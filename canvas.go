package main


import "math"

// ─────────────────────────────────────────────────────────────
// Cell — one character position
// ─────────────────────────────────────────────────────────────

// Cell holds one rendered character on the canvas.
type Cell struct {
	Ch    rune
	Color Color
	Empty bool
}

// ─────────────────────────────────────────────────────────────
// Canvas
// ─────────────────────────────────────────────────────────────

// Canvas is a 2D grid of Cells.
// Set StrokeColor / FillColor before calling draw methods,
// exactly like you set ctx.strokeStyle / ctx.fillStyle in JS.
type Canvas struct {
	// Width and Height of the canvas in character cells.
	W, H int

	// StrokeColor is used by all Stroke* methods and DrawLine / SetPixel.
	StrokeColor Color

	// FillColor is used by all Fill* methods and FillText.
	FillColor Color

	// internal cell buffer — rows first: cells[y][x]
	cells [][]Cell

	// path accumulator for BeginPath / MoveTo / LineTo / Stroke
	pathX []int
	pathY []int
}

// NewCanvas creates a blank canvas of size w×h.
// Equivalent to document.createElement("canvas") + getContext("2d") in JS.
func NewCanvas(w, h int) *Canvas {
	c := &Canvas{
		W:           w,
		H:           h,
		StrokeColor: ColGreen,
		FillColor:   ColGreen,
	}
	c.alloc()
	return c
}

func (c *Canvas) alloc() {
	c.cells = make([][]Cell, c.H)
	for y := range c.cells {
		c.cells[y] = make([]Cell, c.W)
		for x := range c.cells[y] {
			c.cells[y][x] = Cell{Ch: ' ', Empty: true}
		}
	}
}

// ─────────────────────────────────────────────────────────────
// Low-level pixel access
// ─────────────────────────────────────────────────────────────

// set writes one cell — bounds-checked, safe to call with out-of-range coords.
func (c *Canvas) set(x, y int, ch rune, col Color) {
	if x >= 0 && x < c.W && y >= 0 && y < c.H {
		c.cells[y][x] = Cell{Ch: ch, Color: col}
	}
}

// Get returns the cell at (x, y). Returns an empty cell if out of bounds.
func (c *Canvas) Get(x, y int) Cell {
	if x >= 0 && x < c.W && y >= 0 && y < c.H {
		return c.cells[y][x]
	}
	return Cell{Empty: true}
}

// SetPixel draws a single '█' at (x, y) using StrokeColor.
// Equivalent to ctx.fillRect(x, y, 1, 1) in JS.
func (c *Canvas) SetPixel(x, y int) {
	c.set(x, y, freeChar, c.StrokeColor)
}

// ─────────────────────────────────────────────────────────────
// Rectangle  (JS: strokeRect / fillRect / clearRect)
// ─────────────────────────────────────────────────────────────

// StrokeRect draws a rectangle outline using box-drawing characters.
//
//	┌──────────┐
//	│          │
//	└──────────┘
//
// Equivalent to ctx.strokeRect(x, y, width, height) in JS.
func (c *Canvas) StrokeRect(x, y, w, h int) {
	if w <= 0 || h <= 0 { return }
	x2, y2 := x+w-1, y+h-1
	col := c.StrokeColor

	if h == 1 {
		// degenerate: single row
		for i := x; i <= x2; i++ { c.set(i, y, rectTop, col) }
		return
	}
	if w == 1 {
		// degenerate: single column
		for j := y; j <= y2; j++ { c.set(x, j, rectLeft, col) }
		return
	}

	// top & bottom edges
	for i := x + 1; i < x2; i++ {
		c.set(i, y,  rectTop, col)
		c.set(i, y2, rectBottom, col)
	}
	// left & right edges
	for j := y + 1; j < y2; j++ {
		c.set(x,  j, rectLeft, col)
		c.set(x2, j, rectRight, col)
	}
	// corners
	c.set(x,  y,  rectTL, col)
	c.set(x2, y,  rectTR, col)
	c.set(x,  y2, rectBL, col)
	c.set(x2, y2, rectBR, col)
}

// FillRect fills a rectangle with solid '█' blocks using FillColor.
// Equivalent to ctx.fillRect(x, y, width, height) in JS.
func (c *Canvas) FillRect(x, y, w, h int) {
	if w <= 0 || h <= 0 { return }
	col := c.FillColor
	for j := y; j < y+h; j++ {
		for i := x; i < x+w; i++ {
			c.set(i, j, rectFill, col)
		}
	}
}

// ClearRect erases a rectangular region, leaving it blank.
// Equivalent to ctx.clearRect(x, y, width, height) in JS.
func (c *Canvas) ClearRect(x, y, w, h int) {
	for j := y; j < y+h; j++ {
		for i := x; i < x+w; i++ {
			if i >= 0 && i < c.W && j >= 0 && j < c.H {
				c.cells[j][i] = Cell{Ch: ' ', Empty: true}
			}
		}
	}
}

// ─────────────────────────────────────────────────────────────
// Circle  (terminal extension — no direct JS equivalent)
// ─────────────────────────────────────────────────────────────

// StrokeCircle draws a circle outline centered at (cx, cy) with the given radius.
// Characters are chosen by tangent angle: ─ │ ╱ ╲
// The x-axis is doubled internally to compensate for terminal character aspect ratio.
func (c *Canvas) StrokeCircle(cx, cy, radius int) {
	if radius <= 0 { c.SetPixel(cx, cy); return }
	col := c.StrokeColor

	plot := func(px, py, ox, oy int) {
		tx  := float64(-oy)
		ty  := float64(ox) * 0.5
		ang := math.Atan2(ty, tx) * 180 / math.Pi
		if ang < 0 { ang += 180 }
		var ch rune
		switch {
		case ang < 22.5 || ang >= 157.5: ch = circH
		case ang < 67.5:                 ch = circD1
		case ang < 112.5:                ch = circV
		default:                         ch = circD2
		}
		c.set(px, py, ch, col)
	}

	// Bresenham midpoint — x*2 corrects for terminal aspect ratio
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

// FillCircle draws a filled solid circle centered at (cx, cy).
func (c *Canvas) FillCircle(cx, cy, radius int) {
	if radius <= 0 { c.set(cx, cy, circFill, c.FillColor); return }
	rf  := float64(radius)
	col := c.FillColor
	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius*2; x <= cx+radius*2; x++ {
			dx := float64(x-cx) * 0.5
			dy := float64(y - cy)
			if math.Sqrt(dx*dx+dy*dy) <= rf+0.3 {
				c.set(x, y, circFill, col)
			}
		}
	}
}

// ─────────────────────────────────────────────────────────────
// Ellipse  (JS: ellipse — but terminal version takes rx, ry)
// ─────────────────────────────────────────────────────────────

// StrokeEllipse draws an ellipse outline centered at (cx, cy)
// with horizontal radius rx and vertical radius ry.
// Equivalent to ctx.ellipse(cx, cy, rx, ry, 0, 0, 2*Math.PI); ctx.stroke()
func (c *Canvas) StrokeEllipse(cx, cy, rx, ry int) {
	if rx <= 0 || ry <= 0 { c.SetPixel(cx, cy); return }
	col := c.StrokeColor
	// parametric walk around the ellipse at fine angular steps
	steps := (rx + ry) * 8
	prev := false
	var px0, py0 int
	for i := 0; i <= steps; i++ {
		t   := 2 * math.Pi * float64(i) / float64(steps)
		px  := cx + int(math.Round(float64(rx)*math.Cos(t)))
		py  := cy + int(math.Round(float64(ry)*math.Sin(t)))
		if prev { c.lineRaw(px0, py0, px, py, col) }
		px0, py0 = px, py
		prev = true
	}
}

// FillEllipse draws a filled solid ellipse centered at (cx, cy).
func (c *Canvas) FillEllipse(cx, cy, rx, ry int) {
	if rx <= 0 || ry <= 0 { c.set(cx, cy, '█', c.FillColor); return }
	col := c.FillColor
	rxf, ryf := float64(rx), float64(ry)
	for y := cy - ry; y <= cy+ry; y++ {
		for x := cx - rx*2; x <= cx+rx*2; x++ {
			dx := float64(x-cx) * 0.5
			dy := float64(y - cy)
			if (dx*dx)/(rxf*rxf)+(dy*dy)/(ryf*ryf) <= 1.0 {
				c.set(x, y, '█', col)
			}
		}
	}
}

// ─────────────────────────────────────────────────────────────
// Line  (JS: moveTo / lineTo / stroke  +  shorthand DrawLine)
// ─────────────────────────────────────────────────────────────

// DrawLine draws a straight line from (x1, y1) to (x2, y2) using StrokeColor.
// Character is chosen by the line's slope: ─ │ ╱ ╲
// Shorthand for BeginPath / MoveTo / LineTo / Stroke in one call.
func (c *Canvas) DrawLine(x1, y1, x2, y2 int) {
	c.bresenham(x1, y1, x2, y2, c.StrokeColor)
}

// bresenham walks a line and stamps the slope-appropriate character.
func (c *Canvas) bresenham(x1, y1, x2, y2 int, col Color) {
	dx, dy   := x2-x1, y2-y1
	adx, ady := iabs(dx), iabs(dy)
	var ch rune
	switch {
	case ady == 0:              ch = lineH
	case adx == 0:              ch = lineV
	case (dx > 0) == (dy > 0): ch = lineD2
	default:                    ch = lineD1
	}
	sx, sy := isign(dx), isign(dy)
	err := adx - ady
	x, y := x1, y1
	for {
		c.set(x, y, ch, col)
		if x == x2 && y == y2 { break }
		e2 := 2 * err
		if e2 > -ady { err -= ady; x += sx }
		if e2 < adx  { err += adx; y += sy }
	}
}

// lineRaw is like bresenham but picks '·' for diagonals (used inside ellipse).
func (c *Canvas) lineRaw(x1, y1, x2, y2 int, col Color) {
	c.bresenham(x1, y1, x2, y2, col)
}

// ── Path API (JS-style) ───────────────────────────────────────

// BeginPath resets the current path, discarding any previous MoveTo / LineTo calls.
// Equivalent to ctx.beginPath() in JS.
func (c *Canvas) BeginPath() {
	c.pathX = c.pathX[:0]
	c.pathY = c.pathY[:0]
}

// MoveTo moves the "pen" to (x, y) without drawing anything.
// Equivalent to ctx.moveTo(x, y) in JS.
func (c *Canvas) MoveTo(x, y int) {
	c.pathX = append(c.pathX, x)
	c.pathY = append(c.pathY, y)
}

// LineTo adds a straight line from the last point to (x, y).
// Equivalent to ctx.lineTo(x, y) in JS.
func (c *Canvas) LineTo(x, y int) {
	c.pathX = append(c.pathX, x)
	c.pathY = append(c.pathY, y)
}

// Stroke draws all lines in the current path using StrokeColor.
// Equivalent to ctx.stroke() in JS.
func (c *Canvas) Stroke() {
	if len(c.pathX) < 2 { return }
	for i := 1; i < len(c.pathX); i++ {
		c.bresenham(c.pathX[i-1], c.pathY[i-1], c.pathX[i], c.pathY[i], c.StrokeColor)
	}
}

// ClosePath adds a line from the last point back to the first point.
// Equivalent to ctx.closePath() in JS.
func (c *Canvas) ClosePath() {
	if len(c.pathX) < 2 { return }
	c.pathX = append(c.pathX, c.pathX[0])
	c.pathY = append(c.pathY, c.pathY[0])
}

// ─────────────────────────────────────────────────────────────
// Triangle  (terminal extension)
// ─────────────────────────────────────────────────────────────

// StrokeTriangle draws a triangle outline through three points.
func (c *Canvas) StrokeTriangle(x1, y1, x2, y2, x3, y3 int) {
	col := c.StrokeColor
	c.bresenham(x1, y1, x2, y2, col)
	c.bresenham(x2, y2, x3, y3, col)
	c.bresenham(x3, y3, x1, y1, col)
}

// FillTriangle draws a solid filled triangle through three points.
// Uses scanline rasterisation.
func (c *Canvas) FillTriangle(x1, y1, x2, y2, x3, y3 int) {
	col := c.FillColor
	// sort vertices by Y
	if y1 > y2 { x1, y1, x2, y2 = x2, y2, x1, y1 }
	if y1 > y3 { x1, y1, x3, y3 = x3, y3, x1, y1 }
	if y2 > y3 { x2, y2, x3, y3 = x3, y3, x2, y2 }

	lerp := func(ya, yb, xa, xb, y int) int {
		if ya == yb { return xa }
		return xa + (xb-xa)*(y-ya)/(yb-ya)
	}
	for y := y1; y <= y3; y++ {
		var lx, rx int
		lx = lerp(y1, y3, x1, x3, y)
		if y < y2 {
			rx = lerp(y1, y2, x1, x2, y)
		} else {
			rx = lerp(y2, y3, x2, x3, y)
		}
		if lx > rx { lx, rx = rx, lx }
		for x := lx; x <= rx; x++ { c.set(x, y, '█', col) }
	}
}

// ─────────────────────────────────────────────────────────────
// Text  (JS: fillText / strokeText)
// ─────────────────────────────────────────────────────────────

// FillText draws the string s starting at (x, y) using FillColor.
// Each character in s occupies exactly one cell.
// Equivalent to ctx.fillText(text, x, y) in JS.
func (c *Canvas) FillText(s string, x, y int) {
	col := c.FillColor
	for i, ch := range s {
		c.set(x+i, y, ch, col)
	}
}

// StrokeText draws the string s inside a thin box border starting at (x, y).
// The text is drawn with StrokeColor; the border uses the same color.
// Roughly equivalent to ctx.strokeText(text, x, y) in JS
// (terminal adaptation — stroked text isn't a native concept in terminals).
func (c *Canvas) StrokeText(s string, x, y int) {
	col  := c.StrokeColor
	runes := []rune(s)
	w    := len(runes) + 2  // +2 for left/right border
	// top border
	c.set(x, y, '┌', col)
	for i := 0; i < len(runes); i++ { c.set(x+1+i, y, '─', col) }
	c.set(x+w-1, y, '┐', col)
	// text row
	c.set(x, y+1, '│', col)
	for i, ch := range runes { c.set(x+1+i, y+1, ch, col) }
	c.set(x+w-1, y+1, '│', col)
	// bottom border
	c.set(x, y+2, '└', col)
	for i := 0; i < len(runes); i++ { c.set(x+1+i, y+2, '─', col) }
	c.set(x+w-1, y+2, '┘', col)
}

// ─────────────────────────────────────────────────────────────
// Canvas-level operations  (JS: save / restore / clearRect on whole canvas)
// ─────────────────────────────────────────────────────────────

// Clear wipes the entire canvas to blank cells.
// Equivalent to ctx.clearRect(0, 0, canvas.width, canvas.height) in JS.
func (c *Canvas) Clear() { c.alloc() }

// Save returns a deep copy of the current cell buffer.
// Equivalent to ctx.save() in JS (but we store pixels, not state stack).
func (c *Canvas) Save() [][]Cell {
	snap := make([][]Cell, c.H)
	for i := range snap {
		snap[i] = make([]Cell, c.W)
		copy(snap[i], c.cells[i])
	}
	return snap
}

// Restore replaces the cell buffer with a previously saved snapshot.
// Equivalent to ctx.restore() in JS.
func (c *Canvas) Restore(snap [][]Cell) {
	for i := range snap {
		if i < c.H { copy(c.cells[i], snap[i]) }
	}
}

// DrawCanvas composites another Canvas on top of this one at offset (ox, oy).
// Only non-empty cells from src overwrite cells in dst.
// Equivalent to ctx.drawImage(canvas, ox, oy) in JS.
func (c *Canvas) DrawCanvas(src *Canvas, ox, oy int) {
	for y := 0; y < src.H; y++ {
		for x := 0; x < src.W; x++ {
			cell := src.cells[y][x]
			if !cell.Empty { c.set(x+ox, y+oy, cell.Ch, cell.Color) }
		}
	}
}

// ─────────────────────────────────────────────────────────────
// Math helpers
// ─────────────────────────────────────────────────────────────

func iabs(x int) int    { if x < 0 { return -x }; return x }
func isign(x int) int   { if x < 0 { return -1 } else if x > 0 { return 1 }; return 0 }
func imin(a, b int) int { if a < b { return a }; return b }
func imax(a, b int) int { if a > b { return a }; return b }

func isqrt(x float64) float64 {
	if x <= 0 { return 0 }
	z := x
	for i := 0; i < 60; i++ { z = (z + x/z) / 2 }
	return z
}

func idist(x1, y1, x2, y2 int) int {
	dx := float64(x2-x1) * 0.5
	dy := float64(y2 - y1)
	r  := int(isqrt(dx*dx + dy*dy))
	if r < 1 { r = 1 }
	return r
}