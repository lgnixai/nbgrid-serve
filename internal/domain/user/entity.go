package user

import (
	"regexp"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"
	
	"teable-go-backend/pkg/utils"
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

// 安全级别枚举
type SecurityLevel int

const (
	SecurityLevelUser SecurityLevel = iota
	SecurityLevelAdmin
	SecurityLevelSystem
)

func (s SecurityLevel) String() string {
	switch s {
	case SecurityLevelUser:
		return "user"
	case SecurityLevelAdmin:
		return "admin"
	case SecurityLevelSystem:
		return "system"
	default:
		return "unknown"
	}
}

// 领域错误定义
type DomainError struct {
	Code    string
	Message string
}

func (e DomainError) Error() string {
	return e.Message
}

// 业务规则错误 - 纯领域错误，不依赖外部包
var (
	ErrInvalidEmail    = DomainError{Code: "INVALID_EMAIL", Message: "invalid email format"}
	ErrWeakPassword    = DomainError{Code: "WEAK_PASSWORD", Message: "password is too weak"}
	ErrEmailExists     = DomainError{Code: "EMAIL_EXISTS", Message: "email already exists"}
	ErrPhoneExists     = DomainError{Code: "PHONE_EXISTS", Message: "phone already exists"}
	ErrUserNotFound    = DomainError{Code: "USER_NOT_FOUND", Message: "user not found"}
	ErrUserDeactivated = DomainError{Code: "USER_DEACTIVATED", Message: "user is deactivated"}
	ErrUserDeleted     = DomainError{Code: "USER_DELETED", Message: "user is deleted"}
	ErrInvalidPassword = DomainError{Code: "INVALID_PASSWORD", Message: "invalid password"}
	ErrInvalidUserID   = DomainError{Code: "INVALID_USER_ID", Message: "invalid user ID format"}
	ErrInvalidPhone    = DomainError{Code: "INVALID_PHONE", Message: "invalid phone number format"}
)

// NewUser 创建新用户
func NewUser(name, email string) (*User, error) {
	// 验证输入参数
	if err := validateUserName(name); err != nil {
		return nil, err
	}
	
	if err := validateEmail(email); err != nil {
		return nil, err
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
	if err := validatePassword(password); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return DomainError{Code: "PASSWORD_HASH_FAILED", Message: "failed to hash password"}
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
func (u *User) UpdateProfile(name, phone *string, avatar *string) error {
	if name != nil {
		if err := validateUserName(*name); err != nil {
			return err
		}
		u.Name = *name
	}
	
	if phone != nil {
		if err := validatePhone(*phone); err != nil {
			return err
		}
		u.Phone = phone
	}
	
	if avatar != nil {
		if err := validateAvatar(*avatar); err != nil {
			return err
		}
		u.Avatar = avatar
	}
	
	u.updateModifiedTime()
	return nil
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
	// 已删除或停用的用户没有任何权限
	if !u.IsActive() {
		return false
	}

	// 系统用户拥有所有权限
	if u.IsSystem {
		return true
	}

	// 管理员拥有管理权限
	if u.IsAdmin && isAdminPermission(permission) {
		return true
	}

	// 基础用户权限
	if isBasicUserPermission(permission) {
		return true
	}

	return false
}

// CanAccessResource 检查是否可以访问资源
func (u *User) CanAccessResource(resourceType, resourceID string) bool {
	if !u.IsActive() {
		return false
	}

	// 系统用户可以访问所有资源
	if u.IsSystem {
		return true
	}

	// TODO: 实现基于资源的权限检查
	// 这里需要与权限服务集成
	return false
}

// ValidateForUpdate 验证用户是否可以被更新
func (u *User) ValidateForUpdate() error {
	if u.DeletedTime != nil {
		return ErrUserDeleted
	}
	
	// 系统用户不能被普通操作修改
	if u.IsSystem {
		return DomainError{Code: "SYSTEM_USER_READONLY", Message: "system user cannot be modified"}
	}
	
	return nil
}

// ValidateForDeletion 验证用户是否可以被删除
func (u *User) ValidateForDeletion() error {
	if u.DeletedTime != nil {
		return ErrUserDeleted
	}
	
	// 系统用户不能被删除
	if u.IsSystem {
		return DomainError{Code: "SYSTEM_USER_UNDELETABLE", Message: "system user cannot be deleted"}
	}
	
	return nil
}

// GetSecurityLevel 获取用户安全级别
func (u *User) GetSecurityLevel() SecurityLevel {
	if u.IsSystem {
		return SecurityLevelSystem
	}
	if u.IsAdmin {
		return SecurityLevelAdmin
	}
	return SecurityLevelUser
}

// IsPasswordExpired 检查密码是否过期
func (u *User) IsPasswordExpired(maxAge time.Duration) bool {
	if u.LastModifiedTime == nil {
		return false
	}
	
	// 如果密码修改时间超过最大年龄，则认为过期
	return time.Since(*u.LastModifiedTime) > maxAge
}

// ShouldChangePassword 检查是否应该更改密码
func (u *User) ShouldChangePassword() bool {
	// 新用户或没有密码的用户应该设置密码
	if u.Password == nil {
		return true
	}
	
	// 检查密码是否过期（90天）
	return u.IsPasswordExpired(90 * 24 * time.Hour)
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

// 验证函数

// validateUserName 验证用户名
func validateUserName(name string) error {
	if len(name) == 0 {
		return DomainError{Code: "EMPTY_NAME", Message: "user name cannot be empty"}
	}
	if len(name) > 100 {
		return DomainError{Code: "NAME_TOO_LONG", Message: "user name cannot exceed 100 characters"}
	}
	
	// 检查是否包含非法字符
	for _, char := range name {
		if char < 32 || char == 127 { // 控制字符
			return DomainError{Code: "INVALID_NAME_CHARS", Message: "user name contains invalid characters"}
		}
	}
	
	return nil
}

// validateEmail 验证邮箱格式
func validateEmail(email string) error {
	if len(email) == 0 {
		return DomainError{Code: "EMPTY_EMAIL", Message: "email cannot be empty"}
	}
	if len(email) > 255 {
		return DomainError{Code: "EMAIL_TOO_LONG", Message: "email cannot exceed 255 characters"}
	}
	
	// 使用正则表达式验证邮箱格式
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}
	
	return nil
}

// validatePassword 验证密码强度
func validatePassword(password string) error {
	if len(password) < 8 {
		return DomainError{Code: "PASSWORD_TOO_SHORT", Message: "password must be at least 8 characters long"}
	}
	if len(password) > 128 {
		return DomainError{Code: "PASSWORD_TOO_LONG", Message: "password cannot exceed 128 characters"}
	}

	var (
		hasLower   = false
		hasUpper   = false
		hasDigit   = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// 至少包含3种类型的字符
	charTypes := 0
	if hasLower {
		charTypes++
	}
	if hasUpper {
		charTypes++
	}
	if hasDigit {
		charTypes++
	}
	if hasSpecial {
		charTypes++
	}

	if charTypes < 3 {
		return DomainError{Code: "PASSWORD_TOO_WEAK", Message: "password must contain at least 3 types of characters (lowercase, uppercase, digit, special)"}
	}

	return nil
}

// validatePhone 验证手机号格式
func validatePhone(phone string) error {
	if len(phone) == 0 {
		return nil // 手机号可以为空
	}
	if len(phone) > 50 {
		return DomainError{Code: "PHONE_TOO_LONG", Message: "phone number cannot exceed 50 characters"}
	}
	
	// 简单的手机号验证，支持国际格式
	phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	if !phoneRegex.MatchString(phone) {
		return ErrInvalidPhone
	}
	
	return nil
}

// validateAvatar 验证头像URL
func validateAvatar(avatar string) error {
	if len(avatar) == 0 {
		return nil // 头像可以为空
	}
	if len(avatar) > 500 {
		return DomainError{Code: "AVATAR_URL_TOO_LONG", Message: "avatar URL cannot exceed 500 characters"}
	}
	
	// 简单的URL验证
	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	if !urlRegex.MatchString(avatar) {
		return DomainError{Code: "INVALID_AVATAR_URL", Message: "invalid avatar URL format"}
	}
	
	return nil
}

// isAdminPermission 检查是否为管理员权限
func isAdminPermission(permission string) bool {
	adminPermissions := map[string]bool{
		"user:manage":        true,
		"user:create":        true,
		"user:delete":        true,
		"user:promote":       true,
		"space:manage":       true,
		"space:delete":       true,
		"base:manage":        true,
		"base:delete":        true,
		"table:manage":       true,
		"system:config":      true,
		"system:monitor":     true,
		"permission:manage":  true,
	}

	return adminPermissions[permission]
}

// isBasicUserPermission 检查是否为基础用户权限
func isBasicUserPermission(permission string) bool {
	basicPermissions := map[string]bool{
		"user:read":         true,
		"user:update_self":  true,
		"space:read":        true,
		"space:create":      true,
		"base:read":         true,
		"base:create":       true,
		"table:read":        true,
		"table:create":      true,
		"record:read":       true,
		"record:create":     true,
		"record:update":     true,
		"record:delete":     true,
		"view:read":         true,
		"view:create":       true,
		"view:update":       true,
		"attachment:upload": true,
		"attachment:read":   true,
	}

	return basicPermissions[permission]
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
