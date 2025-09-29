package table

import (
	"fmt"
	"regexp"
	"strconv"
	// "strings" // 暂时注释掉，如果需要可以取消注释
	"time"
)

// BaseFieldHandler 基础字段处理器 - 提供通用实现
type BaseFieldHandler struct {
	fieldType FieldType
	info      FieldTypeInfo
}

// GetType 获取字段类型
func (h *BaseFieldHandler) GetType() FieldType {
	return h.fieldType
}

// GetInfo 获取字段类型信息
func (h *BaseFieldHandler) GetInfo() FieldTypeInfo {
	return h.info
}

// ValidateOptions 验证字段选项配置 - 默认实现
func (h *BaseFieldHandler) ValidateOptions(options *FieldOptions) error {
	if options == nil {
		return nil
	}

	// 验证通用选项
	if options.MinLength < 0 {
		return fmt.Errorf("最小长度不能为负数")
	}
	if options.MaxLength < 0 {
		return fmt.Errorf("最大长度不能为负数")
	}
	if options.MinLength > 0 && options.MaxLength > 0 && options.MinLength > options.MaxLength {
		return fmt.Errorf("最小长度不能大于最大长度")
	}

	return nil
}

// SupportsUnique 是否支持唯一性约束 - 默认实现
func (h *BaseFieldHandler) SupportsUnique() bool {
	return true
}

// RequiresOptions 是否需要选项配置 - 默认实现
func (h *BaseFieldHandler) RequiresOptions() bool {
	return false
}

// GetValidationRules 获取验证规则 - 默认实现
func (h *BaseFieldHandler) GetValidationRules(options *FieldOptions) []FieldValidationRule {
	var rules []FieldValidationRule

	if options != nil {
		if options.MinLength > 0 {
			rules = append(rules, FieldValidationRule{
				Type:    "min_length",
				Value:   options.MinLength,
				Message: fmt.Sprintf("最小长度为 %d", options.MinLength),
			})
		}
		if options.MaxLength > 0 {
			rules = append(rules, FieldValidationRule{
				Type:    "max_length",
				Value:   options.MaxLength,
				Message: fmt.Sprintf("最大长度为 %d", options.MaxLength),
			})
		}
	}

	return rules
}

// ValidateValue 验证字段值 - 默认实现
func (h *BaseFieldHandler) ValidateValue(value interface{}, options *FieldOptions) error {
	// 默认实现：接受任何值
	return nil
}

// GetDefaultOptions 获取默认选项配置 - 默认实现
func (h *BaseFieldHandler) GetDefaultOptions() *FieldOptions {
	return &FieldOptions{}
}

// IsCompatibleWith 检查是否与目标类型兼容 - 默认实现
func (h *BaseFieldHandler) IsCompatibleWith(targetType FieldType) bool {
	// 默认情况下，只有相同类型才兼容
	return h.fieldType == targetType
}

// ConvertValue 转换值到目标类型 - 默认实现
func (h *BaseFieldHandler) ConvertValue(value interface{}, targetType FieldType, targetOptions *FieldOptions) (interface{}, error) {
	// 默认情况下，如果类型兼容，直接返回原值
	if h.IsCompatibleWith(targetType) {
		return value, nil
	}

	return nil, fmt.Errorf("不支持从 %s 到 %s 的值转换", h.fieldType, targetType)
}

// TextFieldHandler 文本字段处理器
type TextFieldHandler struct {
	BaseFieldHandler
}

// NewTextFieldHandler 创建文本字段处理器
func NewTextFieldHandler() FieldTypeHandler {
	return &TextFieldHandler{
		BaseFieldHandler: BaseFieldHandler{
			fieldType: FieldTypeText,
			info: FieldTypeInfo{
				Type:        FieldTypeText,
				Name:        "文本",
				Description: "单行文本输入",
				Category:    "基础",
				Icon:        "text",
				Color:       "#4CAF50",
			},
		},
	}
}

// ValidateValue 验证文本值
func (h *TextFieldHandler) ValidateValue(value interface{}, options *FieldOptions) error {
	if value == nil || value == "" {
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("文本字段值必须是字符串")
	}

	if options != nil {
		if options.MinLength > 0 && len(str) < options.MinLength {
			return fmt.Errorf("文本长度不能少于 %d 个字符", options.MinLength)
		}
		if options.MaxLength > 0 && len(str) > options.MaxLength {
			return fmt.Errorf("文本长度不能超过 %d 个字符", options.MaxLength)
		}
		if options.Pattern != "" {
			matched, err := regexp.MatchString(options.Pattern, str)
			if err != nil {
				return fmt.Errorf("正则表达式错误: %v", err)
			}
			if !matched {
				return fmt.Errorf("文本格式不正确")
			}
		}
	}

	return nil
}

// GetDefaultOptions 获取默认选项
func (h *TextFieldHandler) GetDefaultOptions() *FieldOptions {
	return &FieldOptions{
		MaxLength: 255,
	}
}

// IsCompatibleWith 检查兼容性
func (h *TextFieldHandler) IsCompatibleWith(targetType FieldType) bool {
	compatibleTypes := []FieldType{
		FieldTypeText, FieldTypeEmail, FieldTypeURL, FieldTypePhone,
		FieldTypeSelect, FieldTypeMultiSelect,
	}

	for _, t := range compatibleTypes {
		if t == targetType {
			return true
		}
	}
	return false
}

// ConvertValue 转换值
func (h *TextFieldHandler) ConvertValue(value interface{}, targetType FieldType, targetOptions *FieldOptions) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	str, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("无法转换非字符串值")
	}

	switch targetType {
	case FieldTypeText, FieldTypeEmail, FieldTypeURL, FieldTypePhone:
		return str, nil
	case FieldTypeSelect:
		// 检查值是否在选项中
		if targetOptions != nil && len(targetOptions.Choices) > 0 {
			for _, choice := range targetOptions.Choices {
				if choice.Value == str {
					return str, nil
				}
			}
			return nil, fmt.Errorf("值 '%s' 不在目标字段的选项中", str)
		}
		return str, nil
	case FieldTypeMultiSelect:
		// 将单个值转换为数组
		return []string{str}, nil
	default:
		return nil, fmt.Errorf("不支持从文本到 %s 的转换", targetType)
	}
}

// NumberFieldHandler 数字字段处理器
type NumberFieldHandler struct {
	BaseFieldHandler
}

// NewNumberFieldHandler 创建数字字段处理器
func NewNumberFieldHandler() FieldTypeHandler {
	return &NumberFieldHandler{
		BaseFieldHandler: BaseFieldHandler{
			fieldType: FieldTypeNumber,
			info: FieldTypeInfo{
				Type:        FieldTypeNumber,
				Name:        "数字",
				Description: "数字输入，支持整数和小数",
				Category:    "基础",
				Icon:        "number",
				Color:       "#2196F3",
			},
		},
	}
}

// ValidateValue 验证数字值
func (h *NumberFieldHandler) ValidateValue(value interface{}, options *FieldOptions) error {
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
			return fmt.Errorf("数字格式不正确")
		}
	default:
		return fmt.Errorf("数字字段值必须是数字")
	}

	if options != nil {
		if options.MinValue != nil && num < *options.MinValue {
			return fmt.Errorf("数字不能小于 %v", *options.MinValue)
		}
		if options.MaxValue != nil && num > *options.MaxValue {
			return fmt.Errorf("数字不能大于 %v", *options.MaxValue)
		}
	}

	return nil
}

// GetDefaultOptions 获取默认选项
func (h *NumberFieldHandler) GetDefaultOptions() *FieldOptions {
	return &FieldOptions{
		Decimal: 2,
	}
}

// IsCompatibleWith 检查兼容性
func (h *NumberFieldHandler) IsCompatibleWith(targetType FieldType) bool {
	compatibleTypes := []FieldType{
		FieldTypeNumber, FieldTypeCurrency, FieldTypePercent,
		FieldTypeRating, FieldTypeProgress,
	}

	for _, t := range compatibleTypes {
		if t == targetType {
			return true
		}
	}
	return false
}

// ConvertValue 转换值
func (h *NumberFieldHandler) ConvertValue(value interface{}, targetType FieldType, targetOptions *FieldOptions) (interface{}, error) {
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
	case FieldTypeNumber, FieldTypeCurrency:
		return num, nil
	case FieldTypePercent:
		// 确保百分比在0-100范围内
		if num < 0 {
			num = 0
		} else if num > 100 {
			num = 100
		}
		return num, nil
	case FieldTypeRating:
		// 转换为整数评分
		return int(num), nil
	case FieldTypeProgress:
		// 确保进度在0-100范围内
		if num < 0 {
			num = 0
		} else if num > 100 {
			num = 100
		}
		return num, nil
	default:
		return nil, fmt.Errorf("不支持从数字到 %s 的转换", targetType)
	}
}

// SelectFieldHandler 选择字段处理器
type SelectFieldHandler struct {
	BaseFieldHandler
}

// NewSelectFieldHandler 创建选择字段处理器
func NewSelectFieldHandler() FieldTypeHandler {
	return &SelectFieldHandler{
		BaseFieldHandler: BaseFieldHandler{
			fieldType: FieldTypeSelect,
			info: FieldTypeInfo{
				Type:        FieldTypeSelect,
				Name:        "单选",
				Description: "从预定义选项中选择一个",
				Category:    "选择",
				Icon:        "select",
				Color:       "#E91E63",
			},
		},
	}
}

// ValidateValue 验证选择值
func (h *SelectFieldHandler) ValidateValue(value interface{}, options *FieldOptions) error {
	if value == nil || value == "" {
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("选择字段值必须是字符串")
	}

	if options == nil || len(options.Choices) == 0 {
		return fmt.Errorf("选择字段必须配置选项")
	}

	// 检查值是否在选项中
	for _, choice := range options.Choices {
		if choice.Value == str {
			return nil
		}
	}

	return fmt.Errorf("选择的值不在有效选项中")
}

// ValidateOptions 验证选项配置
func (h *SelectFieldHandler) ValidateOptions(options *FieldOptions) error {
	if err := h.BaseFieldHandler.ValidateOptions(options); err != nil {
		return err
	}

	if options == nil || len(options.Choices) == 0 {
		return fmt.Errorf("选择字段必须配置至少一个选项")
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
func (h *SelectFieldHandler) GetDefaultOptions() *FieldOptions {
	return &FieldOptions{
		Choices: []FieldChoice{
			{ID: "1", Label: "选项1", Value: "option1", Color: "#4CAF50"},
			{ID: "2", Label: "选项2", Value: "option2", Color: "#2196F3"},
		},
	}
}

// RequiresOptions 需要选项配置
func (h *SelectFieldHandler) RequiresOptions() bool {
	return true
}

// IsCompatibleWith 检查兼容性
func (h *SelectFieldHandler) IsCompatibleWith(targetType FieldType) bool {
	compatibleTypes := []FieldType{
		FieldTypeSelect, FieldTypeMultiSelect, FieldTypeRadio, FieldTypeText,
	}

	for _, t := range compatibleTypes {
		if t == targetType {
			return true
		}
	}
	return false
}

// ConvertValue 转换值
func (h *SelectFieldHandler) ConvertValue(value interface{}, targetType FieldType, targetOptions *FieldOptions) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	str, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("无法转换非字符串值")
	}

	switch targetType {
	case FieldTypeSelect, FieldTypeRadio, FieldTypeText:
		return str, nil
	case FieldTypeMultiSelect:
		// 将单选值转换为多选数组
		return []string{str}, nil
	default:
		return nil, fmt.Errorf("不支持从选择到 %s 的转换", targetType)
	}
}

// EmailFieldHandler 邮箱字段处理器
type EmailFieldHandler struct {
	BaseFieldHandler
}

// NewEmailFieldHandler 创建邮箱字段处理器
func NewEmailFieldHandler() FieldTypeHandler {
	return &EmailFieldHandler{
		BaseFieldHandler: BaseFieldHandler{
			fieldType: FieldTypeEmail,
			info: FieldTypeInfo{
				Type:        FieldTypeEmail,
				Name:        "邮箱",
				Description: "邮箱地址输入",
				Category:    "高级",
				Icon:        "email",
				Color:       "#00BCD4",
			},
		},
	}
}

// ValidateValue 验证邮箱值
func (h *EmailFieldHandler) ValidateValue(value interface{}, options *FieldOptions) error {
	if value == nil || value == "" {
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("邮箱字段值必须是字符串")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(str) {
		return fmt.Errorf("邮箱格式不正确")
	}

	return nil
}

// GetDefaultOptions 获取默认选项
func (h *EmailFieldHandler) GetDefaultOptions() *FieldOptions {
	return &FieldOptions{}
}

// IsCompatibleWith 检查兼容性
func (h *EmailFieldHandler) IsCompatibleWith(targetType FieldType) bool {
	compatibleTypes := []FieldType{
		FieldTypeEmail, FieldTypeText, FieldTypeURL,
	}

	for _, t := range compatibleTypes {
		if t == targetType {
			return true
		}
	}
	return false
}

// DateFieldHandler 日期字段处理器
type DateFieldHandler struct {
	BaseFieldHandler
}

// NewDateFieldHandler 创建日期字段处理器
func NewDateFieldHandler() FieldTypeHandler {
	return &DateFieldHandler{
		BaseFieldHandler: BaseFieldHandler{
			fieldType: FieldTypeDate,
			info: FieldTypeInfo{
				Type:        FieldTypeDate,
				Name:        "日期",
				Description: "日期选择器",
				Category:    "基础",
				Icon:        "calendar",
				Color:       "#9C27B0",
			},
		},
	}
}

// ValidateValue 验证日期值
func (h *DateFieldHandler) ValidateValue(value interface{}, options *FieldOptions) error {
	if value == nil || value == "" {
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("日期字段值必须是字符串")
	}

	format := "2006-01-02"
	if options != nil && options.DateFormat != "" {
		format = options.DateFormat
	}

	_, err := time.Parse(format, str)
	if err != nil {
		return fmt.Errorf("日期格式不正确，期望格式: %s", format)
	}

	return nil
}

// GetDefaultOptions 获取默认选项
func (h *DateFieldHandler) GetDefaultOptions() *FieldOptions {
	return &FieldOptions{
		DateFormat: "2006-01-02",
	}
}

// IsCompatibleWith 检查兼容性
func (h *DateFieldHandler) IsCompatibleWith(targetType FieldType) bool {
	compatibleTypes := []FieldType{
		FieldTypeDate, FieldTypeDateTime, FieldTypeText,
	}

	for _, t := range compatibleTypes {
		if t == targetType {
			return true
		}
	}
	return false
}

// ConvertValue 转换值
func (h *DateFieldHandler) ConvertValue(value interface{}, targetType FieldType, targetOptions *FieldOptions) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	str, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("无法转换非字符串值")
	}

	// 解析当前日期
	currentFormat := "2006-01-02"
	date, err := time.Parse(currentFormat, str)
	if err != nil {
		return nil, fmt.Errorf("无法解析日期: %v", err)
	}

	switch targetType {
	case FieldTypeDate:
		return date.Format("2006-01-02"), nil
	case FieldTypeDateTime:
		// 转换为日期时间，时间部分设为00:00:00
		return date.Format("2006-01-02 15:04:05"), nil
	case FieldTypeText:
		return str, nil
	default:
		return nil, fmt.Errorf("不支持从日期到 %s 的转换", targetType)
	}
}

// 占位符构造函数 - 这些需要在后续实现中完善

// NewRatingFieldHandler 创建评分字段处理器
func NewRatingFieldHandler() FieldTypeHandler {
	return &BaseFieldHandler{
		fieldType: FieldTypeRating,
		info: FieldTypeInfo{
			Name:        "评分",
			Description: "评分字段",
			Category:    "数值",
		},
	}
}

// NewProgressFieldHandler 创建进度字段处理器
func NewProgressFieldHandler() FieldTypeHandler {
	return &BaseFieldHandler{
		fieldType: FieldTypeProgress,
		info: FieldTypeInfo{
			Name:        "进度",
			Description: "进度字段",
			Category:    "数值",
		},
	}
}

// NewImageFieldHandler 创建图片字段处理器
func NewImageFieldHandler() FieldTypeHandler {
	return &BaseFieldHandler{
		fieldType: FieldTypeImage,
		info: FieldTypeInfo{
			Name:        "图片",
			Description: "图片字段",
			Category:    "媒体",
		},
	}
}

// NewFileFieldHandler 创建文件字段处理器
func NewFileFieldHandler() FieldTypeHandler {
	return &BaseFieldHandler{
		fieldType: FieldTypeFile,
		info: FieldTypeInfo{
			Name:        "文件",
			Description: "文件字段",
			Category:    "媒体",
		},
	}
}

// NewVideoFieldHandler 创建视频字段处理器
func NewVideoFieldHandler() FieldTypeHandler {
	return &BaseFieldHandler{
		fieldType: FieldTypeVideo,
		info: FieldTypeInfo{
			Name:        "视频",
			Description: "视频字段",
			Category:    "媒体",
		},
	}
}

// NewAudioFieldHandler 创建音频字段处理器
func NewAudioFieldHandler() FieldTypeHandler {
	return &BaseFieldHandler{
		fieldType: FieldTypeAudio,
		info: FieldTypeInfo{
			Name:        "音频",
			Description: "音频字段",
			Category:    "媒体",
		},
	}
}

// NewLookupFieldHandler 创建查找字段处理器
func NewLookupFieldHandler() FieldTypeHandler {
	return &BaseFieldHandler{
		fieldType: FieldTypeLookup,
		info: FieldTypeInfo{
			Name:        "查找",
			Description: "查找字段",
			Category:    "关系",
		},
	}
}

// NewRollupFieldHandler 创建汇总字段处理器
func NewRollupFieldHandler() FieldTypeHandler {
	return &BaseFieldHandler{
		fieldType: FieldTypeRollup,
		info: FieldTypeInfo{
			Name:        "汇总",
			Description: "汇总字段",
			Category:    "关系",
		},
	}
}

// NewFormulaFieldHandler 创建公式字段处理器
func NewFormulaFieldHandler() FieldTypeHandler {
	return &BaseFieldHandler{
		fieldType: FieldTypeFormula,
		info: FieldTypeInfo{
			Name:        "公式",
			Description: "公式字段",
			Category:    "计算",
		},
	}
}

// NewAutoNumberFieldHandler 创建自动编号字段处理器
func NewAutoNumberFieldHandler() FieldTypeHandler {
	return &BaseFieldHandler{
		fieldType: FieldTypeAutoNumber,
		info: FieldTypeInfo{
			Name:        "自动编号",
			Description: "自动编号字段",
			Category:    "特殊",
		},
	}
}

// NewCreatedTimeFieldHandler 创建创建时间字段处理器
func NewCreatedTimeFieldHandler() FieldTypeHandler {
	return &BaseFieldHandler{
		fieldType: FieldTypeCreatedTime,
		info: FieldTypeInfo{
			Name:        "创建时间",
			Description: "创建时间字段",
			Category:    "特殊",
		},
	}
}

// NewLastModifiedTimeFieldHandler 创建最后修改时间字段处理器
func NewLastModifiedTimeFieldHandler() FieldTypeHandler {
	return &BaseFieldHandler{
		fieldType: FieldTypeLastModifiedTime,
		info: FieldTypeInfo{
			Name:        "最后修改时间",
			Description: "最后修改时间字段",
			Category:    "特殊",
		},
	}
}

// NewCreatedByFieldHandler 创建创建者字段处理器
func NewCreatedByFieldHandler() FieldTypeHandler {
	return &BaseFieldHandler{
		fieldType: FieldTypeCreatedBy,
		info: FieldTypeInfo{
			Name:        "创建者",
			Description: "创建者字段",
			Category:    "特殊",
		},
	}
}

// NewLastModifiedByFieldHandler 创建最后修改者字段处理器
func NewLastModifiedByFieldHandler() FieldTypeHandler {
	return &BaseFieldHandler{
		fieldType: FieldTypeLastModifiedBy,
		info: FieldTypeInfo{
			Name:        "最后修改者",
			Description: "最后修改者字段",
			Category:    "特殊",
		},
	}
}
