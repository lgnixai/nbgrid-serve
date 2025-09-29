package sharedb

import (
	"sync"
	"time"

	"go.uber.org/zap"

	"teable-go-backend/pkg/logger"
)

// MemoryPubSub 内存发布订阅实现
type MemoryPubSub struct {
	subscribers map[string][]func(channel string, data interface{})
	mu          sync.RWMutex
	logger      *zap.Logger
	closed      bool
}

// NewMemoryPubSub 创建内存发布订阅服务
func NewMemoryPubSub() *MemoryPubSub {
	return &MemoryPubSub{
		subscribers: make(map[string][]func(channel string, data interface{})),
		logger:      logger.Logger,
		closed:      false,
	}
}

// Subscribe 订阅频道
func (m *MemoryPubSub) Subscribe(channel string, callback func(channel string, data interface{})) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return ErrPubSubClosed
	}

	m.subscribers[channel] = append(m.subscribers[channel], callback)

	m.logger.Debug("Subscribed to channel",
		logger.String("channel", channel),
		logger.Int("subscriber_count", len(m.subscribers[channel])),
	)

	return nil
}

// Unsubscribe 取消订阅
func (m *MemoryPubSub) Unsubscribe(channel string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return ErrPubSubClosed
	}

	delete(m.subscribers, channel)

	m.logger.Debug("Unsubscribed from channel",
		logger.String("channel", channel),
	)

	return nil
}

// Publish 发布消息
func (m *MemoryPubSub) Publish(channels []string, data interface{}) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return ErrPubSubClosed
	}

	for _, channel := range channels {
		callbacks, exists := m.subscribers[channel]
		if !exists {
			continue
		}

		// 异步调用回调函数
		for _, callback := range callbacks {
			go func(cb func(channel string, data interface{}), ch string, d interface{}) {
				defer func() {
					if r := recover(); r != nil {
						m.logger.Error("Panic in pub/sub callback",
							logger.String("channel", ch),
							logger.Any("panic", r),
						)
					}
				}()

				cb(ch, d)
			}(callback, channel, data)
		}

		m.logger.Debug("Published message to channel",
			logger.String("channel", channel),
			logger.Int("subscriber_count", len(callbacks)),
		)
	}

	return nil
}

// Close 关闭发布订阅服务
func (m *MemoryPubSub) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil
	}

	m.closed = true
	m.subscribers = make(map[string][]func(channel string, data interface{}))

	m.logger.Info("Memory pub/sub service closed")
	return nil
}

// GetStats 获取统计信息
func (m *MemoryPubSub) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	totalSubscribers := 0
	for _, callbacks := range m.subscribers {
		totalSubscribers += len(callbacks)
	}

	return map[string]interface{}{
		"channels":          len(m.subscribers),
		"total_subscribers": totalSubscribers,
		"closed":            m.closed,
	}
}

// WaitForSubscribers 等待订阅者
func (m *MemoryPubSub) WaitForSubscribers(channel string, timeout time.Duration) error {
	start := time.Now()
	for {
		m.mu.RLock()
		callbacks, exists := m.subscribers[channel]
		m.mu.RUnlock()

		if exists && len(callbacks) > 0 {
			return nil
		}

		if time.Since(start) > timeout {
			return ErrTimeout
		}

		time.Sleep(10 * time.Millisecond)
	}
}

// 错误定义
var (
	ErrPubSubClosed = &PubSubError{Message: "pub/sub service is closed"}
	ErrTimeout      = &PubSubError{Message: "operation timeout"}
)

// PubSubError 发布订阅错误
type PubSubError struct {
	Message string
}

func (e *PubSubError) Error() string {
	return e.Message
}
