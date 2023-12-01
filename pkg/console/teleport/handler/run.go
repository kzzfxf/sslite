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

package handler

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/kzzfxf/teleport/pkg/logkit"
	"github.com/kzzfxf/teleport/pkg/port/http"
	"github.com/kzzfxf/teleport/pkg/port/socket"
	"github.com/kzzfxf/teleport/pkg/service"
	"github.com/kzzfxf/teleport/pkg/ui"
)

type RunFlags struct {
	*GlobalFlags
	HttpPort   int
	SocketPort int
	OpenUI     bool
}

func NewRunFlags(gflags *GlobalFlags) (flags *RunFlags) {
	flags = &RunFlags{GlobalFlags: gflags}
	flags.HttpPort = 8998
	flags.SocketPort = 8999
	flags.OpenUI = false
	return
}

func OnRunHandler(ctx context.Context, flags *RunFlags, args []string) (err error) {
	data0, err := os.ReadFile(flags.BaseConfigFile)
	if err != nil {
		logkit.Error("call os.ReadFile failed", logkit.WithAttr("error", err), logkit.WithAttr("config", flags.BaseConfigFile))
		return
	}

	conf, err := service.Config.LoadConfig(data0)
	if err != nil {
		logkit.Error("call service.Config.LoadConfig failed", logkit.WithAttr("error", err))
		return
	}

	data1, err := os.ReadFile(flags.RulesConfigFile)
	if err != nil {
		logkit.Error("call os.ReadFile failed", logkit.WithAttr("error", err), logkit.WithAttr("config", flags.RulesConfigFile))
		return
	}

	rulesConf, err := service.Config.LoadRules(data1)
	if err != nil {
		logkit.Error("call service.Config.LoadRules failed", logkit.WithAttr("error", err))
		return
	}

	err = service.Teleport.Init(conf, rulesConf)
	if err != nil {
		logkit.Error("call service.Teleport.Init failed", logkit.WithAttr("error", err))
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)

	if flags.HttpPort > 0 {
		go func() {
			defer wg.Done()
			err = http.Start(ctx, fmt.Sprintf(":%d", flags.HttpPort))
			if err != nil {
				logkit.Error("call http.Start failed", logkit.WithAttr("error", err), logkit.WithAttr("port", flags.HttpPort))
			}
		}()
	}

	if flags.HttpPort > 0 {
		go func() {
			defer wg.Done()
			err = socket.Start(ctx, "tcp", fmt.Sprintf(":%d", flags.SocketPort))
			if err != nil {
				logkit.Error("call socket.Start failed", logkit.WithAttr("error", err), logkit.WithAttr("port", flags.SocketPort))
			}
		}()
	}

	if flags.OpenUI {
		go func() {
			defer wg.Done()
			err = ui.ShowMainUI()
			if err != nil {
				logkit.Error("call ui.ShowMainUI failed", logkit.WithAttr("error", err))
			}
		}()
	}

	wg.Wait()

	return
}
