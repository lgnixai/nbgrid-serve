package response

import (
	"net/http"

	"teable-go-backend/pkg/errors"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
	Code    int         `json:"code"`
}

// PaginatedResponse 分页响应结构
type PaginatedResponse struct {
	Response
	Pagination Pagination `json:"pagination"`
}

// Pagination 分页信息
type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
		Code:    http.StatusOK,
	})
}

// SuccessWithMessage 带消息的成功响应
func SuccessWithMessage(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
		Message: message,
		Code:    http.StatusOK,
	})
}

// PaginatedSuccess 分页成功响应
func PaginatedSuccess(c *gin.Context, data interface{}, pagination Pagination) {
	c.JSON(http.StatusOK, PaginatedResponse{
		Response: Response{
			Success: true,
			Data:    data,
			Code:    http.StatusOK,
		},
		Pagination: pagination,
	})
}

// Error 错误响应
func Error(c *gin.Context, err error) {
	var appErr *errors.AppError
	var ok bool

	if appErr, ok = err.(*errors.AppError); !ok {
		appErr = errors.ErrInternalServer
	}

	c.JSON(appErr.HTTPStatus, Response{
		Success: false,
		Error:   appErr.Message,
		Code:    appErr.HTTPStatus,
	})
}

// BadRequest 400错误响应
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Error:   message,
		Code:    http.StatusBadRequest,
	})
}

// Unauthorized 401错误响应
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Success: false,
		Error:   message,
		Code:    http.StatusUnauthorized,
	})
}

// Forbidden 403错误响应
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Success: false,
		Error:   message,
		Code:    http.StatusForbidden,
	})
}

// NotFound 404错误响应
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{
		Success: false,
		Error:   message,
		Code:    http.StatusNotFound,
	})
}

// InternalServerError 500错误响应
func InternalServerError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Error:   message,
		Code:    http.StatusInternalServerError,
	})
}

// ValidationError 验证错误响应
func ValidationError(c *gin.Context, message string) {
	c.JSON(http.StatusUnprocessableEntity, Response{
		Success: false,
		Error:   message,
		Code:    http.StatusUnprocessableEntity,
	})
}
