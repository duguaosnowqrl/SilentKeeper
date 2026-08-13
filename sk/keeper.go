package sk

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

const VERSION = "1.0.2"

type Keeper struct {
	config         *AppConfig
	mailSender     *MailSender
	ticker         *time.Ticker
	processInfoArr []*ProcessInfo
	ecsInfo        *EcsInfo
	mailReceivers  []string
	cc             *CommandComponent
	hc             *HttpComponent
	startTime      int64
}

func (k *Keeper) Show() {
	k.cc.show()
}

func (k *Keeper) Start() {
	k.startTime = time.Now().Unix()
	k.showBanner()
	//加载配置
	k.loadConfig()
	//解析邮件接收者
	k.parseMailReceiver()
	//初始化mailSender
	k.initMailSender()
	//创建ecs监控信息
	k.createEcsInfo()
	//创建进程监控信息
	k.createProcessInfo()
	//启动ticker
	go k.startTicker()
	//启动http服务
	go k.startHttpService()
	//显示UI
	k.cc = NewCommandComponent(k)
	k.cc.show()
}

func (k *Keeper) showBanner() {
	fmt.Println("=====================================================")
	fmt.Println("=================== Silent Keeper ===================")
	fmt.Println("=====================================================")
	fmt.Println()
}

func (k *Keeper) loadConfig() {
	data, err := os.ReadFile("config.json")
	if err != nil {
		fmt.Println("加载配置失败:", err)
		panic(err)
	}

	k.config = &AppConfig{}
	err = json.Unmarshal(data, k.config)
	if err != nil {
		fmt.Println("解析配置失败:", err)
		panic(err)
	}
}

func (k *Keeper) parseMailReceiver() {
	parts := strings.Split(k.config.Mail.Receiver, ";")
	k.mailReceivers = parts
}

func (k *Keeper) initMailSender() {
	mailConfig := k.config.Mail
	k.mailSender = NewMailSender(mailConfig.Host, mailConfig.Port, mailConfig.Sender, mailConfig.Password)
}

func (k *Keeper) createProcessInfo() {
	for _, v := range k.config.Process {
		info := &ProcessInfo{
			Config:         v,
			Alive:          false,
			LastUpdateTime: time.Now().UnixMilli(),
		}
		k.processInfoArr = append(k.processInfoArr, info)
	}
}

func (k *Keeper) createEcsInfo() {
	k.ecsInfo = &EcsInfo{
		Config:         k.config.Ecs,
		LastUpdateTime: time.Now().UnixMilli(),
	}

	for _, m := range k.config.Ecs.Partitions {
		partitionInfo := &PartitionInfo{
			MountPoint:        m.MountPoint,
			UsagePercent:      0,
			UsagePercentAlert: m.UsagePercentAlert,
		}
		k.ecsInfo.Partitions = append(k.ecsInfo.Partitions, partitionInfo)
	}
}

func (k *Keeper) updateEcsInfo() {
	cpuPercent, _ := cpu.Percent(time.Second, false)
	k.ecsInfo.CpuPercent = FloatPercentToInt(cpuPercent[0])

	m, _ := mem.VirtualMemory()
	if m != nil {
		k.ecsInfo.MemPercent = FloatPercentToInt(m.UsedPercent)
	}

	for _, p := range k.ecsInfo.Partitions {
		usage, _ := disk.Usage(p.MountPoint)
		if usage != nil {
			p.UsagePercent = FloatPercentToInt(usage.UsedPercent)
		}
	}
}

func (k *Keeper) printEcsInfo() {
	fmt.Println("====================   服务器   ====================")
	fmt.Println(k.ecsInfo.String())
}

func (k *Keeper) startHttpService() {
	if !k.config.Http.Enabled {
		return
	}
	k.hc = NewHttpComponent(k)
	k.hc.Listen()
}

func (k *Keeper) startTicker() {
	sec := k.config.TickerInterval
	if sec <= 0 {
		return
	}

	k.ticker = time.NewTicker(time.Duration(sec) * time.Second)
	defer k.ticker.Stop()
	for {
		select {
		case <-k.ticker.C:
			{
				k.onTick()
				break
			}
		}
	}
}

func (k *Keeper) onTick() {
	k.updateEcsInfo()
	k.updateProcesses()
	k.printEcsInfo()
	fmt.Println()
	k.printProcesses()
	fmt.Println()
	k.checkEcsInfo()
	k.checkProcesses()
}

func (k *Keeper) Alert1(subject string, content string) {
	if k.config.ConsoleAlertEnabled {
		fmt.Println(content)
	}
	if k.config.FileAlertEnabled {
		WriteFile(k.config.AlertLogFile, content, true)
	}
	if k.config.MailAlertEnabled {
		_ = k.mailSender.SendAll(k.mailReceivers, subject, content)
	}
}

func (k *Keeper) Alert2(content string) {
	k.Alert1(k.config.AlertEmailSubject, content)
}

func (k *Keeper) Restart() {
	exe, _ := os.Executable()
	k.release()
	time.Sleep(2 * time.Second)

	cmd := exec.Command(exe)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	_ = cmd.Start()

	os.Exit(0)
}

func (k *Keeper) updateProcesses() {
	for _, info := range k.processInfoArr {
		var proc *process.Process
		pid, _ := ReadPid(info.Config.PidPath)
		proc, _ = GetProcess(pid, info.Config.Key)
		updateProcessInfo(info, proc)
	}
}

func (k *Keeper) printProcesses() {
	fmt.Println("====================   进程   ====================")
	for _, info := range k.processInfoArr {
		fmt.Println(info.String())
	}
}

func (k *Keeper) checkProcesses() {
	for _, info := range k.processInfoArr {
		k.checkAlertForProcess(info)
	}
}

func (k *Keeper) checkEcsInfo() {
	if !k.ecsInfo.Config.Enabled {
		return
	}
	//服务器CPU占用率告警
	now := time.Now()
	timeZone := k.config.UserTimeZone
	if k.ecsInfo.CpuPercent >= k.config.Ecs.CpuPercentAlert {
		msg := fmt.Sprintf("[%s] [警告] [%s] CPU占用率已达%d%%", FormatTimeWithTimeZone(now, timeZone), k.ecsInfo.NameAndIp(), k.ecsInfo.CpuPercent)
		k.Alert2(msg)
	}
	//服务器内存占用率
	if k.ecsInfo.MemPercent >= k.config.Ecs.MemPercentAlert {
		msg := fmt.Sprintf("[%s] [警告] [%s] 内存占用率已达%d%%", FormatTimeWithTimeZone(now, timeZone), k.ecsInfo.NameAndIp(), k.ecsInfo.MemPercent)
		k.Alert2(msg)
	}
	//硬盘占用率
	for _, partition := range k.ecsInfo.Partitions {
		if partition.UsagePercent >= partition.UsagePercentAlert {
			msg := fmt.Sprintf("[%s] [警告] [%s] [%s] 硬盘分区占用率已达%d%%", FormatTimeWithTimeZone(now, timeZone), k.ecsInfo.NameAndIp(), partition.MountPoint, partition.UsagePercent)
			k.Alert2(msg)
		}
	}
}

func (k *Keeper) checkAlertForProcess(info *ProcessInfo) {
	if info == nil || !info.Config.Enabled {
		return
	}
	now := time.Now()
	timeZone := k.config.UserTimeZone
	if !info.Alive {
		//进程丢失告警
		if info.AliveAlertTimes == 0 { //每次检测到进程丢失只警告一次
			msg := fmt.Sprintf("[%s] [严重] [%s] 未检测到进程(若维护中请忽略)", FormatTimeWithTimeZone(now, timeZone), info.Config.Name)
			k.Alert2(msg)
			info.AliveAlertTimes++
		}
	} else {
		//进程内存使用量告警
		if info.MemoryUsage >= info.Config.MemMbAlert {
			msg := fmt.Sprintf("[%s] [警告] [%s] 内存使用量已达%dMB", FormatTimeWithTimeZone(now, timeZone), info.Config.Name, info.MemoryUsage)
			k.Alert2(msg)
		}
		//进程CPU占用率告警
		if info.CpuPercent >= info.Config.CpuPercentAlert {
			msg := fmt.Sprintf("[%s] [警告] [%s] CPU占用率已达%d%%", FormatTimeWithTimeZone(now, timeZone), info.Config.Name, info.CpuPercent)
			k.Alert2(msg)
		}
	}
}

func (k *Keeper) addReceiver(receiver string) {
	for _, v := range k.mailReceivers {
		if v == receiver {
			return
		}
	}
	k.mailReceivers = append(k.mailReceivers, receiver)
}

func (k *Keeper) SelfInfo() string {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	all := memStats.Sys
	//used := memStats.Alloc
	var dhms [4]int
	_ = CalcDayHourMinuteSecond(k.startTime, time.Now().Unix(), &dhms)
	builder := &strings.Builder{}
	_, _ = fmt.Fprintf(builder, "版本：%s\n", VERSION)
	_, _ = fmt.Fprintf(builder, "内存：%dKB\n", BToKB(all))
	_, _ = fmt.Fprintf(builder, "Heap：%dKB/%dKB\n", BToKB(memStats.HeapInuse), BToKB(memStats.HeapSys))
	_, _ = fmt.Fprintf(builder, "GC次数：%d\n", memStats.NumGC)
	_, _ = fmt.Fprintf(builder, "运行时长：%d天%d时%d分%d秒", dhms[0], dhms[1], dhms[2], dhms[3])
	return builder.String()
}

func (k *Keeper) exit() {
	fmt.Println("拜拜~")
	k.release()
	os.Exit(0)
}

func (k *Keeper) release() {
	if k.ticker != nil {
		k.ticker.Stop()
		k.ticker = nil
	}
	if k.hc != nil {
		k.hc.Close()
	}
}

func updateProcessInfo(info *ProcessInfo, proc *process.Process) {
	if info == nil {
		return
	}
	if proc != nil { //找对对应的进程，刷新进程数据
		//标记为存活
		info.Alive = true
		//清理存活警告次数
		info.AliveAlertTimes = 0
		//更新内存信息
		memValue, memPercent, _ := GetProcessMemoryInfo(proc)
		info.MemoryUsage = memValue
		info.MemoryPercent = memPercent
		//更新CPU信息
		cpuPercent, _ := GetProcessCpuInfo(proc)
		info.CpuPercent = cpuPercent
		//更新上次检查时间
		info.LastUpdateTime = time.Now().UnixMilli()
	} else {
		//标记为死亡
		info.Alive = false
		//更新内存信息
		info.MemoryUsage = 0
		info.MemoryPercent = 0
		//更新CPU信息
		info.CpuPercent = 0
		//更新上次检查时间
		info.LastUpdateTime = time.Now().UnixMilli()
	}
}
