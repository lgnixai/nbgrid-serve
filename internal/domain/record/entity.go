package record

import (
	"time"

	"teable-go-backend/pkg/utils"
)

// Record 数据记录实体
type Record struct {
	ID               string                 `json:"id"`
	TableID          string                 `json:"table_id"`
	Data             map[string]interface{} `json:"data"` // 记录数据，键为字段名，值为字段值
	CreatedBy        string                 `json:"created_by"`
	CreatedTime      time.Time              `json:"created_at"`
	DeletedTime      *time.Time             `json:"deleted_time"`
	LastModifiedTime *time.Time             `json:"updated_at"`
}

// CreateRecordRequest 创建记录请求
type CreateRecordRequest struct {
	TableID   string                 `json:"table_id" validate:"required"`
	Data      map[string]interface{} `json:"data"`
	CreatedBy string                 `json:"created_by,omitempty"`
}

// UpdateRecordRequest 更新记录请求
type UpdateRecordRequest struct {
	Data map[string]interface{} `json:"data"`
}

// ListRecordFilter 记录列表过滤条件
type ListRecordFilter struct {
	TableID   *string                `json:"table_id,omitempty" form:"table_id"`
	CreatedBy *string                `json:"created_by,omitempty" form:"created_by"`
	Search    string                 `json:"search,omitempty" form:"search"`
	OrderBy   string                 `json:"order_by,omitempty" form:"order_by"`
	Order     string                 `json:"order,omitempty" form:"order"`
	Limit     int                    `json:"limit,omitempty" form:"limit"`
	Offset    int                    `json:"offset,omitempty" form:"offset"`
	// 复杂查询条件
	FieldFilters map[string]interface{} `json:"field_filters,omitempty"` // 字段过滤条件
	DateRange    *DateRange             `json:"date_range,omitempty"`    // 日期范围过滤
	IsDeleted    *bool                  `json:"is_deleted,omitempty"`    // 是否包含已删除记录
}

// DateRange 日期范围
type DateRange struct {
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
}

// BulkUpdateRequest 批量更新请求
type BulkUpdateRequest struct {
	RecordIDs []string               `json:"record_ids"`
	Updates   map[string]interface{} `json:"updates"`
}

// BulkDeleteRequest 批量删除请求
type BulkDeleteRequest struct {
	RecordIDs []string `json:"record_ids"`
}

// RecordStats 记录统计信息
type RecordStats struct {
	TotalRecords   int64            `json:"total_records"`
	ActiveRecords  int64            `json:"active_records"`
	DeletedRecords int64            `json:"deleted_records"`
	RecordsByTable map[string]int64 `json:"records_by_table"`
	RecordsByUser  map[string]int64 `json:"records_by_user"`
	RecentActivity int64            `json:"recent_activity"` // 最近7天创建的记录数
}

// ExportRequest 导出请求
type ExportRequest struct {
	TableID      *string                `json:"table_id"`
	Format       string                 `json:"format"` // csv, json, excel
	FieldFilters map[string]interface{} `json:"field_filters"`
	DateRange    *DateRange             `json:"date_range"`
	Fields       []string               `json:"fields"` // 指定导出的字段
}

// ImportRequest 导入请求
type ImportRequest struct {
	TableID string        `json:"table_id"`
	Format  string        `json:"format"` // csv, json, excel
	Data    interface{}   `json:"data"`
	Options ImportOptions `json:"options"`
}

// ImportOptions 导入选项
type ImportOptions struct {
	SkipFirstRow   bool              `json:"skip_first_row"`  // 跳过第一行（标题行）
	UpdateExisting bool              `json:"update_existing"` // 更新已存在的记录
	CreateMissing  bool              `json:"create_missing"`  // 创建缺失的记录
	FieldMapping   map[string]string `json:"field_mapping"`   // 字段映射
}

// ComplexQueryRequest 复杂查询请求
type ComplexQueryRequest struct {
	TableID      string           `json:"table_id"`
	Conditions   []QueryCondition `json:"conditions"`
	GroupBy      []string         `json:"group_by"`
	Aggregations []Aggregation    `json:"aggregations"`
	OrderBy      []OrderBy        `json:"order_by"`
	Limit        int              `json:"limit"`
	Offset       int              `json:"offset"`
}

// QueryCondition 查询条件
type QueryCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // eq, ne, gt, gte, lt, lte, in, not_in, like, ilike, between
	Value    interface{} `json:"value"`
	Logic    string      `json:"logic"` // and, or
}

// Aggregation 聚合操作
type Aggregation struct {
	Field string `json:"field"`
	Type  string `json:"type"` // count, sum, avg, min, max
	Alias string `json:"alias"`
}

// OrderBy 排序条件
type OrderBy struct {
	Field string `json:"field"`
	Order string `json:"order"` // asc, desc
}

// NewRecord 创建新的数据记录
func NewRecord(req CreateRecordRequest) *Record {
	now := time.Now()
	return &Record{
		ID:               utils.GenerateRecordID(),
		TableID:          req.TableID,
		Data:             req.Data,
		CreatedBy:        req.CreatedBy,
		CreatedTime:      now,
		LastModifiedTime: &now,
	}
}

// Update 更新记录数据
func (r *Record) Update(req UpdateRecordRequest) {
	if req.Data != nil {
		r.Data = req.Data
	}
	now := time.Now()
	r.LastModifiedTime = &now
}

// SoftDelete 软删除记录
func (r *Record) SoftDelete() {
	now := time.Now()
	r.DeletedTime = &now
	r.LastModifiedTime = &now
}
