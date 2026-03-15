package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/dongjunkim/tw/internal/config"
	"github.com/dongjunkim/tw/internal/git"
	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:     "project",
	Aliases: []string{"p"},
	Short:   "Manage registered projects",
}

var projectAddCmd = &cobra.Command{
	Use:   "add <name> <path>",
	Short: "Register a git repository as a project",
	Example: `  # Basic registration (auto-detects default branch)
  tw project add myapp ~/dev/myapp

  # Specify default branch
  tw project add api ~/dev/api-server --default-branch develop

  # With setup automation
  tw project add myapp ~/dev/myapp \
    --copy '.env*' \
    --setup-run 'bun install' \
    --teardown-run 'docker compose down'

  # Multiple setup commands
  tw project add myapp ~/dev/myapp \
    --copy '.env*' \
    --setup-run 'npm install' \
    --setup-run 'npm run build'`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		path, err := filepath.Abs(args[1])
		if err != nil {
			return err
		}

		// Verify it's a git repo
		if _, err := os.Stat(filepath.Join(path, ".git")); err != nil {
			return fmt.Errorf("%s is not a git repository", path)
		}

		defaultBranch, _ := cmd.Flags().GetString("default-branch")
		if defaultBranch == "" {
			defaultBranch = git.CurrentBranch(path)
		}

		worktreeDir, _ := cmd.Flags().GetString("worktree-dir")
		copyFiles, _ := cmd.Flags().GetStringSlice("copy")
		setupRun, _ := cmd.Flags().GetStringSlice("setup-run")
		teardownRun, _ := cmd.Flags().GetStringSlice("teardown-run")

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		p := config.Project{
			Name:          name,
			Path:          path,
			DefaultBranch: defaultBranch,
			WorktreeDir:   worktreeDir,
		}

		if len(copyFiles) > 0 || len(setupRun) > 0 {
			p.Setup = &config.SetupConfig{
				Copy: copyFiles,
				Run:  setupRun,
			}
		}
		if len(teardownRun) > 0 {
			p.Teardown = &config.TeardownConfig{
				Run: teardownRun,
			}
		}

		if err := cfg.AddProject(p); err != nil {
			return err
		}

		fmt.Printf("Added project %q (%s, default branch: %s)\n", name, path, defaultBranch)
		return nil
	},
}

var projectRmCmd = &cobra.Command{
	Use:     "rm <name>",
	Aliases: []string{"remove"},
	Short:   "Unregister a project",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if err := cfg.RemoveProject(args[0]); err != nil {
			return err
		}
		fmt.Printf("Removed project %q\n", args[0])
		return nil
	},
}

var projectListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List registered projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if len(cfg.Projects) == 0 {
			fmt.Println("No projects registered. Use 'tw project add <name> <path>' to add one.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tPATH\tDEFAULT BRANCH")
		for _, p := range cfg.Projects {
			fmt.Fprintf(w, "%s\t%s\t%s\n", p.Name, p.Path, p.DefaultBranch)
		}
		return w.Flush()
	},
}

func init() {
	projectAddCmd.Flags().String("default-branch", "", "default branch (auto-detected if empty)")
	projectAddCmd.Flags().String("worktree-dir", "", "custom worktree directory")
	projectAddCmd.Flags().StringSlice("copy", nil, "files to copy from main repo to worktree (e.g. --copy .env,.env.local)")
	projectAddCmd.Flags().StringSlice("setup-run", nil, "commands to run after worktree creation (e.g. --setup-run 'bun install')")
	projectAddCmd.Flags().StringSlice("teardown-run", nil, "commands to run before worktree removal")

	projectCmd.AddCommand(projectAddCmd, projectRmCmd, projectListCmd)
	rootCmd.AddCommand(projectCmd)
}
