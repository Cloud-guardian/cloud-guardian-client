package linux_lsblk

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const sys = "/sys/class/block"

type BlockDevice struct {
	Name       string `json:"name"`
	KName      string `json:"kname"`
	PKName     string `json:"pkname,omitempty"`
	UUID       string `json:"uuid,omitempty"`
	Label      string `json:"label,omitempty"`
	FSType     string `json:"fstype,omitempty"`
	Path       string `json:"path"`
	MajMin     string `json:"maj:min"`
	Size       uint64 `json:"size"`
	RO         bool   `json:"ro"`
	Type       string `json:"type"`
	Serial     string `json:"serial,omitempty"`
	Mountpoint string `json:"mountpoint,omitempty"`
	Vendor     string `json:"vendor,omitempty"`
	State      string `json:"state,omitempty"`
	WWN        string `json:"wwn,omitempty"`
	Model      string `json:"model,omitempty"`
}

// type Out struct {
// 	Blockdevices []*Row `json:"blockdevices"`
// }

func GetLsBlk() (blockdevices []*BlockDevice) {
	mounts := readMounts()
	labels := readLinks("/dev/disk/by-label")
	uuids := readLinks("/dev/disk/by-uuid")
	fstypes := readLinks("/dev/disk/by-type")

	// var rows []*Row

	entries, _ := os.ReadDir(sys)
	for _, e := range entries {
		n := e.Name()
		p := filepath.Join(sys, n)
		majmin := read(p + "/dev")

		blockdevices = append(blockdevices, &BlockDevice{
			Name:       n,
			KName:      n,
			PKName:     parent(n),
			Path:       "/dev/" + n,
			MajMin:     majmin,
			Size:       u(read(p+"/size")) * 512,
			RO:         read(p+"/ro") == "1",
			Type:       dtype(n, p),
			Serial:     read(p + "/device/serial"),
			Vendor:     read(p + "/device/vendor"),
			State:      read(p + "/device/state"),
			WWN:        read(p + "/device/wwn"),
			Model:      read(p + "/device/model"),
			Mountpoint: mounts[majmin],
			Label:      labels[n],
			UUID:       uuids[n],
			FSType:     fstypes[n],
		})
	}

	return blockdevices
}

/* helpers */

func readMounts() map[string]string {
	m := map[string]string{}
	f, _ := os.Open("/proc/self/mountinfo")
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(strings.Split(sc.Text(), " - ")[0])
		m[fields[2]] = fields[4]
	}
	return m
}

func readLinks(dir string) map[string]string {
	m := map[string]string{}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if t, _ := os.Readlink(filepath.Join(dir, e.Name())); t != "" {
			m[filepath.Base(t)] = e.Name()
		}
	}
	return m
}

func parent(n string) string {
	if i := strings.LastIndex(n, "p"); i > 0 && isNum(n[i+1:]) {
		return n[:i]
	}
	return strings.TrimRight(n, "0123456789")
}

func dtype(n, p string) string {
	if strings.HasPrefix(n, "loop") {
		return "loop"
	}
	if _, err := os.Stat(p + "/partition"); err == nil {
		return "part"
	}
	return "disk"
}

func read(p string) string {
	b, _ := os.ReadFile(p)
	return strings.TrimSpace(string(b))
}

func u(s string) uint64 {
	v, _ := strconv.ParseUint(s, 10, 64)
	return v
}

func isNum(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
