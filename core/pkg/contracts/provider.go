package contracts

import "context"

// Prompt represents a prompt to be sent to an AI provider.
type Prompt struct {
	SystemMessage string
	UserMessage   string
	Messages      []Message // For chat history
}

// Message represents a single message in a conversation.
type Message struct {
	Role    string // "system", "user", "assistant"
	Content string
}

// Response represents the response from an AI provider.
type Response struct {
	Text         string
	FinishReason string
	TokenCount   int
	Model        string
	Metadata     map[string]interface{}
}

// AIProvider abstracts the underlying LLM (Claude, Gemini, Antigravity, etc.).
type AIProvider interface {
	// Chat sends a prompt and returns a response.
	Chat(ctx context.Context, prompt Prompt) (Response, error)

	// Stream sends a prompt and returns a stream of tokens.
	Stream(ctx context.Context, prompt Prompt) (<-chan string, error)

	// ProviderName returns the name of the provider.
	ProviderName() string

	// AvailableModels returns the list of available models.
	AvailableModels() []string
}

// ProviderConfig contains configuration for an AI provider.
type ProviderConfig struct {
	APIKey      string
	BaseURL     string
	Model       string
	MaxTokens   int
	Temperature float64
	Timeout     int // seconds
}

// ClaudeConfig contains Claude-specific configuration.
type ClaudeConfig struct {
	ProviderConfig
	Region string // e.g., "us-east-1"
}

// AntigravityConfig contains Antigravity-specific configuration.
type AntigravityConfig struct {
	ProviderConfig
	WorkspaceID string
	ProjectID   string
}

// OllamaConfig contains Ollama-specific configuration for local LLMs.
type OllamaConfig struct {
	ProviderConfig
	Host string // e.g., "localhost:11434"
}
