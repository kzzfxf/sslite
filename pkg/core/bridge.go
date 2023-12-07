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
	"io"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kzzfxf/sslite/pkg/core/internal"
)

const (
	BridgeStatusFailed     int32 = -1
	BridgeStatusReady      int32 = 0
	BridgeStatusConnect    int32 = 1
	BridgeStatusTraffic    int32 = 2
	BridgeStatusDisconnect int32 = 3
)

// ParseBridgeStatus
func GetBridgeStatus(status int32) string {
	switch status {
	case BridgeStatusFailed:
		return "failed"
	case BridgeStatusReady:
		return "ready"
	case BridgeStatusConnect:
		return "connect"
	case BridgeStatusTraffic:
		return "traffic"
	case BridgeStatusDisconnect:
		return "disconnect"
	}
	return "unknown"
}

type DialFunc func(network, addr string) (conn net.Conn, err error)

type Bridge interface {

	// Status
	Status() (status int32)

	// InBound returns the source address.
	InBound() (addr string)

	// OutBound returns the destination address.
	OutBound() (addr string)

	// OutBoundReal returns the real destination address.
	OutBoundReal() (addr string)

	// Forward returns the forward address.
	Forward() (addr string)

	// Tunnel
	Tunnel() (tunnel *Tunnel)

	// Transport
	Transport(ctx context.Context, tunnel *Tunnel) (err error)
}

type HttpBridge struct {
	w       http.ResponseWriter
	r       *http.Request
	tunnel  *Tunnel
	dstAddr string
	forward string
	status  int32
}

// NewHttpBridge
func NewHttpBridge(w http.ResponseWriter, r *http.Request, dstAddr, forward string) (hb *HttpBridge) {
	return &HttpBridge{w: w, r: r, dstAddr: dstAddr, forward: forward}
}

// Status
func (hb *HttpBridge) Status() (status int32) {
	return atomic.LoadInt32(&hb.status)
}

// InBound
func (hb *HttpBridge) InBound() (addr string) {
	return hb.r.RemoteAddr
}

// OutBound
func (hb *HttpBridge) OutBound() (addr string) {
	return hb.dstAddr
}

// OutBoundReal
func (hb *HttpBridge) OutBoundReal() (addr string) {
	if hb.forward != "" {
		return hb.forward
	}
	return hb.dstAddr
}

// Forward
func (hb *HttpBridge) Forward() (addr string) {
	return hb.forward
}

// Tunnel
func (hb *HttpBridge) Tunnel() (tunnel *Tunnel) {
	return hb.tunnel
}

// Transport
func (hb *HttpBridge) Transport(ctx context.Context, tunnel *Tunnel) (err error) {
	hb.tunnel = tunnel

	transport := &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			return tunnel.Dial("tcp", hb.OutBoundReal())
		},
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	atomic.StoreInt32(&hb.status, BridgeStatusConnect)

	resp, err := transport.RoundTrip(hb.r)
	if err != nil {
		atomic.StoreInt32(&hb.status, BridgeStatusFailed)
		return
	}

	atomic.StoreInt32(&hb.status, BridgeStatusTraffic)

	defer func() {
		resp.Body.Close()
		atomic.StoreInt32(&hb.status, BridgeStatusDisconnect)
	}()

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

	return
}

type SocketBridge struct {
	client  net.Conn
	tunnel  *Tunnel
	dstAddr string
	forward string
	status  int32
}

// NewSocketBridge
func NewSocketBridge(client net.Conn, dstAddr, forward string) (sb *SocketBridge) {
	return &SocketBridge{client: client, dstAddr: dstAddr, forward: forward}
}

// Status
func (sb *SocketBridge) Status() (status int32) {
	return atomic.LoadInt32(&sb.status)
}

// InBound
func (sb *SocketBridge) InBound() (addr string) {
	return sb.client.RemoteAddr().String()
}

// OutBound
func (sb *SocketBridge) OutBound() (addr string) {
	return sb.dstAddr
}

// OutBoundReal
func (sb *SocketBridge) OutBoundReal() (addr string) {
	if sb.forward != "" {
		return sb.forward
	}
	return sb.dstAddr
}

// Forward
func (sb *SocketBridge) Forward() (addr string) {
	return sb.forward
}

// Tunnel
func (sb *SocketBridge) Tunnel() (tunnel *Tunnel) {
	return sb.tunnel
}

// Transport
func (sb *SocketBridge) Transport(ctx context.Context, tunnel *Tunnel) (err error) {
	sb.tunnel = tunnel

	atomic.StoreInt32(&sb.status, BridgeStatusConnect)

	server, err := tunnel.Dial("tcp", sb.OutBoundReal())
	if err != nil {
		atomic.StoreInt32(&sb.status, BridgeStatusFailed)
		return
	}
	defer func() {
		server.Close()
		atomic.StoreInt32(&sb.status, BridgeStatusDisconnect)
	}()

	var wg sync.WaitGroup
	wg.Add(2)

	atomic.StoreInt32(&sb.status, BridgeStatusTraffic)

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
