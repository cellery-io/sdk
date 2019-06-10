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

import {Close} from "@material-ui/icons";
import React from "react";
import {IconButton, Snackbar} from "@material-ui/core";
import * as PropTypes from "prop-types";

const Notification = ({message, onClose, open}) => (
    <Snackbar open={open} ContentProps={{"aria-describedby": "notification"}}
        anchorOrigin={{
            vertical: "bottom",
            horizontal: "left"
        }}
        onClose={onClose} message={message} autoHideDuration={5000}
        action={[
            <IconButton key="close" aria-label="Close" color="inherit" onClick={onClose}>
                <Close/>
            </IconButton>
        ]}
    />
);

Notification.propTypes = {
    open: PropTypes.bool.isRequired,
    message: PropTypes.string.isRequired,
    onClose: PropTypes.func.isRequired
};

export default Notification;
