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
import React from "react";
import {withRouter} from "react-router-dom";
import {
    Button, Checkbox, FormControl, FormControlLabel, FormGroup, Grid, Step, StepContent, StepLabel, Stepper, TextField
} from "@material-ui/core";
import * as PropTypes from "prop-types";
import * as utils from "../../utils";

class SignUp extends React.Component {

    constructor(props) {
        super(props);
        this.state = {
            profile: props.profile ? props.profile : {pets: []},
            activeStep: 0,
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

    handleBack = () => {
        this.setState((prevState) => ({
            activeStep: prevState.activeStep - 1
        }));
    };

    handleNext = () => {
        this.setState((prevState) => ({
            activeStep: prevState.activeStep + 1
        }));
    };

    handleFinish = () => {
        const self = this;
        const {history} = self.props;
        const {profile} = self.state;
        const config = {
            url: "/profile",
            method: "POST",
            data: profile
        };
        utils.callApi(config)
            .then(() => {
                history.replace("/");
                self.setState({
                    notification: {
                        open: true,
                        message: "Successfully created your profile"
                    }
                });
            })
            .catch(() => {
                self.setState({
                    notification: {
                        open: true,
                        message: "Failed to create profile"
                    }
                });
            });
    };

    handleChange = (name) => (event) => {
        const value = event.currentTarget.value;
        this.setState((prevState) => ({
            profile: {
                ...prevState.profile,
                [name]: value
            }
        }));
    };

    handlePetsChange = (pet) => (event) => {
        const checked = event.currentTarget.checked;
        this.setState((prevState) => {
            const pets = prevState.profile.pets;
            if (checked) {
                pets.push(pet);
            } else {
                pets.splice(pets.indexOf(pet), 1);
            }
            return {
                profile: {
                    ...prevState.profile,
                    pets: pets
                }
            };
        });
    };

    render() {
        const {activeStep, notification, profile} = this.state;
        const allowedPets = [
            "Dog", "Cat", "Hamster"
        ];
        return (
            <Grid container justify={"center"}>
                <Grid item lg={4} md={8} xs={12}>
                    <Stepper activeStep={activeStep} orientation={"vertical"}>
                        <Step>
                            <StepLabel>Personal Information</StepLabel>
                            <StepContent>
                                <FormControl fullWidth={true}>
                                    <TextField required id={"first-name"} label={"First Name"} margin={"normal"}
                                        value={profile.firstName} onChange={this.handleChange("firstName")}/>
                                </FormControl>
                                <FormControl fullWidth={true}>
                                    <TextField required id={"last-name"} label={"Last Name"} margin={"normal"}
                                        value={profile.lastName} onChange={this.handleChange("lastName")}/>
                                </FormControl>
                                <FormControl fullWidth={true}>
                                    <TextField required multiline id={"address"} label={"Address"} margin={"normal"}
                                        value={profile.address} onChange={this.handleChange("address")}/>
                                </FormControl>
                                <div>
                                    <Button disabled onClick={this.handleBack}>Back</Button>
                                    <Button disabled={!profile.firstName || !profile.lastName || !profile.address}
                                        variant={"contained"} color={"primary"} onClick={this.handleNext}>
                                        Next
                                    </Button>
                                </div>
                            </StepContent>
                        </Step>
                        <Step>
                            <StepLabel>Choose your Pets</StepLabel>
                            <StepContent>
                                <FormGroup>
                                    {
                                        allowedPets.map((pet) => (
                                            <FormControlLabel key={pet} label={pet}
                                                control={
                                                    <Checkbox checked={profile.pets.includes(pet)}
                                                        onChange={this.handlePetsChange(pet)} value={pet}/>
                                                }
                                            />
                                        ))
                                    }
                                </FormGroup>
                                <div>
                                    <Button onClick={this.handleBack}>Back</Button>
                                    <Button disabled={profile.pets.length === 0} variant={"contained"} color={"primary"}
                                        onClick={this.handleFinish}>
                                        Continue
                                    </Button>
                                </div>
                            </StepContent>
                        </Step>
                    </Stepper>
                    <Notification open={notification.open} onClose={this.handleNotificationClose}
                        message={notification.message}/>
                </Grid>
            </Grid>
        );
    }

}

SignUp.propTypes = {
    profile: PropTypes.shape({
        username: PropTypes.string.isRequired,
        firstName: PropTypes.string.isRequired,
        lastName: PropTypes.string.isRequired,
        pets: PropTypes.arrayOf(PropTypes.string.isRequired)
    }),
    history: PropTypes.shape({
        replace: PropTypes.func.isRequired
    }).isRequired
};

export default withRouter(SignUp);
