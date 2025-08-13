package linux_debian_apt

import (
	"fmt"
	"os/exec"
	"strings"
)

type AptPackage struct {
	Name    string
	Version string
	Repo    string
}

type UpdateType int

const (
	AllUpdates UpdateType = iota
	SecurityUpdates
)

// runCommand executes a given command and captures both stdout and stderr.
// It returns the standard output, standard error, and any error that occurred during execution.
//
// Parameters:
//   - command: The exec.Cmd to execute
//
// Returns:
//   - string: Standard output from the command
//   - string: Standard error output from the command
//   - error: Any error that occurred during execution
func runCommand(command *exec.Cmd) (string, string, error) {
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

// UpdateAllPackages upgrades all packages on the system using APT.
// It runs the equivalent of 'apt upgrade --assume-yes --quiet' command.
//
// Returns:
//   - string: Standard output from the APT upgrade command
//   - string: Standard error output from the APT upgrade command
//   - error: Any error that occurred during the upgrade process
func UpdateAllPackages() (string, string, error) {
	command := exec.Command("apt", "upgrade", "--assume-yes", "--quiet")
	return runCommand(command)
}

// UpdatePackages updates the specified packages using the APT package manager.
// It takes a slice of package names and attempts to update them to their latest versions.
//
// Parameters:
//   - packages: A slice of strings containing the names of packages to update
//
// Returns:
//   - string: Standard output from the APT update command
//   - string: Standard error output from the APT update command
//   - error: Any error that occurred during the update process
func UpdatePackages(packages []string) (string, string, error) {
	command := exec.Command("apt", "--only-upgrade", "--assume-yes", "--quiet", "install")
	command.Args = append(command.Args, packages...)
	return runCommand(command)
}

// InstallPackages installs the specified packages using the APT package manager.
// It takes a slice of package names and attempts to install them.
//
// Parameters:
//   - packages: A slice of strings containing the names of packages to install
//
// Returns:
//   - string: Standard output from the APT install command
//   - string: Standard error output from the APT install command
//   - error: Any error that occurred during the installation process
func InstallPackages(packages []string) (string, string, error) {
	command := exec.Command("apt", "install", "--assume-yes", "--quiet", strings.Join(packages, " "))
	return runCommand(command)
}

// GetInstalledPackages retrieves a list of all installed packages on the system.
// It executes 'apt list --installed' and parses the output.
//
// Returns:
//   - []AptPackage: A slice of AptPackage structs containing package information
//   - error: Any error that occurred during the retrieval process
func GetInstalledPackages() ([]AptPackage, error) {
	command := exec.Command("apt", "list", "--installed")
	var out strings.Builder
	command.Stdout = &out
	err := command.Run()
	if err != nil {
		return nil, fmt.Errorf("command failed: %s", out.String())
	}
	return parseInstalledPackages(out.String()), nil
}

// parseInstalledPackages parses the output from 'apt list --installed' command.
// It extracts package information from each line and returns a slice of AptPackage structs.
//
// Parameters:
//   - output: The raw output string from the APT list installed command
//
// Returns:
//   - []AptPackage: A slice of parsed AptPackage structs
func parseInstalledPackages(output string) []AptPackage {
	lines := strings.Split(output, "\n")
	packages := []AptPackage{}
	for _, line := range lines {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "Listing...") {
			continue // Skip empty lines and listing header
		}
		// Split the line by whitespace and take the first part as the package name
		parts := strings.Split(line, "/")
		if len(parts) < 2 {
			continue // Skip lines that do not have enough parts
		}
		name := parts[0]
		repoVersion := strings.Split(parts[1], " ")
		if len(repoVersion) < 2 {
			continue // Skip if repo/version info is incomplete
		}
		repo := repoVersion[0]
		version := repoVersion[1]

		pkg := AptPackage{
			Name:    name,
			Version: version,
			Repo:    repo,
		}
		packages = append(packages, pkg)
	}
	return packages
}

// AptUpdate updates the package lists using APT.
// It runs the equivalent of 'apt update' command.
//
// Returns:
//   - error: Any error that occurred during the update process
func AptUpdate() error {
	command := exec.Command("apt", "update")
	var out strings.Builder
	command.Stdout = &out
	err := command.Run()
	if err != nil {
		return fmt.Errorf("command failed: %s", out.String())
	}
	return nil
}

// CheckUpdates checks for available package updates using APT.
// It can check for all updates or security-only updates based on the updateType parameter.
//
// Parameters:
//   - updateType: UpdateType enum specifying whether to check all updates or security updates only
//
// Returns:
//   - []AptPackage: A slice of packages that have updates available
//   - []AptPackage: A slice of obsolete packages (empty for APT)
//   - error: Any error that occurred during the check process
func CheckUpdates(updateType UpdateType) ([]AptPackage, []AptPackage, error) {
	var command *exec.Cmd
	command = exec.Command("apt", "list", "--upgradable")
	var out strings.Builder
	command.Stdout = &out

	err := command.Run()
	if err != nil {
		return nil, nil, fmt.Errorf("command failed: %s", out.String())
	}
	updates, obsolete := parseUpdates(out.String(), updateType)
	return updates, obsolete, nil
}

// parseUpdates parses the output from 'apt list --upgradable' command.
// It extracts package information and filters by update type if specified.
//
// Parameters:
//   - output: The raw output string from the APT list upgradable command
//   - updateType: The type of updates to filter for (all or security)
//
// Returns:
//   - []AptPackage: A slice of packages with available updates
//   - []AptPackage: A slice of obsolete packages (empty for APT)
func parseUpdates(output string, updateType UpdateType) ([]AptPackage, []AptPackage) {
	lines := strings.Split(output, "\n")
	updates := []AptPackage{}
	obsolete := []AptPackage{}
	for _, line := range lines {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "Listing...") || strings.HasPrefix(line, "WARNING:") {
			continue // Skip empty lines and listing header
		}
		if updateType == SecurityUpdates && !strings.Contains(line, "-security") {
			continue // Skip non-security updates if security flag is set
		}
		// Split the line by whitespace and take the first part as the package name
		parts := strings.Split(line, "/")
		if len(parts) < 2 {
			continue // Skip lines that do not have enough parts
		}
		name := parts[0]
		repoVersion := strings.Split(parts[1], " ")
		if len(repoVersion) < 2 {
			continue // Skip if repo/version info is incomplete
		}
		repo := repoVersion[0]
		version := repoVersion[1]

		pkg := AptPackage{
			Name:    name,
			Version: version,
			Repo:    repo,
		}

		if strings.Contains(repo, "obsolete") {
			obsolete = append(obsolete, pkg)
		} else {
			updates = append(updates, pkg)
		}
	}
	return updates, obsolete
}
