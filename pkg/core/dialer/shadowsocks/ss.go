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

package shadowsocks

import (
	"errors"
	"net"

	"github.com/kzzfxf/teleport/pkg/utils"
	"github.com/riobard/go-shadowsocks2/core"
	"github.com/riobard/go-shadowsocks2/socks"
)

var (
	ErrInvalidAddress       = errors.New("invalid address")
	ErrProtocolNotSupported = errors.New("protocol not supported")
)

type ShadowSocks struct {
	node string
	ciph core.Cipher
}

// NewShadowsocks
func NewShadowsocks(node, cipher, password string) (ss *ShadowSocks, err error) {
	ciph, err := core.PickCipher(cipher, nil, password)
	if err != nil {
		return
	}
	ss = &ShadowSocks{
		node: node,
		ciph: ciph,
	}
	return
}

// Dial
func (ss *ShadowSocks) Dial(network, addr string) (conn net.Conn, err error) {
	if network != "tcp" {
		return nil, ErrProtocolNotSupported
	}

	target := socks.ParseAddr(addr)
	if target == nil {
		return nil, ErrInvalidAddress
	}

	conn, err = net.Dial(network, ss.node)
	if err != nil {
		return
	}

	// Keepalive
	utils.SetKeepAlive(conn)
	// Stream conn
	conn = ss.ciph.StreamConn(conn)
	// Write target
	_, err = conn.Write(target)
	if err != nil {
		return
	}

	return
}

// Close
func (ss *ShadowSocks) Close() (err error) {
	return
}
