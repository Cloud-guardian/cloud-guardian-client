package linux_mdstat

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type MdStat struct {
	Personalities []string `json:"personalities"`
	Arrays        []Array  `json:"arrays"`
}

type Array struct {
	Name        string     `json:"name"`
	Level       string     `json:"level"`
	State       string     `json:"state"`
	Devices     []Device   `json:"devices"`
	Blocks      uint64     `json:"blocks"`
	RaidDisks   int        `json:"raid_disks"`
	ActiveDisks int        `json:"active_disks"`
	Health      string     `json:"health"`
	Progress    *Progress  `json:"progress,omitempty"`
	Sys         *SysBlock  `json:"sys,omitempty"`
}

type Device struct {
	Name  string `json:"name"`
	Slot  int    `json:"slot"`
	Size  uint64 `json:"size_blocks,omitempty"`
}

type Progress struct {
	Type     string  `json:"type"`
	Percent  float64 `json:"percent"`
	SpeedKPS int64   `json:"speed_kps"`
	ETA      string  `json:"eta"`
}

type SysBlock struct {
	ChunkSize string `json:"chunk_size"`
	Layout    string `json:"layout"`
	Metadata  string `json:"metadata"`
}

var (
	devRe    = regexp.MustCompile(`(\w+)\[(\d+)\]`)
	sizeRe   = regexp.MustCompile(`(\d+)\s+blocks.*\[(\d+)/(\d+)\]\s+\[([U_]+)\]`)
	progRe   = regexp.MustCompile(`\((recovery|resync|rebuild)=\s*([\d.]+)%.*?speed=([\d]+)K/sec.*?finish=([^)]+)\)`)
)

func GetMdStat() (mdstat MdStat) {
	partitions := parsePartitions()

	f, _ := os.Open("/proc/mdstat")
	defer f.Close()
	s := bufio.NewScanner(f)

	// var mdstat MdStat
	var cur *Array

	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "Personalities") {
			for _, p := range strings.Fields(line) {
				if strings.HasPrefix(p, "[") {
					mdstat.Personalities = append(mdstat.Personalities, strings.Trim(p, "[]"))
				}
			}
			continue
		}

		if strings.Contains(line, " : ") {
			f := strings.Fields(line)
			cur = &Array{
				Name:  strings.TrimSuffix(f[0], ":"),
				State: f[2],
				Level: f[3],
			}

			for _, m := range devRe.FindAllStringSubmatch(line, -1) {
				slot, _ := strconv.Atoi(m[2])
				cur.Devices = append(cur.Devices, Device{
					Name: m[1],
					Slot: slot,
					Size: partitions[m[1]],
				})
			}
			mdstat.Arrays = append(mdstat.Arrays, *cur)
			continue
		}

		if cur == nil {
			continue
		}

		if m := sizeRe.FindStringSubmatch(line); m != nil {
			blocks, _ := strconv.ParseUint(m[1], 10, 64)
			rd, _ := strconv.Atoi(m[2])
			ad, _ := strconv.Atoi(m[3])
			a := &mdstat.Arrays[len(mdstat.Arrays)-1]
			a.Blocks = blocks
			a.RaidDisks = rd
			a.ActiveDisks = ad
			a.Health = m[4]
		}

		if m := progRe.FindStringSubmatch(line); m != nil {
			pct, _ := strconv.ParseFloat(m[2], 64)
			speed, _ := strconv.ParseInt(m[3], 10, 64)
			mdstat.Arrays[len(mdstat.Arrays)-1].Progress = &Progress{
				Type:    m[1],
				Percent: pct,
				SpeedKPS: speed,
				ETA:     strings.TrimSpace(m[4]),
			}
		}
	}

	// Correlate sysfs
	for i := range mdstat.Arrays {
		base := filepath.Join("/sys/block", mdstat.Arrays[i].Name, "md")
		mdstat.Arrays[i].Sys = &SysBlock{
			ChunkSize: readFirst(filepath.Join(base, "chunk_size")),
			Layout:    readFirst(filepath.Join(base, "layout")),
			Metadata:  readFirst(filepath.Join(base, "metadata_version")),
		}
	}
	return mdstat
}

func readFirst(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func parsePartitions() map[string]uint64 {
	out := map[string]uint64{}
	f, _ := os.Open("/proc/partitions")
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		f := strings.Fields(s.Text())
		if len(f) == 4 {
			if v, err := strconv.ParseUint(f[2], 10, 64); err == nil {
				out[f[3]] = v
			}
		}
	}
	return out
}
