package table

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// VirtualFieldService manages virtual field calculations
type VirtualFieldService struct {
	tableService   TableService
	recordService  RecordService
	fieldHandlers  map[FieldType]FieldTypeHandler
	aiProvider     AIProvider
	cache          VirtualFieldCache
	mu             sync.RWMutex
}

// VirtualFieldCache caches computed virtual field values
type VirtualFieldCache interface {
	Get(recordID, fieldID string) (interface{}, bool)
	Set(recordID, fieldID string, value interface{}, ttl time.Duration)
	Delete(recordID, fieldID string)
	DeleteByRecord(recordID string)
	DeleteByField(fieldID string)
}

// RecordService interface for accessing record data
type RecordService interface {
	GetRecord(ctx context.Context, tableID, recordID string) (map[string]interface{}, error)
	GetLinkedRecords(ctx context.Context, tableID, recordID, linkFieldID string) ([]map[string]interface{}, error)
}

// NewVirtualFieldService creates a new virtual field service
func NewVirtualFieldService(
	tableService TableService,
	recordService RecordService,
	aiProvider AIProvider,
	cache VirtualFieldCache,
) *VirtualFieldService {
	service := &VirtualFieldService{
		tableService:  tableService,
		recordService: recordService,
		fieldHandlers: make(map[FieldType]FieldTypeHandler),
		aiProvider:    aiProvider,
		cache:         cache,
	}

	// Register virtual field handlers
	service.RegisterHandler(FieldTypeVirtualFormula, NewFormulaFieldHandler())
	service.RegisterHandler(FieldTypeVirtualAI, NewAIFieldHandler(aiProvider))
	// TODO: Register lookup and rollup handlers

	return service
}

// RegisterHandler registers a field type handler
func (s *VirtualFieldService) RegisterHandler(fieldType FieldType, handler FieldTypeHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fieldHandlers[fieldType] = handler
}

// CalculateVirtualFields calculates all virtual fields for a record
func (s *VirtualFieldService) CalculateVirtualFields(
	ctx context.Context,
	table *Table,
	recordData map[string]interface{},
	fields []string, // specific fields to calculate, empty for all
) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// Get fields to calculate
	fieldsToCalc := s.getFieldsToCalculate(table, fields)
	
	// Calculate each virtual field
	for _, field := range fieldsToCalc {
		if !IsVirtualField(field.Type) {
			continue
		}
		
		value, err := s.CalculateField(ctx, table, field, recordData)
		if err != nil {
			// Store error in result but continue with other fields
			result[field.Code] = map[string]interface{}{
				"error": err.Error(),
				"value": nil,
			}
		} else {
			result[field.Code] = value
		}
	}
	
	return result, nil
}

// CalculateField calculates a single virtual field value
func (s *VirtualFieldService) CalculateField(
	ctx context.Context,
	table *Table,
	field *Field,
	recordData map[string]interface{},
) (interface{}, error) {
	if !IsVirtualField(field.Type) {
		return nil, fmt.Errorf("field %s is not a virtual field", field.Code)
	}
	
	// Check cache first
	recordID, _ := recordData["id"].(string)
	if recordID != "" && s.cache != nil {
		if cachedValue, found := s.cache.Get(recordID, field.ID); found {
			return cachedValue, nil
		}
	}
	
	// Get handler for field type
	s.mu.RLock()
	handler, exists := s.fieldHandlers[field.Type]
	s.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("no handler registered for field type %s", field.Type)
	}
	
	// Create calculation context
	calcCtx := CalculationContext{
		RecordData: recordData,
		Table:      table,
		Field:      field,
		UserID:     getUserIDFromContext(ctx),
		Context: map[string]interface{}{
			"ctx": ctx,
		},
	}
	
	// Check if handler implements VirtualFieldCalculator
	calculator, ok := handler.(VirtualFieldCalculator)
	if !ok {
		return nil, fmt.Errorf("handler for %s does not implement VirtualFieldCalculator", field.Type)
	}
	
	// Calculate value
	value, err := calculator.Calculate(calcCtx)
	if err != nil {
		return nil, err
	}
	
	// Cache the result
	if recordID != "" && s.cache != nil && value != nil {
		// Cache for 5 minutes by default
		s.cache.Set(recordID, field.ID, value, 5*time.Minute)
	}
	
	return value, nil
}

// InvalidateCache invalidates cached values
func (s *VirtualFieldService) InvalidateCache(recordID, fieldID string) {
	if s.cache == nil {
		return
	}
	
	if recordID != "" && fieldID != "" {
		s.cache.Delete(recordID, fieldID)
	} else if recordID != "" {
		s.cache.DeleteByRecord(recordID)
	} else if fieldID != "" {
		s.cache.DeleteByField(fieldID)
	}
}

// GetFieldDependencies returns the fields that a virtual field depends on
func (s *VirtualFieldService) GetFieldDependencies(field *Field) ([]string, error) {
	if !IsVirtualField(field.Type) {
		return nil, nil
	}
	
	s.mu.RLock()
	handler, exists := s.fieldHandlers[field.Type]
	s.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("no handler registered for field type %s", field.Type)
	}
	
	calculator, ok := handler.(VirtualFieldCalculator)
	if !ok {
		return nil, nil
	}
	
	return calculator.GetDependencies(), nil
}

// UpdateDependentFields updates virtual fields that depend on changed fields
func (s *VirtualFieldService) UpdateDependentFields(
	ctx context.Context,
	table *Table,
	recordID string,
	changedFields []string,
) error {
	// Find virtual fields that depend on the changed fields
	dependentFields := s.findDependentFields(table, changedFields)
	
	if len(dependentFields) == 0 {
		return nil
	}
	
	// Get the record data
	recordData, err := s.recordService.GetRecord(ctx, table.ID, recordID)
	if err != nil {
		return fmt.Errorf("failed to get record: %w", err)
	}
	
	// Invalidate cache for dependent fields
	for _, field := range dependentFields {
		s.InvalidateCache(recordID, field.ID)
	}
	
	// Recalculate dependent fields
	fieldCodes := make([]string, len(dependentFields))
	for i, field := range dependentFields {
		fieldCodes[i] = field.Code
	}
	
	_, err = s.CalculateVirtualFields(ctx, table, recordData, fieldCodes)
	return err
}

// Helper methods

func (s *VirtualFieldService) getFieldsToCalculate(table *Table, requestedFields []string) []*Field {
	if len(requestedFields) == 0 {
		// Return all virtual fields
		var fields []*Field
		for _, field := range table.Fields {
			if IsVirtualField(field.Type) {
				fields = append(fields, field)
			}
		}
		return fields
	}
	
	// Return only requested fields that are virtual
	fieldMap := make(map[string]*Field)
	for _, field := range table.Fields {
		fieldMap[field.Code] = field
	}
	
	var fields []*Field
	for _, code := range requestedFields {
		if field, exists := fieldMap[code]; exists && IsVirtualField(field.Type) {
			fields = append(fields, field)
		}
	}
	
	return fields
}

func (s *VirtualFieldService) findDependentFields(table *Table, changedFields []string) []*Field {
	changedSet := make(map[string]bool)
	for _, field := range changedFields {
		changedSet[field] = true
	}
	
	var dependentFields []*Field
	
	for _, field := range table.Fields {
		if !IsVirtualField(field.Type) {
			continue
		}
		
		// Get dependencies for this virtual field
		deps, err := s.GetFieldDependencies(field)
		if err != nil {
			continue
		}
		
		// Check if any dependency was changed
		for _, dep := range deps {
			if changedSet[dep] {
				dependentFields = append(dependentFields, field)
				break
			}
		}
	}
	
	return dependentFields
}

func getUserIDFromContext(ctx context.Context) string {
	// TODO: Extract user ID from context
	return ""
}

// InMemoryVirtualFieldCache is a simple in-memory cache implementation
type InMemoryVirtualFieldCache struct {
	data map[string]map[string]cacheEntry
	mu   sync.RWMutex
}

type cacheEntry struct {
	value     interface{}
	expiresAt time.Time
}

// NewInMemoryVirtualFieldCache creates a new in-memory cache
func NewInMemoryVirtualFieldCache() *InMemoryVirtualFieldCache {
	cache := &InMemoryVirtualFieldCache{
		data: make(map[string]map[string]cacheEntry),
	}
	
	// Start cleanup goroutine
	go cache.cleanup()
	
	return cache
}

func (c *InMemoryVirtualFieldCache) Get(recordID, fieldID string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if recordCache, exists := c.data[recordID]; exists {
		if entry, exists := recordCache[fieldID]; exists {
			if time.Now().Before(entry.expiresAt) {
				return entry.value, true
			}
		}
	}
	
	return nil, false
}

func (c *InMemoryVirtualFieldCache) Set(recordID, fieldID string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if _, exists := c.data[recordID]; !exists {
		c.data[recordID] = make(map[string]cacheEntry)
	}
	
	c.data[recordID][fieldID] = cacheEntry{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
}

func (c *InMemoryVirtualFieldCache) Delete(recordID, fieldID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if recordCache, exists := c.data[recordID]; exists {
		delete(recordCache, fieldID)
		if len(recordCache) == 0 {
			delete(c.data, recordID)
		}
	}
}

func (c *InMemoryVirtualFieldCache) DeleteByRecord(recordID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	delete(c.data, recordID)
}

func (c *InMemoryVirtualFieldCache) DeleteByField(fieldID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	for recordID, recordCache := range c.data {
		delete(recordCache, fieldID)
		if len(recordCache) == 0 {
			delete(c.data, recordID)
		}
	}
}

func (c *InMemoryVirtualFieldCache) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		
		for recordID, recordCache := range c.data {
			for fieldID, entry := range recordCache {
				if now.After(entry.expiresAt) {
					delete(recordCache, fieldID)
				}
			}
			if len(recordCache) == 0 {
				delete(c.data, recordID)
			}
		}
		
		c.mu.Unlock()
	}
}