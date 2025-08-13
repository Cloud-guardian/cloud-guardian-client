// Manage dnf packages
package linux_redhat_dnf

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

type DnfPackage struct {
	Name    string
	Version string
	Repo    string
}

type DnfUpdateSummary struct {
	SecuritImportant int
	SecurityModerate int
	Bugfix           int
	Enhancement      int
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

// UpdateAllPackages updates all packages on the system using DNF.
// It runs the equivalent of 'dnf update --assumeyes --quiet' command.
//
// Returns:
//   - string: Standard output from the DNF update command
//   - string: Standard error output from the DNF update command
//   - error: Any error that occurred during the update process
func UpdateAllPackages() (string, string, error) {
	command := exec.Command("dnf", "update", "--assumeyes", "--quiet")
	return runCommand(command)
}

// UpdatePackages updates the specified packages using the DNF package manager.
// It takes a slice of package names and attempts to update them to their latest versions.
//
// Parameters:
//   - packages: A slice of strings containing the names of packages to update
//
// Returns:
//   - string: Standard output from the DNF update command
//   - string: Standard error output from the DNF update command
//   - error: Any error that occurred during the update process
//
// Example:
//
//	stdout, stderr, err := dnf.UpdatePackages([]string{"nginx", "curl"})
//	if err != nil {
//	    log.Printf("Update failed: %v, stderr: %s", err, stderr)
//	}
func UpdatePackages(packages []string) (string, string, error) {
	command := exec.Command("dnf", "update", "--assumeyes", "--quiet")
	command.Args = append(command.Args, packages...)
	return runCommand(command)
}

// InstallPackages installs the specified packages using the DNF package manager.
// It takes a slice of package names and attempts to install them.
//
// Parameters:
//   - packages: A slice of strings containing the names of packages to install
//
// Returns:
//   - string: Standard output from the DNF install command
//   - string: Standard error output from the DNF install command
//   - error: Any error that occurred during the installation process
func InstallPackages(packages []string) (string, string, error) {
	command := exec.Command("dnf", "install", "--assumeyes", "--quiet")
	command.Args = append(command.Args, packages...)
	return runCommand(command)
}

// GetInstalledPackages retrieves a list of all installed packages on the system.
// It executes 'dnf list installed --quiet' and parses the output.
//
// Returns:
//   - []DnfPackage: A slice of DnfPackage structs containing package information
//   - error: Any error that occurred during the retrieval process
func GetInstalledPackages() ([]DnfPackage, error) {
	command := exec.Command("dnf", "list", "installed", "--quiet")
	var out strings.Builder
	command.Stdout = &out
	err := command.Run()
	if err != nil {
		return nil, fmt.Errorf("command failed: %s", out.String())
	}

	return parseInstalledPackages(out.String()), nil
}

// parseInstalledPackages parses the output from 'dnf list installed' command.
// It extracts package information from each line and returns a slice of DnfPackage structs.
//
// Parameters:
//   - output: The raw output string from the DNF list installed command
//
// Returns:
//   - []DnfPackage: A slice of parsed DnfPackage structs
func parseInstalledPackages(output string) []DnfPackage {
	lines := strings.Split(output, "\n")
	packages := []DnfPackage{}
	for _, line := range lines {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "Installed Packages") {
			continue // Skip empty lines and header
		}
		// Split the line by whitespace and take the first three parts as package name, version, and repo
		parts := regexp.MustCompile(`\s+`).Split(line, -1)
		if len(parts) >= 3 {
			pkg := DnfPackage{
				Name:    parts[0],
				Version: parts[1],
				Repo:    parts[2],
			}
			packages = append(packages, pkg)
		}
	}
	return packages
}

// parseUpdateSummary parses the output from 'dnf updateinfo --summary' command.
// It extracts update information including security, bugfix, and enhancement counts.
//
// Parameters:
//   - output: The raw output string from the DNF updateinfo summary command
//
// Returns:
//   - DnfUpdateSummary: A struct containing categorized update counts
func parseUpdateSummary(output string) DnfUpdateSummary {

	// Parse the output of `dnf updateinfo --summary` to extract the summary information
	// Updates Information Summary: available
	//     8 Security notice(s)
	//         3 Important Security notice(s)
	//         5 Moderate Security notice(s)
	//     3 Bugfix notice(s)
	//     1 Enhancement notice(s)

	lines := strings.Split(output, "\n")
	summary := DnfUpdateSummary{}
	for _, line := range lines {
		parts := strings.Fields(line)
		if strings.Contains(line, "Important Security notice(s)") && len(parts) >= 3 {
			fmt.Sscanf(parts[0], "%d", &summary.SecuritImportant)
		} else if strings.Contains(line, "Moderate Security notice(s)") && len(parts) >= 3 {
			fmt.Sscanf(parts[0], "%d", &summary.SecurityModerate)
		} else if strings.Contains(line, "Bugfix notice(s)") && len(parts) >= 3 {
			fmt.Sscanf(parts[0], "%d", &summary.Bugfix)
		} else if strings.Contains(line, "Enhancement notice(s)") && len(parts) >= 3 {
			fmt.Sscanf(parts[0], "%d", &summary.Enhancement)
		}
	}
	return summary
}

// CheckUpdates checks for available package updates using DNF.
// It can check for all updates or security-only updates based on the updateType parameter.
//
// Parameters:
//   - updateType: UpdateType enum specifying whether to check all updates or security updates only
//
// Returns:
//   - []DnfPackage: A slice of packages that have updates available
//   - []DnfPackage: A slice of obsolete packages
//   - error: Any error that occurred during the check process
func CheckUpdates(updateType UpdateType) ([]DnfPackage, []DnfPackage, error) {
	command := exec.Command("dnf", "check-update", "--quiet")
	if updateType == SecurityUpdates {
		command.Args = append(command.Args, "--security")
	}
	var out strings.Builder
	command.Stdout = &out
	err := command.Run()
	if err != nil {
		// Exit code 100 indicates updates are available, which is not an error in this context
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 100 {
			// Treat exit code 100 as a success
		} else {
			return nil, nil, fmt.Errorf("command failed: %s", out.String())
		}
	}

	updates, obsolete := parseUpdates(out.String())
	return updates, obsolete, nil
}

// CheckUpdateSummary retrieves a summary of available updates categorized by type.
// It executes 'dnf updateinfo --summary --quiet' and parses the results.
//
// Returns:
//   - DnfUpdateSummary: A struct containing counts of different update types
//   - error: Any error that occurred during the summary retrieval process
func CheckUpdateSummary() (DnfUpdateSummary, error) {
	command := exec.Command("dnf", "updateinfo", "--summary", "--quiet")
	var out strings.Builder
	command.Stdout = &out
	err := command.Run()
	if err != nil {
		return DnfUpdateSummary{}, fmt.Errorf("command failed: %s", out.String())
	}
	summary := parseUpdateSummary(out.String())

	return summary, nil
}

// CheckUpdateInfoList retrieves a detailed list of available updates with advisory information.
// It executes 'dnf updateinfo list --quiet' and parses the output.
//
// Returns:
//   - []DnfPackage: A slice of DnfPackage structs containing update information
//   - error: Any error that occurred during the retrieval process
func CheckUpdateInfoList() ([]DnfPackage, error) {
	command := exec.Command("dnf", "updateinfo", "list", "--quiet")
	var out strings.Builder
	command.Stdout = &out
	err := command.Run()
	if err != nil {
		return nil, fmt.Errorf("command failed: %s", out.String())
	}

	lines := strings.Split(out.String(), "\n")
	packages := []DnfPackage{}
	for _, line := range lines {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "Last metadata expiration check") {
			continue // Skip empty lines and metadata expiration messages
		}
		// Split the line by whitespace and take the first three parts as package name, version, and repo
		parts := regexp.MustCompile(`\s+`).Split(line, -1)
		if len(parts) >= 3 {
			pkg := DnfPackage{
				Name:    parts[0],
				Version: parts[1],
				Repo:    parts[2],
			}
			packages = append(packages, pkg)
		}
	}
	return packages, nil
}

// parseUpdates parses the output from 'dnf check-update' command.
// It separates regular updates from obsolete packages and returns both lists.
//
// Parameters:
//   - output: The raw output string from the DNF check-update command
//
// Returns:
//   - []DnfPackage: A slice of packages with available updates
//   - []DnfPackage: A slice of obsolete packages
func parseUpdates(output string) ([]DnfPackage, []DnfPackage) {
	// Parse the output of `dnf check-update` to extract package names
	// Return the package names and obsolete packages
	lines := strings.Split(output, "\n")
	updates := []DnfPackage{}
	obsolete := []DnfPackage{}
	isObsoleteSection := false
	for _, line := range lines {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "Last metadata expiration check") {
			continue // Skip empty lines and metadata expiration messages
		}
		if strings.HasPrefix(line, "Obsoleting Packages") {
			isObsoleteSection = true
			continue // Skip obsolete packages line
		}
		// Split the line by whitespace and take the first part as the package name
		parts := regexp.MustCompile(`\s+`).Split(line, -1)
		if len(parts) >= 3 {
			pkg := DnfPackage{
				Name:    parts[0],
				Version: parts[1],
				Repo:    parts[2],
			}
			if isObsoleteSection {
				obsolete = append(obsolete, pkg)
			} else {
				updates = append(updates, pkg)
			}
		}
	}
	return updates, obsolete
}
