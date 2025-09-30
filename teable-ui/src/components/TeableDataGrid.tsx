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
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';

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
  const [showAddField, setShowAddField] = useState(false);
  const [showAddRecord, setShowAddRecord] = useState(false);
  const [newFieldName, setNewFieldName] = useState('');
  const [newFieldType, setNewFieldType] = useState('singleLineText');
  const [creating, setCreating] = useState(false);

  // 列宽度状态管理
  const [columnWidths, setColumnWidths] = useState<Record<string, number>>({});
  // 列顺序状态管理
  const [columnOrder, setColumnOrder] = useState<string[]>([]);

  // 加载表格结构和数据
  useEffect(() => {
    const loadTableData = async () => {
      if (!tableId || !baseId) return;
      
      setLoading(true);
      try {
        // 1. 获取表格结构 + 字段 + 记录
        const [tableResp, fieldsResp, recordsResp] = await Promise.all([
          teable.getTable({ table_id: tableId }),
          teable.listFields({ table_id: tableId, limit: 200 }),
          teable.listRecords({ table_id: tableId, limit: 1000 })
        ]);
        const fieldsArr = Array.isArray(fieldsResp?.data) ? fieldsResp.data : [];
        const tbl = tableResp?.data || { id: tableId, name: '' };
        setTable({ ...(tbl as any), fields: fieldsArr });
        setRecords(recordsResp.data);

        // 初始化列宽度和顺序
        const initialColumnWidths: Record<string, number> = {};
        const initialColumnOrder: string[] = [];
        fieldsArr.forEach(field => {
          initialColumnWidths[field.id] = getColumnWidth(field.type);
          initialColumnOrder.push(field.id);
        });
        setColumnWidths(initialColumnWidths);
        setColumnOrder(initialColumnOrder);
        
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

  // 当字段列表变化时，更新列宽度和顺序
  useEffect(() => {
    if (table?.fields) {
      const newColumnWidths = { ...columnWidths };
      const newColumnOrder: string[] = [];
      const existingFieldIds = new Set(columnOrder);

      // 保持现有字段的顺序
      columnOrder.forEach(fieldId => {
        if (table.fields.some(f => f.id === fieldId)) {
          newColumnOrder.push(fieldId);
        }
      });

      // 添加新字段到末尾
      table.fields.forEach(field => {
        if (!existingFieldIds.has(field.id)) {
          newColumnWidths[field.id] = getColumnWidth(field.type);
          newColumnOrder.push(field.id);
        }
      });

      setColumnWidths(newColumnWidths);
      setColumnOrder(newColumnOrder);
    }
  }, [table?.fields]); // 依赖于table.fields的变化

  // 根据字段类型确定列宽（使用函数声明以避免 TDZ 错误）
  function getColumnWidth(fieldType: string): number {
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
  }

  // 根据 Teable 字段类型转换为 Glide 列
  const columns: GridColumn[] = useMemo(() => {
    if (!table || !columnOrder.length) return [];

    const fieldsMap = new Map(table.fields.map(field => [field.id, field]));

    return columnOrder.map((fieldId) => {
      const field = fieldsMap.get(fieldId);
      if (!field) {
        return { id: fieldId, title: 'Unknown', width: 100, resizable: true, movable: true };
      }
      return {
        id: field.id,
        title: field.name,
        width: columnWidths[field.id] || getColumnWidth(field.type),
        resizable: true,
        movable: true,
      };
    });
  }, [table, columnOrder, columnWidths]);

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
          allowOverlay: true,
        } as any;

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

      case 'currency':
      case 'percent':
        return {
          kind: GridCellKind.Number,
          data: value || 0,
          displayData: value?.toString() || '0',
          ...baseCell,
        };

      case 'longText':
        return {
          kind: GridCellKind.Text,
          data: value || '',
          displayData: value || '',
          ...baseCell,
        };

      case 'autoNumber':
        return {
          kind: GridCellKind.Number,
          data: value || 0,
          displayData: value?.toString() || '0',
          allowOverlay: false, // AutoNumber is not editable
        } as any;

      case 'formula':
        return {
          kind: GridCellKind.Text,
          data: value || '',
          displayData: value || '',
          allowOverlay: false, // Formula fields are not editable
        } as any;

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
    const column = columns[col];
    
    if (!record || !column) {
      return {
        kind: GridCellKind.Loading,
        allowOverlay: false,
      };
    }

    // 使用 column.id 来获取字段数据
    const field = table?.fields.find(f => f.id === column.id);
    if (!field) {
      return {
        kind: GridCellKind.Text,
        data: '',
        displayData: '',
        allowOverlay: true,
      };
    }

    const value = record[field.id];
    return createCell(field, value);
  }, [records, columns, table?.fields]);

  // 单元格校验（使用官方 validateCell 回调）
  const validateCell = useCallback((cell: Item, newValue: EditableGridCell) => {
    const [col, row] = cell;
    const column = columns[col];
    if (!column) return true;

    // 使用 column.id 来获取字段数据
    const field = table?.fields.find(f => f.id === column.id);
    if (!field) return true;

    // 简单类型前端校验；其余交给后端更新失败再提示
    switch (field.type) {
      case 'number':
      case 'currency':
      case 'percent': {
        const v = (newValue as any).data;
        const ok = typeof v === 'number' && !Number.isNaN(v);
        return ok;
      }
      case 'checkbox':
      case 'boolean':
        return (newValue as any).kind === GridCellKind.Boolean;
      case 'email':
        if (newValue.kind === GridCellKind.Text || newValue.kind === GridCellKind.Uri) {
          const email = (newValue as any).data;
          const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
          return emailRegex.test(email) || email === '';
        }
        return false;
      case 'url':
        if (newValue.kind === GridCellKind.Text || newValue.kind === GridCellKind.Uri) {
          const url = (newValue as any).data;
          if (url === '') return true;
          try {
            new URL(url);
            return true;
          } catch {
            return false;
          }
        }
        return false;
      case 'autoNumber':
      case 'formula':
        // 这些字段不应该被编辑
        return false;
      default:
        return true;
    }
  }, [columns, table?.fields]);

  // 单元格编辑处理
  const onCellEdited = useCallback(async (cell: Item, newValue: EditableGridCell) => {
    const [col, row] = cell;
    const record = records[row];
    const column = columns[col];
    
    if (!record || !column) return;

    // 使用 column.id 来获取字段数据
    const field = table?.fields.find(f => f.id === column.id);
    if (!field) return;

    try {
      let newData: any;
      
      // 根据字段类型处理新值
      switch (field.type) {
        case 'singleLineText':
        case 'longText':
        case 'email':
        case 'url':
        case 'singleSelect':
        case 'multipleSelect':
          if (newValue.kind === GridCellKind.Text || newValue.kind === GridCellKind.Uri) {
            newData = newValue.data;
          }
          break;
        
        case 'number':
        case 'currency':
        case 'percent':
          if (newValue.kind === GridCellKind.Number) {
            newData = newValue.data;
          }
          break;
        
        case 'checkbox':
        case 'boolean':
          if (newValue.kind === GridCellKind.Boolean) {
            newData = newValue.data;
          }
          break;
        
        case 'date':
        case 'dateTime':
          if (newValue.kind === GridCellKind.Text) {
            try {
              newData = new Date(newValue.data).toISOString();
            } catch {
              newData = newValue.data; // 如果日期解析失败，保持原始值
            }
          }
          break;
        
        case 'autoNumber':
        case 'formula':
          // 这些字段不应该被编辑，但为了安全起见
          return;
        
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
  }, [records, columns, table?.fields, tableId, onRecordEdit, toast]);

  // 行选择处理
  const onRowSelectionChange = useCallback((newSelection: CompactSelection) => {
    setSelectedRows(newSelection);
    
    // 触发选择回调
    if (newSelection.length > 0) {
      const firstIndex = Array.from(newSelection)[0];
      const selectedRecord = records[firstIndex];
      onRecordSelect?.(selectedRecord);
    }
  }, [records, onRecordSelect]);

  // 添加新记录（弹窗）
  const addNewRecord = useCallback(async () => {
    setShowAddRecord(true);
  }, []);

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
      
      setRecords(prev => prev.filter((_, index) => Array.from(selectedRows).indexOf(index) === -1));
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

  // 列大小调整处理
  const onColumnResize = useCallback((column: GridColumn, newSize: number) => {
    setColumnWidths(prev => ({
      ...prev,
      [column.id]: newSize,
    }));
  }, []);

  // 列移动处理
  const onColumnMoved = useCallback((startIndex: number, endIndex: number) => {
    setColumnOrder(prev => {
      const newOrder = [...prev];
      const [movedColumnId] = newOrder.splice(startIndex, 1);
      newOrder.splice(endIndex, 0, movedColumnId);
      return newOrder;
    });
  }, []);

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
            <h3 className="text-lg font-semibold">{table?.name || ''}</h3>
            <p className="text-sm text-gray-600">
              {records.length} 条记录，{table?.fields?.length ?? 0} 个字段
            </p>
            <p className="text-xs text-blue-600 mt-1">
              💡 双击单元格编辑 | 拖拽列边界调整宽度 | 拖拽列标题重新排序
            </p>
          </div>
          
          <div className="flex items-center space-x-2">
            <Button onClick={() => setShowAddField(true)}>添加字段</Button>
            <Button onClick={addNewRecord}>添加记录</Button>
            {selectedRows.length > 0 && (
              <Button variant="destructive" onClick={deleteSelectedRecords}>删除选中 ({selectedRows.length})</Button>
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
          validateCell={validateCell}
          
          // 编辑功能
          onCellEdited={onCellEdited}
          fillHandle={true}
          rowMarkers="both"
          headerHeight={36}
          rowHeight={36}
          trailingRowOptions={{ hint: '添加记录', addIcon: 'plus', sticky: true }}
          onRowAppended={async () => {
            try {
              const resp = await teable.createRecord({ table_id: tableId, fields: {} });
              setRecords(prev => [...prev, resp.data]);
              toast({ title: '记录已创建' });
              return 0;
            } catch (e: any) {
              toast({ title: '创建记录失败', description: e?.message || String(e), variant: 'destructive' });
              return 0;
            }
          }}
          
          // 选择（受控）
          onGridSelectionChange={(sel) => {
            if (sel?.rows) {
              setSelectedRows(sel.rows);
            } else {
              setSelectedRows(CompactSelection.empty());
            }
          }}
          
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
          
          // 列大小调整和移动
          onColumnResize={onColumnResize}
          onColumnMoved={onColumnMoved}
        />
      </div>

      {/* 添加字段弹窗 */}
      <Dialog open={showAddField} onOpenChange={setShowAddField}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>添加字段</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            <div className="space-y-1">
              <div className="text-sm">字段名</div>
              <input className="w-full px-3 py-2 border rounded" value={newFieldName} onChange={(e) => setNewFieldName(e.target.value)} placeholder="例如：姓名" />
            </div>
            <div className="space-y-1">
              <div className="text-sm">字段类型</div>
              <select className="w-full px-3 py-2 border rounded" value={newFieldType} onChange={(e) => setNewFieldType(e.target.value)}>
                <option value="singleLineText">单行文本</option>
                <option value="number">数字</option>
                <option value="checkbox">复选框</option>
                <option value="date">日期</option>
                <option value="singleSelect">单选</option>
                <option value="multipleSelect">多选</option>
              </select>
            </div>
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setShowAddField(false)}>取消</Button>
            <Button disabled={creating || !newFieldName.trim()} onClick={async () => {
              setCreating(true);
              try {
                const resp = await teable.createField({ table_id: tableId, name: newFieldName.trim(), type: newFieldType });
                setTable(prev => prev ? { ...prev, fields: [...(Array.isArray(prev.fields) ? prev.fields : []), resp.data as any] } : { id: tableId, name: '', fields: [resp.data as any] } as any);
                setShowAddField(false);
                setNewFieldName('');
                setNewFieldType('singleLineText');
                toast({ title: '字段已创建' });
              } catch (e: any) {
                toast({ title: '创建字段失败', description: e?.message || String(e), variant: 'destructive' });
              } finally {
                setCreating(false);
              }
            }}>创建</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 添加记录弹窗 */}
      <Dialog open={showAddRecord} onOpenChange={setShowAddRecord}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>添加记录</DialogTitle>
          </DialogHeader>
          <div className="text-sm text-gray-600">将创建一条空记录，之后可在网格中编辑。</div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setShowAddRecord(false)}>取消</Button>
            <Button disabled={creating} onClick={async () => {
              setCreating(true);
              try {
                const resp = await teable.createRecord({ table_id: tableId, fields: {} });
                setRecords(prev => [...prev, resp.data]);
                setShowAddRecord(false);
                toast({ title: '记录已创建' });
              } catch (e: any) {
                toast({ title: '创建记录失败', description: e?.message || String(e), variant: 'destructive' });
              } finally {
                setCreating(false);
              }
            }}>创建</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      
      {/* Portal容器 - 用于编辑覆盖层 */}
      <div id="portal"></div>
    </div>
  );
};
