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
	Name       string  `json:"name"`
	KName      string  `json:"kname"`
	PKName     *string `json:"pkname"`
	UUID       *string `json:"uuid"`
	Label      *string `json:"label"`
	FSType     *string `json:"fstype"`
	Path       string  `json:"path"`
	MajMin     string  `json:"maj:min"`
	Size       uint64  `json:"size"`
	RO         bool    `json:"ro"`
	Type       string  `json:"type"`
	Serial     *string `json:"serial"`
	Mountpoint *string `json:"mountpoint"`
	Vendor     *string `json:"vendor"`
	State      *string `json:"state"`
	WWN        *string `json:"wwn"`
	Model      *string `json:"model"`
}

type Device struct {
	BlockDevice
	Slaves  []string
	Holders []string
}

type Output struct {
	Blockdevices []*BlockDevice `json:"blockdevices"`
}

func GetLsBlk() (blockdevices []*BlockDevice) {
	mounts := readMounts()
	labels := readLinks("/dev/disk/by-label")
	uuids := readLinks("/dev/disk/by-uuid")
	fstypes := readLinks("/dev/disk/by-type")

	devs := map[string]*Device{}

	entries, _ := os.ReadDir(sys)
	for _, e := range entries {
		n := e.Name()
		p := filepath.Join(sys, n)
		majmin := read(p + "/dev")

		dev := &Device{
			BlockDevice: BlockDevice{
				Name:       n,
				KName:      n,
				Path:       "/dev/" + n,
				MajMin:     majmin,
				Size:       u(read(p+"/size")) * 512,
				RO:         read(p+"/ro") == "1",
				Type:       detectType(n, p),
				Serial:     strPtr(read(p + "/device/serial")),
				Vendor:     strPtr(read(p + "/device/vendor")),
				Model:      strPtr(read(p + "/device/model")),
				WWN:        strPtr(read(p + "/device/wwn")),
				State:      strPtr(read(p + "/device/state")),
				Mountpoint: strPtr(mounts[majmin]),
				Label:      strPtr(labels[n]),
				UUID:       strPtr(uuids[n]),
				FSType:     strPtr(fstypes[n]),
			},
			Slaves:  listDir(p + "/slaves"),
			Holders: listDir(p + "/holders"),
		}

		devs[n] = dev
	}

	// Determine PKNAME from slaves
	for _, d := range devs {
		if len(d.Slaves) > 0 {
			d.PKName = &d.Slaves[0]
		}
	}

	// var blockdevices []*BlockDevice
	for _, d := range devs {
		blockdevices = append(blockdevices, &d.BlockDevice)
	}

	return blockdevices
	// json.NewEncoder(os.Stdout).Encode(Output{Blockdevices: rows})
}

/* ---------- helpers ---------- */

func detectType(name, path string) string {
	if _, err := os.Stat(path + "/partition"); err == nil {
		return "part"
	}
	if _, err := os.Stat(path + "/md"); err == nil {
		return "raid"
	}
	if strings.HasPrefix(name, "dm-") {
		u := read(path + "/dm/uuid")
		if strings.HasPrefix(u, "LVM-") {
			return "lvm"
		}
		if strings.HasPrefix(u, "CRYPT-") {
			return "crypt"
		}
		return "dm"
	}
	if strings.HasPrefix(name, "loop") {
		return "loop"
	}
	return "disk"
}

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
	ents, err := os.ReadDir(dir)
	if err != nil {
		return m
	}
	for _, e := range ents {
		if t, err := os.Readlink(filepath.Join(dir, e.Name())); err == nil {
			m[filepath.Base(t)] = e.Name()
		}
	}
	return m
}

func listDir(dir string) []string {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range ents {
		out = append(out, e.Name())
	}
	return out
}

func strPtr(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

func read(p string) string {
	b, err := os.ReadFile(p)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func u(s string) uint64 {
	v, _ := strconv.ParseUint(strings.TrimSpace(s), 10, 64)
	return v
}
