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

package instance

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/cellery-io/sdk/components/cli/internal/test"
)

func TestRunLogs(t *testing.T) {
	logsMap := make(map[string]string)
	componentLogsMap := make(map[string]string)
	logsMap["employee"] = "employee cell is running"
	componentLogsMap["employee:job"] = "job component is running"
	mockKubeCli := test.NewMockKubeCli(test.WithCellLogs(logsMap), test.WithComponentLogs(componentLogsMap))
	mockCli := test.NewMockCli(test.SetKubeCli(mockKubeCli))

	tests := []struct {
		name      string
		instance  string
		component string
		sysLog    bool
	}{
		{
			name:      "logs of cell instance (all components)",
			instance:  "employee",
			component: "",
			sysLog:    false,
		},
		{
			name:      "logs of cell instance (user components)",
			instance:  "employee",
			component: "",
			sysLog:    true,
		},
		{
			name:      "logs of cell component",
			instance:  "employee",
			component: "job",
			sysLog:    false,
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			err := RunLogs(mockCli, tst.instance, tst.component, tst.sysLog)
			if err != nil {
				t.Errorf("error in RunLogs, %v", err)
			}
		})
	}
}

func TestRunLogsError(t *testing.T) {
	tests := []struct {
		name       string
		instance   string
		component  string
		errMessage string
	}{
		{
			name:       "No logs of cell instance",
			instance:   "employee",
			component:  "",
			errMessage: "No logs found%!(EXTRA *errors.errorString=cannot find cell instance employee)",
		},
		{
			name:       "No logs of cell component",
			instance:   "employee",
			component:  "job",
			errMessage: "No logs found%!(EXTRA *errors.errorString=cannot find component job of cell instance employee)",
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			err := RunLogs(test.NewMockCli(test.SetKubeCli(test.NewMockKubeCli())), tst.instance, tst.component, false)
			if diff := cmp.Diff(tst.errMessage, err.Error()); diff != "" {
				t.Errorf("RunLogs: unexpected error (-want, +got)\n%v", diff)
			}
		})
	}
}
