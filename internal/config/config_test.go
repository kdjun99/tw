package config

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestConfig(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	return tmpDir, func() { os.Setenv("HOME", origHome) }
}

func TestLoadEmptyConfig(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(cfg.Projects))
	}
}

func TestAddAndFindProject(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	p := Project{
		Name:          "myapp",
		Path:          "/dev/myapp",
		DefaultBranch: "main",
	}
	if err := cfg.AddProject(p); err != nil {
		t.Fatal(err)
	}

	// Find existing
	found := cfg.FindProject("myapp")
	if found == nil {
		t.Fatal("FindProject returned nil")
	}
	if found.Path != "/dev/myapp" {
		t.Errorf("path = %q, want %q", found.Path, "/dev/myapp")
	}

	// Find non-existing
	if cfg.FindProject("nope") != nil {
		t.Error("FindProject(nope) should return nil")
	}
}

func TestAddDuplicateProject(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	cfg, _ := Load()
	p := Project{Name: "myapp", Path: "/dev/myapp", DefaultBranch: "main"}
	if err := cfg.AddProject(p); err != nil {
		t.Fatal(err)
	}
	if err := cfg.AddProject(p); err == nil {
		t.Error("expected error for duplicate project")
	}
}

func TestRemoveProject(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	cfg, _ := Load()
	cfg.AddProject(Project{Name: "a", Path: "/a", DefaultBranch: "main"})
	cfg.AddProject(Project{Name: "b", Path: "/b", DefaultBranch: "main"})

	if err := cfg.RemoveProject("a"); err != nil {
		t.Fatal(err)
	}
	if cfg.FindProject("a") != nil {
		t.Error("project 'a' should be removed")
	}
	if cfg.FindProject("b") == nil {
		t.Error("project 'b' should still exist")
	}
}

func TestRemoveNonexistentProject(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	cfg, _ := Load()
	if err := cfg.RemoveProject("nope"); err == nil {
		t.Error("expected error removing nonexistent project")
	}
}

func TestSaveAndReload(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	cfg, _ := Load()
	cfg.AddProject(Project{
		Name:          "myapp",
		Path:          "/dev/myapp",
		DefaultBranch: "develop",
		Setup: &SetupConfig{
			Copy: []string{".env"},
			Run:  []string{"npm install"},
		},
		Teardown: &TeardownConfig{
			Run: []string{"docker compose down"},
		},
	})

	// Reload from disk
	cfg2, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	p := cfg2.FindProject("myapp")
	if p == nil {
		t.Fatal("project not found after reload")
	}
	if p.DefaultBranch != "develop" {
		t.Errorf("defaultBranch = %q, want %q", p.DefaultBranch, "develop")
	}
	if p.Setup == nil || len(p.Setup.Copy) != 1 || p.Setup.Copy[0] != ".env" {
		t.Errorf("setup.copy not preserved: %+v", p.Setup)
	}
	if p.Teardown == nil || len(p.Teardown.Run) != 1 {
		t.Errorf("teardown not preserved: %+v", p.Teardown)
	}
}

func TestResolveWorktreeDir(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		name        string
		project     Project
		wantSuffix  string
		wantExact   string
	}{
		{
			name:       "default",
			project:    Project{Name: "myapp"},
			wantSuffix: filepath.Join(".tw", "myapp"),
		},
		{
			name:      "custom",
			project:   Project{Name: "myapp", WorktreeDir: "/custom/path"},
			wantExact: "/custom/path",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.project.ResolveWorktreeDir()
			if tt.wantExact != "" {
				if got != tt.wantExact {
					t.Errorf("got %q, want %q", got, tt.wantExact)
				}
			} else {
				want := filepath.Join(home, tt.wantSuffix)
				if got != want {
					t.Errorf("got %q, want %q", got, want)
				}
			}
		})
	}
}
