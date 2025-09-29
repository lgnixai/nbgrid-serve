import React, { useState, useCallback, useMemo, useRef } from 'react';
import { 
  DataEditor, 
  GridCellKind, 
  GridColumn, 
  Item, 
  GridCell, 
  EditableGridCell,
  CompactSelection,
  GridSelection,
  CustomCellRenderer,
  CustomCellRendererProps
} from '@glideapps/glide-data-grid';
import '@glideapps/glide-data-grid/dist/index.css';

// 高级数据类型
interface AdvancedDataRow {
  id: string;
  name: string;
  avatar: string;
  status: 'online' | 'offline' | 'away';
  progress: number;
  tags: string[];
  rating: number;
  lastActive: Date;
  customData: any;
}

// 示例数据
const advancedSampleData: AdvancedDataRow[] = [
  {
    id: '1',
    name: '张三',
    avatar: '👨‍💻',
    status: 'online',
    progress: 75,
    tags: ['React', 'TypeScript'],
    rating: 4.5,
    lastActive: new Date('2024-01-15'),
    customData: { department: 'Engineering', level: 'Senior' }
  },
  {
    id: '2',
    name: '李四',
    avatar: '👩‍🎨',
    status: 'away',
    progress: 60,
    tags: ['Design', 'UI/UX'],
    rating: 4.2,
    lastActive: new Date('2024-01-14'),
    customData: { department: 'Design', level: 'Mid' }
  },
  {
    id: '3',
    name: '王五',
    avatar: '👨‍💼',
    status: 'offline',
    progress: 90,
    tags: ['Management', 'Strategy'],
    rating: 4.8,
    lastActive: new Date('2024-01-13'),
    customData: { department: 'Management', level: 'Director' }
  }
];

// 1. 自定义单元格渲染器 - 状态指示器
const StatusCellRenderer: CustomCellRenderer = {
  kind: GridCellKind.Custom,
  isMatch: (cell: GridCell): cell is GridCell & { data: string } => 
    cell.kind === GridCellKind.Custom && cell.data?.type === 'status',
  
  draw: (args: CustomCellRendererProps) => {
    const { ctx, cell, theme, rect } = args;
    const { data } = cell as GridCell & { data: { status: string } };
    
    // 根据状态设置颜色
    let color = '#6b7280'; // 默认灰色
    switch (data.status) {
      case 'online':
        color = '#10b981'; // 绿色
        break;
      case 'away':
        color = '#f59e0b'; // 黄色
        break;
      case 'offline':
        color = '#ef4444'; // 红色
        break;
    }
    
    // 绘制状态圆点
    ctx.fillStyle = color;
    ctx.beginPath();
    ctx.arc(rect.x + 10, rect.y + rect.height / 2, 4, 0, 2 * Math.PI);
    ctx.fill();
    
    // 绘制状态文本
    ctx.fillStyle = theme.textDark;
    ctx.font = theme.baseFontStyle;
    ctx.fillText(data.status, rect.x + 20, rect.y + rect.height / 2 + 4);
  },
  
  provideEditor: () => undefined,
  onPaste: (value: string) => {
    return {
      kind: GridCellKind.Custom,
      data: { type: 'status', status: value },
      allowOverlay: true,
    };
  }
};

// 2. 自定义单元格渲染器 - 进度条
const ProgressCellRenderer: CustomCellRenderer = {
  kind: GridCellKind.Custom,
  isMatch: (cell: GridCell): cell is GridCell & { data: number } => 
    cell.kind === GridCellKind.Custom && cell.data?.type === 'progress',
  
  draw: (args: CustomCellRendererProps) => {
    const { ctx, cell, theme, rect } = args;
    const { data } = cell as GridCell & { data: { progress: number } };
    
    const progress = Math.max(0, Math.min(100, data.progress));
    const barWidth = (rect.width - 20) * (progress / 100);
    
    // 绘制背景
    ctx.fillStyle = theme.bgCellMedium;
    ctx.fillRect(rect.x + 10, rect.y + 8, rect.width - 20, 8);
    
    // 绘制进度条
    ctx.fillStyle = progress > 80 ? '#10b981' : progress > 50 ? '#f59e0b' : '#ef4444';
    ctx.fillRect(rect.x + 10, rect.y + 8, barWidth, 8);
    
    // 绘制百分比文本
    ctx.fillStyle = theme.textDark;
    ctx.font = theme.baseFontStyle;
    ctx.fillText(`${progress}%`, rect.x + 10, rect.y + rect.height - 4);
  },
  
  provideEditor: () => undefined,
  onPaste: (value: string) => {
    const num = parseFloat(value);
    return {
      kind: GridCellKind.Custom,
      data: { type: 'progress', progress: isNaN(num) ? 0 : num },
      allowOverlay: true,
    };
  }
};

// 3. 自定义单元格渲染器 - 标签
const TagsCellRenderer: CustomCellRenderer = {
  kind: GridCellKind.Custom,
  isMatch: (cell: GridCell): cell is GridCell & { data: string[] } => 
    cell.kind === GridCellKind.Custom && cell.data?.type === 'tags',
  
  draw: (args: CustomCellRendererProps) => {
    const { ctx, cell, theme, rect } = args;
    const { data } = cell as GridCell & { data: { tags: string[] } };
    
    let x = rect.x + 5;
    const y = rect.y + rect.height / 2 + 4;
    
    data.tags.forEach((tag, index) => {
      if (x > rect.x + rect.width - 20) return; // 防止溢出
      
      // 绘制标签背景
      ctx.fillStyle = '#e0f2fe';
      const textWidth = ctx.measureText(tag).width;
      const tagWidth = textWidth + 8;
      
      ctx.fillRect(x, rect.y + 4, tagWidth, rect.height - 8);
      
      // 绘制标签文本
      ctx.fillStyle = '#0369a1';
      ctx.font = theme.baseFontStyle;
      ctx.fillText(tag, x + 4, y);
      
      x += tagWidth + 4;
    });
  },
  
  provideEditor: () => undefined,
  onPaste: (value: string) => {
    return {
      kind: GridCellKind.Custom,
      data: { type: 'tags', tags: value.split(',').map(t => t.trim()) },
      allowOverlay: true,
    };
  }
};

export const AdvancedGlideGrid: React.FC = () => {
  const [data, setData] = useState<AdvancedDataRow[]>(advancedSampleData);
  const [searchText, setSearchText] = useState('');
  const [selectedRows, setSelectedRows] = useState<CompactSelection>(CompactSelection.empty());
  const dataEditorRef = useRef<DataEditor>(null);

  // 高级列定义
  const columns: GridColumn[] = useMemo(() => [
    {
      title: 'ID',
      id: 'id',
      width: 60,
      resizable: false,
    },
    {
      title: '头像',
      id: 'avatar',
      width: 80,
      resizable: false,
    },
    {
      title: '姓名',
      id: 'name',
      width: 120,
    },
    {
      title: '状态',
      id: 'status',
      width: 100,
    },
    {
      title: '进度',
      id: 'progress',
      width: 120,
    },
    {
      title: '标签',
      id: 'tags',
      width: 200,
    },
    {
      title: '评分',
      id: 'rating',
      width: 100,
    },
    {
      title: '最后活跃',
      id: 'lastActive',
      width: 150,
    }
  ], []);

  // 过滤数据
  const filteredData = useMemo(() => {
    if (!searchText) return data;
    
    return data.filter(row => 
      row.name.toLowerCase().includes(searchText.toLowerCase()) ||
      row.tags.some(tag => tag.toLowerCase().includes(searchText.toLowerCase()))
    );
  }, [data, searchText]);

  // 获取单元格内容
  const getCellContent = useCallback((cell: Item): GridCell => {
    const [col, row] = cell;
    const dataRow = filteredData[row];
    
    if (!dataRow) {
      return {
        kind: GridCellKind.Loading,
        allowOverlay: false,
      };
    }

    const columnId = columns[col]?.id;
    
    switch (columnId) {
      case 'id':
        return {
          kind: GridCellKind.Text,
          data: dataRow.id,
          displayData: dataRow.id,
          allowOverlay: false,
          readonly: true,
        };
      
      case 'avatar':
        return {
          kind: GridCellKind.Text,
          data: dataRow.avatar,
          displayData: dataRow.avatar,
          allowOverlay: false,
          readonly: true,
        };
      
      case 'name':
        return {
          kind: GridCellKind.Text,
          data: dataRow.name,
          displayData: dataRow.name,
          allowOverlay: true,
        };
      
      case 'status':
        return {
          kind: GridCellKind.Custom,
          data: { type: 'status', status: dataRow.status },
          allowOverlay: true,
        };
      
      case 'progress':
        return {
          kind: GridCellKind.Custom,
          data: { type: 'progress', progress: dataRow.progress },
          allowOverlay: true,
        };
      
      case 'tags':
        return {
          kind: GridCellKind.Custom,
          data: { type: 'tags', tags: dataRow.tags },
          allowOverlay: true,
        };
      
      case 'rating':
        return {
          kind: GridCellKind.Number,
          data: dataRow.rating,
          displayData: dataRow.rating.toString(),
          allowOverlay: true,
        };
      
      case 'lastActive':
        return {
          kind: GridCellKind.Text,
          data: dataRow.lastActive.toLocaleDateString(),
          displayData: dataRow.lastActive.toLocaleDateString(),
          allowOverlay: true,
        };
      
      default:
        return {
          kind: GridCellKind.Text,
          data: '',
          displayData: '',
          allowOverlay: false,
        };
    }
  }, [filteredData, columns]);

  // 单元格编辑处理
  const onCellEdited = useCallback((cell: Item, newValue: EditableGridCell) => {
    const [col, row] = cell;
    const columnId = columns[col]?.id;
    
    if (!columnId) return;

    setData(prevData => {
      const newData = [...prevData];
      const originalRowIndex = data.findIndex(d => d.id === filteredData[row].id);
      
      if (originalRowIndex === -1) return prevData;
      
      const updatedRow = { ...newData[originalRowIndex] };
      
      switch (columnId) {
        case 'name':
          if (newValue.kind === GridCellKind.Text) {
            updatedRow.name = newValue.data;
          }
          break;
        
        case 'rating':
          if (newValue.kind === GridCellKind.Number) {
            updatedRow.rating = newValue.data;
          }
          break;
      }
      
      newData[originalRowIndex] = updatedRow;
      return newData;
    });
  }, [columns, data, filteredData]);

  // 行选择处理
  const onRowSelectionChange = useCallback((newSelection: CompactSelection) => {
    setSelectedRows(newSelection);
  }, []);

  // 拖拽处理
  const onDragStart = useCallback((event: React.DragEvent) => {
    if (selectedRows.length > 0) {
      const selectedData = Array.from(selectedRows).map(index => filteredData[index]);
      event.dataTransfer.setData('application/json', JSON.stringify(selectedData));
    }
  }, [selectedRows, filteredData]);

  // 搜索功能
  const handleSearch = useCallback((text: string) => {
    setSearchText(text);
  }, []);

  // 添加新行
  const addNewRow = useCallback(() => {
    const newRow: AdvancedDataRow = {
      id: (data.length + 1).toString(),
      name: '新用户',
      avatar: '👤',
      status: 'offline',
      progress: 0,
      tags: [],
      rating: 0,
      lastActive: new Date(),
      customData: {}
    };
    
    setData(prev => [...prev, newRow]);
  }, [data.length]);

  return (
    <div className="w-full h-full p-4">
      <div className="mb-4">
        <h2 className="text-2xl font-bold mb-2">Glide Data Grid 高级功能示例</h2>
        <p className="text-gray-600 mb-4">
          展示自定义单元格、拖拽、搜索、行选择等高级功能
        </p>
        
        {/* 控制面板 */}
        <div className="mb-4 p-4 bg-gray-100 rounded-lg">
          <div className="flex items-center space-x-4 mb-4">
            <div>
              <label className="block text-sm font-medium mb-1">搜索:</label>
              <input
                type="text"
                value={searchText}
                onChange={(e) => handleSearch(e.target.value)}
                placeholder="搜索姓名或标签..."
                className="px-3 py-1 border rounded-md"
              />
            </div>
            
            <button
              onClick={addNewRow}
              className="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600"
            >
              添加新行
            </button>
            
            <div className="text-sm text-gray-600">
              已选择 {selectedRows.length} 行
            </div>
          </div>
          
          <div className="text-sm text-gray-600">
            <p>功能说明：</p>
            <ul className="list-disc list-inside space-y-1">
              <li>状态列：显示在线状态指示器</li>
              <li>进度列：可视化进度条</li>
              <li>标签列：多标签显示</li>
              <li>支持行选择和拖拽</li>
              <li>实时搜索过滤</li>
            </ul>
          </div>
        </div>
      </div>

      <div 
        className="border rounded-lg overflow-hidden" 
        style={{ height: '600px' }}
        onDragStart={onDragStart}
      >
        <DataEditor
          ref={dataEditorRef}
          getCellContent={getCellContent}
          columns={columns}
          rows={filteredData.length}
          
          // 编辑功能
          onCellEdited={onCellEdited}
          
          // 行选择
          rowSelect="multi"
          onRowSelectionChange={onRowSelectionChange}
          
          // 自定义渲染器
          customRenderers={[StatusCellRenderer, ProgressCellRenderer, TagsCellRenderer]}
          
          // 拖拽支持
          onDragStart={onDragStart}
          
          // 搜索功能
          searchResults={searchText ? 
            filteredData.map((_, index) => index) : 
            undefined
          }
          
          // 主题配置
          theme={{
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
          }}
          
          // 其他配置
          smoothScrollX={true}
          smoothScrollY={true}
          overscrollX={0}
          overscrollY={0}
          gridSelection={{
            columns: CompactSelection.empty(),
            rows: selectedRows,
          }}
        />
      </div>
    </div>
  );
};
