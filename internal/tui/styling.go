package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
)

var (
	PrimaryColor = lipgloss.AdaptiveColor{
		Light: "#1db954",
		Dark:  "#1db954",
	}
	SecondaryColor = lipgloss.AdaptiveColor{
		Light: "#f5f5f5",
		Dark:  "#212121",
	}
	BorderColor = lipgloss.AdaptiveColor{
		Light: "#e0e0e0",
		Dark:  "#535353",
	}
	TextColor = lipgloss.AdaptiveColor{
		Light: "#333333",
		Dark:  "#b3b3b3",
	}
	WhiteTextColor = lipgloss.AdaptiveColor{
		Light: "#000000",
		Dark:  "#ffffff",
	}
	DangerColor = lipgloss.AdaptiveColor{
		Light: "#ff4444",
		Dark:  "#ff4444",
	}
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
