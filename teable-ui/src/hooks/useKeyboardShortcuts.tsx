import { useEffect } from "react";

interface ShortcutActions {
  onNewFile?: () => void;
  onOpenFile?: () => void;
  onSave?: () => void;
  onSearch?: () => void;
  onCloseTab?: () => void;
  onTogglePreview?: () => void;
}

export const useKeyboardShortcuts = (actions: ShortcutActions) => {
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      const { ctrlKey, metaKey, key, shiftKey } = event;
      const isCtrlOrCmd = ctrlKey || metaKey;

      if (isCtrlOrCmd) {
        switch (key.toLowerCase()) {
          case 'n':
            event.preventDefault();
            actions.onNewFile?.();
            break;
          case 'o':
            event.preventDefault();
            actions.onOpenFile?.();
            break;
          case 's':
            event.preventDefault();
            actions.onSave?.();
            break;
          case 'f':
            event.preventDefault();
            actions.onSearch?.();
            break;
          case 'w':
            event.preventDefault();
            actions.onCloseTab?.();
            break;
          case 'e':
            if (shiftKey) {
              event.preventDefault();
              actions.onTogglePreview?.();
            }
            break;
        }
      }

      // ESC key
      if (key === 'Escape') {
        // Can be used to close search, dialogs, etc.
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [actions]);
};