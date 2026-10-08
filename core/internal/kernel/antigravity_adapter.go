package kernel

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/omardelahoz/fenix-ai-sdlc/core/pkg/contracts"
)

// AntigravityAdapter implements AIProvider for Antigravity platform.
// Antigravity is an advanced AI platform for software engineering.
type AntigravityAdapter struct {
	config *contracts.AntigravityConfig
	client *http.Client
}

// NewAntigravityAdapter creates a new Antigravity adapter.
func NewAntigravityAdapter(config *contracts.AntigravityConfig) *AntigravityAdapter {
	timeout := time.Duration(config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &AntigravityAdapter{
		config: config,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Chat sends a prompt to Antigravity and returns the response.
func (a *AntigravityAdapter) Chat(ctx context.Context, prompt contracts.Prompt) (contracts.Response, error) {
	// Build the request body (format depends on Antigravity's API)
	requestBody := map[string]interface{}{
		"model":        a.config.Model,
		"max_tokens":   a.config.MaxTokens,
		"temperature":  a.config.Temperature,
		"messages":     buildMessages(prompt),
		"workspace_id": a.config.WorkspaceID,
		"project_id":   a.config.ProjectID,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return contracts.Response{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	// Note: BaseURL should be configured to point to Antigravity's API endpoint
	url := fmt.Sprintf("%s/v1/chat/completions", a.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return contracts.Response{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.config.APIKey))

	// Send request
	resp, err := a.client.Do(req)
	if err != nil {
		return contracts.Response{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return contracts.Response{}, fmt.Errorf("Antigravity API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response (format depends on Antigravity's API response structure)
	var antigravityResponse struct {
		Choices []struct {
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
		Model string `json:"model"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&antigravityResponse); err != nil {
		return contracts.Response{}, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract text from response
	var text string
	if len(antigravityResponse.Choices) > 0 {
		text = antigravityResponse.Choices[0].Message.Content
	}

	finishReason := "stop"
	if len(antigravityResponse.Choices) > 0 {
		finishReason = antigravityResponse.Choices[0].FinishReason
	}

	return contracts.Response{
		Text:         text,
		TokenCount:   antigravityResponse.Usage.TotalTokens,
		Model:        antigravityResponse.Model,
		FinishReason: finishReason,
		Metadata: map[string]interface{}{
			"prompt_tokens":     antigravityResponse.Usage.PromptTokens,
			"completion_tokens": antigravityResponse.Usage.CompletionTokens,
			"workspace_id":      a.config.WorkspaceID,
			"project_id":        a.config.ProjectID,
		},
	}, nil
}

// Stream sends a prompt to Antigravity and returns a stream of tokens.
func (a *AntigravityAdapter) Stream(ctx context.Context, prompt contracts.Prompt) (<-chan string, error) {
	ch := make(chan string, 100) // Buffered channel for better performance

	// Build the request body for streaming
	requestBody := map[string]interface{}{
		"model":        a.config.Model,
		"max_tokens":   a.config.MaxTokens,
		"temperature":  a.config.Temperature,
		"stream":       true,
		"messages":     buildMessages(prompt),
		"workspace_id": a.config.WorkspaceID,
		"project_id":   a.config.ProjectID,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		close(ch)
		return ch, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/v1/chat/completions", a.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		close(ch)
		return ch, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.config.APIKey))
	req.Header.Set("Accept", "text/event-stream")

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
		return ch, fmt.Errorf("Antigravity API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Stream responses (SSE format)
	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			// Skip empty lines and comments
			if line == "" || strings.HasPrefix(line, ":") {
				continue
			}

			// SSE format: "data: {...}"
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")

			// Check for [DONE] marker
			if data == "[DONE]" {
				return
			}

			// Parse JSON chunk
			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
					FinishReason string `json:"finish_reason"`
				} `json:"choices"`
				Error *struct {
					Message string `json:"message"`
					Type    string `json:"type"`
				} `json:"error"`
			}

			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				// Skip malformed chunks
				continue
			}

			// Handle errors
			if chunk.Error != nil {
				ch <- fmt.Sprintf("Error: %s", chunk.Error.Message)
				return
			}

			// Extract content from choices
			if len(chunk.Choices) > 0 {
				content := chunk.Choices[0].Delta.Content
				if content != "" {
					ch <- content
				}

				// Check if generation is complete
				if chunk.Choices[0].FinishReason != "" {
					return
				}
			}
		}

		// Check for scanner errors
		if err := scanner.Err(); err != nil {
			ch <- fmt.Sprintf("Stream error: %v", err)
		}
	}()

	return ch, nil
}

// ProviderName returns the name of the provider.
func (a *AntigravityAdapter) ProviderName() string {
	return "antigravity"
}

// AvailableModels returns the list of available Antigravity models.
func (a *AntigravityAdapter) AvailableModels() []string {
	// These are example models - actual models depend on Antigravity's offerings
	return []string{
		"antigravity-coding-v1",
		"antigravity-architecture-v1",
		"antigravity-analysis-v1",
	}
}

// buildMessages converts a Prompt to Antigravity's message format.
func buildMessages(prompt contracts.Prompt) []map[string]string {
	messages := []map[string]string{}

	// Add system message if present
	if prompt.SystemMessage != "" {
		messages = append(messages, map[string]string{
			"role":    "system",
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
