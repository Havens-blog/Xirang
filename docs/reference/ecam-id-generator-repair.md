# e-cam-service ID 生成器计数落后修复备忘

> 2026-09-22 服务树三期实测时暴露。现象：新建规则撞 `E11000 duplicate key error ... id_1 dup key: { id: 1 }`；新建节点得到 id=1 而实际节点已到 id=10。

## 机制与根因

ID 生成靠 `pkg/mongox/mongo.go:GetIdGenerator(collection)`：

```go
coll := m.Database().Collection("ecam_id_generator")
// doc: { name: <collection名称>, next_id: N }
// FindOneAndUpdate: $inc next_id + 1，返回递增后的值作为新 id
```

即 `ecam_id_generator` 集合里，每个用自增的 collection 一条 `{name, next_id}` 计数。**当计数 < 该 collection 实际 max(id) 时**（典型：数据被导入/恢复但计数集合没跟上），下一次 `$inc` 就会撞已有 id 的唯一索引 → E11000。

为什么规则那次会自己愈合：每次失败也推进了计数（0→1→2 撞 id 1/2，2→3→4 成功），所以规则计数现在跟上了。**但这不可靠**——节点集合计数仍落后（给 id=1，实际 max=10），再建 2 个节点后第 3 个就会撞 id=4。

## 修复脚本

修复 = 把每个 `next_id` 抬到对应集合的 `max(id)`（下次 $inc 即得 max+1，永不撞）。**通用写法**：遍历 `ecam_id_generator` 里所有 name，逐个对齐其数据集合的 max(id)，无需手列清单。

```js
// 连接目标库后，在 ecam 库下执行：
// mongosh "mongodb://<host>:27017/ecam?authSource=admin" --authenticationDatabase admin -u ecmdb -p '123456' --eval '...'

const gens = db.ecam_id_generator.find().toArray();
for (const g of gens) {
  let maxId = 0;
  try {
    const m = db.getCollection(g.name).find({}, { id: 1 }).sort({ id: -1 }).limit(1).toArray();
    if (m.length && typeof m[0].id === 'number') maxId = m[0].id;
  } catch (e) {
    print(g.name + ': 集合不存在或读取失败，跳过 (' + e.message + ')');
    continue;
  }
  if (maxId > (g.next_id || 0)) {
    db.ecam_id_generator.updateOne(
      { name: g.name },
      { $set: { next_id: maxId } }
    );
    print(g.name + ': next_id ' + (g.next_id || 0) + ' -> ' + maxId);
  } else {
    print(g.name + ': OK (next_id=' + (g.next_id || 0) + ' >= max=' + maxId + ')');
  }
}
```

如只想修服务树相关 4 个（本次暴露的），显式清单版：

```js
const cols = ["ecam_service_tree_node", "ecam_binding_rule", "ecam_resource_binding", "ecam_environment"];
for (const c of cols) {
  const m = db.getCollection(c).find({}, { id: 1 }).sort({ id: -1 }).limit(1).toArray();
  const maxId = m.length ? m[0].id : 0;
  const g = db.ecam_id_generator.findOne({ name: c }) || { next_id: 0 };
  if (maxId > g.next_id) {
    db.ecam_id_generator.updateOne({ name: c }, { $set: { next_id: maxId } }, { upsert: true });
    print(c + ': ' + g.next_id + ' -> ' + maxId);
  } else print(c + ': OK (' + g.next_id + ')');
}
```

## 在哪执行

配置（`e-cam-service/config/`），服务通过 `cmd start -f <yaml>` 选择，**默认 `config/prod.yaml`**（`cmd/root.go` 默认值；`dev.ps1` 不带 `-f`，即正在跑的就是它）：

| 配置 | DSN | 库 |
|------|-----|-----|
| prod.yaml | `mongodb://106.52.187.69:27017/ecam?authSource=admin` | ecam |
| test.yaml | `mongodb://118.145.73.93:27017/ecam?authSource=admin` | ecam |

凭证均为 `ecmdb / 123456`（DSN 里没带，`config/*.yaml` 的 `mongodb.username/password` 两块）。

mongosh 连接命令示例：

```bash
mongosh "mongodb://106.52.187.69:27017/ecam?authSource=admin" \
  --authenticationDatabase admin -u ecmdb -p '123456'
```

## 安全与验证

- 抬 `next_id = max(id)` 永远安全：只可能让后续 id 变大、稀疏，不会撞索引；只有漏 COLLECTION 才撞。
- 有数据集合是云导入的大 id（如 `ecam_instance` 里 id 到千万级）——抬到那么大无害，只是后续自增 id 稀疏；别把大 id 当异常。
- 建议低峰或暂停服务窗口执行（`$inc` 原子、`$set` 只会往上抬，风险小，但避开并发窗口更稳）。
- 执行后可复跑一遍：第二遍应全部 `OK`（幂等）。再 `forge`/页面各建一条规则与节点验证不撞。

## 使用 GetIdGenerator 的 collection 清单（供全量自查，不全则补）

`ecam_service_tree_node`、`ecam_binding_rule`、`ecam_resource_binding`、`ecam_environment`、
`ecam_cloud_account`、`ecam_instance`、`ecam_cloud_asset`、`ecam_attribute`、`ecam_attribute_group`、
`ecam_instance_relation`、`ecam_alert_rule`、`ecam_alert_event`、`ecam_notification_channel`、
`ecam_audit_log`、`ecam_change_history`、`ecam_change_order_audit`、`ecam_cost_*`（alloc/bill/budget/anomaly/recommendation…）、
`ecam_dict_type`、`ecam_dict_item`、`ecam_iam_user`、`ecam_iam_user_group`、`ecam_iam_policy_template`、
`ecam_iam_sync_task`、`ecam_tag_policy`、`ecam_tag_rule`、`ecam_vm_template`、`ecam_endpoint` 等
（`grep -rn "GetIdGenerator" internal/` 可拿全量调用点并对齐 collection 常量）。

## 长期可选加固

`GetIdGenerator` 可改成「首次遇到该 collection 时用 `max(id)` 兜底」（懒初始化计数 = max，而不是从 0 起），从根上避免计数落后。当前未做，仅靠本修复脚本兜。