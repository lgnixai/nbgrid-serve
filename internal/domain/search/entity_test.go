package search

import (
	"testing"
	"time"
)

func TestNewSearchIndex(t *testing.T) {
	tests := []struct {
		name       string
		searchType SearchType
		title      string
		content    string
		wantErr    bool
	}{
		{
			name:       "valid search index",
			searchType: SearchTypeRecord,
			title:      "Test Record",
			content:    "This is test content for the search index",
			wantErr:    false,
		},
		{
			name:       "empty title",
			searchType: SearchTypeRecord,
			title:      "",
			content:    "This is test content",
			wantErr:    false, // 允许空标题
		},
		{
			name:       "empty content",
			searchType: SearchTypeRecord,
			title:      "Test Title",
			content:    "",
			wantErr:    false, // 允许空内容
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index := NewSearchIndex(tt.searchType, tt.title, tt.content)

			if index == nil {
				t.Error("NewSearchIndex() returned nil")
				return
			}

			if index.Type != tt.searchType {
				t.Errorf("NewSearchIndex() Type = %v, want %v", index.Type, tt.searchType)
			}

			if index.Title != tt.title {
				t.Errorf("NewSearchIndex() Title = %v, want %v", index.Title, tt.title)
			}

			if index.Content != tt.content {
				t.Errorf("NewSearchIndex() Content = %v, want %v", index.Content, tt.content)
			}

			if index.ID == "" {
				t.Error("NewSearchIndex() ID should not be empty")
			}

			if index.Keywords == nil {
				t.Error("NewSearchIndex() Keywords should not be nil")
			}

			if index.Metadata == nil {
				t.Error("NewSearchIndex() Metadata should not be nil")
			}

			if index.Permissions == nil {
				t.Error("NewSearchIndex() Permissions should not be nil")
			}

			if index.Tags == nil {
				t.Error("NewSearchIndex() Tags should not be nil")
			}

			if index.CreatedTime.IsZero() {
				t.Error("NewSearchIndex() CreatedTime should not be zero")
			}

			if index.UpdatedTime.IsZero() {
				t.Error("NewSearchIndex() UpdatedTime should not be zero")
			}
		})
	}
}

func TestSearchIndex_AddKeyword(t *testing.T) {
	index := NewSearchIndex(SearchTypeRecord, "Test", "Content")

	// 添加关键词
	index.AddKeyword("test")
	index.AddKeyword("record")

	if len(index.Keywords) != 2 {
		t.Errorf("AddKeyword() Keywords length = %v, want %v", len(index.Keywords), 2)
	}

	if index.Keywords[0] != "test" {
		t.Errorf("AddKeyword() Keywords[0] = %v, want %v", index.Keywords[0], "test")
	}

	if index.Keywords[1] != "record" {
		t.Errorf("AddKeyword() Keywords[1] = %v, want %v", index.Keywords[1], "record")
	}

	// 测试重复添加
	index.AddKeyword("test") // 应该不会重复添加

	if len(index.Keywords) != 2 {
		t.Errorf("AddKeyword() duplicate Keywords length = %v, want %v", len(index.Keywords), 2)
	}

	// 测试更新时间是否改变
	oldUpdatedTime := index.UpdatedTime
	time.Sleep(time.Millisecond)
	index.AddKeyword("new_keyword")

	if index.UpdatedTime == oldUpdatedTime {
		t.Error("AddKeyword() should update UpdatedTime")
	}
}

func TestSearchIndex_AddKeywords(t *testing.T) {
	index := NewSearchIndex(SearchTypeRecord, "Test", "Content")

	// 批量添加关键词
	keywords := []string{"keyword1", "keyword2", "keyword3"}
	index.AddKeywords(keywords)

	if len(index.Keywords) != 3 {
		t.Errorf("AddKeywords() Keywords length = %v, want %v", len(index.Keywords), 3)
	}

	for i, keyword := range keywords {
		if index.Keywords[i] != keyword {
			t.Errorf("AddKeywords() Keywords[%d] = %v, want %v", i, index.Keywords[i], keyword)
		}
	}

	// 测试更新时间是否改变
	oldUpdatedTime := index.UpdatedTime
	time.Sleep(time.Millisecond)
	index.AddKeywords([]string{"new_keyword"})

	if index.UpdatedTime == oldUpdatedTime {
		t.Error("AddKeywords() should update UpdatedTime")
	}
}

func TestSearchIndex_SetSource(t *testing.T) {
	index := NewSearchIndex(SearchTypeRecord, "Test", "Content")

	// 设置来源
	sourceID := "source123"
	sourceType := "table"
	sourceURL := "https://example.com/table/123"
	index.SetSource(sourceID, sourceType, sourceURL)

	if index.SourceID != sourceID {
		t.Errorf("SetSource() SourceID = %v, want %v", index.SourceID, sourceID)
	}

	if index.SourceType != sourceType {
		t.Errorf("SetSource() SourceType = %v, want %v", index.SourceType, sourceType)
	}

	if index.SourceURL != sourceURL {
		t.Errorf("SetSource() SourceURL = %v, want %v", index.SourceURL, sourceURL)
	}

	// 测试更新时间是否改变
	oldUpdatedTime := index.UpdatedTime
	time.Sleep(time.Millisecond)
	index.SetSource("newSource", "newType", "https://example.com/new")

	if index.UpdatedTime == oldUpdatedTime {
		t.Error("SetSource() should update UpdatedTime")
	}
}

func TestSearchIndex_SetMetadata(t *testing.T) {
	index := NewSearchIndex(SearchTypeRecord, "Test", "Content")

	// 设置元数据
	index.SetMetadata("author", "John Doe")
	index.SetMetadata("category", "documentation")

	if index.Metadata["author"] != "John Doe" {
		t.Errorf("SetMetadata() author = %v, want %v", index.Metadata["author"], "John Doe")
	}

	if index.Metadata["category"] != "documentation" {
		t.Errorf("SetMetadata() category = %v, want %v", index.Metadata["category"], "documentation")
	}

	// 测试更新时间是否改变
	oldUpdatedTime := index.UpdatedTime
	time.Sleep(time.Millisecond)
	index.SetMetadata("new_key", "new_value")

	if index.UpdatedTime == oldUpdatedTime {
		t.Error("SetMetadata() should update UpdatedTime")
	}
}

func TestSearchIndex_SetContext(t *testing.T) {
	index := NewSearchIndex(SearchTypeRecord, "Test", "Content")

	// 设置上下文
	userID := "user123"
	spaceID := "space456"
	tableID := "table789"
	fieldID := "field101"
	index.SetContext(userID, spaceID, tableID, fieldID)

	if index.UserID != userID {
		t.Errorf("SetContext() UserID = %v, want %v", index.UserID, userID)
	}

	if index.SpaceID != spaceID {
		t.Errorf("SetContext() SpaceID = %v, want %v", index.SpaceID, spaceID)
	}

	if index.TableID != tableID {
		t.Errorf("SetContext() TableID = %v, want %v", index.TableID, tableID)
	}

	if index.FieldID != fieldID {
		t.Errorf("SetContext() FieldID = %v, want %v", index.FieldID, fieldID)
	}

	// 测试更新时间是否改变
	oldUpdatedTime := index.UpdatedTime
	time.Sleep(time.Millisecond)
	index.SetContext("newUser", "newSpace", "newTable", "newField")

	if index.UpdatedTime == oldUpdatedTime {
		t.Error("SetContext() should update UpdatedTime")
	}
}

func TestSearchIndex_AddPermission(t *testing.T) {
	index := NewSearchIndex(SearchTypeRecord, "Test", "Content")

	// 添加权限
	index.AddPermission("read")
	index.AddPermission("write")

	if len(index.Permissions) != 2 {
		t.Errorf("AddPermission() Permissions length = %v, want %v", len(index.Permissions), 2)
	}

	if index.Permissions[0] != "read" {
		t.Errorf("AddPermission() Permissions[0] = %v, want %v", index.Permissions[0], "read")
	}

	if index.Permissions[1] != "write" {
		t.Errorf("AddPermission() Permissions[1] = %v, want %v", index.Permissions[1], "write")
	}

	// 测试重复添加
	index.AddPermission("read") // 应该不会重复添加

	if len(index.Permissions) != 2 {
		t.Errorf("AddPermission() duplicate Permissions length = %v, want %v", len(index.Permissions), 2)
	}

	// 测试更新时间是否改变
	oldUpdatedTime := index.UpdatedTime
	time.Sleep(time.Millisecond)
	index.AddPermission("delete")

	if index.UpdatedTime == oldUpdatedTime {
		t.Error("AddPermission() should update UpdatedTime")
	}
}

func TestSearchIndex_AddTag(t *testing.T) {
	index := NewSearchIndex(SearchTypeRecord, "Test", "Content")

	// 添加标签
	index.AddTag("important")
	index.AddTag("urgent")

	if len(index.Tags) != 2 {
		t.Errorf("AddTag() Tags length = %v, want %v", len(index.Tags), 2)
	}

	if index.Tags[0] != "important" {
		t.Errorf("AddTag() Tags[0] = %v, want %v", index.Tags[0], "important")
	}

	if index.Tags[1] != "urgent" {
		t.Errorf("AddTag() Tags[1] = %v, want %v", index.Tags[1], "urgent")
	}

	// 测试重复添加
	index.AddTag("important") // 应该不会重复添加

	if len(index.Tags) != 2 {
		t.Errorf("AddTag() duplicate Tags length = %v, want %v", len(index.Tags), 2)
	}

	// 测试更新时间是否改变
	oldUpdatedTime := index.UpdatedTime
	time.Sleep(time.Millisecond)
	index.AddTag("new_tag")

	if index.UpdatedTime == oldUpdatedTime {
		t.Error("AddTag() should update UpdatedTime")
	}
}

func TestNewSearchSuggestion(t *testing.T) {
	tests := []struct {
		name  string
		query string
		count int64
	}{
		{
			name:  "valid suggestion",
			query: "test query",
			count: 10,
		},
		{
			name:  "empty query",
			query: "",
			count: 0,
		},
		{
			name:  "zero count",
			query: "zero count query",
			count: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestion := NewSearchSuggestion(tt.query, tt.count)

			if suggestion == nil {
				t.Error("NewSearchSuggestion() returned nil")
				return
			}

			if suggestion.Query != tt.query {
				t.Errorf("NewSearchSuggestion() Query = %v, want %v", suggestion.Query, tt.query)
			}

			if suggestion.Count != tt.count {
				t.Errorf("NewSearchSuggestion() Count = %v, want %v", suggestion.Count, tt.count)
			}

			if suggestion.ID == "" {
				t.Error("NewSearchSuggestion() ID should not be empty")
			}

			if suggestion.CreatedTime.IsZero() {
				t.Error("NewSearchSuggestion() CreatedTime should not be zero")
			}

			if suggestion.UpdatedTime.IsZero() {
				t.Error("NewSearchSuggestion() UpdatedTime should not be zero")
			}
		})
	}
}

func TestSearchType_String(t *testing.T) {
	tests := []struct {
		searchType SearchType
		want       string
	}{
		{SearchTypeGlobal, "global"},
		{SearchTypeSpace, "space"},
		{SearchTypeTable, "table"},
		{SearchTypeRecord, "record"},
		{SearchTypeField, "field"},
		{SearchTypeUser, "user"},
		{SearchTypeComment, "comment"},
		{SearchTypeAttachment, "attachment"},
	}

	for _, tt := range tests {
		t.Run(string(tt.searchType), func(t *testing.T) {
			if got := string(tt.searchType); got != tt.want {
				t.Errorf("SearchType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSearchScope_String(t *testing.T) {
	tests := []struct {
		searchScope SearchScope
		want        string
	}{
		{SearchScopeAll, "all"},
		{SearchScopeTitle, "title"},
		{SearchScopeContent, "content"},
		{SearchScopeMetadata, "metadata"},
		{SearchScopeComments, "comments"},
		{SearchScopeAttachments, "attachments"},
	}

	for _, tt := range tests {
		t.Run(string(tt.searchScope), func(t *testing.T) {
			if got := string(tt.searchScope); got != tt.want {
				t.Errorf("SearchScope.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// 这些测试已经在notification包的测试中覆盖，这里不需要重复
