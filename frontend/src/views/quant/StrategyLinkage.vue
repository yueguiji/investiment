<template>
  <div class="fade-in linkage-page">
    <n-page-header
      title="联动推荐"
      subtitle="结合宽基风格、行业热度、资金强弱和 AI 周期判断，给出更稳健的脚本切换建议。"
    >
      <template #extra>
        <n-space>
          <n-button :loading="refreshing" @click="runMonitorCycle(true)">刷新信号</n-button>
          <n-button type="primary" @click="$router.push('/quant/templates')">管理脚本库</n-button>
        </n-space>
      </template>
    </n-page-header>

    <div class="platform-card monitor-card">
      <div class="monitor-head">
        <div>
          <div class="section-title">推荐提醒</div>
          <div class="section-subtitle">
            每天首次打开默认关闭。手动开启后，后台会按周期继续检查，即使离开当前菜单也不会停止。
          </div>
        </div>
        <div class="monitor-controls">
          <n-tag :type="monitorEnabled ? 'success' : 'default'" :bordered="false">
            {{ monitorEnabled ? `已开启，每 ${formatInterval(intervalSeconds)} 检查一次` : '当前已关闭' }}
          </n-tag>
          <n-select v-model:value="intervalSeconds" :options="intervalOptions" style="width: 150px" />
          <n-select
            v-model:value="selectedAiConfigId"
            @update:value="persistSelectedAiConfig"
            :options="aiConfigOptions"
            style="width: 220px"
            placeholder="选择 AI 源"
          />
          <n-switch v-model:value="monitorEnabled" @update:value="handleMonitorToggle" />
        </div>
      </div>

      <n-grid :cols="2" :x-gap="16" :y-gap="12" responsive="screen" item-responsive style="margin-top: 16px">
        <n-gi span="2 m:1">
          <n-alert type="warning" :show-icon="false">
            推荐提醒每天都会自动恢复为关闭，避免打扰工作。需要时再手动开启。
          </n-alert>
        </n-gi>
        <n-gi span="2 m:1">
          <n-alert type="info" :show-icon="false">
            Python 规则推荐与 AI 周期推荐会分别给出理由；只要任一命中，脚本就会被推到前列。
          </n-alert>
        </n-gi>
      </n-grid>
    </div>

    <n-grid :cols="4" :x-gap="16" :y-gap="16" responsive="screen" item-responsive style="margin-top: 20px">
      <n-gi span="4 m:2 l:1">
        <div class="platform-card metric-card">
          <div class="metric-label">市场情绪</div>
          <div class="metric-value">{{ signalSummary.emotion }}</div>
          <div class="metric-desc">{{ signalSummary.emotionReason }}</div>
        </div>
      </n-gi>
      <n-gi span="4 m:2 l:1">
        <div class="platform-card metric-card">
          <div class="metric-label">量能强弱</div>
          <div class="metric-value">{{ signalSummary.volume }}</div>
          <div class="metric-desc">{{ signalSummary.volumeReason }}</div>
        </div>
      </n-gi>
      <n-gi span="4 m:2 l:1">
        <div class="platform-card metric-card">
          <div class="metric-label">宽基风格</div>
          <div class="metric-value small">{{ signalSummary.broadStyle }}</div>
          <div class="metric-desc">{{ signalSummary.broadStyleReason }}</div>
        </div>
      </n-gi>
      <n-gi span="4 m:2 l:1">
        <div class="platform-card metric-card">
          <div class="metric-label">行业焦点</div>
          <div class="metric-value small">{{ signalSummary.topIndustry || '暂无' }}</div>
          <div class="metric-desc">{{ signalSummary.industryReason }}</div>
        </div>
      </n-gi>
    </n-grid>

    <div class="platform-card" style="margin-top: 20px">
      <div class="section-head">
        <div>
          <div class="section-title">重点监控篮子</div>
          <div class="section-subtitle">宽基、红利、成长和行业轮动都会单独打分，直接参与脚本排序。</div>
        </div>
        <div class="section-meta">最近更新：{{ lastSignalAt || '暂无' }}</div>
      </div>

      <div class="bucket-grid">
        <div v-for="item in focusedBuckets" :key="item.id" class="bucket-card">
          <div class="bucket-head">
            <div class="bucket-title">{{ item.label }}</div>
            <n-tag :type="item.active ? 'success' : 'default'" :bordered="false">
              {{ item.active ? '已触发' : '观察中' }}
            </n-tag>
          </div>
          <div class="bucket-score">强度 {{ item.score }}/10</div>
          <div class="bucket-reason">{{ item.reason }}</div>
        </div>
      </div>
    </div>

    <div class="platform-card" style="margin-top: 20px">
      <div class="section-head">
        <div>
          <div class="section-title">双轨推荐</div>
          <div class="section-subtitle">规则推荐和 AI 推荐会一起展示，通知也会分别发出。</div>
        </div>
        <div class="section-meta">{{ aiModelLabel || '未启用 AI' }}</div>
      </div>

      <n-grid :cols="2" :x-gap="16" :y-gap="16" responsive="screen" item-responsive>
        <n-gi span="2 l:1">
          <n-alert :type="topRuleRecommendation ? 'success' : 'default'" title="Python 规则推荐">
            <template v-if="topRuleRecommendation">
              <div class="rec-name">{{ topRuleRecommendation.name }}</div>
              <div class="rec-action">{{ topRuleRecommendation.ruleAction }}</div>
              <div class="rec-reason">{{ topRuleRecommendation.ruleReason }}</div>
              <div class="rec-meta">最后刷新：{{ lastRuleRefreshAt || '暂无' }}</div>
            </template>
            <template v-else>当前暂无规则推荐结果。</template>
          </n-alert>
        </n-gi>
        <n-gi span="2 l:1">
          <n-alert :type="topAiRecommendation ? 'info' : 'default'" title="AI 周期推荐">
            <template v-if="topAiRecommendation">
              <div class="rec-name">{{ topAiRecommendation.name }}</div>
              <div class="rec-action">
                {{ topAiRecommendation.aiAction }}
                <span v-if="topAiRecommendation.aiConfidenceText">{{ topAiRecommendation.aiConfidenceText }}</span>
              </div>
              <div class="rec-reason">{{ topAiRecommendation.aiReason }}</div>
              <div class="rec-meta">最后刷新：{{ lastAiRefreshAt || '暂无' }}</div>
            </template>
            <template v-else>{{ aiStatusMessage || '当前没有可用的 AI 推荐结果。' }}</template>
          </n-alert>
        </n-gi>
      </n-grid>
    </div>

    <div class="platform-card" style="margin-top: 20px">
      <div class="section-head">
        <div>
          <div class="section-title">脚本切换建议</div>
          <div class="section-subtitle">规则命中或 AI 命中的脚本都会排到前列，并分别写明原因。</div>
        </div>
      </div>

      <div v-if="displayRecommendations.length" class="recommendation-list">
        <div v-for="item in displayRecommendations" :key="item.id || item.ID || item.name" class="recommendation-card">
          <div class="recommendation-top">
            <div>
              <div class="recommendation-name">{{ item.name }}</div>
              <div class="recommendation-tags">
                <n-tag
                  v-for="tag in item.visibleTags"
                  :key="`${item.name}-${tag}`"
                  size="small"
                  :bordered="false"
                  type="info"
                >
                  {{ formatTag(tag) }}
                </n-tag>
              </div>
            </div>
            <div class="recommendation-score">{{ item.sortPriority }} 分</div>
          </div>

          <div class="recommendation-block">
            <div class="block-title">规则建议</div>
            <div class="block-action">{{ item.ruleAction }}</div>
            <div class="block-text">{{ item.ruleReason }}</div>
          </div>

          <div class="recommendation-block">
            <div class="block-title">AI 建议</div>
            <div class="block-action">{{ item.aiAction }}</div>
            <div class="block-text">{{ item.aiReason }}</div>
          </div>

          <div class="recommendation-block">
            <div class="block-title">综合判断</div>
            <div class="block-action">{{ item.finalSuggestion }}</div>
            <div class="block-text">切换时请同时参考规则和 AI 原因，再结合仓位执行。</div>
          </div>
        </div>
      </div>
      <n-empty v-else description="暂无可展示的脚本建议" />
    </div>

    <div class="platform-card" style="margin-top: 20px">
      <div class="section-head">
        <div>
          <div class="section-title">推荐历史</div>
          <div class="section-subtitle">保留最近每次周期刷新得到的规则推荐和 AI 推荐，便于回看切换轨迹。</div>
        </div>
        <div class="section-meta">共 {{ recommendationHistory.length }} 条</div>
      </div>

      <div v-if="recommendationHistory.length" class="history-list">
        <div v-for="item in recommendationHistory" :key="item.id" class="history-card">
          <div class="history-head">
            <div class="history-time">{{ item.timestamp }}</div>
            <n-tag size="small" :bordered="false" type="info">{{ item.trigger }}</n-tag>
          </div>
          <div class="history-grid">
            <div>
              <div class="history-label">Python 规则推荐</div>
              <div class="history-name">{{ item.rule?.name || '暂无' }}</div>
              <div class="history-text">{{ item.rule?.action || '无动作' }}</div>
              <div class="history-text">{{ item.rule?.reason || '暂无说明' }}</div>
            </div>
            <div>
              <div class="history-label">AI 周期推荐</div>
              <div class="history-name">{{ item.ai?.name || '暂无' }}</div>
              <div class="history-text">{{ item.ai?.action || '无动作' }}</div>
              <div class="history-text">{{ item.ai?.reason || '暂无说明' }}</div>
            </div>
          </div>
        </div>
      </div>
      <n-empty v-else description="开启周期推荐或手动刷新后，这里会开始累积历史。" />
    </div>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useMessage } from 'naive-ui'
import { collectTemplateTags, getTagLabel } from '../../utils/quant'

const DEFAULT_AI_OPTION = 'default'
const DEFAULT_INTERVAL_SECONDS = 60
const RUNTIME_KEY = '__investmentQuantLinkageMonitor'
const DAILY_DISABLE_KEY = 'investment.quant.linkage.daily-disable'
const AI_CONFIG_STORAGE_KEY = 'investment.quant.linkage.ai-config-id'
const HISTORY_STORAGE_KEY = 'investment.quant.linkage.history'
const MAX_HISTORY_ITEMS = 20

const message = useMessage()

const templates = ref([])
const taxonomy = ref([])
const hotTopics = ref([])
const hotStocks = ref([])
const moneyRank = ref([])
const industryRank = ref([])
const globalIndexes = ref({})
const aiConfigs = ref([])

const refreshing = ref(false)
const aiAnalyzing = ref(false)
const monitorEnabled = ref(false)
const intervalSeconds = ref(DEFAULT_INTERVAL_SECONDS)
const selectedAiConfigId = ref(DEFAULT_AI_OPTION)
const lastSignalAt = ref('')
const lastRuleRefreshAt = ref('')
const lastAiRefreshAt = ref('')
const aiStatusMessage = ref('')
const aiDecision = ref(null)
const aiModelLabel = ref('')
const recommendationHistory = ref([])

let lastRuleNotificationFingerprint = ''
let lastAiNotificationFingerprint = ''

const intervalOptions = [
  { label: '每 30 秒', value: 30 },
  { label: '每 1 分钟', value: 60 },
  { label: '每 3 分钟', value: 180 },
  { label: '每 5 分钟', value: 300 },
  { label: '每 10 分钟', value: 600 }
]

function getMonitorRuntime() {
  if (!window[RUNTIME_KEY]) {
    window[RUNTIME_KEY] = {
      timer: null,
      intervalSeconds: DEFAULT_INTERVAL_SECONDS,
      enabled: false,
      runner: null,
      selectedAiConfigId: DEFAULT_AI_OPTION,
      lastRuleFingerprint: '',
      lastAiFingerprint: '',
      lastRuleRefreshAt: '',
      lastAiRefreshAt: '',
      history: []
    }
  }
  return window[RUNTIME_KEY]
}

const aiConfigOptions = computed(() => {
  const options = [{ label: '默认 AI 源', value: DEFAULT_AI_OPTION }]
  for (const item of aiConfigs.value) {
    const label = [item.name, item.modelName || item.model].filter(Boolean).join(' / ') || `AI-${item.id || item.ID}`
    options.push({ label, value: Number(item.id || item.ID) })
  }
  return options
})

function todayStamp() {
  const now = new Date()
  return `${now.getFullYear()}-${now.getMonth() + 1}-${now.getDate()}`
}

function resetMonitorDailyDefault() {
  const stamp = todayStamp()
  if (localStorage.getItem(DAILY_DISABLE_KEY) !== stamp) {
    localStorage.setItem(DAILY_DISABLE_KEY, stamp)
    monitorEnabled.value = false
    const runtime = getMonitorRuntime()
    runtime.enabled = false
    stopMonitor()
  }
}

function normalizeList(result) {
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

function normalizeTemplateResult(result) {
  if (Array.isArray(result) && Array.isArray(result[0])) {
    return result[0]
  }
  if (Array.isArray(result)) {
    return result
  }
  return normalizeList(result)
}

function parseNumber(value) {
  if (typeof value === 'number') {
    return Number.isFinite(value) ? value : 0
  }
  const match = String(value ?? '').replace(/,/g, '').match(/-?\d+(\.\d+)?/)
  return match ? Number(match[0]) : 0
}

function pickFirstValue(item, keys) {
  for (const key of keys) {
    if (item && item[key] !== undefined && item[key] !== null && String(item[key]).trim() !== '') {
      return item[key]
    }
  }
  return ''
}

function dedupeBy(list, keyFn) {
  const seen = new Set()
  return (list || []).filter((item) => {
    const key = keyFn(item)
    if (!key || seen.has(key)) {
      return false
    }
    seen.add(key)
    return true
  })
}

function flattenIndexCandidates(raw) {
  const items = []
  const visit = (value) => {
    if (!value) return
    if (Array.isArray(value)) {
      value.forEach(visit)
      return
    }
    if (typeof value !== 'object') return
    const name = pickFirstValue(value, ['name', '名称', '简称', 'label'])
    const code = pickFirstValue(value, ['code', 'symbol', '代码'])
    const change = pickFirstValue(value, ['zdf', 'zd', 'change', 'pct_chg', '涨跌幅'])
    if (name || code) {
      items.push({
        name: String(name || code),
        code: String(code || ''),
        changePct: parseNumber(change)
      })
    }
    Object.values(value).forEach((child) => {
      if (child && typeof child === 'object') {
        visit(child)
      }
    })
  }
  visit(raw)
  return dedupeBy(items, (item) => `${item.code}-${item.name}`)
}

function flattenTopicNames(list) {
  return dedupeBy(
    normalizeList(list)
      .map((item) => String(pickFirstValue(item, ['name', 'title', '概念名称', '板块名称'])).trim())
      .filter(Boolean),
    (item) => item
  ).slice(0, 6)
}

function flattenStockNames(list) {
  return dedupeBy(
    normalizeList(list)
      .map((item) => String(pickFirstValue(item, ['name', '股票名称', 'title', 'f14'])).trim())
      .filter(Boolean),
    (item) => item
  ).slice(0, 6)
}

function averageChange(items, keywords) {
  const matched = items.filter((item) => keywords.some((keyword) => item.name.includes(keyword)))
  if (!matched.length) {
    return { average: 0, items: [] }
  }
  const total = matched.reduce((sum, item) => sum + item.changePct, 0)
  return { average: total / matched.length, items: matched }
}

const broadStyleSignal = computed(() => {
  const items = flattenIndexCandidates(globalIndexes.value)
  const largeCap = averageChange(items, ['上证50', '沪深300', '中证A50', '红利'])
  const smallCap = averageChange(items, ['中证500', '中证1000', '国证2000'])
  const growth = averageChange(items, ['创业板', '科创50'])

  const candidates = [
    { key: 'large-cap', label: '大盘价值占优', value: largeCap.average, keywords: ['宽基', 'ETF', '沪深300', '上证50', '红利', '低波'] },
    { key: 'small-cap', label: '中小盘弹性更强', value: smallCap.average, keywords: ['中证500', '中证1000', '小盘', '弹性', '成长', '轮动'] },
    { key: 'growth', label: '成长风格偏强', value: growth.average, keywords: ['创业板', '科创', '成长', '科技', '突破'] }
  ].sort((a, b) => b.value - a.value)

  const leader = candidates[0]
  if (!leader || leader.value === 0) {
    return {
      label: '宽基分化不明显',
      reason: '宽基指数之间暂未形成明确领涨风格，建议降低切换频率。',
      keywords: ['防守', '观察', '均衡'],
      snapshots: items.slice(0, 8),
      largeCapAverage: largeCap.average,
      smallCapAverage: smallCap.average,
      growthAverage: growth.average
    }
  }

  return {
    label: leader.label,
    reason: `${leader.label}，代表性宽基指数平均涨跌幅约 ${leader.value.toFixed(2)}%。`,
    keywords: leader.keywords,
    snapshots: [...largeCap.items, ...smallCap.items, ...growth.items].slice(0, 8),
    largeCapAverage: largeCap.average,
    smallCapAverage: smallCap.average,
    growthAverage: growth.average
  }
})

const topIndustryList = computed(() =>
  normalizeList(industryRank.value)
    .map((item) => ({
      name: String(pickFirstValue(item, ['板块名称', 'name', 'plate_name', 'industry_name', 'f14'])).trim(),
      changePct: parseNumber(pickFirstValue(item, ['涨跌幅', 'zdf', 'f3']))
    }))
    .filter((item) => item.name)
    .slice(0, 5)
)

const signalSummary = computed(() => {
  const topicHeat = flattenTopicNames(hotTopics.value).length
  const stockHeat = flattenStockNames(hotStocks.value).length
  const moneyHeat = normalizeList(moneyRank.value).slice(0, 10).length
  const topIndustry = topIndustryList.value[0]?.name || ''
  const topIndustryChange = topIndustryList.value[0]?.changePct || 0
  const broadStyle = broadStyleSignal.value.label

  let emotion = '中性偏谨慎'
  let emotionReason = '热点、热股和资金强度没有同步抬升，适合先看确认。'
  if (topicHeat >= 4 && stockHeat >= 4 && moneyHeat >= 8) {
    emotion = '风险偏好抬升'
    emotionReason = '热点、热股与资金活跃度同步走强，适合趋势与轮动脚本。'
  } else if (moneyHeat <= 3 && topIndustryChange < 0.5) {
    emotion = '风险偏好回落'
    emotionReason = '资金和板块带动偏弱，更适合防守与低波脚本。'
  }

  const snapshots = broadStyleSignal.value.snapshots || []
  const avgAbsMove = snapshots.length
    ? snapshots.reduce((sum, item) => sum + Math.abs(item.changePct || 0), 0) / snapshots.length
    : 0

  let volume = '量能均衡'
  let volumeReason = '市场尚未形成极端缩量或放量。'
  if (moneyHeat >= 8 || avgAbsMove >= 1.2) {
    volume = '量能放大'
    volumeReason = '资金排名和指数波动都在抬升，适合趋势跟随与突破脚本。'
  } else if (moneyHeat <= 3 && avgAbsMove <= 0.5) {
    volume = '量能偏弱'
    volumeReason = '资金与价格波动都较弱，更适合防守、均衡或低频脚本。'
  }

  return {
    emotion,
    emotionReason,
    volume,
    volumeReason,
    broadStyle,
    broadStyleReason: broadStyleSignal.value.reason,
    topIndustry,
    industryReason: topIndustry ? `当前行业焦点偏向 ${topIndustry}。` : '当前没有明显行业焦点。',
    topThemes: flattenTopicNames(hotTopics.value),
    topStocks: flattenStockNames(hotStocks.value)
  }
})

const focusedBuckets = computed(() => {
  const emotion = signalSummary.value.emotion
  const volume = signalSummary.value.volume
  const largeCapAverage = broadStyleSignal.value.largeCapAverage || 0
  const smallCapAverage = broadStyleSignal.value.smallCapAverage || 0
  const growthAverage = broadStyleSignal.value.growthAverage || 0
  const topIndustryChange = topIndustryList.value[0]?.changePct || 0
  const hotThemeText = signalSummary.value.topThemes.join(' ')

  return [
    {
      id: 'hs300',
      label: '沪深300 / 上证50 / A50',
      scriptKeywords: ['沪深300', '上证50', 'a50', '大盘', '价值', '宽基', 'etf', '蓝筹'],
      active: largeCapAverage >= 0.8 && emotion !== '风险偏好回落',
      score: Math.max(0, Math.min(10, Math.round(largeCapAverage * 4 + (volume === '量能放大' ? 2 : 0)))),
      reason: largeCapAverage >= 0.8 ? `大盘价值风格更稳，宽基大盘指数平均涨跌幅约 ${largeCapAverage.toFixed(2)}%。` : '大盘宽基尚未形成明显优势，先列入观察。'
    },
    {
      id: 'csi500',
      label: '中证500',
      scriptKeywords: ['中证500', '中盘', '轮动', '宽基', 'etf'],
      active: smallCapAverage >= 0.7 && emotion === '风险偏好抬升',
      score: Math.max(0, Math.min(10, Math.round(smallCapAverage * 4 + (emotion === '风险偏好抬升' ? 2 : 0)))),
      reason: smallCapAverage >= 0.7 ? '中盘弹性开始回暖，中证500 更适合轮动型脚本。' : '中证500 弹性尚不充分，先等待确认。'
    },
    {
      id: 'csi1000',
      label: '中证1000 / 小盘弹性',
      scriptKeywords: ['中证1000', '小盘', '弹性', '高波动', '动量', '突破'],
      active: smallCapAverage >= 1.1 && volume === '量能放大' && emotion === '风险偏好抬升',
      score: Math.max(0, Math.min(10, Math.round(smallCapAverage * 4 + (volume === '量能放大' ? 2 : 0) + (emotion === '风险偏好抬升' ? 2 : 0)))),
      reason: smallCapAverage >= 1.1 ? '小盘弹性和量能同时转强，更适合高贝塔或突破型脚本。' : '小盘风格还未形成足够强度，不建议过早切入高波动脚本。'
    },
    {
      id: 'dividend',
      label: '红利 / 低波防守',
      scriptKeywords: ['红利', '低波', '防守', '股息', '价值', '稳健'],
      active: emotion === '风险偏好回落' || largeCapAverage - smallCapAverage >= 0.4,
      score: Math.max(0, Math.min(10, Math.round((emotion === '风险偏好回落' ? 5 : 2) + Math.max(0, largeCapAverage - smallCapAverage) * 4))),
      reason: emotion === '风险偏好回落' ? '市场风险偏好回落，红利与低波更适合作为防守切换方向。' : '大盘相对小盘更抗跌，红利低波具备防御价值。'
    },
    {
      id: 'growth',
      label: '创业板 / 科创50 / 成长科技',
      scriptKeywords: ['创业板', '科创50', '成长', '科技', '芯片', 'ai', '半导体'],
      active: growthAverage >= 0.8 && /AI|算力|芯片|半导体|机器人|科技/i.test(hotThemeText),
      score: Math.max(0, Math.min(10, Math.round(growthAverage * 4 + (/AI|算力|芯片|半导体|机器人|科技/i.test(hotThemeText) ? 3 : 0)))),
      reason: growthAverage >= 0.8 ? '成长指数与科技主题共振，适合科技成长或景气跟随脚本。' : '成长风格缺少指数与主题的双重确认。'
    },
    {
      id: 'industry',
      label: '行业轮动 / 热门行业 ETF',
      scriptKeywords: ['行业', '板块', '轮动', '主题', 'sector', 'etf'],
      active: topIndustryChange >= 1.2 || signalSummary.value.topThemes.length >= 3,
      score: Math.max(0, Math.min(10, Math.round(topIndustryChange * 3 + Math.min(signalSummary.value.topThemes.length, 4)))),
      reason: topIndustryList.value.length > 0 ? `行业前排集中在 ${topIndustryList.value.slice(0, 3).map((item) => item.name).join(' / ')}。` : '行业强度不够集中，暂不建议重仓轮动。'
    }
  ].sort((a, b) => b.score - a.score)
})

function includesAny(text, keywords) {
  const lower = String(text || '').toLowerCase()
  return (keywords || []).some((keyword) => lower.includes(String(keyword).toLowerCase()))
}

function buildKeywordCorpus(template) {
  return [
    template.name,
    template.description,
    template.searchKeywords,
    template.strategyType,
    template.scriptCategory,
    template.brokerPlatform,
    template.language,
    ...collectTemplateTags(template)
  ]
    .join(' ')
    .toLowerCase()
}

const ruleRecommendations = computed(() => {
  const industryKeywords = topIndustryList.value.map((item) => item.name).filter(Boolean)
  const themeKeywords = signalSummary.value.topThemes
  const stockKeywords = signalSummary.value.topStocks
  const broadKeywords = broadStyleSignal.value.keywords
  const activeBuckets = focusedBuckets.value.filter((item) => item.active)
  const emotionKeywords =
    signalSummary.value.emotion === '风险偏好抬升'
      ? ['情绪', '热点', '趋势', '轮动', '强势']
      : signalSummary.value.emotion === '风险偏好回落'
        ? ['防守', '低波', '观望', '回撤', '价值']
        : ['均衡', '中性', '宽基', '对冲']
  const volumeKeywords =
    signalSummary.value.volume === '量能放大'
      ? ['放量', '突破', '趋势', '追踪', '动量']
      : signalSummary.value.volume === '量能偏弱'
        ? ['低频', '防守', '低波', '等待']
        : ['轮动', '均衡', '宽基']

  return templates.value
    .filter((item) => String(item.language || '').toLowerCase().includes('python'))
    .map((item) => {
      const tags = collectTemplateTags(item)
      const corpus = buildKeywordCorpus(item)
      let score = 0
      const reasons = []

      for (const bucket of activeBuckets) {
        if (includesAny(corpus, bucket.scriptKeywords)) {
          score += bucket.score
          reasons.push(`命中重点篮子：${bucket.label}，${bucket.reason}`)
        }
      }
      if (includesAny(corpus, [...industryKeywords, ...themeKeywords])) {
        score += 4
        reasons.push('与当前行业或主题焦点相关。')
      }
      if (includesAny(corpus, stockKeywords)) {
        score += 2
        reasons.push('与热股线索存在关联。')
      }
      if (includesAny(corpus, broadKeywords)) {
        score += 3
        reasons.push(`适配当前宽基风格：${signalSummary.value.broadStyle}。`)
      }
      if (includesAny(corpus, emotionKeywords)) {
        score += 2
        reasons.push(`更贴合当前市场情绪：${signalSummary.value.emotion}。`)
      }
      if (includesAny(corpus, volumeKeywords)) {
        score += 2
        reasons.push(`更贴合当前量能环境：${signalSummary.value.volume}。`)
      }
      if (String(item.status || '').toLowerCase() === 'active') {
        score += 1
      }

      const ruleAction = score >= 14 ? '优先切换' : score >= 9 ? '列入观察' : '暂不切换'

      return {
        ...item,
        id: item.id || item.ID || item.name,
        ruleScore: score,
        ruleAction,
        ruleReason: reasons.join('；') || '标签和关键词信息不足，建议先补充脚本画像。',
        visibleTags: tags.slice(0, 6)
      }
    })
    .sort((a, b) => b.ruleScore - a.ruleScore)
})

const topRuleRecommendation = computed(() => ruleRecommendations.value[0] || null)

function normalizeAiAction(action) {
  const text = String(action || '').trim()
  if (text.includes('优先')) return '优先切换'
  if (text.includes('观察')) return '列入观察'
  if (text.includes('暂不')) return '暂不切换'
  return text || '待确认'
}

function findTemplateByName(name) {
  const clean = String(name || '').trim()
  if (!clean) return null
  return (
    ruleRecommendations.value.find((item) => item.name === clean) ||
    ruleRecommendations.value.find((item) => item.name.includes(clean) || clean.includes(item.name)) ||
    null
  )
}

const topAiRecommendation = computed(() => {
  const decision = aiDecision.value
  if (!decision) return null
  const template = findTemplateByName(decision.scriptName)
  if (!template) {
    return {
      id: `ai-${decision.scriptName}`,
      name: decision.scriptName || 'AI 未匹配到本地脚本',
      aiAction: normalizeAiAction(decision.action),
      aiReason: decision.reason || aiStatusMessage.value || 'AI 未返回明确原因',
      aiRiskHint: decision.riskHint || '',
      aiConfidenceText: Number.isFinite(decision.confidence) ? `置信度 ${decision.confidence}/10` : ''
    }
  }
  return {
    ...template,
    aiAction: normalizeAiAction(decision.action),
    aiReason: decision.reason || 'AI 已结合多维信号给出建议。',
    aiRiskHint: decision.riskHint || '',
    aiConfidenceText: Number.isFinite(decision.confidence) ? `置信度 ${decision.confidence}/10` : ''
  }
})

const displayRecommendations = computed(() =>
  ruleRecommendations.value
    .map((item) => {
      const aiMatched = topAiRecommendation.value && findTemplateByName(topAiRecommendation.value.name)?.name === item.name ? topAiRecommendation.value : null
      let finalSuggestion = item.ruleAction
      if (aiMatched?.aiAction === '优先切换' || item.ruleAction === '优先切换') {
        finalSuggestion = '建议优先切换'
      } else if (aiMatched?.aiAction === '列入观察' || item.ruleAction === '列入观察') {
        finalSuggestion = '建议列入观察'
      }
      const aiPriority = aiMatched ? (aiMatched.aiAction === '优先切换' ? 40 : aiMatched.aiAction === '列入观察' ? 20 : 8) : 0
      const rulePriority = item.ruleAction === '优先切换' ? 20 : item.ruleAction === '列入观察' ? 8 : 0
      return {
        ...item,
        aiAction: aiMatched?.aiAction || '待确认',
        aiReason: aiMatched?.aiReason || 'AI 当前没有给出明确推荐原因',
        finalSuggestion,
        sortPriority: aiPriority + rulePriority + item.ruleScore
      }
    })
    .sort((a, b) => b.sortPriority - a.sortPriority)
    .slice(0, 12)
)

function formatTag(value) {
  return getTagLabel(taxonomy.value, value)
}

function formatInterval(seconds) {
  if (seconds < 60) return `${seconds} 秒`
  return `${Math.round(seconds / 60)} 分钟`
}

function formatNow() {
  const now = new Date()
  const pad = (value) => String(value).padStart(2, '0')
  return `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())} ${pad(now.getHours())}:${pad(now.getMinutes())}:${pad(now.getSeconds())}`
}

function loadHistoryFromStorage() {
  try {
    const parsed = JSON.parse(localStorage.getItem(HISTORY_STORAGE_KEY) || '[]')
    recommendationHistory.value = Array.isArray(parsed) ? parsed : []
  } catch {
    recommendationHistory.value = []
  }
}

function saveHistoryToStorage() {
  localStorage.setItem(HISTORY_STORAGE_KEY, JSON.stringify(recommendationHistory.value))
}

function appendRecommendationHistory(trigger = '自动刷新') {
  const rule = topRuleRecommendation.value
  const ai = topAiRecommendation.value
  if (!rule && !ai) {
    return
  }
  const timestamp = formatNow()
  recommendationHistory.value = [
    {
      id: `${timestamp}-${ruleFingerprint(rule)}-${aiFingerprint(ai)}`,
      timestamp,
      trigger,
      rule: rule
        ? {
            name: rule.name,
            action: rule.ruleAction,
            reason: rule.ruleReason
          }
        : null,
      ai: ai
        ? {
            name: ai.name,
            action: ai.aiAction,
            reason: ai.aiReason
          }
        : null
    },
    ...recommendationHistory.value
  ].slice(0, MAX_HISTORY_ITEMS)
  saveHistoryToStorage()
  const runtime = getMonitorRuntime()
  runtime.history = recommendationHistory.value
}

function buildAiSummary() {
  return [
    `市场情绪: ${signalSummary.value.emotion}。${signalSummary.value.emotionReason}`,
    `量能强弱: ${signalSummary.value.volume}。${signalSummary.value.volumeReason}`,
    `宽基风格: ${signalSummary.value.broadStyle}。${signalSummary.value.broadStyleReason}`,
    `行业焦点: ${signalSummary.value.topIndustry || '暂无'}。${signalSummary.value.industryReason}`,
    `热门主题: ${signalSummary.value.topThemes.join(' / ') || '暂无'}`,
    `热股线索: ${signalSummary.value.topStocks.join(' / ') || '暂无'}`,
    `重点篮子触发: ${focusedBuckets.value.filter((item) => item.active).map((item) => `${item.label}(${item.score}/10)`).join('；') || '暂无明显触发'}`
  ].join('\n')
}

async function refreshAiRecommendation() {
  if (!window.go?.main?.App?.AnalyzeQuantLinkageWithAI) {
    aiDecision.value = null
    aiStatusMessage.value = '当前版本尚未暴露 AI 联动推荐接口。'
    return
  }

  const candidates = ruleRecommendations.value.slice(0, 12)
  if (!candidates.length) {
    aiDecision.value = null
    aiStatusMessage.value = '当前没有可供 AI 评审的 Python 脚本。'
    return
  }

  aiAnalyzing.value = true
  try {
    const result = await window.go.main.App.AnalyzeQuantLinkageWithAI(
      {
        summary: buildAiSummary(),
        templates: candidates
      },
      selectedAiConfigId.value === DEFAULT_AI_OPTION ? 0 : Number(selectedAiConfigId.value)
    )
    aiModelLabel.value = result?.model || ''
    if (!result?.aiEnabled) {
      aiDecision.value = null
      aiStatusMessage.value = result?.message || 'AI 推荐未开启。'
      return
    }
    const parsed = result?.parsed || {}
    aiDecision.value = {
      scriptName: String(parsed.scriptName || ''),
      action: String(parsed.action || ''),
      reason: String(parsed.reason || ''),
      riskHint: String(parsed.riskHint || ''),
      confidence: Number(parsed.confidence || 0)
    }
    aiStatusMessage.value = result?.message || 'AI 推荐已更新。'
    lastAiRefreshAt.value = formatNow()
  } catch (error) {
    console.error('refresh linkage ai recommendation failed', error)
    aiDecision.value = null
    aiStatusMessage.value = `AI 推荐失败：${error?.message || error}`
    lastAiRefreshAt.value = formatNow()
  } finally {
    aiAnalyzing.value = false
  }
}

function ruleFingerprint(item) {
  return item ? `${item.name}|${item.ruleAction}|${item.ruleReason}` : ''
}

function aiFingerprint(item) {
  return item ? `${item.name}|${item.aiAction}|${item.aiReason}` : ''
}

async function maybeNotifyRecommendation() {
  const runtime = getMonitorRuntime()
  const topRule = topRuleRecommendation.value
  const topAi = topAiRecommendation.value
  const currentRuleFingerprint = ruleFingerprint(topRule)
  const currentAiFingerprint = aiFingerprint(topAi)

  if (
    currentRuleFingerprint &&
    currentRuleFingerprint !== lastRuleNotificationFingerprint &&
    window.go?.main?.App?.SendLocalNotification
  ) {
    const ok = await window.go.main.App.SendLocalNotification(
      '联动推荐更新',
      `Python 规则推荐：${topRule.name}，${topRule.ruleAction}。原因：${topRule.ruleReason}`,
      `quant-linkage-rule-${topRule.name}-${topRule.ruleAction}`
    )
    if (ok) {
      lastRuleNotificationFingerprint = currentRuleFingerprint
      runtime.lastRuleFingerprint = currentRuleFingerprint
    }
  }

  if (
    currentAiFingerprint &&
    currentAiFingerprint !== lastAiNotificationFingerprint &&
    window.go?.main?.App?.SendLocalNotification
  ) {
    const ok = await window.go.main.App.SendLocalNotification(
      '联动推荐更新',
      `AI 周期推荐：${topAi.name}，${topAi.aiAction}。原因：${topAi.aiReason}`,
      `quant-linkage-ai-${topAi.name}-${topAi.aiAction}`
    )
    if (ok) {
      lastAiNotificationFingerprint = currentAiFingerprint
      runtime.lastAiFingerprint = currentAiFingerprint
    }
  }
}

async function loadSignals() {
  if (!window.go?.main?.App) return
  const app = window.go.main.App

  if (app.GetQuantTemplates) {
    const result = await app.GetQuantTemplates(0, '', 1, 200)
    templates.value = normalizeTemplateResult(result)
  }
  if (app.GetQuantTagTaxonomy) {
    taxonomy.value = normalizeList(await app.GetQuantTagTaxonomy())
  }
  if (app.GetAiConfigs) {
    aiConfigs.value = normalizeList(await app.GetAiConfigs())
  }
  if (app.HotTopic) {
    hotTopics.value = normalizeList(await app.HotTopic(10))
  }
  if (app.HotStock) {
    hotStocks.value = normalizeList(await app.HotStock('10'))
  }
  if (app.GetMoneyRankSina) {
    moneyRank.value = normalizeList(await app.GetMoneyRankSina('netamount'))
  }
  if (app.GetIndustryRank) {
    industryRank.value = normalizeList(await app.GetIndustryRank('涨跌幅', 10))
  }
  if (app.GlobalStockIndexes) {
    globalIndexes.value = await app.GlobalStockIndexes()
  }
  lastSignalAt.value = formatNow()
}

async function runMonitorCycle(showMessage = false) {
  refreshing.value = true
  try {
    await loadSignals()
    lastRuleRefreshAt.value = formatNow()
    await refreshAiRecommendation()
    await maybeNotifyRecommendation()
    appendRecommendationHistory(showMessage ? '手动刷新' : '周期刷新')
    if (showMessage) {
      message.success('联动推荐已刷新')
    }
  } catch (error) {
    console.error('run linkage cycle failed', error)
    if (showMessage) {
      message.error(`联动推荐刷新失败：${error?.message || error}`)
    }
  } finally {
    refreshing.value = false
  }
}

function stopMonitor() {
  const runtime = getMonitorRuntime()
  if (runtime.timer) {
    clearInterval(runtime.timer)
    runtime.timer = null
  }
}

function startMonitor() {
  const runtime = getMonitorRuntime()
  runtime.enabled = monitorEnabled.value
  runtime.intervalSeconds = intervalSeconds.value
  runtime.runner = runMonitorCycle

  stopMonitor()
  if (!monitorEnabled.value) return

  runtime.timer = window.setInterval(() => {
    const activeRuntime = getMonitorRuntime()
    if (activeRuntime.enabled && typeof activeRuntime.runner === 'function') {
      activeRuntime.runner(false)
    }
  }, intervalSeconds.value * 1000)
}

function handleMonitorToggle(value) {
  monitorEnabled.value = value
  const runtime = getMonitorRuntime()
  runtime.enabled = value
  if (value) {
    startMonitor()
    runMonitorCycle(false)
  } else {
    stopMonitor()
  }
}

function persistSelectedAiConfig(value) {
  const normalized = value ?? DEFAULT_AI_OPTION
  selectedAiConfigId.value = normalized
  localStorage.setItem(AI_CONFIG_STORAGE_KEY, String(normalized))
  getMonitorRuntime().selectedAiConfigId = normalized
}

function restoreSelectedAiConfig() {
  const runtime = getMonitorRuntime()
  const stored = localStorage.getItem(AI_CONFIG_STORAGE_KEY)
  const preferred = runtime.selectedAiConfigId ?? stored ?? DEFAULT_AI_OPTION
  selectedAiConfigId.value = preferred
}

watch(aiConfigs, (configs) => {
  if (!configs.length) {
    persistSelectedAiConfig(DEFAULT_AI_OPTION)
    return
  }
  const current = String(selectedAiConfigId.value ?? DEFAULT_AI_OPTION)
  if (current === DEFAULT_AI_OPTION) {
    return
  }
  const matched = configs.some((item) => String(item.id || item.ID) === current)
  if (!matched) {
    persistSelectedAiConfig(DEFAULT_AI_OPTION)
  }
})

onMounted(async () => {
  const runtime = getMonitorRuntime()
  intervalSeconds.value = runtime.intervalSeconds || DEFAULT_INTERVAL_SECONDS
  monitorEnabled.value = Boolean(runtime.enabled)
  runtime.selectedAiConfigId = runtime.selectedAiConfigId ?? localStorage.getItem(AI_CONFIG_STORAGE_KEY) ?? DEFAULT_AI_OPTION
  lastRuleNotificationFingerprint = runtime.lastRuleFingerprint || ''
  lastAiNotificationFingerprint = runtime.lastAiFingerprint || ''
  lastRuleRefreshAt.value = runtime.lastRuleRefreshAt || ''
  lastAiRefreshAt.value = runtime.lastAiRefreshAt || ''
  recommendationHistory.value = Array.isArray(runtime.history) ? runtime.history : []

  restoreSelectedAiConfig()
  if (!recommendationHistory.value.length) {
    loadHistoryFromStorage()
  }
  resetMonitorDailyDefault()
  await runMonitorCycle(false)
  startMonitor()
})

onBeforeUnmount(() => {
  const runtime = getMonitorRuntime()
  runtime.enabled = monitorEnabled.value
  runtime.intervalSeconds = intervalSeconds.value
  runtime.runner = runMonitorCycle
  runtime.selectedAiConfigId = selectedAiConfigId.value
  runtime.lastRuleFingerprint = lastRuleNotificationFingerprint
  runtime.lastAiFingerprint = lastAiNotificationFingerprint
  runtime.lastRuleRefreshAt = lastRuleRefreshAt.value
  runtime.lastAiRefreshAt = lastAiRefreshAt.value
  runtime.history = recommendationHistory.value

  if (!monitorEnabled.value) {
    stopMonitor()
  }
})
</script>

<style scoped>
.linkage-page {
  padding-bottom: 24px;
}

.monitor-card,
.metric-card,
.bucket-card,
.recommendation-card {
  border: 1px solid rgba(148, 163, 184, 0.12);
}

.monitor-head,
.section-head,
.recommendation-top,
.bucket-head {
  display: flex;
  justify-content: space-between;
  gap: 16px;
}

.monitor-head,
.section-head {
  align-items: flex-start;
}

.monitor-controls {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.section-title {
  font-size: 18px;
  font-weight: 700;
  color: #f8fafc;
}

.section-subtitle,
.metric-desc,
.bucket-reason,
.block-text,
.rec-reason,
.rec-meta,
.history-text {
  color: #94a3b8;
  line-height: 1.7;
}

.section-meta {
  color: #94a3b8;
  font-size: 13px;
}

.metric-card {
  min-height: 146px;
}

.metric-label {
  color: #94a3b8;
  font-size: 13px;
  margin-bottom: 10px;
}

.metric-value {
  font-size: 34px;
  font-weight: 700;
  color: #f8fafc;
  margin-bottom: 10px;
}

.metric-value.small {
  font-size: 24px;
}

.bucket-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 16px;
  margin-top: 16px;
}

.bucket-card,
.recommendation-card {
  background: rgba(15, 23, 42, 0.3);
  border-radius: 18px;
  padding: 18px;
}

.bucket-title,
.recommendation-name,
.rec-name {
  color: #f8fafc;
  font-size: 18px;
  font-weight: 700;
}

.bucket-score,
.recommendation-score,
.block-title {
  color: #67e8f9;
  font-size: 13px;
  margin-top: 10px;
}

.recommendation-list {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 16px;
  margin-top: 16px;
}

.recommendation-top {
  align-items: flex-start;
}

.recommendation-tags {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-top: 10px;
}

.recommendation-score {
  min-width: 64px;
  text-align: right;
}

.recommendation-block {
  margin-top: 14px;
  padding-top: 14px;
  border-top: 1px solid rgba(148, 163, 184, 0.12);
}

.block-action,
.rec-action {
  margin-top: 6px;
  color: #f8fafc;
  font-weight: 600;
}

.rec-meta {
  margin-top: 10px;
  font-size: 12px;
}

.history-list {
  display: grid;
  gap: 14px;
  margin-top: 16px;
}

.history-card {
  background: rgba(15, 23, 42, 0.3);
  border: 1px solid rgba(148, 163, 184, 0.12);
  border-radius: 16px;
  padding: 16px 18px;
}

.history-head,
.history-grid {
  display: flex;
  justify-content: space-between;
  gap: 16px;
}

.history-grid > div {
  flex: 1;
}

.history-time,
.history-name {
  color: #f8fafc;
  font-weight: 600;
}

.history-label {
  margin-top: 12px;
  margin-bottom: 6px;
  color: #67e8f9;
  font-size: 12px;
}

.rec-action span {
  color: #67e8f9;
  margin-left: 8px;
  font-size: 12px;
}

@media (max-width: 900px) {
  .monitor-head,
  .section-head,
  .recommendation-top,
  .bucket-head,
  .history-head,
  .history-grid {
    flex-direction: column;
  }

  .monitor-controls {
    width: 100%;
  }
}
</style>
