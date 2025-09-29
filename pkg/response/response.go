package response

import (
	"net/http"
	"time"

	"teable-go-backend/pkg/errors"

	"github.com/gin-gonic/gin"
)

// Pagination 分页信息
type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// --- 新版统一响应结构（数字码） ---

// APIResponse 统一响应结构（V2）
type APIResponse struct {
	Code       int           `json:"code"`
	Message    string        `json:"message,omitempty"`
	Data       interface{}   `json:"data"`
	Error      *ErrorPayload `json:"error,omitempty"`
	RequestID  string        `json:"request_id,omitempty"`
	Timestamp  string        `json:"timestamp,omitempty"`
	DurationMs int64         `json:"duration_ms,omitempty"`
	Version    string        `json:"version,omitempty"`
}

// ErrorPayload 错误详情载荷（V2）
type ErrorPayload struct {
	Details interface{} `json:"details,omitempty"`
}

// metaFromContext 提取元信息
func metaFromContext(c *gin.Context) (requestID, ts string, durationMs int64) {
	requestID = c.GetString("request_id")
	ts = time.Now().UTC().Format(time.RFC3339)
	if v, exists := c.Get("start_time"); exists {
		if t, ok := v.(time.Time); ok {
			durationMs = time.Since(t).Milliseconds()
		}
	}
	return
}

// Success 成功响应（数字码）
func Success(c *gin.Context, data interface{}, message string) {
	reqID, ts, dur := metaFromContext(c)
	c.JSON(http.StatusOK, APIResponse{
		Code:       errors.CodeOK,
		Message:    message,
		Data:       data,
		RequestID:  reqID,
		Timestamp:  ts,
		DurationMs: dur,
	})
}

// PaginatedSuccess 分页成功响应（data: list+pagination）
func PaginatedSuccess(c *gin.Context, list interface{}, pagination Pagination, message string) {
	reqID, ts, dur := metaFromContext(c)
	c.JSON(http.StatusOK, APIResponse{
		Code:    errors.CodeOK,
		Message: message,
		Data: gin.H{
			"list":       list,
			"pagination": pagination,
		},
		RequestID:  reqID,
		Timestamp:  ts,
		DurationMs: dur,
	})
}

// Error 错误响应（数字码）
func Error(c *gin.Context, err error) {
	reqID, ts, dur := metaFromContext(c)

	// 默认内部错误
	httpStatus := http.StatusInternalServerError
	code := errors.CodeInternalError
	message := "服务器内部错误"
	var details interface{}

	if appErr, ok := err.(*errors.AppError); ok {
		httpStatus = appErr.HTTPStatus
		code = errors.NumericCodeFromString(appErr.Code, appErr.HTTPStatus)
		message = appErr.Message
		details = appErr.Details
	}

	c.JSON(httpStatus, APIResponse{
		Code:    code,
		Message: message,
		Data:    nil,
		Error: &ErrorPayload{
			Details: details,
		},
		RequestID:  reqID,
		Timestamp:  ts,
		DurationMs: dur,
	})
}

// SuccessWithMessage 便捷函数：成功并带 message
func SuccessWithMessage(c *gin.Context, data interface{}, message string) {
	Success(c, data, message)
}
