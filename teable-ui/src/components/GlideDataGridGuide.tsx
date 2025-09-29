import React, { useState, useCallback, useMemo } from 'react';
import { DataEditor, GridCellKind, GridColumn, Item, GridCell, EditableGridCell, CompactSelection } from '@glideapps/glide-data-grid';
import '@glideapps/glide-data-grid/dist/index.css';

// 基础数据类型定义
interface DataRow {
  id: string;
  name: string;
  company: string;
  email: string;
  phone: string;
  status: 'active' | 'inactive';
  score: number;
  date: string;
}

// 示例数据
const sampleData: DataRow[] = [
  {
    id: '1',
    name: '张三',
    company: 'ABC公司',
    email: 'zhangsan@abc.com',
    phone: '+86 123 4567 8901',
    status: 'active',
    score: 95,
    date: '2024-01-15'
  },
  {
    id: '2',
    name: '李四',
    company: 'XYZ科技',
    email: 'lisi@xyz.com',
    phone: '+86 987 6543 2109',
    status: 'inactive',
    score: 87,
    date: '2024-01-20'
  },
  {
    id: '3',
    name: '王五',
    company: 'DEF集团',
    email: 'wangwu@def.com',
    phone: '+86 555 1234 5678',
    status: 'active',
    score: 92,
    date: '2024-01-25'
  }
];

export const GlideDataGridGuide: React.FC = () => {
  const [data, setData] = useState<DataRow[]>(sampleData);
  const [selectedCells, setSelectedCells] = useState<Item[]>([]);

  // 1. 基础列定义
  const columns: GridColumn[] = useMemo(() => [
    {
      title: 'ID',
      id: 'id',
      width: 80,
      resizable: false,
    },
    {
      title: '姓名',
      id: 'name',
      width: 120,
    },
    {
      title: '公司',
      id: 'company',
      width: 150,
    },
    {
      title: '邮箱',
      id: 'email',
      width: 200,
    },
    {
      title: '电话',
      id: 'phone',
      width: 150,
    },
    {
      title: '状态',
      id: 'status',
      width: 100,
    },
    {
      title: '评分',
      id: 'score',
      width: 100,
    },
    {
      title: '日期',
      id: 'date',
      width: 120,
    }
  ], []);

  // 2. 核心回调函数 - 获取单元格内容
  const getCellContent = useCallback((cell: Item): GridCell => {
    const [col, row] = cell;
    const dataRow = data[row];
    
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
      
      case 'name':
        return {
          kind: GridCellKind.Text,
          data: dataRow.name,
          displayData: dataRow.name,
          allowOverlay: true,
        };
      
      case 'company':
        return {
          kind: GridCellKind.Text,
          data: dataRow.company,
          displayData: dataRow.company,
          allowOverlay: true,
        };
      
      case 'email':
        return {
          kind: GridCellKind.Uri,
          data: dataRow.email,
          displayData: dataRow.email,
          allowOverlay: true,
        };
      
      case 'phone':
        return {
          kind: GridCellKind.Text,
          data: dataRow.phone,
          displayData: dataRow.phone,
          allowOverlay: true,
        };
      
      case 'status':
        return {
          kind: GridCellKind.Boolean,
          data: dataRow.status === 'active',
          allowOverlay: false,
        };
      
      case 'score':
        return {
          kind: GridCellKind.Number,
          data: dataRow.score,
          displayData: dataRow.score.toString(),
          allowOverlay: true,
        };
      
      case 'date':
        return {
          kind: GridCellKind.Text,
          data: dataRow.date,
          displayData: dataRow.date,
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
  }, [data, columns]);

  // 3. 单元格编辑处理
  const onCellEdited = useCallback((cell: Item, newValue: EditableGridCell) => {
    const [col, row] = cell;
    const columnId = columns[col]?.id;
    
    if (!columnId) return;

    setData(prevData => {
      const newData = [...prevData];
      const updatedRow = { ...newData[row] };

      switch (columnId) {
        case 'name':
        case 'company':
        case 'phone':
        case 'date':
          if (newValue.kind === GridCellKind.Text) {
            updatedRow[columnId] = newValue.data;
          }
          break;
        
        case 'email':
          if (newValue.kind === GridCellKind.Uri) {
            updatedRow.email = newValue.data;
          }
          break;
        
        case 'status':
          if (newValue.kind === GridCellKind.Boolean) {
            updatedRow.status = newValue.data ? 'active' : 'inactive';
          }
          break;
        
        case 'score':
          if (newValue.kind === GridCellKind.Number) {
            updatedRow.score = newValue.data;
          }
          break;
      }

      newData[row] = updatedRow;
      return newData;
    });
  }, [columns]);

  // 4. 选择处理
  const onSelectionChange = useCallback((newSelection: Item[]) => {
    setSelectedCells(newSelection);
  }, []);

  // 5. 复制粘贴支持
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

  // 6. 粘贴处理
  const onPaste = useCallback((target: Item, values: readonly (readonly string[])[]): boolean => {
    const [startCol, startRow] = target;
    
    setData(prevData => {
      const newData = [...prevData];
      
      values.forEach((row, rowIndex) => {
        const targetRow = startRow + rowIndex;
        if (targetRow >= newData.length) return;
        
        row.forEach((cellValue, colIndex) => {
          const targetCol = startCol + colIndex;
          const columnId = columns[targetCol]?.id;
          
          if (!columnId) return;
          
          const updatedRow = { ...newData[targetRow] };
          
          switch (columnId) {
            case 'name':
            case 'company':
            case 'phone':
            case 'date':
              updatedRow[columnId] = cellValue;
              break;
            case 'email':
              updatedRow.email = cellValue;
              break;
            case 'status':
              updatedRow.status = cellValue === 'true' || cellValue === 'active' ? 'active' : 'inactive';
              break;
            case 'score':
              const numValue = parseFloat(cellValue);
              if (!isNaN(numValue)) {
                updatedRow.score = numValue;
              }
              break;
          }
          
          newData[targetRow] = updatedRow;
        });
      });
      
      return newData;
    });
    
    return true;
  }, [columns]);

  // 7. 行标记配置
  const rowMarkers = "number"; // 可选: "number" | "checkbox" | "both" | "none"

  return (
    <div className="w-full h-full p-4">
      <div className="mb-4">
        <h2 className="text-2xl font-bold mb-2">Glide Data Grid 完整使用指南</h2>
        <p className="text-gray-600 mb-4">
          这是一个展示 Glide Data Grid 各种功能的完整示例
        </p>
        
        <div className="mb-4 p-4 bg-gray-100 rounded-lg">
          <h3 className="font-semibold mb-2">功能特性：</h3>
          <ul className="list-disc list-inside space-y-1 text-sm">
            <li>多种单元格类型：文本、数字、布尔值、URI</li>
            <li>可编辑单元格</li>
            <li>复制粘贴支持</li>
            <li>行标记</li>
            <li>列宽调整</li>
            <li>选择处理</li>
          </ul>
        </div>
        
        {selectedCells.length > 0 && (
          <div className="mb-4 p-2 bg-blue-100 rounded">
            <p className="text-sm">
              已选择 {selectedCells.length} 个单元格
            </p>
          </div>
        )}
      </div>

      <div className="border rounded-lg overflow-hidden" style={{ height: '600px' }}>
        <DataEditor
          // 必需属性
          getCellContent={getCellContent}
          columns={columns}
          rows={data.length}
          
          // 编辑功能
          onCellEdited={onCellEdited}
          
          // 选择功能
          onGridSelectionChange={onSelectionChange}
          
          // 复制粘贴
          getCellsForSelection={getCellsForSelection}
          onPaste={onPaste}
          
          // 行标记
          rowMarkers={rowMarkers}
          
          // 样式配置
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
            rows: CompactSelection.empty(),
          }}
        />
      </div>
    </div>
  );
};

