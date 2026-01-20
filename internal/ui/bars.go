package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Block characters for bar visualization (increasing height)
var barBlocks = []rune{' ', '▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// BarRenderer renders spectrum data as vertical bars
type BarRenderer struct {
	width       int
	height      int
	useGradient bool
	showPeaks   bool
	mirror      bool
}

// NewBarRenderer creates a new bar renderer
func NewBarRenderer(width, height int) *BarRenderer {
	return &BarRenderer{
		width:       width,
		height:      height,
		useGradient: true,
		showPeaks:   true,
		mirror:      false,
	}
}

// SetSize updates the renderer dimensions
func (r *BarRenderer) SetSize(width, height int) {
	r.width = width
	r.height = height
}

// SetOptions updates rendering options
func (r *BarRenderer) SetOptions(useGradient, showPeaks, mirror bool) {
	r.useGradient = useGradient
	r.showPeaks = showPeaks
	r.mirror = mirror
}

// Render renders the spectrum as vertical bars
func (r *BarRenderer) Render(spectrum, peaks []float64) string {
	if r.width <= 0 || r.height <= 0 || len(spectrum) == 0 {
		return ""
	}

	// Calculate how many bars we can fit and how to map spectrum data
	numBars := r.width
	if numBars > len(spectrum) {
		numBars = len(spectrum)
	}

	// Map spectrum data to bars
	barValues := make([]float64, numBars)
	peakValues := make([]float64, numBars)
	step := float64(len(spectrum)) / float64(numBars)

	for i := 0; i < numBars; i++ {
		// Average the spectrum values for this bar
		start := int(float64(i) * step)
		end := int(float64(i+1) * step)
		if end > len(spectrum) {
			end = len(spectrum)
		}
		if start >= end {
			start = end - 1
		}
		if start < 0 {
			start = 0
		}

		var sum, peakSum float64
		count := 0
		for j := start; j < end; j++ {
			sum += spectrum[j]
			if peaks != nil && j < len(peaks) {
				peakSum += peaks[j]
			}
			count++
		}
		if count > 0 {
			barValues[i] = sum / float64(count)
			peakValues[i] = peakSum / float64(count)
		}
	}

	// Build the visualization row by row (top to bottom)
	var lines []string
	effectiveHeight := r.height
	if r.mirror {
		effectiveHeight = r.height / 2
	}

	for row := effectiveHeight - 1; row >= 0; row-- {
		var line strings.Builder
		threshold := float64(row) / float64(effectiveHeight)

		for col := 0; col < numBars; col++ {
			value := barValues[col]
			peakValue := peakValues[col]
			char := r.getBarChar(value, threshold, peakValue)

			// Determine color
			var color lipgloss.Color
			if r.useGradient {
				color = GetHeightColor(value)
			} else {
				color = GetBarColor(float64(col)/float64(numBars), false)
			}

			// Check if this is a peak position
			if r.showPeaks && value < threshold && peakValue >= threshold && peakValue < threshold+1.0/float64(effectiveHeight) {
				line.WriteString(PeakStyle().Render("▔"))
			} else {
				line.WriteString(BarStyle(color).Render(string(char)))
			}
		}
		lines = append(lines, line.String())
	}

	// If mirroring, add the mirrored bottom half
	if r.mirror {
		for i := 0; i < effectiveHeight; i++ {
			var line strings.Builder
			threshold := float64(i) / float64(effectiveHeight)

			for col := 0; col < numBars; col++ {
				value := barValues[col]
				char := r.getBarCharInverted(value, threshold)

				var color lipgloss.Color
				if r.useGradient {
					color = GetHeightColor(value)
				} else {
					color = GetBarColor(float64(col)/float64(numBars), false)
				}

				line.WriteString(BarStyle(color).Render(string(char)))
			}
			lines = append(lines, line.String())
		}
	}

	return strings.Join(lines, "\n")
}

// getBarChar returns the appropriate bar character for the given value and threshold
func (r *BarRenderer) getBarChar(value, threshold, peakValue float64) rune {
	if value <= threshold {
		return ' '
	}

	// Calculate how much of this row is filled
	rowHeight := 1.0 / float64(r.height)
	fillAmount := (value - threshold) / rowHeight

	if fillAmount >= 1 {
		return '█'
	}

	// Map to partial block character
	index := int(fillAmount * float64(len(barBlocks)-1))
	if index < 0 {
		index = 0
	}
	if index >= len(barBlocks) {
		index = len(barBlocks) - 1
	}

	return barBlocks[index]
}

// getBarCharInverted returns bar character for the mirrored (inverted) bottom half
func (r *BarRenderer) getBarCharInverted(value, threshold float64) rune {
	if value <= threshold {
		return ' '
	}

	rowHeight := 1.0 / float64(r.height/2)
	fillAmount := (value - threshold) / rowHeight

	if fillAmount >= 1 {
		return '█'
	}

	// Use top-aligned blocks for inverted display
	invertedBlocks := []rune{' ', '▔', '▔', '▀', '▀', '▀', '█', '█', '█'}
	index := int(fillAmount * float64(len(invertedBlocks)-1))
	if index < 0 {
		index = 0
	}
	if index >= len(invertedBlocks) {
		index = len(invertedBlocks) - 1
	}

	return invertedBlocks[index]
}

// RenderSimple renders a simple single-line bar visualization
func RenderSimple(spectrum []float64, width int) string {
	if width <= 0 || len(spectrum) == 0 {
		return ""
	}

	numBars := width
	if numBars > len(spectrum) {
		numBars = len(spectrum)
	}

	step := float64(len(spectrum)) / float64(numBars)
	var result strings.Builder

	for i := 0; i < numBars; i++ {
		start := int(float64(i) * step)
		end := int(float64(i+1) * step)
		if end > len(spectrum) {
			end = len(spectrum)
		}

		var sum float64
		count := 0
		for j := start; j < end; j++ {
			sum += spectrum[j]
			count++
		}
		value := sum / float64(count)

		// Map to block character
		index := int(value * float64(len(barBlocks)-1))
		if index < 0 {
			index = 0
		}
		if index >= len(barBlocks) {
			index = len(barBlocks) - 1
		}

		color := GetHeightColor(value)
		result.WriteString(BarStyle(color).Render(string(barBlocks[index])))
	}

	return result.String()
}
