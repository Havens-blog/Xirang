# cmd/tmp-k8s-cred 诊断逻辑摘录留档

> 留档原因：本工具为**凭证处理工具**（从 Mongo 云账号记录读取并解密 AccessKey 后调用云 API，原 `main.go:55`），
> 按 dead-modules-cleanup 提案 Risk 5 **不整目录入库**（涉敏调用链一旦写入 git 历史不可逆），仅摘录诊断逻辑后删除。
> 本文件只描述查询结构与调用链，**不含凭证读取/解密代码本身**。
> 留档时间：2026-09-09（对应任务 docs/features/dead-modules-cleanup/tasks/2）

## 一、工具定位与目的

`package main` 一次性只读诊断（用后即删），非部署单元。

- **目的**：定位"集群凭证 fetch 无响应"的延迟来源——对阿里云 ACK
  `DescribeClusterUserKubeconfig` 接口计时，比较 `PrivateIpAddress` 两种形态
  （公网 / 私网）的响应耗时，判断无响应是公网/私网形态差异还是账号/区域问题。
- **输入**：`config/prod.yaml`（经 viper 读取运行环境连接配置，与另两只 tmp 工具同模式）。
- **硬编码诊断目标**：一个杭州区域的 ACK 目标 clusterId（诊断时指定的具体值，非凭证，已随目录删除，不在本文复制）。
- **总超时**：120s 整体 context；Mongo server selection 10s。

## 二、Mongo 查询结构

数据访问复用 `internal/account` 既有仓储，不直接构造查询：

| 步骤 | 对象 | 说明 |
|------|------|------|
| 建连 | `mongox.NewMongo(client, v.GetString("mongodb.db"))` | 连接参数全部来自 viper 读 `config/prod.yaml`（`mongodb.dsn/username/password/db`），与 `cmd/tmp-cas-diag`、`cmd/tmp-k8s-diag` 同模式 |
| 仓储 | `accountrepo.NewCloudAccountRepository(accountdao.NewCloudAccountDAO(m))` | `internal/account/repository` + dao，即 account 半活模块的消费方之一（提案 Evidence 已登记 `cmd/tmp-k8s-cred` 这处引用） |
| 查询 | `repo.List(ctx, domain.CloudAccountFilter{Provider: domain.CloudProviderAliyun, Status: domain.CloudAccountStatusActive})` | 过滤条件：**云厂商 = aliyun 且状态 = active**，一次性列出全部合格云账号记录；返回 `(_, total, err)`，工具只取列表 |

查询到的账号记录字段使用：`acc.Name`（展示名）、`acc.Regions`（取 `[0]` 为 region，缺省回退 `cn-hangzhou`）、`acc.AccessKeyID` / `acc.AccessKeySecret`（见下）。

## 三、调用链（涉敏路径描述，不含代码）

```
viper 读 config/prod.yaml
  ├─ security.encryption_key → 初始化默认加解密器（pkg/crypto.InitDefaultCrypto）
  └─ mongodb.* → Mongo 连接
        ↓
internal/account CloudAccountRepository.List（aliyun + active 账号）
        ↓
逐账号：acc.AccessKeyID / acc.AccessKeySecret
   （存储态为加密数据，经 pkg/crypto 按 security.encryption_key 解密得到明文 AK——此段即"涉敏调用链"本体，故本工具不整目录入库）
        ↓
cs.NewClientWithAccessKey(region, AK)   ← 阿里云 CS SDK
        ↓
DescribeClusterUserKubeconfig(req)      × 每账号 2 次（PrivateIpAddress=false/true）
```

触达的云 API（共 1 个，`github.com/aliyun/alibaba-cloud-sdk-go/services/cs`）：

- **`DescribeClusterUserKubeconfig`**：获取指定 ACK 集群的用户 kubeconfig。
  - `Scheme = "https"`
  - `ClusterId` = 硬编码诊断目标集群
  - `PrivateIpAddress` = `requests.NewBoolean(false / true)`，两种形态各调一次
- 客户端按账号 region 创建（`acc.Regions[0]`，缺省 `cn-hangzhou`）；某账号创建客户端失败则跳过该账号（`continue`）。

## 四、关键诊断逻辑

1. **双重形态计时对比**：对同一集群、同一账号，分别以 `PrivateIpAddress=false`（公网）与 `true`（私网）调用 `DescribeClusterUserKubeconfig`，各自记录 `time.Since(t0)`（毫秒精度）。这是工具的核心——若仅公网形态超时/极慢而私网正常，则"fetch 无响应"定位为网络形态问题；两种形态均慢则指向 API 服务端或账号侧。
2. **失败不中断**：单次调用出错仅打印 `account=%s region=%s private=%v: %v (耗时 %v)` 后继续下一形态/账号；打印格式中错误与耗时同行，便于横向对比。
3. **成功输出**：`OK config=%d 字节 (耗时 %v)`——用 kubeconfig 响应字节数佐证响应完整性。
4. **账号遍历顺序**：按 `repo.List` 返回顺序遍历全部 aliyun active 账号，逐个尝试，用于排查"是否只有某个账号的凭证/区域有问题"。
5. **全程只读**：对 Mongo 只读（List），对云 API 只读（Describe），无任何写入。

## 五、诊断结论的去向

该工具的诊断产出（PrivateIpAddress 两形态耗时对比结论）已用于排查集群凭证 fetch 无响应问题；工具本体按提案 In Scope 于本清理轮删除，如需复现同类诊断，可依据本摘录的查询结构（第二节）与调用链（第三节）重写，凭证读取复用 `internal/account` 仓储既有解密路径即可。

## 关联

- 提案：`docs/proposals/dead-modules-cleanup/proposal.md`（Risk 4 / Risk 5、SC-1）
- 任务：`docs/features/dead-modules-cleanup/tasks/2-archive-tmp-tools.md`
- 另两只同族工具：`cmd/tmp-cas-diag`、`cmd/tmp-k8s-diag` 已各自存档 commit 入库后删除
