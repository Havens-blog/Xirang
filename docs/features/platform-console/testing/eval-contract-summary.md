# eval-contract Summary — platform-console

**Target**: 850/1000（且每维度 ≥ 门槛）

| Contract Group | Iteration 1 | Iteration 2 | Result |
|----------------|-------------|-------------|--------|
| identity-permission-governance-golden-path (5) | 840 | **918** | PASS |
| login-and-session-lifecycle (4) | 813 | **937** | PASS |
| tenant-boundary-enforcement (6) | 792 | **886** | PASS |
| readonly-audit-access (4) | 819 | **930** | PASS |

## 首轮共性缺陷（已全修复）

1. 语义纯度：Side-effect/Input/State 维度嵌入 file:line、函数名、API 路径、casbin 等实现耦合 → 全部下沉至 `<!-- 实现锚点 -->` 注释。
2. 事实依据：契约引用 journey 旧 `F-*` ID，与 fact-table.json 的 `IAM_*` 规范 ID 错配 → 全量重映射；409/60s 等 UNKNOWN 事实不再当作既定行为断言（标注 PARTIAL/UNKNOWN）。
3. 前置数据声明：entity_type 用抽象类型（Caller/TargetResource/LoginProvider/EiamService）不匹配领域模型 → 对齐 User/Role/Policy/IdentitySource；补 relationship_type/parent_entity；inline 与 bottom fixture_spec 双声收敛为单一权威。
4. 完整性：Web/API 面缺强制派生 Outcome（validation-error/session-expired/unauthorized）→ 补齐。
5. 一致性：Journey Invariants 清单各契约不一致 → 统一 5 条。
6. 前置互斥：空态/错误/成功 Outcome 前置重叠 → 精化为唯一可选。

## 残余观察（终轮通过后，非阻塞）

- 全部 2 轮通过；残余均为「某 Step 漏同步同类修复、entity 建模细节、事实引用主语更精确」级。
- 首轮 4 组全部 FAIL 的主因是「总分 <850 但各维度均过阈」——语义纯度/前置数据两维度是主要失分位。
- 跨文档一致性遗留（可选后续）：golden-path journey.md 的「≤30 秒」注记，契约已统一为 60 秒（代码实测）；journey.md 的 U2 引用可后续同步为 60 秒。