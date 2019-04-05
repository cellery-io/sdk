package kubectl

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
)

func WaitForDeployment(condition string, timeoutSeconds int, resourceName string, namespace string) error {
	return WaitForCondition(condition, timeoutSeconds, fmt.Sprintf("deployment/%s", resourceName), namespace)
}

func WaitForCondition(condition string, timeoutSeconds int, resourceName string, namespace string) error {
	cmd := exec.Command(
		constants.KUBECTL,
		"wait",
		fmt.Sprintf("--for=condition=%s", condition),
		fmt.Sprintf("--timeout=%ds", timeoutSeconds),
		resourceName,
		"-n", namespace,
	)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func WaitForCluster(timeout time.Duration) error {
	exitCode := 0
	for start := time.Now(); time.Since(start) < timeout; {
		cmd := exec.Command(constants.KUBECTL, "get", "nodes", "--request-timeout=10s")
		err := cmd.Run()
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				ws := exitError.Sys().(syscall.WaitStatus)
				exitCode = ws.ExitStatus()
				if exitCode == 1 {
					continue
				} else {
					return fmt.Errorf("kubectl exit with unknown error code %d", exitCode)
				}
			} else {
				return err
			}
		} else {
			return nil
		}
	}
	return fmt.Errorf("cluster ready check timeout")
}

func WaitForDeployments(namespace string) error {
	names, err := GetDeploymentNames(namespace)
	if err != nil {
		return err
	}
	for _, v := range names {
		if len(v) == 0 {
			continue
		}
		err := WaitForDeployment("available", -1, v, namespace)
		if err != nil {
			return err
		}
	}
	return nil
}
