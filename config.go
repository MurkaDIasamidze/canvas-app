package main

// ════════════════════════════════════════════════════════════
//  CONFIG — the only file you need to edit
// ════════════════════════════════════════════════════════════

// ── Canvas size ──────────────────────────────────────────────
// true  = fill the whole terminal (recommended)
// false = use canvasW / canvasH below
const fullConsole = true

const (
	canvasW = 120
	canvasH = 40
)

// ── Default tool, fill, color ────────────────────────────────
// tool:  KindRect | KindCircle | KindLine | KindFree
// fill:  false = outline   true = solid
// color: ColGreen | ColCyan | ColYellow | ColRed | ColMagenta | ColBlue | ColWhite
const defaultTool  = KindRect
const defaultFill  = false
const defaultColor = ColGreen

// ── Characters ───────────────────────────────────────────────
// Rectangle corners & edges
const rectTL     = '┌'
const rectTR     = '┐'
const rectBL     = '└'
const rectBR     = '┘'
const rectTop    = '─'
const rectBottom = '─'
const rectLeft   = '│'
const rectRight  = '│'

// Fill character used by fillRect / fillCircle / fillTriangle / fill
const rectFill = '█' // alternatives: '▓' '▒' '░' '#' '@'
const circFill = '█'

// Circle outline chars — picked by tangent angle at each pixel
const circH  = '─' // horizontal
const circV  = '│' // vertical
const circD1 = '╱' // diagonal /
const circD2 = '╲' // diagonal \

// Line chars — picked by slope
const lineH  = '─'
const lineV  = '│'
const lineD1 = '╱'
const lineD2 = '╲'

// Freehand (F tool) pixel stamp
const freeChar = '█' // alternatives: '·' '•' '+' '*'

// ── Preset shapes ─────────────────────────────────────────────
// Drawn once at startup as a background layer.

func presetShapes(ctx *Ctx) {
	ctx.strokeStyle = ColCyan
	strokeRect(ctx, 0, 0, 0, 0)

	ctx.fillStyle = ColYellow
	fillCircle(ctx, 0, 0, 0)

	ctx.strokeStyle = ColRed
	drawLine(ctx, 0, 0, 0, 0)

	ctx.fillStyle = ColGreen
	fillText(ctx, "", 0, 0)
}