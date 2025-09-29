package user

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

// 用户相关事件类型常量
const (
	UserCreatedEventType            = "user.created"
	UserUpdatedEventType            = "user.updated"
	UserDeletedEventType            = "user.deleted"
	UserActivatedEventType          = "user.activated"
	UserDeactivatedEventType        = "user.deactivated"
	UserPasswordChangedEventType    = "user.password_changed"
	UserPromotedEventType           = "user.promoted"
	UserDemotedEventType            = "user.demoted"
	UserSignedInEventType           = "user.signed_in"
	UserAccountLinkedEventType      = "user.account_linked"
	UserAccountUnlinkedEventType    = "user.account_unlinked"
	UserPreferencesUpdatedEventType = "user.preferences_updated"
)

// UserCreatedEvent 用户创建事件
type UserCreatedEvent struct {
	BaseDomainEvent
	UserID    string `json:"user_id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	IsAdmin   bool   `json:"is_admin"`
	IsSystem  bool   `json:"is_system"`
	CreatedBy string `json:"created_by,omitempty"`
}

// NewUserCreatedEvent 创建用户创建事件
func NewUserCreatedEvent(user *User, createdBy string) *UserCreatedEvent {
	return &UserCreatedEvent{
		BaseDomainEvent: BaseDomainEvent{
			ID:          generateEventID(),
			Type:        UserCreatedEventType,
			AggregateId: user.ID,
			OccurredOn:  time.Now(),
		},
		UserID:    user.ID,
		Name:      user.Name,
		Email:     user.Email,
		IsAdmin:   user.IsAdmin,
		IsSystem:  user.IsSystem,
		CreatedBy: createdBy,
	}
}

// UserUpdatedEvent 用户更新事件
type UserUpdatedEvent struct {
	BaseDomainEvent
	UserID       string                 `json:"user_id"`
	Changes      map[string]interface{} `json:"changes"`
	UpdatedBy    string                 `json:"updated_by,omitempty"`
	PreviousData map[string]interface{} `json:"previous_data,omitempty"`
}

// NewUserUpdatedEvent 创建用户更新事件
func NewUserUpdatedEvent(userID string, changes map[string]interface{}, updatedBy string) *UserUpdatedEvent {
	return &UserUpdatedEvent{
		BaseDomainEvent: BaseDomainEvent{
			ID:          generateEventID(),
			Type:        UserUpdatedEventType,
			AggregateId: userID,
			OccurredOn:  time.Now(),
		},
		UserID:    userID,
		Changes:   changes,
		UpdatedBy: updatedBy,
	}
}

// UserDeletedEvent 用户删除事件
type UserDeletedEvent struct {
	BaseDomainEvent
	UserID    string `json:"user_id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	DeletedBy string `json:"deleted_by,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

// NewUserDeletedEvent 创建用户删除事件
func NewUserDeletedEvent(user *User, deletedBy, reason string) *UserDeletedEvent {
	return &UserDeletedEvent{
		BaseDomainEvent: BaseDomainEvent{
			ID:          generateEventID(),
			Type:        UserDeletedEventType,
			AggregateId: user.ID,
			OccurredOn:  time.Now(),
		},
		UserID:    user.ID,
		Name:      user.Name,
		Email:     user.Email,
		DeletedBy: deletedBy,
		Reason:    reason,
	}
}

// UserActivatedEvent 用户激活事件
type UserActivatedEvent struct {
	BaseDomainEvent
	UserID      string `json:"user_id"`
	ActivatedBy string `json:"activated_by,omitempty"`
}

// NewUserActivatedEvent 创建用户激活事件
func NewUserActivatedEvent(userID, activatedBy string) *UserActivatedEvent {
	return &UserActivatedEvent{
		BaseDomainEvent: BaseDomainEvent{
			ID:          generateEventID(),
			Type:        UserActivatedEventType,
			AggregateId: userID,
			OccurredOn:  time.Now(),
		},
		UserID:      userID,
		ActivatedBy: activatedBy,
	}
}

// UserDeactivatedEvent 用户停用事件
type UserDeactivatedEvent struct {
	BaseDomainEvent
	UserID        string `json:"user_id"`
	DeactivatedBy string `json:"deactivated_by,omitempty"`
	Reason        string `json:"reason,omitempty"`
}

// NewUserDeactivatedEvent 创建用户停用事件
func NewUserDeactivatedEvent(userID, deactivatedBy, reason string) *UserDeactivatedEvent {
	return &UserDeactivatedEvent{
		BaseDomainEvent: BaseDomainEvent{
			ID:          generateEventID(),
			Type:        UserDeactivatedEventType,
			AggregateId: userID,
			OccurredOn:  time.Now(),
		},
		UserID:        userID,
		DeactivatedBy: deactivatedBy,
		Reason:        reason,
	}
}

// UserPasswordChangedEvent 用户密码修改事件
type UserPasswordChangedEvent struct {
	BaseDomainEvent
	UserID    string `json:"user_id"`
	ChangedBy string `json:"changed_by,omitempty"`
	IsReset   bool   `json:"is_reset"`
}

// NewUserPasswordChangedEvent 创建用户密码修改事件
func NewUserPasswordChangedEvent(userID, changedBy string, isReset bool) *UserPasswordChangedEvent {
	return &UserPasswordChangedEvent{
		BaseDomainEvent: BaseDomainEvent{
			ID:          generateEventID(),
			Type:        UserPasswordChangedEventType,
			AggregateId: userID,
			OccurredOn:  time.Now(),
		},
		UserID:    userID,
		ChangedBy: changedBy,
		IsReset:   isReset,
	}
}

// UserPromotedEvent 用户提升事件
type UserPromotedEvent struct {
	BaseDomainEvent
	UserID     string `json:"user_id"`
	PromotedBy string `json:"promoted_by,omitempty"`
	ToRole     string `json:"to_role"`
	FromRole   string `json:"from_role"`
}

// NewUserPromotedEvent 创建用户提升事件
func NewUserPromotedEvent(userID, promotedBy, fromRole, toRole string) *UserPromotedEvent {
	return &UserPromotedEvent{
		BaseDomainEvent: BaseDomainEvent{
			ID:          generateEventID(),
			Type:        UserPromotedEventType,
			AggregateId: userID,
			OccurredOn:  time.Now(),
		},
		UserID:     userID,
		PromotedBy: promotedBy,
		FromRole:   fromRole,
		ToRole:     toRole,
	}
}

// UserDemotedEvent 用户降级事件
type UserDemotedEvent struct {
	BaseDomainEvent
	UserID    string `json:"user_id"`
	DemotedBy string `json:"demoted_by,omitempty"`
	ToRole    string `json:"to_role"`
	FromRole  string `json:"from_role"`
}

// NewUserDemotedEvent 创建用户降级事件
func NewUserDemotedEvent(userID, demotedBy, fromRole, toRole string) *UserDemotedEvent {
	return &UserDemotedEvent{
		BaseDomainEvent: BaseDomainEvent{
			ID:          generateEventID(),
			Type:        UserDemotedEventType,
			AggregateId: userID,
			OccurredOn:  time.Now(),
		},
		UserID:    userID,
		DemotedBy: demotedBy,
		FromRole:  fromRole,
		ToRole:    toRole,
	}
}

// UserSignedInEvent 用户登录事件
type UserSignedInEvent struct {
	BaseDomainEvent
	UserID    string `json:"user_id"`
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
	Location  string `json:"location,omitempty"`
}

// NewUserSignedInEvent 创建用户登录事件
func NewUserSignedInEvent(userID, ipAddress, userAgent, location string) *UserSignedInEvent {
	return &UserSignedInEvent{
		BaseDomainEvent: BaseDomainEvent{
			ID:          generateEventID(),
			Type:        UserSignedInEventType,
			AggregateId: userID,
			OccurredOn:  time.Now(),
		},
		UserID:    userID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Location:  location,
	}
}

// UserAccountLinkedEvent 用户账户关联事件
type UserAccountLinkedEvent struct {
	BaseDomainEvent
	UserID     string `json:"user_id"`
	AccountID  string `json:"account_id"`
	Provider   string `json:"provider"`
	ProviderID string `json:"provider_id"`
}

// NewUserAccountLinkedEvent 创建用户账户关联事件
func NewUserAccountLinkedEvent(userID, accountID, provider, providerID string) *UserAccountLinkedEvent {
	return &UserAccountLinkedEvent{
		BaseDomainEvent: BaseDomainEvent{
			ID:          generateEventID(),
			Type:        UserAccountLinkedEventType,
			AggregateId: userID,
			OccurredOn:  time.Now(),
		},
		UserID:     userID,
		AccountID:  accountID,
		Provider:   provider,
		ProviderID: providerID,
	}
}

// UserAccountUnlinkedEvent 用户账户取消关联事件
type UserAccountUnlinkedEvent struct {
	BaseDomainEvent
	UserID     string `json:"user_id"`
	AccountID  string `json:"account_id"`
	Provider   string `json:"provider"`
	ProviderID string `json:"provider_id"`
}

// NewUserAccountUnlinkedEvent 创建用户账户取消关联事件
func NewUserAccountUnlinkedEvent(userID, accountID, provider, providerID string) *UserAccountUnlinkedEvent {
	return &UserAccountUnlinkedEvent{
		BaseDomainEvent: BaseDomainEvent{
			ID:          generateEventID(),
			Type:        UserAccountUnlinkedEventType,
			AggregateId: userID,
			OccurredOn:  time.Now(),
		},
		UserID:     userID,
		AccountID:  accountID,
		Provider:   provider,
		ProviderID: providerID,
	}
}

// UserPreferencesUpdatedEvent 用户偏好设置更新事件
type UserPreferencesUpdatedEvent struct {
	BaseDomainEvent
	UserID              string                 `json:"user_id"`
	UpdatedPreferences  map[string]interface{} `json:"updated_preferences"`
	PreviousPreferences map[string]interface{} `json:"previous_preferences,omitempty"`
}

// NewUserPreferencesUpdatedEvent 创建用户偏好设置更新事件
func NewUserPreferencesUpdatedEvent(userID string, updated, previous map[string]interface{}) *UserPreferencesUpdatedEvent {
	return &UserPreferencesUpdatedEvent{
		BaseDomainEvent: BaseDomainEvent{
			ID:          generateEventID(),
			Type:        UserPreferencesUpdatedEventType,
			AggregateId: userID,
			OccurredOn:  time.Now(),
		},
		UserID:              userID,
		UpdatedPreferences:  updated,
		PreviousPreferences: previous,
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
	return generateUserID() // 复用用户ID生成逻辑
}
