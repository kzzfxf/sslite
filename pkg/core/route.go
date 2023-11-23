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
	"fmt"
	"strings"

	"github.com/gobwas/glob"
	"github.com/kzzfxf/teleport/pkg/config"
)

func init() {
	glob.MustCompile("", '.')
}

type pattern struct {
	g     glob.Glob
	route config.Route
}

type Route struct {
	t0 map[string]config.Route
	t1 []pattern
}

// NewRoute
func NewRoute() (r *Route) {
	return &Route{
		t0: make(map[string]config.Route),
		t1: make([]pattern, 0),
	}
}

// Init
func (r *Route) Init(routes []config.Route) (err error) {
	for _, route := range routes {
		if strings.ContainsAny(route.Server, "*?[!]{},\\") {
			g, err := glob.Compile(route.Server, '.')
			if err != nil {
				return err
			}
			r.t1 = append(r.t1, pattern{g: g, route: route})
		} else {
			if _, ok := r.t0[route.Server]; ok {
				return fmt.Errorf("duplicate item: %s", route.Server)
			}
			r.t0[route.Server] = route
		}
	}
	return
}

// Match
func (r *Route) Match(server string) (ip, selector string, ok bool) {
	if !strings.ContainsAny(server, "*?[!]{},\\") {
		if route, ok := r.t0[server]; ok {
			return route.IP, route.Selector, ok
		}
	}
	for _, p := range r.t1 {
		if p.g.Match(server) {
			return p.route.IP, p.route.Selector, true
		}
	}
	return
}
