package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tw",
	Short: "tmux workspace manager - manage git worktrees with tmux",
	Long:  `tw manages multiple git repositories and worktrees using tmux sessions and windows, inspired by Superset's workspace UI.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
