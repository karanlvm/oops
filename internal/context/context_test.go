package context_test

import (
	"os"
	"testing"

	ctx "github.com/karan/oops/internal/context"
)

func TestIsValid(t *testing.T) {
	cases := []struct {
		c     ctx.ShellContext
		valid bool
	}{
		{ctx.ShellContext{LastCommand: "git psuh", LastExitCode: 127}, true},
		{ctx.ShellContext{LastCommand: "npm rnu dev", LastExitCode: 1}, true},
		{ctx.ShellContext{LastCommand: "", LastExitCode: 1}, false},
		{ctx.ShellContext{LastCommand: "ls", LastExitCode: 0}, false},
		{ctx.ShellContext{LastCommand: "", LastExitCode: 0}, false},
	}
	for _, tc := range cases {
		got := tc.c.IsValid()
		if got != tc.valid {
			t.Errorf("IsValid(%+v) = %v, want %v", tc.c, got, tc.valid)
		}
	}
}

func TestFromEnv(t *testing.T) {
	os.Setenv("OOPS_LAST_CMD", "gti status")
	os.Setenv("OOPS_LAST_EXIT", "127")
	os.Setenv("OOPS_CWD", "/home/user/project")
	os.Setenv("OOPS_GIT_BRANCH", "main")
	defer func() {
		os.Unsetenv("OOPS_LAST_CMD")
		os.Unsetenv("OOPS_LAST_EXIT")
		os.Unsetenv("OOPS_CWD")
		os.Unsetenv("OOPS_GIT_BRANCH")
	}()

	c := ctx.FromEnv()
	if c.LastCommand != "gti status" {
		t.Errorf("LastCommand = %q, want %q", c.LastCommand, "gti status")
	}
	if c.LastExitCode != 127 {
		t.Errorf("LastExitCode = %d, want 127", c.LastExitCode)
	}
	if c.WorkingDir != "/home/user/project" {
		t.Errorf("WorkingDir = %q", c.WorkingDir)
	}
	if c.GitBranch != "main" {
		t.Errorf("GitBranch = %q", c.GitBranch)
	}
	if !c.IsValid() {
		t.Error("expected IsValid() = true")
	}
}

func TestHistoryText(t *testing.T) {
	c := ctx.ShellContext{
		RecentHistory: []ctx.HistoryEntry{
			{Index: 1, Command: "git status"},
			{Index: 2, Command: "git add ."},
		},
	}
	got := c.HistoryText()
	if got == "" {
		t.Error("expected non-empty HistoryText")
	}
}
