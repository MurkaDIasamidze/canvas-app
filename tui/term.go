package tui

import (
	"bufio"
	"fmt"
	"os"
)

// ANSI escape codes
const (
	Reset      = "\033[0m"
	Bold       = "\033[1m"
	Dim        = "\033[2m"
	Underline  = "\033[4m"

	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	BgBlack   = "\033[40m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"

	BrightBlack   = "\033[90m"
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"
)

var writer = bufio.NewWriterSize(os.Stdout, 1<<16) // 64KB buffer for fast redraws

// Flush flushes the output buffer to screen
func Flush() {
	writer.Flush()
}

// Clear clears the screen and moves cursor to 1,1
func Clear() {
	fmt.Fprint(writer, "\033[2J\033[H")
}

// ClearLine clears from cursor to end of line
func ClearLine() {
	fmt.Fprint(writer, "\033[2K")
}

// MoveCursor moves the cursor to col, row (1-based)
func MoveCursor(col, row int) {
	fmt.Fprintf(writer, "\033[%d;%dH", row, col)
}

// HideCursor hides the terminal cursor
func HideCursor() {
	fmt.Fprint(writer, "\033[?25l")
}

// ShowCursor shows the terminal cursor
func ShowCursor() {
	fmt.Fprint(writer, "\033[?25h")
}

// SaveCursor saves cursor position
func SaveCursor() {
	fmt.Fprint(writer, "\033[s")
}

// RestoreCursor restores saved cursor position
func RestoreCursor() {
	fmt.Fprint(writer, "\033[u")
}

// Print writes a string to the buffer
func Print(s string) {
	fmt.Fprint(writer, s)
}

// Printf writes formatted string to the buffer
func Printf(format string, args ...interface{}) {
	fmt.Fprintf(writer, format, args...)
}

// PrintAt prints text at a specific position
func PrintAt(col, row int, s string) {
	MoveCursor(col, row)
	fmt.Fprint(writer, s)
}

// PrintAtf prints formatted text at a specific position
func PrintAtf(col, row int, format string, args ...interface{}) {
	MoveCursor(col, row)
	fmt.Fprintf(writer, format, args...)
}

// Box draws a box border
func Box(x, y, w, h int, color string) {
	// Top border
	PrintAt(x, y, color+"┌"+repeat("─", w-2)+"┐"+Reset)
	// Sides
	for row := 1; row < h-1; row++ {
		PrintAt(x, y+row, color+"│"+Reset)
		PrintAt(x+w-1, y+row, color+"│"+Reset)
	}
	// Bottom border
	PrintAt(x, y+h-1, color+"└"+repeat("─", w-2)+"┘"+Reset)
}

// FillBox fills a box with a background color
func FillBox(x, y, w, h int, bg string) {
	line := bg + spaces(w) + Reset
	for row := 0; row < h; row++ {
		PrintAt(x, y+row, line)
	}
}

func repeat(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

func spaces(n int) string {
	return repeat(" ", n)
}