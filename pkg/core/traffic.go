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

package core

import (
	"net"
)

type ConnTrafficTracker struct {
	net.Conn
	down, up chan<- int
}

func (t *ConnTrafficTracker) Read(b []byte) (n int, err error) {
	defer func() {
		if n > 0 {
			t.down <- n
		}
	}()
	return t.Conn.Read(b)
}

func (t *ConnTrafficTracker) Write(b []byte) (n int, err error) {
	defer func() {
		if n > 0 {
			t.up <- n
		}
	}()
	return t.Conn.Write(b)
}
