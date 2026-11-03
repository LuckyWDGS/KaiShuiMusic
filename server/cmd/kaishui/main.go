// 命令行 / 服务端版本：启动后在浏览器访问。
// 加 --addr 0.0.0.0:8686 可让同一局域网的手机、电视直接连接使用。
package main

import (
	"flag"
	"log"

	"github.com/LuckyWDGS/KaiShuiMusic/server/internal/app"
	"github.com/guohuiyuan/go-music-dl/core"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8686", "监听地址")
	dataDir := flag.String("data", "", "数据目录（默认系统用户配置目录）")
	open := flag.Bool("open", false, "启动后自动打开浏览器")
	flag.Parse()

	url, err := app.Start(app.Options{Addr: *addr, DataDir: *dataDir})
	if err != nil {
		log.Fatal(err)
	}
	if *open {
		core.OpenBrowser(url)
	}
	select {}
}
