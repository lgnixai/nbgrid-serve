package sharedb

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"teable-go-backend/internal/domain/record"
	"teable-go-backend/internal/domain/sharedb"
	"teable-go-backend/internal/domain/table"
	"teable-go-backend/internal/domain/view"
	"teable-go-backend/internal/infrastructure/database"
	"teable-go-backend/pkg/logger"
)

// Adapter ShareDB数据库适配器
type Adapter struct {
	dbConn     *database.Connection
	recordRepo record.Repository
	// fieldRepo  repository.FieldRepository // TODO: 实现FieldRepository
	viewRepo  view.Repository
	tableRepo table.Repository
	logger    *zap.Logger
	closed    bool
}

// NewAdapter 创建ShareDB适配器
func NewAdapter(
	dbConn *database.Connection,
	recordRepo record.Repository,
	// fieldRepo repository.FieldRepository, // TODO: 实现FieldRepository
	viewRepo view.Repository,
	tableRepo table.Repository,
	logger *zap.Logger,
) *Adapter {
	return &Adapter{
		dbConn:     dbConn,
		recordRepo: recordRepo,
		// fieldRepo:  fieldRepo, // TODO: 实现FieldRepository
		viewRepo:  viewRepo,
		tableRepo: tableRepo,
		logger:    logger,
		closed:    false,
	}
}

// GetSnapshot 获取快照
func (a *Adapter) GetSnapshot(collection, id string, projection sharedb.Projection, options interface{}) (*sharedb.Snapshot, error) {
	if a.closed {
		return nil, fmt.Errorf("adapter is closed")
	}

	ctx := context.Background()

	// 根据集合类型获取快照
	docType := a.extractDocType(collection)
	collectionID := a.extractCollectionID(collection)

	switch docType {
	case "record":
		return a.getRecordSnapshot(ctx, collectionID, id, projection)
	case "field":
		return a.getFieldSnapshot(ctx, collectionID, id, projection)
	case "view":
		return a.getViewSnapshot(ctx, collectionID, id, projection)
	case "table":
		return a.getTableSnapshot(ctx, collectionID, id, projection)
	default:
		return nil, fmt.Errorf("unknown document type: %s", docType)
	}
}

// GetSnapshotBulk 批量获取快照
func (a *Adapter) GetSnapshotBulk(collection string, ids []string, projection sharedb.Projection, options interface{}) (map[string]*sharedb.Snapshot, error) {
	if a.closed {
		return nil, fmt.Errorf("adapter is closed")
	}

	ctx := context.Background()
	result := make(map[string]*sharedb.Snapshot)

	// 根据集合类型批量获取快照
	docType := a.extractDocType(collection)
	collectionID := a.extractCollectionID(collection)

	for _, id := range ids {
		var snapshot *sharedb.Snapshot
		var err error

		switch docType {
		case "record":
			snapshot, err = a.getRecordSnapshot(ctx, collectionID, id, projection)
		case "field":
			snapshot, err = a.getFieldSnapshot(ctx, collectionID, id, projection)
		case "view":
			snapshot, err = a.getViewSnapshot(ctx, collectionID, id, projection)
		case "table":
			snapshot, err = a.getTableSnapshot(ctx, collectionID, id, projection)
		default:
			err = fmt.Errorf("unknown document type: %s", docType)
		}

		if err != nil {
			a.logger.Error("Failed to get snapshot",
				logger.String("collection", collection),
				logger.String("id", id),
				logger.ErrorField(err),
			)
			// 创建空快照
			snapshot = &sharedb.Snapshot{
				ID:   id,
				V:    0,
				Type: "",
				Data: nil,
			}
		}

		result[id] = snapshot
	}

	return result, nil
}

// GetOps 获取操作
func (a *Adapter) GetOps(collection, id string, from, to int, options interface{}) ([]*sharedb.RawOperation, error) {
	if a.closed {
		return nil, fmt.Errorf("adapter is closed")
	}

	// 获取文档的当前版本和类型
	docType := a.extractDocType(collection)
	collectionID := a.extractCollectionID(collection)

	version, opType, err := a.getVersionAndType(docType, collectionID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get version and type: %w", err)
	}

	if opType == sharedb.OperationTypeDelete {
		return []*sharedb.RawOperation{}, nil
	}

	if from > version {
		return []*sharedb.RawOperation{}, nil
	}

	// 获取快照数据
	snapshot, err := a.GetSnapshot(collection, id, nil, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}

	if snapshot == nil || snapshot.Data == nil {
		return []*sharedb.RawOperation{}, nil
	}

	// 生成操作
	var ops []*sharedb.RawOperation

	if opType == sharedb.OperationTypeCreate {
		// 创建操作
		createOp := &sharedb.RawOperation{
			Src: a.generateRandomString(21),
			Seq: 1,
			V:   version,
			Create: &sharedb.CreateData{
				Type: "json0",
				Data: snapshot.Data,
			},
		}
		ops = append(ops, createOp)
	} else {
		// 编辑操作
		editOps := a.getOpsFromSnapshot(docType, snapshot.Data)
		gapVersion := to - from
		if to == 0 {
			gapVersion = version - from + 1
		}

		for i := 0; i < gapVersion; i++ {
			editOp := &sharedb.RawOperation{
				Src: a.generateRandomString(21),
				Seq: 1,
				V:   from + i,
			}

			if i == gapVersion-1 && len(editOps) > 0 {
				editOp.Op = editOps
			}

			ops = append(ops, editOp)
		}
	}

	return ops, nil
}

// Query 查询文档
func (a *Adapter) Query(collection string, query *sharedb.Query, projection sharedb.Projection, options interface{}) ([]*sharedb.Snapshot, interface{}, error) {
	if a.closed {
		return nil, nil, fmt.Errorf("adapter is closed")
	}

	ctx := context.Background()

	// 根据集合类型查询文档ID
	docType := a.extractDocType(collection)
	collectionID := a.extractCollectionID(collection)

	ids, extra, err := a.getDocIdsByQuery(ctx, docType, collectionID, query)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query documents: %w", err)
	}

	if len(ids) == 0 {
		return []*sharedb.Snapshot{}, extra, nil
	}

	// 批量获取快照
	snapshots, err := a.GetSnapshotBulk(collection, ids, projection, options)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get snapshots: %w", err)
	}

	// 转换为切片
	result := make([]*sharedb.Snapshot, 0, len(ids))
	for _, id := range ids {
		if snapshot, exists := snapshots[id]; exists {
			result = append(result, snapshot)
		}
	}

	return result, extra, nil
}

// Commit 提交操作
func (a *Adapter) Commit(collection, id string, op *sharedb.RawOperation) error {
	if a.closed {
		return fmt.Errorf("adapter is closed")
	}

	// 这里应该实现实际的数据提交逻辑
	// 暂时只记录日志
	a.logger.Info("Committing operation",
		logger.String("collection", collection),
		logger.String("id", id),
		logger.String("operation_type", a.getOperationType(op)),
	)

	return nil
}

// Close 关闭适配器
func (a *Adapter) Close() error {
	if a.closed {
		return nil
	}

	a.closed = true
	a.logger.Info("ShareDB adapter closed")
	return nil
}

// 辅助方法

// extractDocType 提取文档类型
func (a *Adapter) extractDocType(collection string) string {
	parts := a.splitCollection(collection)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// extractCollectionID 提取集合ID
func (a *Adapter) extractCollectionID(collection string) string {
	parts := a.splitCollection(collection)
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

// splitCollection 分割集合名称
func (a *Adapter) splitCollection(collection string) []string {
	var parts []string
	var current string

	for _, char := range collection {
		if char == '_' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

// getRecordSnapshot 获取记录快照
func (a *Adapter) getRecordSnapshot(ctx context.Context, tableID, recordID string, projection sharedb.Projection) (*sharedb.Snapshot, error) {
	// 这里应该调用实际的记录服务
	// 暂时返回模拟数据
	return &sharedb.Snapshot{
		ID:   recordID,
		V:    1,
		Type: "json0",
		Data: map[string]interface{}{
			"id":     recordID,
			"fields": map[string]interface{}{},
		},
	}, nil
}

// getFieldSnapshot 获取字段快照
func (a *Adapter) getFieldSnapshot(ctx context.Context, tableID, fieldID string, projection sharedb.Projection) (*sharedb.Snapshot, error) {
	// 这里应该调用实际的字段服务
	// 暂时返回模拟数据
	return &sharedb.Snapshot{
		ID:   fieldID,
		V:    1,
		Type: "json0",
		Data: map[string]interface{}{
			"id":   fieldID,
			"name": "Field Name",
			"type": "text",
		},
	}, nil
}

// getViewSnapshot 获取视图快照
func (a *Adapter) getViewSnapshot(ctx context.Context, tableID, viewID string, projection sharedb.Projection) (*sharedb.Snapshot, error) {
	// 这里应该调用实际的视图服务
	// 暂时返回模拟数据
	return &sharedb.Snapshot{
		ID:   viewID,
		V:    1,
		Type: "json0",
		Data: map[string]interface{}{
			"id":   viewID,
			"name": "View Name",
			"type": "grid",
		},
	}, nil
}

// getTableSnapshot 获取表快照
func (a *Adapter) getTableSnapshot(ctx context.Context, tableID, id string, projection sharedb.Projection) (*sharedb.Snapshot, error) {
	// 这里应该调用实际的表服务
	// 暂时返回模拟数据
	return &sharedb.Snapshot{
		ID:   id,
		V:    1,
		Type: "json0",
		Data: map[string]interface{}{
			"id":   id,
			"name": "Table Name",
		},
	}, nil
}

// getVersionAndType 获取版本和类型
func (a *Adapter) getVersionAndType(docType, collectionID, docID string) (int, sharedb.OperationType, error) {
	// 这里应该调用实际的服务获取版本和类型
	// 暂时返回模拟数据
	return 1, sharedb.OperationTypeEdit, nil
}

// getDocIdsByQuery 根据查询获取文档ID
func (a *Adapter) getDocIdsByQuery(ctx context.Context, docType, collectionID string, query *sharedb.Query) ([]string, interface{}, error) {
	// 这里应该调用实际的服务进行查询
	// 暂时返回模拟数据
	return []string{"doc1", "doc2"}, nil, nil
}

// getOpsFromSnapshot 从快照生成操作
func (a *Adapter) getOpsFromSnapshot(docType string, snapshot interface{}) []sharedb.OTOperation {
	// 这里应该根据文档类型和快照数据生成操作
	// 暂时返回空操作
	return []sharedb.OTOperation{}
}

// getOperationType 获取操作类型
func (a *Adapter) getOperationType(op *sharedb.RawOperation) string {
	if op.Create != nil {
		return "create"
	} else if op.Del {
		return "delete"
	} else if len(op.Op) > 0 {
		return "edit"
	}
	return "unknown"
}

// generateRandomString 生成随机字符串
func (a *Adapter) generateRandomString(length int) string {
	// 简单的随机字符串生成
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[time.Now().UnixNano()%int64(len(chars))]
	}
	return string(result)
}
