package models

import (
	"time"
)

// Permission 权限模型
type Permission struct {
	ID           string     `gorm:"primaryKey;type:varchar(30)" json:"id"`
	UserID       string     `gorm:"not null;type:varchar(30);index" json:"user_id"`
	ResourceType string     `gorm:"not null;type:varchar(50);index" json:"resource_type"`
	ResourceID   string     `gorm:"not null;type:varchar(30);index" json:"resource_id"`
	Role         string     `gorm:"not null;type:varchar(50)" json:"role"`
	GrantedBy    string     `gorm:"not null;type:varchar(30)" json:"granted_by"`
	GrantedAt    time.Time  `gorm:"not null" json:"granted_at"`
	ExpiresAt    *time.Time `gorm:"index" json:"expires_at,omitempty"`
	IsActive     bool       `gorm:"default:true;index" json:"is_active"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// 复合索引
	// UNIQUE(user_id, resource_type, resource_id)
}

// TableName 指定表名
func (Permission) TableName() string {
	return "permissions"
}

// BaseCollaborator 基础表协作者模型
type BaseCollaborator struct {
	ID        string     `gorm:"primaryKey;type:varchar(30)" json:"id"`
	BaseID    string     `gorm:"not null;type:varchar(30);index" json:"base_id"`
	UserID    string     `gorm:"not null;type:varchar(30);index" json:"user_id"`
	Role      string     `gorm:"not null;type:varchar(50)" json:"role"`
	Email     *string    `gorm:"type:varchar(255)" json:"email,omitempty"`
	InvitedBy string     `gorm:"not null;type:varchar(30)" json:"invited_by"`
	InvitedAt time.Time  `gorm:"not null" json:"invited_at"`
	JoinedAt  *time.Time `json:"joined_at,omitempty"`
	IsActive  bool       `gorm:"default:true;index" json:"is_active"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// 复合索引
	// UNIQUE(base_id, user_id)
}

// TableName 指定表名
func (BaseCollaborator) TableName() string {
	return "base_collaborators"
}
