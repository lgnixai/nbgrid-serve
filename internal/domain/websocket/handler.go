package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	// 写等待时间
	writeWait = 10 * time.Second
	// 读取下一个pong消息的等待时间
	pongWait = 60 * time.Second
	// 发送ping消息的间隔时间
	pingPeriod = (pongWait * 9) / 10
	// 最大消息大小
	maxMessageSize = 512
)

// Handler WebSocket处理器
type Handler struct {
	manager  *Manager
	logger   *zap.Logger
	upgrader websocket.Upgrader
}

// NewHandler 创建新的WebSocket处理器
func NewHandler(manager *Manager, logger *zap.Logger) *Handler {
	return &Handler{
		manager: manager,
		logger:  logger,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// 在生产环境中应该检查Origin
				return true
			},
		},
	}
}

// HandleWebSocket 处理WebSocket连接
func (h *Handler) HandleWebSocket(c *gin.Context) {
	// 从查询参数或头部获取用户信息
	userID := c.Query("user_id")
	if userID == "" {
		userID = c.GetHeader("X-User-ID")
	}

	if userID == "" {
		h.logger.Error("WebSocket connection rejected: missing user_id")
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user_id"})
		return
	}

	// 升级HTTP连接为WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade connection to WebSocket", zap.Error(err))
		return
	}

	// 创建连接对象
	connection := &Connection{
		ID:            generateConnectionID(),
		UserID:        userID,
		SessionID:     c.Query("session_id"),
		Conn:          conn,
		Send:          make(chan *Message, 256),
		Manager:       h.manager,
		Subscriptions: make(map[string]bool),
		LastPing:      time.Now(),
	}

	// 注册连接
	h.manager.register <- connection

	// 启动读写协程
	go h.writePump(connection)
	go h.readPump(connection)

	h.logger.Info("WebSocket connection established",
		zap.String("connection_id", connection.ID),
		zap.String("user_id", userID),
		zap.String("session_id", connection.SessionID),
	)
}

// readPump 读取消息
func (h *Handler) readPump(conn *Connection) {
	defer func() {
		h.manager.unregister <- conn
		conn.Conn.Close()
	}()

	conn.Conn.SetReadLimit(maxMessageSize)
	conn.Conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.Conn.SetPongHandler(func(string) error {
		conn.mu.Lock()
		conn.LastPing = time.Now()
		conn.mu.Unlock()
		conn.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, messageData, err := conn.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Error("WebSocket read error", zap.Error(err))
			}
			break
		}

		// 解析消息
		var msg Message
		if err := json.Unmarshal(messageData, &msg); err != nil {
			h.logger.Error("Failed to parse WebSocket message", zap.Error(err))
			continue
		}

		// 处理消息
		h.handleMessage(conn, &msg)
	}
}

// writePump 写入消息
func (h *Handler) writePump(conn *Connection) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-conn.Send:
			conn.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				conn.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := conn.Conn.WriteJSON(message); err != nil {
				h.logger.Error("Failed to write WebSocket message", zap.Error(err))
				return
			}

		case <-ticker.C:
			conn.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage 处理消息
func (h *Handler) handleMessage(conn *Connection, msg *Message) {
	switch msg.Type {
	case MessageTypePing:
		h.handlePing(conn, msg)
	case MessageTypeSubscribe:
		h.handleSubscribe(conn, msg)
	case MessageTypeUnsubscribe:
		h.handleUnsubscribe(conn, msg)
	case MessageTypeQuery:
		h.handleQuery(conn, msg)
	case MessageTypeSubmit:
		h.handleSubmit(conn, msg)
	case MessageTypePresence:
		h.handlePresence(conn, msg)
	default:
		h.logger.Warn("Unknown message type", zap.String("type", string(msg.Type)))
	}
}

// handlePing 处理心跳
func (h *Handler) handlePing(conn *Connection, msg *Message) {
	conn.mu.Lock()
	conn.LastPing = time.Now()
	conn.mu.Unlock()

	// 发送pong响应
	pongMsg := NewMessage(MessageTypePong, nil)
	select {
	case conn.Send <- pongMsg:
	default:
		h.logger.Error("Failed to send pong message")
	}
}

// handleSubscribe 处理订阅
func (h *Handler) handleSubscribe(conn *Connection, msg *Message) {
	if msg.Collection == "" {
		h.sendError(conn, msg.ID, 400, "collection is required")
		return
	}

	channel := msg.Collection
	if msg.Document != "" {
		channel = fmt.Sprintf("%s.%s", msg.Collection, msg.Document)
	}

	h.manager.Subscribe(conn.ID, channel)

	// 发送订阅确认
	response := NewMessage(MessageTypeSubscribe, map[string]string{
		"channel": channel,
		"status":  "subscribed",
	})
	response.ID = msg.ID

	select {
	case conn.Send <- response:
	default:
		h.logger.Error("Failed to send subscribe response")
	}
}

// handleUnsubscribe 处理取消订阅
func (h *Handler) handleUnsubscribe(conn *Connection, msg *Message) {
	if msg.Collection == "" {
		h.sendError(conn, msg.ID, 400, "collection is required")
		return
	}

	channel := msg.Collection
	if msg.Document != "" {
		channel = fmt.Sprintf("%s.%s", msg.Collection, msg.Document)
	}

	h.manager.Unsubscribe(conn.ID, channel)

	// 发送取消订阅确认
	response := NewMessage(MessageTypeUnsubscribe, map[string]string{
		"channel": channel,
		"status":  "unsubscribed",
	})
	response.ID = msg.ID

	select {
	case conn.Send <- response:
	default:
		h.logger.Error("Failed to send unsubscribe response")
	}
}

// handleQuery 处理查询
func (h *Handler) handleQuery(conn *Connection, msg *Message) {
	// TODO: 实现查询逻辑
	h.logger.Info("Query message received", zap.String("collection", msg.Collection))

	// 发送查询响应
	response := NewMessage(MessageTypeQueryResponse, QueryResponse{
		Data: []interface{}{},
	})
	response.ID = msg.ID

	select {
	case conn.Send <- response:
	default:
		h.logger.Error("Failed to send query response")
	}
}

// handleSubmit 处理提交
func (h *Handler) handleSubmit(conn *Connection, msg *Message) {
	// TODO: 实现提交逻辑
	h.logger.Info("Submit message received",
		zap.String("collection", msg.Collection),
		zap.String("document", msg.Document),
	)

	// 发送提交响应
	response := NewMessage(MessageTypeSubmitResponse, SubmitResponse{})
	response.ID = msg.ID

	select {
	case conn.Send <- response:
	default:
		h.logger.Error("Failed to send submit response")
	}
}

// handlePresence 处理在线状态
func (h *Handler) handlePresence(conn *Connection, msg *Message) {
	// TODO: 实现在线状态逻辑
	h.logger.Info("Presence message received", zap.String("collection", msg.Collection))
}

// sendError 发送错误消息
func (h *Handler) sendError(conn *Connection, msgID string, code int, message string) {
	errorMsg := NewErrorMessage(code, message)
	errorMsg.ID = msgID

	select {
	case conn.Send <- errorMsg:
	default:
		h.logger.Error("Failed to send error message")
	}
}

// generateConnectionID 生成连接ID
func generateConnectionID() string {
	return fmt.Sprintf("conn_%d", time.Now().UnixNano())
}
