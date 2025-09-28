//go:build integration

package notification

import (
	"testing"
	"time"
)

// TestNotificationIntegration 集成测试示例
// 注意：这是一个示例测试，实际的集成测试需要数据库连接
func TestNotificationIntegration(t *testing.T) {
	// 这个测试需要在有数据库连接的环境中运行
	// 使用 -tags=integration 标志运行

	t.Run("CreateAndRetrieveNotification", func(t *testing.T) {
		// 创建通知
		notification := &Notification{
			ID:      "test-notification-1",
			UserID:  "user123",
			Type:    NotificationTypeUser,
			Title:   "Integration Test",
			Content: "This is an integration test",
			Status:  NotificationStatusUnread,
		}

		// 验证通知创建成功
		if notification == nil {
			t.Fatal("Failed to create notification")
		}

		if notification.UserID != "user123" {
			t.Errorf("Expected UserID 'user123', got '%s'", notification.UserID)
		}

		if notification.Type != NotificationTypeUser {
			t.Errorf("Expected Type '%s', got '%s'", NotificationTypeUser, notification.Type)
		}

		if notification.Title != "Integration Test" {
			t.Errorf("Expected Title 'Integration Test', got '%s'", notification.Title)
		}

		if notification.Content != "This is an integration test" {
			t.Errorf("Expected Content 'This is an integration test', got '%s'", notification.Content)
		}

		// 测试状态变更
		if notification.Status != NotificationStatusUnread {
			t.Errorf("Expected Status '%s', got '%s'", NotificationStatusUnread, notification.Status)
		}

		notification.Status = NotificationStatusRead
		now := time.Now()
		notification.ReadAt = &now

		if notification.Status != NotificationStatusRead {
			t.Errorf("Expected Status '%s', got '%s'", NotificationStatusRead, notification.Status)
		}

		if notification.ReadAt == nil {
			t.Error("Expected ReadAt to be set after marking as read")
		}
	})

	t.Run("NotificationTemplateIntegration", func(t *testing.T) {
		// 创建通知模板
		template := &NotificationTemplate{
			ID:       "template-1",
			Type:     NotificationTypeUser,
			Name:     "welcome",
			Title:    "Welcome {{user_name}}",
			Content:  "Welcome to our platform, {{user_name}}!",
			IsActive: true,
		}

		// 验证模板创建成功
		if template == nil {
			t.Fatal("Failed to create notification template")
		}

		if template.Type != NotificationTypeUser {
			t.Errorf("Expected Type '%s', got '%s'", NotificationTypeUser, template.Type)
		}

		if template.Name != "welcome" {
			t.Errorf("Expected Name 'welcome', got '%s'", template.Name)
		}

		if !template.IsActive {
			t.Error("Expected template to be active by default")
		}

		// 测试添加变量
		template.Variables = []string{"user_name", "platform_name"}

		if len(template.Variables) != 2 {
			t.Errorf("Expected 2 variables, got %d", len(template.Variables))
		}

		// 测试设置默认数据
		template.DefaultData = map[string]interface{}{
			"platform_name": "Teable",
			"version":       "1.0.0",
		}

		if template.DefaultData["platform_name"] != "Teable" {
			t.Errorf("Expected platform_name 'Teable', got '%v'", template.DefaultData["platform_name"])
		}

		if template.DefaultData["version"] != "1.0.0" {
			t.Errorf("Expected version '1.0.0', got '%v'", template.DefaultData["version"])
		}
	})

	t.Run("NotificationSubscriptionIntegration", func(t *testing.T) {
		// 创建通知订阅
		subscription := &NotificationSubscription{
			ID:       "subscription-1",
			UserID:   "user123",
			Type:     NotificationTypeUser,
			IsActive: true,
			Channels: []string{"in_app"},
			Settings: make(map[string]interface{}),
		}

		// 验证订阅创建成功
		if subscription == nil {
			t.Fatal("Failed to create notification subscription")
		}

		if subscription.UserID != "user123" {
			t.Errorf("Expected UserID 'user123', got '%s'", subscription.UserID)
		}

		if subscription.Type != NotificationTypeUser {
			t.Errorf("Expected Type '%s', got '%s'", NotificationTypeUser, subscription.Type)
		}

		if !subscription.IsActive {
			t.Error("Expected subscription to be active by default")
		}

		// 验证默认渠道
		if len(subscription.Channels) != 1 || subscription.Channels[0] != "in_app" {
			t.Errorf("Expected default channel 'in_app', got %v", subscription.Channels)
		}

		// 测试添加渠道
		subscription.Channels = append(subscription.Channels, "email", "push")

		if len(subscription.Channels) != 3 {
			t.Errorf("Expected 3 channels, got %d", len(subscription.Channels))
		}

		// 测试移除渠道
		newChannels := make([]string, 0)
		for _, channel := range subscription.Channels {
			if channel != "email" {
				newChannels = append(newChannels, channel)
			}
		}
		subscription.Channels = newChannels

		if len(subscription.Channels) != 2 {
			t.Errorf("Expected 2 channels after removal, got %d", len(subscription.Channels))
		}

		// 测试设置配置
		subscription.Settings["frequency"] = "daily"
		subscription.Settings["quiet_hours_start"] = "22:00"

		if subscription.Settings["frequency"] != "daily" {
			t.Errorf("Expected frequency 'daily', got '%v'", subscription.Settings["frequency"])
		}

		if subscription.Settings["quiet_hours_start"] != "22:00" {
			t.Errorf("Expected quiet_hours_start '22:00', got '%v'", subscription.Settings["quiet_hours_start"])
		}
	})
}
