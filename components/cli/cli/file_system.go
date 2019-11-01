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
	"os"
	"runtime"
)

type FileSystemManager interface {
	CurrentDir() (string, error)
	UserHome() (string, error)
}

type celleyFileSystem struct {
}

// NewCelleryFileSystem returns a celleyFileSystem instance.
func NewCelleryFileSystem() *celleyFileSystem {
	fs := &celleyFileSystem{}
	return fs
}

// CurrentDir returns the current directory.
func (fs *celleyFileSystem) CurrentDir() (string, error) {
	return os.Getwd()
}

// UserHome returns user home.
func (fs *celleyFileSystem) UserHome() (string, error) {
	return userHomeDir(), nil
}

func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}
