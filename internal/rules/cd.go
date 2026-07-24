package rules

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/karan/oops/internal/context"
)

type cdRule struct{}

func (r *cdRule) Name() string { return "cd" }

func (r *cdRule) Match(cmd string, exitCode int, _ context.ShellContext) bool {
	return first(cmd) == "cd" && exitCode != 0
}

func (r *cdRule) Fix(cmd string, exitCode int, ctx context.ShellContext) []string {
	parts := strings.Fields(cmd)
	if len(parts) < 2 {
		return nil
	}
	target := parts[1]

	// Resolve the search directory and the base we're trying to match.
	searchIn := ctx.WorkingDir
	base := target

	if filepath.IsAbs(target) {
		searchIn = filepath.Dir(target)
		base = filepath.Base(target)
	} else if strings.Contains(target, "/") {
		// Relative path with slashes: fix only the last component.
		dir := filepath.Dir(target)
		base = filepath.Base(target)
		if ctx.WorkingDir != "" {
			searchIn = filepath.Join(ctx.WorkingDir, dir)
		}
	}

	if searchIn == "" {
		return nil
	}

	entries, err := os.ReadDir(searchIn)
	if err != nil {
		return nil
	}

	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}
	if len(dirs) == 0 {
		return nil
	}

	best, dist := closestMatch(base, dirs)
	if dist == 0 || dist > 2 {
		return nil
	}

	// Reconstruct the full path.
	var corrected string
	if filepath.IsAbs(target) {
		corrected = filepath.Join(filepath.Dir(target), best)
	} else if strings.Contains(target, "/") {
		corrected = filepath.Join(filepath.Dir(target), best)
	} else {
		corrected = best
	}

	return []string{"cd " + corrected}
}
