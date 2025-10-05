package main

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"teable-go-backend/internal/config"
	"teable-go-backend/internal/container"
	"teable-go-backend/internal/domain/base"
	"teable-go-backend/internal/domain/record"
	"teable-go-backend/internal/domain/space"
	"teable-go-backend/internal/domain/table"
	"teable-go-backend/pkg/logger"
)

// helper: marshal any to JSON bytes
func toJSONBytes(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"error": "config load failed", "detail": err.Error()})
		os.Exit(1)
	}
	_ = logger.Init(logger.LoggerConfig{Level: cfg.Logger.Level, Format: cfg.Logger.Format, OutputPath: cfg.Logger.OutputPath})
	defer logger.Sync()

	c := container.NewContainer(cfg)
	if err := c.Initialize(); err != nil {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"error": "container init failed", "detail": err.Error()})
		os.Exit(1)
	}
	defer c.Close()

	// 创建 MCP 服务器，启用工具能力
	s := server.NewMCPServer("Teable MCP", "1.0.0", server.WithToolCapabilities(true))

	// 工具：createSpace
	s.AddTool(
		mcp.NewTool(
			"createSpace",
			mcp.WithDescription("创建空间"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) *mcp.CallToolResult {
			var p struct {
				Name        string  `json:"name"`
				Description *string `json:"description"`
				Icon        *string `json:"icon"`
				UserID      string  `json:"user_id"`
			}
			_ = json.Unmarshal(toJSONBytes(req.Params.Arguments), &p)

			sp, err := c.SpaceService().CreateSpace(ctx, space.CreateSpaceRequest{
				Name:        p.Name,
				Description: p.Description,
				Icon:        p.Icon,
				CreatedBy:   p.UserID,
			})
			if err != nil {
				res, _ := mcp.NewToolResultError(err.Error())
				return res
			}
			b, _ := json.Marshal(sp)
			res, _ := mcp.NewToolResultJSON(b)
			return res
		},
	)

	// 工具：createBase
	s.AddTool(
		mcp.NewTool(
			"createBase",
			mcp.WithDescription("在指定空间创建基础表"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) *mcp.CallToolResult {
			var p struct {
				SpaceID     string  `json:"space_id"`
				Name        string  `json:"name"`
				Description *string `json:"description"`
				Icon        *string `json:"icon"`
				UserID      string  `json:"user_id"`
			}
			_ = json.Unmarshal(toJSONBytes(req.Params.Arguments), &p)
			resBase, err := c.BaseService().CreateBase(ctx, base.CreateBaseRequest{SpaceID: p.SpaceID, Name: p.Name, Description: p.Description, Icon: p.Icon, CreatedBy: p.UserID})
			if err != nil {
				res, _ := mcp.NewToolResultError(err.Error())
				return res
			}
			b, _ := json.Marshal(resBase)
			res, _ := mcp.NewToolResultJSON(b)
			return res
		},
	)

	// 工具：createTable
	s.AddTool(
		mcp.NewTool(
			"createTable",
			mcp.WithDescription("在基础表下创建数据表"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) *mcp.CallToolResult {
			var p struct {
				BaseID      string  `json:"base_id"`
				Name        string  `json:"name"`
				Description *string `json:"description"`
				Icon        *string `json:"icon"`
				UserID      string  `json:"user_id"`
			}
			_ = json.Unmarshal(toJSONBytes(req.Params.Arguments), &p)
			resTable, err := c.TableService().CreateTable(ctx, table.CreateTableRequest{BaseID: p.BaseID, Name: p.Name, Description: p.Description, Icon: p.Icon, CreatedBy: p.UserID})
			if err != nil {
				res, _ := mcp.NewToolResultError(err.Error())
				return res
			}
			b, _ := json.Marshal(resTable)
			res, _ := mcp.NewToolResultJSON(b)
			return res
		},
	)

	// 工具：createField
	s.AddTool(
		mcp.NewTool(
			"createField",
			mcp.WithDescription("在数据表下创建字段"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) *mcp.CallToolResult {
			var p struct {
				TableID      string         `json:"table_id"`
				Name         string         `json:"name"`
				Type         string         `json:"type"`
				Description  *string        `json:"description"`
				Required     bool           `json:"required"`
				IsUnique     bool           `json:"is_unique"`
				IsPrimary    bool           `json:"is_primary"`
				DefaultValue *string        `json:"default_value"`
				FieldOptions map[string]any `json:"field_options"`
				FieldOrder   int            `json:"field_order"`
				UserID       string         `json:"user_id"`
			}
			_ = json.Unmarshal(toJSONBytes(req.Params.Arguments), &p)

			var optionsJSON *string
			if p.FieldOptions != nil {
				b, _ := json.Marshal(p.FieldOptions)
				s := string(b)
				optionsJSON = &s
			}
			resField, err := c.TableService().CreateField(ctx, table.CreateFieldRequest{
				TableID:      p.TableID,
				Name:         p.Name,
				Type:         p.Type,
				Description:  p.Description,
				IsRequired:   p.Required,
				IsUnique:     p.IsUnique,
				IsPrimary:    p.IsPrimary,
				DefaultValue: p.DefaultValue,
				Options:      optionsJSON,
				FieldOrder:   p.FieldOrder,
				CreatedBy:    p.UserID,
			})
			if err != nil {
				res, _ := mcp.NewToolResultError(err.Error())
				return res
			}
			b, _ := json.Marshal(resField)
			res, _ := mcp.NewToolResultJSON(b)
			return res
		},
	)

	// 工具：createRecord
	s.AddTool(
		mcp.NewTool(
			"createRecord",
			mcp.WithDescription("在数据表下创建记录"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) *mcp.CallToolResult {
			var p struct {
				TableID string                 `json:"table_id"`
				Data    map[string]interface{} `json:"data"`
				UserID  string                 `json:"user_id"`
			}
			_ = json.Unmarshal(toJSONBytes(req.Params.Arguments), &p)
			ctx2, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()
			resRec, err := c.RecordAppService().CreateRecord(ctx2, record.CreateRecordRequest{TableID: p.TableID, Data: p.Data, CreatedBy: p.UserID}, p.UserID)
			if err != nil {
				res, _ := mcp.NewToolResultError(err.Error())
				return res
			}
			b, _ := json.Marshal(resRec)
			res, _ := mcp.NewToolResultJSON(b)
			return res
		},
	)

	// 以 stdio 方式提供 MCP 服务
	if err := server.ServeStdio(s); err != nil {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"error": "serve stdio failed", "detail": err.Error()})
		os.Exit(1)
	}
}
