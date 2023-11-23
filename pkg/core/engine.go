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
	tunnels map[string]*Tunnel
	bridges map[string]Bridge
	locker  sync.RWMutex
}

// NewEngine
func NewEngine() (tp *Engine) {
	tp = &Engine{
		tunnels: make(map[string]*Tunnel),
		bridges: make(map[string]Bridge),
	}
	tp.tunnels[TunnelDirectID] = NewTunnel("direct", direct.NewDirect(3*time.Second))
	tp.tunnels[TunnelRejectID] = NewTunnel("reject", reject.NewReject())
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
func (tp *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tunnel := tp.MatchTunnel(r.Host)
	if tunnel == nil {
		fmt.Printf("Select tunnel for %s failed\n", r.Host)
		return
	}
	tp.transport(NewHttpBridge(w, r, tunnel))
}

// ServeSocket
func (tp *Engine) ServeSocket(client net.Conn, server string) {
	tunnel := tp.MatchTunnel(server)
	if tunnel == nil {
		fmt.Printf("Select tunnel for %s failed\n", server)
		return
	}
	tp.transport(NewSocketBridge(client, server, tunnel))
}

// transport
func (tp *Engine) transport(bridge Bridge) {
	bridgeID := tp.AddBridge(bridge)
	defer func() {
		tp.RemoveBridge(bridgeID)
	}()
	err := bridge.Transport()
	if err != nil {
		fmt.Printf("Transport failed, error = %s\n", err.Error())
		return
	}
}
