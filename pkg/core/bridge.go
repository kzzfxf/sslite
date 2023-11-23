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
	"io"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kzzfxf/teleport/pkg/core/internal"
)

const (
	BridgeStatusFailed       int32 = -1
	BridgeStatusReady        int32 = 0
	BridgeStatusConnecting   int32 = 1
	BridgeStatusTransporting int32 = 2
	BridgeStatusDisconnected int32 = 3
)

type Bridge interface {

	// Status
	Status() (status int32)

	// Transport
	Transport() (err error)
}

type HttpBridge struct {
	w      http.ResponseWriter
	r      *http.Request
	tunnel *Tunnel
	status int32
}

// NewHttpBridge
func NewHttpBridge(w http.ResponseWriter, r *http.Request, tunnel *Tunnel) (hb *HttpBridge) {
	return &HttpBridge{w: w, r: r, tunnel: tunnel}
}

// Status
func (hb *HttpBridge) Status() (status int32) {
	return atomic.LoadInt32(&hb.status)
}

// Transport
func (hb *HttpBridge) Transport() (err error) {

	fmt.Printf("%s -> %s -> %s\n", hb.r.RemoteAddr, hb.tunnel.Name(), hb.r.Host)

	transport := &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			return hb.tunnel.Dial("tcp", hb.r.Host)
		},
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	atomic.StoreInt32(&hb.status, BridgeStatusConnecting)

	resp, err := transport.RoundTrip(hb.r)
	if err != nil {
		atomic.StoreInt32(&hb.status, BridgeStatusFailed)
		return
	}

	atomic.StoreInt32(&hb.status, BridgeStatusTransporting)

	defer resp.Body.Close()

	header := hb.w.Header()
	for k, vv := range resp.Header {
		for _, v := range vv {
			header.Add(k, v)
		}
	}
	hb.w.WriteHeader(resp.StatusCode)

	var writer io.Writer = hb.w
	if len(resp.TransferEncoding) > 0 && resp.TransferEncoding[0] == "chunked" {
		writer = internal.ChunkWriter{Writer: hb.w}
	}
	io.Copy(writer, resp.Body)

	atomic.StoreInt32(&hb.status, BridgeStatusDisconnected)

	return
}

type SocketBridge struct {
	client net.Conn
	server string
	tunnel *Tunnel
	status int32
}

// NewSocketBridge
func NewSocketBridge(client net.Conn, server string, tunnel *Tunnel) (sb *SocketBridge) {
	return &SocketBridge{client: client, server: server, tunnel: tunnel}
}

// Status
func (sb *SocketBridge) Status() (status int32) {
	return atomic.LoadInt32(&sb.status)
}

// Transport
func (sb *SocketBridge) Transport() (err error) {

	fmt.Printf("%s -> %s -> %s\n", sb.client.RemoteAddr(), sb.tunnel.Name(), sb.server)

	atomic.StoreInt32(&sb.status, BridgeStatusConnecting)

	server, err := sb.tunnel.Dial("tcp", sb.server)
	if err != nil {
		atomic.StoreInt32(&sb.status, BridgeStatusFailed)
		return
	}
	defer func() {
		server.Close()
		atomic.StoreInt32(&sb.status, BridgeStatusDisconnected)
	}()

	var wg sync.WaitGroup
	wg.Add(2)

	atomic.StoreInt32(&sb.status, BridgeStatusTransporting)

	go func() {
		defer wg.Done()
		io.Copy(server, sb.client)
	}()
	go func() {
		defer wg.Done()
		io.Copy(sb.client, server)
	}()

	wg.Wait()

	return
}
