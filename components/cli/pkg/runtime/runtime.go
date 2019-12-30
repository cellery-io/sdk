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

package runtime

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/hashicorp/go-version"

	"cellery.io/cellery/components/cli/pkg/constants"
	"cellery.io/cellery/components/cli/pkg/kubernetes"
	"cellery.io/cellery/components/cli/pkg/osexec"
	"cellery.io/cellery/components/cli/pkg/util"
)

const kubernetesVersionMax = "v1.16.0"
const kubernetesVersionMaxExpected = "v1.15.x"
const kubernetesVersionMin = "v1.13"

type Selection int

const (
	NoChange Selection = iota
	Enable
	Disable
)

type Runtime interface {
	SetArtifactsPath(artifactsPath string)
	AddComponent(component SystemComponent) error
	DeleteComponent(component SystemComponent) error
	IsComponentEnabled(component SystemComponent) (bool, error)
	CreateCelleryNameSpace() error
	CreateNameSpace(namespace string) error
	IsGcpRuntime() bool
	CreatePersistentVolumeDirs() error
	UpdateNfsServerDetails(ipAddress, fileShare string) error
	UpdateMysqlCredentials(dbUserName, dbPassword, dbHost string) error
	UpdateInitSql(dbUserName, dbPassword string) error
	ApplyIstioCrds() error
	InstallIstio() error
	InstallIngressNginx(isLoadBalancerIngressMode bool, nodePortIpAddress string) error
	ApplyKnativeCrds() error
	InstallKnativeServing() error
	InstallController() error
	InstallMysql(isPersistentVolume bool) error
	CreateConfigMaps() error
	AddApim(isPersistentVolume bool, nfs Nfs, db MysqlDb) error
	DeleteApim() error
	AddObservability() error
	AddIdp(db MysqlDb) error
	DeleteIdp() error
	UpdateNodePortIpAddress(nodePortIpAddress string) error
	CreatePersistentVolume(hasNfs bool) error
	IsHpaEnabled() (bool, error)
	WaitFor(checkKnative, hpaEnabled bool) error
	Validate() error
	UnmarshalHelmValues(chartName string) error
	MarshalHelmValues(chartName string) error
	InstallHPA() error
}

type CelleryRuntime struct {
	artifactsPath          string
	nfs                    Nfs
	db                     MysqlDb
	celleryRuntimeVals     CelleryRuntimeVals
	celleryRuntimeYaml     string
	ingressControllerVals  IngressController
	ingressControllerYamls string
}

// NewCelleryRuntime returns a CelleryRuntime instance.
func NewCelleryRuntime(opts ...func(*CelleryRuntime)) *CelleryRuntime {
	runtime := &CelleryRuntime{}
	for _, opt := range opts {
		opt(runtime)
	}
	return runtime
}

func (runtime *CelleryRuntime) SetArtifactsPath(artifactsPath string) {
	runtime.artifactsPath = artifactsPath
}

func (runtime *CelleryRuntime) CreatePersistentVolumeDirs() error {
	// Create folders required by the mysql PVC
	if err := util.CreateDir(filepath.Join(constants.RootDir, constants.VAR, constants.TMP, constants.CELLERY, constants.MySql)); err != nil {
		return err
	}
	// Create folders required by the APIM PVC
	if err := util.CreateDir(filepath.Join(constants.RootDir, constants.VAR, constants.TMP, constants.CELLERY,
		constants.ApimRepositoryDeploymentServer)); err != nil {
		return err
	}
	return nil
}

func (runtime *CelleryRuntime) CreateConfigMaps() error {
	if err := createGlobalGatewayConfigMaps(runtime.artifactsPath); err != nil {
		return fmt.Errorf("error creating gateway configmaps: %v", err)
	}
	if err := createObservabilityConfigMaps(runtime.artifactsPath); err != nil {
		return fmt.Errorf("error creating observability configmaps: %v", err)
	}
	if err := createIdpConfigMaps(runtime.artifactsPath); err != nil {
		return fmt.Errorf("error creating idp configmaps: %v", err)
	}
	return nil
}

func (runtime *CelleryRuntime) AddComponent(component SystemComponent) error {
	switch component {
	case ApiManager:
		return runtime.AddApim(false, Nfs{}, MysqlDb{})
	case IdentityProvider:
		return runtime.AddIdp(MysqlDb{})
	case Observability:
		return runtime.AddObservability()
	case ScaleToZero:
		return runtime.InstallKnativeServing()
	case HPA:
		//return InstallHPA(filepath.Join(util.CelleryInstallationDir(), constants.K8sArtifacts))
		return runtime.InstallHPA()
	default:
		return fmt.Errorf("unknown system componenet %q", component)
	}
}

func (runtime *CelleryRuntime) DeleteComponent(component SystemComponent) error {
	switch component {
	case ApiManager:
		//return deleteApim(filepath.Join(util.CelleryInstallationDir(), constants.K8sArtifacts))
		return runtime.DeleteApim()
	case IdentityProvider:
		//return deleteIdp(filepath.Join(util.CelleryInstallationDir(), constants.K8sArtifacts))
		return runtime.DeleteIdp()
	case Observability:
		//return deleteObservability(filepath.Join(util.CelleryInstallationDir(), constants.K8sArtifacts))
		return runtime.DeleteObservability()
	case ScaleToZero:
		return deleteKnative()
	case HPA:
		//return deleteHpa(filepath.Join(util.CelleryInstallationDir(), constants.K8sArtifacts))
		return runtime.DeleteHPA()
	default:
		return fmt.Errorf("unknown system componenet %q", component)
	}
}

func (runtime *CelleryRuntime) IsComponentEnabled(component SystemComponent) (bool, error) {
	switch component {
	case ApiManager:
		return IsApimEnabled()
	case Observability:
		return IsObservabilityEnabled()
	case ScaleToZero:
		return IsKnativeEnabled()
	case HPA:
		return runtime.IsHpaEnabled()
	default:
		return false, fmt.Errorf("unknown system componenet %q", component)
	}
}

func buildArtifactsPath(component SystemComponent, artifactsPath string) string {
	switch component {
	case ApiManager:
		return filepath.Join(artifactsPath, "global-apim")
	case IdentityProvider:
		return filepath.Join(artifactsPath, "global-idp")
	case Observability:
		return filepath.Join(artifactsPath, "observability")
	case Controller:
		return filepath.Join(artifactsPath, "controller")
	case System:
		return filepath.Join(artifactsPath, "system")
	case Mysql:
		return filepath.Join(artifactsPath, "mysql")
	case HPA:
		return filepath.Join(artifactsPath, "metrics-server/")
	default:
		return filepath.Join(artifactsPath)
	}
}

func (runtime *CelleryRuntime) IsGcpRuntime() bool {
	nodes, err := kubernetes.GetNodes()
	if err != nil {
		util.ExitWithErrorMessage("failed to check if runtime is gcp", err)
	}
	for _, node := range nodes.Items {
		version := node.Status.NodeInfo.KubeletVersion
		if strings.Contains(version, "gke") {
			return true
		}
	}
	return false
}

func (runtime *CelleryRuntime) WaitFor(checkKnative, hpaEnabled bool) error {
	spinner := util.StartNewSpinner("Checking cluster status...")
	wtCluster, err := waitingTimeCluster()
	if err != nil {
		spinner.Stop(false)
		util.ExitWithErrorMessage("Error getting waiting time for cluster", err)
	}
	err = kubernetes.WaitForCluster(wtCluster)
	if err != nil {
		spinner.Stop(false)
		util.ExitWithErrorMessage("Error while checking cluster status", err)
	}
	spinner.SetNewAction("Cluster status...OK")
	spinner.Stop(true)

	spinner = util.StartNewSpinner("Checking runtime status (Istio)...")
	err = kubernetes.WaitForDeployments("istio-system", time.Minute*15)
	if err != nil {
		spinner.Stop(false)
		util.ExitWithErrorMessage("Error while checking runtime status (Istio)", err)
	}
	spinner.SetNewAction("Runtime status (Istio)...OK")
	spinner.Stop(true)

	if checkKnative {
		spinner = util.StartNewSpinner("Checking runtime status (Knative Serving)...")
		err = kubernetes.WaitForDeployments("knative-serving", time.Minute*15)
		if err != nil {
			spinner.Stop(false)
			util.ExitWithErrorMessage("Error while checking runtime status (Knative Serving)", err)
		}
		spinner.SetNewAction("Runtime status (Knative Serving)...OK")
		spinner.Stop(true)
	}

	if hpaEnabled {
		spinner = util.StartNewSpinner("Checking runtime status (Metrics server)...")
		err = kubernetes.WaitForDeployment("available", 900, "metrics-server", "kube-system")
		if err != nil {
			spinner.Stop(false)
			util.ExitWithErrorMessage("Error while checking runtime status (Metrics server)", err)
		}
		spinner.SetNewAction("Runtime status (Metrics server)...OK")
		spinner.Stop(true)
	}

	spinner = util.StartNewSpinner("Checking runtime status (Cellery)...")
	wrCellerySysterm, err := waitingTimeCellerySystem()
	if err != nil {
		spinner.Stop(false)
		util.ExitWithErrorMessage("Error getting waiting time for cellery system", err)
	}
	err = kubernetes.WaitForDeployments("cellery-system", wrCellerySysterm)
	if err != nil {
		spinner.Stop(false)
		util.ExitWithErrorMessage("Error while checking runtime status (Cellery)", err)
	}
	spinner.SetNewAction("Runtime status (Cellery)...OK")
	spinner.Stop(true)
	return nil
}

func waitingTimeCluster() (time.Duration, error) {
	waitingTime := time.Minute * 60
	envVar := os.Getenv("CELLERY_CLUSTER_WAIT_TIME_MINUTES")
	if envVar != "" {
		wt, err := strconv.Atoi(envVar)
		if err != nil {
			return waitingTime, err
		}
		waitingTime = time.Duration(time.Minute * time.Duration(wt))
	}
	return waitingTime, nil
}

func waitingTimeCellerySystem() (time.Duration, error) {
	waitingTime := time.Minute * 30
	envVar := os.Getenv("CELLERY_SYSTEM_WAIT_TIME_MINUTES")
	if envVar != "" {
		wt, err := strconv.Atoi(envVar)
		if err != nil {
			return waitingTime, err
		}
		waitingTime = time.Duration(time.Minute * time.Duration(wt))
	}
	return waitingTime, nil
}

func (runtime *CelleryRuntime) Validate() error {
	errorMessage := ""
	cmd := exec.Command(
		constants.KubeCtl,
		"version",
		"-o",
		"json",
	)
	jsonOutput := kubernetes.K8sVersion{}
	out, err := osexec.GetCommandOutputFromTextFile(cmd)
	if err != nil {
		return fmt.Errorf("failed to get kubectl versions, %v", err)
	}
	err = json.Unmarshal(out, &jsonOutput)
	if err != nil {
		return fmt.Errorf("failed to unmarshall kubectl versions, %v", err)
	}
	kubectlServerVersion := jsonOutput.ServerVersion.GitVersion
	actualKubectlServerVersion, err := version.NewVersion(kubectlServerVersion)
	expectedKubernetesServerVersionMax, err := version.NewVersion(kubernetesVersionMax)

	if actualKubectlServerVersion.GreaterThanOrEqual(expectedKubernetesServerVersionMax) {
		errorMessage = errorMessage + "\n" + fmt.Sprintf("kubectl server version (%s) is newer than the maximum version expected (%s)", actualKubectlServerVersion, kubernetesVersionMaxExpected)
	}
	expectedKubernetesServerVersionMin, err := version.NewVersion(kubernetesVersionMin)
	if actualKubectlServerVersion.LessThan(expectedKubernetesServerVersionMin) {
		errorMessage = errorMessage + "\n" + fmt.Sprintf("kubectl server version (%s) is older than the minimum version required (%s)", actualKubectlServerVersion, expectedKubernetesServerVersionMin)
	}
	kubectlClientVersion := jsonOutput.ClientVersion.GitVersion
	actualKubectlClientVersion, err := version.NewVersion(kubectlClientVersion)
	expectedKubernetesClientVersionMax, err := version.NewVersion(kubernetesVersionMax)

	if actualKubectlClientVersion.GreaterThanOrEqual(expectedKubernetesClientVersionMax) {
		errorMessage = errorMessage + "\n" + fmt.Sprintf("kubectl client version (%s) is newer than the maximum version expected (%s)", actualKubectlClientVersion, kubernetesVersionMaxExpected)
	}
	expectedKubernetesClientVersionMin, err := version.NewVersion(kubernetesVersionMin)
	if actualKubectlClientVersion.LessThan(expectedKubernetesClientVersionMin) {
		errorMessage = errorMessage + "\n" + fmt.Sprintf("kubectl client version (%s) is older than the minimum version required (%s)", actualKubectlClientVersion, expectedKubernetesClientVersionMin)
	}
	if errorMessage != "" {
		return fmt.Errorf(errorMessage)
	}
	return nil
}

func (runtime *CelleryRuntime) UnmarshalHelmValues(chartName string) error {
	defaultValues, err := util.GetHelmChartDefaultValues(chartName)
	if err != nil {
		log.Printf("Unmarshal error: %v", err)
		return err
	}
	if chartName == "cellery-runtime" {
		err = yaml.Unmarshal([]byte(defaultValues), &runtime.celleryRuntimeVals)
		if err != nil {
			log.Printf("Unmarshal error: %v", err)
			return err
		}
	} else if chartName == "ingress-controller" {
		err = yaml.Unmarshal([]byte(defaultValues), &runtime.ingressControllerVals)
		if err != nil {
			log.Printf("Unmarshal error: %v", err)
			return err
		}
	}
	return nil
}

func (runtime *CelleryRuntime) MarshalHelmValues(chartName string) error {
	if chartName == "cellery-runtime" {
		templateYamls, err := yaml.Marshal(&runtime.celleryRuntimeVals)
		if err != nil {
			log.Printf("Marshal error: %v", err)
			return err
		}
		runtime.celleryRuntimeYaml = string(templateYamls)
	} else if chartName == "ingress-controller" {
		templateYamls, err := yaml.Marshal(&runtime.ingressControllerVals)
		if err != nil {
			log.Printf("error: %v", err)
			return err
		}
		runtime.ingressControllerYamls = string(templateYamls)
	}
	return nil
}

func (runtime *CelleryRuntime) CreateNameSpace(namespace string) error {
	var cmd *exec.Cmd
	cmd = exec.Command(
		constants.KubeCtl,
		"create",
		"ns",
		namespace)

	cmd.Stderr = os.Stderr
	return cmd.Run()
}
