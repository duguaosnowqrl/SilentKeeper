package sk

import (
	"fmt"
	"strings"
)

type EcsInfo struct {
	Config         *EcsConfig
	LastUpdateTime int64
	CpuPercent     int32
	MemPercent     int32
	Partitions     []*PartitionInfo
}

type PartitionInfo struct {
	MountPoint        string
	UsagePercent      int32
	UsagePercentAlert int32
}

func (e *EcsInfo) String() string {
	builder := new(strings.Builder)
	base := fmt.Sprintf("[%s] [CPU使用率=%d%%] [内存使用率=%d%%]", e.NameAndIp(), e.CpuPercent, e.MemPercent)
	builder.WriteString(base)
	for _, partition := range e.Partitions {
		builder.WriteString("\n")
		disk := fmt.Sprintf("- [%s] [使用率=%d%%]", partition.MountPoint, partition.UsagePercent)
		builder.WriteString(disk)
	}
	return builder.String()
}

func (e *EcsInfo) NameAndIp() string {
	return e.Config.Name + "-" + e.ProtectedIP()
}

func (e *EcsInfo) ProtectedIP() string {
	if !e.Config.IPSecure {
		return e.Config.IP
	}
	arr := strings.Split(e.Config.IP, ".")
	if len(arr) != 4 {
		return e.Config.IP
	}
	arr[1] = "*"
	arr[2] = "*"
	return strings.Join(arr, ".")
}
