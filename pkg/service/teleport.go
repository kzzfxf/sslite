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
	"net"
	"net/http"

	"github.com/kzzfxf/teleport/pkg/core"
)

type teleport interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	ServeHTTPS(client net.Conn, addr string)
	ServeSocket(client net.Conn, addr string)
}

type teleportImpl struct {
	engine *core.Engine
}

var Teleport teleport = &teleportImpl{
	engine: core.NewEngine(),
}

func (tp *teleportImpl) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tp.engine.ServeHTTP(w, r)
}

func (tp *teleportImpl) ServeHTTPS(client net.Conn, addr string) {
	tp.engine.ServeSocket(client, addr)
}

func (tp *teleportImpl) ServeSocket(client net.Conn, addr string) {
	tp.engine.ServeSocket(client, addr)
}
