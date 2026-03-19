package paster

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"time"
)

const defaultTimeout = 2 * time.Second

// PasteText intelligently detects the display server and available tools
// to paste text or emulate keyboard input.
func PasteText(text string) error {
	if isWayland() {
		return pasteWayland(text)
	}
	return pasteX11(text)
}

func isWayland() bool {
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		return true
	}
	if os.Getenv("XDG_SESSION_TYPE") == "wayland" {
		return true
	}
	return false
}

func pasteWayland(text string) error {
	if hasTool("wl-copy") {
		_ = runCmdWithStdin(text, "wl-copy")
		_ = runCmdWithStdin(text, "wl-copy", "-p")
		
		time.Sleep(100 * time.Millisecond)
		
		if hasTool("ydotool") {
			if err := runCmd("ydotool", "key", "42:1", "110:1", "110:0", "42:0"); err == nil {
				return nil
			}
		}
		
		if hasTool("wtype") {
			if err := runCmd("wtype", "-M", "shift", "-k", "Insert", "-m", "shift"); err == nil {
				return nil
			}
		}
		
		_ = runCmd("notify-send", "Copied to clipboard (install ydotoold to auto-paste)")
		return nil
	}

	return errors.New("wl-copy is missing")
}

func pasteX11(text string) error {
	if hasTool("xclip") {
		_ = runCmdWithStdin(text, "xclip", "-selection", "clipboard")
		_ = runCmdWithStdin(text, "xclip", "-selection", "primary")
		
		time.Sleep(100 * time.Millisecond)
		
		if hasTool("xdotool") {
			if err := runCmd("xdotool", "key", "shift+Insert"); err == nil {
				return nil
			}
		}
		
		_ = runCmd("notify-send", "Copied to clipboard")
		return nil
	}
	return errors.New("xclip is missing")
}

func hasTool(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func runCmd(name string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.Run()
}

func runCmdWithStdin(stdin string, name string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = bytes.NewBufferString(stdin)
	return cmd.Run()
}
