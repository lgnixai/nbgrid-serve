import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { 
  Bold, 
  Italic, 
  Strikethrough, 
  List, 
  ListOrdered, 
  Quote, 
  Code, 
  Link, 
  Image,
  Heading1,
  Heading2,
  Heading3
} from "lucide-react";

interface MarkdownToolbarProps {
  onInsert: (syntax: string) => void;
}

export const MarkdownToolbar = ({ onInsert }: MarkdownToolbarProps) => {
  const toolbarItems = [
    { icon: Bold, syntax: "**", label: "粗体" },
    { icon: Italic, syntax: "*", label: "斜体" },
    { icon: Strikethrough, syntax: "~~", label: "删除线" },
    { separator: true },
    { icon: Heading1, syntax: "# ", label: "标题1", prefix: true },
    { icon: Heading2, syntax: "## ", label: "标题2", prefix: true },
    { icon: Heading3, syntax: "### ", label: "标题3", prefix: true },
    { separator: true },
    { icon: List, syntax: "- ", label: "无序列表", prefix: true },
    { icon: ListOrdered, syntax: "1. ", label: "有序列表", prefix: true },
    { icon: Quote, syntax: "> ", label: "引用", prefix: true },
    { separator: true },
    { icon: Code, syntax: "`", label: "行内代码" },
    { icon: Link, syntax: "[文本](链接)", label: "链接", custom: true },
    { icon: Image, syntax: "![alt](链接)", label: "图片", custom: true },
  ];

  const handleInsert = (item: any) => {
    if (item.custom) {
      onInsert(item.syntax);
    } else if (item.prefix) {
      onInsert(item.syntax);
    } else {
      onInsert(item.syntax);
    }
  };

  return (
    <div className="flex items-center gap-1 px-3 py-2 border-b border-obsidian-border bg-obsidian-surface">
      {toolbarItems.map((item, index) => (
        item.separator ? (
          <Separator key={index} orientation="vertical" className="h-4 mx-1 bg-obsidian-border" />
        ) : (
          <Button
            key={index}
            variant="ghost"
            size="sm"
            className="h-7 w-7 p-0 text-obsidian-text-muted hover:text-obsidian-text hover:bg-obsidian-surface-hover"
            onClick={() => handleInsert(item)}
            title={item.label}
          >
            <item.icon className="w-4 h-4" />
          </Button>
        )
      ))}
    </div>
  );
};