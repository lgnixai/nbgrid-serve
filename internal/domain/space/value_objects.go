package space

import (
	"encoding/json"
	"time"
)

// SpaceName 空间名称值对象
type SpaceName struct {
	value string
}

// NewSpaceName 创建空间名称值对象
func NewSpaceName(name string) (*SpaceName, error) {
	if err := validateSpaceName(name); err != nil {
		return nil, err
	}
	
	return &SpaceName{value: name}, nil
}

// Value 获取空间名称值
func (sn SpaceName) Value() string {
	return sn.value
}

// String 实现Stringer接口
func (sn SpaceName) String() string {
	return sn.value
}

// Equals 比较两个空间名称是否相等
func (sn SpaceName) Equals(other SpaceName) bool {
	return sn.value == other.value
}

// SpaceSettings 空间设置值对象
type SpaceSettings struct {
	AllowPublicAccess    bool                   `json:"allow_public_access"`
	DefaultRole          CollaboratorRole       `json:"default_role"`
	EnableComments       bool                   `json:"enable_comments"`
	EnableNotifications  bool                   `json:"enable_notifications"`
	TimeZone             string                 `json:"timezone"`
	Language             string                 `json:"language"`
	Theme                string                 `json:"theme"`
	CustomFields         map[string]interface{} `json:"custom_fields"`
	IntegrationSettings  IntegrationSettings    `json:"integration_settings"`
	SecuritySettings     SecuritySettings       `json:"security_settings"`
}

// IntegrationSettings 集成设置
type IntegrationSettings struct {
	EnableWebhooks    bool              `json:"enable_webhooks"`
	WebhookURL        string            `json:"webhook_url"`
	EnableAPIAccess   bool              `json:"enable_api_access"`
	APIRateLimit      int               `json:"api_rate_limit"`
	AllowedDomains    []string          `json:"allowed_domains"`
	ExternalServices  map[string]string `json:"external_services"`
}

// SecuritySettings 安全设置
type SecuritySettings struct {
	RequireTwoFactor     bool     `json:"require_two_factor"`
	AllowedIPRanges      []string `json:"allowed_ip_ranges"`
	SessionTimeout       int      `json:"session_timeout"` // 分钟
	PasswordPolicy       string   `json:"password_policy"`
	EnableAuditLog       bool     `json:"enable_audit_log"`
	DataRetentionDays    int      `json:"data_retention_days"`
}

// NewDefaultSpaceSettings 创建默认空间设置
func NewDefaultSpaceSettings() *SpaceSettings {
	return &SpaceSettings{
		AllowPublicAccess:   false,
		DefaultRole:         RoleViewer,
		EnableComments:      true,
		EnableNotifications: true,
		TimeZone:            "Asia/Shanghai",
		Language:            "zh-CN",
		Theme:               "light",
		CustomFields:        make(map[string]interface{}),
		IntegrationSettings: IntegrationSettings{
			EnableWebhooks:   false,
			EnableAPIAccess:  true,
			APIRateLimit:     1000,
			AllowedDomains:   []string{},
			ExternalServices: make(map[string]string),
		},
		SecuritySettings: SecuritySettings{
			RequireTwoFactor:  false,
			AllowedIPRanges:   []string{},
			SessionTimeout:    480, // 8小时
			PasswordPolicy:    "standard",
			EnableAuditLog:    true,
			DataRetentionDays: 365,
		},
	}
}

// Validate 验证空间设置
func (ss *SpaceSettings) Validate() error {
	// 验证默认角色
	if !ss.DefaultRole.IsValid() {
		return ErrInvalidRole
	}

	// 验证时区
	if _, err := time.LoadLocation(ss.TimeZone); err != nil {
		return DomainError{Code: "INVALID_TIMEZONE", Message: "invalid timezone"}
	}

	// 验证语言代码
	validLanguages := map[string]bool{
		"zh-CN": true, "zh-TW": true, "en-US": true, "en-GB": true,
		"ja-JP": true, "ko-KR": true, "fr-FR": true, "de-DE": true,
	}
	if !validLanguages[ss.Language] {
		return DomainError{Code: "INVALID_LANGUAGE", Message: "invalid language code"}
	}

	// 验证主题
	validThemes := map[string]bool{
		"light": true, "dark": true, "auto": true,
	}
	if !validThemes[ss.Theme] {
		return DomainError{Code: "INVALID_THEME", Message: "invalid theme"}
	}

	// 验证集成设置
	if err := ss.IntegrationSettings.Validate(); err != nil {
		return err
	}

	// 验证安全设置
	if err := ss.SecuritySettings.Validate(); err != nil {
		return err
	}

	return nil
}

// Validate 验证集成设置
func (is *IntegrationSettings) Validate() error {
	// 验证API速率限制
	if is.APIRateLimit < 0 || is.APIRateLimit > 10000 {
		return DomainError{Code: "INVALID_API_RATE_LIMIT", Message: "API rate limit must be between 0 and 10000"}
	}

	// 验证Webhook URL
	if is.EnableWebhooks && is.WebhookURL == "" {
		return DomainError{Code: "WEBHOOK_URL_REQUIRED", Message: "webhook URL is required when webhooks are enabled"}
	}

	return nil
}

// Validate 验证安全设置
func (ss *SecuritySettings) Validate() error {
	// 验证会话超时
	if ss.SessionTimeout < 30 || ss.SessionTimeout > 1440 { // 30分钟到24小时
		return DomainError{Code: "INVALID_SESSION_TIMEOUT", Message: "session timeout must be between 30 and 1440 minutes"}
	}

	// 验证数据保留天数
	if ss.DataRetentionDays < 30 || ss.DataRetentionDays > 2555 { // 30天到7年
		return DomainError{Code: "INVALID_DATA_RETENTION", Message: "data retention days must be between 30 and 2555"}
	}

	// 验证密码策略
	validPolicies := map[string]bool{
		"weak": true, "standard": true, "strong": true, "custom": true,
	}
	if !validPolicies[ss.PasswordPolicy] {
		return DomainError{Code: "INVALID_PASSWORD_POLICY", Message: "invalid password policy"}
	}

	return nil
}

// ToJSON 转换为JSON字符串
func (ss *SpaceSettings) ToJSON() (string, error) {
	data, err := json.Marshal(ss)
	if err != nil {
		return "", DomainError{Code: "SETTINGS_MARSHAL_FAILED", Message: "failed to marshal settings to JSON"}
	}
	return string(data), nil
}

// SpaceSettingsFromJSON 从JSON字符串创建空间设置
func SpaceSettingsFromJSON(jsonStr string) (*SpaceSettings, error) {
	if jsonStr == "" {
		return NewDefaultSpaceSettings(), nil
	}

	var settings SpaceSettings
	if err := json.Unmarshal([]byte(jsonStr), &settings); err != nil {
		return nil, DomainError{Code: "SETTINGS_UNMARSHAL_FAILED", Message: "failed to unmarshal settings from JSON"}
	}

	// 验证设置
	if err := settings.Validate(); err != nil {
		return nil, err
	}

	return &settings, nil
}

// Merge 合并空间设置
func (ss *SpaceSettings) Merge(other *SpaceSettings) *SpaceSettings {
	if other == nil {
		return ss
	}

	merged := *ss // 复制当前设置

	// 合并基础设置
	merged.AllowPublicAccess = other.AllowPublicAccess
	merged.DefaultRole = other.DefaultRole
	merged.EnableComments = other.EnableComments
	merged.EnableNotifications = other.EnableNotifications

	if other.TimeZone != "" {
		merged.TimeZone = other.TimeZone
	}
	if other.Language != "" {
		merged.Language = other.Language
	}
	if other.Theme != "" {
		merged.Theme = other.Theme
	}

	// 合并自定义字段
	if other.CustomFields != nil {
		if merged.CustomFields == nil {
			merged.CustomFields = make(map[string]interface{})
		}
		for k, v := range other.CustomFields {
			merged.CustomFields[k] = v
		}
	}

	// 合并集成设置
	merged.IntegrationSettings = ss.IntegrationSettings.Merge(&other.IntegrationSettings)

	// 合并安全设置
	merged.SecuritySettings = ss.SecuritySettings.Merge(&other.SecuritySettings)

	return &merged
}

// Merge 合并集成设置
func (is *IntegrationSettings) Merge(other *IntegrationSettings) IntegrationSettings {
	if other == nil {
		return *is
	}

	merged := *is
	merged.EnableWebhooks = other.EnableWebhooks
	merged.EnableAPIAccess = other.EnableAPIAccess

	if other.WebhookURL != "" {
		merged.WebhookURL = other.WebhookURL
	}
	if other.APIRateLimit > 0 {
		merged.APIRateLimit = other.APIRateLimit
	}
	if other.AllowedDomains != nil {
		merged.AllowedDomains = other.AllowedDomains
	}
	if other.ExternalServices != nil {
		if merged.ExternalServices == nil {
			merged.ExternalServices = make(map[string]string)
		}
		for k, v := range other.ExternalServices {
			merged.ExternalServices[k] = v
		}
	}

	return merged
}

// Merge 合并安全设置
func (ss *SecuritySettings) Merge(other *SecuritySettings) SecuritySettings {
	if other == nil {
		return *ss
	}

	merged := *ss
	merged.RequireTwoFactor = other.RequireTwoFactor
	merged.EnableAuditLog = other.EnableAuditLog

	if other.SessionTimeout > 0 {
		merged.SessionTimeout = other.SessionTimeout
	}
	if other.DataRetentionDays > 0 {
		merged.DataRetentionDays = other.DataRetentionDays
	}
	if other.PasswordPolicy != "" {
		merged.PasswordPolicy = other.PasswordPolicy
	}
	if other.AllowedIPRanges != nil {
		merged.AllowedIPRanges = other.AllowedIPRanges
	}

	return merged
}

// SpaceMetrics 空间指标值对象
type SpaceMetrics struct {
	TotalBases         int64     `json:"total_bases"`
	TotalTables        int64     `json:"total_tables"`
	TotalRecords       int64     `json:"total_records"`
	TotalCollaborators int64     `json:"total_collaborators"`
	StorageUsed        int64     `json:"storage_used"`        // 字节
	APICallsThisMonth  int64     `json:"api_calls_this_month"`
	LastActivityAt     time.Time `json:"last_activity_at"`
	CreatedAt          time.Time `json:"created_at"`
}

// NewSpaceMetrics 创建空间指标
func NewSpaceMetrics() *SpaceMetrics {
	return &SpaceMetrics{
		TotalBases:         0,
		TotalTables:        0,
		TotalRecords:       0,
		TotalCollaborators: 0,
		StorageUsed:        0,
		APICallsThisMonth:  0,
		LastActivityAt:     time.Now(),
		CreatedAt:          time.Now(),
	}
}

// UpdateActivity 更新活动时间
func (sm *SpaceMetrics) UpdateActivity() {
	sm.LastActivityAt = time.Now()
}

// AddBase 增加基础表数量
func (sm *SpaceMetrics) AddBase() {
	sm.TotalBases++
	sm.UpdateActivity()
}

// RemoveBase 减少基础表数量
func (sm *SpaceMetrics) RemoveBase() {
	if sm.TotalBases > 0 {
		sm.TotalBases--
	}
	sm.UpdateActivity()
}

// AddTable 增加表格数量
func (sm *SpaceMetrics) AddTable() {
	sm.TotalTables++
	sm.UpdateActivity()
}

// RemoveTable 减少表格数量
func (sm *SpaceMetrics) RemoveTable() {
	if sm.TotalTables > 0 {
		sm.TotalTables--
	}
	sm.UpdateActivity()
}

// AddRecord 增加记录数量
func (sm *SpaceMetrics) AddRecord() {
	sm.TotalRecords++
	sm.UpdateActivity()
}

// RemoveRecord 减少记录数量
func (sm *SpaceMetrics) RemoveRecord() {
	if sm.TotalRecords > 0 {
		sm.TotalRecords--
	}
	sm.UpdateActivity()
}

// AddCollaborator 增加协作者数量
func (sm *SpaceMetrics) AddCollaborator() {
	sm.TotalCollaborators++
	sm.UpdateActivity()
}

// RemoveCollaborator 减少协作者数量
func (sm *SpaceMetrics) RemoveCollaborator() {
	if sm.TotalCollaborators > 0 {
		sm.TotalCollaborators--
	}
	sm.UpdateActivity()
}

// AddStorageUsage 增加存储使用量
func (sm *SpaceMetrics) AddStorageUsage(bytes int64) {
	sm.StorageUsed += bytes
	sm.UpdateActivity()
}

// RemoveStorageUsage 减少存储使用量
func (sm *SpaceMetrics) RemoveStorageUsage(bytes int64) {
	sm.StorageUsed -= bytes
	if sm.StorageUsed < 0 {
		sm.StorageUsed = 0
	}
	sm.UpdateActivity()
}

// IncrementAPICall 增加API调用次数
func (sm *SpaceMetrics) IncrementAPICall() {
	sm.APICallsThisMonth++
	sm.UpdateActivity()
}