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
	"fmt"
	"net"
	"net/http"

	json "github.com/json-iterator/go"
	"github.com/kzzfxf/teleport/pkg/config"
	"github.com/kzzfxf/teleport/pkg/core"
	"github.com/kzzfxf/teleport/pkg/core/dialer"
)

type teleport interface {
	Init(config []byte) (err error)
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	ServeHTTPS(client net.Conn, server string)
	ServeSocket(client net.Conn, server string)
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
		tp.engine.Rules().Put(route.Server, route.IP, route.Selector)
	}
	for _, group := range tp.config.Groups {
		if len(group.Servers) <= 0 {
			continue
		} else if len(group.Servers) == 1 {
			tp.engine.Rules().Group(group.Name, group.Servers[0])
		} else {
			tp.engine.Rules().Group(group.Name, group.Servers[0], group.Servers[1:]...)
		}
	}
	for _, proxy := range tp.config.Proxies {
		dialer, err := dialer.NewDialerWithURL(proxy.Type, proxy.URL)
		if err != nil {
			return err
		}
		tunnel := core.NewTunnel(proxy.Name, dialer)
		tunnel.SetLabel(dialer.Addr())
		for _, label := range proxy.Labels {
			tunnel.SetLabel(label)
		}
		tp.engine.AddTunnel(tunnel)
		fmt.Printf("New tunnel %s\n", tunnel.Name())
	}
	return
}

func (tp *teleportImpl) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tp.engine.ServeHTTP(w, r)
}

func (tp *teleportImpl) ServeHTTPS(client net.Conn, server string) {
	tp.engine.ServeSocket(client, server)
}

func (tp *teleportImpl) ServeSocket(client net.Conn, server string) {
	tp.engine.ServeSocket(client, server)
}
