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

package util

import (
	"archive/zip"
	"bufio"
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/tj/go-spin"
	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/cheggaaa/pb.v1"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/kubectl"
)

var Bold = color.New(color.Bold).SprintFunc()
var CyanBold = color.New(color.FgCyan).Add(color.Bold).SprintFunc()
var Faint = color.New(color.Faint).SprintFunc()
var Green = color.New(color.FgGreen).SprintfFunc()
var GreenBold = color.New(color.FgGreen).Add(color.Bold).SprintFunc()
var YellowBold = color.New(color.FgYellow).Add(color.Bold).SprintFunc()
var Red = color.New(color.FgRed).Add(color.Bold).SprintFunc()

func PrintWhatsNextMessage(action string, cmd string) {
	fmt.Println()
	fmt.Println(Bold("What's next?"))
	fmt.Println("--------------------------------------------------------")
	fmt.Printf("Execute the following command to %s:\n", action)
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
		ExitWithErrorMessage("Error parsing time", err)
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

func CelleryInstallationDir() string {
	celleryHome := ""
	if runtime.GOOS == "darwin" {
		celleryHome = constants.CELLERY_INSTALLATION_PATH_MAC
	}
	if runtime.GOOS == "linux" {
		celleryHome = constants.CELLERY_INSTALLATION_PATH_UBUNTU
	}
	return celleryHome
}

func BallerinaInstallationDir() string {
	ballerinaHome := ""
	if runtime.GOOS == "darwin" {
		ballerinaHome = constants.BALLERINA_INSTALLATION_PATH_MAC
	}
	if runtime.GOOS == "linux" {
		ballerinaHome = constants.BALLERINA_INSTALLATION_PATH_UBUNTU
	}
	return ballerinaHome

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

func ReplaceFile(fileToBeReplaced, fileToReplace string) error {
	errRename := RenameFile(fileToBeReplaced, fileToBeReplaced+"-old")
	if errRename != nil {
		return errRename
	}
	errCopy := CopyFile(fileToReplace, fileToBeReplaced)
	if errCopy != nil {
		return errCopy
	}
	return nil
}

func ExecuteCommand(cmd *exec.Cmd, errorMessage string) error {
	stderrReader, _ := cmd.StderrPipe()
	stderrScanner := bufio.NewScanner(stderrReader)

	go func() {
		for stderrScanner.Scan() {
			fmt.Println(stderrScanner.Text())
		}
	}()
	err := cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

func DownloadFromS3Bucket(bucket, item, path string, displayProgressBar bool) {
	file, err := os.Create(filepath.Join(path, item))
	if err != nil {
		ExitWithErrorMessage("Failed to create file path "+path, err)
	}

	defer file.Close()

	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(constants.AWS_REGION), Credentials: credentials.AnonymousCredentials},
	)

	// Get the object size
	s3ObjectSize := GetS3ObjectSize(bucket, item)

	// Create a downloader with the session and custom options
	downloader := s3manager.NewDownloader(sess, func(d *s3manager.Downloader) {
		d.PartSize = 64 * 1024 * 1024 // 64MB per part
		d.Concurrency = 6
	})

	writer := &progressWriter{writer: file, size: s3ObjectSize}
	writer.display = displayProgressBar

	writer.init(s3ObjectSize)
	numBytes, err := downloader.Download(writer,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(item),
		})
	if err != nil {
		ExitWithErrorMessage("Failed to download "+item+" from s3 bucket "+bucket, fmt.Errorf("Error downloading from file", err))
	}

	writer.finish()
	fmt.Println("Download completed", file.Name(), numBytes, "bytes")
}

func GetS3ObjectSize(bucket, item string) int64 {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(constants.AWS_REGION), Credentials: credentials.AnonymousCredentials},
	)

	svc := s3.New(sess)
	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(item),
	}

	result, err := svc.HeadObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			ExitWithErrorMessage("Error getting size of file", aerr)
		} else {
			ExitWithErrorMessage("Error getting size of file", err)
		}
	}
	return *result.ContentLength
}

func ExtractTarGzFile(extractTo, archive_name string) error {
	cmd := exec.Command("tar", "-zxvf", archive_name)
	cmd.Dir = extractTo

	ExecuteCommand(cmd, "Error occured in extracting file :"+archive_name)

	return nil
}

// RequestCredentials requests the credentials form the user and returns them
func RequestCredentials(credentialType string, usernameOverride string) (string, string, error) {
	fmt.Println()
	fmt.Println(YellowBold("?") + " " + credentialType + " credentials required")

	var username string
	var err error
	if usernameOverride == "" {
		// Requesting the username from the user
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Username: ")
		username, err = reader.ReadString('\n')
		if err != nil {
			return "", "", err
		}
	} else {
		username = usernameOverride
	}

	// Requesting the password from the user
	fmt.Print("Password: ")
	bytePassword, err := terminal.ReadPassword(0)
	if err != nil {
		return username, "", err
	}
	password := string(bytePassword)

	fmt.Println()
	return strings.TrimSpace(username), strings.TrimSpace(password), nil
}

// StartNewSpinner starts a new spinner with the provided message
func StartNewSpinner(action string) *Spinner {
	newSpinner := &Spinner{
		core:           spin.New(),
		action:         action,
		previousAction: action,
		isRunning:      true,
		isSpinning:     true,
		error:          false,
	}
	go func() {
		for newSpinner.isRunning {
			newSpinner.mux.Lock()
			newSpinner.spin()
			newSpinner.mux.Unlock()
			time.Sleep(100 * time.Millisecond)
		}
	}()
	return newSpinner
}

// SetNewAction sets the current action of a spinner
func (s *Spinner) SetNewAction(action string) {
	s.mux.Lock()
	s.action = action
	s.spin()
	s.mux.Unlock()
}

// Pause the spinner and clear the line
func (s *Spinner) Pause() {
	s.mux.Lock()
	s.isSpinning = false
	fmt.Printf("\r\x1b[2K")
	s.mux.Unlock()
}

// Resume the previous action
func (s *Spinner) Resume() {
	s.mux.Lock()
	s.isSpinning = true
	s.spin()
	s.mux.Unlock()
}

// Stop causes the spinner to stop
func (s *Spinner) Stop(isSuccess bool) {
	s.mux.Lock()
	s.error = !isSuccess
	s.action = ""
	s.spin()
	s.isRunning = false
	s.mux.Unlock()
}

// spin causes the spinner to do one spin
func (s *Spinner) spin() {
	if s.isRunning && s.isSpinning {
		if s.action != s.previousAction {
			if s.previousAction != "" {
				var icon string
				if s.error {
					icon = Red("\U0000274C")
				} else {
					icon = Green("\U00002714")
				}
				fmt.Printf("\r\x1b[2K%s %s\n", icon, s.previousAction)
			}
			s.previousAction = s.action
		}
		if s.action != "" {
			fmt.Printf("\r\x1b[2K\033[36m%s\033[m %s", s.core.Next(), s.action)
		}
	}
}

// ParseImageTag parses the given image name string and returns a CellImage struct with the relevant information.
func ParseImageTag(cellImageString string) (parsedCellImage *CellImage, err error) {
	cellImage := &CellImage{
		constants.CENTRAL_REGISTRY_HOST,
		"",
		"",
		"",
	}

	if cellImageString == "" {
		return cellImage, errors.New("no cell image specified")
	}

	const IMAGE_FORMAT_ERROR_MESSAGE = "incorrect image name format. Image name should be " +
		"[REGISTRY[:REGISTRY_PORT]/]ORGANIZATION/IMAGE_NAME:VERSION"

	// Parsing the cell image string
	strArr := strings.Split(cellImageString, "/")
	if len(strArr) == 3 {
		cellImage.Registry = strArr[0]
		cellImage.Organization = strArr[1]
		imageTag := strings.Split(strArr[2], ":")
		if len(imageTag) != 2 {
			return cellImage, errors.New(IMAGE_FORMAT_ERROR_MESSAGE)
		}
		cellImage.ImageName = imageTag[0]
		cellImage.ImageVersion = imageTag[1]
	} else if len(strArr) == 2 {
		cellImage.Organization = strArr[0]
		imageNameSplit := strings.Split(strArr[1], ":")
		if len(imageNameSplit) != 2 {
			return cellImage, errors.New(IMAGE_FORMAT_ERROR_MESSAGE)
		}
		cellImage.ImageName = imageNameSplit[0]
		cellImage.ImageVersion = imageNameSplit[1]
	} else {
		return cellImage, errors.New(IMAGE_FORMAT_ERROR_MESSAGE)
	}

	return cellImage, nil
}

// ValidateImageTag validates the image tag (without the registry in it).
func ValidateImageTag(imageTag string) error {
	r := regexp.MustCompile("^([^/:]*)/([^/:]*):([^/:]*)$")
	subMatch := r.FindStringSubmatch(imageTag)

	if subMatch == nil {
		return fmt.Errorf("expects <organization>/<cell-image>:<version> as the tag, received %s", imageTag)
	}

	organization := subMatch[1]
	isValid, err := regexp.MatchString(fmt.Sprintf("^%s$", constants.CELLERY_ID_PATTERN), organization)
	if err != nil || !isValid {
		return fmt.Errorf("expects a valid organization name (lower case letters, numbers and dashes "+
			"with only letters and numbers at the begining and end), received %s", organization)
	}

	imageName := subMatch[2]
	isValid, err = regexp.MatchString(fmt.Sprintf("^%s$", constants.CELLERY_ID_PATTERN), imageName)
	if err != nil || !isValid {
		return fmt.Errorf("expects a valid image name (lower case letters, numbers and dashes "+
			"with only letters and numbers at the begining and end), received %s", imageName)
	}

	imageVersion := subMatch[3]
	isValid, err = regexp.MatchString(fmt.Sprintf("^%s$", constants.IMAGE_VERSION_PATTERN), imageVersion)
	if err != nil || !isValid {
		return fmt.Errorf("expects a valid image version (lower case letters, numbers, dashes and dots "+
			"with only letters and numbers at the begining and end), received %s", imageVersion)
	}

	return nil
}

// ValidateImageTag validates the image tag (with the registry in it). The registry is an option element
// in this validation.
func ValidateImageTagWithRegistry(imageTag string) error {
	r := regexp.MustCompile("^(?:([^/]*)/)?([^/:]*)/([^/:]*):([^/:]*)$")
	subMatch := r.FindStringSubmatch(imageTag)

	if subMatch == nil {
		return fmt.Errorf("expects [<registry>]/<organization>/<cell-image>:<version> as the tag, received %s",
			imageTag)
	}

	registry := subMatch[1]
	isValid, err := regexp.MatchString(fmt.Sprintf("^%s$", constants.DOMAIN_NAME_PATTERN), registry)
	if registry != "" && (err != nil || !isValid) {
		return fmt.Errorf("expects a valid URL as the registry, received %s", registry)
	}

	organization := subMatch[2]
	isValid, err = regexp.MatchString(fmt.Sprintf("^%s$", constants.CELLERY_ID_PATTERN), organization)
	if err != nil || !isValid {
		return fmt.Errorf("expects a valid organization name (lower case letters, numbers and dashes "+
			"with only letters and numbers at the begining and end), received %s", organization)
	}

	imageName := subMatch[3]
	isValid, err = regexp.MatchString(fmt.Sprintf("^%s$", constants.CELLERY_ID_PATTERN), imageName)
	if err != nil || !isValid {
		return fmt.Errorf("expects a valid image name (lower case letters, numbers and dashes "+
			"with only letters and numbers at the begining and end), received %s", imageName)
	}

	imageVersion := subMatch[4]
	isValid, err = regexp.MatchString(fmt.Sprintf("^%s$", constants.IMAGE_VERSION_PATTERN), imageVersion)
	if err != nil || !isValid {
		return fmt.Errorf("expects a valid image version (lower case letters, numbers, dashes and dots "+
			"with only letters and numbers at the begining and end), received %s", imageVersion)
	}

	return nil
}

// ExitWithErrorMessage prints an error message and exits the command
func ExitWithErrorMessage(message string, err error) {
	fmt.Printf("\n\n\x1b[31;1m%s:\x1b[0m %v\n\n", message, err)
	os.Exit(1)
}

// PrintSuccessMessage prints the standard command success message
func PrintSuccessMessage(message string) {
	fmt.Println()
	fmt.Printf("\n%s %s\n", GreenBold("\U00002714"), message)
}

func PrintWarningMessage(message string) {
	fmt.Println()
	fmt.Printf("%s\n", YellowBold("\U000026A0 "+message))
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

// RunMethodExists checks if the run method exists in ballerina file
func RunMethodExists(sourceFile string) (bool, error) {
	sourceFileBytes, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		return false, err
	}

	// Check whether run method exists
	return regexp.MatchString(
		`.*public(\s)+function(\s)+run(\s)*\((s)*cellery:ImageName(\s)+.+(\s)*,(\s)*map<cellery:ImageName>(\s)+.+(\s)*\)(\s)+returns(\s)+error\\?`,
		string(sourceFileBytes))
}

// TestMethodExists checks if the test method exists in ballerina file
func TestMethodExists(sourceFile string) (bool, error) {
	sourceFileBytes, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		return false, err
	}

	// Check whether test method exists
	return regexp.MatchString(
		`.*public(\s)+function(\s)+test(\s)*\((s)*cellery:ImageName(\s)+.+(\s)*,(\s)*map<cellery:ImageName>(\s)+.+(\s)*\)(\s)+returns(\s)+error\\?`,
		string(sourceFileBytes))
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

func GetCurrentPath() (string, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", err
	}
	return dir, nil
}

func ContainsInStringArray(array []string, item string) bool {
	for _, element := range array {
		if element == item {
			return true
		}
	}
	return false
}

func GetYesOrNoFromUser(question string, withBackOption bool) (bool, bool, error) {
	options := []string{}
	var isBackSelected = false
	if withBackOption {
		options = []string{"Yes", "No", constants.CELLERY_SETUP_BACK}
	} else {
		options = []string{"Yes", "No"}
	}
	prompt := promptui.Select{
		Label: question,
		Items: options,
	}
	_, result, err := prompt.Run()
	if result == constants.CELLERY_SETUP_BACK {
		isBackSelected = true
	}
	if err != nil {
		return false, isBackSelected, fmt.Errorf("Prompt failed %v\n", err)
	}
	return result == "Yes", isBackSelected, nil
}

// OpenBrowser opens up the provided URL in a browser
func OpenBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "openbsd":
		fallthrough
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		r := strings.NewReplacer("&", "^&")
		cmd = exec.Command("cmd", "/c", "start", r.Replace(url))
	}
	if cmd != nil {
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Start()
		if err != nil {
			errStr := string(stderr.Bytes())
			fmt.Printf("%s\n", errStr)
			fmt.Println("Failed to open browser: " + err.Error())
		}
		err = cmd.Wait()
		if err != nil {
			errStr := string(stderr.Bytes())
			fmt.Printf("\x1b[31;1m%s\x1b[0m", errStr)
			fmt.Println("Failed to open browser: " + err.Error())
		}
		outStr := string(stdout.Bytes())
		fmt.Println(outStr)
		return err
	} else {
		return errors.New("unsupported platform")
	}
}

func ReadCellImageYaml(cellImage string) []byte {
	parsedCellImage, err := ParseImageTag(cellImage)
	cellImageZip := path.Join(UserHomeDir(), constants.CELLERY_HOME, "repo", parsedCellImage.Organization,
		parsedCellImage.ImageName, parsedCellImage.ImageVersion, parsedCellImage.ImageName+constants.CELL_IMAGE_EXT)

	// Create tmp directory
	tmpPath := filepath.Join(UserHomeDir(), constants.CELLERY_HOME, "tmp", "imageExtracted")
	err = CleanOrCreateDir(tmpPath)
	if err != nil {
		panic(err)
	}

	err = Unzip(cellImageZip, tmpPath)
	if err != nil {
		panic(err)
	}
	if err != nil {
		ExitWithErrorMessage("Error occurred while extracting cell image", err)
	}

	cellYamlContent, err := ioutil.ReadFile(filepath.Join(tmpPath, constants.ZIP_ARTIFACTS, "cellery",
		parsedCellImage.ImageName+".yaml"))
	if err != nil {
		ExitWithErrorMessage("Error while reading cell image content", err)
	}
	// Delete tmp directory
	err = CleanOrCreateDir(tmpPath)
	if err != nil {
		ExitWithErrorMessage("Error while reading cell image content", err)
	}
	return cellYamlContent
}

func WaitForRuntime(checkKnative, hpaEnabled bool) {
	spinner := StartNewSpinner("Checking cluster status...")
	err := kubectl.WaitForCluster(time.Hour)
	if err != nil {
		spinner.Stop(false)
		ExitWithErrorMessage("Error while checking cluster status", err)
	}
	spinner.SetNewAction("Cluster status...OK")
	spinner.Stop(true)

	spinner = StartNewSpinner("Checking runtime status (Istio)...")
	err = kubectl.WaitForDeployments("istio-system", time.Minute*15)
	if err != nil {
		spinner.Stop(false)
		ExitWithErrorMessage("Error while checking runtime status (Istio)", err)
	}
	spinner.SetNewAction("Runtime status (Istio)...OK")
	spinner.Stop(true)

	if checkKnative {
		spinner = StartNewSpinner("Checking runtime status (Knative Serving)...")
		err = kubectl.WaitForDeployments("knative-serving", time.Minute*15)
		if err != nil {
			spinner.Stop(false)
			ExitWithErrorMessage("Error while checking runtime status (Knative Serving)", err)
		}
		spinner.SetNewAction("Runtime status (Knative Serving)...OK")
		spinner.Stop(true)
	}

	if hpaEnabled {
		spinner = StartNewSpinner("Checking runtime status (Metrics server)...")
		err = kubectl.WaitForDeployment("available", 900, "metrics-server", "kube-system")
		if err != nil {
			spinner.Stop(false)
			ExitWithErrorMessage("Error while checking runtime status (Metrics server)", err)
		}
		spinner.SetNewAction("Runtime status (Metrics server)...OK")
		spinner.Stop(true)
	}

	spinner = StartNewSpinner("Checking runtime status (Cellery)...")
	err = kubectl.WaitForDeployments("cellery-system", time.Minute*15)
	if err != nil {
		spinner.Stop(false)
		ExitWithErrorMessage("Error while checking runtime status (Cellery)", err)
	}
	spinner.SetNewAction("Runtime status (Cellery)...OK")
	spinner.Stop(true)

}

func FormatBytesToString(size int64) string {
	return pb.Format(size).To(pb.U_BYTES_DEC).String()
}

func MergeKubeConfig(newConfigFile string) error {
	newConf, err := kubectl.ReadConfig(newConfigFile)
	if err != nil {
		return err
	}

	confFile, err := kubectl.DefaultConfigFile()
	if err != nil {
		return err
	}

	if _, err := os.Stat(confFile); err != nil {
		if os.IsNotExist(err) {
			// kube-config does not exist. Create a new one
			confDir, err := kubectl.DefaultConfigDir()
			if err != nil {
				return err
			}
			// Check for .kube directory and create if not present
			if _, err := os.Stat(confDir); os.IsNotExist(err) {
				err = os.Mkdir(confDir, 0755)
				if err != nil {
					return err
				}
			}
			return kubectl.WriteConfig(confFile, newConf)
		} else {
			return err
		}
	}

	oldConf, err := kubectl.ReadConfig(confFile)
	if err != nil {
		return err
	}
	merged := kubectl.MergeConfig(oldConf, newConf)
	return kubectl.WriteConfig(confFile, merged)
}

func IsCompleteSetupSelected() (bool, bool) {
	var isCompleteSelected = false
	var isBackSelected = false
	cellTemplate := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Help:     Faint("[Use arrow keys]"),
	}

	cellPrompt := promptui.Select{
		Label:     YellowBold("?") + " Select the type of runtime",
		Items:     []string{constants.BASIC, constants.COMPLETE, constants.CELLERY_SETUP_BACK},
		Templates: cellTemplate,
	}
	_, value, err := cellPrompt.Run()
	if err != nil {
		ExitWithErrorMessage("Failed to select an option: %v", err)
	}
	if value == constants.CELLERY_SETUP_BACK {
		isBackSelected = true
	}
	if value == constants.COMPLETE {
		isCompleteSelected = true
	}
	return isCompleteSelected, isBackSelected
}

func IsLoadBalancerIngressTypeSelected() (bool, bool) {
	var isLoadBalancerSelected = false
	var isBackSelected = false
	cellTemplate := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U000027A4 {{ .| bold }}",
		Inactive: "  {{ . | faint }}",
		Help:     Faint("[Use arrow keys]"),
	}

	cellPrompt := promptui.Select{
		Label:     YellowBold("?") + " Select ingress mode",
		Items:     []string{constants.INGRESS_MODE_NODE_PORT, constants.INGRESS_MODE_LOAD_BALANCER, constants.CELLERY_SETUP_BACK},
		Templates: cellTemplate,
	}
	_, value, err := cellPrompt.Run()
	if err != nil {
		ExitWithErrorMessage("Failed to select an option: %v", err)
	}
	if value == constants.CELLERY_SETUP_BACK {
		isBackSelected = true
	}
	if value == constants.INGRESS_MODE_LOAD_BALANCER {
		isLoadBalancerSelected = true
	}
	return isLoadBalancerSelected, isBackSelected
}

func IsCommandAvailable(name string) bool {
	cmd := exec.Command("/bin/sh", "-c", "command -v "+name)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func CreateTempExecutableBalFile(file string, action string) (string, error) {
	var ballerinaMain = ""
	if action == "build" {
		ballerinaMain = `
public function main(string action, cellery:ImageName iName, map<cellery:ImageName> instances) returns error? {
	return build(iName);
}`
	} else if action == "run" {
		ballerinaMain = `
public function main(string action, cellery:ImageName iName, map<cellery:ImageName> instances) returns error? {
	return run(iName, instances);
}`
	} else if action == "test" {
		ballerinaMain = `
public function main(string action, cellery:ImageName iName, map<cellery:ImageName> instances) returns error? {
	return test(iName, instances);
}`
	} else {
		return "", errors.New("invalid action:" + action)
	}

	originalFilePath, _ := filepath.Abs(file)
	input, err := ioutil.ReadFile(originalFilePath)
	if err != nil {
		return "", err
	}
	var newFileContent = string(input) + ballerinaMain

	balFileName := filepath.Base(originalFilePath)
	var newFileName = strings.Replace(balFileName, ".bal", "", 1) + "_" + action + ".bal"
	originalFileDir := filepath.Dir(originalFilePath)
	targetAbs := filepath.Join(originalFileDir, "target")
	err = os.Mkdir(targetAbs, 0777)
	if err != nil {
		return "", err
	}
	targetFilePath := filepath.Join(targetAbs, newFileName)
	err = ioutil.WriteFile(targetFilePath, []byte(newFileContent), 0644)
	if err != nil {
		return "", err
	}

	return targetFilePath, nil
}
