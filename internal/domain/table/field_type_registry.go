package table

import (
	"context"
	"fmt"
	"sync"
)

// FieldTypeHandler 字段类型处理器接口 - 支持插件化扩展
type FieldTypeHandler interface {
	// GetType 获取字段类型
	GetType() FieldType
	
	// GetInfo 获取字段类型信息
	GetInfo() FieldTypeInfo
	
	// ValidateValue 验证字段值
	ValidateValue(value interface{}, options *FieldOptions) error
	
	// ValidateOptions 验证字段选项配置
	ValidateOptions(options *FieldOptions) error
	
	// GetDefaultOptions 获取默认选项配置
	GetDefaultOptions() *FieldOptions
	
	// IsCompatibleWith 检查与其他类型的兼容性
	IsCompatibleWith(targetType FieldType) bool
	
	// ConvertValue 转换值到目标类型
	ConvertValue(value interface{}, targetType FieldType, targetOptions *FieldOptions) (interface{}, error)
	
	// SupportsUnique 是否支持唯一性约束
	SupportsUnique() bool
	
	// RequiresOptions 是否需要选项配置
	RequiresOptions() bool
	
	// GetValidationRules 获取验证规则
	GetValidationRules(options *FieldOptions) []FieldValidationRule
}

// FieldTypeRegistry 字段类型注册表 - 支持插件化扩展
type FieldTypeRegistry struct {
	handlers map[FieldType]FieldTypeHandler
	mutex    sync.RWMutex
}

// NewFieldTypeRegistry 创建字段类型注册表
func NewFieldTypeRegistry() *FieldTypeRegistry {
	registry := &FieldTypeRegistry{
		handlers: make(map[FieldType]FieldTypeHandler),
	}
	
	// 注册内置字段类型处理器
	registry.registerBuiltinHandlers()
	
	return registry
}

// RegisterHandler 注册字段类型处理器
func (r *FieldTypeRegistry) RegisterHandler(handler FieldTypeHandler) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	fieldType := handler.GetType()
	if _, exists := r.handlers[fieldType]; exists {
		return fmt.Errorf("字段类型 %s 已注册", fieldType)
	}
	
	r.handlers[fieldType] = handler
	return nil
}

// GetHandler 获取字段类型处理器
func (r *FieldTypeRegistry) GetHandler(fieldType FieldType) (FieldTypeHandler, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	handler, exists := r.handlers[fieldType]
	if !exists {
		return nil, fmt.Errorf("未找到字段类型 %s 的处理器", fieldType)
	}
	
	return handler, nil
}

// GetAllHandlers 获取所有字段类型处理器
func (r *FieldTypeRegistry) GetAllHandlers() map[FieldType]FieldTypeHandler {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	result := make(map[FieldType]FieldTypeHandler)
	for k, v := range r.handlers {
		result[k] = v
	}
	
	return result
}

// GetAllFieldTypes 获取所有字段类型信息
func (r *FieldTypeRegistry) GetAllFieldTypes() []FieldTypeInfo {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	var infos []FieldTypeInfo
	for _, handler := range r.handlers {
		infos = append(infos, handler.GetInfo())
	}
	
	return infos
}

// ValidateFieldValue 验证字段值
func (r *FieldTypeRegistry) ValidateFieldValue(fieldType FieldType, value interface{}, options *FieldOptions) error {
	handler, err := r.GetHandler(fieldType)
	if err != nil {
		return err
	}
	
	return handler.ValidateValue(value, options)
}

// ValidateFieldOptions 验证字段选项配置
func (r *FieldTypeRegistry) ValidateFieldOptions(fieldType FieldType, options *FieldOptions) error {
	handler, err := r.GetHandler(fieldType)
	if err != nil {
		return err
	}
	
	return handler.ValidateOptions(options)
}

// IsFieldTypeCompatible 检查字段类型兼容性
func (r *FieldTypeRegistry) IsFieldTypeCompatible(sourceType, targetType FieldType) bool {
	if sourceType == targetType {
		return true
	}
	
	handler, err := r.GetHandler(sourceType)
	if err != nil {
		return false
	}
	
	return handler.IsCompatibleWith(targetType)
}

// ConvertFieldValue 转换字段值
func (r *FieldTypeRegistry) ConvertFieldValue(value interface{}, sourceType, targetType FieldType, targetOptions *FieldOptions) (interface{}, error) {
	if sourceType == targetType {
		return value, nil
	}
	
	handler, err := r.GetHandler(sourceType)
	if err != nil {
		return nil, err
	}
	
	return handler.ConvertValue(value, targetType, targetOptions)
}

// GetFieldTypeInfo 获取字段类型信息
func (r *FieldTypeRegistry) GetFieldTypeInfo(fieldType FieldType) (FieldTypeInfo, error) {
	handler, err := r.GetHandler(fieldType)
	if err != nil {
		return FieldTypeInfo{}, err
	}
	
	return handler.GetInfo(), nil
}

// GetDefaultFieldOptions 获取字段类型的默认选项
func (r *FieldTypeRegistry) GetDefaultFieldOptions(fieldType FieldType) (*FieldOptions, error) {
	handler, err := r.GetHandler(fieldType)
	if err != nil {
		return nil, err
	}
	
	return handler.GetDefaultOptions(), nil
}

// registerBuiltinHandlers 注册内置字段类型处理器
func (r *FieldTypeRegistry) registerBuiltinHandlers() {
	// 基础类型
	r.handlers[FieldTypeText] = NewTextFieldHandler()
	r.handlers[FieldTypeNumber] = NewNumberFieldHandler()
	r.handlers[FieldTypeBoolean] = NewBooleanFieldHandler()
	r.handlers[FieldTypeDate] = NewDateFieldHandler()
	r.handlers[FieldTypeDateTime] = NewDateTimeFieldHandler()
	r.handlers[FieldTypeTime] = NewTimeFieldHandler()
	
	// 选择类型
	r.handlers[FieldTypeSelect] = NewSelectFieldHandler()
	r.handlers[FieldTypeMultiSelect] = NewMultiSelectFieldHandler()
	r.handlers[FieldTypeRadio] = NewRadioFieldHandler()
	r.handlers[FieldTypeCheckbox] = NewCheckboxFieldHandler()
	
	// 高级类型
	r.handlers[FieldTypeEmail] = NewEmailFieldHandler()
	r.handlers[FieldTypeURL] = NewURLFieldHandler()
	r.handlers[FieldTypePhone] = NewPhoneFieldHandler()
	r.handlers[FieldTypeCurrency] = NewCurrencyFieldHandler()
	r.handlers[FieldTypePercent] = NewPercentFieldHandler()
	r.handlers[FieldTypeRating] = NewRatingFieldHandler()
	r.handlers[FieldTypeProgress] = NewProgressFieldHandler()
	
	// 媒体类型
	r.handlers[FieldTypeImage] = NewImageFieldHandler()
	r.handlers[FieldTypeFile] = NewFileFieldHandler()
	r.handlers[FieldTypeVideo] = NewVideoFieldHandler()
	r.handlers[FieldTypeAudio] = NewAudioFieldHandler()
	
	// 关系类型
	// r.handlers[FieldTypeLink] = NewLinkFieldHandler() // 暂时禁用，需要修复类型问题
	r.handlers[FieldTypeLookup] = NewLookupFieldHandler()
	r.handlers[FieldTypeRollup] = NewRollupFieldHandler()
	r.handlers[FieldTypeFormula] = NewFormulaFieldHandler()
	
	// 特殊类型
	r.handlers[FieldTypeAutoNumber] = NewAutoNumberFieldHandler()
	r.handlers[FieldTypeCreatedTime] = NewCreatedTimeFieldHandler()
	r.handlers[FieldTypeLastModifiedTime] = NewLastModifiedTimeFieldHandler()
	r.handlers[FieldTypeCreatedBy] = NewCreatedByFieldHandler()
	r.handlers[FieldTypeLastModifiedBy] = NewLastModifiedByFieldHandler()
}

// FieldTypeService 字段类型服务 - 提供高级功能
type FieldTypeService struct {
	registry *FieldTypeRegistry
}

// NewFieldTypeService 创建字段类型服务
func NewFieldTypeService(registry *FieldTypeRegistry) *FieldTypeService {
	return &FieldTypeService{
		registry: registry,
	}
}

// ValidateFieldConfiguration 验证字段配置的完整性
func (s *FieldTypeService) ValidateFieldConfiguration(ctx context.Context, field *Field) error {
	// 验证字段类型是否存在
	handler, err := s.registry.GetHandler(field.Type)
	if err != nil {
		return fmt.Errorf("无效的字段类型: %v", err)
	}
	
	// 验证选项配置
	if err := handler.ValidateOptions(field.Options); err != nil {
		return fmt.Errorf("字段选项配置无效: %v", err)
	}
	
	// 验证必需的选项配置
	if handler.RequiresOptions() && field.Options == nil {
		return fmt.Errorf("字段类型 %s 需要配置选项", field.Type)
	}
	
	// 验证唯一性约束
	if field.IsUnique && !handler.SupportsUnique() {
		return fmt.Errorf("字段类型 %s 不支持唯一性约束", field.Type)
	}
	
	// 验证默认值
	if field.DefaultValue != nil {
		if err := handler.ValidateValue(*field.DefaultValue, field.Options); err != nil {
			return fmt.Errorf("默认值无效: %v", err)
		}
	}
	
	return nil
}

// GetFieldTypeCompatibilityMatrix 获取字段类型兼容性矩阵
func (s *FieldTypeService) GetFieldTypeCompatibilityMatrix(ctx context.Context) map[FieldType][]FieldType {
	matrix := make(map[FieldType][]FieldType)
	
	handlers := s.registry.GetAllHandlers()
	for sourceType, sourceHandler := range handlers {
		var compatibleTypes []FieldType
		for targetType := range handlers {
			if sourceHandler.IsCompatibleWith(targetType) {
				compatibleTypes = append(compatibleTypes, targetType)
			}
		}
		matrix[sourceType] = compatibleTypes
	}
	
	return matrix
}

// PreviewFieldTypeChange 预览字段类型变更的影响
func (s *FieldTypeService) PreviewFieldTypeChange(ctx context.Context, field *Field, newType FieldType, newOptions *FieldOptions) (*FieldTypeChangePreview, error) {
	preview := &FieldTypeChangePreview{
		FieldID:       field.ID,
		CurrentType:   field.Type,
		NewType:       newType,
		IsCompatible:  false,
		Warnings:      []string{},
		Errors:        []string{},
	}
	
	// 检查类型兼容性
	if s.registry.IsFieldTypeCompatible(field.Type, newType) {
		preview.IsCompatible = true
	} else {
		preview.Errors = append(preview.Errors, fmt.Sprintf("字段类型从 %s 到 %s 不兼容", field.Type, newType))
		return preview, nil
	}
	
	// 检查约束兼容性
	newHandler, err := s.registry.GetHandler(newType)
	if err != nil {
		preview.Errors = append(preview.Errors, fmt.Sprintf("无效的目标字段类型: %v", err))
		return preview, nil
	}
	
	if field.IsUnique && !newHandler.SupportsUnique() {
		preview.Warnings = append(preview.Warnings, "新字段类型不支持唯一性约束，该约束将被移除")
	}
	
	// 检查选项配置
	if newHandler.RequiresOptions() && (newOptions == nil || len(newOptions.Choices) == 0) {
		preview.Errors = append(preview.Errors, fmt.Sprintf("字段类型 %s 需要配置选项", newType))
	}
	
	// 检查默认值兼容性
	if field.DefaultValue != nil {
		if err := newHandler.ValidateValue(*field.DefaultValue, newOptions); err != nil {
			preview.Warnings = append(preview.Warnings, "当前默认值与新字段类型不兼容，需要重新设置")
		}
	}
	
	return preview, nil
}

// FieldTypeChangePreview 字段类型变更预览
type FieldTypeChangePreview struct {
	FieldID      string      `json:"field_id"`
	CurrentType  FieldType   `json:"current_type"`
	NewType      FieldType   `json:"new_type"`
	IsCompatible bool        `json:"is_compatible"`
	Warnings     []string    `json:"warnings"`
	Errors       []string    `json:"errors"`
}

// 全局字段类型注册表实例
var globalFieldTypeRegistry *FieldTypeRegistry
var registryOnce sync.Once

// GetGlobalFieldTypeRegistry 获取全局字段类型注册表
func GetGlobalFieldTypeRegistry() *FieldTypeRegistry {
	registryOnce.Do(func() {
		globalFieldTypeRegistry = NewFieldTypeRegistry()
	})
	return globalFieldTypeRegistry
}

// RegisterFieldTypeHandler 注册字段类型处理器到全局注册表
func RegisterFieldTypeHandler(handler FieldTypeHandler) error {
	return GetGlobalFieldTypeRegistry().RegisterHandler(handler)
}