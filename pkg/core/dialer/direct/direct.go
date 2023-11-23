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

package direct

import (
	"net"
	"time"
)

type Direct struct {
	timeout time.Duration
}

// NewDirect
func NewDirect(timeout time.Duration) (d *Direct) {
	d = &Direct{
		timeout: timeout,
	}
	if d.timeout < 0 {
		d.timeout = 0
	}
	return
}

// Dial
func (d *Direct) Dial(network string, addr string) (conn net.Conn, err error) {
	return net.DialTimeout(network, addr, d.timeout)
}

// Close
func (d *Direct) Close() (err error) {
	return
}
