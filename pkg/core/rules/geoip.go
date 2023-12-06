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
	"net"
	"strings"

	"github.com/kzzfxf/sslite/pkg/logkit"
	"github.com/oschwald/geoip2-golang"
)

type geoip struct {
	isoCode  string
	selector selector
}

var (
	geoipdb *geoip2.Reader
	//go:embed country-only-cn-private.mmdb
	geoipmmdb []byte
)

func init() {
	db, err := geoip2.FromBytes(geoipmmdb)
	if err == nil {
		geoipdb = db
	} else {
		logkit.Error("read geoip database failed", logkit.Any("error", err))
	}
}

// lookupGeoIPIsoCode
func lookupGeoIPIsoCode(ip net.IP) (code string, known bool) {
	if geoipdb == nil {
		return
	}
	city, err := geoipdb.City(ip)
	if err != nil {
		return "", false
	}
	return strings.ToLower(city.Country.IsoCode), true
}
