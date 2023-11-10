//go:build windows
// +build windows

package log

import (
	"syscall"
	"unsafe"
)

var (
	kernel32           *syscall.LazyDLL
	procGetConsoleMode *syscall.LazyProc
	procSetConsoleMode *syscall.LazyProc
)

const EnableVirtualTerminalProcessingMode uint32 = 0x4

var colorEnable = true

func init() {
	colorEnable = false
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	// https://docs.microsoft.com/en-us/windows/console/setconsolemode
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode = kernel32.NewProc("SetConsoleMode")
	handle, err := syscall.Open("CONOUT$", syscall.O_RDWR, 0)
	if err != nil {
		Warn("open CONOUT$ error", err)
		return
	}
	var mode uint32
	if r1, _, err := procGetConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode))); r1 == 0 {
		Warn("console mode get error", err)
	}
	mode |= EnableVirtualTerminalProcessingMode
	if r1, _, err := procSetConsoleMode.Call(uintptr(handle), uintptr(mode)); r1 == 0 {
		Warn("console mode set error", err)
	}
	colorEnable = true
}
