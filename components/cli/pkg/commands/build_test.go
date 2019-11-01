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
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/cellery-io/sdk/components/cli/internal/test"
	image2 "github.com/cellery-io/sdk/components/cli/pkg/image"
)

var expectedTempBalFileContent = "import ballerina/config;\n" +
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
	"}\n" +
	"\n" +

	"public function main(string action, cellery:ImageName iName," +
	" map<cellery:ImageName> instances) returns error? {\n	return build(iName);" +
	"\n}"

func TestCreateTempBalFile(t *testing.T) {
	mockCli := test.NewMockCli()
	currentDir, _ := mockCli.FileSystem().CurrentDir()
	tmpBal, err := os.Create(filepath.Join(currentDir, "hello.bal"))
	defer func() { os.RemoveAll(currentDir) }()
	err = ioutil.WriteFile(tmpBal.Name(), []byte(expectedInitContent), 0644)
	if err != nil {
		t.Errorf("Failed to create bal file: %v", err)
	}
	tmpExecutableBalFile, err := createTempBalFile(mockCli, tmpBal.Name())
	if err != nil {
		t.Errorf("Failed to create temp executable bal file: %v", err)
	}
	tmpExecutableBalFileContent, err := ioutil.ReadFile(tmpExecutableBalFile)
	if err != nil {
		t.Errorf("Failed to read temp executable bal file: %v", err)
	}
	got := string(tmpExecutableBalFileContent)
	if diff := cmp.Diff(expectedTempBalFileContent, got); diff != "" {
		t.Errorf("Write (-want, +got)\n%v", diff)
	}
}

func TestExecuteTempBalFile(t *testing.T) {
	mockCli := test.NewMockCli()
	currentDir, _ := mockCli.FileSystem().CurrentDir()
	mockBalExecutor := &test.MockBalExecutor{
		CurrentDir: currentDir,
	}
	defer func() { os.RemoveAll(currentDir) }()
	executableBal, _ := os.Create(filepath.Join(currentDir, "executable.bal"))
	err := executeTempBalFile(mockBalExecutor, executableBal.Name(), []byte("test"))
	if err != nil {
		t.Errorf("Failed to run executable bal file: %v", err)
	}
	var artifacts []string
	filepath.Walk(currentDir, func(path string, info os.FileInfo, err error) error {
		artifacts = append(artifacts, path)
		return nil
	})
	got := artifacts
	expected := []string{currentDir, currentDir + "/metadata.json"}
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Errorf("Write (-want, +got)\n%v", diff)
	}
}

func TestGenerateMetaData(t *testing.T) {
	mockCli := test.NewMockCli()
	projectDir, _ := mockCli.FileSystem().CurrentDir()
	targetDir := filepath.Join(projectDir, "target")
	os.MkdirAll(filepath.Join(targetDir, "cellery"), os.ModePerm)
	defer func() { os.RemoveAll(projectDir) }()
	cellImage := &image2.CellImage{
		Registry:     "registry.hub.cellery.io",
		Organization: "myorg",
		ImageName:    "aaa",
		ImageVersion: "1.0.0",
	}
	mockMetaData, err := ioutil.ReadFile(filepath.Join("testdata", "metadata.json"))
	if err != nil {
		t.Errorf("Failed to read metadatajson.golden: %v", err)
	}
	mokcMetaDataJson, err := os.Create(filepath.Join(targetDir, "cellery", "metadata.json"))
	if err != nil {
		t.Errorf("Failed to create mock metadatajson: %v", err)
	}
	err = ioutil.WriteFile(mokcMetaDataJson.Name(), []byte(mockMetaData), 0644)
	if err != nil {
		t.Errorf("Failed to write to mock metadatajson: %v", err)
	}
	mockYamlData, err := ioutil.ReadFile(filepath.Join("testdata", "hello.yaml"))
	if err != nil {
		t.Errorf("Failed to read yaml.golden: %v", err)
	}
	mockYamlFile, err := os.Create(filepath.Join(targetDir, "cellery", "aaa.yaml"))
	if err != nil {
		t.Errorf("Failed to create mock yaml file: %v", err)
	}
	err = ioutil.WriteFile(mockYamlFile.Name(), []byte(mockYamlData), 0644)
	if err != nil {
		t.Errorf("Failed to write to mock yaml file: %v", err)
	}
	err = generateMetaData(cellImage, projectDir)
	if err != nil {
		t.Errorf("Failed to generate metadata: %v", err)
	}
	metaData, err := ioutil.ReadFile(mokcMetaDataJson.Name())
	if err != nil {
		t.Errorf("Failed to read temp executable bal file: %v", err)
	}
	got := &image2.MetaData{}
	if err = json.Unmarshal(metaData, got); err != nil {
		t.Errorf("error unmarshalling metadata json: %v", err)
	}
	if diff := cmp.Diff("myorg", got.Organization); diff != "" {
		t.Errorf("Write (-want, +got)\n%v", diff)
	}
	if diff := cmp.Diff("aaa", got.Name); diff != "" {
		t.Errorf("Write (-want, +got)\n%v", diff)
	}
	if diff := cmp.Diff("1.0.0", got.Version); diff != "" {
		t.Errorf("Write (-want, +got)\n%v", diff)
	}
}
