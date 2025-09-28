package sharedb

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"teable-go-backend/pkg/logger"
)

// Service ShareDB服务实现
type Service struct {
	db          Database
	pubsub      PubSub
	otEngine    *OTEngine
	logger      *zap.Logger
	connections map[string]*Connection
	mu          sync.RWMutex
	middlewares map[string][]MiddlewareFunc
	closing     bool
}

// NewService 创建ShareDB服务
func NewService(db Database, pubsub PubSub, logger *zap.Logger) *Service {
	service := &Service{
		db:          db,
		pubsub:      pubsub,
		otEngine:    NewOTEngine(),
		logger:      logger,
		connections: make(map[string]*Connection),
		middlewares: make(map[string][]MiddlewareFunc),
		closing:     false,
	}

	// 注册默认中间件
	service.Use("submit", service.defaultSubmitMiddleware)

	return service
}

// GetConnection 获取连接
func (s *Service) GetConnection() *Connection {
	connID := fmt.Sprintf("conn_%d", time.Now().UnixNano())
	conn := &Connection{
		ID:        connID,
		UserID:    "anonymous",
		SessionID: fmt.Sprintf("session_%d", time.Now().UnixNano()),
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
	}

	s.mu.Lock()
	s.connections[connID] = conn
	s.mu.Unlock()

	s.logger.Info("New connection created",
		logger.String("connection_id", connID),
	)

	return conn
}

// OnSubmit 处理提交
func (s *Service) OnSubmit(context *SubmitContext, next func(error error)) {
	if s.closing {
		next(fmt.Errorf("service is closing"))
		return
	}

	// 执行中间件
	middlewares, exists := s.middlewares["submit"]
	if !exists {
		next(nil)
		return
	}

	s.executeMiddlewares(middlewares, context, next)
}

// executeMiddlewares 执行中间件链
func (s *Service) executeMiddlewares(middlewares []MiddlewareFunc, context *SubmitContext, finalNext func(error error)) {
	if len(middlewares) == 0 {
		finalNext(nil)
		return
	}

	current := 0
	var next func(error error)
	next = func(err error) {
		if err != nil {
			finalNext(err)
			return
		}

		if current >= len(middlewares) {
			finalNext(nil)
			return
		}

		middleware := middlewares[current]
		current++
		middleware(context, next)
	}

	next(nil)
}

// defaultSubmitMiddleware 默认提交中间件
func (s *Service) defaultSubmitMiddleware(context *SubmitContext, next func(error error)) {
	// 验证操作
	if context.Op == nil {
		next(fmt.Errorf("operation is required"))
		return
	}

	// 验证集合和文档ID
	if context.Collection == "" || context.ID == "" {
		next(fmt.Errorf("collection and document ID are required"))
		return
	}

	// 记录操作
	s.logger.Debug("Processing submit",
		logger.String("collection", context.Collection),
		logger.String("document_id", context.ID),
		logger.String("operation_type", s.getOperationType(context.Op)),
	)

	next(nil)
}

// getOperationType 获取操作类型
func (s *Service) getOperationType(op *RawOperation) string {
	if op.Create != nil {
		return "create"
	} else if op.Del {
		return "delete"
	} else if len(op.Op) > 0 {
		return "edit"
	}
	return "unknown"
}

// PublishOpsMap 发布操作映射
func (s *Service) PublishOpsMap(rawOpMaps []RawOpMap) error {
	if s.closing {
		return fmt.Errorf("service is closing")
	}

	for _, rawOpMap := range rawOpMaps {
		for collection, docs := range rawOpMap {
			for docID, rawOp := range docs {
				// 设置集合和文档ID
				rawOp.C = collection
				rawOp.D = docID

				// 发布到相关频道
				channels := []string{collection, fmt.Sprintf("%s.%s", collection, docID)}

				if err := s.pubsub.Publish(channels, rawOp); err != nil {
					s.logger.Error("Failed to publish operation",
						logger.String("collection", collection),
						logger.String("document_id", docID),
						logger.ErrorField(err),
					)
					continue
				}

				// 发布到相关频道（如果需要）
				if s.shouldPublishAction(rawOp) {
					tableID := s.extractTableID(collection)
					if tableID != "" {
						s.publishRelatedChannels(tableID, rawOp)
					}
				}

				s.logger.Debug("Published operation",
					logger.String("collection", collection),
					logger.String("document_id", docID),
					logger.String("channels", fmt.Sprintf("%v", channels)),
				)
			}
		}
	}

	return nil
}

// shouldPublishAction 检查是否应该发布操作
func (s *Service) shouldPublishAction(rawOp *RawOperation) bool {
	if len(rawOp.Op) == 0 {
		return false
	}

	// 检查视图相关操作
	viewKeys := []string{"filter", "sort", "group", "lastModifiedTime"}
	fieldKeys := []string{"options"}

	for _, op := range rawOp.Op {
		if len(op.P) > 0 {
			key := fmt.Sprintf("%v", op.P[len(op.P)-1])
			for _, viewKey := range viewKeys {
				if key == viewKey {
					return true
				}
			}
			for _, fieldKey := range fieldKeys {
				if key == fieldKey {
					return true
				}
			}
		}
	}

	return false
}

// extractTableID 提取表ID
func (s *Service) extractTableID(collection string) string {
	// 假设集合格式为 "type_tableID"
	parts := splitCollection(collection)
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

// splitCollection 分割集合名称
func splitCollection(collection string) []string {
	// 简单的分割实现，实际可能需要更复杂的逻辑
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

// publishRelatedChannels 发布到相关频道
func (s *Service) publishRelatedChannels(tableID string, rawOp *RawOperation) {
	channels := []string{
		fmt.Sprintf("record_%s", tableID),
		fmt.Sprintf("field_%s", tableID),
	}

	for _, channel := range channels {
		if err := s.pubsub.Publish([]string{channel}, rawOp); err != nil {
			s.logger.Error("Failed to publish to related channel",
				logger.String("channel", channel),
				logger.ErrorField(err),
			)
		}
	}
}

// PublishRecordChannel 发布记录频道
func (s *Service) PublishRecordChannel(tableID string, rawOp *RawOperation) error {
	channel := fmt.Sprintf("record_%s", tableID)
	return s.pubsub.Publish([]string{channel}, rawOp)
}

// Use 添加中间件
func (s *Service) Use(event string, middleware MiddlewareFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.middlewares[event] = append(s.middlewares[event], middleware)

	s.logger.Debug("Middleware registered",
		logger.String("event", event),
	)
}

// GetSnapshot 获取快照
func (s *Service) GetSnapshot(collection, id string, projection Projection, options interface{}) (*Snapshot, error) {
	return s.db.GetSnapshot(collection, id, projection, options)
}

// GetSnapshotBulk 批量获取快照
func (s *Service) GetSnapshotBulk(collection string, ids []string, projection Projection, options interface{}) (map[string]*Snapshot, error) {
	return s.db.GetSnapshotBulk(collection, ids, projection, options)
}

// GetOps 获取操作
func (s *Service) GetOps(collection, id string, from, to int, options interface{}) ([]*RawOperation, error) {
	return s.db.GetOps(collection, id, from, to, options)
}

// Query 查询文档
func (s *Service) Query(collection string, query *Query, projection Projection, options interface{}) ([]*Snapshot, interface{}, error) {
	return s.db.Query(collection, query, projection, options)
}

// Commit 提交操作
func (s *Service) Commit(collection, id string, op *RawOperation) error {
	return s.db.Commit(collection, id, op)
}

// ApplyOperation 应用操作
func (s *Service) ApplyOperation(doc interface{}, op OTOperation, typeName string) (interface{}, error) {
	return s.otEngine.ApplyOperation(doc, op, typeName)
}

// TransformOperations 转换操作
func (s *Service) TransformOperations(op1, op2 OTOperation, typeName string) (OTOperation, OTOperation, error) {
	return s.otEngine.TransformOperations(op1, op2, typeName)
}

// ValidateOperation 验证操作
func (s *Service) ValidateOperation(op OTOperation, doc interface{}, typeName string) error {
	return s.otEngine.ValidateOperation(op, doc, typeName)
}

// GetStats 获取统计信息
func (s *Service) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"connections": len(s.connections),
		"middlewares": len(s.middlewares),
		"closing":     s.closing,
	}
}

// Close 关闭服务
func (s *Service) Close() error {
	if s.closing {
		return nil
	}

	s.closing = true

	// 关闭数据库连接
	if err := s.db.Close(); err != nil {
		s.logger.Error("Failed to close database", logger.ErrorField(err))
	}

	// 关闭发布订阅
	if err := s.pubsub.Close(); err != nil {
		s.logger.Error("Failed to close pubsub", logger.ErrorField(err))
	}

	// 清理连接
	s.mu.Lock()
	s.connections = make(map[string]*Connection)
	s.mu.Unlock()

	s.logger.Info("ShareDB service closed")
	return nil
}
