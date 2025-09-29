package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockRepository Mock仓储实现
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, user *User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) Update(ctx context.Context, user *User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) GetByID(ctx context.Context, id string) (*User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) List(ctx context.Context, offset, limit int) ([]*User, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*User), args.Error(1)
}

func (m *MockRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// UserServiceTestSuite 用户服务测试套件
type UserServiceTestSuite struct {
	suite.Suite
	service  Service
	mockRepo *MockRepository
	ctx      context.Context
}

func (suite *UserServiceTestSuite) SetupTest() {
	suite.mockRepo = new(MockRepository)
	suite.service = NewService(suite.mockRepo)
	suite.ctx = context.Background()
}

func (suite *UserServiceTestSuite) TearDownTest() {
	suite.mockRepo.AssertExpectations(suite.T())
}

// TestCreateUser 测试创建用户
func (suite *UserServiceTestSuite) TestCreateUser() {
	tests := []struct {
		name      string
		inputName string
		inputEmail string
		setupMock func()
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "成功创建用户",
			inputName: "Test User",
			inputEmail: "test@example.com",
			setupMock: func() {
				suite.mockRepo.On("GetByEmail", suite.ctx, "test@example.com").Return(nil, errors.New("not found"))
				suite.mockRepo.On("Create", suite.ctx, mock.AnythingOfType("*user.User")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "邮箱已存在",
			inputName: "Test User",
			inputEmail: "existing@example.com",
			setupMock: func() {
				existingUser := &User{Email: "existing@example.com"}
				suite.mockRepo.On("GetByEmail", suite.ctx, "existing@example.com").Return(existingUser, nil)
			},
			wantErr: true,
			errMsg:  "email already exists",
		},
		{
			name:      "无效邮箱格式",
			inputName: "Test User",
			inputEmail: "invalid-email",
			setupMock: func() {},
			wantErr:   true,
			errMsg:    "invalid email format",
		},
		{
			name:      "空用户名",
			inputName: "",
			inputEmail: "test@example.com",
			setupMock: func() {},
			wantErr:   true,
			errMsg:    "user name cannot be empty",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tt.setupMock()
			
			user, err := suite.service.CreateUser(suite.ctx, tt.inputName, tt.inputEmail)
			
			if tt.wantErr {
				suite.Error(err)
				if tt.errMsg != "" {
					suite.Contains(err.Error(), tt.errMsg)
				}
				suite.Nil(user)
			} else {
				suite.NoError(err)
				suite.NotNil(user)
				suite.Equal(tt.inputName, user.Name)
				suite.Equal(tt.inputEmail, user.Email)
				suite.NotEmpty(user.ID)
			}
			
			suite.mockRepo.AssertExpectations(suite.T())
			suite.mockRepo.ExpectedCalls = nil
		})
	}
}

// TestUpdateUser 测试更新用户
func (suite *UserServiceTestSuite) TestUpdateUser() {
	userID := "user123"
	existingUser := &User{
		ID:          userID,
		Name:        "Old Name",
		Email:       "old@example.com",
		CreatedTime: time.Now().Add(-24 * time.Hour),
	}

	tests := []struct {
		name      string
		userID    string
		updates   map[string]interface{}
		setupMock func()
		wantErr   bool
		errMsg    string
	}{
		{
			name:   "成功更新用户名",
			userID: userID,
			updates: map[string]interface{}{
				"name": "New Name",
			},
			setupMock: func() {
				suite.mockRepo.On("GetByID", suite.ctx, userID).Return(existingUser, nil)
				suite.mockRepo.On("Update", suite.ctx, mock.AnythingOfType("*user.User")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "用户不存在",
			userID: "nonexistent",
			updates: map[string]interface{}{
				"name": "New Name",
			},
			setupMock: func() {
				suite.mockRepo.On("GetByID", suite.ctx, "nonexistent").Return(nil, errors.New("not found"))
			},
			wantErr: true,
			errMsg:  "user not found",
		},
		{
			name:   "用户已删除",
			userID: userID,
			updates: map[string]interface{}{
				"name": "New Name",
			},
			setupMock: func() {
				deletedTime := time.Now()
				deletedUser := &User{
					ID:          userID,
					Name:        "Deleted User",
					Email:       "deleted@example.com",
					DeletedTime: &deletedTime,
				}
				suite.mockRepo.On("GetByID", suite.ctx, userID).Return(deletedUser, nil)
			},
			wantErr: true,
			errMsg:  "user is deleted",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tt.setupMock()
			
			user, err := suite.service.UpdateUser(suite.ctx, tt.userID, tt.updates)
			
			if tt.wantErr {
				suite.Error(err)
				if tt.errMsg != "" {
					suite.Contains(err.Error(), tt.errMsg)
				}
				suite.Nil(user)
			} else {
				suite.NoError(err)
				suite.NotNil(user)
			}
			
			suite.mockRepo.AssertExpectations(suite.T())
			suite.mockRepo.ExpectedCalls = nil
		})
	}
}

// TestDeleteUser 测试删除用户
func (suite *UserServiceTestSuite) TestDeleteUser() {
	userID := "user123"
	
	tests := []struct {
		name      string
		userID    string
		setupMock func()
		wantErr   bool
		errMsg    string
	}{
		{
			name:   "成功删除用户",
			userID: userID,
			setupMock: func() {
				user := &User{ID: userID, Name: "Test User"}
				suite.mockRepo.On("GetByID", suite.ctx, userID).Return(user, nil)
				suite.mockRepo.On("Update", suite.ctx, mock.AnythingOfType("*user.User")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "系统用户不能删除",
			userID: userID,
			setupMock: func() {
				systemUser := &User{ID: userID, Name: "System", IsSystem: true}
				suite.mockRepo.On("GetByID", suite.ctx, userID).Return(systemUser, nil)
			},
			wantErr: true,
			errMsg:  "system user cannot be deleted",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tt.setupMock()
			
			err := suite.service.DeleteUser(suite.ctx, tt.userID)
			
			if tt.wantErr {
				suite.Error(err)
				if tt.errMsg != "" {
					suite.Contains(err.Error(), tt.errMsg)
				}
			} else {
				suite.NoError(err)
			}
			
			suite.mockRepo.AssertExpectations(suite.T())
			suite.mockRepo.ExpectedCalls = nil
		})
	}
}

// TestConcurrentUserCreation 测试并发创建用户
func (suite *UserServiceTestSuite) TestConcurrentUserCreation() {
	// 设置mock期望
	suite.mockRepo.On("GetByEmail", suite.ctx, mock.AnythingOfType("string")).Return(nil, errors.New("not found"))
	suite.mockRepo.On("Create", suite.ctx, mock.AnythingOfType("*user.User")).Return(nil)

	// 并发创建用户
	concurrency := 10
	done := make(chan bool, concurrency)
	errors := make([]error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(index int) {
			defer func() { done <- true }()
			
			name := assert.Sprintf("User %d", index)
			email := assert.Sprintf("user%d@example.com", index)
			
			_, err := suite.service.CreateUser(suite.ctx, name, email)
			errors[index] = err
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < concurrency; i++ {
		<-done
	}

	// 验证所有创建都成功
	for i, err := range errors {
		suite.NoError(err, "User creation %d should succeed", i)
	}
}

// BenchmarkCreateUser 性能测试
func BenchmarkCreateUser(b *testing.B) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetByEmail", ctx, mock.AnythingOfType("string")).Return(nil, errors.New("not found"))
	mockRepo.On("Create", ctx, mock.AnythingOfType("*user.User")).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		name := assert.Sprintf("User %d", i)
		email := assert.Sprintf("user%d@example.com", i)
		_, _ = service.CreateUser(ctx, name, email)
	}
}

// TestUserServiceTestSuite 运行测试套件
func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}