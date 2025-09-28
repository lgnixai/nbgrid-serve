package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"teable-go-backend/internal/domain/notification"
	"teable-go-backend/internal/infrastructure/database/models"
	"teable-go-backend/pkg/errors"
)

// NotificationRepository 通知仓储实现
type NotificationRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewNotificationRepository 创建新的NotificationRepository
func NewNotificationRepository(db *gorm.DB, logger *zap.Logger) *NotificationRepository {
	return &NotificationRepository{
		db:     db,
		logger: logger,
	}
}

// CreateNotification 创建通知
func (r *NotificationRepository) CreateNotification(ctx context.Context, n *notification.Notification) error {
	model := r.domainToModel(n)
	
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		r.logger.Error("Failed to create notification", 
			zap.String("id", n.ID),
			zap.Error(err))
		return err
	}

	return nil
}

// GetNotification 获取通知
func (r *NotificationRepository) GetNotification(ctx context.Context, id string) (*notification.Notification, error) {
	var model models.Notification
	
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		r.logger.Error("Failed to get notification", 
			zap.String("id", id),
			zap.Error(err))
		return nil, err
	}

	return r.modelToDomain(&model), nil
}

// UpdateNotification 更新通知
func (r *NotificationRepository) UpdateNotification(ctx context.Context, n *notification.Notification) error {
	model := r.domainToModel(n)
	
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		r.logger.Error("Failed to update notification", 
			zap.String("id", n.ID),
			zap.Error(err))
		return err
	}

	return nil
}

// DeleteNotification 删除通知
func (r *NotificationRepository) DeleteNotification(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Notification{}).Error; err != nil {
		r.logger.Error("Failed to delete notification", 
			zap.String("id", id),
			zap.Error(err))
		return err
	}

	return nil
}

// ListNotifications 列出通知
func (r *NotificationRepository) ListNotifications(ctx context.Context, req *notification.ListNotificationsRequest) (*notification.ListNotificationsResponse, error) {
	query := r.db.WithContext(ctx).Model(&models.Notification{}).Where("user_id = ?", req.UserID)

	// 添加过滤条件
	if req.Type != nil {
		query = query.Where("type = ?", string(*req.Type))
	}
	if req.Status != nil {
		query = query.Where("status = ?", string(*req.Status))
	}
	if req.Priority != nil {
		query = query.Where("priority = ?", string(*req.Priority))
	}
	if req.SourceID != "" {
		query = query.Where("source_id = ?", req.SourceID)
	}
	if req.SourceType != "" {
		query = query.Where("source_type = ?", req.SourceType)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("Failed to count notifications", zap.Error(err))
		return nil, err
	}

	// 排序
	orderBy := fmt.Sprintf("%s %s", req.SortBy, req.SortOrder)
	query = query.Order(orderBy)

	// 分页
	offset := (req.Page - 1) * req.PageSize
	query = query.Offset(offset).Limit(req.PageSize)

	// 查询数据
	var models []models.Notification
	if err := query.Find(&models).Error; err != nil {
		r.logger.Error("Failed to list notifications", zap.Error(err))
		return nil, err
	}

	// 转换为领域对象
	notifications := make([]*notification.Notification, len(models))
	for i, model := range models {
		notifications[i] = r.modelToDomain(&model)
	}

	// 计算总页数
	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	return &notification.ListNotificationsResponse{
		Notifications: notifications,
		Total:         total,
		Page:          req.Page,
		PageSize:      req.PageSize,
		TotalPages:    totalPages,
	}, nil
}

// MarkNotificationsRead 标记通知为已读
func (r *NotificationRepository) MarkNotificationsRead(ctx context.Context, notificationIDs []string) error {
	now := time.Now()
	
	if err := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("id IN ?", notificationIDs).
		Updates(map[string]interface{}{
			"status":      string(notification.NotificationStatusRead),
			"read_at":     &now,
			"updated_time": now,
		}).Error; err != nil {
		r.logger.Error("Failed to mark notifications as read", 
			zap.Strings("notification_ids", notificationIDs),
			zap.Error(err))
		return err
	}

	return nil
}

// MarkAllNotificationsRead 标记所有通知为已读
func (r *NotificationRepository) MarkAllNotificationsRead(ctx context.Context, userID string) error {
	now := time.Now()
	
	if err := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("user_id = ? AND status = ?", userID, string(notification.NotificationStatusUnread)).
		Updates(map[string]interface{}{
			"status":      string(notification.NotificationStatusRead),
			"read_at":     &now,
			"updated_time": now,
		}).Error; err != nil {
		r.logger.Error("Failed to mark all notifications as read", 
			zap.String("user_id", userID),
			zap.Error(err))
		return err
	}

	return nil
}

// GetNotificationStats 获取通知统计
func (r *NotificationRepository) GetNotificationStats(ctx context.Context, userID string) (*notification.NotificationStats, error) {
	stats := &notification.NotificationStats{
		ByType:     make(map[notification.NotificationType]int64),
		ByPriority: make(map[notification.NotificationPriority]int64),
	}

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("user_id = ?", userID).
		Count(&stats.TotalNotifications).Error; err != nil {
		r.logger.Error("Failed to count total notifications", zap.Error(err))
		return nil, err
	}

	// 获取未读数量
	if err := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("user_id = ? AND status = ?", userID, string(notification.NotificationStatusUnread)).
		Count(&stats.UnreadCount).Error; err != nil {
		r.logger.Error("Failed to count unread notifications", zap.Error(err))
		return nil, err
	}

	// 获取已读数量
	if err := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("user_id = ? AND status = ?", userID, string(notification.NotificationStatusRead)).
		Count(&stats.ReadCount).Error; err != nil {
		r.logger.Error("Failed to count read notifications", zap.Error(err))
		return nil, err
	}

	// 获取已归档数量
	if err := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("user_id = ? AND status = ?", userID, string(notification.NotificationStatusArchived)).
		Count(&stats.ArchivedCount).Error; err != nil {
		r.logger.Error("Failed to count archived notifications", zap.Error(err))
		return nil, err
	}

	// 按类型统计
	var typeStats []struct {
		Type  string `json:"type"`
		Count int64  `json:"count"`
	}
	if err := r.db.WithContext(ctx).Model(&models.Notification{}).
		Select("type, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("type").
		Scan(&typeStats).Error; err != nil {
		r.logger.Error("Failed to get notifications by type", zap.Error(err))
		return nil, err
	}

	for _, stat := range typeStats {
		stats.ByType[notification.NotificationType(stat.Type)] = stat.Count
	}

	// 按优先级统计
	var priorityStats []struct {
		Priority string `json:"priority"`
		Count    int64  `json:"count"`
	}
	if err := r.db.WithContext(ctx).Model(&models.Notification{}).
		Select("priority, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("priority").
		Scan(&priorityStats).Error; err != nil {
		r.logger.Error("Failed to get notifications by priority", zap.Error(err))
		return nil, err
	}

	for _, stat := range priorityStats {
		stats.ByPriority[notification.NotificationPriority(stat.Priority)] = stat.Count
	}

	// 获取最近活动
	var recentModels []models.Notification
	if err := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("user_id = ?", userID).
		Order("created_time DESC").
		Limit(10).
		Find(&recentModels).Error; err != nil {
		r.logger.Error("Failed to get recent notifications", zap.Error(err))
		return nil, err
	}

	stats.RecentActivity = make([]*notification.Notification, len(recentModels))
	for i, model := range recentModels {
		stats.RecentActivity[i] = r.modelToDomain(&model)
	}

	return stats, nil
}

// CleanupExpiredNotifications 清理过期通知
func (r *NotificationRepository) CleanupExpiredNotifications(ctx context.Context) error {
	now := time.Now()
	
	if err := r.db.WithContext(ctx).Where("expires_at IS NOT NULL AND expires_at < ?", now).
		Delete(&models.Notification{}).Error; err != nil {
		r.logger.Error("Failed to cleanup expired notifications", zap.Error(err))
		return err
	}

	return nil
}

// domainToModel 领域对象转模型
func (r *NotificationRepository) domainToModel(n *notification.Notification) *models.Notification {
	model := &models.Notification{
		ID:          n.ID,
		UserID:      n.UserID,
		Type:        string(n.Type),
		Title:       n.Title,
		Content:     n.Content,
		Status:      string(n.Status),
		Priority:    string(n.Priority),
		SourceID:    n.SourceID,
		SourceType:  n.SourceType,
		ActionURL:   n.ActionURL,
		ExpiresAt:   n.ExpiresAt,
		ReadAt:      n.ReadAt,
		CreatedTime: n.CreatedTime,
		UpdatedTime: n.UpdatedTime,
	}

	// 序列化Data
	if n.Data != nil {
		if dataBytes, err := json.Marshal(n.Data); err == nil {
			model.Data = string(dataBytes)
		}
	}

	return model
}

// modelToDomain 模型转领域对象
func (r *NotificationRepository) modelToDomain(m *models.Notification) *notification.Notification {
	n := &notification.Notification{
		ID:          m.ID,
		UserID:      m.UserID,
		Type:        notification.NotificationType(m.Type),
		Title:       m.Title,
		Content:     m.Content,
		Status:      notification.NotificationStatus(m.Status),
		Priority:    notification.NotificationPriority(m.Priority),
		SourceID:    m.SourceID,
		SourceType:  m.SourceType,
		ActionURL:   m.ActionURL,
		ExpiresAt:   m.ExpiresAt,
		ReadAt:      m.ReadAt,
		CreatedTime: m.CreatedTime,
		UpdatedTime: m.UpdatedTime,
	}

	// 反序列化Data
	if m.Data != "" {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(m.Data), &data); err == nil {
			n.Data = data
		}
	}

	return n
}

// NotificationTemplateRepository 通知模板仓储实现
type NotificationTemplateRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewNotificationTemplateRepository 创建新的NotificationTemplateRepository
func NewNotificationTemplateRepository(db *gorm.DB, logger *zap.Logger) *NotificationTemplateRepository {
	return &NotificationTemplateRepository{
		db:     db,
		logger: logger,
	}
}

// CreateTemplate 创建模板
func (r *NotificationTemplateRepository) CreateTemplate(ctx context.Context, template *notification.NotificationTemplate) error {
	model := r.domainToModel(template)
	
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		r.logger.Error("Failed to create notification template", 
			zap.String("id", template.ID),
			zap.Error(err))
		return err
	}

	return nil
}

// GetTemplate 获取模板
func (r *NotificationTemplateRepository) GetTemplate(ctx context.Context, id string) (*notification.NotificationTemplate, error) {
	var model models.NotificationTemplate
	
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		r.logger.Error("Failed to get notification template", 
			zap.String("id", id),
			zap.Error(err))
		return nil, err
	}

	return r.modelToDomain(&model), nil
}

// GetTemplateByType 根据类型获取模板
func (r *NotificationTemplateRepository) GetTemplateByType(ctx context.Context, notificationType notification.NotificationType) (*notification.NotificationTemplate, error) {
	var model models.NotificationTemplate
	
	if err := r.db.WithContext(ctx).Where("type = ?", string(notificationType)).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		r.logger.Error("Failed to get notification template by type", 
			zap.String("type", string(notificationType)),
			zap.Error(err))
		return nil, err
	}

	return r.modelToDomain(&model), nil
}

// UpdateTemplate 更新模板
func (r *NotificationTemplateRepository) UpdateTemplate(ctx context.Context, template *notification.NotificationTemplate) error {
	model := r.domainToModel(template)
	
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		r.logger.Error("Failed to update notification template", 
			zap.String("id", template.ID),
			zap.Error(err))
		return err
	}

	return nil
}

// DeleteTemplate 删除模板
func (r *NotificationTemplateRepository) DeleteTemplate(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.NotificationTemplate{}).Error; err != nil {
		r.logger.Error("Failed to delete notification template", 
			zap.String("id", id),
			zap.Error(err))
		return err
	}

	return nil
}

// ListTemplates 列出模板
func (r *NotificationTemplateRepository) ListTemplates(ctx context.Context, notificationType *notification.NotificationType, isActive *bool) ([]*notification.NotificationTemplate, error) {
	query := r.db.WithContext(ctx).Model(&models.NotificationTemplate{})

	// 添加过滤条件
	if notificationType != nil {
		query = query.Where("type = ?", string(*notificationType))
	}
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	// 查询数据
	var models []models.NotificationTemplate
	if err := query.Find(&models).Error; err != nil {
		r.logger.Error("Failed to list notification templates", zap.Error(err))
		return nil, err
	}

	// 转换为领域对象
	templates := make([]*notification.NotificationTemplate, len(models))
	for i, model := range models {
		templates[i] = r.modelToDomain(&model)
	}

	return templates, nil
}

// domainToModel 领域对象转模型
func (r *NotificationTemplateRepository) domainToModel(t *notification.NotificationTemplate) *models.NotificationTemplate {
	model := &models.NotificationTemplate{
		ID:          t.ID,
		Type:        string(t.Type),
		Name:        t.Name,
		Title:       t.Title,
		Content:     t.Content,
		IsActive:    t.IsActive,
		CreatedTime: t.CreatedTime,
		UpdatedTime: t.UpdatedTime,
	}

	// 序列化Variables
	if t.Variables != nil {
		if variablesBytes, err := json.Marshal(t.Variables); err == nil {
			model.Variables = string(variablesBytes)
		}
	}

	// 序列化DefaultData
	if t.DefaultData != nil {
		if defaultDataBytes, err := json.Marshal(t.DefaultData); err == nil {
			model.DefaultData = string(defaultDataBytes)
		}
	}

	return model
}

// modelToDomain 模型转领域对象
func (r *NotificationTemplateRepository) modelToDomain(m *models.NotificationTemplate) *notification.NotificationTemplate {
	t := &notification.NotificationTemplate{
		ID:          m.ID,
		Type:        notification.NotificationType(m.Type),
		Name:        m.Name,
		Title:       m.Title,
		Content:     m.Content,
		IsActive:    m.IsActive,
		CreatedTime: m.CreatedTime,
		UpdatedTime: m.UpdatedTime,
	}

	// 反序列化Variables
	if m.Variables != "" {
		var variables []string
		if err := json.Unmarshal([]byte(m.Variables), &variables); err == nil {
			t.Variables = variables
		}
	}

	// 反序列化DefaultData
	if m.DefaultData != "" {
		var defaultData map[string]interface{}
		if err := json.Unmarshal([]byte(m.DefaultData), &defaultData); err == nil {
			t.DefaultData = defaultData
		}
	}

	return t
}

// NotificationSubscriptionRepository 通知订阅仓储实现
type NotificationSubscriptionRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewNotificationSubscriptionRepository 创建新的NotificationSubscriptionRepository
func NewNotificationSubscriptionRepository(db *gorm.DB, logger *zap.Logger) *NotificationSubscriptionRepository {
	return &NotificationSubscriptionRepository{
		db:     db,
		logger: logger,
	}
}

// CreateSubscription 创建订阅
func (r *NotificationSubscriptionRepository) CreateSubscription(ctx context.Context, subscription *notification.NotificationSubscription) error {
	model := r.domainToModel(subscription)
	
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		r.logger.Error("Failed to create notification subscription", 
			zap.String("id", subscription.ID),
			zap.Error(err))
		return err
	}

	return nil
}

// GetSubscription 获取订阅
func (r *NotificationSubscriptionRepository) GetSubscription(ctx context.Context, id string) (*notification.NotificationSubscription, error) {
	var model models.NotificationSubscription
	
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		r.logger.Error("Failed to get notification subscription", 
			zap.String("id", id),
			zap.Error(err))
		return nil, err
	}

	return r.modelToDomain(&model), nil
}

// GetUserSubscriptions 获取用户订阅
func (r *NotificationSubscriptionRepository) GetUserSubscriptions(ctx context.Context, userID string, notificationType *notification.NotificationType) ([]*notification.NotificationSubscription, error) {
	query := r.db.WithContext(ctx).Model(&models.NotificationSubscription{}).Where("user_id = ? AND is_active = ?", userID, true)

	// 添加类型过滤
	if notificationType != nil {
		query = query.Where("type = ?", string(*notificationType))
	}

	// 查询数据
	var models []models.NotificationSubscription
	if err := query.Find(&models).Error; err != nil {
		r.logger.Error("Failed to get user notification subscriptions", 
			zap.String("user_id", userID),
			zap.Error(err))
		return nil, err
	}

	// 转换为领域对象
	subscriptions := make([]*notification.NotificationSubscription, len(models))
	for i, model := range models {
		subscriptions[i] = r.modelToDomain(&model)
	}

	return subscriptions, nil
}

// GetSubscriptionsBySource 根据来源获取订阅
func (r *NotificationSubscriptionRepository) GetSubscriptionsBySource(ctx context.Context, sourceID, sourceType string, notificationType *notification.NotificationType) ([]*notification.NotificationSubscription, error) {
	query := r.db.WithContext(ctx).Model(&models.NotificationSubscription{}).Where("is_active = ?", true)

	// 添加来源过滤
	if sourceID != "" {
		query = query.Where("source_id = ?", sourceID)
	}
	if sourceType != "" {
		query = query.Where("source_type = ?", sourceType)
	}

	// 添加类型过滤
	if notificationType != nil {
		query = query.Where("type = ?", string(*notificationType))
	}

	// 查询数据
	var models []models.NotificationSubscription
	if err := query.Find(&models).Error; err != nil {
		r.logger.Error("Failed to get subscriptions by source", 
			zap.String("source_id", sourceID),
			zap.String("source_type", sourceType),
			zap.Error(err))
		return nil, err
	}

	// 转换为领域对象
	subscriptions := make([]*notification.NotificationSubscription, len(models))
	for i, model := range models {
		subscriptions[i] = r.modelToDomain(&model)
	}

	return subscriptions, nil
}

// UpdateSubscription 更新订阅
func (r *NotificationSubscriptionRepository) UpdateSubscription(ctx context.Context, subscription *notification.NotificationSubscription) error {
	model := r.domainToModel(subscription)
	
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		r.logger.Error("Failed to update notification subscription", 
			zap.String("id", subscription.ID),
			zap.Error(err))
		return err
	}

	return nil
}

// DeleteSubscription 删除订阅
func (r *NotificationSubscriptionRepository) DeleteSubscription(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.NotificationSubscription{}).Error; err != nil {
		r.logger.Error("Failed to delete notification subscription", 
			zap.String("id", id),
			zap.Error(err))
		return err
	}

	return nil
}

// DeleteUserSubscriptions 删除用户订阅
func (r *NotificationSubscriptionRepository) DeleteUserSubscriptions(ctx context.Context, userID string, notificationType *notification.NotificationType) error {
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)

	// 添加类型过滤
	if notificationType != nil {
		query = query.Where("type = ?", string(*notificationType))
	}

	if err := query.Delete(&models.NotificationSubscription{}).Error; err != nil {
		r.logger.Error("Failed to delete user notification subscriptions", 
			zap.String("user_id", userID),
			zap.Error(err))
		return err
	}

	return nil
}

// domainToModel 领域对象转模型
func (r *NotificationSubscriptionRepository) domainToModel(s *notification.NotificationSubscription) *models.NotificationSubscription {
	model := &models.NotificationSubscription{
		ID:          s.ID,
		UserID:      s.UserID,
		Type:        string(s.Type),
		SourceID:    s.SourceID,
		SourceType:  s.SourceType,
		IsActive:    s.IsActive,
		CreatedTime: s.CreatedTime,
		UpdatedTime: s.UpdatedTime,
	}

	// 序列化Channels
	if s.Channels != nil {
		if channelsBytes, err := json.Marshal(s.Channels); err == nil {
			model.Channels = string(channelsBytes)
		}
	}

	// 序列化Settings
	if s.Settings != nil {
		if settingsBytes, err := json.Marshal(s.Settings); err == nil {
			model.Settings = string(settingsBytes)
		}
	}

	return model
}

// modelToDomain 模型转领域对象
func (r *NotificationSubscriptionRepository) modelToDomain(m *models.NotificationSubscription) *notification.NotificationSubscription {
	s := &notification.NotificationSubscription{
		ID:          m.ID,
		UserID:      m.UserID,
		Type:        notification.NotificationType(m.Type),
		SourceID:    m.SourceID,
		SourceType:  m.SourceType,
		IsActive:    m.IsActive,
		CreatedTime: m.CreatedTime,
		UpdatedTime: m.UpdatedTime,
	}

	// 反序列化Channels
	if m.Channels != "" {
		var channels []string
		if err := json.Unmarshal([]byte(m.Channels), &channels); err == nil {
			s.Channels = channels
		}
	}

	// 反序列化Settings
	if m.Settings != "" {
		var settings map[string]interface{}
		if err := json.Unmarshal([]byte(m.Settings), &settings); err == nil {
			s.Settings = settings
		}
	}

	return s
}
