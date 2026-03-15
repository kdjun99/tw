package terminal

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CmuxBackend implements Backend using the cmux CLI.
type CmuxBackend struct{}

func (c *CmuxBackend) Name() string { return "cmux" }

func (c *CmuxBackend) IsActive() bool {
	return os.Getenv("CMUX_WORKSPACE_ID") != ""
}

func (c *CmuxBackend) SessionExists(name string) bool {
	workspaces, err := c.listWorkspacesRaw()
	if err != nil {
		return false
	}
	for _, ws := range workspaces {
		if ws.title == name {
			return true
		}
	}
	return false
}

func (c *CmuxBackend) CreateSession(name, path string) error {
	out, err := runCmux("new-workspace", "--cwd", path)
	if err != nil {
		return err
	}
	ref := parseNewWorkspaceRef(out)
	return c.renameWorkspace(ref, name)
}

func (c *CmuxBackend) NewWindow(session, name, path string) error {
	out, err := runCmux("new-workspace", "--cwd", path)
	if err != nil {
		return err
	}
	ref := parseNewWorkspaceRef(out)
	return c.renameWorkspace(ref, name)
}

func (c *CmuxBackend) KillWindow(session, name string) error {
	workspaces, err := c.listWorkspacesRaw()
	if err != nil {
		return err
	}
	for _, ws := range workspaces {
		if ws.title == name {
			_, err := runCmux("close-workspace", "--workspace", ws.ref)
			return err
		}
	}
	return fmt.Errorf("workspace %q not found", name)
}

func (c *CmuxBackend) SwitchTo(session, window string) error {
	target := session
	if window != "" {
		target = window
	}

	workspaces, err := c.listWorkspacesRaw()
	if err != nil {
		return err
	}
	for _, ws := range workspaces {
		if ws.title == target {
			_, err := runCmux("select-workspace", "--workspace", ws.ref)
			return err
		}
	}
	return fmt.Errorf("workspace %q not found", target)
}

func (c *CmuxBackend) ListWindows(session string) ([]Window, error) {
	workspaces, err := c.listWorkspacesRaw()
	if err != nil {
		return nil, err
	}
	var windows []Window
	for _, ws := range workspaces {
		windows = append(windows, Window{Name: ws.title, Path: ws.cwd})
	}
	return windows, nil
}

type cmuxWorkspace struct {
	ref      string // e.g., "workspace:1"
	title    string // custom name or directory path
	cwd      string // always the directory path
	selected bool
}

func (c *CmuxBackend) listWorkspacesRaw() ([]cmuxWorkspace, error) {
	out, err := runCmux("list-workspaces")
	if err != nil {
		return nil, err
	}
	var workspaces []cmuxWorkspace
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		ws := parseCmuxWorkspaceLine(line)
		if ws.ref != "" {
			workspaces = append(workspaces, ws)
		}
	}
	return workspaces, nil
}

// parseCmuxWorkspaceLine parses a line from `cmux list-workspaces`.
// Formats:
//
//	"  workspace:1  ~/dev/tw"
//	"* workspace:2  ~/dev/tw  [selected]"
//	"  workspace:3  my-project"           (renamed workspace)
func parseCmuxWorkspaceLine(line string) cmuxWorkspace {
	var ws cmuxWorkspace

	ws.selected = strings.HasPrefix(line, "*")

	// Remove leading "* " or "  "
	if len(line) >= 2 {
		line = line[2:]
	}

	// Remove trailing "[selected]"
	line = strings.TrimSuffix(line, "[selected]")
	line = strings.TrimRight(line, " ")

	// Split into ref and title (separated by two or more spaces)
	parts := strings.SplitN(strings.TrimSpace(line), "  ", 2)
	if len(parts) < 2 {
		return ws
	}

	ws.ref = strings.TrimSpace(parts[0])
	ws.title = strings.TrimSpace(parts[1])
	ws.cwd = ws.title

	return ws
}

// parseNewWorkspaceRef extracts workspace ref from new-workspace output.
// Format: "OK workspace:8"
func parseNewWorkspaceRef(output string) string {
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		// "OK workspace:8" format
		if strings.HasPrefix(line, "OK ") {
			parts := strings.Fields(line)
			for _, p := range parts {
				if strings.HasPrefix(p, "workspace:") {
					return p
				}
			}
		}
		// Direct "workspace:8" format
		if strings.HasPrefix(line, "workspace:") {
			return line
		}
	}
	return ""
}

func (c *CmuxBackend) renameWorkspace(ref, name string) error {
	args := []string{"workspace-action", "--action", "rename", "--title", name}
	if ref != "" {
		args = append(args, "--workspace", ref)
	}
	_, err := runCmux(args...)
	return err
}

func runCmux(args ...string) (string, error) {
	cmd := exec.Command("cmux", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("cmux %s: %s: %w", strings.Join(args, " "), strings.TrimSpace(string(out)), err)
	}
	return string(out), nil
}
