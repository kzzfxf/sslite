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

package socket

import (
	"context"
	"net"

	"github.com/kzzfxf/teleport/pkg/service"
	"github.com/kzzfxf/teleport/pkg/utils"
	"github.com/riobard/go-shadowsocks2/socks"
)

func Start(ctx context.Context, network, addr string) (err error) {
	ln, err := net.Listen(network, addr)
	if err != nil {
		return
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			break
		}

		server, err := socks.Handshake(conn)
		if err != nil {
			conn.Close()
			continue
		}
		// Keepalive
		utils.SetKeepAlive(conn)
		//
		go func() {
			service.Teleport.ServeSocket(conn, server.String())
		}()
	}
	return
}
