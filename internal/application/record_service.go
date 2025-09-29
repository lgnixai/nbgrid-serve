package application

import (
	"context"
	"fmt"
	"time"

	"teable-go-backend/internal/domain/permission"
	"teable-go-backend/internal/domain/record"
	"teable-go-backend/internal/domain/table"
	"teable-go-backend/pkg/errors"
	"teable-go-backend/pkg/logger"
)

// RecordService 记录应用服务 - 重构版本
// 支持动态schema、版本历史、权限控制和变更追踪
type RecordService struct {
	recordRepo     record.Repository
	tableService   table.Service
	permissionSvc  permission.Service
	changeTracker  *RecordChangeTracker
	validator      *RecordValidator
	versionManager *RecordVersionManager
}

// NewRecordService 创建记录应用服务
func NewRecordService(
	recordRepo record.Repository,
	tableService table.Service,
	permissionSvc permission.Service,
) *RecordService {
	return &RecordService{
		recordRepo:     recordRepo,
		tableService:   tableService,
		permissionSvc:  permissionSvc,
		changeTracker:  NewRecordChangeTracker(),
		validator:      NewRecordValidator(tableService),
		versionManager: NewRecordVersionManager(recordRepo),
	}
}

// CreateRecord 创建记录 - 支持动态schema验证和权限控制
func (s *RecordService) CreateRecord(ctx context.Context, req record.CreateRecordRequest, userID string) (*record.Record, error) {
	// 1. 权限检查
	if err := s.checkPermission(ctx, userID, req.TableID, "record:create"); err != nil {
		return nil, err
	}

	// 2. 获取表格schema
	tableSchema, err := s.tableService.GetTable(ctx, req.TableID)
	if err != nil {
		return nil, fmt.Errorf("获取表格schema失败: %v", err)
	}

	// 3. 创建记录实体
	req.CreatedBy = userID
	newRecord := record.NewRecord(req)
	newRecord.SetTableSchema(tableSchema)

	// 4. 动态schema验证
	if err := s.validator.ValidateForCreate(ctx, newRecord, tableSchema); err != nil {
		return nil, fmt.Errorf("记录验证失败: %v", err)
	}

	// 5. 应用字段默认值和系统字段
	if err := newRecord.ApplyFieldDefaults(); err != nil {
		return nil, fmt.Errorf("应用字段默认值失败: %v", err)
	}

	// 6. 保存记录
	if err := s.recordRepo.Create(ctx, newRecord); err != nil {
		return nil, fmt.Errorf("保存记录失败: %v", err)
	}

	// 7. 记录变更事件
	changeEvent := newRecord.CreateChangeEvent("create", nil, userID)
	if err := s.changeTracker.TrackChange(ctx, changeEvent); err != nil {
		logger.Error("记录变更追踪失败", logger.ErrorField(err))
	}

	// 8. 创建版本历史
	if err := s.versionManager.CreateVersion(ctx, newRecord, "create", userID); err != nil {
		logger.Error("创建版本历史失败", logger.ErrorField(err))
	}

	logger.Info("记录创建成功",
		logger.String("record_id", newRecord.ID),
		logger.String("table_id", req.TableID),
		logger.String("user_id", userID),
	)

	return newRecord, nil
}

// GetRecord 获取记录 - 支持权限控制
func (s *RecordService) GetRecord(ctx context.Context, recordID string, userID string) (*record.Record, error) {
	// 1. 获取记录
	rec, err := s.recordRepo.GetByID(ctx, recordID)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, errors.ErrNotFound.WithDetails("记录未找到")
	}

	// 2. 权限检查
	if err := s.checkPermission(ctx, userID, rec.TableID, "record:read"); err != nil {
		return nil, err
	}

	// 3. 获取表格schema用于数据处理
	tableSchema, err := s.tableService.GetTable(ctx, rec.TableID)
	if err != nil {
		logger.Error("获取表格schema失败", logger.ErrorField(err))
	} else {
		rec.SetTableSchema(tableSchema)
	}

	return rec, nil
}

// UpdateRecord 更新记录 - 支持动态schema验证、版本控制和变更追踪
func (s *RecordService) UpdateRecord(ctx context.Context, recordID string, req record.UpdateRecordRequest, userID string) (*record.Record, error) {
	// 1. 获取现有记录
	existingRecord, err := s.recordRepo.GetByID(ctx, recordID)
	if err != nil {
		return nil, err
	}
	if existingRecord == nil {
		return nil, errors.ErrNotFound.WithDetails("记录未找到")
	}

	// 2. 权限检查
	if err := s.checkPermission(ctx, userID, existingRecord.TableID, "record:update"); err != nil {
		return nil, err
	}

	// 3. 获取表格schema
	tableSchema, err := s.tableService.GetTable(ctx, existingRecord.TableID)
	if err != nil {
		return nil, fmt.Errorf("获取表格schema失败: %v", err)
	}

	// 4. 保存旧数据用于变更追踪
	oldData := make(map[string]interface{})
	for k, v := range existingRecord.Data {
		oldData[k] = v
	}

	// 5. 设置schema并更新记录
	existingRecord.SetTableSchema(tableSchema)
	req.UpdatedBy = userID

	if err := existingRecord.Update(req, userID); err != nil {
		return nil, fmt.Errorf("更新记录失败: %v", err)
	}

	// 6. 动态schema验证
	if err := s.validator.ValidateForUpdate(ctx, existingRecord, tableSchema); err != nil {
		return nil, fmt.Errorf("记录验证失败: %v", err)
	}

	// 7. 保存更新
	if err := s.recordRepo.Update(ctx, existingRecord); err != nil {
		return nil, fmt.Errorf("保存记录更新失败: %v", err)
	}

	// 8. 记录变更事件
	changeEvent := existingRecord.CreateChangeEvent("update", oldData, userID)
	if err := s.changeTracker.TrackChange(ctx, changeEvent); err != nil {
		logger.Error("记录变更追踪失败", logger.ErrorField(err))
	}

	// 9. 创建版本历史
	if err := s.versionManager.CreateVersion(ctx, existingRecord, "update", userID); err != nil {
		logger.Error("创建版本历史失败", logger.ErrorField(err))
	}

	logger.Info("记录更新成功",
		logger.String("record_id", recordID),
		logger.String("table_id", existingRecord.TableID),
		logger.String("user_id", userID),
		logger.Int64("version", existingRecord.Version),
	)

	return existingRecord, nil
}

// DeleteRecord 删除记录 - 支持权限控制和变更追踪
func (s *RecordService) DeleteRecord(ctx context.Context, recordID string, userID string) error {
	// 1. 获取记录
	rec, err := s.recordRepo.GetByID(ctx, recordID)
	if err != nil {
		return err
	}
	if rec == nil {
		return errors.ErrNotFound.WithDetails("记录未找到")
	}

	// 2. 权限检查
	if err := s.checkPermission(ctx, userID, rec.TableID, "record:delete"); err != nil {
		return err
	}

	// 3. 保存旧数据用于变更追踪
	oldData := make(map[string]interface{})
	for k, v := range rec.Data {
		oldData[k] = v
	}

	// 4. 软删除记录
	rec.SoftDelete()
	if err := s.recordRepo.Update(ctx, rec); err != nil {
		return fmt.Errorf("删除记录失败: %v", err)
	}

	// 5. 记录变更事件
	changeEvent := rec.CreateChangeEvent("delete", oldData, userID)
	if err := s.changeTracker.TrackChange(ctx, changeEvent); err != nil {
		logger.Error("记录变更追踪失败", logger.ErrorField(err))
	}

	// 6. 创建版本历史
	if err := s.versionManager.CreateVersion(ctx, rec, "delete", userID); err != nil {
		logger.Error("创建版本历史失败", logger.ErrorField(err))
	}

	logger.Info("记录删除成功",
		logger.String("record_id", recordID),
		logger.String("table_id", rec.TableID),
		logger.String("user_id", userID),
	)

	return nil
}

// ListRecords 列出记录 - 支持权限过滤
func (s *RecordService) ListRecords(ctx context.Context, filter record.ListRecordFilter, userID string) ([]*record.Record, int64, error) {
	// 1. 权限检查 - 如果指定了表ID，检查表的读取权限
	if filter.TableID != nil {
		if err := s.checkPermission(ctx, userID, *filter.TableID, "record:read"); err != nil {
			return nil, 0, err
		}
	}

	// 2. 获取记录列表
	records, err := s.recordRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// 3. 获取总数
	total, err := s.recordRepo.Count(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// 4. 为每个记录设置表格schema（用于数据处理）
	tableSchemas := make(map[string]*table.Table)
	for _, rec := range records {
		if _, exists := tableSchemas[rec.TableID]; !exists {
			if schema, err := s.tableService.GetTable(ctx, rec.TableID); err == nil {
				tableSchemas[rec.TableID] = schema
			}
		}
		if schema, exists := tableSchemas[rec.TableID]; exists {
			rec.SetTableSchema(schema)
		}
	}

	return records, total, nil
}

// GetRecordHistory 获取记录版本历史
func (s *RecordService) GetRecordHistory(ctx context.Context, recordID string, userID string) ([]*RecordVersion, error) {
	// 1. 获取记录以检查权限
	rec, err := s.recordRepo.GetByID(ctx, recordID)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, errors.ErrNotFound.WithDetails("记录未找到")
	}

	// 2. 权限检查
	if err := s.checkPermission(ctx, userID, rec.TableID, "record:read"); err != nil {
		return nil, err
	}

	// 3. 获取版本历史
	return s.versionManager.GetRecordHistory(ctx, recordID)
}

// RestoreRecordVersion 恢复记录到指定版本
func (s *RecordService) RestoreRecordVersion(ctx context.Context, recordID string, versionID string, userID string) (*record.Record, error) {
	// 1. 获取记录以检查权限
	rec, err := s.recordRepo.GetByID(ctx, recordID)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, errors.ErrNotFound.WithDetails("记录未找到")
	}

	// 2. 权限检查
	if err := s.checkPermission(ctx, userID, rec.TableID, "record:update"); err != nil {
		return nil, err
	}

	// 3. 恢复版本
	restoredRecord, err := s.versionManager.RestoreVersion(ctx, recordID, versionID, userID)
	if err != nil {
		return nil, err
	}

	// 4. 记录变更事件
	changeEvent := restoredRecord.CreateChangeEvent("restore", rec.Data, userID)
	if err := s.changeTracker.TrackChange(ctx, changeEvent); err != nil {
		logger.Error("记录变更追踪失败", logger.ErrorField(err))
	}

	logger.Info("记录版本恢复成功",
		logger.String("record_id", recordID),
		logger.String("version_id", versionID),
		logger.String("user_id", userID),
	)

	return restoredRecord, nil
}

// GetRecordChanges 获取记录变更历史
func (s *RecordService) GetRecordChanges(ctx context.Context, recordID string, userID string) ([]*record.RecordChangeEvent, error) {
	// 1. 获取记录以检查权限
	rec, err := s.recordRepo.GetByID(ctx, recordID)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, errors.ErrNotFound.WithDetails("记录未找到")
	}

	// 2. 权限检查
	if err := s.checkPermission(ctx, userID, rec.TableID, "record:read"); err != nil {
		return nil, err
	}

	// 3. 获取变更历史
	return s.changeTracker.GetRecordChanges(ctx, recordID)
}

// ValidateRecordData 验证记录数据 - 支持动态schema
func (s *RecordService) ValidateRecordData(ctx context.Context, tableID string, data map[string]interface{}) error {
	// 1. 获取表格schema
	tableSchema, err := s.tableService.GetTable(ctx, tableID)
	if err != nil {
		return fmt.Errorf("获取表格schema失败: %v", err)
	}

	// 2. 创建临时记录进行验证
	tempRecord := &record.Record{
		TableID: tableID,
		Data:    data,
	}
	tempRecord.SetTableSchema(tableSchema)

	// 3. 执行验证
	return s.validator.ValidateData(ctx, tempRecord, tableSchema)
}

// checkPermission 检查用户权限
func (s *RecordService) checkPermission(ctx context.Context, userID, tableID, action string) error {
	hasPermission, err := s.permissionSvc.CheckPermission(ctx, userID, "table", tableID, permission.Action(action))
	if err != nil {
		return fmt.Errorf("权限检查失败: %v", err)
	}
	if !hasPermission {
		return errors.ErrForbidden.WithDetails(fmt.Sprintf("用户 %s 没有权限执行操作 %s", userID, action))
	}
	return nil
}

// RecordVersion 记录版本信息
type RecordVersion struct {
	ID         string                 `json:"id"`
	RecordID   string                 `json:"record_id"`
	Version    int64                  `json:"version"`
	Data       map[string]interface{} `json:"data"`
	ChangeType string                 `json:"change_type"`
	ChangedBy  string                 `json:"changed_by"`
	ChangedAt  time.Time              `json:"changed_at"`
	Comment    string                 `json:"comment,omitempty"`
}
