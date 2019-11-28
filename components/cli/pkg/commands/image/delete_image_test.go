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

package image

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"

	"cellery.io/cellery/components/cli/internal/test"
)

func TestDeleteImage(t *testing.T) {
	tests := []struct {
		name      string
		images    []string
		deleteAll bool
		regex     string
	}{
		{
			name:      "delete single image",
			images:    []string{"myorg/hello:1.0.0"},
			deleteAll: false,
			regex:     "",
		},
		{
			name:      "delete multiple images",
			images:    []string{"myorg/hello:1.0.0", "myorg/foo:1.0.0"},
			deleteAll: false,
			regex:     "",
		},
		{
			name:      "delete all images",
			images:    []string{},
			deleteAll: true,
			regex:     "",
		},
		{
			name:      "delete images with regex",
			images:    []string{},
			deleteAll: false,
			regex:     ".*/hello:.*",
		},
	}
	for _, testIteration := range tests {
		tempRepo, err := ioutil.TempDir("", "repo")
		if err != nil {
			t.Errorf("error creating temp repo, %v", err)
		}
		mockRepo := filepath.Join("testdata", "repo")
		if copyDir(mockRepo, tempRepo); err != nil {
			t.Errorf("error copying mock repo to temp repo, %v", err)
		}
		mockFileSystem := test.NewMockFileSystem(test.SetRepository(tempRepo))
		mockCli := test.NewMockCli(test.SetFileSystem(mockFileSystem))
		t.Run(testIteration.name, func(t *testing.T) {
			err := RunDeleteImage(mockCli, testIteration.images, testIteration.regex, testIteration.deleteAll)
			if err != nil {
				t.Errorf("error in RunDeleteImage, %v", err)
			}
		})
	}
}

// Dir copies a whole directory recursively
func copyDir(src string, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = copyDir(srcfp, dstfp); err != nil {
				return err
			}
		} else {
			if _, err = copyFile(srcfp, dstfp); err != nil {
				return err
			}
		}
	}
	return nil
}

// File copies a single file from src to dst
func copyFile(src, dst string) (*os.File, error) {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return nil, err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return nil, err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return nil, err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return nil, err
	}
	return dstfd, os.Chmod(dst, srcinfo.Mode())
}
