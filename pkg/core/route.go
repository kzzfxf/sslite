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
	"strings"
	"sync"
	"time"

	"github.com/gobwas/glob"
)

type RouteItem struct {
	Server  string
	IP      string
	Tunnel  *Tunnel
	Timeout time.Time
}

type Route struct {
	rt     map[string]RouteItem
	locker sync.RWMutex
}

// NewRoute
func NewRoute() (r *Route) {
	return &Route{
		rt: make(map[string]RouteItem),
	}
}

// Put
func (r *Route) Put(server, ip string, tunnel *Tunnel, timeout time.Time) {
	r.locker.Lock()
	defer r.locker.Unlock()
	r.rt[server] = RouteItem{Server: server, IP: ip, Tunnel: tunnel, Timeout: timeout}
}

// Get
func (r *Route) Get(server string) (ip string, tunnel *Tunnel, exists bool) {
	r.locker.RLock()
	item, ok := r.rt[server]
	r.locker.RUnlock()
	if !ok {
		return "", nil, false
	} else if item.Timeout.Before(time.Now()) {
		r.locker.Lock()
		defer r.locker.Unlock()
		delete(r.rt, server)
		return "", nil, false
	}
	return item.IP, item.Tunnel, true
}

type RoutePattern struct {
	Pattern  glob.Glob
	Server   string
	IP       string
	Selector string
}

type RouteRules struct {
	t0 map[string]RoutePattern
	tg map[string]*RouteGroup
	tp []RoutePattern
}

// NewRouteRules
func NewRouteRules() (r *RouteRules) {
	return &RouteRules{
		t0: make(map[string]RoutePattern),
		tg: make(map[string]*RouteGroup),
		tp: make([]RoutePattern, 0),
	}
}

// Put
func (r *RouteRules) Put(server, ip, selector string) {
	if !strings.ContainsAny(server, "*?[!]{},\\") {
		r.t0[server] = RoutePattern{Server: server, IP: ip, Selector: selector}
	} else {
		p, err := glob.Compile(server, '.')
		if err != nil {
			return
		}
		r.tp = append(r.tp, RoutePattern{Pattern: p, Server: server, IP: ip, Selector: selector})
	}
}

// Group
func (r *RouteRules) Group(name string, server string, others ...string) {
	if _, ok := r.tg[name]; !ok {
		r.tg[name] = NewRouteGroup()
	}
	for _, server := range append(others, server) {
		if !strings.ContainsAny(server, "*?[!]{},\\") {
			r.tg[name].t0[server] = struct{}{}
		} else {
			p, err := glob.Compile(server, '.')
			if err != nil {
				return
			}
			r.tg[name].tp = append(r.tg[name].tp, p)
		}
	}
}

// Match
func (r *RouteRules) Match(server string) (ip, selector string, ok bool) {
	if route, ok := r.t0[server]; ok {
		return route.IP, route.Selector, ok
	}
	for _, p := range r.tp {
		if p.Pattern.Match(server) {
			return p.IP, p.Selector, true
		}
	}
	for gn, rg := range r.tg {
		if rp, ok := r.t0[gn]; !ok {
			continue
		} else {
			if rg.Match(server) {
				return rp.IP, rp.Selector, true
			}
		}
	}
	return
}

type RouteGroup struct {
	t0 map[string]struct{}
	tp []glob.Glob
}

// NewRouteGroup
func NewRouteGroup() (rg *RouteGroup) {
	return &RouteGroup{
		t0: make(map[string]struct{}),
		tp: make([]glob.Glob, 0),
	}
}

// Match
func (rg *RouteGroup) Match(server string) (matched bool) {
	if _, ok := rg.t0[server]; ok {
		return true
	}
	for _, p := range rg.tp {
		if p.Match(server) {
			return true
		}
	}
	return
}
