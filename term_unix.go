//go:build !windows

package main

import (
	"fmt"
	"os"
	"os/exec"
)

func termSize() (w, h int) {
	out, err := exec.Command("stty", "size").Output()
	if err == nil {
		var hh, ww int
		fmt.Sscanf(string(out), "%d %d", &hh, &ww)
		if ww > 0 && hh > 0 { return ww, hh }
	}
	return 80, 24
}

func initRaw()    { exec.Command("stty", "-F", "/dev/tty", "raw", "-echo").Run() }
func restoreRaw() { exec.Command("stty", "-F", "/dev/tty", "sane").Run() }

func readKey() int {
	buf := make([]byte, 8)
	n, _ := os.Stdin.Read(buf)
	if n == 0 { return 0 }
	// arrow keys arrive as ESC [ A/B/C/D
	if buf[0] == 27 && n > 2 && buf[1] == '[' {
		switch buf[2] {
		case 'A': return KeyUp
		case 'B': return KeyDown
		case 'C': return KeyRight
		case 'D': return KeyLeft
		}
	}
	if buf[0] == 27 && n == 1 { return KeyEsc }
	switch buf[0] {
	case '\r', '\n': return KeyEnter
	case 127:        return KeyBack
	case ' ':        return KeySpace
	}
	if buf[0] >= 32 && buf[0] < 127 { return int(buf[0]) }
	return int(buf[0])
}