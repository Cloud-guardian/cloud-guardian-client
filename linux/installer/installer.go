package linux_installer

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
)

const (
	targetPath         = "/usr/bin/patchmaster-client"
	serviceName        = "patchmaster-client.service"
	serviceFilePath    = "/etc/systemd/system/" + serviceName
	serviceDescription = "Patchmaster Client Service"
)

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

	// Reload systemd and enable/start service
	execCommand("systemctl", "daemon-reexec")
	execCommand("systemctl", "daemon-reload")
	execCommand("systemctl", "enable", serviceName)
	execCommand("systemctl", "start", serviceName)

	fmt.Println("SystemD service installed and started.")
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

	// Copy binary to /usr/bin
	if err := copyFile(selfPath, targetPath, 0755); err != nil {
		log.Fatalf("Error copying binary: %v\n", err)
	}

	// Create a systemd service file
	if err := createSystemdService(); err != nil {
		log.Fatalf("Error creating systemd service: %v\n", err)
	}
	return nil
}
