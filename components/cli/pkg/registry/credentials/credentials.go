/*
 * Copyright (c) 2019, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/cellery-io/sdk/components/cli/pkg/config"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

const callBackDefaultPort = 8888
const callBackUrlContext = "/auth"
const callBackUrl = "http://localhost:%d" + callBackUrlContext

// FromBrowser requests the credentials from the user
func FromBrowser(username string, isAuthorized chan bool, done chan bool) (string, string, error) {
	conf := config.LoadConfig()
	timeout := make(chan bool)
	authCode := make(chan string)
	var code string
	httpPortString := ":" + strconv.Itoa(callBackDefaultPort)
	var codeReceiverPort = callBackDefaultPort
	// This is to start the CLI auth in a different port is the default port is already occupied
	for {
		_, err := net.Dial("tcp", httpPortString)
		if err != nil {
			break
		}
		codeReceiverPort++
		httpPortString = ":" + strconv.Itoa(codeReceiverPort)
	}
	redirectUrl := url.QueryEscape(fmt.Sprintf(callBackUrl, codeReceiverPort))
	var hubAuthUrl = conf.Hub.Url + "/sdk/fidp-select?redirectUrl=" + redirectUrl

	go func() {
		mux := http.NewServeMux()
		server := http.Server{Addr: httpPortString, Handler: mux}
		//var timer *time.Timer
		mux.HandleFunc(callBackUrlContext, func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseForm()
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				util.ExitWithErrorMessage("Error parsing received query parameters", err)
			}
			code = r.Form.Get("code")
			ping := r.Form.Get("ping")
			if ping == "true" {
				w.Header().Set("Access-Control-Allow-Origin", conf.Hub.Url)
				w.Header().Set("Access-Control-Allow-Methods", http.MethodGet)
				w.WriteHeader(http.StatusOK)
			}
			if code != "" {
				authCode <- code
				authorized := <-isAuthorized
				if authorized {
					http.Redirect(w, r, conf.Hub.Url+"/sdk/auth-success", http.StatusSeeOther)
				} else {
					http.Redirect(w, r, conf.Hub.Url+"/sdk/auth-failure", http.StatusSeeOther)
					fmt.Println("\n\U0000274C Failed to authenticate")
				}
				flusher, ok := w.(http.Flusher)
				if !ok {
					util.ExitWithErrorMessage("Error in casting the flusher", err)
				}
				flusher.Flush()
				done <- true
			}
		})
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			util.ExitWithErrorMessage("Error while establishing the service", err)
		}
	}()

	fmt.Printf("\nOpening %s\n\n", hubAuthUrl)
	err := util.OpenBrowser(hubAuthUrl)
	if err != nil {
		fmt.Printf("\r\x1b[2K%s Could not open browser. Operating in the headless mode.",
			util.YellowBold("\U000026A0"))
		username, token, err := FromTerminal(username)
		go func() {
			// Mocking the channels used by the server to avoid hanging
			<-isAuthorized
			done <- true
		}()
		return username, token, err
	}
	// Setting up a timeout
	go func() {
		time.Sleep(15 * time.Minute)
		timeout <- true
	}()
	// Wait for a code, or timeout
	select {
	case <-authCode:
	case <-timeout:
		return "", "", errors.New("time out waiting for authentication")
	}
	token, err := getTokenFromCode(code, codeReceiverPort, conf)
	if err != nil {
		return "", "", fmt.Errorf("failed to get token for the authorized user: %v", err)
	}
	username, accessToken, err := getUsernameAndTokenFromJwt(token)
	if err != nil {
		return "", "", fmt.Errorf("failed to identify user from received token: %v", err)
	}
	return username, accessToken, nil
}

// FromTerminal is to allow this login flow to work in headless mode
func FromTerminal(username string) (string, string, error) {
	var password string
	if username == "" {
		fmt.Print("Enter username: ")
		_, err := fmt.Scanln(&username)
		if err != nil {
			return "", "", fmt.Errorf("failed to read the input username: %v", err)
		}
	}
	fmt.Print("Enter password/token: ")
	bytePassword, err := terminal.ReadPassword(0)
	if err != nil {
		return "", "", fmt.Errorf("failed to read the input password/token: %v", err)
	}
	password = strings.TrimSpace(string(bytePassword))
	username = strings.TrimSpace(username)
	fmt.Println()
	return username, password, nil
}

// getUsernameAndToken returns the extracted subject from the JWT
func getUsernameAndTokenFromJwt(response string) (string, string, error) {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(response), &result)
	if err != nil {
		return "", "", fmt.Errorf("failed to unmarshal the id_token: %v", err)
	}
	idToken, ok := (result["id_token"]).(string)
	accessToken, ok := (result["access_token"]).(string)
	if !ok {
		return "", "", fmt.Errorf("failed to retrieve the access token: %v", err)
	}
	jwtToken, _ := jwt.Parse(idToken, nil)
	claims := jwtToken.Claims.(jwt.MapClaims)
	sub, ok := claims["sub"].(string)
	if !ok {
		return "", "", fmt.Errorf("failed to read the user ID: %v", err)
	}
	return sub, accessToken, nil
}

// getTokenFromCode returns the JWT from the auth code provided
func getTokenFromCode(code string, port int, conf *config.Conf) (string, error) {
	tokenUrl := conf.Idp.Url + "/oauth2/token"
	responseBody := "client_id=" + conf.Idp.ClientId +
		"&grant_type=authorization_code&code=" + code +
		"&redirect_uri=" + fmt.Sprintf(callBackUrl, port)
	body := strings.NewReader(responseBody)
	// Token request
	req, err := http.NewRequest("POST", tokenUrl, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request to connect to Cellery Hub IdP: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to connect to Cellery Hub IdP: %v", err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	respBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read the response body: %v", err)
	}
	return string(respBody), nil
}
