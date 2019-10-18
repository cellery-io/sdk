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

package commands

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/cellery-io/sdk/components/cli/internal/test"
	"github.com/cellery-io/sdk/components/cli/kubernetes"
)

func TestGetCellTableData(t *testing.T) {
	tests := []struct {
		name        string
		want        string
		MockKubeCli *test.MockKubeCli
	}{
		{
			name: "list instances with single cell instance",
			want: "employee",
			MockKubeCli: test.NewMockKubeCli(kubernetes.Cells{
				Items: []kubernetes.Cell{
					{
						CellMetaData: kubernetes.K8SMetaData{
							Name:              "employee",
							CreationTimestamp: "2019-10-18T11:40:36Z",
						},
					},
				},
			}),
		},
	}
	for _, testIteration := range tests {
		t.Run(testIteration.name, func(t *testing.T) {
			got := getCellTableData(testIteration.MockKubeCli)[0][0]
			if diff := cmp.Diff(testIteration.want, got); diff != "" {
				t.Errorf("getCellTableData (-want, +got)\n%v", diff)
			}
		})
	}
}
