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
	"path/filepath"
	"runtime"
)

const celleryHome = ".cellery"

type FileSystemManager interface {
	CurrentDir() string
	UserHome() (string, error)
	Repository() string
	RemoveAll(path string) error
}

type celleyFileSystem struct {
	currentDir string
}

// NewCelleryFileSystem returns a celleyFileSystem instance.
func NewCelleryFileSystem() (*celleyFileSystem, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	fs := &celleyFileSystem{
		currentDir: currentDir,
	}
	return fs, nil
}

// CurrentDir returns the current directory.
func (fs *celleyFileSystem) CurrentDir() string {
	return fs.currentDir
}

// UserHome returns user home.
func (fs *celleyFileSystem) UserHome() (string, error) {
	return userHomeDir(), nil
}

// UserHome returns user home.
func (fs *celleyFileSystem) Repository() string {
	return filepath.Join(userHomeDir(), celleryHome, "repo")
}

// RemoveAll deletes files in a given location.
func (fs *celleyFileSystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
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
