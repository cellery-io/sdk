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

package kubectl

import (
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/osexec"

	"github.com/ghodss/yaml"
)

func DefaultConfigDir() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(u.HomeDir, ".kube"), nil
}

func DefaultConfigFile() (string, error) {
	dir, err := DefaultConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config"), nil
}

func ReadConfig(file string) (*Config, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	conf := &Config{}
	err = yaml.Unmarshal(b, conf)
	if err != nil {
		return nil, err
	}
	return conf, err
}

func WriteConfig(file string, conf *Config) error {
	b, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(file, b, 0644)
	if err != nil {
		return err
	}
	return nil
}

func MergeConfig(old, new *Config) *Config {
	merged := *old
	merged.Clusters = mergeClusters(merged.Clusters, new.Clusters)
	merged.AuthInfos = mergeUsers(merged.AuthInfos, new.AuthInfos)
	merged.Contexts = mergeContexts(merged.Contexts, new.Contexts)
	merged.CurrentContext = new.CurrentContext
	return &merged
}

func mergeClusters(old []*ClusterInfo, new []*ClusterInfo) []*ClusterInfo {
	existing := make(map[string]*ClusterInfo)
	for _, ov := range old {
		existing[ov.Name] = ov
	}
	for _, nv := range new {
		if val, ok := existing[nv.Name]; ok {
			val.Cluster = nv.Cluster
		} else {
			old = append(old, nv)
		}
	}
	return old
}

func mergeUsers(old []*UserInfo, new []*UserInfo) []*UserInfo {
	existing := make(map[string]*UserInfo)
	for _, ov := range old {
		existing[ov.Name] = ov
	}
	for _, nv := range new {
		if val, ok := existing[nv.Name]; ok {
			val.User = nv.User
		} else {
			old = append(old, nv)
		}
	}
	return old
}

func mergeContexts(old []*ContextInfo, new []*ContextInfo) []*ContextInfo {
	existing := make(map[string]*ContextInfo)
	for _, ov := range old {
		existing[ov.Name] = ov
	}
	for _, nv := range new {
		if val, ok := existing[nv.Name]; ok {
			val.Context = nv.Context
		} else {
			old = append(old, nv)
		}
	}
	return old
}

func UseContext(context string) error {
	cmd := exec.Command(
		constants.KUBECTL,
		"config",
		"use-context",
		context,
	)
	displayVerboseOutput(cmd)
	cmd.Stderr = os.Stderr
	return cmd.Run()
	return nil
}

func GetContexts() ([]byte, error) {
	cmd := exec.Command(
		constants.KUBECTL,
		"config",
		"view",
		"-o",
		"json",
	)
	displayVerboseOutput(cmd)
	return osexec.GetCommandOutputFromTextFile(cmd)
}

func GetContext() (string, error) {
	cmd := exec.Command(
		constants.KUBECTL,
		"config",
		"current-context",
	)
	displayVerboseOutput(cmd)
	out, err := osexec.GetCommandOutput(cmd)
	return out, err
}
