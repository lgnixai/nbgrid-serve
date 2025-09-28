package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"teable-go-backend/internal/domain/table"
	"teable-go-backend/internal/infrastructure/database/models"
	"teable-go-backend/pkg/errors"
)

// TableRepository 数据表仓储实现
type TableRepository struct {
	db *gorm.DB
}

// NewTableRepository 创建新的数据表仓储
func NewTableRepository(db *gorm.DB) table.Repository {
	return &TableRepository{db: db}
}

// CreateTable 创建数据表
func (r *TableRepository) CreateTable(ctx context.Context, t *table.Table) error {
	model := r.domainToModel(t)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return r.handleDBError(err)
	}
	return nil
}

// GetTableByID 通过ID获取数据表
func (r *TableRepository) GetTableByID(ctx context.Context, id string) (*table.Table, error) {
	var model models.Table
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, r.handleDBError(err)
	}
	return r.modelToDomain(&model), nil
}

// UpdateTable 更新数据表
func (r *TableRepository) UpdateTable(ctx context.Context, t *table.Table) error {
	model := r.domainToModel(t)
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return r.handleDBError(err)
	}
	return nil
}

// DeleteTable 删除数据表 (软删除)
func (r *TableRepository) DeleteTable(ctx context.Context, id string) error {
	// GORM的软删除会自动处理DeletedAt字段
	if err := r.db.WithContext(ctx).Delete(&models.Table{}, "id = ?", id).Error; err != nil {
		return r.handleDBError(err)
	}
	return nil
}

// ListTables 列出数据表
func (r *TableRepository) ListTables(ctx context.Context, filter table.ListTableFilter) ([]*table.Table, error) {
	var modelTables []models.Table
	query := r.db.WithContext(ctx).Model(&models.Table{})

	if filter.BaseID != nil {
		query = query.Where("base_id = ?", *filter.BaseID)
	}
	if filter.Name != nil {
		query = query.Where("name LIKE ?", "%"+*filter.Name+"%")
	}
	if filter.CreatedBy != nil {
		query = query.Where("created_by = ?", *filter.CreatedBy)
	}
	if filter.Search != "" {
		like := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", like, like)
	}
	if filter.OrderBy != "" && filter.Order != "" {
		query = query.Order(fmt.Sprintf("%s %s", filter.OrderBy, filter.Order))
	} else {
		query = query.Order("created_time desc")
	}
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	if err := query.Find(&modelTables).Error; err != nil {
		return nil, r.handleDBError(err)
	}

	tables := make([]*table.Table, len(modelTables))
	for i, model := range modelTables {
		tables[i] = r.modelToDomain(&model)
	}
	return tables, nil
}

// CountTables 统计数据表数量
func (r *TableRepository) CountTables(ctx context.Context, filter table.ListTableFilter) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Table{})

	if filter.BaseID != nil {
		query = query.Where("base_id = ?", *filter.BaseID)
	}
	if filter.Name != nil {
		query = query.Where("name LIKE ?", "%"+*filter.Name+"%")
	}
	if filter.CreatedBy != nil {
		query = query.Where("created_by = ?", *filter.CreatedBy)
	}
	if filter.Search != "" {
		like := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", like, like)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, r.handleDBError(err)
	}
	return count, nil
}

// ExistsTable 检查数据表是否存在
func (r *TableRepository) ExistsTable(ctx context.Context, filter table.ListTableFilter) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Table{})

	if filter.BaseID != nil {
		query = query.Where("base_id = ?", *filter.BaseID)
	}
	if filter.Name != nil {
		query = query.Where("name = ?", *filter.Name)
	}
	if filter.CreatedBy != nil {
		query = query.Where("created_by = ?", *filter.CreatedBy)
	}

	if err := query.Count(&count).Error; err != nil {
		return false, r.handleDBError(err)
	}
	return count > 0, nil
}

// CreateField 创建字段
func (r *TableRepository) CreateField(ctx context.Context, f *table.Field) error {
	model := r.fieldDomainToModel(f)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return r.handleDBError(err)
	}
	return nil
}

// GetFieldByID 通过ID获取字段
func (r *TableRepository) GetFieldByID(ctx context.Context, id string) (*table.Field, error) {
	var model models.Field
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, r.handleDBError(err)
	}
	return r.fieldModelToDomain(&model), nil
}

// UpdateField 更新字段
func (r *TableRepository) UpdateField(ctx context.Context, f *table.Field) error {
	model := r.fieldDomainToModel(f)
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return r.handleDBError(err)
	}
	return nil
}

// DeleteField 删除字段 (软删除)
func (r *TableRepository) DeleteField(ctx context.Context, id string) error {
	// GORM的软删除会自动处理DeletedAt字段
	if err := r.db.WithContext(ctx).Delete(&models.Field{}, "id = ?", id).Error; err != nil {
		return r.handleDBError(err)
	}
	return nil
}

// ListFields 列出字段
func (r *TableRepository) ListFields(ctx context.Context, filter table.ListFieldFilter) ([]*table.Field, error) {
	var modelFields []models.Field
	query := r.db.WithContext(ctx).Model(&models.Field{})

	if filter.TableID != nil {
		query = query.Where("table_id = ?", *filter.TableID)
	}
	if filter.Name != nil {
		query = query.Where("name LIKE ?", "%"+*filter.Name+"%")
	}
	if filter.Type != nil {
		query = query.Where("type = ?", *filter.Type)
	}
	if filter.CreatedBy != nil {
		query = query.Where("created_by = ?", *filter.CreatedBy)
	}
	if filter.OrderBy != "" && filter.Order != "" {
		query = query.Order(fmt.Sprintf("%s %s", filter.OrderBy, filter.Order))
	} else {
		query = query.Order("field_order asc, created_time asc")
	}
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	if err := query.Find(&modelFields).Error; err != nil {
		return nil, r.handleDBError(err)
	}

	fields := make([]*table.Field, len(modelFields))
	for i, model := range modelFields {
		fields[i] = r.fieldModelToDomain(&model)
	}
	return fields, nil
}

// CountFields 统计字段数量
func (r *TableRepository) CountFields(ctx context.Context, filter table.ListFieldFilter) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Field{})

	if filter.TableID != nil {
		query = query.Where("table_id = ?", *filter.TableID)
	}
	if filter.Name != nil {
		query = query.Where("name LIKE ?", "%"+*filter.Name+"%")
	}
	if filter.Type != nil {
		query = query.Where("type = ?", *filter.Type)
	}
	if filter.CreatedBy != nil {
		query = query.Where("created_by = ?", *filter.CreatedBy)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, r.handleDBError(err)
	}
	return count, nil
}

// ExistsField 检查字段是否存在
func (r *TableRepository) ExistsField(ctx context.Context, filter table.ListFieldFilter) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Field{})

	if filter.TableID != nil {
		query = query.Where("table_id = ?", *filter.TableID)
	}
	if filter.Name != nil {
		query = query.Where("name = ?", *filter.Name)
	}
	if filter.Type != nil {
		query = query.Where("type = ?", *filter.Type)
	}
	if filter.CreatedBy != nil {
		query = query.Where("created_by = ?", *filter.CreatedBy)
	}

	if err := query.Count(&count).Error; err != nil {
		return false, r.handleDBError(err)
	}
	return count > 0, nil
}

// domainToModel 领域对象转数据模型
func (r *TableRepository) domainToModel(t *table.Table) *models.Table {
	model := &models.Table{
		ID:               t.ID,
		BaseID:           t.BaseID,
		Name:             t.Name,
		Description:      t.Description,
		Icon:             t.Icon,
		IsSystem:         t.IsSystem,
		CreatedBy:        t.CreatedBy,
		CreatedTime:      t.CreatedTime,
		LastModifiedTime: t.LastModifiedTime,
	}
	if t.DeletedTime != nil {
		model.DeletedTime = gorm.DeletedAt{Time: *t.DeletedTime, Valid: true}
	}
	return model
}

// modelToDomain 数据模型转领域对象
func (r *TableRepository) modelToDomain(model *models.Table) *table.Table {
	var deletedTime *time.Time
	if model.DeletedTime.Valid {
		deletedTime = &model.DeletedTime.Time
	}
	return &table.Table{
		ID:               model.ID,
		BaseID:           model.BaseID,
		Name:             model.Name,
		Description:      model.Description,
		Icon:             model.Icon,
		IsSystem:         model.IsSystem,
		CreatedBy:        model.CreatedBy,
		CreatedTime:      model.CreatedTime,
		DeletedTime:      deletedTime,
		LastModifiedTime: model.LastModifiedTime,
	}
}

// fieldDomainToModel 字段领域对象转数据模型
func (r *TableRepository) fieldDomainToModel(f *table.Field) *models.Field {
	model := &models.Field{
		ID:               f.ID,
		TableID:          f.TableID,
		Name:             f.Name,
		Type:             f.Type,
		CellValueType:    f.Type, // 默认使用相同的类型
		DBFieldType:      f.Type, // 默认使用相同的类型
		DBFieldName:      f.Name, // 默认使用相同的名称
		Description:      f.Description,
		IsRequired:       f.IsRequired,
		IsUnique:         f.IsUnique,
		IsPrimary:        &f.IsPrimary,
		DefaultValue:     f.DefaultValue,
		Options:          f.Options,
		FieldOrder:       float64(f.FieldOrder),
		CreatedBy:        f.CreatedBy,
		CreatedTime:      f.CreatedTime,
		LastModifiedTime: f.LastModifiedTime,
	}
	if f.DeletedTime != nil {
		model.DeletedTime = gorm.DeletedAt{Time: *f.DeletedTime, Valid: true}
	}
	return model
}

// fieldModelToDomain 字段数据模型转领域对象
func (r *TableRepository) fieldModelToDomain(model *models.Field) *table.Field {
	var deletedTime *time.Time
	if model.DeletedTime.Valid {
		deletedTime = &model.DeletedTime.Time
	}

	var isPrimary bool
	if model.IsPrimary != nil {
		isPrimary = *model.IsPrimary
	}

	return &table.Field{
		ID:               model.ID,
		TableID:          model.TableID,
		Name:             model.Name,
		Type:             model.Type,
		Description:      model.Description,
		IsRequired:       model.IsRequired,
		IsUnique:         model.IsUnique,
		IsPrimary:        isPrimary,
		DefaultValue:     model.DefaultValue,
		Options:          model.Options,
		FieldOrder:       int(model.FieldOrder),
		CreatedBy:        model.CreatedBy,
		CreatedTime:      model.CreatedTime,
		DeletedTime:      deletedTime,
		LastModifiedTime: model.LastModifiedTime,
	}
}

// handleDBError 处理数据库错误
func (r *TableRepository) handleDBError(err error) error {
	// TODO: 根据具体的数据库错误类型返回对应的业务错误
	if strings.Contains(err.Error(), "duplicate key") {
		if strings.Contains(err.Error(), "name") {
			return errors.ErrEmailExists.WithDetails("数据表或字段名称已存在") // This should be ErrResourceExists
		}
	}

	return errors.ErrDatabaseOperation.WithDetails(err.Error())
}
