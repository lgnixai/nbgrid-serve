package application

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	userDomain "teable-go-backend/internal/domain/user"
	"teable-go-backend/internal/testing"
	"teable-go-backend/pkg/errors"
)

func TestUserService_Register(t *testing.T) {
	tests := []struct {
		name           string
		request        RegisterRequest
		loginCtx       *LoginContext
		mockSetup      func(*testing.MockUserRepository, *testing.MockTokenService, *testing.MockSessionService, *testing.MockCacheService)
		expectedError  error
		expectedResult func(*AuthResponse) bool
	}{
		{
			name: "成功注册用户",
			request: RegisterRequest{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "password123",
			},
			loginCtx: &LoginContext{
				IPAddress: "127.0.0.1",
				UserAgent: "test-agent",
				DeviceID:  "device-123",
			},
			mockSetup: func(mockRepo *testing.MockUserRepository, mockToken *testing.MockTokenService, mockSession *testing.MockSessionService, mockCache *testing.MockCacheService) {
				// 模拟用户创建
				user := testing.NewUserBuilder().
					WithName("Test User").
					WithEmail("test@example.com").
					WithPassword("password123").
					Build()
				mockRepo.On("Exists", mock.Anything, mock.MatchedBy(func(filter userDomain.ExistsFilter) bool {
					return filter.Email != nil && *filter.Email == "test@example.com"
				})).Return(false, nil)
				mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)

				// 模拟会话创建
				mockSession.On("CreateSession", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*testing.UserSession")).Return("session-123", nil)

				// 模拟令牌生成
				tokenPair := &testing.TokenPair{
					AccessToken:  "access-token",
					RefreshToken: "refresh-token",
					ExpiresIn:    3600,
					TokenType:    "Bearer",
				}
				mockToken.On("GenerateTokenPair", mock.Anything, mock.AnythingOfType("*user.User"), "session-123").Return(tokenPair, nil)

				// 模拟缓存
				mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*user.User"), mock.AnythingOfType("time.Duration")).Return(nil)
			},
			expectedError: nil,
			expectedResult: func(response *AuthResponse) bool {
				return response.User.Name == "Test User" &&
					response.User.Email == "test@example.com" &&
					response.AccessToken == "access-token" &&
					response.RefreshToken == "refresh-token" &&
					response.SessionID == "session-123"
			},
		},
		{
			name: "邮箱已存在",
			request: RegisterRequest{
				Name:     "Test User",
				Email:    "existing@example.com",
				Password: "password123",
			},
			loginCtx: &LoginContext{
				IPAddress: "127.0.0.1",
				UserAgent: "test-agent",
			},
			mockSetup: func(mockRepo *testing.MockUserRepository, mockToken *testing.MockTokenService, mockSession *testing.MockSessionService, mockCache *testing.MockCacheService) {
				mockRepo.On("Exists", mock.Anything, mock.MatchedBy(func(filter userDomain.ExistsFilter) bool {
					return filter.Email != nil && *filter.Email == "existing@example.com"
				})).Return(true, nil)
			},
			expectedError: errors.ErrEmailExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟服务
			mockRepo := &testing.MockUserRepository{}
			mockToken := &testing.MockTokenService{}
			mockSession := &testing.MockSessionService{}
			mockCache := &testing.MockCacheService{}

			// 设置模拟
			tt.mockSetup(mockRepo, mockToken, mockSession, mockCache)

			// 创建用户领域服务
			userDomainService := userDomain.NewService(mockRepo)

			// 创建应用服务
			service := NewUserService(userDomainService, mockToken, mockSession, mockCache)

			// 执行测试
			result, err := service.Register(context.Background(), tt.request, tt.loginCtx)

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
			mockToken.AssertExpectations(t)
			mockSession.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

func TestUserService_Login(t *testing.T) {
	tests := []struct {
		name           string
		request        LoginRequest
		loginCtx       *LoginContext
		mockSetup      func(*testing.MockUserRepository, *testing.MockTokenService, *testing.MockSessionService, *testing.MockCacheService)
		expectedError  error
		expectedResult func(*AuthResponse) bool
	}{
		{
			name: "成功登录",
			request: LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			loginCtx: &LoginContext{
				IPAddress: "127.0.0.1",
				UserAgent: "test-agent",
				DeviceID:  "device-123",
			},
			mockSetup: func(mockRepo *testing.MockUserRepository, mockToken *testing.MockTokenService, mockSession *testing.MockSessionService, mockCache *testing.MockCacheService) {
				// 模拟用户认证
				user := testing.NewUserBuilder().
					WithEmail("test@example.com").
					WithPassword("password123").
					Build()
				mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)

				// 模拟会话创建
				mockSession.On("CreateSession", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*testing.UserSession")).Return("session-123", nil)

				// 模拟令牌生成
				tokenPair := &testing.TokenPair{
					AccessToken:  "access-token",
					RefreshToken: "refresh-token",
					ExpiresIn:    3600,
					TokenType:    "Bearer",
				}
				mockToken.On("GenerateTokenPair", mock.Anything, mock.AnythingOfType("*user.User"), "session-123").Return(tokenPair, nil)

				// 模拟缓存
				mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*user.User"), mock.AnythingOfType("time.Duration")).Return(nil)
			},
			expectedError: nil,
			expectedResult: func(response *AuthResponse) bool {
				return response.User.Email == "test@example.com" &&
					response.AccessToken == "access-token" &&
					response.RefreshToken == "refresh-token" &&
					response.SessionID == "session-123"
			},
		},
		{
			name: "用户不存在",
			request: LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			loginCtx: &LoginContext{
				IPAddress: "127.0.0.1",
				UserAgent: "test-agent",
			},
			mockSetup: func(mockRepo *testing.MockUserRepository, mockToken *testing.MockTokenService, mockSession *testing.MockSessionService, mockCache *testing.MockCacheService) {
				mockRepo.On("GetByEmail", mock.Anything, "nonexistent@example.com").Return(nil, nil)
			},
			expectedError: errors.ErrUserNotFound,
		},
		{
			name: "密码错误",
			request: LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			loginCtx: &LoginContext{
				IPAddress: "127.0.0.1",
				UserAgent: "test-agent",
			},
			mockSetup: func(mockRepo *testing.MockUserRepository, mockToken *testing.MockTokenService, mockSession *testing.MockSessionService, mockCache *testing.MockCacheService) {
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
			// 创建模拟服务
			mockRepo := &testing.MockUserRepository{}
			mockToken := &testing.MockTokenService{}
			mockSession := &testing.MockSessionService{}
			mockCache := &testing.MockCacheService{}

			// 设置模拟
			tt.mockSetup(mockRepo, mockToken, mockSession, mockCache)

			// 创建用户领域服务
			userDomainService := userDomain.NewService(mockRepo)

			// 创建应用服务
			service := NewUserService(userDomainService, mockToken, mockSession, mockCache)

			// 执行测试
			result, err := service.Login(context.Background(), tt.request, tt.loginCtx)

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
			mockToken.AssertExpectations(t)
			mockSession.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

func TestUserService_Logout(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		token         string
		sessionID     string
		mockSetup     func(*testing.MockTokenService, *testing.MockSessionService, *testing.MockCacheService)
		expectedError error
	}{
		{
			name:      "成功登出",
			userID:    "user-123",
			token:     "access-token",
			sessionID: "session-123",
			mockSetup: func(mockToken *testing.MockTokenService, mockSession *testing.MockSessionService, mockCache *testing.MockCacheService) {
				mockToken.On("BlacklistToken", mock.Anything, "access-token").Return(nil)
				mockSession.On("InvalidateSession", mock.Anything, "session-123").Return(nil)
				mockCache.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:      "令牌黑名单失败",
			userID:    "user-123",
			token:     "access-token",
			sessionID: "session-123",
			mockSetup: func(mockToken *testing.MockTokenService, mockSession *testing.MockSessionService, mockCache *testing.MockCacheService) {
				mockToken.On("BlacklistToken", mock.Anything, "access-token").Return(assert.AnError)
			},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟服务
			mockToken := &testing.MockTokenService{}
			mockSession := &testing.MockSessionService{}
			mockCache := &testing.MockCacheService{}

			// 设置模拟
			tt.mockSetup(mockToken, mockSession, mockCache)

			// 创建应用服务（不需要用户领域服务）
			service := &UserService{
				tokenService:   mockToken,
				sessionService: mockSession,
				cacheService:   mockCache,
			}

			// 执行测试
			err := service.Logout(context.Background(), tt.userID, tt.token, tt.sessionID)

			// 验证结果
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			// 验证模拟调用
			mockToken.AssertExpectations(t)
			mockSession.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

func TestUserService_GetProfile(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockSetup      func(*testing.MockUserRepository)
		expectedError  error
		expectedResult func(*UserResponse) bool
	}{
		{
			name:   "成功获取用户资料",
			userID: "user-123",
			mockSetup: func(mockRepo *testing.MockUserRepository) {
				user := testing.NewUserBuilder().
					WithID("user-123").
					WithName("Test User").
					WithEmail("test@example.com").
					Build()
				mockRepo.On("GetByID", mock.Anything, "user-123").Return(user, nil)
			},
			expectedError: nil,
			expectedResult: func(response *UserResponse) bool {
				return response.ID == "user-123" &&
					response.Name == "Test User" &&
					response.Email == "test@example.com"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟服务
			mockRepo := &testing.MockUserRepository{}

			// 设置模拟
			tt.mockSetup(mockRepo)

			// 创建用户领域服务
			userDomainService := userDomain.NewService(mockRepo)

			// 创建应用服务
			service := NewUserService(userDomainService, nil, nil, nil)

			// 执行测试
			result, err := service.GetProfile(context.Background(), tt.userID)

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

func TestUserService_UpdateProfile(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		request        UpdateProfileRequest
		mockSetup      func(*testing.MockUserRepository, *testing.MockCacheService)
		expectedError  error
		expectedResult func(*UserResponse) bool
	}{
		{
			name:   "成功更新用户资料",
			userID: "user-123",
			request: UpdateProfileRequest{
				Name:  stringPtr("Updated Name"),
				Phone: stringPtr("1234567890"),
			},
			mockSetup: func(mockRepo *testing.MockUserRepository, mockCache *testing.MockCacheService) {
				user := testing.NewUserBuilder().
					WithID("user-123").
					WithName("Original Name").
					WithEmail("test@example.com").
					Build()
				mockRepo.On("GetByID", mock.Anything, "user-123").Return(user, nil)
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)
				mockCache.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(nil)
			},
			expectedError: nil,
			expectedResult: func(response *UserResponse) bool {
				return response.ID == "user-123" && response.Name == "Updated Name"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟服务
			mockRepo := &testing.MockUserRepository{}
			mockCache := &testing.MockCacheService{}

			// 设置模拟
			tt.mockSetup(mockRepo, mockCache)

			// 创建用户领域服务
			userDomainService := userDomain.NewService(mockRepo)

			// 创建应用服务
			service := NewUserService(userDomainService, nil, nil, mockCache)

			// 执行测试
			result, err := service.UpdateProfile(context.Background(), tt.userID, tt.request)

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
			mockCache.AssertExpectations(t)
		})
	}
}

func TestUserService_ChangePassword(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		request       ChangePasswordRequest
		mockSetup     func(*testing.MockUserRepository)
		expectedError error
	}{
		{
			name:   "成功修改密码",
			userID: "user-123",
			request: ChangePasswordRequest{
				OldPassword: "oldpassword",
				NewPassword: "newpassword123",
			},
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
			name:   "旧密码错误",
			userID: "user-123",
			request: ChangePasswordRequest{
				OldPassword: "wrongpassword",
				NewPassword: "newpassword123",
			},
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
			// 创建模拟服务
			mockRepo := &testing.MockUserRepository{}

			// 设置模拟
			tt.mockSetup(mockRepo)

			// 创建用户领域服务
			userDomainService := userDomain.NewService(mockRepo)

			// 创建应用服务
			service := NewUserService(userDomainService, nil, nil, nil)

			// 执行测试
			err := service.ChangePassword(context.Background(), tt.userID, tt.request)

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

func TestUserService_RefreshToken(t *testing.T) {
	tests := []struct {
		name           string
		request        RefreshTokenRequest
		mockSetup      func(*testing.MockUserRepository, *testing.MockTokenService, *testing.MockSessionService, *testing.MockCacheService)
		expectedError  error
		expectedResult func(*AuthResponse) bool
	}{
		{
			name: "成功刷新令牌",
			request: RefreshTokenRequest{
				RefreshToken: "refresh-token",
			},
			mockSetup: func(mockRepo *testing.MockUserRepository, mockToken *testing.MockTokenService, mockSession *testing.MockSessionService, mockCache *testing.MockCacheService) {
				// 模拟令牌验证
				claims := &testing.TokenClaims{
					UserID:    "user-123",
					Email:     "test@example.com",
					IsAdmin:   false,
					SessionID: "session-123",
					TokenType: "refresh",
					ExpiresAt: time.Now().Add(time.Hour).Unix(),
				}
				mockToken.On("ValidateToken", mock.Anything, "refresh-token").Return(claims, nil)

				// 模拟获取用户
				user := testing.NewUserBuilder().
					WithID("user-123").
					WithEmail("test@example.com").
					Build()
				mockRepo.On("GetByID", mock.Anything, "user-123").Return(user, nil)

				// 模拟更新会话活动
				mockSession.On("UpdateSessionActivity", mock.Anything, "session-123").Return(nil)

				// 模拟刷新令牌
				tokenPair := &testing.TokenPair{
					AccessToken:  "new-access-token",
					RefreshToken: "new-refresh-token",
					ExpiresIn:    3600,
					TokenType:    "Bearer",
				}
				mockToken.On("RefreshToken", mock.Anything, "refresh-token", mock.AnythingOfType("*user.User")).Return(tokenPair, nil)

				// 模拟缓存
				mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*user.User"), mock.AnythingOfType("time.Duration")).Return(nil)
			},
			expectedError: nil,
			expectedResult: func(response *AuthResponse) bool {
				return response.User.ID == "user-123" &&
					response.AccessToken == "new-access-token" &&
					response.RefreshToken == "new-refresh-token" &&
					response.SessionID == "session-123"
			},
		},
		{
			name: "无效的刷新令牌",
			request: RefreshTokenRequest{
				RefreshToken: "invalid-token",
			},
			mockSetup: func(mockRepo *testing.MockUserRepository, mockToken *testing.MockTokenService, mockSession *testing.MockSessionService, mockCache *testing.MockCacheService) {
				mockToken.On("ValidateToken", mock.Anything, "invalid-token").Return(nil, errors.ErrInvalidToken)
			},
			expectedError: errors.ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟服务
			mockRepo := &testing.MockUserRepository{}
			mockToken := &testing.MockTokenService{}
			mockSession := &testing.MockSessionService{}
			mockCache := &testing.MockCacheService{}

			// 设置模拟
			tt.mockSetup(mockRepo, mockToken, mockSession, mockCache)

			// 创建用户领域服务
			userDomainService := userDomain.NewService(mockRepo)

			// 创建应用服务
			service := NewUserService(userDomainService, mockToken, mockSession, mockCache)

			// 执行测试
			result, err := service.RefreshToken(context.Background(), tt.request)

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
			mockToken.AssertExpectations(t)
			mockSession.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}
