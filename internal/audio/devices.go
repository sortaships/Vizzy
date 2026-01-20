package audio

import (
	"fmt"
	"strings"

	"github.com/gen2brain/malgo"
)

// Device represents an audio device
type Device struct {
	ID         string
	Name       string
	IsDefault  bool
	IsCapture  bool
	IsPlayback bool
	IsLoopback bool
}

// DeviceList holds available audio devices
type DeviceList struct {
	CaptureDevices  []Device
	PlaybackDevices []Device
	LoopbackDevices []Device
}

// GetDevices enumerates all available audio devices
func GetDevices() (*DeviceList, error) {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize audio context: %w", err)
	}
	defer func() {
		_ = ctx.Uninit()
		ctx.Free()
	}()

	list := &DeviceList{
		CaptureDevices:  make([]Device, 0),
		PlaybackDevices: make([]Device, 0),
		LoopbackDevices: make([]Device, 0),
	}

	// Get capture devices
	captureDevices, err := ctx.Devices(malgo.Capture)
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate capture devices: %w", err)
	}

	for _, info := range captureDevices {
		dev := Device{
			ID:         string(info.ID[:]),
			Name:       info.Name(),
			IsDefault:  info.IsDefault != 0,
			IsCapture:  true,
			IsPlayback: false,
			IsLoopback: isLoopbackDevice(info.Name()),
		}
		list.CaptureDevices = append(list.CaptureDevices, dev)

		if dev.IsLoopback {
			list.LoopbackDevices = append(list.LoopbackDevices, dev)
		}
	}

	// Get playback devices
	playbackDevices, err := ctx.Devices(malgo.Playback)
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate playback devices: %w", err)
	}

	for _, info := range playbackDevices {
		dev := Device{
			ID:         string(info.ID[:]),
			Name:       info.Name(),
			IsDefault:  info.IsDefault != 0,
			IsCapture:  false,
			IsPlayback: true,
			IsLoopback: false,
		}
		list.PlaybackDevices = append(list.PlaybackDevices, dev)
	}

	return list, nil
}

// GetDefaultCaptureDevice returns the default capture device
func GetDefaultCaptureDevice() (*Device, error) {
	list, err := GetDevices()
	if err != nil {
		return nil, err
	}

	for _, dev := range list.CaptureDevices {
		if dev.IsDefault {
			return &dev, nil
		}
	}

	if len(list.CaptureDevices) > 0 {
		return &list.CaptureDevices[0], nil
	}

	return nil, fmt.Errorf("no capture devices found")
}

// GetDeviceByName finds a device by name
func GetDeviceByName(name string) (*Device, error) {
	list, err := GetDevices()
	if err != nil {
		return nil, err
	}

	for _, dev := range list.CaptureDevices {
		if dev.Name == name {
			return &dev, nil
		}
	}

	nameLower := strings.ToLower(name)
	for _, dev := range list.CaptureDevices {
		if strings.Contains(strings.ToLower(dev.Name), nameLower) {
			return &dev, nil
		}
	}

	return nil, fmt.Errorf("device not found: %s", name)
}

// GetCaptureDeviceByIndex returns a capture device by its index
func GetCaptureDeviceByIndex(index int) (*Device, error) {
	list, err := GetDevices()
	if err != nil {
		return nil, err
	}

	if index < 0 || index >= len(list.CaptureDevices) {
		return nil, fmt.Errorf("invalid capture device index: %d", index)
	}

	return &list.CaptureDevices[index], nil
}

// isLoopbackDevice checks if a device is likely a loopback/stereo mix device
func isLoopbackDevice(name string) bool {
	nameLower := strings.ToLower(name)
	return strings.Contains(nameLower, "loopback") ||
		strings.Contains(nameLower, "stereo mix") ||
		strings.Contains(nameLower, "what u hear") ||
		strings.Contains(nameLower, "wave out") ||
		strings.Contains(nameLower, "what you hear")
}

// String returns a formatted string representation of the device
func (d Device) String() string {
	flags := ""
	if d.IsDefault {
		flags += " [Default]"
	}
	if d.IsLoopback {
		flags += " [Loopback]"
	}

	deviceType := "Capture"
	if d.IsPlayback {
		deviceType = "Playback"
	}

	return fmt.Sprintf("%s (%s)%s", d.Name, deviceType, flags)
}

// ListDevices returns a formatted list of all devices
func ListDevices() (string, error) {
	list, err := GetDevices()
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("Capture Devices:\n")
	for i, dev := range list.CaptureDevices {
		sb.WriteString(fmt.Sprintf("  %d: %s\n", i, dev.String()))
	}

	sb.WriteString("\nPlayback Devices:\n")
	for i, dev := range list.PlaybackDevices {
		sb.WriteString(fmt.Sprintf("  %d: %s\n", i, dev.String()))
	}

	if len(list.LoopbackDevices) > 0 {
		sb.WriteString("\nLoopback Devices:\n")
		for i, dev := range list.LoopbackDevices {
			sb.WriteString(fmt.Sprintf("  %d: %s\n", i, dev.String()))
		}
	}

	return sb.String(), nil
}
