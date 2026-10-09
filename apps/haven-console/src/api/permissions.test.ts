// @vitest-environment happy-dom
/**
 * permissions.ts manifest 客户端映射单测（task 6.2 图形化策略编辑器词表）。
 *
 * 覆盖：
 * - mapManifestAction：snake_case → camelCase（menu_urns → menuUrns；has_menu → hasMenu）；
 *   缺 code → null；name 缺省回退 code。
 * - mapManifestActions：过滤 null / 非数组输入空集。
 * - fetchPermissionManifest：GET /api/iam/permission/manifest 载荷走 eiamAxios，
 *   归一 actions → ManifestAction[]；非 0 码 / 网络错抛 ApiError。
 */
import {
  AxiosError,
  type AxiosResponse,
  type InternalAxiosRequestConfig,
} from "axios";
import { afterEach, describe, expect, it } from "vitest";
import {
  fetchPermissionManifest,
  mapManifestAction,
  mapManifestActions,
} from "./permissions";
import { eiamAxios } from "@/api/request/eiam";

const STATUS_TEXT: Partial<Record<number, string>> = {
  200: "OK",
};

interface MockReply {
  status?: number;
  body?: unknown;
  networkFailure?: { code: string; message: string };
}

function settleLike(response: AxiosResponse): AxiosResponse {
  const validate =
    response.config.validateStatus ?? ((s: number) => s >= 200 && s < 300);
  if (validate(response.status)) return response;
  throw new AxiosError(
    `Request failed with status code ${response.status}`,
    AxiosError.ERR_BAD_REQUEST,
    response.config,
    undefined,
    response,
  );
}

function adapterFor(
  replyFor: (config: InternalAxiosRequestConfig) => MockReply,
  capture: InternalAxiosRequestConfig[] = [],
): void {
  eiamAxios.defaults.adapter = (async (config: InternalAxiosRequestConfig) => {
    capture.push(config);
    const reply = replyFor(config);
    if (reply.networkFailure) {
      throw new AxiosError(
        reply.networkFailure.message,
        reply.networkFailure.code,
        config,
      );
    }
    return settleLike({
      data: reply.body ?? null,
      status: reply.status ?? 200,
      statusText: STATUS_TEXT[reply.status ?? 200] ?? "",
      headers: {},
      config,
    } as AxiosResponse);
  }) as unknown as typeof eiamAxios.defaults.adapter;
}

afterEach(() => {
  eiamAxios.defaults.adapter = undefined;
});

const RAW_ACTION = {
  id: 7,
  service: "iam",
  group: "身份源管理",
  code: "iam:identity_source:toggle",
  name: "启停身份源",
  has_menu: true,
  menu_urns: ["iam_identity_source_toggle"],
};

describe("mapManifestAction（snake_case → camelCase 归一）", () => {
  it("完整字段归一：code/name/group/service/has_menu/menu_urns", () => {
    const a = mapManifestAction(RAW_ACTION);
    expect(a).toEqual({
      code: "iam:identity_source:toggle",
      name: "启停身份源",
      group: "身份源管理",
      service: "iam",
      hasMenu: true,
      menuUrns: ["iam_identity_source_toggle"],
    });
  });

  it("name 缺失 → 回退 code", () => {
    const a = mapManifestAction({ code: "x", name: "" });
    expect(a?.name).toBe("x");
  });

  it("缺 code（词表键）→ null", () => {
    expect(mapManifestAction({ name: "无码" })).toBeNull();
    expect(mapManifestAction(null)).toBeNull();
    expect(mapManifestAction("str")).toBeNull();
  });

  it("menu_urns 非字符串数组 → []", () => {
    const a = mapManifestAction({ code: "x", menu_urns: [1, "u", null] });
    expect(a?.menuUrns).toEqual(["u"]);
  });
});

describe("mapManifestActions（批量归一）", () => {
  it("过滤 null 保留有效项", () => {
    const items = mapManifestActions({
      actions: [
        RAW_ACTION,
        { name: "缺码" },
        null,
        { code: "iam:user:view", name: "用户查看", group: "用户管理" },
      ],
    });
    expect(items.map((a) => a.code)).toEqual([
      "iam:identity_source:toggle",
      "iam:user:view",
    ]);
  });

  it("actions 非数组 → []", () => {
    expect(mapManifestActions({ actions: "x" })).toEqual([]);
    expect(mapManifestActions(null)).toEqual([]);
  });
});

describe("fetchPermissionManifest（GET /api/iam/permission/manifest）", () => {
  it("归一 actions → ManifestAction[]（GET 方法 + 正确路径）", async () => {
    const capture: InternalAxiosRequestConfig[] = [];
    adapterFor(
      () => ({
        status: 200,
        body: {
          code: 0,
          msg: "",
          data: {
            actions: [RAW_ACTION, { code: "iam:user:view", name: "用户查看" }],
            services: [{ code: "iam", name: "统一身份", entries: [] }],
          },
        },
      }),
      capture,
    );
    const actions = await fetchPermissionManifest();
    expect(capture[0]?.method).toBe("get");
    expect(capture[0]?.url).toContain("/api/iam/permission/manifest");
    expect(actions).toHaveLength(2);
    expect(actions[0]).toMatchObject({ code: "iam:identity_source:toggle" });
    // services 树不消费（P1 口径）
    expect(actions[0]).not.toHaveProperty("entries");
  });

  it("信封非 0 码 → 抛 ApiError（调用方降级自由输入）", async () => {
    adapterFor(() => ({
      status: 200,
      body: { code: 40100, msg: "无权", data: null },
    }));
    await expect(fetchPermissionManifest()).rejects.toThrow();
  });

  it("网络错 → 抛 ApiError unavailable", async () => {
    adapterFor(() => ({
      networkFailure: { code: "ECONNABORTED", message: "timeout" },
    }));
    await expect(fetchPermissionManifest()).rejects.toThrow();
  });
});