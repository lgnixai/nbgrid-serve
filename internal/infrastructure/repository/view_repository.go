package repository

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"teable-go-backend/internal/domain/view"
	"teable-go-backend/internal/infrastructure/database/models"
	"teable-go-backend/pkg/errors"
)

// ViewRepository 视图仓储实现
type ViewRepository struct {
	db *gorm.DB
}

// NewViewRepository 创建新的视图仓储
func NewViewRepository(db *gorm.DB) view.Repository {
	return &ViewRepository{db: db}
}

// Create 创建视图
func (r *ViewRepository) Create(ctx context.Context, v *view.View) error {
	model := r.domainToModel(v)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return r.handleDBError(err)
	}
	return nil
}

// GetByID 通过ID获取视图
func (r *ViewRepository) GetByID(ctx context.Context, id string) (*view.View, error) {
	var model models.View
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, r.handleDBError(err)
	}
	return r.modelToDomain(&model), nil
}

// Update 更新视图
func (r *ViewRepository) Update(ctx context.Context, v *view.View) error {
	model := r.domainToModel(v)
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return r.handleDBError(err)
	}
	return nil
}

// Delete 删除视图 (软删除)
func (r *ViewRepository) Delete(ctx context.Context, id string) error {
	// GORM的软删除会自动处理DeletedAt字段
	if err := r.db.WithContext(ctx).Delete(&models.View{}, "id = ?", id).Error; err != nil {
		return r.handleDBError(err)
	}
	return nil
}

// List 列出视图
func (r *ViewRepository) List(ctx context.Context, filter view.ListViewFilter) ([]*view.View, error) {
	var modelViews []models.View
	query := r.db.WithContext(ctx).Model(&models.View{})

	if filter.TableID != nil {
		query = query.Where("table_id = ?", *filter.TableID)
	}
	if filter.Name != nil {
		query = query.Where("name LIKE ?", "%"+*filter.Name+"%")
	}
	if filter.Type != nil {
		query = query.Where("type = ?", *filter.Type)
	}
	if filter.CreatedBy != nil {
		query = query.Where("created_by = ?", *filter.CreatedBy)
	}
	if filter.Search != "" {
		like := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", like, like)
	}
	if filter.OrderBy != "" && filter.Order != "" {
		query = query.Order(fmt.Sprintf("%s %s", filter.OrderBy, filter.Order))
	} else {
		query = query.Order("created_time desc")
	}
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	if err := query.Find(&modelViews).Error; err != nil {
		return nil, r.handleDBError(err)
	}

	views := make([]*view.View, len(modelViews))
	for i, model := range modelViews {
		views[i] = r.modelToDomain(&model)
	}
	return views, nil
}

// Count 统计视图数量
func (r *ViewRepository) Count(ctx context.Context, filter view.ListViewFilter) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.View{})

	if filter.TableID != nil {
		query = query.Where("table_id = ?", *filter.TableID)
	}
	if filter.Name != nil {
		query = query.Where("name LIKE ?", "%"+*filter.Name+"%")
	}
	if filter.Type != nil {
		query = query.Where("type = ?", *filter.Type)
	}
	if filter.CreatedBy != nil {
		query = query.Where("created_by = ?", *filter.CreatedBy)
	}
	if filter.Search != "" {
		like := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", like, like)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, r.handleDBError(err)
	}
	return count, nil
}

// Exists 检查视图是否存在
func (r *ViewRepository) Exists(ctx context.Context, filter view.ListViewFilter) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.View{})

	if filter.TableID != nil {
		query = query.Where("table_id = ?", *filter.TableID)
	}
	if filter.Name != nil {
		query = query.Where("name = ?", *filter.Name)
	}
	if filter.Type != nil {
		query = query.Where("type = ?", *filter.Type)
	}
	if filter.CreatedBy != nil {
		query = query.Where("created_by = ?", *filter.CreatedBy)
	}

	if err := query.Count(&count).Error; err != nil {
		return false, r.handleDBError(err)
	}
	return count > 0, nil
}

// domainToModel 领域对象转数据模型
func (r *ViewRepository) domainToModel(v *view.View) *models.View {
	model := &models.View{
		ID:               v.ID,
		TableID:          v.TableID,
		Name:             v.Name,
		Description:      v.Description,
		Type:             v.Type,
		IsDefault:        v.IsDefault,
		CreatedBy:        v.CreatedBy,
		CreatedTime:      v.CreatedTime,
		LastModifiedTime: v.LastModifiedTime,
	}

	// 转换配置
	if err := model.SetConfigFromMap(v.Config); err != nil {
		// 如果转换失败，设置为空JSON
		model.Config = "{}"
	}

	return model
}

// modelToDomain 数据模型转领域对象
func (r *ViewRepository) modelToDomain(model *models.View) *view.View {
	config, err := model.GetConfigAsMap()
	if err != nil {
		// 如果转换失败，设置为空map
		config = make(map[string]interface{})
	}

	return &view.View{
		ID:               model.ID,
		TableID:          model.TableID,
		Name:             model.Name,
		Description:      model.Description,
		Type:             model.Type,
		Config:           config,
		IsDefault:        model.IsDefault,
		CreatedBy:        model.CreatedBy,
		CreatedTime:      model.CreatedTime,
		LastModifiedTime: model.LastModifiedTime,
	}
}

// GetGridViewData 获取网格视图数据
func (r *ViewRepository) GetGridViewData(ctx context.Context, req view.GridViewDataRequest) (*view.GridViewData, error) {
	// 获取视图信息
	viewInfo, err := r.GetByID(ctx, req.ViewID)
	if err != nil {
		return nil, err
	}
	if viewInfo == nil {
		return nil, errors.ErrNotFound.WithDetails("视图未找到")
	}

	// 解析网格视图配置
	var config view.GridViewConfig
	if viewInfo.Config != nil {
		// TODO: 将map转换为GridViewConfig结构体
		// 这里需要实现JSON序列化/反序列化
	}

	// 构建记录查询条件
	_ = view.ListViewFilter{
		TableID: &viewInfo.TableID,
		Limit:   req.PageSize,
		Offset:  (req.Page - 1) * req.PageSize,
	}

	// 这里需要调用记录仓储来获取数据
	// 由于视图仓储不应该直接依赖记录仓储，我们需要通过服务层来处理
	// 暂时返回一个空的数据结构
	data := &view.GridViewData{
		Records:  []map[string]interface{}{},
		Total:    0,
		Page:     req.Page,
		PageSize: req.PageSize,
		Columns:  config.Columns,
		Config:   config,
	}

	return data, nil
}

// GetFormViewData 获取表单视图数据
func (r *ViewRepository) GetFormViewData(ctx context.Context, req view.FormViewDataRequest) (*view.FormViewData, error) {
	// 获取视图信息
	viewInfo, err := r.GetByID(ctx, req.ViewID)
	if err != nil {
		return nil, err
	}
	if viewInfo == nil {
		return nil, errors.ErrNotFound.WithDetails("视图未找到")
	}

	// 解析表单视图配置
	var config view.FormViewConfig
	if viewInfo.Config != nil {
		// TODO: 将map转换为FormViewConfig结构体
		// 这里需要实现JSON序列化/反序列化
	}

	// 暂时返回一个空的数据结构
	data := &view.FormViewData{
		Fields: config.Fields,
		Config: config,
	}

	return data, nil
}

// GetKanbanViewData 获取看板视图数据
func (r *ViewRepository) GetKanbanViewData(ctx context.Context, req view.KanbanViewDataRequest) (*view.KanbanViewData, error) {
	// 获取视图信息
	viewInfo, err := r.GetByID(ctx, req.ViewID)
	if err != nil {
		return nil, err
	}
	if viewInfo == nil {
		return nil, errors.ErrNotFound.WithDetails("视图未找到")
	}

	// 解析看板视图配置
	var config view.KanbanViewConfig
	if viewInfo.Config != nil {
		// TODO: 将map转换为KanbanViewConfig结构体
		// 这里需要实现JSON序列化/反序列化
	}

	// 暂时返回一个空的数据结构
	data := &view.KanbanViewData{
		Groups: []view.KanbanGroup{},
		Config: config,
	}

	return data, nil
}

// GetCalendarViewData 获取日历视图数据
func (r *ViewRepository) GetCalendarViewData(ctx context.Context, req view.CalendarViewDataRequest) (*view.CalendarViewData, error) {
	// 获取视图信息
	viewInfo, err := r.GetByID(ctx, req.ViewID)
	if err != nil {
		return nil, err
	}
	if viewInfo == nil {
		return nil, errors.ErrNotFound.WithDetails("视图未找到")
	}

	// 解析日历视图配置
	var config view.CalendarViewConfig
	if viewInfo.Config != nil {
		// TODO: 将map转换为CalendarViewConfig结构体
		// 这里需要实现JSON序列化/反序列化
	}

	// 暂时返回一个空的数据结构
	data := &view.CalendarViewData{
		Events: []view.CalendarEvent{},
		Config: config,
	}

	return data, nil
}

// GetGalleryViewData 获取画廊视图数据
func (r *ViewRepository) GetGalleryViewData(ctx context.Context, req view.GalleryViewDataRequest) (*view.GalleryViewData, error) {
	// 获取视图信息
	viewInfo, err := r.GetByID(ctx, req.ViewID)
	if err != nil {
		return nil, err
	}
	if viewInfo == nil {
		return nil, errors.ErrNotFound.WithDetails("视图未找到")
	}

	// 解析画廊视图配置
	var config view.GalleryViewConfig
	if viewInfo.Config != nil {
		// TODO: 将map转换为GalleryViewConfig结构体
		// 这里需要实现JSON序列化/反序列化
	}

	// 暂时返回一个空的数据结构
	data := &view.GalleryViewData{
		Cards:    []view.GalleryCard{},
		Total:    0,
		Page:     req.Page,
		PageSize: req.PageSize,
		Config:   config,
	}

	return data, nil
}

// handleDBError 处理数据库错误
func (r *ViewRepository) handleDBError(err error) error {
	// TODO: 根据具体的数据库错误类型返回对应的业务错误
	if strings.Contains(err.Error(), "duplicate key") {
		if strings.Contains(err.Error(), "name") {
			return errors.ErrResourceExists.WithDetails("视图名称已存在")
		}
	}

	return errors.ErrDatabaseOperation.WithDetails(err.Error())
}
