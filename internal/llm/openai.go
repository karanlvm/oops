package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/karan/oops/internal/context"
)

const (
	defaultOpenAIBase  = "https://api.openai.com/v1"
	defaultOpenAIModel = "gpt-4o-mini"
)

type OpenAIBackend struct {
	apiKey  string
	baseURL string
	model   string
}

func NewOpenAIBackend(apiKey, baseURL string) *OpenAIBackend {
	if baseURL == "" {
		baseURL = defaultOpenAIBase
	}
	return &OpenAIBackend{apiKey: apiKey, baseURL: baseURL, model: defaultOpenAIModel}
}

func (b *OpenAIBackend) Name() string { return "OpenAI-compatible" }

func (b *OpenAIBackend) Fix(ctx context.ShellContext) ([]string, error) {
	prompt := BuildPrompt(ctx)

	body, _ := json.Marshal(map[string]any{
		"model":      b.model,
		"max_tokens": 256,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	})

	req, err := http.NewRequest("POST", b.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+b.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai API request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai API error %d: %s", resp.StatusCode, string(data))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse openai response: %w", err)
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("empty response from openai")
	}

	return ParseCommands(result.Choices[0].Message.Content), nil
}
