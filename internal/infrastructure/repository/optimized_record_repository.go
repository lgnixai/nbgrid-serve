//go:build ignore

package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"teable-go-backend/internal/domain/record"
	"teable-go-backend/internal/infrastructure/database/models"
)

// OptimizedRecordRepository 优化的记录仓储实现
type OptimizedRecordRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewOptimizedRecordRepository 创建优化的记录仓储
func NewOptimizedRecordRepository(db *gorm.DB, logger *zap.Logger) record.Repository {
	return &OptimizedRecordRepository{
		db:     db,
		logger: logger,
	}
}

// Create 创建记录（优化版）
func (r *OptimizedRecordRepository) Create(ctx context.Context, record *record.Record) error {
	model := r.domainToModel(record)

	// 使用 Create 的批量插入优化
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		r.logger.Error("Failed to create record",
			zap.String("record_id", record.ID),
			zap.Error(err),
		)
		return err
	}

	// 更新领域对象
	record.ID = model.ID
	record.CreatedTime = model.CreatedTime
	record.LastModifiedTime = model.LastModifiedTime

	return nil
}

// GetByID 根据ID获取记录（优化版）
func (r *OptimizedRecordRepository) GetByID(ctx context.Context, id string) (*record.Record, error) {
	var model models.Record

	// 使用预加载减少查询次数
	query := r.db.WithContext(ctx).
		Preload("Fields").
		Preload("Creator").
		Where("id = ? AND deleted_time IS NULL", id)

	if err := query.First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, record.ErrRecordNotFound
		}
		r.logger.Error("Failed to get record by ID",
			zap.String("record_id", id),
			zap.Error(err),
		)
		return err
	}

	return r.modelToDomain(&model), nil
}

// GetByTableID 根据表ID获取记录列表（优化版）
func (r *OptimizedRecordRepository) GetByTableID(ctx context.Context, tableID string, offset, limit int) ([]*record.Record, error) {
	var models []models.Record

	// 优化查询：使用索引，避免 N+1 问题
	query := r.db.WithContext(ctx).
		Preload("Fields", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
		Where("table_id = ? AND deleted_time IS NULL", tableID).
		Order("created_time DESC").
		Offset(offset).
		Limit(limit)

	// 添加查询提示以使用特定索引
	query = query.Clauses(clause.Index{Name: "idx_records_table_created"})

	if err := query.Find(&models).Error; err != nil {
		r.logger.Error("Failed to get records by table ID",
			zap.String("table_id", tableID),
			zap.Error(err),
		)
		return nil, err
	}

	// 批量转换，减少内存分配
	records := make([]*record.Record, 0, len(models))
	for i := range models {
		records = append(records, r.modelToDomain(&models[i]))
	}

	return records, nil
}

// BatchCreate 批量创建记录（优化版）
func (r *OptimizedRecordRepository) BatchCreate(ctx context.Context, records []*record.Record) error {
	if len(records) == 0 {
		return nil
	}

	// 分批处理，避免单次插入数据过多
	batchSize := 1000
	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}

		batch := records[i:end]
		models := make([]*models.Record, len(batch))

		for j, rec := range batch {
			models[j] = r.domainToModel(rec)
		}

		// 使用 CreateInBatches 进行批量插入
		if err := r.db.WithContext(ctx).CreateInBatches(models, 100).Error; err != nil {
			r.logger.Error("Failed to batch create records",
				zap.Int("batch_start", i),
				zap.Int("batch_end", end),
				zap.Error(err),
			)
			return err
		}

		// 更新领域对象
		for j, model := range models {
			batch[j].ID = model.ID
			batch[j].CreatedTime = model.CreatedTime
			batch[j].LastModifiedTime = model.LastModifiedTime
		}
	}

	return nil
}

// BatchUpdate 批量更新记录（优化版）
func (r *OptimizedRecordRepository) BatchUpdate(ctx context.Context, records []*record.Record) error {
	if len(records) == 0 {
		return nil
	}

	// 使用事务确保数据一致性
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()

		// 批量更新，减少数据库往返
		for _, rec := range records {
			updates := map[string]interface{}{
				"data":               rec.Data,
				"last_modified_by":   rec.LastModifiedBy,
				"last_modified_time": now,
			}

			if err := tx.Model(&models.Record{}).
				Where("id = ? AND deleted_time IS NULL", rec.ID).
				Updates(updates).Error; err != nil {
				return err
			}

			// 更新领域对象
			rec.LastModifiedTime = &now
		}

		return nil
	})
}

// ComplexQuery 复杂查询（优化版）
func (r *OptimizedRecordRepository) ComplexQuery(ctx context.Context, query record.Query) ([]*record.Record, int64, error) {
	// 构建基础查询
	baseQuery := r.db.WithContext(ctx).Model(&models.Record{}).
		Where("table_id = ? AND deleted_time IS NULL", query.TableID)

	// 应用过滤条件
	for _, filter := range query.Filters {
		baseQuery = r.applyFilter(baseQuery, filter)
	}

	// 计数查询（优化：使用单独的查询避免加载不必要的数据）
	var total int64
	countQuery := baseQuery
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 如果没有数据，直接返回
	if total == 0 {
		return []*record.Record{}, 0, nil
	}

	// 数据查询
	var models []models.Record
	dataQuery := baseQuery.
		Preload("Fields").
		Order(r.buildOrderClause(query.Sort)).
		Offset(query.Offset).
		Limit(query.Limit)

	if err := dataQuery.Find(&models).Error; err != nil {
		return nil, 0, err
	}

	// 转换结果
	records := make([]*record.Record, 0, len(models))
	for i := range models {
		records = append(records, r.modelToDomain(&models[i]))
	}

	return records, total, nil
}

// applyFilter 应用过滤条件
func (r *OptimizedRecordRepository) applyFilter(query *gorm.DB, filter record.Filter) *gorm.DB {
	// 使用 JSONB 操作符进行高效查询
	switch filter.Operator {
	case "=":
		return query.Where("data->? = ?", filter.Field, filter.Value)
	case "!=":
		return query.Where("data->? != ?", filter.Field, filter.Value)
	case ">":
		return query.Where("(data->>?)::numeric > ?", filter.Field, filter.Value)
	case "<":
		return query.Where("(data->>?)::numeric < ?", filter.Field, filter.Value)
	case ">=":
		return query.Where("(data->>?)::numeric >= ?", filter.Field, filter.Value)
	case "<=":
		return query.Where("(data->>?)::numeric <= ?", filter.Field, filter.Value)
	case "like":
		return query.Where("data->>? LIKE ?", filter.Field, fmt.Sprintf("%%%v%%", filter.Value))
	case "in":
		return query.Where("data->? IN ?", filter.Field, filter.Value)
	case "not_in":
		return query.Where("data->? NOT IN ?", filter.Field, filter.Value)
	case "is_null":
		return query.Where("data->? IS NULL", filter.Field)
	case "is_not_null":
		return query.Where("data->? IS NOT NULL", filter.Field)
	default:
		return query
	}
}

// buildOrderClause 构建排序子句
func (r *OptimizedRecordRepository) buildOrderClause(sort []record.Sort) string {
	if len(sort) == 0 {
		return "created_time DESC"
	}

	var clauses []string
	for _, s := range sort {
		direction := "ASC"
		if s.Desc {
			direction = "DESC"
		}

		// 对 JSONB 字段进行排序
		clause := fmt.Sprintf("data->>'%s' %s", s.Field, direction)
		clauses = append(clauses, clause)
	}

	return strings.Join(clauses, ", ")
}

// OptimizeTable 优化表（创建索引等）
func (r *OptimizedRecordRepository) OptimizeTable(ctx context.Context, tableID string) error {
	// 创建常用查询的索引
	indexes := []string{
		// 复合索引：表ID + 创建时间
		`CREATE INDEX IF NOT EXISTS idx_records_table_created ON records(table_id, created_time DESC) WHERE deleted_time IS NULL`,

		// 复合索引：表ID + 修改时间
		`CREATE INDEX IF NOT EXISTS idx_records_table_modified ON records(table_id, last_modified_time DESC) WHERE deleted_time IS NULL`,

		// JSONB GIN 索引：加速 JSONB 查询
		`CREATE INDEX IF NOT EXISTS idx_records_data_gin ON records USING gin(data)`,

		// 部分索引：只索引未删除的记录
		`CREATE INDEX IF NOT EXISTS idx_records_active ON records(table_id) WHERE deleted_time IS NULL`,
	}

	for _, index := range indexes {
		if err := r.db.WithContext(ctx).Exec(index).Error; err != nil {
			r.logger.Warn("Failed to create index",
				zap.String("index", index),
				zap.Error(err),
			)
		}
	}

	// 更新表统计信息
	if err := r.db.WithContext(ctx).Exec("ANALYZE records").Error; err != nil {
		r.logger.Warn("Failed to analyze table", zap.Error(err))
	}

	return nil
}

// 其他必需的接口方法实现...

// Update 更新记录
func (r *OptimizedRecordRepository) Update(ctx context.Context, record *record.Record) error {
	model := r.domainToModel(record)
	model.LastModifiedTime = timePtr(time.Now())

	result := r.db.WithContext(ctx).
		Model(&models.Record{}).
		Where("id = ? AND deleted_time IS NULL", record.ID).
		Updates(model)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return record.ErrRecordNotFound
	}

	record.LastModifiedTime = model.LastModifiedTime
	return nil
}

// Delete 删除记录
func (r *OptimizedRecordRepository) Delete(ctx context.Context, id string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&models.Record{}).
		Where("id = ? AND deleted_time IS NULL", id).
		Update("deleted_time", now)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return record.ErrRecordNotFound
	}

	return nil
}

// CountByTableID 统计表中的记录数
func (r *OptimizedRecordRepository) CountByTableID(ctx context.Context, tableID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Record{}).
		Where("table_id = ? AND deleted_time IS NULL", tableID).
		Count(&count).Error

	return count, err
}

// BatchDelete 批量删除
func (r *OptimizedRecordRepository) BatchDelete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.Record{}).
		Where("id IN ? AND deleted_time IS NULL", ids).
		Update("deleted_time", now).Error
}

// 辅助方法

func (r *OptimizedRecordRepository) domainToModel(domain *record.Record) *models.Record {
	return &models.Record{
		ID:               domain.ID,
		TableID:          domain.TableID,
		Data:             domain.Data,
		CreatedBy:        domain.CreatedBy,
		CreatedTime:      domain.CreatedTime,
		LastModifiedBy:   domain.LastModifiedBy,
		LastModifiedTime: domain.LastModifiedTime,
		DeletedTime:      domain.DeletedTime,
	}
}

func (r *OptimizedRecordRepository) modelToDomain(model *models.Record) *record.Record {
	return &record.Record{
		ID:               model.ID,
		TableID:          model.TableID,
		Data:             model.Data,
		CreatedBy:        model.CreatedBy,
		CreatedTime:      model.CreatedTime,
		LastModifiedBy:   model.LastModifiedBy,
		LastModifiedTime: model.LastModifiedTime,
		DeletedTime:      model.DeletedTime,
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}
