package main

import (
	"canvas-tui/tui"

	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load() // loads .env if present, silently ignores if not
}

func main() {
	tui.Run()
}