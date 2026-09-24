// Package orchestrator 编排流水线：意图识别（Coordinator）→ 证据采集
// （LogAnalyst/Monitor/Inspector）→ LLM 解读 → 报告组装落库。
//
// 入口接口为 Orchestrator.RunChat（对话式排障）与 Orchestrator.RunAlert
// （告警主动触发），Go 类型签名见 tech-design › Interface 3；同步编排延迟
// 预算 ≤ 25s，超时转 202 异步回退。具体实现在后续任务交付。
package orchestrator
