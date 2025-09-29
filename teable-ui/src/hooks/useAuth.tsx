import { useState, useEffect, useCallback } from 'react';
import { useToast } from './use-toast';
import teable, { ensureLogin } from '@/lib/teable-simple';

interface AuthState {
  isAuthenticated: boolean;
  isLoading: boolean;
  user: any | null;
  error: string | null;
}

export const useAuth = () => {
  const [authState, setAuthState] = useState<AuthState>({
    isAuthenticated: false,
    isLoading: true,
    user: null,
    error: null,
  });
  const { toast } = useToast();

  const login = useCallback(async (email?: string, password?: string) => {
    setAuthState(prev => ({ ...prev, isLoading: true, error: null }));
    
    try {
      await ensureLogin(email && password ? { email, password } : undefined);
      setAuthState({
        isAuthenticated: true,
        isLoading: false,
        user: { email: email || "test@example.com" },
        error: null,
      });
      toast({
        title: "登录成功",
        description: "已成功连接到 Teable 后端",
      });
    } catch (error: any) {
      const errorMessage = error?.message || "登录失败";
      setAuthState({
        isAuthenticated: false,
        isLoading: false,
        user: null,
        error: errorMessage,
      });
      toast({
        title: "登录失败",
        description: errorMessage,
        variant: "destructive",
      });
      throw error;
    }
  }, [toast]);

  const logout = useCallback(async () => {
    setAuthState(prev => ({ ...prev, isLoading: true }));
    
    try {
      await teable.logout();
      setAuthState({
        isAuthenticated: false,
        isLoading: false,
        user: null,
        error: null,
      });
      toast({
        title: "已登出",
        description: "已成功登出",
      });
    } catch (error: any) {
      // 即使登出失败，也清除本地状态
      setAuthState({
        isAuthenticated: false,
        isLoading: false,
        user: null,
        error: null,
      });
      toast({
        title: "登出完成",
        description: "已清除本地登录状态",
      });
    }
  }, [toast]);

  const clearError = useCallback(() => {
    setAuthState(prev => ({ ...prev, error: null }));
  }, []);

  // 自动登录
  useEffect(() => {
    const autoLogin = async () => {
      try {
        await login();
      } catch (error) {
        // 自动登录失败，保持未认证状态
        console.warn('自动登录失败:', error);
      }
    };

    autoLogin();
  }, [login]);

  return {
    ...authState,
    login,
    logout,
    clearError,
  };
};
