// Package server provides node identity and health metric collection.
package server

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Snapshot captures server health at a point in time.
type Snapshot struct {
	CPUUsage     float64   `json:"cpu_usage"`
	RAMUsage     float64   `json:"ram_usage"`
	RAMTotal     uint64    `json:"ram_total"`
	RAMUsed      uint64    `json:"ram_used"`
	DiskUsage    float64   `json:"disk_usage"`
	DiskTotal    uint64    `json:"disk_total"`
	DiskFree     uint64    `json:"disk_free"`
	LoadAvg1     float64   `json:"load_avg_1"`
	LoadAvg5     float64   `json:"load_avg_5"`
	LoadAvg15    float64   `json:"load_avg_15"`
	Uptime       float64   `json:"uptime_seconds"`
	Hostname     string    `json:"hostname"`
	ProcessCount int       `json:"process_count"`
	RecordedAt   time.Time `json:"recorded_at"`
}

// Collector gathers health metrics from the local system.
type Collector struct {
	dataDir string
	prevCPU struct {
		user, nice, system, idle, iowait, irq, softirq, steal uint64
		set bool
	}
}

// NewCollector creates a new health metric collector.
func NewCollector(dataDir string) *Collector {
	return &Collector{dataDir: dataDir}
}

// Collect gathers a full health snapshot.
func (c *Collector) Collect() Snapshot {
	s := Snapshot{RecordedAt: time.Now()}

	s.Hostname, _ = os.Hostname()

	// CPU
	s.CPUUsage = c.cpuUsage()

	// RAM
	s.RAMTotal, s.RAMUsed, s.RAMUsage = c.ramUsage()

	// Disk (data directory or /)
	path := "/"
	if c.dataDir != "" {
		path = c.dataDir
	}
	s.DiskTotal, s.DiskFree, s.DiskUsage = c.diskUsage(path)

	// Load
	s.LoadAvg1, s.LoadAvg5, s.LoadAvg15 = c.loadAvg()

	// Uptime
	s.Uptime = c.uptime()

	// Process count
	s.ProcessCount = c.processCount()

	return s
}

// cpuUsage reads /proc/stat and computes CPU usage since last call.
func (c *Collector) cpuUsage() float64 {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return 0
	}
	defer f.Close()

	var (
		user, nice, system, idle, iowait, irq, softirq, steal uint64
	)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "cpu ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 9 {
			return 0
		}
		user, _ = strconv.ParseUint(fields[1], 10, 64)
		nice, _ = strconv.ParseUint(fields[2], 10, 64)
		system, _ = strconv.ParseUint(fields[3], 10, 64)
		idle, _ = strconv.ParseUint(fields[4], 10, 64)
		iowait, _ = strconv.ParseUint(fields[5], 10, 64)
		irq, _ = strconv.ParseUint(fields[6], 10, 64)
		softirq, _ = strconv.ParseUint(fields[7], 10, 64)
		steal, _ = strconv.ParseUint(fields[8], 10, 64)
		break
	}

	if !c.prevCPU.set {
		c.prevCPU.user = user
		c.prevCPU.nice = nice
		c.prevCPU.system = system
		c.prevCPU.idle = idle
		c.prevCPU.iowait = iowait
		c.prevCPU.irq = irq
		c.prevCPU.softirq = softirq
		c.prevCPU.steal = steal
		c.prevCPU.set = true
		return 0
	}

	prevIdle := c.prevCPU.idle + c.prevCPU.iowait
	idle2 := idle + iowait

	prevTotal := c.prevCPU.user + c.prevCPU.nice + c.prevCPU.system + c.prevCPU.idle + c.prevCPU.iowait + c.prevCPU.irq + c.prevCPU.softirq + c.prevCPU.steal
	total := user + nice + system + idle + iowait + irq + softirq + steal

	totalDelta := total - prevTotal
	idleDelta := idle2 - prevIdle

	c.prevCPU.user = user
	c.prevCPU.nice = nice
	c.prevCPU.system = system
	c.prevCPU.idle = idle
	c.prevCPU.iowait = iowait
	c.prevCPU.irq = irq
	c.prevCPU.softirq = softirq
	c.prevCPU.steal = steal

	if totalDelta == 0 {
		return 0
	}

	return 100.0 * float64(totalDelta-idleDelta) / float64(totalDelta)
}

// ramUsage reads /proc/meminfo and returns total, used, and usage percentage.
func (c *Collector) ramUsage() (uint64, uint64, float64) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0, 0
	}
	defer f.Close()

	var total, available uint64
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "MemTotal:"):
			fmt.Sscanf(line, "MemTotal: %d kB", &total)
		case strings.HasPrefix(line, "MemAvailable:"):
			fmt.Sscanf(line, "MemAvailable: %d kB", &available)
		}
	}

	if total == 0 {
		return 0, 0, 0
	}

	used := total - available
	usage := 100.0 * float64(used) / float64(total)

	return total * 1024, used * 1024, usage
}

// diskUsage returns total, free, and usage percentage for the given path.
func (c *Collector) diskUsage(path string) (uint64, uint64, float64) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, 0, 0
	}

	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)

	if total == 0 {
		return 0, 0, 0
	}

	usage := 100.0 * float64(total-free) / float64(total)
	return total, free, usage
}

// loadAvg reads /proc/loadavg.
func (c *Collector) loadAvg() (float64, float64, float64) {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0, 0, 0
	}

	var l1, l5, l15 float64
	fmt.Sscanf(string(data), "%f %f %f", &l1, &l5, &l15)
	return l1, l5, l15
}

// uptime reads /proc/uptime.
func (c *Collector) uptime() float64 {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0
	}

	var up float64
	fmt.Sscanf(string(data), "%f", &up)
	return up
}

// processCount counts processes by listing /proc directories.
func (c *Collector) processCount() int {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return 0
	}

	count := 0
	for _, e := range entries {
		if _, err := strconv.Atoi(e.Name()); err == nil {
			count++
		}
	}
	return count
}
