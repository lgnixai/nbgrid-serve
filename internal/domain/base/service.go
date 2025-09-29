package base

import (
	"context"
	"errors"
)

// 业务错误定义
var (
	ErrBaseNotFound     = errors.New("基础表不存在")
	ErrInvalidName      = errors.New("无效的基础表名称")
	ErrPermissionDenied = errors.New("权限不足")
)

// Service 基础表服务接口
type Service interface {
	// 创建基础表
	CreateBase(ctx context.Context, req CreateBaseRequest) (*Base, error)

	// 获取基础表
	GetBase(ctx context.Context, id string) (*Base, error)

	// 更新基础表
	UpdateBase(ctx context.Context, id string, req UpdateBaseRequest) (*Base, error)

	// 删除基础表
	DeleteBase(ctx context.Context, id string) error

	// 列出基础表
	ListBases(ctx context.Context, filter ListFilter) ([]*Base, error)

	// 统计基础表数量
	CountBases(ctx context.Context, filter CountFilter) (int64, error)

	// 批量操作
	BulkUpdateBases(ctx context.Context, updates []BulkUpdateRequest) error
	BulkDeleteBases(ctx context.Context, baseIDs []string) error

	// 权限检查
	CheckUserPermission(ctx context.Context, baseID, userID, permission string) (bool, error)

	// 统计信息
	GetBaseStats(ctx context.Context, baseID string) (*BaseStats, error)
	GetSpaceBaseStats(ctx context.Context, spaceID string) (*SpaceBaseStats, error)

	// 导出/导入
	ExportBases(ctx context.Context, filter ListFilter) ([]*Base, error)
	ImportBases(ctx context.Context, bases []CreateBaseRequest) ([]*Base, error)
}

// CreateBaseRequest 创建基础表请求
type CreateBaseRequest struct {
	SpaceID     string  `json:"space_id" binding:"required"`
	Name        string  `json:"name" binding:"required,min=1,max=255"`
	Description *string `json:"description"`
	Icon        *string `json:"icon"`
	CreatedBy   string  `json:"-"` // 从JWT中获取
}

// UpdateBaseRequest 更新基础表请求
type UpdateBaseRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=255"`
	Description *string `json:"description"`
	Icon        *string `json:"icon"`
}

// PaginatedResult 分页结果
type PaginatedResult struct {
	Data   []*Base `json:"data"`
	Total  int64   `json:"total"`
	Limit  int     `json:"limit"`
	Offset int     `json:"offset"`
}

// BulkUpdateRequest 批量更新请求
type BulkUpdateRequest struct {
	BaseID  string            `json:"base_id" validate:"required"`
	Updates UpdateBaseRequest `json:"updates" validate:"required"`
}

// BaseStats 基础表统计信息
type BaseStats struct {
	BaseID         string  `json:"base_id"`
	TotalTables    int64   `json:"total_tables"`
	TotalRecords   int64   `json:"total_records"`
	TotalFields    int64   `json:"total_fields"`
	LastActivityAt *string `json:"last_activity_at,omitempty"`
}

// SpaceBaseStats 空间基础表统计信息
type SpaceBaseStats struct {
	SpaceID      string `json:"space_id"`
	TotalBases   int64  `json:"total_bases"`
	TotalTables  int64  `json:"total_tables"`
	TotalRecords int64  `json:"total_records"`
	TotalFields  int64  `json:"total_fields"`
}

// ServiceImpl 基础表服务实现 - 重构后的版本
type ServiceImpl struct {
	repo Repository
}

// NewService 创建基础表服务 - 重构后的版本
func NewService(repo Repository) Service {
	return &ServiceImpl{
		repo: repo,
	}
}

// CreateBase 创建基础表 - 重构后的版本
func (s *ServiceImpl) CreateBase(ctx context.Context, req CreateBaseRequest) (*Base, error) {
	// 检查同一空间下名称是否已存在
	exists, err := s.repo.Exists(ctx, ExistsFilter{
		SpaceID: &req.SpaceID,
		Name:    &req.Name,
	})
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrBaseExists
	}

	// 创建基础表
	base, err := NewBase(req.SpaceID, req.Name, req.CreatedBy)
	if err != nil {
		return nil, err
	}
	
	base.Description = req.Description
	base.Icon = req.Icon

	if err := base.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, base); err != nil {
		return nil, err
	}

	return base, nil
}

// GetBase 获取基础表
func (s *ServiceImpl) GetBase(ctx context.Context, id string) (*Base, error) {
	base, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if base == nil {
		return nil, ErrBaseNotFound
	}
	return base, nil
}

// UpdateBase 更新基础表 - 重构后的版本
func (s *ServiceImpl) UpdateBase(ctx context.Context, id string, req UpdateBaseRequest) (*Base, error) {
	base, err := s.GetBase(ctx, id)
	if err != nil {
		return nil, err
	}

	// 验证基础表是否可以更新
	if err := base.ValidateForUpdate(); err != nil {
		return nil, err
	}

	// 如果名称改变，检查新名称是否已存在
	if base.Name != req.Name {
		exists, err := s.repo.Exists(ctx, ExistsFilter{
			SpaceID: &base.SpaceID,
			Name:    &req.Name,
		})
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrBaseExists
		}
	}

	if err := base.Update(&req.Name, req.Description, req.Icon); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, base); err != nil {
		return nil, err
	}

	return base, nil
}

// DeleteBase 删除基础表 - 重构后的版本
func (s *ServiceImpl) DeleteBase(ctx context.Context, id string) error {
	base, err := s.GetBase(ctx, id)
	if err != nil {
		return err
	}

	// 验证基础表是否可以删除
	if err := base.ValidateForDeletion(); err != nil {
		return err
	}

	base.SoftDelete()
	return s.repo.Update(ctx, base)
}

// ListBases 列出基础表
func (s *ServiceImpl) ListBases(ctx context.Context, filter ListFilter) ([]*Base, error) {
	return s.repo.List(ctx, filter)
}

// CountBases 统计基础表数量
func (s *ServiceImpl) CountBases(ctx context.Context, filter CountFilter) (int64, error) {
	return s.repo.Count(ctx, filter)
}

// BulkUpdateBases 批量更新基础表
func (s *ServiceImpl) BulkUpdateBases(ctx context.Context, updates []BulkUpdateRequest) error {
	for _, update := range updates {
		_, err := s.UpdateBase(ctx, update.BaseID, update.Updates)
		if err != nil {
			return err
		}
	}
	return nil
}

// BulkDeleteBases 批量删除基础表
func (s *ServiceImpl) BulkDeleteBases(ctx context.Context, baseIDs []string) error {
	for _, baseID := range baseIDs {
		if err := s.DeleteBase(ctx, baseID); err != nil {
			return err
		}
	}
	return nil
}

// CheckUserPermission 检查用户权限
func (s *ServiceImpl) CheckUserPermission(ctx context.Context, baseID, userID, permission string) (bool, error) {
	// 获取基础表信息
	base, err := s.GetBase(ctx, baseID)
	if err != nil {
		return false, err
	}

	// 如果是基础表创建者，拥有所有权限
	if base.CreatedBy == userID {
		return true, nil
	}

	// TODO: 检查空间权限
	// 这里需要集成空间服务来检查用户对空间的权限
	// 暂时返回false，需要后续集成
	return false, nil
}

// GetBaseStats 获取基础表统计信息
func (s *ServiceImpl) GetBaseStats(ctx context.Context, baseID string) (*BaseStats, error) {
	// TODO: 实现获取基础表统计信息的逻辑
	// 这里需要查询数据表、记录和字段的数量
	return &BaseStats{
		BaseID: baseID,
		// 暂时返回默认值，需要集成其他服务
		TotalTables:  0,
		TotalRecords: 0,
		TotalFields:  0,
	}, nil
}

// GetSpaceBaseStats 获取空间基础表统计信息
func (s *ServiceImpl) GetSpaceBaseStats(ctx context.Context, spaceID string) (*SpaceBaseStats, error) {
	// TODO: 实现获取空间基础表统计信息的逻辑
	return &SpaceBaseStats{
		SpaceID: spaceID,
		// 暂时返回默认值，需要集成其他服务
		TotalBases:   0,
		TotalTables:  0,
		TotalRecords: 0,
		TotalFields:  0,
	}, nil
}

// ExportBases 导出基础表
func (s *ServiceImpl) ExportBases(ctx context.Context, filter ListFilter) ([]*Base, error) {
	return s.ListBases(ctx, filter)
}

// ImportBases 导入基础表
func (s *ServiceImpl) ImportBases(ctx context.Context, bases []CreateBaseRequest) ([]*Base, error) {
	var result []*Base
	for _, baseReq := range bases {
		base, err := s.CreateBase(ctx, baseReq)
		if err != nil {
			return nil, err
		}
		result = append(result, base)
	}
	return result, nil
}
