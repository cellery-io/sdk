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

import {CircularProgress} from "@material-ui/core";
import React from "react";
import {withStyles} from "@material-ui/core/styles";
import * as PropTypes from "prop-types";

const styles = (theme) => ({
    toolbar: {
        display: "flex",
        alignItems: "center",
        justifyContent: "flex-end",
        padding: "0 8px",
        ...theme.mixins.toolbar
    },
    progressOverlayContainer: {
        position: "absolute",
        zIndex: 9999,
        top: 0,
        left: 0,
        width: "100%",
        height: "100%"
    },
    progressOverlay: {
        position: "relative",
        display: "grid",
        top: 0,
        left: 0,
        width: "100%",
        height: "100%",
        backgroundColor: "rgb(0, 0, 0, 0.5)"
    },
    progress: {
        textAlign: "center",
        margin: "auto"
    },
    progressIndicator: {
        margin: theme.spacing.unit * 2
    },
    progressContent: {
        fontSize: "large",
        fontWeight: 500,
        width: "100%",
        color: "#ffffff"
    }
});

const ProgressIndicator = ({classes, message}) => (
    <div className={classes.progressOverlayContainer}>
        <div className={classes.toolbar}/>
        <div className={classes.progressOverlay}>
            <div className={classes.progress}>
                <CircularProgress className={classes.progressIndicator} thickness={4.5} size={45}/>
                <div className={classes.progressContent}>
                    {message ? message : "Loading"}...
                </div>
            </div>
        </div>
    </div>
);

ProgressIndicator.propTypes = {
    classes: PropTypes.object.isRequired,
    message: PropTypes.string
};

export default withStyles(styles)(ProgressIndicator);
