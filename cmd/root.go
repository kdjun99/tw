package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tw",
	Short: "tmux workspace manager - manage git worktrees with tmux",
	Long: `tw manages multiple git repositories and worktrees using tmux sessions
and windows, inspired by Superset's workspace UI.

Workflow:
  1. Register projects:  tw project add myapp ~/dev/myapp
  2. Create workspace:   tw add myapp feature/login
  3. View all:           tw list
  4. Switch workspace:   tw switch
  5. Remove workspace:   tw rm myapp feature/login`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
