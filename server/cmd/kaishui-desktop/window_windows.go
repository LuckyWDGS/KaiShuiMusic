//go:build windows

package main

import (
	"strings"
	"syscall"
	"unsafe"

	"github.com/jchv/go-webview2"
)

var (
	user32             = syscall.NewLazyDLL("user32.dll")
	procSetWindowLng   = user32.NewProc("SetWindowLongPtrW")
	procCallWindowProc = user32.NewProc("CallWindowProcW")
	procSetWindowPos   = user32.NewProc("SetWindowPos")
	procShowWindow     = user32.NewProc("ShowWindow")
	procIsZoomed       = user32.NewProc("IsZoomed")
	procReleaseCap     = user32.NewProc("ReleaseCapture")
	procSendMessage    = user32.NewProc("SendMessageW")
	procPostMessage    = user32.NewProc("PostMessageW")
	procGetSysMetrics  = user32.NewProc("GetSystemMetrics")
)

const (
	gwlpWndProc      = ^uintptr(3) // -4
	swpFrameChanged  = 0x0020
	swpNoMove        = 0x0002
	swpNoSize        = 0x0001
	swpNoZOrder      = 0x0004
	swMinimize       = 6
	swMaximize       = 3
	swRestore        = 9
	swShow           = 5
	wmNCCalcSize     = 0x0083
	wmNCLButtonDown  = 0x00A1
	wmClose          = 0x0010
	htCaption        = 2
	smCyFrame        = 33
	smCxPaddedBorder = 92
)

type rect struct{ Left, Top, Right, Bottom int32 }

type ncCalcSizeParams struct {
	Rgrc  [3]rect
	Lppos uintptr
}

var (
	origWndProc uintptr
	tray        *trayIcon
	// 只有从托盘菜单"退出"时才真正关闭程序，否则关闭按钮只是隐藏到托盘
	quitting bool
)

// framelessWndProc 主窗口的窗口过程：
//   - 去掉系统标题栏和顶部边框，保留左右下的缩放边框、贴边分屏和窗口动画（同 Windows Terminal 的做法）
//   - 关闭窗口时隐藏到托盘
//   - 处理托盘图标消息
func framelessWndProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	switch {
	case msg == wmNCCalcSize && wParam != 0:
		params := (*ncCalcSizeParams)(unsafe.Pointer(lParam))
		top := params.Rgrc[0].Top
		ret, _, _ := procCallWindowProc.Call(origWndProc, hwnd, msg, wParam, lParam)
		params.Rgrc[0].Top = top
		// 最大化时窗口会超出屏幕一个边框的宽度，顶部需要补回来
		if zoomed, _, _ := procIsZoomed.Call(hwnd); zoomed != 0 {
			frame, _, _ := procGetSysMetrics.Call(smCyFrame)
			pad, _, _ := procGetSysMetrics.Call(smCxPaddedBorder)
			params.Rgrc[0].Top += int32(frame + pad)
		}
		return ret
	case msg == wmClose && !quitting && tray != nil:
		procShowWindow.Call(hwnd, swHide)
		return 0
	case tray != nil && tray.handle(msg, lParam):
		return 0
	}
	ret, _, _ := procCallWindowProc.Call(origWndProc, hwnd, msg, wParam, lParam)
	return ret
}

// makeFrameless 用自定义窗口过程替换系统标题栏，标题栏由前端自绘。
func makeFrameless(hwnd uintptr) {
	origWndProc, _, _ = procSetWindowLng.Call(hwnd, gwlpWndProc, syscall.NewCallback(framelessWndProc))
	procSetWindowPos.Call(hwnd, 0, 0, 0, 0, 0, swpFrameChanged|swpNoMove|swpNoSize|swpNoZOrder)
}

// quitApp 真正退出程序（托盘菜单"退出"）
func quitApp(hwnd uintptr) {
	quitting = true
	if tray != nil {
		tray.Remove()
	}
	procPostMessage.Call(hwnd, wmClose, 0, 0)
}

// bindWindowControls 向前端暴露 window.kaishuiWin(action)：
// drag / min / max / close（隐藏到托盘）/ mini:on / mini:off / tip:<0|1>:<歌名 - 歌手>
func bindWindowControls(w webview2.WebView, hwnd uintptr, mini *miniPlayer) {
	_ = w.Bind("kaishuiWin", func(action string) {
		switch {
		case action == "mini:on", action == "mini:off":
			// 创建 WebView2 会嵌套消息循环，放到主循环里执行，避免在消息回调中重入
			on := action == "mini:on"
			w.Dispatch(func() { mini.SetVisible(on) })
		case action == "drag":
			procReleaseCap.Call()
			procSendMessage.Call(hwnd, wmNCLButtonDown, htCaption, 0)
		case action == "min":
			procShowWindow.Call(hwnd, swMinimize)
		case action == "max":
			if zoomed, _, _ := procIsZoomed.Call(hwnd); zoomed != 0 {
				procShowWindow.Call(hwnd, swRestore)
			} else {
				procShowWindow.Call(hwnd, swMaximize)
			}
		case action == "close":
			procPostMessage.Call(hwnd, wmClose, 0, 0)
		case strings.HasPrefix(action, "tip:") && tray != nil:
			// tip:1:歌名 - 歌手（1 表示正在播放）
			rest := strings.TrimPrefix(action, "tip:")
			playing, title, _ := strings.Cut(rest, ":")
			tray.Update(playing == "1", title)
		}
	})
}
