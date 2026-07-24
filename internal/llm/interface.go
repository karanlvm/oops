package llm

import (
	"fmt"
	"os"
	"strings"

	"github.com/karanlvm/oops/internal/context"
)

type Backend interface {
	Fix(ctx context.ShellContext) ([]string, error)
	Name() string
}

func BuildPrompt(ctx context.ShellContext) string {
	var sb strings.Builder
	sb.WriteString("You are a shell expert. The user ran a sequence of commands and the last one failed.\n\n")

	if history := ctx.HistoryText(); history != "" {
		sb.WriteString("Session history (newest last):\n")
		sb.WriteString(history)
		sb.WriteByte('\n')
	}

	if ctx.WorkingDir != "" {
		sb.WriteString("Working directory: ")
		sb.WriteString(ctx.WorkingDir)
		sb.WriteByte('\n')
	}
	if ctx.GitBranch != "" {
		sb.WriteString("Git branch: ")
		sb.WriteString(ctx.GitBranch)
		sb.WriteByte('\n')
	}
	if ctx.Output != "" {
		sb.WriteString("Command output (stderr):\n")
		sb.WriteString(ctx.Output)
		sb.WriteByte('\n')
	}

	sb.WriteString(fmt.Sprintf("\nThe last command was: %s\n", ctx.LastCommand))
	sb.WriteString(fmt.Sprintf("It failed with exit code %d.\n\n", ctx.LastExitCode))
	sb.WriteString("Use the command output above (if any) to understand the specific error before suggesting a fix.\n")
	sb.WriteString("Output ONLY the corrected command(s), one per line. No explanation, no markdown, no code blocks.\n")
	sb.WriteString("If multiple commands are needed to achieve the goal, output them in execution order.\n")
	sb.WriteString("Maximum 5 commands.")

	return sb.String()
}

// ParseCommands splits LLM output into individual commands, stripping markdown fences.
func ParseCommands(raw string) []string {
	var cmds []string
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line == "```" || strings.HasPrefix(line, "```") {
			continue
		}
		// Strip leading $ prompt if the model added one
		line = strings.TrimPrefix(line, "$ ")
		if line != "" {
			cmds = append(cmds, line)
		}
	}
	return cmds
}

// Detect picks the best available backend based on env vars.
func Detect() (Backend, error) {
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		return NewAnthropicBackend(key), nil
	}
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		return NewOpenAIBackend(key, os.Getenv("OPENAI_BASE_URL")), nil
	}
	if ollamaAvailable() {
		return NewOllamaBackend(""), nil
	}
	return nil, fmt.Errorf("no LLM backend configured\n\nSet one of:\n  ANTHROPIC_API_KEY  — Claude (recommended)\n  OPENAI_API_KEY     — OpenAI / Groq / any OpenAI-compatible API\n\nOr install Ollama (ollama.com) for a fully local, private backend.")
}
