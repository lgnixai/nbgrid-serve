import React, { useState, useCallback, useMemo } from 'react';
import { DataEditor, GridCellKind, GridColumn, Item, GridCell } from '@glideapps/glide-data-grid';
import '@glideapps/glide-data-grid/dist/index.css';

// 简单的数据类型
interface SimpleDataRow {
  id: string;
  name: string;
  age: number;
  city: string;
}

// 示例数据
const simpleData: SimpleDataRow[] = [
  { id: '1', name: '张三', age: 25, city: '北京' },
  { id: '2', name: '李四', age: 30, city: '上海' },
  { id: '3', name: '王五', age: 28, city: '广州' },
];

export const SimpleGlideGrid: React.FC = () => {
  const [data, setData] = useState<SimpleDataRow[]>(simpleData);

  // 列定义
  const columns: GridColumn[] = useMemo(() => [
    {
      title: 'ID',
      id: 'id',
      width: 80,
      resizable: true,
    },
    {
      title: '姓名',
      id: 'name',
      width: 120,
      resizable: true,
    },
    {
      title: '年龄',
      id: 'age',
      width: 100,
      resizable: true,
    },
    {
      title: '城市',
      id: 'city',
      width: 150,
      resizable: true,
    }
  ], []);

  // 获取单元格内容
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
      
      case 'age':
        return {
          kind: GridCellKind.Number,
          data: dataRow.age,
          displayData: dataRow.age.toString(),
          allowOverlay: true,
        };
      
      case 'city':
        return {
          kind: GridCellKind.Text,
          data: dataRow.city,
          displayData: dataRow.city,
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

  return (
    <div className="w-full h-full p-4">
      <div className="mb-4">
        <h2 className="text-2xl font-bold mb-2">简单 Glide Data Grid 测试</h2>
        <p className="text-gray-600 mb-4">
          这是一个简化的测试版本，用于验证基本功能
        </p>
      </div>

      <div className="border rounded-lg overflow-hidden" style={{ height: '400px' }}>
        <DataEditor
          getCellContent={getCellContent}
          columns={columns}
          rows={data.length}
          
          // 基础配置
          smoothScrollX={true}
          smoothScrollY={true}
          overscrollX={0}
          overscrollY={0}
          
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
        />
      </div>
    </div>
  );
};
