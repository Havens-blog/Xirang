# UI Design Evaluation — Iteration 1

**Document**: `ui-design.md`
**Rubric**: `ui-web.md` (1000-point scale, target 950)
**PRD**: `prd-ui-functions.md`
**Scorer persona**: Senior UX Engineer
**Date**: 2026-10-09

---

## Phase 1: Reasoning Audit

### Problem → Solution
The PRD defines 3 UI functions: UF-1 (eiam unified login), UF-2 (grace banner), UF-3 (read-only view). The design provides 4 components mapping to these: Component 1 → UF-1, Component 2 → UF-2, Component 3 → UF-3, Component 4 → UF-1's 403 in-place state. All UI functions are covered. The design also adds a Page Composition table expanding coverage to 告警规则 and 故障中心 pages, aligning with PRD Navigation Architecture's 3 primary pages.

### Solution → Evidence
Layout structures are provided as ASCII tree diagrams for each component. States, Interactions, and Data Binding tables exist per component. A Design System section defines color palette, typography, and component specs. Navigation & App Shell defines the shared top nav bar. This is structurally sound.

### Self-contradiction / Pre-score Anchors
1. **Component 3 state gap**: Component 3 fetches `readonlyAlerts: list` from a backend read-only API but defines only 2 states (只读免登录 / 已恢复). No Loading, Empty, or Error state for the data-fetching lifecycle. A read-only dashboard that cannot show "loading data" or "no alerts" or "fetch failed" is incomplete.
2. **Dark mode palette absent**: The document claims "全部通过 CSS 变量实现 light/dark 切换" but only provides the Light color palette table. No dark mode token values are given anywhere.
3. **Accessibility underspecified**: No ARIA labels, roles, live regions, keyboard navigation, or focus management specified anywhere in the document.
4. **Icon system missing**: PRD Navigation Architecture defines icon keywords (告警/alert, 规则/rule, 故障/event) but the design's Navigation & App Shell section does not specify any icon library, icon set, or visual treatment for navigation icons.
5. **Component 3 layout lacks dimensions**: "告警概览卡片 / 规则列表 / 事件时间线" listed without widths, grid columns, gap, or max-width.
6. **SSE/polling connection failure**: "后端 SSE/轮询（10s 探针，滞回 3 次）" — the design specifies eiam health probe failure behavior but not what happens if the SSE/polling transport itself drops independently.
7. **Contrast violation**: muted-foreground (#71717a) on muted background (#f4f4f5) yields 4.4:1 — below WCAG AA 4.5:1 for normal text (14px, 400 weight).
8. **Cross-component 403 flow**: Component 1 says eiam auth success but no role → render Component 4. Component 4 renders on "受保护页面". But the user is on the login page after auth — the redirect from login → protected page → 403 rendering transition is not explicitly described.

---

## Phase 2: Rubric Scoring

### Required Sections Check

| Section | Present? |
|---------|----------|
| Design System Reference | Yes |
| Component definitions (layout, states, interactions, data binding) | Yes (all 4 components) |
| States table (Default/Loading/Empty/Error) | Partially — tables exist but not all components cover all 4 states |
| Interactions table (Trigger/Action/Feedback) | Yes |
| Data Binding table (UI Element/Field/Source) | Yes |

No dimension receives 0 for missing required sections.

---

### Dimension 1: Requirement Coverage (250 pts) — Product Manager

#### UI function coverage (76/80)
All 3 PRD UI functions have corresponding components. Component 4 explicitly covers UF-1's 403 edge case (PRD flow step 5: "eiam 认证成功但无 WatchAlert 内部角色 → 返回 403"). No PRD UI function is unaddressed.

Deduction (-4): UF-1 is split across Component 1 (login entry) and Component 4 (403 placeholder) without an explicit cross-component flow description. The transition from login page → eiam auth success → redirect to protected page → 403 rendering is implied but not stated. A PM reviewing this would ask "what exactly happens between auth success and the 403 appearing?"

#### Navigation Architecture coverage (32/40)
Primary navigation (3 entries: 告警看板 / 告警规则 / 故障中心) is covered in Navigation & App Shell with "品牌名「WatchAlert」+ 三个一级导航入口". Secondary page (eiam 登录跳转页) is covered: "eiam 登录跳转页 由 UF-1 通过浏览器整页跳转实现". Navigation targets are consistent with PRD page names. The design also notes "原型中所有共享导航 HTML 在各页面必须字节一致（仅 active 态不同），标签与目标严格对齐 PRD Navigation Architecture 表".

Deduction (-8): PRD Navigation Architecture defines an "Icon Keyword" column (告警/alert, 规则/rule, 故障/event) but the design does not specify any icon system, icon library, or visual treatment for these navigation icons. The nav bar description omits icons entirely.

#### State requirement coverage (62/80)
PRD state coverage:
- UF-1 PRD states (未登录, 登录中, 登录失败, 无权限): all 4 covered — Component 1 States table has Default(未登录), Loading(跳转中), Error(登录失败); Component 4 covers 403(无权限).
- UF-2 PRD states (正常, 宽限中, 已恢复, 宽限过期): all 4 covered in Component 2 States table.
- UF-3 PRD states (只读免登录, 已恢复): both covered in Component 3 States table.

All 10 PRD-defined states are addressed.

Deduction (-18): The rubric asks about loading/empty/error/populated states generally. Component 3 is a data-display component that fetches `readonlyAlerts: list` from a backend API, but its States table only has 2 entries (只读免登录 / 已恢复). There is no Loading state (data still fetching), no Empty state (no alerts to display), and no Error state (read-only API itself fails). These are applicable states for a dashboard component that the design does not address.

#### Edge case handling (32/50)
Addressed boundary conditions:
- Permission denied: Component 4 (403 placeholder).
- Concurrent clicks: Component 1 Loading state "锁定按钮，防重复点击".
- Open redirect: redirect_uri whitelist validation ("仅 host 匹配告警平台域名、path 以 / 开头的相对路径；禁止绝对外链、// 开头协议相对 URL、javascript:/data: scheme").
- Hysteresis/debounce: "连续 3 次失败才判定不可用；连续 3 次成功才判恢复".
- Sensitive field masking: "通知渠道凭据（webhook secret、API key）、数据源密码、webhook URL 查询参数等一律 *** 不展示".

Not addressed:
- Long text overflow: banner text "剩余 X 小时 X 分；此期间告警规则/通知等写操作不可用" could overflow on narrow screens — no truncation, wrap, or responsive text rules.
- No data / empty state: Component 3's read-only dashboard has no empty state for when `readonlyAlerts` list is empty.
- Slow network for Component 3 data fetch: no skeleton/spinner/loading state for read-only data.
- SSE/polling connection failure: "后端 SSE/轮询（10s 探针，滞回 3 次）" — no behavior specified for when the SSE/polling transport itself drops.
- Browser back button during loading: Component 1 Loading state doesn't address browser back navigation.

Deduction (-18): Five edge case categories not addressed.

**Dimension 1 total: 202/250**

---

### Dimension 2: User Experience (250 pts) — End User

#### Information hierarchy (58/80)
Login page (Component 1) has clear hierarchy: brand name (H3, centered) → primary button (full-width) → guidance text (Muted, centered). Banner (Component 2) has main text (left-aligned, font-medium) + sub text (Small, muted) — good priority separation. 403 card (Component 4) has title → description, clear.

Deduction (-22): Component 3's read-only view lists "告警概览卡片 / 规则列表 (Table) / 事件时间线" in its layout structure without specifying visual priority, ordering rationale, or relative prominence. A user scanning the read-only dashboard cannot tell from the spec what matters most — are alert cards more prominent than the event timeline? Is the rule list the primary content? The design doesn't guide visual weight distribution for this multi-section page.

#### Interaction intuitiveness (68/80)
Interaction patterns are conventional: click button → browser redirect (Component 1), banner appears → write buttons disabled (Component 2), read-only view → no buttons (Component 3), 403 → static placeholder (Component 4). No custom interaction behaviors that require learning.

Deduction (-12): Component 3's "已恢复" state triggers "前端自动整页跳转 eiam 登录页" — "无按钮入口，无提示停顿". An automatic page redirect without any user action or transition indication could disorient users. The design explicitly states "无提示停顿" (no prompt pause), which means the page changes suddenly. While justified by the PRD requirement to avoid re-introducing a login button, the UX of an unannounced redirect is poor — at minimum a brief "鉴权服务已恢复，正在跳转登录…" transition state would improve usability without violating the constraint.

#### Accessibility (38/90)
**Contrast**: Calculated WCAG contrast ratios:
- muted-foreground (#71717a) on muted background (#f4f4f5): **4.4:1** — FAILS WCAG AA (requires 4.5:1 for normal text at 14px/400 weight). This affects table headers ("表头 bg-muted") and Component 3's read-only notice bar ("muted 底").
- muted-foreground on white (#ffffff): 4.83:1 — passes but barely.
- muted-foreground on amber-50 (#fffbeb): 4.66:1 — passes but barely.
- amber-800 on amber-50: 6.84:1 — passes.
- foreground on background: 19.9:1 — passes.
- primary-foreground on primary: 16.97:1 — passes.

**ARIA/roles**: No ARIA labels, roles, or live regions specified anywhere. The grace banner (Component 2) is a status notification that should use `role="alert"` or `aria-live="polite"` — not specified. The loading spinner in Component 1 ("内联 spinner") has no text alternative or `aria-label`. The 403 state change (Component 4 rendering) has no `aria-live` announcement.

**Keyboard navigation**: Not mentioned at all. No focus management for state transitions — when the banner appears, when 403 renders, when auto-redirect fires. No specification of focus movement or keyboard operability for any component.

**Other**: No skip navigation link specified. No screen reader behavior for async state changes. No `prefers-reduced-motion` consideration for the 150ms hover transition or banner expand/collapse.

Deductions: no ARIA/roles/live regions (-20), no keyboard/focus management (-15), contrast failure on muted-on-muted (-10), no skip link/reduced-motion/other (-7).

**Dimension 2 total: 164/250**

---

### Dimension 3: Design Integrity (250 pts) — Designer

#### Design system adherence (68/80)
Shadcn system is well-defined: specific CSS variable tokens for all color roles, typography table with font/weight/size, component specs (Button with rounded-md/h-10/px-4, Card with rounded-lg/p-6, Badge, Table, Banner). Semantic color exception is thoroughly documented with specific amber values ("amber-50 底 / amber-800 文字 / amber-600 1px 边"). Do's and Don'ts table reinforces rules.

Deduction (-12): 
1. Dark mode palette values absent — the document claims "全部通过 CSS 变量实现 light/dark 切换" and "默认浅色，支持暗色切换" but only provides the Light color palette table. No dark mode token values are given. A designer reviewing this cannot verify dark mode coherence.
2. No icon system specified — PRD defines icon keywords but the design doesn't specify an icon library (Lucide? Heroicons?) or icon visual treatment, despite the banner layout mentioning "图标" ("图标 + 主文案").

#### Visual coherence (72/90)
All components use the same design tokens (zinc neutrals + semantic exceptions). Cross-page navigation is specified as "字节一致" (byte-identical) across pages. Consistent border-radius pattern (rounded-md for inputs/buttons, rounded-lg for cards). Consistent shadow depth (none — "无阴影" for Card, "无阴影" implied throughout). Consistent border treatment (1px solid).

Deduction (-18): Component 3's read-only notice bar uses "muted 底" with no border or radius specification. This is visually ambiguous relative to Component 2's amber banner which has explicit specs ("1px 底边（amber-600），圆角 0（贴顶）"). Are they the same visual pattern (full-width banner) or different? The read-only notice lacks border, radius, padding, and height specs — a designer cannot determine if it's visually coherent with the grace banner or a different pattern. Additionally, Component 3's "告警概览卡片" — is this a standard Card (rounded-lg, p-6, 1px border) or a modified read-only variant? Not specified.

#### State completeness (52/80)
- Component 1: Default, Loading, Error, 403 (via Component 4). State transitions described: click → Loading → Default/Error. Empty state N/A for login. Adequate.
- Component 2: 4 states (正常 / 宽限中 / 已恢复 / 宽限过期) with transitions described (3 failures → degraded, 3 successes → recovered, 2h timeout → expired). Good — state transitions are explicit.
- Component 3: Only 2 states (只读免登录 / 已恢复). Missing Loading (data fetching), Empty (no alerts), Error (read-only API failure). This is a significant gap — a dashboard component displaying `readonlyAlerts: list` must handle the data-fetching lifecycle. How does the component go from entering the read-only state to showing populated data? What shows while fetching? What if the fetch fails?
- Component 4: Only Default (403). Acceptable for a static placeholder, but no transition out is described (what happens if the user's role is later assigned?).

Deduction (-28): Primarily for Component 3 missing 3 applicable states (Loading, Empty, Error) and Component 4 lacking any transition-out description.

**Dimension 3 total: 192/250**

---

### Dimension 4: Implementability (250 pts) — Developer

#### Layout specificity (48/80)
- Component 1: Specific — `min-height: 100vh`, flex center, `max-w-[400px]`, `p-6`. Developer can build this.
- Component 2: Partially specific — "全宽常驻插槽" and "height 0 或 auto" but expanded banner height is not specified (h-12? h-14? h-16?). No padding specification for the banner content. A developer would need to guess the banner height.
- Component 3: Largely unspecified — "告警概览卡片 / 规则列表 (Table) / 事件时间线" listed in ASCII tree with no widths, grid columns, gap, max-width, or responsive breakpoint behavior. A developer cannot build this dashboard layout from the spec alone — they must design the entire grid system themselves.
- Component 4: Specific — `max-w-[460px]`, `p-6`, centered.
- Dark mode palette: No dark token values provided — developer cannot implement dark mode despite the design claiming it's supported.

Deduction (-32): Component 3 layout dimensions entirely missing (-15), banner expanded height missing (-5), dark mode palette absent (-8), Component 2 padding unspecified (-4).

#### Data binding explicit (73/80)
All 4 components have Data Binding tables with UI Element → Data Field → Source:
- Component 1: `isAuthenticated: boolean` (session), `redirect_uri: string` (current URL), `eiamAuthUrl: string` (backend config).
- Component 2: `graceRemaining: string` (backend state), `writeAllowed: boolean` (backend state), `authHealth: 'normal' | 'degraded' | 'recovered'` (backend SSE/polling).
- Component 3: `readonlyAlerts: list` (backend read-only API), `authHealth: boolean` (backend probe), `recoverySignal` (backend SSE/polling), `masked: '***'` (backend).
- Component 4: `roleBinding: null | 'admin' | 'user'` (backend auth middleware).

No orphan UI elements detected — every interactive element has a data source.

Deduction (-7): Component 1's Error state message text "登录失败，请重试" is not in the Data Binding table — unclear whether this is static frontend copy or backend-sourced. Component 2's "只读" Badge is mentioned in layout ("(可选) 只读说明徽标 Badge「只读」") but not in Data Binding — its visibility condition is implied from `writeAllowed` but not explicitly mapped.

#### Interaction unambiguity (68/90)
- Component 1: All 4 triggers have explicit action + feedback. "点击「使用平台账号登录」" → "拼接 eiam 授权地址（携带 redirect_uri = 当前页 URL）并整页跳转" → "按钮进入 Loading 态". Clear.
- Component 2: Triggers mostly explicit. However, "后端返回 503（绕前端的直接写请求）→ 前端捕获 503 → toast「鉴权服务暂不可用，操作失败」" is ambiguous — if the request "bypasses the frontend" ("绕前端的直接写请求"), how does the frontend catch it? Is this referring to API calls that bypass the button-disabling logic (e.g., background sync)? The mechanism is unclear.
- Component 3: "用户尝试任何写操作 → 无写操作按钮可点（已移除）→ —" — clear but terse. The "—" feedback is adequate since there's nothing to click.
- Component 4: Single trigger, explicit.

Missing interaction scenarios:
- SSE/polling connection drop (Components 2, 3): "后端 SSE/轮询（10s 探针，滞回 3 次）" — what happens if the SSE connection itself drops or polling requests fail (independent of eiam health)? Does the frontend retry? Show a warning?
- `eiamAuthUrl` config absent (Component 1): what happens if the backend doesn't provide `eiamAuthUrl`? The Data Binding says it comes from "后端配置下发" but no error handling for missing config.
- Browser back button during Loading (Component 1): "跳转发起后锁定按钮，防重复点击；不阻塞浏览器跳转" — but if the user hits back during the redirect, what state is the button in?

Deduction (-22): Ambiguous 503 handling (-7), missing SSE/polling connection failure behavior (-8), missing config-absent scenario (-4), missing back-button-during-loading (-3).

**Dimension 4 total: 189/250**

---

## Cross-Dimension Coherence Check

1. **Component 1 → Component 4 flow gap**: Component 1 Interactions says "eiam 认证成功但无 watchalert_role_binding → 渲染组件 4（403 占位）". Component 4 Placement says "任意受保护页面（告警看板/规则/故障中心）". But after eiam auth success, the user is redirected from the login page to a protected page — the design doesn't explicitly describe this redirect-then-403 flow. A developer implementing this would need to infer that the redirect happens first, then the 403 is rendered on the target page.

2. **Recovery path coherence**: Component 2 "宽限过期" → Component 3 (read-only). Component 3 "已恢复" → "自动整页跳转 eiam 登录页". This skips Component 2's "已恢复" state entirely — the read-only view goes directly to login, not back through the banner. This is consistent (the user needs to re-authenticate) but the full cycle (normal → degraded → read-only → login → normal) is implicit, not documented as a flow.

3. **authHealth type inconsistency**: Component 2 Data Binding types `authHealth` as `'normal' | 'degraded' | 'recovered'` (3-value enum). Component 3 Data Binding types `authHealth` as `boolean`. These are different types for the same conceptual field sourced from the same backend probe. A developer would need to reconcile this.

---

## Phase 3: Blindspot Hunt

### [blindspot-1] SSE/polling transport failure not specified
> "后端 SSE/轮询（10s 探针，滞回 3 次）"

The design specifies what happens when eiam health probe fails (banner appears, write disabled) but never addresses what happens if the SSE/polling **transport itself** drops — e.g., network interruption, SSE connection closed, polling requests timing out. This is distinct from eiam being unhealthy. If the transport fails, the frontend cannot know eiam's health — does it assume degraded? Show a different banner? Retry the connection? This is an operational gap outside all rubric dimensions.

### [blindspot-2] Session expiry behavior (BIZ-auth-008) not addressed
The injected context specifies BIZ-auth-008: "会话过期 → 跳转控制台登录页（不白屏、不无限 loading）；重新登录后回到原操作路径". The design covers eiam unavailability (Components 2, 3) but never addresses session expiry during normal use — what happens when a logged-in user's session expires while browsing the alert dashboard? The design should specify the transition (redirect to login page, preserve return path) to satisfy BIZ-auth-008. This is outside the PRD UI functions but within the business rules the design must respect.

### [blindspot-3] Responsive behavior beyond nav collapse
> "响应式：视口 <768px 时一级导航折叠为汉堡菜单"

The design addresses responsive behavior only for the top nav bar (collapse to hamburger at <768px). No responsive behavior is specified for: the grace banner (Component 2) — long text on mobile; the 403 card (Component 4) — max-w-[460px] may be fine but no breakpoint behavior stated; the read-only dashboard (Component 3) — dashboard cards, table, and timeline on mobile screens. A Senior UX Engineer would flag that responsive interactions can break when layouts designed for desktop are viewed on mobile — the banner text "剩余 X 小时 X 分；此期间告警规则/通知等写操作不可用" could wrap awkwardly or overflow.

### [blindspot-4] Transition timing for state changes
> "hover 仅 opacity 变化、无阴影、150ms 过渡"

The design specifies button hover transition timing (150ms) but no transition timing for: banner expand/collapse (Component 2 — "height 0 或 auto" — instant or animated?), state change between normal/degraded/recovered, 403 card appearance (Component 4), read-only view entry (Component 3). Without timing specs, developers will implement inconsistent transitions across components.

### [blindspot-5] authHealth type inconsistency across components
> Component 2: `authHealth: 'normal' | 'degraded' | 'recovered'`
> Component 3: `authHealth: boolean`

The same conceptual field (`authHealth`, sourced from the same backend probe) is typed differently in Component 2 (3-value enum) vs Component 3 (boolean). This is a cross-component data contract inconsistency that a developer would need to reconcile — does the boolean mean "is healthy" (normal/recovered = true, degraded = false)? Or "is degraded"? The mapping is unspecified.

---

## Score Summary

| Dimension | Score | Max |
|-----------|-------|-----|
| Requirement Coverage | 202 | 250 |
| User Experience | 164 | 250 |
| Design Integrity | 192 | 250 |
| Implementability | 189 | 250 |
| **Total** | **747** | **1000** |

### Score Breakdown by Criterion

| Dimension | Criterion | Score | Max |
|-----------|-----------|-------|-----|
| Requirement Coverage | UI function coverage | 76 | 80 |
| Requirement Coverage | Navigation Architecture coverage | 32 | 40 |
| Requirement Coverage | State requirement coverage | 62 | 80 |
| Requirement Coverage | Edge case handling | 32 | 50 |
| User Experience | Information hierarchy | 58 | 80 |
| User Experience | Interaction intuitiveness | 68 | 80 |
| User Experience | Accessibility | 38 | 90 |
| Design Integrity | Design system adherence | 68 | 80 |
| Design Integrity | Visual coherence | 72 | 90 |
| Design Integrity | State completeness | 52 | 80 |
| Implementability | Layout specificity | 48 | 80 |
| Implementability | Data binding explicit | 73 | 80 |
| Implementability | Interaction unambiguity | 68 | 90 |

---

## Attack Points (Prioritized)

1. **[User Experience]**: Accessibility severely underspecified — no ARIA labels, roles, live regions, keyboard navigation, or focus management anywhere in the document. The grace banner is a status notification with no `role="alert"` or `aria-live` specification. The loading spinner has no text alternative. Contrast violation: muted-foreground (#71717a) on muted background (#f4f4f5) = 4.4:1, failing WCAG AA 4.5:1 for normal text. — Quote: "表头 bg-muted" and "muted 底" — Must add ARIA specifications, keyboard/focus management, and fix contrast (use zinc-600 #52525b which yields ~5.7:1 on muted, or darken muted-foreground).

2. **[Design Integrity]**: Component 3 missing Loading/Empty/Error states for data-fetching lifecycle. — Quote: Component 3 States table has only "只读免登录" and "已恢复" — Must add Loading (skeleton/spinner while fetching `readonlyAlerts`), Empty (no alerts to display), and Error (read-only API failure) states with transitions.

3. **[Requirement Coverage]**: Edge cases not addressed — no empty/no-data state, no long text overflow handling, no slow network handling for Component 3 data fetch. — Quote: "告警概览卡片 / 规则列表 (Table) / 事件时间线" with no empty or error handling — Must add empty states, overflow/truncation rules, and loading skeletons for all data-display components.

4. **[Implementability]**: Component 3 layout lacks all dimensions — "告警概览卡片 / 规则列表 (Table) / 事件时间线" listed without widths, grid columns, gap, or max-width. Developer cannot build this dashboard layout from the spec. — Quote: layout structure has no dimensional specs — Must specify grid layout, card widths, table column behavior, and responsive breakpoints.

5. **[Design Integrity]**: Dark mode palette absent despite claiming CSS variable light/dark switching. — Quote: "全部通过 CSS 变量实现 light/dark 切换" but only Light palette table provided — Must add dark mode token values for all color roles.

6. **[Implementability]**: Component 2 503 handling ambiguous — "后端返回 503（绕前端的直接写请求）→ 前端捕获 503" — if the request bypasses the frontend, how does the frontend catch it? — Must clarify the mechanism (is this about API calls that bypass button-disabling logic?).

7. **[blindspot]**: SSE/polling transport failure not specified — "后端 SSE/轮询（10s 探针，滞回 3 次）" — what happens if the SSE connection itself drops or polling fails independently of eiam health? — Must specify transport-level failure handling (retry, fallback, assumed state).

8. **[blindspot]**: Session expiry behavior (BIZ-auth-008) not addressed — the design covers eiam unavailability but not session expiry during normal use. — Must specify session-expiry → login redirect with return-path preservation per BIZ-auth-008.

9. **[blindspot]**: authHealth type inconsistency — Component 2 types it as `'normal' | 'degraded' | 'recovered'` (enum), Component 3 types it as `boolean`. Same field, same source, different types. — Must unify the type contract across components.

10. **[blindspot]**: Responsive behavior beyond nav collapse — only "<768px 时一级导航折叠为汉堡菜单" specified. No responsive behavior for banner, 403 card, or read-only dashboard. — Must add responsive breakpoints for all components, especially banner text overflow on mobile.

11. **[blindspot]**: Transition timing unspecified for state changes — only button hover (150ms) is specified. Banner expand/collapse, 403 appearance, read-only view entry have no timing. — Must specify transition durations for all state-change animations.

12. **[blindspot]**: Icon system not specified — PRD defines icon keywords (告警/alert, 规则/rule, 故障/event) but design doesn't specify icon library or visual treatment. Banner layout mentions "图标" without specifying source. — Must specify icon library (e.g., Lucide) and icon mappings.
