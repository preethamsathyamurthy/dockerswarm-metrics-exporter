package core

import (
	"fmt"
	"time"

	"strconv"
	"syscall"

	"github.com/docker/docker/api/types/swarm"
	humanize "github.com/dustin/go-humanize"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

type cpuMetrics struct {
	CpuInfoStats             []cpu.InfoStat
	CpuPercentageUtilization []float64
	CpuLoadAverageStat       *load.AvgStat
}

type diskInfo struct {
	Partition disk.PartitionStat
	DiskUsage *disk.UsageStat
}

type diskMetrics struct {
	DiskInfo []diskInfo
	DiskIO   map[string]disk.IOCountersStat
}

type systemMetrics struct {
	CpuMetrics      cpuMetrics
	MemoryInfoStats *mem.VirtualMemoryStat
	DiskMetrics     diskMetrics
	HostInfo        *host.InfoStat
	NetInfo         []net.IOCountersStat
}

type dockerMetrics struct {
	RootDiskInfo  DockerRootInfo
	Status        bool
	IsManager     bool
	ManagerStatus *swarm.ManagerStatus
	NodeList      []swarm.Node
}

type metrics struct {
	DockerMetrics dockerMetrics
	SystemMetrics systemMetrics
}

// File System Info
type DockerRootInfo struct {
	Path               string
	Total              uint64
	Free               uint64
	Used               uint64
	Files              uint64
	Ffree              uint64
	FSType             string
	PercentageUtilized float64
}

// File System Type
var fsType2StringMap = map[string]string{
	"1021994":  "TMPFS",
	"137d":     "EXT",
	"4244":     "HFS",
	"4d44":     "MSDOS",
	"52654973": "REISERFS",
	"5346544e": "NTFS",
	"58465342": "XFS",
	"61756673": "AUFS",
	"6969":     "NFS",
	"ef51":     "EXT2OLD",
	"ef53":     "EXT4",
	"f15f":     "ecryptfs",
	"794c7630": "overlayfs",
	"2fc12fc1": "zfs",
	"ff534d42": "cifs",
	"53464846": "wslfs",
}

// Function to get Filesystem Type
func GetFSType(ftype int64) string {
	fsTypeHex := strconv.FormatInt(ftype, 16)
	fsTypeString, ok := fsType2StringMap[fsTypeHex]
	if !ok {
		return "UNKNOWN"
	}
	return fsTypeString
}

// Get Disk Info
func GetPathDiskInfo(path string) (info DockerRootInfo, err error) {
	s := syscall.Statfs_t{}
	err = syscall.Statfs(path, &s)
	if err != nil {
		// err = errors.New(fmt.Sprint(err) + "||" + trace2())
		return DockerRootInfo{}, err
	}
	reservedBlocks := s.Bfree - s.Bavail

	info = DockerRootInfo{
		Path:  path,
		Total: uint64(s.Ffree) * (s.Blocks - reservedBlocks),
		Free:  uint64(s.Ffree) * s.Bavail,
		Files: s.Files,
		Ffree: s.Ffree,
		//nolint:unconvert
		FSType: GetFSType(int64(s.Type)),
		//Used:               (uint64(s.Ffree) * (s.Blocks - reservedBlocks)) - (uint64(s.Ffree) * s.Bavail),
		//PercentageUtilized: float64((((uint64(s.Ffree) * (s.Blocks - reservedBlocks)) - (uint64(s.Ffree) * s.Bavail)) / (uint64(s.Ffree) * (s.Blocks - reservedBlocks))) * 100),
		// Percentage: (float64((uint64(s.Ffree)*(s.Blocks-reservedBlocks))-(uint64(s.Ffree)*s.Bavail)) / float64(uint64(s.Ffree)*(s.Blocks-reservedBlocks))) * 100,
	}
	// Check for overflows.
	// https://github.com/minio/minio/issues/8035
	// XFS can show wrong values at times error out
	// in such scenarios.
	if info.Free > info.Total {
		return info, fmt.Errorf("detected free space (%d) > total disk space (%d), fs corruption at (%s). please run 'fsck'", info.Free, info.Total, path)
	}
	info.Used = info.Total - info.Free
	info.PercentageUtilized = ((float64(info.Used) / float64(info.Total)) * 100)

	return info, nil
}

// Get Disk usage
func PrintDiskUsage(path string) string {
	di, err := GetPathDiskInfo(path)
	if err != nil {
		panic(err)
	}

	return (fmt.Sprintf("%s of %s disk space used (%0.2f%%)\n",
		humanize.Bytes(di.Used),
		humanize.Bytes(di.Total),
		di.PercentageUtilized,
	))
}

// Errors could be handled here rather than passing
// if we anyways pass error variable
func GetCPUInfo() []cpu.InfoStat {
	cpuInfos, err := cpu.Info()
	if err != nil {
		writeLogs.Error(err.Error())
		return []cpu.InfoStat{}
	}
	return cpuInfos
}

func GetCPUPercentageUtilization() []float64 {
	//CPU utilization
	percent, err := cpu.Percent(time.Second, false)
	if err != nil {
		writeLogs.Error(err.Error())
		return []float64{}
	}
	return percent
}

// CPU Load
func GetCpuLoad() *load.AvgStat {
	info, err := load.Avg()
	if err != nil {
		writeLogs.Error(err.Error())
		return nil
	}
	return info
}

// mem info
func GetMemInfo() *mem.VirtualMemoryStat {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		writeLogs.Error(err.Error())
		return nil
	}
	return memInfo
}

// disk info
func GetDiskInfo() []diskInfo {
	parts, err := disk.Partitions(true)
	if err != nil {
		writeLogs.Error(err.Error())
		return nil
	}
	diskInfos := []diskInfo{}
	for _, part := range parts {
		diskInformation, _ := disk.Usage(part.Mountpoint)
		diskInfos = append(diskInfos, diskInfo{
			Partition: part,
			DiskUsage: diskInformation,
		})
	}

	return diskInfos

}

func GetDiskIOCounterStat() map[string]disk.IOCountersStat {
	ioStat, err := disk.IOCounters()
	if err != nil {
		writeLogs.Error(err.Error())
		return map[string]disk.IOCountersStat{}
	}
	return ioStat
}

// host info
func GetHostInfo() *host.InfoStat {
	hInfo, _ := host.Info()
	return hInfo
}

func GetNetInfo() []net.IOCountersStat {
	info, _ := net.IOCounters(true)
	return info
}

func GetCurrentInfo() metrics {

	dockerStatus := true

	//returning error method
	// systemFunctions -> errors are handled in function
	// Here we are returning all errors
	dockerRootDiskInfo, errorDockerRoot := DockerRootInfo{}, error(nil)
	isManager, errorNodeManager := false, error(nil)
	var managerStatus *swarm.ManagerStatus
	nodeList, nodeListError := []swarm.Node{}, error(nil)
	dockerRootPath, err := GetDockerRootInfo()

	if err != nil {
		writeLogs.Error(err.Error())
		dockerStatus = false
		dockerRootDiskInfo = DockerRootInfo{}

	} else {
		dockerRootDiskInfo, errorDockerRoot = GetPathDiskInfo(dockerRootPath)

		if errorDockerRoot != nil {
			dockerRootDiskInfo.Path = dockerRootPath
			writeLogs.Error(errorDockerRoot.Error())
		}

		isManager, managerStatus, errorNodeManager = isANodeManager()

		if errorNodeManager != nil {
			writeLogs.Error(errorNodeManager.Error())
		}
		if isManager {
			nodeList, nodeListError = GetNodeList(true)
			if nodeListError != nil {
				writeLogs.Error(nodeListError.Error())
			}
		}
	}

	return (metrics{
		SystemMetrics: systemMetrics{
			CpuMetrics: cpuMetrics{
				CpuInfoStats:             GetCPUInfo(),
				CpuPercentageUtilization: GetCPUPercentageUtilization(),
				CpuLoadAverageStat:       GetCpuLoad(),
			},
			MemoryInfoStats: GetMemInfo(),
			DiskMetrics: diskMetrics{
				DiskInfo: GetDiskInfo(),
				DiskIO:   GetDiskIOCounterStat(),
			},
			HostInfo: GetHostInfo(),
		},
		DockerMetrics: dockerMetrics{
			RootDiskInfo:  dockerRootDiskInfo,
			Status:        dockerStatus,
			IsManager:     isManager,
			ManagerStatus: managerStatus,
			NodeList:      nodeList,
		},
	})
}
