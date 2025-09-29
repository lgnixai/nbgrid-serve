package application

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"teable-go-backend/internal/domain/space"
	"teable-go-backend/pkg/logger"
)

// MockSpaceRepository 模拟空间仓储
type MockSpaceRepository struct {
	mock.Mock
}

func (m *MockSpaceRepository) Create(ctx context.Context, space *space.Space) error {
	args := m.Called(ctx, space)
	return args.Error(0)
}

func (m *MockSpaceRepository) GetByID(ctx context.Context, id string) (*space.Space, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*space.Space), args.Error(1)
}

func (m *MockSpaceRepository) Update(ctx context.Context, space *space.Space) error {
	args := m.Called(ctx, space)
	return args.Error(0)
}

func (m *MockSpaceRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSpaceRepository) List(ctx context.Context, filter space.ListFilter) ([]*space.Space, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*space.Space), args.Error(1)
}

func (m *MockSpaceRepository) Count(ctx context.Context, filter space.CountFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockSpaceRepository) AddCollaborator(ctx context.Context, collab *space.SpaceCollaborator) error {
	args := m.Called(ctx, collab)
	return args.Error(0)
}

func (m *MockSpaceRepository) RemoveCollaborator(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSpaceRepository) ListCollaborators(ctx context.Context, spaceID string) ([]*space.SpaceCollaborator, error) {
	args := m.Called(ctx, spaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*space.SpaceCollaborator), args.Error(1)
}

func (m *MockSpaceRepository) ListDeleted(ctx context.Context, filter space.ListFilter) ([]*space.Space, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*space.Space), args.Error(1)
}

func (m *MockSpaceRepository) CountDeleted(ctx context.Context, filter space.CountFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// MockSpaceDomainService 模拟空间领域服务
type MockSpaceDomainService struct {
	mock.Mock
}

func (m *MockSpaceDomainService) CreateSpace(ctx context.Context, req space.CreateSpaceRequest) (*space.Space, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*space.Space), args.Error(1)
}

func (m *MockSpaceDomainService) GetSpace(ctx context.Context, id string) (*space.Space, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*space.Space), args.Error(1)
}

func (m *MockSpaceDomainService) UpdateSpace(ctx context.Context, id string, req space.UpdateSpaceRequest) (*space.Space, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*space.Space), args.Error(1)
}

func (m *MockSpaceDomainService) DeleteSpace(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSpaceDomainService) ListSpaces(ctx context.Context, filter space.ListFilter) ([]*space.Space, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*space.Space), args.Get(1).(int64), args.Error(2)
}

func (m *MockSpaceDomainService) AddCollaborator(ctx context.Context, spaceID, userID, role string) error {
	args := m.Called(ctx, spaceID, userID, role)
	return args.Error(0)
}

func (m *MockSpaceDomainService) RemoveCollaborator(ctx context.Context, collabID string) error {
	args := m.Called(ctx, collabID)
	return args.Error(0)
}

func (m *MockSpaceDomainService) ListCollaborators(ctx context.Context, spaceID string) ([]*space.SpaceCollaborator, error) {
	args := m.Called(ctx, spaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*space.SpaceCollaborator), args.Error(1)
}

func (m *MockSpaceDomainService) UpdateCollaboratorRole(ctx context.Context, collabID, role string) error {
	args := m.Called(ctx, collabID, role)
	return args.Error(0)
}

func (m *MockSpaceDomainService) BulkUpdateSpaces(ctx context.Context, updates []space.BulkUpdateRequest) error {
	args := m.Called(ctx, updates)
	return args.Error(0)
}

func (m *MockSpaceDomainService) BulkDeleteSpaces(ctx context.Context, spaceIDs []string) error {
	args := m.Called(ctx, spaceIDs)
	return args.Error(0)
}

func (m *MockSpaceDomainService) CheckUserPermission(ctx context.Context, spaceID, userID, permission string) (bool, error) {
	args := m.Called(ctx, spaceID, userID, permission)
	return args.Bool(0), args.Error(1)
}

func (m *MockSpaceDomainService) GetUserSpaces(ctx context.Context, userID string, filter space.ListFilter) ([]*space.Space, int64, error) {
	args := m.Called(ctx, userID, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*space.Space), args.Get(1).(int64), args.Error(2)
}

func (m *MockSpaceDomainService) GetSpaceStats(ctx context.Context, spaceID string) (*space.SpaceStats, error) {
	args := m.Called(ctx, spaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*space.SpaceStats), args.Error(1)
}

func (m *MockSpaceDomainService) GetUserDeletedSpaces(ctx context.Context, userID string, filter space.ListFilter) ([]*space.Space, int64, error) {
	args := m.Called(ctx, userID, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*space.Space), args.Get(1).(int64), args.Error(2)
}

func (m *MockSpaceDomainService) GetUserSpaceStats(ctx context.Context, userID string) (*space.UserSpaceStats, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*space.UserSpaceStats), args.Error(1)
}

// MockLogger 模拟日志记录器
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Info(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Warn(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Error(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Fatal(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

// 测试用例

func TestSpaceApplicationService_CreateSpace(t *testing.T) {
	// 准备测试数据
	mockRepo := new(MockSpaceRepository)
	mockDomainService := new(MockSpaceDomainService)
	mockLogger := new(MockLogger)

	service := NewSpaceApplicationService(mockRepo, mockDomainService, mockLogger)

	ctx := context.Background()
	req := CreateSpaceRequest{
		Name:        "测试空间",
		Description: stringPtr("这是一个测试空间"),
		CreatedBy:   "user123",
	}

	// 创建预期的空间实体
	expectedSpace := &space.Space{
		ID:          "space123",
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   req.CreatedBy,
		CreatedTime: time.Now(),
	}

	// 设置模拟期望
	mockLogger.On("Info", "Creating space", mock.Anything).Return()
	mockLogger.On("Info", "Space created successfully", mock.Anything).Return()
	
	mockDomainService.On("CreateSpace", ctx, mock.MatchedBy(func(domainReq space.CreateSpaceRequest) bool {
		return domainReq.Name == req.Name && 
			   domainReq.CreatedBy == req.CreatedBy &&
			   *domainReq.Description == *req.Description
	})).Return(expectedSpace, nil)

	// 执行测试
	result, err := service.CreateSpace(ctx, req)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedSpace.ID, result.ID)
	assert.Equal(t, expectedSpace.Name, result.Name)
	assert.Equal(t, expectedSpace.Description, result.Description)
	assert.Equal(t, expectedSpace.CreatedBy, result.CreatedBy)

	// 验证模拟调用
	mockDomainService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestSpaceApplicationService_GetSpace(t *testing.T) {
	// 准备测试数据
	mockRepo := new(MockSpaceRepository)
	mockDomainService := new(MockSpaceDomainService)
	mockLogger := new(MockLogger)

	service := NewSpaceApplicationService(mockRepo, mockDomainService, mockLogger)

	ctx := context.Background()
	spaceID := "space123"
	userID := "user123"

	// 创建预期的空间实体
	expectedSpace := &space.Space{
		ID:          spaceID,
		Name:        "测试空间",
		CreatedBy:   userID,
		CreatedTime: time.Now(),
	}

	// 设置模拟期望
	mockLogger.On("Info", "Getting space", mock.Anything).Return()
	
	mockDomainService.On("GetSpace", ctx, spaceID).Return(expectedSpace, nil)
	mockDomainService.On("CheckUserPermission", ctx, spaceID, userID, "read").Return(true, nil)

	// 执行测试
	result, err := service.GetSpace(ctx, spaceID, userID)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedSpace.ID, result.ID)
	assert.Equal(t, expectedSpace.Name, result.Name)

	// 验证模拟调用
	mockDomainService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestSpaceApplicationService_UpdateSpace(t *testing.T) {
	// 准备测试数据
	mockRepo := new(MockSpaceRepository)
	mockDomainService := new(MockSpaceDomainService)
	mockLogger := new(MockLogger)

	service := NewSpaceApplicationService(mockRepo, mockDomainService, mockLogger)

	ctx := context.Background()
	spaceID := "space123"
	userID := "user123"
	req := UpdateSpaceRequest{
		Name: stringPtr("更新后的空间名称"),
	}

	// 创建预期的空间实体
	updatedSpace := &space.Space{
		ID:               spaceID,
		Name:             *req.Name,
		CreatedBy:        userID,
		CreatedTime:      time.Now(),
		LastModifiedTime: timePtr(time.Now()),
	}

	// 设置模拟期望
	mockLogger.On("Info", "Updating space", mock.Anything).Return()
	mockLogger.On("Info", "Space updated successfully", mock.Anything).Return()
	
	mockDomainService.On("CheckUserPermission", ctx, spaceID, userID, "update").Return(true, nil)
	mockDomainService.On("UpdateSpace", ctx, spaceID, mock.MatchedBy(func(domainReq space.UpdateSpaceRequest) bool {
		return domainReq.Name != nil && *domainReq.Name == *req.Name
	})).Return(updatedSpace, nil)

	// 执行测试
	result, err := service.UpdateSpace(ctx, spaceID, req, userID)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, updatedSpace.ID, result.ID)
	assert.Equal(t, updatedSpace.Name, result.Name)

	// 验证模拟调用
	mockDomainService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestSpaceApplicationService_DeleteSpace(t *testing.T) {
	// 准备测试数据
	mockRepo := new(MockSpaceRepository)
	mockDomainService := new(MockSpaceDomainService)
	mockLogger := new(MockLogger)

	service := NewSpaceApplicationService(mockRepo, mockDomainService, mockLogger)

	ctx := context.Background()
	spaceID := "space123"
	userID := "user123"

	// 设置模拟期望
	mockLogger.On("Info", "Deleting space", mock.Anything).Return()
	mockLogger.On("Info", "Space deleted successfully", mock.Anything).Return()
	
	mockDomainService.On("CheckUserPermission", ctx, spaceID, userID, "delete").Return(true, nil)
	mockDomainService.On("DeleteSpace", ctx, spaceID).Return(nil)

	// 执行测试
	err := service.DeleteSpace(ctx, spaceID, userID)

	// 验证结果
	assert.NoError(t, err)

	// 验证模拟调用
	mockDomainService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestSpaceApplicationService_ListSpaces(t *testing.T) {
	// 准备测试数据
	mockRepo := new(MockSpaceRepository)
	mockDomainService := new(MockSpaceDomainService)
	mockLogger := new(MockLogger)

	service := NewSpaceApplicationService(mockRepo, mockDomainService, mockLogger)

	ctx := context.Background()
	userID := "user123"
	req := ListSpacesRequest{
		Offset: 0,
		Limit:  20,
	}

	// 创建预期的空间列表
	expectedSpaces := []*space.Space{
		{
			ID:          "space1",
			Name:        "空间1",
			CreatedBy:   userID,
			CreatedTime: time.Now(),
		},
		{
			ID:          "space2",
			Name:        "空间2",
			CreatedBy:   userID,
			CreatedTime: time.Now(),
		},
	}
	expectedTotal := int64(2)

	// 设置模拟期望
	mockLogger.On("Info", "Listing spaces", mock.Anything).Return()
	mockLogger.On("Info", "Spaces listed successfully", mock.Anything).Return()
	
	mockDomainService.On("GetUserSpaces", ctx, userID, mock.MatchedBy(func(filter space.ListFilter) bool {
		return filter.Offset == req.Offset && filter.Limit == req.Limit
	})).Return(expectedSpaces, expectedTotal, nil)

	// 执行测试
	result, err := service.ListSpaces(ctx, req, userID)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedTotal, result.Total)
	assert.Len(t, result.Data, len(expectedSpaces))
	assert.Equal(t, expectedSpaces[0].ID, result.Data[0].ID)
	assert.Equal(t, expectedSpaces[1].ID, result.Data[1].ID)

	// 验证模拟调用
	mockDomainService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestSpaceApplicationService_RestoreSpace(t *testing.T) {
	// 准备测试数据
	mockRepo := new(MockSpaceRepository)
	mockDomainService := new(MockSpaceDomainService)
	mockLogger := new(MockLogger)

	service := NewSpaceApplicationService(mockRepo, mockDomainService, mockLogger)

	ctx := context.Background()
	spaceID := "space123"
	userID := "user123"

	// 创建已删除的空间实体
	deletedTime := time.Now()
	deletedSpace := &space.Space{
		ID:          spaceID,
		Name:        "已删除的空间",
		CreatedBy:   userID,
		CreatedTime: time.Now(),
		DeletedTime: &deletedTime,
	}

	// 设置模拟期望
	mockLogger.On("Info", "Restoring space", mock.Anything).Return()
	mockLogger.On("Info", "Space restored successfully", mock.Anything).Return()
	
	mockDomainService.On("CheckUserPermission", ctx, spaceID, userID, "restore").Return(true, nil)
	mockRepo.On("GetByID", ctx, spaceID).Return(deletedSpace, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(s *space.Space) bool {
		return s.ID == spaceID && !s.IsDeleted()
	})).Return(nil)

	// 执行测试
	result, err := service.RestoreSpace(ctx, spaceID, userID)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, spaceID, result.ID)
	assert.False(t, result.IsDeleted)

	// 验证模拟调用
	mockDomainService.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestSpaceApplicationService_PermanentDeleteSpace(t *testing.T) {
	// 准备测试数据
	mockRepo := new(MockSpaceRepository)
	mockDomainService := new(MockSpaceDomainService)
	mockLogger := new(MockLogger)

	service := NewSpaceApplicationService(mockRepo, mockDomainService, mockLogger)

	ctx := context.Background()
	spaceID := "space123"
	userID := "user123"

	// 创建已删除的空间实体
	deletedTime := time.Now()
	deletedSpace := &space.Space{
		ID:          spaceID,
		Name:        "已删除的空间",
		CreatedBy:   userID,
		CreatedTime: time.Now(),
		DeletedTime: &deletedTime,
	}

	// 设置模拟期望
	mockLogger.On("Info", "Permanently deleting space", mock.Anything).Return()
	mockLogger.On("Info", "Space permanently deleted successfully", mock.Anything).Return()
	
	mockDomainService.On("CheckUserPermission", ctx, spaceID, userID, "delete").Return(true, nil)
	mockRepo.On("GetByID", ctx, spaceID).Return(deletedSpace, nil)
	mockRepo.On("Delete", ctx, spaceID).Return(nil)

	// 执行测试
	err := service.PermanentDeleteSpace(ctx, spaceID, userID)

	// 验证结果
	assert.NoError(t, err)

	// 验证模拟调用
	mockDomainService.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestSpaceApplicationService_GetDeletedSpaces(t *testing.T) {
	// 准备测试数据
	mockRepo := new(MockSpaceRepository)
	mockDomainService := new(MockSpaceDomainService)
	mockLogger := new(MockLogger)

	service := NewSpaceApplicationService(mockRepo, mockDomainService, mockLogger)

	ctx := context.Background()
	userID := "user123"
	req := ListDeletedSpacesRequest{
		Offset: 0,
		Limit:  20,
	}

	// 创建已删除的空间列表
	deletedTime := time.Now()
	expectedSpaces := []*space.Space{
		{
			ID:          "space1",
			Name:        "已删除空间1",
			CreatedBy:   userID,
			CreatedTime: time.Now(),
			DeletedTime: &deletedTime,
		},
		{
			ID:          "space2",
			Name:        "已删除空间2",
			CreatedBy:   userID,
			CreatedTime: time.Now(),
			DeletedTime: &deletedTime,
		},
	}
	expectedTotal := int64(2)

	// 设置模拟期望
	mockLogger.On("Info", "Getting deleted spaces", mock.Anything).Return()
	mockLogger.On("Info", "Deleted spaces retrieved successfully", mock.Anything).Return()
	
	mockDomainService.On("GetUserDeletedSpaces", ctx, userID, mock.MatchedBy(func(filter space.ListFilter) bool {
		return filter.Offset == req.Offset && filter.Limit == req.Limit
	})).Return(expectedSpaces, expectedTotal, nil)

	// 执行测试
	result, err := service.GetDeletedSpaces(ctx, userID, req)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedTotal, result.Total)
	assert.Len(t, result.Data, len(expectedSpaces))
	assert.Equal(t, expectedSpaces[0].ID, result.Data[0].ID)
	assert.Equal(t, expectedSpaces[1].ID, result.Data[1].ID)
	assert.True(t, result.Data[0].IsDeleted)
	assert.True(t, result.Data[1].IsDeleted)

	// 验证模拟调用
	mockDomainService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestSpaceApplicationService_BulkRestoreSpaces(t *testing.T) {
	// 准备测试数据
	mockRepo := new(MockSpaceRepository)
	mockDomainService := new(MockSpaceDomainService)
	mockLogger := new(MockLogger)

	service := NewSpaceApplicationService(mockRepo, mockDomainService, mockLogger)

	ctx := context.Background()
	userID := "user123"
	spaceIDs := []string{"space1", "space2"}

	// 创建已删除的空间实体
	deletedTime := time.Now()
	deletedSpaces := []*space.Space{
		{
			ID:          "space1",
			Name:        "已删除空间1",
			CreatedBy:   userID,
			CreatedTime: time.Now(),
			DeletedTime: &deletedTime,
		},
		{
			ID:          "space2",
			Name:        "已删除空间2",
			CreatedBy:   userID,
			CreatedTime: time.Now(),
			DeletedTime: &deletedTime,
		},
	}

	// 设置模拟期望
	mockLogger.On("Info", "Bulk restoring spaces", mock.Anything).Return()
	mockLogger.On("Info", "Spaces bulk restored successfully", mock.Anything).Return()
	
	// 权限检查
	for _, spaceID := range spaceIDs {
		mockDomainService.On("CheckUserPermission", ctx, spaceID, userID, "restore").Return(true, nil)
	}
	
	// 获取空间实体
	for i, spaceID := range spaceIDs {
		mockRepo.On("GetByID", ctx, spaceID).Return(deletedSpaces[i], nil)
		mockRepo.On("Update", ctx, mock.MatchedBy(func(s *space.Space) bool {
			return s.ID == spaceID && !s.IsDeleted()
		})).Return(nil)
	}

	// 执行测试
	err := service.BulkRestoreSpaces(ctx, spaceIDs, userID)

	// 验证结果
	assert.NoError(t, err)

	// 验证模拟调用
	mockDomainService.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestSpaceApplicationService_GetSpaceStats(t *testing.T) {
	// 准备测试数据
	mockRepo := new(MockSpaceRepository)
	mockDomainService := new(MockSpaceDomainService)
	mockLogger := new(MockLogger)

	service := NewSpaceApplicationService(mockRepo, mockDomainService, mockLogger)

	ctx := context.Background()
	spaceID := "space123"
	userID := "user123"

	// 创建空间实体
	spaceEntity := &space.Space{
		ID:          spaceID,
		Name:        "测试空间",
		CreatedBy:   userID,
		CreatedTime: time.Now(),
	}

	// 创建统计信息
	lastActivity := time.Now().Format(time.RFC3339)
	stats := &space.SpaceStats{
		SpaceID:            spaceID,
		TotalBases:         5,
		TotalTables:        15,
		TotalRecords:       1000,
		TotalCollaborators: 3,
		LastActivityAt:     &lastActivity,
	}

	// 设置模拟期望
	mockLogger.On("Info", "Getting space stats", mock.Anything).Return()
	mockLogger.On("Info", "Space stats retrieved successfully", mock.Anything).Return()
	
	mockDomainService.On("CheckUserPermission", ctx, spaceID, userID, "read").Return(true, nil)
	mockDomainService.On("GetSpace", ctx, spaceID).Return(spaceEntity, nil)
	mockDomainService.On("GetSpaceStats", ctx, spaceID).Return(stats, nil)

	// 执行测试
	result, err := service.GetSpaceStats(ctx, spaceID, userID)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, spaceID, result.SpaceID)
	assert.Equal(t, stats.TotalBases, result.TotalBases)
	assert.Equal(t, stats.TotalTables, result.TotalTables)
	assert.Equal(t, stats.TotalRecords, result.TotalRecords)
	assert.Equal(t, stats.TotalCollaborators, result.TotalCollaborators)
	assert.NotNil(t, result.LastActivityAt)

	// 验证模拟调用
	mockDomainService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}