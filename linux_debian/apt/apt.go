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

func UpdatePackages(packages []string) (string, error) {
	command := exec.Command("apt", "--only-upgrade", "--assume-yes", "--quiet", "install", strings.Join(packages, " "))
	var out strings.Builder
	command.Stdout = &out
	command.Stderr = &out // Capture stderr as well
	err := command.Run()
	if err != nil {
		return "", fmt.Errorf("command failed: %s", out.String())
	}
	return out.String(), nil
}

func InstallPackages(packages []string) (string, error) {
	command := exec.Command("apt", "install", "--assume-yes", "--quiet", strings.Join(packages, " "))
	var out strings.Builder
	command.Stdout = &out
	command.Stderr = &out // Capture stderr as well
	err := command.Run()
	if err != nil {
		return "", fmt.Errorf("command failed: %s", out.String())
	}
	return out.String(), nil
}

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
