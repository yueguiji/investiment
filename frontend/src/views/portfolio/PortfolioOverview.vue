<template>
  <div class="fade-in">
    <n-page-header title="持仓总览" subtitle="实时跟踪你的股票和基金持仓收益">
      <template #extra>
        <n-space>
          <n-button @click="showAddModal = true" type="primary">+ 新增持仓</n-button>
          <n-button @click="$router.push('/portfolio/history')">收益历史</n-button>
        </n-space>
      </template>
    </n-page-header>

    <n-grid :cols="4" :x-gap="16" :y-gap="16" style="margin: 20px 0;">
      <n-gi>
        <div class="platform-card">
          <div style="font-size: 13px; color: var(--text-secondary); margin-bottom: 8px;">总市值</div>
          <div style="font-size: 22px; font-weight: 700;">¥ {{ formatNum(summary.totalValue) }}</div>
          <div style="font-size: 12px; color: var(--text-muted);">成本 ¥{{ formatNum(summary.totalCost) }}</div>
        </div>
      </n-gi>
      <n-gi>
        <div class="platform-card">
          <div style="font-size: 13px; color: var(--text-secondary); margin-bottom: 8px;">总盈亏</div>
          <div style="font-size: 22px; font-weight: 700;" :class="summary.totalProfit >= 0 ? 'profit' : 'loss'">
            {{ summary.totalProfit >= 0 ? '+' : '' }}¥ {{ formatNum(summary.totalProfit) }}
          </div>
          <div style="font-size: 12px;" :class="summary.totalProfitRate >= 0 ? 'profit' : 'loss'">
            {{ summary.totalProfitRate >= 0 ? '+' : '' }}{{ summary.totalProfitRate }}%
          </div>
        </div>
      </n-gi>
      <n-gi>
        <div class="platform-card">
          <div style="font-size: 13px; color: var(--text-secondary); margin-bottom: 8px;">今日盈亏</div>
          <div style="font-size: 22px; font-weight: 700;" :class="summary.todayProfit >= 0 ? 'profit' : 'loss'">
            {{ summary.todayProfit >= 0 ? '+' : '' }}¥ {{ formatNum(summary.todayProfit) }}
          </div>
        </div>
      </n-gi>
      <n-gi>
        <div class="platform-card">
          <div style="font-size: 13px; color: var(--text-secondary); margin-bottom: 8px;">持仓数量</div>
          <div style="font-size: 22px; font-weight: 700;">{{ summary.stockCount + summary.fundCount }}</div>
          <div style="font-size: 12px; color: var(--text-muted);">{{ summary.stockCount }} 股票 / {{ summary.fundCount }} 基金</div>
        </div>
      </n-gi>
    </n-grid>

    <div class="platform-card" style="padding: 0;">
      <n-data-table
        :columns="columns"
        :data="holdings"
        :pagination="{ pageSize: 20 }"
        :row-class-name="rowClassName"
        striped
      />
    </div>

    <n-modal v-model:show="showAddModal" title="新增持仓" preset="card" style="width: 520px;">
      <n-form label-placement="left" label-width="80">
        <n-form-item label="代码">
          <n-input v-model:value="newHolding.stockCode" placeholder="如：sh600519" />
        </n-form-item>
        <n-form-item label="名称">
          <n-input v-model:value="newHolding.stockName" placeholder="如：贵州茅台" />
        </n-form-item>
        <n-form-item label="类型">
          <n-select v-model:value="newHolding.holdingType" :options="[{label:'股票',value:'stock'},{label:'基金',value:'fund'}]" />
        </n-form-item>
        <n-form-item label="持仓数量">
          <n-input-number v-model:value="newHolding.quantity" :min="0" style="width: 100%;" />
        </n-form-item>
        <n-form-item label="持仓成本">
          <n-input-number v-model:value="newHolding.avgCost" :min="0" :precision="3" style="width: 100%;">
            <template #prefix>¥</template>
          </n-input-number>
        </n-form-item>
        <n-form-item label="券商">
          <n-input v-model:value="newHolding.brokerName" placeholder="如：华泰证券" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-button type="primary" @click="handleAddHolding">保存</n-button>
      </template>
    </n-modal>
  </div>
</template>

<script setup>
import { ref, h, onMounted } from 'vue'
import { NButton, NTag } from 'naive-ui'
import { useRouter } from 'vue-router'

const router = useRouter()
const showAddModal = ref(false)
const holdings = ref([])
const summary = ref({ totalCost: 0, totalValue: 0, totalProfit: 0, totalProfitRate: 0, todayProfit: 0, stockCount: 0, fundCount: 0 })
const newHolding = ref({ stockCode: '', stockName: '', holdingType: 'stock', quantity: 0, avgCost: 0, brokerName: '' })

function formatNum(n) {
  return Number(n || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function rowClassName(row) {
  return row.profitLoss >= 0 ? 'row-profit' : 'row-loss'
}

const columns = [
  { title: '代码', key: 'stockCode', width: 100 },
  {
    title: '名称',
    key: 'stockName',
    width: 100,
    render: (row) => h('a', {
      style: { cursor: 'pointer', color: 'var(--primary-light)', textDecoration: 'none' },
      onClick: () => router.push('/invest/monitor')
    }, row.stockName)
  },
  {
    title: '类型',
    key: 'holdingType',
    width: 60,
    render: (row) => h(NTag, { size: 'tiny', round: true, type: row.holdingType === 'stock' ? 'info' : 'warning' },
      () => row.holdingType === 'stock' ? '股票' : '基金')
  },
  { title: '持仓', key: 'quantity', width: 80, render: (row) => row.quantity },
  { title: '成本价', key: 'avgCost', width: 90, render: (row) => h('span', { style: { fontFamily: 'var(--font-mono)' } }, row.avgCost?.toFixed(3)) },
  { title: '现价', key: 'currentPrice', width: 90, render: (row) => h('span', { style: { fontFamily: 'var(--font-mono)' } }, row.currentPrice?.toFixed(3) || '-') },
  {
    title: '盈亏',
    key: 'profitLoss',
    width: 120,
    render: (row) => h('span', {
      style: { fontFamily: 'var(--font-mono)', fontWeight: '600', color: row.profitLoss >= 0 ? 'var(--profit)' : 'var(--loss)' }
    }, `${row.profitLoss >= 0 ? '+' : ''}${formatNum(row.profitLoss)}`)
  },
  {
    title: '收益率',
    key: 'profitRate',
    width: 90,
    render: (row) => h('span', {
      style: { fontFamily: 'var(--font-mono)', fontWeight: '600', color: row.profitRate >= 0 ? 'var(--profit)' : 'var(--loss)' }
    }, `${row.profitRate >= 0 ? '+' : ''}${row.profitRate?.toFixed(2)}%`)
  },
  {
    title: '今日',
    key: 'todayRate',
    width: 80,
    render: (row) => h('span', {
      style: { fontFamily: 'var(--font-mono)', color: row.todayRate >= 0 ? 'var(--profit)' : 'var(--loss)' }
    }, `${row.todayRate >= 0 ? '+' : ''}${row.todayRate?.toFixed(2)}%`)
  },
  { title: '券商', key: 'brokerName', width: 100 },
  {
    title: '操作',
    key: 'actions',
    width: 100,
    render: (row) => h('div', { style: { display: 'flex', gap: '4px' } }, [
      h(NButton, { size: 'tiny', quaternary: true, type: 'error', onClick: () => handleDelete(row.ID) }, () => '删除')
    ])
  }
]

async function loadData() {
  try {
    if (window.go?.main?.App?.GetPortfolioSummary) {
      const s = await window.go.main.App.GetPortfolioSummary()
      if (s) {
        summary.value = s
        holdings.value = s.holdings || []
      }
    }
  } catch (e) {
    console.log('Loading portfolio...', e)
  }
}

async function handleAddHolding() {
  try {
    if (window.go?.main?.App?.CreateHolding) {
      await window.go.main.App.CreateHolding(newHolding.value)
      showAddModal.value = false
      newHolding.value = { stockCode: '', stockName: '', holdingType: 'stock', quantity: 0, avgCost: 0, brokerName: '' }
      await loadData()
    }
  } catch (e) {
    console.error(e)
  }
}

async function handleDelete(id) {
  try {
    if (window.go?.main?.App?.DeleteHolding) {
      await window.go.main.App.DeleteHolding(id)
      await loadData()
    }
  } catch (e) {
    console.error(e)
  }
}

onMounted(loadData)
</script>
