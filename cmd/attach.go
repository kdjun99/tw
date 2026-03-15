package cmd

import (
	"fmt"
	"strings"

	"github.com/dongjunkim/tw/internal/config"
	"github.com/dongjunkim/tw/internal/git"
	"github.com/dongjunkim/tw/internal/ui"
	"github.com/spf13/cobra"
)

var attachCmd = &cobra.Command{
	Use:     "attach [project[/window]]",
	Aliases: []string{"a"},
	Short:   "Attach to a project's terminal session",
	Example: `  # Attach to myapp session (creates if needed)
  tw attach myapp

  # Attach to myapp session, switch to specific window
  tw attach myapp/feature-login

  # Interactive project picker
  tw attach

  # Short alias
  tw a`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if len(cfg.Projects) == 0 {
			return fmt.Errorf("no projects registered. Use 'tw project add' first")
		}

		var projectName, windowName string

		if len(args) == 0 {
			// Interactive fzf picker
			if !ui.HasFzf() {
				return fmt.Errorf("fzf is required for interactive selection. Install: brew install fzf")
			}

			var items []ui.Item
			for _, proj := range cfg.Projects {
				label := fmt.Sprintf("\033[1;36m%s\033[0m (%s)", proj.Name, shortenPath(proj.Path))
				items = append(items, ui.Item{Display: label, Value: proj.Name})
			}

			selected, err := ui.FzfSelect(items, "Attach to")
			if err != nil {
				return err
			}
			projectName = selected
		} else {
			projectName, windowName = parseAttachArg(args[0])
		}

		proj := cfg.FindProject(projectName)
		if proj == nil {
			return fmt.Errorf("project %q not found", projectName)
		}

		b := getBackend()

		// Ensure session exists
		sessionCreated := false
		if !b.SessionExists(projectName) {
			if err := b.CreateSession(projectName, proj.Path); err != nil {
				return fmt.Errorf("create session: %w", err)
			}
			fmt.Printf("Created %s session %q\n", b.Name(), projectName)
			sessionCreated = true
		}

		// Sync worktree windows: create missing windows for existing worktrees
		worktrees, err := git.ListWorktrees(proj.Path)
		if err == nil {
			// Get existing windows to avoid duplicates
			existingWindows := map[string]bool{}
			if !sessionCreated {
				if windows, err := b.ListWindows(projectName); err == nil {
					for _, w := range windows {
						existingWindows[w.Name] = true
					}
				}
			}

			firstWindow := true
			for _, wt := range worktrees {
				if wt.Bare {
					continue
				}
				branchName := wt.Branch
				if branchName == "" {
					continue
				}
				winName := shortBranch(branchName)

				// For main worktree on fresh session, rename the default surface
				if wt.Path == proj.Path && sessionCreated && firstWindow {
					b.RenameWindow(projectName, "", winName)
					firstWindow = false
					continue
				}

				if existingWindows[winName] {
					continue
				}
				if err := b.NewWindow(projectName, winName, wt.Path); err != nil {
					fmt.Printf("Warning: could not create window for %s: %v\n", branchName, err)
				} else {
					fmt.Printf("  Window %q (%s)\n", winName, branchName)
				}
			}
		}

		// Attach/switch to session (optionally with window)
		if err := b.SwitchTo(projectName, windowName); err != nil {
			if windowName != "" {
				return fmt.Errorf("window %q not found in session %q", windowName, projectName)
			}
			return fmt.Errorf("attach to session: %w", err)
		}

		return nil
	},
}

func parseAttachArg(arg string) (project, window string) {
	parts := strings.SplitN(arg, "/", 2)
	project = parts[0]
	if len(parts) == 2 {
		window = parts[1]
	}
	return
}

func init() {
	rootCmd.AddCommand(attachCmd)
}
