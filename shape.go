package main

// ShapeKind identifies the type of a drawn shape.
type ShapeKind string

const (
	KindRect   ShapeKind = "rect"
	KindCircle ShapeKind = "circle"
	KindLine   ShapeKind = "line"
	KindFree   ShapeKind = "free"
)

// Shape is a fully described drawing command stored in memory.
// All drawn shapes are kept in a slice; undo just pops the last one.
type Shape struct {
	Kind           ShapeKind
	X1, Y1, X2, Y2 int
	Radius         int
	Filled         bool
	Color          Color
}

// renderShapes draws every shape from the list onto canvas c.
func renderShapes(c *Canvas, shapes []Shape) {
	for _, s := range shapes {
		drawShape(c, s)
	}
}

// drawShape draws one Shape onto canvas c.
func drawShape(c *Canvas, s Shape) {
	c.StrokeColor = s.Color
	c.FillColor   = s.Color
	switch s.Kind {
	case KindRect:
		x := imin(s.X1, s.X2)
		y := imin(s.Y1, s.Y2)
		w := iabs(s.X2-s.X1) + 1
		h := iabs(s.Y2-s.Y1) + 1
		if s.Filled { c.FillRect(x, y, w, h) } else { c.StrokeRect(x, y, w, h) }
	case KindCircle:
		if s.Filled { c.FillCircle(s.X1, s.Y1, s.Radius) } else { c.StrokeCircle(s.X1, s.Y1, s.Radius) }
	case KindLine:
		c.DrawLine(s.X1, s.Y1, s.X2, s.Y2)
	case KindFree:
		c.set(s.X1, s.Y1, '█', s.Color)
	}
}