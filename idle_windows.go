//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

type LASTINPUTINFO struct {
	CbSize uint32
	DwTime uint32
}

var (
	user32           = syscall.NewLazyDLL("user32.dll")
	getLastInputInfo = user32.NewProc("GetLastInputInfo")
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	getTickCount     = kernel32.NewProc("GetTickCount")
)

// GetSystemIdleTime retrieves the user idle time on Windows in seconds.
func GetSystemIdleTime() (uint32, error) {
	var lii LASTINPUTINFO
	lii.CbSize = uint32(unsafe.Sizeof(lii))

	r, _, err := getLastInputInfo.Call(uintptr(unsafe.Pointer(&lii)))
	if r == 0 {
		return 0, err
	}

	tickCount, _, _ := getTickCount.Call()
	idleMs := uint32(tickCount) - lii.DwTime
	return idleMs / 1000, nil
}

var (
	getForegroundWindow = user32.NewProc("GetForegroundWindow")
	getWindowText       = user32.NewProc("GetWindowTextW")
)

// GetActiveWindowTitle retrieves the active window's title on Windows.
func GetActiveWindowTitle() string {
	hwnd, _, _ := getForegroundWindow.Call()
	if hwnd == 0 {
		return "Idle"
	}

	buf := make([]uint16, 512)
	r, _, _ := getWindowText.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), 512)
	if r == 0 {
		return "System"
	}

	return syscall.UTF16ToString(buf[:r])
}
