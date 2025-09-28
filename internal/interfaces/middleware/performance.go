package middleware

import (
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	RequestCount    int64            `json:"request_count"`
	TotalDuration   time.Duration    `json:"total_duration"`
	AvgDuration     time.Duration    `json:"avg_duration"`
	MaxDuration     time.Duration    `json:"max_duration"`
	MinDuration     time.Duration    `json:"min_duration"`
	MemoryUsage     MemoryStats      `json:"memory_usage"`
	ResponseSizes   map[string]int64 `json:"response_sizes"`
	StatusCodeCount map[int]int64    `json:"status_code_count"`
	LastUpdated     time.Time        `json:"last_updated"`
	mu              sync.RWMutex     `json:"-"`
}

// MemoryStats 内存统计
type MemoryStats struct {
	Alloc         uint64  `json:"alloc"`
	TotalAlloc    uint64  `json:"total_alloc"`
	Sys           uint64  `json:"sys"`
	NumGC         uint32  `json:"num_gc"`
	GCCPUFraction float64 `json:"gc_cpu_fraction"`
}

// PerformanceMonitor 性能监控器
type PerformanceMonitor struct {
	metrics   *PerformanceMetrics
	logger    *zap.Logger
	config    PerformanceConfig
	startTime time.Time
}

// PerformanceConfig 性能监控配置
type PerformanceConfig struct {
	// 是否启用性能监控
	Enabled bool
	// 是否记录详细指标
	DetailedMetrics bool
	// 慢请求阈值
	SlowRequestThreshold time.Duration
	// 是否启用内存监控
	EnableMemoryMonitoring bool
	// 指标收集间隔
	MetricsInterval time.Duration
	// 是否记录响应体大小
	LogResponseSize bool
}

// DefaultPerformanceConfig 默认配置
func DefaultPerformanceConfig() PerformanceConfig {
	return PerformanceConfig{
		Enabled:                true,
		DetailedMetrics:        true,
		SlowRequestThreshold:   100 * time.Millisecond,
		EnableMemoryMonitoring: true,
		MetricsInterval:        30 * time.Second,
		LogResponseSize:        true,
	}
}

// NewPerformanceMonitor 创建性能监控器
func NewPerformanceMonitor(config PerformanceConfig) *PerformanceMonitor {
	pm := &PerformanceMonitor{
		metrics: &PerformanceMetrics{
			ResponseSizes:   make(map[string]int64),
			StatusCodeCount: make(map[int]int64),
		},
		logger:    zap.L(),
		config:    config,
		startTime: time.Now(),
	}

	// 启动指标收集goroutine
	if config.EnableMemoryMonitoring {
		go pm.collectMetrics()
	}

	return pm
}

// PerformanceMiddleware 性能监控中间件
func (pm *PerformanceMonitor) PerformanceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !pm.config.Enabled {
			c.Next()
			return
		}

		start := time.Now()

		// 记录请求开始时的内存使用
		var memBefore runtime.MemStats
		if pm.config.EnableMemoryMonitoring {
			runtime.ReadMemStats(&memBefore)
		}

		// 包装响应写入器以捕获响应大小
		var responseSize int64
		writer := &responseSizeWriter{
			ResponseWriter: c.Writer,
			size:           &responseSize,
		}
		c.Writer = writer

		// 处理请求
		c.Next()

		// 计算处理时间
		duration := time.Since(start)

		// 记录请求结束时的内存使用
		var memAfter runtime.MemStats
		if pm.config.EnableMemoryMonitoring {
			runtime.ReadMemStats(&memAfter)
		}

		// 更新指标
		pm.updateMetrics(duration, c.Writer.Status(), responseSize, memBefore, memAfter)

		// 记录慢请求
		if duration > pm.config.SlowRequestThreshold {
			pm.logSlowRequest(c, duration, responseSize)
		}

		// 记录详细指标
		if pm.config.DetailedMetrics {
			pm.logDetailedMetrics(c, duration, responseSize, memBefore, memAfter)
		}
	}
}

// updateMetrics 更新性能指标
func (pm *PerformanceMonitor) updateMetrics(duration time.Duration, statusCode int, responseSize int64, memBefore, memAfter runtime.MemStats) {
	pm.metrics.mu.Lock()
	defer pm.metrics.mu.Unlock()

	// 更新请求计数
	pm.metrics.RequestCount++

	// 更新持续时间统计
	pm.metrics.TotalDuration += duration
	pm.metrics.AvgDuration = pm.metrics.TotalDuration / time.Duration(pm.metrics.RequestCount)

	if duration > pm.metrics.MaxDuration {
		pm.metrics.MaxDuration = duration
	}

	if pm.metrics.MinDuration == 0 || duration < pm.metrics.MinDuration {
		pm.metrics.MinDuration = duration
	}

	// 更新状态码计数
	pm.metrics.StatusCodeCount[statusCode]++

	// 更新响应大小统计
	path := pm.getRequestPath()
	pm.metrics.ResponseSizes[path] += responseSize

	// 更新内存使用统计
	if pm.config.EnableMemoryMonitoring {
		pm.metrics.MemoryUsage = MemoryStats{
			Alloc:         memAfter.Alloc,
			TotalAlloc:    memAfter.TotalAlloc,
			Sys:           memAfter.Sys,
			NumGC:         memAfter.NumGC,
			GCCPUFraction: memAfter.GCCPUFraction,
		}
	}

	pm.metrics.LastUpdated = time.Now()
}

// logSlowRequest 记录慢请求
func (pm *PerformanceMonitor) logSlowRequest(c *gin.Context, duration time.Duration, responseSize int64) {
	fields := []zap.Field{
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("query", c.Request.URL.RawQuery),
		zap.String("ip", c.ClientIP()),
		zap.Duration("duration", duration),
		zap.Int("status", c.Writer.Status()),
		zap.Int64("response_size", responseSize),
	}

	if userID, exists := c.Get("user_id"); exists {
		fields = append(fields, zap.String("user_id", userID.(string)))
	}

	if requestID, exists := c.Get("request_id"); exists {
		fields = append(fields, zap.String("request_id", requestID.(string)))
	}

	pm.logger.Warn("Slow request detected", fields...)
}

// logDetailedMetrics 记录详细指标
func (pm *PerformanceMonitor) logDetailedMetrics(c *gin.Context, duration time.Duration, responseSize int64, memBefore, memAfter runtime.MemStats) {
	fields := []zap.Field{
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.Duration("duration", duration),
		zap.Int("status", c.Writer.Status()),
		zap.Int64("response_size", responseSize),
		zap.Uint64("mem_alloc", memAfter.Alloc),
		zap.Uint64("mem_total_alloc", memAfter.TotalAlloc),
		zap.Uint32("num_gc", memAfter.NumGC),
	}

	pm.logger.Debug("Request metrics", fields...)
}

// getRequestPath 获取请求路径（用于分组）
func (pm *PerformanceMonitor) getRequestPath() string {
	// 简化实现，实际项目中可能需要更复杂的路径规范化
	return ""
}

// collectMetrics 收集系统指标
func (pm *PerformanceMonitor) collectMetrics() {
	ticker := time.NewTicker(pm.config.MetricsInterval)
	defer ticker.Stop()

	for range ticker.C {
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		pm.logger.Info("System metrics",
			zap.Uint64("alloc", memStats.Alloc),
			zap.Uint64("total_alloc", memStats.TotalAlloc),
			zap.Uint64("sys", memStats.Sys),
			zap.Uint32("num_gc", memStats.NumGC),
			zap.Float64("gc_cpu_fraction", memStats.GCCPUFraction),
			zap.Int("goroutines", runtime.NumGoroutine()),
		)
	}
}

// GetMetrics 获取当前性能指标
func (pm *PerformanceMonitor) GetMetrics() *PerformanceMetrics {
	pm.metrics.mu.RLock()
	defer pm.metrics.mu.RUnlock()

	// 返回指标的副本
	return &PerformanceMetrics{
		RequestCount:    pm.metrics.RequestCount,
		TotalDuration:   pm.metrics.TotalDuration,
		AvgDuration:     pm.metrics.AvgDuration,
		MaxDuration:     pm.metrics.MaxDuration,
		MinDuration:     pm.metrics.MinDuration,
		MemoryUsage:     pm.metrics.MemoryUsage,
		ResponseSizes:   copyMap(pm.metrics.ResponseSizes),
		StatusCodeCount: copyIntMap(pm.metrics.StatusCodeCount),
		LastUpdated:     pm.metrics.LastUpdated,
	}
}

// ResetMetrics 重置性能指标
func (pm *PerformanceMonitor) ResetMetrics() {
	pm.metrics.mu.Lock()
	defer pm.metrics.mu.Unlock()

	pm.metrics.RequestCount = 0
	pm.metrics.TotalDuration = 0
	pm.metrics.AvgDuration = 0
	pm.metrics.MaxDuration = 0
	pm.metrics.MinDuration = 0
	pm.metrics.ResponseSizes = make(map[string]int64)
	pm.metrics.StatusCodeCount = make(map[int]int64)
	pm.metrics.LastUpdated = time.Now()
}

// responseSizeWriter 响应大小写入器
type responseSizeWriter struct {
	gin.ResponseWriter
	size *int64
}

func (w *responseSizeWriter) Write(data []byte) (int, error) {
	*w.size += int64(len(data))
	return w.ResponseWriter.Write(data)
}

func (w *responseSizeWriter) WriteString(s string) (int, error) {
	*w.size += int64(len(s))
	return w.ResponseWriter.WriteString(s)
}

// copyMap 复制字符串到int64的映射
func copyMap(original map[string]int64) map[string]int64 {
	copied := make(map[string]int64)
	for k, v := range original {
		copied[k] = v
	}
	return copied
}

// copyIntMap 复制int到int64的映射
func copyIntMap(original map[int]int64) map[int]int64 {
	copied := make(map[int]int64)
	for k, v := range original {
		copied[k] = v
	}
	return copied
}

// RateLimiter 限流器
type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	config   RateLimitConfig
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	// 时间窗口
	Window time.Duration
	// 最大请求数
	MaxRequests int
	// 是否启用IP限流
	EnableIPLimit bool
	// 是否启用用户限流
	EnableUserLimit bool
}

// DefaultRateLimitConfig 默认限流配置
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Window:          1 * time.Minute,
		MaxRequests:     100,
		EnableIPLimit:   true,
		EnableUserLimit: true,
	}
}

// NewRateLimiter 创建限流器
func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		config:   config,
	}
}

// RateLimitMiddleware 限流中间件
func (rl *RateLimiter) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取标识符
		identifier := rl.getIdentifier(c)

		// 检查限流
		if !rl.isAllowed(identifier) {
			c.JSON(429, gin.H{
				"error":       "Too many requests",
				"code":        "RATE_LIMIT_EXCEEDED",
				"retry_after": int(rl.config.Window.Seconds()),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// getIdentifier 获取限流标识符
func (rl *RateLimiter) getIdentifier(c *gin.Context) string {
	// 优先使用用户ID，如果没有则使用IP
	if rl.config.EnableUserLimit {
		if userID, exists := c.Get("user_id"); exists {
			return "user:" + userID.(string)
		}
	}

	if rl.config.EnableIPLimit {
		return "ip:" + c.ClientIP()
	}

	return "default"
}

// isAllowed 检查是否允许请求
func (rl *RateLimiter) isAllowed(identifier string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.config.Window)

	// 获取该标识符的请求历史
	requests, exists := rl.requests[identifier]
	if !exists {
		rl.requests[identifier] = []time.Time{now}
		return true
	}

	// 清理过期请求
	var validRequests []time.Time
	for _, reqTime := range requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}

	// 检查是否超过限制
	if len(validRequests) >= rl.config.MaxRequests {
		return false
	}

	// 添加当前请求
	validRequests = append(validRequests, now)
	rl.requests[identifier] = validRequests

	return true
}

// CleanupExpired 清理过期的限流记录
func (rl *RateLimiter) CleanupExpired() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.config.Window)

	for identifier, requests := range rl.requests {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if reqTime.After(cutoff) {
				validRequests = append(validRequests, reqTime)
			}
		}

		if len(validRequests) == 0 {
			delete(rl.requests, identifier)
		} else {
			rl.requests[identifier] = validRequests
		}
	}
}
