---
tags: [interface, error-handling]
date: 2026-09-22
---

# Gotcha: Vue v-model 绑定 const reactive + 子组件 emit 新对象 = 静默失败

## Problem

`resource/tag` 标签页筛选功能「坏而不崩」：输入关键字或切换下拉后，筛选条件不生效、搜索不带新参数，界面无任何报错。

## Root Cause

父组件 `const filters = reactive({...})` 经 `v-model="filters"` 绑定子组件；子组件 `emit('update:modelValue', { ...localFilters })` 发射**新对象**。`v-model` 编译为 `filters = $event`，对 `const` 重新赋值抛 `TypeError: Assignment to constant variable`。Vue 3 的组件事件处理器经 `callWithErrorHandling` 包裹，同步 throw 被 `handleError` **静默吞掉**（仅 console 告警，不 rethrow、不冒泡），于是 `filters` 未更新、后续 `emit('search')` 照常执行但读到空值。构建期 esbuild 报 `"This assignment will throw because 'filters' is a constant"` 警告是唯一可靠线索，但 dev server 下极易被忽略。

## Solution

把对 reactive 对象的「整体重赋值」改为「原地合并」：

```vue
<TagFilters v-model="filters" @search="handleSearch" />
<!-- ↓ 改为 ↓ -->
<TagFilters :model-value="filters" @update:model-value="(v) => Object.assign(filters, v)" @search="handleSearch" />
```

## Reusable Pattern

- 子组件若做「整体替换」式 emit（如 `{ ...localFilters }`），父组件用 `const reactive` 时**不能直接 `v-model`**；要么 `:model-value` + `@update:model-value` 的 `Object.assign` 原地更新，要么改用 `ref` 允许整体替换（`filters.value = $event`）。
- 排查「坏而不崩」的状态不同步：先查是否 v-model 对 const 重赋值，再查 Vue 静默吞错；esbuild build 警告是定位这类 bug 的可靠入口。

## Related Files

- src/views/resource/tag/index.vue
- src/views/resource/tag/components/TagFilters.vue