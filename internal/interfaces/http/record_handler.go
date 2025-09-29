package http

import (
	"github.com/gin-gonic/gin"

	"teable-go-backend/internal/application"
	recdomain "teable-go-backend/internal/domain/record"
	"teable-go-backend/pkg/errors"
	"teable-go-backend/pkg/response"
)

// RecordHandler 记录处理器
type RecordHandler struct {
	recordService *application.RecordService
}

// NewRecordHandler 创建新的记录处理器
func NewRecordHandler(recordService *application.RecordService) *RecordHandler {
	return &RecordHandler{recordService: recordService}
}

// CreateRecord 创建记录
// @Summary 创建记录
// @Description 创建一个新的数据记录
// @Tags 记录
// @Accept json
// @Produce json
// @Param request body record.CreateRecordRequest true "创建记录请求"
// @Success 201 {object} Response{data=record.Record}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/records [post]
func (h *RecordHandler) CreateRecord(c *gin.Context) {
	var req recdomain.CreateRecordRequest
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

	newRecord, err := h.recordService.CreateRecord(c.Request.Context(), req, userID.(string))
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, newRecord, "")
}

// GetRecord 获取记录详情
// @Summary 获取记录详情
// @Description 根据ID获取记录详情
// @Tags 记录
// @Produce json
// @Param id path string true "记录ID"
// @Success 200 {object} Response{data=record.Record}
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/records/{id} [get]
func (h *RecordHandler) GetRecord(c *gin.Context) {
	recordID := c.Param("id")
	userID, _ := c.Get("user_id")
	r, err := h.recordService.GetRecord(c.Request.Context(), recordID, userID.(string))
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.SuccessWithMessage(c, r, "")
}

// UpdateRecord 更新记录
// @Summary 更新记录
// @Description 更新指定ID的记录数据
// @Tags 记录
// @Accept json
// @Produce json
// @Param id path string true "记录ID"
// @Param request body record.UpdateRecordRequest true "更新记录请求"
// @Success 200 {object} Response{data=record.Record}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/records/{id} [put]
func (h *RecordHandler) UpdateRecord(c *gin.Context) {
	recordID := c.Param("id")
	var req recdomain.UpdateRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	userID, _ := c.Get("user_id")
	updatedRecord, err := h.recordService.UpdateRecord(c.Request.Context(), recordID, req, userID.(string))
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.SuccessWithMessage(c, updatedRecord, "")
}

// DeleteRecord 删除记录
// @Summary 删除记录
// @Description 删除指定ID的记录
// @Tags 记录
// @Produce json
// @Param id path string true "记录ID"
// @Success 200 {object} Response{success=boolean}
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/records/{id} [delete]
func (h *RecordHandler) DeleteRecord(c *gin.Context) {
	recordID := c.Param("id")
	userID, _ := c.Get("user_id")
	err := h.recordService.DeleteRecord(c.Request.Context(), recordID, userID.(string))
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.SuccessWithMessage(c, map[string]bool{"success": true}, "")
}

// ListRecords 列出记录
// @Summary 列出记录
// @Description 获取记录列表
// @Tags 记录
// @Produce json
// @Param table_id query string false "数据表ID"
// @Param created_by query string false "创建者ID"
// @Param search query string false "搜索关键词"
// @Param order_by query string false "排序字段"
// @Param order query string false "排序方式 (asc/desc)"
// @Param limit query int false "每页数量"
// @Param offset query int false "偏移量"
// @Success 200 {object} Response{data=[]record.Record,total=int64,limit=int,offset=int}
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/records [get]
func (h *RecordHandler) ListRecords(c *gin.Context) {
	var filter recdomain.ListRecordFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	// 从JWT中获取用户ID，用于过滤用户自己的记录
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

	records, total, err := h.recordService.ListRecords(c.Request.Context(), filter, userID.(string))
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.PaginatedSuccess(c, records, response.Pagination{
		Page:       0,
		Limit:      filter.Limit,
		Total:      int(total),
		TotalPages: 0,
	}, "")
}

// BulkCreateRecords 批量创建记录
// @Summary 批量创建记录
// @Description 批量创建多个数据记录
// @Tags 记录
// @Accept json
// @Produce json
// @Param request body []record.CreateRecordRequest true "批量创建记录请求"
// @Success 201 {object} Response{data=[]record.Record}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/records/bulk [post]
func (h *RecordHandler) BulkCreateRecords(c *gin.Context) {
	var reqs []recdomain.CreateRecordRequest
	if err := c.ShouldBindJSON(&reqs); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	// 从JWT中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		h.handleError(c, errors.ErrUnauthorized)
		return
	}

	// 设置创建者
	for i := range reqs {
		reqs[i].CreatedBy = userID.(string)
	}

	// 应用服务未提供批量接口，这里逐条调用以保持行为
	created := make([]*recdomain.Record, 0, len(reqs))
	for _, r := range reqs {
		rec, err := h.recordService.CreateRecord(c.Request.Context(), r, userID.(string))
		if err != nil {
			h.handleError(c, err)
			return
		}
		created = append(created, rec)
	}
	response.SuccessWithMessage(c, created, "")
}

// BulkUpdateRecords 批量更新记录
// @Summary 批量更新记录
// @Description 批量更新多个记录的数据
// @Tags 记录
// @Accept json
// @Produce json
// @Param request body record.BulkUpdateRequest true "批量更新记录请求"
// @Success 200 {object} Response{success=boolean}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/records/bulk [put]
func (h *RecordHandler) BulkUpdateRecords(c *gin.Context) {
	var req recdomain.BulkUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	// 未提供批量更新：返回 400 或者逐条更新（这里简单返回不支持）
	h.handleError(c, errors.ErrInvalidRequest.WithDetails("bulk update not supported in application service"))
	return
	// 已返回错误，上面已 return
}

// BulkDeleteRecords 批量删除记录
// @Summary 批量删除记录
// @Description 批量删除多个记录
// @Tags 记录
// @Accept json
// @Produce json
// @Param request body record.BulkDeleteRequest true "批量删除记录请求"
// @Success 200 {object} Response{success=boolean}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/records/bulk [delete]
func (h *RecordHandler) BulkDeleteRecords(c *gin.Context) {
	var req recdomain.BulkDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	h.handleError(c, errors.ErrInvalidRequest.WithDetails("bulk delete not supported in application service"))
	return
	// 已返回错误
}

// ComplexQuery 复杂查询
// @Summary 复杂查询记录
// @Description 使用复杂条件查询记录
// @Tags 记录
// @Accept json
// @Produce json
// @Param request body record.ComplexQueryRequest true "复杂查询请求"
// @Success 200 {object} Response{data=[]map[string]interface{}}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/records/query [post]
func (h *RecordHandler) ComplexQuery(c *gin.Context) {
	var req recdomain.ComplexQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	h.handleError(c, errors.ErrInvalidRequest.WithDetails("complex query not supported in application service"))
	return
}

// GetRecordStats 获取记录统计信息
// @Summary 获取记录统计信息
// @Description 获取记录的统计信息
// @Tags 记录
// @Produce json
// @Param table_id query string false "表ID"
// @Success 200 {object} Response{data=record.RecordStats}
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/records/stats [get]
func (h *RecordHandler) GetRecordStats(c *gin.Context) {
	_ = c.Query("table_id")
	h.handleError(c, errors.ErrInvalidRequest.WithDetails("record stats not supported in application service"))
	return
}

// ExportRecords 导出记录
// @Summary 导出记录
// @Description 导出记录数据
// @Tags 记录
// @Accept json
// @Produce application/octet-stream
// @Param request body record.ExportRequest true "导出请求"
// @Success 200 {file} file
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/records/export [post]
func (h *RecordHandler) ExportRecords(c *gin.Context) {
	var req recdomain.ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	h.handleError(c, errors.ErrInvalidRequest.WithDetails("export not supported in application service"))
	return
}

// ImportRecords 导入记录
// @Summary 导入记录
// @Description 导入记录数据
// @Tags 记录
// @Accept json
// @Produce json
// @Param request body record.ImportRequest true "导入请求"
// @Success 200 {object} Response{data=int}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/records/import [post]
func (h *RecordHandler) ImportRecords(c *gin.Context) {
	var req recdomain.ImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	h.handleError(c, errors.ErrInvalidRequest.WithDetails("import not supported in application service"))
	return
}

func (h *RecordHandler) handleError(c *gin.Context, err error) {
	response.Error(c, err)
}
