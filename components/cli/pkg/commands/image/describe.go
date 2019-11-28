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
	"fmt"
	"regexp"

	"cellery.io/cellery/components/cli/cli"
	"cellery.io/cellery/components/cli/pkg/constants"
	"cellery.io/cellery/components/cli/pkg/image"
)

func RunDescribe(cli cli.Cli, name string) error {
	instancePattern, _ := regexp.MatchString(fmt.Sprintf("^%s$", constants.CelleryIdPattern), name)
	if instancePattern {
		// If the input of user is an instance describe running cell instance.
		err := cli.KubeCli().DescribeCell(name)
		if err != nil {
			return fmt.Errorf("error describing cell instance, %v", err)
		}
	} else {
		// If the input of user is a cell image print the cell yaml
		cellYamlContent, err := image.ReadCellImageYaml(cli.FileSystem().Repository(), name)
		if err != nil {
			return fmt.Errorf("error describing cell image, %v", err)
		}
		fmt.Fprintln(cli.Out(), string(cellYamlContent))
	}
	return nil
}
