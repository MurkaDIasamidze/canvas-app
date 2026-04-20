// Package canvas provides a JS-canvas-style drawing API for the terminal.
//
// Usage — just like JavaScript canvas:
//
//	ctx := canvas.New(80, 40)          // create a canvas
//	ctx.StrokeColor = canvas.Green     // set color
//	ctx.FillColor   = canvas.Red
//
//	ctx.StrokeRect(5, 5, 30, 10)       // outline rectangle (x, y, width, height)
//	ctx.FillRect(5, 5, 30, 10)         // filled rectangle
//	ctx.StrokeCircle(40, 20, 8)        // outline circle (cx, cy, radius)
//	ctx.FillCircle(40, 20, 8)          // filled circle
//	ctx.DrawLine(0, 0, 79, 39)         // line from point to point
//	ctx.SetPixel(10, 10)               // single dot
//	ctx.ClearRect(5, 5, 30, 10)        // erase a region
//	ctx.Clear()                        // erase everything
//
// All coordinates are (x, y) from top-left, just like JS canvas.
// Width and height are in terminal character cells.
package canvas

import (
	"canvas-tui/models"
	"math"
)

// ── Colors ────────────────────────────────────────────────────

const (
	Green   = models.ColGreen
	Cyan    = models.ColCyan
	Yellow  = models.ColYellow
	Red     = models.ColRed
	Magenta = models.ColMagenta
	Blue    = models.ColBlue
	White   = models.ColWhite
)

// ── Cell ──────────────────────────────────────────────────────

type Cell struct {
	Ch    rune
	Color models.ColorName
	Empty bool
}

// ── Chars — configurable characters for each shape part ───────

type Chars struct {
	// Rectangle outline
	RectTop, RectBottom, RectLeft, RectRight rune
	RectTL, RectTR, RectBL, RectBR          rune
	// Rectangle filled
	RectFill rune
	// Circle outline (chosen by tangent angle at each point)
	CircH, CircV, CircD1, CircD2 rune
	// Circle filled
	CircFill rune
	// Line (chosen by slope)
	LineH, LineV, LineD1, LineD2 rune
	// Freehand / pixel
	Free rune
}

// DefaultChars returns the standard box-drawing character set.
func DefaultChars() Chars {
	return Chars{
		RectTop: '─', RectBottom: '─', RectLeft: '│', RectRight: '│',
		RectTL: '┌', RectTR: '┐', RectBL: '└', RectBR: '┘',
		RectFill: '█',
		CircH: '─', CircV: '│', CircD1: '╱', CircD2: '╲',
		CircFill: '█',
		LineH: '─', LineV: '│', LineD1: '╱', LineD2: '╲',
		Free: '█',
	}
}

// ── Context — the main drawing surface ────────────────────────
//
// Modelled after the HTML5 Canvas 2D API.
// Set StrokeColor / FillColor before calling draw methods.

type Context struct {
	// Width and Height of the canvas in characters
	Width, Height int

	// StrokeColor is used by StrokeRect, StrokeCircle, DrawLine, SetPixel
	StrokeColor models.ColorName

	// FillColor is used by FillRect, FillCircle
	FillColor models.ColorName

	// Chars controls which characters are used for each shape element.
	// Change any field to customise the look.
	Chars Chars

	// internal pixel buffer
	cells [][]Cell
}

// New creates a new canvas Context of the given width and height.
// Equivalent to creating a <canvas> element in JS.
func New(width, height int) *Context {
	ctx := &Context{
		Width:       width,
		Height:      height,
		StrokeColor: Green,
		FillColor:   Green,
		Chars:       DefaultChars(),
	}
	ctx.allocate()
	return ctx
}

func (ctx *Context) allocate() {
	ctx.cells = make([][]Cell, ctx.Height)
	for i := range ctx.cells {
		ctx.cells[i] = make([]Cell, ctx.Width)
		for j := range ctx.cells[i] {
			ctx.cells[i][j] = Cell{Ch: ' ', Empty: true}
		}
	}
}

// ── Pixel access ──────────────────────────────────────────────

// set writes a character and color to a cell (bounds-checked).
func (ctx *Context) set(x, y int, ch rune, color models.ColorName) {
	if x >= 0 && x < ctx.Width && y >= 0 && y < ctx.Height {
		ctx.cells[y][x] = Cell{Ch: ch, Color: color}
	}
}

// Get returns the cell at (x, y). Returns an empty cell if out of bounds.
func (ctx *Context) Get(x, y int) Cell {
	if x >= 0 && x < ctx.Width && y >= 0 && y < ctx.Height {
		return ctx.cells[y][x]
	}
	return Cell{Empty: true}
}

// SetPixel draws a single pixel at (x, y) using StrokeColor and Free char.
// Equivalent to JS: ctx.fillRect(x, y, 1, 1)
func (ctx *Context) SetPixel(x, y int) {
	ctx.set(x, y, ctx.Chars.Free, ctx.StrokeColor)
}

// ── Rect ──────────────────────────────────────────────────────

// StrokeRect draws a rectangle outline at (x, y) with given width and height.
// Uses box-drawing characters: ┌─┐ │ └┘
// Equivalent to JS: ctx.strokeRect(x, y, width, height)
func (ctx *Context) StrokeRect(x, y, width, height int) {
	x2, y2 := x+width-1, y+height-1
	ch := ctx.Chars
	c  := ctx.StrokeColor

	// top and bottom edges
	for i := x + 1; i < x2; i++ {
		ctx.set(i, y,  ch.RectTop, c)
		ctx.set(i, y2, ch.RectBottom, c)
	}
	// left and right edges
	for j := y + 1; j < y2; j++ {
		ctx.set(x,  j, ch.RectLeft, c)
		ctx.set(x2, j, ch.RectRight, c)
	}
	// corners
	ctx.set(x,  y,  ch.RectTL, c)
	ctx.set(x2, y,  ch.RectTR, c)
	ctx.set(x,  y2, ch.RectBL, c)
	ctx.set(x2, y2, ch.RectBR, c)
}

// FillRect draws a solid filled rectangle at (x, y) with given width and height.
// Uses the RectFill character (default: █).
// Equivalent to JS: ctx.fillRect(x, y, width, height)
func (ctx *Context) FillRect(x, y, width, height int) {
	for j := y; j < y+height; j++ {
		for i := x; i < x+width; i++ {
			ctx.set(i, j, ctx.Chars.RectFill, ctx.FillColor)
		}
	}
}

// ClearRect erases a rectangular region (fills with empty cells).
// Equivalent to JS: ctx.clearRect(x, y, width, height)
func (ctx *Context) ClearRect(x, y, width, height int) {
	for j := y; j < y+height; j++ {
		for i := x; i < x+width; i++ {
			if i >= 0 && i < ctx.Width && j >= 0 && j < ctx.Height {
				ctx.cells[j][i] = Cell{Ch: ' ', Empty: true}
			}
		}
	}
}

// ── Circle ────────────────────────────────────────────────────

// StrokeCircle draws a circle outline centered at (cx, cy) with given radius.
// Uses box-drawing characters chosen by tangent angle: ─ │ ╱ ╲
// Equivalent to JS: ctx.arc(cx, cy, r, 0, 2*Math.PI); ctx.stroke()
func (ctx *Context) StrokeCircle(cx, cy, radius int) {
	if radius <= 0 {
		ctx.SetPixel(cx, cy)
		return
	}
	c  := ctx.StrokeColor
	ch := ctx.Chars

	plot := func(px, py, ox, oy int) {
		// compute tangent angle to pick the right char
		tx  := float64(-oy)
		ty  := float64(ox) * 0.5 // 0.5 corrects terminal aspect ratio
		ang := math.Atan2(ty, tx) * 180 / math.Pi
		if ang < 0 { ang += 180 }
		var r rune
		switch {
		case ang < 22.5 || ang >= 157.5: r = ch.CircH
		case ang < 67.5:                 r = ch.CircD1
		case ang < 112.5:                r = ch.CircV
		default:                         r = ch.CircD2
		}
		ctx.set(px, py, r, c)
	}

	// Bresenham midpoint circle — 8-fold symmetry
	// x is doubled to correct for terminal char aspect ratio
	x, y := 0, radius
	d := 1 - radius
	for x <= y {
		plot(cx+x*2, cy-y,  x, -y); plot(cx-x*2, cy-y, -x, -y)
		plot(cx+x*2, cy+y,  x,  y); plot(cx-x*2, cy+y, -x,  y)
		plot(cx+y*2, cy-x,  y, -x); plot(cx-y*2, cy-x, -y, -x)
		plot(cx+y*2, cy+x,  y,  x); plot(cx-y*2, cy+x, -y,  x)
		if d < 0 { d += 2*x + 3 } else { d += 2*(x-y) + 5; y-- }
		x++
	}
}

// FillCircle draws a filled circle centered at (cx, cy) with given radius.
// Uses the CircFill character (default: █).
// Equivalent to JS: ctx.arc(cx, cy, r, 0, 2*Math.PI); ctx.fill()
func (ctx *Context) FillCircle(cx, cy, radius int) {
	if radius <= 0 {
		ctx.set(cx, cy, ctx.Chars.CircFill, ctx.FillColor)
		return
	}
	rf := float64(radius)
	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius*2; x <= cx+radius*2; x++ {
			dx := float64(x-cx) * 0.5
			dy := float64(y - cy)
			if math.Sqrt(dx*dx+dy*dy) <= rf+0.3 {
				ctx.set(x, y, ctx.Chars.CircFill, ctx.FillColor)
			}
		}
	}
}

// ── Line ──────────────────────────────────────────────────────

// DrawLine draws a straight line from (x1, y1) to (x2, y2).
// Automatically picks the best character based on slope:
//   - horizontal  → ─
//   - vertical    → │
//   - diagonal \  → ╲
//   - diagonal /  → ╱
//
// Equivalent to JS: ctx.moveTo(x1,y1); ctx.lineTo(x2,y2); ctx.stroke()
func (ctx *Context) DrawLine(x1, y1, x2, y2 int) {
	dx, dy := x2-x1, y2-y1
	adx, ady := iabs(dx), iabs(dy)
	ch := ctx.Chars
	var r rune
	switch {
	case ady == 0:             r = ch.LineH
	case adx == 0:             r = ch.LineV
	case (dx > 0) == (dy > 0): r = ch.LineD2
	default:                   r = ch.LineD1
	}
	// Bresenham walk
	sx, sy := isign(dx), isign(dy)
	err := adx - ady
	for x, y := x1, y1; ; {
		ctx.set(x, y, r, ctx.StrokeColor)
		if x == x2 && y == y2 { break }
		e2 := 2 * err
		if e2 > -ady { err -= ady; x += sx }
		if e2 < adx  { err += adx; y += sy }
	}
}

// ── Canvas-level ops ──────────────────────────────────────────

// Clear erases the entire canvas.
// Equivalent to JS: ctx.clearRect(0, 0, canvas.width, canvas.height)
func (ctx *Context) Clear() {
	ctx.allocate()
}

// ── Snapshot / restore ────────────────────────────────────────

// Snapshot returns a deep copy of the current canvas state.
// Useful for saving state before a preview draw.
func (ctx *Context) Snapshot() [][]Cell {
	snap := make([][]Cell, ctx.Height)
	for i := range snap {
		snap[i] = make([]Cell, ctx.Width)
		copy(snap[i], ctx.cells[i])
	}
	return snap
}

// Restore replaces the canvas cells with a previously taken snapshot.
func (ctx *Context) Restore(snap [][]Cell) {
	for i := range snap {
		if i < ctx.Height {
			copy(ctx.cells[i], snap[i])
		}
	}
}

// ── Merge / composite ─────────────────────────────────────────

// DrawContext draws another Context onto this one at offset (ox, oy).
// Non-empty cells from src overwrite cells in dst.
// Equivalent to JS: ctx.drawImage(src, ox, oy)
func (ctx *Context) DrawContext(src *Context, ox, oy int) {
	for y := 0; y < src.Height; y++ {
		for x := 0; x < src.Width; x++ {
			c := src.cells[y][x]
			if !c.Empty {
				ctx.set(x+ox, y+oy, c.Ch, c.Color)
			}
		}
	}
}

// ── Render from saved shapes (used by the app) ────────────────

// RenderAll creates a fresh Context and draws all saved shapes onto it.
func RenderAll(w, h int, shapes []models.Shape, ch Chars) *Context {
	ctx := New(w, h)
	ctx.Chars = ch
	for _, s := range shapes {
		drawShape(ctx, s)
	}
	return ctx
}

// drawShape draws one saved shape onto a context.
func drawShape(ctx *Context, s models.Shape) {
	if s.Color != "" {
		ctx.StrokeColor = s.Color
		ctx.FillColor   = s.Color
	}
	switch s.Type {
	case models.ShapeRect:
		// stored as x1,y1 → x2,y2 (two corners)
		x, y := min2(s.X1, s.X2), min2(s.Y1, s.Y2)
		w, h := iabs(s.X2-s.X1)+1, iabs(s.Y2-s.Y1)+1
		if s.Filled { ctx.FillRect(x, y, w, h) } else { ctx.StrokeRect(x, y, w, h) }

	case models.ShapeCircle:
		if s.Filled { ctx.FillCircle(s.X1, s.Y1, s.Radius) } else { ctx.StrokeCircle(s.X1, s.Y1, s.Radius) }

	case models.ShapeLine:
		ctx.DrawLine(s.X1, s.Y1, s.X2, s.Y2)

	case models.ShapeFree:
		ctx.set(s.X1, s.Y1, ctx.Chars.Free, s.Color)
	}
}

// ── Helpers ───────────────────────────────────────────────────

func iabs(x int) int   { if x < 0 { return -x }; return x }
func isign(x int) int  { if x < 0 { return -1 } else if x > 0 { return 1 }; return 0 }
func min2(a, b int) int { if a < b { return a }; return b }