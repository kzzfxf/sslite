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

package service

import (
	"encoding/json"

	"github.com/kzzfxf/sslite/pkg/config"
)

type configuration interface {
	LoadConfig(data []byte) (conf *config.Config, err error)
	LoadRules(data []byte) (conf *config.Rules, err error)
}

type configurationImpl struct {
}

var Config configuration = &configurationImpl{}

// LoadConfig
func (c *configurationImpl) LoadConfig(data []byte) (conf *config.Config, err error) {
	var config config.Config
	err = json.Unmarshal(data, &config)
	if err == nil {
		conf = &config
	}
	return
}

// LoadRules
func (c *configurationImpl) LoadRules(data []byte) (conf *config.Rules, err error) {
	var config config.Rules
	err = json.Unmarshal(data, &config)
	if err == nil {
		conf = &config
	}
	return
}
