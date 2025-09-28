package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// View 视图模型
type View struct {
	ID               string         `gorm:"primaryKey;type:varchar(50)" json:"id"`
	TableID          string         `gorm:"type:varchar(50);not null;index" json:"table_id"`
	Name             string         `gorm:"type:varchar(255);not null" json:"name"`
	Description      *string        `gorm:"type:text" json:"description"`
	Type             string         `gorm:"type:varchar(50);not null" json:"type"`
	Config           string         `gorm:"type:jsonb" json:"config"` // 存储为JSON格式
	IsDefault        bool           `gorm:"default:false" json:"is_default"`
	CreatedBy        string         `gorm:"type:varchar(50);not null;index" json:"created_by"`
	CreatedTime      time.Time      `gorm:"not null" json:"created_time"`
	DeletedTime      gorm.DeletedAt `gorm:"index" json:"deleted_time"`
	LastModifiedTime *time.Time     `json:"last_modified_time"`
}

// TableName 指定表名
func (View) TableName() string {
	return "view"
}

// GetConfigAsMap 将JSON配置转换为map
func (v *View) GetConfigAsMap() (map[string]interface{}, error) {
	var config map[string]interface{}
	if v.Config != "" {
		err := json.Unmarshal([]byte(v.Config), &config)
		if err != nil {
			return nil, err
		}
	}
	return config, nil
}

// SetConfigFromMap 将map配置转换为JSON存储
func (v *View) SetConfigFromMap(config map[string]interface{}) error {
	if config == nil {
		v.Config = "{}"
		return nil
	}

	jsonConfig, err := json.Marshal(config)
	if err != nil {
		return err
	}
	v.Config = string(jsonConfig)
	return nil
}
