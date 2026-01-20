# Go Audio Visualizer

Real-time audio visualizer with waveform display, isometric 3D mode, and customizable themes.

## Requirements

- Go 1.21+
- GCC (for CGO/audio)
  - **Windows**: MSYS2 MinGW64 (`pacman -S mingw-w64-x86_64-gcc`)
  - **macOS**: Xcode Command Line Tools
  - **Linux**: `sudo apt install build-essential libasound2-dev`

## Install

```bash
git clone <repo>
cd go_audio_visualizer
go mod tidy
```

## Build & Run

```bash
# Windows (MSYS2/MinGW)
PATH="/c/msys64/mingw64/bin:$PATH" CGO_ENABLED=1 go build ./cmd/visualizer-gui
./visualizer-gui.exe

# macOS/Linux
CGO_ENABLED=1 go build ./cmd/visualizer-gui
./visualizer-gui
```

## Controls

| Key | Action |
|-----|--------|
| `H` | Help |
| `M` | Settings menu |
| `L` | Style panel |
| `I` | Toggle isometric 3D |
| `C` | Cycle colors |
| `G` | Toggle glow |
| `T` | Line thickness |
| `O` | Overlay lines |
| `+/-` | Sensitivity |
| `F` | Fullscreen |
| `R` | Reset |
| `S` | Save config |
| `Q/Esc` | Quit |

## Features

- 23 color themes (including Rainbow)
- Isometric 3D mode with auto-rotation
- Multi-line overlays
- Smooth curve rendering
- System audio capture (loopback)
- Configurable frequency bands
