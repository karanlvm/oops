package rules

import (
	"strings"

	"github.com/karan/oops/internal/context"
)

var dockerSubcmds = []string{
	"build", "buildx", "commit", "container", "cp", "create", "diff",
	"exec", "export", "history", "image", "images", "import", "info",
	"inspect", "kill", "load", "login", "logout", "logs", "network",
	"pause", "plugin", "port", "ps", "pull", "push", "rename", "restart",
	"rm", "rmi", "run", "save", "search", "start", "stats", "stop",
	"system", "tag", "top", "trust", "unpause", "update", "version",
	"volume", "wait",
}

var dockerComposeSubcmds = []string{
	"build", "config", "create", "down", "events", "exec", "images",
	"kill", "logs", "pause", "port", "ps", "pull", "push", "restart",
	"rm", "run", "scale", "start", "stop", "top", "unpause", "up",
	"version",
}

type dockerRule struct{}

func (r *dockerRule) Name() string { return "docker" }

func (r *dockerRule) Match(cmd string, exitCode int, _ context.ShellContext) bool {
	name := first(cmd)
	return name == "docker" || name == "docker-compose" ||
		name == "dokcer" || name == "dcoker" || name == "docekr" || name == "doker"
}

func (r *dockerRule) Fix(cmd string, exitCode int, _ context.ShellContext) []string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil
	}
	name := parts[0]

	// ── binary typo ──────────────────────────────────────────────────────────
	if name != "docker" && name != "docker-compose" {
		rest := strings.Join(parts[1:], " ")
		if strings.Contains(name, "compose") {
			return []string{"docker compose " + rest}
		}
		return []string{"docker " + rest}
	}

	// ── docker-compose → docker compose ─────────────────────────────────────
	if name == "docker-compose" {
		return []string{"docker compose " + strings.Join(parts[1:], " ")}
	}

	// ── docker subcommand ────────────────────────────────────────────────────
	if len(parts) < 2 {
		return nil
	}
	sub := parts[1]

	// docker <subcmd-typo>
	if best, dist := closestMatch(sub, dockerSubcmds); dist <= 2 && dist > 0 && best != sub {
		fixed := append([]string{"docker", best}, parts[2:]...)
		return []string{strings.Join(fixed, " ")}
	}

	// docker run / exec / start permission denied → sudo
	if exitCode == 1 && (sub == "run" || sub == "ps" || sub == "exec" || sub == "images") {
		return []string{"sudo " + strings.Join(parts, " ")}
	}

	return nil
}
