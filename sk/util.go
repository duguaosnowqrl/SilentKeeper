package sk

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

const TIME_LAYOUT = "2006-01-02 15:04:05"

func BToMB(b uint64) int64 {
	return int64(b / 1024 / 1024)
}

func BToKB(b uint64) uint64 {
	return b / 1024
}

func FormatTime(t time.Time) string {
	return t.Format(TIME_LAYOUT)
}

func FormattedNow() string {
	return time.Now().Format(TIME_LAYOUT)
}

func FormatTimeWithTimeZone(t time.Time, zoneName string) string {
	tz, err := time.LoadLocation(zoneName)
	if err != nil {
		tz = time.Local
	}
	t = t.In(tz)
	_, offset := t.Zone()
	offset = offset / 3600
	if offset < 0 {
		return fmt.Sprintf("%s(UTC%s)", t.Format(TIME_LAYOUT), strconv.Itoa(offset))
	}
	return fmt.Sprintf("%s(UTC+%s)", t.Format(TIME_LAYOUT), strconv.Itoa(offset))
}

func FloatPercentToInt(percent float64) int32 {
	percent = math.Ceil(percent)
	return int32(percent)
}

func ReadPid(pidPath string) (int32, error) {
	data, err := os.ReadFile(pidPath)
	if err != nil {
		return 0, err
	}

	str := strings.TrimSpace(string(data))
	pid, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(pid), nil
}

func GetProcess(pid int32, key string) (*process.Process, error) {
	if pid <= 0 {
		return nil, fmt.Errorf("pid不合法:%d", pid)
	}
	p, err := process.NewProcess(pid)
	if err != nil {
		return nil, err
	}
	cmdLine, err := p.Cmdline()
	if err != nil {
		return nil, err
	}
	if !strings.Contains(cmdLine, key) {
		return nil, fmt.Errorf("启动命令不包含关键字:%s", key)
	}
	return p, nil
}

func GetProcessMemoryInfo(proc *process.Process) (int64, int32, error) {
	var percent float64
	var value int64

	memInfo, err := proc.MemoryInfo()
	if err != nil {
		return 0, 0, err
	}

	memPercent, err := proc.MemoryPercent()
	if err != nil {
		return 0, 0, err
	}

	value = BToMB(memInfo.RSS)
	percent = math.Ceil(float64(memPercent))

	return value, int32(percent), nil
}

func GetProcessCpuInfo(proc *process.Process) (int32, error) {
	var percent float64
	p, err := proc.CPUPercent()
	if err != nil {
		return 0, err
	}
	cpuNum := runtime.NumCPU()
	percent = math.Ceil(p / float64(cpuNum))
	return int32(percent), nil
}

func CalcPercent(a int, b int) int {
	f := math.Ceil((float64(a) / float64(b)) * 100)
	return int(f)
}

func CalcDayHourMinuteSecond(beginSec int64, endSec int64, arr *[4]int) error {
	sec := endSec - beginSec
	if sec < 0 {
		return fmt.Errorf("CalcDayHourMinuteSecond: endSec < beginSec")
	}
	arr[0] = int(sec / 86400)
	sec %= 86400
	arr[1] = int(sec / 3600)
	sec %= 3600
	arr[2] = int(sec / 60)
	arr[3] = int(sec % 60)
	return nil
}

func WriteFile(path string, msg string, lineBreak bool) {
	if path == "" || msg == "" {
		return
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return
	}
	defer file.Close()
	_, _ = file.WriteString(msg)
	if lineBreak {
		_, _ = file.WriteString("\n")
	}
}
