package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"teable-go-backend/internal/config"
	"teable-go-backend/pkg/logger"
)

// MessageHandler 消息处理器函数类型
type MessageHandler func(channel string, data interface{}) error

// RedisPubSub Redis发布订阅服务
type RedisPubSub struct {
	client   *redis.Client
	observer *redis.Client
	handlers map[string][]MessageHandler
	mu       sync.RWMutex
	logger   *zap.Logger
	prefix   string
	closing  bool
}

// PubSubMessage 发布订阅消息结构
type PubSubMessage struct {
	Type      string      `json:"type"`
	Channel   string      `json:"channel"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	Source    string      `json:"source,omitempty"`
}

// NewRedisPubSub 创建Redis发布订阅服务
func NewRedisPubSub(cfg config.RedisConfig, prefix string) (*RedisPubSub, error) {
	// 创建发布客户端
	client := redis.NewClient(&redis.Options{
		Addr:        cfg.GetRedisAddr(),
		Password:    cfg.Password,
		DB:          cfg.DB,
		PoolSize:    cfg.PoolSize,
		DialTimeout: cfg.DialTimeout,
	})

	// 创建订阅客户端（需要独立的连接）
	observer := redis.NewClient(&redis.Options{
		Addr:        cfg.GetRedisAddr(),
		Password:    cfg.Password,
		DB:          cfg.DB,
		PoolSize:    1, // 订阅客户端只需要一个连接
		DialTimeout: cfg.DialTimeout,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis client: %w", err)
	}

	if err := observer.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis observer: %w", err)
	}

	pubsub := &RedisPubSub{
		client:   client,
		observer: observer,
		handlers: make(map[string][]MessageHandler),
		logger:   logger.Logger,
		prefix:   prefix,
	}

	// 启动消息监听
	go pubsub.listen()

	logger.Info("Redis Pub/Sub service initialized",
		logger.String("addr", cfg.GetRedisAddr()),
		logger.String("prefix", prefix),
	)

	return pubsub, nil
}

// listen 监听Redis消息
func (r *RedisPubSub) listen() {
	pubsub := r.observer.Subscribe(context.Background())
	defer pubsub.Close()

	// 处理订阅消息
	ch := pubsub.Channel()
	for msg := range ch {
		if r.closing {
			break
		}

		// 复制消息以避免闭包问题
		msgCopy := msg

		// 解析消息
		var pubsubMsg PubSubMessage
		if err := json.Unmarshal([]byte(msgCopy.Payload), &pubsubMsg); err != nil {
			r.logger.Error("Failed to unmarshal pubsub message",
				logger.String("channel", msgCopy.Channel),
				logger.ErrorField(err),
			)
			continue
		}

		// 调用处理器
		r.mu.RLock()
		handlers, exists := r.handlers[msgCopy.Channel]
		r.mu.RUnlock()

		if exists {
			for _, handler := range handlers {
				go func(h MessageHandler, channel string, data interface{}) {
					if err := h(channel, data); err != nil {
						r.logger.Error("Message handler error",
							logger.String("channel", channel),
							logger.ErrorField(err),
						)
					}
				}(handler, msgCopy.Channel, pubsubMsg.Data)
			}
		}
	}
}

// Subscribe 订阅频道
func (r *RedisPubSub) Subscribe(channel string, callback func(channel string, data interface{})) error {
	if r.closing {
		return fmt.Errorf("pubsub service is closing")
	}

	// 添加处理器
	r.mu.Lock()
	handler := func(channel string, data interface{}) error {
		callback(channel, data)
		return nil
	}
	r.handlers[channel] = append(r.handlers[channel], handler)
	r.mu.Unlock()

	// 订阅Redis频道
	ctx := context.Background()
	pubsub := r.observer.Subscribe(ctx, channel)
	// 检查订阅是否成功
	if pubsub == nil {
		// 如果订阅失败，移除处理器
		r.mu.Lock()
		if handlers, exists := r.handlers[channel]; exists && len(handlers) > 0 {
			// 移除最后一个添加的处理器
			r.handlers[channel] = handlers[:len(handlers)-1]
		}
		r.mu.Unlock()
		return fmt.Errorf("failed to subscribe to channel %s", channel)
	}

	r.logger.Info("Subscribed to channel",
		logger.String("channel", channel),
	)

	return nil
}

// Unsubscribe 取消订阅频道
func (r *RedisPubSub) Unsubscribe(channel string) error {
	if r.closing {
		return nil
	}

	// 移除处理器
	r.mu.Lock()
	if _, exists := r.handlers[channel]; exists {
		// 删除所有处理器
		delete(r.handlers, channel)
		r.mu.Unlock()
		// 取消Redis订阅
		ctx := context.Background()
		pubsub := r.observer.Subscribe(ctx, channel)
		if pubsub != nil {
			pubsub.Unsubscribe(ctx, channel)
		}
	} else {
		r.mu.Unlock()
	}

	r.logger.Info("Unsubscribed from channel",
		logger.String("channel", channel),
	)

	return nil
}

// Publish 发布消息到频道
func (r *RedisPubSub) Publish(channels []string, data interface{}) error {
	if r.closing {
		return fmt.Errorf("pubsub service is closing")
	}

	message := PubSubMessage{
		Type:      "message",
		Data:      data,
		Timestamp: time.Now(),
		Source:    "server",
	}

	ctx := context.Background()
	pipe := r.client.Pipeline()

	for _, channel := range channels {
		message.Channel = channel
		payload, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("failed to marshal message for channel %s: %w", channel, err)
		}
		pipe.Publish(ctx, channel, payload)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to publish to channels: %w", err)
	}

	r.logger.Debug("Published message to channels",
		logger.Int("channel_count", len(channels)),
	)

	return nil
}

// PublishToChannels 发布消息到多个频道
func (r *RedisPubSub) PublishToChannels(channels []string, data interface{}) error {
	if r.closing {
		return fmt.Errorf("pubsub service is closing")
	}

	message := PubSubMessage{
		Type:      "message",
		Data:      data,
		Timestamp: time.Now(),
		Source:    "server",
	}

	ctx := context.Background()
	pipe := r.client.Pipeline()

	for _, channel := range channels {
		message.Channel = channel
		channelPayload, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("failed to marshal message for channel %s: %w", channel, err)
		}
		pipe.Publish(ctx, channel, channelPayload)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to publish to channels: %w", err)
	}

	r.logger.Debug("Published message to multiple channels",
		logger.Int("channel_count", len(channels)),
	)

	return nil
}

// GetSubscribedChannels 获取已订阅的频道列表
func (r *RedisPubSub) GetSubscribedChannels() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	channels := make([]string, 0, len(r.handlers))
	for channel := range r.handlers {
		channels = append(channels, channel)
	}

	return channels
}

// GetChannelSubscriberCount 获取频道的订阅者数量
func (r *RedisPubSub) GetChannelSubscriberCount(channel string) (int64, error) {
	ctx := context.Background()
	result, err := r.client.PubSubNumSub(ctx, channel).Result()
	if err != nil {
		return 0, err
	}
	if count, exists := result[channel]; exists {
		return count, nil
	}
	return 0, nil
}

// GetStats 获取统计信息
func (r *RedisPubSub) GetStats() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return map[string]interface{}{
		"subscribed_channels": len(r.handlers),
		"total_handlers":      r.getTotalHandlers(),
		"closing":             r.closing,
	}
}

// getTotalHandlers 获取处理器总数
func (r *RedisPubSub) getTotalHandlers() int {
	total := 0
	for _, handlers := range r.handlers {
		total += len(handlers)
	}
	return total
}

// Close 关闭发布订阅服务
func (r *RedisPubSub) Close() error {
	if r.closing {
		return nil
	}

	r.closing = true

	// 关闭Redis连接
	if err := r.client.Close(); err != nil {
		r.logger.Error("Failed to close redis client", logger.ErrorField(err))
	}

	if err := r.observer.Close(); err != nil {
		r.logger.Error("Failed to close redis observer", logger.ErrorField(err))
	}

	r.logger.Info("Redis Pub/Sub service closed")
	return nil
}

// Health 检查服务健康状态
func (r *RedisPubSub) Health(ctx context.Context) error {
	if r.closing {
		return fmt.Errorf("pubsub service is closing")
	}

	if err := r.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis client health check failed: %w", err)
	}

	if err := r.observer.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis observer health check failed: %w", err)
	}

	return nil
}
