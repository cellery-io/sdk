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

const codeTemplate = `
import celleryio/cellery;

public function build(cellery:ImageName iName) returns error? {
    {{$s := separator ", "}}
    {{range $index, $value := .Components}}
    cellery:Component {{$value.Name}} = {
        name: "{{$value.Name}}",
        src: {
            image: "{{$value.Image}}"
        },
        ingresses: {},
        envVars: {},
        dependencies: {
          cells: {
            {{ range $index, $value := $value.Dependencies -}}{{if eq $value.Type "cell" -}}{{call $s}}
            {{$value.Alias}}: { org:"{{$value.Org}}", name:"{{$value.Name}}", ver:"{{$value.Ver}}" }{{end -}}
            {{end }}
            },
            composites: {
            {{ range $index, $value := $value.Dependencies -}}{{if eq $value.Type "composite" -}}{{call $s}}
            {{$value.Alias}}: { org:"{{$value.Org}}", name:"{{$value.Name}}", ver:"{{$value.Ver}}" }{{end -}}
            {{end }}
            }
        }
    };
    {{end}}

    {{if eq .Type "cell" -}}
    // Cell Initialization
    {{$s = separator ", " -}}
    cellery:CellImage {{.Name}} = {
        components: {
        {{range $index, $value := .Components -}}{{call $s}}
        {{$value.Name}}Comp: {{$value.Name}}{{end }}
        }
    };
    {{else -}}
    // Composite Initialization
    {{$s = separator ", "}}
    cellery:Composite {{.Name}} = {
    components: {
        {{range $index, $value := .Components -}}{{call $s}}
        {{$value.Name}}Comp: {{$value.Name}}{{end }}
        }
    };
    {{end -}}
    return <@untainted> cellery:createImage({{.Name}}, iName);
}



public function run(cellery:ImageName iName, map<cellery:ImageName> instances, boolean startDependencies, boolean shareDependencies) returns (cellery:InstanceState[]|error?) {
    cellery:CellImage|cellery:Composite {{.Name}} = cellery:constructImage(<@untainted> iName);
    return <@untainted> cellery:createInstance({{.Name}}, iName, instances, startDependencies, shareDependencies);
}
`
