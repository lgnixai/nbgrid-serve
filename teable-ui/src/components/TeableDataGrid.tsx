import React, { useState, useCallback, useMemo, useEffect } from 'react';
import { 
  DataEditor, 
  GridCellKind, 
  GridColumn, 
  Item, 
  GridCell, 
  EditableGridCell,
  CompactSelection
} from '@glideapps/glide-data-grid';
import '@glideapps/glide-data-grid/dist/index.css';
import teable from '@/lib/teable-simple';
import { useToast } from '@/hooks/use-toast';

// Teable 数据类型
interface TeableRecord {
  id: string;
  [key: string]: any;
}

interface TeableField {
  id: string;
  name: string;
  type: string;
  options?: any;
}

interface TeableTable {
  id: string;
  name: string;
  fields: TeableField[];
}

interface TeableDataGridProps {
  tableId: string;
  baseId: string;
  onRecordSelect?: (record: TeableRecord) => void;
  onRecordEdit?: (record: TeableRecord) => void;
  onRecordDelete?: (recordId: string) => void;
}

export const TeableDataGrid: React.FC<TeableDataGridProps> = ({
  tableId,
  baseId,
  onRecordSelect,
  onRecordEdit,
  onRecordDelete
}) => {
  const [records, setRecords] = useState<TeableRecord[]>([]);
  const [table, setTable] = useState<TeableTable | null>(null);
  const [loading, setLoading] = useState(false);
  const [selectedRows, setSelectedRows] = useState<CompactSelection>(CompactSelection.empty());
  const { toast } = useToast();

  // 加载表格结构和数据
  useEffect(() => {
    const loadTableData = async () => {
      if (!tableId || !baseId) return;
      
      setLoading(true);
      try {
        // 1. 获取表格结构
        const tableResp = await teable.getTable({ table_id: tableId });
        setTable(tableResp.data);
        
        // 2. 获取记录数据
        const recordsResp = await teable.listRecords({ 
          table_id: tableId,
          limit: 1000 // 根据需要调整
        });
        setRecords(recordsResp.data);
        
      } catch (error: any) {
        toast({
          title: "数据加载失败",
          description: error?.message || "无法加载表格数据",
          variant: "destructive"
        });
      } finally {
        setLoading(false);
      }
    };

    loadTableData();
  }, [tableId, baseId, toast]);

  // 根据 Teable 字段类型转换为 Glide 列
  const columns: GridColumn[] = useMemo(() => {
    if (!table) return [];
    
    return table.fields.map(field => ({
      title: field.name,
      id: field.id,
      width: getColumnWidth(field.type),
      resizable: true,
    }));
  }, [table]);

  // 根据字段类型确定列宽
  const getColumnWidth = (fieldType: string): number => {
    switch (fieldType) {
      case 'singleLineText':
      case 'longText':
        return 200;
      case 'number':
        return 120;
      case 'date':
      case 'dateTime':
        return 150;
      case 'singleSelect':
      case 'multipleSelect':
        return 150;
      case 'checkbox':
        return 80;
      case 'email':
      case 'url':
        return 200;
      default:
        return 150;
    }
  };

  // 根据字段类型创建单元格
  const createCell = (field: TeableField, value: any): GridCell => {
    const baseCell = {
      allowOverlay: true,
    };

    switch (field.type) {
      case 'singleLineText':
      case 'longText':
        return {
          kind: GridCellKind.Text,
          data: value || '',
          displayData: value || '',
          ...baseCell,
        };

      case 'number':
        return {
          kind: GridCellKind.Number,
          data: value || 0,
          displayData: value?.toString() || '0',
          ...baseCell,
        };

      case 'checkbox':
        return {
          kind: GridCellKind.Boolean,
          data: Boolean(value),
          displayData: value ? '是' : '否',
          ...baseCell,
        };

      case 'email':
        return {
          kind: GridCellKind.Uri,
          data: value || '',
          displayData: value || '',
          ...baseCell,
        };

      case 'url':
        return {
          kind: GridCellKind.Uri,
          data: value || '',
          displayData: value || '',
          ...baseCell,
        };

      case 'date':
      case 'dateTime':
        return {
          kind: GridCellKind.Text,
          data: value ? new Date(value).toLocaleDateString() : '',
          displayData: value ? new Date(value).toLocaleDateString() : '',
          ...baseCell,
        };

      case 'singleSelect':
        return {
          kind: GridCellKind.Text,
          data: value || '',
          displayData: value || '',
          ...baseCell,
        };

      case 'multipleSelect':
        return {
          kind: GridCellKind.Text,
          data: Array.isArray(value) ? value.join(', ') : value || '',
          displayData: Array.isArray(value) ? value.join(', ') : value || '',
          ...baseCell,
        };

      default:
        return {
          kind: GridCellKind.Text,
          data: value || '',
          displayData: value || '',
          ...baseCell,
        };
    }
  };

  // 获取单元格内容
  const getCellContent = useCallback((cell: Item): GridCell => {
    const [col, row] = cell;
    const record = records[row];
    const field = table?.fields[col];
    
    if (!record || !field) {
      return {
        kind: GridCellKind.Loading,
        allowOverlay: false,
      };
    }

    const value = record[field.id];
    return createCell(field, value);
  }, [records, table]);

  // 单元格编辑处理
  const onCellEdited = useCallback(async (cell: Item, newValue: EditableGridCell) => {
    const [col, row] = cell;
    const record = records[row];
    const field = table?.fields[col];
    
    if (!record || !field) return;

    try {
      let newData: any;
      
      // 根据字段类型处理新值
      switch (field.type) {
        case 'singleLineText':
        case 'longText':
        case 'email':
        case 'url':
          if (newValue.kind === GridCellKind.Text || newValue.kind === GridCellKind.Uri) {
            newData = newValue.data;
          }
          break;
        
        case 'number':
          if (newValue.kind === GridCellKind.Number) {
            newData = newValue.data;
          }
          break;
        
        case 'checkbox':
          if (newValue.kind === GridCellKind.Boolean) {
            newData = newValue.data;
          }
          break;
        
        case 'date':
        case 'dateTime':
          if (newValue.kind === GridCellKind.Text) {
            newData = new Date(newValue.data).toISOString();
          }
          break;
        
        default:
          if (newValue.kind === GridCellKind.Text) {
            newData = newValue.data;
          }
          break;
      }

      // 更新本地状态
      setRecords(prev => prev.map((r, index) => 
        index === row ? { ...r, [field.id]: newData } : r
      ));

      // 调用 Teable API 更新记录
      await teable.updateRecord({
        table_id: tableId,
        record_id: record.id,
        fields: { [field.id]: newData }
      });

      // 触发编辑回调
      onRecordEdit?.(record);

      toast({
        title: "更新成功",
        description: "记录已成功更新",
      });

    } catch (error: any) {
      toast({
        title: "更新失败",
        description: error?.message || "无法更新记录",
        variant: "destructive"
      });
      
      // 恢复原始数据
      setRecords(prev => [...prev]);
    }
  }, [records, table, tableId, onRecordEdit, toast]);

  // 行选择处理
  const onRowSelectionChange = useCallback((newSelection: CompactSelection) => {
    setSelectedRows(newSelection);
    
    // 触发选择回调
    if (newSelection.length > 0) {
      const selectedRecord = records[Array.from(newSelection)[0]];
      onRecordSelect?.(selectedRecord);
    }
  }, [records, onRecordSelect]);

  // 添加新记录
  const addNewRecord = useCallback(async () => {
    if (!table) return;
    
    try {
      const newRecord = await teable.createRecord({
        table_id: tableId,
        fields: {}
      });
      
      setRecords(prev => [...prev, newRecord.data]);
      
      toast({
        title: "记录已创建",
        description: "新记录已成功创建",
      });
    } catch (error: any) {
      toast({
        title: "创建失败",
        description: error?.message || "无法创建新记录",
        variant: "destructive"
      });
    }
  }, [table, tableId, toast]);

  // 删除选中记录
  const deleteSelectedRecords = useCallback(async () => {
    if (selectedRows.length === 0) return;
    
    try {
      const recordIds = Array.from(selectedRows).map(index => records[index].id);
      
      for (const recordId of recordIds) {
        await teable.deleteRecord({
          table_id: tableId,
          record_id: recordId
        });
        
        onRecordDelete?.(recordId);
      }
      
      setRecords(prev => prev.filter((_, index) => !selectedRows.has(index)));
      setSelectedRows(CompactSelection.empty());
      
      toast({
        title: "删除成功",
        description: `已删除 ${recordIds.length} 条记录`,
      });
    } catch (error: any) {
      toast({
        title: "删除失败",
        description: error?.message || "无法删除记录",
        variant: "destructive"
      });
    }
  }, [selectedRows, records, tableId, onRecordDelete, toast]);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-lg">加载中...</div>
      </div>
    );
  }

  if (!table) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-lg text-gray-500">未找到表格数据</div>
      </div>
    );
  }

  return (
    <div className="w-full h-full">
      {/* 工具栏 */}
      <div className="mb-4 p-4 bg-gray-100 rounded-lg">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-lg font-semibold">{table.name}</h3>
            <p className="text-sm text-gray-600">
              {records.length} 条记录，{table.fields.length} 个字段
            </p>
          </div>
          
          <div className="flex items-center space-x-2">
            <button
              onClick={addNewRecord}
              className="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600"
            >
              添加记录
            </button>
            
            {selectedRows.length > 0 && (
              <button
                onClick={deleteSelectedRecords}
                className="px-4 py-2 bg-red-500 text-white rounded-md hover:bg-red-600"
              >
                删除选中 ({selectedRows.length})
              </button>
            )}
          </div>
        </div>
      </div>

      {/* 数据网格 */}
      <div className="border rounded-lg overflow-hidden" style={{ height: '600px' }}>
        <DataEditor
          getCellContent={getCellContent}
          columns={columns}
          rows={records.length}
          
          // 编辑功能
          onCellEdited={onCellEdited}
          
          // 行选择
          rowSelect="multi"
          onRowSelectionChange={onRowSelectionChange}
          
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
