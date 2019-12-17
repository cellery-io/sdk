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

import CelleryLogo from "../icons/CelleryLogo";
import AppBar from "@material-ui/core/AppBar";
import Button from "@material-ui/core/Button";
import IconButton from "@material-ui/core/IconButton/IconButton";
import React from "react";
import Toolbar from "@material-ui/core/Toolbar";
import Tooltip from "@material-ui/core/Tooltip/Tooltip";
import Typography from "@material-ui/core/Typography";
import {withStyles} from "@material-ui/core/styles";
import {SaveOutlined, OpenInBrowserRounded} from '@material-ui/icons';
import * as PropTypes from "prop-types";
import Files from "react-files";

const styles = {
    celleryLogo: {
        width: 70,
        marginRight: 5
    },
    title: {
        color: "#464646",
        flexGrow: 1,
        fontSize: 16
    },
    appBar: {
        boxShadow: "none"
    },
    headerBtn: {
        fontSize: "1rem"
    },
    toolbar: {
        minHeight: 35
    },
    primaryBtn: {
        color: "#fff",
        fontSize: 10,
        boxShadow: "none"
    },
    openFile: {
        display: "inline"
    }
};

class DesignerView extends React.Component {

    render = () => {
        const {classes, saveData, fileReader, validateGraph} = this.props;

        return (
            <AppBar position="static" color="default" className={classes.appBar}>
                <Toolbar className={classes.toolbar}>
                    <CelleryLogo className={classes.celleryLogo} fontSize="large"/>
                    <Typography variant="h6" className={classes.title}>
                        Designer
                    </Typography>
                    <div className={classes.actions}>
                        <span class="file-open">
                            <Files
                                className={classes.openFile}
                                onChange={file => {
                                    fileReader.readAsText(file[0]);
                                }}
                                onError={err => console.log(err)}
                                accepts={[".json"]}
                                maxFileSize={10000000}
                                minFileSize={0}
                                clickable
                            >
                            <Tooltip title="Open" placement="bottom">
                            <IconButton color="inherit" className={classes.headerBtn}>
                                <OpenInBrowserRounded fontSize="small"/>
                            </IconButton>
                            </Tooltip>
                        </Files>
                        </span>
                        <Tooltip title="Save" placement="bottom">
                            <IconButton color="inherit" fontSize="small" onClick={saveData}>
                                <SaveOutlined fontSize="small"/>
                            </IconButton>
                        </Tooltip>
                        <Button variant="contained" className={classes.primaryBtn} color="primary" size="small"
                                onClick={validateGraph}>
                            Generate Code
                        </Button>
                    </div>
                </Toolbar>
            </AppBar>
        );
    };

}

DesignerView.propTypes = {
    classes: PropTypes.object.isRequired
};

export default withStyles(styles)(DesignerView);
