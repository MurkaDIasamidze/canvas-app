package tui

import (
	"canvas-tui/canvas"
	"canvas-tui/db"
	"canvas-tui/models"
	"fmt"
	"strings"
)

type screen int
const (
	scrProjects screen = iota
	scrCanvas
	scrNewProject
	scrShapeMenu
	scrHelp
)

type App struct {
	projects   []models.Project
	project    *models.Project
	shapes     []models.Shape
	scr        screen
	cx, cy     int
	drawing    bool
	sx, sy     int       // drawing start point
	shape      models.ShapeType
	fill       bool
	color      models.ColorName
	projCursor int
	inputBuf   string
	status     string
}

func Run() {
	InitTerminal(); HideCursor(); ClearScreen(); Flush()
	a := &App{shape: models.ShapeRect, color: models.ColorGreen}
	db.Connect()
	a.loadProjects()
	defer func() { ClearScreen(); ShowCursor(); Flush(); restoreTerminal() }()
	initRawMode()
	for a.draw(); ; a.draw() {
		if !a.key(ReadKey()) { return }
	}
}

// ── render ────────────────────────────────────────────────────────────────────

func (a *App) draw() {
	switch a.scr {
	case scrProjects:   a.drawProjects()
	case scrCanvas:     a.drawCanvas()
	case scrNewProject: a.drawNewProject()
	case scrShapeMenu:  a.drawMenu()
	case scrHelp:       a.drawHelp()
	}
	Flush()
}

func ln(s string) { Print(s + "\033[K\n") }
func hdr(s string) { ln(BgBlue + FgBrightWhite + Bold + s + Reset) }
func sep(n int)    { ln(Dim + Repeat("─", n) + Reset) }

func (a *App) drawProjects() {
	MoveTo(1, 1)
	hdr("  ASCII CANVAS STUDIO  ")
	sep(56)
	ln("")
	ln(Bold + "  Projects" + Reset)
	sep(36)
	if len(a.projects) == 0 {
		ln(Dim + "  No projects — press " + Reset + FgBrightYellow + "N" + Reset + Dim + " to create one." + Reset)
	}
	for i, p := range a.projects {
		if i == a.projCursor {
			ln(BgBlue + FgBrightWhite + fmt.Sprintf("  ▶  %-24s  %dx%d", p.Name, p.Width, p.Height) + Reset)
		} else {
			ln(fmt.Sprintf("     %-24s  "+Dim+"%dx%d"+Reset, p.Name, p.Width, p.Height))
		}
	}
	ln(""); sep(56)
	ln("  " + key("↑↓") + " navigate  " + key("Enter") + " open  " +
		key("N") + " new  " + key("D") + " delete  " + key("Q") + " quit")
	EraseDown()
}

func (a *App) drawNewProject() {
	MoveTo(1, 1)
	hdr("  NEW PROJECT  ")
	ln("")
	ln("  Name: " + FgBrightGreen + a.inputBuf + FgBrightYellow + "█" + Reset)
	ln("")
	ln(Dim + "  Enter to create  |  Esc to cancel" + Reset)
	EraseDown()
}

func (a *App) drawMenu() {
	MoveTo(1, 1)
	hdr("  SHAPE OPTIONS  ")
	sep(50); ln("")

	ln(Bold + "  Shape  " + Reset + Dim + "(press number)" + Reset)
	for i, s := range []struct{ t models.ShapeType; icon, label string }{
		{models.ShapeRect,   "▭", "Rectangle"},
		{models.ShapeCircle, "◯", "Circle"},
		{models.ShapeLine,   "─", "Line"},
	} {
		num := fmt.Sprintf("%d", i+1)
		if a.shape == s.t {
			ln(FgBrightGreen + "  ▶ [" + num + "] " + s.icon + "  " + s.label + Reset)
		} else {
			ln(Dim + "    [" + num + "] " + s.icon + "  " + s.label + Reset)
		}
	}
	ln("")

	ln(Bold + "  Fill  " + Reset + Dim + "(press F)" + Reset)
	outline, filled := Dim+"    [□] Outline"+Reset, Dim+"    [■] Filled"+Reset
	if !a.fill { outline = FgBrightCyan + "  ▶ [□] Outline" + Reset
	} else      { filled  = FgBrightCyan + "  ▶ [■] Filled"  + Reset }
	ln(outline); ln(filled); ln("")

	ln(Bold + "  Color  " + Reset + Dim + "(← →)" + Reset)
	row := "  "
	for _, c := range models.AllColors {
		ch := string([]rune(string(c))[:1])
		if c == a.color { row += ColorANSI(c) + Bold + "[" + ch + "]" + Reset + " "
		} else           { row += Dim + ColorANSI(c) + " " + ch + " " + Reset + " " }
	}
	ln(row)
	ln("  " + ColorANSI(a.color) + Bold + string(a.color) + Reset)
	ln(""); sep(50)
	ln("  " + key("1/2/3") + " shape  " + key("F") + " fill  " +
		key("←→") + " color  " + key("Enter") + " back")
	EraseDown()
}

func (a *App) drawHelp() {
	MoveTo(1, 1)
	hdr("  HELP  ")
	for _, l := range []string{
		"", Bold + "  CONTROLS" + Reset,
		"  " + key("↑↓←→") + "    Move cursor",
		"  " + key("Space") + "    1st press = start point  |  2nd press = draw",
		"  " + key("S") + "       Shape / fill / color menu",
		"  " + key("U") + "       Undo last shape",
		"  " + key("X") + "       Clear canvas",
		"  " + key("Q") + "       Back to project list",
		"  " + key("Esc") + "     Cancel drawing",
		"", Bold + "  SHAPES" + Reset,
		"  Line:      auto chars  ─ │ ╱ ╲",
		"  Rectangle: outline  ┌─┐│└┘  |  filled  █",
		"  Circle:    outline  ─│╱╲    |  filled  █",
		"", "  Press any key to return...",
	} { ln(l) }
	EraseDown()
}

func (a *App) drawCanvas() {
	if a.project == nil { a.scr = scrProjects; return }
	W, H := a.project.Width, a.project.Height
	const numW, borderRow, dataRow = 4, 3, 4
	leftCol := numW + 1

	// Header
	shapeLabel := map[models.ShapeType]string{
		models.ShapeRect: "Rect", models.ShapeCircle: "Circle", models.ShapeLine: "Line",
	}[a.shape]
	fillLabel := "Outline"; if a.fill { fillLabel = "Filled" }
	PrintAt(1, 1, BgBlue+FgBrightWhite+Bold+" "+a.project.Name+" "+Reset+
		Dim+"  "+Reset+shapeLabel+Dim+"  "+Reset+fillLabel+
		Dim+"  Color:"+Reset+" "+ColorANSI(a.color)+string(a.color)+Reset+
		Dim+"  "+Reset+fmt.Sprintf("%d,%d", a.cx, a.cy)+
		Dim+"  Shapes:"+Reset+fmt.Sprintf(" %d", len(a.shapes))+"\033[K")

	// Hint bar
	if a.drawing {
		PrintAt(1, 2, FgBrightYellow+"  DRAWING "+Reset+
			"move cursor → "+FgBrightGreen+"Space"+Reset+" to draw  |  "+
			FgBrightYellow+"Esc"+Reset+" cancel\033[K")
	} else {
		PrintAt(1, 2, Dim+"  "+Reset+
			key("↑↓←→")+" move  "+key("Space")+" place  "+
			key("S")+" options  "+key("U")+" undo  "+
			key("X")+" clear  "+key("H")+" help  "+key("Q")+" quit\033[K")
	}

	// Build grid
	g := canvas.RenderAll(W, H, a.shapes)
	if a.drawing {
		p := canvas.NewGrid(W, H)
		switch a.shape {
		case models.ShapeRect:   canvas.DrawRect(p, a.sx, a.sy, a.cx, a.cy, a.color, a.fill)
		case models.ShapeCircle: canvas.DrawCircle(p, a.sx, a.sy, dist(a.sx, a.sy, a.cx, a.cy), a.color, a.fill)
		case models.ShapeLine:   canvas.DrawLine(p, a.sx, a.sy, a.cx, a.cy, a.color)
		}
		for y := 0; y < H; y++ {
			for x := 0; x < W; x++ {
				if c := p.Get(x, y); c.Ch != ' ' { g.Set(x, y, c.Ch, c.Color) }
			}
		}
	}

	// Border top
	PrintAt(1, borderRow, Dim+Repeat(" ", numW)+Reset+FgCyan+"┌"+Repeat("─", W)+"┐"+Reset+"\033[K")

	// Rows
	for y := 0; y < H; y++ {
		row := dataRow + y
		PrintfAt(1, row, Dim+"%3d "+Reset, y)
		PrintAt(leftCol, row, FgCyan+"│"+Reset)
		for x := 0; x < W; x++ {
			col  := leftCol + 1 + x
			cell := g.Get(x, y)
			isCur   := x == a.cx && y == a.cy
			isStart := a.drawing && x == a.sx && y == a.sy
			switch {
			case isCur && isStart: PrintAt(col, row, BgYellow+FgBlack+"@"+Reset)
			case isCur:            PrintAt(col, row, BgGreen+FgBlack+string(cell.Ch)+Reset)
			case isStart:          PrintAt(col, row, BgYellow+FgBlack+"·"+Reset)
			case cell.Ch != ' ':   PrintAt(col, row, ColorANSI(cell.Color)+Bold+string(cell.Ch)+Reset)
			default:               PrintAt(col, row, " ")
			}
		}
		PrintAt(leftCol+W+1, row, FgCyan+"│"+Reset+"\033[K")
	}

	// Border bottom + ruler
	rulerRow := borderRow + H + 1
	PrintAt(1, rulerRow-1, Dim+Repeat(" ", numW)+Reset+FgCyan+"└"+Repeat("─", W)+"┘"+Reset+"\033[K")
	ruler := Repeat(" ", numW+1)
	for i := 0; i < W; i += 10 { ruler += Dim + fmt.Sprintf("%-10d", i) + Reset }
	PrintAt(1, rulerRow, ruler+"\033[K")

	// Status
	statusRow := rulerRow + 1
	if a.status != "" {
		PrintAt(1, statusRow, "  "+FgBrightGreen+a.status+Reset+"\033[K")
		a.status = ""
	} else {
		PrintAt(1, statusRow, "\033[K")
	}

	// Sidebar: shape list + colors
	sideX := leftCol + W + 3
	rows := []string{Bold + "  Shapes" + Reset, Dim + "  " + Repeat("─", 20) + Reset}
	if len(a.shapes) == 0 {
		rows = append(rows, Dim+"  (empty)"+Reset)
	} else {
		icons := map[models.ShapeType]string{models.ShapeRect: "▭", models.ShapeCircle: "◯", models.ShapeLine: "─"}
		start := 0; if len(a.shapes) > H-4 { start = len(a.shapes) - (H - 4) }
		for _, s := range a.shapes[start:] {
			fill := "□"; if s.Filled { fill = "■" }
			rows = append(rows, fmt.Sprintf("  %s %s%s%s %s (%d,%d)",
				icons[s.Type], ColorANSI(s.Color), string([]rune(string(s.Color))[:2]), Reset, fill, s.X1, s.Y1))
		}
	}
	rows = append(rows, "", Bold+"  Colors"+Reset)
	for _, c := range models.AllColors {
		bul := "  "; if c == a.color { bul = FgBrightWhite + " ▶" + Reset }
		rows = append(rows, bul+" "+ColorANSI(c)+Bold+"██"+Reset+" "+string(c))
	}
	for i, r := range rows {
		if sr := borderRow + i; sr <= borderRow+H { PrintAt(sideX, sr, r+"\033[K") }
	}
}

// ── keys ──────────────────────────────────────────────────────────────────────

func (a *App) key(k int) bool {
	switch a.scr {
	case scrProjects:   return a.keyProjects(k)
	case scrCanvas:     return a.keyCanvas(k)
	case scrNewProject: return a.keyNewProject(k)
	case scrShapeMenu:  return a.keyMenu(k)
	case scrHelp:       a.scr = scrCanvas
	}
	return true
}

func (a *App) keyProjects(k int) bool {
	n := len(a.projects)
	switch k {
	case KeyArrowUp:   if a.projCursor > 0   { a.projCursor-- }
	case KeyArrowDown: if a.projCursor < n-1 { a.projCursor++ }
	case KeyEnter:     if n > 0 { a.open(a.projects[a.projCursor]) }
	case 'n', 'N':    a.inputBuf = ""; a.scr = scrNewProject
	case 'd', 'D':    if n > 0 { db.DB.Delete(&models.Project{}, a.projects[a.projCursor].ID); a.loadProjects() }
	case 'q', 'Q':    return false
	}
	return true
}

func (a *App) keyNewProject(k int) bool {
	switch k {
	case KeyEscape: a.scr = scrProjects
	case KeyEnter:
		if name := strings.TrimSpace(a.inputBuf); name != "" {
			p := models.Project{Name: name, Width: 70, Height: 28}
			db.DB.Create(&p); a.project = &p; a.shapes = nil
			a.cx, a.cy = p.Width/2, p.Height/2
			a.scr = scrCanvas; a.status = "Press H for help"
			ClearScreen(); Flush()
		}
	case KeyBackspace:
		if len(a.inputBuf) > 0 { r := []rune(a.inputBuf); a.inputBuf = string(r[:len(r)-1]) }
	default:
		if k >= 32 && k < 127 && len(a.inputBuf) < 28 { a.inputBuf += string(rune(k)) }
	}
	return true
}

func (a *App) keyMenu(k int) bool {
	switch k {
	case '1': a.shape = models.ShapeRect
	case '2': a.shape = models.ShapeCircle
	case '3': a.shape = models.ShapeLine
	case 'f', 'F': a.fill = !a.fill
	case KeyArrowLeft:  a.color = shiftColor(a.color, -1)
	case KeyArrowRight: a.color = shiftColor(a.color, +1)
	case KeyEnter, KeyEscape, KeySpace:
		ClearScreen(); Flush(); a.scr = scrCanvas
	}
	return true
}

func (a *App) keyCanvas(k int) bool {
	if a.project == nil { return true }
	W, H := a.project.Width, a.project.Height
	switch k {
	case KeyArrowUp:    if a.cy > 0   { a.cy-- }
	case KeyArrowDown:  if a.cy < H-1 { a.cy++ }
	case KeyArrowLeft:  if a.cx > 0   { a.cx-- }
	case KeyArrowRight: if a.cx < W-1 { a.cx++ }
	case KeySpace, KeyEnter:
		if !a.drawing {
			a.drawing = true; a.sx, a.sy = a.cx, a.cy
			a.status = fmt.Sprintf("Start (%d,%d) — move cursor, Space to draw", a.sx, a.sy)
		} else {
			a.commit(); a.drawing = false
		}
	case KeyEscape:
		if a.drawing { a.drawing = false; a.status = "Cancelled"
		} else { a.scr = scrProjects; a.loadProjects(); a.projCursor = 0 }
	case 's', 'S': a.scr = scrShapeMenu; a.drawing = false
	case 'u', 'U': a.undo()
	case 'x', 'X': a.clear()
	case 'h', 'H': a.scr = scrHelp
	case 'q', 'Q': a.drawing = false; a.scr = scrProjects; a.loadProjects(); a.projCursor = 0
	}
	return true
}

// ── data ──────────────────────────────────────────────────────────────────────

func (a *App) loadProjects() {
	db.DB.Preload("Shapes").Order("created_at desc").Find(&a.projects)
	if a.projCursor >= len(a.projects) { a.projCursor = 0 }
}

func (a *App) open(p models.Project) {
	db.DB.First(&p, p.ID)
	a.project = &p
	db.DB.Where("project_id = ? AND deleted_at IS NULL", p.ID).Order("created_at asc").Find(&a.shapes)
	a.cx, a.cy = p.Width/2, p.Height/2
	a.drawing = false; a.scr = scrCanvas
	a.status = fmt.Sprintf("Opened %q — H for help", p.Name)
	ClearScreen(); Flush()
}

func (a *App) commit() {
	if a.project == nil { return }
	s := models.Shape{ProjectID: a.project.ID, Type: a.shape,
		X1: a.sx, Y1: a.sy, Filled: a.fill, Color: a.color}
	switch a.shape {
	case models.ShapeRect, models.ShapeLine: s.X2, s.Y2 = a.cx, a.cy
	case models.ShapeCircle: s.Radius = dist(a.sx, a.sy, a.cx, a.cy); s.X2, s.Y2 = a.cx, a.cy
	}
	db.DB.Create(&s); a.shapes = append(a.shapes, s)
	fill := "outline"; if s.Filled { fill = "filled" }
	a.status = fmt.Sprintf("Drew %s %s at (%d,%d)", s.Type, fill, a.sx, a.sy)
}

func (a *App) undo() {
	if len(a.shapes) == 0 { a.status = "Nothing to undo"; return }
	db.DB.Delete(&models.Shape{}, a.shapes[len(a.shapes)-1].ID)
	a.shapes = a.shapes[:len(a.shapes)-1]; a.status = "Undone"
}

func (a *App) clear() {
	if a.project == nil { return }
	db.DB.Where("project_id = ?", a.project.ID).Delete(&models.Shape{})
	a.shapes = nil; a.drawing = false; a.status = "Canvas cleared"
}

// ── helpers ───────────────────────────────────────────────────────────────────

func key(k string) string { return FgBrightYellow + k + Reset }

func dist(x1, y1, x2, y2 int) int {
	dx := float64(x2-x1) * 0.5
	dy := float64(y2 - y1)
	r := int(math_sqrt(dx*dx + dy*dy))
	if r < 1 { r = 1 }
	return r
}

func math_sqrt(x float64) float64 {
	if x <= 0 { return 0 }
	z := x
	for i := 0; i < 50; i++ { z = (z + x/z) / 2 }
	return z
}

func shiftColor(c models.ColorName, d int) models.ColorName {
	all := models.AllColors
	for i, v := range all {
		if v == c { return all[(i+d+len(all))%len(all)] }
	}
	return all[0]
}