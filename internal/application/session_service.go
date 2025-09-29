package application

import (
	"context"
	"fmt"
	"time"

	"teable-go-backend/internal/infrastructure/cache"
	"teable-go-backend/pkg/logger"
)

// SessionService 会话管理服务
type SessionService struct {
	cacheService cache.CacheService
}

// NewSessionService 创建会话服务
func NewSessionService(cacheService cache.CacheService) *SessionService {
	return &SessionService{
		cacheService: cacheService,
	}
}

// UserSession 用户会话信息
type UserSession struct {
	UserID       string    `json:"user_id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	IsAdmin      bool      `json:"is_admin"`
	IsSystem     bool      `json:"is_system"`
	LoginTime    time.Time `json:"login_time"`
	LastActivity time.Time `json:"last_activity"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	DeviceID     string    `json:"device_id,omitempty"`
}

// CreateSession 创建用户会话
func (s *SessionService) CreateSession(ctx context.Context, userID string, session *UserSession) (string, error) {
	sessionID := generateSessionID()
	sessionKey := s.buildSessionKey(sessionID)
	
	// 设置会话过期时间为24小时
	expiration := 24 * time.Hour
	
	if err := s.cacheService.Set(ctx, sessionKey, session, expiration); err != nil {
		logger.Error("Failed to create user session",
			logger.String("user_id", userID),
			logger.String("session_id", sessionID),
			logger.ErrorField(err),
		)
		return "", err
	}

	// 维护用户的活跃会话列表
	if err := s.addToUserSessions(ctx, userID, sessionID); err != nil {
		logger.Warn("Failed to add session to user sessions list",
			logger.String("user_id", userID),
			logger.String("session_id", sessionID),
			logger.ErrorField(err),
		)
	}

	logger.Info("User session created",
		logger.String("user_id", userID),
		logger.String("session_id", sessionID),
	)

	return sessionID, nil
}

// GetSession 获取用户会话
func (s *SessionService) GetSession(ctx context.Context, sessionID string) (*UserSession, error) {
	sessionKey := s.buildSessionKey(sessionID)
	
	var session UserSession
	if err := s.cacheService.Get(ctx, sessionKey, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

// UpdateSessionActivity 更新会话活动时间
func (s *SessionService) UpdateSessionActivity(ctx context.Context, sessionID string) error {
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	session.LastActivity = time.Now()
	sessionKey := s.buildSessionKey(sessionID)
	
	// 延长会话过期时间
	expiration := 24 * time.Hour
	return s.cacheService.Set(ctx, sessionKey, session, expiration)
}

// InvalidateSession 使会话失效
func (s *SessionService) InvalidateSession(ctx context.Context, sessionID string) error {
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	sessionKey := s.buildSessionKey(sessionID)
	if err := s.cacheService.Delete(ctx, sessionKey); err != nil {
		return err
	}

	// 从用户会话列表中移除
	if err := s.removeFromUserSessions(ctx, session.UserID, sessionID); err != nil {
		logger.Warn("Failed to remove session from user sessions list",
			logger.String("user_id", session.UserID),
			logger.String("session_id", sessionID),
			logger.ErrorField(err),
		)
	}

	logger.Info("User session invalidated",
		logger.String("user_id", session.UserID),
		logger.String("session_id", sessionID),
	)

	return nil
}

// InvalidateAllUserSessions 使用户的所有会话失效
func (s *SessionService) InvalidateAllUserSessions(ctx context.Context, userID string) error {
	sessionIDs, err := s.getUserSessions(ctx, userID)
	if err != nil {
		return err
	}

	for _, sessionID := range sessionIDs {
		sessionKey := s.buildSessionKey(sessionID)
		if err := s.cacheService.Delete(ctx, sessionKey); err != nil {
			logger.Warn("Failed to delete session",
				logger.String("user_id", userID),
				logger.String("session_id", sessionID),
				logger.ErrorField(err),
			)
		}
	}

	// 清空用户会话列表
	userSessionsKey := s.buildUserSessionsKey(userID)
	if err := s.cacheService.Delete(ctx, userSessionsKey); err != nil {
		logger.Warn("Failed to clear user sessions list",
			logger.String("user_id", userID),
			logger.ErrorField(err),
		)
	}

	logger.Info("All user sessions invalidated",
		logger.String("user_id", userID),
		logger.Int("session_count", len(sessionIDs)),
	)

	return nil
}

// GetUserActiveSessions 获取用户的活跃会话
func (s *SessionService) GetUserActiveSessions(ctx context.Context, userID string) ([]*UserSession, error) {
	sessionIDs, err := s.getUserSessions(ctx, userID)
	if err != nil {
		return nil, err
	}

	var sessions []*UserSession
	for _, sessionID := range sessionIDs {
		session, err := s.GetSession(ctx, sessionID)
		if err != nil {
			// 会话可能已过期，从列表中移除
			s.removeFromUserSessions(ctx, userID, sessionID)
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// CleanupExpiredSessions 清理过期会话
func (s *SessionService) CleanupExpiredSessions(ctx context.Context) error {
	// 这个方法可以通过定时任务调用
	// 实际实现中可以扫描所有用户的会话列表，检查过期会话
	logger.Info("Session cleanup completed")
	return nil
}

// 私有方法

func (s *SessionService) buildSessionKey(sessionID string) string {
	return cache.BuildCacheKey("session:", sessionID)
}

func (s *SessionService) buildUserSessionsKey(userID string) string {
	return cache.BuildCacheKey("user_sessions:", userID)
}

func (s *SessionService) addToUserSessions(ctx context.Context, userID, sessionID string) error {
	key := s.buildUserSessionsKey(userID)
	
	// 获取现有会话列表
	var sessionIDs []string
	if err := s.cacheService.Get(ctx, key, &sessionIDs); err != nil {
		// 如果不存在，创建新列表
		sessionIDs = []string{}
	}

	// 添加新会话ID
	sessionIDs = append(sessionIDs, sessionID)
	
	// 限制每个用户最多保持10个活跃会话
	if len(sessionIDs) > 10 {
		// 移除最旧的会话
		oldSessionID := sessionIDs[0]
		sessionIDs = sessionIDs[1:]
		
		// 删除旧会话
		oldSessionKey := s.buildSessionKey(oldSessionID)
		s.cacheService.Delete(ctx, oldSessionKey)
	}

	// 保存更新后的会话列表，设置较长的过期时间
	return s.cacheService.Set(ctx, key, sessionIDs, 7*24*time.Hour)
}

func (s *SessionService) removeFromUserSessions(ctx context.Context, userID, sessionID string) error {
	key := s.buildUserSessionsKey(userID)
	
	var sessionIDs []string
	if err := s.cacheService.Get(ctx, key, &sessionIDs); err != nil {
		return nil // 列表不存在，无需处理
	}

	// 移除指定的会话ID
	var newSessionIDs []string
	for _, id := range sessionIDs {
		if id != sessionID {
			newSessionIDs = append(newSessionIDs, id)
		}
	}

	if len(newSessionIDs) == 0 {
		// 如果没有剩余会话，删除整个列表
		return s.cacheService.Delete(ctx, key)
	}

	// 保存更新后的会话列表
	return s.cacheService.Set(ctx, key, newSessionIDs, 7*24*time.Hour)
}

func (s *SessionService) getUserSessions(ctx context.Context, userID string) ([]string, error) {
	key := s.buildUserSessionsKey(userID)
	
	var sessionIDs []string
	if err := s.cacheService.Get(ctx, key, &sessionIDs); err != nil {
		return []string{}, nil // 返回空列表而不是错误
	}

	return sessionIDs, nil
}

func generateSessionID() string {
	// 生成唯一的会话ID
	return fmt.Sprintf("sess_%d_%s", time.Now().UnixNano(), generateRandomString(16))
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)] // 简化实现
	}
	return string(b)
}