package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"go_audio_visualizer/internal/audio"
	"go_audio_visualizer/internal/config"
	"go_audio_visualizer/internal/fft"
	"go_audio_visualizer/internal/ui"
)

const (
	defaultConfigPath = "config.json"
	version           = "1.0.0"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", defaultConfigPath, "Path to configuration file")
	listDevices := flag.Bool("list-devices", false, "List available audio devices and exit")
	showVersion := flag.Bool("version", false, "Show version and exit")
	deviceName := flag.String("device", "", "Audio device to use (overrides config)")
	flag.Parse()

	if *showVersion {
		fmt.Printf("Audio Visualizer v%s\n", version)
		os.Exit(0)
	}

	// Handle list devices
	if *listDevices {
		if err := printDevices(); err != nil {
			log.Fatalf("Failed to list devices: %v", err)
		}
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Printf("Warning: Failed to load config from %s: %v", *configPath, err)
		cfg = config.Default()
	}
	cfg.Validate()

	// Override device if specified
	if *deviceName != "" {
		cfg.Audio.Device = *deviceName
	}

	// Run the application
	if err := run(cfg, *configPath); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}

func run(cfg *config.Config, configPath string) error {
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Get device list
	deviceList, err := audio.GetDevices()
	if err != nil {
		log.Printf("Warning: Failed to enumerate devices: %v", err)
	}

	// Create audio capture
	captureCfg := audio.CaptureConfig{
		SampleRate: cfg.Audio.SampleRate,
		ChunkSize:  cfg.Audio.ChunkSize,
		DeviceName: cfg.Audio.Device,
	}
	capture, err := audio.NewCapture(captureCfg)
	if err != nil {
		return fmt.Errorf("failed to create audio capture: %w", err)
	}

	// Start audio capture
	if err := capture.Start(); err != nil {
		log.Printf("Warning: Failed to start audio capture: %v", err)
		log.Println("Running in demo mode without audio input")
	}

	// Create FFT processor
	processorCfg := fft.ProcessorConfig{
		SampleRate:  cfg.Audio.SampleRate,
		ChunkSize:   cfg.Audio.ChunkSize,
		NumBands:    cfg.Display.BarCount,
		AttackSpeed: cfg.Smoothing.AttackSpeed,
		DecaySpeed:  cfg.Smoothing.DecaySpeed,
		RestDecay:   cfg.Smoothing.RestDecay,
		BassBoost:   cfg.Sensitivity.BassBoost,
		MidBoost:    cfg.Sensitivity.MidBoost,
		TrebleBoost: cfg.Sensitivity.TrebleBoost,
		MinFreq:     cfg.Sensitivity.MinFrequency,
		MaxFreq:     cfg.Sensitivity.MaxFrequency,
		NoiseFloor:  cfg.Sensitivity.NoiseFloor,
		Sensitivity: cfg.Sensitivity.Sensitivity,
	}
	processor := fft.NewProcessor(processorCfg)

	// Create UI model
	model := ui.NewModel(cfg, configPath)
	model.SetCapture(capture)
	model.SetProcessor(processor)
	model.SetDeviceList(deviceList)

	// Create bubbletea program
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Handle signals in a goroutine
	go func() {
		<-sigChan
		capture.Close()
		p.Quit()
	}()

	// Run the program
	if _, err := p.Run(); err != nil {
		capture.Close()
		return fmt.Errorf("failed to run program: %w", err)
	}

	// Cleanup
	capture.Close()
	return nil
}

func printDevices() error {
	list, err := audio.ListDevices()
	if err != nil {
		return err
	}
	fmt.Println(list)
	return nil
}
