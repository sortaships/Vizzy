package audio

import (
	"encoding/binary"
	"fmt"
	"math"
	"runtime"
	"sync"

	"github.com/gen2brain/malgo"
)

// CaptureConfig holds audio capture configuration
type CaptureConfig struct {
	SampleRate int
	ChunkSize  int
	DeviceName string // Empty for default device
	Loopback   bool   // If true, capture from playback device (system audio)
}

// DefaultCaptureConfig returns sensible defaults
func DefaultCaptureConfig() CaptureConfig {
	return CaptureConfig{
		SampleRate: 44100,
		ChunkSize:  1024,
		DeviceName: "",
	}
}

// Capture manages audio capture from an input device
type Capture struct {
	config     CaptureConfig
	context    *malgo.AllocatedContext
	device     *malgo.Device
	deviceName string
	dataChan   chan []float64
	stopChan   chan struct{}
	running    bool
	mu         sync.Mutex
	demoMode   bool
	isLoopback bool
}

// NewCapture creates a new audio capture instance
func NewCapture(cfg CaptureConfig) (*Capture, error) {
	return &Capture{
		config:   cfg,
		dataChan: make(chan []float64, 4),
		stopChan: make(chan struct{}),
		demoMode: false,
	}, nil
}

// Start begins audio capture
func (c *Capture) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return nil
	}

	// Initialize context
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return fmt.Errorf("failed to initialize audio context: %w", err)
	}
	c.context = ctx

	// Determine device type based on loopback setting
	deviceType := malgo.Capture
	if c.config.Loopback {
		// On macOS, loopback capture is not natively supported by miniaudio
		// Users need to install a virtual audio device like BlackHole or Soundflower
		if runtime.GOOS == "darwin" {
			// Try to use loopback anyway - it may work with virtual audio devices
			// but warn that it might not capture system audio
			fmt.Println("Note: macOS loopback capture requires a virtual audio device (BlackHole/Soundflower)")
		}
		deviceType = malgo.Loopback
	}

	// Configure device
	deviceConfig := malgo.DefaultDeviceConfig(deviceType)
	deviceConfig.Capture.Format = malgo.FormatF32
	deviceConfig.Capture.Channels = 1
	deviceConfig.SampleRate = uint32(c.config.SampleRate)
	deviceConfig.PeriodSizeInFrames = uint32(c.config.ChunkSize)
	deviceConfig.Periods = 2

	// Find device by name if specified
	if c.config.DeviceName != "" {
		searchType := malgo.Capture
		if c.config.Loopback {
			searchType = malgo.Playback // For loopback, search in playback devices
		}
		dev, err := c.findDeviceByNameAndType(c.config.DeviceName, searchType)
		if err != nil {
			c.context.Uninit()
			c.context.Free()
			return fmt.Errorf("device not found: %w", err)
		}
		if c.config.Loopback {
			deviceConfig.Playback.DeviceID = dev.ID.Pointer()
		} else {
			deviceConfig.Capture.DeviceID = dev.ID.Pointer()
		}
		c.deviceName = dev.Name()
		c.isLoopback = c.config.Loopback
	} else {
		if c.config.Loopback {
			c.deviceName = "Default Output (Loopback)"
		} else {
			c.deviceName = "Default"
		}
		c.isLoopback = c.config.Loopback
	}

	// Create capture callback
	onRecvFrames := func(pOutputSample, pInputSamples []byte, frameCount uint32) {
		if len(pInputSamples) == 0 {
			return
		}

		// Convert bytes to float64
		samples := make([]float64, 0, frameCount)
		for i := uint32(0); i < frameCount && i*4+4 <= uint32(len(pInputSamples)); i++ {
			bits := binary.LittleEndian.Uint32(pInputSamples[i*4:])
			f32 := math.Float32frombits(bits)
			samples = append(samples, float64(f32))
		}

		if len(samples) > 0 {
			// Non-blocking send
			select {
			case c.dataChan <- samples:
			default:
				// Channel full, drop frame
			}
		}
	}

	deviceCallbacks := malgo.DeviceCallbacks{
		Data: onRecvFrames,
	}

	device, err := malgo.InitDevice(c.context.Context, deviceConfig, deviceCallbacks)
	if err != nil {
		// On macOS, if loopback fails, try falling back to regular capture
		if c.config.Loopback && runtime.GOOS == "darwin" {
			fmt.Println("Loopback capture failed on macOS, falling back to microphone capture")
			fmt.Println("For system audio capture on macOS, install BlackHole: https://existential.audio/blackhole/")
			deviceConfig = malgo.DefaultDeviceConfig(malgo.Capture)
			deviceConfig.Capture.Format = malgo.FormatF32
			deviceConfig.Capture.Channels = 1
			deviceConfig.SampleRate = uint32(c.config.SampleRate)
			deviceConfig.PeriodSizeInFrames = uint32(c.config.ChunkSize)
			deviceConfig.Periods = 2
			c.isLoopback = false
			c.deviceName = "Default (Microphone)"

			device, err = malgo.InitDevice(c.context.Context, deviceConfig, deviceCallbacks)
			if err != nil {
				c.context.Uninit()
				c.context.Free()
				return fmt.Errorf("failed to initialize fallback device: %w", err)
			}
		} else {
			c.context.Uninit()
			c.context.Free()
			return fmt.Errorf("failed to initialize device: %w", err)
		}
	}
	c.device = device

	if err := c.device.Start(); err != nil {
		c.device.Uninit()
		c.context.Uninit()
		c.context.Free()
		return fmt.Errorf("failed to start device: %w", err)
	}

	c.running = true
	return nil
}

// findDeviceByName finds a device info by name (searches capture devices)
func (c *Capture) findDeviceByName(name string) (*malgo.DeviceInfo, error) {
	return c.findDeviceByNameAndType(name, malgo.Capture)
}

// findDeviceByNameAndType finds a device info by name and type
func (c *Capture) findDeviceByNameAndType(name string, deviceType malgo.DeviceType) (*malgo.DeviceInfo, error) {
	devices, err := c.context.Devices(deviceType)
	if err != nil {
		return nil, err
	}

	for i := range devices {
		if devices[i].Name() == name {
			return &devices[i], nil
		}
	}

	return nil, fmt.Errorf("device not found: %s", name)
}

// Stop stops audio capture
func (c *Capture) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return nil
	}

	if c.device != nil {
		c.device.Stop()
		c.device.Uninit()
		c.device = nil
	}

	if c.context != nil {
		c.context.Uninit()
		c.context.Free()
		c.context = nil
	}

	c.running = false
	return nil
}

// Close releases all resources
func (c *Capture) Close() error {
	return c.Stop()
}

// DataChannel returns the channel for receiving audio data
func (c *Capture) DataChannel() <-chan []float64 {
	return c.dataChan
}

// IsRunning returns whether capture is active
func (c *Capture) IsRunning() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.running
}

// GetDeviceName returns the current device name
func (c *Capture) GetDeviceName() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.deviceName != "" {
		return c.deviceName
	}
	return c.config.DeviceName
}

// SetDevice changes the capture device (requires restart)
func (c *Capture) SetDevice(name string) error {
	return c.SetDeviceWithLoopback(name, false)
}

// SetDeviceWithLoopback changes the capture device with optional loopback mode
func (c *Capture) SetDeviceWithLoopback(name string, loopback bool) error {
	c.mu.Lock()
	wasRunning := c.running
	c.mu.Unlock()

	if wasRunning {
		if err := c.Stop(); err != nil {
			return err
		}
	}

	c.mu.Lock()
	c.config.DeviceName = name
	c.config.Loopback = loopback
	c.mu.Unlock()

	if wasRunning {
		return c.Start()
	}
	return nil
}

// IsLoopback returns whether currently capturing in loopback mode
func (c *Capture) IsLoopback() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isLoopback
}

// GetSampleRate returns the current sample rate
func (c *Capture) GetSampleRate() int {
	return c.config.SampleRate
}

// GetChunkSize returns the current chunk size
func (c *Capture) GetChunkSize() int {
	return c.config.ChunkSize
}

// IsDemoMode returns whether running in demo mode
func (c *Capture) IsDemoMode() bool {
	return c.demoMode
}
