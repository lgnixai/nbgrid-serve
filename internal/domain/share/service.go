package share

import (
	"context"

	"go.uber.org/zap"

	"teable-go-backend/pkg/errors"
	"teable-go-backend/pkg/logger"
)

// Service 分享服务接口
type Service interface {
	// CreateShareView 创建分享视图
	CreateShareView(ctx context.Context, viewID, tableID, createdBy string) (*ShareView, error)
	// GetShareView 获取分享视图
	GetShareView(ctx context.Context, shareID string) (*ShareView, error)
	// EnableShareView 启用分享视图
	EnableShareView(ctx context.Context, shareID string, meta *ShareViewMeta) error
	// DisableShareView 禁用分享视图
	DisableShareView(ctx context.Context, shareID string) error
	// UpdateShareMeta 更新分享元数据
	UpdateShareMeta(ctx context.Context, shareID string, meta *ShareViewMeta) error
	// ValidateShareAccess 验证分享访问权限
	ValidateShareAccess(ctx context.Context, shareID, password string) (*ShareView, error)
	// GetShareViewInfo 获取分享视图信息
	GetShareViewInfo(ctx context.Context, shareID string) (*ShareViewInfo, error)
	// SubmitForm 提交表单
	SubmitForm(ctx context.Context, shareID string, req *ShareFormSubmitRequest) (*ShareFormSubmitResponse, error)
	// CopyData 复制数据
	CopyData(ctx context.Context, shareID string, req *ShareCopyRequest) (*ShareCopyResponse, error)
	// GetCollaborators 获取协作者
	GetCollaborators(ctx context.Context, shareID string, req *ShareCollaboratorsRequest) (*ShareCollaboratorsResponse, error)
	// GetLinkRecords 获取链接记录
	GetLinkRecords(ctx context.Context, shareID string, req *ShareLinkRecordsRequest) (*ShareLinkRecordsResponse, error)
	// GetShareStats 获取分享统计
	GetShareStats(ctx context.Context, tableID string) (*ShareStats, error)
}

// service 分享服务实现
type service struct {
	repo   Repository
	logger *zap.Logger
}

// NewService 创建分享服务
func NewService(repo Repository, logger *zap.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

// CreateShareView 创建分享视图
func (s *service) CreateShareView(ctx context.Context, viewID, tableID, createdBy string) (*ShareView, error) {
	// 检查是否已存在分享视图
	existing, err := s.repo.GetShareViewByViewID(ctx, viewID)
	if err != nil && err != errors.ErrNotFound {
		s.logger.Error("Failed to check existing share view",
			logger.String("view_id", viewID),
			logger.ErrorField(err),
		)
		return nil, errors.ErrInternalServer.WithDetails("Failed to create share view")
	}
	if existing != nil {
		return existing, nil
	}

	shareView := NewShareView(viewID, tableID, createdBy)
	if err := s.repo.CreateShareView(ctx, shareView); err != nil {
		s.logger.Error("Failed to create share view",
			logger.String("view_id", viewID),
			logger.String("table_id", tableID),
			logger.ErrorField(err),
		)
		return nil, errors.ErrInternalServer.WithDetails("Failed to create share view")
	}

	s.logger.Info("Share view created",
		logger.String("share_id", shareView.ShareID),
		logger.String("view_id", viewID),
		logger.String("table_id", tableID),
	)
	return shareView, nil
}

// GetShareView 获取分享视图
func (s *service) GetShareView(ctx context.Context, shareID string) (*ShareView, error) {
	shareView, err := s.repo.GetShareViewByShareID(ctx, shareID)
	if err != nil {
		if err == errors.ErrNotFound {
			return nil, errors.ErrNotFound.WithDetails("Share view not found")
		}
		s.logger.Error("Failed to get share view",
			logger.String("share_id", shareID),
			logger.ErrorField(err),
		)
		return nil, errors.ErrInternalServer.WithDetails("Failed to get share view")
	}
	return shareView, nil
}

// EnableShareView 启用分享视图
func (s *service) EnableShareView(ctx context.Context, shareID string, meta *ShareViewMeta) error {
	shareView, err := s.GetShareView(ctx, shareID)
	if err != nil {
		return err
	}

	shareView.Enable(meta)
	if err := s.repo.UpdateShareView(ctx, shareView); err != nil {
		s.logger.Error("Failed to enable share view",
			logger.String("share_id", shareID),
			logger.ErrorField(err),
		)
		return errors.ErrInternalServer.WithDetails("Failed to enable share view")
	}

	s.logger.Info("Share view enabled",
		logger.String("share_id", shareID),
	)
	return nil
}

// DisableShareView 禁用分享视图
func (s *service) DisableShareView(ctx context.Context, shareID string) error {
	shareView, err := s.GetShareView(ctx, shareID)
	if err != nil {
		return err
	}

	shareView.Disable()
	if err := s.repo.UpdateShareView(ctx, shareView); err != nil {
		s.logger.Error("Failed to disable share view",
			logger.String("share_id", shareID),
			logger.ErrorField(err),
		)
		return errors.ErrInternalServer.WithDetails("Failed to disable share view")
	}

	s.logger.Info("Share view disabled",
		logger.String("share_id", shareID),
	)
	return nil
}

// UpdateShareMeta 更新分享元数据
func (s *service) UpdateShareMeta(ctx context.Context, shareID string, meta *ShareViewMeta) error {
	shareView, err := s.GetShareView(ctx, shareID)
	if err != nil {
		return err
	}

	shareView.UpdateMeta(meta)
	if err := s.repo.UpdateShareView(ctx, shareView); err != nil {
		s.logger.Error("Failed to update share meta",
			logger.String("share_id", shareID),
			logger.ErrorField(err),
		)
		return errors.ErrInternalServer.WithDetails("Failed to update share meta")
	}

	s.logger.Info("Share meta updated",
		logger.String("share_id", shareID),
	)
	return nil
}

// ValidateShareAccess 验证分享访问权限
func (s *service) ValidateShareAccess(ctx context.Context, shareID, password string) (*ShareView, error) {
	shareView, err := s.GetShareView(ctx, shareID)
	if err != nil {
		return nil, err
	}

	if !shareView.EnableShare {
		return nil, errors.ErrForbidden.WithDetails("Share view is disabled")
	}

	if !shareView.ValidatePassword(password) {
		return nil, errors.ErrUnauthorized.WithDetails("Invalid password")
	}

	return shareView, nil
}

// GetShareViewInfo 获取分享视图信息
func (s *service) GetShareViewInfo(ctx context.Context, shareID string) (*ShareViewInfo, error) {
	shareView, err := s.GetShareView(ctx, shareID)
	if err != nil {
		return nil, err
	}

	if !shareView.EnableShare {
		return nil, errors.ErrForbidden.WithDetails("Share view is disabled")
	}

	// TODO: 获取视图、表格和字段数据
	// 这里需要调用其他服务来获取相关数据
	info := &ShareViewInfo{
		ShareView: shareView,
		View:      nil, // TODO: 获取视图数据
		Table:     nil, // TODO: 获取表格数据
		Fields:    nil, // TODO: 获取字段数据
	}

	return info, nil
}

// SubmitForm 提交表单
func (s *service) SubmitForm(ctx context.Context, shareID string, req *ShareFormSubmitRequest) (*ShareFormSubmitResponse, error) {
	shareView, err := s.GetShareView(ctx, shareID)
	if err != nil {
		return nil, err
	}

	if !shareView.EnableShare {
		return nil, errors.ErrForbidden.WithDetails("Share view is disabled")
	}

	if !shareView.AllowSubmit() {
		return nil, errors.ErrForbidden.WithDetails("Form submission is not allowed")
	}

	// TODO: 实现表单提交逻辑
	// 这里需要调用记录服务来创建记录
	response := &ShareFormSubmitResponse{
		RecordID: "temp_record_id", // TODO: 生成真实的记录ID
		Fields:   req.Fields,
	}

	s.logger.Info("Form submitted via share",
		logger.String("share_id", shareID),
		logger.String("record_id", response.RecordID),
	)
	return response, nil
}

// CopyData 复制数据
func (s *service) CopyData(ctx context.Context, shareID string, req *ShareCopyRequest) (*ShareCopyResponse, error) {
	shareView, err := s.GetShareView(ctx, shareID)
	if err != nil {
		return nil, err
	}

	if !shareView.EnableShare {
		return nil, errors.ErrForbidden.WithDetails("Share view is disabled")
	}

	if !shareView.AllowCopy() {
		return nil, errors.ErrForbidden.WithDetails("Copy is not allowed")
	}

	// TODO: 实现数据复制逻辑
	response := &ShareCopyResponse{
		Data: "copied_data", // TODO: 实现真实的数据复制
	}

	s.logger.Info("Data copied via share",
		logger.String("share_id", shareID),
	)
	return response, nil
}

// GetCollaborators 获取协作者
func (s *service) GetCollaborators(ctx context.Context, shareID string, req *ShareCollaboratorsRequest) (*ShareCollaboratorsResponse, error) {
	shareView, err := s.GetShareView(ctx, shareID)
	if err != nil {
		return nil, err
	}

	if !shareView.EnableShare {
		return nil, errors.ErrForbidden.WithDetails("Share view is disabled")
	}

	// TODO: 实现协作者获取逻辑
	response := &ShareCollaboratorsResponse{
		Collaborators: []interface{}{}, // TODO: 获取真实的协作者数据
	}

	return response, nil
}

// GetLinkRecords 获取链接记录
func (s *service) GetLinkRecords(ctx context.Context, shareID string, req *ShareLinkRecordsRequest) (*ShareLinkRecordsResponse, error) {
	shareView, err := s.GetShareView(ctx, shareID)
	if err != nil {
		return nil, err
	}

	if !shareView.EnableShare {
		return nil, errors.ErrForbidden.WithDetails("Share view is disabled")
	}

	// TODO: 实现链接记录获取逻辑
	response := &ShareLinkRecordsResponse{
		Records: []interface{}{}, // TODO: 获取真实的链接记录数据
	}

	return response, nil
}

// GetShareStats 获取分享统计
func (s *service) GetShareStats(ctx context.Context, tableID string) (*ShareStats, error) {
	stats, err := s.repo.GetShareStats(ctx, tableID)
	if err != nil {
		s.logger.Error("Failed to get share stats",
			logger.String("table_id", tableID),
			logger.ErrorField(err),
		)
		return nil, errors.ErrInternalServer.WithDetails("Failed to get share stats")
	}
	return stats, nil
}
