package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

type Worktree struct {
	Path   string
	Branch string
	Bare   bool
}

type DiffStat struct {
	Added   int
	Removed int
}

func ListWorktrees(repoPath string) ([]Worktree, error) {
	out, err := runGit(repoPath, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	var worktrees []Worktree
	var current Worktree
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			current = Worktree{Path: strings.TrimPrefix(line, "worktree ")}
		case strings.HasPrefix(line, "branch "):
			ref := strings.TrimPrefix(line, "branch ")
			current.Branch = strings.TrimPrefix(ref, "refs/heads/")
		case line == "bare":
			current.Bare = true
		case line == "":
			if current.Path != "" {
				worktrees = append(worktrees, current)
			}
			current = Worktree{}
		}
	}
	if current.Path != "" {
		worktrees = append(worktrees, current)
	}
	return worktrees, nil
}

func AddWorktree(repoPath, worktreePath, branch, baseBranch string) error {
	args := []string{"worktree", "add"}
	if baseBranch != "" {
		args = append(args, "-b", branch, worktreePath, baseBranch)
	} else {
		args = append(args, "-b", branch, worktreePath)
	}
	_, err := runGit(repoPath, args...)
	return err
}

func AddWorktreeExisting(repoPath, worktreePath, branch string) error {
	_, err := runGit(repoPath, "worktree", "add", worktreePath, branch)
	return err
}

func RemoveWorktree(repoPath, worktreePath string) error {
	_, err := runGit(repoPath, "worktree", "remove", "--force", worktreePath)
	return err
}

func CurrentBranch(path string) string {
	out, err := runGit(path, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(out)
}

func GetDiffStat(path string) DiffStat {
	out, err := runGit(path, "diff", "--numstat", "HEAD")
	if err != nil {
		return DiffStat{}
	}
	var stat DiffStat
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		var added, removed int
		fmt.Sscanf(line, "%d\t%d", &added, &removed)
		stat.Added += added
		stat.Removed += removed
	}
	// Also count untracked/staged
	outStaged, err := runGit(path, "diff", "--numstat", "--cached")
	if err == nil {
		for _, line := range strings.Split(strings.TrimSpace(outStaged), "\n") {
			if line == "" {
				continue
			}
			var added, removed int
			fmt.Sscanf(line, "%d\t%d", &added, &removed)
			stat.Added += added
			stat.Removed += removed
		}
	}
	return stat
}

func BranchExists(repoPath, branch string) bool {
	_, err := runGit(repoPath, "rev-parse", "--verify", "refs/heads/"+branch)
	return err == nil
}

func ListBranches(repoPath string) ([]string, error) {
	out, err := runGit(repoPath, "branch", "--format=%(refname:short)")
	if err != nil {
		return nil, err
	}
	var branches []string
	for _, b := range strings.Split(strings.TrimSpace(out), "\n") {
		if b != "" {
			branches = append(branches, b)
		}
	}
	return branches, nil
}

func ResolveWorktreePath(baseDir, branch string) string {
	safe := strings.ReplaceAll(branch, "/", "-")
	return filepath.Join(baseDir, safe)
}

func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %s: %w", strings.Join(args, " "), strings.TrimSpace(string(out)), err)
	}
	return string(out), nil
}
