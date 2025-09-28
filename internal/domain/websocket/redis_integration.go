package websocket

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"teable-go-backend/pkg/logger"

	"go.uber.org/zap"

	"teable-go-backend/internal/infrastructure/pubsub"
)

// RedisIntegration Redis集成服务
type RedisIntegration struct {
	pubsub  *pubsub.RedisPubSub
	manager *Manager
	logger  *zap.Logger
	prefix  string
}

// NewRedisIntegration 创建Redis集成服务
func NewRedisIntegration(pubsub *pubsub.RedisPubSub, manager *Manager, logger *zap.Logger, prefix string) *RedisIntegration {
	integration := &RedisIntegration{
		pubsub:  pubsub,
		manager: manager,
		logger:  logger,
		prefix:  prefix,
	}

	// 订阅所有WebSocket相关频道
	integration.subscribeToChannels()

	return integration
}

// subscribeToChannels 订阅所有相关频道
func (r *RedisIntegration) subscribeToChannels() {
	// 订阅WebSocket广播频道
	wsBroadcastChannel := r.buildChannel("ws", "broadcast")
	r.pubsub.Subscribe(wsBroadcastChannel, func(channel string, data interface{}) {
		r.handleWebSocketBroadcast(channel, data)
	})

	// 订阅文档操作频道
	docOpChannel := r.buildChannel("doc", "op")
	r.pubsub.Subscribe(docOpChannel, func(channel string, data interface{}) {
		r.handleDocumentOperation(channel, data)
	})

	// 订阅记录操作频道
	recordOpChannel := r.buildChannel("record", "op")
	r.pubsub.Subscribe(recordOpChannel, func(channel string, data interface{}) {
		r.handleRecordOperation(channel, data)
	})

	// 订阅视图操作频道
	viewOpChannel := r.buildChannel("view", "op")
	r.pubsub.Subscribe(viewOpChannel, func(channel string, data interface{}) {
		r.handleViewOperation(channel, data)
	})

	// 订阅字段操作频道
	fieldOpChannel := r.buildChannel("field", "op")
	r.pubsub.Subscribe(fieldOpChannel, func(channel string, data interface{}) {
		r.handleFieldOperation(channel, data)
	})

	// 订阅在线状态频道
	presenceChannel := r.buildChannel("presence", "update")
	r.pubsub.Subscribe(presenceChannel, func(channel string, data interface{}) {
		r.handlePresenceUpdate(channel, data)
	})

	// 订阅系统消息频道
	systemChannel := r.buildChannel("system", "message")
	r.pubsub.Subscribe(systemChannel, func(channel string, data interface{}) {
		r.handleSystemMessage(channel, data)
	})

	r.logger.Info("Redis integration subscribed to channels")
}

// buildChannel 构建频道名称
func (r *RedisIntegration) buildChannel(parts ...string) string {
	channel := strings.Join(parts, ":")
	if r.prefix != "" {
		channel = r.prefix + ":" + channel
	}
	return channel
}

// handleWebSocketBroadcast 处理WebSocket广播消息
func (r *RedisIntegration) handleWebSocketBroadcast(channel string, data interface{}) error {
	var broadcast BroadcastMessage
	if err := r.unmarshalData(data, &broadcast); err != nil {
		return fmt.Errorf("failed to unmarshal broadcast message: %w", err)
	}

	// 创建WebSocket消息
	wsMessage := NewMessage(MessageTypeOp, broadcast.Message)

	// 广播到本地连接
	r.manager.BroadcastToChannel(broadcast.Channel, wsMessage, broadcast.Exclude...)

	r.logger.Debug("Handled WebSocket broadcast from Redis",
		logger.String("channel", broadcast.Channel),
		logger.Int("exclude_count", len(broadcast.Exclude)),
	)

	return nil
}

// handleDocumentOperation 处理文档操作
func (r *RedisIntegration) handleDocumentOperation(channel string, data interface{}) error {
	var op DocumentOperationMessage
	if err := r.unmarshalData(data, &op); err != nil {
		return fmt.Errorf("failed to unmarshal document operation: %w", err)
	}

	// 创建WebSocket消息
	wsMessage := NewMessage(MessageTypeOp, DocumentOperation{
		Op:     op.Operation,
		Source: op.Source,
	})
	wsMessage.Collection = op.Collection
	wsMessage.Document = op.Document

	// 广播到相关频道
	channels := []string{op.Collection}
	if op.Document != "" {
		channels = append(channels, fmt.Sprintf("%s.%s", op.Collection, op.Document))
	}

	for _, ch := range channels {
		r.manager.BroadcastToChannel(ch, wsMessage)
	}

	r.logger.Debug("Handled document operation from Redis",
		logger.String("collection", op.Collection),
		logger.String("document", op.Document),
		logger.Int("op_count", len(op.Operation)),
	)

	return nil
}

// handleRecordOperation 处理记录操作
func (r *RedisIntegration) handleRecordOperation(channel string, data interface{}) error {
	var op RecordOperationMessage
	if err := r.unmarshalData(data, &op); err != nil {
		return fmt.Errorf("failed to unmarshal record operation: %w", err)
	}

	collection := fmt.Sprintf("record_%s", op.TableID)
	wsMessage := NewMessage(MessageTypeOp, DocumentOperation{
		Op:     op.Operation,
		Source: op.Source,
	})
	wsMessage.Collection = collection
	wsMessage.Document = op.RecordID

	// 广播到记录频道
	channels := []string{collection}
	if op.RecordID != "" {
		channels = append(channels, fmt.Sprintf("%s.%s", collection, op.RecordID))
	}

	for _, ch := range channels {
		r.manager.BroadcastToChannel(ch, wsMessage)
	}

	r.logger.Debug("Handled record operation from Redis",
		logger.String("table_id", op.TableID),
		logger.String("record_id", op.RecordID),
		logger.Int("op_count", len(op.Operation)),
	)

	return nil
}

// handleViewOperation 处理视图操作
func (r *RedisIntegration) handleViewOperation(channel string, data interface{}) error {
	var op ViewOperationMessage
	if err := r.unmarshalData(data, &op); err != nil {
		return fmt.Errorf("failed to unmarshal view operation: %w", err)
	}

	collection := fmt.Sprintf("view_%s", op.TableID)
	wsMessage := NewMessage(MessageTypeOp, DocumentOperation{
		Op:     op.Operation,
		Source: op.Source,
	})
	wsMessage.Collection = collection
	wsMessage.Document = op.ViewID

	// 广播到视图频道
	channels := []string{collection}
	if op.ViewID != "" {
		channels = append(channels, fmt.Sprintf("%s.%s", collection, op.ViewID))
	}

	for _, ch := range channels {
		r.manager.BroadcastToChannel(ch, wsMessage)
	}

	r.logger.Debug("Handled view operation from Redis",
		logger.String("table_id", op.TableID),
		logger.String("view_id", op.ViewID),
		logger.Int("op_count", len(op.Operation)),
	)

	return nil
}

// handleFieldOperation 处理字段操作
func (r *RedisIntegration) handleFieldOperation(channel string, data interface{}) error {
	var op FieldOperationMessage
	if err := r.unmarshalData(data, &op); err != nil {
		return fmt.Errorf("failed to unmarshal field operation: %w", err)
	}

	collection := fmt.Sprintf("field_%s", op.TableID)
	wsMessage := NewMessage(MessageTypeOp, DocumentOperation{
		Op:     op.Operation,
		Source: op.Source,
	})
	wsMessage.Collection = collection
	wsMessage.Document = op.FieldID

	// 广播到字段频道
	channels := []string{collection}
	if op.FieldID != "" {
		channels = append(channels, fmt.Sprintf("%s.%s", collection, op.FieldID))
	}

	for _, ch := range channels {
		r.manager.BroadcastToChannel(ch, wsMessage)
	}

	r.logger.Debug("Handled field operation from Redis",
		logger.String("table_id", op.TableID),
		logger.String("field_id", op.FieldID),
		logger.Int("op_count", len(op.Operation)),
	)

	return nil
}

// handlePresenceUpdate 处理在线状态更新
func (r *RedisIntegration) handlePresenceUpdate(channel string, data interface{}) error {
	var presence PresenceUpdateMessage
	if err := r.unmarshalData(data, &presence); err != nil {
		return fmt.Errorf("failed to unmarshal presence update: %w", err)
	}

	// 创建在线状态消息
	wsMessage := NewMessage(MessageTypePresence, PresenceInfo{
		UserID:    presence.UserID,
		SessionID: presence.SessionID,
		Data:      presence.Data,
	})

	// 广播到相关用户
	if presence.UserID != "" {
		r.manager.BroadcastToUser(presence.UserID, wsMessage)
	}

	// 如果指定了频道，也广播到频道
	if presence.Channel != "" {
		r.manager.BroadcastToChannel(presence.Channel, wsMessage)
	}

	r.logger.Debug("Handled presence update from Redis",
		logger.String("user_id", presence.UserID),
		logger.String("session_id", presence.SessionID),
	)

	return nil
}

// handleSystemMessage 处理系统消息
func (r *RedisIntegration) handleSystemMessage(channel string, data interface{}) error {
	var sysMsg SystemMessage
	if err := r.unmarshalData(data, &sysMsg); err != nil {
		return fmt.Errorf("failed to unmarshal system message: %w", err)
	}

	// 创建系统消息
	wsMessage := NewMessage(MessageTypeOp, map[string]interface{}{
		"type":    "system",
		"message": sysMsg.Message,
		"level":   sysMsg.Level,
		"time":    time.Now().Format(time.RFC3339),
	})

	// 广播到所有连接
	r.manager.mu.RLock()
	connections := make([]*Connection, 0, len(r.manager.connections))
	for _, conn := range r.manager.connections {
		connections = append(connections, conn)
	}
	r.manager.mu.RUnlock()

	for _, conn := range connections {
		select {
		case conn.Send <- wsMessage:
		default:
			// 发送失败，关闭连接
			r.manager.unregister <- conn
		}
	}

	r.logger.Debug("Handled system message from Redis",
		logger.String("level", sysMsg.Level),
	)

	return nil
}

// unmarshalData 反序列化数据
func (r *RedisIntegration) unmarshalData(data interface{}, dest interface{}) error {
	// 如果data是字符串，先解析为JSON
	if str, ok := data.(string); ok {
		return json.Unmarshal([]byte(str), dest)
	}

	// 如果data是[]byte，直接解析
	if bytes, ok := data.([]byte); ok {
		return json.Unmarshal(bytes, dest)
	}

	// 否则尝试直接转换
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	return json.Unmarshal(jsonData, dest)
}

// PublishWebSocketBroadcast 发布WebSocket广播消息
func (r *RedisIntegration) PublishWebSocketBroadcast(broadcast BroadcastMessage) error {
	channel := r.buildChannel("ws", "broadcast")
	return r.pubsub.Publish([]string{channel}, broadcast)
}

// PublishDocumentOperation 发布文档操作
func (r *RedisIntegration) PublishDocumentOperation(collection, document string, operation []interface{}, source string) error {
	op := DocumentOperationMessage{
		Collection: collection,
		Document:   document,
		Operation:  operation,
		Source:     source,
		Timestamp:  time.Now(),
	}

	channel := r.buildChannel("doc", "op")
	return r.pubsub.Publish([]string{channel}, op)
}

// PublishRecordOperation 发布记录操作
func (r *RedisIntegration) PublishRecordOperation(tableID, recordID string, operation []interface{}, source string) error {
	op := RecordOperationMessage{
		TableID:   tableID,
		RecordID:  recordID,
		Operation: operation,
		Source:    source,
		Timestamp: time.Now(),
	}

	channel := r.buildChannel("record", "op")
	return r.pubsub.Publish([]string{channel}, op)
}

// PublishViewOperation 发布视图操作
func (r *RedisIntegration) PublishViewOperation(tableID, viewID string, operation []interface{}, source string) error {
	op := ViewOperationMessage{
		TableID:   tableID,
		ViewID:    viewID,
		Operation: operation,
		Source:    source,
		Timestamp: time.Now(),
	}

	channel := r.buildChannel("view", "op")
	return r.pubsub.Publish([]string{channel}, op)
}

// PublishFieldOperation 发布字段操作
func (r *RedisIntegration) PublishFieldOperation(tableID, fieldID string, operation []interface{}, source string) error {
	op := FieldOperationMessage{
		TableID:   tableID,
		FieldID:   fieldID,
		Operation: operation,
		Source:    source,
		Timestamp: time.Now(),
	}

	channel := r.buildChannel("field", "op")
	return r.pubsub.Publish([]string{channel}, op)
}

// PublishPresenceUpdate 发布在线状态更新
func (r *RedisIntegration) PublishPresenceUpdate(userID, sessionID, channel string, data map[string]interface{}) error {
	presence := PresenceUpdateMessage{
		UserID:    userID,
		SessionID: sessionID,
		Channel:   channel,
		Data:      data,
		Timestamp: time.Now(),
	}

	channelName := r.buildChannel("presence", "update")
	return r.pubsub.Publish([]string{channelName}, presence)
}

// PublishSystemMessage 发布系统消息
func (r *RedisIntegration) PublishSystemMessage(message, level string) error {
	sysMsg := SystemMessage{
		Message:   message,
		Level:     level,
		Timestamp: time.Now(),
	}

	channel := r.buildChannel("system", "message")
	return r.pubsub.Publish([]string{channel}, sysMsg)
}

// Close 关闭Redis集成服务
func (r *RedisIntegration) Close() error {
	return r.pubsub.Close()
}

// 消息结构定义

// DocumentOperationMessage 文档操作消息
type DocumentOperationMessage struct {
	Collection string        `json:"collection"`
	Document   string        `json:"document"`
	Operation  []interface{} `json:"operation"`
	Source     string        `json:"source"`
	Timestamp  time.Time     `json:"timestamp"`
}

// RecordOperationMessage 记录操作消息
type RecordOperationMessage struct {
	TableID   string        `json:"table_id"`
	RecordID  string        `json:"record_id"`
	Operation []interface{} `json:"operation"`
	Source    string        `json:"source"`
	Timestamp time.Time     `json:"timestamp"`
}

// ViewOperationMessage 视图操作消息
type ViewOperationMessage struct {
	TableID   string        `json:"table_id"`
	ViewID    string        `json:"view_id"`
	Operation []interface{} `json:"operation"`
	Source    string        `json:"source"`
	Timestamp time.Time     `json:"timestamp"`
}

// FieldOperationMessage 字段操作消息
type FieldOperationMessage struct {
	TableID   string        `json:"table_id"`
	FieldID   string        `json:"field_id"`
	Operation []interface{} `json:"operation"`
	Source    string        `json:"source"`
	Timestamp time.Time     `json:"timestamp"`
}

// PresenceUpdateMessage 在线状态更新消息
type PresenceUpdateMessage struct {
	UserID    string                 `json:"user_id"`
	SessionID string                 `json:"session_id"`
	Channel   string                 `json:"channel,omitempty"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// SystemMessage 系统消息
type SystemMessage struct {
	Message   string    `json:"message"`
	Level     string    `json:"level"`
	Timestamp time.Time `json:"timestamp"`
}
