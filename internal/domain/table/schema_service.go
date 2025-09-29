package table

import (
	"context"
	"fmt"
)

// SchemaChangeType 表示schema变更类型
type SchemaChangeType string

const (
	SchemaChangeAddField     SchemaChangeType = "add_field"
	SchemaChangeUpdateField  SchemaChangeType = "update_field"
	SchemaChangeDeleteField  SchemaChangeType = "delete_field"
	SchemaChangeReorderField SchemaChangeType = "reorder_field"
)

// SchemaChange 表示一个schema变更
type SchemaChange struct {
	Type        SchemaChangeType `json:"type"`
	FieldID     string           `json:"field_id,omitempty"`
	OldField    *Field           `json:"old_field,omitempty"`
	NewField    *Field           `json:"new_field,omitempty"`
	Description string           `json:"description"`
}

// SchemaChangeRequest 表示schema变更请求
type SchemaChangeRequest struct {
	TableID string         `json:"table_id"`
	Changes []SchemaChange `json:"changes"`
	UserID  string         `json:"user_id"`
}

// SchemaChangeResult 表示schema变更结果
type SchemaChangeResult struct {
	Success    bool           `json:"success"`
	Changes    []SchemaChange `json:"changes"`
	Errors     []string       `json:"errors,omitempty"`
	Warnings   []string       `json:"warnings,omitempty"`
	NewVersion int64          `json:"new_version"`
}

// SchemaService 提供安全的schema变更服务
type SchemaService interface {
	// ValidateSchemaChange 验证schema变更的安全性
	ValidateSchemaChange(ctx context.Context, table *Table, changes []SchemaChange) error

	// ApplySchemaChanges 应用schema变更
	ApplySchemaChanges(ctx context.Context, req SchemaChangeRequest) (*SchemaChangeResult, error)

	// PreviewSchemaChanges 预览schema变更的影响
	PreviewSchemaChanges(ctx context.Context, table *Table, changes []SchemaChange) (*SchemaChangeResult, error)

	// CanSafelyChangeFieldType 检查是否可以安全地更改字段类型
	CanSafelyChangeFieldType(ctx context.Context, field *Field, newType FieldType) (bool, []string, error)

	// GetSchemaHistory 获取表格的schema变更历史
	GetSchemaHistory(ctx context.Context, tableID string) ([]SchemaChange, error)
}

// SchemaServiceImpl schema服务实现
type SchemaServiceImpl struct {
	tableRepo Repository
}

// NewSchemaService 创建schema服务
func NewSchemaService(tableRepo Repository) SchemaService {
	return &SchemaServiceImpl{
		tableRepo: tableRepo,
	}
}

// ValidateSchemaChange 验证schema变更的安全性
func (s *SchemaServiceImpl) ValidateSchemaChange(ctx context.Context, table *Table, changes []SchemaChange) error {
	for _, change := range changes {
		switch change.Type {
		case SchemaChangeAddField:
			if err := s.validateAddField(table, change.NewField); err != nil {
				return fmt.Errorf("添加字段验证失败: %v", err)
			}
		case SchemaChangeUpdateField:
			if err := s.validateUpdateField(table, change.OldField, change.NewField); err != nil {
				return fmt.Errorf("更新字段验证失败: %v", err)
			}
		case SchemaChangeDeleteField:
			if err := s.validateDeleteField(table, change.OldField); err != nil {
				return fmt.Errorf("删除字段验证失败: %v", err)
			}
		}
	}
	return nil
}

// ApplySchemaChanges 应用schema变更
func (s *SchemaServiceImpl) ApplySchemaChanges(ctx context.Context, req SchemaChangeRequest) (*SchemaChangeResult, error) {
	// 获取表格
	table, err := s.tableRepo.GetTableByID(ctx, req.TableID)
	if err != nil {
		return nil, fmt.Errorf("获取表格失败: %v", err)
	}
	if table == nil {
		return nil, fmt.Errorf("表格不存在")
	}

	// 验证变更
	if err := s.ValidateSchemaChange(ctx, table, req.Changes); err != nil {
		return &SchemaChangeResult{
			Success: false,
			Errors:  []string{err.Error()},
		}, nil
	}

	// 应用变更
	var appliedChanges []SchemaChange
	var warnings []string

	for _, change := range req.Changes {
		switch change.Type {
		case SchemaChangeAddField:
			if err := table.AddField(change.NewField); err != nil {
				return &SchemaChangeResult{
					Success: false,
					Errors:  []string{fmt.Sprintf("添加字段失败: %v", err)},
				}, nil
			}
			appliedChanges = append(appliedChanges, change)

		case SchemaChangeUpdateField:
			field := table.GetFieldByID(change.FieldID)
			if field == nil {
				return &SchemaChangeResult{
					Success: false,
					Errors:  []string{"字段不存在"},
				}, nil
			}

			// 应用字段更新
			if change.NewField.Name != field.Name {
				field.Name = change.NewField.Name
			}
			if change.NewField.Type != field.Type {
				if err := field.ChangeType(change.NewField.Type, change.NewField.Options); err != nil {
					return &SchemaChangeResult{
						Success: false,
						Errors:  []string{fmt.Sprintf("更改字段类型失败: %v", err)},
					}, nil
				}
				warnings = append(warnings, "字段类型变更可能影响现有数据")
			}
			if change.NewField.IsRequired != field.IsRequired {
				if err := field.SetRequired(change.NewField.IsRequired); err != nil {
					return &SchemaChangeResult{
						Success: false,
						Errors:  []string{fmt.Sprintf("设置必填属性失败: %v", err)},
					}, nil
				}
			}
			if change.NewField.IsUnique != field.IsUnique {
				if err := field.SetUnique(change.NewField.IsUnique); err != nil {
					return &SchemaChangeResult{
						Success: false,
						Errors:  []string{fmt.Sprintf("设置唯一性约束失败: %v", err)},
					}, nil
				}
			}
			appliedChanges = append(appliedChanges, change)

		case SchemaChangeDeleteField:
			if err := table.RemoveField(change.FieldID); err != nil {
				return &SchemaChangeResult{
					Success: false,
					Errors:  []string{fmt.Sprintf("删除字段失败: %v", err)},
				}, nil
			}
			appliedChanges = append(appliedChanges, change)
			warnings = append(warnings, "删除字段将永久丢失该字段的所有数据")
		}
	}

	// 保存表格
	if err := s.tableRepo.UpdateTable(ctx, table); err != nil {
		return &SchemaChangeResult{
			Success: false,
			Errors:  []string{fmt.Sprintf("保存表格失败: %v", err)},
		}, nil
	}

	return &SchemaChangeResult{
		Success:    true,
		Changes:    appliedChanges,
		Warnings:   warnings,
		NewVersion: table.GetSchemaVersion(),
	}, nil
}

// PreviewSchemaChanges 预览schema变更的影响
func (s *SchemaServiceImpl) PreviewSchemaChanges(ctx context.Context, table *Table, changes []SchemaChange) (*SchemaChangeResult, error) {
	// 创建表格副本进行预览
	tableCopy := *table

	var previewChanges []SchemaChange
	var warnings []string
	var errors []string

	for _, change := range changes {
		switch change.Type {
		case SchemaChangeAddField:
			if err := s.validateAddField(&tableCopy, change.NewField); err != nil {
				errors = append(errors, fmt.Sprintf("添加字段 '%s': %v", change.NewField.Name, err))
			} else {
				previewChanges = append(previewChanges, change)
			}

		case SchemaChangeUpdateField:
			if err := s.validateUpdateField(&tableCopy, change.OldField, change.NewField); err != nil {
				errors = append(errors, fmt.Sprintf("更新字段 '%s': %v", change.NewField.Name, err))
			} else {
				previewChanges = append(previewChanges, change)
				if change.NewField.Type != change.OldField.Type {
					warnings = append(warnings, fmt.Sprintf("字段 '%s' 类型变更可能影响现有数据", change.NewField.Name))
				}
			}

		case SchemaChangeDeleteField:
			if err := s.validateDeleteField(&tableCopy, change.OldField); err != nil {
				errors = append(errors, fmt.Sprintf("删除字段 '%s': %v", change.OldField.Name, err))
			} else {
				previewChanges = append(previewChanges, change)
				warnings = append(warnings, fmt.Sprintf("删除字段 '%s' 将永久丢失该字段的所有数据", change.OldField.Name))
			}
		}
	}

	return &SchemaChangeResult{
		Success:  len(errors) == 0,
		Changes:  previewChanges,
		Errors:   errors,
		Warnings: warnings,
	}, nil
}

// CanSafelyChangeFieldType 检查是否可以安全地更改字段类型
func (s *SchemaServiceImpl) CanSafelyChangeFieldType(ctx context.Context, field *Field, newType FieldType) (bool, []string, error) {
	var warnings []string

	// 检查类型兼容性
	canChange, err := field.CanChangeTypeTo(newType)
	if err != nil {
		return false, nil, err
	}
	if !canChange {
		return false, []string{fmt.Sprintf("字段类型从 %s 到 %s 不兼容", field.Type, newType)}, nil
	}

	// 检查数据兼容性警告
	if !field.Type.IsCompatibleWith(newType) {
		warnings = append(warnings, "类型转换可能导致数据丢失或格式错误")
	}

	// 检查约束兼容性
	if field.IsUnique && !newType.SupportsUnique() {
		warnings = append(warnings, "新类型不支持唯一性约束，该约束将被移除")
	}

	return true, warnings, nil
}

// GetSchemaHistory 获取表格的schema变更历史
func (s *SchemaServiceImpl) GetSchemaHistory(ctx context.Context, tableID string) ([]SchemaChange, error) {
	// TODO: 实现schema变更历史记录
	// 这需要在数据库中存储schema变更日志
	return []SchemaChange{}, nil
}

// validateAddField 验证添加字段
func (s *SchemaServiceImpl) validateAddField(table *Table, field *Field) error {
	if field == nil {
		return fmt.Errorf("字段不能为空")
	}

	if field.Name == "" {
		return fmt.Errorf("字段名称不能为空")
	}

	if table.HasFieldWithName(field.Name) {
		return fmt.Errorf("字段名称 '%s' 已存在", field.Name)
	}

	if field.IsPrimary && table.HasPrimaryField() {
		return fmt.Errorf("表格已存在主键字段")
	}

	// 验证字段类型是否需要选项配置
	if field.Type.RequiresOptions() && (field.Options == nil || len(field.Options.Choices) == 0) {
		return fmt.Errorf("字段类型 %s 需要配置选项", field.Type)
	}

	return nil
}

// validateUpdateField 验证更新字段
func (s *SchemaServiceImpl) validateUpdateField(table *Table, oldField, newField *Field) error {
	if oldField == nil || newField == nil {
		return fmt.Errorf("字段不能为空")
	}

	if newField.Name == "" {
		return fmt.Errorf("字段名称不能为空")
	}

	// 检查名称冲突（除了自己）
	if newField.Name != oldField.Name && table.HasFieldWithName(newField.Name) {
		return fmt.Errorf("字段名称 '%s' 已存在", newField.Name)
	}

	// 检查类型变更
	if newField.Type != oldField.Type {
		canChange, err := oldField.CanChangeTypeTo(newField.Type)
		if err != nil {
			return err
		}
		if !canChange {
			return fmt.Errorf("字段类型从 %s 到 %s 不兼容", oldField.Type, newField.Type)
		}
	}

	// 验证新字段类型的选项配置
	if newField.Type.RequiresOptions() && (newField.Options == nil || len(newField.Options.Choices) == 0) {
		return fmt.Errorf("字段类型 %s 需要配置选项", newField.Type)
	}

	return nil
}

// validateDeleteField 验证删除字段
func (s *SchemaServiceImpl) validateDeleteField(table *Table, field *Field) error {
	if field == nil {
		return fmt.Errorf("字段不能为空")
	}

	canDelete, err := field.CanBeDeleted()
	if err != nil {
		return err
	}
	if !canDelete {
		return fmt.Errorf("字段不能被删除")
	}

	return nil
}
