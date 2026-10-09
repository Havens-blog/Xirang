---
title: "性能约定"
domains: [performance, latency, lcp, bundle, caching, cdn, gzip]
---

# 技术约定：性能

> 跨特性可复用约定。源：feature/platform-console。

## 首屏预算

### TECH-performance-001: 首屏静态资源 gzip ≤ 300 KB 预算

**Requirement**: 首屏静态资源 gzip ≤ 300 KB，预算分解：entry 自研 ≤ 60 + element-plus ≤ 140 + vue/router/pinia ≤ 50 + 其余（css / 字体 / 图标）≤ 50（单位 KB）。手段：11 页路由级全 `import()` 懒加载；Element Plus unplugin 按需 + 图标按需；`manualChunks` 三分包；构建期预生成 `.gz` + nginx `gzip_static on`。测量：构建脚本汇总 `dist/**/*.gz` 体积，CI 超 300 KB 即 fail。
**Source**: feature/platform-console tech-design §性能预算

## LCP

### TECH-performance-002: LCP ≤ 2.5s / 二次 ≤ 1s

**Requirement**: 同域内网 LCP ≤ 2.5s（二次访问 ≤ 1s）。手段：首屏仅 entry + workbench chunk + element-plus 核心 css；静态资源 `Cache-Control: max-age=31536000, immutable`（content-hash 文件名），`index.html` no-cache。测量：Playwright trace + Lighthouse，登录页与工作台两个 LCP 采样点，同域内网基线。
**Source**: feature/platform-console tech-design §性能预算

## CDN / 部署形态

### TECH-performance-003: 静态站可 CDN 但须同登录域

**Requirement**: 纯静态前端站可 CDN 加速，但须与 e-cam-web / eiam 同登录域；跨域 CDN 直调 eiam 不支持（SameSite / CORS 限制）。仅当未来变更部署形态并专项核对 CORS / SameSite 后才可重议。
**Source**: feature/platform-console prd-spec.md §Performance Requirements
