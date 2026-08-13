package sk

import "fmt"

type ProcessInfo struct {
	Config          *ProcessConfig
	Alive           bool
	MemoryUsage     int64
	MemoryPercent   int32
	CpuPercent      int32
	LastUpdateTime  int64
	LastNoticeTime  int64
	AliveAlertTimes int64
}

func (p *ProcessInfo) String() string {
	var aliveStr = "存活"
	if !p.Alive {
		aliveStr = "丢失"
	}
	str := fmt.Sprintf("[%s] [%s] [CPU使用率=%d%%] [内存使用量=%dMB] [内存使用率=%d%%]", p.Config.Name, aliveStr, p.CpuPercent, p.MemoryUsage, p.MemoryPercent)
	return str
}
