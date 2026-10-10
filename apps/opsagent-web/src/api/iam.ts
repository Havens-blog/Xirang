/**
 * eiam 用户档案 API（仅 profile 一档，收敛于本应用需求）。
 *
 * 路径口径与平台一致：/api/iam/*（nginx 与 vite dev proxy 重写为 eiam :9000 /api/*，
 * 不直连 /api/... —— 2026-10-09 平台事故口径）。
 */
import { eiamAxios, unwrapEiam } from "./request/eiam";

/** eiam GET /api/iam/user/profile 响应 data（RetrieveUser） */
export interface UserProfile {
  /** eiam RetrieveUser.User（嵌套） */
  user?: {
    username?: string;
    nickname?: string;
    job_title?: string;
  };
  /** 平台管理员标记（roles 含 admin） */
  is_admin?: boolean;
  /** 当前租户 id（eiam current_tenant_id） */
  current_tenant_id?: number;
}

/** 获取当前用户档案（cookie → Bearer，403/401 由请求层收敛） */
export function fetchUserProfile(): Promise<UserProfile> {
  return unwrapEiam<UserProfile>(eiamAxios.get("/api/iam/user/profile"));
}