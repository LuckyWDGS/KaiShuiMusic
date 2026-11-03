//go:build windows

package main

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/jchv/go-webview2"
	"golang.org/x/sys/windows"
)

// 系统托盘（通知区域）图标：程序常驻后台。关闭主窗口只是隐藏到托盘，音乐继续播放；
// 左键单击图标打开主窗口，右键菜单可以控制播放或退出。

var (
	shell32              = syscall.NewLazyDLL("shell32.dll")
	procShellNotifyIcon  = shell32.NewProc("Shell_NotifyIconW")
	procLoadImage        = user32.NewProc("LoadImageW")
	procCreatePopupMenu  = user32.NewProc("CreatePopupMenu")
	procAppendMenu       = user32.NewProc("AppendMenuW")
	procTrackPopupMenu   = user32.NewProc("TrackPopupMenu")
	procDestroyMenu      = user32.NewProc("DestroyMenu")
	procGetCursorPos     = user32.NewProc("GetCursorPos")
	procRegisterWindowMs = user32.NewProc("RegisterWindowMessageW")
)

const (
	nimAdd        = 0
	nimModify     = 1
	nimDelete     = 2
	nifMessage    = 0x1
	nifIcon       = 0x2
	nifTip        = 0x4
	wmApp         = 0x8000
	wmTrayIcon    = wmApp + 17
	wmNull        = 0x0000
	wmLButtonUp   = 0x0202
	wmRButtonUp   = 0x0205
	imageIcon     = 1
	lrShared      = 0x8000
	smCxSmIcon    = 49
	smCySmIcon    = 50
	mfString      = 0x0
	mfSeparator   = 0x800
	tpmReturnCmd  = 0x0100
	tpmRightBtn   = 0x0002
	tpmBottomAlgn = 0x0020
)

// 托盘菜单命令
const (
	cmdShow = iota + 1
	cmdToggle
	cmdPrev
	cmdNext
	cmdQuit
)

type notifyIconData struct {
	CbSize           uint32
	HWnd             uintptr
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            uintptr
	SzTip            [128]uint16
	DwState          uint32
	DwStateMask      uint32
	SzInfo           [256]uint16
	UVersion         uint32
	SzInfoTitle      [64]uint16
	DwInfoFlags      uint32
	GuidItem         [16]byte
	HBalloonIcon     uintptr
}

type trayIcon struct {
	w       webview2.WebView
	hwnd    uintptr
	data    notifyIconData
	playing bool
	// 资源管理器重启后任务栏重建时会广播此消息，需要重新添加图标
	taskbarCreated uintptr
}

func newTrayIcon(w webview2.WebView, hwnd uintptr) *trayIcon {
	t := &trayIcon{w: w, hwnd: hwnd}
	var hinst windows.Handle
	_ = windows.GetModuleHandleEx(0, nil, &hinst)
	cx, _, _ := procGetSysMetrics.Call(smCxSmIcon)
	cy, _, _ := procGetSysMetrics.Call(smCySmIcon)
	icon, _, _ := procLoadImage.Call(uintptr(hinst), iconID, imageIcon, cx, cy, lrShared)

	t.data.CbSize = uint32(unsafe.Sizeof(t.data))
	t.data.HWnd = hwnd
	t.data.UID = 1
	t.data.UFlags = nifMessage | nifIcon | nifTip
	t.data.UCallbackMessage = wmTrayIcon
	t.data.HIcon = icon
	t.setTip("开水音乐")
	procShellNotifyIcon.Call(nimAdd, uintptr(unsafe.Pointer(&t.data)))

	name, _ := syscall.UTF16PtrFromString("TaskbarCreated")
	t.taskbarCreated, _, _ = procRegisterWindowMs.Call(uintptr(unsafe.Pointer(name)))
	return t
}

func (t *trayIcon) setTip(text string) {
	u, _ := syscall.UTF16FromString(text)
	if len(u) > len(t.data.SzTip) {
		u = append(u[:len(t.data.SzTip)-1], 0)
	}
	t.data.SzTip = [128]uint16{}
	copy(t.data.SzTip[:], u)
}

// Update 更新悬停提示（正在播放的歌曲）和播放状态
func (t *trayIcon) Update(playing bool, title string) {
	t.playing = playing
	tip := "开水音乐"
	if title != "" {
		tip = fmt.Sprintf("%s\n开水音乐", title)
	}
	t.setTip(tip)
	procShellNotifyIcon.Call(nimModify, uintptr(unsafe.Pointer(&t.data)))
}

func (t *trayIcon) Remove() {
	procShellNotifyIcon.Call(nimDelete, uintptr(unsafe.Pointer(&t.data)))
}

// handle 处理托盘相关的窗口消息，返回 true 表示已处理
func (t *trayIcon) handle(msg, lParam uintptr) bool {
	if t.taskbarCreated != 0 && msg == t.taskbarCreated {
		procShellNotifyIcon.Call(nimAdd, uintptr(unsafe.Pointer(&t.data)))
		return true
	}
	if msg != wmTrayIcon {
		return false
	}
	switch lParam & 0xFFFF {
	case wmLButtonUp:
		showMain(t.hwnd)
	case wmRButtonUp:
		t.showMenu()
	}
	return true
}

func (t *trayIcon) showMenu() {
	menu, _, _ := procCreatePopupMenu.Call()
	defer procDestroyMenu.Call(menu)
	add := func(id uintptr, text string) {
		s, _ := syscall.UTF16PtrFromString(text)
		procAppendMenu.Call(menu, mfString, id, uintptr(unsafe.Pointer(s)))
	}
	add(cmdShow, "打开开水音乐")
	procAppendMenu.Call(menu, mfSeparator, 0, 0)
	if t.playing {
		add(cmdToggle, "暂停")
	} else {
		add(cmdToggle, "播放")
	}
	add(cmdPrev, "上一首")
	add(cmdNext, "下一首")
	procAppendMenu.Call(menu, mfSeparator, 0, 0)
	add(cmdQuit, "退出")

	var pt struct{ X, Y int32 }
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	// 菜单弹出前需要把窗口设为前台，否则点击菜单外部时菜单不会消失
	procSetForeground.Call(t.hwnd)
	cmd, _, _ := procTrackPopupMenu.Call(menu, tpmReturnCmd|tpmRightBtn|tpmBottomAlgn, uintptr(pt.X), uintptr(pt.Y), 0, t.hwnd, 0)
	procPostMessage.Call(t.hwnd, wmNull, 0, 0)

	switch cmd {
	case cmdShow:
		showMain(t.hwnd)
	case cmdToggle:
		t.w.Eval(`window.__kaishuiTray && window.__kaishuiTray('toggle')`)
	case cmdPrev:
		t.w.Eval(`window.__kaishuiTray && window.__kaishuiTray('prev')`)
	case cmdNext:
		t.w.Eval(`window.__kaishuiTray && window.__kaishuiTray('next')`)
	case cmdQuit:
		quitApp(t.hwnd)
	}
}

// showMain 显示主窗口（可能已隐藏到托盘或最小化）并切到前台
func showMain(hwnd uintptr) {
	if iconic, _, _ := procIsIconic.Call(hwnd); iconic != 0 {
		procShowWindow.Call(hwnd, swRestore)
	} else {
		procShowWindow.Call(hwnd, swShow)
	}
	procSetForeground.Call(hwnd)
}
