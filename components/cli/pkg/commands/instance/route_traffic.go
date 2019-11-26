/*
 * Copyright (c) 2019 WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
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

package instance

import (
	"fmt"
	"os"

	"github.com/cellery-io/sdk/components/cli/cli"
	errorpkg "github.com/cellery-io/sdk/components/cli/pkg/error"
	"github.com/cellery-io/sdk/components/cli/pkg/routing"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func RunRouteTrafficCommand(cli cli.Cli, sourceInstances []string, dependencyInstance string, targetInstance string, percentage int,
	enableUserBasedSessionAwareness bool, assumeYes bool) error {
	var err error
	artifactFile := fmt.Sprintf("./%s-routing-artifacts.yaml", dependencyInstance)
	defer func() error {
		return os.Remove(artifactFile)
	}()
	BuildRouteArtifact(cli, sourceInstances, dependencyInstance, targetInstance, percentage,
		enableUserBasedSessionAwareness, assumeYes)

	if err = cli.ExecuteTask("Applying modified rules", "Failed to apply modified rules", "", func() error {
		err = cli.KubeCli().ApplyFile(artifactFile)
		if err != nil {
			return fmt.Errorf("error occurred while applying modified rules, %v", err)
		}
		return nil
	}); err != nil {
		return err
	}
	util.PrintSuccessMessage(fmt.Sprintf("Successfully routed %d%% of traffic to instance %s", percentage,
		targetInstance))
	return nil
}

func BuildRouteArtifact(cli cli.Cli, sourceInstances []string, dependencyInstance string, targetInstance string, percentage int,
	enableUserBasedSessionAwareness bool, assumeYes bool) error {
	fmt.Fprintln(cli.Out(), fmt.Sprintf("Starting to route %d%% of traffic to instance %s", percentage,
		targetInstance))

	// check the source instance and see if the dependency exists in the source
	routes, err := routing.GetRoutes(cli, sourceInstances, dependencyInstance, targetInstance)
	if err != nil {
		return err
	}
	// now we have the source instance list which actually depend on the given dependency instance.
	// get the virtual services corresponding to the given source instances and modify accordingly.
	if len(routes) == 0 {
		// no depending instances
		return fmt.Errorf("cell/composite instance %s not found among dependencies of source instance(s)",
			dependencyInstance)
	}
	artifactFile := fmt.Sprintf("./%s-routing-artifacts.yaml", dependencyInstance)
	for _, route := range routes {
		err := route.Check()
		if err != nil {
			// if this is a CellGwApiVersionMismatchError, need to print a warning and prompt user for action
			if versionErr, match := err.(errorpkg.CellGwApiVersionMismatchError); match {
				if !assumeYes {
					// prompt confirmation from user
					canContinue, err := canContinueWithWarning(versionErr.ApiContext, versionErr.CurrentTargetApiVersion, versionErr.NewTargetApiVersion)
					if err != nil {
						return err
					}
					if !canContinue {
						fmt.Fprintln(cli.Out(), "Aborting traffic routing")
						return nil
					}
				}
			} else {
				return err
			}
		}
		if err = cli.ExecuteTask("Building modified rules", "Failed to build modified rules", "", func() error {
			return route.Build(cli, percentage, enableUserBasedSessionAwareness, artifactFile)
			if err != nil {
				return fmt.Errorf("error occurred while building modified rules, %v", err)
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

func canContinueWithWarning(context string, currVersion string, newVersion string) (bool, error) {
	var warnMsg string
	if newVersion != "" {
		warnMsg = fmt.Sprintf("No matching API found in target for context: '%s', version: '%s'. Available version in target: '%s' \n", context, currVersion, newVersion)
	} else {
		warnMsg = fmt.Sprintf("No matching API found in target for context: '%s', version: '%s' \n", context, currVersion)
	}
	util.PrintWarningMessage(warnMsg)
	canContinue, _, err := util.GetYesOrNoFromUser("Continue traffic routing", false)
	if err != nil {
		return false, err
	}
	return canContinue, nil
}
