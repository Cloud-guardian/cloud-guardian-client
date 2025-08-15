package linux

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// HasRootPrivileges checks if the current process is running with root privileges.
// It returns true if the effective user ID is 0 (root), false otherwise.
//
// Returns:
//   - bool: true if running as root, false otherwise
func HasRootPrivileges() bool {
	// Check if the current user has root privileges
	return os.Geteuid() == 0
}

// RunCommand executes a given command and captures both stdout and stderr.
// It returns the standard output, standard error, and any error that occurred during execution.
//
// Parameters:
//   - command: The exec.Cmd to execute
//
// Returns:
//   - string: Standard output from the command
//   - string: Standard error output from the command
//   - error: Any error that occurred during execution
func RunCommand(command *exec.Cmd) (string, string, error) {
	var stdout strings.Builder
	var stderr strings.Builder
	command.Stdout = &stdout
	command.Stderr = &stderr // Capture stderr as well
	err := command.Run()
	if err != nil {
		return stdout.String(), stderr.String(), fmt.Errorf("command failed: %s", stderr.String())
	}
	return stdout.String(), stderr.String(), nil
}
