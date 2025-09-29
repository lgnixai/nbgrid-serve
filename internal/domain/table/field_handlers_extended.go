package table

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// MultiSelectFieldHandler 多选字段处理器
type MultiSelectFieldHandler struct {
	BaseFieldHandler
}

// NewMultiSelectFieldHandler 创建多选字段处理器
func NewMultiSelectFieldHandler() FieldTypeHandler {
	return &MultiSelectFieldHandler{
		BaseFieldHandler: BaseFieldHandler{
			fieldType: FieldTypeMultiSelect,
			info: FieldTypeInfo{
				Type:        FieldTypeMultiSelect,
				Name:        "多选",
				Description: "从预定义选项中选择多个",
				Category:    "选择",
				Icon:        "multi_select",
				Color:       "#F06292",
			},
		},
	}
}

// ValidateValue 验证多选值
func (h *MultiSelectFieldHandler) ValidateValue(value interface{}, options *FieldOptions) error {
	if value == nil || value == "" {
		return nil
	}

	var values []string
	switch v := value.(type) {
	case string:
		// 如果是逗号分隔的字符串
		if v != "" {
			values = strings.Split(v, ",")
		}
	case []string:
		values = v
	case []interface{}:
		for _, item := range v {
			if str, ok := item.(string); ok {
				values = append(values, str)
			} else {
				return fmt.Errorf("多选字段值必须是字符串数组")
			}
		}
	default:
		return fmt.Errorf("多选字段值必须是字符串数组或逗号分隔的字符串")
	}

	if options == nil || len(options.Choices) == 0 {
		return fmt.Errorf("多选字段必须配置选项")
	}

	// 检查每个值是否在选项中
	for _, val := range values {
		val = strings.TrimSpace(val)
		if val == "" {
			continue
		}

		found := false
		for _, choice := range options.Choices {
			if choice.Value == val {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("选择的值 %s 不在有效选项中", val)
		}
	}

	return nil
}

// ValidateOptions 验证选项配置
func (h *MultiSelectFieldHandler) ValidateOptions(options *FieldOptions) error {
	if err := h.BaseFieldHandler.ValidateOptions(options); err != nil {
		return err
	}

	if options == nil || len(options.Choices) == 0 {
		return fmt.Errorf("多选字段必须配置至少一个选项")
	}

	// 检查选项值的唯一性
	valueMap := make(map[string]bool)
	for _, choice := range options.Choices {
		if choice.Value == "" {
			return fmt.Errorf("选项值不能为空")
		}
		if valueMap[choice.Value] {
			return fmt.Errorf("选项值 '%s' 重复", choice.Value)
		}
		valueMap[choice.Value] = true
	}

	return nil
}

// GetDefaultOptions 获取默认选项
func (h *MultiSelectFieldHandler) GetDefaultOptions() *FieldOptions {
	return &FieldOptions{
		Choices: []FieldChoice{
			{ID: "1", Label: "选项1", Value: "option1", Color: "#4CAF50"},
			{ID: "2", Label: "选项2", Value: "option2", Color: "#2196F3"},
			{ID: "3", Label: "选项3", Value: "option3", Color: "#FF9800"},
		},
	}
}

// RequiresOptions 需要选项配置
func (h *MultiSelectFieldHandler) RequiresOptions() bool {
	return true
}

// SupportsUnique 不支持唯一性约束
func (h *MultiSelectFieldHandler) SupportsUnique() bool {
	return false
}

// IsCompatibleWith 检查兼容性
func (h *MultiSelectFieldHandler) IsCompatibleWith(targetType FieldType) bool {
	compatibleTypes := []FieldType{
		FieldTypeMultiSelect, FieldTypeSelect, FieldTypeCheckbox, FieldTypeText,
	}

	for _, t := range compatibleTypes {
		if t == targetType {
			return true
		}
	}
	return false
}

// ConvertValue 转换值
func (h *MultiSelectFieldHandler) ConvertValue(value interface{}, targetType FieldType, targetOptions *FieldOptions) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	var values []string
	switch v := value.(type) {
	case string:
		if v != "" {
			values = strings.Split(v, ",")
		}
	case []string:
		values = v
	case []interface{}:
		for _, item := range v {
			if str, ok := item.(string); ok {
				values = append(values, str)
			}
		}
	default:
		return nil, fmt.Errorf("无法转换多选值")
	}

	switch targetType {
	case FieldTypeMultiSelect, FieldTypeCheckbox:
		return values, nil
	case FieldTypeSelect:
		// 取第一个值作为单选值
		if len(values) > 0 {
			return values[0], nil
		}
		return "", nil
	case FieldTypeText:
		// 将数组转换为逗号分隔的字符串
		return strings.Join(values, ","), nil
	default:
		return nil, fmt.Errorf("不支持从多选到 %s 的转换", targetType)
	}
}

// BooleanFieldHandler 布尔字段处理器
type BooleanFieldHandler struct {
	BaseFieldHandler
}

// NewBooleanFieldHandler 创建布尔字段处理器
func NewBooleanFieldHandler() FieldTypeHandler {
	return &BooleanFieldHandler{
		BaseFieldHandler: BaseFieldHandler{
			fieldType: FieldTypeBoolean,
			info: FieldTypeInfo{
				Type:        FieldTypeBoolean,
				Name:        "布尔值",
				Description: "是/否选择",
				Category:    "基础",
				Icon:        "checkbox",
				Color:       "#FF9800",
			},
		},
	}
}

// ValidateValue 验证布尔值
func (h *BooleanFieldHandler) ValidateValue(value interface{}, options *FieldOptions) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case bool:
		return nil
	case string:
		if v == "true" || v == "false" || v == "1" || v == "0" || v == "" {
			return nil
		}
		return fmt.Errorf("布尔字段值必须是 true/false 或 1/0")
	case int:
		if v == 0 || v == 1 {
			return nil
		}
		return fmt.Errorf("布尔字段整数值必须是 0 或 1")
	default:
		return fmt.Errorf("布尔字段值必须是布尔类型")
	}
}

// GetDefaultOptions 获取默认选项
func (h *BooleanFieldHandler) GetDefaultOptions() *FieldOptions {
	return &FieldOptions{}
}

// IsCompatibleWith 检查兼容性
func (h *BooleanFieldHandler) IsCompatibleWith(targetType FieldType) bool {
	compatibleTypes := []FieldType{
		FieldTypeBoolean, FieldTypeText,
	}

	for _, t := range compatibleTypes {
		if t == targetType {
			return true
		}
	}
	return false
}

// ConvertValue 转换值
func (h *BooleanFieldHandler) ConvertValue(value interface{}, targetType FieldType, targetOptions *FieldOptions) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	var boolVal bool
	switch v := value.(type) {
	case bool:
		boolVal = v
	case string:
		boolVal = v == "true" || v == "1"
	case int:
		boolVal = v == 1
	default:
		return nil, fmt.Errorf("无法转换布尔值")
	}

	switch targetType {
	case FieldTypeBoolean:
		return boolVal, nil
	case FieldTypeText:
		if boolVal {
			return "true", nil
		}
		return "false", nil
	default:
		return nil, fmt.Errorf("不支持从布尔到 %s 的转换", targetType)
	}
}

// URLFieldHandler URL字段处理器
type URLFieldHandler struct {
	BaseFieldHandler
}

// NewURLFieldHandler 创建URL字段处理器
func NewURLFieldHandler() FieldTypeHandler {
	return &URLFieldHandler{
		BaseFieldHandler: BaseFieldHandler{
			fieldType: FieldTypeURL,
			info: FieldTypeInfo{
				Type:        FieldTypeURL,
				Name:        "链接",
				Description: "URL链接输入",
				Category:    "高级",
				Icon:        "link",
				Color:       "#009688",
			},
		},
	}
}

// ValidateValue 验证URL值
func (h *URLFieldHandler) ValidateValue(value interface{}, options *FieldOptions) error {
	if value == nil || value == "" {
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("URL字段值必须是字符串")
	}

	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	if !urlRegex.MatchString(str) {
		return fmt.Errorf("URL格式不正确")
	}

	return nil
}

// GetDefaultOptions 获取默认选项
func (h *URLFieldHandler) GetDefaultOptions() *FieldOptions {
	return &FieldOptions{}
}

// IsCompatibleWith 检查兼容性
func (h *URLFieldHandler) IsCompatibleWith(targetType FieldType) bool {
	compatibleTypes := []FieldType{
		FieldTypeURL, FieldTypeText, FieldTypeEmail,
	}

	for _, t := range compatibleTypes {
		if t == targetType {
			return true
		}
	}
	return false
}

// PhoneFieldHandler 电话字段处理器
type PhoneFieldHandler struct {
	BaseFieldHandler
}

// NewPhoneFieldHandler 创建电话字段处理器
func NewPhoneFieldHandler() FieldTypeHandler {
	return &PhoneFieldHandler{
		BaseFieldHandler: BaseFieldHandler{
			fieldType: FieldTypePhone,
			info: FieldTypeInfo{
				Type:        FieldTypePhone,
				Name:        "电话",
				Description: "电话号码输入",
				Category:    "高级",
				Icon:        "phone",
				Color:       "#4DB6AC",
			},
		},
	}
}

// ValidateValue 验证电话值
func (h *PhoneFieldHandler) ValidateValue(value interface{}, options *FieldOptions) error {
	if value == nil || value == "" {
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("电话字段值必须是字符串")
	}

	phoneRegex := regexp.MustCompile(`^[\+]?[1-9][\d]{0,15}$`)
	if !phoneRegex.MatchString(str) {
		return fmt.Errorf("电话格式不正确")
	}

	return nil
}

// GetDefaultOptions 获取默认选项
func (h *PhoneFieldHandler) GetDefaultOptions() *FieldOptions {
	return &FieldOptions{}
}

// IsCompatibleWith 检查兼容性
func (h *PhoneFieldHandler) IsCompatibleWith(targetType FieldType) bool {
	compatibleTypes := []FieldType{
		FieldTypePhone, FieldTypeText,
	}

	for _, t := range compatibleTypes {
		if t == targetType {
			return true
		}
	}
	return false
}

// CurrencyFieldHandler 货币字段处理器
type CurrencyFieldHandler struct {
	BaseFieldHandler
}

// NewCurrencyFieldHandler 创建货币字段处理器
func NewCurrencyFieldHandler() FieldTypeHandler {
	return &CurrencyFieldHandler{
		BaseFieldHandler: BaseFieldHandler{
			fieldType: FieldTypeCurrency,
			info: FieldTypeInfo{
				Type:        FieldTypeCurrency,
				Name:        "货币",
				Description: "货币金额输入",
				Category:    "高级",
				Icon:        "currency",
				Color:       "#26A69A",
			},
		},
	}
}

// ValidateValue 验证货币值
func (h *CurrencyFieldHandler) ValidateValue(value interface{}, options *FieldOptions) error {
	// 货币字段本质上是数字字段
	numberHandler := NewNumberFieldHandler()
	return numberHandler.ValidateValue(value, options)
}

// GetDefaultOptions 获取默认选项
func (h *CurrencyFieldHandler) GetDefaultOptions() *FieldOptions {
	return &FieldOptions{
		Decimal: 2,
	}
}

// IsCompatibleWith 检查兼容性
func (h *CurrencyFieldHandler) IsCompatibleWith(targetType FieldType) bool {
	compatibleTypes := []FieldType{
		FieldTypeCurrency, FieldTypeNumber, FieldTypePercent,
	}

	for _, t := range compatibleTypes {
		if t == targetType {
			return true
		}
	}
	return false
}

// ConvertValue 转换值
func (h *CurrencyFieldHandler) ConvertValue(value interface{}, targetType FieldType, targetOptions *FieldOptions) (interface{}, error) {
	// 使用数字字段处理器的转换逻辑
	numberHandler := NewNumberFieldHandler()
	return numberHandler.ConvertValue(value, targetType, targetOptions)
}

// PercentFieldHandler 百分比字段处理器
type PercentFieldHandler struct {
	BaseFieldHandler
}

// NewPercentFieldHandler 创建百分比字段处理器
func NewPercentFieldHandler() FieldTypeHandler {
	return &PercentFieldHandler{
		BaseFieldHandler: BaseFieldHandler{
			fieldType: FieldTypePercent,
			info: FieldTypeInfo{
				Type:        FieldTypePercent,
				Name:        "百分比",
				Description: "百分比输入",
				Category:    "高级",
				Icon:        "percent",
				Color:       "#66BB6A",
			},
		},
	}
}

// ValidateValue 验证百分比值
func (h *PercentFieldHandler) ValidateValue(value interface{}, options *FieldOptions) error {
	if value == nil || value == "" {
		return nil
	}

	var num float64
	switch v := value.(type) {
	case float64:
		num = v
	case int:
		num = float64(v)
	case string:
		var err error
		num, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf("百分比格式不正确")
		}
	default:
		return fmt.Errorf("百分比字段值必须是数字")
	}

	if num < 0 || num > 100 {
		return fmt.Errorf("百分比值必须在0-100之间")
	}

	return nil
}

// GetDefaultOptions 获取默认选项
func (h *PercentFieldHandler) GetDefaultOptions() *FieldOptions {
	return &FieldOptions{
		Decimal:  1,
		MinValue: func() *float64 { v := 0.0; return &v }(),
		MaxValue: func() *float64 { v := 100.0; return &v }(),
	}
}

// IsCompatibleWith 检查兼容性
func (h *PercentFieldHandler) IsCompatibleWith(targetType FieldType) bool {
	compatibleTypes := []FieldType{
		FieldTypePercent, FieldTypeNumber, FieldTypeCurrency, FieldTypeProgress,
	}

	for _, t := range compatibleTypes {
		if t == targetType {
			return true
		}
	}
	return false
}

// ConvertValue 转换值
func (h *PercentFieldHandler) ConvertValue(value interface{}, targetType FieldType, targetOptions *FieldOptions) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	var num float64
	switch v := value.(type) {
	case float64:
		num = v
	case int:
		num = float64(v)
	case string:
		var err error
		num, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("无法转换为数字: %v", err)
		}
	default:
		return nil, fmt.Errorf("无法转换非数字值")
	}

	switch targetType {
	case FieldTypePercent, FieldTypeProgress:
		// 确保在0-100范围内
		if num < 0 {
			num = 0
		} else if num > 100 {
			num = 100
		}
		return num, nil
	case FieldTypeNumber, FieldTypeCurrency:
		return num, nil
	default:
		return nil, fmt.Errorf("不支持从百分比到 %s 的转换", targetType)
	}
}

// 创建其他字段类型的简单处理器
func NewRadioFieldHandler() FieldTypeHandler {
	handler := NewSelectFieldHandler().(*SelectFieldHandler)
	handler.fieldType = FieldTypeRadio
	handler.info = FieldTypeInfo{
		Type:        FieldTypeRadio,
		Name:        "单选按钮",
		Description: "单选按钮组",
		Category:    "选择",
		Icon:        "radio",
		Color:       "#BA68C8",
	}
	return handler
}

func NewCheckboxFieldHandler() FieldTypeHandler {
	handler := NewMultiSelectFieldHandler().(*MultiSelectFieldHandler)
	handler.fieldType = FieldTypeCheckbox
	handler.info = FieldTypeInfo{
		Type:        FieldTypeCheckbox,
		Name:        "复选框",
		Description: "复选框组",
		Category:    "选择",
		Icon:        "checkbox",
		Color:       "#9575CD",
	}
	return handler
}

func NewDateTimeFieldHandler() FieldTypeHandler {
	return &DateTimeFieldHandler{
		BaseFieldHandler: BaseFieldHandler{
			fieldType: FieldTypeDateTime,
			info: FieldTypeInfo{
				Type:        FieldTypeDateTime,
				Name:        "日期时间",
				Description: "日期和时间选择器",
				Category:    "基础",
				Icon:        "datetime",
				Color:       "#673AB7",
			},
		},
	}
}

func NewTimeFieldHandler() FieldTypeHandler {
	return &TimeFieldHandler{
		BaseFieldHandler: BaseFieldHandler{
			fieldType: FieldTypeTime,
			info: FieldTypeInfo{
				Type:        FieldTypeTime,
				Name:        "时间",
				Description: "时间选择器",
				Category:    "基础",
				Icon:        "time",
				Color:       "#3F51B5",
			},
		},
	}
}

// DateTimeFieldHandler 日期时间字段处理器
type DateTimeFieldHandler struct {
	BaseFieldHandler
}

// ValidateValue 验证日期时间值
func (h *DateTimeFieldHandler) ValidateValue(value interface{}, options *FieldOptions) error {
	if value == nil || value == "" {
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("日期时间字段值必须是字符串")
	}

	format := "2006-01-02 15:04:05"
	if options != nil && options.DateFormat != "" {
		format = options.DateFormat
	}

	_, err := time.Parse(format, str)
	if err != nil {
		return fmt.Errorf("日期时间格式不正确，期望格式: %s", format)
	}

	return nil
}

// GetDefaultOptions 获取默认选项
func (h *DateTimeFieldHandler) GetDefaultOptions() *FieldOptions {
	return &FieldOptions{
		DateFormat: "2006-01-02 15:04:05",
	}
}

// IsCompatibleWith 检查兼容性
func (h *DateTimeFieldHandler) IsCompatibleWith(targetType FieldType) bool {
	compatibleTypes := []FieldType{
		FieldTypeDateTime, FieldTypeDate, FieldTypeTime, FieldTypeText,
	}

	for _, t := range compatibleTypes {
		if t == targetType {
			return true
		}
	}
	return false
}

// TimeFieldHandler 时间字段处理器
type TimeFieldHandler struct {
	BaseFieldHandler
}

// ValidateValue 验证时间值
func (h *TimeFieldHandler) ValidateValue(value interface{}, options *FieldOptions) error {
	if value == nil || value == "" {
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("时间字段值必须是字符串")
	}

	format := "15:04:05"
	if options != nil && options.TimeFormat != "" {
		format = options.TimeFormat
	}

	_, err := time.Parse(format, str)
	if err != nil {
		return fmt.Errorf("时间格式不正确，期望格式: %s", format)
	}

	return nil
}

// GetDefaultOptions 获取默认选项
func (h *TimeFieldHandler) GetDefaultOptions() *FieldOptions {
	return &FieldOptions{
		TimeFormat: "15:04:05",
	}
}

// IsCompatibleWith 检查兼容性
func (h *TimeFieldHandler) IsCompatibleWith(targetType FieldType) bool {
	compatibleTypes := []FieldType{
		FieldTypeTime, FieldTypeDateTime, FieldTypeText,
	}

	for _, t := range compatibleTypes {
		if t == targetType {
			return true
		}
	}
	return false
}
