import { useState, useRef } from "react";
import { X, Plus, MoreHorizontal } from "lucide-react";
import { cn } from "@/lib/utils";
import { ContextMenu, ContextMenuContent, ContextMenuItem, ContextMenuTrigger } from "@/components/ui/context-menu";
import { Button } from "@/components/ui/button";
import { MarkdownEditor } from "./MarkdownEditor";

interface Tab {
  id: string;
  title: string;
  content: string;
  type: "markdown";
}

interface TabSystemProps {
  tabs: Tab[];
  activeTab: string;
  onTabChange: (tabId: string) => void;
  onTabClose: (tabId: string) => void;
  onTabSplit: (tabId: string, direction: "horizontal" | "vertical") => void;
}

export const TabSystem = ({ tabs, activeTab, onTabChange, onTabClose, onTabSplit }: TabSystemProps) => {
  const [contextMenuTab, setContextMenuTab] = useState<string | null>(null);
  const activeTabData = tabs.find(tab => tab.id === activeTab);

  const handleContextMenu = (tabId: string) => {
    setContextMenuTab(tabId);
  };

  const handleTabAction = (action: string, tabId: string) => {
    switch (action) {
      case "close":
        onTabClose(tabId);
        break;
      case "close-others":
        tabs.forEach(tab => {
          if (tab.id !== tabId) {
            onTabClose(tab.id);
          }
        });
        break;
      case "close-all":
        tabs.forEach(tab => onTabClose(tab.id));
        break;
      case "split-horizontal":
        onTabSplit(tabId, "horizontal");
        break;
      case "split-vertical":
        onTabSplit(tabId, "vertical");
        break;
      default:
        console.log("Action:", action, "for tab:", tabId);
    }
  };

  const TabContextMenu = ({ tabId, children }: { tabId: string; children: React.ReactNode }) => (
    <ContextMenu>
      <ContextMenuTrigger onContextMenu={() => handleContextMenu(tabId)}>
        {children}
      </ContextMenuTrigger>
      <ContextMenuContent className="w-48 bg-obsidian-surface border-obsidian-border">
        <ContextMenuItem
          className="text-obsidian-text hover:bg-obsidian-surface-hover focus:bg-obsidian-surface-hover"
          onClick={() => handleTabAction("close", tabId)}
        >
          关闭
        </ContextMenuItem>
        <ContextMenuItem
          className="text-obsidian-text hover:bg-obsidian-surface-hover focus:bg-obsidian-surface-hover"
          onClick={() => handleTabAction("close-others", tabId)}
        >
          关闭其他标签页
        </ContextMenuItem>
        <ContextMenuItem
          className="text-obsidian-text hover:bg-obsidian-surface-hover focus:bg-obsidian-surface-hover"
          onClick={() => handleTabAction("close-all", tabId)}
        >
          全部关闭
        </ContextMenuItem>
        <ContextMenuItem
          className="text-obsidian-text hover:bg-obsidian-surface-hover focus:bg-obsidian-surface-hover"
          onClick={() => handleTabAction("pin", tabId)}
        >
          锁定
        </ContextMenuItem>
        <ContextMenuItem
          className="text-obsidian-text hover:bg-obsidian-surface-hover focus:bg-obsidian-surface-hover"
          onClick={() => handleTabAction("link", tabId)}
        >
          关联标签页......
        </ContextMenuItem>
        <ContextMenuItem
          className="text-obsidian-text hover:bg-obsidian-surface-hover focus:bg-obsidian-surface-hover"
          onClick={() => handleTabAction("move-window", tabId)}
        >
          移动至新窗口
        </ContextMenuItem>
        <ContextMenuItem
          className="text-obsidian-text hover:bg-obsidian-surface-hover focus:bg-obsidian-surface-hover"
          onClick={() => handleTabAction("split-horizontal", tabId)}
        >
          左右分屏
        </ContextMenuItem>
        <ContextMenuItem
          className="text-obsidian-text hover:bg-obsidian-surface-hover focus:bg-obsidian-surface-hover"
          onClick={() => handleTabAction("split-vertical", tabId)}
        >
          上下分屏
        </ContextMenuItem>
      </ContextMenuContent>
    </ContextMenu>
  );

  return (
    <div className="h-full bg-obsidian-bg flex flex-col">
      {/* Tab Bar */}
      <div className="flex items-center border-b border-obsidian-border bg-obsidian-surface">
        <div className="flex overflow-x-auto scrollbar-hide">
          {tabs.map((tab) => (
            <TabContextMenu key={tab.id} tabId={tab.id}>
              <div
                className={cn(
                  "flex items-center px-4 py-2 text-sm border-r border-obsidian-border cursor-pointer transition-colors min-w-0 group",
                  activeTab === tab.id
                    ? "bg-tab-active text-obsidian-text"
                    : "bg-tab-inactive text-obsidian-text-muted hover:bg-tab-hover"
                )}
                onClick={() => onTabChange(tab.id)}
              >
                <span className="truncate max-w-32">{tab.title}</span>
                <Button
                  variant="ghost"
                  size="sm"
                  className="ml-2 p-0 w-4 h-4 text-obsidian-text-muted hover:text-obsidian-text opacity-0 group-hover:opacity-100 transition-opacity"
                  onClick={(e) => {
                    e.stopPropagation();
                    onTabClose(tab.id);
                  }}
                >
                  <X className="w-3 h-3" />
                </Button>
              </div>
            </TabContextMenu>
          ))}
        </div>
        
        <Button
          variant="ghost"
          size="sm"
          className="ml-auto mr-2 text-obsidian-text-muted hover:text-obsidian-text"
          onClick={() => {
            // Add new tab functionality - generate unique ID
            const newTab = {
              id: `tab-${Date.now()}`,
              title: "新标签页",
              content: "# 新文档\n\n开始编写...",
              type: "markdown" as const
            };
            console.log("Add new tab", newTab);
          }}
        >
          <Plus className="w-4 h-4" />
        </Button>
      </div>

      {/* Tab Content */}
      <div className="flex-1 overflow-hidden">
        {activeTabData && (
          <MarkdownEditor
            key={activeTabData.id}
            content={activeTabData.content}
            onChange={(content) => {
              // Update tab content
              const updatedTabs = tabs.map(tab => 
                tab.id === activeTabData.id 
                  ? { ...tab, content }
                  : tab
              );
              // In a real app, you'd update the parent state here
              console.log("Content changed:", content);
            }}
          />
        )}
      </div>
    </div>
  );
};