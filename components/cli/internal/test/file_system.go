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

package test

type MockFileSystem struct {
	currentDir             string
	repository             string
	userHome               string
	tempDir                string
	celleryInstallationDir string
}

// NewMockFileSystem returns a mockFileSystem instance.
func NewMockFileSystem(opts ...func(*MockFileSystem)) *MockFileSystem {
	fs := &MockFileSystem{}
	for _, opt := range opts {
		opt(fs)
	}
	return fs
}

func SetRepository(repository string) func(*MockFileSystem) {
	return func(fs *MockFileSystem) {
		fs.repository = repository
	}
}

func SetCurrentDir(currentDir string) func(*MockFileSystem) {
	return func(fs *MockFileSystem) {
		fs.currentDir = currentDir
	}
}

func SetCelleryInstallationDir(dir string) func(*MockFileSystem) {
	return func(fs *MockFileSystem) {
		fs.celleryInstallationDir = dir
	}
}

// CurrentDir returns the current directory.
func (fs *MockFileSystem) CurrentDir() string {
	return fs.currentDir
}

// TempDir returns the temp directory.
func (fs *MockFileSystem) TempDir() string {
	return fs.tempDir
}

// UserHome returns the user home.
func (fs *MockFileSystem) UserHome() string {
	return fs.userHome
}

// UserHome returns user home.
func (fs *MockFileSystem) Repository() string {
	return fs.repository
}

func (fs *MockFileSystem) CelleryInstallationDir() string {
	return fs.celleryInstallationDir
}

// WorkingDirRelativePath returns the relative path of working directory.
// For mock file system, WorkingDirRelativePath is also the current dir path.
func (fs *MockFileSystem) WorkingDirRelativePath() string {
	return fs.currentDir
}
