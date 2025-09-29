package space

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"teable-go-backend/internal/testing"
)

func TestSpaceService_CreateSpace(t *testing.T) {
	tests := []struct {
		name           string
		request        CreateSpaceRequest
		mockSetup      func(*testing.MockSpaceRepository)
		expectedError  error
		expectedResult func(*Space) bool
	}{
		{
			name: "成功创建空间",
			request: CreateSpaceRequest{
				Name:        "Test Space",
				Description: stringPtr("Test space description"),
				CreatedBy:   "user-123",
			},
			mockSetup: func(mockRepo *testing.MockSpaceRepository) {
				mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*space.Space")).Return(nil)
			},
			expectedError: nil,
			expectedResult: func(space *Space) bool {
				return space.Name == "Test Space" && space.CreatedBy == "user-123"
			},
		},
		{
			name: "创建空间失败",
			request: CreateSpaceRequest{
				Name:        "Test Space",
				Description: stringPtr("Test space description"),
				CreatedBy:   "user-123",
			},
			mockSetup: func(mockRepo *testing.MockSpaceRepository) {
				mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*space.Space")).Return(assert.AnError)
			},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟仓储
			mockRepo := &testing.MockSpaceRepository{}
			tt.mockSetup(mockRepo)

			// 创建服务
			service := NewService(mockRepo)

			// 执行测试
			result, err := service.CreateSpace(context.Background(), tt.request)

			// 验证结果
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.expectedResult != nil {
					assert.True(t, tt.expectedResult(result))
				}
			}

			// 验证模拟调用
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSpaceService_GetSpace(t *testing.T) {
	tests := []struct {
		name           string
		spaceID        string
		mockSetup      func(*testing.MockSpaceRepository)
		expectedError  error
		expectedResult func(*Space) bool
	}{
		{
			name:    "成功获取空间",
			spaceID: "space-123",
			mockSetup: func(mockRepo *testing.MockSpaceRepository) {
				expectedSpace := testing.NewSpaceBuilder().
					WithID("space-123").
					WithName("Test Space").
					WithCreatedBy("user-123").
					Build()
				mockRepo.On("GetByID", mock.Anything, "space-123").Return(expectedSpace, nil)
			},
			expectedError: nil,
			expectedResult: func(space *Space) bool {
				return space.ID == "space-123" && space.Name == "Test Space"
			},
		},
		{
			name:    "空间不存在",
			spaceID: "non-existent",
			mockSetup: func(mockRepo *testing.MockSpaceRepository) {
				mockRepo.On("GetByID", mock.Anything, "non-existent").Return(nil, nil)
			},
			expectedError: assert.AnError, // 期望返回错误
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟仓储
			mockRepo := &testing.MockSpaceRepository{}
			tt.mockSetup(mockRepo)

			// 创建服务
			service := NewService(mockRepo)

			// 执行测试
			result, err := service.GetSpace(context.Background(), tt.spaceID)

			// 验证结果
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.expectedResult != nil {
					assert.True(t, tt.expectedResult(result))
				}
			}

			// 验证模拟调用
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSpaceService_UpdateSpace(t *testing.T) {
	tests := []struct {
		name           string
		spaceID        string
		request        UpdateSpaceRequest
		mockSetup      func(*testing.MockSpaceRepository)
		expectedError  error
		expectedResult func(*Space) bool
	}{
		{
			name:    "成功更新空间",
			spaceID: "space-123",
			request: UpdateSpaceRequest{
				Name:        stringPtr("Updated Space"),
				Description: stringPtr("Updated description"),
			},
			mockSetup: func(mockRepo *testing.MockSpaceRepository) {
				existingSpace := testing.NewSpaceBuilder().
					WithID("space-123").
					WithName("Original Space").
					Build()
				mockRepo.On("GetByID", mock.Anything, "space-123").Return(existingSpace, nil)
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*space.Space")).Return(nil)
			},
			expectedError: nil,
			expectedResult: func(space *Space) bool {
				return space.Name == "Updated Space"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟仓储
			mockRepo := &testing.MockSpaceRepository{}
			tt.mockSetup(mockRepo)

			// 创建服务
			service := NewService(mockRepo)

			// 执行测试
			result, err := service.UpdateSpace(context.Background(), tt.spaceID, tt.request)

			// 验证结果
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.expectedResult != nil {
					assert.True(t, tt.expectedResult(result))
				}
			}

			// 验证模拟调用
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSpaceService_DeleteSpace(t *testing.T) {
	tests := []struct {
		name          string
		spaceID       string
		mockSetup     func(*testing.MockSpaceRepository)
		expectedError error
	}{
		{
			name:    "成功删除空间",
			spaceID: "space-123",
			mockSetup: func(mockRepo *testing.MockSpaceRepository) {
				existingSpace := testing.NewSpaceBuilder().
					WithID("space-123").
					WithName("Test Space").
					Build()
				mockRepo.On("GetByID", mock.Anything, "space-123").Return(existingSpace, nil)
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*space.Space")).Return(nil)
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟仓储
			mockRepo := &testing.MockSpaceRepository{}
			tt.mockSetup(mockRepo)

			// 创建服务
			service := NewService(mockRepo)

			// 执行测试
			err := service.DeleteSpace(context.Background(), tt.spaceID)

			// 验证结果
			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// 验证模拟调用
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSpaceService_ListSpaces(t *testing.T) {
	tests := []struct {
		name           string
		filter         ListFilter
		mockSetup      func(*testing.MockSpaceRepository)
		expectedError  error
		expectedResult func([]*Space, int64) bool
	}{
		{
			name: "成功获取空间列表",
			filter: ListFilter{
				Offset: 0,
				Limit:  10,
			},
			mockSetup: func(mockRepo *testing.MockSpaceRepository) {
				spaces := []*Space{
					testing.NewSpaceBuilder().WithID("space-1").WithName("Space 1").Build(),
					testing.NewSpaceBuilder().WithID("space-2").WithName("Space 2").Build(),
				}
				mockRepo.On("List", mock.Anything, mock.AnythingOfType("space.ListFilter")).Return(spaces, nil)
				mockRepo.On("Count", mock.Anything, mock.AnythingOfType("space.CountFilter")).Return(int64(2), nil)
			},
			expectedError: nil,
			expectedResult: func(spaces []*Space, total int64) bool {
				return total == 2 && len(spaces) == 2
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟仓储
			mockRepo := &testing.MockSpaceRepository{}
			tt.mockSetup(mockRepo)

			// 创建服务
			service := NewService(mockRepo)

			// 执行测试
			spaces, total, err := service.ListSpaces(context.Background(), tt.filter)

			// 验证结果
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, spaces)
				assert.Equal(t, int64(0), total)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, spaces)
				if tt.expectedResult != nil {
					assert.True(t, tt.expectedResult(spaces, total))
				}
			}

			// 验证模拟调用
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSpaceService_AddCollaborator(t *testing.T) {
	tests := []struct {
		name          string
		spaceID       string
		userID        string
		role          string
		mockSetup     func(*testing.MockSpaceRepository)
		expectedError error
	}{
		{
			name:    "成功添加协作者",
			spaceID: "space-123",
			userID:  "user-456",
			role:    "editor",
			mockSetup: func(mockRepo *testing.MockSpaceRepository) {
				space := testing.NewSpaceBuilder().
					WithID("space-123").
					WithCreatedBy("user-123").
					Build()
				mockRepo.On("GetByID", mock.Anything, "space-123").Return(space, nil)
				mockRepo.On("AddCollaborator", mock.Anything, mock.AnythingOfType("*space.SpaceCollaborator")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:    "无效角色",
			spaceID: "space-123",
			userID:  "user-456",
			role:    "invalid-role",
			mockSetup: func(mockRepo *testing.MockSpaceRepository) {
				space := testing.NewSpaceBuilder().
					WithID("space-123").
					WithCreatedBy("user-123").
					Build()
				mockRepo.On("GetByID", mock.Anything, "space-123").Return(space, nil)
			},
			expectedError: ErrInvalidRole,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟仓储
			mockRepo := &testing.MockSpaceRepository{}
			tt.mockSetup(mockRepo)

			// 创建服务
			service := NewService(mockRepo)

			// 执行测试
			err := service.AddCollaborator(context.Background(), tt.spaceID, tt.userID, tt.role)

			// 验证结果
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			// 验证模拟调用
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSpaceService_RemoveCollaborator(t *testing.T) {
	tests := []struct {
		name          string
		collabID      string
		mockSetup     func(*testing.MockSpaceRepository)
		expectedError error
	}{
		{
			name:     "成功移除协作者",
			collabID: "collab-123",
			mockSetup: func(mockRepo *testing.MockSpaceRepository) {
				mockRepo.On("RemoveCollaborator", mock.Anything, "collab-123").Return(nil)
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟仓储
			mockRepo := &testing.MockSpaceRepository{}
			tt.mockSetup(mockRepo)

			// 创建服务
			service := NewService(mockRepo)

			// 执行测试
			err := service.RemoveCollaborator(context.Background(), tt.collabID)

			// 验证结果
			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// 验证模拟调用
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSpaceService_ListCollaborators(t *testing.T) {
	tests := []struct {
		name           string
		spaceID        string
		mockSetup      func(*testing.MockSpaceRepository)
		expectedError  error
		expectedResult func([]*SpaceCollaborator) bool
	}{
		{
			name:    "成功获取协作者列表",
			spaceID: "space-123",
			mockSetup: func(mockRepo *testing.MockSpaceRepository) {
				collaborators := []*SpaceCollaborator{
					{
						ID:      "collab-1",
						SpaceID: "space-123",
						UserID:  "user-456",
						Role:    CollaboratorRoleEditor,
					},
				}
				mockRepo.On("ListCollaborators", mock.Anything, "space-123").Return(collaborators, nil)
			},
			expectedError: nil,
			expectedResult: func(collaborators []*SpaceCollaborator) bool {
				return len(collaborators) == 1 && collaborators[0].UserID == "user-456"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟仓储
			mockRepo := &testing.MockSpaceRepository{}
			tt.mockSetup(mockRepo)

			// 创建服务
			service := NewService(mockRepo)

			// 执行测试
			result, err := service.ListCollaborators(context.Background(), tt.spaceID)

			// 验证结果
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.expectedResult != nil {
					assert.True(t, tt.expectedResult(result))
				}
			}

			// 验证模拟调用
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSpaceService_CheckUserPermission(t *testing.T) {
	tests := []struct {
		name           string
		spaceID        string
		userID         string
		permission     string
		mockSetup      func(*testing.MockSpaceRepository)
		expectedError  error
		expectedResult bool
	}{
		{
			name:       "空间创建者有所有权限",
			spaceID:    "space-123",
			userID:     "user-123",
			permission: "read",
			mockSetup: func(mockRepo *testing.MockSpaceRepository) {
				space := testing.NewSpaceBuilder().
					WithID("space-123").
					WithCreatedBy("user-123").
					Build()
				mockRepo.On("GetByID", mock.Anything, "space-123").Return(space, nil)
			},
			expectedError:  nil,
			expectedResult: true,
		},
		{
			name:       "协作者有相应权限",
			spaceID:    "space-123",
			userID:     "user-456",
			permission: "read",
			mockSetup: func(mockRepo *testing.MockSpaceRepository) {
				space := testing.NewSpaceBuilder().
					WithID("space-123").
					WithCreatedBy("user-123").
					Build()
				collaborators := []*SpaceCollaborator{
					{
						ID:      "collab-1",
						SpaceID: "space-123",
						UserID:  "user-456",
						Role:    CollaboratorRoleEditor,
					},
				}
				mockRepo.On("GetByID", mock.Anything, "space-123").Return(space, nil)
				mockRepo.On("ListCollaborators", mock.Anything, "space-123").Return(collaborators, nil)
			},
			expectedError:  nil,
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟仓储
			mockRepo := &testing.MockSpaceRepository{}
			tt.mockSetup(mockRepo)

			// 创建服务
			service := NewService(mockRepo)

			// 执行测试
			result, err := service.CheckUserPermission(context.Background(), tt.spaceID, tt.userID, tt.permission)

			// 验证结果
			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			// 验证模拟调用
			mockRepo.AssertExpectations(t)
		})
	}
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}
