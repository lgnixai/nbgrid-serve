package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"teable-go-backend/internal/infrastructure/cache"
)

// UserFlowTestSuite 用户流程集成测试套件
type UserFlowTestSuite struct {
	IntegrationTestSuite
}

// TestCompleteUserRegistrationFlow 测试完整的用户注册流程
func (s *UserFlowTestSuite) TestCompleteUserRegistrationFlow() {
	// 1. 注册新用户
	userService := s.Container().UserAppService()
	authService := s.Container().AuthService()
	
	// 测试数据
	userName := "Test User"
	userEmail := "testuser@example.com"
	userPassword := "Test123!@#"
	
	// 执行注册
	user, err := userService.Register(s.Context(), userName, userEmail, userPassword)
	s.Require().NoError(err)
	s.Require().NotNil(user)
	
	// 验证用户数据
	s.Equal(userName, user.Name)
	s.Equal(userEmail, user.Email)
	s.NotEmpty(user.ID)
	s.False(user.IsAdmin)
	s.False(user.IsSystem)
	
	// 2. 验证用户已存储到数据库
	var dbUser struct {
		ID    string
		Name  string
		Email string
	}
	err = s.DB().Table("users").Where("id = ?", user.ID).First(&dbUser).Error
	s.Require().NoError(err)
	s.Equal(user.ID, dbUser.ID)
	s.Equal(userName, dbUser.Name)
	s.Equal(userEmail, dbUser.Email)
	
	// 3. 测试登录
	tokens, err := userService.Login(s.Context(), userEmail, userPassword)
	s.Require().NoError(err)
	s.Require().NotNil(tokens)
	s.NotEmpty(tokens.AccessToken)
	s.NotEmpty(tokens.RefreshToken)
	
	// 4. 验证会话已创建
	sessionKey := cache.BuildCacheKey("session:", user.ID)
	s.AssertCacheExists(sessionKey)
	
	// 5. 验证token可以正常使用
	claims, err := authService.ValidateToken(tokens.AccessToken)
	s.Require().NoError(err)
	s.Equal(user.ID, claims.UserID)
	s.Equal(userEmail, claims.Email)
	s.Equal("access", claims.TokenType)
	
	// 6. 测试获取用户信息
	profile, err := userService.GetUserByID(s.Context(), user.ID)
	s.Require().NoError(err)
	s.Equal(user.ID, profile.ID)
	
	// 7. 测试更新用户信息
	newName := "Updated User"
	newPhone := "+1234567890"
	updates := map[string]interface{}{
		"name":  newName,
		"phone": &newPhone,
	}
	
	updatedUser, err := userService.UpdateUser(s.Context(), user.ID, updates)
	s.Require().NoError(err)
	s.Equal(newName, updatedUser.Name)
	s.Equal(newPhone, *updatedUser.Phone)
	
	// 8. 测试修改密码
	newPassword := "NewTest123!@#"
	err = userService.ChangePassword(s.Context(), user.ID, userPassword, newPassword)
	s.Require().NoError(err)
	
	// 9. 验证新密码登录
	newTokens, err := userService.Login(s.Context(), userEmail, newPassword)
	s.Require().NoError(err)
	s.NotEmpty(newTokens.AccessToken)
	
	// 10. 测试刷新令牌
	refreshedTokens, err := userService.RefreshToken(s.Context(), tokens.RefreshToken)
	s.Require().NoError(err)
	s.NotEmpty(refreshedTokens.AccessToken)
	s.NotEqual(tokens.AccessToken, refreshedTokens.AccessToken)
	
	// 11. 测试登出
	err = userService.Logout(s.Context(), user.ID, refreshedTokens.AccessToken)
	s.Require().NoError(err)
	
	// 12. 验证token已失效
	_, err = authService.ValidateToken(refreshedTokens.AccessToken)
	s.Error(err)
}

// TestConcurrentUserRegistration 测试并发用户注册
func (s *UserFlowTestSuite) TestConcurrentUserRegistration() {
	userService := s.Container().UserAppService()
	
	// 并发注册10个用户
	concurrency := 10
	errors := make([]error, concurrency)
	users := make([]interface{}, concurrency)
	
	s.RunConcurrent(concurrency, func(index int) {
		userName := s.T().Name() + "_User_" + string(rune(index))
		userEmail := s.T().Name() + "_user" + string(rune(index)) + "@example.com"
		userPassword := "Test123!@#"
		
		user, err := userService.Register(s.Context(), userName, userEmail, userPassword)
		errors[index] = err
		users[index] = user
	})
	
	// 验证所有注册都成功
	successCount := 0
	for i, err := range errors {
		if err == nil {
			successCount++
			s.NotNil(users[i])
		}
	}
	
	s.Equal(concurrency, successCount, "All concurrent registrations should succeed")
}

// TestUserPermissionFlow 测试用户权限流程
func (s *UserFlowTestSuite) TestUserPermissionFlow() {
	// 1. 创建管理员用户
	adminID := s.CreateTestUser("Admin User", "admin@example.com", "Admin123!@#")
	
	// 2. 提升为管理员
	userService := s.Container().UserAppService()
	err := userService.PromoteToAdmin(s.Context(), adminID, "system")
	s.Require().NoError(err)
	
	// 3. 创建普通用户
	userID := s.CreateTestUser("Normal User", "user@example.com", "User123!@#")
	
	// 4. 创建空间
	spaceID := s.CreateTestSpace(adminID, "Test Space")
	
	// 5. 验证管理员可以访问空间
	spaceService := s.Container().SpaceService()
	space, err := spaceService.GetSpace(s.Context(), adminID, spaceID)
	s.Require().NoError(err)
	s.Equal(spaceID, space.ID)
	
	// 6. 验证普通用户无法访问空间
	_, err = spaceService.GetSpace(s.Context(), userID, spaceID)
	s.Error(err)
	
	// 7. 添加普通用户为协作者
	err = spaceService.AddCollaborator(s.Context(), adminID, spaceID, userID, "editor")
	s.Require().NoError(err)
	
	// 8. 验证普通用户现在可以访问空间
	space, err = spaceService.GetSpace(s.Context(), userID, spaceID)
	s.Require().NoError(err)
	s.Equal(spaceID, space.ID)
}

// TestUserSessionManagement 测试用户会话管理
func (s *UserFlowTestSuite) TestUserSessionManagement() {
	userService := s.Container().UserAppService()
	
	// 1. 创建用户并登录
	userID := s.CreateTestUser("Session User", "session@example.com", "Session123!@#")
	tokens1, err := userService.Login(s.Context(), "session@example.com", "Session123!@#")
	s.Require().NoError(err)
	
	// 2. 从不同设备登录（模拟）
	time.Sleep(100 * time.Millisecond) // 确保token不同
	tokens2, err := userService.Login(s.Context(), "session@example.com", "Session123!@#")
	s.Require().NoError(err)
	s.NotEqual(tokens1.AccessToken, tokens2.AccessToken)
	
	// 3. 验证两个会话都有效
	authService := s.Container().AuthService()
	_, err = authService.ValidateToken(tokens1.AccessToken)
	s.NoError(err)
	_, err = authService.ValidateToken(tokens2.AccessToken)
	s.NoError(err)
	
	// 4. 登出第一个会话
	err = userService.Logout(s.Context(), userID, tokens1.AccessToken)
	s.Require().NoError(err)
	
	// 5. 验证第一个会话已失效，第二个仍有效
	_, err = authService.ValidateToken(tokens1.AccessToken)
	s.Error(err)
	_, err = authService.ValidateToken(tokens2.AccessToken)
	s.NoError(err)
}

// TestUserDataConsistency 测试用户数据一致性
func (s *UserFlowTestSuite) TestUserDataConsistency() {
	userService := s.Container().UserAppService()
	
	// 1. 创建用户
	userName := "Consistency User"
	userEmail := "consistency@example.com"
	user, err := userService.Register(s.Context(), userName, userEmail, "Test123!@#")
	s.Require().NoError(err)
	
	// 2. 并发更新用户信息
	concurrency := 5
	s.RunConcurrent(concurrency, func(index int) {
		updates := map[string]interface{}{
			"name": s.T().Name() + "_Updated_" + string(rune(index)),
		}
		_, _ = userService.UpdateUser(s.Context(), user.ID, updates)
	})
	
	// 3. 等待所有更新完成
	s.WaitForAsync(100 * time.Millisecond)
	
	// 4. 验证最终数据一致性
	finalUser, err := userService.GetUserByID(s.Context(), user.ID)
	s.Require().NoError(err)
	s.NotEqual(userName, finalUser.Name) // 名字应该已更新
	s.Equal(userEmail, finalUser.Email)  // 邮箱应该保持不变
	
	// 5. 验证数据库和缓存一致性
	var dbUser struct {
		ID    string
		Name  string
		Email string
	}
	err = s.DB().Table("users").Where("id = ?", user.ID).First(&dbUser).Error
	s.Require().NoError(err)
	s.Equal(finalUser.Name, dbUser.Name)
	s.Equal(finalUser.Email, dbUser.Email)
}

// TestUserDeletionCascade 测试用户删除级联
func (s *UserFlowTestSuite) TestUserDeletionCascade() {
	// 1. 创建用户
	userID := s.CreateTestUser("Delete User", "delete@example.com", "Delete123!@#")
	
	// 2. 创建用户相关数据
	spaceID := s.CreateTestSpace(userID, "User Space")
	baseID := s.CreateTestBase(spaceID, "User Base")
	tableID := s.CreateTestTable(baseID, "User Table")
	
	// 3. 软删除用户
	userService := s.Container().UserAppService()
	err := userService.DeleteUser(s.Context(), userID)
	s.Require().NoError(err)
	
	// 4. 验证用户已标记为删除
	user, err := userService.GetUserByID(s.Context(), userID)
	s.Require().NoError(err)
	s.NotNil(user.DeletedTime)
	
	// 5. 验证无法使用已删除用户登录
	_, err = userService.Login(s.Context(), "delete@example.com", "Delete123!@#")
	s.Error(err)
	
	// 6. 验证相关数据的访问权限
	spaceService := s.Container().SpaceService()
	_, err = spaceService.GetSpace(s.Context(), userID, spaceID)
	s.Error(err) // 已删除用户无法访问其空间
}

// TestUserActivityTracking 测试用户活动跟踪
func (s *UserFlowTestSuite) TestUserActivityTracking() {
	// 1. 创建用户
	userID := s.CreateTestUser("Activity User", "activity@example.com", "Activity123!@#")
	
	// 2. 执行各种操作
	userService := s.Container().UserAppService()
	
	// 登录
	_, err := userService.Login(s.Context(), "activity@example.com", "Activity123!@#")
	s.Require().NoError(err)
	
	// 更新信息
	_, err = userService.UpdateUser(s.Context(), userID, map[string]interface{}{
		"name": "Updated Activity User",
	})
	s.Require().NoError(err)
	
	// 创建空间
	spaceID := s.CreateTestSpace(userID, "Activity Space")
	
	// 3. 获取用户活动记录
	// 注意：这需要实现活动跟踪功能
	user, err := userService.GetUserByID(s.Context(), userID)
	s.Require().NoError(err)
	s.NotNil(user.LastSignTime)
	s.NotNil(user.LastModifiedTime)
	
	// 4. 验证活动时间戳
	s.True(user.LastSignTime.After(user.CreatedTime))
	s.True(user.LastModifiedTime.After(user.CreatedTime))
}

// TestUserFlowTestSuite 运行用户流程测试套件
func TestUserFlowTestSuite(t *testing.T) {
	suite.Run(t, new(UserFlowTestSuite))
}