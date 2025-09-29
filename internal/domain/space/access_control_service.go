package space

import (
	"context"
)

// AccessControlService 访问控制领域服务
type AccessControlService struct {
	memberService *MemberService
}

// NewAccessControlService 创建访问控制服务
func NewAccessControlService(memberService *MemberService) *AccessControlService {
	return &AccessControlService{
		memberService: memberService,
	}
}

// CheckSpaceAccess 检查用户对空间的访问权限
func (acs *AccessControlService) CheckSpaceAccess(ctx context.Context, space *Space, userID string, collaborator *SpaceCollaborator) bool {
	// 检查空间是否活跃
	if !space.IsActive() {
		return false
	}
	
	// 检查用户是否有读取权限
	return acs.memberService.CheckPermission(space, userID, PermissionRead, collaborator)
}

// CheckSpaceOperation 检查用户对空间的操作权限
func (acs *AccessControlService) CheckSpaceOperation(ctx context.Context, space *Space, userID, operation string, collaborator *SpaceCollaborator) bool {
	// 检查空间是否活跃（某些操作如恢复可以在非活跃状态下执行）
	if operation != PermissionRestore && !space.IsActive() {
		return false
	}
	
	// 检查用户是否有指定操作权限
	return acs.memberService.CheckPermission(space, userID, operation, collaborator)
}

// CheckBaseAccess 检查用户对基础表的访问权限
func (acs *AccessControlService) CheckBaseAccess(ctx context.Context, space *Space, userID string, collaborator *SpaceCollaborator) bool {
	// 首先检查空间访问权限
	if !acs.CheckSpaceAccess(ctx, space, userID, collaborator) {
		return false
	}
	
	// 基础表访问权限继承自空间权限
	return true
}

// CheckBaseOperation 检查用户对基础表的操作权限
func (acs *AccessControlService) CheckBaseOperation(ctx context.Context, space *Space, userID, operation string, collaborator *SpaceCollaborator) bool {
	// 首先检查空间操作权限
	return acs.CheckSpaceOperation(ctx, space, userID, operation, collaborator)
}

// GetUserAccessLevel 获取用户的访问级别
func (acs *AccessControlService) GetUserAccessLevel(space *Space, userID string, collaborator *SpaceCollaborator) AccessLevel {
	role := acs.memberService.GetUserRole(space, userID, collaborator)
	
	switch role {
	case RoleOwner:
		return AccessLevelOwner
	case RoleAdmin:
		return AccessLevelAdmin
	case RoleEditor:
		return AccessLevelEditor
	case RoleViewer:
		return AccessLevelViewer
	default:
		return AccessLevelNone
	}
}

// CanInviteMembers 检查用户是否可以邀请成员
func (acs *AccessControlService) CanInviteMembers(space *Space, userID string, collaborator *SpaceCollaborator) bool {
	return acs.memberService.CheckPermission(space, userID, PermissionManageMembers, collaborator)
}

// CanManageSettings 检查用户是否可以管理设置
func (acs *AccessControlService) CanManageSettings(space *Space, userID string, collaborator *SpaceCollaborator) bool {
	return acs.memberService.CheckPermission(space, userID, PermissionManageSettings, collaborator)
}

// CanTransferOwnership 检查用户是否可以转移所有权
func (acs *AccessControlService) CanTransferOwnership(space *Space, userID string) bool {
	// 只有空间所有者可以转移所有权
	return space.CreatedBy == userID
}

// AccessLevel 访问级别枚举
type AccessLevel string

const (
	AccessLevelNone   AccessLevel = "none"
	AccessLevelViewer AccessLevel = "viewer"
	AccessLevelEditor AccessLevel = "editor"
	AccessLevelAdmin  AccessLevel = "admin"
	AccessLevelOwner  AccessLevel = "owner"
)

// IsValid 检查访问级别是否有效
func (al AccessLevel) IsValid() bool {
	switch al {
	case AccessLevelNone, AccessLevelViewer, AccessLevelEditor, AccessLevelAdmin, AccessLevelOwner:
		return true
	default:
		return false
	}
}

// CanRead 检查是否可以读取
func (al AccessLevel) CanRead() bool {
	return al != AccessLevelNone
}

// CanWrite 检查是否可以写入
func (al AccessLevel) CanWrite() bool {
	return al == AccessLevelEditor || al == AccessLevelAdmin || al == AccessLevelOwner
}

// CanManage 检查是否可以管理
func (al AccessLevel) CanManage() bool {
	return al == AccessLevelAdmin || al == AccessLevelOwner
}

// CanOwn 检查是否拥有所有权
func (al AccessLevel) CanOwn() bool {
	return al == AccessLevelOwner
}