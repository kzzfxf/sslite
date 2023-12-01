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
	"errors"
	"net"
	"regexp"
	"strconv"
)

// SetKeepAlive
func SetKeepAlive(conn net.Conn) {
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
	}
}

// ParseAddr
func ParseAddr(addr string) (hostname string, port uint, err error) {
	hostname, p, err := net.SplitHostPort(addr)
	if err != nil {
		return
	}
	if hostname == "" {
		return "", 0, errors.New("invalid hostname")
	}
	nport, err := strconv.ParseUint(p, 10, 16)
	if err != nil {
		return
	}
	return hostname, uint(nport), nil
}

var (
	domainRegExp = regexp.MustCompile(`^([a-zA-Z0-9]+(-[a-zA-Z0-9]+)*\.)+[a-zA-Z]{2,}$`)
)

// IsIPV4
func IsIPV4(s string) bool {
	if ip := net.ParseIP(s); ip != nil {
		return ip.To4() != nil
	}
	return false
}

// IsIPV6
func IsIPV6(s string) bool {
	if ip := net.ParseIP(s); ip != nil {
		return ip.To4() == nil
	}
	return false
}

// IsDomain
func IsDomain(s string) bool {
	return domainRegExp.MatchString(s)
}

// IsValidAddr
func IsValidAddr(addr string) bool {
	_, _, err := ParseAddr(addr)
	return err == nil
}

// LookupIP
func LookupIP(domain string) net.IP {
	IPs, err := net.LookupIP(domain)
	if err == nil {
		for _, v := range IPs {
			return v
		}
	}
	return nil
}
