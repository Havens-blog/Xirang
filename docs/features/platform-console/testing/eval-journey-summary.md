# eval-journey Summary — platform-console

**Target**: 850/1000（且每维度 ≥ 门槛）

| Journey | Iteration 1 | Iteration 2 | Iteration 3 | Result |
|---------|-------------|-------------|-------------|--------|
| identity-permission-governance-golden-path | 735 | **862** | — | PASS |
| login-and-session-lifecycle | 550 | **854** | — | PASS |
| tenant-boundary-enforcement | 613 | **872** | — | PASS |
| readonly-audit-access | 570 | 839 | **890** | PASS |

## 首轮共性缺陷（已全部修复）

1. `golden_path: false` 元数据与正文 Happy Path 矛盾（否决项）。
2. 无 Fact Table / fact_id / UNKNOWN / `source: inferred` 标注。
3. 缺 surface 强制派生 Outcome（web 的 validation-error + session-expired、api 的 unauthorized 401）。
4. 自然语言耦合实现细节（eiam/nginx/正则/HTTP 状态码）。
5. 单 Outcome/Step 结构、分支拆成独立 Step、无 Precondition 互斥。
6. 无 @web-e2e / @api-functional 测试类型标签。

## 残余观察（终轮通过后，非阻塞）

- 3 条在 iteration 2 通过，readonly 需 3 轮（前置互斥与 Fact 越界为主因）。
- 通过后的残余攻击点均为「事实引用越界需补事实 / inferred 标注更具体 / 编号对齐」级，不影响通过。
- readonly contracts/ 目录为空，已显式登记由下游 gen-contracts 生成（50/50 比例待契约产出后闭环）。