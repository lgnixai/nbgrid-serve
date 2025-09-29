package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/yourusername/teable-app/internal/domain/table"
)

// OpenAIProvider implements AI provider using OpenAI API
type OpenAIProvider struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey string) *OpenAIProvider {
	return &OpenAIProvider{
		apiKey:  apiKey,
		baseURL: "https://api.openai.com/v1",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Generate generates content based on prompt
func (p *OpenAIProvider) Generate(ctx context.Context, prompt string, options table.AIGenerateOptions) (string, error) {
	model := options.Model
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	requestBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"temperature": options.Temperature,
		"max_tokens":  options.MaxTokens,
	}

	if options.TopP > 0 {
		requestBody["top_p"] = options.TopP
	}
	if len(options.Stop) > 0 {
		requestBody["stop"] = options.Stop
	}

	response, err := p.makeRequest(ctx, "/chat/completions", requestBody)
	if err != nil {
		return "", err
	}

	// Extract content from response
	choices, ok := response["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("invalid response format")
	}

	firstChoice := choices[0].(map[string]interface{})
	message := firstChoice["message"].(map[string]interface{})
	content := message["content"].(string)

	return content, nil
}

// Classify classifies content into categories
func (p *OpenAIProvider) Classify(ctx context.Context, content string, categories []string, options table.AIClassifyOptions) (string, error) {
	model := options.Model
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	// Build classification prompt
	categoriesStr := ""
	for i, cat := range categories {
		if i > 0 {
			categoriesStr += ", "
		}
		categoriesStr += cat
	}

	prompt := fmt.Sprintf(`Classify the following text into one of these categories: %s

Text: %s

Category:`, categoriesStr, content)

	requestBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{
				"role": "system",
				"content": "You are a text classifier. Respond with only the category name, nothing else.",
			},
			{"role": "user", "content": prompt},
		},
		"temperature": options.Temperature,
		"max_tokens":  50,
	}

	response, err := p.makeRequest(ctx, "/chat/completions", requestBody)
	if err != nil {
		return "", err
	}

	// Extract content from response
	choices, ok := response["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("invalid response format")
	}

	firstChoice := choices[0].(map[string]interface{})
	message := firstChoice["message"].(map[string]interface{})
	category := message["content"].(string)

	return category, nil
}

// Extract extracts information from content
func (p *OpenAIProvider) Extract(ctx context.Context, content string, schema interface{}, options table.AIExtractOptions) (interface{}, error) {
	model := options.Model
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	// Build extraction prompt
	prompt := fmt.Sprintf(`Extract structured information from the following text and return it as JSON:

Text: %s

Return the extracted information as a valid JSON object.`, content)

	requestBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{
				"role": "system",
				"content": "You are a data extraction assistant. Always respond with valid JSON only, no explanations.",
			},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.1, // Low temperature for consistent extraction
		"max_tokens":  1000,
	}

	response, err := p.makeRequest(ctx, "/chat/completions", requestBody)
	if err != nil {
		return nil, err
	}

	// Extract content from response
	choices, ok := response["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return nil, fmt.Errorf("invalid response format")
	}

	firstChoice := choices[0].(map[string]interface{})
	message := firstChoice["message"].(map[string]interface{})
	jsonStr := message["content"].(string)

	// Parse JSON response
	var result interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		// If parsing fails, return as string
		return jsonStr, nil
	}

	return result, nil
}

// Summarize creates a summary of content
func (p *OpenAIProvider) Summarize(ctx context.Context, content string, options table.AISummarizeOptions) (string, error) {
	model := options.Model
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	maxLength := options.MaxLength
	if maxLength == 0 {
		maxLength = 200
	}

	prompt := fmt.Sprintf(`Summarize the following text in no more than %d words:

%s`, maxLength/5, content) // Approximate words from tokens

	requestBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{
				"role": "system",
				"content": "You are a summarization assistant. Provide concise, informative summaries.",
			},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.3,
		"max_tokens":  maxLength,
	}

	response, err := p.makeRequest(ctx, "/chat/completions", requestBody)
	if err != nil {
		return "", err
	}

	// Extract content from response
	choices, ok := response["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("invalid response format")
	}

	firstChoice := choices[0].(map[string]interface{})
	message := firstChoice["message"].(map[string]interface{})
	summary := message["content"].(string)

	return summary, nil
}

// Translate translates content to target language
func (p *OpenAIProvider) Translate(ctx context.Context, content string, targetLang string, options table.AITranslateOptions) (string, error) {
	model := options.Model
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	sourceLang := options.SourceLang
	if sourceLang == "" {
		sourceLang = "auto-detect"
	}

	prompt := fmt.Sprintf(`Translate the following text to %s:

%s`, targetLang, content)

	if sourceLang != "auto-detect" {
		prompt = fmt.Sprintf(`Translate the following text from %s to %s:

%s`, sourceLang, targetLang, content)
	}

	requestBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{
				"role": "system",
				"content": "You are a professional translator. Provide accurate translations preserving the original meaning and tone.",
			},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.3,
		"max_tokens":  2000,
	}

	response, err := p.makeRequest(ctx, "/chat/completions", requestBody)
	if err != nil {
		return "", err
	}

	// Extract content from response
	choices, ok := response["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("invalid response format")
	}

	firstChoice := choices[0].(map[string]interface{})
	message := firstChoice["message"].(map[string]interface{})
	translation := message["content"].(string)

	return translation, nil
}

// makeRequest makes an HTTP request to OpenAI API
func (p *OpenAIProvider) makeRequest(ctx context.Context, endpoint string, body interface{}) (map[string]interface{}, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result, nil
}