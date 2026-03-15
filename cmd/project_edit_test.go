package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/dongjunkim/tw/internal/config"
)

func TestStripComments(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "removes comment lines",
			input: "// comment\n{\"name\": \"test\"}\n",
			want:  "{\"name\": \"test\"}\n",
		},
		{
			name:  "preserves indented json",
			input: "// header\n{\n  \"name\": \"test\"\n}\n",
			want:  "{\n  \"name\": \"test\"\n}\n",
		},
		{
			name:  "handles comment with leading spaces",
			input: "  // indented comment\n{}\n",
			want:  "{}\n",
		},
		{
			name:  "no comments",
			input: "{\"name\": \"test\"}\n",
			want:  "{\"name\": \"test\"}\n",
		},
		{
			name:  "does not strip // inside json values",
			input: "{\n  \"url\": \"https://example.com\"\n}\n",
			want:  "{\n  \"url\": \"https://example.com\"\n}\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripComments(tt.input)
			if got != tt.want {
				t.Errorf("stripComments() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateAnnotatedConfig(t *testing.T) {
	p := config.Project{
		Name:          "myapp",
		Path:          "/dev/myapp",
		DefaultBranch: "main",
	}

	result := generateAnnotatedConfig(p)

	// Should contain comment header
	if !strings.Contains(result, "// tw project config") {
		t.Error("missing comment header")
	}

	// Should contain field descriptions
	for _, field := range []string{"name:", "path:", "defaultBranch:", "worktreeDir:", "setup.copy:", "setup.run:", "teardown.run:"} {
		if !strings.Contains(result, field) {
			t.Errorf("missing field description for %s", field)
		}
	}

	// Should be parseable after stripping comments
	cleaned := stripComments(result)
	var parsed config.Project
	if err := json.Unmarshal([]byte(cleaned), &parsed); err != nil {
		t.Fatalf("generated config not parseable: %v", err)
	}
	if parsed.Name != "myapp" {
		t.Errorf("name = %q, want %q", parsed.Name, "myapp")
	}
	if parsed.Path != "/dev/myapp" {
		t.Errorf("path = %q, want %q", parsed.Path, "/dev/myapp")
	}
}

func TestGenerateAnnotatedConfig_WithSetup(t *testing.T) {
	p := config.Project{
		Name:          "api",
		Path:          "/dev/api",
		DefaultBranch: "develop",
		Setup: &config.SetupConfig{
			Copy: []string{".env*"},
			Run:  []string{"npm install"},
		},
		Teardown: &config.TeardownConfig{
			Run: []string{"docker compose down"},
		},
	}

	result := generateAnnotatedConfig(p)
	cleaned := stripComments(result)

	var parsed config.Project
	if err := json.Unmarshal([]byte(cleaned), &parsed); err != nil {
		t.Fatalf("generated config not parseable: %v", err)
	}
	if parsed.Setup == nil || len(parsed.Setup.Copy) != 1 || parsed.Setup.Copy[0] != ".env*" {
		t.Errorf("setup.copy not preserved: %+v", parsed.Setup)
	}
	if parsed.Teardown == nil || len(parsed.Teardown.Run) != 1 {
		t.Errorf("teardown not preserved: %+v", parsed.Teardown)
	}
}
