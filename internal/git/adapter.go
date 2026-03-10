// Package git provides Git runtime integration helpers.
package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// CapabilityRequest defines inputs for Git capability checks.
type CapabilityRequest struct {
	Mode       string
	WorkingDir string
}

var lookPath = exec.LookPath

var runCommand = func(workingDir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...) //#nosec G204 -- args are controlled by internal call sites
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

// CheckCapabilities validates Git runtime requirements for enabled Git modes.
func CheckCapabilities(request CapabilityRequest) error {
	mode := strings.TrimSpace(request.Mode)
	if mode == "" || mode == "off" {
		return nil
	}

	if _, err := lookPath("git"); err != nil {
		return fmt.Errorf("git mode %s requires git executable", mode)
	}

	output, err := runCommand(request.WorkingDir, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		return fmt.Errorf("git mode %s requires a git repository", mode)
	}
	if strings.TrimSpace(output) != "true" {
		return fmt.Errorf("git mode %s requires a git repository", mode)
	}

	return nil
}
