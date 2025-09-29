package http

import (
	"github.com/gin-gonic/gin"

	"teable-go-backend/internal/domain/space"
	"teable-go-backend/internal/interfaces/middleware"
	"teable-go-backend/pkg/errors"
	"teable-go-backend/pkg/response"
)

// SpaceHandler 空间相关HTTP处理器
type SpaceHandler struct {
	service space.Service
}

func NewSpaceHandler(service space.Service) *SpaceHandler { return &SpaceHandler{service: service} }

// CreateSpace 创建空间
func (h *SpaceHandler) CreateSpace(c *gin.Context) {
	var req struct {
		Name        string  `json:"name" binding:"required"`
		Description *string `json:"description"`
		Icon        *string `json:"icon"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrBadRequest.WithDetails(err.Error()))
		return
	}

	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	sp, err := h.service.CreateSpace(c.Request.Context(), space.CreateSpaceRequest{
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
		CreatedBy:   userID,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithMessage(c, sp, "")
}

// GetSpace 获取空间
func (h *SpaceHandler) GetSpace(c *gin.Context) {
	id := c.Param("id")
	sp, err := h.service.GetSpace(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	if sp == nil {
		response.Error(c, errors.ErrSpaceNotFound)
		return
	}
	response.SuccessWithMessage(c, sp, "")
}

// UpdateSpace 更新空间
func (h *SpaceHandler) UpdateSpace(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Icon        *string `json:"icon"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrBadRequest.WithDetails(err.Error()))
		return
	}
	sp, err := h.service.UpdateSpace(c.Request.Context(), id, space.UpdateSpaceRequest{
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessWithMessage(c, sp, "")
}

// DeleteSpace 删除空间
func (h *SpaceHandler) DeleteSpace(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteSpace(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessWithMessage(c, map[string]bool{"success": true}, "")
}

// ListSpaces 列出空间
func (h *SpaceHandler) ListSpaces(c *gin.Context) {
	var query struct {
		Offset    int     `form:"offset"`
		Limit     int     `form:"limit"`
		OrderBy   string  `form:"order_by"`
		Order     string  `form:"order"`
		Name      *string `form:"name"`
		Search    string  `form:"search"`
		CreatedBy *string `form:"created_by"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, errors.ErrBadRequest.WithDetails(err.Error()))
		return
	}
	filter := space.ListFilter{Offset: query.Offset, Limit: query.Limit, OrderBy: query.OrderBy, Order: query.Order, Name: query.Name, Search: query.Search, CreatedBy: query.CreatedBy}
	items, total, err := h.service.ListSpaces(c.Request.Context(), filter)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.PaginatedSuccess(c, items, response.Pagination{Page: 0, Limit: filter.Limit, Total: int(total), TotalPages: 0}, "")
}

// AddCollaborator 添加协作者
// @Summary 添加空间协作者
// @Description 为指定空间添加协作者
// @Tags 空间管理
// @Accept json
// @Produce json
// @Param id path string true "空间ID"
// @Param request body AddCollaboratorRequest true "协作者信息"
// @Success 201 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/spaces/{id}/collaborators [post]
func (h *SpaceHandler) AddCollaborator(c *gin.Context) {
	spaceID := c.Param("id")
	if spaceID == "" {
		response.Error(c, errors.ErrBadRequest.WithDetails("空间ID不能为空"))
		return
	}

	var req struct {
		UserID string `json:"user_id" binding:"required"`
		Role   string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrBadRequest.WithDetails(err.Error()))
		return
	}

	if err := h.service.AddCollaborator(c.Request.Context(), spaceID, req.UserID, req.Role); err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessWithMessage(c, map[string]string{"message": "协作者添加成功"}, "")
}

// RemoveCollaborator 移除协作者
// @Summary 移除空间协作者
// @Description 从指定空间移除协作者
// @Tags 空间管理
// @Accept json
// @Produce json
// @Param id path string true "空间ID"
// @Param collab_id path string true "协作者ID"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/spaces/{id}/collaborators/{collab_id} [delete]
func (h *SpaceHandler) RemoveCollaborator(c *gin.Context) {
	collabID := c.Param("collab_id")
	if collabID == "" {
		response.Error(c, errors.ErrBadRequest.WithDetails("协作者ID不能为空"))
		return
	}

	if err := h.service.RemoveCollaborator(c.Request.Context(), collabID); err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessWithMessage(c, map[string]string{"message": "协作者移除成功"}, "")
}

// ListCollaborators 列出协作者
// @Summary 列出空间协作者
// @Description 获取指定空间的所有协作者
// @Tags 空间管理
// @Accept json
// @Produce json
// @Param id path string true "空间ID"
// @Success 200 {array} space.SpaceCollaborator
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/spaces/{id}/collaborators [get]
func (h *SpaceHandler) ListCollaborators(c *gin.Context) {
	spaceID := c.Param("id")
	if spaceID == "" {
		response.Error(c, errors.ErrBadRequest.WithDetails("空间ID不能为空"))
		return
	}

	collaborators, err := h.service.ListCollaborators(c.Request.Context(), spaceID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessWithMessage(c, collaborators, "")
}

// UpdateCollaboratorRole 更新协作者角色
// @Summary 更新协作者角色
// @Description 更新指定协作者的角色
// @Tags 空间管理
// @Accept json
// @Produce json
// @Param id path string true "空间ID"
// @Param collab_id path string true "协作者ID"
// @Param request body UpdateCollaboratorRoleRequest true "角色信息"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/spaces/{id}/collaborators/{collab_id}/role [put]
func (h *SpaceHandler) UpdateCollaboratorRole(c *gin.Context) {
	collabID := c.Param("collab_id")
	if collabID == "" {
		response.Error(c, errors.ErrBadRequest.WithDetails("协作者ID不能为空"))
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrBadRequest.WithDetails(err.Error()))
		return
	}

	if err := h.service.UpdateCollaboratorRole(c.Request.Context(), collabID, req.Role); err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessWithMessage(c, map[string]string{"message": "协作者角色更新成功"}, "")
}

// BulkUpdateSpaces 批量更新空间
// @Summary 批量更新空间
// @Description 批量更新多个空间的信息
// @Tags 空间管理
// @Accept json
// @Produce json
// @Param request body []space.BulkUpdateRequest true "批量更新信息"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/spaces/bulk-update [post]
func (h *SpaceHandler) BulkUpdateSpaces(c *gin.Context) {
	var updates []space.BulkUpdateRequest
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.Error(c, errors.ErrBadRequest.WithDetails(err.Error()))
		return
	}

	if err := h.service.BulkUpdateSpaces(c.Request.Context(), updates); err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessWithMessage(c, map[string]string{"message": "批量更新成功"}, "")
}

// BulkDeleteSpaces 批量删除空间
// @Summary 批量删除空间
// @Description 批量删除多个空间
// @Tags 空间管理
// @Accept json
// @Produce json
// @Param request body []string true "空间ID列表"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/spaces/bulk-delete [post]
func (h *SpaceHandler) BulkDeleteSpaces(c *gin.Context) {
	var spaceIDs []string
	if err := c.ShouldBindJSON(&spaceIDs); err != nil {
		response.Error(c, errors.ErrBadRequest.WithDetails(err.Error()))
		return
	}

	if err := h.service.BulkDeleteSpaces(c.Request.Context(), spaceIDs); err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessWithMessage(c, map[string]string{"message": "批量删除成功"}, "")
}

// CheckUserPermission 检查用户权限
// @Summary 检查用户权限
// @Description 检查指定用户对空间的权限
// @Tags 空间管理
// @Accept json
// @Produce json
// @Param id path string true "空间ID"
// @Param user_id query string true "用户ID"
// @Param permission query string true "权限类型" Enums(read,write,admin)
// @Success 200 {object} PermissionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/spaces/{id}/permissions [get]
func (h *SpaceHandler) CheckUserPermission(c *gin.Context) {
	spaceID := c.Param("id")
	if spaceID == "" {
		response.Error(c, errors.ErrBadRequest.WithDetails("空间ID不能为空"))
		return
	}

	userID := c.Query("user_id")
	if userID == "" {
		response.Error(c, errors.ErrBadRequest.WithDetails("用户ID不能为空"))
		return
	}

	permission := c.Query("permission")
	if permission == "" {
		response.Error(c, errors.ErrBadRequest.WithDetails("权限类型不能为空"))
		return
	}

	hasPermission, err := h.service.CheckUserPermission(c.Request.Context(), spaceID, userID, permission)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessWithMessage(c, gin.H{
		"has_permission": hasPermission,
		"user_id":        userID,
		"space_id":       spaceID,
		"permission":     permission,
	}, "")
}

// GetUserSpaces 获取用户空间
// @Summary 获取用户空间
// @Description 获取指定用户的空间列表
// @Tags 空间管理
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Param offset query int false "偏移量" default(0)
// @Param limit query int false "限制数量" default(20)
// @Param search query string false "搜索关键词"
// @Success 200 {object} UserSpacesResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/spaces/user/{user_id} [get]
func (h *SpaceHandler) GetUserSpaces(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		response.Error(c, errors.ErrBadRequest.WithDetails("用户ID不能为空"))
		return
	}

	var query struct {
		Offset int    `form:"offset"`
		Limit  int    `form:"limit"`
		Search string `form:"search"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, errors.ErrBadRequest.WithDetails(err.Error()))
		return
	}

	filter := space.ListFilter{
		Offset:  query.Offset,
		Limit:   query.Limit,
		Search:  query.Search,
		OrderBy: "created_time",
		Order:   "DESC",
	}

	spaces, total, err := h.service.GetUserSpaces(c.Request.Context(), userID, filter)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.PaginatedSuccess(c, spaces, response.Pagination{Page: 0, Limit: filter.Limit, Total: int(total), TotalPages: 0}, "")
}

// GetSpaceStats 获取空间统计信息
// @Summary 获取空间统计信息
// @Description 获取指定空间的统计信息
// @Tags 空间管理
// @Accept json
// @Produce json
// @Param id path string true "空间ID"
// @Success 200 {object} space.SpaceStats
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/spaces/{id}/stats [get]
func (h *SpaceHandler) GetSpaceStats(c *gin.Context) {
	spaceID := c.Param("id")
	if spaceID == "" {
		response.Error(c, errors.ErrBadRequest.WithDetails("空间ID不能为空"))
		return
	}

	stats, err := h.service.GetSpaceStats(c.Request.Context(), spaceID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessWithMessage(c, stats, "")
}

// GetUserSpaceStats 获取用户空间统计信息
// @Summary 获取用户空间统计信息
// @Description 获取指定用户的空间统计信息
// @Tags 空间管理
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Success 200 {object} space.UserSpaceStats
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/spaces/user/{user_id}/stats [get]
func (h *SpaceHandler) GetUserSpaceStats(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		response.Error(c, errors.ErrBadRequest.WithDetails("用户ID不能为空"))
		return
	}

	stats, err := h.service.GetUserSpaceStats(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessWithMessage(c, stats, "")
}
