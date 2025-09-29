package table

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// AIFieldHandler handles AI-powered field operations
type AIFieldHandler struct {
	BaseFieldHandler
	aiProvider AIProvider
}

// AIProvider interface for different AI service providers
type AIProvider interface {
	// Generate generates content based on prompt
	Generate(ctx context.Context, prompt string, options AIGenerateOptions) (string, error)
	
	// Classify classifies content into categories
	Classify(ctx context.Context, content string, categories []string, options AIClassifyOptions) (string, error)
	
	// Extract extracts information from content
	Extract(ctx context.Context, content string, schema interface{}, options AIExtractOptions) (interface{}, error)
	
	// Summarize creates a summary of content
	Summarize(ctx context.Context, content string, options AISummarizeOptions) (string, error)
	
	// Translate translates content to target language
	Translate(ctx context.Context, content string, targetLang string, options AITranslateOptions) (string, error)
}

// AI operation options
type AIGenerateOptions struct {
	Model       string
	MaxTokens   int
	Temperature float32
	TopP        float32
	Stop        []string
}

type AIClassifyOptions struct {
	Model       string
	Temperature float32
}

type AIExtractOptions struct {
	Model string
}

type AISummarizeOptions struct {
	Model     string
	MaxLength int
}

type AITranslateOptions struct {
	Model      string
	SourceLang string
}

// NewAIFieldHandler creates a new AI field handler
func NewAIFieldHandler(provider AIProvider) FieldTypeHandler {
	return &AIFieldHandler{
		BaseFieldHandler: BaseFieldHandler{
			fieldType: FieldTypeVirtualAI,
			info: FieldTypeInfo{
				Type:        FieldTypeVirtualAI,
				Name:        "AI智能字段",
				Description: "使用AI生成、提取或处理内容",
				Category:    "虚拟字段",
				Icon:        "ai",
				Color:       "#00BCD4",
			},
		},
		aiProvider: provider,
	}
}

// ValidateOptions validates AI field options
func (h *AIFieldHandler) ValidateOptions(options *FieldOptions) error {
	if options == nil {
		return fmt.Errorf("AI field requires options")
	}

	aiOpts, err := ParseAIOptions(options)
	if err != nil {
		return err
	}

	// Validate operation type
	validOps := []string{"generate", "extract", "classify", "summarize", "translate"}
	if !contains(validOps, aiOpts.OperationType) {
		return fmt.Errorf("invalid operation type: %s", aiOpts.OperationType)
	}

	// Validate provider
	if aiOpts.Provider == "" {
		return fmt.Errorf("AI provider is required")
	}

	// Validate prompt template
	if aiOpts.PromptTemplate == "" {
		return fmt.Errorf("prompt template is required")
	}

	// Validate source fields
	if len(aiOpts.SourceFields) == 0 {
		return fmt.Errorf("at least one source field is required")
	}

	return nil
}

// ValidateValue validates an AI field value (always returns nil as values are computed)
func (h *AIFieldHandler) ValidateValue(value interface{}, options *FieldOptions) error {
	// AI fields are read-only, values are always computed
	return nil
}

// FormatValue formats an AI field value
func (h *AIFieldHandler) FormatValue(value interface{}, options *FieldOptions) (string, error) {
	if value == nil {
		return "", nil
	}

	aiOpts, _ := ParseAIOptions(options)
	if aiOpts != nil && aiOpts.OutputFormat == "json" {
		// Pretty print JSON
		jsonBytes, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			return fmt.Sprintf("%v", value), nil
		}
		return string(jsonBytes), nil
	}

	return fmt.Sprintf("%v", value), nil
}

// ParseValue parses an AI field value (not applicable for AI fields)
func (h *AIFieldHandler) ParseValue(str string, options *FieldOptions) (interface{}, error) {
	// AI fields are read-only
	return nil, fmt.Errorf("AI fields are read-only")
}

// GetDefaultValue returns the default value for an AI field
func (h *AIFieldHandler) GetDefaultValue(options *FieldOptions) interface{} {
	return nil // AI fields have no default value
}

// IsCompatibleWith checks if AI field is compatible with another type
func (h *AIFieldHandler) IsCompatibleWith(targetType FieldType) bool {
	// AI fields can convert to text
	return targetType == FieldTypeText
}

// ConvertValue converts an AI value to another type
func (h *AIFieldHandler) ConvertValue(value interface{}, targetType FieldType, targetOptions *FieldOptions) (interface{}, error) {
	if targetType == FieldTypeText {
		return fmt.Sprintf("%v", value), nil
	}
	return nil, fmt.Errorf("AI fields can only be converted to text")
}

// Calculate computes the AI field value
func (h *AIFieldHandler) Calculate(ctx CalculationContext) (interface{}, error) {
	if h.aiProvider == nil {
		return nil, fmt.Errorf("AI provider not configured")
	}

	options := ctx.Field.Options
	if options == nil {
		return nil, fmt.Errorf("AI field requires options")
	}

	aiOpts, err := ParseAIOptions(options)
	if err != nil {
		return nil, err
	}

	// Check cache if enabled
	if aiOpts.CacheResults {
		if cachedValue := h.getCachedValue(ctx); cachedValue != nil {
			return cachedValue, nil
		}
	}

	// Prepare prompt with field values
	prompt, err := h.preparePrompt(aiOpts.PromptTemplate, ctx.RecordData, aiOpts.SourceFields)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare prompt: %w", err)
	}

	// Execute AI operation
	var result interface{}
	switch aiOpts.OperationType {
	case "generate":
		result, err = h.generate(ctx.Context, prompt, aiOpts)
	case "classify":
		result, err = h.classify(ctx.Context, prompt, aiOpts)
	case "extract":
		result, err = h.extract(ctx.Context, prompt, aiOpts)
	case "summarize":
		result, err = h.summarize(ctx.Context, prompt, aiOpts)
	case "translate":
		result, err = h.translate(ctx.Context, prompt, aiOpts)
	default:
		return nil, fmt.Errorf("unsupported operation type: %s", aiOpts.OperationType)
	}

	if err != nil {
		return nil, err
	}

	// Cache result if enabled
	if aiOpts.CacheResults && result != nil {
		h.setCachedValue(ctx, result)
	}

	return result, nil
}

// preparePrompt replaces field references in the prompt template
func (h *AIFieldHandler) preparePrompt(template string, recordData map[string]interface{}, sourceFields []string) (string, error) {
	prompt := template

	// Replace field references
	for _, fieldName := range sourceFields {
		placeholder := fmt.Sprintf("{%s}", fieldName)
		if value, exists := recordData[fieldName]; exists {
			valueStr := fmt.Sprintf("%v", value)
			prompt = strings.ReplaceAll(prompt, placeholder, valueStr)
		} else {
			prompt = strings.ReplaceAll(prompt, placeholder, "")
		}
	}

	return prompt, nil
}

// AI operation implementations

func (h *AIFieldHandler) generate(ctx context.Context, prompt string, opts *AIFieldOptions) (interface{}, error) {
	genOpts := AIGenerateOptions{
		Model:       opts.Model,
		MaxTokens:   opts.MaxTokens,
		Temperature: opts.Temperature,
	}

	result, err := h.aiProvider.Generate(ctx, prompt, genOpts)
	if err != nil {
		return nil, err
	}

	return h.formatOutput(result, opts.OutputFormat)
}

func (h *AIFieldHandler) classify(ctx context.Context, content string, opts *AIFieldOptions) (interface{}, error) {
	// Extract categories from prompt or options
	categories := h.extractCategories(opts)
	
	classOpts := AIClassifyOptions{
		Model:       opts.Model,
		Temperature: opts.Temperature,
	}

	result, err := h.aiProvider.Classify(ctx, content, categories, classOpts)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (h *AIFieldHandler) extract(ctx context.Context, content string, opts *AIFieldOptions) (interface{}, error) {
	extOpts := AIExtractOptions{
		Model: opts.Model,
	}

	// Define extraction schema based on output format
	var schema interface{}
	if opts.OutputFormat == "json" {
		// Use a generic schema for now
		schema = map[string]interface{}{
			"type": "object",
		}
	}

	result, err := h.aiProvider.Extract(ctx, content, schema, extOpts)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (h *AIFieldHandler) summarize(ctx context.Context, content string, opts *AIFieldOptions) (interface{}, error) {
	sumOpts := AISummarizeOptions{
		Model:     opts.Model,
		MaxLength: opts.MaxTokens,
	}

	result, err := h.aiProvider.Summarize(ctx, content, sumOpts)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (h *AIFieldHandler) translate(ctx context.Context, content string, opts *AIFieldOptions) (interface{}, error) {
	// Extract target language from prompt or options
	targetLang := h.extractTargetLanguage(opts)
	
	transOpts := AITranslateOptions{
		Model: opts.Model,
	}

	result, err := h.aiProvider.Translate(ctx, content, targetLang, transOpts)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Helper methods

func (h *AIFieldHandler) formatOutput(result string, format string) (interface{}, error) {
	switch format {
	case "json":
		var jsonResult interface{}
		if err := json.Unmarshal([]byte(result), &jsonResult); err != nil {
			// If not valid JSON, return as string
			return result, nil
		}
		return jsonResult, nil
	case "markdown":
		// TODO: Add markdown processing if needed
		return result, nil
	default:
		return result, nil
	}
}

func (h *AIFieldHandler) extractCategories(opts *AIFieldOptions) []string {
	// TODO: Extract categories from prompt template or provider options
	if categories, ok := opts.ProviderOptions["categories"].([]string); ok {
		return categories
	}
	return []string{}
}

func (h *AIFieldHandler) extractTargetLanguage(opts *AIFieldOptions) string {
	// TODO: Extract target language from prompt template or provider options
	if lang, ok := opts.ProviderOptions["target_language"].(string); ok {
		return lang
	}
	return "en"
}

func (h *AIFieldHandler) getCachedValue(ctx CalculationContext) interface{} {
	// TODO: Implement caching mechanism
	return nil
}

func (h *AIFieldHandler) setCachedValue(ctx CalculationContext, value interface{}) {
	// TODO: Implement caching mechanism
}

// ParseAIOptions parses AI field options
func ParseAIOptions(options *FieldOptions) (*AIFieldOptions, error) {
	if options == nil {
		return nil, fmt.Errorf("options is nil")
	}

	// TODO: Parse from actual options structure
	// For now, create default options
	aiOpts := &AIFieldOptions{
		OperationType:  "generate",
		Provider:       "openai",
		Model:          "gpt-3.5-turbo",
		PromptTemplate: options.Formula, // Use formula field for prompt
		SourceFields:   []string{},
		OutputFormat:   "text",
		CacheResults:   true,
		MaxTokens:      500,
		Temperature:    0.7,
	}

	// Extract source fields from prompt template
	// Look for {field_name} patterns
	prompt := aiOpts.PromptTemplate
	start := 0
	for {
		idx := strings.Index(prompt[start:], "{")
		if idx == -1 {
			break
		}
		idx += start
		
		endIdx := strings.Index(prompt[idx:], "}")
		if endIdx == -1 {
			break
		}
		endIdx += idx
		
		fieldName := prompt[idx+1 : endIdx]
		if fieldName != "" && !contains(aiOpts.SourceFields, fieldName) {
			aiOpts.SourceFields = append(aiOpts.SourceFields, fieldName)
		}
		
		start = endIdx + 1
	}

	return aiOpts, nil
}

// contains checks if a string slice contains a value
func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}