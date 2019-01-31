/*
 * Copyright (c) 2018 WSO2 Inc. (http:www.wso2.org) All Rights Reserved.
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

package util

import (
	"archive/zip"
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/fatih/color"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var Cyan = color.New(color.FgCyan)
var CyanF = color.New(color.FgCyan).SprintFunc()
var CyanBold = Cyan.Add(color.Bold).SprintFunc()
var Bold = color.New(color.Bold).SprintFunc()
var Faint = color.New(color.Faint).SprintFunc()
var Green = color.New(color.FgGreen)
var GreenBold = Green.Add(color.Bold).SprintFunc()
var Yellow = color.New(color.FgYellow)
var YellowBold = Yellow.Add(color.Bold).SprintFunc()

func ExitWithImageFormatError() {
	fmt.Printf("Incorrect tag name. Tag name should be [REPOSITORY]/ORGANIZATION/IMAGE_NAME:VERSION\n")
	os.Exit(1)
}

func PrintWhatsNextMessage(cmd string) {
	fmt.Println()
	fmt.Println(Bold("What's next?"))
	fmt.Println("--------------------------------------------------------")
	fmt.Println("Execute the following command to run the image: ")
	fmt.Println("  $ " + cmd)
	fmt.Println("--------------------------------------------------------")
}

//func CopyFile(oldFile string, newFile string) {
//	input, err := os.Open(oldFile)
//	if err != nil {
//		panic(err)
//	}
//	defer input.Close()
//
//	output, err := os.Create(newFile)
//	if err != nil {
//		panic(err)
//	}
//	defer output.Close()
//
//	_, err = io.Copy(output, input)
//	if err != nil {
//		panic(err)
//	}
//}

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

func Trim(stream string) string {
	var trimmedString string
	if strings.Contains(stream, ".cell.balx") {
		trimmedString = strings.Replace(stream, ".cell.balx", ".celx", -1)
	} else if strings.Contains(stream, ".bal") {
		trimmedString = strings.Replace(stream, ".bal", "", -1)
	} else if strings.Contains(stream, ".cell") {
		trimmedString = strings.Replace(stream, ".cell", "", -1)
	} else {
		trimmedString = stream
	}
	return trimmedString
}

func ZipFiles(filename string, files []string) error {

	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {

		zipfile, err := os.Open(file)
		if err != nil {
			return err
		}
		defer zipfile.Close()

		// Get the file information
		info, err := zipfile.Stat()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Using FileInfoHeader() above only uses the basename of the file. If we want
		// to preserve the folder structure we can overwrite this with the full path.
		header.Name = file

		// Change to deflate to gain better compression
		// see http://golang.org/pkg/archive/zip/#pkg-constants
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		if _, err = io.Copy(writer, zipfile); err != nil {
			return err
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

func GetDuration(startTime time.Time) string {
	duration := ""
	var year, month, day, hour, min, sec int
	currentTime := time.Now()
	if startTime.Location() != currentTime.Location() {
		currentTime = currentTime.In(startTime.Location())
	}
	if startTime.After(currentTime) {
		startTime, currentTime = currentTime, startTime
	}
	startYear, startMonth, startDay := startTime.Date()
	currentYear, currentMonth, currentDay := currentTime.Date()

	startHour, startMinute, startSecond := startTime.Clock()
	currentHour, currentMinute, currentSecond := currentTime.Clock()

	year = int(currentYear - startYear)
	month = int(currentMonth - startMonth)
	day = int(currentDay - startDay)
	hour = int(currentHour - startHour)
	min = int(currentMinute - startMinute)
	sec = int(currentSecond - startSecond)

	// Normalize negative values
	if sec < 0 {
		sec += 60
		min--
	}
	if min < 0 {
		min += 60
		hour--
	}
	if hour < 0 {
		hour += 24
		day--
	}
	if day < 0 {
		// days in month:
		t := time.Date(startYear, startMonth, 32, 0, 0, 0, 0, time.UTC)
		day += 32 - t.Day()
		month--
	}
	if month < 0 {
		month += 12
		year--
	}

	numOfTimeUnits := 0
	if year > 0 && numOfTimeUnits < 2 {
		duration += strconv.Itoa(year) + " years "
		numOfTimeUnits++
	}
	if month > 0 && numOfTimeUnits < 2 {
		duration += strconv.Itoa(month) + " months "
		numOfTimeUnits++
	}
	if day > 0 && numOfTimeUnits < 2 {
		duration += strconv.Itoa(day) + " days "
		numOfTimeUnits++
	}
	if hour > 0 && numOfTimeUnits < 2 {
		duration += strconv.Itoa(hour) + " hours "
		numOfTimeUnits++
	}
	if min > 0 && numOfTimeUnits < 2 {
		duration += strconv.Itoa(min) + " minutes "
		numOfTimeUnits++
	}
	if sec > 0 && numOfTimeUnits < 2 {
		duration += strconv.Itoa(sec) + " seconds"
		numOfTimeUnits++
	}
	return duration
}

func ConvertStringToTime(timeString string) time.Time {
	convertedTime, err := time.Parse(time.RFC3339, timeString)
	if err != nil {
		fmt.Printf("\x1b[31;1m Error parsing time: \x1b[0m %v \n", err)
		os.Exit(1)
	}
	return convertedTime
}

// Creates a new file upload http request with optional extra params
func FileUploadRequest(uri string, params map[string]string, paramName, path string, secure bool) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	if !secure {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	req, err := http.NewRequest("POST", uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}

func DownloadFile(filepath string, url string) (*http.Response, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transport}

	resp, err := client.Get(url)
	if err != nil {
		return resp, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		out, err := os.Create(filepath)
		if err != nil {
			return nil, err
		}
		defer out.Close()
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return nil, err
		}
	}
	return resp, nil
}

func UserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
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
		directoryNames = append(directoryNames, subdirectory.Name())
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
