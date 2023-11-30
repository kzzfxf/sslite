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
	"github.com/kzzfxf/teleport/pkg/console/teleport/handler"
	"github.com/spf13/cobra"
)

var (
	Version = "1.0.0"
)

var (
	teleportc = &cobra.Command{}
	gflags    = handler.NewGlobalFlags()
)

func init() {
	// var (
	//    flags = handler.NewTeleportFlags(gflags)
	// )
	teleportc.Use = "teleport"
	teleportc.Short = "A short description"
	teleportc.Long = "A long description"
	teleportc.Version = Version
	teleportc.SilenceUsage = true
	teleportc.CompletionOptions.HiddenDefaultCmd = true
	// Events
	teleportc.RunE = func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
		// return handler.OnTeleportHandler(cmd.Context(), flags, args)
	}
	teleportc.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return handler.OnGlobalBeforeHandler(cmd.Context(), gflags, args)
	}
	teleportc.PersistentPostRunE = func(cmd *cobra.Command, args []string) error {
		return handler.OnGlobalAfterHandler(cmd.Context(), gflags, args)
	}
	// Flags
	// if f := teleportc.Flags(); f != nil {
	//     f.StringVarP(&flags.Test, "test", "t", flags.Test, "a test flag")
	// }
	if pf := teleportc.PersistentFlags(); pf != nil {
		pf.StringVarP(&gflags.BaseConfigFile, "config", "c", gflags.BaseConfigFile, "base config file")
		pf.StringVarP(&gflags.RulesConfigFile, "config-rules", "r", gflags.RulesConfigFile, "rules config file")
	}
}

func main() {
	var (
		cmds []*cobra.Command
	)

	// Register sub commands
	cmds = append(cmds, runc)
	// sub command placeholder

	teleportc.AddCommand(cmds...)
	defer func() {
		teleportc.RemoveCommand(cmds...)
	}()

	teleportc.Execute()
}
