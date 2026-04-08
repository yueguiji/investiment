<template>
  <div class="fade-in history-page">
    <n-page-header title="收益历史" subtitle="按日快照看你的基金账户变化，适合观察稳健仓的回撤和恢复速度">
      <template #extra>
        <n-space>
          <n-select v-model:value="days" :options="dayOptions" style="width: 120px;" />
          <n-button @click="loadData">刷新</n-button>
          <n-button type="primary" @click="handleSaveSnapshot">立即保存快照</n-button>
        </n-space>
      </template>
    </n-page-header>

    <div class="metric-grid">
      <div class="metric-card">
        <div class="metric-label">当前累计收益</div>
        <div class="metric-value" :class="latestProfit >= 0 ? 'profit' : 'loss'">{{ formatSignedMoney(latestProfit) }}</div>
        <div class="metric-sub">{{ signedPercent(latestProfitRate) }}%</div>
      </div>
      <div class="metric-card">
        <div class="metric-label">最新基金市值</div>
        <div class="metric-value">{{ formatMoney(latestFundValue) }}</div>
        <div class="metric-sub">股票市值 {{ formatMoney(latestStockValue) }}</div>
      </div>
      <div class="metric-card">
        <div class="metric-label">近 7 日变化</div>
        <div class="metric-value" :class="recentDelta >= 0 ? 'profit' : 'loss'">{{ formatSignedMoney(recentDelta) }}</div>
        <div class="metric-sub">按总收益口径计算</div>
      </div>
      <div class="metric-card">
        <div class="metric-label">区间最大回撤</div>
        <div class="metric-value loss">{{ maxDrawdown.toFixed(2) }}%</div>
        <div class="metric-sub">越低越稳</div>
      </div>
    </div>

    <div class="chart-card">
      <div class="chart-head">
        <div>
          <div class="chart-title">收益曲线</div>
          <div class="chart-sub">蓝线是总收益，绿线是基金市值。债基账户更适合看趋势是否平滑。</div>
        </div>
      </div>
      <div v-if="history.length > 1" class="chart-shell">
        <svg viewBox="0 0 100 34" preserveAspectRatio="none" class="trend-chart">
          <polyline :points="profitLinePoints" fill="none" stroke="#38bdf8" stroke-width="1.8" />
          <polyline :points="fundValueLinePoints" fill="none" stroke="#4ade80" stroke-width="1.6" opacity="0.9" />
        </svg>
      </div>
      <div v-else class="empty-text">当前快照还不够，连续使用几天后这里会更有参考价值。</div>
    </div>

    <div class="table-card">
      <n-data-table
        :columns="columns"
        :data="history"
        :pagination="{ pageSize: 15 }"
        :scroll-x="960"
        striped
      />
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useMessage } from 'naive-ui'

const message = useMessage()
const days = ref(90)
const history = ref([])

const dayOptions = [
  { label: '近 30 天', value: 30 },
  { label: '近 90 天', value: 90 },
  { label: '近 180 天', value: 180 },
  { label: '近 365 天', value: 365 }
]

const latest = computed(() => history.value[history.value.length - 1] || {})
const latestProfit = computed(() => Number(latest.value.totalProfit || 0))
const latestProfitRate = computed(() => Number(latest.value.profitRate || 0))
const latestFundValue = computed(() => Number(latest.value.fundValue || 0))
const latestStockValue = computed(() => Number(latest.value.stockValue || 0))

const recentDelta = computed(() => {
  if (!history.value.length) return 0
  const start = history.value[Math.max(0, history.value.length - 7)]
  return Number(latest.value.totalProfit || 0) - Number(start?.totalProfit || 0)
})

const maxDrawdown = computed(() => {
  let peak = Number.NEGATIVE_INFINITY
  let drawdown = 0
  for (const item of history.value) {
    const value = Number(item.totalValue || 0)
    peak = Math.max(peak, value)
    if (peak > 0) {
      drawdown = Math.min(drawdown, (value - peak) / peak)
    }
  }
  return Math.abs(drawdown * 100)
})

const profitLinePoints = computed(() => buildLine(history.value.map((item) => Number(item.totalProfit || 0))))
const fundValueLinePoints = computed(() => buildLine(history.value.map((item) => Number(item.fundValue || 0))))

const columns = [
  {
    title: '日期',
    key: 'snapshotDate',
    width: 130,
    render: (row) => formatDate(row.snapshotDate)
  },
  {
    title: '总成本',
    key: 'totalCost',
    width: 130,
    render: (row) => formatMoney(row.totalCost)
  },
  {
    title: '总市值',
    key: 'totalValue',
    width: 130,
    render: (row) => formatMoney(row.totalValue)
  },
  {
    title: '累计收益',
    key: 'totalProfit',
    width: 130,
    render: (row) => formatSignedMoney(row.totalProfit)
  },
  {
    title: '收益率',
    key: 'profitRate',
    width: 100,
    render: (row) => `${signedPercent(row.profitRate)}%`
  },
  {
    title: '基金市值',
    key: 'fundValue',
    width: 130,
    render: (row) => formatMoney(row.fundValue)
  },
  {
    title: '股票市值',
    key: 'stockValue',
    width: 130,
    render: (row) => formatMoney(row.stockValue)
  },
  {
    title: '基金占比',
    key: 'fundRatio',
    width: 100,
    render: (row) => {
      const totalValue = Number(row.totalValue || 0)
      const fundValue = Number(row.fundValue || 0)
      if (!totalValue) return '0.00%'
      return `${((fundValue / totalValue) * 100).toFixed(2)}%`
    }
  }
]

function buildLine(values) {
  if (values.length <= 1) return ''
  const min = Math.min(...values)
  const max = Math.max(...values)
  const span = max - min || 1
  return values.map((value, index) => {
    const x = (index / (values.length - 1)) * 100
    const y = 30 - ((value - min) / span) * 26
    return `${x},${y}`
  }).join(' ')
}

function formatMoney(value) {
  return `¥${Number(value || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

function formatSignedMoney(value) {
  const amount = Number(value || 0)
  return `${amount >= 0 ? '+' : '-'}${formatMoney(Math.abs(amount))}`
}

function signedPercent(value) {
  const amount = Number(value || 0)
  return `${amount >= 0 ? '+' : ''}${amount.toFixed(2)}`
}

function formatDate(value) {
  if (!value) return '-'
  return new Date(value).toLocaleDateString('zh-CN')
}

async function loadData() {
  if (!window.go?.main?.App?.GetProfitHistory) return
  history.value = (await window.go.main.App.GetProfitHistory(days.value)) || []
}

async function handleSaveSnapshot() {
  try {
    if (window.go?.main?.App?.SavePortfolioSnapshot) {
      await window.go.main.App.SavePortfolioSnapshot()
      await loadData()
      message.success('快照已保存')
    }
  } catch (error) {
    console.error(error)
    message.error('保存快照失败')
  }
}

watch(days, () => {
  loadData()
})

onMounted(loadData)
</script>

<style scoped>
.history-page {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 16px;
}

.metric-card,
.chart-card,
.table-card {
  border-radius: var(--radius-lg);
  border: 1px solid rgba(97, 118, 148, 0.24);
  background: linear-gradient(180deg, rgba(14, 24, 39, 0.96), rgba(18, 30, 48, 0.98));
  box-shadow: 0 18px 40px rgba(7, 12, 18, 0.16);
}

.metric-card,
.chart-card,
.table-card {
  padding: 18px;
}

.metric-label,
.metric-sub,
.chart-sub,
.empty-text {
  color: var(--text-secondary);
}

.metric-value {
  margin-top: 10px;
  font-size: 28px;
  font-weight: 700;
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

.chart-head {
  margin-bottom: 14px;
}

.chart-title {
  font-size: 16px;
  font-weight: 700;
}

.chart-shell {
  height: 240px;
  padding: 12px;
  border-radius: var(--radius-md);
  background: rgba(11, 19, 31, 0.32);
  border: 1px solid rgba(97, 118, 148, 0.18);
}

.trend-chart {
  width: 100%;
  height: 100%;
}
</style>
