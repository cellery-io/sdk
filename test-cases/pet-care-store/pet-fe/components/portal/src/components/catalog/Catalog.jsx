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

import {AddShoppingCart} from "@material-ui/icons";
import AddToCartPopup from "./AddToCartPopup";
import React from "react";
import classNames from "classnames";
import {withRouter} from "react-router-dom";
import withState from "../common/state";
import {withStyles} from "@material-ui/core/styles";
import {Button, Card, CardContent, CardMedia, Grid, Typography} from "@material-ui/core";
import * as PropTypes from "prop-types";

const styles = (theme) => ({
    grow: {
        flexGrow: 1
    },
    root: {
        backgroundColor: theme.palette.background.paper
    },
    titleContent: {
        maxWidth: 600,
        margin: "0 auto",
        padding: `${theme.spacing.unit * 8}px 0 ${theme.spacing.unit * 6}px`
    },
    titleButtons: {
        marginTop: theme.spacing.unit * 4
    },
    layout: {
        width: "auto",
        marginLeft: theme.spacing.unit * 3,
        marginRight: theme.spacing.unit * 3,
        [theme.breakpoints.up(1100 + (theme.spacing.unit * 3 * 2))]: {
            width: 1100,
            marginLeft: "auto",
            marginRight: "auto"
        }
    },
    cardGrid: {
        padding: `${theme.spacing.unit * 8}px 0`
    },
    card: {
        height: "100%",
        display: "flex",
        flexDirection: "column"
    },
    cardMedia: {
        paddingTop: "56.25%" // 16:9
    },
    cardContent: {
        flexGrow: 0
    },
    priceTag: {
        textAlign: "right"
    },
    catalogActionPane: {
        marginTop: theme.spacing.unit * 3
    }
});

class Catalog extends React.Component {

    constructor(props) {
        super(props);
        this.state = {
            addToCartPopup: {
                open: false,
                item: null
            }
        };
    }

    handleOpenAddToCartPopup = (item) => () => {
        this.setState({
            addToCartPopup: {
                open: true,
                item: item
            }
        });
    };

    handleCloseAddToCartPopup = () => {
        this.setState({
            addToCartPopup: {
                open: false,
                item: null
            }
        });
    };

    render() {
        const {catalog, classes, history, user} = this.props;
        const {addToCartPopup} = this.state;
        return (
            <div className={classes.root}>
                <div className={classes.titleContent}>
                    <Typography component="h1" variant="h2" align="center" color="textPrimary" gutterBottom>
                        Pet Accessories
                    </Typography>
                    <Typography variant="h6" align="center" color="textSecondary" paragraph>
                        Buy accessories for your Pets
                    </Typography>
                    <div className={classes.titleButtons}>
                        <Grid container spacing={16} justify="center">
                            <Grid item>
                                {
                                    user
                                        ? (
                                            <Button variant="contained" color="primary" onClick={() => {
                                                history.push("/orders");
                                            }}>
                                                Check My Orders
                                            </Button>
                                        )
                                        : null
                                }
                            </Grid>
                        </Grid>
                    </div>
                </div>
                <div className={classNames(classes.layout, classes.cardGrid)}>
                    {
                        catalog.accessories.length > 0
                            ? (
                                <Grid container spacing={40}>
                                    {
                                        catalog.accessories.map((item, index) => (
                                            <Grid key={index} item sm={6} md={4} lg={3}>
                                                <Card className={classes.card}>
                                                    <CardMedia
                                                        className={classes.cardMedia}
                                                        image={"./app/assets/catalog/"
                                                        + `${item.name.replace(/\s+/g, "-")
                                                            .toLowerCase()}.svg`}
                                                        title={item.name}
                                                    />
                                                    <CardContent className={classes.cardContent}>
                                                        <Typography gutterBottom variant="h5" component="h2">
                                                            {item.name}
                                                        </Typography>
                                                        <Typography>
                                                            {item.description}
                                                        </Typography>
                                                    </CardContent>
                                                    <div className={classes.grow}/>
                                                    <CardContent>
                                                        <Grid container direction={"row"} justify={"space-between"}
                                                            alignItems={"center"}>
                                                            <Grid item sm={6}>
                                                                <Typography color={"textSecondary"}>
                                                                    In Stock: {item.inStock}
                                                                </Typography>
                                                            </Grid>
                                                            <Grid item sm={6} className={classes.priceTag}>
                                                                <Typography color={"textSecondary"}>
                                                                    $ {item.unitPrice}
                                                                </Typography>
                                                            </Grid>
                                                        </Grid>
                                                        {
                                                            user
                                                                ? (
                                                                    <Grid container direction={"row"}
                                                                        className={classes.catalogActionPane}
                                                                        justify={"flex-end"} alignItems={"flex-end"}>
                                                                        <Button disabled={item.inStock === 0}
                                                                            color={"primary"} variant={"contained"}
                                                                            size={"small"}
                                                                            onClick={
                                                                                this.handleOpenAddToCartPopup(item)}>
                                                                            <AddShoppingCart/>
                                                                        </Button>
                                                                    </Grid>
                                                                )
                                                                : null
                                                        }
                                                    </CardContent>
                                                </Card>
                                            </Grid>
                                        ))
                                    }
                                </Grid>

                            )
                            : (
                                <Typography variant={"body1"} align={"center"} color={"textSecondary"}>
                                    No Accessories Available for Sale
                                </Typography>
                            )
                    }
                </div>
                {
                    addToCartPopup.item
                        ? (
                            <AddToCartPopup open={addToCartPopup.open} item={addToCartPopup.item}
                                onClose={this.handleCloseAddToCartPopup}/>
                        )
                        : null
                }
            </div>
        );
    }

}

Catalog.propTypes = {
    user: PropTypes.string.isRequired,
    classes: PropTypes.object.isRequired,
    history: PropTypes.shape({
        push: PropTypes.func.isRequired
    }).isRequired,
    catalog: PropTypes.shape({
        accessories: PropTypes.arrayOf(PropTypes.shape({
            id: PropTypes.number.isRequired,
            name: PropTypes.string.isRequired,
            description: PropTypes.string.isRequired,
            unitPrice: PropTypes.number.isRequired,
            inStock: PropTypes.number.isRequired
        })).isRequired
    })
};

export default withStyles(styles)(withRouter(withState(Catalog)));
