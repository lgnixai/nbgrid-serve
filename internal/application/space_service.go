package application

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"teable-go-backend/internal/domain/space"
	"teable-go-backend/pkg/errors"
)

// SpaceService 空间应用服务接口 - 重构后的版本，提高可测试性
// 实现需求: 1.1, 1.2, 1.5 - 空间创建、管理和软删除
type SpaceService interface {
	// 基础CRUD操作
	CreateSpace(ctx context.Context, req CreateSpaceRequest) (*SpaceResponse, error)
	GetSpace(ctx context.Context, id string, userID string) (*SpaceResponse, error)
	UpdateSpace(ctx context.Context, id string, req UpdateSpaceRequest, userID string) (*SpaceResponse, error)
	DeleteSpace(ctx context.Context, id string, userID string) error
	ListSpaces(ctx context.Context, req ListSpacesRequest, userID string) (*ListSpacesResponse, error)
	
	// 软删除和恢复机制
	RestoreSpace(ctx context.Context, id string, userID string) (*SpaceResponse, error)
	ArchiveSpace(ctx context.Context, id string, userID string) (*SpaceResponse, error)
	PermanentDeleteSpace(ctx context.Context, id string, userID string) error
	
	// 批量操作（支持事务管理）
	BulkUpdateSpaces(ctx context.Context, updates []BulkUpdateSpaceRequest, userID string) error
	BulkDeleteSpaces(ctx context.Context, spaceIDs []string, userID string) error
	BulkRestoreSpaces(ctx context.Context, spaceIDs []string, userID string) error
	
	// 统计和查询
	GetSpaceStats(ctx context.Context, spaceID string, userID string) (*SpaceStatsResponse, error)
	GetDeletedSpaces(ctx context.Context, userID string, req ListDeletedSpacesRequest) (*ListSpacesResponse, error)
}

// SpaceApplicationService 空间应用服务实现 - 重构后的版本
// 实现需求: 1.1, 1.2, 1.5 - 空间创建、管理和软删除
type SpaceApplicationService struct {
	spaceRepo     space.Repository
	domainService space.Service
	logger        *zap.Logger
}

// NewSpaceApplicationService 创建空间应用服务
func NewSpaceApplicationService(
	spaceRepo space.Repository,
	domainService space.Service,
	logger *zap.Logger,
) SpaceService {
	return &SpaceApplicationService{
		spaceRepo:     spaceRepo,
		domainService: domainService,
		logger:        logger,
	}
}

// CreateSpaceRequest 创建空间请求
type CreateSpaceRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=255"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=2000"`
	Icon        *string `json:"icon,omitempty" validate:"omitempty,max=100"`
	CreatedBy   string  `json:"created_by" validate:"required"`
}

// UpdateSpaceRequest 更新空间请求
type UpdateSpaceRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=2000"`
	Icon        *string `json:"icon,omitempty" validate:"omitempty,max=100"`
}

// SpaceResponse 空间响应
type SpaceResponse struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	Description      *string    `json:"description"`
	Icon             *string    `json:"icon"`
	CreatedBy        string     `json:"created_by"`
	CreatedTime      time.Time  `json:"created_time"`
	LastModifiedTime *time.Time `json:"last_modified_time"`
	Status           string     `json:"status"`
	MemberCount      int        `json:"member_count"`
	IsDeleted        bool       `json:"is_deleted"`
}

// ListSpacesRequest 列出空间请求
type ListSpacesRequest struct {
	Offset    int     `json:"offset" validate:"min=0"`
	Limit     int     `json:"limit" validate:"min=1,max=100"`
	OrderBy   string  `json:"order_by"`
	Order     string  `json:"order" validate:"oneof=ASC DESC"`
	Name      *string `json:"name,omitempty"`
	CreatedBy *string `json:"created_by,omitempty"`
	Search    string  `json:"search,omitempty"`
}

// ListSpacesResponse 列出空间响应
type ListSpacesResponse struct {
	Data   []*SpaceResponse `json:"data"`
	Total  int64            `json:"total"`
	Offset int              `json:"offset"`
	Limit  int              `json:"limit"`
}

// CreateSpace 创建空间 - 重构后的版本
// 实现需求 1.1: 用户创建工作空间时生成唯一ID并设置创建者为所有者
func (s *SpaceApplicationService) CreateSpace(ctx context.Context, req CreateSpaceRequest) (*SpaceResponse, error) {
	s.logger.Info("Creating space", 
		zap.String("name", req.Name), 
		zap.String("created_by", req.CreatedBy))

	// 使用领域服务创建空间
	spaceEntity, err := s.domainService.CreateSpace(ctx, space.CreateSpaceRequest{
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
		CreatedBy:   req.CreatedBy,
	})
	if err != nil {
		s.logger.Error("Failed to create space", zap.Error(err))
		return nil, fmt.Errorf("创建空间失败: %w", err)
	}

	s.logger.Info("Space created successfully", zap.String("space_id", spaceEntity.ID))
	return s.toSpaceResponse(spaceEntity), nil
}

// GetSpace 获取空间 - 重构后的版本
// 实现需求 1.2: 用户访问工作空间时返回有权限访问的空间
func (s *SpaceApplicationService) GetSpace(ctx context.Context, id string, userID string) (*SpaceResponse, error) {
	s.logger.Info("Getting space", zap.String("space_id", id), zap.String("user_id", userID))

	spaceEntity, err := s.domainService.GetSpace(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get space", zap.String("space_id", id), zap.Error(err))
		return nil, fmt.Errorf("获取空间失败: %w", err)
	}

	// 检查用户是否有权限访问空间
	hasPermission, err := s.domainService.CheckUserPermission(ctx, id, userID, "read")
	if err != nil {
		s.logger.Error("Failed to check user permission", zap.String("space_id", id), zap.String("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("检查权限失败: %w", err)
	}

	if !hasPermission {
		s.logger.Warn("User has no permission to access space", zap.String("space_id", id), zap.String("user_id", userID))
		return nil, errors.ErrForbidden.WithMessage("没有权限访问该空间")
	}

	return s.toSpaceResponse(spaceEntity), nil
}

// UpdateSpace 更新空间 - 重构后的版本
// 实现需求 1.4: 用户更新工作空间信息时验证权限并记录修改历史
func (s *SpaceApplicationService) UpdateSpace(ctx context.Context, id string, req UpdateSpaceRequest, userID string) (*SpaceResponse, error) {
	s.logger.Info("Updating space", zap.String("space_id", id), zap.String("user_id", userID))

	// 检查用户是否有更新权限
	hasPermission, err := s.domainService.CheckUserPermission(ctx, id, userID, "update")
	if err != nil {
		s.logger.Error("Failed to check user permission", zap.String("space_id", id), zap.String("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("检查权限失败: %w", err)
	}

	if !hasPermission {
		s.logger.Warn("User has no permission to update space", zap.String("space_id", id), zap.String("user_id", userID))
		return nil, errors.ErrForbidden.WithMessage("没有权限更新该空间")
	}

	// 使用领域服务更新空间
	spaceEntity, err := s.domainService.UpdateSpace(ctx, id, space.UpdateSpaceRequest{
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
	})
	if err != nil {
		s.logger.Error("Failed to update space", zap.String("space_id", id), zap.Error(err))
		return nil, fmt.Errorf("更新空间失败: %w", err)
	}

	s.logger.Info("Space updated successfully", zap.String("space_id", id))
	return s.toSpaceResponse(spaceEntity), nil
}

// DeleteSpace 删除空间 - 重构后的版本
// 实现需求 1.5: 用户删除工作空间时执行软删除并保留数据30天
func (s *SpaceApplicationService) DeleteSpace(ctx context.Context, id string, userID string) error {
	s.logger.Info("Deleting space", zap.String("space_id", id), zap.String("user_id", userID))

	// 检查用户是否有删除权限
	hasPermission, err := s.domainService.CheckUserPermission(ctx, id, userID, "delete")
	if err != nil {
		s.logger.Error("Failed to check user permission", zap.String("space_id", id), zap.String("user_id", userID), zap.Error(err))
		return fmt.Errorf("检查权限失败: %w", err)
	}

	if !hasPermission {
		s.logger.Warn("User has no permission to delete space", zap.String("space_id", id), zap.String("user_id", userID))
		return errors.ErrForbidden.WithMessage("没有权限删除该空间")
	}

	// 使用领域服务删除空间（软删除）
	if err := s.domainService.DeleteSpace(ctx, id); err != nil {
		s.logger.Error("Failed to delete space", zap.String("space_id", id), zap.Error(err))
		return fmt.Errorf("删除空间失败: %w", err)
	}

	s.logger.Info("Space deleted successfully", zap.String("space_id", id))
	return nil
}

// ListSpaces 列出空间 - 重构后的版本
// 实现需求 1.2: 返回用户有权限访问的所有工作空间
func (s *SpaceApplicationService) ListSpaces(ctx context.Context, req ListSpacesRequest, userID string) (*ListSpacesResponse, error) {
	s.logger.Info("Listing spaces", zap.String("user_id", userID), zap.Int("offset", req.Offset), zap.Int("limit", req.Limit))

	// 设置默认值
	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.OrderBy == "" {
		req.OrderBy = "created_time"
	}
	if req.Order == "" {
		req.Order = "DESC"
	}

	// 构建过滤条件
	filter := space.ListFilter{
		Offset:    req.Offset,
		Limit:     req.Limit,
		OrderBy:   req.OrderBy,
		Order:     req.Order,
		Name:      req.Name,
		CreatedBy: req.CreatedBy,
		Search:    req.Search,
	}

	// 获取用户有权限访问的空间
	spaces, total, err := s.domainService.GetUserSpaces(ctx, userID, filter)
	if err != nil {
		s.logger.Error("Failed to list spaces", zap.String("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("获取空间列表失败: %w", err)
	}

	// 转换为响应格式
	spaceResponses := make([]*SpaceResponse, len(spaces))
	for i, space := range spaces {
		spaceResponses[i] = s.toSpaceResponse(space)
	}

	response := &ListSpacesResponse{
		Data:   spaceResponses,
		Total:  total,
		Offset: req.Offset,
		Limit:  req.Limit,
	}

	s.logger.Info("Spaces listed successfully", zap.String("user_id", userID), zap.Int("count", len(spaceResponses)), zap.Int64("total", total))
	return response, nil
}

// RestoreSpace 恢复空间 - 重构后的版本
// 实现需求 1.5: 支持数据恢复机制
func (s *SpaceApplicationService) RestoreSpace(ctx context.Context, id string, userID string) (*SpaceResponse, error) {
	s.logger.Info("Restoring space", zap.String("space_id", id), zap.String("user_id", userID))

	// 检查用户是否有恢复权限
	hasPermission, err := s.domainService.CheckUserPermission(ctx, id, userID, "restore")
	if err != nil {
		s.logger.Error("Failed to check user permission", zap.String("space_id", id), zap.String("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("检查权限失败: %w", err)
	}

	if !hasPermission {
		s.logger.Warn("User has no permission to restore space", zap.String("space_id", id), zap.String("user_id", userID))
		return nil, errors.ErrForbidden.WithMessage("没有权限恢复该空间")
	}

	// 获取空间实体
	spaceEntity, err := s.spaceRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get space for restore", zap.String("space_id", id), zap.Error(err))
		return nil, fmt.Errorf("获取空间失败: %w", err)
	}

	if spaceEntity == nil {
		return nil, errors.ErrNotFound.WithMessage("空间不存在")
	}

	// 恢复空间
	if err := spaceEntity.Restore(); err != nil {
		s.logger.Error("Failed to restore space", zap.String("space_id", id), zap.Error(err))
		return nil, fmt.Errorf("恢复空间失败: %w", err)
	}

	// 保存更新
	if err := s.spaceRepo.Update(ctx, spaceEntity); err != nil {
		s.logger.Error("Failed to save restored space", zap.String("space_id", id), zap.Error(err))
		return nil, fmt.Errorf("保存恢复的空间失败: %w", err)
	}

	s.logger.Info("Space restored successfully", zap.String("space_id", id))
	return s.toSpaceResponse(spaceEntity), nil
}

// ArchiveSpace 归档空间 - 重构后的版本
func (s *SpaceApplicationService) ArchiveSpace(ctx context.Context, id string, userID string) (*SpaceResponse, error) {
	s.logger.Info("Archiving space", zap.String("space_id", id), zap.String("user_id", userID))

	// 检查用户是否有归档权限
	hasPermission, err := s.domainService.CheckUserPermission(ctx, id, userID, "archive")
	if err != nil {
		s.logger.Error("Failed to check user permission", zap.String("space_id", id), zap.String("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("检查权限失败: %w", err)
	}

	if !hasPermission {
		s.logger.Warn("User has no permission to archive space", zap.String("space_id", id), zap.String("user_id", userID))
		return nil, errors.ErrForbidden.WithMessage("没有权限归档该空间")
	}

	// 获取空间实体
	spaceEntity, err := s.spaceRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get space for archive", zap.String("space_id", id), zap.Error(err))
		return nil, fmt.Errorf("获取空间失败: %w", err)
	}

	if spaceEntity == nil {
		return nil, errors.ErrNotFound.WithMessage("空间不存在")
	}

	// 归档空间
	if err := spaceEntity.Archive(); err != nil {
		s.logger.Error("Failed to archive space", zap.String("space_id", id), zap.Error(err))
		return nil, fmt.Errorf("归档空间失败: %w", err)
	}

	// 保存更新
	if err := s.spaceRepo.Update(ctx, spaceEntity); err != nil {
		s.logger.Error("Failed to save archived space", zap.String("space_id", id), zap.Error(err))
		return nil, fmt.Errorf("保存归档的空间失败: %w", err)
	}

	s.logger.Info("Space archived successfully", zap.String("space_id", id))
	return s.toSpaceResponse(spaceEntity), nil
}

// BulkUpdateSpaces 批量更新空间 - 重构后的版本
// 实现事务管理确保数据一致性
func (s *SpaceApplicationService) BulkUpdateSpaces(ctx context.Context, updates []BulkUpdateSpaceRequest, userID string) error {
	s.logger.Info("Bulk updating spaces", zap.String("user_id", userID), zap.Int("count", len(updates)))

	// 验证所有空间的权限
	for _, update := range updates {
		hasPermission, err := s.domainService.CheckUserPermission(ctx, update.SpaceID, userID, "update")
		if err != nil {
			s.logger.Error("Failed to check permission for bulk update", zap.String("space_id", update.SpaceID), zap.String("user_id", userID), zap.Error(err))
			return fmt.Errorf("检查空间 %s 权限失败: %w", update.SpaceID, err)
		}

		if !hasPermission {
			s.logger.Warn("User has no permission to update space in bulk", zap.String("space_id", update.SpaceID), zap.String("user_id", userID))
			return errors.ErrForbidden.WithMessage(fmt.Sprintf("没有权限更新空间 %s", update.SpaceID))
		}
	}

	// 执行批量更新
	bulkUpdates := make([]space.BulkUpdateRequest, len(updates))
	for i, update := range updates {
		bulkUpdates[i] = space.BulkUpdateRequest{
			SpaceID: update.SpaceID,
			Updates: space.UpdateSpaceRequest{
				Name:        update.Updates.Name,
				Description: update.Updates.Description,
				Icon:        update.Updates.Icon,
			},
		}
	}

	if err := s.domainService.BulkUpdateSpaces(ctx, bulkUpdates); err != nil {
		s.logger.Error("Failed to bulk update spaces", zap.String("user_id", userID), zap.Error(err))
		return fmt.Errorf("批量更新空间失败: %w", err)
	}

	s.logger.Info("Spaces bulk updated successfully", zap.String("user_id", userID), zap.Int("count", len(updates)))
	return nil
}

// BulkDeleteSpaces 批量删除空间 - 重构后的版本
// 实现事务管理确保数据一致性
func (s *SpaceApplicationService) BulkDeleteSpaces(ctx context.Context, spaceIDs []string, userID string) error {
	s.logger.Info("Bulk deleting spaces", zap.String("user_id", userID), zap.Int("count", len(spaceIDs)))

	// 验证所有空间的权限
	for _, spaceID := range spaceIDs {
		hasPermission, err := s.domainService.CheckUserPermission(ctx, spaceID, userID, "delete")
		if err != nil {
			s.logger.Error("Failed to check permission for bulk delete", zap.String("space_id", spaceID), zap.String("user_id", userID), zap.Error(err))
			return fmt.Errorf("检查空间 %s 权限失败: %w", spaceID, err)
		}

		if !hasPermission {
			s.logger.Warn("User has no permission to delete space in bulk", zap.String("space_id", spaceID), zap.String("user_id", userID))
			return errors.ErrForbidden.WithMessage(fmt.Sprintf("没有权限删除空间 %s", spaceID))
		}
	}

	// 执行批量删除
	if err := s.domainService.BulkDeleteSpaces(ctx, spaceIDs); err != nil {
		s.logger.Error("Failed to bulk delete spaces", zap.String("user_id", userID), zap.Error(err))
		return fmt.Errorf("批量删除空间失败: %w", err)
	}

	s.logger.Info("Spaces bulk deleted successfully", zap.String("user_id", userID), zap.Int("count", len(spaceIDs)))
	return nil
}

// PermanentDeleteSpace 永久删除空间 - 重构后的版本
// 实现需求 1.5: 支持彻底删除过期的软删除数据
func (s *SpaceApplicationService) PermanentDeleteSpace(ctx context.Context, id string, userID string) error {
	s.logger.Info("Permanently deleting space", zap.String("space_id", id), zap.String("user_id", userID))

	// 检查用户是否有删除权限
	hasPermission, err := s.domainService.CheckUserPermission(ctx, id, userID, "delete")
	if err != nil {
		s.logger.Error("Failed to check user permission", zap.String("space_id", id), zap.String("user_id", userID), zap.Error(err))
		return fmt.Errorf("检查权限失败: %w", err)
	}

	if !hasPermission {
		s.logger.Warn("User has no permission to permanently delete space", zap.String("space_id", id), zap.String("user_id", userID))
		return errors.ErrForbidden.WithMessage("没有权限永久删除该空间")
	}

	// 获取空间实体确认其已被软删除
	spaceEntity, err := s.spaceRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get space for permanent deletion", zap.String("space_id", id), zap.Error(err))
		return fmt.Errorf("获取空间失败: %w", err)
	}

	if spaceEntity == nil {
		return errors.ErrNotFound.WithMessage("空间不存在")
	}

	if !spaceEntity.IsDeleted() {
		return errors.ErrBadRequest.WithMessage("只能永久删除已软删除的空间")
	}

	// 执行永久删除
	if err := s.spaceRepo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to permanently delete space", zap.String("space_id", id), zap.Error(err))
		return fmt.Errorf("永久删除空间失败: %w", err)
	}

	s.logger.Info("Space permanently deleted successfully", zap.String("space_id", id))
	return nil
}

// BulkRestoreSpaces 批量恢复空间 - 重构后的版本
// 实现需求 1.5: 支持批量数据恢复机制
func (s *SpaceApplicationService) BulkRestoreSpaces(ctx context.Context, spaceIDs []string, userID string) error {
	s.logger.Info("Bulk restoring spaces", zap.String("user_id", userID), zap.Int("count", len(spaceIDs)))

	// 验证所有空间的权限
	for _, spaceID := range spaceIDs {
		hasPermission, err := s.domainService.CheckUserPermission(ctx, spaceID, userID, "restore")
		if err != nil {
			s.logger.Error("Failed to check permission for bulk restore", zap.String("space_id", spaceID), zap.String("user_id", userID), zap.Error(err))
			return fmt.Errorf("检查空间 %s 权限失败: %w", spaceID, err)
		}

		if !hasPermission {
			s.logger.Warn("User has no permission to restore space in bulk", zap.String("space_id", spaceID), zap.String("user_id", userID))
			return errors.ErrForbidden.WithMessage(fmt.Sprintf("没有权限恢复空间 %s", spaceID))
		}
	}

	// 执行批量恢复
	for _, spaceID := range spaceIDs {
		// 获取空间实体
		spaceEntity, err := s.spaceRepo.GetByID(ctx, spaceID)
		if err != nil {
			s.logger.Error("Failed to get space for bulk restore", zap.String("space_id", spaceID), zap.Error(err))
			return fmt.Errorf("获取空间 %s 失败: %w", spaceID, err)
		}

		if spaceEntity == nil {
			s.logger.Warn("Space not found for bulk restore", zap.String("space_id", spaceID))
			continue // 跳过不存在的空间
		}

		if !spaceEntity.IsDeleted() {
			s.logger.Warn("Space is not deleted, skipping restore", zap.String("space_id", spaceID))
			continue // 跳过未删除的空间
		}

		// 恢复空间
		if err := spaceEntity.Restore(); err != nil {
			s.logger.Error("Failed to restore space in bulk", zap.String("space_id", spaceID), zap.Error(err))
			return fmt.Errorf("恢复空间 %s 失败: %w", spaceID, err)
		}

		// 保存更新
		if err := s.spaceRepo.Update(ctx, spaceEntity); err != nil {
			s.logger.Error("Failed to save restored space in bulk", zap.String("space_id", spaceID), zap.Error(err))
			return fmt.Errorf("保存恢复的空间 %s 失败: %w", spaceID, err)
		}
	}

	s.logger.Info("Spaces bulk restored successfully", zap.String("user_id", userID), zap.Int("count", len(spaceIDs)))
	return nil
}

// GetSpaceStats 获取空间统计信息 - 重构后的版本
func (s *SpaceApplicationService) GetSpaceStats(ctx context.Context, spaceID string, userID string) (*SpaceStatsResponse, error) {
	s.logger.Info("Getting space stats", zap.String("space_id", spaceID), zap.String("user_id", userID))

	// 检查用户是否有读取权限
	hasPermission, err := s.domainService.CheckUserPermission(ctx, spaceID, userID, "read")
	if err != nil {
		s.logger.Error("Failed to check user permission", zap.String("space_id", spaceID), zap.String("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("检查权限失败: %w", err)
	}

	if !hasPermission {
		s.logger.Warn("User has no permission to view space stats", zap.String("space_id", spaceID), zap.String("user_id", userID))
		return nil, errors.ErrForbidden.WithMessage("没有权限查看该空间统计信息")
	}

	// 获取空间信息
	spaceEntity, err := s.domainService.GetSpace(ctx, spaceID)
	if err != nil {
		s.logger.Error("Failed to get space for stats", zap.String("space_id", spaceID), zap.Error(err))
		return nil, fmt.Errorf("获取空间失败: %w", err)
	}

	// 获取统计信息
	stats, err := s.domainService.GetSpaceStats(ctx, spaceID)
	if err != nil {
		s.logger.Error("Failed to get space stats", zap.String("space_id", spaceID), zap.Error(err))
		return nil, fmt.Errorf("获取空间统计信息失败: %w", err)
	}

	response := &SpaceStatsResponse{
		SpaceID:            stats.SpaceID,
		TotalBases:         stats.TotalBases,
		TotalTables:        stats.TotalTables,
		TotalRecords:       stats.TotalRecords,
		TotalCollaborators: stats.TotalCollaborators,
		StorageUsed:        0, // TODO: 实现存储使用量统计
		CreatedAt:          spaceEntity.CreatedTime,
		Status:             string(spaceEntity.GetStatus()),
	}

	if stats.LastActivityAt != nil {
		if lastActivity, parseErr := time.Parse(time.RFC3339, *stats.LastActivityAt); parseErr == nil {
			response.LastActivityAt = &lastActivity
		}
	}

	s.logger.Info("Space stats retrieved successfully", zap.String("space_id", spaceID))
	return response, nil
}

// GetDeletedSpaces 获取已删除的空间列表 - 重构后的版本
// 实现需求 1.5: 支持查看软删除的数据
func (s *SpaceApplicationService) GetDeletedSpaces(ctx context.Context, userID string, req ListDeletedSpacesRequest) (*ListSpacesResponse, error) {
	s.logger.Info("Getting deleted spaces", zap.String("user_id", userID), zap.Int("offset", req.Offset), zap.Int("limit", req.Limit))

	// 设置默认值
	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.OrderBy == "" {
		req.OrderBy = "deleted_time"
	}
	if req.Order == "" {
		req.Order = "DESC"
	}

	// 构建过滤条件，只查询已删除的空间
	filter := space.ListFilter{
		Offset:    req.Offset,
		Limit:     req.Limit,
		OrderBy:   req.OrderBy,
		Order:     req.Order,
		Name:      req.Name,
		CreatedBy: &userID, // 只查询用户自己的空间
		Search:    req.Search,
	}

	// 获取已删除的空间（需要在仓储层实现专门的查询方法）
	spaces, total, err := s.domainService.GetUserDeletedSpaces(ctx, userID, filter)
	if err != nil {
		s.logger.Error("Failed to get deleted spaces", zap.String("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("获取已删除空间列表失败: %w", err)
	}

	// 转换为响应格式
	spaceResponses := make([]*SpaceResponse, len(spaces))
	for i, space := range spaces {
		spaceResponses[i] = s.toSpaceResponse(space)
	}

	response := &ListSpacesResponse{
		Data:   spaceResponses,
		Total:  total,
		Offset: req.Offset,
		Limit:  req.Limit,
	}

	s.logger.Info("Deleted spaces retrieved successfully", zap.String("user_id", userID), zap.Int("count", len(spaceResponses)), zap.Int64("total", total))
	return response, nil
}

// BulkUpdateSpaceRequest 批量更新空间请求
type BulkUpdateSpaceRequest struct {
	SpaceID string             `json:"space_id" validate:"required"`
	Updates UpdateSpaceRequest `json:"updates" validate:"required"`
}

// ListDeletedSpacesRequest 列出已删除空间请求
type ListDeletedSpacesRequest struct {
	Offset         int     `json:"offset" validate:"min=0"`
	Limit          int     `json:"limit" validate:"min=1,max=100"`
	OrderBy        string  `json:"order_by"`
	Order          string  `json:"order" validate:"oneof=ASC DESC"`
	Name           *string `json:"name,omitempty"`
	DeletedAfter   *string `json:"deleted_after,omitempty"`  // ISO 8601 格式
	DeletedBefore  *string `json:"deleted_before,omitempty"` // ISO 8601 格式
	Search         string  `json:"search,omitempty"`
}

// SpaceStatsResponse 空间统计响应
type SpaceStatsResponse struct {
	SpaceID            string     `json:"space_id"`
	TotalBases         int64      `json:"total_bases"`
	TotalTables        int64      `json:"total_tables"`
	TotalRecords       int64      `json:"total_records"`
	TotalCollaborators int64      `json:"total_collaborators"`
	StorageUsed        int64      `json:"storage_used_bytes"`
	LastActivityAt     *time.Time `json:"last_activity_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	Status             string     `json:"status"`
}

// toSpaceResponse 转换为空间响应
func (s *SpaceApplicationService) toSpaceResponse(spaceEntity *space.Space) *SpaceResponse {
	return &SpaceResponse{
		ID:               spaceEntity.ID,
		Name:             spaceEntity.Name,
		Description:      spaceEntity.Description,
		Icon:             spaceEntity.Icon,
		CreatedBy:        spaceEntity.CreatedBy,
		CreatedTime:      spaceEntity.CreatedTime,
		LastModifiedTime: spaceEntity.LastModifiedTime,
		Status:           string(spaceEntity.GetStatus()),
		MemberCount:      spaceEntity.GetMemberCount(),
		IsDeleted:        spaceEntity.IsDeleted(),
	}
}