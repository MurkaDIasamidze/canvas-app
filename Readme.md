# ASCII Canvas Studio — Pure Terminal TUI

A fully interactive terminal drawing app. **No browser, no frontend.**
Runs directly in CMD, PowerShell, or any Unix terminal.

## Folder Structure

```
canvas-tui/
├── main.go
├── go.mod
├── .env               ← copy from .env.example
├── db/
│   └── db.go          ← PostgreSQL connection
├── models/
│   └── models.go      ← Project + Shape models
├── canvas/
│   └── canvas.go      ← Drawing engine (rect/circle/line)
└── tui/
    ├── app.go          ← Main TUI app & all screens
    ├── term.go         ← ANSI terminal helpers
    ├── init_windows.go ← Windows: enable ANSI + UTF-8 via kernel32
    ├── init_unix.go    ← Unix: no-op
    ├── input_windows.go← Windows: ReadConsoleInputW (arrow keys etc)
    ├── input_unix.go   ← Unix: raw mode via stty
    ├── raw_windows.go  ← Windows: raw mode stubs
    └── raw_unix.go     ← Unix: raw mode wrappers
```

## Setup

### 1. Copy .env
```
copy .env.example .env
```
Edit `.env` with your DB credentials (defaults match your setup).

### 2. Install dependencies
```
go mod tidy
```

### 3. Run
```
go run main.go
```

## Controls

### Project List
| Key | Action |
|-----|--------|
| ↑ ↓ | Navigate projects |
| Enter | Open project |
| N | New project |
| D | Delete project |
| Q | Quit |

### Canvas (inside a project)
| Key | Action |
|-----|--------|
| ↑ ↓ ← → | Move cursor |
| Space / Enter | Place point (1st = start, 2nd = draw) |
| S | Open shape/fill/char menu |
| U | Undo last shape |
| X | Clear entire canvas |
| H | Help screen |
| Q | Back to project list |
| Esc | Cancel drawing / back |

### Shape Menu (press S)
| Key | Action |
|-----|--------|
| 1 | Rectangle |
| 2 | Circle |
| 3 | Line |
| F | Toggle filled / outline |
| C | Change draw character |
| Enter | Back to canvas |

## How Drawing Works

1. Press **S** → pick shape, fill mode, and character
2. Move cursor with **arrow keys** to where you want to start
3. Press **Space** — this sets the start point (shown in yellow)
4. Move cursor to the end point — **live preview** shows as you move
5. Press **Space** again — shape is drawn and saved to DB

For circles: start = center, end = any point on the edge (sets radius).