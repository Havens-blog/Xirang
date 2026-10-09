# UI Design Evaluation — Iteration 2

**Document**: `ui-design.md`
**Rubric**: `ui-web.md` (1000-point scale, target 950)
**PRD**: `prd-ui-functions.md`
**Scorer persona**: Senior UX Engineer
**Date**: 2026-10-09
**Previous iteration**: 747/1000

---

## Iteration-1 Issue Tracker

| Iteration-1 issue | Status in iteration 2 | Evidence |
|------------------|----------------------|----------|
| Component 3 missing Loading/Empty/Error | **Fixed** | States table now has Loading / Populated / Empty / Error + state-transfer note |
| Dark mode palette absent | **Fixed** | "Color Palette（Dark）" table with all token values + contrast verification |
| Accessibility underspecified (no ARIA/keyboard/focus) | **Fixed** | Full "Accessibility" section: skip link, focus management, keyboard nav, ARIA/live regions, semantic structure |
| Icon system missing | **Fixed** | "Icons" section: Lucide React + mapping table |
| Component 3 layout lacks dimensions | **Fixed** | max-w-[1200px], grid cols, gap-4, mt-6, column widths, overflow-x |
| SSE/polling transport failure | **Fixed** | "SSE/Polling 传输层故障处理" table |
| Contrast violation muted-on-muted | **Fixed** | muted-foreground raised to zinc-600 (#52525b), 5.74:1 |
| Cross-component 403 flow gap | **Partially fixed** | Component 1 States now says "不渲染在登录页：渲染在受保护页「原地」" — but redirect→403 sequence still implied |
| authHealth type inconsistency | **Fixed** | Shared `AuthHealth` enum contract, both components reference it |
| Component 3 information hierarchy | **Partially fixed** | Grid + ordering implies hierarchy, but no explicit visual-weight rationale |
| Auto-redirect without transition | **Fixed (then contradicted)** | States table adds 150ms transition message — BUT Interactions table still says "无提示停顿" (see Conflict 1) |
| 503 handling ambiguous | **Fixed (then partially broken)** | 503 捕获机制说明 added — BUT case (b) misclassifies a 401 path as 503 (see Conflict 3) |
| Responsive beyond nav | **Fixed** | "全局响应式行为" table covering all components × 3 breakpoints |
| Transition timing unspecified | **Fixed** | "Motion & Transitions" table |
| Session expiry (BIZ-auth-008) | **Fixed** | "会话过期处理" section, 401 → redirect with return path |
| eiamAuthUrl config absent | **Not fixed** | Still no handling for missing backend config |
| Back button during Component 1 Loading | **Not fixed** | Still unspecified |
| Component 4 transition-out | **Not fixed** | Still only Default state, no recovery mechanism |

Net: 13 of 17 prior issues resolved; 2 resolved-but-then-contradicted; 2 partially; 3 not addressed. The document is structurally much stronger, but three new cross-section contradictions were introduced or left standing.

---

## Phase 1: Reasoning Audit

### Problem → Solution
3 PRD UI functions → 4 components. UF-1 split across Component 1 (login entry) and Component 4 (403 in-place). UF-2 → Component 2. UF-3 → Component 3. Page Composition table extends coverage to 告警规则 and 故障中心. All UI functions mapped.

### Solution → Evidence
Layout ASCII trees per component. States / Interactions / Data Binding tables per component. Design System with Light + Dark palettes, typography, component specs, icons, motion, accessibility. Navigation & App Shell with shared nav + banner slot + responsive table. SSE transport-failure table. Session-expiry (401) handling. Shared `AuthHealth` type contract. This is structurally sound and substantially more complete than iteration 1.

### Self-contradiction / Pre-score Anchors (independent findings)

1. **Conflict 1 — Component 3 recovery: "150ms transition message" vs "无提示停顿".** The States table for "已恢复" specifies "跳转前 150ms 显示瞬时过渡文案「鉴权服务已恢复，正在跳转登录…」". The Interactions table for the same trigger specifies feedback "无按钮入口，无提示停顿". These directly contradict — one says show a 150ms message, the other says no pause. A developer cannot determine which is authoritative.

2. **Conflict 2 — Component 3 Error aria-live: polite vs assertive.** The global Accessibility section states "Component 3 数据加载/空/错误态切换：`role="status"` + `aria-live="polite"`". The Component 3 Error state row states "`role="alert"` + `aria-live="assertive"`". For the Error case, one spec says polite/status, the other says assertive/alert. No override relationship is stated.

3. **Conflict 3 — Component 2 503 note case (b) misclassifies a 401 path.** The 503 捕获机制说明 lists "(b) 已登录标签页在另一标签页登出后（BIZ-auth-005 会话销话）发起写请求" as a 503-triggering path and says both cases "均由全局 HTTP 拦截器统一捕获 503". But BIZ-auth-005 (injected business rule) explicitly states cross-tab logout causes the next request to return **401**, not 503. The doc's own 会话过期处理 section confirms 401 → login redirect. So case (b) returns 401, is handled by the 401 flow, and cannot also be a 503 path. The note conflates two distinct error codes.

4. **`recoverySignal` vs `authHealth='recovered'` redundancy.** Component 3 Data Binding lists `recoverySignal` as a separate field sourced from "后端 SSE/轮询", while the shared type contract already defines `recovered` as a value of `AuthHealth` (also sourced from the same backend probe). Whether `recoverySignal` is a separate channel or derived from `authHealth` is unspecified.

5. **Component 4 has no transition-out / recovery mechanism.** A user staring at the 403 placeholder after their role is assigned by an admin has no documented path back (refresh? poll? auto-detect?). The States table has only Default.

6. **`prefers-color-scheme` vs `[data-theme="dark"]` precedence unresolved.** Dark mode is enabled via "`[data-theme="dark"]` 或 `@media (prefers-color-scheme: dark)`" (OR). TECH-frontend-state-001 implies a manual theme toggle (`hc_theme`). If both selectors apply, a manual toggle cannot override an OS dark preference. Precedence is unspecified.

7. **Banner `height 0↔auto` is not directly CSS-animatable.** Motion table specifies "Banner 展开/收起（height 0↔auto）200ms / height, opacity". Animating to/from `height: auto` requires a technique (measured height, max-height, or FLIP); a developer cannot implement this with a plain CSS transition on `height`.

---

## Phase 2: Rubric Scoring (Verification Stance)

### Required Sections Check

| Section | Present? |
|---------|----------|
| Design System Reference | Yes |
| Component definitions (layout, states, interactions, data binding) | Yes (all 4 components) |
| States table (Default/Loading/Empty/Error) | Yes — Component 3 now covers all four; Component 4 N/A for Loading/Empty |
| Interactions table (Trigger/Action/Feedback) | Yes |
| Data Binding table (UI Element/Field/Source) | Yes |

No dimension receives 0 for missing required sections.

---

### Dimension 1: Requirement Coverage (250 pts) — Product Manager

#### UI function coverage (76/80)
All 3 PRD UI functions have corresponding components. Component 4 covers UF-1's 403 edge case (PRD flow step 5). Component 1 States now explicitly clarifies "403 无权限… **不渲染在登录页**：渲染在受保护页「原地」的全宽占位（见组件 4）", resolving the iteration-1 placement ambiguity.

Deduction (-4): The full redirect-then-403 sequence (eiam auth success → app receives callback → app checks `watchalert_role_binding` → no role → render 403 on the target protected page) is still implied, not traced as an explicit step list. A PM reviewing would still need to infer the callback→role-check→in-place-render chain.

#### Navigation Architecture coverage (38/40)
3 primary entries covered in Navigation & App Shell with "品牌名「WatchAlert」+ 三个一级导航入口". Secondary page (eiam 登录跳转页) covered: "由 UF-1 通过浏览器整页跳转实现（无站内二级路由），浏览器「返回」为回退机制". Icon keywords now mapped via Lucide (bell/shield/activity). Byte-identical nav requirement stated. Responsive breakpoints defined.

Deduction (-2): The secondary page's "Return Target" per PRD is "原访问页面" via redirect_uri whitelist — the design frames return as "浏览器「返回」为回退机制", which describes abandonment, not the success return path. The success-return mechanism (redirect_uri whitelist → re-entry to original page) lives in Component 1, not in the Navigation section, so the navigation-level description of the secondary page is slightly incomplete.

#### State requirement coverage (76/80)
All 10 PRD-defined states addressed:
- UF-1 (未登录/登录中/登录失败/无权限): all 4 covered.
- UF-2 (正常/宽限中/已恢复/宽限过期): all 4 covered.
- UF-3 (只读免登录/已恢复): both covered, plus Component 3 now adds Loading/Populated/Empty/Error lifecycle states with an explicit state-transfer note ("进入只读态 → Loading → (Populated | Empty | Error)").

Deduction (-4): Component 4 has only a Default state with no transition-out described.

#### Edge case handling (42/50)
Addressed boundary conditions:
- Long text overflow: "单行 `truncate` + `title` 全文；表格列设 `min-w` 防压缩".
- No data / empty: Component 3 Empty state with per-region empty placeholders.
- Slow network: "fetch 超过 5s 未返回时骨架上方追加 Small 文案「加载较慢，请稍候…」".
- Permission denied: Component 4.
- Concurrent clicks: Component 1 Loading "锁定按钮，防重复点击".
- Open redirect: redirect_uri whitelist validation.
- Hysteresis/debounce: "连续 3 次失败才判定不可用；连续 3 次成功才判恢复".
- Sensitive field masking: front-end secondary filter for known sensitive field names.
- SSE/polling transport failure: full table with reconnect/backoff.
- Session expiry: 401 → redirect with return path.
- Pagination: "客户端分页（pageSize=20，符合 TECH-frontend-state-002）".
- 404 read-only source: "归入 Error 态，文案改为「只读数据暂不可用」".

Not addressed (-8):
- `eiamAuthUrl` config absent (Component 1): Data Binding says it comes from "后端配置下发" but no behavior specified if config is missing.
- Browser back button during Component 1 Loading: "不阻塞浏览器跳转" — but back-button state during redirect still unspecified.
- Component 4 role-later-assigned recovery: no auto-refresh/poll mechanism.

**Subtotal: 232/250**

#### Cross-section inconsistency (Conflict 3): -30
Component 2's 503 note case (b) states "已登录标签页在另一标签页登出后（BIZ-auth-005 会话销毁）发起写请求" is a 503 path, contradicting BIZ-auth-005 (which returns 401 for cross-tab logout) and the doc's own 会话过期处理 section. This is a business-rule alignment error in the requirements-coverage domain.

**Dimension 1 total: 202/250**

---

### Dimension 2: User Experience (250 pts) — End User

#### Information hierarchy (70/80)
Login page: brand (H3 centered) → primary button (full-width) → guidance (muted centered) — clear. Banner: main text (font-medium, left) + sub text (Small, muted) + Badge (ml-auto) — good priority separation. Component 3 now has an explicit structure: read-only notice bar (top context) → overview cards (3-col grid, primary scan) → rule list (full-width table) → timeline (max-w-[800px]). The ordering and max-widths imply visual priority.

Deduction (-10): No explicit statement of visual-weight rationale — the doc never says "概览卡片为首要扫描区" or explains why overview precedes rules precedes timeline. The structure implies it, but a user-scannability rationale is absent. Also the read-only notice bar (h-12, muted) and the grace banner (h-14, amber) are both full-width top bars at different vertical positions; their relative visual weight and whether the notice bar competes with the banner for attention is unaddressed (though they don't co-occur — banner is logged-in grace, notice bar is read-only — so this is minor).

#### Interaction intuitiveness (74/80)
Conventional patterns throughout: click button → redirect, banner → write disabled, read-only → no buttons, 403 → static. Auto-redirect (Component 3 已恢复) now has a 150ms transition message per the States table — improving on iteration 1's abrupt redirect.

Deduction (-6): The 503 capture has two paths (button-intercept + HTTP-interceptor) plus the SSE transport-failure conservative-degradation banner ("鉴权服务连接中断，写操作已暂停") which appears while "假定 eiam 健康不变" — a banner saying "connection interrupted" while assuming health is unchanged is a potentially confusing message for end users. And the SSE transport table introduces 4 sub-cases that a user would experience as different banner texts without a clear mental model of when each appears.

#### Accessibility (84/90)
Major improvement over iteration 1. Full Accessibility section: skip link ("跳到主内容"), focus management (focus moves to banner/403 title; Loading keeps focus on button with aria-busy), keyboard nav (Tab-focusable, visible focus ring `outline: 2px solid var(--ring)`, disabled = tabindex=-1 + aria-disabled), ARIA/live regions per component, semantic structure (`<nav aria-label>`, `<main id="main-content">`, `<aside>`, `<table>` with scope). Contrast fixed: muted-foreground = zinc-600 (#52525b) on muted = 5.74:1 (passes AA). Dark mode contrast verified ≥ 4.5:1. `prefers-reduced-motion` downgrade specified.

Deduction (-6):
- Component 1 Loading keeps focus on the button while the button is set to `tabindex="-1"` + `aria-disabled="true"` per the disabled-element rule. A programmatically-focused tabindex=-1 element works, but this nuance is unexplained — an implementer might assume disabled elements can't hold focus.
- No colorblind-specific accommodation note (semantic icons + text labels are present, which mitigates, but no explicit statement that status is never conveyed by color alone).

**Subtotal: 228/250**

#### Cross-section inconsistency (Conflict 2): -30
The global Accessibility rule "Component 3 数据加载/空/错误态切换：`role="status"` + `aria-live="polite"`" contradicts Component 3 Error state "`role="alert"` + `aria-live="assertive"`" for the same Error case. This is directly an accessibility-spec contradiction — a screen reader user's error-announcement behavior is ambiguous. No specific-over-general override is stated.

**Dimension 2 total: 198/250**

---

### Dimension 3: Design Integrity (250 pts) — Designer

#### Design system adherence (76/80)
Shadcn well-defined: Light + Dark palettes with all token values, semantic color exceptions documented for both modes with specific amber/destructive values and contrast ratios, typography table, component specs (Button/Card/Badge/Table/Banner), Lucide icon system with size/strokeWidth rules, Motion table with easing + durations + reduced-motion, Do's/Don'ts.

Deduction (-4):
- Card component spec defines padding `p-6`, but Component 3 overview cards use "Card（rounded-lg, p-4, 1px border）" — a deviation with no rationale. A designer would question whether overview cards are a denser variant or an error.
- Banner motion "height 0↔auto 200ms" — `height: auto` is not a CSS-animatable value; the motion spec is internally consistent but not directly implementable as written (see Implementability).

#### Visual coherence (83/90)
All components use the same zinc-neutral + semantic-exception tokens. Cross-page nav specified as byte-identical. Consistent radius pattern (rounded-md inputs/buttons, rounded-lg cards). Consistent border treatment (1px). Both full-width top bars now have specs: banner (h-14, px-6, amber, 1px bottom border, rounded-0) and read-only notice bar (h-12, px-6, bg-muted).

Deduction (-7):
- Read-only notice bar (Component 3) spec says "bg-muted、h-12、px-6" with no border mentioned, while the Do's and Don'ts mandates "全部 1px 细边框". Whether the notice bar has a border is ambiguous — if the rule is "all 1px borders", the notice bar should have one, but it's not stated. Visual coherence with the amber banner (which has an explicit border) is uncertain.
- Card padding inconsistency (p-6 global vs p-4 in Component 3 overview) — see above; coheres poorly with the global Card spec.

#### State completeness (74/80)
- Component 1: Default / Loading / Error / 403 (via C4). Transitions: click → Loading → Default/Error. Complete.
- Component 2: 正常 / 宽限中 / 已恢复 / 宽限过期 with explicit transitions (3 failures → degraded, 3 successes → recovered, 2h timeout → expired). Complete.
- Component 3: Loading / Populated / Empty / Error / 只读免登录 / 已恢复 + explicit state-transfer note. Complete lifecycle.
- Component 4: Default only.

Deduction (-6): Component 4 has only Default with no transition-out (what happens if the user's role is assigned while viewing the 403, or if they refresh). State transitions for Components 1-3 are well-described.

**Dimension 3 total: 233/250**

---

### Dimension 4: Implementability (250 pts) — Developer

#### Layout specificity (70/80)
- Component 1: specific — min-h-screen, max-w-[400px], p-6.
- Component 2: now specific — h-14(56px), px-6, flex items-center gap-3.
- Component 3: now specific — max-w-[1200px], grid grid-cols-3/gap-4 (≥1024), grid-cols-2 (768-1023), grid-cols-1 (<768); table column widths (min-w-[160px], w-[80px], w-[100px], min-w-[200px], w-[120px]); timeline max-w-[800px]; mt-6 spacing; overflow-x:auto.
- Component 4: max-w-[460px], p-6.
- Dark palette: full token values provided.
- Responsive: full table.

Deduction (-10):
- Banner "height 0↔auto 200ms" is not directly CSS-animatable — developer must choose a technique (measured height, max-height, FLIP) with no guidance (-5).
- Component 3 overview cards: equal-height behavior unspecified (do cards in a row stretch to the tallest? grid auto-rows?) (-2).
- Table row height / line-height unspecified (-1).
- Component 2 banner sub-text and Badge vertical alignment within the h-14 bar (are they centered together? does sub-text wrap below main text on ≥1024px "同一行"?) — the responsive table says "同一行" for ≥1024px but the layout structure lists them as flex siblings with no wrap rule (-2).

#### Data binding explicit (72/80)
All 4 components have Data Binding tables. Shared `AuthHealth` type contract documented. No orphan interactive elements.

Deduction (-8):
- Component 1 Error message "登录失败，请重试" not in Data Binding — unclear if static frontend copy or backend-sourced (iteration-1 issue, unresolved).
- Component 2 Badge「只读」mentioned in layout ("Badge「只读」(secondary variant，ml-auto) // writeAllowed=false 时显示") but not in Data Binding table — visibility condition implied from `writeAllowed` but not explicitly mapped.
- `recoverySignal` (Component 3) vs `authHealth='recovered'` (shared contract) — two fields for the same conceptual signal; relationship unspecified (-3).

#### Interaction unambiguity (76/90)
- Component 1: 4 triggers explicit.
- Component 2: 4 triggers + 503 mechanism explained + SSE transport table — mostly explicit and improved over iteration 1.
- Component 3: 3 triggers — BUT the "已恢复" row says "无按钮入口，无提示停顿" while the States table says "跳转前 150ms 显示瞬时过渡文案" (Conflict 1).
- Component 4: 1 trigger explicit, no transition-out.

Deduction (-14): Component 3 已恢复 interaction contradicts its States table (treated as cross-section -30 below); `eiamAuthUrl`-absent scenario unhandled (-3); Component 4 no transition-out interaction (-3); `recoverySignal` vs `authHealth` recovery-channel ambiguity (-3); remaining minor (-5).

**Subtotal: 218/250**

#### Cross-section inconsistency (Conflict 1): -30
Component 3 States "已恢复" specifies "跳转前 150ms 显示瞬时过渡文案「鉴权服务已恢复，正在跳转登录…」", while Component 3 Interactions for the same trigger specifies feedback "无按钮入口，无提示停顿". A developer implementing the recovery flow cannot determine whether to render the 150ms transition message or not. This is a direct interaction-table-vs-state-table contradiction.

**Dimension 4 total: 188/250**

---

## Cross-Dimension Coherence Check

1. **401 / 403 / 503 dispatch model is fragmented.** Three HTTP error codes are handled in three separate places: 401 (会话过期处理), 403 (Component 4), 503 (Component 2 503 note + SSE table). There is no unified HTTP-interceptor dispatch table showing how a single response is routed. Conflict 3 (503 note case (b) misclassifying a 401 path) is a symptom of this fragmentation. A developer must synthesize the global interceptor from three sections.

2. **Recovery-path coherence.** Component 2 "已恢复" → banner collapses silently, user continues (logged-in, session valid). Component 3 "已恢复" → redirect to eiam login (anonymous read-only). These are different behaviors for the same `authHealth='recovered'` signal, which is correct given different contexts (logged-in vs anonymous), but the full state cycle (normal → degraded → read-only → login → normal) is implicit, not documented as a single flow diagram.

3. **`recoverySignal` vs `authHealth='recovered'`.** Component 3 Data Binding lists both as separate fields from the same backend probe source. Whether they are two transport channels or one derived from the other is unspecified — a developer must decide the data contract.

---

## Phase 3: Blindspot Hunt

### [blindspot-1] `recoverySignal` vs `authHealth='recovered'` data-contract ambiguity
> Component 3 Data Binding: "恢复信号 | `recoverySignal` | 后端 SSE/轮询"
> Shared type contract: "`recovered` = 探针连续 3 次成功判定恢复（瞬时态，前端收到后触发恢复流程，随后回落 `normal`）"

The shared `AuthHealth` enum already carries `recovered` from the same backend probe. Component 3 also binds a separate `recoverySignal` field. The relationship between these two is unspecified — is `recoverySignal` a distinct SSE event channel, or is it just `authHealth === 'recovered'`? A developer designing the store cannot determine whether to subscribe to one signal or two. Outside all rubric dimensions (data-contract design ambiguity, not orphan-element checking).

### [blindspot-2] Multi-tab state staleness not modeled (BIZ-auth-005)
> BIZ-auth-005: "多标签页不做实时广播（不引入 WebSocket / BroadcastChannel）"
> Component 2 Data Binding: "`authHealth: AuthHealth` | 后端 SSE/轮询（10s 探针，滞回 3 次）"

BIZ-auth-005 explicitly forbids real-time cross-tab broadcast. This means Tab B's `authHealth` / `writeAllowed` / grace-banner state is stale until its own next probe tick. The design does not specify: (a) how long Tab B's banner can be stale relative to Tab A's logout/recovery; (b) whether Tab B's grace countdown drifts from Tab A's; (c) what Tab B shows if eiam recovered in Tab A but Tab B's probe hasn't yet confirmed. This cross-tab staleness window is an operational UX gap outside the rubric.

### [blindspot-3] Component 4 (403) has no recovery / refresh mechanism
> Component 4 States: only "Default（403） | 居中 403 占位卡 | 原地停留，不跳登录页，不提供「重新登录」按钮"

If an administrator assigns a `watchalert_role_binding` to the user while they are viewing the 403 placeholder, there is no documented mechanism (auto-poll, manual refresh hint, SSE signal) for the user to recover. The user must manually refresh, which is undocumented. A static 403 with no recovery path strands the user. Outside the rubric (state-transition-out, but the rubric's state-completeness criterion doesn't explicitly require recovery-from-terminal-state).

### [blindspot-4] `prefers-color-scheme` vs `[data-theme="dark"]` precedence unresolved
> "暗色模式通过 `[data-theme="dark"]` 或 `@media (prefers-color-scheme: dark)` 重写同名 CSS 变量"

TECH-frontend-state-001 implies a persisted manual theme toggle (`hc_theme`). Using OR (`或`) between a manual attribute selector and an OS media query means a user with OS dark preference cannot be overridden by a manual light toggle (the media-query branch still applies). The design does not specify precedence (manual overrides auto, or vice versa) or how `hc_theme` interacts with `prefers-color-scheme`. This is a theme-system design gap contradicting the frontend-state convention, outside the rubric.

---

## Score Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| Requirement Coverage | 202 | 250 |
| User Experience | 198 | 250 |
| Design Integrity | 233 | 250 |
| Implementability | 188 | 250 |
| **Total** | **821** | **1000** |

### Score Breakdown by Criterion

| Dimension | Criterion | Score | Max |
|-----------|-----------|-------|-----|
| Requirement Coverage | UI function coverage | 76 | 80 |
| Requirement Coverage | Navigation Architecture coverage | 38 | 40 |
| Requirement Coverage | State requirement coverage | 76 | 80 |
| Requirement Coverage | Edge case handling | 42 | 50 |
| Requirement Coverage | *Cross-section -30 (Conflict 3: 503 vs BIZ-auth-005)* | -30 | — |
| User Experience | Information hierarchy | 70 | 80 |
| User Experience | Interaction intuitiveness | 74 | 80 |
| User Experience | Accessibility | 84 | 90 |
| User Experience | *Cross-section -30 (Conflict 2: aria-live polite vs assertive)* | -30 | — |
| Design Integrity | Design system adherence | 76 | 80 |
| Design Integrity | Visual coherence | 83 | 90 |
| Design Integrity | State completeness | 74 | 80 |
| Implementability | Layout specificity | 70 | 80 |
| Implementability | Data binding explicit | 72 | 80 |
| Implementability | Interaction unambiguity | 76 | 90 |
| Implementability | *Cross-section -30 (Conflict 1: 150ms message vs 无提示停顿)* | -30 | — |

**Movement vs iteration 1**: +74 (747 → 821). The structural improvements (dark palette, accessibility, icons, Component 3 lifecycle, SSE transport, responsive, motion, session expiry, type unification) recovered ~140 points, but three newly-surfaced or left-standing cross-section contradictions (-90) cap the gain.

---

## Attack Points (Prioritized)

1. **[Implementability]**: Component 3 recovery interaction contradicts its own States table — States says "跳转前 150ms 显示瞬时过渡文案「鉴权服务已恢复，正在跳转登录…」", Interactions says "无按钮入口，无提示停顿". — Must reconcile: either the 150ms transition message exists (update Interactions) or it doesn't (update States). Cannot be both.

2. **[User Experience]**: Component 3 Error aria-live spec contradicts the global Accessibility rule — global says "数据加载/空/错误态切换：`role="status"` + `aria-live="polite"`", Error state says "`role="alert"` + `aria-live="assertive`". — Must state explicitly whether Error overrides the global polite rule, and unify the two specs.

3. **[Requirement Coverage]**: Component 2 503 note case (b) misclassifies a 401 path as 503 — "已登录标签页在另一标签页登出后（BIZ-auth-005 会话销毁）发起写请求…均由全局 HTTP 拦截器统一捕获 503" contradicts BIZ-auth-005 (cross-tab logout → 401) and the doc's own 会话过期处理 section. — Must remove case (b) from the 503 mechanism (it is a 401 path) and add a unified 401/403/503 dispatch table.

4. **[Implementability]**: `recoverySignal` vs `authHealth='recovered'` data-contract redundancy — both sourced from "后端 SSE/轮询", relationship unspecified. — Must clarify whether `recoverySignal` is a distinct channel or derived from `authHealth`, and collapse to one if redundant.

5. **[Design Integrity]**: Card padding inconsistency — global Card spec is `p-6`, Component 3 overview cards use `p-4` with no rationale. — Must either align overview cards to p-6 or explicitly document a denser "overview card" variant.

6. **[Design Integrity]**: Read-only notice bar border ambiguous — Do's and Don'ts mandates "全部 1px 细边框" but notice bar spec ("bg-muted、h-12、px-6") omits any border. — Must state explicitly whether the notice bar has a border (and its color/radius).

7. **[Implementability]**: Banner `height 0↔auto` 200ms animation not directly CSS-animatable. — Must specify the technique (measured height via ResizeObserver, max-height, or FLIP) or change the animated property.

8. **[Requirement Coverage]**: Component 4 has no transition-out / recovery mechanism — only Default state, no path back after role assignment. — Must specify refresh/poll behavior or document that manual refresh is the only recovery.

9. **[blindspot]**: Multi-tab state staleness (BIZ-auth-005 forbids broadcast) — Tab B's `authHealth`/banner/grace countdown can drift from Tab A's with no staleness window or reconciliation specified. — Must document the cross-tab staleness bound and any reconciliation.

10. **[blindspot]**: `prefers-color-scheme` vs `[data-theme="dark"]` precedence unresolved — OR semantics prevents a manual toggle from overriding OS dark preference, conflicting with TECH-frontend-state-001 (`hc_theme`). — Must specify precedence (manual > auto) and the toggle interaction.

11. **[blindspot]**: 401/403/503 dispatch model fragmented across three sections with no unified interceptor table — a developer must synthesize global HTTP error routing from 会话过期处理 (401), Component 4 (403), and Component 2 503 note (503). — Must add a single HTTP-error dispatch table.

12. **[blindspot]**: `eiamAuthUrl` config-absent scenario unhandled (Component 1) — Data Binding says "后端配置下发" but no behavior if config missing. — Must specify fallback (e.g., render an error state instead of a dead login button).
