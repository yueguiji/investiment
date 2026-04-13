<template>
  <n-modal :show="show" preset="card" :title="titleText" style="width: 980px" @update:show="handleShowUpdate">
    <div class="analysis-shell">
      <div class="analysis-toolbar">
        <div class="analysis-toolbar-grid">
          <div class="analysis-info-card">
            <div class="analysis-info-item">
              <span class="analysis-info-label">AI 配置</span>
              <div class="analysis-note">{{ modelNote }}</div>
            </div>
            <div class="analysis-info-item">
              <span class="analysis-info-label">当前模板</span>
              <div class="analysis-template">{{ templateNote }}</div>
            </div>
          </div>
          <div class="analysis-select-shell">
            <span class="analysis-select-label">切换模板</span>
            <n-select
              v-model:value="selectedPromptTemplateId"
              :options="promptOptions"
              size="small"
              placeholder="使用当前场景默认模板"
            />
          </div>
          <n-button class="analysis-run-button" type="primary" :loading="loading" @click="runAnalysis">
            {{ actionText }}
          </n-button>
        </div>
      </div>

      <n-spin :show="loading">
        <div v-if="analysis" class="analysis-preview-wrap">
          <MdPreview :editor-id="editorId" :model-value="analysis" theme="dark" preview-theme="github" />
        </div>
        <div v-else class="analysis-empty">
          {{ emptyText }}
        </div>
      </n-spin>
    </div>
  </n-modal>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useMessage } from 'naive-ui'
import { MdPreview } from 'md-editor-v3'

const BASE_TIMEOUT_MS = 90000
const HEAVY_TIMEOUT_MS = 150000
const EXTRA_HEAVY_TIMEOUT_MS = 210000
const SYSTEM_PROMPT_TYPE = '模型系统Prompt'

const props = defineProps({
  show: { type: Boolean, default: false },
  mode: { type: String, default: 'single' },
  fundCode: { type: String, default: '' },
  scope: { type: String, default: 'holdings' },
  scopeLabel: { type: String, default: '' },
  fundCodes: { type: Array, default: () => [] },
  title: { type: String, default: '' },
  betterReferenceCode: { type: String, default: '' },
  betterDimension: { type: String, default: 'balanced' },
  sameTypeOnly: { type: Boolean, default: false },
  sameSubTypeOnly: { type: Boolean, default: false },
  feeFree7: { type: Boolean, default: true },
  feeFree30: { type: Boolean, default: true },
  includeAClass: { type: Boolean, default: false },
  onlyAClass: { type: Boolean, default: false },
  betterTopN: { type: Number, default: 3 }
})

const emit = defineEmits(['update:show'])

const message = useMessage()
const loading = ref(false)
const analysis = ref('')
const errorMessage = ref('')
const modelName = ref('')
const templateName = ref('')
const requestSerial = ref(0)
const promptTemplates = ref([])
const selectedPromptTemplateId = ref(null)

const normalizedFundCodes = computed(() =>
  Array.from(new Set((props.fundCodes || []).map((item) => String(item || '').trim()).filter(Boolean)))
)

const editorId = computed(() => [
  'fund-ai',
  props.mode,
  props.fundCode || props.scopeLabel || props.scope || props.betterReferenceCode || 'panel'
].join('-'))

const titleText = computed(() => {
  if (props.title) return props.title
  if (props.mode === 'selection') return '勾选基金 AI 对比'
  if (props.mode === 'collection' && (props.scope === 'tab' || props.scope === 'watchlist')) return '当前页签 AI 分析'
  if (props.mode === 'collection') return '基金持仓 AI 分析'
  if (props.mode === 'better') return '推荐基金 AI 对比分析'
  return '基金 AI 分析'
})

const defaultTemplateName = computed(() => {
  if (props.mode === 'selection') return '基金横向对比-优劣对比模板'
  if (props.mode === 'collection' && (props.scope === 'tab' || props.scope === 'watchlist')) return '基金页签分析-关联性与缺口模板'
  if (props.mode === 'collection') return '基金持仓分析-结构诊断模板'
  if (props.mode === 'better') return '基金对比推荐分析-标准模板'
  return '基金分析-标准模板'
})

const promptOptions = computed(() => promptTemplates.value.map((item) => ({
  label: item.name || item.Name || '未命名模板',
  value: Number(item.ID || item.id || 0)
})))

const selectedPromptTemplateName = computed(() =>
  promptOptions.value.find((item) => item.value === selectedPromptTemplateId.value)?.label || defaultTemplateName.value
)

const modelNote = computed(() => (modelName.value ? `模型：${modelName.value}` : '使用当前默认 AI 配置'))
const templateNote = computed(() => `提示词模板：${templateName.value || selectedPromptTemplateName.value || defaultTemplateName.value}`)
const actionText = computed(() => (analysis.value ? '重新分析' : '开始分析'))
const emptyText = computed(() => errorMessage.value || '点击“开始分析”后，会按当前场景和所选模板生成基金 AI 分析。')

const requestTimeoutMs = computed(() => {
  if (props.mode === 'selection') {
    return EXTRA_HEAVY_TIMEOUT_MS
  }
  if (props.mode === 'collection') {
    const fundCount = normalizedFundCodes.value.length
    if (props.scope === 'tab' || props.scope === 'watchlist') {
      if (fundCount >= 20) return EXTRA_HEAVY_TIMEOUT_MS
      if (fundCount >= 8) return HEAVY_TIMEOUT_MS
    }
    return HEAVY_TIMEOUT_MS
  }
  if (props.mode === 'better') {
    return HEAVY_TIMEOUT_MS
  }
  return BASE_TIMEOUT_MS
})

const timeoutMessage = computed(() => {
  const seconds = Math.round(requestTimeoutMs.value / 1000)
  return `AI 分析超时（>${seconds} 秒），这通常是因为基金场景一次性提交的数据更大。请稍后重试，或减少当前分析的基金数量。`
})

watch(
  () => props.show,
  async (show) => {
    if (show) {
      resetState()
      await loadPromptTemplates()
      return
    }
    cancelPendingRequest()
  }
)

watch(
  () => [
    props.mode,
    props.fundCode,
    props.scope,
    props.scopeLabel,
    normalizedFundCodes.value.join(','),
    props.betterReferenceCode,
    props.betterDimension,
    props.sameTypeOnly,
    props.sameSubTypeOnly,
    props.feeFree7,
    props.feeFree30,
    props.includeAClass,
    props.onlyAClass,
    props.betterTopN
  ].join('|'),
  async () => {
    if (props.show) {
      resetState()
      await loadPromptTemplates()
    }
  }
)

async function loadPromptTemplates() {
  if (!window.go?.main?.App?.GetPromptTemplates) {
    promptTemplates.value = []
    selectedPromptTemplateId.value = null
    return
  }

  try {
    const result = (await window.go.main.App.GetPromptTemplates('', SYSTEM_PROMPT_TYPE)) || []
    promptTemplates.value = result.filter((item) => {
      const name = String(item?.name || item?.Name || '')
      return name.includes('基金') || name.includes('页签') || name.includes('持仓')
    })
  } catch (error) {
    console.error(error)
    promptTemplates.value = []
  }

  selectedPromptTemplateId.value = pickDefaultPromptId()
}

function pickDefaultPromptId() {
  const exact = promptTemplates.value.find((item) => String(item?.name || item?.Name || '') === defaultTemplateName.value)
  if (exact) {
    return Number(exact.ID || exact.id || 0)
  }
  const first = promptTemplates.value[0]
  return first ? Number(first.ID || first.id || 0) : null
}

async function runAnalysis() {
  if (loading.value) {
    return
  }

  const validationError = validateAnalysisTarget()
  if (validationError) {
    errorMessage.value = validationError
    message.warning(validationError)
    return
  }

  const currentRequest = requestSerial.value + 1
  requestSerial.value = currentRequest
  loading.value = true
  errorMessage.value = ''

  try {
    const result = await withTimeout(invokeAnalysis(), requestTimeoutMs.value, timeoutMessage.value)

    if (requestSerial.value !== currentRequest) {
      return
    }

    modelName.value = result?.model || ''
    templateName.value = result?.template || selectedPromptTemplateName.value || defaultTemplateName.value

    if (!result?.success) {
      analysis.value = ''
      errorMessage.value = result?.message || 'AI 分析失败'
      message.error(errorMessage.value)
      return
    }

    analysis.value = result?.analysis || ''
    if (!analysis.value) {
      errorMessage.value = 'AI 没有返回有效内容，请稍后再试。'
      message.warning(errorMessage.value)
    }
  } catch (error) {
    if (requestSerial.value !== currentRequest) {
      return
    }
    console.error(error)
    analysis.value = ''
    templateName.value = selectedPromptTemplateName.value || defaultTemplateName.value
    errorMessage.value = error instanceof Error ? error.message : 'AI 分析失败'
    message.error(errorMessage.value)
  } finally {
    if (requestSerial.value === currentRequest) {
      loading.value = false
    }
  }
}

function validateAnalysisTarget() {
  if (props.mode === 'single' && !String(props.fundCode || '').trim()) {
    return '当前基金代码为空，无法发起 AI 分析。'
  }
  if (props.mode === 'selection' && normalizedFundCodes.value.length < 2) {
    return '至少勾选 2 只基金后才能发起 AI 对比分析。'
  }
  if (props.mode === 'collection' && props.scope === 'tab' && normalizedFundCodes.value.length === 0) {
    return '当前页签还没有基金，无法发起 AI 分析。'
  }
  if (props.mode === 'better' && !String(props.betterReferenceCode || '').trim()) {
    return '当前推荐基金列表还没有参考基金，无法发起 AI 分析。'
  }
  return ''
}

function invokeAnalysis() {
  const app = window.go?.main?.App
  if (!app) {
    return Promise.reject(new Error('当前环境未接入基金 AI 分析。'))
  }

  const promptTemplateId = Number(selectedPromptTemplateId.value || 0)

  if (props.mode === 'collection') {
    if (typeof app.AnalyzeFundCollectionWithAI !== 'function') {
      return Promise.reject(new Error('当前版本暂不支持基金页签 AI 分析，请重新打包后再试。'))
    }
    return app.AnalyzeFundCollectionWithAI(
      props.scope || 'holdings',
      props.scopeLabel || '',
      normalizedFundCodes.value,
      0,
      promptTemplateId
    )
  }

  if (props.mode === 'selection') {
    if (typeof app.AnalyzeFundCollectionWithAI !== 'function') {
      return Promise.reject(new Error('当前版本暂不支持勾选基金 AI 对比，请重新打包后再试。'))
    }
    return app.AnalyzeFundCollectionWithAI(
      'selection',
      props.scopeLabel || '',
      normalizedFundCodes.value,
      0,
      promptTemplateId
    )
  }

  if (props.mode === 'better') {
    if (typeof app.AnalyzeBetterFundsWithAI !== 'function') {
      return Promise.reject(new Error('当前版本暂不支持推荐基金 AI 对比分析，请重新打包后再试。'))
    }
    return app.AnalyzeBetterFundsWithAI(
      {
        referenceCode: props.betterReferenceCode || '',
        sameTypeOnly: props.sameTypeOnly,
        sameSubTypeOnly: props.sameSubTypeOnly,
        dimension: props.betterDimension || 'balanced',
        feeFree7: props.feeFree7,
        feeFree30: props.feeFree30,
        includeAClass: props.includeAClass,
        onlyAClass: props.onlyAClass,
        page: 1,
        pageSize: Math.max(Number(props.betterTopN || 3), 8)
      },
      Number(props.betterTopN || 3),
      0,
      promptTemplateId
    )
  }

  if (typeof app.AnalyzeFundWithAI !== 'function') {
    return Promise.reject(new Error('当前版本暂不支持基金 AI 分析，请重新打包后再试。'))
  }
  return app.AnalyzeFundWithAI(props.fundCode || '', 0, promptTemplateId)
}

function resetState() {
  analysis.value = ''
  errorMessage.value = ''
  modelName.value = ''
  templateName.value = defaultTemplateName.value
  selectedPromptTemplateId.value = null
  loading.value = false
  requestSerial.value += 1
}

function cancelPendingRequest() {
  requestSerial.value += 1
  loading.value = false
}

function handleShowUpdate(value) {
  if (!value) {
    cancelPendingRequest()
  }
  emit('update:show', value)
}

function withTimeout(promise, timeoutMs, messageText) {
  return new Promise((resolve, reject) => {
    const timer = window.setTimeout(() => {
      reject(new Error(messageText))
    }, timeoutMs)

    promise
      .then((result) => {
        window.clearTimeout(timer)
        resolve(result)
      })
      .catch((error) => {
        window.clearTimeout(timer)
        reject(error)
      })
  })
}
</script>

<style scoped>
.analysis-shell {
  min-height: 260px;
}

.analysis-toolbar {
  margin-bottom: 14px;
  padding: 14px 16px;
  border-radius: 14px;
  border: 1px solid rgba(148, 163, 184, 0.14);
  background: rgba(15, 23, 42, 0.32);
}

.analysis-toolbar-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(280px, 320px) auto;
  gap: 14px;
  align-items: end;
}

.analysis-info-card {
  min-width: 0;
  display: grid;
  gap: 10px;
}

.analysis-info-item {
  min-width: 0;
  display: grid;
  gap: 4px;
}

.analysis-select-shell {
  min-width: 0;
  display: grid;
  gap: 6px;
  align-self: end;
}

.analysis-info-label,
.analysis-select-label {
  font-size: 12px;
  line-height: 1.2;
  color: var(--text-secondary);
}

.analysis-note,
.analysis-template {
  min-width: 0;
  font-size: 13px;
  line-height: 1.5;
  color: var(--text-primary);
  word-break: break-word;
}

.analysis-run-button {
  align-self: end;
  min-width: 112px;
}

.analysis-preview-wrap {
  max-height: 70vh;
  overflow: auto;
}

.analysis-empty {
  min-height: 220px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  text-align: center;
  border-radius: 14px;
  border: 1px dashed rgba(148, 163, 184, 0.26);
  color: var(--text-secondary);
}

@media (max-width: 900px) {
  .analysis-toolbar-grid {
    grid-template-columns: 1fr;
    align-items: stretch;
  }

  .analysis-select-shell {
    min-width: 0;
  }

  .analysis-run-button {
    width: 100%;
  }
}
</style>
