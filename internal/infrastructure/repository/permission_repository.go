package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"teable-go-backend/internal/domain/permission"
	"teable-go-backend/internal/infrastructure/database/models"
	"teable-go-backend/pkg/logger"
)

// PermissionRepository 权限仓储实现
type PermissionRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewPermissionRepository 创建权限仓储
func NewPermissionRepository(db *gorm.DB, logger *zap.Logger) permission.Repository {
	return &PermissionRepository{
		db:     db,
		logger: logger,
	}
}

// 权限管理

// CreatePermission 创建权限
func (r *PermissionRepository) CreatePermission(ctx context.Context, perm *permission.Permission) error {
	model := &models.Permission{
		ID:           perm.ID,
		UserID:       perm.UserID,
		ResourceType: perm.ResourceType,
		ResourceID:   perm.ResourceID,
		Role:         string(perm.Role),
		GrantedBy:    perm.GrantedBy,
		GrantedAt:    perm.GrantedAt,
		ExpiresAt:    perm.ExpiresAt,
		IsActive:     perm.IsActive,
		CreatedAt:    perm.CreatedAt,
		UpdatedAt:    perm.UpdatedAt,
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		r.logger.Error("Failed to create permission",
			logger.String("user_id", perm.UserID),
			logger.String("resource_type", perm.ResourceType),
			logger.String("resource_id", perm.ResourceID),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to create permission: %w", err)
	}

	return nil
}

// GetPermission 获取权限
func (r *PermissionRepository) GetPermission(ctx context.Context, id string) (*permission.Permission, error) {
	var model models.Permission
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, permission.ErrPermissionNotFound
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return r.modelToPermission(&model), nil
}

// GetPermissionsByUser 获取用户权限
func (r *PermissionRepository) GetPermissionsByUser(ctx context.Context, userID string) ([]*permission.Permission, error) {
	var models []models.Permission
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	var permissions []*permission.Permission
	for _, model := range models {
		permissions = append(permissions, r.modelToPermission(&model))
	}

	return permissions, nil
}

// GetPermissionsByResource 获取资源权限
func (r *PermissionRepository) GetPermissionsByResource(ctx context.Context, resourceType, resourceID string) ([]*permission.Permission, error) {
	var models []models.Permission
	if err := r.db.WithContext(ctx).Where("resource_type = ? AND resource_id = ?", resourceType, resourceID).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get resource permissions: %w", err)
	}

	var permissions []*permission.Permission
	for _, model := range models {
		permissions = append(permissions, r.modelToPermission(&model))
	}

	return permissions, nil
}

// GetUserPermission 获取用户权限
func (r *PermissionRepository) GetUserPermission(ctx context.Context, userID, resourceType, resourceID string) (*permission.Permission, error) {
	var model models.Permission
	if err := r.db.WithContext(ctx).Where("user_id = ? AND resource_type = ? AND resource_id = ?", userID, resourceType, resourceID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, permission.ErrPermissionNotFound
		}
		return nil, fmt.Errorf("failed to get user permission: %w", err)
	}

	return r.modelToPermission(&model), nil
}

// UpdatePermission 更新权限
func (r *PermissionRepository) UpdatePermission(ctx context.Context, perm *permission.Permission) error {
	model := &models.Permission{
		ID:           perm.ID,
		UserID:       perm.UserID,
		ResourceType: perm.ResourceType,
		ResourceID:   perm.ResourceID,
		Role:         string(perm.Role),
		GrantedBy:    perm.GrantedBy,
		GrantedAt:    perm.GrantedAt,
		ExpiresAt:    perm.ExpiresAt,
		IsActive:     perm.IsActive,
		CreatedAt:    perm.CreatedAt,
		UpdatedAt:    perm.UpdatedAt,
	}

	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		r.logger.Error("Failed to update permission",
			logger.String("permission_id", perm.ID),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to update permission: %w", err)
	}

	return nil
}

// DeletePermission 删除权限
func (r *PermissionRepository) DeletePermission(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Permission{}).Error; err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}
	return nil
}

// DeleteUserPermissions 删除用户权限
func (r *PermissionRepository) DeleteUserPermissions(ctx context.Context, userID, resourceType, resourceID string) error {
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if resourceType != "" {
		query = query.Where("resource_type = ?", resourceType)
	}
	if resourceID != "" {
		query = query.Where("resource_id = ?", resourceID)
	}

	if err := query.Delete(&models.Permission{}).Error; err != nil {
		return fmt.Errorf("failed to delete user permissions: %w", err)
	}
	return nil
}

// 空间协作者管理

// CreateSpaceCollaborator 创建空间协作者
func (r *PermissionRepository) CreateSpaceCollaborator(ctx context.Context, collaborator *permission.SpaceCollaborator) error {
	model := &models.SpaceCollaborator{
		ID:          collaborator.ID,
		SpaceID:     collaborator.SpaceID,
		UserID:      collaborator.UserID,
		Role:        string(collaborator.Role),
		CreatedTime: collaborator.CreatedAt,
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		r.logger.Error("Failed to create space collaborator",
			logger.String("space_id", collaborator.SpaceID),
			logger.String("user_id", collaborator.UserID),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to create space collaborator: %w", err)
	}

	return nil
}

// GetSpaceCollaborator 获取空间协作者
func (r *PermissionRepository) GetSpaceCollaborator(ctx context.Context, id string) (*permission.SpaceCollaborator, error) {
	var model models.SpaceCollaborator
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, permission.ErrCollaboratorNotFound
		}
		return nil, fmt.Errorf("failed to get space collaborator: %w", err)
	}

	return r.modelToSpaceCollaborator(&model), nil
}

// GetSpaceCollaborators 获取空间协作者列表
func (r *PermissionRepository) GetSpaceCollaborators(ctx context.Context, spaceID string) ([]*permission.SpaceCollaborator, error) {
	var models []models.SpaceCollaborator
	if err := r.db.WithContext(ctx).Where("space_id = ?", spaceID).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get space collaborators: %w", err)
	}

	var collaborators []*permission.SpaceCollaborator
	for _, model := range models {
		collaborators = append(collaborators, r.modelToSpaceCollaborator(&model))
	}

	return collaborators, nil
}

// GetUserSpaceCollaborator 获取用户空间协作者
func (r *PermissionRepository) GetUserSpaceCollaborator(ctx context.Context, userID, spaceID string) (*permission.SpaceCollaborator, error) {
	var model models.SpaceCollaborator
	if err := r.db.WithContext(ctx).Where("user_id = ? AND space_id = ?", userID, spaceID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, permission.ErrCollaboratorNotFound
		}
		return nil, fmt.Errorf("failed to get user space collaborator: %w", err)
	}

	return r.modelToSpaceCollaborator(&model), nil
}

// UpdateSpaceCollaborator 更新空间协作者
func (r *PermissionRepository) UpdateSpaceCollaborator(ctx context.Context, collaborator *permission.SpaceCollaborator) error {
	model := &models.SpaceCollaborator{
		ID:          collaborator.ID,
		SpaceID:     collaborator.SpaceID,
		UserID:      collaborator.UserID,
		Role:        string(collaborator.Role),
		CreatedTime: collaborator.CreatedAt,
	}

	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		r.logger.Error("Failed to update space collaborator",
			logger.String("collaborator_id", collaborator.ID),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to update space collaborator: %w", err)
	}

	return nil
}

// DeleteSpaceCollaborator 删除空间协作者
func (r *PermissionRepository) DeleteSpaceCollaborator(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.SpaceCollaborator{}).Error; err != nil {
		return fmt.Errorf("failed to delete space collaborator: %w", err)
	}
	return nil
}

// DeleteUserSpaceCollaborator 删除用户空间协作者
func (r *PermissionRepository) DeleteUserSpaceCollaborator(ctx context.Context, userID, spaceID string) error {
	if err := r.db.WithContext(ctx).Where("user_id = ? AND space_id = ?", userID, spaceID).Delete(&models.SpaceCollaborator{}).Error; err != nil {
		return fmt.Errorf("failed to delete user space collaborator: %w", err)
	}
	return nil
}

// 基础表协作者管理

// CreateBaseCollaborator 创建基础表协作者
func (r *PermissionRepository) CreateBaseCollaborator(ctx context.Context, collaborator *permission.BaseCollaborator) error {
	model := &models.BaseCollaborator{
		ID:        collaborator.ID,
		BaseID:    collaborator.BaseID,
		UserID:    collaborator.UserID,
		Role:      string(collaborator.Role),
		Email:     collaborator.Email,
		InvitedBy: collaborator.InvitedBy,
		InvitedAt: collaborator.InvitedAt,
		JoinedAt:  collaborator.JoinedAt,
		IsActive:  collaborator.IsActive,
		CreatedAt: collaborator.CreatedAt,
		UpdatedAt: collaborator.UpdatedAt,
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		r.logger.Error("Failed to create base collaborator",
			logger.String("base_id", collaborator.BaseID),
			logger.String("user_id", collaborator.UserID),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to create base collaborator: %w", err)
	}

	return nil
}

// GetBaseCollaborator 获取基础表协作者
func (r *PermissionRepository) GetBaseCollaborator(ctx context.Context, id string) (*permission.BaseCollaborator, error) {
	var model models.BaseCollaborator
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, permission.ErrCollaboratorNotFound
		}
		return nil, fmt.Errorf("failed to get base collaborator: %w", err)
	}

	return r.modelToBaseCollaborator(&model), nil
}

// GetBaseCollaborators 获取基础表协作者列表
func (r *PermissionRepository) GetBaseCollaborators(ctx context.Context, baseID string) ([]*permission.BaseCollaborator, error) {
	var models []models.BaseCollaborator
	if err := r.db.WithContext(ctx).Where("base_id = ?", baseID).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get base collaborators: %w", err)
	}

	var collaborators []*permission.BaseCollaborator
	for _, model := range models {
		collaborators = append(collaborators, r.modelToBaseCollaborator(&model))
	}

	return collaborators, nil
}

// GetUserBaseCollaborator 获取用户基础表协作者
func (r *PermissionRepository) GetUserBaseCollaborator(ctx context.Context, userID, baseID string) (*permission.BaseCollaborator, error) {
	var model models.BaseCollaborator
	if err := r.db.WithContext(ctx).Where("user_id = ? AND base_id = ?", userID, baseID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, permission.ErrCollaboratorNotFound
		}
		return nil, fmt.Errorf("failed to get user base collaborator: %w", err)
	}

	return r.modelToBaseCollaborator(&model), nil
}

// UpdateBaseCollaborator 更新基础表协作者
func (r *PermissionRepository) UpdateBaseCollaborator(ctx context.Context, collaborator *permission.BaseCollaborator) error {
	model := &models.BaseCollaborator{
		ID:        collaborator.ID,
		BaseID:    collaborator.BaseID,
		UserID:    collaborator.UserID,
		Role:      string(collaborator.Role),
		Email:     collaborator.Email,
		InvitedBy: collaborator.InvitedBy,
		InvitedAt: collaborator.InvitedAt,
		JoinedAt:  collaborator.JoinedAt,
		IsActive:  collaborator.IsActive,
		CreatedAt: collaborator.CreatedAt,
		UpdatedAt: collaborator.UpdatedAt,
	}

	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		r.logger.Error("Failed to update base collaborator",
			logger.String("collaborator_id", collaborator.ID),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to update base collaborator: %w", err)
	}

	return nil
}

// DeleteBaseCollaborator 删除基础表协作者
func (r *PermissionRepository) DeleteBaseCollaborator(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.BaseCollaborator{}).Error; err != nil {
		return fmt.Errorf("failed to delete base collaborator: %w", err)
	}
	return nil
}

// DeleteUserBaseCollaborator 删除用户基础表协作者
func (r *PermissionRepository) DeleteUserBaseCollaborator(ctx context.Context, userID, baseID string) error {
	if err := r.db.WithContext(ctx).Where("user_id = ? AND base_id = ?", userID, baseID).Delete(&models.BaseCollaborator{}).Error; err != nil {
		return fmt.Errorf("failed to delete user base collaborator: %w", err)
	}
	return nil
}

// 批量操作

// GetUserRoles 获取用户角色
func (r *PermissionRepository) GetUserRoles(ctx context.Context, userID string) (map[string]permission.Role, error) {
	var permissions []models.Permission
	if err := r.db.WithContext(ctx).Where("user_id = ? AND is_active = ?", userID, true).Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	roles := make(map[string]permission.Role)
	for _, perm := range permissions {
		roles[perm.ResourceType] = permission.Role(perm.Role)
	}

	return roles, nil
}

// GetResourceCollaborators 获取资源协作者
func (r *PermissionRepository) GetResourceCollaborators(ctx context.Context, resourceType, resourceID string) ([]*permission.CollaboratorInfo, error) {
	var collaborators []*permission.CollaboratorInfo

	switch resourceType {
	case "space":
		var models []models.SpaceCollaborator
		if err := r.db.WithContext(ctx).Where("space_id = ?", resourceID).Find(&models).Error; err != nil {
			return nil, fmt.Errorf("failed to get space collaborators: %w", err)
		}

		for _, model := range models {
			collaborators = append(collaborators, &permission.CollaboratorInfo{
				ID:        model.ID,
				UserID:    model.UserID,
				Role:      permission.Role(model.Role),
				Email:     nil, // 空间协作者模型中没有Email字段
				InvitedBy: "",  // 空间协作者模型中没有InvitedBy字段
				InvitedAt: time.Time{}, // 空间协作者模型中没有InvitedAt字段
				JoinedAt:  nil, // 空间协作者模型中没有JoinedAt字段
				IsActive:  true, // 空间协作者模型中没有IsActive字段，默认为true
			})
		}

	case "base":
		var models []models.BaseCollaborator
		if err := r.db.WithContext(ctx).Where("base_id = ? AND is_active = ?", resourceID, true).Find(&models).Error; err != nil {
			return nil, fmt.Errorf("failed to get base collaborators: %w", err)
		}

		for _, model := range models {
			collaborators = append(collaborators, &permission.CollaboratorInfo{
				ID:        model.ID,
				UserID:    model.UserID,
				Role:      permission.Role(model.Role),
				Email:     model.Email,
				InvitedBy: model.InvitedBy,
				InvitedAt: model.InvitedAt,
				JoinedAt:  model.JoinedAt,
				IsActive:  model.IsActive,
			})
		}

	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	return collaborators, nil
}

// GetUserResources 获取用户资源
func (r *PermissionRepository) GetUserResources(ctx context.Context, userID, resourceType string) ([]string, error) {
	var permissions []models.Permission
	query := r.db.WithContext(ctx).Where("user_id = ? AND is_active = ?", userID, true)
	if resourceType != "" {
		query = query.Where("resource_type = ?", resourceType)
	}

	if err := query.Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get user resources: %w", err)
	}

	var resources []string
	for _, perm := range permissions {
		resources = append(resources, perm.ResourceID)
	}

	return resources, nil
}

// 权限检查

// CheckUserPermission 检查用户权限
func (r *PermissionRepository) CheckUserPermission(ctx context.Context, userID, resourceType, resourceID string, action permission.Action) (bool, error) {
	// 获取用户有效角色
	role, err := r.GetUserEffectiveRole(ctx, userID, resourceType, resourceID)
	if err != nil {
		return false, err
	}

	// 检查角色是否有权限
	return permission.HasRolePermission(role, action), nil
}

// GetUserEffectiveRole 获取用户有效角色
func (r *PermissionRepository) GetUserEffectiveRole(ctx context.Context, userID, resourceType, resourceID string) (permission.Role, error) {
	// 首先检查直接权限
	var perm models.Permission
	if err := r.db.WithContext(ctx).Where("user_id = ? AND resource_type = ? AND resource_id = ? AND is_active = ?", userID, resourceType, resourceID, true).First(&perm).Error; err == nil {
		// 检查权限是否过期
		if perm.ExpiresAt == nil || time.Now().Before(*perm.ExpiresAt) {
			return permission.Role(perm.Role), nil
		}
	}

	// 如果没有直接权限，检查继承权限
	switch resourceType {
	case "base":
		// 检查空间权限
		// 这里需要查询基础表所属的空间，暂时返回空角色
		return "", permission.ErrPermissionNotFound
	case "table", "view", "field", "record":
		// 检查基础表权限
		// 这里需要查询表格所属的基础表，暂时返回空角色
		return "", permission.ErrPermissionNotFound
	}

	return "", permission.ErrPermissionNotFound
}

// 统计

// GetPermissionStats 获取权限统计
func (r *PermissionRepository) GetPermissionStats(ctx context.Context) (*permission.PermissionStats, error) {
	stats := &permission.PermissionStats{
		RoleDistribution: make(map[permission.Role]int64),
	}

	// 权限统计
	if err := r.db.WithContext(ctx).Model(&models.Permission{}).Count(&stats.TotalPermissions).Error; err != nil {
		return nil, fmt.Errorf("failed to count total permissions: %w", err)
	}

	if err := r.db.WithContext(ctx).Model(&models.Permission{}).Where("is_active = ?", true).Count(&stats.ActivePermissions).Error; err != nil {
		return nil, fmt.Errorf("failed to count active permissions: %w", err)
	}

	if err := r.db.WithContext(ctx).Model(&models.Permission{}).Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).Count(&stats.ExpiredPermissions).Error; err != nil {
		return nil, fmt.Errorf("failed to count expired permissions: %w", err)
	}

	// 空间协作者统计
	if err := r.db.WithContext(ctx).Model(&models.SpaceCollaborator{}).Count(&stats.TotalSpaceCollaborators).Error; err != nil {
		return nil, fmt.Errorf("failed to count total space collaborators: %w", err)
	}

	if err := r.db.WithContext(ctx).Model(&models.SpaceCollaborator{}).Where("is_active = ?", true).Count(&stats.ActiveSpaceCollaborators).Error; err != nil {
		return nil, fmt.Errorf("failed to count active space collaborators: %w", err)
	}

	// 基础表协作者统计
	if err := r.db.WithContext(ctx).Model(&models.BaseCollaborator{}).Count(&stats.TotalBaseCollaborators).Error; err != nil {
		return nil, fmt.Errorf("failed to count total base collaborators: %w", err)
	}

	if err := r.db.WithContext(ctx).Model(&models.BaseCollaborator{}).Where("is_active = ?", true).Count(&stats.ActiveBaseCollaborators).Error; err != nil {
		return nil, fmt.Errorf("failed to count active base collaborators: %w", err)
	}

	// 角色分布统计
	var roleCounts []struct {
		Role  string `json:"role"`
		Count int64  `json:"count"`
	}

	if err := r.db.WithContext(ctx).Model(&models.Permission{}).Select("role, COUNT(*) as count").Where("is_active = ?", true).Group("role").Scan(&roleCounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get role distribution: %w", err)
	}

	for _, rc := range roleCounts {
		stats.RoleDistribution[permission.Role(rc.Role)] = rc.Count
	}

	return stats, nil
}

// 辅助方法

// modelToPermission 模型转权限实体
func (r *PermissionRepository) modelToPermission(model *models.Permission) *permission.Permission {
	return &permission.Permission{
		ID:           model.ID,
		UserID:       model.UserID,
		ResourceType: model.ResourceType,
		ResourceID:   model.ResourceID,
		Role:         permission.Role(model.Role),
		GrantedBy:    model.GrantedBy,
		GrantedAt:    model.GrantedAt,
		ExpiresAt:    model.ExpiresAt,
		IsActive:     model.IsActive,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}
}

// modelToSpaceCollaborator 模型转空间协作者实体
func (r *PermissionRepository) modelToSpaceCollaborator(model *models.SpaceCollaborator) *permission.SpaceCollaborator {
	return &permission.SpaceCollaborator{
		ID:        model.ID,
		SpaceID:   model.SpaceID,
		UserID:    model.UserID,
		Role:      permission.Role(model.Role),
		Email:     nil, // 空间协作者模型中没有Email字段
		InvitedBy: "",  // 空间协作者模型中没有InvitedBy字段
		InvitedAt: time.Time{}, // 空间协作者模型中没有InvitedAt字段
		JoinedAt:  nil, // 空间协作者模型中没有JoinedAt字段
		IsActive:  true, // 空间协作者模型中没有IsActive字段，默认为true
		CreatedAt: model.CreatedTime,
		UpdatedAt: model.CreatedTime, // 空间协作者模型中没有UpdatedAt字段，使用CreatedTime
	}
}

// modelToBaseCollaborator 模型转基础表协作者实体
func (r *PermissionRepository) modelToBaseCollaborator(model *models.BaseCollaborator) *permission.BaseCollaborator {
	return &permission.BaseCollaborator{
		ID:        model.ID,
		BaseID:    model.BaseID,
		UserID:    model.UserID,
		Role:      permission.Role(model.Role),
		Email:     model.Email,
		InvitedBy: model.InvitedBy,
		InvitedAt: model.InvitedAt,
		JoinedAt:  model.JoinedAt,
		IsActive:  model.IsActive,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}
