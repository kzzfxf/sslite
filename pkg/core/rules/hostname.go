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

package rules

import (
	"net"
	"regexp"
	"strings"

	"github.com/gobwas/glob"
)

type hostname struct {
	hostname string
	selector selector
}

type pattern struct {
	pattern  glob.Glob
	selector selector
}

var (
	domainRegExp = regexp.MustCompile(`^([a-zA-Z0-9]+(-[a-zA-Z0-9]+)*\.)+[a-zA-Z]{2,}$`)
)

// isIPV4
func isIPV4(s string) bool {
	if ip := net.ParseIP(s); ip != nil {
		return ip.To4() != nil
	}
	return false
}

// isIPV6
func isIPV6(s string) bool {
	if ip := net.ParseIP(s); ip != nil {
		return ip.To4() == nil
	}
	return false
}

// isDomain
func isDomain(s string) bool {
	return domainRegExp.MatchString(s)
}

// isPattern
func isPattern(s string) bool {
	return strings.ContainsAny(s, "*?[!]{},\\")
}
