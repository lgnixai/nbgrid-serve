package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"teable-go-backend/internal/domain/record"
	"teable-go-backend/pkg/errors"
	"teable-go-backend/pkg/logger"
)

// RecordHandler 记录处理器
type RecordHandler struct {
	recordService record.Service
}

// NewRecordHandler 创建新的记录处理器
func NewRecordHandler(recordService record.Service) *RecordHandler {
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
	var req record.CreateRecordRequest
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

	newRecord, err := h.recordService.CreateRecord(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{Data: newRecord})
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
	r, err := h.recordService.GetRecord(c.Request.Context(), recordID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, SuccessResponse{Data: r})
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
	var req record.UpdateRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	updatedRecord, err := h.recordService.UpdateRecord(c.Request.Context(), recordID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, SuccessResponse{Data: updatedRecord})
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
	err := h.recordService.DeleteRecord(c.Request.Context(), recordID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, SuccessResponse{Success: true})
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
	var filter record.ListRecordFilter
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

	records, total, err := h.recordService.ListRecords(c.Request.Context(), filter)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, PaginatedResponse{
		Data:   records,
		Total:  total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	})
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
	var reqs []record.CreateRecordRequest
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

	records, err := h.recordService.BulkCreateRecords(c.Request.Context(), reqs)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{Data: records})
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
	var req record.BulkUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	err := h.recordService.BulkUpdateRecords(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Success: true})
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
	var req record.BulkDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	err := h.recordService.BulkDeleteRecords(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Success: true})
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
	var req record.ComplexQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	results, err := h.recordService.ComplexQuery(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Data: results})
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
	tableID := c.Query("table_id")
	var tableIDPtr *string
	if tableID != "" {
		tableIDPtr = &tableID
	}

	stats, err := h.recordService.GetRecordStats(c.Request.Context(), tableIDPtr)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Data: stats})
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
	var req record.ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	data, err := h.recordService.ExportRecords(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 设置响应头
	filename := fmt.Sprintf("records_export_%d.%s", time.Now().Unix(), req.Format)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/octet-stream")
	c.Data(http.StatusOK, "application/octet-stream", data)
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
	var req record.ImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	count, err := h.recordService.ImportRecords(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Data: count})
}

func (h *RecordHandler) handleError(c *gin.Context, err error) {
	traceID := c.GetString("request_id")

	if appErr, ok := errors.IsAppError(err); ok {
		logger.Error("Application error",
			logger.String("error", appErr.Message),
			logger.String("code", appErr.Code),
			logger.String("trace_id", traceID),
		)

		c.JSON(appErr.HTTPStatus, ErrorResponse{
			Error:   appErr.Message,
			Code:    appErr.Code,
			Details: appErr.Details,
			TraceID: traceID,
		})
		return
	}

	logger.Error("Internal server error",
		logger.ErrorField(err),
		logger.String("trace_id", traceID),
	)

	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error:   "服务器内部错误",
		Code:    "INTERNAL_SERVER_ERROR",
		TraceID: traceID,
	})
}
