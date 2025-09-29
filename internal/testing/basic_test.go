package testing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"teable-go-backend/internal/domain/table"
	"teable-go-backend/internal/testing/builders"
)

// BasicTestSuite 基础测试套件
type BasicTestSuite struct {
	suite.Suite
}

// TestBasicAssertions 测试基本断言
func (s *BasicTestSuite) TestBasicAssertions() {
	s.T().Log("Testing basic assertions")

	// 测试基本断言
	assert.True(s.T(), true, "Basic assertion should pass")

	// 测试字符串
	expected := "hello"
	actual := "hello"
	assert.Equal(s.T(), expected, actual, "Strings should be equal")

	// 测试数字
	assert.Equal(s.T(), 42, 42, "Numbers should be equal")

	// 测试布尔值
	assert.True(s.T(), true, "True should be true")
	assert.False(s.T(), false, "False should be false")
}

// TestUserBuilder 测试用户构建器
func (s *BasicTestSuite) TestUserBuilder() {
	s.T().Log("Testing user builder")

	// 创建用户构建器
	builder := builders.NewUserBuilder()

	// 设置用户属性
	user := builder.
		WithName("Test User").
		WithEmail("test@example.com").
		WithPassword("Password123!").
		AsAdmin().
		Build()

	// 验证用户属性
	assert.NotNil(s.T(), user, "User should not be nil")
	assert.Equal(s.T(), "Test User", user.Name, "User name should match")
	assert.Equal(s.T(), "test@example.com", user.Email, "User email should match")
	assert.True(s.T(), user.IsAdmin, "User should be admin")
	assert.NoError(s.T(), user.CheckPassword("Password123!"), "Password should match")
}

// TestSpaceBuilder 测试空间构建器
func (s *BasicTestSuite) TestSpaceBuilder() {
	s.T().Log("Testing space builder")

	// 创建空间构建器
	builder := builders.NewSpaceBuilder()

	// 设置空间属性
	space := builder.
		WithName("Test Space").
		WithCreatedBy("user123").
		Build()

	// 验证空间属性
	assert.NotNil(s.T(), space, "Space should not be nil")
	assert.Equal(s.T(), "Test Space", space.Name, "Space name should match")
	assert.Equal(s.T(), "user123", space.CreatedBy, "Space creator should match")
}

// TestTableBuilder 测试表构建器
func (s *BasicTestSuite) TestTableBuilder() {
	s.T().Log("Testing table builder")

	// 创建表构建器
	builder := builders.NewTableBuilder()

	// 设置表属性
	table := builder.
		WithName("Test Table").
		WithBaseID("base123").
		WithCreatedBy("user123").
		Build()

	// 验证表属性
	assert.NotNil(s.T(), table, "Table should not be nil")
	assert.Equal(s.T(), "Test Table", table.Name, "Table name should match")
	assert.Equal(s.T(), "base123", table.BaseID, "Table base ID should match")
	assert.Equal(s.T(), "user123", table.CreatedBy, "Table creator should match")
}

// TestFieldBuilder 测试字段构建器
func (s *BasicTestSuite) TestFieldBuilder() {
	s.T().Log("Testing field builder")

	// 创建字段构建器
	builder := builders.NewFieldBuilder()

	// 设置字段属性
	field := builder.
		WithName("Test Field").
		WithType(table.FieldTypeText).
		WithTableID("table123").
		AsPrimary().
		Build()

	// 验证字段属性
	assert.NotNil(s.T(), field, "Field should not be nil")
	assert.Equal(s.T(), "Test Field", field.Name, "Field name should match")
	assert.Equal(s.T(), table.FieldTypeText, field.Type, "Field type should match")
	assert.Equal(s.T(), "table123", field.TableID, "Field table ID should match")
	assert.True(s.T(), field.IsPrimary, "Field should be primary")
}

// TestMockUserRepository 测试模拟用户仓储
func (s *BasicTestSuite) TestMockUserRepository() {
	s.T().Log("Testing mock user repository")

	// 创建模拟仓储
	mockRepo := new(MockUserRepository)

	// 设置期望
	mockRepo.On("GetByID", mock.Anything, "user123").Return(nil, nil)

	// 调用方法
	user, err := mockRepo.GetByID(nil, "user123")

	// 验证结果
	assert.Nil(s.T(), user, "User should be nil")
	assert.NoError(s.T(), err, "Error should be nil")

	// 验证期望
	mockRepo.AssertExpectations(s.T())
}

// TestMockSpaceRepository 测试模拟空间仓储
func (s *BasicTestSuite) TestMockSpaceRepository() {
	s.T().Log("Testing mock space repository")

	// 创建模拟仓储
	mockRepo := new(MockSpaceRepository)

	// 设置期望
	mockRepo.On("GetByID", mock.Anything, "space123").Return(nil, nil)

	// 调用方法
	space, err := mockRepo.GetByID(nil, "space123")

	// 验证结果
	assert.Nil(s.T(), space, "Space should be nil")
	assert.NoError(s.T(), err, "Error should be nil")

	// 验证期望
	mockRepo.AssertExpectations(s.T())
}

// TestMockTableRepository 测试模拟表仓储
func (s *BasicTestSuite) TestMockTableRepository() {
	s.T().Log("Testing mock table repository")

	// 创建模拟仓储
	mockRepo := new(MockTableRepository)

	// 设置期望
	mockRepo.On("GetByID", mock.Anything, "table123").Return(nil, nil)

	// 调用方法
	table, err := mockRepo.GetByID(nil, "table123")

	// 验证结果
	assert.Nil(s.T(), table, "Table should be nil")
	assert.NoError(s.T(), err, "Error should be nil")

	// 验证期望
	mockRepo.AssertExpectations(s.T())
}

// TestMockFieldRepository 测试模拟字段仓储
func (s *BasicTestSuite) TestMockFieldRepository() {
	s.T().Log("Testing mock field repository")

	// 创建模拟仓储
	mockRepo := new(MockFieldRepository)

	// 设置期望
	mockRepo.On("GetByID", mock.Anything, "field123").Return(nil, nil)

	// 调用方法
	field, err := mockRepo.GetByID(nil, "field123")

	// 验证结果
	assert.Nil(s.T(), field, "Field should be nil")
	assert.NoError(s.T(), err, "Error should be nil")

	// 验证期望
	mockRepo.AssertExpectations(s.T())
}

// TestMockUserDomainService 测试模拟用户领域服务
func (s *BasicTestSuite) TestMockUserDomainService() {
	s.T().Log("Testing mock user domain service")

	// 创建模拟服务
	mockService := new(MockUserDomainService)

	// 设置期望
	mockService.On("GetUser", mock.Anything, "user123").Return(nil, nil)

	// 调用方法
	user, err := mockService.GetUser(nil, "user123")

	// 验证结果
	assert.Nil(s.T(), user, "User should be nil")
	assert.NoError(s.T(), err, "Error should be nil")

	// 验证期望
	mockService.AssertExpectations(s.T())
}

// TestMockTokenService 测试模拟令牌服务
func (s *BasicTestSuite) TestMockTokenService() {
	s.T().Log("Testing mock token service")

	// 创建模拟服务
	mockService := new(MockTokenService)

	// 设置期望
	mockService.On("ValidateToken", mock.Anything, "token123").Return(nil, nil)

	// 调用方法
	claims, err := mockService.ValidateToken(nil, "token123")

	// 验证结果
	assert.Nil(s.T(), claims, "Claims should be nil")
	assert.NoError(s.T(), err, "Error should be nil")

	// 验证期望
	mockService.AssertExpectations(s.T())
}

// TestMockSessionService 测试模拟会话服务
func (s *BasicTestSuite) TestMockSessionService() {
	s.T().Log("Testing mock session service")

	// 创建模拟服务
	mockService := new(MockSessionService)

	// 设置期望
	mockService.On("GetSession", mock.Anything, "session123").Return(nil, nil)

	// 调用方法
	session, err := mockService.GetSession(nil, "session123")

	// 验证结果
	assert.Nil(s.T(), session, "Session should be nil")
	assert.NoError(s.T(), err, "Error should be nil")

	// 验证期望
	mockService.AssertExpectations(s.T())
}

// TestMockCacheService 测试模拟缓存服务
func (s *BasicTestSuite) TestMockCacheService() {
	s.T().Log("Testing mock cache service")

	// 创建模拟服务
	mockService := new(MockCacheService)

	// 设置期望
	mockService.On("Get", mock.Anything, "key123").Return(nil, nil)

	// 调用方法
	value, err := mockService.Get(nil, "key123")

	// 验证结果
	assert.Nil(s.T(), value, "Value should be nil")
	assert.NoError(s.T(), err, "Error should be nil")

	// 验证期望
	mockService.AssertExpectations(s.T())
}

// TestRunnerSuite 运行所有测试
func TestRunnerSuite(t *testing.T) {
	suite.Run(t, new(BasicTestSuite))
}
