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
	"net"
	"net/http"

	"github.com/kzzfxf/sslite/pkg/config"
	"github.com/kzzfxf/sslite/pkg/core"
)

type sslite interface {
	Init(conf *config.Config, rulesConf *config.Rules) (err error)
	ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request)
	ServeHTTPS(ctx context.Context, client net.Conn, dstAddr string)
	ServeSocket(ctx context.Context, client net.Conn, dstAddr string)
}

type ssliteImpl struct {
	engine *core.Engine
}

var SSLite sslite = &ssliteImpl{}

func (tp *ssliteImpl) Init(conf *config.Config, rulesConf *config.Rules) (err error) {
	engine, err := core.NewEngine(conf, rulesConf)
	if err != nil {
		return
	}
	tp.engine = engine
	return
}

func (tp *ssliteImpl) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	tp.engine.ServeHTTP(ctx, w, r)
}

func (tp *ssliteImpl) ServeHTTPS(ctx context.Context, client net.Conn, dstAddr string) {
	tp.engine.ServeSocket(ctx, client, dstAddr)
}

func (tp *ssliteImpl) ServeSocket(ctx context.Context, client net.Conn, dstAddr string) {
	tp.engine.ServeSocket(ctx, client, dstAddr)
}
