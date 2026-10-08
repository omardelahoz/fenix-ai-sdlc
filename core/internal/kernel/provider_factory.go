package kernel

import (
	"fmt"

	"github.com/omardelahoz/fenix-ai-sdlc/core/pkg/contracts"
)

// ProviderFactory creates AIProvider instances based on configuration.
type ProviderFactory struct{}

// NewProviderFactory creates a new ProviderFactory.
func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{}
}

// CreateProvider creates an AIProvider based on the provider name and configuration.
func (f *ProviderFactory) CreateProvider(providerName string, config interface{}) (contracts.AIProvider, error) {
	switch providerName {
	case "claude":
		claudeConfig, ok := config.(*contracts.ClaudeConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config type for claude provider")
		}
		return NewClaudeAdapter(claudeConfig), nil

	case "antigravity":
		antigravityConfig, ok := config.(*contracts.AntigravityConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config type for antigravity provider")
		}
		return NewAntigravityAdapter(antigravityConfig), nil

	case "ollama":
		ollamaConfig, ok := config.(*contracts.OllamaConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config type for ollama provider")
		}
		return NewOllamaAdapter(ollamaConfig), nil

	case "openai":
		// TODO: Implement OpenAI adapter
		return nil, fmt.Errorf("openai provider not yet implemented")

	case "gemini":
		// TODO: Implement Gemini adapter
		return nil, fmt.Errorf("gemini provider not yet implemented")

	default:
		return nil, fmt.Errorf("unknown provider: %s", providerName)
	}
}

// CreateProviderFromExecutionPackage creates a provider based on ExecutionPackage resources.
func (f *ProviderFactory) CreateProviderFromExecutionPackage(pkg *contracts.ExecutionPackage, configs map[string]interface{}) (contracts.AIProvider, error) {
	providerName := pkg.Resources.Provider
	
	// Get the appropriate config for this provider
	config, exists := configs[providerName]
	if !exists {
		return nil, fmt.Errorf("no configuration found for provider: %s", providerName)
	}

	return f.CreateProvider(providerName, config)
}

// ListAvailableProviders returns a list of all available providers.
func (f *ProviderFactory) ListAvailableProviders() []string {
	return []string{
		"claude",
		"antigravity",
		"ollama",
		// "openai", // Uncomment when implemented
		// "gemini", // Uncomment when implemented
	}
}
