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
	"io"
	"net"
	"net/http"
	"time"

	"github.com/kzzfxf/teleport/pkg/core/dialer/direct"
	"github.com/kzzfxf/teleport/pkg/core/dialer/reject"
	"github.com/kzzfxf/teleport/pkg/core/dialer/shadowsocks"
)

var (
	TunnelDirectID = "DIRECT"
	TunnelRejectID = "REJECT"
)

type Engine struct {
	tunnels map[string]*Tunnel
}

// NewEngine
func NewEngine() (tp *Engine) {
	tp = &Engine{
		tunnels: make(map[string]*Tunnel),
	}
	tp.tunnels[TunnelDirectID] = NewTunnel("direct", direct.NewDirect())
	tp.tunnels[TunnelRejectID] = NewTunnel("reject", reject.NewReject())
	ss, err := shadowsocks.NewShadowsocks("", "", "")
	if err != nil {
		panic(err)
	}
	tp.tunnels["test"] = NewTunnel("miaona", ss)
	return
}

// Mount
func (tp *Engine) Mount(tun *Tunnel) {
	tp.tunnels[randomN(12)] = tun
}

// Direct
func (tp *Engine) Direct() (tun *Tunnel) {
	return tp.tunnels[TunnelDirectID]
}

// Reject
func (tp *Engine) Reject() (tun *Tunnel) {
	return tp.tunnels[TunnelRejectID]
}

// Select
func (tp *Engine) Select(addr string) (tun *Tunnel) {
	return tp.tunnels["test"]
}

// ServeHTTP
func (tp *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	transport := &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			return tp.Select(r.Host).Dial("tcp", r.Host)
		},
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	resp, err := transport.RoundTrip(r)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	header := w.Header()
	for k, vv := range resp.Header {
		for _, v := range vv {
			header.Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)

	var writer io.Writer = w
	if len(resp.TransferEncoding) > 0 && resp.TransferEncoding[0] == "chunked" {
		writer = ChunkWriter{Writer: w}
	}
	io.Copy(writer, resp.Body)
}

// ServeSocket
func (tp *Engine) ServeSocket(client net.Conn, addr string) {
	server, err := tp.Select(addr).Dial("tcp", addr)
	if err != nil {
		client.Close()
		return
	}
	ladder := Ladder{client: client, server: server}
	ladder.Go()
}
