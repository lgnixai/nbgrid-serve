package http

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"teable-go-backend/internal/domain/table"
	"teable-go-backend/pkg/errors"
	"teable-go-backend/pkg/response"
)

// TableHandler 数据表处理器
type TableHandler struct {
	tableService table.Service
}

// NewTableHandler 创建新的数据表处理器
func NewTableHandler(tableService table.Service) *TableHandler {
	return &TableHandler{tableService: tableService}
}

// CreateTable 创建数据表
// @Summary 创建数据表
// @Description 创建一个新的数据表
// @Tags 数据表
// @Accept json
// @Produce json
// @Param request body table.CreateTableRequest true "创建数据表请求"
// @Success 201 {object} Response{data=table.Table}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/tables [post]
func (h *TableHandler) CreateTable(c *gin.Context) {
	var req table.CreateTableRequest
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

	// 从URL参数中获取base_id（如果存在）
	if baseID := c.Param("id"); baseID != "" {
		req.BaseID = baseID
	}

	newTable, err := h.tableService.CreateTable(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, newTable, "")
}

// GetTable 获取数据表详情
// @Summary 获取数据表详情
// @Description 根据ID获取数据表详情
// @Tags 数据表
// @Produce json
// @Param id path string true "数据表ID"
// @Success 200 {object} Response{data=table.Table}
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/tables/{id} [get]
func (h *TableHandler) GetTable(c *gin.Context) {
	tableID := c.Param("id")
	t, err := h.tableService.GetTable(c.Request.Context(), tableID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.SuccessWithMessage(c, t, "")
}

// UpdateTable 更新数据表
// @Summary 更新数据表
// @Description 更新指定ID的数据表信息
// @Tags 数据表
// @Accept json
// @Produce json
// @Param id path string true "数据表ID"
// @Param request body table.UpdateTableRequest true "更新数据表请求"
// @Success 200 {object} Response{data=table.Table}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/tables/{id} [put]
func (h *TableHandler) UpdateTable(c *gin.Context) {
	tableID := c.Param("id")
	var req table.UpdateTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	updatedTable, err := h.tableService.UpdateTable(c.Request.Context(), tableID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.SuccessWithMessage(c, updatedTable, "")
}

// DeleteTable 删除数据表
// @Summary 删除数据表
// @Description 删除指定ID的数据表
// @Tags 数据表
// @Produce json
// @Param id path string true "数据表ID"
// @Success 200 {object} Response{success=boolean}
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/tables/{id} [delete]
func (h *TableHandler) DeleteTable(c *gin.Context) {
	tableID := c.Param("id")
	err := h.tableService.DeleteTable(c.Request.Context(), tableID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.SuccessWithMessage(c, map[string]bool{"success": true}, "")
}

// ListTables 列出数据表
// @Summary 列出数据表
// @Description 获取数据表列表
// @Tags 数据表
// @Produce json
// @Param base_id query string false "基础表ID"
// @Param name query string false "名称"
// @Param created_by query string false "创建者ID"
// @Param search query string false "搜索关键词"
// @Param order_by query string false "排序字段"
// @Param order query string false "排序方式 (asc/desc)"
// @Param limit query int false "每页数量"
// @Param offset query int false "偏移量"
// @Success 200 {object} Response{data=[]table.Table,total=int64,limit=int,offset=int}
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/tables [get]
func (h *TableHandler) ListTables(c *gin.Context) {
	var filter table.ListTableFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	// 从JWT中获取用户ID，用于过滤用户自己的数据表
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

	tables, total, err := h.tableService.ListTables(c.Request.Context(), filter)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.PaginatedSuccess(c, tables, response.Pagination{
		Page:       0,
		Limit:      filter.Limit,
		Total:      int(total),
		TotalPages: 0,
	}, "")
}

// CreateField 创建字段
// @Summary 创建字段
// @Description 创建一个新的字段
// @Tags 字段
// @Accept json
// @Produce json
// @Param request body table.CreateFieldRequest true "创建字段请求"
// @Success 201 {object} Response{data=table.Field}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/fields [post]
func (h *TableHandler) CreateField(c *gin.Context) {
	var req table.CreateFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	// 添加调试日志
	fmt.Printf("CreateField request: %+v\n", req)

	// 从JWT中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		h.handleError(c, errors.ErrUnauthorized)
		return
	}
	req.CreatedBy = userID.(string)

	newField, err := h.tableService.CreateField(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 添加调试日志
	fmt.Printf("Created field: %+v\n", newField)

	response.SuccessWithMessage(c, newField, "")
}

// GetField 获取字段详情
// @Summary 获取字段详情
// @Description 根据ID获取字段详情
// @Tags 字段
// @Produce json
// @Param id path string true "字段ID"
// @Success 200 {object} Response{data=table.Field}
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/fields/{id} [get]
func (h *TableHandler) GetField(c *gin.Context) {
	fieldID := c.Param("id")
	f, err := h.tableService.GetField(c.Request.Context(), fieldID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.SuccessWithMessage(c, f, "")
}

// UpdateField 更新字段
// @Summary 更新字段
// @Description 更新指定ID的字段信息
// @Tags 字段
// @Accept json
// @Produce json
// @Param id path string true "字段ID"
// @Param request body table.UpdateFieldRequest true "更新字段请求"
// @Success 200 {object} Response{data=table.Field}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/fields/{id} [put]
func (h *TableHandler) UpdateField(c *gin.Context) {
	fieldID := c.Param("id")
	var req table.UpdateFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	updatedField, err := h.tableService.UpdateField(c.Request.Context(), fieldID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.SuccessWithMessage(c, updatedField, "")
}

// DeleteField 删除字段
// @Summary 删除字段
// @Description 删除指定ID的字段
// @Tags 字段
// @Produce json
// @Param id path string true "字段ID"
// @Success 200 {object} Response{success=boolean}
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/fields/{id} [delete]
func (h *TableHandler) DeleteField(c *gin.Context) {
	fieldID := c.Param("id")
	err := h.tableService.DeleteField(c.Request.Context(), fieldID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.SuccessWithMessage(c, map[string]bool{"success": true}, "")
}

// ListFields 列出字段
// @Summary 列出字段
// @Description 获取字段列表
// @Tags 字段
// @Produce json
// @Param table_id query string false "数据表ID"
// @Param name query string false "名称"
// @Param type query string false "类型"
// @Param created_by query string false "创建者ID"
// @Param order_by query string false "排序字段"
// @Param order query string false "排序方式 (asc/desc)"
// @Param limit query int false "每页数量"
// @Param offset query int false "偏移量"
// @Success 200 {object} Response{data=[]table.Field,total=int64,limit=int,offset=int}
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/fields [get]
func (h *TableHandler) ListFields(c *gin.Context) {
	var filter table.ListFieldFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	// 从JWT中获取用户ID，用于过滤用户自己的字段
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

	fields, total, err := h.tableService.ListFields(c.Request.Context(), filter)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.PaginatedSuccess(c, fields, response.Pagination{
		Page:       0,
		Limit:      filter.Limit,
		Total:      int(total),
		TotalPages: 0,
	}, "")
}

// GetFieldTypes 获取字段类型列表
// @Summary 获取字段类型列表
// @Description 获取所有可用的字段类型及其信息
// @Tags 字段管理
// @Accept json
// @Produce json
// @Success 200 {array} table.FieldTypeInfo
// @Failure 500 {object} ErrorResponse
// @Router /api/fields/types [get]
func (h *TableHandler) GetFieldTypes(c *gin.Context) {
	types, err := h.tableService.GetFieldTypes(c.Request.Context())
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.SuccessWithMessage(c, types, "")
}

// GetFieldTypeInfo 获取字段类型信息
// @Summary 获取字段类型信息
// @Description 获取指定字段类型的详细信息
// @Tags 字段管理
// @Accept json
// @Produce json
// @Param type path string true "字段类型"
// @Success 200 {object} table.FieldTypeInfo
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/fields/types/{type} [get]
func (h *TableHandler) GetFieldTypeInfo(c *gin.Context) {
	fieldType := c.Param("type")
	if fieldType == "" {
		response.Error(c, errors.ErrBadRequest.WithDetails("字段类型不能为空"))
		return
	}

	info, err := h.tableService.GetFieldTypeInfo(c.Request.Context(), table.FieldType(fieldType))
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.SuccessWithMessage(c, info, "")
}

// GetFieldShortcuts 获取字段捷径列表
// @Summary 获取字段捷径列表
// @Description 获取所有可用的字段捷径模板
// @Tags 字段管理
// @Accept json
// @Produce json
// @Param category query string false "分类过滤"
// @Param tag query string false "标签过滤"
// @Success 200 {array} table.FieldShortcut
// @Failure 500 {object} ErrorResponse
// @Router /api/fields/shortcuts [get]
func (h *TableHandler) GetFieldShortcuts(c *gin.Context) {
	category := c.Query("category")
	tag := c.Query("tag")
	
	var shortcuts []table.FieldShortcut
	
	if category != "" {
		shortcuts = table.GetFieldShortcutsByCategory(category)
	} else if tag != "" {
		shortcuts = table.GetFieldShortcutsByTag(tag)
	} else {
		shortcuts = table.FieldShortcuts
	}
	
	response.SuccessWithMessage(c, shortcuts, "")
}

// GetFieldShortcut 获取字段捷径详情
// @Summary 获取字段捷径详情
// @Description 获取指定字段捷径的详细信息
// @Tags 字段管理
// @Accept json
// @Produce json
// @Param id path string true "字段捷径ID"
// @Success 200 {object} table.FieldShortcut
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/fields/shortcuts/{id} [get]
func (h *TableHandler) GetFieldShortcut(c *gin.Context) {
	shortcutID := c.Param("id")
	if shortcutID == "" {
		response.Error(c, errors.ErrBadRequest.WithDetails("字段捷径ID不能为空"))
		return
	}

	shortcut, err := table.GetFieldShortcutByID(shortcutID)
	if err != nil {
		response.Error(c, errors.ErrNotFound.WithDetails(err.Error()))
		return
	}
	
	response.SuccessWithMessage(c, shortcut, "")
}

// ValidateFieldValue 验证字段值
// @Summary 验证字段值
// @Description 验证字段值是否符合字段的验证规则
// @Tags 字段管理
// @Accept json
// @Produce json
// @Param field_id path string true "字段ID"
// @Param request body ValidateFieldValueRequest true "验证请求"
// @Success 200 {object} ValidateFieldValueResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/fields/{field_id}/validate [post]
func (h *TableHandler) ValidateFieldValue(c *gin.Context) {
	fieldID := c.Param("field_id")
	if fieldID == "" {
		response.Error(c, errors.ErrBadRequest.WithDetails("字段ID不能为空"))
		return
	}

	var req struct {
		Value interface{} `json:"value" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrBadRequest.WithDetails(err.Error()))
		return
	}

	// 获取字段信息
	field, err := h.tableService.GetField(c.Request.Context(), fieldID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 验证字段值
	err = h.tableService.ValidateFieldValue(c.Request.Context(), field, req.Value)
	if err != nil {
		response.SuccessWithMessage(c, gin.H{"valid": false, "error": err.Error()}, "")
		return
	}

	response.SuccessWithMessage(c, gin.H{"valid": true, "error": nil}, "")
}

func (h *TableHandler) handleError(c *gin.Context, err error) {
	response.Error(c, err)
}
