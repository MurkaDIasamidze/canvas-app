//go:build windows

package tui

import (
	"unsafe"
)

// Key constants
const (
	KeyArrowUp    = 1000
	KeyArrowDown  = 1001
	KeyArrowLeft  = 1002
	KeyArrowRight = 1003
	KeyEnter      = 13
	KeyEscape     = 27
	KeyBackspace  = 8
	KeyTab        = 9
	KeySpace      = 32
	KeyDelete     = 1004
	KeyHome       = 1005
	KeyEnd        = 1006
	KeyPageUp     = 1007
	KeyPageDown   = 1008
	KeyF1         = 1009
	KeyF2         = 1010
	KeyF5         = 1013
)

var procReadConsoleInput = kernel32.NewProc("ReadConsoleInputW")

// inputRecord matches the Windows INPUT_RECORD structure
type inputRecord struct {
	EventType uint16
	_         [2]byte
	Event     [16]byte
}

// keyEventRecord matches the Windows KEY_EVENT_RECORD structure
type keyEventRecord struct {
	KeyDown         int32
	RepeatCount     uint16
	VirtualKeyCode  uint16
	VirtualScanCode uint16
	UnicodeChar     uint16
	ControlKeyState uint32
}

// ReadKey reads a single key from the Windows console (blocking).
// Uses ReadConsoleInputW so it captures arrow keys and special keys correctly.
func ReadKey() int {
	hIn, _, _ := procGetStdHandle.Call(stdInputHandle)

	for {
		var rec inputRecord
		var numRead uint32
		procReadConsoleInput.Call(
			hIn,
			uintptr(unsafe.Pointer(&rec)),
			1,
			uintptr(unsafe.Pointer(&numRead)),
		)

		// Only process KEY_EVENT (EventType == 1), key-down only
		if rec.EventType != 1 {
			continue
		}
		keyEvent := (*keyEventRecord)(unsafe.Pointer(&rec.Event[0]))
		if keyEvent.KeyDown == 0 {
			continue
		}

		vk := keyEvent.VirtualKeyCode
		ch := keyEvent.UnicodeChar

		switch vk {
		case 0x26: return KeyArrowUp
		case 0x28: return KeyArrowDown
		case 0x25: return KeyArrowLeft
		case 0x27: return KeyArrowRight
		case 0x0D: return KeyEnter
		case 0x1B: return KeyEscape
		case 0x08: return KeyBackspace
		case 0x09: return KeyTab
		case 0x20: return KeySpace
		case 0x2E: return KeyDelete
		case 0x24: return KeyHome
		case 0x23: return KeyEnd
		case 0x21: return KeyPageUp
		case 0x22: return KeyPageDown
		case 0x70: return KeyF1
		case 0x71: return KeyF2
		case 0x74: return KeyF5
		}

		if ch >= 32 && ch < 127 {
			return int(ch)
		}
	}
}