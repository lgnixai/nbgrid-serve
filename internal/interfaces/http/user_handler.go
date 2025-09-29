package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"teable-go-backend/internal/application"
	userDomain "teable-go-backend/internal/domain/user"
	"teable-go-backend/internal/interfaces/middleware"
	"teable-go-backend/pkg/errors"
	"teable-go-backend/pkg/logger"
	"teable-go-backend/pkg/response"
)

// UserHandler 用户HTTP处理器
type UserHandler struct {
	userService *application.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService *application.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// Register 用户注册
// @Summary 用户注册
// @Description 创建新用户账户
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body application.RegisterRequest true "注册信息"
// @Success 201 {object} application.AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /api/auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req application.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid register request", logger.ErrorField(err))
		response.Error(c, errors.ErrBadRequest.WithDetails(err.Error()))
		return
	}

	loginCtx := &application.LoginContext{
		IPAddress: c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		DeviceID:  c.GetHeader("X-Device-ID"),
	}

	responseData, err := h.userService.Register(c.Request.Context(), req, loginCtx)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 统一响应包装
	response.Success(c, responseData, "")
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录认证
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body application.LoginRequest true "登录信息"
// @Success 200 {object} application.AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req application.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid login request", logger.ErrorField(err))
		response.Error(c, errors.ErrBadRequest.WithDetails(err.Error()))
		return
	}

	loginCtx := &application.LoginContext{
		IPAddress: c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		DeviceID:  c.GetHeader("X-Device-ID"),
	}

	responseData, err := h.userService.Login(c.Request.Context(), req, loginCtx)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 统一响应包装
	response.Success(c, responseData, "")
}

// Logout 用户登出
// @Summary 用户登出
// @Description 用户登出，令牌失效
// @Tags 认证
// @Accept json
// @Produce json
// @Success 200 {object} MessageResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/auth/logout [post]
func (h *UserHandler) Logout(c *gin.Context) {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	token := c.GetString("token")
	if token == "" {
		response.Error(c, errors.ErrBadRequest.WithDetails("令牌不能为空"))
		return
	}

	sessionID := c.GetString("session_id") // or extract from token
	if sessionID == "" {
		sessionID = "default" // fallback
	}

	if err := h.userService.Logout(c.Request.Context(), userID, token, sessionID); err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, nil, "登出成功")
}

// RefreshToken 刷新访问令牌
// @Summary 刷新访问令牌
// @Description 使用refresh token刷新访问令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body application.RefreshTokenRequest true "刷新令牌信息"
// @Success 200 {object} application.AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/auth/refresh [post]
func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req application.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid refresh token request", logger.ErrorField(err))
		response.Error(c, errors.ErrBadRequest.WithDetails(err.Error()))
		return
	}

	responseData, err := h.userService.RefreshToken(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 统一响应包装
	response.Success(c, responseData, "")
}

// GetProfile 获取当前用户资料
// @Summary 获取用户资料
// @Description 获取当前登录用户的资料信息
// @Tags 用户
// @Accept json
// @Produce json
// @Success 200 {object} application.UserResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/users/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response, err := h.userService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateProfile 更新用户资料
// @Summary 更新用户资料
// @Description 更新当前登录用户的资料信息
// @Tags 用户
// @Accept json
// @Produce json
// @Param request body application.UpdateProfileRequest true "更新资料信息"
// @Success 200 {object} application.UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/users/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	var req application.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid update profile request", logger.ErrorField(err))
		response.Error(c, errors.ErrBadRequest.WithDetails(err.Error()))
		return
	}

	response, err := h.userService.UpdateProfile(c.Request.Context(), userID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改当前登录用户的密码
// @Tags 用户
// @Accept json
// @Produce json
// @Param request body application.ChangePasswordRequest true "修改密码信息"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/users/change-password [post]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	var req application.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid change password request", logger.ErrorField(err))
		response.Error(c, errors.ErrBadRequest.WithDetails(err.Error()))
		return
	}

	if err := h.userService.ChangePassword(c.Request.Context(), userID, req); err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, nil, "密码修改成功")
}

// GetUser 获取用户信息(管理员功能)
// @Summary 获取用户信息
// @Description 根据用户ID获取用户详细信息(需要管理员权限)
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Success 200 {object} application.UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/admin/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.Error(c, errors.ErrBadRequest.WithDetails("用户ID不能为空"))
		return
	}

	response, err := h.userService.GetUser(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// ListUsers 列出用户(管理员功能)
// @Summary 列出用户
// @Description 分页获取用户列表(需要管理员权限)
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param search query string false "搜索关键词"
// @Param is_active query bool false "是否激活"
// @Param is_admin query bool false "是否管理员"
// @Success 200 {object} user.PaginatedResult
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/admin/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	search := c.Query("search")

	// 限制分页参数
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 构建过滤器
	filter := userDomain.NewListFilter()
	filter.Offset = (page - 1) * pageSize
	filter.Limit = pageSize
	filter.Search = search

	// 处理布尔查询参数
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
			filter.IsActive = &isActive
		}
	}

	if isAdminStr := c.Query("is_admin"); isAdminStr != "" {
		if isAdmin, err := strconv.ParseBool(isAdminStr); err == nil {
			filter.IsAdmin = &isAdmin
		}
	}

	result, err := h.userService.ListUsers(c.Request.Context(), filter)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// BulkUpdateUsers 批量更新用户(管理员功能)
// @Summary 批量更新用户
// @Description 批量更新多个用户的信息(需要管理员权限)
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body []userDomain.BulkUpdateRequest true "批量更新信息"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/admin/users/bulk-update [post]
func (h *UserHandler) BulkUpdateUsers(c *gin.Context) {
	var updates []userDomain.BulkUpdateRequest
	if err := c.ShouldBindJSON(&updates); err != nil {
		logger.Warn("Invalid bulk update request", logger.ErrorField(err))
		response.Error(c, errors.ErrBadRequest.WithDetails(err.Error()))
		return
	}

	if err := h.userService.BulkUpdateUsers(c.Request.Context(), updates); err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, nil, "批量更新成功")
}

// BulkDeleteUsers 批量删除用户(管理员功能)
// @Summary 批量删除用户
// @Description 批量删除多个用户(需要管理员权限)
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body []string true "用户ID列表"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/admin/users/bulk-delete [post]
func (h *UserHandler) BulkDeleteUsers(c *gin.Context) {
	var userIDs []string
	if err := c.ShouldBindJSON(&userIDs); err != nil {
		logger.Warn("Invalid bulk delete request", logger.ErrorField(err))
		response.Error(c, errors.ErrBadRequest.WithDetails(err.Error()))
		return
	}

	if err := h.userService.BulkDeleteUsers(c.Request.Context(), userIDs); err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, nil, "批量删除成功")
}

// ExportUsers 导出用户数据(管理员功能)
// @Summary 导出用户数据
// @Description 导出用户数据为JSON格式(需要管理员权限)
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param search query string false "搜索关键词"
// @Param is_active query bool false "是否激活"
// @Param is_admin query bool false "是否管理员"
// @Success 200 {array} application.UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/admin/users/export [get]
func (h *UserHandler) ExportUsers(c *gin.Context) {
	// 构建过滤器
	filter := userDomain.NewListFilter()
	filter.Search = c.Query("search")

	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
			filter.IsActive = &isActive
		}
	}

	if isAdminStr := c.Query("is_admin"); isAdminStr != "" {
		if isAdmin, err := strconv.ParseBool(isAdminStr); err == nil {
			filter.IsAdmin = &isAdmin
		}
	}

	users, err := h.userService.ExportUsers(c.Request.Context(), filter)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, users)
}

// ImportUsers 导入用户数据(管理员功能)
// @Summary 导入用户数据
// @Description 从JSON格式导入用户数据(需要管理员权限)
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body []userDomain.CreateUserRequest true "用户数据列表"
// @Success 201 {array} application.UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/admin/users/import [post]
func (h *UserHandler) ImportUsers(c *gin.Context) {
	var userReqs []userDomain.CreateUserRequest
	if err := c.ShouldBindJSON(&userReqs); err != nil {
		logger.Warn("Invalid import request", logger.ErrorField(err))
		response.Error(c, errors.ErrBadRequest.WithDetails(err.Error()))
		return
	}

	users, err := h.userService.ImportUsers(c.Request.Context(), userReqs)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, users)
}

// GetUserStats 获取用户统计信息(管理员功能)
// @Summary 获取用户统计信息
// @Description 获取用户统计和分析数据(需要管理员权限)
// @Tags 用户管理
// @Accept json
// @Produce json
// @Success 200 {object} userDomain.UserStats
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/admin/users/stats [get]
func (h *UserHandler) GetUserStats(c *gin.Context) {
	stats, err := h.userService.GetUserStats(c.Request.Context())
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetUserActivity 获取用户活动信息
// @Summary 获取用户活动信息
// @Description 获取指定用户的活动信息
// @Tags 用户
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Param days query int false "天数" default(30)
// @Success 200 {object} userDomain.UserActivity
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/users/{id}/activity [get]
func (h *UserHandler) GetUserActivity(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.Error(c, errors.ErrBadRequest.WithDetails("用户ID不能为空"))
		return
	}

	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days <= 0 {
		days = 30
	}

	activity, err := h.userService.GetUserActivity(c.Request.Context(), userID, days)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, activity)
}

// UpdateUserPreferences 更新用户偏好设置
// @Summary 更新用户偏好设置
// @Description 更新当前登录用户的偏好设置
// @Tags 用户
// @Accept json
// @Produce json
// @Param request body userDomain.UserPreferences true "偏好设置"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/users/preferences [put]
func (h *UserHandler) UpdateUserPreferences(c *gin.Context) {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	var prefs userDomain.UserPreferences
	if err := c.ShouldBindJSON(&prefs); err != nil {
		logger.Warn("Invalid preferences request", logger.ErrorField(err))
		response.Error(c, errors.ErrBadRequest.WithDetails(err.Error()))
		return
	}

	if err := h.userService.UpdateUserPreferences(c.Request.Context(), userID, prefs); err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, nil, "偏好设置更新成功")
}

// GetUserPreferences 获取用户偏好设置
// @Summary 获取用户偏好设置
// @Description 获取当前登录用户的偏好设置
// @Tags 用户
// @Accept json
// @Produce json
// @Success 200 {object} userDomain.UserPreferences
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/users/preferences [get]
func (h *UserHandler) GetUserPreferences(c *gin.Context) {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	prefs, err := h.userService.GetUserPreferences(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, prefs)
}

// UpdateUser 更新用户信息(管理员功能)
// @Summary 更新用户信息
// @Description 管理员更新指定用户的信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Param request body userDomain.UpdateUserRequest true "用户信息"
// @Success 200 {object} application.UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/admin/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.Error(c, errors.ErrBadRequest.WithDetails("用户ID不能为空"))
		return
	}

	var req userDomain.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid update user request", logger.ErrorField(err))
		response.Error(c, errors.ErrBadRequest.WithDetails(err.Error()))
		return
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), userID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteUser 删除用户(管理员功能)
// @Summary 删除用户
// @Description 管理员删除指定用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/admin/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.Error(c, errors.ErrBadRequest.WithDetails("用户ID不能为空"))
		return
	}

	if err := h.userService.DeleteUser(c.Request.Context(), userID); err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, nil, "用户删除成功")
}

// PromoteToAdmin 提升用户为管理员(管理员功能)
// @Summary 提升用户为管理员
// @Description 管理员提升指定用户为管理员
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/admin/users/{id}/promote [post]
func (h *UserHandler) PromoteToAdmin(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.Error(c, errors.ErrBadRequest.WithDetails("用户ID不能为空"))
		return
	}

	if err := h.userService.PromoteToAdmin(c.Request.Context(), userID); err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, nil, "用户已提升为管理员")
}

// DemoteFromAdmin 撤销管理员权限(管理员功能)
// @Summary 撤销管理员权限
// @Description 管理员撤销指定用户的管理员权限
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/admin/users/{id}/demote [post]
func (h *UserHandler) DemoteFromAdmin(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.Error(c, errors.ErrBadRequest.WithDetails("用户ID不能为空"))
		return
	}

	if err := h.userService.DemoteFromAdmin(c.Request.Context(), userID); err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, nil, "管理员权限已撤销")
}

// ActivateUser 激活用户(管理员功能)
// @Summary 激活用户
// @Description 管理员激活指定用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/admin/users/{id}/activate [post]
func (h *UserHandler) ActivateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.Error(c, errors.ErrBadRequest.WithDetails("用户ID不能为空"))
		return
	}

	if err := h.userService.ActivateUser(c.Request.Context(), userID); err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, nil, "用户已激活")
}

// DeactivateUser 停用用户(管理员功能)
// @Summary 停用用户
// @Description 管理员停用指定用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/admin/users/{id}/deactivate [post]
func (h *UserHandler) DeactivateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.Error(c, errors.ErrBadRequest.WithDetails("用户ID不能为空"))
		return
	}

	if err := h.userService.DeactivateUser(c.Request.Context(), userID); err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, nil, "用户已停用")
}

// handleError 统一错误处理
func (h *UserHandler) handleError(c *gin.Context, err error) {
	response.Error(c, err)
}
