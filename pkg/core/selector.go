// Copyright 2023 kzzfxf
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/kzzfxf/teleport/pkg/utils"
)

type SelectOperation int

const (
	SelectOpAnd SelectOperation = 0
	SelectOpOr  SelectOperation = 1
)

// MatchTunnel
func (tp *Engine) MatchTunnel(addr string) (tunnel *Tunnel, forward string) {
	hostname, port, err := utils.ParseAddr(addr)
	if err != nil {
		return
	}
	tunnel, forward = tp.match(hostname, port)
	if forward != "" {
		if !utils.IsValidAddr(forward) {
			forward = fmt.Sprintf("%s:%d", forward, port)
		}
	}
	return
}

// match
func (tp *Engine) match(hostname string, port uint) (tunnel *Tunnel, forward string) {
	tunnel, forward, ok := tp.route.Get(hostname)
	if ok {
		return tunnel, forward
	}

	selector, forward, matched := tp.rules.Match(hostname)
	if matched == "" {
		return nil, ""
	}

	defer func() {
		if tunnel != nil {
			fmt.Printf("%s => %s => (%s)\n", hostname, matched, selector)
			tp.route.Set(hostname, forward, tunnel, time.Now().Add(60*time.Second))
		}
	}()

	if selector == TunnelGlobalName {
		selector = tp.global
	} else if selector == TunnelDirectName {
		return tp.GetDirectTunnel(), forward
	} else if selector == TunnelRejectName {
		return tp.GetRejectTunnel(), forward
	}

	var labels []string
	labels = append(labels, strings.Split(selector, ",")...)

	if len(labels) <= 0 {
		return nil, ""
	}

	tunnels := tp.SelectTunnels(SelectOpAnd, labels...)
	if len(tunnels) <= 0 {
		return nil, ""
	}

	// Sort
	sort.Slice(tunnels, func(i, j int) bool {
		if tunnels[i].latency.value <= 0 && tunnels[j].latency.value > 0 {
			return false
		}
		if tunnels[i].latency.value > 0 && tunnels[j].latency.value <= 0 {
			return true
		}
		return tunnels[i].latency.value <= tunnels[j].latency.value
	})

	for _, tunnel := range tunnels {
		return tunnel, forward
	}

	return nil, ""
}

// SelectTunnels
func (tp *Engine) SelectTunnels(op SelectOperation, labels ...string) (tunnels []*Tunnel) {
	tp.locker.RLock()
	defer tp.locker.RUnlock()
	if len(labels) <= 0 {
		return
	}
	for _, tunnel := range tp.tunnels {
		hits := 0
		for _, label := range labels {
			if tunnel.Is(label) {
				if op == SelectOpAnd {
					hits++
				} else if op == SelectOpOr {
					tunnels = append(tunnels, tunnel)
					break
				}
			}
		}
		if op == SelectOpAnd && hits == len(labels) {
			tunnels = append(tunnels, tunnel)
		}
	}
	return
}
