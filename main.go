package main

import (
	"canvas-tui/canvas"
	"canvas-tui/models"
	"canvas-tui/tui"

	"github.com/joho/godotenv"
)

func init() { godotenv.Load() }

// ═══════════════════════════════════════════════════════════════
//  CONFIG — change anything here to customise the app
// ═══════════════════════════════════════════════════════════════

// FullConsole: true  → canvas fills the entire terminal window
//              false → use CanvasW / CanvasH instead
const fullConsole = true
const canvasW     = 120
const canvasH     = 40

// Show coordinate numbers along edges of the canvas
const showRulers = false

// Default tool on startup: "rect" "circle" "line" "free"
const defaultTool = "rect"

// Default fill mode: false = outline, true = filled
const defaultFill = false

// Default color: "green" "cyan" "yellow" "red" "magenta" "blue" "white"
const defaultColor = "green"

// ── Characters used for drawing ──────────────────────────────
// Rectangle outline
const rectTop      = '─'
const rectBottom   = '─'
const rectLeft     = '│'
const rectRight    = '│'
const rectCornerTL = '┌'
const rectCornerTR = '┐'
const rectCornerBL = '└'
const rectCornerBR = '┘'
// Rectangle filled
const rectFill     = '█'

// Circle outline (chosen by tangent angle)
const circleH     = '─'
const circleV     = '│'
const circleDiag1 = '╱'
const circleDiag2 = '╲'
// Circle filled
const circleFill  = '█'

// Line (chosen by slope)
const lineH     = '─'
const lineV     = '│'
const lineDiag1 = '╱'
const lineDiag2 = '╲'

// Freehand paint character
const freeChar = '█'

// ═══════════════════════════════════════════════════════════════
//  ENTRY POINT — no need to edit below
// ═══════════════════════════════════════════════════════════════

func main() {
	tui.Run(tui.Config{
		FullConsole:  fullConsole,
		CanvasW:      canvasW,
		CanvasH:      canvasH,
		ShowRulers:   showRulers,
		DefaultTool:  defaultTool,
		DefaultFill:  defaultFill,
		DefaultColor: defaultColor,
		Chars: canvas.Chars{
			RectTop: rectTop, RectBottom: rectBottom,
			RectLeft: rectLeft, RectRight: rectRight,
			RectTL: rectCornerTL, RectTR: rectCornerTR,
			RectBL: rectCornerBL, RectBR: rectCornerBR,
			RectFill: rectFill,
			CircH: circleH, CircV: circleV,
			CircD1: circleDiag1, CircD2: circleDiag2,
			CircFill: circleFill,
			LineH: lineH, LineV: lineV,
			LineD1: lineDiag1, LineD2: lineDiag2,
			Free: freeChar,
		},
	})
}

// keep models imported for AllColors
var _ = models.AllColors