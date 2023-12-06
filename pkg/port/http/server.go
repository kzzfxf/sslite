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

	"github.com/kzzfxf/sslite/pkg/common"
	"github.com/kzzfxf/sslite/pkg/logkit"
	"github.com/kzzfxf/sslite/pkg/service"
)

// Start
func Start(ctx context.Context, addr string) (err error) {
	return httpPkg.ListenAndServe(addr, handler(ctx, addr))
}

// handler
func handler(ctx context.Context, addr string) httpPkg.HandlerFunc {
	return func(w httpPkg.ResponseWriter, r *httpPkg.Request) {
		reqCtx, cancel := context.WithCancel(r.Context())
		reqCtx = context.WithValue(reqCtx, common.ContextEntry, addr)

		go func() {
			defer cancel()
			select {
			case <-ctx.Done():
				return
			case <-reqCtx.Done():
				return
			}
		}()
		defer cancel()

		if r.Method != httpPkg.MethodConnect {
			service.SSLite.ServeHTTP(reqCtx, w, r)
		} else {
			hijacker, ok := w.(httpPkg.Hijacker)
			if !ok {
				return
			}
			conn, _, err := hijacker.Hijack()
			if err != nil {
				logkit.Error("handler call Hijack failed", logkit.Any("error", err))
				return
			}
			conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
			service.SSLite.ServeHTTPS(reqCtx, conn, r.Host)
		}
	}
}
