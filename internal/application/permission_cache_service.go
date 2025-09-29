package application

import (
	"context"
	"sync"
	"time"

	"teable-go-backend/internal/domain/permission"
	"teable-go-backend/internal/infrastructure/cache"
	"teable-go-backend/pkg/logger"
)

// PermissionCacheService 权限缓存服务
type PermissionCacheService struct {
	permissionService *PermissionService
	cacheService      cache.CacheService
	warmupInterval    time.Duration
	stopChan          chan struct{}
	wg                sync.WaitGroup
}

// NewPermissionCacheService 创建权限缓存服务
func NewPermissionCacheService(
	permissionService *PermissionService,
	cacheService cache.CacheService,
) *PermissionCacheService {
	return &PermissionCacheService{
		permissionService: permissionService,
		cacheService:      cacheService,
		warmupInterval:    30 * time.Minute, // 每30分钟预热一次
		stopChan:          make(chan struct{}),
	}
}

// CacheWarmupConfig 缓存预热配置
type CacheWarmupConfig struct {
	UserIDs       []string            `json:"user_ids"`
	ResourceTypes []string            `json:"resource_types"`
	CommonActions []permission.Action `json:"common_actions"`
	BatchSize     int                 `json:"batch_size"`
}

// Start 启动缓存预热服务
func (s *PermissionCacheService) Start(ctx context.Context) {
	s.wg.Add(1)
	go s.warmupWorker(ctx)

	logger.Info("Permission cache service started",
		logger.Duration("warmup_interval", s.warmupInterval),
	)
}

// Stop 停止缓存预热服务
func (s *PermissionCacheService) Stop() {
	close(s.stopChan)
	s.wg.Wait()

	logger.Info("Permission cache service stopped")
}

// WarmupUserPermissions 预热用户权限
func (s *PermissionCacheService) WarmupUserPermissions(ctx context.Context, config *CacheWarmupConfig) error {
	if config.BatchSize <= 0 {
		config.BatchSize = 10
	}

	// 默认资源类型
	if len(config.ResourceTypes) == 0 {
		config.ResourceTypes = []string{"space", "base", "table", "view"}
	}

	// 默认常用操作
	if len(config.CommonActions) == 0 {
		config.CommonActions = []permission.Action{
			permission.ActionSpaceRead,
			permission.ActionBaseRead,
			permission.ActionTableRead,
			permission.ActionRecordRead,
			permission.ActionRecordCreate,
			permission.ActionRecordUpdate,
			permission.ActionViewRead,
		}
	}

	// 分批处理用户
	for i := 0; i < len(config.UserIDs); i += config.BatchSize {
		end := i + config.BatchSize
		if end > len(config.UserIDs) {
			end = len(config.UserIDs)
		}

		batch := config.UserIDs[i:end]
		if err := s.warmupUserBatch(ctx, batch, config); err != nil {
			logger.Error("Failed to warmup user batch",
				logger.Int("batch_start", i),
				logger.Int("batch_end", end),
				logger.ErrorField(err),
			)
		}

		// 检查是否需要停止
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.stopChan:
			return nil
		default:
		}
	}

	logger.Info("User permissions warmup completed",
		logger.Int("total_users", len(config.UserIDs)),
		logger.Strings("resource_types", config.ResourceTypes),
	)

	return nil
}

// WarmupResourcePermissions 预热资源权限
func (s *PermissionCacheService) WarmupResourcePermissions(ctx context.Context, resourceType string, resourceIDs []string) error {
	for _, resourceID := range resourceIDs {
		// 获取资源的所有协作者
		collaborators, err := s.getResourceCollaborators(ctx, resourceType, resourceID)
		if err != nil {
			logger.Warn("Failed to get resource collaborators",
				logger.String("resource_type", resourceType),
				logger.String("resource_id", resourceID),
				logger.ErrorField(err),
			)
			continue
		}

		// 预热每个协作者的权限
		for _, userID := range collaborators {
			_, err := s.permissionService.GetUserResourcePermissions(ctx, userID, resourceType, resourceID)
			if err != nil {
				logger.Warn("Failed to warmup user resource permissions",
					logger.String("user_id", userID),
					logger.String("resource_type", resourceType),
					logger.String("resource_id", resourceID),
					logger.ErrorField(err),
				)
			}
		}

		// 检查是否需要停止
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.stopChan:
			return nil
		default:
		}
	}

	logger.Info("Resource permissions warmup completed",
		logger.String("resource_type", resourceType),
		logger.Int("resource_count", len(resourceIDs)),
	)

	return nil
}

// GetCacheStats 获取缓存统计信息
func (s *PermissionCacheService) GetCacheStats(ctx context.Context) (*CacheStats, error) {
	// 获取权限相关的缓存键数量
	permissionKeys, err := s.countCacheKeys(ctx, "permission:*")
	if err != nil {
		return nil, err
	}

	userResourceKeys, err := s.countCacheKeys(ctx, "user_resources:*")
	if err != nil {
		return nil, err
	}

	userPermissionKeys, err := s.countCacheKeys(ctx, "user_resource_permissions:*")
	if err != nil {
		return nil, err
	}

	return &CacheStats{
		PermissionKeys:     permissionKeys,
		UserResourceKeys:   userResourceKeys,
		UserPermissionKeys: userPermissionKeys,
		TotalKeys:          permissionKeys + userResourceKeys + userPermissionKeys,
		LastWarmupTime:     time.Now(), // 这里应该记录实际的预热时间
		CacheHitRate:       0.0,        // 这里应该从缓存服务获取命中率
	}, nil
}

// ClearExpiredCache 清理过期缓存
func (s *PermissionCacheService) ClearExpiredCache(ctx context.Context) error {
	patterns := []string{
		"permission:*",
		"user_resources:*",
		"user_resource_permissions:*",
	}

	for _, pattern := range patterns {
		if err := s.cacheService.DeletePattern(ctx, pattern); err != nil {
			logger.Error("Failed to clear expired cache",
				logger.String("pattern", pattern),
				logger.ErrorField(err),
			)
		}
	}

	logger.Info("Expired permission cache cleared")
	return nil
}

// 私有方法

func (s *PermissionCacheService) warmupWorker(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(s.warmupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.performScheduledWarmup(ctx)
		}
	}
}

func (s *PermissionCacheService) performScheduledWarmup(ctx context.Context) {
	logger.Info("Starting scheduled permission cache warmup")

	// 获取活跃用户列表（这里应该从用户服务获取）
	activeUsers := s.getActiveUsers(ctx)
	if len(activeUsers) == 0 {
		logger.Info("No active users found for cache warmup")
		return
	}

	// 执行预热
	config := &CacheWarmupConfig{
		UserIDs:       activeUsers,
		ResourceTypes: []string{"space", "base", "table"},
		BatchSize:     20,
	}

	if err := s.WarmupUserPermissions(ctx, config); err != nil {
		logger.Error("Scheduled permission cache warmup failed", logger.ErrorField(err))
	} else {
		logger.Info("Scheduled permission cache warmup completed",
			logger.Int("user_count", len(activeUsers)),
		)
	}
}

func (s *PermissionCacheService) warmupUserBatch(ctx context.Context, userIDs []string, config *CacheWarmupConfig) error {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5) // 限制并发数

	for _, userID := range userIDs {
		wg.Add(1)
		go func(uid string) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 预热用户的资源权限
			for _, resourceType := range config.ResourceTypes {
				resources, err := s.permissionService.GetUserAccessibleResources(ctx, uid, resourceType)
				if err != nil {
					logger.Warn("Failed to get user accessible resources",
						logger.String("user_id", uid),
						logger.String("resource_type", resourceType),
						logger.ErrorField(err),
					)
					continue
				}

				// 预热前几个资源的权限
				maxResources := 10
				if len(resources) > maxResources {
					resources = resources[:maxResources]
				}

				for _, resourceID := range resources {
					_, err := s.permissionService.GetUserResourcePermissions(ctx, uid, resourceType, resourceID)
					if err != nil {
						logger.Warn("Failed to warmup user resource permissions",
							logger.String("user_id", uid),
							logger.String("resource_type", resourceType),
							logger.String("resource_id", resourceID),
							logger.ErrorField(err),
						)
					}
				}
			}
		}(userID)
	}

	wg.Wait()
	return nil
}

func (s *PermissionCacheService) getActiveUsers(ctx context.Context) []string {
	// 这里应该从用户服务获取活跃用户列表
	// 暂时返回空列表
	return []string{}
}

func (s *PermissionCacheService) getResourceCollaborators(ctx context.Context, resourceType, resourceID string) ([]string, error) {
	// 这里应该从权限服务获取资源协作者
	// 暂时返回空列表
	return []string{}, nil
}

func (s *PermissionCacheService) countCacheKeys(ctx context.Context, pattern string) (int64, error) {
	// 这里应该从缓存服务获取匹配模式的键数量
	// 暂时返回0
	return 0, nil
}

// CacheStats 缓存统计信息
type CacheStats struct {
	PermissionKeys     int64     `json:"permission_keys"`
	UserResourceKeys   int64     `json:"user_resource_keys"`
	UserPermissionKeys int64     `json:"user_permission_keys"`
	TotalKeys          int64     `json:"total_keys"`
	LastWarmupTime     time.Time `json:"last_warmup_time"`
	CacheHitRate       float64   `json:"cache_hit_rate"`
}
