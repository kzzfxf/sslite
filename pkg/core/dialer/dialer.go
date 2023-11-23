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

package dialer

import (
	"fmt"
	"net"

	"github.com/kzzfxf/teleport/pkg/core/dialer/shadowsocks"
)

type Dialer interface {

	// Addr returns the dialer address.
	Addr() (addr string)

	// Dial returns a new connection by the dialer.
	Dial(network, addr string) (conn net.Conn, err error)

	// Close close the dialer.
	Close() (err error)
}

var (
	dialers = make(map[string]func(URL string) (dialer Dialer, err error))
)

func init() {
	dialers["ss"] = func(URL string) (dialer Dialer, err error) {
		return shadowsocks.NewShadowsocksWithURL(URL)
	}
}

// NewDialerWithURL
func NewDialerWithURL(t, URL string) (dialer Dialer, err error) {
	if fn, ok := dialers[t]; ok {
		return fn(URL)
	}
	return nil, fmt.Errorf("unrecognized proxy type '%s'", t)
}
