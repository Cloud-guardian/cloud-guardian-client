package linux_loggedinusers

import (
	"os/exec"
	"strings"
	"fmt"
)

type LoggedInUser struct {
	Username string
	Terminal string
	LoginTime string
	Host     string
}

// GetLoggedInUsers retrieves the list of currently logged-in users on a Linux system.
func GetLoggedInUsers() ([]LoggedInUser, error) {
	command := exec.Command("who")
	var out strings.Builder
	command.Stdout = &out
	err := command.Run()
	if err != nil {
		return nil, fmt.Errorf("command failed: %s", out.String())
	}

	return parseLoggedInUsers(out.String()), nil
}

func parseLoggedInUsers(output string) []LoggedInUser {
	lines := strings.Split(output, "\n")
	users := []LoggedInUser{}
	for _, line := range lines {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "NAME") {
			continue // Skip empty lines and header
		}
		// Split the line by whitespace and take the first part as username
		parts := strings.Fields(line)
		host := parts[4]
		// strip the ( and ) from the host if it exists
		if strings.HasPrefix(host, "(") && strings.HasSuffix(host, ")") {
			host = strings.TrimSuffix(strings.TrimPrefix(host, "("), ")")
		}
		if len(parts) > 0 {
			users = append(users, LoggedInUser{
				Username:  parts[0],
				Terminal:  parts[1],
				LoginTime: parts[2] + " " + parts[3], // Combine date and time
				Host:      host, // Host is the last part
			})
		}
	}
	return users
}
