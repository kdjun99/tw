package cmd

import (
	"fmt"

	"github.com/dongjunkim/tw/internal/config"
	"github.com/dongjunkim/tw/internal/git"
	"github.com/dongjunkim/tw/internal/setup"
	"github.com/dongjunkim/tw/internal/tmux"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <project> <branch>",
	Short: "Create a new workspace (git worktree + tmux window)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]
		branch := args[1]

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		proj := cfg.FindProject(projectName)
		if proj == nil {
			return fmt.Errorf("project %q not found. Use 'tw project add' first", projectName)
		}

		baseBranch, _ := cmd.Flags().GetString("base")
		if baseBranch == "" {
			baseBranch = proj.DefaultBranch
		}

		useExisting, _ := cmd.Flags().GetBool("existing")
		noTmux, _ := cmd.Flags().GetBool("no-tmux")

		// Resolve worktree path
		wtDir := proj.ResolveWorktreeDir()
		wtPath := git.ResolveWorktreePath(wtDir, branch)

		// Create worktree
		if useExisting {
			if !git.BranchExists(proj.Path, branch) {
				return fmt.Errorf("branch %q does not exist", branch)
			}
			fmt.Printf("Checking out existing branch %q...\n", branch)
			if err := git.AddWorktreeExisting(proj.Path, wtPath, branch); err != nil {
				return fmt.Errorf("create worktree: %w", err)
			}
		} else {
			fmt.Printf("Creating worktree %q from %q...\n", branch, baseBranch)
			if err := git.AddWorktree(proj.Path, wtPath, branch, baseBranch); err != nil {
				return fmt.Errorf("create worktree: %w", err)
			}
		}

		fmt.Printf("Worktree created at %s\n", wtPath)

		// Run setup (copy files + run commands)
		noSetup, _ := cmd.Flags().GetBool("no-setup")
		if !noSetup {
			setupCfg, err := setup.LoadConfig(proj.Path)
			if err != nil {
				fmt.Printf("Warning: failed to load .tw.toml: %v\n", err)
			} else if setupCfg != nil {
				fmt.Println("Running workspace setup...")
				if err := setup.RunSetup(proj.Path, wtPath, setupCfg); err != nil {
					fmt.Printf("Warning: setup failed: %v\n", err)
				}
			}
		}

		if noTmux {
			return nil
		}

		// Create tmux session if not exists, then add window
		sessionName := projectName
		if !tmux.SessionExists(sessionName) {
			if err := tmux.CreateSession(sessionName, proj.Path); err != nil {
				return fmt.Errorf("create tmux session: %w", err)
			}
			fmt.Printf("Created tmux session %q\n", sessionName)
		}

		windowName := shortBranch(branch)
		if err := tmux.NewWindow(sessionName, windowName, wtPath); err != nil {
			return fmt.Errorf("create tmux window: %w", err)
		}

		fmt.Printf("Created tmux window %q in session %q\n", windowName, sessionName)

		// Auto-switch if inside tmux
		switchFlag, _ := cmd.Flags().GetBool("switch")
		if switchFlag && tmux.IsInsideTmux() {
			return tmux.SwitchTo(sessionName, windowName)
		}

		return nil
	},
}

func shortBranch(branch string) string {
	// feature/add-career -> add-career
	parts := splitLast(branch, "/")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return branch
}

func splitLast(s, sep string) []string {
	for i := len(s) - 1; i >= 0; i-- {
		if string(s[i]) == sep {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

func init() {
	addCmd.Flags().String("base", "", "base branch (defaults to project's default branch)")
	addCmd.Flags().Bool("existing", false, "checkout existing branch instead of creating new")
	addCmd.Flags().Bool("no-tmux", false, "create worktree only, skip tmux")
	addCmd.Flags().Bool("no-setup", false, "skip .tw.toml setup steps")
	addCmd.Flags().BoolP("switch", "s", true, "auto-switch to new window")

	rootCmd.AddCommand(addCmd)
}
