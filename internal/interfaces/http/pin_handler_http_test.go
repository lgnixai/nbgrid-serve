package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"teable-go-backend/pkg/logger"
)

type minimalAPIResponse struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

func TestPinHandler_ListPins_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Initialize a no-op global logger to avoid nil panics
	logger.Logger = zap.NewNop()
	logger.Sugar = logger.Logger.Sugar()

	r := gin.New()
	h := NewPinHandler()

	r.GET("/api/pin/list", func(c *gin.Context) {
		c.Set("user_id", "u_test")
		h.ListPins(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/pin/list", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp minimalAPIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Code != 200000 {
		t.Fatalf("expected code 200000, got %d", resp.Code)
	}
}
