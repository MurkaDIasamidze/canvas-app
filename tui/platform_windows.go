//go:build windows

package tui

import (
	"bufio"
	"canvas-tui/models"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

// ── Windows kernel32 ──────────────────────────────────────────────────────────

var (
	kernel32               = syscall.NewLazyDLL("kernel32.dll")
	procGetStdHandle       = kernel32.NewProc("GetStdHandle")
	procGetConsoleMode     = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode     = kernel32.NewProc("SetConsoleMode")
	procSetConsoleOutputCP = kernel32.NewProc("SetConsoleOutputCP")
	procReadConsoleInput   = kernel32.NewProc("ReadConsoleInputW")
)

const (
	stdOutputHandle                 = ^uintptr(10)
	stdInputHandle                  = ^uintptr(9)
	enableVirtualTerminalProcessing = 0x0004
	cpUTF8                          = 65001
)

// InitTerminal enables ANSI colors and UTF-8 on Windows
func InitTerminal() {
	procSetConsoleOutputCP.Call(uintptr(cpUTF8))
	hOut, _, _ := procGetStdHandle.Call(stdOutputHandle)
	if hOut != 0 {
		var mode uint32
		procGetConsoleMode.Call(hOut, uintptr(unsafe.Pointer(&mode)))
		procSetConsoleMode.Call(hOut, uintptr(mode|enableVirtualTerminalProcessing))
	}
}

func initRawMode()    {}
func restoreTerminal() {}

// ── Key constants ─────────────────────────────────────────────────────────────

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
)

// ── Input ─────────────────────────────────────────────────────────────────────

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

func ReadKey() int {
	hIn, _, _ := procGetStdHandle.Call(stdInputHandle)
	for {
		var rec inputRecord
		var n uint32
		procReadConsoleInput.Call(hIn, uintptr(unsafe.Pointer(&rec)), 1, uintptr(unsafe.Pointer(&n)))
		if rec.EventType != 1 { continue }
		ke := (*keyEventRecord)(unsafe.Pointer(&rec.Event[0]))
		if ke.KeyDown == 0 { continue }
		switch ke.VirtualKeyCode {
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
		}
		if ke.UnicodeChar >= 32 && ke.UnicodeChar < 127 {
			return int(ke.UnicodeChar)
		}
	}
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