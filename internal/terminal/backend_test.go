package terminal

import (
	"os/exec"
	"testing"
)

func TestNewBackend_Tmux(t *testing.T) {
	b := NewBackend("tmux")
	if b.Name() != "tmux" {
		t.Errorf("NewBackend(tmux).Name() = %q, want %q", b.Name(), "tmux")
	}
}

func TestNewBackend_Cmux(t *testing.T) {
	b := NewBackend("cmux")
	if b.Name() != "cmux" {
		t.Errorf("NewBackend(cmux).Name() = %q, want %q", b.Name(), "cmux")
	}
}

func TestNewBackend_Auto(t *testing.T) {
	b := NewBackend("")
	name := b.Name()
	if name != "tmux" && name != "cmux" {
		t.Errorf("NewBackend('').Name() = %q, want tmux or cmux", name)
	}
}

func TestNewBackend_AutoSameAsDefault(t *testing.T) {
	b1 := NewBackend("auto")
	b2 := NewBackend("")
	if b1.Name() != b2.Name() {
		t.Errorf("NewBackend(auto) = %q, NewBackend('') = %q, should be equal", b1.Name(), b2.Name())
	}
}

func TestTmuxBackend_Interface(t *testing.T) {
	var b Backend = &TmuxBackend{}
	if b.Name() != "tmux" {
		t.Errorf("TmuxBackend.Name() = %q", b.Name())
	}
}

func TestCmuxBackend_Interface(t *testing.T) {
	var b Backend = &CmuxBackend{}
	if b.Name() != "cmux" {
		t.Errorf("CmuxBackend.Name() = %q", b.Name())
	}
}

func TestHasCmux(t *testing.T) {
	_, err := exec.LookPath("cmux")
	expected := err == nil
	if hasCmux() != expected {
		t.Errorf("hasCmux() = %v, expected %v", hasCmux(), expected)
	}
}

func TestTmuxBackend_SessionExists_NonExistent(t *testing.T) {
	b := &TmuxBackend{}
	// A session with this random name should not exist
	if b.SessionExists("tw-test-nonexistent-session-xyz") {
		t.Error("SessionExists should return false for nonexistent session")
	}
}

func TestCmuxBackend_IsActive_OutsideCmux(t *testing.T) {
	// When running tests outside cmux, CMUX_WORKSPACE_ID is not set
	b := &CmuxBackend{}
	// We can't assert true/false since it depends on environment,
	// but it should not panic
	_ = b.IsActive()
}

func TestParseCmuxWorkspaceLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		wantRef  string
		wantTitle string
		wantSel  bool
	}{
		{
			name:     "normal workspace",
			line:     "  workspace:1  ~/dev/tw",
			wantRef:  "workspace:1",
			wantTitle: "~/dev/tw",
			wantSel:  false,
		},
		{
			name:     "selected workspace",
			line:     "* workspace:2  ~/dev/tw  [selected]",
			wantRef:  "workspace:2",
			wantTitle: "~/dev/tw",
			wantSel:  true,
		},
		{
			name:     "renamed workspace",
			line:     "  workspace:3  my-project",
			wantRef:  "workspace:3",
			wantTitle: "my-project",
			wantSel:  false,
		},
		{
			name:     "renamed and selected",
			line:     "* workspace:5  feature-login  [selected]",
			wantRef:  "workspace:5",
			wantTitle: "feature-login",
			wantSel:  true,
		},
		{
			name:     "path with spaces in title",
			line:     "  workspace:4  ~/dev/my app",
			wantRef:  "workspace:4",
			wantTitle: "~/dev/my app",
			wantSel:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := parseCmuxWorkspaceLine(tt.line)
			if ws.ref != tt.wantRef {
				t.Errorf("ref = %q, want %q", ws.ref, tt.wantRef)
			}
			if ws.title != tt.wantTitle {
				t.Errorf("title = %q, want %q", ws.title, tt.wantTitle)
			}
			if ws.selected != tt.wantSel {
				t.Errorf("selected = %v, want %v", ws.selected, tt.wantSel)
			}
		})
	}
}

func TestParseCmuxSurfaceLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		wantRef  string
		wantTitle string
		wantSel  bool
	}{
		{
			name:     "normal surface",
			line:     "  surface:1  test-rename",
			wantRef:  "surface:1",
			wantTitle: "test-rename",
			wantSel:  false,
		},
		{
			name:     "selected surface",
			line:     "* surface:27  Terminal  [selected]",
			wantRef:  "surface:27",
			wantTitle: "Terminal",
			wantSel:  true,
		},
		{
			name:     "surface with path title",
			line:     "  surface:5  ~/dev/myapp",
			wantRef:  "surface:5",
			wantTitle: "~/dev/myapp",
			wantSel:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := parseCmuxSurfaceLine(tt.line)
			if s.ref != tt.wantRef {
				t.Errorf("ref = %q, want %q", s.ref, tt.wantRef)
			}
			if s.title != tt.wantTitle {
				t.Errorf("title = %q, want %q", s.title, tt.wantTitle)
			}
			if s.selected != tt.wantSel {
				t.Errorf("selected = %v, want %v", s.selected, tt.wantSel)
			}
		})
	}
}

func TestParseNewSurfaceRef(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   string
	}{
		{
			name:   "OK with surface ref",
			output: "OK surface:27 pane:1 workspace:1\n",
			want:   "surface:27",
		},
		{
			name:   "no surface in output",
			output: "OK\n",
			want:   "",
		},
		{
			name:   "empty",
			output: "",
			want:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseNewSurfaceRef(tt.output)
			if got != tt.want {
				t.Errorf("parseNewSurfaceRef() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestShellEscape(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/dev/myapp", "'/dev/myapp'"},
		{"/path/with spaces", "'/path/with spaces'"},
		{"it's a test", "'it'\\''s a test'"},
	}
	for _, tt := range tests {
		got := shellEscape(tt.input)
		if got != tt.want {
			t.Errorf("shellEscape(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseNewWorkspaceRef(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   string
	}{
		{
			name:   "OK with workspace ref",
			output: "OK workspace:8\n",
			want:   "workspace:8",
		},
		{
			name:   "direct workspace ref",
			output: "workspace:3\n",
			want:   "workspace:3",
		},
		{
			name:   "no ref in output",
			output: "OK\n",
			want:   "",
		},
		{
			name:   "empty output",
			output: "",
			want:   "",
		},
		{
			name:   "multiline with ref",
			output: "some info\nworkspace:5\ndone\n",
			want:   "workspace:5",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseNewWorkspaceRef(tt.output)
			if got != tt.want {
				t.Errorf("parseNewWorkspaceRef() = %q, want %q", got, tt.want)
			}
		})
	}
}
