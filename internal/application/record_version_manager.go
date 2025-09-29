package application

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"teable-go-backend/internal/domain/record"
	"teable-go-backend/pkg/logger"
	"teable-go-backend/pkg/utils"
)

// RecordVersionManager 记录版本管理器
type RecordVersionManager struct {
	recordRepo  record.Repository
	versionRepo VersionRepository
}

// NewRecordVersionManager 创建记录版本管理器
func NewRecordVersionManager(recordRepo record.Repository) *RecordVersionManager {
	return &RecordVersionManager{
		recordRepo: recordRepo,
		// TODO: 注入实际的版本仓储实现
		versionRepo: &InMemoryVersionRepository{
			versions: make(map[string][]*RecordVersion),
		},
	}
}

// CreateVersion 创建记录版本
func (m *RecordVersionManager) CreateVersion(ctx context.Context, rec *record.Record, changeType string, changedBy string) error {
	version := &RecordVersion{
		ID:         utils.GenerateID(),
		RecordID:   rec.ID,
		Version:    rec.Version,
		Data:       m.copyData(rec.Data),
		ChangeType: changeType,
		ChangedBy:  changedBy,
		ChangedAt:  time.Now(),
	}

	if err := m.versionRepo.SaveVersion(ctx, version); err != nil {
		return fmt.Errorf("保存记录版本失败: %v", err)
	}

	logger.Info("记录版本创建成功",
		logger.String("version_id", version.ID),
		logger.String("record_id", rec.ID),
		logger.Int64("version", rec.Version),
		logger.String("change_type", changeType),
	)

	return nil
}

// GetRecordHistory 获取记录版本历史
func (m *RecordVersionManager) GetRecordHistory(ctx context.Context, recordID string) ([]*RecordVersion, error) {
	versions, err := m.versionRepo.GetRecordVersions(ctx, recordID)
	if err != nil {
		return nil, fmt.Errorf("获取记录版本历史失败: %v", err)
	}

	// 按版本号倒序排序
	for i := 0; i < len(versions)-1; i++ {
		for j := i + 1; j < len(versions); j++ {
			if versions[i].Version < versions[j].Version {
				versions[i], versions[j] = versions[j], versions[i]
			}
		}
	}

	return versions, nil
}

// GetVersion 获取指定版本
func (m *RecordVersionManager) GetVersion(ctx context.Context, versionID string) (*RecordVersion, error) {
	return m.versionRepo.GetVersion(ctx, versionID)
}

// RestoreVersion 恢复记录到指定版本
func (m *RecordVersionManager) RestoreVersion(ctx context.Context, recordID string, versionID string, userID string) (*record.Record, error) {
	// 1. 获取目标版本
	targetVersion, err := m.versionRepo.GetVersion(ctx, versionID)
	if err != nil {
		return nil, fmt.Errorf("获取目标版本失败: %v", err)
	}
	if targetVersion == nil {
		return nil, fmt.Errorf("版本 %s 不存在", versionID)
	}
	if targetVersion.RecordID != recordID {
		return nil, fmt.Errorf("版本 %s 不属于记录 %s", versionID, recordID)
	}

	// 2. 获取当前记录
	currentRecord, err := m.recordRepo.GetByID(ctx, recordID)
	if err != nil {
		return nil, fmt.Errorf("获取当前记录失败: %v", err)
	}
	if currentRecord == nil {
		return nil, fmt.Errorf("记录 %s 不存在", recordID)
	}

	// 3. 创建当前版本的备份
	if err := m.CreateVersion(ctx, currentRecord, "backup_before_restore", userID); err != nil {
		logger.Error("创建恢复前备份失败", logger.ErrorField(err))
	}

	// 4. 恢复数据
	currentRecord.Data = m.copyData(targetVersion.Data)
	currentRecord.Version++
	now := time.Now()
	currentRecord.LastModifiedTime = &now
	currentRecord.UpdatedBy = &userID

	// 5. 保存恢复后的记录
	if err := m.recordRepo.Update(ctx, currentRecord); err != nil {
		return nil, fmt.Errorf("保存恢复后的记录失败: %v", err)
	}

	// 6. 创建恢复版本
	if err := m.CreateVersion(ctx, currentRecord, "restore", userID); err != nil {
		logger.Error("创建恢复版本失败", logger.ErrorField(err))
	}

	logger.Info("记录版本恢复成功",
		logger.String("record_id", recordID),
		logger.String("version_id", versionID),
		logger.String("user_id", userID),
		logger.Int64("new_version", currentRecord.Version),
	)

	return currentRecord, nil
}

// CompareVersions 比较两个版本的差异
func (m *RecordVersionManager) CompareVersions(ctx context.Context, versionID1, versionID2 string) (*VersionComparison, error) {
	// 获取两个版本
	version1, err := m.versionRepo.GetVersion(ctx, versionID1)
	if err != nil {
		return nil, fmt.Errorf("获取版本1失败: %v", err)
	}
	if version1 == nil {
		return nil, fmt.Errorf("版本 %s 不存在", versionID1)
	}

	version2, err := m.versionRepo.GetVersion(ctx, versionID2)
	if err != nil {
		return nil, fmt.Errorf("获取版本2失败: %v", err)
	}
	if version2 == nil {
		return nil, fmt.Errorf("版本 %s 不存在", versionID2)
	}

	// 比较数据差异
	comparison := &VersionComparison{
		Version1:    version1,
		Version2:    version2,
		Differences: m.calculateDifferences(version1.Data, version2.Data),
	}

	return comparison, nil
}

// DeleteOldVersions 删除旧版本（保留策略）
func (m *RecordVersionManager) DeleteOldVersions(ctx context.Context, recordID string, keepCount int) error {
	versions, err := m.versionRepo.GetRecordVersions(ctx, recordID)
	if err != nil {
		return fmt.Errorf("获取记录版本失败: %v", err)
	}

	if len(versions) <= keepCount {
		return nil // 不需要删除
	}

	// 按版本号排序，保留最新的版本
	for i := 0; i < len(versions)-1; i++ {
		for j := i + 1; j < len(versions); j++ {
			if versions[i].Version < versions[j].Version {
				versions[i], versions[j] = versions[j], versions[i]
			}
		}
	}

	// 删除多余的旧版本
	toDelete := versions[keepCount:]
	for _, version := range toDelete {
		if err := m.versionRepo.DeleteVersion(ctx, version.ID); err != nil {
			logger.Error("删除旧版本失败",
				logger.String("version_id", version.ID),
				logger.ErrorField(err),
			)
		}
	}

	logger.Info("删除旧版本完成",
		logger.String("record_id", recordID),
		logger.Int("deleted_count", len(toDelete)),
		logger.Int("kept_count", keepCount),
	)

	return nil
}

// copyData 深拷贝数据
func (m *RecordVersionManager) copyData(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		return nil
	}

	// 使用JSON序列化/反序列化进行深拷贝
	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.Error("数据序列化失败", logger.ErrorField(err))
		return make(map[string]interface{})
	}

	var copied map[string]interface{}
	if err := json.Unmarshal(jsonData, &copied); err != nil {
		logger.Error("数据反序列化失败", logger.ErrorField(err))
		return make(map[string]interface{})
	}

	return copied
}

// calculateDifferences 计算数据差异
func (m *RecordVersionManager) calculateDifferences(data1, data2 map[string]interface{}) []FieldDifference {
	var differences []FieldDifference

	// 检查所有字段
	allFields := make(map[string]bool)
	for field := range data1 {
		allFields[field] = true
	}
	for field := range data2 {
		allFields[field] = true
	}

	for field := range allFields {
		value1, exists1 := data1[field]
		value2, exists2 := data2[field]

		if !exists1 && exists2 {
			// 字段在版本2中新增
			differences = append(differences, FieldDifference{
				Field:      field,
				ChangeType: "added",
				OldValue:   nil,
				NewValue:   value2,
			})
		} else if exists1 && !exists2 {
			// 字段在版本2中删除
			differences = append(differences, FieldDifference{
				Field:      field,
				ChangeType: "deleted",
				OldValue:   value1,
				NewValue:   nil,
			})
		} else if exists1 && exists2 {
			// 检查字段值是否变化
			if !m.isEqual(value1, value2) {
				differences = append(differences, FieldDifference{
					Field:      field,
					ChangeType: "modified",
					OldValue:   value1,
					NewValue:   value2,
				})
			}
		}
	}

	return differences
}

// isEqual 比较两个值是否相等
func (m *RecordVersionManager) isEqual(a, b interface{}) bool {
	aJSON, aErr := json.Marshal(a)
	bJSON, bErr := json.Marshal(b)

	if aErr != nil || bErr != nil {
		return false
	}

	return string(aJSON) == string(bJSON)
}

// VersionComparison 版本比较结果
type VersionComparison struct {
	Version1    *RecordVersion    `json:"version1"`
	Version2    *RecordVersion    `json:"version2"`
	Differences []FieldDifference `json:"differences"`
}

// FieldDifference 字段差异
type FieldDifference struct {
	Field      string      `json:"field"`
	ChangeType string      `json:"change_type"` // added, deleted, modified
	OldValue   interface{} `json:"old_value"`
	NewValue   interface{} `json:"new_value"`
}

// VersionRepository 版本仓储接口
type VersionRepository interface {
	SaveVersion(ctx context.Context, version *RecordVersion) error
	GetVersion(ctx context.Context, versionID string) (*RecordVersion, error)
	GetRecordVersions(ctx context.Context, recordID string) ([]*RecordVersion, error)
	DeleteVersion(ctx context.Context, versionID string) error
}

// InMemoryVersionRepository 内存版本仓储实现（用于测试和开发）
type InMemoryVersionRepository struct {
	versions map[string][]*RecordVersion
}

func (r *InMemoryVersionRepository) SaveVersion(ctx context.Context, version *RecordVersion) error {
	if r.versions[version.RecordID] == nil {
		r.versions[version.RecordID] = make([]*RecordVersion, 0)
	}
	r.versions[version.RecordID] = append(r.versions[version.RecordID], version)

	logger.Debug("版本已保存到内存",
		logger.String("version_id", version.ID),
		logger.String("record_id", version.RecordID),
	)

	return nil
}

func (r *InMemoryVersionRepository) GetVersion(ctx context.Context, versionID string) (*RecordVersion, error) {
	for _, versions := range r.versions {
		for _, version := range versions {
			if version.ID == versionID {
				return version, nil
			}
		}
	}
	return nil, nil
}

func (r *InMemoryVersionRepository) GetRecordVersions(ctx context.Context, recordID string) ([]*RecordVersion, error) {
	versions := r.versions[recordID]
	if versions == nil {
		return []*RecordVersion{}, nil
	}

	// 返回副本以避免外部修改
	result := make([]*RecordVersion, len(versions))
	copy(result, versions)
	return result, nil
}

func (r *InMemoryVersionRepository) DeleteVersion(ctx context.Context, versionID string) error {
	for recordID, versions := range r.versions {
		for i, version := range versions {
			if version.ID == versionID {
				// 从切片中删除该版本
				r.versions[recordID] = append(versions[:i], versions[i+1:]...)
				return nil
			}
		}
	}
	return fmt.Errorf("版本 %s 不存在", versionID)
}
