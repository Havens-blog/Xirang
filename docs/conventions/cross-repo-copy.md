---
title: "跨仓拷贝漂移防线约定"
domains: [copy, hash, sha256, drift, ci, shared-files, theme]
---

# 技术约定：跨仓拷贝漂移防线

> 跨特性可复用约定。源：feature/platform-console。

## 登记

### TECH-cross-repo-copy-001: shared-hashes.json 登记拷贝面

**Requirement**: 新仓 `shared-hashes.json` 逐条登记拷贝文件 × sha256 × 源仓 commit，覆盖 `theme-variables.scss`、`element-theme.scss`、`api/request/eiam.ts`（eiamAxios 模式文件）三处拷贝面。
**Source**: feature/platform-console tech-design §拷贝面漂移防线

## CI 防线

### TECH-cross-repo-copy-002: CI sha256 比对防漂移

**Requirement**: CI job 从 e-cam-web checkout 对应源文件重算 sha256 比对，不一致 → 红；恢复绿必须同步拷贝 + 更新 hash + 跑一条主题冒烟 E2E，漂移来源 commit 记入变更记录。
**Source**: feature/platform-console tech-design §拷贝面漂移防线 CI 防线

## 升级流程

### TECH-cross-repo-copy-003: 拷贝面升级流程

**Requirement**: e-cam-web 侧改动主题 / 请求层文件 → 控制台 CI 当日变红 → 人工评估差异 → 同步拷贝 + hash 更新 + 主题冒烟回归。此防线覆盖 PRD ≥ 5 工作日通知条款未触及的 e-cam-web 侧文件（该条款只覆盖 eiam 端点）。
**Source**: feature/platform-console tech-design §拷贝面漂移防线 升级流程
