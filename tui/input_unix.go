//go:build !windows

package tui

import (
	"os"
	"os/exec"
)

const (
	KeyArrowUp    = 1000
	KeyArrowDown  = 1001
	KeyArrowLeft  = 1002
	KeyArrowRight = 1003
	KeyEnter      = 13
	KeyEscape     = 27
	KeyBackspace  = 127
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

// InitRawMode puts the terminal into raw mode using stty (works on Linux + macOS)
func InitRawMode() {
	exec.Command("stty", "-F", "/dev/tty", "raw", "-echo").Run()
}

// RestoreTerminal restores the terminal using stty
func RestoreTerminal() {
	exec.Command("stty", "-F", "/dev/tty", "sane").Run()
}

// ReadKey reads a single keypress from stdin
func ReadKey() int {
	buf := make([]byte, 8)
	n, _ := os.Stdin.Read(buf)
	if n == 0 {
		return 0
	}
	// ESC sequence
	if buf[0] == 27 && n > 2 && buf[1] == '[' {
		switch buf[2] {
		case 'A': return KeyArrowUp
		case 'B': return KeyArrowDown
		case 'C': return KeyArrowRight
		case 'D': return KeyArrowLeft
		case 'H': return KeyHome
		case 'F': return KeyEnd
		case '3':
			if n > 3 && buf[3] == '~' { return KeyDelete }
		case '5':
			if n > 3 && buf[3] == '~' { return KeyPageUp }
		case '6':
			if n > 3 && buf[3] == '~' { return KeyPageDown }
		}
	}
	if buf[0] == 27 && n == 1 { return KeyEscape }
	switch buf[0] {
	case '\r', '\n': return KeyEnter
	case 127:        return KeyBackspace
	case '\t':       return KeyTab
	case ' ':        return KeySpace
	}
	if buf[0] >= 32 && buf[0] < 127 {
		return int(buf[0])
	}
	return int(buf[0])
}