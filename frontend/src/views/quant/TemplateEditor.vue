<template>
  <div class="fade-in">
    <n-page-header :title="isEdit ? '编辑量化脚本' : '新建量化脚本'" subtitle="维护脚本代码、标签画像、适配场景和搜索关键词">
      <template #extra>
        <n-space>
          <n-button @click="router.push('/quant/templates')">返回程序库</n-button>
          <n-button type="primary" secondary @click="openAiGenerator">{{ isEdit ? 'AI 迭代修改' : 'AI 生成脚本' }}</n-button>
          <n-button v-if="isEdit" @click="handleExport">导出 .py</n-button>
          <n-button v-if="isEdit && form.ID" type="success" @click="handleActivate">设为启用</n-button>
          <n-button type="primary" :loading="saving" @click="handleSave">保存脚本</n-button>
        </n-space>
      </template>
    </n-page-header>

    <n-alert v-if="isEdit" type="info" :show-icon="false" style="margin-top: 16px;">
      当前脚本可以直接发给 AI 做增量修改。点击右上角“AI 迭代修改”后，会自动带入现有代码和说明。
    </n-alert>

    <n-grid :cols="2" :x-gap="16" style="margin-top: 20px;">
      <n-gi>
        <div class="platform-card">
          <n-form label-placement="top">
            <n-form-item label="脚本名称">
              <n-input v-model:value="form.name" placeholder="例如：A股多因子轮动_情绪增强版" />
            </n-form-item>

            <n-grid :cols="2" :x-gap="12">
              <n-gi>
                <n-form-item label="脚本分类">
                  <n-select v-model:value="form.scriptCategory" :options="scriptCategoryOptions" />
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item label="策略类型">
                  <n-select v-model:value="form.strategyType" :options="strategyTypeOptions" />
                </n-form-item>
              </n-gi>
            </n-grid>

            <n-grid :cols="2" :x-gap="12">
              <n-gi>
                <n-form-item label="平台适配">
                  <n-input v-model:value="form.brokerPlatform" placeholder="QMT / 掘金 / 聚宽 / 通用" />
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item label="状态">
                  <n-select v-model:value="form.status" :options="statusOptions" />
                </n-form-item>
              </n-gi>
            </n-grid>

            <n-form-item label="描述">
              <n-input v-model:value="form.description" type="textarea" :rows="4" placeholder="说明策略目标、适用行情、核心逻辑。" />
            </n-form-item>

            <n-form-item label="关联标的">
              <n-input v-model:value="form.linkedStocks" placeholder="例如：sh600519, sz000858" />
            </n-form-item>

            <n-form-item label="搜索关键词">
              <n-input v-model:value="form.searchKeywords" placeholder="例如：多因子, 轮动, 情绪增强, 中盘成长" />
            </n-form-item>

            <n-form-item label="来源平台">
              <n-select v-model:value="sourcePlatforms" multiple filterable tag :options="sourcePlatformOptions" placeholder="记录灵感来源平台" />
            </n-form-item>
          </n-form>
        </div>
      </n-gi>

      <n-gi>
        <div class="platform-card">
          <div class="tag-grid">
            <div v-for="group in editableGroups" :key="group.key" class="tag-group">
              <div class="tag-group-title">{{ group.label }}</div>
              <div class="tag-group-desc">{{ group.description }}</div>
              <n-select
                v-model:value="groupValues[group.key]"
                multiple
                filterable
                tag
                :options="groupOptions(group.key)"
                :placeholder="`选择${group.label}`"
              />
            </div>
          </div>
        </div>
      </n-gi>
    </n-grid>

    <div class="platform-card" style="margin-top: 20px;">
      <div class="code-header">
        <div>
          <div class="code-title">Python 策略代码</div>
          <div class="code-subtitle">建议包含参数区、因子计算、信号逻辑、仓位管理和风控。</div>
        </div>
        <n-space>
          <n-tag round>{{ form.language || 'python' }}</n-tag>
          <n-tag round type="info">v{{ form.version || 1 }}</n-tag>
        </n-space>
      </div>
      <n-input v-model:value="form.code" type="textarea" :autosize="{ minRows: 20, maxRows: 30 }" placeholder="请输入 Python 量化脚本代码" />
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useDialog, useMessage } from 'naive-ui'
import { useRoute, useRouter } from 'vue-router'
import { buildTagOptions, joinTagArray, splitTagString } from '../../utils/quant'

const AI_DRAFT_KEY = 'investment.quant.ai-generator.draft'

const route = useRoute()
const router = useRouter()
const dialog = useDialog()
const message = useMessage()
const saving = ref(false)
const taxonomy = ref([])

const form = reactive({
  ID: 0,
  name: '',
  categoryId: 0,
  description: '',
  language: 'python',
  code: '',
  brokerPlatform: '通用',
  strategyType: 'multi-factor',
  scriptCategory: 'strategy-main',
  styleTags: '',
  emotionTags: '',
  volumeTags: '',
  scenarioTags: '',
  capitalTags: '',
  factorTags: '',
  searchKeywords: '',
  sourcePlatforms: '',
  isAiGenerated: false,
  aiPrompt: '',
  aiModel: '',
  linkedStocks: '',
  parameters: '',
  backtestResult: '',
  version: 1,
  status: 'draft',
  lastUsedAt: ''
})

const groupValues = reactive({
  style: [],
  emotion: [],
  volume: [],
  scenario: [],
  capital: [],
  factor: []
})

const sourcePlatforms = ref([])

const sourcePlatformOptions = [
  { label: '掘金量化', value: '掘金量化' },
  { label: '聚宽', value: '聚宽' },
  { label: 'Ricequant', value: 'Ricequant' },
  { label: 'BigQuant', value: 'BigQuant' },
  { label: 'GitHub', value: 'GitHub' },
  { label: '本地原创', value: '本地原创' }
]

const strategyTypeOptions = [
  { label: '多因子', value: 'multi-factor' },
  { label: '网格交易', value: 'grid-trading' },
  { label: '趋势跟随', value: 'trend-following' },
  { label: '均值回归', value: 'mean-reversion' },
  { label: '事件驱动', value: 'event-driven' },
  { label: '套利', value: 'arbitrage' },
  { label: '自定义', value: 'custom' }
]

const statusOptions = [
  { label: '草稿', value: 'draft' },
  { label: '启用', value: 'active' },
  { label: '归档', value: 'archived' }
]

const isEdit = computed(() => Boolean(route.params.id))
const scriptCategoryOptions = computed(() => groupOptions('script-category'))
const editableGroups = computed(() => taxonomy.value.filter((item) => ['style', 'emotion', 'volume', 'scenario', 'capital', 'factor'].includes(item.key)))

function groupOptions(key) {
  const group = taxonomy.value.find((item) => item.key === key)
  return buildTagOptions(group ? [group] : [])
}

function applyTemplate(template) {
  Object.assign(form, {
    ID: template.ID || 0,
    name: template.name || '',
    categoryId: template.categoryId || 0,
    description: template.description || '',
    language: template.language || 'python',
    code: template.code || '',
    brokerPlatform: template.brokerPlatform || '通用',
    strategyType: template.strategyType || 'multi-factor',
    scriptCategory: template.scriptCategory || 'strategy-main',
    styleTags: template.styleTags || '',
    emotionTags: template.emotionTags || '',
    volumeTags: template.volumeTags || '',
    scenarioTags: template.scenarioTags || '',
    capitalTags: template.capitalTags || '',
    factorTags: template.factorTags || '',
    searchKeywords: template.searchKeywords || '',
    sourcePlatforms: template.sourcePlatforms || '',
    isAiGenerated: !!template.isAiGenerated,
    aiPrompt: template.aiPrompt || '',
    aiModel: template.aiModel || '',
    linkedStocks: template.linkedStocks || '',
    parameters: template.parameters || '',
    backtestResult: template.backtestResult || '',
    version: template.version || 1,
    status: template.status || 'draft',
    lastUsedAt: template.lastUsedAt || ''
  })

  groupValues.style = splitTagString(template.styleTags)
  groupValues.emotion = splitTagString(template.emotionTags)
  groupValues.volume = splitTagString(template.volumeTags)
  groupValues.scenario = splitTagString(template.scenarioTags)
  groupValues.capital = splitTagString(template.capitalTags)
  groupValues.factor = splitTagString(template.factorTags)
  sourcePlatforms.value = splitTagString(template.sourcePlatforms)
}

function syncGroupValuesToForm() {
  form.styleTags = joinTagArray(groupValues.style)
  form.emotionTags = joinTagArray(groupValues.emotion)
  form.volumeTags = joinTagArray(groupValues.volume)
  form.scenarioTags = joinTagArray(groupValues.scenario)
  form.capitalTags = joinTagArray(groupValues.capital)
  form.factorTags = joinTagArray(groupValues.factor)
  form.sourcePlatforms = joinTagArray(sourcePlatforms.value)
}

async function loadTaxonomy() {
  if (window.go?.main?.App?.GetQuantTagTaxonomy) {
    taxonomy.value = (await window.go.main.App.GetQuantTagTaxonomy()) || []
  }
}

async function loadTemplate() {
  if (!isEdit.value || !window.go?.main?.App?.GetQuantTemplate) {
    return
  }
  const template = await window.go.main.App.GetQuantTemplate(Number(route.params.id))
  if (!template) {
    message.error('没有找到对应脚本')
    router.push('/quant/templates')
    return
  }
  applyTemplate(template)
}

function buildAiDraft() {
  syncGroupValuesToForm()
  return {
    mode: isEdit.value ? 'edit' : 'create',
    sourceTemplateId: form.ID || 0,
    sourceTemplateName: form.name || '',
    sourceDescription: form.description || '',
    strategyDescription: form.description || '',
    brokerPlatform: form.brokerPlatform || 'QMT',
    strategyType: form.strategyType || 'multi-factor',
    scriptCategory: form.scriptCategory || 'strategy-main',
    stockCodes: form.linkedStocks || '',
    factorTags: form.factorTags || '',
    sceneTags: form.scenarioTags || '',
    baseCode: form.code || '',
  }
}

function openAiGenerator() {
  sessionStorage.setItem(AI_DRAFT_KEY, JSON.stringify(buildAiDraft()))
  const query = isEdit.value ? `mode=edit&sourceTemplateId=${form.ID}` : 'mode=create'
  router.push(`/quant/ai-generate?${query}`)
}

async function handleSave() {
  syncGroupValuesToForm()
  if (!form.name.trim()) {
    message.warning('请先填写脚本名称')
    return
  }
  if (!form.code.trim()) {
    message.warning('请先填写 Python 代码')
    return
  }

  saving.value = true
  try {
    if (form.ID && window.go?.main?.App?.UpdateQuantTemplate) {
      const saved = await window.go.main.App.UpdateQuantTemplate({ ...form })
      if (saved) {
        applyTemplate(saved)
      }
      message.success('脚本已更新')
    } else if (window.go?.main?.App?.CreateQuantTemplate) {
      const created = await window.go.main.App.CreateQuantTemplate({ ...form })
      if (created?.ID) {
        applyTemplate(created)
        router.replace(`/quant/editor/${created.ID}`)
      }
      message.success('脚本已创建')
    }
  } catch (error) {
    console.error(error)
    message.error('保存失败')
  } finally {
    saving.value = false
  }
}

async function handleActivate() {
  if (!form.ID || !window.go?.main?.App?.ActivateQuantTemplate) {
    return
  }
  const ok = await window.go.main.App.ActivateQuantTemplate(form.ID)
  if (ok) {
    form.status = 'active'
    message.success('已设为启用')
  } else {
    message.error('启用失败')
  }
}

async function handleExport() {
  if (!form.ID || !window.go?.main?.App?.ExportQuantTemplate) {
    return
  }
  try {
    const path = await window.go.main.App.ExportQuantTemplate(form.ID)
    message.success(`已导出到 ${path}`)
  } catch (error) {
    console.error(error)
    message.error('导出失败')
  }
}

function handleDelete() {
  if (!form.ID || !window.go?.main?.App?.DeleteQuantTemplate) {
    return
  }
  dialog.warning({
    title: '删除脚本',
    content: '删除后不可恢复，确认继续吗？',
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      const ok = await window.go.main.App.DeleteQuantTemplate(form.ID)
      if (ok) {
        message.success('脚本已删除')
        router.push('/quant/templates')
      } else {
        message.error('删除失败')
      }
    }
  })
}

onMounted(async () => {
  await loadTaxonomy()
  await loadTemplate()
})
</script>

<style scoped>
.tag-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 14px;
}

.tag-group-title {
  font-weight: 600;
  margin-bottom: 4px;
}

.tag-group-desc {
  font-size: 12px;
  color: var(--text-muted);
  margin-bottom: 8px;
}

.code-header {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: center;
  margin-bottom: 12px;
}

.code-title {
  font-weight: 700;
}

.code-subtitle {
  color: var(--text-muted);
  font-size: 12px;
  margin-top: 4px;
}
</style>
