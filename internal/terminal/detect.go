package terminal

import (
	"os/exec"
)

// NewBackend creates a Backend based on the given name.
// Valid names: "tmux", "cmux", "auto" (or empty for auto-detect).
func NewBackend(name string) Backend {
	switch name {
	case "tmux":
		return &TmuxBackend{}
	case "cmux":
		return &CmuxBackend{}
	default:
		return detectBackend()
	}
}

// detectBackend auto-detects the best available backend.
// Priority: cmux (if inside cmux) > tmux (if inside tmux) > cmux (if available) > tmux.
func detectBackend() Backend {
	cmux := &CmuxBackend{}
	tmux := &TmuxBackend{}

	// Prefer the backend we're currently inside
	if cmux.IsActive() {
		return cmux
	}
	if tmux.IsActive() {
		return tmux
	}

	// Fall back to whichever is installed
	if hasCmux() {
		return cmux
	}
	return tmux
}

func hasCmux() bool {
	_, err := exec.LookPath("cmux")
	return err == nil
}
