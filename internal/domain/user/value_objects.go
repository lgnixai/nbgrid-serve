package user

import (
	"encoding/json"
	"time"
)

// Email 邮箱值对象
type Email struct {
	value string
}

// NewEmail 创建邮箱值对象
func NewEmail(email string) (*Email, error) {
	if err := validateEmail(email); err != nil {
		return nil, err
	}

	return &Email{value: email}, nil
}

// Value 获取邮箱值
func (e Email) Value() string {
	return e.value
}

// String 实现Stringer接口
func (e Email) String() string {
	return e.value
}

// Equals 比较两个邮箱是否相等
func (e Email) Equals(other Email) bool {
	return e.value == other.value
}

// Phone 手机号值对象
type Phone struct {
	value string
}

// NewPhone 创建手机号值对象
func NewPhone(phone string) (*Phone, error) {
	if err := validatePhone(phone); err != nil {
		return nil, err
	}

	return &Phone{value: phone}, nil
}

// Value 获取手机号值
func (p Phone) Value() string {
	return p.value
}

// String 实现Stringer接口
func (p Phone) String() string {
	return p.value
}

// Equals 比较两个手机号是否相等
func (p Phone) Equals(other Phone) bool {
	return p.value == other.value
}

// UserPreferences 用户偏好设置值对象
type UserPreferences struct {
	Language      string            `json:"language"`
	Timezone      string            `json:"timezone"`
	DateFormat    string            `json:"date_format"`
	TimeFormat    string            `json:"time_format"`
	Theme         string            `json:"theme"`
	Notifications NotificationPrefs `json:"notifications"`
	Display       DisplayPrefs      `json:"display"`
	Privacy       PrivacyPrefs      `json:"privacy"`
}

// NotificationPrefs 通知偏好
type NotificationPrefs struct {
	EmailNotifications bool `json:"email_notifications"`
	PushNotifications  bool `json:"push_notifications"`
	SpaceInvites       bool `json:"space_invites"`
	RecordUpdates      bool `json:"record_updates"`
	CommentMentions    bool `json:"comment_mentions"`
	SystemAlerts       bool `json:"system_alerts"`
}

// DisplayPrefs 显示偏好
type DisplayPrefs struct {
	DefaultViewType string `json:"default_view_type"`
	RecordsPerPage  int    `json:"records_per_page"`
	ShowGridLines   bool   `json:"show_grid_lines"`
	CompactMode     bool   `json:"compact_mode"`
	ShowFieldTypes  bool   `json:"show_field_types"`
	AutoSave        bool   `json:"auto_save"`
}

// PrivacyPrefs 隐私偏好
type PrivacyPrefs struct {
	ProfileVisibility   string `json:"profile_visibility"`
	ShowOnlineStatus    bool   `json:"show_online_status"`
	AllowDirectMessages bool   `json:"allow_direct_messages"`
	ShareAnalytics      bool   `json:"share_analytics"`
}

// NewDefaultUserPreferences 创建默认用户偏好设置
func NewDefaultUserPreferences() *UserPreferences {
	return &UserPreferences{
		Language:   "zh-CN",
		Timezone:   "Asia/Shanghai",
		DateFormat: "YYYY-MM-DD",
		TimeFormat: "24h",
		Theme:      "light",
		Notifications: NotificationPrefs{
			EmailNotifications: true,
			PushNotifications:  true,
			SpaceInvites:       true,
			RecordUpdates:      true,
			CommentMentions:    true,
			SystemAlerts:       true,
		},
		Display: DisplayPrefs{
			DefaultViewType: "grid",
			RecordsPerPage:  50,
			ShowGridLines:   true,
			CompactMode:     false,
			ShowFieldTypes:  false,
			AutoSave:        true,
		},
		Privacy: PrivacyPrefs{
			ProfileVisibility:   "public",
			ShowOnlineStatus:    true,
			AllowDirectMessages: true,
			ShareAnalytics:      false,
		},
	}
}

// Validate 验证用户偏好设置
func (p *UserPreferences) Validate() error {
	// 验证语言代码
	validLanguages := map[string]bool{
		"zh-CN": true, "zh-TW": true, "en-US": true, "en-GB": true,
		"ja-JP": true, "ko-KR": true, "fr-FR": true, "de-DE": true,
		"es-ES": true, "pt-BR": true, "ru-RU": true, "it-IT": true,
	}
	if !validLanguages[p.Language] {
		return DomainError{Code: "INVALID_LANGUAGE", Message: "invalid language code"}
	}

	// 验证时区
	if _, err := time.LoadLocation(p.Timezone); err != nil {
		return DomainError{Code: "INVALID_TIMEZONE", Message: "invalid timezone"}
	}

	// 验证日期格式
	validDateFormats := map[string]bool{
		"YYYY-MM-DD": true, "MM/DD/YYYY": true, "DD/MM/YYYY": true,
		"YYYY年MM月DD日": true, "MM月DD日": true,
	}
	if !validDateFormats[p.DateFormat] {
		return DomainError{Code: "INVALID_DATE_FORMAT", Message: "invalid date format"}
	}

	// 验证时间格式
	if p.TimeFormat != "12h" && p.TimeFormat != "24h" {
		return DomainError{Code: "INVALID_TIME_FORMAT", Message: "time format must be 12h or 24h"}
	}

	// 验证主题
	validThemes := map[string]bool{
		"light": true, "dark": true, "auto": true,
	}
	if !validThemes[p.Theme] {
		return DomainError{Code: "INVALID_THEME", Message: "invalid theme"}
	}

	// 验证显示偏好
	if err := p.Display.Validate(); err != nil {
		return err
	}

	// 验证隐私偏好
	if err := p.Privacy.Validate(); err != nil {
		return err
	}

	return nil
}

// Validate 验证显示偏好
func (d *DisplayPrefs) Validate() error {
	// 验证默认视图类型
	validViewTypes := map[string]bool{
		"grid": true, "kanban": true, "calendar": true, "gallery": true, "form": true,
	}
	if !validViewTypes[d.DefaultViewType] {
		return DomainError{Code: "INVALID_VIEW_TYPE", Message: "invalid default view type"}
	}

	// 验证每页记录数
	if d.RecordsPerPage < 10 || d.RecordsPerPage > 200 {
		return DomainError{Code: "INVALID_RECORDS_PER_PAGE", Message: "records per page must be between 10 and 200"}
	}

	return nil
}

// Validate 验证隐私偏好
func (p *PrivacyPrefs) Validate() error {
	// 验证资料可见性
	validVisibilities := map[string]bool{
		"public": true, "private": true, "friends": true,
	}
	if !validVisibilities[p.ProfileVisibility] {
		return DomainError{Code: "INVALID_PROFILE_VISIBILITY", Message: "invalid profile visibility"}
	}

	return nil
}

// ToJSON 转换为JSON字符串
func (p *UserPreferences) ToJSON() (string, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return "", DomainError{Code: "PREFERENCES_MARSHAL_FAILED", Message: "failed to marshal preferences to JSON"}
	}
	return string(data), nil
}

// FromJSON 从JSON字符串创建用户偏好设置
func UserPreferencesFromJSON(jsonStr string) (*UserPreferences, error) {
	if jsonStr == "" {
		return NewDefaultUserPreferences(), nil
	}

	var prefs UserPreferences
	if err := json.Unmarshal([]byte(jsonStr), &prefs); err != nil {
		return nil, DomainError{Code: "PREFERENCES_UNMARSHAL_FAILED", Message: "failed to unmarshal preferences from JSON"}
	}

	// 验证偏好设置
	if err := prefs.Validate(); err != nil {
		return nil, err
	}

	return &prefs, nil
}

// Merge 合并用户偏好设置
func (p *UserPreferences) Merge(other *UserPreferences) *UserPreferences {
	if other == nil {
		return p
	}

	merged := *p // 复制当前偏好设置

	// 合并非零值
	if other.Language != "" {
		merged.Language = other.Language
	}
	if other.Timezone != "" {
		merged.Timezone = other.Timezone
	}
	if other.DateFormat != "" {
		merged.DateFormat = other.DateFormat
	}
	if other.TimeFormat != "" {
		merged.TimeFormat = other.TimeFormat
	}
	if other.Theme != "" {
		merged.Theme = other.Theme
	}

	// 合并通知偏好
	merged.Notifications = p.Notifications.Merge(&other.Notifications)

	// 合并显示偏好
	merged.Display = p.Display.Merge(&other.Display)

	// 合并隐私偏好
	merged.Privacy = p.Privacy.Merge(&other.Privacy)

	return &merged
}

// Merge 合并通知偏好
func (n *NotificationPrefs) Merge(other *NotificationPrefs) NotificationPrefs {
	if other == nil {
		return *n
	}

	return NotificationPrefs{
		EmailNotifications: other.EmailNotifications,
		PushNotifications:  other.PushNotifications,
		SpaceInvites:       other.SpaceInvites,
		RecordUpdates:      other.RecordUpdates,
		CommentMentions:    other.CommentMentions,
		SystemAlerts:       other.SystemAlerts,
	}
}

// Merge 合并显示偏好
func (d *DisplayPrefs) Merge(other *DisplayPrefs) DisplayPrefs {
	if other == nil {
		return *d
	}

	merged := *d
	if other.DefaultViewType != "" {
		merged.DefaultViewType = other.DefaultViewType
	}
	if other.RecordsPerPage > 0 {
		merged.RecordsPerPage = other.RecordsPerPage
	}
	merged.ShowGridLines = other.ShowGridLines
	merged.CompactMode = other.CompactMode
	merged.ShowFieldTypes = other.ShowFieldTypes
	merged.AutoSave = other.AutoSave

	return merged
}

// Merge 合并隐私偏好
func (p *PrivacyPrefs) Merge(other *PrivacyPrefs) PrivacyPrefs {
	if other == nil {
		return *p
	}

	merged := *p
	if other.ProfileVisibility != "" {
		merged.ProfileVisibility = other.ProfileVisibility
	}
	merged.ShowOnlineStatus = other.ShowOnlineStatus
	merged.AllowDirectMessages = other.AllowDirectMessages
	merged.ShareAnalytics = other.ShareAnalytics

	return merged
}
