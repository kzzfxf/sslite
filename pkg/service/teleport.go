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

package service

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	json "github.com/json-iterator/go"
	"github.com/kzzfxf/teleport/pkg/config"
	"github.com/kzzfxf/teleport/pkg/core"
)

type teleport interface {
	Init(config []byte) (err error)
	ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request)
	ServeHTTPS(ctx context.Context, client net.Conn, dstAddr string)
	ServeSocket(ctx context.Context, client net.Conn, dstAddr string)
}

type teleportImpl struct {
	config config.Config
	engine *core.Engine
}

var Teleport teleport = &teleportImpl{
	config: config.Config{},
	engine: core.NewEngine(),
}

func (tp *teleportImpl) Init(config []byte) (err error) {
	err = json.Unmarshal(config, &tp.config)
	if err != nil {
		return
	}
	for _, route := range tp.config.Routes {
		tp.engine.Rules().Put(route.Hostname, "", route.Selector)
	}
	for _, group := range tp.config.Groups {
		if len(group.Hostnames) <= 0 {
			continue
		} else {
			tp.engine.Rules().Group(group.Name, group.Hostnames...)
		}
	}
	for _, proxy := range tp.config.Proxies {
		dialer, err := core.NewDialerWithURL(proxy.Type, proxy.URL)
		if err != nil {
			return err
		}
		tunnel := core.NewTunnel(proxy.Name, dialer)
		tunnel.SetLabel(dialer.Addr())
		for _, label := range proxy.Labels {
			tunnel.SetLabel(label)
		}
		// Ignore direct and reject
		if proxy.Type != "direct" && proxy.Type != "reject" {
			tunnel.SetupLatencyTester(tp.config.Latency.URL, time.Duration(tp.config.Latency.Timeout)*time.Millisecond)
		}
		tp.engine.AddTunnel(tunnel)
		fmt.Printf("New tunnel %s\n", tunnel.Name())
	}
	return
}

// func (tp *teleportImpl) ImportConfig() {

// }

func (tp *teleportImpl) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	tp.engine.ServeHTTP(ctx, w, r)
}

func (tp *teleportImpl) ServeHTTPS(ctx context.Context, client net.Conn, dstAddr string) {
	tp.engine.ServeSocket(ctx, client, dstAddr)
}

func (tp *teleportImpl) ServeSocket(ctx context.Context, client net.Conn, dstAddr string) {
	tp.engine.ServeSocket(ctx, client, dstAddr)
}
