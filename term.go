package main

// term.go — shared terminal types, ANSI escape codes, buffered output.
// Platform-specific raw mode / input / size: term_unix.go & term_windows.go

import (
	"bufio"
	"fmt"
	"os"
)

// ── Color ─────────────────────────────────────────────────────

type Color string

const (
	ColGreen   Color = "green"
	ColCyan    Color = "cyan"
	ColYellow  Color = "yellow"
	ColRed     Color = "red"
	ColMagenta Color = "magenta"
	ColBlue    Color = "blue"
	ColWhite   Color = "white"
)

var allColors = []Color{
	ColGreen, ColCyan, ColYellow, ColRed, ColMagenta, ColBlue, ColWhite,
}

func colorANSI(c Color) string {
	switch c {
	case ColGreen:   return FgBrightGreen
	case ColCyan:    return FgBrightCyan
	case ColYellow:  return FgBrightYellow
	case ColRed:     return FgBrightRed
	case ColMagenta: return FgBrightMagenta
	case ColBlue:    return FgBrightBlue
	case ColWhite:   return FgBrightWhite
	}
	return FgBrightGreen
}

// ky wraps a key hint in bright yellow for status bars.
func ky(s string) string { return FgBrightYellow + s + Reset }

// ── ANSI escape codes ─────────────────────────────────────────

const (
	Reset   = "\033[0m"
	Bold    = "\033[1m"
	Dim     = "\033[2m"

	FgBlack         = "\033[30m"
	FgRed           = "\033[31m"
	FgGreen         = "\033[32m"
	FgYellow        = "\033[33m"
	FgBlue          = "\033[34m"
	FgMagenta       = "\033[35m"
	FgCyan          = "\033[36m"
	FgWhite         = "\033[37m"
	FgBrightRed     = "\033[91m"
	FgBrightGreen   = "\033[92m"
	FgBrightYellow  = "\033[93m"
	FgBrightBlue    = "\033[94m"
	FgBrightMagenta = "\033[95m"
	FgBrightCyan    = "\033[96m"
	FgBrightWhite   = "\033[97m"

	BgBlack  = "\033[40m"
	BgBlue   = "\033[44m"
	BgGreen  = "\033[42m"
	BgYellow = "\033[43m"
)

// ── Key codes ─────────────────────────────────────────────────

const (
	KeyUp    = 1000
	KeyDown  = 1001
	KeyLeft  = 1002
	KeyRight = 1003
	KeyEnter = 13
	KeyEsc   = 27
	KeyBack  = 127
	KeySpace = 32
)

// ── Buffered stdout ───────────────────────────────────────────

var stdout = bufio.NewWriterSize(os.Stdout, 1<<18)

func flush()                     { stdout.Flush() }
func moveTo(x, y int)            { fmt.Fprintf(stdout, "\033[%d;%dH", y, x) }
func print_(s string)            { fmt.Fprint(stdout, s) }
func printAt(x, y int, s string) { moveTo(x, y); fmt.Fprint(stdout, s) }
func clearScreen()               { fmt.Fprint(stdout, "\033[2J\033[H") }
func hideCursor()                { fmt.Fprint(stdout, "\033[?25l") }
func showCursor()                { fmt.Fprint(stdout, "\033[?25h") }
func eraseDown()                 { fmt.Fprint(stdout, "\033[J") }
func eraseLineRight()            { fmt.Fprint(stdout, "\033[K") }

func repeatStr(s string, n int) string {
	r := ""
	for i := 0; i < n; i++ { r += s }
	return r
}