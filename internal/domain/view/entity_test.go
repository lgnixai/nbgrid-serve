package view

import (
	"testing"
)

func TestNewView(t *testing.T) {
	req := CreateViewRequest{
		TableID:     "tbl_test123",
		Name:        "测试视图",
		Description: stringPtr("这是一个测试视图"),
		Type:        string(ViewTypeGrid),
		Config:      map[string]interface{}{"page_size": 50},
		IsDefault:   true,
		CreatedBy:   "usr_test123",
	}

	view := NewView(req)

	if view.ID == "" {
		t.Error("视图ID不能为空")
	}
	if view.TableID != req.TableID {
		t.Errorf("期望TableID为%s，实际为%s", req.TableID, view.TableID)
	}
	if view.Name != req.Name {
		t.Errorf("期望Name为%s，实际为%s", req.Name, view.Name)
	}
	if view.Type != ViewTypeGrid {
		t.Errorf("期望Type为%s，实际为%s", ViewTypeGrid, view.Type)
	}
	if view.Version != 1 {
		t.Errorf("期望Version为1，实际为%d", view.Version)
	}
	if view.IsPublic {
		t.Error("新视图不应该是公共视图")
	}
}

func TestViewTypeValidation(t *testing.T) {
	tests := []struct {
		viewType string
		valid    bool
	}{
		{"grid", true},
		{"kanban", true},
		{"calendar", true},
		{"gallery", true},
		{"form", true},
		{"chart", true},
		{"timeline", true},
		{"invalid", false},
		{"", false},
	}

	for _, test := range tests {
		t.Run(test.viewType, func(t *testing.T) {
			result := IsValidViewType(test.viewType)
			if result != test.valid {
				t.Errorf("视图类型%s的验证结果期望为%v，实际为%v", test.viewType, test.valid, result)
			}
		})
	}
}

func TestGridViewConfig(t *testing.T) {
	config := &GridViewConfig{
		BaseViewConfig: BaseViewConfig{
			Type: ViewTypeGrid,
		},
		PageSize:  100,
		RowHeight: 40,
	}

	// 测试验证
	if err := config.Validate(); err != nil {
		t.Errorf("网格视图配置验证失败: %v", err)
	}

	// 测试无效配置
	invalidConfig := &GridViewConfig{
		BaseViewConfig: BaseViewConfig{
			Type: ViewTypeGrid,
		},
		PageSize: 2000, // 超过限制
	}

	if err := invalidConfig.Validate(); err == nil {
		t.Error("期望验证失败，但验证通过了")
	}

	// 测试默认值
	defaultConfig := &GridViewConfig{
		BaseViewConfig: BaseViewConfig{
			Type: ViewTypeGrid,
		},
	}

	if err := defaultConfig.Validate(); err != nil {
		t.Errorf("默认配置验证失败: %v", err)
	}

	if defaultConfig.PageSize != 50 {
		t.Errorf("期望默认PageSize为50，实际为%d", defaultConfig.PageSize)
	}
}

func TestKanbanViewConfig(t *testing.T) {
	config := &KanbanViewConfig{
		BaseViewConfig: BaseViewConfig{
			Type: ViewTypeKanban,
		},
		GroupFieldID: "fld_status",
		CardHeight:   150,
		ShowCount:    true,
		AllowDrag:    true,
	}

	// 测试验证
	if err := config.Validate(); err != nil {
		t.Errorf("看板视图配置验证失败: %v", err)
	}

	// 测试缺少分组字段
	invalidConfig := &KanbanViewConfig{
		BaseViewConfig: BaseViewConfig{
			Type: ViewTypeKanban,
		},
		// 缺少GroupFieldID
	}

	if err := invalidConfig.Validate(); err == nil {
		t.Error("期望验证失败，但验证通过了")
	}
}

func TestCalendarViewConfig(t *testing.T) {
	config := &CalendarViewConfig{
		BaseViewConfig: BaseViewConfig{
			Type: ViewTypeCalendar,
		},
		DateFieldID:  "fld_date",
		TitleFieldID: "fld_title",
		DefaultView:  "month",
	}

	// 测试验证
	if err := config.Validate(); err != nil {
		t.Errorf("日历视图配置验证失败: %v", err)
	}

	// 测试缺少日期字段
	invalidConfig := &CalendarViewConfig{
		BaseViewConfig: BaseViewConfig{
			Type: ViewTypeCalendar,
		},
		// 缺少DateFieldID
	}

	if err := invalidConfig.Validate(); err == nil {
		t.Error("期望验证失败，但验证通过了")
	}

	// 测试无效的默认视图
	invalidViewConfig := &CalendarViewConfig{
		BaseViewConfig: BaseViewConfig{
			Type: ViewTypeCalendar,
		},
		DateFieldID: "fld_date",
		DefaultView: "invalid",
	}

	if err := invalidViewConfig.Validate(); err == nil {
		t.Error("期望验证失败，但验证通过了")
	}
}

func TestViewUpdate(t *testing.T) {
	view := NewView(CreateViewRequest{
		TableID:   "tbl_test123",
		Name:      "原始视图",
		Type:      string(ViewTypeGrid),
		IsDefault: false,
		CreatedBy: "usr_test123",
	})

	originalVersion := view.Version

	// 更新视图
	updateReq := UpdateViewRequest{
		Name:        stringPtr("更新后的视图"),
		Description: stringPtr("更新后的描述"),
		IsDefault:   boolPtr(true),
	}

	if err := view.Update(updateReq); err != nil {
		t.Errorf("更新视图失败: %v", err)
	}

	// 验证更新结果
	if view.Version <= originalVersion {
		t.Error("版本号应该递增")
	}
	if view.Name != "更新后的视图" {
		t.Error("名称更新不正确")
	}
	if view.Description == nil || *view.Description != "更新后的描述" {
		t.Error("描述更新不正确")
	}
	if !view.IsDefault {
		t.Error("默认状态更新不正确")
	}
}

func TestViewSharing(t *testing.T) {
	view := NewView(CreateViewRequest{
		TableID:   "tbl_test123",
		Name:      "测试视图",
		Type:      string(ViewTypeGrid),
		CreatedBy: "usr_test123",
	})

	// 测试初始状态
	if view.IsPublic {
		t.Error("新视图不应该是公共视图")
	}
	if view.ShareToken != nil {
		t.Error("新视图不应该有分享令牌")
	}

	// 测试设置为公共视图
	view.SetPublic(true)

	if !view.IsPublic {
		t.Error("视图应该被设置为公共视图")
	}
	if view.ShareToken == nil {
		t.Error("公共视图应该有分享令牌")
	}

	// 测试生成新的分享令牌
	oldToken := *view.ShareToken
	newToken := view.GenerateShareToken()

	if newToken == oldToken {
		t.Error("新生成的分享令牌应该不同")
	}
	if view.ShareToken == nil || *view.ShareToken != newToken {
		t.Error("分享令牌更新不正确")
	}

	// 测试撤销分享
	view.RevokeShareToken()

	if view.IsPublic {
		t.Error("撤销分享后视图不应该是公共视图")
	}
	if view.ShareToken != nil {
		t.Error("撤销分享后不应该有分享令牌")
	}
}

func TestViewClone(t *testing.T) {
	original := NewView(CreateViewRequest{
		TableID:     "tbl_test123",
		Name:        "原始视图",
		Description: stringPtr("原始描述"),
		Type:        string(ViewTypeGrid),
		Config:      map[string]interface{}{"page_size": 50},
		IsDefault:   true,
		CreatedBy:   "usr_test123",
	})

	cloned := original.Clone("克隆视图", "usr_test456")

	// 验证克隆结果
	if cloned.ID == original.ID {
		t.Error("克隆视图的ID应该不同")
	}
	if cloned.TableID != original.TableID {
		t.Error("克隆视图的TableID应该相同")
	}
	if cloned.Name != "克隆视图" {
		t.Error("克隆视图的名称应该是指定的名称")
	}
	if cloned.Type != original.Type {
		t.Error("克隆视图的类型应该相同")
	}
	if cloned.IsDefault {
		t.Error("克隆视图不应该是默认视图")
	}
	if cloned.IsPublic {
		t.Error("克隆视图不应该是公共视图")
	}
	if cloned.ShareToken != nil {
		t.Error("克隆视图不应该有分享令牌")
	}
	if cloned.CreatedBy != "usr_test456" {
		t.Error("克隆视图的创建者应该是指定的用户")
	}

	// 验证配置深拷贝
	cloned.Config["page_size"] = 100
	if original.Config["page_size"] == 100 {
		t.Error("修改克隆视图配置不应该影响原视图")
	}
}

func TestViewSoftDelete(t *testing.T) {
	view := NewView(CreateViewRequest{
		TableID:   "tbl_test123",
		Name:      "测试视图",
		Type:      string(ViewTypeGrid),
		CreatedBy: "usr_test123",
	})

	if view.IsDeleted() {
		t.Error("新视图不应该被标记为已删除")
	}

	view.SoftDelete()

	if !view.IsDeleted() {
		t.Error("视图应该被标记为已删除")
	}
	if view.DeletedTime == nil {
		t.Error("删除时间不能为空")
	}

	// 测试恢复
	view.Restore()

	if view.IsDeleted() {
		t.Error("恢复后视图不应该被标记为已删除")
	}
	if view.DeletedTime != nil {
		t.Error("恢复后删除时间应该为空")
	}
}

func TestGetViewTypeInfo(t *testing.T) {
	info := GetViewTypeInfo(ViewTypeGrid)

	if info["name"] == nil {
		t.Error("视图类型信息应该包含名称")
	}
	if info["description"] == nil {
		t.Error("视图类型信息应该包含描述")
	}
	if info["features"] == nil {
		t.Error("视图类型信息应该包含功能列表")
	}

	// 测试未知类型
	unknownInfo := GetViewTypeInfo("unknown")
	if unknownInfo["name"] != "未知视图" {
		t.Error("未知类型应该返回默认信息")
	}
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
