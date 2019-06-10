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
	"context"
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
func FromBrowser(username string) (string, string, error) {
	conf := config.LoadConfig()
	timeout := make(chan bool)
	ch := make(chan string)
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

	fmt.Printf("\n%s\n\n", hubAuthUrl)
	go func() {
		mux := http.NewServeMux()
		server := http.Server{Addr: httpPortString, Handler: mux}
		//var timer *time.Timer
		mux.HandleFunc(callBackUrlContext, func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseForm()
			if err != nil {
				util.ExitWithErrorMessage("Error parsing the code", err)
			}
			code = r.Form.Get("code")
			ch <- code
			if len(code) != 0 {
				http.Redirect(w, r, conf.Hub.Url+"/sdk/auth-success", http.StatusSeeOther)
			} else {
				util.ExitWithErrorMessage("Did not receive any code", err)
			}
			flusher, ok := w.(http.Flusher)
			if !ok {
				util.ExitWithErrorMessage("Error in casting the flusher", err)
			}
			flusher.Flush()
			err = server.Shutdown(context.Background())
			if err != nil {
				util.ExitWithErrorMessage("Error while shutting down the server\n", err)
			}
		})
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			util.ExitWithErrorMessage("Error while establishing the service", err)
		}
	}()

	err := util.OpenBrowser(hubAuthUrl)
	if err != nil {
		fmt.Printf("Could not resolve the given url %s. Started to operate in the headless "+
			"mode\n", hubAuthUrl)
		return FromTerminal(username)
	}
	// Setting up a timeout
	go func() {
		time.Sleep(5 * time.Minute)
		timeout <- true
	}()
	// Wait for a code, or timeout
	select {
	case <-ch:
	case <-timeout:
		util.ExitWithErrorMessage("Failed to authenticate", errors.
			New("time out. Did not receive any code"))
	}
	token := getTokenFromCode(code, codeReceiverPort, conf)
	username, accessToken := getUsernameAndTokenFromJwt(token)
	return username, accessToken, nil
}

// FromTerminal is to allow this login flow to work in headless mode
func FromTerminal(username string) (string, string, error) {
	var password string
	if username == "" {
		fmt.Print("Enter username: ")
		_, err := fmt.Scanln(&username)
		if err != nil {
			util.ExitWithErrorMessage("Error reading the input username", err)
		}
	}
	fmt.Print("Enter password/token: ")
	bytePassword, err := terminal.ReadPassword(0)
	if err != nil {
		util.ExitWithErrorMessage("Error reading the input token", err)
	}
	password = strings.TrimSpace(string(bytePassword))
	username = strings.TrimSpace(username)
	fmt.Println()
	return username, password, nil
}

// getUsernameAndToken returns the extracted subject from the JWT
func getUsernameAndTokenFromJwt(response string) (string, string) {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(response), &result)
	if err != nil {
		util.ExitWithErrorMessage("Error while unmarshal the id_token", err)
	}
	idToken, ok := (result["id_token"]).(string)
	accessToken, ok := (result["access_token"]).(string)
	if !ok {
		util.ExitWithErrorMessage("Error while retrieving the access token", err)
	}
	jwtToken, _ := jwt.Parse(idToken, nil)
	claims := jwtToken.Claims.(jwt.MapClaims)
	sub, ok := claims["sub"].(string)
	if !ok {
		util.ExitWithErrorMessage("Error in casting the subject", err)
	}
	return sub, accessToken
}

// getTokenFromCode returns the JWT from the auth code provided
func getTokenFromCode(code string, port int, conf *config.Conf) string {
	tokenUrl := conf.Idp.Url + "/oauth2/token"
	responseBody := "client_id=" + conf.Idp.ClientId +
		"&grant_type=authorization_code&code=" + code +
		"&redirect_uri=" + fmt.Sprintf(callBackUrl, port)
	body := strings.NewReader(responseBody)
	// Token request
	req, err := http.NewRequest("POST", tokenUrl, body)
	if err != nil {
		util.ExitWithErrorMessage("Error while creating the code receiving request", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Could not connect to the client at %s", tokenUrl)
		util.ExitWithErrorMessage("Error occurred while connecting to the client", err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	respBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		util.ExitWithErrorMessage("Error occurred while reading the response body", err)
	}
	return string(respBody)
}
