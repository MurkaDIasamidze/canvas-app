package models

import (
	"time"

	"gorm.io/gorm"
)

type Project struct {
	ID        uint           `gorm:"primaryKey"`
	Name      string         `gorm:"not null"`
	Width     int            `gorm:"default:60"`
	Height    int            `gorm:"default:28"`
	Shapes    []Shape        `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type ShapeType string

const (
	ShapeRect   ShapeType = "rectangle"
	ShapeCircle ShapeType = "circle"
	ShapeLine   ShapeType = "line"
)

// ColorName is an ANSI color choice stored by name
type ColorName string

const (
	ColorGreen   ColorName = "green"
	ColorCyan    ColorName = "cyan"
	ColorYellow  ColorName = "yellow"
	ColorRed     ColorName = "red"
	ColorMagenta ColorName = "magenta"
	ColorBlue    ColorName = "blue"
	ColorWhite   ColorName = "white"
)

// AllColors lists every selectable color in menu order
var AllColors = []ColorName{
	ColorGreen, ColorCyan, ColorYellow,
	ColorRed, ColorMagenta, ColorBlue, ColorWhite,
}

type Shape struct {
	ID        uint           `gorm:"primaryKey"`
	ProjectID uint           `gorm:"not null;index"`
	Type      ShapeType      `gorm:"not null"`
	X1        int
	Y1        int
	X2        int
	Y2        int
	Radius    int
	Filled    bool
	Char      string    `gorm:"default:'*'"`
	Color     ColorName `gorm:"default:'green'"`
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}