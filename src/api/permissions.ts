/**
 * 权限资产清单 api 客户端（新增，task 6.2 图形化策略编辑器动作词表数据源）。
 *
 * 端点实核：GET /api/iam/permission/manifest
 * （eiam/internal/web/permission/handler.go:51-53；vo.go:5-19）。
 * - Capability("权限资产清单","manifest") + Needs("iam:permission:menus_by_urns")
 *   → 调用者须同时持有 `iam:permission:manifest` 与 `iam:permission:menus_by_urns`
 *   双码；缺任一，调用方应降级为自由输入词表（PolicyList 图形化编辑器口径）。
 * - 响应信封 data = Manifest{actions:[PermissionVO], services:[ServicePermissionEntry]}：
 *   PermissionVO{id, service, group, code, name, has_menu, menu_urns, access_scope_presets}
 *   （vo.go:5-14 实核）。本客户端仅归一 actions → ManifestAction（词表用）；
 *   services 树（Entry 递归 children）P1 不消费（parity §5.0 明示不消费口径）。
 *
 * 归一规则（Hard Rule）：snake_case → camelCase（menu_urns → menuUrns）；
 * 缺 code 项过滤；name 缺省回退 code。
 */
import { eiamAxios, unwrapEnvelope } from "@/api/request/eiam";
import type { Envelope } from "@/api/types";

/** 归一后的权限动作词表条目（图形化编辑器选项用） */
export interface ManifestAction {
  /** 能力码（eiam 动作码，statement.action 取值） */
  code: string;
  /** 动作名（选项展示） */
  name: string;
  /** 模块归属（选项分组键） */
  group: string;
  /** 服务（恒 iam），供跨服务策略上下文标注 */
  service: string;
  /** 是否有菜单（P1 菜单视图数据源，暂不消费） */
  hasMenu: boolean;
  /** 菜单 URN（P1 不消费，透传保留） */
  menuUrns: string[];
}

// ---- eiam 原生形状（私有，仅本模块映射用；snake_case 不越出本模块）----

interface EiamPermissionRaw {
  id?: number;
  service?: string;
  group?: string;
  code?: string;
  name?: string;
  has_menu?: boolean;
  menu_urns?: unknown;
}

function isRecord(v: unknown): v is Record<string, unknown> {
  return typeof v === "object" && v !== null && !Array.isArray(v);
}

function asString(v: unknown): string {
  return typeof v === "string" && v !== "" ? v : "";
}

/**
 * eiam PermissionVO → 归一 ManifestAction。
 * 缺 code（词表键）→ null（调用方过滤）。
 */
export function mapManifestAction(raw: unknown): ManifestAction | null {
  if (!isRecord(raw)) return null;
  const eiam = raw as EiamPermissionRaw;
  const code = asString(eiam.code);
  if (!code) return null;
  return {
    code,
    name: asString(eiam.name) || code,
    group: asString(eiam.group),
    service: asString(eiam.service),
    hasMenu: eiam.has_menu === true,
    menuUrns: Array.isArray(eiam.menu_urns)
      ? (eiam.menu_urns.filter((u): u is string => typeof u === "string") as string[])
      : [],
  };
}

/** 归一 Manifest.actions → ManifestAction[]（过滤 null；空缺省 []，调用方降级自由输入） */
export function mapManifestActions(raw: unknown): ManifestAction[] {
  if (!isRecord(raw)) return [];
  const actions = raw.actions;
  if (!Array.isArray(actions)) return [];
  return actions
    .map(mapManifestAction)
    .filter((a): a is ManifestAction => a !== null);
}

/**
 * 拉取权限资产清单（GET /api/iam/permission/manifest）。
 * @returns ManifestAction[]（动作词表；失败由调用方 catch 降级自由输入）
 */
export async function fetchPermissionManifest(): Promise<ManifestAction[]> {
  const data = await unwrapEnvelope<{ actions?: unknown[] }>(
    eiamAxios.get<Envelope<{ actions?: unknown[] }>>(
      "/api/iam/permission/manifest",
    ),
  );
  const actions = Array.isArray(data.actions) ? data.actions : [];
  return actions
    .map(mapManifestAction)
    .filter((a): a is ManifestAction => a !== null);
}