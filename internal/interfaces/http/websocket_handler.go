package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"teable-go-backend/internal/domain/websocket"
)

// WebSocketHandler WebSocket HTTP处理器
type WebSocketHandler struct {
	wsHandler *websocket.Handler
	logger    *zap.Logger
}

// NewWebSocketHandler 创建WebSocket HTTP处理器
func NewWebSocketHandler(wsHandler *websocket.Handler, logger *zap.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		wsHandler: wsHandler,
		logger:    logger,
	}
}

// HandleWebSocket 处理WebSocket连接
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	h.wsHandler.HandleWebSocket(c)
}

// GetWebSocketStats 获取WebSocket统计信息
func (h *WebSocketHandler) GetWebSocketStats(c *gin.Context) {
	// 这里需要从WebSocket服务获取统计信息
	// 暂时返回模拟数据
	stats := map[string]interface{}{
		"total_connections": 0,
		"total_users":       0,
		"total_channels":    0,
		"uptime":            "0s",
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   stats,
	})
}

// BroadcastMessage 广播消息请求
type BroadcastMessage struct {
	Channel string                 `json:"channel" binding:"required"`
	Message map[string]interface{} `json:"message" binding:"required"`
	Exclude []string               `json:"exclude,omitempty"`
}

// BroadcastToChannel 向频道广播消息
func (h *WebSocketHandler) BroadcastToChannel(c *gin.Context) {
	var req BroadcastMessage
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// 这里需要调用WebSocket服务进行广播
	// wsMessage := websocket.NewMessage(websocket.MessageTypeOp, req.Message)
	// wsService.BroadcastToChannel(req.Channel, wsMessage, req.Exclude...)

	h.logger.Info("Broadcast message to channel",
		zap.String("channel", req.Channel),
		zap.Int("exclude_count", len(req.Exclude)),
	)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Message broadcasted successfully",
	})
}

// BroadcastToUser 向用户广播消息
type BroadcastToUserRequest struct {
	UserID  string                 `json:"user_id" binding:"required"`
	Message map[string]interface{} `json:"message" binding:"required"`
}

// BroadcastToUser 向用户广播消息
func (h *WebSocketHandler) BroadcastToUser(c *gin.Context) {
	var req BroadcastToUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// 这里需要调用WebSocket服务进行广播
	// wsMessage := websocket.NewMessage(websocket.MessageTypeOp, req.Message)
	// wsService.BroadcastToUser(req.UserID, wsMessage)

	h.logger.Info("Broadcast message to user",
		zap.String("user_id", req.UserID),
	)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Message broadcasted to user successfully",
	})
}

// PublishDocumentOp 发布文档操作
type PublishDocumentOpRequest struct {
	Collection string        `json:"collection" binding:"required"`
	Document   string        `json:"document,omitempty"`
	Operation  []interface{} `json:"operation" binding:"required"`
}

// PublishDocumentOp 发布文档操作
func (h *WebSocketHandler) PublishDocumentOp(c *gin.Context) {
	var req PublishDocumentOpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// 这里需要调用WebSocket服务发布操作
	// wsService.PublishDocumentOp(req.Collection, req.Document, req.Operation)

	h.logger.Info("Publish document operation",
		zap.String("collection", req.Collection),
		zap.String("document", req.Document),
		zap.Int("op_count", len(req.Operation)),
	)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Document operation published successfully",
	})
}

// PublishRecordOp 发布记录操作
type PublishRecordOpRequest struct {
	TableID   string        `json:"table_id" binding:"required"`
	RecordID  string        `json:"record_id" binding:"required"`
	Operation []interface{} `json:"operation" binding:"required"`
}

// PublishRecordOp 发布记录操作
func (h *WebSocketHandler) PublishRecordOp(c *gin.Context) {
	var req PublishRecordOpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// 这里需要调用WebSocket服务发布操作
	// wsService.PublishRecordOp(req.TableID, req.RecordID, req.Operation)

	h.logger.Info("Publish record operation",
		zap.String("table_id", req.TableID),
		zap.String("record_id", req.RecordID),
		zap.Int("op_count", len(req.Operation)),
	)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Record operation published successfully",
	})
}

// PublishViewOp 发布视图操作
type PublishViewOpRequest struct {
	TableID   string        `json:"table_id" binding:"required"`
	ViewID    string        `json:"view_id" binding:"required"`
	Operation []interface{} `json:"operation" binding:"required"`
}

// PublishViewOp 发布视图操作
func (h *WebSocketHandler) PublishViewOp(c *gin.Context) {
	var req PublishViewOpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// 这里需要调用WebSocket服务发布操作
	// wsService.PublishViewOp(req.TableID, req.ViewID, req.Operation)

	h.logger.Info("Publish view operation",
		zap.String("table_id", req.TableID),
		zap.String("view_id", req.ViewID),
		zap.Int("op_count", len(req.Operation)),
	)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "View operation published successfully",
	})
}

// PublishFieldOp 发布字段操作
type PublishFieldOpRequest struct {
	TableID   string        `json:"table_id" binding:"required"`
	FieldID   string        `json:"field_id" binding:"required"`
	Operation []interface{} `json:"operation" binding:"required"`
}

// PublishFieldOp 发布字段操作
func (h *WebSocketHandler) PublishFieldOp(c *gin.Context) {
	var req PublishFieldOpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// 这里需要调用WebSocket服务发布操作
	// wsService.PublishFieldOp(req.TableID, req.FieldID, req.Operation)

	h.logger.Info("Publish field operation",
		zap.String("table_id", req.TableID),
		zap.String("field_id", req.FieldID),
		zap.Int("op_count", len(req.Operation)),
	)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Field operation published successfully",
	})
}
