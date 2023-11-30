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
	_ "embed"
	"fmt"
	"net"
	"strings"

	"github.com/gobwas/glob"
	geoip2 "github.com/oschwald/geoip2-golang"
)

const (
	PatternChars = "*?[!]{},\\"
)

//go:embed resources/country-only-cn-private.mmdb
var mmdb []byte

type RuleRoute struct {
	Pattern  glob.Glob
	Hostname string
	IP       string
	Selector string
}

type RuleGroup struct {
	Group    string
	IP       string
	Selector string
}

type RuleCIDR struct {
	ipnet    *net.IPNet
	ip       string
	selector string
}

type Rules struct {
	r0    map[string]RuleRoute
	r1    map[string]RuleGroup
	rg    map[string]*RulesGroup
	rp    []RuleRoute
	db    *geoip2.Reader
	geoip map[string]string
	cidrs []RuleCIDR
	final string
}

// NewRules
func NewRules() (r *Rules) {
	r = &Rules{
		r0:    make(map[string]RuleRoute),
		r1:    make(map[string]RuleGroup),
		rg:    make(map[string]*RulesGroup),
		rp:    make([]RuleRoute, 0),
		geoip: make(map[string]string),
	}
	db, err := geoip2.FromBytes(mmdb)
	if err == nil {
		r.db = db
	}
	return
}

// Put
func (r *Rules) Put(hostname, ip, selector string) {
	if hostname == "**" {
		r.putFinal(selector)
		return
	}
	if strings.HasPrefix(hostname, "group:") {
		r.putGroup(hostname[6:], ip, selector)
		return
	}
	if strings.HasPrefix(hostname, "geoip:") {
		r.putGeoIP(hostname[6:], ip, selector)
		return
	}
	if strings.HasPrefix(hostname, "ip-cidr:") {
		r.putCIDR(hostname[8:], ip, selector)
		return
	}
	r.putHost(hostname, ip, selector)
}

// putHost
func (r *Rules) putHost(hostname, ip, selector string) {
	if !strings.ContainsAny(hostname, PatternChars) {
		if _, ok := r.r0[hostname]; !ok {
			r.r0[hostname] = RuleRoute{Hostname: hostname, IP: ip, Selector: selector}
		}
	} else {
		p, err := glob.Compile(hostname, '.')
		if err != nil {
			fmt.Printf("invalid pattern: %s\n", hostname)
			return
		}
		r.rp = append(r.rp, RuleRoute{Pattern: p, Hostname: hostname, IP: ip, Selector: selector})
	}
}

// putGroup
func (r *Rules) putGroup(name, ip, selector string) {
	if len(name) <= 0 {
		return
	}
	if _, ok := r.r1[name]; !ok {
		r.r1[name] = RuleGroup{Group: name, IP: ip, Selector: selector}
	}
}

// putGeoIP
func (r *Rules) putGeoIP(isoName, ip, selector string) {
	if len(isoName) <= 0 {
		return
	}
	if _, ok := r.geoip[isoName]; !ok {
		r.geoip[isoName] = selector
	}
}

// putCIDR
func (r *Rules) putCIDR(cidr, ip, selector string) {
	if len(cidr) <= 0 {
		return
	}
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return
	}
	r.cidrs = append(r.cidrs, RuleCIDR{ipnet: ipnet, ip: ip, selector: selector})
}

// putFinal
func (r *Rules) putFinal(final string) {
	r.final = final
}

// Group
func (r *Rules) Group(name string, hostnames ...string) {
	g, ok := r.rg[name]
	if !ok {
		r.rg[name] = NewRulesGroup()
		g = r.rg[name]
	}
	for _, hostname := range hostnames {
		if hostname == "**" {
			continue
		} else if !strings.ContainsAny(hostname, PatternChars) {
			if _, ok := g.r0[hostname]; !ok {
				g.r0[hostname] = struct{}{}
			}
		} else {
			p, err := glob.Compile(hostname, '.')
			if err != nil {
				fmt.Printf("invalid pattern: %s\n", hostname)
			} else {
				g.rp = append(g.rp, p)
			}
		}
	}
}

// Match
func (r *Rules) Match(hostname string) (ip, selector string, ok bool) {
	if route, ok := r.r0[hostname]; ok {
		return route.IP, route.Selector, true
	}
	for gn, rg := range r.rg {
		if rp, ok := r.r1[gn]; !ok {
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
	var IP net.IP
	if IP = net.ParseIP(hostname); IP == nil {
		IPs, err := net.LookupIP(hostname)
		if err == nil {
			for _, v := range IPs {
				IP = v
				break
			}
		}
	}
	if IP != nil {
		// geoip
		if r.db != nil && len(r.geoip) > 0 {
			if city, err := r.db.City(IP); err == nil {
				if selector, ok := r.geoip[strings.ToLower(city.Country.IsoCode)]; ok {
					return "", selector, true
				}
			}
		}
		// ip-cidr
		if len(r.cidrs) > 0 {
			for _, cidr := range r.cidrs {
				if cidr.ipnet.Contains(IP) {
					return cidr.ip, cidr.selector, true
				}
			}
		}
	}
	if r.final != "" {
		return "", r.final, true
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
