package tui

import (
	"canvas-tui/canvas"
	"canvas-tui/db"
	"canvas-tui/models"
	"fmt"
	"strings"
)

// App holds all application state
type App struct {
	projects      []models.Project
	project       *models.Project
	shapes        []models.Shape
	cursorX       int
	cursorY       int
	drawChar      rune
	fillMode      bool
	selectedShape models.ShapeType
	screen        Screen

	// For drawing — first point already placed
	drawing   bool
	startX    int
	startY    int

	// terminal size (detected or default)
	termW int
	termH int

	// message shown in status bar
	statusMsg string

	// input text buffer (for name entry etc.)
	inputBuf string
}

type Screen int

const (
	ScreenProjects Screen = iota
	ScreenCanvas
	ScreenNewProject
	ScreenShapeMenu
	ScreenHelp
)

func NewApp() *App {
	return &App{
		drawChar:      '*',
		fillMode:      false,
		selectedShape: models.ShapeRect,
		screen:        ScreenProjects,
		termW:         120,
		termH:         40,
		cursorX:       10,
		cursorY:       10,
	}
}

// Run is the main loop
func Run() {
	InitTerminal()
	HideCursor()

	app := NewApp()
	db.Connect()
	app.loadProjects()

	// Cleanup on exit
	defer func() {
		ShowCursor()
		Clear()
		Flush()
		restoreTerminal()
	}()

	initRawMode()

	for {
		app.render()
		key := ReadKey()
		if !app.handleKey(key) {
			return // quit
		}
	}
}

// ---------- rendering ----------

func (a *App) render() {
	switch a.screen {
	case ScreenProjects:
		a.renderProjectList()
	case ScreenCanvas:
		a.renderCanvas()
	case ScreenNewProject:
		a.renderNewProject()
	case ScreenShapeMenu:
		a.renderShapeMenu()
	case ScreenHelp:
		a.renderHelp()
	}
	Flush()
}

func (a *App) renderProjectList() {
	Clear()
	// Header
	PrintAt(1, 1, BgBlue+BrightWhite+Bold+" ASCII CANVAS STUDIO "+Reset+
		Dim+"  v1.0  |  Interactive Terminal Drawing"+Reset)
	PrintAt(1, 2, Dim+strings.Repeat("─", 60)+Reset)

	PrintAt(1, 4, Bold+White+"  Projects"+Reset)
	PrintAt(1, 5, Dim+"  ──────────────────────────────"+Reset)

	if len(a.projects) == 0 {
		PrintAt(3, 7, Dim+"  No projects yet. Press "+Reset+Yellow+"N"+Reset+Dim+" to create one."+Reset)
	} else {
		for i, p := range a.projects {
			prefix := "  "
			style := ""
			if i == a.cursorY {
				prefix = BgBlue + BrightWhite + " ▶"
				style = BgBlue + BrightWhite
			}
			PrintAtf(1, 7+i, "%s%-3d  %-20s  %dx%d  (%d shapes)%s",
				prefix+style, i+1, p.Name, p.Width, p.Height, len(p.Shapes), Reset)
		}
	}

	row := 9 + len(a.projects)
	PrintAt(1, row, Dim+strings.Repeat("─", 60)+Reset)
	row++
	PrintAt(1, row, "  "+Yellow+"N"+Reset+" New project   "+
		Yellow+"Enter"+Reset+" Open   "+
		Yellow+"D"+Reset+" Delete   "+
		Yellow+"Q"+Reset+" Quit")

	if a.statusMsg != "" {
		PrintAt(1, row+2, BrightGreen+"  "+a.statusMsg+Reset)
	}
}

func (a *App) renderNewProject() {
	Clear()
	PrintAt(1, 1, BgBlue+BrightWhite+Bold+" NEW PROJECT "+Reset)
	PrintAt(1, 3, "  Project name: "+BrightGreen+a.inputBuf+BrightYellow+"█"+Reset)
	PrintAt(1, 5, Dim+"  Type a name and press "+Reset+Yellow+"Enter"+Reset+Dim+" to create"+Reset)
	PrintAt(1, 6, Dim+"  Press "+Reset+Yellow+"Esc"+Reset+Dim+" to cancel"+Reset)
	Flush()
}

func (a *App) renderShapeMenu() {
	if a.project == nil {
		a.screen = ScreenProjects
		return
	}
	Clear()
	PrintAt(1, 1, BgBlue+BrightWhite+Bold+" SHAPE MENU "+Reset+
		Dim+"  project: "+Reset+BrightCyan+a.project.Name+Reset)
	PrintAt(1, 2, Dim+strings.Repeat("─", 50)+Reset)

	items := []struct{ key, label string }{
		{"1", "Rectangle"},
		{"2", "Circle"},
		{"3", "Line"},
	}
	PrintAt(3, 4, Bold+"Shape type:"+Reset)
	for _, item := range items {
		mark := "  "
		col := White
		if (item.key == "1" && a.selectedShape == models.ShapeRect) ||
			(item.key == "2" && a.selectedShape == models.ShapeCircle) ||
			(item.key == "3" && a.selectedShape == models.ShapeLine) {
			mark = BrightGreen + "▶ " + Reset + BrightGreen
			col = BrightGreen
		}
		row := 5
		if item.key == "2" { row = 6 }
		if item.key == "3" { row = 7 }
		PrintAtf(3, row, "%s%s%s"+Reset, mark, col, item.label)
	}

	PrintAt(3, 9, Bold+"Fill mode:"+Reset)
	fillOff := White + "  ○ Outline"
	fillOn  := White + "  ○ Filled"
	if !a.fillMode {
		fillOff = BrightGreen + "  ● Outline" + Reset
	} else {
		fillOn = BrightGreen + "  ● Filled" + Reset
	}
	PrintAt(3, 10, fillOff+Reset)
	PrintAt(3, 11, fillOn+Reset)

	PrintAt(3, 13, Bold+"Draw character:"+Reset+"  "+BrightYellow+Bold+string(a.drawChar)+Reset)

	PrintAt(1, 15, Dim+strings.Repeat("─", 50)+Reset)
	PrintAt(3, 16, Yellow+"1/2/3"+Reset+" shape   "+
		Yellow+"F"+Reset+" toggle fill   "+
		Yellow+"C"+Reset+" change char")
	PrintAt(3, 17, Yellow+"Enter"+Reset+" / "+Yellow+"Space"+Reset+" → draw on canvas   "+
		Yellow+"Esc"+Reset+" back")
}

func (a *App) renderHelp() {
	Clear()
	PrintAt(1, 1, BgBlue+BrightWhite+Bold+" HELP "+Reset)
	lines := []string{
		"",
		Bold + "  CANVAS CONTROLS" + Reset,
		"  " + Yellow + "Arrow keys" + Reset + "        Move cursor",
		"  " + Yellow + "Shift+Arrows" + Reset + "      Move cursor faster (×5)",
		"  " + Yellow + "Space / Enter" + Reset + "     Place point (first=start, second=draw)",
		"  " + Yellow + "Esc" + Reset + "               Cancel drawing / Go back",
		"  " + Yellow + "S" + Reset + "                 Open shape/fill/char menu",
		"  " + Yellow + "U" + Reset + "                 Undo last shape",
		"  " + Yellow + "X" + Reset + "                 Clear entire canvas",
		"  " + Yellow + "H" + Reset + "                 This help screen",
		"  " + Yellow + "Q" + Reset + "                 Quit to project list",
		"",
		Bold + "  DRAWING WORKFLOW" + Reset,
		"  1. Press " + Yellow + "S" + Reset + " to pick shape, fill mode, and character",
		"  2. Move cursor with arrow keys to start position",
		"  3. Press " + Yellow + "Space" + Reset + " to place start point",
		"  4. Move cursor to end position (live preview shown)",
		"  5. Press " + Yellow + "Space" + Reset + " again to draw the shape",
		"",
		Bold + "  SHAPES" + Reset,
		"  Rectangle  x1,y1 → x2,y2 corner",
		"  Circle     center x,y + radius (move away from center)",
		"  Line       x1,y1 → x2,y2 endpoint",
		"",
		"  Press any key to return...",
	}
	for i, l := range lines {
		PrintAt(1, 2+i, l)
	}
}

func (a *App) renderCanvas() {
	if a.project == nil {
		a.screen = ScreenProjects
		return
	}

	Clear()

	// Build the base grid from all saved shapes
	g := canvas.RenderAll(a.project.Width, a.project.Height, a.shapes)

	// If currently drawing, overlay preview
	if a.drawing {
		prev := canvas.NewGrid(a.project.Width, a.project.Height)
		ch := a.drawChar
		switch a.selectedShape {
		case models.ShapeRect:
			canvas.DrawRect(prev, a.startX, a.startY, a.cursorX, a.cursorY, ch, a.fillMode)
		case models.ShapeCircle:
			r := radius(a.startX, a.startY, a.cursorX, a.cursorY)
			canvas.DrawCircle(prev, a.startX, a.startY, r, ch, a.fillMode)
		case models.ShapeLine:
			canvas.DrawLine(prev, a.startX, a.startY, a.cursorX, a.cursorY, ch)
		}
		// Merge preview into base grid (use different color — handled in render)
		for y := 0; y < a.project.Height; y++ {
			for x := 0; x < a.project.Width; x++ {
				if prev.Get(x, y) != ' ' {
					g.Set(x, y, prev.Get(x, y))
				}
			}
		}
	}

	// Canvas offset on screen (leave room for header and sidebar)
	offX := 6  // left offset (for row numbers)
	offY := 3  // top offset (for header)

	// Header bar
	shapeLabel := map[models.ShapeType]string{
		models.ShapeRect:   "Rectangle",
		models.ShapeCircle: "Circle",
		models.ShapeLine:   "Line",
	}[a.selectedShape]
	fillLabel := "Outline"
	if a.fillMode { fillLabel = "Filled" }

	header := fmt.Sprintf(
		BgBlue+BrightWhite+Bold+" %s "+Reset+
			Dim+"  Shape:"+Reset+" %s"+
			Dim+"  Fill:"+Reset+" %s"+
			Dim+"  Char:"+Reset+" "+BrightYellow+"%s"+Reset+
			Dim+"  Pos:"+Reset+" %d,%d"+
			Dim+"  Shapes:"+Reset+" %d",
		a.project.Name, shapeLabel, fillLabel,
		string(a.drawChar), a.cursorX, a.cursorY, len(a.shapes),
	)
	PrintAt(1, 1, header)

	// Sub-header
	if a.drawing {
		PrintAt(1, 2, BrightYellow+"  DRAWING: "+Reset+
			"Move cursor to end point, press "+BrightGreen+"Space"+Reset+" to place  |  "+
			Yellow+"Esc"+Reset+" cancel")
	} else {
		PrintAt(1, 2, Dim+
			"  Arrows"+Reset+":move  "+
			Yellow+"Space"+Reset+":place  "+
			Yellow+"S"+Reset+":shape menu  "+
			Yellow+"U"+Reset+":undo  "+
			Yellow+"X"+Reset+":clear  "+
			Yellow+"H"+Reset+":help  "+
			Yellow+"Q"+Reset+":quit")
	}

	// Top canvas border
	PrintAt(offX, offY, Cyan+"┌"+strings.Repeat("─", a.project.Width)+"┐"+Reset)

	// Draw each row
	for y := 0; y < a.project.Height; y++ {
		screenRow := offY + 1 + y
		// Row number
		PrintAtf(1, screenRow, Dim+"%3d "+Reset, y)
		// Left border
		PrintAt(offX, screenRow, Cyan+"│"+Reset)

		// Cells
		for x := 0; x < a.project.Width; x++ {
			screenCol := offX + 1 + x
			ch := g.Get(x, y)

			isCursor := (x == a.cursorX && y == a.cursorY)
			isStart  := (a.drawing && x == a.startX && y == a.startY)

			if isCursor {
				PrintAt(screenCol, screenRow, BgGreen+Black+string(ch)+Reset)
			} else if isStart {
				PrintAt(screenCol, screenRow, BgYellow+Black+string(ch)+Reset)
			} else if ch != ' ' {
				PrintAt(screenCol, screenRow, BrightGreen+string(ch)+Reset)
			} else {
				PrintAt(screenCol, screenRow, Dim+"."+Reset)
			}
		}

		// Right border
		PrintAt(offX+a.project.Width+1, screenRow, Cyan+"│"+Reset)
	}

	// Bottom border
	PrintAt(offX, offY+a.project.Height+1, Cyan+"└"+strings.Repeat("─", a.project.Width)+"┘"+Reset)

	// X ruler
	rulerRow := offY + a.project.Height + 2
	PrintAt(offX+1, rulerRow, Dim)
	for i := 0; i < a.project.Width; i += 10 {
		PrintAtf(offX+1+i, rulerRow, "%-10d", i)
	}
	Print(Reset)

	// Status message
	if a.statusMsg != "" {
		PrintAt(1, rulerRow+2, BrightGreen+"  "+a.statusMsg+Reset)
		a.statusMsg = ""
	}
}

// ---------- key handling ----------

func (a *App) handleKey(key int) bool {
	switch a.screen {
	case ScreenProjects:
		return a.handleProjectsKey(key)
	case ScreenCanvas:
		return a.handleCanvasKey(key)
	case ScreenNewProject:
		return a.handleNewProjectKey(key)
	case ScreenShapeMenu:
		return a.handleShapeMenuKey(key)
	case ScreenHelp:
		a.screen = ScreenCanvas
		return true
	}
	return true
}

func (a *App) handleProjectsKey(key int) bool {
	switch key {
	case KeyArrowUp:
		if a.cursorY > 0 { a.cursorY-- }
	case KeyArrowDown:
		if a.cursorY < len(a.projects)-1 { a.cursorY++ }
	case KeyEnter:
		if len(a.projects) > 0 && a.cursorY < len(a.projects) {
			a.openProject(&a.projects[a.cursorY])
		}
	case 'n', 'N':
		a.inputBuf = ""
		a.screen = ScreenNewProject
	case 'd', 'D':
		if len(a.projects) > 0 && a.cursorY < len(a.projects) {
			a.deleteProject(a.projects[a.cursorY].ID)
		}
	case 'q', 'Q':
		return false
	}
	return true
}

func (a *App) handleNewProjectKey(key int) bool {
	switch key {
	case KeyEscape:
		a.screen = ScreenProjects
	case KeyEnter:
		if strings.TrimSpace(a.inputBuf) != "" {
			a.createProject(strings.TrimSpace(a.inputBuf))
			a.screen = ScreenCanvas
		}
	case KeyBackspace:
		if len(a.inputBuf) > 0 {
			runes := []rune(a.inputBuf)
			a.inputBuf = string(runes[:len(runes)-1])
		}
	default:
		if key >= 32 && key < 127 && len(a.inputBuf) < 30 {
			a.inputBuf += string(rune(key))
		}
	}
	return true
}

func (a *App) handleShapeMenuKey(key int) bool {
	switch key {
	case '1': a.selectedShape = models.ShapeRect
	case '2': a.selectedShape = models.ShapeCircle
	case '3': a.selectedShape = models.ShapeLine
	case 'f', 'F': a.fillMode = !a.fillMode
	case 'c', 'C':
		// Prompt for char inline
		PrintAt(3, 19, "Enter new draw character: "+BrightYellow)
		Flush()
		ch := ReadKey()
		if ch >= 32 && ch < 127 {
			a.drawChar = rune(ch)
		}
		Print(Reset)
	case KeyEnter, KeySpace:
		a.screen = ScreenCanvas
	case KeyEscape:
		a.screen = ScreenCanvas
	}
	return true
}

func (a *App) handleCanvasKey(key int) bool {
	step := 1
	switch key {
	// Movement — arrow keys move cursor
	case KeyArrowUp:
		if a.cursorY > 0 { a.cursorY -= step }
	case KeyArrowDown:
		if a.cursorY < a.project.Height-1 { a.cursorY += step }
	case KeyArrowLeft:
		if a.cursorX > 0 { a.cursorX -= step }
	case KeyArrowRight:
		if a.cursorX < a.project.Width-1 { a.cursorX += step }

	// Place point / draw
	case KeySpace, KeyEnter:
		if !a.drawing {
			// First point: set start
			a.drawing = true
			a.startX = a.cursorX
			a.startY = a.cursorY
			a.statusMsg = fmt.Sprintf("Start set at (%d,%d) — move and press Space to draw", a.startX, a.startY)
		} else {
			// Second point: draw shape and save
			a.commitShape()
			a.drawing = false
		}

	case KeyEscape:
		if a.drawing {
			a.drawing = false
			a.statusMsg = "Drawing cancelled"
		} else {
			a.screen = ScreenProjects
			a.loadProjects()
		}

	case 's', 'S':
		a.screen = ScreenShapeMenu

	case 'u', 'U':
		a.undoLast()

	case 'x', 'X':
		a.clearCanvas()

	case 'h', 'H':
		a.screen = ScreenHelp

	case 'q', 'Q':
		a.drawing = false
		a.screen = ScreenProjects
		a.project = nil
		a.cursorY = 0
		a.loadProjects()
	}
	return true
}

// ---------- data operations ----------

func (a *App) loadProjects() {
	db.DB.Preload("Shapes").Order("created_at desc").Find(&a.projects)
	if a.cursorY >= len(a.projects) {
		a.cursorY = 0
	}
}

func (a *App) openProject(p *models.Project) {
	a.project = p
	db.DB.Where("project_id = ? AND deleted_at IS NULL", p.ID).
		Order("created_at asc").Find(&a.shapes)
	a.cursorX = p.Width / 2
	a.cursorY = p.Height / 2
	a.drawing = false
	a.screen = ScreenCanvas
	a.statusMsg = fmt.Sprintf("Opened %q  —  press H for help", p.Name)
}

func (a *App) createProject(name string) {
	p := models.Project{Name: name, Width: 60, Height: 28}
	db.DB.Create(&p)
	a.project = &p
	a.shapes = nil
	a.cursorX = p.Width / 2
	a.cursorY = p.Height / 2
	a.statusMsg = fmt.Sprintf("Created %q", p.Name)
}

func (a *App) deleteProject(id uint) {
	db.DB.Delete(&models.Project{}, id)
	a.loadProjects()
	a.statusMsg = "Project deleted"
}

func (a *App) commitShape() {
	if a.project == nil { return }

	s := models.Shape{
		ProjectID: a.project.ID,
		Type:      a.selectedShape,
		X1:        a.startX,
		Y1:        a.startY,
		Filled:    a.fillMode,
		Char:      string(a.drawChar),
	}

	switch a.selectedShape {
	case models.ShapeRect, models.ShapeLine:
		s.X2 = a.cursorX
		s.Y2 = a.cursorY
	case models.ShapeCircle:
		s.Radius = radius(a.startX, a.startY, a.cursorX, a.cursorY)
		s.X2 = a.cursorX
		s.Y2 = a.cursorY
	}

	db.DB.Create(&s)
	a.shapes = append(a.shapes, s)
	a.statusMsg = fmt.Sprintf("Drew %s at (%d,%d) [%s] %s",
		string(a.selectedShape), a.startX, a.startY,
		string(a.drawChar),
		map[bool]string{true: "filled", false: "outline"}[a.fillMode])
}

func (a *App) undoLast() {
	if len(a.shapes) == 0 { return }
	last := a.shapes[len(a.shapes)-1]
	db.DB.Delete(&models.Shape{}, last.ID)
	a.shapes = a.shapes[:len(a.shapes)-1]
	a.statusMsg = "Undone last shape"
}

func (a *App) clearCanvas() {
	if a.project == nil { return }
	db.DB.Where("project_id = ?", a.project.ID).Delete(&models.Shape{})
	a.shapes = nil
	a.drawing = false
	a.statusMsg = "Canvas cleared"
}

// ---------- helpers ----------

func radius(x1, y1, x2, y2 int) int {
	dx := float64(x2-x1) * 0.5
	dy := float64(y2 - y1)
	r := int(sqrt(dx*dx + dy*dy))
	if r < 1 { r = 1 }
	return r
}

func sqrt(x float64) float64 {
	if x <= 0 { return 0 }
	z := x / 2
	for i := 0; i < 20; i++ {
		z -= (z*z - x) / (2 * z)
	}
	return z
}