# LokSTT

A lightweight, blazing-fast, and stylish GTK4 dictation daemon for Linux. It runs whisper.cpp locally, respects your privacy, and types transcribed text directly into your active window (Wayland & X11 supported).

## Features
- **Local & Private:** Everything runs on your machine using quantized `whisper.cpp` models.
- **Hardware Acceleration:** Native support for CPU, Vulkan (AMD/Intel), and CUDA (NVIDIA).
- **Beautiful GTK4 UI:** A modern, transparent, squircle-shaped overlay with a dynamic audio waveform and timer.
- **Hot-Reloading:** Change languages (99 supported) and models (Tiny to Large-v3-Turbo) on the fly without restarting.
- **Universal Paste:** Works on GNOME Wayland, wlroots, and X11 via smart fallback (`ydotool`, `wtype`, `wl-copy`, `xdotool`).
- **Runglish Ready:** Perfectly handles Russian/English mixed dictation for programmers.

## Installation (Arch Linux)

The easiest way to install is via the AUR. Choose the package that matches your hardware:

```bash
# For standard CPU (works everywhere, best for Tiny/Base/Small models)
yay -S lokstt

# For AMD / Intel GPUs (Vulkan acceleration)
yay -S lokstt-vulkan

# For NVIDIA GPUs (CUDA acceleration)
yay -S lokstt-cuda
```

*Note: The AUR package will automatically download the required highly-optimized quantized models (~1GB).*

## Configuration

**Important for GNOME Wayland users:** 
To enable automatic typing across all applications, enable the `ydotoold` service:
```bash
sudo systemctl enable --now ydotoold
```

## Usage

1. Start the background daemon (usually handled by your autostart script):
   ```bash
   lokstt &
   ```
2. Open the Settings GUI to choose your model and language:
   ```bash
   lokstt-client SETTINGS
   ```
3. Map your favorite keyboard shortcut (e.g., `F9`) in your desktop environment to:
   ```bash
   lokstt-client TOGGLE
   ```
   *Press once to start recording, press again to transcribe and paste!*
4. Map another shortcut (e.g., `Shift+F9`) to cancel a recording:
   ```bash
   lokstt-client CANCEL
   ```
