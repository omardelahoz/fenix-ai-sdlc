package kernel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/omardelahoz/fenix-ai-sdlc/core/pkg/contracts"
)

// ClaudeAdapter implements AIProvider for Anthropic's Claude.
type ClaudeAdapter struct {
	config *contracts.ClaudeConfig
	client *http.Client
}

// NewClaudeAdapter creates a new Claude adapter.
func NewClaudeAdapter(config *contracts.ClaudeConfig) *ClaudeAdapter {
	timeout := time.Duration(config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &ClaudeAdapter{
		config: config,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Chat sends a prompt to Claude and returns the response.
func (a *ClaudeAdapter) Chat(ctx context.Context, prompt contracts.Prompt) (contracts.Response, error) {
	// Build the request body
	requestBody := map[string]interface{}{
		"model":      a.config.Model,
		"max_tokens": a.config.MaxTokens,
		"messages":   buildMessages(prompt),
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return contracts.Response{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("https://api.anthropic.com/v1/messages")
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return contracts.Response{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("anthropic-dangerous-direct-browser-access", "true")

	// Send request
	resp, err := a.client.Do(req)
	if err != nil {
		return contracts.Response{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return contracts.Response{}, fmt.Errorf("Claude API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var claudeResponse struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
		Model string `json:"model"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&claudeResponse); err != nil {
		return contracts.Response{}, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract text from response
	var text string
	for _, content := range claudeResponse.Content {
		if content.Type == "text" {
			text = content.Text
			break
		}
	}

	return contracts.Response{
		Text:         text,
		TokenCount:   claudeResponse.Usage.InputTokens + claudeResponse.Usage.OutputTokens,
		Model:        claudeResponse.Model,
		FinishReason: "stop",
		Metadata: map[string]interface{}{
			"input_tokens":  claudeResponse.Usage.InputTokens,
			"output_tokens": claudeResponse.Usage.OutputTokens,
		},
	}, nil
}

// Stream sends a prompt to Claude and returns a stream of tokens.
func (a *ClaudeAdapter) Stream(ctx context.Context, prompt contracts.Prompt) (<-chan string, error) {
	ch := make(chan string, 100) // Buffered channel for better performance

	// Build the request body for streaming
	requestBody := map[string]interface{}{
		"model":      a.config.Model,
		"max_tokens": a.config.MaxTokens,
		"stream":     true,
		"messages":   buildMessages(prompt),
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		close(ch)
		return ch, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("https://api.anthropic.com/v1/messages")
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		close(ch)
		return ch, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("anthropic-dangerous-direct-browser-access", "true")

	// Send request
	resp, err := a.client.Do(req)
	if err != nil {
		close(ch)
		return ch, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		close(ch)
		return ch, fmt.Errorf("Claude API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Stream responses
	go func() {
		defer close(ch)
		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)
		for {
			var chunk struct {
				Type  string `json:"type"`
				Delta *struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"delta"`
				Message *struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				Error *struct {
					Type    string `json:"type"`
					Message string `json:"message"`
				} `json:"error"`
			}

			if err := decoder.Decode(&chunk); err != nil {
				if err == io.EOF {
					break
				}
				ch <- fmt.Sprintf("Error: %v", err)
				return
			}

			// Handle different event types
			switch chunk.Type {
			case "message_start":
				// Message started, no content yet
				continue

			case "content_block_start":
				// Content block started
				continue

			case "content_block_delta":
				// Streaming text content
				if chunk.Delta != nil && chunk.Delta.Type == "text_delta" {
					ch <- chunk.Delta.Text
				}

			case "message_delta":
				// Alternative message format
				if chunk.Message != nil && chunk.Message.Content != "" {
					ch <- chunk.Message.Content
				}

			case "message_stop":
				// Message completed
				return

			case "error":
				// Error occurred
				if chunk.Error != nil {
					ch <- fmt.Sprintf("Error: %s", chunk.Error.Message)
				}
				return

			case "ping":
				// Keep-alive ping, ignore
				continue
			}
		}
	}()

	return ch, nil
}

// ProviderName returns the name of the provider.
func (a *ClaudeAdapter) ProviderName() string {
	return "claude"
}

// AvailableModels returns the list of available Claude models.
func (a *ClaudeAdapter) AvailableModels() []string {
	return []string{
		"claude-3-5-sonnet-20241022",
		"claude-3-5-haiku-20241022",
		"claude-3-opus-20240229",
		"claude-3-sonnet-20240229",
		"claude-3-haiku-20240307",
	}
}

// buildMessages converts a Prompt to Claude's message format.
func buildMessages(prompt contracts.Prompt) []map[string]string {
	messages := []map[string]string{}

	// Add system message if present
	if prompt.SystemMessage != "" {
		messages = append(messages, map[string]string{
			"role":    "user",
			"content": prompt.SystemMessage,
		})
	}

	// Add user message
	if prompt.UserMessage != "" {
		messages = append(messages, map[string]string{
			"role":    "user",
			"content": prompt.UserMessage,
		})
	}

	// Add conversation history
	for _, msg := range prompt.Messages {
		messages = append(messages, map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	return messages
}
