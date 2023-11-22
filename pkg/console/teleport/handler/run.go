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

	"github.com/kzzfxf/teleport/pkg/port/http"
)

type RunFlags struct {
	*GlobalFlags
	// Test string
}

func NewRunFlags(gflags *GlobalFlags) (flags *RunFlags) {
	flags = &RunFlags{GlobalFlags: gflags}
	return
}

func OnRunHandler(ctx context.Context, flags *RunFlags, args []string) (err error) {
	return http.Start(ctx, ":8999")
}
