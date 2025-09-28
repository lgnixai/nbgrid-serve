package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"teable-go-backend/internal/domain/search"
	"teable-go-backend/internal/infrastructure/database/models"
	"teable-go-backend/pkg/errors"
)

// SearchRepository 搜索仓储实现
type SearchRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewSearchRepository 创建新的SearchRepository
func NewSearchRepository(db *gorm.DB, logger *zap.Logger) *SearchRepository {
	return &SearchRepository{
		db:     db,
		logger: logger,
	}
}

// CreateIndex 创建搜索索引
func (r *SearchRepository) CreateIndex(ctx context.Context, index *search.SearchIndex) error {
	model := r.domainToModel(index)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		r.logger.Error("Failed to create search index",
			zap.String("id", index.ID),
			zap.Error(err))
		return err
	}

	return nil
}

// GetIndex 获取搜索索引
func (r *SearchRepository) GetIndex(ctx context.Context, id string) (*search.SearchIndex, error) {
	var model models.SearchIndex

	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		r.logger.Error("Failed to get search index",
			zap.String("id", id),
			zap.Error(err))
		return nil, err
	}

	return r.modelToDomain(&model), nil
}

// UpdateIndex 更新搜索索引
func (r *SearchRepository) UpdateIndex(ctx context.Context, index *search.SearchIndex) error {
	model := r.domainToModel(index)

	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		r.logger.Error("Failed to update search index",
			zap.String("id", index.ID),
			zap.Error(err))
		return err
	}

	return nil
}

// DeleteIndex 删除搜索索引
func (r *SearchRepository) DeleteIndex(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.SearchIndex{}).Error; err != nil {
		r.logger.Error("Failed to delete search index",
			zap.String("id", id),
			zap.Error(err))
		return err
	}

	return nil
}

// DeleteIndexesBySource 根据来源删除搜索索引
func (r *SearchRepository) DeleteIndexesBySource(ctx context.Context, sourceID, sourceType string) error {
	if err := r.db.WithContext(ctx).Where("source_id = ? AND source_type = ?", sourceID, sourceType).
		Delete(&models.SearchIndex{}).Error; err != nil {
		r.logger.Error("Failed to delete search indexes by source",
			zap.String("source_id", sourceID),
			zap.String("source_type", sourceType),
			zap.Error(err))
		return err
	}

	return nil
}

// ListIndexes 列出搜索索引
func (r *SearchRepository) ListIndexes(ctx context.Context, searchType *search.SearchType, sourceID, sourceType string, page, pageSize int) ([]*search.SearchIndex, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.SearchIndex{})

	// 添加过滤条件
	if searchType != nil {
		query = query.Where("type = ?", string(*searchType))
	}
	if sourceID != "" {
		query = query.Where("source_id = ?", sourceID)
	}
	if sourceType != "" {
		query = query.Where("source_type = ?", sourceType)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("Failed to count search indexes", zap.Error(err))
		return nil, 0, err
	}

	// 分页
	offset := (page - 1) * pageSize
	query = query.Offset(offset).Limit(pageSize).Order("updated_time DESC")

	// 查询数据
	var models []models.SearchIndex
	if err := query.Find(&models).Error; err != nil {
		r.logger.Error("Failed to list search indexes", zap.Error(err))
		return nil, 0, err
	}

	// 转换为领域对象
	indexes := make([]*search.SearchIndex, len(models))
	for i, model := range models {
		indexes[i] = r.modelToDomain(&model)
	}

	return indexes, total, nil
}

// Search 搜索
func (r *SearchRepository) Search(ctx context.Context, req *search.SearchRequest) (*search.SearchResponse, error) {
	// 构建搜索查询
	query := r.buildSearchQuery(req)

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("Failed to count search results", zap.Error(err))
		return nil, err
	}

	// 分页
	offset := (req.Page - 1) * req.PageSize
	query = query.Offset(offset).Limit(req.PageSize)

	// 排序
	orderBy := fmt.Sprintf("%s %s", req.SortBy, req.SortOrder)
	query = query.Order(orderBy)

	// 查询数据
	var models []models.SearchIndex
	if err := query.Find(&models).Error; err != nil {
		r.logger.Error("Failed to search", zap.Error(err))
		return nil, err
	}

	// 转换为搜索结果
	results := make([]*search.SearchResult, len(models))
	for i, model := range models {
		results[i] = r.modelToSearchResult(&model, req.Query)
	}

	// 计算总页数
	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	return &search.SearchResponse{
		Results:     results,
		Total:       total,
		Page:        req.Page,
		PageSize:    req.PageSize,
		TotalPages:  totalPages,
		SearchType:  req.Type,
		SearchScope: req.Scope,
	}, nil
}

// AdvancedSearch 高级搜索
func (r *SearchRepository) AdvancedSearch(ctx context.Context, req *search.AdvancedSearchRequest) (*search.SearchResponse, error) {
	// 构建高级搜索查询
	query := r.buildAdvancedSearchQuery(req)

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("Failed to count advanced search results", zap.Error(err))
		return nil, err
	}

	// 分页
	offset := (req.Page - 1) * req.PageSize
	query = query.Offset(offset).Limit(req.PageSize)

	// 排序
	if len(req.Sorts) > 0 {
		var orderParts []string
		for _, sort := range req.Sorts {
			orderParts = append(orderParts, fmt.Sprintf("%s %s", sort.Field, sort.Order))
		}
		query = query.Order(strings.Join(orderParts, ", "))
	} else {
		query = query.Order("updated_time DESC")
	}

	// 查询数据
	var models []models.SearchIndex
	if err := query.Find(&models).Error; err != nil {
		r.logger.Error("Failed to perform advanced search", zap.Error(err))
		return nil, err
	}

	// 转换为搜索结果
	results := make([]*search.SearchResult, len(models))
	for i, model := range models {
		results[i] = r.modelToSearchResult(&model, strings.Join(req.Queries, " "))
	}

	// 计算总页数
	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	return &search.SearchResponse{
		Results:     results,
		Total:       total,
		Page:        req.Page,
		PageSize:    req.PageSize,
		TotalPages:  totalPages,
		SearchType:  req.Type,
		SearchScope: req.Scope,
	}, nil
}

// SearchSuggestions 搜索建议
func (r *SearchRepository) SearchSuggestions(ctx context.Context, query string, limit int) ([]*search.SearchSuggestion, error) {
	var models []models.SearchSuggestion

	if err := r.db.WithContext(ctx).Where("query LIKE ?", "%"+query+"%").
		Order("count DESC").
		Limit(limit).
		Find(&models).Error; err != nil {
		r.logger.Error("Failed to get search suggestions", zap.Error(err))
		return nil, err
	}

	// 转换为领域对象
	suggestions := make([]*search.SearchSuggestion, len(models))
	for i, model := range models {
		suggestions[i] = r.modelToSuggestion(&model)
	}

	return suggestions, nil
}

// GetPopularQueries 获取热门查询
func (r *SearchRepository) GetPopularQueries(ctx context.Context, limit int) ([]*search.SearchSuggestion, error) {
	var models []models.SearchSuggestion

	if err := r.db.WithContext(ctx).
		Order("count DESC").
		Limit(limit).
		Find(&models).Error; err != nil {
		r.logger.Error("Failed to get popular queries", zap.Error(err))
		return nil, err
	}

	// 转换为领域对象
	suggestions := make([]*search.SearchSuggestion, len(models))
	for i, model := range models {
		suggestions[i] = r.modelToSuggestion(&model)
	}

	return suggestions, nil
}

// GetSearchStats 获取搜索统计
func (r *SearchRepository) GetSearchStats(ctx context.Context) (*search.SearchStats, error) {
	stats := &search.SearchStats{
		SearchByType:  make(map[search.SearchType]int64),
		SearchByScope: make(map[search.SearchScope]int64),
	}

	// 获取总搜索数
	if err := r.db.WithContext(ctx).Model(&models.SearchSuggestion{}).
		Select("SUM(count) as total").
		Scan(&stats.TotalSearches).Error; err != nil {
		r.logger.Error("Failed to count total searches", zap.Error(err))
		return nil, err
	}

	// 获取总索引数
	if err := r.db.WithContext(ctx).Model(&models.SearchIndex{}).
		Count(&stats.TotalIndexes).Error; err != nil {
		r.logger.Error("Failed to count total indexes", zap.Error(err))
		return nil, err
	}

	// 按类型统计
	var typeStats []struct {
		Type  string `json:"type"`
		Count int64  `json:"count"`
	}
	if err := r.db.WithContext(ctx).Model(&models.SearchSuggestion{}).
		Select("type, SUM(count) as count").
		Where("type IS NOT NULL AND type != ''").
		Group("type").
		Scan(&typeStats).Error; err != nil {
		r.logger.Error("Failed to get searches by type", zap.Error(err))
		return nil, err
	}

	for _, stat := range typeStats {
		stats.SearchByType[search.SearchType(stat.Type)] = stat.Count
	}

	return stats, nil
}

// IncrementSearchCount 增加搜索计数
func (r *SearchRepository) IncrementSearchCount(ctx context.Context, query string, searchType search.SearchType, scope search.SearchScope) error {
	// 查找现有记录
	var model models.SearchSuggestion
	err := r.db.WithContext(ctx).Where("query = ?", query).First(&model).Error

	if err == gorm.ErrRecordNotFound {
		// 创建新记录
		model = models.SearchSuggestion{
			Query:      query,
			Count:      1,
			Type:       string(searchType),
			SourceType: string(scope),
		}
		return r.db.WithContext(ctx).Create(&model).Error
	} else if err != nil {
		return err
	}

	// 更新计数
	return r.db.WithContext(ctx).Model(&model).
		Update("count", model.Count+1).Error
}

// RebuildIndex 重建索引
func (r *SearchRepository) RebuildIndex(ctx context.Context, searchType *search.SearchType) error {
	// 这里可以实现重建索引的逻辑
	// 例如：清空现有索引，重新从数据源构建索引
	r.logger.Info("Rebuilding search index", zap.String("type", string(*searchType)))
	return nil
}

// OptimizeIndex 优化索引
func (r *SearchRepository) OptimizeIndex(ctx context.Context) error {
	// 这里可以实现索引优化的逻辑
	// 例如：清理无用索引，优化索引结构
	r.logger.Info("Optimizing search index")
	return nil
}

// GetIndexStats 获取索引统计
func (r *SearchRepository) GetIndexStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 获取总索引数
	var totalIndexes int64
	if err := r.db.WithContext(ctx).Model(&models.SearchIndex{}).Count(&totalIndexes).Error; err != nil {
		return nil, err
	}
	stats["total_indexes"] = totalIndexes

	// 按类型统计索引
	var typeStats []struct {
		Type  string `json:"type"`
		Count int64  `json:"count"`
	}
	if err := r.db.WithContext(ctx).Model(&models.SearchIndex{}).
		Select("type, COUNT(*) as count").
		Group("type").
		Scan(&typeStats).Error; err != nil {
		return nil, err
	}

	typeMap := make(map[string]int64)
	for _, stat := range typeStats {
		typeMap[stat.Type] = stat.Count
	}
	stats["by_type"] = typeMap

	return stats, nil
}

// buildSearchQuery 构建搜索查询
func (r *SearchRepository) buildSearchQuery(req *search.SearchRequest) *gorm.DB {
	query := r.db.Model(&models.SearchIndex{})

	// 基本搜索条件
	if req.Query != "" {
		searchTerm := "%" + req.Query + "%"
		query = query.Where("title LIKE ? OR content LIKE ?", searchTerm, searchTerm)
	}

	// 类型过滤
	if req.Type != "" {
		query = query.Where("type = ?", string(req.Type))
	}

	// 来源过滤
	if req.SourceID != "" {
		query = query.Where("source_id = ?", req.SourceID)
	}
	if req.SourceType != "" {
		query = query.Where("source_type = ?", req.SourceType)
	}

	// 用户过滤
	if req.UserID != "" {
		query = query.Where("user_id = ?", req.UserID)
	}

	// 空间过滤
	if req.SpaceID != "" {
		query = query.Where("space_id = ?", req.SpaceID)
	}

	// 表格过滤
	if req.TableID != "" {
		query = query.Where("table_id = ?", req.TableID)
	}

	// 字段过滤
	if len(req.FieldIDs) > 0 {
		query = query.Where("field_id IN ?", req.FieldIDs)
	}

	return query
}

// buildAdvancedSearchQuery 构建高级搜索查询
func (r *SearchRepository) buildAdvancedSearchQuery(req *search.AdvancedSearchRequest) *gorm.DB {
	query := r.db.Model(&models.SearchIndex{})

	// 多查询条件
	if len(req.Queries) > 0 {
		var conditions []string
		var args []interface{}
		for _, q := range req.Queries {
			conditions = append(conditions, "(title LIKE ? OR content LIKE ?)")
			args = append(args, "%"+q+"%", "%"+q+"%")
		}
		query = query.Where(strings.Join(conditions, " OR "), args...)
	}

	// 类型过滤
	if req.Type != "" {
		query = query.Where("type = ?", string(req.Type))
	}

	// 来源过滤
	if req.SourceID != "" {
		query = query.Where("source_id = ?", req.SourceID)
	}
	if req.SourceType != "" {
		query = query.Where("source_type = ?", req.SourceType)
	}

	// 用户过滤
	if req.UserID != "" {
		query = query.Where("user_id = ?", req.UserID)
	}

	// 空间过滤
	if req.SpaceID != "" {
		query = query.Where("space_id = ?", req.SpaceID)
	}

	// 表格过滤
	if req.TableID != "" {
		query = query.Where("table_id = ?", req.TableID)
	}

	// 字段过滤
	if len(req.FieldIDs) > 0 {
		query = query.Where("field_id IN ?", req.FieldIDs)
	}

	// 应用过滤器
	for _, filter := range req.Filters {
		switch filter.Operator {
		case "eq":
			query = query.Where(fmt.Sprintf("%s = ?", filter.Field), filter.Value)
		case "ne":
			query = query.Where(fmt.Sprintf("%s != ?", filter.Field), filter.Value)
		case "gt":
			query = query.Where(fmt.Sprintf("%s > ?", filter.Field), filter.Value)
		case "gte":
			query = query.Where(fmt.Sprintf("%s >= ?", filter.Field), filter.Value)
		case "lt":
			query = query.Where(fmt.Sprintf("%s < ?", filter.Field), filter.Value)
		case "lte":
			query = query.Where(fmt.Sprintf("%s <= ?", filter.Field), filter.Value)
		case "like":
			query = query.Where(fmt.Sprintf("%s LIKE ?", filter.Field), "%"+fmt.Sprintf("%v", filter.Value)+"%")
		}
	}

	return query
}

// domainToModel 领域对象转模型
func (r *SearchRepository) domainToModel(index *search.SearchIndex) *models.SearchIndex {
	model := &models.SearchIndex{
		ID:          index.ID,
		Type:        string(index.Type),
		Title:       index.Title,
		Content:     index.Content,
		SourceID:    index.SourceID,
		SourceType:  index.SourceType,
		SourceURL:   index.SourceURL,
		UserID:      index.UserID,
		SpaceID:     index.SpaceID,
		TableID:     index.TableID,
		FieldID:     index.FieldID,
		CreatedTime: index.CreatedTime,
		UpdatedTime: index.UpdatedTime,
	}

	// 序列化Keywords
	if index.Keywords != nil {
		if keywordsBytes, err := json.Marshal(index.Keywords); err == nil {
			model.Keywords = string(keywordsBytes)
		}
	}

	// 序列化Metadata
	if index.Metadata != nil {
		if metadataBytes, err := json.Marshal(index.Metadata); err == nil {
			model.Metadata = string(metadataBytes)
		}
	}

	// 序列化Permissions
	if index.Permissions != nil {
		if permissionsBytes, err := json.Marshal(index.Permissions); err == nil {
			model.Permissions = string(permissionsBytes)
		}
	}

	// 序列化Tags
	if index.Tags != nil {
		if tagsBytes, err := json.Marshal(index.Tags); err == nil {
			model.Tags = string(tagsBytes)
		}
	}

	return model
}

// modelToDomain 模型转领域对象
func (r *SearchRepository) modelToDomain(model *models.SearchIndex) *search.SearchIndex {
	index := &search.SearchIndex{
		ID:          model.ID,
		Type:        search.SearchType(model.Type),
		Title:       model.Title,
		Content:     model.Content,
		SourceID:    model.SourceID,
		SourceType:  model.SourceType,
		SourceURL:   model.SourceURL,
		UserID:      model.UserID,
		SpaceID:     model.SpaceID,
		TableID:     model.TableID,
		FieldID:     model.FieldID,
		CreatedTime: model.CreatedTime,
		UpdatedTime: model.UpdatedTime,
	}

	// 反序列化Keywords
	if model.Keywords != "" {
		var keywords []string
		if err := json.Unmarshal([]byte(model.Keywords), &keywords); err == nil {
			index.Keywords = keywords
		}
	}

	// 反序列化Metadata
	if model.Metadata != "" {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(model.Metadata), &metadata); err == nil {
			index.Metadata = metadata
		}
	}

	// 反序列化Permissions
	if model.Permissions != "" {
		var permissions []string
		if err := json.Unmarshal([]byte(model.Permissions), &permissions); err == nil {
			index.Permissions = permissions
		}
	}

	// 反序列化Tags
	if model.Tags != "" {
		var tags []string
		if err := json.Unmarshal([]byte(model.Tags), &tags); err == nil {
			index.Tags = tags
		}
	}

	return index
}

// modelToSearchResult 模型转搜索结果
func (r *SearchRepository) modelToSearchResult(model *models.SearchIndex, query string) *search.SearchResult {
	result := &search.SearchResult{
		ID:          model.ID,
		Type:        search.SearchType(model.Type),
		Title:       model.Title,
		Content:     model.Content,
		SourceID:    model.SourceID,
		SourceType:  model.SourceType,
		SourceURL:   model.SourceURL,
		CreatedTime: model.CreatedTime,
		UpdatedTime: model.UpdatedTime,
	}

	// 反序列化Metadata
	if model.Metadata != "" {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(model.Metadata), &metadata); err == nil {
			result.Metadata = metadata
		}
	}

	// 计算简单的匹配分数
	result.Score = r.calculateScore(model, query)

	return result
}

// modelToSuggestion 模型转搜索建议
func (r *SearchRepository) modelToSuggestion(model *models.SearchSuggestion) *search.SearchSuggestion {
	return &search.SearchSuggestion{
		ID:          model.ID,
		Query:       model.Query,
		Count:       model.Count,
		Type:        search.SearchType(model.Type),
		SourceID:    model.SourceID,
		SourceType:  model.SourceType,
		CreatedTime: model.CreatedTime,
		UpdatedTime: model.UpdatedTime,
	}
}

// calculateScore 计算匹配分数
func (r *SearchRepository) calculateScore(model *models.SearchIndex, query string) float64 {
	// 简单的分数计算逻辑
	// 在实际应用中，可以使用更复杂的算法

	score := 0.0
	query = strings.ToLower(query)
	title := strings.ToLower(model.Title)
	content := strings.ToLower(model.Content)

	// 标题匹配权重更高
	if strings.Contains(title, query) {
		score += 10.0
	}

	// 内容匹配
	if strings.Contains(content, query) {
		score += 5.0
	}

	// 关键词匹配
	if model.Keywords != "" {
		var keywords []string
		if err := json.Unmarshal([]byte(model.Keywords), &keywords); err == nil {
			for _, keyword := range keywords {
				if strings.Contains(strings.ToLower(keyword), query) {
					score += 3.0
				}
			}
		}
	}

	return score
}
