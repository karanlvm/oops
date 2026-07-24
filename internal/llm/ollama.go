package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/karanlvm/oops/internal/context"
)

const (
	defaultOllamaBase  = "http://localhost:11434"
	defaultOllamaModel = "llama3.2"
)

type OllamaBackend struct {
	baseURL string
	model   string
}

func NewOllamaBackend(model string) *OllamaBackend {
	if model == "" {
		model = defaultOllamaModel
	}
	return &OllamaBackend{baseURL: defaultOllamaBase, model: model}
}

func (b *OllamaBackend) Name() string { return fmt.Sprintf("Ollama (%s)", b.model) }

func (b *OllamaBackend) Fix(ctx context.ShellContext) ([]string, error) {
	prompt := BuildPrompt(ctx)

	body, _ := json.Marshal(map[string]any{
		"model":      b.model,
		"prompt":     prompt,
		"stream":     false,
		"keep_alive": "30m",
	})

	req, err := http.NewRequest("POST", b.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Response string `json:"response"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse ollama response: %w", err)
	}

	return ParseCommands(result.Response), nil
}

func ollamaAvailable() bool {
	resp, err := http.Get(defaultOllamaBase + "/api/tags")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
