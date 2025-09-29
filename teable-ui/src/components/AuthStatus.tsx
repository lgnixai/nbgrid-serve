import { Button } from "@/components/ui/button";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { LogOut, RefreshCw, AlertCircle } from "lucide-react";
import { useAuth } from "@/hooks/useAuth";

export const AuthStatus = () => {
  const { isAuthenticated, isLoading, user, error, logout, clearError } = useAuth();

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 text-sm text-obsidian-muted">
        <RefreshCw className="h-4 w-4 animate-spin" />
        连接中...
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center gap-2">
        <Alert className="border-red-500 bg-red-50 dark:bg-red-950">
          <AlertCircle className="h-4 w-4 text-red-500" />
          <AlertDescription className="text-red-700 dark:text-red-300">
            {error}
          </AlertDescription>
        </Alert>
        <Button
          variant="outline"
          size="sm"
          onClick={clearError}
          className="h-8"
        >
          清除
        </Button>
      </div>
    );
  }

  if (!isAuthenticated) {
    return (
      <div className="flex items-center gap-2 text-sm text-obsidian-muted">
        <AlertCircle className="h-4 w-4" />
        未连接
      </div>
    );
  }

  return (
    <div className="flex items-center gap-2">
      <div className="text-sm text-obsidian-text">
        已连接: {user?.email || "test@example.com"}
      </div>
      <Button
        variant="outline"
        size="sm"
        onClick={logout}
        className="h-8 gap-1"
      >
        <LogOut className="h-3 w-3" />
        登出
      </Button>
    </div>
  );
};
