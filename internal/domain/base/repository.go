package base

import "context"

// Repository 基础表仓储接口
type Repository interface {
	// 创建基础表
	Create(ctx context.Context, base *Base) error

	// 根据ID获取基础表
	GetByID(ctx context.Context, id string) (*Base, error)

	// 更新基础表
	Update(ctx context.Context, base *Base) error

	// 删除基础表
	Delete(ctx context.Context, id string) error

	// 列出基础表
	List(ctx context.Context, filter ListFilter) ([]*Base, error)

	// 统计基础表数量
	Count(ctx context.Context, filter CountFilter) (int64, error)

	// 检查基础表是否存在
	Exists(ctx context.Context, filter ExistsFilter) (bool, error)
}

// ListFilter 列表过滤器
type ListFilter struct {
	SpaceID   *string
	Name      *string
	CreatedBy *string
	Search    string
	OrderBy   string
	Order     string
	Limit     int
	Offset    int
}

// CountFilter 计数过滤器
type CountFilter struct {
	SpaceID   *string
	Name      *string
	CreatedBy *string
	Search    string
}

// ExistsFilter 存在性检查过滤器
type ExistsFilter struct {
	ID      *string
	SpaceID *string
	Name    *string
}
