package provider

import (
	"context"
)

type Prompt struct {
	SystemMessage string
	UserMessage   string
}

type Response struct {
	Text string
	// Additional metadata like TokenCount can be added here
}

// AIProvider abstracts the underlying LLM (Claude, Gemini, Antigravity, etc.)
type AIProvider interface {
	Chat(ctx context.Context, prompt Prompt) (Response, error)
	Stream(ctx context.Context, prompt Prompt) (<-chan string, error)
}

