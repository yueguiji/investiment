<template>
  <div class="fade-in debt-plans">
    <n-page-header title="负债计划" subtitle="集中查看原始月供、真实月供占用和还款计划">
      <template #extra>
        <n-space>
          <n-button @click="$router.push('/asset/detail?tab=liabilities')">管理负债台账</n-button>
          <n-button @click="$router.push('/asset/benchmarks')">基准数据页</n-button>
          <n-button secondary @click="loadData">刷新</n-button>
        </n-space>
      </template>
    </n-page-header>

    <n-grid :cols="4" :x-gap="16" :y-gap="16" style="margin-top: 20px;">
      <n-gi v-for="item in metricCards" :key="item.label">
        <div class="metric-card" :style="{ '--accent': item.accent }">
          <div class="metric-label">{{ item.label }}</div>
          <div class="metric-value" :class="item.tone">{{ item.value }}</div>
          <div class="metric-sub">{{ item.sub }}</div>
          <div v-if="item.detail" class="metric-detail">{{ item.detail }}</div>
        </div>
      </n-gi>
    </n-grid>

    <n-grid :cols="2" :x-gap="16" :y-gap="16" style="margin-top: 20px;">
      <n-gi>
        <div class="platform-card chart-card">
          <div class="section-title">负债趋势图</div>
          <div class="section-sub">用趋势看未来还款压力、剩余本金下降节奏和利息变化。</div>
          <div v-if="liabilityTrend.length" class="chart-shell">
            <div ref="trendChartRef" class="chart-canvas"></div>
          </div>
          <div v-else class="empty-text">暂无可展示的负债趋势数据。</div>
        </div>
      </n-gi>
      <n-gi>
        <div class="platform-card chart-card">
          <div class="section-title">收入与支出</div>
          <div class="section-sub">优先展示真实月供占用，并保留原始月供作为参考口径。</div>
          <div class="breakdown-grid">
            <div class="breakdown-item"><span>月收入</span><strong>{{ formatMoney(summary.monthlyIncome) }}</strong></div>
            <div class="breakdown-item"><span>月净收入</span><strong>{{ formatMoney(summary.monthlyNetIncome) }}</strong></div>
            <div class="breakdown-item"><span>真实月供占用</span><strong>{{ formatMoney(effectiveDebtPayment) }}</strong></div>
            <div class="breakdown-item"><span>原始月供</span><strong>{{ formatMoney(summary.monthlyDebtPayment) }}</strong></div>
            <div class="breakdown-item"><span>真实占比</span><strong>{{ formatPercent(realDebtRatio) }}</strong></div>
            <div class="breakdown-item"><span>原始占比</span><strong>{{ formatPercent(originalDebtRatio) }}</strong></div>
          </div>
          <div class="formula-strip">税后 {{ formatMoney(summary.monthlyNetIncome) }} - 真实月供 {{ formatMoney(effectiveDebtPayment) }} = 剩余可支配 {{ formatMoney(freeCashAfterDebt) }}</div>
        </div>
      </n-gi>
    </n-grid>

    <n-grid :cols="2" :x-gap="16" :y-gap="16" style="margin-top: 20px;">
      <n-gi>
        <div class="platform-card">
          <div class="section-title">当前还款结构</div>
          <div class="section-sub">聚合当前所有负债在本月的本金、利息和应还金额。</div>
          <div class="breakdown-grid">
            <div class="breakdown-item"><span>本期应还</span><strong>{{ formatMoney(currentBreakdown.paymentAmount) }}</strong></div>
            <div class="breakdown-item"><span>本金</span><strong>{{ formatMoney(currentBreakdown.principalPaid) }}</strong></div>
            <div class="breakdown-item"><span>利息</span><strong>{{ formatMoney(currentBreakdown.interestPaid) }}</strong></div>
            <div class="breakdown-item"><span>期末剩余</span><strong>{{ formatMoney(currentBreakdown.closingPrincipal) }}</strong></div>
          </div>
        </div>
      </n-gi>
      <n-gi>
        <div class="platform-card">
          <div class="section-title">当前选择负债</div>
          <div class="section-sub">{{ selectedLiabilityName }}</div>
          <div class="breakdown-grid">
            <div class="breakdown-item"><span>贷款机构</span><strong>{{ selectedLiability?.lender || '-' }}</strong></div>
            <div class="breakdown-item"><span>剩余本金</span><strong>{{ formatMoney(selectedLiability?.outstandingPrincipal) }}</strong></div>
            <div class="breakdown-item"><span>原始月供</span><strong>{{ formatMoney(totalMonthlyPayment(selectedLiability)) }}</strong></div>
            <div class="breakdown-item"><span>利率/期限</span><strong>{{ selectedLiabilityRateText }}</strong></div>
          </div>
        </div>
      </n-gi>
    </n-grid>

    <div class="platform-card" style="margin-top: 20px;">
      <div class="section-row">
        <div>
          <div class="section-title">负债列表</div>
          <div class="section-sub">这里展示每笔负债的原始月供、剩余本金和期限，点击可切换下方还款计划。</div>
        </div>
      </div>
      <n-data-table :columns="liabilityColumns" :data="liabilities" :pagination="{ pageSize: 8 }" :bordered="false" />
    </div>

    <div class="platform-card" style="margin-top: 20px;">
      <div class="section-row">
        <div>
          <div class="section-title">{{ selectedLiabilityName }}</div>
          <div class="section-sub">展示当前所选负债的自动摊还计划。</div>
        </div>
      </div>
      <n-data-table :columns="scheduleColumns" :data="selectedSchedules" :pagination="{ pageSize: 10 }" :bordered="false" />
    </div>
  </div>
</template>

<script setup>
import { computed, h, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { NButton, useMessage } from 'naive-ui'
import * as echarts from 'echarts'

const message = useMessage()
const summary = ref({ totalLiabilities: 0, monthlyDebtPayment: 0, monthlyEffectiveDebtPayment: 0, monthlyIncome: 0, monthlyNetIncome: 0, monthlyCoverageRate: 0, monthlyEffectiveCoverageRate: 0, liabilityCount: 0 })
const liabilities = ref([])
const liabilityTrend = ref([])
const schedulesByLiability = ref({})
const selectedLiabilityId = ref(null)
const trendChartRef = ref(null)
let trendChartInstance = null

const effectiveDebtPayment = computed(() => Number(summary.value.monthlyEffectiveDebtPayment || summary.value.monthlyDebtPayment || 0))
const averageLiability = computed(() => { const count = Number(summary.value.liabilityCount || 0); return count ? Number(summary.value.totalLiabilities || 0) / count : 0 })
const originalDebtRatio = computed(() => { const income = Number(summary.value.monthlyNetIncome || 0); return income > 0 ? Number(summary.value.monthlyDebtPayment || 0) / income * 100 : 0 })
const realDebtRatio = computed(() => { const income = Number(summary.value.monthlyNetIncome || 0); return income > 0 ? effectiveDebtPayment.value / income * 100 : 0 })
const freeCashAfterDebt = computed(() => Number(summary.value.monthlyNetIncome || 0) - effectiveDebtPayment.value)
const selectedSchedules = computed(() => schedulesByLiability.value[selectedLiabilityId.value] || [])
const selectedLiability = computed(() => liabilities.value.find((item) => item.ID === selectedLiabilityId.value) || null)
const selectedLiabilityName = computed(() => selectedLiability.value ? `${selectedLiability.value.name} 还款计划` : '还款计划明细')
const selectedLiabilityRateText = computed(() => selectedLiability.value ? `${Number(selectedLiability.value.annualRate || 0).toFixed(2)}% / ${selectedLiability.value.loanTermMonths || 0}期` : '-')
const metricCards = computed(() => [
  { label: '总负债', value: formatMoney(summary.value.totalLiabilities), sub: `${summary.value.liabilityCount || 0} 笔在管贷款`, detail: '当前所有房贷、车贷和消费贷合计', tone: 'loss', accent: '#ef4444' },
  { label: '真实月供占用', value: formatMoney(effectiveDebtPayment.value), sub: `原始月供 ${formatMoney(summary.value.monthlyDebtPayment)}`, detail: '已扣除固定公积金回流后的真实占用', accent: '#8b5cf6' },
  { label: '真实收入覆盖倍数', value: `${Number(summary.value.monthlyEffectiveCoverageRate || summary.value.monthlyCoverageRate || 0).toFixed(2)}x`, sub: `税后收入 ${formatMoney(summary.value.monthlyNetIncome)}`, detail: `原始覆盖 ${Number(summary.value.monthlyCoverageRate || 0).toFixed(2)}x`, tone: Number(summary.value.monthlyEffectiveCoverageRate || summary.value.monthlyCoverageRate || 0) >= 2 ? 'profit' : 'loss', accent: '#10b981' },
  { label: '平均单笔负债', value: formatMoney(averageLiability.value), sub: '便于判断负债集中度', detail: '可结合趋势图看后续下降速度', accent: '#3b82f6' }
])
const currentBreakdown = computed(() => Object.values(schedulesByLiability.value).reduce((acc, rows) => {
  const currentMonth = new Date().toISOString().slice(0, 7)
  const current = rows?.find((item) => String(item.dueDate || '').slice(0, 7) === currentMonth) || rows?.[0]
  if (!current) return acc
  acc.paymentAmount += Number(current.paymentAmount || 0)
  acc.principalPaid += Number(current.principalPaid || 0)
  acc.interestPaid += Number(current.interestPaid || 0)
  acc.closingPrincipal += Number(current.closingPrincipal || 0)
  return acc
}, { paymentAmount: 0, principalPaid: 0, interestPaid: 0, closingPrincipal: 0 }))
const liabilityColumns = [{ title: '名称', key: 'name', width: 160 }, { title: '类型', key: 'liabilityType', width: 100, render: (row) => liabilityTypeLabel(row.liabilityType) }, { title: '贷款机构', key: 'lender', width: 120, render: (row) => row.lender || '-' }, { title: '剩余本金', key: 'outstandingPrincipal', width: 150, render: (row) => formatMoney(row.outstandingPrincipal) }, { title: '原始月供', key: 'monthlyPayment', width: 130, render: (row) => formatMoney(totalMonthlyPayment(row)) }, { title: '利率/期限', key: 'annualRate', width: 140, render: (row) => `${Number(row.annualRate || 0).toFixed(2)}% / ${row.loanTermMonths || 0}期` }, { title: '操作', key: 'actions', width: 120, render: (row) => h(NButton, { size: 'tiny', quaternary: true, type: 'primary', onClick: () => selectLiability(row.ID) }, () => '查看计划') }]
const scheduleColumns = [{ title: '期次', key: 'periodNumber', width: 70 }, { title: '还款日', key: 'dueDate', width: 120, render: (row) => formatDate(row.dueDate) }, { title: '期初本金', key: 'openingPrincipal', width: 140, render: (row) => formatMoney(row.openingPrincipal) }, { title: '本金', key: 'principalPaid', width: 120, render: (row) => formatMoney(row.principalPaid) }, { title: '利息', key: 'interestPaid', width: 120, render: (row) => formatMoney(row.interestPaid) }, { title: '本期还款', key: 'paymentAmount', width: 130, render: (row) => formatMoney(row.paymentAmount) }, { title: '期末本金', key: 'closingPrincipal', width: 140, render: (row) => formatMoney(row.closingPrincipal) }]

function totalMonthlyPayment(row) { return Number(row?.monthlyPayment || 0) + Number(row?.extraMonthlyPayment || 0) }
function formatMoney(value) { return `¥ ${Number(value || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}` }
function formatPercent(value) { return `${Number(value || 0).toFixed(2)}%` }
function formatDate(value) { return value ? new Date(value).toLocaleDateString('zh-CN') : '-' }
function liabilityTypeLabel(value) { return { mortgage: '房贷', car_loan: '车贷', consumer_loan: '消费贷', other: '其他负债' }[value] || value || '-' }
function ensureChart(el, existing) { if (!el) return existing; if (existing) return existing; return echarts.init(el) }
function renderTrendChart() {
  if (!trendChartRef.value || !liabilityTrend.value.length) return
  trendChartInstance = ensureChart(trendChartRef.value, trendChartInstance)
  trendChartInstance.setOption({ animation: false, tooltip: { trigger: 'axis' }, legend: { top: 0, textStyle: { color: '#cbd5e1' } }, grid: { left: 56, right: 24, top: 48, bottom: 40 }, xAxis: { type: 'category', data: liabilityTrend.value.map((item) => item.month), axisLabel: { color: '#94a3b8', rotate: 35 }, axisLine: { lineStyle: { color: 'rgba(148,163,184,0.2)' } } }, yAxis: [{ type: 'value', name: '剩余本金', axisLabel: { color: '#94a3b8', formatter: (value) => `${Math.round(value / 10000)}万` }, splitLine: { lineStyle: { color: 'rgba(148,163,184,0.08)' } } }, { type: 'value', name: '月供', axisLabel: { color: '#94a3b8' }, splitLine: { show: false } }], series: [{ name: '剩余本金', type: 'line', smooth: true, data: liabilityTrend.value.map((item) => Number(item.totalOutstanding || 0)), lineStyle: { width: 3, color: '#38bdf8' }, itemStyle: { color: '#38bdf8' }, areaStyle: { color: 'rgba(56,189,248,0.16)' } }, { name: '月供', type: 'bar', yAxisIndex: 1, data: liabilityTrend.value.map((item) => Number(item.totalPayment || 0)), itemStyle: { color: 'rgba(139,92,246,0.72)', borderRadius: [4, 4, 0, 0] }, barMaxWidth: 22 }] })
}
async function callApp(name, ...args) { const fn = window.go?.main?.App?.[name]; if (!fn) throw new Error(`missing app method: ${name}`); return fn(...args) }
async function loadData() { summary.value = { ...summary.value, ...((await callApp('GetHouseholdDashboardSummary')) || {}) }; liabilities.value = (await callApp('GetHouseholdLiabilities')) || []; liabilityTrend.value = (await callApp('GetHouseholdLiabilityTrend', 0, 23)) || []; const pairs = await Promise.all(liabilities.value.map(async (item) => [item.ID, (await callApp('GetHouseholdLiabilitySchedules', item.ID)) || []])); schedulesByLiability.value = Object.fromEntries(pairs); if (!selectedLiabilityId.value && liabilities.value.length) selectedLiabilityId.value = liabilities.value[0].ID; await nextTick(); renderTrendChart() }
function selectLiability(id) { selectedLiabilityId.value = id }
function handleResize() { trendChartInstance?.resize() }
watch(liabilityTrend, async () => { await nextTick(); renderTrendChart() }, { deep: true })
onMounted(async () => { try { window.addEventListener('resize', handleResize); await loadData() } catch (error) { console.error(error); message.error('加载负债计划失败') } })
onBeforeUnmount(() => { window.removeEventListener('resize', handleResize); trendChartInstance?.dispose() })
</script>

<style scoped>
.debt-plans { max-width: 1280px; margin: 0 auto; }
.metric-card { position: relative; min-height: 158px; padding: 20px; border-radius: var(--radius-md); border: 1px solid rgba(148, 163, 184, 0.16); background: linear-gradient(180deg, rgba(15, 23, 42, 0.96), rgba(30, 41, 59, 0.9)); overflow: hidden; }
.metric-card::before { content: ''; position: absolute; inset: 0 auto 0 0; width: 4px; background: var(--accent, #38bdf8); }
.metric-label { font-size: 13px; color: var(--text-secondary); }
.metric-value { margin-top: 12px; font-size: 30px; font-weight: 700; font-family: var(--font-mono); color: #f8fafc; }
.metric-value.profit { color: var(--accent-success); }
.metric-value.loss { color: #f87171; }
.metric-sub, .metric-detail { margin-top: 8px; font-size: 12px; color: var(--text-muted); line-height: 1.6; }
.metric-detail { color: rgba(226, 232, 240, 0.78); }
.platform-card { background: var(--bg-card); border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 20px; }
.section-row { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; margin-bottom: 18px; }
.section-title { font-size: 16px; font-weight: 700; }
.section-sub { margin-top: 6px; font-size: 12px; color: var(--text-muted); line-height: 1.6; }
.chart-card { min-height: 360px; }
.chart-shell { height: 280px; margin-top: 16px; }
.chart-canvas { width: 100%; height: 100%; }
.breakdown-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; margin-top: 16px; }
.breakdown-item { padding: 14px; border-radius: var(--radius-sm); background: rgba(15, 23, 42, 0.52); border: 1px solid rgba(148, 163, 184, 0.12); }
.breakdown-item span { display: block; color: var(--text-secondary); font-size: 12px; }
.breakdown-item strong { display: block; margin-top: 8px; font-size: 24px; font-family: var(--font-mono); }
.formula-strip { margin-top: 16px; padding: 12px 14px; border-radius: var(--radius-sm); background: rgba(16, 185, 129, 0.08); color: #d1fae5; font-size: 12px; line-height: 1.6; }
.empty-text { color: var(--text-muted); font-size: 13px; line-height: 1.6; padding-top: 24px; }
@media (max-width: 980px) { .section-row { flex-direction: column; } .breakdown-grid { grid-template-columns: 1fr; } }
</style>
