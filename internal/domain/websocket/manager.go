package websocket

import (
	"context"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Connection WebSocket连接封装
type Connection struct {
	ID            string
	UserID        string
	SessionID     string
	Conn          *websocket.Conn
	Send          chan *Message
	Manager       *Manager
	Subscriptions map[string]bool // 订阅的频道
	LastPing      time.Time
	mu            sync.RWMutex
}

// Manager WebSocket连接管理器
type Manager struct {
	connections map[string]*Connection // 连接ID -> 连接
	userConns   map[string][]string    // 用户ID -> 连接ID列表
	channels    map[string][]string    // 频道名 -> 连接ID列表
	register    chan *Connection
	unregister  chan *Connection
	broadcast   chan *BroadcastMessage
	mu          sync.RWMutex
	logger      *zap.Logger
}

// BroadcastMessage 广播消息
type BroadcastMessage struct {
	Channel string
	Message *Message
	Exclude []string // 排除的连接ID
}

// NewManager 创建新的连接管理器
func NewManager(logger *zap.Logger) *Manager {
	return &Manager{
		connections: make(map[string]*Connection),
		userConns:   make(map[string][]string),
		channels:    make(map[string][]string),
		register:    make(chan *Connection),
		unregister:  make(chan *Connection),
		broadcast:   make(chan *BroadcastMessage),
		logger:      logger,
	}
}

// Run 启动管理器
func (m *Manager) Run(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // 心跳检查间隔
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case conn := <-m.register:
			m.registerConnection(conn)
		case conn := <-m.unregister:
			m.unregisterConnection(conn)
		case broadcast := <-m.broadcast:
			m.broadcastToChannel(broadcast)
		case <-ticker.C:
			m.checkHeartbeat()
		}
	}
}

// registerConnection 注册连接
func (m *Manager) registerConnection(conn *Connection) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.connections[conn.ID] = conn
	m.userConns[conn.UserID] = append(m.userConns[conn.UserID], conn.ID)

	m.logger.Info("WebSocket connection registered",
		zap.String("connection_id", conn.ID),
		zap.String("user_id", conn.UserID),
		zap.String("session_id", conn.SessionID),
	)
}

// unregisterConnection 注销连接
func (m *Manager) unregisterConnection(conn *Connection) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 从连接映射中删除
	delete(m.connections, conn.ID)

	// 从用户连接映射中删除
	if conns, exists := m.userConns[conn.UserID]; exists {
		for i, connID := range conns {
			if connID == conn.ID {
				m.userConns[conn.UserID] = append(conns[:i], conns[i+1:]...)
				break
			}
		}
		if len(m.userConns[conn.UserID]) == 0 {
			delete(m.userConns, conn.UserID)
		}
	}

	// 从所有频道中删除
	for channel := range conn.Subscriptions {
		m.removeFromChannel(channel, conn.ID)
	}

	// 关闭连接
	close(conn.Send)

	m.logger.Info("WebSocket connection unregistered",
		zap.String("connection_id", conn.ID),
		zap.String("user_id", conn.UserID),
	)
}

// broadcastToChannel 向频道广播消息
func (m *Manager) broadcastToChannel(broadcast *BroadcastMessage) {
	m.mu.RLock()
	connIDs, exists := m.channels[broadcast.Channel]
	m.mu.RUnlock()

	if !exists {
		return
	}

	excludeMap := make(map[string]bool)
	for _, id := range broadcast.Exclude {
		excludeMap[id] = true
	}

	for _, connID := range connIDs {
		if excludeMap[connID] {
			continue
		}

		m.mu.RLock()
		conn, exists := m.connections[connID]
		m.mu.RUnlock()

		if exists {
			select {
			case conn.Send <- broadcast.Message:
			default:
				// 发送失败，关闭连接
				m.unregister <- conn
			}
		}
	}
}

// Subscribe 订阅频道
func (m *Manager) Subscribe(connID, channel string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn, exists := m.connections[connID]
	if !exists {
		return
	}

	conn.mu.Lock()
	conn.Subscriptions[channel] = true
	conn.mu.Unlock()

	m.channels[channel] = append(m.channels[channel], connID)

	m.logger.Info("Connection subscribed to channel",
		zap.String("connection_id", connID),
		zap.String("channel", channel),
	)
}

// Unsubscribe 取消订阅频道
func (m *Manager) Unsubscribe(connID, channel string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn, exists := m.connections[connID]
	if !exists {
		return
	}

	conn.mu.Lock()
	delete(conn.Subscriptions, channel)
	conn.mu.Unlock()

	m.removeFromChannel(channel, connID)

	m.logger.Info("Connection unsubscribed from channel",
		zap.String("connection_id", connID),
		zap.String("channel", channel),
	)
}

// removeFromChannel 从频道中移除连接
func (m *Manager) removeFromChannel(channel, connID string) {
	if connIDs, exists := m.channels[channel]; exists {
		for i, id := range connIDs {
			if id == connID {
				m.channels[channel] = append(connIDs[:i], connIDs[i+1:]...)
				break
			}
		}
		if len(m.channels[channel]) == 0 {
			delete(m.channels, channel)
		}
	}
}

// checkHeartbeat 检查心跳
func (m *Manager) checkHeartbeat() {
	m.mu.RLock()
	connections := make([]*Connection, 0, len(m.connections))
	for _, conn := range m.connections {
		connections = append(connections, conn)
	}
	m.mu.RUnlock()

	now := time.Now()
	for _, conn := range connections {
		conn.mu.RLock()
		lastPing := conn.LastPing
		conn.mu.RUnlock()

		if now.Sub(lastPing) > 60*time.Second { // 60秒超时
			m.logger.Info("Connection heartbeat timeout, closing",
				zap.String("connection_id", conn.ID),
				zap.String("user_id", conn.UserID),
			)
			m.unregister <- conn
		}
	}
}

// GetConnection 获取连接
func (m *Manager) GetConnection(connID string) (*Connection, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	conn, exists := m.connections[connID]
	return conn, exists
}

// GetUserConnections 获取用户的所有连接
func (m *Manager) GetUserConnections(userID string) []*Connection {
	m.mu.RLock()
	defer m.mu.RUnlock()

	connIDs, exists := m.userConns[userID]
	if !exists {
		return nil
	}

	connections := make([]*Connection, 0, len(connIDs))
	for _, connID := range connIDs {
		if conn, exists := m.connections[connID]; exists {
			connections = append(connections, conn)
		}
	}

	return connections
}

// BroadcastToChannel 向频道广播消息
func (m *Manager) BroadcastToChannel(channel string, message *Message, exclude ...string) {
	m.broadcast <- &BroadcastMessage{
		Channel: channel,
		Message: message,
		Exclude: exclude,
	}
}

// BroadcastToUser 向用户的所有连接广播消息
func (m *Manager) BroadcastToUser(userID string, message *Message) {
	connections := m.GetUserConnections(userID)
	for _, conn := range connections {
		select {
		case conn.Send <- message:
		default:
			// 发送失败，关闭连接
			m.unregister <- conn
		}
	}
}

// GetStats 获取统计信息
func (m *Manager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"total_connections": len(m.connections),
		"total_users":       len(m.userConns),
		"total_channels":    len(m.channels),
	}
}
