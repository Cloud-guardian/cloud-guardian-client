package linux_redhat_dnf

import (
	"testing"
)

const testCaseDnfInstalled = `Installed Packages
alternatives.x86_64                                            1.24-1.el9_5.1                                             @baseos   
attr.x86_64                                                    2.5.1-3.el9                                                @baseos   
audit-libs.x86_64                                              3.1.5-1.el9                                                @baseos   
basesystem.noarch                                              11-13.el9.0.1                                              @baseos   
bash.x86_64                                                    5.1.8-9.el9                                                @baseos   
binutils.x86_64                                                2.35.2-54.el9                                              @baseos   
binutils-gold.x86_64                                           2.35.2-54.el9                                              @baseos   
bzip2-libs.x86_64                                              1.0.8-8.el9                                                @baseos   
ca-certificates.noarch                                         2024.2.69_v8.0.303-91.4.el9_4                              @baseos   
coreutils-single.x86_64                                        8.32-36.el9                                                @baseos   
cracklib.x86_64                                                2.9.6-27.el9                                               @baseos   
`

const testCaseDnfCheckUpdate1 = `

NetworkManager.x86_64                                                    1:1.48.10-8.el9_5                                                     baseos
NetworkManager-libnm.x86_64                                              1:1.48.10-8.el9_5                                                     baseos
NetworkManager-team.x86_64                                               1:1.48.10-8.el9_5                                                     baseos
NetworkManager-tui.x86_64                                                1:1.48.10-8.el9_5                                                     baseos
consul.x86_64                                                            1.21.1-1                                                              hashicorp
containerd.io.x86_64                                                     1.7.27-3.1.el9                                                        docker-ce-stable
docker-buildx-plugin.x86_64                                              0.23.0-1.el9                                                          docker-ce-stable
docker-ce.x86_64                                                         3:28.1.1-1.el9                                                        docker-ce-stable
docker-ce-cli.x86_64                                                     1:28.1.1-1.el9                                                        docker-ce-stable
docker-ce-rootless-extras.x86_64                                         28.1.1-1.el9                                                          docker-ce-stable
docker-compose-plugin.x86_64                                             2.35.1-1.el9                                                          docker-ce-stable
dracut.x86_64                                                            057-80.git20250411.el9_5                                              baseos
dracut-config-rescue.x86_64                                              057-80.git20250411.el9_5                                              baseos
dracut-network.x86_64                                                    057-80.git20250411.el9_5                                              baseos
dracut-squash.x86_64                                                     057-80.git20250411.el9_5                                              baseos
e2fsprogs.x86_64                                                         1.46.5-6.el9_5                                                        baseos
e2fsprogs-libs.x86_64                                                    1.46.5-6.el9_5                                                        baseos
emacs-filesystem.noarch                                                  1:27.2-11.el9_5.2                                                     appstream
epel-release.noarch                                                      9-10.el9                                                              epel
expat.x86_64                                                             2.5.0-3.el9_5.3                                                       baseos
freetype.x86_64                                                          2.10.4-10.el9_5                                                       baseos
glibc.x86_64                                                             2.34-125.el9_5.8                                                      baseos
glibc-common.x86_64                                                      2.34-125.el9_5.8                                                      baseos
glibc-devel.x86_64                                                       2.34-125.el9_5.8                                                      appstream
glibc-gconv-extra.x86_64                                                 2.34-125.el9_5.8                                                      baseos
glibc-headers.x86_64                                                     2.34-125.el9_5.8                                                      appstream
glibc-langpack-en.x86_64                                                 2.34-125.el9_5.8                                                      baseos
golang.x86_64                                                            1.23.6-2.el9_5                                                        appstream
golang-bin.x86_64                                                        1.23.6-2.el9_5                                                        appstream
golang-race.x86_64                                                       1.23.6-2.el9_5                                                        appstream
golang-src.noarch                                                        1.23.6-2.el9_5                                                        appstream
grub2-common.noarch                                                      1:2.06-94.el9_5                                                       baseos
grub2-efi-x64.x86_64                                                     1:2.06-94.el9_5                                                       baseos
grub2-pc.x86_64                                                          1:2.06-94.el9_5                                                       baseos
grub2-pc-modules.noarch                                                  1:2.06-94.el9_5                                                       baseos
grub2-tools.x86_64                                                       1:2.06-94.el9_5                                                       baseos
grub2-tools-efi.x86_64                                                   1:2.06-94.el9_5                                                       baseos
grub2-tools-extra.x86_64                                                 1:2.06-94.el9_5                                                       baseos
grub2-tools-minimal.x86_64                                               1:2.06-94.el9_5                                                       baseos
kernel.x86_64                                                            5.14.0-503.40.1.el9_5                                                 baseos
kernel-core.x86_64                                                       5.14.0-503.40.1.el9_5                                                 baseos
kernel-headers.x86_64                                                    5.14.0-503.40.1.el9_5                                                 appstream
kernel-modules.x86_64                                                    5.14.0-503.40.1.el9_5                                                 baseos
kernel-modules-core.x86_64                                               5.14.0-503.40.1.el9_5                                                 baseos
kernel-tools.x86_64                                                      5.14.0-503.40.1.el9_5                                                 baseos
kernel-tools-libs.x86_64                                                 5.14.0-503.40.1.el9_5                                                 baseos
libcom_err.x86_64                                                        1.46.5-6.el9_5                                                        baseos
libnfsidmap.x86_64                                                       1:2.5.4-27.el9_5.1                                                    baseos
libss.x86_64                                                             1.46.5-6.el9_5                                                        baseos
libwbclient.x86_64                                                       4.20.2-2.el9_5.1                                                      baseos
libxml2.x86_64                                                           2.9.13-6.el9_5.2                                                      baseos
libxslt.x86_64                                                           1.1.34-9.el9_5.3                                                      appstream
linux-firmware.noarch                                                    20250415-146.5.el9_5                                                  baseos
linux-firmware-whence.noarch                                             20250415-146.5.el9_5                                                  baseos
mdadm.x86_64                                                             4.3-4.el9_5                                                           baseos
microcode_ctl.noarch                                                     4:20240910-1.20250211.1.el9_5                                         baseos
pciutils-libs.x86_64                                                     3.7.0-5.el9_5.1                                                       baseos
python3-perf.x86_64                                                      5.14.0-503.40.1.el9_5                                                 baseos
qemu-guest-agent.x86_64                                                  17:9.0.0-10.el9_5.3                                                   appstream
rocky-gpg-keys.noarch                                                    9.5-1.3.el9                                                           baseos
rocky-release.noarch                                                     9.5-1.3.el9                                                           baseos
rocky-repos.noarch                                                       9.5-1.3.el9                                                           baseos
samba-client-libs.x86_64                                                 4.20.2-2.el9_5.1                                                      baseos
samba-common.noarch                                                      4.20.2-2.el9_5.1                                                      baseos
samba-common-libs.x86_64                                                 4.20.2-2.el9_5.1                                                      baseos
systemd.x86_64                                                           252-46.el9_5.3                                                        baseos
systemd-boot-unsigned.x86_64                                             252-46.el9_5.3                                                        appstream
systemd-libs.x86_64                                                      252-46.el9_5.3                                                        baseos
systemd-pam.x86_64                                                       252-46.el9_5.3                                                        baseos
systemd-resolved.x86_64                                                  252-46.el9_5.3                                                        baseos
systemd-rpm-macros.noarch                                                252-46.el9_5.3                                                        baseos
systemd-udev.x86_64                                                      252-46.el9_5.3                                                        baseos
tzdata.noarch                                                            2025b-1.el9                                                           baseos
tzdata-java.noarch                                                       2025b-1.el9                                                           appstream
webkit2gtk3-jsc.x86_64                                                   2.48.1-1.el9_5                                                        appstream
zabbix-agent2.x86_64                                                     7.0.13-rc1.release1.el9                                               zabbix
Obsoleting Packages
grub2-tools.x86_64                                                       1:2.06-94.el9_5                                                       baseos
    grub2-tools.x86_64                                                   1:2.06-93.el9_5                                                       @baseos
grub2-tools-efi.x86_64                                                   1:2.06-94.el9_5                                                       baseos
    grub2-tools.x86_64                                                   1:2.06-93.el9_5                                                       @baseos
grub2-tools-extra.x86_64                                                 1:2.06-94.el9_5                                                       baseos
    grub2-tools.x86_64                                                   1:2.06-93.el9_5                                                       @baseos
grub2-tools-minimal.x86_64                                               1:2.06-94.el9_5                                                       baseos
    grub2-tools.x86_64                                                   1:2.06-93.el9_5                                                       @baseos
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
	const expectedPackageCount = 11
	const expectedPackageName = "bash.x86_64"
	const expectedPackageVersion = "5.1.8-9.el9"

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
	const expectedUpdate = "systemd-pam.x86_64 252-46.el9_5.3 baseos"
	const expectedUpdates = 76
	updates, obsolete := parseUpdates(testCaseDnfCheckUpdate1)

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

	if len(obsolete) == 0 {
		t.Error("Expected obsolete packages, but got none")
	}
}

func TestParseUpdatesNoObsolete(t *testing.T) {
	updates, obsolete := parseUpdates(testCaseDnfCheckUpdate2)

	if len(updates) == 0 {
		t.Error("Expected updates, but got none")
	}

	if len(obsolete) != 0 {
		t.Error("Expected no obsolete packages, but got some")
	}
}

func TestParseUpdatesNoUpdates(t *testing.T) {
	updates, obsolete := parseUpdates(testCaseDnfCheckUpdate3)

	if len(updates) != 0 {
		t.Error("Expected no updates, but got some")
	}

	if len(obsolete) != 0 {
		t.Error("Expected no obsolete packages, but got some")
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
