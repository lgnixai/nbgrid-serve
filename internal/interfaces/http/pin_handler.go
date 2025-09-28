package http

import (
	"net/http"

	"teable-go-backend/internal/interfaces/middleware"
	"teable-go-backend/pkg/errors"
	"teable-go-backend/pkg/logger"

	"github.com/gin-gonic/gin"
)

// PinHandler Pin HTTP处理器
type PinHandler struct {
	// 这里可以添加 pin 相关的服务
}

// NewPinHandler 创建 Pin 处理器
func NewPinHandler() *PinHandler {
	return &PinHandler{}
}

// PinItem Pin 项目结构
type PinItem struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Order int    `json:"order"`
}

// ListPins 获取 Pin 列表
// @Summary 获取 Pin 列表
// @Description 获取用户的 Pin 列表
// @Tags Pin
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse{data=[]PinItem}
// @Failure 401 {object} ErrorResponse
// @Router /api/pin/list [get]
func (h *PinHandler) ListPins(c *gin.Context) {
	// 获取当前用户ID
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	logger.Info("Getting pin list for user", logger.String("user_id", userID))

	// 暂时返回空数组，避免前端错误
	// TODO: 实现真正的 pin 数据获取逻辑
	pins := []PinItem{}

	c.JSON(http.StatusOK, SuccessResponse{
		Data: pins,
	})
}

// handleError 统一错误处理
func (h *PinHandler) handleError(c *gin.Context, err error) {
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
