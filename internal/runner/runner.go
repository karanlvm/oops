package runner

import (
	"fmt"
	"os"
	"os/exec"
)

func userShell() string {
	if s := os.Getenv("SHELL"); s != "" {
		return s
	}
	return "/bin/sh"
}

// ShellCmd returns an exec.Cmd that runs the given string in the user's shell.
func ShellCmd(command string) *exec.Cmd {
	return exec.Command(userShell(), "-c", command)
}

// Run executes commands in sequence, stopping on first non-zero exit.
func Run(commands []string) error {
	for _, cmd := range commands {
		fmt.Printf("  \033[2m$\033[0m %s\n", cmd)
		c := ShellCmd(cmd)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return fmt.Errorf("command failed: %s\n%w", cmd, err)
		}
	}
	return nil
}
