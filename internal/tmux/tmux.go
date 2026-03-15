package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func IsInsideTmux() bool {
	return os.Getenv("TMUX") != ""
}

func SessionExists(name string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", name)
	return cmd.Run() == nil
}

func CreateSession(name, path string) error {
	args := []string{"new-session", "-d", "-s", name, "-c", path}
	return run(args...)
}

func NewWindow(session, name, path string) error {
	target := session
	args := []string{"new-window", "-t", target, "-n", name, "-c", path}
	return run(args...)
}

func KillWindow(session, name string) error {
	target := session + ":" + name
	return run("kill-window", "-t", target)
}

func SwitchTo(session, window string) error {
	target := session
	if window != "" {
		target = session + ":" + window
	}
	if IsInsideTmux() {
		return run("switch-client", "-t", target)
	}
	cmd := exec.Command("tmux", "attach-session", "-t", target)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type Window struct {
	Name string
	Path string
}

func ListWindows(session string) ([]Window, error) {
	out, err := runOutput("list-windows", "-t", session, "-F", "#{window_name}\t#{pane_current_path}")
	if err != nil {
		return nil, err
	}
	var windows []Window
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		w := Window{Name: parts[0]}
		if len(parts) > 1 {
			w.Path = parts[1]
		}
		windows = append(windows, w)
	}
	return windows, nil
}

func RenameWindow(session, oldName, newName string) error {
	target := session + ":" + oldName
	return run("rename-window", "-t", target, newName)
}

func run(args ...string) error {
	cmd := exec.Command("tmux", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tmux %s: %s: %w", strings.Join(args, " "), strings.TrimSpace(string(out)), err)
	}
	return nil
}

func runOutput(args ...string) (string, error) {
	cmd := exec.Command("tmux", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("tmux %s: %s: %w", strings.Join(args, " "), strings.TrimSpace(string(out)), err)
	}
	return string(out), nil
}
