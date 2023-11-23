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
func (tp *Engine) MatchTunnel(addr string) (tunnel *Tunnel) {
	domain, ip, port, err := utils.ParseAddr(addr)
	if err != nil {
		return
	}
	return tp.match(domain, ip, port)
}

// match
func (tp *Engine) match(domain, ip string, port uint) (tunnel *Tunnel) {
	server := domain
	if server == "" {
		server = ip
	}

	_, tunnel, ok := tp.route.Get(server)
	if ok {
		return tunnel
	}

	expectedIP, selector, ok := tp.rules.Match(server)
	if !ok {
		return nil
	}

	var labels []string
	labels = append(labels, strings.Split(selector, ",")...)

	if len(labels) <= 0 {
		return
	}

	defer func() {
		tp.route.Put(server, expectedIP, tunnel, time.Now().Add(600*time.Second))
	}()

	tunnels := tp.SelectTunnels(SelectOpAnd, labels...)
	if len(tunnels) <= 0 {
		return
	}
	return tunnels[0]
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
