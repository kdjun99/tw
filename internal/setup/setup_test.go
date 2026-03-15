package setup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dongjunkim/tw/internal/config"
)

func TestCopyFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")

	content := []byte("hello world\n")
	if err := os.WriteFile(src, content, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := copyFile(src, dst); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(content) {
		t.Errorf("content = %q, want %q", got, content)
	}

	// Verify permissions match
	srcInfo, _ := os.Stat(src)
	dstInfo, _ := os.Stat(dst)
	if srcInfo.Mode() != dstInfo.Mode() {
		t.Errorf("mode = %v, want %v", dstInfo.Mode(), srcInfo.Mode())
	}
}

func TestRunSetup_CopyFiles(t *testing.T) {
	repoDir := t.TempDir()
	wtDir := t.TempDir()

	// Create source files
	os.WriteFile(filepath.Join(repoDir, ".env"), []byte("SECRET=123\n"), 0o644)
	os.WriteFile(filepath.Join(repoDir, ".env.local"), []byte("LOCAL=456\n"), 0o644)

	proj := &config.Project{
		Name: "test",
		Path: repoDir,
		Setup: &config.SetupConfig{
			Copy: []string{".env*"},
		},
	}

	if err := RunSetup(proj, wtDir); err != nil {
		t.Fatal(err)
	}

	// Verify files copied
	for _, name := range []string{".env", ".env.local"} {
		data, err := os.ReadFile(filepath.Join(wtDir, name))
		if err != nil {
			t.Errorf("file %s not copied: %v", name, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("file %s is empty", name)
		}
	}
}

func TestRunSetup_RunCommands(t *testing.T) {
	wtDir := t.TempDir()

	proj := &config.Project{
		Name: "test",
		Path: t.TempDir(),
		Setup: &config.SetupConfig{
			Run: []string{"touch setup_marker.txt"},
		},
	}

	if err := RunSetup(proj, wtDir); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(wtDir, "setup_marker.txt")); err != nil {
		t.Error("setup command did not execute")
	}
}

func TestRunSetup_NilSetup(t *testing.T) {
	proj := &config.Project{Name: "test", Path: t.TempDir()}
	if err := RunSetup(proj, t.TempDir()); err != nil {
		t.Errorf("nil setup should not error: %v", err)
	}
}

func TestRunSetup_CommandFailure(t *testing.T) {
	proj := &config.Project{
		Name: "test",
		Path: t.TempDir(),
		Setup: &config.SetupConfig{
			Run: []string{"false"},
		},
	}

	err := RunSetup(proj, t.TempDir())
	if err == nil {
		t.Error("expected error from failing command")
	}
}

func TestRunTeardown(t *testing.T) {
	wtDir := t.TempDir()

	// Create a file to be removed by teardown
	marker := filepath.Join(wtDir, "teardown_test.txt")
	os.WriteFile(marker, []byte("exists"), 0o644)

	proj := &config.Project{
		Name: "test",
		Path: t.TempDir(),
		Teardown: &config.TeardownConfig{
			Run: []string{"rm teardown_test.txt"},
		},
	}

	RunTeardown(proj, wtDir)

	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Error("teardown command did not execute")
	}
}

func TestRunTeardown_NilTeardown(t *testing.T) {
	proj := &config.Project{Name: "test", Path: t.TempDir()}
	// Should not panic
	RunTeardown(proj, t.TempDir())
}
