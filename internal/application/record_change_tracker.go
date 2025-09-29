package application

import (
	"context"
	"encoding/json"
	"time"

	"teable-go-backend/internal/domain/record"
	"teable-go-backend/pkg/logger"
	"teable-go-backend/pkg/utils"
)

// RecordChangeTracker 记录变更追踪器
type RecordChangeTracker struct {
	changeRepo ChangeRepository
}

// NewRecordChangeTracker 创建记录变更追踪器
func NewRecordChangeTracker() *RecordChangeTracker {
	return &RecordChangeTracker{
		// TODO: 注入实际的变更仓储实现
		changeRepo: &InMemoryChangeRepository{
			changes: make(map[string][]*record.RecordChangeEvent),
		},
	}
}

// TrackChange 追踪记录变更
func (t *RecordChangeTracker) TrackChange(ctx context.Context, event *record.RecordChangeEvent) error {
	// 设置事件ID和时间戳
	if event.ChangedAt.IsZero() {
		event.ChangedAt = time.Now()
	}

	// 计算变更差异
	if err := t.calculateChangeDiff(event); err != nil {
		logger.Error("计算变更差异失败", logger.ErrorField(err))
	}

	// 保存变更事件
	if err := t.changeRepo.SaveChange(ctx, event); err != nil {
		return err
	}

	logger.Info("记录变更追踪成功",
		logger.String("record_id", event.RecordID),
		logger.String("table_id", event.TableID),
		logger.String("change_type", event.ChangeType),
		logger.String("changed_by", event.ChangedBy),
	)

	return nil
}

// GetRecordChanges 获取记录的变更历史
func (t *RecordChangeTracker) GetRecordChanges(ctx context.Context, recordID string) ([]*record.RecordChangeEvent, error) {
	return t.changeRepo.GetRecordChanges(ctx, recordID)
}

// GetTableChanges 获取表格的变更历史
func (t *RecordChangeTracker) GetTableChanges(ctx context.Context, tableID string, limit int) ([]*record.RecordChangeEvent, error) {
	return t.changeRepo.GetTableChanges(ctx, tableID, limit)
}

// GetUserChanges 获取用户的变更历史
func (t *RecordChangeTracker) GetUserChanges(ctx context.Context, userID string, limit int) ([]*record.RecordChangeEvent, error) {
	return t.changeRepo.GetUserChanges(ctx, userID, limit)
}

// calculateChangeDiff 计算变更差异
func (t *RecordChangeTracker) calculateChangeDiff(event *record.RecordChangeEvent) error {
	if event.ChangeType == "create" {
		// 创建操作，所有字段都是新增
		return nil
	}

	if event.ChangeType == "delete" {
		// 删除操作，所有字段都被删除
		return nil
	}

	if event.ChangeType == "update" && event.OldData != nil && event.NewData != nil {
		// 更新操作，计算字段级别的差异
		diff := make(map[string]*FieldChange)

		// 检查修改和删除的字段
		for key, oldValue := range event.OldData {
			if newValue, exists := event.NewData[key]; exists {
				if !t.isEqual(oldValue, newValue) {
					diff[key] = &FieldChange{
						Field:    key,
						OldValue: oldValue,
						NewValue: newValue,
						Action:   "modified",
					}
				}
			} else {
				diff[key] = &FieldChange{
					Field:    key,
					OldValue: oldValue,
					NewValue: nil,
					Action:   "deleted",
				}
			}
		}

		// 检查新增的字段
		for key, newValue := range event.NewData {
			if _, exists := event.OldData[key]; !exists {
				diff[key] = &FieldChange{
					Field:    key,
					OldValue: nil,
					NewValue: newValue,
					Action:   "added",
				}
			}
		}

		// 将差异信息序列化并存储
		if len(diff) > 0 {
			if _, err := json.Marshal(diff); err == nil {
				// 可以将差异信息存储在事件的扩展字段中
				// 这里暂时不实现，留给具体的存储层处理
			}
		}
	}

	return nil
}

// isEqual 比较两个值是否相等
func (t *RecordChangeTracker) isEqual(a, b interface{}) bool {
	aJSON, aErr := json.Marshal(a)
	bJSON, bErr := json.Marshal(b)

	if aErr != nil || bErr != nil {
		return false
	}

	return string(aJSON) == string(bJSON)
}

// FieldChange 字段变更信息
type FieldChange struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
	Action   string      `json:"action"` // added, modified, deleted
}

// ChangeRepository 变更仓储接口
type ChangeRepository interface {
	SaveChange(ctx context.Context, event *record.RecordChangeEvent) error
	GetRecordChanges(ctx context.Context, recordID string) ([]*record.RecordChangeEvent, error)
	GetTableChanges(ctx context.Context, tableID string, limit int) ([]*record.RecordChangeEvent, error)
	GetUserChanges(ctx context.Context, userID string, limit int) ([]*record.RecordChangeEvent, error)
}

// InMemoryChangeRepository 内存变更仓储实现（用于测试和开发）
type InMemoryChangeRepository struct {
	changes map[string][]*record.RecordChangeEvent
}

func (r *InMemoryChangeRepository) SaveChange(ctx context.Context, event *record.RecordChangeEvent) error {
	// 生成事件ID
	eventID := utils.GenerateID()

	// 创建事件副本
	eventCopy := &record.RecordChangeEvent{
		RecordID:   event.RecordID,
		TableID:    event.TableID,
		ChangeType: event.ChangeType,
		OldData:    event.OldData,
		NewData:    event.NewData,
		ChangedBy:  event.ChangedBy,
		ChangedAt:  event.ChangedAt,
		Version:    event.Version,
	}

	// 存储到内存中
	if r.changes[event.RecordID] == nil {
		r.changes[event.RecordID] = make([]*record.RecordChangeEvent, 0)
	}
	r.changes[event.RecordID] = append(r.changes[event.RecordID], eventCopy)

	logger.Debug("变更事件已保存到内存",
		logger.String("event_id", eventID),
		logger.String("record_id", event.RecordID),
	)

	return nil
}

func (r *InMemoryChangeRepository) GetRecordChanges(ctx context.Context, recordID string) ([]*record.RecordChangeEvent, error) {
	changes := r.changes[recordID]
	if changes == nil {
		return []*record.RecordChangeEvent{}, nil
	}

	// 返回副本以避免外部修改
	result := make([]*record.RecordChangeEvent, len(changes))
	copy(result, changes)
	return result, nil
}

func (r *InMemoryChangeRepository) GetTableChanges(ctx context.Context, tableID string, limit int) ([]*record.RecordChangeEvent, error) {
	var result []*record.RecordChangeEvent

	for _, changes := range r.changes {
		for _, change := range changes {
			if change.TableID == tableID {
				result = append(result, change)
			}
		}
	}

	// 按时间倒序排序
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].ChangedAt.Before(result[j].ChangedAt) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	// 应用限制
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}

	return result, nil
}

func (r *InMemoryChangeRepository) GetUserChanges(ctx context.Context, userID string, limit int) ([]*record.RecordChangeEvent, error) {
	var result []*record.RecordChangeEvent

	for _, changes := range r.changes {
		for _, change := range changes {
			if change.ChangedBy == userID {
				result = append(result, change)
			}
		}
	}

	// 按时间倒序排序
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].ChangedAt.Before(result[j].ChangedAt) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	// 应用限制
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}

	return result, nil
}
