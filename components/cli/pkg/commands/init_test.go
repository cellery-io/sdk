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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/cellery-io/sdk/components/cli/internal/test"
)

func TestInitcreateBalFile(t *testing.T) {
	projectName := "my-project"
	tmpdir, _ := ioutil.TempDir("", "cellery-current-dir")
	defer func() { os.RemoveAll(tmpdir) }()
	balFile, _ := createBalFile(tmpdir, projectName)
	// Check if bal file is getting successfully created with the correct name.
	got := balFile.Name()
	if diff := cmp.Diff(filepath.Join(tmpdir, projectName, projectName+".bal"), got); diff != "" {
		t.Errorf("Write (-want, +got)\n%v", diff)
	}
}

func TestInitWriteCellTemplate(t *testing.T) {
	expected := "import ballerina/config;\n" +
		"import celleryio/cellery;\n" +
		"\n" +
		"public function build(cellery:ImageName iName) returns error? {\n" +
		"    // Hello Component\n" +
		"    // This Components exposes the HTML hello world page\n" +
		"    cellery:Component helloComponent = {\n" +
		"        name: \"hello\",\n" +
		"        source: {\n" +
		"            image: \"wso2cellery/samples-hello-world-webapp\"\n" +
		"        },\n" +
		"        ingresses: {\n" +
		"            webUI: <cellery:WebIngress>{ // Web ingress will be always exposed globally.\n" +
		"                port: 80,\n" +
		"                gatewayConfig: {\n" +
		"                    vhost: \"hello-world.com\",\n" +
		"                    context: \"/\"\n" +
		"                }\n" +
		"            }\n" +
		"        },\n" +
		"        envVars: {\n" +
		"            HELLO_NAME: { value: \"Cellery\" }\n" +
		"        }\n" +
		"    };\n" +
		"\n" +
		"    // Cell Initialization\n" +
		"    cellery:CellImage helloCell = {\n" +
		"        components: {\n" +
		"            helloComp: helloComponent\n" +
		"        }\n" +
		"    };\n" +
		"    return cellery:createImage(helloCell, untaint iName);\n" +
		"}\n" +
		"\n" +
		"public function run(cellery:ImageName iName, map<cellery:ImageName> instances, boolean startDependencies, boolean shareDependencies) returns (cellery:InstanceState[]|error?) {\n" +
		"    cellery:CellImage helloCell = check cellery:constructCellImage(untaint iName);\n" +
		"    string vhostName = config:getAsString(\"VHOST_NAME\");\n" +
		"    if (vhostName !== \"\") {\n" +
		"        cellery:WebIngress web = <cellery:WebIngress>helloCell.components.helloComp.ingresses.webUI;" +
		"\n" +
		"        web.gatewayConfig.vhost = vhostName;\n" +
		"    }\n" +
		"\n" +
		"    string helloName = config:getAsString(\"HELLO_NAME\");\n" +
		"    if (helloName !== \"\") {\n" +
		"        helloCell.components.helloComp.envVars.HELLO_NAME.value = helloName;\n" +
		"    }\n" +
		"    return cellery:createInstance(helloCell, iName, instances, startDependencies, shareDependencies);\n" +
		"}\n"

	mockCli := test.NewMockCli()
	writeCellTemplate(mockCli.Out(), CellTemplate)
	got := mockCli.OutBuffer().String()
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Errorf("Write (-want, +got)\n%v", diff)
	}
}
