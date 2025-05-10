package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
)

const (
	PrimaryColor    = lipgloss.Color("#1db954")
	SecondaryColor  = lipgloss.Color("#212121")
	BackgroundColor = lipgloss.Color("#121212")
	BorderColor     = lipgloss.Color("#535353")
	TextColor       = lipgloss.Color("#b3b3b3")
	WhiteTextColor  = lipgloss.Color("#ffffff")
)

const SHRINKWIDTH = 95

var RoundedTableBorders = table.Border{
	Top:    "─",
	Left:   "│",
	Right:  "│",
	Bottom: "─",

	TopRight:    "╮",
	TopLeft:     "╭",
	BottomRight: "╯",
	BottomLeft:  "╰",

	TopJunction:    "─",
	LeftJunction:   "├",
	RightJunction:  "┤",
	BottomJunction: "─",
	InnerJunction:  "─",

	InnerDivider: " ",
}

// https://github.com/charmbracelet/lipgloss/issues/209
// Look into this in order to consistently change the color of background
