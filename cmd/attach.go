package cmd

import (
	"fmt"
	"strings"

	"github.com/dongjunkim/tw/internal/config"
	"github.com/dongjunkim/tw/internal/tmux"
	"github.com/dongjunkim/tw/internal/ui"
	"github.com/spf13/cobra"
)

var attachCmd = &cobra.Command{
	Use:     "attach [project[/window]]",
	Aliases: []string{"a"},
	Short:   "Attach to a project's tmux session",
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

		// Ensure tmux session exists
		if !tmux.SessionExists(projectName) {
			if err := tmux.CreateSession(projectName, proj.Path); err != nil {
				return fmt.Errorf("create session: %w", err)
			}
			fmt.Printf("Created tmux session %q\n", projectName)
		}

		// Attach/switch to session (optionally with window)
		if err := tmux.SwitchTo(projectName, windowName); err != nil {
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
