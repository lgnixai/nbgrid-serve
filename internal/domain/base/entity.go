package base

import (
	"time"

	"teable-go-backend/pkg/utils"
)

// 领域错误定义
type DomainError struct {
	Code    string
	Message string
}

func (e DomainError) Error() string {
	return e.Message
}

// 业务规则错误
var (
	ErrInvalidBaseName  = DomainError{Code: "INVALID_BASE_NAME", Message: "invalid base name"}
	ErrBaseNameTooLong  = DomainError{Code: "BASE_NAME_TOO_LONG", Message: "base name too long"}
	ErrBaseNameEmpty    = DomainError{Code: "BASE_NAME_EMPTY", Message: "base name cannot be empty"}
	ErrInvalidSpaceID   = DomainError{Code: "INVALID_SPACE_ID", Message: "invalid space ID"}
	ErrInvalidCreatedBy = DomainError{Code: "INVALID_CREATED_BY", Message: "invalid created by user ID"}
	ErrBaseDeleted      = DomainError{Code: "BASE_DELETED", Message: "base is deleted"}
	ErrInvalidIcon      = DomainError{Code: "INVALID_ICON", Message: "invalid icon format"}
	ErrBaseExists       = DomainError{Code: "BASE_EXISTS", Message: "base already exists"}
)

// BaseStatus 基础表状态枚举
type BaseStatus string

const (
	BaseStatusActive   BaseStatus = "active"
	BaseStatusArchived BaseStatus = "archived"
	BaseStatusDeleted  BaseStatus = "deleted"
)

// IsValid 检查基础表状态是否有效
func (bs BaseStatus) IsValid() bool {
	switch bs {
	case BaseStatusActive, BaseStatusArchived, BaseStatusDeleted:
		return true
	default:
		return false
	}
}

// Base 基础表领域实体 - 重构后的版本
type Base struct {
	ID               string     `json:"id"`
	SpaceID          string     `json:"space_id"`
	Name             string     `json:"name"`
	Description      *string    `json:"description"`
	Icon             *string    `json:"icon"`
	IsSystem         bool       `json:"is_system"`
	CreatedBy        string     `json:"created_by"`
	CreatedTime      time.Time  `json:"created_at"`
	DeletedTime      *time.Time `json:"deleted_time"`
	LastModifiedTime *time.Time `json:"updated_at"`
	
	// 业务状态
	status BaseStatus `json:"status"`
	
	// 统计信息
	tableCount int `json:"-"` // 不序列化，由聚合根管理
}

// NewBase 创建新的基础表 - 重构后的版本
func NewBase(spaceID, name, createdBy string) (*Base, error) {
	// 验证输入参数
	if err := validateSpaceID(spaceID); err != nil {
		return nil, err
	}
	
	if err := validateBaseName(name); err != nil {
		return nil, err
	}
	
	if err := validateUserID(createdBy); err != nil {
		return nil, err
	}

	now := time.Now()
	return &Base{
		ID:               utils.GenerateBaseID(),
		SpaceID:          spaceID,
		Name:             name,
		IsSystem:         false,
		CreatedBy:        createdBy,
		CreatedTime:      now,
		LastModifiedTime: &now,
		status:           BaseStatusActive,
		tableCount:       0,
	}, nil
}

// Update 更新基础表信息
func (b *Base) Update(name *string, description *string, icon *string) error {
	if name != nil {
		if err := validateBaseName(*name); err != nil {
			return err
		}
		b.Name = *name
	}
	
	if description != nil {
		if err := validateDescription(*description); err != nil {
			return err
		}
		b.Description = description
	}
	
	if icon != nil {
		if err := validateIcon(*icon); err != nil {
			return err
		}
		b.Icon = icon
	}
	
	b.updateModifiedTime()
	return nil
}

// SoftDelete 软删除基础表 - 重构后的版本
func (b *Base) SoftDelete() {
	now := time.Now()
	b.DeletedTime = &now
	b.LastModifiedTime = &now
	b.status = BaseStatusDeleted
}

// IsDeleted 检查是否已删除
func (b *Base) IsDeleted() bool {
	return b.DeletedTime != nil
}

// Archive 归档基础表
func (b *Base) Archive() error {
	if b.IsDeleted() {
		return ErrBaseDeleted
	}
	
	if b.status == BaseStatusArchived {
		return DomainError{Code: "BASE_ALREADY_ARCHIVED", Message: "base is already archived"}
	}
	
	b.status = BaseStatusArchived
	b.updateModifiedTime()
	return nil
}

// Restore 恢复基础表
func (b *Base) Restore() error {
	if b.IsDeleted() {
		return ErrBaseDeleted
	}
	
	if b.status == BaseStatusActive {
		return DomainError{Code: "BASE_ALREADY_ACTIVE", Message: "base is already active"}
	}
	
	b.status = BaseStatusActive
	b.updateModifiedTime()
	return nil
}

// GetStatus 获取基础表状态
func (b *Base) GetStatus() BaseStatus {
	if b.IsDeleted() {
		return BaseStatusDeleted
	}
	return b.status
}

// IsArchived 检查基础表是否已归档
func (b *Base) IsArchived() bool {
	return b.status == BaseStatusArchived
}

// IsActive 检查基础表是否活跃
func (b *Base) IsActive() bool {
	return b.status == BaseStatusActive && !b.IsDeleted()
}

// GetTableCount 获取表格数量
func (b *Base) GetTableCount() int {
	return b.tableCount
}

// UpdateTableCount 更新表格数量（由聚合根调用）
func (b *Base) UpdateTableCount(count int) {
	if count >= 0 {
		b.tableCount = count
		b.updateModifiedTime()
	}
}

// updateModifiedTime 更新修改时间
func (b *Base) updateModifiedTime() {
	now := time.Now()
	b.LastModifiedTime = &now
}

// Validate 验证基础表数据 - 重构后的版本
func (b *Base) Validate() error {
	if b.SpaceID == "" {
		return ErrInvalidSpaceID
	}
	if b.Name == "" {
		return ErrBaseNameEmpty
	}
	if b.CreatedBy == "" {
		return ErrInvalidCreatedBy
	}
	if !b.status.IsValid() {
		return DomainError{Code: "INVALID_BASE_STATUS", Message: "invalid base status"}
	}
	return nil
}

// ValidateForUpdate 验证基础表是否可以被更新
func (b *Base) ValidateForUpdate() error {
	if b.IsDeleted() {
		return ErrBaseDeleted
	}
	return nil
}

// ValidateForDeletion 验证基础表是否可以被删除
func (b *Base) ValidateForDeletion() error {
	if b.IsDeleted() {
		return ErrBaseDeleted
	}
	
	// 系统基础表不能删除
	if b.IsSystem {
		return DomainError{Code: "CANNOT_DELETE_SYSTEM_BASE", Message: "cannot delete system base"}
	}
	
	return nil
}
// 验证函数

// validateSpaceID 验证空间ID
func validateSpaceID(spaceID string) error {
	if len(spaceID) == 0 {
		return ErrInvalidSpaceID
	}
	if len(spaceID) > 50 {
		return DomainError{Code: "SPACE_ID_TOO_LONG", Message: "space ID too long"}
	}
	return nil
}

// validateBaseName 验证基础表名称
func validateBaseName(name string) error {
	if len(name) == 0 {
		return ErrBaseNameEmpty
	}
	if len(name) > 255 {
		return ErrBaseNameTooLong
	}
	
	// 检查是否包含非法字符
	for _, char := range name {
		if char < 32 || char == 127 { // 控制字符
			return DomainError{Code: "INVALID_NAME_CHARS", Message: "base name contains invalid characters"}
		}
	}
	
	return nil
}

// validateDescription 验证描述
func validateDescription(description string) error {
	if len(description) > 2000 {
		return DomainError{Code: "DESCRIPTION_TOO_LONG", Message: "description cannot exceed 2000 characters"}
	}
	return nil
}

// validateIcon 验证图标
func validateIcon(icon string) error {
	if len(icon) == 0 {
		return nil // 图标可以为空
	}
	if len(icon) > 100 {
		return DomainError{Code: "ICON_TOO_LONG", Message: "icon cannot exceed 100 characters"}
	}
	return nil
}

// validateUserID 验证用户ID格式
func validateUserID(userID string) error {
	if len(userID) == 0 {
		return DomainError{Code: "EMPTY_USER_ID", Message: "user ID cannot be empty"}
	}
	if len(userID) > 50 {
		return DomainError{Code: "USER_ID_TOO_LONG", Message: "user ID too long"}
	}
	return nil
}