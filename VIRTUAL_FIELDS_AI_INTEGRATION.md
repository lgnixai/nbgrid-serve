# 虚拟字段和AI字段集成文档

## 概述

本文档介绍了如何在现有系统中集成飞书多维表格风格的虚拟字段和AI支持字段功能。

## 新增功能

### 1. 虚拟字段类型

#### 1.1 公式字段 (Formula Field)
- **类型标识**: `virtual_formula`
- **功能**: 基于其他字段值进行计算
- **特性**:
  - 支持数学运算、逻辑运算、字符串操作
  - 实时计算或缓存计算结果
  - 支持自定义函数（ABS、ROUND、CONCAT、IF、LEN等）

```json
{
  "type": "virtual_formula",
  "options": {
    "expression": "ROUND({price} * {quantity}, 2)",
    "result_type": "number",
    "dynamic_calculation": true
  }
}
```

#### 1.2 查找引用字段 (Lookup Field)
- **类型标识**: `virtual_lookup`
- **功能**: 从关联记录中查找并显示字段值
- **特性**:
  - 支持多种处理方式（first、last、array、comma_separated）
  - 自动跟随关联记录更新

```json
{
  "type": "virtual_lookup",
  "options": {
    "link_field_id": "field_123",
    "lookup_field_id": "field_456",
    "multiple_record_handling": "comma_separated"
  }
}
```

#### 1.3 汇总统计字段 (Rollup Field)
- **类型标识**: `virtual_rollup`
- **功能**: 对关联记录的字段值进行聚合计算
- **特性**:
  - 支持多种聚合函数（sum、avg、count、min、max、unique_count）
  - 支持过滤条件

```json
{
  "type": "virtual_rollup",
  "options": {
    "link_field_id": "field_123",
    "rollup_field_id": "field_456",
    "aggregation_function": "sum",
    "filter_expression": "status = 'completed'"
  }
}
```

#### 1.4 AI智能字段 (AI Field)
- **类型标识**: `virtual_ai`
- **功能**: 使用AI生成、提取或处理内容
- **操作类型**:
  - `generate`: 生成内容
  - `extract`: 提取信息
  - `classify`: 分类
  - `summarize`: 总结
  - `translate`: 翻译

```json
{
  "type": "virtual_ai",
  "options": {
    "operation_type": "summarize",
    "provider": "openai",
    "model": "gpt-3.5-turbo",
    "prompt_template": "请总结以下内容：{content}",
    "source_fields": ["content"],
    "output_format": "text",
    "cache_results": true
  }
}
```

### 2. 字段捷径系统

预配置的字段模板，快速创建常用字段：

- **自动排名**: 根据指定字段自动计算排名
- **AI摘要**: 使用AI自动生成内容摘要
- **AI分类**: 使用AI自动对内容进行分类
- **关联总数**: 统计关联记录的总数
- **关联求和**: 对关联记录的数值字段求和

### 3. AI提供商支持

#### 3.1 已实现的提供商
- **OpenAI**: 支持GPT-3.5、GPT-4等模型
- **DeepSeek**: 支持DeepSeek Chat模型（OpenAI兼容API）

#### 3.2 计划支持的提供商
- **Anthropic**: Claude系列模型
- **百度文心一言**
- **阿里通义千问**

### 4. 核心组件

#### 4.1 虚拟字段服务 (VirtualFieldService)
- 管理虚拟字段的计算
- 处理字段依赖关系
- 实现缓存机制
- 自动更新依赖字段

#### 4.2 AI提供商工厂 (ProviderFactory)
- 根据配置创建AI提供商
- 支持多提供商切换
- 统一的API接口

#### 4.3 字段类型注册表 (FieldTypeRegistry)
- 插件化的字段类型管理
- 支持自定义字段类型
- 字段类型验证和转换

## API接口

### 1. 获取字段类型
```
GET /api/fields/types
```

### 2. 获取字段捷径
```
GET /api/fields/shortcuts
GET /api/fields/shortcuts?category=AI
GET /api/fields/shortcuts?tag=summary
```

### 3. 获取字段捷径详情
```
GET /api/fields/shortcuts/{id}
```

## 配置示例

```yaml
# AI配置
ai:
  default_provider: openai
  providers:
    openai:
      type: openai
      api_key: ${OPENAI_API_KEY}
      default_model: gpt-3.5-turbo
      timeout: 30
      rate_limit:
        requests_per_minute: 60
        tokens_per_minute: 90000
        concurrent_requests: 10
    deepseek:
      type: deepseek
      api_key: ${DEEPSEEK_API_KEY}
      base_url: https://api.deepseek.com/v1
      default_model: deepseek-chat
```

## 使用示例

### 1. 创建公式字段
```javascript
const formulaField = {
  name: "总价",
  type: "virtual_formula",
  options: {
    expression: "{单价} * {数量}",
    result_type: "number"
  }
};
```

### 2. 创建AI摘要字段
```javascript
const aiSummaryField = {
  name: "内容摘要",
  type: "virtual_ai",
  options: {
    operation_type: "summarize",
    provider: "openai",
    prompt_template: "请用100字以内总结：{content}",
    source_fields: ["content"]
  }
};
```

### 3. 使用字段捷径
```javascript
// 获取AI摘要捷径
const shortcut = await getFieldShortcut("ai_summary");

// 基于捷径创建字段
const field = {
  name: "产品描述摘要",
  ...shortcut,
  options: {
    ...shortcut.options,
    source_fields: ["product_description"]
  }
};
```

## 架构优势

1. **可扩展性**: 插件化架构，易于添加新的字段类型和AI提供商
2. **性能优化**: 内置缓存机制，避免重复计算
3. **依赖管理**: 自动处理字段间的依赖关系
4. **错误处理**: 优雅的错误处理，不影响其他字段
5. **配置灵活**: 支持多种AI提供商，可按需切换

## 后续计划

1. 实现Lookup和Rollup字段的完整功能
2. 添加更多AI提供商支持
3. 实现字段计算的批量优化
4. 添加更多预定义的字段捷径
5. 支持自定义函数扩展
6. 实现字段权限控制
7. 添加字段计算的监控和日志

## 注意事项

1. AI字段需要配置相应的API密钥
2. 虚拟字段是只读的，不能直接修改其值
3. 公式字段的表达式需要正确引用其他字段
4. AI操作可能产生费用，建议合理使用缓存
5. 某些虚拟字段类型可能影响查询性能