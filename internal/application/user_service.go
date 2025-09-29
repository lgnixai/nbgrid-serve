package application

import (
	"context"
	"time"

	userDomain "teable-go-backend/internal/domain/user"
	"teable-go-backend/internal/infrastructure/cache"
	"teable-go-backend/pkg/errors"
	"teable-go-backend/pkg/logger"
)

// UserService 用户应用服务
type UserService struct {
	userDomainService userDomain.Service
	tokenService      *TokenService
	sessionService    *SessionService
	cacheService      cache.CacheService
}

// NewUserService 创建用户应用服务
func NewUserService(
	userDomainService userDomain.Service,
	tokenService *TokenService,
	sessionService *SessionService,
	cacheService cache.CacheService,
) *UserService {
	return &UserService{
		userDomainService: userDomainService,
		tokenService:      tokenService,
		sessionService:    sessionService,
		cacheService:      cacheService,
	}
}

// AuthResponse 认证响应
type AuthResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token,omitempty"`
	ExpiresIn    int64         `json:"expires_in"`
	TokenType    string        `json:"token_type"`
	SessionID    string        `json:"session_id,omitempty"`
}

// LoginContext 登录上下文信息
type LoginContext struct {
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
	DeviceID  string `json:"device_id,omitempty"`
}

// UserResponse 用户响应
type UserResponse struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	Email            string     `json:"email"`
	Phone            *string    `json:"phone"`
	Avatar           *string    `json:"avatar"`
	IsAdmin          bool       `json:"is_admin"`
	IsSystem         bool       `json:"is_system"`
	IsTrialUsed      bool       `json:"is_trial_used"`
	LastSignTime     *time.Time `json:"last_sign_time"`
	CreatedTime      time.Time  `json:"created_time"`
	LastModifiedTime *time.Time `json:"last_modified_time"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Name     string  `json:"name" validate:"required,min=1,max=100"`
	Email    string  `json:"email" validate:"required,email,max=255"`
	Password string  `json:"password" validate:"required,min=8,max=128"`
	Phone    *string `json:"phone,omitempty" validate:"omitempty,max=50"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=128"`
}

// UpdateProfileRequest 更新资料请求
type UpdateProfileRequest struct {
	Name   *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Phone  *string `json:"phone,omitempty" validate:"omitempty,max=50"`
	Avatar *string `json:"avatar,omitempty" validate:"omitempty,url,max=500"`
}

// Register 用户注册
func (s *UserService) Register(ctx context.Context, req RegisterRequest, loginCtx *LoginContext) (*AuthResponse, error) {
	// 创建用户
	createReq := userDomain.CreateUserRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: &req.Password,
		Phone:    req.Phone,
	}

	domainUser, err := s.userDomainService.CreateUser(ctx, createReq)
	if err != nil {
		// 对于业务错误，只记录基本信息，不记录堆栈
		if appErr, ok := errors.IsAppError(err); ok && appErr.HTTPStatus < 500 {
			logger.Warn("User creation failed",
				logger.String("email", req.Email),
				logger.String("error_code", appErr.Code),
				logger.String("error_message", appErr.Message),
			)
		} else {
			logger.Error("Failed to create user",
				logger.String("email", req.Email),
				logger.ErrorField(err),
			)
		}
		return nil, err
	}

	// 创建用户会话
	session := &UserSession{
		UserID:       domainUser.ID,
		Email:        domainUser.Email,
		Name:         domainUser.Name,
		IsAdmin:      domainUser.IsAdmin,
		IsSystem:     domainUser.IsSystem,
		LoginTime:    time.Now(),
		LastActivity: time.Now(),
		IPAddress:    loginCtx.IPAddress,
		UserAgent:    loginCtx.UserAgent,
		DeviceID:     loginCtx.DeviceID,
	}

	sessionID, err := s.sessionService.CreateSession(ctx, domainUser.ID, session)
	if err != nil {
		logger.Error("Failed to create user session",
			logger.String("user_id", domainUser.ID),
			logger.ErrorField(err),
		)
		return nil, err
	}

	// 生成令牌对
	tokens, err := s.tokenService.GenerateTokenPair(ctx, domainUser, sessionID)
	if err != nil {
		logger.Error("Failed to generate tokens",
			logger.String("user_id", domainUser.ID),
			logger.ErrorField(err),
		)
		return nil, err
	}

	// 缓存用户信息
	if err := s.cacheUserInfo(ctx, domainUser); err != nil {
		logger.Warn("Failed to cache user info",
			logger.String("user_id", domainUser.ID),
			logger.ErrorField(err),
		)
	}

	logger.Info("User registered successfully",
		logger.String("user_id", domainUser.ID),
		logger.String("email", domainUser.Email),
		logger.String("session_id", sessionID),
	)

	return &AuthResponse{
		User:         s.toUserResponse(domainUser),
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		TokenType:    tokens.TokenType,
		SessionID:    sessionID,
	}, nil
}

// Login 用户登录
func (s *UserService) Login(ctx context.Context, req LoginRequest, loginCtx *LoginContext) (*AuthResponse, error) {
	// 认证用户
	domainUser, err := s.userDomainService.Authenticate(ctx, req.Email, req.Password)
	if err != nil {
		// 对于认证失败，只记录基本信息
		if appErr, ok := errors.IsAppError(err); ok && appErr.HTTPStatus < 500 {
			logger.Warn("Authentication failed",
				logger.String("email", req.Email),
				logger.String("error_code", appErr.Code),
			)
		} else {
			logger.Error("Login failed",
				logger.String("email", req.Email),
				logger.ErrorField(err),
			)
		}
		return nil, err
	}

	// 创建用户会话
	session := &UserSession{
		UserID:       domainUser.ID,
		Email:        domainUser.Email,
		Name:         domainUser.Name,
		IsAdmin:      domainUser.IsAdmin,
		IsSystem:     domainUser.IsSystem,
		LoginTime:    time.Now(),
		LastActivity: time.Now(),
		IPAddress:    loginCtx.IPAddress,
		UserAgent:    loginCtx.UserAgent,
		DeviceID:     loginCtx.DeviceID,
	}

	sessionID, err := s.sessionService.CreateSession(ctx, domainUser.ID, session)
	if err != nil {
		logger.Error("Failed to create user session",
			logger.String("user_id", domainUser.ID),
			logger.ErrorField(err),
		)
		return nil, err
	}

	// 生成令牌对
	tokens, err := s.tokenService.GenerateTokenPair(ctx, domainUser, sessionID)
	if err != nil {
		logger.Error("Failed to generate tokens",
			logger.String("user_id", domainUser.ID),
			logger.ErrorField(err),
		)
		return nil, err
	}

	// 缓存用户信息
	if err := s.cacheUserInfo(ctx, domainUser); err != nil {
		logger.Warn("Failed to cache user info",
			logger.String("user_id", domainUser.ID),
			logger.ErrorField(err),
		)
	}

	logger.Info("User logged in successfully",
		logger.String("user_id", domainUser.ID),
		logger.String("email", domainUser.Email),
		logger.String("session_id", sessionID),
	)

	return &AuthResponse{
		User:         s.toUserResponse(domainUser),
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		TokenType:    tokens.TokenType,
		SessionID:    sessionID,
	}, nil
}

// Logout 用户登出
func (s *UserService) Logout(ctx context.Context, userID, token, sessionID string) error {
	// 将令牌加入黑名单
	if err := s.tokenService.BlacklistToken(ctx, token); err != nil {
		logger.Error("Failed to blacklist token",
			logger.String("user_id", userID),
			logger.ErrorField(err),
		)
		return err
	}

	// 使会话失效
	if sessionID != "" {
		if err := s.sessionService.InvalidateSession(ctx, sessionID); err != nil {
			logger.Warn("Failed to invalidate session",
				logger.String("user_id", userID),
				logger.String("session_id", sessionID),
				logger.ErrorField(err),
			)
		}
	}

	// 清除用户缓存
	if err := s.clearUserCache(ctx, userID); err != nil {
		logger.Warn("Failed to clear user cache",
			logger.String("user_id", userID),
			logger.ErrorField(err),
		)
	}

	logger.Info("User logged out successfully",
		logger.String("user_id", userID),
		logger.String("session_id", sessionID),
	)

	return nil
}

// GetProfile 获取用户资料
func (s *UserService) GetProfile(ctx context.Context, userID string) (*UserResponse, error) {
	domainUser, err := s.userDomainService.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.toUserResponse(domainUser), nil
}

// UpdateProfile 更新用户资料
func (s *UserService) UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) (*UserResponse, error) {
	updateReq := userDomain.UpdateUserRequest{
		Name:   req.Name,
		Phone:  req.Phone,
		Avatar: req.Avatar,
	}

	domainUser, err := s.userDomainService.UpdateUser(ctx, userID, updateReq)
	if err != nil {
		logger.Error("Failed to update user profile",
			logger.String("user_id", userID),
			logger.ErrorField(err),
		)
		return nil, err
	}

	// 清除用户缓存
	if err := s.clearUserCache(ctx, userID); err != nil {
		logger.Warn("Failed to clear user cache after update",
			logger.String("user_id", userID),
			logger.ErrorField(err),
		)
	}

	logger.Info("User profile updated successfully",
		logger.String("user_id", userID),
	)

	return s.toUserResponse(domainUser), nil
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(ctx context.Context, userID string, req ChangePasswordRequest) error {
	err := s.userDomainService.ChangePassword(ctx, userID, req.OldPassword, req.NewPassword)
	if err != nil {
		logger.Error("Failed to change password",
			logger.String("user_id", userID),
			logger.ErrorField(err),
		)
		return err
	}

	logger.Info("Password changed successfully",
		logger.String("user_id", userID),
	)

	return nil
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshToken 刷新访问令牌
func (s *UserService) RefreshToken(ctx context.Context, req RefreshTokenRequest) (*AuthResponse, error) {
	// 验证refresh token
	claims, err := s.tokenService.ValidateToken(ctx, req.RefreshToken)
	if err != nil {
		logger.Warn("Invalid refresh token", logger.ErrorField(err))
		return nil, err
	}

	// 检查令牌类型
	if claims.TokenType != "refresh" {
		return nil, errors.ErrInvalidToken
	}

	// 获取用户信息
	domainUser, err := s.userDomainService.GetUser(ctx, claims.UserID)
	if err != nil {
		logger.Error("Failed to get user for token refresh",
			logger.String("user_id", claims.UserID),
			logger.ErrorField(err),
		)
		return nil, err
	}

	// 检查用户状态
	if !domainUser.IsActive() {
		return nil, errors.ErrUserDeactivated
	}

	// 更新会话活动时间
	if claims.SessionID != "" {
		if err := s.sessionService.UpdateSessionActivity(ctx, claims.SessionID); err != nil {
			logger.Warn("Failed to update session activity",
				logger.String("user_id", domainUser.ID),
				logger.String("session_id", claims.SessionID),
				logger.ErrorField(err),
			)
		}
	}

	// 生成新的令牌对
	tokens, err := s.tokenService.RefreshToken(ctx, req.RefreshToken, domainUser)
	if err != nil {
		logger.Error("Failed to refresh tokens",
			logger.String("user_id", domainUser.ID),
			logger.ErrorField(err),
		)
		return nil, err
	}

	// 缓存用户信息
	if err := s.cacheUserInfo(ctx, domainUser); err != nil {
		logger.Warn("Failed to cache user info",
			logger.String("user_id", domainUser.ID),
			logger.ErrorField(err),
		)
	}

	logger.Info("Token refreshed successfully",
		logger.String("user_id", domainUser.ID),
		logger.String("session_id", claims.SessionID),
	)

	return &AuthResponse{
		User:         s.toUserResponse(domainUser),
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		TokenType:    tokens.TokenType,
		SessionID:    claims.SessionID,
	}, nil
}

// GetUser 获取用户信息(管理员功能)
func (s *UserService) GetUser(ctx context.Context, userID string) (*UserResponse, error) {
	domainUser, err := s.userDomainService.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.toUserResponse(domainUser), nil
}

// ListUsers 列出用户(管理员功能)
func (s *UserService) ListUsers(ctx context.Context, filter userDomain.ListFilter) (*userDomain.PaginatedResult, error) {
	result, err := s.userDomainService.ListUsers(ctx, filter)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// BulkUpdateUsers 批量更新用户(管理员功能)
func (s *UserService) BulkUpdateUsers(ctx context.Context, updates []userDomain.BulkUpdateRequest) error {
	return s.userDomainService.BulkUpdateUsers(ctx, updates)
}

// BulkDeleteUsers 批量删除用户(管理员功能)
func (s *UserService) BulkDeleteUsers(ctx context.Context, userIDs []string) error {
	return s.userDomainService.BulkDeleteUsers(ctx, userIDs)
}

// ExportUsers 导出用户数据(管理员功能)
func (s *UserService) ExportUsers(ctx context.Context, filter userDomain.ListFilter) ([]*UserResponse, error) {
	users, err := s.userDomainService.ExportUsers(ctx, filter)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	responses := make([]*UserResponse, len(users))
	for i, user := range users {
		responses[i] = s.toUserResponse(user)
	}

	return responses, nil
}

// ImportUsers 导入用户数据(管理员功能)
func (s *UserService) ImportUsers(ctx context.Context, userReqs []userDomain.CreateUserRequest) ([]*UserResponse, error) {
	users, err := s.userDomainService.ImportUsers(ctx, userReqs)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	responses := make([]*UserResponse, len(users))
	for i, user := range users {
		responses[i] = s.toUserResponse(user)
	}

	return responses, nil
}

// GetUserStats 获取用户统计信息(管理员功能)
func (s *UserService) GetUserStats(ctx context.Context) (*userDomain.UserStats, error) {
	return s.userDomainService.GetUserStats(ctx)
}

// GetUserActivity 获取用户活动信息
func (s *UserService) GetUserActivity(ctx context.Context, userID string, days int) (*userDomain.UserActivity, error) {
	return s.userDomainService.GetUserActivity(ctx, userID, days)
}

// GetUserSessions 获取用户的活跃会话
func (s *UserService) GetUserSessions(ctx context.Context, userID string) ([]*UserSession, error) {
	return s.sessionService.GetUserActiveSessions(ctx, userID)
}

// InvalidateUserSession 使指定会话失效
func (s *UserService) InvalidateUserSession(ctx context.Context, userID, sessionID string) error {
	// 验证会话属于该用户
	session, err := s.sessionService.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	if session.UserID != userID {
		return errors.ErrForbidden
	}

	return s.sessionService.InvalidateSession(ctx, sessionID)
}

// InvalidateAllUserSessions 使用户的所有会话失效
func (s *UserService) InvalidateAllUserSessions(ctx context.Context, userID string) error {
	// 使所有会话失效
	if err := s.sessionService.InvalidateAllUserSessions(ctx, userID); err != nil {
		return err
	}

	// 使所有令牌失效
	return s.tokenService.InvalidateUserTokens(ctx, userID)
}

// GenerateAPIKey 生成API密钥
func (s *UserService) GenerateAPIKey(ctx context.Context, userID, name string, expiresAt *time.Time) (string, error) {
	// 验证用户存在
	_, err := s.userDomainService.GetUser(ctx, userID)
	if err != nil {
		return "", err
	}

	return s.tokenService.GenerateAPIKey(ctx, userID, name, expiresAt)
}

// ValidateAPIKey 验证API密钥
func (s *UserService) ValidateAPIKey(ctx context.Context, apiKey string) (*UserResponse, error) {
	userID, err := s.tokenService.ValidateAPIKey(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	domainUser, err := s.userDomainService.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !domainUser.IsActive() {
		return nil, errors.ErrUserDeactivated
	}

	return s.toUserResponse(domainUser), nil
}

// RevokeAPIKey 撤销API密钥
func (s *UserService) RevokeAPIKey(ctx context.Context, apiKey string) error {
	return s.tokenService.RevokeAPIKey(ctx, apiKey)
}

// UpdateUserPreferences 更新用户偏好设置
func (s *UserService) UpdateUserPreferences(ctx context.Context, userID string, prefs userDomain.UserPreferences) error {
	return s.userDomainService.UpdateUserPreferences(ctx, userID, prefs)
}

// GetUserPreferences 获取用户偏好设置
func (s *UserService) GetUserPreferences(ctx context.Context, userID string) (*userDomain.UserPreferences, error) {
	return s.userDomainService.GetUserPreferences(ctx, userID)
}

// UpdateUser 更新用户信息(管理员功能)
func (s *UserService) UpdateUser(ctx context.Context, userID string, req userDomain.UpdateUserRequest) (*UserResponse, error) {
	domainUser, err := s.userDomainService.UpdateUser(ctx, userID, req)
	if err != nil {
		return nil, err
	}

	// 清除缓存
	if err := s.clearUserCache(ctx, userID); err != nil {
		logger.Warn("Failed to clear user cache after update",
			logger.String("user_id", userID),
			logger.ErrorField(err),
		)
	}

	return s.toUserResponse(domainUser), nil
}

// DeleteUser 删除用户(管理员功能)
func (s *UserService) DeleteUser(ctx context.Context, userID string) error {
	err := s.userDomainService.DeleteUser(ctx, userID)
	if err != nil {
		return err
	}

	// 清除缓存
	if err := s.clearUserCache(ctx, userID); err != nil {
		logger.Warn("Failed to clear user cache after deletion",
			logger.String("user_id", userID),
			logger.ErrorField(err),
		)
	}

	return nil
}

// PromoteToAdmin 提升用户为管理员(管理员功能)
func (s *UserService) PromoteToAdmin(ctx context.Context, userID string) error {
	err := s.userDomainService.PromoteToAdmin(ctx, userID)
	if err != nil {
		return err
	}

	// 清除缓存
	if err := s.clearUserCache(ctx, userID); err != nil {
		logger.Warn("Failed to clear user cache after promotion",
			logger.String("user_id", userID),
			logger.ErrorField(err),
		)
	}

	return nil
}

// DemoteFromAdmin 撤销管理员权限(管理员功能)
func (s *UserService) DemoteFromAdmin(ctx context.Context, userID string) error {
	err := s.userDomainService.DemoteFromAdmin(ctx, userID)
	if err != nil {
		return err
	}

	// 清除缓存
	if err := s.clearUserCache(ctx, userID); err != nil {
		logger.Warn("Failed to clear user cache after demotion",
			logger.String("user_id", userID),
			logger.ErrorField(err),
		)
	}

	return nil
}

// ActivateUser 激活用户(管理员功能)
func (s *UserService) ActivateUser(ctx context.Context, userID string) error {
	err := s.userDomainService.ActivateUser(ctx, userID)
	if err != nil {
		return err
	}

	// 清除缓存
	if err := s.clearUserCache(ctx, userID); err != nil {
		logger.Warn("Failed to clear user cache after activation",
			logger.String("user_id", userID),
			logger.ErrorField(err),
		)
	}

	return nil
}

// DeactivateUser 停用用户(管理员功能)
func (s *UserService) DeactivateUser(ctx context.Context, userID string) error {
	err := s.userDomainService.DeactivateUser(ctx, userID)
	if err != nil {
		return err
	}

	// 清除缓存
	if err := s.clearUserCache(ctx, userID); err != nil {
		logger.Warn("Failed to clear user cache after deactivation",
			logger.String("user_id", userID),
			logger.ErrorField(err),
		)
	}

	return nil
}

// 私有方法

// cacheUserInfo 缓存用户信息
func (s *UserService) cacheUserInfo(ctx context.Context, user *userDomain.User) error {
	key := cache.BuildCacheKey(cache.UserCachePrefix, user.ID)
	return s.cacheService.Set(ctx, key, user, 24*time.Hour)
}

// clearUserCache 清除用户缓存
func (s *UserService) clearUserCache(ctx context.Context, userID string) error {
	key := cache.BuildCacheKey(cache.UserCachePrefix, userID)
	return s.cacheService.Delete(ctx, key)
}

// toUserResponse 转换为用户响应
func (s *UserService) toUserResponse(user *userDomain.User) *UserResponse {
	return &UserResponse{
		ID:               user.ID,
		Name:             user.Name,
		Email:            user.Email,
		Phone:            user.Phone,
		Avatar:           user.Avatar,
		IsAdmin:          user.IsAdmin,
		IsSystem:         user.IsSystem,
		IsTrialUsed:      user.IsTrialUsed,
		LastSignTime:     user.LastSignTime,
		CreatedTime:      user.CreatedTime,
		LastModifiedTime: user.LastModifiedTime,
	}
}
