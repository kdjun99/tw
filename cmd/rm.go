package cmd

import (
	"fmt"

	"github.com/dongjunkim/tw/internal/config"
	"github.com/dongjunkim/tw/internal/git"
	"github.com/dongjunkim/tw/internal/setup"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:     "rm <project> <branch>",
	Aliases: []string{"remove"},
	Short:   "Remove a workspace (git worktree + terminal window)",
	Example: `  # Remove worktree + terminal window
  tw rm myapp feature/login

  # Keep worktree, only close terminal window
  tw rm myapp feature/login --keep-worktree

  # Keep terminal window, only remove worktree
  tw rm myapp feature/login --keep-terminal`,
	Args: cobra.ExactArgs(2),
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
		keepTerminal, _ := cmd.Flags().GetBool("keep-terminal")
		if keepTmux, _ := cmd.Flags().GetBool("keep-tmux"); keepTmux {
			keepTerminal = true
		}

		// Remove terminal window
		if !keepTerminal {
			b := getBackend()
			sessionName := projectName
			windowName := shortBranch(branch)
			if b.SessionExists(sessionName) {
				if err := b.KillWindow(sessionName, windowName); err != nil {
					fmt.Printf("Warning: could not close window: %v\n", err)
				} else {
					fmt.Printf("Closed window %q (%s)\n", windowName, b.Name())
				}
			}
		}

		// Run teardown and remove worktree
		if !keepWorktree {
			wtDir := proj.ResolveWorktreeDir()
			wtPath := git.ResolveWorktreePath(wtDir, branch)

			setup.RunTeardown(proj, wtPath)

			if err := git.RemoveWorktree(proj.Path, wtPath); err != nil {
				return fmt.Errorf("remove worktree: %w", err)
			}
			fmt.Printf("Removed worktree at %s\n", wtPath)
		}

		return nil
	},
}

func init() {
	rmCmd.Flags().Bool("keep-worktree", false, "keep the git worktree, only close terminal window")
	rmCmd.Flags().Bool("keep-terminal", false, "keep terminal window, only remove worktree")
	rmCmd.Flags().Bool("keep-tmux", false, "alias for --keep-terminal (deprecated)")

	rootCmd.AddCommand(rmCmd)
}
