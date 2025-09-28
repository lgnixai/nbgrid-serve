package record

import "context"

// Repository 记录仓储接口
type Repository interface {
	Create(ctx context.Context, record *Record) error
	GetByID(ctx context.Context, id string) (*Record, error)
	Update(ctx context.Context, record *Record) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter ListRecordFilter) ([]*Record, error)
	Count(ctx context.Context, filter ListRecordFilter) (int64, error)
	Exists(ctx context.Context, filter ListRecordFilter) (bool, error)

	// 批量操作
	BulkCreate(ctx context.Context, records []*Record) error
	BulkUpdate(ctx context.Context, updates map[string]map[string]interface{}) error
	BulkDelete(ctx context.Context, recordIDs []string) error

	// 复杂查询
	ComplexQuery(ctx context.Context, req ComplexQueryRequest) ([]map[string]interface{}, error)

	// 统计
	GetRecordStats(ctx context.Context, tableID *string) (*RecordStats, error)

	// 导出导入
	ExportRecords(ctx context.Context, req ExportRequest) ([]byte, error)
	ImportRecords(ctx context.Context, req ImportRequest) (int, error)
}
