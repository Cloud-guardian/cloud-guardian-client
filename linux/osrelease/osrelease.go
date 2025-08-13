// Package osrelease parse /etc/os-release
// More about the os-release: https://www.linux.org/docs/man5/os-release.html
package linux_osrelease

import (
	"fmt"
	"os"
	"strings"
)

// Path contains the default path to the os-release file
var Path = "/etc/os-release"

var Release OSRelease

type OSRelease struct {
	Name       string
	Version    string
	ID         string
	IDLike     string
	PrettyName string
	VersionID  string
	HomeURL    string
	// DocumentationURL string
	// SupportURL string
	// BugReportURL string
	// PrivacyPolicyURL string
	// VersionCodename string
	// UbuntuCodename string
	// ANSIColor string
	// CPEName string
	// BuildID string
	// Variant string
	// VariantID string
	// Logo string
}

// getLines reads the os-release file and returns it line by line.
// Empty lines and comments (beginning with a "#") are ignored.
//
// Returns:
//   - []string: A slice of non-empty, non-comment lines from the os-release file
//   - error: Any error that occurred while reading the file
func getLines() ([]string, error) {

	output, err := os.ReadFile(Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %s", Path, err)
	}

	lines := make([]string, 0)

	for _, line := range strings.Split(string(output), "\n") {

		switch true {
		case line == "":
			continue
		case []byte(line)[0] == '#':
			continue
		}

		lines = append(lines, line)
	}

	return lines, nil
}

// parseLine parses a single line from the os-release file.
// It splits the line on the first '=' character to extract key-value pairs.
//
// Parameters:
//   - line: A single line from the os-release file
//
// Returns:
//   - string: The key portion of the key-value pair
//   - string: The value portion with quotes removed
//   - error: Any error that occurred during parsing
func parseLine(line string) (string, string, error) {

	subs := strings.SplitN(line, "=", 2)

	if len(subs) != 2 {
		return "", "", fmt.Errorf("invalid length of the substrings: %d", len(subs))
	}

	return subs[0], strings.Trim(subs[1], "\"'"), nil
}

// GetOsReleaseInfo reads and parses the os-release file to populate the global Release variable.
// It combines reading the file and parsing the content into a single operation.
//
// Returns:
//   - error: Any error that occurred during reading or parsing
func GetOsReleaseInfo() error {

	lines, err := getLines()
	if err != nil {
		return fmt.Errorf("failed to get lines of %s: %s", Path, err)
	}
	if err := Parse(lines); err != nil {
		return fmt.Errorf("failed to parse os-release file: %s", err)
	}

	return nil
}

// Parse parses the lines from an os-release file and populates the global Release variable.
// It processes each line to extract key-value pairs and maps them to the OSRelease struct fields.
//
// Parameters:
//   - lines: A slice of strings containing the lines from the os-release file
//
// Returns:
//   - error: Any error that occurred during parsing
func Parse(lines []string) error {

	for i := range lines {

		key, value, err := parseLine(lines[i])
		if err != nil {
			return fmt.Errorf("failed to parse line '%s': %s", lines[i], err)
		}

		switch key {
		case "NAME":
			Release.Name = value
		case "VERSION":
			Release.Version = value
		case "ID":
			Release.ID = value
		case "ID_LIKE":
			Release.IDLike = value
		case "PRETTY_NAME":
			Release.PrettyName = value
		case "VERSION_ID":
			Release.VersionID = value
		case "HOME_URL":
			Release.HomeURL = value
			// case "DOCUMENTATION_URL":
			// 	Release.DocumentationURL = value
			// case "SUPPORT_URL":
			// 	Release.SupportURL = value
			// case "BUG_REPORT_URL":
			// 	Release.BugReportURL = value
			// case "PRIVACY_POLICY_URL":
			// 	Release.PrivacyPolicyURL = value
			// case "VERSION_CODENAME":
			// 	Release.VersionCodename = value
			// case "UBUNTU_CODENAME":
			// 	Release.UbuntuCodename = value
			// case "ANSI_COLOR":
			// 	Release.ANSIColor = value
			// case "CPE_NAME":
			// 	Release.CPEName = value
			// case "BUILD_ID":
			// 	Release.BuildID = value
			// case "VARIANT":
			// 	Release.Variant = value
			// case "VARIANT_ID":
			// 	Release.VariantID = value
			// case "LOGO":
			// 	Release.Logo = value
			// default:
			// 	return fmt.Errorf("unknown key found: %s", key)
		}
	}

	return nil
}
