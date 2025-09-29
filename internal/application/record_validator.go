package application

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"teable-go-backend/internal/domain/record"
	"teable-go-backend/internal/domain/table"
)

// RecordValidator 记录验证器 - 支持动态schema验证
type RecordValidator struct {
	tableService table.Service
}

// NewRecordValidator 创建记录验证器
func NewRecordValidator(tableService table.Service) *RecordValidator {
	return &RecordValidator{
		tableService: tableService,
	}
}

// ValidateForCreate 创建时验证
func (v *RecordValidator) ValidateForCreate(ctx context.Context, rec *record.Record, tableSchema *table.Table) error {
	if err := v.ValidateData(ctx, rec, tableSchema); err != nil {
		return err
	}

	// 创建时的特殊验证
	return v.validateUniqueConstraints(ctx, rec, tableSchema, true)
}

// ValidateForUpdate 更新时验证
func (v *RecordValidator) ValidateForUpdate(ctx context.Context, rec *record.Record, tableSchema *table.Table) error {
	if err := v.ValidateData(ctx, rec, tableSchema); err != nil {
		return err
	}

	// 更新时的特殊验证
	return v.validateUniqueConstraints(ctx, rec, tableSchema, false)
}

// ValidateData 验证记录数据
func (v *RecordValidator) ValidateData(ctx context.Context, rec *record.Record, tableSchema *table.Table) error {
	if tableSchema == nil {
		return fmt.Errorf("缺少表格schema信息")
	}

	fields := tableSchema.GetFields()
	if len(fields) == 0 {
		return fmt.Errorf("表格没有定义字段")
	}

	var validationErrors []record.ValidationError

	// 验证每个字段
	for _, field := range fields {
		if field.DeletedTime != nil {
			continue // 跳过已删除的字段
		}

		value, exists := rec.Data[field.Name]

		// 验证必填字段
		if field.IsRequired {
			if !exists || v.isEmpty(value) {
				validationErrors = append(validationErrors, record.ValidationError{
					FieldName: field.Name,
					Message:   fmt.Sprintf("字段 '%s' 是必填的", field.Name),
					Code:      "REQUIRED_FIELD_MISSING",
				})
				continue
			}
		}

		// 如果字段有值，进行类型和格式验证
		if exists && !v.isEmpty(value) {
			if err := v.validateFieldValue(field, value); err != nil {
				validationErrors = append(validationErrors, record.ValidationError{
					FieldName: field.Name,
					Message:   err.Error(),
					Code:      "FIELD_VALIDATION_FAILED",
				})
			}
		}
	}

	// 检查未知字段
	for fieldName := range rec.Data {
		if tableSchema.GetFieldByName(fieldName) == nil {
			validationErrors = append(validationErrors, record.ValidationError{
				FieldName: fieldName,
				Message:   fmt.Sprintf("未知字段 '%s'", fieldName),
				Code:      "UNKNOWN_FIELD",
			})
		}
	}

	if len(validationErrors) > 0 {
		return &ValidationError{
			Message: fmt.Sprintf("记录验证失败，共 %d 个错误", len(validationErrors)),
			Errors:  validationErrors,
		}
	}

	return nil
}

// validateFieldValue 验证字段值
func (v *RecordValidator) validateFieldValue(field *table.Field, value interface{}) error {
	switch field.Type {
	case table.FieldTypeText:
		return v.validateTextField(field, value)
	case table.FieldTypeNumber:
		return v.validateNumberField(field, value)
	case table.FieldTypeDate:
		return v.validateDateField(field, value)
	case table.FieldTypeSelect:
		return v.validateSelectField(field, value)
	case table.FieldTypeMultiSelect:
		return v.validateMultiSelectField(field, value)
	case table.FieldTypeBoolean:
		return v.validateBooleanField(field, value)
	case table.FieldTypeEmail:
		return v.validateEmailField(field, value)
	case table.FieldTypeURL:
		return v.validateURLField(field, value)
	case table.FieldTypePhone:
		return v.validatePhoneField(field, value)
	case table.FieldTypeAttachment:
		return v.validateAttachmentField(field, value)
	case table.FieldTypeLink:
		return v.validateLinkField(field, value)
	default:
		// 对于未知类型，只进行基本验证
		return nil
	}
}

// validateTextField 验证文本字段
func (v *RecordValidator) validateTextField(field *table.Field, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("字段 '%s' 必须是字符串类型", field.Name)
	}

	// 检查长度限制
	if field.Options != nil {
		options := field.GetOptionsAsMap()
		if maxLength, exists := options["maxLength"]; exists {
			if maxLen, ok := maxLength.(float64); ok && len(str) > int(maxLen) {
				return fmt.Errorf("字段 '%s' 长度不能超过 %d 个字符", field.Name, int(maxLen))
			}
		}
		if minLength, exists := options["minLength"]; exists {
			if minLen, ok := minLength.(float64); ok && len(str) < int(minLen) {
				return fmt.Errorf("字段 '%s' 长度不能少于 %d 个字符", field.Name, int(minLen))
			}
		}
	}

	return nil
}

// validateNumberField 验证数字字段
func (v *RecordValidator) validateNumberField(field *table.Field, value interface{}) error {
	var num float64
	var ok bool

	switch v := value.(type) {
	case float64:
		num = v
		ok = true
	case int:
		num = float64(v)
		ok = true
	case string:
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			num = parsed
			ok = true
		}
	}

	if !ok {
		return fmt.Errorf("字段 '%s' 必须是数字类型", field.Name)
	}

	// 检查数值范围
	if field.Options != nil {
		options := field.GetOptionsAsMap()
		if min, exists := options["min"]; exists {
			if minVal, ok := min.(float64); ok && num < minVal {
				return fmt.Errorf("字段 '%s' 的值不能小于 %g", field.Name, minVal)
			}
		}
		if max, exists := options["max"]; exists {
			if maxVal, ok := max.(float64); ok && num > maxVal {
				return fmt.Errorf("字段 '%s' 的值不能大于 %g", field.Name, maxVal)
			}
		}
	}

	return nil
}

// validateDateField 验证日期字段
func (v *RecordValidator) validateDateField(field *table.Field, value interface{}) error {
	var dateStr string
	var ok bool

	switch v := value.(type) {
	case string:
		dateStr = v
		ok = true
	case time.Time:
		dateStr = v.Format(time.RFC3339)
		ok = true
	}

	if !ok {
		return fmt.Errorf("字段 '%s' 必须是日期类型", field.Name)
	}

	// 尝试解析日期
	formats := []string{
		time.RFC3339,
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006/01/02",
		"2006/01/02 15:04:05",
	}

	var parsed time.Time
	var parseErr error
	for _, format := range formats {
		if parsed, parseErr = time.Parse(format, dateStr); parseErr == nil {
			break
		}
	}

	if parseErr != nil {
		return fmt.Errorf("字段 '%s' 的日期格式无效", field.Name)
	}

	// 检查日期范围
	if field.Options != nil {
		options := field.GetOptionsAsMap()
		if minDate, exists := options["minDate"]; exists {
			if minDateStr, ok := minDate.(string); ok {
				if minTime, err := time.Parse(time.RFC3339, minDateStr); err == nil && parsed.Before(minTime) {
					return fmt.Errorf("字段 '%s' 的日期不能早于 %s", field.Name, minTime.Format("2006-01-02"))
				}
			}
		}
		if maxDate, exists := options["maxDate"]; exists {
			if maxDateStr, ok := maxDate.(string); ok {
				if maxTime, err := time.Parse(time.RFC3339, maxDateStr); err == nil && parsed.After(maxTime) {
					return fmt.Errorf("字段 '%s' 的日期不能晚于 %s", field.Name, maxTime.Format("2006-01-02"))
				}
			}
		}
	}

	return nil
}

// validateSelectField 验证单选字段
func (v *RecordValidator) validateSelectField(field *table.Field, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("字段 '%s' 必须是字符串类型", field.Name)
	}

	// 检查选项是否有效
	if field.Options != nil {
		options := field.GetOptionsAsMap()
		if choices, exists := options["choices"]; exists {
			if choiceList, ok := choices.([]interface{}); ok {
				for _, choice := range choiceList {
					if choiceStr, ok := choice.(string); ok && choiceStr == str {
						return nil
					}
				}
				return fmt.Errorf("字段 '%s' 的值 '%s' 不在有效选项中", field.Name, str)
			}
		}
	}

	return nil
}

// validateMultiSelectField 验证多选字段
func (v *RecordValidator) validateMultiSelectField(field *table.Field, value interface{}) error {
	var values []string

	switch v := value.(type) {
	case []interface{}:
		values = make([]string, len(v))
		for i, item := range v {
			if str, ok := item.(string); ok {
				values[i] = str
			} else {
				return fmt.Errorf("字段 '%s' 的多选值必须是字符串数组", field.Name)
			}
		}
	case []string:
		values = v
	default:
		return fmt.Errorf("字段 '%s' 必须是字符串数组类型", field.Name)
	}

	// 检查选项是否有效
	if field.Options != nil {
		options := field.GetOptionsAsMap()
		if choices, exists := options["choices"]; exists {
			if choiceList, ok := choices.([]interface{}); ok {
				validChoices := make(map[string]bool)
				for _, choice := range choiceList {
					if choiceStr, ok := choice.(string); ok {
						validChoices[choiceStr] = true
					}
				}

				for _, value := range values {
					if !validChoices[value] {
						return fmt.Errorf("字段 '%s' 的值 '%s' 不在有效选项中", field.Name, value)
					}
				}
			}
		}
	}

	return nil
}

// validateBooleanField 验证布尔字段
func (v *RecordValidator) validateBooleanField(field *table.Field, value interface{}) error {
	switch value.(type) {
	case bool:
		return nil
	case string:
		str := strings.ToLower(value.(string))
		if str == "true" || str == "false" || str == "1" || str == "0" {
			return nil
		}
	case int:
		intVal := value.(int)
		if intVal == 0 || intVal == 1 {
			return nil
		}
	}

	return fmt.Errorf("字段 '%s' 必须是布尔类型", field.Name)
}

// validateEmailField 验证邮箱字段
func (v *RecordValidator) validateEmailField(field *table.Field, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("字段 '%s' 必须是字符串类型", field.Name)
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(str) {
		return fmt.Errorf("字段 '%s' 的邮箱格式无效", field.Name)
	}

	return nil
}

// validateURLField 验证URL字段
func (v *RecordValidator) validateURLField(field *table.Field, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("字段 '%s' 必须是字符串类型", field.Name)
	}

	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	if !urlRegex.MatchString(str) {
		return fmt.Errorf("字段 '%s' 的URL格式无效", field.Name)
	}

	return nil
}

// validatePhoneField 验证电话字段
func (v *RecordValidator) validatePhoneField(field *table.Field, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("字段 '%s' 必须是字符串类型", field.Name)
	}

	// 简单的电话号码验证（支持国际格式）
	phoneRegex := regexp.MustCompile(`^[\+]?[1-9][\d]{0,15}$`)
	cleanPhone := regexp.MustCompile(`[^\d\+]`).ReplaceAllString(str, "")

	if !phoneRegex.MatchString(cleanPhone) {
		return fmt.Errorf("字段 '%s' 的电话号码格式无效", field.Name)
	}

	return nil
}

// validateAttachmentField 验证附件字段
func (v *RecordValidator) validateAttachmentField(field *table.Field, value interface{}) error {
	// 附件字段可以是文件ID数组或文件对象数组
	switch v := value.(type) {
	case []interface{}:
		// 验证每个附件项
		for _, item := range v {
			if itemMap, ok := item.(map[string]interface{}); ok {
				// 检查必要的附件属性
				if _, hasID := itemMap["id"]; !hasID {
					return fmt.Errorf("字段 '%s' 的附件缺少ID", field.Name)
				}
				if _, hasName := itemMap["name"]; !hasName {
					return fmt.Errorf("字段 '%s' 的附件缺少文件名", field.Name)
				}
			} else if _, ok := item.(string); !ok {
				return fmt.Errorf("字段 '%s' 的附件格式无效", field.Name)
			}
		}
	case []string:
		// 文件ID数组格式
		return nil
	default:
		return fmt.Errorf("字段 '%s' 必须是附件数组类型", field.Name)
	}

	return nil
}

// validateLinkField 验证链接字段
func (v *RecordValidator) validateLinkField(field *table.Field, value interface{}) error {
	// 链接字段可以是记录ID数组
	switch v := value.(type) {
	case []interface{}:
		for _, item := range v {
			if _, ok := item.(string); !ok {
				return fmt.Errorf("字段 '%s' 的链接值必须是字符串数组", field.Name)
			}
		}
	case []string:
		return nil
	case string:
		// 单个链接值
		return nil
	default:
		return fmt.Errorf("字段 '%s' 的链接格式无效", field.Name)
	}

	return nil
}

// validateUniqueConstraints 验证唯一性约束
func (v *RecordValidator) validateUniqueConstraints(ctx context.Context, rec *record.Record, tableSchema *table.Table, isCreate bool) error {
	// TODO: 实现唯一性约束验证
	// 这需要查询数据库检查是否存在重复值
	return nil
}

// isEmpty 检查值是否为空
func (v *RecordValidator) isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	default:
		return reflect.ValueOf(value).IsZero()
	}
}

// ValidationError 验证错误
type ValidationError struct {
	Message string                   `json:"message"`
	Errors  []record.ValidationError `json:"errors"`
}

func (e *ValidationError) Error() string {
	return e.Message
}
