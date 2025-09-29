package space

import (
	"regexp"
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
	ErrInvalidSpaceName     = DomainError{Code: "INVALID_SPACE_NAME", Message: "invalid space name"}
	ErrSpaceNameTooLong     = DomainError{Code: "SPACE_NAME_TOO_LONG", Message: "space name too long"}
	ErrSpaceNameEmpty       = DomainError{Code: "SPACE_NAME_EMPTY", Message: "space name cannot be empty"}
	ErrInvalidIcon          = DomainError{Code: "INVALID_ICON", Message: "invalid icon format"}
	ErrInvalidCreatedBy     = DomainError{Code: "INVALID_CREATED_BY", Message: "invalid created by user ID"}
	ErrSpaceDeleted         = DomainError{Code: "SPACE_DELETED", Message: "space is deleted"}
	ErrInvalidRole          = DomainError{Code: "INVALID_ROLE", Message: "invalid collabor role"}
	ErrCollaboratorExists = DomainError{Code: "COLLABORATOR_EXISTS", Message: "collaborator already exists"}
)

// Space 空间领域实体 - 重构后的版本
type Space struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	Description      *string    `json:"description"`
	Icon             *string    `json:"icon"`
	CreatedBy        string     `json:"created_by"`
	CreatedTime      time.Time  `json:"created_at"`
	DeletedTime      *time.Time `json:"deleted_time"`
	LastModifiedTime *time.Time `json:"updated_at"`
	
	// 业务状态
	status SpaceStatus `json:"status"`
	
	// 成员管理相关
	memberCount int `json:"-"` // 不序列化到JSON，由聚合根管理
}

// NewSpace 创建新空间 - 重构后的版本
func NewSpace(name string, createdBy string) (*Space, error) {
	// 验证输入参数
	if err := validateSpaceName(name); err != nil {
		return nil, err
	}
	
	if err := validateUserID(createdBy); err != nil {
		return nil, err
	}

	now := time.Now()
	return &Space{
		ID:               utils.GenerateSpaceID(),
		Name:             name,
		CreatedBy:        createdBy,
		CreatedTime:      now,
		LastModifiedTime: &now,
		status:           SpaceStatusActive,
		memberCount:      1, // 创建者自动成为第一个成员
	}, nil
}

// Update 更新空间信息
func (s *Space) Update(name *string, description *string, icon *string) error {
	if name != nil {
		if err := validateSpaceName(*name); err != nil {
			return err
		}
		s.Name = *name
	}
	
	if description != nil {
		if err := validateDescription(*description); err != nil {
			return err
		}
		s.Description = description
	}
	
	if icon != nil {
		if err := validateIcon(*icon); err != nil {
			return err
		}
		s.Icon = icon
	}
	
	s.updateModifiedTime()
	return nil
}

// SoftDelete 软删除空间
func (s *Space) SoftDelete() {
	now := time.Now()
	s.DeletedTime = &now
	s.updateModifiedTime()
}

// IsDeleted 检查空间是否已删除
func (s *Space) IsDeleted() bool {
	return s.DeletedTime != nil
}

// ValidateForUpdate 验证空间是否可以被更新
func (s *Space) ValidateForUpdate() error {
	if s.IsDeleted() {
		return ErrSpaceDeleted
	}
	return nil
}

// ValidateForDeletion 验证空间是否可以被删除
func (s *Space) ValidateForDeletion() error {
	if s.IsDeleted() {
		return ErrSpaceDeleted
	}
	return nil
}

// CanUserAccess 检查用户是否可以访问空间
func (s *Space) CanUserAccess(userID string) bool {
	if s.IsDeleted() || s.status != SpaceStatusActive {
		return false
	}
	
	// 创建者可以访问
	if s.CreatedBy == userID {
		return true
	}
	
	// TODO: 检查协作者权限 - 需要通过聚合根或领域服务检查
	return false
}

// Archive 归档空间
func (s *Space) Archive() error {
	if s.IsDeleted() {
		return ErrSpaceDeleted
	}
	
	if s.status == SpaceStatusArchived {
		return DomainError{Code: "SPACE_ALREADY_ARCHIVED", Message: "space is already archived"}
	}
	
	s.status = SpaceStatusArchived
	s.updateModifiedTime()
	return nil
}

// Restore 恢复空间
func (s *Space) Restore() error {
	if s.IsDeleted() {
		return ErrSpaceDeleted
	}
	
	if s.status == SpaceStatusActive {
		return DomainError{Code: "SPACE_ALREADY_ACTIVE", Message: "space is already active"}
	}
	
	s.status = SpaceStatusActive
	s.updateModifiedTime()
	return nil
}

// GetStatus 获取空间状态
func (s *Space) GetStatus() SpaceStatus {
	if s.IsDeleted() {
		return SpaceStatusDeleted
	}
	return s.status
}

// IsArchived 检查空间是否已归档
func (s *Space) IsArchived() bool {
	return s.status == SpaceStatusArchived
}

// IsActive 检查空间是否活跃
func (s *Space) IsActive() bool {
	return s.status == SpaceStatusActive && !s.IsDeleted()
}

// GetMemberCount 获取成员数量
func (s *Space) GetMemberCount() int {
	return s.memberCount
}

// UpdateMemberCount 更新成员数量（由聚合根调用）
func (s *Space) UpdateMemberCount(count int) {
	if count >= 1 { // 至少有创建者一个成员
		s.memberCount = count
		s.updateModifiedTime()
	}
}

// updateModifiedTime 更新修改时间
func (s *Space) updateModifiedTime() {
	now := time.Now()
	s.LastModifiedTime = &now
}

// SpaceStatus 空间状态枚举
type SpaceStatus string

const (
	SpaceStatusActive   SpaceStatus = "active"
	SpaceStatusArchived SpaceStatus = "archived"
	SpaceStatusDeleted  SpaceStatus = "deleted"
)

// IsValid 检查空间状态是否有效
func (s SpaceStatus) IsValid() bool {
	switch s {
	case SpaceStatusActive, SpaceStatusArchived, SpaceStatusDeleted:
		return true
	default:
		return false
	}
}

// 协作者角色枚举
type CollaboratorRole string

const (
	RoleOwner  CollaboratorRole = "owner"
	RoleAdmin  CollaboratorRole = "admin"
	RoleEditor CollaboratorRole = "editor"
	RoleViewer CollaboratorRole = "viewer"
)

// String 实现Stringer接口
func (r CollaboratorRole) String() string {
	return string(r)
}

// IsValid 检查角色是否有效
func (r CollaboratorRole) IsValid() bool {
	switch r {
	case RoleOwner, RoleAdmin, RoleEditor, RoleViewer:
		return true
	default:
		return false
	}
}

// 权限常量定义
const (
	PermissionRead             = "read"
	PermissionWrite            = "write"
	PermissionCreate           = "create"
	PermissionUpdate           = "update"
	PermissionDelete           = "delete"
	PermissionManageMembers    = "manage_members"
	PermissionManageSettings   = "manage_settings"
	PermissionTransferOwnership = "transfer_ownership"
	PermissionArchive          = "archive"
	PermissionRestore          = "restore"
)

// HasPermission 检查角色是否有指定权限 - 重构后的版本
func (r CollaboratorRole) HasPermission(permission string) bool {
	permissions := r.GetPermissions()
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// GetPermissions 获取角色的所有权限
func (r CollaboratorRole) GetPermissions() []string {
	switch r {
	case RoleOwner:
		return []string{
			PermissionRead, PermissionWrite, PermissionCreate, PermissionUpdate, PermissionDelete,
			PermissionManageMembers, PermissionManageSettings, PermissionTransferOwnership,
			PermissionArchive, PermissionRestore,
		}
	case RoleAdmin:
		return []string{
			PermissionRead, PermissionWrite, PermissionCreate, PermissionUpdate, PermissionDelete,
			PermissionManageMembers, PermissionManageSettings,
			PermissionArchive, PermissionRestore,
		}
	case RoleEditor:
		return []string{
			PermissionRead, PermissionWrite, PermissionCreate, PermissionUpdate,
		}
	case RoleViewer:
		return []string{
			PermissionRead,
		}
	default:
		return []string{}
	}
}

// CanManageRole 检查当前角色是否可以管理目标角色
func (r CollaboratorRole) CanManageRole(targetRole CollaboratorRole) bool {
	// 只有Owner可以管理Admin
	if targetRole == RoleAdmin {
		return r == RoleOwner
	}
	
	// Owner和Admin可以管理Editor和Viewer
	if targetRole == RoleEditor || targetRole == RoleViewer {
		return r == RoleOwner || r == RoleAdmin
	}
	
	// 只有Owner可以管理Owner
	if targetRole == RoleOwner {
		return r == RoleOwner
	}
	
	return false
}

// CollaboratorStatus 协作者状态枚举
type CollaboratorStatus string

const (
	CollaboratorStatusPending  CollaboratorStatus = "pending"
	CollaboratorStatusAccepted CollaboratorStatus = "accepted"
	CollaboratorStatusRejected CollaboratorStatus = "rejected"
	CollaboratorStatusRevoked  CollaboratorStatus = "revoked"
)

// IsValid 检查协作者状态是否有效
func (cs CollaboratorStatus) IsValid() bool {
	switch cs {
	case CollaboratorStatusPending, CollaboratorStatusAccepted, CollaboratorStatusRejected, CollaboratorStatusRevoked:
		return true
	default:
		return false
	}
}

// SpaceCollaborator 空间协作者实体 - 重构后的版本
type SpaceCollaborator struct {
	ID          string             `json:"id"`
	SpaceID     string             `json:"space_id"`
	UserID      string             `json:"user_id"`
	Role        CollaboratorRole   `json:"role"`
	InvitedBy   string             `json:"invited_by"`
	CreatedTime time.Time          `json:"created_time"`
	AcceptedAt  *time.Time         `json:"accepted_at"`
	RevokedAt   *time.Time         `json:"revoked_at"`
	Status      CollaboratorStatus `json:"status"`
	
	// 权限相关
	permissions []string `json:"-"` // 不序列化，由角色动态计算
}

// NewSpaceCollaborator 创建新的空间协作者 - 重构后的版本
func NewSpaceCollaborator(spaceID, userID string, role CollaboratorRole, invitedBy string) (*SpaceCollaborator, error) {
	// 验证参数
	if err := validateUserID(spaceID); err != nil {
		return nil, err
	}
	
	if err := validateUserID(userID); err != nil {
		return nil, err
	}
	
	if !role.IsValid() {
		return nil, ErrInvalidRole
	}
	
	if err := validateUserID(invitedBy); err != nil {
		return nil, err
	}

	collaborator := &SpaceCollaborator{
		ID:          utils.GenerateIDWithPrefix("spcusr"),
		SpaceID:     spaceID,
		UserID:      userID,
		Role:        role,
		InvitedBy:   invitedBy,
		CreatedTime: time.Now(),
		Status:      CollaboratorStatusPending,
	}
	
	// 初始化权限
	collaborator.refreshPermissions()
	
	return collaborator, nil
}

// Accept 接受邀请 - 重构后的版本
func (sc *SpaceCollaborator) Accept() error {
	if sc.Status != CollaboratorStatusPending {
		return DomainError{Code: "INVALID_COLLABORATOR_STATUS", Message: "can only accept pending invitations"}
	}
	
	now := time.Now()
	sc.AcceptedAt = &now
	sc.Status = CollaboratorStatusAccepted
	sc.refreshPermissions()
	return nil
}

// Reject 拒绝邀请 - 重构后的版本
func (sc *SpaceCollaborator) Reject() error {
	if sc.Status != CollaboratorStatusPending {
		return DomainError{Code: "INVALID_COLLABORATOR_STATUS", Message: "can only reject pending invitations"}
	}
	
	sc.Status = CollaboratorStatusRejected
	return nil
}

// Revoke 撤销协作者权限
func (sc *SpaceCollaborator) Revoke() error {
	if sc.Status == CollaboratorStatusRevoked {
		return DomainError{Code: "COLLABORATOR_ALREADY_REVOKED", Message: "collaborator is already revoked"}
	}
	
	now := time.Now()
	sc.RevokedAt = &now
	sc.Status = CollaboratorStatusRevoked
	sc.permissions = nil // 清空权限
	return nil
}

// UpdateRole 更新角色 - 重构后的版本
func (sc *SpaceCollaborator) UpdateRole(newRole CollaboratorRole) error {
	if !newRole.IsValid() {
		return ErrInvalidRole
	}
	
	if sc.Status != CollaboratorStatusAccepted {
		return DomainError{Code: "COLLABORATOR_NOT_ACTIVE", Message: "can only update role for active collaborators"}
	}
	
	sc.Role = newRole
	sc.refreshPermissions()
	return nil
}

// HasPermission 检查协作者是否有指定权限 - 重构后的版本
func (sc *SpaceCollaborator) HasPermission(permission string) bool {
	if sc.Status != CollaboratorStatusAccepted {
		return false
	}
	
	// 检查缓存的权限列表
	for _, p := range sc.permissions {
		if p == permission {
			return true
		}
	}
	
	return false
}

// IsActive 检查协作者是否活跃 - 重构后的版本
func (sc *SpaceCollaborator) IsActive() bool {
	return sc.Status == CollaboratorStatusAccepted
}

// IsPending 检查协作者是否待处理
func (sc *SpaceCollaborator) IsPending() bool {
	return sc.Status == CollaboratorStatusPending
}

// IsRevoked 检查协作者是否已撤销
func (sc *SpaceCollaborator) IsRevoked() bool {
	return sc.Status == CollaboratorStatusRevoked
}

// GetPermissions 获取权限列表
func (sc *SpaceCollaborator) GetPermissions() []string {
	return sc.permissions
}

// refreshPermissions 刷新权限列表
func (sc *SpaceCollaborator) refreshPermissions() {
	sc.permissions = sc.Role.GetPermissions()
}

// 验证函数

// validateSpaceName 验证空间名称
func validateSpaceName(name string) error {
	if len(name) == 0 {
		return ErrSpaceNameEmpty
	}
	if len(name) > 255 {
		return ErrSpaceNameTooLong
	}
	
	// 检查是否包含非法字符
	for _, char := range name {
		if char < 32 || char == 127 { // 控制字符
			return DomainError{Code: "INVALID_NAME_CHARS", Message: "space name contains invalid characters"}
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
	
	// 简单的图标格式验证（可以是emoji或图标名称）
	if len(icon) > 10 {
		// 如果长度超过10，可能是URL，验证URL格式
		urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
		if !urlRegex.MatchString(icon) {
			return ErrInvalidIcon
		}
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
	
	// 简单的ID格式验证
	idRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !idRegex.MatchString(userID) {
		return ErrInvalidCreatedBy
	}
	
	return nil
}


