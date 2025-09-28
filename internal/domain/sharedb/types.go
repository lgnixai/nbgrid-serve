package sharedb

import (
	"time"
)

// DocumentType 文档类型
type DocumentType string

const (
	DocumentTypeRecord DocumentType = "record"
	DocumentTypeField  DocumentType = "field"
	DocumentTypeView   DocumentType = "view"
	DocumentTypeTable  DocumentType = "table"
)

// OperationType 操作类型
type OperationType string

const (
	OperationTypeCreate OperationType = "create"
	OperationTypeEdit   OperationType = "edit"
	OperationTypeDelete OperationType = "delete"
)

// OTOperation 操作转换操作
type OTOperation struct {
	P  []interface{} `json:"p"`  // 路径
	OI interface{}   `json:"oi"` // 插入值
	OD interface{}   `json:"od"` // 删除值
}

// RawOperation 原始操作
type RawOperation struct {
	Src    string        `json:"src"` // 源ID
	Seq    int           `json:"seq"` // 序列号
	V      int           `json:"v"`   // 版本
	C      string        `json:"c"`   // 集合
	D      string        `json:"d"`   // 文档ID
	Op     []OTOperation `json:"op,omitempty"`
	Create *CreateData   `json:"create,omitempty"`
	Del    bool          `json:"del,omitempty"`
}

// CreateData 创建数据
type CreateData struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Snapshot 文档快照
type Snapshot struct {
	ID   string      `json:"id"`
	V    int         `json:"v"`
	Type string      `json:"type"`
	Data interface{} `json:"data"`
	M    interface{} `json:"m,omitempty"` // 元数据
}

// Document 文档
type Document struct {
	ID         string      `json:"id"`
	Collection string      `json:"collection"`
	Version    int         `json:"version"`
	Type       string      `json:"type"`
	Data       interface{} `json:"data"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

// Query 查询
type Query struct {
	Fields map[string]interface{} `json:"fields,omitempty"`
	Sort   []SortField            `json:"sort,omitempty"`
	Limit  int                    `json:"limit,omitempty"`
	Skip   int                    `json:"skip,omitempty"`
}

// SortField 排序字段
type SortField struct {
	Field string `json:"field"`
	Order string `json:"order"` // "asc" or "desc"
}

// Projection 投影
type Projection map[string]bool

// Connection 连接
type Connection struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	SessionID string                 `json:"session_id"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
}

// Agent 代理
type Agent struct {
	Connection *Connection            `json:"connection"`
	Custom     map[string]interface{} `json:"custom"`
}

// SubmitContext 提交上下文
type SubmitContext struct {
	Agent      *Agent        `json:"agent"`
	Collection string        `json:"collection"`
	ID         string        `json:"id"`
	Op         *RawOperation `json:"op"`
	Source     string        `json:"source"`
}

// MiddlewareFunc 中间件函数
type MiddlewareFunc func(context *SubmitContext, next func(error error))

// RawOpMap 原始操作映射
type RawOpMap map[string]map[string]*RawOperation

// CollectionSnapshot 集合快照
type CollectionSnapshot struct {
	Type string      `json:"type"`
	V    int         `json:"v"`
	Data interface{} `json:"data"`
}

// QueryResult 查询结果
type QueryResult struct {
	IDs   []string    `json:"ids"`
	Extra interface{} `json:"extra,omitempty"`
}

// SnapshotBulkResult 批量快照结果
type SnapshotBulkResult struct {
	Snapshots map[string]*Snapshot `json:"snapshots"`
	Extra     interface{}          `json:"extra,omitempty"`
}

// VersionAndType 版本和类型
type VersionAndType struct {
	Version int           `json:"version"`
	Type    OperationType `json:"type"`
}

// Database 数据库接口
type Database interface {
	// 获取快照
	GetSnapshot(collection, id string, projection Projection, options interface{}) (*Snapshot, error)

	// 批量获取快照
	GetSnapshotBulk(collection string, ids []string, projection Projection, options interface{}) (map[string]*Snapshot, error)

	// 获取操作
	GetOps(collection, id string, from, to int, options interface{}) ([]*RawOperation, error)

	// 查询文档
	Query(collection string, query *Query, projection Projection, options interface{}) ([]*Snapshot, interface{}, error)

	// 提交操作
	Commit(collection, id string, op *RawOperation) error

	// 关闭连接
	Close() error
}

// PubSub 发布订阅接口
type PubSub interface {
	// 订阅频道
	Subscribe(channel string, callback func(channel string, data interface{})) error

	// 取消订阅
	Unsubscribe(channel string) error

	// 发布消息
	Publish(channels []string, data interface{}) error

	// 关闭
	Close() error
}

// ShareDB ShareDB服务接口
type ShareDB interface {
	// 获取连接
	GetConnection() *Connection

	// 处理提交
	OnSubmit(context *SubmitContext, next func(error error))

	// 发布操作映射
	PublishOpsMap(rawOpMaps []RawOpMap) error

	// 发布记录频道
	PublishRecordChannel(tableID string, rawOp *RawOperation) error

	// 添加中间件
	Use(event string, middleware MiddlewareFunc)

	// 数据库操作
	GetSnapshot(collection, id string, projection Projection, options interface{}) (*Snapshot, error)
	GetSnapshotBulk(collection string, ids []string, projection Projection, options interface{}) (map[string]*Snapshot, error)
	GetOps(collection, id string, from, to int, options interface{}) ([]*RawOperation, error)
	Query(collection string, query *Query, projection Projection, options interface{}) ([]*Snapshot, interface{}, error)
	Commit(collection, id string, op *RawOperation) error

	// 操作转换
	ApplyOperation(doc interface{}, op OTOperation, typeName string) (interface{}, error)
	TransformOperations(op1, op2 OTOperation, typeName string) (OTOperation, OTOperation, error)
	ValidateOperation(op OTOperation, doc interface{}, typeName string) error

	// 统计信息
	GetStats() map[string]interface{}

	// 关闭
	Close() error
}
