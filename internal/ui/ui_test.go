package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"go_audio_visualizer/internal/config"
)

func TestVisualizationModeString(t *testing.T) {
	tests := []struct {
		mode     VisualizationMode
		expected string
	}{
		{ModeBars, "Bars"},
		{ModeWaveform, "Waveform"},
		{VisualizationMode(99), "Unknown"},
	}

	for _, tt := range tests {
		result := tt.mode.String()
		if result != tt.expected {
			t.Errorf("VisualizationMode(%d).String() = %q, expected %q", tt.mode, result, tt.expected)
		}
	}
}

func TestNewModel(t *testing.T) {
	cfg := config.Default()
	m := NewModel(cfg, "test_config.json")

	if m == nil {
		t.Fatal("NewModel returned nil")
	}
	if m.config != cfg {
		t.Error("Model config not set correctly")
	}
	if m.mode != ModeBars {
		t.Errorf("expected default mode=ModeBars, got %v", m.mode)
	}
	if len(m.spectrum) != cfg.Display.BarCount {
		t.Errorf("expected spectrum length=%d, got %d", cfg.Display.BarCount, len(m.spectrum))
	}
}

func TestNewModelWithWaveformMode(t *testing.T) {
	cfg := config.Default()
	cfg.Visualization.Mode = "waveform"
	m := NewModel(cfg, "test_config.json")

	if m.mode != ModeWaveform {
		t.Errorf("expected mode=ModeWaveform, got %v", m.mode)
	}
}

func TestHandleKeyPress(t *testing.T) {
	tests := []struct {
		key      string
		expected Action
	}{
		{"q", ActionQuit},
		{"esc", ActionQuit},
		{"ctrl+c", ActionQuit},
		{"v", ActionToggleMode},
		{"d", ActionCycleDevice},
		{"+", ActionIncreaseSensitivity},
		{"=", ActionIncreaseSensitivity},
		{"-", ActionDecreaseSensitivity},
		{"[", ActionDecreaseSmoothing},
		{"]", ActionIncreaseSmoothing},
		{"p", ActionTogglePeaks},
		{"g", ActionToggleGradient},
		{"m", ActionToggleMirror},
		{"h", ActionToggleHelp},
		{"?", ActionToggleHelp},
		{"s", ActionSaveConfig},
		{"r", ActionReset},
		{"x", ActionNone},
	}

	for _, tt := range tests {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
		if tt.key == "esc" {
			msg = tea.KeyMsg{Type: tea.KeyEscape}
		} else if tt.key == "ctrl+c" {
			msg = tea.KeyMsg{Type: tea.KeyCtrlC}
		}

		result := HandleKeyPress(msg)
		if result != tt.expected {
			t.Errorf("HandleKeyPress(%q) = %v, expected %v", tt.key, result, tt.expected)
		}
	}
}

func TestActionString(t *testing.T) {
	tests := []struct {
		action   Action
		expected string
	}{
		{ActionQuit, "Quit"},
		{ActionToggleMode, "Toggle Mode"},
		{ActionNone, "None"},
	}

	for _, tt := range tests {
		result := tt.action.String()
		if result != tt.expected {
			t.Errorf("Action(%d).String() = %q, expected %q", tt.action, result, tt.expected)
		}
	}
}

func TestBarRenderer(t *testing.T) {
	r := NewBarRenderer(80, 10)

	if r == nil {
		t.Fatal("NewBarRenderer returned nil")
	}
	if r.width != 80 {
		t.Errorf("expected width=80, got %d", r.width)
	}
	if r.height != 10 {
		t.Errorf("expected height=10, got %d", r.height)
	}
}

func TestBarRendererRenderEmpty(t *testing.T) {
	r := NewBarRenderer(80, 10)
	result := r.Render([]float64{}, nil)
	if result != "" {
		t.Errorf("expected empty result for empty spectrum, got %q", result)
	}
}

func TestBarRendererRender(t *testing.T) {
	r := NewBarRenderer(20, 5)

	// Create a simple spectrum
	spectrum := make([]float64, 20)
	for i := range spectrum {
		spectrum[i] = float64(i) / 20.0
	}

	result := r.Render(spectrum, nil)

	if result == "" {
		t.Error("expected non-empty result")
	}

	lines := strings.Split(result, "\n")
	if len(lines) != 5 {
		t.Errorf("expected 5 lines, got %d", len(lines))
	}
}

func TestWaveformRenderer(t *testing.T) {
	r := NewWaveformRenderer(40, 8)

	if r == nil {
		t.Fatal("NewWaveformRenderer returned nil")
	}
	if r.width != 40 {
		t.Errorf("expected width=40, got %d", r.width)
	}
	if r.height != 8 {
		t.Errorf("expected height=8, got %d", r.height)
	}
}

func TestWaveformRendererRenderEmpty(t *testing.T) {
	r := NewWaveformRenderer(40, 8)
	result := r.Render([]float64{})
	if result != "" {
		t.Errorf("expected empty result for empty spectrum, got %q", result)
	}
}

func TestWaveformRendererRender(t *testing.T) {
	r := NewWaveformRenderer(20, 4)

	// Create a simple spectrum (sine-like shape)
	spectrum := make([]float64, 40)
	for i := range spectrum {
		spectrum[i] = 0.5 + 0.5*float64(i%10)/10.0
	}

	result := r.Render(spectrum)

	if result == "" {
		t.Error("expected non-empty result")
	}

	lines := strings.Split(result, "\n")
	if len(lines) != 4 {
		t.Errorf("expected 4 lines, got %d", len(lines))
	}
}

func TestResampleSpectrum(t *testing.T) {
	input := []float64{0, 1, 2, 3}

	// Upsample
	result := resampleSpectrum(input, 8)
	if len(result) != 8 {
		t.Errorf("expected length=8, got %d", len(result))
	}

	// Downsample
	result = resampleSpectrum(input, 2)
	if len(result) != 2 {
		t.Errorf("expected length=2, got %d", len(result))
	}

	// Empty input
	result = resampleSpectrum([]float64{}, 4)
	if len(result) != 4 {
		t.Errorf("expected length=4 for empty input, got %d", len(result))
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello world", 8, "hello..."},
		{"hi", 2, "hi"},
		{"hello", 3, "hel"},
	}

	for _, tt := range tests {
		result := truncateString(tt.input, tt.maxLen)
		if result != tt.expected {
			t.Errorf("truncateString(%q, %d) = %q, expected %q", tt.input, tt.maxLen, result, tt.expected)
		}
	}
}

func TestRenderSimple(t *testing.T) {
	spectrum := []float64{0.1, 0.5, 0.9, 0.3}
	result := RenderSimple(spectrum, 4)

	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestRenderWaveformSimple(t *testing.T) {
	spectrum := []float64{0.1, 0.5, 0.9, 0.3}
	result := RenderWaveformSimple(spectrum, 4)

	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestValueToBlock(t *testing.T) {
	tests := []struct {
		value    float64
		expected rune
	}{
		{0.0, ' '},
		{0.1, ' '},
		{0.2, '▁'},
		{0.3, '▂'},
		{0.4, '▃'},
		{0.55, '▄'},
		{0.65, '▅'},
		{0.8, '▆'},
		{0.9, '▇'},
		{1.0, '█'},
	}

	for _, tt := range tests {
		result := valueToBlock(tt.value)
		if result != tt.expected {
			t.Errorf("valueToBlock(%f) = %c, expected %c", tt.value, result, tt.expected)
		}
	}
}

func TestGetBarColor(t *testing.T) {
	// Test gradient
	color1 := GetBarColor(0.0, true)
	color2 := GetBarColor(0.5, true)
	color3 := GetBarColor(1.0, true)

	if color1 == "" || color2 == "" || color3 == "" {
		t.Error("GetBarColor should return non-empty colors")
	}

	// Test rainbow
	color4 := GetBarColor(0.0, false)
	color5 := GetBarColor(0.5, false)

	if color4 == "" || color5 == "" {
		t.Error("GetBarColor should return non-empty colors for rainbow")
	}
}
