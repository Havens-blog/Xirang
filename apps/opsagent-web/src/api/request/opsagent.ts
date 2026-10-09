/**
 * 运维 Agent（D:/Haven/opsagent，/api/v1/opsagent）专用 axios 实例。
 *
 * 后端信封为 {code, message, data}，但成功 code 存在两态：多数端点经 web.OK 写
 * code=0（数字），chat/correct 端点手写 code="0"（字符串）。主机实（./eiam.ts）
 * 按 ginx 信封契约解包，不适用于双态成功码；故独立成实例，信封解包逻辑收敛在
 * src/api/opsagent.ts 的 unwrapOpsagent。
 *
 * baseURL 同源（VITE_API_BASE_URL || '/api/v1'）；dev 经 vite proxy 指向 opsagent
 * 服务 :8081（与 nginx.dev.conf location ^~ /api/v1/opsagent/ 同构）。
 * 认证：共享 cookie ecmdb-token-key → Bearer；401 跳运维平台登录页（见 ./session.ts）。
 */
import { getSessionToken, redirectToLogin } from "./session";
import axios, { type AxiosInstance } from "axios";

export const opsagentAxios: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || "/api/v1",
  // chat 端点为同步编排（后端 syncBudget 25s），超时须高于 25s
  timeout: 30000,
  withCredentials: false,
});

opsagentAxios.interceptors.request.use((config) => {
  const token = getSessionToken();
  if (token && config.headers) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

opsagentAxios.interceptors.response.use(
  (response) => response,
  (error) => {
    // 401 未认证与会话过期：统一跳运维平台登录页（不清共享 cookie）
    if (error.response?.status === 401) {
      redirectToLogin();
    }
    return Promise.reject(error);
  },
);