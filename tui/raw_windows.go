//go:build windows

package tui

func initRawMode()      {}  // Windows uses ReadConsoleInput, no raw mode needed
func restoreTerminal()  {}