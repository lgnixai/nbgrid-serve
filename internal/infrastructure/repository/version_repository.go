package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"teable-go-backend/internal/application"
	"teable-go-backend/internal/infrastructure/database/models"
	"teable-go-backend/pkg/utils"
)

// VersionRepository 记录版本仓储实现
type VersionRepository struct {
	db *gorm.DB
}

// NewVersionRepository 创建记录版本仓储
func NewVersionRepository(db *gorm.DB) *VersionRepository {
	return &VersionRepository{
		db: db,
	}
}

// SaveVersion 保存版本
func (r *VersionRepository) SaveVersion(ctx context.Context, version *application.RecordVersion) error {
	model := &models.RecordVersion{
		ID:         version.ID,
		RecordID:   version.RecordID,
		Version:    version.Version,
		Data:       version.Data,
		ChangeType: version.ChangeType,
		ChangedBy:  version.ChangedBy,
		ChangedAt:  version.ChangedAt,
	}

	if model.ID == "" {
		model.ID = utils.GenerateID()
	}

	if model.ChangedAt.IsZero() {
		model.ChangedAt = time.Now()
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to save version: %w", err)
	}

	// 更新传入的version的ID（如果是新生成的）
	version.ID = model.ID

	return nil
}

// GetVersion 获取指定版本
func (r *VersionRepository) GetVersion(ctx context.Context, versionID string) (*application.RecordVersion, error) {
	var model models.RecordVersion
	
	if err := r.db.WithContext(ctx).
		Where("id = ?", versionID).
		First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get version: %w", err)
	}

	return r.toRecordVersion(&model), nil
}

// GetRecordVersions 获取记录的所有版本
func (r *VersionRepository) GetRecordVersions(ctx context.Context, recordID string) ([]*application.RecordVersion, error) {
	var models []models.RecordVersion
	
	if err := r.db.WithContext(ctx).
		Where("record_id = ?", recordID).
		Order("version DESC, changed_at DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get record versions: %w", err)
	}

	return r.toRecordVersions(models), nil
}

// DeleteVersion 删除版本
func (r *VersionRepository) DeleteVersion(ctx context.Context, versionID string) error {
	result := r.db.WithContext(ctx).
		Where("id = ?", versionID).
		Delete(&models.RecordVersion{})
	
	if result.Error != nil {
		return fmt.Errorf("failed to delete version: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("version not found: %s", versionID)
	}
	
	return nil
}

// GetVersionByRecordAndVersion 根据记录ID和版本号获取版本
func (r *VersionRepository) GetVersionByRecordAndVersion(ctx context.Context, recordID string, version int64) (*application.RecordVersion, error) {
	var model models.RecordVersion
	
	if err := r.db.WithContext(ctx).
		Where("record_id = ? AND version = ?", recordID, version).
		First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get version: %w", err)
	}

	return r.toRecordVersion(&model), nil
}

// CleanupOldVersions 清理旧版本（保留最近N个版本）
func (r *VersionRepository) CleanupOldVersions(ctx context.Context, recordID string, keepCount int) error {
	if keepCount <= 0 {
		return fmt.Errorf("keepCount must be positive")
	}

	// 获取所有版本，按版本号降序
	var versions []models.RecordVersion
	if err := r.db.WithContext(ctx).
		Where("record_id = ?", recordID).
		Order("version DESC").
		Find(&versions).Error; err != nil {
		return fmt.Errorf("failed to get versions: %w", err)
	}

	// 如果版本数量不超过保留数量，无需清理
	if len(versions) <= keepCount {
		return nil
	}

	// 获取要删除的版本ID
	toDelete := versions[keepCount:]
	versionIDs := make([]string, len(toDelete))
	for i, v := range toDelete {
		versionIDs[i] = v.ID
	}

	// 批量删除
	if err := r.db.WithContext(ctx).
		Where("id IN ?", versionIDs).
		Delete(&models.RecordVersion{}).Error; err != nil {
		return fmt.Errorf("failed to delete old versions: %w", err)
	}

	return nil
}

// toRecordVersion 将模型转换为应用层版本对象
func (r *VersionRepository) toRecordVersion(model *models.RecordVersion) *application.RecordVersion {
	return &application.RecordVersion{
		ID:         model.ID,
		RecordID:   model.RecordID,
		Version:    model.Version,
		Data:       model.Data,
		ChangeType: model.ChangeType,
		ChangedBy:  model.ChangedBy,
		ChangedAt:  model.ChangedAt,
	}
}

// toRecordVersions 批量转换
func (r *VersionRepository) toRecordVersions(models []models.RecordVersion) []*application.RecordVersion {
	versions := make([]*application.RecordVersion, len(models))
	for i, model := range models {
		versions[i] = r.toRecordVersion(&model)
	}
	return versions
}