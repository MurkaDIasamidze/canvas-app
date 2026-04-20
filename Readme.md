# ASCII Canvas Studio

A terminal drawing application written in pure Go.
Runs directly in **CMD**, **PowerShell**, **Linux** or **macOS** terminal.
No browser. No external TUI library. Everything built from raw ANSI escape codes.

---

## Quick Start

```powershell
# 1. Copy env file and fill in your DB credentials
copy .env.example .env

# 2. Download Go dependencies
go mod tidy

# 3. Run  (always use the dot — it compiles all files together)
go run .
```

---

## Project Structure

```
canvas-tui/
├── main.go                  ← entry point + ALL configuration (edit this)
├── go.mod                   ← Go dependencies
├── .env                     ← database credentials (copy from .env.example)
├── db/
│   └── db.go                ← PostgreSQL connection via GORM
├── models/
│   └── models.go            ← data types: Project, Shape, Color
├── canvas/
│   └── canvas.go            ← drawing engine (JS-canvas-style API)
└── tui/
    ├── app.go               ← all screens, input handling, render loop
    ├── platform_windows.go  ← Windows: ANSI enable, UTF-8, ReadConsoleInputW
    └── platform_unix.go     ← Linux/macOS: stty raw mode, stdin read
```

---

## Configuration

Everything is at the **top of `main.go`** — no separate config file.

```go
// Canvas size
const fullConsole = true   // true  → fills entire terminal window
                           // false → use fixed canvasW / canvasH below
const canvasW = 120
const canvasH = 40

// Starting tool: "rect"  "circle"  "line"  "free"
const defaultTool = "rect"

// Starting fill mode: false = outline,  true = filled
const defaultFill = false

// Starting color: "green" "cyan" "yellow" "red" "magenta" "blue" "white"
const defaultColor = "green"

// Characters used for drawing — change to anything you like
const rectCornerTL = '┌'   // top-left corner of rectangle
const rectTop      = '─'   // top edge
const rectFill     = '█'   // filled rectangle
const circleH      = '─'   // horizontal parts of circle
const lineH        = '─'   // horizontal line
const freeChar     = '█'   // freehand pixel dot
// ... all other chars are also configurable
```

---

## Drawing Tools

### Rectangle (`R`)
Draws a rectangle between two points you place.

- **Outline** — uses box-drawing characters:
  ```
  ┌────────────┐
  │            │
  └────────────┘
  ```
- **Filled** — fills the area with `█`

### Circle (`C`)
Draws a circle. First point = center, move cursor outward = radius.

- **Outline** — characters chosen by tangent angle at each point:
  ```
       ────
     ╱      ╲
    │          │
     ╲      ╱
       ────
  ```
- **Filled** — fills with `█`

### Line (`L`)
Draws a straight line between two points.
Character chosen automatically by slope:

| Direction   | Character |
|-------------|-----------|
| Horizontal  | `─`       |
| Vertical    | `│`       |
| Diagonal `\`| `╲`       |
| Diagonal `/`| `╱`       |

### Freehand (`F`)
A **pencil / pixel tool**. Each press of `Space` stamps one `█` cell at the cursor position. Move with arrow keys and press `Space` wherever you want to paint. Good for:
- Drawing irregular shapes
- Writing letters or patterns manually
- Filling gaps between other shapes

---

## Controls

### Project List

| Key | Action |
|-----|--------|
| `↑` `↓` | Navigate projects |
| `Enter` | Open project |
| `N` | New project |
| `D` | Delete project |
| `Q` | Quit |

### Canvas

| Key | Action |
|-----|--------|
| `↑` `↓` `←` `→` | Move cursor |
| `Space` | **1st press** = set start point · **2nd press** = draw shape |
| `O` | Open options menu (tool / fill / color) |
| `U` | Undo last shape |
| `X` | Clear entire canvas |
| `H` | Help screen |
| `Q` | Back to project list |
| `Esc` | Cancel current drawing |

> Freehand mode: each `Space` press paints one dot at the cursor.

### Options Menu (`O`)

| Key | Action |
|-----|--------|
| `R` | Switch to Rectangle tool |
| `C` | Switch to Circle tool |
| `L` | Switch to Line tool |
| `F` | Switch to Freehand tool |
| `T` | Toggle outline / filled |
| `←` `→` | Cycle through colors |
| `Enter` | Back to canvas |

---

## Drawing Workflow

```
1. Open or create a project from the project list

2. Press O → choose:
   - Tool (R / C / L / F)
   - Fill mode (T to toggle)
   - Color (← →)
   Press Enter to return to canvas

3. Move cursor with arrow keys to your start position

4. Press Space → start point is locked (shown in yellow)

5. Move cursor to end position
   → live preview of the shape follows your cursor

6. Press Space again → shape is drawn and saved to DB

7. Press U to undo,  X to clear everything
```

---

## Canvas API (`canvas/canvas.go`)

The drawing engine uses a **JavaScript canvas–style API**.
You can use it directly from Go code:

```go
// Create a canvas
ctx := canvas.New(80, 40)

// Set colors
ctx.StrokeColor = canvas.Green
ctx.FillColor   = canvas.Red

// Draw shapes
ctx.StrokeRect(5, 5, 30, 10)          // outline rectangle (x, y, width, height)
ctx.FillRect(5, 5, 30, 10)            // filled rectangle

ctx.StrokeCircle(40, 20, 8)           // outline circle (cx, cy, radius)
ctx.FillCircle(40, 20, 8)             // filled circle

ctx.DrawLine(0, 0, 79, 39)            // line from (x1,y1) to (x2,y2)

ctx.SetPixel(10, 10)                  // single dot at (x, y)

ctx.ClearRect(5, 5, 30, 10)           // erase a region
ctx.Clear()                           // erase everything

// Compositing
ctx.DrawContext(other, ox, oy)        // draw one canvas onto another

// Save / restore
snap := ctx.Snapshot()                // copy current state
ctx.Restore(snap)                     // put it back
```

Compared to JavaScript:

| JavaScript | Go (this project) |
|---|---|
| `ctx.strokeStyle = 'green'` | `ctx.StrokeColor = canvas.Green` |
| `ctx.strokeRect(x, y, w, h)` | `ctx.StrokeRect(x, y, w, h)` |
| `ctx.fillRect(x, y, w, h)` | `ctx.FillRect(x, y, w, h)` |
| `ctx.clearRect(x, y, w, h)` | `ctx.ClearRect(x, y, w, h)` |
| `ctx.arc(...); ctx.stroke()` | `ctx.StrokeCircle(cx, cy, r)` |
| `ctx.arc(...); ctx.fill()` | `ctx.FillCircle(cx, cy, r)` |
| `ctx.moveTo(); ctx.lineTo(); ctx.stroke()` | `ctx.DrawLine(x1, y1, x2, y2)` |
| `ctx.fillRect(x, y, 1, 1)` | `ctx.SetPixel(x, y)` |
| `ctx.drawImage(src, ox, oy)` | `ctx.DrawContext(src, ox, oy)` |

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.21 |
| Database | PostgreSQL |
| ORM | GORM |
| Terminal colors | Raw ANSI escape codes |
| Windows input | `ReadConsoleInputW` via `kernel32.dll` syscall |
| Unix input | `stty raw` + `os.Stdin.Read` |
| TUI library | None — built from scratch |

---

## Database Setup

```sql
-- PostgreSQL: the default 'postgres' database works fine
-- Just make sure the user and password match your .env
```

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=root
DB_NAME=postgres
```

All tables are created automatically on first run via GORM `AutoMigrate`.