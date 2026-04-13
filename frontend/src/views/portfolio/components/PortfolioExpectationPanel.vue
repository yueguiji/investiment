<template>
  <div class="expectation-shell">
    <div class="expectation-tabs-head">
      <div class="panel-title">收益预期与 AI 分析</div>
      <div class="panel-sub">把目标收益测算和 AI 预期建议拆成页签，持仓总览会更清爽一些。</div>
    </div>

    <n-tabs v-model:value="activeTab" type="line" animated class="expectation-tabs">
      <n-tab-pane name="summary" tab="收益预估设定">
        <div class="panel-card expectation-summary-card">
      <div class="panel-head">
        <div>
          <div class="panel-title">收益预期设定</div>
          <div class="panel-sub">按家庭流动资产口径对齐年度目标，再看当前持仓离目标到底差多少。</div>
        </div>
        <n-space>
          <n-button secondary :loading="loadingSummary" @click="loadAll">刷新预估</n-button>
          <n-button type="primary" :loading="savingTarget" @click="saveTarget">保存目标</n-button>
        </n-space>
      </div>

      <div class="setting-grid">
        <div class="setting-card">
          <span class="setting-label">家庭流动资产</span>
          <strong>{{ formatMoney(summary.householdLiquidAssets) }}</strong>
          <em>来自家庭资产里的流动资产台账</em>
        </div>
        <div class="setting-card setting-card-input">
          <span class="setting-label">目标年收益率</span>
          <n-input-number
            v-model:value="draftTargetRate"
            :show-button="false"
            :min="0"
            :max="100"
            :precision="2"
            placeholder="例如 8 / 12"
            style="width: 100%;"
          />
          <em>会写入家庭画像，后续 AI 分析也会直接带入</em>
        </div>
        <div class="setting-card">
          <span class="setting-label">年度目标收益</span>
          <strong>{{ formatMoney(summary.targetAnnualProfit) }}</strong>
          <em>{{ summary.targetDifficultyLabel || '未设定' }} · {{ formatPercent(summary.targetAnnualReturnRate) }}</em>
        </div>
        <div class="setting-card">
          <span class="setting-label">固收建议上限</span>
          <strong>{{ formatPercent(summary.suggestedFixedIncomeMaxRatio) }}</strong>
          <em>约 {{ formatMoney(summary.suggestedFixedIncomeMaxAmount) }}</em>
        </div>
      </div>

      <div class="expectation-metrics">
        <div v-for="card in metricCards" :key="card.label" class="expectation-metric">
          <span class="metric-label">{{ card.label }}</span>
          <strong :class="card.tone">{{ card.value }}</strong>
          <em>{{ card.sub }}</em>
        </div>
      </div>

      <n-alert v-if="warningLines.length" type="warning" :show-icon="false" class="warning-box">
        <div v-for="item in warningLines" :key="item" class="warning-line">{{ item }}</div>
      </n-alert>

      <div class="bucket-grid">
        <div v-for="bucket in summary.buckets || []" :key="bucket.key" class="bucket-card">
          <div class="bucket-head">
            <span>{{ bucket.label }}</span>
            <span>{{ formatPercent(bucket.weight) }}</span>
          </div>
          <div class="bucket-track">
            <div class="bucket-fill" :style="{ width: `${Math.min(Number(bucket.weight || 0), 100)}%` }"></div>
          </div>
          <div class="bucket-meta">
            <span>{{ formatMoney(bucket.value) }}</span>
            <span>{{ formatSignedPercent(bucket.estimatedAnnualReturnRate) }} / {{ formatSignedMoney(bucket.estimatedAnnualProfit) }}</span>
          </div>
        </div>
      </div>

      <div class="driver-grid">
        <div class="driver-panel">
          <div class="driver-title">主要收益贡献</div>
          <div v-if="positiveDrivers.length" class="driver-list">
            <div v-for="item in positiveDrivers" :key="`${item.code}-up`" class="driver-item">
              <div class="driver-name">{{ item.name || item.code }}</div>
              <div class="driver-meta">{{ item.code }} · {{ item.categoryLabel || item.bucketLabel }}</div>
              <div class="driver-stats">{{ formatMoney(item.totalValue) }} · {{ formatSignedPercent(item.estimatedAnnualReturnRate) }} · {{ formatSignedMoney(item.estimatedAnnualProfit) }}</div>
            </div>
          </div>
          <div v-else class="empty-text">当前还没有明显的正向收益贡献项。</div>
        </div>

        <div class="driver-panel">
          <div class="driver-title">主要收益拖累</div>
          <div v-if="negativeDrivers.length" class="driver-list">
            <div v-for="item in negativeDrivers" :key="`${item.code}-down`" class="driver-item">
              <div class="driver-name">{{ item.name || item.code }}</div>
              <div class="driver-meta">{{ item.code }} · {{ item.categoryLabel || item.bucketLabel }}</div>
              <div class="driver-stats">{{ formatMoney(item.totalValue) }} · {{ formatSignedPercent(item.estimatedAnnualReturnRate) }} · {{ formatSignedMoney(item.estimatedAnnualProfit) }}</div>
            </div>
          </div>
          <div v-else class="empty-text">当前没有明显的负向拖累项。</div>
        </div>
      </div>
        </div>
      </n-tab-pane>

      <n-tab-pane name="analysis" tab="AI 预期分析">
        <div class="panel-card expectation-ai-card">
      <div class="panel-head">
        <div>
          <div class="panel-title">AI 预期分析</div>
          <div class="panel-sub">把目标收益、当前持仓和差距一起交给 AI，直接给出比例和金额建议。</div>
        </div>
        <n-button type="primary" :loading="runningAnalysis" @click="runAnalysis">开始分析</n-button>
      </div>

      <div class="ai-toolbar">
        <div class="ai-meta-card">
          <span class="setting-label">最近状态</span>
          <strong>{{ statusLabel(latestRecord?.status) }}</strong>
          <em>{{ latestRecordTimeText }}</em>
        </div>
        <div class="ai-meta-card">
          <span class="setting-label">提示词模板</span>
          <n-select
            v-model:value="selectedPromptTemplateId"
            :options="promptOptions"
            size="small"
            placeholder="使用默认预期分析模板"
          />
          <em>{{ selectedPromptTemplateName }}</em>
        </div>
      </div>

      <n-alert v-if="analysisMessage" :type="analysisMessageType" :show-icon="false" class="warning-box">
        {{ analysisMessage }}
      </n-alert>

      <n-spin :show="runningAnalysis || loadingSummary">
        <div v-if="latestRecord?.analysisMarkdown" class="analysis-preview-wrap">
          <MdPreview :editor-id="editorId" :model-value="latestRecord.analysisMarkdown" theme="dark" preview-theme="github" />
        </div>
        <div v-else class="analysis-empty">
          保存目标收益后，点“开始分析”就会按当前模板生成持仓预期分析。
        </div>
      </n-spin>
        </div>
      </n-tab-pane>
    </n-tabs>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { NAlert, NButton, NInputNumber, NSelect, NSpin, NTabPane, NTabs, useMessage } from 'naive-ui'
import { MdPreview } from 'md-editor-v3'

const REQUEST_TIMEOUT_MS = 150000
const SYSTEM_PROMPT_TYPE = '模型系统Prompt'

const props = defineProps({
  refreshKey: { type: Number, default: 0 }
})

const message = useMessage()
const activeTab = ref('summary')
const loadingSummary = ref(false)
const savingTarget = ref(false)
const runningAnalysis = ref(false)
const profile = ref({ householdName: '', riskPreference: '', targetAnnualReturnRate: 0 })
const summary = ref(createEmptySummary())
const latestRecord = ref(null)
const promptTemplates = ref([])
const selectedPromptTemplateId = ref(null)
const draftTargetRate = ref(0)
const analysisMessage = ref('')
const analysisMessageType = ref('info')

const editorId = computed(() => `portfolio-expectation-${latestRecord.value?.ID || latestRecord.value?.id || 'latest'}`)

const promptOptions = computed(() => promptTemplates.value.map((item) => ({
  label: item.name || item.Name || '未命名模板',
  value: Number(item.ID || item.id || 0)
})))

const selectedPromptTemplateName = computed(() =>
  promptOptions.value.find((item) => item.value === selectedPromptTemplateId.value)?.label || '持仓收益预期分析-标准模板'
)

const latestRecordTimeText = computed(() => {
  const value = latestRecord.value?.createdAt || latestRecord.value?.CreatedAt
  return value ? shortDateTime(value) : '等待首次分析'
})

const warningLines = computed(() => (summary.value.warnings || []).slice(0, 4))

const positiveDrivers = computed(() =>
  (summary.value.topDrivers || []).filter((item) => Number(item.estimatedAnnualProfit || 0) > 0).slice(0, 4)
)

const negativeDrivers = computed(() =>
  (summary.value.bottomDraggers || []).filter((item) => Number(item.estimatedAnnualProfit || 0) < 0).slice(0, 4)
)

const metricCards = computed(() => [
  {
    label: '仓位年度预估',
    value: formatSignedMoney(summary.value.estimatedPortfolioAnnualProfit),
    sub: `${formatPercent(summary.value.estimatedLiquidAnnualReturnRate)} 占流动资产`,
    tone: Number(summary.value.estimatedPortfolioAnnualProfit || 0) >= 0 ? 'profit' : 'loss'
  },
  {
    label: '年度缺口',
    value: formatGapMoney(summary.value.annualGap),
    sub: `达成度 ${formatPercent(summary.value.projectedCompletionRatio)}`,
    tone: Number(summary.value.annualGap || 0) <= 0 ? 'profit' : 'loss'
  },
  {
    label: '当前累计收益',
    value: formatSignedMoney(summary.value.currentTotalProfit),
    sub: `当前持仓累计 ${formatPercent(summary.value.currentTotalProfitRate)}`,
    tone: Number(summary.value.currentTotalProfit || 0) >= 0 ? 'profit' : 'loss'
  },
  {
    label: '年内应达收益',
    value: formatMoney(summary.value.targetProfitToDate),
    sub: `当前差额 ${formatGapMoney(summary.value.toDateGap)}`,
    tone: Number(summary.value.toDateGap || 0) <= 0 ? 'profit' : 'loss'
  },
  {
    label: '闲置流动资产',
    value: formatMoney(summary.value.idleLiquidAssets),
    sub: `${formatPercent(summary.value.investedRatioOfLiquidAssets)} 已投入持仓`,
    tone: ''
  },
  {
    label: '补缺所需收益率',
    value: formatPercent(summary.value.requiredReturnOnIdleLiquidAssets),
    sub: `${formatPercent(summary.value.requiredReturnOnInvestedCapital)} 按全部持仓口径`,
    tone: Number(summary.value.requiredReturnOnIdleLiquidAssets || 0) > 12 ? 'loss' : 'profit'
  }
])

watch(
  () => props.refreshKey,
  async () => {
    await loadAll()
  }
)

onMounted(async () => {
  await loadAll()
})

async function callApp(name, ...args) {
  const fn = window.go?.main?.App?.[name]
  if (!fn) {
    throw new Error(`missing app method: ${name}`)
  }
  return fn(...args)
}

async function loadPromptTemplates() {
  if (!window.go?.main?.App?.GetPromptTemplates) {
    promptTemplates.value = []
    selectedPromptTemplateId.value = null
    return
  }
  const previousSelectedId = Number(selectedPromptTemplateId.value || 0)
  const result = (await callApp('GetPromptTemplates', '', SYSTEM_PROMPT_TYPE)) || []
  promptTemplates.value = result.filter((item) => {
    const name = String(item?.name || item?.Name || '')
    return name.includes('收益预期分析') || name.includes('持仓收益')
  })
  const keepSelected = promptTemplates.value.find((item) => Number(item.ID || item.id || 0) === previousSelectedId)
  if (keepSelected) {
    selectedPromptTemplateId.value = previousSelectedId
    return
  }
  const exact = promptTemplates.value.find((item) => String(item?.name || item?.Name || '') === '持仓收益预期分析-标准模板')
  selectedPromptTemplateId.value = exact ? Number(exact.ID || exact.id || 0) : Number(promptTemplates.value[0]?.ID || promptTemplates.value[0]?.id || 0) || null
}

async function loadAll() {
  loadingSummary.value = true
  analysisMessage.value = ''
  try {
    await loadPromptTemplates()
    const [profileResult, summaryResult, latestResult] = await Promise.all([
      callApp('GetHouseholdProfile'),
      callApp('GetPortfolioExpectationSummary'),
      callApp('GetLatestPortfolioExpectationAIAnalysis')
    ])
    profile.value = { ...profile.value, ...(profileResult || {}) }
    draftTargetRate.value = Number(profile.value.targetAnnualReturnRate || 0)
    summary.value = { ...createEmptySummary(), ...(summaryResult || {}) }
    latestRecord.value = latestResult || null
  } catch (error) {
    console.error(error)
    analysisMessageType.value = 'error'
    analysisMessage.value = '收益预期数据加载失败。'
  } finally {
    loadingSummary.value = false
  }
}

async function saveTarget(silent = false) {
  savingTarget.value = true
  try {
    const saved = await callApp('UpsertHouseholdProfile', {
      ...profile.value,
      targetAnnualReturnRate: Number(draftTargetRate.value || 0)
    })
    profile.value = { ...profile.value, ...(saved || {}) }
    draftTargetRate.value = Number(profile.value.targetAnnualReturnRate || 0)
    const summaryResult = await callApp('GetPortfolioExpectationSummary')
    summary.value = { ...createEmptySummary(), ...(summaryResult || {}) }
    if (!silent) {
      message.success('目标年收益率已保存')
    }
  } catch (error) {
    console.error(error)
    if (!silent) {
      message.error('保存目标年收益率失败')
    }
    throw error
  } finally {
    savingTarget.value = false
  }
}

async function runAnalysis() {
  if (runningAnalysis.value) return
  if (Number(draftTargetRate.value || 0) <= 0) {
    message.warning('先设置目标年收益率，再做预期分析。')
    return
  }

  runningAnalysis.value = true
  analysisMessage.value = ''
  try {
    if (Math.abs(Number(draftTargetRate.value || 0) - Number(profile.value.targetAnnualReturnRate || 0)) > 0.0001) {
      await saveTarget(true)
    }
    const result = await withTimeout(
      callApp('RunPortfolioExpectationAIAnalysis', 0, Number(selectedPromptTemplateId.value || 0), 'portfolio-overview:manual'),
      REQUEST_TIMEOUT_MS,
      'AI 预期分析超时，请稍后重试。'
    )
    if (!result?.success) {
      analysisMessageType.value = 'error'
      analysisMessage.value = result?.message || 'AI 预期分析失败'
      message.error(analysisMessage.value)
      return
    }
    latestRecord.value = result?.record || latestRecord.value
    const summaryResult = await callApp('GetPortfolioExpectationSummary')
    summary.value = { ...createEmptySummary(), ...(summaryResult || {}) }
    analysisMessageType.value = 'success'
    analysisMessage.value = 'AI 预期分析已完成。'
    message.success('AI 预期分析已完成')
  } catch (error) {
    console.error(error)
    analysisMessageType.value = 'error'
    analysisMessage.value = error instanceof Error ? error.message : 'AI 预期分析失败'
    message.error(analysisMessage.value)
  } finally {
    runningAnalysis.value = false
  }
}

function createEmptySummary() {
  return {
    householdLiquidAssets: 0,
    targetAnnualReturnRate: 0,
    targetAnnualProfit: 0,
    targetDifficultyLabel: '未设定',
    targetProfitToDate: 0,
    estimatedPortfolioAnnualProfit: 0,
    estimatedLiquidAnnualReturnRate: 0,
    annualGap: 0,
    projectedCompletionRatio: 0,
    currentTotalProfit: 0,
    currentTotalProfitRate: 0,
    toDateGap: 0,
    idleLiquidAssets: 0,
    investedRatioOfLiquidAssets: 0,
    requiredReturnOnIdleLiquidAssets: 0,
    requiredReturnOnInvestedCapital: 0,
    suggestedFixedIncomeMaxRatio: 0,
    suggestedFixedIncomeMaxAmount: 0,
    buckets: [],
    topDrivers: [],
    bottomDraggers: [],
    warnings: []
  }
}

function withTimeout(promise, timeoutMs, timeoutMessage) {
  return Promise.race([
    promise,
    new Promise((_, reject) => {
      setTimeout(() => reject(new Error(timeoutMessage)), timeoutMs)
    })
  ])
}

function formatMoney(value) {
  return `¥${Number(value || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

function formatSignedMoney(value) {
  const amount = Number(value || 0)
  return `${amount >= 0 ? '+' : '-'}${formatMoney(Math.abs(amount))}`
}

function formatGapMoney(value) {
  const amount = Number(value || 0)
  if (amount > 0) return `还差 ${formatMoney(amount)}`
  if (amount < 0) return `超出 ${formatMoney(Math.abs(amount))}`
  return '已对齐'
}

function formatPercent(value) {
  return `${Number(value || 0).toFixed(2)}%`
}

function formatSignedPercent(value) {
  const amount = Number(value || 0)
  return `${amount >= 0 ? '+' : ''}${amount.toFixed(2)}%`
}

function shortDateTime(value) {
  if (!value) return '-'
  return String(value).replace('T', ' ').slice(0, 16)
}

function statusLabel(status) {
  const normalized = String(status || '').toLowerCase()
  if (!normalized) return '未分析'
  if (normalized === 'success') return '分析完成'
  if (normalized === 'failed') return '分析失败'
  if (normalized === 'skipped') return '未执行'
  return status
}
</script>

<style scoped>
.expectation-shell {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.expectation-tabs-head {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.expectation-tabs {
  margin-top: -2px;
}

.expectation-tabs :deep(.n-tabs-nav) {
  margin-bottom: 14px;
}

.expectation-tabs :deep(.n-tabs-tab) {
  padding-left: 0;
  padding-right: 0;
  margin-right: 24px;
}

.expectation-tabs :deep(.n-tabs-tab__label) {
  font-weight: 600;
}

.panel-card {
  border-radius: var(--radius-lg);
  background: linear-gradient(180deg, rgba(14, 24, 39, 0.96), rgba(18, 30, 48, 0.98));
  border: 1px solid rgba(97, 118, 148, 0.26);
  box-shadow: 0 18px 40px rgba(7, 12, 18, 0.18);
  padding: 18px;
}

.panel-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 16px;
}

.panel-title {
  font-size: 16px;
  font-weight: 700;
}

.panel-sub,
.setting-label,
.analysis-empty,
.driver-meta,
.warning-line,
.setting-card em,
.expectation-metric em {
  color: var(--text-secondary);
}

.panel-sub {
  margin-top: 6px;
  font-size: 12px;
  line-height: 1.6;
}

.setting-grid,
.expectation-metrics,
.bucket-grid {
  display: grid;
  gap: 12px;
}

.setting-grid {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.setting-card,
.expectation-metric,
.bucket-card,
.driver-panel,
.ai-meta-card {
  border-radius: var(--radius-md);
  border: 1px solid rgba(97, 118, 148, 0.18);
  background: rgba(9, 16, 27, 0.34);
}

.setting-card {
  min-height: 112px;
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.setting-card strong,
.expectation-metric strong,
.ai-meta-card strong {
  font-size: 24px;
  font-weight: 700;
  line-height: 1.1;
}

.setting-card em,
.expectation-metric em,
.ai-meta-card em,
.driver-stats {
  font-size: 12px;
  font-style: normal;
  line-height: 1.6;
}

.setting-card-input :deep(.n-input-number) {
  margin-top: auto;
}

.expectation-metrics {
  grid-template-columns: repeat(3, minmax(0, 1fr));
  margin-top: 14px;
}

.expectation-metric {
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.expectation-metric strong.profit,
.driver-stats .profit {
  color: var(--profit);
}

.expectation-metric strong.loss {
  color: var(--loss);
}

.warning-box {
  margin-top: 14px;
}

.warning-line + .warning-line {
  margin-top: 6px;
}

.bucket-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
  margin-top: 14px;
}

.bucket-card {
  padding: 14px;
}

.bucket-head,
.bucket-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.bucket-track {
  margin: 10px 0;
  height: 8px;
  border-radius: 999px;
  background: rgba(124, 146, 177, 0.18);
  overflow: hidden;
}

.bucket-fill {
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, #34d399, #38bdf8);
}

.bucket-meta {
  font-size: 12px;
  color: var(--text-secondary);
}

.driver-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  margin-top: 14px;
}

.driver-panel {
  padding: 14px;
}

.driver-title {
  font-size: 13px;
  font-weight: 700;
  margin-bottom: 10px;
}

.driver-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.driver-item {
  padding: 12px;
  border-radius: var(--radius-sm);
  background: rgba(15, 23, 42, 0.52);
  border: 1px solid rgba(148, 163, 184, 0.12);
}

.driver-name {
  font-weight: 600;
}

.driver-meta {
  margin-top: 4px;
  font-size: 12px;
}

.driver-stats {
  margin-top: 6px;
  color: rgba(226, 232, 240, 0.82);
}

.ai-toolbar {
  display: grid;
  grid-template-columns: 1fr 1.1fr;
  gap: 12px;
  margin-bottom: 14px;
}

.ai-meta-card {
  min-height: 98px;
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.analysis-preview-wrap {
  min-height: 480px;
  padding: 16px;
  border-radius: var(--radius-md);
  border: 1px solid rgba(97, 118, 148, 0.18);
  background: rgba(9, 16, 27, 0.34);
  overflow: auto;
}

.analysis-empty {
  min-height: 240px;
  border-radius: var(--radius-md);
  border: 1px dashed rgba(97, 118, 148, 0.28);
  display: flex;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 20px;
  line-height: 1.8;
}

.empty-text {
  color: var(--text-secondary);
  font-size: 12px;
  line-height: 1.7;
}

@media (max-width: 960px) {
  .setting-grid,
  .expectation-metrics,
  .bucket-grid,
  .driver-grid,
  .ai-toolbar {
    grid-template-columns: 1fr;
  }

  .panel-head {
    flex-direction: column;
  }

  .expectation-tabs :deep(.n-tabs-tab) {
    margin-right: 18px;
  }
}
</style>
