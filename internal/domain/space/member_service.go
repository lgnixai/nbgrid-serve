package space

import (
	"fmt"
	"time"
)

// MemberService 空间成员管理领域服务
type MemberService struct{}

// NewMemberService 创建成员管理服务
func NewMemberService() *MemberService {
	return &MemberService{}
}

// ValidateInvitation 验证邀请的有效性
func (ms *MemberService) ValidateInvitation(space *Space, inviterID, inviteeID string, role CollaboratorRole) error {
	// 检查空间状态
	if !space.IsActive() {
		return DomainError{Code: "SPACE_NOT_ACTIVE", Message: "cannot invite members to inactive space"}
	}
	
	// 检查邀请者权限
	if space.CreatedBy != inviterID {
		// TODO: 需要检查邀请者是否有管理成员的权限
		// 这里需要通过聚合根或仓储查询邀请者的协作者信息
		return DomainError{Code: "INSUFFICIENT_PERMISSION", Message: "insufficient permission to invite members"}
	}
	
	// 检查角色有效性
	if !role.IsValid() {
		return ErrInvalidRole
	}
	
	// 检查是否邀请自己
	if inviterID == inviteeID {
		return DomainError{Code: "CANNOT_INVITE_SELF", Message: "cannot invite yourself"}
	}
	
	return nil
}

// ValidateRoleUpdate 验证角色更新的有效性
func (ms *MemberService) ValidateRoleUpdate(updaterRole CollaboratorRole, targetRole, newRole CollaboratorRole) error {
	// 检查更新者是否有权限管理目标角色
	if !updaterRole.CanManageRole(targetRole) {
		return DomainError{Code: "INSUFFICIENT_PERMISSION", Message: "insufficient permission to manage this role"}
	}
	
	// 检查更新者是否有权限设置新角色
	if !updaterRole.CanManageRole(newRole) {
		return DomainError{Code: "INSUFFICIENT_PERMISSION", Message: "insufficient permission to assign this role"}
	}
	
	return nil
}

// ValidateRemoval 验证移除成员的有效性
func (ms *MemberService) ValidateRemoval(removerRole, targetRole CollaboratorRole, isOwnerRemoval bool) error {
	// 不能移除空间所有者
	if isOwnerRemoval {
		return DomainError{Code: "CANNOT_REMOVE_OWNER", Message: "cannot remove space owner"}
	}
	
	// 检查移除者是否有权限管理目标角色
	if !removerRole.CanManageRole(targetRole) {
		return DomainError{Code: "INSUFFICIENT_PERMISSION", Message: "insufficient permission to remove this member"}
	}
	
	return nil
}

// CalculatePermissions 计算用户在空间中的有效权限
func (ms *MemberService) CalculatePermissions(space *Space, userID string, collaborator *SpaceCollaborator) []string {
	// 如果是空间创建者，拥有所有权限
	if space.CreatedBy == userID {
		return RoleOwner.GetPermissions()
	}
	
	// 如果是活跃的协作者，返回角色权限
	if collaborator != nil && collaborator.IsActive() {
		return collaborator.GetPermissions()
	}
	
	// 否则没有权限
	return []string{}
}

// CheckPermission 检查用户是否有指定权限
func (ms *MemberService) CheckPermission(space *Space, userID, permission string, collaborator *SpaceCollaborator) bool {
	permissions := ms.CalculatePermissions(space, userID, collaborator)
	
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	
	return false
}

// GetUserRole 获取用户在空间中的角色
func (ms *MemberService) GetUserRole(space *Space, userID string, collaborator *SpaceCollaborator) CollaboratorRole {
	// 如果是空间创建者，角色为Owner
	if space.CreatedBy == userID {
		return RoleOwner
	}
	
	// 如果是活跃的协作者，返回协作者角色
	if collaborator != nil && collaborator.IsActive() {
		return collaborator.Role
	}
	
	// 否则没有角色
	return ""
}

// ValidateOwnershipTransfer 验证所有权转移的有效性
func (ms *MemberService) ValidateOwnershipTransfer(currentOwnerID, newOwnerID string, newOwnerCollaborator *SpaceCollaborator) error {
	// 检查新所有者是否是当前活跃的协作者
	if newOwnerCollaborator == nil || !newOwnerCollaborator.IsActive() {
		return DomainError{Code: "INVALID_NEW_OWNER", Message: "new owner must be an active collaborator"}
	}
	
	// 检查是否转移给自己
	if currentOwnerID == newOwnerID {
		return DomainError{Code: "CANNOT_TRANSFER_TO_SELF", Message: "cannot transfer ownership to yourself"}
	}
	
	return nil
}

// GenerateInviteToken 生成邀请令牌
func (ms *MemberService) GenerateInviteToken(spaceID, userID string) string {
	// 简单的令牌生成逻辑，实际应该使用更安全的方法
	return fmt.Sprintf("invite_%s_%s_%d", spaceID, userID, getCurrentTimestamp())
}

// ValidateInviteToken 验证邀请令牌
func (ms *MemberService) ValidateInviteToken(token, expectedSpaceID, expectedUserID string) bool {
	// 简单的令牌验证逻辑，实际应该使用更安全的方法
	expectedToken := fmt.Sprintf("invite_%s_%s_", expectedSpaceID, expectedUserID)
	return len(token) > len(expectedToken) && token[:len(expectedToken)] == expectedToken
}

// getCurrentTimestamp 获取当前时间戳
func getCurrentTimestamp() int64 {
	return getCurrentTime().Unix()
}

// getCurrentTime 获取当前时间
func getCurrentTime() time.Time {
	return time.Now()
}