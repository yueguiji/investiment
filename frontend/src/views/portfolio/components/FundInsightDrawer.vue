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
                <div class="muted">{{ activeEstimateSourceLabel }} · {{ estimateUpdatedAt || mergedFund.netEstimatedTime || '-' }}</div>
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
                <div class="value" :class="repairInfo.recovered ? 'profit-text' : ''">{{ repairInfo.repairDaysText }}</div>
                <div class="muted">{{ repairInfo.status }}</div>
              </div>
            </section>

            <section class="card block">
              <div class="head">
                <div>
                  <div class="section-title">今日估值走势</div>
                  <div class="muted">当前估算源：{{ activeEstimateSourceLabel }}。Tooltip 里保留估算净值和时间。</div>
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
                  <div class="section-title">业绩走势 / 回撤修复</div>
                  <div class="muted">标出区间内最大回撤，以及从回撤低点修复到前高所用天数。</div>
                </div>
                <div class="switches">
                  <button v-for="item in rangeOptions" :key="item.value" type="button" class="pill-btn" :class="{ active: rangeKey === item.value }" @click="rangeKey = item.value">{{ item.label }}</button>
                </div>
              </div>
              <div v-if="performanceTrend.length" ref="performanceChartRef" class="chart"></div>
              <div v-else class="empty muted">暂无走势数据</div>
              <div class="mini-grid">
                <div class="mini-card"><span class="muted">最大回撤</span><strong class="loss-text">{{ formatPercentNoSign(repairInfo.maxDrawdown) }}</strong></div>
                <div class="mini-card"><span class="muted">修复天数</span><strong>{{ repairInfo.repairDaysText }}</strong></div>
                <div class="mini-card"><span class="muted">前高</span><strong>{{ formatPrice(repairInfo.peakValue, 4) }}</strong></div>
                <div class="mini-card"><span class="muted">回撤低点</span><strong>{{ formatPrice(repairInfo.valleyValue, 4) }}</strong></div>
              </div>
            </section>

            <section class="card block">
              <div class="head">
                <div>
                  <div class="section-title">估算与实际对照</div>
                  <div class="muted">{{ estimateCompareDescription }}</div>
                </div>
                <div class="muted">{{ estimateCompareSummary }}</div>
              </div>
              <div v-if="estimateCompareTrend.length" ref="estimateCompareChartRef" class="chart compact-chart"></div>
              <div v-else class="empty muted">本地还没有足够的历史估算快照可对照</div>
            </section>
          </template>

          <template v-else-if="activeTab === 'holdings'">
            <section class="card block">
              <div class="head">
                <div>
                  <div class="section-title">前十大重仓股</div>
                  <div class="muted">来自基金季报持仓，辅助判断估值偏差来自哪些股票暴露。</div>
                </div>
                <div class="muted">{{ holdingsSummary }}</div>
              </div>
              <n-spin :show="holdingsLoading">
                <div v-if="fundHoldings.length" class="holdings-list">
                  <div class="holdings-row header">
                    <span>排名</span>
                    <span>股票</span>
                    <span>市场</span>
                    <span>占净值</span>
                    <span>最新价</span>
                    <span>涨跌幅</span>
                  </div>
                  <div v-for="item in fundHoldings" :key="`${item.rank}-${item.stockCode}`" class="holdings-row">
                    <span>{{ item.rank || '-' }}</span>
                    <span>
                      <strong>{{ item.stockName || '-' }}</strong>
                      <em>{{ item.stockCode || '-' }}</em>
                    </span>
                    <span>{{ item.market || '-' }}</span>
                    <span>{{ formatPercentNoSign(item.ratio) }}</span>
                    <span>{{ formatPrice(item.price, 2) }}</span>
                    <span :class="numberTone(item.changeRate)">{{ formatPercent(item.changeRate) }}</span>
                  </div>
                </div>
                <div v-else class="empty muted">暂无重仓股数据</div>
              </n-spin>
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

const props = defineProps({
  show: { type: Boolean, default: false },
  fund: { type: Object, default: null },
  estimateSource: { type: String, default: 'eastmoney_js' }
})
const emit = defineEmits(['update:show', 'refreshed'])
const message = useMessage()
const loading = ref(false)
const profile = ref(null)
const fundHoldings = ref([])
const estimateChartRef = ref(null)
const performanceChartRef = ref(null)
const estimateCompareChartRef = ref(null)
const rangeKey = ref('3m')
const activeTab = ref('overview')
const holdingsLoading = ref(false)
let estimateChart = null
let performanceChart = null
let estimateCompareChart = null

const tabs = [
  { label: '概览', value: 'overview' },
  { label: '排行与走势', value: 'analysis' },
  { label: '重仓股', value: 'holdings' },
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
const activeEstimateSource = computed(() => normalizeEstimateSource(profile.value?.estimateSource || props.estimateSource))
const activeEstimateSourceLabel = computed(() => estimateSourceOptions[activeEstimateSource.value] || '天天基金 JS')
const estimateCompareEstimateLabel = computed(() => activeEstimateSource.value === 'ai_corrected' ? 'AI纠偏估算' : '盘中估算')
const estimateCompareDescription = computed(() => `按每天最后一次${estimateCompareEstimateLabel.value}对比实际净值涨跌，帮助判断估算偏差。`)
const rawEstimateTrend = computed(() => Array.isArray(profile.value?.estimateTrend) ? profile.value.estimateTrend : [])
const estimateTrend = computed(() => normalizeEstimateTrend(rawEstimateTrend.value, mergedFund.value?.netUnitValue))
const estimateLatestRate = computed(() => num(profile.value?.estimateLatestRate) ?? num(mergedFund.value?.netEstimatedRate) ?? estimateTrend.value.at(-1)?.rate ?? null)
const estimateHasCurve = computed(() => estimateTrend.value.length >= 2)
const estimateStats = computed(() => buildEstimateStats(estimateTrend.value))
const estimateEmptyText = computed(() => estimateTrend.value.length === 1 ? '今天目前只采样到 1 个估值点，继续开着详情页或手动刷新，系统会自动补成曲线。' : (mergedFund.value?.netEstimatedTime ? `今天还没有新的盘中估值序列，最近一次估值时间是 ${mergedFund.value.netEstimatedTime}。` : '当前基金今天还没有盘中估值数据。'))
const estimateCompareTrend = computed(() => normalizeEstimateCompareTrend(profile.value?.estimateCompareTrend || []))
const estimateCompareSummary = computed(() => {
  const points = estimateCompareTrend.value
  if (!points.length) return '等待历史快照'
  const errors = points
    .map((item) => Math.abs(Number(item.estimatedRate || 0) - Number(item.actualRate || 0)))
    .filter((item) => Number.isFinite(item))
  const mae = errors.length ? errors.reduce((sum, item) => sum + item, 0) / errors.length : 0
  return `${points.length} 天样本 · 平均偏差 ${mae.toFixed(2)}%`
})
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
const repairInfo = computed(() => buildRepairInfo(performanceTrend.value))
const holdingsSummary = computed(() => {
  if (!fundHoldings.value.length) return '等待持仓数据'
  const quarter = fundHoldings.value.find((item) => item.quarter)?.quarter || ''
  const ratio = fundHoldings.value.reduce((sum, item) => sum + Number(item.ratio || 0), 0)
  return `${quarter || '最近披露'} · 合计 ${ratio.toFixed(2)}%`
})

const estimateSourceOptions = {
  eastmoney_js: '天天基金 JS',
  eastmoney_mobile: '天天基金移动端',
  danjuan_position: '蛋卷持仓模拟',
  ai_corrected: 'AI估值纠偏'
}

watch(() => [props.show, props.fund?.stockCode, props.estimateSource], async ([show, code, source], oldValue) => {
  const [prevShow, prevCode, prevSource] = Array.isArray(oldValue) ? oldValue : [false, '', '']
  if (show && code && (code !== prevCode || source !== prevSource || (!prevShow && show) || !profile.value)) {
    activeTab.value = 'overview'
    if (code !== prevCode) fundHoldings.value = []
    await loadProfile({ silent: true })
  }
}, { immediate: true })

watch([estimateTrend, performanceTrend, estimateCompareTrend], async () => {
  if (!props.show) return
  await nextTick()
  renderCharts()
}, { deep: true })

watch(activeTab, async () => {
  if (!props.show) return
  if (activeTab.value === 'holdings') await loadFundHoldings()
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
    const source = normalizeEstimateSource(props.estimateSource)
    const loader = refresh
      ? (window.go?.main?.App?.RefreshFundProfileByFundEstimateSource
        ? window.go.main.App.RefreshFundProfileByFundEstimateSource(code, source)
        : window.go.main.App.RefreshFundProfile(code))
      : (window.go?.main?.App?.GetFundProfileByFundEstimateSource
        ? window.go.main.App.GetFundProfileByFundEstimateSource(code, source)
        : window.go.main.App.GetFundProfile(code))
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
  if (activeTab.value === 'holdings') await loadFundHoldings({ force: true })
}

function handleShowUpdate(value) { emit('update:show', value) }
function openFundPage() {
  const code = mergedFund.value?.stockCode
  if (code && window.go?.main?.App?.OpenURL) window.go.main.App.OpenURL(`https://fund.eastmoney.com/${code}.html`)
}

async function loadFundHoldings(options = {}) {
  const { force = false } = options
  const code = props.fund?.stockCode
  if (!code || !window.go?.main?.App?.GetFundTop10Holdings) return
  if (fundHoldings.value.length && !force) return
  holdingsLoading.value = true
  try {
    const result = await withTimeout(window.go.main.App.GetFundTop10Holdings(code), 10000)
    fundHoldings.value = (Array.isArray(result) ? result : []).map(normalizeHoldingStock)
  } catch (error) {
    console.error(error)
    message.error('重仓股加载失败')
  } finally {
    holdingsLoading.value = false
  }
}

function normalizeHoldingStock(item) {
  return {
    rank: Number(item?.rank || 0),
    stockCode: String(item?.stockCode || '').trim(),
    stockName: String(item?.stockName || '').trim(),
    ratio: num(item?.ratio),
    quarter: String(item?.quarter || '').trim(),
    price: num(item?.price),
    changeRate: num(item?.changeRate),
    market: String(item?.market || '').trim()
  }
}

function normalizeEstimateSource(value) {
  return Object.prototype.hasOwnProperty.call(estimateSourceOptions, value) ? value : 'eastmoney_js'
}

function renderCharts() { renderEstimateChart(); renderPerformanceChart(); renderEstimateCompareChart() }
function resizeCharts() { estimateChart?.resize(); performanceChart?.resize(); estimateCompareChart?.resize() }
function disposeCharts() {
  if (estimateChart) { estimateChart.dispose(); estimateChart = null }
  if (performanceChart) { performanceChart.dispose(); performanceChart = null }
  if (estimateCompareChart) { estimateCompareChart.dispose(); estimateCompareChart = null }
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
  const repair = repairInfo.value
  const drawdownData = buildSegmentSeriesData(points, repair.peakIndex, repair.valleyIndex, 'percent')
  const repairData = buildSegmentSeriesData(points, repair.valleyIndex, repair.recoveredIndex ?? points.length - 1, 'percent')
  performanceChart.setOption({
    backgroundColor: 'transparent',
    animation: false,
    legend: {
      top: 0,
      right: 12,
      itemWidth: 18,
      itemHeight: 8,
      textStyle: { color: '#91a3b8', fontSize: 11 },
      data: ['业绩走势', '最大回撤', repair.recovered ? '修复区间' : '修复中']
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
        const drawdown = repair.peakValue > 0 ? ((Number(point?.value || 0) - repair.peakValue) / repair.peakValue) * 100 : null
        return [
          point?.date || '-',
          '业绩走势 ' + formatPercent(point?.percent),
          '相对前高回撤 ' + (drawdown === null ? '-' : formatPercent(drawdown)),
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
        axisLabel: { color: '#91a3b8', formatter: (value) => Number(value || 0).toFixed(2) + '%' }
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
        },
        markArea: buildRepairMarkArea(points, repair)
      },
      {
        name: '最大回撤',
        type: 'line',
        data: drawdownData,
        smooth: false,
        symbol: 'none',
        connectNulls: true,
        lineStyle: { width: 3, color: '#2dd4bf' },
        itemStyle: { color: '#2dd4bf' },
        label: {
          show: true,
          formatter: (params) => params.dataIndex === repair.valleyIndex ? `最大回撤${formatPercentNoSign(repair.maxDrawdown)}` : '',
          color: '#eafffb',
          backgroundColor: 'rgba(45, 212, 191, 0.82)',
          borderRadius: 4,
          padding: [5, 7],
          position: 'top'
        }
      },
      {
        name: repair.recovered ? '修复区间' : '修复中',
        type: 'line',
        data: repairData,
        smooth: false,
        symbol: 'none',
        connectNulls: true,
        lineStyle: { width: 3, color: repair.recovered ? '#fb7185' : '#f59e0b' },
        itemStyle: { color: repair.recovered ? '#fb7185' : '#f59e0b' },
        label: {
          show: true,
          formatter: (params) => params.dataIndex === (repair.recoveredIndex ?? points.length - 1) ? repair.repairDaysText : '',
          color: '#fff1f2',
          backgroundColor: repair.recovered ? 'rgba(244, 63, 94, 0.82)' : 'rgba(245, 158, 11, 0.82)',
          borderRadius: 4,
          padding: [5, 7],
          position: 'top'
        }
      }
    ]
  })
}

function renderEstimateCompareChart() {
  if (activeTab.value !== 'analysis' || !estimateCompareChartRef.value || !estimateCompareTrend.value.length) {
    if (estimateCompareChart) { estimateCompareChart.dispose(); estimateCompareChart = null }
    return
  }
  if (!estimateCompareChart) estimateCompareChart = echarts.init(estimateCompareChartRef.value)
  const points = estimateCompareTrend.value
  estimateCompareChart.setOption({
    backgroundColor: 'transparent',
    animation: false,
    legend: {
      top: 0,
      right: 12,
      itemWidth: 18,
      itemHeight: 8,
      textStyle: { color: '#91a3b8', fontSize: 11 },
      data: [estimateCompareEstimateLabel.value, '实际涨跌']
    },
    grid: { left: 18, right: 18, top: 42, bottom: 26, containLabel: true },
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(10,16,26,0.94)',
      borderColor: 'rgba(148,163,184,0.18)',
      textStyle: { color: '#e5eef9' },
      formatter: (params) => {
        const point = points[params?.[0]?.dataIndex ?? 0]
        return [
          point?.date || '-',
          estimateCompareEstimateLabel.value + ' ' + formatPercent(point?.estimatedRate),
          '实际涨跌 ' + formatPercent(point?.actualRate),
          '估算时间 ' + (point?.estimateTime || '-'),
          '来源 ' + (point?.sourceLabel || point?.source || '-')
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
    yAxis: {
      type: 'value',
      scale: true,
      axisLine: { show: false },
      splitLine: { lineStyle: { color: 'rgba(148,163,184,0.12)' } },
      axisLabel: { color: '#91a3b8', formatter: (value) => `${Number(value || 0).toFixed(2)}%` }
    },
    series: [
      {
        name: estimateCompareEstimateLabel.value,
        type: 'line',
        data: points.map((p) => p.estimatedRate),
        smooth: true,
        symbol: 'circle',
        symbolSize: 4,
        lineStyle: { width: 2, color: '#60a5fa' }
      },
      {
        name: '实际涨跌',
        type: 'line',
        data: points.map((p) => p.actualRate),
        smooth: true,
        symbol: 'circle',
        symbolSize: 4,
        lineStyle: { width: 2, color: '#34d399' }
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
  if (!Array.isArray(points) || !points.length) {
    return {
      peakValue: 0,
      valleyValue: 0,
      currentValue: 0,
      maxDrawdown: 0,
      peakIndex: 0,
      valleyIndex: 0,
      recoveredIndex: null,
      recovered: false,
      repairDays: null,
      repairDaysText: '-',
      status: '暂无走势数据'
    }
  }

  let runningPeak = Number(points[0]?.value || 0)
  let runningPeakIndex = 0
  let peakIndex = 0
  let valleyIndex = 0
  let maxDrawdown = 0

  points.forEach((point, index) => {
    const value = Number(point?.value || 0)
    if (value > runningPeak) {
      runningPeak = value
      runningPeakIndex = index
    }
    const drawdown = runningPeak > 0 ? ((value - runningPeak) / runningPeak) * 100 : 0
    if (drawdown < maxDrawdown) {
      maxDrawdown = drawdown
      peakIndex = runningPeakIndex
      valleyIndex = index
    }
  })

  const peakValue = Number(points[peakIndex]?.value || 0)
  const valleyValue = Number(points[valleyIndex]?.value || peakValue)
  let recoveredIndex = null
  if (Math.abs(maxDrawdown) > 0) {
    for (let i = valleyIndex + 1; i < points.length; i += 1) {
      if (Number(points[i]?.value || 0) >= peakValue) {
        recoveredIndex = i
        break
      }
    }
  }

  const currentValue = Number(points.at(-1)?.value || 0)
  const recovered = recoveredIndex !== null
  const repairDays = recovered
    ? daysBetween(points[valleyIndex]?.date, points[recoveredIndex]?.date)
    : daysBetween(points[valleyIndex]?.date, points.at(-1)?.date)
  const repairDaysText = Math.abs(maxDrawdown) <= 0
    ? '无明显回撤'
    : (recovered ? `${repairDays}天修复` : `修复中 ${repairDays}天`)
  const status = Math.abs(maxDrawdown) <= 0
    ? '区间内未形成回撤'
    : (recovered ? `从 ${shortDate(points[valleyIndex]?.date)} 到 ${shortDate(points[recoveredIndex]?.date)}` : `低点后尚未回到前高 ${formatPrice(peakValue, 4)}`)

  return {
    peakValue,
    valleyValue,
    currentValue,
    maxDrawdown: Math.abs(maxDrawdown),
    peakIndex,
    valleyIndex,
    recoveredIndex,
    recovered,
    repairDays,
    repairDaysText,
    status
  }
}

function buildSegmentSeriesData(points, startIndex, endIndex, key) {
  if (!Array.isArray(points) || !points.length || startIndex === null || startIndex === undefined || endIndex === null || endIndex === undefined) {
    return []
  }
  return points.map((point, index) => (index >= startIndex && index <= endIndex ? point?.[key] : null))
}

function buildRepairMarkArea(points, repair) {
  if (!repair?.recovered || !Array.isArray(points) || !points[repair.valleyIndex] || !points[repair.recoveredIndex]) {
    return undefined
  }
  return {
    silent: true,
    itemStyle: { color: 'rgba(244, 63, 94, 0.10)' },
    data: [[
      { xAxis: points[repair.valleyIndex].date.slice(5) },
      { xAxis: points[repair.recoveredIndex].date.slice(5) }
    ]]
  }
}

function daysBetween(startDate, endDate) {
  const start = parseDateOnly(startDate)
  const end = parseDateOnly(endDate)
  if (!start || !end) return 0
  return Math.max(0, Math.round((end - start) / 86400000))
}

function parseDateOnly(value) {
  if (!value) return null
  const match = String(value).match(/^(\d{4})-(\d{2})-(\d{2})/)
  if (!match) return null
  return new Date(Number(match[1]), Number(match[2]) - 1, Number(match[3]))
}

function shortDate(value) {
  return value ? String(value).slice(5, 10) : '-'
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

function normalizeEstimateCompareTrend(points) {
  return (Array.isArray(points) ? points : [])
    .map((item) => ({
      ...item,
      estimatedRate: num(item?.estimatedRate),
      actualRate: num(item?.actualRate),
      sourceLabel: estimateSourceLabel(item?.source)
    }))
    .filter((item) => item.date && item.estimatedRate !== null && item.actualRate !== null)
}

function estimateSourceLabel(source) {
  switch (source) {
    case 'eastmoney_js':
      return '天天基金 JS'
    case 'eastmoney_mobile':
      return '天天基金移动端'
    case 'danjuan_position':
      return '蛋卷持仓模拟'
    case 'ai_corrected':
      return 'AI估值纠偏'
    case 'eastmoney':
      return '天天基金估算'
    default:
      return source || '-'
  }
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

.holdings-list {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid rgba(97, 118, 148, 0.14);
  border-radius: 18px;
}

.holdings-row {
  display: grid;
  grid-template-columns: 56px minmax(150px, 1.6fr) 72px 88px 92px 92px;
  gap: 12px;
  align-items: center;
  padding: 14px 16px;
  background: rgba(11, 19, 31, 0.38);
  border-bottom: 1px solid rgba(97, 118, 148, 0.10);
}

.holdings-row:last-child {
  border-bottom: 0;
}

.holdings-row.header {
  color: var(--text-secondary);
  font-size: 12px;
  font-weight: 600;
  background: rgba(148, 163, 184, 0.08);
}

.holdings-row strong,
.holdings-row em {
  display: block;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.holdings-row em {
  margin-top: 4px;
  color: var(--text-secondary);
  font-style: normal;
  font-size: 12px;
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

  .holdings-row {
    grid-template-columns: 44px minmax(120px, 1.4fr) 58px 72px;
  }

  .holdings-row span:nth-child(5),
  .holdings-row span:nth-child(6) {
    display: none;
  }
}
</style>
