---
title: "技术栈与版本约定"
domains: [stack, vue, vite, element-plus, pinia, typescript, pnpm, versions]
---

# 技术约定：技术栈与版本

> 跨特性可复用约定。源：feature/platform-console。

## 技术栈锁同版本

### TECH-tech-stack-001: 技术栈锁同版本与 e-cam-web 同源

**Requirement**: Vue 3.5 + Vite 7 + TypeScript 5.9 + Vue Router 4.5 + Pinia 3（persistedstate）+ Element Plus 2.10（锁与 e-cam-web 同版本）+ @element-plus/icons-vue + unplugin-vue-components / auto-import 按需自动引入 + Sass。v1 采用「拷贝 + 锁同版本」共享机制。
**Source**: feature/platform-console tech-design §Overview

## 包管理

### TECH-tech-stack-002: packageManager 字段锚定

**Requirement**: pnpm ≥ 9，新仓 `package.json` 写入 `packageManager` 字段锚定精确版本（e-cam-web 现有 manifest 未声明 packageManager，控制台新仓为首个显式锚点；建仓时以 lockfile 锁死）。
**Source**: feature/platform-console tech-design §Overview 包管理
