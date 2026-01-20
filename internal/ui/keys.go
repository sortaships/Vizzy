package ui

import (
	"github.com/charmbracelet/bubbletea"
)

// KeyBinding represents a keyboard shortcut
type KeyBinding struct {
	Key         string
	Description string
}

// KeyBindings holds all key bindings
var KeyBindings = []KeyBinding{
	{Key: "q/Esc", Description: "Quit"},
	{Key: "v", Description: "Toggle visualization mode"},
	{Key: "d", Description: "Cycle audio devices"},
	{Key: "+/=", Description: "Increase sensitivity"},
	{Key: "-", Description: "Decrease sensitivity"},
	{Key: "[", Description: "Decrease smoothing"},
	{Key: "]", Description: "Increase smoothing"},
	{Key: "p", Description: "Toggle peak indicators"},
	{Key: "g", Description: "Toggle gradient coloring"},
	{Key: "m", Description: "Toggle mirror mode"},
	{Key: "h/?", Description: "Toggle help"},
	{Key: "s", Description: "Save config"},
	{Key: "r", Description: "Reset to defaults"},
}

// HandleKeyPress processes a key press and returns the appropriate action
func HandleKeyPress(msg tea.KeyMsg) Action {
	switch msg.String() {
	case "q", "esc", "ctrl+c":
		return ActionQuit

	case "v":
		return ActionToggleMode

	case "d":
		return ActionCycleDevice

	case "+", "=":
		return ActionIncreaseSensitivity

	case "-":
		return ActionDecreaseSensitivity

	case "[":
		return ActionDecreaseSmoothing

	case "]":
		return ActionIncreaseSmoothing

	case "p":
		return ActionTogglePeaks

	case "g":
		return ActionToggleGradient

	case "m":
		return ActionToggleMirror

	case "h", "?":
		return ActionToggleHelp

	case "s":
		return ActionSaveConfig

	case "r":
		return ActionReset

	default:
		return ActionNone
	}
}

// Action represents a user action
type Action int

const (
	ActionNone Action = iota
	ActionQuit
	ActionToggleMode
	ActionCycleDevice
	ActionIncreaseSensitivity
	ActionDecreaseSensitivity
	ActionIncreaseSmoothing
	ActionDecreaseSmoothing
	ActionTogglePeaks
	ActionToggleGradient
	ActionToggleMirror
	ActionToggleHelp
	ActionSaveConfig
	ActionReset
)

// String returns a string representation of the action
func (a Action) String() string {
	switch a {
	case ActionQuit:
		return "Quit"
	case ActionToggleMode:
		return "Toggle Mode"
	case ActionCycleDevice:
		return "Cycle Device"
	case ActionIncreaseSensitivity:
		return "Increase Sensitivity"
	case ActionDecreaseSensitivity:
		return "Decrease Sensitivity"
	case ActionIncreaseSmoothing:
		return "Increase Smoothing"
	case ActionDecreaseSmoothing:
		return "Decrease Smoothing"
	case ActionTogglePeaks:
		return "Toggle Peaks"
	case ActionToggleGradient:
		return "Toggle Gradient"
	case ActionToggleMirror:
		return "Toggle Mirror"
	case ActionToggleHelp:
		return "Toggle Help"
	case ActionSaveConfig:
		return "Save Config"
	case ActionReset:
		return "Reset"
	default:
		return "None"
	}
}
