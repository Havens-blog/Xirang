// @vitest-environment happy-dom
import { describe, expect, it } from "vitest";
import { routes } from "./index";

/**
 * 运维 Agent 独立前端路由表。
 * 口径：全部路由经 AppShell（非公开），六页懒加载；未匹配回落 /chat。
 */
describe("opsagent-web 路由表", () => {
  it("根路由 = AppShell + 六页子路由 + catch-all 回落 /chat", () => {
    const shell = routes[0]!;
    expect(shell.path).toBe("/");
    const childPaths = (shell.children ?? []).map((r) => r.path);
    expect(childPaths).toEqual([
      "",
      "chat",
      "risk-center",
      "diagnosis/:id",
      "history",
      "settings",
      "agents",
      ":pathMatch(.*)*",
    ]);
  });

  it("根子路由重定向到 /chat；未匹配路径回落 /chat", () => {
    const children = routes[0]!.children ?? [];
    expect(children.find((r) => r.path === "")?.redirect).toBe("/chat");
    const fallback = children.find((r) => r.path === ":pathMatch(.*)*");
    expect(fallback?.redirect).toBe("/chat");
  });

  it("六页路由组件均为动态 import()（懒加载，路由级代码分割）", () => {
    const children = routes[0]!.children ?? [];
    const pages = children.filter((r) => r.component !== undefined);
    for (const r of pages) {
      expect(typeof r.component).toBe("function");
    }
  });

  it("不声明权限元数据（登录态凭平台共享 cookie；租户边界由后端 EiamAuth 承接）", () => {
    for (const child of routes[0]!.children ?? []) {
      expect(child.meta?.permissions).toBeUndefined();
      expect(child.meta?.requiresAdmin).toBeUndefined();
    }
  });
});