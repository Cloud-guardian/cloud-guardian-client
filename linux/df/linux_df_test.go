package linux_df

import (
	"reflect"
	"testing"
)

type testCase struct {
	testCase       string
	expectedResult []Df
}

var testCases = []testCase{
	{
		testCase: `Filesystem                        Type 1K-blocks    Used   Avail Mounted on
/dev/mapper/ubuntu--vg-ubuntu--lv ext4  15371208 6316136 8252464 /
/dev/sda2                         ext4   1992552  196508 1674804 /boot
`,
		expectedResult: []Df{
			{
				Source: "/dev/mapper/ubuntu--vg-ubuntu--lv",
				FSType: "ext4",
				Size:   15371208,
				Used:   6316136,
				Avail:  8252464,
				Target: "/",
			},
			{
				Source: "/dev/sda2",
				FSType: "ext4",
				Size:   1992552,
				Used:   196508,
				Avail:  1674804,
				Target: "/boot",
			},
		},
	},
	{
		testCase: `Filesystem           Type  1K-blocks       Used     Avail Mounted on
/dev/mapper/vg0-root ext4 1835140560 1279740136 470256536 /
/dev/md2             ext3    1012428     882180     77924 /boot
/dev/md0             vfat     261804       7212    254592 /boot/efi
`,
		expectedResult: []Df{
			{
				Source: "/dev/mapper/vg0-root",
				FSType: "ext4",
				Size:   1835140560,
				Used:   1279740136,
				Avail:  470256536,
				Target: "/",
			},
			{
				Source: "/dev/md2",
				FSType: "ext3",
				Size:   1012428,
				Used:   882180,
				Avail:  77924,
				Target: "/boot",
			},
			{
				Source: "/dev/md0",
				FSType: "vfat",
				Size:   261804,
				Used:   7212,
				Avail:  254592,
				Target: "/boot/efi",
			},
		},
	},
}

func TestParseDfOutput(t *testing.T) {
	for _, testCase := range testCases {
		t.Run(testCase.testCase, func(t *testing.T) {
			result := parseDfOutput(testCase.testCase)
			if !reflect.DeepEqual(result, testCase.expectedResult) {
				t.Errorf("Expected %v, got %v", testCase.expectedResult, result)
			}
		})
	}
}
