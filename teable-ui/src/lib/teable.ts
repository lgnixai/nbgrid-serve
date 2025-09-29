import * as TeableSDK from "@teable/sdk";

// 基础配置：可根据需要改为从环境变量读取
const BASE_URL = import.meta.env.VITE_TEABLE_BASE_URL || "http://localhost:3000";

const Teable = TeableSDK.Teable || TeableSDK.default;
const LoginRequest = TeableSDK.LoginRequest;

const teable = new Teable({ baseUrl: BASE_URL, debug: true });

let loginPromise: Promise<void> | null = null;

export const ensureLogin = (creds?: LoginRequest): Promise<void> => {
  if (teable.isAuthenticated()) return Promise.resolve();
  if (loginPromise) return loginPromise;

  const credentials: LoginRequest = creds ?? {
    email: "test@example.com",
    password: "TestPassword123!",
  };

  loginPromise = teable
    .login(credentials)
    .then(() => {})
    .finally(() => {
      loginPromise = null;
    });

  return loginPromise;
};

export default teable;


