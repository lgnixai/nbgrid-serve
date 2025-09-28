package repository

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"teable-go-backend/internal/domain/record"
	"teable-go-backend/internal/infrastructure/database/models"
	"teable-go-backend/pkg/errors"
)

// RecordRepository 记录仓储实现
type RecordRepository struct {
	db *gorm.DB
}

// NewRecordRepository 创建新的记录仓储
func NewRecordRepository(db *gorm.DB) record.Repository {
	return &RecordRepository{db: db}
}

// Create 创建记录
func (r *RecordRepository) Create(ctx context.Context, rec *record.Record) error {
	model := r.domainToModel(rec)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return r.handleDBError(err)
	}
	return nil
}

// GetByID 通过ID获取记录
func (r *RecordRepository) GetByID(ctx context.Context, id string) (*record.Record, error) {
	var model models.Record
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, r.handleDBError(err)
	}
	return r.modelToDomain(&model), nil
}

// Update 更新记录
func (r *RecordRepository) Update(ctx context.Context, rec *record.Record) error {
	model := r.domainToModel(rec)
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return r.handleDBError(err)
	}
	return nil
}

// Delete 删除记录 (软删除)
func (r *RecordRepository) Delete(ctx context.Context, id string) error {
	// GORM的软删除会自动处理DeletedAt字段
	if err := r.db.WithContext(ctx).Delete(&models.Record{}, "id = ?", id).Error; err != nil {
		return r.handleDBError(err)
	}
	return nil
}

// List 列出记录
func (r *RecordRepository) List(ctx context.Context, filter record.ListRecordFilter) ([]*record.Record, error) {
	var modelRecords []models.Record
	query := r.db.WithContext(ctx).Model(&models.Record{})

	if filter.TableID != nil {
		query = query.Where("table_id = ?", *filter.TableID)
	}
	if filter.CreatedBy != nil {
		query = query.Where("created_by = ?", *filter.CreatedBy)
	}
	if filter.Search != "" {
		// 在JSON数据中搜索
		query = query.Where("data::text ILIKE ?", "%"+filter.Search+"%")
	}
	if filter.OrderBy != "" && filter.Order != "" {
		query = query.Order(fmt.Sprintf("%s %s", filter.OrderBy, filter.Order))
	} else {
		query = query.Order("created_time desc")
	}
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	if err := query.Find(&modelRecords).Error; err != nil {
		return nil, r.handleDBError(err)
	}

	records := make([]*record.Record, len(modelRecords))
	for i, model := range modelRecords {
		records[i] = r.modelToDomain(&model)
	}
	return records, nil
}

// Count 统计记录数量
func (r *RecordRepository) Count(ctx context.Context, filter record.ListRecordFilter) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Record{})

	if filter.TableID != nil {
		query = query.Where("table_id = ?", *filter.TableID)
	}
	if filter.CreatedBy != nil {
		query = query.Where("created_by = ?", *filter.CreatedBy)
	}
	if filter.Search != "" {
		// 在JSON数据中搜索
		query = query.Where("data::text ILIKE ?", "%"+filter.Search+"%")
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, r.handleDBError(err)
	}
	return count, nil
}

// Exists 检查记录是否存在
func (r *RecordRepository) Exists(ctx context.Context, filter record.ListRecordFilter) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Record{})

	if filter.TableID != nil {
		query = query.Where("table_id = ?", *filter.TableID)
	}
	if filter.CreatedBy != nil {
		query = query.Where("created_by = ?", *filter.CreatedBy)
	}

	if err := query.Count(&count).Error; err != nil {
		return false, r.handleDBError(err)
	}
	return count > 0, nil
}

// domainToModel 领域对象转数据模型
func (r *RecordRepository) domainToModel(rec *record.Record) *models.Record {
	model := &models.Record{
		ID:               rec.ID,
		TableID:          rec.TableID,
		CreatedBy:        rec.CreatedBy,
		CreatedTime:      rec.CreatedTime,
		LastModifiedTime: rec.LastModifiedTime,
	}

	// 转换数据
	if err := model.SetDataFromMap(rec.Data); err != nil {
		// 如果转换失败，设置为空JSON
		model.Data = "{}"
	}

	return model
}

// modelToDomain 数据模型转领域对象
func (r *RecordRepository) modelToDomain(model *models.Record) *record.Record {
	data, err := model.GetDataAsMap()
	if err != nil {
		// 如果转换失败，设置为空map
		data = make(map[string]interface{})
	}

	return &record.Record{
		ID:               model.ID,
		TableID:          model.TableID,
		Data:             data,
		CreatedBy:        model.CreatedBy,
		CreatedTime:      model.CreatedTime,
		LastModifiedTime: model.LastModifiedTime,
	}
}

// BulkCreate 批量创建记录
func (r *RecordRepository) BulkCreate(ctx context.Context, records []*record.Record) error {
	if len(records) == 0 {
		return nil
	}

	models := make([]models.Record, len(records))
	for i, rec := range records {
		models[i] = *r.domainToModel(rec)
	}

	if err := r.db.WithContext(ctx).CreateInBatches(models, 100).Error; err != nil {
		return r.handleDBError(err)
	}

	return nil
}

// BulkUpdate 批量更新记录
func (r *RecordRepository) BulkUpdate(ctx context.Context, updates map[string]map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	// 使用事务进行批量更新
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for recordID, data := range updates {
			// 将数据转换为JSON
			jsonData, err := json.Marshal(data)
			if err != nil {
				return err
			}

			// 更新记录
			if err := tx.Model(&models.Record{}).
				Where("id = ?", recordID).
				Updates(map[string]interface{}{
					"data":               string(jsonData),
					"last_modified_time": time.Now(),
				}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// BulkDelete 批量删除记录
func (r *RecordRepository) BulkDelete(ctx context.Context, recordIDs []string) error {
	if len(recordIDs) == 0 {
		return nil
	}

	// 软删除
	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.Record{}).
		Where("id IN ?", recordIDs).
		Updates(map[string]interface{}{
			"deleted_time":       now,
			"last_modified_time": now,
		}).Error
}

// ComplexQuery 复杂查询
func (r *RecordRepository) ComplexQuery(ctx context.Context, req record.ComplexQueryRequest) ([]map[string]interface{}, error) {
	query := r.db.WithContext(ctx).Model(&models.Record{}).
		Where("table_id = ?", req.TableID)

	// 构建查询条件
	for _, condition := range req.Conditions {
		query = r.buildCondition(query, condition)
	}

	// 分组
	if len(req.GroupBy) > 0 {
		query = query.Group(strings.Join(req.GroupBy, ", "))
	}

	// 聚合
	if len(req.Aggregations) > 0 {
		selectFields := make([]string, len(req.Aggregations))
		for i, agg := range req.Aggregations {
			selectFields[i] = fmt.Sprintf("%s(%s) as %s", agg.Type, agg.Field, agg.Alias)
		}
		query = query.Select(strings.Join(selectFields, ", "))
	}

	// 排序
	if len(req.OrderBy) > 0 {
		orderClauses := make([]string, len(req.OrderBy))
		for i, order := range req.OrderBy {
			orderClauses[i] = fmt.Sprintf("%s %s", order.Field, order.Order)
		}
		query = query.Order(strings.Join(orderClauses, ", "))
	}

	// 分页
	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}
	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	var results []map[string]interface{}
	if err := query.Find(&results).Error; err != nil {
		return nil, r.handleDBError(err)
	}

	return results, nil
}

// buildCondition 构建查询条件
func (r *RecordRepository) buildCondition(query *gorm.DB, condition record.QueryCondition) *gorm.DB {
	fieldPath := fmt.Sprintf("data->>'%s'", condition.Field)

	switch condition.Operator {
	case "eq":
		return query.Where(fmt.Sprintf("%s = ?", fieldPath), condition.Value)
	case "ne":
		return query.Where(fmt.Sprintf("%s != ?", fieldPath), condition.Value)
	case "gt":
		return query.Where(fmt.Sprintf("%s > ?", fieldPath), condition.Value)
	case "gte":
		return query.Where(fmt.Sprintf("%s >= ?", fieldPath), condition.Value)
	case "lt":
		return query.Where(fmt.Sprintf("%s < ?", fieldPath), condition.Value)
	case "lte":
		return query.Where(fmt.Sprintf("%s <= ?", fieldPath), condition.Value)
	case "like":
		return query.Where(fmt.Sprintf("%s LIKE ?", fieldPath), condition.Value)
	case "ilike":
		return query.Where(fmt.Sprintf("%s ILIKE ?", fieldPath), condition.Value)
	case "in":
		return query.Where(fmt.Sprintf("%s IN ?", fieldPath), condition.Value)
	case "not_in":
		return query.Where(fmt.Sprintf("%s NOT IN ?", fieldPath), condition.Value)
	case "between":
		if values, ok := condition.Value.([]interface{}); ok && len(values) == 2 {
			return query.Where(fmt.Sprintf("%s BETWEEN ? AND ?", fieldPath), values[0], values[1])
		}
	}

	return query
}

// GetRecordStats 获取记录统计信息
func (r *RecordRepository) GetRecordStats(ctx context.Context, tableID *string) (*record.RecordStats, error) {
	stats := &record.RecordStats{
		RecordsByTable: make(map[string]int64),
		RecordsByUser:  make(map[string]int64),
	}

	baseQuery := r.db.WithContext(ctx).Model(&models.Record{})
	if tableID != nil {
		baseQuery = baseQuery.Where("table_id = ?", *tableID)
	}

	// 总记录数
	if err := baseQuery.Count(&stats.TotalRecords).Error; err != nil {
		return nil, r.handleDBError(err)
	}

	// 活跃记录数（未删除）
	if err := baseQuery.Where("deleted_time IS NULL").Count(&stats.ActiveRecords).Error; err != nil {
		return nil, r.handleDBError(err)
	}

	// 已删除记录数
	stats.DeletedRecords = stats.TotalRecords - stats.ActiveRecords

	// 按表统计
	var tableStats []struct {
		TableID string `json:"table_id"`
		Count   int64  `json:"count"`
	}
	if err := r.db.WithContext(ctx).Model(&models.Record{}).
		Select("table_id, COUNT(*) as count").
		Group("table_id").
		Find(&tableStats).Error; err != nil {
		return nil, r.handleDBError(err)
	}

	for _, stat := range tableStats {
		stats.RecordsByTable[stat.TableID] = stat.Count
	}

	// 按用户统计
	var userStats []struct {
		CreatedBy string `json:"created_by"`
		Count     int64  `json:"count"`
	}
	if err := r.db.WithContext(ctx).Model(&models.Record{}).
		Select("created_by, COUNT(*) as count").
		Group("created_by").
		Find(&userStats).Error; err != nil {
		return nil, r.handleDBError(err)
	}

	for _, stat := range userStats {
		stats.RecordsByUser[stat.CreatedBy] = stat.Count
	}

	// 最近7天活动
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	if err := baseQuery.Where("created_time >= ?", sevenDaysAgo).Count(&stats.RecentActivity).Error; err != nil {
		return nil, r.handleDBError(err)
	}

	return stats, nil
}

// ExportRecords 导出记录
func (r *RecordRepository) ExportRecords(ctx context.Context, req record.ExportRequest) ([]byte, error) {
	// 构建查询
	query := r.db.WithContext(ctx).Model(&models.Record{})

	if req.TableID != nil {
		query = query.Where("table_id = ?", *req.TableID)
	}

	// 应用字段过滤
	for field, value := range req.FieldFilters {
		fieldPath := fmt.Sprintf("data->>'%s'", field)
		query = query.Where(fmt.Sprintf("%s = ?", fieldPath), value)
	}

	// 应用日期范围过滤
	if req.DateRange != nil {
		if req.DateRange.StartTime != nil {
			query = query.Where("created_time >= ?", *req.DateRange.StartTime)
		}
		if req.DateRange.EndTime != nil {
			query = query.Where("created_time <= ?", *req.DateRange.EndTime)
		}
	}

	var records []models.Record
	if err := query.Find(&records).Error; err != nil {
		return nil, r.handleDBError(err)
	}

	// 转换为导出格式
	switch req.Format {
	case "json":
		return json.Marshal(records)
	case "csv":
		return r.exportToCSV(records, req.Fields)
	default:
		return nil, errors.ErrInvalidRequest.WithDetails(fmt.Sprintf("不支持的导出格式: %s", req.Format))
	}
}

// exportToCSV 导出为CSV格式
func (r *RecordRepository) exportToCSV(records []models.Record, fields []string) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// 写入标题行
	if len(fields) > 0 {
		writer.Write(fields)
	} else {
		// 默认字段
		writer.Write([]string{"id", "table_id", "created_by", "created_time", "data"})
	}

	// 写入数据行
	for _, record := range records {
		row := []string{
			record.ID,
			record.TableID,
			record.CreatedBy,
			record.CreatedTime.Format(time.RFC3339),
			record.Data,
		}
		writer.Write(row)
	}

	writer.Flush()
	return buf.Bytes(), nil
}

// ImportRecords 导入记录
func (r *RecordRepository) ImportRecords(ctx context.Context, req record.ImportRequest) (int, error) {
	var records []models.Record

	switch req.Format {
	case "json":
		if err := json.Unmarshal(req.Data.([]byte), &records); err != nil {
			return 0, errors.ErrInvalidRequest.WithDetails("JSON格式错误: " + err.Error())
		}
	case "csv":
		// TODO: 实现CSV解析
		return 0, errors.ErrNotImplemented.WithDetails("CSV导入功能暂未实现")
	default:
		return 0, errors.ErrInvalidRequest.WithDetails(fmt.Sprintf("不支持的导入格式: %s", req.Format))
	}

	// 设置表ID
	for i := range records {
		records[i].TableID = req.TableID
	}

	// 批量创建
	if err := r.db.WithContext(ctx).CreateInBatches(records, 100).Error; err != nil {
		return 0, r.handleDBError(err)
	}

	return len(records), nil
}

// handleDBError 处理数据库错误
func (r *RecordRepository) handleDBError(err error) error {
	// TODO: 根据具体的数据库错误类型返回对应的业务错误
	if strings.Contains(err.Error(), "duplicate key") {
		return errors.ErrResourceExists.WithDetails("记录已存在")
	}

	return errors.ErrDatabaseOperation.WithDetails(err.Error())
}
