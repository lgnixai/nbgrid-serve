package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"teable-go-backend/internal/domain/record"
	"teable-go-backend/internal/infrastructure/database/models"
	"teable-go-backend/pkg/utils"
)

// ChangeRepository 记录变更仓储实现
type ChangeRepository struct {
	db *gorm.DB
}

// NewChangeRepository 创建记录变更仓储
func NewChangeRepository(db *gorm.DB) *ChangeRepository {
	return &ChangeRepository{
		db: db,
	}
}

// SaveChange 保存变更事件
func (r *ChangeRepository) SaveChange(ctx context.Context, event *record.RecordChangeEvent) error {
	change := &models.RecordChange{
		ID:         utils.GenerateID(),
		RecordID:   event.RecordID,
		TableID:    event.TableID,
		ChangeType: event.ChangeType,
		OldData:    event.OldData,
		NewData:    event.NewData,
		ChangedBy:  event.ChangedBy,
		ChangedAt:  event.ChangedAt,
		Version:    event.Version,
	}

	if change.ChangedAt.IsZero() {
		change.ChangedAt = time.Now()
	}

	if err := r.db.WithContext(ctx).Create(change).Error; err != nil {
		return fmt.Errorf("failed to save change: %w", err)
	}

	return nil
}

// GetRecordChanges 获取记录的变更历史
func (r *ChangeRepository) GetRecordChanges(ctx context.Context, recordID string) ([]*record.RecordChangeEvent, error) {
	var changes []models.RecordChange
	
	if err := r.db.WithContext(ctx).
		Where("record_id = ?", recordID).
		Order("changed_at DESC").
		Find(&changes).Error; err != nil {
		return nil, fmt.Errorf("failed to get record changes: %w", err)
	}

	return r.toChangeEvents(changes), nil
}

// GetTableChanges 获取表格的变更历史
func (r *ChangeRepository) GetTableChanges(ctx context.Context, tableID string, limit int) ([]*record.RecordChangeEvent, error) {
	var changes []models.RecordChange
	
	query := r.db.WithContext(ctx).
		Where("table_id = ?", tableID).
		Order("changed_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if err := query.Find(&changes).Error; err != nil {
		return nil, fmt.Errorf("failed to get table changes: %w", err)
	}

	return r.toChangeEvents(changes), nil
}

// GetUserChanges 获取用户的变更历史
func (r *ChangeRepository) GetUserChanges(ctx context.Context, userID string, limit int) ([]*record.RecordChangeEvent, error) {
	var changes []models.RecordChange
	
	query := r.db.WithContext(ctx).
		Where("changed_by = ?", userID).
		Order("changed_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if err := query.Find(&changes).Error; err != nil {
		return nil, fmt.Errorf("failed to get user changes: %w", err)
	}

	return r.toChangeEvents(changes), nil
}

// toChangeEvents 将模型转换为领域事件
func (r *ChangeRepository) toChangeEvents(changes []models.RecordChange) []*record.RecordChangeEvent {
	events := make([]*record.RecordChangeEvent, len(changes))
	for i, change := range changes {
		events[i] = &record.RecordChangeEvent{
			RecordID:   change.RecordID,
			TableID:    change.TableID,
			ChangeType: change.ChangeType,
			OldData:    change.OldData,
			NewData:    change.NewData,
			ChangedBy:  change.ChangedBy,
			ChangedAt:  change.ChangedAt,
			Version:    change.Version,
		}
	}
	return events
}

// CleanupOldChanges 清理旧的变更记录
func (r *ChangeRepository) CleanupOldChanges(ctx context.Context, retentionDays int) error {
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)
	
	result := r.db.WithContext(ctx).
		Where("changed_at < ?", cutoffTime).
		Delete(&models.RecordChange{})
	
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup old changes: %w", result.Error)
	}
	
	return nil
}