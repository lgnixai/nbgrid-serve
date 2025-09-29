package user

import (
	"time"
)

// UserFactory 用户工厂接口
type UserFactory interface {
	CreateUser(req CreateUserRequest) (*UserAggregate, error)
	CreateSystemUser(name, email string) (*UserAggregate, error)
	CreateAdminUser(name, email, password string) (*UserAggregate, error)
	ReconstructUser(user *User) *UserAggregate
}

// UserFactoryImpl 用户工厂实现
type UserFactoryImpl struct {
	// 可以注入一些依赖，比如密码策略、ID生成器等
}

// NewUserFactory 创建用户工厂
func NewUserFactory() UserFactory {
	return &UserFactoryImpl{}
}

// CreateUser 创建普通用户
func (f *UserFactoryImpl) CreateUser(req CreateUserRequest) (*UserAggregate, error) {
	// 验证请求参数
	if err := f.validateCreateUserRequest(req); err != nil {
		return nil, err
	}

	// 创建用户实体
	now := time.Now()
	user := &User{
		ID:               generateUserID(),
		Name:             req.Name,
		Email:            req.Email,
		Phone:            req.Phone,
		Avatar:           req.Avatar,
		IsSystem:         false,
		IsAdmin:          false,
		IsTrialUsed:      false,
		CreatedTime:      now,
		LastModifiedTime: &now,
	}

	// 设置密码（如果提供）
	if req.Password != nil && *req.Password != "" {
		if err := user.SetPassword(*req.Password); err != nil {
			return nil, err
		}
	}

	// 创建聚合根
	aggregate := NewUserAggregate(user)

	// 添加用户创建事件
	event := NewUserCreatedEvent(user, "system")
	aggregate.addEvent(event)

	return aggregate, nil
}

// CreateSystemUser 创建系统用户
func (f *UserFactoryImpl) CreateSystemUser(name, email string) (*UserAggregate, error) {
	// 验证参数
	if err := validateUserName(name); err != nil {
		return nil, err
	}
	if err := validateEmail(email); err != nil {
		return nil, err
	}

	// 创建系统用户实体
	now := time.Now()
	user := &User{
		ID:               generateUserID(),
		Name:             name,
		Email:            email,
		IsSystem:         true, // 系统用户
		IsAdmin:          true, // 系统用户默认是管理员
		IsTrialUsed:      false,
		CreatedTime:      now,
		LastModifiedTime: &now,
	}

	// 创建聚合根
	aggregate := NewUserAggregate(user)

	// 添加用户创建事件
	event := NewUserCreatedEvent(user, "system")
	aggregate.addEvent(event)

	return aggregate, nil
}

// CreateAdminUser 创建管理员用户
func (f *UserFactoryImpl) CreateAdminUser(name, email, password string) (*UserAggregate, error) {
	// 验证参数
	if err := validateUserName(name); err != nil {
		return nil, err
	}
	if err := validateEmail(email); err != nil {
		return nil, err
	}
	if err := validatePassword(password); err != nil {
		return nil, err
	}

	// 创建管理员用户实体
	now := time.Now()
	user := &User{
		ID:               generateUserID(),
		Name:             name,
		Email:            email,
		IsSystem:         false,
		IsAdmin:          true, // 管理员用户
		IsTrialUsed:      false,
		CreatedTime:      now,
		LastModifiedTime: &now,
	}

	// 设置密码
	if err := user.SetPassword(password); err != nil {
		return nil, err
	}

	// 创建聚合根
	aggregate := NewUserAggregate(user)

	// 添加用户创建事件
	event := NewUserCreatedEvent(user, "system")
	aggregate.addEvent(event)

	return aggregate, nil
}

// ReconstructUser 从持久化数据重构用户聚合根
func (f *UserFactoryImpl) ReconstructUser(user *User) *UserAggregate {
	if user == nil {
		return nil
	}

	// 重构聚合根，不添加事件（因为是从持久化数据恢复）
	return NewUserAggregate(user)
}

// validateCreateUserRequest 验证创建用户请求
func (f *UserFactoryImpl) validateCreateUserRequest(req CreateUserRequest) error {
	// 验证用户名
	if err := validateUserName(req.Name); err != nil {
		return err
	}

	// 验证邮箱
	if err := validateEmail(req.Email); err != nil {
		return err
	}

	// 验证密码（如果提供）
	if req.Password != nil && *req.Password != "" {
		if err := validatePassword(*req.Password); err != nil {
			return err
		}
	}

	// 验证手机号（如果提供）
	if req.Phone != nil && *req.Phone != "" {
		if err := validatePhone(*req.Phone); err != nil {
			return err
		}
	}

	// 验证头像URL（如果提供）
	if req.Avatar != nil && *req.Avatar != "" {
		if err := validateAvatar(*req.Avatar); err != nil {
			return err
		}
	}

	return nil
}

// UserBuilder 用户构建器（用于复杂的用户创建场景）
type UserBuilder struct {
	user   *User
	errors []error
}

// NewUserBuilder 创建用户构建器
func NewUserBuilder() *UserBuilder {
	now := time.Now()
	return &UserBuilder{
		user: &User{
			ID:               generateUserID(),
			IsSystem:         false,
			IsAdmin:          false,
			IsTrialUsed:      false,
			CreatedTime:      now,
			LastModifiedTime: &now,
		},
		errors: make([]error, 0),
	}
}

// WithName 设置用户名
func (b *UserBuilder) WithName(name string) *UserBuilder {
	if err := validateUserName(name); err != nil {
		b.errors = append(b.errors, err)
		return b
	}
	b.user.Name = name
	return b
}

// WithEmail 设置邮箱
func (b *UserBuilder) WithEmail(email string) *UserBuilder {
	if err := validateEmail(email); err != nil {
		b.errors = append(b.errors, err)
		return b
	}
	b.user.Email = email
	return b
}

// WithPassword 设置密码
func (b *UserBuilder) WithPassword(password string) *UserBuilder {
	if err := b.user.SetPassword(password); err != nil {
		b.errors = append(b.errors, err)
		return b
	}
	return b
}

// WithPhone 设置手机号
func (b *UserBuilder) WithPhone(phone string) *UserBuilder {
	if err := validatePhone(phone); err != nil {
		b.errors = append(b.errors, err)
		return b
	}
	b.user.Phone = &phone
	return b
}

// WithAvatar 设置头像
func (b *UserBuilder) WithAvatar(avatar string) *UserBuilder {
	if err := validateAvatar(avatar); err != nil {
		b.errors = append(b.errors, err)
		return b
	}
	b.user.Avatar = &avatar
	return b
}

// AsAdmin 设置为管理员
func (b *UserBuilder) AsAdmin() *UserBuilder {
	b.user.IsAdmin = true
	return b
}

// AsSystem 设置为系统用户
func (b *UserBuilder) AsSystem() *UserBuilder {
	b.user.IsSystem = true
	b.user.IsAdmin = true // 系统用户默认是管理员
	return b
}

// WithTrialUsed 设置试用已使用
func (b *UserBuilder) WithTrialUsed() *UserBuilder {
	b.user.IsTrialUsed = true
	return b
}

// Build 构建用户聚合根
func (b *UserBuilder) Build() (*UserAggregate, error) {
	// 检查是否有错误
	if len(b.errors) > 0 {
		return nil, b.errors[0] // 返回第一个错误
	}

	// 验证必填字段
	if b.user.Name == "" {
		return nil, DomainError{Code: "EMPTY_NAME", Message: "user name is required"}
	}
	if b.user.Email == "" {
		return nil, DomainError{Code: "EMPTY_EMAIL", Message: "user email is required"}
	}

	// 创建聚合根
	aggregate := NewUserAggregate(b.user)

	// 添加用户创建事件
	event := NewUserCreatedEvent(b.user, "system")
	aggregate.addEvent(event)

	return aggregate, nil
}

// GetErrors 获取构建过程中的所有错误
func (b *UserBuilder) GetErrors() []error {
	return b.errors
}

// HasErrors 检查是否有错误
func (b *UserBuilder) HasErrors() bool {
	return len(b.errors) > 0
}
