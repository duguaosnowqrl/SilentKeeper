package sk

type AppConfig struct {
	Mail                 *MailConfig      `json:"mail"`
	Ecs                  *EcsConfig       `json:"ecs"`
	Process              []*ProcessConfig `json:"process"`
	TickerInterval       int              `json:"tickerInterval"`
	AlertInterval        int              `json:"alertInterval"`
	AlertEmailSubject    string           `json:"alertEmailSubject"`
	Http                 *HttpConfig      `json:"http"`
	OnlinePercentAlert   int              `json:"onlinePercentAlert"`
	RegisterPercentAlert int              `json:"registerPercentAlert"`
	MailAlertEnabled     bool             `json:"mailAlertEnabled"`
	ConsoleAlertEnabled  bool             `json:"consoleAlertEnabled"`
	AlertLogFile         string           `json:"alertLogFile"`
	FileAlertEnabled     bool             `json:"fileAlertEnabled"`
	UserTimeZone         string           `json:"userTimeZone"`
}

type MailConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Password string `json:"password"`
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
}

type EcsConfig struct {
	Enabled         bool              `json:"enabled"`
	Name            string            `json:"name"`
	IP              string            `json:"ip"`
	IPSecure        bool              `json:"ipSecure"`
	AlertInterval   int               `json:"alertInterval"`
	CpuPercentAlert int32             `json:"cpuPercentAlert"`
	MemPercentAlert int32             `json:"memPercentAlert"`
	Partitions      []PartitionConfig `json:"partitions"`
}

type ProcessConfig struct {
	Enabled         bool   `json:"enabled"`
	Name            string `json:"name"`
	PidPath         string `json:"pidPath"`
	Key             string `json:"key"`
	CpuPercentAlert int32  `json:"cpuPercentAlert"`
	MemMbAlert      int64  `json:"memMbAlert"`
	AlertInterval   int    `json:"alertInterval"`
}

type PartitionConfig struct {
	MountPoint        string `json:"mountPoint"`
	UsagePercentAlert int32  `json:"usagePercentAlert"`
}

type HttpConfig struct {
	Enabled bool   `json:"enabled"`
	Bind    string `json:"bind"`
	Port    int    `json:"port"`
}
