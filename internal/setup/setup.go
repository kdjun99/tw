package setup

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dongjunkim/tw/internal/config"
)

func RunSetup(proj *config.Project, worktreePath string) error {
	if proj.Setup == nil {
		return nil
	}

	// Copy files
	for _, pattern := range proj.Setup.Copy {
		matches, err := filepath.Glob(filepath.Join(proj.Path, pattern))
		if err != nil {
			return fmt.Errorf("glob %q: %w", pattern, err)
		}
		for _, src := range matches {
			rel, _ := filepath.Rel(proj.Path, src)
			dst := filepath.Join(worktreePath, rel)

			if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
				return fmt.Errorf("mkdir for %s: %w", rel, err)
			}
			if err := copyFile(src, dst); err != nil {
				return fmt.Errorf("copy %s: %w", rel, err)
			}
			fmt.Printf("  Copied %s\n", rel)
		}
	}

	// Run commands
	for _, cmdStr := range proj.Setup.Run {
		fmt.Printf("  Running: %s\n", cmdStr)
		cmd := exec.Command("sh", "-c", cmdStr)
		cmd.Dir = worktreePath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("command %q failed: %w", cmdStr, err)
		}
	}

	return nil
}

func RunTeardown(proj *config.Project, worktreePath string) {
	if proj.Teardown == nil {
		return
	}

	for _, cmdStr := range proj.Teardown.Run {
		fmt.Printf("  Running teardown: %s\n", cmdStr)
		cmd := exec.Command("sh", "-c", cmdStr)
		cmd.Dir = worktreePath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("  Warning: teardown %q failed: %v\n", cmdStr, err)
		}
	}
}

func copyFile(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
