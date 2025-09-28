package websocket

import (
	"encoding/json"
	"time"
)

// MessageType WebSocket消息类型
type MessageType string

const (
	// 连接相关
	MessageTypeConnect    MessageType = "connect"
	MessageTypeConnected  MessageType = "connected"
	MessageTypeDisconnect MessageType = "disconnect"
	MessageTypeError      MessageType = "error"

	// 文档操作相关
	MessageTypeSubscribe      MessageType = "subscribe"
	MessageTypeUnsubscribe    MessageType = "unsubscribe"
	MessageTypeQuery          MessageType = "query"
	MessageTypeQueryResponse  MessageType = "queryResponse"
	MessageTypeSubmit         MessageType = "submit"
	MessageTypeSubmitResponse MessageType = "submitResponse"
	MessageTypeOp             MessageType = "op"
	MessageTypePresence       MessageType = "presence"

	// 协作相关
	MessageTypeCursor       MessageType = "cursor"
	MessageTypeNotification MessageType = "notification"
	MessageTypeConflict     MessageType = "conflict"

	// 心跳相关
	MessageTypePing MessageType = "ping"
	MessageTypePong MessageType = "pong"
)

// Message WebSocket消息结构
type Message struct {
	Type       MessageType `json:"type"`
	ID         string      `json:"id,omitempty"`
	Collection string      `json:"collection,omitempty"`
	Document   string      `json:"document,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	Error      *Error      `json:"error,omitempty"`
	Timestamp  time.Time   `json:"timestamp"`
}

// Error 错误信息
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ConnectionInfo 连接信息
type ConnectionInfo struct {
	UserID    string            `json:"user_id"`
	SessionID string            `json:"session_id"`
	Metadata  map[string]string `json:"metadata"`
}

// DocumentOperation 文档操作
type DocumentOperation struct {
	Op     []interface{} `json:"op"`
	Source string        `json:"source,omitempty"`
}

// QueryRequest 查询请求
type QueryRequest struct {
	Collection string                 `json:"collection"`
	Query      map[string]interface{} `json:"query"`
	Options    map[string]interface{} `json:"options,omitempty"`
}

// QueryResponse 查询响应
type QueryResponse struct {
	Data  []interface{} `json:"data"`
	Extra interface{}   `json:"extra,omitempty"`
	Error *Error        `json:"error,omitempty"`
}

// SubmitRequest 提交请求
type SubmitRequest struct {
	Collection string                 `json:"collection"`
	Document   string                 `json:"document"`
	Op         []interface{}          `json:"op"`
	Source     string                 `json:"source,omitempty"`
	Options    map[string]interface{} `json:"options,omitempty"`
}

// SubmitResponse 提交响应
type SubmitResponse struct {
	Error *Error `json:"error,omitempty"`
}

// PresenceInfo 在线状态信息
type PresenceInfo struct {
	UserID    string                 `json:"user_id"`
	SessionID string                 `json:"session_id"`
	Data      map[string]interface{} `json:"data"`
	LastSeen  time.Time              `json:"last_seen"`
}

// Channel 频道信息
type Channel struct {
	Name     string            `json:"name"`
	Type     string            `json:"type"` // collection, document, presence
	Metadata map[string]string `json:"metadata,omitempty"`
}

// CursorInfo 光标信息
type CursorInfo struct {
	UserID    string                 `json:"user_id"`
	SessionID string                 `json:"session_id"`
	Position  map[string]interface{} `json:"position"`
	Selection map[string]interface{} `json:"selection,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// CollaborationMessage 协作消息
type CollaborationMessage struct {
	Type       string                 `json:"type"`
	UserID     string                 `json:"user_id"`
	SessionID  string                 `json:"session_id"`
	Collection string                 `json:"collection"`
	Document   string                 `json:"document,omitempty"`
	Data       map[string]interface{} `json:"data"`
	Timestamp  time.Time              `json:"timestamp"`
}

// ToJSON 将消息转换为JSON
func (m *Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// FromJSON 从JSON解析消息
func (m *Message) FromJSON(data []byte) error {
	return json.Unmarshal(data, m)
}

// NewMessage 创建新消息
func NewMessage(msgType MessageType, data interface{}) *Message {
	return &Message{
		Type:      msgType,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// NewErrorMessage 创建错误消息
func NewErrorMessage(code int, message string) *Message {
	return &Message{
		Type: MessageTypeError,
		Error: &Error{
			Code:    code,
			Message: message,
		},
		Timestamp: time.Now(),
	}
}
