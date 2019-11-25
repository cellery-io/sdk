/*
 * Copyright (c) 2019 WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
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

package credentials

import (
	"regexp"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
)

type CredReader interface {
	SetRegistry(registry string)
	SetUserName(username string)
	Read() (string, string, error)
	Shutdown(authorized bool)
}

type CelleryCredReader struct {
	registry     string
	userName     string
	isAuthorized chan bool
	done         chan bool
}

// NewCelleryCredReader returns a CelleryCredReader instance.
func NewCelleryCredReader(opts ...func(*CelleryCredReader)) *CelleryCredReader {
	reader := &CelleryCredReader{}
	for _, opt := range opts {
		opt(reader)
	}
	return reader
}

func (celleryCredReader *CelleryCredReader) Read() (string, string, error) {
	regex, err := regexp.Compile(constants.CentralRegistryHostRegx)
	if err != nil {
		return "", "", err
	}
	if celleryCredReader.userName == "" && regex.MatchString(celleryCredReader.registry) {
		celleryCredReader.isAuthorized = make(chan bool)
		celleryCredReader.done = make(chan bool)
		return FromBrowser(celleryCredReader.userName, celleryCredReader.isAuthorized, celleryCredReader.done)
	} else {
		return FromTerminal(celleryCredReader.userName)
	}
}

func (celleryCredReader *CelleryCredReader) SetRegistry(registry string) {
	celleryCredReader.registry = registry
}

func (celleryCredReader *CelleryCredReader) SetUserName(username string) {
	celleryCredReader.userName = username
}

func (celleryCredReader *CelleryCredReader) Shutdown(authorized bool) {
	if celleryCredReader.isAuthorized != nil {
		celleryCredReader.isAuthorized <- authorized
	}
	if celleryCredReader.done != nil {
		<-celleryCredReader.done
	}
}
