package main

import "idlekeeper/sk"

func main() {
	//如果使用 -ldflags "-H windowsgui" 来模拟在windows后台运行，需要加上下面的重定向，否则GC疯狂跑
	//f, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	//if err == nil {
	//	os.Stdout = f
	//	os.Stderr = f
	//}
	keeper := new(sk.Keeper)
	keeper.Start()
}
