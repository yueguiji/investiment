<template>
  <n-drawer :show="show" :width="820" placement="right" @update:show="handleShowUpdate">
    <n-drawer-content :title="mergedFund ? `${mergedFund.stockName || mergedFund.stockCode}` : '基金详情'" closable>
      <div v-if="mergedFund" class="shell">
        <n-spin :show="loading">
          <section class="card hero">
            <div>
              <div class="title-row">
                <div class="title">{{ mergedFund.stockName || mergedFund.stockCode }}</div>
                <n-tag size="small" round :bordered="false" :type="tagType">{{ displayCategory }}</n-tag>
              </div>
              <div class="sub">{{ mergedFund.stockCode }}<span v-if="mergedFund.fundCompany"> · {{ mergedFund.fundCompany }}</span></div>
              <div class="status-row">
                <span class="pill" :class="mergedFund.estimateUpdated ? 'live' : 'rest'">{{ mergedFund.estimateStatus || '暂无盘中估值' }}</span>
                <span class="muted">{{ latestNavNote }}</span>
              </div>
            </div>
            <n-space>
              <n-button @click="handleRefreshProfile">刷新详情</n-button>
              <n-button type="primary" @click="openFundPage">基金主页</n-button>
            </n-space>
          </section>

          <section class="tab-nav" aria-label="基金详情页签">
            <button
              v-for="tab in tabs"
              :key="tab.value"
              type="button"
              class="tab-btn"
              :class="{ active: activeTab === tab.value }"
              @click="activeTab = tab.value"
            >
              {{ tab.label }}
            </button>
          </section>

          <template v-if="activeTab === 'overview'">
            <section class="stat-grid">
              <div class="card stat">
                <div class="label">今日估算涨跌幅</div>
                <div class="value" :class="numberTone(estimateLatestRate)">{{ formatPercent(estimateLatestRate) }}</div>
                <div class="muted">{{ estimateUpdatedAt || mergedFund.netEstimatedTime || '-' }}</div>
              </div>
              <div class="card stat">
                <div class="label">最近1日</div>
                <div class="value" :class="numberTone(latestReturnValue)">{{ formatPercent(latestReturnValue) }}</div>
                <div class="muted">{{ mergedFund.latestDailyUpdatedAt || mergedFund.netUnitValueDate || '-' }}</div>
              </div>
              <div class="card stat">
                <div class="label">最大回撤</div>
                <div class="value loss-text">{{ formatPercent(repairInfo.maxDrawdown) }}</div>
                <div class="muted">按近6个月走势测算</div>
              </div>
              <div class="card stat">
                <div class="label">回撤修复</div>
                <div class="value" :class="repairInfo.progress >= 100 ? 'profit-text' : ''">{{ repairInfo.progressText }}</div>
                <div class="muted">{{ repairInfo.status }}</div>
              </div>
            </section>

            <section class="card block">
              <div class="head">
                <div>
                  <div class="section-title">今日估值走势</div>
                  <div class="muted">主看今日涨跌率曲线，Tooltip 里保留估算净值和时间。</div>
                </div>
                <div class="muted">更新至 {{ estimateUpdatedAt || mergedFund.netEstimatedTime || '-' }}</div>
              </div>
              <div class="mini-grid">
                <div class="mini-card"><span class="muted">最新</span><strong :class="numberTone(estimateStats.latestRate)">{{ formatPercent(estimateStats.latestRate) }}</strong></div>
                <div class="mini-card"><span class="muted">日内高点</span><strong :class="numberTone(estimateStats.highRate)">{{ formatPercent(estimateStats.highRate) }}</strong></div>
                <div class="mini-card"><span class="muted">日内低点</span><strong :class="numberTone(estimateStats.lowRate)">{{ formatPercent(estimateStats.lowRate) }}</strong></div>
                <div class="mini-card"><span class="muted">已采样</span><strong>{{ estimateStats.pointCount }} 点</strong></div>
              </div>
              <div v-if="estimateHasCurve" ref="estimateChartRef" class="chart"></div>
              <div v-else class="empty">
                <div class="empty-title">今天还没有形成估值曲线</div>
                <div class="muted">{{ estimateEmptyText }}</div>
              </div>
            </section>
          </template>

          <template v-else-if="activeTab === 'analysis'">
            <section class="card block">
              <div class="head">
                <div>
                  <div class="section-title">同类排行</div>
                  <div class="muted">看这只基金在同类里的位置，按最近 1 周、1 月、3 月、6 月展示。</div>
                </div>
                <div class="muted">更新至 {{ profile?.stageRankingsUpdatedAt || mergedFund.netUnitValueDate || '-' }}</div>
              </div>
              <div v-if="preferredRankings.length" class="ranking-grid">
                <article v-for="item in preferredRankings" :key="item.period" class="mini-card rank-card">
                  <div class="rank-top"><span class="period">{{ item.period }}</span><span class="quartile">{{ item.quartile || '暂无分位' }}</span></div>
                  <div class="rank-pos">{{ formatRank(item) }}</div>
                  <div class="rank-meta muted"><span>击败 {{ formatPercentNoSign(item.rankPercentile) }}</span><span>{{ betterThanText(item) }}</span></div>
                  <div class="track"><div class="track-fill" :style="{ width: `${rankBeatPercent(item)}%` }"></div></div>
                  <div class="mini-metrics">
                    <div><span class="muted">本基金</span><strong :class="numberTone(item.returnRate)">{{ formatPercent(item.returnRate) }}</strong></div>
                    <div><span class="muted">同类平均</span><strong :class="numberTone(item.similarAverageRate)">{{ formatPercent(item.similarAverageRate) }}</strong></div>
                    <div><span class="muted">{{ item.benchmarkLabel || '基准' }}</span><strong :class="numberTone(item.benchmarkRate)">{{ formatPercent(item.benchmarkRate) }}</strong></div>
                    <div><span class="muted">排名变化</span><strong :class="rankDeltaTone(item)">{{ formatRankDelta(item) }}</strong></div>
                  </div>
                </article>
              </div>
              <div v-else class="empty muted">暂无同类排名数据</div>
            </section>

            <section class="card block">
              <div class="head">
                <div>
                  <div class="section-title">业绩走势</div>
                  <div class="muted">按区间起点归一化后的涨跌幅走势，单位净值只在 Tooltip 里显示。</div>
                </div>
                <div class="switches">
                  <button v-for="item in rangeOptions" :key="item.value" type="button" class="pill-btn" :class="{ active: rangeKey === item.value }" @click="rangeKey = item.value">{{ item.label }}</button>
                </div>
              </div>
              <div v-if="performanceTrend.length" ref="performanceChartRef" class="chart"></div>
              <div v-else class="empty muted">暂无走势数据</div>
              <div class="mini-grid">
                <div class="mini-card"><span class="muted">前高</span><strong>{{ formatPrice(repairInfo.peakValue, 4) }}</strong></div>
                <div class="mini-card"><span class="muted">回撤低点</span><strong>{{ formatPrice(repairInfo.valleyValue, 4) }}</strong></div>
                <div class="mini-card"><span class="muted">当前净值</span><strong>{{ formatPrice(repairInfo.currentValue, 4) }}</strong></div>
                <div class="mini-card"><span class="muted">走势更新时间</span><strong>{{ profile?.trendUpdatedAt || mergedFund.netUnitValueDate || '-' }}</strong></div>
              </div>
            </section>
          </template>

          <template v-else>
            <section class="card block">
              <div class="head">
                <div>
                  <div class="section-title">基金资料</div>
                  <div class="muted">管理人、基金经理、规模、评级和最近估值时间都放在这里。</div>
                </div>
              </div>
              <div class="mini-grid">
                <div class="mini-card"><span class="muted">基金类型</span><strong>{{ mergedFund.fundType || displayCategory }}</strong></div>
                <div class="mini-card"><span class="muted">基金经理</span><strong>{{ mergedFund.fundManager || '-' }}</strong></div>
                <div class="mini-card"><span class="muted">基金公司</span><strong>{{ mergedFund.fundCompany || '-' }}</strong></div>
                <div class="mini-card"><span class="muted">基金规模</span><strong>{{ mergedFund.fundScale || '-' }}</strong></div>
                <div class="mini-card"><span class="muted">基金评级</span><strong>{{ mergedFund.fundRating || '-' }}</strong></div>
                <div class="mini-card"><span class="muted">最近估值时间</span><strong>{{ estimateUpdatedAt || mergedFund.netEstimatedTime || mergedFund.netUnitValueDate || '-' }}</strong></div>
              </div>
            </section>
          </template>
        </n-spin>
      </div>
    </n-drawer-content>
  </n-drawer>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { useMessage } from 'naive-ui'
import * as echarts from 'echarts'

const props = defineProps({ show: { type: Boolean, default: false }, fund: { type: Object, default: null } })
const emit = defineEmits(['update:show', 'refreshed'])
const message = useMessage()
const loading = ref(false)
const profile = ref(null)
const estimateChartRef = ref(null)
const performanceChartRef = ref(null)
const rangeKey = ref('3m')
const activeTab = ref('overview')
let estimateChart = null
let performanceChart = null

const tabs = [
  { label: '概览', value: 'overview' },
  { label: '排行与走势', value: 'analysis' },
  { label: '基金详情', value: 'profile' }
]

const rangeOptions = [
  { label: '近1月', value: '1m' }, { label: '近3月', value: '3m' }, { label: '近6月', value: '6m' },
  { label: '近1年', value: '1y' }, { label: '今年来', value: 'ytd' }, { label: '成立来', value: 'all' }
]

const mergedFund = computed(() => profile.value || props.fund || null)
const tagType = computed(() => ({ bond: 'success', cash: 'warning', equity: 'info' }[mergedFund.value?.category] || 'default'))
const displayCategory = computed(() => mergedFund.value?.categoryLabel || inferCategoryLabel(mergedFund.value?.fundType, mergedFund.value?.stockName))
const latestNavNote = computed(() => mergedFund.value?.netUnitValueDate ? `最近净值更新 ${mergedFund.value.netUnitValueDate}` : '净值更新时间待同步')
const latestReturnValue = computed(() => num(profile.value?.latestReturn))
const estimateUpdatedAt = computed(() => profile.value?.estimateTrendUpdatedAt || '')
const rawEstimateTrend = computed(() => Array.isArray(profile.value?.estimateTrend) ? profile.value.estimateTrend : [])
const estimateTrend = computed(() => normalizeEstimateTrend(rawEstimateTrend.value, mergedFund.value?.netUnitValue))
const estimateLatestRate = computed(() => num(profile.value?.estimateLatestRate) ?? num(mergedFund.value?.netEstimatedRate) ?? estimateTrend.value.at(-1)?.rate ?? null)
const estimateHasCurve = computed(() => estimateTrend.value.length >= 2)
const estimateStats = computed(() => buildEstimateStats(estimateTrend.value))
const estimateEmptyText = computed(() => estimateTrend.value.length === 1 ? '今天目前只采样到 1 个估值点，继续开着详情页或手动刷新，系统会自动补成曲线。' : (mergedFund.value?.netEstimatedTime ? `今天还没有新的盘中估值序列，最近一次估值时间是 ${mergedFund.value.netEstimatedTime}。` : '当前基金今天还没有盘中估值数据。'))
const preferredRankings = computed(() => {
  const order = ['近1周', '近1月', '近3月', '近6月']
  const map = new Map((profile.value?.stageRankings || []).map((i) => [i.period, i]))
  return order.map((k) => map.get(k)).filter(Boolean)
})
const stageMetrics = computed(() => {
  const f = mergedFund.value || {}
  return [
    { key: '1m', label: '近1月', value: f.netGrowth1 }, { key: '3m', label: '近3月', value: f.netGrowth3 },
    { key: '6m', label: '近6月', value: f.netGrowth6 }, { key: '1y', label: '近1年', value: f.netGrowth12 },
    { key: '3y', label: '近3年', value: f.netGrowth36 }, { key: 'ytd', label: '今年来', value: f.netGrowthYTD }
  ]
})
const activeTrend = computed(() => filterTrend(profile.value?.trend || [], rangeKey.value))
const performanceTrend = computed(() => normalizePerformanceTrend(activeTrend.value))
const repairTrend = computed(() => buildRepairTrend(activeTrend.value))
const repairInfo = computed(() => buildRepairInfo(filterTrend(profile.value?.trend || [], '6m')))

watch(() => [props.show, props.fund?.stockCode], async ([show, code], oldValue) => {
  const [prevShow, prevCode] = Array.isArray(oldValue) ? oldValue : [false, '']
  if (show && code && (code !== prevCode || (!prevShow && show) || !profile.value)) {
    activeTab.value = 'overview'
    await loadProfile({ silent: true })
  }
}, { immediate: true })

watch([estimateTrend, performanceTrend], async () => {
  if (!props.show) return
  await nextTick()
  renderCharts()
}, { deep: true })

watch(activeTab, async () => {
  if (!props.show) return
  await nextTick()
  renderCharts()
})

watch(() => props.show, async (show) => {
  if (show) {
    await nextTick()
    renderCharts()
    window.addEventListener('resize', resizeCharts)
  } else {
    window.removeEventListener('resize', resizeCharts)
    disposeCharts()
  }
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', resizeCharts)
  disposeCharts()
})

async function loadProfile(options = {}) {
  const { silent = false, refresh = false } = options
  const code = props.fund?.stockCode
  if (!code || !window.go?.main?.App?.GetFundProfile) return
  if (!silent) loading.value = true
  try {
    const loader = refresh && window.go?.main?.App?.RefreshFundProfile
      ? window.go.main.App.RefreshFundProfile(code)
      : window.go.main.App.GetFundProfile(code)
    profile.value = (await withTimeout(loader, refresh ? 20000 : 5000)) || props.fund
    if (refresh) emit('refreshed')
    await nextTick()
    renderCharts()
  } catch (error) {
    console.error(error)
    if (!silent) message.error(refresh ? '刷新详情超时或失败，请稍后再试' : '基金详情加载失败')
  } finally {
    if (!silent) loading.value = false
  }
}

async function handleRefreshProfile() {
  await loadProfile({ refresh: true })
}

function handleShowUpdate(value) { emit('update:show', value) }
function openFundPage() {
  const code = mergedFund.value?.stockCode
  if (code && window.go?.main?.App?.OpenURL) window.go.main.App.OpenURL(`https://fund.eastmoney.com/${code}.html`)
}

function renderCharts() { renderEstimateChart(); renderPerformanceChart() }
function resizeCharts() { estimateChart?.resize(); performanceChart?.resize() }
function disposeCharts() {
  if (estimateChart) { estimateChart.dispose(); estimateChart = null }
  if (performanceChart) { performanceChart.dispose(); performanceChart = null }
}

function renderEstimateChart() {
  if (activeTab.value !== 'overview' || !estimateChartRef.value || !estimateHasCurve.value) {
    if (estimateChart) { estimateChart.dispose(); estimateChart = null }
    return
  }
  if (!estimateChart) estimateChart = echarts.init(estimateChartRef.value)
  estimateChart.setOption(buildLineChartOption({
    points: estimateTrend.value,
    x: (p) => p.timeLabel,
    y: (p) => p.rate,
    color: '#5eead4',
    tooltip: (p) => [p.timeLabel, `今日涨跌率 ${formatPercent(p.rate)}`, `估算净值 ${formatPrice(p.estimatedUnit, 4)}`]
  }))
}

function renderPerformanceChart() {
  if (activeTab.value !== 'analysis' || !performanceChartRef.value || !performanceTrend.value.length) {
    if (performanceChart) { performanceChart.dispose(); performanceChart = null }
    return
  }
  if (!performanceChart) performanceChart = echarts.init(performanceChartRef.value)
  const points = performanceTrend.value
  const repairs = repairTrend.value
  performanceChart.setOption({
    backgroundColor: 'transparent',
    animation: false,
    legend: {
      top: 0,
      right: 12,
      itemWidth: 18,
      itemHeight: 8,
      textStyle: { color: '#91a3b8', fontSize: 11 },
      data: ['业绩走势', '回撤修复']
    },
    grid: { left: 18, right: 18, top: 42, bottom: 26, containLabel: true },
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(10,16,26,0.94)',
      borderColor: 'rgba(148,163,184,0.18)',
      textStyle: { color: '#e5eef9' },
      formatter: (params) => {
        const index = params?.[0]?.dataIndex ?? 0
        const point = points[index]
        const repairPoint = repairs[index]
        return [
          point?.date || '-',
          '业绩走势 ' + formatPercent(point?.percent),
          '回撤修复 ' + formatPercentNoSign(repairPoint?.progress),
          '单位净值 ' + formatPrice(point?.value, 4)
        ].join('<br/>')
      }
    },
    xAxis: {
      type: 'category',
      data: points.map((p) => p.date.slice(5)),
      boundaryGap: false,
      axisLine: { lineStyle: { color: 'rgba(148,163,184,0.18)' } },
      axisLabel: { color: '#91a3b8', fontSize: 11 }
    },
    yAxis: [
      {
        type: 'value',
        scale: true,
        axisLine: { show: false },
        splitLine: { lineStyle: { color: 'rgba(148,163,184,0.12)' } },
        axisLabel: { color: '#91a3b8', formatter: (value) =>           Number(value || 0).toFixed(2) + '%' }
      },
      {
        type: 'value',
        min: 0,
        max: 100,
        splitLine: { show: false },
        axisLabel: { color: '#fbbf24', formatter: (value) =>           Number(value || 0).toFixed(0) + '%' }
      }
    ],
    series: [
      {
        name: '业绩走势',
        type: 'line',
        data: points.map((p) => p.percent),
        smooth: true,
        symbol: 'none',
        lineStyle: { width: 3, color: '#3ecf8e' },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: '#3ecf8e55' },
            { offset: 1, color: '#3ecf8e05' }
          ])
        },
        markLine: {
          symbol: 'none',
          lineStyle: { color: 'rgba(148,163,184,0.24)', type: 'dashed' },
          label: { color: '#91a3b8' },
          data: [{ yAxis: 0 }]
        }
      },
      {
        name: '回撤修复',
        type: 'line',
        yAxisIndex: 1,
        data: repairs.map((p) => p.progress),
        smooth: true,
        symbol: 'none',
        lineStyle: { width: 2, color: '#f59e0b', type: 'dashed' }
      }
    ]
  })
}

function buildLineChartOption({ points, x, y, color, tooltip }) {
  return {
    backgroundColor: 'transparent',
    animation: false,
    grid: { left: 18, right: 18, top: 18, bottom: 26, containLabel: true },
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(10,16,26,0.94)',
      borderColor: 'rgba(148,163,184,0.18)',
      textStyle: { color: '#e5eef9' },
      formatter: (params) => (tooltip(points[params?.[0]?.dataIndex ?? 0]) || []).join('<br/>')
    },
    xAxis: {
      type: 'category', data: points.map(x), boundaryGap: false,
      axisLine: { lineStyle: { color: 'rgba(148,163,184,0.18)' } },
      axisLabel: { color: '#91a3b8', fontSize: 11 }
    },
    yAxis: {
      type: 'value', scale: true, axisLine: { show: false },
      splitLine: { lineStyle: { color: 'rgba(148,163,184,0.12)' } },
      axisLabel: { color: '#91a3b8', formatter: (value) => `${Number(value || 0).toFixed(2)}%` }
    },
    series: [{
      type: 'line', data: points.map(y), smooth: true, symbol: 'none',
      lineStyle: { width: 3, color },
      areaStyle: { color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [{ offset: 0, color: `${color}55` }, { offset: 1, color: `${color}05` }]) },
      markLine: { symbol: 'none', lineStyle: { color: 'rgba(148,163,184,0.24)', type: 'dashed' }, label: { color: '#91a3b8' }, data: [{ yAxis: 0 }] }
    }]
  }
}

function filterTrend(points, key) {
  if (!Array.isArray(points) || !points.length) return []
  const end = points.at(-1)?.timestamp || Date.now()
  const startOfYear = new Date(new Date(end).getFullYear(), 0, 1).getTime()
  const dayMap = { '1m': 31, '3m': 93, '6m': 186, '1y': 366 }
  if (key === 'all') return points
  if (key === 'ytd') return points.filter((item) => item.timestamp >= startOfYear)
  const start = end - (dayMap[key] || 93) * 86400000
  const filtered = points.filter((item) => item.timestamp >= start)
  return filtered.length ? filtered : points.slice(-Math.min(points.length, 60))
}

function normalizePerformanceTrend(points) {
  if (!Array.isArray(points) || !points.length) return []
  const base = Number(points[0]?.value || 0)
  return points.map((item) => ({ ...item, percent: base ? ((Number(item.value || 0) - base) / base) * 100 : 0 }))
}

function buildRepairTrend(points) {
  if (!Array.isArray(points) || !points.length) return []
  let peak = Number(points[0]?.value || 0)
  let valley = peak
  let peakAtValley = peak
  return points.map((item) => {
    const value = Number(item?.value || 0)
    if (value >= peak) {
      peak = value
      valley = value
      peakAtValley = value
      return { ...item, progress: 100 }
    }
    if (value < valley) {
      valley = value
      peakAtValley = peak
    }
    const denominator = peakAtValley - valley
    const progress = denominator > 0 ? Math.max(0, Math.min(100, ((value - valley) / denominator) * 100)) : 100
    return { ...item, progress }
  })
}

function normalizeEstimateTrend(points, confirmedUnit) {
  const base = Number(confirmedUnit || 0)
  return (Array.isArray(points) ? points : []).map((item) => {
    const explicitRate = num(item?.estimatedRate)
    const fallbackRate = base > 0 ? ((Number(item?.estimatedUnit || 0) - base) / base) * 100 : null
    return {
      ...item,
      rate: explicitRate ?? fallbackRate,
      timeLabel: formatTimeLabel(item?.time, Number(item?.timestamp || 0))
    }
  }).filter((item) => item.timeLabel)
}

function buildEstimateStats(points) {
  const rates = points.map((i) => num(i.rate)).filter((i) => i !== null)
  return { latestRate: num(points.at(-1)?.rate), highRate: rates.length ? Math.max(...rates) : null, lowRate: rates.length ? Math.min(...rates) : null, pointCount: points.length }
}

function buildRepairInfo(points) {
  if (!Array.isArray(points) || !points.length) return { peakValue: 0, valleyValue: 0, currentValue: 0, maxDrawdown: 0, progress: 0, progressText: '-', status: '暂无走势数据' }
  let peak = points[0].value; let peakAtMax = peak; let valley = peak; let maxDrawdown = 0
  for (const point of points) {
    if (point.value > peak) peak = point.value
    const drawdown = peak > 0 ? ((point.value - peak) / peak) * 100 : 0
    if (drawdown < maxDrawdown) { maxDrawdown = drawdown; peakAtMax = peak; valley = point.value }
  }
  const currentValue = points.at(-1).value
  const denominator = peakAtMax - valley
  const progress = Math.min(100, Math.max(0, denominator > 0 ? ((currentValue - valley) / denominator) * 100 : 100))
  const status = currentValue >= peakAtMax ? '已修复' : (progress >= 70 ? '修复接近完成' : (progress > 0 ? '修复中' : '仍在低位'))
  return { peakValue: peakAtMax, valleyValue: valley, currentValue, maxDrawdown: Math.abs(maxDrawdown), progress, progressText: `${progress.toFixed(0)}%`, status }
}

function inferCategoryLabel(fundType, fundName) {
  const text = `${fundType || ''} ${fundName || ''}`.toUpperCase()
  if (text.includes('货币') || text.includes('现金') || text.includes('同业存单')) return '现金类基金'
  if (text.includes('一级债')) return '一级债基'
  if (text.includes('二级债')) return '二级债基'
  if (text.includes('偏债')) return '偏债混合'
  if (text.includes('可转债')) return '可转债基金'
  if (text.includes('中短债') || text.includes('短债')) return '中短债基金'
  if (text.includes('长债')) return '长债基金'
  if (text.includes('纯债') || text.includes('债')) return '债券基金'
  if (text.includes('ETF') || text.includes('指数') || text.includes('股票') || text.includes('混合') || text.includes('QDII') || text.includes('FOF') || text.includes('REIT')) return '权益类基金'
  return '其他基金'
}

function num(value) { return value === null || value === undefined || Number.isNaN(Number(value)) ? null : Number(value) }
function formatPrice(value, digits = 4) { return num(value) === null ? '-' : Number(value).toFixed(digits) }
function formatPercent(value) { return num(value) === null ? '-' : `${Number(value) >= 0 ? '+' : ''}${Number(value).toFixed(2)}%` }
function formatPercentNoSign(value) { return num(value) === null ? '-' : `${Number(value).toFixed(1)}%` }
function formatRank(item) { return item?.rank && item?.rankTotal ? `${item.rank} / ${item.rankTotal}` : '-' }
function formatRankDelta(item) { return !item?.rankDelta ? '—' : (item.rankDeltaDirection === 'down' ? `改善 ${item.rankDelta}` : (item.rankDeltaDirection === 'up' ? `回落 ${item.rankDelta}` : `${item.rankDelta}`)) }
function rankDeltaTone(item) { return !item?.rankDelta ? '' : (item.rankDeltaDirection === 'down' ? 'profit-text' : (item.rankDeltaDirection === 'up' ? 'loss-text' : '')) }
function rankBeatPercent(item) { return Math.max(0, Math.min(100, Number(item?.rankPercentile || 0))) }
function betterThanText(item) { return item?.rankPercentile ? `超过 ${formatPercentNoSign(item.rankPercentile)} 同类基金` : '同类位置待同步' }
function numberTone(value) { return num(value) === null ? '' : (Number(value) >= 0 ? 'profit-text' : 'loss-text') }
function formatTimeLabel(text, timestamp) {
  if (typeof text === 'string' && text.trim()) return text.trim().length >= 16 ? text.trim().slice(11, 16) : text.trim()
  if (!timestamp) return ''
  const d = new Date(timestamp)
  return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

function withTimeout(promise, timeoutMs) {
  return Promise.race([
    promise,
    new Promise((_, reject) => {
      window.setTimeout(() => reject(new Error(`timeout:${timeoutMs}`)), timeoutMs)
    })
  ])
}
</script>

<style scoped>
.shell {
  display: flex;
  flex-direction: column;
  gap: 40px;
  padding-bottom: 8px;
}

.card {
  border-radius: 22px;
  border: 1px solid rgba(97, 118, 148, 0.18);
  background:
    radial-gradient(circle at top right, rgba(56, 189, 248, 0.16), transparent 32%),
    linear-gradient(180deg, rgba(18, 30, 48, 0.98), rgba(10, 17, 28, 0.98));
  box-shadow: 0 18px 40px rgba(6, 11, 18, 0.18);
}

.hero {
  padding: 24px;
  display: flex;
  justify-content: space-between;
  gap: 18px;
  align-items: flex-start;
}

.block {
  padding: 22px;
}

.title-row,
.status-row,
.head,
.rank-top,
.rank-meta,
.switches,
.tab-nav {
  display: flex;
}

.title-row,
.status-row,
.switches,
.tab-nav {
  gap: 10px;
  flex-wrap: wrap;
}

.head,
.rank-top,
.rank-meta {
  justify-content: space-between;
  gap: 12px;
  align-items: center;
}

.head {
  margin-bottom: 18px;
}

.title {
  font-size: 28px;
  font-weight: 700;
  letter-spacing: -0.03em;
}

.sub,
.muted {
  color: var(--text-secondary);
}

.sub {
  margin-top: 6px;
}

.status-row {
  margin-top: 14px;
}

.pill {
  padding: 6px 12px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
}

.pill.live {
  color: #0f172a;
  background: #89f0c3;
}

.pill.rest {
  color: #cbd5e1;
  background: rgba(148, 163, 184, 0.16);
}

.tab-nav {
  padding: 8px;
  border-radius: 20px;
  background: rgba(11, 19, 31, 0.58);
  border: 1px solid rgba(97, 118, 148, 0.14);
  width: max-content;
  max-width: 100%;
  margin-top: 0;
}

.tab-btn,
.period,
.quartile,
.pill-btn {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  padding: 6px 10px;
  font-size: 12px;
  font-weight: 600;
}

.tab-btn {
  border: 0;
  color: #b7c7d8;
  background: transparent;
  cursor: pointer;
  transition: background-color 0.2s ease, color 0.2s ease;
}

.tab-btn.active,
.period,
.pill-btn.active {
  color: #08101b;
  background: linear-gradient(90deg, #9ae6b4, #7dd3fc);
}

.stat-grid,
.mini-grid,
.ranking-grid,
.mini-metrics {
  display: grid;
  gap: 16px;
}

.stat-grid,
.mini-grid {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.ranking-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.mini-metrics {
  grid-template-columns: repeat(2, minmax(0, 1fr));
  margin-top: 18px;
}

.stat,
.mini-card,
.rank-card {
  border-radius: 18px;
  background: rgba(11, 19, 31, 0.54);
  border: 1px solid rgba(97, 118, 148, 0.14);
}

.stat,
.mini-card {
  padding: 18px;
}

.label,
.section-title {
  font-weight: 600;
}

.value {
  margin-top: 10px;
  font-size: 24px;
  font-weight: 700;
}

.mini-card strong {
  display: block;
  margin-top: 8px;
  font-size: 18px;
}

.quartile {
  color: #cbd5e1;
  background: rgba(148, 163, 184, 0.12);
}

.rank-pos {
  margin-top: 12px;
  font-size: 28px;
  font-weight: 700;
}

.track {
  margin-top: 12px;
  height: 10px;
  border-radius: 999px;
  background: rgba(100, 116, 139, 0.18);
  overflow: hidden;
}

.track-fill {
  height: 100%;
  border-radius: 999px;
  background: linear-gradient(90deg, #3ecf8e, #7dd3fc);
}

.pill-btn {
  border: 0;
  color: #b7c7d8;
  background: rgba(100, 116, 139, 0.14);
  cursor: pointer;
}

.chart {
  height: 320px;
  margin-top: 6px;
}

.empty {
  padding: 48px 0 30px;
  text-align: center;
}

.empty-title {
  font-size: 18px;
  font-weight: 600;
}

.profit-text {
  color: var(--profit);
}

.loss-text {
  color: var(--loss);
}

@media (max-width: 920px) {
  .hero,
  .head {
    flex-direction: column;
    align-items: flex-start;
  }

  .tab-nav {
    width: 100%;
    margin-top: 0;
  }

  .tab-btn {
    justify-content: center;
    flex: 1;
  }

  .stat-grid,
  .mini-grid,
  .ranking-grid,
  .mini-metrics {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
