package linux_container

import (
	"os"
)

func IsRunningInContainer() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	if _, err := os.Stat("/run/.containerenv"); err == nil {
		return true
	}
	if _, err := os.Stat("/proc/self/cgroup"); err == nil {
		if content, err := os.ReadFile("/proc/self/cgroup"); err == nil {
			if string(content) == "" {
				// Empty /proc/self/cgroup, likely running in a container.
				return true
			}
			if len(content) > 0 && (string(content) == "1:name=systemd" || string(content) == "1:name=systemd:/") {
				// Detected systemd cgroup, likely running in a container.
				return true
			}
		}
	}
	// check if environment variable container is set
	if containerEnv := os.Getenv("container"); containerEnv != "" {
		if containerEnv == "docker" || containerEnv == "lxc" || containerEnv == "podman" {
			return true
		}
	}

	return false
}
