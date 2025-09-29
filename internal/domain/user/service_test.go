package user

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"teable-go-backend/internal/testing"
	"teable-go-backend/pkg/errors"
)

func TestUserService_CreateUser(t *testing.T) {
	tests := []struct {
		name           string
		request        CreateUserRequest
		mockSetup      func(*testing.MockUserRepository)
		expectedError  error
		expectedResult func(*User) bool
	}{
		{
			name: "成功创建用户",
			request: CreateUserRequest{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: stringPtr("password123"),
			},
			mockSetup: func(mockRepo *testing.MockUserRepository) {
				mockRepo.On("Exists", mock.Anything, mock.MatchedBy(func(filter ExistsFilter) bool {
					return filter.Email != nil && *filter.Email == "test@example.com"
				})).Return(false, nil)
				mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)
			},
			expectedError: nil,
			expectedResult: func(user *User) bool {
				return user.Name == "Test User" && user.Email == "test@example.com"
			},
		},
		{
			name: "邮箱已存在",
			request: CreateUserRequest{
				Name:     "Test User",
				Email:    "existing@example.com",
				Password: stringPtr("password123"),
			},
			mockSetup: func(mockRepo *testing.MockUserRepository) {
				mockRepo.On("Exists", mock.Anything, mock.MatchedBy(func(filter ExistsFilter) bool {
					return filter.Email != nil && *filter.Email == "existing@example.com"
				})).Return(true, nil)
			},
			expectedError: errors.ErrEmailExists,
		},
		{
			name: "手机号已存在",
			request: CreateUserRequest{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: stringPtr("password123"),
				Phone:    stringPtr("1234567890"),
			},
			mockSetup: func(mockRepo *testing.MockUserRepository) {
				mockRepo.On("Exists", mock.Anything, mock.MatchedBy(func(filter ExistsFilter) bool {
					return filter.Email != nil && *filter.Email == "test@example.com"
				})).Return(false, nil)
				mockRepo.On("Exists", mock.Anything, mock.MatchedBy(func(filter ExistsFilter) bool {
					return filter.Phone != nil && *filter.Phone == "1234567890"
				})).Return(true, nil)
			},
			expectedError: errors.ErrPhoneExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟仓储
			mockRepo := &testing.MockUserRepository{}
			tt.mockSetup(mockRepo)

			// 创建服务
			service := NewService(mockRepo)

			// 执行测试
			result, err := service.CreateUser(context.Background(), tt.request)

			// 验证结果
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Equal(t, tt.expectedError, err)
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

func TestUserService_GetUser(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockSetup      func(*testing.MockUserRepository)
		expectedError  error
		expectedResult func(*User) bool
	}{
		{
			name:   "成功获取用户",
			userID: "user-123",
			mockSetup: func(mockRepo *testing.MockUserRepository) {
				expectedUser := testing.NewUserBuilder().
					WithID("user-123").
					WithName("Test User").
					WithEmail("test@example.com").
					Build()
				mockRepo.On("GetByID", mock.Anything, "user-123").Return(expectedUser, nil)
			},
			expectedError: nil,
			expectedResult: func(user *User) bool {
				return user.ID == "user-123" && user.Name == "Test User"
			},
		},
		{
			name:   "用户不存在",
			userID: "non-existent",
			mockSetup: func(mockRepo *testing.MockUserRepository) {
				mockRepo.On("GetByID", mock.Anything, "non-existent").Return(nil, nil)
			},
			expectedError: errors.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟仓储
			mockRepo := &testing.MockUserRepository{}
			tt.mockSetup(mockRepo)

			// 创建服务
			service := NewService(mockRepo)

			// 执行测试
			result, err := service.GetUser(context.Background(), tt.userID)

			// 验证结果
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Equal(t, tt.expectedError, err)
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

func TestUserService_Authenticate(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		password       string
		mockSetup      func(*testing.MockUserRepository)
		expectedError  error
		expectedResult func(*User) bool
	}{
		{
			name:     "成功认证",
			email:    "test@example.com",
			password: "password123",
			mockSetup: func(mockRepo *testing.MockUserRepository) {
				user := testing.NewUserBuilder().
					WithEmail("test@example.com").
					WithPassword("password123").
					Build()
				mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)
			},
			expectedError: nil,
			expectedResult: func(user *User) bool {
				return user.Email == "test@example.com"
			},
		},
		{
			name:     "用户不存在",
			email:    "nonexistent@example.com",
			password: "password123",
			mockSetup: func(mockRepo *testing.MockUserRepository) {
				mockRepo.On("GetByEmail", mock.Anything, "nonexistent@example.com").Return(nil, nil)
			},
			expectedError: errors.ErrUserNotFound,
		},
		{
			name:     "密码错误",
			email:    "test@example.com",
			password: "wrongpassword",
			mockSetup: func(mockRepo *testing.MockUserRepository) {
				user := testing.NewUserBuilder().
					WithEmail("test@example.com").
					WithPassword("password123").
					Build()
				mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)
			},
			expectedError: errors.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟仓储
			mockRepo := &testing.MockUserRepository{}
			tt.mockSetup(mockRepo)

			// 创建服务
			service := NewService(mockRepo)

			// 执行测试
			result, err := service.Authenticate(context.Background(), tt.email, tt.password)

			// 验证结果
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Equal(t, tt.expectedError, err)
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

func TestUserService_ChangePassword(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		oldPassword   string
		newPassword   string
		mockSetup     func(*testing.MockUserRepository)
		expectedError error
	}{
		{
			name:        "成功修改密码",
			userID:      "user-123",
			oldPassword: "oldpassword",
			newPassword: "newpassword123",
			mockSetup: func(mockRepo *testing.MockUserRepository) {
				user := testing.NewUserBuilder().
					WithID("user-123").
					WithPassword("oldpassword").
					Build()
				mockRepo.On("GetByID", mock.Anything, "user-123").Return(user, nil)
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:        "旧密码错误",
			userID:      "user-123",
			oldPassword: "wrongpassword",
			newPassword: "newpassword123",
			mockSetup: func(mockRepo *testing.MockUserRepository) {
				user := testing.NewUserBuilder().
					WithID("user-123").
					WithPassword("oldpassword").
					Build()
				mockRepo.On("GetByID", mock.Anything, "user-123").Return(user, nil)
			},
			expectedError: errors.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟仓储
			mockRepo := &testing.MockUserRepository{}
			tt.mockSetup(mockRepo)

			// 创建服务
			service := NewService(mockRepo)

			// 执行测试
			err := service.ChangePassword(context.Background(), tt.userID, tt.oldPassword, tt.newPassword)

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

func TestUserService_ListUsers(t *testing.T) {
	tests := []struct {
		name           string
		filter         ListFilter
		mockSetup      func(*testing.MockUserRepository)
		expectedError  error
		expectedResult func(*PaginatedResult) bool
	}{
		{
			name: "成功获取用户列表",
			filter: ListFilter{
				Offset: 0,
				Limit:  10,
			},
			mockSetup: func(mockRepo *testing.MockUserRepository) {
				users := []*User{
					testing.NewUserBuilder().WithID("user-1").WithName("User 1").Build(),
					testing.NewUserBuilder().WithID("user-2").WithName("User 2").Build(),
				}
				mockRepo.On("List", mock.Anything, mock.AnythingOfType("user.ListFilter")).Return(users, nil)
				mockRepo.On("Count", mock.Anything, mock.AnythingOfType("user.CountFilter")).Return(int64(2), nil)
			},
			expectedError: nil,
			expectedResult: func(result *PaginatedResult) bool {
				return result.Total == 2 && len(result.Users) == 2
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟仓储
			mockRepo := &testing.MockUserRepository{}
			tt.mockSetup(mockRepo)

			// 创建服务
			service := NewService(mockRepo)

			// 执行测试
			result, err := service.ListUsers(context.Background(), tt.filter)

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

func TestUserService_GetUserStats(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*testing.MockUserRepository)
		expectedError  error
		expectedResult func(*UserStats) bool
	}{
		{
			name: "成功获取用户统计",
			mockSetup: func(mockRepo *testing.MockUserRepository) {
				mockRepo.On("Count", mock.Anything, mock.MatchedBy(func(filter CountFilter) bool {
					return true // 匹配所有Count调用
				})).Return(int64(100), nil).Times(7) // 7次不同的统计查询
			},
			expectedError: nil,
			expectedResult: func(stats *UserStats) bool {
				return stats.TotalUsers == 100
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟仓储
			mockRepo := &testing.MockUserRepository{}
			tt.mockSetup(mockRepo)

			// 创建服务
			service := NewService(mockRepo)

			// 执行测试
			result, err := service.GetUserStats(context.Background())

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

// 辅助函数
func stringPtr(s string) *string {
	return &s
}
