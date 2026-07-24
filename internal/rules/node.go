package rules

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/karanlvm/oops/internal/context"
)

var npmSubcmds = []string{
	"install", "uninstall", "update", "run", "start", "test", "build",
	"publish", "pack", "link", "unlink", "ci", "audit", "fund",
	"init", "outdated", "ls", "prune", "dedupe", "exec", "diff",
	"version", "view", "search", "login", "logout", "whoami",
	"config", "set", "get", "cache", "help",
}

var yarnSubcmds = []string{
	"add", "remove", "install", "upgrade", "run", "start", "test", "build",
	"publish", "link", "unlink", "init", "why", "audit", "outdated",
	"workspace", "workspaces", "version", "info", "list", "config",
	"cache", "login", "logout", "exec", "dlx", "set",
}

var pnpmSubcmds = []string{
	"add", "remove", "install", "update", "run", "start", "test", "build",
	"publish", "link", "unlink", "init", "why", "audit", "outdated",
	"exec", "dlx", "list", "recursive", "store", "config",
}

type nodeRule struct{}

func (r *nodeRule) Name() string { return "node" }

func (r *nodeRule) Match(cmd string, exitCode int, _ context.ShellContext) bool {
	name := first(cmd)
	return name == "npm" || name == "yarn" || name == "pnpm" ||
		name == "npx" || name == "node" ||
		name == "nom" || name == "nmp" || name == "noed" || name == "ndoe"
}

func (r *nodeRule) Fix(cmd string, exitCode int, ctx context.ShellContext) []string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil
	}
	name := parts[0]

	// ── plain binary typos ────────────────────────────────────────────────────
	if name == "nom" || name == "nmp" {
		return []string{"npm " + strings.Join(parts[1:], " ")}
	}
	if name == "noed" || name == "ndoe" {
		return []string{"node " + strings.Join(parts[1:], " ")}
	}

	// ── npm ───────────────────────────────────────────────────────────────────
	if name == "npm" {
		if len(parts) < 2 {
			return nil
		}
		sub := parts[1]

		// npm run <script-typo> → find closest script in package.json
		if sub == "run" && len(parts) >= 3 {
			scripts := packageJSONScripts(ctx.WorkingDir)
			if len(scripts) > 0 {
				typo := parts[2]
				if best, dist := closestMatch(typo, scripts); dist <= 2 && dist > 0 {
					return []string{"npm run " + best}
				}
			}
		}

		// npm <subcmd-typo>
		if best, dist := closestMatch(sub, npmSubcmds); dist <= 2 && dist > 0 && best != sub {
			fixed := append([]string{"npm", best}, parts[2:]...)
			return []string{strings.Join(fixed, " ")}
		}

		// npm install missing (ran a script before installing)
		if sub == "run" && exitCode != 0 {
			return []string{"npm install", strings.Join(parts, " ")}
		}
	}

	// ── yarn ──────────────────────────────────────────────────────────────────
	if name == "yarn" {
		if len(parts) < 2 {
			return nil
		}
		sub := parts[1]

		// yarn <script-typo> — yarn runs scripts directly
		scripts := packageJSONScripts(ctx.WorkingDir)
		if len(scripts) > 0 {
			if best, dist := closestMatch(sub, scripts); dist <= 2 && dist > 0 {
				return []string{"yarn " + best}
			}
		}

		// yarn <subcmd-typo>
		if best, dist := closestMatch(sub, yarnSubcmds); dist <= 1 && dist > 0 && best != sub {
			fixed := append([]string{"yarn", best}, parts[2:]...)
			return []string{strings.Join(fixed, " ")}
		}
	}

	// ── pnpm ─────────────────────────────────────────────────────────────────
	if name == "pnpm" {
		if len(parts) < 2 {
			return nil
		}
		sub := parts[1]

		if sub == "run" && len(parts) >= 3 {
			scripts := packageJSONScripts(ctx.WorkingDir)
			if len(scripts) > 0 {
				if best, dist := closestMatch(parts[2], scripts); dist <= 2 && dist > 0 {
					return []string{"pnpm run " + best}
				}
			}
		}

		if best, dist := closestMatch(sub, pnpmSubcmds); dist <= 1 && dist > 0 && best != sub {
			fixed := append([]string{"pnpm", best}, parts[2:]...)
			return []string{strings.Join(fixed, " ")}
		}
	}

	// ── npx ──────────────────────────────────────────────────────────────────
	if name == "npx" && exitCode == 1 && len(parts) >= 2 {
		// Package not found — suggest installing it
		pkg := parts[1]
		if !strings.HasPrefix(pkg, "-") {
			return []string{
				"npm install -g " + pkg,
				strings.Join(parts, " "),
			}
		}
	}

	return nil
}

// packageJSONScripts reads script names from the nearest package.json.
func packageJSONScripts(cwd string) []string {
	if cwd == "" {
		return nil
	}
	data, err := os.ReadFile(filepath.Join(cwd, "package.json"))
	if err != nil {
		return nil
	}
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil
	}
	names := make([]string, 0, len(pkg.Scripts))
	for k := range pkg.Scripts {
		names = append(names, k)
	}
	return names
}
