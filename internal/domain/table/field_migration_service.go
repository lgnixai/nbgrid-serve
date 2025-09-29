package table

import (
	"context"
	// "fmt" // 暂时注释掉，如果需要可以取消注释
	"time"
)

// FieldMigrationService 字段迁移服务 - 处理安全的schema变更和数据转换
type FieldMigrationService interface {
	// AnalyzeFieldChange 分析字段变更的影响
	AnalyzeFieldChange(ctx context.Context, req FieldChangeAnalysisRequest) (*FieldChangeAnalysis, error)

	// ExecuteFieldMigration 执行字段迁移
	ExecuteFieldMigration(ctx context.Context, req FieldMigrationRequest) (*FieldMigrationResult, error)

	// RollbackFieldMigration 回滚字段迁移
	RollbackFieldMigration(ctx context.Context, migrationID string) (*FieldMigrationResult, error)

	// GetMigrationHistory 获取迁移历史
	GetMigrationHistory(ctx context.Context, tableID string) ([]*FieldMigration, error)

	// ValidateDataConversion 验证数据转换的可行性
	ValidateDataConversion(ctx context.Context, req DataConversionRequest) (*DataConversionValidation, error)
}

// FieldChangeAnalysisRequest 字段变更分析请求
type FieldChangeAnalysisRequest struct {
	TableID    string     `json:"table_id"`
	FieldID    string     `json:"field_id"`
	OldField   *Field     `json:"old_field"`
	NewField   *Field     `json:"new_field"`
	ChangeType ChangeType `json:"change_type"`
	UserID     string     `json:"user_id"`
}

// FieldChangeAnalysis 字段变更分析结果
type FieldChangeAnalysis struct {
	ChangeID              string              `json:"change_id"`
	TableID               string              `json:"table_id"`
	FieldID               string              `json:"field_id"`
	ChangeType            ChangeType          `json:"change_type"`
	IsSafe                bool                `json:"is_safe"`
	RequiresDataMigration bool                `json:"requires_data_migration"`
	AffectedRecords       int64               `json:"affected_records"`
	EstimatedDuration     time.Duration       `json:"estimated_duration"`
	Risks                 []FieldChangeRisk   `json:"risks"`
	Warnings              []string            `json:"warnings"`
	DataConversionPlan    *DataConversionPlan `json:"data_conversion_plan,omitempty"`
	Dependencies          []FieldDependency   `json:"dependencies"`
	CreatedAt             time.Time           `json:"created_at"`
}

// ChangeType 变更类型
type ChangeType string

const (
	ChangeTypeAdd        ChangeType = "add"
	ChangeTypeUpdate     ChangeType = "update"
	ChangeTypeDelete     ChangeType = "delete"
	ChangeTypeReorder    ChangeType = "reorder"
	ChangeTypeRename     ChangeType = "rename"
	ChangeTypeTypeChange ChangeType = "type_change"
)

// FieldChangeRisk 字段变更风险
type FieldChangeRisk struct {
	Level       RiskLevel `json:"level"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Impact      string    `json:"impact"`
	Mitigation  string    `json:"mitigation"`
}

// RiskLevel 风险级别
type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

// DataConversionPlan 数据转换计划
type DataConversionPlan struct {
	ConversionType    ConversionType         `json:"conversion_type"`
	BatchSize         int                    `json:"batch_size"`
	EstimatedBatches  int                    `json:"estimated_batches"`
	ConversionRules   []DataConversionRule   `json:"conversion_rules"`
	ValidationRules   []DataValidationRule   `json:"validation_rules"`
	FallbackStrategy  FallbackStrategy       `json:"fallback_strategy"`
	PreservationRules []DataPreservationRule `json:"preservation_rules"`
}

// ConversionType 转换类型
type ConversionType string

const (
	ConversionTypeAutomatic ConversionType = "automatic"
	ConversionTypeManual    ConversionType = "manual"
	ConversionTypeBatch     ConversionType = "batch"
	ConversionTypeStreaming ConversionType = "streaming"
)

// DataConversionRule 数据转换规则
type DataConversionRule struct {
	SourcePattern string      `json:"source_pattern"`
	TargetFormat  string      `json:"target_format"`
	Transformer   string      `json:"transformer"`
	Parameters    interface{} `json:"parameters"`
	Priority      int         `json:"priority"`
}

// DataValidationRule 数据验证规则
type DataValidationRule struct {
	RuleType    string      `json:"rule_type"`
	Expression  string      `json:"expression"`
	ErrorAction string      `json:"error_action"`
	Parameters  interface{} `json:"parameters"`
}

// FallbackStrategy 回退策略
type FallbackStrategy struct {
	Strategy     string      `json:"strategy"`
	DefaultValue interface{} `json:"default_value"`
	ErrorAction  string      `json:"error_action"`
}

// DataPreservationRule 数据保留规则
type DataPreservationRule struct {
	PreservationType string `json:"preservation_type"`
	Duration         string `json:"duration"`
	StorageLocation  string `json:"storage_location"`
}

// FieldDependency 字段依赖关系
type FieldDependency struct {
	DependentFieldID string `json:"dependent_field_id"`
	DependentTableID string `json:"dependent_table_id"`
	DependencyType   string `json:"dependency_type"`
	Impact           string `json:"impact"`
	RequiresUpdate   bool   `json:"requires_update"`
}

// FieldMigrationRequest 字段迁移请求
type FieldMigrationRequest struct {
	AnalysisID    string           `json:"analysis_id"`
	TableID       string           `json:"table_id"`
	FieldID       string           `json:"field_id"`
	Changes       []SchemaChange   `json:"changes"`
	MigrationMode MigrationMode    `json:"migration_mode"`
	BatchSize     int              `json:"batch_size"`
	MaxDuration   time.Duration    `json:"max_duration"`
	UserID        string           `json:"user_id"`
	Options       MigrationOptions `json:"options"`
}

// MigrationMode 迁移模式
type MigrationMode string

const (
	MigrationModeImmediate  MigrationMode = "immediate"
	MigrationModeScheduled  MigrationMode = "scheduled"
	MigrationModeBatch      MigrationMode = "batch"
	MigrationModeBackground MigrationMode = "background"
)

// MigrationOptions 迁移选项
type MigrationOptions struct {
	CreateBackup            bool          `json:"create_backup"`
	ValidateBeforeMigration bool          `json:"validate_before_migration"`
	ValidateAfterMigration  bool          `json:"validate_after_migration"`
	StopOnError             bool          `json:"stop_on_error"`
	NotifyOnCompletion      bool          `json:"notify_on_completion"`
	ScheduledTime           *time.Time    `json:"scheduled_time,omitempty"`
	MaxRetries              int           `json:"max_retries"`
	RetryDelay              time.Duration `json:"retry_delay"`
}

// FieldMigrationResult 字段迁移结果
type FieldMigrationResult struct {
	MigrationID       string           `json:"migration_id"`
	Status            MigrationStatus  `json:"status"`
	StartTime         time.Time        `json:"start_time"`
	EndTime           *time.Time       `json:"end_time,omitempty"`
	Duration          time.Duration    `json:"duration"`
	ProcessedRecords  int64            `json:"processed_records"`
	SuccessfulRecords int64            `json:"successful_records"`
	FailedRecords     int64            `json:"failed_records"`
	Errors            []MigrationError `json:"errors"`
	Warnings          []string         `json:"warnings"`
	BackupLocation    string           `json:"backup_location,omitempty"`
	RollbackData      *RollbackData    `json:"rollback_data,omitempty"`
}

// MigrationStatus 迁移状态
type MigrationStatus string

const (
	MigrationStatusPending    MigrationStatus = "pending"
	MigrationStatusRunning    MigrationStatus = "running"
	MigrationStatusCompleted  MigrationStatus = "completed"
	MigrationStatusFailed     MigrationStatus = "failed"
	MigrationStatusRolledBack MigrationStatus = "rolled_back"
	MigrationStatusCancelled  MigrationStatus = "cancelled"
)

// MigrationError 迁移错误
type MigrationError struct {
	RecordID  string      `json:"record_id"`
	FieldID   string      `json:"field_id"`
	ErrorType string      `json:"error_type"`
	Message   string      `json:"message"`
	OldValue  interface{} `json:"old_value"`
	NewValue  interface{} `json:"new_value"`
	Timestamp time.Time   `json:"timestamp"`
}

// RollbackData 回滚数据
type RollbackData struct {
	OriginalSchema  *Field    `json:"original_schema"`
	BackupLocation  string    `json:"backup_location"`
	AffectedRecords []string  `json:"affected_records"`
	RollbackScript  string    `json:"rollback_script"`
	CreatedAt       time.Time `json:"created_at"`
	ExpiresAt       time.Time `json:"expires_at"`
}

// FieldMigration 字段迁移记录
type FieldMigration struct {
	ID          string                `json:"id"`
	TableID     string                `json:"table_id"`
	FieldID     string                `json:"field_id"`
	ChangeType  ChangeType            `json:"change_type"`
	OldSchema   *Field                `json:"old_schema"`
	NewSchema   *Field                `json:"new_schema"`
	Status      MigrationStatus       `json:"status"`
	Result      *FieldMigrationResult `json:"result"`
	CreatedBy   string                `json:"created_by"`
	CreatedAt   time.Time             `json:"created_at"`
	CompletedAt *time.Time            `json:"completed_at,omitempty"`
}

// DataConversionRequest 数据转换请求
type DataConversionRequest struct {
	TableID       string        `json:"table_id"`
	FieldID       string        `json:"field_id"`
	SourceType    FieldType     `json:"source_type"`
	TargetType    FieldType     `json:"target_type"`
	SourceOptions *FieldOptions `json:"source_options"`
	TargetOptions *FieldOptions `json:"target_options"`
	SampleSize    int           `json:"sample_size"`
}

// DataConversionValidation 数据转换验证结果
type DataConversionValidation struct {
	IsValid            bool                `json:"is_valid"`
	ConversionRate     float64             `json:"conversion_rate"`
	SampleSize         int                 `json:"sample_size"`
	SuccessfulSamples  int                 `json:"successful_samples"`
	FailedSamples      int                 `json:"failed_samples"`
	ConversionExamples []ConversionExample `json:"conversion_examples"`
	ValidationErrors   []ValidationError   `json:"validation_errors"`
	Recommendations    []string            `json:"recommendations"`
}

// ConversionExample 转换示例
type ConversionExample struct {
	OriginalValue  interface{} `json:"original_value"`
	ConvertedValue interface{} `json:"converted_value"`
	IsSuccessful   bool        `json:"is_successful"`
	ErrorMessage   string      `json:"error_message,omitempty"`
}

// ValidationError 验证错误
type ValidationError struct {
	RecordID     string      `json:"record_id"`
	FieldValue   interface{} `json:"field_value"`
	ErrorType    string      `json:"error_type"`
	ErrorMessage string      `json:"error_message"`
}
