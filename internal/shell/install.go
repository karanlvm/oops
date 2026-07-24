package shell

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed *.zsh *.bash *.fish
var shellFiles embed.FS

type Shell struct {
	Name      string
	RCFile    string
	HookFile  string
	SourceTag string
}

func Detect() (*Shell, error) {
	shellPath := os.Getenv("SHELL")
	switch {
	case strings.HasSuffix(shellPath, "zsh"):
		return &Shell{
			Name:     "zsh",
			RCFile:   filepath.Join(os.Getenv("HOME"), ".zshrc"),
			HookFile: "oops.zsh",
		}, nil
	case strings.HasSuffix(shellPath, "bash"):
		return &Shell{
			Name:     "bash",
			RCFile:   filepath.Join(os.Getenv("HOME"), ".bashrc"),
			HookFile: "oops.bash",
		}, nil
	case strings.HasSuffix(shellPath, "fish"):
		return &Shell{
			Name:     "fish",
			RCFile:   filepath.Join(os.Getenv("HOME"), ".config", "fish", "config.fish"),
			HookFile: "oops.fish",
		}, nil
	default:
		return nil, fmt.Errorf("unsupported shell: %s (supported: zsh, bash, fish)", shellPath)
	}
}

const sourceLine = "# oops shell integration\nsource ~/.config/oops/"

func Install() error {
	sh, err := Detect()
	if err != nil {
		return err
	}
	fmt.Printf("  ✓ Detected shell: %s\n", sh.Name)

	// Write hook file to ~/.config/oops/
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "oops")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	hookContent, err := shellFiles.ReadFile(sh.HookFile)
	if err != nil {
		return fmt.Errorf("embedded hook not found for %s: %w", sh.Name, err)
	}

	hookDest := filepath.Join(configDir, sh.HookFile)
	if err := os.WriteFile(hookDest, hookContent, 0644); err != nil {
		return fmt.Errorf("failed to write hook file: %w", err)
	}
	fmt.Printf("  ✓ Wrote hooks to %s\n", hookDest)

	// Add source line to RC file if not already present
	rcContent, _ := os.ReadFile(sh.RCFile)
	line := sourceLine + sh.HookFile
	if strings.Contains(string(rcContent), line) {
		fmt.Printf("  ✓ %s already configured\n", sh.RCFile)
	} else {
		f, err := os.OpenFile(sh.RCFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", sh.RCFile, err)
		}
		defer f.Close()
		fmt.Fprintf(f, "\n%s\n", line)
		fmt.Printf("  ✓ Added source line to %s\n", sh.RCFile)
	}

	// Check for LLM backend
	switch {
	case os.Getenv("ANTHROPIC_API_KEY") != "":
		fmt.Println("  ✓ LLM backend: Claude (ANTHROPIC_API_KEY set)")
	case os.Getenv("OPENAI_API_KEY") != "":
		fmt.Println("  ✓ LLM backend: OpenAI (OPENAI_API_KEY set)")
	default:
		fmt.Println("  ⚠ No LLM key found. Set ANTHROPIC_API_KEY or OPENAI_API_KEY in your shell profile.")
	}

	fmt.Printf("\n  Ready. Restart your shell or run: source %s\n", sh.RCFile)
	fmt.Println("  Then run a failing command and type: oops")
	return nil
}

func Uninstall() error {
	sh, err := Detect()
	if err != nil {
		return err
	}

	rcContent, err := os.ReadFile(sh.RCFile)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", sh.RCFile, err)
	}

	line := sourceLine + sh.HookFile
	updated := strings.ReplaceAll(string(rcContent), "\n"+line+"\n", "\n")
	updated = strings.ReplaceAll(updated, line+"\n", "")
	updated = strings.ReplaceAll(updated, "\n"+line, "")

	if err := os.WriteFile(sh.RCFile, []byte(updated), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", sh.RCFile, err)
	}
	fmt.Printf("  ✓ Removed oops from %s\n", sh.RCFile)
	return nil
}
