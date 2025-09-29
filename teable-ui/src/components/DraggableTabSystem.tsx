import { useState } from "react";
import { X, Plus, MoreHorizontal } from "lucide-react";
import { cn } from "@/lib/utils";
import { ContextMenu, ContextMenuContent, ContextMenuItem, ContextMenuTrigger } from "@/components/ui/context-menu";
import { Button } from "@/components/ui/button";
import { MarkdownEditor } from "./MarkdownEditor";
import { AdvancedGlideGrid } from "./AdvancedGlideGrid";
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  DragEndEvent,
} from '@dnd-kit/core';
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  horizontalListSortingStrategy,
} from '@dnd-kit/sortable';
import {
  useSortable,
} from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';

interface Tab {
  id: string;
  title: string;
  content: string;
  type: "markdown";
}

interface DraggableTabSystemProps {
  tabs: Tab[];
  activeTab: string;
  onTabChange: (tabId: string) => void;
  onTabClose: (tabId: string) => void;
  onTabSplit: (tabId: string, direction: "horizontal" | "vertical") => void;
  onTabsReorder: (tabs: Tab[]) => void;
  onTabAdd: () => void;
  onContentChange: (tabId: string, content: string) => void;
}

interface SortableTabProps {
  tab: Tab;
  isActive: boolean;
  onTabChange: (tabId: string) => void;
  onTabClose: (tabId: string) => void;
  onTabSplit: (tabId: string, direction: "horizontal" | "vertical") => void;
}

const SortableTab = ({ tab, isActive, onTabChange, onTabClose, onTabSplit }: SortableTabProps) => {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: tab.id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  const handleTabAction = (action: string, tabId: string) => {
    switch (action) {
      case "close":
        onTabClose(tabId);
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

  return (
    <ContextMenu>
      <ContextMenuTrigger>
        <div
          ref={setNodeRef}
          style={style}
          {...attributes}
          {...listeners}
          className={cn(
            "flex items-center px-4 py-2 text-sm border-r border-obsidian-border cursor-pointer transition-colors min-w-0 group select-none",
            isActive
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
      </ContextMenuTrigger>
      <ContextMenuContent className="w-48 bg-obsidian-surface border-obsidian-border">
        <ContextMenuItem
          className="text-obsidian-text hover:bg-obsidian-surface-hover focus:bg-obsidian-surface-hover"
          onClick={() => handleTabAction("close", tab.id)}
        >
          关闭
        </ContextMenuItem>
        <ContextMenuItem
          className="text-obsidian-text hover:bg-obsidian-surface-hover focus:bg-obsidian-surface-hover"
          onClick={() => handleTabAction("split-horizontal", tab.id)}
        >
          左右分屏
        </ContextMenuItem>
        <ContextMenuItem
          className="text-obsidian-text hover:bg-obsidian-surface-hover focus:bg-obsidian-surface-hover"
          onClick={() => handleTabAction("split-vertical", tab.id)}
        >
          上下分屏
        </ContextMenuItem>
      </ContextMenuContent>
    </ContextMenu>
  );
};

export const DraggableTabSystem = ({ 
  tabs, 
  activeTab, 
  onTabChange, 
  onTabClose, 
  onTabSplit, 
  onTabsReorder,
  onTabAdd,
  onContentChange
}: DraggableTabSystemProps) => {
  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  );

  const activeTabData = tabs.find(tab => tab.id === activeTab);

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;

    if (over && active.id !== over.id) {
      const oldIndex = tabs.findIndex(tab => tab.id === active.id);
      const newIndex = tabs.findIndex(tab => tab.id === over.id);
      const newTabs = arrayMove(tabs, oldIndex, newIndex);
      onTabsReorder(newTabs);
    }
  };

  return (
    <div className="h-full bg-obsidian-bg flex flex-col">
      {/* Tab Bar */}
      <div className="flex items-center border-b border-obsidian-border bg-obsidian-surface">
        <DndContext
          sensors={sensors}
          collisionDetection={closestCenter}
          onDragEnd={handleDragEnd}
        >
          <div className="flex overflow-x-auto scrollbar-hide">
            <SortableContext items={tabs.map(tab => tab.id)} strategy={horizontalListSortingStrategy}>
              {tabs.map((tab) => (
                <SortableTab
                  key={tab.id}
                  tab={tab}
                  isActive={activeTab === tab.id}
                  onTabChange={onTabChange}
                  onTabClose={onTabClose}
                  onTabSplit={onTabSplit}
                />
              ))}
            </SortableContext>
          </div>
        </DndContext>
        
        <Button
          variant="ghost"
          size="sm"
          className="ml-auto mr-2 text-obsidian-text-muted hover:text-obsidian-text"
          onClick={onTabAdd}
        >
          <Plus className="w-4 h-4" />
        </Button>
      </div>

      {/* Tab Content */}
      <div className="flex-1 overflow-hidden">
        {activeTabData && (
          // <MarkdownEditor
          //   key={activeTabData.id}
          //   content={activeTabData.content}
          //   title={activeTabData.title}
          //   onChange={(content) => onContentChange(activeTabData.id, content)}
          // />
          <AdvancedGlideGrid />
        )}
      </div>
    </div>
  );
};