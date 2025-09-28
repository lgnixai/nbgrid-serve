package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"teable-go-backend/pkg/logger"
)

// CollaborationService 实时协作服务
type CollaborationService struct {
	manager          *Manager
	logger           *zap.Logger
	redisIntegration *RedisIntegration
	presenceStore    *PresenceStore
	cursorStore      *CursorStore
	conflictResolver *ConflictResolver
}

// NewCollaborationService 创建协作服务
func NewCollaborationService(manager *Manager, redisIntegration *RedisIntegration, logger *zap.Logger) *CollaborationService {
	return &CollaborationService{
		manager:          manager,
		logger:           logger,
		redisIntegration: redisIntegration,
		presenceStore:    NewPresenceStore(),
		cursorStore:      NewCursorStore(),
		conflictResolver: NewConflictResolver(),
	}
}

// PresenceStore 在线状态存储
type PresenceStore struct {
	presence map[string]map[string]*PresenceInfo // collection -> userID -> presence
	mu       sync.RWMutex
}

// NewPresenceStore 创建在线状态存储
func NewPresenceStore() *PresenceStore {
	return &PresenceStore{
		presence: make(map[string]map[string]*PresenceInfo),
	}
}

// UpdatePresence 更新在线状态
func (p *PresenceStore) UpdatePresence(collection, userID, sessionID string, data map[string]interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.presence[collection] == nil {
		p.presence[collection] = make(map[string]*PresenceInfo)
	}

	p.presence[collection][userID] = &PresenceInfo{
		UserID:    userID,
		SessionID: sessionID,
		Data:      data,
		LastSeen:  time.Now(),
	}
}

// RemovePresence 移除在线状态
func (p *PresenceStore) RemovePresence(collection, userID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.presence[collection] != nil {
		delete(p.presence[collection], userID)
		if len(p.presence[collection]) == 0 {
			delete(p.presence, collection)
		}
	}
}

// GetPresence 获取在线状态
func (p *PresenceStore) GetPresence(collection string) []PresenceInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var result []PresenceInfo
	if p.presence[collection] != nil {
		for _, presence := range p.presence[collection] {
			result = append(result, *presence)
		}
	}
	return result
}

// CursorStore 光标位置存储
type CursorStore struct {
	cursors map[string]map[string]*CursorInfo // collection -> userID -> cursor
	mu      sync.RWMutex
}

// NewCursorStore 创建光标存储
func NewCursorStore() *CursorStore {
	return &CursorStore{
		cursors: make(map[string]map[string]*CursorInfo),
	}
}

// UpdateCursor 更新光标位置
func (c *CursorStore) UpdateCursor(collection, userID string, cursor *CursorInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cursors[collection] == nil {
		c.cursors[collection] = make(map[string]*CursorInfo)
	}

	c.cursors[collection][userID] = cursor
}

// RemoveCursor 移除光标
func (c *CursorStore) RemoveCursor(collection, userID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cursors[collection] != nil {
		delete(c.cursors[collection], userID)
		if len(c.cursors[collection]) == 0 {
			delete(c.cursors, collection)
		}
	}
}

// GetCursors 获取所有光标
func (c *CursorStore) GetCursors(collection string) []CursorInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []CursorInfo
	if c.cursors[collection] != nil {
		for _, cursor := range c.cursors[collection] {
			result = append(result, *cursor)
		}
	}
	return result
}

// ConflictResolver 冲突解决器
type ConflictResolver struct {
	mu sync.RWMutex
}

// NewConflictResolver 创建冲突解决器
func NewConflictResolver() *ConflictResolver {
	return &ConflictResolver{}
}

// ResolveConflict 解决冲突
func (cr *ConflictResolver) ResolveConflict(operation1, operation2 []interface{}) ([]interface{}, error) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	// 简单的冲突解决策略：操作1优先
	// 在实际应用中，这里应该实现更复杂的OT算法
	return operation1, nil
}

// UpdateUserPresence 更新用户在线状态
func (cs *CollaborationService) UpdateUserPresence(ctx context.Context, userID, sessionID, collection string, data map[string]interface{}) error {
	// 更新本地存储
	cs.presenceStore.UpdatePresence(collection, userID, sessionID, data)

	// 创建在线状态消息
	presence := PresenceInfo{
		UserID:    userID,
		SessionID: sessionID,
		Data:      data,
		LastSeen:  time.Now(),
	}

	message := NewMessage(MessageTypePresence, presence)
	message.Collection = collection

	// 广播到频道
	cs.manager.BroadcastToChannel(collection, message, sessionID)

	// 如果有Redis集成，通过Redis发布
	if cs.redisIntegration != nil {
		if err := cs.redisIntegration.PublishPresenceUpdate(userID, sessionID, collection, data); err != nil {
			cs.logger.Error("Failed to publish presence update to Redis",
				logger.String("collection", collection),
				logger.String("user_id", userID),
				logger.ErrorField(err),
			)
		}
	}

	cs.logger.Debug("User presence updated",
		logger.String("collection", collection),
		logger.String("user_id", userID),
		logger.String("session_id", sessionID),
	)

	return nil
}

// RemoveUserPresence 移除用户在线状态
func (cs *CollaborationService) RemoveUserPresence(ctx context.Context, userID, sessionID, collection string) error {
	// 从本地存储移除
	cs.presenceStore.RemovePresence(collection, userID)

	// 创建离线消息
	presence := PresenceInfo{
		UserID:    userID,
		SessionID: sessionID,
		Data:      map[string]interface{}{"status": "offline"},
		LastSeen:  time.Now(),
	}

	message := NewMessage(MessageTypePresence, presence)
	message.Collection = collection

	// 广播到频道
	cs.manager.BroadcastToChannel(collection, message)

	cs.logger.Debug("User presence removed",
		logger.String("collection", collection),
		logger.String("user_id", userID),
		logger.String("session_id", sessionID),
	)

	return nil
}

// UpdateUserCursor 更新用户光标位置
func (cs *CollaborationService) UpdateUserCursor(ctx context.Context, userID, sessionID, collection, document string, cursor *CursorInfo) error {
	// 更新本地存储
	cs.cursorStore.UpdateCursor(collection, userID, cursor)

	// 创建光标消息
	message := NewMessage(MessageTypeCursor, cursor)
	message.Collection = collection
	message.Document = document

	// 广播到频道
	cs.manager.BroadcastToChannel(collection, message, sessionID)

	cs.logger.Debug("User cursor updated",
		logger.String("collection", collection),
		logger.String("document", document),
		logger.String("user_id", userID),
	)

	return nil
}

// RemoveUserCursor 移除用户光标
func (cs *CollaborationService) RemoveUserCursor(ctx context.Context, userID, sessionID, collection, document string) error {
	// 从本地存储移除
	cs.cursorStore.RemoveCursor(collection, userID)

	// 创建光标移除消息
	cursor := &CursorInfo{
		UserID:    userID,
		SessionID: sessionID,
		Position:  nil,
		Timestamp: time.Now(),
	}

	message := NewMessage(MessageTypeCursor, cursor)
	message.Collection = collection
	message.Document = document

	// 广播到频道
	cs.manager.BroadcastToChannel(collection, message)

	cs.logger.Debug("User cursor removed",
		logger.String("collection", collection),
		logger.String("document", document),
		logger.String("user_id", userID),
	)

	return nil
}

// GetPresenceInfo 获取在线状态信息
func (cs *CollaborationService) GetPresenceInfo(collection string) []PresenceInfo {
	return cs.presenceStore.GetPresence(collection)
}

// GetCursorInfo 获取光标信息
func (cs *CollaborationService) GetCursorInfo(collection string) []CursorInfo {
	return cs.cursorStore.GetCursors(collection)
}

// HandleCollaborationMessage 处理协作消息
func (cs *CollaborationService) HandleCollaborationMessage(message *Message) error {
	switch message.Type {
	case MessageTypePresence:
		return cs.handlePresenceMessage(message)
	case MessageTypeCursor:
		return cs.handleCursorMessage(message)
	case MessageTypeOp:
		return cs.handleOperationMessage(message)
	default:
		return fmt.Errorf("unknown collaboration message type: %s", message.Type)
	}
}

// handlePresenceMessage 处理在线状态消息
func (cs *CollaborationService) handlePresenceMessage(message *Message) error {
	var presence PresenceInfo
	dataBytes, err := json.Marshal(message.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal message data: %w", err)
	}
	if err := json.Unmarshal(dataBytes, &presence); err != nil {
		return fmt.Errorf("failed to unmarshal presence message: %w", err)
	}

	// 更新本地存储
	cs.presenceStore.UpdatePresence(message.Collection, presence.UserID, presence.SessionID, presence.Data)

	return nil
}

// handleCursorMessage 处理光标消息
func (cs *CollaborationService) handleCursorMessage(message *Message) error {
	var cursor CursorInfo
	dataBytes, err := json.Marshal(message.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal message data: %w", err)
	}
	if err := json.Unmarshal(dataBytes, &cursor); err != nil {
		return fmt.Errorf("failed to unmarshal cursor message: %w", err)
	}

	// 更新本地存储
	cs.cursorStore.UpdateCursor(message.Collection, cursor.UserID, &cursor)

	return nil
}

// handleOperationMessage 处理操作消息
func (cs *CollaborationService) handleOperationMessage(message *Message) error {
	// 这里可以添加操作冲突检测和解决逻辑
	// 目前直接转发消息
	return nil
}

// PublishCollaborationNotification 发布协作通知
func (cs *CollaborationService) PublishCollaborationNotification(collection, userID, notificationType string, data map[string]interface{}) error {
	notification := map[string]interface{}{
		"type":      notificationType,
		"data":      data,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	message := NewMessage(MessageTypeNotification, notification)
	message.Collection = collection

	// 向特定用户发送通知
	cs.manager.BroadcastToUser(userID, message)

	cs.logger.Debug("Collaboration notification sent",
		logger.String("collection", collection),
		logger.String("user_id", userID),
		logger.String("notification_type", notificationType),
	)

	return nil
}

// CleanupStalePresence 清理过期的在线状态
func (cs *CollaborationService) CleanupStalePresence() {
	cs.presenceStore.mu.Lock()
	defer cs.presenceStore.mu.Unlock()

	now := time.Now()
	for collection, users := range cs.presenceStore.presence {
		for userID, presence := range users {
			// 如果超过5分钟没有活动，移除在线状态
			if now.Sub(presence.LastSeen) > 5*time.Minute {
				delete(users, userID)
				cs.logger.Debug("Removed stale presence",
					logger.String("collection", collection),
					logger.String("user_id", userID),
				)
			}
		}
		if len(users) == 0 {
			delete(cs.presenceStore.presence, collection)
		}
	}
}

// StartPresenceCleanup 启动在线状态清理任务
func (cs *CollaborationService) StartPresenceCleanup(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cs.CleanupStalePresence()
		}
	}
}
