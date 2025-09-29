package record

import (
	"context"
	"fmt"

	"teable-go-backend/internal/domain/table"
	"teable-go-backend/pkg/errors"
)

// Service 记录服务接口
type Service interface {
	CreateRecord(ctx context.Context, req CreateRecordRequest) (*Record, error)
	GetRecord(ctx context.Context, id string) (*Record, error)
	UpdateRecord(ctx context.Context, id string, req UpdateRecordRequest) (*Record, error)
	DeleteRecord(ctx context.Context, id string) error
	ListRecords(ctx context.Context, filter ListRecordFilter) ([]*Record, int64, error)

	// 批量操作
	BulkCreateRecords(ctx context.Context, reqs []CreateRecordRequest) ([]*Record, error)
	BulkUpdateRecords(ctx context.Context, req BulkUpdateRequest) error
	BulkDeleteRecords(ctx context.Context, req BulkDeleteRequest) error

	// 复杂查询
	ComplexQuery(ctx context.Context, req ComplexQueryRequest) ([]map[string]interface{}, error)

	// 统计
	GetRecordStats(ctx context.Context, tableID *string) (*RecordStats, error)

	// 导出导入
	ExportRecords(ctx context.Context, req ExportRequest) ([]byte, error)
	ImportRecords(ctx context.Context, req ImportRequest) (int, error)
}

// ServiceImpl 记录服务实现
type ServiceImpl struct {
	repo Repository
}

// NewService 创建记录服务
func NewService(repo Repository) Service {
	return &ServiceImpl{repo: repo}
}

// CreateRecord 创建记录
func (s *ServiceImpl) CreateRecord(ctx context.Context, req CreateRecordRequest) (*Record, error) {
	record := NewRecord(req)

	// 获取表格schema进行验证
	tableSchema, err := s.getTableSchema(ctx, req.TableID)
	if err != nil {
		return nil, fmt.Errorf("获取表格schema失败: %v", err)
	}

	record.SetTableSchema(tableSchema)

	// 应用字段默认值
	if err := record.ApplyFieldDefaults(); err != nil {
		return nil, fmt.Errorf("应用字段默认值失败: %v", err)
	}

	// 验证记录数据
	if err := record.ValidateData(); err != nil {
		return nil, fmt.Errorf("记录数据验证失败: %v", err)
	}

	if err := s.repo.Create(ctx, record); err != nil {
		return nil, err
	}
	return record, nil
}

// GetRecord 获取记录
func (s *ServiceImpl) GetRecord(ctx context.Context, id string) (*Record, error) {
	record, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, errors.ErrNotFound.WithDetails("记录未找到")
	}
	return record, nil
}

// UpdateRecord 更新记录
func (s *ServiceImpl) UpdateRecord(ctx context.Context, id string, req UpdateRecordRequest) (*Record, error) {
	record, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, errors.ErrNotFound.WithDetails("记录未找到")
	}

	// 获取表格schema进行验证
	tableSchema, err := s.getTableSchema(ctx, record.TableID)
	if err != nil {
		return nil, fmt.Errorf("获取表格schema失败: %v", err)
	}

	record.SetTableSchema(tableSchema)

	// 更新记录数据
	updatedBy := req.UpdatedBy
	if updatedBy == "" {
		updatedBy = record.CreatedBy // 如果没有指定更新者，使用创建者
	}

	if err := record.Update(req, updatedBy); err != nil {
		return nil, fmt.Errorf("更新记录失败: %v", err)
	}

	if err := s.repo.Update(ctx, record); err != nil {
		return nil, err
	}
	return record, nil
}

// DeleteRecord 删除记录
func (s *ServiceImpl) DeleteRecord(ctx context.Context, id string) error {
	record, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if record == nil {
		return errors.ErrNotFound.WithDetails("记录未找到")
	}

	record.SoftDelete() // 软删除
	if err := s.repo.Update(ctx, record); err != nil {
		return err
	}
	return nil
}

// ListRecords 列出记录
func (s *ServiceImpl) ListRecords(ctx context.Context, filter ListRecordFilter) ([]*Record, int64, error) {
	records, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.Count(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

// BulkCreateRecords 批量创建记录
func (s *ServiceImpl) BulkCreateRecords(ctx context.Context, reqs []CreateRecordRequest) ([]*Record, error) {
	if len(reqs) == 0 {
		return []*Record{}, nil
	}

	records := make([]*Record, len(reqs))
	for i, req := range reqs {
		records[i] = NewRecord(req)
	}

	if err := s.repo.BulkCreate(ctx, records); err != nil {
		return nil, err
	}

	return records, nil
}

// BulkUpdateRecords 批量更新记录
func (s *ServiceImpl) BulkUpdateRecords(ctx context.Context, req BulkUpdateRequest) error {
	if len(req.RecordIDs) == 0 {
		return errors.ErrInvalidRequest.WithDetails("记录ID列表不能为空")
	}

	updates := make(map[string]map[string]interface{})
	for _, recordID := range req.RecordIDs {
		updates[recordID] = req.Updates
	}

	return s.repo.BulkUpdate(ctx, updates)
}

// BulkDeleteRecords 批量删除记录
func (s *ServiceImpl) BulkDeleteRecords(ctx context.Context, req BulkDeleteRequest) error {
	if len(req.RecordIDs) == 0 {
		return errors.ErrInvalidRequest.WithDetails("记录ID列表不能为空")
	}

	return s.repo.BulkDelete(ctx, req.RecordIDs)
}

// ComplexQuery 复杂查询
func (s *ServiceImpl) ComplexQuery(ctx context.Context, req ComplexQueryRequest) ([]map[string]interface{}, error) {
	if req.TableID == "" {
		return nil, errors.ErrInvalidRequest.WithDetails("表ID不能为空")
	}

	return s.repo.ComplexQuery(ctx, req)
}

// GetRecordStats 获取记录统计信息
func (s *ServiceImpl) GetRecordStats(ctx context.Context, tableID *string) (*RecordStats, error) {
	return s.repo.GetRecordStats(ctx, tableID)
}

// ExportRecords 导出记录
func (s *ServiceImpl) ExportRecords(ctx context.Context, req ExportRequest) ([]byte, error) {
	if req.Format == "" {
		req.Format = "json" // 默认格式
	}

	return s.repo.ExportRecords(ctx, req)
}

// ImportRecords 导入记录
func (s *ServiceImpl) ImportRecords(ctx context.Context, req ImportRequest) (int, error) {
	if req.TableID == "" {
		return 0, errors.ErrInvalidRequest.WithDetails("表ID不能为空")
	}
	if req.Format == "" {
		req.Format = "json" // 默认格式
	}

	return s.repo.ImportRecords(ctx, req)
}

// getTableSchema 获取表格schema（这里需要依赖注入表格服务）
func (s *ServiceImpl) getTableSchema(ctx context.Context, tableID string) (*table.Table, error) {
	// 这里应该通过依赖注入的表格服务获取schema
	// 为了简化，暂时返回一个模拟的schema
	// 在实际实现中，需要注入TableService
	return &table.Table{
		ID: tableID,
		// 其他字段...
	}, nil
}
