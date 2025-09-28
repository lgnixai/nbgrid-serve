package user

import (
	"errors"
	"teable-go-backend/pkg/utils"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User 用户实体
type User struct {
	ID                   string
	Name                 string
	Email                string
	Password             *string
	Salt                 *string
	Phone                *string
	Avatar               *string
	IsSystem             bool
	IsAdmin              bool
	IsTrialUsed          bool
	NotifyMeta           *string
	LastSignTime         *time.Time
	DeactivatedTime      *time.Time
	CreatedTime          time.Time
	DeletedTime          *time.Time
	LastModifiedTime     *time.Time
	PermanentDeletedTime *time.Time
	RefMeta              *string
}

// Account 第三方账户实体
type Account struct {
	ID          string
	UserID      string
	Type        string
	Provider    string
	ProviderID  string
	CreatedTime time.Time
}

// 用户状态枚举
type UserStatus string

const (
	UserStatusActive      UserStatus = "active"
	UserStatusDeactivated UserStatus = "deactivated"
	UserStatusDeleted     UserStatus = "deleted"
)

// 账户类型枚举
type AccountType string

const (
	AccountTypeLocal  AccountType = "local"
	AccountTypeOAuth  AccountType = "oauth"
	AccountTypeSocial AccountType = "social"
)

// 提供商枚举
type Provider string

const (
	ProviderLocal  Provider = "local"
	ProviderGitHub Provider = "github"
	ProviderGoogle Provider = "google"
	ProviderOIDC   Provider = "oidc"
)

// 业务规则错误 - 这些错误会被转换为应用错误
var (
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrWeakPassword    = errors.New("password is too weak")
	ErrEmailExists     = errors.New("email already exists")
	ErrPhoneExists     = errors.New("phone already exists")
	ErrUserNotFound    = errors.New("user not found")
	ErrUserDeactivated = errors.New("user is deactivated")
	ErrUserDeleted     = errors.New("user is deleted")
	ErrInvalidPassword = errors.New("invalid password")
)

// NewUser 创建新用户
func NewUser(name, email string) (*User, error) {
	if !isValidEmail(email) {
		return nil, ErrInvalidEmail
	}

	now := time.Now()
	return &User{
		ID:               generateUserID(),
		Name:             name,
		Email:            email,
		IsSystem:         false,
		IsAdmin:          false,
		IsTrialUsed:      false,
		CreatedTime:      now,
		LastModifiedTime: &now,
	}, nil
}

// NewUserWithPassword 创建带密码的新用户
func NewUserWithPassword(name, email, password string) (*User, error) {
	user, err := NewUser(name, email)
	if err != nil {
		return nil, err
	}

	if err := user.SetPassword(password); err != nil {
		return nil, err
	}

	return user, nil
}

// SetPassword 设置密码
func (u *User) SetPassword(password string) error {
	if !isValidPassword(password) {
		return ErrWeakPassword
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	passwordStr := string(hashedPassword)
	u.Password = &passwordStr
	u.updateModifiedTime()

	return nil
}

// CheckPassword 验证密码
func (u *User) CheckPassword(password string) error {
	if u.Password == nil {
		return ErrInvalidPassword
	}

	return bcrypt.CompareHashAndPassword([]byte(*u.Password), []byte(password))
}

// IsActive 检查用户是否激活
func (u *User) IsActive() bool {
	return u.DeactivatedTime == nil && u.DeletedTime == nil
}

// GetStatus 获取用户状态
func (u *User) GetStatus() UserStatus {
	if u.DeletedTime != nil {
		return UserStatusDeleted
	}
	if u.DeactivatedTime != nil {
		return UserStatusDeactivated
	}
	return UserStatusActive
}

// Deactivate 停用用户
func (u *User) Deactivate() {
	now := time.Now()
	u.DeactivatedTime = &now
	u.updateModifiedTime()
}

// Activate 激活用户
func (u *User) Activate() {
	u.DeactivatedTime = nil
	u.updateModifiedTime()
}

// SoftDelete 软删除用户
func (u *User) SoftDelete() {
	now := time.Now()
	u.DeletedTime = &now
	u.updateModifiedTime()
}

// UpdateProfile 更新用户资料
func (u *User) UpdateProfile(name, phone *string, avatar *string) {
	if name != nil {
		u.Name = *name
	}
	if phone != nil {
		u.Phone = phone
	}
	if avatar != nil {
		u.Avatar = avatar
	}
	u.updateModifiedTime()
}

// PromoteToAdmin 提升为管理员
func (u *User) PromoteToAdmin() {
	u.IsAdmin = true
	u.updateModifiedTime()
}

// DemoteFromAdmin 取消管理员
func (u *User) DemoteFromAdmin() {
	u.IsAdmin = false
	u.updateModifiedTime()
}

// MarkTrialUsed 标记试用已使用
func (u *User) MarkTrialUsed() {
	u.IsTrialUsed = true
	u.updateModifiedTime()
}

// RecordSignIn 记录登录时间
func (u *User) RecordSignIn() {
	now := time.Now()
	u.LastSignTime = &now
	u.updateModifiedTime()
}

// GetDisplayName 获取显示名称
func (u *User) GetDisplayName() string {
	if u.Name != "" {
		return u.Name
	}
	return u.Email
}

// HasPermission 检查权限
func (u *User) HasPermission(permission string) bool {
	// 系统用户拥有所有权限
	if u.IsSystem {
		return true
	}

	// 管理员拥有管理权限
	if u.IsAdmin && isAdminPermission(permission) {
		return true
	}

	// TODO: 实现更细粒度的权限检查
	return false
}

// AddAccount 添加第三方账户
func (u *User) AddAccount(accountType AccountType, provider Provider, providerID string) *Account {
	return &Account{
		ID:          generateAccountID(),
		UserID:      u.ID,
		Type:        string(accountType),
		Provider:    string(provider),
		ProviderID:  providerID,
		CreatedTime: time.Now(),
	}
}

// updateModifiedTime 更新修改时间
func (u *User) updateModifiedTime() {
	now := time.Now()
	u.LastModifiedTime = &now
}

// 辅助函数

// isValidEmail 验证邮箱格式
func isValidEmail(email string) bool {
	// 简单的邮箱验证，实际应该使用更严格的正则表达式
	return len(email) > 0 &&
		len(email) <= 255 &&
		containsChar(email, '@') &&
		containsChar(email, '.')
}

// isValidPassword 验证密码强度
func isValidPassword(password string) bool {
	// 密码至少8位，包含字母和数字
	if len(password) < 8 {
		return false
	}

	hasLetter := false
	hasDigit := false

	for _, char := range password {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
			hasLetter = true
		}
		if char >= '0' && char <= '9' {
			hasDigit = true
		}
	}

	return hasLetter && hasDigit
}

// containsChar 检查字符串是否包含字符
func containsChar(str string, char rune) bool {
	for _, c := range str {
		if c == char {
			return true
		}
	}
	return false
}

// isAdminPermission 检查是否为管理员权限
func isAdminPermission(permission string) bool {
	adminPermissions := []string{
		"user:manage",
		"space:manage",
		"base:manage",
		"system:config",
	}

	for _, perm := range adminPermissions {
		if perm == permission {
			return true
		}
	}
	return false
}

// generateUserID 生成用户ID
func generateUserID() string {
	return utils.GenerateUserID()
}

// generateAccountID 生成账户ID
func generateAccountID() string {
	return utils.GenerateAccountID()
}

// generateNanoID 生成NanoID(简化实现)
func generateNanoID(length int) string {
	const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	b := make([]byte, length)
	for i := range b {
		b[i] = alphabet[i%len(alphabet)] // 简化实现，实际应该使用crypto/rand
	}
	return string(b)
}
