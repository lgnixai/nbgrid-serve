package user

import (
	"time"
)

// UserAggregate 用户聚合根
type UserAggregate struct {
	*User
	events []DomainEvent
}

// NewUserAggregate 创建用户聚合根
func NewUserAggregate(user *User) *UserAggregate {
	return &UserAggregate{
		User:   user,
		events: make([]DomainEvent, 0),
	}
}

// GetUncommittedEvents 获取未提交的事件
func (a *UserAggregate) GetUncommittedEvents() []DomainEvent {
	return a.events
}

// MarkEventsAsCommitted 标记事件为已提交
func (a *UserAggregate) MarkEventsAsCommitted() {
	a.events = make([]DomainEvent, 0)
}

// addEvent 添加领域事件
func (a *UserAggregate) addEvent(event DomainEvent) {
	a.events = append(a.events, event)
}

// CreateUser 创建用户（聚合根方法）
func (a *UserAggregate) CreateUser(name, email, password string, createdBy string) error {
	// 验证输入参数
	if err := validateUserName(name); err != nil {
		return err
	}
	
	if err := validateEmail(email); err != nil {
		return err
	}

	// 创建用户实体
	now := time.Now()
	a.User = &User{
		ID:               generateUserID(),
		Name:             name,
		Email:            email,
		IsSystem:         false,
		IsAdmin:          false,
		IsTrialUsed:      false,
		CreatedTime:      now,
		LastModifiedTime: &now,
	}

	// 设置密码（如果提供）
	if password != "" {
		if err := a.User.SetPassword(password); err != nil {
			return err
		}
	}

	// 发布用户创建事件
	event := NewUserCreatedEvent(a.User, createdBy)
	a.addEvent(event)

	return nil
}

// UpdateProfile 更新用户资料（聚合根方法）
func (a *UserAggregate) UpdateProfile(name, phone *string, avatar *string, updatedBy string) error {
	// 验证用户是否可以被更新
	if err := a.User.ValidateForUpdate(); err != nil {
		return err
	}

	// 记录变更前的数据
	previousData := map[string]interface{}{
		"name":   a.User.Name,
		"phone":  a.User.Phone,
		"avatar": a.User.Avatar,
	}

	// 更新用户信息
	if err := a.User.UpdateProfile(name, phone, avatar); err != nil {
		return err
	}

	// 记录变更
	changes := make(map[string]interface{})
	if name != nil {
		changes["name"] = *name
	}
	if phone != nil {
		changes["phone"] = *phone
	}
	if avatar != nil {
		changes["avatar"] = *avatar
	}

	// 发布用户更新事件
	event := NewUserUpdatedEvent(a.User.ID, changes, updatedBy)
	event.PreviousData = previousData
	a.addEvent(event)

	return nil
}

// ChangePassword 修改密码（聚合根方法）
func (a *UserAggregate) ChangePassword(oldPassword, newPassword, changedBy string) error {
	// 验证旧密码
	if err := a.User.CheckPassword(oldPassword); err != nil {
		return ErrInvalidPassword
	}

	// 设置新密码
	if err := a.User.SetPassword(newPassword); err != nil {
		return err
	}

	// 发布密码修改事件
	event := NewUserPasswordChangedEvent(a.User.ID, changedBy, false)
	a.addEvent(event)

	return nil
}

// ResetPassword 重置密码（聚合根方法）
func (a *UserAggregate) ResetPassword(newPassword, resetBy string) error {
	// 设置新密码
	if err := a.User.SetPassword(newPassword); err != nil {
		return err
	}

	// 发布密码重置事件
	event := NewUserPasswordChangedEvent(a.User.ID, resetBy, true)
	a.addEvent(event)

	return nil
}

// Activate 激活用户（聚合根方法）
func (a *UserAggregate) Activate(activatedBy string) {
	a.User.Activate()

	// 发布用户激活事件
	event := NewUserActivatedEvent(a.User.ID, activatedBy)
	a.addEvent(event)
}

// Deactivate 停用用户（聚合根方法）
func (a *UserAggregate) Deactivate(deactivatedBy, reason string) {
	a.User.Deactivate()

	// 发布用户停用事件
	event := NewUserDeactivatedEvent(a.User.ID, deactivatedBy, reason)
	a.addEvent(event)
}

// Delete 删除用户（聚合根方法）
func (a *UserAggregate) Delete(deletedBy, reason string) error {
	// 验证用户是否可以被删除
	if err := a.User.ValidateForDeletion(); err != nil {
		return err
	}

	a.User.SoftDelete()

	// 发布用户删除事件
	event := NewUserDeletedEvent(a.User, deletedBy, reason)
	a.addEvent(event)

	return nil
}

// PromoteToAdmin 提升为管理员（聚合根方法）
func (a *UserAggregate) PromoteToAdmin(promotedBy string) {
	fromRole := "user"
	if a.User.IsAdmin {
		return // 已经是管理员
	}

	a.User.PromoteToAdmin()

	// 发布用户提升事件
	event := NewUserPromotedEvent(a.User.ID, promotedBy, fromRole, "admin")
	a.addEvent(event)
}

// DemoteFromAdmin 取消管理员（聚合根方法）
func (a *UserAggregate) DemoteFromAdmin(demotedBy string) {
	fromRole := "admin"
	if !a.User.IsAdmin {
		return // 不是管理员
	}

	a.User.DemoteFromAdmin()

	// 发布用户降级事件
	event := NewUserDemotedEvent(a.User.ID, demotedBy, fromRole, "user")
	a.addEvent(event)
}

// RecordSignIn 记录登录（聚合根方法）
func (a *UserAggregate) RecordSignIn(ipAddress, userAgent, location string) {
	a.User.RecordSignIn()

	// 发布用户登录事件
	event := NewUserSignedInEvent(a.User.ID, ipAddress, userAgent, location)
	a.addEvent(event)
}

// LinkAccount 关联第三方账户（聚合根方法）
func (a *UserAggregate) LinkAccount(accountType AccountType, provider Provider, providerID string) *Account {
	account := a.User.AddAccount(accountType, provider, providerID)

	// 发布账户关联事件
	event := NewUserAccountLinkedEvent(a.User.ID, account.ID, string(provider), providerID)
	a.addEvent(event)

	return account
}

// UnlinkAccount 取消关联第三方账户（聚合根方法）
func (a *UserAggregate) UnlinkAccount(accountID, provider, providerID string) {
	// 发布账户取消关联事件
	event := NewUserAccountUnlinkedEvent(a.User.ID, accountID, provider, providerID)
	a.addEvent(event)
}

// UpdatePreferences 更新用户偏好设置（聚合根方法）
func (a *UserAggregate) UpdatePreferences(prefs *UserPreferences) error {
	// 验证偏好设置
	if err := prefs.Validate(); err != nil {
		return err
	}

	// 获取当前偏好设置
	var currentPrefsJSON string
	if a.User.NotifyMeta != nil {
		currentPrefsJSON = *a.User.NotifyMeta
	}
	currentPrefs, _ := UserPreferencesFromJSON(currentPrefsJSON)

	// 将偏好设置序列化为JSON
	prefsJSON, err := prefs.ToJSON()
	if err != nil {
		return err
	}

	a.User.NotifyMeta = &prefsJSON
	a.User.updateModifiedTime()

	// 准备事件数据
	updatedData := map[string]interface{}{
		"language":      prefs.Language,
		"timezone":      prefs.Timezone,
		"date_format":   prefs.DateFormat,
		"time_format":   prefs.TimeFormat,
		"theme":         prefs.Theme,
		"notifications": prefs.Notifications,
		"display":       prefs.Display,
		"privacy":       prefs.Privacy,
	}

	var previousData map[string]interface{}
	if currentPrefs != nil {
		previousData = map[string]interface{}{
			"language":      currentPrefs.Language,
			"timezone":      currentPrefs.Timezone,
			"date_format":   currentPrefs.DateFormat,
			"time_format":   currentPrefs.TimeFormat,
			"theme":         currentPrefs.Theme,
			"notifications": currentPrefs.Notifications,
			"display":       currentPrefs.Display,
			"privacy":       currentPrefs.Privacy,
		}
	}

	// 发布偏好设置更新事件
	event := NewUserPreferencesUpdatedEvent(a.User.ID, updatedData, previousData)
	a.addEvent(event)

	return nil
}

// GetUser 获取用户实体
func (a *UserAggregate) GetUser() *User {
	return a.User
}

// GetID 获取聚合根ID
func (a *UserAggregate) GetID() string {
	if a.User == nil {
		return ""
	}
	return a.User.ID
}

// GetVersion 获取聚合根版本（基于最后修改时间）
func (a *UserAggregate) GetVersion() int64 {
	if a.User == nil || a.User.LastModifiedTime == nil {
		return 0
	}
	return a.User.LastModifiedTime.Unix()
}

// IsDeleted 检查聚合根是否已删除
func (a *UserAggregate) IsDeleted() bool {
	if a.User == nil {
		return true
	}
	return a.User.DeletedTime != nil
}