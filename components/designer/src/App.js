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

import DesignerView from "./components/DesignerView";
import CssBaseline from "@material-ui/core/CssBaseline/CssBaseline";
import React from "react";
import {MuiThemeProvider, createMuiTheme} from "@material-ui/core/styles";
import {makeStyles} from "@material-ui/styles";
import * as PropTypes from "prop-types";

const useStyles = makeStyles({
    root: {
        flexGrow: 1,
        background: "#fff"
    }
});

const theme = createMuiTheme({
    typography: {
        useNextVariants: true
    },
    palette: {
        primary: {
            main: "#69b26d"
        }
    }
});

const App = () => {

    const classes = useStyles();
    return (
        <MuiThemeProvider theme={theme}>
            <CssBaseline/>
            <div className={classes.root}>
                <DesignerView/>
            </div>
        </MuiThemeProvider>
    );
};

App.propTypes = {
    classes: PropTypes.object.isRequired
};

export default App;
