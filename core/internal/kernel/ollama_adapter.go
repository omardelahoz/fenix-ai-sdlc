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

// OllamaAdapter implements AIProvider for Ollama (local LLMs).
// Ollama allows running models like Llama, Mistral, etc. locally.
type OllamaAdapter struct {
	config *contracts.OllamaConfig
	client *http.Client
}

// NewOllamaAdapter creates a new Ollama adapter.
func NewOllamaAdapter(config *contracts.OllamaConfig) *OllamaAdapter {
	// Set default host if not specified
	if config.Host == "" {
		config.Host = "localhost:11434"
	}

	// Set default base URL
	if config.BaseURL == "" {
		config.BaseURL = fmt.Sprintf("http://%s", config.Host)
	}

	timeout := time.Duration(config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 60 * time.Second // Local models may be slower
	}

	return &OllamaAdapter{
		config: config,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Chat sends a prompt to Ollama and returns the response.
func (a *OllamaAdapter) Chat(ctx context.Context, prompt contracts.Prompt) (contracts.Response, error) {
	// Build the request body in Ollama's format
	requestBody := map[string]interface{}{
		"model":  a.config.Model,
		"stream": false,
		"options": map[string]interface{}{
			"num_predict": a.config.MaxTokens,
			"temperature": a.config.Temperature,
		},
	}

	// Build messages
	messages := buildOllamaMessages(prompt)
	if len(messages) > 0 {
		requestBody["messages"] = messages
	} else {
		// Fallback to prompt format for compatibility
		requestBody["prompt"] = prompt.UserMessage
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return contracts.Response{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/chat", a.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return contracts.Response{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := a.client.Do(req)
	if err != nil {
		return contracts.Response{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return contracts.Response{}, fmt.Errorf("Ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var ollamaResponse struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		DoneReason string `json:"done_reason"`
		EvalCount  int `json:"eval_count"`
		PromptEvalCount int `json:"prompt_eval_count"`
		TotalEvalCount int `json:"total_eval_count"`
		Model      string `json:"model"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ollamaResponse); err != nil {
		return contracts.Response{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return contracts.Response{
		Text:         ollamaResponse.Message.Content,
		TokenCount:   ollamaResponse.TotalEvalCount,
		Model:        ollamaResponse.Model,
		FinishReason: ollamaResponse.DoneReason,
		Metadata: map[string]interface{}{
			"prompt_eval_count": ollamaResponse.PromptEvalCount,
			"eval_count":        ollamaResponse.EvalCount,
			"done_reason":       ollamaResponse.DoneReason,
		},
	}, nil
}

// Stream sends a prompt to Ollama and returns a stream of tokens.
func (a *OllamaAdapter) Stream(ctx context.Context, prompt contracts.Prompt) (<-chan string, error) {
	ch := make(chan string)

	// Build the request body with streaming enabled
	requestBody := map[string]interface{}{
		"model":  a.config.Model,
		"stream": true,
		"options": map[string]interface{}{
			"num_predict": a.config.MaxTokens,
			"temperature": a.config.Temperature,
		},
	}

	messages := buildOllamaMessages(prompt)
	if len(messages) > 0 {
		requestBody["messages"] = messages
	} else {
		requestBody["prompt"] = prompt.UserMessage
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		close(ch)
		return ch, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/chat", a.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		close(ch)
		return ch, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

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
		return ch, fmt.Errorf("Ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Stream responses
	go func() {
		defer close(ch)
		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)
		for {
			var chunk struct {
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				Done bool `json:"done"`
			}

			if err := decoder.Decode(&chunk); err != nil {
				if err == io.EOF {
					break
				}
				ch <- fmt.Sprintf("Error: %v", err)
				return
			}

			if chunk.Message.Content != "" {
				ch <- chunk.Message.Content
			}

			if chunk.Done {
				break
			}
		}
	}()

	return ch, nil
}

// ProviderName returns the name of the provider.
func (a *OllamaAdapter) ProviderName() string {
	return "ollama"
}

// AvailableModels returns the list of available Ollama models.
func (a *OllamaAdapter) AvailableModels() []string {
	// Query Ollama's /api/tags endpoint to get available models
	url := fmt.Sprintf("%s/api/tags", a.config.BaseURL)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		// Return default models if we can't query
		return getDefaultOllamaModels()
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return getDefaultOllamaModels()
	}

	var tagsResponse struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tagsResponse); err != nil {
		return getDefaultOllamaModels()
	}

	models := make([]string, 0, len(tagsResponse.Models))
	for _, model := range tagsResponse.Models {
		models = append(models, model.Name)
	}

	return models
}

// buildOllamaMessages converts a Prompt to Ollama's message format.
func buildOllamaMessages(prompt contracts.Prompt) []map[string]string {
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

// getDefaultOllamaModels returns a list of common Ollama models.
func getDefaultOllamaModels() []string {
	return []string{
		"llama2:13b",
		"llama2:7b",
		"codellama:13b",
		"codellama:7b",
		"mistral:7b",
		"mixtral:8x7b",
		"phi:2.7b",
		"gemma:2b",
	}
}
