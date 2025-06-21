package monitor

import (
	"context"
	"fmt"
	"net"
	"os"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	psnet "github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// SystemStats represents system resource statistics
type SystemStats struct {
	CPU     CPUStats     `json:"cpu"`
	Memory  MemoryStats  `json:"memory"`
	Disk    DiskStats    `json:"disk"`
	Network NetworkStats `json:"network"`
	Runtime RuntimeStats `json:"runtime"`
}

// CPUStats represents CPU usage statistics
type CPUStats struct {
	Usage       float64   `json:"usage"`         // Overall CPU usage percentage
	UsagePerCPU []float64 `json:"usage_per_cpu"` // Usage per CPU core
	Cores       int       `json:"cores"`         // Number of CPU cores
	ModelName   string    `json:"model_name"`    // CPU model name
}

// MemoryStats represents memory usage statistics
type MemoryStats struct {
	Total       uint64  `json:"total"`        // Total memory in bytes
	Available   uint64  `json:"available"`    // Available memory in bytes
	Used        uint64  `json:"used"`         // Used memory in bytes
	UsedPercent float64 `json:"used_percent"` // Used memory percentage
	Free        uint64  `json:"free"`         // Free memory in bytes
	Buffers     uint64  `json:"buffers"`      // Buffer memory in bytes
	Cached      uint64  `json:"cached"`       // Cached memory in bytes
}

// DiskStats represents disk usage statistics
type DiskStats struct {
	Total       uint64  `json:"total"`        // Total disk space in bytes
	Used        uint64  `json:"used"`         // Used disk space in bytes
	Free        uint64  `json:"free"`         // Free disk space in bytes
	UsedPercent float64 `json:"used_percent"` // Used disk space percentage
	Path        string  `json:"path"`         // Disk path
}

// NetworkStats represents network statistics
type NetworkStats struct {
	BytesSent   uint64             `json:"bytes_sent"`   // Total bytes sent
	BytesRecv   uint64             `json:"bytes_recv"`   // Total bytes received
	PacketsSent uint64             `json:"packets_sent"` // Total packets sent
	PacketsRecv uint64             `json:"packets_recv"` // Total packets received
	Interfaces  []NetworkInterface `json:"interfaces"`   // Network interfaces
}

// NetworkInterface represents a network interface
type NetworkInterface struct {
	Name      string `json:"name"`       // Interface name
	BytesSent uint64 `json:"bytes_sent"` // Bytes sent on this interface
	BytesRecv uint64 `json:"bytes_recv"` // Bytes received on this interface
	IsUp      bool   `json:"is_up"`      // Interface status
}

// RuntimeStats represents Go runtime statistics
type RuntimeStats struct {
	Goroutines    int     `json:"goroutines"`      // Number of goroutines
	HeapAlloc     uint64  `json:"heap_alloc"`      // Heap allocated memory
	HeapSys       uint64  `json:"heap_sys"`        // Heap system memory
	HeapIdle      uint64  `json:"heap_idle"`       // Heap idle memory
	HeapInuse     uint64  `json:"heap_inuse"`      // Heap in-use memory
	HeapReleased  uint64  `json:"heap_released"`   // Heap released memory
	GCCPUFraction float64 `json:"gc_cpu_fraction"` // GC CPU fraction
	NextGC        uint64  `json:"next_gc"`         // Next GC threshold
	NumGC         uint32  `json:"num_gc"`          // Number of GC runs
}

// SystemMonitor provides system monitoring capabilities
type SystemMonitor struct {
	ctx context.Context
}

// NewSystemMonitor creates a new system monitor
func NewSystemMonitor(ctx context.Context) *SystemMonitor {
	return &SystemMonitor{
		ctx: ctx,
	}
}

// GetSystemStats retrieves current system statistics
func (sm *SystemMonitor) GetSystemStats() (*SystemStats, error) {
	stats := &SystemStats{}

	// Get CPU stats
	cpuStats, err := sm.getCPUStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU stats: %w", err)
	}
	stats.CPU = *cpuStats

	// Get Memory stats
	memStats, err := sm.getMemoryStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory stats: %w", err)
	}
	stats.Memory = *memStats

	// Get Disk stats
	diskStats, err := sm.getDiskStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get disk stats: %w", err)
	}
	stats.Disk = *diskStats

	// Get Network stats
	netStats, err := sm.getNetworkStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get network stats: %w", err)
	}
	stats.Network = *netStats

	// Get Runtime stats
	runtimeStats := sm.getRuntimeStats()
	stats.Runtime = *runtimeStats

	return stats, nil
}

// getCPUStats retrieves CPU statistics
func (sm *SystemMonitor) getCPUStats() (*CPUStats, error) {
	// Get CPU usage percentage using cached values for immediate response
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		return nil, err
	}

	// Get per-CPU usage using cached values for immediate response
	cpuPercentPerCPU, err := cpu.Percent(0, true)
	if err != nil {
		return nil, err
	}

	// Get CPU info
	cpuInfo, err := cpu.Info()
	if err != nil {
		return nil, err
	}

	modelName := "Unknown"
	if len(cpuInfo) > 0 {
		modelName = cpuInfo[0].ModelName
	}

	return &CPUStats{
		Usage:       cpuPercent[0],
		UsagePerCPU: cpuPercentPerCPU,
		Cores:       runtime.NumCPU(),
		ModelName:   modelName,
	}, nil
}

// getMemoryStats retrieves memory statistics
func (sm *SystemMonitor) getMemoryStats() (*MemoryStats, error) {
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	return &MemoryStats{
		Total:       vmStat.Total,
		Available:   vmStat.Available,
		Used:        vmStat.Used,
		UsedPercent: vmStat.UsedPercent,
		Free:        vmStat.Free,
		Buffers:     vmStat.Buffers,
		Cached:      vmStat.Cached,
	}, nil
}

// getDiskStats retrieves disk statistics
func (sm *SystemMonitor) getDiskStats() (*DiskStats, error) {
	diskStat, err := disk.Usage("/")
	if err != nil {
		return nil, err
	}

	return &DiskStats{
		Total:       diskStat.Total,
		Used:        diskStat.Used,
		Free:        diskStat.Free,
		UsedPercent: diskStat.UsedPercent,
		Path:        diskStat.Path,
	}, nil
}

// getNetworkStats retrieves network statistics
func (sm *SystemMonitor) getNetworkStats() (*NetworkStats, error) {
	netIO, err := psnet.IOCounters(false)
	if err != nil {
		return nil, err
	}

	netInterfaces, err := psnet.IOCounters(true)
	if err != nil {
		return nil, err
	}

	// Get total network stats
	var totalBytesSent, totalBytesRecv, totalPacketsSent, totalPacketsRecv uint64
	if len(netIO) > 0 {
		totalBytesSent = netIO[0].BytesSent
		totalBytesRecv = netIO[0].BytesRecv
		totalPacketsSent = netIO[0].PacketsSent
		totalPacketsRecv = netIO[0].PacketsRecv
	}

	// Get interface-specific stats
	interfaces := make([]NetworkInterface, 0, len(netInterfaces))
	for _, iFace := range netInterfaces {
		// Check if interface is up
		isUp := sm.isInterfaceUp(iFace.Name)

		interfaces = append(interfaces, NetworkInterface{
			Name:      iFace.Name,
			BytesSent: iFace.BytesSent,
			BytesRecv: iFace.BytesRecv,
			IsUp:      isUp,
		})
	}

	return &NetworkStats{
		BytesSent:   totalBytesSent,
		BytesRecv:   totalBytesRecv,
		PacketsSent: totalPacketsSent,
		PacketsRecv: totalPacketsRecv,
		Interfaces:  interfaces,
	}, nil
}

// isInterfaceUp checks if a network interface is up
func (sm *SystemMonitor) isInterfaceUp(name string) bool {
	iFace, err := net.InterfaceByName(name)
	if err != nil {
		return false
	}
	return iFace.Flags&net.FlagUp == net.FlagUp
}

// getRuntimeStats retrieves Go runtime statistics
func (sm *SystemMonitor) getRuntimeStats() *RuntimeStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &RuntimeStats{
		Goroutines:    runtime.NumGoroutine(),
		HeapAlloc:     m.HeapAlloc,
		HeapSys:       m.HeapSys,
		HeapIdle:      m.HeapIdle,
		HeapInuse:     m.HeapInuse,
		HeapReleased:  m.HeapReleased,
		GCCPUFraction: m.GCCPUFraction,
		NextGC:        m.NextGC,
		NumGC:         m.NumGC,
	}
}

// GetProcessStats retrieves statistics for the current process
func (sm *SystemMonitor) GetProcessStats() (*ProcessStats, error) {
	pid := int32(os.Getpid())
	proc, err := process.NewProcess(pid)
	if err != nil {
		return nil, err
	}

	cpuPercent, err := proc.CPUPercent()
	if err != nil {
		return nil, err
	}

	memInfo, err := proc.MemoryInfo()
	if err != nil {
		return nil, err
	}

	createTime, err := proc.CreateTime()
	if err != nil {
		return nil, err
	}

	numThreads, err := proc.NumThreads()
	if err != nil {
		return nil, err
	}

	return &ProcessStats{
		PID:        int(pid),
		CPUPercent: cpuPercent,
		MemoryRSS:  memInfo.RSS,
		MemoryVMS:  memInfo.VMS,
		NumThreads: numThreads,
		CreateTime: time.Unix(createTime/1000, 0),
		Uptime:     time.Since(time.Unix(createTime/1000, 0)),
	}, nil
}

// ProcessStats represents process-specific statistics
type ProcessStats struct {
	PID        int           `json:"pid"`         // Process ID
	CPUPercent float64       `json:"cpu_percent"` // CPU usage percentage
	MemoryRSS  uint64        `json:"memory_rss"`  // Resident set size
	MemoryVMS  uint64        `json:"memory_vms"`  // Virtual memory size
	NumThreads int32         `json:"num_threads"` // Number of threads
	CreateTime time.Time     `json:"create_time"` // Process creation time
	Uptime     time.Duration `json:"uptime"`      // Process uptime
}
