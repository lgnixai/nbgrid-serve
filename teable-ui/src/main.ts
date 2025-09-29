import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import './style.css'

function Sidebar() {
  return (
    <div className="w-72 border-r h-full flex flex-col">
      <div className="p-3 border-b">
        <select className="w-full border rounded px-2 py-1">
          <option>选择 Space...</option>
        </select>
        <button className="mt-2 w-full border rounded px-2 py-1">+ 创建 Space</button>
      </div>
      <div className="flex-1 overflow-auto p-2">
        <div className="text-sm text-gray-500 mb-2">Bases & Tables</div>
        <ul className="text-sm space-y-1">
          <li className="font-semibold">示例 Base</li>
          <li className="pl-4">示例表 Table</li>
        </ul>
      </div>
    </div>
  )
}

function ContentPlaceholder() {
  return (
    <div className="flex-1 h-full flex items-center justify-center text-gray-500">
      内容区占位符（后续设计表视图）
    </div>
  )
}

function AppShell() {
  return (
    <div className="h-full flex">
      <Sidebar />
      <ContentPlaceholder />
    </div>
  )
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <BrowserRouter>
      <Routes>
        <Route path="/*" element={<AppShell />} />
      </Routes>
    </BrowserRouter>
  </React.StrictMode>,
)
