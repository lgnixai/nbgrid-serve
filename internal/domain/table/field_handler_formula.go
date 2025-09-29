package table

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Knetic/govaluate"
)

// FormulaFieldHandler handles formula field operations
type FormulaFieldHandler struct {
	BaseFieldHandler
}

// NewFormulaFieldHandler creates a new formula field handler
func NewFormulaFieldHandler() FieldTypeHandler {
	return &FormulaFieldHandler{
		BaseFieldHandler: BaseFieldHandler{
			fieldType: FieldTypeVirtualFormula,
			info: FieldTypeInfo{
				Type:        FieldTypeVirtualFormula,
				Name:        "公式",
				Description: "基于其他字段值计算的虚拟字段",
				Category:    "虚拟字段",
				Icon:        "formula",
				Color:       "#9C27B0",
			},
		},
	}
}

// ValidateOptions validates formula field options
func (h *FormulaFieldHandler) ValidateOptions(options *FieldOptions) error {
	if options == nil {
		return fmt.Errorf("formula field requires options")
	}

	// Parse formula options
	formulaOpts, err := ParseFormulaOptions(options)
	if err != nil {
		return err
	}

	if formulaOpts.Expression == "" {
		return fmt.Errorf("formula expression is required")
	}

	// Validate expression syntax
	if err := h.validateExpression(formulaOpts.Expression); err != nil {
		return fmt.Errorf("invalid formula expression: %w", err)
	}

	// Validate result type
	if !isValidResultType(formulaOpts.ResultType) {
		return fmt.Errorf("invalid result type: %s", formulaOpts.ResultType)
	}

	return nil
}

// ValidateValue validates a formula field value (always returns nil as values are computed)
func (h *FormulaFieldHandler) ValidateValue(value interface{}, options *FieldOptions) error {
	// Formula fields are read-only, values are always computed
	return nil
}

// FormatValue formats a formula field value
func (h *FormulaFieldHandler) FormatValue(value interface{}, options *FieldOptions) (string, error) {
	if value == nil {
		return "", nil
	}

	// Format based on result type
	formulaOpts, _ := ParseFormulaOptions(options)
	if formulaOpts != nil {
		switch formulaOpts.ResultType {
		case FieldTypeNumber, FieldTypeCurrency, FieldTypePercent:
			return formatNumberValue(value)
		case FieldTypeBoolean:
			return formatBooleanValue(value)
		case FieldTypeDate, FieldTypeDateTime:
			return formatDateValue(value)
		default:
			return fmt.Sprintf("%v", value), nil
		}
	}

	return fmt.Sprintf("%v", value), nil
}

// ParseValue parses a formula field value (not applicable for formulas)
func (h *FormulaFieldHandler) ParseValue(str string, options *FieldOptions) (interface{}, error) {
	// Formula fields are read-only
	return nil, fmt.Errorf("formula fields are read-only")
}

// GetDefaultValue returns the default value for a formula field
func (h *FormulaFieldHandler) GetDefaultValue(options *FieldOptions) interface{} {
	return nil // Formulas have no default value
}

// IsCompatibleWith checks if formula field is compatible with another type
func (h *FormulaFieldHandler) IsCompatibleWith(targetType FieldType) bool {
	// Formula fields can only convert to their result type
	return false
}

// ConvertValue converts a formula value to another type
func (h *FormulaFieldHandler) ConvertValue(value interface{}, targetType FieldType, targetOptions *FieldOptions) (interface{}, error) {
	return nil, fmt.Errorf("formula fields cannot be converted to other types")
}

// Calculate computes the formula value
func (h *FormulaFieldHandler) Calculate(ctx CalculationContext) (interface{}, error) {
	options := ctx.Field.Options
	if options == nil {
		return nil, fmt.Errorf("formula field requires options")
	}

	formulaOpts, err := ParseFormulaOptions(options)
	if err != nil {
		return nil, err
	}

	// Prepare expression with field values
	expression, err := h.prepareExpression(formulaOpts.Expression, ctx.RecordData)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare expression: %w", err)
	}

	// Evaluate expression
	result, err := h.evaluateExpression(expression)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate expression: %w", err)
	}

	// Convert result to expected type
	return h.convertResult(result, formulaOpts.ResultType)
}

// validateExpression validates the formula expression syntax
func (h *FormulaFieldHandler) validateExpression(expression string) error {
	// Remove field references for validation
	testExpr := regexp.MustCompile(`\{[^}]+\}`).ReplaceAllString(expression, "1")
	
	// Try to parse the expression
	_, err := govaluate.NewEvaluableExpression(testExpr)
	return err
}

// prepareExpression replaces field references with actual values
func (h *FormulaFieldHandler) prepareExpression(expression string, recordData map[string]interface{}) (string, error) {
	// Find all field references like {field_name}
	re := regexp.MustCompile(`\{([^}]+)\}`)
	
	result := re.ReplaceAllStringFunc(expression, func(match string) string {
		fieldName := strings.Trim(match, "{}")
		
		// Get field value from record data
		if value, exists := recordData[fieldName]; exists {
			// Convert value to string representation for expression
			return h.valueToExpressionString(value)
		}
		
		// Field not found, use null
		return "null"
	})
	
	return result, nil
}

// valueToExpressionString converts a value to its string representation in the expression
func (h *FormulaFieldHandler) valueToExpressionString(value interface{}) string {
	if value == nil {
		return "null"
	}
	
	switch v := value.(type) {
	case string:
		// Escape strings
		return fmt.Sprintf(`"%s"`, strings.ReplaceAll(v, `"`, `\"`))
	case bool:
		return fmt.Sprintf("%t", v)
	case int, int32, int64, float32, float64:
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf(`"%v"`, v)
	}
}

// evaluateExpression evaluates the prepared expression
func (h *FormulaFieldHandler) evaluateExpression(expression string) (interface{}, error) {
	// Create evaluable expression
	expr, err := govaluate.NewEvaluableExpression(expression)
	if err != nil {
		return nil, err
	}
	
	// Add custom functions
	functions := map[string]govaluate.ExpressionFunction{
		"ABS": func(args ...interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("ABS requires 1 argument")
			}
			val, err := toFloat64(args[0])
			if err != nil {
				return nil, err
			}
			if val < 0 {
				return -val, nil
			}
			return val, nil
		},
		"ROUND": func(args ...interface{}) (interface{}, error) {
			if len(args) < 1 || len(args) > 2 {
				return nil, fmt.Errorf("ROUND requires 1 or 2 arguments")
			}
			val, err := toFloat64(args[0])
			if err != nil {
				return nil, err
			}
			precision := 0
			if len(args) == 2 {
				p, err := toFloat64(args[1])
				if err != nil {
					return nil, err
				}
				precision = int(p)
			}
			multiplier := 1.0
			for i := 0; i < precision; i++ {
				multiplier *= 10
			}
			return float64(int(val*multiplier+0.5)) / multiplier, nil
		},
		"CONCAT": func(args ...interface{}) (interface{}, error) {
			var result strings.Builder
			for _, arg := range args {
				result.WriteString(fmt.Sprintf("%v", arg))
			}
			return result.String(), nil
		},
		"IF": func(args ...interface{}) (interface{}, error) {
			if len(args) != 3 {
				return nil, fmt.Errorf("IF requires 3 arguments")
			}
			condition, err := toBool(args[0])
			if err != nil {
				return nil, err
			}
			if condition {
				return args[1], nil
			}
			return args[2], nil
		},
		"LEN": func(args ...interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("LEN requires 1 argument")
			}
			str := fmt.Sprintf("%v", args[0])
			return float64(len(str)), nil
		},
	}
	
	// Evaluate with custom functions
	result, err := expr.Evaluate(functions)
	if err != nil {
		return nil, err
	}
	
	return result, nil
}

// convertResult converts the evaluation result to the expected type
func (h *FormulaFieldHandler) convertResult(result interface{}, resultType FieldType) (interface{}, error) {
	if result == nil {
		return nil, nil
	}
	
	switch resultType {
	case FieldTypeText:
		return fmt.Sprintf("%v", result), nil
		
	case FieldTypeNumber, FieldTypeCurrency, FieldTypePercent:
		return toFloat64(result)
		
	case FieldTypeBoolean:
		return toBool(result)
		
	case FieldTypeDate, FieldTypeDateTime:
		// TODO: Implement date conversion
		return fmt.Sprintf("%v", result), nil
		
	default:
		return result, nil
	}
}

// Helper functions

func toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case string:
		return strconv.ParseFloat(val, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}

func toBool(v interface{}) (bool, error) {
	switch val := v.(type) {
	case bool:
		return val, nil
	case string:
		return strconv.ParseBool(val)
	case float64:
		return val != 0, nil
	case int:
		return val != 0, nil
	default:
		return false, fmt.Errorf("cannot convert %T to bool", v)
	}
}

func formatNumberValue(value interface{}) (string, error) {
	f, err := toFloat64(value)
	if err != nil {
		return "", err
	}
	// Check if it's a whole number
	if f == float64(int64(f)) {
		return fmt.Sprintf("%.0f", f), nil
	}
	return fmt.Sprintf("%.2f", f), nil
}

func formatBooleanValue(value interface{}) (string, error) {
	b, err := toBool(value)
	if err != nil {
		return "", err
	}
	if b {
		return "是", nil
	}
	return "否", nil
}

func formatDateValue(value interface{}) (string, error) {
	// TODO: Implement proper date formatting
	return fmt.Sprintf("%v", value), nil
}

// ParseFormulaOptions parses formula field options
func ParseFormulaOptions(options *FieldOptions) (*FormulaFieldOptions, error) {
	if options == nil {
		return nil, fmt.Errorf("options is nil")
	}
	
	// Create a temporary map to extract formula-specific fields
	tempMap := map[string]interface{}{
		"expression":           options.Formula,
		"result_type":         "text", // default
		"referenced_fields":   []string{},
		"dynamic_calculation": true,
	}
	
	// If there's a Formula field, use it as expression
	if options.Formula != "" {
		tempMap["expression"] = options.Formula
	}
	
	// TODO: Parse from options JSON structure if needed
	
	formulaOpts := &FormulaFieldOptions{
		Expression:         tempMap["expression"].(string),
		ResultType:        FieldType(tempMap["result_type"].(string)),
		ReferencedFields:  tempMap["referenced_fields"].([]string),
		DynamicCalculation: tempMap["dynamic_calculation"].(bool),
	}
	
	// Extract referenced fields from expression
	re := regexp.MustCompile(`\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(formulaOpts.Expression, -1)
	for _, match := range matches {
		if len(match) > 1 {
			formulaOpts.ReferencedFields = append(formulaOpts.ReferencedFields, match[1])
		}
	}
	
	return formulaOpts, nil
}

// isValidResultType checks if the result type is valid
func isValidResultType(fieldType FieldType) bool {
	switch fieldType {
	case FieldTypeText, FieldTypeNumber, FieldTypeBoolean,
	     FieldTypeDate, FieldTypeDateTime, FieldTypeCurrency,
	     FieldTypePercent:
		return true
	default:
		return false
	}
}