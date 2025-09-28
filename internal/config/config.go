package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Storage   StorageConfig   `mapstructure:"storage"`
	Logger    LoggerConfig    `mapstructure:"logger"`
	Queue     QueueConfig     `mapstructure:"queue"`
	WebSocket WebSocketConfig `mapstructure:"websocket"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Mode            string        `mapstructure:"mode"` // debug, release, test
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	MaxHeaderBytes  int           `mapstructure:"max_header_bytes"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	EnableCORS      bool          `mapstructure:"enable_cors"`
	EnableSwagger   bool          `mapstructure:"enable_swagger"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Name            string        `mapstructure:"name"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	LogLevel        string        `mapstructure:"log_level"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host        string        `mapstructure:"host"`
	Port        int           `mapstructure:"port"`
	Password    string        `mapstructure:"password"`
	DB          int           `mapstructure:"db"`
	PoolSize    int           `mapstructure:"pool_size"`
	DialTimeout time.Duration `mapstructure:"dial_timeout"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret          string        `mapstructure:"secret"`
	AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl"`
	Issuer          string        `mapstructure:"issuer"`
	EnableRefresh   bool          `mapstructure:"enable_refresh"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type      string      `mapstructure:"type"` // local, s3, minio
	Local     LocalConfig `mapstructure:"local"`
	S3        S3Config    `mapstructure:"s3"`
	CDNDomain string      `mapstructure:"cdn_domain"`
	UploadPath string     `mapstructure:"upload_path"` // 兼容性字段
}

type LocalConfig struct {
	UploadPath string `mapstructure:"upload_path"`
	URLPrefix  string `mapstructure:"url_prefix"`
}

type S3Config struct {
	Region    string `mapstructure:"region"`
	Bucket    string `mapstructure:"bucket"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Endpoint  string `mapstructure:"endpoint"`
	UseSSL    bool   `mapstructure:"use_ssl"`
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"` // json, console
	OutputPath string `mapstructure:"output_path"`
}

// QueueConfig 队列配置
type QueueConfig struct {
	RedisAddr     string `mapstructure:"redis_addr"`
	RedisPassword string `mapstructure:"redis_password"`
	RedisDB       int    `mapstructure:"redis_db"`
	MaxRetries    int    `mapstructure:"max_retries"`
	QueueCritical string `mapstructure:"queue_critical"`
	QueueDefault  string `mapstructure:"queue_default"`
	QueueLow      string `mapstructure:"queue_low"`
}

// WebSocketConfig WebSocket配置
type WebSocketConfig struct {
	EnableRedisPubSub bool          `mapstructure:"enable_redis_pubsub"`
	RedisPrefix       string        `mapstructure:"redis_prefix"`
	HeartbeatInterval time.Duration `mapstructure:"heartbeat_interval"`
	ConnectionTimeout time.Duration `mapstructure:"connection_timeout"`
	MaxConnections    int           `mapstructure:"max_connections"`
	EnablePresence    bool          `mapstructure:"enable_presence"`
}

// Load 加载配置
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("./deployments")

	// 设置环境变量前缀
	viper.SetEnvPrefix("TEABLE")
	viper.AutomaticEnv()
	
	// 设置环境变量键替换规则，避免系统环境变量干扰
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 设置默认值
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// 配置文件不存在时使用默认值和环境变量
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// setDefaults 设置默认配置值
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 3000)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "120s")
	viper.SetDefault("server.max_header_bytes", 1<<20) // 1MB
	viper.SetDefault("server.shutdown_timeout", "10s")
	viper.SetDefault("server.enable_cors", true)
	viper.SetDefault("server.enable_swagger", true)

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.name", "teable")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.max_open_conns", 100)
	viper.SetDefault("database.conn_max_lifetime", "1h")
	viper.SetDefault("database.log_level", "info")

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.dial_timeout", "5s")

	// JWT defaults
	viper.SetDefault("jwt.secret", "your-secret-key")
	viper.SetDefault("jwt.access_token_ttl", "24h")
	viper.SetDefault("jwt.refresh_token_ttl", "168h") // 7 days
	viper.SetDefault("jwt.issuer", "teable-api")
	viper.SetDefault("jwt.enable_refresh", true)

	// Storage defaults
	viper.SetDefault("storage.type", "local")
	viper.SetDefault("storage.local.upload_path", "./uploads")
	viper.SetDefault("storage.local.url_prefix", "/uploads")

	// Logger defaults
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.format", "json")
	viper.SetDefault("logger.output_path", "stdout")

	// Queue defaults
	viper.SetDefault("queue.redis_addr", "localhost:6379")
	viper.SetDefault("queue.redis_password", "")
	viper.SetDefault("queue.redis_db", 1)
	viper.SetDefault("queue.max_retries", 3)
	viper.SetDefault("queue.queue_critical", "critical")
	viper.SetDefault("queue.queue_default", "default")
	viper.SetDefault("queue.queue_low", "low")

	// WebSocket defaults
	viper.SetDefault("websocket.enable_redis_pubsub", true)
	viper.SetDefault("websocket.redis_prefix", "teable:ws")
	viper.SetDefault("websocket.heartbeat_interval", "30s")
	viper.SetDefault("websocket.connection_timeout", "60s")
	viper.SetDefault("websocket.max_connections", 1000)
	viper.SetDefault("websocket.enable_presence", true)
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	// 使用URL格式，GORM的PostgreSQL驱动对此格式支持更好
	if c.Password != "" {
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			c.User, c.Password, c.Host, c.Port, c.Name, c.SSLMode)
	}
	return fmt.Sprintf("postgres://%s@%s:%d/%s?sslmode=%s",
		c.User, c.Host, c.Port, c.Name, c.SSLMode)
}

// GetRedisAddr 获取Redis连接地址
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetServerAddr 获取服务器监听地址
func (c *ServerConfig) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
