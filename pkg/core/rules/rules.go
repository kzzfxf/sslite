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
	_ "embed"
	"fmt"
	"net"
	"strings"

	"github.com/gobwas/glob"
	"github.com/kzzfxf/teleport/pkg/config"
	"github.com/kzzfxf/teleport/pkg/utils"
)

type selector struct {
	forward  string
	selector string
}

// empty
func (s selector) empty() bool {
	return s.selector == ""
}

type Rules struct {
	conf      *config.Rules
	hostnames map[string]hostname
	patterns  []pattern
	groups    map[string]*group
	geoips    map[string]geoip
	ipcidrs   []ipcidr
	final     selector
}

// NewRules
func NewRules(conf *config.Rules) (r *Rules) {
	r = &Rules{
		conf:      conf,
		hostnames: make(map[string]hostname),
		patterns:  make([]pattern, 0),
		groups:    make(map[string]*group),
		geoips:    make(map[string]geoip),
		ipcidrs:   make([]ipcidr, 0),
	}
	if conf != nil {
		r.init()
	}

	return
}

// init
func (r *Rules) init() {
	for _, route := range r.conf.Routes {
		prefix, rule, known := WhatRule(route.Rule)
		if !known || rule == "" {
			continue
		}
		selector := selector{selector: route.Selector}
		switch prefix {
		case "domain", "ipv4", "ipv6":
			if _, ok := r.hostnames[rule]; !ok {
				if utils.IsValidAddr(route.Forward) ||
					(utils.IsDomain(route.Forward) || utils.IsIPV4(route.Forward) || utils.IsIPV6(route.Forward)) {
					selector.forward = route.Forward
				}
				r.hostnames[rule] = hostname{
					hostname: rule,
					selector: selector,
				}
			}
		case "pattern":
			p, err := glob.Compile(rule, '.')
			if err == nil {
				if utils.IsValidAddr(route.Forward) ||
					(utils.IsDomain(route.Forward) || utils.IsIPV4(route.Forward) || utils.IsIPV6(route.Forward)) {
					selector.forward = route.Forward
				}
				r.patterns = append(r.patterns, pattern{
					pattern:  p,
					hostname: rule,
					selector: selector,
				})
			}
		case "geoip":
			if _, ok := r.geoips[rule]; !ok {
				r.geoips[rule] = geoip{
					isoCode:  rule,
					selector: selector,
				}
			}
		case "ip-cidr":
			_, ipnet, err := net.ParseCIDR(rule)
			if err == nil {
				r.ipcidrs = append(r.ipcidrs, ipcidr{
					ipnet:    ipnet,
					cidr:     rule,
					selector: selector,
				})
			}
		case "group":
			if _, ok := r.groups[rule]; !ok {
				r.groups[rule] = &group{
					name:      rule,
					hostnames: make(map[string]struct{}),
					patterns:  make([]glob.Glob, 0),
					selector:  selector,
				}
			}
		case "final":
			if r.final.empty() {
				r.final = selector
			}
		}
	}
	for _, group := range r.conf.Groups {
		g, ok := r.groups[group.Name]
		if !ok {
			continue
		}
		for _, hostname := range group.Hostnames {
			prefix, rule, known := WhatRule(hostname)
			if !known || rule == "" {
				continue
			}
			switch prefix {
			case "domain", "ipv4", "ipv6":
				if _, ok := g.hostnames[rule]; !ok {
					g.hostnames[rule] = struct{}{}
				}
			case "pattern":
				p, err := glob.Compile(rule, '.')
				if err == nil {
					g.patterns = append(g.patterns, p)
				}
			}
		}
	}
}

// Match
func (r *Rules) Match(hostname string) (selector, forward, matched string) {
	if route, ok := r.hostnames[hostname]; ok {
		return route.selector.selector, route.selector.forward, hostname
	}
	for _, p := range r.patterns {
		if p.pattern.Match(hostname) {
			return p.selector.selector, p.selector.forward, p.hostname
		}
	}
	for _, g := range r.groups {
		if g.match(hostname) {
			return g.selector.selector, g.selector.forward, fmt.Sprintf("group:%s", g.name)
		}
	}
	var IP net.IP
	if IP = net.ParseIP(hostname); IP == nil {
		IP = utils.LookupIP(hostname)
	}
	if IP != nil {
		// geoip
		if len(r.geoips) > 0 && geoipdb != nil {
			if isoCode, known := lookupGeoIPIsoCode(IP); known {
				if geoip, ok := r.geoips[isoCode]; ok {
					return geoip.selector.selector, geoip.selector.forward, fmt.Sprintf("group:%s", isoCode)
				}
			}
		}
		// ip-cidr
		for _, cidr := range r.ipcidrs {
			if cidr.ipnet.Contains(IP) {
				return cidr.selector.selector, cidr.selector.forward, fmt.Sprintf("group:%s", cidr.cidr)
			}
		}
	}
	if !r.final.empty() {
		return r.final.selector, r.final.forward, "**"
	}
	return
}

// WhatRule
func WhatRule(hostname string) (prefix, rule string, known bool) {
	if hostname == "**" {
		return "final", hostname, true
	}
	if strings.HasPrefix(hostname, "geoip:") {
		return "geoip", hostname[6:], true
	}
	if strings.HasPrefix(hostname, "ip-cidr:") {
		return "ip-cidr", hostname[8:], true
	}
	if strings.HasPrefix(hostname, "group:") {
		return "group", hostname[6:], true
	}
	if isPattern(hostname) {
		return "pattern", hostname, true
	}
	if utils.IsDomain(hostname) {
		return "domain", hostname, true
	}
	if utils.IsIPV4(hostname) {
		return "ipv4", hostname, true
	}
	if utils.IsIPV6(hostname) {
		return "ipv6", hostname, true
	}
	return "", "", false
}
