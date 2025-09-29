package space

import (
// "time" // 暂时注释掉，如果需要可以取消注释
)

// SpaceAggregate 空间聚合根
type SpaceAggregate struct {
	*Space
	collaborators map[string]*SpaceCollaborator
	settings      *SpaceSettings
	metrics       *SpaceMetrics
	events        []DomainEvent
}

// NewSpaceAggregate 创建空间聚合根
func NewSpaceAggregate(space *Space) *SpaceAggregate {
	return &SpaceAggregate{
		Space:         space,
		collaborators: make(map[string]*SpaceCollaborator),
		settings:      NewDefaultSpaceSettings(),
		metrics:       NewSpaceMetrics(),
		events:        make([]DomainEvent, 0),
	}
}

// GetUncommittedEvents 获取未提交的事件
func (a *SpaceAggregate) GetUncommittedEvents() []DomainEvent {
	return a.events
}

// MarkEventsAsCommitted 标记事件为已提交
func (a *SpaceAggregate) MarkEventsAsCommitted() {
	a.events = make([]DomainEvent, 0)
}

// addEvent 添加领域事件
func (a *SpaceAggregate) addEvent(event DomainEvent) {
	a.events = append(a.events, event)
}

// CreateSpace 创建空间（聚合根方法）
func (a *SpaceAggregate) CreateSpace(name, createdBy string, description, icon *string) error {
	// 创建空间实体
	space, err := NewSpace(name, createdBy)
	if err != nil {
		return err
	}

	// 设置可选属性
	if description != nil {
		space.Description = description
	}
	if icon != nil {
		space.Icon = icon
	}

	a.Space = space
	a.settings = NewDefaultSpaceSettings()
	a.metrics = NewSpaceMetrics()

	// 发布空间创建事件
	event := NewSpaceCreatedEvent(a.Space)
	a.addEvent(event)

	return nil
}

// UpdateSpace 更新空间信息（聚合根方法）
func (a *SpaceAggregate) UpdateSpace(name, description, icon *string, updatedBy string) error {
	// 验证空间是否可以被更新
	if err := a.Space.ValidateForUpdate(); err != nil {
		return err
	}

	// 记录变更前的数据
	previousData := map[string]interface{}{
		"name":        a.Space.Name,
		"description": a.Space.Description,
		"icon":        a.Space.Icon,
	}

	// 更新空间信息
	if err := a.Space.Update(name, description, icon); err != nil {
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

	// 发布空间更新事件
	event := NewSpaceUpdatedEvent(a.Space.ID, changes, updatedBy)
	event.PreviousData = previousData
	a.addEvent(event)

	// 更新活动时间
	a.metrics.UpdateActivity()

	return nil
}

// DeleteSpace 删除空间（聚合根方法）
func (a *SpaceAggregate) DeleteSpace(deletedBy, reason string) error {
	// 验证空间是否可以被删除
	if err := a.Space.ValidateForDeletion(); err != nil {
		return err
	}

	a.Space.SoftDelete()

	// 发布空间删除事件
	event := NewSpaceDeletedEvent(a.Space, deletedBy, reason)
	a.addEvent(event)

	return nil
}

// UpdateSettings 更新空间设置（聚合根方法）
func (a *SpaceAggregate) UpdateSettings(newSettings *SpaceSettings, updatedBy string) error {
	// 验证新设置
	if err := newSettings.Validate(); err != nil {
		return err
	}

	// 记录旧设置
	oldSettingsJSON, _ := a.settings.ToJSON()
	newSettingsJSON, _ := newSettings.ToJSON()

	// 更新设置
	a.settings = newSettings
	a.Space.updateModifiedTime()

	// 准备事件数据
	var oldData, newData map[string]interface{}
	if oldSettingsJSON != "" {
		// 这里简化处理，实际应该解析JSON
		oldData = map[string]interface{}{"settings": oldSettingsJSON}
	}
	if newSettingsJSON != "" {
		newData = map[string]interface{}{"settings": newSettingsJSON}
	}

	// 发布设置更新事件
	event := NewSpaceSettingsUpdatedEvent(a.Space.ID, newData, oldData, updatedBy)
	a.addEvent(event)

	// 更新活动时间
	a.metrics.UpdateActivity()

	return nil
}

// AddCollaborator 添加协作者（聚合根方法）
func (a *SpaceAggregate) AddCollaborator(userID string, role CollaboratorRole, invitedBy string) (*SpaceCollaborator, error) {
	// 检查用户是否已经是协作者
	for _, collab := range a.collaborators {
		if collab.UserID == userID {
			return nil, ErrCollaboratorExists
		}
	}

	// 创建协作者
	collaborator, err := NewSpaceCollaborator(a.Space.ID, userID, role, invitedBy)
	if err != nil {
		return nil, err
	}

	// 添加到协作者列表
	a.collaborators[collaborator.ID] = collaborator

	// 发布协作者添加事件
	event := NewSpaceCollaboratorAddedEvent(collaborator)
	a.addEvent(event)

	// 发布邀请事件
	inviteEvent := NewSpaceCollaboratorInvitedEvent(collaborator, "")
	a.addEvent(inviteEvent)

	// 更新指标
	a.metrics.AddCollaborator()

	return collaborator, nil
}

// RemoveCollaborator 移除协作者（聚合根方法）
func (a *SpaceAggregate) RemoveCollaborator(collaboratorID, removedBy, reason string) error {
	collaborator, exists := a.collaborators[collaboratorID]
	if !exists {
		return DomainError{Code: "COLLABORATOR_NOT_FOUND", Message: "collaborator not found"}
	}

	// 删除协作者
	delete(a.collaborators, collaboratorID)

	// 发布协作者移除事件
	event := NewSpaceCollaboratorRemovedEvent(collaborator, removedBy, reason)
	a.addEvent(event)

	// 更新指标
	a.metrics.RemoveCollaborator()

	return nil
}

// UpdateCollaboratorRole 更新协作者角色（聚合根方法）
func (a *SpaceAggregate) UpdateCollaboratorRole(collaboratorID string, newRole CollaboratorRole, updatedBy string) error {
	collaborator, exists := a.collaborators[collaboratorID]
	if !exists {
		return DomainError{Code: "COLLABORATOR_NOT_FOUND", Message: "collaborator not found"}
	}

	oldRole := collaborator.Role

	// 更新角色
	if err := collaborator.UpdateRole(newRole); err != nil {
		return err
	}

	// 发布角色更新事件
	event := NewSpaceCollaboratorRoleUpdatedEvent(collaborator, oldRole, updatedBy)
	a.addEvent(event)

	// 更新活动时间
	a.metrics.UpdateActivity()

	return nil
}

// AcceptCollaboratorInvite 接受协作者邀请（聚合根方法）
func (a *SpaceAggregate) AcceptCollaboratorInvite(collaboratorID string) error {
	collaborator, exists := a.collaborators[collaboratorID]
	if !exists {
		return DomainError{Code: "COLLABORATOR_NOT_FOUND", Message: "collaborator not found"}
	}

	if collaborator.Status != "pending" {
		return DomainError{Code: "INVALID_INVITE_STATUS", Message: "invite is not pending"}
	}

	// 接受邀请
	collaborator.Accept()

	// 发布接受邀请事件
	event := NewSpaceCollaboratorAcceptedEvent(collaborator)
	a.addEvent(event)

	// 更新活动时间
	a.metrics.UpdateActivity()

	return nil
}

// RejectCollaboratorInvite 拒绝协作者邀请（聚合根方法）
func (a *SpaceAggregate) RejectCollaboratorInvite(collaboratorID string) error {
	collaborator, exists := a.collaborators[collaboratorID]
	if !exists {
		return DomainError{Code: "COLLABORATOR_NOT_FOUND", Message: "collaborator not found"}
	}

	if collaborator.Status != "pending" {
		return DomainError{Code: "INVALID_INVITE_STATUS", Message: "invite is not pending"}
	}

	// 拒绝邀请
	collaborator.Reject()

	// 发布拒绝邀请事件
	event := NewSpaceCollaboratorRejectedEvent(collaborator)
	a.addEvent(event)

	// 从协作者列表中移除
	delete(a.collaborators, collaboratorID)

	return nil
}

// UpdateMetrics 更新空间指标（聚合根方法）
func (a *SpaceAggregate) UpdateMetrics(metricType string, oldValue, newValue interface{}) {
	// 发布指标更新事件
	event := NewSpaceMetricsUpdatedEvent(a.Space.ID, metricType, oldValue, newValue)
	a.addEvent(event)

	// 更新活动时间
	a.metrics.UpdateActivity()
}

// CheckUserPermission 检查用户权限
func (a *SpaceAggregate) CheckUserPermission(userID, permission string) bool {
	// 空间创建者拥有所有权限
	if a.Space.CreatedBy == userID {
		return true
	}

	// 检查协作者权限
	for _, collaborator := range a.collaborators {
		if collaborator.UserID == userID && collaborator.IsActive() {
			return collaborator.HasPermission(permission)
		}
	}

	return false
}

// GetCollaborators 获取协作者列表
func (a *SpaceAggregate) GetCollaborators() []*SpaceCollaborator {
	collaborators := make([]*SpaceCollaborator, 0, len(a.collaborators))
	for _, collaborator := range a.collaborators {
		collaborators = append(collaborators, collaborator)
	}
	return collaborators
}

// GetActiveCollaborators 获取活跃协作者列表
func (a *SpaceAggregate) GetActiveCollaborators() []*SpaceCollaborator {
	collaborators := make([]*SpaceCollaborator, 0)
	for _, collaborator := range a.collaborators {
		if collaborator.IsActive() {
			collaborators = append(collaborators, collaborator)
		}
	}
	return collaborators
}

// GetSettings 获取空间设置
func (a *SpaceAggregate) GetSettings() *SpaceSettings {
	return a.settings
}

// GetMetrics 获取空间指标
func (a *SpaceAggregate) GetMetrics() *SpaceMetrics {
	return a.metrics
}

// LoadCollaborators 加载协作者（从仓储重构时使用）
func (a *SpaceAggregate) LoadCollaborators(collaborators []*SpaceCollaborator) {
	a.collaborators = make(map[string]*SpaceCollaborator)
	for _, collaborator := range collaborators {
		a.collaborators[collaborator.ID] = collaborator
	}
}

// LoadSettings 加载设置（从仓储重构时使用）
func (a *SpaceAggregate) LoadSettings(settings *SpaceSettings) {
	if settings != nil {
		a.settings = settings
	} else {
		a.settings = NewDefaultSpaceSettings()
	}
}

// LoadMetrics 加载指标（从仓储重构时使用）
func (a *SpaceAggregate) LoadMetrics(metrics *SpaceMetrics) {
	if metrics != nil {
		a.metrics = metrics
	} else {
		a.metrics = NewSpaceMetrics()
	}
}

// GetSpace 获取空间实体
func (a *SpaceAggregate) GetSpace() *Space {
	return a.Space
}

// GetID 获取聚合根ID
func (a *SpaceAggregate) GetID() string {
	if a.Space == nil {
		return ""
	}
	return a.Space.ID
}

// GetVersion 获取聚合根版本（基于最后修改时间）
func (a *SpaceAggregate) GetVersion() int64 {
	if a.Space == nil || a.Space.LastModifiedTime == nil {
		return 0
	}
	return a.Space.LastModifiedTime.Unix()
}

// IsDeleted 检查聚合根是否已删除
func (a *SpaceAggregate) IsDeleted() bool {
	if a.Space == nil {
		return true
	}
	return a.Space.IsDeleted()
}
