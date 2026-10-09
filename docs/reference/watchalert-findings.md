# WatchAlert 安全/正确性审查结论与遗留 Backlog

> 审查时间 2026-09-24，范围 `D:\WatchAlert`(后端) + `D:\WatchAlert-web`(前端)。
> 由 correctness + security 两轮 agent 并行审查，合并去重后记录如下。

## 一、已修复（6 个 PR，均已推送到 fork → opsre/WatchAlert）

| PR | 分支 | 内容 |
|---|---|---|
| #257 | fix/mute-recover-notify-nil-pointer | mute.RecoverNotify nil 指针 panic |
| #258 | chore/ignore-local-config | config.yaml 凭据隔离 + example 模板 |
| #259 | fix/oidc-discovery-and-userinfo | OIDC discovery URL / Bearer 前缀 / refresh_token 校验 |
| #260 | fix/auth-hardening | JWT 空密钥+Valid 恒真 / userInfo 密码哈希外泄 / oidcInfo 下发 secret / ginlog 打请求体 |
| #261 | fix/ssrf-and-http-guard | http NewRequest 判空前 Header.Set nil panic / SSRF 出站校验 |
| #262 | fix/eval-goroutine-leak-and-nil-guard | Submit goroutine 泄漏 / *instance.Enabled & *oldRule.Enabled nil panic / oidc return-nil-err |

额外：`fix/datasource-forward-http-headers`(datasource Bearer 转发，早期 PR)。

## 二、遗留 Backlog（未修，需拍板或先记录）

### Tier 2 —— 需决策/有操作成本（HIGH）

1. **生产凭据已进 git 历史**：`config/config.yaml` 历史提交含 `pass:w8t.123`、Redis `1234567890`、飞书 `appSecret`/`token` 等。需 `git filter-repo` 洗历史 + **你在阿里云/飞书侧轮换密码**（操作项，非代码）。
2. **数据源/云厂商密钥明文入库 + 跨租户 IDOR**：`models/datasource.go` Auth.Pass/AliCloudSk/AWS.SecretKey 明文；`repo/datasource.go` `Get`/`GetInstance` 只按 id、不含 `tenant_id`。建议 KMS 加密 + 响应脱敏 + 查询强制带租户。
3. **SSRF 端点未收紧鉴权**：#261 已加出站地址拦截（回环/链路本地/元数据），但 `datasource/ping`、`probing/once`、`notice/test` 仍只有 `Auth()` 无 `Permission()`。内网 SSRF 需按角色收紧（另立 PR，注意别误伤前端“测试连接”）。
4. **错误响应回显内部细节**：`api/entry.go`、`middleware/Tenant.go`、`pkg/tools/http.go:121` 回显 `err.Error()`/原始 body，可能泄露 SQL/内部结构。

### Tier 3 —— 中低优先（MEDIUM/LOW，含正确性）

**安全：**
- 密码用**无盐 MD5**（`pkg/tools/hash.go`），改 bcrypt/argon2 会作废现有用户，需迁移方案。
- OIDC 硬编码弱口令 `types/oidc.go:3` `OidcPassword="watchAlert"`。
- 开放注册：`/api/system/register` 可自选 UserId/Role；`checkUser` 可枚举用户名/邮箱/手机。
- 通知内容注入：`pkg/templates/parser.go` 用 `text/template` 不转义，labels/annotations 拼进飞书/钉钉/邮件。
- `pkg/provider/kubernetes.go:32` `os.ReadFile(kubeConfigContent)` 用户可控路径 → 任意文件读。
- API Key 控制 `api/apikey.go` 仅 `Auth()`，明文返 Key。
- CORS `*` + credentials；OIDC 缺 state/PKCE；`InsecureSkipVerify:true`(全部出站)；前端 token 存 localStorage；`cmd/main.go` 开 9999 pprof。

**正确性：**
- `services/rule.go Import`：`forDuration=1` 被 `d.Seconds()` 无条件覆盖为 0（死代码，改默认语义）。
- `services/silence.go`：未来生效逻辑改 `r.Status` 但 `silence.Status` 硬编码 1，计划静默立即生效。
- `Where("enabled = ?", "1")`（4 处）字符串 vs `*bool`，MySQL 碰巧能跑、Postgres 报错。
- `alert/process/process.go:59` 丢弃 `GetEventFromCache` error，Redis 抖动时丢事件连续性。
- `internal/cache/faultCenter.go:50`、`internal/repo/user.go:180` 丢弃 `sonic.Unmarshal` error。
- `regexp.MustCompile` 用户正则非法（`[`）panic（consumer/mute handleAlert 链路）。
- `GetFingerprint` 标签 FNV + XOR 满足交换律，`{a:x,b:y}` 与 `{a:y,b:x}` 同指纹。
- `oidc.CookieConvertToken` 忽略 user 是否存在的 bool。
- `repo/fault_center.go:116 Reset` 三字段相互覆盖；`repo/gorm.go` 传 `[]string` 当标量。
- `services/alert.go RePushRule` 无缓冲通道 + 死锁隐患。

> 未发现 SQL 注入（全参数化）、Web 上传路径遍历；GORM `db.Where()` 不赋值写法经实测**不是 bug**。