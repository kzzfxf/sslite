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

package main

import (
	"github.com/kzzfxf/sslite/pkg/console/sslite/handler"
	"github.com/spf13/cobra"
)

var (
	runc = &cobra.Command{}
)

func init() {
	var (
		flags = handler.NewRunFlags(gflags)
	)
	runc.Use = "run"
	runc.Short = "Start proxy process"
	runc.Long = "Start proxy process and listen for proxy connections."
	// Events
	runc.RunE = func(cmd *cobra.Command, args []string) error {
		return handler.OnRunHandler(cmd.Context(), flags, args)
	}
	// Flags
	if f := runc.Flags(); f != nil {
		f.IntVarP(&flags.HttpPort, "with-http", "", flags.HttpPort, "set a http port")
		f.IntVarP(&flags.SocketPort, "with-socket", "", flags.SocketPort, "set a socket port")
		f.BoolVarP(&flags.ShowUI, "with-gui", "", flags.ShowUI, "show terminal GUI")
	}
}
