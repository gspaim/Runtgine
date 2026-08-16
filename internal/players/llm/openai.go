package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gspaim/Runtgine/internal/core/contextpack"
	"github.com/gspaim/Runtgine/internal/core/result"
)

type OpenAICompat struct {
	BaseURL string
	APIKey  string
	Model   string
	Client  *http.Client
}

func NewOpenAICompat(baseURL, apiKey, model string) *OpenAICompat {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &OpenAICompat{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
		Model:   model,
		Client:  &http.Client{Timeout: 90 * time.Second},
	}
}

func (c *OpenAICompat) Complete(ctx context.Context, pack contextpack.Pack, outputSchema json.RawMessage) (json.RawMessage, error) {
	body := map[string]any{
		"model": c.Model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt(outputSchema)},
			{"role": "user", "content": userPrompt(pack)},
		},
		"response_format": map[string]string{"type": "json_object"},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, result.Runtime(result.CodePlayerError, err.Error(), true, nil)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, result.Runtime(result.CodePlayerError, fmt.Sprintf("openai-compat %d: %s", resp.StatusCode, truncate(string(b), 400)), true, nil)
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		return nil, result.Runtime(result.CodePlayerError, "openai-compat decode: "+err.Error(), true, nil)
	}
	if len(parsed.Choices) == 0 {
		return nil, result.Runtime(result.CodePlayerError, "openai-compat: empty choices", true, nil)
	}
	return json.RawMessage(parsed.Choices[0].Message.Content), nil
}
