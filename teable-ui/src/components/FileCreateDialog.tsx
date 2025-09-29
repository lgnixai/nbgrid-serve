import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

interface FileCreateDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onCreateFile: (name: string, type: string) => void;
}

export const FileCreateDialog = ({ open, onOpenChange, onCreateFile }: FileCreateDialogProps) => {
  const [fileName, setFileName] = useState("");
  const [fileType, setFileType] = useState("markdown");

  const handleCreate = () => {
    if (fileName.trim()) {
      const extension = fileType === "markdown" ? ".md" : ".txt";
      const fullName = fileName.endsWith(extension) ? fileName : fileName + extension;
      onCreateFile(fullName, fileType);
      setFileName("");
      onOpenChange(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="bg-obsidian-surface border-obsidian-border text-obsidian-text">
        <DialogHeader>
          <DialogTitle>创建新文件</DialogTitle>
        </DialogHeader>
        
        <div className="space-y-4">
          <div>
            <Label htmlFor="fileName" className="text-obsidian-text">文件名</Label>
            <Input
              id="fileName"
              value={fileName}
              onChange={(e) => setFileName(e.target.value)}
              placeholder="输入文件名..."
              className="mt-2 bg-obsidian-bg border-obsidian-border text-obsidian-text placeholder-obsidian-text-muted"
              autoFocus
            />
          </div>
          
          <div>
            <Label htmlFor="fileType" className="text-obsidian-text">文件类型</Label>
            <Select value={fileType} onValueChange={setFileType}>
              <SelectTrigger className="mt-2 bg-obsidian-bg border-obsidian-border text-obsidian-text">
                <SelectValue />
              </SelectTrigger>
              <SelectContent className="bg-obsidian-surface border-obsidian-border">
                <SelectItem value="markdown" className="text-obsidian-text">Markdown (.md)</SelectItem>
                <SelectItem value="text" className="text-obsidian-text">文本文件 (.txt)</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>

        <DialogFooter>
          <Button
            variant="ghost"
            onClick={() => onOpenChange(false)}
            className="text-obsidian-text-muted hover:text-obsidian-text"
          >
            取消
          </Button>
          <Button
            onClick={handleCreate}
            className="bg-obsidian-accent text-obsidian-bg hover:bg-obsidian-accent-hover"
          >
            创建
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};