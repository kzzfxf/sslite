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

package utils

import (
	"fmt"
	"net"
	"strconv"
)

// SetKeepAlive
func SetKeepAlive(conn net.Conn) {
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
	}
}

// ParseAddr
func ParseAddr(addr string) (domain, ip string, port uint, err error) {
	h, p, err := net.SplitHostPort(addr)
	if err != nil {
		return
	}
	if IP := net.ParseIP(h); IP != nil {
		ip = h
	} else if len(h) <= 255 {
		domain = h
	} else {
		return "", "", 0, fmt.Errorf("invalid host: %s", h)
	}
	nport, err := strconv.ParseUint(p, 10, 16)
	if err != nil {
		return
	}
	return domain, ip, uint(nport), nil
}
