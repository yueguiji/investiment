<template>
  <div class="fade-in">
    <n-page-header
      title="量化脚本搜索 Agent"
      subtitle="先聚合平台搜索结果，再由 AI 帮你筛选最值得看的 Python 策略脚本"
    />

    <div class="platform-card" style="margin-top: 20px;">
      <n-grid :cols="4" :x-gap="12" :y-gap="12">
        <n-gi :span="2">
          <n-input
            v-model:value="query"
            placeholder="输入策略、因子、场景或脚本关键词，例如：宽基ETF 网格交易 Python"
          />
        </n-gi>
        <n-gi>
          <n-select
            v-model:value="selectedAiConfigId"
            :options="aiSourceOptions"
            placeholder="可选 AI 源"
            clearable
          />
        </n-gi>
        <n-gi>
          <n-input v-model:value="preferPlatform" placeholder="平台偏好，例如：QMT / 聚宽 / 掘金" />
        </n-gi>
      </n-grid>

      <n-grid :cols="3" :x-gap="12" :y-gap="12" style="margin-top: 12px;">
        <n-gi :span="2">
          <n-select
            v-model:value="selectedSources"
            multiple
            :options="sourceOptions"
            placeholder="选择搜索源"
          />
        </n-gi>
        <n-gi>
          <n-space style="width: 100%">
            <n-checkbox v-model:checked="requirePython">优先 Python</n-checkbox>
            <n-input-number v-model:value="resultLimit" :min="6" :max="30" style="width: 120px;" />
          </n-space>
        </n-gi>
      </n-grid>

      <div class="search-hints">
        <n-tag v-for="item in quickQueries" :key="item" size="small" round @click="query = item">{{ item }}</n-tag>
      </div>

      <n-space style="margin-top: 16px;">
        <n-button type="primary" :loading="searching" @click="handleSearch">AI 检索筛选</n-button>
        <n-button @click="handleOpenLinksOnly">仅生成搜索入口</n-button>
      </n-space>
    </div>

    <n-alert v-if="messageText" type="info" class="message-box" :show-icon="false">
      {{ messageText }}
    </n-alert>

    <n-grid :cols="2" :x-gap="16" :y-gap="16" style="margin-top: 20px;">
      <n-gi>
        <div class="platform-card result-panel">
          <div class="panel-title">AI 检索结论</div>
          <div v-if="analysis" class="analysis-box">{{ analysis }}</div>
          <div v-else class="empty-box">完成搜索后，这里会显示 AI 对候选脚本的筛选和推荐理由。</div>
        </div>
      </n-gi>
      <n-gi>
        <div class="platform-card result-panel">
          <div class="panel-title">本次检索 Prompt</div>
          <div v-if="usedPrompt" class="analysis-box">{{ usedPrompt }}</div>
          <div v-else class="empty-box">这里会展示发给 AI agent 的实际检索 Prompt。</div>
        </div>
      </n-gi>
    </n-grid>

    <div class="platform-card" style="margin-top: 20px;">
      <div class="panel-title">聚合搜索结果</div>
      <n-grid :cols="2" :x-gap="16" :y-gap="16" style="margin-top: 12px;">
        <n-gi v-for="hit in hits" :key="hit.url">
          <div class="search-card">
            <div class="search-head">
              <div>
                <div class="search-title">{{ hit.title }}</div>
                <div class="search-source">{{ hit.source }}</div>
              </div>
              <n-space>
                <n-button size="small" @click="openLink(hit.url)">打开</n-button>
                <n-button size="small" type="primary" @click="saveAsSearchSeed(hit)">入库候选</n-button>
              </n-space>
            </div>
            <div class="search-desc">{{ hit.snippet }}</div>
            <div v-if="hit.contentSnippet" class="search-extra">{{ hit.contentSnippet }}</div>
            <div v-if="hit.importHint" class="search-hint">{{ hit.importHint }}</div>
            <div v-if="hit.codePreview" class="code-preview">{{ hit.codePreview }}</div>
            <div class="search-url">{{ hit.url }}</div>
          </div>
        </n-gi>
      </n-grid>
      <div v-if="hits.length === 0" class="empty-box" style="margin-top: 12px;">
        还没有搜索结果。你可以先输入关键词，再让 Agent 帮你筛选。
      </div>
    </div>

    <div class="platform-card" style="margin-top: 20px;">
      <div class="panel-title">外部搜索入口</div>
      <n-grid :cols="2" :x-gap="16" :y-gap="16" style="margin-top: 12px;">
        <n-gi v-for="link in links" :key="link.url">
          <div class="search-card">
            <div class="search-head">
              <div class="search-title">{{ link.name }}</div>
              <n-button size="small" type="primary" @click="openLink(link.url)">打开搜索</n-button>
            </div>
            <div class="search-desc">{{ link.description }}</div>
            <div class="search-url">{{ link.url }}</div>
          </div>
        </n-gi>
      </n-grid>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useMessage } from 'naive-ui'

const message = useMessage()
const query = ref('宽基ETF 网格交易 Python')
const preferPlatform = ref('QMT')
const requirePython = ref(true)
const resultLimit = ref(12)
const searching = ref(false)
const selectedAiConfigId = ref(null)
const selectedSources = ref(['myquant', 'joinquant', 'ricequant', 'bigquant', 'github'])
const aiConfigs = ref([])
const links = ref([])
const hits = ref([])
const analysis = ref('')
const usedPrompt = ref('')
const messageText = ref('')

const quickQueries = [
  '宽基ETF 网格交易 Python',
  'ETF 轮动 网格 回测 Python',
  '多因子 低情绪 缩量 Python',
  '行业轮动 资金流 热点共振 Python',
  'QMT 可运行 ETF 网格交易',
]

const sourceOptions = [
  { label: '掘金量化', value: 'myquant' },
  { label: '聚宽', value: 'joinquant' },
  { label: 'Ricequant', value: 'ricequant' },
  { label: 'BigQuant', value: 'bigquant' },
  { label: 'GitHub', value: 'github' },
  { label: '论文/研究', value: 'papers' },
]

const aiSourceOptions = computed(() =>
  aiConfigs.value.map((item) => ({
    label: `${item.name || '未命名源'} / ${item.modelName || '未配置模型'}`,
    value: item.ID,
  }))
)

async function loadAiConfigs() {
  if (!window.go?.main?.App?.GetAiConfigs) {
    return
  }
  aiConfigs.value = (await window.go.main.App.GetAiConfigs()) || []
  if (!selectedAiConfigId.value && aiConfigs.value.length > 0) {
    selectedAiConfigId.value = aiConfigs.value[0].ID
  }
}

async function handleOpenLinksOnly() {
  if (!query.value.trim() || !window.go?.main?.App?.BuildQuantScriptSearchLinks) {
    return
  }
  links.value = (await window.go.main.App.BuildQuantScriptSearchLinks(query.value.trim())) || []
  messageText.value = '已生成外部搜索入口'
}

async function handleSearch() {
  const cleanQuery = query.value.trim()
  if (!cleanQuery || !window.go?.main?.App?.SearchQuantScriptsWithAI) {
    return
  }

  searching.value = true
  analysis.value = ''
  usedPrompt.value = ''
  hits.value = []
  messageText.value = ''

  try {
    const result = await window.go.main.App.SearchQuantScriptsWithAI(
      {
        query: cleanQuery,
        sources: selectedSources.value,
        resultLimit: resultLimit.value,
        requirePython: requirePython.value,
        preferPlatform: preferPlatform.value
      },
      selectedAiConfigId.value || 0
    )
    hits.value = result?.hits || []
    analysis.value = result?.analysis || ''
    usedPrompt.value = result?.prompt || ''
    messageText.value = result?.message || '检索完成'
    if (window.go?.main?.App?.BuildQuantScriptSearchLinks) {
      links.value = (await window.go.main.App.BuildQuantScriptSearchLinks(cleanQuery)) || []
    }
  } catch (error) {
    console.error(error)
    messageText.value = error?.message || '脚本搜索失败'
    message.error(messageText.value)
  } finally {
    searching.value = false
  }
}

async function saveAsSearchSeed(hit) {
  if (!window.go?.main?.App?.CreateQuantTemplate) {
    return
  }
  const importedCode = normalizeImportedCode(hit)
  const descriptionParts = [
    hit.source,
    hit.snippet,
    hit.contentSnippet,
    hit.importHint
  ].filter(Boolean)
  await window.go.main.App.CreateQuantTemplate({
    name: hit.title?.slice(0, 80) || `搜索候选_${new Date().toLocaleDateString('zh-CN')}`,
    description: descriptionParts.join('\n\n'),
    code: importedCode,
    brokerPlatform: preferPlatform.value || '通用',
    strategyType: 'custom',
    scriptCategory: 'stock-screener',
    searchKeywords: query.value,
    sourcePlatforms: hit.source,
    isAiGenerated: false,
    aiPrompt: '',
    aiModel: '',
    linkedStocks: '',
    status: 'draft'
  })
  message.success('已保存为程序库候选')
}

function normalizeImportedCode(hit) {
  const preview = String(hit?.codePreview || '').trim()
  if (preview) {
    return `# Source URL\n# ${hit.url}\n# Candidate Type: ${hit.candidateType || 'unknown'}\n\n${preview}\n`
  }
  return `# Source URL\n# ${hit.url}\n# Candidate Type: ${hit?.candidateType || 'unknown'}\n\n# Continue implementing the strategy based on this search candidate.\n`
}

function openLink(url) {
  if (window.go?.main?.App?.OpenURL) {
    window.go.main.App.OpenURL(url)
    return
  }
  window.open(url, '_blank')
}

onMounted(async () => {
  await loadAiConfigs()
  await handleOpenLinksOnly()
})
</script>

<style scoped>
.search-hints {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 14px;
}

.message-box {
  margin-top: 20px;
}

.result-panel {
  min-height: 240px;
}

.panel-title {
  font-size: 16px;
  font-weight: 700;
}

.analysis-box,
.empty-box {
  margin-top: 12px;
  padding: 16px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
  background: rgba(255, 255, 255, 0.03);
  white-space: pre-wrap;
  line-height: 1.7;
  color: var(--text-secondary);
}

.search-card {
  min-height: 190px;
  padding: 16px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
  background: rgba(255, 255, 255, 0.03);
}

.search-head {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: flex-start;
}

.search-title {
  font-size: 16px;
  font-weight: 700;
}

.search-source {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 4px;
}

.search-desc {
  margin-top: 12px;
  color: var(--text-secondary);
  line-height: 1.6;
  white-space: pre-wrap;
}

.search-url {
  margin-top: 14px;
  font-size: 12px;
  color: var(--text-muted);
  word-break: break-all;
}

.search-extra {
  margin-top: 10px;
  color: var(--text-secondary);
  line-height: 1.6;
  white-space: pre-wrap;
}

.search-hint {
  margin-top: 10px;
  padding: 10px 12px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
  color: var(--text-secondary);
  background: rgba(255, 255, 255, 0.04);
}

.code-preview {
  margin-top: 12px;
  max-height: 220px;
  overflow: auto;
  padding: 12px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
  background: rgba(0, 0, 0, 0.2);
  color: var(--text-secondary);
  font-family: "Cascadia Code", "Fira Code", monospace;
  font-size: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
}
</style>
