package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the model
func (m *Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	var b strings.Builder

	// Title bar
	title := m.renderTitle()
	b.WriteString(title)
	b.WriteString("\n")

	// Main visualization area
	if m.showHelp {
		b.WriteString(m.renderHelp())
	} else {
		b.WriteString(m.renderVisualization())
	}

	// Status bar
	b.WriteString("\n")
	b.WriteString(m.renderStatusBar())

	return b.String()
}

// renderTitle renders the title bar
func (m *Model) renderTitle() string {
	title := " Audio Visualizer "
	modeStr := fmt.Sprintf(" [%s] ", m.mode)
	deviceStr := fmt.Sprintf(" %s ", truncateString(m.GetDeviceName(), 30))

	// Calculate padding
	leftPart := TitleStyle.Render(title)
	modePart := ModeStyle.Render(modeStr)
	devicePart := DeviceStyle.Render(deviceStr)

	contentWidth := lipgloss.Width(leftPart) + lipgloss.Width(modePart) + lipgloss.Width(devicePart)
	padding := m.width - contentWidth
	if padding < 0 {
		padding = 0
	}

	paddingStr := strings.Repeat(" ", padding)

	return leftPart + modePart + paddingStr + devicePart
}

// renderVisualization renders the main visualization
func (m *Model) renderVisualization() string {
	visHeight := m.height - 4
	if visHeight < 1 {
		visHeight = 1
	}

	var visualization string
	switch m.mode {
	case ModeBars:
		if m.barRenderer != nil {
			m.barRenderer.SetOptions(m.useGradient, m.showPeaks, m.mirrorBars)
			visualization = m.barRenderer.Render(m.spectrum, m.peaks)
		}
	case ModeWaveform:
		if m.waveformRenderer != nil {
			m.waveformRenderer.SetOptions(m.useGradient, true)
			visualization = m.waveformRenderer.Render(m.spectrum)
		}
	}

	// Ensure we have enough lines
	lines := strings.Split(visualization, "\n")
	for len(lines) < visHeight {
		lines = append(lines, "")
	}

	return strings.Join(lines[:visHeight], "\n")
}

// renderHelp renders the help overlay
func (m *Model) renderHelp() string {
	var lines []string
	lines = append(lines, "")
	lines = append(lines, HelpStyle.Render(" Keyboard Shortcuts "))
	lines = append(lines, "")

	for _, kb := range KeyBindings {
		key := HelpKeyStyle.Render(kb.Key)
		desc := HelpDescStyle.Render(kb.Description)
		lines = append(lines, "  "+key+"  "+desc)
	}

	lines = append(lines, "")
	lines = append(lines, HelpDescStyle.Render("  Press 'h' or '?' to close"))

	// Center the help content
	visHeight := m.height - 4
	helpContent := strings.Join(lines, "\n")
	helpLines := strings.Split(helpContent, "\n")

	// Pad to fill height
	for len(helpLines) < visHeight {
		helpLines = append(helpLines, "")
	}

	return strings.Join(helpLines[:visHeight], "\n")
}

// renderStatusBar renders the bottom status bar
func (m *Model) renderStatusBar() string {
	// Left side: status message or controls hint
	leftPart := "  Press 'h' for help"
	if m.statusMsg != "" {
		leftPart = "  " + m.statusMsg
	}

	// Right side: FPS and settings
	rightPart := fmt.Sprintf("FPS: %d | Sens: %.1f | Smooth: %.2f  ",
		m.fps,
		m.config.Sensitivity.Sensitivity,
		m.config.Smoothing.AttackSpeed,
	)

	// Calculate padding
	contentWidth := len(leftPart) + len(rightPart)
	padding := m.width - contentWidth
	if padding < 0 {
		padding = 0
	}
	paddingStr := strings.Repeat(" ", padding)

	statusLeft := StatusBarStyle.Render(leftPart)
	statusRight := StatusBarStyle.Render(rightPart)

	return statusLeft + paddingStr + statusRight
}

// truncateString truncates a string to max length with ellipsis
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// renderError renders an error message (used during initialization)
func renderError(err error) string {
	return ErrorStyle.Render(fmt.Sprintf("Error: %v", err))
}

// RenderLoadingScreen renders a loading screen
func RenderLoadingScreen() string {
	return lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true).
		Render("Loading audio visualizer...")
}
