package main

import (
	"canvas-tui/canvas"
	"canvas-tui/models"
	"canvas-tui/tui"

	"github.com/joho/godotenv"
)

func init() { godotenv.Load() }

// ╔══════════════════════════════════════════════════════════════════╗
// ║                        CANVAS SETTINGS                          ║
// ╚══════════════════════════════════════════════════════════════════╝

// ── Size ──────────────────────────────────────────────────────────────
//
//	fullConsole = true  → canvas fills your entire terminal window
//	fullConsole = false → use the fixed width/height below
const fullConsole = true
const canvasW     = 120
const canvasH     = 40

// ── Starting tool ─────────────────────────────────────────────────────
//	"rect"    rectangle
//	"circle"  circle
//	"line"    straight line
//	"free"    freehand pixel paint
const defaultTool = "rect"

// ── Starting fill mode ────────────────────────────────────────────────
//	false = outline only
//	true  = filled solid
const defaultFill = false

// ── Starting color ────────────────────────────────────────────────────
//	"green"  "cyan"  "yellow"  "red"  "magenta"  "blue"  "white"
const defaultColor = "green"


// ── Rectangle outline ─────────────────────────────────────────────────
const (
	rectCornerTL = '┌' // top-left corner
	rectCornerTR = '┐' // top-right corner
	rectCornerBL = '└' // bottom-left corner
	rectCornerBR = '┘' // bottom-right corner
	rectTop      = '─' // top edge
	rectBottom   = '─' // bottom edge
	rectLeft     = '│' // left edge
	rectRight    = '│' // right edge
)

// ── Rectangle filled ──────────────────────────────────────────────────
const rectFill = '█'

// ── Circle outline ────────────────────────────────────────────────────
const (
	circleH     = '─' // horizontal segments
	circleV     = '│' // vertical segments
	circleDiag1 = '╱' // forward-slash diagonal
	circleDiag2 = '╲' // back-slash diagonal
)

// ── Circle filled ─────────────────────────────────────────────────────
const circleFill = '█'

// ── Line ──────────────────────────────────────────────────────────────
const (
	lineH     = '─'
	lineV     = '│'
	lineDiag1 = '╱'
	lineDiag2 = '╲'
)

// ── Freehand pixel ────────────────────────────────────────────────────
//   Character stamped when using the Freehand (F) tool
const freeChar = '█'


//
// Example — uncomment to try:
//
	// func presetShapes(ctx *canvas.Context) {
	//     ctx.StrokeColor = canvas.Cyan
	//     ctx.StrokeRect(2, 2, 20, 10)       // outline box at x=2, y=2, w=20, h=10

	//     ctx.FillColor = canvas.Yellow
	//     ctx.FillCircle(50, 15, 8)           // filled circle at center (50,15) r=8

	//     ctx.StrokeColor = canvas.Red
	//     ctx.DrawLine(0, 0, 30, 15)          // diagonal line

	//     ctx.StrokeColor = canvas.Green
	//     ctx.DrawLine(10, 0, 10, 20)         // vertical line at x=10
	// }

func presetShapes(_ *canvas.Context) {
	// empty by default — interactive drawing only
	// replace _ with ctx and add your draw calls above
}

func main() {
	tui.Run(tui.Config{
		FullConsole:   fullConsole,
		CanvasW:       canvasW,
		CanvasH:       canvasH,
		DefaultTool:   defaultTool,
		DefaultFill:   defaultFill,
		DefaultColor:  defaultColor,
		PresetShapes:  presetShapes,
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

var _ = models.AllColors // keep import