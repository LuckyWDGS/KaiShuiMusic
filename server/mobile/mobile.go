// Package mobile 供安卓 APP 调用（gomobile bind）：在 APP 进程内启动本地服务，界面由 WebView 加载。
package mobile

import (
	"github.com/LuckyWDGS/KaiShuiMusic/server/internal/api"
	"github.com/LuckyWDGS/KaiShuiMusic/server/internal/app"
	"github.com/guohuiyuan/go-music-dl/core"
	_ "golang.org/x/mobile/bind" // gomobile bind 需要，保留在 go.mod 里
)

// 固定端口：WebView 的本地存储（音量、播放列表等）按地址区分，被占用时才退回随机端口
const (
	addr         = "127.0.0.1:37618"
	fallbackAddr = "127.0.0.1:0"
)

// Start 启动服务并返回访问地址。dataDir 是 APP 私有目录，downloadDir 是下载保存位置，version 是 APP 版本号。
func Start(dataDir, downloadDir, version string) (string, error) {
	api.Version = version
	url, err := app.Start(app.Options{Addr: addr, DataDir: dataDir, FallbackAddr: fallbackAddr})
	if err != nil {
		return "", err
	}
	if settings := core.GetWebSettings(); settings.DownloadDir != downloadDir {
		settings.DownloadDir = downloadDir
		if err := core.SaveWebSettings(settings); err != nil {
			return "", err
		}
	}
	return url, nil
}
