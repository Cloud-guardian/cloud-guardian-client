package linux_debian_apt

import (
	"fmt"
	"testing"
)

const testSample1 = `
Listing... Done
base-files/noble-updates 13ubuntu10.2 arm64 [upgradable from: 13ubuntu10.1]
gpgv/noble-updates,noble-security 2.4.4-2ubuntu17.2 arm64 [upgradable from: 2.4.4-2ubuntu17]
libattr1/noble-updates 1:2.5.2-1build1.1 arm64 [upgradable from: 1:2.5.2-1build1]
libc-bin/noble-updates,noble-security 2.39-0ubuntu8.4 arm64 [upgradable from: 2.39-0ubuntu8.3]
libc6/noble-updates,noble-security 2.39-0ubuntu8.4 arm64 [upgradable from: 2.39-0ubuntu8.3]
libcap2/noble-updates,noble-security 1:2.66-5ubuntu2.2 arm64 [upgradable from: 1:2.66-5ubuntu2]
libgmp10/noble-updates 2:6.3.0+dfsg-2ubuntu6.1 arm64 [upgradable from: 2:6.3.0+dfsg-2ubuntu6]
libgnutls30t64/noble-updates,noble-security 3.8.3-1.1ubuntu3.3 arm64 [upgradable from: 3.8.3-1.1ubuntu3.2]
libgpg-error0/noble-updates 1.47-3build2.1 arm64 [upgradable from: 1.47-3build2]
libidn2-0/noble-updates 2.3.7-2build1.1 arm64 [upgradable from: 2.3.7-2build1]
liblzma5/noble-updates,noble-security 5.6.1+really5.4.5-1ubuntu0.2 arm64 [upgradable from: 5.6.1+really5.4.5-1build0.1]
libmd0/noble-updates 1.1.0-2build1.1 arm64 [upgradable from: 1.1.0-2build1]
libpcre2-8-0/noble-updates 10.42-4ubuntu2.1 arm64 [upgradable from: 10.42-4ubuntu2]
libselinux1/noble-updates 3.5-2ubuntu2.1 arm64 [upgradable from: 3.5-2ubuntu2]
libssl3t64/noble-updates,noble-security 3.0.13-0ubuntu3.5 arm64 [upgradable from: 3.0.13-0ubuntu3.4]
libsystemd0/noble-updates 255.4-1ubuntu8.6 arm64 [upgradable from: 255.4-1ubuntu8.4]
libtasn1-6/noble-updates,noble-security 4.19.0-3ubuntu0.24.04.1 arm64 [upgradable from: 4.19.0-3build1]
libudev1/noble-updates 255.4-1ubuntu8.6 arm64 [upgradable from: 255.4-1ubuntu8.4]
libunistring5/noble-updates 1.1-2build1.1 arm64 [upgradable from: 1.1-2build1]
perl-base/noble-updates,noble-security 5.38.2-3.2ubuntu0.1 arm64 [upgradable from: 5.38.2-3.2build2]
`

func TestParseUpdates(t *testing.T) {
	expectedUpdate := AptPackage{ // "libidn2-0 1:2.66-5ubuntu2.2 noble-updates,noble-security"
		Name:    "libidn2-0",
		Version: "2.3.7-2build1.1",
		Repo:    "noble-updates",
	}
	expectedUpdateCount := 20

	updates, obsolete := parseUpdates(testSample1, AllUpdates)

	if len(updates) != expectedUpdateCount {
		t.Errorf("Expected %d updates, got %d", expectedUpdateCount, len(updates))
	}

	if len(obsolete) != 0 {
		t.Errorf("Expected 0 obsolete packages, got %d", len(obsolete))
	}

	// Check if expected update is present
	found := false
	for _, update := range updates {
		fmt.Println(update.Name + " - " + update.Version + " (" + update.Repo + ")")
		if update == expectedUpdate {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected update %+v not found in updates", expectedUpdate)
	}

}
