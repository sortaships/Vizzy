package audio

import (
	"testing"
	"time"
)

func TestDefaultCaptureConfig(t *testing.T) {
	cfg := DefaultCaptureConfig()

	if cfg.SampleRate != 44100 {
		t.Errorf("expected SampleRate=44100, got %d", cfg.SampleRate)
	}
	if cfg.ChunkSize != 1024 {
		t.Errorf("expected ChunkSize=1024, got %d", cfg.ChunkSize)
	}
	if cfg.DeviceName != "" {
		t.Errorf("expected DeviceName empty, got %s", cfg.DeviceName)
	}
}

func TestNewCapture(t *testing.T) {
	cfg := DefaultCaptureConfig()
	cap, err := NewCapture(cfg)

	if err != nil {
		t.Fatalf("NewCapture failed: %v", err)
	}
	if cap == nil {
		t.Fatal("NewCapture returned nil")
	}
	if cap.dataChan == nil {
		t.Error("dataChan should not be nil")
	}
}

func TestCaptureStartStop(t *testing.T) {
	cfg := DefaultCaptureConfig()
	cap, err := NewCapture(cfg)
	if err != nil {
		t.Fatalf("NewCapture failed: %v", err)
	}

	err = cap.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if !cap.IsRunning() {
		t.Error("Capture should be running after Start")
	}

	// Let it run briefly
	time.Sleep(100 * time.Millisecond)

	err = cap.Stop()
	if err != nil {
		t.Errorf("Stop failed: %v", err)
	}

	if cap.IsRunning() {
		t.Error("Capture should not be running after Stop")
	}
}

func TestCaptureDoubleStart(t *testing.T) {
	cfg := DefaultCaptureConfig()
	cap, err := NewCapture(cfg)
	if err != nil {
		t.Fatalf("NewCapture failed: %v", err)
	}
	defer cap.Close()

	err = cap.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Double start should be safe
	err = cap.Start()
	if err != nil {
		t.Errorf("Double start should be safe: %v", err)
	}
}

func TestCaptureDoubleStop(t *testing.T) {
	cfg := DefaultCaptureConfig()
	cap, err := NewCapture(cfg)
	if err != nil {
		t.Fatalf("NewCapture failed: %v", err)
	}

	// Stop without start should be safe
	err = cap.Stop()
	if err != nil {
		t.Errorf("Stop without start should be safe: %v", err)
	}

	// Double stop should be safe
	err = cap.Stop()
	if err != nil {
		t.Errorf("Double stop should be safe: %v", err)
	}
}

func TestCaptureGetters(t *testing.T) {
	cfg := CaptureConfig{
		SampleRate: 48000,
		ChunkSize:  2048,
		DeviceName: "Test",
	}
	cap, err := NewCapture(cfg)
	if err != nil {
		t.Fatalf("NewCapture failed: %v", err)
	}

	if cap.GetSampleRate() != 48000 {
		t.Errorf("expected SampleRate=48000, got %d", cap.GetSampleRate())
	}
	if cap.GetChunkSize() != 2048 {
		t.Errorf("expected ChunkSize=2048, got %d", cap.GetChunkSize())
	}
}

func TestCaptureDataChannel(t *testing.T) {
	cfg := DefaultCaptureConfig()
	cap, err := NewCapture(cfg)
	if err != nil {
		t.Fatalf("NewCapture failed: %v", err)
	}

	ch := cap.DataChannel()
	if ch == nil {
		t.Error("DataChannel should not return nil")
	}
}

func TestCaptureReceiveData(t *testing.T) {
	cfg := DefaultCaptureConfig()
	cap, err := NewCapture(cfg)
	if err != nil {
		t.Fatalf("NewCapture failed: %v", err)
	}
	defer cap.Close()

	err = cap.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Try to receive some data (demo mode should always produce data)
	select {
	case data := <-cap.DataChannel():
		if len(data) == 0 {
			t.Error("Received empty data")
		}
		t.Logf("Received %d samples", len(data))
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for data")
	}
}

func TestCaptureDemoMode(t *testing.T) {
	cfg := DefaultCaptureConfig()
	cap, err := NewCapture(cfg)
	if err != nil {
		t.Fatalf("NewCapture failed: %v", err)
	}

	if !cap.IsDemoMode() {
		t.Error("Capture should be in demo mode without CGO")
	}
}

func TestCaptureSetDevice(t *testing.T) {
	cfg := DefaultCaptureConfig()
	cap, err := NewCapture(cfg)
	if err != nil {
		t.Fatalf("NewCapture failed: %v", err)
	}

	err = cap.SetDevice("TestDevice")
	if err != nil {
		t.Errorf("SetDevice should not fail: %v", err)
	}

	// In demo mode, device name should still be "Demo Mode"
	name := cap.GetDeviceName()
	if name != "Demo Mode" {
		t.Errorf("expected GetDeviceName to return 'Demo Mode', got %s", name)
	}
}
