package linux_uptime

import (
	"os"
	"strconv"
	"strings"
)

func GetUptime() (int64, error) {
	// Read the /proc/uptime file to get system uptime
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0, err
	}

	// Split the data into two parts: uptime and idle time
	parts := string(data)
	uptime := parts[:len(parts)-1] // Remove the trailing newline character
	uptimeParts := strings.Split(uptime, " ")
	// uptimeStr := strings.TrimSpace(string(uptime))
	uptimeSeconds, err := strconv.ParseFloat(uptimeParts[0], 64)
	if err != nil {
		return 0, err
	}
	uptimeSecondsInt := int64(uptimeSeconds)
	return uptimeSecondsInt, nil
}
