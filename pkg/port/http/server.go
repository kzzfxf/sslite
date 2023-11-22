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

package http

import (
	"context"
	httpPkg "net/http"

	"github.com/kzzfxf/teleport/pkg/service"
)

func Start(ctx context.Context, addr string) (err error) {
	server := &httpPkg.Server{
		Addr: addr,
		Handler: httpPkg.HandlerFunc(func(w httpPkg.ResponseWriter, r *httpPkg.Request) {
			if r.Method == httpPkg.MethodConnect {
				handleTunneling(w, r)
			} else {
				handleHTTP(w, r)
			}
		}),
	}
	return server.ListenAndServe()
}

func handleTunneling(w httpPkg.ResponseWriter, r *httpPkg.Request) {
	hijacker, ok := w.(httpPkg.Hijacker)
	if !ok {
		return
	}
	conn, _, err := hijacker.Hijack()
	if err != nil {
		return
	}
	conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	service.Teleport.ServeHTTPS(conn, r.Host)
}

func handleHTTP(w httpPkg.ResponseWriter, r *httpPkg.Request) {
	service.Teleport.ServeHTTP(w, r)
}
