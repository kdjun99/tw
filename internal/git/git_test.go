package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestResolveWorktreePath(t *testing.T) {
	tests := []struct {
		baseDir string
		branch  string
		want    string
	}{
		{"/home/user/.tw/myapp", "feature/login", "/home/user/.tw/myapp/feature-login"},
		{"/home/user/.tw/myapp", "main", "/home/user/.tw/myapp/main"},
		{"/home/user/.tw/myapp", "fix/bug/nested", "/home/user/.tw/myapp/fix-bug-nested"},
	}
	for _, tt := range tests {
		got := ResolveWorktreePath(tt.baseDir, tt.branch)
		if got != tt.want {
			t.Errorf("ResolveWorktreePath(%q, %q) = %q, want %q", tt.baseDir, tt.branch, got, tt.want)
		}
	}
}

// initTestRepo creates a temporary bare-style git repo with an initial commit.
func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init", "-b", "main"},
		{"config", "user.email", "test@test.com"},
		{"config", "user.name", "Test"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s: %v", args, out, err)
		}
	}
	// Create initial commit
	f := filepath.Join(dir, "README.md")
	if err := os.WriteFile(f, []byte("# test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"add", "."},
		{"commit", "-m", "init"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s: %v", args, out, err)
		}
	}
	return dir
}

func TestCurrentBranch(t *testing.T) {
	repo := initTestRepo(t)
	branch := CurrentBranch(repo)
	if branch != "main" {
		t.Errorf("CurrentBranch() = %q, want %q", branch, "main")
	}
}

func TestBranchExists(t *testing.T) {
	repo := initTestRepo(t)

	if !BranchExists(repo, "main") {
		t.Error("BranchExists(main) = false, want true")
	}
	if BranchExists(repo, "nonexistent") {
		t.Error("BranchExists(nonexistent) = true, want false")
	}
}

func TestListBranches(t *testing.T) {
	repo := initTestRepo(t)

	// Create another branch
	cmd := exec.Command("git", "branch", "develop")
	cmd.Dir = repo
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git branch develop: %s: %v", out, err)
	}

	branches, err := ListBranches(repo)
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]bool{"main": true, "develop": true}
	got := map[string]bool{}
	for _, b := range branches {
		got[b] = true
	}
	for k := range want {
		if !got[k] {
			t.Errorf("ListBranches missing %q", k)
		}
	}
}

func TestListWorktrees(t *testing.T) {
	repo := initTestRepo(t)
	worktrees, err := ListWorktrees(repo)
	if err != nil {
		t.Fatal(err)
	}

	if len(worktrees) != 1 {
		t.Fatalf("expected 1 worktree, got %d", len(worktrees))
	}
	if worktrees[0].Branch != "main" {
		t.Errorf("worktree branch = %q, want %q", worktrees[0].Branch, "main")
	}
	if worktrees[0].Bare {
		t.Error("worktree should not be bare")
	}
}

func TestAddAndRemoveWorktree(t *testing.T) {
	repo := initTestRepo(t)
	wtPath := filepath.Join(t.TempDir(), "feature-test")

	// Add worktree with new branch
	if err := AddWorktree(repo, wtPath, "feature/test", "main"); err != nil {
		t.Fatalf("AddWorktree: %v", err)
	}

	// Verify worktree was created
	if _, err := os.Stat(wtPath); err != nil {
		t.Fatalf("worktree dir not created: %v", err)
	}

	// Verify it shows in list
	worktrees, err := ListWorktrees(repo)
	if err != nil {
		t.Fatal(err)
	}
	if len(worktrees) != 2 {
		t.Fatalf("expected 2 worktrees, got %d", len(worktrees))
	}

	// Verify branch exists
	if !BranchExists(repo, "feature/test") {
		t.Error("branch feature/test should exist after AddWorktree")
	}

	// Remove worktree
	if err := RemoveWorktree(repo, wtPath); err != nil {
		t.Fatalf("RemoveWorktree: %v", err)
	}

	worktrees, err = ListWorktrees(repo)
	if err != nil {
		t.Fatal(err)
	}
	if len(worktrees) != 1 {
		t.Fatalf("expected 1 worktree after remove, got %d", len(worktrees))
	}
}

func TestAddWorktreeExisting(t *testing.T) {
	repo := initTestRepo(t)

	// Create branch first
	cmd := exec.Command("git", "branch", "existing-branch")
	cmd.Dir = repo
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git branch: %s: %v", out, err)
	}

	wtPath := filepath.Join(t.TempDir(), "existing")
	if err := AddWorktreeExisting(repo, wtPath, "existing-branch"); err != nil {
		t.Fatalf("AddWorktreeExisting: %v", err)
	}

	branch := CurrentBranch(wtPath)
	if branch != "existing-branch" {
		t.Errorf("worktree branch = %q, want %q", branch, "existing-branch")
	}
}

func TestRemoveWorktreeWithUncommittedChanges(t *testing.T) {
	repo := initTestRepo(t)
	wtPath := filepath.Join(t.TempDir(), "dirty-wt")

	if err := AddWorktree(repo, wtPath, "dirty-branch", "main"); err != nil {
		t.Fatalf("AddWorktree: %v", err)
	}

	// Create uncommitted change
	if err := os.WriteFile(filepath.Join(wtPath, "dirty.txt"), []byte("uncommitted\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "add", "dirty.txt")
	cmd.Dir = wtPath
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add: %s: %v", out, err)
	}

	// RemoveWorktree should fail without --force
	err := RemoveWorktree(repo, wtPath)
	if err == nil {
		t.Error("RemoveWorktree should fail with uncommitted changes")
	}
}

func TestGetDiffStat_Clean(t *testing.T) {
	repo := initTestRepo(t)
	stat := GetDiffStat(repo)
	if stat.Added != 0 || stat.Removed != 0 {
		t.Errorf("clean repo: got +%d -%d, want +0 -0", stat.Added, stat.Removed)
	}
}

func TestGetDiffStat_Unstaged(t *testing.T) {
	repo := initTestRepo(t)

	// Modify existing file (unstaged)
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("# updated\nline2\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	stat := GetDiffStat(repo)
	if stat.Added == 0 {
		t.Error("expected Added > 0 for unstaged change")
	}
}

func TestGetDiffStat_Staged(t *testing.T) {
	repo := initTestRepo(t)

	// Create and stage a new file
	if err := os.WriteFile(filepath.Join(repo, "new.txt"), []byte("new\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "add", "new.txt")
	cmd.Dir = repo
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add: %s: %v", out, err)
	}

	stat := GetDiffStat(repo)
	if stat.Added != 1 {
		t.Errorf("staged file: got Added=%d, want 1", stat.Added)
	}
}

func TestGetDiffStat_NoDoubleCount(t *testing.T) {
	repo := initTestRepo(t)

	// Stage a change
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "add", "README.md")
	cmd.Dir = repo
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add: %s: %v", out, err)
	}

	stat := GetDiffStat(repo)

	// "# test\n" -> "changed\n" = 1 added, 1 removed (staged only, not double-counted)
	if stat.Added != 1 || stat.Removed != 1 {
		t.Errorf("staged-only change: got +%d -%d, want +1 -1", stat.Added, stat.Removed)
	}
}
