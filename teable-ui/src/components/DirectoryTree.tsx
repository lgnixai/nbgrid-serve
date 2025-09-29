import { useState, useEffect } from "react";
import { ChevronRight, ChevronDown, File, Folder, FolderOpen, Plus, Search, MoreHorizontal, Trash2 } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { FileCreateDialog } from "./FileCreateDialog";
import { SearchBar } from "./SearchBar";
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuTrigger,
} from "@/components/ui/context-menu";

interface FileItem {
  id: string;
  name: string;
  type: "file" | "folder";
  children?: FileItem[];
  isExpanded?: boolean;
}

interface DirectoryTreeProps {
  onFileOpen: (fileName: string) => void;
  onFileCreate?: (fileName: string, type: string) => void;
  onFileDelete?: (fileName: string) => void;
  items?: string[];
}

const mockFiles = [
  { id: "1", name: "欢迎.md", type: "file" as const },
  { id: "2", name: "项目笔记.md", type: "file" as const },
  { id: "3", name: "待办事项.md", type: "file" as const },
  { id: "4", name: "会议记录.md", type: "file" as const },
  { id: "5", name: "想法收集.md", type: "file" as const },
];

export const DirectoryTree = ({ onFileOpen, onFileCreate, onFileDelete, items }: DirectoryTreeProps) => {
  const [expandedFolders, setExpandedFolders] = useState<string[]>(["root"]);
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [showSearch, setShowSearch] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [files, setFiles] = useState(mockFiles);

  // 使用 useEffect 来响应 items 变化
  useEffect(() => {
    if (items && items.length > 0) {
      const external = items.map((name, idx) => ({ id: `ext-${idx}`, name, type: "file" as const }));
      setFiles(external);
    } else {
      setFiles(mockFiles);
    }
  }, [items]);

  const toggleFolder = (folderId: string) => {
    setExpandedFolders(prev =>
      prev.includes(folderId)
        ? prev.filter(id => id !== folderId)
        : [...prev, folderId]
    );
  };

  const handleCreateFile = (fileName: string, type: string) => {
    const newFile = {
      id: `file-${Date.now()}`,
      name: fileName,
      type: "file" as const
    };
    setFiles(prev => [...prev, newFile]);
    onFileCreate?.(fileName, type);
  };

  const handleDeleteFile = (fileName: string) => {
    setFiles(prev => prev.filter(file => file.name !== fileName));
    onFileDelete?.(fileName);
  };

  const filteredFiles = searchQuery 
    ? files.filter(file => 
        file.name.toLowerCase().includes(searchQuery.toLowerCase())
      )
    : files;

  return (
    <div className="h-full bg-obsidian-surface flex flex-col">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-obsidian-border">
        <h2 className="text-sm font-medium text-obsidian-text">表列表</h2>
        <div className="flex items-center gap-1">
          <Button 
            variant="ghost" 
            size="sm"
            onClick={() => setShowSearch(!showSearch)}
            className="text-obsidian-text-muted hover:text-obsidian-text h-6 w-6 p-0"
            title="搜索文件"
          >
            <Search className="w-4 h-4" />
          </Button>
          <Button 
            variant="ghost" 
            size="sm"
            onClick={() => setShowCreateDialog(true)}
            className="text-obsidian-text-muted hover:text-obsidian-text h-6 w-6 p-0"
            title="新建文件"
          >
            <Plus className="w-4 h-4" />
          </Button>
          <Button 
            variant="ghost" 
            size="sm"
            className="text-obsidian-text-muted hover:text-obsidian-text h-6 w-6 p-0"
          >
            <MoreHorizontal className="w-4 h-4" />
          </Button>
        </div>
      </div>

      {/* Search Bar */}
      {showSearch && (
        <SearchBar 
          onSearch={setSearchQuery}
          onClose={() => {
            setShowSearch(false);
            setSearchQuery("");
          }}
        />
      )}

      {/* File Tree */}
      <div className="flex-1 overflow-y-auto">
        <div className="p-2">
          {filteredFiles.map((file) => (
            <ContextMenu key={file.id}>
              <ContextMenuTrigger>
                <div className="mb-1">
                  <Button
                    variant="ghost"
                    className="w-full justify-start text-left p-2 h-auto text-obsidian-text hover:bg-obsidian-surface-hover"
                    onClick={() => onFileOpen(file.name)}
                  >
                    <File className="w-4 h-4 mr-2 text-obsidian-text-muted" />
                    <span className="text-sm">{file.name}</span>
                  </Button>
                </div>
              </ContextMenuTrigger>
              <ContextMenuContent className="bg-obsidian-surface border-obsidian-border">
                <ContextMenuItem
                  className="text-obsidian-text hover:bg-obsidian-surface-hover focus:bg-obsidian-surface-hover cursor-pointer"
                  onClick={() => onFileOpen(file.name)}
                >
                  <File className="w-4 h-4 mr-2" />
                  打开
                </ContextMenuItem>
                <ContextMenuItem
                  className="text-destructive hover:bg-obsidian-surface-hover focus:bg-obsidian-surface-hover cursor-pointer"
                  onClick={() => handleDeleteFile(file.name)}
                >
                  <Trash2 className="w-4 h-4 mr-2" />
                  删除
                </ContextMenuItem>
              </ContextMenuContent>
            </ContextMenu>
          ))}
        </div>
      </div>

      {/* Create File Dialog */}
      <FileCreateDialog
        open={showCreateDialog}
        onOpenChange={setShowCreateDialog}
        onCreateFile={handleCreateFile}
      />
    </div>
  );
};