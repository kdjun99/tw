package terminal

import (
	"github.com/dongjunkim/tw/internal/tmux"
)

// TmuxBackend wraps the existing tmux package as a Backend implementation.
type TmuxBackend struct{}

func (t *TmuxBackend) Name() string { return "tmux" }

func (t *TmuxBackend) IsActive() bool { return tmux.IsInsideTmux() }

func (t *TmuxBackend) SessionExists(name string) bool {
	return tmux.SessionExists(name)
}

func (t *TmuxBackend) CreateSession(name, path string) error {
	return tmux.CreateSession(name, path)
}

func (t *TmuxBackend) NewWindow(session, name, path string) error {
	return tmux.NewWindow(session, name, path)
}

func (t *TmuxBackend) KillWindow(session, name string) error {
	return tmux.KillWindow(session, name)
}

func (t *TmuxBackend) SwitchTo(session, window string) error {
	return tmux.SwitchTo(session, window)
}

func (t *TmuxBackend) RenameWindow(session, oldName, newName string) error {
	return tmux.RenameWindow(session, oldName, newName)
}

func (t *TmuxBackend) ListWindows(session string) ([]Window, error) {
	windows, err := tmux.ListWindows(session)
	if err != nil {
		return nil, err
	}
	result := make([]Window, len(windows))
	for i, w := range windows {
		result[i] = Window{Name: w.Name, Path: w.Path}
	}
	return result, nil
}
