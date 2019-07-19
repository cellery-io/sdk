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

package image

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func ReadMetaData(organization, project, version string) (*CellImageMetaData, error) {
	r, err := zip.OpenReader(filepath.Join(util.UserHomeDir(), constants.CELLERY_HOME, "repo",
		organization, project, version, project+".zip"))
	if err != nil {
		return nil, err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name != MetaDataFile() {
			continue
		}
		metaReader, err := f.Open()
		if err != nil {
			return nil, err
		}
		meta := &CellImageMetaData{}
		err = json.NewDecoder(metaReader).Decode(meta)
		if err != nil {
			return nil, err
		}
		metaReader.Close()
		return meta, nil
	}
	return nil, fmt.Errorf("missing metadata infomation in %s/%s:%s", organization, project, version)
}

func MetaDataFile() string {
	return filepath.Join("artifacts", "cellery", "metadata.json")
}
