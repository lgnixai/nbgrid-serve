package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"teable-go-backend/internal/domain/space"
	"teable-go-backend/internal/infrastructure/database/models"
	"teable-go-backend/pkg/errors"
)

// SpaceRepository 空间仓储实现 - 重构后的版本
// 支持事务管理和软删除恢复机制
type SpaceRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewSpaceRepository(db *gorm.DB, logger *zap.Logger) space.Repository {
	return &SpaceRepository{
		db:     db,
		logger: logger,
	}
}

// Create 创建空间 - 重构后的版本，支持事务管理
func (r *SpaceRepository) Create(ctx context.Context, s *space.Space) error {
	r.logger.Debug("Creating space in repository", 
		zap.String("space_id", s.ID), 
		zap.String("name", s.Name))
	
	model := r.domainToModel(s)
	
	// 使用事务确保数据一致性
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(model).Error; err != nil {
			r.logger.Error("Failed to create space in database", 
				zap.String("space_id", s.ID), 
				zap.Error(err))
			return r.handleDBError(err)
		}
		return nil
	})
	
	if err != nil {
		return err
	}
	
	r.logger.Debug("Space created successfully in repository", 
		zap.String("space_id", s.ID))
	return nil
}

// GetByID 根据ID获取空间 - 重构后的版本，支持软删除查询
func (r *SpaceRepository) GetByID(ctx context.Context, id string) (*space.Space, error) {
	r.logger.Debug("Getting space by ID", zap.String("space_id", id))
	
	var m models.Space
	// 使用Unscoped()查询包括软删除的记录，用于恢复功能
	if err := r.db.WithContext(ctx).Unscoped().Where("id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Debug("Space not found", zap.String("space_id", id))
			return nil, nil
		}
		r.logger.Error("Failed to get space from database", zap.String("space_id", id), zap.Error(err))
		return nil, r.handleDBError(err)
	}
	
	domainSpace := r.modelToDomain(&m)
	r.logger.Debug("Space retrieved successfully", zap.String("space_id", id), zap.Bool("is_deleted", domainSpace.IsDeleted()))
	return domainSpace, nil
}

// Update 更新空间 - 重构后的版本，支持事务管理和软删除状态更新
func (r *SpaceRepository) Update(ctx context.Context, s *space.Space) error {
	r.logger.Debug("Updating space in repository", zap.String("space_id", s.ID))
	
	model := r.domainToModel(s)
	
	// 使用事务确保数据一致性
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 使用Unscoped()更新包括软删除的记录
		if err := tx.Unscoped().Save(model).Error; err != nil {
			r.logger.Error("Failed to update space in database", zap.String("space_id", s.ID), zap.Error(err))
			return r.handleDBError(err)
		}
		return nil
	})
	
	if err != nil {
		return err
	}
	
	r.logger.Debug("Space updated successfully in repository", zap.String("space_id", s.ID))
	return nil
}

// Delete 硬删除空间 - 重构后的版本，支持事务管理
// 注意：通常使用软删除，此方法用于彻底清理过期数据
func (r *SpaceRepository) Delete(ctx context.Context, id string) error {
	r.logger.Debug("Hard deleting space", zap.String("space_id", id))
	
	// 使用事务确保数据一致性
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先删除相关的协作者记录
		if err := tx.Where("space_id = ?", id).Delete(&models.SpaceCollaborator{}).Error; err != nil {
			r.logger.Error("Failed to delete space collaborators", zap.String("space_id", id), zap.Error(err))
			return r.handleDBError(err)
		}
		
		// 然后硬删除空间记录
		if err := tx.Unscoped().Delete(&models.Space{}, "id = ?", id).Error; err != nil {
			r.logger.Error("Failed to hard delete space", zap.String("space_id", id), zap.Error(err))
			return r.handleDBError(err)
		}
		
		return nil
	})
	
	if err != nil {
		return err
	}
	
	r.logger.Debug("Space hard deleted successfully", zap.String("space_id", id))
	return nil
}

// List 列出空间 - 重构后的版本，优化查询性能
func (r *SpaceRepository) List(ctx context.Context, filter space.ListFilter) ([]*space.Space, error) {
	r.logger.Debug("Listing spaces", zap.String("filter", fmt.Sprintf("%+v", filter)))
	
	var rows []models.Space
	q := r.db.WithContext(ctx).Model(&models.Space{})
	
	// 默认只查询未删除的记录
	q = q.Where("deleted_time IS NULL")
	
	// 应用过滤条件
	if filter.Name != nil {
		q = q.Where("name LIKE ?", "%"+*filter.Name+"%")
	}
	if filter.CreatedBy != nil {
		q = q.Where("created_by = ?", *filter.CreatedBy)
	}
	if filter.Search != "" {
		like := "%" + strings.ToLower(filter.Search) + "%"
		q = q.Where("LOWER(name) LIKE ? OR LOWER(COALESCE(description, '')) LIKE ?", like, like)
	}
	
	// 应用排序
	if filter.OrderBy != "" && filter.Order != "" {
		q = q.Order(fmt.Sprintf("%s %s", filter.OrderBy, filter.Order))
	} else {
		q = q.Order("created_time DESC") // 默认按创建时间倒序
	}
	
	// 应用分页
	if filter.Limit > 0 {
		q = q.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		q = q.Offset(filter.Offset)
	}
	
	if err := q.Find(&rows).Error; err != nil {
		r.logger.Error("Failed to list spaces from database", zap.Error(err))
		return nil, r.handleDBError(err)
	}
	
	items := make([]*space.Space, len(rows))
	for i := range rows {
		items[i] = r.modelToDomain(&rows[i])
	}
	
	r.logger.Debug("Spaces listed successfully", zap.Int("count", len(items)))
	return items, nil
}

// Count 统计空间数量 - 重构后的版本，优化查询性能
func (r *SpaceRepository) Count(ctx context.Context, filter space.CountFilter) (int64, error) {
	r.logger.Debug("Counting spaces", zap.String("filter", fmt.Sprintf("%+v", filter)))
	
	var count int64
	q := r.db.WithContext(ctx).Model(&models.Space{})
	
	// 默认只统计未删除的记录
	q = q.Where("deleted_time IS NULL")
	
	// 应用过滤条件
	if filter.Name != nil {
		q = q.Where("name LIKE ?", "%"+*filter.Name+"%")
	}
	if filter.CreatedBy != nil {
		q = q.Where("created_by = ?", *filter.CreatedBy)
	}
	if filter.Search != "" {
		like := "%" + strings.ToLower(filter.Search) + "%"
		q = q.Where("LOWER(name) LIKE ? OR LOWER(COALESCE(description, '')) LIKE ?", like, like)
	}
	
	if err := q.Count(&count).Error; err != nil {
		r.logger.Error("Failed to count spaces from database", zap.Error(err))
		return 0, r.handleDBError(err)
	}
	
	r.logger.Debug("Spaces counted successfully", zap.Int64("count", count))
	return count, nil
}

// AddCollaborator 添加协作者 - 重构后的版本，支持事务管理
func (r *SpaceRepository) AddCollaborator(ctx context.Context, collab *space.SpaceCollaborator) error {
	r.logger.Debug("Adding collaborator", zap.String("space_id", collab.SpaceID), zap.String("user_id", collab.UserID), zap.String("role", string(collab.Role)))
	
	m := &models.SpaceCollaborator{
		ID:          collab.ID,
		SpaceID:     collab.SpaceID,
		UserID:      collab.UserID,
		Role:        string(collab.Role),
		CreatedTime: collab.CreatedTime,
	}
	
	// 使用事务确保数据一致性
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 检查是否已存在相同的协作者
		var existing models.SpaceCollaborator
		if err := tx.Where("space_id = ? AND user_id = ?", collab.SpaceID, collab.UserID).First(&existing).Error; err == nil {
			return errors.ErrConflict.WithMessage("协作者已存在")
		} else if err != gorm.ErrRecordNotFound {
			return r.handleDBError(err)
		}
		
		if err := tx.Create(m).Error; err != nil {
			r.logger.Error("Failed to add collaborator to database", zap.String("space_id", collab.SpaceID), zap.String("user_id", collab.UserID), zap.Error(err))
			return r.handleDBError(err)
		}
		return nil
	})
	
	if err != nil {
		return err
	}
	
	r.logger.Debug("Collaborator added successfully", zap.String("space_id", collab.SpaceID), zap.String("user_id", collab.UserID))
	return nil
}

// RemoveCollaborator 移除协作者 - 重构后的版本，支持事务管理
func (r *SpaceRepository) RemoveCollaborator(ctx context.Context, id string) error {
	r.logger.Debug("Removing collaborator", zap.String("collaborator_id", id))
	
	// 使用事务确保数据一致性
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Delete(&models.SpaceCollaborator{}, "id = ?", id)
		if result.Error != nil {
			r.logger.Error("Failed to remove collaborator from database", zap.String("collaborator_id", id), zap.Error(result.Error))
			return r.handleDBError(result.Error)
		}
		
		if result.RowsAffected == 0 {
			return errors.ErrNotFound.WithMessage("协作者不存在")
		}
		
		return nil
	})
	
	if err != nil {
		return err
	}
	
	r.logger.Debug("Collaborator removed successfully", zap.String("collaborator_id", id))
	return nil
}

// ListCollaborators 列出协作者 - 重构后的版本，优化查询性能
func (r *SpaceRepository) ListCollaborators(ctx context.Context, spaceID string) ([]*space.SpaceCollaborator, error) {
	r.logger.Debug("Listing collaborators", zap.String("space_id", spaceID))
	
	var rows []models.SpaceCollaborator
	if err := r.db.WithContext(ctx).Where("space_id = ?", spaceID).Order("created_time ASC").Find(&rows).Error; err != nil {
		r.logger.Error("Failed to list collaborators from database", zap.String("space_id", spaceID), zap.Error(err))
		return nil, r.handleDBError(err)
	}
	
	items := make([]*space.SpaceCollaborator, len(rows))
	for i := range rows {
		items[i] = &space.SpaceCollaborator{
			ID:          rows[i].ID,
			SpaceID:     rows[i].SpaceID,
			UserID:      rows[i].UserID,
			Role:        space.CollaboratorRole(rows[i].Role),
			CreatedTime: rows[i].CreatedTime,
			Status:      space.CollaboratorStatusAccepted, // 默认为已接受状态
		}
	}
	
	r.logger.Debug("Collaborators listed successfully", zap.String("space_id", spaceID), zap.Int("count", len(items)))
	return items, nil
}

// domainToModel 领域实体转数据库模型 - 重构后的版本
func (r *SpaceRepository) domainToModel(s *space.Space) *models.Space {
	model := &models.Space{
		ID:               s.ID,
		Name:             s.Name,
		Description:      s.Description,
		Icon:             s.Icon,
		CreatedBy:        s.CreatedBy,
		CreatedTime:      s.CreatedTime,
		LastModifiedTime: s.LastModifiedTime,
	}
	
	// 处理软删除时间
	if s.DeletedTime != nil {
		model.DeletedTime = gorm.DeletedAt{
			Time:  *s.DeletedTime,
			Valid: true,
		}
	}
	
	return model
}

// modelToDomain 数据库模型转领域实体 - 重构后的版本
func (r *SpaceRepository) modelToDomain(m *models.Space) *space.Space {
	var deleted *time.Time
	if m.DeletedTime.Valid {
		deleted = &m.DeletedTime.Time
	}
	
	// 重新构建领域实体，保持业务规则
	spaceEntity := &space.Space{
		ID:               m.ID,
		Name:             m.Name,
		Description:      m.Description,
		Icon:             m.Icon,
		CreatedBy:        m.CreatedBy,
		CreatedTime:      m.CreatedTime,
		DeletedTime:      deleted,
		LastModifiedTime: m.LastModifiedTime,
	}
	
	return spaceEntity
}

// ListDeleted 列出已删除的空间 - 重构后的版本，支持软删除查询
func (r *SpaceRepository) ListDeleted(ctx context.Context, filter space.ListFilter) ([]*space.Space, error) {
	r.logger.Debug("Listing deleted spaces", zap.String("filter", fmt.Sprintf("%+v", filter)))
	
	var rows []models.Space
	q := r.db.WithContext(ctx).Model(&models.Space{}).Unscoped()
	
	// 只查询已删除的记录
	q = q.Where("deleted_time IS NOT NULL")
	
	// 应用过滤条件
	if filter.Name != nil {
		q = q.Where("name LIKE ?", "%"+*filter.Name+"%")
	}
	if filter.CreatedBy != nil {
		q = q.Where("created_by = ?", *filter.CreatedBy)
	}
	if filter.Search != "" {
		like := "%" + strings.ToLower(filter.Search) + "%"
		q = q.Where("LOWER(name) LIKE ? OR LOWER(COALESCE(description, '')) LIKE ?", like, like)
	}
	
	// 应用排序
	if filter.OrderBy != "" && filter.Order != "" {
		q = q.Order(fmt.Sprintf("%s %s", filter.OrderBy, filter.Order))
	} else {
		q = q.Order("deleted_time DESC") // 默认按删除时间倒序
	}
	
	// 应用分页
	if filter.Limit > 0 {
		q = q.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		q = q.Offset(filter.Offset)
	}
	
	if err := q.Find(&rows).Error; err != nil {
		r.logger.Error("Failed to list deleted spaces from database", zap.Error(err))
		return nil, r.handleDBError(err)
	}
	
	items := make([]*space.Space, len(rows))
	for i := range rows {
		items[i] = r.modelToDomain(&rows[i])
	}
	
	r.logger.Debug("Deleted spaces listed successfully", zap.Int("count", len(items)))
	return items, nil
}

// CountDeleted 统计已删除空间数量 - 重构后的版本，支持软删除统计
func (r *SpaceRepository) CountDeleted(ctx context.Context, filter space.CountFilter) (int64, error) {
	r.logger.Debug("Counting deleted spaces", zap.String("filter", fmt.Sprintf("%+v", filter)))
	
	var count int64
	q := r.db.WithContext(ctx).Model(&models.Space{}).Unscoped()
	
	// 只统计已删除的记录
	q = q.Where("deleted_time IS NOT NULL")
	
	// 应用过滤条件
	if filter.Name != nil {
		q = q.Where("name LIKE ?", "%"+*filter.Name+"%")
	}
	if filter.CreatedBy != nil {
		q = q.Where("created_by = ?", *filter.CreatedBy)
	}
	if filter.Search != "" {
		like := "%" + strings.ToLower(filter.Search) + "%"
		q = q.Where("LOWER(name) LIKE ? OR LOWER(COALESCE(description, '')) LIKE ?", like, like)
	}
	
	if err := q.Count(&count).Error; err != nil {
		r.logger.Error("Failed to count deleted spaces from database", zap.Error(err))
		return 0, r.handleDBError(err)
	}
	
	r.logger.Debug("Deleted spaces counted successfully", zap.Int64("count", count))
	return count, nil
}

// handleDBError 处理数据库错误 - 重构后的版本，提供更详细的错误信息
func (r *SpaceRepository) handleDBError(err error) error {
	errMsg := err.Error()
	
	// 处理常见的数据库错误
	if strings.Contains(errMsg, "duplicate key") || strings.Contains(errMsg, "UNIQUE constraint") {
		return errors.ErrConflict.WithMessage("数据已存在，违反唯一性约束")
	}
	
	if strings.Contains(errMsg, "foreign key constraint") {
		return errors.ErrConflict.WithMessage("违反外键约束")
	}
	
	if strings.Contains(errMsg, "connection") {
		return errors.ErrDatabaseConnection.WithMessage("数据库连接失败")
	}
	
	if strings.Contains(errMsg, "timeout") {
		return errors.ErrTimeout.WithMessage("数据库操作超时")
	}
	
	// 默认数据库操作错误
	return errors.ErrDatabaseOperation.WithDetails(errMsg)
}


