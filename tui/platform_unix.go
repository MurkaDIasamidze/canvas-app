//go:build !windows

package tui

import (
	"bufio"
	"canvas-tui/models"
	"fmt"
	"os"
	"os/exec"
)

// ── Terminal init ─────────────────────────────────────────────────────────────

func InitTerminal() {}

func initRawMode() {
	exec.Command("stty", "-F", "/dev/tty", "raw", "-echo").Run()
}

func restoreTerminal() {
	exec.Command("stty", "-F", "/dev/tty", "sane").Run()
}

// ── Key constants ─────────────────────────────────────────────────────────────

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
)

// ── Input ─────────────────────────────────────────────────────────────────────

func ReadKey() int {
	buf := make([]byte, 8)
	n, _ := os.Stdin.Read(buf)
	if n == 0 { return 0 }
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
	if buf[0] >= 32 && buf[0] < 127 { return int(buf[0]) }
	return int(buf[0])
}

// ── ANSI constants ────────────────────────────────────────────────────────────

const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	FgBlack   = "\033[30m"
	FgRed     = "\033[31m"
	FgGreen   = "\033[32m"
	FgYellow  = "\033[33m"
	FgBlue    = "\033[34m"
	FgMagenta = "\033[35m"
	FgCyan    = "\033[36m"
	FgWhite   = "\033[37m"
	FgBrightRed     = "\033[91m"
	FgBrightGreen   = "\033[92m"
	FgBrightYellow  = "\033[93m"
	FgBrightBlue    = "\033[94m"
	FgBrightMagenta = "\033[95m"
	FgBrightCyan    = "\033[96m"
	FgBrightWhite   = "\033[97m"
	BgBlack   = "\033[40m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"
)

// ── Terminal output ───────────────────────────────────────────────────────────

var out = bufio.NewWriterSize(os.Stdout, 1<<17)

func Flush()  { out.Flush() }
func MoveTo(x, y int) { fmt.Fprintf(out, "\033[%d;%dH", y, x) }
func Print(s string)  { fmt.Fprint(out, s) }
func Printf(f string, a ...interface{}) { fmt.Fprintf(out, f, a...) }
func PrintAt(x, y int, s string) { MoveTo(x, y); fmt.Fprint(out, s) }
func PrintfAt(x, y int, f string, a ...interface{}) { MoveTo(x, y); fmt.Fprintf(out, f, a...) }
func ClearScreen() { fmt.Fprint(out, "\033[2J\033[H") }
func HideCursor()  { fmt.Fprint(out, "\033[?25l") }
func ShowCursor()  { fmt.Fprint(out, "\033[?25h") }
func EraseDown()   { fmt.Fprint(out, "\033[J") }
func Repeat(s string, n int) string {
	r := ""
	for i := 0; i < n; i++ { r += s }
	return r
}

// ── Color helpers ─────────────────────────────────────────────────────────────

func ColorANSI(c models.ColorName) string {
	switch c {
	case models.ColorGreen:   return FgBrightGreen
	case models.ColorCyan:    return FgBrightCyan
	case models.ColorYellow:  return FgBrightYellow
	case models.ColorRed:     return FgBrightRed
	case models.ColorMagenta: return FgBrightMagenta
	case models.ColorBlue:    return FgBrightBlue
	case models.ColorWhite:   return FgBrightWhite
	default:                  return FgBrightGreen
	}
}