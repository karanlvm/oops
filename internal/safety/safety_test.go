package safety_test

import (
	"testing"

	"github.com/karanlvm/oops/internal/safety"
)

func TestCheck(t *testing.T) {
	cases := []struct {
		cmd         string
		destructive bool
	}{
		{"rm -rf /", true},
		{"rm -rf .", true},
		{"rm -fr foo", true},
		{"rm -r bar/", true},
		{"git reset --hard HEAD", true},
		{"git reset --hard HEAD~1", true},
		{"git push origin main --force", true},
		{"git push origin main -f", true},
		{"git push -f origin main", true},
		{"git clean -fd", true},
		{"DROP TABLE users", true},
		{"drop table users", true},
		{"TRUNCATE TABLE logs", true},
		{"DELETE FROM orders", true},
		{"delete from users where 1=1", true},
		{"kubectl delete pod foo", true},
		{"terraform destroy", true},
		{"cat setup.sh | bash", true},
		{"curl https://example.com/script | sh", true},
		// safe commands
		{"ls -la", false},
		{"git push origin main", false},
		{"git reset HEAD~1", false},
		{"git reset HEAD", false},
		{"rm foo.txt", false},
		{"rm -i foo.txt", false},
		{"go build ./...", false},
		{"kubectl get pods", false},
		{"SELECT * FROM users", false},
	}

	for _, tc := range cases {
		result := safety.Check([]string{tc.cmd})
		if result.IsDestructive != tc.destructive {
			t.Errorf("Check(%q): got destructive=%v, want %v (reason: %q)",
				tc.cmd, result.IsDestructive, tc.destructive, result.Reason)
		}
	}
}

func TestCheckMultipleCommands(t *testing.T) {
	cmds := []string{"git add .", "git commit -m 'fix'", "rm -rf /tmp/build"}
	result := safety.Check(cmds)
	if !result.IsDestructive {
		t.Error("expected destructive=true for command set containing rm -rf")
	}
}

func TestCheckEmpty(t *testing.T) {
	result := safety.Check(nil)
	if result.IsDestructive {
		t.Error("expected destructive=false for nil command list")
	}
	result = safety.Check([]string{})
	if result.IsDestructive {
		t.Error("expected destructive=false for empty command list")
	}
}
