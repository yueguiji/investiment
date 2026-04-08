<template>
  <n-modal :show="show" preset="card" :title="titleText" style="width: 960px" @update:show="handleShowUpdate">
    <n-spin :show="loading">
      <div class="analysis-shell">
        <div class="analysis-toolbar">
          <div class="analysis-note">{{ modelName ? `模型：${modelName}` : '使用当前默认 AI 配置' }}</div>
          <n-button :loading="loading" @click="runAnalysis">重新分析</n-button>
        </div>
        <div v-if="analysis" class="analysis-preview-wrap">
          <MdPreview :editor-id="editorId" :model-value="analysis" theme="dark" preview-theme="github" />
        </div>
        <div v-else class="analysis-empty">
          {{ errorMessage || '点击“重新分析”生成基金 AI 分析。' }}
        </div>
      </div>
    </n-spin>
  </n-modal>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useMessage } from 'naive-ui'
import { MdPreview } from 'md-editor-v3'
import { AnalyzeFundCollectionWithAI, AnalyzeFundWithAI } from '../../../../wailsjs/go/main/App'

const props = defineProps({
  show: { type: Boolean, default: false },
  mode: { type: String, default: 'single' },
  fundCode: { type: String, default: '' },
  scope: { type: String, default: 'holdings' },
  title: { type: String, default: '' }
})

const emit = defineEmits(['update:show'])

const message = useMessage()
const loading = ref(false)
const analysis = ref('')
const errorMessage = ref('')
const modelName = ref('')

const editorId = computed(() => `fund-ai-${props.mode}-${props.fundCode || props.scope || 'panel'}`)
const titleText = computed(() => props.title || (props.mode === 'collection' ? '组合 AI 分析' : '基金 AI 分析'))

watch(
  () => props.show,
  async (show) => {
    if (!show) return
    await runAnalysis()
  }
)

async function runAnalysis() {
  loading.value = true
  errorMessage.value = ''
  try {
    const result = props.mode === 'collection'
      ? await AnalyzeFundCollectionWithAI(props.scope || 'holdings', 0)
      : await AnalyzeFundWithAI(props.fundCode || '', 0)

    if (!result?.success) {
      analysis.value = ''
      modelName.value = result?.model || ''
      errorMessage.value = result?.message || 'AI 分析失败'
      message.error(errorMessage.value)
      return
    }

    analysis.value = result.analysis || ''
    modelName.value = result.model || ''
  } catch (error) {
    console.error(error)
    analysis.value = ''
    errorMessage.value = 'AI 分析失败'
    message.error(errorMessage.value)
  } finally {
    loading.value = false
  }
}

function handleShowUpdate(value) {
  emit('update:show', value)
}
</script>

<style scoped>
.analysis-shell {
  min-height: 260px;
}

.analysis-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.analysis-note {
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
  border-radius: 14px;
  border: 1px dashed rgba(148, 163, 184, 0.26);
  color: var(--text-secondary);
}
</style>
