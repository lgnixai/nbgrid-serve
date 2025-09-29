package container

import (
	"context"
	"fmt"

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
	"teable-go-backend/internal/infrastructure/pubsub"
	"teable-go-backend/internal/infrastructure/repository"
	sharedbInfra "teable-go-backend/internal/infrastructure/sharedb"
	"teable-go-backend/internal/infrastructure/storage"
	"teable-go-backend/internal/interfaces/middleware"
	"teable-go-backend/pkg/logger"
)

// Container 依赖注入容器
type Container struct {
	config *config.Config

	// 基础设施
	dbConn      *database.Connection
	redisClient *cache.RedisClient
	redisPubSub *pubsub.RedisPubSub

	// 仓储层
	userRepo       user.Repository
	spaceRepo      space.Repository
	baseRepo       base.Repository
	tableRepo      table.Repository
	recordRepo     record.Repository
	viewRepo       view.Repository
	permissionRepo permission.Repository
	shareRepo      share.Repository
	attachmentRepo attachment.Repository

	// 领域服务
	userDomainService       user.Service
	spaceDomainService      space.Service
	baseDomainService       base.Service
	tableDomainService      table.Service
	recordDomainService     record.Service
	viewDomainService       view.Service
	permissionDomainService permission.Service
	shareDomainService      share.Service
	attachmentDomainService attachment.Service

	// 应用服务
	userAppService        *application.UserService
	authService           *application.AuthService
	recordAppService      *application.RecordService
	permissionAppService  *application.PermissionService
	middlewareAuthService middleware.AuthService

	// WebSocket和实时协作
	wsManager            *websocket.Manager
	wsService            websocket.Service
	wsHandler            *websocket.Handler
	collaborationService *websocket.CollaborationService

	// ShareDB
	sharedbAdapter       *sharedbInfra.Adapter
	sharedbService       sharedb.ShareDB
	sharedbWSIntegration *sharedb.WebSocketIntegration

	// 存储服务
	localStorage       *storage.LocalStorage
	fileValidator      *storage.FileValidator
	thumbnailGenerator attachment.ThumbnailGenerator
	uploadTokenRepo    attachment.UploadTokenRepository
}

// NewContainer 创建新的容器
func NewContainer(cfg *config.Config) *Container {
	return &Container{
		config: cfg,
	}
}

// Initialize 初始化所有依赖
func (c *Container) Initialize() error {
	// 初始化基础设施
	if err := c.initInfrastructure(); err != nil {
		return fmt.Errorf("failed to initialize infrastructure: %w", err)
	}

	// 初始化仓储层
	c.initRepositories()

	// 初始化领域服务
	c.initDomainServices()

	// 初始化应用服务
	c.initApplicationServices()

	// 初始化WebSocket和实时协作
	if err := c.initWebSocketServices(); err != nil {
		return fmt.Errorf("failed to initialize websocket services: %w", err)
	}

	// 初始化ShareDB
	if err := c.initShareDBServices(); err != nil {
		return fmt.Errorf("failed to initialize sharedb services: %w", err)
	}

	// 初始化存储服务
	if err := c.initStorageServices(); err != nil {
		return fmt.Errorf("failed to initialize storage services: %w", err)
	}

	return nil
}

// initInfrastructure 初始化基础设施
func (c *Container) initInfrastructure() error {
	// 初始化数据库连接
	dbConn, err := database.NewConnection(c.config.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	c.dbConn = dbConn

	// 初始化Redis连接
	redisClient, err := cache.NewRedisClient(c.config.Redis)
	if err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}
	c.redisClient = redisClient

	// 初始化Redis Pub/Sub（如果启用）
	if c.config.WebSocket.EnableRedisPubSub {
		redisPubSub, err := pubsub.NewRedisPubSub(c.config.Redis, c.config.WebSocket.RedisPrefix)
		if err != nil {
			return fmt.Errorf("failed to create redis pub/sub: %w", err)
		}
		c.redisPubSub = redisPubSub
	}

	return nil
}

// initRepositories 初始化仓储层
func (c *Container) initRepositories() {
	db := c.dbConn.GetDB()

	c.userRepo = repository.NewUserRepository(db)
	c.spaceRepo = repository.NewSpaceRepository(db, logger.Logger)
	c.baseRepo = repository.NewBaseRepository(db)
	c.tableRepo = repository.NewTableRepository(db)
	c.recordRepo = repository.NewRecordRepository(db)
	c.viewRepo = repository.NewViewRepository(db)
	c.permissionRepo = repository.NewPermissionRepository(db, logger.Logger)
	c.shareRepo = repository.NewShareRepository(db, logger.Logger)
	c.attachmentRepo = repository.NewAttachmentRepository(db, logger.Logger)
}

// initDomainServices 初始化领域服务
func (c *Container) initDomainServices() {
	c.userDomainService = user.NewService(c.userRepo)
	c.spaceDomainService = space.NewService(c.spaceRepo)
	c.baseDomainService = base.NewService(c.baseRepo)
	c.tableDomainService = table.NewService(c.tableRepo)
	c.recordDomainService = record.NewService(c.recordRepo)
	c.viewDomainService = view.NewService(c.viewRepo)
	c.permissionDomainService = permission.NewService(c.permissionRepo, logger.Logger)
	c.shareDomainService = share.NewService(c.shareRepo, logger.Logger)
}

// initApplicationServices 初始化应用服务
func (c *Container) initApplicationServices() {
	// 创建令牌服务
	tokenService := application.NewTokenService(c.config.JWT, c.redisClient)

	// 创建会话服务
	sessionService := application.NewSessionService(c.redisClient)

	// 创建用户应用服务
	c.userAppService = application.NewUserService(c.userDomainService, tokenService, sessionService, c.redisClient)

	// 创建新的认证服务
	c.authService = application.NewAuthService(tokenService, c.userDomainService, c.redisClient)

	// 创建记录应用服务
	c.recordAppService = application.NewRecordService(c.recordRepo, c.tableDomainService, c.permissionDomainService)

	// 创建权限应用服务
	c.permissionAppService = application.NewPermissionService(c.permissionDomainService, c.redisClient)

	// 创建中间件认证服务（保持兼容性）
	c.middlewareAuthService = middleware.NewJWTAuthService(c.config.JWT, c.redisClient)
}

// initWebSocketServices 初始化WebSocket服务
func (c *Container) initWebSocketServices() error {
	c.wsManager = websocket.NewManager(logger.Logger)

	// 创建WebSocket服务
	if c.config.WebSocket.EnableRedisPubSub && c.redisPubSub != nil {
		c.wsService = websocket.NewServiceWithRedis(c.wsManager, c.redisPubSub, logger.Logger, c.config.WebSocket.RedisPrefix)
		logger.Info("WebSocket service initialized with Redis Pub/Sub")
	} else {
		c.wsService = websocket.NewService(c.wsManager, logger.Logger)
		logger.Info("WebSocket service initialized without Redis Pub/Sub")
	}

	c.wsHandler = websocket.NewHandler(c.wsManager, logger.Logger)

	// 创建协作服务
	var redisIntegration *websocket.RedisIntegration
	if c.config.WebSocket.EnableRedisPubSub && c.redisPubSub != nil {
		redisIntegration = websocket.NewRedisIntegration(c.redisPubSub, c.wsManager, logger.Logger, c.config.WebSocket.RedisPrefix)
	}
	c.collaborationService = websocket.NewCollaborationService(c.wsManager, redisIntegration, logger.Logger)

	return nil
}

// initShareDBServices 初始化ShareDB服务
func (c *Container) initShareDBServices() error {
	c.sharedbAdapter = sharedbInfra.NewAdapter(
		c.dbConn,
		c.recordRepo,
		c.viewRepo,
		c.tableRepo,
		logger.Logger,
	)

	// 创建ShareDB服务
	if c.config.WebSocket.EnableRedisPubSub && c.redisPubSub != nil {
		c.sharedbService = sharedb.NewService(c.sharedbAdapter, c.redisPubSub, logger.Logger)
		logger.Info("ShareDB service initialized with Redis Pub/Sub")
	} else {
		memoryPubSub := sharedb.NewMemoryPubSub()
		c.sharedbService = sharedb.NewService(c.sharedbAdapter, memoryPubSub, logger.Logger)
		logger.Info("ShareDB service initialized with memory Pub/Sub")
	}

	// 创建ShareDB与WebSocket的集成
	c.sharedbWSIntegration = sharedb.NewWebSocketIntegration(c.sharedbService.(*sharedb.Service), c.wsService, logger.Logger)

	return nil
}

// initStorageServices 初始化存储服务
func (c *Container) initStorageServices() error {
	uploadPath := c.config.Storage.Local.UploadPath
	if uploadPath == "" {
		uploadPath = c.config.Storage.UploadPath // 兼容性字段
	}

	c.localStorage = storage.NewLocalStorage(uploadPath, logger.Logger)
	c.fileValidator = storage.NewFileValidator(logger.Logger)

	// 创建存储配置
	storageConfig := &attachment.AttachmentStorageConfig{
		Type:         c.config.Storage.Type,
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
	c.uploadTokenRepo = &memoryUploadTokenRepository{
		tokens: make(map[string]*attachment.UploadToken),
	}

	// 创建简单的缩略图生成器（占位符实现）
	c.thumbnailGenerator = &placeholderThumbnailGenerator{}

	c.attachmentDomainService = attachment.NewService(
		c.attachmentRepo,
		c.uploadTokenRepo,
		c.localStorage,
		c.thumbnailGenerator,
		c.fileValidator,
		storageConfig,
		thumbnailConfig,
		logger.Logger,
	)

	return nil
}

// StartServices 启动服务
func (c *Container) StartServices(ctx context.Context) {
	// 启动WebSocket管理器
	go c.wsManager.Run(ctx)

	// 启动协作服务清理任务
	go c.collaborationService.StartPresenceCleanup(ctx)
}

// Close 关闭所有连接
func (c *Container) Close() error {
	var errors []error

	if c.sharedbService != nil {
		c.sharedbService.Close()
	}

	if c.redisPubSub != nil {
		if err := c.redisPubSub.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close redis pub/sub: %w", err))
		}
	}

	if c.redisClient != nil {
		if err := c.redisClient.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close redis client: %w", err))
		}
	}

	if c.dbConn != nil {
		if err := c.dbConn.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close database connection: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors occurred while closing container: %v", errors)
	}

	return nil
}

// Getters for accessing services

func (c *Container) Config() *config.Config                       { return c.config }
func (c *Container) DBConnection() *database.Connection           { return c.dbConn }
func (c *Container) RedisClient() *cache.RedisClient              { return c.redisClient }
func (c *Container) UserAppService() *application.UserService     { return c.userAppService }
func (c *Container) AuthService() *application.AuthService        { return c.authService }
func (c *Container) RecordAppService() *application.RecordService { return c.recordAppService }
func (c *Container) PermissionAppService() *application.PermissionService {
	return c.permissionAppService
}
func (c *Container) MiddlewareAuthService() middleware.AuthService { return c.middlewareAuthService }
func (c *Container) SpaceService() space.Service                   { return c.spaceDomainService }
func (c *Container) BaseService() base.Service                     { return c.baseDomainService }
func (c *Container) TableService() table.Service                   { return c.tableDomainService }
func (c *Container) RecordService() record.Service                 { return c.recordDomainService }
func (c *Container) ViewService() view.Service                     { return c.viewDomainService }
func (c *Container) PermissionService() permission.Service         { return c.permissionDomainService }
func (c *Container) ShareService() share.Service                   { return c.shareDomainService }
func (c *Container) AttachmentService() attachment.Service         { return c.attachmentDomainService }
func (c *Container) WebSocketService() websocket.Service           { return c.wsService }
func (c *Container) WebSocketHandler() *websocket.Handler          { return c.wsHandler }
func (c *Container) ShareDBService() sharedb.ShareDB               { return c.sharedbService }
func (c *Container) ShareDBWSIntegration() *sharedb.WebSocketIntegration {
	return c.sharedbWSIntegration
}
func (c *Container) CollaborationService() *websocket.CollaborationService {
	return c.collaborationService
}

// 简单的内存上传令牌仓储实现
type memoryUploadTokenRepository struct {
	tokens map[string]*attachment.UploadToken
}

func (r *memoryUploadTokenRepository) CreateUploadToken(ctx context.Context, token *attachment.UploadToken) error {
	r.tokens[token.Token] = token
	return nil
}

func (r *memoryUploadTokenRepository) GetUploadToken(ctx context.Context, token string) (*attachment.UploadToken, error) {
	t, exists := r.tokens[token]
	if !exists {
		return nil, fmt.Errorf("token not found")
	}
	return t, nil
}

func (r *memoryUploadTokenRepository) DeleteUploadToken(ctx context.Context, token string) error {
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
