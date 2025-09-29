package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"teable-go-backend/internal/application"
	"teable-go-backend/internal/domain/permission"
	"teable-go-backend/pkg/errors"
	"teable-go-backend/pkg/logger"
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
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
				"code":  "AUTH_REQUIRED",
			})
			c.Abort()
			return
		}

		// 从路径参数获取资源ID
		resourceID := m.extractResourceID(c, resourceType)
		if resourceID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Resource ID required",
				"code":  "RESOURCE_ID_REQUIRED",
			})
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
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Permission check failed",
				"code":  "PERMISSION_CHECK_FAILED",
			})
			c.Abort()
			return
		}

		if !result.Allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error":  "Permission denied",
				"code":   "PERMISSION_DENIED",
				"reason": result.Reason,
			})
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
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
				"code":  "AUTH_REQUIRED",
			})
			c.Abort()
			return
		}

		resourceID := m.extractResourceID(c, resourceType)
		if resourceID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Resource ID required",
				"code":  "RESOURCE_ID_REQUIRED",
			})
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
			reason := "Permission denied for all requested actions"
			if lastResult != nil {
				reason = lastResult.Reason
			}

			c.JSON(http.StatusForbidden, gin.H{
				"error":  "Permission denied",
				"code":   "PERMISSION_DENIED",
				"reason": reason,
			})
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
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
				"code":  "AUTH_REQUIRED",
			})
			c.Abort()
			return
		}

		resourceID := m.extractResourceID(c, resourceType)
		if resourceID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Resource ID required",
				"code":  "RESOURCE_ID_REQUIRED",
			})
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
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Permission check failed",
				"code":  "PERMISSION_CHECK_FAILED",
			})
			c.Abort()
			return
		}

		if !result.Allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error":  "Permission denied",
				"code":   "PERMISSION_DENIED",
				"reason": result.Reason,
			})
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
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
				"code":  "AUTH_REQUIRED",
			})
			c.Abort()
			return
		}

		resourceID := m.extractResourceID(c, resourceType)
		if resourceID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Resource ID required",
				"code":  "RESOURCE_ID_REQUIRED",
			})
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
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Permission check failed",
				"code":  "PERMISSION_CHECK_FAILED",
			})
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
			c.JSON(http.StatusForbidden, gin.H{
				"error":      "Insufficient role",
				"code":       "INSUFFICIENT_ROLE",
				"user_role":  result.Role,
				"required_roles": roles,
			})
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
	// 根据资源类型从不同的路径参数中提取资源ID
	switch resourceType {
	case "space":
		if spaceID := c.Param("spaceId"); spaceID != "" {
			return spaceID
		}
		return c.Param("space_id")
	case "base":
		if baseID := c.Param("baseId"); baseID != "" {
			return baseID
		}
		return c.Param("base_id")
	case "table":
		if tableID := c.Param("tableId"); tableID != "" {
			return tableID
		}
		return c.Param("table_id")
	case "view":
		if viewID := c.Param("viewId"); viewID != "" {
			return viewID
		}
		return c.Param("view_id")
	case "record":
		if recordID := c.Param("recordId"); recordID != "" {
			return recordID
		}
		return c.Param("record_id")
	case "field":
		if fieldID := c.Param("fieldId"); fieldID != "" {
			return fieldID
		}
		return c.Param("field_id")
	default:
		// 尝试通用的ID参数
		if id := c.Param("id"); id != "" {
			return id
		}
		// 尝试从查询参数获取
		return c.Query(resourceType + "_id")
	}
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