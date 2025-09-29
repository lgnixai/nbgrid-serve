package user

import (
	"context"
	"errors"
	"time"

	pkgErrors "teable-go-backend/pkg/errors"
)

// Service 用户领域服务接口
type Service interface {
	// 用户管理
	CreateUser(ctx context.Context, req CreateUserRequest) (*User, error)
	GetUser(ctx context.Context, id string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, id string, req UpdateUserRequest) (*User, error)
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, filter ListFilter) (*PaginatedResult, error)

	// 认证相关
	Authenticate(ctx context.Context, email, password string) (*User, error)
	ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error
	ResetPassword(ctx context.Context, email string) error

	// 用户状态管理
	ActivateUser(ctx context.Context, userID string) error
	DeactivateUser(ctx context.Context, userID string) error

	// 权限管理
	PromoteToAdmin(ctx context.Context, userID string) error
	DemoteFromAdmin(ctx context.Context, userID string) error

	// 第三方账户
	LinkAccount(ctx context.Context, userID string, req LinkAccountRequest) error
	UnlinkAccount(ctx context.Context, userID, accountID string) error
	GetUserByProvider(ctx context.Context, provider, providerID string) (*User, error)

	// 高级用户管理功能
	BulkUpdateUsers(ctx context.Context, updates []BulkUpdateRequest) error
	BulkDeleteUsers(ctx context.Context, userIDs []string) error
	ExportUsers(ctx context.Context, filter ListFilter) ([]*User, error)
	ImportUsers(ctx context.Context, users []CreateUserRequest) ([]*User, error)

	// 用户统计和分析
	GetUserStats(ctx context.Context) (*UserStats, error)
	GetUserActivity(ctx context.Context, userID string, days int) (*UserActivity, error)

	// 用户偏好设置
	UpdateUserPreferences(ctx context.Context, userID string, prefs UserPreferences) error
	GetUserPreferences(ctx context.Context, userID string) (*UserPreferences, error)
}

// ServiceImpl 用户领域服务实现
type ServiceImpl struct {
	repo Repository
}

// NewService 创建用户服务
func NewService(repo Repository) Service {
	return &ServiceImpl{
		repo: repo,
	}
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Name     string  `json:"name" validate:"required,min=1,max=100"`
	Email    string  `json:"email" validate:"required,email,max=255"`
	Password *string `json:"password,omitempty" validate:"omitempty,min=8,max=128"`
	Phone    *string `json:"phone,omitempty" validate:"omitempty,max=50"`
	Avatar   *string `json:"avatar,omitempty" validate:"omitempty,url,max=500"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Name   *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Phone  *string `json:"phone,omitempty" validate:"omitempty,max=50"`
	Avatar *string `json:"avatar,omitempty" validate:"omitempty,url,max=500"`
}

// LinkAccountRequest 关联账户请求
type LinkAccountRequest struct {
	Type       AccountType `json:"type" validate:"required"`
	Provider   Provider    `json:"provider" validate:"required"`
	ProviderID string      `json:"provider_id" validate:"required"`
}

// BulkUpdateRequest 批量更新请求
type BulkUpdateRequest struct {
	UserID string                 `json:"user_id" validate:"required"`
	Fields map[string]interface{} `json:"fields" validate:"required"`
}

// UserStats 用户统计信息
type UserStats struct {
	TotalUsers        int64 `json:"total_users"`
	ActiveUsers       int64 `json:"active_users"`
	InactiveUsers     int64 `json:"inactive_users"`
	AdminUsers        int64 `json:"admin_users"`
	SystemUsers       int64 `json:"system_users"`
	NewUsersToday     int64 `json:"new_users_today"`
	NewUsersThisWeek  int64 `json:"new_users_this_week"`
	NewUsersThisMonth int64 `json:"new_users_this_month"`
}

// UserActivity 用户活动信息
type UserActivity struct {
	UserID           string     `json:"user_id"`
	LastLoginTime    *time.Time `json:"last_login_time"`
	LoginCount       int64      `json:"login_count"`
	SpacesCreated    int64      `json:"spaces_created"`
	BasesCreated     int64      `json:"bases_created"`
	TablesCreated    int64      `json:"tables_created"`
	RecordsCreated   int64      `json:"records_created"`
	LastActivityTime *time.Time `json:"last_activity_time"`
}

// 注意：UserPreferences 和相关结构已移动到 value_objects.go

// CreateUser 创建用户
func (s *ServiceImpl) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
	// 检查邮箱是否已存在
	exists, err := s.repo.Exists(ctx, ExistsFilter{Email: &req.Email})
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, pkgErrors.ErrEmailExists
	}

	// 检查手机号是否已存在
	if req.Phone != nil {
		exists, err := s.repo.Exists(ctx, ExistsFilter{Phone: req.Phone})
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, pkgErrors.ErrPhoneExists
		}
	}

	// 创建用户
	var user *User
	if req.Password != nil {
		user, err = NewUserWithPassword(req.Name, req.Email, *req.Password)
	} else {
		user, err = NewUser(req.Name, req.Email)
	}
	if err != nil {
		return nil, convertDomainError(err)
	}

	// 设置其他属性
	if req.Phone != nil {
		user.Phone = req.Phone
	}
	if req.Avatar != nil {
		user.Avatar = req.Avatar
	}

	// 保存到数据库
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// GetUser 获取用户
func (s *ServiceImpl) GetUser(ctx context.Context, id string) (*User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, pkgErrors.ErrUserNotFound
	}
	return user, nil
}

// GetUserByEmail 通过邮箱获取用户
func (s *ServiceImpl) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, pkgErrors.ErrUserNotFound
	}
	return user, nil
}

// UpdateUser 更新用户
func (s *ServiceImpl) UpdateUser(ctx context.Context, id string, req UpdateUserRequest) (*User, error) {
	user, err := s.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}

	// 验证用户是否可以被更新
	if err := user.ValidateForUpdate(); err != nil {
		return nil, convertDomainError(err)
	}

	// 更新用户信息
	if err := user.UpdateProfile(req.Name, req.Phone, req.Avatar); err != nil {
		return nil, convertDomainError(err)
	}

	// 保存更新
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser 删除用户
func (s *ServiceImpl) DeleteUser(ctx context.Context, id string) error {
	user, err := s.GetUser(ctx, id)
	if err != nil {
		return err
	}

	// 验证用户是否可以被删除
	if err := user.ValidateForDeletion(); err != nil {
		return convertDomainError(err)
	}

	// 软删除
	user.SoftDelete()

	return s.repo.Update(ctx, user)
}

// ListUsers 列出用户
func (s *ServiceImpl) ListUsers(ctx context.Context, filter ListFilter) (*PaginatedResult, error) {
	users, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	// 计算总数
	countFilter := CountFilter{
		Name:           filter.Name,
		Email:          filter.Email,
		IsActive:       filter.IsActive,
		IsAdmin:        filter.IsAdmin,
		IsSystem:       filter.IsSystem,
		CreatedAfter:   filter.CreatedAfter,
		CreatedBefore:  filter.CreatedBefore,
		ModifiedAfter:  filter.ModifiedAfter,
		ModifiedBefore: filter.ModifiedBefore,
		Search:         filter.Search,
	}

	total, err := s.repo.Count(ctx, countFilter)
	if err != nil {
		return nil, err
	}

	// 计算分页信息
	page := filter.Offset/filter.Limit + 1
	totalPages := int(total) / filter.Limit
	if int(total)%filter.Limit > 0 {
		totalPages++
	}

	return &PaginatedResult{
		Users:      users,
		Total:      total,
		Page:       page,
		PageSize:   filter.Limit,
		TotalPages: totalPages,
	}, nil
}

// Authenticate 用户认证
func (s *ServiceImpl) Authenticate(ctx context.Context, email, password string) (*User, error) {
	user, err := s.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	// 检查用户状态
	if !user.IsActive() {
		if user.GetStatus() == UserStatusDeactivated {
			return nil, pkgErrors.ErrUserDeactivated
		}
		return nil, pkgErrors.ErrUserDeleted
	}

	// 验证密码
	if err := user.CheckPassword(password); err != nil {
		return nil, pkgErrors.ErrInvalidCredentials
	}

	// 记录登录时间
	user.RecordSignIn()
	if err := s.repo.Update(ctx, user); err != nil {
		// 登录时间更新失败不影响认证结果
		// TODO: 记录日志
	}

	return user, nil
}

// ChangePassword 修改密码
func (s *ServiceImpl) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	user, err := s.GetUser(ctx, userID)
	if err != nil {
		return err
	}

	// 检查用户状态
	if !user.IsActive() {
		return pkgErrors.ErrUserDeactivated
	}

	// 验证旧密码
	if err := user.CheckPassword(oldPassword); err != nil {
		return pkgErrors.ErrInvalidCredentials
	}

	// 设置新密码
	if err := user.SetPassword(newPassword); err != nil {
		return convertDomainError(err)
	}

	return s.repo.Update(ctx, user)
}

// ResetPassword 重置密码
func (s *ServiceImpl) ResetPassword(ctx context.Context, email string) error {
	user, err := s.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}

	// 检查用户状态
	if !user.IsActive() {
		return pkgErrors.ErrUserDeactivated
	}

	// TODO: 实现密码重置逻辑
	// 1. 生成重置令牌
	// 2. 发送重置邮件
	// 3. 存储令牌到缓存

	return errors.New("password reset not implemented yet")
}

// ActivateUser 激活用户
func (s *ServiceImpl) ActivateUser(ctx context.Context, userID string) error {
	user, err := s.GetUser(ctx, userID)
	if err != nil {
		return err
	}

	user.Activate()
	return s.repo.Update(ctx, user)
}

// DeactivateUser 停用用户
func (s *ServiceImpl) DeactivateUser(ctx context.Context, userID string) error {
	user, err := s.GetUser(ctx, userID)
	if err != nil {
		return err
	}

	user.Deactivate()
	return s.repo.Update(ctx, user)
}

// PromoteToAdmin 提升为管理员
func (s *ServiceImpl) PromoteToAdmin(ctx context.Context, userID string) error {
	user, err := s.GetUser(ctx, userID)
	if err != nil {
		return err
	}

	user.PromoteToAdmin()
	return s.repo.Update(ctx, user)
}

// DemoteFromAdmin 取消管理员
func (s *ServiceImpl) DemoteFromAdmin(ctx context.Context, userID string) error {
	user, err := s.GetUser(ctx, userID)
	if err != nil {
		return err
	}

	user.DemoteFromAdmin()
	return s.repo.Update(ctx, user)
}

// LinkAccount 关联第三方账户
func (s *ServiceImpl) LinkAccount(ctx context.Context, userID string, req LinkAccountRequest) error {
	user, err := s.GetUser(ctx, userID)
	if err != nil {
		return err
	}

	// 检查提供商账户是否已被其他用户使用
	existingAccount, err := s.repo.GetAccountByProvider(ctx, string(req.Provider), req.ProviderID)
	if err != nil {
		return err
	}
	if existingAccount != nil {
		return errors.New("provider account already linked to another user")
	}

	// 创建账户关联
	account := user.AddAccount(req.Type, req.Provider, req.ProviderID)
	return s.repo.CreateAccount(ctx, account)
}

// UnlinkAccount 取消关联第三方账户
func (s *ServiceImpl) UnlinkAccount(ctx context.Context, userID, accountID string) error {
	// TODO: 验证账户属于该用户
	return s.repo.DeleteAccount(ctx, accountID)
}

// GetUserByProvider 通过第三方提供商获取用户
func (s *ServiceImpl) GetUserByProvider(ctx context.Context, provider, providerID string) (*User, error) {
	account, err := s.repo.GetAccountByProvider(ctx, provider, providerID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, pkgErrors.ErrUserNotFound
	}

	return s.GetUser(ctx, account.UserID)
}

// BulkUpdateUsers 批量更新用户
func (s *ServiceImpl) BulkUpdateUsers(ctx context.Context, updates []BulkUpdateRequest) error {
	if len(updates) == 0 {
		return nil
	}

	// 获取所有需要更新的用户
	userIDs := make([]string, len(updates))
	for i, update := range updates {
		userIDs[i] = update.UserID
	}

	// 批量获取用户
	users := make([]*User, 0, len(updates))
	for _, userID := range userIDs {
		user, err := s.GetUser(ctx, userID)
		if err != nil {
			return err
		}
		users = append(users, user)
	}

	// 批量更新用户
	for i, user := range users {
		update := updates[i]
		// 应用更新字段
		if name, ok := update.Fields["name"].(string); ok {
			user.Name = name
		}
		if phone, ok := update.Fields["phone"].(string); ok {
			user.Phone = &phone
		}
		if avatar, ok := update.Fields["avatar"].(string); ok {
			user.Avatar = &avatar
		}
		if isAdmin, ok := update.Fields["is_admin"].(bool); ok {
			user.IsAdmin = isAdmin
		}
		user.updateModifiedTime()
	}

	// 批量保存到数据库
	return s.repo.BatchUpdate(ctx, users)
}

// BulkDeleteUsers 批量删除用户
func (s *ServiceImpl) BulkDeleteUsers(ctx context.Context, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}

	// 验证所有用户都存在
	for _, userID := range userIDs {
		_, err := s.GetUser(ctx, userID)
		if err != nil {
			return err
		}
	}

	// 批量软删除
	return s.repo.BatchDelete(ctx, userIDs)
}

// ExportUsers 导出用户数据
func (s *ServiceImpl) ExportUsers(ctx context.Context, filter ListFilter) ([]*User, error) {
	// 设置较大的限制以获取所有用户
	filter.Limit = 10000
	return s.repo.List(ctx, filter)
}

// ImportUsers 导入用户数据
func (s *ServiceImpl) ImportUsers(ctx context.Context, userReqs []CreateUserRequest) ([]*User, error) {
	if len(userReqs) == 0 {
		return nil, nil
	}

	users := make([]*User, 0, len(userReqs))

	// 创建用户
	for _, req := range userReqs {
		user, err := s.CreateUser(ctx, req)
		if err != nil {
			// 如果创建失败，记录错误但继续处理其他用户
			continue
		}
		users = append(users, user)
	}

	// 批量保存到数据库
	if err := s.repo.BatchCreate(ctx, users); err != nil {
		return nil, err
	}

	return users, nil
}

// GetUserStats 获取用户统计信息
func (s *ServiceImpl) GetUserStats(ctx context.Context) (*UserStats, error) {
	now := time.Now()
	today := now.Truncate(24 * time.Hour)
	weekAgo := now.AddDate(0, 0, -7)
	monthAgo := now.AddDate(0, -1, 0)

	// 获取总用户数
	totalUsers, err := s.repo.Count(ctx, CountFilter{})
	if err != nil {
		return nil, err
	}

	// 获取活跃用户数
	activeUsers, err := s.repo.Count(ctx, CountFilter{IsActive: boolPtr(true)})
	if err != nil {
		return nil, err
	}

	// 获取非活跃用户数
	inactiveUsers, err := s.repo.Count(ctx, CountFilter{IsActive: boolPtr(false)})
	if err != nil {
		return nil, err
	}

	// 获取管理员用户数
	adminUsers, err := s.repo.Count(ctx, CountFilter{IsAdmin: boolPtr(true)})
	if err != nil {
		return nil, err
	}

	// 获取系统用户数
	systemUsers, err := s.repo.Count(ctx, CountFilter{IsSystem: boolPtr(true)})
	if err != nil {
		return nil, err
	}

	// 获取今日新用户数
	todayStr := today.Format("2006-01-02")
	newUsersToday, err := s.repo.Count(ctx, CountFilter{CreatedAfter: &todayStr})
	if err != nil {
		return nil, err
	}

	// 获取本周新用户数
	weekAgoStr := weekAgo.Format("2006-01-02")
	newUsersThisWeek, err := s.repo.Count(ctx, CountFilter{CreatedAfter: &weekAgoStr})
	if err != nil {
		return nil, err
	}

	// 获取本月新用户数
	monthAgoStr := monthAgo.Format("2006-01-02")
	newUsersThisMonth, err := s.repo.Count(ctx, CountFilter{CreatedAfter: &monthAgoStr})
	if err != nil {
		return nil, err
	}

	return &UserStats{
		TotalUsers:        totalUsers,
		ActiveUsers:       activeUsers,
		InactiveUsers:     inactiveUsers,
		AdminUsers:        adminUsers,
		SystemUsers:       systemUsers,
		NewUsersToday:     newUsersToday,
		NewUsersThisWeek:  newUsersThisWeek,
		NewUsersThisMonth: newUsersThisMonth,
	}, nil
}

// GetUserActivity 获取用户活动信息
func (s *ServiceImpl) GetUserActivity(ctx context.Context, userID string, days int) (*UserActivity, error) {
	user, err := s.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// TODO: 实现从其他服务获取用户活动数据
	// 这里需要与空间、基础表、表格等服务集成
	// 暂时返回基础信息
	return &UserActivity{
		UserID:           user.ID,
		LastLoginTime:    user.LastSignTime,
		LoginCount:       0, // TODO: 从登录日志中获取
		SpacesCreated:    0, // TODO: 从空间服务获取
		BasesCreated:     0, // TODO: 从基础表服务获取
		TablesCreated:    0, // TODO: 从表格服务获取
		RecordsCreated:   0, // TODO: 从记录服务获取
		LastActivityTime: user.LastSignTime,
	}, nil
}

// UpdateUserPreferences 更新用户偏好设置
func (s *ServiceImpl) UpdateUserPreferences(ctx context.Context, userID string, prefs UserPreferences) error {
	user, err := s.GetUser(ctx, userID)
	if err != nil {
		return err
	}

	// 验证偏好设置
	if err := prefs.Validate(); err != nil {
		return convertDomainError(err)
	}

	// 将偏好设置序列化为JSON并存储到NotifyMeta字段
	prefsJSON, err := prefs.ToJSON()
	if err != nil {
		return convertDomainError(err)
	}

	user.NotifyMeta = &prefsJSON
	user.updateModifiedTime()

	return s.repo.Update(ctx, user)
}

// GetUserPreferences 获取用户偏好设置
func (s *ServiceImpl) GetUserPreferences(ctx context.Context, userID string) (*UserPreferences, error) {
	user, err := s.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 从NotifyMeta字段反序列化偏好设置
	var prefsJSON string
	if user.NotifyMeta != nil {
		prefsJSON = *user.NotifyMeta
	}

	prefs, err := UserPreferencesFromJSON(prefsJSON)
	if err != nil {
		// 如果反序列化失败，返回默认偏好设置
		return NewDefaultUserPreferences(), nil
	}

	return prefs, nil
}

// 辅助函数
func boolPtr(b bool) *bool {
	return &b
}

func stringPtr(s string) *string {
	return &s
}

// convertDomainError 将领域错误转换为应用错误
func convertDomainError(err error) error {
	if domainErr, ok := err.(DomainError); ok {
		switch domainErr.Code {
		case "INVALID_EMAIL", "EMPTY_EMAIL", "EMAIL_TOO_LONG":
			return pkgErrors.ErrBadRequest.WithDetails(domainErr.Message)
		case "WEAK_PASSWORD", "PASSWORD_TOO_SHORT", "PASSWORD_TOO_LONG", "PASSWORD_TOO_WEAK":
			return pkgErrors.ErrInvalidPassword.WithDetails(domainErr.Message)
		case "EMAIL_EXISTS":
			return pkgErrors.ErrEmailExists
		case "PHONE_EXISTS":
			return pkgErrors.ErrPhoneExists
		case "USER_NOT_FOUND":
			return pkgErrors.ErrUserNotFound
		case "USER_DEACTIVATED":
			return pkgErrors.ErrUserDeactivated
		case "USER_DELETED":
			return pkgErrors.ErrUserDeleted
		case "INVALID_PASSWORD":
			return pkgErrors.ErrInvalidCredentials
		case "SYSTEM_USER_READONLY", "SYSTEM_USER_UNDELETABLE":
			return pkgErrors.ErrOperationNotAllowed.WithDetails(domainErr.Message)
		default:
			return pkgErrors.ErrBadRequest.WithDetails(domainErr.Message)
		}
	}
	return err
}
