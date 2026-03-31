<template>
  <div class="fade-in household-overview">
    <n-page-header title="资产总览" subtitle="围绕家庭资产、负债、保障和数字分析的统一总览">
      <template #extra>
        <n-space>
          <n-button @click="$router.push('/asset/detail')">新增台账</n-button>
          <n-button secondary @click="handleRefresh">刷新快照</n-button>
          <n-button @click="$router.push('/asset/debt-plans')">负债计划</n-button>
          <n-button @click="$router.push('/asset/benchmarks')">基准数据</n-button>
          <n-button type="primary" @click="$router.push('/asset/detail')">进入资产台账</n-button>
        </n-space>
      </template>
    </n-page-header>

    <n-grid class="summary-grid" :cols="3" :x-gap="16" :y-gap="16">
      <n-gi v-for="item in metricCards" :key="item.label">
        <div class="metric-card" :style="{ '--accent': item.accent }">
          <div class="metric-topline">
            <span class="metric-label">{{ item.label }}</span>
            <span v-if="item.badge" class="metric-badge">{{ item.badge }}</span>
          </div>
          <div class="metric-value" :class="item.tone">{{ item.value }}</div>
          <div class="metric-sub">{{ item.sub }}</div>
          <div v-if="item.detail" class="metric-detail">{{ item.detail }}</div>
        </div>
      </n-gi>
    </n-grid>

    <n-grid :cols="2" :x-gap="16" :y-gap="16" style="margin-top: 20px;">
      <n-gi>
        <div class="platform-card chart-card">
          <div class="section-row">
            <div>
              <div class="section-title">家庭负债曲线</div>
              <div class="section-sub">观察未来 24 个月的剩余本金和月供变化，方便判断负债压力拐点。</div>
            </div>
            <n-button text type="primary" @click="$router.push('/asset/debt-plans')">查看计划页</n-button>
          </div>
          <div v-if="liabilityTrend.length" class="chart-shell">
            <div ref="liabilityTrendChartRef" class="chart-canvas"></div>
          </div>
          <div v-else class="empty-text">暂无可展示的负债趋势数据。</div>
        </div>
      </n-gi>
      <n-gi>
        <div class="platform-card">
          <div class="section-row">
            <div>
              <div class="section-title">重点负债</div>
              <div class="section-sub">优先看剩余本金、原始月供和期限。</div>
            </div>
            <n-button text type="primary" @click="$router.push('/asset/debt-plans')">查看详情</n-button>
          </div>
          <n-data-table :columns="liabilityColumns" :data="liabilities" :pagination="{ pageSize: 5 }" :bordered="false" />
        </div>
      </n-gi>
    </n-grid>

    <n-grid :cols="2" :x-gap="16" :y-gap="16" style="margin-top: 20px;">
      <n-gi>
        <div class="platform-card">
          <div class="section-row">
            <div>
              <div class="section-title">保障沉淀</div>
              <div class="section-sub">五险一金、年金和商业保险统一沉淀在这里。</div>
            </div>
            <n-button text type="primary" @click="$router.push('/asset/detail?tab=protections')">去管理</n-button>
          </div>
          <n-data-table :columns="protectionColumns" :data="protections" :pagination="{ pageSize: 5 }" :bordered="false" />
        </div>
      </n-gi>
      <n-gi>
        <div class="platform-card benchmark-card">
          <div class="section-row">
            <div>
              <div class="section-title">地区基准观察</div>
              <div class="section-sub">收入、负债、资产和保障基准会一起进入数字分析。</div>
            </div>
            <n-button text type="primary" @click="$router.push('/asset/benchmarks')">管理基准</n-button>
          </div>
          <div v-if="benchmarkHighlights.length" class="benchmark-grid">
            <div v-for="item in benchmarkHighlights" :key="`${item.region}-${item.name}-${item.year}`" class="benchmark-item">
              <div class="benchmark-topline">
                <span class="benchmark-region">{{ item.region }}</span>
                <span class="benchmark-category">{{ benchmarkCategoryLabel(item.category) }}</span>
              </div>
              <div class="benchmark-name">{{ item.name }}</div>
              <div class="benchmark-value">{{ formatCompactMetric(item.value, item.unit) }}</div>
              <div class="benchmark-sub">{{ item.year }} / {{ item.version || 'manual' }}</div>
            </div>
          </div>
          <div v-else class="empty-text">暂无基准数据。</div>
        </div>
      </n-gi>
    </n-grid>

    <div class="platform-card ai-panel-card" style="margin-top: 20px;">
      <div class="section-row">
        <div>
          <div class="section-title">数字分析</div>
          <div class="section-sub">把当前家庭资产、负债、税后收入和地区基准一起带入数字分析。</div>
        </div>
        <n-button type="primary" @click="$router.push('/asset/digital-analysis')">进入数字分析</n-button>
      </div>
      <div class="analysis-entry-grid">
        <div class="analysis-entry-item">
          <span>当前地区</span>
          <strong>{{ profile.region || defaultRegion }}</strong>
          <em>用于地区基准对比</em>
        </div>
        <div class="analysis-entry-item">
          <span>家庭画像</span>
          <strong>{{ profile.householdName || '我的家庭' }}</strong>
          <em>成员、支出、抵扣都会进入分析</em>
        </div>
        <div class="analysis-entry-item">
          <span>最近状态</span>
          <strong><span class="analysis-status-chip" :class="analysisStatusTone(latestRecord?.status)">{{ analysisStatusLabel(latestRecord?.status) }}</span></strong>
          <em>{{ latestRecord?.triggerSource || '等待首次分析' }}</em>
        </div>
      </div>
      <div class="analysis-entry-actions">
        <n-button secondary @click="$router.push('/asset/benchmarks')">管理基准</n-button>
        <n-button secondary @click="$router.push('/asset/detail')">完善台账</n-button>
        <n-button type="primary" :loading="runningAnalysis" @click="runAnalysis('asset-overview:manual')">手动触发</n-button>
      </div>
      <n-alert v-if="analysisMessage" :type="analysisMessageType" :show-icon="false" style="margin-top: 16px;">{{ analysisMessage }}</n-alert>
    </div>
  </div>
</template>

<script setup>
import { computed, h, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { NButton, NTag, useMessage } from 'naive-ui'
import * as echarts from 'echarts'

const message = useMessage()
const defaultRegion = '天津市'
const summary = ref({ totalAssets: 0, totalLiquidAssets: 0, totalFixedAssets: 0, totalProtection: 0, totalLiabilities: 0, netAssets: 0, debtRatio: 0, monthlyIncome: 0, monthlyNetIncome: 0, monthlyDebtPayment: 0, monthlyCoverageRate: 0, monthlyEffectiveDebtPayment: 0, monthlyEffectiveCoverageRate: 0, accountCount: 0, fixedAssetCount: 0, protectionCount: 0, liabilityCount: 0 })
const liabilities = ref([])
const protections = ref([])
const benchmarks = ref([])
const liabilityTrend = ref([])
const profile = ref({ householdName: '我的家庭', region: defaultRegion })
const latestRecord = ref(null)
const runningAnalysis = ref(false)
const analysisMessage = ref('')
const analysisMessageType = ref('info')
const liabilityTrendChartRef = ref(null)
let liabilityTrendChartInstance = null

const metricCards = computed(() => {
  const effectiveCoverage = Number(summary.value.monthlyEffectiveCoverageRate || summary.value.monthlyCoverageRate || 0)
  const rawCoverage = Number(summary.value.monthlyCoverageRate || 0)
  const effectiveDebt = Number(summary.value.monthlyEffectiveDebtPayment || summary.value.monthlyDebtPayment || 0)
  const rawDebt = Number(summary.value.monthlyDebtPayment || 0)
  const totalLedgerCount = Number(summary.value.accountCount || 0) + Number(summary.value.fixedAssetCount || 0) + Number(summary.value.liabilityCount || 0)
  return [
    { label: '总资产', value: formatMoney(summary.value.totalAssets), sub: `流动 ${formatMoney(summary.value.totalLiquidAssets)} / 固定 ${formatMoney(summary.value.totalFixedAssets)}`, detail: '覆盖现金、房产、车辆、股权等家庭资产', accent: '#14b8a6' },
    { label: '净资产', value: formatMoney(summary.value.netAssets), sub: `负债 ${formatMoney(summary.value.totalLiabilities)} / 负债率 ${formatPercent(summary.value.debtRatio)}`, detail: '净资产 = 总资产 - 总负债', accent: '#38bdf8', tone: summary.value.netAssets >= 0 ? 'profit' : 'loss' },
    { label: '真实月度现金流覆盖', value: `${effectiveCoverage.toFixed(2)}x`, sub: `税后 ${formatMoney(summary.value.monthlyNetIncome)} / 真实月供 ${formatMoney(effectiveDebt)}`, detail: `原始月供 ${formatMoney(rawDebt)} / 原始覆盖 ${rawCoverage.toFixed(2)}x`, badge: '新口径', accent: '#8b5cf6', tone: effectiveCoverage >= 2 ? 'profit' : 'loss' },
    { label: '保障沉淀', value: formatMoney(summary.value.totalProtection), sub: `${summary.value.protectionCount || 0} 项保障与福利沉淀`, detail: '包含社保、公积金、年金与商业保险', accent: '#f59e0b' },
    { label: '台账规模', value: `${totalLedgerCount}`, sub: `${summary.value.accountCount || 0} 账户 / ${summary.value.fixedAssetCount || 0} 固定资产 / ${summary.value.liabilityCount || 0} 负债`, detail: '帮助确认资产与负债录入完整度', accent: '#3b82f6' },
    { label: '快照数量', value: String(benchmarks.value.length || 0), sub: '用于地区基准、对比和数字分析', detail: profile.value.region || defaultRegion, accent: '#ef4444' }
  ]
})
const benchmarkHighlights = computed(() => benchmarks.value.slice(0, 6))
const liabilityColumns = [{ title: '名称', key: 'name', width: 160 }, { title: '类型', key: 'liabilityType', width: 100, render: (row) => liabilityTypeLabel(row.liabilityType) }, { title: '剩余本金', key: 'outstandingPrincipal', width: 150, render: (row) => formatMoney(row.outstandingPrincipal) }, { title: '原始月供', key: 'monthlyPayment', width: 130, render: (row) => formatMoney(totalMonthlyPayment(row)) }, { title: '利率/期限', key: 'annualRate', width: 130, render: (row) => `${Number(row.annualRate || 0).toFixed(2)}% / ${row.loanTermMonths || 0}期` }]
const protectionColumns = [{ title: '名称', key: 'name', width: 160 }, { title: '类型', key: 'protectionType', width: 120, render: (row) => h(NTag, { size: 'small', round: true, type: protectionTagType(row.protectionType) }, () => protectionTypeLabel(row.protectionType)) }, { title: '当前余额', key: 'currentBalance', width: 130, render: (row) => formatMoney(row.currentBalance) }, { title: '月缴/保费', key: 'monthlyValue', width: 130, render: (row) => formatMoney(Number(row.monthlyPremium || 0) + Number(row.monthlyPersonalContribution || 0) + Number(row.monthlyEmployerContribution || 0)) }]

function totalMonthlyPayment(row) { return Number(row?.monthlyPayment || 0) + Number(row?.extraMonthlyPayment || 0) }
function ensureChart(el, existing) { if (!el) return existing; if (existing) return existing; return echarts.init(el) }
function renderLiabilityTrendChart() {
  if (!liabilityTrendChartRef.value || !liabilityTrend.value.length) return
  liabilityTrendChartInstance = ensureChart(liabilityTrendChartRef.value, liabilityTrendChartInstance)
  liabilityTrendChartInstance.setOption({ animation: false, tooltip: { trigger: 'axis' }, legend: { top: 0, textStyle: { color: '#cbd5e1' } }, grid: { left: 56, right: 24, top: 48, bottom: 40 }, xAxis: { type: 'category', data: liabilityTrend.value.map((item) => item.month), axisLabel: { color: '#94a3b8', rotate: 35 }, axisLine: { lineStyle: { color: 'rgba(148,163,184,0.2)' } } }, yAxis: [{ type: 'value', name: '剩余本金', axisLabel: { color: '#94a3b8', formatter: (value) => `${Math.round(value / 10000)}万` }, splitLine: { lineStyle: { color: 'rgba(148,163,184,0.08)' } } }, { type: 'value', name: '月供', axisLabel: { color: '#94a3b8' }, splitLine: { show: false } }], series: [{ name: '剩余本金', type: 'line', smooth: true, data: liabilityTrend.value.map((item) => Number(item.totalOutstanding || 0)), lineStyle: { width: 3, color: '#38bdf8' }, itemStyle: { color: '#38bdf8' }, areaStyle: { color: 'rgba(56,189,248,0.16)' } }, { name: '月供', type: 'bar', yAxisIndex: 1, data: liabilityTrend.value.map((item) => Number(item.totalPayment || 0)), itemStyle: { color: 'rgba(139,92,246,0.7)', borderRadius: [4, 4, 0, 0] }, barMaxWidth: 22 }] })
}
function formatMoney(value) { return `¥ ${Number(value || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}` }
function formatPercent(value) { return `${Number(value || 0).toFixed(2)}%` }
function formatCompactMetric(value, unit) { return `${Number(value || 0).toLocaleString('zh-CN')} ${unit || ''}`.trim() }
function benchmarkCategoryLabel(value) { return { income: '收入', debt: '负债', asset: '资产', protection: '保障', other: '其他' }[value] || value || '基准' }
function protectionTypeLabel(value) { return { social_insurance: '五险', housing_fund: '公积金', enterprise_annuity: '企业年金', commercial_insurance: '商业保险' }[value] || value || '-' }
function protectionTagType(value) { return { social_insurance: 'info', housing_fund: 'success', enterprise_annuity: 'warning', commercial_insurance: 'default' }[value] || 'default' }
function liabilityTypeLabel(value) { return { mortgage: '房贷', car_loan: '车贷', consumer_loan: '消费贷', other: '其他负债' }[value] || value || '-' }
function analysisStatusLabel(status) { const normalized = String(status || '').toLowerCase(); if (!normalized) return '未运行'; if (normalized === 'success') return '分析完成'; if (normalized === 'running') return '分析中'; if (normalized === 'failed') return '分析失败'; if (normalized === 'skipped') return '未执行'; return status }
function analysisStatusTone(status) { const normalized = String(status || '').toLowerCase(); if (normalized === 'success') return 'success'; if (normalized === 'running') return 'info'; if (normalized === 'failed') return 'error'; return 'muted' }
async function callApp(name, ...args) { const fn = window.go?.main?.App?.[name]; if (!fn) throw new Error(`missing app method: ${name}`); return fn(...args) }
async function loadData() { const profileResult = (await callApp('GetHouseholdProfile')) || {}; profile.value = { ...profile.value, ...profileResult }; summary.value = { ...summary.value, ...((await callApp('GetHouseholdDashboardSummary')) || {}) }; liabilities.value = (await callApp('GetHouseholdLiabilities')) || []; protections.value = (await callApp('GetHouseholdProtections')) || []; benchmarks.value = (await callApp('GetHouseholdBenchmarks', profile.value.region || defaultRegion)) || []; liabilityTrend.value = (await callApp('GetHouseholdLiabilityTrend', 0, 23)) || []; latestRecord.value = await callApp('GetLatestHouseholdAIAnalysis'); await nextTick(); renderLiabilityTrendChart() }
async function handleRefresh() { try { const snapshot = window.go?.main?.App?.SaveHouseholdSnapshot; if (snapshot) await snapshot('asset-overview:manual-refresh'); await loadData(); message.success('已刷新快照') } catch (error) { console.error(error); message.error('刷新失败') } }
async function runAnalysis(triggerSource) { runningAnalysis.value = true; analysisMessage.value = ''; try { const aiConfigs = (await callApp('GetAiConfigs')) || []; const promptTemplates = (await callApp('GetPromptTemplates', '', '模型系统Prompt')) || []; const promptTemplate = promptTemplates.find((item) => item.name === '家庭资产分析-标准模板') || promptTemplates[0]; if (!aiConfigs.length || !promptTemplate) { analysisMessageType.value = 'warning'; analysisMessage.value = '请先配置 AI 源和提示词模板。'; return } await callApp('RunHouseholdAIAnalysis', profile.value.region || defaultRegion, Number(aiConfigs[0]?.ID || 0), Number(promptTemplate?.ID || 0), triggerSource); await loadData(); analysisMessageType.value = 'success'; analysisMessage.value = '已触发数字分析。' } catch (error) { console.error(error); analysisMessageType.value = 'error'; analysisMessage.value = '触发数字分析失败。' } finally { runningAnalysis.value = false } }
function handleResize() { liabilityTrendChartInstance?.resize() }
watch(liabilityTrend, async () => { await nextTick(); renderLiabilityTrendChart() }, { deep: true })
onMounted(async () => { window.addEventListener('resize', handleResize); await loadData() })
onBeforeUnmount(() => { window.removeEventListener('resize', handleResize); liabilityTrendChartInstance?.dispose() })
</script>

<style scoped>
.household-overview { max-width: 1280px; margin: 0 auto; }
.summary-grid { margin-top: 20px; }
.metric-card { position: relative; min-height: 180px; padding: 20px; border-radius: var(--radius-md); border: 1px solid rgba(148, 163, 184, 0.16); background: linear-gradient(180deg, rgba(15, 23, 42, 0.96), rgba(30, 41, 59, 0.9)); overflow: hidden; }
.metric-card::before { content: ''; position: absolute; inset: 0 auto 0 0; width: 4px; background: var(--accent, #38bdf8); }
.metric-topline { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.metric-label { font-size: 13px; color: var(--text-secondary); }
.metric-badge { padding: 2px 8px; border-radius: 999px; background: rgba(99, 102, 241, 0.18); color: #c4b5fd; font-size: 11px; }
.metric-value { margin-top: 18px; font-size: 34px; line-height: 1.1; font-weight: 700; font-family: var(--font-mono); color: #f8fafc; }
.metric-value.profit { color: var(--accent-success); }
.metric-value.loss { color: #fda4af; }
.metric-sub, .metric-detail { margin-top: 10px; color: var(--text-muted); font-size: 12px; line-height: 1.6; }
.metric-detail { color: rgba(226, 232, 240, 0.78); }
.platform-card { background: var(--bg-card); border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 20px; }
.section-row { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; margin-bottom: 18px; }
.section-title { font-size: 16px; font-weight: 700; }
.section-sub { margin-top: 6px; color: var(--text-muted); font-size: 12px; line-height: 1.6; }
.chart-card { min-height: 360px; }
.chart-shell { height: 280px; }
.chart-canvas { width: 100%; height: 100%; }
.benchmark-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; }
.benchmark-item, .analysis-entry-item { padding: 14px; border-radius: var(--radius-sm); background: rgba(15, 23, 42, 0.52); border: 1px solid rgba(148, 163, 184, 0.12); }
.benchmark-topline { display: flex; justify-content: space-between; gap: 8px; color: var(--text-secondary); font-size: 12px; }
.benchmark-name { margin-top: 10px; font-weight: 600; }
.benchmark-value { margin-top: 8px; font-size: 22px; font-family: var(--font-mono); color: #f8fafc; }
.benchmark-sub { margin-top: 8px; color: var(--text-muted); font-size: 12px; }
.analysis-entry-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 12px; margin-top: 16px; }
.analysis-entry-item span, .analysis-entry-item em { display: block; }
.analysis-entry-item span { color: var(--text-secondary); font-size: 12px; }
.analysis-entry-item strong { display: block; margin-top: 10px; font-size: 18px; font-weight: 700; }
.analysis-entry-item em { margin-top: 8px; color: var(--text-muted); font-size: 12px; font-style: normal; }
.analysis-entry-actions { display: flex; justify-content: flex-end; gap: 12px; margin-top: 16px; }
.analysis-status-chip { display: inline-flex; align-items: center; border-radius: 999px; padding: 4px 10px; font-size: 12px; }
.analysis-status-chip.success { background: rgba(16, 185, 129, 0.15); color: #6ee7b7; }
.analysis-status-chip.info { background: rgba(59, 130, 246, 0.15); color: #93c5fd; }
.analysis-status-chip.error { background: rgba(239, 68, 68, 0.15); color: #fca5a5; }
.analysis-status-chip.muted { background: rgba(148, 163, 184, 0.12); color: #cbd5e1; }
.empty-text { color: var(--text-muted); font-size: 13px; line-height: 1.6; padding-top: 24px; }
@media (max-width: 1100px) { .benchmark-grid, .analysis-entry-grid { grid-template-columns: 1fr; } .section-row { flex-direction: column; } }
</style>
