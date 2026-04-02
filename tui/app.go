package tui

import (
	"canvas-tui/canvas"
	"canvas-tui/db"
	"canvas-tui/models"
	"fmt"
	"strings"
)

// ── Screen enum ─────────────────────────────────────────────────────────────

type Screen int

const (
	ScrProjects Screen = iota
	ScrCanvas
	ScrNewProject
	ScrShapeMenu
	ScrHelp
)

// ── App state ────────────────────────────────────────────────────────────────

type App struct {
	// data
	projects []models.Project
	project  *models.Project
	shapes   []models.Shape

	// current screen
	screen Screen

	// canvas cursor position
	cx, cy int

	// drawing state
	drawing        bool
	startX, startY int

	// shape options
	shapeType models.ShapeType
	fillMode  bool
	drawChar  rune
	drawColor models.ColorName

	// shape menu cursor
	menuCursor int // 0=type, 1=fill, 2=char, 3=color

	// project list cursor
	projCursor int

	// text input buffer (new project name)
	inputBuf string

	// one-line status message shown on canvas
	status string

	// canvas screen offset — where on screen the canvas top-left is drawn
	// set once and never changes: fixes the no-scroll requirement
	canvasOffX int // column of the left │ border
	canvasOffY int // row of the ┌ border

	// layout constants (computed once)
	sideW  int // width of right sidebar
	headerH int // rows used above the canvas border
	footerH int // rows used below the canvas border
}

func newApp() *App {
	return &App{
		screen:    ScrProjects,
		shapeType: models.ShapeRect,
		fillMode:  false,
		drawChar:  '*',
		drawColor: models.ColorGreen,
		sideW:     28,
		headerH:   3, // row1=header bar, row2=keybind bar, row3=border top
		footerH:   3, // border bottom + ruler + status
	}
}

// ── Entry point ──────────────────────────────────────────────────────────────

func Run() {
	InitTerminal()
	HideCursor()
	ClearScreen()
	Flush()

	app := newApp()
	db.Connect()
	app.loadProjects()

	defer func() {
		// restore terminal cleanly
		MoveTo(1, 1)
		ClearScreen()
		ShowCursor()
		EraseDown()
		Flush()
		restoreTerminal()
	}()

	initRawMode()

	// Draw once, then enter event loop
	app.draw()
	for {
		key := ReadKey()
		if !app.handleKey(key) {
			return
		}
		app.draw()
	}
}

// ── Top-level draw dispatcher ─────────────────────────────────────────────────

func (a *App) draw() {
	switch a.screen {
	case ScrProjects:
		a.drawProjectList()
	case ScrCanvas:
		a.drawCanvas()
	case ScrNewProject:
		a.drawNewProject()
	case ScrShapeMenu:
		a.drawShapeMenu()
	case ScrHelp:
		a.drawHelp()
	}
	Flush()
}

// ── Project list screen ───────────────────────────────────────────────────────

func (a *App) drawProjectList() {
	// Paint from top-left — no clearing needed, we overwrite every line
	MoveTo(1, 1)

	line := func(s string) { Print(s + "\033[K\n") } // \033[K clears to EOL

	line(BgBlue + FgBrightWhite + Bold + "  ASCII CANVAS STUDIO  " + Reset +
		Dim + "  Terminal Drawing App  |  Q=quit" + Reset)
	line(Dim + Repeat("─", 60) + Reset)
	line("")
	line(Bold + FgBrightWhite + "  Projects" + Reset)
	line(Dim + "  " + Repeat("─", 40) + Reset)

	if len(a.projects) == 0 {
		line(Dim + "  No projects yet." + Reset)
		line(Dim + "  Press " + Reset + FgBrightYellow + "N" + Reset + Dim + " to create one." + Reset)
	} else {
		for i, p := range a.projects {
			if i == a.projCursor {
				line(BgBlue + FgBrightWhite + fmt.Sprintf("  ▶  %-22s  %dx%d", p.Name, p.Width, p.Height) + Reset)
			} else {
				line(fmt.Sprintf("     %-22s  "+Dim+"%dx%d"+Reset, p.Name, p.Width, p.Height))
			}
		}
	}

	line("")
	line(Dim + Repeat("─", 60) + Reset)
	line("  " + FgBrightYellow + "↑↓" + Reset + " navigate   " +
		FgBrightYellow + "Enter" + Reset + " open   " +
		FgBrightYellow + "N" + Reset + " new   " +
		FgBrightYellow + "D" + Reset + " delete   " +
		FgBrightYellow + "Q" + Reset + " quit")
	// Erase anything below
	EraseDown()
}

// ── New project screen ────────────────────────────────────────────────────────

func (a *App) drawNewProject() {
	MoveTo(1, 1)
	line := func(s string) { Print(s + "\033[K\n") }
	line(BgBlue + FgBrightWhite + Bold + "  NEW PROJECT  " + Reset)
	line("")
	line("  Name: " + FgBrightGreen + a.inputBuf + FgBrightYellow + "█" + Reset)
	line("")
	line(Dim + "  Type name → Enter to create   Esc to cancel" + Reset)
	EraseDown()
}

// ── Shape menu screen ─────────────────────────────────────────────────────────

func (a *App) drawShapeMenu() {
	MoveTo(1, 1)
	line := func(s string) { Print(s + "\033[K\n") }

	line(BgBlue + FgBrightWhite + Bold + "  SHAPE OPTIONS  " + Reset +
		Dim + "  Enter/Esc = back to canvas" + Reset)
	line(Dim + Repeat("─", 50) + Reset)
	line("")

	// Shape type
	shapes := []struct {
		key   string
		stype models.ShapeType
		label string
		icon  string
	}{
		{"1", models.ShapeRect, "Rectangle", "▭"},
		{"2", models.ShapeCircle, "Circle", "◯"},
		{"3", models.ShapeLine, "Line", "╱"},
	}
	line(Bold + "  Shape Type" + Reset + Dim + "  (press number key)" + Reset)
	for _, s := range shapes {
		sel := a.shapeType == s.stype
		if sel {
			line(FgBrightGreen + "  ▶ [" + s.key + "] " + s.icon + " " + s.label + Reset)
		} else {
			line(Dim + "    [" + s.key + "] " + s.icon + " " + s.label + Reset)
		}
	}
	line("")

	// Fill mode
	line(Bold + "  Fill Mode" + Reset + Dim + "  (press F to toggle)" + Reset)
	if !a.fillMode {
		line(FgBrightCyan + "  ▶ [□] Outline" + Reset)
		line(Dim + "    [■] Filled" + Reset)
	} else {
		line(Dim + "    [□] Outline" + Reset)
		line(FgBrightCyan + "  ▶ [■] Filled" + Reset)
	}
	line("")

	// Draw character
	line(Bold + "  Draw Char" + Reset + Dim + "  (press C, type new char)" + Reset)
	presets := []rune{'*', '#', '@', '+', 'X', 'O', '.', '=', '~', '%', '&', '$'}
	charRow := "  "
	for _, ch := range presets {
		if rune(a.drawChar) == ch {
			charRow += FgBrightYellow + "[" + string(ch) + "]" + Reset + " "
		} else {
			charRow += Dim + " " + string(ch) + " " + Reset + " "
		}
	}
	line(charRow)
	line("  Current: " + FgBrightYellow + Bold + string(a.drawChar) + Reset +
		"   Press " + FgBrightYellow + "C" + Reset + " then type any character")
	line("")

	// Color picker
	line(Bold + "  Color" + Reset + Dim + "  (press ← → to change)" + Reset)
	colorRow := "  "
	for _, c := range models.AllColors {
		if c == a.drawColor {
			colorRow += ColorANSI(c) + Bold + "[" + string(c)[0:1] + "]" + Reset + " "
		} else {
			colorRow += Dim + ColorANSI(c) + " " + string(c)[0:1] + " " + Reset + " "
		}
	}
	line(colorRow)
	line("  Selected: " + ColorANSI(a.drawColor) + Bold + string(a.drawColor) + Reset)

	line("")
	line(Dim + Repeat("─", 50) + Reset)
	line("  " + FgBrightYellow + "1/2/3" + Reset + " shape   " +
		FgBrightYellow + "F" + Reset + " fill   " +
		FgBrightYellow + "C" + Reset + " char   " +
		FgBrightYellow + "←→" + Reset + " color   " +
		FgBrightYellow + "Enter" + Reset + " draw")
	EraseDown()
}

// ── Help screen ───────────────────────────────────────────────────────────────

func (a *App) drawHelp() {
	MoveTo(1, 1)
	line := func(s string) { Print(s + "\033[K\n") }
	line(BgBlue + FgBrightWhite + Bold + "  HELP  " + Reset)
	helps := []string{
		"",
		Bold + "  CONTROLS" + Reset,
		"  " + FgBrightYellow + "↑ ↓ ← →" + Reset + "     Move cursor on canvas",
		"  " + FgBrightYellow + "Space/Enter" + Reset + "  Place point (1st=start, 2nd=draw shape)",
		"  " + FgBrightYellow + "S" + Reset + "           Shape/fill/char/color menu",
		"  " + FgBrightYellow + "U" + Reset + "           Undo last shape",
		"  " + FgBrightYellow + "X" + Reset + "           Clear canvas",
		"  " + FgBrightYellow + "H" + Reset + "           This help screen",
		"  " + FgBrightYellow + "Q" + Reset + "           Back to project list",
		"  " + FgBrightYellow + "Esc" + Reset + "         Cancel drawing",
		"",
		Bold + "  DRAWING" + Reset,
		"  1. Press S → choose shape, fill, character, color",
		"  2. Move cursor to start position",
		"  3. Press Space → start point locked (yellow cursor)",
		"  4. Move cursor → live preview of shape follows",
		"  5. Press Space again → shape committed to canvas",
		"",
		Bold + "  CIRCLE" + Reset,
		"  Start = center, move cursor outward = radius",
		"",
		"  Press any key to return...",
	}
	for _, h := range helps {
		line(h)
	}
	EraseDown()
}

// ── Canvas screen — the main drawing area ─────────────────────────────────────
//
// Layout (fixed, never scrolls):
//
//   Row 1:   [header bar]
//   Row 2:   [keybind/status bar]
//   Row 3:   ┌────────── canvas ──────────┐   [sidebar top]
//   Row 4..N │  . . . cells . . .         │   [sidebar rows]
//   Row N+1: └────────────────────────────┘   [sidebar bottom]
//   Row N+2: [x-axis ruler]
//   Row N+3: [status message]
//
// The canvas border is always at the same screen rows — no scroll possible.

func (a *App) drawCanvas() {
	if a.project == nil {
		a.screen = ScrProjects
		return
	}

	W := a.project.Width
	H := a.project.Height

	// Fixed layout constants
	const (
		headerRow = 1
		keybindRow = 2
		borderTop = 3       // row of the ┌ line
		dataStart = 4       // first canvas data row
		numColW   = 4       // "  0 " row-number column width
	)
	canvasCol := numColW + 1  // column where ┌ is drawn (1-based)
	rulerRow  := borderTop + H + 1
	statusRow := borderTop + H + 2

	// ── Header bar (row 1) ────────────────────────────────────────────────
	shapeNames := map[models.ShapeType]string{
		models.ShapeRect:   "Rect",
		models.ShapeCircle: "Circle",
		models.ShapeLine:   "Line",
	}
	fillStr := "Outline"
	if a.fillMode { fillStr = "Filled" }

	PrintAt(1, headerRow,
		BgBlue+FgBrightWhite+Bold+" "+a.project.Name+" "+Reset+
			Dim+"  Shape:"+Reset+" "+shapeNames[a.shapeType]+
			Dim+"  Fill:"+Reset+" "+fillStr+
			Dim+"  Char:"+Reset+" "+FgBrightYellow+string(a.drawChar)+Reset+
			Dim+"  Color:"+Reset+" "+ColorANSI(a.drawColor)+string(a.drawColor)+Reset+
			Dim+"  Pos:"+Reset+fmt.Sprintf(" %d,%d", a.cx, a.cy)+
			Dim+"  Shapes:"+Reset+fmt.Sprintf(" %d", len(a.shapes))+
			"\033[K")

	// ── Keybind / drawing-status bar (row 2) ─────────────────────────────
	PrintAt(1, keybindRow, "\033[K") // clear line first
	if a.drawing {
		PrintAt(1, keybindRow,
			FgBrightYellow+"  DRAWING "+Reset+
				"Move cursor to end → press "+FgBrightGreen+"Space"+Reset+" to place  |  "+
				FgBrightYellow+"Esc"+Reset+" cancel"+
				"\033[K")
	} else {
		PrintAt(1, keybindRow,
			Dim+"  "+Reset+
				FgBrightYellow+"↑↓←→"+Reset+" move  "+
				FgBrightYellow+"Space"+Reset+" place  "+
				FgBrightYellow+"S"+Reset+" options  "+
				FgBrightYellow+"U"+Reset+" undo  "+
				FgBrightYellow+"X"+Reset+" clear  "+
				FgBrightYellow+"H"+Reset+" help  "+
				FgBrightYellow+"Q"+Reset+" quit"+
				"\033[K")
	}

	// ── Build grid (saved shapes + live preview) ─────────────────────────
	g := canvas.RenderAll(W, H, a.shapes)

	// Preview overlay while drawing
	if a.drawing {
		prev := canvas.NewGrid(W, H)
		ch := a.drawChar
		previewColor := a.drawColor
		switch a.shapeType {
		case models.ShapeRect:
			canvas.DrawRect(prev, a.startX, a.startY, a.cx, a.cy, ch, previewColor, a.fillMode)
		case models.ShapeCircle:
			r := dist(a.startX, a.startY, a.cx, a.cy)
			canvas.DrawCircle(prev, a.startX, a.startY, r, ch, previewColor, a.fillMode)
		case models.ShapeLine:
			canvas.DrawLine(prev, a.startX, a.startY, a.cx, a.cy, ch, previewColor)
		}
		for y := 0; y < H; y++ {
			for x := 0; x < W; x++ {
				if c := prev.Get(x, y); c.Ch != ' ' {
					g.Set(x, y, c.Ch, c.Color)
				}
			}
		}
	}

	// ── Canvas border top (row 3) ────────────────────────────────────────
	PrintAt(1, borderTop, Dim+Repeat(" ", numColW)+Reset+
		FgCyan+"┌"+Repeat("─", W)+"┐"+Reset+"\033[K")

	// ── Canvas rows (rows 4 .. 4+H-1) ────────────────────────────────────
	for y := 0; y < H; y++ {
		screenRow := dataStart + y

		// Row number
		PrintfAt(1, screenRow, Dim+"%3d "+Reset, y)

		// Left border
		PrintAt(canvasCol, screenRow, FgCyan+"│"+Reset)

		// Cells — write each one at its exact screen column
		for x := 0; x < W; x++ {
			col := canvasCol + 1 + x
			cell := g.Get(x, y)

			isCursor := x == a.cx && y == a.cy
			isStart  := a.drawing && x == a.startX && y == a.startY

			switch {
			case isCursor && isStart:
				// cursor is sitting on start point
				PrintAt(col, screenRow, BgYellow+FgBlack+"@"+Reset)
			case isCursor:
				PrintAt(col, screenRow, BgGreen+FgBlack+string(cell.Ch)+Reset)
			case isStart:
				PrintAt(col, screenRow, BgYellow+FgBlack+string(cell.Ch)+Reset)
			case cell.Ch != ' ':
				PrintAt(col, screenRow, ColorANSI(cell.Color)+Bold+string(cell.Ch)+Reset)
			default:
				PrintAt(col, screenRow, " ")
			}
		}

		// Right border
		PrintAt(canvasCol+W+1, screenRow, FgCyan+"│"+Reset+"\033[K")
	}

	// ── Canvas border bottom ──────────────────────────────────────────────
	PrintAt(1, rulerRow-1, Dim+Repeat(" ", numColW)+Reset+
		FgCyan+"└"+Repeat("─", W)+"┘"+Reset+"\033[K")

	// ── X-axis ruler ─────────────────────────────────────────────────────
	ruler := Repeat(" ", numColW+1) // align with canvas content
	for i := 0; i < W; i += 10 {
		marker := fmt.Sprintf("%-10d", i)
		ruler += Dim + marker + Reset
	}
	PrintAt(1, rulerRow, ruler+"\033[K")

	// ── Status message ────────────────────────────────────────────────────
	statusStr := ""
	if a.status != "" {
		statusStr = "  " + FgBrightGreen + a.status + Reset
		a.status = "" // consume it
	}
	PrintAt(1, statusRow, statusStr+"\033[K")

	// ── Sidebar (right of canvas) ─────────────────────────────────────────
	sideX := canvasCol + W + 3 // 2 cols gap after right border

	// Sidebar rows alongside canvas rows
	sideRows := []string{
		Bold + "  Shape list" + Reset,
		Dim + "  " + Repeat("─", 22) + Reset,
	}
	if len(a.shapes) == 0 {
		sideRows = append(sideRows, Dim+"  (empty)"+Reset)
	} else {
		// Show last N shapes that fit
		maxShow := H - 4
		start := 0
		if len(a.shapes) > maxShow {
			start = len(a.shapes) - maxShow
		}
		for i := start; i < len(a.shapes); i++ {
			s := a.shapes[i]
			icon := map[models.ShapeType]string{
				models.ShapeRect:   "▭",
				models.ShapeCircle: "◯",
				models.ShapeLine:   "╱",
			}[s.Type]
			fill := "□"
			if s.Filled { fill = "■" }
			row := fmt.Sprintf("  %s %s"+ColorANSI(s.Color)+"%s"+Reset+" %s (%d,%d)",
				icon, Dim, string(s.Color)[0:2], fill, s.X1, s.Y1)
			sideRows = append(sideRows, row)
		}
	}
	sideRows = append(sideRows, "")
	sideRows = append(sideRows, Bold+"  Colors"+Reset)
	for _, c := range models.AllColors {
		bullet := "  "
		if c == a.drawColor { bullet = FgBrightWhite + " ▶" + Reset }
		sideRows = append(sideRows, bullet+" "+ColorANSI(c)+Bold+"██"+Reset+" "+string(c))
	}

	for i, row := range sideRows {
		screenRow := borderTop + i
		if screenRow > borderTop+H { break }
		PrintAt(sideX, screenRow, row+"\033[K")
	}
}

// ── Key handlers ──────────────────────────────────────────────────────────────

func (a *App) handleKey(key int) bool {
	switch a.screen {
	case ScrProjects:   return a.keyProjects(key)
	case ScrCanvas:     return a.keyCanvas(key)
	case ScrNewProject: return a.keyNewProject(key)
	case ScrShapeMenu:  return a.keyShapeMenu(key)
	case ScrHelp:
		a.screen = ScrCanvas
	}
	return true
}

func (a *App) keyProjects(key int) bool {
	switch key {
	case KeyArrowUp:
		if a.projCursor > 0 { a.projCursor-- }
	case KeyArrowDown:
		if a.projCursor < len(a.projects)-1 { a.projCursor++ }
	case KeyEnter:
		if len(a.projects) > 0 {
			a.openProject(a.projects[a.projCursor])
		}
	case 'n', 'N':
		a.inputBuf = ""
		a.screen = ScrNewProject
	case 'd', 'D':
		if len(a.projects) > 0 {
			a.deleteProject(a.projects[a.projCursor].ID)
		}
	case 'q', 'Q':
		return false
	}
	return true
}

func (a *App) keyNewProject(key int) bool {
	switch key {
	case KeyEscape:
		a.screen = ScrProjects
	case KeyEnter:
		name := strings.TrimSpace(a.inputBuf)
		if name != "" {
			a.createProject(name)
			ClearScreen() // clear once before entering canvas
			Flush()
		}
	case KeyBackspace:
		if len(a.inputBuf) > 0 {
			runes := []rune(a.inputBuf)
			a.inputBuf = string(runes[:len(runes)-1])
		}
	default:
		if key >= 32 && key < 127 && len(a.inputBuf) < 28 {
			a.inputBuf += string(rune(key))
		}
	}
	return true
}

func (a *App) keyShapeMenu(key int) bool {
	switch key {
	case '1': a.shapeType = models.ShapeRect
	case '2': a.shapeType = models.ShapeCircle
	case '3': a.shapeType = models.ShapeLine
	case 'f', 'F': a.fillMode = !a.fillMode
	case 'c', 'C':
		// inline: show prompt, read one char
		PrintAt(3, 30, FgBrightYellow+"Type new draw character: "+Reset)
		Flush()
		ch := ReadKey()
		if ch >= 32 && ch < 127 {
			a.drawChar = rune(ch)
		}
	case KeyArrowLeft:
		a.drawColor = prevColor(a.drawColor)
	case KeyArrowRight:
		a.drawColor = nextColor(a.drawColor)
	case KeyEnter, KeyEscape, KeySpace:
		ClearScreen() // clear before returning to canvas
		Flush()
		a.screen = ScrCanvas
	}
	return true
}

func (a *App) keyCanvas(key int) bool {
	if a.project == nil { return true }
	W, H := a.project.Width, a.project.Height

	switch key {
	// ── cursor movement ──────────────────────────────────────────────────
	case KeyArrowUp:
		if a.cy > 0 { a.cy-- }
	case KeyArrowDown:
		if a.cy < H-1 { a.cy++ }
	case KeyArrowLeft:
		if a.cx > 0 { a.cx-- }
	case KeyArrowRight:
		if a.cx < W-1 { a.cx++ }

	// ── place / draw ─────────────────────────────────────────────────────
	case KeySpace, KeyEnter:
		if !a.drawing {
			a.drawing = true
			a.startX, a.startY = a.cx, a.cy
			a.status = fmt.Sprintf("Start at (%d,%d) — move cursor, Space to draw", a.startX, a.startY)
		} else {
			a.commitShape()
			a.drawing = false
		}

	// ── cancel / quit ────────────────────────────────────────────────────
	case KeyEscape:
		if a.drawing {
			a.drawing = false
			a.status = "Cancelled"
		} else {
			a.drawing = false
			a.screen = ScrProjects
			a.loadProjects()
			a.projCursor = 0
		}

	// ── actions ──────────────────────────────────────────────────────────
	case 's', 'S':
		a.screen = ScrShapeMenu
		a.drawing = false
	case 'u', 'U':
		a.undoLast()
	case 'x', 'X':
		a.clearCanvas()
	case 'h', 'H':
		a.screen = ScrHelp
	case 'q', 'Q':
		a.drawing = false
		a.screen = ScrProjects
		a.loadProjects()
		a.projCursor = 0
	}
	return true
}

// ── Data operations ───────────────────────────────────────────────────────────

func (a *App) loadProjects() {
	db.DB.Preload("Shapes").Order("created_at desc").Find(&a.projects)
	if a.projCursor >= len(a.projects) { a.projCursor = 0 }
}

func (a *App) openProject(p models.Project) {
	// Take a fresh copy from DB
	var proj models.Project
	db.DB.First(&proj, p.ID)
	a.project = &proj
	db.DB.Where("project_id = ? AND deleted_at IS NULL", proj.ID).
		Order("created_at asc").Find(&a.shapes)
	a.cx = proj.Width / 2
	a.cy = proj.Height / 2
	a.drawing = false
	a.screen = ScrCanvas
	a.status = fmt.Sprintf("Opened %q  —  H for help", proj.Name)
	ClearScreen()
	Flush()
}

func (a *App) createProject(name string) {
	p := models.Project{Name: name, Width: 70, Height: 28}
	db.DB.Create(&p)
	a.project = &p
	a.shapes = nil
	a.cx = p.Width / 2
	a.cy = p.Height / 2
	a.drawing = false
	a.screen = ScrCanvas
	a.status = "Project created — H for help"
}

func (a *App) deleteProject(id uint) {
	db.DB.Delete(&models.Project{}, id)
	a.loadProjects()
	a.status = "Deleted"
}

func (a *App) commitShape() {
	if a.project == nil { return }
	s := models.Shape{
		ProjectID: a.project.ID,
		Type:      a.shapeType,
		X1:        a.startX,
		Y1:        a.startY,
		Filled:    a.fillMode,
		Char:      string(a.drawChar),
		Color:     a.drawColor,
	}
	switch a.shapeType {
	case models.ShapeRect, models.ShapeLine:
		s.X2, s.Y2 = a.cx, a.cy
	case models.ShapeCircle:
		r := dist(a.startX, a.startY, a.cx, a.cy)
		s.Radius = r
		s.X2, s.Y2 = a.cx, a.cy
	}
	db.DB.Create(&s)
	a.shapes = append(a.shapes, s)
	fillLabel := "outline"
	if s.Filled { fillLabel = "filled" }
	a.status = fmt.Sprintf("Drew %s [%s] %s at (%d,%d)",
		s.Type, string(a.drawChar), fillLabel, a.startX, a.startY)
}

func (a *App) undoLast() {
	if len(a.shapes) == 0 {
		a.status = "Nothing to undo"
		return
	}
	last := a.shapes[len(a.shapes)-1]
	db.DB.Delete(&models.Shape{}, last.ID)
	a.shapes = a.shapes[:len(a.shapes)-1]
	a.status = "Undone"
}

func (a *App) clearCanvas() {
	if a.project == nil { return }
	db.DB.Where("project_id = ?", a.project.ID).Delete(&models.Shape{})
	a.shapes = nil
	a.drawing = false
	a.status = "Canvas cleared"
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func dist(x1, y1, x2, y2 int) int {
	dx := float64(x2-x1) * 0.5
	dy := float64(y2 - y1)
	r := int(sqrtf(dx*dx + dy*dy))
	if r < 1 { r = 1 }
	return r
}

func sqrtf(x float64) float64 {
	if x <= 0 { return 0 }
	z := x
	for i := 0; i < 50; i++ { z = (z + x/z) / 2 }
	return z
}

func nextColor(c models.ColorName) models.ColorName {
	all := models.AllColors
	for i, v := range all {
		if v == c { return all[(i+1)%len(all)] }
	}
	return all[0]
}

func prevColor(c models.ColorName) models.ColorName {
	all := models.AllColors
	for i, v := range all {
		if v == c { return all[(i-1+len(all))%len(all)] }
	}
	return all[0]
}