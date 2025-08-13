package linux

import (
	"os"
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
