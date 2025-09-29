import { useState } from "react";
import { ResizablePanelGroup, ResizablePanel, ResizableHandle } from "@/components/ui/resizable";
import { DirectoryTree } from "./DirectoryTree";
import { DraggableTabSystem } from "./DraggableTabSystem";
import { RightSidebar } from "./RightSidebar";
import { ThemeSwitch } from "./ThemeSwitch";
import { SpaceBaseSelector, SpaceOption } from "./SpaceBaseSelector";
import { AuthStatus } from "./AuthStatus";
import teable from "@/lib/teable-simple";
import { useEffect } from "react";
import { useKeyboardShortcuts } from "@/hooks/useKeyboardShortcuts";
import { useToast } from "@/hooks/use-toast";
import { useAuth } from "@/hooks/useAuth";
import { Toaster } from "@/components/ui/toaster";

interface Tab {
  id: string;
  title: string;
  content: string;
  type: "markdown";
}

export const ObsidianLayout = () => {
  const [openTabs, setOpenTabs] = useState<Tab[]>([
    { id: "tab1", title: "新标签页", content: "# 欢迎使用 Obsidian 风格编辑器\n\n开始编写您的内容...", type: "markdown" },
  ]);
  const [activeTab, setActiveTab] = useState("tab1");
  const { toast } = useToast();
  const { isAuthenticated, error: authError } = useAuth();
  const [spaces, setSpaces] = useState<SpaceOption[]>([]);
  const [selectedSpaceId, setSelectedSpaceId] = useState<string>("");
  const [selectedBaseId, setSelectedBaseId] = useState<string>("");
  const [loadingData, setLoadingData] = useState(false);
  const [currentTables, setCurrentTables] = useState<string[]>([]);

  // 初始化时获取所有 spaces
  useEffect(() => {
    const loadSpaces = async () => {
      if (!isAuthenticated) return;
      
      setLoadingData(true);
      try {
        const spaceResp = await teable.listSpaces({ limit: 50 });
        const spaceItems = spaceResp.data;
        const combined: SpaceOption[] = spaceItems.map(s => ({
          id: s.id,
          name: s.name,
          bases: [], // 初始为空，等选择 space 后再加载
        }));
        setSpaces(combined);
        if (combined.length) {
          setSelectedSpaceId(combined[0].id);
        }
      } catch (e: any) {
        toast({ 
          title: "获取空间列表失败", 
          description: String(e?.message || e), 
          variant: "destructive" 
        });
      } finally {
        setLoadingData(false);
      }
    };

    loadSpaces();
  }, [isAuthenticated, toast]);

  // 当 selectedSpaceId 变化时，获取该 space 下的 bases
  useEffect(() => {
    const loadBases = async () => {
      if (!selectedSpaceId || !isAuthenticated) return;
      
      setLoadingData(true);
      try {
        const baseResp = await teable.listBases({ limit: 100 });
        // 过滤出当前 space 下的 bases
        const spaceBases = baseResp.data
          .filter(b => b.space_id === selectedSpaceId)
          .map(b => ({ id: b.id, name: b.name, tables: [] }));
        
        // 更新 spaces 中对应 space 的 bases
        setSpaces(prev => prev.map(space => 
          space.id === selectedSpaceId 
            ? { ...space, bases: spaceBases }
            : space
        ));
        
        // 默认选中第一个 base
        if (spaceBases.length > 0) {
          setSelectedBaseId(spaceBases[0].id);
        } else {
          setSelectedBaseId("");
          setCurrentTables([]);
        }
      } catch (e: any) {
        toast({ 
          title: "获取数据库列表失败", 
          description: String(e?.message || e), 
          variant: "destructive" 
        });
      } finally {
        setLoadingData(false);
      }
    };

    loadBases();
  }, [selectedSpaceId, isAuthenticated, toast]);

  // 当 selectedBaseId 变化时，获取该 base 下的 tables
  useEffect(() => {
    const loadTables = async () => {
      if (!selectedBaseId || !isAuthenticated) return;
      
      setLoadingData(true);
      try {
        const tablesResp = await teable.listTables({ base_id: selectedBaseId, limit: 200 });
        if (tablesResp.data.length > 0) {
          const tableNames = tablesResp.data.map(t => `${t.name}.md`);
          setCurrentTables(tableNames);
        } else {
          // 如果没有 tables，显示一些示例数据
          const spaceName = spaces.find(s => s.id === selectedSpaceId)?.name || "未命名空间";
          const baseName = spaces.find(s => s.id === selectedSpaceId)?.bases.find(b => b.id === selectedBaseId)?.name || "未命名数据库";
          setCurrentTables([
            `${baseName}表1.md`,
            `${baseName}表2.md`, 
            `${baseName}表3.md`
          ]);
        }
      } catch (e: any) {
        toast({ 
          title: "获取表格列表失败", 
          description: String(e?.message || e), 
          variant: "destructive" 
        });
        setCurrentTables([]);
      } finally {
        setLoadingData(false);
      }
    };

    loadTables();
  }, [selectedBaseId, isAuthenticated, toast, selectedSpaceId, spaces]);

  const handleTabClose = (tabId: string) => {
    const newTabs = openTabs.filter(tab => tab.id !== tabId);
    setOpenTabs(newTabs);
    
    if (activeTab === tabId && newTabs.length > 0) {
      setActiveTab(newTabs[0].id);
    }
  };

  const handleFileOpen = (fileName: string) => {
    // Check if file is already open
    const existingTab = openTabs.find(tab => tab.title === fileName);
    if (existingTab) {
      setActiveTab(existingTab.id);
      return;
    }

    const newTab: Tab = {
      id: `tab-${Date.now()}`,
      title: fileName,
      content: `# ${fileName}\n\n这是 ${fileName} 的内容...`,
      type: "markdown"
    };
    
    setOpenTabs([...openTabs, newTab]);
    setActiveTab(newTab.id);
  };

  const handleFileCreate = (fileName: string, type: string) => {
    const newTab: Tab = {
      id: `tab-${Date.now()}`,
      title: fileName,
      content: type === "markdown" ? `# ${fileName}\n\n` : "",
      type: "markdown"
    };
    
    setOpenTabs([...openTabs, newTab]);
    setActiveTab(newTab.id);
    toast({
      title: "文件已创建",
      description: `${fileName} 已成功创建`,
    });
  };

  const handleFileDelete = (fileName: string) => {
    // Close tab if open
    const tabToClose = openTabs.find(tab => tab.title === fileName);
    if (tabToClose) {
      handleTabClose(tabToClose.id);
    }
    
    toast({
      title: "文件已删除",
      description: `${fileName} 已被删除`,
      variant: "destructive",
    });
  };

  const handleNewFile = () => {
    const newTab: Tab = {
      id: `tab-${Date.now()}`,
      title: "新标签页",
      content: "# 新文档\n\n开始编写...",
      type: "markdown"
    };
    
    setOpenTabs([...openTabs, newTab]);
    setActiveTab(newTab.id);
  };

  const handleSave = () => {
    toast({
      title: "已保存",
      description: "文档已保存到本地",
    });
  };

  const handleTabAdd = () => {
    const newTab: Tab = {
      id: `tab-${Date.now()}`,
      title: "新标签页",
      content: "# 新文档\n\n开始编写...",
      type: "markdown"
    };
    
    setOpenTabs([...openTabs, newTab]);
    setActiveTab(newTab.id);
  };

  const handleTabsReorder = (newTabs: Tab[]) => {
    setOpenTabs(newTabs);
  };

  const handleContentChange = (tabId: string, content: string) => {
    setOpenTabs(tabs => 
      tabs.map(tab => 
        tab.id === tabId ? { ...tab, content } : tab
      )
    );
  };

  // Keyboard shortcuts
  useKeyboardShortcuts({
    onNewFile: handleNewFile,
    onSave: handleSave,
    onCloseTab: () => {
      if (activeTab) {
        handleTabClose(activeTab);
      }
    },
  });

  return (
    <div className="h-screen bg-obsidian-bg text-obsidian-text overflow-hidden">
      {/* Top Bar */}
      <div className="flex items-center justify-between px-4 py-2 border-b border-obsidian-border bg-obsidian-surface">
        <div className="flex items-center space-x-4">
          <h1 className="text-lg font-semibold text-obsidian-text">Nbgrid</h1>
          {isAuthenticated && !loadingData && (
            <SpaceBaseSelector
              spaces={spaces}
              spaceId={selectedSpaceId}
              baseId={selectedBaseId}
              onChange={(sid, bid) => {
                setSelectedSpaceId(sid);
                // 当 space 变化时，重置 base 选择，让 useEffect 来处理
                if (sid !== selectedSpaceId) {
                  setSelectedBaseId("");
                } else if (bid) {
                  setSelectedBaseId(bid);
                }
              }}
            />
          )}
        </div>
        <div className="flex items-center space-x-2">
          <AuthStatus />
          <ThemeSwitch />
        </div>
      </div>

      <ResizablePanelGroup direction="horizontal" className="h-full">
        {/* Left Sidebar - Directory Tree */}
        <ResizablePanel defaultSize={20} minSize={15} maxSize={35}>
          <DirectoryTree 
            onFileOpen={handleFileOpen}
            onFileCreate={handleFileCreate}
            onFileDelete={handleFileDelete}
            items={currentTables}
          />
        </ResizablePanel>
        
        <ResizableHandle className="w-1 bg-obsidian-border hover:bg-obsidian-accent transition-colors" />
        
        {/* Main Content Area */}
        <ResizablePanel defaultSize={60} minSize={40}>
          <DraggableTabSystem 
            tabs={openTabs}
            activeTab={activeTab}
            onTabChange={setActiveTab}
            onTabClose={handleTabClose}
            onTabSplit={(tabId, direction) => {
              console.log("Split tab", tabId, direction);
            }}
            onTabsReorder={handleTabsReorder}
            onTabAdd={handleTabAdd}
            onContentChange={handleContentChange}
          />
        </ResizablePanel>
        
        <ResizableHandle className="w-1 bg-obsidian-border hover:bg-obsidian-accent transition-colors" />
        
        {/* Right Sidebar */}
        <ResizablePanel defaultSize={20} minSize={15} maxSize={35}>
          <RightSidebar />
        </ResizablePanel>
      </ResizablePanelGroup>
      <Toaster />
    </div>
  );
};