package kernel

import (
	"context"
	"fmt"
	"time"

	"github.com/omardelahoz/fenix-ai-sdlc/core/pkg/contracts"
)

// ExampleStreaming demonstrates how to use streaming with different providers.
func ExampleStreaming() {
	ctx := context.Background()

	// Example 1: Streaming with Claude
	fmt.Println("=== Streaming with Claude ===")
	claudeConfig := &contracts.ClaudeConfig{
		ProviderConfig: contracts.ProviderConfig{
			APIKey:      "sk-ant-xxx", // Replace with actual API key
			Model:       "claude-3-5-sonnet-20241022",
			MaxTokens:   1000,
			Temperature: 0.7,
			Timeout:     30,
		},
		Region: "us-east-1",
	}

	claudeAdapter := NewClaudeAdapter(claudeConfig)
	claudePrompt := contracts.Prompt{
		SystemMessage: "You are a helpful assistant.",
		UserMessage:   "Explain what is Fénix AI SDLC in one sentence.",
	}

	claudeStream, err := claudeAdapter.Stream(ctx, claudePrompt)
	if err != nil {
		fmt.Printf("Claude stream error: %v\n", err)
	} else {
		fmt.Print("Claude response: ")
		for chunk := range claudeStream {
			fmt.Print(chunk)
		}
		fmt.Println()
	}

	// Example 2: Streaming with Antigravity
	fmt.Println("\n=== Streaming with Antigravity ===")
	antigravityConfig := &contracts.AntigravityConfig{
		ProviderConfig: contracts.ProviderConfig{
			APIKey:      "ag-xxx", // Replace with actual API key
			BaseURL:     "https://api.antigravity.com",
			Model:       "antigravity-coding-v1",
			MaxTokens:   1000,
			Temperature: 0.7,
			Timeout:     30,
		},
		WorkspaceID: "ws-123",
		ProjectID:   "proj-456",
	}

	antigravityAdapter := NewAntigravityAdapter(antigravityConfig)
	antigravityPrompt := contracts.Prompt{
		SystemMessage: "You are a software engineer.",
		UserMessage:   "Write a simple Go function that adds two numbers.",
	}

	antigravityStream, err := antigravityAdapter.Stream(ctx, antigravityPrompt)
	if err != nil {
		fmt.Printf("Antigravity stream error: %v\n", err)
	} else {
		fmt.Print("Antigravity response: ")
		for chunk := range antigravityStream {
			fmt.Print(chunk)
		}
		fmt.Println()
	}

	// Example 3: Streaming with Ollama (local)
	fmt.Println("\n=== Streaming with Ollama (Local) ===")
	ollamaConfig := &contracts.OllamaConfig{
		ProviderConfig: contracts.ProviderConfig{
			BaseURL:     "http://localhost:11434",
			Model:       "codellama:13b",
			MaxTokens:   500,
			Temperature: 0.7,
			Timeout:     60,
		},
		Host: "localhost:11434",
	}

	ollamaAdapter := NewOllamaAdapter(ollamaConfig)
	ollamaPrompt := contracts.Prompt{
		UserMessage: "What is 2 + 2?",
	}

	ollamaStream, err := ollamaAdapter.Stream(ctx, ollamaPrompt)
	if err != nil {
		fmt.Printf("Ollama stream error: %v\n", err)
	} else {
		fmt.Print("Ollama response: ")
		for chunk := range ollamaStream {
			fmt.Print(chunk)
		}
		fmt.Println()
	}
}

// ExampleStreamingWithCancellation demonstrates how to cancel a streaming request.
func ExampleStreamingWithCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ollamaConfig := &contracts.OllamaConfig{
		ProviderConfig: contracts.ProviderConfig{
			BaseURL: "http://localhost:11434",
			Model:   "codellama:13b",
		},
	}

	ollamaAdapter := NewOllamaAdapter(ollamaConfig)
	stream, err := ollamaAdapter.Stream(ctx, contracts.Prompt{
		UserMessage: "Write a long explanation of Go programming.",
	})
	if err != nil {
		fmt.Printf("Stream error: %v\n", err)
		return
	}

	// Read for 2 seconds then cancel
	go func() {
		time.Sleep(2 * time.Second)
		fmt.Println("\n--- Cancelling stream ---")
		cancel()
	}()

	fmt.Print("Streaming (will cancel after 2s): ")
	for chunk := range stream {
		fmt.Print(chunk)
	}
	fmt.Println("\nStream cancelled")
}

// ExampleStreamingWithBuffer demonstrates collecting stream into a buffer.
func ExampleStreamingWithBuffer() {
	ctx := context.Background()

	ollamaConfig := &contracts.OllamaConfig{
		ProviderConfig: contracts.ProviderConfig{
			BaseURL: "http://localhost:11434",
			Model:   "codellama:13b",
		},
	}

	ollamaAdapter := NewOllamaAdapter(ollamaConfig)
	stream, err := ollamaAdapter.Stream(ctx, contracts.Prompt{
		UserMessage: "What is recursion?",
	})
	if err != nil {
		fmt.Printf("Stream error: %v\n", err)
		return
	}

	// Collect stream into buffer
	var buffer string
	for chunk := range stream {
		buffer += chunk
	}

	fmt.Printf("Complete response: %s\n", buffer)
	fmt.Printf("Response length: %d characters\n", len(buffer))
}

// ExampleStreamingWordByWord demonstrates streaming word by word.
func ExampleStreamingWordByWord() {
	ctx := context.Background()

	ollamaConfig := &contracts.OllamaConfig{
		ProviderConfig: contracts.ProviderConfig{
			BaseURL: "http://localhost:11434",
			Model:   "codellama:13b",
		},
	}

	ollamaAdapter := NewOllamaAdapter(ollamaConfig)
	stream, err := ollamaAdapter.Stream(ctx, contracts.Prompt{
		UserMessage: "Say hello world.",
	})
	if err != nil {
		fmt.Printf("Stream error: %v\n", err)
		return
	}

	// Stream word by word
	var currentWord string
	for chunk := range stream {
		currentWord += chunk
		// Split on spaces and print complete words
		for len(currentWord) > 0 {
			spaceIdx := -1
			for i, c := range currentWord {
				if c == ' ' || c == '\n' {
					spaceIdx = i
					break
				}
			}

			if spaceIdx == -1 {
				// No space yet, keep accumulating
				break
			}

			word := currentWord[:spaceIdx]
			fmt.Printf("Word: %s\n", word)
			currentWord = currentWord[spaceIdx+1:]
		}
	}

	// Print remaining word
	if currentWord != "" {
		fmt.Printf("Word: %s\n", currentWord)
	}
}

// ExampleStreamingComparison compares streaming vs non-streaming latency.
func ExampleStreamingComparison() {
	ctx := context.Background()

	ollamaConfig := &contracts.OllamaConfig{
		ProviderConfig: contracts.ProviderConfig{
			BaseURL: "http://localhost:11434",
			Model:   "codellama:13b",
		},
	}

	ollamaAdapter := NewOllamaAdapter(ollamaConfig)
	prompt := contracts.Prompt{
		UserMessage: "What is 2 + 2?",
	}

	// Non-streaming (chat)
	fmt.Println("=== Non-streaming (Chat) ===")
	start := time.Now()
	response, err := ollamaAdapter.Chat(ctx, prompt)
	if err != nil {
		fmt.Printf("Chat error: %v\n", err)
		return
	}
	elapsed := time.Since(start)
	fmt.Printf("Response: %s\n", response.Text)
	fmt.Printf("Time to complete: %v\n", elapsed)

	// Streaming
	fmt.Println("\n=== Streaming ===")
	start = time.Now()
	stream, err := ollamaAdapter.Stream(ctx, prompt)
	if err != nil {
		fmt.Printf("Stream error: %v\n", err)
		return
	}

	firstChunkTime := time.Time{}
	totalChunks := 0
	for chunk := range stream {
		if firstChunkTime.IsZero() {
			firstChunkTime = time.Now()
			fmt.Printf("First chunk received after: %v\n", firstChunkTime.Sub(start))
		}
		totalChunks++
		fmt.Print(chunk)
	}
	fmt.Println()
	fmt.Printf("Total chunks: %d\n", totalChunks)
	fmt.Printf("Total time: %v\n", time.Since(start))
}
