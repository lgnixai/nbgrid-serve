package table

import (
	"fmt"
	"time"
)

// RelationType 关系类型枚举
type RelationType string

const (
	RelationTypeOneToOne   RelationType = "one_to_one"
	RelationTypeOneToMany  RelationType = "one_to_many"
	RelationTypeManyToOne  RelationType = "many_to_one"
	RelationTypeManyToMany RelationType = "many_to_many"
)

// RelationshipConfig 关系配置
type RelationshipConfig struct {
	ID                    string       `json:"id"`
	SourceTableID         string       `json:"source_table_id"`
	SourceFieldID         string       `json:"source_field_id"`
	TargetTableID         string       `json:"target_table_id"`
	TargetFieldID         string       `json:"target_field_id"`
	RelationType          RelationType `json:"relation_type"`
	DisplayField          string       `json:"display_field"`          // 显示字段名称
	AllowLinkToMultiple   bool         `json:"allow_link_to_multiple"` // 是否允许链接到多个记录
	IsSymmetric           bool         `json:"is_symmetric"`           // 是否为对称关系
	CascadeDelete         bool         `json:"cascade_delete"`         // 是否级联删除
	OnDeleteAction        string       `json:"on_delete_action"`       // 删除时的动作: cascade, set_null, restrict
	OnUpdateAction        string       `json:"on_update_action"`       // 更新时的动作: cascade, set_null, restrict
	CreatedBy             string       `json:"created_by"`
	CreatedTime           time.Time    `json:"created_at"`
	LastModifiedTime      *time.Time   `json:"updated_at"`
}

// LinkFieldOptions 链接字段选项
type LinkFieldOptions struct {
	TargetTableID       string       `json:"target_table_id"`
	TargetFieldID       string       `json:"target_field_id"`
	RelationType        RelationType `json:"relation_type"`
	DisplayField        string       `json:"display_field"`
	AllowLinkToMultiple bool         `json:"allow_link_to_multiple"`
	IsSymmetric         bool         `json:"is_symmetric"`
	CascadeDelete       bool         `json:"cascade_delete"`
	OnDeleteAction      string       `json:"on_delete_action"`
	OnUpdateAction      string       `json:"on_update_action"`
}

// RelationshipConstraint 关系约束
type RelationshipConstraint struct {
	Type        string      `json:"type"`        // unique, foreign_key, check
	Expression  string      `json:"expression"`  // 约束表达式
	ErrorMessage string     `json:"error_message"`
	IsActive    bool        `json:"is_active"`
}

// RelationshipImpactAnalysis 关系影响分析
type RelationshipImpactAnalysis struct {
	AffectedTables   []string `json:"affected_tables"`
	AffectedRecords  int64    `json:"affected_records"`
	AffectedFields   []string `json:"affected_fields"`
	BreakingChanges  []string `json:"breaking_changes"`
	Warnings         []string `json:"warnings"`
	RequiredMigrations []string `json:"required_migrations"`
}

// RelationshipManager 关系管理器
type RelationshipManager struct {
	relationships map[string]*RelationshipConfig
	constraints   map[string][]RelationshipConstraint
}

// NewRelationshipManager 创建关系管理器
func NewRelationshipManager() *RelationshipManager {
	return &RelationshipManager{
		relationships: make(map[string]*RelationshipConfig),
		constraints:   make(map[string][]RelationshipConstraint),
	}
}

// CreateRelationship 创建关系
func (rm *RelationshipManager) CreateRelationship(config *RelationshipConfig) error {
	// 验证关系配置
	if err := rm.validateRelationshipConfig(config); err != nil {
		return fmt.Errorf("关系配置验证失败: %v", err)
	}

	// 检查关系冲突
	if err := rm.checkRelationshipConflicts(config); err != nil {
		return fmt.Errorf("关系冲突检查失败: %v", err)
	}

	// 存储关系配置
	rm.relationships[config.ID] = config

	// 如果是对称关系，创建反向关系
	if config.IsSymmetric {
		reverseConfig := rm.createReverseRelationship(config)
		rm.relationships[reverseConfig.ID] = reverseConfig
	}

	return nil
}

// UpdateRelationship 更新关系
func (rm *RelationshipManager) UpdateRelationship(relationshipID string, updates *RelationshipConfig) error {
	existing, exists := rm.relationships[relationshipID]
	if !exists {
		return fmt.Errorf("关系不存在: %s", relationshipID)
	}

	// 分析影响
	impact := rm.analyzeRelationshipImpact(existing, updates)
	if len(impact.BreakingChanges) > 0 {
		return fmt.Errorf("更新会导致破坏性变更: %v", impact.BreakingChanges)
	}

	// 验证更新后的配置
	if err := rm.validateRelationshipConfig(updates); err != nil {
		return fmt.Errorf("更新后的关系配置验证失败: %v", err)
	}

	// 更新关系
	now := time.Now()
	updates.LastModifiedTime = &now
	rm.relationships[relationshipID] = updates

	return nil
}

// DeleteRelationship 删除关系
func (rm *RelationshipManager) DeleteRelationship(relationshipID string) error {
	config, exists := rm.relationships[relationshipID]
	if !exists {
		return fmt.Errorf("关系不存在: %s", relationshipID)
	}

	// 分析删除影响
	impact := rm.analyzeRelationshipDeletionImpact(config)
	if len(impact.BreakingChanges) > 0 {
		return fmt.Errorf("删除会导致破坏性变更: %v", impact.BreakingChanges)
	}

	// 删除关系
	delete(rm.relationships, relationshipID)

	// 如果是对称关系，删除反向关系
	if config.IsSymmetric {
		reverseID := rm.generateReverseRelationshipID(config)
		delete(rm.relationships, reverseID)
	}

	return nil
}

// GetRelationship 获取关系
func (rm *RelationshipManager) GetRelationship(relationshipID string) (*RelationshipConfig, error) {
	config, exists := rm.relationships[relationshipID]
	if !exists {
		return nil, fmt.Errorf("关系不存在: %s", relationshipID)
	}
	return config, nil
}

// GetRelationshipsByTable 获取表的所有关系
func (rm *RelationshipManager) GetRelationshipsByTable(tableID string) []*RelationshipConfig {
	var relationships []*RelationshipConfig
	for _, config := range rm.relationships {
		if config.SourceTableID == tableID || config.TargetTableID == tableID {
			relationships = append(relationships, config)
		}
	}
	return relationships
}

// GetRelationshipsByField 获取字段的关系
func (rm *RelationshipManager) GetRelationshipsByField(fieldID string) []*RelationshipConfig {
	var relationships []*RelationshipConfig
	for _, config := range rm.relationships {
		if config.SourceFieldID == fieldID || config.TargetFieldID == fieldID {
			relationships = append(relationships, config)
		}
	}
	return relationships
}

// ValidateRelationshipIntegrity 验证关系完整性
func (rm *RelationshipManager) ValidateRelationshipIntegrity(tableID string, recordData map[string]interface{}) error {
	relationships := rm.GetRelationshipsByTable(tableID)
	
	for _, rel := range relationships {
		if rel.SourceTableID == tableID {
			// 验证源表的关系完整性
			if err := rm.validateSourceRelationshipIntegrity(rel, recordData); err != nil {
				return err
			}
		}
	}
	
	return nil
}

// AnalyzeRelationshipImpact 分析关系变更影响
func (rm *RelationshipManager) AnalyzeRelationshipImpact(oldConfig, newConfig *RelationshipConfig) *RelationshipImpactAnalysis {
	return rm.analyzeRelationshipImpact(oldConfig, newConfig)
}

// validateRelationshipConfig 验证关系配置
func (rm *RelationshipManager) validateRelationshipConfig(config *RelationshipConfig) error {
	if config.SourceTableID == "" {
		return fmt.Errorf("源表ID不能为空")
	}
	if config.TargetTableID == "" {
		return fmt.Errorf("目标表ID不能为空")
	}
	if config.SourceFieldID == "" {
		return fmt.Errorf("源字段ID不能为空")
	}
	if config.RelationType == "" {
		return fmt.Errorf("关系类型不能为空")
	}

	// 验证关系类型
	validTypes := []RelationType{
		RelationTypeOneToOne, RelationTypeOneToMany,
		RelationTypeManyToOne, RelationTypeManyToMany,
	}
	isValidType := false
	for _, validType := range validTypes {
		if config.RelationType == validType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return fmt.Errorf("无效的关系类型: %s", config.RelationType)
	}

	// 验证删除和更新动作
	validActions := []string{"cascade", "set_null", "restrict"}
	if config.OnDeleteAction != "" {
		isValid := false
		for _, action := range validActions {
			if config.OnDeleteAction == action {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("无效的删除动作: %s", config.OnDeleteAction)
		}
	}

	if config.OnUpdateAction != "" {
		isValid := false
		for _, action := range validActions {
			if config.OnUpdateAction == action {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("无效的更新动作: %s", config.OnUpdateAction)
		}
	}

	return nil
}

// checkRelationshipConflicts 检查关系冲突
func (rm *RelationshipManager) checkRelationshipConflicts(config *RelationshipConfig) error {
	for _, existing := range rm.relationships {
		// 检查是否已存在相同的关系
		if existing.SourceTableID == config.SourceTableID &&
			existing.TargetTableID == config.TargetTableID &&
			existing.SourceFieldID == config.SourceFieldID {
			return fmt.Errorf("已存在相同的关系")
		}

		// 检查一对一关系的唯一性
		if config.RelationType == RelationTypeOneToOne {
			if existing.TargetTableID == config.TargetTableID &&
				existing.RelationType == RelationTypeOneToOne {
				return fmt.Errorf("目标表已存在一对一关系")
			}
		}
	}

	return nil
}

// createReverseRelationship 创建反向关系
func (rm *RelationshipManager) createReverseRelationship(config *RelationshipConfig) *RelationshipConfig {
	reverseType := rm.getReverseRelationType(config.RelationType)
	
	return &RelationshipConfig{
		ID:                  rm.generateReverseRelationshipID(config),
		SourceTableID:       config.TargetTableID,
		SourceFieldID:       config.TargetFieldID,
		TargetTableID:       config.SourceTableID,
		TargetFieldID:       config.SourceFieldID,
		RelationType:        reverseType,
		DisplayField:        config.DisplayField,
		AllowLinkToMultiple: config.AllowLinkToMultiple,
		IsSymmetric:         true,
		CascadeDelete:       config.CascadeDelete,
		OnDeleteAction:      config.OnDeleteAction,
		OnUpdateAction:      config.OnUpdateAction,
		CreatedBy:           config.CreatedBy,
		CreatedTime:         config.CreatedTime,
		LastModifiedTime:    config.LastModifiedTime,
	}
}

// getReverseRelationType 获取反向关系类型
func (rm *RelationshipManager) getReverseRelationType(relType RelationType) RelationType {
	switch relType {
	case RelationTypeOneToOne:
		return RelationTypeOneToOne
	case RelationTypeOneToMany:
		return RelationTypeManyToOne
	case RelationTypeManyToOne:
		return RelationTypeOneToMany
	case RelationTypeManyToMany:
		return RelationTypeManyToMany
	default:
		return relType
	}
}

// generateReverseRelationshipID 生成反向关系ID
func (rm *RelationshipManager) generateReverseRelationshipID(config *RelationshipConfig) string {
	return fmt.Sprintf("reverse_%s", config.ID)
}

// analyzeRelationshipImpact 分析关系影响
func (rm *RelationshipManager) analyzeRelationshipImpact(oldConfig, newConfig *RelationshipConfig) *RelationshipImpactAnalysis {
	impact := &RelationshipImpactAnalysis{
		AffectedTables:     []string{},
		AffectedRecords:    0,
		AffectedFields:     []string{},
		BreakingChanges:    []string{},
		Warnings:           []string{},
		RequiredMigrations: []string{},
	}

	// 检查关系类型变更
	if oldConfig.RelationType != newConfig.RelationType {
		impact.BreakingChanges = append(impact.BreakingChanges, 
			fmt.Sprintf("关系类型从 %s 变更为 %s", oldConfig.RelationType, newConfig.RelationType))
		impact.RequiredMigrations = append(impact.RequiredMigrations, "重建关系索引")
	}

	// 检查目标表变更
	if oldConfig.TargetTableID != newConfig.TargetTableID {
		impact.BreakingChanges = append(impact.BreakingChanges, "目标表变更")
		impact.AffectedTables = append(impact.AffectedTables, oldConfig.TargetTableID, newConfig.TargetTableID)
		impact.RequiredMigrations = append(impact.RequiredMigrations, "迁移关联数据")
	}

	// 检查级联删除变更
	if oldConfig.CascadeDelete != newConfig.CascadeDelete {
		if newConfig.CascadeDelete {
			impact.Warnings = append(impact.Warnings, "启用级联删除可能导致数据丢失")
		} else {
			impact.Warnings = append(impact.Warnings, "禁用级联删除可能导致孤立数据")
		}
	}

	return impact
}

// analyzeRelationshipDeletionImpact 分析关系删除影响
func (rm *RelationshipManager) analyzeRelationshipDeletionImpact(config *RelationshipConfig) *RelationshipImpactAnalysis {
	impact := &RelationshipImpactAnalysis{
		AffectedTables:     []string{config.SourceTableID, config.TargetTableID},
		AffectedRecords:    0, // 需要从数据库查询
		AffectedFields:     []string{config.SourceFieldID, config.TargetFieldID},
		BreakingChanges:    []string{"删除关系将断开表间连接"},
		Warnings:           []string{"删除关系后，相关的查找字段和汇总字段将失效"},
		RequiredMigrations: []string{"清理关联数据", "更新相关视图"},
	}

	return impact
}

// validateSourceRelationshipIntegrity 验证源关系完整性
func (rm *RelationshipManager) validateSourceRelationshipIntegrity(rel *RelationshipConfig, recordData map[string]interface{}) error {
	// 获取关联字段的值
	linkValue, exists := recordData[rel.SourceFieldID]
	if !exists || linkValue == nil {
		return nil // 空值不需要验证
	}

	// 根据关系类型验证
	switch rel.RelationType {
	case RelationTypeOneToOne, RelationTypeManyToOne:
		// 单值关系，验证目标记录是否存在
		if err := rm.validateSingleLinkValue(rel, linkValue); err != nil {
			return err
		}
	case RelationTypeOneToMany, RelationTypeManyToMany:
		// 多值关系，验证所有目标记录是否存在
		if err := rm.validateMultipleLinkValues(rel, linkValue); err != nil {
			return err
		}
	}

	return nil
}

// validateSingleLinkValue 验证单个链接值
func (rm *RelationshipManager) validateSingleLinkValue(rel *RelationshipConfig, value interface{}) error {
	// 这里需要查询目标表验证记录是否存在
	// 实际实现需要依赖记录仓储
	return nil
}

// validateMultipleLinkValues 验证多个链接值
func (rm *RelationshipManager) validateMultipleLinkValues(rel *RelationshipConfig, value interface{}) error {
	// 这里需要查询目标表验证所有记录是否存在
	// 实际实现需要依赖记录仓储
	return nil
}