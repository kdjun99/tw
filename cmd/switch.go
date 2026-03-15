package cmd

import (
	"fmt"
	"strings"

	"github.com/dongjunkim/tw/internal/config"
	"github.com/dongjunkim/tw/internal/git"
	"github.com/dongjunkim/tw/internal/tmux"
	"github.com/dongjunkim/tw/internal/ui"
	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:     "switch",
	Aliases: []string{"sw", "s"},
	Short:   "Interactively switch workspace (fzf)",
	Example: `  tw switch    # opens fzf selector
  tw sw        # alias
  tw s         # short alias`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !ui.HasFzf() {
			return fmt.Errorf("fzf is required for interactive switching. Install: brew install fzf")
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if len(cfg.Projects) == 0 {
			return fmt.Errorf("no projects registered")
		}

		var items []ui.Item
		for _, proj := range cfg.Projects {
			worktrees, err := git.ListWorktrees(proj.Path)
			if err != nil {
				continue
			}

			for _, wt := range worktrees {
				if wt.Bare {
					continue
				}

				stat := git.GetDiffStat(wt.Path)
				diffStr := ""
				if stat.Added > 0 || stat.Removed > 0 {
					parts := []string{}
					if stat.Added > 0 {
						parts = append(parts, fmt.Sprintf("+%d", stat.Added))
					}
					if stat.Removed > 0 {
						parts = append(parts, fmt.Sprintf("-%d", stat.Removed))
					}
					diffStr = " " + strings.Join(parts, " ")
				}

				branchName := wt.Branch
				if branchName == "" {
					branchName = "(detached)"
				}

				isMain := wt.Path == proj.Path
				label := ""
				if isMain {
					label = fmt.Sprintf("\033[1;36m%s\033[0m / \033[33m%s\033[0m (local)%s", proj.Name, branchName, diffStr)
				} else {
					label = fmt.Sprintf("\033[1;36m%s\033[0m / \033[32m%s\033[0m%s", proj.Name, branchName, diffStr)
				}

				value := fmt.Sprintf("%s:%s:%s", proj.Name, wt.Branch, wt.Path)
				items = append(items, ui.Item{Display: label, Value: value})
			}
		}

		if len(items) == 0 {
			return fmt.Errorf("no workspaces found")
		}

		selected, err := ui.FzfSelect(items, "Switch to")
		if err != nil {
			return err
		}

		// Parse selection: project:branch:path
		parts := strings.SplitN(selected, ":", 3)
		if len(parts) < 3 {
			return fmt.Errorf("invalid selection")
		}
		projectName := parts[0]
		windowName := shortBranch(parts[1])
		wtPath := parts[2]

		// Ensure tmux session exists
		sessionName := projectName
		if !tmux.SessionExists(sessionName) {
			if err := tmux.CreateSession(sessionName, wtPath); err != nil {
				return fmt.Errorf("create session: %w", err)
			}
		}

		// Try to switch to existing window, or create one
		if err := tmux.SwitchTo(sessionName, windowName); err != nil {
			// Window might not exist, create it
			if createErr := tmux.NewWindow(sessionName, windowName, wtPath); createErr != nil {
				return fmt.Errorf("create window: %w", createErr)
			}
			return tmux.SwitchTo(sessionName, windowName)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(switchCmd)
}
