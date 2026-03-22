# LokSTT - AI Agent Context File

**Hello fellow AI Agent!** This file contains the concentrated architectural context, quirks, and constraints of the LokSTT project. Read this before modifying the codebase to understand the "why" behind the code.

## Project Overview
LokSTT is a Linux dictation tool written in **Go**, utilizing **GTK4** (`diamondburned/gotk4`) for the UI and **whisper.cpp** (`ggerganov/whisper.cpp/bindings/go`) for local speech-to-text.

## Core Components
- **Daemon (`main.go`)**: Manages the CGO whisper context, listens to a Unix socket (`/tmp/lokstt.sock`), and handles threaded audio recording via PortAudio.
- **Client (`client/main.go`)**: A dummy CLI that sends RPC-like commands (`TOGGLE`, `CANCEL`, `SETTINGS`) to the daemon's socket.
- **UI (`ui/`)**:
  - `overlay.go`: A borderless, transparent, squircle GTK4 window acting as a HUD. It reads an audio level channel to animate a 5-bar waveform.
  - `ui.go`: A settings window with dropdowns for Model and Language selection. Settings are saved to `~/.config/lokstt/config.json`.
- **Paster (`paster/paster.go`)**: The module responsible for inserting text. It uses `wl-copy`/`xclip` to put text in the clipboard, then uses `ydotool` (or `xdotool`/`wtype`) to emulate a `Shift+Insert` keystroke. This solves keyboard layout dependency issues.

## Critical Technical Constraints & Quirks (DO NOT BREAK)

1. **Memory Leaks in Whisper Bindings**: 
   When hot-reloading a model (changing models in the settings UI), you **MUST explicitly call `Free()` or `Close()`** on the old C CGO whisper context. If the pointer is lost without freeing, RAM usage will explode on every setting change.

2. **GNOME Wayland Security**: 
   GNOME's Mutter compositor aggressively blocks virtual keystrokes (like `wtype`). We rely on `ydotool` (which uses `/dev/uinput` via the `ydotoold` daemon) as the primary Wayland fallback. Do not attempt to replace this with pure Wayland protocols without testing on GNOME.

3. **GTK4 Transparency on Wayland**: 
   True transparency requires removing the `.background` CSS class and hiding window decorations. This is implemented in `overlay.go`. Modifying the window setup may result in a black/white solid background instead of a transparent one.

4. **GTK4 DropDown Search**: 
   GTK4 string lists require a `PropertyExpression` targeting `GTypeStringObject` to make the search feature work in `GtkDropDown`. This boilerplate is intentional in `ui.go`.

5. **Concurrency & Audio**: 
   Audio recording, waveform UI updates, and Whisper transcription run in separate goroutines. Ensure channels are closed cleanly on `CANCEL` commands to avoid deadlocks. The overlay uses `glib.IdleAdd` to safely update the GTK4 UI from the audio goroutine.

## Build Variants
The project supports multiple inference backends via standard CGO flags (see the `packaging/` directory for exact PKGBUILD instructions):
- **CPU**: Default (`CGO_CXXFLAGS="-std=c++11"`).
- **Vulkan**: Requires `CGO_CFLAGS="-DGGML_USE_VULKAN=1"` and `CGO_LDFLAGS="-lvulkan"`.
- **CUDA**: Requires `CGO_CFLAGS="-DGGML_USE_CUDA=1"` and appropriate linking.
