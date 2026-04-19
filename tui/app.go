package tui

import (
	"canvas-tui/canvas"
	"canvas-tui/db"
	"canvas-tui/models"
	"fmt"
	"strings"
)

// ── Screens ───────────────────────────────────────────────────

type screen int

const (
	scrProjects screen = iota
	scrCanvas
	scrNew
	scrMenu
	scrHelp
)

// ── App ───────────────────────────────────────────────────────

type App struct {
	// data
	projects   []models.Project
	project    *models.Project
	shapes     []models.Shape

	// screen
	scr        screen

	// canvas dimensions (set from config or terminal size)
	cw, ch     int

	// cursor
	cx, cy     int

	// drawing state
	drawing    bool
	sx, sy     int

	// tool state
	tool       string           // "rect" "circle" "line" "free"
	fill       bool
	color      models.ColorName

	// ui state
	projCursor int
	inputBuf   string
	status     string
	statusErr  bool

	// char config converted from global Cfg
	chars      canvas.Chars
}

// ── Entry point ───────────────────────────────────────────────

// Config holds all startup parameters passed from main.go.
type Config struct {
	FullConsole  bool
	CanvasW      int
	CanvasH      int
	ShowRulers   bool
	DefaultTool  string
	DefaultFill  bool
	DefaultColor string
	Chars        canvas.Chars
}

func Run(cfg Config) {
	InitTerminal()
	HideCursor()
	ClearScreen()
	Flush()

	tw, th := TermSize()

	cw, ch := tw, th-1 // -1 for status bar at bottom
	if !cfg.FullConsole {
		cw, ch = cfg.CanvasW, cfg.CanvasH
	}

	color := models.ColorName(cfg.DefaultColor)
	if color == "" { color = models.ColGreen }

	a := &App{
		scr:   scrProjects,
		cw:    cw,
		ch:    ch,
		tool:  cfg.DefaultTool,
		fill:  cfg.DefaultFill,
		color: color,
		chars: cfg.Chars,
	}

	db.Connect()
	a.loadProjects()

	defer func() {
		ClearScreen()
		ShowCursor()
		MoveTo(1, 1)
		Flush()
		restoreTerminal()
	}()

	initRawMode()

	for a.draw(); ; a.draw() {
		if !a.key(ReadKey()) { return }
	}
}

// ── Draw dispatcher ───────────────────────────────────────────

func (a *App) draw() {
	switch a.scr {
	case scrProjects: a.drawProjects()
	case scrCanvas:   a.drawCanvas()
	case scrNew:      a.drawNew()
	case scrMenu:     a.drawMenu()
	case scrHelp:     a.drawHelp()
	}
	Flush()
}

// ── Project list — occupies full terminal ─────────────────────

func (a *App) drawProjects() {
	ClearScreen()
	row := 1
	p := func(s string) { PrintAt(1, row, s+"\033[K"); row++ }

	p(BgBlue + FgBrightWhite + Bold + "  ASCII CANVAS  " + Reset +
		Dim + "  Q=quit" + Reset)
	p(Dim + Repeat("─", 50) + Reset)
	p("")
	if len(a.projects) == 0 {
		p(Dim + "  No projects — press N to create one" + Reset)
	}
	for i, pr := range a.projects {
		if i == a.projCursor {
			p(BgBlue + FgBrightWhite + fmt.Sprintf("  ▶  %-24s  %dx%d", pr.Name, pr.Width, pr.Height) + Reset)
		} else {
			p(fmt.Sprintf("     %-24s  "+Dim+"%dx%d"+Reset, pr.Name, pr.Width, pr.Height))
		}
	}
	p("")
	p(Dim + Repeat("─", 50) + Reset)
	p("  " + ky("↑↓") + " move  " + ky("Enter") + " open  " +
		ky("N") + " new  " + ky("D") + " delete  " + ky("Q") + " quit")
	EraseDown()
}

// ── New project form ──────────────────────────────────────────

func (a *App) drawNew() {
	ClearScreen()
	PrintAt(1, 1, BgBlue+FgBrightWhite+Bold+"  NEW PROJECT  "+Reset)
	PrintAt(1, 3, "  Name: "+FgBrightGreen+a.inputBuf+FgBrightYellow+"█"+Reset)
	PrintAt(1, 5, Dim+"  Enter = create  |  Esc = cancel"+Reset)
	EraseDown()
}

// ── Shape / tool menu ─────────────────────────────────────────

func (a *App) drawMenu() {
	ClearScreen()
	row := 1
	p := func(s string) { PrintAt(1, row, s+"\033[K"); row++ }

	p(BgBlue + FgBrightWhite + Bold + "  OPTIONS  " + Reset +
		Dim + "  Enter/Esc = back" + Reset)
	p(Dim + Repeat("─", 40) + Reset); p("")

	p(Bold + "  Tool" + Reset + Dim + "  (press key)" + Reset)
	for _, t := range []struct{ k, id, icon, label string }{
		{"R", "rect",   "▭", "Rectangle"},
		{"C", "circle", "◯", "Circle"},
		{"L", "line",   "─", "Line"},
		{"F", "free",   "·", "Freehand"},
	} {
		if a.tool == t.id {
			p(FgBrightGreen + "  ▶ [" + t.k + "] " + t.icon + "  " + t.label + Reset)
		} else {
			p(Dim + "    [" + t.k + "] " + t.icon + "  " + t.label + Reset)
		}
	}
	p("")

	p(Bold + "  Fill" + Reset + Dim + "  (press T)" + Reset)
	if !a.fill {
		p(FgBrightCyan + "  ▶ [□] Outline" + Reset)
		p(Dim           + "    [■] Filled"  + Reset)
	} else {
		p(Dim           + "    [□] Outline" + Reset)
		p(FgBrightCyan + "  ▶ [■] Filled"  + Reset)
	}
	p("")

	p(Bold + "  Color" + Reset + Dim + "  (← →)" + Reset)
	row2 := "  "
	for _, c := range models.AllColors {
		ch := string([]rune(string(c))[:1])
		if c == a.color {
			row2 += ColorANSI(c) + Bold + "[" + ch + "]" + Reset + " "
		} else {
			row2 += Dim + ColorANSI(c) + " " + ch + " " + Reset + " "
		}
	}
	p(row2)
	p("  " + ColorANSI(a.color) + Bold + string(a.color) + Reset)
	p(""); p(Dim + Repeat("─", 40) + Reset)
	p("  " + ky("R/C/L/F") + " tool  " + ky("T") + " fill  " +
		ky("←→") + " color  " + ky("Enter") + " back")
	EraseDown()
}

// ── Help screen ───────────────────────────────────────────────

func (a *App) drawHelp() {
	ClearScreen()
	row := 1
	p := func(s string) { PrintAt(1, row, s+"\033[K"); row++ }
	p(BgBlue + FgBrightWhite + Bold + "  HELP  " + Reset)
	for _, l := range []string{
		"", Bold + "  CANVAS CONTROLS" + Reset,
		"  " + ky("↑↓←→") + "   Move cursor (1 cell)",
		"  " + ky("Space") + "    1st press = set start  |  2nd press = draw",
		"  " + ky("O") + "       Options (tool / fill / color)",
		"  " + ky("U") + "       Undo last shape",
		"  " + ky("X") + "       Clear all shapes",
		"  " + ky("H") + "       Toggle this help overlay",
		"  " + ky("Q") + "       Back to project list",
		"  " + ky("Esc") + "     Cancel current drawing",
		"", Bold + "  TOOLS" + Reset,
		"  R  Rectangle   (outline uses box chars, filled uses █)",
		"  C  Circle      (outline uses ─ │ ╱ ╲, filled uses █)",
		"  L  Line        (auto char: ─ │ ╱ ╲ by angle)",
		"  F  Freehand    (paint single cells while moving)",
		"", Bold + "  TIPS" + Reset,
		"  Canvas fills the entire terminal window",
		"  Resize your terminal to get more space",
		"  Edit config.go to change all chars and behaviour",
		"", "  Press any key to return...",
	} { p(l) }
	EraseDown()
}

// ── Canvas — full terminal, no fixed box ──────────────────────
//
// The canvas IS the terminal. Row 1 = top of screen.
// The last row is reserved for the status bar.
// No borders, no rulers unless ShowRulers=true in config.
// Cursor and drawing overlays are painted on top of the grid.

func (a *App) drawCanvas() {
	if a.project == nil { a.scr = scrProjects; return }

	// detect terminal size every frame so resize just works
	tw, th := TermSize()
	_ = tw

	W, H := a.cw, a.ch
	statusRow := th // last terminal row = status bar

	// Build grid from saved shapes
	g := canvas.RenderAll(W, H, a.shapes, ' ', a.chars)

	// Freehand: paint current cursor position while drawing
	if a.drawing && a.tool == "free" {
		canvas.Dot(g, a.cx, a.cy, a.color, a.chars.Free)
	}

	// Live preview for other tools
	if a.drawing && a.tool != "free" {
		p := canvas.New(W, H, ' ')
		switch a.tool {
		case "rect":
			canvas.Rect(p, a.sx, a.sy, a.cx, a.cy, a.color, a.fill, a.chars)
		case "circle":
			canvas.Circle(p, a.sx, a.sy, idist(a.sx, a.sy, a.cx, a.cy), a.color, a.fill, a.chars)
		case "line":
			canvas.Line(p, a.sx, a.sy, a.cx, a.cy, a.color, a.chars)
		}
		for y := 0; y < H; y++ {
			for x := 0; x < W; x++ {
				if c := p.Get(x, y); !c.Empty {
					g.Set(x, y, c.Ch, c.Color)
				}
			}
		}
	}

	// ── Render every cell directly onto the terminal ──────────
	// Row 1..H = canvas rows, each cell at its exact screen position
	for y := 0; y < H; y++ {
		screenRow := y + 1 // terminal rows are 1-based

		// Move to start of this row once, then print chars consecutively
		// to avoid hundreds of MoveTo calls per row (much faster)
		MoveTo(1, screenRow)
		for x := 0; x < W; x++ {
			cell    := g.Get(x, y)
			isCur   := x == a.cx && y == a.cy
			isStart := a.drawing && x == a.sx && y == a.sy && a.tool != "free"

			switch {
			case isCur && isStart:
				Print(BgYellow + FgBlack + "@" + Reset)
			case isCur:
				if cell.Empty {
					Print(BgGreen + FgBlack + " " + Reset)
				} else {
					Print(BgGreen + FgBlack + string(cell.Ch) + Reset)
				}
			case isStart:
				Print(BgYellow + FgBlack + "·" + Reset)
			case !cell.Empty:
				Print(ColorANSI(cell.Color) + Bold + string(cell.Ch) + Reset)
			default:
				Print(" ")
			}
		}
		Print("\033[K") // erase any leftover chars to the right
	}

	// ── Status bar (last terminal row) ────────────────────────
	toolLabel := map[string]string{
		"rect": "Rect", "circle": "Circle", "line": "Line", "free": "Free",
	}[a.tool]
	fillLabel := "Outline"; if a.fill { fillLabel = "Filled" }

	var statusText string
	if a.drawing && a.tool != "free" {
		statusText = FgBrightYellow + " DRAWING " + Reset +
			"end point → " + FgBrightGreen + "Space" + Reset + " draw  " +
			FgBrightYellow + "Esc" + Reset + " cancel"
	} else if a.status != "" {
		col := FgBrightGreen
		if a.statusErr { col = FgBrightRed }
		statusText = col + " " + a.status + Reset
		a.status = ""
	} else {
		statusText = Dim + " " + toolLabel + " · " + fillLabel + " · " +
			ColorANSI(a.color) + string(a.color) + Reset +
			Dim + "  |  " + Reset +
			ky("O") + " options  " +
			ky("Space") + " place  " +
			ky("U") + " undo  " +
			ky("X") + " clear  " +
			ky("H") + " help  " +
			ky("Q") + " quit" +
			Dim + fmt.Sprintf("  |  %d,%d", a.cx, a.cy) + Reset
	}

	PrintAt(1, statusRow, BgBlack+statusText+"\033[K")
}

// ── Key dispatcher ────────────────────────────────────────────

func (a *App) key(k int) bool {
	switch a.scr {
	case scrProjects: return a.keyProjects(k)
	case scrCanvas:   return a.keyCanvas(k)
	case scrNew:      return a.keyNew(k)
	case scrMenu:     return a.keyMenu(k)
	case scrHelp:     a.scr = scrCanvas
	}
	return true
}

func (a *App) keyProjects(k int) bool {
	n := len(a.projects)
	switch k {
	case KeyUp:    if a.projCursor > 0   { a.projCursor-- }
	case KeyDown:  if a.projCursor < n-1 { a.projCursor++ }
	case KeyEnter: if n > 0 { a.open(a.projects[a.projCursor]) }
	case 'n', 'N': a.inputBuf = ""; a.scr = scrNew
	case 'd', 'D':
		if n > 0 {
			db.DB.Delete(&models.Project{}, a.projects[a.projCursor].ID)
			a.loadProjects()
		}
	case 'q', 'Q': return false
	}
	return true
}

func (a *App) keyNew(k int) bool {
	switch k {
	case KeyEsc: a.scr = scrProjects
	case KeyEnter:
		if name := strings.TrimSpace(a.inputBuf); name != "" {
			p := models.Project{Name: name, Width: a.cw, Height: a.ch}
			db.DB.Create(&p)
			a.project = &p
			a.shapes = nil
			a.cx, a.cy = a.cw/2, a.ch/2
			a.scr = scrCanvas
			ClearScreen(); Flush()
		}
	case KeyBack:
		if len(a.inputBuf) > 0 {
			r := []rune(a.inputBuf)
			a.inputBuf = string(r[:len(r)-1])
		}
	default:
		if k >= 32 && k < 127 && len(a.inputBuf) < 40 {
			a.inputBuf += string(rune(k))
		}
	}
	return true
}

func (a *App) keyMenu(k int) bool {
	switch k {
	case 'r', 'R': a.tool = "rect"
	case 'c', 'C': a.tool = "circle"
	case 'l', 'L': a.tool = "line"
	case 'f', 'F': a.tool = "free"
	case 't', 'T': a.fill = !a.fill
	case KeyLeft:  a.color = shiftColor(a.color, -1)
	case KeyRight: a.color = shiftColor(a.color, +1)
	case KeyEnter, KeyEsc, KeySpace:
		ClearScreen(); Flush(); a.scr = scrCanvas
	}
	return true
}

func (a *App) keyCanvas(k int) bool {
	if a.project == nil { return true }
	W, H := a.cw, a.ch

	switch k {
	// ── movement ─────────────────────────────────────────────
	case KeyUp:    if a.cy > 0   { a.cy-- }
	case KeyDown:  if a.cy < H-1 { a.cy++ }
	case KeyLeft:  if a.cx > 0   { a.cx-- }
	case KeyRight: if a.cx < W-1 { a.cx++ }

	// ── freehand: draw while moving ──────────────────────────
	// (handled in drawCanvas — just move)

	// ── place / draw ─────────────────────────────────────────
	case KeySpace:
		if a.tool == "free" {
			// freehand: each space press commits one dot
			s := models.Shape{
				ProjectID: a.project.ID, Type: models.ShapeFree,
				X1: a.cx, Y1: a.cy, Color: a.color,
			}
			db.DB.Create(&s)
			a.shapes = append(a.shapes, s)
		} else if !a.drawing {
			a.drawing = true
			a.sx, a.sy = a.cx, a.cy
			a.status = fmt.Sprintf("Start (%d,%d) — move then Space", a.sx, a.sy)
		} else {
			a.commit()
			a.drawing = false
		}

	case KeyEnter:
		if !a.drawing {
			a.drawing = true
			a.sx, a.sy = a.cx, a.cy
		} else {
			a.commit()
			a.drawing = false
		}

	case KeyEsc:
		if a.drawing { a.drawing = false; a.status = "Cancelled"
		} else { a.scr = scrProjects; a.loadProjects() }

	// ── actions ──────────────────────────────────────────────
	case 'o', 'O': a.scr = scrMenu; a.drawing = false
	case 'u', 'U': a.undo()
	case 'x', 'X': a.clearAll()
	case 'h', 'H': a.scr = scrHelp
	case 'q', 'Q': a.drawing = false; a.scr = scrProjects; a.loadProjects()
	}

	// clamp cursor
	if a.cx < 0 { a.cx = 0 }; if a.cx >= W { a.cx = W - 1 }
	if a.cy < 0 { a.cy = 0 }; if a.cy >= H { a.cy = H - 1 }

	return true
}

// ── Data ops ──────────────────────────────────────────────────

func (a *App) loadProjects() {
	db.DB.Preload("Shapes").Order("created_at desc").Find(&a.projects)
	if a.projCursor >= len(a.projects) { a.projCursor = 0 }
}

func (a *App) open(p models.Project) {
	db.DB.First(&p, p.ID)
	a.project = &p
	a.cw, a.ch = p.Width, p.Height
	db.DB.Where("project_id = ? AND deleted_at IS NULL", p.ID).
		Order("created_at asc").Find(&a.shapes)
	a.cx, a.cy = a.cw/2, a.ch/2
	a.drawing = false
	a.scr = scrCanvas
	a.status = fmt.Sprintf("Opened %q", p.Name)
	ClearScreen(); Flush()
}

func (a *App) commit() {
	if a.project == nil { return }
	s := models.Shape{
		ProjectID: a.project.ID,
		Type:      models.ShapeType(a.tool),
		X1: a.sx, Y1: a.sy,
		Filled: a.fill,
		Color:  a.color,
	}
	switch a.tool {
	case "rect", "line": s.X2, s.Y2 = a.cx, a.cy
	case "circle":
		s.Radius = idist(a.sx, a.sy, a.cx, a.cy)
		s.X2, s.Y2 = a.cx, a.cy
	}
	db.DB.Create(&s)
	a.shapes = append(a.shapes, s)
	fill := "outline"; if s.Filled { fill = "filled" }
	a.status = fmt.Sprintf("Drew %s %s (%d,%d)→(%d,%d)", a.tool, fill, a.sx, a.sy, a.cx, a.cy)
}

func (a *App) undo() {
	if len(a.shapes) == 0 { a.status = "Nothing to undo"; return }
	db.DB.Delete(&models.Shape{}, a.shapes[len(a.shapes)-1].ID)
	a.shapes = a.shapes[:len(a.shapes)-1]
	a.status = "Undone"
}

func (a *App) clearAll() {
	if a.project == nil { return }
	db.DB.Where("project_id = ?", a.project.ID).Delete(&models.Shape{})
	a.shapes = nil; a.drawing = false
	a.status = "Cleared"
}

// ── Helpers ───────────────────────────────────────────────────

func ky(s string) string { return FgBrightYellow + s + Reset }

func idist(x1, y1, x2, y2 int) int {
	dx := float64(x2-x1) * 0.5
	dy := float64(y2 - y1)
	r := int(isqrt(dx*dx + dy*dy))
	if r < 1 { r = 1 }
	return r
}

func isqrt(x float64) float64 {
	if x <= 0 { return 0 }
	z := x
	for i := 0; i < 60; i++ { z = (z + x/z) / 2 }
	return z
}

func shiftColor(c models.ColorName, d int) models.ColorName {
	all := models.AllColors
	for i, v := range all {
		if v == c { return all[(i+d+len(all))%len(all)] }
	}
	return all[0]
}