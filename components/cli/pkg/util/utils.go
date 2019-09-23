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
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
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
	if year > 0 {
		duration += strconv.Itoa(year) + " years "
		numOfTimeUnits++
	}
	if month > 0 {
		duration += strconv.Itoa(month) + " months "
		numOfTimeUnits++
	}
	if day > 0 {
		duration += strconv.Itoa(day) + " days "
		numOfTimeUnits++
	}
	if hour > 0 {
		duration += strconv.Itoa(hour) + " hours "
		numOfTimeUnits++
	}
	if min > 0 {
		duration += strconv.Itoa(min) + " minutes "
		numOfTimeUnits++
	}
	if sec > 0 {
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

func UserHomeCelleryDir() string {
	return filepath.Join(UserHomeDir(), constants.CELLERY_HOME)
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

func ExecuteCommand(cmd *exec.Cmd) error {
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

// RunMethodExists checks if the run method exists in ballerina file
func RunMethodExists(sourceFile string) (bool, error) {
	sourceFileBytes, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		return false, err
	}

	// Check whether run method exists
	return regexp.MatchString(
		`.*public(\s)+function(\s)+run(\s)*\((s)*cellery:ImageName(\s)+.+(\s)*,(\s)*map<cellery:ImageName>(\s)+.+(\s)*\)(\s)+returns(\s)+\(cellery:InstanceState\[\]\|error\?\)`,
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

func ContainsInStringArray(array []string, item string) bool {
	for _, element := range array {
		if element == item {
			return true
		}
	}
	return false
}

func GetYesOrNoFromUser(question string, withBackOption bool) (bool, bool, error) {
	var options []string
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
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Start()
		if err != nil {
			log.Printf("Failed to open browser due to error %v", err)
			return fmt.Errorf("Failed to open browser: " + err.Error())
		}
		err = cmd.Wait()
		if err != nil {
			log.Printf("Failed to wait for open browser command to finish due to error %v", err)
			return fmt.Errorf("Failed to wait for open browser command to finish: " + err.Error())
		}
		return nil
	} else {
		return errors.New("unsupported platform")
	}
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
public function main(string action, cellery:ImageName iName, map<cellery:ImageName> instances, boolean startDependencies, boolean shareDependencies) returns error? {
	cellery:InstanceState[]|error? result = run(iName, instances, startDependencies, shareDependencies);
    if (result is error?) {
		return result;
	}
}`
	} else if action == "test" {
		ballerinaMain = `
public function main(string action, cellery:ImageName iName, map<cellery:ImageName> instances, boolean startDependencies, boolean shareDependencies) returns error? {
	return test(iName, instances, startDependencies, shareDependencies);
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

func ConvertToAlphanumeric(input, replacement string) string {
	reg, err := regexp.Compile("[^A-Za-z0-9]+")
	if err != nil {
		ExitWithErrorMessage("Error making regex", err)
	}
	processedString := reg.ReplaceAllString(input, replacement)
	return processedString
}
