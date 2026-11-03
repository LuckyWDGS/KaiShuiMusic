//go:build !windows

// 非 Windows 平台暂时以浏览器方式打开（macOS / Linux 原生窗口待后续接入）。
package main

import (
	"log"

	"github.com/LuckyWDGS/KaiShuiMusic/server/internal/app"
	"github.com/guohuiyuan/go-music-dl/core"
)

func main() {
	url, err := app.Start(app.Options{Addr: "127.0.0.1:0"})
	if err != nil {
		log.Fatal(err)
	}
	core.OpenBrowser(url + "/?platform=pc")
	select {}
}
