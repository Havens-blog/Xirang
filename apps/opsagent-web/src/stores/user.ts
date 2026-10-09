import { defineStore } from "pinia";
import { ref } from "vue";
import { fetchUserProfile } from "@/api/iam";

/**
 * 当前用户档案（最小面：settings 页 isAdmin 门控 + AppShell 用户名展示）。
 * 会话凭证为平台共享 cookie；档案拉取失败不阻断使用（保持未登录降级，
 * 后端 EiamAuth 会对业务请求做权威拒绝）。
 */
export const useUserStore = defineStore("user", () => {
  /** 用户名（未拉取/失败时为空） */
  const username = ref<string>("");
  /** 平台管理员标记（settings 页编辑门控） */
  const isAdmin = ref<boolean>(false);
  /** 档案是否已拉取成功（防重复请求） */
  const loaded = ref<boolean>(false);

  /** 拉取档案：失败静默（业务请求由后端权威鉴权） */
  async function fetchProfile(): Promise<void> {
    if (loaded.value) return;
    try {
      const profile = await fetchUserProfile();
      username.value = profile.username ?? "";
      isAdmin.value = profile.is_admin === true;
      loaded.value = true;
    } catch {
      // 网络失败/会话失效：401 已由请求层跳登录；此处静默保持降级态
    }
  }

  return { username, isAdmin, loaded, fetchProfile };
});