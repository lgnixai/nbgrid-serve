package space

import (
	"context"
	"errors"
)

// Service 空间领域服务接口
type Service interface {
	CreateSpace(ctx context.Context, req CreateSpaceRequest) (*Space, error)
	GetSpace(ctx context.Context, id string) (*Space, error)
	UpdateSpace(ctx context.Context, id string, req UpdateSpaceRequest) (*Space, error)
	DeleteSpace(ctx context.Context, id string) error
	ListSpaces(ctx context.Context, filter ListFilter) ([]*Space, int64, error)

	// 协作者管理
	AddCollaborator(ctx context.Context, spaceID, userID, role string) error
	RemoveCollaborator(ctx context.Context, collabID string) error
	ListCollaborators(ctx context.Context, spaceID string) ([]*SpaceCollaborator, error)
	UpdateCollaboratorRole(ctx context.Context, collabID, role string) error

	// 批量操作
	BulkUpdateSpaces(ctx context.Context, updates []BulkUpdateRequest) error
	BulkDeleteSpaces(ctx context.Context, spaceIDs []string) error

	// 权限检查
	CheckUserPermission(ctx context.Context, spaceID, userID, permission string) (bool, error)
	GetUserSpaces(ctx context.Context, userID string, filter ListFilter) ([]*Space, int64, error)
	GetUserDeletedSpaces(ctx context.Context, userID string, filter ListFilter) ([]*Space, int64, error)

	// 统计信息
	GetSpaceStats(ctx context.Context, spaceID string) (*SpaceStats, error)
	GetUserSpaceStats(ctx context.Context, userID string) (*UserSpaceStats, error)
}

type ServiceImpl struct {
	repo                 Repository
	memberService        *MemberService
	accessControlService *AccessControlService
}

func NewService(repo Repository) Service {
	memberService := NewMemberService()
	accessControlService := NewAccessControlService(memberService)

	return &ServiceImpl{
		repo:                 repo,
		memberService:        memberService,
		accessControlService: accessControlService,
	}
}

type CreateSpaceRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=255"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=2000"`
	Icon        *string `json:"icon,omitempty" validate:"omitempty,max=100"`
	CreatedBy   string  `json:"created_by" validate:"required"`
}

type UpdateSpaceRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=2000"`
	Icon        *string `json:"icon,omitempty" validate:"omitempty,max=100"`
}

// BulkUpdateRequest 批量更新请求
type BulkUpdateRequest struct {
	SpaceID string             `json:"space_id" validate:"required"`
	Updates UpdateSpaceRequest `json:"updates" validate:"required"`
}

// SpaceStats 空间统计信息
type SpaceStats struct {
	SpaceID            string  `json:"space_id"`
	TotalBases         int64   `json:"total_bases"`
	TotalTables        int64   `json:"total_tables"`
	TotalRecords       int64   `json:"total_records"`
	TotalCollaborators int64   `json:"total_collaborators"`
	LastActivityAt     *string `json:"last_activity_at,omitempty"`
}

// UserSpaceStats 用户空间统计信息
type UserSpaceStats struct {
	UserID             string `json:"user_id"`
	TotalSpaces        int64  `json:"total_spaces"`
	OwnedSpaces        int64  `json:"owned_spaces"`
	CollaboratedSpaces int64  `json:"collaborated_spaces"`
	TotalBases         int64  `json:"total_bases"`
	TotalTables        int64  `json:"total_tables"`
}

func (s *ServiceImpl) CreateSpace(ctx context.Context, req CreateSpaceRequest) (*Space, error) {
	space, err := NewSpace(req.Name, req.CreatedBy)
	if err != nil {
		return nil, err
	}

	space.Description = req.Description
	space.Icon = req.Icon

	if err := s.repo.Create(ctx, space); err != nil {
		return nil, err
	}
	return space, nil
}

func (s *ServiceImpl) GetSpace(ctx context.Context, id string) (*Space, error) {
	sp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if sp == nil {
		return nil, errors.New("space not found")
	}
	return sp, nil
}

func (s *ServiceImpl) UpdateSpace(ctx context.Context, id string, req UpdateSpaceRequest) (*Space, error) {
	sp, err := s.GetSpace(ctx, id)
	if err != nil {
		return nil, err
	}
	sp.Update(req.Name, req.Description, req.Icon)
	if err := s.repo.Update(ctx, sp); err != nil {
		return nil, err
	}
	return sp, nil
}

func (s *ServiceImpl) DeleteSpace(ctx context.Context, id string) error {
	sp, err := s.GetSpace(ctx, id)
	if err != nil {
		return err
	}
	sp.SoftDelete()
	return s.repo.Update(ctx, sp)
}

func (s *ServiceImpl) ListSpaces(ctx context.Context, filter ListFilter) ([]*Space, int64, error) {
	items, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.Count(ctx, CountFilter{Name: filter.Name, CreatedBy: filter.CreatedBy, Search: filter.Search})
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *ServiceImpl) AddCollaborator(ctx context.Context, spaceID, userID, role string) error {
	// 获取空间信息
	space, err := s.GetSpace(ctx, spaceID)
	if err != nil {
		return err
	}

	// 验证角色
	collaboratorRole := CollaboratorRole(role)
	if !collaboratorRole.IsValid() {
		return ErrInvalidRole
	}

	// 验证邀请（这里简化处理，实际需要获取邀请者信息）
	if err := s.memberService.ValidateInvitation(space, space.CreatedBy, userID, collaboratorRole); err != nil {
		return err
	}

	// 创建协作者
	collaborator, err := NewSpaceCollaborator(spaceID, userID, collaboratorRole, space.CreatedBy)
	if err != nil {
		return err
	}

	return s.repo.AddCollaborator(ctx, collaborator)
}

func (s *ServiceImpl) RemoveCollaborator(ctx context.Context, collabID string) error {
	return s.repo.RemoveCollaborator(ctx, collabID)
}

func (s *ServiceImpl) ListCollaborators(ctx context.Context, spaceID string) ([]*SpaceCollaborator, error) {
	return s.repo.ListCollaborators(ctx, spaceID)
}

// UpdateCollaboratorRole 更新协作者角色
func (s *ServiceImpl) UpdateCollaboratorRole(ctx context.Context, collabID, role string) error {
	// TODO: 实现更新协作者角色的逻辑
	// 这里需要先获取协作者信息，然后更新角色
	return errors.New("not implemented")
}

// BulkUpdateSpaces 批量更新空间
func (s *ServiceImpl) BulkUpdateSpaces(ctx context.Context, updates []BulkUpdateRequest) error {
	for _, update := range updates {
		_, err := s.UpdateSpace(ctx, update.SpaceID, update.Updates)
		if err != nil {
			return err
		}
	}
	return nil
}

// BulkDeleteSpaces 批量删除空间
func (s *ServiceImpl) BulkDeleteSpaces(ctx context.Context, spaceIDs []string) error {
	for _, spaceID := range spaceIDs {
		if err := s.DeleteSpace(ctx, spaceID); err != nil {
			return err
		}
	}
	return nil
}

// CheckUserPermission 检查用户权限 - 重构后的版本
func (s *ServiceImpl) CheckUserPermission(ctx context.Context, spaceID, userID, permission string) (bool, error) {
	// 获取空间信息
	space, err := s.GetSpace(ctx, spaceID)
	if err != nil {
		return false, err
	}

	// 获取用户的协作者信息
	var collaborator *SpaceCollaborator
	if space.CreatedBy != userID {
		collaborators, err := s.ListCollaborators(ctx, spaceID)
		if err != nil {
			return false, err
		}

		for _, collab := range collaborators {
			if collab.UserID == userID {
				collaborator = collab
				break
			}
		}
	}

	// 使用访问控制服务检查权限
	return s.accessControlService.CheckSpaceOperation(ctx, space, userID, permission, collaborator), nil
}

// GetUserSpaces 获取用户的空间列表
func (s *ServiceImpl) GetUserSpaces(ctx context.Context, userID string, filter ListFilter) ([]*Space, int64, error) {
	// 设置过滤条件为当前用户
	filter.CreatedBy = &userID
	return s.ListSpaces(ctx, filter)
}

// GetSpaceStats 获取空间统计信息
func (s *ServiceImpl) GetSpaceStats(ctx context.Context, spaceID string) (*SpaceStats, error) {
	// TODO: 实现获取空间统计信息的逻辑
	// 这里需要查询基础表、数据表、记录和协作者的数量
	return &SpaceStats{
		SpaceID: spaceID,
		// 暂时返回默认值，需要集成其他服务
		TotalBases:         0,
		TotalTables:        0,
		TotalRecords:       0,
		TotalCollaborators: 0,
	}, nil
}

// GetUserDeletedSpaces 获取用户已删除的空间列表
func (s *ServiceImpl) GetUserDeletedSpaces(ctx context.Context, userID string, filter ListFilter) ([]*Space, int64, error) {
	// 设置过滤条件为当前用户的已删除空间
	filter.CreatedBy = &userID

	// 获取已删除的空间（需要在仓储层实现专门的查询方法）
	spaces, err := s.repo.ListDeleted(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// 统计总数
	countFilter := CountFilter{
		Name:      filter.Name,
		CreatedBy: filter.CreatedBy,
		Search:    filter.Search,
	}
	total, err := s.repo.CountDeleted(ctx, countFilter)
	if err != nil {
		return nil, 0, err
	}

	return spaces, total, nil
}

// GetUserSpaceStats 获取用户空间统计信息
func (s *ServiceImpl) GetUserSpaceStats(ctx context.Context, userID string) (*UserSpaceStats, error) {
	// TODO: 实现获取用户空间统计信息的逻辑
	return &UserSpaceStats{
		UserID: userID,
		// 暂时返回默认值，需要集成其他服务
		TotalSpaces:        0,
		OwnedSpaces:        0,
		CollaboratedSpaces: 0,
		TotalBases:         0,
		TotalTables:        0,
	}, nil
}
