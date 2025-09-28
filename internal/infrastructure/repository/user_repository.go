package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	userDomain "teable-go-backend/internal/domain/user"
	"teable-go-backend/internal/infrastructure/database/models"
	"teable-go-backend/pkg/errors"
)

// UserRepository 用户仓储实现
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓储
func NewUserRepository(db *gorm.DB) userDomain.Repository {
	return &UserRepository{db: db}
}

// Create 创建用户
func (r *UserRepository) Create(ctx context.Context, user *userDomain.User) error {
	model := r.domainToModel(user)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return r.handleDBError(err)
	}

	return nil
}

// GetByID 通过ID获取用户
func (r *UserRepository) GetByID(ctx context.Context, id string) (*userDomain.User, error) {
	var model models.User

	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, r.handleDBError(err)
	}

	return r.modelToDomain(&model), nil
}

// GetByEmail 通过邮箱获取用户
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*userDomain.User, error) {
	var model models.User

	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, r.handleDBError(err)
	}

	return r.modelToDomain(&model), nil
}

// GetByPhone 通过手机号获取用户
func (r *UserRepository) GetByPhone(ctx context.Context, phone string) (*userDomain.User, error) {
	var model models.User

	if err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, r.handleDBError(err)
	}

	return r.modelToDomain(&model), nil
}

// Update 更新用户
func (r *UserRepository) Update(ctx context.Context, user *userDomain.User) error {
	model := r.domainToModel(user)

	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return r.handleDBError(err)
	}

	return nil
}

// Delete 删除用户
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&models.User{}, "id = ?", id).Error; err != nil {
		return r.handleDBError(err)
	}

	return nil
}

// List 列出用户
func (r *UserRepository) List(ctx context.Context, filter userDomain.ListFilter) ([]*userDomain.User, error) {
	var modelUsers []models.User

	query := r.db.WithContext(ctx).Model(&models.User{})

	// 应用过滤条件
	query = r.applyListFilter(query, filter)

	// 应用排序
	if filter.OrderBy != "" && filter.Order != "" {
		query = query.Order(fmt.Sprintf("%s %s", filter.OrderBy, filter.Order))
	}

	// 应用分页
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	if err := query.Find(&modelUsers).Error; err != nil {
		return nil, r.handleDBError(err)
	}

	// 转换为领域对象
	users := make([]*userDomain.User, len(modelUsers))
	for i, model := range modelUsers {
		users[i] = r.modelToDomain(&model)
	}

	return users, nil
}

// Count 统计用户数量
func (r *UserRepository) Count(ctx context.Context, filter userDomain.CountFilter) (int64, error) {
	var count int64

	query := r.db.WithContext(ctx).Model(&models.User{})

	// 应用过滤条件
	query = r.applyCountFilter(query, filter)

	if err := query.Count(&count).Error; err != nil {
		return 0, r.handleDBError(err)
	}

	return count, nil
}

// Exists 检查用户是否存在
func (r *UserRepository) Exists(ctx context.Context, filter userDomain.ExistsFilter) (bool, error) {
	var count int64

	query := r.db.WithContext(ctx).Model(&models.User{})

	// 应用过滤条件
	if filter.ID != nil {
		query = query.Where("id = ?", *filter.ID)
	}
	if filter.Email != nil {
		query = query.Where("email = ?", *filter.Email)
	}
	if filter.Phone != nil {
		query = query.Where("phone = ?", *filter.Phone)
	}

	if err := query.Count(&count).Error; err != nil {
		return false, r.handleDBError(err)
	}

	return count > 0, nil
}

// BatchCreate 批量创建用户
func (r *UserRepository) BatchCreate(ctx context.Context, users []*userDomain.User) error {
	if len(users) == 0 {
		return nil
	}

	models := make([]models.User, len(users))
	for i, user := range users {
		models[i] = *r.domainToModel(user)
	}

	if err := r.db.WithContext(ctx).CreateInBatches(models, 100).Error; err != nil {
		return r.handleDBError(err)
	}

	return nil
}

// BatchUpdate 批量更新用户
func (r *UserRepository) BatchUpdate(ctx context.Context, users []*userDomain.User) error {
	if len(users) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, user := range users {
			model := r.domainToModel(user)
			if err := tx.Save(model).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// BatchDelete 批量删除用户
func (r *UserRepository) BatchDelete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).Delete(&models.User{}, "id IN ?", ids).Error; err != nil {
		return r.handleDBError(err)
	}

	return nil
}

// CreateAccount 创建账户
func (r *UserRepository) CreateAccount(ctx context.Context, account *userDomain.Account) error {
	model := r.accountDomainToModel(account)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return r.handleDBError(err)
	}

	return nil
}

// GetAccountsByUserID 获取用户的所有账户
func (r *UserRepository) GetAccountsByUserID(ctx context.Context, userID string) ([]*userDomain.Account, error) {
	var models []models.Account

	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&models).Error; err != nil {
		return nil, r.handleDBError(err)
	}

	accounts := make([]*userDomain.Account, len(models))
	for i, model := range models {
		accounts[i] = r.accountModelToDomain(&model)
	}

	return accounts, nil
}

// GetAccountByProvider 通过提供商获取账户
func (r *UserRepository) GetAccountByProvider(ctx context.Context, provider, providerID string) (*userDomain.Account, error) {
	var model models.Account

	if err := r.db.WithContext(ctx).Where("provider = ? AND provider_id = ?", provider, providerID).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, r.handleDBError(err)
	}

	return r.accountModelToDomain(&model), nil
}

// DeleteAccount 删除账户
func (r *UserRepository) DeleteAccount(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&models.Account{}, "id = ?", id).Error; err != nil {
		return r.handleDBError(err)
	}

	return nil
}

// 辅助方法

// domainToModel 领域对象转数据模型
func (r *UserRepository) domainToModel(user *userDomain.User) *models.User {
	model := &models.User{
		ID:                   user.ID,
		Name:                 user.Name,
		Email:                user.Email,
		Password:             user.Password,
		Salt:                 user.Salt,
		Phone:                user.Phone,
		Avatar:               user.Avatar,
		IsSystem:             &user.IsSystem,
		IsAdmin:              &user.IsAdmin,
		IsTrialUsed:          &user.IsTrialUsed,
		NotifyMeta:           user.NotifyMeta,
		LastSignTime:         user.LastSignTime,
		DeactivatedTime:      user.DeactivatedTime,
		PermanentDeletedTime: user.PermanentDeletedTime,
		RefMeta:              user.RefMeta,
	}

	// 处理软删除字段
	if user.DeletedTime != nil {
		model.DeletedTime = gorm.DeletedAt{
			Time:  *user.DeletedTime,
			Valid: true,
		}
	}

	// 只有在更新时才设置CreatedTime和LastModifiedTime
	if !user.CreatedTime.IsZero() {
		model.CreatedTime = user.CreatedTime
	}
	if user.LastModifiedTime != nil {
		model.LastModifiedTime = user.LastModifiedTime
	}

	return model
}

// modelToDomain 数据模型转领域对象
func (r *UserRepository) modelToDomain(model *models.User) *userDomain.User {
	isSystem := false
	isAdmin := false
	isTrialUsed := false

	if model.IsSystem != nil {
		isSystem = *model.IsSystem
	}
	if model.IsAdmin != nil {
		isAdmin = *model.IsAdmin
	}
	if model.IsTrialUsed != nil {
		isTrialUsed = *model.IsTrialUsed
	}

	var deletedTime *time.Time
	if model.DeletedTime.Valid {
		deletedTime = &model.DeletedTime.Time
	}

	return &userDomain.User{
		ID:                   model.ID,
		Name:                 model.Name,
		Email:                model.Email,
		Password:             model.Password,
		Salt:                 model.Salt,
		Phone:                model.Phone,
		Avatar:               model.Avatar,
		IsSystem:             isSystem,
		IsAdmin:              isAdmin,
		IsTrialUsed:          isTrialUsed,
		NotifyMeta:           model.NotifyMeta,
		LastSignTime:         model.LastSignTime,
		DeactivatedTime:      model.DeactivatedTime,
		CreatedTime:          model.CreatedTime,
		DeletedTime:          deletedTime,
		LastModifiedTime:     model.LastModifiedTime,
		PermanentDeletedTime: model.PermanentDeletedTime,
		RefMeta:              model.RefMeta,
	}
}

// accountDomainToModel 账户领域对象转数据模型
func (r *UserRepository) accountDomainToModel(account *userDomain.Account) *models.Account {
	return &models.Account{
		ID:          account.ID,
		UserID:      account.UserID,
		Type:        account.Type,
		Provider:    account.Provider,
		ProviderID:  account.ProviderID,
		CreatedTime: account.CreatedTime,
	}
}

// accountModelToDomain 账户数据模型转领域对象
func (r *UserRepository) accountModelToDomain(model *models.Account) *userDomain.Account {
	return &userDomain.Account{
		ID:          model.ID,
		UserID:      model.UserID,
		Type:        model.Type,
		Provider:    model.Provider,
		ProviderID:  model.ProviderID,
		CreatedTime: model.CreatedTime,
	}
}

// applyListFilter 应用列表过滤条件
func (r *UserRepository) applyListFilter(query *gorm.DB, filter userDomain.ListFilter) *gorm.DB {
	if filter.Name != nil {
		query = query.Where("name LIKE ?", "%"+*filter.Name+"%")
	}
	if filter.Email != nil {
		query = query.Where("email LIKE ?", "%"+*filter.Email+"%")
	}
	if filter.IsActive != nil {
		if *filter.IsActive {
			query = query.Where("deactivated_time IS NULL AND deleted_time IS NULL")
		} else {
			query = query.Where("deactivated_time IS NOT NULL OR deleted_time IS NOT NULL")
		}
	}
	if filter.IsAdmin != nil {
		query = query.Where("is_admin = ?", *filter.IsAdmin)
	}
	if filter.IsSystem != nil {
		query = query.Where("is_system = ?", *filter.IsSystem)
	}
	if filter.CreatedAfter != nil {
		query = query.Where("created_time >= ?", *filter.CreatedAfter)
	}
	if filter.CreatedBefore != nil {
		query = query.Where("created_time <= ?", *filter.CreatedBefore)
	}
	if filter.ModifiedAfter != nil {
		query = query.Where("last_modified_time >= ?", *filter.ModifiedAfter)
	}
	if filter.ModifiedBefore != nil {
		query = query.Where("last_modified_time <= ?", *filter.ModifiedBefore)
	}
	if filter.Search != "" {
		searchTerm := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ?", searchTerm, searchTerm)
	}

	return query
}

// applyCountFilter 应用计数过滤条件
func (r *UserRepository) applyCountFilter(query *gorm.DB, filter userDomain.CountFilter) *gorm.DB {
	if filter.Name != nil {
		query = query.Where("name LIKE ?", "%"+*filter.Name+"%")
	}
	if filter.Email != nil {
		query = query.Where("email LIKE ?", "%"+*filter.Email+"%")
	}
	if filter.IsActive != nil {
		if *filter.IsActive {
			query = query.Where("deactivated_time IS NULL AND deleted_time IS NULL")
		} else {
			query = query.Where("deactivated_time IS NOT NULL OR deleted_time IS NOT NULL")
		}
	}
	if filter.IsAdmin != nil {
		query = query.Where("is_admin = ?", *filter.IsAdmin)
	}
	if filter.IsSystem != nil {
		query = query.Where("is_system = ?", *filter.IsSystem)
	}
	if filter.CreatedAfter != nil {
		query = query.Where("created_time >= ?", *filter.CreatedAfter)
	}
	if filter.CreatedBefore != nil {
		query = query.Where("created_time <= ?", *filter.CreatedBefore)
	}
	if filter.ModifiedAfter != nil {
		query = query.Where("last_modified_time >= ?", *filter.ModifiedAfter)
	}
	if filter.ModifiedBefore != nil {
		query = query.Where("last_modified_time <= ?", *filter.ModifiedBefore)
	}
	if filter.Search != "" {
		searchTerm := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ?", searchTerm, searchTerm)
	}

	return query
}

// handleDBError 处理数据库错误
func (r *UserRepository) handleDBError(err error) error {
	// TODO: 根据具体的数据库错误类型返回对应的业务错误
	if strings.Contains(err.Error(), "duplicate key") {
		if strings.Contains(err.Error(), "email") {
			return errors.ErrEmailExists
		}
		if strings.Contains(err.Error(), "phone") {
			return errors.ErrPhoneExists
		}
	}

	return errors.ErrDatabaseOperation.WithDetails(err.Error())
}
