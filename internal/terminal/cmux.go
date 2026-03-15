package terminal

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CmuxBackend implements Backend using the cmux CLI.
// Project = cmux workspace (sidebar entry), Branch = cmux surface (tab within workspace).
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
	// Find the project workspace by session name
	wsRef, err := c.findWorkspaceRef(session)
	if err != nil {
		return fmt.Errorf("session %q not found: %w", session, err)
	}

	// Create a new surface (tab) within the project workspace
	out, err := runCmux("new-surface", "--workspace", wsRef)
	if err != nil {
		return err
	}

	// Parse surface ref from output: "OK surface:27 pane:1 workspace:1"
	surfaceRef := parseNewSurfaceRef(out)

	// Navigate to the worktree directory
	if surfaceRef != "" {
		_, _ = runCmux("send", "--workspace", wsRef, "--surface", surfaceRef,
			fmt.Sprintf("cd %s && clear", shellEscape(path)))
		// Send Enter key to execute the command
		_, _ = runCmux("send-key", "--workspace", wsRef, "--surface", surfaceRef, "Return")
	}

	// Rename the surface tab
	renameArgs := []string{"rename-tab", "--workspace", wsRef, name}
	if surfaceRef != "" {
		renameArgs = []string{"rename-tab", "--workspace", wsRef, "--surface", surfaceRef, name}
	}
	_, err = runCmux(renameArgs...)
	return err
}

func (c *CmuxBackend) KillWindow(session, name string) error {
	wsRef, err := c.findWorkspaceRef(session)
	if err != nil {
		return fmt.Errorf("session %q not found: %w", session, err)
	}

	// Find surface by name within the workspace
	surfaces, err := c.listSurfacesRaw(wsRef)
	if err != nil {
		return err
	}

	for _, s := range surfaces {
		if s.title == name {
			_, err := runCmux("close-surface", "--workspace", wsRef, "--surface", s.ref)
			if err != nil && strings.Contains(err.Error(), "Cannot close the last surface") {
				// Last surface — close the entire workspace
				_, err = runCmux("close-workspace", "--workspace", wsRef)
			}
			return err
		}
	}
	return fmt.Errorf("surface %q not found in workspace %q", name, session)
}

func (c *CmuxBackend) SwitchTo(session, window string) error {
	wsRef, err := c.findWorkspaceRef(session)
	if err != nil {
		return fmt.Errorf("workspace %q not found", session)
	}

	// First, switch to the project workspace
	if _, err := runCmux("select-workspace", "--workspace", wsRef); err != nil {
		return err
	}

	// If a specific window (surface) is requested, select it
	if window != "" {
		surfaces, err := c.listSurfacesRaw(wsRef)
		if err != nil {
			return nil // workspace switched, surface listing failed — acceptable
		}
		for _, s := range surfaces {
			if s.title == window {
				_, err := runCmux("tab-action", "--action", "select", "--tab", s.ref, "--workspace", wsRef)
				return err
			}
		}
		return fmt.Errorf("surface %q not found in workspace %q", window, session)
	}

	return nil
}

func (c *CmuxBackend) ListWindows(session string) ([]Window, error) {
	wsRef, err := c.findWorkspaceRef(session)
	if err != nil {
		return nil, err
	}

	surfaces, err := c.listSurfacesRaw(wsRef)
	if err != nil {
		return nil, err
	}

	var windows []Window
	for _, s := range surfaces {
		windows = append(windows, Window{Name: s.title, Path: s.cwd})
	}
	return windows, nil
}

// --- workspace types and helpers ---

type cmuxWorkspace struct {
	ref      string
	title    string
	cwd      string
	selected bool
}

type cmuxSurface struct {
	ref      string // e.g., "surface:27"
	title    string
	cwd      string
	selected bool
}

func (c *CmuxBackend) findWorkspaceRef(name string) (string, error) {
	workspaces, err := c.listWorkspacesRaw()
	if err != nil {
		return "", err
	}
	for _, ws := range workspaces {
		if ws.title == name {
			return ws.ref, nil
		}
	}
	return "", fmt.Errorf("workspace %q not found", name)
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

func (c *CmuxBackend) listSurfacesRaw(wsRef string) ([]cmuxSurface, error) {
	out, err := runCmux("list-pane-surfaces", "--workspace", wsRef)
	if err != nil {
		return nil, err
	}
	var surfaces []cmuxSurface
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		s := parseCmuxSurfaceLine(line)
		if s.ref != "" {
			surfaces = append(surfaces, s)
		}
	}
	return surfaces, nil
}

// parseCmuxWorkspaceLine parses a line from `cmux list-workspaces`.
// Formats:
//
//	"  workspace:1  ~/dev/tw"
//	"* workspace:2  my-project  [selected]"
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

// parseCmuxSurfaceLine parses a line from `cmux list-pane-surfaces`.
// Formats:
//
//	"  surface:1  test-rename"
//	"* surface:27  Terminal  [selected]"
func parseCmuxSurfaceLine(line string) cmuxSurface {
	var s cmuxSurface

	s.selected = strings.HasPrefix(line, "*")

	// Remove leading "* " or "  "
	if len(line) >= 2 {
		line = line[2:]
	}

	// Remove trailing "[selected]"
	line = strings.TrimSuffix(line, "[selected]")
	line = strings.TrimRight(line, " ")

	parts := strings.SplitN(strings.TrimSpace(line), "  ", 2)
	if len(parts) < 2 {
		return s
	}

	s.ref = strings.TrimSpace(parts[0])
	s.title = strings.TrimSpace(parts[1])
	s.cwd = s.title

	return s
}

// parseNewWorkspaceRef extracts workspace ref from new-workspace output.
// Format: "OK workspace:8"
func parseNewWorkspaceRef(output string) string {
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "OK ") {
			parts := strings.Fields(line)
			for _, p := range parts {
				if strings.HasPrefix(p, "workspace:") {
					return p
				}
			}
		}
		if strings.HasPrefix(line, "workspace:") {
			return line
		}
	}
	return ""
}

// parseNewSurfaceRef extracts surface ref from new-surface output.
// Format: "OK surface:27 pane:1 workspace:1"
func parseNewSurfaceRef(output string) string {
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "OK ") {
			parts := strings.Fields(line)
			for _, p := range parts {
				if strings.HasPrefix(p, "surface:") {
					return p
				}
			}
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

func shellEscape(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

func runCmux(args ...string) (string, error) {
	cmd := exec.Command("cmux", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("cmux %s: %s: %w", strings.Join(args, " "), strings.TrimSpace(string(out)), err)
	}
	return string(out), nil
}
