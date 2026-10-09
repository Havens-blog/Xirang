/**
 * useCan —— 权限码判定组合式函数（P0 按键级裁剪的判定源，与路由守卫/menuConfig 同口径）。
 *
 * 数据源 = user store 的 permissions（profile.permissions 归一，eiam 权威）；
 * 返回 can / canEvery / canSome 三个判定函数，随 store 响应式更新（登录/切租户/登出后自动收敛）。
 *
 * 使用形态（各页面）：
 *   const { can, canEvery, canSome } = useCan()
 *   <button v-if="can(IAM_USER.ADD)">新增用户</button>
 *   <button v-if="canSome([IAM_ROLE.BATCH_ASSIGN, IAM_POLICY.BATCH_ATTACH])">新增授权</button>
 *
 * 边界（BIZ-auth-001）：本判定仅体验层裁剪；服务端 OPA 403 兜底仍是唯一安全边界。
 */
import { computed } from "vue";
import { useUserStore } from "@/stores/user";

export interface CanApi {
  /** 是否持有单个能力码 */
  can: (code: string) => boolean;
  /** 是否持有全部能力码 */
  canEvery: (codes: readonly string[]) => boolean;
  /** 是否持有任一能力码 */
  canSome: (codes: readonly string[]) => boolean;
}

export function useCan(): CanApi {
  const userStore = useUserStore();

  const owned = computed(() => new Set(userStore.permissions));

  function can(code: string): boolean {
    return owned.value.has(code);
  }

  function canEvery(codes: readonly string[]): boolean {
    return codes.every((code) => owned.value.has(code));
  }

  function canSome(codes: readonly string[]): boolean {
    return codes.some((code) => owned.value.has(code));
  }

  return { can, canEvery, canSome };
}