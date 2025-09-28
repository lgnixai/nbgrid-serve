package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"teable-go-backend/internal/application"
	"teable-go-backend/internal/config"
	"teable-go-backend/internal/domain/attachment"
	"teable-go-backend/internal/domain/base"
	"teable-go-backend/internal/domain/permission"
	"teable-go-backend/internal/domain/record"
	"teable-go-backend/internal/domain/share"
	"teable-go-backend/internal/domain/sharedb"
	"teable-go-backend/internal/domain/space"
	"teable-go-backend/internal/domain/table"
	"teable-go-backend/internal/domain/user"
	"teable-go-backend/internal/domain/view"
	"teable-go-backend/internal/domain/websocket"
	"teable-go-backend/internal/infrastructure/cache"
	"teable-go-backend/internal/infrastructure/database"
	"teable-go-backend/internal/infrastructure/database/models"
	"teable-go-backend/internal/infrastructure/pubsub"
	"teable-go-backend/internal/infrastructure/repository"
	sharedbInfra "teable-go-backend/internal/infrastructure/sharedb"
	"teable-go-backend/internal/infrastructure/storage"
	httpHandlers "teable-go-backend/internal/interfaces/http"
	"teable-go-backend/internal/interfaces/middleware"
	"teable-go-backend/pkg/errors"
	"teable-go-backend/pkg/logger"
)

// @title Teable API
// @version 1.0
// @description Teable后端API服务
// @termsOfService https://teable.ai/terms

// @contact.name API Support
// @contact.url https://teable.ai/support
// @contact.email support@teable.ai

// @license.name AGPL-3.0
// @license.url https://github.com/teableio/teable/blob/main/LICENSE

// @host localhost:3000
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	if err := logger.Init(logger.LoggerConfig{
		Level:      cfg.Logger.Level,
		Format:     cfg.Logger.Format,
		OutputPath: cfg.Logger.OutputPath,
	}); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting Teable Go Backend",
		logger.String("version", "1.0.0"),
		logger.String("mode", cfg.Server.Mode),
		logger.String("port", fmt.Sprintf("%d", cfg.Server.Port)),
	)

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 初始化数据库连接
	dbConn, err := database.NewConnection(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", logger.ErrorField(err))
	}
	defer dbConn.Close()

	// 初始化Redis连接
	redisClient, err := cache.NewRedisClient(cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to connect to redis", logger.ErrorField(err))
	}
	defer redisClient.Close()

	// 初始化服务依赖
	userRepo := repository.NewUserRepository(dbConn.GetDB())
	userDomainService := user.NewService(userRepo)
	userAppService := application.NewUserService(userDomainService, redisClient, cfg.JWT)
	authService := middleware.NewJWTAuthService(cfg.JWT, redisClient)

	// 空间依赖
	spaceRepo := repository.NewSpaceRepository(dbConn.GetDB())
	spaceDomainService := space.NewService(spaceRepo)

	// 基础表依赖
	baseRepo := repository.NewBaseRepository(dbConn.GetDB())
	baseDomainService := base.NewService(baseRepo)

	// 数据表依赖
	tableRepo := repository.NewTableRepository(dbConn.GetDB())
	tableDomainService := table.NewService(tableRepo)

	// 记录依赖
	recordRepo := repository.NewRecordRepository(dbConn.GetDB())
	recordDomainService := record.NewService(recordRepo)

	// 视图依赖
	viewRepo := repository.NewViewRepository(dbConn.GetDB())
	viewDomainService := view.NewService(viewRepo)

	// 权限依赖
	permissionRepo := repository.NewPermissionRepository(dbConn.GetDB(), logger.Logger)
	permissionService := permission.NewService(permissionRepo, logger.Logger)

	// 分享依赖
	shareRepo := repository.NewShareRepository(dbConn.GetDB(), logger.Logger)
	shareService := share.NewService(shareRepo, logger.Logger)

	// 文件管理依赖
	attachmentRepo := repository.NewAttachmentRepository(dbConn.GetDB(), logger.Logger)
	uploadPath := cfg.Storage.Local.UploadPath
	if uploadPath == "" {
		uploadPath = cfg.Storage.UploadPath // 兼容性字段
	}
	localStorage := storage.NewLocalStorage(uploadPath, logger.Logger)
	fileValidator := storage.NewFileValidator(logger.Logger)

	// 创建存储配置
	storageConfig := &attachment.StorageConfig{
		Type:         cfg.Storage.Type,
		LocalPath:    uploadPath,
		MaxFileSize:  100 * 1024 * 1024, // 100MB
		AllowedTypes: []string{"image/*", "video/*", "audio/*", "application/pdf", "text/*"},
	}

	// 创建缩略图配置
	thumbnailConfig := &attachment.ThumbnailConfig{
		Enabled:     true,
		SmallWidth:  150,
		SmallHeight: 150,
		LargeWidth:  300,
		LargeHeight: 300,
		Quality:     80,
		Format:      "jpeg",
	}

	// 创建简单的上传令牌仓储（内存实现）
	tokenRepo := &memoryUploadTokenRepository{}

	// 创建简单的缩略图生成器（占位符实现）
	thumbnailGenerator := &placeholderThumbnailGenerator{}

	attachmentService := attachment.NewService(attachmentRepo, tokenRepo, localStorage, thumbnailGenerator, fileValidator, storageConfig, thumbnailConfig, logger.Logger)

	// WebSocket依赖
	wsManager := websocket.NewManager(logger.Logger)

	// 创建WebSocket服务，根据配置决定是否启用Redis Pub/Sub
	var wsService websocket.Service
	var redisPubSub *pubsub.RedisPubSub
	if cfg.WebSocket.EnableRedisPubSub {
		// 创建Redis Pub/Sub服务
		var err error
		redisPubSub, err = pubsub.NewRedisPubSub(cfg.Redis, cfg.WebSocket.RedisPrefix)
		if err != nil {
			logger.Fatal("Failed to create Redis Pub/Sub service", logger.ErrorField(err))
		}
		defer redisPubSub.Close()

		// 创建带Redis集成的WebSocket服务
		wsService = websocket.NewServiceWithRedis(wsManager, redisPubSub, logger.Logger, cfg.WebSocket.RedisPrefix)
		logger.Info("WebSocket service initialized with Redis Pub/Sub",
			logger.String("prefix", cfg.WebSocket.RedisPrefix),
		)
	} else {
		// 创建普通WebSocket服务
		wsService = websocket.NewService(wsManager, logger.Logger)
		logger.Info("WebSocket service initialized without Redis Pub/Sub")
	}

	wsHandler := websocket.NewHandler(wsManager, logger.Logger)

	// 协作服务依赖
	var redisIntegration *websocket.RedisIntegration
	if cfg.WebSocket.EnableRedisPubSub {
		redisIntegration = websocket.NewRedisIntegration(redisPubSub, wsManager, logger.Logger, cfg.WebSocket.RedisPrefix)
	}
	collaborationService := websocket.NewCollaborationService(wsManager, redisIntegration, logger.Logger)

	// ShareDB依赖
	sharedbAdapter := sharedbInfra.NewAdapter(
		dbConn,
		recordRepo,
		viewRepo,
		tableRepo,
		logger.Logger,
	)

	// 创建ShareDB服务
	var sharedbService sharedb.ShareDB
	if cfg.WebSocket.EnableRedisPubSub {
		// 使用Redis Pub/Sub
		sharedbService = sharedb.NewService(sharedbAdapter, redisPubSub, logger.Logger)
		logger.Info("ShareDB service initialized with Redis Pub/Sub")
	} else {
		// 使用内存Pub/Sub
		memoryPubSub := sharedb.NewMemoryPubSub()
		sharedbService = sharedb.NewService(sharedbAdapter, memoryPubSub, logger.Logger)
		logger.Info("ShareDB service initialized with memory Pub/Sub")
	}
	defer sharedbService.Close()

	// 创建ShareDB与WebSocket的集成
	sharedbWSIntegration := sharedb.NewWebSocketIntegration(sharedbService.(*sharedb.Service), wsService, logger.Logger)

	// 启动WebSocket管理器
	wsCtx, wsCancel := context.WithCancel(context.Background())
	defer wsCancel()
	go wsManager.Run(wsCtx)

	// 启动协作服务清理任务
	go collaborationService.StartPresenceCleanup(wsCtx)

	// 创建Gin引擎
	router := setupRouter(cfg, dbConn, redisClient, userAppService, authService, spaceDomainService, baseDomainService, tableDomainService, recordDomainService, viewDomainService, permissionService, shareService, attachmentService, wsService, wsHandler, sharedbService, sharedbWSIntegration, collaborationService)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:           cfg.Server.GetServerAddr(),
		Handler:        router,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		IdleTimeout:    cfg.Server.IdleTimeout,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	// 启动服务器
	go func() {
		logger.Info("Server starting",
			logger.String("addr", server.Addr),
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", logger.ErrorField(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// 优雅关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", logger.ErrorField(err))
	}

	logger.Info("Server exited")
}

// 简单的内存上传令牌仓储实现
type memoryUploadTokenRepository struct {
	tokens map[string]*attachment.UploadToken
}

func (r *memoryUploadTokenRepository) CreateUploadToken(ctx context.Context, token *attachment.UploadToken) error {
	if r.tokens == nil {
		r.tokens = make(map[string]*attachment.UploadToken)
	}
	r.tokens[token.Token] = token
	return nil
}

func (r *memoryUploadTokenRepository) GetUploadToken(ctx context.Context, token string) (*attachment.UploadToken, error) {
	if r.tokens == nil {
		return nil, errors.ErrNotFound
	}
	t, exists := r.tokens[token]
	if !exists {
		return nil, errors.ErrNotFound
	}
	return t, nil
}

func (r *memoryUploadTokenRepository) DeleteUploadToken(ctx context.Context, token string) error {
	if r.tokens == nil {
		return nil
	}
	delete(r.tokens, token)
	return nil
}

func (r *memoryUploadTokenRepository) CleanupExpiredTokens(ctx context.Context) error {
	// 简单的实现，实际应该检查过期时间
	return nil
}

// 占位符缩略图生成器实现
type placeholderThumbnailGenerator struct{}

func (g *placeholderThumbnailGenerator) GenerateThumbnail(ctx context.Context, sourcePath, targetPath string, width, height int, quality int) error {
	// 占位符实现，实际应该生成缩略图
	return nil
}

func (g *placeholderThumbnailGenerator) GenerateThumbnails(ctx context.Context, sourcePath string, config *attachment.ThumbnailConfig) (map[string]string, error) {
	// 占位符实现，实际应该生成多种尺寸的缩略图
	return make(map[string]string), nil
}

func (g *placeholderThumbnailGenerator) IsSupported(mimeType string) bool {
	// 支持常见的图片格式
	return mimeType == "image/jpeg" || mimeType == "image/png" || mimeType == "image/gif"
}

// setupRouter 设置路由
func setupRouter(cfg *config.Config, dbConn *database.Connection, redisClient *cache.RedisClient, userService *application.UserService, authService middleware.AuthService, spaceService space.Service, baseService base.Service, tableService table.Service, recordService record.Service, viewService view.Service, permissionService permission.Service, shareService share.Service, attachmentService attachment.Service, wsService websocket.Service, wsHandler *websocket.Handler, sharedbService sharedb.ShareDB, sharedbWSIntegration *sharedb.WebSocketIntegration, collaborationService *websocket.CollaborationService) *gin.Engine {
	router := gin.New()

	// 基础中间件
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.LoggingMiddleware())
	router.Use(gin.Recovery())

	// CORS中间件
	if cfg.Server.EnableCORS {
		router.Use(middleware.CORS())
	}

	// 简单的ping检查
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// ORM 自动迁移（开发期）
	_ = dbConn.Migrate(&models.User{}, &models.Account{}, &models.Space{}, &models.SpaceCollaborator{}, &models.Base{}, &models.Table{}, &models.Field{}, &models.Record{}, &models.View{}, &models.Permission{}, &models.ShareView{}, &models.Attachment{})

	// 设置API路由
	httpHandlers.SetupRoutes(router, httpHandlers.RouterConfig{
		UserService:          userService,
		AuthService:          authService,
		SpaceService:         spaceService,
		BaseService:          baseService,
		TableService:         tableService,
		RecordService:        recordService,
		ViewService:          viewService,
		PermissionService:    permissionService,
		ShareService:         shareService,
		AttachmentService:    attachmentService,
		WebSocketService:     wsService,
		WebSocketHandler:     wsHandler,
		ShareDBService:       sharedbService,
		ShareDBWSIntegration: sharedbWSIntegration,
		CollaborationService: collaborationService,
		DB:                   dbConn,
		ErrorMonitor:         nil, // 暂时传nil，稍后修复
	})

	// Swagger文档
	if cfg.Server.EnableSwagger {
		// 这里需要添加swagger中间件
		// router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	}

	return router
}

// healthCheckHandler 健康检查处理器
func healthCheckHandler(dbConn *database.Connection, redisClient *cache.RedisClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		status := gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
		}

		// 检查数据库连接
		if err := dbConn.Health(); err != nil {
			status["database"] = "unhealthy"
			status["database_error"] = err.Error()
			status["status"] = "unhealthy"
		} else {
			status["database"] = "healthy"
		}

		// 检查Redis连接
		if err := redisClient.Health(ctx); err != nil {
			status["redis"] = "unhealthy"
			status["redis_error"] = err.Error()
			status["status"] = "unhealthy"
		} else {
			status["redis"] = "healthy"
		}

		httpStatus := 200
		if status["status"] == "unhealthy" {
			httpStatus = 503
		}

		c.JSON(httpStatus, status)
	}
}
