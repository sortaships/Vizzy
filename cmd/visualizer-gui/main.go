package main

import (
	"flag"
	"fmt"
	"image/color"
	"log"
	"math"
	"os"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"go_audio_visualizer/internal/audio"
	"go_audio_visualizer/internal/config"
	"go_audio_visualizer/internal/fft"
)

const (
	defaultWidth  = 1280
	defaultHeight = 720
	version       = "1.0.0"
)

// Visualization mode names
var vizModeNames = []string{"Line", "Bars", "Rising", "Spiral", "Isometric", "Particles", "Circular", "Wave"}

// ColorPreset represents a color scheme for the waveform
type ColorPreset struct {
	Name        string
	LineColor   color.RGBA
	MirrorColor color.RGBA
	GlowColor   color.RGBA
}

// Predefined color presets
var colorPresets = []ColorPreset{
	// Classic colors
	{"Cyan", color.RGBA{0, 255, 200, 255}, color.RGBA{0, 127, 100, 255}, color.RGBA{0, 255, 200, 80}},
	{"Purple", color.RGBA{180, 100, 255, 255}, color.RGBA{90, 50, 127, 255}, color.RGBA{180, 100, 255, 80}},
	{"Orange", color.RGBA{255, 150, 50, 255}, color.RGBA{127, 75, 25, 255}, color.RGBA{255, 150, 50, 80}},
	{"Green", color.RGBA{100, 255, 100, 255}, color.RGBA{50, 127, 50, 255}, color.RGBA{100, 255, 100, 80}},
	{"Blue", color.RGBA{80, 150, 255, 255}, color.RGBA{40, 75, 127, 255}, color.RGBA{80, 150, 255, 80}},
	{"Pink", color.RGBA{255, 100, 180, 255}, color.RGBA{127, 50, 90, 255}, color.RGBA{255, 100, 180, 80}},
	{"Gold", color.RGBA{255, 215, 0, 255}, color.RGBA{127, 107, 0, 255}, color.RGBA{255, 215, 0, 80}},
	{"White", color.RGBA{240, 240, 240, 255}, color.RGBA{120, 120, 120, 255}, color.RGBA{240, 240, 240, 80}},
	// Vibrant colors
	{"Neon Red", color.RGBA{255, 50, 50, 255}, color.RGBA{180, 25, 25, 255}, color.RGBA{255, 50, 50, 100}},
	{"Electric Blue", color.RGBA{30, 144, 255, 255}, color.RGBA{15, 72, 127, 255}, color.RGBA{30, 144, 255, 100}},
	{"Lime", color.RGBA{180, 255, 0, 255}, color.RGBA{90, 127, 0, 255}, color.RGBA{180, 255, 0, 80}},
	{"Magenta", color.RGBA{255, 0, 255, 255}, color.RGBA{127, 0, 127, 255}, color.RGBA{255, 0, 255, 80}},
	{"Aqua", color.RGBA{0, 255, 255, 255}, color.RGBA{0, 127, 127, 255}, color.RGBA{0, 255, 255, 80}},
	{"Coral", color.RGBA{255, 127, 80, 255}, color.RGBA{200, 80, 50, 255}, color.RGBA{255, 127, 80, 80}},
	{"Violet", color.RGBA{138, 43, 226, 255}, color.RGBA{69, 21, 113, 255}, color.RGBA{138, 43, 226, 80}},
	{"Mint", color.RGBA{152, 255, 152, 255}, color.RGBA{76, 127, 76, 255}, color.RGBA{152, 255, 152, 80}},
	// Themed colors
	{"Sunset", color.RGBA{255, 100, 50, 255}, color.RGBA{200, 50, 100, 255}, color.RGBA{255, 150, 50, 100}},
	{"Ocean", color.RGBA{0, 180, 220, 255}, color.RGBA{0, 80, 120, 255}, color.RGBA{0, 200, 255, 80}},
	{"Forest", color.RGBA{34, 180, 34, 255}, color.RGBA{17, 90, 17, 255}, color.RGBA{50, 200, 50, 80}},
	{"Fire", color.RGBA{255, 80, 0, 255}, color.RGBA{200, 40, 0, 255}, color.RGBA{255, 120, 0, 100}},
	{"Ice", color.RGBA{180, 220, 255, 255}, color.RGBA{100, 150, 200, 255}, color.RGBA{200, 230, 255, 80}},
	{"Lava", color.RGBA{255, 50, 0, 255}, color.RGBA{180, 20, 0, 255}, color.RGBA{255, 100, 0, 100}},
	// Two-tone color themes (contrasting line/mirror colors)
	{"Cyber", color.RGBA{0, 255, 255, 255}, color.RGBA{255, 0, 128, 255}, color.RGBA{0, 255, 255, 80}},           // Cyan line, hot pink mirror
	{"Neon Night", color.RGBA{255, 20, 147, 255}, color.RGBA{0, 255, 127, 255}, color.RGBA{255, 20, 147, 80}},    // Deep pink line, spring green mirror
	{"Electric", color.RGBA{255, 255, 0, 255}, color.RGBA{138, 43, 226, 255}, color.RGBA{255, 255, 0, 80}},       // Yellow line, blue-violet mirror
	{"Toxic", color.RGBA{57, 255, 20, 255}, color.RGBA{148, 0, 211, 255}, color.RGBA{57, 255, 20, 80}},           // Neon green line, dark violet mirror
	{"Bloodmoon", color.RGBA{220, 20, 60, 255}, color.RGBA{255, 165, 0, 255}, color.RGBA{220, 20, 60, 80}},       // Crimson line, orange mirror
	{"Arctic", color.RGBA{135, 206, 250, 255}, color.RGBA{255, 182, 193, 255}, color.RGBA{135, 206, 250, 80}},    // Light sky blue line, light pink mirror
	{"Synthwave", color.RGBA{255, 0, 255, 255}, color.RGBA{0, 191, 255, 255}, color.RGBA{255, 0, 255, 80}},       // Magenta line, deep sky blue mirror
	{"Autumn", color.RGBA{255, 140, 0, 255}, color.RGBA{139, 69, 19, 255}, color.RGBA{255, 140, 0, 80}},          // Dark orange line, saddle brown mirror
	{"Vaporwave", color.RGBA{255, 105, 180, 255}, color.RGBA{64, 224, 208, 255}, color.RGBA{255, 105, 180, 80}},  // Hot pink line, turquoise mirror
	{"Retrowave", color.RGBA{255, 215, 0, 255}, color.RGBA{255, 20, 147, 255}, color.RGBA{255, 215, 0, 80}},      // Gold line, deep pink mirror
	{"Matrix", color.RGBA{0, 255, 65, 255}, color.RGBA{0, 128, 0, 255}, color.RGBA{0, 255, 65, 80}},              // Bright green line, dark green mirror
	{"Twilight", color.RGBA{148, 0, 211, 255}, color.RGBA{255, 140, 0, 255}, color.RGBA{148, 0, 211, 80}},        // Dark violet line, dark orange mirror
	// Special animated
	{"Rainbow", color.RGBA{255, 0, 0, 255}, color.RGBA{0, 0, 255, 255}, color.RGBA{255, 255, 0, 80}}, // Cycles through hues
}

// LineStyle holds line rendering settings
type LineStyle struct {
	Thickness    float32 // Line thickness (1.0 - 8.0)
	ColorPreset  int     // Index into colorPresets
	GlowEnabled  bool    // Enable glow effect behind line
	GlowSize     float32 // Glow thickness multiplier
}

// OverlaySettings holds multi-line overlay settings
type OverlaySettings struct {
	Count       int     // Number of overlaid lines (1-5)
	Offset      float32 // Vertical offset between lines in pixels
	FadeEnabled bool    // Fade distant overlay lines
	TimeOffset  int     // Frame delay between overlays (creates trailing effect)
	SyncLines   bool    // If true, all overlay lines use same spectrum (no overlap)
}

// IsometricSettings holds 3D isometric rendering settings
type IsometricSettings struct {
	Enabled       bool    // Enable isometric 3D mode
	DepthLayers   int     // Number of depth layers (2-30)
	DepthSpacing  float32 // Spacing between depth layers
	Angle         float32 // Isometric angle in degrees (15-60)
	ScaleFactor   float32 // How much further layers shrink (0.8-0.98)
	DepthFade     bool    // Fade distant layers
	Rotation      float32 // Y-axis rotation angle (0-360)
	RotationSpeed float32 // Rotation speed in degrees per frame
	AutoRotate    bool    // Enable automatic rotation
	ForwardMotion bool    // Enable forward motion effect (moving through landscape)
	ForwardSpeed  float32 // Forward motion speed (-2.0 to 2.0)
	ForwardOffset float32 // Current forward motion offset (internal)
}

// BarSettings holds settings for the bar visualizer mode
type BarSettings struct {
	BarWidth     float32 // Width of each bar (0.5-1.0 as ratio of spacing)
	GapWidth     float32 // Gap between bars in pixels
	RoundedCaps  bool    // Use rounded tops on bars
	Mirror       bool    // Mirror bars below center line
	Gradient     bool    // Use gradient coloring based on amplitude
	MinHeight    float32 // Minimum bar height in pixels
	GlowEnabled  bool    // Enable glow effect on bars
	GlowSize     float32 // Size of glow effect
	Outline      bool    // Draw outline around bars
	OutlineWidth float32 // Width of outline
}

// RisingSettings holds settings for the rising/falling line visualizer mode
type RisingSettings struct {
	Gravity      float32 // How fast the line falls (0.01-0.5)
	MaxRise      float32 // Maximum rise speed (0.1-1.0)
	Bounce       float32 // Bounce factor when hitting bottom (0-0.5)
	TrailEnabled bool    // Show trail behind falling line
	TrailLength  int     // Number of trail segments
	TrailFade    float32 // How fast trail fades (0-1)
	Peaks        bool    // Show peak markers
	PeakHold     int     // Frames to hold peak before falling
	PeakFallRate float32 // How fast peaks fall
}

// Particle represents a single particle in the particles visualizer
type Particle struct {
	X, Y     float32 // Position
	VX, VY   float32 // Velocity
	Life     float32 // Remaining life (0-1)
	Size     float32 // Particle size
	ColorIdx int     // Which frequency band spawned this (for coloring)
}

// ParticleSettings holds settings for the particles visualizer mode
type ParticleSettings struct {
	MaxParticles  int     // Maximum number of particles
	SpawnRate     float32 // Particles spawned per frame per band (0.1-5)
	Gravity       float32 // Gravity pull on particles
	InitialSpeed  float32 // Initial upward velocity
	SpeedVariance float32 // Random variance in speed
	Spread        float32 // Horizontal spread angle
	MinLife       float32 // Minimum particle lifetime
	MaxLife       float32 // Maximum particle lifetime
	SizeMin       float32 // Minimum particle size
	SizeMax       float32 // Maximum particle size
	FadeOut       bool    // Fade particles as they age
	Glow          bool    // Add glow effect to particles
	Explosion     bool    // Burst mode - explode on beats
}

// CircularSettings holds settings for the circular/radial visualizer mode
type CircularSettings struct {
	InnerRadius   float32 // Inner radius of the circle
	BarLength     float32 // Maximum bar length
	BarWidth      float32 // Width of each bar
	Rotation      float32 // Current rotation angle
	RotationSpeed float32 // Rotation speed (degrees per frame)
	Mirror        bool    // Draw bars both inward and outward
	Dots          bool    // Use dots instead of bars
	DotSize       float32 // Size of dots
	Spiral        bool    // Spiral effect (offset based on frequency)
	Rainbow       bool    // Use rainbow coloring around circle
}

// WaveSettings holds settings for the wave/ripple visualizer mode
type WaveSettings struct {
	MaxRings      int     // Maximum number of rings
	RingSpeed     float32 // How fast rings expand
	RingWidth     float32 // Width of each ring
	Decay         float32 // How fast rings fade (0-1)
	SpawnRate     float32 // Rings spawned per frame when active
	ReactToBass   bool    // React primarily to bass frequencies
	Fill          bool    // Fill rings instead of stroke
	Wobble        bool    // Add wobble effect to rings
	WobbleAmount  float32 // Amount of wobble
	ColorByAge    bool    // Color rings based on age
}

// Ring represents a single expanding ring in the wave visualizer
type Ring struct {
	Radius    float32 // Current radius
	Intensity float32 // Ring intensity (affects opacity and width)
	Age       float32 // Age of ring (0-1, 1 = dead)
}

// DeviceEntry represents a device in the dropdown
type DeviceEntry struct {
	Name       string
	IsLoopback bool
	IsDefault  bool
}

// BeatDetector detects beats in audio based on energy levels
type BeatDetector struct {
	Enabled         bool      // Enable beat-reactive color changes
	energyHistory   []float64 // Rolling history of bass energy
	historySize     int       // Size of energy history buffer
	historyIndex    int       // Current position in history buffer
	threshold       float64   // Beat detection threshold (1.0-3.0)
	cooldown        int       // Frames to wait between beats
	cooldownCounter int       // Current cooldown counter
	beatDetected    bool      // Was a beat detected this frame
	beatIntensity   float64   // Intensity of the detected beat (0-1)
	decayRate       float64   // How fast the beat intensity decays
}

// SettingsMenu holds the collapsible settings menu state
type SettingsMenu struct {
	Visible       bool
	X, Y          int
	Width, Height int
	ActiveDial    int // -1 = none, otherwise index of dial being dragged
	DragStartY    int
	DragStartVal  float64
}

// DialControl represents a dial/slider in the settings menu
type DialControl struct {
	Key      string   // Unique identifier for applying values
	Name     string   // Display name
	Value    *float64
	Min, Max float64
	Step     float64
	Format   string // e.g., "%.1f" or "%.0f"
}

type Visualizer struct {
	// Window
	width  int
	height int

	// Config
	config     *config.Config
	configPath string

	// Audio
	capture    *audio.Capture
	deviceList *audio.DeviceList
	devices    []DeviceEntry // Combined capture + output (loopback) devices

	// FFT
	processor *fft.Processor

	// Spectrum data
	spectrum        []float64
	peaks           []float64
	spectrumHistory [][]float64 // History buffer for time-offset overlay
	historyIndex    int
	mu              sync.RWMutex

	// UI state
	showHelp          bool
	showDropdown      bool
	showVizDropdown   bool
	showColorDropdown bool
	showStylePanel    bool
	selectedDevice    int
	hoverDevice       int
	hoverVizMode      int
	hoverColorPreset  int
	colorDropdownScroll int // Scroll offset for color dropdown
	statusMessage     string
	statusTimer       int

	// Frequency shift slider
	freqShift      float64 // -1.0 to 1.0, shifts frequencies along the line
	draggingSlider bool

	// Frequency bands slider
	numBands            int // Number of frequency bands (16-256)
	draggingBandsSlider bool

	// Line style settings
	lineStyle    LineStyle
	overlay      OverlaySettings
	isometric    IsometricSettings
	barSettings  BarSettings      // Bar visualizer settings
	rising       RisingSettings   // Rising line visualizer settings
	particles    ParticleSettings // Particle visualizer settings
	circular     CircularSettings // Circular visualizer settings
	wave         WaveSettings     // Wave visualizer settings
	smoothCurves bool             // Use smooth curve interpolation
	frameCount   int              // For rainbow color animation
	damping      float64          // Damping factor to reduce reactivity (0.1 = very damped, 1.0 = full reactivity)

	// Rising line state (velocities and positions for physics simulation)
	risingPositions  []float64 // Current Y positions for each frequency band
	risingVelocities []float64 // Current velocities for each frequency band
	risingPeaks      []float64 // Peak positions for each band
	risingPeakHold   []int     // Frames remaining to hold peak

	// Particles state
	particleList []Particle // Active particles

	// Wave state
	rings []Ring // Active rings

	// Beat detection
	beatDetector BeatDetector

	// Visualization mode
	vizMode int // 0=Line, 1=Bars, 2=Rising, 3=Spiral, 4=Isometric, 5=Particles, 6=Circular, 7=Wave

	// Collapsible settings menu
	settingsMenu SettingsMenu
	// Dial values (intermediate float64 for smooth control)
	dialThickness    float64
	dialGlowSize     float64
	dialOverlayOff   float64
	dialOverlayCnt   float64
	dialDepthLayers  float64
	dialAngle        float64
	dialSpacing      float64
	dialRotation     float64 // Manual rotation control
	dialRotSpeed     float64 // Rotation speed
	dialForwardSpeed float64 // Forward motion speed
	dialDamping      float64 // Damping/reactivity control
	dialFreqShift    float64 // Frequency shift control
	dialBands        float64 // Number of frequency bands

	// Bar mode dial values
	dialBarWidth  float64
	dialBarGap    float64
	dialMinHeight float64

	// Rising mode dial values
	dialGravity     float64
	dialBounce      float64
	dialTrailLength float64
	dialPeakHold    float64

	// Particles mode dial values
	dialParticleSpawn   float64
	dialParticleGravity float64
	dialParticleSpeed   float64
	dialParticleLife    float64
	dialParticleSize    float64

	// Circular mode dial values
	dialCircularRadius   float64
	dialCircularBarLen   float64
	dialCircularRotSpeed float64

	// Wave mode dial values
	dialWaveSpeed    float64
	dialWaveWidth    float64
	dialWaveDecay    float64
	dialWaveWobble   float64

	// Colors
	bgColor       color.RGBA
	lineColor     color.RGBA
	mirrorColor   color.RGBA
	centerColor   color.RGBA
	textColor     color.RGBA
	dropdownBg    color.RGBA
	dropdownHover color.RGBA
	sliderBg      color.RGBA
	sliderFill    color.RGBA
	sliderHandle  color.RGBA
}

func NewVisualizer(cfg *config.Config, configPath string) *Visualizer {
	// Initialize spectrum history buffer (for time-offset overlay)
	historySize := 30 // Store up to 30 frames of history
	history := make([][]float64, historySize)
	for i := range history {
		history[i] = make([]float64, cfg.Display.BarCount)
	}

	return &Visualizer{
		width:           defaultWidth,
		height:          defaultHeight,
		config:          cfg,
		configPath:      configPath,
		spectrum:        make([]float64, cfg.Display.BarCount),
		peaks:           make([]float64, cfg.Display.BarCount),
		spectrumHistory: history,
		numBands:        cfg.Display.BarCount,
		smoothCurves:    true, // Enable smooth curves by default

		// Default line style
		lineStyle: LineStyle{
			Thickness:   2.0,
			ColorPreset: 28, // Synthwave (Magenta line, deep sky blue mirror)
			GlowEnabled: true,
			GlowSize:    3.0,
		},

		// Default overlay settings
		overlay: OverlaySettings{
			Count:       1,     // Single line by default
			Offset:      30.0,  // 30 pixels between overlay lines
			FadeEnabled: true,  // Fade distant lines
			TimeOffset:  3,     // 3 frame delay between overlays
			SyncLines:   true,  // Sync lines by default to prevent overlap
		},

		// Damping (reactivity reduction)
		damping: 1.0, // Full reactivity by default

		// Beat detection
		beatDetector: BeatDetector{
			Enabled:       false, // Disabled by default
			energyHistory: make([]float64, 43), // ~43 frames at 60fps = ~0.7 seconds
			historySize:   43,
			threshold:     1.5,  // Beat threshold multiplier
			cooldown:      8,    // Minimum frames between beats
			decayRate:     0.15, // How fast beat intensity fades
		},

		// Visualization mode
		vizMode: 0, // 0 = normal line

		// Default isometric settings
		isometric: IsometricSettings{
			Enabled:       false,
			DepthLayers:   6,
			DepthSpacing:  25.0,
			Angle:         30.0,
			ScaleFactor:   0.92,
			DepthFade:     true,
			Rotation:      0.0,
			RotationSpeed: 0.5,  // Slow rotation, 0.5 degrees per frame
			AutoRotate:    true, // Auto-rotate enabled by default
			ForwardMotion: false,
			ForwardSpeed:  0.5,
			ForwardOffset: 0.0,
		},

		// Default bar visualizer settings
		barSettings: BarSettings{
			BarWidth:     0.8,   // 80% of spacing
			GapWidth:     2.0,   // 2 pixels between bars
			RoundedCaps:  true,  // Rounded tops
			Mirror:       true,  // Mirror below center
			Gradient:     true,  // Gradient coloring
			MinHeight:    2.0,   // Minimum 2px height
			GlowEnabled:  true,  // Glow enabled
			GlowSize:     2.0,   // Glow size
			Outline:      false, // No outline by default
			OutlineWidth: 1.0,
		},

		// Default rising line settings
		rising: RisingSettings{
			Gravity:      0.08,  // Gravity pull
			MaxRise:      0.6,   // Max rise speed
			Bounce:       0.2,   // Bounce factor
			TrailEnabled: true,  // Show trail
			TrailLength:  8,     // Trail segments
			TrailFade:    0.15,  // Trail fade rate
			Peaks:        true,  // Show peaks
			PeakHold:     30,    // Hold peak for ~0.5s
			PeakFallRate: 0.02,  // Slow peak fall
		},

		// Rising line physics state
		risingPositions:  make([]float64, cfg.Display.BarCount),
		risingVelocities: make([]float64, cfg.Display.BarCount),
		risingPeaks:      make([]float64, cfg.Display.BarCount),
		risingPeakHold:   make([]int, cfg.Display.BarCount),

		// Default particles settings
		particles: ParticleSettings{
			MaxParticles:  2000,
			SpawnRate:     2.0,
			Gravity:       0.15,
			InitialSpeed:  8.0,
			SpeedVariance: 3.0,
			Spread:        0.8,
			MinLife:       0.5,
			MaxLife:       2.0,
			SizeMin:       2.0,
			SizeMax:       6.0,
			FadeOut:       true,
			Glow:          true,
			Explosion:     false,
		},

		// Particles state
		particleList: make([]Particle, 0, 2000),

		// Default circular settings
		circular: CircularSettings{
			InnerRadius:   80.0,
			BarLength:     150.0,
			BarWidth:      4.0,
			Rotation:      0.0,
			RotationSpeed: 0.5,
			Mirror:        true,
			Dots:          false,
			DotSize:       8.0,
			Spiral:        false,
			Rainbow:       true,
		},

		// Default wave settings
		wave: WaveSettings{
			MaxRings:     30,
			RingSpeed:    4.0,
			RingWidth:    3.0,
			Decay:        0.02,
			SpawnRate:    0.3,
			ReactToBass:  true,
			Fill:         false,
			Wobble:       true,
			WobbleAmount: 0.3,
			ColorByAge:   true,
		},

		// Wave state
		rings: make([]Ring, 0, 30),

		// Settings menu
		settingsMenu: SettingsMenu{
			Visible:    false,
			X:          10,
			Y:          45,
			Width:      340,
			Height:     500,
			ActiveDial: -1,
		},

		// Initialize dial values
		dialThickness:    2.0,
		dialGlowSize:     3.0,
		dialOverlayOff:   30.0,
		dialOverlayCnt:   1.0,
		dialDepthLayers:  6.0,
		dialAngle:        30.0,
		dialSpacing:      25.0,
		dialRotation:     0.0,
		dialRotSpeed:     0.5,
		dialForwardSpeed: 0.0,  // 0 = no motion (center position)
		dialDamping:      1.0,  // Full reactivity
		dialFreqShift:    0.0,  // Center = no shift
		dialBands:        64.0, // Default bands

		// Bar mode dial values
		dialBarWidth:  0.8,
		dialBarGap:    2.0,
		dialMinHeight: 2.0,

		// Rising mode dial values
		dialGravity:     0.08,
		dialBounce:      0.2,
		dialTrailLength: 8.0,
		dialPeakHold:    30.0,

		// Particles mode dial values
		dialParticleSpawn:   2.0,
		dialParticleGravity: 0.15,
		dialParticleSpeed:   8.0,
		dialParticleLife:    2.0,
		dialParticleSize:    4.0,

		// Circular mode dial values
		dialCircularRadius:   80.0,
		dialCircularBarLen:   150.0,
		dialCircularRotSpeed: 0.5,

		// Wave mode dial values
		dialWaveSpeed:  4.0,
		dialWaveWidth:  3.0,
		dialWaveDecay:  0.02,
		dialWaveWobble: 0.3,

		// Colors matching Python version
		bgColor:       color.RGBA{10, 10, 15, 255},
		lineColor:     color.RGBA{0, 255, 200, 255},
		mirrorColor:   color.RGBA{0, 127, 100, 255},
		centerColor:   color.RGBA{40, 40, 40, 255},
		textColor:     color.RGBA{200, 200, 200, 255},
		dropdownBg:    color.RGBA{40, 40, 40, 255},
		dropdownHover: color.RGBA{60, 60, 60, 255},
		sliderBg:      color.RGBA{30, 30, 35, 255},
		sliderFill:    color.RGBA{0, 150, 120, 255},
		sliderHandle:  color.RGBA{0, 255, 200, 255},
	}
}

func (v *Visualizer) SetCapture(cap *audio.Capture) {
	v.capture = cap
}

func (v *Visualizer) SetProcessor(proc *fft.Processor) {
	v.processor = proc
}

func (v *Visualizer) SetDeviceList(list *audio.DeviceList) {
	v.deviceList = list
	// Combine capture and playback (loopback) devices for the dropdown
	v.devices = make([]DeviceEntry, 0)

	// Add capture devices (microphones)
	for _, dev := range list.CaptureDevices {
		v.devices = append(v.devices, DeviceEntry{
			Name:       dev.Name,
			IsLoopback: false,
			IsDefault:  dev.IsDefault,
		})
	}

	// Track where loopback devices start
	loopbackStartIdx := len(v.devices)

	// Add playback devices as loopback options (system audio capture)
	for _, dev := range list.PlaybackDevices {
		v.devices = append(v.devices, DeviceEntry{
			Name:       dev.Name,
			IsLoopback: true,
			IsDefault:  dev.IsDefault,
		})
	}

	// Default to first loopback (output) device, or default output device if available
	if len(list.PlaybackDevices) > 0 {
		// Try to find the default playback device
		v.selectedDevice = loopbackStartIdx // Default to first loopback
		for i, dev := range list.PlaybackDevices {
			if dev.IsDefault {
				v.selectedDevice = loopbackStartIdx + i
				break
			}
		}
	}
}

func (v *Visualizer) Update() error {
	// Handle input
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		return ebiten.Termination
	}

	// Toggle help
	if inpututil.IsKeyJustPressed(ebiten.KeyH) {
		v.showHelp = !v.showHelp
	}

	// Toggle fullscreen
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
		v.showStatus("Fullscreen: " + boolToOnOff(ebiten.IsFullscreen()))
	}

	// Save config
	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		if err := v.config.Save(v.configPath); err != nil {
			v.showStatus("Failed to save config")
		} else {
			v.showStatus("Config saved")
		}
	}

	// Reset
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		v.config.Reset()
		if v.processor != nil {
			v.processor.SetSensitivity(v.config.Sensitivity.Sensitivity)
			v.processor.SetSmoothing(
				v.config.Smoothing.AttackSpeed,
				v.config.Smoothing.DecaySpeed,
				v.config.Smoothing.RestDecay,
			)
			v.processor.SetNumBands(v.config.Display.BarCount)
		}
		v.vizMode = 0 // Reset to Line mode
		v.lineStyle = LineStyle{Thickness: 2.0, ColorPreset: 28, GlowEnabled: true, GlowSize: 3.0} // Synthwave
		v.overlay = OverlaySettings{Count: 1, Offset: 30.0, FadeEnabled: true, TimeOffset: 3, SyncLines: true}
		v.isometric = IsometricSettings{Enabled: false, DepthLayers: 6, DepthSpacing: 25.0, Angle: 30.0, ScaleFactor: 0.92, DepthFade: true, Rotation: 0, RotationSpeed: 0.5, AutoRotate: true, ForwardMotion: false, ForwardSpeed: 0.5, ForwardOffset: 0}
		v.damping = 1.0
		v.freqShift = 0
		v.dialFreqShift = 0
		v.numBands = v.config.Display.BarCount
		v.dialBands = float64(v.numBands)
		v.smoothCurves = true
		// Reset dial values
		v.dialThickness = 2.0
		v.dialGlowSize = 3.0
		v.dialOverlayOff = 30.0
		v.dialOverlayCnt = 1.0
		v.dialDepthLayers = 6.0
		v.dialAngle = 30.0
		v.dialSpacing = 25.0
		v.dialRotation = 0.0
		v.dialRotSpeed = 0.5
		v.dialForwardSpeed = 0.0
		v.dialDamping = 1.0
		v.resizeSpectrumBuffers(v.numBands)
		v.showStatus("Reset to defaults")
	}

	// Toggle smooth curves: B
	if inpututil.IsKeyJustPressed(ebiten.KeyB) {
		v.smoothCurves = !v.smoothCurves
		v.showStatus("Smooth curves: " + boolToOnOff(v.smoothCurves))
	}

	// Update isometric rotation (animate Y-axis rotation) if auto-rotate is enabled
	if v.vizMode == 4 && v.isometric.AutoRotate {
		v.isometric.Rotation += v.isometric.RotationSpeed
		if v.isometric.Rotation >= 360 {
			v.isometric.Rotation -= 360
		}
		if v.isometric.Rotation < 0 {
			v.isometric.Rotation += 360
		}
	}

	// Update forward motion offset
	if v.vizMode == 4 && v.isometric.ForwardMotion {
		v.isometric.ForwardOffset += v.isometric.ForwardSpeed
	}

	// Toggle style panel
	if inpututil.IsKeyJustPressed(ebiten.KeyL) {
		v.showStylePanel = !v.showStylePanel
	}

	// Toggle settings menu (collapsible dial menu)
	if inpututil.IsKeyJustPressed(ebiten.KeyM) {
		v.settingsMenu.Visible = !v.settingsMenu.Visible
		v.showStatus("Settings Menu: " + boolToOnOff(v.settingsMenu.Visible))
	}

	// Glow toggle: G
	if inpututil.IsKeyJustPressed(ebiten.KeyG) {
		v.lineStyle.GlowEnabled = !v.lineStyle.GlowEnabled
		v.showStatus("Glow: " + boolToOnOff(v.lineStyle.GlowEnabled))
	}

	// Toggle SyncLines: Y (makes all overlay lines react identically, preventing overlap)
	if inpututil.IsKeyJustPressed(ebiten.KeyY) {
		v.overlay.SyncLines = !v.overlay.SyncLines
		if v.overlay.SyncLines {
			v.showStatus("Overlay sync: ON (lines move together)")
		} else {
			v.showStatus("Overlay sync: OFF (lines trail)")
		}
	}

	// Handle dropdown
	v.handleDropdown()

	// Handle visualization mode dropdown
	v.handleVizModeDropdown()

	// Handle color picker dropdown
	v.handleColorDropdown()

	// Handle settings menu dials
	v.handleSettingsMenu()

	// Poll audio
	v.pollAudio()

	// Update status timer
	if v.statusTimer > 0 {
		v.statusTimer--
	}

	return nil
}

func (v *Visualizer) handleDropdown() {
	mx, my := ebiten.CursorPosition()
	dropdownX := v.width - 320
	dropdownY := 10
	dropdownW := 300
	itemH := 28

	// Check if clicked on dropdown button
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if mx >= dropdownX && mx <= dropdownX+dropdownW && my >= dropdownY && my <= dropdownY+itemH {
			v.showDropdown = !v.showDropdown
			return
		}

		// Check if clicked on dropdown item
		if v.showDropdown {
			for i := range v.devices {
				itemY := dropdownY + itemH + i*itemH
				if mx >= dropdownX && mx <= dropdownX+dropdownW && my >= itemY && my <= itemY+itemH {
					v.selectDevice(i)
					v.showDropdown = false
					return
				}
			}
			// Clicked elsewhere, close dropdown
			v.showDropdown = false
		}
	}

	// Update hover state
	if v.showDropdown {
		v.hoverDevice = -1
		for i := range v.devices {
			itemY := dropdownY + itemH + i*itemH
			if mx >= dropdownX && mx <= dropdownX+dropdownW && my >= itemY && my <= itemY+itemH {
				v.hoverDevice = i
				break
			}
		}
	}
}

func (v *Visualizer) handleVizModeDropdown() {
	mx, my := ebiten.CursorPosition()
	dropdownX := 10
	dropdownY := 10
	dropdownW := 120
	itemH := 24

	// Check if clicked on dropdown button
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if mx >= dropdownX && mx <= dropdownX+dropdownW && my >= dropdownY && my <= dropdownY+itemH {
			v.showVizDropdown = !v.showVizDropdown
			v.showDropdown = false // Close device dropdown if open
			return
		}

		// Check if clicked on dropdown item
		if v.showVizDropdown {
			for i := range vizModeNames {
				itemY := dropdownY + itemH + i*itemH
				if mx >= dropdownX && mx <= dropdownX+dropdownW && my >= itemY && my <= itemY+itemH {
					v.vizMode = i
					v.showVizDropdown = false
					v.showStatus("Viz mode: " + vizModeNames[i])
					return
				}
			}
			// Clicked elsewhere, close dropdown
			v.showVizDropdown = false
		}
	}

	// Update hover state
	if v.showVizDropdown {
		v.hoverVizMode = -1
		for i := range vizModeNames {
			itemY := dropdownY + itemH + i*itemH
			if mx >= dropdownX && mx <= dropdownX+dropdownW && my >= itemY && my <= itemY+itemH {
				v.hoverVizMode = i
				break
			}
		}
	}
}

func (v *Visualizer) handleColorDropdown() {
	mx, my := ebiten.CursorPosition()
	dropdownX := 140 // Position next to viz mode dropdown
	dropdownY := 10
	dropdownW := 160
	itemH := 20
	maxVisible := 12 // Maximum visible items before scrolling

	// Check if clicked on dropdown button
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if mx >= dropdownX && mx <= dropdownX+dropdownW && my >= dropdownY && my <= dropdownY+24 {
			v.showColorDropdown = !v.showColorDropdown
			v.showVizDropdown = false // Close viz dropdown if open
			v.showDropdown = false    // Close device dropdown if open
			return
		}

		// Check if clicked on dropdown item
		if v.showColorDropdown {
			visibleCount := len(colorPresets)
			if visibleCount > maxVisible {
				visibleCount = maxVisible
			}
			for i := 0; i < visibleCount; i++ {
				actualIndex := i + v.colorDropdownScroll
				if actualIndex >= len(colorPresets) {
					break
				}
				itemY := dropdownY + 24 + i*itemH
				if mx >= dropdownX && mx <= dropdownX+dropdownW && my >= itemY && my <= itemY+itemH {
					v.lineStyle.ColorPreset = actualIndex
					v.showColorDropdown = false
					v.showStatus("Color: " + colorPresets[actualIndex].Name)
					return
				}
			}
			// Clicked elsewhere, close dropdown
			v.showColorDropdown = false
		}
	}

	// Handle scroll wheel for color dropdown
	if v.showColorDropdown {
		_, scrollY := ebiten.Wheel()
		if scrollY != 0 {
			v.colorDropdownScroll -= int(scrollY)
			maxScroll := len(colorPresets) - maxVisible
			if maxScroll < 0 {
				maxScroll = 0
			}
			if v.colorDropdownScroll < 0 {
				v.colorDropdownScroll = 0
			}
			if v.colorDropdownScroll > maxScroll {
				v.colorDropdownScroll = maxScroll
			}
		}
	}

	// Update hover state
	if v.showColorDropdown {
		v.hoverColorPreset = -1
		visibleCount := len(colorPresets)
		if visibleCount > maxVisible {
			visibleCount = maxVisible
		}
		for i := 0; i < visibleCount; i++ {
			actualIndex := i + v.colorDropdownScroll
			if actualIndex >= len(colorPresets) {
				break
			}
			itemY := dropdownY + 24 + i*itemH
			if mx >= dropdownX && mx <= dropdownX+dropdownW && my >= itemY && my <= itemY+itemH {
				v.hoverColorPreset = actualIndex
				break
			}
		}
	}
}

func (v *Visualizer) handleSliders() {
	mx, my := ebiten.CursorPosition()

	// === Frequency Shift Slider (bottom) ===
	shiftSliderX := 100
	shiftSliderY := v.height - 40
	shiftSliderW := (v.width - 220) / 2
	shiftSliderH := 20

	// Check if mouse is pressed/held on freq shift slider
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if mx >= shiftSliderX && mx <= shiftSliderX+shiftSliderW && my >= shiftSliderY-10 && my <= shiftSliderY+shiftSliderH+10 {
			v.draggingSlider = true
		}
	} else if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		v.draggingSlider = false
	}

	// Update freq shift slider value while dragging
	if v.draggingSlider {
		relX := float64(mx-shiftSliderX) / float64(shiftSliderW)
		v.freqShift = (relX - 0.5) * 2.0
		v.freqShift = math.Max(-1.0, math.Min(1.0, v.freqShift))
	}

	// Right-click to reset freq shift slider
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		if mx >= shiftSliderX && mx <= shiftSliderX+shiftSliderW && my >= shiftSliderY-10 && my <= shiftSliderY+shiftSliderH+10 {
			v.freqShift = 0
			v.showStatus("Frequency shift reset")
		}
	}

	// === Frequency Bands Slider (right side) ===
	bandsSliderX := shiftSliderX + shiftSliderW + 20
	bandsSliderY := v.height - 40
	bandsSliderW := (v.width - 220) / 2
	bandsSliderH := 20

	// Check if mouse is pressed/held on bands slider
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if mx >= bandsSliderX && mx <= bandsSliderX+bandsSliderW && my >= bandsSliderY-10 && my <= bandsSliderY+bandsSliderH+10 {
			v.draggingBandsSlider = true
		}
	} else if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		v.draggingBandsSlider = false
	}

	// Update bands slider value while dragging
	if v.draggingBandsSlider {
		relX := float64(mx-bandsSliderX) / float64(bandsSliderW)
		// Map to 16-256 bands (logarithmic scale for better feel)
		minBands := 16.0
		maxBands := 256.0
		logMin := math.Log(minBands)
		logMax := math.Log(maxBands)
		newBands := int(math.Exp(logMin + relX*(logMax-logMin)))
		newBands = int(math.Max(minBands, math.Min(maxBands, float64(newBands))))

		if newBands != v.numBands {
			v.numBands = newBands
			if v.processor != nil {
				v.processor.SetNumBands(newBands)
			}
			v.resizeSpectrumBuffers(newBands)
			v.showStatus(fmt.Sprintf("Bands: %d", newBands))
		}
	}

	// Right-click to reset bands slider
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		if mx >= bandsSliderX && mx <= bandsSliderX+bandsSliderW && my >= bandsSliderY-10 && my <= bandsSliderY+bandsSliderH+10 {
			v.numBands = 64
			if v.processor != nil {
				v.processor.SetNumBands(64)
			}
			v.resizeSpectrumBuffers(64)
			v.showStatus("Bands reset to 64")
		}
	}
}

// handleSettingsMenu handles mouse interaction with the settings menu dials
func (v *Visualizer) handleSettingsMenu() {
	if !v.settingsMenu.Visible {
		return
	}

	mx, my := ebiten.CursorPosition()

	// Define the dial controls
	dials := v.getDialControls()

	dialStartY := v.settingsMenu.Y + 50
	dialHeight := 36
	sliderX := v.settingsMenu.X + 100
	sliderW := 180

	// Handle mouse release
	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		v.settingsMenu.ActiveDial = -1
	}

	// Handle mouse press on dials
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		for i, dial := range dials {
			dialY := dialStartY + i*dialHeight
			// Check if click is on slider area
			if mx >= sliderX && mx <= sliderX+sliderW && my >= dialY && my <= dialY+24 {
				v.settingsMenu.ActiveDial = i
				v.settingsMenu.DragStartY = mx
				v.settingsMenu.DragStartVal = *dial.Value
				break
			}
		}
	}

	// Handle dragging
	if v.settingsMenu.ActiveDial >= 0 && v.settingsMenu.ActiveDial < len(dials) {
		dial := dials[v.settingsMenu.ActiveDial]

		// Calculate new value based on horizontal drag
		relX := float64(mx-sliderX) / float64(sliderW)
		relX = math.Max(0, math.Min(1, relX))
		newVal := dial.Min + relX*(dial.Max-dial.Min)

		// Snap to step if defined
		if dial.Step > 0 {
			newVal = math.Round(newVal/dial.Step) * dial.Step
		}

		*dial.Value = newVal
		v.applyDialValue(v.settingsMenu.ActiveDial)
	}

	// Handle right-click to reset dial to default
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		for i := range dials {
			dialY := dialStartY + i*dialHeight
			if mx >= sliderX && mx <= sliderX+sliderW && my >= dialY && my <= dialY+24 {
				v.resetDialToDefault(i)
				break
			}
		}
	}
}

// getDialControls returns the list of dial controls for the settings menu
// Filters dials based on current visualization mode
func (v *Visualizer) getDialControls() []DialControl {
	// Common dials for all modes
	commonDials := []DialControl{
		{"sensitivity", "Sensitivity", &v.config.Sensitivity.Sensitivity, 0.1, 5.0, 0.1, "%.1f"},
		{"smoothing", "Smoothing", &v.config.Smoothing.AttackSpeed, 0.1, 1.0, 0.05, "%.2f"},
		{"damping", "Damping", &v.dialDamping, 0.1, 1.0, 0.05, "%.2f"},
		{"freqshift", "Freq Shift", &v.dialFreqShift, -100.0, 100.0, 5.0, "%+.0f%%"},
		{"bands", "Bands", &v.dialBands, 16.0, 256.0, 8.0, "%.0f"},
	}

	// Line mode (0): Add line-specific controls
	if v.vizMode == 0 {
		return append(commonDials, []DialControl{
			{"thickness", "Thickness", &v.dialThickness, 0.5, 8.0, 0.5, "%.1f"},
			{"glowsize", "Glow Size", &v.dialGlowSize, 1.0, 8.0, 0.5, "%.1f"},
			{"overlays", "Overlays", &v.dialOverlayCnt, 1.0, 50.0, 1.0, "%.0f"},
			{"overlayoff", "Overlay Gap", &v.dialOverlayOff, 2.0, 60.0, 2.0, "%.0f"},
		}...)
	}

	// Bars mode (1): Add bar-specific controls
	if v.vizMode == 1 {
		return append(commonDials, []DialControl{
			{"barwidth", "Bar Width", &v.dialBarWidth, 0.3, 1.0, 0.05, "%.0f%%"},
			{"bargap", "Bar Gap", &v.dialBarGap, 0.0, 10.0, 0.5, "%.1f"},
			{"minheight", "Min Height", &v.dialMinHeight, 0.0, 20.0, 1.0, "%.0f"},
			{"glowsize", "Glow Size", &v.dialGlowSize, 0.0, 8.0, 0.5, "%.1f"},
		}...)
	}

	// Rising mode (2): Add rising-specific controls
	if v.vizMode == 2 {
		return append(commonDials, []DialControl{
			{"thickness", "Thickness", &v.dialThickness, 0.5, 8.0, 0.5, "%.1f"},
			{"gravity", "Gravity", &v.dialGravity, 0.01, 0.3, 0.01, "%.2f"},
			{"bounce", "Bounce", &v.dialBounce, 0.0, 0.5, 0.05, "%.2f"},
			{"traillength", "Trail", &v.dialTrailLength, 0.0, 20.0, 1.0, "%.0f"},
			{"peakhold", "Peak Hold", &v.dialPeakHold, 0.0, 120.0, 5.0, "%.0f"},
		}...)
	}

	// Spiral mode (3): Add spiral-specific rotations control
	if v.vizMode == 3 {
		return append(commonDials, []DialControl{
			{"thickness", "Thickness", &v.dialThickness, 0.5, 8.0, 0.5, "%.1f"},
			{"glowsize", "Glow Size", &v.dialGlowSize, 1.0, 8.0, 0.5, "%.1f"},
			{"overlays", "Rotations", &v.dialOverlayCnt, 1.0, 10.0, 1.0, "%.0f"},
		}...)
	}

	// Isometric mode (4): Add isometric-specific controls
	if v.vizMode == 4 {
		return append(commonDials, []DialControl{
			{"thickness", "Thickness", &v.dialThickness, 0.5, 8.0, 0.5, "%.1f"},
			{"glowsize", "Glow Size", &v.dialGlowSize, 1.0, 8.0, 0.5, "%.1f"},
			{"overlays", "Overlays", &v.dialOverlayCnt, 1.0, 50.0, 1.0, "%.0f"},
			{"overlayoff", "Overlay Gap", &v.dialOverlayOff, 2.0, 60.0, 2.0, "%.0f"},
			{"depthlayers", "Depth Layers", &v.dialDepthLayers, 2.0, 30.0, 1.0, "%.0f"},
			{"isoangle", "Iso Angle", &v.dialAngle, 15.0, 60.0, 5.0, "%.0f°"},
			{"isospacing", "Iso Spacing", &v.dialSpacing, 10.0, 60.0, 5.0, "%.0f"},
			{"rotation", "Rotation", &v.dialRotation, -180.0, 180.0, 5.0, "%.0f°"},
			{"rotspeed", "Rot Speed", &v.dialRotSpeed, -2.0, 2.0, 0.1, "%.1f"},
			{"forward", "Forward", &v.dialForwardSpeed, -2.0, 2.0, 0.1, "%.1f"},
		}...)
	}

	// Particles mode (5): Add particle-specific controls
	if v.vizMode == 5 {
		return append(commonDials, []DialControl{
			{"particlespawn", "Spawn Rate", &v.dialParticleSpawn, 0.5, 10.0, 0.5, "%.1f"},
			{"particlegravity", "Gravity", &v.dialParticleGravity, 0.0, 0.5, 0.02, "%.2f"},
			{"particlespeed", "Speed", &v.dialParticleSpeed, 2.0, 20.0, 1.0, "%.0f"},
			{"particlelife", "Lifetime", &v.dialParticleLife, 0.5, 5.0, 0.25, "%.2f"},
			{"particlesize", "Size", &v.dialParticleSize, 1.0, 12.0, 0.5, "%.1f"},
		}...)
	}

	// Circular mode (6): Add circular-specific controls
	if v.vizMode == 6 {
		return append(commonDials, []DialControl{
			{"circularradius", "Inner Radius", &v.dialCircularRadius, 20.0, 200.0, 10.0, "%.0f"},
			{"circularbarlen", "Bar Length", &v.dialCircularBarLen, 50.0, 300.0, 10.0, "%.0f"},
			{"circularrotspeed", "Rotation", &v.dialCircularRotSpeed, -2.0, 2.0, 0.1, "%.1f"},
			{"thickness", "Bar Width", &v.dialThickness, 1.0, 12.0, 0.5, "%.1f"},
		}...)
	}

	// Wave mode (7): Add wave-specific controls
	if v.vizMode == 7 {
		return append(commonDials, []DialControl{
			{"wavespeed", "Speed", &v.dialWaveSpeed, 1.0, 15.0, 0.5, "%.1f"},
			{"wavewidth", "Width", &v.dialWaveWidth, 1.0, 10.0, 0.5, "%.1f"},
			{"wavedecay", "Decay", &v.dialWaveDecay, 0.005, 0.1, 0.005, "%.3f"},
			{"wavewobble", "Wobble", &v.dialWaveWobble, 0.0, 1.0, 0.1, "%.1f"},
		}...)
	}

	return commonDials
}

// applyDialValue applies the dial value to the actual settings using dial key
func (v *Visualizer) applyDialValue(dialIndex int) {
	dials := v.getDialControls()
	if dialIndex < 0 || dialIndex >= len(dials) {
		return
	}
	dial := dials[dialIndex]

	switch dial.Key {
	case "sensitivity":
		if v.processor != nil {
			v.processor.SetSensitivity(v.config.Sensitivity.Sensitivity)
		}
	case "smoothing":
		if v.processor != nil {
			v.processor.SetSmoothing(v.config.Smoothing.AttackSpeed, v.config.Smoothing.DecaySpeed, v.config.Smoothing.RestDecay)
		}
	case "damping":
		v.damping = v.dialDamping
	case "freqshift":
		v.freqShift = v.dialFreqShift / 100.0
	case "bands":
		newBands := int(v.dialBands)
		if newBands != v.numBands {
			v.numBands = newBands
			if v.processor != nil {
				v.processor.SetNumBands(newBands)
			}
			v.resizeSpectrumBuffers(newBands)
		}
	case "thickness":
		v.lineStyle.Thickness = float32(v.dialThickness)
	case "glowsize":
		v.lineStyle.GlowSize = float32(v.dialGlowSize)
	case "overlays":
		v.overlay.Count = int(v.dialOverlayCnt)
	case "overlayoff":
		v.overlay.Offset = float32(v.dialOverlayOff)
	case "depthlayers":
		v.isometric.DepthLayers = int(v.dialDepthLayers)
	case "isoangle":
		v.isometric.Angle = float32(v.dialAngle)
	case "isospacing":
		v.isometric.DepthSpacing = float32(v.dialSpacing)
	case "rotation":
		v.isometric.Rotation = float32(v.dialRotation)
		v.isometric.AutoRotate = false
	case "rotspeed":
		v.isometric.RotationSpeed = float32(v.dialRotSpeed)
		if v.dialRotSpeed != 0 {
			v.isometric.AutoRotate = true
		}
	case "forward":
		v.isometric.ForwardSpeed = float32(v.dialForwardSpeed)
		v.isometric.ForwardMotion = v.dialForwardSpeed != 0
	// Bar mode settings
	case "barwidth":
		v.barSettings.BarWidth = float32(v.dialBarWidth)
	case "bargap":
		v.barSettings.GapWidth = float32(v.dialBarGap)
	case "minheight":
		v.barSettings.MinHeight = float32(v.dialMinHeight)
	// Rising mode settings
	case "gravity":
		v.rising.Gravity = float32(v.dialGravity)
	case "bounce":
		v.rising.Bounce = float32(v.dialBounce)
	case "traillength":
		v.rising.TrailLength = int(v.dialTrailLength)
	case "peakhold":
		v.rising.PeakHold = int(v.dialPeakHold)
	// Particles mode settings
	case "particlespawn":
		v.particles.SpawnRate = float32(v.dialParticleSpawn)
	case "particlegravity":
		v.particles.Gravity = float32(v.dialParticleGravity)
	case "particlespeed":
		v.particles.InitialSpeed = float32(v.dialParticleSpeed)
	case "particlelife":
		v.particles.MaxLife = float32(v.dialParticleLife)
	case "particlesize":
		v.particles.SizeMax = float32(v.dialParticleSize)
		v.particles.SizeMin = float32(v.dialParticleSize) * 0.5
	// Circular mode settings
	case "circularradius":
		v.circular.InnerRadius = float32(v.dialCircularRadius)
	case "circularbarlen":
		v.circular.BarLength = float32(v.dialCircularBarLen)
	case "circularrotspeed":
		v.circular.RotationSpeed = float32(v.dialCircularRotSpeed)
	// Wave mode settings
	case "wavespeed":
		v.wave.RingSpeed = float32(v.dialWaveSpeed)
	case "wavewidth":
		v.wave.RingWidth = float32(v.dialWaveWidth)
	case "wavedecay":
		v.wave.Decay = float32(v.dialWaveDecay)
	case "wavewobble":
		v.wave.WobbleAmount = float32(v.dialWaveWobble)
		v.wave.Wobble = v.dialWaveWobble > 0
	}
}

// resetDialToDefault resets a dial to its default value
func (v *Visualizer) resetDialToDefault(dialIndex int) {
	dials := v.getDialControls()
	if dialIndex < 0 || dialIndex >= len(dials) {
		return
	}
	dial := dials[dialIndex]

	switch dial.Key {
	case "sensitivity":
		v.config.Sensitivity.Sensitivity = 1.0
		if v.processor != nil {
			v.processor.SetSensitivity(1.0)
		}
	case "smoothing":
		v.config.Smoothing.AttackSpeed = 0.8
		if v.processor != nil {
			v.processor.SetSmoothing(0.8, v.config.Smoothing.DecaySpeed, v.config.Smoothing.RestDecay)
		}
	case "damping":
		v.dialDamping = 1.0
		v.damping = 1.0
	case "freqshift":
		v.dialFreqShift = 0.0
		v.freqShift = 0.0
	case "bands":
		v.dialBands = 64.0
		v.numBands = 64
		if v.processor != nil {
			v.processor.SetNumBands(64)
		}
		v.resizeSpectrumBuffers(64)
	case "thickness":
		v.dialThickness = 2.0
		v.lineStyle.Thickness = 2.0
	case "glowsize":
		v.dialGlowSize = 3.0
		v.lineStyle.GlowSize = 3.0
	case "overlays":
		v.dialOverlayCnt = 1.0
		v.overlay.Count = 1
	case "overlayoff":
		v.dialOverlayOff = 30.0
		v.overlay.Offset = 30.0
	case "depthlayers":
		v.dialDepthLayers = 6.0
		v.isometric.DepthLayers = 6
	case "isoangle":
		v.dialAngle = 30.0
		v.isometric.Angle = 30.0
	case "isospacing":
		v.dialSpacing = 25.0
		v.isometric.DepthSpacing = 25.0
	case "rotation":
		v.dialRotation = 0.0
		v.isometric.Rotation = 0.0
	case "rotspeed":
		v.dialRotSpeed = 0.5
		v.isometric.RotationSpeed = 0.5
		v.isometric.AutoRotate = true
	case "forward":
		v.dialForwardSpeed = 0.0
		v.isometric.ForwardSpeed = 0.0
		v.isometric.ForwardMotion = false
	// Bar mode resets
	case "barwidth":
		v.dialBarWidth = 0.8
		v.barSettings.BarWidth = 0.8
	case "bargap":
		v.dialBarGap = 2.0
		v.barSettings.GapWidth = 2.0
	case "minheight":
		v.dialMinHeight = 2.0
		v.barSettings.MinHeight = 2.0
	// Rising mode resets
	case "gravity":
		v.dialGravity = 0.08
		v.rising.Gravity = 0.08
	case "bounce":
		v.dialBounce = 0.2
		v.rising.Bounce = 0.2
	case "traillength":
		v.dialTrailLength = 8.0
		v.rising.TrailLength = 8
	case "peakhold":
		v.dialPeakHold = 30.0
		v.rising.PeakHold = 30
	// Particles mode resets
	case "particlespawn":
		v.dialParticleSpawn = 2.0
		v.particles.SpawnRate = 2.0
	case "particlegravity":
		v.dialParticleGravity = 0.15
		v.particles.Gravity = 0.15
	case "particlespeed":
		v.dialParticleSpeed = 8.0
		v.particles.InitialSpeed = 8.0
	case "particlelife":
		v.dialParticleLife = 2.0
		v.particles.MaxLife = 2.0
	case "particlesize":
		v.dialParticleSize = 4.0
		v.particles.SizeMax = 6.0
		v.particles.SizeMin = 2.0
	// Circular mode resets
	case "circularradius":
		v.dialCircularRadius = 80.0
		v.circular.InnerRadius = 80.0
	case "circularbarlen":
		v.dialCircularBarLen = 150.0
		v.circular.BarLength = 150.0
	case "circularrotspeed":
		v.dialCircularRotSpeed = 0.5
		v.circular.RotationSpeed = 0.5
	// Wave mode resets
	case "wavespeed":
		v.dialWaveSpeed = 4.0
		v.wave.RingSpeed = 4.0
	case "wavewidth":
		v.dialWaveWidth = 3.0
		v.wave.RingWidth = 3.0
	case "wavedecay":
		v.dialWaveDecay = 0.02
		v.wave.Decay = 0.02
	case "wavewobble":
		v.dialWaveWobble = 0.3
		v.wave.WobbleAmount = 0.3
		v.wave.Wobble = true
	}
	v.showStatus("Reset to default")
}

// resizeSpectrumBuffers resizes all spectrum-related buffers to match new band count
func (v *Visualizer) resizeSpectrumBuffers(numBands int) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.spectrum = make([]float64, numBands)
	v.peaks = make([]float64, numBands)

	// Resize history buffer
	for i := range v.spectrumHistory {
		v.spectrumHistory[i] = make([]float64, numBands)
	}

	// Resize rising line buffers
	v.risingPositions = make([]float64, numBands)
	v.risingVelocities = make([]float64, numBands)
	v.risingPeaks = make([]float64, numBands)
	v.risingPeakHold = make([]int, numBands)
}

func (v *Visualizer) selectDevice(index int) {
	if index < 0 || index >= len(v.devices) {
		return
	}

	device := v.devices[index]
	v.selectedDevice = index

	if v.capture != nil {
		if err := v.capture.SetDeviceWithLoopback(device.Name, device.IsLoopback); err != nil {
			v.showStatus("Failed to switch device")
		} else {
			prefix := ""
			if device.IsLoopback {
				prefix = "[Loopback] "
			}
			v.showStatus("Device: " + prefix + truncate(device.Name, 25))
		}
	}
}

func (v *Visualizer) pollAudio() {
	if v.capture == nil {
		return
	}

	select {
	case data, ok := <-v.capture.DataChannel():
		if ok && len(data) > 0 && v.processor != nil {
			v.mu.Lock()
			v.spectrum = v.processor.Process(data)
			v.peaks = v.processor.GetPeaks()

			// Store in history buffer for time-offset overlay
			if len(v.spectrumHistory) > 0 {
				copy(v.spectrumHistory[v.historyIndex], v.spectrum)
				v.historyIndex = (v.historyIndex + 1) % len(v.spectrumHistory)
			}

			// Beat detection - analyze bass frequencies
			if v.beatDetector.Enabled && len(v.spectrum) > 0 {
				v.detectBeat(v.spectrum)
			}

			v.mu.Unlock()
		}
	default:
	}

	// Update beat intensity decay
	if v.beatDetector.beatIntensity > 0 {
		v.beatDetector.beatIntensity -= v.beatDetector.decayRate
		if v.beatDetector.beatIntensity < 0 {
			v.beatDetector.beatIntensity = 0
		}
	}

	// Decrement beat cooldown
	if v.beatDetector.cooldownCounter > 0 {
		v.beatDetector.cooldownCounter--
	}

	// Increment frame counter for rainbow animation
	v.frameCount++
}

// detectBeat analyzes the spectrum to detect beats based on bass energy
func (v *Visualizer) detectBeat(spectrum []float64) {
	// Calculate bass energy (first 1/8 of spectrum = low frequencies)
	bassEnd := len(spectrum) / 8
	if bassEnd < 1 {
		bassEnd = 1
	}

	var bassEnergy float64
	for i := 0; i < bassEnd; i++ {
		bassEnergy += spectrum[i] * spectrum[i] // Use squared for energy
	}
	bassEnergy /= float64(bassEnd)

	// Add to history
	v.beatDetector.energyHistory[v.beatDetector.historyIndex] = bassEnergy
	v.beatDetector.historyIndex = (v.beatDetector.historyIndex + 1) % v.beatDetector.historySize

	// Calculate average energy from history
	var avgEnergy float64
	for _, e := range v.beatDetector.energyHistory {
		avgEnergy += e
	}
	avgEnergy /= float64(v.beatDetector.historySize)

	// Detect beat: current energy significantly higher than average
	v.beatDetector.beatDetected = false
	if v.beatDetector.cooldownCounter <= 0 && avgEnergy > 0.001 {
		if bassEnergy > avgEnergy*v.beatDetector.threshold {
			v.beatDetector.beatDetected = true
			v.beatDetector.beatIntensity = 1.0
			v.beatDetector.cooldownCounter = v.beatDetector.cooldown
		}
	}
}

func (v *Visualizer) Draw(screen *ebiten.Image) {
	// Clear background
	screen.Fill(v.bgColor)

	v.width, v.height = screen.Bounds().Dx(), screen.Bounds().Dy()
	centerY := float32(v.height) / 2

	// Draw center line
	vector.StrokeLine(screen, 0, centerY, float32(v.width), centerY, 1, v.centerColor, false)

	// Get current color preset
	preset := colorPresets[v.lineStyle.ColorPreset]

	// Draw spectrum with overlays
	v.mu.RLock()
	spectrum := make([]float64, len(v.spectrum))
	copy(spectrum, v.spectrum)

	// Apply damping to reduce reactivity
	if v.damping < 1.0 {
		for i := range spectrum {
			spectrum[i] *= v.damping
		}
	}

	// Get history spectrums for overlays
	historySpectrums := make([][]float64, v.overlay.Count)
	for i := 0; i < v.overlay.Count; i++ {
		historySpectrums[i] = make([]float64, len(v.spectrum))
		histIdx := (v.historyIndex - 1 - i*v.overlay.TimeOffset + len(v.spectrumHistory)) % len(v.spectrumHistory)
		if histIdx >= 0 && histIdx < len(v.spectrumHistory) {
			copy(historySpectrums[i], v.spectrumHistory[histIdx])
		}
		// Apply damping to history spectrums as well
		if v.damping < 1.0 {
			for j := range historySpectrums[i] {
				historySpectrums[i][j] *= v.damping
			}
		}
	}
	v.mu.RUnlock()

	if len(spectrum) > 1 {
		switch v.vizMode {
		case 1: // Bars visualization mode
			v.drawBars(screen, spectrum, preset, centerY)
		case 2: // Rising line visualization mode
			v.drawRising(screen, spectrum, preset, centerY)
		case 3: // Spiral visualization mode
			v.drawSpiral(screen, spectrum, preset)
		case 4: // Isometric 3D rendering
			v.drawIsometric(screen, spectrum, historySpectrums, preset, centerY)
		case 5: // Particles visualization mode
			v.drawParticles(screen, spectrum, preset, centerY)
		case 6: // Circular visualization mode
			v.drawCircular(screen, spectrum, preset)
		case 7: // Wave visualization mode
			v.drawWave(screen, spectrum, preset)
		default: // Line mode (0) - Standard 2D rendering with overlays
			for i := v.overlay.Count - 1; i >= 0; i-- {
				// Determine which spectrum to draw
				// If SyncLines is enabled, all lines use the same current spectrum (no overlap)
				spectrumToDraw := spectrum
				if !v.overlay.SyncLines && i > 0 && i < len(historySpectrums) {
					spectrumToDraw = historySpectrums[i]
				}

				// Calculate offset for this overlay line
				yOffset := float32(i) * v.overlay.Offset

				// Calculate alpha for fading effect
				alpha := uint8(255)
				if v.overlay.FadeEnabled && v.overlay.Count > 1 {
					alpha = uint8(255 - (200 * i / v.overlay.Count))
				}

				// Get colors for this overlay (glow now matches line color dynamically)
				lineColor, mirrorColor, glowColor := v.getOverlayColors(preset, i, alpha)

				// Draw glow first (behind the line)
				if v.lineStyle.GlowEnabled && i == 0 {
					v.drawWaveformStyled(screen, spectrumToDraw, centerY-yOffset, false, v.lineStyle.Thickness*v.lineStyle.GlowSize, glowColor)
					v.drawWaveformStyled(screen, spectrumToDraw, centerY+yOffset, true, v.lineStyle.Thickness*v.lineStyle.GlowSize, glowColor)
				}

				// Draw the main line
				v.drawWaveformStyled(screen, spectrumToDraw, centerY-yOffset, false, v.lineStyle.Thickness, lineColor)
				v.drawWaveformStyled(screen, spectrumToDraw, centerY+yOffset, true, v.lineStyle.Thickness*0.75, mirrorColor)
			}
		}
	}

	// Draw UI
	v.drawUI(screen)
}

// drawBars renders the spectrum as vertical bars
func (v *Visualizer) drawBars(screen *ebiten.Image, spectrum []float64, preset ColorPreset, centerY float32) {
	numBars := len(spectrum)
	if numBars < 1 {
		return
	}

	// Calculate bar dimensions
	totalWidth := float32(v.width)
	barSpacing := totalWidth / float32(numBars)
	barWidth := barSpacing * v.barSettings.BarWidth
	gap := v.barSettings.GapWidth

	// Adjust bar width to account for gap
	if barWidth > barSpacing-gap {
		barWidth = barSpacing - gap
	}

	// Get colors
	lineColor, mirrorColor, glowColor := v.getOverlayColors(preset, 0, 255)

	// Apply frequency shift
	shiftAmount := int(v.freqShift * float64(numBars) * 0.5)

	maxHeight := float32(v.height) * 0.45 // Max bar height (from center to edge)

	for i := 0; i < numBars; i++ {
		// Get spectrum value with shift
		shiftedIdx := (i + shiftAmount + numBars) % numBars
		amp := float32(spectrum[shiftedIdx])

		// Calculate bar height
		barHeight := amp * maxHeight
		if barHeight < v.barSettings.MinHeight {
			barHeight = v.barSettings.MinHeight
		}

		// Calculate bar position
		barX := float32(i)*barSpacing + (barSpacing-barWidth)/2

		// Calculate gradient color based on amplitude if enabled
		barColor := lineColor
		if v.barSettings.Gradient {
			// Interpolate from mirror color (low) to line color (high)
			t := amp
			if t > 1 {
				t = 1
			}
			barColor = color.RGBA{
				R: uint8(float32(mirrorColor.R) + t*float32(lineColor.R-mirrorColor.R)),
				G: uint8(float32(mirrorColor.G) + t*float32(lineColor.G-mirrorColor.G)),
				B: uint8(float32(mirrorColor.B) + t*float32(lineColor.B-mirrorColor.B)),
				A: 255,
			}
		}

		// Draw glow behind bar if enabled
		if v.barSettings.GlowEnabled && v.lineStyle.GlowSize > 0 {
			glowExpand := v.lineStyle.GlowSize
			glowBarColor := glowColor
			glowBarColor.A = 60

			// Upper glow
			vector.DrawFilledRect(screen, barX-glowExpand, centerY-barHeight-glowExpand,
				barWidth+glowExpand*2, barHeight+glowExpand, glowBarColor, false)

			// Lower glow (mirror)
			if v.barSettings.Mirror {
				vector.DrawFilledRect(screen, barX-glowExpand, centerY,
					barWidth+glowExpand*2, barHeight+glowExpand, glowBarColor, false)
			}
		}

		// Draw upper bar (above center line)
		if v.barSettings.RoundedCaps {
			// Draw rounded cap (circle at top)
			capRadius := barWidth / 2
			capY := centerY - barHeight
			vector.DrawFilledCircle(screen, barX+barWidth/2, capY, capRadius, barColor, true)
			// Draw rectangle body below cap
			if barHeight > capRadius {
				vector.DrawFilledRect(screen, barX, capY, barWidth, barHeight-capRadius+1, barColor, false)
			}
		} else {
			vector.DrawFilledRect(screen, barX, centerY-barHeight, barWidth, barHeight, barColor, false)
		}

		// Draw outline if enabled
		if v.barSettings.Outline {
			outlineColor := color.RGBA{255, 255, 255, 100}
			vector.StrokeRect(screen, barX, centerY-barHeight, barWidth, barHeight, v.barSettings.OutlineWidth, outlineColor, false)
		}

		// Draw lower bar (mirror below center line) if enabled
		if v.barSettings.Mirror {
			mirrorBarColor := mirrorColor
			if v.barSettings.Gradient {
				t := amp
				if t > 1 {
					t = 1
				}
				// Dimmer gradient for mirror
				mirrorBarColor = color.RGBA{
					R: uint8(float32(mirrorColor.R) * (0.5 + 0.5*t)),
					G: uint8(float32(mirrorColor.G) * (0.5 + 0.5*t)),
					B: uint8(float32(mirrorColor.B) * (0.5 + 0.5*t)),
					A: 200,
				}
			}

			if v.barSettings.RoundedCaps {
				capRadius := barWidth / 2
				capY := centerY + barHeight
				vector.DrawFilledCircle(screen, barX+barWidth/2, capY, capRadius, mirrorBarColor, true)
				if barHeight > capRadius {
					vector.DrawFilledRect(screen, barX, centerY, barWidth, barHeight-capRadius+1, mirrorBarColor, false)
				}
			} else {
				vector.DrawFilledRect(screen, barX, centerY, barWidth, barHeight, mirrorBarColor, false)
			}
		}
	}
}

// drawRising renders the spectrum as a line that rises with audio and falls with gravity
func (v *Visualizer) drawRising(screen *ebiten.Image, spectrum []float64, preset ColorPreset, centerY float32) {
	numPoints := len(spectrum)
	if numPoints < 2 {
		return
	}

	// Ensure buffers are the right size
	if len(v.risingPositions) != numPoints {
		v.risingPositions = make([]float64, numPoints)
		v.risingVelocities = make([]float64, numPoints)
		v.risingPeaks = make([]float64, numPoints)
		v.risingPeakHold = make([]int, numPoints)
	}

	// Apply frequency shift
	shiftAmount := int(v.freqShift * float64(numPoints) * 0.5)

	maxHeight := float64(v.height) * 0.4

	// Update physics for each point
	for i := 0; i < numPoints; i++ {
		// Get spectrum value with shift
		shiftedIdx := (i + shiftAmount + numPoints) % numPoints
		targetHeight := spectrum[shiftedIdx] * maxHeight

		// Current position
		currentPos := v.risingPositions[i]
		currentVel := v.risingVelocities[i]

		// If spectrum is pushing up higher than current position
		if targetHeight > currentPos {
			// Rise quickly to meet the target
			riseSpeed := (targetHeight - currentPos) * float64(v.rising.MaxRise)
			currentVel = riseSpeed
			currentPos = targetHeight
		} else {
			// Apply gravity (fall)
			currentVel -= float64(v.rising.Gravity)
			currentPos += currentVel

			// Bounce off the bottom
			if currentPos < 0 {
				currentPos = 0
				if currentVel < 0 {
					currentVel = -currentVel * float64(v.rising.Bounce)
				}
			}
		}

		// Update state
		v.risingPositions[i] = currentPos
		v.risingVelocities[i] = currentVel

		// Update peaks
		if currentPos > v.risingPeaks[i] {
			v.risingPeaks[i] = currentPos
			v.risingPeakHold[i] = v.rising.PeakHold
		} else if v.risingPeakHold[i] > 0 {
			v.risingPeakHold[i]--
		} else {
			// Peak falls
			v.risingPeaks[i] -= float64(v.rising.PeakFallRate) * maxHeight
			if v.risingPeaks[i] < 0 {
				v.risingPeaks[i] = 0
			}
		}
	}

	// Get colors
	lineColor, mirrorColor, glowColor := v.getOverlayColors(preset, 0, 255)

	pointSpacing := float32(v.width) / float32(numPoints-1)

	// Draw trail if enabled
	if v.rising.TrailEnabled && v.rising.TrailLength > 0 {
		for t := v.rising.TrailLength; t >= 1; t-- {
			trailAlpha := uint8(float32(80) * (1.0 - float32(t)/float32(v.rising.TrailLength+1)))
			trailColor := color.RGBA{lineColor.R, lineColor.G, lineColor.B, trailAlpha}
			trailOffset := float32(t) * 2.0

			// Draw trail line
			for i := 0; i < numPoints-1; i++ {
				x1 := float32(i) * pointSpacing
				x2 := float32(i+1) * pointSpacing
				y1 := centerY - float32(v.risingPositions[i]) + trailOffset
				y2 := centerY - float32(v.risingPositions[i+1]) + trailOffset

				vector.StrokeLine(screen, x1, y1, x2, y2, v.lineStyle.Thickness*0.5, trailColor, true)
			}
		}
	}

	// Draw glow
	if v.lineStyle.GlowEnabled {
		glowThickness := v.lineStyle.Thickness * v.lineStyle.GlowSize
		for i := 0; i < numPoints-1; i++ {
			x1 := float32(i) * pointSpacing
			x2 := float32(i+1) * pointSpacing
			y1 := centerY - float32(v.risingPositions[i])
			y2 := centerY - float32(v.risingPositions[i+1])

			vector.StrokeLine(screen, x1, y1, x2, y2, glowThickness, glowColor, true)
		}
	}

	// Draw main rising line
	for i := 0; i < numPoints-1; i++ {
		x1 := float32(i) * pointSpacing
		x2 := float32(i+1) * pointSpacing
		y1 := centerY - float32(v.risingPositions[i])
		y2 := centerY - float32(v.risingPositions[i+1])

		vector.StrokeLine(screen, x1, y1, x2, y2, v.lineStyle.Thickness, lineColor, true)
	}

	// Draw mirror (below center)
	for i := 0; i < numPoints-1; i++ {
		x1 := float32(i) * pointSpacing
		x2 := float32(i+1) * pointSpacing
		y1 := centerY + float32(v.risingPositions[i])*0.5
		y2 := centerY + float32(v.risingPositions[i+1])*0.5

		vector.StrokeLine(screen, x1, y1, x2, y2, v.lineStyle.Thickness*0.6, mirrorColor, true)
	}

	// Draw peak markers if enabled
	if v.rising.Peaks {
		peakColor := lineColor
		peakColor.A = 200
		for i := 0; i < numPoints; i++ {
			if v.risingPeaks[i] > 0 {
				x := float32(i) * pointSpacing
				y := centerY - float32(v.risingPeaks[i])
				// Draw small peak marker
				vector.DrawFilledCircle(screen, x, y, 2, peakColor, true)
			}
		}
	}
}

// drawParticles renders an explosive particle system driven by the spectrum
func (v *Visualizer) drawParticles(screen *ebiten.Image, spectrum []float64, preset ColorPreset, centerY float32) {
	numBands := len(spectrum)
	if numBands < 1 {
		return
	}

	// Get base colors
	lineColor, _, glowColor := v.getOverlayColors(preset, 0, 255)

	// Spawn new particles based on spectrum
	bandWidth := float32(v.width) / float32(numBands)
	for i := 0; i < numBands; i++ {
		amp := spectrum[i]
		if amp > 0.1 { // Only spawn if there's significant audio
			// Spawn rate based on amplitude
			spawnChance := amp * float64(v.particles.SpawnRate)
			for spawnChance > 0 {
				if spawnChance >= 1 || (spawnChance > 0 && float64(v.frameCount%10)/10.0 < spawnChance) {
					if len(v.particleList) < v.particles.MaxParticles {
						// Create new particle
						x := float32(i)*bandWidth + bandWidth/2
						speedVariance := (float32(v.frameCount%100)/100.0 - 0.5) * v.particles.SpeedVariance
						spreadAngle := (float32(v.frameCount%100)/100.0 - 0.5) * v.particles.Spread

						p := Particle{
							X:        x,
							Y:        centerY,
							VX:       spreadAngle * v.particles.InitialSpeed * 0.5,
							VY:       -(v.particles.InitialSpeed + speedVariance) * float32(amp),
							Life:     1.0,
							Size:     v.particles.SizeMin + (v.particles.SizeMax-v.particles.SizeMin)*float32(amp),
							ColorIdx: i,
						}
						v.particleList = append(v.particleList, p)
					}
				}
				spawnChance -= 1
			}
		}
	}

	// Update and draw particles
	aliveParticles := make([]Particle, 0, len(v.particleList))
	for _, p := range v.particleList {
		// Apply gravity
		p.VY += v.particles.Gravity

		// Update position
		p.X += p.VX
		p.Y += p.VY

		// Decay life
		lifeDecay := 1.0 / (float32(60) * v.particles.MaxLife)
		p.Life -= lifeDecay

		// Keep if still alive and on screen
		if p.Life > 0 && p.Y < float32(v.height)+50 && p.Y > -50 && p.X > -50 && p.X < float32(v.width)+50 {
			aliveParticles = append(aliveParticles, p)

			// Calculate color based on position in spectrum (rainbow effect)
			var particleColor color.RGBA
			if v.circular.Rainbow || preset.Name == "Rainbow" {
				hue := float64(p.ColorIdx) / float64(numBands) * 360
				particleColor = hueToRGB(hue, 1.0, 1.0)
			} else {
				particleColor = lineColor
			}

			// Fade alpha based on life
			alpha := uint8(255)
			if v.particles.FadeOut {
				alpha = uint8(p.Life * 255)
			}
			particleColor.A = alpha

			// Draw glow if enabled
			if v.particles.Glow {
				glowParticle := glowColor
				glowParticle.A = uint8(float32(alpha) * 0.3)
				vector.DrawFilledCircle(screen, p.X, p.Y, p.Size*1.5, glowParticle, true)
			}

			// Draw particle
			vector.DrawFilledCircle(screen, p.X, p.Y, p.Size*p.Life, particleColor, true)
		}
	}
	v.particleList = aliveParticles
}

// drawCircular renders a circular/radial bar visualizer
func (v *Visualizer) drawCircular(screen *ebiten.Image, spectrum []float64, preset ColorPreset) {
	numBands := len(spectrum)
	if numBands < 1 {
		return
	}

	centerX := float32(v.width) / 2
	centerY := float32(v.height) / 2

	// Update rotation
	v.circular.Rotation += v.circular.RotationSpeed
	if v.circular.Rotation >= 360 {
		v.circular.Rotation -= 360
	}

	// Get base colors
	lineColor, mirrorColor, glowColor := v.getOverlayColors(preset, 0, 255)

	// Angle per band
	angleStep := 2 * math.Pi / float64(numBands)
	baseAngle := float64(v.circular.Rotation) * math.Pi / 180

	for i := 0; i < numBands; i++ {
		amp := float32(spectrum[i])
		angle := baseAngle + float64(i)*angleStep

		// Calculate bar color
		var barColor color.RGBA
		if v.circular.Rainbow {
			hue := float64(i) / float64(numBands) * 360
			barColor = hueToRGB(hue, 1.0, 0.9)
		} else {
			barColor = lineColor
		}

		// Calculate bar length based on amplitude
		barLength := amp * v.circular.BarLength

		// Inner and outer points
		innerR := v.circular.InnerRadius
		outerR := innerR + barLength

		// Calculate positions
		cosA := float32(math.Cos(angle))
		sinA := float32(math.Sin(angle))

		innerX := centerX + innerR*cosA
		innerY := centerY + innerR*sinA
		outerX := centerX + outerR*cosA
		outerY := centerY + outerR*sinA

		if v.circular.Dots {
			// Draw dots at the end of each "bar"
			dotSize := v.circular.DotSize * (0.5 + amp*0.5)

			// Glow
			glowDot := glowColor
			glowDot.A = 80
			vector.DrawFilledCircle(screen, outerX, outerY, dotSize*1.5, glowDot, true)

			// Main dot
			vector.DrawFilledCircle(screen, outerX, outerY, dotSize, barColor, true)

			// Mirror dot (inner)
			if v.circular.Mirror {
				mirrorR := innerR - barLength*0.5
				if mirrorR < 10 {
					mirrorR = 10
				}
				mirrorX := centerX + mirrorR*cosA
				mirrorY := centerY + mirrorR*sinA
				mirrorDot := mirrorColor
				mirrorDot.A = 180
				vector.DrawFilledCircle(screen, mirrorX, mirrorY, dotSize*0.7, mirrorDot, true)
			}
		} else {
			// Draw bars
			thickness := v.lineStyle.Thickness

			// Glow
			if v.lineStyle.GlowEnabled {
				glowBar := glowColor
				glowBar.A = 60
				vector.StrokeLine(screen, innerX, innerY, outerX, outerY, thickness*v.lineStyle.GlowSize, glowBar, true)
			}

			// Main bar
			vector.StrokeLine(screen, innerX, innerY, outerX, outerY, thickness, barColor, true)

			// Mirror bar (inward)
			if v.circular.Mirror {
				mirrorR := innerR - barLength*0.4
				if mirrorR < 5 {
					mirrorR = 5
				}
				mirrorX := centerX + mirrorR*cosA
				mirrorY := centerY + mirrorR*sinA
				mirrorBar := mirrorColor
				mirrorBar.A = 180
				vector.StrokeLine(screen, centerX+innerR*0.9*cosA, centerY+innerR*0.9*sinA, mirrorX, mirrorY, thickness*0.6, mirrorBar, true)
			}
		}
	}

	// Draw center circle
	vector.StrokeCircle(screen, centerX, centerY, v.circular.InnerRadius*0.9, 1, color.RGBA{60, 60, 60, 255}, true)
}

// drawWave renders concentric wave rings that pulse outward
func (v *Visualizer) drawWave(screen *ebiten.Image, spectrum []float64, preset ColorPreset) {
	numBands := len(spectrum)
	if numBands < 1 {
		return
	}

	centerX := float32(v.width) / 2
	centerY := float32(v.height) / 2
	maxRadius := float32(math.Sqrt(float64(v.width*v.width+v.height*v.height))) / 2

	// Get colors
	lineColor, _, glowColor := v.getOverlayColors(preset, 0, 255)

	// Calculate average intensity (or bass for bass-reactive mode)
	var intensity float64
	if v.wave.ReactToBass {
		// Focus on lower frequencies (first 1/4 of spectrum)
		bassEnd := numBands / 4
		if bassEnd < 1 {
			bassEnd = 1
		}
		for i := 0; i < bassEnd; i++ {
			intensity += spectrum[i]
		}
		intensity /= float64(bassEnd)
	} else {
		// Average of all frequencies
		for _, amp := range spectrum {
			intensity += amp
		}
		intensity /= float64(numBands)
	}

	// Spawn new rings based on intensity
	if intensity > 0.15 {
		spawnChance := intensity * float64(v.wave.SpawnRate)
		if spawnChance > 0 && (len(v.rings) == 0 || v.rings[len(v.rings)-1].Radius > v.wave.RingSpeed*5) {
			if len(v.rings) < v.wave.MaxRings {
				v.rings = append(v.rings, Ring{
					Radius:    1,
					Intensity: float32(intensity),
					Age:       0,
				})
			}
		}
	}

	// Update and draw rings
	aliveRings := make([]Ring, 0, len(v.rings))
	for _, ring := range v.rings {
		// Update ring
		ring.Radius += v.wave.RingSpeed
		ring.Age += v.wave.Decay

		// Keep if still visible
		if ring.Age < 1.0 && ring.Radius < maxRadius {
			aliveRings = append(aliveRings, ring)

			// Calculate alpha based on age
			alpha := uint8((1.0 - ring.Age) * 255 * float32(ring.Intensity))

			// Calculate color
			var ringColor color.RGBA
			if v.wave.ColorByAge {
				// Color shifts from line color to dim as it ages
				hue := float64((v.frameCount + int(ring.Radius)) % 360)
				ringColor = hueToRGB(hue, 1.0-float64(ring.Age)*0.5, 1.0-float64(ring.Age)*0.3)
			} else {
				ringColor = lineColor
			}
			ringColor.A = alpha

			// Apply wobble if enabled
			radius := ring.Radius
			if v.wave.Wobble {
				wobbleOffset := float32(math.Sin(float64(ring.Radius)*0.1+float64(v.frameCount)*0.1)) * v.wave.WobbleAmount * 20
				radius += wobbleOffset
			}

			// Draw glow
			glowRing := glowColor
			glowRing.A = alpha / 3
			vector.StrokeCircle(screen, centerX, centerY, radius, v.wave.RingWidth*2, glowRing, true)

			// Draw ring
			if v.wave.Fill {
				// Filled ring (actually a thick stroke)
				vector.StrokeCircle(screen, centerX, centerY, radius, v.wave.RingWidth*3, ringColor, true)
			} else {
				vector.StrokeCircle(screen, centerX, centerY, radius, v.wave.RingWidth, ringColor, true)
			}
		}
	}
	v.rings = aliveRings

	// Draw center pulse indicator
	pulseSize := float32(20 + intensity*40)
	pulseColor := lineColor
	pulseColor.A = uint8(intensity * 200)
	vector.DrawFilledCircle(screen, centerX, centerY, pulseSize, pulseColor, true)
}

// drawIsometric renders the waveform in isometric 3D perspective with Y-axis rotation
func (v *Visualizer) drawIsometric(screen *ebiten.Image, spectrum []float64, historySpectrums [][]float64, preset ColorPreset, centerY float32) {
	angleRad := float64(v.isometric.Angle) * math.Pi / 180.0
	cosAngle := float32(math.Cos(angleRad))
	sinAngle := float32(math.Sin(angleRad))

	// Y-axis rotation angle
	rotationRad := float64(v.isometric.Rotation) * math.Pi / 180.0
	cosRotation := float32(math.Cos(rotationRad))
	sinRotation := float32(math.Sin(rotationRad))

	centerX := float32(v.width) / 2

	// Draw depth layers from back to front
	for layer := v.isometric.DepthLayers - 1; layer >= 0; layer-- {
		// Get spectrum for this layer (use history for trailing effect)
		layerSpectrum := spectrum
		histIdx := (v.historyIndex - 1 - layer*2 + len(v.spectrumHistory)) % len(v.spectrumHistory)
		if histIdx >= 0 && histIdx < len(v.spectrumHistory) && layer > 0 {
			layerSpectrum = v.spectrumHistory[histIdx]
		}

		// Calculate depth offset and scale with forward motion
		// Forward motion creates the effect of traveling through a 3D landscape
		baseDepth := float32(layer) * v.isometric.DepthSpacing

		// Apply forward motion offset - layers cycle through positions
		var effectiveDepth float32
		if v.isometric.ForwardMotion {
			// Offset wraps around based on forward position
			maxDepth := float32(v.isometric.DepthLayers) * v.isometric.DepthSpacing
			effectiveDepth = baseDepth + v.isometric.ForwardOffset
			// Wrap around to create continuous motion
			for effectiveDepth >= maxDepth {
				effectiveDepth -= maxDepth
			}
			for effectiveDepth < 0 {
				effectiveDepth += maxDepth
			}
		} else {
			effectiveDepth = baseDepth
		}

		depth := effectiveDepth
		// Scale based on effective depth for proper perspective
		effectiveLayer := effectiveDepth / v.isometric.DepthSpacing
		scale := float32(math.Pow(float64(v.isometric.ScaleFactor), float64(effectiveLayer)))

		// Base isometric offset
		baseXOffset := depth * cosAngle * 0.3
		baseYOffset := -depth * sinAngle

		// Apply Y-axis rotation to the depth offset
		// This creates a swinging/rotating effect around the center
		xOffset := baseXOffset*cosRotation - depth*sinRotation*0.15
		yOffset := baseYOffset + depth*sinRotation*cosAngle*0.1

		// Calculate alpha for depth fade (use effective layer for proper depth during forward motion)
		alpha := uint8(255)
		if v.isometric.DepthFade && v.isometric.DepthLayers > 1 {
			fadeLayer := int(effectiveLayer)
			if fadeLayer >= v.isometric.DepthLayers {
				fadeLayer = v.isometric.DepthLayers - 1
			}
			alpha = uint8(255 - (200 * fadeLayer / v.isometric.DepthLayers))
		}

		// Get colors (glow now matches line color dynamically)
		lineColor, mirrorColor, glowColor := v.getOverlayColors(preset, layer, alpha)

		// Adjust thickness based on depth
		layerThickness := v.lineStyle.Thickness * scale

		// Draw glow for front layer
		if v.lineStyle.GlowEnabled && layer == 0 {
			v.drawWaveformIsometric(screen, layerSpectrum, centerX+xOffset, centerY+yOffset, scale, false, layerThickness*v.lineStyle.GlowSize, glowColor)
			v.drawWaveformIsometric(screen, layerSpectrum, centerX+xOffset, centerY+yOffset, scale, true, layerThickness*v.lineStyle.GlowSize, glowColor)
		}

		// Draw the waveform layer
		v.drawWaveformIsometric(screen, layerSpectrum, centerX+xOffset, centerY+yOffset, scale, false, layerThickness, lineColor)
		v.drawWaveformIsometric(screen, layerSpectrum, centerX+xOffset, centerY+yOffset, scale, true, layerThickness*0.75, mirrorColor)

		// Draw connecting lines between layers (depth lines) for front layers
		if layer < v.isometric.DepthLayers-1 && layer < 2 {
			v.drawDepthConnectors(screen, layerSpectrum, centerX+xOffset, centerY+yOffset, scale, depth, cosAngle, sinAngle, lineColor)
		}
	}
}

// drawWaveformIsometric draws a waveform with isometric scaling and offset
func (v *Visualizer) drawWaveformIsometric(screen *ebiten.Image, spectrum []float64, centerX, centerY, scale float32, mirror bool, thickness float32, lineColor color.RGBA) {
	numPoints := len(spectrum)
	if numPoints < 2 {
		return
	}

	// Scale the width based on depth
	scaledWidth := float32(v.width) * scale
	startX := centerX - scaledWidth/2
	pointSpacing := scaledWidth / float32(numPoints-1)

	// Apply frequency shift
	shiftAmount := int(v.freqShift * float64(numPoints) * 0.5)

	// Helper to get Y position for a given index
	getY := func(idx int) float32 {
		shiftedIdx := (idx + shiftAmount + numPoints) % numPoints
		amp := float32(spectrum[shiftedIdx]) * float32(v.height) * 0.35 * scale
		if mirror {
			return centerY + amp
		}
		return centerY - amp
	}

	if v.smoothCurves && numPoints >= 4 {
		// Catmull-Rom spline interpolation
		segmentsPerPoint := 6

		for i := 0; i < numPoints-1; i++ {
			i0 := i - 1
			if i0 < 0 {
				i0 = 0
			}
			i1 := i
			i2 := i + 1
			i3 := i + 2
			if i3 >= numPoints {
				i3 = numPoints - 1
			}

			x0, y0 := startX+float32(i0)*pointSpacing, getY(i0)
			x1, y1 := startX+float32(i1)*pointSpacing, getY(i1)
			x2, y2 := startX+float32(i2)*pointSpacing, getY(i2)
			x3, y3 := startX+float32(i3)*pointSpacing, getY(i3)

			prevX, prevY := x1, y1
			for s := 1; s <= segmentsPerPoint; s++ {
				t := float32(s) / float32(segmentsPerPoint)
				t2 := t * t
				t3 := t2 * t

				nextX := 0.5 * ((2 * x1) + (-x0+x2)*t + (2*x0-5*x1+4*x2-x3)*t2 + (-x0+3*x1-3*x2+x3)*t3)
				nextY := 0.5 * ((2 * y1) + (-y0+y2)*t + (2*y0-5*y1+4*y2-y3)*t2 + (-y0+3*y1-3*y2+y3)*t3)

				vector.StrokeLine(screen, prevX, prevY, nextX, nextY, thickness, lineColor, true)
				prevX, prevY = nextX, nextY
			}
		}
	} else {
		// Simple linear
		for i := 0; i < numPoints-1; i++ {
			x1 := startX + float32(i)*pointSpacing
			x2 := startX + float32(i+1)*pointSpacing
			y1 := getY(i)
			y2 := getY(i + 1)
			vector.StrokeLine(screen, x1, y1, x2, y2, thickness, lineColor, true)
		}
	}
}

// drawDepthConnectors draws vertical lines connecting depth layers
func (v *Visualizer) drawDepthConnectors(screen *ebiten.Image, spectrum []float64, centerX, centerY, scale, depth, cosAngle, sinAngle float32, lineColor color.RGBA) {
	numPoints := len(spectrum)
	if numPoints < 2 {
		return
	}

	scaledWidth := float32(v.width) * scale
	startX := centerX - scaledWidth/2
	pointSpacing := scaledWidth / float32(numPoints-1)

	// Draw connectors at regular intervals
	step := numPoints / 16
	if step < 1 {
		step = 1
	}

	nextDepth := depth + v.isometric.DepthSpacing
	nextScale := scale * v.isometric.ScaleFactor
	nextXOffset := nextDepth * cosAngle * 0.3
	nextYOffset := -nextDepth * sinAngle

	nextScaledWidth := float32(v.width) * nextScale
	nextStartX := (centerX + nextXOffset) - nextScaledWidth/2
	nextPointSpacing := nextScaledWidth / float32(numPoints-1)

	shiftAmount := int(v.freqShift * float64(numPoints) * 0.5)

	// Faded color for connectors
	connectorColor := lineColor
	connectorColor.A = lineColor.A / 4

	for i := 0; i < numPoints; i += step {
		shiftedIdx := (i + shiftAmount + numPoints) % numPoints
		amp := float32(spectrum[shiftedIdx]) * float32(v.height) * 0.35

		x1 := startX + float32(i)*pointSpacing
		y1 := centerY - amp*scale

		x2 := nextStartX + float32(i)*nextPointSpacing
		y2 := (centerY + nextYOffset) - amp*nextScale

		vector.StrokeLine(screen, x1, y1, x2, y2, 0.5, connectorColor, true)
	}
}

// getOverlayColors returns the line, mirror, and glow colors for an overlay layer
func (v *Visualizer) getOverlayColors(preset ColorPreset, overlayIndex int, alpha uint8) (color.RGBA, color.RGBA, color.RGBA) {
	lineColor := preset.LineColor
	mirrorColor := preset.MirrorColor

	// Special handling for rainbow mode
	if preset.Name == "Rainbow" {
		hue := float64((v.frameCount + overlayIndex*30) % 360)
		lineColor = hueToRGB(hue, 1.0, 1.0)
		mirrorColor = hueToRGB(hue, 0.7, 0.5)
	}

	// Beat-reactive color shift: on beats, shift hue and increase brightness
	if v.beatDetector.Enabled && v.beatDetector.beatIntensity > 0 {
		intensity := v.beatDetector.beatIntensity

		// Convert current color to HSV-like values and shift
		// Simple approach: brighten and shift towards complementary color
		lineColor = shiftColorOnBeat(lineColor, intensity)
		mirrorColor = shiftColorOnBeat(mirrorColor, intensity)
	}

	// Derive glow color from the current line color (so it matches dynamically)
	glowColor := deriveGlowColor(lineColor)

	lineColor.A = alpha
	mirrorColor.A = alpha / 2
	glowColor.A = alpha / 3
	return lineColor, mirrorColor, glowColor
}

// shiftColorOnBeat shifts a color based on beat intensity
func shiftColorOnBeat(c color.RGBA, intensity float64) color.RGBA {
	// Brighten the color and shift hue slightly
	brighten := 1.0 + intensity*0.5 // Up to 50% brighter

	// Calculate new RGB values with brightening
	r := float64(c.R) * brighten
	g := float64(c.G) * brighten
	b := float64(c.B) * brighten

	// Also add a color shift: rotate RGB channels slightly based on intensity
	shift := intensity * 0.3 // 30% shift at full intensity
	newR := r*(1-shift) + b*shift
	newG := g*(1-shift) + r*shift
	newB := b*(1-shift) + g*shift

	// Clamp values
	if newR > 255 {
		newR = 255
	}
	if newG > 255 {
		newG = 255
	}
	if newB > 255 {
		newB = 255
	}

	return color.RGBA{
		R: uint8(newR),
		G: uint8(newG),
		B: uint8(newB),
		A: c.A,
	}
}

// deriveGlowColor creates a glow color from a line color
// The glow is a semi-transparent version of the line color
func deriveGlowColor(lineColor color.RGBA) color.RGBA {
	return color.RGBA{
		R: lineColor.R,
		G: lineColor.G,
		B: lineColor.B,
		A: 80, // Base glow alpha, will be adjusted by caller
	}
}

// drawSpiral renders the waveform as a spiral pattern around a center point
func (v *Visualizer) drawSpiral(screen *ebiten.Image, spectrum []float64, preset ColorPreset) {
	numPoints := len(spectrum)
	if numPoints < 2 {
		return
	}

	centerX := float32(v.width) / 2
	centerY := float32(v.height) / 2

	// Base radius and max amplitude
	baseRadius := float32(math.Min(float64(v.width), float64(v.height))) * 0.15
	maxRadius := float32(math.Min(float64(v.width), float64(v.height))) * 0.4

	// Get colors
	lineColor, _, glowColor := v.getOverlayColors(preset, 0, 255)

	// Apply frequency shift
	shiftAmount := int(v.freqShift * float64(numPoints) * 0.5)

	// Number of full rotations the spiral makes
	rotations := 2.0 + float64(v.overlay.Count-1)*0.5

	// Draw glow layer first if enabled
	if v.lineStyle.GlowEnabled {
		v.drawSpiralLayer(screen, spectrum, centerX, centerY, baseRadius, maxRadius, rotations, shiftAmount, v.lineStyle.Thickness*v.lineStyle.GlowSize, glowColor)
	}

	// Draw main spiral
	v.drawSpiralLayer(screen, spectrum, centerX, centerY, baseRadius, maxRadius, rotations, shiftAmount, v.lineStyle.Thickness, lineColor)

	// Draw mirrored (inner) spiral with dimmer color
	mirrorColor := preset.MirrorColor
	mirrorColor.A = 180
	innerBaseRadius := baseRadius * 0.6
	innerMaxRadius := maxRadius * 0.6
	v.drawSpiralLayer(screen, spectrum, centerX, centerY, innerBaseRadius, innerMaxRadius, rotations, shiftAmount, v.lineStyle.Thickness*0.75, mirrorColor)
}

// drawSpiralLayer draws a single spiral layer
func (v *Visualizer) drawSpiralLayer(screen *ebiten.Image, spectrum []float64, centerX, centerY, baseRadius, maxRadius float32, rotations float64, shiftAmount int, thickness float32, lineColor color.RGBA) {
	numPoints := len(spectrum)

	// Animation offset based on frame count for rotation effect
	animOffset := float64(v.frameCount) * 0.02

	// Helper to get point on spiral
	getPoint := func(idx int) (float32, float32) {
		// Angle progresses through the spiral
		t := float64(idx) / float64(numPoints-1)
		angle := t*rotations*2*math.Pi + animOffset

		// Get spectrum value with shift
		shiftedIdx := (idx + shiftAmount + numPoints) % numPoints
		amp := float32(spectrum[shiftedIdx]) * float32(v.damping)

		// Radius varies based on position in spiral and amplitude
		radius := baseRadius + (maxRadius-baseRadius)*float32(t) + amp*maxRadius*0.5

		x := centerX + radius*float32(math.Cos(angle))
		y := centerY + radius*float32(math.Sin(angle))
		return x, y
	}

	if v.smoothCurves && numPoints >= 4 {
		// Catmull-Rom spline for smooth spiral
		segmentsPerPoint := 5

		for i := 0; i < numPoints-1; i++ {
			i0 := i - 1
			if i0 < 0 {
				i0 = 0
			}
			i1 := i
			i2 := i + 1
			i3 := i + 2
			if i3 >= numPoints {
				i3 = numPoints - 1
			}

			x0, y0 := getPoint(i0)
			x1, y1 := getPoint(i1)
			x2, y2 := getPoint(i2)
			x3, y3 := getPoint(i3)

			prevX, prevY := x1, y1
			for s := 1; s <= segmentsPerPoint; s++ {
				t := float32(s) / float32(segmentsPerPoint)
				t2 := t * t
				t3 := t2 * t

				nextX := 0.5 * ((2 * x1) + (-x0+x2)*t + (2*x0-5*x1+4*x2-x3)*t2 + (-x0+3*x1-3*x2+x3)*t3)
				nextY := 0.5 * ((2 * y1) + (-y0+y2)*t + (2*y0-5*y1+4*y2-y3)*t2 + (-y0+3*y1-3*y2+y3)*t3)

				vector.StrokeLine(screen, prevX, prevY, nextX, nextY, thickness, lineColor, true)
				prevX, prevY = nextX, nextY
			}
		}
	} else {
		// Simple linear
		for i := 0; i < numPoints-1; i++ {
			x1, y1 := getPoint(i)
			x2, y2 := getPoint(i + 1)
			vector.StrokeLine(screen, x1, y1, x2, y2, thickness, lineColor, true)
		}
	}
}

// getUIAccentColor returns the current accent color for UI elements based on the active color preset
func (v *Visualizer) getUIAccentColor() color.RGBA {
	preset := colorPresets[v.lineStyle.ColorPreset]

	// Special handling for rainbow mode - animate the UI color too
	if preset.Name == "Rainbow" {
		hue := float64(v.frameCount % 360)
		return hueToRGB(hue, 1.0, 1.0)
	}

	return preset.LineColor
}

// getUIDimColor returns a dimmed version of the accent color for backgrounds/inactive elements
func (v *Visualizer) getUIDimColor() color.RGBA {
	accent := v.getUIAccentColor()
	return color.RGBA{
		R: accent.R / 3,
		G: accent.G / 3,
		B: accent.B / 3,
		A: 255,
	}
}

// hueToRGB converts HSV to RGB (saturation and value are 0-1)
func hueToRGB(hue, saturation, value float64) color.RGBA {
	c := value * saturation
	x := c * (1 - math.Abs(math.Mod(hue/60, 2)-1))
	m := value - c

	var r, g, b float64
	switch {
	case hue < 60:
		r, g, b = c, x, 0
	case hue < 120:
		r, g, b = x, c, 0
	case hue < 180:
		r, g, b = 0, c, x
	case hue < 240:
		r, g, b = 0, x, c
	case hue < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	return color.RGBA{
		R: uint8((r + m) * 255),
		G: uint8((g + m) * 255),
		B: uint8((b + m) * 255),
		A: 255,
	}
}

// drawWaveformStyled draws the waveform with custom color and thickness
func (v *Visualizer) drawWaveformStyled(screen *ebiten.Image, spectrum []float64, centerY float32, mirror bool, thickness float32, lineColor color.RGBA) {
	numPoints := len(spectrum)
	if numPoints < 2 {
		return
	}

	pointSpacing := float32(v.width) / float32(numPoints-1)

	// Apply frequency shift - shifts which frequency appears at each position
	shiftAmount := int(v.freqShift * float64(numPoints) * 0.5)

	// Helper to get Y position for a given index
	getY := func(idx int) float32 {
		shiftedIdx := (idx + shiftAmount + numPoints) % numPoints
		amp := float32(spectrum[shiftedIdx]) * float32(v.height) * 0.4
		if mirror {
			return centerY + amp
		}
		return centerY - amp
	}

	if v.smoothCurves && numPoints >= 4 {
		// Use Catmull-Rom spline interpolation for smooth curves
		segmentsPerPoint := 6 // Number of interpolated segments between each data point for fluid motion

		for i := 0; i < numPoints-1; i++ {
			// Get 4 control points for Catmull-Rom: p0, p1, p2, p3
			// p1 and p2 are the segment endpoints, p0 and p3 are neighbors for tangent calculation
			i0 := i - 1
			if i0 < 0 {
				i0 = 0
			}
			i1 := i
			i2 := i + 1
			i3 := i + 2
			if i3 >= numPoints {
				i3 = numPoints - 1
			}

			x0, y0 := float32(i0)*pointSpacing, getY(i0)
			x1, y1 := float32(i1)*pointSpacing, getY(i1)
			x2, y2 := float32(i2)*pointSpacing, getY(i2)
			x3, y3 := float32(i3)*pointSpacing, getY(i3)

			// Draw interpolated segments
			prevX, prevY := x1, y1
			for s := 1; s <= segmentsPerPoint; s++ {
				t := float32(s) / float32(segmentsPerPoint)

				// Catmull-Rom spline formula
				t2 := t * t
				t3 := t2 * t

				nextX := 0.5 * ((2 * x1) +
					(-x0+x2)*t +
					(2*x0-5*x1+4*x2-x3)*t2 +
					(-x0+3*x1-3*x2+x3)*t3)

				nextY := 0.5 * ((2 * y1) +
					(-y0+y2)*t +
					(2*y0-5*y1+4*y2-y3)*t2 +
					(-y0+3*y1-3*y2+y3)*t3)

				vector.StrokeLine(screen, prevX, prevY, nextX, nextY, thickness, lineColor, true)
				prevX, prevY = nextX, nextY
			}
		}
	} else {
		// Simple linear interpolation (original behavior)
		for i := 0; i < numPoints-1; i++ {
			x1 := float32(i) * pointSpacing
			x2 := float32(i+1) * pointSpacing
			y1 := getY(i)
			y2 := getY(i + 1)

			vector.StrokeLine(screen, x1, y1, x2, y2, thickness, lineColor, true)
		}
	}
}

func (v *Visualizer) drawUI(screen *ebiten.Image) {
	// Draw visualization mode dropdown (top-left)
	v.drawVizModeDropdown(screen)

	// Draw color picker dropdown (next to viz mode)
	v.drawColorDropdown(screen)

	// Status message (moved to not overlap with dropdowns)
	if v.statusTimer > 0 && v.statusMessage != "" {
		ebitenutil.DebugPrintAt(screen, v.statusMessage, 310, 14)
	}

	// Help hint
	if !v.showHelp && !v.showStylePanel && !v.settingsMenu.Visible {
		ebitenutil.DebugPrintAt(screen, "H: Help  M: Menu", v.width-140, v.height-25)
	}

	// Settings info (hide when settings menu is visible to avoid clutter)
	if !v.settingsMenu.Visible {
		settingsY := v.height - 60
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Sensitivity: %.2f  Smoothing: %.2f  FPS: %.0f",
			v.config.Sensitivity.Sensitivity, v.config.Smoothing.AttackSpeed, ebiten.ActualFPS()), 10, settingsY)
	}

	// Draw device dropdown
	v.drawDropdown(screen)

	// Draw settings menu (collapsible dial menu)
	if v.settingsMenu.Visible {
		v.drawSettingsMenu(screen)
	}

	// Draw style panel
	if v.showStylePanel {
		v.drawStylePanel(screen)
	}

	// Draw help overlay
	if v.showHelp {
		v.drawHelp(screen)
	}
}

// drawSettingsMenu draws the collapsible settings menu with dial controls
func (v *Visualizer) drawSettingsMenu(screen *ebiten.Image) {
	dials := v.getDialControls()

	// Get dynamic accent colors
	accentColor := v.getUIAccentColor()
	dimColor := v.getUIDimColor()
	fillColor := color.RGBA{accentColor.R / 2, accentColor.G / 2, accentColor.B / 2, 255}

	// Calculate menu height based on number of dials
	dialHeight := 36
	headerHeight := 50
	footerHeight := 30
	menuHeight := headerHeight + len(dials)*dialHeight + footerHeight

	menuX := float32(v.settingsMenu.X)
	menuY := float32(v.settingsMenu.Y)
	menuW := float32(v.settingsMenu.Width)
	menuH := float32(menuHeight)

	// Draw semi-transparent background
	vector.DrawFilledRect(screen, menuX, menuY, menuW, menuH, color.RGBA{15, 15, 20, 240}, false)
	vector.StrokeRect(screen, menuX, menuY, menuW, menuH, 2, accentColor, false)

	// Draw header
	ebitenutil.DebugPrintAt(screen, "Settings Menu (M to close)", int(menuX)+10, int(menuY)+10)

	// Draw separator line
	vector.StrokeLine(screen, menuX+10, menuY+35, menuX+menuW-10, menuY+35, 1, dimColor, false)

	// Draw each dial
	dialStartY := int(menuY) + headerHeight
	labelX := int(menuX) + 10
	sliderX := int(menuX) + 100
	sliderW := 180

	for i, dial := range dials {
		dialY := dialStartY + i*dialHeight

		// Draw label
		ebitenutil.DebugPrintAt(screen, dial.Name, labelX, dialY+4)

		// Draw slider track
		trackY := float32(dialY + 8)
		trackH := float32(8)
		vector.DrawFilledRect(screen, float32(sliderX), trackY, float32(sliderW), trackH, color.RGBA{30, 30, 35, 255}, false)
		vector.StrokeRect(screen, float32(sliderX), trackY, float32(sliderW), trackH, 1, dimColor, false)

		// Check if this is a bi-directional dial (has negative min or center at 0)
		isBidirectional := dial.Min < 0 && dial.Max > 0

		// Calculate fill ratio
		ratio := (*dial.Value - dial.Min) / (dial.Max - dial.Min)
		ratio = math.Max(0, math.Min(1, ratio))

		// Draw filled portion
		dialFillColor := fillColor
		if v.settingsMenu.ActiveDial == i {
			dialFillColor = accentColor // Brighter when active
		}

		if isBidirectional {
			// For bi-directional dials, fill from center
			centerX := float32(sliderX) + float32(sliderW)/2
			handleX := float32(sliderX) + float32(sliderW)*float32(ratio)

			// Draw center mark
			vector.StrokeLine(screen, centerX, trackY-2, centerX, trackY+trackH+2, 1, dimColor, false)

			// Fill from center to handle position
			if handleX > centerX {
				vector.DrawFilledRect(screen, centerX, trackY+1, handleX-centerX, trackH-2, dialFillColor, false)
			} else if handleX < centerX {
				vector.DrawFilledRect(screen, handleX, trackY+1, centerX-handleX, trackH-2, dialFillColor, false)
			}

			// Draw handle
			handleW := float32(12)
			handleH := float32(18)
			handleY := trackY - (handleH-trackH)/2
			handleColor := accentColor
			if v.settingsMenu.ActiveDial == i {
				handleColor = color.RGBA{
					R: uint8(min(255, int(accentColor.R)+50)),
					G: uint8(min(255, int(accentColor.G)+50)),
					B: uint8(min(255, int(accentColor.B)+50)),
					A: 255,
				}
			}
			vector.DrawFilledRect(screen, handleX-handleW/2, handleY, handleW, handleH, handleColor, false)
			vector.StrokeRect(screen, handleX-handleW/2, handleY, handleW, handleH, 1, accentColor, false)
		} else {
			// Standard left-to-right fill
			fillW := float32(sliderW) * float32(ratio)
			if fillW > 0 {
				vector.DrawFilledRect(screen, float32(sliderX), trackY+1, fillW, trackH-2, dialFillColor, false)
			}

			// Draw handle
			handleX := float32(sliderX) + fillW
			handleW := float32(12)
			handleH := float32(18)
			handleY := trackY - (handleH-trackH)/2
			handleColor := accentColor
			if v.settingsMenu.ActiveDial == i {
				handleColor = color.RGBA{
					R: uint8(min(255, int(accentColor.R)+50)),
					G: uint8(min(255, int(accentColor.G)+50)),
					B: uint8(min(255, int(accentColor.B)+50)),
					A: 255,
				}
			}
			vector.DrawFilledRect(screen, handleX-handleW/2, handleY, handleW, handleH, handleColor, false)
			vector.StrokeRect(screen, handleX-handleW/2, handleY, handleW, handleH, 1, accentColor, false)
		}

		// Draw value
		valueStr := fmt.Sprintf(dial.Format, *dial.Value)
		ebitenutil.DebugPrintAt(screen, valueStr, sliderX+sliderW+8, dialY+4)
	}

	// Draw footer hint
	footerY := dialStartY + len(dials)*dialHeight + 5
	ebitenutil.DebugPrintAt(screen, "Drag sliders | Right-click to reset", int(menuX)+10, footerY)
}

func (v *Visualizer) drawVizModeDropdown(screen *ebiten.Image) {
	dropdownX := 10
	dropdownY := 10
	dropdownW := 120
	itemH := 24

	accentColor := v.getUIAccentColor()

	// Draw button background
	vector.DrawFilledRect(screen, float32(dropdownX), float32(dropdownY), float32(dropdownW), float32(itemH), v.dropdownBg, false)
	vector.StrokeRect(screen, float32(dropdownX), float32(dropdownY), float32(dropdownW), float32(itemH), 1, accentColor, false)

	// Draw selected mode text
	selectedText := "Line"
	if v.vizMode >= 0 && v.vizMode < len(vizModeNames) {
		selectedText = vizModeNames[v.vizMode]
	}
	ebitenutil.DebugPrintAt(screen, "Mode: "+selectedText, dropdownX+8, dropdownY+4)

	// Draw arrow
	arrowX := float32(dropdownX + dropdownW - 16)
	arrowY := float32(dropdownY + itemH/2)
	if v.showVizDropdown {
		// Up arrow
		vector.StrokeLine(screen, arrowX, arrowY+3, arrowX+4, arrowY-3, 1, color.RGBA{150, 150, 150, 255}, false)
		vector.StrokeLine(screen, arrowX+4, arrowY-3, arrowX+8, arrowY+3, 1, color.RGBA{150, 150, 150, 255}, false)
	} else {
		// Down arrow
		vector.StrokeLine(screen, arrowX, arrowY-3, arrowX+4, arrowY+3, 1, color.RGBA{150, 150, 150, 255}, false)
		vector.StrokeLine(screen, arrowX+4, arrowY+3, arrowX+8, arrowY-3, 1, color.RGBA{150, 150, 150, 255}, false)
	}

	// Draw dropdown list if open
	if v.showVizDropdown {
		listH := len(vizModeNames) * itemH
		vector.DrawFilledRect(screen, float32(dropdownX), float32(dropdownY+itemH), float32(dropdownW), float32(listH), v.dropdownBg, false)
		vector.StrokeRect(screen, float32(dropdownX), float32(dropdownY+itemH), float32(dropdownW), float32(listH), 1, accentColor, false)

		for i, modeName := range vizModeNames {
			itemY := dropdownY + itemH + i*itemH

			// Highlight
			if i == v.hoverVizMode {
				vector.DrawFilledRect(screen, float32(dropdownX), float32(itemY), float32(dropdownW), float32(itemH), v.dropdownHover, false)
			} else if i == v.vizMode {
				vector.DrawFilledRect(screen, float32(dropdownX), float32(itemY), float32(dropdownW), float32(itemH), color.RGBA{accentColor.R / 3, accentColor.G / 3, accentColor.B / 3, 255}, false)
			}

			ebitenutil.DebugPrintAt(screen, modeName, dropdownX+8, itemY+4)
		}
	}
}

func (v *Visualizer) drawColorDropdown(screen *ebiten.Image) {
	dropdownX := 140
	dropdownY := 10
	dropdownW := 160
	buttonH := 24
	itemH := 20
	maxVisible := 12

	// Get current preset for button display
	currentPreset := colorPresets[v.lineStyle.ColorPreset]

	// Draw button background
	vector.DrawFilledRect(screen, float32(dropdownX), float32(dropdownY), float32(dropdownW), float32(buttonH), v.dropdownBg, false)
	vector.StrokeRect(screen, float32(dropdownX), float32(dropdownY), float32(dropdownW), float32(buttonH), 1, currentPreset.LineColor, false)

	// Draw selected color name in its own color
	ebitenutil.DebugPrintAt(screen, currentPreset.Name, dropdownX+8, dropdownY+4)

	// Draw arrow
	arrowX := float32(dropdownX + dropdownW - 16)
	arrowY := float32(dropdownY + buttonH/2)
	if v.showColorDropdown {
		vector.StrokeLine(screen, arrowX, arrowY+3, arrowX+4, arrowY-3, 1, currentPreset.LineColor, false)
		vector.StrokeLine(screen, arrowX+4, arrowY-3, arrowX+8, arrowY+3, 1, currentPreset.LineColor, false)
	} else {
		vector.StrokeLine(screen, arrowX, arrowY-3, arrowX+4, arrowY+3, 1, currentPreset.LineColor, false)
		vector.StrokeLine(screen, arrowX+4, arrowY+3, arrowX+8, arrowY-3, 1, currentPreset.LineColor, false)
	}

	// Draw dropdown list if open
	if v.showColorDropdown {
		visibleCount := len(colorPresets)
		if visibleCount > maxVisible {
			visibleCount = maxVisible
		}
		listH := visibleCount * itemH

		// Draw background
		vector.DrawFilledRect(screen, float32(dropdownX), float32(dropdownY+buttonH), float32(dropdownW), float32(listH), color.RGBA{20, 20, 25, 250}, false)
		vector.StrokeRect(screen, float32(dropdownX), float32(dropdownY+buttonH), float32(dropdownW), float32(listH), 1, currentPreset.LineColor, false)

		for i := 0; i < visibleCount; i++ {
			actualIndex := i + v.colorDropdownScroll
			if actualIndex >= len(colorPresets) {
				break
			}
			preset := colorPresets[actualIndex]
			itemY := dropdownY + buttonH + i*itemH

			// Highlight on hover or selection
			if actualIndex == v.hoverColorPreset {
				vector.DrawFilledRect(screen, float32(dropdownX+1), float32(itemY), float32(dropdownW-2), float32(itemH), color.RGBA{50, 50, 55, 255}, false)
			} else if actualIndex == v.lineStyle.ColorPreset {
				vector.DrawFilledRect(screen, float32(dropdownX+1), float32(itemY), float32(dropdownW-2), float32(itemH), color.RGBA{30, 30, 35, 255}, false)
			}

			// Draw color swatch
			swatchX := float32(dropdownX + 6)
			swatchY := float32(itemY + 3)
			swatchW := float32(14)
			swatchH := float32(itemH - 6)

			// Draw gradient swatch showing both line and mirror colors
			vector.DrawFilledRect(screen, swatchX, swatchY, swatchW/2, swatchH, preset.LineColor, false)
			vector.DrawFilledRect(screen, swatchX+swatchW/2, swatchY, swatchW/2, swatchH, preset.MirrorColor, false)
			vector.StrokeRect(screen, swatchX, swatchY, swatchW, swatchH, 1, color.RGBA{80, 80, 80, 255}, false)

			// Draw preset name in its line color
			textColor := preset.LineColor
			// For dark colors, use the line color; for light backgrounds it shows well
			// Draw text with colored background hint
			ebitenutil.DebugPrintAt(screen, preset.Name, dropdownX+26, itemY+2)

			// Draw a small colored line under the text to indicate the color
			vector.StrokeLine(screen, float32(dropdownX+26), float32(itemY+itemH-2), float32(dropdownX+26)+float32(len(preset.Name)*7), float32(itemY+itemH-2), 2, textColor, false)
		}

		// Draw scroll indicators if needed
		if len(colorPresets) > maxVisible {
			if v.colorDropdownScroll > 0 {
				// Up arrow indicator
				ebitenutil.DebugPrintAt(screen, "▲", dropdownX+dropdownW-16, dropdownY+buttonH+2)
			}
			if v.colorDropdownScroll < len(colorPresets)-maxVisible {
				// Down arrow indicator
				ebitenutil.DebugPrintAt(screen, "▼", dropdownX+dropdownW-16, dropdownY+buttonH+listH-12)
			}
		}
	}
}

func (v *Visualizer) drawSliders(screen *ebiten.Image) {
	sliderY := float32(v.height - 40)
	sliderH := float32(8)
	handleW := float32(12)
	handleH := float32(18)

	// Get dynamic accent colors
	accentColor := v.getUIAccentColor()
	dimColor := v.getUIDimColor()

	// === Frequency Shift Slider (left side) ===
	shiftSliderX := float32(100)
	shiftSliderW := float32((v.width - 240) / 2)

	// Draw label
	ebitenutil.DebugPrintAt(screen, "Freq Shift", 10, v.height-44)

	// Draw slider background (track)
	vector.DrawFilledRect(screen, shiftSliderX, sliderY, shiftSliderW, sliderH, v.sliderBg, false)
	vector.StrokeRect(screen, shiftSliderX, sliderY, shiftSliderW, sliderH, 1, dimColor, false)

	// Draw center mark
	centerX := shiftSliderX + shiftSliderW/2
	vector.StrokeLine(screen, centerX, sliderY-2, centerX, sliderY+sliderH+2, 1, dimColor, false)

	// Draw filled portion from center
	shiftHandlePos := shiftSliderX + shiftSliderW*float32((v.freqShift+1.0)/2.0)
	fillColor := color.RGBA{accentColor.R / 2, accentColor.G / 2, accentColor.B / 2, 255}
	if v.freqShift > 0 {
		fillW := shiftHandlePos - centerX
		vector.DrawFilledRect(screen, centerX, sliderY+1, fillW, sliderH-2, fillColor, false)
	} else if v.freqShift < 0 {
		fillW := centerX - shiftHandlePos
		vector.DrawFilledRect(screen, shiftHandlePos, sliderY+1, fillW, sliderH-2, fillColor, false)
	}

	// Draw handle
	handleY := sliderY - (handleH-sliderH)/2
	vector.DrawFilledRect(screen, shiftHandlePos-handleW/2, handleY, handleW, handleH, accentColor, false)
	vector.StrokeRect(screen, shiftHandlePos-handleW/2, handleY, handleW, handleH, 1, accentColor, false)

	// Draw value
	shiftPercent := v.freqShift * 100
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%.0f%%", shiftPercent), int(shiftSliderX+shiftSliderW)+5, v.height-44)

	// === Frequency Bands Slider (right side) ===
	bandsSliderX := shiftSliderX + shiftSliderW + 50
	bandsSliderW := shiftSliderW

	// Draw label
	ebitenutil.DebugPrintAt(screen, "Bands", int(bandsSliderX)-45, v.height-44)

	// Draw slider background (track)
	vector.DrawFilledRect(screen, bandsSliderX, sliderY, bandsSliderW, sliderH, v.sliderBg, false)
	vector.StrokeRect(screen, bandsSliderX, sliderY, bandsSliderW, sliderH, 1, dimColor, false)

	// Calculate handle position (logarithmic scale)
	minBands := 16.0
	maxBands := 256.0
	logMin := math.Log(minBands)
	logMax := math.Log(maxBands)
	logCurrent := math.Log(float64(v.numBands))
	bandsRatio := float32((logCurrent - logMin) / (logMax - logMin))
	bandsHandlePos := bandsSliderX + bandsSliderW*bandsRatio

	// Draw filled portion from left
	vector.DrawFilledRect(screen, bandsSliderX, sliderY+1, bandsHandlePos-bandsSliderX, sliderH-2, fillColor, false)

	// Draw handle
	vector.DrawFilledRect(screen, bandsHandlePos-handleW/2, handleY, handleW, handleH, accentColor, false)
	vector.StrokeRect(screen, bandsHandlePos-handleW/2, handleY, handleW, handleH, 1, accentColor, false)

	// Draw value
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", v.numBands), int(bandsSliderX+bandsSliderW)+5, v.height-44)

	// Draw hint
	ebitenutil.DebugPrintAt(screen, "(Right-click sliders to reset)", v.width-220, v.height-25)
}

func (v *Visualizer) drawDropdown(screen *ebiten.Image) {
	dropdownX := v.width - 350
	dropdownY := 10
	dropdownW := 340
	itemH := 24

	// Draw button background
	vector.DrawFilledRect(screen, float32(dropdownX), float32(dropdownY), float32(dropdownW), float32(itemH), v.dropdownBg, false)
	vector.StrokeRect(screen, float32(dropdownX), float32(dropdownY), float32(dropdownW), float32(itemH), 1, color.RGBA{80, 80, 80, 255}, false)

	// Draw selected device text
	selectedText := "Select audio device..."
	if v.selectedDevice >= 0 && v.selectedDevice < len(v.devices) {
		dev := v.devices[v.selectedDevice]
		prefix := "[Mic] "
		if dev.IsLoopback {
			prefix = "[Output] "
		}
		selectedText = prefix + truncate(dev.Name, 30)
	}
	ebitenutil.DebugPrintAt(screen, selectedText, dropdownX+8, dropdownY+4)

	// Draw arrow
	arrowX := float32(dropdownX + dropdownW - 20)
	arrowY := float32(dropdownY + itemH/2)
	if v.showDropdown {
		// Up arrow
		vector.StrokeLine(screen, arrowX, arrowY+3, arrowX+4, arrowY-3, 1, color.RGBA{150, 150, 150, 255}, false)
		vector.StrokeLine(screen, arrowX+4, arrowY-3, arrowX+8, arrowY+3, 1, color.RGBA{150, 150, 150, 255}, false)
	} else {
		// Down arrow
		vector.StrokeLine(screen, arrowX, arrowY-3, arrowX+4, arrowY+3, 1, color.RGBA{150, 150, 150, 255}, false)
		vector.StrokeLine(screen, arrowX+4, arrowY+3, arrowX+8, arrowY-3, 1, color.RGBA{150, 150, 150, 255}, false)
	}

	// Draw dropdown list if open
	if v.showDropdown && len(v.devices) > 0 {
		// Limit visible items and enable scrolling if needed
		maxVisible := 12
		visibleCount := len(v.devices)
		if visibleCount > maxVisible {
			visibleCount = maxVisible
		}

		listH := visibleCount * itemH
		vector.DrawFilledRect(screen, float32(dropdownX), float32(dropdownY+itemH), float32(dropdownW), float32(listH), v.dropdownBg, false)
		vector.StrokeRect(screen, float32(dropdownX), float32(dropdownY+itemH), float32(dropdownW), float32(listH), 1, color.RGBA{80, 80, 80, 255}, false)

		for i := 0; i < visibleCount && i < len(v.devices); i++ {
			dev := v.devices[i]
			itemY := dropdownY + itemH + i*itemH

			// Highlight
			if i == v.hoverDevice {
				vector.DrawFilledRect(screen, float32(dropdownX), float32(itemY), float32(dropdownW), float32(itemH), v.dropdownHover, false)
			} else if i == v.selectedDevice {
				vector.DrawFilledRect(screen, float32(dropdownX), float32(itemY), float32(dropdownW), float32(itemH), color.RGBA{0, 100, 80, 255}, false)
			}

			// Build display text with type prefix
			prefix := "[Mic] "
			if dev.IsLoopback {
				prefix = "[Output] "
			}
			text := prefix + truncate(dev.Name, 28)
			if dev.IsDefault {
				text += " *"
			}
			ebitenutil.DebugPrintAt(screen, text, dropdownX+8, itemY+4)
		}
	}
}

func (v *Visualizer) drawStylePanel(screen *ebiten.Image) {
	preset := colorPresets[v.lineStyle.ColorPreset]
	accentColor := v.getUIAccentColor()

	lines := []string{
		"Line Style (L to close)",
		"",
		fmt.Sprintf("Color: %s (C/Shift+C)", preset.Name),
		fmt.Sprintf("Thickness: %.1f (T/Shift+T)", v.lineStyle.Thickness),
		fmt.Sprintf("Glow: %s (G)", boolToOnOff(v.lineStyle.GlowEnabled)),
		fmt.Sprintf("Smooth Curves: %s (B)", boolToOnOff(v.smoothCurves)),
		fmt.Sprintf("Bands: %d (slider)", v.numBands),
		"",
		fmt.Sprintf("Isometric 3D: %s (I)", boolToOnOff(v.isometric.Enabled)),
	}

	// Add isometric details if enabled
	if v.isometric.Enabled {
		lines = append(lines,
			fmt.Sprintf("  Layers: %d (D/Shift+D)", v.isometric.DepthLayers),
			fmt.Sprintf("  Angle: %.0f° (A/Shift+A)", v.isometric.Angle),
			fmt.Sprintf("  Spacing: %.0f (W/Shift+W)", v.isometric.DepthSpacing),
			fmt.Sprintf("  Rotation: %.0f° (auto)", v.isometric.Rotation),
		)
	}

	// Add overlay section if not in isometric mode
	if !v.isometric.Enabled {
		lines = append(lines,
			"",
			"Overlay:",
			fmt.Sprintf("  Lines: %d (O/Shift+O)", v.overlay.Count),
			fmt.Sprintf("  Offset: %.0fpx (Up/Down)", v.overlay.Offset),
			fmt.Sprintf("  Sync: %s (Y)", boolToOnOff(v.overlay.SyncLines)),
		)
	}

	boxW := 260
	boxH := len(lines)*16 + 20
	boxX := v.width - boxW - 10
	boxY := 50

	// Semi-transparent background
	vector.DrawFilledRect(screen, float32(boxX), float32(boxY), float32(boxW), float32(boxH), color.RGBA{20, 20, 20, 230}, false)
	vector.StrokeRect(screen, float32(boxX), float32(boxY), float32(boxW), float32(boxH), 1, accentColor, false)

	// Draw color preview swatch (using dynamic accent color for rainbow mode)
	swatchX := boxX + boxW - 50
	swatchY := boxY + 32
	vector.DrawFilledRect(screen, float32(swatchX), float32(swatchY), 40, 16, accentColor, false)
	vector.StrokeRect(screen, float32(swatchX), float32(swatchY), 40, 16, 1, color.RGBA{100, 100, 100, 255}, false)

	for i, line := range lines {
		ebitenutil.DebugPrintAt(screen, line, boxX+10, boxY+10+i*16)
	}
}

func (v *Visualizer) drawHelp(screen *ebiten.Image) {
	helpLines := []string{
		"Controls:",
		"",
		"ESC/Q - Exit  F - Fullscreen",
		"H - Help  M - Settings Menu",
		"",
		"8 Visualization Modes:",
		"  Line, Bars, Rising, Spiral,",
		"  Isometric, Particles,",
		"  Circular, Wave",
		"",
		"Audio:",
		"  Dropdown - Select device",
		"  +/- Sensitivity",
		"",
		"Press M for mode-specific",
		"settings and customization",
		"",
		"R - Reset  S - Save config",
	}

	boxW := 260
	boxH := len(helpLines)*16 + 20
	boxX := 10
	boxY := 40

	// Semi-transparent background
	vector.DrawFilledRect(screen, float32(boxX), float32(boxY), float32(boxW), float32(boxH), color.RGBA{20, 20, 20, 220}, false)

	for i, line := range helpLines {
		ebitenutil.DebugPrintAt(screen, line, boxX+10, boxY+10+i*16)
	}
}

func (v *Visualizer) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func (v *Visualizer) adjustSensitivity(delta float64) {
	newSens := v.config.Sensitivity.Sensitivity + delta
	newSens = math.Max(0.1, math.Min(10.0, newSens))
	v.config.Sensitivity.Sensitivity = newSens
	if v.processor != nil {
		v.processor.SetSensitivity(newSens)
	}
	v.showStatus(fmt.Sprintf("Sensitivity: %.2f", newSens))
}

func (v *Visualizer) adjustSmoothing(delta float64) {
	newAttack := v.config.Smoothing.AttackSpeed + delta
	newAttack = math.Max(0.1, math.Min(1.0, newAttack))
	v.config.Smoothing.AttackSpeed = newAttack
	if v.processor != nil {
		v.processor.SetSmoothing(newAttack, v.config.Smoothing.DecaySpeed, v.config.Smoothing.RestDecay)
	}
	v.showStatus(fmt.Sprintf("Smoothing: %.2f", newAttack))
}

func (v *Visualizer) showStatus(msg string) {
	v.statusMessage = msg
	v.statusTimer = 120 // ~2 seconds at 60fps
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func boolToOnOff(b bool) string {
	if b {
		return "ON"
	}
	return "OFF"
}

func main() {
	configPath := flag.String("config", "config.json", "Path to configuration file")
	listDevices := flag.Bool("list-devices", false, "List available audio devices and exit")
	showVersion := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("Audio Visualizer (GUI) v%s\n", version)
		os.Exit(0)
	}

	if *listDevices {
		list, err := audio.ListDevices()
		if err != nil {
			log.Fatalf("Failed to list devices: %v", err)
		}
		fmt.Println(list)
		os.Exit(0)
	}

	// Load config
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Printf("Warning: Failed to load config: %v", err)
		cfg = config.Default()
	}
	cfg.Validate()

	// Get device list
	deviceList, err := audio.GetDevices()
	if err != nil {
		log.Printf("Warning: Failed to enumerate devices: %v", err)
	}

	// Create audio capture - default to loopback (system audio) for better out-of-box experience
	captureCfg := audio.CaptureConfig{
		SampleRate: cfg.Audio.SampleRate,
		ChunkSize:  cfg.Audio.ChunkSize,
		DeviceName: cfg.Audio.Device,
		Loopback:   true, // Default to loopback (output device) for system audio capture
	}
	capture, err := audio.NewCapture(captureCfg)
	if err != nil {
		log.Fatalf("Failed to create audio capture: %v", err)
	}

	// Start capture
	if err := capture.Start(); err != nil {
		log.Printf("Warning: Failed to start audio capture: %v", err)
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

	// Create visualizer
	vis := NewVisualizer(cfg, *configPath)
	vis.SetCapture(capture)
	vis.SetProcessor(processor)
	vis.SetDeviceList(deviceList)

	// Configure window
	ebiten.SetWindowSize(defaultWidth, defaultHeight)
	ebiten.SetWindowTitle("Audio Visualizer")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// Run
	if err := ebiten.RunGame(vis); err != nil && err != ebiten.Termination {
		log.Fatal(err)
	}

	// Cleanup
	capture.Close()
}
