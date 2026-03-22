# LokSTT

A blazing-fast, privacy-first GTK4 dictation daemon for Linux. It records audio via a hotkey, transcribes it locally using `whisper.cpp`, and pastes the text directly into your active window.

## Architecture

LokSTT is split into two parts:
1. **The Daemon (`lokstt`)**: Runs in the background, holds the AI model in memory for instant transcription, and manages the GTK4 UI.
2. **The Client (`lokstt-client`)**: A lightweight CLI tool that you bind to your keyboard shortcuts to send commands to the daemon.

## Features

- **Fully Local**: Uses quantized Whisper models. No cloud, no subscriptions, complete privacy.
- **Runglish & Multi-Language**: Perfectly handles mixed Russian/English speech for programmers, and supports 99 languages out of the box.
- **Universal Paster**: Works seamlessly across GNOME Wayland, wlroots, and X11. It gracefully falls back between `ydotool`, `wtype`, and `xdotool` to ensure text is pasted regardless of your keyboard layout.
- **Modern UI**: Features a transparent GTK4 recording overlay with a dynamic audio waveform, and a clean settings window for on-the-fly model and language switching.
- **Hardware Acceleration**: Built-in support for CPU and Vulkan backends.

## Building from Source

If you are not using the AUR packages, you can build LokSTT manually.

### Dependencies
Ensure you have the following installed on your system:
- Go (1.21+)
- GTK4 (`gtk4`)
- PortAudio (`portaudio`)
- pkg-config (`pkgconf`)
- `ydotool`, `wl-clipboard`, `xdotool` (for pasting functionality)

### Build Instructions
```bash
# Enable C++11 for whisper.cpp bindings
export CGO_CXXFLAGS="-std=c++11"

# Build the daemon
go build -o lokstt main.go

# Build the client
cd client
go build -o lokstt-client main.go
```

*(Note: For Vulkan support, you need to add `-DGGML_USE_VULKAN=1` to CFLAGS and `-lvulkan` to LDFLAGS).*

## Usage

1. **Start the daemon**: If you installed via AUR, it starts automatically on login. Otherwise, run `./lokstt &` in the background.
2. **Configure**: Open the settings to select your preferred model and language:
   ```bash
   lokstt-client SETTINGS
   ```
3. **Record**: Bind a shortcut in your Desktop Environment (e.g., `F9`) to:
   ```bash
   lokstt-client TOGGLE
   ```
   *Press once to start recording, press again to stop, transcribe, and paste.*
4. **Cancel**: Bind another shortcut (e.g., `Shift+F9`) to cancel an ongoing recording without pasting:
   ```bash
   lokstt-client CANCEL
   ```

## Configuration

**Important for GNOME Wayland users:** 
To enable automatic typing across all applications, you must enable the `ydotoold` service, as GNOME blocks standard virtual keyboard protocols:
```bash
sudo systemctl enable --now ydotoold
```
