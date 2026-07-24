package rules

import (
	"strings"

	"github.com/karan/oops/internal/context"
)

var goSubcmds = []string{
	"build", "clean", "doc", "env", "bug", "fix", "fmt",
	"generate", "get", "install", "list", "mod", "work",
	"run", "test", "tool", "version", "vet",
}

var goModSubcmds = []string{
	"download", "edit", "graph", "init", "tidy", "vendor", "verify", "why",
}

type goRule struct{}

func (r *goRule) Name() string { return "go" }

func (r *goRule) Match(cmd string, exitCode int, _ context.ShellContext) bool {
	name := first(cmd)
	return name == "go" || name == "og"
}

func (r *goRule) Fix(cmd string, exitCode int, _ context.ShellContext) []string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil
	}
	name := parts[0]

	// ── binary typo ──────────────────────────────────────────────────────────
	if name != "go" {
		return []string{"go " + strings.Join(parts[1:], " ")}
	}

	if len(parts) < 2 {
		return nil
	}
	sub := parts[1]

	// ── go mod <sub-typo> ────────────────────────────────────────────────────
	if sub == "mod" && len(parts) >= 3 {
		modSub := parts[2]
		if best, dist := closestMatch(modSub, goModSubcmds); dist <= 1 && dist > 0 && best != modSub {
			fixed := append([]string{"go", "mod", best}, parts[3:]...)
			return []string{strings.Join(fixed, " ")}
		}
	}

	// ── subcommand typo ──────────────────────────────────────────────────────
	if best, dist := closestMatch(sub, goSubcmds); dist <= 2 && dist > 0 && best != sub {
		fixed := append([]string{"go", best}, parts[2:]...)
		return []string{strings.Join(fixed, " ")}
	}

	// ── go run without .go suffix ─────────────────────────────────────────────
	if sub == "run" && len(parts) >= 3 && exitCode != 0 {
		file := parts[2]
		if !strings.HasSuffix(file, ".go") && !strings.HasPrefix(file, "-") {
			return []string{"go run " + file + ".go" + " " + strings.Join(parts[3:], " ")}
		}
	}

	// ── go test needs ./... ──────────────────────────────────────────────────
	if sub == "test" && len(parts) == 2 && exitCode != 0 {
		return []string{"go test ./..."}
	}

	// ── go build: missing module deps ────────────────────────────────────────
	if (sub == "build" || sub == "run") && exitCode != 0 {
		return []string{"go mod tidy", strings.Join(parts, " ")}
	}

	return nil
}
