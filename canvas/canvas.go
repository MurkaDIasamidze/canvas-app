package canvas

import (
	"canvas-tui/models"
	"math"
)

// Grid is a 2D character buffer
type Grid struct {
	Width  int
	Height int
	Cells  [][]rune
}

func NewGrid(w, h int) *Grid {
	cells := make([][]rune, h)
	for i := range cells {
		cells[i] = make([]rune, w)
		for j := range cells[i] {
			cells[i][j] = ' '
		}
	}
	return &Grid{Width: w, Height: h, Cells: cells}
}

func (g *Grid) Set(x, y int, ch rune) {
	if x >= 0 && x < g.Width && y >= 0 && y < g.Height {
		g.Cells[y][x] = ch
	}
}

func (g *Grid) Get(x, y int) rune {
	if x >= 0 && x < g.Width && y >= 0 && y < g.Height {
		return g.Cells[y][x]
	}
	return ' '
}

// DrawShape draws a single shape onto the grid
func DrawShape(g *Grid, s models.Shape) {
	ch := '*'
	if len(s.Char) > 0 {
		ch = rune(s.Char[0])
	}
	switch s.Type {
	case models.ShapeRect:
		DrawRect(g, s.X1, s.Y1, s.X2, s.Y2, ch, s.Filled)
	case models.ShapeCircle:
		DrawCircle(g, s.X1, s.Y1, s.Radius, ch, s.Filled)
	case models.ShapeLine:
		DrawLine(g, s.X1, s.Y1, s.X2, s.Y2, ch)
	}
}

// DrawRect draws a rectangle (outline or filled)
func DrawRect(g *Grid, x1, y1, x2, y2 int, ch rune, filled bool) {
	if x1 > x2 { x1, x2 = x2, x1 }
	if y1 > y2 { y1, y2 = y2, y1 }
	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			if filled || x == x1 || x == x2 || y == y1 || y == y2 {
				g.Set(x, y, ch)
			}
		}
	}
}

// DrawCircle draws a circle (outline or filled)
// Uses aspect-ratio correction (x*0.5) since terminal chars are taller than wide
func DrawCircle(g *Grid, cx, cy, r int, ch rune, filled bool) {
	if r <= 0 {
		g.Set(cx, cy, ch)
		return
	}
	rf := float64(r)
	for y := cy - r; y <= cy+r; y++ {
		for x := cx - r*2; x <= cx+r*2; x++ {
			dx := float64(x-cx) * 0.5
			dy := float64(y - cy)
			dist := math.Sqrt(dx*dx + dy*dy)
			if filled {
				if dist <= rf {
					g.Set(x, y, ch)
				}
			} else {
				if dist >= rf-0.5 && dist <= rf+0.5 {
					g.Set(x, y, ch)
				}
			}
		}
	}
}

// DrawLine draws a line using Bresenham's algorithm
func DrawLine(g *Grid, x1, y1, x2, y2 int, ch rune) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx, sy := 1, 1
	if x1 > x2 { sx = -1 }
	if y1 > y2 { sy = -1 }
	err := dx - dy
	for {
		g.Set(x1, y1, ch)
		if x1 == x2 && y1 == y2 { break }
		e2 := 2 * err
		if e2 > -dy { err -= dy; x1 += sx }
		if e2 < dx  { err += dx; y1 += sy }
	}
}

// RenderAll draws all shapes onto a fresh grid
func RenderAll(width, height int, shapes []models.Shape) *Grid {
	g := NewGrid(width, height)
	for _, s := range shapes {
		DrawShape(g, s)
	}
	return g
}

func abs(x int) int {
	if x < 0 { return -x }
	return x
}