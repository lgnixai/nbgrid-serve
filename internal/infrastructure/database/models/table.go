package models

import (
	"time"

	"gorm.io/gorm"
)

// Table 数据表模型
type Table struct {
	ID               string         `gorm:"primaryKey;type:varchar(50)" json:"id"`
	BaseID           string         `gorm:"type:varchar(50);not null;index" json:"base_id"`
	Name             string         `gorm:"type:varchar(255);not null" json:"name"`
	Description      *string        `gorm:"type:text" json:"description"`
	Icon             *string        `gorm:"type:varchar(255)" json:"icon"`
	IsSystem         bool           `gorm:"default:false" json:"is_system"`
	CreatedBy        string         `gorm:"type:varchar(50);not null;index" json:"created_by"`
	CreatedTime      time.Time      `gorm:"not null" json:"created_time"`
	DeletedTime      gorm.DeletedAt `gorm:"index" json:"deleted_time"`
	LastModifiedTime *time.Time     `json:"last_modified_time"`
}

// TableName 指定表名
func (Table) TableName() string {
	return "table"
}
