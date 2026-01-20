package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Braille dot positions:
// 1 4
// 2 5
// 3 6
// 7 8
//
// Braille character = 0x2800 + dot pattern
// Dot 1 = 0x01, Dot 2 = 0x02, Dot 3 = 0x04, Dot 4 = 0x08
// Dot 5 = 0x10, Dot 6 = 0x20, Dot 7 = 0x40, Dot 8 = 0x80

const brailleBase = 0x2800

// Braille dot offsets (rows 0-3 for left column, 4-7 for right column)
var brailleDots = [8]int{0x01, 0x02, 0x04, 0x40, 0x08, 0x10, 0x20, 0x80}

// WaveformRenderer renders spectrum data as a smooth waveform using braille
type WaveformRenderer struct {
	width       int
	height      int
	useGradient bool
	fillBelow   bool
}

// NewWaveformRenderer creates a new waveform renderer
func NewWaveformRenderer(width, height int) *WaveformRenderer {
	return &WaveformRenderer{
		width:       width,
		height:      height,
		useGradient: true,
		fillBelow:   true,
	}
}

// SetSize updates the renderer dimensions
func (r *WaveformRenderer) SetSize(width, height int) {
	r.width = width
	r.height = height
}

// SetOptions updates rendering options
func (r *WaveformRenderer) SetOptions(useGradient, fillBelow bool) {
	r.useGradient = useGradient
	r.fillBelow = fillBelow
}

// Render renders the spectrum as a waveform using braille characters
func (r *WaveformRenderer) Render(spectrum []float64) string {
	if r.width <= 0 || r.height <= 0 || len(spectrum) == 0 {
		return ""
	}

	// Each braille character is 2 dots wide and 4 dots tall
	// So we have width*2 horizontal dots and height*4 vertical dots
	dotsWidth := r.width * 2
	dotsHeight := r.height * 4

	// Resample spectrum to match dots width
	points := resampleSpectrum(spectrum, dotsWidth)

	// Create a 2D grid of dots
	grid := make([][]bool, dotsHeight)
	for i := range grid {
		grid[i] = make([]bool, dotsWidth)
	}

	// Plot the waveform
	for x := 0; x < dotsWidth; x++ {
		// Map value (0-1) to y position (0 to dotsHeight-1)
		value := points[x]
		if value < 0 {
			value = 0
		}
		if value > 1 {
			value = 1
		}

		// y=0 is top, so invert
		y := int((1 - value) * float64(dotsHeight-1))
		if y < 0 {
			y = 0
		}
		if y >= dotsHeight {
			y = dotsHeight - 1
		}

		// Set the point
		grid[y][x] = true

		// Fill below if enabled
		if r.fillBelow {
			for yy := y + 1; yy < dotsHeight; yy++ {
				grid[yy][x] = true
			}
		}
	}

	// Convert grid to braille characters
	var lines []string
	for row := 0; row < r.height; row++ {
		var line strings.Builder
		for col := 0; col < r.width; col++ {
			char := r.gridToBraille(grid, col, row)

			// Determine color based on position or height
			var color lipgloss.Color
			if r.useGradient {
				// Use average height in this cell for color
				avgHeight := r.getCellHeight(points, col)
				color = GetHeightColor(avgHeight)
			} else {
				color = GetBarColor(float64(col)/float64(r.width), false)
			}

			line.WriteString(BarStyle(color).Render(string(char)))
		}
		lines = append(lines, line.String())
	}

	return strings.Join(lines, "\n")
}

// gridToBraille converts a 2x4 section of the grid to a braille character
func (r *WaveformRenderer) gridToBraille(grid [][]bool, col, row int) rune {
	pattern := 0
	baseX := col * 2
	baseY := row * 4

	// Map grid positions to braille dots
	// Left column (dots 1,2,3,7)
	if baseY < len(grid) && baseX < len(grid[0]) && grid[baseY][baseX] {
		pattern |= brailleDots[0] // dot 1
	}
	if baseY+1 < len(grid) && baseX < len(grid[0]) && grid[baseY+1][baseX] {
		pattern |= brailleDots[1] // dot 2
	}
	if baseY+2 < len(grid) && baseX < len(grid[0]) && grid[baseY+2][baseX] {
		pattern |= brailleDots[2] // dot 3
	}
	if baseY+3 < len(grid) && baseX < len(grid[0]) && grid[baseY+3][baseX] {
		pattern |= brailleDots[3] // dot 7
	}

	// Right column (dots 4,5,6,8)
	if baseY < len(grid) && baseX+1 < len(grid[0]) && grid[baseY][baseX+1] {
		pattern |= brailleDots[4] // dot 4
	}
	if baseY+1 < len(grid) && baseX+1 < len(grid[0]) && grid[baseY+1][baseX+1] {
		pattern |= brailleDots[5] // dot 5
	}
	if baseY+2 < len(grid) && baseX+1 < len(grid[0]) && grid[baseY+2][baseX+1] {
		pattern |= brailleDots[6] // dot 6
	}
	if baseY+3 < len(grid) && baseX+1 < len(grid[0]) && grid[baseY+3][baseX+1] {
		pattern |= brailleDots[7] // dot 8
	}

	return rune(brailleBase + pattern)
}

// getCellHeight returns the average height value for a braille cell
func (r *WaveformRenderer) getCellHeight(points []float64, col int) float64 {
	baseX := col * 2
	if baseX >= len(points) {
		return 0
	}

	sum := points[baseX]
	count := 1

	if baseX+1 < len(points) {
		sum += points[baseX+1]
		count++
	}

	return sum / float64(count)
}

// resampleSpectrum resamples the spectrum to the target width
func resampleSpectrum(spectrum []float64, targetWidth int) []float64 {
	if len(spectrum) == 0 || targetWidth <= 0 {
		return make([]float64, targetWidth)
	}

	result := make([]float64, targetWidth)
	step := float64(len(spectrum)) / float64(targetWidth)

	for i := 0; i < targetWidth; i++ {
		// Use linear interpolation for smoother result
		pos := float64(i) * step
		index := int(pos)
		frac := pos - float64(index)

		if index >= len(spectrum)-1 {
			result[i] = spectrum[len(spectrum)-1]
		} else {
			// Linear interpolation
			result[i] = spectrum[index]*(1-frac) + spectrum[index+1]*frac
		}
	}

	return result
}

// RenderWaveformSimple renders a simple single-line waveform using block characters
func RenderWaveformSimple(spectrum []float64, width int) string {
	if width <= 0 || len(spectrum) == 0 {
		return ""
	}

	// Resample spectrum
	points := resampleSpectrum(spectrum, width)

	// Use half-block characters for basic waveform
	var result strings.Builder
	for i, value := range points {
		char := valueToBlock(value)
		color := GetHeightColor(value)
		result.WriteString(BarStyle(color).Render(string(char)))
		_ = i // suppress unused variable warning
	}

	return result.String()
}

// valueToBlock converts a 0-1 value to a block character
func valueToBlock(value float64) rune {
	if value < 0.125 {
		return ' '
	} else if value < 0.25 {
		return '▁'
	} else if value < 0.375 {
		return '▂'
	} else if value < 0.5 {
		return '▃'
	} else if value < 0.625 {
		return '▄'
	} else if value < 0.75 {
		return '▅'
	} else if value < 0.875 {
		return '▆'
	} else if value < 1.0 {
		return '▇'
	}
	return '█'
}
