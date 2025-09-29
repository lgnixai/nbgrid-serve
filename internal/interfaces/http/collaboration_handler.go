package http

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"teable-go-backend/internal/domain/websocket"
	"teable-go-backend/pkg/errors"
	"teable-go-backend/pkg/response"
)

// CollaborationHandler 协作功能HTTP处理器
type CollaborationHandler struct {
	collaborationService *websocket.CollaborationService
	logger               *zap.Logger
}

// NewCollaborationHandler 创建协作功能HTTP处理器
func NewCollaborationHandler(collaborationService *websocket.CollaborationService, logger *zap.Logger) *CollaborationHandler {
	return &CollaborationHandler{
		collaborationService: collaborationService,
		logger:               logger,
	}
}

// UpdatePresenceRequest 更新在线状态请求
type UpdatePresenceRequest struct {
	Collection string                 `json:"collection" binding:"required"`
	Data       map[string]interface{} `json:"data"`
}

// UpdatePresence 更新用户在线状态
// @Summary 更新用户在线状态
// @Description 更新用户在指定集合中的在线状态
// @Tags 协作
// @Accept json
// @Produce json
// @Param request body UpdatePresenceRequest true "更新在线状态请求"
// @Success 200 {object} Response{success=boolean}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/collaboration/presence [post]
func (h *CollaborationHandler) UpdatePresence(c *gin.Context) {
	var req UpdatePresenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	// 从JWT中获取用户ID和会话ID
	userID, exists := c.Get("user_id")
	if !exists {
		h.handleError(c, errors.ErrUnauthorized.WithDetails("User ID not found"))
		return
	}

	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = "default"
	}

	// 更新在线状态
	if err := h.collaborationService.UpdateUserPresence(c.Request.Context(), userID.(string), sessionID, req.Collection, req.Data); err != nil {
		h.handleError(c, errors.ErrInternalServer.WithDetails(err.Error()))
		return
	}

	response.SuccessWithMessage(c, map[string]bool{"success": true}, "")
}

// RemovePresence 移除用户在线状态
// @Summary 移除用户在线状态
// @Description 移除用户在指定集合中的在线状态
// @Tags 协作
// @Produce json
// @Param collection query string true "集合名称"
// @Success 200 {object} Response{success=boolean}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/collaboration/presence [delete]
func (h *CollaborationHandler) RemovePresence(c *gin.Context) {
	collection := c.Query("collection")
	if collection == "" {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails("Collection is required"))
		return
	}

	// 从JWT中获取用户ID和会话ID
	userID, exists := c.Get("user_id")
	if !exists {
		h.handleError(c, errors.ErrUnauthorized.WithDetails("User ID not found"))
		return
	}

	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = "default"
	}

	// 移除在线状态
	if err := h.collaborationService.RemoveUserPresence(c.Request.Context(), userID.(string), sessionID, collection); err != nil {
		h.handleError(c, errors.ErrInternalServer.WithDetails(err.Error()))
		return
	}

	response.SuccessWithMessage(c, map[string]bool{"success": true}, "")
}

// GetPresence 获取在线状态
// @Summary 获取在线状态
// @Description 获取指定集合中的在线用户状态
// @Tags 协作
// @Produce json
// @Param collection query string true "集合名称"
// @Success 200 {object} Response{data=[]websocket.PresenceInfo}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/collaboration/presence [get]
func (h *CollaborationHandler) GetPresence(c *gin.Context) {
	collection := c.Query("collection")
	if collection == "" {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails("Collection is required"))
		return
	}

	// 获取在线状态
	presence := h.collaborationService.GetPresenceInfo(collection)

	response.SuccessWithMessage(c, presence, "")
}

// UpdateCursorRequest 更新光标位置请求
type UpdateCursorRequest struct {
	Collection string                 `json:"collection" binding:"required"`
	Document   string                 `json:"document" binding:"required"`
	Position   map[string]interface{} `json:"position" binding:"required"`
	Selection  map[string]interface{} `json:"selection,omitempty"`
}

// UpdateCursor 更新用户光标位置
// @Summary 更新用户光标位置
// @Description 更新用户在指定文档中的光标位置
// @Tags 协作
// @Accept json
// @Produce json
// @Param request body UpdateCursorRequest true "更新光标位置请求"
// @Success 200 {object} Response{success=boolean}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/collaboration/cursor [post]
func (h *CollaborationHandler) UpdateCursor(c *gin.Context) {
	var req UpdateCursorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	// 从JWT中获取用户ID和会话ID
	userID, exists := c.Get("user_id")
	if !exists {
		h.handleError(c, errors.ErrUnauthorized.WithDetails("User ID not found"))
		return
	}

	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = "default"
	}

	// 创建光标信息
	cursor := &websocket.CursorInfo{
		UserID:    userID.(string),
		SessionID: sessionID,
		Position:  req.Position,
		Selection: req.Selection,
		Timestamp: time.Now(),
	}

	// 更新光标位置
	if err := h.collaborationService.UpdateUserCursor(c.Request.Context(), userID.(string), sessionID, req.Collection, req.Document, cursor); err != nil {
		h.handleError(c, errors.ErrInternalServer.WithDetails(err.Error()))
		return
	}

	response.SuccessWithMessage(c, map[string]bool{"success": true}, "")
}

// RemoveCursor 移除用户光标
// @Summary 移除用户光标
// @Description 移除用户在指定文档中的光标
// @Tags 协作
// @Produce json
// @Param collection query string true "集合名称"
// @Param document query string true "文档ID"
// @Success 200 {object} Response{success=boolean}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/collaboration/cursor [delete]
func (h *CollaborationHandler) RemoveCursor(c *gin.Context) {
	collection := c.Query("collection")
	document := c.Query("document")
	if collection == "" || document == "" {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails("Collection and document are required"))
		return
	}

	// 从JWT中获取用户ID和会话ID
	userID, exists := c.Get("user_id")
	if !exists {
		h.handleError(c, errors.ErrUnauthorized.WithDetails("User ID not found"))
		return
	}

	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = "default"
	}

	// 移除光标
	if err := h.collaborationService.RemoveUserCursor(c.Request.Context(), userID.(string), sessionID, collection, document); err != nil {
		h.handleError(c, errors.ErrInternalServer.WithDetails(err.Error()))
		return
	}

	response.SuccessWithMessage(c, map[string]bool{"success": true}, "")
}

// GetCursors 获取光标信息
// @Summary 获取光标信息
// @Description 获取指定集合中的所有用户光标信息
// @Tags 协作
// @Produce json
// @Param collection query string true "集合名称"
// @Success 200 {object} Response{data=[]websocket.CursorInfo}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/collaboration/cursor [get]
func (h *CollaborationHandler) GetCursors(c *gin.Context) {
	collection := c.Query("collection")
	if collection == "" {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails("Collection is required"))
		return
	}

	// 获取光标信息
	cursors := h.collaborationService.GetCursorInfo(collection)

	response.SuccessWithMessage(c, cursors, "")
}

// SendNotificationRequest 发送通知请求
type SendNotificationRequest struct {
	Collection string                 `json:"collection" binding:"required"`
	UserID     string                 `json:"user_id" binding:"required"`
	Type       string                 `json:"type" binding:"required"`
	Data       map[string]interface{} `json:"data"`
}

// SendNotification 发送协作通知
// @Summary 发送协作通知
// @Description 向指定用户发送协作通知
// @Tags 协作
// @Accept json
// @Produce json
// @Param request body SendNotificationRequest true "发送通知请求"
// @Success 200 {object} Response{success=boolean}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/collaboration/notification [post]
func (h *CollaborationHandler) SendNotification(c *gin.Context) {
	var req SendNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, errors.ErrInvalidRequest.WithDetails(err.Error()))
		return
	}

	// 发送通知
	if err := h.collaborationService.PublishCollaborationNotification(req.Collection, req.UserID, req.Type, req.Data); err != nil {
		h.handleError(c, errors.ErrInternalServer.WithDetails(err.Error()))
		return
	}

	response.SuccessWithMessage(c, map[string]bool{"success": true}, "")
}

// GetCollaborationStats 获取协作统计信息
// @Summary 获取协作统计信息
// @Description 获取实时协作的统计信息
// @Tags 协作
// @Produce json
// @Success 200 {object} Response{data=object}
// @Failure 500 {object} ErrorResponse
// @Router /api/collaboration/stats [get]
func (h *CollaborationHandler) GetCollaborationStats(c *gin.Context) {
	// 获取统计信息
	stats := map[string]interface{}{
		"active_connections": 0, // TODO: 从WebSocket管理器获取
		"active_presence":    0, // TODO: 从协作服务获取
		"active_cursors":     0, // TODO: 从协作服务获取
		"timestamp":          time.Now().Format(time.RFC3339),
	}

	response.SuccessWithMessage(c, stats, "")
}

func (h *CollaborationHandler) handleError(c *gin.Context, err error) {
	response.Error(c, err)
}
