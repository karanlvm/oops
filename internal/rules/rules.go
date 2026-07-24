package rules

import (
	"strings"

	"github.com/karanlvm/oops/internal/context"
)

// Rule is a single correction rule.
type Rule interface {
	Name() string
	Match(cmd string, exitCode int, ctx context.ShellContext) bool
	Fix(cmd string, exitCode int, ctx context.ShellContext) []string
}

var registry []Rule

func init() {
	registry = []Rule{
		new(gitRule),
		new(sudoRule),
		new(cdRule),
		new(pythonRule),
		new(nodeRule),
		new(dockerRule),
		new(brewRule),
		new(cargoRule),
		new(goRule),
		new(makeRule),
		new(typoRule), // generic fallback — must be last
	}
}

// Suggest returns the first matching set of corrected commands, or nil.
func Suggest(cmd string, exitCode int, ctx context.ShellContext) []string {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return nil
	}
	for _, r := range registry {
		if r.Match(cmd, exitCode, ctx) {
			if fix := r.Fix(cmd, exitCode, ctx); len(fix) > 0 {
				return fix
			}
		}
	}
	return nil
}

// first returns the first token of cmd.
func first(cmd string) string {
	f := strings.Fields(cmd)
	if len(f) == 0 {
		return ""
	}
	return f[0]
}
