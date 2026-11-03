//go:build windows

// PC 桌面版：内置服务 + WebView2 无边框窗口 + 任务栏播放器。
// 构建: go build -ldflags "-H windowsgui" ./cmd/kaishui-desktop
// 图标与高 DPI 清单来自 rsrc_windows_*.syso（由 winres/winres.json 经 go-winres 生成）。
package main

import (
	"log"
	"os/exec"
	"syscall"
	"unsafe"

	"github.com/LuckyWDGS/KaiShuiMusic/server/internal/app"
	"github.com/jchv/go-webview2"
)

// 窗口尺寸（逻辑像素，按系统缩放换算成物理像素）
const (
	winWidth     = 1100
	winHeight    = 720
	winMinWidth  = 960
	winMinHeight = 640
	iconID       = 1
)

func main() {
	url, err := app.Start(app.Options{Addr: "127.0.0.1:0"})
	if err != nil {
		messageBox("开水音乐", "启动失败: "+err.Error(), 0)
		log.Fatal(err)
	}

	scale := systemScale()
	w := webview2.NewWithOptions(webview2.WebViewOptions{
		AutoFocus: true,
		WindowOptions: webview2.WindowOptions{
			Title:  "开水音乐",
			Width:  uint(float64(winWidth) * scale),
			Height: uint(float64(winHeight) * scale),
			IconId: iconID,
			Center: true,
		},
	})
	if w == nil {
		if messageBox("开水音乐", "无法打开 WebView2 运行时，请安装 Microsoft Edge WebView2 后重试。\n点击“确定”前往下载。", 1) == 1 {
			_ = exec.Command("cmd", "/c", "start", "https://developer.microsoft.com/microsoft-edge/webview2/").Start()
		}
		log.Fatal("webview2 unavailable")
	}
	defer w.Destroy()
	w.SetSize(int(winMinWidth*scale), int(winMinHeight*scale), webview2.HintMin)
	hwnd := uintptr(w.Window())
	makeFrameless(hwnd)
	mini := newMiniPlayer(w, hwnd, url)
	bindWindowControls(w, hwnd, mini)
	w.Navigate(url + "/?platform=pc")
	w.Run()
}

// systemScale 返回系统缩放比例（如 175% → 1.75）。程序清单声明了 Per-Monitor V2 DPI 感知，窗口尺寸需要自己换算。
func systemScale() float64 {
	proc := syscall.NewLazyDLL("user32.dll").NewProc("GetDpiForSystem")
	if proc.Find() != nil {
		return 1
	}
	dpi, _, _ := proc.Call()
	if dpi == 0 {
		return 1
	}
	return float64(dpi) / 96
}

func messageBox(title, text string, flags uintptr) int {
	t, _ := syscall.UTF16PtrFromString(title)
	m, _ := syscall.UTF16PtrFromString(text)
	ret, _, _ := syscall.NewLazyDLL("user32.dll").NewProc("MessageBoxW").Call(0, uintptr(unsafe.Pointer(m)), uintptr(unsafe.Pointer(t)), flags)
	return int(ret)
}
