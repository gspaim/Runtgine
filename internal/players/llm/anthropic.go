package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gspaim/Runtgine/internal/core/contextpack"
	"github.com/gspaim/Runtgine/internal/core/result"
)

type Anthropic struct {
	APIKey  string
	Model   string
	BaseURL string
	Client  *http.Client
}

func NewAnthropic(apiKey, model string) *Anthropic {
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}
	return &Anthropic{
		APIKey:  apiKey,
		Model:   model,
		BaseURL: "https://api.anthropic.com/v1/messages",
		Client:  &http.Client{Timeout: 90 * time.Second},
	}
}

func (c *Anthropic) Complete(ctx context.Context, pack contextpack.Pack, outputSchema json.RawMessage) (json.RawMessage, error) {
	body := map[string]any{
		"model":      c.Model,
		"max_tokens": 2048,
		"system":     systemPrompt(outputSchema),
		"messages": []map[string]string{
			{"role": "user", "content": userPrompt(pack)},
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, result.Runtime(result.CodePlayerError, err.Error(), true, nil)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, result.Runtime(result.CodePlayerError, fmt.Sprintf("anthropic %d: %s", resp.StatusCode, truncate(string(b), 400)), true, nil)
	}
	var parsed struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		return nil, result.Runtime(result.CodePlayerError, "anthropic decode: "+err.Error(), true, nil)
	}
	var text string
	for _, c := range parsed.Content {
		if c.Type == "text" {
			text += c.Text
		}
	}
	if text == "" {
		return nil, result.Runtime(result.CodePlayerError, "anthropic: empty content", true, nil)
	}
	return json.RawMessage(text), nil
}
