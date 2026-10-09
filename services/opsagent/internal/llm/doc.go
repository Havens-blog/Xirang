// Package llm LLM Provider 抽象 + GuardedLLM 降级封装。
//
// 接口签名见 tech-design › Interface 4；降级谓词：单次调用超时 > 12s
// （可配）或连续 2 次失败 → LLM 不可用，进入三级降级（L1 模板意图匹配 →
// L2 预置查询入口 → L3 明示降级），不中断主链路。语义与
// e-cam-service/internal/logquery/llm 一致，在本服务内独立实现，不跨模块 import。
//
// 三个 provider（qwen/deepseek/openai）共享 OpenAI 兼容 chat/completions 协议，
// 落地为 provider.go 的 httpProvider；未配置 apiKey 走 disabledProvider（空串降级）。
// 三级降级决策（L1/L2/L3）与滚动 1h 预算由 3.3 DegradationController 负责。
package llm
