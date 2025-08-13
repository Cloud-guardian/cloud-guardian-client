package linux_top

import (
	"bufio"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

// GetUptime retrieves the system uptime in seconds by reading from /proc/uptime.
//
// Returns:
//   - int64: System uptime in seconds
//   - error: Any error that occurred while reading or parsing the uptime
func GetUptime() (int64, error) {
	// Read the /proc/uptime file to get system uptime
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0, err
	}

	// Split the data into two parts: uptime and idle time
	parts := string(data)
	uptime := parts[:len(parts)-1] // Remove the trailing newline character
	uptimeParts := strings.Split(uptime, " ")
	// uptimeStr := strings.TrimSpace(string(uptime))
	uptimeSeconds, err := strconv.ParseFloat(uptimeParts[0], 64)
	if err != nil {
		return 0, err
	}
	uptimeSecondsInt := int64(uptimeSeconds)
	return uptimeSecondsInt, nil
}

type MemoryUsage struct {
	Total        float64
	Free         float64
	Used         float64
	Buffers      float64
	Cached       float64
	Available    float64
	Committed_As float64
	SwapTotal    float64
	SwapFree     float64
	SwapUsed     float64
}

// GetMemory retrieves memory usage information from /proc/meminfo.
//
// Returns:
//   - MemoryUsage: A struct containing detailed memory usage statistics
func GetMemory() MemoryUsage {
	mem := map[string]float64{}
	file, _ := os.Open("/proc/meminfo")
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			val, _ := strconv.ParseFloat(fields[1], 64)
			mem[fields[0][:len(fields[0])-1]] = val / 1024 // kB to MiB
		}
	}

	cached := mem["Cached"] + mem["SReclaimable"] - mem["Shmem"]
	used := mem["MemTotal"] - mem["MemFree"] - mem["Buffers"] - cached

	return MemoryUsage{
		Total:        round(mem["MemTotal"], 2),
		Free:         round(mem["MemFree"], 2),
		Used:         round(used, 2),
		Buffers:      round(mem["Buffers"], 2),
		Cached:       round(cached, 2),
		Available:    round(mem["MemAvailable"], 2),
		Committed_As: round(mem["Committed_AS"], 2),
		SwapTotal:    round(mem["SwapTotal"], 2),
		SwapFree:     round(mem["SwapFree"], 2),
		SwapUsed:     round(mem["SwapTotal"]-mem["SwapFree"], 2),
	}
}

type LoadAverage struct {
	OneMinute      float64
	FiveMinutes    float64
	FifteenMinutes float64
}

// GetLoad retrieves the system load averages from /proc/loadavg.
//
// Returns:
//   - LoadAverage: A struct containing 1, 5, and 15 minute load averages
func GetLoad() LoadAverage {
	// Load averages
	loadBytes, _ := os.ReadFile("/proc/loadavg")
	loadFields := strings.Fields(string(loadBytes))

	return LoadAverage{
		OneMinute:      parseLoad(loadFields[0]),
		FiveMinutes:    parseLoad(loadFields[1]),
		FifteenMinutes: parseLoad(loadFields[2]),
	}
}

// parseLoad parses a load average string value into a float64.
//
// Parameters:
//   - load: The load average string to parse
//
// Returns:
//   - float64: The parsed load average value, or 0.0 if parsing fails
func parseLoad(load string) float64 {
	value, err := strconv.ParseFloat(load, 64)
	if err != nil {
		log.Printf("Error parsing load average: %v\n", err)
		return 0.0
	}
	return value
}

type CpuUsage struct {
	User              float64
	System            float64
	Nice              float64
	Idle              float64
	IOWait            float64
	HardwareInterrupt float64
	SoftwareInterrupt float64
	Steal             float64
}

// GetCpuUsage calculates CPU usage percentages by taking two snapshots of /proc/stat.
// It computes the percentage of time spent in different CPU states.
//
// Returns:
//   - CpuUsage: A struct containing CPU usage percentages for different states
func GetCpuUsage() CpuUsage {
	stat1 := readCpuStat()
	time.Sleep(100 * time.Millisecond)
	stat2 := readCpuStat()

	total1 := sum(stat1)
	total2 := sum(stat2)

	deltaTotal := float64(total2 - total1)
	deltas := make([]float64, len(stat1))
	for i := range stat1 {
		deltas[i] = float64(stat2[i]-stat1[i]) / deltaTotal * 100
	}

	return CpuUsage{
		User:              round(deltas[0], 2),
		System:            round(deltas[2], 2),
		Nice:              round(deltas[1], 2),
		Idle:              round(deltas[3], 2),
		IOWait:            round(deltas[4], 2),
		HardwareInterrupt: round(deltas[5], 2),
		SoftwareInterrupt: round(deltas[6], 2),
		Steal:             round(deltas[7], 2),
	}
}

// readCpuStat reads CPU statistics from /proc/stat and returns the numeric values.
//
// Returns:
//   - []int64: Array of CPU time values from /proc/stat, or nil if reading fails
func readCpuStat() []int64 {
	file, _ := os.Open("/proc/stat")
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)[1:]
			values := make([]int64, len(fields))
			for i, f := range fields {
				values[i], _ = strconv.ParseInt(f, 10, 64)
			}
			return values
		}
	}
	return nil
}

type CpuInfo struct {
	ModelName string
	Cores     int
	Threads   int
	Mhz       float64
}

// GetCpuInfo retrieves CPU information from /proc/cpuinfo.
//
// Returns:
//   - CpuInfo: A struct containing CPU model name, core count, thread count, and frequency
func GetCpuInfo() CpuInfo {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return CpuInfo{}
	}
	defer file.Close()

	var modelName string
	var cores, threads int
	var mhz float64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "model name") {
			modelName = extractStringFromKV(line)
		} else if strings.HasPrefix(line, "cpu cores") {
			cores = extractIntFromKV(line)
		} else if strings.HasPrefix(line, "siblings") {
			threads = extractIntFromKV(line)
		} else if strings.HasPrefix(line, "cpu MHz") {
			mhz = extractFloatFromKV(line)
		}
	}

	return CpuInfo{
		ModelName: modelName,
		Cores:     cores,
		Threads:   threads,
		Mhz:       round(mhz, 2),
	}
}

type TaskStats struct {
	Total           int
	Running         int
	Sleeping        int
	Stopped         int
	Zombie          int
	Uninterruptible int
	Idle            int
}

// GetTasks retrieves task/process statistics by scanning /proc directory.
// It counts processes in different states by reading their status files.
//
// Returns:
//   - TaskStats: A struct containing counts of processes in different states
func GetTasks() TaskStats {
	total, running, sleeping, uninterruptible, idle, stopped, zombie := 0, 0, 0, 0, 0, 0, 0
	proc, _ := os.ReadDir("/proc")
	for _, entry := range proc {
		if pid := entry.Name(); isNumeric(pid) {
			if stat, err := os.ReadFile("/proc/" + pid + "/status"); err == nil {
				for _, line := range strings.Split(string(stat), "\n") {
					if strings.HasPrefix(line, "State:") {
						fields := strings.Fields(line)
						if len(fields) > 1 {
							state := fields[1]
							switch state {
							case "R":
								running++
							case "S":
								sleeping++
							case "D":
								uninterruptible++
							case "I":
								idle++
							case "T", "t":
								stopped++
							case "Z":
								zombie++
							}
							total++ // increment only when we know it's a process
						}
						break
					}
				}
			}
		}
	}
	return TaskStats{
		Total:           total,
		Running:         running,
		Sleeping:        sleeping + uninterruptible + idle, // Combine sleeping, uninterruptible, and idle states
		Stopped:         stopped,
		Zombie:          zombie,
		Uninterruptible: uninterruptible,
		Idle:            idle,
	}
}

// round rounds a float64 value to the specified number of decimal places.
//
// Parameters:
//   - value: The float64 value to round
//   - precision: The number of decimal places to round to
//
// Returns:
//   - float64: The rounded value
func round(value float64, precision int) float64 {
	pow := math.Pow(10, float64(precision))
	return math.Round(value*pow) / pow
}

// isNumeric checks if a string represents a numeric value.
//
// Parameters:
//   - s: The string to check
//
// Returns:
//   - bool: true if the string is numeric, false otherwise
func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// sum calculates the sum of all values in an int64 slice.
//
// Parameters:
//   - arr: The slice of int64 values to sum
//
// Returns:
//   - int64: The sum of all values in the slice
func sum(arr []int64) int64 {
	var total int64
	for _, v := range arr {
		total += v
	}
	return total
}

// extractIntFromKV extracts an integer value from a key-value line.
// It parses lines in the format "key: value" and returns the integer value.
//
// Parameters:
//   - line: The key-value line to parse
//
// Returns:
//   - int: The extracted integer value, or 0 if parsing fails
func extractIntFromKV(line string) int {
	result, err := strconv.Atoi(extractStringFromKV(line))
	if err != nil {
		return 0 // Default to 0 if parsing fails
	}
	return result
}

// extractStringFromKV extracts a string value from a key-value line.
// It parses lines in the format "key: value" and returns the trimmed string value.
//
// Parameters:
//   - line: The key-value line to parse
//
// Returns:
//   - string: The extracted and trimmed string value
func extractStringFromKV(line string) string {
	return strings.TrimSpace(strings.Split(line, ":")[1])
}

// extractFloatFromKV extracts a float64 value from a key-value line.
// It parses lines in the format "key: value" and returns the float64 value.
//
// Parameters:
//   - line: The key-value line to parse
//
// Returns:
//   - float64: The extracted float64 value, or 0.0 if parsing fails
func extractFloatFromKV(line string) float64 {
	result, err := strconv.ParseFloat(extractStringFromKV(line), 64)
	if err != nil {
		return 0.0 // Default to 0 if parsing fails
	}
	return result
}
