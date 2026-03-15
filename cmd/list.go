package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/dongjunkim/tw/internal/config"
	"github.com/dongjunkim/tw/internal/git"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all projects and their workspaces",
	Example: `  tw list       # show all projects and worktrees
  tw ls         # alias
  tw list -a    # show all details`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if len(cfg.Projects) == 0 {
			fmt.Println("No projects registered. Use 'tw project add <name> <path>' to add one.")
			return nil
		}

		for i, proj := range cfg.Projects {
			if i > 0 {
				fmt.Println()
			}

			branch := git.CurrentBranch(proj.Path)
			fmt.Printf("\033[1;36m%s\033[0m (%s)\n", proj.Name, branch)

			worktrees, err := git.ListWorktrees(proj.Path)
			if err != nil {
				fmt.Printf("  Error listing worktrees: %v\n", err)
				continue
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

			for _, wt := range worktrees {
				if wt.Bare {
					continue
				}

				stat := git.GetDiffStat(wt.Path)
				diffStr := ""
				if stat.Added > 0 || stat.Removed > 0 {
					parts := []string{}
					if stat.Added > 0 {
						parts = append(parts, fmt.Sprintf("\033[32m+%d\033[0m", stat.Added))
					}
					if stat.Removed > 0 {
						parts = append(parts, fmt.Sprintf("\033[31m-%d\033[0m", stat.Removed))
					}
					diffStr = strings.Join(parts, " ")
				}

				branchDisplay := wt.Branch
				if branchDisplay == "" {
					branchDisplay = "(detached)"
				}

				fmt.Fprintf(w, "  \u251c\u2500\u2500 %s\t%s\t%s\n", branchDisplay, diffStr, shortenPath(wt.Path))
			}
			w.Flush()
		}

		return nil
	},
}

func shortenPath(path string) string {
	home, _ := os.UserHomeDir()
	if strings.HasPrefix(path, home) {
		return "~" + path[len(home):]
	}
	return path
}

func init() {
	listCmd.Flags().BoolP("all", "a", false, "show all details")
	rootCmd.AddCommand(listCmd)
}
