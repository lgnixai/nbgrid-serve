package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"teable-go-backend/internal/domain/attachment"
	"teable-go-backend/internal/infrastructure/database/models"
	"teable-go-backend/pkg/errors"
	"teable-go-backend/pkg/logger"
)

// AttachmentRepository 附件仓储实现
type AttachmentRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewAttachmentRepository 创建新的AttachmentRepository
func NewAttachmentRepository(db *gorm.DB, logger *zap.Logger) *AttachmentRepository {
	return &AttachmentRepository{
		db:     db,
		logger: logger,
	}
}

// CreateAttachment 创建附件
func (r *AttachmentRepository) CreateAttachment(ctx context.Context, attachment *attachment.AttachmentItem) error {
	model := r.domainToModel(attachment)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		r.logger.Error("Failed to create attachment in DB", logger.ErrorField(err))
		return fmt.Errorf("failed to create attachment: %w", err)
	}
	return nil
}

// GetAttachmentByID 通过ID获取附件
func (r *AttachmentRepository) GetAttachmentByID(ctx context.Context, id string) (*attachment.AttachmentItem, error) {
	var model models.Attachment
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		r.logger.Error("Failed to get attachment by ID from DB", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to get attachment: %w", err)
	}
	return r.modelToDomain(&model), nil
}

// GetAttachmentByToken 通过令牌获取附件
func (r *AttachmentRepository) GetAttachmentByToken(ctx context.Context, token string) (*attachment.AttachmentItem, error) {
	var model models.Attachment
	if err := r.db.WithContext(ctx).Where("token = ?", token).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		r.logger.Error("Failed to get attachment by token from DB", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to get attachment: %w", err)
	}
	return r.modelToDomain(&model), nil
}

// GetAttachmentByPath 通过路径获取附件
func (r *AttachmentRepository) GetAttachmentByPath(ctx context.Context, path string) (*attachment.AttachmentItem, error) {
	var model models.Attachment
	if err := r.db.WithContext(ctx).Where("path = ?", path).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		r.logger.Error("Failed to get attachment by path from DB", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to get attachment: %w", err)
	}
	return r.modelToDomain(&model), nil
}

// UpdateAttachment 更新附件
func (r *AttachmentRepository) UpdateAttachment(ctx context.Context, attachment *attachment.AttachmentItem) error {
	model := r.domainToModel(attachment)
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		r.logger.Error("Failed to update attachment in DB", logger.ErrorField(err))
		return fmt.Errorf("failed to update attachment: %w", err)
	}
	return nil
}

// DeleteAttachment 删除附件
func (r *AttachmentRepository) DeleteAttachment(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Attachment{}).Error; err != nil {
		r.logger.Error("Failed to delete attachment from DB", logger.ErrorField(err))
		return fmt.Errorf("failed to delete attachment: %w", err)
	}
	return nil
}

// ListAttachments 列出附件
func (r *AttachmentRepository) ListAttachments(ctx context.Context, tableID, fieldID, recordID string) ([]*attachment.AttachmentItem, error) {
	var models []*models.Attachment
	query := r.db.WithContext(ctx).Where("table_id = ?", tableID)

	if fieldID != "" {
		query = query.Where("field_id = ?", fieldID)
	}
	if recordID != "" {
		query = query.Where("record_id = ?", recordID)
	}

	if err := query.Find(&models).Error; err != nil {
		r.logger.Error("Failed to list attachments from DB", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to list attachments: %w", err)
	}

	attachments := make([]*attachment.AttachmentItem, len(models))
	for i, model := range models {
		attachments[i] = r.modelToDomain(model)
	}
	return attachments, nil
}

// GetAttachmentStats 获取附件统计信息
func (r *AttachmentRepository) GetAttachmentStats(ctx context.Context, tableID string) (*attachment.AttachmentStats, error) {
	var totalFiles int64
	if err := r.db.WithContext(ctx).Model(&models.Attachment{}).Where("table_id = ?", tableID).Count(&totalFiles).Error; err != nil {
		r.logger.Error("Failed to count total files from DB", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to get attachment stats: %w", err)
	}

	var totalSize int64
	if err := r.db.WithContext(ctx).Model(&models.Attachment{}).Where("table_id = ?", tableID).Select("COALESCE(SUM(size), 0)").Scan(&totalSize).Error; err != nil {
		r.logger.Error("Failed to sum total size from DB", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to get attachment stats: %w", err)
	}

	var imageFiles int64
	if err := r.db.WithContext(ctx).Model(&models.Attachment{}).Where("table_id = ? AND mime_type LIKE ?", tableID, "image/%").Count(&imageFiles).Error; err != nil {
		r.logger.Error("Failed to count image files from DB", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to get attachment stats: %w", err)
	}

	var videoFiles int64
	if err := r.db.WithContext(ctx).Model(&models.Attachment{}).Where("table_id = ? AND mime_type LIKE ?", tableID, "video/%").Count(&videoFiles).Error; err != nil {
		r.logger.Error("Failed to count video files from DB", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to get attachment stats: %w", err)
	}

	var audioFiles int64
	if err := r.db.WithContext(ctx).Model(&models.Attachment{}).Where("table_id = ? AND mime_type LIKE ?", tableID, "audio/%").Count(&audioFiles).Error; err != nil {
		r.logger.Error("Failed to count audio files from DB", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to get attachment stats: %w", err)
	}

	var documentFiles int64
	documentTypes := []string{
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/vnd.ms-powerpoint",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation",
		"text/plain",
		"text/csv",
	}

	for _, docType := range documentTypes {
		var count int64
		if err := r.db.WithContext(ctx).Model(&models.Attachment{}).Where("table_id = ? AND mime_type = ?", tableID, docType).Count(&count).Error; err != nil {
			r.logger.Error("Failed to count document files from DB", logger.ErrorField(err))
			return nil, fmt.Errorf("failed to get attachment stats: %w", err)
		}
		documentFiles += count
	}

	var otherFiles int64
	otherFiles = totalFiles - imageFiles - videoFiles - audioFiles - documentFiles

	var lastUploaded time.Time
	if err := r.db.WithContext(ctx).Model(&models.Attachment{}).Where("table_id = ?", tableID).Select("MAX(created_time)").Scan(&lastUploaded).Error; err != nil {
		r.logger.Error("Failed to get last uploaded time from DB", logger.ErrorField(err))
		lastUploaded = time.Time{} // 使用零值
	}

	return &attachment.AttachmentStats{
		TotalFiles:    totalFiles,
		TotalSize:     totalSize,
		ImageFiles:    imageFiles,
		VideoFiles:    videoFiles,
		AudioFiles:    audioFiles,
		DocumentFiles: documentFiles,
		OtherFiles:    otherFiles,
		LastUploaded:  lastUploaded,
	}, nil
}

// domainToModel 领域实体转模型
func (r *AttachmentRepository) domainToModel(attachment *attachment.AttachmentItem) *models.Attachment {
	model := &models.Attachment{
		ID:             attachment.ID,
		Name:           attachment.Name,
		Path:           attachment.Path,
		Token:          attachment.Token,
		Size:           attachment.Size,
		MimeType:       attachment.MimeType,
		PresignedURL:   attachment.PresignedURL,
		Width:          attachment.Width,
		Height:         attachment.Height,
		SmallThumbnail: attachment.SmallThumbnail,
		LargeThumbnail: attachment.LargeThumbnail,
		CreatedTime:    attachment.CreatedTime,
		UpdatedTime:    attachment.UpdatedTime,
	}

	// TODO: 从上下文中获取这些信息，或者作为参数传递
	// 这里暂时使用默认值
	model.TableID = "temp_table_id"
	model.FieldID = "temp_field_id"
	model.RecordID = "temp_record_id"
	model.CreatedBy = "temp_user_id"

	return model
}

// modelToDomain 模型转领域实体
func (r *AttachmentRepository) modelToDomain(model *models.Attachment) *attachment.AttachmentItem {
	return &attachment.AttachmentItem{
		ID:             model.ID,
		Name:           model.Name,
		Path:           model.Path,
		Token:          model.Token,
		Size:           model.Size,
		MimeType:       model.MimeType,
		PresignedURL:   model.PresignedURL,
		Width:          model.Width,
		Height:         model.Height,
		SmallThumbnail: model.SmallThumbnail,
		LargeThumbnail: model.LargeThumbnail,
		CreatedTime:    model.CreatedTime,
		UpdatedTime:    model.UpdatedTime,
	}
}

// UploadTokenRepository 上传令牌仓储实现
type UploadTokenRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewUploadTokenRepository 创建新的UploadTokenRepository
func NewUploadTokenRepository(db *gorm.DB, logger *zap.Logger) *UploadTokenRepository {
	return &UploadTokenRepository{
		db:     db,
		logger: logger,
	}
}

// CreateUploadToken 创建上传令牌
func (r *UploadTokenRepository) CreateUploadToken(ctx context.Context, token *attachment.UploadToken) error {
	model := r.domainToModel(token)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		r.logger.Error("Failed to create upload token in DB", logger.ErrorField(err))
		return fmt.Errorf("failed to create upload token: %w", err)
	}
	return nil
}

// GetUploadToken 获取上传令牌
func (r *UploadTokenRepository) GetUploadToken(ctx context.Context, token string) (*attachment.UploadToken, error) {
	var model models.UploadToken
	if err := r.db.WithContext(ctx).Where("token = ?", token).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		r.logger.Error("Failed to get upload token from DB", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to get upload token: %w", err)
	}
	return r.modelToDomain(&model), nil
}

// DeleteUploadToken 删除上传令牌
func (r *UploadTokenRepository) DeleteUploadToken(ctx context.Context, token string) error {
	if err := r.db.WithContext(ctx).Where("token = ?", token).Delete(&models.UploadToken{}).Error; err != nil {
		r.logger.Error("Failed to delete upload token from DB", logger.ErrorField(err))
		return fmt.Errorf("failed to delete upload token: %w", err)
	}
	return nil
}

// CleanupExpiredTokens 清理过期令牌
func (r *UploadTokenRepository) CleanupExpiredTokens(ctx context.Context) error {
	if err := r.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&models.UploadToken{}).Error; err != nil {
		r.logger.Error("Failed to cleanup expired tokens from DB", logger.ErrorField(err))
		return fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}
	return nil
}

// domainToModel 领域实体转模型
func (r *UploadTokenRepository) domainToModel(token *attachment.UploadToken) *models.UploadToken {
	model := &models.UploadToken{
		Token:       token.Token,
		UserID:      token.UserID,
		TableID:     token.TableID,
		FieldID:     token.FieldID,
		RecordID:    token.RecordID,
		ExpiresAt:   token.ExpiresAt,
		MaxSize:     token.MaxSize,
		CreatedTime: token.CreatedTime,
	}

	// 序列化允许的文件类型
	if len(token.AllowedTypes) > 0 {
		typesBytes, err := json.Marshal(token.AllowedTypes)
		if err != nil {
			r.logger.Error("Failed to marshal allowed types", logger.ErrorField(err))
		} else {
			typesStr := string(typesBytes)
			model.AllowedTypes = &typesStr
		}
	}

	return model
}

// modelToDomain 模型转领域实体
func (r *UploadTokenRepository) modelToDomain(model *models.UploadToken) *attachment.UploadToken {
	token := &attachment.UploadToken{
		Token:       model.Token,
		UserID:      model.UserID,
		TableID:     model.TableID,
		FieldID:     model.FieldID,
		RecordID:    model.RecordID,
		ExpiresAt:   model.ExpiresAt,
		MaxSize:     model.MaxSize,
		CreatedTime: model.CreatedTime,
	}

	// 反序列化允许的文件类型
	if model.AllowedTypes != nil {
		var allowedTypes []string
		if err := json.Unmarshal([]byte(*model.AllowedTypes), &allowedTypes); err != nil {
			r.logger.Error("Failed to unmarshal allowed types", logger.ErrorField(err))
		} else {
			token.AllowedTypes = allowedTypes
		}
	}

	return token
}
