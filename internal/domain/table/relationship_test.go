package table

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// RelationshipManagerTestSuite 关系管理器测试套件
type RelationshipManagerTestSuite struct {
	suite.Suite
	manager *RelationshipManager
}

// SetupTest 设置测试
func (s *RelationshipManagerTestSuite) SetupTest() {
	s.manager = NewRelationshipManager()
}

// TestCreateRelationship 测试创建关系
func (s *RelationshipManagerTestSuite) TestCreateRelationship() {
	config := &RelationshipConfig{
		ID:                "rel-1",
		SourceTableID:     "table-1",
		SourceFieldID:     "field-1",
		TargetTableID:     "table-2",
		TargetFieldID:     "field-2",
		RelationType:      RelationTypeOneToMany,
		DisplayField:      "name",
		AllowLinkToMultiple: true,
		IsSymmetric:       false,
		CascadeDelete:     false,
		OnDeleteAction:    "restrict",
		OnUpdateAction:    "cascade",
		CreatedBy:         "user-1",
		CreatedTime:       time.Now(),
	}

	err := s.manager.CreateRelationship(config)
	assert.NoError(s.T(), err)

	// 验证关系已创建
	retrieved, err := s.manager.GetRelationship("rel-1")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), config.ID, retrieved.ID)
	assert.Equal(s.T(), config.RelationType, retrieved.RelationType)
}

// TestCreateSymmetricRelationship 测试创建对称关系
func (s *RelationshipManagerTestSuite) TestCreateSymmetricRelationship() {
	config := &RelationshipConfig{
		ID:                "rel-1",
		SourceTableID:     "table-1",
		SourceFieldID:     "field-1",
		TargetTableID:     "table-2",
		TargetFieldID:     "field-2",
		RelationType:      RelationTypeManyToMany,
		IsSymmetric:       true,
		CreatedBy:         "user-1",
		CreatedTime:       time.Now(),
	}

	err := s.manager.CreateRelationship(config)
	assert.NoError(s.T(), err)

	// 验证反向关系已创建
	reverseID := "reverse_rel-1"
	reverse, err := s.manager.GetRelationship(reverseID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), config.TargetTableID, reverse.SourceTableID)
	assert.Equal(s.T(), config.SourceTableID, reverse.TargetTableID)
}

// TestValidateRelationshipConfig 测试验证关系配置
func (s *RelationshipManagerTestSuite) TestValidateRelationshipConfig() {
	// 测试有效配置
	validConfig := &RelationshipConfig{
		ID:            "rel-1",
		SourceTableID: "table-1",
		SourceFieldID: "field-1",
		TargetTableID: "table-2",
		RelationType:  RelationTypeOneToOne,
		CreatedBy:     "user-1",
		CreatedTime:   time.Now(),
	}

	err := s.manager.validateRelationshipConfig(validConfig)
	assert.NoError(s.T(), err)

	// 测试无效配置 - 缺少源表ID
	invalidConfig := &RelationshipConfig{
		ID:           "rel-1",
		TargetTableID: "table-2",
		RelationType: RelationTypeOneToOne,
	}

	err = s.manager.validateRelationshipConfig(invalidConfig)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "源表ID不能为空")

	// 测试无效关系类型
	invalidTypeConfig := &RelationshipConfig{
		ID:            "rel-1",
		SourceTableID: "table-1",
		SourceFieldID: "field-1",
		TargetTableID: "table-2",
		RelationType:  RelationType("invalid"),
	}

	err = s.manager.validateRelationshipConfig(invalidTypeConfig)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "无效的关系类型")
}

// TestCheckRelationshipConflicts 测试检查关系冲突
func (s *RelationshipManagerTestSuite) TestCheckRelationshipConflicts() {
	// 创建第一个关系
	config1 := &RelationshipConfig{
		ID:            "rel-1",
		SourceTableID: "table-1",
		SourceFieldID: "field-1",
		TargetTableID: "table-2",
		RelationType:  RelationTypeOneToOne,
		CreatedBy:     "user-1",
		CreatedTime:   time.Now(),
	}

	err := s.manager.CreateRelationship(config1)
	assert.NoError(s.T(), err)

	// 尝试创建冲突的关系
	config2 := &RelationshipConfig{
		ID:            "rel-2",
		SourceTableID: "table-1",
		SourceFieldID: "field-1",
		TargetTableID: "table-2",
		RelationType:  RelationTypeOneToOne,
		CreatedBy:     "user-1",
		CreatedTime:   time.Now(),
	}

	err = s.manager.CreateRelationship(config2)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "已存在相同的关系")
}

// TestUpdateRelationship 测试更新关系
func (s *RelationshipManagerTestSuite) TestUpdateRelationship() {
	// 创建关系
	config := &RelationshipConfig{
		ID:            "rel-1",
		SourceTableID: "table-1",
		SourceFieldID: "field-1",
		TargetTableID: "table-2",
		RelationType:  RelationTypeOneToMany,
		DisplayField:  "name",
		CreatedBy:     "user-1",
		CreatedTime:   time.Now(),
	}

	err := s.manager.CreateRelationship(config)
	assert.NoError(s.T(), err)

	// 更新关系
	updates := &RelationshipConfig{
		ID:            "rel-1",
		SourceTableID: "table-1",
		SourceFieldID: "field-1",
		TargetTableID: "table-2",
		RelationType:  RelationTypeOneToMany,
		DisplayField:  "title", // 更改显示字段
		CreatedBy:     "user-1",
		CreatedTime:   config.CreatedTime,
	}

	err = s.manager.UpdateRelationship("rel-1", updates)
	assert.NoError(s.T(), err)

	// 验证更新
	updated, err := s.manager.GetRelationship("rel-1")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "title", updated.DisplayField)
	assert.NotNil(s.T(), updated.LastModifiedTime)
}

// TestDeleteRelationship 测试删除关系
func (s *RelationshipManagerTestSuite) TestDeleteRelationship() {
	// 创建关系
	config := &RelationshipConfig{
		ID:            "rel-1",
		SourceTableID: "table-1",
		SourceFieldID: "field-1",
		TargetTableID: "table-2",
		RelationType:  RelationTypeOneToMany,
		CreatedBy:     "user-1",
		CreatedTime:   time.Now(),
	}

	err := s.manager.CreateRelationship(config)
	assert.NoError(s.T(), err)

	// 删除关系
	err = s.manager.DeleteRelationship("rel-1")
	assert.NoError(s.T(), err)

	// 验证关系已删除
	_, err = s.manager.GetRelationship("rel-1")
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "关系不存在")
}

// TestGetRelationshipsByTable 测试按表获取关系
func (s *RelationshipManagerTestSuite) TestGetRelationshipsByTable() {
	// 创建多个关系
	configs := []*RelationshipConfig{
		{
			ID:            "rel-1",
			SourceTableID: "table-1",
			SourceFieldID: "field-1",
			TargetTableID: "table-2",
			RelationType:  RelationTypeOneToMany,
			CreatedBy:     "user-1",
			CreatedTime:   time.Now(),
		},
		{
			ID:            "rel-2",
			SourceTableID: "table-2",
			SourceFieldID: "field-2",
			TargetTableID: "table-3",
			RelationType:  RelationTypeManyToOne,
			CreatedBy:     "user-1",
			CreatedTime:   time.Now(),
		},
		{
			ID:            "rel-3",
			SourceTableID: "table-3",
			SourceFieldID: "field-3",
			TargetTableID: "table-1",
			RelationType:  RelationTypeOneToOne,
			CreatedBy:     "user-1",
			CreatedTime:   time.Now(),
		},
	}

	for _, config := range configs {
		err := s.manager.CreateRelationship(config)
		assert.NoError(s.T(), err)
	}

	// 获取table-1的关系
	relationships := s.manager.GetRelationshipsByTable("table-1")
	assert.Len(s.T(), relationships, 2) // rel-1 (作为源表) 和 rel-3 (作为目标表)

	// 获取table-2的关系
	relationships = s.manager.GetRelationshipsByTable("table-2")
	assert.Len(s.T(), relationships, 2) // rel-1 (作为目标表) 和 rel-2 (作为源表)
}

// TestAnalyzeRelationshipImpact 测试分析关系影响
func (s *RelationshipManagerTestSuite) TestAnalyzeRelationshipImpact() {
	oldConfig := &RelationshipConfig{
		ID:            "rel-1",
		SourceTableID: "table-1",
		TargetTableID: "table-2",
		RelationType:  RelationTypeOneToMany,
		CascadeDelete: false,
	}

	newConfig := &RelationshipConfig{
		ID:            "rel-1",
		SourceTableID: "table-1",
		TargetTableID: "table-3", // 更改目标表
		RelationType:  RelationTypeManyToMany, // 更改关系类型
		CascadeDelete: true, // 启用级联删除
	}

	impact := s.manager.analyzeRelationshipImpact(oldConfig, newConfig)

	assert.NotEmpty(s.T(), impact.BreakingChanges)
	assert.Contains(s.T(), impact.BreakingChanges[0], "关系类型从")
	assert.Contains(s.T(), impact.BreakingChanges, "目标表变更")
	assert.NotEmpty(s.T(), impact.Warnings)
	assert.Contains(s.T(), impact.Warnings[0], "启用级联删除")
	assert.NotEmpty(s.T(), impact.RequiredMigrations)
}

// TestValidateRelationshipIntegrity 测试验证关系完整性
func (s *RelationshipManagerTestSuite) TestValidateRelationshipIntegrity() {
	// 创建关系
	config := &RelationshipConfig{
		ID:            "rel-1",
		SourceTableID: "table-1",
		SourceFieldID: "field-1",
		TargetTableID: "table-2",
		RelationType:  RelationTypeOneToMany,
		CreatedBy:     "user-1",
		CreatedTime:   time.Now(),
	}

	err := s.manager.CreateRelationship(config)
	assert.NoError(s.T(), err)

	// 测试有效的记录数据
	validData := map[string]interface{}{
		"field-1": "target-record-1",
		"other_field": "some_value",
	}

	err = s.manager.ValidateRelationshipIntegrity("table-1", validData)
	assert.NoError(s.T(), err) // 由于是模拟实现，应该返回nil

	// 测试空值数据
	emptyData := map[string]interface{}{
		"field-1": nil,
		"other_field": "some_value",
	}

	err = s.manager.ValidateRelationshipIntegrity("table-1", emptyData)
	assert.NoError(s.T(), err) // 空值应该被允许
}

// 运行测试套件
func TestRelationshipManagerSuite(t *testing.T) {
	suite.Run(t, new(RelationshipManagerTestSuite))
}

// TestRelationTypeValidation 测试关系类型验证
func TestRelationTypeValidation(t *testing.T) {
	validTypes := []RelationType{
		RelationTypeOneToOne,
		RelationTypeOneToMany,
		RelationTypeManyToOne,
		RelationTypeManyToMany,
	}

	for _, relType := range validTypes {
		config := &RelationshipConfig{
			ID:            "test-rel",
			SourceTableID: "table-1",
			SourceFieldID: "field-1",
			TargetTableID: "table-2",
			RelationType:  relType,
			CreatedBy:     "user-1",
			CreatedTime:   time.Now(),
		}

		manager := NewRelationshipManager()
		err := manager.validateRelationshipConfig(config)
		assert.NoError(t, err, "Valid relation type %s should not cause error", relType)
	}

	// 测试无效类型
	invalidConfig := &RelationshipConfig{
		ID:            "test-rel",
		SourceTableID: "table-1",
		SourceFieldID: "field-1",
		TargetTableID: "table-2",
		RelationType:  RelationType("invalid_type"),
		CreatedBy:     "user-1",
		CreatedTime:   time.Now(),
	}

	manager := NewRelationshipManager()
	err := manager.validateRelationshipConfig(invalidConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无效的关系类型")
}

// TestReverseRelationshipGeneration 测试反向关系生成
func TestReverseRelationshipGeneration(t *testing.T) {
	manager := NewRelationshipManager()

	testCases := []struct {
		original RelationType
		expected RelationType
	}{
		{RelationTypeOneToOne, RelationTypeOneToOne},
		{RelationTypeOneToMany, RelationTypeManyToOne},
		{RelationTypeManyToOne, RelationTypeOneToMany},
		{RelationTypeManyToMany, RelationTypeManyToMany},
	}

	for _, tc := range testCases {
		result := manager.getReverseRelationType(tc.original)
		assert.Equal(t, tc.expected, result, 
			"Reverse of %s should be %s, got %s", tc.original, tc.expected, result)
	}
}

// BenchmarkCreateRelationship 基准测试创建关系
func BenchmarkCreateRelationship(b *testing.B) {
	manager := NewRelationshipManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config := &RelationshipConfig{
			ID:            fmt.Sprintf("rel-%d", i),
			SourceTableID: "table-1",
			SourceFieldID: "field-1",
			TargetTableID: "table-2",
			RelationType:  RelationTypeOneToMany,
			CreatedBy:     "user-1",
			CreatedTime:   time.Now(),
		}

		manager.CreateRelationship(config)
	}
}

// BenchmarkGetRelationshipsByTable 基准测试按表获取关系
func BenchmarkGetRelationshipsByTable(b *testing.B) {
	manager := NewRelationshipManager()

	// 预先创建一些关系
	for i := 0; i < 1000; i++ {
		config := &RelationshipConfig{
			ID:            fmt.Sprintf("rel-%d", i),
			SourceTableID: fmt.Sprintf("table-%d", i%10),
			SourceFieldID: "field-1",
			TargetTableID: "table-target",
			RelationType:  RelationTypeOneToMany,
			CreatedBy:     "user-1",
			CreatedTime:   time.Now(),
		}
		manager.CreateRelationship(config)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.GetRelationshipsByTable("table-1")
	}
}

import "fmt"