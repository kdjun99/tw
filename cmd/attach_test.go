package cmd

import "testing"

func TestParseAttachArg(t *testing.T) {
	tests := []struct {
		arg         string
		wantProject string
		wantWindow  string
	}{
		{"myapp", "myapp", ""},
		{"myapp/feature-login", "myapp", "feature-login"},
		{"myapp/nested/window", "myapp", "nested/window"},
		{"", "", ""},
	}
	for _, tt := range tests {
		project, window := parseAttachArg(tt.arg)
		if project != tt.wantProject || window != tt.wantWindow {
			t.Errorf("parseAttachArg(%q) = (%q, %q), want (%q, %q)",
				tt.arg, project, window, tt.wantProject, tt.wantWindow)
		}
	}
}
