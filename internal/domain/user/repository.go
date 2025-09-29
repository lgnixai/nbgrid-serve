package user

import (
	"context"
)

// Repository 用户仓储接口
type Repository interface {
	// 基础CRUD操作
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByPhone(ctx context.Context, phone string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error

	// 查询操作
	List(ctx context.Context, filter ListFilter) ([]*User, error)
	Count(ctx context.Context, filter CountFilter) (int64, error)
	Exists(ctx context.Context, filter ExistsFilter) (bool, error)

	// 批量操作
	BatchCreate(ctx context.Context, users []*User) error
	BatchUpdate(ctx context.Context, users []*User) error
	BatchDelete(ctx context.Context, ids []string) error

	// 账户相关
	CreateAccount(ctx context.Context, account *Account) error
	GetAccountsByUserID(ctx context.Context, userID string) ([]*Account, error)
	GetAccountByProvider(ctx context.Context, provider, providerID string) (*Account, error)
	DeleteAccount(ctx context.Context, id string) error
}

// ListFilter 列表过滤器
type ListFilter struct {
	// 分页参数
	Offset int
	Limit  int

	// 排序参数
	OrderBy string
	Order   string // ASC, DESC

	// 过滤条件
	Name           *string
	Email          *string
	IsActive       *bool
	IsAdmin        *bool
	IsSystem       *bool
	CreatedAfter   *string
	CreatedBefore  *string
	ModifiedAfter  *string
	ModifiedBefore *string

	// 搜索
	Search string
}

// CountFilter 计数过滤器
type CountFilter struct {
	Name           *string
	Email          *string
	IsActive       *bool
	IsAdmin        *bool
	IsSystem       *bool
	CreatedAfter   *string
	CreatedBefore  *string
	ModifiedAfter  *string
	ModifiedBefore *string
	Search         string
}

// ExistsFilter 存在性检查过滤器
type ExistsFilter struct {
	ID         *string
	Email      *string
	Phone      *string
	Provider   *string
	ProviderID *string
}

// 分页结果
type PaginatedResult struct {
	Users      []*User
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

// 默认过滤器构造函数
func NewListFilter() ListFilter {
	return ListFilter{
		Offset:  0,
		Limit:   20,
		OrderBy: "created_time",
		Order:   "DESC",
	}
}

func NewCountFilter() CountFilter {
	return CountFilter{}
}

func NewExistsFilter() ExistsFilter {
	return ExistsFilter{}
}
