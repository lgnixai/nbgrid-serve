package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"teable-go-backend/internal/config"
	userDomain "teable-go-backend/internal/domain/user"
	"teable-go-backend/internal/infrastructure/cache"
	"teable-go-backend/pkg/errors"
	"teable-go-backend/pkg/logger"
)

// TokenService JWT令牌服务
type TokenService struct {
	config       config.JWTConfig
	cacheService cache.CacheService
}

// NewTokenService 创建令牌服务
func NewTokenService(jwtConfig config.JWTConfig, cacheService cache.CacheService) *TokenService {
	return &TokenService{
		config:       jwtConfig,
		cacheService: cacheService,
	}
}

// TokenClaims JWT声明
type TokenClaims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	IsAdmin   bool   `json:"is_admin"`
	IsSystem  bool   `json:"is_system"`
	TokenType string `json:"token_type"` // access, refresh
	SessionID string `json:"session_id,omitempty"`
	jwt.RegisteredClaims
}

// TokenPair 令牌对
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// GenerateTokenPair 生成访问令牌和刷新令牌对
func (s *TokenService) GenerateTokenPair(ctx context.Context, user *userDomain.User, sessionID string) (*TokenPair, error) {
	now := time.Now()

	// 生成访问令牌
	accessToken, err := s.generateToken(user, "access", sessionID, now, s.config.AccessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	var refreshToken string
	if s.config.EnableRefresh {
		refreshToken, err = s.generateToken(user, "refresh", sessionID, now, s.config.RefreshTokenTTL)
		if err != nil {
			return nil, fmt.Errorf("failed to generate refresh token: %w", err)
		}
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.config.AccessTokenTTL.Seconds()),
	}, nil
}

// ValidateToken 验证令牌
func (s *TokenService) ValidateToken(ctx context.Context, tokenString string) (*TokenClaims, error) {
	// 检查令牌是否在黑名单中
	if blacklisted, err := s.isTokenBlacklisted(ctx, tokenString); err != nil {
		logger.Error("Failed to check token blacklist", logger.ErrorField(err))
		return nil, errors.ErrInternalServer
	} else if blacklisted {
		return nil, errors.ErrInvalidToken
	}

	// 解析和验证令牌
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 检查签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.ErrInvalidToken
		}
		return []byte(s.config.Secret), nil
	})

	if err != nil {
		if err == jwt.ErrTokenExpired {
			return nil, errors.ErrTokenExpired
		}
		return nil, errors.ErrInvalidToken
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, errors.ErrInvalidToken
	}

	return claims, nil
}

// RefreshToken 刷新访问令牌
func (s *TokenService) RefreshToken(ctx context.Context, refreshTokenString string, user *userDomain.User) (*TokenPair, error) {
	// 验证刷新令牌
	claims, err := s.ValidateToken(ctx, refreshTokenString)
	if err != nil {
		return nil, err
	}

	// 检查令牌类型
	if claims.TokenType != "refresh" {
		return nil, errors.ErrInvalidToken
	}

	// 检查用户ID是否匹配
	if claims.UserID != user.ID {
		return nil, errors.ErrInvalidToken
	}

	// 将旧的刷新令牌加入黑名单
	if err := s.blacklistToken(ctx, refreshTokenString, s.config.RefreshTokenTTL); err != nil {
		logger.Warn("Failed to blacklist old refresh token",
			logger.String("user_id", user.ID),
			logger.ErrorField(err),
		)
	}

	// 生成新的令牌对
	return s.GenerateTokenPair(ctx, user, claims.SessionID)
}

// BlacklistToken 将令牌加入黑名单
func (s *TokenService) BlacklistToken(ctx context.Context, tokenString string) error {
	// 解析令牌以获取过期时间
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.Secret), nil
	})

	var ttl time.Duration = s.config.AccessTokenTTL // 默认TTL
	if err == nil {
		if claims, ok := token.Claims.(*TokenClaims); ok {
			// 计算令牌剩余有效时间
			if claims.ExpiresAt != nil {
				remaining := time.Until(claims.ExpiresAt.Time)
				if remaining > 0 {
					ttl = remaining
				}
			}
		}
	}

	return s.blacklistToken(ctx, tokenString, ttl)
}

// InvalidateUserTokens 使用户的所有令牌失效
func (s *TokenService) InvalidateUserTokens(ctx context.Context, userID string) error {
	// 将用户ID加入令牌失效列表
	key := cache.BuildCacheKey("user_token_invalidated:", userID)
	
	// 设置较长的过期时间，确保覆盖所有可能的令牌
	ttl := s.config.RefreshTokenTTL
	if s.config.AccessTokenTTL > ttl {
		ttl = s.config.AccessTokenTTL
	}
	
	return s.cacheService.Set(ctx, key, time.Now().Unix(), ttl)
}

// IsUserTokensInvalidated 检查用户令牌是否已失效
func (s *TokenService) IsUserTokensInvalidated(ctx context.Context, userID string, tokenIssuedAt time.Time) (bool, error) {
	key := cache.BuildCacheKey("user_token_invalidated:", userID)
	
	var invalidatedAt int64
	if err := s.cacheService.Get(ctx, key, &invalidatedAt); err != nil {
		// 如果键不存在，说明没有失效
		return false, nil
	}

	// 如果令牌签发时间早于失效时间，则令牌无效
	return tokenIssuedAt.Unix() < invalidatedAt, nil
}

// GenerateAPIKey 生成API密钥
func (s *TokenService) GenerateAPIKey(ctx context.Context, userID string, name string, expiresAt *time.Time) (string, error) {
	// 生成随机API密钥
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", fmt.Errorf("failed to generate random key: %w", err)
	}
	
	apiKey := "tk_" + hex.EncodeToString(keyBytes)
	
	// 存储API密钥信息
	keyInfo := map[string]interface{}{
		"user_id":    userID,
		"name":       name,
		"created_at": time.Now().Unix(),
	}
	
	if expiresAt != nil {
		keyInfo["expires_at"] = expiresAt.Unix()
	}
	
	key := cache.BuildCacheKey("api_key:", apiKey)
	var ttl time.Duration = 365 * 24 * time.Hour // 默认1年
	if expiresAt != nil {
		ttl = time.Until(*expiresAt)
	}
	
	if err := s.cacheService.Set(ctx, key, keyInfo, ttl); err != nil {
		return "", fmt.Errorf("failed to store API key: %w", err)
	}
	
	logger.Info("API key generated",
		logger.String("user_id", userID),
		logger.String("key_name", name),
	)
	
	return apiKey, nil
}

// ValidateAPIKey 验证API密钥
func (s *TokenService) ValidateAPIKey(ctx context.Context, apiKey string) (string, error) {
	key := cache.BuildCacheKey("api_key:", apiKey)
	
	var keyInfo map[string]interface{}
	if err := s.cacheService.Get(ctx, key, &keyInfo); err != nil {
		return "", errors.ErrInvalidToken
	}
	
	userID, ok := keyInfo["user_id"].(string)
	if !ok {
		return "", errors.ErrInvalidToken
	}
	
	// 检查是否过期
	if expiresAt, exists := keyInfo["expires_at"]; exists {
		if expTime, ok := expiresAt.(float64); ok {
			if time.Now().Unix() > int64(expTime) {
				return "", errors.ErrTokenExpired
			}
		}
	}
	
	return userID, nil
}

// RevokeAPIKey 撤销API密钥
func (s *TokenService) RevokeAPIKey(ctx context.Context, apiKey string) error {
	key := cache.BuildCacheKey("api_key:", apiKey)
	return s.cacheService.Delete(ctx, key)
}

// 私有方法

func (s *TokenService) generateToken(user *userDomain.User, tokenType, sessionID string, issuedAt time.Time, ttl time.Duration) (string, error) {
	expiresAt := issuedAt.Add(ttl)

	claims := TokenClaims{
		UserID:    user.ID,
		Email:     user.Email,
		Name:      user.Name,
		IsAdmin:   user.IsAdmin,
		IsSystem:  user.IsSystem,
		TokenType: tokenType,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.config.Issuer,
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(issuedAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.Secret))
}

func (s *TokenService) blacklistToken(ctx context.Context, tokenString string, ttl time.Duration) error {
	key := cache.BuildCacheKey("blacklist:", tokenString)
	return s.cacheService.Set(ctx, key, true, ttl)
}

func (s *TokenService) isTokenBlacklisted(ctx context.Context, tokenString string) (bool, error) {
	key := cache.BuildCacheKey("blacklist:", tokenString)
	return s.cacheService.Exists(ctx, key)
}