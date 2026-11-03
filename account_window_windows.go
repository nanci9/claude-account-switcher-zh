package main

import (
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	findWindowEx     = user32.NewProc("FindWindowExW")
	setWindowPos     = user32.NewProc("SetWindowPos")
	getSystemMetrics = user32.NewProc("GetSystemMetrics")
	isZoomed         = user32.NewProc("IsZoomed")
	showWindow       = user32.NewProc("ShowWindow")
	isIconic         = user32.NewProc("IsIconic")
	setForeground    = user32.NewProc("SetForegroundWindow")
	setThreadDPI     = user32.NewProc("SetThreadDpiAwarenessContext")
)

func accountWindowHandles() map[uintptr]bool {
	title, _ := windows.UTF16PtrFromString("Claude 账号管理")
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
		if strings.EqualFold(filepath.Base(processPath(pid)), "msedge.exe") && windows.IsWindowVisible(windows.HWND(hwnd)) {
			result[hwnd] = true
		}
	}
}

// Edge may reuse its existing process and ignore --window-size; set physical pixels.
func resizeAccountWindow(previous map[uintptr]bool) uintptr {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	old, _, _ := setThreadDPI.Call(^uintptr(3)) // DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2
	if old != 0 {
		defer setThreadDPI.Call(old)
	}
	var target uintptr
	for attempt := 0; attempt < 30; attempt++ {
		for hwnd := range accountWindowHandles() {
			if !previous[hwnd] {
				target = hwnd
				break
			}
		}
		if target != 0 {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if target == 0 {
		return 0
	}
	if zoomed, _, _ := isZoomed.Call(target); zoomed != 0 {
		showWindow.Call(target, 9)
	} // SW_RESTORE
	w, h := uintptr(760), uintptr(620)
	sw, _, _ := getSystemMetrics.Call(0)
	sh, _, _ := getSystemMetrics.Call(1)
	if sw > 0 && w > sw {
		w = sw
	}
	if sh > 0 && h > sh {
		h = sh
	}
	x, y := uintptr(0), uintptr(0)
	if sw > w {
		x = (sw - w) / 2
	}
	if sh > h {
		y = (sh - h) / 2
	}
	setWindowPos.Call(target, 0, x, y, w, h, 0x0014) // NOZORDER | NOACTIVATE
	return target
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
