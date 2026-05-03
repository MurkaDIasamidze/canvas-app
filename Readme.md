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
func presetShapes(ctx *Ctx) {
    ctx.strokeStyle = ColCyan
    strokeRect(ctx, 1, 1, 30, 10)

    ctx.fillStyle = ColYellow
    fillCircle(ctx, 50, 15, 8)

    ctx.strokeStyle = ColRed
    drawLine(ctx, 0, 0, 40, 20)

    ctx.fillStyle = ColWhite
    fillText(ctx, "hello!", 5, 5)

    ctx.strokeStyle = ColMagenta
    beginPath(ctx)
    moveTo(ctx, 10, 1)
    lineTo(ctx, 20, 10)
    lineTo(ctx, 0, 10)
    closePath(ctx)
    stroke(ctx)
}
```

Leave the body empty (or `_ = c`) to start with a blank canvas.

---

## Canvas API reference

The API is modelled **exactly** on the browser's Canvas 2D context.
Every drawing function takes `ctx` as its first argument — just like
JavaScript helper functions that receive `ctx`.

```
JS:                                   Go:
────────────────────────────────────  ──────────────────────────────────────────
const canvas = document.createElement canvas := newCanvas(120, 40)
const ctx = canvas.getContext("2d")   ctx    := getContext(canvas)
```

### Context style properties

```go
ctx.strokeStyle = ColCyan     // JS: ctx.strokeStyle = "cyan"
ctx.fillStyle   = ColYellow   // JS: ctx.fillStyle   = "yellow"
```

### Rectangles

```go
strokeRect(ctx, x, y, w, h)  // JS: ctx.strokeRect(x, y, w, h)   → ┌──┐ outline
fillRect(ctx, x, y, w, h)    // JS: ctx.fillRect(x, y, w, h)     → solid █
clearRect(ctx, x, y, w, h)   // JS: ctx.clearRect(x, y, w, h)    → erase region
```

### Circles

```go
strokeCircle(ctx, cx, cy, r) // JS: ctx.arc(...); ctx.stroke()   → ─ │ ╱ ╲ outline
fillCircle(ctx, cx, cy, r)   // JS: ctx.arc(...); ctx.fill()     → solid █
```

### Ellipses

```go
strokeEllipse(ctx, cx, cy, rx, ry)  // JS: ctx.ellipse(...); ctx.stroke()
fillEllipse(ctx, cx, cy, rx, ry)    // JS: ctx.ellipse(...); ctx.fill()
```

### Lines

```go
drawLine(ctx, x1, y1, x2, y2)  // shorthand — slope picks ─ │ ╱ ╲ automatically
```

### Path API

```go
beginPath(ctx)        // JS: ctx.beginPath()
moveTo(ctx, x, y)     // JS: ctx.moveTo(x, y)
lineTo(ctx, x, y)     // JS: ctx.lineTo(x, y)
closePath(ctx)        // JS: ctx.closePath()
stroke(ctx)           // JS: ctx.stroke()   — render path with strokeStyle
fill(ctx)             // JS: ctx.fill()     — fill closed path with fillStyle
```

### Triangles

```go
strokeTriangle(ctx, x1,y1, x2,y2, x3,y3)  // outline
fillTriangle(ctx, x1,y1, x2,y2, x3,y3)    // scanline fill
```

### Text

```go
fillText(ctx, "hello", x, y)   // JS: ctx.fillText("hello", x, y)
strokeText(ctx, "hi",  x, y)   // draws text inside a ┌─┐ border box
```

### Pixel

```go
setPixel(ctx, x, y)   // JS: ctx.fillRect(x, y, 1, 1)
```

### Canvas operations

```go
clearCanvas(ctx)                   // JS: ctx.clearRect(0,0,w,h)
snap := saveCanvas(ctx)            // JS: ctx.save()   — pixel snapshot
restoreCanvas(ctx, snap)           // JS: ctx.restore()
drawCanvas(dstCtx, srcCtx, ox, oy) // JS: dstCtx.drawImage(srcCanvas, ox, oy)
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