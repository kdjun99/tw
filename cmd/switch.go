package cmd

import (
	"fmt"
	"strings"

	"github.com/dongjunkim/tw/internal/config"
	"github.com/dongjunkim/tw/internal/git"
	"github.com/dongjunkim/tw/internal/ui"
	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:     "switch [project]",
	Aliases: []string{"sw", "s"},
	Short:   "Switch to a project or workspace",
	Example: `  tw switch              # fzf project picker → attach to project session
  tw switch myapp        # directly switch to myapp session
  tw switch -w           # fzf workspace picker (branch-level)
  tw sw                  # alias
  tw s                   # short alias`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		workspaceMode, _ := cmd.Flags().GetBool("workspace")

		if workspaceMode {
			return switchWorkspace()
		}

		// Project mode (default)
		if len(args) == 1 {
			return switchToProject(args[0])
		}
		return switchProjectPicker()
	},
}

func switchProjectPicker() error {
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
		branch := git.CurrentBranch(proj.Path)
		worktrees, _ := git.ListWorktrees(proj.Path)
		wtCount := 0
		for _, wt := range worktrees {
			if !wt.Bare && wt.Path != proj.Path {
				wtCount++
			}
		}

		label := fmt.Sprintf("\033[1;36m%s\033[0m (%s)", proj.Name, branch)
		if wtCount > 0 {
			label += fmt.Sprintf(" \033[90m+%d worktrees\033[0m", wtCount)
		}

		items = append(items, ui.Item{Display: label, Value: proj.Name})
	}

	selected, err := ui.FzfSelect(items, "Switch project")
	if err != nil {
		return err
	}

	return switchToProject(selected)
}

func switchToProject(projectName string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	proj := cfg.FindProject(projectName)
	if proj == nil {
		return fmt.Errorf("project %q not found", projectName)
	}

	b := getBackend()

	if !b.SessionExists(projectName) {
		if err := b.CreateSession(projectName, proj.Path); err != nil {
			return fmt.Errorf("create session: %w", err)
		}
		fmt.Printf("Created %s session %q\n", b.Name(), projectName)
	}

	return b.SwitchTo(projectName, "")
}

func switchWorkspace() error {
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

			value := fmt.Sprintf("%s:%s:%s", proj.Name, branchName, wt.Path)
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

	b := getBackend()

	sessionName := projectName
	if !b.SessionExists(sessionName) {
		if err := b.CreateSession(sessionName, wtPath); err != nil {
			return fmt.Errorf("create session: %w", err)
		}
	}

	if err := b.SwitchTo(sessionName, windowName); err != nil {
		if createErr := b.NewWindow(sessionName, windowName, wtPath); createErr != nil {
			return fmt.Errorf("create window: %w", createErr)
		}
		return b.SwitchTo(sessionName, windowName)
	}

	return nil
}

func init() {
	switchCmd.Flags().BoolP("workspace", "w", false, "switch by workspace (branch) instead of project")
	rootCmd.AddCommand(switchCmd)
}
