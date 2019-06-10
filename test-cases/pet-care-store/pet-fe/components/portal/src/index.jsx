/*
 * Copyright (c) 2019, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
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

import App from "./components/App";
import {BrowserRouter} from "react-router-dom";
import {CssBaseline} from "@material-ui/core";
import {JssProvider} from "react-jss";
import React from "react";
import ReactDOM from "react-dom";
import {StateProvider} from "./components/common/state";
import {generateTheme} from "./utils";
import {MuiThemeProvider, createGenerateClassName} from "@material-ui/core/styles";

const initialState = window.__INITIAL_STATE__ ? window.__INITIAL_STATE__ : {};
Reflect.deleteProperty(window, "__INITIAL_STATE__");

if (!window.__BASE_PATH__) {
    window.__BASE_PATH__ = "";
} else if (window.__BASE_PATH__ === "/") {
    window.__BASE_PATH__ = ".";
} else if (window.__BASE_PATH__.endsWith("/")) {
    window.__BASE_PATH__ = window.__BASE_PATH__.substr(0, window.__BASE_PATH__.length - 1);
}

class Main extends React.Component {

    componentDidMount() {
        // Remove the server-side injected CSS.
        const jssStyles = document.getElementById("jss-server-side");
        if (jssStyles && jssStyles.parentNode) {
            jssStyles.parentNode.removeChild(jssStyles);
        }
    }

    render() {
        return <App/>;
    }

}

ReactDOM.hydrate((
    <JssProvider generateClassName={createGenerateClassName()}>
        <MuiThemeProvider theme={generateTheme()}>
            <CssBaseline/>
            <BrowserRouter basename={window.__BASE_PATH__}>
                <StateProvider catalog={initialState.catalog} user={initialState.user}>
                    <Main/>
                </StateProvider>
            </BrowserRouter>
        </MuiThemeProvider>
    </JssProvider>
), document.getElementById("app"));
