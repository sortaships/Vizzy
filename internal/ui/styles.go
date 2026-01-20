package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	ColorPrimary   = lipgloss.Color("#00FFFF")
	ColorSecondary = lipgloss.Color("#FF00FF")
	ColorAccent    = lipgloss.Color("#FFFF00")
	ColorMuted     = lipgloss.Color("#666666")
	ColorBright    = lipgloss.Color("#FFFFFF")
	ColorDim       = lipgloss.Color("#333333")
	ColorError     = lipgloss.Color("#FF0000")
	ColorSuccess   = lipgloss.Color("#00FF00")
)

// Rainbow colors for spectrum visualization
var RainbowColors = []lipgloss.Color{
	"#FF0000", // Red
	"#FF7F00", // Orange
	"#FFFF00", // Yellow
	"#00FF00", // Green
	"#0000FF", // Blue
	"#4B0082", // Indigo
	"#9400D3", // Violet
}

// GradientColors for gradient visualization
var GradientColors = []lipgloss.Color{
	"#00FFFF", // Cyan
	"#00FF00", // Green
	"#FFFF00", // Yellow
	"#FF7F00", // Orange
	"#FF0000", // Red
}

// Styles for various UI elements
var (
	// Base styles
	BaseStyle = lipgloss.NewStyle()

	// Title bar
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorBright).
			Background(lipgloss.Color("#333366")).
			Padding(0, 1)

	// Status bar
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Background(lipgloss.Color("#1A1A1A")).
			Padding(0, 1)

	StatusKeyStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	StatusValueStyle = lipgloss.NewStyle().
			Foreground(ColorBright)

	// Help overlay
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorBright).
			Background(lipgloss.Color("#1A1A3A")).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			Width(12)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// Visualization area
	VisualizationStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Device selector
	DeviceStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary)

	DeviceSelectedStyle = lipgloss.NewStyle().
			Foreground(ColorBright).
			Background(ColorPrimary).
			Bold(true)

	// Mode indicator
	ModeStyle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)

	// Error message
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)
)

// GetBarColor returns a color for a bar at the given position (0-1)
func GetBarColor(pos float64, useGradient bool) lipgloss.Color {
	if useGradient {
		return getGradientColor(pos)
	}
	return getRainbowColor(pos)
}

// getRainbowColor returns a rainbow color based on position (0-1)
func getRainbowColor(pos float64) lipgloss.Color {
	if pos < 0 {
		pos = 0
	}
	if pos >= 1 {
		pos = 0.999
	}
	index := int(pos * float64(len(RainbowColors)))
	return RainbowColors[index]
}

// getGradientColor returns a gradient color based on height (0-1)
func getGradientColor(height float64) lipgloss.Color {
	if height < 0 {
		height = 0
	}
	if height >= 1 {
		height = 0.999
	}
	index := int(height * float64(len(GradientColors)))
	return GradientColors[index]
}

// GetHeightColor returns a color based on bar height for gradient coloring
func GetHeightColor(height float64) lipgloss.Color {
	return getGradientColor(height)
}

// BarStyle returns a style for a bar character with the given color
func BarStyle(color lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(color)
}

// PeakStyle returns a style for peak indicators
func PeakStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ColorBright)
}
