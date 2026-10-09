/**
 * eiam 统一身份服务请求层（/api/iam/*，nginx 与 vite dev proxy 重写为 eiam :9000 /api/*）。
 *
 * 用途收敛：仅拉取当前用户档案（settings 页 isAdmin 等）。行为口径与
 * haven-console src/api/request/eiam.ts 一致：withCredentials、cookie→Bearer 注入、
 * 401 跳运维平台登录（不清理共享 cookie）、15s 超时。
 */
import { getSessionToken, redirectToLogin } from "./session";
import axios, { type AxiosInstance, type AxiosResponse } from "axios";

/** 请求超时（与平台侧 eiam 请求层一致） */
const REQUEST_TIMEOUT_MS = 15000;

export const eiamAxios: AxiosInstance = axios.create({
  timeout: REQUEST_TIMEOUT_MS,
  withCredentials: true,
  headers: { "Content-Type": "application/json" },
});

eiamAxios.interceptors.request.use((config) => {
  const token = getSessionToken();
  if (token && config.headers) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

eiamAxios.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      redirectToLogin();
    }
    return Promise.reject(error);
  },
);

/** ginx Result 信封（eiam 形态 {code,msg,data}） */
export interface EiamEnvelope<T = unknown> {
  code: number;
  msg?: string;
  data?: T;
}

/** 单出口解包：2xx + code===0 → data；2xx + code!=0 → 抛错（msg 原文）；非 2xx 原样上抛 */
export async function unwrapEiam<T>(
  request: Promise<AxiosResponse<EiamEnvelope<T>>>,
): Promise<T> {
  const response = await request;
  const body: unknown = response.data;
  if (typeof body !== "object" || body === null || !("code" in body)) {
    throw new Error("eiam 接口响应格式错误（缺少 code 信封字段）");
  }
  const envelope = body as EiamEnvelope<T>;
  if (typeof envelope.code !== "number" || envelope.code !== 0) {
    throw new Error(
      typeof envelope.msg === "string" && envelope.msg.length > 0
        ? envelope.msg
        : "eiam 请求失败",
    );
  }
  return envelope.data as T;
}