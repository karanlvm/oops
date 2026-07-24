package safety

import (
	"regexp"
	"strings"
)

var destructivePatterns = []*regexp.Regexp{
	regexp.MustCompile(`\brm\s+-[a-zA-Z]*r[a-zA-Z]*f[a-zA-Z]*\b`), // rm -rf, rm -fr
	regexp.MustCompile(`\brm\s+-[a-zA-Z]*f[a-zA-Z]*r[a-zA-Z]*\b`),
	regexp.MustCompile(`\brm\s+-r\b`),                                      // rm -r
	regexp.MustCompile(`\bgit\s+reset\s+--hard\b`),                         // git reset --hard
	regexp.MustCompile(`\bgit\s+(push|p)\s+.*--force\b`),                   // git push --force
	regexp.MustCompile(`\bgit\s+(push|p)\s+.*-f\b`),                        // git push -f
	regexp.MustCompile(`\bgit\s+clean\s+.*-[a-zA-Z]*f`),                    // git clean -f, -fd, -fx
	regexp.MustCompile(`(?i)\bDROP\s+(TABLE|DATABASE|SCHEMA)\b`),            // SQL DROP
	regexp.MustCompile(`(?i)\bTRUNCATE\s+TABLE\b`),                         // SQL TRUNCATE
	regexp.MustCompile(`(?i)\bDELETE\s+FROM\b`),                            // SQL DELETE (no WHERE check — always flag)
	regexp.MustCompile(`\bkubectl\s+(delete|destroy)\b`),                   // kubectl delete
	regexp.MustCompile(`\bterraform\s+destroy\b`),                          // terraform destroy
	regexp.MustCompile(`\|\s*(bash|sh|zsh|fish)\b`),                        // pipe to shell
}

type Result struct {
	IsDestructive bool
	Reason        string
}

// Check returns whether any command in the list is potentially destructive.
func Check(commands []string) Result {
	for _, cmd := range commands {
		for _, re := range destructivePatterns {
			if re.MatchString(cmd) {
				return Result{
					IsDestructive: true,
					Reason:        "contains: " + strings.TrimSpace(re.FindString(cmd)),
				}
			}
		}
	}
	return Result{}
}
