package llm_test

import (
	"strings"
	"testing"

	"github.com/karan/oops/internal/context"
	"github.com/karan/oops/internal/llm"
)

func TestParseCommands(t *testing.T) {
	cases := []struct {
		input string
		want  []string
	}{
		{"git push origin main", []string{"git push origin main"}},
		{
			"git add .\ngit commit -m 'fix'\ngit push",
			[]string{"git add .", "git commit -m 'fix'", "git push"},
		},
		{"```\ngit push\n```", []string{"git push"}},
		{"```sh\ngit push\n```", []string{"git push"}},
		{"$ git push", []string{"git push"}},
		{"  git push  ", []string{"git push"}},
		{"", nil},
		{"   ", nil},
	}

	for _, tc := range cases {
		got := llm.ParseCommands(tc.input)
		if len(got) != len(tc.want) {
			t.Errorf("ParseCommands(%q): got %v (len %d), want %v (len %d)",
				tc.input, got, len(got), tc.want, len(tc.want))
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("ParseCommands(%q)[%d]: got %q, want %q", tc.input, i, got[i], tc.want[i])
			}
		}
	}
}

func TestParseCommandsLimit(t *testing.T) {
	input := "a\nb\nc\nd\ne\nf\ng"
	got := llm.ParseCommands(input)
	if len(got) != 7 {
		t.Errorf("expected 7 commands, got %d", len(got))
	}
}

func TestBuildPrompt(t *testing.T) {
	c := context.ShellContext{
		LastCommand:  "git psuh origin main",
		LastExitCode: 127,
		WorkingDir:   "/home/user/project",
		GitBranch:    "main",
		RecentHistory: []context.HistoryEntry{
			{Index: 1, Command: "git status"},
			{Index: 2, Command: "git add ."},
		},
	}
	prompt := llm.BuildPrompt(c)
	for _, want := range []string{
		"git psuh origin main",
		"127",
		"/home/user/project",
		"main",
		"git status",
		"git add .",
	} {
		if !strings.Contains(prompt, want) {
			t.Errorf("BuildPrompt: expected prompt to contain %q", want)
		}
	}
}
