package main

import "math"

// Canvas is the pixel buffer. Create with newCanvas, draw via getContext.
type Canvas struct {
	width, height int
	cells         [][]cell
}

func newCanvas(w, h int) *Canvas {
	c := &Canvas{width: w, height: h}
	c.alloc()
	return c
}

func (c *Canvas) alloc() {
	c.cells = make([][]cell, c.height)
	for y := range c.cells {
		c.cells[y] = make([]cell, c.width)
		for x := range c.cells[y] {
			c.cells[y][x] = cell{ch: ' ', empty: true}
		}
	}
}

// Ctx is the drawing context. Get one with getContext(canvas).
// Set strokeStyle / fillStyle before calling draw functions.
type Ctx struct {
	strokeStyle Color // used by stroke*, drawLine, setPixel
	fillStyle   Color // used by fill*, fillText
	canvas      *Canvas
	pathX, pathY []int // accumulated by moveTo/lineTo, consumed by stroke/fill
}

func getContext(c *Canvas) *Ctx {
	return &Ctx{canvas: c, strokeStyle: ColGreen, fillStyle: ColGreen}
}

type cell struct {
	ch    rune
	color Color
	empty bool
}

func (ctx *Ctx) set(x, y int, ch rune, col Color) {
	c := ctx.canvas
	if x >= 0 && x < c.width && y >= 0 && y < c.height {
		c.cells[y][x] = cell{ch: ch, color: col}
	}
}

func (ctx *Ctx) getCell(x, y int) cell {
	c := ctx.canvas
	if x >= 0 && x < c.width && y >= 0 && y < c.height {
		return c.cells[y][x]
	}
	return cell{empty: true}
}

// ── setPixel ─────────────────────────────────────────────────

func setPixel(ctx *Ctx, x, y int) {
	ctx.set(x, y, freeChar, ctx.strokeStyle)
}

// ── strokeRect / fillRect / clearRect ────────────────────────

func strokeRect(ctx *Ctx, x, y, w, h int) {
	if w <= 0 || h <= 0 { return }
	x2, y2 := x+w-1, y+h-1
	col := ctx.strokeStyle
	if h == 1 {
		for i := x; i <= x2; i++ { ctx.set(i, y, rectTop, col) }
		return
	}
	if w == 1 {
		for j := y; j <= y2; j++ { ctx.set(x, j, rectLeft, col) }
		return
	}
	for i := x + 1; i < x2; i++ {
		ctx.set(i, y,  rectTop,    col)
		ctx.set(i, y2, rectBottom, col)
	}
	for j := y + 1; j < y2; j++ {
		ctx.set(x,  j, rectLeft,  col)
		ctx.set(x2, j, rectRight, col)
	}
	ctx.set(x,  y,  rectTL, col)
	ctx.set(x2, y,  rectTR, col)
	ctx.set(x,  y2, rectBL, col)
	ctx.set(x2, y2, rectBR, col)
}

func fillRect(ctx *Ctx, x, y, w, h int) {
	if w <= 0 || h <= 0 { return }
	col := ctx.fillStyle
	for j := y; j < y+h; j++ {
		for i := x; i < x+w; i++ {
			ctx.set(i, j, rectFill, col)
		}
	}
}

func clearRect(ctx *Ctx, x, y, w, h int) {
	c := ctx.canvas
	for j := y; j < y+h; j++ {
		for i := x; i < x+w; i++ {
			if i >= 0 && i < c.width && j >= 0 && j < c.height {
				c.cells[j][i] = cell{ch: ' ', empty: true}
			}
		}
	}
}

// ── strokeCircle / fillCircle ─────────────────────────────────

func strokeCircle(ctx *Ctx, cx, cy, radius int) {
	if radius <= 0 { setPixel(ctx, cx, cy); return }
	col := ctx.strokeStyle
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
		ctx.set(px, py, ch, col)
	}
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

func fillCircle(ctx *Ctx, cx, cy, radius int) {
	if radius <= 0 { ctx.set(cx, cy, circFill, ctx.fillStyle); return }
	rf, col := float64(radius), ctx.fillStyle
	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius*2; x <= cx+radius*2; x++ {
			dx := float64(x-cx) * 0.5
			dy := float64(y - cy)
			if math.Sqrt(dx*dx+dy*dy) <= rf+0.3 {
				ctx.set(x, y, circFill, col)
			}
		}
	}
}

// ── strokeEllipse / fillEllipse ───────────────────────────────

func strokeEllipse(ctx *Ctx, cx, cy, rx, ry int) {
	if rx <= 0 || ry <= 0 { setPixel(ctx, cx, cy); return }
	col   := ctx.strokeStyle
	steps := (rx + ry) * 8
	prev  := false
	var px0, py0 int
	for i := 0; i <= steps; i++ {
		t  := 2 * math.Pi * float64(i) / float64(steps)
		px := cx + int(math.Round(float64(rx)*math.Cos(t)))
		py := cy + int(math.Round(float64(ry)*math.Sin(t)))
		if prev { bresenham(ctx, px0, py0, px, py, col) }
		px0, py0 = px, py
		prev = true
	}
}

func fillEllipse(ctx *Ctx, cx, cy, rx, ry int) {
	if rx <= 0 || ry <= 0 { ctx.set(cx, cy, rectFill, ctx.fillStyle); return }
	col      := ctx.fillStyle
	rxf, ryf := float64(rx), float64(ry)
	for y := cy - ry; y <= cy+ry; y++ {
		for x := cx - rx*2; x <= cx+rx*2; x++ {
			dx := float64(x-cx) * 0.5
			dy := float64(y - cy)
			if (dx*dx)/(rxf*rxf)+(dy*dy)/(ryf*ryf) <= 1.0 {
				ctx.set(x, y, rectFill, col)
			}
		}
	}
}

// ── drawLine ─────────────────────────────────────────────────

// drawLine draws from (x1,y1) to (x2,y2). Slope picks char: ─ │ ╱ ╲
func drawLine(ctx *Ctx, x1, y1, x2, y2 int) {
	bresenham(ctx, x1, y1, x2, y2, ctx.strokeStyle)
}

func bresenham(ctx *Ctx, x1, y1, x2, y2 int, col Color) {
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
		ctx.set(x, y, ch, col)
		if x == x2 && y == y2 { break }
		e2 := 2 * err
		if e2 > -ady { err -= ady; x += sx }
		if e2 < adx  { err += adx; y += sy }
	}
}

// ── Path API: beginPath / moveTo / lineTo / closePath / stroke / fill ─

func beginPath(ctx *Ctx) {
	ctx.pathX = ctx.pathX[:0]
	ctx.pathY = ctx.pathY[:0]
}

func moveTo(ctx *Ctx, x, y int) {
	ctx.pathX = append(ctx.pathX, x)
	ctx.pathY = append(ctx.pathY, y)
}

func lineTo(ctx *Ctx, x, y int) {
	ctx.pathX = append(ctx.pathX, x)
	ctx.pathY = append(ctx.pathY, y)
}

func closePath(ctx *Ctx) {
	if len(ctx.pathX) < 2 { return }
	ctx.pathX = append(ctx.pathX, ctx.pathX[0])
	ctx.pathY = append(ctx.pathY, ctx.pathY[0])
}

func stroke(ctx *Ctx) {
	for i := 1; i < len(ctx.pathX); i++ {
		bresenham(ctx, ctx.pathX[i-1], ctx.pathY[i-1], ctx.pathX[i], ctx.pathY[i], ctx.strokeStyle)
	}
}

// fill fills the closed polygon defined by the current path.
func fill(ctx *Ctx) {
	n := len(ctx.pathX)
	if n < 3 { return }
	col := ctx.fillStyle
	minY, maxY := ctx.pathY[0], ctx.pathY[0]
	for _, y := range ctx.pathY {
		if y < minY { minY = y }
		if y > maxY { maxY = y }
	}
	for y := minY; y <= maxY; y++ {
		var xs []int
		for i := 0; i < n; i++ {
			x1, y1 := ctx.pathX[i], ctx.pathY[i]
			x2, y2 := ctx.pathX[(i+1)%n], ctx.pathY[(i+1)%n]
			if (y1 <= y && y < y2) || (y2 <= y && y < y1) {
				xs = append(xs, x1+(x2-x1)*(y-y1)/(y2-y1))
			}
		}
		for i := 0; i < len(xs)-1; i++ { // sort
			for j := i + 1; j < len(xs); j++ {
				if xs[i] > xs[j] { xs[i], xs[j] = xs[j], xs[i] }
			}
		}
		for i := 0; i+1 < len(xs); i += 2 {
			for x := xs[i]; x <= xs[i+1]; x++ { ctx.set(x, y, rectFill, col) }
		}
	}
}

// ── strokeTriangle / fillTriangle ────────────────────────────

func strokeTriangle(ctx *Ctx, x1, y1, x2, y2, x3, y3 int) {
	col := ctx.strokeStyle
	bresenham(ctx, x1, y1, x2, y2, col)
	bresenham(ctx, x2, y2, x3, y3, col)
	bresenham(ctx, x3, y3, x1, y1, col)
}

func fillTriangle(ctx *Ctx, x1, y1, x2, y2, x3, y3 int) {
	col := ctx.fillStyle
	if y1 > y2 { x1, y1, x2, y2 = x2, y2, x1, y1 }
	if y1 > y3 { x1, y1, x3, y3 = x3, y3, x1, y1 }
	if y2 > y3 { x2, y2, x3, y3 = x3, y3, x2, y2 }
	lerp := func(ya, yb, xa, xb, y int) int {
		if ya == yb { return xa }
		return xa + (xb-xa)*(y-ya)/(yb-ya)
	}
	for y := y1; y <= y3; y++ {
		lx := lerp(y1, y3, x1, x3, y)
		var rx int
		if y < y2 { rx = lerp(y1, y2, x1, x2, y) } else { rx = lerp(y2, y3, x2, x3, y) }
		if lx > rx { lx, rx = rx, lx }
		for x := lx; x <= rx; x++ { ctx.set(x, y, rectFill, col) }
	}
}

// ── fillText / strokeText ─────────────────────────────────────

func fillText(ctx *Ctx, s string, x, y int) {
	col := ctx.fillStyle
	for i, ch := range s {
		ctx.set(x+i, y, ch, col)
	}
}

// strokeText draws text inside a box border:  ┌───────┐ │ text │ └───────┘
func strokeText(ctx *Ctx, s string, x, y int) {
	col   := ctx.strokeStyle
	runes := []rune(s)
	w     := len(runes) + 2
	ctx.set(x, y, rectTL, col)
	for i := 0; i < len(runes); i++ { ctx.set(x+1+i, y, rectTop, col) }
	ctx.set(x+w-1, y, rectTR, col)
	ctx.set(x, y+1, rectLeft, col)
	for i, ch := range runes { ctx.set(x+1+i, y+1, ch, col) }
	ctx.set(x+w-1, y+1, rectRight, col)
	ctx.set(x, y+2, rectBL, col)
	for i := 0; i < len(runes); i++ { ctx.set(x+1+i, y+2, rectBottom, col) }
	ctx.set(x+w-1, y+2, rectBR, col)
}

// ── Canvas-level ops ─────────────────────────────────────────

func clearCanvas(ctx *Ctx)              { ctx.canvas.alloc() }
func saveCanvas(ctx *Ctx) [][]cell      {
	c := ctx.canvas
	snap := make([][]cell, c.height)
	for i := range snap {
		snap[i] = make([]cell, c.width)
		copy(snap[i], c.cells[i])
	}
	return snap
}
func restoreCanvas(ctx *Ctx, snap [][]cell) {
	c := ctx.canvas
	for i := range snap {
		if i < c.height { copy(c.cells[i], snap[i]) }
	}
}

// drawCanvas composites src onto dst at offset (ox, oy). Non-empty cells overwrite.
func drawCanvas(dst *Ctx, src *Ctx, ox, oy int) {
	sc := src.canvas
	for y := 0; y < sc.height; y++ {
		for x := 0; x < sc.width; x++ {
			cl := sc.cells[y][x]
			if !cl.empty { dst.set(x+ox, y+oy, cl.ch, cl.color) }
		}
	}
}

// ── Helpers ───────────────────────────────────────────────────

func iabs(x int) int    { if x < 0 { return -x }; return x }
func isign(x int) int   { if x < 0 { return -1 } else if x > 0 { return 1 }; return 0 }
func imin(a, b int) int { if a < b { return a }; return b }

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