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
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrNoLatencyTesterFound = errors.New("no latency tester found")
)

type Tunnel struct {
	name       string
	dialer     Dialer
	down, up   chan int
	downNBytes uint64
	upNBytes   uint64
	latency    struct {
		addr    string
		url     string
		timeout time.Duration
		value   time.Duration
	}
	labels map[string]string
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	locker sync.RWMutex
}

// OpenTunnel
func NewTunnel(name string, dialer Dialer) (tun *Tunnel) {
	tun = &Tunnel{
		name:   name,
		dialer: dialer,
		down:   make(chan int, 10240),
		up:     make(chan int, 10240),
		labels: make(map[string]string),
	}
	tun.SetLabel(name)

	tun.ctx, tun.cancel = context.WithCancel(context.Background())
	tun.wg.Add(1)
	// Background goroutine
	go tun.background()

	return
}

// background
func (tun *Tunnel) background() {
	ticker := time.NewTicker(1 * time.Second)
	defer func() {
		ticker.Stop()
		tun.wg.Done()
	}()
	for {
		select {
		case <-tun.ctx.Done():
			return
		case <-ticker.C:
			tun.evaluate()
		case n := <-tun.down:
			atomic.AddUint64(&tun.downNBytes, uint64(n))
		case n := <-tun.up:
			atomic.AddUint64(&tun.upNBytes, uint64(n))
		}
	}
}

// Name
func (tun *Tunnel) Name() (name string) {
	return tun.name
}

// Dial
func (tun *Tunnel) Dial(network, addr string) (conn net.Conn, err error) {
	conn, err = tun.dialer.Dial(network, addr)
	if err == nil {
		conn = &ConnTrafficTracker{Conn: conn, down: tun.down, up: tun.up}
	}
	return
}

// DownNBytes
func (tun *Tunnel) DownNBytes() uint64 {
	return atomic.LoadUint64(&tun.downNBytes)
}

// UpNBytes
func (tun *Tunnel) UpNBytes() uint64 {
	return atomic.LoadUint64(&tun.upNBytes)
}

// Is
func (tun *Tunnel) Is(label string) (hit bool) {
	tun.locker.RLock()
	defer tun.locker.RUnlock()
	if label == "" {
		return false
	}
	_, hit = tun.labels[label]
	return
}

// SetLabel
func (tun *Tunnel) SetLabel(label string) {
	tun.locker.Lock()
	defer tun.locker.Unlock()
	if label != "" {
		tun.labels[label] = label
	}
}

// RemoveLabel
func (tun *Tunnel) RemoveLabel(label string) {
	tun.locker.Lock()
	defer tun.locker.Unlock()
	if label != "" {
		delete(tun.labels, label)
	}
}

// SetupLatencyTester
func (tun *Tunnel) SetupLatencyTester(URL string, timeout time.Duration) {
	if URL != "" {
		u, err := url.Parse(URL)
		if err != nil {
			return
		}
		addr := ""
		if u.Hostname() == "" {
			return
		} else if port := u.Port(); port != "" {
			addr = u.Host
		} else {
			if u.Scheme == "http" {
				addr = fmt.Sprintf("%s:%d", u.Hostname(), 80)
			} else if u.Scheme == "https" {
				addr = fmt.Sprintf("%s:%d", u.Hostname(), 443)
			} else {
				return
			}
		}

		tun.latency.addr = addr
		tun.latency.url = URL

		if timeout <= 0 {
			tun.latency.timeout = 3000 * time.Millisecond
		}
	}
}

// TestLatency
func (tun *Tunnel) TestLatency() (latency time.Duration, err error) {
	if tun.latency.addr == "" ||
		tun.latency.url == "" {
		return 0, ErrNoLatencyTesterFound
	}
	t := &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			return tun.Dial("tcp", tun.latency.addr)
		},
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	c := &http.Client{
		Transport: t,
		Timeout:   tun.latency.timeout,
	}

	start := time.Now()
	_, err = c.Get(tun.latency.url)
	if err != nil {
		latency = -1
	} else {
		latency = time.Since(start)
	}

	return
}

// evaluate
func (tun *Tunnel) evaluate() {

}

// Close
func (tun *Tunnel) Close() (err error) {
	if tun.cancel != nil {
		tun.cancel()
	}
	tun.wg.Wait()
	return tun.dialer.Close()
}
