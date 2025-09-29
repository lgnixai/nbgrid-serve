import { Search } from "lucide-react";
import { Input } from "@/components/ui/input";

export const RightSidebar = () => {
  return (
    <div className="h-full bg-obsidian-surface border-l border-obsidian-border">
      <div className="p-4 border-b border-obsidian-border">
        <h2 className="text-sm font-medium text-obsidian-text mb-3">链接当前文件</h2>
        <div className="text-xs text-obsidian-text-muted mb-4">
          没有笔记链接当前文件
        </div>
        
        <div className="relative">
          <Search className="absolute left-2 top-2.5 w-4 h-4 text-obsidian-text-muted" />
          <Input
            placeholder="搜索当前文件名"
            className="pl-8 bg-obsidian-bg border-obsidian-border text-obsidian-text placeholder-obsidian-text-muted"
          />
        </div>
      </div>
      
      <div className="p-4">
        <div className="text-xs text-obsidian-text-muted">
          暂无链接内容
        </div>
      </div>
    </div>
  );
};