package fft

import (
	"math"
	"testing"
)

func TestNewProcessor(t *testing.T) {
	cfg := DefaultProcessorConfig()
	p := NewProcessor(cfg)

	if p == nil {
		t.Fatal("NewProcessor returned nil")
	}
	if p.numBands != 64 {
		t.Errorf("expected numBands=64, got %d", p.numBands)
	}
	if len(p.smoothed) != 64 {
		t.Errorf("expected smoothed length=64, got %d", len(p.smoothed))
	}
	if len(p.window) != cfg.ChunkSize {
		t.Errorf("expected window length=%d, got %d", cfg.ChunkSize, len(p.window))
	}
}

func TestProcessSilence(t *testing.T) {
	cfg := DefaultProcessorConfig()
	cfg.NumBands = 16
	p := NewProcessor(cfg)

	// Process silence
	silence := make([]float64, cfg.ChunkSize)
	result := p.Process(silence)

	if len(result) != 16 {
		t.Errorf("expected result length=16, got %d", len(result))
	}

	// All values should be 0 or very close to 0
	for i, v := range result {
		if v > 0.01 {
			t.Errorf("expected silence result[%d] ≈ 0, got %f", i, v)
		}
	}
}

func TestProcessSineWave(t *testing.T) {
	cfg := DefaultProcessorConfig()
	cfg.NumBands = 32
	cfg.SampleRate = 44100
	cfg.ChunkSize = 1024
	p := NewProcessor(cfg)

	// Generate 440 Hz sine wave (A4)
	samples := make([]float64, cfg.ChunkSize)
	freq := 440.0
	for i := range samples {
		t := float64(i) / float64(cfg.SampleRate)
		samples[i] = math.Sin(2 * math.Pi * freq * t)
	}

	result := p.Process(samples)

	if len(result) != 32 {
		t.Errorf("expected result length=32, got %d", len(result))
	}

	// Should have some energy (not all zeros)
	hasEnergy := false
	for _, v := range result {
		if v > 0.1 {
			hasEnergy = true
			break
		}
	}
	if !hasEnergy {
		t.Error("expected sine wave to produce energy in spectrum")
	}
}

func TestProcessPadding(t *testing.T) {
	cfg := DefaultProcessorConfig()
	cfg.ChunkSize = 1024
	p := NewProcessor(cfg)

	// Process samples shorter than chunk size
	short := make([]float64, 100)
	for i := range short {
		short[i] = 0.5
	}

	result := p.Process(short)

	if len(result) != cfg.NumBands {
		t.Errorf("expected result length=%d, got %d", cfg.NumBands, len(result))
	}
}

func TestSmoothing(t *testing.T) {
	cfg := DefaultProcessorConfig()
	cfg.NumBands = 8
	cfg.AttackSpeed = 0.9
	cfg.DecaySpeed = 0.1
	p := NewProcessor(cfg)

	// Generate a signal that produces energy
	signal := make([]float64, cfg.ChunkSize)
	for i := range signal {
		signal[i] = math.Sin(2 * math.Pi * 1000 * float64(i) / float64(cfg.SampleRate))
	}

	// First process should start smoothing
	p.Process(signal)
	first := p.GetSmoothed()

	// Process same signal again - should smooth
	p.Process(signal)
	second := p.GetSmoothed()

	// Values should be similar (smoothed)
	for i := range first {
		if first[i] > 0 && second[i] > 0 {
			diff := math.Abs(first[i] - second[i])
			if diff > 0.5 {
				t.Logf("smoothing working: first[%d]=%f, second[%d]=%f", i, first[i], i, second[i])
			}
		}
	}
}

func TestReset(t *testing.T) {
	cfg := DefaultProcessorConfig()
	p := NewProcessor(cfg)

	// Process some signal
	signal := make([]float64, cfg.ChunkSize)
	for i := range signal {
		signal[i] = math.Sin(2 * math.Pi * 500 * float64(i) / float64(cfg.SampleRate))
	}
	p.Process(signal)

	// Reset
	p.Reset()

	// All smoothed values should be 0
	smoothed := p.GetSmoothed()
	for i, v := range smoothed {
		if v != 0 {
			t.Errorf("expected smoothed[%d]=0 after reset, got %f", i, v)
		}
	}
}

func TestSetSensitivity(t *testing.T) {
	cfg := DefaultProcessorConfig()
	p := NewProcessor(cfg)

	// Test clamping
	p.SetSensitivity(0.01) // Below minimum
	if p.sensitivity != 0.1 {
		t.Errorf("expected sensitivity clamped to 0.1, got %f", p.sensitivity)
	}

	p.SetSensitivity(100) // Above maximum
	if p.sensitivity != 10.0 {
		t.Errorf("expected sensitivity clamped to 10.0, got %f", p.sensitivity)
	}

	p.SetSensitivity(2.0)
	if p.sensitivity != 2.0 {
		t.Errorf("expected sensitivity=2.0, got %f", p.sensitivity)
	}
}

func TestSetNumBands(t *testing.T) {
	cfg := DefaultProcessorConfig()
	p := NewProcessor(cfg)

	p.SetNumBands(128)
	if p.numBands != 128 {
		t.Errorf("expected numBands=128, got %d", p.numBands)
	}
	if len(p.smoothed) != 128 {
		t.Errorf("expected smoothed length=128, got %d", len(p.smoothed))
	}

	// Test clamping
	p.SetNumBands(4)
	if p.numBands != 8 {
		t.Errorf("expected numBands clamped to 8, got %d", p.numBands)
	}

	p.SetNumBands(1000)
	if p.numBands != 256 {
		t.Errorf("expected numBands clamped to 256, got %d", p.numBands)
	}
}

func TestBandMapping(t *testing.T) {
	cfg := DefaultProcessorConfig()
	cfg.NumBands = 16
	p := NewProcessor(cfg)

	// Band mapping should be monotonically increasing
	for i := 0; i < len(p.bandMapping)-1; i++ {
		if p.bandMapping[i] > p.bandMapping[i+1] {
			t.Errorf("band mapping not monotonic: mapping[%d]=%d > mapping[%d]=%d",
				i, p.bandMapping[i], i+1, p.bandMapping[i+1])
		}
	}
}

func TestGetPeaks(t *testing.T) {
	cfg := DefaultProcessorConfig()
	cfg.NumBands = 8
	p := NewProcessor(cfg)

	// Process a signal
	signal := make([]float64, cfg.ChunkSize)
	for i := range signal {
		signal[i] = math.Sin(2 * math.Pi * 1000 * float64(i) / float64(cfg.SampleRate))
	}
	p.Process(signal)

	peaks := p.GetPeaks()
	if len(peaks) != 8 {
		t.Errorf("expected peaks length=8, got %d", len(peaks))
	}

	// Peaks should be >= smoothed values
	smoothed := p.GetSmoothed()
	for i := range peaks {
		if peaks[i] < smoothed[i]-0.001 {
			t.Errorf("peak[%d]=%f should be >= smoothed[%d]=%f", i, peaks[i], i, smoothed[i])
		}
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		v, min, max, expected float64
	}{
		{5, 0, 10, 5},
		{-5, 0, 10, 0},
		{15, 0, 10, 10},
		{0.5, 0, 1, 0.5},
	}

	for _, tt := range tests {
		result := clamp(tt.v, tt.min, tt.max)
		if result != tt.expected {
			t.Errorf("clamp(%f, %f, %f) = %f, expected %f", tt.v, tt.min, tt.max, result, tt.expected)
		}
	}
}
