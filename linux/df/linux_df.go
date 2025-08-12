package linux_df

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type Df struct {
	Source string
	FSType string
	Size   float64 // Size in KB
	Used   float64 // Used space in KB
	Avail  float64 // Available space in KB
	Target string  // Mount point
}

func GetDf() ([]Df, error) {
	fileSystemTypes := []string{"ext3", "ext4", "xfs", "vfat"}
	var typeFlags []string
	for _, fsType := range fileSystemTypes {
		typeFlags = append(typeFlags, "--type="+fsType)
	}
	args := append([]string{"--block-size=1K", "--local"}, typeFlags...)
	args = append(args, "--output=source,fstype,size,used,avail,target")
	command := exec.Command("df", args...)

	var out strings.Builder
	command.Stdout = &out
	err := command.Run()
	if err != nil {
		return nil, fmt.Errorf("command failed: %s", out.String())
	}

	return parseDfOutput(out.String()), nil
}

func parseDfOutput(output string) []Df {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		return nil
	}

	var dfList []Df

	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) != 6 {
			continue
		}

		var df Df
		df.Source = fields[0]
		df.FSType = fields[1]
		df.Size, _ = strconv.ParseFloat(fields[2], 64)
		df.Used, _ = strconv.ParseFloat(fields[3], 64)
		df.Avail, _ = strconv.ParseFloat(fields[4], 64)
		df.Target = fields[5]

		dfList = append(dfList, df)
	}
	return dfList
}
