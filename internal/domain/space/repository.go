package space

import "context"

// Repository 空间仓储接口
type Repository interface {
	Create(ctx context.Context, space *Space) error
	GetByID(ctx context.Context, id string) (*Space, error)
	Update(ctx context.Context, space *Space) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter ListFilter) ([]*Space, error)
	Count(ctx context.Context, filter CountFilter) (int64, error)

	// 协作者
    AddCollaborator(ctx context.Context, collab *SpaceCollaborator) error
    RemoveCollaborator(ctx context.Context, id string) error
    ListCollaborators(ctx context.Context, spaceID string) ([]*SpaceCollaborator, error)
}

type ListFilter struct {
	Offset int
	Limit  int
	OrderBy string
	Order   string

	Name *string
	CreatedBy *string

	Search string
}

type CountFilter struct {
	Name *string
	CreatedBy *string
	Search string
}

func NewListFilter() ListFilter {
	return ListFilter{
		Offset: 0,
		Limit: 20,
		OrderBy: "created_time",
		Order: "DESC",
	}
}

func NewCountFilter() CountFilter { return CountFilter{} }


