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
