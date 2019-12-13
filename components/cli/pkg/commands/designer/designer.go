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
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

// Starts a http server to serve Cellery designer react app and API
func RunDesigner(cli cli.Cli) error {
	r := mux.NewRouter()

	// It's important that this is before your catch-all route ("/")
	api := r.PathPrefix("/api/").Subrouter()
	api.HandleFunc("/load", loadCodeHandler).Methods(http.MethodGet)
	api.HandleFunc("/save", saveCodeHandler).Methods(http.MethodPost)
	api.HandleFunc("/generate", generateCodeHandler).Methods(http.MethodPost)

	// Serve static assets directly.
	designerDir := path.Join(cli.FileSystem().CelleryInstallationDir(), "designer")
	r.PathPrefix("/static").Handler(http.FileServer(http.Dir(designerDir)))

	// Catch-all: Serve Designer app entry-point (index.html).
	r.PathPrefix("/").HandlerFunc(indexHandler(path.Join(designerDir, "index.html")))

	srv := &http.Server{
		Handler:      handlers.LoggingHandler(os.Stdout, r),
		Addr:         "127.0.0.1:8088",
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

func loadCodeHandler(w http.ResponseWriter, r *http.Request) {
	//Read content from the give path and send as a JSON response
	var fileContent []byte
	var nodeData Data
	metaJsonPath := r.URL.Query().Get("path")
	fileContent, err := ioutil.ReadFile(metaJsonPath)
	if err != nil {
		log.Println("Error while reading json file: "+metaJsonPath, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(fileContent, &nodeData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	designerJSON, err := json.Marshal(&nodeData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(designerJSON)
}

func saveCodeHandler(w http.ResponseWriter, r *http.Request) {
	var nodeData DesignMeta

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(r.Body).Decode(&nodeData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//Extract the Data from body and save to the given path.
	nodeDataJson, err := json.Marshal(nodeData.Data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := ioutil.WriteFile(nodeData.Path, nodeDataJson, 0644); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(nodeData.Path))
}

func generateCodeHandler(w http.ResponseWriter, r *http.Request) {
	var nodeData DesignMeta

	// Try to decode the request body into the struct. If there is an error,
	// respond to the client with the error message and a 400 status code.
	err := json.NewDecoder(r.Body).Decode(&nodeData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := generateCode(nodeData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(nodeData.Path))
}

func generateCode(designerMeta DesignMeta) error {
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
		tpl := template.Must(template.New("cell template").Funcs(template.FuncMap{"separator": separator}).
			Parse(codeTemplate))

		dirExist, _ := util.FileExists(designerMeta.Path)
		if !dirExist {
			err := os.MkdirAll(designerMeta.Path, os.ModePerm)
			if err != nil {
				log.Println("Error while creating the directory: "+designerMeta.Path, err)
				return err
			}
		}
		f, err := os.Create(filepath.Join(designerMeta.Path, cell.Name+".bal"))
		if err != nil {
			log.Println("Error while creating the file: "+filepath.Join(designerMeta.Path, cell.Name+".bal"), err)
			return err
		}
		err = tpl.Execute(f, cell)
		if err != nil {
			log.Println("Error while executing the template: ", err)
			return err
		}
	}
	return nil
}

// Template function to print commas(,)
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
