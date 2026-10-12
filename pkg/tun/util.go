package tun

import (
	"fmt"
	"os/exec"
	"strings"
)

// runCmd executes a command and returns an error if it fails.
func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %w (output: %s)",
			name, strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return nil
}