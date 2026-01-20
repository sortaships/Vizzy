package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.Display.BarCount != 64 {
		t.Errorf("expected BarCount=64, got %d", cfg.Display.BarCount)
	}
	if cfg.Audio.SampleRate != 44100 {
		t.Errorf("expected SampleRate=44100, got %d", cfg.Audio.SampleRate)
	}
	if cfg.Visualization.Mode != "bars" {
		t.Errorf("expected Mode=bars, got %s", cfg.Visualization.Mode)
	}
	if cfg.Smoothing.AttackSpeed != 0.8 {
		t.Errorf("expected AttackSpeed=0.8, got %f", cfg.Smoothing.AttackSpeed)
	}
}

func TestLoadNonExistent(t *testing.T) {
	cfg, err := Load("nonexistent.json")
	if err != nil {
		t.Fatalf("Load should not error on nonexistent file: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load should return default config for nonexistent file")
	}
	if cfg.Display.BarCount != 64 {
		t.Errorf("expected default BarCount=64, got %d", cfg.Display.BarCount)
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test_config.json")

	cfg := Default()
	cfg.Display.BarCount = 128
	cfg.Audio.SampleRate = 48000
	cfg.Visualization.Mode = "waveform"

	if err := cfg.Save(path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Display.BarCount != 128 {
		t.Errorf("expected BarCount=128, got %d", loaded.Display.BarCount)
	}
	if loaded.Audio.SampleRate != 48000 {
		t.Errorf("expected SampleRate=48000, got %d", loaded.Audio.SampleRate)
	}
	if loaded.Visualization.Mode != "waveform" {
		t.Errorf("expected Mode=waveform, got %s", loaded.Visualization.Mode)
	}
}

func TestSaveCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "subdir", "nested", "config.json")

	cfg := Default()
	if err := cfg.Save(path); err != nil {
		t.Fatalf("Save should create directories: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("config file should exist after save")
	}
}

func TestClone(t *testing.T) {
	cfg := Default()
	cfg.Display.BarCount = 32

	clone := cfg.Clone()
	clone.Display.BarCount = 64

	if cfg.Display.BarCount != 32 {
		t.Error("Clone should not modify original")
	}
	if clone.Display.BarCount != 64 {
		t.Error("Clone should have modified value")
	}
}

func TestReset(t *testing.T) {
	cfg := Default()
	cfg.Display.BarCount = 256
	cfg.Audio.SampleRate = 96000

	cfg.Reset()

	if cfg.Display.BarCount != 64 {
		t.Errorf("Reset should restore default BarCount=64, got %d", cfg.Display.BarCount)
	}
	if cfg.Audio.SampleRate != 44100 {
		t.Errorf("Reset should restore default SampleRate=44100, got %d", cfg.Audio.SampleRate)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name     string
		modify   func(*Config)
		expected func(*Config) bool
	}{
		{
			name:   "BarCount too low",
			modify: func(c *Config) { c.Display.BarCount = 2 },
			expected: func(c *Config) bool {
				return c.Display.BarCount == 8
			},
		},
		{
			name:   "BarCount too high",
			modify: func(c *Config) { c.Display.BarCount = 500 },
			expected: func(c *Config) bool {
				return c.Display.BarCount == 256
			},
		},
		{
			name:   "SampleRate too low",
			modify: func(c *Config) { c.Audio.SampleRate = 100 },
			expected: func(c *Config) bool {
				return c.Audio.SampleRate == 8000
			},
		},
		{
			name:   "AttackSpeed negative",
			modify: func(c *Config) { c.Smoothing.AttackSpeed = -0.5 },
			expected: func(c *Config) bool {
				return c.Smoothing.AttackSpeed == 0.0
			},
		},
		{
			name:   "Sensitivity too low",
			modify: func(c *Config) { c.Sensitivity.Sensitivity = 0.01 },
			expected: func(c *Config) bool {
				return c.Sensitivity.Sensitivity == 0.1
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Default()
			tt.modify(cfg)
			cfg.Validate()
			if !tt.expected(cfg) {
				t.Error("Validate did not correct the value as expected")
			}
		})
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "invalid.json")

	if err := os.WriteFile(path, []byte("not valid json"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Error("Load should error on invalid JSON")
	}
}
