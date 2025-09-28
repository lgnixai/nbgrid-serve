package base

import (
	"teable-go-backend/pkg/utils"
	"time"
)

// Base 基础表实体
type Base struct {
	ID               string     `json:"id"`
	SpaceID          string     `json:"space_id"`
	Name             string     `json:"name"`
	Description      *string    `json:"description"`
	Icon             *string    `json:"icon"`
	CreatedBy        string     `json:"created_by"`
	CreatedTime      time.Time  `json:"created_at"`
	DeletedTime      *time.Time `json:"deleted_time"`
	LastModifiedTime *time.Time `json:"updated_at"`
}

// NewBase 创建新的基础表
func NewBase(spaceID, name, createdBy string) *Base {
	now := time.Now()
	return &Base{
		ID:               utils.GenerateBaseID(),
		SpaceID:          spaceID,
		Name:             name,
		CreatedBy:        createdBy,
		CreatedTime:      now,
		LastModifiedTime: &now,
	}
}

// Update 更新基础表信息
func (b *Base) Update(name string, description *string, icon *string) {
	b.Name = name
	b.Description = description
	b.Icon = icon
	now := time.Now()
	b.LastModifiedTime = &now
}

// SoftDelete 软删除基础表
func (b *Base) SoftDelete() {
	now := time.Now()
	b.DeletedTime = &now
	b.LastModifiedTime = &now
}

// IsDeleted 检查是否已删除
func (b *Base) IsDeleted() bool {
	return b.DeletedTime != nil
}

// Validate 验证基础表数据
func (b *Base) Validate() error {
	if b.SpaceID == "" {
		return ErrInvalidSpaceID
	}
	if b.Name == "" {
		return ErrInvalidName
	}
	if b.CreatedBy == "" {
		return ErrInvalidCreatedBy
	}
	return nil
}
