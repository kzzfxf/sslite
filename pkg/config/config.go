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

package config

type Config struct {
	Global  string  `json:"global"`
	Latency Latency `json:"latency"`
	Proxies []Proxy `json:"proxies"`
}

type Proxy struct {
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	URL    string   `json:"url"`
	Labels []string `json:"labels"`
}

type Latency struct {
	URL     string `json:"url"`
	Timeout int64  `json:"timeout"`
}

type Rules struct {
	Routes []Route `json:"routes"`
	Groups []Group `json:"groups"`
}

type Route struct {
	Rule     string `json:"rule"`
	Forward  string `json:"forward"`
	Selector string `json:"selector"`
}

type Group struct {
	Name      string   `json:"name"`
	Hostnames []string `json:"hostnames"`
}
