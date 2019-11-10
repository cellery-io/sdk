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
	"fmt"
	"time"

	"github.com/tj/go-spin"
)

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
