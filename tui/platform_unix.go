//go:build !windows

package tui

import (
	"bufio"
	"canvas-tui/models"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// TermSize returns current terminal dimensions using stty.
func TermSize() (w, h int) {
	out, err := exec.Command("stty", "size").Output()
	if err == nil {
		parts := strings.Fields(strings.TrimSpace(string(out)))
		if len(parts) == 2 {
			h, _ = strconv.Atoi(parts[0])
			w, _ = strconv.Atoi(parts[1])
		}
	}
	if w <= 0 { w = 80 }
	if h <= 0 { h = 24 }
	return
}

func InitTerminal() {}

func initRawMode()    { exec.Command("stty", "-F", "/dev/tty", "raw", "-echo").Run() }
func restoreTerminal() { exec.Command("stty", "-F", "/dev/tty", "sane").Run() }

// ── Key constants ─────────────────────────────────────────────

const (
	KeyUp    = 1000; KeyDown  = 1001
	KeyLeft  = 1002; KeyRight = 1003
	KeyEnter = 13;   KeyEsc   = 27
	KeyBack  = 127;  KeySpace = 32
	KeyTab   = 9;    KeyDel   = 1004
)

func ReadKey() int {
	buf := make([]byte, 8)
	n, _ := os.Stdin.Read(buf)
	if n == 0 { return 0 }
	if buf[0] == 27 && n > 2 && buf[1] == '[' {
		switch buf[2] {
		case 'A': return KeyUp;   case 'B': return KeyDown
		case 'C': return KeyRight; case 'D': return KeyLeft
		}
	}
	if buf[0] == 27 && n == 1 { return KeyEsc }
	switch buf[0] {
	case '\r', '\n': return KeyEnter
	case 127:        return KeyBack
	case ' ':        return KeySpace
	case '\t':       return KeyTab
	}
	if buf[0] >= 32 && buf[0] < 127 { return int(buf[0]) }
	return int(buf[0])
}

// ── ANSI / terminal output ────────────────────────────────────

const (
	Reset = "\033[0m"; Bold = "\033[1m"; Dim = "\033[2m"
	FgBlack = "\033[30m"; FgRed = "\033[31m"; FgGreen = "\033[32m"
	FgYellow = "\033[33m"; FgBlue = "\033[34m"; FgMagenta = "\033[35m"
	FgCyan = "\033[36m"; FgWhite = "\033[37m"
	FgBrightRed = "\033[91m"; FgBrightGreen = "\033[92m"
	FgBrightYellow = "\033[93m"; FgBrightBlue = "\033[94m"
	FgBrightMagenta = "\033[95m"; FgBrightCyan = "\033[96m"
	FgBrightWhite = "\033[97m"
	BgBlack = "\033[40m"; BgRed = "\033[41m"; BgGreen = "\033[42m"
	BgYellow = "\033[43m"; BgBlue = "\033[44m"; BgMagenta = "\033[45m"
	BgCyan = "\033[46m"; BgWhite = "\033[47m"
)

func ColorANSI(c models.ColorName) string {
	switch c {
	case models.ColGreen:   return FgBrightGreen
	case models.ColCyan:    return FgBrightCyan
	case models.ColYellow:  return FgBrightYellow
	case models.ColRed:     return FgBrightRed
	case models.ColMagenta: return FgBrightMagenta
	case models.ColBlue:    return FgBrightBlue
	case models.ColWhite:   return FgBrightWhite
	}
	return FgBrightGreen
}

var out = bufio.NewWriterSize(os.Stdout, 1<<18)

func Flush()                              { out.Flush() }
func MoveTo(x, y int)                     { fmt.Fprintf(out, "\033[%d;%dH", y, x) }
func Print(s string)                      { fmt.Fprint(out, s) }
func Printf(f string, a ...interface{})   { fmt.Fprintf(out, f, a...) }
func PrintAt(x, y int, s string)          { MoveTo(x, y); fmt.Fprint(out, s) }
func PrintfAt(x, y int, f string, a ...interface{}) { MoveTo(x, y); fmt.Fprintf(out, f, a...) }
func ClearScreen()                        { fmt.Fprint(out, "\033[2J\033[H") }
func HideCursor()                         { fmt.Fprint(out, "\033[?25l") }
func ShowCursor()                         { fmt.Fprint(out, "\033[?25h") }
func ClearLine()                          { fmt.Fprint(out, "\033[2K") }
func EraseDown()                          { fmt.Fprint(out, "\033[J") }
func Repeat(s string, n int) string       { r := ""; for i := 0; i < n; i++ { r += s }; return r }