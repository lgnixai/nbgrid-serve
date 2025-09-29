package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"teable-go-backend/internal/domain/permission"
	"teable-go-backend/internal/infrastructure/cache"
	"teable-go-backend/pkg/logger"
)

// PermissionService 权限应用服务
type PermissionService struct {
	permissionDomainService permission.Service
	cacheService            cache.CacheService
}

// NewPermissionService 创建权限服务
func NewPermissionService(
	permissionDomainService permission.Service,
	cacheService cache.CacheService,
) *PermissionService {
	return &PermissionService{
		permissionDomainService: permissionDomainService,
		cacheService:            cacheService,
	}
}

// PermissionCheckRequest 权限检查请求
type PermissionCheckRequest struct {
	UserID       string                 `json:"user_id"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   string                 `json:"resource_id"`
	Action       permission.Action      `json:"action"`
	Actions      []permission.Action    `json:"actions,omitempty"`
	Context      map[string]interface{} `json:"context,omitempty"`
}

// PermissionCheckResponse 权限检查响应
type PermissionCheckResponse struct {
	Allowed     bool                `json:"allowed"`
	Role        permission.Role     `json:"role,omitempty"`
	Permissions []permission.Action `json:"permissions,omitempty"`
	Reason      string              `json:"reason,omitempty"`
	CachedAt    *time.Time          `json:"cached_at,omitempty"`
}

// BatchPermissionCheckRequest 批量权限检查请求
type BatchPermissionCheckRequest struct {
	UserID   string                   `json:"user_id"`
	Requests []PermissionCheckRequest `json:"requests"`
}

// BatchPermissionCheckResponse 批量权限检查响应
type BatchPermissionCheckResponse struct {
	Results map[string]*PermissionCheckResponse `json:"results"`
}

// CheckPermission 检查单个权限（带缓存）
func (s *PermissionService) CheckPermission(ctx context.Context, req *PermissionCheckRequest) (*PermissionCheckResponse, error) {
	// 构建缓存键
	cacheKey := s.buildPermissionCacheKey(req.UserID, req.ResourceType, req.ResourceID, req.Action)

	// 尝试从缓存获取
	var cachedResult PermissionCheckResponse
	if err := s.cacheService.Get(ctx, cacheKey, &cachedResult); err == nil {
		logger.Debug("Permission check cache hit",
			logger.String("user_id", req.UserID),
			logger.String("resource_type", req.ResourceType),
			logger.String("resource_id", req.ResourceID),
			logger.String("action", string(req.Action)),
		)
		return &cachedResult, nil
	}

	// 缓存未命中，执行权限检查
	allowed, err := s.permissionDomainService.CheckPermission(ctx, req.UserID, req.ResourceType, req.ResourceID, req.Action)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}

	// 获取用户角色
	role, err := s.permissionDomainService.GetUserEffectiveRole(ctx, req.UserID, req.ResourceType, req.ResourceID)
	if err != nil {
		logger.Warn("Failed to get user effective role",
			logger.String("user_id", req.UserID),
			logger.String("resource_type", req.ResourceType),
			logger.String("resource_id", req.ResourceID),
			logger.ErrorField(err),
		)
		role = ""
	}

	result := &PermissionCheckResponse{
		Allowed: allowed,
		Role:    role,
		Reason:  s.getPermissionReason(allowed, role, req.Action),
	}

	// 缓存结果（短时间缓存，避免权限变更延迟）
	cacheTTL := 5 * time.Minute
	if err := s.cacheService.Set(ctx, cacheKey, result, cacheTTL); err != nil {
		logger.Warn("Failed to cache permission result",
			logger.String("cache_key", cacheKey),
			logger.ErrorField(err),
		)
	}

	logger.Debug("Permission check completed",
		logger.String("user_id", req.UserID),
		logger.String("resource_type", req.ResourceType),
		logger.String("resource_id", req.ResourceID),
		logger.String("action", string(req.Action)),
		logger.Bool("allowed", allowed),
		logger.String("role", string(role)),
	)

	return result, nil
}

// CheckMultiplePermissions 检查多个权限
func (s *PermissionService) CheckMultiplePermissions(ctx context.Context, req *PermissionCheckRequest) (*PermissionCheckResponse, error) {
	if len(req.Actions) == 0 {
		return s.CheckPermission(ctx, req)
	}

	// 检查所有权限
	allowed, err := s.permissionDomainService.CheckMultiplePermissions(ctx, req.UserID, req.ResourceType, req.ResourceID, req.Actions)
	if err != nil {
		return nil, fmt.Errorf("failed to check multiple permissions: %w", err)
	}

	// 获取用户角色和权限
	role, err := s.permissionDomainService.GetUserEffectiveRole(ctx, req.UserID, req.ResourceType, req.ResourceID)
	if err != nil {
		logger.Warn("Failed to get user effective role",
			logger.String("user_id", req.UserID),
			logger.String("resource_type", req.ResourceType),
			logger.String("resource_id", req.ResourceID),
			logger.ErrorField(err),
		)
		role = ""
	}

	permissions, err := s.permissionDomainService.GetUserEffectivePermissions(ctx, req.UserID, req.ResourceType, req.ResourceID)
	if err != nil {
		logger.Warn("Failed to get user effective permissions",
			logger.String("user_id", req.UserID),
			logger.String("resource_type", req.ResourceType),
			logger.String("resource_id", req.ResourceID),
			logger.ErrorField(err),
		)
		permissions = []permission.Action{}
	}

	return &PermissionCheckResponse{
		Allowed:     allowed,
		Role:        role,
		Permissions: permissions,
		Reason:      s.getMultiplePermissionReason(allowed, role, req.Actions),
	}, nil
}

// BatchCheckPermissions 批量检查权限
func (s *PermissionService) BatchCheckPermissions(ctx context.Context, req *BatchPermissionCheckRequest) (*BatchPermissionCheckResponse, error) {
	results := make(map[string]*PermissionCheckResponse)

	for i, checkReq := range req.Requests {
		// 确保用户ID一致
		checkReq.UserID = req.UserID

		key := fmt.Sprintf("%d", i)
		if checkReq.ResourceType != "" && checkReq.ResourceID != "" {
			key = fmt.Sprintf("%s:%s", checkReq.ResourceType, checkReq.ResourceID)
		}

		var result *PermissionCheckResponse
		var err error

		if len(checkReq.Actions) > 0 {
			result, err = s.CheckMultiplePermissions(ctx, &checkReq)
		} else {
			result, err = s.CheckPermission(ctx, &checkReq)
		}

		if err != nil {
			logger.Error("Failed to check permission in batch",
				logger.String("user_id", req.UserID),
				logger.String("resource_type", checkReq.ResourceType),
				logger.String("resource_id", checkReq.ResourceID),
				logger.ErrorField(err),
			)
			result = &PermissionCheckResponse{
				Allowed: false,
				Reason:  "Permission check failed",
			}
		}

		results[key] = result
	}

	return &BatchPermissionCheckResponse{
		Results: results,
	}, nil
}

// GetUserResourcePermissions 获取用户对资源的所有权限
func (s *PermissionService) GetUserResourcePermissions(ctx context.Context, userID, resourceType, resourceID string) (*PermissionCheckResponse, error) {
	// 构建缓存键
	cacheKey := s.buildUserResourcePermissionsCacheKey(userID, resourceType, resourceID)

	// 尝试从缓存获取
	var cachedResult PermissionCheckResponse
	if err := s.cacheService.Get(ctx, cacheKey, &cachedResult); err == nil {
		return &cachedResult, nil
	}

	// 获取用户角色
	role, err := s.permissionDomainService.GetUserEffectiveRole(ctx, userID, resourceType, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user effective role: %w", err)
	}

	// 获取用户权限
	permissions, err := s.permissionDomainService.GetUserEffectivePermissions(ctx, userID, resourceType, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user effective permissions: %w", err)
	}

	result := &PermissionCheckResponse{
		Allowed:     len(permissions) > 0,
		Role:        role,
		Permissions: permissions,
		Reason:      fmt.Sprintf("User has %s role with %d permissions", role, len(permissions)),
	}

	// 缓存结果
	cacheTTL := 10 * time.Minute
	if err := s.cacheService.Set(ctx, cacheKey, result, cacheTTL); err != nil {
		logger.Warn("Failed to cache user resource permissions",
			logger.String("cache_key", cacheKey),
			logger.ErrorField(err),
		)
	}

	return result, nil
}

// GetUserAccessibleResources 获取用户可访问的资源列表
func (s *PermissionService) GetUserAccessibleResources(ctx context.Context, userID, resourceType string) ([]string, error) {
	// 构建缓存键
	cacheKey := s.buildUserResourcesCacheKey(userID, resourceType)

	// 尝试从缓存获取
	var cachedResources []string
	if err := s.cacheService.Get(ctx, cacheKey, &cachedResources); err == nil {
		return cachedResources, nil
	}

	// 从领域服务获取
	resources, err := s.permissionDomainService.GetUserResources(ctx, userID, resourceType)
	if err != nil {
		return nil, fmt.Errorf("failed to get user resources: %w", err)
	}

	// 缓存结果
	cacheTTL := 15 * time.Minute
	if err := s.cacheService.Set(ctx, cacheKey, resources, cacheTTL); err != nil {
		logger.Warn("Failed to cache user resources",
			logger.String("cache_key", cacheKey),
			logger.ErrorField(err),
		)
	}

	return resources, nil
}

// InvalidateUserPermissionCache 使用户权限缓存失效
func (s *PermissionService) InvalidateUserPermissionCache(ctx context.Context, userID string) error {
	// 构建缓存键模式
	patterns := []string{
		s.buildUserPermissionCachePattern(userID),
		s.buildUserResourcesCachePattern(userID),
	}

	for _, pattern := range patterns {
		if err := s.cacheService.DeletePattern(ctx, pattern); err != nil {
			logger.Warn("Failed to invalidate permission cache",
				logger.String("user_id", userID),
				logger.String("pattern", pattern),
				logger.ErrorField(err),
			)
		}
	}

	logger.Info("User permission cache invalidated",
		logger.String("user_id", userID),
	)

	return nil
}

// InvalidateResourcePermissionCache 使资源权限缓存失效
func (s *PermissionService) InvalidateResourcePermissionCache(ctx context.Context, resourceType, resourceID string) error {
	// 构建缓存键模式
	pattern := s.buildResourcePermissionCachePattern(resourceType, resourceID)

	if err := s.cacheService.DeletePattern(ctx, pattern); err != nil {
		logger.Warn("Failed to invalidate resource permission cache",
			logger.String("resource_type", resourceType),
			logger.String("resource_id", resourceID),
			logger.String("pattern", pattern),
			logger.ErrorField(err),
		)
		return err
	}

	logger.Info("Resource permission cache invalidated",
		logger.String("resource_type", resourceType),
		logger.String("resource_id", resourceID),
	)

	return nil
}

// PreloadUserPermissions 预加载用户权限到缓存
func (s *PermissionService) PreloadUserPermissions(ctx context.Context, userID string, resourceTypes []string) error {
	for _, resourceType := range resourceTypes {
		// 获取用户可访问的资源
		resources, err := s.GetUserAccessibleResources(ctx, userID, resourceType)
		if err != nil {
			logger.Warn("Failed to preload user resources",
				logger.String("user_id", userID),
				logger.String("resource_type", resourceType),
				logger.ErrorField(err),
			)
			continue
		}

		// 预加载每个资源的权限
		for _, resourceID := range resources {
			_, err := s.GetUserResourcePermissions(ctx, userID, resourceType, resourceID)
			if err != nil {
				logger.Warn("Failed to preload resource permissions",
					logger.String("user_id", userID),
					logger.String("resource_type", resourceType),
					logger.String("resource_id", resourceID),
					logger.ErrorField(err),
				)
			}
		}
	}

	logger.Info("User permissions preloaded",
		logger.String("user_id", userID),
		logger.Strings("resource_types", resourceTypes),
	)

	return nil
}

// 私有方法

func (s *PermissionService) buildPermissionCacheKey(userID, resourceType, resourceID string, action permission.Action) string {
	return cache.BuildCacheKey("permission", fmt.Sprintf("%s:%s:%s:%s", userID, resourceType, resourceID, action))
}

func (s *PermissionService) buildUserResourcePermissionsCacheKey(userID, resourceType, resourceID string) string {
	return cache.BuildCacheKey("user_resource_permissions", fmt.Sprintf("%s:%s:%s", userID, resourceType, resourceID))
}

func (s *PermissionService) buildUserResourcesCacheKey(userID, resourceType string) string {
	return cache.BuildCacheKey("user_resources", fmt.Sprintf("%s:%s", userID, resourceType))
}

func (s *PermissionService) buildUserPermissionCachePattern(userID string) string {
	return cache.BuildCacheKey("permission", userID+":*")
}

func (s *PermissionService) buildUserResourcesCachePattern(userID string) string {
	return cache.BuildCacheKey("user_resources", userID+":*")
}

func (s *PermissionService) buildResourcePermissionCachePattern(resourceType, resourceID string) string {
	return cache.BuildCacheKey("permission", fmt.Sprintf("*:%s:%s:*", resourceType, resourceID))
}

func (s *PermissionService) getPermissionReason(allowed bool, role permission.Role, action permission.Action) string {
	if allowed {
		return fmt.Sprintf("User has %s role which allows %s action", role, action)
	}

	if role == "" {
		return "User has no role for this resource"
	}

	return fmt.Sprintf("User has %s role which does not allow %s action", role, action)
}

func (s *PermissionService) getMultiplePermissionReason(allowed bool, role permission.Role, actions []permission.Action) string {
	actionNames := make([]string, len(actions))
	for i, action := range actions {
		actionNames[i] = string(action)
	}

	if allowed {
		return fmt.Sprintf("User has %s role which allows all requested actions: %s", role, strings.Join(actionNames, ", "))
	}

	if role == "" {
		return "User has no role for this resource"
	}

	return fmt.Sprintf("User has %s role which does not allow some of the requested actions: %s", role, strings.Join(actionNames, ", "))
}
