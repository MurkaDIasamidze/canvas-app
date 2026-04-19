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

// ── Windows API ───────────────────────────────────────────────

var (
	kernel32               = syscall.NewLazyDLL("kernel32.dll")
	procGetStdHandle       = kernel32.NewProc("GetStdHandle")
	procGetConsoleMode     = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode     = kernel32.NewProc("SetConsoleMode")
	procSetConsoleOutputCP = kernel32.NewProc("SetConsoleOutputCP")
	procReadConsoleInput   = kernel32.NewProc("ReadConsoleInputW")
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
)

const (
	stdOut = ^uintptr(10)
	stdIn  = ^uintptr(9)
	vtFlag = 0x0004
	utf8CP = 65001
)

type consoleScreenBufferInfo struct {
	dwSizeX, dwSizeY           int16
	dwCursorPositionX, dwCursorPositionY int16
	wAttributes                uint16
	srWindowLeft, srWindowTop  int16
	srWindowRight, srWindowBottom int16
	dwMaximumWindowSizeX, dwMaximumWindowSizeY int16
}

// TermSize returns the current terminal width and height.
func TermSize() (w, h int) {
	hOut, _, _ := procGetStdHandle.Call(stdOut)
	var info consoleScreenBufferInfo
	procGetConsoleScreenBufferInfo.Call(hOut, uintptr(unsafe.Pointer(&info)))
	w = int(info.srWindowRight-info.srWindowLeft) + 1
	h = int(info.srWindowBottom-info.srWindowTop) + 1
	if w <= 0 { w = 80 }
	if h <= 0 { h = 24 }
	return
}

func InitTerminal() {
	procSetConsoleOutputCP.Call(uintptr(utf8CP))
	hOut, _, _ := procGetStdHandle.Call(stdOut)
	if hOut != 0 {
		var mode uint32
		procGetConsoleMode.Call(hOut, uintptr(unsafe.Pointer(&mode)))
		procSetConsoleMode.Call(hOut, uintptr(mode|vtFlag))
	}
}

func initRawMode()    {}
func restoreTerminal() {}

// ── Key constants ─────────────────────────────────────────────

const (
	KeyUp    = 1000; KeyDown  = 1001
	KeyLeft  = 1002; KeyRight = 1003
	KeyEnter = 13;   KeyEsc   = 27
	KeyBack  = 8;    KeySpace = 32
	KeyTab   = 9;    KeyDel   = 1004
)

// ── Input ─────────────────────────────────────────────────────

type inputRecord struct {
	EventType uint16; _ [2]byte; Event [16]byte
}
type keyEvent struct {
	KeyDown int32; RepeatCount, VKCode, ScanCode, UnicodeChar uint16
	ControlKeys uint32
}

func ReadKey() int {
	hIn, _, _ := procGetStdHandle.Call(stdIn)
	for {
		var rec inputRecord; var n uint32
		procReadConsoleInput.Call(hIn, uintptr(unsafe.Pointer(&rec)), 1, uintptr(unsafe.Pointer(&n)))
		if rec.EventType != 1 { continue }
		ke := (*keyEvent)(unsafe.Pointer(&rec.Event[0]))
		if ke.KeyDown == 0 { continue }
		switch ke.VKCode {
		case 0x26: return KeyUp;    case 0x28: return KeyDown
		case 0x25: return KeyLeft;  case 0x27: return KeyRight
		case 0x0D: return KeyEnter; case 0x1B: return KeyEsc
		case 0x08: return KeyBack;  case 0x20: return KeySpace
		case 0x09: return KeyTab;   case 0x2E: return KeyDel
		}
		if ke.UnicodeChar >= 32 && ke.UnicodeChar < 127 {
			return int(ke.UnicodeChar)
		}
	}
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