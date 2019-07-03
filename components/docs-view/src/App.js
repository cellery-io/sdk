/*
 * Copyright (c) 2019, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import AppBar from "@material-ui/core/AppBar";
import CellDiagramView from "./components/CellDiagramView";
import CelleryLogo from "./icons/CelleryLogo";
import CssBaseline from "@material-ui/core/CssBaseline/CssBaseline";
import React from "react";
import Toolbar from "@material-ui/core/Toolbar";
import Typography from "@material-ui/core/Typography";
import {MuiThemeProvider, createMuiTheme, withStyles} from "@material-ui/core/styles";
import * as PropTypes from "prop-types";

const styles = {
    root: {
        flexGrow: 1,
        background: "#fff"
    },
    celleryLogo: {
        width: 100,
        marginRight: 10
    },
    title: {
        color: "#464646"
    },
    appBar: {
        boxShadow: "none"
    }
};

// Create the main theme of the App
const theme = createMuiTheme({
    typography: {
        useNextVariants: true
    },
    palette: {
        primary: {
            main: "#43ab01"
        }
    }
});

const App = ({data, classes}) => {
    const getCellName = (cellData) => `${cellData.org}/${cellData.name}:${cellData.ver}`;
    const diagramData = {
        cells: [getCellName(data)],
        components: data.components.map((component) => (
            {
                cell: getCellName(data),
                name: component
            }
        )),
        metaInfo: {},
        dependencyLinks: []
    };

    // Recursively extract dependencies (including transitive dependencies if available)
    const extractData = (cell) => {
        if (cell.dependencies) {
            Object.entries(cell.dependencies).forEach(([alias, dependency]) => {
                const dependencyName = getCellName(dependency);
                if (!diagramData.cells.includes(dependencyName)) {
                    diagramData.cells.push(dependencyName);
                }
                diagramData.dependencyLinks.push({
                    alias: alias,
                    from: getCellName(cell),
                    to: dependencyName
                });

                if (dependency.components) {
                    dependency.components.forEach((component) => {
                        const matches = diagramData.components.find(
                            (datum) => datum.cell === dependencyName && datum.name === component);
                        if (!matches) {
                            diagramData.components.push({
                                cell: dependencyName,
                                name: component
                            });
                        }
                    });
                }
                extractData(dependency);
            });
        }

        if (!diagramData.metaInfo.hasOwnProperty(getCellName(cell))) {
            const cellMetaInfo = {
                cell: getCellName(cell),
                ingresses: [],
                componentDependencyLinks: []
            };

            if (cell.componentDep) {
                Object.entries(cell.componentDep).forEach(([component, dependency]) => {
                    dependency.forEach((dependentComponent) => {
                        cellMetaInfo.componentDependencyLinks.push({
                            from: `${getCellName(cell)} ${component}`,
                            to: `${getCellName(cell)} ${dependentComponent}`
                        });
                    });
                });
            }

            if (cell.exposed) {
                cell.exposed.forEach((component) => {
                    cellMetaInfo.componentDependencyLinks.push({
                        from: `${getCellName(cell)} gateway`,
                        to: `${getCellName(cell)} ${component}`
                    });
                });
            }

            if (cell.ingresses) {
                cellMetaInfo.ingresses = cell.ingresses;
            }
            diagramData.metaInfo[getCellName(cell)] = cellMetaInfo;
        }
    };
    extractData(data);

    return (
        <MuiThemeProvider theme={theme}>
            <CssBaseline/>
            <div className={classes.root}>
                <AppBar position="static" color="default" className={classes.appBar}>
                    <Toolbar>
                        <CelleryLogo className={classes.celleryLogo} fontSize="large"/>
                        <Typography variant="h6" className={classes.title}>
                                Image View
                        </Typography>
                    </Toolbar>
                </AppBar>
                <CellDiagramView data={diagramData} focusedCell={getCellName(data)}/>
            </div>
        </MuiThemeProvider>
    );
};

App.propTypes = {
    data: PropTypes.object.isRequired,
    classes: PropTypes.object.isRequired
};

export default withStyles(styles)(App);
