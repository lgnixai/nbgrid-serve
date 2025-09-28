package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"teable-go-backend/internal/domain/share"
	"teable-go-backend/internal/infrastructure/database/models"
	"teable-go-backend/pkg/errors"
	"teable-go-backend/pkg/logger"
)

// ShareRepository 分享仓储实现
type ShareRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewShareRepository 创建新的ShareRepository
func NewShareRepository(db *gorm.DB, logger *zap.Logger) *ShareRepository {
	return &ShareRepository{
		db:     db,
		logger: logger,
	}
}

// CreateShareView 创建分享视图
func (r *ShareRepository) CreateShareView(ctx context.Context, shareView *share.ShareView) error {
	model := r.domainToModel(shareView)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		r.logger.Error("Failed to create share view in DB", logger.ErrorField(err))
		return fmt.Errorf("failed to create share view: %w", err)
	}
	return nil
}

// GetShareViewByShareID 通过分享ID获取分享视图
func (r *ShareRepository) GetShareViewByShareID(ctx context.Context, shareID string) (*share.ShareView, error) {
	var model models.ShareView
	if err := r.db.WithContext(ctx).Where("share_id = ?", shareID).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		r.logger.Error("Failed to get share view by share ID from DB", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to get share view: %w", err)
	}
	return r.modelToDomain(&model), nil
}

// GetShareViewByViewID 通过视图ID获取分享视图
func (r *ShareRepository) GetShareViewByViewID(ctx context.Context, viewID string) (*share.ShareView, error) {
	var model models.ShareView
	if err := r.db.WithContext(ctx).Where("view_id = ?", viewID).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		r.logger.Error("Failed to get share view by view ID from DB", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to get share view: %w", err)
	}
	return r.modelToDomain(&model), nil
}

// UpdateShareView 更新分享视图
func (r *ShareRepository) UpdateShareView(ctx context.Context, shareView *share.ShareView) error {
	model := r.domainToModel(shareView)
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		r.logger.Error("Failed to update share view in DB", logger.ErrorField(err))
		return fmt.Errorf("failed to update share view: %w", err)
	}
	return nil
}

// DeleteShareView 删除分享视图
func (r *ShareRepository) DeleteShareView(ctx context.Context, shareID string) error {
	if err := r.db.WithContext(ctx).Where("share_id = ?", shareID).Delete(&models.ShareView{}).Error; err != nil {
		r.logger.Error("Failed to delete share view from DB", logger.ErrorField(err))
		return fmt.Errorf("failed to delete share view: %w", err)
	}
	return nil
}

// ListShareViews 列出分享视图
func (r *ShareRepository) ListShareViews(ctx context.Context, tableID string) ([]*share.ShareView, error) {
	var models []*models.ShareView
	if err := r.db.WithContext(ctx).Where("table_id = ?", tableID).Find(&models).Error; err != nil {
		r.logger.Error("Failed to list share views from DB", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to list share views: %w", err)
	}

	shareViews := make([]*share.ShareView, len(models))
	for i, model := range models {
		shareViews[i] = r.modelToDomain(model)
	}
	return shareViews, nil
}

// GetShareStats 获取分享统计信息
func (r *ShareRepository) GetShareStats(ctx context.Context, tableID string) (*share.ShareStats, error) {
	var totalShares int64
	if err := r.db.WithContext(ctx).Model(&models.ShareView{}).Where("table_id = ?", tableID).Count(&totalShares).Error; err != nil {
		r.logger.Error("Failed to count total shares from DB", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to get share stats: %w", err)
	}

	var activeShares int64
	if err := r.db.WithContext(ctx).Model(&models.ShareView{}).Where("table_id = ? AND enable_share = ?", tableID, true).Count(&activeShares).Error; err != nil {
		r.logger.Error("Failed to count active shares from DB", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to get share stats: %w", err)
	}

	var passwordProtected int64
	if err := r.db.WithContext(ctx).Model(&models.ShareView{}).
		Where("table_id = ? AND enable_share = ? AND share_meta LIKE ?", tableID, true, "%password%").
		Count(&passwordProtected).Error; err != nil {
		r.logger.Error("Failed to count password protected shares from DB", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to get share stats: %w", err)
	}

	// 获取最后访问时间（这里简化处理，实际可能需要单独的表来记录访问日志）
	lastAccessed := time.Now()

	return &share.ShareStats{
		TotalShares:        totalShares,
		ActiveShares:       activeShares,
		PasswordProtected:  passwordProtected,
		LastAccessed:       lastAccessed,
	}, nil
}

// domainToModel 领域实体转模型
func (r *ShareRepository) domainToModel(shareView *share.ShareView) *models.ShareView {
	model := &models.ShareView{
		ID:          shareView.ID,
		ViewID:      shareView.ViewID,
		TableID:     shareView.TableID,
		ShareID:     shareView.ShareID,
		EnableShare: shareView.EnableShare,
		CreatedBy:   shareView.CreatedBy,
		CreatedTime: shareView.CreatedTime,
		UpdatedTime: shareView.UpdatedTime,
	}

	// 序列化分享元数据
	if shareView.ShareMeta != nil {
		metaBytes, err := json.Marshal(shareView.ShareMeta)
		if err != nil {
			r.logger.Error("Failed to marshal share meta", logger.ErrorField(err))
		} else {
			metaStr := string(metaBytes)
			model.ShareMeta = &metaStr
		}
	}

	return model
}

// modelToDomain 模型转领域实体
func (r *ShareRepository) modelToDomain(model *models.ShareView) *share.ShareView {
	shareView := &share.ShareView{
		ID:          model.ID,
		ViewID:      model.ViewID,
		TableID:     model.TableID,
		ShareID:     model.ShareID,
		EnableShare: model.EnableShare,
		CreatedBy:   model.CreatedBy,
		CreatedTime: model.CreatedTime,
		UpdatedTime: model.UpdatedTime,
	}

	// 反序列化分享元数据
	if model.ShareMeta != nil {
		var meta share.ShareViewMeta
		if err := json.Unmarshal([]byte(*model.ShareMeta), &meta); err != nil {
			r.logger.Error("Failed to unmarshal share meta", logger.ErrorField(err))
		} else {
			shareView.ShareMeta = &meta
		}
	}

	return shareView
}
