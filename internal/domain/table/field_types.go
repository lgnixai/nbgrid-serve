package table

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// FieldType 字段类型枚举
type FieldType string

const (
	// 基础类型
	FieldTypeText     FieldType = "text"
	FieldTypeNumber   FieldType = "number"
	FieldTypeBoolean  FieldType = "boolean"
	FieldTypeDate     FieldType = "date"
	FieldTypeDateTime FieldType = "datetime"
	FieldTypeTime     FieldType = "time"

	// 选择类型
	FieldTypeSelect      FieldType = "select"
	FieldTypeMultiSelect FieldType = "multi_select"
	FieldTypeRadio       FieldType = "radio"
	FieldTypeCheckbox    FieldType = "checkbox"

	// 高级类型
	FieldTypeEmail    FieldType = "email"
	FieldTypeURL      FieldType = "url"
	FieldTypePhone    FieldType = "phone"
	FieldTypeCurrency FieldType = "currency"
	FieldTypePercent  FieldType = "percent"
	FieldTypeRating   FieldType = "rating"
	FieldTypeProgress FieldType = "progress"

	// 媒体类型
	FieldTypeImage FieldType = "image"
	FieldTypeFile  FieldType = "file"
	FieldTypeVideo FieldType = "video"
	FieldTypeAudio FieldType = "audio"

	// 关系类型
	FieldTypeLink    FieldType = "link"
	FieldTypeLookup  FieldType = "lookup"
	FieldTypeRollup  FieldType = "rollup"
	FieldTypeFormula FieldType = "formula"

	// 特殊类型
	FieldTypeAutoNumber       FieldType = "auto_number"
	FieldTypeCreatedTime      FieldType = "created_time"
	FieldTypeLastModifiedTime FieldType = "last_modified_time"
	FieldTypeCreatedBy        FieldType = "created_by"
	FieldTypeLastModifiedBy   FieldType = "last_modified_by"
)

// FieldTypeInfo 字段类型信息
type FieldTypeInfo struct {
	Type        FieldType `json:"type"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Icon        string    `json:"icon"`
	Color       string    `json:"color"`
}

// GetFieldTypeInfo 获取字段类型信息
func GetFieldTypeInfo(fieldType FieldType) FieldTypeInfo {
	infos := map[FieldType]FieldTypeInfo{
		// 基础类型
		FieldTypeText: {
			Type:        FieldTypeText,
			Name:        "文本",
			Description: "单行文本输入",
			Category:    "基础",
			Icon:        "text",
			Color:       "#4CAF50",
		},
		FieldTypeNumber: {
			Type:        FieldTypeNumber,
			Name:        "数字",
			Description: "数字输入，支持整数和小数",
			Category:    "基础",
			Icon:        "number",
			Color:       "#2196F3",
		},
		FieldTypeBoolean: {
			Type:        FieldTypeBoolean,
			Name:        "布尔值",
			Description: "是/否选择",
			Category:    "基础",
			Icon:        "checkbox",
			Color:       "#FF9800",
		},
		FieldTypeDate: {
			Type:        FieldTypeDate,
			Name:        "日期",
			Description: "日期选择器",
			Category:    "基础",
			Icon:        "calendar",
			Color:       "#9C27B0",
		},
		FieldTypeDateTime: {
			Type:        FieldTypeDateTime,
			Name:        "日期时间",
			Description: "日期和时间选择器",
			Category:    "基础",
			Icon:        "datetime",
			Color:       "#673AB7",
		},
		FieldTypeTime: {
			Type:        FieldTypeTime,
			Name:        "时间",
			Description: "时间选择器",
			Category:    "基础",
			Icon:        "time",
			Color:       "#3F51B5",
		},

		// 选择类型
		FieldTypeSelect: {
			Type:        FieldTypeSelect,
			Name:        "单选",
			Description: "从预定义选项中选择一个",
			Category:    "选择",
			Icon:        "select",
			Color:       "#E91E63",
		},
		FieldTypeMultiSelect: {
			Type:        FieldTypeMultiSelect,
			Name:        "多选",
			Description: "从预定义选项中选择多个",
			Category:    "选择",
			Icon:        "multi_select",
			Color:       "#F06292",
		},
		FieldTypeRadio: {
			Type:        FieldTypeRadio,
			Name:        "单选按钮",
			Description: "单选按钮组",
			Category:    "选择",
			Icon:        "radio",
			Color:       "#BA68C8",
		},
		FieldTypeCheckbox: {
			Type:        FieldTypeCheckbox,
			Name:        "复选框",
			Description: "复选框组",
			Category:    "选择",
			Icon:        "checkbox",
			Color:       "#9575CD",
		},

		// 高级类型
		FieldTypeEmail: {
			Type:        FieldTypeEmail,
			Name:        "邮箱",
			Description: "邮箱地址输入",
			Category:    "高级",
			Icon:        "email",
			Color:       "#00BCD4",
		},
		FieldTypeURL: {
			Type:        FieldTypeURL,
			Name:        "链接",
			Description: "URL链接输入",
			Category:    "高级",
			Icon:        "link",
			Color:       "#009688",
		},
		FieldTypePhone: {
			Type:        FieldTypePhone,
			Name:        "电话",
			Description: "电话号码输入",
			Category:    "高级",
			Icon:        "phone",
			Color:       "#4DB6AC",
		},
		FieldTypeCurrency: {
			Type:        FieldTypeCurrency,
			Name:        "货币",
			Description: "货币金额输入",
			Category:    "高级",
			Icon:        "currency",
			Color:       "#26A69A",
		},
		FieldTypePercent: {
			Type:        FieldTypePercent,
			Name:        "百分比",
			Description: "百分比输入",
			Category:    "高级",
			Icon:        "percent",
			Color:       "#66BB6A",
		},
		FieldTypeRating: {
			Type:        FieldTypeRating,
			Name:        "评分",
			Description: "星级评分",
			Category:    "高级",
			Icon:        "star",
			Color:       "#FFA726",
		},
		FieldTypeProgress: {
			Type:        FieldTypeProgress,
			Name:        "进度条",
			Description: "进度百分比",
			Category:    "高级",
			Icon:        "progress",
			Color:       "#AB47BC",
		},

		// 媒体类型
		FieldTypeImage: {
			Type:        FieldTypeImage,
			Name:        "图片",
			Description: "图片文件上传",
			Category:    "媒体",
			Icon:        "image",
			Color:       "#8BC34A",
		},
		FieldTypeFile: {
			Type:        FieldTypeFile,
			Name:        "文件",
			Description: "文件上传",
			Category:    "媒体",
			Icon:        "file",
			Color:       "#CDDC39",
		},
		FieldTypeVideo: {
			Type:        FieldTypeVideo,
			Name:        "视频",
			Description: "视频文件上传",
			Category:    "媒体",
			Icon:        "video",
			Color:       "#FFEB3B",
		},
		FieldTypeAudio: {
			Type:        FieldTypeAudio,
			Name:        "音频",
			Description: "音频文件上传",
			Category:    "媒体",
			Icon:        "audio",
			Color:       "#FFC107",
		},

		// 关系类型
		FieldTypeLink: {
			Type:        FieldTypeLink,
			Name:        "关联",
			Description: "关联到其他表格",
			Category:    "关系",
			Icon:        "link",
			Color:       "#FF7043",
		},
		FieldTypeLookup: {
			Type:        FieldTypeLookup,
			Name:        "查找",
			Description: "从关联表格查找数据",
			Category:    "关系",
			Icon:        "lookup",
			Color:       "#FF5722",
		},
		FieldTypeRollup: {
			Type:        FieldTypeRollup,
			Name:        "汇总",
			Description: "汇总关联表格的数据",
			Category:    "关系",
			Icon:        "rollup",
			Color:       "#795548",
		},
		FieldTypeFormula: {
			Type:        FieldTypeFormula,
			Name:        "公式",
			Description: "基于其他字段计算",
			Category:    "关系",
			Icon:        "formula",
			Color:       "#607D8B",
		},

		// 特殊类型
		FieldTypeAutoNumber: {
			Type:        FieldTypeAutoNumber,
			Name:        "自动编号",
			Description: "自动递增的数字",
			Category:    "特殊",
			Icon:        "auto_number",
			Color:       "#9E9E9E",
		},
		FieldTypeCreatedTime: {
			Type:        FieldTypeCreatedTime,
			Name:        "创建时间",
			Description: "记录创建时间",
			Category:    "特殊",
			Icon:        "created_time",
			Color:       "#424242",
		},
		FieldTypeLastModifiedTime: {
			Type:        FieldTypeLastModifiedTime,
			Name:        "最后修改时间",
			Description: "记录最后修改时间",
			Category:    "特殊",
			Icon:        "last_modified_time",
			Color:       "#616161",
		},
		FieldTypeCreatedBy: {
			Type:        FieldTypeCreatedBy,
			Name:        "创建者",
			Description: "记录创建者",
			Category:    "特殊",
			Icon:        "created_by",
			Color:       "#757575",
		},
		FieldTypeLastModifiedBy: {
			Type:        FieldTypeLastModifiedBy,
			Name:        "最后修改者",
			Description: "记录最后修改者",
			Category:    "特殊",
			Icon:        "last_modified_by",
			Color:       "#9E9E9E",
		},
	}

	if info, exists := infos[fieldType]; exists {
		return info
	}

	return FieldTypeInfo{
		Type:        fieldType,
		Name:        "未知类型",
		Description: "未知的字段类型",
		Category:    "其他",
		Icon:        "unknown",
		Color:       "#CCCCCC",
	}
}

// GetAllFieldTypes 获取所有字段类型
func GetAllFieldTypes() []FieldTypeInfo {
	types := []FieldType{
		// 基础类型
		FieldTypeText, FieldTypeNumber, FieldTypeBoolean, FieldTypeDate, FieldTypeDateTime, FieldTypeTime,
		// 选择类型
		FieldTypeSelect, FieldTypeMultiSelect, FieldTypeRadio, FieldTypeCheckbox,
		// 高级类型
		FieldTypeEmail, FieldTypeURL, FieldTypePhone, FieldTypeCurrency, FieldTypePercent, FieldTypeRating, FieldTypeProgress,
		// 媒体类型
		FieldTypeImage, FieldTypeFile, FieldTypeVideo, FieldTypeAudio,
		// 关系类型
		FieldTypeLink, FieldTypeLookup, FieldTypeRollup, FieldTypeFormula,
		// 特殊类型
		FieldTypeAutoNumber, FieldTypeCreatedTime, FieldTypeLastModifiedTime, FieldTypeCreatedBy, FieldTypeLastModifiedBy,
	}

	var infos []FieldTypeInfo
	for _, t := range types {
		infos = append(infos, GetFieldTypeInfo(t))
	}

	return infos
}

// FieldValidationRule 字段验证规则
type FieldValidationRule struct {
	Type     string      `json:"type"`
	Value    interface{} `json:"value"`
	Message  string      `json:"message"`
	Required bool        `json:"required"`
}

// FieldOptions 字段选项配置
type FieldOptions struct {
	// 通用选项
	Placeholder string `json:"placeholder,omitempty"`
	HelpText    string `json:"help_text,omitempty"`

	// 选择类型选项
	Choices []FieldChoice `json:"choices,omitempty"`

	// 数字类型选项
	MinValue *float64 `json:"min_value,omitempty"`
	MaxValue *float64 `json:"max_value,omitempty"`
	Decimal  int      `json:"decimal,omitempty"`

	// 文本类型选项
	MinLength int    `json:"min_length,omitempty"`
	MaxLength int    `json:"max_length,omitempty"`
	Pattern   string `json:"pattern,omitempty"`

	// 日期类型选项
	DateFormat string `json:"date_format,omitempty"`
	TimeFormat string `json:"time_format,omitempty"`

	// 文件类型选项
	MaxFileSize  int64    `json:"max_file_size,omitempty"`
	AllowedTypes []string `json:"allowed_types,omitempty"`

	// 关联类型选项
	LinkTableID string `json:"link_table_id,omitempty"`
	LinkFieldID string `json:"link_field_id,omitempty"`

	// 公式类型选项
	Formula string `json:"formula,omitempty"`

	// 验证规则
	ValidationRules []FieldValidationRule `json:"validation_rules,omitempty"`
}

// FieldChoice 字段选择项
type FieldChoice struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Value string `json:"value"`
	Color string `json:"color,omitempty"`
}

// ValidateFieldValue 验证字段值
func ValidateFieldValue(field *Field, value interface{}) error {
	if field.IsRequired && (value == nil || value == "") {
		return fmt.Errorf("字段 %s 是必填的", field.Name)
	}

	if value == nil || value == "" {
		return nil // 非必填字段可以为空
	}

	// 解析字段选项
	var options FieldOptions
	if field.Options != nil && *field.Options != "" {
		if err := json.Unmarshal([]byte(*field.Options), &options); err != nil {
			return fmt.Errorf("解析字段选项失败: %v", err)
		}
	}

	// 根据字段类型验证
	switch FieldType(field.Type) {
	case FieldTypeText:
		return validateTextValue(value, &options)
	case FieldTypeNumber:
		return validateNumberValue(value, &options)
	case FieldTypeEmail:
		return validateEmailValue(value, &options)
	case FieldTypeURL:
		return validateURLValue(value, &options)
	case FieldTypePhone:
		return validatePhoneValue(value, &options)
	case FieldTypeDate:
		return validateDateValue(value, &options)
	case FieldTypeDateTime:
		return validateDateTimeValue(value, &options)
	case FieldTypeTime:
		return validateTimeValue(value, &options)
	case FieldTypeSelect, FieldTypeRadio:
		return validateSelectValue(value, &options)
	case FieldTypeMultiSelect, FieldTypeCheckbox:
		return validateMultiSelectValue(value, &options)
	case FieldTypeBoolean:
		return validateBooleanValue(value, &options)
	case FieldTypeCurrency:
		return validateCurrencyValue(value, &options)
	case FieldTypePercent:
		return validatePercentValue(value, &options)
	case FieldTypeRating:
		return validateRatingValue(value, &options)
	case FieldTypeProgress:
		return validateProgressValue(value, &options)
	default:
		return nil // 其他类型暂时不验证
	}
}

// validateTextValue 验证文本值
func validateTextValue(value interface{}, options *FieldOptions) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("文本字段值必须是字符串")
	}

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

	return nil
}

// validateNumberValue 验证数字值
func validateNumberValue(value interface{}, options *FieldOptions) error {
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

	if options.MinValue != nil && num < *options.MinValue {
		return fmt.Errorf("数字不能小于 %v", *options.MinValue)
	}

	if options.MaxValue != nil && num > *options.MaxValue {
		return fmt.Errorf("数字不能大于 %v", *options.MaxValue)
	}

	return nil
}

// validateEmailValue 验证邮箱值
func validateEmailValue(value interface{}, options *FieldOptions) error {
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

// validateURLValue 验证URL值
func validateURLValue(value interface{}, options *FieldOptions) error {
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

// validatePhoneValue 验证电话值
func validatePhoneValue(value interface{}, options *FieldOptions) error {
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

// validateDateValue 验证日期值
func validateDateValue(value interface{}, options *FieldOptions) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("日期字段值必须是字符串")
	}

	format := "2006-01-02"
	if options.DateFormat != "" {
		format = options.DateFormat
	}

	_, err := time.Parse(format, str)
	if err != nil {
		return fmt.Errorf("日期格式不正确，期望格式: %s", format)
	}

	return nil
}

// validateDateTimeValue 验证日期时间值
func validateDateTimeValue(value interface{}, options *FieldOptions) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("日期时间字段值必须是字符串")
	}

	format := "2006-01-02 15:04:05"
	if options.DateFormat != "" {
		format = options.DateFormat
	}

	_, err := time.Parse(format, str)
	if err != nil {
		return fmt.Errorf("日期时间格式不正确，期望格式: %s", format)
	}

	return nil
}

// validateTimeValue 验证时间值
func validateTimeValue(value interface{}, options *FieldOptions) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("时间字段值必须是字符串")
	}

	format := "15:04:05"
	if options.TimeFormat != "" {
		format = options.TimeFormat
	}

	_, err := time.Parse(format, str)
	if err != nil {
		return fmt.Errorf("时间格式不正确，期望格式: %s", format)
	}

	return nil
}

// validateSelectValue 验证选择值
func validateSelectValue(value interface{}, options *FieldOptions) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("选择字段值必须是字符串")
	}

	// 检查值是否在选项中
	for _, choice := range options.Choices {
		if choice.Value == str {
			return nil
		}
	}

	return fmt.Errorf("选择的值不在有效选项中")
}

// validateMultiSelectValue 验证多选值
func validateMultiSelectValue(value interface{}, options *FieldOptions) error {
	var values []string
	switch v := value.(type) {
	case string:
		// 如果是逗号分隔的字符串
		values = strings.Split(v, ",")
	case []string:
		values = v
	case []interface{}:
		for _, item := range v {
			if str, ok := item.(string); ok {
				values = append(values, str)
			}
		}
	default:
		return fmt.Errorf("多选字段值必须是字符串数组或逗号分隔的字符串")
	}

	// 检查每个值是否在选项中
	for _, val := range values {
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

// validateBooleanValue 验证布尔值
func validateBooleanValue(value interface{}, options *FieldOptions) error {
	switch value.(type) {
	case bool:
		return nil
	case string:
		str := value.(string)
		if str == "true" || str == "false" || str == "1" || str == "0" {
			return nil
		}
		return fmt.Errorf("布尔字段值必须是 true/false 或 1/0")
	default:
		return fmt.Errorf("布尔字段值必须是布尔类型")
	}
}

// validateCurrencyValue 验证货币值
func validateCurrencyValue(value interface{}, options *FieldOptions) error {
	return validateNumberValue(value, options)
}

// validatePercentValue 验证百分比值
func validatePercentValue(value interface{}, options *FieldOptions) error {
	err := validateNumberValue(value, options)
	if err != nil {
		return err
	}

	// 百分比值应该在0-100之间
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
	}

	if num < 0 || num > 100 {
		return fmt.Errorf("百分比值必须在0-100之间")
	}

	return nil
}

// validateRatingValue 验证评分值
func validateRatingValue(value interface{}, options *FieldOptions) error {
	err := validateNumberValue(value, options)
	if err != nil {
		return err
	}

	// 评分值应该是整数
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
			return fmt.Errorf("评分格式不正确")
		}
	}

	if num != float64(int(num)) {
		return fmt.Errorf("评分值必须是整数")
	}

	return nil
}

// validateProgressValue 验证进度值
func validateProgressValue(value interface{}, options *FieldOptions) error {
	return validatePercentValue(value, options)
}
