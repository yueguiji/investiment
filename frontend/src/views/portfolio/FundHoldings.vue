<template>
  <div class="fade-in">
    <n-page-header title="基金持仓" subtitle="只查看基金类型持仓" />
    <div class="platform-card" style="margin-top: 20px; padding: 0;">
      <n-data-table :columns="columns" :data="rows" :pagination="{ pageSize: 15 }" striped />
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'

const rows = ref([])

const columns = [
  { title: '代码', key: 'stockCode', width: 110 },
  { title: '名称', key: 'stockName', width: 160 },
  { title: '份额', key: 'quantity', width: 100 },
  { title: '成本净值', key: 'avgCost', width: 120, render: (row) => formatPrice(row.avgCost) },
  { title: '当前净值', key: 'currentPrice', width: 120, render: (row) => formatPrice(row.currentPrice) },
  { title: '总成本', key: 'totalCost', width: 130, render: (row) => formatPrice(row.totalCost) },
  { title: '市值', key: 'totalValue', width: 130, render: (row) => formatPrice(row.totalValue) },
  { title: '浮盈亏', key: 'profitLoss', width: 130, render: (row) => formatSigned(row.profitLoss) },
  { title: '收益率', key: 'profitRate', width: 100, render: (row) => formatSigned(row.profitRate, '%') },
  { title: '账户标签', key: 'accountTag', render: (row) => row.accountTag || '-' }
]

function formatPrice(value) {
  return `¥ ${Number(value || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

function formatSigned(value, suffix = '') {
  const num = Number(value || 0)
  return `${num >= 0 ? '+' : ''}${num.toFixed(2)}${suffix}`
}

async function loadData() {
  if (window.go?.main?.App?.GetHoldingsByType) {
    rows.value = (await window.go.main.App.GetHoldingsByType('fund')) || []
  }
}

onMounted(loadData)
</script>
