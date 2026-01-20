package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"go_audio_visualizer/internal/audio"
	"go_audio_visualizer/internal/config"
	"go_audio_visualizer/internal/fft"
)

// VisualizationMode represents the current visualization type
type VisualizationMode int

const (
	ModeBars VisualizationMode = iota
	ModeWaveform
)

func (m VisualizationMode) String() string {
	switch m {
	case ModeBars:
		return "Bars"
	case ModeWaveform:
		return "Waveform"
	default:
		return "Unknown"
	}
}

// Model is the main bubbletea model
type Model struct {
	// Dimensions
	width  int
	height int

	// Configuration
	config     *config.Config
	configPath string

	// Audio capture
	capture     *audio.Capture
	deviceList  *audio.DeviceList
	deviceIndex int

	// FFT processor
	processor *fft.Processor

	// Visualization state
	mode       VisualizationMode
	spectrum   []float64
	peaks      []float64
	showHelp   bool
	showPeaks  bool
	useGradient bool
	mirrorBars bool

	// Renderers
	barRenderer      *BarRenderer
	waveformRenderer *WaveformRenderer

	// Status
	lastError   string
	statusMsg   string
	fps         int
	frameCount  int
	lastFPSTime time.Time

	// Ready state
	ready bool
}

// NewModel creates a new UI model
func NewModel(cfg *config.Config, configPath string) *Model {
	m := &Model{
		config:      cfg,
		configPath:  configPath,
		mode:        ModeBars,
		spectrum:    make([]float64, cfg.Display.BarCount),
		peaks:       make([]float64, cfg.Display.BarCount),
		showPeaks:   cfg.Visualization.ShowPeaks,
		useGradient: cfg.Visualization.UseGradient,
		mirrorBars:  cfg.Visualization.MirrorBars,
		lastFPSTime: time.Now(),
	}

	// Set visualization mode from config
	if cfg.Visualization.Mode == "waveform" {
		m.mode = ModeWaveform
	}

	return m
}

// SetCapture sets the audio capture instance
func (m *Model) SetCapture(cap *audio.Capture) {
	m.capture = cap
}

// SetProcessor sets the FFT processor instance
func (m *Model) SetProcessor(proc *fft.Processor) {
	m.processor = proc
}

// SetDeviceList sets the available devices
func (m *Model) SetDeviceList(list *audio.DeviceList) {
	m.deviceList = list
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		tickCmd(),
	)
}

// tickMsg is sent on every frame tick
type tickMsg time.Time

// audioDataMsg contains new audio data
type audioDataMsg struct {
	data []float64
}

// tickCmd returns a command that sends tick messages
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second/60, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateRenderers()
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tickMsg:
		m.updateFPS()
		return m, tea.Batch(tickCmd(), m.pollAudio())

	case audioDataMsg:
		if m.processor != nil {
			m.spectrum = m.processor.Process(msg.data)
			m.peaks = m.processor.GetPeaks()
		}
		return m, nil
	}

	return m, nil
}

// handleKeyPress processes keyboard input
func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	action := HandleKeyPress(msg)

	switch action {
	case ActionQuit:
		return m, tea.Quit

	case ActionToggleMode:
		m.toggleMode()

	case ActionCycleDevice:
		m.cycleDevice()

	case ActionIncreaseSensitivity:
		m.adjustSensitivity(0.1)

	case ActionDecreaseSensitivity:
		m.adjustSensitivity(-0.1)

	case ActionIncreaseSmoothing:
		m.adjustSmoothing(0.05)

	case ActionDecreaseSmoothing:
		m.adjustSmoothing(-0.05)

	case ActionTogglePeaks:
		m.showPeaks = !m.showPeaks
		m.config.Visualization.ShowPeaks = m.showPeaks

	case ActionToggleGradient:
		m.useGradient = !m.useGradient
		m.config.Visualization.UseGradient = m.useGradient

	case ActionToggleMirror:
		m.mirrorBars = !m.mirrorBars
		m.config.Visualization.MirrorBars = m.mirrorBars

	case ActionToggleHelp:
		m.showHelp = !m.showHelp

	case ActionSaveConfig:
		m.saveConfig()

	case ActionReset:
		m.resetConfig()
	}

	return m, nil
}

// updateRenderers creates/updates the visualization renderers
func (m *Model) updateRenderers() {
	// Reserve space for title and status bars
	visHeight := m.height - 4
	if visHeight < 1 {
		visHeight = 1
	}

	m.barRenderer = NewBarRenderer(m.width-2, visHeight)
	m.barRenderer.SetOptions(m.useGradient, m.showPeaks, m.mirrorBars)

	m.waveformRenderer = NewWaveformRenderer(m.width-2, visHeight)
	m.waveformRenderer.SetOptions(m.useGradient, true)
}

// toggleMode switches between visualization modes
func (m *Model) toggleMode() {
	if m.mode == ModeBars {
		m.mode = ModeWaveform
		m.config.Visualization.Mode = "waveform"
	} else {
		m.mode = ModeBars
		m.config.Visualization.Mode = "bars"
	}
	m.statusMsg = fmt.Sprintf("Mode: %s", m.mode)
}

// cycleDevice switches to the next audio device
func (m *Model) cycleDevice() {
	if m.deviceList == nil || len(m.deviceList.CaptureDevices) == 0 {
		m.statusMsg = "No input devices available"
		return
	}

	m.deviceIndex = (m.deviceIndex + 1) % len(m.deviceList.CaptureDevices)
	device := m.deviceList.CaptureDevices[m.deviceIndex]

	if m.capture != nil {
		if err := m.capture.SetDevice(device.Name); err != nil {
			m.lastError = err.Error()
			m.statusMsg = fmt.Sprintf("Failed to switch device: %s", err)
		} else {
			m.config.Audio.Device = device.Name
			m.statusMsg = fmt.Sprintf("Device: %s", device.Name)
		}
	}
}

// adjustSensitivity changes the sensitivity
func (m *Model) adjustSensitivity(delta float64) {
	newSens := m.config.Sensitivity.Sensitivity + delta
	if newSens < 0.1 {
		newSens = 0.1
	}
	if newSens > 10.0 {
		newSens = 10.0
	}
	m.config.Sensitivity.Sensitivity = newSens

	if m.processor != nil {
		m.processor.SetSensitivity(newSens)
	}

	m.statusMsg = fmt.Sprintf("Sensitivity: %.1f", newSens)
}

// adjustSmoothing changes the smoothing values
func (m *Model) adjustSmoothing(delta float64) {
	newAttack := m.config.Smoothing.AttackSpeed + delta
	newDecay := m.config.Smoothing.DecaySpeed + delta*0.5

	if newAttack < 0.1 {
		newAttack = 0.1
	}
	if newAttack > 1.0 {
		newAttack = 1.0
	}
	if newDecay < 0.05 {
		newDecay = 0.05
	}
	if newDecay > 0.9 {
		newDecay = 0.9
	}

	m.config.Smoothing.AttackSpeed = newAttack
	m.config.Smoothing.DecaySpeed = newDecay

	if m.processor != nil {
		m.processor.SetSmoothing(newAttack, newDecay, m.config.Smoothing.RestDecay)
	}

	m.statusMsg = fmt.Sprintf("Smoothing: attack=%.2f decay=%.2f", newAttack, newDecay)
}

// saveConfig saves the current configuration
func (m *Model) saveConfig() {
	if err := m.config.Save(m.configPath); err != nil {
		m.lastError = err.Error()
		m.statusMsg = fmt.Sprintf("Failed to save config: %s", err)
	} else {
		m.statusMsg = "Config saved"
	}
}

// resetConfig resets to default configuration
func (m *Model) resetConfig() {
	m.config.Reset()

	// Apply reset values
	m.showPeaks = m.config.Visualization.ShowPeaks
	m.useGradient = m.config.Visualization.UseGradient
	m.mirrorBars = m.config.Visualization.MirrorBars

	if m.config.Visualization.Mode == "waveform" {
		m.mode = ModeWaveform
	} else {
		m.mode = ModeBars
	}

	if m.processor != nil {
		m.processor.SetSensitivity(m.config.Sensitivity.Sensitivity)
		m.processor.SetSmoothing(
			m.config.Smoothing.AttackSpeed,
			m.config.Smoothing.DecaySpeed,
			m.config.Smoothing.RestDecay,
		)
	}

	m.updateRenderers()
	m.statusMsg = "Reset to defaults"
}

// pollAudio checks for new audio data
func (m *Model) pollAudio() tea.Cmd {
	if m.capture == nil {
		return nil
	}

	return func() tea.Msg {
		select {
		case data, ok := <-m.capture.DataChannel():
			if ok && len(data) > 0 {
				return audioDataMsg{data: data}
			}
		default:
		}
		return nil
	}
}

// updateFPS updates the FPS counter
func (m *Model) updateFPS() {
	m.frameCount++
	now := time.Now()
	elapsed := now.Sub(m.lastFPSTime)
	if elapsed >= time.Second {
		m.fps = m.frameCount
		m.frameCount = 0
		m.lastFPSTime = now
	}
}

// GetFPS returns the current FPS
func (m *Model) GetFPS() int {
	return m.fps
}

// GetDeviceName returns the current device name
func (m *Model) GetDeviceName() string {
	if m.capture != nil {
		return m.capture.GetDeviceName()
	}
	return "No device"
}
