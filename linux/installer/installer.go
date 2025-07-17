package linux_installer

import (
	cgconfig "cloud-guardian/cloudguardian_config"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"encoding/json"
)

const (
	targetPath         = "/usr/bin/cloud-guardian"
	serviceName        = "cloud-guardian.service"
	serviceFilePath    = "/etc/systemd/system/" + serviceName
	serviceDescription = "Cloud Gardian Client Service"
	configFilePath	   = "/etc/cloud-guardian.json"
)

var Config *cgconfig.CloudGardianConfig

func HasRootPrivileges() bool {
	// Check if the current user has root privileges
	return os.Geteuid() == 0
}

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
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to run command: %s %v\nError: %v\n", name, args, err)
	}
}

func createConfigFile(filename string) error {

	defaultApiUrl := cgconfig.DefaultConfig().ApiUrl

	configFileContent := map[string]any{
		"api_key": Config.ApiKey,
	}

	if Config.ApiUrl != defaultApiUrl {
		configFileContent["api_url"] = Config.ApiUrl
	}

	if Config.Debug {
		configFileContent["debug"] = true
	}

	// Marshal the config map into JSON with indentation
	jsonConfig, err := json.MarshalIndent(configFileContent, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(filename, []byte(jsonConfig), 0644); err != nil {
		log.Fatalf("Error writing config file: %v\n", err)
	}
	return nil
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
	fmt.Printf("Installed systemd service at %s\n", serviceFilePath)
	return nil
}

func EnableAndStartService() error {
	println("Enabling and starting the service...")
	if !HasRootPrivileges() {
		return os.ErrPermission // User does not have root privileges
	}

	// // Create the service file if it doesn't exist
	// if _, err := os.Stat(serviceFilePath); os.IsNotExist(err) {
	// 	if err := createSystemdService(); err != nil {
	// 		log.Fatalf("Error creating systemd service: %v\n", err)
	// 	}
	// }

	// Reload systemd to ensure it recognizes the new service file
	execCommand("systemctl", "daemon-reexec")
	execCommand("systemctl", "daemon-reload")

	// Enable the service
	execCommand("systemctl", "enable", serviceName)

	// Start the service
	execCommand("systemctl", "start", serviceName)

	fmt.Println("Service enabled and started successfully.")
	return nil
}

func IsServiceRunning() bool {
	cmd := exec.Command("systemctl", "is-active", "--quiet", serviceName)
	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 3 {
			// Service is inactive
			return false
		}
		log.Fatalf("Failed to check service status: %v\n", err)
	}
	return true // Service is active
}

func IsServiceEnabled() bool {
	cmd := exec.Command("systemctl", "is-enabled", "--quiet", serviceName)
	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			// Service is not enabled
			return false
		}
		log.Fatalf("Failed to check service enabled status: %v\n", err)
	}
	return true // Service is enabled
}

func DisableAndStopService() error {
	println("Disabling and stopping the service...")
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

	fmt.Println("Service disabled and stopped successfully.")
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

	DisableAndStopService()

	// Copy binary to /usr/bin
	if err := copyFile(selfPath, targetPath, 0755); err != nil {
		log.Fatalf("Error copying binary: %v\n", err)
	}

	// Create a systemd service file
	if err := createSystemdService(); err != nil {
		log.Fatalf("Error creating systemd service: %v\n", err)
	}

	EnableAndStartService()

	// Create the configuration file
	if err := createConfigFile(configFilePath); err != nil {
		log.Fatalf("Error creating config file: %v\n", err)
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

	fmt.Println("Uninstalled successfully.")
	return nil
}
