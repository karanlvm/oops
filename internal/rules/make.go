package rules

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/karanlvm/oops/internal/context"
)

var makeTargetRe = regexp.MustCompile(`^([a-zA-Z0-9][a-zA-Z0-9_\-\.]*)\s*:`)

type makeRule struct{}

func (r *makeRule) Name() string { return "make" }

func (r *makeRule) Match(cmd string, exitCode int, _ context.ShellContext) bool {
	name := first(cmd)
	return (name == "make" || name == "mkae" || name == "amke" || name == "maek") && exitCode != 0
}

func (r *makeRule) Fix(cmd string, exitCode int, ctx context.ShellContext) []string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil
	}
	name := parts[0]

	// ── binary typo ──────────────────────────────────────────────────────────
	if name != "make" {
		return []string{"make " + strings.Join(parts[1:], " ")}
	}

	// ── no target specified ───────────────────────────────────────────────────
	if len(parts) == 1 {
		return nil
	}

	// ── find target typo ─────────────────────────────────────────────────────
	// The target is usually the first non-flag argument.
	var target string
	var flags []string
	for _, p := range parts[1:] {
		if strings.HasPrefix(p, "-") || strings.Contains(p, "=") {
			flags = append(flags, p)
		} else {
			target = p
			break
		}
	}
	if target == "" {
		return nil
	}

	targets := makefile(ctx.WorkingDir)
	if len(targets) == 0 {
		return nil
	}

	best, dist := closestMatch(target, targets)
	if dist == 0 || dist > 2 {
		return nil
	}

	fixed := append([]string{"make"}, flags...)
	fixed = append(fixed, best)
	return []string{strings.Join(fixed, " ")}
}

// makefile returns the list of non-hidden targets from the nearest Makefile.
func makefile(cwd string) []string {
	if cwd == "" {
		return nil
	}
	for _, name := range []string{"Makefile", "makefile", "GNUmakefile"} {
		f, err := os.Open(filepath.Join(cwd, name))
		if err != nil {
			continue
		}
		defer f.Close()

		var targets []string
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if m := makeTargetRe.FindStringSubmatch(line); m != nil {
				t := m[1]
				// Skip internal/phony-style targets
				if !strings.HasPrefix(t, ".") && t != "all" {
					targets = append(targets, t)
				}
			}
		}
		return targets
	}
	return nil
}
