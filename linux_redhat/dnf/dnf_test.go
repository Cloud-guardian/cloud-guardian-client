package linux_redhat_dnf

import (
	"testing"
)

const testCaseDnfInstalled = `
NetworkManager-tui.x86_64 1:1.54.0-3.el9_7 @baseos
bubblewrap.x86_64 0:0.6.3-1.el9 @baseos
bzip2-libs.x86_64 0:1.0.8-10.el9_5 @anaconda
yum.noarch 0:4.14.0-31.el9.rocky.0.1 @baseos
zabbix-agent2-plugin-ember-plus.x86_64 0:7.0.23-rc2.release1.el9 @zabbix
`

const testCaseDnfCheckUpdate1 = `
docker-ce.x86_64 3:29.3.1-1.el9 docker-ce-stable
docker-compose-plugin.x86_64 0:5.1.1-1.el9 docker-ce-stable
glibc-common.x86_64 0:2.34-231.el9_7.10 baseos
glibc-devel.x86_64 0:2.34-231.el9_7.10 appstream
glibc-gconv-extra.x86_64 0:2.34-231.el9_7.10 baseos
glibc-headers.x86_64 0:2.34-231.el9_7.10 appstream
glibc-langpack-en.x86_64 0:2.34-231.el9_7.10 baseos
glibc.x86_64 0:2.34-231.el9_7.10 baseos
selinux-policy.noarch 0:38.1.65-1.el9_7.1 baseos
util-linux-core.x86_64 0:2.37.4-21.el9_7 baseos
util-linux.x86_64 0:2.37.4-21.el9_7 baseos
zabbix-agent2-plugin-ember-plus.x86_64 0:7.0.24-rc3.release1.el9 zabbix
zabbix-agent2-plugin-mongodb.x86_64 0:7.0.24-rc3.release1.el9 zabbix
`

const testCaseDnfCheckUpdate2 = `

consul.x86_64                                                                  1.21.1-1                                                                     hashicorp
dracut.x86_64                                                                  057-80.git20250411.el9_5                                                     baseos
dracut-config-rescue.x86_64                                                    057-80.git20250411.el9_5                                                     baseos
dracut-network.x86_64                                                          057-80.git20250411.el9_5                                                     baseos
dracut-squash.x86_64                                                           057-80.git20250411.el9_5                                                     baseos
e2fsprogs.x86_64                                                               1.46.5-6.el9_5                                                               baseos
e2fsprogs-libs.x86_64                                                          1.46.5-6.el9_5                                                               baseos
epel-release.noarch                                                            9-10.el9                                                                     epel
kernel.x86_64                                                                  5.14.0-503.40.1.el9_5                                                        baseos
kernel-core.x86_64                                                             5.14.0-503.40.1.el9_5                                                        baseos
kernel-modules.x86_64                                                          5.14.0-503.40.1.el9_5                                                        baseos
kernel-modules-core.x86_64                                                     5.14.0-503.40.1.el9_5                                                        baseos
kernel-tools.x86_64                                                            5.14.0-503.40.1.el9_5                                                        baseos
kernel-tools-libs.x86_64                                                       5.14.0-503.40.1.el9_5                                                        baseos
libcom_err.x86_64                                                              1.46.5-6.el9_5                                                               baseos
libss.x86_64                                                                   1.46.5-6.el9_5                                                               baseos
pciutils-libs.x86_64                                                           3.7.0-5.el9_5.1                                                              baseos
qemu-guest-agent.x86_64                                                        17:9.0.0-10.el9_5.3                                                          appstream
`

const testCaseDnfCheckUpdate3 = ``

func TestParseInstalledPackages(t *testing.T) {
	const expectedPackageCount = 5
	const expectedPackageName = "bubblewrap.x86_64"
	const expectedPackageVersion = "0:0.6.3-1.el9"

	packages := parseInstalledPackages(testCaseDnfInstalled)

	if len(packages) != expectedPackageCount {
		t.Errorf("Expected %d installed packages, got %d", expectedPackageCount, len(packages))
	}

	found := false
	for _, pkg := range packages {
		if pkg.Name == expectedPackageName && pkg.Version == expectedPackageVersion {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected package %s with version %s not found in installed packages", expectedPackageName, expectedPackageVersion)
	}
}

func TestParseUpdates(t *testing.T) {
	const expectedUpdate = "util-linux-core.x86_64 0:2.37.4-21.el9_7 baseos"
	const expectedUpdates = 13
	updates := parseUpdates(testCaseDnfCheckUpdate1)

	if len(updates) == 0 {
		t.Error("Expected updates, but got none")
	} else {
		if len(updates) != expectedUpdates {
			t.Errorf("Expected %d updates, but got %d", expectedUpdates, len(updates))
		}
		var foundExpectedUpdate bool
		foundExpectedUpdate = false
		for _, update := range updates {
			if update.Name+" "+update.Version+" "+update.Repo == expectedUpdate {
				foundExpectedUpdate = true
			}
		}
		if !foundExpectedUpdate {
			t.Errorf("Expected update '%s' not found in updates", expectedUpdate)
		}
	}
}

func TestParseUpdatesNoObsolete(t *testing.T) {
	updates := parseUpdates(testCaseDnfCheckUpdate2)

	if len(updates) == 0 {
		t.Error("Expected updates, but got none")
	}
}

func TestParseUpdatesNoUpdates(t *testing.T) {
	updates := parseUpdates(testCaseDnfCheckUpdate3)

	if len(updates) != 0 {
		t.Error("Expected no updates, but got some")
	}
}

// Update summary test cases
const testCaseUpdateSummary = `Updates Information Summary: available
    8 Security notice(s)
        3 Important Security notice(s)
        5 Moderate Security notice(s)
    3 Bugfix notice(s)
    1 Enhancement notice(s)
`

const testCaseUpdateSummaryEmpty = `
`

const testCaseUpdateSummaryNoUpdates = `Updates Information Summary: available
`

const testCaseOnlySecurityUpdates = `Updates Information Summary: available
    8 Security notice(s)
        3 Important Security notice(s)
        5 Moderate Security notice(s)
`

func TestParseUpdateSummary(t *testing.T) {
	summary := parseUpdateSummary(testCaseUpdateSummary)

	expectedSummary := DnfUpdateSummary{
		SecuritImportant: 3,
		SecurityModerate: 5,
		Bugfix:           3,
		Enhancement:      1,
	}

	if summary != expectedSummary {
		t.Errorf("Expected summary %+v, got %+v", expectedSummary, summary)
	}
}

func TestParseUpdateSummaryEmpty(t *testing.T) {
	summary := parseUpdateSummary(testCaseUpdateSummaryEmpty)

	expectedSummary := DnfUpdateSummary{
		SecuritImportant: 0,
		SecurityModerate: 0,
		Bugfix:           0,
		Enhancement:      0,
	}

	if summary != expectedSummary {
		t.Errorf("Expected empty summary %+v, got %+v", expectedSummary, summary)
	}
}

func TestParseUpdateSummaryNoUpdates(t *testing.T) {
	summary := parseUpdateSummary(testCaseUpdateSummaryNoUpdates)

	expectedSummary := DnfUpdateSummary{
		SecuritImportant: 0,
		SecurityModerate: 0,
		Bugfix:           0,
		Enhancement:      0,
	}

	if summary != expectedSummary {
		t.Errorf("Expected no updates summary %+v, got %+v", expectedSummary, summary)
	}
}

func TestParseUpdateSummaryOnlySecurityUpdates(t *testing.T) {
	summary := parseUpdateSummary(testCaseOnlySecurityUpdates)

	expectedSummary := DnfUpdateSummary{
		SecuritImportant: 3,
		SecurityModerate: 5,
		Bugfix:           0,
		Enhancement:      0,
	}

	if summary != expectedSummary {
		t.Errorf("Expected only security updates summary %+v, got %+v", expectedSummary, summary)
	}
}
