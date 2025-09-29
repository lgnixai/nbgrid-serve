package builders

import (
	"time"

	"teable-go-backend/internal/domain/space"
	"teable-go-backend/internal/domain/table"
	"teable-go-backend/internal/domain/user"
	"teable-go-backend/pkg/utils"
)

// UserBuilder 用户构建器
type UserBuilder struct {
	user *user.User
}

// NewUserBuilder 创建用户构建器
func NewUserBuilder() *UserBuilder {
	return &UserBuilder{
		user: &user.User{
			ID:               utils.GenerateUserID(),
			Name:             "Test User",
			Email:            "test@example.com",
			IsAdmin:          false,
			IsSystem:         false,
			CreatedTime:      time.Now(),
			LastModifiedTime: &[]time.Time{time.Now()}[0],
		},
	}
}

// WithID 设置用户ID
func (b *UserBuilder) WithID(id string) *UserBuilder {
	b.user.ID = id
	return b
}

// WithName 设置用户名
func (b *UserBuilder) WithName(name string) *UserBuilder {
	b.user.Name = name
	return b
}

// WithEmail 设置邮箱
func (b *UserBuilder) WithEmail(email string) *UserBuilder {
	b.user.Email = email
	return b
}

// WithPassword 设置密码
func (b *UserBuilder) WithPassword(password string) *UserBuilder {
	b.user.SetPassword(password)
	return b
}

// AsAdmin 设置为管理员
func (b *UserBuilder) AsAdmin() *UserBuilder {
	b.user.IsAdmin = true
	return b
}

// AsSystem 设置为系统用户
func (b *UserBuilder) AsSystem() *UserBuilder {
	b.user.IsSystem = true
	return b
}

// AsInactive 设置为非活跃用户
func (b *UserBuilder) AsInactive() *UserBuilder {
	// 注意：User结构体可能没有IsActive字段，这里暂时注释掉
	// b.user.IsActive = false
	return b
}

// Build 构建用户
func (b *UserBuilder) Build() *user.User {
	return b.user
}

// SpaceBuilder 空间构建器
type SpaceBuilder struct {
	space *space.Space
}

// NewSpaceBuilder 创建空间构建器
func NewSpaceBuilder() *SpaceBuilder {
	return &SpaceBuilder{
		space: &space.Space{
			ID:               utils.GenerateSpaceID(),
			Name:             "Test Space",
			CreatedBy:        utils.GenerateUserID(),
			CreatedTime:      time.Now(),
			LastModifiedTime: &[]time.Time{time.Now()}[0],
		},
	}
}

// WithID 设置空间ID
func (b *SpaceBuilder) WithID(id string) *SpaceBuilder {
	b.space.ID = id
	return b
}

// WithName 设置空间名称
func (b *SpaceBuilder) WithName(name string) *SpaceBuilder {
	b.space.Name = name
	return b
}

// WithCreatedBy 设置创建者
func (b *SpaceBuilder) WithCreatedBy(userID string) *SpaceBuilder {
	b.space.CreatedBy = userID
	return b
}

// Build 构建空间
func (b *SpaceBuilder) Build() *space.Space {
	return b.space
}

// TableBuilder 表构建器
type TableBuilder struct {
	table *table.Table
}

// NewTableBuilder 创建表构建器
func NewTableBuilder() *TableBuilder {
	return &TableBuilder{
		table: &table.Table{
			ID:               utils.GenerateTableID(),
			Name:             "Test Table",
			BaseID:           utils.GenerateBaseID(),
			CreatedBy:        utils.GenerateUserID(),
			CreatedTime:      time.Now(),
			LastModifiedTime: &[]time.Time{time.Now()}[0],
		},
	}
}

// WithID 设置表ID
func (b *TableBuilder) WithID(id string) *TableBuilder {
	b.table.ID = id
	return b
}

// WithName 设置表名称
func (b *TableBuilder) WithName(name string) *TableBuilder {
	b.table.Name = name
	return b
}

// WithBaseID 设置基础表ID
func (b *TableBuilder) WithBaseID(baseID string) *TableBuilder {
	b.table.BaseID = baseID
	return b
}

// WithCreatedBy 设置创建者
func (b *TableBuilder) WithCreatedBy(userID string) *TableBuilder {
	b.table.CreatedBy = userID
	return b
}

// Build 构建表
func (b *TableBuilder) Build() *table.Table {
	return b.table
}

// FieldBuilder 字段构建器
type FieldBuilder struct {
	field *table.Field
}

// NewFieldBuilder 创建字段构建器
func NewFieldBuilder() *FieldBuilder {
	return &FieldBuilder{
		field: &table.Field{
			ID:               utils.GenerateFieldID(),
			Name:             "Test Field",
			Type:             table.FieldTypeText,
			TableID:          utils.GenerateTableID(),
			CreatedTime:      time.Now(),
			LastModifiedTime: &[]time.Time{time.Now()}[0],
		},
	}
}

// WithID 设置字段ID
func (b *FieldBuilder) WithID(id string) *FieldBuilder {
	b.field.ID = id
	return b
}

// WithName 设置字段名称
func (b *FieldBuilder) WithName(name string) *FieldBuilder {
	b.field.Name = name
	return b
}

// WithType 设置字段类型
func (b *FieldBuilder) WithType(fieldType table.FieldType) *FieldBuilder {
	b.field.Type = fieldType
	return b
}

// WithTableID 设置表ID
func (b *FieldBuilder) WithTableID(tableID string) *FieldBuilder {
	b.field.TableID = tableID
	return b
}

// AsPrimary 设置为主键
func (b *FieldBuilder) AsPrimary() *FieldBuilder {
	b.field.IsPrimary = true
	return b
}

// Build 构建字段
func (b *FieldBuilder) Build() *table.Field {
	return b.field
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
