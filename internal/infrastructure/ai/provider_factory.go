package ai

import (
	"fmt"

	"teable-go-backend/internal/config"
	"teable-go-backend/internal/domain/table"
)

// ProviderFactory creates AI providers based on configuration
type ProviderFactory struct {
	config config.AIConfig
}

// NewProviderFactory creates a new provider factory
func NewProviderFactory(cfg config.AIConfig) *ProviderFactory {
	return &ProviderFactory{
		config: cfg,
	}
}

// CreateProvider creates an AI provider based on the provider name
func (f *ProviderFactory) CreateProvider(providerName string) (table.AIProvider, error) {
	// Use default provider if not specified
	if providerName == "" {
		providerName = f.config.DefaultProvider
	}

	providerConfig, exists := f.config.Providers[providerName]
	if !exists {
		return nil, fmt.Errorf("provider %s not configured", providerName)
	}

	switch providerConfig.Type {
	case "openai":
		return f.createOpenAIProvider(providerConfig)
	case "deepseek":
		return f.createDeepSeekProvider(providerConfig)
	case "anthropic":
		return f.createAnthropicProvider(providerConfig)
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerConfig.Type)
	}
}

// CreateDefaultProvider creates the default AI provider
func (f *ProviderFactory) CreateDefaultProvider() (table.AIProvider, error) {
	return f.CreateProvider(f.config.DefaultProvider)
}

func (f *ProviderFactory) createOpenAIProvider(cfg config.AIProviderConfig) (table.AIProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key not configured")
	}

	provider := NewOpenAIProvider(cfg.APIKey)

	// Set custom base URL if provided
	if cfg.BaseURL != "" {
		provider.baseURL = cfg.BaseURL
	}

	return provider, nil
}

func (f *ProviderFactory) createDeepSeekProvider(cfg config.AIProviderConfig) (table.AIProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("DeepSeek API key not configured")
	}

	// DeepSeek uses OpenAI-compatible API
	provider := NewOpenAIProvider(cfg.APIKey)

	// Set DeepSeek base URL
	if cfg.BaseURL != "" {
		provider.baseURL = cfg.BaseURL
	} else {
		provider.baseURL = "https://api.deepseek.com/v1"
	}

	return provider, nil
}

func (f *ProviderFactory) createAnthropicProvider(cfg config.AIProviderConfig) (table.AIProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("Anthropic API key not configured")
	}

	// TODO: Implement Anthropic provider
	return nil, fmt.Errorf("Anthropic provider not yet implemented")
}
