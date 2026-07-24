package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/karanlvm/oops/internal/context"
	"github.com/karanlvm/oops/internal/llm"
	"github.com/karanlvm/oops/internal/rules"
	"github.com/karanlvm/oops/internal/runner"
	"github.com/karanlvm/oops/internal/safety"
	"github.com/karanlvm/oops/internal/shell"
)

var version = "0.1.0"

func main() {
	args := os.Args[1:]

	if len(args) > 0 {
		switch args[0] {
		case "--install":
			fmt.Println("oops: installing shell integration")
			if err := shell.Install(); err != nil {
				fatalf("install failed: %v\n", err)
			}
			return
		case "--uninstall":
			fmt.Println("oops: removing shell integration")
			if err := shell.Uninstall(); err != nil {
				fatalf("uninstall failed: %v\n", err)
			}
			return
		case "--version", "-v":
			fmt.Println("oops", version)
			return
		case "--help", "-h":
			printHelp()
			return
		}
	}

	yesFlag := containsFlag(args, "--yes", "-y")
	dryRun := containsFlag(args, "--dry-run")
	explain := containsFlag(args, "--explain")

	ctx := context.FromEnv()
	if !ctx.IsValid() {
		fatalf("oops: no failed command detected\n\nMake sure oops is installed in your shell:\n  oops --install\n")
	}

	// ── Layer 1: local rule engine (instant, no API key needed) ─────────────
	commands := rules.Suggest(ctx.LastCommand, ctx.LastExitCode, ctx)
	source := "local"

	// ── Layer 2: LLM fallback ────────────────────────────────────────────────
	if len(commands) == 0 {
		backend, err := llm.Detect()
		if err != nil {
			fatalf("oops: no fix found locally and no LLM configured\n\n%v\n", err)
		}
		fmt.Printf("oops: asking %s...\n", backend.Name())
		commands, err = backend.Fix(ctx)
		if err != nil {
			fatalf("oops: LLM request failed: %v\n", err)
		}
		source = backend.Name()
	}

	if len(commands) == 0 {
		fatalf("oops: no fix found\n")
	}
	_ = source

	// Limit to 5 commands
	if len(commands) > 5 {
		commands = commands[:5]
	}

	fmt.Println()
	if len(commands) == 1 {
		fmt.Printf("oops: detected a fix\n\n  %s\n\n", commands[0])
	} else {
		fmt.Printf("oops: detected a fix (%d commands)\n\n", len(commands))
		for i, cmd := range commands {
			fmt.Printf("  %d  %s\n", i+1, cmd)
		}
		fmt.Println()
	}

	if explain {
		fmt.Println("(--explain not yet implemented — coming in a future release)")
		fmt.Println()
	}

	if dryRun {
		fmt.Println("(dry run — not executing)")
		return
	}

	safetyResult := safety.Check(commands)
	if safetyResult.IsDestructive {
		fmt.Printf("\033[33mWarning:\033[0m potentially destructive command detected (%s)\n", safetyResult.Reason)
		fmt.Print("Type YES to confirm, or anything else to abort: ")
		var response string
		fmt.Scanln(&response)
		if response != "YES" {
			fmt.Println("Aborted.")
			return
		}
	} else if !yesFlag {
		fmt.Print("Execute? [Y/n/e(dit)] ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		switch response {
		case "", "y":
			// proceed
		case "n":
			fmt.Println("Aborted.")
			return
		case "e":
			edited, err := editCommands(commands)
			if err != nil {
				fatalf("oops: edit failed: %v\n", err)
			}
			commands = edited
		default:
			fmt.Println("Aborted.")
			return
		}
	}

	fmt.Println()
	if err := runner.Run(commands); err != nil {
		fatalf("oops: %v\n", err)
	}
}

func editCommands(commands []string) ([]string, error) {
	tmp, err := os.CreateTemp("", "oops-edit-*.sh")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmp.Name())

	for _, cmd := range commands {
		fmt.Fprintln(tmp, cmd)
	}
	tmp.Close()

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	c := fmt.Sprintf("%s %s", editor, tmp.Name())
	editCmd := runner.ShellCmd(c)
	editCmd.Stdin = os.Stdin
	editCmd.Stdout = os.Stdout
	editCmd.Stderr = os.Stderr
	if err := editCmd.Run(); err != nil {
		return nil, err
	}

	content, err := os.ReadFile(tmp.Name())
	if err != nil {
		return nil, err
	}

	return llm.ParseCommands(string(content)), nil
}

func containsFlag(args []string, flags ...string) bool {
	for _, arg := range args {
		for _, f := range flags {
			if arg == f {
				return true
			}
		}
	}
	return false
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

func printHelp() {
	fmt.Print(`oops — self-healing shell powered by session-aware AI

Usage:
  oops              fix the last failed command
  oops --yes        skip confirmation prompt (safety layer still applies)
  oops --dry-run    show the fix without executing
  oops --explain    show why the fix works (coming soon)
  oops --install    add shell hooks to your rc file
  oops --uninstall  remove shell hooks
  oops --version    print version

LLM backends (checked in order):
  ANTHROPIC_API_KEY  → Claude haiku (recommended)
  OPENAI_API_KEY     → gpt-4o-mini (or any OpenAI-compatible API)
  Ollama             → local model on localhost:11434

`)
}
