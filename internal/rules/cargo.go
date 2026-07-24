package rules

import (
	"strings"

	"github.com/karanlvm/oops/internal/context"
)

var cargoSubcmds = []string{
	"add", "bench", "build", "check", "clean", "clippy", "doc",
	"fetch", "fix", "fmt", "generate-lockfile", "help", "info",
	"init", "install", "locate-project", "login", "logout",
	"metadata", "new", "owner", "package", "pkgid", "publish",
	"remove", "report", "run", "rustc", "rustdoc", "search",
	"test", "tree", "uninstall", "update", "vendor", "verify-project",
	"version", "yank",
}

type cargoRule struct{}

func (r *cargoRule) Name() string { return "cargo" }

func (r *cargoRule) Match(cmd string, exitCode int, _ context.ShellContext) bool {
	name := first(cmd)
	return name == "cargo" || name == "cargoo" || name == "craog" || name == "carog"
}

func (r *cargoRule) Fix(cmd string, exitCode int, _ context.ShellContext) []string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil
	}
	name := parts[0]

	// ── binary typo ──────────────────────────────────────────────────────────
	if name != "cargo" {
		return []string{"cargo " + strings.Join(parts[1:], " ")}
	}

	if len(parts) < 2 {
		return nil
	}
	sub := parts[1]

	// ── subcommand typo ──────────────────────────────────────────────────────
	if best, dist := closestMatch(sub, cargoSubcmds); dist <= 2 && dist > 0 && best != sub {
		fixed := append([]string{"cargo", best}, parts[2:]...)
		return []string{strings.Join(fixed, " ")}
	}

	// ── cargo build/run failed → check first ─────────────────────────────────
	if (sub == "build" || sub == "run") && exitCode != 0 {
		return []string{"cargo check"}
	}

	// ── cargo test with no tests found ───────────────────────────────────────
	if sub == "test" && exitCode == 101 {
		return []string{"cargo test -- --nocapture"}
	}

	return nil
}
