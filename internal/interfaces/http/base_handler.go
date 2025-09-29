package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"teable-go-backend/internal/domain/base"
	"teable-go-backend/pkg/errors"
	"teable-go-backend/pkg/response"
)

// BaseHandler 基础表处理器
type BaseHandler struct {
	baseService base.Service
}

// NewBaseHandler 创建基础表处理器
func NewBaseHandler(baseService base.Service) *BaseHandler {
	return &BaseHandler{
		baseService: baseService,
	}
}

// CreateBase 创建基础表
// @Summary 创建基础表
// @Description 在指定空间中创建新的基础表
// @Tags 基础表
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body base.CreateBaseRequest true "创建基础表请求"
// @Success 201 {object} base.Base "创建成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 409 {object} ErrorResponse "基础表已存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/bases [post]
func (h *BaseHandler) CreateBase(c *gin.Context) {
	var req base.CreateBaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrBadRequest.WithDetails(err.Error()))
		return
	}

	// 从JWT中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}
	req.CreatedBy = userID.(string)

	b, err := h.baseService.CreateBase(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, b, "")
}

// GetBase 获取基础表详情
// @Summary 获取基础表详情
// @Description 根据ID获取基础表详情
// @Tags 基础表
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "基础表ID"
// @Success 200 {object} base.Base "获取成功"
// @Failure 404 {object} ErrorResponse "基础表不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/bases/{id} [get]
func (h *BaseHandler) GetBase(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, errors.ErrBadRequest.WithDetails("基础表ID不能为空"))
		return
	}

	b, err := h.baseService.GetBase(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, b, "")
}

// UpdateBase 更新基础表
// @Summary 更新基础表
// @Description 更新基础表信息
// @Tags 基础表
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "基础表ID"
// @Param request body base.UpdateBaseRequest true "更新基础表请求"
// @Success 200 {object} base.Base "更新成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "基础表不存在"
// @Failure 409 {object} ErrorResponse "基础表名称已存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/bases/{id} [put]
func (h *BaseHandler) UpdateBase(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, errors.ErrBadRequest.WithDetails("基础表ID不能为空"))
		return
	}

	var req base.UpdateBaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrBadRequest.WithDetails(err.Error()))
		return
	}

	b, err := h.baseService.UpdateBase(c.Request.Context(), id, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, b, "")
}

// DeleteBase 删除基础表
// @Summary 删除基础表
// @Description 软删除基础表
// @Tags 基础表
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "基础表ID"
// @Success 200 {object} object{success=bool} "删除成功"
// @Failure 404 {object} ErrorResponse "基础表不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/bases/{id} [delete]
func (h *BaseHandler) DeleteBase(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, errors.ErrBadRequest.WithDetails("基础表ID不能为空"))
		return
	}

	err := h.baseService.DeleteBase(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, map[string]bool{"success": true}, "")
}

// ListBases 获取基础表列表
// @Summary 获取基础表列表
// @Description 获取基础表列表，支持分页和过滤
// @Tags 基础表
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param space_id query string false "空间ID"
// @Param name query string false "基础表名称（模糊搜索）"
// @Param search query string false "搜索关键词"
// @Param order_by query string false "排序字段" default(created_time)
// @Param order query string false "排序方向" Enums(asc,desc) default(desc)
// @Param limit query int false "每页数量" default(20)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {object} base.PaginatedResult "获取成功"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/bases [get]
func (h *BaseHandler) ListBases(c *gin.Context) {
	// 解析查询参数
	filter := base.ListFilter{
		OrderBy: "created_time",
		Order:   "desc",
		Limit:   20,
		Offset:  0,
	}

	if spaceID := c.Query("space_id"); spaceID != "" {
		filter.SpaceID = &spaceID
	}
	if name := c.Query("name"); name != "" {
		filter.Name = &name
	}
	if search := c.Query("search"); search != "" {
		filter.Search = search
	}
	if orderBy := c.Query("order_by"); orderBy != "" {
		filter.OrderBy = orderBy
	}
	if order := c.Query("order"); order != "" {
		filter.Order = order
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	// 获取基础表列表
	bases, err := h.baseService.ListBases(c.Request.Context(), filter)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 获取总数
	countFilter := base.CountFilter{
		SpaceID: filter.SpaceID,
		Name:    filter.Name,
		Search:  filter.Search,
	}
	total, err := h.baseService.CountBases(c.Request.Context(), countFilter)
	if err != nil {
		h.handleError(c, err)
		return
	}

	result := base.PaginatedResult{
		Data:   bases,
		Total:  total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}

	response.PaginatedSuccess(c, result.Data, response.Pagination{Page: 0, Limit: result.Limit, Total: int(result.Total), TotalPages: 0}, "")
}

// BulkUpdateBases 批量更新基础表
// @Summary 批量更新基础表
// @Description 批量更新多个基础表的信息
// @Tags 基础表
// @Accept json
// @Produce json
// @Param request body []base.BulkUpdateRequest true "批量更新信息"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/bases/bulk-update [post]
func (h *BaseHandler) BulkUpdateBases(c *gin.Context) {
	var updates []base.BulkUpdateRequest
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.Error(c, errors.ErrBadRequest.WithDetails(err.Error()))
		return
	}

	if err := h.baseService.BulkUpdateBases(c.Request.Context(), updates); err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, map[string]string{"message": "批量更新成功"}, "")
}

// BulkDeleteBases 批量删除基础表
// @Summary 批量删除基础表
// @Description 批量删除多个基础表
// @Tags 基础表
// @Accept json
// @Produce json
// @Param request body []string true "基础表ID列表"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/bases/bulk-delete [post]
func (h *BaseHandler) BulkDeleteBases(c *gin.Context) {
	var baseIDs []string
	if err := c.ShouldBindJSON(&baseIDs); err != nil {
		response.Error(c, errors.ErrBadRequest.WithDetails(err.Error()))
		return
	}

	if err := h.baseService.BulkDeleteBases(c.Request.Context(), baseIDs); err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, map[string]string{"message": "批量删除成功"}, "")
}

// CheckUserPermission 检查用户权限
// @Summary 检查用户权限
// @Description 检查指定用户对基础表的权限
// @Tags 基础表
// @Accept json
// @Produce json
// @Param id path string true "基础表ID"
// @Param user_id query string true "用户ID"
// @Param permission query string true "权限类型" Enums(read,write,admin)
// @Success 200 {object} PermissionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/bases/{id}/permissions [get]
func (h *BaseHandler) CheckUserPermission(c *gin.Context) {
	baseID := c.Param("id")
	if baseID == "" {
		response.Error(c, errors.ErrBadRequest.WithDetails("基础表ID不能为空"))
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

	hasPermission, err := h.baseService.CheckUserPermission(c.Request.Context(), baseID, userID, permission)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, gin.H{
		"has_permission": hasPermission,
		"user_id":        userID,
		"base_id":        baseID,
		"permission":     permission,
	}, "")
}

// GetBaseStats 获取基础表统计信息
// @Summary 获取基础表统计信息
// @Description 获取指定基础表的统计信息
// @Tags 基础表
// @Accept json
// @Produce json
// @Param id path string true "基础表ID"
// @Success 200 {object} base.BaseStats
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/bases/{id}/stats [get]
func (h *BaseHandler) GetBaseStats(c *gin.Context) {
	baseID := c.Param("id")
	if baseID == "" {
		response.Error(c, errors.ErrBadRequest.WithDetails("基础表ID不能为空"))
		return
	}

	stats, err := h.baseService.GetBaseStats(c.Request.Context(), baseID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, stats, "")
}

// GetSpaceBaseStats 获取空间基础表统计信息
// @Summary 获取空间基础表统计信息
// @Description 获取指定空间的基础表统计信息
// @Tags 基础表
// @Accept json
// @Produce json
// @Param space_id path string true "空间ID"
// @Success 200 {object} base.SpaceBaseStats
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/bases/space/{space_id}/stats [get]
func (h *BaseHandler) GetSpaceBaseStats(c *gin.Context) {
	spaceID := c.Param("space_id")
	if spaceID == "" {
		response.Error(c, errors.ErrBadRequest.WithDetails("空间ID不能为空"))
		return
	}

	stats, err := h.baseService.GetSpaceBaseStats(c.Request.Context(), spaceID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, stats, "")
}

// ExportBases 导出基础表
// @Summary 导出基础表
// @Description 导出基础表数据为JSON格式
// @Tags 基础表
// @Accept json
// @Produce json
// @Param space_id query string false "空间ID"
// @Param search query string false "搜索关键词"
// @Success 200 {array} base.Base
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/bases/export [get]
func (h *BaseHandler) ExportBases(c *gin.Context) {
	// 构建过滤器
	filter := base.ListFilter{
		OrderBy: "created_time",
		Order:   "desc",
		Limit:   1000, // 导出时设置较大的限制
		Offset:  0,
	}

	if spaceID := c.Query("space_id"); spaceID != "" {
		filter.SpaceID = &spaceID
	}
	if search := c.Query("search"); search != "" {
		filter.Search = search
	}

	bases, err := h.baseService.ExportBases(c.Request.Context(), filter)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, bases)
}

// ImportBases 导入基础表
// @Summary 导入基础表
// @Description 从JSON格式导入基础表数据
// @Tags 基础表
// @Accept json
// @Produce json
// @Param request body []base.CreateBaseRequest true "基础表数据列表"
// @Success 201 {array} base.Base
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/bases/import [post]
func (h *BaseHandler) ImportBases(c *gin.Context) {
	var baseReqs []base.CreateBaseRequest
	if err := c.ShouldBindJSON(&baseReqs); err != nil {
		response.Error(c, errors.ErrBadRequest.WithDetails(err.Error()))
		return
	}

	// 从JWT中获取用户ID并设置到所有请求中
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	for i := range baseReqs {
		baseReqs[i].CreatedBy = userID.(string)
	}

	bases, err := h.baseService.ImportBases(c.Request.Context(), baseReqs)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, bases, "")
}

// handleError 处理错误
func (h *BaseHandler) handleError(c *gin.Context, err error) {
	response.Error(c, err)
}
