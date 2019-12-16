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

type DesignMeta struct {
	Path string `json:"path"`
	Data Data   `json:"data"`
}
type Components struct {
	Label        string            `json:"label"`
	ID           string            `json:"id"`
	X            int               `json:"x"`
	Y            int               `json:"y"`
	Shape        string            `json:"shape"`
	Value        int               `json:"value"`
	Iterator     int               `json:"iterator"`
	Type         string            `json:"type"`
	Image        string            `json:"image"`
	Parent       string            `json:"parent"`
	SourceImage  string            `json:"sourceImage"`
	Dependencies map[string]string `json:"dependencies"`
}
type Gateway struct {
	Label    string `json:"label"`
	ID       string `json:"id"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
	Shape    string `json:"shape"`
	Value    int    `json:"value"`
	Iterator int    `json:"iterator"`
	Type     string `json:"type"`
	Parent   string `json:"parent"`
}
type Nodes struct {
	Label      string       `json:"label"`
	ID         string       `json:"id"`
	X          int          `json:"x"`
	Y          int          `json:"y"`
	Shape      string       `json:"shape"`
	Iterator   int          `json:"iterator"`
	Value      int          `json:"value"`
	Type       string       `json:"type"`
	Image      string       `json:"image"`
	Org        string       `json:"org"`
	Version    string       `json:"version"`
	Components []Components `json:"components"`
	Gateway    Gateway      `json:"gateway"`
	Name       string       `json:"name,omitempty"`
}

type Data struct {
	Nodes []Nodes       `json:"nodes"`
	Edges []interface{} `json:"edges"`
}

type TemplateDependency struct {
	Org   string
	Name  string
	Ver   string
	Alias string
	Type  string
}

type TemplateComponent struct {
	Name         string
	Image        string
	Dependencies []TemplateDependency
}

type TemplateImage struct {
	Name       string
	Type       string
	Components []TemplateComponent
}
