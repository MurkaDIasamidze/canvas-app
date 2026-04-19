package canvas

import (
	"canvas-tui/models"
	"math"
)

// ── Cell ──────────────────────────────────────────────────────

type Cell struct {
	Ch    rune
	Color models.ColorName
	Empty bool // true = no shape was drawn here
}

// ── Grid ──────────────────────────────────────────────────────

type Grid struct {
	W, H  int
	Cells [][]Cell
}

func New(w, h int, empty rune) *Grid {
	cells := make([][]Cell, h)
	for i := range cells {
		cells[i] = make([]Cell, w)
		for j := range cells[i] {
			cells[i][j] = Cell{Ch: empty, Empty: true}
		}
	}
	return &Grid{W: w, H: h, Cells: cells}
}

func (g *Grid) Set(x, y int, ch rune, c models.ColorName) {
	if x >= 0 && x < g.W && y >= 0 && y < g.H {
		g.Cells[y][x] = Cell{Ch: ch, Color: c}
	}
}

func (g *Grid) Get(x, y int) Cell {
	if x >= 0 && x < g.W && y >= 0 && y < g.H {
		return g.Cells[y][x]
	}
	return Cell{Empty: true}
}

// ── Config-driven chars ────────────────────────────────────────

type Chars struct {
	// Rect outline
	RectTop, RectBottom, RectLeft, RectRight rune
	RectTL, RectTR, RectBL, RectBR          rune
	RectFill                                 rune
	// Circle
	CircH, CircV, CircD1, CircD2            rune
	CircFill                                 rune
	// Line
	LineH, LineV, LineD1, LineD2            rune
	// Free
	Free                                    rune
}

// ── Drawing functions ──────────────────────────────────────────

// Rect draws a rectangle using the provided char config.
func Rect(g *Grid, x1, y1, x2, y2 int, c models.ColorName, filled bool, ch Chars) {
	if x1 > x2 { x1, x2 = x2, x1 }
	if y1 > y2 { y1, y2 = y2, y1 }

	if filled {
		for y := y1; y <= y2; y++ {
			for x := x1; x <= x2; x++ {
				g.Set(x, y, ch.RectFill, c)
			}
		}
		return
	}
	// top & bottom edges
	for x := x1 + 1; x < x2; x++ {
		g.Set(x, y1, ch.RectTop, c)
		g.Set(x, y2, ch.RectBottom, c)
	}
	// left & right edges
	for y := y1 + 1; y < y2; y++ {
		g.Set(x1, y, ch.RectLeft, c)
		g.Set(x2, y, ch.RectRight, c)
	}
	// corners
	g.Set(x1, y1, ch.RectTL, c)
	g.Set(x2, y1, ch.RectTR, c)
	g.Set(x1, y2, ch.RectBL, c)
	g.Set(x2, y2, ch.RectBR, c)
}

// Circle draws a circle using midpoint algorithm + tangent char selection.
func Circle(g *Grid, cx, cy, r int, c models.ColorName, filled bool, ch Chars) {
	if r <= 0 { g.Set(cx, cy, ch.CircFill, c); return }

	if filled {
		rf := float64(r)
		for y := cy - r; y <= cy+r; y++ {
			for x := cx - r*2; x <= cx+r*2; x++ {
				dx := float64(x-cx) * 0.5
				dy := float64(y - cy)
				if math.Sqrt(dx*dx+dy*dy) <= rf+0.3 {
					g.Set(x, y, ch.CircFill, c)
				}
			}
		}
		return
	}

	plot := func(px, py, ox, oy int) {
		tx := float64(-oy)
		ty := float64(ox) * 0.5
		ang := math.Atan2(ty, tx) * 180 / math.Pi
		if ang < 0 { ang += 180 }
		var r rune
		switch {
		case ang < 22.5 || ang >= 157.5: r = ch.CircH
		case ang < 67.5:                 r = ch.CircD1
		case ang < 112.5:                r = ch.CircV
		default:                         r = ch.CircD2
		}
		g.Set(px, py, r, c)
	}

	x, y := 0, r
	d := 1 - r
	for x <= y {
		plot(cx+x*2, cy-y,  x, -y); plot(cx-x*2, cy-y, -x, -y)
		plot(cx+x*2, cy+y,  x,  y); plot(cx-x*2, cy+y, -x,  y)
		plot(cx+y*2, cy-x,  y, -x); plot(cx-y*2, cy-x, -y, -x)
		plot(cx+y*2, cy+x,  y,  x); plot(cx-y*2, cy+x, -y,  x)
		if d < 0 { d += 2*x + 3 } else { d += 2*(x-y) + 5; y-- }
		x++
	}
}

// Line draws a line using Bresenham + slope-based char selection.
func Line(g *Grid, x1, y1, x2, y2 int, c models.ColorName, ch Chars) {
	dx, dy := x2-x1, y2-y1
	adx, ady := iabs(dx), iabs(dy)
	var r rune
	switch {
	case ady == 0:            r = ch.LineH
	case adx == 0:            r = ch.LineV
	case (dx > 0) == (dy > 0): r = ch.LineD2
	default:                  r = ch.LineD1
	}
	sx, sy, err := isign(dx), isign(dy), adx-ady
	for x, y := x1, y1; ; {
		g.Set(x, y, r, c)
		if x == x2 && y == y2 { break }
		e2 := 2 * err
		if e2 > -ady { err -= ady; x += sx }
		if e2 < adx  { err += adx; y += sy }
	}
}

// Dot sets a single cell (used for freehand drawing).
func Dot(g *Grid, x, y int, c models.ColorName, ch rune) {
	g.Set(x, y, ch, c)
}

// RenderAll draws all saved shapes onto a fresh grid.
func RenderAll(w, h int, shapes []models.Shape, empty rune, ch Chars) *Grid {
	g := New(w, h, empty)
	for _, s := range shapes {
		DrawShape(g, s, ch)
	}
	return g
}

// DrawShape dispatches to the correct drawing function.
func DrawShape(g *Grid, s models.Shape, ch Chars) {
	c := s.Color
	if c == "" { c = models.ColGreen }
	switch s.Type {
	case models.ShapeRect:   Rect(g, s.X1, s.Y1, s.X2, s.Y2, c, s.Filled, ch)
	case models.ShapeCircle: Circle(g, s.X1, s.Y1, s.Radius, c, s.Filled, ch)
	case models.ShapeLine:   Line(g, s.X1, s.Y1, s.X2, s.Y2, c, ch)
	case models.ShapeFree:   Dot(g, s.X1, s.Y1, c, ch.Free)
	}
}

func iabs(x int) int  { if x < 0 { return -x }; return x }
func isign(x int) int { if x < 0 { return -1 } else if x > 0 { return 1 }; return 0 }