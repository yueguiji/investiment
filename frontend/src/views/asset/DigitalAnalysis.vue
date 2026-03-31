<template>
  <div class="fade-in digital-analysis-page">
    <n-page-header title="数字分析" subtitle="围绕家庭资产、负债、收入、成员与基准数据的结构化分析和连续问答">
      <template #extra>
        <n-space>
          <n-button @click="$router.push('/asset/overview')">返回总览</n-button>
          <n-button @click="$router.push('/asset/members')">家庭成员</n-button>
          <n-button @click="$router.push('/asset/benchmarks')">基准数据</n-button>
          <n-button secondary @click="loadBootstrap">刷新数据</n-button>
          <n-button type="primary" :loading="runningAnalysis" @click="runAnalysis('manual')">手动重新分析</n-button>
        </n-space>
      </template>
    </n-page-header>

    <n-grid :cols="3" :x-gap="16" :y-gap="16" style="margin-top: 20px;">
      <n-gi>
        <div class="metric-card">
          <div class="metric-label">当前地区</div>
          <div class="metric-value">{{ region }}</div>
          <div class="metric-sub">数字分析和地区对比默认优先使用当前地区口径。</div>
        </div>
      </n-gi>
      <n-gi>
        <div class="metric-card">
          <div class="metric-label">分析模型</div>
          <div class="metric-model">{{ selectedAiLabel }}</div>
          <div class="metric-sub">统一使用当前 AI 源生成报告和连续问答。</div>
        </div>
      </n-gi>
      <n-gi>
        <div class="metric-card">
          <div class="metric-label">最近状态</div>
          <div class="metric-value">
            <span class="status-chip" :class="analysisStatusTone(latestRecord?.status)">{{ analysisStatusLabel(latestRecord?.status) }}</span>
          </div>
          <div class="metric-sub">{{ latestStatusText }}</div>
        </div>
      </n-gi>
    </n-grid>

    <div class="platform-card" style="margin-top: 20px;">
      <div class="section-row">
        <div>
          <div class="section-title">大致水平</div>
          <div class="section-sub">基于当前净资产、负债率、家庭成员和地区基准做粗略层级判断，仅用于辅助观察。</div>
        </div>
      </div>
      <n-grid :cols="2" :x-gap="16" :y-gap="16">
        <n-gi v-for="item in rankingCards" :key="item.label">
          <div class="mini-card">
            <div class="mini-label">{{ item.label }}</div>
            <div class="mini-value">{{ item.value }}</div>
            <div class="mini-sub">{{ item.sub }}</div>
          </div>
        </n-gi>
      </n-grid>
    </div>

    <div class="platform-card" style="margin-top: 20px;">
      <div class="section-row">
        <div>
          <div class="section-title">AI分析</div>
          <div class="section-sub">已自动注入当前家庭资产、负债、收入、保障、成员和地区基准，你可以直接继续追问。</div>
        </div>
      </div>

      <div class="context-banner">
        <div class="context-title">本轮已注入家庭快照</div>
        <div class="context-text">{{ contextDigest }}</div>
        <div class="context-grid">
          <div class="context-block">
            <div class="context-block-title">家庭成员</div>
            <div v-if="memberDigest.length" class="context-list">
              <div v-for="item in memberDigest" :key="item" class="context-item">{{ item }}</div>
            </div>
            <div v-else class="context-empty">暂无成员明细</div>
          </div>
          <div class="context-block">
            <div class="context-block-title">主要资产明细</div>
            <div v-if="assetDigest.length" class="context-list">
              <div v-for="item in assetDigest" :key="item" class="context-item">{{ item }}</div>
            </div>
            <div v-else class="context-empty">暂无资产明细</div>
          </div>
          <div class="context-block">
            <div class="context-block-title">月收入明细</div>
            <div v-if="incomeDigest.length" class="context-list">
              <div v-for="item in incomeDigest" :key="item" class="context-item">{{ item }}</div>
            </div>
            <div v-else class="context-empty">暂无收入明细</div>
          </div>
          <div class="context-block">
            <div class="context-block-title">负债明细</div>
            <div v-if="liabilityDigest.length" class="context-list">
              <div v-for="item in liabilityDigest" :key="item" class="context-item">{{ item }}</div>
            </div>
            <div v-else class="context-empty">暂无负债明细</div>
          </div>
          <div class="context-block">
            <div class="context-block-title">流动资金趋势</div>
            <div v-if="liquidTrendDigest.length" class="context-list">
              <div v-for="item in liquidTrendDigest" :key="item" class="context-item">{{ item }}</div>
            </div>
            <div v-else class="context-empty">暂无流动资金趋势</div>
          </div>
          <div class="context-block">
            <div class="context-block-title">流动资金分布</div>
            <div v-if="liquidDistributionDigest.length" class="context-list">
              <div v-for="item in liquidDistributionDigest" :key="item" class="context-item">{{ item }}</div>
            </div>
            <div v-else class="context-empty">暂无流动资金分布</div>
          </div>
        </div>
      </div>

      <div class="chat-shell">
        <div ref="messageContainerRef" class="message-list">
          <div
            v-for="(item, index) in chatMessages"
            :key="`${item.role}-${index}`"
            class="message-item"
            :class="item.role"
          >
            <div class="message-header">
              <span>{{ item.role === 'user' ? '我' : '数字分析助手' }}</span>
              <span>{{ item.datetime }}</span>
            </div>
            <div v-if="item.role === 'assistant'" class="message-content">
              <MdPreview :editor-id="`household-chat-${index}`" :model-value="item.content" theme="dark" preview-theme="github" />
            </div>
            <div v-else class="message-content plain">{{ item.content }}</div>
          </div>
        </div>

        <div class="chat-input">
          <n-input
            v-model:value="chatInput"
            type="textarea"
            :autosize="{ minRows: 3, maxRows: 5 }"
            placeholder="例如：以我目前的净资产、月供和家庭结构来看，未来三年最该盯住哪几个数字？"
            @keydown.enter.exact.prevent="sendChat"
          />
          <div class="chat-actions">
            <n-button @click="resetChat">清空对话</n-button>
            <n-button type="primary" :loading="chatLoading" @click="sendChat">发送问题</n-button>
          </div>
        </div>
      </div>
    </div>

    <div class="platform-card" style="margin-top: 20px;">
      <div class="section-row">
        <div>
          <div class="section-title">最近分析结果</div>
          <div class="section-sub">结构化报告按纵向阅读展示，便于复盘核心结论、指标、风险和建议。</div>
        </div>
      </div>

      <n-alert v-if="statusMessage" :type="statusType" :show-icon="false" style="margin-bottom: 16px;">
        {{ statusMessage }}
      </n-alert>

      <n-descriptions v-if="latestRecord" label-placement="left" :column="3" bordered style="margin-bottom: 18px;">
        <n-descriptions-item label="触发来源">{{ latestRecord.triggerSource || '-' }}</n-descriptions-item>
        <n-descriptions-item label="模型">{{ latestRecord.modelName || '-' }}</n-descriptions-item>
        <n-descriptions-item label="分析时间">{{ latestRecord.createdAt ? new Date(latestRecord.createdAt).toLocaleString('zh-CN') : '-' }}</n-descriptions-item>
      </n-descriptions>

      <div v-if="parsedSections.length" class="analysis-stack">
        <div v-for="section in parsedSections" :key="section.title" class="analysis-card">
          <div class="analysis-title">{{ section.title }}</div>
          <MdPreview class="analysis-preview" :editor-id="analysisSectionEditorId(section.title)" :model-value="section.content" theme="dark" preview-theme="github" />
        </div>
      </div>
      <div v-else class="empty-text">还没有分析结果。录入资产数据后点击“手动重新分析”即可生成第一版报告。</div>

      <n-collapse style="margin-top: 20px;">
        <n-collapse-item title="本次实际 Prompt" name="prompt">
          <pre class="code-block">{{ latestRecord?.prompt || usedPrompt || '暂无 Prompt' }}</pre>
        </n-collapse-item>
        <n-collapse-item title="输入快照(JSON)" name="payload">
          <pre class="code-block">{{ latestRecord?.inputPayload || latestPayload || '暂无输入快照' }}</pre>
        </n-collapse-item>
      </n-collapse>
    </div>
  </div>
</template>

<script setup>
import { computed, nextTick, onMounted, ref } from 'vue'
import { useMessage } from 'naive-ui'
import { MdPreview } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'

const message = useMessage()

const region = ref('天津市')
const aiConfigs = ref([])
const promptTemplates = ref([])
const selectedAiConfigId = ref(null)
const selectedAnalysisPromptTemplateId = ref(null)
const selectedChatPromptTemplateId = ref(null)
const latestRecord = ref(null)
const latestAnalysis = ref('')
const latestPayload = ref('')
const usedPrompt = ref('')
const profile = ref({ householdName: '我的家庭', region: '天津市', membersCount: 2 })
const benchmarks = ref([])
const runningAnalysis = ref(false)
const statusMessage = ref('')
const statusType = ref('info')
const chatInput = ref('')
const chatLoading = ref(false)
const messageContainerRef = ref(null)
const chatMessages = ref(createInitialMessages())

const aiOptions = computed(() => aiConfigs.value.map((item) => ({
  label: `${item.Name || item.name || '未命名'} / ${item.ModelName || item.modelName || '未设模型'}`,
  value: Number(item.ID || item.id || 0)
})))

const selectedAiLabel = computed(() => aiOptions.value.find((item) => item.value === selectedAiConfigId.value)?.label || '未选择')

const latestStatusDetail = computed(() => {
  if (!latestRecord.value?.createdAt) {
    return '还没有分析记录'
  }
  return `最近分析时间：${new Date(latestRecord.value.createdAt).toLocaleString('zh-CN')}`
})

const latestPayloadObject = computed(() => {
  if (!latestPayload.value) return {}
  try {
    return JSON.parse(latestPayload.value)
  } catch {
    return {}
  }
})

const latestPayloadSummary = computed(() => latestPayloadObject.value?.summary || {})
const latestPayloadMembers = computed(() => latestPayloadObject.value?.members || [])
const latestPayloadMemberProfiles = computed(() => latestPayloadObject.value?.memberProfiles || [])
const latestPayloadAssets = computed(() => latestPayloadObject.value?.assetDetails || latestPayloadObject.value?.topAssetDetails || [])
const latestPayloadIncomes = computed(() => latestPayloadObject.value?.incomeDetails || [])
const latestPayloadLiabilities = computed(() => latestPayloadObject.value?.liabilityDetails || [])
const latestPayloadLiquidTrend = computed(() => latestPayloadObject.value?.liquidAssetTrend || [])
const latestPayloadLiquidDistribution = computed(() => latestPayloadObject.value?.liquidAssetDistribution || [])

const contextDigest = computed(() => {
  const summary = latestPayloadSummary.value
  const memberNames = latestPayloadMembers.value.map((item) => item.name).filter(Boolean)
  return [
    `家庭：${profile.value.householdName || '我的家庭'}`,
    `成员：${memberNames.length ? memberNames.join('、') : `${profile.value.membersCount || 1} 人`}`,
    `总资产：${formatMoney(summary.totalAssets)}`,
    `净资产：${formatMoney(summary.netAssets)}`,
    `总负债：${formatMoney(summary.totalLiabilities)}`,
    `负债率：${Number(summary.debtRatio || 0).toFixed(2)}%`,
    `月税后收入：${formatMoney(summary.monthlyNetIncome)}`,
    `月供：${formatMoney(summary.monthlyDebtPayment)}`,
    `真实月供占用：${formatMoney(summary.monthlyEffectiveDebtPayment)}`,
    `家庭月均支出：${formatMoney(latestPayloadObject.value?.profile?.monthlyHouseholdSpend)}`,
    `月结余：${formatMoney(Number(summary.monthlyNetIncome || 0) - Number(summary.monthlyEffectiveDebtPayment || 0) - Number(latestPayloadObject.value?.profile?.monthlyHouseholdSpend || 0))}`
  ].join(' · ')
})

const memberDigest = computed(() =>
  latestPayloadMemberProfiles.value.map((item) => {
    const protection = []
    if (item?.protectionStatus?.hasSocialInsurance) protection.push('五险')
    if (item?.protectionStatus?.hasHousingFund) protection.push('公积金')
    if (Array.isArray(item?.protectionStatus?.commercialCoverage) && item.protectionStatus.commercialCoverage.length) {
      protection.push(`商保:${item.protectionStatus.commercialCoverage.join('、')}`)
    }
    return [
      item.name,
      item.relationship || '关系未填',
      item.age ? `${item.age}岁` : '年龄未填',
      item.occupation || '职业未填',
      protection.join('/') || '保障未录入'
    ].join(' · ')
  })
)

const assetDigest = computed(() =>
  latestPayloadAssets.value.map((item) => `${item.name} · ${item.category || '资产'} · ${formatMoney(item.value)} · 占比${Number(item.shareOfTotal || 0).toFixed(2)}%`)
)

const incomeDigest = computed(() =>
  latestPayloadIncomes.value.map((item) => {
    const parts = [
      item.owner || item.name,
      item.type || '收入',
      `税前${formatMoney(item.monthlyGross)}`,
    ]
    if (Number(item.pretaxDeduction || 0) > 0) parts.push(`个人五险一金${formatMoney(item.pretaxDeduction)}`)
    if (Number(item.otherPretaxDeduction || 0) > 0) parts.push(`其他税前扣除${formatMoney(item.otherPretaxDeduction)}`)
    if (Number(item.specialDeduction || 0) > 0) parts.push(`专项附加${formatMoney(item.specialDeduction)}`)
    if (Number(item.basicDeduction || 0) > 0) parts.push(`基本减除${formatMoney(item.basicDeduction)}`)
    if (item.type === 'salary') parts.push(`应纳税所得额${formatMoney(item.taxableIncome)}`)
    parts.push(`个税${formatMoney(item.monthlyTax)}`)
    parts.push(`税后${formatMoney(item.monthlyNet)}`)
    if (item.formulaText) parts.push(item.formulaText)
    return parts.join(' · ')
  })
)

const liabilityDigest = computed(() =>
  latestPayloadLiabilities.value.map((item) => `${item.name} · ${formatMoney(item.outstandingPrincipal)} · 月供${formatMoney(item.monthlyPayment)} · ${item.remainingMonths || 0}个月`)
)

const liquidTrendDigest = computed(() =>
  latestPayloadLiquidTrend.value.slice(-6).map((item) => `${item.date} · 流动资金${formatMoney(item.totalLiquidAssets)}`)
)

const liquidDistributionDigest = computed(() =>
  latestPayloadLiquidDistribution.value.map((item) => `${item.name} · ${formatMoney(item.balance)} · 占流动资金${Number(item.shareOfLiquid || 0).toFixed(2)}%`)
)

const parsedSections = computed(() => {
  if (!latestAnalysis.value) return []
  const titles = ['核心结论', '关键指标表', '风险点', '优化建议', '地区/全国对比', '后续关注项']
  return titles.map((title, index) => {
    const nextTitle = titles[index + 1]
    const pattern = nextTitle
      ? new RegExp(`(?:^|\\n)#{0,3}\\s*${title}[：\\n\\r]*([\\s\\S]*?)(?=(?:\\n#{0,3}\\s*${nextTitle}[：\\n\\r]))`)
      : new RegExp(`(?:^|\\n)#{0,3}\\s*${title}[：\\n\\r]*([\\s\\S]*)`)
    const match = latestAnalysis.value.match(pattern)
    return {
      title,
      content: normalizeSectionContent(match?.[1] || '')
    }
  }).filter((item) => item.content)
})

const rankingCards = computed(() => {
  const memberCount = Math.max(1, Number(latestPayloadMembers.value.length || profile.value.membersCount || 1))
  const netAssetsPerAdult = Number(latestPayloadSummary.value.netAssets || 0) / memberCount
  const debtRatio = Number(latestPayloadSummary.value.debtRatio || 0)
  return [
    buildAssetRankingCard('天津资产水平', '天津市', netAssetsPerAdult),
    buildAssetRankingCard('全国资产水平', '全国', netAssetsPerAdult),
    buildDebtRankingCard('天津负债率水平', '天津市', debtRatio),
    buildDebtRankingCard('全国负债率水平', '全国', debtRatio)
  ]
})

const latestAnalysisTimeDisplay = computed(() => {
  const timestamp = getRecordTimestamp(latestRecord.value)
  return timestamp ? formatDateTime(timestamp) : '-'
})

const latestStatusText = computed(() => {
  const timestamp = getRecordTimestamp(latestRecord.value)
  return timestamp ? `最近分析时间：${formatDateTime(timestamp)}` : '还没有分析记录'
})

function createInitialMessages() {
  return [
    {
      role: 'assistant',
      content: '我会直接基于你已经录入的家庭资产、负债、收入、保障、成员和地区基准来回答。你可以继续追问月供压力、税后现金流、保险缺口、季度更新重点等问题。',
      datetime: formatNow()
    }
  ]
}

function getRecordTimestamp(record) {
  if (!record) return ''
  return record.createdAt || record.CreatedAt || record.created_at || ''
}

function formatDateTime(value) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return String(value)
  return date.toLocaleString('zh-CN')
}

function normalizeSectionContent(content) {
  return String(content || '')
    .trim()
    .replace(/^---+\s*/g, '')
    .replace(/\s*---+\s*$/g, '')
    .replace(/^```(?:markdown|md)?\s*/i, '')
    .replace(/\s*```$/i, '')
    .trim()
}

function analysisSectionEditorId(title) {
  return `digital-analysis-${String(title || '')
    .replace(/[^a-zA-Z0-9\u4e00-\u9fa5_-]+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '') || 'section'}`
}

function analysisStatusLabel(status) {
  const normalized = String(status || '').toLowerCase()
  if (!normalized) return '未运行'
  if (normalized === 'success') return '分析完成'
  if (normalized === 'running') return '分析中'
  if (normalized === 'failed') return '分析失败'
  if (normalized === 'skipped') return '未执行'
  return status
}

function analysisStatusTone(status) {
  const normalized = String(status || '').toLowerCase()
  if (normalized === 'success') return 'success'
  if (normalized === 'running') return 'info'
  if (normalized === 'failed') return 'error'
  return 'muted'
}

function findBenchmark(regionName, candidates) {
  return benchmarks.value.find((item) => item.region === regionName && candidates.includes(item.name))
}

function assetLevelByMultiple(multiple) {
  if (multiple <= 0) return '基准不足'
  if (multiple < 1) return '约后 50%'
  if (multiple < 2) return '约前 50%-35%'
  if (multiple < 4) return '约前 35%-20%'
  if (multiple < 8) return '约前 20%-10%'
  if (multiple < 15) return '约前 10%-5%'
  return '约前 5%以内'
}

function debtLevelByRatio(relativeRatio) {
  if (relativeRatio <= 0) return '基准不足'
  if (relativeRatio <= 0.3) return '优于约 90%'
  if (relativeRatio <= 0.6) return '优于约 75%'
  if (relativeRatio <= 1.0) return '优于约 60%'
  if (relativeRatio <= 1.4) return '接近中位'
  if (relativeRatio <= 2.0) return '约后 40%'
  return '约后 20%'
}

function buildAssetRankingCard(label, regionName, netAssetsPerAdult) {
  const incomeBenchmark = findBenchmark(regionName, ['天津居民人均可支配收入', '全国居民人均可支配收入'])
  if (!incomeBenchmark?.value || netAssetsPerAdult <= 0) {
    return { label, value: '基准不足', sub: '缺少收入基准或当前净资产数据，暂时不能估算资产层级。' }
  }
  const multiple = netAssetsPerAdult / Number(incomeBenchmark.value)
  return {
    label,
    value: assetLevelByMultiple(multiple),
    sub: `按人均净资产 ${formatMoney(netAssetsPerAdult)} 对比 ${regionName} 人均收入 ${formatMoney(incomeBenchmark.value)}，约 ${multiple.toFixed(1)} 倍。`
  }
}

function buildDebtRankingCard(label, regionName, debtRatio) {
  const benchmark = findBenchmark(regionName, ['天津负债率参考', '全国负债率参考', '天津贷款/存款余额比', '全国住户贷款/存款余额比'])
  if (benchmark?.value == null) {
    return { label, value: '基准不足', sub: '缺少地区负债率参考，暂时不能估算所处层级。' }
  }
  const relativeRatio = debtRatio / Number(benchmark.value)
  return {
    label,
    value: debtLevelByRatio(relativeRatio),
    sub: `当前负债率 ${Number(debtRatio).toFixed(2)}%，对比 ${regionName} 参考值 ${Number(benchmark.value).toFixed(2)}%。`
  }
}

function formatMoney(value) {
  return `¥ ${Number(value || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

function formatNow() {
  return new Date().toLocaleString('zh-CN', { hour12: false })
}

function pickPromptId(name) {
  const exact = promptTemplates.value.find((item) => item.name === name)
  return exact?.ID || promptTemplates.value[0]?.ID || null
}

async function loadBootstrap() {
  await Promise.all([loadProfile(), loadConfigs(), loadLatest()])
  await loadBenchmarks()
}

async function loadConfigs() {
  if (window.go?.main?.App?.GetAiConfigs) {
    aiConfigs.value = (await window.go.main.App.GetAiConfigs()) || []
  }
  if (window.go?.main?.App?.GetPromptTemplates) {
    promptTemplates.value = (await window.go.main.App.GetPromptTemplates('', '模型系统Prompt')) || []
  }
  if (!selectedAiConfigId.value && aiConfigs.value.length > 0) {
    selectedAiConfigId.value = Number(aiConfigs.value[0].ID || aiConfigs.value[0].id || 0)
  }
  if (!selectedAnalysisPromptTemplateId.value && promptTemplates.value.length > 0) {
    selectedAnalysisPromptTemplateId.value = pickPromptId('家庭资产分析-标准模板')
  }
  if (!selectedChatPromptTemplateId.value && promptTemplates.value.length > 0) {
    selectedChatPromptTemplateId.value = pickPromptId('家庭数字分析-连续对话模板')
  }
}

async function loadProfile() {
  if (!window.go?.main?.App?.GetHouseholdProfile) return
  const data = await window.go.main.App.GetHouseholdProfile()
  profile.value = { ...profile.value, ...(data || {}) }
  region.value = profile.value.region || region.value || '天津市'
}

async function loadLatest() {
  if (!window.go?.main?.App?.GetLatestHouseholdAIAnalysis) return
  latestRecord.value = await window.go.main.App.GetLatestHouseholdAIAnalysis()
  latestAnalysis.value = latestRecord.value?.analysisMarkdown || ''
  latestPayload.value = latestRecord.value?.inputPayload || ''
  usedPrompt.value = latestRecord.value?.prompt || ''
}

async function loadBenchmarks() {
  if (!window.go?.main?.App?.GetHouseholdBenchmarks) return
  benchmarks.value = (await window.go.main.App.GetHouseholdBenchmarks(region.value)) || []
}

async function runAnalysis(triggerSource) {
  if (!window.go?.main?.App?.RunHouseholdAIAnalysis) return
  runningAnalysis.value = true
  statusMessage.value = ''
  try {
    const result = await window.go.main.App.RunHouseholdAIAnalysis(
      region.value,
      selectedAiConfigId.value || 0,
      selectedAnalysisPromptTemplateId.value || 0,
      triggerSource
    )
    statusType.value = result.success ? 'success' : 'warning'
    statusMessage.value = result.message || (result.success ? '分析完成' : '分析失败')
    await loadLatest()
    await loadBenchmarks()
    if (result.success) {
      message.success('数字分析报告已更新')
    } else {
      message.warning(statusMessage.value)
    }
  } catch (error) {
    console.error(error)
    statusType.value = 'error'
    statusMessage.value = error?.message || '分析失败'
    message.error(statusMessage.value)
  } finally {
    runningAnalysis.value = false
  }
}

function scrollChatToBottom() {
  nextTick(() => {
    const el = messageContainerRef.value
    if (el) {
      el.scrollTop = el.scrollHeight
    }
  })
}

async function sendChat() {
  const question = (chatInput.value || '').trim()
  if (!question || chatLoading.value) return

  chatMessages.value.push({
    role: 'user',
    content: question,
    datetime: formatNow()
  })
  chatInput.value = ''
  chatLoading.value = true
  scrollChatToBottom()

  try {
    const result = await window.go.main.App.ChatWithHouseholdDigitalAnalysis(
      region.value,
      selectedAiConfigId.value || 0,
      selectedChatPromptTemplateId.value || 0,
      chatMessages.value.map((item) => ({ role: item.role, content: item.content }))
    )
    if (!result?.success) {
      throw new Error(result?.message || '对话失败')
    }
    chatMessages.value.push({
      role: 'assistant',
      content: result.reply || '模型没有返回内容',
      datetime: formatNow()
    })
    scrollChatToBottom()
  } catch (error) {
    console.error(error)
    message.error(error?.message || '数字分析对话失败')
  } finally {
    chatLoading.value = false
  }
}

function resetChat() {
  chatMessages.value = createInitialMessages()
}

onMounted(loadBootstrap)
</script>

<style scoped>
.digital-analysis-page {
  max-width: 1240px;
  margin: 0 auto;
}

.metric-card,
.mini-card,
.analysis-card {
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: 20px;
  background:
    radial-gradient(circle at top right, rgba(56, 189, 248, 0.1), transparent 38%),
    linear-gradient(180deg, rgba(15, 23, 42, 0.96), rgba(30, 41, 59, 0.86));
}

.metric-card {
  padding: 20px 22px;
  min-height: 136px;
}

.metric-label,
.mini-label {
  font-size: 12px;
  color: var(--text-secondary);
}

.metric-value {
  margin-top: 12px;
  font-size: 22px;
  font-weight: 700;
}

.metric-model {
  margin-top: 12px;
  font-size: 18px;
  line-height: 1.5;
  font-weight: 700;
  color: #f8fafc;
  word-break: break-word;
}

.metric-sub,
.mini-sub,
.section-sub {
  margin-top: 8px;
  color: var(--text-muted);
  line-height: 1.7;
  font-size: 12px;
}

.status-chip {
  display: inline-flex;
  align-items: center;
  padding: 6px 12px;
  border-radius: 999px;
  font-size: 13px;
  font-weight: 700;
}

.status-chip.success {
  background: rgba(16, 185, 129, 0.18);
  color: #d1fae5;
}

.status-chip.info {
  background: rgba(59, 130, 246, 0.18);
  color: #dbeafe;
}

.status-chip.error {
  background: rgba(239, 68, 68, 0.18);
  color: #fee2e2;
}

.status-chip.muted {
  background: rgba(148, 163, 184, 0.12);
  color: #e2e8f0;
}

.section-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 16px;
}

.section-title {
  font-size: 16px;
  font-weight: 700;
}

.mini-card {
  padding: 18px 20px;
  min-height: 126px;
}

.mini-value {
  margin-top: 12px;
  font-size: 20px;
  font-weight: 700;
  color: #f8fafc;
}

.context-banner {
  margin-bottom: 16px;
  padding: 14px 16px;
  border-radius: 16px;
  border: 1px solid rgba(45, 212, 191, 0.18);
  background: rgba(15, 118, 110, 0.12);
}

.context-title {
  font-size: 12px;
  font-weight: 700;
  color: #99f6e4;
}

.context-text {
  margin-top: 8px;
  font-size: 13px;
  line-height: 1.7;
  color: #d6fbf4;
}

.context-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  margin-top: 14px;
}

.context-block {
  border-radius: 14px;
  border: 1px solid rgba(148, 163, 184, 0.14);
  background: rgba(15, 23, 42, 0.28);
  padding: 12px 14px;
}

.context-block-title {
  font-size: 12px;
  font-weight: 700;
  color: #c7f9f1;
}

.context-list {
  display: grid;
  gap: 8px;
  margin-top: 10px;
}

.context-item,
.context-empty {
  font-size: 12px;
  line-height: 1.7;
  color: #d8eeea;
}

.analysis-stack {
  display: grid;
  gap: 16px;
}

.analysis-card {
  padding: 22px 24px;
}

.analysis-title {
  font-size: 15px;
  font-weight: 700;
  margin-bottom: 14px;
}

.empty-text {
  color: var(--text-secondary);
  padding: 18px 0 8px;
}

.code-block {
  white-space: pre-wrap;
  line-height: 1.7;
  margin: 0;
  font-size: 12px;
}

.chat-shell {
  display: grid;
  gap: 16px;
}

.message-list {
  min-height: 360px;
  max-height: 620px;
  overflow-y: auto;
  padding: 18px;
  border-radius: 20px;
  border: 1px solid rgba(148, 163, 184, 0.16);
  background: rgba(8, 12, 22, 0.78);
}

.message-item {
  max-width: 88%;
  margin-bottom: 16px;
  padding: 14px 16px;
  border-radius: 18px;
  border: 1px solid rgba(148, 163, 184, 0.16);
}

.message-item.user {
  margin-left: auto;
  background: rgba(20, 83, 45, 0.35);
  border-color: rgba(74, 222, 128, 0.22);
}

.message-item.assistant {
  background: rgba(15, 23, 42, 0.84);
}

.message-header {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 10px;
  font-size: 12px;
  color: var(--text-muted);
}

.message-content.plain {
  white-space: pre-wrap;
  line-height: 1.8;
}

.chat-input {
  display: grid;
  gap: 12px;
}

.chat-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

:deep(.analysis-preview),
:deep(.message-content) {
  background: transparent;
}

:deep(.analysis-preview .md-editor-preview),
:deep(.message-content .md-editor-preview) {
  background: transparent;
  padding: 12px 16px 14px;
  color: #e2e8f0;
  line-height: 1.85;
  font-size: 14px;
  border-radius: 12px;
}

:deep(.analysis-preview p),
:deep(.message-content p) {
  margin: 0 0 12px;
}

:deep(.analysis-preview ul),
:deep(.analysis-preview ol),
:deep(.message-content ul),
:deep(.message-content ol) {
  margin: 0 0 12px;
  padding-left: 22px;
}

:deep(.analysis-preview li),
:deep(.message-content li) {
  margin-bottom: 8px;
}

:deep(.analysis-preview table),
:deep(.message-content table) {
  width: 100%;
  border-collapse: collapse;
  margin: 12px 0;
}

:deep(.analysis-preview th),
:deep(.analysis-preview td),
:deep(.message-content th),
:deep(.message-content td) {
  border: 1px solid rgba(148, 163, 184, 0.18);
  padding: 10px 12px;
  vertical-align: top;
}

@media (max-width: 980px) {
  .section-row {
    flex-direction: column;
  }

  .message-item {
    max-width: 100%;
  }

  .context-grid {
    grid-template-columns: 1fr;
  }
}
</style>
