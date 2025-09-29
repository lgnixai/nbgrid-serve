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

	"teable-go-backend/internal/config"
	"teable-go-backend/internal/container"
	"teable-go-backend/internal/infrastructure/database/models"
	httpHandlers "teable-go-backend/internal/interfaces/http"
	"teable-go-backend/internal/interfaces/middleware"
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

	// 创建依赖注入容器
	container := container.NewContainer(cfg)
	
	// 初始化所有依赖
	if err := container.Initialize(); err != nil {
		logger.Fatal("Failed to initialize container", logger.ErrorField(err))
	}
	defer container.Close()

	// 启动服务
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	container.StartServices(ctx)

	// 创建Gin引擎
	router := setupRouter(container)

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
	ctx, cancel = context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", logger.ErrorField(err))
	}

	logger.Info("Server exited")
}



// setupRouter 设置路由
func setupRouter(container *container.Container) *gin.Engine {
	cfg := container.Config()
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
	_ = container.DBConnection().Migrate(&models.User{}, &models.Account{}, &models.Space{}, &models.SpaceCollaborator{}, &models.Base{}, &models.Table{}, &models.Field{}, &models.Record{}, &models.View{}, &models.Permission{}, &models.ShareView{}, &models.Attachment{})

	// 设置API路由
	httpHandlers.SetupRoutes(router, httpHandlers.RouterConfig{
		UserService:          container.UserAppService(),
		AuthService:          container.MiddlewareAuthService(),
		SpaceService:         container.SpaceService(),
		BaseService:          container.BaseService(),
		TableService:         container.TableService(),
		RecordService:        container.RecordService(),
		ViewService:          container.ViewService(),
		PermissionService:    container.PermissionService(),
		ShareService:         container.ShareService(),
		AttachmentService:    container.AttachmentService(),
		WebSocketService:     container.WebSocketService(),
		WebSocketHandler:     container.WebSocketHandler(),
		ShareDBService:       container.ShareDBService(),
		ShareDBWSIntegration: container.ShareDBWSIntegration(),
		CollaborationService: container.CollaborationService(),
		DB:                   container.DBConnection(),
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
func healthCheckHandler(container *container.Container) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		status := gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
		}

		// 检查数据库连接
		if err := container.DBConnection().Health(); err != nil {
			status["database"] = "unhealthy"
			status["database_error"] = err.Error()
			status["status"] = "unhealthy"
		} else {
			status["database"] = "healthy"
		}

		// 检查Redis连接
		if err := container.RedisClient().Health(ctx); err != nil {
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
