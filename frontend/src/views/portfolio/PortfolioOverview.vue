<template>
  <div class="fade-in portfolio-overview">
    <n-page-header title="我的持仓" subtitle="基金优先的仓位总览，适合支付宝基金账户和稳健仓位管理。">
      <template #extra>
        <n-space>
          <n-button :loading="refreshing" @click="handleRefresh">刷新估值</n-button>
          <n-button @click="handleSaveSnapshot">保存快照</n-button>
          <n-button @click="$router.push('/portfolio/transactions')">交易记录</n-button>
          <n-button type="primary" @click="showAddModal = true">新增持仓</n-button>
        </n-space>
      </template>
    </n-page-header>

    <div class="source-banner">
      页面先读本地缓存，打开速度会更快。需要最新估值时再手动刷新，避免每次进来都把全部基金重抓一遍。    </div>

    <div class="metric-grid">
      <div v-for="card in metricCards" :key="card.label" class="metric-card">
        <div class="metric-label">{{ card.label }}</div>
        <div class="metric-value" :class="card.tone">{{ card.value }}</div>
        <div class="metric-sub">{{ card.sub }}</div>
      </div>
    </div>

    <div class="content-grid">
      <div class="panel-card">
        <div class="panel-head">
          <div>
            <div class="panel-title">基金结构</div>
            <div class="panel-sub">先看保守资产和债基占比，判断账户是不是还在你的舒适区间里。</div>
          </div>
        </div>
        <div class="allocation-list">
          <div v-for="item in dashboard.typeAllocation || []" :key="item.label" class="allocation-row">
            <div class="allocation-meta">
              <span>{{ item.label }}</span>
              <span>{{ item.ratio }}%</span>
            </div>
            <div class="allocation-track">
              <div class="allocation-fill" :style="{ width: `${Math.min(item.ratio || 0, 100)}%` }"></div>
            </div>
            <div class="allocation-note">{{ formatMoney(item.value) }} / {{ item.count }} 只</div>
          </div>
          <div v-if="!(dashboard.typeAllocation || []).length" class="empty-text">还没有基金持仓，先录入一笔当前仓位试试。</div>
        </div>
      </div>

      <div class="panel-card">
        <div class="panel-head">
          <div>
            <div class="panel-title">平台与账户</div>
            <div class="panel-sub">把支付宝主账户、家庭账户和备用账户分开看，会更清楚。</div>
          </div>
        </div>
        <div class="mini-section">
          <div class="mini-title">平台分布</div>
          <div v-for="item in topPlatformAllocation" :key="item.label" class="stack-row">
            <span>{{ item.label }}</span>
            <span>{{ formatMoney(item.value) }} · {{ item.ratio }}%</span>
          </div>
        </div>
        <div class="mini-section">
          <div class="mini-title">账户分组</div>
          <div v-for="item in topAccountAllocation" :key="item.label" class="stack-row">
            <span>{{ item.label }}</span>
            <span>{{ formatMoney(item.value) }} · {{ item.ratio }}%</span>
          </div>
        </div>
      </div>
    </div>

    <div class="panel-card">
      <div class="panel-head">
        <div>
          <div class="panel-title">核心持仓</div>
          <div class="panel-sub">把最近1日、今日估算和当前持仓放在一张表里，方便你快速看仓位。</div>
        </div>
        <n-button text type="primary" @click="$router.push('/portfolio/funds')">查看全部基金</n-button>
      </div>
      <n-data-table
        :columns="columns"
        :data="dashboard.positions || []"
        :pagination="{ pageSize: 8 }"
        :scroll-x="1320"
        striped
      />
    </div>

    <n-modal v-model:show="showAddModal" preset="card" title="快速录入当前仓位" style="width: 620px;">
      <n-form label-placement="top">
        <n-grid :cols="2" :x-gap="12">
          <n-gi>
            <n-form-item label="基金代码">
              <n-input v-model:value="newHolding.stockCode" placeholder="例如 013449" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="基金名称">
              <n-input v-model:value="newHolding.stockName" placeholder="会根据代码自动回填" />
            </n-form-item>
          </n-gi>
        </n-grid>

        <n-grid :cols="2" :x-gap="12">
          <n-gi>
            <n-form-item label="当前持仓金额">
              <n-input-number v-model:value="newHolding.positionAmount" :min="0" :precision="2" style="width: 100%;" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="累计成本金额（可选）">
              <n-input-number v-model:value="newHolding.costAmount" :min="0" :precision="2" style="width: 100%;" />
            </n-form-item>
          </n-gi>
        </n-grid>

        <div class="helper-note">
          保存时会自动抓当前净值并换算份额。你只要告诉系统“这只基金现在有多少钱”，不用自己算份额。        </div>

        <n-grid :cols="2" :x-gap="12">
          <n-gi>
            <n-form-item label="平台">
              <n-input v-model:value="newHolding.brokerName" placeholder="默认支付宝" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="账户分组">
              <n-input v-model:value="newHolding.accountTag" placeholder="例如 主账户 / 家庭账户" />
            </n-form-item>
          </n-gi>
        </n-grid>

        <n-form-item label="备注">
          <n-input v-model:value="newHolding.remark" type="textarea" :rows="3" placeholder="例如 稳健仓 / 备用金 / 养老仓" />
        </n-form-item>
      </n-form>

      <template #action>
        <n-space justify="end">
          <n-button @click="showAddModal = false">取消</n-button>
          <n-button type="primary" :loading="saving" @click="handleAddHolding">保存</n-button>
        </n-space>
      </template>
    </n-modal>

    <FundInsightDrawer
      v-model:show="showDetail"
      :fund="activeFund"
      @refreshed="loadDashboard"
    />
  </div>
</template>

<script setup>
import { computed, h, onMounted, ref, watch } from 'vue'
import { NButton, NTag, useMessage } from 'naive-ui'
import FundInsightDrawer from './components/FundInsightDrawer.vue'

const message = useMessage()

const showAddModal = ref(false)
const showDetail = ref(false)
const saving = ref(false)
const refreshing = ref(false)
const lookupTimer = ref(null)
const activeFund = ref(null)
const dashboard = ref({
  summary: {},
  positions: [],
  typeAllocation: [],
  platformAllocation: [],
  accountAllocation: [],
  companyAllocation: [],
  conservativeRatio: 0,
  bondAllocationRatio: 0,
  estimatedProfitToday: 0
})

const newHolding = ref({
  stockCode: '',
  stockName: '',
  positionAmount: 0,
  costAmount: 0,
  brokerName: '支付宝',
  accountTag: '主账户',
  remark: ''
})

const metricCards = computed(() => {
  const summary = dashboard.value.summary || {}
  const totalValue = Number(summary.fundValue || summary.totalValue || 0)
  const totalProfit = Number(summary.totalProfit || 0)
  const todayProfit = Number(dashboard.value.estimatedProfitToday || summary.todayProfit || 0)
  const conservativeRatio = Number(dashboard.value.conservativeRatio || 0)
  const bondRatio = Number(dashboard.value.bondAllocationRatio || 0)
  const fundCount = Number(summary.fundCount || 0)

  return [
    {
      label: '基金总市值',
      value: formatMoney(totalValue),
      sub: `总成本 ${formatMoney(summary.totalCost || 0)}`,
      tone: ''
    },
    {
      label: '累计收益',
      value: formatSignedMoney(totalProfit),
      sub: `${signedPercent(summary.totalProfitRate)}%`,
      tone: totalProfit >= 0 ? 'profit' : 'loss'
    },
    {
      label: '今日估算',
      value: formatSignedMoney(todayProfit),
      sub: '仅在盘中估算更新时生效',
      tone: todayProfit >= 0 ? 'profit' : 'loss'
    },
    {
      label: '保守资产占比',
      value: `${conservativeRatio.toFixed(2)}%`,
      sub: '债基 + 现金管理',
      tone: ''
    },
    {
      label: '债基占比',
      value: `${bondRatio.toFixed(2)}%`,
      sub: `${summary.bondFundCount || 0} 只债基`,
      tone: ''
    },
    {
      label: '基金数量',
      value: `${fundCount}`,
      sub: `平台 ${topPlatformAllocation.value.length} 个 / 账户 ${topAccountAllocation.value.length} 个`,
      tone: ''
    }
  ]
})

const topPlatformAllocation = computed(() => (dashboard.value.platformAllocation || []).slice(0, 4))
const topAccountAllocation = computed(() => (dashboard.value.accountAllocation || []).slice(0, 4))

const columns = [
  {
    title: '基金',
    key: 'stockName',
    width: 200,
    fixed: 'left',
    render: (row) => h('div', { class: 'fund-cell' }, [
      h(NButton, {
        text: true,
        type: 'primary',
        class: 'fund-link',
        onClick: () => openFundDetail(row)
      }, () => row.stockName || row.stockCode),
      h('div', { class: 'fund-code' }, row.stockCode)
    ])
  },
  {
    title: '分类',
    key: 'categoryLabel',
    width: 110,
    render: (row) => h(NTag, {
      size: 'small',
      round: true,
      bordered: false,
      type: row.category === 'bond' ? 'success' : row.category === 'cash' ? 'warning' : row.category === 'equity' ? 'info' : 'default'
    }, () => row.categoryLabel || '其他基金')
  },
  {
    title: '平台 / 账户',
    key: 'brokerName',
    width: 150,
    render: (row) => `${row.brokerName || '未标记平台'} / ${row.accountTag || '未分组账户'}`
  },
  {
    title: '份额',
    key: 'quantity',
    width: 110,
    render: (row) => formatUnits(row.quantity)
  },
  {
    title: '最近1日',
    key: 'latestDailyRate',
    title: '最近1日',
    width: 100,
    render: (row) => {
      const text = latestDailyRateText(row)
      if (text === '--') {
        return h('span', { class: 'fund-code' }, text)
      }
      return h('span', {
        class: latestDailyRateValue(row) >= 0 ? 'profit-text' : 'loss-text'
      }, text)
    }
  },
  {
    title: '更新时间',
    key: 'latestDailyUpdatedAt',
    width: 132,
    render: (row) => h('span', { class: 'fund-code' }, latestDailyTimeText(row))
  },
  {
    title: '今日估算',
    key: 'todayChange',
    width: 140,
    render: (row) => {
      if (!row.estimateUpdated) {
        return h('span', { class: 'fund-code' }, '休市或暂无盘中估算')
      }
      return h('span', { class: row.todayChange >= 0 ? 'profit-text' : 'loss-text' }, `${formatSignedMoney(row.todayChange)} / ${signedPercent(row.todayRate)}%`)
    }
  },
  {
    title: '持仓市值',
    key: 'totalValue',
    width: 120,
    render: (row) => formatMoney(row.totalValue)
  },
  {
    title: '近1月',
    key: 'netGrowth1',
    width: 90,
    render: (row) => formatMaybePercent(row.netGrowth1)
  },
  {
    title: '近3月',
    key: 'netGrowth3',
    width: 90,
    render: (row) => formatMaybePercent(row.netGrowth3)
  },
  {
    title: '近1年',
    key: 'netGrowth12',
    width: 90,
    render: (row) => formatMaybePercent(row.netGrowth12)
  },
  {
    title: '操作',
    key: 'actions',
    width: 96,
    render: (row) => h('div', { class: 'table-actions' }, [
      h(NButton, { text: true, onClick: () => openFundPage(row.stockCode) }, () => '外链')
    ])
  }
]

function formatMoney(value) {
  return `¥${Number(value || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

function formatSignedMoney(value) {
  const amount = Number(value || 0)
  return `${amount >= 0 ? '+' : '-'}${formatMoney(Math.abs(amount))}`
}

function formatPrice(value, digits = 4) {
  return Number(value || 0).toFixed(digits)
}

function formatUnits(value) {
  return Number(value || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function formatMaybePercent(value) {
  if (value === null || value === undefined) return '-'
  return `${signedPercent(value)}%`
}

function signedPercent(value) {
  const amount = Number(value || 0)
  return `${amount >= 0 ? '+' : ''}${amount.toFixed(2)}`
}

function shortDateTime(value) {
  if (!value) return '-'
  return value.replace(/^(\d{4})-/, '').slice(0, 11)
}

function formatSignedPercentText(value) {
  if (value === null || value === undefined) return '--'
  const amount = Number(value || 0)
  return `${amount >= 0 ? '+' : ''}${amount.toFixed(2)}%`
}

function latestMetaText(row) {
  const timeText = row.estimateUpdated
    ? shortDateTime(row.netEstimatedTime)
    : (row.netUnitValueDate || '-')
  const ratio = row.netEstimatedRate ?? (row.estimateUpdated ? row.todayRate : null)
  return `${timeText} · ${formatSignedPercentText(ratio)}`
}

function latestDailyRateValue(row) {
  if (row.latestDailyRate !== null && row.latestDailyRate !== undefined) {
    return Number(row.latestDailyRate || 0)
  }
  return 0
}

function latestDailyRateText(row) {
  const hasLatest = row.latestDailyRate !== null && row.latestDailyRate !== undefined
  if (!hasLatest) return '--'
  return `${signedPercent(latestDailyRateValue(row))}%`
}

function latestDailyTimeText(row) {
  const raw = row.latestDailyUpdatedAt || row.netUnitValueDate
  return shortDateTime(raw)
}

function openFundPage(code) {
  if (!code || !window.go?.main?.App?.OpenURL) return
  window.go.main.App.OpenURL(`https://fund.eastmoney.com/${code}.html`)
}

function openFundDetail(row) {
  activeFund.value = row
  showDetail.value = true
}

async function autofillFundName(code, target) {
  const trimmed = String(code || '').trim()
  if (!trimmed || trimmed.length < 4 || !window.go?.main?.App?.GetfundList) return
  const result = await window.go.main.App.GetfundList(trimmed)
  const exact = (result || []).find((item) => item.code === trimmed)
  if (exact?.name) {
    target.value.stockName = exact.name
  }
}

async function loadDashboard() {
  if (!window.go?.main?.App?.GetFundPortfolioDashboard) return
  const result = await window.go.main.App.GetFundPortfolioDashboard()
  if (result) {
    dashboard.value = result
  }
}

async function handleRefresh() {
  refreshing.value = true
  try {
    if (window.go?.main?.App?.SyncPortfolioQuotes) {
      await window.go.main.App.SyncPortfolioQuotes()
    }
    await loadDashboard()
    message.success('基金估值已刷新')
  } catch (error) {
    console.error(error)
    message.error('刷新失败')
  } finally {
    refreshing.value = false
  }
}

async function handleSaveSnapshot() {
  try {
    if (window.go?.main?.App?.SavePortfolioSnapshot) {
      await window.go.main.App.SavePortfolioSnapshot()
      message.success('当日快照已保存')
    }
  } catch (error) {
    console.error(error)
      message.error('快照保存失败')
  }
}

async function handleAddHolding() {
  if (!newHolding.value.stockCode.trim() || !newHolding.value.positionAmount) {
    message.warning('请先填写基金代码和当前持仓金额')
    return
  }

  saving.value = true
  try {
    const result = await window.go.main.App.UpsertFundHoldingByAmount({ ...newHolding.value })
    if (!result) {
      message.error('暂时没有拿到这只基金的净值，稍后再试')
      return
    }
    showAddModal.value = false
    newHolding.value = {
      stockCode: '',
      stockName: '',
      positionAmount: 0,
      costAmount: 0,
      brokerName: '支付宝',
      accountTag: '主账户',
      remark: ''
    }
    await loadDashboard()
    message.success('持仓已保存')
  } catch (error) {
    console.error(error)
    message.error('持仓保存失败')
  } finally {
    saving.value = false
  }
}

watch(() => newHolding.value.stockCode, (value) => {
  if (lookupTimer.value) {
    clearTimeout(lookupTimer.value)
  }
  lookupTimer.value = setTimeout(() => {
    autofillFundName(value, newHolding)
  }, 220)
})

onMounted(loadDashboard)
</script>

<style scoped>
.portfolio-overview {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 16px;
}

.metric-card,
.panel-card {
  border-radius: var(--radius-lg);
  background: linear-gradient(180deg, rgba(14, 24, 39, 0.96), rgba(18, 30, 48, 0.98));
  border: 1px solid rgba(97, 118, 148, 0.26);
  box-shadow: 0 18px 40px rgba(7, 12, 18, 0.18);
}

.source-banner {
  padding: 12px 14px;
  border-radius: var(--radius-md);
  background: rgba(56, 189, 248, 0.08);
  border: 1px solid rgba(56, 189, 248, 0.18);
  color: var(--text-secondary);
  font-size: 12px;
  line-height: 1.6;
}

.metric-card {
  padding: 18px;
}

.metric-label,
.panel-sub,
.allocation-note,
.empty-text,
.helper-note,
.mini-title,
.fund-code {
  color: var(--text-secondary);
}

.helper-note {
  margin-bottom: 12px;
  font-size: 12px;
  line-height: 1.6;
}

.metric-value {
  margin-top: 10px;
  font-size: 28px;
  font-weight: 700;
  letter-spacing: -0.03em;
}

.metric-sub {
  margin-top: 8px;
  font-size: 12px;
}

.profit {
  color: var(--profit);
}

.loss {
  color: var(--loss);
}

.content-grid {
  display: grid;
  grid-template-columns: 1.2fr 1fr;
  gap: 16px;
}

.panel-card {
  padding: 18px;
}

.panel-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.panel-title {
  font-size: 16px;
  font-weight: 700;
}

.allocation-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.allocation-row {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.allocation-meta,
.stack-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.allocation-track {
  height: 8px;
  border-radius: 999px;
  background: rgba(124, 146, 177, 0.18);
  overflow: hidden;
}

.allocation-fill {
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, #4ade80, #38bdf8);
}

.mini-section + .mini-section {
  margin-top: 18px;
  padding-top: 18px;
  border-top: 1px solid rgba(97, 118, 148, 0.18);
}

.mini-title {
  margin-bottom: 10px;
  font-size: 13px;
}

.stack-row + .stack-row {
  margin-top: 10px;
}

.fund-cell,
.value-stack {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

:deep(.table-actions) {
  display: flex;
  gap: 16px;
  align-items: center;
  white-space: nowrap;
}

:deep(.table-actions .n-button) {
  margin-right: 4px;
}

:deep(.table-actions .n-button:last-child) {
  margin-right: 0;
}

.profit-text {
  color: var(--profit);
  font-weight: 600;
}

.loss-text {
  color: var(--loss);
  font-weight: 600;
}

@media (max-width: 960px) {
  .content-grid {
    grid-template-columns: 1fr;
  }
}
</style>
