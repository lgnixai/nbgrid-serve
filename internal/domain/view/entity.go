package view

import (
	"time"

	"teable-go-backend/pkg/utils"
)

// View 数据视图实体
type View struct {
	ID               string
	TableID          string
	Name             string
	Description      *string
	Type             string                 // 视图类型：grid, kanban, calendar, gallery等
	Config           map[string]interface{} // 视图配置，存储为JSON
	IsDefault        bool
	CreatedBy        string
	CreatedTime      time.Time
	DeletedTime      *time.Time
	LastModifiedTime *time.Time
}

// CreateViewRequest 创建视图请求
type CreateViewRequest struct {
	TableID     string
	Name        string
	Description *string
	Type        string
	Config      map[string]interface{}
	IsDefault   bool
	CreatedBy   string
}

// UpdateViewRequest 更新视图请求
type UpdateViewRequest struct {
	Name        *string
	Description *string
	Type        *string
	Config      map[string]interface{}
	IsDefault   *bool
}

// ListViewFilter 视图列表过滤条件
type ListViewFilter struct {
	TableID   *string
	Name      *string
	Type      *string
	CreatedBy *string
	Search    string
	OrderBy   string
	Order     string
	Limit     int
	Offset    int
}

// NewView 创建新的数据视图
func NewView(req CreateViewRequest) *View {
	now := time.Now()
	return &View{
		ID:               utils.GenerateViewID(),
		TableID:          req.TableID,
		Name:             req.Name,
		Description:      req.Description,
		Type:             req.Type,
		Config:           req.Config,
		IsDefault:        req.IsDefault,
		CreatedBy:        req.CreatedBy,
		CreatedTime:      now,
		LastModifiedTime: &now,
	}
}

// Update 更新视图信息
func (v *View) Update(req UpdateViewRequest) {
	if req.Name != nil {
		v.Name = *req.Name
	}
	if req.Description != nil {
		v.Description = req.Description
	}
	if req.Type != nil {
		v.Type = *req.Type
	}
	if req.Config != nil {
		v.Config = req.Config
	}
	if req.IsDefault != nil {
		v.IsDefault = *req.IsDefault
	}
	now := time.Now()
	v.LastModifiedTime = &now
}

// SoftDelete 软删除视图
func (v *View) SoftDelete() {
	now := time.Now()
	v.DeletedTime = &now
	v.LastModifiedTime = &now
}

// GridViewConfig 网格视图配置
type GridViewConfig struct {
	// 列配置
	Columns []GridViewColumn `json:"columns"`
	// 行配置
	RowHeight int `json:"row_height"` // 行高
	// 排序配置
	Sorts []GridViewSort `json:"sorts"`
	// 过滤配置
	Filters []GridViewFilter `json:"filters"`
	// 分组配置
	Groups []GridViewGroup `json:"groups"`
	// 显示配置
	ShowRowNumbers bool `json:"show_row_numbers"` // 显示行号
	ShowCheckboxes bool `json:"show_checkboxes"`  // 显示复选框
	// 分页配置
	PageSize int `json:"page_size"` // 每页显示数量
}

// GridViewColumn 网格视图列配置
type GridViewColumn struct {
	FieldID    string `json:"field_id"`   // 字段ID
	FieldName  string `json:"field_name"` // 字段名称
	FieldType  string `json:"field_type"` // 字段类型
	Width      int    `json:"width"`      // 列宽
	Visible    bool   `json:"visible"`    // 是否可见
	Frozen     bool   `json:"frozen"`     // 是否冻结
	Sortable   bool   `json:"sortable"`   // 是否可排序
	Filterable bool   `json:"filterable"` // 是否可过滤
	Order      int    `json:"order"`      // 显示顺序
}

// GridViewSort 网格视图排序配置
type GridViewSort struct {
	FieldID string `json:"field_id"` // 字段ID
	Order   string `json:"order"`    // 排序方向：asc, desc
}

// GridViewFilter 网格视图过滤配置
type GridViewFilter struct {
	FieldID  string      `json:"field_id"` // 字段ID
	Operator string      `json:"operator"` // 操作符：eq, ne, gt, gte, lt, lte, like, in, not_in, between
	Value    interface{} `json:"value"`    // 过滤值
	Logic    string      `json:"logic"`    // 逻辑关系：and, or
}

// GridViewGroup 网格视图分组配置
type GridViewGroup struct {
	FieldID   string `json:"field_id"`  // 字段ID
	Order     string `json:"order"`     // 分组顺序：asc, desc
	Collapsed bool   `json:"collapsed"` // 是否折叠
}

// KanbanViewConfig 看板视图配置
type KanbanViewConfig struct {
	GroupFieldID string   `json:"group_field_id"` // 分组字段ID
	CardFields   []string `json:"card_fields"`    // 卡片显示字段
	CardHeight   int      `json:"card_height"`    // 卡片高度
	ShowCount    bool     `json:"show_count"`     // 显示数量
}

// CalendarViewConfig 日历视图配置
type CalendarViewConfig struct {
	DateFieldID    string `json:"date_field_id"`    // 日期字段ID
	TitleFieldID   string `json:"title_field_id"`   // 标题字段ID
	ColorFieldID   string `json:"color_field_id"`   // 颜色字段ID
	StartTimeField string `json:"start_time_field"` // 开始时间字段
	EndTimeField   string `json:"end_time_field"`   // 结束时间字段
	AllDay         bool   `json:"all_day"`          // 全天事件
}

// GalleryViewConfig 画廊视图配置
type GalleryViewConfig struct {
	ImageFieldID    string   `json:"image_field_id"`    // 图片字段ID
	TitleFieldID    string   `json:"title_field_id"`    // 标题字段ID
	SubtitleFieldID string   `json:"subtitle_field_id"` // 副标题字段ID
	CardFields      []string `json:"card_fields"`       // 卡片显示字段
	CardSize        string   `json:"card_size"`         // 卡片大小：small, medium, large
	Columns         int      `json:"columns"`           // 列数
}

// FormViewConfig 表单视图配置
type FormViewConfig struct {
	Fields []FormViewField `json:"fields"` // 表单字段配置
	Layout string          `json:"layout"` // 布局：single, multi
}

// FormViewField 表单视图字段配置
type FormViewField struct {
	FieldID   string `json:"field_id"`   // 字段ID
	FieldName string `json:"field_name"` // 字段名称
	FieldType string `json:"field_type"` // 字段类型
	Required  bool   `json:"required"`   // 是否必填
	Visible   bool   `json:"visible"`    // 是否可见
	ReadOnly  bool   `json:"read_only"`  // 是否只读
	Width     int    `json:"width"`      // 字段宽度
	Order     int    `json:"order"`      // 显示顺序
}

// GridViewData 网格视图数据
type GridViewData struct {
	Records  []map[string]interface{} `json:"records"`   // 记录数据
	Total    int64                    `json:"total"`     // 总记录数
	Page     int                      `json:"page"`      // 当前页码
	PageSize int                      `json:"page_size"` // 每页大小
	Columns  []GridViewColumn         `json:"columns"`   // 列配置
	Config   GridViewConfig           `json:"config"`    // 视图配置
}

// GridViewDataRequest 网格视图数据请求
type GridViewDataRequest struct {
	ViewID   string           `json:"view_id"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
	Sorts    []GridViewSort   `json:"sorts"`
	Filters  []GridViewFilter `json:"filters"`
	Groups   []GridViewGroup  `json:"groups"`
}

// FormViewData 表单视图数据
type FormViewData struct {
	Fields []FormViewField `json:"fields"` // 表单字段配置
	Config FormViewConfig  `json:"config"` // 表单配置
}

// FormViewDataRequest 表单视图数据请求
type FormViewDataRequest struct {
	ViewID string `json:"view_id"`
}

// KanbanViewData 看板视图数据
type KanbanViewData struct {
	Groups []KanbanGroup    `json:"groups"` // 看板分组
	Config KanbanViewConfig `json:"config"` // 看板配置
}

// KanbanGroup 看板分组
type KanbanGroup struct {
	ID        string                   `json:"id"`        // 分组ID
	Name      string                   `json:"name"`      // 分组名称
	Value     interface{}              `json:"value"`     // 分组值
	Cards     []map[string]interface{} `json:"cards"`     // 卡片数据
	Count     int                      `json:"count"`     // 卡片数量
	Collapsed bool                     `json:"collapsed"` // 是否折叠
}

// KanbanViewDataRequest 看板视图数据请求
type KanbanViewDataRequest struct {
	ViewID string `json:"view_id"`
}

// MoveKanbanCardRequest 移动看板卡片请求
type MoveKanbanCardRequest struct {
	ViewID    string `json:"view_id"`
	RecordID  string `json:"record_id"`
	FromGroup string `json:"from_group"`
	ToGroup   string `json:"to_group"`
}

// CalendarViewData 日历视图数据
type CalendarViewData struct {
	Events []CalendarEvent    `json:"events"` // 日历事件
	Config CalendarViewConfig `json:"config"` // 日历配置
}

// CalendarEvent 日历事件
type CalendarEvent struct {
	ID        string                 `json:"id"`         // 事件ID（记录ID）
	Title     string                 `json:"title"`      // 事件标题
	StartTime string                 `json:"start_time"` // 开始时间
	EndTime   string                 `json:"end_time"`   // 结束时间
	AllDay    bool                   `json:"all_day"`    // 是否全天事件
	Color     string                 `json:"color"`      // 事件颜色
	Data      map[string]interface{} `json:"data"`       // 其他数据
}

// CalendarViewDataRequest 日历视图数据请求
type CalendarViewDataRequest struct {
	ViewID    string `json:"view_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// GalleryViewData 画廊视图数据
type GalleryViewData struct {
	Cards    []GalleryCard     `json:"cards"`     // 画廊卡片
	Total    int64             `json:"total"`     // 总卡片数
	Page     int               `json:"page"`      // 当前页码
	PageSize int               `json:"page_size"` // 每页大小
	Config   GalleryViewConfig `json:"config"`    // 画廊配置
}

// GalleryCard 画廊卡片
type GalleryCard struct {
	ID       string                 `json:"id"`       // 卡片ID（记录ID）
	Image    string                 `json:"image"`    // 图片URL
	Title    string                 `json:"title"`    // 标题
	Subtitle string                 `json:"subtitle"` // 副标题
	Data     map[string]interface{} `json:"data"`     // 其他数据
}

// GalleryViewDataRequest 画廊视图数据请求
type GalleryViewDataRequest struct {
	ViewID   string `json:"view_id"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}
