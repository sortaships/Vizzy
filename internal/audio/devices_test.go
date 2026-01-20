package audio

import (
	"strings"
	"testing"
)

func TestIsLoopbackDevice(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"Stereo Mix", true},
		{"Loopback Audio", true},
		{"What U Hear", true},
		{"What You Hear", true},
		{"Microphone", false},
		{"Speakers", false},
		{"Built-in Microphone", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isLoopbackDevice(tt.name)
			if result != tt.expected {
				t.Errorf("isLoopbackDevice(%q) = %v, expected %v",
					tt.name, result, tt.expected)
			}
		})
	}
}

func TestDeviceString(t *testing.T) {
	dev := Device{
		Name:       "Test Device",
		IsDefault:  true,
		IsCapture:  true,
		IsPlayback: false,
		IsLoopback: false,
	}

	str := dev.String()

	if !strings.Contains(str, "Test Device") {
		t.Error("Device string should contain device name")
	}
	if !strings.Contains(str, "Capture") {
		t.Error("Device string should contain device type")
	}
	if !strings.Contains(str, "[Default]") {
		t.Error("Device string should indicate default")
	}
}

func TestDeviceStringWithLoopback(t *testing.T) {
	dev := Device{
		Name:       "Stereo Mix",
		IsDefault:  false,
		IsCapture:  true,
		IsPlayback: false,
		IsLoopback: true,
	}

	str := dev.String()

	if !strings.Contains(str, "[Loopback]") {
		t.Error("Device string should indicate loopback")
	}
}

// Integration tests - require audio hardware

func TestGetDevices(t *testing.T) {
	list, err := GetDevices()
	if err != nil {
		t.Skipf("Skipping test - audio not available: %v", err)
	}

	if list == nil {
		t.Fatal("GetDevices returned nil list")
	}

	t.Logf("Found %d capture devices, %d playback devices",
		len(list.CaptureDevices), len(list.PlaybackDevices))
}

func TestGetDefaultCaptureDevice(t *testing.T) {
	dev, err := GetDefaultCaptureDevice()
	if err != nil {
		t.Skipf("Skipping test - no default capture device: %v", err)
	}

	if dev == nil {
		t.Fatal("GetDefaultCaptureDevice returned nil")
	}

	t.Logf("Default capture device: %s", dev.String())
}

func TestListDevices(t *testing.T) {
	str, err := ListDevices()
	if err != nil {
		t.Skipf("Skipping test - audio not available: %v", err)
	}

	if str == "" {
		t.Error("ListDevices returned empty string")
	}

	if !strings.Contains(str, "Capture Devices:") {
		t.Error("ListDevices should contain 'Capture Devices:'")
	}

	t.Logf("Device list:\n%s", str)
}
