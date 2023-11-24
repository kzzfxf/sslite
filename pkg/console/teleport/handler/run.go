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
	config, err := os.ReadFile(flags.ConfigFile)
	if err != nil {
		return
	}
	err = service.Teleport.Init(config)
	if err != nil {
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)

	if flags.HttpPort > 0 {
		go func() {
			defer wg.Done()
			err = http.Start(ctx, fmt.Sprintf(":%d", flags.HttpPort))
		}()
	}

	if flags.HttpPort > 0 {
		go func() {
			defer wg.Done()
			err = socket.Start(ctx, "tcp", fmt.Sprintf(":%d", flags.SocketPort))
		}()
	}

	if flags.OpenUI {
		go func() {
			defer wg.Done()
			err = ui.ShowMainUI()
		}()
	}

	wg.Wait()

	return
}
