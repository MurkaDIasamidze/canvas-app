package canvas

import (
	"canvas-tui/models"
	"math"
)

type Cell struct {
	Ch    rune
	Color models.ColorName
}

type Grid struct {
	Width, Height int
	Cells         [][]Cell
}

func NewGrid(w, h int) *Grid {
	cells := make([][]Cell, h)
	for i := range cells {
		cells[i] = make([]Cell, w)
		for j := range cells[i] { cells[i][j] = Cell{Ch: ' '} }
	}
	return &Grid{w, h, cells}
}

func (g *Grid) Set(x, y int, ch rune, c models.ColorName) {
	if x >= 0 && x < g.Width && y >= 0 && y < g.Height {
		g.Cells[y][x] = Cell{ch, c}
	}
}

func (g *Grid) Get(x, y int) Cell {
	if x >= 0 && x < g.Width && y >= 0 && y < g.Height { return g.Cells[y][x] }
	return Cell{Ch: ' '}
}

func DrawShape(g *Grid, s models.Shape) {
	c := s.Color
	if c == "" { c = models.ColorGreen }
	switch s.Type {
	case models.ShapeRect:   drawRect(g, s.X1, s.Y1, s.X2, s.Y2, c, s.Filled)
	case models.ShapeCircle: drawCircle(g, s.X1, s.Y1, s.Radius, c, s.Filled)
	case models.ShapeLine:   drawLine(g, s.X1, s.Y1, s.X2, s.Y2, c)
	}
}

// drawRect — outline uses ┌─┐│└┘, filled uses █
func drawRect(g *Grid, x1, y1, x2, y2 int, c models.ColorName, filled bool) {
	if x1 > x2 { x1, x2 = x2, x1 }
	if y1 > y2 { y1, y2 = y2, y1 }
	if filled {
		for y := y1; y <= y2; y++ {
			for x := x1; x <= x2; x++ { g.Set(x, y, '█', c) }
		}
		return
	}
	for x := x1 + 1; x < x2; x++ { g.Set(x, y1, '─', c); g.Set(x, y2, '─', c) }
	for y := y1 + 1; y < y2; y++ { g.Set(x1, y, '│', c); g.Set(x2, y, '│', c) }
	g.Set(x1, y1, '┌', c); g.Set(x2, y1, '┐', c)
	g.Set(x1, y2, '└', c); g.Set(x2, y2, '┘', c)
}

// drawCircle — uses Bresenham midpoint algorithm for a crisp single-pixel border.
// Each point on the circle is assigned ─ │ ╱ ╲ based on the local tangent angle,
// so the outline looks like a smooth curve made of box-drawing characters.
// Filled uses █.
func drawCircle(g *Grid, cx, cy, r int, c models.ColorName, filled bool) {
	if r <= 0 { g.Set(cx, cy, '█', c); return }

	if filled {
		rf := float64(r)
		for y := cy - r; y <= cy+r; y++ {
			for x := cx - r*2; x <= cx+r*2; x++ {
				dx := float64(x-cx) * 0.5
				dy := float64(y - cy)
				if math.Sqrt(dx*dx+dy*dy) <= rf+0.3 { g.Set(x, y, '█', c) }
			}
		}
		return
	}

	// Outline: Bresenham midpoint circle, then pick char from tangent angle
	// The tangent at angle θ is perpendicular to the radius.
	// We compute the 8 octants and map to the best char.
	plotCircle := func(px, py, ox, oy int) {
		// ox,oy is the offset from center — the radius vector points outward.
		// Tangent is perpendicular: (-oy, ox). Use the tangent to pick char.
		// Account for terminal aspect ratio (chars ~2x taller than wide).
		tx := float64(-oy)
		ty := float64(ox) * 0.5
		angle := math.Atan2(ty, tx) * 180 / math.Pi
		if angle < 0 { angle += 180 }
		var ch rune
		switch {
		case angle < 22.5 || angle >= 157.5: ch = '─'
		case angle < 67.5:                   ch = '╱'
		case angle < 112.5:                  ch = '│'
		default:                             ch = '╲'
		}
		g.Set(px, py, ch, c)
	}

	x, y := 0, r
	d := 1 - r
	for x <= y {
		// All 8 symmetry points — terminal x is doubled for aspect ratio
		plotCircle(cx+x*2, cy-y,  x,  -y)
		plotCircle(cx-x*2, cy-y,  -x, -y)
		plotCircle(cx+x*2, cy+y,  x,  y)
		plotCircle(cx-x*2, cy+y,  -x, y)
		plotCircle(cx+y*2, cy-x,  y,  -x)
		plotCircle(cx-y*2, cy-x,  -y, -x)
		plotCircle(cx+y*2, cy+x,  y,  x)
		plotCircle(cx-y*2, cy+x,  -y, x)
		if d < 0 { d += 2*x + 3 } else { d += 2*(x-y) + 5; y-- }
		x++
	}
}

// drawLine — Bresenham, picks ─ │ ╱ ╲ from overall slope
func drawLine(g *Grid, x1, y1, x2, y2 int, c models.ColorName) {
	dx, dy := x2-x1, y2-y1
	adx, ady := abs(dx), abs(dy)
	var ch rune
	switch {
	case ady == 0:                               ch = '─'
	case adx == 0:                               ch = '│'
	case (dx > 0) == (dy > 0):                  ch = '╲'
	default:                                     ch = '╱'
	}
	sx, sy, err := sign(dx), sign(dy), adx-ady
	for x, y := x1, y1; ; {
		g.Set(x, y, ch, c)
		if x == x2 && y == y2 { break }
		e2 := 2 * err
		if e2 > -ady { err -= ady; x += sx }
		if e2 < adx  { err += adx; y += sy }
	}
}

func RenderAll(w, h int, shapes []models.Shape) *Grid {
	g := NewGrid(w, h)
	for _, s := range shapes { DrawShape(g, s) }
	return g
}

// DrawRect / DrawCircle / DrawLine exported for preview in app.go
func DrawRect(g *Grid, x1, y1, x2, y2 int, c models.ColorName, filled bool) {
	drawRect(g, x1, y1, x2, y2, c, filled)
}
func DrawCircle(g *Grid, cx, cy, r int, c models.ColorName, filled bool) {
	drawCircle(g, cx, cy, r, c, filled)
}
func DrawLine(g *Grid, x1, y1, x2, y2 int, c models.ColorName) {
	drawLine(g, x1, y1, x2, y2, c)
}

func abs(x int) int { if x < 0 { return -x }; return x }
func sign(x int) int { if x < 0 { return -1 } else if x > 0 { return 1 }; return 0 }