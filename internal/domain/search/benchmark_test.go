package search

import (
	"testing"
)

// BenchmarkNewSearchIndex 测试创建搜索索引的性能
func BenchmarkNewSearchIndex(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewSearchIndex(SearchTypeRecord, "Test Title", "Test Content")
	}
}

// BenchmarkSearchIndex_AddKeyword 测试添加关键词的性能
func BenchmarkSearchIndex_AddKeyword(b *testing.B) {
	index := NewSearchIndex(SearchTypeRecord, "Test Title", "Test Content")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		index.AddKeyword("keyword")
	}
}

// BenchmarkSearchIndex_AddKeywords 测试批量添加关键词的性能
func BenchmarkSearchIndex_AddKeywords(b *testing.B) {
	index := NewSearchIndex(SearchTypeRecord, "Test Title", "Test Content")
	keywords := []string{"keyword1", "keyword2", "keyword3", "keyword4", "keyword5"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		index.AddKeywords(keywords)
	}
}

// BenchmarkSearchIndex_SetMetadata 测试设置元数据的性能
func BenchmarkSearchIndex_SetMetadata(b *testing.B) {
	index := NewSearchIndex(SearchTypeRecord, "Test Title", "Test Content")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		index.SetMetadata("key", "value")
	}
}

// BenchmarkSearchIndex_AddPermission 测试添加权限的性能
func BenchmarkSearchIndex_AddPermission(b *testing.B) {
	index := NewSearchIndex(SearchTypeRecord, "Test Title", "Test Content")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		index.AddPermission("read")
	}
}

// BenchmarkSearchIndex_AddTag 测试添加标签的性能
func BenchmarkSearchIndex_AddTag(b *testing.B) {
	index := NewSearchIndex(SearchTypeRecord, "Test Title", "Test Content")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		index.AddTag("tag")
	}
}

// BenchmarkNewSearchSuggestion 测试创建搜索建议的性能
func BenchmarkNewSearchSuggestion(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewSearchSuggestion("test query", 10)
	}
}
