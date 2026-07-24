package rules

import (
	"strings"

	"github.com/karanlvm/oops/internal/context"
)

var pythonVerbs = []string{
	"build", "clean", "compile", "develop", "dist", "egg-info",
	"install", "sdist", "test", "upload", "wheel",
}

var pipSubcmds = []string{
	"install", "uninstall", "freeze", "list", "show", "search",
	"download", "wheel", "hash", "completion", "debug", "config",
	"cache", "index", "inspect",
}

type pythonRule struct{}

func (r *pythonRule) Name() string { return "python" }

func (r *pythonRule) Match(cmd string, exitCode int, _ context.ShellContext) bool {
	name := first(cmd)
	return name == "python" || name == "python3" || name == "python2" ||
		name == "pip" || name == "pip3" || name == "pip2" ||
		name == "pyhton" || name == "pytohn" || name == "pyton" ||
		name == "pyhton3" || name == "piip" || name == "pipp"
}

func (r *pythonRule) Fix(cmd string, exitCode int, _ context.ShellContext) []string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil
	}
	name := parts[0]
	rest := ""
	if len(parts) > 1 {
		rest = " " + strings.Join(parts[1:], " ")
	}

	// ── pip typos / version ──────────────────────────────────────────────────
	if name == "pip" || name == "piip" || name == "pipp" || name == "pip2" {
		// pip subcmd typo?
		if len(parts) >= 2 {
			sub := parts[1]
			if best, dist := closestMatch(sub, pipSubcmds); dist <= 1 && dist > 0 {
				return []string{"pip3 " + best + " " + strings.Join(parts[2:], " ")}
			}
		}
		return []string{"pip3" + rest}
	}

	// ── python typos → python3 ────────────────────────────────────────────────
	if name != "python3" {
		return []string{"python3" + rest}
	}

	// ── python3 verb typo ─────────────────────────────────────────────────────
	if exitCode != 0 && len(parts) >= 2 {
		// e.g. python3 -m pytest typo
		if parts[1] == "-m" && len(parts) >= 3 {
			return nil // don't try to fix module names
		}
		// Maybe they forgot -m for a known module runner?
		sub := parts[1]
		knownModules := []string{"pytest", "unittest", "http.server", "json.tool", "venv", "ensurepip"}
		for _, m := range knownModules {
			if sub == m {
				return []string{"python3 -m " + strings.Join(parts[1:], " ")}
			}
		}
	}

	return nil
}
