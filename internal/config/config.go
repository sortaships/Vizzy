package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// DisplayConfig holds display-related settings
type DisplayConfig struct {
	BarCount    int    `json:"bar_count"`
	RefreshRate int    `json:"refresh_rate"`
	ColorScheme string `json:"color_scheme"`
}

// AudioConfig holds audio capture settings
type AudioConfig struct {
	SampleRate int    `json:"sample_rate"`
	ChunkSize  int    `json:"chunk_size"`
	Device     string `json:"device"`
}

// VisualizationConfig holds visualization mode settings
type VisualizationConfig struct {
	Mode         string `json:"mode"` // "bars" or "waveform"
	ShowPeaks    bool   `json:"show_peaks"`
	MirrorBars   bool   `json:"mirror_bars"`
	UseGradient  bool   `json:"use_gradient"`
	GradientHigh string `json:"gradient_high"`
	GradientLow  string `json:"gradient_low"`
}

// SmoothingConfig holds smoothing parameters
type SmoothingConfig struct {
	AttackSpeed float64 `json:"attack_speed"`
	DecaySpeed  float64 `json:"decay_speed"`
	RestDecay   float64 `json:"rest_decay"`
}

// SensitivityConfig holds sensitivity and boost settings
type SensitivityConfig struct {
	Sensitivity   float64 `json:"sensitivity"`
	NoiseFloor    float64 `json:"noise_floor"`
	BassBoost     float64 `json:"bass_boost"`
	MidBoost      float64 `json:"mid_boost"`
	TrebleBoost   float64 `json:"treble_boost"`
	MinFrequency  float64 `json:"min_frequency"`
	MaxFrequency  float64 `json:"max_frequency"`
}

// Config holds all application configuration
type Config struct {
	Display       DisplayConfig       `json:"display"`
	Audio         AudioConfig         `json:"audio"`
	Visualization VisualizationConfig `json:"visualization"`
	Smoothing     SmoothingConfig     `json:"smoothing"`
	Sensitivity   SensitivityConfig   `json:"sensitivity"`
}

// Default returns a Config with sensible default values
func Default() *Config {
	return &Config{
		Display: DisplayConfig{
			BarCount:    64,
			RefreshRate: 60,
			ColorScheme: "rainbow",
		},
		Audio: AudioConfig{
			SampleRate: 44100,
			ChunkSize:  1024,
			Device:     "",
		},
		Visualization: VisualizationConfig{
			Mode:         "bars",
			ShowPeaks:    true,
			MirrorBars:   false,
			UseGradient:  true,
			GradientHigh: "#ff00ff",
			GradientLow:  "#00ffff",
		},
		Smoothing: SmoothingConfig{
			AttackSpeed: 0.8,
			DecaySpeed:  0.3,
			RestDecay:   0.05,
		},
		Sensitivity: SensitivityConfig{
			Sensitivity:  1.0,
			NoiseFloor:   0.01,
			BassBoost:    0.8,  // Reduced - bass is attenuated by frequency weighting
			MidBoost:     1.0,
			TrebleBoost:  1.0,  // Increased to balance with bass attenuation
			MinFrequency: 50.0, // Increased to cut sub-bass rumble
			MaxFrequency: 18000.0,
		},
	}
}

// Load reads configuration from a JSON file
// If the file doesn't exist, returns default configuration
func Load(path string) (*Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save writes configuration to a JSON file
func (c *Config) Save(path string) error {
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Clone creates a deep copy of the configuration
func (c *Config) Clone() *Config {
	clone := *c
	return &clone
}

// Reset resets the configuration to defaults
func (c *Config) Reset() {
	*c = *Default()
}

// Validate checks if the configuration values are valid
func (c *Config) Validate() error {
	if c.Display.BarCount < 8 {
		c.Display.BarCount = 8
	}
	if c.Display.BarCount > 256 {
		c.Display.BarCount = 256
	}

	if c.Display.RefreshRate < 10 {
		c.Display.RefreshRate = 10
	}
	if c.Display.RefreshRate > 120 {
		c.Display.RefreshRate = 120
	}

	if c.Audio.SampleRate < 8000 {
		c.Audio.SampleRate = 8000
	}
	if c.Audio.SampleRate > 192000 {
		c.Audio.SampleRate = 192000
	}

	if c.Audio.ChunkSize < 256 {
		c.Audio.ChunkSize = 256
	}
	if c.Audio.ChunkSize > 8192 {
		c.Audio.ChunkSize = 8192
	}

	if c.Smoothing.AttackSpeed < 0.0 {
		c.Smoothing.AttackSpeed = 0.0
	}
	if c.Smoothing.AttackSpeed > 1.0 {
		c.Smoothing.AttackSpeed = 1.0
	}

	if c.Smoothing.DecaySpeed < 0.0 {
		c.Smoothing.DecaySpeed = 0.0
	}
	if c.Smoothing.DecaySpeed > 1.0 {
		c.Smoothing.DecaySpeed = 1.0
	}

	if c.Sensitivity.Sensitivity < 0.1 {
		c.Sensitivity.Sensitivity = 0.1
	}
	if c.Sensitivity.Sensitivity > 10.0 {
		c.Sensitivity.Sensitivity = 10.0
	}

	return nil
}
