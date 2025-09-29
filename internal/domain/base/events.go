package base

import (
	"time"
)

// 基础表相关事件类型常量
const (
	BaseCreatedEventType   = "base.created"
	BaseUpdatedEventType   = "base.updated"
	BaseDeletedEventType   = "base.deleted"
	BaseArchivedEventType  = "base.archived"
	BaseRestoredEventType  = "base.restored"
	TableAddedEventType    = "base.table_added"
	TableRemovedEventType  = "base.table_removed"
)

// BaseDomainEvent 基础领域事件
type BaseDomainEvent struct {
	ID          string      `json:"id"`
	Type        string      `json:"type"`
	AggregateId string      `json:"aggregate_id"`
	OccurredOn  time.Time   `json:"occurred_on"`
	Data        interface{} `json:"data"`
}

func (e BaseDomainEvent) EventID() string {
	return e.ID
}

func (e BaseDomainEvent) EventType() string {
	return e.Type
}

func (e BaseDomainEvent) AggregateID() string {
	return e.AggregateId
}

func (e BaseDomainEvent) OccurredAt() time.Time {
	return e.OccurredOn
}

func (e BaseDomainEvent) EventData() interface{} {
	return e.Data
}

// BaseCreatedEvent 基础表创建事件
type BaseCreatedEvent struct {
	BaseDomainEvent
	BaseID      string `json:"base_id"`
	SpaceID     string `json:"space_id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CreatedBy   string `json:"created_by"`
}

// NewBaseCreatedEvent 创建基础表创建事件
func NewBaseCreatedEvent(base *Base) *BaseCreatedEvent {
	var description string
	if base.Description != nil {
		description = *base.Description
	}

	return &BaseCreatedEvent{
		BaseDomainEvent: BaseDomainEvent{
			ID:          generateEventID(),
			Type:        BaseCreatedEventType,
			AggregateId: base.ID,
			OccurredOn:  time.Now(),
		},
		BaseID:      base.ID,
		SpaceID:     base.SpaceID,
		Name:        base.Name,
		Description: description,
		CreatedBy:   base.CreatedBy,
	}
}

// BaseUpdatedEvent 基础表更新事件
type BaseUpdatedEvent struct {
	BaseDomainEvent
	BaseID       string                 `json:"base_id"`
	Changes      map[string]interface{} `json:"changes"`
	UpdatedBy    string                 `json:"updated_by,omitempty"`
	PreviousData map[string]interface{} `json:"previous_data,omitempty"`
}

// NewBaseUpdatedEvent 创建基础表更新事件
func NewBaseUpdatedEvent(baseID string, changes map[string]interface{}, updatedBy string) *BaseUpdatedEvent {
	return &BaseUpdatedEvent{
		BaseDomainEvent: BaseDomainEvent{
			ID:          generateEventID(),
			Type:        BaseUpdatedEventType,
			AggregateId: baseID,
			OccurredOn:  time.Now(),
		},
		BaseID:    baseID,
		Changes:   changes,
		UpdatedBy: updatedBy,
	}
}

// BaseDeletedEvent 基础表删除事件
type BaseDeletedEvent struct {
	BaseDomainEvent
	BaseID    string `json:"base_id"`
	SpaceID   string `json:"space_id"`
	Name      string `json:"name"`
	DeletedBy string `json:"deleted_by,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

// NewBaseDeletedEvent 创建基础表删除事件
func NewBaseDeletedEvent(base *Base, deletedBy, reason string) *BaseDeletedEvent {
	return &BaseDeletedEvent{
		BaseDomainEvent: BaseDomainEvent{
			ID:          generateEventID(),
			Type:        BaseDeletedEventType,
			AggregateId: base.ID,
			OccurredOn:  time.Now(),
		},
		BaseID:    base.ID,
		SpaceID:   base.SpaceID,
		Name:      base.Name,
		DeletedBy: deletedBy,
		Reason:    reason,
	}
}

// BaseArchivedEvent 基础表归档事件
type BaseArchivedEvent struct {
	BaseDomainEvent
	BaseID     string `json:"base_id"`
	SpaceID    string `json:"space_id"`
	Name       string `json:"name"`
	ArchivedBy string `json:"archived_by,omitempty"`
}

// NewBaseArchivedEvent 创建基础表归档事件
func NewBaseArchivedEvent(base *Base, archivedBy string) *BaseArchivedEvent {
	return &BaseArchivedEvent{
		BaseDomainEvent: BaseDomainEvent{
			ID:          generateEventID(),
			Type:        BaseArchivedEventType,
			AggregateId: base.ID,
			OccurredOn:  time.Now(),
		},
		BaseID:     base.ID,
		SpaceID:    base.SpaceID,
		Name:       base.Name,
		ArchivedBy: archivedBy,
	}
}

// BaseRestoredEvent 基础表恢复事件
type BaseRestoredEvent struct {
	BaseDomainEvent
	BaseID     string `json:"base_id"`
	SpaceID    string `json:"space_id"`
	Name       string `json:"name"`
	RestoredBy string `json:"restored_by,omitempty"`
}

// NewBaseRestoredEvent 创建基础表恢复事件
func NewBaseRestoredEvent(base *Base, restoredBy string) *BaseRestoredEvent {
	return &BaseRestoredEvent{
		BaseDomainEvent: BaseDomainEvent{
			ID:          generateEventID(),
			Type:        BaseRestoredEventType,
			AggregateId: base.ID,
			OccurredOn:  time.Now(),
		},
		BaseID:     base.ID,
		SpaceID:    base.SpaceID,
		Name:       base.Name,
		RestoredBy: restoredBy,
	}
}

// TableAddedEvent 表格添加事件
type TableAddedEvent struct {
	BaseDomainEvent
	BaseID    string `json:"base_id"`
	TableID   string `json:"table_id"`
	TableName string `json:"table_name"`
	CreatedBy string `json:"created_by"`
}

// NewTableAddedEvent 创建表格添加事件
func NewTableAddedEvent(table *Table) *TableAddedEvent {
	return &TableAddedEvent{
		BaseDomainEvent: BaseDomainEvent{
			ID:          generateEventID(),
			Type:        TableAddedEventType,
			AggregateId: table.BaseID,
			OccurredOn:  time.Now(),
		},
		BaseID:    table.BaseID,
		TableID:   table.ID,
		TableName: table.Name,
		CreatedBy: table.CreatedBy,
	}
}

// TableRemovedEvent 表格移除事件
type TableRemovedEvent struct {
	BaseDomainEvent
	BaseID    string `json:"base_id"`
	TableID   string `json:"table_id"`
	TableName string `json:"table_name"`
	RemovedBy string `json:"removed_by,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

// NewTableRemovedEvent 创建表格移除事件
func NewTableRemovedEvent(table *Table, removedBy, reason string) *TableRemovedEvent {
	return &TableRemovedEvent{
		BaseDomainEvent: BaseDomainEvent{
			ID:          generateEventID(),
			Type:        TableRemovedEventType,
			AggregateId: table.BaseID,
			OccurredOn:  time.Now(),
		},
		BaseID:    table.BaseID,
		TableID:   table.ID,
		TableName: table.Name,
		RemovedBy: removedBy,
		Reason:    reason,
	}
}

// generateEventID 生成事件ID
func generateEventID() string {
	return "evt_" + time.Now().Format("20060102150405") + "_" + generateRandomString(8)
}

// generateRandomString 生成随机字符串
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)] // 简化实现
	}
	return string(b)
}