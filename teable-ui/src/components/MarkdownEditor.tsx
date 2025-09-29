import { useState, useRef } from "react";
import { Button } from "@/components/ui/button";
import { ArrowLeft, ArrowRight, MoreHorizontal } from "lucide-react";
import { MarkdownToolbar } from "./MarkdownToolbar";

interface MarkdownEditorProps {
  content: string;
  onChange: (content: string) => void;
  title?: string;
}

export const MarkdownEditor = ({ content, onChange, title }: MarkdownEditorProps) => {
  const [editorContent, setEditorContent] = useState(content);
  const [isPreview, setIsPreview] = useState(false);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const handleContentChange = (value: string) => {
    setEditorContent(value);
    onChange(value);
  };

  const handleToolbarInsert = (syntax: string) => {
    if (!textareaRef.current) return;
    
    const textarea = textareaRef.current;
    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const selectedText = editorContent.substring(start, end);
    
    let newText = "";
    let cursorOffset = 0;
    
    if (syntax.includes("[文本](链接)")) {
      newText = `[${selectedText || "文本"}](链接)`;
      cursorOffset = selectedText ? newText.length - 4 : 1;
    } else if (syntax.includes("![alt](链接)")) {
      newText = `![${selectedText || "alt"}](链接)`;
      cursorOffset = selectedText ? newText.length - 4 : 3;
    } else if (syntax.endsWith(" ")) {
      // Prefix syntax like headings, lists
      const lines = selectedText.split('\n');
      newText = lines.map(line => syntax + line).join('\n');
      cursorOffset = newText.length;
    } else {
      // Wrapper syntax like bold, italic
      newText = `${syntax}${selectedText}${syntax}`;
      cursorOffset = selectedText ? newText.length : syntax.length;
    }
    
    const newContent = editorContent.substring(0, start) + newText + editorContent.substring(end);
    handleContentChange(newContent);
    
    setTimeout(() => {
      textarea.focus();
      textarea.setSelectionRange(start + cursorOffset, start + cursorOffset);
    }, 0);
  };

  // Simple markdown to HTML conversion for preview
  const markdownToHtml = (markdown: string) => {
    return markdown
      .replace(/^# (.*$)/gim, '<h1 class="text-2xl font-bold mb-4 text-obsidian-text">$1</h1>')
      .replace(/^## (.*$)/gim, '<h2 class="text-xl font-semibold mb-3 text-obsidian-text">$1</h2>')
      .replace(/^### (.*$)/gim, '<h3 class="text-lg font-medium mb-2 text-obsidian-text">$1</h3>')
      .replace(/\*\*(.*)\*\*/gim, '<strong class="font-semibold">$1</strong>')
      .replace(/\*(.*)\*/gim, '<em class="italic">$1</em>')
      .replace(/\n/gim, '<br>');
  };

  return (
    <div className="h-full flex flex-col bg-obsidian-bg">
      {/* Editor Header */}
      <div className="flex items-center justify-between px-4 py-2 border-b border-obsidian-border bg-obsidian-surface">
        <div className="flex items-center space-x-2">
          <Button variant="ghost" size="sm" className="text-obsidian-text-muted hover:text-obsidian-text">
            <ArrowLeft className="w-4 h-4" />
          </Button>
          <Button variant="ghost" size="sm" className="text-obsidian-text-muted hover:text-obsidian-text">
            <ArrowRight className="w-4 h-4" />
          </Button>
          <span className="text-sm text-obsidian-text-muted ml-4">{title ?? "新标签页"}</span>
        </div>
        
        <div className="flex items-center space-x-2">
          <Button
            variant="ghost"
            size="sm"
            className="text-obsidian-text-muted hover:text-obsidian-text"
            onClick={() => setIsPreview(!isPreview)}
          >
            {isPreview ? "编辑" : "预览"}
          </Button>
          <Button variant="ghost" size="sm" className="text-obsidian-text-muted hover:text-obsidian-text">
            <MoreHorizontal className="w-4 h-4" />
          </Button>
        </div>
      </div>

      {/* Markdown Toolbar */}
      {!isPreview && <MarkdownToolbar onInsert={handleToolbarInsert} />}

      {/* Editor Content */}
      <div className="flex-1 overflow-hidden">
        {isPreview ? (
          <div className="h-full p-6 overflow-y-auto">
            <div
              className="prose prose-invert max-w-none text-obsidian-text"
              dangerouslySetInnerHTML={{ __html: markdownToHtml(editorContent) }}
            />
          </div>
        ) : (
          <div className="h-full p-6">
            <textarea
              ref={textareaRef}
              value={editorContent}
              onChange={(e) => handleContentChange(e.target.value)}
              className="w-full h-full bg-transparent text-obsidian-text placeholder-obsidian-text-muted border-none outline-none resize-none font-mono text-sm leading-relaxed"
              placeholder="开始写作..."
              spellCheck={false}
            />
          </div>
        )}
      </div>

      {/* Quick Actions Bar */}
      <div className="px-6 py-4 bg-obsidian-bg">
        <div className="flex flex-col space-y-2">
          <Button
            variant="ghost"
            size="sm"
            className="w-full justify-start text-obsidian-accent hover:text-obsidian-accent-hover hover:bg-obsidian-surface"
          >
            创建新文件 (⌘ N)
          </Button>
          <Button
            variant="ghost"
            size="sm"
            className="w-full justify-start text-obsidian-accent hover:text-obsidian-accent-hover hover:bg-obsidian-surface"
          >
            打开文件 (⌘ O)
          </Button>
          <Button
            variant="ghost"
            size="sm"
            className="w-full justify-start text-obsidian-accent hover:text-obsidian-accent-hover hover:bg-obsidian-surface"
          >
            查看近期文件 (⌘ O)
          </Button>
          <Button
            variant="ghost"
            size="sm"
            className="w-full justify-start text-obsidian-accent hover:text-obsidian-accent-hover hover:bg-obsidian-surface"
          >
            关闭标签页
          </Button>
        </div>
      </div>
    </div>
  );
};