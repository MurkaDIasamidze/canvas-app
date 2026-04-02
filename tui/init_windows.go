//go:build windows

package tui

import (
	"syscall"
	"unsafe"
)

var (
	kernel32               = syscall.NewLazyDLL("kernel32.dll")
	procGetStdHandle       = kernel32.NewProc("GetStdHandle")
	procGetConsoleMode     = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode     = kernel32.NewProc("SetConsoleMode")
	procSetConsoleOutputCP = kernel32.NewProc("SetConsoleOutputCP")
)

const (
	stdOutputHandle                 = ^uintptr(10) // -11
	stdInputHandle                  = ^uintptr(9)  // -10
	enableVirtualTerminalProcessing = 0x0004
	enableVirtualTerminalInput      = 0x0200
	cp_UTF8                         = 65001
)

// InitTerminal enables ANSI colors and UTF-8 on Windows CMD/PowerShell
func InitTerminal() {
	// UTF-8 output
	procSetConsoleOutputCP.Call(uintptr(cp_UTF8))

	// Enable VT on stdout
	hOut, _, _ := procGetStdHandle.Call(stdOutputHandle)
	if hOut != 0 {
		var mode uint32
		procGetConsoleMode.Call(hOut, uintptr(unsafe.Pointer(&mode)))
		procSetConsoleMode.Call(hOut, uintptr(mode|enableVirtualTerminalProcessing))
	}
}