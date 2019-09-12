///*
// * Copyright (c) 2019 WSO2 Inc. (http:www.wso2.org) All Rights Reserved.
// *
// * WSO2 Inc. licenses this file to you under the Apache License,
// * Version 2.0 (the "License"); you may not use this file except
// * in compliance with the License.
// * You may obtain a copy of the License at
// *
// * http:www.apache.org/licenses/LICENSE-2.0
// *
// * Unless required by applicable law or agreed to in writing,
// * software distributed under the License is distributed on an
// * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// * KIND, either express or implied.  See the License for the
// * specific language governing permissions and limitations
// * under the License.
// */

package commands

import (
	"fmt"
	"os"

	"github.com/cellery-io/sdk/components/cli/pkg/kubectl"

	errorpkg "github.com/cellery-io/sdk/components/cli/pkg/error"
	"github.com/cellery-io/sdk/components/cli/pkg/routing"
	"github.com/cellery-io/sdk/components/cli/pkg/util"
)

func RunRouteTrafficCommand(sourceInstances []string, dependencyInstance string, targetInstance string, percentage int,
	enableUserBasedSessionAwareness bool) error {
	spinner := util.StartNewSpinner(fmt.Sprintf("Starting to route %d%% of traffic to instance %s", percentage,
		targetInstance))

	// check the source instance and see if the dependency exists in the source
	routes, err := routing.GetRoutes(sourceInstances, dependencyInstance, targetInstance)
	if err != nil {
		spinner.Stop(false)
		return err
	}
	// now we have the source instance list which actually depend on the given dependency instance.
	// get the virtual services corresponding to the given source instances and modify accordingly.
	if len(routes) == 0 {
		// no depending instances
		spinner.Stop(false)
		return fmt.Errorf("cell/composite instance %s not found among dependencies of source instance(s)",
			dependencyInstance)
	}
	artifactFile := fmt.Sprintf("./%s-routing-artifacts.yaml", dependencyInstance)
	defer func() {
		_ = os.Remove(artifactFile)
	}()
	spinner.SetNewAction("Building modified rules")
	for _, route := range routes {
		err := route.Check()
		if err != nil {
			// if this is a CellGwApiVersionMismatchError, need to print a warning and prompt user for action
			if versionErr, match := err.(errorpkg.CellGwApiVersionMismatchError); match {
				spinner.Pause()
				canContinue, err := canContinueWithWarning(versionErr.CurrentTargetApiContext, versionErr.CurrentTargetApiVersion)
				spinner.Resume()
				if err != nil {
					spinner.Stop(false)
					return err
				}
				if !canContinue {
					spinner.Stop(false)
					return err
				}
			} else {
				spinner.Stop(false)
				return err
			}
		}
		err = route.Build(percentage, enableUserBasedSessionAwareness, artifactFile)
		if err != nil {
			spinner.Stop(false)
			return err
		}
	}

	spinner.SetNewAction("Applying modified rules")
	// perform kubectl apply
	err = kubectl.ApplyFile(artifactFile)
	if err != nil {
		spinner.Stop(false)
		return err
	}

	spinner.Stop(true)
	util.PrintSuccessMessage(fmt.Sprintf("Successfully routed %d%% of traffic to instance %s", percentage,
		targetInstance))
	return nil
}

func canContinueWithWarning(currCtxt string, currVersion string) (bool, error) {
	util.PrintWarningMessage(fmt.Sprintf("No API with matching version found in target instance for context: %s, version: %s \n", currCtxt, currVersion))
	canContinue, _, err := util.GetYesOrNoFromUser("Continue traffic routing", false)
	if err != nil {
		return false, err
	}
	return canContinue, nil
}
