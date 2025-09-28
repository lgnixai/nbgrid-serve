package models

import (
	"time"

	"gorm.io/gorm"
)

// Notification 通知模型
type Notification struct {
	ID          string    `gorm:"primaryKey;type:varchar(20)" json:"id"`
	UserID      string    `gorm:"type:varchar(20);not null;index" json:"user_id"`
	Type        string    `gorm:"type:varchar(50);not null;index" json:"type"`
	Title       string    `gorm:"type:varchar(255);not null" json:"title"`
	Content     string    `gorm:"type:text;not null" json:"content"`
	Data        string    `gorm:"type:json" json:"data"` // JSON格式存储
	Status      string    `gorm:"type:varchar(20);not null;default:'unread';index" json:"status"`
	Priority    string    `gorm:"type:varchar(20);not null;default:'normal';index" json:"priority"`
	SourceID    string    `gorm:"type:varchar(20);index" json:"source_id"`
	SourceType  string    `gorm:"type:varchar(50);index" json:"source_type"`
	ActionURL   string    `gorm:"type:varchar(500)" json:"action_url"`
	ExpiresAt   *time.Time `gorm:"index" json:"expires_at"`
	ReadAt      *time.Time `json:"read_at"`
	CreatedTime time.Time `gorm:"autoCreateTime;index" json:"created_time"`
	UpdatedTime time.Time `gorm:"autoUpdateTime" json:"updated_time"`
}

// TableName 返回表名
func (Notification) TableName() string {
	return "notifications"
}

// NotificationTemplate 通知模板模型
type NotificationTemplate struct {
	ID          string    `gorm:"primaryKey;type:varchar(20)" json:"id"`
	Type        string    `gorm:"type:varchar(50);not null;uniqueIndex" json:"type"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Title       string    `gorm:"type:varchar(255);not null" json:"title"`
	Content     string    `gorm:"type:text;not null" json:"content"`
	Variables   string    `gorm:"type:json" json:"variables"` // JSON格式存储
	DefaultData string    `gorm:"type:json" json:"default_data"` // JSON格式存储
	IsActive    bool      `gorm:"default:true;index" json:"is_active"`
	CreatedTime time.Time `gorm:"autoCreateTime" json:"created_time"`
	UpdatedTime time.Time `gorm:"autoUpdateTime" json:"updated_time"`
}

// TableName 返回表名
func (NotificationTemplate) TableName() string {
	return "notification_templates"
}

// NotificationSubscription 通知订阅模型
type NotificationSubscription struct {
	ID          string    `gorm:"primaryKey;type:varchar(20)" json:"id"`
	UserID      string    `gorm:"type:varchar(20);not null;index" json:"user_id"`
	Type        string    `gorm:"type:varchar(50);not null;index" json:"type"`
	SourceID    string    `gorm:"type:varchar(20);index" json:"source_id"`
	SourceType  string    `gorm:"type:varchar(50);index" json:"source_type"`
	Channels    string    `gorm:"type:json;not null" json:"channels"` // JSON格式存储
	Settings    string    `gorm:"type:json" json:"settings"` // JSON格式存储
	IsActive    bool      `gorm:"default:true;index" json:"is_active"`
	CreatedTime time.Time `gorm:"autoCreateTime" json:"created_time"`
	UpdatedTime time.Time `gorm:"autoUpdateTime" json:"updated_time"`
}

// TableName 返回表名
func (NotificationSubscription) TableName() string {
	return "notification_subscriptions"
}

// BeforeCreate 创建前钩子
func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == "" {
		// 这里应该生成ID，但通常由应用层处理
	}
	return nil
}

// BeforeCreate 创建前钩子
func (nt *NotificationTemplate) BeforeCreate(tx *gorm.DB) error {
	if nt.ID == "" {
		// 这里应该生成ID，但通常由应用层处理
	}
	return nil
}

// BeforeCreate 创建前钩子
func (ns *NotificationSubscription) BeforeCreate(tx *gorm.DB) error {
	if ns.ID == "" {
		// 这里应该生成ID，但通常由应用层处理
	}
	return nil
}
