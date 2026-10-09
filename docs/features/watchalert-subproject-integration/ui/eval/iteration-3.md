# UI Design Evaluation — Iteration 3

**Document**: `ui-design.md`
**Rubric**: `ui-web.md` (1000-point scale, target 950)
**PRD**: `prd-ui-functions.md`
**Scorer persona**: Senior UX Engineer
**Date**: 2026-10-09
**Previous iteration**: 821/1000

---

## Iteration-2 Issue Tracker

| Iteration-2 issue | Status in iteration 3 | Evidence |
|------------------|----------------------|----------|
| Conflict 1 — Component 3 recovery: 150ms message vs "无提示停顿" | **Fixed** | Interactions table now says "先渲染 150ms 瞬时过渡文案…过渡文案与 States 表「已恢复」行为一致——以 States 表为准，Interactions 不再表述「无提示停顿」" |
| Conflict 2 — Component 3 Error aria-live: polite vs assertive | **Fixed** | Accessibility section adds explicit override: "Error 态为该规则的特例覆盖：升级为 `role="alert"` + `aria-live="assertive"`…二者不冲突——通用规则管 Loading/Empty/Populated，Error 行管请求失败态" |
| Conflict 3 — Component 2 503 note case (b) misclassifies 401 | **Fixed** | 503 note now says "跨标签页登出（BIZ-auth-005）场景在其余标签页的下一请求返回 **401** 而非 503，归入下表 401 分派，不混入 503"; plus unified HTTP error dispatch table added |
| `recoverySignal` vs `authHealth='recovered'` redundancy | **Fixed** | Component 3 Data Binding: "不引入独立 `recoverySignal` 字段，恢复信号是 `authHealth` 枚举值之一，避免双通道冗余与数据契约歧义" |
| Component 4 no transition-out / recovery | **Fixed** | Component 4 States adds "角色已分配（恢复）" with 60s polling + manual refresh trigger |
| `prefers-color-scheme` vs `[data-theme="dark"]` precedence | **Fixed** | Dark palette section: "`hc_theme` 存在时以手动选择为准（覆盖系统偏好），`hc_theme` 缺省时回退系统 `prefers-color-scheme`" with CSS cascade guidance |
| Banner `height 0↔auto` not CSS-animatable | **Fixed** | Motion table note: "采用「测量高度 + max-height」技术——展开前以 `scrollHeight` 测量目标高度并设为 `max-height`，过渡结束后清除 `max-height` 改回 `auto`" |
| Card padding inconsistency (p-6 vs p-4) | **Fixed** | New "Card（compact 变体）" component spec: "同 Card，但内边距 p-4；用于信息密度高的概览/统计卡片…这是对 p-6 基准的显式有依据偏离，不是遗漏" |
| Read-only notice bar border ambiguous | **Fixed** | Component 3 layout: "1px solid var(--border)（底边，遵循「全部 1px 细边框」Do）" |
| Component 1 Error message not in Data Binding | **Not fixed** | Still absent from Data Binding table; "登录失败，请重试" only in States |
| Component 2 Badge「只读」not in Data Binding | **Not fixed** | Still only in layout ("writeAllowed=false 时显示"), not explicitly mapped in Data Binding |
| `eiamAuthUrl` config-absent scenario | **Fixed** | Component 1 States adds "配置缺失（`eiamAuthUrl` 为 null）" state with disabled button + 30s retry |
| Back button during Component 1 Loading | **Not fixed** | "不阻塞浏览器跳转" stated but back-button state during redirect still unspecified |
| Multi-tab state staleness (BIZ-auth-005) | **Fixed** | New "多标签页状态陈旧窗口" section with staleness bounds + reconciliation table |
| 401/403/503 dispatch fragmented | **Fixed** | New "全局 HTTP 错误分派表（401 / 403 / 503）" with unified interceptor dispatch |

Net: 12 of 15 prior issues resolved; 3 minor items not addressed (Component 1 Error in Data Binding, Component 2 Badge mapping, back-button state). All three cross-section contradictions from iteration 2 are resolved. No new contradictions introduced.

---

## Phase 1: Reasoning Audit

### Problem → Solution
3 PRD UI functions → 4 components. UF-1 split across Component 1 (login entry) and Component 4 (403 in-place). UF-2 → Component 2. UF-3 → Component 3. Page Composition table extends coverage to 告警规则 and 故障中心. All UI functions mapped.

### Solution → Evidence
Layout ASCII trees per component. States / Interactions / Data Binding tables per component. Design System with Light + Dark palettes (precedence resolved), typography, component specs (including compact Card variant), icons, motion (with banner height animation technique), accessibility (with explicit Error override). Navigation & App Shell with shared nav + banner slot + responsive table. SSE transport-failure table. Multi-tab staleness table. Session-expiry (401) handling. Unified HTTP error dispatch table. Shared `AuthHealth` type contract (recoverySignal removed). Component 4 recovery state added. eiamAuthUrl config-absent state added. Structurally excellent — this is a near-complete spec.

### Self-contradiction / Pre-score Anchors (independent findings)

1. **No direct contradictions found.** All three iteration-2 conflicts are resolved with explicit override/reconciliation language. The recoverySignal redundancy is eliminated. The aria-live override is explicit. The 503/401 dispatch is unified.

2. **Page-specific read-only layouts absent.** Component 3 is defined for 告警看板 (layout: 概览 + 规则 + 时间线). Page Composition says 告警规则 has "组件 3（只读态脱敏）" and 故障中心 has "组件 3（只读事件时间线）" — but no separate layouts are documented for these pages. A developer must infer which sections to render.

3. **Page-level state dispatch implicit.** The page has 4 states (正常/宽限/只读/403) but no unified page-level dispatch table states which component renders given `authHealth` + session state + `graceRemaining`.

4. **`graceRemaining=0` trigger for 宽限过期 → Component 3 transition is implied, not stated.** Component 2 States says "宽限过期 | 2 小时窗口到期，整页降级" but the exact condition (`graceRemaining === 0`) is not pinned.

5. **Component 4 polling source is "后端鉴权中间件" — not an API endpoint.** A developer implementing the 60s poll needs an HTTP endpoint, not a backend concept.

---

## Phase 2: Rubric Scoring (Verification Stance)

### Required Sections Check

| Section | Present? |
|---------|----------|
| Design System Reference | Yes |
| Component definitions (layout, states, interactions, data binding) | Yes (all 4 components) |
| States table (Default/Loading/Empty/Error) | Yes — Components 1-3 complete; Component 4 has Default + Recovery |
| Interactions table (Trigger/Action/Feedback) | Yes |
| Data Binding table (UI Element/Field/Source) | Yes |

No dimension receives 0 for missing required sections.

---

### Dimension 1: Requirement Coverage (250 pts) — Product Manager

#### UI function coverage (78/80)
All 3 PRD UI functions have corresponding components. Component 4 covers UF-1's 403 edge case (PRD flow step 5). Component 1 States explicitly clarifies "403 无权限…**不渲染在登录页**：渲染在受保护页「原地」的全宽占位（见组件 4）". The HTTP dispatch table now traces the 403 path: "eiam 认证成功但角色绑定缺失" → "原地渲染 403 占位". The eiamAuthUrl config-absent state is now covered.

Deduction (-2): The exact timing of the 403 check — whether it occurs during eiam callback processing or on first authenticated page load — is not explicitly stated. The HTTP dispatch table says "eiam 认证成功但角色绑定缺失" but doesn't specify whether this is a callback-time check or a middleware check on the first authenticated request. A PM tracing the full redirect→callback→role-check→in-place-render chain must infer this.

#### Navigation Architecture coverage (39/40)
3 primary entries covered in Navigation & App Shell: "品牌名「WatchAlert」+ 三个一级导航入口 `告警看板 / 告警规则 / 故障中心`". Secondary page (eiam 登录跳转页) covered: "由 UF-1 通过浏览器整页跳转实现（无站内二级路由）". Icon keywords mapped via Lucide (bell/shield/activity). Byte-identical nav requirement stated: "所有共享导航 HTML 在各页面必须字节一致". Responsive breakpoints defined for all components.

Deduction (-1): The navigation section frames the secondary page's return as "浏览器「返回」为回退机制" — this describes the abandonment path, not the success return. The success return (redirect_uri whitelist → re-entry to original page) is documented in Component 1 Interactions, not in the Navigation section. The navigation-level description of the secondary page is slightly incomplete.

#### State requirement coverage (80/80)
All 10 PRD-defined states addressed:
- UF-1 (未登录/登录中/登录失败/无权限): all 4 covered, plus a new "配置缺失" state.
- UF-2 (正常/宽限中/已恢复/宽限过期): all 4 covered.
- UF-3 (只读免登录/已恢复): both covered, plus Component 3 adds Loading/Populated/Empty/Error lifecycle states with explicit state-transfer note: "进入只读态 → Loading → (Populated | Empty | Error)".

Component 4 now has Default + Recovery (polling). No state gaps found.

#### Edge case handling (49/50)
Addressed boundary conditions:
- Long text overflow: "单行 `truncate` + `title` 全文；表格列设 `min-w` 防压缩".
- No data / empty: Component 3 Empty state with per-region empty placeholders.
- Slow network: "fetch 超过 5s 未返回时骨架上方追加 Small 文案「加载较慢，请稍候…」".
- Permission denied: Component 4 with recovery polling.
- Concurrent clicks: Component 1 Loading "锁定按钮，防重复点击".
- Open redirect: redirect_uri whitelist validation with scheme/blocklist.
- Hysteresis/debounce: "连续 3 次失败才判定不可用；连续 3 次成功才判恢复".
- Sensitive field masking: front-end secondary filter + backend non-delivery.
- SSE/polling transport failure: full table with reconnect/backoff + conservative degradation.
- Session expiry: 401 → redirect with return path + 150ms overlay.
- Multi-tab staleness: full table with staleness bounds + reconciliation.
- Pagination: "客户端分页（pageSize=20，符合 TECH-frontend-state-002）".
- 404 read-only source: "归入 Error 态，文案改为「只读数据暂不可用」".
- eiamAuthUrl config absent: now covered with disabled button + 30s retry.
- Component 4 role-later-assigned: now covered with 60s polling + manual refresh.

Not addressed (-1): Browser back button during Component 1 Loading — "不阻塞浏览器跳转" is stated, but what state the user sees if they hit back mid-redirect is unspecified.

**Dimension 1 total: 246/250**

---

### Dimension 2: User Experience (250 pts) — End User

#### Information hierarchy (77/80)
Login page: brand (H3 centered) → primary button (full-width) → guidance (muted centered) — clear. Banner: main text (font-medium, left) + sub text (Small, muted) + Badge (ml-auto) — good priority separation. Component 3: read-only notice bar (top context) → overview cards (3-col grid, primary scan) → rule list (full-width table) → timeline (max-w-[800px]). The ordering and max-widths imply visual priority.

Deduction (-3): No explicit statement of visual-weight rationale — the doc never says "概览卡片为首要扫描区" or explains why overview precedes rules precedes timeline. The structure implies it, but a user-scannability rationale is absent. The SSE transport table introduces 4 banner sub-cases with different texts; the relative visual weight of these variant banners is unaddressed (though they're all amber h-14, so weight is consistent — the concern is cognitive load, not visual weight).

#### Interaction intuitiveness (76/80)
Conventional patterns throughout: click button → redirect, banner → write disabled, read-only → no buttons, 403 → static. Auto-redirect (Component 3 已恢复) has a 150ms transition message per both States and Interactions tables — reconciled. Component 4 recovery has clear polling + feedback.

Deduction (-4): The SSE transport-failure table introduces a banner variant "鉴权服务连接中断，写操作已暂停" shown when "传输中断 > 60s 且上一状态为 normal" — but this is a transport failure, not necessarily an eiam failure. The message implies eiam itself is the problem, which may not be true. A user seeing "鉴权服务连接中断" when the actual issue is the SSE channel being down receives a potentially misleading message. Additionally, there are now 3 distinct banner text variants (normal grace, transport-interrupt-from-normal, transport-interrupt-from-degraded) that a user would experience without a clear mental model of when each appears.

#### Accessibility (87/90)
Comprehensive. Skip link ("跳到主内容"). Focus management (banner expand → focus to banner; 403 render → focus to title; Loading → focus on button with aria-busy). Keyboard nav (Tab-focusable, visible focus ring `outline: 2px solid var(--ring); outline-offset: 2px`, disabled = tabindex=-1 + aria-disabled). ARIA/live regions per component with explicit Error override ("Error 态为该规则的特例覆盖"). Semantic structure (`<nav aria-label>`, `<main id="main-content">`, `<aside>`, `<table>` with scope). Contrast: light muted-foreground = zinc-600 (#52525b) on muted = 5.74:1 (AA pass); dark muted-foreground = zinc-400 (#a1a1aa) on #27272a = 5.89:1; destructive dark = 5.91:1. `prefers-reduced-motion` downgrade specified. Dark mode precedence resolved (hc_theme > prefers-color-scheme).

Deduction (-3):
- No colorblind-specific accommodation note. Icons + text labels are present throughout (banner has alert-triangle + text, badges have text, status states have icons), which mitigates the concern, but the document never explicitly states "status is never conveyed by color alone" as a principle. A colorblind user auditor would flag this gap. (-2)
- Component 1 Loading keeps focus on the button while the button is set to `tabindex="-1"` + `aria-disabled="true"` per the disabled-element rule. A `tabindex="-1"` element is programmatically focusable, but this nuance is unexplained — an implementer might assume disabled elements can't hold focus. (-1)

**Dimension 2 total: 240/250**

---

### Dimension 3: Design Integrity (250 pts) — Designer

#### Design system adherence (78/80)
Shadcn well-defined: Light + Dark palettes with all token values, semantic color exceptions documented for both modes with specific amber/destructive values and contrast ratios, typography table, component specs (Button/Card/Card-compact/Badge/Table/Banner), Lucide icon system with size/strokeWidth rules, Motion table with easing + durations + reduced-motion + banner height technique, Do's/Don'ts. Dark mode precedence explicitly resolved.

Deduction (-2): The Do's and Don'ts mandates "disabled 用 opacity-50" but Component 1 Loading button — which is set to `aria-disabled="true"` + `aria-busy="true"` + `tabindex="-1"` — does not mention `opacity-50` in its visual spec. Whether the Loading button (which is "locked" not traditionally "disabled") should follow the opacity-50 rule is ambiguous. A designer would question whether the Loading visual (spinner + text change) is sufficient or if opacity-50 should also apply.

#### Visual coherence (86/90)
All components use the same zinc-neutral + semantic-exception tokens. Cross-page nav specified as byte-identical. Consistent radius pattern (rounded-md inputs/buttons, rounded-lg cards). Consistent border treatment (1px). Both full-width top bars have specs: banner (h-14, px-6, amber, 1px bottom border, rounded-0) and read-only notice bar (h-12, px-6, bg-muted, 1px bottom border). Compact Card variant documented as intentional.

Deduction (-4): Component 3's layout structure shows 概览 + 规则 + 时间线 for 告警看板. But the Page Composition table says 告警规则 has "组件 3（只读态脱敏）" and 故障中心 has "组件 3（只读事件时间线）" — no separate layouts are documented for these pages. A designer reviewing would not know whether the 告警规则 read-only view shows only the rule table (with notice bar) or also includes overview cards. The cross-page read-only layout coherence is unspecified.

#### State completeness (78/80)
- Component 1: Default / Loading / Error / 403 (via C4) / 配置缺失. Transitions: click → Loading → Default/Error; config-absent → 30s retry → Default. Complete.
- Component 2: 正常 / 宽限中 / 已恢复 / 宽限过期 with explicit transitions (3 failures → degraded, 3 successes → recovered, 2h timeout → expired). Complete.
- Component 3: Loading / Populated / Empty / Error / 只读免登录 / 已恢复 + explicit state-transfer note. Complete lifecycle.
- Component 4: Default / 角色已分配（恢复）with 60s polling + manual refresh. Complete.

Deduction (-2): The transition from Component 2 "宽限中" to Component 3 "只读免登录" (宽限过期) is described as "2 小时窗口到期，整页降级" but the exact trigger condition — `graceRemaining === 0` — is implied, not explicitly stated as a condition. A developer implementing the state machine must infer that when the backend-pushed `graceRemaining` reaches 0, the page switches from banner-mode to read-only-mode.

**Dimension 3 total: 242/250**

---

### Dimension 4: Implementability (250 pts) — Developer

#### Layout specificity (74/80)
- Component 1: specific — min-h-screen, max-w-[400px], p-6.
- Component 2: specific — h-14(56px), px-6, flex items-center gap-3.
- Component 3: specific — max-w-[1200px], grid grid-cols-3/gap-4 (≥1024), grid-cols-2 (768-1023), grid-cols-1 (<768); table column widths (min-w-[160px], w-[80px], w-[100px], min-w-[200px], w-[120px]); timeline max-w-[800px]; mt-6 spacing; overflow-x:auto.
- Component 4: max-w-[460px], p-6.
- Banner height animation: technique specified (scrollHeight + max-height).
- Responsive: full table covering all components × 3 breakpoints.

Deduction (-6):
- Page-specific read-only layouts for 告警规则 and 故障中心 are not documented — only the 告警看板 layout (概览 + 规则 + 时间线) is shown. A developer building the 告警规则 page's read-only view must infer which sections to include. (-3)
- Component 3 overview cards equal-height behavior unspecified — do cards in a row stretch to the tallest? `grid auto-rows`? (-1)
- Table row height / line-height unspecified. (-1)
- Component 2 banner: the responsive table says "主文案 + 副文案同一行" for ≥1024px, but the layout structure lists main + sub + badge as flex siblings with no `flex-wrap` or `nowrap` rule — does the sub-text wrap below on narrower desktop widths? (-1)

#### Data binding explicit (76/80)
All 4 components have Data Binding tables. Shared `AuthHealth` type contract documented with `recoverySignal` eliminated. eiamAuthUrl null case covered. No orphan interactive elements.

Deduction (-4):
- Component 4 `roleBinding` source is "后端鉴权中间件（`watchalert_role_binding` 表查询）" — this is a backend concept, not an API endpoint. The States table specifies "每 60s 轮询" which requires the developer to know what HTTP endpoint to poll. Other Data Binding entries (e.g., "后端 SSE/轮询", "后端只读接口") are also not endpoints, but they don't involve active polling — Component 4's polling interaction creates a higher bar for source specificity. (-2)
- Component 1 Error message "登录失败，请重试" not in Data Binding — unclear if static frontend copy or backend-sourced (BIZ-auth-006 implies static, but not stated). (-1)
- `graceRemaining: string` format: the multi-tab table says "后端统一下发绝对剩余时长，前端仅展示不下发本地计秒" implying a pre-formatted display string, but the exact format ("2小时15分" vs seconds vs ISO duration) is not stated in Data Binding. (-1)

#### Interaction unambiguity (84/90)
- Component 1: 4 triggers explicit. Config-absent behavior documented.
- Component 2: 4 triggers + 503 mechanism explained (401 correctly excluded) + SSE transport table + multi-tab staleness table.
- Component 3: 3 triggers, recovery reconciled with States table.
- Component 4: 2 triggers (403 render + polling recovery) explicit.

Deduction (-6):
- No unified page-level state dispatch table — which component renders on 告警看板 given `authHealth` + session state + `graceRemaining` is implicit. The Page Composition table lists which components are "involved" per page but doesn't specify the dispatch logic (e.g., "if authHealth=degraded && session valid && graceRemaining>0 → render Component 2 banner; if authHealth=degraded && graceRemaining=0 → render Component 3"). (-2)
- The `graceRemaining === 0` trigger for 宽限过期 → Component 3 transition is implied by "2 小时窗口到期" but not stated as an explicit condition. (-2)
- Component 3 recovery redirect does not specify the `redirect_uri` value — is it the current read-only page URL or the default 告警看板 path? Component 1 says "redirect_uri = 当前页 URL" but Component 3's recovery says only "前端自动整页跳转 eiam 登录页". (-1)
- Component 4 polling: the 60s interval and manual-refresh trigger are specified, but the polled resource is "后端鉴权中间件" — the interaction assumes a pollable endpoint that isn't specified. (-1)

**Dimension 4 total: 234/250**

---

## Cross-Dimension Coherence Check

1. **Recovery-path coherence is now well-documented.** Component 2 "已恢复" → banner collapses silently (logged-in, session valid). Component 3 "已恢复" → redirect to eiam login (anonymous read-only). These are different behaviors for the same `authHealth='recovered'` signal, which is correct given different contexts (logged-in vs anonymous). The shared type contract clarifies `recovered` as a transient state that triggers recovery then falls back to `normal`. This is coherent.

2. **HTTP error dispatch is now unified.** The "全局 HTTP 错误分派表" consolidates 401/403/503 with trigger scenarios, interceptor actions, and target components. The 503 note explicitly excludes cross-tab logout (401). 401 and 403 semantic separation is stated: "401 走跳转流程，403 走 Component 4 原地占位". This resolves the iteration-2 fragmentation.

3. **Page-level orchestration remains implicit.** While each component's states and interactions are well-specified, the page-level logic that decides which component to render (given the combination of authHealth, session validity, and graceRemaining) is never consolidated into a single dispatch table. A developer building the 告警看板 page must synthesize this from Component 2 States, Component 3 States, and Component 4 States. This is the primary remaining coherence gap.

---

## Phase 3: Blindspot Hunt

### [blindspot-1] Session expiry during grace period — interaction between session TTL and 2h grace window
> Component 2 States: "宽限中 | 黄色（amber）横幅展开 | 写操作按钮整体禁用"
> BIZ-auth-008: "会话过期 → 跳登录页（不白屏、不无限 loading）；重登后回原操作路径"

If the Redis session TTL is shorter than the 2-hour grace window, the user's session could expire while the grace banner is showing. The next request would return 401 → redirect to eiam login → but eiam is down (that's why grace is active) → the user cannot re-authenticate. The design does not address this interaction: it assumes the session remains valid throughout the 2-hour grace window, but doesn't state that the session TTL is extended or that 401 during grace is handled differently (e.g., fall through to Component 3 read-only instead of redirecting to an unavailable eiam). This is an operational edge case outside all rubric dimensions.

### [blindspot-2] Page-level state dispatch table absent
> Page Composition: "告警看板 | existing | 组件 2（横幅）+ 组件 3（只读）+ 组件 4（403）| 正常/宽限/只读/403 四态"

The page has 4 states, but no table specifies the dispatch logic: given `authHealth` (normal/degraded/recovered), session validity (valid/expired), and `graceRemaining` (>0/=0), which component renders? A developer must infer: `normal && session-valid` → no banner, normal content; `degraded && session-valid && graceRemaining>0` → Component 2 banner; `degraded && (graceRemaining=0 || !session)` → Component 3 read-only; `403` → Component 4. This inference is reasonable but undocumented. Outside the rubric (the rubric asks for component-level interactions, not page-level orchestration).

### [blindspot-3] Transition behavior for non-看板 pages during grace expiry
> Component 2 States: "宽限过期 | 横幅消失，页面切至只读免登录视图（组件 3） | 2 小时窗口到期，整页降级"
> Page Composition: "告警规则 | existing | 组件 2 + 组件 3（只读态脱敏）+ 组件 4"

When the grace period expires while the user is on the 告警规则 page (not 告警看板), does the page navigate to 告警看板 (Component 3's defined target) or does the current page switch to a read-only rendering of the rule list? The phrase "页面切至只读免登录视图（组件 3）" implies navigation to Component 3's defined target (告警看板), but the Page Composition table says 告警规则 also has Component 3 (只读态脱敏). Whether the user stays on 告警规则 with a read-only rule list or is redirected to 告警看板 is unspecified. Outside the rubric (cross-page transition behavior during state changes).

---

## Score Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| Requirement Coverage | 246 | 250 |
| User Experience | 240 | 250 |
| Design Integrity | 242 | 250 |
| Implementability | 234 | 250 |
| **Total** | **962** | **1000** |

### Score Breakdown by Criterion

| Dimension | Criterion | Score | Max |
|-----------|-----------|-------|-----|
| Requirement Coverage | UI function coverage | 78 | 80 |
| Requirement Coverage | Navigation Architecture coverage | 39 | 40 |
| Requirement Coverage | State requirement coverage | 80 | 80 |
| Requirement Coverage | Edge case handling | 49 | 50 |
| User Experience | Information hierarchy | 77 | 80 |
| User Experience | Interaction intuitiveness | 76 | 80 |
| User Experience | Accessibility | 87 | 90 |
| Design Integrity | Design system adherence | 78 | 80 |
| Design Integrity | Visual coherence | 86 | 90 |
| Design Integrity | State completeness | 78 | 80 |
| Implementability | Layout specificity | 74 | 80 |
| Implementability | Data binding explicit | 76 | 80 |
| Implementability | Interaction unambiguity | 84 | 90 |

**Movement vs iteration 2**: +141 (821 → 962). All three cross-section contradictions (-90) resolved. 12 of 15 prior issues addressed. The structural additions (unified HTTP dispatch table, multi-tab staleness table, Component 4 recovery, compact Card variant, dark mode precedence, banner animation technique, eiamAuthUrl config-absent state) close the major gaps. Remaining deductions are minor implementability and UX refinements — no contradictions, no missing required sections, no placeholder text.

---

## Attack Points (Prioritized)

1. **[Implementability]**: Page-specific read-only layouts for 告警规则 and 故障中心 are not documented — Component 3 layout shows "告警概览卡片 (grid) + 规则列表 (Table) + 事件时间线" for 告警看板 only, but Page Composition says 告警规则 has "组件 3（只读态脱敏）" and 故障中心 has "组件 3（只读事件时间线）" without separate layouts. — Must document per-page read-only layouts (which sections render, which are omitted) or state that all pages share the full 告警看板 layout with empty sections hidden.

2. **[Implementability]**: Page-level state dispatch logic is implicit — the Page Composition table says "正常/宽限/只读/403 四态" but no table specifies the dispatch condition (authHealth + session + graceRemaining → which component renders). — Must add a page-level state dispatch table pinning the conditions for each state.

3. **[User Experience]**: SSE transport-failure banner messaging is potentially misleading — "鉴权服务连接中断，写操作已暂停" is shown when "传输中断 > 60s 且上一状态为 normal", but the issue is the SSE channel, not necessarily eiam itself. 3 banner text variants exist without a clear user mental model. — Must either reword the transport-failure banner to distinguish transport issues from eiam health, or provide a user-facing explanation of the distinction.

4. **[Implementability]**: Component 4 polling source is "后端鉴权中间件（`watchalert_role_binding` 表查询）" — not an HTTP endpoint. The 60s polling interaction requires a pollable API endpoint. — Must specify the API endpoint or state that the endpoint contract is deferred to tech design.

5. **[Design Integrity]**: `graceRemaining === 0` trigger for 宽限过期 → Component 3 transition is implied — Component 2 States says "2 小时窗口到期，整页降级" but the exact condition is not stated. — Must explicitly state the trigger condition (e.g., "when backend-pushed `graceRemaining` reaches 0, transition from Component 2 to Component 3").

6. **[User Experience]**: No explicit colorblind accommodation note — icons + text labels mitigate but the principle "status is never conveyed by color alone" is never stated. — Must add an explicit statement that status indicators always include icon + text, never color alone.

7. **[Design Integrity]**: Component 1 Loading button opacity-50 ambiguity — Do's mandates "disabled 用 opacity-50" but Component 1 Loading visual spec doesn't mention opacity-50. — Must clarify whether the Loading button (locked, not traditionally disabled) follows the opacity-50 rule or is exempt.

8. **[Requirement Coverage]**: Back button during Component 1 Loading — "不阻塞浏览器跳转" is stated but back-button state during redirect is unspecified. — Must document what state the user sees if they hit back mid-redirect (Loading persists? Default?).

9. **[blindspot]**: Session expiry during grace period — if session TTL < 2h grace window, 401 → redirect to eiam → eiam is down → user cannot re-authenticate. — Must address whether 401 during grace falls through to Component 3 read-only instead of redirecting to an unavailable eiam.

10. **[blindspot]**: Transition behavior for non-看板 pages during grace expiry — does the page navigate to 告警看板 or stay on the current page with read-only rendering? — Must specify cross-page transition behavior during state changes.
