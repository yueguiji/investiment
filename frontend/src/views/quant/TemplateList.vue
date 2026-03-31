<template>
  <div class="fade-in">
    <n-page-header title="量化程序库" subtitle="集中管理 Python 量化脚本、标签、因子和适配场景">
      <template #extra>
        <n-space>
          <n-button @click="$router.push('/quant/search')">脚本搜索</n-button>
          <n-button @click="$router.push('/quant/editor')">手动新建</n-button>
          <n-button type="primary" @click="$router.push('/quant/ai-generate?mode=create')">AI 新建脚本</n-button>
        </n-space>
      </template>
    </n-page-header>

    <div class="platform-card" style="margin-top: 20px;">
      <n-grid :cols="4" :x-gap="12" :y-gap="12">
        <n-gi>
          <n-input v-model:value="filters.keyword" placeholder="搜索脚本名、标签、关键词" clearable />
        </n-gi>
        <n-gi>
          <n-select v-model:value="filters.scriptCategory" :options="scriptCategoryOptions" placeholder="脚本分类" clearable />
        </n-gi>
        <n-gi>
          <n-select v-model:value="filters.status" :options="statusOptions" placeholder="状态" clearable />
        </n-gi>
        <n-gi>
          <n-select v-model:value="filters.styleTag" :options="styleTagOptions" placeholder="策略类型标签" clearable />
        </n-gi>
      </n-grid>

      <div class="stats-row">
        <div class="stat-box">
          <div class="stat-label">脚本总数</div>
          <div class="stat-value">{{ templates.length }}</div>
        </div>
        <div class="stat-box">
          <div class="stat-label">启用中</div>
          <div class="stat-value profit">{{ activeCount }}</div>
        </div>
        <div class="stat-box">
          <div class="stat-label">AI 生成</div>
          <div class="stat-value">{{ aiCount }}</div>
        </div>
        <div class="stat-box">
          <div class="stat-label">筛选结果</div>
          <div class="stat-value">{{ filteredTemplates.length }}</div>
        </div>
      </div>
    </div>

    <n-alert v-if="loadError" type="error" class="load-error" :show-icon="false">
      {{ loadError }}
    </n-alert>

    <n-grid :cols="3" :x-gap="16" :y-gap="16" style="margin-top: 20px;">
      <n-gi v-for="template in filteredTemplates" :key="template.ID || template.id || template.name">
        <div class="template-card platform-card" @click="openTemplate(template)">
          <div class="template-head">
            <div>
              <div class="template-title">{{ safeText(template.name, '未命名脚本') }}</div>
              <div class="template-meta">
                {{ safeText(template.scriptCategory, 'strategy-main') }} / {{ safeText(template.strategyType, 'custom') }} / v{{ template.version || 1 }}
              </div>
            </div>
            <div class="template-actions">
              <n-tag :type="statusType(template.status)" size="small" round>{{ statusLabel(template.status) }}</n-tag>
              <n-space size="small" vertical align="end">
                <n-button quaternary size="tiny" @click.stop="openTemplate(template)">编辑</n-button>
                <n-button quaternary type="primary" size="tiny" @click.stop="openInAi(template)">AI 迭代</n-button>
                <n-button quaternary type="error" size="tiny" @click.stop="handleDelete(template)">删除</n-button>
              </n-space>
            </div>
          </div>

          <div class="template-desc">{{ safeText(template.description, '暂无说明，建议补充适配场景和参数说明。') }}</div>

          <div class="tag-area">
            <n-tag
              v-for="tag in formatVisibleTags(template)"
              :key="`${template.ID || template.id}-${tag}`"
              size="small"
              :bordered="false"
              type="info"
            >
              {{ formatTag(tag) }}
            </n-tag>
          </div>

          <div class="template-foot">
            <span>{{ safeText(template.brokerPlatform, '通用平台') }}</span>
            <span>{{ formatTime(template.UpdatedAt || template.updatedAt) }}</span>
          </div>
        </div>
      </n-gi>

      <n-gi v-if="filteredTemplates.length === 0" :span="3">
        <div class="platform-card empty-state">
          <div class="empty-icon">Q</div>
          <div class="empty-title">暂时没有匹配的量化脚本</div>
          <div class="empty-subtitle">可以先让 AI 帮你生成草稿，或者去脚本搜索页找灵感后再二次整理。</div>
          <n-space justify="center">
            <n-button type="primary" @click="$router.push('/quant/ai-generate?mode=create')">AI 新建脚本</n-button>
            <n-button @click="$router.push('/quant/editor')">手动新建</n-button>
            <n-button @click="$router.push('/quant/search')">去脚本搜索</n-button>
          </n-space>
        </div>
      </n-gi>
    </n-grid>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useDialog, useMessage } from 'naive-ui'
import { collectTemplateTags, getTagLabel } from '../../utils/quant'

const AI_DRAFT_KEY = 'investment.quant.ai-generator.draft'

const message = useMessage()
const dialog = useDialog()
const templates = ref([])
const taxonomy = ref([])
const loadError = ref('')
const filters = ref({
  keyword: '',
  scriptCategory: null,
  status: null,
  styleTag: null,
})

const statusOptions = [
  { label: '草稿', value: 'draft' },
  { label: '启用', value: 'active' },
  { label: '归档', value: 'archived' },
]

const scriptCategoryOptions = computed(() => {
  const group = taxonomy.value.find((item) => item.key === 'script-category')
  return (group?.options || []).map((item) => ({ label: item.label, value: item.value }))
})

const styleTagOptions = computed(() => {
  const group = taxonomy.value.find((item) => item.key === 'style')
  return (group?.options || []).map((item) => ({ label: item.label, value: item.value }))
})

const filteredTemplates = computed(() => {
  const keyword = String(filters.value.keyword || '').trim().toLowerCase()
  return templates.value.filter((item) => {
    try {
      const tags = collectTemplateTags(item)
      const haystack = [
        item?.name,
        item?.description,
        item?.searchKeywords,
        item?.scriptCategory,
        item?.strategyType,
        ...tags,
      ]
        .map((value) => safeText(value))
        .join(' ')
        .toLowerCase()

      if (keyword && !haystack.includes(keyword)) {
        return false
      }
      if (filters.value.scriptCategory && item?.scriptCategory !== filters.value.scriptCategory) {
        return false
      }
      if (filters.value.status && item?.status !== filters.value.status) {
        return false
      }
      if (filters.value.styleTag && !tags.includes(filters.value.styleTag)) {
        return false
      }
      return true
    } catch (error) {
      console.error('filter template failed', item, error)
      return false
    }
  })
})

const activeCount = computed(() => templates.value.filter((item) => item?.status === 'active').length)
const aiCount = computed(() => templates.value.filter((item) => item?.isAiGenerated).length)

function safeText(value, fallback = '') {
  if (value === null || value === undefined) {
    return fallback
  }
  const text = String(value).trim()
  return text || fallback
}

function normalizeTemplates(result) {
  if (Array.isArray(result)) {
    if (Array.isArray(result[0])) {
      return result[0]
    }
    return result
  }
  if (result && Array.isArray(result.list)) {
    return result.list
  }
  if (result && Array.isArray(result.data)) {
    return result.data
  }
  return []
}

function statusType(status) {
  return { draft: 'default', active: 'success', archived: 'warning' }[status] || 'default'
}

function statusLabel(status) {
  return { draft: '草稿', active: '启用', archived: '归档' }[status] || safeText(status, '未知状态')
}

function formatTag(value) {
  return getTagLabel(taxonomy.value, value)
}

function formatVisibleTags(template) {
  try {
    return collectTemplateTags(template).slice(0, 8)
  } catch (error) {
    console.error('collect tags failed', template, error)
    return []
  }
}

function formatTime(value) {
  if (!value) {
    return '未更新'
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return '未更新'
  }
  return `更新于 ${date.toLocaleDateString('zh-CN')}`
}

function openTemplate(template) {
  const id = template?.ID || template?.id
  if (!id) {
    message.warning('这条脚本记录缺少 ID，暂时无法打开编辑页')
    return
  }
  window.location.hash = `#/quant/editor/${id}`
}

function openInAi(template) {
  const id = template?.ID || template?.id
  if (!id) {
    message.warning('这条脚本记录缺少 ID，暂时无法发起 AI 迭代')
    return
  }

  const draft = {
    mode: 'edit',
    sourceTemplateId: id,
    sourceTemplateName: safeText(template?.name, ''),
    sourceDescription: safeText(template?.description, ''),
    baseCode: safeText(template?.code, ''),
    brokerPlatform: safeText(template?.brokerPlatform, 'QMT'),
    strategyType: safeText(template?.strategyType, 'multi-factor'),
    scriptCategory: safeText(template?.scriptCategory, 'strategy-main'),
    stockCodes: safeText(template?.linkedStocks, ''),
    factorTags: safeText(template?.factorTags, ''),
    sceneTags: safeText(template?.scenarioTags, ''),
    strategyDescription: safeText(template?.description, ''),
  }
  sessionStorage.setItem(AI_DRAFT_KEY, JSON.stringify(draft))
  window.location.hash = `#/quant/ai-generate?mode=edit&sourceTemplateId=${id}`
}

function handleDelete(template) {
  const id = template?.ID || template?.id
  if (!id || !window.go?.main?.App?.DeleteQuantTemplate) {
    message.warning('当前脚本无法删除，请刷新后重试')
    return
  }

  dialog.warning({
    title: '删除脚本',
    content: `确认删除“${safeText(template?.name, '未命名脚本')}”吗？删除后不可恢复。`,
    positiveText: '确认删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        const ok = await window.go.main.App.DeleteQuantTemplate(id)
        if (!ok) {
          message.error('删除失败')
          return
        }
        message.success('脚本已删除')
        await loadData()
      } catch (error) {
        console.error(error)
        message.error('删除失败')
      }
    }
  })
}

async function loadData() {
  loadError.value = ''
  try {
    if (window.go?.main?.App?.GetQuantTemplates) {
      const result = await window.go.main.App.GetQuantTemplates(0, '', 1, 200)
      templates.value = normalizeTemplates(result)
    }
    if (window.go?.main?.App?.GetQuantTagTaxonomy) {
      taxonomy.value = (await window.go.main.App.GetQuantTagTaxonomy()) || []
    }
  } catch (error) {
    console.error(error)
    templates.value = []
    loadError.value = error?.message || '量化程序库加载失败'
  }
}

onMounted(loadData)
</script>

<style scoped>
.load-error {
  margin-top: 20px;
}

.stats-row {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
  margin-top: 16px;
}

.stat-box {
  padding: 14px 16px;
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--border-color);
}

.stat-label {
  color: var(--text-secondary);
  font-size: 12px;
  margin-bottom: 8px;
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
}

.template-card {
  cursor: pointer;
  min-height: 230px;
}

.template-card:hover {
  border-color: var(--primary);
  box-shadow: var(--shadow-glow);
}

.template-head {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: flex-start;
}

.template-actions {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 8px;
}

.template-title {
  font-size: 16px;
  font-weight: 700;
  line-height: 1.5;
  word-break: break-word;
}

.template-meta {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 6px;
}

.template-desc {
  margin-top: 14px;
  color: var(--text-secondary);
  line-height: 1.6;
  min-height: 64px;
  white-space: pre-wrap;
  word-break: break-word;
}

.tag-area {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 16px;
}

.template-foot {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-top: 18px;
  color: var(--text-muted);
  font-size: 12px;
}

.empty-state {
  text-align: center;
  padding: 72px 24px;
}

.empty-icon {
  font-size: 40px;
  font-weight: 700;
  margin-bottom: 16px;
}

.empty-title {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 8px;
}

.empty-subtitle {
  color: var(--text-secondary);
  margin-bottom: 20px;
}
</style>
