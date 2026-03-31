<template>
  <div class="fade-in">
    <n-page-header title="收益历史" subtitle="查看每日快照的总收益变化" />

    <div class="platform-card" style="margin-top: 20px;">
      <div class="chart-header">
        <div>
          <div class="metric-label">近 {{ history.length }} 天收益曲线</div>
          <div class="metric-sub">数据来自 `profit_snapshots` 每日快照</div>
        </div>
        <n-button size="small" @click="loadData">刷新</n-button>
      </div>

      <div v-if="history.length > 1" class="chart-shell">
        <svg viewBox="0 0 100 32" preserveAspectRatio="none" class="trend-chart">
          <polyline :points="polylinePoints" fill="none" stroke="var(--primary-light)" stroke-width="1.5" />
        </svg>
      </div>
      <div v-else class="empty-text">快照数量不足，先积累几天数据后这里会显示趋势。</div>
    </div>

    <div class="platform-card" style="margin-top: 20px; padding: 0;">
      <n-data-table :columns="columns" :data="history" :pagination="{ pageSize: 15 }" />
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'

const history = ref([])

const columns = [
  { title: '日期', key: 'snapshotDate', width: 120, render: (row) => formatDate(row.snapshotDate) },
  { title: '总成本', key: 'totalCost', width: 140, render: (row) => formatPrice(row.totalCost) },
  { title: '总市值', key: 'totalValue', width: 140, render: (row) => formatPrice(row.totalValue) },
  { title: '总收益', key: 'totalProfit', width: 140, render: (row) => formatSigned(row.totalProfit) },
  { title: '收益率', key: 'profitRate', width: 100, render: (row) => formatSigned(row.profitRate, '%') },
  { title: '股票市值', key: 'stockValue', width: 140, render: (row) => formatPrice(row.stockValue) },
  { title: '基金市值', key: 'fundValue', width: 140, render: (row) => formatPrice(row.fundValue) }
]

const polylinePoints = computed(() => {
  if (history.value.length <= 1) {
    return ''
  }
  const profits = history.value.map((item) => Number(item.totalProfit || 0))
  const min = Math.min(...profits)
  const max = Math.max(...profits)
  const span = max - min || 1
  return history.value.map((item, index) => {
    const x = (index / (history.value.length - 1)) * 100
    const y = 28 - ((Number(item.totalProfit || 0) - min) / span) * 24
    return `${x},${y}`
  }).join(' ')
})

function formatPrice(value) {
  return `¥ ${Number(value || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

function formatSigned(value, suffix = '') {
  const num = Number(value || 0)
  return `${num >= 0 ? '+' : ''}${num.toFixed(2)}${suffix}`
}

function formatDate(value) {
  if (!value) return '-'
  return new Date(value).toLocaleDateString('zh-CN')
}

async function loadData() {
  if (window.go?.main?.App?.GetProfitHistory) {
    history.value = (await window.go.main.App.GetProfitHistory(90)) || []
  }
}

onMounted(loadData)
</script>

<style scoped>
.chart-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.metric-label {
  font-size: 13px;
  color: var(--text-secondary);
}

.metric-sub {
  margin-top: 4px;
  color: var(--text-muted);
  font-size: 12px;
}

.chart-shell {
  height: 220px;
  border-radius: var(--radius-md);
  background: linear-gradient(180deg, rgba(99, 102, 241, 0.08), rgba(99, 102, 241, 0.02));
  border: 1px solid var(--border-color);
  padding: 12px;
}

.trend-chart {
  width: 100%;
  height: 100%;
}

.empty-text {
  color: var(--text-secondary);
  padding: 20px 0;
}
</style>
