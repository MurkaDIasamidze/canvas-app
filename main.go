package main

import "fmt"

type screen int

const (
	scrMenu   screen = iota
	scrCanvas
	scrHelp
)

type App struct {
	scr       screen
	cw, ch    int
	cx, cy    int
	shapes    []Shape
	drawing   bool
	sx, sy    int
	tool      ShapeKind
	fill      bool
	color     Color
	status    string
	statusErr bool
	colorIdx  int
	preset    func(*Ctx)
}

func main() {
	initRaw()
	defer func() {
		restoreRaw()
		clearScreen()
		showCursor()
		cursorTo(1, 1)
		flush()
	}()

	hideCursor()
	clearScreen()
	flush()

	tw, th := termSize()
	cw, ch := tw, th-1
	if !fullConsole {
		cw, ch = canvasW, canvasH
	}

	a := &App{
		scr:    scrCanvas,
		cw:     cw,
		ch:     ch,
		cx:     cw / 2,
		cy:     ch / 2,
		tool:   defaultTool,
		fill:   defaultFill,
		color:  defaultColor,
		preset: presetShapes,
	}
	for i, c := range allColors {
		if c == a.color { a.colorIdx = i; break }
	}

	for {
		a.draw()
		if !a.handleKey(readKey()) { break }
	}
}

// ── Draw dispatcher ───────────────────────────────────────────

func (a *App) draw() {
	switch a.scr {
	case scrCanvas: a.drawCanvas()
	case scrMenu:   a.drawMenu()
	case scrHelp:   a.drawHelp()
	}
	flush()
}

// ── Canvas screen ─────────────────────────────────────────────

func (a *App) drawCanvas() {
	_, th := termSize()
	W, H  := a.cw, a.ch

	cv  := newCanvas(W, H)
	ctx := getContext(cv)

	if a.preset != nil { a.preset(ctx) }
	renderShapes(ctx, a.shapes)

	// live preview
	if a.drawing && a.tool != KindFree {
		pv   := newCanvas(W, H)
		pCtx := getContext(pv)
		pCtx.strokeStyle = a.color
		pCtx.fillStyle   = a.color
		switch a.tool {
		case KindRect:
			x := imin(a.sx, a.cx); y := imin(a.sy, a.cy)
			w := iabs(a.cx-a.sx) + 1; h := iabs(a.cy-a.sy) + 1
			if a.fill { fillRect(pCtx, x, y, w, h) } else { strokeRect(pCtx, x, y, w, h) }
		case KindCircle:
			r := idist(a.sx, a.sy, a.cx, a.cy)
			if a.fill { fillCircle(pCtx, a.sx, a.sy, r) } else { strokeCircle(pCtx, a.sx, a.sy, r) }
		case KindLine:
			drawLine(pCtx, a.sx, a.sy, a.cx, a.cy)
		}
		drawCanvas(ctx, pCtx, 0, 0)
	}
	if a.drawing && a.tool == KindFree {
		ctx.strokeStyle = a.color
		setPixel(ctx, a.cx, a.cy)
	}

	// render cells to terminal
	for y := 0; y < H; y++ {
		cursorTo(1, y+1)
		for x := 0; x < W; x++ {
			cl      := ctx.getCell(x, y)
			isCur   := x == a.cx && y == a.cy
			isStart := a.drawing && x == a.sx && y == a.sy && a.tool != KindFree
			switch {
			case isCur && isStart:
				print_(BgYellow + FgBlack + "@" + Reset)
			case isCur:
				if cl.empty { print_(BgGreen + FgBlack + " " + Reset) } else { print_(BgGreen + FgBlack + string(cl.ch) + Reset) }
			case isStart:
				print_(BgYellow + FgBlack + "·" + Reset)
			case !cl.empty:
				print_(colorANSI(cl.color) + Bold + string(cl.ch) + Reset)
			default:
				print_(" ")
			}
		}
		eraseLineRight()
	}

	// status bar
	toolLabel := map[ShapeKind]string{KindRect: "Rect", KindCircle: "Circle", KindLine: "Line", KindFree: "Free"}[a.tool]
	fillLabel  := "Outline"; if a.fill { fillLabel = "Filled" }
	var bar string
	switch {
	case a.drawing && a.tool != KindFree:
		bar = FgBrightYellow + " DRAWING " + Reset + "move → " + ky("Space") + " place  " + ky("Esc") + " cancel"
	case a.status != "":
		col := FgBrightGreen; if a.statusErr { col = FgBrightRed }
		bar = col + " " + a.status + Reset
		a.status = ""
	default:
		bar = Dim + " " + toolLabel + " · " + fillLabel + " · " + colorANSI(a.color) + string(a.color) + Reset +
			Dim + "  │  " + Reset +
			ky("M") + " menu  " + ky("Space") + " place  " + ky("U") + " undo  " +
			ky("X") + " clear  " + ky("H") + " help  " + ky("Q") + " quit" +
			Dim + fmt.Sprintf("  │  %d,%d", a.cx, a.cy) + Reset
	}
	printAt(1, th, BgBlack+bar+"\033[K")
}

// ── Menu screen ───────────────────────────────────────────────

func (a *App) drawMenu() {
	clearScreen()
	row := 1
	p := func(s string) { printAt(1, row, s+"\033[K"); row++ }

	p(BgBlue + FgBrightWhite + Bold + "  OPTIONS  " + Reset + Dim + "  Enter/Esc = back" + Reset)
	p(Dim + repeatStr("─", 44) + Reset)
	p("")
	p(Bold + "  Tool" + Reset)
	for _, t := range []struct{ k, id ShapeKind; label string }{
		{"r", KindRect, "▭  Rectangle"},
		{"c", KindCircle, "◯  Circle"},
		{"l", KindLine, "─  Line"},
		{"f", KindFree, "·  Freehand"},
	} {
		if a.tool == t.id { p(FgBrightGreen + "  ▶ [" + string(t.k) + "]  " + t.label + Reset) } else { p(Dim + "    [" + string(t.k) + "]  " + t.label + Reset) }
	}
	p("")
	p(Bold + "  Fill" + Reset + Dim + "  T to toggle" + Reset)
	if !a.fill { p(FgBrightCyan+"  ▶ [□] Outline"+Reset); p(Dim+"    [■] Filled"+Reset) } else { p(Dim+"    [□] Outline"+Reset); p(FgBrightCyan+"  ▶ [■] Filled"+Reset) }
	p("")
	p(Bold + "  Color" + Reset + Dim + "  ← →" + Reset)
	swatch := "  "
	for i, c := range allColors {
		label := string([]rune(string(c))[:1])
		if i == a.colorIdx { swatch += colorANSI(c) + Bold + "[" + label + "]" + Reset + " " } else { swatch += Dim + colorANSI(c) + " " + label + " " + Reset + " " }
	}
	p(swatch)
	p("  " + colorANSI(a.color) + Bold + string(a.color) + Reset)
	p(""); p(Dim + repeatStr("─", 44) + Reset)
	p("  " + ky("R/C/L/F") + " tool   " + ky("T") + " fill   " + ky("←→") + " color   " + ky("Enter") + " back")
	eraseDown()
}

// ── Help screen ───────────────────────────────────────────────

func (a *App) drawHelp() {
	clearScreen()
	row := 1
	p := func(s string) { printAt(1, row, s+"\033[K"); row++ }
	p(BgBlue + FgBrightWhite + Bold + "  HELP  " + Reset)
	p("")
	for _, l := range []string{
		Bold + "  Controls" + Reset,
		"  " + ky("↑↓←→") + "  move   " + ky("Space") + " 1st=start 2nd=commit   " + ky("Esc") + " cancel",
		"  " + ky("M") + " menu   " + ky("U") + " undo   " + ky("X") + " clear   " + ky("H") + " help   " + ky("Q") + " quit",
		"",
		Bold + "  Tools" + Reset,
		"  R  Rectangle   C  Circle   L  Line   F  Freehand",
		"",
		Bold + "  Canvas API  (use in config.go presetShapes)" + Reset,
		"  strokeRect(ctx,x,y,w,h)      fillRect(ctx,x,y,w,h)",
		"  strokeCircle(ctx,cx,cy,r)     fillCircle(ctx,cx,cy,r)",
		"  strokeEllipse(ctx,cx,cy,rx,ry) fillEllipse(ctx,cx,cy,rx,ry)",
		"  strokeTriangle(ctx,x1,y1,...) fillTriangle(ctx,x1,y1,...)",
		"  drawLine(ctx,x1,y1,x2,y2)    setPixel(ctx,x,y)",
		"  fillText(ctx,\"s\",x,y)        strokeText(ctx,\"s\",x,y)",
		"  beginPath  moveTo  lineTo  closePath  stroke  fill",
		"",
		"  Press any key to return...",
	} { p(l) }
	eraseDown()
}

// ── Key handling ──────────────────────────────────────────────

func (a *App) handleKey(k int) bool {
	switch a.scr {
	case scrCanvas: return a.keyCanvas(k)
	case scrMenu:   return a.keyMenu(k)
	case scrHelp:   a.scr = scrCanvas
	}
	return true
}

func (a *App) keyCanvas(k int) bool {
	W, H := a.cw, a.ch
	switch k {
	case KeyUp:    if a.cy > 0   { a.cy-- }
	case KeyDown:  if a.cy < H-1 { a.cy++ }
	case KeyLeft:  if a.cx > 0   { a.cx-- }
	case KeyRight: if a.cx < W-1 { a.cx++ }
	case KeySpace, KeyEnter:
		if a.tool == KindFree {
			a.shapes = append(a.shapes, Shape{Kind: KindFree, X1: a.cx, Y1: a.cy, Color: a.color})
		} else if !a.drawing {
			a.drawing = true
			a.sx, a.sy = a.cx, a.cy
			a.status = fmt.Sprintf("Start (%d,%d) — move, then Space", a.sx, a.sy)
		} else {
			a.commit()
			a.drawing = false
		}
	case KeyEsc:
		if a.drawing { a.drawing = false; a.status = "Cancelled" }
	case 'm', 'M': a.drawing = false; a.scr = scrMenu; clearScreen(); flush()
	case 'u', 'U': a.undo()
	case 'x', 'X': a.clearAll()
	case 'h', 'H': a.scr = scrHelp; clearScreen(); flush()
	case 'q', 'Q': return false
	}
	if a.cx < 0  { a.cx = 0 };  if a.cx >= W { a.cx = W-1 }
	if a.cy < 0  { a.cy = 0 };  if a.cy >= H { a.cy = H-1 }
	return true
}

func (a *App) keyMenu(k int) bool {
	switch k {
	case 'r', 'R': a.tool = KindRect
	case 'c', 'C': a.tool = KindCircle
	case 'l', 'L': a.tool = KindLine
	case 'f', 'F': a.tool = KindFree
	case 't', 'T': a.fill = !a.fill
	case KeyLeft:
		a.colorIdx = (a.colorIdx - 1 + len(allColors)) % len(allColors)
		a.color = allColors[a.colorIdx]
	case KeyRight:
		a.colorIdx = (a.colorIdx + 1) % len(allColors)
		a.color = allColors[a.colorIdx]
	case KeyEnter, KeyEsc:
		clearScreen(); flush(); a.scr = scrCanvas
	}
	return true
}

// ── Shape operations ──────────────────────────────────────────

func (a *App) commit() {
	s := Shape{Kind: a.tool, X1: a.sx, Y1: a.sy, X2: a.cx, Y2: a.cy, Filled: a.fill, Color: a.color}
	if a.tool == KindCircle { s.Radius = idist(a.sx, a.sy, a.cx, a.cy) }
	a.shapes = append(a.shapes, s)
	fill := "outline"; if s.Filled { fill = "filled" }
	a.status = fmt.Sprintf("Drew %s %s (%d,%d)→(%d,%d)", a.tool, fill, a.sx, a.sy, a.cx, a.cy)
}

func (a *App) undo() {
	if len(a.shapes) == 0 { a.status = "Nothing to undo"; return }
	a.shapes = a.shapes[:len(a.shapes)-1]
	a.status = "Undone"
}

func (a *App) clearAll() {
	a.shapes = nil; a.drawing = false; a.status = "Cleared"
}