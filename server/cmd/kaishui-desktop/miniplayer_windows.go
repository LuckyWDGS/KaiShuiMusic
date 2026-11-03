//go:build windows

package main

import (
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"

	"github.com/jchv/go-webview2"
	"golang.org/x/sys/windows/registry"
)

// 任务栏播放器：第二个 WebView2 窗口，作为子窗口嵌入 Windows 任务栏（Shell_TrayWnd），
// 显示封面、歌名、当前歌词和播放控制。界面是同一前端的 /?platform=mini 页面，
// 与主窗口通过 BroadcastChannel 同步状态（两个 WebView2 共用同一个数据目录）。

var (
	procFindWindow      = user32.NewProc("FindWindowW")
	procFindWindowEx    = user32.NewProc("FindWindowExW")
	procGetWindowRect   = user32.NewProc("GetWindowRect")
	procSetParent       = user32.NewProc("SetParent")
	procSetWindowLongW  = user32.NewProc("SetWindowLongPtrW")
	procGetWindowLongW  = user32.NewProc("GetWindowLongPtrW")
	procMoveWindow      = user32.NewProc("MoveWindow")
	procIsWindow        = user32.NewProc("IsWindow")
	procIsIconic        = user32.NewProc("IsIconic")
	procSetForeground   = user32.NewProc("SetForegroundWindow")
	procDefWindowProc   = user32.NewProc("DefWindowProcW")
	procGetDC           = user32.NewProc("GetDC")
	procReleaseDC       = user32.NewProc("ReleaseDC")
	procGetDpiForWindow = user32.NewProc("GetDpiForWindow")
	procGetPixel        = syscall.NewLazyDLL("gdi32.dll").NewProc("GetPixel")
)

const (
	gwlStyle      = ^uintptr(15) // -16
	gwlExStyle    = ^uintptr(19) // -20
	wsChild       = 0x40000000
	wsClipSibling = 0x04000000
	wsExToolWin   = 0x00000080
	wsExAppWin    = 0x00040000
	swHide        = 0
	swShowNA      = 8
	wmDestroy     = 0x0002

	miniWidth  = 300 // 逻辑像素
	miniLeft   = 110 // 居中任务栏时放在左侧小组件（天气）之后
	miniMargin = 8
)

type miniPlayer struct {
	main     webview2.WebView
	mainHwnd uintptr
	url      string

	w       webview2.WebView
	hwnd    uintptr
	tray    uintptr
	visible bool
	bg      uint32 // 上次采样的任务栏颜色（COLORREF）
	ticker  *time.Ticker
}

func newMiniPlayer(main webview2.WebView, mainHwnd uintptr, url string) *miniPlayer {
	return &miniPlayer{main: main, mainHwnd: mainHwnd, url: url}
}

// SetVisible 开关任务栏播放器（在 UI 线程调用）
func (m *miniPlayer) SetVisible(on bool) {
	if !on {
		m.visible = false
		if m.hwnd != 0 {
			procShowWindow.Call(m.hwnd, swHide)
		}
		return
	}
	m.visible = true
	if m.hwnd == 0 || !isWindow(m.hwnd) || !isWindow(m.tray) {
		m.create()
	}
	if m.hwnd != 0 {
		m.place()
		procShowWindow.Call(m.hwnd, swShowNA)
	}
}

func (m *miniPlayer) create() {
	m.tray = findWindow("Shell_TrayWnd")
	if m.tray == 0 {
		return
	}
	// WebView2 背景透明，直接透出任务栏本身（Windows 11 任务栏是随壁纸变化的渐变色，取色无法完全吻合）
	m.bg = m.sampleTaskbarColor()
	_ = os.Setenv("WEBVIEW2_DEFAULT_BACKGROUND_COLOR", "00000000")

	w := webview2.NewWithOptions(webview2.WebViewOptions{
		WindowOptions: webview2.WindowOptions{Title: "开水音乐 任务栏播放器", Width: 1, Height: 1, IconId: iconID},
	})
	if w == nil {
		return
	}
	hwnd := uintptr(w.Window())
	procShowWindow.Call(hwnd, swHide)

	// 变成任务栏的子窗口：去掉标题栏和任务栏按钮
	procSetWindowLongW.Call(hwnd, gwlStyle, wsChild|wsClipSibling)
	ex, _, _ := procGetWindowLongW.Call(hwnd, gwlExStyle)
	procSetWindowLongW.Call(hwnd, gwlExStyle, (ex|wsExToolWin)&^wsExAppWin)
	procSetParent.Call(hwnd, m.tray)
	m.hookDestroy(hwnd)

	// 页面调用 kaishuiMini('colors') 获取任务栏配色，kaishuiMini('show') 唤起主窗口
	_ = w.Bind("kaishuiMini", func(action string) []string {
		switch action {
		case "show":
			if iconic, _, _ := procIsIconic.Call(m.mainHwnd); iconic != 0 {
				procShowWindow.Call(m.mainHwnd, swRestore)
			}
			procSetForeground.Call(m.mainHwnd)
		case "colors":
			bg, fg := m.colors()
			return []string{bg, fg}
		}
		return nil
	})
	w.Navigate(m.url + "/?platform=mini")

	m.w, m.hwnd = w, hwnd
	m.applyColors()
	if m.ticker == nil {
		// 定期跟随任务栏颜色和位置变化（切换深浅色、调整任务栏等）
		m.ticker = time.NewTicker(3 * time.Second)
		go func() {
			for range m.ticker.C {
				m.main.Dispatch(m.refresh)
			}
		}()
	}
}

func (m *miniPlayer) refresh() {
	if !m.visible || m.hwnd == 0 {
		return
	}
	if !isWindow(m.hwnd) || !isWindow(m.tray) {
		// 资源管理器重启后任务栏会重建，重新嵌入
		m.hwnd = 0
		m.SetVisible(true)
		return
	}
	m.place()
	if c := m.sampleTaskbarColor(); c != m.bg {
		m.bg = c
		m.applyColors()
	}
}

// place 计算播放器在任务栏中的位置：任务栏图标居中时放在左侧，靠左时放在通知区域前面
func (m *miniPlayer) place() {
	var tr rect
	procGetWindowRect.Call(m.tray, uintptr(unsafe.Pointer(&tr)))
	dpi, _, _ := procGetDpiForWindow.Call(m.tray)
	if dpi == 0 {
		dpi = 96
	}
	scale := float64(dpi) / 96
	h := tr.Bottom - tr.Top
	w := int32(miniWidth * scale)
	x := int32(miniLeft * scale)
	if !taskbarCentered() {
		if notify := findChild(m.tray, "TrayNotifyWnd"); notify != 0 {
			var nr rect
			procGetWindowRect.Call(notify, uintptr(unsafe.Pointer(&nr)))
			x = nr.Left - tr.Left - w - int32(miniMargin*scale)
		}
	}
	procMoveWindow.Call(m.hwnd, uintptr(x), 0, uintptr(w), uintptr(h), 1)
}

// sampleTaskbarColor 取播放器右侧一小段任务栏的颜色，让播放器背景与任务栏融为一体
func (m *miniPlayer) sampleTaskbarColor() uint32 {
	tray := m.tray
	if tray == 0 {
		return 0x1C1C1C
	}
	var tr rect
	procGetWindowRect.Call(tray, uintptr(unsafe.Pointer(&tr)))
	dpi, _, _ := procGetDpiForWindow.Call(tray)
	if dpi == 0 {
		dpi = 96
	}
	scale := float64(dpi) / 96
	x := tr.Left + int32((miniLeft+miniWidth+6)*scale)
	y := tr.Bottom - int32(4*scale)
	dc, _, _ := procGetDC.Call(0)
	c, _, _ := procGetPixel.Call(dc, uintptr(x), uintptr(y))
	procReleaseDC.Call(0, dc)
	if c == 0xFFFFFFFF { // CLR_INVALID
		return 0x1C1C1C
	}
	return uint32(c)
}

func (m *miniPlayer) applyColors() {
	if m.w == nil {
		return
	}
	bg, fg := m.colors()
	m.w.Eval(fmt.Sprintf(`window.__kaishuiTaskbar && window.__kaishuiTaskbar(%q, %q)`, bg, fg))
}

// colors 返回 CSS 背景色和与之对比的文字颜色
func (m *miniPlayer) colors() (string, string) {
	r, g, b := m.bg&0xFF, (m.bg>>8)&0xFF, (m.bg>>16)&0xFF
	fg := "#ffffff"
	if (r*299+g*587+b*114)/1000 > 150 {
		fg = "#111111"
	}
	return fmt.Sprintf("rgb(%d,%d,%d)", r, g, b), fg
}

var miniOrigProc uintptr

// hookDestroy 拦截播放器窗口的 WM_DESTROY：任务栏被销毁时（资源管理器重启）不能让 go-webview2 退出整个程序
func (m *miniPlayer) hookDestroy(hwnd uintptr) {
	miniOrigProc, _, _ = procSetWindowLng.Call(hwnd, gwlpWndProc, syscall.NewCallback(func(h, msg, wp, lp uintptr) uintptr {
		if msg == wmDestroy {
			m.hwnd = 0
			r, _, _ := procDefWindowProc.Call(h, msg, wp, lp)
			return r
		}
		r, _, _ := procCallWindowProc.Call(miniOrigProc, h, msg, wp, lp)
		return r
	}))
}

// taskbarCentered 判断任务栏图标是否居中（Windows 11 默认居中；Windows 10 始终靠左）
func taskbarCentered() bool {
	if k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE); err == nil {
		build, _, _ := k.GetStringValue("CurrentBuildNumber")
		k.Close()
		var n int
		fmt.Sscan(build, &n)
		if n > 0 && n < 22000 {
			return false
		}
	}
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Explorer\Advanced`, registry.QUERY_VALUE)
	if err != nil {
		return true
	}
	defer k.Close()
	v, _, err := k.GetIntegerValue("TaskbarAl")
	return err != nil || v == 1
}

func findWindow(class string) uintptr {
	c, _ := syscall.UTF16PtrFromString(class)
	h, _, _ := procFindWindow.Call(uintptr(unsafe.Pointer(c)), 0)
	return h
}

func findChild(parent uintptr, class string) uintptr {
	c, _ := syscall.UTF16PtrFromString(class)
	h, _, _ := procFindWindowEx.Call(parent, 0, uintptr(unsafe.Pointer(c)), 0)
	return h
}

func isWindow(h uintptr) bool {
	if h == 0 {
		return false
	}
	r, _, _ := procIsWindow.Call(h)
	return r != 0
}
