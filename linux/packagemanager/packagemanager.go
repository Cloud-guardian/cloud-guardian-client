package linux_packagemanager

import (
	linux_debian_apt "cloud-guardian/linux_debian/apt"
	linux_redhat_dnf "cloud-guardian/linux_redhat/dnf"
	"fmt"
	"os"
)

type UpdateType int

const (
	AllUpdates UpdateType = iota
	SecurityUpdates
)

func (ut UpdateType) String() string {
	switch ut {
	case AllUpdates:
		return "all"
	case SecurityUpdates:
		return "security"
	default:
		return "unknown"
	}
}

// Package interface to standardize package information
type Package struct {
	Name    string
	Version string
	Repo    string
}

// PackageManager interface to abstract package manager operations
type PackageManager interface {
	GetInstalledPackages() ([]Package, error)
	CheckUpdates(updatetype UpdateType) ([]Package, []Package, error)
}

func DetectPackageManager() (PackageManager, error) {
	// Check if dnf is available
	if _, err := os.Stat("/usr/bin/dnf"); err == nil {
		return &Dnf{}, nil
	}

	// Check if apt is available
	if _, err := os.Stat("/usr/bin/apt"); err == nil {
		return &Apt{}, nil
	}

	return nil, fmt.Errorf("no supported package manager found")
}

// DNF Manager implementation
type Dnf struct{}

func (dnf *Dnf) GetInstalledPackages() ([]Package, error) {
	packages, err := linux_redhat_dnf.GetInstalledPackages()
	if err != nil {
		return nil, err
	}

	result := make([]Package, len(packages))
	for i, pkg := range packages {
		result[i] = Package{
			Name:    pkg.Name,
			Version: pkg.Version,
			Repo:    pkg.Repo,
		}
	}
	return result, nil
}

func (dnf *Dnf) CheckUpdates(updateType UpdateType) ([]Package, []Package, error) {
	updates, obsolete, err := linux_redhat_dnf.CheckUpdates(linux_redhat_dnf.UpdateType(updateType))
	if err != nil {
		return nil, nil, err
	}

	updatesResult := make([]Package, len(updates))
	for i, pkg := range updates {
		updatesResult[i] = Package{
			Name:    pkg.Name,
			Version: pkg.Version,
			Repo:    pkg.Repo,
		}
	}

	obsoleteResult := make([]Package, len(obsolete))
	for i, pkg := range obsolete {
		obsoleteResult[i] = Package{
			Name:    pkg.Name,
			Version: pkg.Version,
			Repo:    pkg.Repo,
		}
	}

	return updatesResult, obsoleteResult, nil
}

// APT Manager implementation
type Apt struct{}

func (apt *Apt) GetInstalledPackages() ([]Package, error) {
	packages, err := linux_debian_apt.GetInstalledPackages()
	if err != nil {
		return nil, err
	}

	result := make([]Package, len(packages))
	for i, pkg := range packages {
		result[i] = Package{
			Name:    pkg.Name,
			Version: pkg.Version,
			Repo:    pkg.Repo,
		}
	}
	return result, nil
}

func (apt *Apt) CheckUpdates(updateType UpdateType) ([]Package, []Package, error) {
	linux_debian_apt.AptUpdate() // Ensure apt is updated before checking for updates

	updates, obsolete, err := linux_debian_apt.CheckUpdates(linux_debian_apt.UpdateType(updateType))
	if err != nil {
		return nil, nil, err
	}

	updatesResult := make([]Package, len(updates))
	for i, pkg := range updates {
		updatesResult[i] = Package{
			Name:    pkg.Name,
			Version: pkg.Version,
			Repo:    pkg.Repo,
		}
	}

	obsoleteResult := make([]Package, len(obsolete))
	for i, pkg := range obsolete {
		obsoleteResult[i] = Package{
			Name:    pkg.Name,
			Version: pkg.Version,
			Repo:    pkg.Repo,
		}
	}

	return updatesResult, obsoleteResult, nil
}
