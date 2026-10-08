package kernel

import (
	"context"
	"fmt"
	"log"

	"github.com/omardelahoz/fenix-ai-sdlc/core/pkg/contracts"
)

// ExampleUsage demonstrates how to use different AI providers.
func ExampleUsage() {
	// Create the provider factory
	factory := NewProviderFactory()

	// Configure providers
	configs := map[string]interface{}{
		"claude": &contracts.ClaudeConfig{
			ProviderConfig: contracts.ProviderConfig{
				APIKey:      "sk-ant-xxx",
				Model:       "claude-3-5-sonnet-20241022",
				MaxTokens:   100000,
				Temperature: 0.7,
				Timeout:     30,
			},
			Region: "us-east-1",
		},
		"antigravity": &contracts.AntigravityConfig{
			ProviderConfig: contracts.ProviderConfig{
				APIKey:      "ag-xxx",
				BaseURL:     "https://api.antigravity.com",
				Model:       "antigravity-coding-v1",
				MaxTokens:   100000,
				Temperature: 0.7,
				Timeout:     30,
			},
			WorkspaceID: "ws-123",
			ProjectID:   "proj-456",
		},
		"ollama": &contracts.OllamaConfig{
			ProviderConfig: contracts.ProviderConfig{
				BaseURL:     "http://localhost:11434",
				Model:       "codellama:13b",
				MaxTokens:   8192,
				Temperature: 0.7,
				Timeout:     60,
			},
			Host: "localhost:11434",
		},
	}

	// List available providers
	fmt.Println("Available providers:")
	for _, provider := range factory.ListAvailableProviders() {
		fmt.Printf("  - %s\n", provider)
	}

	// Example 1: Use Claude for architecture
	fmt.Println("\n=== Using Claude for Architecture ===")
	claudeProvider, err := factory.CreateProvider("claude", configs["claude"])
	if err != nil {
		log.Fatalf("Failed to create Claude provider: %v", err)
	}

	architecturePrompt := contracts.Prompt{
		SystemMessage: "You are a senior software architect.",
		UserMessage:   "Design a microservices architecture for an e-commerce platform.",
	}

	response, err := claudeProvider.Chat(context.Background(), architecturePrompt)
	if err != nil {
		log.Printf("Claude chat failed: %v", err)
	} else {
		fmt.Printf("Claude response (%d tokens): %s\n", response.TokenCount, response.Text[:min(200, len(response.Text))])
	}

	// Example 2: Use Antigravity for coding
	fmt.Println("\n=== Using Antigravity for Coding ===")
	antigravityProvider, err := factory.CreateProvider("antigravity", configs["antigravity"])
	if err != nil {
		log.Fatalf("Failed to create Antigravity provider: %v", err)
	}

	codingPrompt := contracts.Prompt{
		SystemMessage: "You are a Go developer.",
		UserMessage:   "Write a REST API handler for user authentication.",
	}

	response, err = antigravityProvider.Chat(context.Background(), codingPrompt)
	if err != nil {
		log.Printf("Antigravity chat failed: %v", err)
	} else {
		fmt.Printf("Antigravity response (%d tokens): %s\n", response.TokenCount, response.Text[:min(200, len(response.Text))])
	}

	// Example 3: Use Ollama for quick local testing
	fmt.Println("\n=== Using Ollama for Local Testing ===")
	ollamaProvider, err := factory.CreateProvider("ollama", configs["ollama"])
	if err != nil {
		log.Fatalf("Failed to create Ollama provider: %v", err)
	}

	// List available models from Ollama
	fmt.Println("Available Ollama models:")
	for _, model := range ollamaProvider.AvailableModels() {
		fmt.Printf("  - %s\n", model)
	}

	quickPrompt := contracts.Prompt{
		UserMessage: "What is 2 + 2?",
	}

	response, err = ollamaProvider.Chat(context.Background(), quickPrompt)
	if err != nil {
		log.Printf("Ollama chat failed: %v", err)
	} else {
		fmt.Printf("Ollama response: %s\n", response.Text)
	}

	// Example 4: Streaming
	fmt.Println("\n=== Streaming with Ollama ===")
	stream, err := ollamaProvider.Stream(context.Background(), contracts.Prompt{
		UserMessage: "Explain recursion in programming.",
	})
	if err != nil {
		log.Fatalf("Failed to start stream: %v", err)
	}

	fmt.Print("Streamed response: ")
	for chunk := range stream {
		fmt.Print(chunk)
	}
	fmt.Println()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ExampleDynamicProviderSelection shows how to select providers dynamically.
func ExampleDynamicProviderSelection() {
	factory := NewProviderFactory()

	// Strategy 1: Provider based on task type
	taskType := "architecture"
	var providerName string

	switch taskType {
	case "architecture":
		providerName = "claude" // Best for reasoning
	case "coding":
		providerName = "antigravity" // Specialized in code
	case "quick_test":
		providerName = "ollama" // Free and local
	default:
		providerName = "claude" // Default
	}

	fmt.Printf("Selected provider for task '%s': %s\n", taskType, providerName)

	// Strategy 2: Fallback chain
	providers := []string{"claude", "antigravity", "ollama"}
	var selectedProvider contracts.AIProvider
	var err error

	for _, provider := range providers {
		config := getDefaultConfig(provider)
		selectedProvider, err = factory.CreateProvider(provider, config)
		if err == nil {
			fmt.Printf("Successfully connected to: %s\n", provider)
			break
		}
		fmt.Printf("Failed to connect to %s: %v, trying next...\n", provider, err)
	}

	if selectedProvider == nil {
		log.Fatal("All providers failed")
	}

	// Use the selected provider
	_ = selectedProvider
}

func getDefaultConfig(provider string) interface{} {
	switch provider {
	case "claude":
		return &contracts.ClaudeConfig{
			ProviderConfig: contracts.ProviderConfig{
				APIKey: "sk-ant-xxx",
				Model:  "claude-3-5-sonnet-20241022",
			},
		}
	case "antigravity":
		return &contracts.AntigravityConfig{
			ProviderConfig: contracts.ProviderConfig{
				APIKey:  "ag-xxx",
				BaseURL: "https://api.antigravity.com",
				Model:   "antigravity-coding-v1",
			},
		}
	case "ollama":
		return &contracts.OllamaConfig{
			ProviderConfig: contracts.ProviderConfig{
				BaseURL: "http://localhost:11434",
				Model:   "codellama:13b",
			},
		}
	default:
		return nil
	}
}
