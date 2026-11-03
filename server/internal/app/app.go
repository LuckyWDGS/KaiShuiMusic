// Package app 组装并启动后端服务，供命令行版和桌面版共用。
package app

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"path/filepath"

	"github.com/LuckyWDGS/KaiShuiMusic/server/internal/api"
	"github.com/LuckyWDGS/KaiShuiMusic/server/internal/appdata"
	"github.com/LuckyWDGS/KaiShuiMusic/server/internal/library"
	"github.com/LuckyWDGS/KaiShuiMusic/server/internal/music"
	"github.com/LuckyWDGS/KaiShuiMusic/server/internal/webui"
	"github.com/guohuiyuan/go-music-dl/core"
)

type Options struct {
	Addr    string // 监听地址，如 127.0.0.1:0（随机端口）或 0.0.0.0:8686（局域网可访问，供手机 / 电视连接）
	DataDir string // 数据目录，为空时使用系统默认目录
}

// Start 启动服务并返回访问地址，服务在后台 goroutine 中运行。
func Start(opts Options) (string, error) {
	dataDir, err := appdata.Prepare(opts.DataDir)
	if err != nil {
		return "", err
	}
	core.CM.Load()

	lib, err := library.Open(filepath.Join(dataDir, "data", "library.json"))
	if err != nil {
		return "", fmt.Errorf("打开音乐库失败: %w", err)
	}
	svc := music.NewService(filepath.Join(dataDir, "data", "cache"))
	go svc.Warmup()
	srv := api.New(svc, lib, webui.FS(), dataDir)

	ln, err := net.Listen("tcp", opts.Addr)
	if err != nil {
		return "", fmt.Errorf("监听 %s 失败: %w", opts.Addr, err)
	}
	go func() {
		if err := http.Serve(ln, srv.Handler()); err != nil {
			log.Printf("server stopped: %v", err)
		}
	}()

	addr := ln.Addr().(*net.TCPAddr)
	host := addr.IP.String()
	if addr.IP.IsUnspecified() {
		host = "127.0.0.1"
	}
	log.Printf("开水音乐已启动: http://%s:%d  数据目录: %s", host, addr.Port, dataDir)
	return fmt.Sprintf("http://%s:%d", host, addr.Port), nil
}
