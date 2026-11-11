// 中文社区版修改文件；变更概述见 docs/MODIFICATIONS.md。
package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unsafe"

	webview2 "github.com/jchv/go-webview2"
	"golang.org/x/sys/windows"
)

var (
	findWindowEx     = user32.NewProc("FindWindowExW")
	getSystemMetrics = user32.NewProc("GetSystemMetrics")
	showWindow       = user32.NewProc("ShowWindow")
	isIconic         = user32.NewProc("IsIconic")
	setForeground    = user32.NewProc("SetForegroundWindow")
)

func accountWindowHandles() map[uintptr]bool {
	title, _ := windows.UTF16PtrFromString("Claude 账号管理")
	exe, _ := os.Executable()
	result := map[uintptr]bool{}
	var after uintptr
	for {
		hwnd, _, _ := findWindowEx.Call(0, after, 0, uintptr(unsafe.Pointer(title)))
		if hwnd == 0 {
			return result
		}
		after = hwnd
		var pid uint32
		if _, err := windows.GetWindowThreadProcessId(windows.HWND(hwnd), &pid); err != nil {
			continue
		}
		if strings.EqualFold(filepath.Clean(processPath(pid)), filepath.Clean(exe)) && windows.IsWindowVisible(windows.HWND(hwnd)) {
			result[hwnd] = true
		}
	}
}

func findAccountWindow(previous map[uintptr]bool) uintptr {
	for attempt := 0; attempt < 30; attempt++ {
		for hwnd := range accountWindowHandles() {
			if !previous[hwnd] {
				return hwnd
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	return 0
}

func runAccountWindow(url string) int {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	w, h := uintptr(760), uintptr(620)
	sw, _, _ := getSystemMetrics.Call(0)
	sh, _, _ := getSystemMetrics.Call(1)
	if sw > 0 && w > sw {
		w = sw
	}
	if sh > 0 && h > sh {
		h = sh
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return 1
	}
	view := webview2.NewWithOptions(webview2.WebViewOptions{
		DataPath:      filepath.Join(home, ".claude-account-switcher", "account-webview"),
		AutoFocus:     true,
		WindowOptions: webview2.WindowOptions{Title: "Claude 账号管理", Width: uint(w), Height: uint(h), IconId: 1, Center: true},
	})
	if view == nil {
		return 1
	}
	defer view.Destroy()
	view.Navigate(url)
	go closeWhenParentExits(os.Stdin, func() { view.Dispatch(view.Terminate) })
	view.Run()
	return 0
}

func activateAccountWindow(hwnd uintptr) bool {
	if !accountWindowHandles()[hwnd] {
		return false
	}
	if minimized, _, _ := isIconic.Call(hwnd); minimized != 0 {
		showWindow.Call(hwnd, 9)
	}
	setForeground.Call(hwnd)
	return true
}
