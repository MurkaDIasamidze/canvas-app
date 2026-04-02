//go:build !windows

package tui

func initRawMode() {
	InitRawMode()
}

func restoreTerminal() {
	RestoreTerminal()
}