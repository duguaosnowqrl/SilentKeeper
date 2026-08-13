package sk

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type CommandComponent struct {
	app      *Keeper
	commands map[string]func([]string)
}

func NewCommandComponent(app *Keeper) *CommandComponent {
	cc := &CommandComponent{
		app: app,
	}
	cc.registerCommands()
	return cc
}

func (cc *CommandComponent) registerCommands() {
	cc.commands = make(map[string]func([]string))
	cc.commands["help"] = cc.handleHelp
	cc.commands["receiver.add"] = cc.handleAddReceivers
	cc.commands["receiver.list"] = cc.handleListReceivers
	cc.commands["ecs.list"] = cc.handlePrintEcsInfo
	cc.commands["process.list"] = cc.handlePrintProcesses
	cc.commands["self"] = cc.handleSelfInfo
	cc.commands["tick"] = cc.handleTick
	cc.commands["mail"] = cc.handleMail
	cc.commands["alert"] = cc.handleAlert
	cc.commands["exit"] = cc.handleExit
}

func (cc *CommandComponent) handleHelp(args []string) {
	fmt.Println("支持以下命令:")
	fmt.Println("help 查看帮助")
	fmt.Println("receiver.list 查看收件人")
	fmt.Println("receiver.add <收件人地址(多个)> 添加收件人")
	fmt.Println("ecs.list 查看服务器")
	fmt.Println("process.list 查看进程")
	fmt.Println("self 查看程序运行状态")
	fmt.Println("tick 立刻触发一次tick")
	fmt.Println("alert <内容> 发起警告")
	fmt.Println("mail <收件人地址> <主题> <内容> 发送邮件")
	fmt.Println("exit 退出")
}

func (cc *CommandComponent) handleListReceivers(args []string) {
	for _, receiver := range cc.app.mailReceivers {
		fmt.Println(receiver)
	}
}

func (cc *CommandComponent) handleAddReceivers(args []string) {
	for _, receiver := range args {
		cc.app.addReceiver(receiver)
	}
}

func (cc *CommandComponent) handleMail(args []string) {
	receiver := args[0]
	subject := args[1]
	content := args[2]
	_ = cc.app.mailSender.Send(receiver, subject, content)
}

func (cc *CommandComponent) handleAlert(args []string) {
	cc.app.Alert2(args[0])
}

func (cc *CommandComponent) handleExit(args []string) {
	cc.app.exit()
}

func (cc *CommandComponent) handlePrintEcsInfo(args []string) {
	cc.app.updateEcsInfo()
	fmt.Println(cc.app.ecsInfo.String())
}

func (cc *CommandComponent) handlePrintProcesses(args []string) {
	cc.app.updateProcesses()
	cc.app.printProcesses()
}

func (cc *CommandComponent) handleSelfInfo(args []string) {
	fmt.Println(cc.app.SelfInfo())
}

func (cc *CommandComponent) handleTick(args []string) {
	cc.app.onTick()
}

func (cc *CommandComponent) show() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print(">")
		if !scanner.Scan() { //nohup启动时，会导致stdin立刻返回eof,导致cpu狂飙。后台运行时应当阻塞住
			fmt.Println("Stdin EOF!")
			select {}
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		parts := strings.Fields(input)
		cmd := parts[0]
		args := parts[1:]
		fn, ok := cc.commands[cmd]
		if !ok {
			fmt.Printf("不支持的命令 %s\n", cmd)
			continue
		}
		executeCommand(fn, cmd, args)
	}
}

func executeCommand(fn func([]string), cmd string, args []string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("执行命令%s出错:%v\n", cmd, r)
			fmt.Println("请使用help命令获取帮助")
		}
	}()
	fn(args)
}
