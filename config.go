package main

// ═══════════════════════════════════════════════════════════════
//  CONFIG — change anything here to customise the app behaviour
// ═══════════════════════════════════════════════════════════════

var Cfg = Config{

	// ── Canvas ────────────────────────────────────────────────
	// FullConsole: true  → canvas fills the entire terminal
	//              false → use CanvasW / CanvasH below
	FullConsole: true,
	CanvasW:     120, // used only when FullConsole = false
	CanvasH:     40,  // used only when FullConsole = false

	// Show row/col numbers along the edges
	ShowRulers: true,

	// Character used for empty cells  (' ' = invisible)
	EmptyCell: ' ',

	// ── Shapes ───────────────────────────────────────────────
	// Characters used for each shape element.
	// You can change any of these to anything you like.
	Chars: CharConfig{
		// Rectangle outline
		RectTop:         '─',
		RectBottom:      '─',
		RectLeft:        '│',
		RectRight:       '│',
		RectCornerTL:    '┌',
		RectCornerTR:    '┐',
		RectCornerBL:    '└',
		RectCornerBR:    '┘',
		// Rectangle filled
		RectFill:        '█',

		// Circle outline chars (chosen per tangent angle)
		CircleH:         '─', // horizontal parts
		CircleV:         '│', // vertical parts
		CircleDiag1:     '╱', // / diagonal
		CircleDiag2:     '╲', // \ diagonal
		// Circle filled
		CircleFill:      '█',

		// Line chars (chosen per slope)
		LineH:           '─',
		LineV:           '│',
		LineDiag1:       '╱',
		LineDiag2:       '╲',

		// Freehand draw char (for paint mode)
		FreeChar:        '█',
	},

	// ── Cursor ───────────────────────────────────────────────
	CursorChar:   '█', // character shown at cursor position
	StartChar:    '·', // character shown at drawing start point

	// ── Colors ───────────────────────────────────────────────
	// Default color when app starts
	DefaultColor: "green",

	// ── Default tool when app starts ─────────────────────────
	// Options: "rect" "circle" "line" "free"
	DefaultTool: "rect",

	// ── Default fill mode ────────────────────────────────────
	DefaultFill: false,

	// ── Canvas border ────────────────────────────────────────
	// Draw a border around the canvas area
	ShowBorder: true,

	// ── Project defaults ─────────────────────────────────────
	DefaultProjectW: 120,
	DefaultProjectH: 40,
}

// ─────────────────────────────────────────────────────────────
// Types — you don't need to edit below this line normally
// ─────────────────────────────────────────────────────────────

type Config struct {
	FullConsole     bool
	CanvasW         int
	CanvasH         int
	ShowRulers      bool
	EmptyCell       rune
	Chars           CharConfig
	CursorChar      rune
	StartChar       rune
	DefaultColor    string
	DefaultTool     string
	DefaultFill     bool
	ShowBorder      bool
	DefaultProjectW int
	DefaultProjectH int
}

type CharConfig struct {
	RectTop, RectBottom, RectLeft, RectRight rune
	RectCornerTL, RectCornerTR              rune
	RectCornerBL, RectCornerBR              rune
	RectFill                                rune
	CircleH, CircleV, CircleDiag1, CircleDiag2 rune
	CircleFill                              rune
	LineH, LineV, LineDiag1, LineDiag2      rune
	FreeChar                                rune
}