package table

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Virtual field types that are computed based on other fields
const (
	// FieldTypeVirtualFormula is a computed field based on a formula expression
	FieldTypeVirtualFormula FieldType = "virtual_formula"
	
	// FieldTypeVirtualLookup retrieves values from linked records
	FieldTypeVirtualLookup FieldType = "virtual_lookup"
	
	// FieldTypeVirtualRollup aggregates values from linked records
	FieldTypeVirtualRollup FieldType = "virtual_rollup"
	
	// FieldTypeVirtualAI uses AI to generate or process content
	FieldTypeVirtualAI FieldType = "virtual_ai"
)

// FormulaFieldOptions represents options for formula fields
type FormulaFieldOptions struct {
	// Formula expression to evaluate
	Expression string `json:"expression"`
	
	// Result type of the formula
	ResultType FieldType `json:"result_type"`
	
	// Referenced field IDs used in the formula
	ReferencedFields []string `json:"referenced_fields"`
	
	// Whether to recalculate on every read
	DynamicCalculation bool `json:"dynamic_calculation"`
}

// LookupFieldOptions represents options for lookup fields
type LookupFieldOptions struct {
	// The link field ID to traverse
	LinkFieldID string `json:"link_field_id"`
	
	// The field ID in the linked table to look up
	LookupFieldID string `json:"lookup_field_id"`
	
	// How to handle multiple linked records
	MultipleRecordHandling string `json:"multiple_record_handling"` // first, last, array, comma_separated
}

// RollupFieldOptions represents options for rollup fields
type RollupFieldOptions struct {
	// The link field ID to traverse
	LinkFieldID string `json:"link_field_id"`
	
	// The field ID in the linked table to aggregate
	RollupFieldID string `json:"rollup_field_id"`
	
	// Aggregation function
	AggregationFunction string `json:"aggregation_function"` // sum, avg, count, min, max, unique_count
	
	// Filter condition for records to include
	FilterExpression string `json:"filter_expression,omitempty"`
}

// AIFieldOptions represents options for AI-powered fields
type AIFieldOptions struct {
	// AI operation type
	OperationType string `json:"operation_type"` // generate, extract, classify, summarize, translate
	
	// AI provider to use
	Provider string `json:"provider"` // openai, deepseek, anthropic, etc.
	
	// Model to use
	Model string `json:"model"`
	
	// Prompt template with field references
	PromptTemplate string `json:"prompt_template"`
	
	// Source fields to use as input
	SourceFields []string `json:"source_fields"`
	
	// Output format
	OutputFormat string `json:"output_format"` // text, json, markdown
	
	// Cache results
	CacheResults bool `json:"cache_results"`
	
	// Max tokens for generation
	MaxTokens int `json:"max_tokens,omitempty"`
	
	// Temperature for generation
	Temperature float32 `json:"temperature,omitempty"`
	
	// Additional provider-specific options
	ProviderOptions map[string]interface{} `json:"provider_options,omitempty"`
}

// VirtualFieldMetadata contains metadata for virtual fields
type VirtualFieldMetadata struct {
	// Last calculation time
	LastCalculatedAt *string `json:"last_calculated_at,omitempty"`
	
	// Calculation error if any
	CalculationError *string `json:"calculation_error,omitempty"`
	
	// Cached value
	CachedValue interface{} `json:"cached_value,omitempty"`
	
	// Dependencies that trigger recalculation
	Dependencies []string `json:"dependencies,omitempty"`
}

// IsVirtualField checks if a field type is virtual
func IsVirtualField(fieldType FieldType) bool {
	switch fieldType {
	case FieldTypeVirtualFormula, FieldTypeVirtualLookup, 
	     FieldTypeVirtualRollup, FieldTypeVirtualAI:
		return true
	default:
		return false
	}
}

// ParseVirtualFieldOptions parses options for virtual field types
func ParseVirtualFieldOptions(fieldType FieldType, options *FieldOptions) (interface{}, error) {
	if options == nil {
		return nil, fmt.Errorf("options required for virtual field type %s", fieldType)
	}
	
	// Convert options to JSON for parsing
	optionsJSON, err := json.Marshal(options)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal options: %w", err)
	}
	
	switch fieldType {
	case FieldTypeVirtualFormula:
		var formulaOpts FormulaFieldOptions
		if err := json.Unmarshal(optionsJSON, &formulaOpts); err != nil {
			return nil, fmt.Errorf("invalid formula field options: %w", err)
		}
		return &formulaOpts, nil
		
	case FieldTypeVirtualLookup:
		var lookupOpts LookupFieldOptions
		if err := json.Unmarshal(optionsJSON, &lookupOpts); err != nil {
			return nil, fmt.Errorf("invalid lookup field options: %w", err)
		}
		return &lookupOpts, nil
		
	case FieldTypeVirtualRollup:
		var rollupOpts RollupFieldOptions
		if err := json.Unmarshal(optionsJSON, &rollupOpts); err != nil {
			return nil, fmt.Errorf("invalid rollup field options: %w", err)
		}
		return &rollupOpts, nil
		
	case FieldTypeVirtualAI:
		var aiOpts AIFieldOptions
		if err := json.Unmarshal(optionsJSON, &aiOpts); err != nil {
			return nil, fmt.Errorf("invalid AI field options: %w", err)
		}
		return &aiOpts, nil
		
	default:
		return nil, fmt.Errorf("unknown virtual field type: %s", fieldType)
	}
}

// GetVirtualFieldInfo returns field type info for virtual fields
func GetVirtualFieldInfo(fieldType FieldType) FieldTypeInfo {
	infos := map[FieldType]FieldTypeInfo{
		FieldTypeVirtualFormula: {
			Type:        FieldTypeVirtualFormula,
			Name:        "公式",
			Description: "基于其他字段值计算的虚拟字段",
			Category:    "虚拟字段",
			Icon:        "formula",
			Color:       "#9C27B0",
		},
		FieldTypeVirtualLookup: {
			Type:        FieldTypeVirtualLookup,
			Name:        "查找引用",
			Description: "从关联记录中查找并显示字段值",
			Category:    "虚拟字段",
			Icon:        "lookup",
			Color:       "#673AB7",
		},
		FieldTypeVirtualRollup: {
			Type:        FieldTypeVirtualRollup,
			Name:        "汇总统计",
			Description: "对关联记录的字段值进行聚合计算",
			Category:    "虚拟字段",
			Icon:        "rollup",
			Color:       "#3F51B5",
		},
		FieldTypeVirtualAI: {
			Type:        FieldTypeVirtualAI,
			Name:        "AI智能字段",
			Description: "使用AI生成、提取或处理内容",
			Category:    "虚拟字段",
			Icon:        "ai",
			Color:       "#00BCD4",
		},
	}
	
	if info, exists := infos[fieldType]; exists {
		return info
	}
	
	return FieldTypeInfo{
		Type:        fieldType,
		Name:        "未知虚拟字段",
		Description: "未知的虚拟字段类型",
		Category:    "虚拟字段",
		Icon:        "unknown",
		Color:       "#CCCCCC",
	}
}

// VirtualFieldCalculator interface for calculating virtual field values
type VirtualFieldCalculator interface {
	// Calculate computes the value for a virtual field
	Calculate(ctx CalculationContext) (interface{}, error)
	
	// ValidateOptions validates the field options
	ValidateOptions() error
	
	// GetDependencies returns field IDs this virtual field depends on
	GetDependencies() []string
}

// CalculationContext provides context for virtual field calculation
type CalculationContext struct {
	// Current record data
	RecordData map[string]interface{}
	
	// Table containing the field
	Table *Table
	
	// Field being calculated
	Field *Field
	
	// User context for permissions
	UserID string
	
	// Additional context data
	Context map[string]interface{}
}

// FieldShortcut represents a pre-configured field template
type FieldShortcut struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Category    string       `json:"category"`
	Icon        string       `json:"icon"`
	FieldType   FieldType    `json:"field_type"`
	Options     interface{}  `json:"options"`
	Tags        []string     `json:"tags"`
}

// Common field shortcuts for quick field creation
var FieldShortcuts = []FieldShortcut{
	{
		ID:          "auto_rank",
		Name:        "自动排名",
		Description: "根据指定字段自动计算排名",
		Category:    "计算",
		Icon:        "ranking",
		FieldType:   FieldTypeVirtualFormula,
		Options: FormulaFieldOptions{
			Expression: "RANK({field}, 'desc')",
			ResultType: FieldTypeNumber,
		},
		Tags: []string{"ranking", "calculation"},
	},
	{
		ID:          "ai_summary",
		Name:        "AI摘要",
		Description: "使用AI自动生成内容摘要",
		Category:    "AI",
		Icon:        "ai_summary",
		FieldType:   FieldTypeVirtualAI,
		Options: AIFieldOptions{
			OperationType:  "summarize",
			Provider:       "openai",
			Model:          "gpt-3.5-turbo",
			PromptTemplate: "请总结以下内容的要点：{content}",
			OutputFormat:   "text",
			CacheResults:   true,
		},
		Tags: []string{"ai", "summary", "text"},
	},
	{
		ID:          "ai_classify",
		Name:        "AI分类",
		Description: "使用AI自动对内容进行分类",
		Category:    "AI",
		Icon:        "ai_classify",
		FieldType:   FieldTypeVirtualAI,
		Options: AIFieldOptions{
			OperationType:  "classify",
			Provider:       "openai",
			Model:          "gpt-3.5-turbo",
			PromptTemplate: "请将以下内容分类到合适的类别中：{content}\n可选类别：{categories}",
			OutputFormat:   "text",
			CacheResults:   true,
		},
		Tags: []string{"ai", "classification"},
	},
	{
		ID:          "total_linked",
		Name:        "关联总数",
		Description: "统计关联记录的总数",
		Category:    "统计",
		Icon:        "count",
		FieldType:   FieldTypeVirtualRollup,
		Options: RollupFieldOptions{
			AggregationFunction: "count",
		},
		Tags: []string{"count", "statistics"},
	},
	{
		ID:          "sum_linked",
		Name:        "关联求和",
		Description: "对关联记录的数值字段求和",
		Category:    "统计",
		Icon:        "sum",
		FieldType:   FieldTypeVirtualRollup,
		Options: RollupFieldOptions{
			AggregationFunction: "sum",
		},
		Tags: []string{"sum", "statistics"},
	},
}

// GetFieldShortcutByID returns a field shortcut by ID
func GetFieldShortcutByID(id string) (*FieldShortcut, error) {
	for _, shortcut := range FieldShortcuts {
		if shortcut.ID == id {
			return &shortcut, nil
		}
	}
	return nil, fmt.Errorf("field shortcut not found: %s", id)
}

// GetFieldShortcutsByCategory returns field shortcuts by category
func GetFieldShortcutsByCategory(category string) []FieldShortcut {
	var shortcuts []FieldShortcut
	for _, shortcut := range FieldShortcuts {
		if strings.EqualFold(shortcut.Category, category) {
			shortcuts = append(shortcuts, shortcut)
		}
	}
	return shortcuts
}

// GetFieldShortcutsByTag returns field shortcuts by tag
func GetFieldShortcutsByTag(tag string) []FieldShortcut {
	var shortcuts []FieldShortcut
	for _, shortcut := range FieldShortcuts {
		for _, t := range shortcut.Tags {
			if strings.EqualFold(t, tag) {
				shortcuts = append(shortcuts, shortcut)
				break
			}
		}
	}
	return shortcuts
}