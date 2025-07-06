package linux_osrelease

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

const testCaseRocky = `NAME="Rocky Linux"
VERSION="9.5 (Blue Onyx)"
ID="rocky"
ID_LIKE="rhel centos fedora"
VERSION_ID="9.5"
PLATFORM_ID="platform:el9"
PRETTY_NAME="Rocky Linux 9.5 (Blue Onyx)"
ANSI_COLOR="0;32"
LOGO="fedora-logo-icon"
CPE_NAME="cpe:/o:rocky:rocky:9::baseos"
HOME_URL="https://rockylinux.org/"
VENDOR_NAME="RESF"
VENDOR_URL="https://resf.org/"
BUG_REPORT_URL="https://bugs.rockylinux.org/"
SUPPORT_END="2032-05-31"
ROCKY_SUPPORT_PRODUCT="Rocky-Linux-9"
ROCKY_SUPPORT_PRODUCT_VERSION="9.5"
REDHAT_SUPPORT_PRODUCT="Rocky Linux"
REDHAT_SUPPORT_PRODUCT_VERSION="9.5"
`

const testCaseUbuntu2104 = `NAME="Ubuntu"
VERSION="21.04 (Hirsute Hippo)"
ID=ubuntu
ID_LIKE=debian
PRETTY_NAME="Ubuntu 21.04"
VERSION_ID="21.04"
HOME_URL="https://www.ubuntu.com/"
DOCUMENTATION_URL=""
SUPPORT_URL="https://help.ubuntu.com/"
BUG_REPORT_URL="https://bugs.launchpad.net/ubuntu/"
PRIVACY_POLICY_URL="https://www.ubuntu.com/legal/terms-and-policies/privacy-policy"
VERSION_CODENAME=hirsute
UBUNTU_CODENAME=hirsute
ANSI_COLOR=""
CPE_NAME=""
BUILD_ID=""
VARIANT=""
VARIANT_ID=""
LOGO=""
`

const testCaseUbuntu2204 = `PRETTY_NAME="Ubuntu 22.04.5 LTS"
NAME="Ubuntu"
VERSION_ID="22.04"
VERSION="22.04.5 LTS (Jammy Jellyfish)"
VERSION_CODENAME=jammy
ID=ubuntu
ID_LIKE=debian
HOME_URL="https://www.ubuntu.com/"
SUPPORT_URL="https://help.ubuntu.com/"
BUG_REPORT_URL="https://bugs.launchpad.net/ubuntu/"
PRIVACY_POLICY_URL="https://www.ubuntu.com/legal/terms-and-policies/privacy-policy"
UBUNTU_CODENAME=jammy`

// This test is written for Ubuntu 21.04
func TestParseUbuntu(t *testing.T) {

	// convert testCaseRocky to a slice of strings
	testCaseRockyLines := make([]string, 0)

	for _, line := range strings.Split(string(testCaseUbuntu2104), "\n") {

		switch true {
		case line == "":
			continue
		case []byte(line)[0] == '#':
			continue
		}

		testCaseRockyLines = append(testCaseRockyLines, line)
	}

	err := Parse(testCaseRockyLines)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse os-relese: %s\n", err)
		os.Exit(1)
	}

	switch true {
	case Release.Name != "Ubuntu":
		t.Errorf("Test failed on NAME: want 'Ubuntu', got '%s'\n", Release.Name)
	case Release.Version != "21.04 (Hirsute Hippo)":
		t.Errorf("Test failed on VERSION: want '21.04 (Hirsute Hippo)', got '%s'\n", Release.Version)
	case Release.ID != "ubuntu":
		t.Errorf("Test failed on ID: want 'ubuntu', got '%s'\n", Release.ID)
	case Release.IDLike != "debian":
		t.Errorf("Test failed on ID_LIKE: want 'debian', got '%s'\n", Release.IDLike)
	case Release.PrettyName != "Ubuntu 21.04":
		t.Errorf("Test failed on PRETTY_NAME: want 'Ubuntu 21.04', got '%s'\n", Release.PrettyName)
	case Release.VersionID != "21.04":
		t.Errorf("Test failed on VERSION_ID: want '21.04', got '%s'\n", Release.VersionID)
	case Release.HomeURL != "https://www.ubuntu.com/":
		t.Errorf("test failed on HOME_URL: want 'https://www.ubuntu.com/', got '%s'\n", Release.HomeURL)
	}
}

func TestParseUbuntu2204(t *testing.T) {
	// convert testCaseRocky to a slice of strings
	testCaseRockyLines := make([]string, 0)

	for _, line := range strings.Split(string(testCaseUbuntu2204), "\n") {

		switch true {
		case line == "":
			continue
		case []byte(line)[0] == '#':
			continue
		}

		testCaseRockyLines = append(testCaseRockyLines, line)
	}

	err := Parse(testCaseRockyLines)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse os-relese: %s\n", err)
		os.Exit(1)
	}

	switch true {
	case Release.Name != "Ubuntu":
		t.Errorf("Test failed on NAME: want 'Ubuntu', got '%s'\n", Release.Name)
	case Release.Version != "22.04.5 LTS (Jammy Jellyfish)":
		t.Errorf("Test failed on VERSION: want '22.04.5 LTS (Jammy Jellyfish)', got '%s'\n", Release.Version)
	case Release.ID != "ubuntu":
		t.Errorf("Test failed on ID: want 'ubuntu', got '%s'\n", Release.ID)
	case Release.IDLike != "debian":
		t.Errorf("Test failed on ID_LIKE: want 'debian', got '%s'\n", Release.IDLike)
	case Release.PrettyName != "Ubuntu 22.04.5 LTS":
		t.Errorf("Test failed on PRETTY_NAME: want 'Ubuntu 22.04.5 LTS', got '%s'\n", Release.PrettyName)
	case Release.VersionID != "22.04":
		t.Errorf("Test failed on VERSION_ID: want '22.04', got '%s'\n", Release.VersionID)
	case Release.HomeURL != "https://www.ubuntu.com/":
		t.Errorf("test failed on HOME_URL: want 'https://www.ubuntu.com/', got '%s'\n", Release.HomeURL)
	}
}

// This test is written for Rocky Linux 9.5
func TestParseRocky(t *testing.T) {

	// convert testCaseRocky to a slice of strings
	testCaseRockyLines := make([]string, 0)

	for _, line := range strings.Split(string(testCaseRocky), "\n") {

		switch true {
		case line == "":
			continue
		case []byte(line)[0] == '#':
			continue
		}

		testCaseRockyLines = append(testCaseRockyLines, line)
	}

	err := Parse(testCaseRockyLines)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse os-relese: %s\n", err)
		os.Exit(1)
	}

	switch true {
	case Release.Name != "Rocky Linux":
		t.Errorf("Test failed on NAME: want 'Rocky Linux', got '%s'\n", Release.Name)
	case Release.Version != "9.5 (Blue Onyx)":
		t.Errorf("Test failed on VERSION: want '9.5 (Blue Onyx)', got '%s'\n", Release.Version)
	case Release.ID != "rocky":
		t.Errorf("Test failed on ID: want 'rocky', got '%s'\n", Release.ID)
	case Release.IDLike != "rhel centos fedora":
		t.Errorf("Test failed on ID_LIKE: want 'rhel centos fedora', got '%s'\n", Release.IDLike)
	case Release.PrettyName != "Rocky Linux 9.5 (Blue Onyx)":
		t.Errorf("Test failed on PRETTY_NAME: want 'Rocky Linux 9.5 (Blue Onyx)', got '%s'\n", Release.PrettyName)
	case Release.VersionID != "9.5":
		t.Errorf("Test failed on VERSION_ID: want '9.5', got '%s'\n", Release.VersionID)
	case Release.HomeURL != "https://rockylinux.org/":
		t.Errorf("test failed on HOME_URL: want 'https://rockylinux.org/', got '%s'\n", Release.HomeURL)
	default:
		fmt.Println("All tests passed for Rocky Linux 9.5")
	}
}
