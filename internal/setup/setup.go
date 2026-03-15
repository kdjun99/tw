package setup

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Setup    SetupConfig    `toml:"setup"`
	Teardown TeardownConfig `toml:"teardown"`
}

type SetupConfig struct {
	Copy []string `toml:"copy"`
	Run  []string `toml:"run"`
}

type TeardownConfig struct {
	Run []string `toml:"run"`
}

const ConfigFileName = ".tw.toml"

func LoadConfig(repoPath string) (*Config, error) {
	path := filepath.Join(repoPath, ConfigFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", ConfigFileName, err)
	}
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", ConfigFileName, err)
	}
	return &cfg, nil
}

func RunSetup(repoPath, worktreePath string, cfg *Config) error {
	if cfg == nil {
		return nil
	}

	// Copy files
	for _, pattern := range cfg.Setup.Copy {
		matches, err := filepath.Glob(filepath.Join(repoPath, pattern))
		if err != nil {
			return fmt.Errorf("glob %q: %w", pattern, err)
		}
		for _, src := range matches {
			rel, _ := filepath.Rel(repoPath, src)
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
	for _, cmdStr := range cfg.Setup.Run {
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

func RunTeardown(repoPath, worktreePath string) error {
	cfg, err := LoadConfig(repoPath)
	if err != nil || cfg == nil {
		return err
	}

	for _, cmdStr := range cfg.Teardown.Run {
		fmt.Printf("  Running teardown: %s\n", cmdStr)
		cmd := exec.Command("sh", "-c", cmdStr)
		cmd.Dir = worktreePath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("  Warning: teardown %q failed: %v\n", cmdStr, err)
		}
	}

	return nil
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
