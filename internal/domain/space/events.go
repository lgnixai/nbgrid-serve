package space

import (
	"time"
)

// DomainEvent 领域事件接口
type DomainEvent interface {
	EventID() string
	EventType() string
	AggregateID() string
	OccurredAt() time.Time
	EventData() interface{}
}

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

// 空间相关事件类型常量
const (
	SpaceCreatedEventType              = "space.created"
	SpaceUpdatedEventType              = "space.updated"
	SpaceDeletedEventType              = "space.deleted"
	SpaceSettingsUpdatedEventType      = "space.settings_updated"
	SpaceCollaboratorAddedEventType    = "space.collaborator_added"
	SpaceCollaboratorRemovedEventType  = "space.collaborator_removed"
	SpaceCollaboratorRoleUpdatedEventType = "space.collaborator_role_updated"
	SpaceCollaboratorInvitedEventType  = "space.collaborator_invited"
	SpaceCollaboratorAcceptedEventType = "space.collaborator_accepted"
	SpaceCollaboratorRejectedEventType = "space.collaborator_rejected"
	SpaceMetricsUpdatedEventType       = "space.metrics_updated"
)

// SpaceCreatedEvent 空间创建事件
type SpaceCreatedEvent struct {
	BaseDomainEvent
	SpaceID     string `json:"space_id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CreatedBy   string `json:"created_by"`
}

// NewSpaceCreatedEvent 创建空间创建事件
func NewSpaceCreatedEvent(space *Space) *SpaceCreatedEvent {
	var description string
	if space.Description != nil {
		description = *space.Description
	}

	return &SpaceCreatedEvent{
		BaseDomainEvent: BaseDomainEvent{
			ID:          generateEventID(),
			Type:        SpaceCreatedEventType,
			AggregateId: space.ID,
			OccurredOn:  time.Now(),
		},
		SpaceID:     space.ID,
		Name:        space.Name,
		Description: description,
		CreatedBy:   space.CreatedBy,
	}
}

// SpaceUpdatedEvent 空间更新事件
type SpaceUpdatedEvent struct {
	BaseDomainEvent
	SpaceID      string                 `json:"space_id"`
	Changes      map[string]interface{} `json:"changes"`
	UpdatedBy    string                 `json:"updated_by,omitempty"`
	PreviousData map[string]interface{} `json:"previous_data,omitempty"`
}

// NewSpaceUpdatedEvent 创建空间更新事件
func NewSpaceUpdatedEvent(spaceID string, changes map[string]interface{}, updatedBy string) *SpaceUpdatedEvent {
	return &SpaceUpdatedEvent{
		BaseDomainEvent: BaseDomainEvent{
			ID:          generateEventID(),
			Type:        SpaceUpdatedEventType,
			AggregateId: spaceID,
			OccurredOn:  time.Now(),
		},
		SpaceID:   spaceID,
		Changes:   changes,
		UpdatedBy: updatedBy,
	}
}

// SpaceDeletedEvent 空间删除事件
type SpaceDeletedEvent struct {
	BaseDomainEvent
	SpaceID   string `json:"space_id"`
	Name      string `json:"name"`
	DeletedBy string `json:"deleted_by,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

// NewSpaceDeletedEvent 创建空间删除事件
func NewSpaceDeletedEvent(space *Space, deletedBy, reason string) *SpaceDeletedEvent {
	return &SpaceDeletedEvent{
		BaseDomainEvent: BaseDomainEvent{
			ID:          generateEventID(),
			Type:        SpaceDeletedEventType,
			AggregateId: space.ID,
			OccurredOn:  time.Now(),
		},
		SpaceID:   space.ID,
		Name:      space.Name,
		DeletedBy: deletedBy,
		Reason:    reason,
	}
}

// SpaceSettingsUpdatedEvent 空间设置更新事件
type SpaceSettingsUpdatedEvent struct {
	BaseDomainEvent
	SpaceID            string                 `json:"space_id"`
	UpdatedSettings    map[string]interface{} `json:"updated_settings"`
	PreviousSettings   map[string]interface{} `json:"previous_settings,omitempty"`
	UpdatedBy          string                 `json:"updated_by,omitempty"`
}

// NewSpaceSettingsUpdatedEvent 创建空间设置更新事件
func NewSpaceSettingsUpdatedEvent(spaceID string, updated, previous map[string]interface{}, updatedBy string) *SpaceSettingsUpdatedEvent {
	return &SpaceSettingsUpdatedEvent{
		BaseDomainEvent: BaseDomainEvent{
			ID:          generateEventID(),
			Type:        SpaceSettingsUpdatedEventType,
			AggregateId: spaceID,
			OccurredOn:  time.Now(),
		},
		SpaceID:          spaceID,
		UpdatedSettings:  updated,
		PreviousSettings: previous,
		UpdatedBy:        updatedBy,
	}
}

// SpaceCollaboratorAddedEvent 空间协作者添加事件
type SpaceCollaboratorAddedEvent struct {
	BaseDomainEvent
	SpaceID         string           `json:"space_id"`
	CollaboratorID  string           `json:"collaborator_id"`
	UserID          string           `json:"user_id"`
	Role            CollaboratorRole `json:"role"`
	InvitedBy       string           `json:"invited_by"`
}

// NewSpaceCollaboratorAddedEvent 创建空间协作者添加事件
func NewSpaceCollaboratorAddedEvent(collaborator *SpaceCollaborator) *SpaceCollaboratorAddedEvent {
	return &SpaceCollaboratorAddedEvent{
		BaseDomainEvent: BaseDomainEvent{
			ID:          generateEventID(),
			Type:        SpaceCollaboratorAddedEventType,
			AggregateId: collaborator.SpaceID,
			OccurredOn:  time.Now(),
		},
		SpaceID:        collaborator.SpaceID,
		CollaboratorID: collaborator.ID,
		UserID:         collaborator.UserID,
		Role:           collaborator.Role,
		InvitedBy:      collaborator.InvitedBy,
	}
}

// SpaceCollaboratorRemovedEvent 空间协作者移除事件
type SpaceCollaboratorRemovedEvent struct {
	BaseDomainEvent
	SpaceID        string           `json:"space_id"`
	CollaboratorID string           `json:"collaborator_id"`
	UserID         string           `json:"user_id"`
	Role           CollaboratorRole `json:"role"`
	RemovedBy      string           `json:"removed_by,omitempty"`
	Reason         string           `json:"reason,omitempty"`
}

// NewSpaceCollaboratorRemovedEvent 创建空间协作者移除事件
func NewSpaceCollaboratorRemovedEvent(collaborator *SpaceCollaborator, removedBy, reason string) *SpaceCollaboratorRemovedEvent {
	return &SpaceCollaboratorRemovedEvent{
		BaseDomainEvent: BaseDomainEvent{
			ID:          generateEventID(),
			Type:        SpaceCollaboratorRemovedEventType,
			AggregateId: collaborator.SpaceID,
			OccurredOn:  time.Now(),
		},
		SpaceID:        collaborator.SpaceID,
		CollaboratorID: collaborator.ID,
		UserID:         collaborator.UserID,
		Role:           collaborator.Role,
		RemovedBy:      removedBy,
		Reason:         reason,
	}
}

// SpaceCollaboratorRoleUpdatedEvent 空间协作者角色更新事件
type SpaceCollaboratorRoleUpdatedEvent struct {
	BaseDomainEvent
	SpaceID        string           `json:"space_id"`
	CollaboratorID string           `json:"collaborator_id"`
	UserID         string           `json:"user_id"`
	NewRole        CollaboratorRole `json:"new_role"`
	OldRole        CollaboratorRole `json:"old_role"`
	UpdatedBy      string           `json:"updated_by,omitempty"`
}

// NewSpaceCollaboratorRoleUpdatedEvent 创建空间协作者角色更新事件
func NewSpaceCollaboratorRoleUpdatedEvent(collaborator *SpaceCollaborator, oldRole CollaboratorRole, updatedBy string) *SpaceCollaboratorRoleUpdatedEvent {
	return &SpaceCollaboratorRoleUpdatedEvent{
		BaseDomainEvent: BaseDomainEvent{
			ID:          generateEventID(),
			Type:        SpaceCollaboratorRoleUpdatedEventType,
			AggregateId: collaborator.SpaceID,
			OccurredOn:  time.Now(),
		},
		SpaceID:        collaborator.SpaceID,
		CollaboratorID: collaborator.ID,
		UserID:         collaborator.UserID,
		NewRole:        collaborator.Role,
		OldRole:        oldRole,
		UpdatedBy:      updatedBy,
	}
}

// SpaceCollaboratorInvitedEvent 空间协作者邀请事件
type SpaceCollaboratorInvitedEvent struct {
	BaseDomainEvent
	SpaceID        string           `json:"space_id"`
	CollaboratorID string           `json:"collaborator_id"`
	UserID         string           `json:"user_id"`
	Role           CollaboratorRole `json:"role"`
	InvitedBy      string           `json:"invited_by"`
	InviteToken    string           `json:"invite_token,omitempty"`
}

// NewSpaceCollaboratorInvitedEvent 创建空间协作者邀请事件
func NewSpaceCollaboratorInvitedEvent(collaborator *SpaceCollaborator, inviteToken string) *SpaceCollaboratorInvitedEvent {
	return &SpaceCollaboratorInvitedEvent{
		BaseDomainEvent: BaseDomainEvent{
			ID:          generateEventID(),
			Type:        SpaceCollaboratorInvitedEventType,
			AggregateId: collaborator.SpaceID,
			OccurredOn:  time.Now(),
		},
		SpaceID:        collaborator.SpaceID,
		CollaboratorID: collaborator.ID,
		UserID:         collaborator.UserID,
		Role:           collaborator.Role,
		InvitedBy:      collaborator.InvitedBy,
		InviteToken:    inviteToken,
	}
}

// SpaceCollaboratorAcceptedEvent 空间协作者接受邀请事件
type SpaceCollaboratorAcceptedEvent struct {
	BaseDomainEvent
	SpaceID        string           `json:"space_id"`
	CollaboratorID string           `json:"collaborator_id"`
	UserID         string           `json:"user_id"`
	Role           CollaboratorRole `json:"role"`
	AcceptedAt     time.Time        `json:"accepted_at"`
}

// NewSpaceCollaboratorAcceptedEvent 创建空间协作者接受邀请事件
func NewSpaceCollaboratorAcceptedEvent(collaborator *SpaceCollaborator) *SpaceCollaboratorAcceptedEvent {
	return &SpaceCollaboratorAcceptedEvent{
		BaseDomainEvent: BaseDomainEvent{
			ID:          generateEventID(),
			Type:        SpaceCollaboratorAcceptedEventType,
			AggregateId: collaborator.SpaceID,
			OccurredOn:  time.Now(),
		},
		SpaceID:        collaborator.SpaceID,
		CollaboratorID: collaborator.ID,
		UserID:         collaborator.UserID,
		Role:           collaborator.Role,
		AcceptedAt:     *collaborator.AcceptedAt,
	}
}

// SpaceCollaboratorRejectedEvent 空间协作者拒绝邀请事件
type SpaceCollaboratorRejectedEvent struct {
	BaseDomainEvent
	SpaceID        string           `json:"space_id"`
	CollaboratorID string           `json:"collaborator_id"`
	UserID         string           `json:"user_id"`
	Role           CollaboratorRole `json:"role"`
	RejectedAt     time.Time        `json:"rejected_at"`
}

// NewSpaceCollaboratorRejectedEvent 创建空间协作者拒绝邀请事件
func NewSpaceCollaboratorRejectedEvent(collaborator *SpaceCollaborator) *SpaceCollaboratorRejectedEvent {
	return &SpaceCollaboratorRejectedEvent{
		BaseDomainEvent: BaseDomainEvent{
			ID:          generateEventID(),
			Type:        SpaceCollaboratorRejectedEventType,
			AggregateId: collaborator.SpaceID,
			OccurredOn:  time.Now(),
		},
		SpaceID:        collaborator.SpaceID,
		CollaboratorID: collaborator.ID,
		UserID:         collaborator.UserID,
		Role:           collaborator.Role,
		RejectedAt:     time.Now(),
	}
}

// SpaceMetricsUpdatedEvent 空间指标更新事件
type SpaceMetricsUpdatedEvent struct {
	BaseDomainEvent
	SpaceID        string                 `json:"space_id"`
	MetricType     string                 `json:"metric_type"`
	OldValue       interface{}            `json:"old_value"`
	NewValue       interface{}            `json:"new_value"`
	AdditionalData map[string]interface{} `json:"additional_data,omitempty"`
}

// NewSpaceMetricsUpdatedEvent 创建空间指标更新事件
func NewSpaceMetricsUpdatedEvent(spaceID, metricType string, oldValue, newValue interface{}) *SpaceMetricsUpdatedEvent {
	return &SpaceMetricsUpdatedEvent{
		BaseDomainEvent: BaseDomainEvent{
			ID:          generateEventID(),
			Type:        SpaceMetricsUpdatedEventType,
			AggregateId: spaceID,
			OccurredOn:  time.Now(),
		},
		SpaceID:    spaceID,
		MetricType: metricType,
		OldValue:   oldValue,
		NewValue:   newValue,
	}
}

// EventPublisher 事件发布器接口
type EventPublisher interface {
	Publish(event DomainEvent) error
	PublishBatch(events []DomainEvent) error
}

// EventHandler 事件处理器接口
type EventHandler interface {
	Handle(event DomainEvent) error
	CanHandle(eventType string) bool
}

// EventStore 事件存储接口
type EventStore interface {
	Save(event DomainEvent) error
	SaveBatch(events []DomainEvent) error
	GetEvents(aggregateID string) ([]DomainEvent, error)
	GetEventsByType(eventType string, limit int) ([]DomainEvent, error)
}

// generateEventID 生成事件ID
func generateEventID() string {
	return generateSpaceID() // 复用空间ID生成逻辑
}

// generateSpaceID 生成空间ID
func generateSpaceID() string {
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