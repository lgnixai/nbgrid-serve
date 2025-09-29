package http

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"teable-go-backend/internal/domain/view"
	"teable-go-backend/pkg/errors"
	"teable-go-backend/pkg/response"
)

// ViewHandler 视图处理器
type ViewHandler struct {
	viewService view.Service
}

// NewViewHandler 创建新的视图处理器
func NewViewHandler(viewService view.Service) *ViewHandler {
	return &ViewHandler{viewService: viewService}
}

// CreateView 创建视图
// @Summary 创建视图
// @Description 创建一个新的数据视图
// @Tags 视图
// @Accept json
// @Produce json
// @Param request body view.CreateViewRequest true "创建视图请求"
// @Success 201 {object} Response{data=view.View}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/views [post]
func (h *ViewHandler) CreateView(c *gin.Context) {
	var req view.CreateViewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	// 从JWT中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		h.handleError(c, errors.ErrUnauthorized)
		return
	}
	req.CreatedBy = userID.(string)

	newView, err := h.viewService.CreateView(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, newView, "")
}

// GetView 获取视图详情
// @Summary 获取视图详情
// @Description 根据ID获取视图详情
// @Tags 视图
// @Produce json
// @Param id path string true "视图ID"
// @Success 200 {object} Response{data=view.View}
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/views/{id} [get]
func (h *ViewHandler) GetView(c *gin.Context) {
	viewID := c.Param("id")
	v, err := h.viewService.GetView(c.Request.Context(), viewID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.SuccessWithMessage(c, v, "")
}

// UpdateView 更新视图
// @Summary 更新视图
// @Description 更新指定ID的视图信息
// @Tags 视图
// @Accept json
// @Produce json
// @Param id path string true "视图ID"
// @Param request body view.UpdateViewRequest true "更新视图请求"
// @Success 200 {object} Response{data=view.View}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/views/{id} [put]
func (h *ViewHandler) UpdateView(c *gin.Context) {
	viewID := c.Param("id")
	var req view.UpdateViewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	updatedView, err := h.viewService.UpdateView(c.Request.Context(), viewID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.SuccessWithMessage(c, updatedView, "")
}

// DeleteView 删除视图
// @Summary 删除视图
// @Description 删除指定ID的视图
// @Tags 视图
// @Produce json
// @Param id path string true "视图ID"
// @Success 200 {object} Response{success=boolean}
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/views/{id} [delete]
func (h *ViewHandler) DeleteView(c *gin.Context) {
	viewID := c.Param("id")
	err := h.viewService.DeleteView(c.Request.Context(), viewID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.SuccessWithMessage(c, map[string]bool{"success": true}, "")
}

// ListViews 列出视图
// @Summary 列出视图
// @Description 获取视图列表
// @Tags 视图
// @Produce json
// @Param table_id query string false "数据表ID"
// @Param name query string false "名称"
// @Param type query string false "视图类型"
// @Param created_by query string false "创建者ID"
// @Param search query string false "搜索关键词"
// @Param order_by query string false "排序字段"
// @Param order query string false "排序方式 (asc/desc)"
// @Param limit query int false "每页数量"
// @Param offset query int false "偏移量"
// @Success 200 {object} Response{data=[]view.View,total=int64,limit=int,offset=int}
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/views [get]
func (h *ViewHandler) ListViews(c *gin.Context) {
	var filter view.ListViewFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	// 从JWT中获取用户ID，用于过滤用户自己的视图
	userID, exists := c.Get("user_id")
	if !exists {
		h.handleError(c, errors.ErrUnauthorized)
		return
	}
	// 如果没有指定created_by，则默认查询当前用户创建的
	if filter.CreatedBy == nil || *filter.CreatedBy == "" {
		filter.CreatedBy = new(string)
		*filter.CreatedBy = userID.(string)
	}

	views, total, err := h.viewService.ListViews(c.Request.Context(), filter)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.PaginatedSuccess(c, views, response.Pagination{
		Page:       0,
		Limit:      filter.Limit,
		Total:      int(total),
		TotalPages: 0,
	}, "")
}

// GetGridViewData 获取网格视图数据
// @Summary 获取网格视图数据
// @Description 获取网格视图的数据和配置
// @Tags 视图
// @Produce json
// @Param id path string true "视图ID"
// @Param page query int false "页码"
// @Param page_size query int false "每页大小"
// @Success 200 {object} Response{data=view.GridViewData}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/views/{id}/grid/data [get]
func (h *ViewHandler) GetGridViewData(c *gin.Context) {
	viewID := c.Param("id")
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "20")

	pageInt := 1
	pageSizeInt := 20

	if p, err := strconv.Atoi(page); err == nil && p > 0 {
		pageInt = p
	}
	if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 && ps <= 1000 {
		pageSizeInt = ps
	}

	data, err := h.viewService.GetGridViewData(c.Request.Context(), viewID, pageInt, pageSizeInt)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, data, "")
}

// UpdateGridViewConfig 更新网格视图配置
// @Summary 更新网格视图配置
// @Description 更新网格视图的配置信息
// @Tags 视图
// @Accept json
// @Produce json
// @Param id path string true "视图ID"
// @Param request body view.GridViewConfig true "网格视图配置"
// @Success 200 {object} Response{success=boolean}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/views/{id}/grid/config [put]
func (h *ViewHandler) UpdateGridViewConfig(c *gin.Context) {
	viewID := c.Param("id")
	var config view.GridViewConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	err := h.viewService.UpdateGridViewConfig(c.Request.Context(), viewID, config)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, map[string]bool{"success": true}, "")
}

// AddGridViewColumn 添加网格视图列
// @Summary 添加网格视图列
// @Description 向网格视图添加新列
// @Tags 视图
// @Accept json
// @Produce json
// @Param id path string true "视图ID"
// @Param request body view.GridViewColumn true "列配置"
// @Success 200 {object} Response{success=boolean}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/views/{id}/grid/columns [post]
func (h *ViewHandler) AddGridViewColumn(c *gin.Context) {
	viewID := c.Param("id")
	var column view.GridViewColumn
	if err := c.ShouldBindJSON(&column); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	err := h.viewService.AddGridViewColumn(c.Request.Context(), viewID, column)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, map[string]bool{"success": true}, "")
}

// UpdateGridViewColumn 更新网格视图列
// @Summary 更新网格视图列
// @Description 更新网格视图中的指定列
// @Tags 视图
// @Accept json
// @Produce json
// @Param id path string true "视图ID"
// @Param field_id path string true "字段ID"
// @Param request body view.GridViewColumn true "列配置"
// @Success 200 {object} Response{success=boolean}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/views/{id}/grid/columns/{field_id} [put]
func (h *ViewHandler) UpdateGridViewColumn(c *gin.Context) {
	viewID := c.Param("id")
	fieldID := c.Param("field_id")
	var column view.GridViewColumn
	if err := c.ShouldBindJSON(&column); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	err := h.viewService.UpdateGridViewColumn(c.Request.Context(), viewID, fieldID, column)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, map[string]bool{"success": true}, "")
}

// RemoveGridViewColumn 移除网格视图列
// @Summary 移除网格视图列
// @Description 从网格视图中移除指定列
// @Tags 视图
// @Produce json
// @Param id path string true "视图ID"
// @Param field_id path string true "字段ID"
// @Success 200 {object} Response{success=boolean}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/views/{id}/grid/columns/{field_id} [delete]
func (h *ViewHandler) RemoveGridViewColumn(c *gin.Context) {
	viewID := c.Param("id")
	fieldID := c.Param("field_id")

	err := h.viewService.RemoveGridViewColumn(c.Request.Context(), viewID, fieldID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, map[string]bool{"success": true}, "")
}

// ReorderGridViewColumns 重新排序网格视图列
// @Summary 重新排序网格视图列
// @Description 重新排序网格视图中的列
// @Tags 视图
// @Accept json
// @Produce json
// @Param id path string true "视图ID"
// @Param request body []string true "字段ID列表"
// @Success 200 {object} Response{success=boolean}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/views/{id}/grid/columns/reorder [put]
func (h *ViewHandler) ReorderGridViewColumns(c *gin.Context) {
	viewID := c.Param("id")
	var fieldIDs []string
	if err := c.ShouldBindJSON(&fieldIDs); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	err := h.viewService.ReorderGridViewColumns(c.Request.Context(), viewID, fieldIDs)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, map[string]bool{"success": true}, "")
}

// GetViewConfig 获取视图配置
// @Summary 获取视图配置
// @Description 获取视图的配置信息
// @Tags 视图
// @Produce json
// @Param id path string true "视图ID"
// @Success 200 {object} Response{data=map[string]interface{}}
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/views/{id}/config [get]
func (h *ViewHandler) GetViewConfig(c *gin.Context) {
	viewID := c.Param("id")

	config, err := h.viewService.GetViewConfig(c.Request.Context(), viewID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, config, "")
}

// UpdateViewConfig 更新视图配置
// @Summary 更新视图配置
// @Description 更新视图的配置信息
// @Tags 视图
// @Accept json
// @Produce json
// @Param id path string true "视图ID"
// @Param request body map[string]interface{} true "视图配置"
// @Success 200 {object} Response{success=boolean}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/views/{id}/config [put]
func (h *ViewHandler) UpdateViewConfig(c *gin.Context) {
	viewID := c.Param("id")
	var config map[string]interface{}
	if err := c.ShouldBindJSON(&config); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	err := h.viewService.UpdateViewConfig(c.Request.Context(), viewID, config)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, map[string]bool{"success": true}, "")
}

// GetFormViewData 获取表单视图数据
// @Summary 获取表单视图数据
// @Description 获取表单视图的字段配置和表单配置
// @Tags 视图
// @Produce json
// @Param id path string true "视图ID"
// @Success 200 {object} Response{data=view.FormViewData}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/views/{id}/form/data [get]
func (h *ViewHandler) GetFormViewData(c *gin.Context) {
	viewID := c.Param("id")

	data, err := h.viewService.GetFormViewData(c.Request.Context(), viewID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, data, "")
}

// UpdateFormViewConfig 更新表单视图配置
// @Summary 更新表单视图配置
// @Description 更新表单视图的配置信息
// @Tags 视图
// @Accept json
// @Produce json
// @Param id path string true "视图ID"
// @Param request body view.FormViewConfig true "表单视图配置"
// @Success 200 {object} Response{success=boolean}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/views/{id}/form/config [put]
func (h *ViewHandler) UpdateFormViewConfig(c *gin.Context) {
	viewID := c.Param("id")
	var config view.FormViewConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	err := h.viewService.UpdateFormViewConfig(c.Request.Context(), viewID, config)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, map[string]bool{"success": true}, "")
}

// AddFormViewField 添加表单视图字段
// @Summary 添加表单视图字段
// @Description 向表单视图添加新字段
// @Tags 视图
// @Accept json
// @Produce json
// @Param id path string true "视图ID"
// @Param request body view.FormViewField true "字段配置"
// @Success 200 {object} Response{success=boolean}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/views/{id}/form/fields [post]
func (h *ViewHandler) AddFormViewField(c *gin.Context) {
	viewID := c.Param("id")
	var field view.FormViewField
	if err := c.ShouldBindJSON(&field); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	err := h.viewService.AddFormViewField(c.Request.Context(), viewID, field)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, map[string]bool{"success": true}, "")
}

// UpdateFormViewField 更新表单视图字段
// @Summary 更新表单视图字段
// @Description 更新表单视图中的指定字段
// @Tags 视图
// @Accept json
// @Produce json
// @Param id path string true "视图ID"
// @Param field_id path string true "字段ID"
// @Param request body view.FormViewField true "字段配置"
// @Success 200 {object} Response{success=boolean}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/views/{id}/form/fields/{field_id} [put]
func (h *ViewHandler) UpdateFormViewField(c *gin.Context) {
	viewID := c.Param("id")
	fieldID := c.Param("field_id")
	var field view.FormViewField
	if err := c.ShouldBindJSON(&field); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	err := h.viewService.UpdateFormViewField(c.Request.Context(), viewID, fieldID, field)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, map[string]bool{"success": true}, "")
}

// RemoveFormViewField 移除表单视图字段
// @Summary 移除表单视图字段
// @Description 从表单视图中移除指定字段
// @Tags 视图
// @Produce json
// @Param id path string true "视图ID"
// @Param field_id path string true "字段ID"
// @Success 200 {object} Response{success=boolean}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/views/{id}/form/fields/{field_id} [delete]
func (h *ViewHandler) RemoveFormViewField(c *gin.Context) {
	viewID := c.Param("id")
	fieldID := c.Param("field_id")

	err := h.viewService.RemoveFormViewField(c.Request.Context(), viewID, fieldID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, map[string]bool{"success": true}, "")
}

// ReorderFormViewFields 重新排序表单视图字段
// @Summary 重新排序表单视图字段
// @Description 重新排序表单视图中的字段
// @Tags 视图
// @Accept json
// @Produce json
// @Param id path string true "视图ID"
// @Param request body []string true "字段ID列表"
// @Success 200 {object} Response{success=boolean}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/views/{id}/form/fields/reorder [put]
func (h *ViewHandler) ReorderFormViewFields(c *gin.Context) {
	viewID := c.Param("id")
	var fieldIDs []string
	if err := c.ShouldBindJSON(&fieldIDs); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	err := h.viewService.ReorderFormViewFields(c.Request.Context(), viewID, fieldIDs)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, map[string]bool{"success": true}, "")
}

// GetKanbanViewData 获取看板视图数据
// @Summary 获取看板视图数据
// @Description 获取看板视图的分组和卡片数据
// @Tags 视图
// @Produce json
// @Param id path string true "视图ID"
// @Success 200 {object} Response{data=view.KanbanViewData}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/views/{id}/kanban/data [get]
func (h *ViewHandler) GetKanbanViewData(c *gin.Context) {
	viewID := c.Param("id")

	data, err := h.viewService.GetKanbanViewData(c.Request.Context(), viewID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, data, "")
}

// UpdateKanbanViewConfig 更新看板视图配置
// @Summary 更新看板视图配置
// @Description 更新看板视图的配置信息
// @Tags 视图
// @Accept json
// @Produce json
// @Param id path string true "视图ID"
// @Param request body view.KanbanViewConfig true "看板视图配置"
// @Success 200 {object} Response{success=boolean}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/views/{id}/kanban/config [put]
func (h *ViewHandler) UpdateKanbanViewConfig(c *gin.Context) {
	viewID := c.Param("id")
	var config view.KanbanViewConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	err := h.viewService.UpdateKanbanViewConfig(c.Request.Context(), viewID, config)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, map[string]bool{"success": true}, "")
}

// MoveKanbanCard 移动看板卡片
// @Summary 移动看板卡片
// @Description 在看板视图中移动卡片到不同的分组
// @Tags 视图
// @Accept json
// @Produce json
// @Param id path string true "视图ID"
// @Param request body view.MoveKanbanCardRequest true "移动卡片请求"
// @Success 200 {object} Response{success=boolean}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/views/{id}/kanban/move [post]
func (h *ViewHandler) MoveKanbanCard(c *gin.Context) {
	viewID := c.Param("id")
	var req view.MoveKanbanCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	req.ViewID = viewID

	err := h.viewService.MoveKanbanCard(c.Request.Context(), req.ViewID, req.RecordID, req.FromGroup, req.ToGroup)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, map[string]bool{"success": true}, "")
}

// GetCalendarViewData 获取日历视图数据
// @Summary 获取日历视图数据
// @Description 获取日历视图的事件数据
// @Tags 视图
// @Produce json
// @Param id path string true "视图ID"
// @Param start_date query string false "开始日期"
// @Param end_date query string false "结束日期"
// @Success 200 {object} Response{data=view.CalendarViewData}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/views/{id}/calendar/data [get]
func (h *ViewHandler) GetCalendarViewData(c *gin.Context) {
	viewID := c.Param("id")
	startDate := c.DefaultQuery("start_date", "")
	endDate := c.DefaultQuery("end_date", "")

	data, err := h.viewService.GetCalendarViewData(c.Request.Context(), viewID, startDate, endDate)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, data, "")
}

// UpdateCalendarViewConfig 更新日历视图配置
// @Summary 更新日历视图配置
// @Description 更新日历视图的配置信息
// @Tags 视图
// @Accept json
// @Produce json
// @Param id path string true "视图ID"
// @Param request body view.CalendarViewConfig true "日历视图配置"
// @Success 200 {object} Response{success=boolean}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/views/{id}/calendar/config [put]
func (h *ViewHandler) UpdateCalendarViewConfig(c *gin.Context) {
	viewID := c.Param("id")
	var config view.CalendarViewConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	err := h.viewService.UpdateCalendarViewConfig(c.Request.Context(), viewID, config)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, map[string]bool{"success": true}, "")
}

// GetGalleryViewData 获取画廊视图数据
// @Summary 获取画廊视图数据
// @Description 获取画廊视图的卡片数据
// @Tags 视图
// @Produce json
// @Param id path string true "视图ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Success 200 {object} Response{data=view.GalleryViewData}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/views/{id}/gallery/data [get]
func (h *ViewHandler) GetGalleryViewData(c *gin.Context) {
	viewID := c.Param("id")
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "20")

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails("页码必须是数字"))
		return
	}

	pageSizeInt, err := strconv.Atoi(pageSize)
	if err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails("每页大小必须是数字"))
		return
	}

	data, err := h.viewService.GetGalleryViewData(c.Request.Context(), viewID, pageInt, pageSizeInt)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, data, "")
}

// UpdateGalleryViewConfig 更新画廊视图配置
// @Summary 更新画廊视图配置
// @Description 更新画廊视图的配置信息
// @Tags 视图
// @Accept json
// @Produce json
// @Param id path string true "视图ID"
// @Param request body view.GalleryViewConfig true "画廊视图配置"
// @Success 200 {object} Response{success=boolean}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/views/{id}/gallery/config [put]
func (h *ViewHandler) UpdateGalleryViewConfig(c *gin.Context) {
	viewID := c.Param("id")
	var config view.GalleryViewConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	err := h.viewService.UpdateGalleryViewConfig(c.Request.Context(), viewID, config)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, map[string]bool{"success": true}, "")
}

func (h *ViewHandler) handleError(c *gin.Context, err error) {
	response.Error(c, err)
}
