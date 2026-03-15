package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Item struct {
	Display string
	Value   string
}

func FzfSelect(items []Item, prompt string) (string, error) {
	if len(items) == 0 {
		return "", fmt.Errorf("no items to select")
	}

	var input strings.Builder
	for _, item := range items {
		fmt.Fprintf(&input, "%s\t%s\n", item.Value, item.Display)
	}

	args := []string{
		"--ansi",
		"--prompt", prompt + " ",
		"--delimiter", "\t",
		"--with-nth", "2",
		"--no-multi",
		"--height", "~40%",
		"--layout", "reverse",
		"--border", "rounded",
		"--border-label", " tw - workspace switcher ",
	}

	cmd := exec.Command("fzf", args...)
	cmd.Stdin = strings.NewReader(input.String())
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 130 {
			return "", fmt.Errorf("cancelled")
		}
		return "", err
	}

	line := strings.TrimSpace(string(out))
	parts := strings.SplitN(line, "\t", 2)
	return parts[0], nil
}

func HasFzf() bool {
	_, err := exec.LookPath("fzf")
	return err == nil
}
