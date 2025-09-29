package http

import (
	"github.com/gin-gonic/gin"

	"teable-go-backend/internal/application"
	"teable-go-backend/internal/domain/attachment"
	"teable-go-backend/internal/domain/base"
	"teable-go-backend/internal/domain/notification"
	"teable-go-backend/internal/domain/permission"
	recdomain "teable-go-backend/internal/domain/record"
	"teable-go-backend/internal/domain/search"
	"teable-go-backend/internal/domain/share"
	"teable-go-backend/internal/domain/sharedb"
	"teable-go-backend/internal/domain/space"
	"teable-go-backend/internal/domain/table"
	"teable-go-backend/internal/domain/view"
	"teable-go-backend/internal/domain/websocket"
	"teable-go-backend/internal/infrastructure/database"
	"teable-go-backend/internal/infrastructure/monitoring"
	"teable-go-backend/internal/interfaces/middleware"
)

// RouterConfig 路由配置
type RouterConfig struct {
	UserService          *application.UserService
	AuthService          middleware.AuthService
	SpaceService         space.Service
	BaseService          base.Service
	TableService         table.Service
	RecordAppService     *application.RecordService
	RecordService        recdomain.Service
	ViewService          view.Service
	PermissionService    permission.Service
	ShareService         share.Service
	AttachmentService    attachment.Service
	NotificationService  notification.Service
	SearchService        search.Service
	WebSocketService     websocket.Service
	WebSocketHandler     *websocket.Handler
	ShareDBService       sharedb.ShareDB
	ShareDBWSIntegration *sharedb.WebSocketIntegration
	CollaborationService *websocket.CollaborationService
	DB                   *database.Connection
	ErrorMonitor         *monitoring.ErrorMonitor
}

// SetupRoutes 设置路由
func SetupRoutes(router *gin.Engine, config RouterConfig) {
	// 创建处理器
	userHandler := NewUserHandler(config.UserService)
	spaceHandler := NewSpaceHandler(config.SpaceService)
	baseHandler := NewBaseHandler(config.BaseService)
	tableHandler := NewTableHandler(config.TableService)
	var recordHandler *RecordHandler
	if config.RecordAppService != nil {
		recordHandler = NewRecordHandler(config.RecordAppService)
	} else {
		// 回退到领域服务（不推荐），但避免空指针
		// 注意：领域服务缺少真实 schema 校验，可能导致 500
		recordHandler = NewRecordHandler(nil)
	}
	viewHandler := NewViewHandler(config.ViewService)
	permissionHandler := NewPermissionHandler(config.PermissionService, nil)       // 暂时传nil logger，稍后修复
	shareHandler := NewShareHandler(config.ShareService, nil)                      // 暂时传nil logger，稍后修复
	attachmentHandler := NewAttachmentHandler(config.AttachmentService, nil)       // 暂时传nil logger，稍后修复
	notificationHandler := NewNotificationHandler(config.NotificationService, nil) // 暂时传nil logger，稍后修复
	searchHandler := NewSearchHandler(config.SearchService, nil)                   // 暂时传nil logger，稍后修复
	pinHandler := NewPinHandler()
	healthHandler := NewHealthHandler(config.DB, config.ErrorMonitor)

	// WebSocket处理器
	wsHandler := NewWebSocketHandler(config.WebSocketHandler, nil) // 暂时传nil logger，稍后修复

	// ShareDB处理器
	sharedbHandler := NewShareDBHandler(config.ShareDBService, config.ShareDBWSIntegration, nil) // 暂时传nil logger，稍后修复

	// 协作处理器
	collaborationHandler := NewCollaborationHandler(config.CollaborationService, nil) // 暂时传nil logger，稍后修复

	// 添加全局中间件
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.PanicRecovery())
	router.Use(middleware.LoggingMiddleware())

	// 系统健康检查路由 (无需认证)
	router.GET("/health", healthHandler.HealthCheck)
	router.GET("/ready", healthHandler.ReadinessCheck)
	router.GET("/alive", healthHandler.LivenessCheck)

	// API v1 路由组
	v1 := router.Group("/api")
	{
		// 认证相关路由 (无需认证)
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", userHandler.Register)
			authGroup.POST("/login", userHandler.Login)
			authGroup.POST("/refresh", userHandler.RefreshToken)
			authGroup.POST("/logout", middleware.AuthMiddleware(config.AuthService), userHandler.Logout)
		}

		// 用户相关路由 (需要认证)
		userGroup := v1.Group("/users")
		userGroup.Use(middleware.AuthMiddleware(config.AuthService))
		{
			userGroup.GET("/profile", userHandler.GetProfile)
			userGroup.PUT("/profile", userHandler.UpdateProfile)
			userGroup.POST("/change-password", userHandler.ChangePassword)
			userGroup.GET("/:id/activity", userHandler.GetUserActivity)
			userGroup.GET("/preferences", userHandler.GetUserPreferences)
			userGroup.PUT("/preferences", userHandler.UpdateUserPreferences)
		}

		// 空间相关路由 (需要认证)
		spaceGroup := v1.Group("/spaces")
		spaceGroup.Use(middleware.AuthMiddleware(config.AuthService))
		{
			spaceGroup.POST("", spaceHandler.CreateSpace)
			spaceGroup.GET("", spaceHandler.ListSpaces)
			spaceGroup.GET(":id", spaceHandler.GetSpace)
			spaceGroup.PUT(":id", spaceHandler.UpdateSpace)
			spaceGroup.DELETE(":id", spaceHandler.DeleteSpace)

			// 协作者管理
			spaceGroup.POST(":id/collaborators", spaceHandler.AddCollaborator)
			spaceGroup.GET(":id/collaborators", spaceHandler.ListCollaborators)
			spaceGroup.DELETE(":id/collaborators/:collab_id", spaceHandler.RemoveCollaborator)
			spaceGroup.PUT(":id/collaborators/:collab_id/role", spaceHandler.UpdateCollaboratorRole)

			// 权限管理
			spaceGroup.GET(":id/permissions", spaceHandler.CheckUserPermission)

			// 统计信息
			spaceGroup.GET(":id/stats", spaceHandler.GetSpaceStats)

			// 用户空间
			spaceGroup.GET("user/:user_id", spaceHandler.GetUserSpaces)
			spaceGroup.GET("user/:user_id/stats", spaceHandler.GetUserSpaceStats)

			// 批量操作
			spaceGroup.POST("bulk-update", spaceHandler.BulkUpdateSpaces)
			spaceGroup.POST("bulk-delete", spaceHandler.BulkDeleteSpaces)
		}

		// 基础表相关路由 (需要认证)
		baseGroup := v1.Group("/bases")
		baseGroup.Use(middleware.AuthMiddleware(config.AuthService))
		{
			baseGroup.POST("", baseHandler.CreateBase)
			baseGroup.GET("", baseHandler.ListBases)
			baseGroup.GET(":id", baseHandler.GetBase)
			baseGroup.PUT(":id", baseHandler.UpdateBase)
			baseGroup.DELETE(":id", baseHandler.DeleteBase)

			// 权限管理
			baseGroup.GET(":id/permissions", baseHandler.CheckUserPermission)

			// 统计信息
			baseGroup.GET(":id/stats", baseHandler.GetBaseStats)
			baseGroup.GET("space/:space_id/stats", baseHandler.GetSpaceBaseStats)

			// 批量操作
			baseGroup.POST("bulk-update", baseHandler.BulkUpdateBases)
			baseGroup.POST("bulk-delete", baseHandler.BulkDeleteBases)

			// 导出/导入
			baseGroup.GET("export", baseHandler.ExportBases)
			baseGroup.POST("import", baseHandler.ImportBases)

			// 基础表下的数据表路由
			baseGroup.POST(":id/tables", tableHandler.CreateTable)
			baseGroup.GET(":id/tables", tableHandler.ListTables)
		}

		// 数据表相关路由 (需要认证)
		tableGroup := v1.Group("/tables")
		tableGroup.Use(middleware.AuthMiddleware(config.AuthService))
		{
			tableGroup.POST("", tableHandler.CreateTable)
			tableGroup.GET("", tableHandler.ListTables)
			tableGroup.GET(":id", tableHandler.GetTable)
			tableGroup.PUT(":id", tableHandler.UpdateTable)
			tableGroup.DELETE(":id", tableHandler.DeleteTable)
		}

		// 字段相关路由 (需要认证)
		fieldGroup := v1.Group("/fields")
		fieldGroup.Use(middleware.AuthMiddleware(config.AuthService))
		{
			fieldGroup.POST("", tableHandler.CreateField)
			fieldGroup.GET("", tableHandler.ListFields)
			fieldGroup.GET(":id", tableHandler.GetField)
			fieldGroup.PUT(":id", tableHandler.UpdateField)
			fieldGroup.DELETE(":id", tableHandler.DeleteField)

			// 字段类型和验证
			fieldGroup.GET("types", tableHandler.GetFieldTypes)
			fieldGroup.GET("types/:type", tableHandler.GetFieldTypeInfo)
			fieldGroup.POST(":field_id/validate", tableHandler.ValidateFieldValue)
		}

		// 记录相关路由 (需要认证)
		recordGroup := v1.Group("/records")
		recordGroup.Use(middleware.AuthMiddleware(config.AuthService))
		{
			recordGroup.POST("", recordHandler.CreateRecord)
			recordGroup.GET("", recordHandler.ListRecords)
			recordGroup.GET(":id", recordHandler.GetRecord)
			recordGroup.PUT(":id", recordHandler.UpdateRecord)
			recordGroup.DELETE(":id", recordHandler.DeleteRecord)

			// 批量操作
			recordGroup.POST("bulk", recordHandler.BulkCreateRecords)
			recordGroup.PUT("bulk", recordHandler.BulkUpdateRecords)
			recordGroup.DELETE("bulk", recordHandler.BulkDeleteRecords)

			// 复杂查询
			recordGroup.POST("query", recordHandler.ComplexQuery)

			// 统计信息
			recordGroup.GET("stats", recordHandler.GetRecordStats)

			// 导出导入
			recordGroup.POST("export", recordHandler.ExportRecords)
			recordGroup.POST("import", recordHandler.ImportRecords)
		}

		// 视图相关路由 (需要认证)
		viewGroup := v1.Group("/views")
		viewGroup.Use(middleware.AuthMiddleware(config.AuthService))
		{
			viewGroup.POST("", viewHandler.CreateView)
			viewGroup.GET("", viewHandler.ListViews)
			viewGroup.GET(":id", viewHandler.GetView)
			viewGroup.PUT(":id", viewHandler.UpdateView)
			viewGroup.DELETE(":id", viewHandler.DeleteView)

			// 视图配置
			viewGroup.GET(":id/config", viewHandler.GetViewConfig)
			viewGroup.PUT(":id/config", viewHandler.UpdateViewConfig)

			// 网格视图特定功能
			viewGroup.GET(":id/grid/data", viewHandler.GetGridViewData)
			viewGroup.PUT(":id/grid/config", viewHandler.UpdateGridViewConfig)
			viewGroup.POST(":id/grid/columns", viewHandler.AddGridViewColumn)
			viewGroup.PUT(":id/grid/columns/:field_id", viewHandler.UpdateGridViewColumn)
			viewGroup.DELETE(":id/grid/columns/:field_id", viewHandler.RemoveGridViewColumn)
			viewGroup.PUT(":id/grid/columns/reorder", viewHandler.ReorderGridViewColumns)

			// 表单视图特定功能
			viewGroup.GET(":id/form/data", viewHandler.GetFormViewData)
			viewGroup.PUT(":id/form/config", viewHandler.UpdateFormViewConfig)
			viewGroup.POST(":id/form/fields", viewHandler.AddFormViewField)
			viewGroup.PUT(":id/form/fields/:field_id", viewHandler.UpdateFormViewField)
			viewGroup.DELETE(":id/form/fields/:field_id", viewHandler.RemoveFormViewField)
			viewGroup.PUT(":id/form/fields/reorder", viewHandler.ReorderFormViewFields)

			// 看板视图特定功能
			viewGroup.GET(":id/kanban/data", viewHandler.GetKanbanViewData)
			viewGroup.PUT(":id/kanban/config", viewHandler.UpdateKanbanViewConfig)
			viewGroup.POST(":id/kanban/move", viewHandler.MoveKanbanCard)

			// 日历视图特定功能
			viewGroup.GET(":id/calendar/data", viewHandler.GetCalendarViewData)
			viewGroup.PUT(":id/calendar/config", viewHandler.UpdateCalendarViewConfig)

			// 画廊视图特定功能
			viewGroup.GET(":id/gallery/data", viewHandler.GetGalleryViewData)
			viewGroup.PUT(":id/gallery/config", viewHandler.UpdateGalleryViewConfig)
		}

		// Pin 相关路由 (需要认证)
		pinGroup := v1.Group("/pin")
		pinGroup.Use(middleware.AuthMiddleware(config.AuthService))
		{
			pinGroup.GET("/list", pinHandler.ListPins)
		}

		// 管理员相关路由 (需要管理员权限)
		adminGroup := v1.Group("/admin")
		adminGroup.Use(middleware.AuthMiddleware(config.AuthService))
		adminGroup.Use(middleware.AdminRequiredMiddleware())
		{
			// 用户管理
			adminGroup.GET("/users", userHandler.ListUsers)
			adminGroup.GET("/users/:id", userHandler.GetUser)
			adminGroup.PUT("/users/:id", userHandler.UpdateUser)
			adminGroup.DELETE("/users/:id", userHandler.DeleteUser)
			adminGroup.POST("/users/:id/promote", userHandler.PromoteToAdmin)
			adminGroup.POST("/users/:id/demote", userHandler.DemoteFromAdmin)
			adminGroup.POST("/users/:id/activate", userHandler.ActivateUser)
			adminGroup.POST("/users/:id/deactivate", userHandler.DeactivateUser)

			// 批量操作
			adminGroup.POST("/users/bulk-update", userHandler.BulkUpdateUsers)
			adminGroup.POST("/users/bulk-delete", userHandler.BulkDeleteUsers)
			adminGroup.GET("/users/export", userHandler.ExportUsers)
			adminGroup.POST("/users/import", userHandler.ImportUsers)
			adminGroup.GET("/users/stats", userHandler.GetUserStats)
		}

		// WebSocket相关路由
		wsGroup := v1.Group("/ws")
		{
			// WebSocket连接 (无需认证，在连接时验证)
			wsGroup.GET("/socket", wsHandler.HandleWebSocket)

			// WebSocket管理接口 (需要认证)
			wsGroup.Use(middleware.AuthMiddleware(config.AuthService))
			{
				wsGroup.GET("/stats", wsHandler.GetWebSocketStats)
				wsGroup.POST("/broadcast/channel", wsHandler.BroadcastToChannel)
				wsGroup.POST("/broadcast/user", wsHandler.BroadcastToUser)
				wsGroup.POST("/publish/document", wsHandler.PublishDocumentOp)
				wsGroup.POST("/publish/record", wsHandler.PublishRecordOp)
				wsGroup.POST("/publish/view", wsHandler.PublishViewOp)
				wsGroup.POST("/publish/field", wsHandler.PublishFieldOp)
			}
		}

		// ShareDB相关路由
		sharedbGroup := v1.Group("/sharedb")
		sharedbGroup.Use(middleware.AuthMiddleware(config.AuthService))
		{
			// ShareDB管理接口
			sharedbGroup.GET("/stats", sharedbHandler.GetShareDBStats)
			sharedbGroup.POST("/submit", sharedbHandler.HandleSubmit)
			sharedbGroup.GET("/snapshot/:collection/:id", sharedbHandler.GetSnapshot)
			sharedbGroup.POST("/query/:collection", sharedbHandler.Query)
			sharedbGroup.GET("/ops/:collection/:id", sharedbHandler.GetOps)
		}

		// 协作功能相关路由
		collaborationGroup := v1.Group("/collaboration")
		collaborationGroup.Use(middleware.AuthMiddleware(config.AuthService))
		{
			// 在线状态管理
			collaborationGroup.POST("/presence", collaborationHandler.UpdatePresence)
			collaborationGroup.DELETE("/presence", collaborationHandler.RemovePresence)
			collaborationGroup.GET("/presence", collaborationHandler.GetPresence)

			// 光标位置管理
			collaborationGroup.POST("/cursor", collaborationHandler.UpdateCursor)
			collaborationGroup.DELETE("/cursor", collaborationHandler.RemoveCursor)
			collaborationGroup.GET("/cursor", collaborationHandler.GetCursors)

			// 通知管理
			collaborationGroup.POST("/notification", collaborationHandler.SendNotification)

			// 统计信息
			collaborationGroup.GET("/stats", collaborationHandler.GetCollaborationStats)
		}

		// 权限管理相关路由
		permissionGroup := v1.Group("/permissions")
		permissionGroup.Use(middleware.AuthMiddleware(config.AuthService))
		{
			// 权限管理
			permissionGroup.POST("/grant", permissionHandler.GrantPermission)
			permissionGroup.POST("/revoke", permissionHandler.RevokePermission)
			permissionGroup.PUT("/:id", permissionHandler.UpdatePermission)
			permissionGroup.GET("/user/:user_id", permissionHandler.GetUserPermissions)
			permissionGroup.GET("/resource", permissionHandler.GetResourcePermissions)
			permissionGroup.POST("/check", permissionHandler.CheckPermission)

			// 角色管理
			permissionGroup.GET("/user/:user_id/role", permissionHandler.GetUserRole)
			permissionGroup.GET("/user/:user_id/resources", permissionHandler.GetUserResources)
			permissionGroup.GET("/resource/collaborators", permissionHandler.GetResourceCollaborators)
			permissionGroup.POST("/transfer-ownership", permissionHandler.TransferOwnership)

			// 统计和查询
			permissionGroup.GET("/stats", permissionHandler.GetPermissionStats)
			permissionGroup.GET("/role/:role/permissions", permissionHandler.GetRolePermissions)
			permissionGroup.POST("/role/check", permissionHandler.CheckRolePermission)
			permissionGroup.GET("/role/:role/level", permissionHandler.GetRoleLevel)
			permissionGroup.POST("/role/compare", permissionHandler.CompareRoles)
		}

		// 分享功能相关路由
		shareGroup := v1.Group("/share")
		{
			// 分享认证（无需认证）
			shareGroup.POST("/auth", shareHandler.ShareAuth)

			// 分享视图访问（无需认证，但需要分享权限验证）
			shareGroup.GET("/:share_id/view", shareHandler.GetShareView)
			shareGroup.POST("/:share_id/form-submit", shareHandler.SubmitForm)
			shareGroup.POST("/:share_id/copy", shareHandler.CopyData)
			shareGroup.GET("/:share_id/collaborators", shareHandler.GetCollaborators)
			shareGroup.GET("/:share_id/link-records", shareHandler.GetLinkRecords)
		}

		// 分享管理相关路由（需要认证）
		shareManageGroup := v1.Group("/share-manage")
		shareManageGroup.Use(middleware.AuthMiddleware(config.AuthService))
		{
			// 分享管理
			shareManageGroup.POST("/:share_id/enable", shareHandler.EnableShareView)
			shareManageGroup.POST("/:share_id/disable", shareHandler.DisableShareView)
			shareManageGroup.PUT("/:share_id/meta", shareHandler.UpdateShareMeta)
		}

		// 视图分享管理路由（需要认证）
		viewShareGroup := v1.Group("/view-shares")
		viewShareGroup.Use(middleware.AuthMiddleware(config.AuthService))
		{
			viewShareGroup.POST("/:view_id/share", shareHandler.CreateShareView)
		}

		// 表格分享统计路由（需要认证）
		tableShareGroup := v1.Group("/table-shares")
		tableShareGroup.Use(middleware.AuthMiddleware(config.AuthService))
		{
			tableShareGroup.GET("/:table_id/share/stats", shareHandler.GetShareStats)
		}

		// 文件管理相关路由（需要认证）
		attachmentGroup := v1.Group("/attachments")
		attachmentGroup.Use(middleware.AuthMiddleware(config.AuthService))
		{
			// 文件上传
			attachmentGroup.POST("/upload", attachmentHandler.UploadFile)

			// 文件读取
			attachmentGroup.GET("/:id/read", attachmentHandler.ReadFile)

			// 文件信息
			attachmentGroup.GET("/:id", attachmentHandler.GetAttachment)

			// 文件列表
			attachmentGroup.GET("", attachmentHandler.ListAttachments)

			// 文件删除
			attachmentGroup.DELETE("/:id", attachmentHandler.DeleteFile)

			// 文件签名生成
			attachmentGroup.POST("/signature", attachmentHandler.GenerateSignature)

			// 上传完成通知
			attachmentGroup.POST("/notify", attachmentHandler.NotifyUpload)

			// 附件统计
			attachmentGroup.GET("/stats", attachmentHandler.GetAttachmentStats)

			// 清理过期令牌
			attachmentGroup.POST("/cleanup", attachmentHandler.CleanupExpiredTokens)
		}

		// 通知管理相关路由（需要认证）
		notificationGroup := v1.Group("/notifications")
		notificationGroup.Use(middleware.AuthMiddleware(config.AuthService))
		{
			// 通知CRUD
			notificationGroup.POST("", notificationHandler.CreateNotification)
			notificationGroup.GET("", notificationHandler.ListNotifications)
			notificationGroup.GET("/:id", notificationHandler.GetNotification)
			notificationGroup.PUT("/:id", notificationHandler.UpdateNotification)
			notificationGroup.DELETE("/:id", notificationHandler.DeleteNotification)

			// 通知操作
			notificationGroup.POST("/mark-read", notificationHandler.MarkNotificationsRead)
			notificationGroup.POST("/send", notificationHandler.SendNotification)
			notificationGroup.POST("/send-to-subscribers", notificationHandler.SendNotificationToSubscribers)
			notificationGroup.POST("/cleanup", notificationHandler.CleanupExpiredNotifications)

			// 用户通知操作
			notificationGroup.POST("/user/:user_id/mark-all-read", notificationHandler.MarkAllNotificationsRead)
			notificationGroup.GET("/user/:user_id/stats", notificationHandler.GetNotificationStats)
		}

		// 通知模板管理相关路由（需要认证）
		templateGroup := v1.Group("/notification-templates")
		templateGroup.Use(middleware.AuthMiddleware(config.AuthService))
		{
			// 模板CRUD
			templateGroup.POST("", notificationHandler.CreateTemplate)
			templateGroup.GET("", notificationHandler.ListTemplates)
			templateGroup.GET("/:id", notificationHandler.GetTemplate)
			templateGroup.PUT("/:id", notificationHandler.UpdateTemplate)
			templateGroup.DELETE("/:id", notificationHandler.DeleteTemplate)

			// 按类型获取模板
			templateGroup.GET("/type/:type", notificationHandler.GetTemplateByType)
		}

		// 通知订阅管理相关路由（需要认证）
		subscriptionGroup := v1.Group("/notification-subscriptions")
		subscriptionGroup.Use(middleware.AuthMiddleware(config.AuthService))
		{
			// 订阅CRUD
			subscriptionGroup.POST("", notificationHandler.CreateSubscription)
			subscriptionGroup.GET("/:id", notificationHandler.GetSubscription)
			subscriptionGroup.PUT("/:id", notificationHandler.UpdateSubscription)
			subscriptionGroup.DELETE("/:id", notificationHandler.DeleteSubscription)

			// 用户订阅管理
			subscriptionGroup.GET("/user/:user_id", notificationHandler.GetUserSubscriptions)
			subscriptionGroup.DELETE("/user/:user_id", notificationHandler.DeleteUserSubscriptions)
		}

		// 搜索功能相关路由（需要认证）
		searchGroup := v1.Group("/search")
		searchGroup.Use(middleware.AuthMiddleware(config.AuthService))
		{
			// 搜索操作
			searchGroup.GET("", searchHandler.Search)
			searchGroup.POST("/advanced", searchHandler.AdvancedSearch)
			searchGroup.GET("/suggestions", searchHandler.SearchSuggestions)
			searchGroup.GET("/popular", searchHandler.GetPopularQueries)
			searchGroup.GET("/stats", searchHandler.GetSearchStats)

			// 搜索索引管理
			searchGroup.POST("/indexes", searchHandler.CreateIndex)
			searchGroup.GET("/indexes", searchHandler.ListIndexes)
			searchGroup.GET("/indexes/:id", searchHandler.GetIndex)
			searchGroup.PUT("/indexes/:id", searchHandler.UpdateIndex)
			searchGroup.DELETE("/indexes/:id", searchHandler.DeleteIndex)
			searchGroup.DELETE("/indexes/by-source", searchHandler.DeleteIndexesBySource)
			searchGroup.POST("/indexes/rebuild", searchHandler.RebuildIndex)
			searchGroup.POST("/indexes/optimize", searchHandler.OptimizeIndex)
			searchGroup.GET("/indexes/stats", searchHandler.GetIndexStats)
		}

		// 健康检查和信息路由 (无需认证)
		v1.GET("/info", InfoHandler)

		// 系统指标路由 (需要认证)
		systemGroup := v1.Group("/system")
		systemGroup.Use(middleware.AuthMiddleware(config.AuthService))
		systemGroup.Use(middleware.AdminRequiredMiddleware())
		{
			systemGroup.GET("/metrics", healthHandler.Metrics)
		}
	}
}

// HealthCheckHandler 健康检查处理器
func HealthCheckHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "healthy",
		"service": "teable-go-backend",
		"version": "1.0.0",
	})
}

// InfoHandler 服务信息处理器
func InfoHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"service":     "teable-go-backend",
		"version":     "1.0.0",
		"description": "Teable后端服务 - Go语言重构版本",
		"author":      "Teable Team",
		"license":     "AGPL-3.0",
	})
}
