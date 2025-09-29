package application

import (
	"context"
	"strings"

	userDomain "teable-go-backend/internal/domain/user"
	"teable-go-backend/internal/infrastructure/cache"
	"teable-go-backend/internal/infrastructure/database/models"
	"teable-go-backend/pkg/errors"
	"teable-go-backend/pkg/logger"
)

// AuthService 认证服务
type AuthService struct {
	tokenService      *TokenService
	userDomainService userDomain.Service
	cacheService      cache.CacheService
}

// NewAuthService 创建认证服务
func NewAuthService(
	tokenService *TokenService,
	userDomainService userDomain.Service,
	cacheService cache.CacheService,
) *AuthService {
	return &AuthService{
		tokenService:      tokenService,
		userDomainService: userDomainService,
		cacheService:      cacheService,
	}
}

// AuthenticatedUser 已认证的用户信息
type AuthenticatedUser struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	IsAdmin   bool   `json:"is_admin"`
	IsSystem  bool   `json:"is_system"`
	SessionID string `json:"session_id,omitempty"`
}

// ValidateToken 验证令牌并返回用户信息
func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*AuthenticatedUser, error) {
	// 验证令牌
	claims, err := s.tokenService.ValidateToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	// 检查令牌类型
	if claims.TokenType != "access" {
		return nil, errors.ErrInvalidToken
	}

	// 检查用户令牌是否已失效
	if claims.IssuedAt != nil {
		invalidated, err := s.tokenService.IsUserTokensInvalidated(ctx, claims.UserID, claims.IssuedAt.Time)
		if err != nil {
			logger.Error("Failed to check user token invalidation",
				logger.String("user_id", claims.UserID),
				logger.ErrorField(err),
			)
		} else if invalidated {
			return nil, errors.ErrInvalidToken
		}
	}

	// 从缓存获取用户信息
	user, err := s.getCachedUser(ctx, claims.UserID)
	if err != nil {
		// 缓存未命中，从数据库获取
		domainUser, err := s.userDomainService.GetUser(ctx, claims.UserID)
		if err != nil {
			return nil, err
		}

		// 检查用户状态
		if !domainUser.IsActive() {
			return nil, errors.ErrUserDeactivated
		}

		user = &AuthenticatedUser{
			ID:       domainUser.ID,
			Email:    domainUser.Email,
			Name:     domainUser.Name,
			IsAdmin:  domainUser.IsAdmin,
			IsSystem: domainUser.IsSystem,
		}

		// 缓存用户信息
		s.cacheUser(ctx, user)
	}

	user.SessionID = claims.SessionID
	return user, nil
}

// ValidateAPIKey 验证API密钥并返回用户信息
func (s *AuthService) ValidateAPIKey(ctx context.Context, apiKey string) (*AuthenticatedUser, error) {
	userID, err := s.tokenService.ValidateAPIKey(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	// 获取用户信息
	domainUser, err := s.userDomainService.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 检查用户状态
	if !domainUser.IsActive() {
		return nil, errors.ErrUserDeactivated
	}

	return &AuthenticatedUser{
		ID:       domainUser.ID,
		Email:    domainUser.Email,
		Name:     domainUser.Name,
		IsAdmin:  domainUser.IsAdmin,
		IsSystem: domainUser.IsSystem,
	}, nil
}

// GetUserFromToken 从令牌获取用户信息 (用于middleware接口兼容)
func (s *AuthService) GetUserFromToken(ctx context.Context, tokenString string) (*models.User, error) {
	authUser, err := s.ValidateToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	// 转换为models.User格式
	user := &models.User{
		ID:    authUser.ID,
		Name:  authUser.Name,
		Email: authUser.Email,
	}

	if authUser.IsAdmin {
		isAdmin := true
		user.IsAdmin = &isAdmin
	}

	if authUser.IsSystem {
		isSystem := true
		user.IsSystem = &isSystem
	}

	return user, nil
}

// ExtractToken 从请求中提取令牌
func (s *AuthService) ExtractToken(authHeader, queryToken, cookieToken string) (string, string) {
	// 从Authorization header获取
	if authHeader != "" {
		// Bearer token格式
		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer "), "bearer"
		}
		// API Key格式
		if strings.HasPrefix(authHeader, "ApiKey ") {
			return strings.TrimPrefix(authHeader, "ApiKey "), "apikey"
		}
		// 直接token格式
		return authHeader, "bearer"
	}

	// 从query参数获取
	if queryToken != "" {
		return queryToken, "bearer"
	}

	// 从cookie获取
	if cookieToken != "" {
		return cookieToken, "bearer"
	}

	return "", ""
}

// CheckPermission 检查用户权限
func (s *AuthService) CheckPermission(user *AuthenticatedUser, permission string) bool {
	// 系统用户拥有所有权限
	if user.IsSystem {
		return true
	}

	// 管理员权限检查
	if user.IsAdmin && isAdminPermission(permission) {
		return true
	}

	// 基础用户权限检查
	return isBasicUserPermission(permission)
}

// 私有方法

func (s *AuthService) getCachedUser(ctx context.Context, userID string) (*AuthenticatedUser, error) {
	key := cache.BuildCacheKey(cache.UserCachePrefix, userID)
	
	var user AuthenticatedUser
	if err := s.cacheService.Get(ctx, key, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *AuthService) cacheUser(ctx context.Context, user *AuthenticatedUser) {
	key := cache.BuildCacheKey(cache.UserCachePrefix, user.ID)
	
	if err := s.cacheService.Set(ctx, key, user, cache.DefaultTTL); err != nil {
		logger.Warn("Failed to cache user info",
			logger.String("user_id", user.ID),
			logger.ErrorField(err),
		)
	}
}

// isAdminPermission 检查是否为管理员权限
func isAdminPermission(permission string) bool {
	adminPermissions := map[string]bool{
		"user:manage":        true,
		"user:create":        true,
		"user:delete":        true,
		"user:promote":       true,
		"space:manage":       true,
		"space:delete":       true,
		"base:manage":        true,
		"base:delete":        true,
		"table:manage":       true,
		"system:config":      true,
		"system:monitor":     true,
		"permission:manage":  true,
	}

	return adminPermissions[permission]
}

// isBasicUserPermission 检查是否为基础用户权限
func isBasicUserPermission(permission string) bool {
	basicPermissions := map[string]bool{
		"user:read":         true,
		"user:update_self":  true,
		"space:read":        true,
		"space:create":      true,
		"base:read":         true,
		"base:create":       true,
		"table:read":        true,
		"table:create":      true,
		"record:read":       true,
		"record:create":     true,
		"record:update":     true,
		"record:delete":     true,
		"view:read":         true,
		"view:create":       true,
		"view:update":       true,
		"attachment:upload": true,
		"attachment:read":   true,
	}

	return basicPermissions[permission]
}