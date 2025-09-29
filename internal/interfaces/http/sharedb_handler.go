package http

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"teable-go-backend/internal/domain/sharedb"
	"teable-go-backend/pkg/errors"
	"teable-go-backend/pkg/response"
)

// ShareDBHandler ShareDB HTTP处理器
type ShareDBHandler struct {
	sharedbService       sharedb.ShareDB
	sharedbWSIntegration *sharedb.WebSocketIntegration
	logger               *zap.Logger
}

// NewShareDBHandler 创建ShareDB HTTP处理器
func NewShareDBHandler(sharedbService sharedb.ShareDB, sharedbWSIntegration *sharedb.WebSocketIntegration, logger *zap.Logger) *ShareDBHandler {
	return &ShareDBHandler{
		sharedbService:       sharedbService,
		sharedbWSIntegration: sharedbWSIntegration,
		logger:               logger,
	}
}

// GetShareDBStats 获取ShareDB统计信息
func (h *ShareDBHandler) GetShareDBStats(c *gin.Context) {
	stats := h.sharedbService.GetStats()

	response.SuccessWithMessage(c, stats, "")
}

// HandleSubmit 处理提交操作
func (h *ShareDBHandler) HandleSubmit(c *gin.Context) {
	var req SubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrBadRequest.WithDetails(err.Error()))
		return
	}

	// 从JWT中获取用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, errors.ErrUnauthorized)
		return
	}

	// 创建提交上下文
	context := &sharedb.SubmitContext{
		Agent: &sharedb.Agent{
			Connection: &sharedb.Connection{
				ID:        req.AgentID,
				UserID:    userID.(string),
				SessionID: req.SessionID,
				Metadata:  req.Metadata,
			},
			Custom: req.Custom,
		},
		Collection: req.Collection,
		ID:         req.ID,
		Op:         req.Op,
		Source:     req.Source,
	}

	// 处理提交
	var submitError error
	h.sharedbService.OnSubmit(context, func(err error) {
		submitError = err
	})

	if submitError != nil {
		h.handleError(c, submitError)
		return
	}

	response.SuccessWithMessage(c, map[string]string{"message": "Operation submitted successfully"}, "")
}

// GetSnapshot 获取快照
func (h *ShareDBHandler) GetSnapshot(c *gin.Context) {
	collection := c.Param("collection")
	id := c.Param("id")

	if collection == "" || id == "" {
		response.Error(c, errors.ErrBadRequest.WithDetails("Collection and ID are required"))
		return
	}

	snapshot, err := h.sharedbService.GetSnapshot(collection, id, nil, nil)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, snapshot, "")
}

// Query 查询文档
func (h *ShareDBHandler) Query(c *gin.Context) {
	collection := c.Param("collection")
	if collection == "" {
		response.Error(c, errors.ErrBadRequest.WithDetails("Collection is required"))
		return
	}

	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrBadRequest.WithDetails(err.Error()))
		return
	}

	// 转换查询请求
	query := &sharedb.Query{
		Fields: req.Fields,
		Sort:   make([]sharedb.SortField, len(req.Sort)),
		Limit:  req.Limit,
		Skip:   req.Skip,
	}

	for i, sort := range req.Sort {
		query.Sort[i] = sharedb.SortField{
			Field: sort.Field,
			Order: sort.Order,
		}
	}

	snapshots, extra, err := h.sharedbService.Query(collection, query, nil, nil)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, gin.H{"snapshots": snapshots, "extra": extra}, "")
}

// GetOps 获取操作
func (h *ShareDBHandler) GetOps(c *gin.Context) {
	collection := c.Param("collection")
	id := c.Param("id")

	if collection == "" || id == "" {
		response.Error(c, errors.ErrBadRequest.WithDetails("Collection and ID are required"))
		return
	}

	// 解析查询参数
	fromStr := c.DefaultQuery("from", "0")
	toStr := c.DefaultQuery("to", "0")

	from, err := strconv.Atoi(fromStr)
	if err != nil {
		response.Error(c, errors.ErrBadRequest.WithDetails("Invalid from parameter"))
		return
	}

	to, err := strconv.Atoi(toStr)
	if err != nil {
		response.Error(c, errors.ErrBadRequest.WithDetails("Invalid to parameter"))
		return
	}

	ops, err := h.sharedbService.GetOps(collection, id, from, to, nil)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.SuccessWithMessage(c, ops, "")
}

// handleError 处理错误
func (h *ShareDBHandler) handleError(c *gin.Context, err error) {
	response.Error(c, err)
}

// 请求结构定义

// SubmitRequest 提交请求
type SubmitRequest struct {
	AgentID    string                 `json:"agent_id" binding:"required"`
	SessionID  string                 `json:"session_id" binding:"required"`
	Collection string                 `json:"collection" binding:"required"`
	ID         string                 `json:"id" binding:"required"`
	Op         *sharedb.RawOperation  `json:"op" binding:"required"`
	Source     string                 `json:"source"`
	Metadata   map[string]interface{} `json:"metadata"`
	Custom     map[string]interface{} `json:"custom"`
}

// QueryRequest 查询请求
type QueryRequest struct {
	Fields map[string]interface{} `json:"fields,omitempty"`
	Sort   []SortField            `json:"sort,omitempty"`
	Limit  int                    `json:"limit,omitempty"`
	Skip   int                    `json:"skip,omitempty"`
}

// SortField 排序字段
type SortField struct {
	Field string `json:"field"`
	Order string `json:"order"` // "asc" or "desc"
}
