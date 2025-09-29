import { useState } from "react";
import { Search, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

interface SearchBarProps {
  onSearch: (query: string) => void;
  onClose: () => void;
}

export const SearchBar = ({ onSearch, onClose }: SearchBarProps) => {
  const [query, setQuery] = useState("");

  const handleSearch = (value: string) => {
    setQuery(value);
    onSearch(value);
  };

  return (
    <div className="flex items-center gap-2 px-4 py-3 border-b border-obsidian-border bg-obsidian-surface">
      <div className="relative flex-1">
        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-obsidian-text-muted" />
        <Input
          value={query}
          onChange={(e) => handleSearch(e.target.value)}
          placeholder="搜索文件和内容..."
          className="pl-10 bg-obsidian-bg border-obsidian-border text-obsidian-text placeholder-obsidian-text-muted focus:border-obsidian-accent"
          autoFocus
        />
      </div>
      <Button
        variant="ghost"
        size="sm"
        onClick={onClose}
        className="text-obsidian-text-muted hover:text-obsidian-text"
      >
        <X className="w-4 h-4" />
      </Button>
    </div>
  );
};