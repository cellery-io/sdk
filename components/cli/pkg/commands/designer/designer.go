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

package designer

import (
	"cellery.io/cellery/components/cli/cli"
	"cellery.io/cellery/components/cli/pkg/constants"
	"cellery.io/cellery/components/cli/pkg/util"
	"encoding/json"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

var tmpSourceDir string
var tmpZipDir string

// Starts a http server to serve Cellery designer react app and API
func RunDesigner(cli cli.Cli) error {
	r := mux.NewRouter()
	tmpSourceDir = filepath.Join(cli.FileSystem().TempDir(), "designer")
	tmpZipDir = filepath.Join(cli.FileSystem().TempDir(), "downloads")

	// It's important that this is before your catch-all route ("/")
	api := r.PathPrefix("/api/").Subrouter()
	api.HandleFunc("/generate", generateHandler).Methods(http.MethodPost)

	// Serve static assets directly.
	designerDir := path.Join(cli.FileSystem().CelleryInstallationDir(), "designer")
	r.PathPrefix("/static").Handler(http.FileServer(http.Dir(designerDir)))

	r.PathPrefix("/download").HandlerFunc(downloadHandler(filepath.Join(tmpZipDir, "downloads.zip")))
	// Catch-all: Serve Designer app entry-point (index.html).
	r.PathPrefix("/").HandlerFunc(indexHandler(path.Join(designerDir, "index.html")))

	srv := &http.Server{
		Handler:      handlers.LoggingHandler(os.Stdout, r),
		Addr:         "0.0.0.0:8088",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	if err := cli.OpenBrowser("http://127.0.0.1:8088"); err != nil {
		return err
	}
	if err := srv.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

func indexHandler(entryPoint string) func(w http.ResponseWriter, r *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, entryPoint)
	}
	return fn
}

func downloadHandler(entryPoint string) func(w http.ResponseWriter, r *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, entryPoint)
	}
	return fn
}

func generateHandler(w http.ResponseWriter, r *http.Request) {
	var nodeData DesignMeta

	if err := util.CleanAndCreateDir(tmpZipDir); err != nil {
		log.Println("error occurred file create dir "+tmpZipDir, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := util.CleanAndCreateDir(tmpSourceDir); err != nil {
		log.Println("error occurred file create dir "+tmpSourceDir, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(r.Body).Decode(&nodeData)
	if err != nil {
		log.Println("error occurred decoding generate request body ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := generateCode(nodeData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte("http://127.0.0.1/download"))
}

func generateCode(designerMeta DesignMeta) error {
	tpl := template.Must(template.New("cell template").Funcs(template.FuncMap{"separator": separator}).
		Parse(codeTemplate))
	for _, node := range designerMeta.Data.Nodes {
		cell := TemplateImage{
			Name:       strings.Replace(node.Label, "-", "_", -1),
			Type:       node.Type,
			Components: nil,
		}
		for _, component := range node.Components {
			templateComponent := TemplateComponent{
				Name:  strings.Replace(component.Label, "-", "_", -1),
				Image: component.Image,
			}
			for alias, dependencyID := range component.Dependencies {
				dependency, err := getDependency(dependencyID, designerMeta)
				if err != nil {
					return err
				}
				dependency.Alias = strings.Replace(alias, "-", "_", -1)
				templateComponent.Dependencies = append(templateComponent.Dependencies, dependency)
			}
			cell.Components = append(cell.Components, templateComponent)
		}
		outputFile := filepath.Join(tmpSourceDir, cell.Name+constants.BalExt)
		f, err := os.Create(outputFile)
		if err != nil {
			log.Println("error occurred while creating the file: "+outputFile, err)
			return err
		}
		err = tpl.Execute(f, cell)
		if err != nil {
			log.Println("error occurred while executing the template: ", err)
			return err
		}
	}
	if err := util.RecursiveZip(nil, []string{tmpSourceDir}, filepath.Join(tmpZipDir, "downloads.zip")); err != nil {
		log.Println("error occurred while creating the zip: ", err)
		return err
	}
	return nil
}

// Template helper function to print commas(,)
func separator(s string) func() string {
	i := -1
	return func() string {
		i++
		if i == 0 {
			return ""
		}
		return s
	}
}

func getDependency(nodeId string, meta DesignMeta) (TemplateDependency, error) {
	for _, node := range meta.Data.Nodes {
		if nodeId == node.Gateway.ID {
			dependency := TemplateDependency{
				Org:  node.Org,
				Name: node.Version,
				Ver:  node.Label,
				Type: constants.CELL,
			}
			return dependency, nil
		}
		for _, component := range node.Components {
			if nodeId == component.ID {
				dependency := TemplateDependency{
					Org:  node.Org,
					Name: node.Version,
					Ver:  node.Label,
					Type: constants.COMPOSITE,
				}
				return dependency, nil
			}
		}
	}
	return TemplateDependency{}, fmt.Errorf("error while generationg code. Invalid dependency node id " + nodeId)
}
