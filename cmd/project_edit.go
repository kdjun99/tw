package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/dongjunkim/tw/internal/config"
	"github.com/spf13/cobra"
)

var projectEditCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Edit a project's configuration in $EDITOR",
	Example: `  # Open myapp config in $EDITOR
  tw project edit myapp

  # Using project alias
  tw p edit myapp`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		originalName := args[0]

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		proj := cfg.FindProject(originalName)
		if proj == nil {
			return fmt.Errorf("project %q not found", originalName)
		}

		// Generate annotated config
		content := generateAnnotatedConfig(*proj)

		// Write to temp file
		tmpFile, err := os.CreateTemp("", "tw-project-*.jsonc")
		if err != nil {
			return fmt.Errorf("create temp file: %w", err)
		}
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.WriteString(content); err != nil {
			tmpFile.Close()
			return fmt.Errorf("write temp file: %w", err)
		}
		tmpFile.Close()

		// Open in editor
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}

		editorCmd := exec.Command(editor, tmpFile.Name())
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr
		if err := editorCmd.Run(); err != nil {
			return fmt.Errorf("editor failed: %w", err)
		}

		// Read back and parse
		data, err := os.ReadFile(tmpFile.Name())
		if err != nil {
			return fmt.Errorf("read temp file: %w", err)
		}

		cleaned := stripComments(string(data))

		var updated config.Project
		if err := json.Unmarshal([]byte(cleaned), &updated); err != nil {
			return fmt.Errorf("invalid config: %w", err)
		}

		// Validate
		if updated.Name == "" {
			return fmt.Errorf("name cannot be empty")
		}
		if updated.Path == "" {
			return fmt.Errorf("path cannot be empty")
		}
		if updated.DefaultBranch == "" {
			return fmt.Errorf("defaultBranch cannot be empty")
		}

		// Update config
		if err := cfg.UpdateProject(originalName, updated); err != nil {
			return err
		}

		fmt.Printf("Updated project %q\n", updated.Name)
		return nil
	},
}

func generateAnnotatedConfig(p config.Project) string {
	// Ensure optional fields are present for visibility
	editProject := struct {
		Name          string               `json:"name"`
		Path          string               `json:"path"`
		DefaultBranch string               `json:"defaultBranch"`
		WorktreeDir   string               `json:"worktreeDir"`
		Setup         *config.SetupConfig  `json:"setup"`
		Teardown      *config.TeardownConfig `json:"teardown"`
	}{
		Name:          p.Name,
		Path:          p.Path,
		DefaultBranch: p.DefaultBranch,
		WorktreeDir:   p.WorktreeDir,
		Setup:         p.Setup,
		Teardown:      p.Teardown,
	}

	if editProject.Setup == nil {
		editProject.Setup = &config.SetupConfig{Copy: []string{}, Run: []string{}}
	}
	if editProject.Teardown == nil {
		editProject.Teardown = &config.TeardownConfig{Run: []string{}}
	}

	data, _ := json.MarshalIndent(editProject, "", "  ")

	var b strings.Builder
	b.WriteString(fmt.Sprintf("// tw project config for %q\n", p.Name))
	b.WriteString("// Edit the values below. Lines starting with // are ignored.\n")
	b.WriteString("//\n")
	b.WriteString("// name:           Project identifier used in all tw commands\n")
	b.WriteString("// path:           Absolute path to the git repository root\n")
	b.WriteString("// defaultBranch:  Base branch for new worktrees (e.g. main, develop)\n")
	b.WriteString("// worktreeDir:    Custom worktree directory (empty = ~/.tw/<name>/)\n")
	b.WriteString("// setup.copy:     Glob patterns to copy from repo to new worktrees (e.g. \".env*\")\n")
	b.WriteString("// setup.run:      Shell commands to run after worktree creation (e.g. \"npm install\")\n")
	b.WriteString("// teardown.run:   Shell commands to run before worktree removal (e.g. \"docker compose down\")\n")
	b.WriteString(string(data))
	b.WriteString("\n")
	return b.String()
}

func stripComments(input string) string {
	var lines []string
	for _, line := range strings.Split(input, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			continue
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func init() {
	projectCmd.AddCommand(projectEditCmd)
}
