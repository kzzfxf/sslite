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

	"github.com/kzzfxf/teleport/pkg/common"
	"github.com/kzzfxf/teleport/pkg/logkit"
	"github.com/kzzfxf/teleport/pkg/service"
	"github.com/kzzfxf/teleport/pkg/utils"
	"github.com/riobard/go-shadowsocks2/socks"
)

func Start(ctx context.Context, network, addr string) (err error) {
	ln, err := net.Listen(network, addr)
	if err != nil {
		logkit.Error("Start call Listen failed", logkit.Any("error", err), logkit.Any("network", network), logkit.Any("addr", addr))
		return
	}

	go func() {
		defer ln.Close()
		<-ctx.Done()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			logkit.Error("Start call Accept failed", logkit.Any("error", err))
			break
		}

		dstAddr, err := socks.Handshake(conn)
		if err != nil {
			logkit.Error("Start call Handshake failed", logkit.Any("error", err))
			conn.Close()
			continue
		}
		// Keepalive
		utils.SetKeepAlive(conn)
		//
		connCtx, cancel := context.WithCancel(ctx)
		connCtx = context.WithValue(connCtx, common.ContextEntry, addr)

		go func() {
			defer cancel()
			select {
			case <-ctx.Done():
				return
			case <-connCtx.Done():
				return
			}
		}()

		go func() {
			defer cancel()
			service.Teleport.ServeSocket(connCtx, conn, dstAddr.String())
		}()
	}
	return
}
