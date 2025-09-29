package base

import (
	"time"
)

// BaseAggregate 基础表聚合根
type BaseAggregate struct {
	*Base
	tables  map[string]*Table // 表格集合
	metrics *BaseMetrics      // 基础表指标
	events  []DomainEvent     // 领域事件
}

// Table 表格实体（简化版本）
type Table struct {
	ID          string    `json:"id"`
	BaseID      string    `json:"base_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	CreatedBy   string    `json:"created_by"`
	CreatedTime time.Time `json:"created_time"`
}

// BaseMetrics 基础表指标
type BaseMetrics struct {
	TotalTables      int64     `json:"total_tables"`
	TotalRecords     int64     `json:"total_records"`
	TotalFields      int64     `json:"total_fields"`
	StorageUsed      int64     `json:"storage_used"`      // 字节
	LastActivityAt   time.Time `json:"last_activity_at"`
	CreatedAt        time.Time `json:"created_at"`
}

// DomainEvent 领域事件接口
type DomainEvent interface {
	EventID() string
	EventType() string
	AggregateID() string
	OccurredAt() time.Time
	EventData() interface{}
}

// NewBaseAggregate 创建基础表聚合根
func NewBaseAggregate(base *Base) *BaseAggregate {
	return &BaseAggregate{
		Base:    base,
		tables:  make(map[string]*Table),
		metrics: NewBaseMetrics(),
		events:  make([]DomainEvent, 0),
	}
}

// NewBaseMetrics 创建基础表指标
func NewBaseMetrics() *BaseMetrics {
	return &BaseMetrics{
		TotalTables:    0,
		TotalRecords:   0,
		TotalFields:    0,
		StorageUsed:    0,
		LastActivityAt: time.Now(),
		CreatedAt:      time.Now(),
	}
}

// GetUncommittedEvents 获取未提交的事件
func (a *BaseAggregate) GetUncommittedEvents() []DomainEvent {
	return a.events
}

// MarkEventsAsCommitted 标记事件为已提交
func (a *BaseAggregate) MarkEventsAsCommitted() {
	a.events = make([]DomainEvent, 0)
}

// addEvent 添加领域事件
func (a *BaseAggregate) addEvent(event DomainEvent) {
	a.events = append(a.events, event)
}

// CreateBase 创建基础表（聚合根方法）
func (a *BaseAggregate) CreateBase(spaceID, name, createdBy string, description, icon *string) error {
	// 创建基础表实体
	base, err := NewBase(spaceID, name, createdBy)
	if err != nil {
		return err
	}

	// 设置可选属性
	if description != nil {
		base.Description = description
	}
	if icon != nil {
		base.Icon = icon
	}

	a.Base = base
	a.metrics = NewBaseMetrics()

	// 发布基础表创建事件
	event := NewBaseCreatedEvent(a.Base)
	a.addEvent(event)

	return nil
}

// UpdateBase 更新基础表信息（聚合根方法）
func (a *BaseAggregate) UpdateBase(name, description, icon *string, updatedBy string) error {
	// 验证基础表是否可以被更新
	if err := a.Base.ValidateForUpdate(); err != nil {
		return err
	}

	// 记录变更前的数据
	previousData := map[string]interface{}{
		"name":        a.Base.Name,
		"description": a.Base.Description,
		"icon":        a.Base.Icon,
	}

	// 更新基础表信息
	if err := a.Base.Update(name, description, icon); err != nil {
		return err
	}

	// 记录变更
	changes := make(map[string]interface{})
	if name != nil {
		changes["name"] = *name
	}
	if description != nil {
		changes["description"] = *description
	}
	if icon != nil {
		changes["icon"] = *icon
	}

	// 发布基础表更新事件
	event := NewBaseUpdatedEvent(a.Base.ID, changes, updatedBy)
	event.PreviousData = previousData
	a.addEvent(event)

	// 更新活动时间
	a.metrics.UpdateActivity()

	return nil
}

// DeleteBase 删除基础表（聚合根方法）
func (a *BaseAggregate) DeleteBase(deletedBy, reason string) error {
	// 验证基础表是否可以被删除
	if err := a.Base.ValidateForDeletion(); err != nil {
		return err
	}

	a.Base.SoftDelete()

	// 发布基础表删除事件
	event := NewBaseDeletedEvent(a.Base, deletedBy, reason)
	a.addEvent(event)

	return nil
}

// ArchiveBase 归档基础表（聚合根方法）
func (a *BaseAggregate) ArchiveBase(archivedBy string) error {
	if err := a.Base.Archive(); err != nil {
		return err
	}

	// 发布基础表归档事件
	event := NewBaseArchivedEvent(a.Base, archivedBy)
	a.addEvent(event)

	return nil
}

// RestoreBase 恢复基础表（聚合根方法）
func (a *BaseAggregate) RestoreBase(restoredBy string) error {
	if err := a.Base.Restore(); err != nil {
		return err
	}

	// 发布基础表恢复事件
	event := NewBaseRestoredEvent(a.Base, restoredBy)
	a.addEvent(event)

	return nil
}

// AddTable 添加表格（聚合根方法）
func (a *BaseAggregate) AddTable(table *Table) error {
	// 检查表格是否已存在
	if _, exists := a.tables[table.ID]; exists {
		return DomainError{Code: "TABLE_EXISTS", Message: "table already exists"}
	}

	// 添加表格
	a.tables[table.ID] = table

	// 更新统计信息
	a.metrics.AddTable()
	a.Base.UpdateTableCount(len(a.tables))

	// 发布表格添加事件
	event := NewTableAddedEvent(table)
	a.addEvent(event)

	return nil
}

// RemoveTable 移除表格（聚合根方法）
func (a *BaseAggregate) RemoveTable(tableID, removedBy, reason string) error {
	table, exists := a.tables[tableID]
	if !exists {
		return DomainError{Code: "TABLE_NOT_FOUND", Message: "table not found"}
	}

	// 删除表格
	delete(a.tables, tableID)

	// 更新统计信息
	a.metrics.RemoveTable()
	a.Base.UpdateTableCount(len(a.tables))

	// 发布表格移除事件
	event := NewTableRemovedEvent(table, removedBy, reason)
	a.addEvent(event)

	return nil
}

// GetTables 获取表格列表
func (a *BaseAggregate) GetTables() []*Table {
	tables := make([]*Table, 0, len(a.tables))
	for _, table := range a.tables {
		tables = append(tables, table)
	}
	return tables
}

// GetTable 获取指定表格
func (a *BaseAggregate) GetTable(tableID string) *Table {
	return a.tables[tableID]
}

// GetMetrics 获取基础表指标
func (a *BaseAggregate) GetMetrics() *BaseMetrics {
	return a.metrics
}

// LoadTables 加载表格（从仓储重构时使用）
func (a *BaseAggregate) LoadTables(tables []*Table) {
	a.tables = make(map[string]*Table)
	for _, table := range tables {
		a.tables[table.ID] = table
	}
	a.Base.UpdateTableCount(len(a.tables))
}

// LoadMetrics 加载指标（从仓储重构时使用）
func (a *BaseAggregate) LoadMetrics(metrics *BaseMetrics) {
	if metrics != nil {
		a.metrics = metrics
	} else {
		a.metrics = NewBaseMetrics()
	}
}

// GetBase 获取基础表实体
func (a *BaseAggregate) GetBase() *Base {
	return a.Base
}

// GetID 获取聚合根ID
func (a *BaseAggregate) GetID() string {
	if a.Base == nil {
		return ""
	}
	return a.Base.ID
}

// GetVersion 获取聚合根版本（基于最后修改时间）
func (a *BaseAggregate) GetVersion() int64 {
	if a.Base == nil || a.Base.LastModifiedTime == nil {
		return 0
	}
	return a.Base.LastModifiedTime.Unix()
}

// IsDeleted 检查聚合根是否已删除
func (a *BaseAggregate) IsDeleted() bool {
	if a.Base == nil {
		return true
	}
	return a.Base.IsDeleted()
}

// UpdateActivity 更新活动时间
func (bm *BaseMetrics) UpdateActivity() {
	bm.LastActivityAt = time.Now()
}

// AddTable 增加表格数量
func (bm *BaseMetrics) AddTable() {
	bm.TotalTables++
	bm.UpdateActivity()
}

// RemoveTable 减少表格数量
func (bm *BaseMetrics) RemoveTable() {
	if bm.TotalTables > 0 {
		bm.TotalTables--
	}
	bm.UpdateActivity()
}

// AddRecord 增加记录数量
func (bm *BaseMetrics) AddRecord() {
	bm.TotalRecords++
	bm.UpdateActivity()
}

// RemoveRecord 减少记录数量
func (bm *BaseMetrics) RemoveRecord() {
	if bm.TotalRecords > 0 {
		bm.TotalRecords--
	}
	bm.UpdateActivity()
}

// AddField 增加字段数量
func (bm *BaseMetrics) AddField() {
	bm.TotalFields++
	bm.UpdateActivity()
}

// RemoveField 减少字段数量
func (bm *BaseMetrics) RemoveField() {
	if bm.TotalFields > 0 {
		bm.TotalFields--
	}
	bm.UpdateActivity()
}

// AddStorageUsage 增加存储使用量
func (bm *BaseMetrics) AddStorageUsage(bytes int64) {
	bm.StorageUsed += bytes
	bm.UpdateActivity()
}

// RemoveStorageUsage 减少存储使用量
func (bm *BaseMetrics) RemoveStorageUsage(bytes int64) {
	bm.StorageUsed -= bytes
	if bm.StorageUsed < 0 {
		bm.StorageUsed = 0
	}
	bm.UpdateActivity()
}