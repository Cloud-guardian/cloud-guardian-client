package linux_needrestart

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type NeedRestart struct {
	RebootRequired bool                `json:"reboot_required"`
	Services       map[string][]string `json:"services"`
	Users          []string            `json:"users"`
	Containers     map[string][]string `json:"containers"`
}

func GetNeedRestart() (needRestart NeedRestart) {
	flag.Parse()

	needRestart = buildResult()

	return needRestart
}

func kernelNeedsReboot() bool {
	running, _ := os.ReadFile("/proc/sys/kernel/osrelease")
	modules, _ := filepath.Glob("/lib/modules/*")
	if len(modules) == 0 {
		return false
	}
	latest := filepath.Base(modules[len(modules)-1])
	return strings.TrimSpace(string(running)) != latest
}

var ignoredDeletedFiles = []string{
	"/dev/zero",
	"SYSV",
	"/memfd:",
	"/tmp",
}

func scanDeletedMappings() map[int][]string {
	result := map[int][]string{}
	filepath.WalkDir("/proc", func(p string, d fs.DirEntry, _ error) error {
		if !strings.HasSuffix(p, "/maps") {
			return nil
		}
		pid, err := strconv.Atoi(strings.Split(p, "/")[2])
		if err != nil {
			return nil
		}
		f, err := os.Open(p)
		if err != nil {
			return nil
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "(deleted)") {
				fields := strings.Fields(line)
				file := fields[len(fields)-2]
				for _, ign := range ignoredDeletedFiles {
					if strings.Contains(file, ign) {
						return nil
					}
				}
				result[pid] = append(result[pid], fields[len(fields)-2])
			}
		}
		return nil
	})
	return result
}

func serviceOfPID(pid int) string {
	data, _ := os.ReadFile(fmt.Sprintf("/proc/%d/cgroup", pid))
	re := regexp.MustCompile(`system.slice/(.+?).service`)
	m := re.FindStringSubmatch(string(data))
	if len(m) > 1 {
		return m[1]
	}
	return ""
}

func containerOfPID(pid int) string {
	data, _ := os.ReadFile(fmt.Sprintf("/proc/%d/cgroup", pid))
	if strings.Contains(string(data), "kubepods") {
		return "kubernetes"
	}
	if strings.Contains(string(data), "docker") {
		return "docker"
	}
	if strings.Contains(string(data), "libpod") {
		return "podman"
	}
	return ""
}

func buildResult() NeedRestart {
	deleted := scanDeletedMappings()
	res := NeedRestart{
		RebootRequired: kernelNeedsReboot(),
		Services:       map[string][]string{},
		Containers:     map[string][]string{},
	}

	users := map[string]bool{}

	for pid, files := range deleted {
		if svc := serviceOfPID(pid); svc != "" {
			res.Services[svc] = append(res.Services[svc], files...)
			continue
		}
		if ctr := containerOfPID(pid); ctr != "" {
			res.Containers[ctr] = append(res.Containers[ctr], files...)
			continue
		}
		status, _ := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
		for _, l := range strings.Split(string(status), "\n") {
			if strings.HasPrefix(l, "Uid:") {
				uid := strings.Fields(l)[1]
				if u, err := user.LookupId(uid); err == nil {
					users[u.Username] = true
				}
			}
		}
	}

	for u := range users {
		res.Users = append(res.Users, u)
	}
	return res
}
