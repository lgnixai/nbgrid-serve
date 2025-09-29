# Glide Data Grid 全面使用指南

## 概述

Glide Data Grid 是一个高性能的 React 数据表格库，专为现代 Web 应用程序设计。它提供了丰富的功能，包括虚拟滚动、自定义单元格、编辑支持、复制粘贴等。

## 安装和设置

### 1. 安装依赖

```bash
npm install @glideapps/glide-data-grid
npm install lodash marked react-responsive-carousel
```

### 2. 导入 CSS

```javascript
import "@glideapps/glide-data-grid/dist/index.css";
```

## 基础使用

### 1. 基本组件结构

```typescript
import { DataEditor, GridCellKind, GridColumn, Item, GridCell } from '@glideapps/glide-data-grid';

const BasicGrid = () => {
  const [data, setData] = useState([
    { name: "张三", company: "ABC公司", email: "zhangsan@abc.com" }
  ]);

  const columns: GridColumn[] = [
    { title: "姓名", id: "name" },
    { title: "公司", id: "company" },
    { title: "邮箱", id: "email" }
  ];

  const getCellContent = useCallback((cell: Item): GridCell => {
    const [col, row] = cell;
    const dataRow = data[row];
    const keys = ["name", "company", "email"];
    const value = dataRow[keys[col]];
    
    return {
      kind: GridCellKind.Text,
      data: value,
      displayData: value,
      allowOverlay: false,
    };
  }, [data]);

  return (
    <DataEditor
      getCellContent={getCellContent}
      columns={columns}
      rows={data.length}
    />
  );
};
```

## 单元格类型

### 1. 文本单元格

```typescript
{
  kind: GridCellKind.Text,
  data: "文本内容",
  displayData: "文本内容",
  allowOverlay: true,
}
```

### 2. 数字单元格

```typescript
{
  kind: GridCellKind.Number,
  data: 123,
  displayData: "123",
  allowOverlay: true,
}
```

### 3. 布尔单元格

```typescript
{
  kind: GridCellKind.Boolean,
  data: true,
  displayData: "是",
  allowOverlay: true,
}
```

### 4. URI 单元格

```typescript
{
  kind: GridCellKind.Uri,
  data: "https://example.com",
  displayData: "https://example.com",
  allowOverlay: true,
}
```

### 5. 图片单元格

```typescript
{
  kind: GridCellKind.Image,
  data: ["image1.jpg", "image2.jpg"],
  displayData: "图片",
  allowOverlay: false,
}
```

### 6. Markdown 单元格

```typescript
{
  kind: GridCellKind.Markdown,
  data: "# 标题\n\n内容",
  displayData: "Markdown 内容",
  allowOverlay: true,
}
```

## 编辑功能

### 1. 启用编辑

```typescript
const onCellEdited = useCallback((cell: Item, newValue: EditableGridCell) => {
  const [col, row] = cell;
  // 处理编辑逻辑
  setData(prev => {
    const newData = [...prev];
    newData[row] = { ...newData[row], [keys[col]]: newValue.data };
    return newData;
  });
}, []);

<DataEditor
  getCellContent={getCellContent}
  columns={columns}
  rows={data.length}
  onCellEdited={onCellEdited}
/>
```

### 2. 只读单元格

```typescript
{
  kind: GridCellKind.Text,
  data: "只读内容",
  displayData: "只读内容",
  allowOverlay: false,
  readonly: true,
}
```

## 选择功能

### 1. 单元格选择

```typescript
const [selectedCells, setSelectedCells] = useState<Item[]>([]);

const onSelectionChange = useCallback((newSelection: Item[]) => {
  setSelectedCells(newSelection);
}, []);

<DataEditor
  // ... 其他属性
  onSelectionChange={onSelectionChange}
/>
```

### 2. 行选择

```typescript
const [selectedRows, setSelectedRows] = useState<CompactSelection>(CompactSelection.empty());

const onRowSelectionChange = useCallback((newSelection: CompactSelection) => {
  setSelectedRows(newSelection);
}, []);

<DataEditor
  // ... 其他属性
  rowSelect="multi" // 或 "single"
  onRowSelectionChange={onRowSelectionChange}
/>
```

## 复制粘贴功能

### 1. 启用复制粘贴

```typescript
const getCellsForSelection = useCallback((selection: any) => {
  const cells: GridCell[][] = [];
  
  for (let row = selection.y; row < selection.y + selection.height; row++) {
    const rowCells: GridCell[] = [];
    for (let col = selection.x; col < selection.x + selection.width; col++) {
      rowCells.push(getCellContent([col, row]));
    }
    cells.push(rowCells);
  }
  
  return cells;
}, [getCellContent]);

const onPaste = useCallback((target: Item, values: string[][]) => {
  // 处理粘贴逻辑
}, []);

<DataEditor
  // ... 其他属性
  getCellsForSelection={getCellsForSelection}
  onPaste={onPaste}
/>
```

## 自定义单元格渲染器

### 1. 创建自定义渲染器

```typescript
const CustomStatusRenderer: CustomCellRenderer = {
  kind: GridCellKind.Custom,
  isMatch: (cell: GridCell): cell is GridCell & { data: string } => 
    cell.kind === GridCellKind.Custom && cell.data?.type === 'status',
  
  draw: (args: CustomCellRendererProps) => {
    const { ctx, cell, theme, rect } = args;
    const { data } = cell as GridCell & { data: { status: string } };
    
    // 绘制状态指示器
    ctx.fillStyle = data.status === 'online' ? '#10b981' : '#ef4444';
    ctx.beginPath();
    ctx.arc(rect.x + 10, rect.y + rect.height / 2, 4, 0, 2 * Math.PI);
    ctx.fill();
    
    // 绘制文本
    ctx.fillStyle = theme.textDark;
    ctx.font = theme.baseFontStyle;
    ctx.fillText(data.status, rect.x + 20, rect.y + rect.height / 2 + 4);
  },
  
  provideEditor: () => undefined,
  onPaste: (value: string) => ({
    kind: GridCellKind.Custom,
    data: { type: 'status', status: value },
    allowOverlay: true,
  })
};
```

### 2. 使用自定义渲染器

```typescript
<DataEditor
  // ... 其他属性
  customRenderers={[CustomStatusRenderer]}
/>
```

## 搜索功能

### 1. 实现搜索

```typescript
const [searchText, setSearchText] = useState('');

const filteredData = useMemo(() => {
  if (!searchText) return data;
  
  return data.filter(row => 
    row.name.toLowerCase().includes(searchText.toLowerCase())
  );
}, [data, searchText]);

const searchResults = useMemo(() => {
  if (!searchText) return undefined;
  
  return filteredData.map((_, index) => index);
}, [filteredData, searchText]);

<DataEditor
  // ... 其他属性
  searchResults={searchResults}
/>
```

## 主题配置

### 1. 自定义主题

```typescript
const customTheme = {
  accentColor: "#0ea5e9",
  accentFg: "#ffffff",
  accentLight: "#e0f2fe",
  textDark: "#1f2937",
  textMedium: "#6b7280",
  textLight: "#9ca3af",
  textBubble: "#ffffff",
  bgIconHeader: "#6b7280",
  fgIconHeader: "#ffffff",
  textHeader: "#374151",
  textHeaderSelected: "#1f2937",
  bgCell: "#ffffff",
  bgCellMedium: "#f9fafb",
  bgHeader: "#f3f4f6",
  bgHeaderHasFocus: "#e5e7eb",
  bgHeaderHovered: "#e5e7eb",
  bgBubble: "#1f2937",
  bgBubbleSelected: "#111827",
  bgSearchResult: "#fef3c7",
  borderColor: "#e5e7eb",
  drilldownBorder: "#d1d5db",
  linkColor: "#0ea5e9",
  headerFontStyle: "600 14px",
  baseFontStyle: "14px",
  fontFamily: "ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, 'Noto Sans', sans-serif",
};

<DataEditor
  // ... 其他属性
  theme={customTheme}
/>
```

## 性能优化

### 1. 虚拟滚动

Glide Data Grid 默认启用虚拟滚动，无需额外配置。

### 2. 数据懒加载

```typescript
const getCellContent = useCallback((cell: Item): GridCell => {
  const [col, row] = cell;
  
  // 检查数据是否已加载
  if (!data[row]) {
    return {
      kind: GridCellKind.Loading,
      allowOverlay: false,
    };
  }
  
  // 返回实际数据
  return createCell(data[row], col);
}, [data]);
```

### 3. 回调函数优化

```typescript
// 使用 useCallback 避免不必要的重新渲染
const getCellContent = useCallback((cell: Item): GridCell => {
  // 实现逻辑
}, [data, columns]);

const onCellEdited = useCallback((cell: Item, newValue: EditableGridCell) => {
  // 实现逻辑
}, [data, columns]);
```

## 与 Teable 集成

### 1. 数据映射

```typescript
// 将 Teable 字段类型映射到 Glide 单元格类型
const mapTeableFieldToGlideCell = (field: TeableField, value: any): GridCell => {
  switch (field.type) {
    case 'singleLineText':
    case 'longText':
      return {
        kind: GridCellKind.Text,
        data: value || '',
        displayData: value || '',
        allowOverlay: true,
      };
    
    case 'number':
      return {
        kind: GridCellKind.Number,
        data: value || 0,
        displayData: value?.toString() || '0',
        allowOverlay: true,
      };
    
    case 'checkbox':
      return {
        kind: GridCellKind.Boolean,
        data: Boolean(value),
        displayData: value ? '是' : '否',
        allowOverlay: true,
      };
    
    // 更多类型映射...
  }
};
```

### 2. 实时同步

```typescript
// 监听 Teable 数据变化
useEffect(() => {
  const subscription = teable.subscribeToTableChanges(tableId, (changes) => {
    setRecords(prev => {
      // 应用变更
      return applyChanges(prev, changes);
    });
  });
  
  return () => subscription.unsubscribe();
}, [tableId]);
```

## 最佳实践

### 1. 数据管理

- 使用状态管理库（如 Redux、Zustand）管理大量数据
- 实现数据分页或虚拟滚动
- 缓存频繁访问的数据

### 2. 性能优化

- 避免在 `getCellContent` 中进行复杂计算
- 使用 `useMemo` 和 `useCallback` 优化渲染
- 合理设置列宽和行高

### 3. 用户体验

- 提供加载状态指示器
- 实现错误处理和重试机制
- 添加键盘快捷键支持

### 4. 可访问性

- 确保键盘导航支持
- 提供屏幕阅读器支持
- 使用语义化的 HTML 结构

## 常见问题

### 1. 性能问题

**问题**: 大量数据时渲染缓慢
**解决方案**: 
- 启用虚拟滚动
- 实现数据分页
- 优化 `getCellContent` 函数

### 2. 编辑问题

**问题**: 编辑后数据不更新
**解决方案**: 
- 检查 `onCellEdited` 回调实现
- 确保状态更新正确
- 验证数据类型匹配

### 3. 样式问题

**问题**: 自定义样式不生效
**解决方案**: 
- 检查 CSS 导入
- 使用正确的主题配置
- 验证样式优先级

## 总结

Glide Data Grid 是一个功能强大且灵活的数据表格库，通过合理使用其各种功能，可以构建出高性能、用户友好的数据展示界面。关键是要理解其核心概念，合理配置各种属性，并根据具体需求进行定制化开发。
