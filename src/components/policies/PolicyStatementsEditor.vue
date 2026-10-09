<script setup lang="ts">
/**
 * 策略 Statement 编辑器（task 6.2，UF-10 增强）—— 图形化编排 / JSON 双模式，v-model 双向同步。
 *
 * 单一事实源 = JSON 文本（modelValue，规范复数形 [{effect, actions, resources, condition?}]，
 * 与 parseStatementsJson/serializeStatementsJson 同源，api/policies.ts）。
 * - 图形化模式：进入时 parseStatementsJson 解析文本 → 本地 graphStatements；
 *   每次图形编辑（effect/actions/resources/增删条）→ serializeStatementsJson 回写 modelValue；
 *   condition 仅 JSON 模式可编辑（图形模式保留并不渲染编辑控件，badge 提示）。
 * - JSON 模式：textarea 直绑 modelValue。
 * - JSON 非法时不允许进入图形化编辑（warn 横幅 + 指引先修 JSON）。
 *
 * 动作词表（可选）：GET /api/iam/permission/manifest（eiam Needs 双码
 * iam:permission:manifest + iam:permission:menus_by_urns）；持有双码才拉取，
 * 失败/缺码降级自由输入（el-select allow-create，不阻断编辑）。
 */
import { computed, onMounted, ref } from "vue";
import {
  parseStatementsJson,
  serializeStatementsJson,
} from "@/api/policies";
import type { PolicyStatement } from "@/api/types";
import { fetchPermissionManifest, type ManifestAction } from "@/api/permissions";
import { useCan } from "@/composables/useCan";
import { IAM_PERMISSION } from "@/config/permission-codes";

const props = defineProps<{
  /** JSON 文本（规范复数形，单一事实源） */
  modelValue: string;
  /** 父级字段错误（JSON 校验失败标红文案；图形化模式经 warn 横幅展呈） */
  error?: string;
}>();

const emit = defineEmits<{
  (e: "update:modelValue", value: string): void;
}>();

const { can } = useCan();

const editorTab = ref<"graphical" | "json">("graphical");

// ---- 图形化模型 ----
const graphStatements = ref<PolicyStatement[]>([]);
/** ok = 可编辑；invalid = JSON 解析失败，须先修 JSON */
const graphState = ref<"ok" | "invalid">("ok");
const graphError = ref<string>("");

function syncFromJson(): void {
  const parsed = parseStatementsJson(props.modelValue);
  if (parsed.ok) {
    graphStatements.value = parsed.statements.map((s) => ({
      effect: s.effect,
      actions: [...s.actions],
      resources: [...s.resources],
      condition: s.condition,
    }));
    graphState.value = "ok";
    graphError.value = "";
  } else {
    graphState.value = "invalid";
    graphError.value = parsed.error;
  }
}

function commitGraph(): void {
  emit("update:modelValue", serializeStatementsJson(graphStatements.value));
}

function switchTab(tab: "graphical" | "json"): void {
  if (tab === "graphical") syncFromJson();
  editorTab.value = tab;
}

function addStatement(): void {
  graphStatements.value.push({ effect: "Allow", actions: [], resources: [] });
  commitGraph();
}

function removeStatement(index: number): void {
  if (graphStatements.value.length <= 1) return; // 至少保留一条
  graphStatements.value.splice(index, 1);
  commitGraph();
}

/** effect 变更（v-model 单值） */
function onEffectChange(stmt: PolicyStatement, value: string): void {
  if (value !== "Allow" && value !== "Deny") return;
  stmt.effect = value;
  commitGraph();
}

/** actions 变更（v-model 数组） */
function onActionsChange(stmt: PolicyStatement, value: unknown): void {
  stmt.actions = Array.isArray(value)
    ? (value.filter((v): v is string => typeof v === "string") as string[])
    : [];
  commitGraph();
}

/** resources 变更（v-model 数组） */
function onResourcesChange(stmt: PolicyStatement, value: unknown): void {
  stmt.resources = Array.isArray(value)
    ? (value.filter((v): v is string => typeof v === "string") as string[])
    : [];
  commitGraph();
}

// ---- JSON 模式 ----
const jsonText = computed({
  get: () => props.modelValue,
  set: (v: string) => emit("update:modelValue", v),
});

// ---- 动作词表（manifest；双码缺一或失败 → 降级自由输入）----
const vocab = ref<ManifestAction[]>([]);
const vocabLoaded = ref(false);
/** 词表可用性：el-select 仍可 allow-create，此态只影响选项丰富度。
 *  服务端 Needs 双码（manifest + menus_by_urns，handler.go:51-53）；实测 dev admin
 *  profile 无 menus_by_urns 码但 super 豁免、manifest 直接 200（scripts/probe-eiam.cjs），
 *  故客户端仅按 manifest 单码 gate，失败/缺码一律降级自由输入（不阻断编辑）。 */
const canFetchVocab = computed(() => can(IAM_PERMISSION.MANIFEST));

/** 选项：按 group 分组（el-option-group），label = code（name） */
const vocabGroups = computed(() => {
  const map = new Map<string, ManifestAction[]>();
  for (const a of vocab.value) {
    const key = a.group || a.service || "其他";
    const list = map.get(key);
    if (list) list.push(a);
    else map.set(key, [a]);
  }
  return [...map.entries()].map(([label, entries]) => ({
    label,
    entries: [...entries].sort((x, y) => x.code.localeCompare(y.code)),
  }));
});

onMounted(() => {
  syncFromJson();
  if (!canFetchVocab.value) return;
  fetchPermissionManifest()
    .then((actions) => {
      vocab.value = actions;
      vocabLoaded.value = true;
    })
    .catch(() => {
      vocabLoaded.value = false; // 降级自由输入（不阻断编辑）
    });
});
</script>

<template>
  <div class="pse" data-testid="policy-statements-editor">
    <!-- 模式切换 -->
    <div class="pse-tabs" role="tablist" aria-label="Statement 编辑模式">
      <button
        type="button"
        role="tab"
        class="pse-tab"
        :class="{ on: editorTab === 'graphical' }"
        :aria-selected="editorTab === 'graphical'"
        data-testid="pse-tab-graphical"
        @click="switchTab('graphical')"
      >
        图形化编排
      </button>
      <button
        type="button"
        role="tab"
        class="pse-tab"
        :class="{ on: editorTab === 'json' }"
        :aria-selected="editorTab === 'json'"
        data-testid="pse-tab-json"
        @click="switchTab('json')"
      >
        JSON 模式
      </button>
    </div>

    <!-- 图形化模式 -->
    <div v-if="editorTab === 'graphical'" class="pse-graph">
      <div v-if="graphState === 'invalid'" class="pse-warn" role="alert">
        {{ graphError }} —— 请先切到 JSON 模式修正，图形化编排暂不可用。
      </div>
      <template v-else>
        <div
          v-for="(s, i) in graphStatements"
          :key="i"
          class="pse-stmt"
          data-testid="pse-statement"
        >
          <header class="pse-stmt-head">
            <span class="pse-stmt-title">Statement {{ i + 1 }}</span>
            <button
              type="button"
              class="btn btn--text btn--danger btn--sm"
              :disabled="graphStatements.length <= 1"
              :data-testid="`pse-stmt-remove-${i}`"
              @click="removeStatement(i)"
            >
              删除此条
            </button>
          </header>

          <div class="pse-grid">
            <label class="field">
              <span class="field__label">Effect</span>
              <el-select
                class="pse-select"
                :model-value="s.effect"
                data-testid="pse-effect"
                @update:model-value="onEffectChange(s, $event)"
              >
                <el-option label="Allow" value="Allow" />
                <el-option label="Deny" value="Deny" />
              </el-select>
            </label>

            <label class="field pse-actions-field">
              <span class="field__label">Actions（可多选/自定义搜索）</span>
              <el-select
                class="pse-select"
                :model-value="s.actions"
                multiple
                filterable
                allow-create
                default-first-option
                placeholder="搜索或输入能力码，如 iam:user:view"
                data-testid="pse-actions-el"
                @update:model-value="onActionsChange(s, $event)"
              >
                <el-option-group
                  v-for="g in vocabGroups"
                  :key="g.label"
                  :label="g.label"
                >
                  <el-option
                    v-for="a in g.entries"
                    :key="a.code"
                    :label="`${a.code}（${a.name}）`"
                    :value="a.code"
                  />
                </el-option-group>
              </el-select>
              <span class="muted">
                动作码词表（manifest 实核）；词表缺项可直接输入自定义码
              </span>
            </label>

            <label class="field">
              <span class="field__label">Resources（URN，可多填）</span>
              <el-select
                class="pse-select"
                :model-value="s.resources"
                multiple
                filterable
                allow-create
                default-first-option
                placeholder="输入资源 URN 后回车，如 cert/*"
                data-testid="pse-resources-el"
                @update:model-value="onResourcesChange(s, $event)"
              />
              <span class="muted">资源 URN 或通配（如 cert/*、tenant/{id}）</span>
            </label>
          </div>

          <p
            v-if="s.condition && s.condition.length > 0"
            class="pse-cond"
            data-testid="pse-condition-note"
          >
            ⚙ 本条含 condition 条件（{{ s.condition.length }} 项）——仅在 JSON
            模式可编辑，图形化保存不丢失。
          </p>
        </div>

        <div class="pse-foot">
          <button
            type="button"
            class="btn btn--ghost btn--sm"
            data-testid="pse-add-statement"
            @click="addStatement"
          >
            ＋ 添加 Statement
          </button>
          <span class="muted">
            {{
              vocabLoaded
                ? `动作码词表已加载（${vocab.length} 项，manifest）`
                : "动作码词表未加载，可自由输入（不影响保存）"
            }}
          </span>
        </div>
      </template>
    </div>

    <!-- JSON 模式 -->
    <div v-else class="pse-json">
      <label class="field">
        <span class="field__label"
          >Statement（JSON）<span class="required">*</span></span
        >
        <textarea
          v-model="jsonText"
          class="input input--code"
          rows="12"
          spellcheck="false"
          data-testid="edit-policy-statements"
          :aria-invalid="!!error"
          :aria-describedby="error ? 'pse-json-error' : 'pse-json-hint'"
        />
        <span id="pse-json-hint" class="muted"
          >JSON 数组，每条含 effect（"Allow"|"Deny"）、actions 与 resources
          非空字符串数组、可选 condition；与图形化模式双向同步</span
        >
        <span v-if="error" id="pse-json-error" class="field__error" role="alert">
          {{ error }}
        </span>
      </label>
    </div>
  </div>
</template>

<style scoped lang="scss">
.pse {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.pse-tabs {
  display: inline-flex;
  gap: 4px;
  background: var(--tag-bg);
  border-radius: 8px;
  padding: 3px;
  align-self: flex-start;
}

.pse-tab {
  border: none;
  background: transparent;
  color: var(--text-regular);
  font-size: 13px;
  padding: 6px 14px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 120ms ease;

  &.on {
    background: var(--bg-surface);
    color: var(--accent-blue);
    font-weight: 600;
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.15);
  }
}

.pse-graph {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.pse-warn {
  padding: 10px 12px;
  border-radius: 8px;
  background: color-mix(in srgb, var(--color-danger) 12%, transparent);
  color: var(--color-danger);
  font-size: 13px;
}

.pse-stmt {
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  padding: 12px 14px;
  background: var(--bg-surface);
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.pse-stmt-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.pse-stmt-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--accent-blue);
}

.pse-grid {
  display: grid;
  grid-template-columns: 150px 1fr;
  gap: 10px 14px;
}

.pse-actions-field,
.pse-cond {
  grid-column: 1 / -1;
}

.pse-cond {
  margin: 0;
  font-size: 12px;
  color: var(--text-secondary);
}

.pse-select {
  width: 100%;
}

.pse-foot {
  display: flex;
  align-items: center;
  gap: 12px;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 4px;

  &__label {
    font-size: 13px;
    color: var(--text-regular);
  }
  &__error {
    font-size: 12px;
    color: var(--color-danger);
  }
}

.required {
  color: var(--color-danger);
  margin-left: 2px;
}

.muted {
  color: var(--text-secondary);
  font-size: 12px;
}

.input {
  height: 36px;
  padding: 0 10px;
  background: var(--input-bg);
  border: 1px solid var(--input-border);
  border-radius: 8px;
  color: var(--text-primary);
  font-size: 14px;
}

.input--code {
  height: auto;
  font-family: ui-monospace, Consolas, monospace;
  font-size: 12px;
  line-height: 1.5;
  padding: 8px 10px;
}

.btn {
  height: 36px;
  padding: 0 14px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 4px;

  &--text {
    background: transparent;
    color: var(--accent-blue);
  }
  &--danger {
    color: var(--color-danger);
    background: transparent;
  }
  &--ghost {
    background: transparent;
    border: 1px solid var(--border-subtle);
    color: var(--text-regular);
  }
  &--sm {
    height: 28px;
    padding: 0 8px;
    font-size: 12px;
  }
  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}
</style>