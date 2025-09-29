package record

import (
	"testing"

	"teable-go-backend/internal/domain/table"
)

func TestNewRecord(t *testing.T) {
	req := CreateRecordRequest{
		TableID:   "tbl_test123",
		Data:      map[string]interface{}{"name": "测试记录", "age": 25},
		CreatedBy: "usr_test123",
	}

	record := NewRecord(req)

	if record.ID == "" {
		t.Error("记录ID不能为空")
	}
	if record.TableID != req.TableID {
		t.Errorf("期望TableID为%s，实际为%s", req.TableID, record.TableID)
	}
	if record.CreatedBy != req.CreatedBy {
		t.Errorf("期望CreatedBy为%s，实际为%s", req.CreatedBy, record.CreatedBy)
	}
	if record.Version != 1 {
		t.Errorf("期望Version为1，实际为%d", record.Version)
	}
	if record.Hash == "" {
		t.Error("记录哈希值不能为空")
	}
}

func TestRecordValidation(t *testing.T) {
	// 创建测试表格schema
	testTable := &table.Table{
		ID: "tbl_test123",
	}

	// 创建测试字段
	nameField := &table.Field{
		ID:         "fld_name",
		Name:       "name",
		Type:       table.FieldTypeText,
		IsRequired: true,
	}

	ageField := &table.Field{
		ID:         "fld_age",
		Name:       "age",
		Type:       table.FieldTypeNumber,
		IsRequired: false,
	}

	testTable.SetFields([]*table.Field{nameField, ageField})

	// 测试有效数据
	t.Run("有效数据", func(t *testing.T) {
		record := NewRecord(CreateRecordRequest{
			TableID:   "tbl_test123",
			Data:      map[string]interface{}{"name": "张三", "age": 25},
			CreatedBy: "usr_test123",
		})

		record.SetTableSchema(testTable)

		if err := record.ValidateData(); err != nil {
			t.Errorf("有效数据验证失败: %v", err)
		}
	})

	// 测试缺少必填字段
	t.Run("缺少必填字段", func(t *testing.T) {
		record := NewRecord(CreateRecordRequest{
			TableID:   "tbl_test123",
			Data:      map[string]interface{}{"age": 25}, // 缺少必填的name字段
			CreatedBy: "usr_test123",
		})

		record.SetTableSchema(testTable)

		if err := record.ValidateData(); err == nil {
			t.Error("期望验证失败，但验证通过了")
		}

		errors := record.GetValidationErrors()
		if len(errors) == 0 {
			t.Error("期望有验证错误，但没有错误")
		}
	})

	// 测试未知字段
	t.Run("未知字段", func(t *testing.T) {
		record := NewRecord(CreateRecordRequest{
			TableID:   "tbl_test123",
			Data:      map[string]interface{}{"name": "张三", "unknown_field": "值"},
			CreatedBy: "usr_test123",
		})

		record.SetTableSchema(testTable)

		if err := record.ValidateData(); err == nil {
			t.Error("期望验证失败，但验证通过了")
		}

		errors := record.GetValidationErrors()
		found := false
		for _, err := range errors {
			if err.Code == "UNKNOWN_FIELD" {
				found = true
				break
			}
		}
		if !found {
			t.Error("期望有未知字段错误，但没有找到")
		}
	})
}

func TestRecordUpdate(t *testing.T) {
	// 创建测试表格schema
	testTable := &table.Table{
		ID: "tbl_test123",
	}

	nameField := &table.Field{
		ID:         "fld_name",
		Name:       "name",
		Type:       table.FieldTypeText,
		IsRequired: true,
	}

	testTable.SetFields([]*table.Field{nameField})

	record := NewRecord(CreateRecordRequest{
		TableID:   "tbl_test123",
		Data:      map[string]interface{}{"name": "张三"},
		CreatedBy: "usr_test123",
	})

	record.SetTableSchema(testTable)

	originalVersion := record.Version
	originalHash := record.Hash

	// 更新记录
	updateReq := UpdateRecordRequest{
		Data:      map[string]interface{}{"name": "李四"},
		UpdatedBy: "usr_test456",
	}

	if err := record.Update(updateReq, "usr_test456"); err != nil {
		t.Errorf("更新记录失败: %v", err)
	}

	// 验证更新结果
	if record.Version <= originalVersion {
		t.Error("版本号应该递增")
	}
	if record.Hash == originalHash {
		t.Error("哈希值应该改变")
	}
	if record.UpdatedBy == nil || *record.UpdatedBy != "usr_test456" {
		t.Error("更新者信息不正确")
	}
	if record.Data["name"] != "李四" {
		t.Error("数据更新不正确")
	}
}

func TestRecordSoftDelete(t *testing.T) {
	record := NewRecord(CreateRecordRequest{
		TableID:   "tbl_test123",
		Data:      map[string]interface{}{"name": "张三"},
		CreatedBy: "usr_test123",
	})

	if record.IsDeleted() {
		t.Error("新记录不应该被标记为已删除")
	}

	record.SoftDelete()

	if !record.IsDeleted() {
		t.Error("记录应该被标记为已删除")
	}
	if record.DeletedTime == nil {
		t.Error("删除时间不能为空")
	}

	// 测试恢复
	record.Restore()

	if record.IsDeleted() {
		t.Error("恢复后记录不应该被标记为已删除")
	}
	if record.DeletedTime != nil {
		t.Error("恢复后删除时间应该为空")
	}
}

func TestRecordClone(t *testing.T) {
	original := NewRecord(CreateRecordRequest{
		TableID:   "tbl_test123",
		Data:      map[string]interface{}{"name": "张三", "age": 25},
		CreatedBy: "usr_test123",
	})

	cloned := original.Clone()

	// 验证克隆结果
	if cloned.ID == original.ID {
		t.Error("克隆记录的ID应该不同")
	}
	if cloned.TableID != original.TableID {
		t.Error("克隆记录的TableID应该相同")
	}
	if cloned.Version != original.Version {
		t.Error("克隆记录的版本应该相同")
	}
	if cloned.Hash != original.Hash {
		t.Error("克隆记录的哈希值应该相同")
	}

	// 验证数据深拷贝
	cloned.Data["name"] = "李四"
	if original.Data["name"] == "李四" {
		t.Error("修改克隆记录不应该影响原记录")
	}
}

func TestRecordFieldOperations(t *testing.T) {
	// 创建测试表格schema
	testTable := &table.Table{
		ID: "tbl_test123",
	}

	nameField := &table.Field{
		ID:         "fld_name",
		Name:       "name",
		Type:       table.FieldTypeText,
		IsRequired: true,
	}

	testTable.SetFields([]*table.Field{nameField})

	record := NewRecord(CreateRecordRequest{
		TableID:   "tbl_test123",
		Data:      map[string]interface{}{"name": "张三"},
		CreatedBy: "usr_test123",
	})

	record.SetTableSchema(testTable)

	// 测试获取字段值
	value, exists := record.GetFieldValue("name")
	if !exists {
		t.Error("字段name应该存在")
	}
	if value != "张三" {
		t.Errorf("期望字段值为'张三'，实际为'%v'", value)
	}

	// 测试更新单个字段
	if err := record.UpdateField("name", "李四", "usr_test456"); err != nil {
		t.Errorf("更新字段失败: %v", err)
	}

	value, _ = record.GetFieldValue("name")
	if value != "李四" {
		t.Errorf("期望字段值为'李四'，实际为'%v'", value)
	}

	// 测试更新不存在的字段
	if err := record.UpdateField("unknown", "值", "usr_test456"); err == nil {
		t.Error("更新不存在的字段应该失败")
	}
}
