package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/karan/oops/internal/context"
)

const anthropicModel = "claude-haiku-4-5"

type AnthropicBackend struct {
	apiKey string
}

func NewAnthropicBackend(apiKey string) *AnthropicBackend {
	return &AnthropicBackend{apiKey: apiKey}
}

func (b *AnthropicBackend) Name() string { return "Claude (Anthropic)" }

func (b *AnthropicBackend) Fix(ctx context.ShellContext) ([]string, error) {
	prompt := BuildPrompt(ctx)

	body, _ := json.Marshal(map[string]any{
		"model":      anthropicModel,
		"max_tokens": 256,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	})

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", b.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("anthropic API request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anthropic API error %d: %s", resp.StatusCode, string(data))
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse anthropic response: %w", err)
	}
	if len(result.Content) == 0 {
		return nil, fmt.Errorf("empty response from anthropic")
	}

	return ParseCommands(result.Content[0].Text), nil
}
