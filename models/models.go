package models

import (
	"time"
	"gorm.io/gorm"
)

type Project struct {
	ID        uint           `gorm:"primaryKey"`
	Name      string         `gorm:"not null"`
	Width     int
	Height    int
	Shapes    []Shape        `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type ShapeType string

const (
	ShapeRect   ShapeType = "rect"
	ShapeCircle ShapeType = "circle"
	ShapeLine   ShapeType = "line"
	ShapeFree   ShapeType = "free"
)

type ColorName string

const (
	ColGreen   ColorName = "green"
	ColCyan    ColorName = "cyan"
	ColYellow  ColorName = "yellow"
	ColRed     ColorName = "red"
	ColMagenta ColorName = "magenta"
	ColBlue    ColorName = "blue"
	ColWhite   ColorName = "white"
)

var AllColors = []ColorName{
	ColGreen, ColCyan, ColYellow, ColRed, ColMagenta, ColBlue, ColWhite,
}

type Shape struct {
	ID        uint           `gorm:"primaryKey"`
	ProjectID uint           `gorm:"not null;index"`
	Type      ShapeType
	X1, Y1    int
	X2, Y2    int
	Radius    int
	Filled    bool
	Color     ColorName      `gorm:"default:'green'"`
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}