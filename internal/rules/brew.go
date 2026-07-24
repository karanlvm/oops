package rules

import (
	"strings"

	"github.com/karanlvm/oops/internal/context"
)

var brewSubcmds = []string{
	"install", "uninstall", "reinstall", "upgrade", "update",
	"list", "info", "search", "options", "home",
	"pin", "unpin", "link", "unlink", "switch",
	"tap", "untap", "tap-info",
	"cask", "services", "cleanup", "autoremove",
	"doctor", "audit", "fetch", "deps",
	"leaves", "uses", "missing", "outdated",
	"config", "env", "prefix", "cellar",
	"bundle", "test", "create", "edit",
	"analytics", "shellenv",
	"completions", "commands",
}

type brewRule struct{}

func (r *brewRule) Name() string { return "brew" }

func (r *brewRule) Match(cmd string, exitCode int, _ context.ShellContext) bool {
	name := first(cmd)
	return name == "brew" || name == "bre" || name == "bew" || name == "brwe" || name == "breu"
}

func (r *brewRule) Fix(cmd string, exitCode int, _ context.ShellContext) []string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil
	}
	name := parts[0]

	// ── binary typo ──────────────────────────────────────────────────────────
	if name != "brew" {
		return []string{"brew " + strings.Join(parts[1:], " ")}
	}

	if len(parts) < 2 {
		return nil
	}
	sub := parts[1]

	// ── subcommand typo ──────────────────────────────────────────────────────
	if best, dist := closestMatch(sub, brewSubcmds); dist <= 2 && dist > 0 && best != sub {
		fixed := append([]string{"brew", best}, parts[2:]...)
		return []string{strings.Join(fixed, " ")}
	}

	// ── brew install <cask> → brew install --cask <pkg> ─────────────────────
	// (heuristic: if install fails and pkg looks like a GUI app)
	if sub == "install" && exitCode != 0 && len(parts) >= 3 {
		pkg := parts[2]
		if !strings.HasPrefix(pkg, "--") {
			return []string{"brew install --cask " + pkg}
		}
	}

	// ── brew update before upgrade (common pattern) ──────────────────────────
	if sub == "upgrade" && exitCode != 0 {
		return []string{"brew update", strings.Join(parts, " ")}
	}

	return nil
}
