import { shallowRef, type Ref } from "vue";

/**
 * topbar 动作总线：页面把自己的动作注册到应用壳 topbar 右侧
 * （原型口径：系统配置「保存配置」、Agent 管理「重启所有/导出日志」都在 topbar）。
 * 页面 onMounted 注册、onBeforeUnmount 清空；AppShell 渲染。
 *
 * 单例实现：路由页面均为懒加载 chunk，普通模块级状态会被 rollup 复制成多份实例，
 * 父子不同步；故共享 ref 存 window 级单例（跨 chunk 恒唯一）。无 window 的
 * 环境（纯 node 测试）返回本地哑 ref，不抛错。
 */
export interface TopbarAction {
  /** 唯一键 */
  key: string;
  /** 按钮文案（原型逐字） */
  label: string;
  /** 主按钮样式（btn-primary）；默认 outline-sm */
  primary?: boolean;
  /** 点击处理（页面闭包绑好自身状态） */
  onClick: () => void;
}

const REF_KEY = "__opsagent_web_topbar_actions__";

type Registry = Record<string, unknown>;

function getWindow(): Registry | null {
  if (typeof window === "undefined") return null;
  return window as unknown as Registry;
}

export function useTopbar(): {
  actions: Ref<TopbarAction[]>;
  setActions: (next: TopbarAction[]) => void;
} {
  const win = getWindow();
  if (!win) {
    // 无 window 环境（node 单测直载）：本地哑 ref
    const dumb = shallowRef<TopbarAction[]>([]);
    return { actions: dumb, setActions: (next) => void (dumb.value = next) };
  }
  let shared = win[REF_KEY] as Ref<TopbarAction[]> | undefined;
  if (!shared) {
    shared = shallowRef<TopbarAction[]>([]);
    win[REF_KEY] = shared;
  }
  return {
    actions: shared,
    setActions: (next) => {
      shared!.value = next;
    },
  };
}