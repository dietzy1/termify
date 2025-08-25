package tui

import (
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/evertras/bubble-table/table"
)

var (
	PrimaryColor    = lipgloss.Color("#1db954")
	SecondaryColor  = lipgloss.Color("#212121")
	BackgroundColor = lipgloss.Color("#121212")
	BorderColor     = lipgloss.Color("#535353")
	TextColor       = lipgloss.Color("#b3b3b3")
	WhiteTextColor  = lipgloss.Color("#ffffff")
)

/* const (
	PrimaryColor    = lipgloss.Color("2")  // Green (closest to your Spotify green)
	SecondaryColor  = lipgloss.Color("0")  // Black
	BackgroundColor = lipgloss.Color("0")  // Black
	BorderColor     = lipgloss.Color("8")  // Bright black/gray
	TextColor       = lipgloss.Color("7")  // White/light gray
	WhiteTextColor  = lipgloss.Color("15") // Bright white
) */

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
