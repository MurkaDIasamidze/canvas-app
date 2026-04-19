# ASCII Canvas Studio

A terminal drawing app. Runs entirely in CMD, PowerShell, or any Unix terminal.
No browser. No external TUI library. Pure Go + ANSI escape codes.

---

## Folder structure

```
canvas-tui/
├── main.go          ← entry point + ALL config (edit this)
├── go.mod
├── .env             ← database credentials
├── db/
│   └── db.go        ← PostgreSQL connection
├── models/
│   └── models.go    ← data structures
├── canvas/
│   └── canvas.go    ← drawing engine
└── tui/
    ├── app.go             ← main app logic
    ├── platform_windows.go ← Windows terminal + input
    └── platform_unix.go   ← Linux/macOS terminal + input
```

---

## Setup

### 1. PostgreSQL

```sql
-- run in psql
CREATE DATABASE postgres; -- already exists by default
```

### 2. Environment

```powershell
copy .env.example .env
```

Edit `.env`:
```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=root
DB_NAME=postgres
```

### 3. Dependencies

```powershell
go mod tidy
```

### 4. Run

```powershell
go run .
```

> **Important:** always use `go run .` (with the dot), not `go run main.go`.
> The dot tells Go to compile all `.go` files in the directory together.

---

## Configuration

Everything is in `main.go` at the top — no separate config file needed.

```go
// Canvas fills the full terminal window (true) or use fixed size (false)
const fullConsole = true
const canvasW     = 120   // used only when fullConsole = false
const canvasH     = 40    // used only when fullConsole = false

// Default tool: "rect" "circle" "line" "free"
const defaultTool = "rect"

// Default fill: false = outline, true = filled
const defaultFill = false

// Default color: "green" "cyan" "yellow" "red" "magenta" "blue" "white"
const defaultColor = "green"

// Characters — change to any ASCII/Unicode char you like
const rectFill     = '█'   // filled rectangle
const rectCornerTL = '┌'   // top-left corner
const lineH        = '─'   // horizontal line
const freeChar     = '█'   // freehand paint
// ... etc
```

---

## Controls

### Project list

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
| `Space` | Place start point, then draw on second press |
| `O` | Open options menu (tool / fill / color) |
| `U` | Undo last shape |
| `X` | Clear all shapes |
| `H` | Help screen |
| `Q` | Back to project list |
| `Esc` | Cancel current drawing |

### Options menu (`O`)

| Key | Action |
|-----|--------|
| `R` | Rectangle tool |
| `C` | Circle tool |
| `L` | Line tool |
| `F` | Freehand tool |
| `T` | Toggle outline / filled |
| `←` `→` | Change color |
| `Enter` | Back to canvas |

---

## Drawing workflow

1. Open or create a project
2. Press `O` → choose tool, fill mode, color
3. Move cursor with arrow keys to start position
4. Press `Space` — start point is set (shown in yellow)
5. Move cursor to end position — **live preview** shows the shape
6. Press `Space` again — shape is drawn and saved to database

For **freehand**: just move cursor and press `Space` to paint individual cells.

---

## Shapes and their characters

| Shape | Outline | Filled |
|-------|---------|--------|
| Rectangle | `┌─┐ │ └┘` | `█` |
| Circle | `─ │ ╱ ╲` (by tangent angle) | `█` |
| Line | `─ │ ╱ ╲` (by slope) | — |
| Freehand | — | `█` |

All characters are configurable in `main.go`.

---

## Tech stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.21 |
| Database | PostgreSQL |
| ORM | GORM |
| Terminal colors | Raw ANSI escape codes (no library) |
| Windows input | `ReadConsoleInputW` via `kernel32.dll` |
| Unix input | `stty raw` + `os.Stdin.Read` |
| TUI framework | None — built from scratch |