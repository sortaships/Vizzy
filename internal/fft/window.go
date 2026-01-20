package fft

import "math"

// WindowFunc represents a window function type
type WindowFunc int

const (
	WindowNone WindowFunc = iota
	WindowHanning
	WindowHamming
	WindowBlackman
)

// ApplyWindow applies a window function to the input samples
func ApplyWindow(samples []float64, windowType WindowFunc) []float64 {
	n := len(samples)
	result := make([]float64, n)

	switch windowType {
	case WindowHanning:
		for i := 0; i < n; i++ {
			w := 0.5 * (1 - math.Cos(2*math.Pi*float64(i)/float64(n-1)))
			result[i] = samples[i] * w
		}
	case WindowHamming:
		for i := 0; i < n; i++ {
			w := 0.54 - 0.46*math.Cos(2*math.Pi*float64(i)/float64(n-1))
			result[i] = samples[i] * w
		}
	case WindowBlackman:
		for i := 0; i < n; i++ {
			w := 0.42 - 0.5*math.Cos(2*math.Pi*float64(i)/float64(n-1)) +
				0.08*math.Cos(4*math.Pi*float64(i)/float64(n-1))
			result[i] = samples[i] * w
		}
	default:
		copy(result, samples)
	}

	return result
}

// GenerateHanningWindow generates a Hanning window of the specified size
func GenerateHanningWindow(size int) []float64 {
	window := make([]float64, size)
	for i := 0; i < size; i++ {
		window[i] = 0.5 * (1 - math.Cos(2*math.Pi*float64(i)/float64(size-1)))
	}
	return window
}

// GenerateHammingWindow generates a Hamming window of the specified size
func GenerateHammingWindow(size int) []float64 {
	window := make([]float64, size)
	for i := 0; i < size; i++ {
		window[i] = 0.54 - 0.46*math.Cos(2*math.Pi*float64(i)/float64(size-1))
	}
	return window
}

// ApplyPrecomputedWindow applies a precomputed window to samples
func ApplyPrecomputedWindow(samples []float64, window []float64) []float64 {
	n := len(samples)
	if len(window) < n {
		n = len(window)
	}
	result := make([]float64, n)
	for i := 0; i < n; i++ {
		result[i] = samples[i] * window[i]
	}
	return result
}
