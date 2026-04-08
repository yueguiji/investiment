<template>
  <div class="fade-in fund-holdings-page">
    <n-page-header title="基金持仓" subtitle="优先读取本地持仓，再按需要刷新估值；支持直接编辑当前持仓金额。">
      <template #extra>
        <n-space>
          <n-button @click="showCollectionAI = true">AI 分析组合</n-button>
          <n-button :loading="refreshing" @click="handleRefresh">刷新估值</n-button>
          <n-button @click="$router.push('/portfolio/transactions')">录入持仓</n-button>
        </n-space>
      </template>
    </n-page-header>

    <div class="source-banner">
      页面默认只读取本地缓存，所以打开会更快。需要最新估值时，再手动刷新。    </div>

    <div class="toolbar-card">
      <n-grid :cols="4" :x-gap="12" :y-gap="12">
        <n-gi>
          <n-input v-model:value="keyword" clearable placeholder="搜索基金代码或名称" />
        </n-gi>
        <n-gi>
          <n-select v-model:value="categoryFilter" :options="categoryOptions" />
        </n-gi>
        <n-gi>
          <n-select v-model:value="platformFilter" :options="platformOptions" />
        </n-gi>
        <n-gi>
          <n-select v-model:value="accountFilter" :options="accountOptions" />
        </n-gi>
      </n-grid>
    </div>

    <div class="metric-grid">
      <div class="mini-metric">
        <div class="mini-label">基金总市值</div>
        <div class="mini-value">{{ formatMoney(summary.fundValue) }}</div>
      </div>
      <div class="mini-metric">
        <div class="mini-label">债基市值</div>
        <div class="mini-value">{{ formatMoney(summary.bondFundValue) }}</div>
        <div class="mini-sub">{{ summary.bondFundCount || 0 }} 只</div>
      </div>
      <div class="mini-metric">
        <div class="mini-label">现金管理</div>
        <div class="mini-value">{{ formatMoney(summary.cashFundValue) }}</div>
        <div class="mini-sub">{{ summary.cashFundCount || 0 }} 只</div>
      </div>
      <div class="mini-metric">
        <div class="mini-label">保守资产占比</div>
        <div class="mini-value">{{ Number(dashboard.conservativeRatio || 0).toFixed(2) }}%</div>
        <div class="mini-sub">债基 + 现金管理</div>
      </div>
    </div>

    <div class="table-card">
      <n-data-table
        :columns="columns"
        :data="filteredRows"
        :pagination="{ pageSize: 12 }"
        :scroll-x="1780"
        striped
      />
    </div>

    <n-modal v-model:show="showEditModal" preset="card" title="编辑持仓" style="width: 620px;">
      <n-form label-placement="top">
        <n-grid :cols="2" :x-gap="12">
          <n-gi>
            <n-form-item label="基金代码">
              <n-input v-model:value="editForm.stockCode" disabled />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="基金名称">
              <n-input v-model:value="editForm.stockName" disabled />
            </n-form-item>
          </n-gi>
        </n-grid>

        <n-grid :cols="2" :x-gap="12">
          <n-gi>
            <n-form-item label="当前持仓金额">
              <n-input-number v-model:value="editForm.positionAmount" :min="0" :precision="2" style="width: 100%;" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="累计成本金额（可选）">
              <n-input-number v-model:value="editForm.costAmount" :min="0" :precision="2" style="width: 100%;" />
            </n-form-item>
          </n-gi>
        </n-grid>

        <n-grid :cols="2" :x-gap="12">
          <n-gi>
            <n-form-item label="平台">
              <n-input v-model:value="editForm.brokerName" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="账户分组">
              <n-input v-model:value="editForm.accountTag" />
            </n-form-item>
          </n-gi>
        </n-grid>

        <n-form-item label="备注">
          <n-input v-model:value="editForm.remark" type="textarea" :rows="3" placeholder="例如 稳健仓 / 备用金 / 养老仓" />
        </n-form-item>
      </n-form>

      <template #action>
        <n-space justify="end">
          <n-button @click="showEditModal = false">取消</n-button>
          <n-button type="primary" :loading="savingEdit" @click="handleSaveEdit">保存</n-button>
        </n-space>
      </template>
    </n-modal>

    <FundInsightDrawer
      v-model:show="showDetail"
      :fund="activeFund"
      @refreshed="loadData"
    />

    <FundAIAnalysisModal
      v-model:show="showSingleAI"
      mode="single"
      :fund-code="activeFund?.stockCode || ''"
      :title="activeFund ? `${activeFund.stockName || activeFund.stockCode} AI 分析` : '基金 AI 分析'"
    />

    <FundAIAnalysisModal
      v-model:show="showCollectionAI"
      mode="collection"
      scope="holdings"
      title="持仓基金 AI 分析"
    />
  </div>
</template>

<script setup>
import { computed, h, onMounted, ref } from 'vue'
import { NButton, NTag, useMessage } from 'naive-ui'
import FundInsightDrawer from './components/FundInsightDrawer.vue'
import FundAIAnalysisModal from './components/FundAIAnalysisModal.vue'

const message = useMessage()
const refreshing = ref(false)
const savingEdit = ref(false)
const showDetail = ref(false)
const showSingleAI = ref(false)
const showCollectionAI = ref(false)
const showEditModal = ref(false)
const keyword = ref('')
const categoryFilter = ref('all')
const platformFilter = ref('all')
const accountFilter = ref('all')
const activeFund = ref(null)
const editingHoldingId = ref(0)
const dashboard = ref({
  summary: {},
  positions: [],
  platformAllocation: [],
  accountAllocation: [],
  conservativeRatio: 0
})

const editForm = ref(createEmptyEditForm())

const summary = computed(() => dashboard.value.summary || {})

const categoryOptions = [
  { label: '全部分类', value: 'all' },
  { label: '债券基金', value: 'bond' },
  { label: '现金管理', value: 'cash' },
  { label: '权益基金', value: 'equity' },
  { label: '其他基金', value: 'other' }
]

const platformOptions = computed(() => [
  { label: '全部平台', value: 'all' },
  ...(dashboard.value.platformAllocation || []).map((item) => ({ label: item.label, value: item.label }))
])

const accountOptions = computed(() => [
  { label: '全部账户', value: 'all' },
  ...(dashboard.value.accountAllocation || []).map((item) => ({ label: item.label, value: item.label }))
])

const filteredRows = computed(() => {
  const list = dashboard.value.positions || []
  const search = keyword.value.trim().toLowerCase()
  return list.filter((row) => {
    const matchesKeyword = !search || [row.stockCode, row.stockName]
      .filter(Boolean)
      .some((value) => String(value).toLowerCase().includes(search))
    const matchesCategory = categoryFilter.value === 'all' || row.category === categoryFilter.value
    const matchesPlatform = platformFilter.value === 'all' || (row.brokerName || '未标记平台') === platformFilter.value
    const matchesAccount = accountFilter.value === 'all' || (row.accountTag || '未分组账户') === accountFilter.value
    return matchesKeyword && matchesCategory && matchesPlatform && matchesAccount
  })
})

const columns = [
  {
    title: '基金',
    key: 'stockName',
    width: 220,
    fixed: 'left',
    render: (row) => h('div', { class: 'fund-cell' }, [
      h(NButton, {
        text: true,
        type: 'primary',
        class: 'fund-link',
        onClick: () => openFundDetail(row)
      }, () => row.stockName || row.stockCode),
      h('div', { class: 'fund-meta' }, row.stockCode)
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
    key: 'accountTag',
    width: 160,
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
    width: 100,
    render: (row) => {
      const text = latestDailyRateText(row)
      if (text === '--') {
        return h('span', { class: 'fund-meta' }, text)
      }
      return h('span', {
        class: latestDailyRateValue(row) >= 0 ? 'profit-text' : 'loss-text'
      }, text)
    }
  },
  {
    title: '更新时间',
    key: 'latestDailyUpdatedAt',
    width: 140,
    render: (row) => h('span', { class: 'fund-meta' }, latestDailyTimeText(row))
  },
  {
    title: '今日估算',
    key: 'todayRate',
    width: 140,
    render: (row) => {
      if (!row.estimateUpdated) {
        return h('span', { class: 'fund-meta' }, '休市或暂无盘中估算')
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
    title: '近6月',
    key: 'netGrowth6',
    width: 90,
    render: (row) => formatMaybePercent(row.netGrowth6)
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
    width: 240,
    render: (row) => h('div', { class: 'table-actions' }, [
      h(NButton, { size: 'tiny', secondary: true, type: 'warning', onClick: () => openSingleAI(row) }, () => 'AI分析'),
      h(NButton, { size: 'tiny', secondary: true, type: 'info', onClick: () => openEditModal(row) }, () => '编辑'),
      h(NButton, { size: 'tiny', secondary: true, type: 'error', onClick: () => handleDelete(row) }, () => '删除'),
      h(NButton, { size: 'tiny', secondary: true, type: 'success', onClick: () => openFundPage(row.stockCode) }, () => '外链')
    ])
  }
]

function createEmptyEditForm() {
  return {
    stockCode: '',
    stockName: '',
    positionAmount: 0,
    costAmount: 0,
    brokerName: '支付宝',
    accountTag: '主账户',
    remark: ''
  }
}

function formatMoney(value) {
  return `¥${Number(value || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

function formatSignedMoney(value) {
  const amount = Number(value || 0)
  return `${amount >= 0 ? '+' : '-'}${formatMoney(Math.abs(amount))}`
}

function formatUnits(value) {
  return Number(value || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function signedPercent(value) {
  const amount = Number(value || 0)
  return `${amount >= 0 ? '+' : ''}${amount.toFixed(2)}`
}

function formatMaybePercent(value) {
  if (value === null || value === undefined) return '-'
  return `${signedPercent(value)}%`
}

function shortDateTime(value) {
  if (!value) return '-'
  return String(value).replace(/^(\d{4})-/, '').slice(0, 11)
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

function openFundDetail(row) {
  activeFund.value = row
  showDetail.value = true
}

function openSingleAI(row) {
  activeFund.value = row
  showSingleAI.value = true
}

function openEditModal(row) {
  editingHoldingId.value = Number(row.ID || 0)
  editForm.value = {
    stockCode: row.stockCode || '',
    stockName: row.stockName || '',
    positionAmount: Number(row.totalValue || 0),
    costAmount: Number(row.totalCost || 0),
    brokerName: row.brokerName || '支付宝',
    accountTag: row.accountTag || '主账户',
    remark: row.remark || ''
  }
  showEditModal.value = true
}

async function handleSaveEdit() {
  if (!editForm.value.stockCode.trim() || !editForm.value.positionAmount) {
    message.warning('请先填写当前持仓金额')
    return
  }

  savingEdit.value = true
  try {
    const result = await window.go.main.App.UpsertFundHoldingByAmount({ ...editForm.value })
    if (!result) {
      message.error('保存失败，请稍后重试')
      return
    }
    showEditModal.value = false
    editingHoldingId.value = 0
    editForm.value = createEmptyEditForm()
    await loadData()
    message.success('持仓已更新')
  } catch (error) {
    console.error(error)
    message.error('更新持仓失败')
  } finally {
    savingEdit.value = false
  }
}

async function handleDelete(row) {
  const id = Number(row.ID || 0)
  if (!id) {
    message.error('这条持仓缺少有效 ID，暂时无法删除')
    return
  }
  if (!window.confirm(`确认删除 ${row.stockName || row.stockCode} 这条持仓吗？`)) {
    return
  }

  try {
    const ok = await window.go.main.App.DeleteHolding(id)
    if (!ok) {
      message.error('删除失败')
      return
    }
    if (editingHoldingId.value === id) {
      showEditModal.value = false
      editingHoldingId.value = 0
      editForm.value = createEmptyEditForm()
    }
    await loadData()
    message.success('持仓已删除')
  } catch (error) {
    console.error(error)
    message.error('删除持仓失败')
  }
}

async function loadData() {
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
    await loadData()
    message.success('基金估值已刷新')
  } catch (error) {
    console.error(error)
    message.error('刷新失败')
  } finally {
    refreshing.value = false
  }
}

onMounted(loadData)
</script>

<style scoped>
.fund-holdings-page {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.toolbar-card,
.metric-grid,
.table-card {
  border-radius: var(--radius-lg);
  border: 1px solid rgba(97, 118, 148, 0.24);
  background: linear-gradient(180deg, rgba(14, 24, 39, 0.96), rgba(18, 30, 48, 0.98));
  box-shadow: 0 18px 40px rgba(7, 12, 18, 0.16);
}

.source-banner {
  padding: 12px 14px;
  border-radius: var(--radius-md);
  border: 1px solid rgba(125, 211, 252, 0.18);
  background: rgba(125, 211, 252, 0.08);
  color: var(--text-secondary);
  font-size: 12px;
  line-height: 1.6;
}

.toolbar-card,
.table-card {
  padding: 16px;
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 1px;
  overflow: hidden;
}

.mini-metric {
  padding: 18px;
  background: rgba(11, 19, 31, 0.22);
}

.mini-label,
.mini-sub,
.fund-meta {
  color: var(--text-secondary);
}

.mini-value {
  margin-top: 10px;
  font-size: 24px;
  font-weight: 700;
}

.mini-sub {
  margin-top: 6px;
  font-size: 12px;
}

.fund-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

:deep(.table-actions) {
  display: flex;
  gap: 8px 10px;
  flex-wrap: wrap;
  align-items: center;
}

:deep(.table-actions .n-button) {
  margin: 0;
}

.profit-text {
  color: var(--profit);
  font-weight: 600;
}

.loss-text {
  color: var(--loss);
  font-weight: 600;
}
</style>
