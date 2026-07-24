package context

import (
	"os"
	"strconv"
	"strings"
)

type HistoryEntry struct {
	Index   int
	Command string
}

type ShellContext struct {
	LastCommand   string
	LastExitCode  int
	RecentHistory []HistoryEntry
	WorkingDir    string
	GitBranch     string
	Output        string // stderr/stdout of the failed command (best-effort)
}

func FromEnv() ShellContext {
	exitCode, _ := strconv.Atoi(os.Getenv("OOPS_LAST_EXIT"))
	return ShellContext{
		LastCommand:   os.Getenv("OOPS_LAST_CMD"),
		LastExitCode:  exitCode,
		RecentHistory: parseHistory(os.Getenv("OOPS_HISTORY")),
		WorkingDir:    os.Getenv("OOPS_CWD"),
		GitBranch:     os.Getenv("OOPS_GIT_BRANCH"),
		Output:        os.Getenv("OOPS_OUTPUT"),
	}
}

// parseHistory parses the output of `fc -l -10` or `history 10`.
// Lines look like: "  123  git commit -m foo" or "123  git commit -m foo"
func parseHistory(raw string) []HistoryEntry {
	var entries []HistoryEntry
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Split off the leading index number
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}
		idx, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			continue
		}
		entries = append(entries, HistoryEntry{
			Index:   idx,
			Command: strings.TrimSpace(parts[1]),
		})
	}
	return entries
}

func (c ShellContext) IsValid() bool {
	return c.LastCommand != "" && c.LastExitCode != 0
}

func (c ShellContext) HistoryText() string {
	var sb strings.Builder
	for _, e := range c.RecentHistory {
		sb.WriteString("  $ ")
		sb.WriteString(e.Command)
		sb.WriteByte('\n')
	}
	return sb.String()
}
