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

import Notification from "../common/Notification";
import ProgressIndicator from "../common/ProgressIndicator";
import React from "react";
import * as utils from "../../utils";

class SignIn extends React.Component {

    constructor(props) {
        super(props);
        this.state = {
            notification: {
                open: false,
                message: ""
            }
        };
    }

    handleNotificationClose = () => {
        this.setState({
            notification: {
                open: false,
                message: ""
            }
        });
    };

    componentDidMount() {
        const self = this;
        const {history} = self.props;
        const config = {
            url: "/profile",
            method: "GET"
        };
        utils.callApi(config)
            .then((response) => {
                if (response.data.profile) {
                    history.replace("/");
                } else {
                    history.replace("/sign-up");
                }
            })
            .catch(() => {
                self.setState({
                    notification: {
                        open: true,
                        message: "Failed to check if profile exists"
                    }
                });
                history.replace("/");
            });
    }

    render() {
        const {notification} = this.state;
        return (
            <React.Fragment>
                <ProgressIndicator/>
                <Notification open={notification.open} onClose={this.handleNotificationClose}
                    message={notification.message}/>
            </React.Fragment>
        );
    }

}

export default SignIn;
