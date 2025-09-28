package sharedb

import (
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"teable-go-backend/internal/domain/websocket"
	"teable-go-backend/pkg/logger"
)

// WebSocketIntegration WebSocket集成服务
type WebSocketIntegration struct {
	sharedbService *Service
	wsService      websocket.Service
	logger         *zap.Logger
}

// NewWebSocketIntegration 创建WebSocket集成服务
func NewWebSocketIntegration(sharedbService *Service, wsService websocket.Service, logger *zap.Logger) *WebSocketIntegration {
	integration := &WebSocketIntegration{
		sharedbService: sharedbService,
		wsService:      wsService,
		logger:         logger,
	}

	// 注册ShareDB中间件
	sharedbService.Use("submit", integration.handleSubmit)

	return integration
}

// handleSubmit 处理提交操作
func (w *WebSocketIntegration) handleSubmit(context *SubmitContext, next func(error error)) {
	// 记录操作
	w.logger.Debug("Processing ShareDB submit",
		logger.String("collection", context.Collection),
		logger.String("document_id", context.ID),
		logger.String("operation_type", w.getOperationType(context.Op)),
	)

	// 处理操作
	if err := w.processOperation(context); err != nil {
		w.logger.Error("Failed to process operation",
			logger.String("collection", context.Collection),
			logger.String("document_id", context.ID),
			logger.ErrorField(err),
		)
		next(err)
		return
	}

	// 继续执行下一个中间件
	next(nil)
}

// processOperation 处理操作
func (w *WebSocketIntegration) processOperation(context *SubmitContext) error {
	// 根据操作类型处理
	switch w.getOperationType(context.Op) {
	case "create":
		return w.handleCreateOperation(context)
	case "edit":
		return w.handleEditOperation(context)
	case "delete":
		return w.handleDeleteOperation(context)
	default:
		return fmt.Errorf("unknown operation type")
	}
}

// handleCreateOperation 处理创建操作
func (w *WebSocketIntegration) handleCreateOperation(context *SubmitContext) error {
	// 通过WebSocket发布创建操作
	op := []interface{}{
		map[string]interface{}{
			"p":  []interface{}{},
			"oi": context.Op.Create.Data,
		},
	}

	// 根据集合类型发布到相应频道
	docType := w.extractDocType(context.Collection)
	collectionID := w.extractCollectionID(context.Collection)

	switch docType {
	case "record":
		return w.wsService.PublishRecordOp(collectionID, context.ID, op)
	case "field":
		return w.wsService.PublishFieldOp(collectionID, context.ID, op)
	case "view":
		return w.wsService.PublishViewOp(collectionID, context.ID, op)
	default:
		return w.wsService.PublishDocumentOp(context.Collection, context.ID, op)
	}
}

// handleEditOperation 处理编辑操作
func (w *WebSocketIntegration) handleEditOperation(context *SubmitContext) error {
	// 转换OT操作为WebSocket操作
	ops := make([]interface{}, len(context.Op.Op))
	for i, otOp := range context.Op.Op {
		ops[i] = map[string]interface{}{
			"p":  otOp.P,
			"oi": otOp.OI,
			"od": otOp.OD,
		}
	}

	// 根据集合类型发布到相应频道
	docType := w.extractDocType(context.Collection)
	collectionID := w.extractCollectionID(context.Collection)

	switch docType {
	case "record":
		return w.wsService.PublishRecordOp(collectionID, context.ID, ops)
	case "field":
		return w.wsService.PublishFieldOp(collectionID, context.ID, ops)
	case "view":
		return w.wsService.PublishViewOp(collectionID, context.ID, ops)
	default:
		return w.wsService.PublishDocumentOp(context.Collection, context.ID, ops)
	}
}

// handleDeleteOperation 处理删除操作
func (w *WebSocketIntegration) handleDeleteOperation(context *SubmitContext) error {
	// 通过WebSocket发布删除操作
	op := []interface{}{
		map[string]interface{}{
			"p":  []interface{}{},
			"od": nil,
		},
	}

	// 根据集合类型发布到相应频道
	docType := w.extractDocType(context.Collection)
	collectionID := w.extractCollectionID(context.Collection)

	switch docType {
	case "record":
		return w.wsService.PublishRecordOp(collectionID, context.ID, op)
	case "field":
		return w.wsService.PublishFieldOp(collectionID, context.ID, op)
	case "view":
		return w.wsService.PublishViewOp(collectionID, context.ID, op)
	default:
		return w.wsService.PublishDocumentOp(context.Collection, context.ID, op)
	}
}

// ProcessWebSocketMessage 处理WebSocket消息
func (w *WebSocketIntegration) ProcessWebSocketMessage(message *websocket.Message) error {
	// 解析消息
	var submitMsg SubmitMessage
	payloadBytes, err := json.Marshal(message.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal message data: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &submitMsg); err != nil {
		return fmt.Errorf("failed to unmarshal submit message: %w", err)
	}

	// 创建提交上下文
	context := &SubmitContext{
		Agent: &Agent{
			Connection: &Connection{
				ID:        submitMsg.AgentID,
				UserID:    submitMsg.UserID,
				SessionID: submitMsg.SessionID,
				Metadata:  submitMsg.Metadata,
				CreatedAt: time.Now(),
			},
			Custom: submitMsg.Custom,
		},
		Collection: submitMsg.Collection,
		ID:         submitMsg.ID,
		Op:         submitMsg.Op,
		Source:     submitMsg.Source,
	}

	// 处理提交
	var submitError error
	w.sharedbService.OnSubmit(context, func(err error) {
		submitError = err
	})

	return submitError
}

// PublishOperationToWebSocket 发布操作到WebSocket
func (w *WebSocketIntegration) PublishOperationToWebSocket(rawOpMap RawOpMap) error {
	// 转换RawOpMap为WebSocket消息
	for collection, docs := range rawOpMap {
		for docID, rawOp := range docs {
			// 创建WebSocket消息
			wsMessage := &websocket.Message{
				Type:       websocket.MessageTypeOp,
				Collection: collection,
				Document:   docID,
				Data:       rawOp,
			}

			// 发布到WebSocket频道
			channels := []string{collection, fmt.Sprintf("%s.%s", collection, docID)}
			for _, channel := range channels {
				if err := w.wsService.BroadcastToChannel(channel, wsMessage); err != nil {
					w.logger.Error("Failed to broadcast to WebSocket channel",
						logger.String("channel", channel),
						logger.ErrorField(err),
					)
				}
			}
		}
	}

	return nil
}

// 辅助方法

// getOperationType 获取操作类型
func (w *WebSocketIntegration) getOperationType(op *RawOperation) string {
	if op.Create != nil {
		return "create"
	} else if op.Del {
		return "delete"
	} else if len(op.Op) > 0 {
		return "edit"
	}
	return "unknown"
}

// extractDocType 提取文档类型
func (w *WebSocketIntegration) extractDocType(collection string) string {
	parts := w.splitCollection(collection)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// extractCollectionID 提取集合ID
func (w *WebSocketIntegration) extractCollectionID(collection string) string {
	parts := w.splitCollection(collection)
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

// splitCollection 分割集合名称
func (w *WebSocketIntegration) splitCollection(collection string) []string {
	var parts []string
	var current string

	for _, char := range collection {
		if char == '_' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

// SubmitMessage WebSocket提交消息
type SubmitMessage struct {
	AgentID    string                 `json:"agent_id"`
	UserID     string                 `json:"user_id"`
	SessionID  string                 `json:"session_id"`
	Collection string                 `json:"collection"`
	ID         string                 `json:"id"`
	Op         *RawOperation          `json:"op"`
	Source     string                 `json:"source"`
	Metadata   map[string]interface{} `json:"metadata"`
	Custom     map[string]interface{} `json:"custom"`
}
