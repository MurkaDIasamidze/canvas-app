package main

import "fmt"

// ── App state ─────────────────────────────────────────────────

type screen int

const (
	scrMenu   screen = iota // tool/color picker
	scrCanvas               // main drawing surface
	scrHelp                 // help overlay
)

// App holds the entire application state.
// No external libraries, no database — everything lives here.
type App struct {
	scr screen

	// canvas dimensions (set to terminal size on start)
	cw, ch int

	// cursor position
	cx, cy int

	// drawn shapes (in-memory — undo just pops)
	shapes []Shape

	// drawing session
	drawing bool
	sx, sy  int // start point

	// current tool settings
	tool  ShapeKind
	fill  bool
	color Color

	// status bar message (shown for one frame)
	status    string
	statusErr bool

	// color picker cursor in menu
	colorIdx int

	// preset drawing function from config.go
	preset func(*Canvas)
}

// ── Entry point ───────────────────────────────────────────────

func main() {
	// terminal setup
	initRaw()
	defer func() {
		restoreRaw()
		clearScreen()
		showCursor()
		moveTo(1, 1)
		flush()
	}()

	hideCursor()
	clearScreen()
	flush()

	tw, th := termSize()
	cw, ch := tw, th-1 // reserve last row for status bar
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
	// set initial colorIdx
	for i, c := range allColors {
		if c == a.color { a.colorIdx = i; break }
	}

	// main loop
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
	// re-read terminal size each frame so resize just works
	tw, th := termSize()
	_ = tw
	W, H := a.cw, a.ch
	statusRow := th

	// --- build the composite canvas ---

	// 1. preset background layer (from config.go presetShapes)
	grid := NewCanvas(W, H)
	if a.preset != nil { a.preset(grid) }

	// 2. all committed shapes on top
	renderShapes(grid, a.shapes)

	// 2. live preview of the shape being drawn
	if a.drawing && a.tool != KindFree {
		prev := NewCanvas(W, H)
		prev.StrokeColor = a.color
		prev.FillColor   = a.color
		switch a.tool {
		case KindRect:
			x := imin(a.sx, a.cx); y := imin(a.sy, a.cy)
			w := iabs(a.cx-a.sx) + 1; h := iabs(a.cy-a.sy) + 1
			if a.fill { prev.FillRect(x, y, w, h) } else { prev.StrokeRect(x, y, w, h) }
		case KindCircle:
			r := idist(a.sx, a.sy, a.cx, a.cy)
			if a.fill { prev.FillCircle(a.sx, a.sy, r) } else { prev.StrokeCircle(a.sx, a.sy, r) }
		case KindLine:
			prev.DrawLine(a.sx, a.sy, a.cx, a.cy)
		}
		grid.DrawCanvas(prev, 0, 0)
	}

	// 3. freehand: paint cursor position while drawing
	if a.drawing && a.tool == KindFree {
		grid.StrokeColor = a.color
		grid.SetPixel(a.cx, a.cy)
	}

	// --- render to terminal ---
	for y := 0; y < H; y++ {
		moveTo(1, y+1) // rows are 1-based in ANSI
		for x := 0; x < W; x++ {
			cell    := grid.Get(x, y)
			isCur   := x == a.cx && y == a.cy
			isStart := a.drawing && x == a.sx && y == a.sy && a.tool != KindFree

			switch {
			case isCur && isStart:
				print_(BgYellow + FgBlack + "@" + Reset)
			case isCur:
				if cell.Empty {
					print_(BgGreen + FgBlack + " " + Reset)
				} else {
					print_(BgGreen + FgBlack + string(cell.Ch) + Reset)
				}
			case isStart:
				print_(BgYellow + FgBlack + "·" + Reset)
			case !cell.Empty:
				print_(colorANSI(cell.Color) + Bold + string(cell.Ch) + Reset)
			default:
				print_(" ")
			}
		}
		eraseLineRight()
	}

	// --- status bar ---
	toolLabel := map[ShapeKind]string{
		KindRect: "Rect", KindCircle: "Circle", KindLine: "Line", KindFree: "Free",
	}[a.tool]
	fillLabel := "Outline"
	if a.fill { fillLabel = "Filled" }

	var bar string
	if a.drawing && a.tool != KindFree {
		bar = FgBrightYellow + " DRAWING " + Reset +
			"move cursor → " + ky("Space") + " to place end  " +
			ky("Esc") + " cancel"
	} else if a.status != "" {
		col := FgBrightGreen
		if a.statusErr { col = FgBrightRed }
		bar = col + " " + a.status + Reset
		a.status = ""
	} else {
		bar = Dim + " " + toolLabel + " · " + fillLabel + " · " +
			colorANSI(a.color) + string(a.color) + Reset +
			Dim + "  │  " + Reset +
			ky("M") + " menu  " +
			ky("Space") + " place  " +
			ky("U") + " undo  " +
			ky("X") + " clear  " +
			ky("H") + " help  " +
			ky("Q") + " quit" +
			Dim + fmt.Sprintf("  │  %d,%d", a.cx, a.cy) + Reset
	}
	printAt(1, statusRow, BgBlack+bar+"\033[K")
}

// ── Menu screen ───────────────────────────────────────────────

func (a *App) drawMenu() {
	clearScreen()
	row := 1
	p := func(s string) { printAt(1, row, s+"\033[K"); row++ }

	p(BgBlue + FgBrightWhite + Bold + "  DRAWING OPTIONS  " + Reset +
		Dim + "  Enter/Esc = back to canvas" + Reset)
	p(Dim + repeatStr("─", 48) + Reset)
	p("")

	// ── Tool ──
	p(Bold + "  Tool" + Reset + Dim + "  (press the key)" + Reset)
	tools := []struct{ k, id ShapeKind; icon, label string }{
		{"r", KindRect,   "▭", "Rectangle"},
		{"c", KindCircle, "◯", "Circle"},
		{"l", KindLine,   "─", "Line"},
		{"f", KindFree,   "·", "Freehand"},
	}
	for _, t := range tools {
		key := string(t.k)
		if a.tool == t.id {
			p(FgBrightGreen + "  ▶ [" + key + "]  " + t.icon + "  " + t.label + Reset)
		} else {
			p(Dim + "    [" + key + "]  " + t.icon + "  " + t.label + Reset)
		}
	}
	p("")

	// ── Fill ──
	p(Bold + "  Fill" + Reset + Dim + "  (press T to toggle)" + Reset)
	if !a.fill {
		p(FgBrightCyan + "  ▶ [□] Outline" + Reset)
		p(Dim           + "    [■] Filled"  + Reset)
	} else {
		p(Dim           + "    [□] Outline" + Reset)
		p(FgBrightCyan + "  ▶ [■] Filled"  + Reset)
	}
	p("")

	// ── Color ──
	p(Bold + "  Color" + Reset + Dim + "  (← →  to cycle)" + Reset)
	swatch := "  "
	for i, c := range allColors {
		label := string([]rune(string(c))[:1])
		if i == a.colorIdx {
			swatch += colorANSI(c) + Bold + "[" + label + "]" + Reset + " "
		} else {
			swatch += Dim + colorANSI(c) + " " + label + " " + Reset + " "
		}
	}
	p(swatch)
	p("  " + colorANSI(a.color) + Bold + string(a.color) + Reset)
	p("")
	p(Dim + repeatStr("─", 48) + Reset)
	p("  " + ky("R/C/L/F") + " tool   " + ky("T") + " fill   " +
		ky("←→") + " color   " + ky("Enter") + " back")
	eraseDown()
}

// ── Help screen ───────────────────────────────────────────────

func (a *App) drawHelp() {
	clearScreen()
	row := 1
	p := func(s string) { printAt(1, row, s+"\033[K"); row++ }

	p(BgBlue + FgBrightWhite + Bold + "  HELP  " + Reset)
	p("")

	lines := []string{
		Bold + "  CANVAS CONTROLS" + Reset,
		"  " + ky("↑ ↓ ← →") + "   Move cursor one cell",
		"  " + ky("Space") + "      First press = set start  │  Second press = commit shape",
		"  " + ky("Enter") + "      Same as Space",
		"  " + ky("Esc") + "        Cancel drawing in progress",
		"  " + ky("M") + "          Open options menu (tool / fill / color)",
		"  " + ky("U") + "          Undo last shape",
		"  " + ky("X") + "          Clear all shapes",
		"  " + ky("H") + "          This help screen",
		"  " + ky("Q") + "          Quit",
		"",
		Bold + "  TOOLS" + Reset,
		"  R  Rectangle   outline uses ┌─┐│└┘  ·  filled uses █",
		"  C  Circle      outline uses ─ │ ╱ ╲  ·  filled uses █",
		"  L  Line        auto-selects ─ │ ╱ ╲ by angle",
		"  F  Freehand    stamp a single █ at the cursor",
		"",
		Bold + "  TIPS" + Reset,
		"  Canvas fills the whole terminal — resize to get more space.",
		"  All drawn shapes are kept in memory; close the app to reset.",
		"",
		"  Press any key to return...",
	}
	for _, l := range lines { p(l) }
	eraseDown()
}

// ── Key handling ──────────────────────────────────────────────

// handleKey processes one keypress. Returns false to quit.
func (a *App) handleKey(k int) bool {
	switch a.scr {
	case scrCanvas: return a.keyCanvas(k)
	case scrMenu:   return a.keyMenu(k)
	case scrHelp:   a.scr = scrCanvas // any key closes help
	}
	return true
}

func (a *App) keyCanvas(k int) bool {
	W, H := a.cw, a.ch
	switch k {
	// movement
	case KeyUp:    if a.cy > 0   { a.cy-- }
	case KeyDown:  if a.cy < H-1 { a.cy++ }
	case KeyLeft:  if a.cx > 0   { a.cx-- }
	case KeyRight: if a.cx < W-1 { a.cx++ }

	// place / draw
	case KeySpace, KeyEnter:
		if a.tool == KindFree {
			// freehand: each press commits one dot immediately
			a.shapes = append(a.shapes, Shape{
				Kind: KindFree, X1: a.cx, Y1: a.cy, Color: a.color,
			})
		} else if !a.drawing {
			a.drawing = true
			a.sx, a.sy = a.cx, a.cy
			a.status = fmt.Sprintf("Start (%d,%d) — move cursor, then Space/Enter", a.sx, a.sy)
		} else {
			a.commit()
			a.drawing = false
		}

	case KeyEsc:
		if a.drawing {
			a.drawing = false
			a.status = "Cancelled"
		}

	// actions
	case 'm', 'M': a.drawing = false; a.scr = scrMenu; clearScreen(); flush()
	case 'u', 'U': a.undo()
	case 'x', 'X': a.clearAll()
	case 'h', 'H': a.scr = scrHelp; clearScreen(); flush()
	case 'q', 'Q': return false
	}

	// clamp cursor
	if a.cx < 0   { a.cx = 0 }
	if a.cx >= W  { a.cx = W - 1 }
	if a.cy < 0   { a.cy = 0 }
	if a.cy >= H  { a.cy = H - 1 }
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
		clearScreen(); flush()
		a.scr = scrCanvas
	}
	return true
}

// ── Shape operations ──────────────────────────────────────────

// commit saves the current drawing session as a Shape.
func (a *App) commit() {
	s := Shape{
		Kind:  a.tool,
		X1: a.sx, Y1: a.sy,
		X2: a.cx, Y2: a.cy,
		Filled: a.fill,
		Color:  a.color,
	}
	if a.tool == KindCircle {
		s.Radius = idist(a.sx, a.sy, a.cx, a.cy)
	}
	a.shapes = append(a.shapes, s)
	fill := "outline"
	if s.Filled { fill = "filled" }
	a.status = fmt.Sprintf("Drew %s %s (%d,%d)→(%d,%d)", a.tool, fill, a.sx, a.sy, a.cx, a.cy)
}

// undo removes the last shape.
func (a *App) undo() {
	if len(a.shapes) == 0 {
		a.status = "Nothing to undo"
		return
	}
	a.shapes = a.shapes[:len(a.shapes)-1]
	a.status = "Undone"
}

// clearAll removes all shapes.
func (a *App) clearAll() {
	a.shapes = nil
	a.drawing = false
	a.status = "Cleared"
}