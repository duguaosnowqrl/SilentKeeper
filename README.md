# Silent Keeper

`Silent Keeper` 是一个使用 Go 编写的轻量级服务巡检与告警工具，用来监控服务器资源、指定进程状态，以及业务侧上报的人数指标，并通过控制台、文件、邮件等方式发送告警。

项目当前启动后会加载 `config.json`，初始化监控对象，按配置周期执行巡检，并可选地启动一个本地 HTTP 服务用于查看状态和触发动作。

## 主要功能

- 监控服务器 CPU、内存、磁盘分区使用率
- 监控指定进程是否存活，以及 CPU、内存占用情况
- 支持业务系统通过 HTTP 上报区服在线人数、注册人数并触发阈值告警
- 支持三种告警输出方式
  - 控制台输出
  - 文件日志输出
  - SMTP邮件发送
  - 后续将提供短信、QQ、微信等告警方式
- 提供简单的控制台命令行交互
- 提供本地 HTTP 接口查看状态、手动触发巡检、重启或关闭程序

## 项目结构

```text
.
|-- main.go                  程序入口
|-- config.template.json     配置模板
|-- config.json              实际运行配置
|-- Makefile                 构建脚本
`-- sk/
    |-- keeper.go            核心调度逻辑
    |-- config.go            配置结构定义
    |-- http.go              HTTP 接口
    |-- command.go           控制台命令
    |-- ecs.go               服务器资源信息
    |-- process.go           进程信息
    |-- mail.go              邮件发送
    `-- util.go              工具函数
```

## 运行环境

- Go `1.22` 或更高版本
- 支持 Windows / Linux
- 需要可访问目标 SMTP 服务，若启用邮件告警

## 快速开始

### 1. 准备配置

先复制配置模板：

```bash
cp config.template.json config.json
```

Windows PowerShell 可使用：

```powershell
Copy-Item config.template.json config.json
```

然后根据实际环境修改 `config.json`。

### 2. 直接运行

```bash
go run main.go
```

### 3. 编译

直接编译当前平台：

```bash
go build -o sk main.go
```

如果本机安装了 `make`，也可以使用仓库内的构建脚本：

```bash
make init
make build
```

其中：

- `make init`：从模板生成根目录 `config.json`,可用于初始化工程
- `make build`：清理并输出 Linux、Windows 64 位、Windows 32 位构建产物到 `out/`

另外：

- `make build-linux`：单独输出 Linux64位构建产物到 `out/`
- `make build-win32`：单独输出 Windows32位构建产物到 `out/`
- `make build-win64`：单独输出 Windows64位构建产物到 `out/`
- `make copy-config`：复制config.template.json到 `out/`

## 配置说明

配置文件示例见 `config.template.json`。

### 顶层配置

| 字段 | 说明 |
| --- | --- |
| `mail` | 邮件服务器配置 |
| `ecs` | 服务器资源监控配置 |
| `process` | 进程监控列表 |
| `http` | HTTP 服务配置 |
| `alertLogFile` | 告警日志文件路径 |
| `fileAlertEnabled` | 是否写入告警日志 |
| `mailAlertEnabled` | 是否发送邮件告警 |
| `consoleAlertEnabled` | 是否在控制台输出告警 |
| `tickerInterval` | 自动巡检周期，单位秒；`<= 0` 表示不自动轮询 |
| `alertInterval` | 预留字段，当前代码中未实际使用 |
| `alertEmailSubject` | 邮件告警标题 |
| `onlinePercentAlert` | 在线人数告警阈值，百分比 |
| `registerPercentAlert` | 注册人数告警阈值，百分比 |
| `userTimeZone` | 告警时间使用的时区，例如 `Asia/Shanghai` |

### `mail`

| 字段 | 说明 |
| --- | --- |
| `host` | SMTP 服务器地址 |
| `port` | SMTP 端口 |
| `sender` | 发件人邮箱 |
| `password` | 发件人密码或授权码 |
| `receiver` | 收件人列表，使用 `;` 分隔多个地址 |

### `ecs`

| 字段 | 说明 |
| --- | --- |
| `enabled` | 是否启用服务器资源监控 |
| `name` | 服务器名称 |
| `ip` | 服务器 IP |
| `ipSecure` | 是否对 IP 做脱敏展示 |
| `cpuPercentAlert` | CPU 使用率告警阈值 |
| `memPercentAlert` | 内存使用率告警阈值 |
| `partitions` | 需要监控的磁盘分区列表 |

### `process`

每个进程项包含以下字段：

| 字段 | 说明 |
| --- | --- |
| `enabled` | 是否启用该进程监控 |
| `name` | 进程名称，用于展示和告警 |
| `pidPath` | PID 文件路径 |
| `key` | 进程启动命令关键字，用于二次校验 |
| `cpuPercentAlert` | CPU 使用率告警阈值 |
| `memMbAlert` | 内存使用量告警阈值，单位 MB |

### `http`

| 字段 | 说明 |
| --- | --- |
| `enabled` | 是否启用 HTTP 服务 |
| `bind` | 监听地址 |
| `port` | 监听端口 |

## 控制台命令

程序启动后可在标准输入中执行以下命令：

| 命令 | 说明 |
| --- | --- |
| `help` | 查看帮助 |
| `receiver.list` | 查看当前收件人 |
| `receiver.add <邮箱1> <邮箱2>` | 追加收件人 |
| `ecs.list` | 查看服务器资源状态 |
| `process.list` | 查看进程状态 |
| `self` | 查看程序自身运行信息 |
| `tick` | 立即执行一次巡检 |
| `alert <内容>` | 主动发送一条告警 |
| `mail <收件人> <主题> <内容>` | 发送测试邮件 |
| `exit` | 退出程序 |

说明：

- 后台运行且没有标准输入时，程序会阻塞等待，而不是持续空转占用 CPU

## HTTP 接口

默认监听地址由 `http.bind` 和 `http.port` 决定，例如 `http://127.0.0.1:8895`。

### 状态与控制

| 路径 | 方法 | 说明 |
| --- | --- | --- |
| `/` | GET | 查看程序信息、服务器状态、进程状态 |
| `/ecs/list` | GET | 查看服务器资源状态 |
| `/process/list` | GET | 查看进程状态 |
| `/receiver/list` | GET | 查看收件人列表 |
| `/tick` | GET | 手动触发一次巡检 |
| `/restart` | GET | 重启程序 |
| `/stop` | GET | 停止程序 |
| `/alert` | GET / POST | 主动发送告警 |

`/alert` 请求示例：

```bash
curl "http://127.0.0.1:8895/alert?msg=test-alert"
```

或：

```bash
curl -X POST "http://127.0.0.1:8895/alert" \
  -H "Content-Type: application/json" \
  -d "{\"msg\":\"test-alert\"}"
```

### 区服上报接口

路径：

```text
/district/report
```

仅支持 `POST` JSON，程序会读取每个区服的以下字段：

- `name`
- `online_num`
- `online_limit`
- `register_num`
- `register_limit`

`POST` JSON 示例：

```json
{
  "1": {
    "name": "1服",
    "online_num": 900,
    "online_limit": 1000,
    "register_num": 950,
    "register_limit": 1000
  },
  "2": {
    "name": "2服",
    "online_num": 300,
    "online_limit": 1000,
    "register_num": 400,
    "register_limit": 1000
  }
}
```

示例请求：

```bash
curl -X POST "http://127.0.0.1:8895/district/report" \
  -H "Content-Type: application/json" \
  -d "{\"1\":{\"name\":\"1服\",\"online_num\":900,\"online_limit\":1000,\"register_num\":950,\"register_limit\":1000}}"
```

## 日志与输出

- 普通状态信息默认输出到控制台
- 告警信息可按配置同时输出到控制台、文件和邮件
- 告警日志路径由 `alertLogFile` 控制

## 注意事项

- 程序运行目录下必须存在 `config.json`
- `process.pidPath` 必须指向一个内容为数字 PID 的文本文件
- `process.key` 用于校验 PID 对应进程是否为目标进程，建议填写有辨识度的命令关键字
- 邮件发送当前使用 TLS 直连 SMTP，需确保服务器与端口匹配
- 若 `tickerInterval <= 0`，程序不会自动巡检，但仍可通过控制台 `tick` 或 HTTP `/tick` 手动触发

## 版本

当前代码中的版本号为 `1.0.2`。

## 开源协议

本项目使用 `MIT License` 开源，详见根目录的 `LICENSE` 文件。
第三方依赖的协议声明见根目录 `THIRD_PARTY_NOTICES.md`。
