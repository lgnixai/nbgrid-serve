package table

import (
	"testing"
	"time"
)

func TestTable_AddField(t *testing.T) {
	table := &Table{
		ID:            "table1",
		Name:          "Test Table",
		fields:        make([]*Field, 0),
		schemaVersion: 1,
	}

	// 测试添加正常字段
	field1 := &Field{
		ID:        "field1",
		TableID:   "table1",
		Name:      "Name",
		Type:      FieldTypeText,
		IsPrimary: true,
	}

	err := table.AddField(field1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(table.fields) != 1 {
		t.Errorf("Expected 1 field, got %d", len(table.fields))
	}

	if table.schemaVersion != 2 {
		t.Errorf("Expected schema version 2, got %d", table.schemaVersion)
	}

	// 测试添加重复名称的字段
	field2 := &Field{
		ID:      "field2",
		TableID: "table1",
		Name:    "Name", // 重复名称
		Type:    FieldTypeText,
	}

	err = table.AddField(field2)
	if err == nil {
		t.Error("Expected error for duplicate field name, got nil")
	}

	// 测试添加重复主键字段
	field3 := &Field{
		ID:        "field3",
		TableID:   "table1",
		Name:      "ID",
		Type:      FieldTypeNumber,
		IsPrimary: true, // 重复主键
	}

	err = table.AddField(field3)
	if err == nil {
		t.Error("Expected error for duplicate primary key, got nil")
	}
}

func TestTable_ValidateSchema(t *testing.T) {
	table := &Table{
		ID:     "table1",
		Name:   "Test Table",
		fields: make([]*Field, 0),
	}

	// 测试空字段列表
	err := table.ValidateSchema()
	if err == nil {
		t.Error("Expected error for empty fields, got nil")
	}

	// 添加非主键字段
	field1 := &Field{
		ID:      "field1",
		Name:    "Name",
		Type:    FieldTypeText,
		TableID: "table1",
	}
	table.fields = append(table.fields, field1)

	// 测试没有主键字段
	err = table.ValidateSchema()
	if err == nil {
		t.Error("Expected error for no primary key, got nil")
	}

	// 添加主键字段
	field2 := &Field{
		ID:        "field2",
		Name:      "ID",
		Type:      FieldTypeNumber,
		TableID:   "table1",
		IsPrimary: true,
	}
	table.fields = append(table.fields, field2)

	// 测试有效schema
	err = table.ValidateSchema()
	if err != nil {
		t.Errorf("Expected no error for valid schema, got %v", err)
	}
}

func TestField_ValidateValue(t *testing.T) {
	// 测试文本字段
	textField := &Field{
		Name:       "Name",
		Type:       FieldTypeText,
		IsRequired: true,
		Options: &FieldOptions{
			MinLength: 2,
			MaxLength: 10,
		},
	}

	// 测试必填字段为空
	err := textField.ValidateValue("")
	if err == nil {
		t.Error("Expected error for empty required field, got nil")
	}

	// 测试长度验证
	err = textField.ValidateValue("A")
	if err == nil {
		t.Error("Expected error for too short text, got nil")
	}

	err = textField.ValidateValue("This is too long")
	if err == nil {
		t.Error("Expected error for too long text, got nil")
	}

	// 测试有效值
	err = textField.ValidateValue("Valid")
	if err != nil {
		t.Errorf("Expected no error for valid text, got %v", err)
	}
}

func TestField_CanChangeTypeTo(t *testing.T) {
	// 测试普通字段类型转换
	field := &Field{
		Type:      FieldTypeText,
		IsPrimary: false,
	}

	canChange, err := field.CanChangeTypeTo(FieldTypeEmail)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !canChange {
		t.Error("Expected text to email conversion to be allowed")
	}

	// 测试主键字段类型转换
	primaryField := &Field{
		Type:      FieldTypeNumber,
		IsPrimary: true,
	}

	canChange, err = primaryField.CanChangeTypeTo(FieldTypeText)
	if err == nil {
		t.Error("Expected error for primary key type change, got nil")
	}
	if canChange {
		t.Error("Expected primary key type change to be disallowed")
	}
}

func TestFieldType_IsCompatibleWith(t *testing.T) {
	// 测试兼容的类型转换
	if !FieldTypeText.IsCompatibleWith(FieldTypeEmail) {
		t.Error("Expected text to email to be compatible")
	}

	if !FieldTypeNumber.IsCompatibleWith(FieldTypeCurrency) {
		t.Error("Expected number to currency to be compatible")
	}

	// 测试不兼容的类型转换
	if FieldTypeText.IsCompatibleWith(FieldTypeNumber) {
		t.Error("Expected text to number to be incompatible")
	}

	// 测试相同类型
	if !FieldTypeText.IsCompatibleWith(FieldTypeText) {
		t.Error("Expected same type to be compatible")
	}
}

func TestFieldType_SupportsUnique(t *testing.T) {
	// 测试支持唯一性约束的类型
	if !FieldTypeText.SupportsUnique() {
		t.Error("Expected text type to support unique constraint")
	}

	if !FieldTypeEmail.SupportsUnique() {
		t.Error("Expected email type to support unique constraint")
	}

	// 测试不支持唯一性约束的类型
	if FieldTypeMultiSelect.SupportsUnique() {
		t.Error("Expected multi-select type to not support unique constraint")
	}

	if FieldTypeFile.SupportsUnique() {
		t.Error("Expected file type to not support unique constraint")
	}
}

func TestField_SetRequired(t *testing.T) {
	field := &Field{
		Name:         "TestField",
		Type:         FieldTypeText,
		IsRequired:   false,
		DefaultValue: nil,
	}

	// 测试设置为必填但没有默认值
	err := field.SetRequired(true)
	if err == nil {
		t.Error("Expected error when setting required without default value, got nil")
	}

	// 设置默认值后再设置为必填
	defaultValue := "default"
	field.DefaultValue = &defaultValue

	err = field.SetRequired(true)
	if err != nil {
		t.Errorf("Expected no error when setting required with default value, got %v", err)
	}

	if !field.IsRequired {
		t.Error("Expected field to be required")
	}
}

func TestField_SetUnique(t *testing.T) {
	// 测试支持唯一性约束的字段类型
	field := &Field{
		Name:     "TestField",
		Type:     FieldTypeText,
		IsUnique: false,
	}

	err := field.SetUnique(true)
	if err != nil {
		t.Errorf("Expected no error for text field unique constraint, got %v", err)
	}

	if !field.IsUnique {
		t.Error("Expected field to be unique")
	}

	// 测试不支持唯一性约束的字段类型
	multiSelectField := &Field{
		Name:     "TestField",
		Type:     FieldTypeMultiSelect,
		IsUnique: false,
	}

	err = multiSelectField.SetUnique(true)
	if err == nil {
		t.Error("Expected error for multi-select field unique constraint, got nil")
	}
}

func TestField_incrementVersion(t *testing.T) {
	field := &Field{
		Version: 1,
	}

	initialVersion := field.Version
	field.incrementVersion()

	if field.Version != initialVersion+1 {
		t.Errorf("Expected version to increment from %d to %d, got %d", initialVersion, initialVersion+1, field.Version)
	}

	if field.LastModifiedTime == nil {
		t.Error("Expected LastModifiedTime to be set")
	}

	// 验证时间是最近的
	timeDiff := time.Since(*field.LastModifiedTime)
	if timeDiff > time.Second {
		t.Errorf("Expected LastModifiedTime to be recent, but it was %v ago", timeDiff)
	}
}