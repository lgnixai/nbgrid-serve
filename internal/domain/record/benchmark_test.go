package record

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// MockRecordRepository 用于基准测试的Mock仓储
type MockRecordRepository struct {
	records map[string]*Record
}

func NewMockRecordRepository() *MockRecordRepository {
	return &MockRecordRepository{
		records: make(map[string]*Record),
	}
}

func (m *MockRecordRepository) Create(ctx context.Context, record *Record) error {
	m.records[record.ID] = record
	return nil
}

func (m *MockRecordRepository) Update(ctx context.Context, record *Record) error {
	m.records[record.ID] = record
	return nil
}

func (m *MockRecordRepository) Delete(ctx context.Context, id string) error {
	delete(m.records, id)
	return nil
}

func (m *MockRecordRepository) GetByID(ctx context.Context, id string) (*Record, error) {
	record, exists := m.records[id]
	if !exists {
		return nil, fmt.Errorf("record not found")
	}
	return record, nil
}

func (m *MockRecordRepository) GetByTableID(ctx context.Context, tableID string, offset, limit int) ([]*Record, error) {
	var result []*Record
	count := 0
	for _, record := range m.records {
		if record.TableID == tableID {
			if count >= offset && len(result) < limit {
				result = append(result, record)
			}
			count++
		}
	}
	return result, nil
}

func (m *MockRecordRepository) CountByTableID(ctx context.Context, tableID string) (int64, error) {
	count := int64(0)
	for _, record := range m.records {
		if record.TableID == tableID {
			count++
		}
	}
	return count, nil
}

func (m *MockRecordRepository) BatchCreate(ctx context.Context, records []*Record) error {
	for _, record := range records {
		m.records[record.ID] = record
	}
	return nil
}

func (m *MockRecordRepository) BatchUpdate(ctx context.Context, records []*Record) error {
	for _, record := range records {
		m.records[record.ID] = record
	}
	return nil
}

func (m *MockRecordRepository) BatchDelete(ctx context.Context, ids []string) error {
	for _, id := range ids {
		delete(m.records, id)
	}
	return nil
}

// BenchmarkRecordService_Create 测试单条记录创建性能
func BenchmarkRecordService_Create(b *testing.B) {
	repo := NewMockRecordRepository()
	service := NewService(repo)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		record := &Record{
			TableID: "table1",
			Data: map[string]interface{}{
				"name":        fmt.Sprintf("Record %d", i),
				"description": "Benchmark test record",
				"value":       i,
				"created_at":  time.Now(),
			},
		}
		_ = service.CreateRecord(ctx, record)
	}
}

// BenchmarkRecordService_BatchCreate 测试批量创建性能
func BenchmarkRecordService_BatchCreate(b *testing.B) {
	batchSizes := []int{10, 100, 1000}

	for _, batchSize := range batchSizes {
		b.Run(fmt.Sprintf("BatchSize-%d", batchSize), func(b *testing.B) {
			repo := NewMockRecordRepository()
			service := NewService(repo)
			ctx := context.Background()

			// 准备批量数据
			records := make([]*Record, batchSize)
			for i := 0; i < batchSize; i++ {
				records[i] = &Record{
					TableID: "table1",
					Data: map[string]interface{}{
						"name":        fmt.Sprintf("Record %d", i),
						"description": "Benchmark test record",
						"value":       i,
						"created_at":  time.Now(),
					},
				}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = service.BatchCreateRecords(ctx, records)
			}
		})
	}
}

// BenchmarkRecordService_Update 测试更新性能
func BenchmarkRecordService_Update(b *testing.B) {
	repo := NewMockRecordRepository()
	service := NewService(repo)
	ctx := context.Background()

	// 预创建记录
	record := &Record{
		ID:      "record1",
		TableID: "table1",
		Data: map[string]interface{}{
			"name": "Original",
		},
	}
	_ = repo.Create(ctx, record)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		record.Data["name"] = fmt.Sprintf("Updated %d", i)
		record.Data["updated_at"] = time.Now()
		_ = service.UpdateRecord(ctx, record)
	}
}

// BenchmarkRecordService_GetByID 测试查询性能
func BenchmarkRecordService_GetByID(b *testing.B) {
	repo := NewMockRecordRepository()
	service := NewService(repo)
	ctx := context.Background()

	// 预创建大量记录
	numRecords := 10000
	for i := 0; i < numRecords; i++ {
		record := &Record{
			ID:      fmt.Sprintf("record%d", i),
			TableID: "table1",
			Data: map[string]interface{}{
				"name": fmt.Sprintf("Record %d", i),
			},
		}
		_ = repo.Create(ctx, record)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		recordID := fmt.Sprintf("record%d", i%numRecords)
		_, _ = service.GetRecord(ctx, recordID)
	}
}

// BenchmarkRecordService_List 测试列表查询性能
func BenchmarkRecordService_List(b *testing.B) {
	repo := NewMockRecordRepository()
	service := NewService(repo)
	ctx := context.Background()

	// 预创建记录
	tableID := "table1"
	for i := 0; i < 1000; i++ {
		record := &Record{
			ID:      fmt.Sprintf("record%d", i),
			TableID: tableID,
			Data: map[string]interface{}{
				"name": fmt.Sprintf("Record %d", i),
			},
		}
		_ = repo.Create(ctx, record)
	}

	pageSizes := []int{10, 50, 100}
	for _, pageSize := range pageSizes {
		b.Run(fmt.Sprintf("PageSize-%d", pageSize), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = service.ListRecords(ctx, tableID, i*pageSize, pageSize)
			}
		})
	}
}

// BenchmarkRecordService_ComplexQuery 测试复杂查询性能
func BenchmarkRecordService_ComplexQuery(b *testing.B) {
	repo := NewMockRecordRepository()
	service := NewService(repo)
	ctx := context.Background()

	// 预创建具有不同属性的记录
	for i := 0; i < 10000; i++ {
		record := &Record{
			ID:      fmt.Sprintf("record%d", i),
			TableID: "table1",
			Data: map[string]interface{}{
				"name":     fmt.Sprintf("Record %d", i),
				"category": fmt.Sprintf("cat%d", i%10),
				"status":   i%3 == 0,
				"value":    i * 100,
				"tags":     []string{"tag1", "tag2", "tag3"},
			},
		}
		_ = repo.Create(ctx, record)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 模拟复杂查询：按类别过滤，按值排序，分页
		records, _ := service.ListRecords(ctx, "table1", 0, 50)
		
		// 模拟过滤和排序逻辑
		filtered := make([]*Record, 0)
		for _, r := range records {
			if category, ok := r.Data["category"].(string); ok && category == "cat5" {
				filtered = append(filtered, r)
			}
		}
	}
}

// BenchmarkRecordService_ConcurrentOperations 测试并发操作性能
func BenchmarkRecordService_ConcurrentOperations(b *testing.B) {
	repo := NewMockRecordRepository()
	service := NewService(repo)
	ctx := context.Background()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			// 混合操作：创建、读取、更新
			switch i % 3 {
			case 0: // 创建
				record := &Record{
					TableID: "table1",
					Data: map[string]interface{}{
						"name": fmt.Sprintf("Concurrent %d", i),
					},
				}
				_ = service.CreateRecord(ctx, record)
			case 1: // 读取
				_, _ = service.GetRecord(ctx, fmt.Sprintf("record%d", i%100))
			case 2: // 更新
				record := &Record{
					ID:      fmt.Sprintf("record%d", i%100),
					TableID: "table1",
					Data: map[string]interface{}{
						"name": fmt.Sprintf("Updated %d", i),
					},
				}
				_ = service.UpdateRecord(ctx, record)
			}
			i++
		}
	})
}

// BenchmarkRecordValidation 测试记录验证性能
func BenchmarkRecordValidation(b *testing.B) {
	// 模拟不同复杂度的数据
	simpleData := map[string]interface{}{
		"name": "Simple Record",
		"age":  25,
	}

	complexData := map[string]interface{}{
		"name":        "Complex Record",
		"age":         25,
		"email":       "test@example.com",
		"address":     "123 Main St",
		"tags":        []string{"tag1", "tag2", "tag3", "tag4", "tag5"},
		"metadata":    map[string]interface{}{"key1": "value1", "key2": "value2"},
		"created_at":  time.Now(),
		"updated_at":  time.Now(),
		"is_active":   true,
		"score":       98.5,
	}

	b.Run("Simple", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = validateRecordData(simpleData)
		}
	})

	b.Run("Complex", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = validateRecordData(complexData)
		}
	})
}

// validateRecordData 模拟记录数据验证
func validateRecordData(data map[string]interface{}) error {
	// 模拟验证逻辑
	for key, value := range data {
		switch v := value.(type) {
		case string:
			if len(v) > 1000 {
				return fmt.Errorf("string too long: %s", key)
			}
		case []string:
			if len(v) > 100 {
				return fmt.Errorf("array too long: %s", key)
			}
		}
	}
	return nil
}

// BenchmarkMemoryAllocation 测试内存分配
func BenchmarkMemoryAllocation(b *testing.B) {
	b.Run("NewRecord", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = &Record{
				ID:      fmt.Sprintf("record%d", i),
				TableID: "table1",
				Data:    make(map[string]interface{}, 10),
			}
		}
	})

	b.Run("RecordWithData", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			record := &Record{
				ID:      fmt.Sprintf("record%d", i),
				TableID: "table1",
				Data:    make(map[string]interface{}, 10),
			}
			record.Data["field1"] = "value1"
			record.Data["field2"] = i
			record.Data["field3"] = true
			record.Data["field4"] = []string{"a", "b", "c"}
		}
	})
}