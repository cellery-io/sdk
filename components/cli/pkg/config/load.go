/*
 * Copyright (c) 2019, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
 *
 * WSO2 Inc. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package config

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

const configFile = "config.json"

const defaultHubUrl = "https://hub.cellery.io"
const defaultIdpUrl = "https://id.choreo.dev"
const defaultClientId = "s8jIVx9uJKE087FosgcSwNVjGd0a"

type Conf struct {
	Hub *HubConf `json:"hub"`
	Idp *IdpConf `json:"idp"`
}

type HubConf struct {
	Url string `json:"url"`
}

type IdpConf struct {
	Url      string `json:"url"`
	ClientId string `json:"clientId"`
}

// LoadConfig reads the config file from the Cellery home and returns the Config struct
func LoadConfig() *Conf {
	// Default config
	var conf = &Conf{
		Hub: &HubConf{
			Url: defaultHubUrl,
		},
		Idp: &IdpConf{
			Url:      defaultIdpUrl,
			ClientId: defaultClientId,
		},
	}

	configFilePath := filepath.Join(util.UserHomeDir(), constants.CelleryHome, configFile)
	configFileBytes, err := ioutil.ReadFile(configFilePath)
	if err == nil {
		err = json.Unmarshal(configFileBytes, conf)
	}
	return conf
}
