package linux_installer

import (
	cgconfig "cloud-guardian/cloudguardian_config"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

const (
	targetPath         = "/usr/bin/cloud-guardian"
	serviceName        = "cloud-guardian.service"
	serviceFilePath    = "/etc/systemd/system/" + serviceName
	serviceDescription = "Cloud Gardian Client Service"
	configFilePath     = "/etc/cloud-guardian.json"
)

var Config *cgconfig.CloudGuardianConfig

// HasRootPrivileges checks if the current process is running with root privileges.
// It returns true if the effective user ID is 0 (root), false otherwise.
//
// Returns:
//   - bool: true if running as root, false otherwise
func HasRootPrivileges() bool {
	// Check if the current user has root privileges
	return os.Geteuid() == 0
}

// copyFile copies a file from source to destination with the specified file mode.
// It handles the opening, copying, and setting permissions of the destination file.
//
// Parameters:
//   - src: The source file path
//   - dst: The destination file path
//   - filemode: The file mode to set on the destination file
//
// Returns:
//   - error: Any error that occurred during the copy operation
func copyFile(src, dst string, filemode os.FileMode) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, filemode)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func execCommand(name string, args ...string) {
	cmd := exec.Command(name, args...)
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to run command: %s %v\nError: %v\n", name, args, err)
	}
}

func createSystemdService() error {
	serviceFileContent := `[Unit]
Description=` + serviceDescription + `
After=network.target

[Service]
ExecStart=` + targetPath + `
Restart=always

[Install]
WantedBy=multi-user.target
`
	if err := os.WriteFile(serviceFilePath, []byte(serviceFileContent), 0644); err != nil {
		log.Fatalf("Error writing service file: %v\n", err)
	}
	log.Printf("Installed systemd service at %s\n", serviceFilePath)
	return nil
}

func EnableAndStartService() error {
	if !HasRootPrivileges() {
		return os.ErrPermission // User does not have root privileges
	}

	// Reload systemd to ensure it recognizes the new service file
	execCommand("systemctl", "daemon-reexec")
	execCommand("systemctl", "daemon-reload")

	// Enable the service
	execCommand("systemctl", "enable", serviceName)

	// Start the service
	execCommand("systemctl", "start", serviceName)

	return nil
}

func IsServiceRunning() bool {
	command := exec.Command("systemctl", "is-active", serviceName)
	var out strings.Builder
	command.Stdout = &out
	err := command.Run()
	if err != nil {
		// Check if service is inactive by examining output
		if string(out.String()) == "inactive\n" {
			return false
		}
		if string(out.String()) == "failed\n" {
			return false
		}
		log.Fatalf("Failed to check service status: %v\n", err)
	}
	return true // Service is active
}

func IsServiceEnabled() bool {
	command := exec.Command("systemctl", "is-enabled", serviceName)
	var stdout strings.Builder
	var stderr strings.Builder
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	if err != nil {
		if strings.Contains(stdout.String(), "disabled") || strings.Contains(stdout.String(), "not-found") {
			return false // Service is not enabled or does not exist
		}
		if strings.Contains(stderr.String(), "Failed to get unit file state for") && strings.Contains(stderr.String(), "No such file or directory") {
			return false // Service does not exist
		}
		log.Fatalf("Failed to check service enabled status: %v\n", err)
	}
	return true // Service is enabled
}

func DisableAndStopService() error {
	if !HasRootPrivileges() {
		return os.ErrPermission // User does not have root privileges
	}

	// Stop the service
	if IsServiceRunning() {
		execCommand("systemctl", "stop", serviceName)
	}

	// // Disable the service
	if IsServiceEnabled() {
		execCommand("systemctl", "disable", serviceName)
	}

	return nil
}

func Update() error {
	if !HasRootPrivileges() {
		return os.ErrPermission // User does not have root privileges
	}

	// Check if service is installed
	if _, err := os.Stat(serviceFilePath); os.IsNotExist(err) {
		log.Fatalf("Service file does not exist at %s. Please install the service first.\n", serviceFilePath)
	}

	// Check if config file exists
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		log.Fatalf("Configuration file does not exist at %s. Please install the service first.\n", configFilePath)
	}

	// Check if service is active
	if !IsServiceEnabled() {
		log.Fatalf("Service is not enabled. Please install and enable the service first.\n")
	}

	selfPath, err := os.Executable()
	if err != nil {
		log.Fatalf("Error getting executable path: %v\n", err)
	}

	if selfPath == targetPath {
		log.Fatalf("I can not update myself while running from the target path: %s\n", targetPath)
	}

	if err := DisableAndStopService(); err != nil {
		log.Fatalf("Error disabling and stopping service: %v\n", err)
	}

	// Copy binary to /usr/bin
	if err := copyFile(selfPath, targetPath, 0755); err != nil {
		log.Fatalf("Error copying binary: %v\n", err)
	}

	if err := EnableAndStartService(); err != nil {
		log.Fatalf("Error enabling and starting service: %v\n", err)
	}

	log.Println("Client updated successfully.")
	return nil
}

func Install() error {
	if !HasRootPrivileges() {
		return os.ErrPermission // User does not have root privileges
	}

	selfPath, err := os.Executable()
	if err != nil {
		log.Fatalf("Error getting executable path: %v\n", err)
	}

	if err := DisableAndStopService(); err != nil {
		log.Fatalf("Error disabling and stopping service: %v\n", err)
	}

	// Copy binary to /usr/bin
	if err := copyFile(selfPath, targetPath, 0755); err != nil {
		log.Fatalf("Error copying binary: %v\n", err)
	}

	// Create a systemd service file
	if err := createSystemdService(); err != nil {
		log.Fatalf("Error creating systemd service: %v\n", err)
	}

	// Create the configuration file
	if err := Config.Save(configFilePath); err != nil {
		log.Fatalf("Error creating config file: %v\n", err)
	}

	if err := EnableAndStartService(); err != nil {
		log.Fatalf("Error enabling and starting service: %v\n", err)
	}

	return nil
}

func Uninstall() error {
	if !HasRootPrivileges() {
		return os.ErrPermission // User does not have root privileges
	}

	// Stop and disable the service
	if err := DisableAndStopService(); err != nil {
		log.Fatalf("Error disabling and stopping service: %v\n", err)
	}

	// Remove the service file
	if _, err := os.Stat(serviceFilePath); !os.IsNotExist(err) {
		if err := os.Remove(serviceFilePath); err != nil {
			log.Fatalf("Error removing service file: %v\n", err)
		}
	}

	// Remove the binary
	if _, err := os.Stat(targetPath); !os.IsNotExist(err) {
		if err := os.Remove(targetPath); err != nil {
			log.Fatalf("Error removing binary: %v\n", err)
		}
	}

	// Remove the configuration file
	if _, err := os.Stat(configFilePath); !os.IsNotExist(err) {
		if err := os.Remove(configFilePath); err != nil {
			log.Fatalf("Error removing config file: %v\n", err)
		}
	}

	log.Println("Uninstalled successfully.")
	return nil
}
