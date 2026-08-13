package sk

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type HttpComponent struct {
	app     *Keeper
	server  *http.Server
	closing bool
}

func NewHttpComponent(app *Keeper) *HttpComponent {
	return &HttpComponent{app: app}
}

func (c *HttpComponent) Listen() {
	mux := http.NewServeMux()
	c.registerHandlers(mux)
	httpConfig := c.app.config.Http
	c.server = &http.Server{
		Addr:    httpConfig.Bind + ":" + strconv.Itoa(httpConfig.Port),
		Handler: mux,
	}
	err := c.server.ListenAndServe()
	if err != nil {
		if !c.closing {
			fmt.Println("Http服务启动失败:", err)
		}
	}
}

func (c *HttpComponent) registerHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/", c.handleRoot)
	mux.HandleFunc("/district/report", c.handleDistrictReport)
	mux.HandleFunc("/alert", c.handleAlert)
	mux.HandleFunc("/stop", c.handleStop)
	mux.HandleFunc("/ecs/list", c.handleListEcs)
	mux.HandleFunc("/process/list", c.handleListProcess)
	mux.HandleFunc("/receiver/list", c.handleListReceiver)
	mux.HandleFunc("/restart", c.handleRestart)
	mux.HandleFunc("/tick", c.handleTick)
}

func (c *HttpComponent) handleTick(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "text/plain;charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
	go c.app.onTick()
}

func (c *HttpComponent) handleRestart(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "text/plain;charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
	time.AfterFunc(1*time.Second, c.app.Restart)
}

func (c *HttpComponent) handleListReceiver(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "text/plain;charset=utf-8")
	w.WriteHeader(http.StatusOK)
	for _, receiver := range c.app.mailReceivers {
		_, _ = w.Write([]byte(receiver))
		_, _ = w.Write([]byte("\n"))
	}
}

func (c *HttpComponent) handleListProcess(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "text/plain;charset=utf-8")
	w.WriteHeader(http.StatusOK)
	c.app.updateProcesses()
	for _, p := range c.app.processInfoArr {
		_, _ = w.Write([]byte(p.String()))
		_, _ = w.Write([]byte("\n"))
	}
}

func (c *HttpComponent) handleListEcs(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "text/plain;charset=utf-8")
	w.WriteHeader(http.StatusOK)
	c.app.updateEcsInfo()
	_, _ = w.Write([]byte(c.app.ecsInfo.String()))
	_, _ = w.Write([]byte("\n"))
}

func (c *HttpComponent) handleRoot(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	c.app.updateEcsInfo()
	c.app.updateProcesses()

	w.Header().Set("Content-Type", "text/plain;charset=utf-8")
	w.WriteHeader(http.StatusOK)

	_, _ = w.Write([]byte("欢迎使用Silent Keeper!\n"))
	_, _ = w.Write([]byte(c.app.SelfInfo()))
	_, _ = w.Write([]byte("\n"))
	_, _ = w.Write([]byte("\n"))

	_, _ = w.Write([]byte(c.app.ecsInfo.String()))
	_, _ = w.Write([]byte("\n"))
	_, _ = w.Write([]byte("\n"))

	for _, p := range c.app.processInfoArr {
		_, _ = w.Write([]byte(p.String()))
		_, _ = w.Write([]byte("\n"))
	}
}

func (c *HttpComponent) handleStop(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "text/plain;charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
	time.AfterFunc(1*time.Second, c.app.exit)
}

func (c *HttpComponent) handleDistrictReport(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var data map[string]interface{}
	err := parseRequest(r, &data)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bat request\n"))
		return
	}
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("handleDistrictReport方法报错:", err)
			}
		}()

		now := time.Now()
		app := c.app
		timeZone := app.config.UserTimeZone
		for _, value := range data {
			district := value.(map[string]interface{})
			//id := district["id"].(string)
			name := district["name"].(string)
			onlineNum := int(district["online_num"].(float64))
			onlineLimit := int(district["online_limit"].(float64))
			registerNum := int(district["register_num"].(float64))
			registerLimit := int(district["register_limit"].(float64))
			onlinePercent := CalcPercent(onlineNum, onlineLimit)
			registerPercent := CalcPercent(registerNum, registerLimit)

			//在线人数过多警告
			if onlinePercent >= app.config.OnlinePercentAlert {
				level := "警告"
				if onlinePercent >= 98 {
					level = "严重"
				}
				msg := fmt.Sprintf("[%s] [%s] [%s] 当前在线人数已达%d/%d,请注意服务器运行状态", FormatTimeWithTimeZone(now, timeZone), level, name, onlineNum, onlineLimit)
				app.Alert2(msg)
			}

			//注册人数过多警告
			if registerPercent >= app.config.RegisterPercentAlert {
				level := "警告"
				if onlinePercent >= 98 {
					level = "严重"
				}
				msg := fmt.Sprintf("[%s] [%s] [%s] 当前注册人数已达%d/%d,请注意调整参数，避免新玩家无法注册", FormatTimeWithTimeZone(now, timeZone), level, name, registerNum, registerLimit)
				app.Alert2(msg)
			}
		}
	}()

	w.Header().Set("Content-Type", "text/plain;charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}

func (c *HttpComponent) handleAlert(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var data map[string]interface{}
	err := parseRequest(r, &data)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bat request\n"))
		return
	}

	msg, ok := data["msg"].(string)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	go c.app.Alert2(msg)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}

func (c *HttpComponent) Close() {
	if c.server != nil {
		c.closing = true
		_ = c.server.Close()
	}
}

func parseRequest(r *http.Request, v interface{}) error {
	switch r.Method {
	case http.MethodGet:
		{
			//将v断言成*map,如果断言成map拿到的将是一个副本
			params, ok := v.(*map[string]interface{})
			if !ok {
				return fmt.Errorf("GET请求需传递 *map[string]interface{}")
			}
			//解引用后的map是nil,则初始化
			if *params == nil {
				*params = make(map[string]interface{})
			}
			//将参数读取到map
			for key, values := range r.URL.Query() {
				if len(values) > 0 {
					(*params)[key] = values[0] //params是一个指向map的指针哦，所以需要解引用
				}
			}
			return nil
		}
	default:
		{
			return json.NewDecoder(r.Body).Decode(v)
		}
	}
}
