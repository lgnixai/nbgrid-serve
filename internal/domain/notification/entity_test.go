package notification

import (
	"testing"
	"time"
)

func TestNewNotification(t *testing.T) {
	tests := []struct {
		name             string
		userID           string
		notificationType NotificationType
		title            string
		content          string
		wantErr          bool
	}{
		{
			name:             "valid notification",
			userID:           "user123",
			notificationType: NotificationTypeUser,
			title:            "Test Notification",
			content:          "This is a test notification",
			wantErr:          false,
		},
		{
			name:             "empty user ID",
			userID:           "",
			notificationType: NotificationTypeUser,
			title:            "Test Notification",
			content:          "This is a test notification",
			wantErr:          false, // 允许空用户ID，可能在系统通知中使用
		},
		{
			name:             "empty title",
			userID:           "user123",
			notificationType: NotificationTypeUser,
			title:            "",
			content:          "This is a test notification",
			wantErr:          false, // 允许空标题
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notification := NewNotification(tt.userID, tt.notificationType, tt.title, tt.content)

			if notification == nil {
				t.Error("NewNotification() returned nil")
				return
			}

			if notification.UserID != tt.userID {
				t.Errorf("NewNotification() UserID = %v, want %v", notification.UserID, tt.userID)
			}

			if notification.Type != tt.notificationType {
				t.Errorf("NewNotification() Type = %v, want %v", notification.Type, tt.notificationType)
			}

			if notification.Title != tt.title {
				t.Errorf("NewNotification() Title = %v, want %v", notification.Title, tt.title)
			}

			if notification.Content != tt.content {
				t.Errorf("NewNotification() Content = %v, want %v", notification.Content, tt.content)
			}

			if notification.Status != NotificationStatusUnread {
				t.Errorf("NewNotification() Status = %v, want %v", notification.Status, NotificationStatusUnread)
			}

			if notification.Priority != NotificationPriorityNormal {
				t.Errorf("NewNotification() Priority = %v, want %v", notification.Priority, NotificationPriorityNormal)
			}

			if notification.ID == "" {
				t.Error("NewNotification() ID should not be empty")
			}

			if notification.Data == nil {
				t.Error("NewNotification() Data should not be nil")
			}

			if notification.CreatedTime.IsZero() {
				t.Error("NewNotification() CreatedTime should not be zero")
			}

			if notification.UpdatedTime.IsZero() {
				t.Error("NewNotification() UpdatedTime should not be zero")
			}
		})
	}
}

func TestNotification_SetData(t *testing.T) {
	notification := NewNotification("user123", NotificationTypeUser, "Test", "Content")

	// 测试设置数据
	notification.SetData("key1", "value1")
	notification.SetData("key2", 123)

	if notification.Data["key1"] != "value1" {
		t.Errorf("SetData() key1 = %v, want %v", notification.Data["key1"], "value1")
	}

	if notification.Data["key2"] != 123 {
		t.Errorf("SetData() key2 = %v, want %v", notification.Data["key2"], 123)
	}

	// 测试更新时间是否改变
	oldUpdatedTime := notification.UpdatedTime
	time.Sleep(time.Millisecond)
	notification.SetData("key3", "value3")

	if notification.UpdatedTime == oldUpdatedTime {
		t.Error("SetData() should update UpdatedTime")
	}
}

func TestNotification_SetPriority(t *testing.T) {
	notification := NewNotification("user123", NotificationTypeUser, "Test", "Content")

	// 测试设置优先级
	notification.SetPriority(NotificationPriorityHigh)

	if notification.Priority != NotificationPriorityHigh {
		t.Errorf("SetPriority() Priority = %v, want %v", notification.Priority, NotificationPriorityHigh)
	}

	// 测试更新时间是否改变
	oldUpdatedTime := notification.UpdatedTime
	time.Sleep(time.Millisecond)
	notification.SetPriority(NotificationPriorityUrgent)

	if notification.UpdatedTime == oldUpdatedTime {
		t.Error("SetPriority() should update UpdatedTime")
	}
}

func TestNotification_SetSource(t *testing.T) {
	notification := NewNotification("user123", NotificationTypeUser, "Test", "Content")

	// 测试设置来源
	sourceID := "source123"
	sourceType := "table"
	notification.SetSource(sourceID, sourceType)

	if notification.SourceID != sourceID {
		t.Errorf("SetSource() SourceID = %v, want %v", notification.SourceID, sourceID)
	}

	if notification.SourceType != sourceType {
		t.Errorf("SetSource() SourceType = %v, want %v", notification.SourceType, sourceType)
	}

	// 测试更新时间是否改变
	oldUpdatedTime := notification.UpdatedTime
	time.Sleep(time.Millisecond)
	notification.SetSource("newSource", "newType")

	if notification.UpdatedTime == oldUpdatedTime {
		t.Error("SetSource() should update UpdatedTime")
	}
}

func TestNotification_SetActionURL(t *testing.T) {
	notification := NewNotification("user123", NotificationTypeUser, "Test", "Content")

	// 测试设置操作链接
	actionURL := "https://example.com/action"
	notification.SetActionURL(actionURL)

	if notification.ActionURL != actionURL {
		t.Errorf("SetActionURL() ActionURL = %v, want %v", notification.ActionURL, actionURL)
	}

	// 测试更新时间是否改变
	oldUpdatedTime := notification.UpdatedTime
	time.Sleep(time.Millisecond)
	notification.SetActionURL("https://example.com/new-action")

	if notification.UpdatedTime == oldUpdatedTime {
		t.Error("SetActionURL() should update UpdatedTime")
	}
}

func TestNotification_SetExpiresAt(t *testing.T) {
	notification := NewNotification("user123", NotificationTypeUser, "Test", "Content")

	// 测试设置过期时间
	expiresAt := time.Now().Add(24 * time.Hour)
	notification.SetExpiresAt(expiresAt)

	if notification.ExpiresAt == nil {
		t.Error("SetExpiresAt() ExpiresAt should not be nil")
	} else if !notification.ExpiresAt.Equal(expiresAt) {
		t.Errorf("SetExpiresAt() ExpiresAt = %v, want %v", notification.ExpiresAt, expiresAt)
	}

	// 测试更新时间是否改变
	oldUpdatedTime := notification.UpdatedTime
	time.Sleep(time.Millisecond)
	newExpiresAt := time.Now().Add(48 * time.Hour)
	notification.SetExpiresAt(newExpiresAt)

	if notification.UpdatedTime == oldUpdatedTime {
		t.Error("SetExpiresAt() should update UpdatedTime")
	}
}

func TestNotification_MarkAsRead(t *testing.T) {
	notification := NewNotification("user123", NotificationTypeUser, "Test", "Content")

	// 初始状态应该是未读
	if notification.Status != NotificationStatusUnread {
		t.Errorf("Initial status = %v, want %v", notification.Status, NotificationStatusUnread)
	}

	if notification.ReadAt != nil {
		t.Error("Initial ReadAt should be nil")
	}

	// 标记为已读
	notification.MarkAsRead()

	if notification.Status != NotificationStatusRead {
		t.Errorf("Status after MarkAsRead = %v, want %v", notification.Status, NotificationStatusRead)
	}

	if notification.ReadAt == nil {
		t.Error("ReadAt should not be nil after MarkAsRead")
	}

	if notification.ReadAt.IsZero() {
		t.Error("ReadAt should not be zero after MarkAsRead")
	}
}

func TestNotification_MarkAsArchived(t *testing.T) {
	notification := NewNotification("user123", NotificationTypeUser, "Test", "Content")

	// 初始状态应该是未读
	if notification.Status != NotificationStatusUnread {
		t.Errorf("Initial status = %v, want %v", notification.Status, NotificationStatusUnread)
	}

	// 标记为已归档
	notification.MarkAsArchived()

	if notification.Status != NotificationStatusArchived {
		t.Errorf("Status after MarkAsArchived = %v, want %v", notification.Status, NotificationStatusArchived)
	}
}

func TestNotification_IsExpired(t *testing.T) {
	notification := NewNotification("user123", NotificationTypeUser, "Test", "Content")

	// 没有设置过期时间的通知不应该过期
	if notification.IsExpired() {
		t.Error("Notification without ExpiresAt should not be expired")
	}

	// 设置未来的过期时间
	futureTime := time.Now().Add(24 * time.Hour)
	notification.SetExpiresAt(futureTime)

	if notification.IsExpired() {
		t.Error("Notification with future ExpiresAt should not be expired")
	}

	// 设置过去的过期时间
	pastTime := time.Now().Add(-24 * time.Hour)
	notification.SetExpiresAt(pastTime)

	if !notification.IsExpired() {
		t.Error("Notification with past ExpiresAt should be expired")
	}
}

func TestNewNotificationTemplate(t *testing.T) {
	tests := []struct {
		name             string
		notificationType NotificationType
		templateName     string
		title            string
		content          string
		wantErr          bool
	}{
		{
			name:             "valid template",
			notificationType: NotificationTypeUser,
			templateName:     "user_welcome",
			title:            "Welcome {{user_name}}",
			content:          "Welcome to our platform, {{user_name}}!",
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := NewNotificationTemplate(tt.notificationType, tt.templateName, tt.title, tt.content)

			if template == nil {
				t.Error("NewNotificationTemplate() returned nil")
				return
			}

			if template.Type != tt.notificationType {
				t.Errorf("NewNotificationTemplate() Type = %v, want %v", template.Type, tt.notificationType)
			}

			if template.Name != tt.templateName {
				t.Errorf("NewNotificationTemplate() Name = %v, want %v", template.Name, tt.templateName)
			}

			if template.Title != tt.title {
				t.Errorf("NewNotificationTemplate() Title = %v, want %v", template.Title, tt.title)
			}

			if template.Content != tt.content {
				t.Errorf("NewNotificationTemplate() Content = %v, want %v", template.Content, tt.content)
			}

			if !template.IsActive {
				t.Error("NewNotificationTemplate() IsActive should be true")
			}

			if template.Variables == nil {
				t.Error("NewNotificationTemplate() Variables should not be nil")
			}

			if template.DefaultData == nil {
				t.Error("NewNotificationTemplate() DefaultData should not be nil")
			}

			if template.ID == "" {
				t.Error("NewNotificationTemplate() ID should not be empty")
			}

			if template.CreatedTime.IsZero() {
				t.Error("NewNotificationTemplate() CreatedTime should not be zero")
			}

			if template.UpdatedTime.IsZero() {
				t.Error("NewNotificationTemplate() UpdatedTime should not be zero")
			}
		})
	}
}

func TestNotificationTemplate_AddVariable(t *testing.T) {
	template := NewNotificationTemplate(NotificationTypeUser, "test", "Title", "Content")

	// 添加变量
	template.AddVariable("user_name")
	template.AddVariable("platform_name")

	if len(template.Variables) != 2 {
		t.Errorf("AddVariable() Variables length = %v, want %v", len(template.Variables), 2)
	}

	if template.Variables[0] != "user_name" {
		t.Errorf("AddVariable() Variables[0] = %v, want %v", template.Variables[0], "user_name")
	}

	if template.Variables[1] != "platform_name" {
		t.Errorf("AddVariable() Variables[1] = %v, want %v", template.Variables[1], "platform_name")
	}

	// 测试更新时间是否改变
	oldUpdatedTime := template.UpdatedTime
	time.Sleep(time.Millisecond)
	template.AddVariable("new_variable")

	if template.UpdatedTime == oldUpdatedTime {
		t.Error("AddVariable() should update UpdatedTime")
	}
}

func TestNotificationTemplate_SetDefaultData(t *testing.T) {
	template := NewNotificationTemplate(NotificationTypeUser, "test", "Title", "Content")

	// 设置默认数据
	template.SetDefaultData("platform_name", "Teable")
	template.SetDefaultData("version", "1.0.0")

	if template.DefaultData["platform_name"] != "Teable" {
		t.Errorf("SetDefaultData() platform_name = %v, want %v", template.DefaultData["platform_name"], "Teable")
	}

	if template.DefaultData["version"] != "1.0.0" {
		t.Errorf("SetDefaultData() version = %v, want %v", template.DefaultData["version"], "1.0.0")
	}

	// 测试更新时间是否改变
	oldUpdatedTime := template.UpdatedTime
	time.Sleep(time.Millisecond)
	template.SetDefaultData("new_key", "new_value")

	if template.UpdatedTime == oldUpdatedTime {
		t.Error("SetDefaultData() should update UpdatedTime")
	}
}

func TestNewNotificationSubscription(t *testing.T) {
	tests := []struct {
		name             string
		userID           string
		notificationType NotificationType
		wantErr          bool
	}{
		{
			name:             "valid subscription",
			userID:           "user123",
			notificationType: NotificationTypeUser,
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subscription := NewNotificationSubscription(tt.userID, tt.notificationType)

			if subscription == nil {
				t.Error("NewNotificationSubscription() returned nil")
				return
			}

			if subscription.UserID != tt.userID {
				t.Errorf("NewNotificationSubscription() UserID = %v, want %v", subscription.UserID, tt.userID)
			}

			if subscription.Type != tt.notificationType {
				t.Errorf("NewNotificationSubscription() Type = %v, want %v", subscription.Type, tt.notificationType)
			}

			if !subscription.IsActive {
				t.Error("NewNotificationSubscription() IsActive should be true")
			}

			if subscription.Channels == nil {
				t.Error("NewNotificationSubscription() Channels should not be nil")
			}

			if len(subscription.Channels) != 1 || subscription.Channels[0] != "in_app" {
				t.Errorf("NewNotificationSubscription() Channels = %v, want %v", subscription.Channels, []string{"in_app"})
			}

			if subscription.Settings == nil {
				t.Error("NewNotificationSubscription() Settings should not be nil")
			}

			if subscription.ID == "" {
				t.Error("NewNotificationSubscription() ID should not be empty")
			}

			if subscription.CreatedTime.IsZero() {
				t.Error("NewNotificationSubscription() CreatedTime should not be zero")
			}

			if subscription.UpdatedTime.IsZero() {
				t.Error("NewNotificationSubscription() UpdatedTime should not be zero")
			}
		})
	}
}

func TestNotificationSubscription_AddChannel(t *testing.T) {
	subscription := NewNotificationSubscription("user123", NotificationTypeUser)

	// 添加渠道
	subscription.AddChannel("email")
	subscription.AddChannel("push")

	if len(subscription.Channels) != 3 { // 初始有 "in_app"
		t.Errorf("AddChannel() Channels length = %v, want %v", len(subscription.Channels), 3)
	}

	// 测试重复添加
	subscription.AddChannel("email") // 应该不会重复添加

	if len(subscription.Channels) != 3 {
		t.Errorf("AddChannel() duplicate Channels length = %v, want %v", len(subscription.Channels), 3)
	}

	// 测试更新时间是否改变
	oldUpdatedTime := subscription.UpdatedTime
	time.Sleep(time.Millisecond)
	subscription.AddChannel("sms")

	if subscription.UpdatedTime == oldUpdatedTime {
		t.Error("AddChannel() should update UpdatedTime")
	}
}

func TestNotificationSubscription_RemoveChannel(t *testing.T) {
	subscription := NewNotificationSubscription("user123", NotificationTypeUser)
	subscription.AddChannel("email")
	subscription.AddChannel("push")

	// 移除渠道
	subscription.RemoveChannel("email")

	if len(subscription.Channels) != 2 {
		t.Errorf("RemoveChannel() Channels length = %v, want %v", len(subscription.Channels), 2)
	}

	// 检查是否还存在
	for _, channel := range subscription.Channels {
		if channel == "email" {
			t.Error("RemoveChannel() should remove the channel")
		}
	}

	// 移除不存在的渠道
	subscription.RemoveChannel("nonexistent")

	if len(subscription.Channels) != 2 {
		t.Errorf("RemoveChannel() nonexistent Channels length = %v, want %v", len(subscription.Channels), 2)
	}

	// 测试更新时间是否改变
	oldUpdatedTime := subscription.UpdatedTime
	time.Sleep(time.Millisecond)
	subscription.RemoveChannel("push")

	if subscription.UpdatedTime == oldUpdatedTime {
		t.Error("RemoveChannel() should update UpdatedTime")
	}
}

func TestNotificationSubscription_SetSetting(t *testing.T) {
	subscription := NewNotificationSubscription("user123", NotificationTypeUser)

	// 设置配置
	subscription.SetSetting("frequency", "daily")
	subscription.SetSetting("quiet_hours_start", "22:00")

	if subscription.Settings["frequency"] != "daily" {
		t.Errorf("SetSetting() frequency = %v, want %v", subscription.Settings["frequency"], "daily")
	}

	if subscription.Settings["quiet_hours_start"] != "22:00" {
		t.Errorf("SetSetting() quiet_hours_start = %v, want %v", subscription.Settings["quiet_hours_start"], "22:00")
	}

	// 测试更新时间是否改变
	oldUpdatedTime := subscription.UpdatedTime
	time.Sleep(time.Millisecond)
	subscription.SetSetting("new_setting", "new_value")

	if subscription.UpdatedTime == oldUpdatedTime {
		t.Error("SetSetting() should update UpdatedTime")
	}
}
