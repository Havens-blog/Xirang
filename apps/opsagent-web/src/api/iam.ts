/**
 * eiam 用户档案 API（仅 profile 一档，收敛于本应用需求）。
 */
import { eiamAxios, unwrapEiam } from "./request/eiam";

/** eiam 用户档案（字段镜像 eiam /api/user/profile 载荷，平台共享形态） */
export interface UserProfile {
  username: string;
  nickname?: string;
  is_admin?: boolean;
  tenant_id?: number;
}

/** 获取当前用户档案（cookie → Bearer，403/401 由请求层收敛） */
export function fetchUserProfile(): Promise<UserProfile> {
  return unwrapEiam<UserProfile>(eiamAxios.get("/user/profile"));
}