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
 *
 */

package util

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
)

// File copies a single file from src to dst
func CopyFile(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

// Dir copies a whole directory recursively
func CopyDir(src string, dst string) error {
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
			if err = CopyDir(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = CopyFile(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

func RecursiveZip(files []string, folders []string, destinationPath string) error {
	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	myZip := zip.NewWriter(destinationFile)
	for _, folder := range folders {
		err = filepath.Walk(folder, func(filePath string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			if err != nil {
				return err
			}
			relPath := strings.TrimPrefix(filePath, filepath.Dir(folder))

			zipFile, err := myZip.Create(relPath)
			if err != nil {
				return err
			}

			fsFile, err := os.Open(filePath)
			if err != nil {
				return err
			}

			_, err = io.Copy(zipFile, fsFile)

			if err != nil {
				return err
			}
			return nil
		})
	}

	// Copy files
	for _, file := range files {
		zipFile, err := myZip.Create(file)
		files, err := os.Open(file)
		if err != nil {
			return err
		}
		_, err = io.Copy(zipFile, files)
	}
	if err != nil {
		return err
	}
	err = myZip.Close()
	if err != nil {
		return err
	}
	return nil
}

func Unzip(zipFolderName string, destinationFolderName string) error {
	var fileNames []string
	zipFolder, err := zip.OpenReader(zipFolderName)
	if err != nil {
		return err
	}
	defer zipFolder.Close()

	for _, file := range zipFolder.File {
		fileContent, err := file.Open()
		if err != nil {
			return err
		}
		defer fileContent.Close()

		fpath := filepath.Join(destinationFolderName, file.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(destinationFolderName)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", fpath)
		}

		fileNames = append(fileNames, fpath)
		if file.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
		} else {
			// Make File
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return err
			}
			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
			if err != nil {
				return err
			}
			_, err = io.Copy(outFile, fileContent)

			// Close the file without defer to close before next iteration of loop
			outFile.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func FindInDirectory(directory, suffix string) []string {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil
	}
	fileList := []string{}
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), suffix) {
			fileList = append(fileList, filepath.Join(directory, f.Name()))
		}
	}
	return fileList
}

func CreateDir(dirPath string) error {
	dirExist, _ := FileExists(dirPath)
	if !dirExist {
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func RemoveDir(dirPath string) error {
	dirExist, _ := FileExists(dirPath)
	if dirExist {
		err := os.RemoveAll(dirPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func CleanOrCreateDir(dirPath string) error {
	dirExist, _ := FileExists(dirPath)
	if !dirExist {
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return err
		}
	} else {
		err := os.RemoveAll(dirPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func CleanAndCreateDir(dirPath string) error {
	dirExist, _ := FileExists(dirPath)
	if dirExist {
		err := os.RemoveAll(dirPath)
		if err != nil {
			return err
		}
	}
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func GetSubDirectoryNames(path string) ([]string, error) {
	directoryNames := []string{}
	subdirectories, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, subdirectory := range subdirectories {
		if subdirectory.IsDir() {
			directoryNames = append(directoryNames, subdirectory.Name())
		}
	}
	return directoryNames, nil
}

func GetFileSize(path string) (int64, error) {
	file, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	return file.Size(), nil
}

func RenameFile(oldName, newName string) error {
	err := os.Rename(oldName, newName)
	if err != nil {
		return err
	}
	return nil
}

func ExtractTarGzFile(extractTo, archive_name string) error {
	cmd := exec.Command("tar", "-zxvf", archive_name)
	cmd.Dir = extractTo
	err := ExecuteCommand(cmd)
	if err != nil {
		return err
	}
	return nil
}

func ReplaceInFile(srcFile, oldString, newString string, replaceCount int) error {
	input, err := ioutil.ReadFile(srcFile)
	if err != nil {
		return err
	}

	output := bytes.Replace(input, []byte(oldString), []byte(newString), replaceCount)

	if err = ioutil.WriteFile(srcFile, output, 0666); err != nil {
		return err
	}
	return nil
}

func GetSourceFileName(filePath string) (string, error) {
	d, err := os.Open(filePath)
	if err != nil {
		ExitWithErrorMessage("Error opening file "+filePath, err)
	}
	defer d.Close()
	fi, err := d.Readdir(-1)
	if err != nil {
		ExitWithErrorMessage("Error reading file "+filePath, err)
	}
	for _, fi := range fi {
		if fi.Mode().IsRegular() && strings.HasSuffix(fi.Name(), ".bal") {
			return fi.Name(), nil
		}
	}
	return "", errors.New("Ballerina source file not found in extracted location: " + filePath)
}

func CopyK8sArtifacts(outPath string) {
	k8sArtifactsDir := filepath.Join(outPath, constants.K8sArtifacts)
	RemoveDir(k8sArtifactsDir)
	CreateDir(k8sArtifactsDir)
	CopyDir(filepath.Join(CelleryInstallationDir(), constants.K8sArtifacts), k8sArtifactsDir)
}

func CreateCelleryDirStructure() {
	celleryHome := filepath.Join(UserHomeDir(), constants.CelleryHome)
	CreateDir(celleryHome)
	CreateDir(filepath.Join(celleryHome, "gcp"))
	CreateDir(filepath.Join(celleryHome, "k8s-artefacts"))
	CreateDir(filepath.Join(celleryHome, "logs"))
	CreateDir(filepath.Join(celleryHome, "repo"))
	CreateDir(filepath.Join(celleryHome, "tmp"))
	CreateDir(filepath.Join(celleryHome, "vm"))
}
