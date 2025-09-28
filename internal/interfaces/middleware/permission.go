package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"teable-go-backend/internal/domain/permission"
	appErrors "teable-go-backend/pkg/errors"
	"teable-go-backend/pkg/logger"
)

// PermissionMiddleware 权限中间件
func PermissionMiddleware(permissionService permission.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户信息
		_, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not found in context",
				"code":  "USER_NOT_FOUND",
			})
			c.Abort()
			return
		}

		// 获取权限要求
		requiredPermissions, exists := c.Get("required_permissions")
		if !exists {
			// 没有权限要求，直接通过
			c.Next()
			return
		}

		permissions, ok := requiredPermissions.([]permission.Action)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid permission format",
				"code":  "INVALID_PERMISSION_FORMAT",
			})
			c.Abort()
			return
		}

		// 获取资源信息
		resourceType, resourceID := extractResourceInfo(c)
		if resourceType == "" || resourceID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Resource information not found",
				"code":  "RESOURCE_NOT_FOUND",
			})
			c.Abort()
			return
		}

		// 获取用户ID
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User ID not found",
				"code":  "USER_ID_NOT_FOUND",
			})
			c.Abort()
			return
		}

		// 检查权限
		hasPermission, err := permissionService.CheckMultiplePermissions(
			c.Request.Context(),
			userID.(string),
			resourceType,
			resourceID,
			permissions,
		)
		if err != nil {
			logger.Error("Failed to check permissions",
				logger.String("user_id", userID.(string)),
				logger.String("resource_type", resourceType),
				logger.String("resource_id", resourceID),
				logger.ErrorField(err),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to check permissions",
				"code":  "PERMISSION_CHECK_FAILED",
			})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
				"code":  "INSUFFICIENT_PERMISSIONS",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermissions 设置权限要求
func RequirePermissions(actions ...permission.Action) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("required_permissions", actions)
		c.Next()
	}
}

// RequireSpacePermission 要求空间权限
func RequireSpacePermission(action permission.Action) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("required_permissions", []permission.Action{action})
		c.Set("resource_type", "space")
		c.Next()
	}
}

// RequireBasePermission 要求基础表权限
func RequireBasePermission(action permission.Action) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("required_permissions", []permission.Action{action})
		c.Set("resource_type", "base")
		c.Next()
	}
}

// RequireTablePermission 要求表格权限
func RequireTablePermission(action permission.Action) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("required_permissions", []permission.Action{action})
		c.Set("resource_type", "table")
		c.Next()
	}
}

// RequireViewPermission 要求视图权限
func RequireViewPermission(action permission.Action) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("required_permissions", []permission.Action{action})
		c.Set("resource_type", "view")
		c.Next()
	}
}

// RequireFieldPermission 要求字段权限
func RequireFieldPermission(action permission.Action) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("required_permissions", []permission.Action{action})
		c.Set("resource_type", "field")
		c.Next()
	}
}

// RequireRecordPermission 要求记录权限
func RequireRecordPermission(action permission.Action) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("required_permissions", []permission.Action{action})
		c.Set("resource_type", "record")
		c.Next()
	}
}

// extractResourceInfo 提取资源信息
func extractResourceInfo(c *gin.Context) (resourceType, resourceID string) {
	// 从路径参数中提取
	if spaceID := c.Param("space_id"); spaceID != "" {
		return "space", spaceID
	}
	if baseID := c.Param("base_id"); baseID != "" {
		return "base", baseID
	}
	if tableID := c.Param("table_id"); tableID != "" {
		return "table", tableID
	}
	if viewID := c.Param("view_id"); viewID != "" {
		return "view", viewID
	}
	if fieldID := c.Param("field_id"); fieldID != "" {
		return "field", fieldID
	}
	if recordID := c.Param("record_id"); recordID != "" {
		return "record", recordID
	}

	// 从查询参数中提取
	if spaceID := c.Query("space_id"); spaceID != "" {
		return "space", spaceID
	}
	if baseID := c.Query("base_id"); baseID != "" {
		return "base", baseID
	}
	if tableID := c.Query("table_id"); tableID != "" {
		return "table", tableID
	}
	if viewID := c.Query("view_id"); viewID != "" {
		return "view", viewID
	}
	if fieldID := c.Query("field_id"); fieldID != "" {
		return "field", fieldID
	}
	if recordID := c.Query("record_id"); recordID != "" {
		return "record", recordID
	}

	// 从请求体中提取（需要解析JSON）
	// 这里可以根据需要添加JSON解析逻辑

	// 从上下文中获取
	if resourceType, exists := c.Get("resource_type"); exists {
		if resourceID, exists := c.Get("resource_id"); exists {
			return resourceType.(string), resourceID.(string)
		}
	}

	return "", ""
}

// CheckPermission 检查权限的辅助函数
func CheckPermission(ctx context.Context, permissionService permission.Service, userID, resourceType, resourceID string, action permission.Action) error {
	hasPermission, err := permissionService.CheckPermission(ctx, userID, resourceType, resourceID, action)
	if err != nil {
		return appErrors.ErrInternalServer.WithDetails(err.Error())
	}

	if !hasPermission {
		return appErrors.ErrForbidden.WithDetails("Insufficient permissions")
	}

	return nil
}

// CheckMultiplePermissions 检查多个权限的辅助函数
func CheckMultiplePermissions(ctx context.Context, permissionService permission.Service, userID, resourceType, resourceID string, actions []permission.Action) error {
	hasPermission, err := permissionService.CheckMultiplePermissions(ctx, userID, resourceType, resourceID, actions)
	if err != nil {
		return appErrors.ErrInternalServer.WithDetails(err.Error())
	}

	if !hasPermission {
		return appErrors.ErrForbidden.WithDetails("Insufficient permissions")
	}

	return nil
}

// GetUserEffectiveRole 获取用户有效角色的辅助函数
func GetUserEffectiveRole(ctx context.Context, permissionService permission.Service, userID, resourceType, resourceID string) (permission.Role, error) {
	role, err := permissionService.GetUserEffectiveRole(ctx, userID, resourceType, resourceID)
	if err != nil {
		return "", appErrors.ErrInternalServer.WithDetails(err.Error())
	}

	return role, nil
}

// IsAdminUser 检查是否为管理员用户
func IsAdminUser(c *gin.Context) bool {
	user, exists := c.Get("user")
	if !exists {
		return false
	}

	// 这里需要根据实际的用户模型来检查
	// 暂时返回false，实际实现时需要检查用户的IsAdmin字段
	_ = user
	return false
}

// IsSystemUser 检查是否为系统用户
func IsSystemUser(c *gin.Context) bool {
	user, exists := c.Get("user")
	if !exists {
		return false
	}

	// 这里需要根据实际的用户模型来检查
	// 暂时返回false，实际实现时需要检查用户的IsSystem字段
	_ = user
	return false
}

// AdminOnlyMiddleware 仅管理员可访问的中间件
func AdminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !IsAdminUser(c) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Admin access required",
				"code":  "ADMIN_ACCESS_REQUIRED",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// SystemOnlyMiddleware 仅系统用户可访问的中间件
func SystemOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !IsSystemUser(c) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "System access required",
				"code":  "SYSTEM_ACCESS_REQUIRED",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// ResourceOwnerMiddleware 资源所有者中间件
func ResourceOwnerMiddleware(permissionService permission.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User ID not found",
				"code":  "USER_ID_NOT_FOUND",
			})
			c.Abort()
			return
		}

		resourceType, resourceID := extractResourceInfo(c)
		if resourceType == "" || resourceID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Resource information not found",
				"code":  "RESOURCE_NOT_FOUND",
			})
			c.Abort()
			return
		}

		// 检查是否为资源所有者
		role, err := permissionService.GetUserEffectiveRole(c.Request.Context(), userID.(string), resourceType, resourceID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to check user role",
				"code":  "ROLE_CHECK_FAILED",
			})
			c.Abort()
			return
		}

		// 检查是否为所有者角色
		isOwner := false
		switch resourceType {
		case "space":
			isOwner = role == permission.RoleOwner
		case "base":
			isOwner = role == permission.RoleBaseOwner
		}

		if !isOwner {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Resource owner access required",
				"code":  "OWNER_ACCESS_REQUIRED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// parseActionsFromString 从字符串解析权限动作
func parseActionsFromString(actionsStr string) []permission.Action {
	if actionsStr == "" {
		return nil
	}

	actionStrs := strings.Split(actionsStr, ",")
	var actions []permission.Action

	for _, actionStr := range actionStrs {
		actionStr = strings.TrimSpace(actionStr)
		if actionStr != "" {
			actions = append(actions, permission.Action(actionStr))
		}
	}

	return actions
}
