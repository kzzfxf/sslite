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
	"os"

	"github.com/kzzfxf/teleport/pkg/logkit"
)

type GlobalFlags struct {
	LogLevel        string
	BaseConfigFile  string
	RulesConfigFile string
}

func NewGlobalFlags() (gflags *GlobalFlags) {
	gflags = &GlobalFlags{}
	gflags.LogLevel = string(logkit.LevelError)
	gflags.BaseConfigFile = "./conf/teleport.json"
	gflags.RulesConfigFile = "./conf/rules.json"
	return
}

type TeleportFlags struct {
	*GlobalFlags
	// Test string
}

func NewTeleportFlags(gflags *GlobalFlags) (flags *TeleportFlags) {
	flags = &TeleportFlags{GlobalFlags: gflags}
	return
}

func OnTeleportHandler(ctx context.Context, flags *TeleportFlags, args []string) (err error) {
	return
}

func OnGlobalBeforeHandler(ctx context.Context, flags *GlobalFlags, args []string) (err error) {
	logkit.Init(os.Stdout, logkit.Level(flags.LogLevel))
	logkit.Info("init logkit")
	return
}

func OnGlobalAfterHandler(ctx context.Context, flags *GlobalFlags, args []string) (err error) {
	return
}
