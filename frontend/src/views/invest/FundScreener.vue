<script setup>
import { computed, h, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { NButton, NTag, NText, useMessage } from 'naive-ui'
import { FollowFund, GetFundScreener, RefreshFundScreenerData, UnFollowFund } from '../../../wailsjs/go/main/App'
import FundInsightDrawer from '../portfolio/components/FundInsightDrawer.vue'

const message = useMessage()
const loading = ref(false)
const refreshing = ref(false)
const rows = ref([])
const total = ref(0)
const universeCount = ref(0)
const screenedCount = ref(0)
const lastRefreshHint = ref('')
const typeOptions = ref([])
const industryOptions = ref([])
const refreshStatus = ref(null)
const showDetail = ref(false)
const activeFund = ref(null)
let pollingTimer = null

const filter = reactive({
  keyword: '',
  fundType: '',
  category: '',
  industry: '',
  minReturn7: null,
  minReturn1: null,
  minReturn3: null,
  maxDrawdown12: null,
  onlyWatchlist: false,
  sortBy: 'growth3',
  sortOrder: 'desc',
  page: 1,
  pageSize: 20
})

const categoryOptions = [
  { label: '全部类别', value: '' },
  { label: '债券/稳健', value: 'bond' },
  { label: '现金管理', value: 'cash' },
  { label: '权益/混合', value: 'equity' },
  { label: '其他', value: 'other' }
]

const sortOptions = [
  { label: '近3个月收益', value: 'growth3' },
  { label: '近1个月收益', value: 'growth1' },
  { label: '近7天收益', value: 'growth7' },
  { label: '近6个月收益', value: 'growth6' },
  { label: '近1年收益', value: 'growth12' },
  { label: '近1年最大回撤', value: 'drawdown12' },
  { label: '更新时间', value: 'updatedAt' }
]

const orderOptions = [
  { label: '从高到低', value: 'desc' },
  { label: '从低到高', value: 'asc' }
]

const refreshBannerType = computed(() => {
  if (refreshStatus.value?.refreshing) return 'warning'
  if (refreshStatus.value?.needsRefresh) return 'info'
  return 'success'
})

const refreshBannerTitle = computed(() => {
  if (refreshStatus.value?.refreshing) return '今日基金指标正在后台更新'
  if (refreshStatus.value?.needsRefresh) return '今日基金指标尚未刷新'
  return '今日基金指标已就绪'
})

const refreshBannerText = computed(() => {
  const status = refreshStatus.value
  if (!status) return ''
  if (status.refreshing) {
    const progress = status.progressTotal > 0 ? `${status.progressCurrent}/${status.progressTotal}` : '准备中'
    const current = status.currentCode ? `，当前：${status.currentCode}` : ''
    return `后台正在增量更新基金池指标，当前进度 ${progress}${current}。你现在看到的是本地缓存结果，更新完成后会更准确。`
  }
  if (status.needsRefresh) {
    return '今天还没有更新过基金池指标。首次进入会自动启动后台更新，筛选结果会随着更新逐步变新。'
  }
  const latest = status.lastRefreshHint || '今日已更新'
  return `可以直接筛选，最近一次指标更新时间：${latest}`
})

function percentText(value) {
  if (value === null || value === undefined || Number.isNaN(Number(value))) {
    return '-'
  }
  const num = Number(value)
  return `${num > 0 ? '+' : ''}${num.toFixed(2)}%`
}

function drawdownText(value) {
  if (value === null || value === undefined || Number.isNaN(Number(value))) {
    return '-'
  }
  return `${Number(value).toFixed(2)}%`
}

function percentType(value, reverse = false) {
  if (value === null || value === undefined || Number.isNaN(Number(value))) {
    return 'default'
  }
  const num = Number(value)
  if (reverse) {
    if (num <= 5) return 'success'
    if (num <= 10) return 'warning'
    return 'error'
  }
  if (num > 0) return 'error'
  if (num < 0) return 'success'
  return 'default'
}

function syncRefreshStatus(status, silent = false) {
  refreshStatus.value = status || null
  if (status?.triggered && !silent) {
    message.info('今日首次进入已自动启动基金池后台更新。')
  }
  if (status?.refreshing) {
    startPolling()
  } else {
    stopPolling()
  }
}

async function loadData(resetPage = false, silent = false) {
  if (resetPage) {
    filter.page = 1
  }
  if (!silent) {
    loading.value = true
  }
  try {
    const result = await GetFundScreener({
      ...filter,
      minReturn7: filter.minReturn7 === null ? null : Number(filter.minReturn7),
      minReturn1: filter.minReturn1 === null ? null : Number(filter.minReturn1),
      minReturn3: filter.minReturn3 === null ? null : Number(filter.minReturn3),
      maxDrawdown12: filter.maxDrawdown12 === null ? null : Number(filter.maxDrawdown12)
    })
    rows.value = result?.items || []
    total.value = result?.total || 0
    universeCount.value = result?.universeCount || 0
    screenedCount.value = result?.screenedCount || 0
    lastRefreshHint.value = result?.lastRefreshHint || ''
    syncRefreshStatus(result?.refreshStatus, silent)
    typeOptions.value = [{ label: '全部类型', value: '' }].concat(
      (result?.typeOptions || []).map((item) => ({ label: item, value: item }))
    )
    industryOptions.value = [{ label: '全部行业', value: '' }].concat(
      (result?.industryOptions || []).map((item) => ({ label: item, value: item }))
    )
  } catch (error) {
    console.error(error)
    if (!silent) {
      message.error('基金筛选加载失败')
    }
  } finally {
    if (!silent) {
      loading.value = false
    }
  }
}

function startPolling() {
  stopPolling()
  pollingTimer = window.setInterval(() => {
    loadData(false, true)
  }, 15000)
}

function stopPolling() {
  if (pollingTimer) {
    window.clearInterval(pollingTimer)
    pollingTimer = null
  }
}

async function handleRefreshMetrics() {
  refreshing.value = true
  try {
    const result = await RefreshFundScreenerData(300)
    message.success(`已增量更新 ${result?.refreshed || 0} 只基金指标`)
    await loadData(false)
  } catch (error) {
    console.error(error)
    message.error('增量更新基金池指标失败')
  } finally {
    refreshing.value = false
  }
}

function resetFilters() {
  filter.keyword = ''
  filter.fundType = ''
  filter.category = ''
  filter.industry = ''
  filter.minReturn7 = null
  filter.minReturn1 = null
  filter.minReturn3 = null
  filter.maxDrawdown12 = null
  filter.onlyWatchlist = false
  filter.sortBy = 'growth3'
  filter.sortOrder = 'desc'
  loadData(true)
}

async function toggleWatch(row) {
  try {
    if (row.watchlist) {
      await UnFollowFund(row.code)
      message.success('已取消自选')
    } else {
      await FollowFund(row.code)
      message.success('已加入自选')
    }
    await loadData(false)
  } catch (error) {
    console.error(error)
    message.error('更新自选状态失败')
  }
}

function openDetail(row) {
  activeFund.value = normalizeScreenerFund(row)
  showDetail.value = true
}

function normalizeScreenerFund(row) {
  return {
    stockCode: row.code,
    stockName: row.name,
    fundType: row.fundType,
    fundCompany: row.company,
    fundManager: row.manager,
    fundScale: row.scale,
    fundRating: row.rating,
    category: row.category,
    categoryLabel: row.categoryLabel,
    netGrowth1: row.netGrowth1,
    netGrowth3: row.netGrowth3,
    netGrowth6: row.netGrowth6,
    netGrowth12: row.netGrowth12
  }
}

async function handleDetailRefreshed() {
  await loadData(false, true)
}

const columns = computed(() => [
  {
    title: '基金',
    key: 'name',
    width: 260,
    render(row) {
      return h('div', { class: 'cell-main' }, [
        h(NButton, { text: true, type: 'primary', onClick: () => openDetail(row) }, () => row.name || row.code),
        h('div', { class: 'fund-meta' }, `${row.code} / ${row.company || '基金公司待补'}`)
      ])
    }
  },
  {
    title: '类型',
    key: 'fundType',
    width: 190,
    render(row) {
      return h('div', { class: 'cell-stack' }, [
        h(NTag, { size: 'small', type: 'info', bordered: false }, { default: () => row.categoryLabel || '未分类' }),
        h('div', { class: 'fund-meta' }, row.fundType || '类型待补')
      ])
    }
  },
  {
    title: '行业',
    key: 'topIndustry',
    width: 180,
    render(row) {
      return h('div', { class: 'cell-stack' }, [
        h('div', { class: 'fund-name small' }, row.topIndustry || '暂无行业口径'),
        h('div', { class: 'fund-meta' }, row.topIndustryWeight != null ? `${Number(row.topIndustryWeight).toFixed(2)}% / ${row.topIndustryDate || '-'}` : '-')
      ])
    }
  },
  {
    title: '近7天',
    key: 'netGrowth7',
    width: 110,
    render(row) {
      return h(NText, { type: percentType(row.netGrowth7) }, { default: () => percentText(row.netGrowth7) })
    }
  },
  {
    title: '近1月',
    key: 'netGrowth1',
    width: 110,
    render(row) {
      return h(NText, { type: percentType(row.netGrowth1) }, { default: () => percentText(row.netGrowth1) })
    }
  },
  {
    title: '近3月',
    key: 'netGrowth3',
    width: 110,
    render(row) {
      return h(NText, { type: percentType(row.netGrowth3) }, { default: () => percentText(row.netGrowth3) })
    }
  },
  {
    title: '近1年最大回撤',
    key: 'maxDrawdown12',
    width: 140,
    render(row) {
      return h(NText, { type: percentType(row.maxDrawdown12, true) }, { default: () => drawdownText(row.maxDrawdown12) })
    }
  },
  {
    title: '更新时间',
    key: 'screenUpdatedAt',
    width: 160,
    render(row) {
      return h('span', { class: 'fund-meta' }, row.screenUpdatedAt || '-')
    }
  },
  {
    title: '操作',
    key: 'actions',
    width: 180,
    fixed: 'right',
    render(row) {
      return h('div', { class: 'table-actions' }, [
        h(
          NButton,
          {
            size: 'tiny',
            secondary: true,
            type: row.watchlist ? 'primary' : 'success',
            onClick: () => toggleWatch(row)
          },
          { default: () => (row.watchlist ? '取消自选' : '加入自选') }
        ),
        h(
          NButton,
          {
            size: 'tiny',
            secondary: true,
            type: 'warning',
            onClick: () => openDetail(row)
          },
          { default: () => '详情' }
        )
      ])
    }
  }
])

onMounted(() => {
  loadData(true)
})

onBeforeUnmount(() => {
  stopPolling()
})
</script>

<template>
  <div class="fade-in screener-page">
    <n-page-header
      class="view-header"
      title="基金筛选"
      subtitle="按类型、阶段收益、回撤和行业快速筛出适合自己复盘的基金池"
    />

    <n-alert v-if="refreshStatus" :type="refreshBannerType" :title="refreshBannerTitle" class="status-alert" show-icon>
      {{ refreshBannerText }}
    </n-alert>

    <div class="platform-card stats-shell">
      <div class="stat-item">
        <div class="stat-value">{{ universeCount }}</div>
        <div class="stat-label">基金池数量</div>
      </div>
      <div class="stat-item">
        <div class="stat-value">{{ screenedCount }}</div>
        <div class="stat-label">已补筛选指标</div>
      </div>
      <div class="stat-item wide">
        <div class="stat-value text">{{ lastRefreshHint || '尚未刷新' }}</div>
        <div class="stat-label">最近一次指标更新时间</div>
      </div>
      <div class="stat-actions">
        <n-button :loading="refreshing" @click="handleRefreshMetrics">增量更新指标</n-button>
      </div>
    </div>

    <div class="platform-card filter-shell">
      <n-grid :cols="24" :x-gap="12" :y-gap="12">
        <n-gi :span="6">
          <n-input v-model:value="filter.keyword" placeholder="基金代码 / 名称 / 公司 / 行业" clearable />
        </n-gi>
        <n-gi :span="5">
          <n-select v-model:value="filter.fundType" :options="typeOptions" placeholder="基金类型" />
        </n-gi>
        <n-gi :span="4">
          <n-select v-model:value="filter.category" :options="categoryOptions" placeholder="大类" />
        </n-gi>
        <n-gi :span="4">
          <n-select v-model:value="filter.industry" :options="industryOptions" placeholder="所属行业" />
        </n-gi>
        <n-gi :span="5">
          <n-space>
            <n-button type="primary" @click="loadData(true)">开始筛选</n-button>
            <n-button tertiary @click="resetFilters">重置</n-button>
          </n-space>
        </n-gi>

        <n-gi :span="4">
          <n-input-number v-model:value="filter.minReturn7" clearable placeholder="近7天最低收益(%)" style="width: 100%" />
        </n-gi>
        <n-gi :span="4">
          <n-input-number v-model:value="filter.minReturn1" clearable placeholder="近1月最低收益(%)" style="width: 100%" />
        </n-gi>
        <n-gi :span="4">
          <n-input-number v-model:value="filter.minReturn3" clearable placeholder="近3月最低收益(%)" style="width: 100%" />
        </n-gi>
        <n-gi :span="4">
          <n-input-number v-model:value="filter.maxDrawdown12" clearable placeholder="近1年最大回撤上限(%)" style="width: 100%" />
        </n-gi>
        <n-gi :span="4">
          <n-select v-model:value="filter.sortBy" :options="sortOptions" placeholder="排序字段" />
        </n-gi>
        <n-gi :span="4">
          <n-select v-model:value="filter.sortOrder" :options="orderOptions" placeholder="排序方向" />
        </n-gi>
      </n-grid>

      <div class="filter-foot">
        <n-checkbox v-model:checked="filter.onlyWatchlist">只看基金自选</n-checkbox>
        <span class="filter-note">基金池保存在本地数据库，进入页面时会自动检查是否需要做今日更新。</span>
      </div>
    </div>

    <div class="platform-card table-shell">
      <n-data-table
        remote
        :loading="loading"
        :columns="columns"
        :data="rows"
        :pagination="{
          page: filter.page,
          pageSize: filter.pageSize,
          itemCount: total,
          showSizePicker: true,
          pageSizes: [20, 50, 100],
          onUpdatePage: (page) => {
            filter.page = page
            loadData(false)
          },
          onUpdatePageSize: (pageSize) => {
            filter.pageSize = pageSize
            filter.page = 1
            loadData(false)
          }
        }"
      />
    </div>

    <FundInsightDrawer
      v-model:show="showDetail"
      :fund="activeFund"
      @refreshed="handleDetailRefreshed"
    />
  </div>
</template>

<style scoped>
.screener-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.status-alert {
  margin-top: -4px;
}

.stats-shell {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
  align-items: center;
}

.stat-item {
  min-height: 88px;
  padding: 18px;
  border-radius: 16px;
  background: linear-gradient(180deg, rgba(18, 40, 68, 0.9), rgba(13, 28, 46, 0.92));
  border: 1px solid rgba(102, 163, 255, 0.14);
}

.stat-item.wide {
  min-width: 0;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #f3f7ff;
}

.stat-value.text {
  font-size: 16px;
  line-height: 1.5;
}

.stat-label {
  margin-top: 6px;
  font-size: 12px;
  color: rgba(222, 234, 255, 0.68);
}

.stat-actions {
  display: flex;
  justify-content: flex-end;
}

.filter-shell {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.filter-foot {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  font-size: 12px;
  color: rgba(222, 234, 255, 0.62);
}

.filter-note {
  text-align: right;
}

.table-shell {
  overflow: hidden;
}

.cell-main,
.cell-stack {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.fund-name {
  font-weight: 600;
  color: #eef5ff;
}

.fund-name.small {
  font-size: 14px;
}

.fund-meta {
  font-size: 12px;
  color: rgba(222, 234, 255, 0.62);
}

:deep(.table-actions) {
  display: flex;
  align-items: center;
  gap: 8px 10px;
  flex-wrap: wrap;
}

:deep(.table-actions .n-button) {
  margin: 0;
}
</style>
