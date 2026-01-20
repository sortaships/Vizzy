package fft

import (
	"math"
	"math/cmplx"

	"github.com/mjibson/go-dsp/fft"
)

// Processor performs FFT analysis with smoothing
type Processor struct {
	// FFT settings
	sampleRate  int
	chunkSize   int
	window      []float64
	numBands    int
	bandMapping []int // Maps FFT bins to output bands

	// Smoothing
	attackSpeed float64
	decaySpeed  float64
	restDecay   float64
	smoothed    []float64
	peaks       []float64
	peakDecay   float64

	// Frequency boost
	bassBoost   float64
	midBoost    float64
	trebleBoost float64

	// Frequency range
	minFreq float64
	maxFreq float64

	// Other settings
	noiseFloor  float64
	sensitivity float64
}

// ProcessorConfig holds configuration for the processor
type ProcessorConfig struct {
	SampleRate  int
	ChunkSize   int
	NumBands    int
	AttackSpeed float64
	DecaySpeed  float64
	RestDecay   float64
	BassBoost   float64
	MidBoost    float64
	TrebleBoost float64
	MinFreq     float64
	MaxFreq     float64
	NoiseFloor  float64
	Sensitivity float64
}

// DefaultProcessorConfig returns sensible defaults
func DefaultProcessorConfig() ProcessorConfig {
	return ProcessorConfig{
		SampleRate:  44100,
		ChunkSize:   1024,
		NumBands:    64,
		AttackSpeed: 0.8,
		DecaySpeed:  0.3,
		RestDecay:   0.05,
		BassBoost:   0.8,  // Reduced from 1.2 - bass is already attenuated by frequency weighting
		MidBoost:    1.0,
		TrebleBoost: 1.0,  // Increased from 0.8 - helps balance with bass attenuation
		MinFreq:     50.0, // Increased from 20.0 - cuts sub-bass rumble
		MaxFreq:     18000.0,
		NoiseFloor:  0.01,
		Sensitivity: 1.0,
	}
}

// NewProcessor creates a new FFT processor
func NewProcessor(cfg ProcessorConfig) *Processor {
	p := &Processor{
		sampleRate:  cfg.SampleRate,
		chunkSize:   cfg.ChunkSize,
		numBands:    cfg.NumBands,
		attackSpeed: cfg.AttackSpeed,
		decaySpeed:  cfg.DecaySpeed,
		restDecay:   cfg.RestDecay,
		bassBoost:   cfg.BassBoost,
		midBoost:    cfg.MidBoost,
		trebleBoost: cfg.TrebleBoost,
		minFreq:     cfg.MinFreq,
		maxFreq:     cfg.MaxFreq,
		noiseFloor:  cfg.NoiseFloor,
		sensitivity: cfg.Sensitivity,
		peakDecay:   0.02,
	}

	// Generate window function
	p.window = GenerateHanningWindow(cfg.ChunkSize)

	// Initialize smoothed values
	p.smoothed = make([]float64, cfg.NumBands)
	p.peaks = make([]float64, cfg.NumBands)

	// Compute logarithmic frequency band mapping
	p.computeBandMapping()

	return p
}

// computeBandMapping creates a logarithmic mapping from FFT bins to output bands
func (p *Processor) computeBandMapping() {
	// Number of useful FFT bins (Nyquist)
	numBins := p.chunkSize / 2

	// Frequency resolution
	freqResolution := float64(p.sampleRate) / float64(p.chunkSize)

	// Min and max bins based on frequency range
	minBin := int(p.minFreq / freqResolution)
	maxBin := int(p.maxFreq / freqResolution)
	if minBin < 1 {
		minBin = 1
	}
	if maxBin >= numBins {
		maxBin = numBins - 1
	}

	// Create logarithmic band boundaries
	p.bandMapping = make([]int, p.numBands+1)
	logMin := math.Log10(float64(minBin))
	logMax := math.Log10(float64(maxBin))
	logRange := logMax - logMin

	for i := 0; i <= p.numBands; i++ {
		logBin := logMin + logRange*float64(i)/float64(p.numBands)
		p.bandMapping[i] = int(math.Pow(10, logBin))
	}
}

// Process takes audio samples and returns the spectrum
func (p *Processor) Process(samples []float64) []float64 {
	if len(samples) < p.chunkSize {
		// Pad with zeros if needed
		padded := make([]float64, p.chunkSize)
		copy(padded, samples)
		samples = padded
	} else if len(samples) > p.chunkSize {
		samples = samples[:p.chunkSize]
	}

	// Apply window function
	windowed := ApplyPrecomputedWindow(samples, p.window)

	// Perform FFT
	spectrum := fft.FFTReal(windowed)

	// Compute magnitude for each band
	bands := make([]float64, p.numBands)
	numBins := p.chunkSize / 2

	for band := 0; band < p.numBands; band++ {
		startBin := p.bandMapping[band]
		endBin := p.bandMapping[band+1]

		if startBin >= numBins {
			startBin = numBins - 1
		}
		if endBin >= numBins {
			endBin = numBins - 1
		}
		if endBin <= startBin {
			endBin = startBin + 1
		}

		// Average magnitude across bins in this band
		var sum float64
		count := 0
		for bin := startBin; bin < endBin && bin < numBins; bin++ {
			mag := cmplx.Abs(spectrum[bin])
			sum += mag
			count++
		}
		if count > 0 {
			bands[band] = sum / float64(count)
		}
	}

	// Apply frequency-dependent weighting to compensate for natural bass energy
	// This is similar to A-weighting used in audio measurements
	// Aggressive attenuation of low frequencies to prevent bass domination
	for i := 0; i < p.numBands; i++ {
		bandPos := float64(i) / float64(p.numBands)

		// Apply a frequency weighting curve - heavily attenuate bass
		// This compensates for the natural loudness curve of audio
		var weight float64
		if bandPos < 0.08 {
			// Sub-bass: very strong attenuation (mostly rumble)
			weight = 0.1 + bandPos*1.5 // 0.1 to 0.22
		} else if bandPos < 0.20 {
			// Low bass: strong attenuation
			weight = 0.22 + (bandPos-0.08)*2.5 // 0.22 to 0.52
		} else if bandPos < 0.35 {
			// Upper bass: moderate attenuation
			weight = 0.52 + (bandPos-0.20)*2.5 // 0.52 to 0.90
		} else if bandPos < 0.55 {
			// Low-mid: slight attenuation
			weight = 0.90 + (bandPos-0.35)*0.5 // 0.90 to 1.0
		} else {
			// Upper-mid and treble: full weight
			weight = 1.0
		}
		bands[i] *= weight
	}

	// Apply user-configurable frequency boost on top of weighting
	for i := 0; i < p.numBands; i++ {
		bandPos := float64(i) / float64(p.numBands)
		var boost float64
		if bandPos < 0.33 {
			// Bass
			boost = p.bassBoost
		} else if bandPos < 0.66 {
			// Mid
			boost = p.midBoost
		} else {
			// Treble
			boost = p.trebleBoost
		}
		bands[i] *= boost
	}

	// Apply sensitivity
	for i := 0; i < p.numBands; i++ {
		bands[i] *= p.sensitivity
	}

	// Apply noise floor
	for i := 0; i < p.numBands; i++ {
		if bands[i] < p.noiseFloor {
			bands[i] = 0
		}
	}

	// Normalize
	maxVal := 0.0
	for _, v := range bands {
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal > 0 {
		normFactor := 1.0 / maxVal
		// Use a softer normalization to avoid excessive scaling
		if normFactor > 10 {
			normFactor = 10
		}
		for i := range bands {
			bands[i] *= normFactor
			if bands[i] > 1.0 {
				bands[i] = 1.0
			}
		}
	}

	// Apply smoothing
	p.applySmoothing(bands)

	return p.smoothed
}

// applySmoothing applies asymmetric smoothing to the bands
func (p *Processor) applySmoothing(bands []float64) {
	for i := 0; i < p.numBands; i++ {
		target := bands[i]
		current := p.smoothed[i]

		if target > current {
			// Attack (rising)
			p.smoothed[i] = current + (target-current)*p.attackSpeed
		} else if target > 0.05 {
			// Decay (falling but still active)
			p.smoothed[i] = current + (target-current)*p.decaySpeed
		} else {
			// Rest decay (return to zero)
			p.smoothed[i] = current * (1 - p.restDecay)
		}

		// Clamp to valid range
		if p.smoothed[i] < 0 {
			p.smoothed[i] = 0
		}
		if p.smoothed[i] > 1 {
			p.smoothed[i] = 1
		}

		// Update peaks
		if p.smoothed[i] > p.peaks[i] {
			p.peaks[i] = p.smoothed[i]
		} else {
			p.peaks[i] *= (1 - p.peakDecay)
		}
	}
}

// GetSmoothed returns the current smoothed spectrum
func (p *Processor) GetSmoothed() []float64 {
	result := make([]float64, len(p.smoothed))
	copy(result, p.smoothed)
	return result
}

// GetPeaks returns the current peak values
func (p *Processor) GetPeaks() []float64 {
	result := make([]float64, len(p.peaks))
	copy(result, p.peaks)
	return result
}

// Reset clears the smoothed values and peaks
func (p *Processor) Reset() {
	for i := range p.smoothed {
		p.smoothed[i] = 0
	}
	for i := range p.peaks {
		p.peaks[i] = 0
	}
}

// SetSensitivity updates the sensitivity value
func (p *Processor) SetSensitivity(s float64) {
	if s < 0.1 {
		s = 0.1
	}
	if s > 10.0 {
		s = 10.0
	}
	p.sensitivity = s
}

// SetSmoothing updates the smoothing values
func (p *Processor) SetSmoothing(attack, decay, rest float64) {
	p.attackSpeed = clamp(attack, 0, 1)
	p.decaySpeed = clamp(decay, 0, 1)
	p.restDecay = clamp(rest, 0, 1)
}

// SetNumBands changes the number of output bands
func (p *Processor) SetNumBands(n int) {
	if n < 8 {
		n = 8
	}
	if n > 256 {
		n = 256
	}
	p.numBands = n
	p.smoothed = make([]float64, n)
	p.peaks = make([]float64, n)
	p.computeBandMapping()
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
