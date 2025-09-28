package table

import (
	"time"

	"teable-go-backend/pkg/utils"
)

// Table 数据表实体
type Table struct {
	ID               string     `json:"id"`
	BaseID           string     `json:"base_id"`
	Name             string     `json:"name"`
	Description      *string    `json:"description"`
	Icon             *string    `json:"icon"`
	IsSystem         bool       `json:"is_system"`
	CreatedBy        string     `json:"created_by"`
	CreatedTime      time.Time  `json:"created_at"`
	DeletedTime      *time.Time `json:"deleted_time"`
	LastModifiedTime *time.Time `json:"updated_at"`
}

// Field 字段实体
type Field struct {
	ID               string     `json:"id"`
	TableID          string     `json:"table_id"`
	Name             string     `json:"name"`
	Type             string     `json:"type"`
	Description      *string    `json:"description"`
	IsRequired       bool       `json:"required"`
	IsUnique         bool       `json:"is_unique"`
	IsPrimary        bool       `json:"is_primary"`
	DefaultValue     *string    `json:"default_value"`
	Options          *string    `json:"options"` // JSON格式的选项配置
	FieldOrder       int        `json:"field_order"`
	CreatedBy        string     `json:"created_by"`
	CreatedTime      time.Time  `json:"created_at"`
	DeletedTime      *time.Time `json:"deleted_time"`
	LastModifiedTime *time.Time `json:"updated_at"`
}

// CreateTableRequest 创建数据表请求
type CreateTableRequest struct {
	BaseID      string
	Name        string
	Description *string
	Icon        *string
	CreatedBy   string
}

// UpdateTableRequest 更新数据表请求
type UpdateTableRequest struct {
	Name        *string
	Description *string
	Icon        *string
}

// CreateFieldRequest 创建字段请求
type CreateFieldRequest struct {
	TableID      string  `json:"table_id"`
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	Description  *string `json:"description"`
	IsRequired   bool    `json:"required"`
	IsUnique     bool    `json:"is_unique"`
	IsPrimary    bool    `json:"is_primary"`
	DefaultValue *string `json:"default_value"`
	Options      *string `json:"options"`
	FieldOrder   int     `json:"field_order"`
	CreatedBy    string  `json:"created_by"`
}

// UpdateFieldRequest 更新字段请求
type UpdateFieldRequest struct {
	Name         *string
	Type         *string
	Description  *string
	IsRequired   *bool
	IsUnique     *bool
	IsPrimary    *bool
	DefaultValue *string
	Options      *string
	FieldOrder   *int
}

// ListTableFilter 表格列表过滤条件
type ListTableFilter struct {
	BaseID    *string `form:"base_id" json:"base_id"`
	Name      *string `form:"name" json:"name"`
	CreatedBy *string `form:"created_by" json:"created_by"`
	Search    string  `form:"search" json:"search"`
	OrderBy   string  `form:"order_by" json:"order_by"`
	Order     string  `form:"order" json:"order"`
	Limit     int     `form:"limit" json:"limit"`
	Offset    int     `form:"offset" json:"offset"`
}

// ListFieldFilter 字段列表过滤条件
type ListFieldFilter struct {
	TableID   *string `form:"table_id" json:"table_id"`
	Name      *string `form:"name" json:"name"`
	Type      *string `form:"type" json:"type"`
	CreatedBy *string `form:"created_by" json:"created_by"`
	OrderBy   string  `form:"order_by" json:"order_by"`
	Order     string  `form:"order" json:"order"`
	Limit     int     `form:"limit" json:"limit"`
	Offset    int     `form:"offset" json:"offset"`
}

// BulkUpdateTableRequest 批量更新表格请求
type BulkUpdateTableRequest struct {
	TableID string             `json:"table_id" validate:"required"`
	Updates UpdateTableRequest `json:"updates" validate:"required"`
}

// BulkUpdateFieldRequest 批量更新字段请求
type BulkUpdateFieldRequest struct {
	FieldID string             `json:"field_id" validate:"required"`
	Updates UpdateFieldRequest `json:"updates" validate:"required"`
}

// TableStats 表格统计信息
type TableStats struct {
	TableID        string  `json:"table_id"`
	TotalFields    int64   `json:"total_fields"`
	TotalRecords   int64   `json:"total_records"`
	TotalViews     int64   `json:"total_views"`
	LastActivityAt *string `json:"last_activity_at,omitempty"`
}

// BaseTableStats 基础表表格统计信息
type BaseTableStats struct {
	BaseID       string `json:"base_id"`
	TotalTables  int64  `json:"total_tables"`
	TotalFields  int64  `json:"total_fields"`
	TotalRecords int64  `json:"total_records"`
	TotalViews   int64  `json:"total_views"`
}

// NewTable 创建新的数据表
func NewTable(req CreateTableRequest) *Table {
	now := time.Now()
	return &Table{
		ID:               utils.GenerateTableID(),
		BaseID:           req.BaseID,
		Name:             req.Name,
		Description:      req.Description,
		Icon:             req.Icon,
		IsSystem:         false,
		CreatedBy:        req.CreatedBy,
		CreatedTime:      now,
		LastModifiedTime: &now,
	}
}

// NewField 创建新的字段
func NewField(req CreateFieldRequest) *Field {
	now := time.Now()
	return &Field{
		ID:               utils.GenerateFieldID(),
		TableID:          req.TableID,
		Name:             req.Name,
		Type:             req.Type,
		Description:      req.Description,
		IsRequired:       req.IsRequired,
		IsUnique:         req.IsUnique,
		IsPrimary:        req.IsPrimary,
		DefaultValue:     req.DefaultValue,
		Options:          req.Options,
		FieldOrder:       req.FieldOrder,
		CreatedBy:        req.CreatedBy,
		CreatedTime:      now,
		LastModifiedTime: &now,
	}
}

// Update 更新数据表信息
func (t *Table) Update(req UpdateTableRequest) {
	if req.Name != nil {
		t.Name = *req.Name
	}
	if req.Description != nil {
		t.Description = req.Description
	}
	if req.Icon != nil {
		t.Icon = req.Icon
	}
	now := time.Now()
	t.LastModifiedTime = &now
}

// Update 更新字段信息
func (f *Field) Update(req UpdateFieldRequest) {
	if req.Name != nil {
		f.Name = *req.Name
	}
	if req.Type != nil {
		f.Type = *req.Type
	}
	if req.Description != nil {
		f.Description = req.Description
	}
	if req.IsRequired != nil {
		f.IsRequired = *req.IsRequired
	}
	if req.IsUnique != nil {
		f.IsUnique = *req.IsUnique
	}
	if req.IsPrimary != nil {
		f.IsPrimary = *req.IsPrimary
	}
	if req.DefaultValue != nil {
		f.DefaultValue = req.DefaultValue
	}
	if req.Options != nil {
		f.Options = req.Options
	}
	if req.FieldOrder != nil {
		f.FieldOrder = *req.FieldOrder
	}
	now := time.Now()
	f.LastModifiedTime = &now
}

// SoftDelete 软删除数据表
func (t *Table) SoftDelete() {
	now := time.Now()
	t.DeletedTime = &now
	t.LastModifiedTime = &now
}

// SoftDelete 软删除字段
func (f *Field) SoftDelete() {
	now := time.Now()
	f.DeletedTime = &now
	f.LastModifiedTime = &now
}
