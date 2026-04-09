<template>
  <n-modal :show="show" preset="card" :title="titleText" style="width: 960px" @update:show="handleShowUpdate">
    <div class="analysis-shell">
      <div class="analysis-toolbar">
        <div class="analysis-meta">
          <div class="analysis-note">{{ modelNote }}</div>
          <div class="analysis-template">{{ templateNote }}</div>
        </div>
        <n-button type="primary" :loading="loading" @click="runAnalysis">
          {{ actionText }}
        </n-button>
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
import { AnalyzeFundCollectionWithAI, AnalyzeFundWithAI } from '../../../../wailsjs/go/main/App'

const REQUEST_TIMEOUT_MS = 45000

const props = defineProps({
  show: { type: Boolean, default: false },
  mode: { type: String, default: 'single' },
  fundCode: { type: String, default: '' },
  scope: { type: String, default: 'holdings' },
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

const editorId = computed(() => [
  'fund-ai',
  props.mode,
  props.fundCode || props.scope || props.betterReferenceCode || 'panel'
].join('-'))

const titleText = computed(() => {
  if (props.title) return props.title
  if (props.mode === 'collection') return '组合 AI 分析'
  if (props.mode === 'better') return '推荐基金 AI 分析'
  return '基金 AI 分析'
})

const defaultTemplateName = computed(() => {
  if (props.mode === 'collection') return '基金组合分析-标准模板'
  if (props.mode === 'better') return '基金对比推荐分析-标准模板'
  return '基金分析-标准模板'
})

const modelNote = computed(() => modelName.value ? `模型：${modelName.value}` : '使用当前默认 AI 配置')
const templateNote = computed(() => `提示词模板：${templateName.value || defaultTemplateName.value}`)
const actionText = computed(() => (analysis.value ? '重新分析' : '开始分析'))
const emptyText = computed(() => errorMessage.value || '点击“开始分析”后，按内置模板生成基金 AI 分析。')

watch(
  () => props.show,
  (show) => {
    if (show) {
      resetState()
      return
    }
    cancelPendingRequest()
  }
)

watch(
  () => `${props.mode}|${props.fundCode}|${props.scope}|${props.betterReferenceCode}|${props.betterDimension}|${props.sameTypeOnly}|${props.sameSubTypeOnly}|${props.feeFree7}|${props.feeFree30}|${props.includeAClass}|${props.onlyAClass}|${props.betterTopN}`,
  () => {
    if (props.show) {
      resetState()
    }
  }
)

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
    const result = await withTimeout(invokeAnalysis(), REQUEST_TIMEOUT_MS, 'AI 分析超时，请稍后重试或检查模型配置。')

    if (requestSerial.value !== currentRequest) {
      return
    }

    modelName.value = result?.model || ''
    templateName.value = result?.template || defaultTemplateName.value

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
    templateName.value = defaultTemplateName.value
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
  if (props.mode === 'better' && !String(props.betterReferenceCode || '').trim()) {
    return '当前推荐基金列表还没有参考基金，无法发起 AI 分析。'
  }
  return ''
}

function invokeAnalysis() {
  if (props.mode === 'collection') {
    return AnalyzeFundCollectionWithAI(props.scope || 'holdings', 0)
  }
  if (props.mode === 'better') {
    const analyzer = window.go?.main?.App?.AnalyzeBetterFundsWithAI
    if (typeof analyzer !== 'function') {
      return Promise.reject(new Error('当前版本暂不支持推荐基金 AI 分析，请重新打包后再试。'))
    }
    return analyzer(
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
      0
    )
  }
  return AnalyzeFundWithAI(props.fundCode || '', 0)
}

function resetState() {
  analysis.value = ''
  errorMessage.value = ''
  modelName.value = ''
  templateName.value = defaultTemplateName.value
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
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 14px;
}

.analysis-meta {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.analysis-note,
.analysis-template {
  font-size: 12px;
  color: var(--text-secondary);
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
</style>
