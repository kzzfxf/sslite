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
	"net"
	"sync"

	"github.com/kzzfxf/teleport/pkg/core/dialer"
)

type Tunnel struct {
	name   string
	dialer dialer.Dialer
	labels map[string]struct{}
	locker sync.RWMutex
}

// OpenTunnel
func NewTunnel(name string, dialer dialer.Dialer) (tun *Tunnel) {
	tun = &Tunnel{
		name:   name,
		dialer: dialer,
		labels: make(map[string]struct{}),
	}
	tun.Label(name)
	return
}

// Name
func (tun *Tunnel) Name() (name string) {
	return tun.name
}

// Dial
func (tun *Tunnel) Dial(network, addr string) (conn net.Conn, err error) {
	return tun.dialer.Dial(network, addr)
}

// Is
func (tun *Tunnel) Is(label string) (yes bool) {
	tun.locker.RLock()
	defer tun.locker.RUnlock()
	_, yes = tun.labels[label]
	return
}

// Label
func (tun *Tunnel) Label(label string) {
	tun.locker.Lock()
	defer tun.locker.Unlock()
	if label != "" {
		tun.labels[label] = struct{}{}
	}
}

// Close
func (tun *Tunnel) Close() (err error) {
	return tun.dialer.Close()
}
