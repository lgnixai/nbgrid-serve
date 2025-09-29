package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"github.com/gin-gonic/gin"

	"teable-go-backend/internal/application"
	"teable-go-backend/internal/domain/permission"
	"teable-go-backend/pkg/errors"
	"teable-go-backend/pkg/logger"
	"teable-go-backend/pkg/response"
)

// PermissionMiddleware 权限检查中间件
type PermissionMiddleware struct {
	permissionService *application.PermissionService
}

// NewPermissionMiddleware 创建权限中间件
func NewPermissionMiddleware(permissionService *application.PermissionService) *PermissionMiddleware {
	return &PermissionMiddleware{
		permissionService: permissionService,
	}
}

// RequirePermission 要求特定权限
func (m *PermissionMiddleware) RequirePermission(resourceType string, action permission.Action) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := GetCurrentUserID(c)
		if err != nil {
			response.Error(c, errors.ErrUnauthorized.WithDetails("Authentication required"))
			c.Abort()
			return
		}

		// 从路径参数获取资源ID
		resourceID := m.extractResourceID(c, resourceType)
		if resourceID == "" {
			response.Error(c, errors.ErrBadRequest.WithDetails("Resource ID required"))
			c.Abort()
			return
		}

		// 检查权限
		req := &application.PermissionCheckRequest{
			UserID:       userID,
			ResourceType: resourceType,
			ResourceID:   resourceID,
			Action:       action,
		}

		result, err := m.permissionService.CheckPermission(c.Request.Context(), req)
		if err != nil {
			logger.Error("Permission check failed",
				logger.String("user_id", userID),
				logger.String("resource_type", resourceType),
				logger.String("resource_id", resourceID),
				logger.String("action", string(action)),
				logger.ErrorField(err),
			)
			response.Error(c, errors.ErrInternalServer.WithDetails("Permission check failed"))
			c.Abort()
			return
		}

		if !result.Allowed {
			response.Error(c, errors.ErrForbidden.WithDetails(result.Reason))
			c.Abort()
			return
		}

		// 将权限信息存储到上下文中
		c.Set("user_role", result.Role)
		c.Set("user_permissions", result.Permissions)

		c.Next()
	}
}

// RequireAnyPermission 要求任意一个权限
func (m *PermissionMiddleware) RequireAnyPermission(resourceType string, actions ...permission.Action) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := GetCurrentUserID(c)
		if err != nil {
			response.Error(c, errors.ErrUnauthorized.WithDetails("Authentication required"))
			c.Abort()
			return
		}

		resourceID := m.extractResourceID(c, resourceType)
		if resourceID == "" {
			response.Error(c, errors.ErrBadRequest.WithDetails("Resource ID required"))
			c.Abort()
			return
		}

		// 检查是否有任意一个权限
		var allowed bool
		var lastResult *application.PermissionCheckResponse

		for _, action := range actions {
			req := &application.PermissionCheckRequest{
				UserID:       userID,
				ResourceType: resourceType,
				ResourceID:   resourceID,
				Action:       action,
			}

			result, err := m.permissionService.CheckPermission(c.Request.Context(), req)
			if err != nil {
				logger.Error("Permission check failed",
					logger.String("user_id", userID),
					logger.String("resource_type", resourceType),
					logger.String("resource_id", resourceID),
					logger.String("action", string(action)),
					logger.ErrorField(err),
				)
				continue
			}

			lastResult = result
			if result.Allowed {
				allowed = true
				break
			}
		}

		if !allowed {
			_reason := "Permission denied for all requested actions"
			if lastResult != nil {
				_reason = lastResult.Reason
			}
			response.Error(c, errors.ErrForbidden.WithDetails(_reason))
			c.Abort()
			return
		}

		// 将权限信息存储到上下文中
		if lastResult != nil {
			c.Set("user_role", lastResult.Role)
			c.Set("user_permissions", lastResult.Permissions)
		}

		c.Next()
	}
}

// RequireAllPermissions 要求所有权限
func (m *PermissionMiddleware) RequireAllPermissions(resourceType string, actions ...permission.Action) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := GetCurrentUserID(c)
		if err != nil {
			response.Error(c, errors.ErrUnauthorized.WithDetails("Authentication required"))
			c.Abort()
			return
		}

		resourceID := m.extractResourceID(c, resourceType)
		if resourceID == "" {
			response.Error(c, errors.ErrBadRequest.WithDetails("Resource ID required"))
			c.Abort()
			return
		}

		// 检查所有权限
		req := &application.PermissionCheckRequest{
			UserID:       userID,
			ResourceType: resourceType,
			ResourceID:   resourceID,
			Actions:      actions,
		}

		result, err := m.permissionService.CheckMultiplePermissions(c.Request.Context(), req)
		if err != nil {
			logger.Error("Multiple permissions check failed",
				logger.String("user_id", userID),
				logger.String("resource_type", resourceType),
				logger.String("resource_id", resourceID),
				logger.ErrorField(err),
			)
			response.Error(c, errors.ErrInternalServer.WithDetails("Permission check failed"))
			c.Abort()
			return
		}

		if !result.Allowed {
			response.Error(c, errors.ErrForbidden.WithDetails(result.Reason))
			c.Abort()
			return
		}

		// 将权限信息存储到上下文中
		c.Set("user_role", result.Role)
		c.Set("user_permissions", result.Permissions)

		c.Next()
	}
}

// RequireRole 要求特定角色
func (m *PermissionMiddleware) RequireRole(resourceType string, roles ...permission.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := GetCurrentUserID(c)
		if err != nil {
			response.Error(c, errors.ErrUnauthorized.WithDetails("Authentication required"))
			c.Abort()
			return
		}

		resourceID := m.extractResourceID(c, resourceType)
		if resourceID == "" {
			response.Error(c, errors.ErrBadRequest.WithDetails("Resource ID required"))
			c.Abort()
			return
		}

		// 获取用户权限信息
		result, err := m.permissionService.GetUserResourcePermissions(c.Request.Context(), userID, resourceType, resourceID)
		if err != nil {
			logger.Error("Failed to get user resource permissions",
				logger.String("user_id", userID),
				logger.String("resource_type", resourceType),
				logger.String("resource_id", resourceID),
				logger.ErrorField(err),
			)
			response.Error(c, errors.ErrInternalServer.WithDetails("Permission check failed"))
			c.Abort()
			return
		}

		// 检查角色
		hasRole := false
		for _, role := range roles {
			if result.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			response.Error(c, errors.ErrForbidden.WithDetails("Insufficient role"))
			c.Abort()
			return
		}

		// 将权限信息存储到上下文中
		c.Set("user_role", result.Role)
		c.Set("user_permissions", result.Permissions)

		c.Next()
	}
}

// RequireResourceOwnership 要求资源所有权
func (m *PermissionMiddleware) RequireResourceOwnership(resourceType string) gin.HandlerFunc {
	var ownerRole permission.Role
	switch resourceType {
	case "space":
		ownerRole = permission.RoleOwner
	case "base":
		ownerRole = permission.RoleBaseOwner
	default:
		// 对于其他资源类型，使用通用的所有者角色
		ownerRole = permission.RoleOwner
	}

	return m.RequireRole(resourceType, ownerRole)
}

// OptionalPermission 可选权限检查（不阻止请求）
func (m *PermissionMiddleware) OptionalPermission(resourceType string, action permission.Action) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := GetCurrentUserID(c)
		if err != nil {
			c.Next()
			return
		}

		resourceID := m.extractResourceID(c, resourceType)
		if resourceID == "" {
			c.Next()
			return
		}

		// 检查权限
		req := &application.PermissionCheckRequest{
			UserID:       userID,
			ResourceType: resourceType,
			ResourceID:   resourceID,
			Action:       action,
		}

		result, err := m.permissionService.CheckPermission(c.Request.Context(), req)
		if err != nil {
			logger.Warn("Optional permission check failed",
				logger.String("user_id", userID),
				logger.String("resource_type", resourceType),
				logger.String("resource_id", resourceID),
				logger.String("action", string(action)),
				logger.ErrorField(err),
			)
			c.Next()
			return
		}

		// 将权限信息存储到上下文中（无论是否有权限）
		c.Set("user_role", result.Role)
		c.Set("user_permissions", result.Permissions)
		c.Set("permission_allowed", result.Allowed)

		c.Next()
	}
}

// 私有方法

func (m *PermissionMiddleware) extractResourceID(c *gin.Context, resourceType string) string {
	// 1) 优先从资源类型特定的路径参数中提取（仅在非空时返回）
	var id string
	switch resourceType {
	case "space":
		if v := c.Param("spaceId"); v != "" {
			return v
		}
		if v := c.Param("space_id"); v != "" {
			return v
		}
	case "base":
		if v := c.Param("baseId"); v != "" {
			return v
		}
		if v := c.Param("base_id"); v != "" {
			return v
		}
	case "table":
		if v := c.Param("tableId"); v != "" {
			return v
		}
		if v := c.Param("table_id"); v != "" {
			return v
		}
	case "view":
		if v := c.Param("viewId"); v != "" {
			return v
		}
		if v := c.Param("view_id"); v != "" {
			return v
		}
	case "record":
		if v := c.Param("recordId"); v != "" {
			return v
		}
		if v := c.Param("record_id"); v != "" {
			return v
		}
	case "field":
		if v := c.Param("fieldId"); v != "" {
			return v
		}
		if v := c.Param("field_id"); v != "" {
			return v
		}
	}

	// 2) 通用路径参数 id
	if id = c.Param("id"); id != "" {
		return id
	}

	// 3) 查询参数 <resourceType>_id
	if id = c.Query(resourceType + "_id"); id != "" {
		return id
	}

	// 4) 尝试从 JSON 请求体读取（保持 body 可再次读取）
	if c.Request != nil && c.Request.Body != nil {
		ct := strings.ToLower(c.GetHeader("Content-Type"))
		if strings.Contains(ct, "application/json") {
			data, _ := io.ReadAll(io.LimitReader(c.Request.Body, 64*1024))
			c.Request.Body = io.NopCloser(bytes.NewReader(data))

			var body map[string]interface{}
			if len(data) > 0 && json.Unmarshal(data, &body) == nil {
				// 4.1) 优先 <resourceType>_id
				if v, ok := body[resourceType+"_id"]; ok {
					if s, ok2 := v.(string); ok2 && s != "" {
						return s
					}
				}
				// 4.2) 常见别名
				for _, k := range []string{"table_id", "base_id", "space_id", "view_id", "field_id", "record_id", "id"} {
					if v, ok := body[k]; ok {
						if s, ok2 := v.(string); ok2 && s != "" {
							return s
						}
					}
				}
			}
		}
	}

	return ""
}

// GetCurrentUserRole 获取当前用户角色
func GetCurrentUserRole(c *gin.Context) (permission.Role, error) {
	role, exists := c.Get("user_role")
	if !exists {
		return "", errors.ErrUnauthorized
	}

	r, ok := role.(permission.Role)
	if !ok {
		return "", errors.ErrInternalServer
	}

	return r, nil
}

// GetCurrentUserPermissions 获取当前用户权限
func GetCurrentUserPermissions(c *gin.Context) ([]permission.Action, error) {
	permissions, exists := c.Get("user_permissions")
	if !exists {
		return nil, errors.ErrUnauthorized
	}

	p, ok := permissions.([]permission.Action)
	if !ok {
		return nil, errors.ErrInternalServer
	}

	return p, nil
}

// HasPermission 检查当前用户是否有指定权限
func HasPermission(c *gin.Context, action permission.Action) bool {
	permissions, err := GetCurrentUserPermissions(c)
	if err != nil {
		return false
	}

	for _, p := range permissions {
		if p == action {
			return true
		}
	}

	return false
}

// IsPermissionAllowed 检查权限是否被允许（用于可选权限检查）
func IsPermissionAllowed(c *gin.Context) bool {
	allowed, exists := c.Get("permission_allowed")
	if !exists {
		return false
	}

	a, ok := allowed.(bool)
	if !ok {
		return false
	}

	return a
}
