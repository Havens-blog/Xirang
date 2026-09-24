// Package llm LLM Provider 抽象 + GuardedLLM 降级封装。
//
// 接口签名见 tech-design › Interface 4；降级谓词：单次调用超时 > 12s
// （可配）或连续 2 次失败或滚动 1h 预算耗尽 → LLM 不可用，进入三级降级
// （L1 模板意图匹配 → L2 预置查询入口 → L3 明示降级），不中断主链路。
// 语义与 e-cam-service/internal/logquery/llm 一致，在本服务内独立实现，
// 不跨模块 import。具体实现在后续任务交付。
package llm
