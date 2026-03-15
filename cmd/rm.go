package cmd

import (
	"fmt"

	"github.com/dongjunkim/tw/internal/config"
	"github.com/dongjunkim/tw/internal/git"
	"github.com/dongjunkim/tw/internal/setup"
	"github.com/dongjunkim/tw/internal/tmux"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:     "rm <project> <branch>",
	Aliases: []string{"remove"},
	Short:   "Remove a workspace (git worktree + tmux window)",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]
		branch := args[1]

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		proj := cfg.FindProject(projectName)
		if proj == nil {
			return fmt.Errorf("project %q not found", projectName)
		}

		keepWorktree, _ := cmd.Flags().GetBool("keep-worktree")
		keepTmux, _ := cmd.Flags().GetBool("keep-tmux")

		// Remove tmux window
		if !keepTmux {
			sessionName := projectName
			windowName := shortBranch(branch)
			if tmux.SessionExists(sessionName) {
				if err := tmux.KillWindow(sessionName, windowName); err != nil {
					fmt.Printf("Warning: could not close tmux window: %v\n", err)
				} else {
					fmt.Printf("Closed tmux window %q\n", windowName)
				}
			}
		}

		// Run teardown before removing worktree
		if !keepWorktree {
			wtDir := proj.ResolveWorktreeDir()
			wtPath := git.ResolveWorktreePath(wtDir, branch)
			setup.RunTeardown(proj.Path, wtPath)
		}

		// Remove worktree
		if !keepWorktree {
			wtDir := proj.ResolveWorktreeDir()
			wtPath := git.ResolveWorktreePath(wtDir, branch)
			if err := git.RemoveWorktree(proj.Path, wtPath); err != nil {
				return fmt.Errorf("remove worktree: %w", err)
			}
			fmt.Printf("Removed worktree at %s\n", wtPath)
		}

		return nil
	},
}

func init() {
	rmCmd.Flags().Bool("keep-worktree", false, "keep the git worktree, only close tmux window")
	rmCmd.Flags().Bool("keep-tmux", false, "keep tmux window, only remove worktree")

	rootCmd.AddCommand(rmCmd)
}
