/*
 * Copyright (c) 2019 WSO2 Inc. (http:www.wso2.org) All Rights Reserved.
 *
 * WSO2 Inc. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http:www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package cli

import (
	"io"
	"os"
)

// Cli represents the cellery command line client.
type Cli interface {
	Out() io.Writer
}

// CelleryCli is an instance of the cellery command line client.
// Instances of the client can be returned from NewCelleryCli.
type CelleryCli struct {
}

// NewCelleryCli returns a CelleryCli instance.
func NewCelleryCli() *CelleryCli {
	cli := &CelleryCli{}
	return cli
}

// Out returns the writer used for the stdout.
func (cli *CelleryCli) Out() io.Writer {
	return os.Stdout
}
