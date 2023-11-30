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
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/kzzfxf/teleport/pkg/common"
	"github.com/kzzfxf/teleport/pkg/config"
	"github.com/kzzfxf/teleport/pkg/core/dialer/direct"
	"github.com/kzzfxf/teleport/pkg/core/dialer/reject"
	"github.com/kzzfxf/teleport/pkg/core/internal"
	"github.com/kzzfxf/teleport/pkg/core/rules"
)

const (
	BuiltinTunnelGlobalName = "GLOBAL"
	BuiltinTunnelDirectName = "DIRECT"
	BuiltinTunnelRejectName = "REJECT"
)

type Engine struct {
	bridges map[string]Bridge
	tunnels map[string]*Tunnel
	conf    *config.Config
	global  string
	route   *Route
	rules   *rules.Rules
	locker  sync.RWMutex
}

// NewEngine
func NewEngine(conf *config.Config, ruleConf *config.Rules) (tp *Engine, err error) {
	tp = &Engine{
		bridges: make(map[string]Bridge),
		tunnels: make(map[string]*Tunnel),
		conf:    conf,
		global:  conf.Global,
		route:   NewRoute(),
	}
	// Builtin tunnels
	tp.tunnels[BuiltinTunnelDirectName] = NewTunnel(BuiltinTunnelDirectName, direct.NewDirect(3000*time.Millisecond))
	tp.tunnels[BuiltinTunnelRejectName] = NewTunnel(BuiltinTunnelRejectName, reject.NewReject())
	// Init engine
	for _, proxy := range conf.Proxies {
		dialer, err := NewDialerWithURL(proxy.Type, proxy.URL)
		if err != nil {
			return nil, err
		}
		tunnel := NewTunnel(proxy.Name, dialer)
		tunnel.SetLabel(dialer.Addr())
		for _, label := range proxy.Labels {
			tunnel.SetLabel(label)
		}
		// Ignore direct and reject
		if proxy.Type != "direct" && proxy.Type != "reject" {
			tunnel.SetupLatencyTester(conf.Latency.URL, time.Duration(conf.Latency.Timeout)*time.Millisecond)
		}
		tp.AddTunnel(tunnel)
	}
	// Init rules
	tp.rules = rules.NewRules(ruleConf)

	return
}

// Route
func (tp *Engine) Route() (r *Route) {
	return tp.route
}

// Rules
func (tp *Engine) Rules() (r *rules.Rules) {
	return tp.rules
}

// GetDirectTunnel
func (tp *Engine) GetDirectTunnel() (tunnel *Tunnel) {
	tunnel, _ = tp.GetTunnel(BuiltinTunnelDirectName)
	return
}

// GetRejectTunnel
func (tp *Engine) GetRejectTunnel() (tunnel *Tunnel) {
	tunnel, _ = tp.GetTunnel(BuiltinTunnelRejectName)
	return
}

// GetTunnel
func (tp *Engine) GetTunnel(tunnelID string) (tunnel *Tunnel, ok bool) {
	tp.locker.RLock()
	defer tp.locker.RUnlock()
	tunnel, ok = tp.tunnels[tunnelID]
	return
}

// AddTunnel
func (tp *Engine) AddTunnel(tunnel *Tunnel) (tunnelID string) {
	tunnelID = internal.RandomN(12)
	tp.locker.Lock()
	defer tp.locker.Unlock()
	tp.tunnels[tunnelID] = tunnel
	return
}

// RemoveTunnel
func (tp *Engine) RemoveTunnel(tunnelID string) {
	tp.locker.Lock()
	defer tp.locker.Unlock()
	delete(tp.tunnels, tunnelID)
}

// GetBridge
func (tp *Engine) GetBridge(bridgeID string) (bridge Bridge, ok bool) {
	tp.locker.RLock()
	defer tp.locker.RUnlock()
	bridge, ok = tp.bridges[bridgeID]
	return
}

// AddBridge
func (tp *Engine) AddBridge(bridge Bridge) (bridgeID string) {
	bridgeID = internal.RandomN(16)
	tp.locker.Lock()
	defer tp.locker.Unlock()
	tp.bridges[bridgeID] = bridge
	return
}

// RemoveBridge
func (tp *Engine) RemoveBridge(bridgeID string) {
	tp.locker.Lock()
	defer tp.locker.Unlock()
	delete(tp.bridges, bridgeID)
}

// ServeHTTP
func (tp *Engine) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	dstAddr := r.Host
	if !strings.Contains(dstAddr, ":") {
		dstAddr = fmt.Sprintf("%s:%d", dstAddr, 80)
	}
	tunnel := tp.MatchTunnel(dstAddr)
	if tunnel == nil {
		fmt.Printf("Select tunnel for %s failed\n", dstAddr)
		return
	}
	tp.transport(ctx, NewHttpBridge(w, r, dstAddr), tunnel)
}

// ServeSocket
func (tp *Engine) ServeSocket(ctx context.Context, client net.Conn, dstAddr string) {
	tunnel := tp.MatchTunnel(dstAddr)
	if tunnel == nil {
		fmt.Printf("Select tunnel for %s failed\n", dstAddr)
		return
	}
	tp.transport(ctx, NewSocketBridge(client, dstAddr), tunnel)
}

// transport
func (tp *Engine) transport(ctx context.Context, bridge Bridge, tunnel *Tunnel) {
	bridgeID := tp.AddBridge(bridge)
	defer func() {
		tp.RemoveBridge(bridgeID)
	}()

	fmt.Printf("%s -> %s -> %s -> %s\n", bridge.InBound(), ctx.Value(common.ContextEntry), tunnel.Name(), bridge.OutBound())

	dialFn := func(network, addr string) (net.Conn, error) {
		return tunnel.Dial(network, addr)
	}
	err := bridge.Transport(ctx, dialFn)
	if err != nil {
		fmt.Printf("Transport %s->%s failed, error = %s\n", bridge.InBound(), bridge.OutBound(), err.Error())
		return
	}
}
