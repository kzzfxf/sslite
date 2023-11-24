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
	"sync"
	"time"

	"github.com/kzzfxf/teleport/pkg/core/dialer/direct"
	"github.com/kzzfxf/teleport/pkg/core/dialer/reject"
	"github.com/kzzfxf/teleport/pkg/core/internal"
)

var (
	TunnelDirectID = "DIRECT"
	TunnelRejectID = "REJECT"
)

type Engine struct {
	bridges map[string]Bridge
	tunnels map[string]*Tunnel
	route   *Route
	rules   *Rules
	locker  sync.RWMutex
}

// NewEngine
func NewEngine() (tp *Engine) {
	tp = &Engine{
		bridges: make(map[string]Bridge),
		tunnels: make(map[string]*Tunnel),
		route:   NewRoute(),
		rules:   NewRules(),
	}
	direct := NewTunnel("direct.proxy", direct.NewDirect(3000*time.Millisecond))
	direct.SetLabel("direct")
	tp.tunnels[TunnelDirectID] = direct
	reject := NewTunnel("reject.proxy", reject.NewReject())
	reject.SetLabel("reject")
	tp.tunnels[TunnelRejectID] = reject
	return
}

// Route
func (tp *Engine) Route() (r *Route) {
	return tp.route
}

// Rules
func (tp *Engine) Rules() (r *Rules) {
	return tp.rules
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

// Direct
func (tp *Engine) Direct() (tun *Tunnel) {
	tp.locker.RLock()
	defer tp.locker.RUnlock()
	return tp.tunnels[TunnelDirectID]
}

// Reject
func (tp *Engine) Reject() (tun *Tunnel) {
	tp.locker.RLock()
	defer tp.locker.RUnlock()
	return tp.tunnels[TunnelRejectID]
}

// ServeHTTP
func (tp *Engine) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	tunnel := tp.MatchTunnel(r.Host)
	if tunnel == nil {
		fmt.Printf("Select tunnel for %s failed\n", r.Host)
		return
	}
	tp.transport(ctx, NewHttpBridge(w, r, r.Host), tunnel)
}

// ServeSocket
func (tp *Engine) ServeSocket(ctx context.Context, client net.Conn, serverAddr string) {
	tunnel := tp.MatchTunnel(serverAddr)
	if tunnel == nil {
		fmt.Printf("Select tunnel for %s failed\n", serverAddr)
		return
	}
	tp.transport(ctx, NewSocketBridge(client, serverAddr), tunnel)
}

// transport
func (tp *Engine) transport(ctx context.Context, bridge Bridge, tunnel *Tunnel) {
	bridgeID := tp.AddBridge(bridge)
	defer func() {
		tp.RemoveBridge(bridgeID)
	}()
	err := bridge.Transport(ctx, tunnel)
	if err != nil {
		fmt.Printf("Transport failed, error = %s\n", err.Error())
		return
	}
}
