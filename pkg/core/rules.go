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
)

const (
	PatternChars = "*?[!]{},\\"
)

type RulePattern struct {
	Pattern  glob.Glob
	Hostname string
	IP       string
	Selector string
}

type Rules struct {
	r0 map[string]RulePattern
	rg map[string]*RulesGroup
	rp []RulePattern
}

// NewRules
func NewRules() (r *Rules) {
	return &Rules{
		r0: make(map[string]RulePattern),
		rg: make(map[string]*RulesGroup),
		rp: make([]RulePattern, 0),
	}
}

// Put
func (r *Rules) Put(hostname, ip, selector string) {
	if !strings.ContainsAny(hostname, PatternChars) {
		if _, ok := r.r0[hostname]; !ok {
			r.r0[hostname] = RulePattern{Hostname: hostname, IP: ip, Selector: selector}
		}
	} else {
		p, err := glob.Compile(hostname, '.')
		if err != nil {
			fmt.Printf("invalid pattern: %s\n", hostname)
			return
		}
		r.rp = append(r.rp, RulePattern{Pattern: p, Hostname: hostname, IP: ip, Selector: selector})
	}
}

// Group
func (r *Rules) Group(name string, hostnames ...string) {
	if len(hostnames) <= 0 {
		return
	}
	if _, ok := r.rg[name]; !ok {
		r.rg[name] = NewRulesGroup()
	}
	for _, hostname := range hostnames {
		if !strings.ContainsAny(hostname, PatternChars) {
			if _, ok := r.rg[name].r0[hostname]; !ok {
				r.rg[name].r0[hostname] = struct{}{}
			}
		} else {
			p, err := glob.Compile(hostname, '.')
			if err != nil {
				fmt.Printf("invalid pattern: %s\n", hostname)
			} else {
				r.rg[name].rp = append(r.rg[name].rp, p)
			}
		}
	}
}

// Match
func (r *Rules) Match(hostname string) (ip, selector string, ok bool) {
	if route, ok := r.r0[hostname]; ok {
		return route.IP, route.Selector, ok
	}
	for gn, rg := range r.rg {
		if rp, ok := r.r0[gn]; !ok {
			continue
		} else {
			if rg.Match(hostname) {
				return rp.IP, rp.Selector, true
			}
		}
	}
	for _, p := range r.rp {
		if p.Pattern.Match(hostname) {
			return p.IP, p.Selector, true
		}
	}
	return
}

type RulesGroup struct {
	r0 map[string]struct{}
	rp []glob.Glob
}

// NewRulesGroup
func NewRulesGroup() (rg *RulesGroup) {
	return &RulesGroup{
		r0: make(map[string]struct{}),
		rp: make([]glob.Glob, 0),
	}
}

// Match
func (rg *RulesGroup) Match(hostname string) (matched bool) {
	if _, ok := rg.r0[hostname]; ok {
		return true
	}
	for _, p := range rg.rp {
		if p.Match(hostname) {
			return true
		}
	}
	return
}
