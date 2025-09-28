package repository

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"teable-go-backend/internal/domain/base"
	"teable-go-backend/internal/infrastructure/database/models"
	"teable-go-backend/pkg/errors"
)

// BaseRepository 基础表仓储实现
type BaseRepository struct {
	db *gorm.DB
}

// NewBaseRepository 创建基础表仓储
func NewBaseRepository(db *gorm.DB) base.Repository {
	return &BaseRepository{db: db}
}

// Create 创建基础表
func (r *BaseRepository) Create(ctx context.Context, b *base.Base) error {
	model := r.domainToModel(b)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return r.handleDBError(err)
	}

	return nil
}

// GetByID 根据ID获取基础表
func (r *BaseRepository) GetByID(ctx context.Context, id string) (*base.Base, error) {
	var model models.Base

	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, r.handleDBError(err)
	}

	return r.modelToDomain(&model), nil
}

// Update 更新基础表
func (r *BaseRepository) Update(ctx context.Context, b *base.Base) error {
	model := r.domainToModel(b)

	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return r.handleDBError(err)
	}

	return nil
}

// Delete 删除基础表
func (r *BaseRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&models.Base{}, "id = ?", id).Error; err != nil {
		return r.handleDBError(err)
	}
	return nil
}

// List 列出基础表
func (r *BaseRepository) List(ctx context.Context, filter base.ListFilter) ([]*base.Base, error) {
	var modelBases []models.Base

	query := r.db.WithContext(ctx).Model(&models.Base{})

	// 应用过滤条件
	if filter.SpaceID != nil {
		query = query.Where("space_id = ?", *filter.SpaceID)
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

	// 排序
	if filter.OrderBy != "" && filter.Order != "" {
		query = query.Order(fmt.Sprintf("%s %s", filter.OrderBy, filter.Order))
	} else {
		query = query.Order("created_time DESC")
	}

	// 分页
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	if err := query.Find(&modelBases).Error; err != nil {
		return nil, r.handleDBError(err)
	}

	// 转换为领域对象
	bases := make([]*base.Base, len(modelBases))
	for i, model := range modelBases {
		bases[i] = r.modelToDomain(&model)
	}

	return bases, nil
}

// Count 统计基础表数量
func (r *BaseRepository) Count(ctx context.Context, filter base.CountFilter) (int64, error) {
	var count int64

	query := r.db.WithContext(ctx).Model(&models.Base{})

	// 应用过滤条件
	if filter.SpaceID != nil {
		query = query.Where("space_id = ?", *filter.SpaceID)
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

// Exists 检查基础表是否存在
func (r *BaseRepository) Exists(ctx context.Context, filter base.ExistsFilter) (bool, error) {
	var count int64

	query := r.db.WithContext(ctx).Model(&models.Base{})

	// 应用过滤条件
	if filter.ID != nil {
		query = query.Where("id = ?", *filter.ID)
	}
	if filter.SpaceID != nil {
		query = query.Where("space_id = ?", *filter.SpaceID)
	}
	if filter.Name != nil {
		query = query.Where("name = ?", *filter.Name)
	}

	if err := query.Count(&count).Error; err != nil {
		return false, r.handleDBError(err)
	}

	return count > 0, nil
}

// 辅助方法

// domainToModel 领域对象转数据模型
func (r *BaseRepository) domainToModel(b *base.Base) *models.Base {
	model := &models.Base{
		ID:               b.ID,
		SpaceID:          b.SpaceID,
		Name:             b.Name,
		Description:      b.Description,
		Icon:             b.Icon,
		CreatedBy:        b.CreatedBy,
		CreatedTime:      b.CreatedTime,
		LastModifiedTime: b.LastModifiedTime,
	}

	// 处理软删除字段
	if b.DeletedTime != nil {
		model.DeletedTime = gorm.DeletedAt{
			Time:  *b.DeletedTime,
			Valid: true,
		}
	}

	return model
}

// modelToDomain 数据模型转领域对象
func (r *BaseRepository) modelToDomain(model *models.Base) *base.Base {
	b := &base.Base{
		ID:               model.ID,
		SpaceID:          model.SpaceID,
		Name:             model.Name,
		Description:      model.Description,
		Icon:             model.Icon,
		CreatedBy:        model.CreatedBy,
		CreatedTime:      model.CreatedTime,
		LastModifiedTime: model.LastModifiedTime,
	}

	// 处理软删除字段
	if model.DeletedTime.Valid {
		b.DeletedTime = &model.DeletedTime.Time
	}

	return b
}

// handleDBError 处理数据库错误
func (r *BaseRepository) handleDBError(err error) error {
	// TODO: 根据具体的数据库错误类型返回对应的业务错误
	if strings.Contains(err.Error(), "duplicate key") {
		if strings.Contains(err.Error(), "name") {
			return errors.ErrEmailExists.WithDetails("基础表名称已存在")
		}
	}

	return errors.ErrDatabaseOperation.WithDetails(err.Error())
}
