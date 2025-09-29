package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"teable-go-backend/internal/application"
	userDomain "teable-go-backend/internal/domain/user"
	"teable-go-backend/internal/testing"
)

func TestUserHandler_Register(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*testing.MockUserRepository, *testing.MockTokenService, *testing.MockSessionService, *testing.MockCacheService)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "成功注册用户",
			requestBody: map[string]interface{}{
				"name":     "Test User",
				"email":    "test@example.com",
				"password": "password123",
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
			expectedStatus: http.StatusCreated,
		},
		{
			name: "邮箱已存在",
			requestBody: map[string]interface{}{
				"name":     "Test User",
				"email":    "existing@example.com",
				"password": "password123",
			},
			mockSetup: func(mockRepo *testing.MockUserRepository, mockToken *testing.MockTokenService, mockSession *testing.MockSessionService, mockCache *testing.MockCacheService) {
				mockRepo.On("Exists", mock.Anything, mock.MatchedBy(func(filter userDomain.ExistsFilter) bool {
					return filter.Email != nil && *filter.Email == "existing@example.com"
				})).Return(true, nil)
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "邮箱已存在",
		},
		{
			name: "无效请求体",
			requestBody: map[string]interface{}{
				"name": "", // 空名称
			},
			mockSetup: func(mockRepo *testing.MockUserRepository, mockToken *testing.MockTokenService, mockSession *testing.MockSessionService, mockCache *testing.MockCacheService) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "请求参数错误",
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
			userService := application.NewUserService(userDomainService, mockToken, mockSession, mockCache)

			// 创建处理器
			handler := NewUserHandler(userService)

			// 设置Gin
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.POST("/register", handler.Register)

			// 准备请求
			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// 创建响应记录器
			w := httptest.NewRecorder()

			// 执行请求
			router.ServeHTTP(w, req)

			// 验证响应
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotNil(t, response["user"])
				assert.NotNil(t, response["access_token"])
			}

			// 验证模拟调用
			mockRepo.AssertExpectations(t)
			mockToken.AssertExpectations(t)
			mockSession.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*testing.MockUserRepository, *testing.MockTokenService, *testing.MockSessionService, *testing.MockCacheService)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "成功登录",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "password123",
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
			expectedStatus: http.StatusOK,
		},
		{
			name: "用户不存在",
			requestBody: map[string]interface{}{
				"email":    "nonexistent@example.com",
				"password": "password123",
			},
			mockSetup: func(mockRepo *testing.MockUserRepository, mockToken *testing.MockTokenService, mockSession *testing.MockSessionService, mockCache *testing.MockCacheService) {
				mockRepo.On("GetByEmail", mock.Anything, "nonexistent@example.com").Return(nil, nil)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "用户不存在",
		},
		{
			name: "密码错误",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "wrongpassword",
			},
			mockSetup: func(mockRepo *testing.MockUserRepository, mockToken *testing.MockTokenService, mockSession *testing.MockSessionService, mockCache *testing.MockCacheService) {
				user := testing.NewUserBuilder().
					WithEmail("test@example.com").
					WithPassword("password123").
					Build()
				mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "无效的凭据",
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
			userService := application.NewUserService(userDomainService, mockToken, mockSession, mockCache)

			// 创建处理器
			handler := NewUserHandler(userService)

			// 设置Gin
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.POST("/login", handler.Login)

			// 准备请求
			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// 创建响应记录器
			w := httptest.NewRecorder()

			// 执行请求
			router.ServeHTTP(w, req)

			// 验证响应
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotNil(t, response["user"])
				assert.NotNil(t, response["access_token"])
			}

			// 验证模拟调用
			mockRepo.AssertExpectations(t)
			mockToken.AssertExpectations(t)
			mockSession.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetProfile(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockSetup      func(*testing.MockUserRepository)
		expectedStatus int
		expectedError  string
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
			expectedStatus: http.StatusOK,
		},
		{
			name:   "用户不存在",
			userID: "non-existent",
			mockSetup: func(mockRepo *testing.MockUserRepository) {
				mockRepo.On("GetByID", mock.Anything, "non-existent").Return(nil, nil)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "用户不存在",
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
			userService := application.NewUserService(userDomainService, nil, nil, nil)

			// 创建处理器
			handler := NewUserHandler(userService)

			// 设置Gin
			gin.SetMode(gin.TestMode)
			router := gin.New()

			// 添加中间件来设置用户ID
			router.Use(func(c *gin.Context) {
				c.Set("user_id", tt.userID)
				c.Next()
			})

			router.GET("/profile", handler.GetProfile)

			// 准备请求
			req, _ := http.NewRequest("GET", "/profile", nil)

			// 创建响应记录器
			w := httptest.NewRecorder()

			// 执行请求
			router.ServeHTTP(w, req)

			// 验证响应
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "Test User", response["name"])
				assert.Equal(t, "test@example.com", response["email"])
			}

			// 验证模拟调用
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserHandler_UpdateProfile(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		requestBody    interface{}
		mockSetup      func(*testing.MockUserRepository, *testing.MockCacheService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "成功更新用户资料",
			userID: "user-123",
			requestBody: map[string]interface{}{
				"name":  "Updated User",
				"phone": "1234567890",
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
			expectedStatus: http.StatusOK,
		},
		{
			name:   "无效请求体",
			userID: "user-123",
			requestBody: map[string]interface{}{
				"name": "", // 空名称
			},
			mockSetup:      func(mockRepo *testing.MockUserRepository, mockCache *testing.MockCacheService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "请求参数错误",
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
			userService := application.NewUserService(userDomainService, nil, nil, mockCache)

			// 创建处理器
			handler := NewUserHandler(userService)

			// 设置Gin
			gin.SetMode(gin.TestMode)
			router := gin.New()

			// 添加中间件来设置用户ID
			router.Use(func(c *gin.Context) {
				c.Set("user_id", tt.userID)
				c.Next()
			})

			router.PUT("/profile", handler.UpdateProfile)

			// 准备请求
			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("PUT", "/profile", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// 创建响应记录器
			w := httptest.NewRecorder()

			// 执行请求
			router.ServeHTTP(w, req)

			// 验证响应
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "Updated User", response["name"])
				assert.Equal(t, "1234567890", response["phone"])
			}

			// 验证模拟调用
			mockRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

func TestUserHandler_ChangePassword(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		requestBody    interface{}
		mockSetup      func(*testing.MockUserRepository)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "成功修改密码",
			userID: "user-123",
			requestBody: map[string]interface{}{
				"old_password": "oldpassword",
				"new_password": "newpassword123",
			},
			mockSetup: func(mockRepo *testing.MockUserRepository) {
				user := testing.NewUserBuilder().
					WithID("user-123").
					WithPassword("oldpassword").
					Build()
				mockRepo.On("GetByID", mock.Anything, "user-123").Return(user, nil)
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "旧密码错误",
			userID: "user-123",
			requestBody: map[string]interface{}{
				"old_password": "wrongpassword",
				"new_password": "newpassword123",
			},
			mockSetup: func(mockRepo *testing.MockUserRepository) {
				user := testing.NewUserBuilder().
					WithID("user-123").
					WithPassword("oldpassword").
					Build()
				mockRepo.On("GetByID", mock.Anything, "user-123").Return(user, nil)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "无效的凭据",
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
			userService := application.NewUserService(userDomainService, nil, nil, nil)

			// 创建处理器
			handler := NewUserHandler(userService)

			// 设置Gin
			gin.SetMode(gin.TestMode)
			router := gin.New()

			// 添加中间件来设置用户ID
			router.Use(func(c *gin.Context) {
				c.Set("user_id", tt.userID)
				c.Next()
			})

			router.POST("/change-password", handler.ChangePassword)

			// 准备请求
			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/change-password", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// 创建响应记录器
			w := httptest.NewRecorder()

			// 执行请求
			router.ServeHTTP(w, req)

			// 验证响应
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "密码修改成功", response["message"])
			}

			// 验证模拟调用
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserHandler_ListUsers(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func(*testing.MockUserRepository)
		expectedStatus int
		expectedError  string
	}{
		{
			name:        "成功获取用户列表",
			queryParams: "?page=1&page_size=10",
			mockSetup: func(mockRepo *testing.MockUserRepository) {
				users := []*userDomain.User{
					testing.NewUserBuilder().WithID("user-1").WithName("User 1").Build(),
					testing.NewUserBuilder().WithID("user-2").WithName("User 2").Build(),
				}
				mockRepo.On("List", mock.Anything, mock.AnythingOfType("user.ListFilter")).Return(users, nil)
				mockRepo.On("Count", mock.Anything, mock.AnythingOfType("user.CountFilter")).Return(int64(2), nil)
			},
			expectedStatus: http.StatusOK,
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
			userService := application.NewUserService(userDomainService, nil, nil, nil)

			// 创建处理器
			handler := NewUserHandler(userService)

			// 设置Gin
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.GET("/users", handler.ListUsers)

			// 准备请求
			req, _ := http.NewRequest("GET", "/users"+tt.queryParams, nil)

			// 创建响应记录器
			w := httptest.NewRecorder()

			// 执行请求
			router.ServeHTTP(w, req)

			// 验证响应
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotNil(t, response["users"])
				assert.Equal(t, float64(2), response["total"])
			}

			// 验证模拟调用
			mockRepo.AssertExpectations(t)
		})
	}
}
