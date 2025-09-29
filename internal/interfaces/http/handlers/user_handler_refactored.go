package handlers

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"teable-go-backend/internal/application"
	"teable-go-backend/internal/interfaces/http/validators"
)

// UserHandler 用户处理器（重构版）
type UserHandler struct {
	*BaseHandler
	userService *application.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService *application.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		BaseHandler: NewBaseHandler(logger),
		userService: userService,
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,password"`
}

// Register 用户注册
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		return
	}
	
	// 执行注册
	user, err := h.userService.Register(c.Request.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	
	// 记录活动
	h.LogActivity(c, "user.register", "user", map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
	})
	
	h.HandleCreated(c, user, "User registered successfully")
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Login 用户登录
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		return
	}
	
	// 执行登录
	tokens, err := h.userService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	
	// 记录活动
	h.LogActivity(c, "user.login", "user", map[string]interface{}{
		"email": req.Email,
	})
	
	h.HandleSuccess(c, tokens, "Login successful")
}

// GetProfile 获取用户资料
func (h *UserHandler) GetProfile(c *gin.Context) {
	user, err := h.RequireAuth(c)
	if err != nil {
		return
	}
	
	// 尝试从缓存获取
	cacheKey := h.CacheKey("user:profile", user.ID)
	var cachedUser interface{}
	if err := h.GetCache(c, cacheKey, &cachedUser); err == nil {
		h.HandleSuccess(c, cachedUser)
		return
	}
	
	// 从服务获取
	profile, err := h.userService.GetUserByID(c.Request.Context(), user.ID)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	
	// 设置缓存
	h.SetCache(c, cacheKey, profile, 5*time.Minute)
	
	h.HandleSuccess(c, profile)
}

// UpdateProfileRequest 更新资料请求
type UpdateProfileRequest struct {
	Name   *string `json:"name" validate:"omitempty,min=2,max=100"`
	Phone  *string `json:"phone" validate:"omitempty,phone"`
	Avatar *string `json:"avatar" validate:"omitempty,url"`
}

// UpdateProfile 更新用户资料
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	user, err := h.RequireAuth(c)
	if err != nil {
		return
	}
	
	var req UpdateProfileRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		return
	}
	
	// 构建更新数据
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Phone != nil {
		updates["phone"] = req.Phone
	}
	if req.Avatar != nil {
		updates["avatar"] = req.Avatar
	}
	
	// 执行更新
	updatedUser, err := h.userService.UpdateUser(c.Request.Context(), user.ID, updates)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	
	// 清除缓存
	h.InvalidateCache(c, h.CacheKey("user:profile", user.ID))
	
	// 记录活动
	h.LogActivity(c, "user.update_profile", "user", updates)
	
	h.HandleSuccess(c, updatedUser, "Profile updated successfully")
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,password"`
}

// ChangePassword 修改密码
func (h *UserHandler) ChangePassword(c *gin.Context) {
	user, err := h.RequireAuth(c)
	if err != nil {
		return
	}
	
	var req ChangePasswordRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		return
	}
	
	// 执行密码修改
	err = h.userService.ChangePassword(
		c.Request.Context(),
		user.ID,
		req.OldPassword,
		req.NewPassword,
	)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	
	// 记录活动
	h.LogActivity(c, "user.change_password", "user", nil)
	
	h.HandleSuccess(c, nil, "Password changed successfully")
}

// ListUsersRequest 用户列表请求
type ListUsersRequest struct {
	validators.PaginationRequest
	validators.SearchRequest
	validators.SortRequest
	Status string `form:"status" validate:"omitempty,oneof=active inactive deleted"`
}

// ListUsers 获取用户列表（管理员）
func (h *UserHandler) ListUsers(c *gin.Context) {
	// 检查管理员权限
	if err := h.CheckPermission(c, "user", "list"); err != nil {
		h.HandleForbidden(c, "Admin access required")
		return
	}
	
	var req ListUsersRequest
	if err := h.ValidateQueryParams(c, &req); err != nil {
		return
	}
	
	// 设置默认值
	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.SortBy == "" {
		req.SortBy = "created_time"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}
	
	// 获取用户列表
	users, total, err := h.userService.ListUsers(
		c.Request.Context(),
		req.Offset,
		req.Limit,
		req.Query,
		req.Status,
		req.SortBy,
		req.SortOrder,
	)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	
	h.HandlePaginatedSuccess(c, users, total, req.Offset, req.Limit)
}

// GetUser 获取指定用户（管理员）
func (h *UserHandler) GetUser(c *gin.Context) {
	// 检查管理员权限
	if err := h.CheckPermission(c, "user", "read"); err != nil {
		h.HandleForbidden(c, "Admin access required")
		return
	}
	
	userID := h.GetPathParam(c, "id")
	if err := validators.ValidateID(userID); err != nil {
		h.HandleBadRequest(c, err.Error())
		return
	}
	
	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	
	h.HandleSuccess(c, user)
}

// DeleteUser 删除用户（管理员）
func (h *UserHandler) DeleteUser(c *gin.Context) {
	// 检查管理员权限
	if err := h.CheckPermission(c, "user", "delete"); err != nil {
		h.HandleForbidden(c, "Admin access required")
		return
	}
	
	userID := h.GetPathParam(c, "id")
	if err := validators.ValidateID(userID); err != nil {
		h.HandleBadRequest(c, err.Error())
		return
	}
	
	// 防止删除自己
	currentUser, _ := h.GetCurrentUser(c)
	if currentUser.ID == userID {
		h.HandleBadRequest(c, "Cannot delete yourself")
		return
	}
	
	err := h.userService.DeleteUser(c.Request.Context(), userID)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	
	// 记录活动
	h.LogActivity(c, "user.delete", "user", map[string]interface{}{
		"deleted_user_id": userID,
	})
	
	h.HandleSuccess(c, nil, "User deleted successfully")
}

// BulkUpdateUsers 批量更新用户（管理员）
func (h *UserHandler) BulkUpdateUsers(c *gin.Context) {
	// 检查管理员权限
	if err := h.CheckPermission(c, "user", "update"); err != nil {
		h.HandleForbidden(c, "Admin access required")
		return
	}
	
	var req struct {
		validators.BatchIDsRequest
		Updates map[string]interface{} `json:"updates" validate:"required"`
	}
	
	if err := h.ValidateRequest(c, &req); err != nil {
		return
	}
	
	// 执行批量更新
	affected, err := h.userService.BulkUpdateUsers(
		c.Request.Context(),
		req.IDs,
		req.Updates,
	)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	
	// 记录活动
	h.LogActivity(c, "user.bulk_update", "user", map[string]interface{}{
		"user_ids": req.IDs,
		"updates":  req.Updates,
		"affected": affected,
	})
	
	h.HandleSuccess(c, map[string]interface{}{
		"affected": affected,
	}, "Users updated successfully")
}

// Logout 用户登出
func (h *UserHandler) Logout(c *gin.Context) {
	user, err := h.RequireAuth(c)
	if err != nil {
		return
	}
	
	// 获取token
	token := c.GetHeader("Authorization")
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}
	
	// 执行登出
	err = h.userService.Logout(c.Request.Context(), user.ID, token)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	
	// 记录活动
	h.LogActivity(c, "user.logout", "user", nil)
	
	h.HandleSuccess(c, nil, "Logout successful")
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshToken 刷新访问令牌
func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		return
	}
	
	// 执行刷新
	tokens, err := h.userService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	
	h.HandleSuccess(c, tokens, "Token refreshed successfully")
}