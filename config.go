package main

// config.go — every setting you might want to change, all in one place.
//
// ┌─────────────────────────────────────────────────────────────┐
// │  Edit this file to customise the canvas without touching    │
// │  any other source file.                                     │
// └─────────────────────────────────────────────────────────────┘

// ── Canvas size ───────────────────────────────────────────────
//
// fullConsole = true  → canvas fills your entire terminal window (recommended)
// fullConsole = false → use the fixed canvasW / canvasH values below
const fullConsole = true

// Fixed size used only when fullConsole = false.
const (
	canvasW = 120
	canvasH = 140
)

// ── Starting tool ─────────────────────────────────────────────
//
// One of: KindRect  KindCircle  KindLine  KindFree
const defaultTool = KindRect

// ── Starting fill mode ────────────────────────────────────────
//
// false = outline only   true = filled solid
const defaultFill = false

// ── Starting color ────────────────────────────────────────────
//
// One of: ColGreen ColCyan ColYellow ColRed ColMagenta ColBlue ColWhite
const defaultColor = ColGreen

// ── Rectangle outline characters ──────────────────────────────
//
// These four corners and four edges make up every StrokeRect call.
// Swap them for ASCII if your terminal doesn't support Unicode:
//
//   rectTL, rectTR, rectBL, rectBR = '+', '+', '+', '+'
//   rectTop, rectBottom            = '-', '-'
//   rectLeft, rectRight            = '|', '|'
const (
	rectTL     = '┌' // top-left corner
	rectTR     = '┐' // top-right corner
	rectBL     = '└' // bottom-left corner
	rectBR     = '┘' // bottom-right corner
	rectTop    = '─' // top edge
	rectBottom = '─' // bottom edge
	rectLeft   = '│' // left edge
	rectRight  = '│' // right edge
)

// ── Rectangle fill character ──────────────────────────────────
const rectFill = '█'

// ── Circle outline characters ─────────────────────────────────
//
// Chosen by tangent angle at each pixel.
// circH  → nearly horizontal segments  (0° / 180°)
// circV  → nearly vertical segments    (90°)
// circD1 → forward-slash diagonal      (45°)
// circD2 → back-slash diagonal         (135°)
const (
	circH  = '─'
	circV  = '│'
	circD1 = '╱'
	circD2 = '╲'
)

// ── Circle fill character ─────────────────────────────────────
const circFill = '█'

// ── Line characters ───────────────────────────────────────────
//
// Chosen automatically by the line's slope.
const (
	lineH  = '─' // horizontal
	lineV  = '│' // vertical
	lineD1 = '╱' // diagonal top-right → bottom-left
	lineD2 = '╲' // diagonal top-left → bottom-right
)

// ── Freehand / pixel character ────────────────────────────────
//
// Stamped at the cursor for every Space press in Freehand mode.
const freeChar = '█'

// ── Preset shapes ─────────────────────────────────────────────
//
// presetShapes is called once at startup with a fresh canvas.
// Anything you draw here appears as a permanent background layer
// behind all interactively drawn shapes.
//
// Use the full JS-canvas API — examples:
//
//	c.StrokeColor = ColCyan
//	c.StrokeRect(2, 2, 20, 10)
//
//	c.FillColor = ColYellow
//	c.FillCircle(50, 15, 8)
//
//	c.StrokeColor = ColRed
//	c.DrawLine(0, 0, 30, 15)
//
//	c.FillColor = ColWhite
//	c.FillText("hello world", 5, 5)
//
//	c.BeginPath()
//	c.MoveTo(0, 0)
//	c.LineTo(40, 20)
//	c.LineTo(0, 20)
//	c.ClosePath()
//	c.Stroke()
func presetShapes(c *Canvas) {
	// empty by default — add your draw calls here
	_ = c
}