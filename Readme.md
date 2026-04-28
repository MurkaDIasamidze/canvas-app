# console-draw

A terminal drawing app written in pure Go — no external libraries, no framework, no database.  
Draws directly to the console using ANSI escape codes and Unicode box-drawing characters.

```
┌──────────────────────────────────────────────┐
│  ◯   ╱╲   ─────   ▭   ·····   FillText!    │
└──────────────────────────────────────────────┘
```

---

## Files

| File | What it does |
|---|---|
| `config.go` | **All settings** — size, chars, default tool, preset shapes |
| `canvas.go` | Drawing engine — JS-canvas-style API |
| `shape.go` | In-memory shape list + renderer |
| `main.go` | App loop, screens, keyboard input |
| `term.go` | Shared ANSI codes, colors, buffered output |
| `term_unix.go` | Unix raw mode + key reading (`stty`) |
| `term_windows.go` | Windows raw mode + key reading (`kernel32.dll`) |

---

## Run

```bash
go run .
```

Or build a binary:

```bash
go build -o draw .
./draw
```

No `go get` needed — zero dependencies.

---

## Controls

| Key | Action |
|---|---|
| `↑ ↓ ← →` | Move cursor |
| `Space` / `Enter` | **1st press** = set start point · **2nd press** = commit shape |
| `Esc` | Cancel drawing in progress |
| `M` | Open options menu (tool / fill / color) |
| `U` | Undo last shape |
| `X` | Clear all shapes |
| `H` | Help screen |
| `Q` | Quit |

**In the menu:**

| Key | Action |
|---|---|
| `R` `C` `L` `F` | Switch tool (Rect / Circle / Line / Freehand) |
| `T` | Toggle outline ↔ filled |
| `← →` | Cycle color |
| `Enter` / `Esc` | Back to canvas |

---

## Changing the canvas size

Open **`config.go`** and edit the size section at the top:

```go
// fullConsole = true  → fills your entire terminal window (recommended)
// fullConsole = false → use the fixed width/height below
const fullConsole = true

const (
    canvasW = 120   // used only when fullConsole = false
    canvasH = 40
)
```

**Full terminal (default):** `fullConsole = true` — the canvas always fills whatever terminal you have open. Resize your terminal to get more space.

**Fixed size:** Set `fullConsole = false` and pick any `canvasW` / `canvasH`. Useful if you want a predictable viewport or are embedding preset artwork.

---

## Changing the drawing characters

Every character used for drawing is a constant in **`config.go`**.  
Change any of them and rebuild — nothing else needs to touch.

### Rectangle outline

```go
const (
    rectTL  = '┌'  // top-left corner
    rectTR  = '┐'  // top-right corner
    rectBL  = '└'  // bottom-left corner
    rectBR  = '┘'  // bottom-right corner
    rectTop    = '─'  // top edge
    rectBottom = '─'  // bottom edge
    rectLeft   = '│'  // left edge
    rectRight  = '│'  // right edge
)
```

**ASCII fallback** (for terminals without Unicode):

```go
const (
    rectTL = '+';  rectTR = '+';  rectBL = '+';  rectBR = '+'
    rectTop = '-'; rectBottom = '-'
    rectLeft = '|'; rectRight = '|'
)
```

### Rectangle fill

```go
const rectFill = '█'   // try: '▓' '▒' '░' '#' '*' '@'
```

### Circle outline

Characters are picked automatically by the tangent angle at each pixel:

```go
const (
    circH  = '─'   // horizontal segments  (~0° / 180°)
    circV  = '│'   // vertical segments    (~90°)
    circD1 = '╱'   // forward diagonal     (~45°)
    circD2 = '╲'   // back diagonal        (~135°)
)
```

**ASCII fallback:**

```go
const (
    circH = '-';  circV = '|';  circD1 = '/';  circD2 = '\\'
)
```

### Circle fill

```go
const circFill = '█'   // try: 'O' '*' '@'
```

### Line

```go
const (
    lineH  = '─'   // horizontal line
    lineV  = '│'   // vertical line
    lineD1 = '╱'   // diagonal /
    lineD2 = '╲'   // diagonal \
)
```

### Freehand pixel

```go
const freeChar = '█'   // try: '·' '•' '+' '*'
```

---

## Changing default tool, fill, and color

Also in `config.go`:

```go
const defaultTool  = KindRect    // KindRect | KindCircle | KindLine | KindFree
const defaultFill  = false       // false = outline   true = filled
const defaultColor = ColGreen    // ColGreen | ColCyan | ColYellow | ColRed
                                 // ColMagenta | ColBlue | ColWhite
```

---

## Drawing preset shapes at startup

`config.go` includes a `presetShapes` function that runs once when the app starts.  
Anything you draw here appears as a permanent background layer — it cannot be undone  
or erased during the session, but interactive shapes draw on top of it normally.

```go
func presetShapes(c *Canvas) {
    // box in the top-left
    c.StrokeColor = ColCyan
    c.StrokeRect(1, 1, 30, 10)

    // filled circle in the center
    c.FillColor = ColYellow
    c.FillCircle(60, 20, 8)

    // diagonal line
    c.StrokeColor = ColRed
    c.DrawLine(0, 0, 40, 20)

    // text label
    c.FillColor = ColWhite
    c.FillText("hello!", 5, 5)

    // path / polygon
    c.StrokeColor = ColMagenta
    c.BeginPath()
    c.MoveTo(10, 1)
    c.LineTo(20, 10)
    c.LineTo(0, 10)
    c.ClosePath()
    c.Stroke()
}
```

Leave the body empty (or `_ = c`) to start with a blank canvas.

---

## Canvas API reference

The `Canvas` type in `canvas.go` mirrors the browser's `CanvasRenderingContext2D`.

### Properties

```go
c.StrokeColor = ColCyan    // color for Stroke* methods, DrawLine, SetPixel
c.FillColor   = ColYellow  // color for Fill* methods, FillText
```

### Rectangles

```go
c.StrokeRect(x, y, w, h)   // ┌──┐ outline box
c.FillRect(x, y, w, h)     // solid filled box
c.ClearRect(x, y, w, h)    // erase region (make blank)
```

### Circles

```go
c.StrokeCircle(cx, cy, r)   // outline circle — chars chosen by angle
c.FillCircle(cx, cy, r)     // solid filled circle
```

### Ellipses

```go
c.StrokeEllipse(cx, cy, rx, ry)   // outline ellipse (horizontal r, vertical r)
c.FillEllipse(cx, cy, rx, ry)     // solid filled ellipse
```

### Lines

```go
c.DrawLine(x1, y1, x2, y2)   // straight line — char chosen by slope
```

### Path API (like JS)

```go
c.BeginPath()          // reset path
c.MoveTo(x, y)         // move pen without drawing
c.LineTo(x, y)         // add segment
c.ClosePath()          // connect last point back to first
c.Stroke()             // render the path with StrokeColor
```

### Triangles

```go
c.StrokeTriangle(x1,y1, x2,y2, x3,y3)   // outline
c.FillTriangle(x1,y1, x2,y2, x3,y3)     // solid (scanline fill)
```

### Text

```go
c.FillText("hello", x, y)    // plain text at position
c.StrokeText("hi",  x, y)    // text inside a ┌─┐ border box
```

### Pixel

```go
c.SetPixel(x, y)   // single pixel using freeChar + StrokeColor
```

### Canvas operations

```go
c.Clear()                    // wipe everything blank
snap := c.Save()             // deep-copy buffer (returns [][]Cell)
c.Restore(snap)              // revert to snapshot
c.DrawCanvas(other, ox, oy)  // composite another canvas on top at offset
```

---

## Colors

| Constant | Color |
|---|---|
| `ColGreen` | Bright green |
| `ColCyan` | Bright cyan |
| `ColYellow` | Bright yellow |
| `ColRed` | Bright red |
| `ColMagenta` | Bright magenta |
| `ColBlue` | Bright blue |
| `ColWhite` | Bright white |

---

## Quick example — custom ASCII art preset

```go
// config.go

const fullConsole = false
const canvasW = 80
const canvasH = 24

const (
    rectTL = '+'; rectTR = '+'; rectBL = '+'; rectBR = '+'
    rectTop = '-'; rectBottom = '-'
    rectLeft = '|'; rectRight = '|'
)
const rectFill  = '#'
const circFill  = 'O'
const freeChar  = '*'

func presetShapes(c *Canvas) {
    c.StrokeColor = ColGreen
    c.StrokeRect(0, 0, 80, 24)   // border around the whole canvas

    c.FillColor = ColYellow
    c.FillText("ASCII CANVAS", 33, 11)
}
```