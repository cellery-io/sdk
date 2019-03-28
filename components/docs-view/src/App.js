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
    const diagramData = {
        cells: [data.name],
        components: data.components.map((component) => (
            {
                cell: data.name,
                name: component
            }
        )),
        dependencyLinks: []
    };

    // Recursively extract dependencies (including transitive dependencies if available)
    const extractData = (cell) => {
        if (cell.dependencies) {
            Object.keys(cell.dependencies).forEach((alias) => {
                const dependency = cell.dependencies[alias];
                diagramData.cells.push(dependency.name);
                diagramData.dependencyLinks.push({
                    from: cell.name,
                    to: dependency.name
                });

                if (dependency.components) {
                    dependency.components.forEach((component) => {
                        diagramData.components.push({
                            cell: dependency.name,
                            name: component
                        });
                    });
                }

                extractData(dependency);
            });
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
                <CellDiagramView data={diagramData} focusedCell={data.name}/>
            </div>
        </MuiThemeProvider>
    );
};

App.propTypes = {
    data: PropTypes.object.isRequired,
    classes: PropTypes.object.isRequired
};

export default withStyles(styles)(App);
