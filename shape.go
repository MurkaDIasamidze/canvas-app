package main

type ShapeKind string

const (
	KindRect   ShapeKind = "rect"
	KindCircle ShapeKind = "circle"
	KindLine   ShapeKind = "line"
	KindFree   ShapeKind = "free"
)

type Shape struct {
	Kind            ShapeKind
	X1, Y1, X2, Y2 int
	Radius          int
	Filled          bool
	Color           Color
}

func renderShapes(ctx *Ctx, shapes []Shape) {
	for _, s := range shapes {
		ctx.strokeStyle = s.Color
		ctx.fillStyle   = s.Color
		switch s.Kind {
		case KindRect:
			x, y := imin(s.X1, s.X2), imin(s.Y1, s.Y2)
			w, h := iabs(s.X2-s.X1)+1, iabs(s.Y2-s.Y1)+1
			if s.Filled { fillRect(ctx, x, y, w, h) } else { strokeRect(ctx, x, y, w, h) }
		case KindCircle:
			if s.Filled { fillCircle(ctx, s.X1, s.Y1, s.Radius) } else { strokeCircle(ctx, s.X1, s.Y1, s.Radius) }
		case KindLine:
			drawLine(ctx, s.X1, s.Y1, s.X2, s.Y2)
		case KindFree:
			ctx.set(s.X1, s.Y1, freeChar, s.Color)
		}
	}
}