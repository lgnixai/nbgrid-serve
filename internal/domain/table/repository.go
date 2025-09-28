package table

import "context"

// Repository 数据表仓储接口
type Repository interface {
	// 表格操作
	CreateTable(ctx context.Context, table *Table) error
	GetTableByID(ctx context.Context, id string) (*Table, error)
	UpdateTable(ctx context.Context, table *Table) error
	DeleteTable(ctx context.Context, id string) error
	ListTables(ctx context.Context, filter ListTableFilter) ([]*Table, error)
	CountTables(ctx context.Context, filter ListTableFilter) (int64, error)
	ExistsTable(ctx context.Context, filter ListTableFilter) (bool, error)

	// 字段操作
	CreateField(ctx context.Context, field *Field) error
	GetFieldByID(ctx context.Context, id string) (*Field, error)
	UpdateField(ctx context.Context, field *Field) error
	DeleteField(ctx context.Context, id string) error
	ListFields(ctx context.Context, filter ListFieldFilter) ([]*Field, error)
	CountFields(ctx context.Context, filter ListFieldFilter) (int64, error)
	ExistsField(ctx context.Context, filter ListFieldFilter) (bool, error)
}
