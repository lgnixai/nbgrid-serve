package table

import (
	"encoding/json"
	"fmt"
	"time"

	"teable-go-backend/pkg/utils"
)

// Table 数据表实体 - 支持动态schema
type Table struct {
	ID               string     `json:"id"`
	BaseID           string     `json:"base_id"`
	Name             string     `json:"name"`
	Description      *string    `json:"description"`
	Icon             *string    `json:"icon"`
	IsSystem         bool       `json:"is_system"`
	CreatedBy        string     `json:"created_by"`
	CreatedTime      time.Time  `json:"created_at"`
	DeletedTime      *time.Time `json:"deleted_time"`
	LastModifiedTime *time.Time `json:"updated_at"`

	// 动态schema支持
	fields        []*Field `json:"-"` // 内部字段缓存，不序列化
	schemaVersion int64    `json:"schema_version"`
}

// Field 字段实体 - 增强的字段类型系统
type Field struct {
	ID               string        `json:"id"`
	TableID          string        `json:"table_id"`
	Name             string        `json:"name"`
	Type             FieldType     `json:"type"`
	Description      *string       `json:"description"`
	IsRequired       bool          `json:"required"`
	IsUnique         bool          `json:"is_unique"`
	IsPrimary        bool          `json:"is_primary"`
	IsComputed       bool          `json:"is_computed"` // 计算字段
	IsLookup         bool          `json:"is_lookup"`   // 查找字段
	DefaultValue     *string       `json:"default_value"`
	Options          *FieldOptions `json:"options"` // 强类型选项配置
	FieldOrder       int           `json:"field_order"`
	Version          int64         `json:"version"` // 字段版本号
	CreatedBy        string        `json:"created_by"`
	CreatedTime      time.Time     `json:"created_at"`
	DeletedTime      *time.Time    `json:"deleted_time"`
	LastModifiedTime *time.Time    `json:"updated_at"`
}

// CreateTableRequest 创建数据表请求
type CreateTableRequest struct {
	BaseID      string  `json:"base_id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Icon        *string `json:"icon"`
	CreatedBy   string  `json:"created_by,omitempty"`
}

// UpdateTableRequest 更新数据表请求
type UpdateTableRequest struct {
	Name        *string
	Description *string
	Icon        *string
}

// CreateFieldRequest 创建字段请求
type CreateFieldRequest struct {
	TableID      string        `json:"table_id"`
	Name         string        `json:"name"`
	Type         string        `json:"type"`
	Description  *string       `json:"description"`
	IsRequired   bool          `json:"required"`
	IsUnique     bool          `json:"is_unique"`
	IsPrimary    bool          `json:"is_primary"`
	DefaultValue *string       `json:"default_value"`
	Options      *string       `json:"options"`                 // JSON字符串，向后兼容
	FieldOptions *FieldOptions `json:"field_options,omitempty"` // 强类型选项
	FieldOrder   int           `json:"field_order"`
	CreatedBy    string        `json:"created_by"`
}

// UpdateFieldRequest 更新字段请求
type UpdateFieldRequest struct {
	Name         *string       `json:"name,omitempty"`
	Type         *string       `json:"type,omitempty"`
	Description  *string       `json:"description,omitempty"`
	IsRequired   *bool         `json:"required,omitempty"`
	IsUnique     *bool         `json:"is_unique,omitempty"`
	IsPrimary    *bool         `json:"is_primary,omitempty"`
	DefaultValue *string       `json:"default_value,omitempty"`
	Options      *string       `json:"options,omitempty"`       // JSON字符串，向后兼容
	FieldOptions *FieldOptions `json:"field_options,omitempty"` // 强类型选项
	FieldOrder   *int          `json:"field_order,omitempty"`
}

// ListTableFilter 表格列表过滤条件
type ListTableFilter struct {
	BaseID    *string `form:"base_id" json:"base_id"`
	Name      *string `form:"name" json:"name"`
	CreatedBy *string `form:"created_by" json:"created_by"`
	Search    string  `form:"search" json:"search"`
	OrderBy   string  `form:"order_by" json:"order_by"`
	Order     string  `form:"order" json:"order"`
	Limit     int     `form:"limit" json:"limit"`
	Offset    int     `form:"offset" json:"offset"`
}

// ListFieldFilter 字段列表过滤条件
type ListFieldFilter struct {
	TableID   *string `form:"table_id" json:"table_id"`
	Name      *string `form:"name" json:"name"`
	Type      *string `form:"type" json:"type"`
	CreatedBy *string `form:"created_by" json:"created_by"`
	OrderBy   string  `form:"order_by" json:"order_by"`
	Order     string  `form:"order" json:"order"`
	Limit     int     `form:"limit" json:"limit"`
	Offset    int     `form:"offset" json:"offset"`
}

// BulkUpdateTableRequest 批量更新表格请求
type BulkUpdateTableRequest struct {
	TableID string             `json:"table_id" validate:"required"`
	Updates UpdateTableRequest `json:"updates" validate:"required"`
}

// BulkUpdateFieldRequest 批量更新字段请求
type BulkUpdateFieldRequest struct {
	FieldID string             `json:"field_id" validate:"required"`
	Updates UpdateFieldRequest `json:"updates" validate:"required"`
}

// TableStats 表格统计信息
type TableStats struct {
	TableID        string  `json:"table_id"`
	TotalFields    int64   `json:"total_fields"`
	TotalRecords   int64   `json:"total_records"`
	TotalViews     int64   `json:"total_views"`
	LastActivityAt *string `json:"last_activity_at,omitempty"`
}

// BaseTableStats 基础表表格统计信息
type BaseTableStats struct {
	BaseID       string `json:"base_id"`
	TotalTables  int64  `json:"total_tables"`
	TotalFields  int64  `json:"total_fields"`
	TotalRecords int64  `json:"total_records"`
	TotalViews   int64  `json:"total_views"`
}

// NewTable 创建新的数据表
func NewTable(req CreateTableRequest) *Table {
	now := time.Now()
	return &Table{
		ID:               utils.GenerateTableID(),
		BaseID:           req.BaseID,
		Name:             req.Name,
		Description:      req.Description,
		Icon:             req.Icon,
		IsSystem:         false,
		CreatedBy:        req.CreatedBy,
		CreatedTime:      now,
		LastModifiedTime: &now,
		fields:           make([]*Field, 0),
		schemaVersion:    1,
	}
}

// GetFields 获取表格字段列表
func (t *Table) GetFields() []*Field {
	return t.fields
}

// SetFields 设置表格字段列表
func (t *Table) SetFields(fields []*Field) {
	t.fields = fields
}

// AddField 添加字段到表格
func (t *Table) AddField(field *Field) error {
	// 验证字段名称唯一性
	if t.HasFieldWithName(field.Name) {
		return fmt.Errorf("字段名称 '%s' 已存在", field.Name)
	}

	// 验证主键字段唯一性
	if field.IsPrimary && t.HasPrimaryField() {
		return fmt.Errorf("表格已存在主键字段")
	}

	t.fields = append(t.fields, field)
	t.incrementSchemaVersion()
	return nil
}

// RemoveField 从表格中移除字段
func (t *Table) RemoveField(fieldID string) error {
	for i, field := range t.fields {
		if field.ID == fieldID {
			// 不允许删除主键字段
			if field.IsPrimary {
				return fmt.Errorf("不能删除主键字段")
			}

			// 软删除字段
			field.SoftDelete()
			t.fields = append(t.fields[:i], t.fields[i+1:]...)
			t.incrementSchemaVersion()
			return nil
		}
	}
	return fmt.Errorf("字段未找到")
}

// HasFieldWithName 检查是否存在指定名称的字段
func (t *Table) HasFieldWithName(name string) bool {
	for _, field := range t.fields {
		if field.Name == name && field.DeletedTime == nil {
			return true
		}
	}
	return false
}

// HasPrimaryField 检查是否存在主键字段
func (t *Table) HasPrimaryField() bool {
	for _, field := range t.fields {
		if field.IsPrimary && field.DeletedTime == nil {
			return true
		}
	}
	return false
}

// GetPrimaryField 获取主键字段
func (t *Table) GetPrimaryField() *Field {
	for _, field := range t.fields {
		if field.IsPrimary && field.DeletedTime == nil {
			return field
		}
	}
	return nil
}

// GetFieldByName 根据名称获取字段
func (t *Table) GetFieldByName(name string) *Field {
	for _, field := range t.fields {
		if field.Name == name && field.DeletedTime == nil {
			return field
		}
	}
	return nil
}

// GetFieldByID 根据ID获取字段
func (t *Table) GetFieldByID(id string) *Field {
	for _, field := range t.fields {
		if field.ID == id && field.DeletedTime == nil {
			return field
		}
	}
	return nil
}

// ValidateSchema 验证表格schema的完整性
func (t *Table) ValidateSchema() error {
	if len(t.fields) == 0 {
		return fmt.Errorf("表格必须至少包含一个字段")
	}

	if !t.HasPrimaryField() {
		return fmt.Errorf("表格必须包含一个主键字段")
	}

	// 验证字段名称唯一性
	nameMap := make(map[string]bool)
	for _, field := range t.fields {
		if field.DeletedTime != nil {
			continue
		}
		if nameMap[field.Name] {
			return fmt.Errorf("字段名称 '%s' 重复", field.Name)
		}
		nameMap[field.Name] = true
	}

	return nil
}

// GetSchemaVersion 获取schema版本号
func (t *Table) GetSchemaVersion() int64 {
	return t.schemaVersion
}

// incrementSchemaVersion 递增schema版本号
func (t *Table) incrementSchemaVersion() {
	t.schemaVersion++
	now := time.Now()
	t.LastModifiedTime = &now
}

// NewField 创建新的字段
func NewField(req CreateFieldRequest) *Field {
	now := time.Now()

	// 解析选项配置
	var options *FieldOptions
	if req.Options != nil && *req.Options != "" {
		var opts FieldOptions
		if err := json.Unmarshal([]byte(*req.Options), &opts); err == nil {
			options = &opts
		}
	}

	return &Field{
		ID:               utils.GenerateFieldID(),
		TableID:          req.TableID,
		Name:             req.Name,
		Type:             FieldType(req.Type),
		Description:      req.Description,
		IsRequired:       req.IsRequired,
		IsUnique:         req.IsUnique,
		IsPrimary:        req.IsPrimary,
		IsComputed:       false,
		IsLookup:         false,
		DefaultValue:     req.DefaultValue,
		Options:          options,
		FieldOrder:       req.FieldOrder,
		Version:          1,
		CreatedBy:        req.CreatedBy,
		CreatedTime:      now,
		LastModifiedTime: &now,
	}
}

// ValidateValue 验证字段值
func (f *Field) ValidateValue(value interface{}) error {
	// 必填字段检查
	if f.IsRequired && (value == nil || value == "") {
		return fmt.Errorf("字段 '%s' 是必填的", f.Name)
	}

	// 空值检查
	if value == nil || value == "" {
		return nil
	}

	// 根据字段类型进行验证
	return f.Type.ValidateValue(value, f.Options)
}

// GetOptionsAsMap 获取选项配置作为map
func (f *Field) GetOptionsAsMap() map[string]interface{} {
	if f.Options == nil {
		return make(map[string]interface{})
	}

	// 将FieldOptions转换为map
	optionsMap := make(map[string]interface{})

	if f.Options.Placeholder != "" {
		optionsMap["placeholder"] = f.Options.Placeholder
	}
	if f.Options.HelpText != "" {
		optionsMap["help_text"] = f.Options.HelpText
	}
	if len(f.Options.Choices) > 0 {
		optionsMap["choices"] = f.Options.Choices
	}
	if f.Options.MinValue != nil {
		optionsMap["min_value"] = *f.Options.MinValue
	}
	if f.Options.MaxValue != nil {
		optionsMap["max_value"] = *f.Options.MaxValue
	}
	if f.Options.Decimal > 0 {
		optionsMap["decimal"] = f.Options.Decimal
	}
	if f.Options.MinLength > 0 {
		optionsMap["min_length"] = f.Options.MinLength
	}
	if f.Options.MaxLength > 0 {
		optionsMap["max_length"] = f.Options.MaxLength
	}
	if f.Options.Pattern != "" {
		optionsMap["pattern"] = f.Options.Pattern
	}
	if f.Options.DateFormat != "" {
		optionsMap["date_format"] = f.Options.DateFormat
	}
	if f.Options.TimeFormat != "" {
		optionsMap["time_format"] = f.Options.TimeFormat
	}
	if f.Options.MaxFileSize > 0 {
		optionsMap["max_file_size"] = f.Options.MaxFileSize
	}
	if len(f.Options.AllowedTypes) > 0 {
		optionsMap["allowed_types"] = f.Options.AllowedTypes
	}
	if f.Options.LinkTableID != "" {
		optionsMap["link_table_id"] = f.Options.LinkTableID
	}
	if f.Options.LinkFieldID != "" {
		optionsMap["link_field_id"] = f.Options.LinkFieldID
	}
	if f.Options.Formula != "" {
		optionsMap["formula"] = f.Options.Formula
	}
	if len(f.Options.ValidationRules) > 0 {
		optionsMap["validation_rules"] = f.Options.ValidationRules
	}

	return optionsMap
}

// CanChangeTypeTo 检查是否可以将字段类型更改为指定类型
func (f *Field) CanChangeTypeTo(newType FieldType) (bool, error) {
	// 主键字段不能更改类型
	if f.IsPrimary {
		return false, fmt.Errorf("主键字段不能更改类型")
	}

	// 计算字段不能更改类型
	if f.IsComputed {
		return false, fmt.Errorf("计算字段不能更改类型")
	}

	// 检查类型兼容性
	return f.Type.IsCompatibleWith(newType), nil
}

// ChangeType 安全地更改字段类型
func (f *Field) ChangeType(newType FieldType, newOptions *FieldOptions) error {
	canChange, err := f.CanChangeTypeTo(newType)
	if err != nil {
		return err
	}
	if !canChange {
		return fmt.Errorf("字段类型从 %s 到 %s 的转换不兼容", f.Type, newType)
	}

	f.Type = newType
	f.Options = newOptions
	f.incrementVersion()
	return nil
}

// SetDefaultValue 设置默认值
func (f *Field) SetDefaultValue(value *string) error {
	// 验证默认值是否符合字段类型
	if value != nil {
		if err := f.ValidateValue(*value); err != nil {
			return fmt.Errorf("默认值验证失败: %v", err)
		}
	}

	f.DefaultValue = value
	f.incrementVersion()
	return nil
}

// SetRequired 设置必填属性
func (f *Field) SetRequired(required bool) error {
	// 如果设置为必填，但有默认值为空，则不允许
	if required && (f.DefaultValue == nil || *f.DefaultValue == "") {
		return fmt.Errorf("设置为必填字段时必须提供默认值")
	}

	f.IsRequired = required
	f.incrementVersion()
	return nil
}

// SetUnique 设置唯一性约束
func (f *Field) SetUnique(unique bool) error {
	// 某些字段类型不支持唯一性约束
	if unique && !f.Type.SupportsUnique() {
		return fmt.Errorf("字段类型 %s 不支持唯一性约束", f.Type)
	}

	f.IsUnique = unique
	f.incrementVersion()
	return nil
}

// GetTypeInfo 获取字段类型信息
func (f *Field) GetTypeInfo() FieldTypeInfo {
	return GetFieldTypeInfo(f.Type)
}

// IsSystemField 检查是否为系统字段
func (f *Field) IsSystemField() bool {
	return f.Type == FieldTypeCreatedTime ||
		f.Type == FieldTypeLastModifiedTime ||
		f.Type == FieldTypeCreatedBy ||
		f.Type == FieldTypeLastModifiedBy ||
		f.Type == FieldTypeAutoNumber
}

// CanBeDeleted 检查字段是否可以被删除
func (f *Field) CanBeDeleted() (bool, error) {
	// 主键字段不能删除
	if f.IsPrimary {
		return false, fmt.Errorf("主键字段不能删除")
	}

	// 系统字段不能删除
	if f.IsSystemField() {
		return false, fmt.Errorf("系统字段不能删除")
	}

	return true, nil
}

// GetVersion 获取字段版本号
func (f *Field) GetVersion() int64 {
	return f.Version
}

// incrementVersion 递增字段版本号
func (f *Field) incrementVersion() {
	f.Version++
	now := time.Now()
	f.LastModifiedTime = &now
}

// Update 更新数据表信息
func (t *Table) Update(req UpdateTableRequest) {
	if req.Name != nil {
		t.Name = *req.Name
	}
	if req.Description != nil {
		t.Description = req.Description
	}
	if req.Icon != nil {
		t.Icon = req.Icon
	}
	now := time.Now()
	t.LastModifiedTime = &now
}

// Update 更新字段信息
func (f *Field) Update(req UpdateFieldRequest) error {
	if req.Name != nil {
		f.Name = *req.Name
	}
	if req.Type != nil {
		newType := FieldType(*req.Type)
		if err := f.ChangeType(newType, req.FieldOptions); err != nil {
			return err
		}
	}
	if req.Description != nil {
		f.Description = req.Description
	}
	if req.IsRequired != nil {
		if err := f.SetRequired(*req.IsRequired); err != nil {
			return err
		}
	}
	if req.IsUnique != nil {
		if err := f.SetUnique(*req.IsUnique); err != nil {
			return err
		}
	}
	if req.IsPrimary != nil {
		f.IsPrimary = *req.IsPrimary
	}
	if req.DefaultValue != nil {
		if err := f.SetDefaultValue(req.DefaultValue); err != nil {
			return err
		}
	}

	// 处理选项配置
	if req.FieldOptions != nil {
		f.Options = req.FieldOptions
	} else if req.Options != nil {
		// 向后兼容：解析JSON字符串
		var options FieldOptions
		if err := json.Unmarshal([]byte(*req.Options), &options); err == nil {
			f.Options = &options
		}
	}

	if req.FieldOrder != nil {
		f.FieldOrder = *req.FieldOrder
	}

	f.incrementVersion()
	return nil
}

// SoftDelete 软删除数据表
func (t *Table) SoftDelete() {
	now := time.Now()
	t.DeletedTime = &now
	t.LastModifiedTime = &now
}

// SoftDelete 软删除字段
func (f *Field) SoftDelete() {
	now := time.Now()
	f.DeletedTime = &now
	f.LastModifiedTime = &now
}
