package record

import (
	"fmt"
	"time"

	"teable-go-backend/internal/domain/table"
	"teable-go-backend/pkg/utils"
)

// Record 数据记录实体 - 支持动态数据存储和验证
type Record struct {
	ID               string                 `json:"id"`
	TableID          string                 `json:"table_id"`
	Data             map[string]interface{} `json:"data"` // 动态数据存储，键为字段名，值为字段值
	Version          int64                  `json:"version"`           // 记录版本号，用于并发控制
	Hash             string                 `json:"hash"`              // 数据哈希值，用于变更检测
	CreatedBy        string                 `json:"created_by"`
	UpdatedBy        *string                `json:"updated_by"`        // 最后更新者
	CreatedTime      time.Time              `json:"created_at"`
	DeletedTime      *time.Time             `json:"deleted_time"`
	LastModifiedTime *time.Time             `json:"updated_at"`
	
	// 内部字段，不序列化
	tableSchema      *table.Table           `json:"-"` // 表格schema缓存
	validationErrors []ValidationError      `json:"-"` // 验证错误列表
}

// CreateRecordRequest 创建记录请求
type CreateRecordRequest struct {
	TableID   string                 `json:"table_id" validate:"required"`
	Data      map[string]interface{} `json:"data"`
	CreatedBy string                 `json:"created_by,omitempty"`
}

// UpdateRecordRequest 更新记录请求
type UpdateRecordRequest struct {
	Data      map[string]interface{} `json:"data"`
	UpdatedBy string                 `json:"updated_by,omitempty"` // 更新者
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

// ValidationError 验证错误
type ValidationError struct {
	FieldName string `json:"field_name"`
	Message   string `json:"message"`
	Code      string `json:"code"`
}

// RecordChangeEvent 记录变更事件
type RecordChangeEvent struct {
	RecordID    string                 `json:"record_id"`
	TableID     string                 `json:"table_id"`
	ChangeType  string                 `json:"change_type"` // create, update, delete
	OldData     map[string]interface{} `json:"old_data,omitempty"`
	NewData     map[string]interface{} `json:"new_data,omitempty"`
	ChangedBy   string                 `json:"changed_by"`
	ChangedAt   time.Time              `json:"changed_at"`
	Version     int64                  `json:"version"`
}

// NewRecord 创建新的数据记录
func NewRecord(req CreateRecordRequest) *Record {
	now := time.Now()
	record := &Record{
		ID:               utils.GenerateRecordID(),
		TableID:          req.TableID,
		Data:             make(map[string]interface{}),
		Version:          1,
		CreatedBy:        req.CreatedBy,
		CreatedTime:      now,
		LastModifiedTime: &now,
		validationErrors: make([]ValidationError, 0),
	}
	
	// 设置数据
	if req.Data != nil {
		record.Data = req.Data
	}
	
	// 计算哈希值
	record.updateHash()
	
	return record
}

// SetTableSchema 设置表格schema用于验证
func (r *Record) SetTableSchema(schema *table.Table) {
	r.tableSchema = schema
}

// ValidateData 验证记录数据
func (r *Record) ValidateData() error {
	r.validationErrors = make([]ValidationError, 0)
	
	if r.tableSchema == nil {
		return fmt.Errorf("缺少表格schema信息")
	}
	
	fields := r.tableSchema.GetFields()
	if len(fields) == 0 {
		return fmt.Errorf("表格没有定义字段")
	}
	
	// 验证必填字段
	for _, field := range fields {
		if field.DeletedTime != nil {
			continue // 跳过已删除的字段
		}
		
		value, exists := r.Data[field.Name]
		
		// 检查必填字段
		if field.IsRequired && (!exists || value == nil || value == "") {
			r.validationErrors = append(r.validationErrors, ValidationError{
				FieldName: field.Name,
				Message:   fmt.Sprintf("字段 '%s' 是必填的", field.Name),
				Code:      "REQUIRED_FIELD_MISSING",
			})
			continue
		}
		
		// 如果字段有值，进行类型验证
		if exists && value != nil && value != "" {
			if err := field.ValidateValue(value); err != nil {
				r.validationErrors = append(r.validationErrors, ValidationError{
					FieldName: field.Name,
					Message:   err.Error(),
					Code:      "FIELD_VALIDATION_FAILED",
				})
			}
		}
	}
	
	// 检查是否有未知字段
	for fieldName := range r.Data {
		if r.tableSchema.GetFieldByName(fieldName) == nil {
			r.validationErrors = append(r.validationErrors, ValidationError{
				FieldName: fieldName,
				Message:   fmt.Sprintf("未知字段 '%s'", fieldName),
				Code:      "UNKNOWN_FIELD",
			})
		}
	}
	
	if len(r.validationErrors) > 0 {
		return fmt.Errorf("记录验证失败，共 %d 个错误", len(r.validationErrors))
	}
	
	return nil
}

// GetValidationErrors 获取验证错误列表
func (r *Record) GetValidationErrors() []ValidationError {
	return r.validationErrors
}

// ApplyFieldDefaults 应用字段默认值
func (r *Record) ApplyFieldDefaults() error {
	if r.tableSchema == nil {
		return fmt.Errorf("缺少表格schema信息")
	}
	
	fields := r.tableSchema.GetFields()
	for _, field := range fields {
		if field.DeletedTime != nil {
			continue
		}
		
		// 如果字段没有值且有默认值，则应用默认值
		if _, exists := r.Data[field.Name]; !exists && field.DefaultValue != nil {
			r.Data[field.Name] = *field.DefaultValue
		}
		
		// 处理系统字段
		switch field.Type {
		case table.FieldTypeCreatedTime:
			if _, exists := r.Data[field.Name]; !exists {
				r.Data[field.Name] = r.CreatedTime.Format(time.RFC3339)
			}
		case table.FieldTypeLastModifiedTime:
			r.Data[field.Name] = r.LastModifiedTime.Format(time.RFC3339)
		case table.FieldTypeCreatedBy:
			if _, exists := r.Data[field.Name]; !exists {
				r.Data[field.Name] = r.CreatedBy
			}
		case table.FieldTypeLastModifiedBy:
			if r.UpdatedBy != nil {
				r.Data[field.Name] = *r.UpdatedBy
			} else {
				r.Data[field.Name] = r.CreatedBy
			}
		}
	}
	
	return nil
}

// Update 更新记录数据
func (r *Record) Update(req UpdateRecordRequest, updatedBy string) error {
	if req.Data == nil {
		return fmt.Errorf("更新数据不能为空")
	}
	
	// 保存旧数据用于变更检测
	oldData := make(map[string]interface{})
	for k, v := range r.Data {
		oldData[k] = v
	}
	
	// 更新数据
	for key, value := range req.Data {
		r.Data[key] = value
	}
	
	// 更新元数据
	r.UpdatedBy = &updatedBy
	now := time.Now()
	r.LastModifiedTime = &now
	r.Version++
	
	// 应用字段默认值
	if err := r.ApplyFieldDefaults(); err != nil {
		return fmt.Errorf("应用字段默认值失败: %v", err)
	}
	
	// 验证数据
	if err := r.ValidateData(); err != nil {
		return fmt.Errorf("数据验证失败: %v", err)
	}
	
	// 更新哈希值
	r.updateHash()
	
	return nil
}

// UpdateField 更新单个字段
func (r *Record) UpdateField(fieldName string, value interface{}, updatedBy string) error {
	if r.tableSchema == nil {
		return fmt.Errorf("缺少表格schema信息")
	}
	
	field := r.tableSchema.GetFieldByName(fieldName)
	if field == nil {
		return fmt.Errorf("字段 '%s' 不存在", fieldName)
	}
	
	// 验证字段值
	if err := field.ValidateValue(value); err != nil {
		return fmt.Errorf("字段值验证失败: %v", err)
	}
	
	// 更新字段值
	r.Data[fieldName] = value
	r.UpdatedBy = &updatedBy
	now := time.Now()
	r.LastModifiedTime = &now
	r.Version++
	
	// 更新哈希值
	r.updateHash()
	
	return nil
}

// GetFieldValue 获取字段值
func (r *Record) GetFieldValue(fieldName string) (interface{}, bool) {
	value, exists := r.Data[fieldName]
	return value, exists
}

// HasChanges 检查记录是否有变更
func (r *Record) HasChanges(otherHash string) bool {
	return r.Hash != otherHash
}

// Clone 克隆记录
func (r *Record) Clone() *Record {
	clonedData := make(map[string]interface{})
	for k, v := range r.Data {
		clonedData[k] = v
	}
	
	cloned := &Record{
		ID:               utils.GenerateRecordID(),
		TableID:          r.TableID,
		Data:             clonedData,
		Version:          r.Version,
		Hash:             r.Hash,
		CreatedBy:        r.CreatedBy,
		UpdatedBy:        r.UpdatedBy,
		CreatedTime:      r.CreatedTime,
		DeletedTime:      r.DeletedTime,
		LastModifiedTime: r.LastModifiedTime,
		tableSchema:      r.tableSchema,
	}
	
	return cloned
}

// SoftDelete 软删除记录
func (r *Record) SoftDelete() {
	now := time.Now()
	r.DeletedTime = &now
	r.LastModifiedTime = &now
	r.Version++
	r.updateHash()
}

// Restore 恢复已删除的记录
func (r *Record) Restore() {
	r.DeletedTime = nil
	now := time.Now()
	r.LastModifiedTime = &now
	r.Version++
	r.updateHash()
}

// IsDeleted 检查记录是否已删除
func (r *Record) IsDeleted() bool {
	return r.DeletedTime != nil
}

// updateHash 更新数据哈希值
func (r *Record) updateHash() {
	r.Hash = utils.GenerateDataHash(r.Data, r.Version)
}

// CreateChangeEvent 创建变更事件
func (r *Record) CreateChangeEvent(changeType string, oldData map[string]interface{}, changedBy string) *RecordChangeEvent {
	return &RecordChangeEvent{
		RecordID:   r.ID,
		TableID:    r.TableID,
		ChangeType: changeType,
		OldData:    oldData,
		NewData:    r.Data,
		ChangedBy:  changedBy,
		ChangedAt:  time.Now(),
		Version:    r.Version,
	}
}
