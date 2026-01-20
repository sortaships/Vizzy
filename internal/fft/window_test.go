package fft

import (
	"math"
	"testing"
)

func TestHanningWindow(t *testing.T) {
	window := GenerateHanningWindow(100)

	if len(window) != 100 {
		t.Errorf("expected window length 100, got %d", len(window))
	}

	// Hanning window should be 0 at endpoints
	if math.Abs(window[0]) > 1e-10 {
		t.Errorf("expected window[0] ≈ 0, got %f", window[0])
	}
	if math.Abs(window[99]) > 1e-10 {
		t.Errorf("expected window[99] ≈ 0, got %f", window[99])
	}

	// Hanning window should be 1 at center
	center := window[49]
	if math.Abs(center-1.0) > 0.02 {
		t.Errorf("expected window center ≈ 1, got %f", center)
	}

	// Window should be symmetric
	for i := 0; i < 50; i++ {
		if math.Abs(window[i]-window[99-i]) > 1e-10 {
			t.Errorf("window not symmetric at index %d: %f vs %f", i, window[i], window[99-i])
		}
	}
}

func TestHammingWindow(t *testing.T) {
	window := GenerateHammingWindow(100)

	if len(window) != 100 {
		t.Errorf("expected window length 100, got %d", len(window))
	}

	// Hamming window should be ~0.08 at endpoints (not exactly 0)
	if math.Abs(window[0]-0.08) > 0.01 {
		t.Errorf("expected window[0] ≈ 0.08, got %f", window[0])
	}

	// Window should be symmetric
	for i := 0; i < 50; i++ {
		if math.Abs(window[i]-window[99-i]) > 1e-10 {
			t.Errorf("window not symmetric at index %d", i)
		}
	}
}

func TestApplyWindow(t *testing.T) {
	samples := make([]float64, 100)
	for i := range samples {
		samples[i] = 1.0
	}

	// Test Hanning
	windowed := ApplyWindow(samples, WindowHanning)
	if len(windowed) != 100 {
		t.Errorf("expected length 100, got %d", len(windowed))
	}

	// Endpoints should be 0 after Hanning
	if math.Abs(windowed[0]) > 1e-10 {
		t.Errorf("expected windowed[0] ≈ 0, got %f", windowed[0])
	}

	// Test no window (should copy)
	noWindow := ApplyWindow(samples, WindowNone)
	for i := range noWindow {
		if noWindow[i] != 1.0 {
			t.Errorf("expected noWindow[%d] = 1.0, got %f", i, noWindow[i])
		}
	}
}

func TestApplyPrecomputedWindow(t *testing.T) {
	samples := []float64{1.0, 1.0, 1.0, 1.0}
	window := []float64{0.5, 1.0, 1.0, 0.5}

	result := ApplyPrecomputedWindow(samples, window)

	expected := []float64{0.5, 1.0, 1.0, 0.5}
	for i := range expected {
		if math.Abs(result[i]-expected[i]) > 1e-10 {
			t.Errorf("result[%d] = %f, expected %f", i, result[i], expected[i])
		}
	}
}

func TestWindowDoesNotModifyInput(t *testing.T) {
	samples := []float64{1.0, 2.0, 3.0, 4.0}
	original := make([]float64, len(samples))
	copy(original, samples)

	ApplyWindow(samples, WindowHanning)

	for i := range samples {
		if samples[i] != original[i] {
			t.Error("ApplyWindow modified input samples")
		}
	}
}
