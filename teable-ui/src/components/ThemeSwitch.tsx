import { useState } from "react";
import { Moon, Sun, Monitor } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

type Theme = "dark" | "light" | "system";

export const ThemeSwitch = () => {
  const [theme, setTheme] = useState<Theme>("dark");

  const handleThemeChange = (newTheme: Theme) => {
    setTheme(newTheme);
    
    if (newTheme === "system") {
      const isDark = window.matchMedia("(prefers-color-scheme: dark)").matches;
      document.documentElement.classList.toggle("dark", isDark);
    } else {
      document.documentElement.classList.toggle("dark", newTheme === "dark");
    }
  };

  const getThemeIcon = () => {
    switch (theme) {
      case "light":
        return <Sun className="w-4 h-4" />;
      case "dark":
        return <Moon className="w-4 h-4" />;
      case "system":
        return <Monitor className="w-4 h-4" />;
    }
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          size="sm"
          className="text-obsidian-text-muted hover:text-obsidian-text hover:bg-obsidian-surface-hover"
        >
          {getThemeIcon()}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent className="bg-obsidian-surface border-obsidian-border">
        <DropdownMenuItem
          onClick={() => handleThemeChange("light")}
          className="text-obsidian-text hover:bg-obsidian-surface-hover focus:bg-obsidian-surface-hover cursor-pointer"
        >
          <Sun className="w-4 h-4 mr-2" />
          浅色主题
        </DropdownMenuItem>
        <DropdownMenuItem
          onClick={() => handleThemeChange("dark")}
          className="text-obsidian-text hover:bg-obsidian-surface-hover focus:bg-obsidian-surface-hover cursor-pointer"
        >
          <Moon className="w-4 h-4 mr-2" />
          深色主题
        </DropdownMenuItem>
        <DropdownMenuItem
          onClick={() => handleThemeChange("system")}
          className="text-obsidian-text hover:bg-obsidian-surface-hover focus:bg-obsidian-surface-hover cursor-pointer"
        >
          <Monitor className="w-4 h-4 mr-2" />
          跟随系统
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
};