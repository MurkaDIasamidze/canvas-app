//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

var (
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	procGetStdHandle               = kernel32.NewProc("GetStdHandle")
	procGetConsoleMode             = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode             = kernel32.NewProc("SetConsoleMode")
	procSetConsoleOutputCP         = kernel32.NewProc("SetConsoleOutputCP")
	procReadConsoleInput           = kernel32.NewProc("ReadConsoleInputW")
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
)

const (
	stdOutputHandle = ^uintptr(10) // -11 as uintptr
	stdInputHandle  = ^uintptr(9)  // -10 as uintptr
)

type smallRect struct{ Left, Top, Right, Bottom int16 }
type coord struct{ X, Y int16 }
type consoleScreenBufferInfo struct {
	Size              coord
	CursorPosition    coord
	Attributes        uint16
	Window            smallRect
	MaximumWindowSize coord
}

func termSize() (w, h int) {
	hOut, _, _ := procGetStdHandle.Call(stdOutputHandle)
	var info consoleScreenBufferInfo
	procGetConsoleScreenBufferInfo.Call(hOut, uintptr(unsafe.Pointer(&info)))
	w = int(info.Window.Right-info.Window.Left) + 1
	h = int(info.Window.Bottom-info.Window.Top) + 1
	if w <= 0 { w = 80 }
	if h <= 0 { h = 24 }
	return
}

func initRaw() {
	// UTF-8 output
	procSetConsoleOutputCP.Call(65001)

	// Enable VT100 escape processing on stdout
	hOut, _, _ := procGetStdHandle.Call(stdOutputHandle)
	var modeOut uint32
	procGetConsoleMode.Call(hOut, uintptr(unsafe.Pointer(&modeOut)))
	procSetConsoleMode.Call(hOut, uintptr(modeOut|0x0004)) // ENABLE_VIRTUAL_TERMINAL_PROCESSING

	// Disable line-input and echo on stdin
	hIn, _, _ := procGetStdHandle.Call(stdInputHandle)
	var modeIn uint32
	procGetConsoleMode.Call(hIn, uintptr(unsafe.Pointer(&modeIn)))
	procSetConsoleMode.Call(hIn, uintptr(modeIn&^uint32(0x0006))) // clear ENABLE_LINE_INPUT | ENABLE_ECHO_INPUT
}

var savedModeIn uint32

func restoreRaw() {
	hIn, _, _ := procGetStdHandle.Call(stdInputHandle)
	procSetConsoleMode.Call(hIn, uintptr(0x0007)) // ENABLE_PROCESSED_INPUT | ENABLE_LINE_INPUT | ENABLE_ECHO_INPUT
}

// inputRecord and keyEventRecord match the Windows INPUT_RECORD / KEY_EVENT_RECORD structs.
type inputRecord struct {
	EventType uint16
	_         [2]byte
	Event     [16]byte
}
type keyEventRecord struct {
	KeyDown         int32
	RepeatCount     uint16
	VirtualKeyCode  uint16
	VirtualScanCode uint16
	UnicodeChar     uint16
	ControlKeyState uint32
}

func readKey() int {
	hIn, _, _ := procGetStdHandle.Call(stdInputHandle)
	for {
		var rec inputRecord
		var n uint32
		procReadConsoleInput.Call(
			hIn,
			uintptr(unsafe.Pointer(&rec)),
			1,
			uintptr(unsafe.Pointer(&n)),
		)
		// only handle KEY_EVENT (type 1) that are key-down
		if rec.EventType != 1 { continue }
		ke := (*keyEventRecord)(unsafe.Pointer(&rec.Event[0]))
		if ke.KeyDown == 0 { continue }

		switch ke.VirtualKeyCode {
		case 0x26: return KeyUp
		case 0x28: return KeyDown
		case 0x25: return KeyLeft
		case 0x27: return KeyRight
		case 0x0D: return KeyEnter
		case 0x1B: return KeyEsc
		case 0x08: return KeyBack
		case 0x20: return KeySpace
		}
		if ke.UnicodeChar >= 32 && ke.UnicodeChar < 127 {
			return int(ke.UnicodeChar)
		}
	}
}