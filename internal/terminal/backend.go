package terminal

// Backend defines the interface for terminal multiplexer backends (tmux, cmux).
type Backend interface {
	// Name returns the backend identifier ("tmux" or "cmux").
	Name() string

	// IsActive returns true if running inside this backend's environment.
	IsActive() bool

	// SessionExists checks if a session/workspace group exists.
	SessionExists(name string) bool

	// CreateSession creates a new session/workspace at the given path.
	CreateSession(name, path string) error

	// NewWindow creates a new window/workspace tab with the given name and path.
	NewWindow(session, name, path string) error

	// KillWindow closes a window/workspace tab.
	KillWindow(session, name string) error

	// SwitchTo switches to a session (optionally a specific window).
	SwitchTo(session, window string) error

	// ListWindows returns all windows in a session.
	ListWindows(session string) ([]Window, error)
}

// Window represents a terminal window/tab.
type Window struct {
	Name string
	Path string
}
