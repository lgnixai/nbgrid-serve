import React, { useState } from 'react';
import { GlideDataGridGuide } from '@/components/GlideDataGridGuide';
import { AdvancedGlideGrid } from '@/components/AdvancedGlideGrid';
import { TeableDataGrid } from '@/components/TeableDataGrid';

export const GlideGridDemo: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'basic' | 'advanced' | 'teable'>('basic');

  const tabs = [
    { id: 'basic', label: '基础示例', component: <GlideDataGridGuide /> },
    { id: 'advanced', label: '高级功能', component: <AdvancedGlideGrid /> },
    { id: 'teable', label: 'Teable 集成', component: <TeableDataGrid tableId="demo" baseId="demo" /> },
  ];

  return (
    <div className="h-screen bg-gray-50">
      {/* 标签页导航 */}
      <div className="bg-white border-b border-gray-200">
        <div className="px-6 py-4">
          <h1 className="text-2xl font-bold text-gray-900 mb-4">
            Glide Data Grid 演示
          </h1>
          
          <div className="flex space-x-1">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id as any)}
                className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
                  activeTab === tab.id
                    ? 'bg-blue-100 text-blue-700 border border-blue-200'
                    : 'text-gray-500 hover:text-gray-700 hover:bg-gray-100'
                }`}
              >
                {tab.label}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* 内容区域 */}
      <div className="flex-1 overflow-hidden">
        {tabs.find(tab => tab.id === activeTab)?.component}
      </div>
    </div>
  );
};
