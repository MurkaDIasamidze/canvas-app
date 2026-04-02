//go:build !windows

package tui

// InitTerminal is a no-op on Unix/macOS — terminals support ANSI natively
func InitTerminal() {}