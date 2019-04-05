package kubectl

import (
	"os/exec"
	"strings"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
)

func GetDeploymentNames(namespace string) ([]string, error) {
	cmd := exec.Command(
		constants.KUBECTL,
		"get",
		"deployments",
		"-o",
		"jsonpath={.items[*].metadata.name}",
		"-n", namespace,
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return strings.Split(string(out), " "), nil
}
