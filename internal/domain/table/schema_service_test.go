package table

import (
	"context"
	"testing"
)

func TestSchemaService_ValidateSchemaChange(t *testing.T) {
	// 创建测试表格
	table := &Table{
		ID:     "table1",
		Name:   "Test Table",
		fields: make([]*Field, 0),
	}

	// 添加现有字段
	existingField := &Field{
		ID:        "field1",
		Name:      "Name",
		Type:      FieldTypeText,
		TableID:   "table1",
		IsPrimary: true,
	}
	table.fields = append(table.fields, existingField)

	service := &SchemaServiceImpl{}

	// 测试添加新字段
	newField := &Field{
		ID:      "field2",
		Name:    "Email",
		Type:    FieldTypeEmail,
		TableID: "table1",
	}

	addChange := SchemaChange{
		Type:     SchemaChangeAddField,
		NewField: newField,
	}

	err := service.ValidateSchemaChange(context.Background(), table, []SchemaChange{addChange})
	if err != nil {
		t.Errorf("Expected no error for valid add field change, got %v", err)
	}

	// 测试添加重复名称的字段
	duplicateField := &Field{
		ID:      "field3",
		Name:    "Name", // 重复名称
		Type:    FieldTypeText,
		TableID: "table1",
	}

	duplicateChange := SchemaChange{
		Type:     SchemaChangeAddField,
		NewField: duplicateField,
	}

	err = service.ValidateSchemaChange(context.Background(), table, []SchemaChange{duplicateChange})
	if err == nil {
		t.Error("Expected error for duplicate field name, got nil")
	}

	// 测试删除主键字段
	deleteChange := SchemaChange{
		Type:     SchemaChangeDeleteField,
		FieldID:  "field1",
		OldField: existingField,
	}

	err = service.ValidateSchemaChange(context.Background(), table, []SchemaChange{deleteChange})
	if err == nil {
		t.Error("Expected error for deleting primary key field, got nil")
	}
}

func TestSchemaService_PreviewSchemaChanges(t *testing.T) {
	// 创建测试表格
	table := &Table{
		ID:     "table1",
		Name:   "Test Table",
		fields: make([]*Field, 0),
	}

	// 添加现有字段
	existingField := &Field{
		ID:      "field1",
		Name:    "Name",
		Type:    FieldTypeText,
		TableID: "table1",
	}
	table.fields = append(table.fields, existingField)

	service := &SchemaServiceImpl{}

	// 测试预览添加字段
	newField := &Field{
		ID:      "field2",
		Name:    "Email",
		Type:    FieldTypeEmail,
		TableID: "table1",
	}

	addChange := SchemaChange{
		Type:     SchemaChangeAddField,
		NewField: newField,
	}

	result, err := service.PreviewSchemaChanges(context.Background(), table, []SchemaChange{addChange})
	if err != nil {
		t.Errorf("Expected no error for preview, got %v", err)
	}

	if !result.Success {
		t.Error("Expected preview to be successful")
	}

	if len(result.Changes) != 1 {
		t.Errorf("Expected 1 change in preview, got %d", len(result.Changes))
	}

	// 测试预览类型变更
	updatedField := &Field{
		ID:      "field1",
		Name:    "Name",
		Type:    FieldTypeEmail, // 从Text变更为Email
		TableID: "table1",
	}

	updateChange := SchemaChange{
		Type:     SchemaChangeUpdateField,
		FieldID:  "field1",
		OldField: existingField,
		NewField: updatedField,
	}

	result, err = service.PreviewSchemaChanges(context.Background(), table, []SchemaChange{updateChange})
	if err != nil {
		t.Errorf("Expected no error for preview, got %v", err)
	}

	if !result.Success {
		t.Error("Expected preview to be successful")
	}

	if len(result.Warnings) == 0 {
		t.Error("Expected warnings for type change")
	}
}

func TestSchemaService_validateAddField(t *testing.T) {
	table := &Table{
		ID:     "table1",
		Name:   "Test Table",
		fields: make([]*Field, 0),
	}

	service := &SchemaServiceImpl{}

	// 测试添加空字段
	err := service.validateAddField(table, nil)
	if err == nil {
		t.Error("Expected error for nil field, got nil")
	}

	// 测试添加空名称字段
	emptyNameField := &Field{
		ID:      "field1",
		Name:    "",
		Type:    FieldTypeText,
		TableID: "table1",
	}

	err = service.validateAddField(table, emptyNameField)
	if err == nil {
		t.Error("Expected error for empty field name, got nil")
	}

	// 测试添加需要选项配置的字段但没有配置
	selectField := &Field{
		ID:      "field1",
		Name:    "Status",
		Type:    FieldTypeSelect,
		TableID: "table1",
		Options: nil,
	}

	err = service.validateAddField(table, selectField)
	if err == nil {
		t.Error("Expected error for select field without options, got nil")
	}

	// 测试添加有效的选择字段
	validSelectField := &Field{
		ID:      "field1",
		Name:    "Status",
		Type:    FieldTypeSelect,
		TableID: "table1",
		Options: &FieldOptions{
			Choices: []FieldChoice{
				{ID: "1", Label: "Active", Value: "active"},
				{ID: "2", Label: "Inactive", Value: "inactive"},
			},
		},
	}

	err = service.validateAddField(table, validSelectField)
	if err != nil {
		t.Errorf("Expected no error for valid select field, got %v", err)
	}
}

func TestSchemaService_validateUpdateField(t *testing.T) {
	table := &Table{
		ID:     "table1",
		Name:   "Test Table",
		fields: make([]*Field, 0),
	}

	// 添加现有字段
	existingField := &Field{
		ID:      "field1",
		Name:    "Name",
		Type:    FieldTypeText,
		TableID: "table1",
	}
	table.fields = append(table.fields, existingField)

	anotherField := &Field{
		ID:      "field2",
		Name:    "Email",
		Type:    FieldTypeEmail,
		TableID: "table1",
	}
	table.fields = append(table.fields, anotherField)

	service := &SchemaServiceImpl{}

	// 测试更新为已存在的名称
	updatedField := &Field{
		ID:      "field1",
		Name:    "Email", // 与field2重复
		Type:    FieldTypeText,
		TableID: "table1",
	}

	err := service.validateUpdateField(table, existingField, updatedField)
	if err == nil {
		t.Error("Expected error for duplicate field name, got nil")
	}

	// 测试不兼容的类型变更
	incompatibleField := &Field{
		ID:      "field1",
		Name:    "Name",
		Type:    FieldTypeNumber, // Text到Number不兼容
		TableID: "table1",
	}

	err = service.validateUpdateField(table, existingField, incompatibleField)
	if err == nil {
		t.Error("Expected error for incompatible type change, got nil")
	}

	// 测试兼容的类型变更
	compatibleField := &Field{
		ID:      "field1",
		Name:    "Name",
		Type:    FieldTypeEmail, // Text到Email兼容
		TableID: "table1",
	}

	err = service.validateUpdateField(table, existingField, compatibleField)
	if err != nil {
		t.Errorf("Expected no error for compatible type change, got %v", err)
	}
}

func TestSchemaService_validateDeleteField(t *testing.T) {
	service := &SchemaServiceImpl{}

	// 测试删除主键字段
	primaryField := &Field{
		ID:        "field1",
		Name:      "ID",
		Type:      FieldTypeNumber,
		IsPrimary: true,
	}

	err := service.validateDeleteField(nil, primaryField)
	if err == nil {
		t.Error("Expected error for deleting primary key field, got nil")
	}

	// 测试删除系统字段
	systemField := &Field{
		ID:   "field2",
		Name: "Created Time",
		Type: FieldTypeCreatedTime,
	}

	err = service.validateDeleteField(nil, systemField)
	if err == nil {
		t.Error("Expected error for deleting system field, got nil")
	}

	// 测试删除普通字段
	normalField := &Field{
		ID:   "field3",
		Name: "Name",
		Type: FieldTypeText,
	}

	err = service.validateDeleteField(nil, normalField)
	if err != nil {
		t.Errorf("Expected no error for deleting normal field, got %v", err)
	}
}
