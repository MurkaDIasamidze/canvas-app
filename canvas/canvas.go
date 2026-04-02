package canvas

import (
	"canvas-tui/models"
	"math"
)

// Cell holds a character and its color
type Cell struct {
	Ch    rune
	Color models.ColorName
}

// Grid is a 2D cell buffer
type Grid struct {
	Width  int
	Height int
	Cells  [][]Cell
}

func NewGrid(w, h int) *Grid {
	cells := make([][]Cell, h)
	for i := range cells {
		cells[i] = make([]Cell, w)
		for j := range cells[i] {
			cells[i][j] = Cell{Ch: ' '}
		}
	}
	return &Grid{Width: w, Height: h, Cells: cells}
}

func (g *Grid) Set(x, y int, ch rune, color models.ColorName) {
	if x >= 0 && x < g.Width && y >= 0 && y < g.Height {
		g.Cells[y][x] = Cell{Ch: ch, Color: color}
	}
}

func (g *Grid) Get(x, y int) Cell {
	if x >= 0 && x < g.Width && y >= 0 && y < g.Height {
		return g.Cells[y][x]
	}
	return Cell{Ch: ' '}
}

// DrawShape draws a shape onto the grid with its color
func DrawShape(g *Grid, s models.Shape) {
	ch := rune('*')
	if len(s.Char) > 0 {
		ch = rune(s.Char[0])
	}
	color := s.Color
	if color == "" {
		color = models.ColorGreen
	}
	switch s.Type {
	case models.ShapeRect:
		DrawRect(g, s.X1, s.Y1, s.X2, s.Y2, ch, color, s.Filled)
	case models.ShapeCircle:
		DrawCircle(g, s.X1, s.Y1, s.Radius, ch, color, s.Filled)
	case models.ShapeLine:
		DrawLine(g, s.X1, s.Y1, s.X2, s.Y2, ch, color)
	}
}

func DrawRect(g *Grid, x1, y1, x2, y2 int, ch rune, color models.ColorName, filled bool) {
	if x1 > x2 { x1, x2 = x2, x1 }
	if y1 > y2 { y1, y2 = y2, y1 }
	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			if filled || x == x1 || x == x2 || y == y1 || y == y2 {
				g.Set(x, y, ch, color)
			}
		}
	}
}

func DrawCircle(g *Grid, cx, cy, r int, ch rune, color models.ColorName, filled bool) {
	if r <= 0 {
		g.Set(cx, cy, ch, color)
		return
	}
	rf := float64(r)
	for y := cy - r; y <= cy+r; y++ {
		for x := cx - r*2; x <= cx+r*2; x++ {
			dx := float64(x-cx) * 0.5
			dy := float64(y - cy)
			dist := math.Sqrt(dx*dx + dy*dy)
			if filled {
				if dist <= rf+0.3 {
					g.Set(x, y, ch, color)
				}
			} else {
				if dist >= rf-0.6 && dist <= rf+0.6 {
					g.Set(x, y, ch, color)
				}
			}
		}
	}
}

func DrawLine(g *Grid, x1, y1, x2, y2 int, ch rune, color models.ColorName) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx, sy := 1, 1
	if x1 > x2 { sx = -1 }
	if y1 > y2 { sy = -1 }
	err := dx - dy
	for {
		g.Set(x1, y1, ch, color)
		if x1 == x2 && y1 == y2 { break }
		e2 := 2 * err
		if e2 > -dy { err -= dy; x1 += sx }
		if e2 < dx  { err += dx; y1 += sy }
	}
}

// RenderAll rasterizes all shapes onto a fresh grid
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