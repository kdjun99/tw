package cmd

import (
	"fmt"
	"strings"

	"github.com/dongjunkim/tw/internal/config"
	"github.com/dongjunkim/tw/internal/git"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"st"},
	Short:   "Show diff stats for all workspaces",
	Example: `  tw status    # show +/- line stats per workspace
  tw st        # alias`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if len(cfg.Projects) == 0 {
			fmt.Println("No projects registered.")
			return nil
		}

		for i, proj := range cfg.Projects {
			if i > 0 {
				fmt.Println()
			}

			fmt.Printf("\033[1;36m%s\033[0m\n", proj.Name)

			worktrees, err := git.ListWorktrees(proj.Path)
			if err != nil {
				fmt.Printf("  Error: %v\n", err)
				continue
			}

			for _, wt := range worktrees {
				if wt.Bare {
					continue
				}

				stat := git.GetDiffStat(wt.Path)
				branchName := wt.Branch
				if branchName == "" {
					branchName = "(detached)"
				}

				isMain := wt.Path == proj.Path

				diffParts := []string{}
				if stat.Added > 0 {
					diffParts = append(diffParts, fmt.Sprintf("\033[32m+%d\033[0m", stat.Added))
				}
				if stat.Removed > 0 {
					diffParts = append(diffParts, fmt.Sprintf("\033[31m-%d\033[0m", stat.Removed))
				}
				diffStr := strings.Join(diffParts, " ")
				if diffStr == "" {
					diffStr = "\033[90mclean\033[0m"
				}

				marker := "├──"
				if isMain {
					fmt.Printf("  %s \033[33m%s\033[0m (local)  %s\n", marker, branchName, diffStr)
				} else {
					fmt.Printf("  %s \033[32m%s\033[0m  %s\n", marker, branchName, diffStr)
				}
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
