package externalprocess

import "os/exec"

func Execute(command []string) ([]byte, error) {
	cmd := exec.Command(command[0], command[1:]...) //nolint
	return cmd.Output()
}
