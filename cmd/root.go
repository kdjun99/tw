package cmd

import (
	"fmt"
	"os"

	"github.com/dongjunkim/tw/internal/config"
	"github.com/dongjunkim/tw/internal/terminal"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "tw",
		Short: "terminal workspace manager - manage git worktrees with tmux/cmux",
		Long: `tw manages multiple git repositories and worktrees using tmux or cmux,
inspired by Superset's workspace UI.

Workflow:
  1. Register projects:  tw project add myapp ~/dev/myapp
  2. Create workspace:   tw add myapp feature/login
  3. View all:           tw list
  4. Switch workspace:   tw switch
  5. Remove workspace:   tw rm myapp feature/login`,
	}

	// backend is the active terminal backend, initialized in Execute().
	backend terminal.Backend
)

func getBackend() terminal.Backend {
	return backend
}

func Execute() {
	cfg, _ := config.Load()
	backendName := ""
	if cfg != nil {
		backendName = cfg.Backend
	}
	backend = terminal.NewBackend(backendName)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
