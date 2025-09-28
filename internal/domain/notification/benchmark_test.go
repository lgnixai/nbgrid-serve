package notification

import (
	"testing"
)

// BenchmarkNewNotification 测试创建通知的性能
func BenchmarkNewNotification(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewNotification("user123", NotificationTypeUser, "Test Title", "Test Content")
	}
}

// BenchmarkNotification_SetData 测试设置数据的性能
func BenchmarkNotification_SetData(b *testing.B) {
	notification := NewNotification("user123", NotificationTypeUser, "Test Title", "Test Content")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		notification.SetData("key", "value")
	}
}

// BenchmarkNotification_MarkAsRead 测试标记为已读的性能
func BenchmarkNotification_MarkAsRead(b *testing.B) {
	notification := NewNotification("user123", NotificationTypeUser, "Test Title", "Test Content")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		notification.MarkAsRead()
	}
}

// BenchmarkNewNotificationTemplate 测试创建通知模板的性能
func BenchmarkNewNotificationTemplate(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewNotificationTemplate(NotificationTypeUser, "template", "Title", "Content")
	}
}

// BenchmarkNewNotificationSubscription 测试创建通知订阅的性能
func BenchmarkNewNotificationSubscription(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewNotificationSubscription("user123", NotificationTypeUser)
	}
}

// BenchmarkNotificationSubscription_AddChannel 测试添加渠道的性能
func BenchmarkNotificationSubscription_AddChannel(b *testing.B) {
	subscription := NewNotificationSubscription("user123", NotificationTypeUser)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		subscription.AddChannel("email")
	}
}
