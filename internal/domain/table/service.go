package table

import (
	"context"

	"teable-go-backend/pkg/errors"
)

// Service 数据表服务接口
type Service interface {
	// 表格管理
	CreateTable(ctx context.Context, req CreateTableRequest) (*Table, error)
	GetTable(ctx context.Context, id string) (*Table, error)
	UpdateTable(ctx context.Context, id string, req UpdateTableRequest) (*Table, error)
	DeleteTable(ctx context.Context, id string) error
	ListTables(ctx context.Context, filter ListTableFilter) ([]*Table, int64, error)

	// 字段管理
	CreateField(ctx context.Context, req CreateFieldRequest) (*Field, error)
	GetField(ctx context.Context, id string) (*Field, error)
	UpdateField(ctx context.Context, id string, req UpdateFieldRequest) (*Field, error)
	DeleteField(ctx context.Context, id string) error
	ListFields(ctx context.Context, filter ListFieldFilter) ([]*Field, int64, error)

	// 批量操作
	BulkUpdateTables(ctx context.Context, updates []BulkUpdateTableRequest) error
	BulkDeleteTables(ctx context.Context, tableIDs []string) error
	BulkUpdateFields(ctx context.Context, updates []BulkUpdateFieldRequest) error
	BulkDeleteFields(ctx context.Context, fieldIDs []string) error

	// 权限检查
	CheckUserPermission(ctx context.Context, tableID, userID, permission string) (bool, error)

	// 统计信息
	GetTableStats(ctx context.Context, tableID string) (*TableStats, error)
	GetBaseTableStats(ctx context.Context, baseID string) (*BaseTableStats, error)

	// 导出/导入
	ExportTables(ctx context.Context, filter ListTableFilter) ([]*Table, error)
	ImportTables(ctx context.Context, tables []CreateTableRequest) ([]*Table, error)
	ExportFields(ctx context.Context, filter ListFieldFilter) ([]*Field, error)
	ImportFields(ctx context.Context, fields []CreateFieldRequest) ([]*Field, error)

	// 字段类型和验证
	GetFieldTypes(ctx context.Context) ([]FieldTypeInfo, error)
	ValidateFieldValue(ctx context.Context, field *Field, value interface{}) error
	GetFieldTypeInfo(ctx context.Context, fieldType FieldType) (FieldTypeInfo, error)
	
	// Schema管理
	ValidateSchemaChange(ctx context.Context, tableID string, changes []SchemaChange) error
	ApplySchemaChanges(ctx context.Context, req SchemaChangeRequest) (*SchemaChangeResult, error)
	PreviewSchemaChanges(ctx context.Context, tableID string, changes []SchemaChange) (*SchemaChangeResult, error)
}

// ServiceImpl 数据表服务实现
type ServiceImpl struct {
	repo          Repository
	schemaService SchemaService
}

// NewService 创建数据表服务
func NewService(repo Repository) Service {
	service := &ServiceImpl{repo: repo}
	service.schemaService = NewSchemaService(repo)
	return service
}

// CreateTable 创建数据表
func (s *ServiceImpl) CreateTable(ctx context.Context, req CreateTableRequest) (*Table, error) {
	// 检查名称是否已存在于同一基础表下
	exists, err := s.repo.ExistsTable(ctx, ListTableFilter{BaseID: &req.BaseID, Name: &req.Name})
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.ErrResourceExists.WithDetails("数据表名称已存在于此基础表")
	}

	table := NewTable(req)
	if err := s.repo.CreateTable(ctx, table); err != nil {
		return nil, err
	}
	return table, nil
}

// GetTable 获取数据表
func (s *ServiceImpl) GetTable(ctx context.Context, id string) (*Table, error) {
	table, err := s.repo.GetTableByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if table == nil {
		return nil, errors.ErrNotFound.WithDetails("数据表未找到")
	}
	return table, nil
}

// UpdateTable 更新数据表
func (s *ServiceImpl) UpdateTable(ctx context.Context, id string, req UpdateTableRequest) (*Table, error) {
	table, err := s.repo.GetTableByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if table == nil {
		return nil, errors.ErrNotFound.WithDetails("数据表未找到")
	}

	// 检查更新后的名称是否冲突
	if req.Name != nil && *req.Name != table.Name {
		exists, err := s.repo.ExistsTable(ctx, ListTableFilter{BaseID: &table.BaseID, Name: req.Name})
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.ErrResourceExists.WithDetails("更新后的数据表名称已存在于此基础表")
		}
	}

	table.Update(req)
	if err := s.repo.UpdateTable(ctx, table); err != nil {
		return nil, err
	}
	return table, nil
}

// DeleteTable 删除数据表
func (s *ServiceImpl) DeleteTable(ctx context.Context, id string) error {
	table, err := s.repo.GetTableByID(ctx, id)
	if err != nil {
		return err
	}
	if table == nil {
		return errors.ErrNotFound.WithDetails("数据表未找到")
	}

	table.SoftDelete() // 软删除
	if err := s.repo.UpdateTable(ctx, table); err != nil {
		return err
	}
	return nil
}

// ListTables 列出数据表
func (s *ServiceImpl) ListTables(ctx context.Context, filter ListTableFilter) ([]*Table, int64, error) {
	tables, err := s.repo.ListTables(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.CountTables(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	return tables, total, nil
}

// CreateField 创建字段
func (s *ServiceImpl) CreateField(ctx context.Context, req CreateFieldRequest) (*Field, error) {
	// 检查名称是否已存在于同一数据表下
	exists, err := s.repo.ExistsField(ctx, ListFieldFilter{TableID: &req.TableID, Name: &req.Name})
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.ErrResourceExists.WithDetails("字段名称已存在于此数据表")
	}

	field := NewField(req)
	if err := s.repo.CreateField(ctx, field); err != nil {
		return nil, err
	}
	return field, nil
}

// GetField 获取字段
func (s *ServiceImpl) GetField(ctx context.Context, id string) (*Field, error) {
	field, err := s.repo.GetFieldByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if field == nil {
		return nil, errors.ErrNotFound.WithDetails("字段未找到")
	}
	return field, nil
}

// UpdateField 更新字段
func (s *ServiceImpl) UpdateField(ctx context.Context, id string, req UpdateFieldRequest) (*Field, error) {
	field, err := s.repo.GetFieldByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if field == nil {
		return nil, errors.ErrNotFound.WithDetails("字段未找到")
	}

	// 检查更新后的名称是否冲突
	if req.Name != nil && *req.Name != field.Name {
		exists, err := s.repo.ExistsField(ctx, ListFieldFilter{TableID: &field.TableID, Name: req.Name})
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.ErrResourceExists.WithDetails("更新后的字段名称已存在于此数据表")
		}
	}

	// 使用新的Update方法，包含验证逻辑
	if err := field.Update(req); err != nil {
		return nil, errors.ErrValidationFailed.WithDetails(err.Error())
	}
	
	if err := s.repo.UpdateField(ctx, field); err != nil {
		return nil, err
	}
	return field, nil
}

// DeleteField 删除字段
func (s *ServiceImpl) DeleteField(ctx context.Context, id string) error {
	field, err := s.repo.GetFieldByID(ctx, id)
	if err != nil {
		return err
	}
	if field == nil {
		return errors.ErrNotFound.WithDetails("字段未找到")
	}

	field.SoftDelete() // 软删除
	if err := s.repo.UpdateField(ctx, field); err != nil {
		return err
	}
	return nil
}

// ListFields 列出字段
func (s *ServiceImpl) ListFields(ctx context.Context, filter ListFieldFilter) ([]*Field, int64, error) {
	fields, err := s.repo.ListFields(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.CountFields(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	return fields, total, nil
}

// BulkUpdateTables 批量更新表格
func (s *ServiceImpl) BulkUpdateTables(ctx context.Context, updates []BulkUpdateTableRequest) error {
	for _, update := range updates {
		_, err := s.UpdateTable(ctx, update.TableID, update.Updates)
		if err != nil {
			return err
		}
	}
	return nil
}

// BulkDeleteTables 批量删除表格
func (s *ServiceImpl) BulkDeleteTables(ctx context.Context, tableIDs []string) error {
	for _, tableID := range tableIDs {
		if err := s.DeleteTable(ctx, tableID); err != nil {
			return err
		}
	}
	return nil
}

// BulkUpdateFields 批量更新字段
func (s *ServiceImpl) BulkUpdateFields(ctx context.Context, updates []BulkUpdateFieldRequest) error {
	for _, update := range updates {
		_, err := s.UpdateField(ctx, update.FieldID, update.Updates)
		if err != nil {
			return err
		}
	}
	return nil
}

// BulkDeleteFields 批量删除字段
func (s *ServiceImpl) BulkDeleteFields(ctx context.Context, fieldIDs []string) error {
	for _, fieldID := range fieldIDs {
		if err := s.DeleteField(ctx, fieldID); err != nil {
			return err
		}
	}
	return nil
}

// CheckUserPermission 检查用户权限
func (s *ServiceImpl) CheckUserPermission(ctx context.Context, tableID, userID, permission string) (bool, error) {
	// 获取表格信息
	table, err := s.GetTable(ctx, tableID)
	if err != nil {
		return false, err
	}

	// 如果是表格创建者，拥有所有权限
	if table.CreatedBy == userID {
		return true, nil
	}

	// TODO: 检查基础表权限
	// 这里需要集成基础表服务来检查用户对基础表的权限
	// 暂时返回false，需要后续集成
	return false, nil
}

// GetTableStats 获取表格统计信息
func (s *ServiceImpl) GetTableStats(ctx context.Context, tableID string) (*TableStats, error) {
	// TODO: 实现获取表格统计信息的逻辑
	// 这里需要查询字段、记录和视图的数量
	return &TableStats{
		TableID: tableID,
		// 暂时返回默认值，需要集成其他服务
		TotalFields:  0,
		TotalRecords: 0,
		TotalViews:   0,
	}, nil
}

// GetBaseTableStats 获取基础表表格统计信息
func (s *ServiceImpl) GetBaseTableStats(ctx context.Context, baseID string) (*BaseTableStats, error) {
	// TODO: 实现获取基础表表格统计信息的逻辑
	return &BaseTableStats{
		BaseID: baseID,
		// 暂时返回默认值，需要集成其他服务
		TotalTables:  0,
		TotalFields:  0,
		TotalRecords: 0,
		TotalViews:   0,
	}, nil
}

// ExportTables 导出表格
func (s *ServiceImpl) ExportTables(ctx context.Context, filter ListTableFilter) ([]*Table, error) {
	tables, _, err := s.ListTables(ctx, filter)
	return tables, err
}

// ImportTables 导入表格
func (s *ServiceImpl) ImportTables(ctx context.Context, tables []CreateTableRequest) ([]*Table, error) {
	var result []*Table
	for _, tableReq := range tables {
		table, err := s.CreateTable(ctx, tableReq)
		if err != nil {
			return nil, err
		}
		result = append(result, table)
	}
	return result, nil
}

// ExportFields 导出字段
func (s *ServiceImpl) ExportFields(ctx context.Context, filter ListFieldFilter) ([]*Field, error) {
	fields, _, err := s.ListFields(ctx, filter)
	return fields, err
}

// ImportFields 导入字段
func (s *ServiceImpl) ImportFields(ctx context.Context, fields []CreateFieldRequest) ([]*Field, error) {
	var result []*Field
	for _, fieldReq := range fields {
		field, err := s.CreateField(ctx, fieldReq)
		if err != nil {
			return nil, err
		}
		result = append(result, field)
	}
	return result, nil
}

// GetFieldTypes 获取所有字段类型
func (s *ServiceImpl) GetFieldTypes(ctx context.Context) ([]FieldTypeInfo, error) {
	return GetAllFieldTypes(), nil
}

// ValidateFieldValue 验证字段值
func (s *ServiceImpl) ValidateFieldValue(ctx context.Context, field *Field, value interface{}) error {
	return ValidateFieldValue(field, value)
}

// GetFieldTypeInfo 获取字段类型信息
func (s *ServiceImpl) GetFieldTypeInfo(ctx context.Context, fieldType FieldType) (FieldTypeInfo, error) {
	return GetFieldTypeInfo(fieldType), nil
}
// ValidateSchemaChange 验证schema变更
func (s *ServiceImpl) ValidateSchemaChange(ctx context.Context, tableID string, changes []SchemaChange) error {
	table, err := s.GetTable(ctx, tableID)
	if err != nil {
		return err
	}
	
	return s.schemaService.ValidateSchemaChange(ctx, table, changes)
}

// ApplySchemaChanges 应用schema变更
func (s *ServiceImpl) ApplySchemaChanges(ctx context.Context, req SchemaChangeRequest) (*SchemaChangeResult, error) {
	return s.schemaService.ApplySchemaChanges(ctx, req)
}

// PreviewSchemaChanges 预览schema变更
func (s *ServiceImpl) PreviewSchemaChanges(ctx context.Context, tableID string, changes []SchemaChange) (*SchemaChangeResult, error) {
	table, err := s.GetTable(ctx, tableID)
	if err != nil {
		return nil, err
	}
	
	return s.schemaService.PreviewSchemaChanges(ctx, table, changes)
}