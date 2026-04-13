<template>
  <div class="fade-in fund-watch-page">
    <n-page-header title="基金自选" subtitle="跟踪关注基金的估算、阶段收益，并支持对比推荐和 AI 解读。">
      <template #extra>
        <n-space>
          <n-button :disabled="checkedRowKeys.length < 2" @click="showCompareModal = true">
            基金对比{{ checkedRowKeys.length ? ` (${checkedRowKeys.length})` : '' }}
          </n-button>
          <n-button :disabled="groupRows.length === 0" @click="showCollectionAI = true">AI分析当前页签</n-button>
          <n-button :loading="loading" @click="loadData">刷新列表</n-button>
        </n-space>
      </template>
    </n-page-header>

    <div class="platform-card add-shell">
      <div class="watch-search-bar">
        <div class="watch-search-main">
          <n-input
            v-model:value="filterKeyword"
            clearable
            placeholder="筛选当前页签里的自选基金，支持代码 / 名称 / 类型 / 跟踪标的"
          />
          <n-button type="primary" @click="openAddFundModal">加入基金</n-button>
        </div>
        <div class="watch-search-meta">
          <n-text depth="3">{{ filterSummary }}</n-text>
          <n-text depth="3">需要新增基金时，点右侧“加入基金”后再搜索代码或名称。</n-text>
        </div>
      </div>
    </div>

    <div class="platform-card watch-toolbar-card">
      <div class="watch-toolbar">
        <n-tabs v-model:value="activeGroupTab" type="segment" animated>
          <n-tab-pane
            v-for="tab in watchTabs"
            :key="tab.value"
            :name="tab.value"
            :tab="tab.label"
          />
        </n-tabs>
      </div>
    </div>

    <div class="platform-card recommendation-status-card" v-if="recommendationStatus">
      <div class="recommendation-status-head">
        <div class="recommendation-status-title">推荐检索状态</div>
        <div class="recommendation-status-badge" :class="{ running: recommendationStatus.refreshing, done: recommendationStatus.state === 'completed' }">
          {{ recommendationStatus.stateLabel || '未开始' }}
        </div>
      </div>
      <div class="recommendation-status-summary">{{ recommendationSummary }}</div>
      <div class="recommendation-status-hint">{{ recommendationHint }}</div>
    </div>

    <div class="platform-card table-shell">
      <div class="table-toolbar">
        <n-space>
          <n-button v-if="activeCustomGroup" size="small" @click="openRenameGroup">重命名</n-button>
          <n-button v-if="activeCustomGroup" size="small" type="error" ghost @click="handleDeleteGroup">删除页签</n-button>
          <n-button size="small" @click="openCreateGroup">新建页签</n-button>
        </n-space>
      </div>
      <n-data-table
        :loading="loading"
        :columns="columns"
        :data="sortedRows"
        :pagination="{ pageSize: 10 }"
        :scroll-x="1520"
        :row-key="(row) => row.code"
        :checked-row-keys="checkedRowKeys"
        striped
        @update:checked-row-keys="handleCheckedRowKeys"
      />
    </div>

    <FundInsightDrawer
      v-model:show="showDetail"
      :fund="activeFund"
      @refreshed="loadData"
    />

    <FundAIAnalysisModal
      v-model:show="showSingleAI"
      mode="single"
      :fund-code="activeFund?.code || activeFund?.stockCode || ''"
      :title="activeFund ? ((activeFund.name || activeFund.stockName || activeFund.code) + ' AI 分析') : '基金 AI 分析'"
    />

    <FundAIAnalysisModal
      v-model:show="showCollectionAI"
      mode="collection"
      scope="tab"
      :scope-label="currentWatchTabName"
      :fund-codes="groupRows.map((item) => item.code)"
      :title="`${currentWatchTabName} AI 分析`"
    />

    <FundAIAnalysisModal
      v-model:show="showBetterAI"
      mode="better"
      :better-reference-code="betterReferenceCode"
      :better-dimension="betterDimension"
      :same-type-only="sameTypeOnly"
      :same-sub-type-only="betterSameSubTypeOnly"
      :fee-free7="betterFeeFree7"
      :fee-free30="betterFeeFree30"
      :include-a-class="betterIncludeAClass"
      :only-a-class="betterIncludeAClass"
      :better-top-n="3"
      :title="betterAITitle"
    />

    <FundCompareModal
      v-model:show="showCompareModal"
      :codes="checkedRowKeys"
      :tab-label="currentWatchTabName"
    />

    <n-modal
      v-model:show="showAddFundModal"
      preset="card"
      title="加入自选基金"
      style="width: 560px; max-width: calc(100vw - 32px);"
    >
      <n-space vertical size="large">
        <n-auto-complete
          v-model:value="addKeyword"
          :options="fundOptions"
          placeholder="输入基金代码或名称，加入自选"
          clearable
          @search="searchFunds"
          @update:value="handleKeywordUpdate"
          @select="handleSelectFund"
        />
        <n-text depth="3">支持代码、名称模糊检索；如果你当前停留在某个自定义页签，加入后会自动归到该页签。</n-text>
      </n-space>
    </n-modal>

    <n-modal
      v-model:show="showGroupModal"
      preset="card"
      :title="groupModalTitle"
      style="width: 420px;"
    >
      <n-space vertical size="large">
        <n-select
          v-if="groupModalMode === 'assign'"
          v-model:value="assignGroupValue"
          :options="groupSelectOptions"
          placeholder="选择一个页签"
        />
        <n-input
          v-else
          v-model:value="editingGroupName"
          placeholder="例如：核心债基 / 进攻观察 / 长期定投"
          @keydown.enter.prevent="saveGroup"
        />
        <n-text depth="3">
          {{ groupModalHint }}
        </n-text>
      </n-space>
      <template #action>
        <n-space justify="end">
          <n-button @click="closeGroupModal">取消</n-button>
          <n-button type="primary" @click="saveGroup">{{ groupModalActionText }}</n-button>
        </n-space>
      </template>
    </n-modal>

    <n-modal
      v-model:show="showBetterModal"
      preset="card"
      :title="betterTitle"
      style="width: 1380px; max-width: calc(100vw - 32px);"
    >
      <div class="better-toolbar">
        <div class="better-toolbar-left">
          <n-checkbox v-model:checked="sameTypeOnly">只看同类型</n-checkbox>
          <div class="better-filter-group">
            <n-checkbox
              v-model:checked="betterSameSubTypeOnly"
              :disabled="betterSubTypeDisabled"
            >
              {{ betterSubTypeLabelText }}
            </n-checkbox>
            <n-checkbox v-model:checked="betterFeeFree7">免7</n-checkbox>
            <n-checkbox v-model:checked="betterFeeFree30">免30</n-checkbox>
            <n-checkbox v-model:checked="betterIncludeAClass">只看A类</n-checkbox>
          </div>
          <n-text depth="3">当前栏目支持直接对 Top3 推荐基金做 AI 对比分析。</n-text>
        </div>
        <div class="better-toolbar-right">
          <n-button :disabled="!betterRows.length || betterLoading" @click="openBetterAI">
            AI分析{{ currentBetterDimensionLabel }}Top3
          </n-button>
          <n-button :loading="betterLoading" @click="loadBetterFunds(true)">刷新结果</n-button>
        </div>
      </div>
      <div v-if="betterResult" class="better-context" :class="{ warning: betterResult.fallbackApplied }">
        <div class="better-context-head">
          <div class="better-context-title">{{ betterResult.sortLabel || '按推荐得分排序' }}</div>
          <div class="better-context-meta">
            <span>{{ betterResult.scopeLabel || '基金池对比' }}</span>
            <span>推荐 {{ betterResult.total || 0 }} / 样本 {{ betterResult.comparedUniverse || 0 }} / 目标池 {{ betterResult.universeTotal || betterResult.comparedUniverse || 0 }} 只</span>
            <span>更新时间 {{ betterResult.refreshStatus?.lastRefreshHint || '本地缓存' }}</span>
          </div>
        </div>
        <div class="better-context-hint">{{ betterResult.dataHint || '推荐结果来自后台已生成的推荐缓存。' }} 同类排名来自东方财富官网阶段排名，回撤基于基金净值趋势测算。</div>
      </div>
      <n-tabs v-model:value="betterDimension" type="segment" animated @update:value="loadBetterFunds()">
        <n-tab name="lower_drawdown" tab="回撤更低" />
        <n-tab name="higher_return" tab="收益更高" />
        <n-tab name="balanced" tab="实力均衡更优" />
      </n-tabs>
      <div class="better-table-wrap">
        <n-data-table
          class="better-result-table"
          :loading="betterLoading"
          :columns="betterColumns"
          :data="betterRows"
          :pagination="{ pageSize: 8 }"
          :scroll-x="1680"
        />
      </div>
    </n-modal>
  </div>
</template>

<script setup>
import { computed, h, onMounted, onUnmounted, ref, watch } from 'vue'
import { NButton, NIcon, NTag, NText, NTooltip, useDialog, useMessage } from 'naive-ui'
import { AlertCircleOutline } from '@vicons/ionicons5'
import { FollowFund, GetBetterFunds, GetFollowedFund, GetfundList, OpenURL, UnFollowFund } from '../../../wailsjs/go/main/App'
import FundInsightDrawer from '../portfolio/components/FundInsightDrawer.vue'
import FundAIAnalysisModal from '../portfolio/components/FundAIAnalysisModal.vue'
import FundCompareModal from './components/FundCompareModal.vue'

const dialog = useDialog()
const message = useMessage()
const loading = ref(false)
const rows = ref([])
const fundOptions = ref([])
const addKeyword = ref('')
const filterKeyword = ref('')
const activeFund = ref(null)
const showDetail = ref(false)
const showSingleAI = ref(false)
const showCollectionAI = ref(false)
const showBetterAI = ref(false)
const showBetterModal = ref(false)
const betterLoading = ref(false)
const sameTypeOnly = ref(false)
const betterSameSubTypeOnly = ref(false)
const betterFeeFree7 = ref(true)
const betterFeeFree30 = ref(true)
const betterIncludeAClass = ref(false)
const betterDimension = ref('balanced')
const betterResult = ref(null)
const recommendationStatus = ref(null)
const checkedRowKeys = ref([])
const showCompareModal = ref(false)
const activeGroupTab = ref('__all__')
const customGroups = ref([])
const showAddFundModal = ref(false)
const showGroupModal = ref(false)
const groupingFund = ref(null)
const editingGroupName = ref('')
const groupModalMode = ref('create')
const editingGroupSource = ref('')

const betterReferenceCode = ref('')
const watchSortState = ref({
  key: '',
  order: 'desc'
})
const ALL_GROUP_TAB = '__all__'
const UNGROUPED_GROUP_TAB = '__ungrouped__'
const CUSTOM_GROUP_STORAGE_KEY = 'fund-watch-custom-groups'
const RESERVED_GROUP_NAMES = [ALL_GROUP_TAB, UNGROUPED_GROUP_TAB, '全部', '未分组']

const groupNames = computed(() => Array.from(new Set([
  ...customGroups.value,
  ...rows.value.map((row) => String(row.watchGroup || '').trim()).filter(Boolean)
])))

const groupCounts = computed(() => {
  const counts = new Map()
  for (const row of rows.value) {
    const group = String(row.watchGroup || '').trim()
    if (group) {
      counts.set(group, (counts.get(group) || 0) + 1)
    }
  }
  return counts
})

const groupSelectOptions = computed(() => [
  { label: '未分组', value: UNGROUPED_GROUP_TAB },
  ...groupNames.value.map((group) => ({
    label: `${group} (${groupCounts.value.get(group) || 0})`,
    value: group
  }))
])

const assignGroupValue = computed({
  get: () => editingGroupName.value || UNGROUPED_GROUP_TAB,
  set: (value) => {
    editingGroupName.value = value === UNGROUPED_GROUP_TAB ? '' : String(value || '').trim()
  }
})

const watchTabs = computed(() => {
  let ungroupedCount = 0
  for (const row of rows.value) {
    const group = String(row.watchGroup || '').trim()
    if (!group) {
      ungroupedCount += 1
    }
  }

  return [
    { label: `全部 (${rows.value.length})`, value: ALL_GROUP_TAB },
    ...groupNames.value.map((group) => ({ label: `${group} (${groupCounts.value.get(group) || 0})`, value: group })),
    { label: `未分组 (${ungroupedCount})`, value: UNGROUPED_GROUP_TAB }
  ]
})

const currentWatchTabLabel = computed(() => {
  const current = watchTabs.value.find((tab) => tab.value === activeGroupTab.value)
  return current?.label || '当前页签'
})
const currentWatchTabName = computed(() => String(currentWatchTabLabel.value || '').replace(/\s*\(\d+\)\s*$/, '') || '当前页签')

const groupRows = computed(() => {
  if (activeGroupTab.value === ALL_GROUP_TAB) return rows.value
  if (activeGroupTab.value === UNGROUPED_GROUP_TAB) {
    return rows.value.filter((row) => !String(row.watchGroup || '').trim())
  }
  return rows.value.filter((row) => String(row.watchGroup || '').trim() === activeGroupTab.value)
})

const filteredRows = computed(() => {
  const keyword = normalizeWatchKeyword(filterKeyword.value)
  if (!keyword) {
    return groupRows.value
  }
  return groupRows.value.filter((row) => matchesWatchKeyword(row, keyword))
})

const filterSummary = computed(() => {
  const total = groupRows.value.length
  const keyword = String(filterKeyword.value || '').trim()
  if (!keyword) {
    return `${currentWatchTabName.value} 当前共 ${total} 只基金。`
  }
  return `${currentWatchTabName.value} 中，关键词“${keyword}”筛出 ${filteredRows.value.length}/${total} 只基金。`
})

const sortedRows = computed(() => {
  const list = [...filteredRows.value]
  const { key, order } = watchSortState.value
  if (!key) {
    return list
  }
  return list.sort((left, right) => compareNullableValues(
    getWatchSortValue(left, key),
    getWatchSortValue(right, key),
    order
  ))
})

const betterTitle = computed(() => {
  const name = activeFund.value?.name || activeFund.value?.stockName || activeFund.value?.code || activeFund.value?.stockCode || '当前基金'
  return name + ' 的对比推荐'
})
const betterSubTypeLabel = computed(() => String(
  betterResult.value?.reference?.trackingTarget ||
  activeFund.value?.trackingTarget ||
  ''
).trim())
const betterSubTypeDisabled = computed(() => !sameTypeOnly.value || !betterSubTypeLabel.value)
const betterSubTypeLabelText = computed(() => {
  if (!betterSubTypeLabel.value) {
    return '同子类型'
  }
  return `同子类型（${betterSubTypeLabel.value}）`
})
const currentBetterDimensionLabel = computed(() => {
  switch (betterDimension.value) {
    case 'lower_drawdown':
      return '回撤更低'
    case 'higher_return':
      return '收益更高'
    default:
      return '实力均衡更优'
  }
})
const betterAITitle = computed(() => `${betterTitle.value} · ${currentBetterDimensionLabel.value} Top3 AI分析`)
const betterRows = computed(() => betterResult.value?.candidates || [])
const recommendationSummary = computed(() => {
  const status = recommendationStatus.value
  if (!status) return ''
  const parts = [
    `当前自选 ${status.watchlistCount || 0} 只`,
    `推荐缓存已完成 ${status.completedCount || 0} 只`
  ]
  if (status.pendingCount > 0) {
    parts.push(`剩余 ${status.pendingCount || 0} 只`)
  }
  if (status.refreshing && status.progressTotal > 0) {
    parts.push(`后台进度 ${status.progressCurrent || 0}/${status.progressTotal || 0}`)
  }
  if (status.currentCode) {
    parts.push(`当前 ${status.currentCode}`)
  }
  let summary = parts.join('，') + '。'
  if (status.lastRefreshHint) {
    summary += ` 最近完成时间：${status.lastRefreshHint}`
  }
  return summary
})
const recommendationHint = computed(() => {
  const status = recommendationStatus.value
  if (!status) {
    return '后台会按今日自选基金逐只补齐推荐缓存。'
  }
  const base = status.message || '后台会按今日自选基金逐只补齐推荐缓存。'
  if (status.state === 'completed') {
    return base
  }
  return `${base} 推荐结果每天只生成一次，未完成时下次会从剩余基金继续补齐。`
})
const activeCustomGroup = computed(() => isCustomGroupTab(activeGroupTab.value) ? activeGroupTab.value : '')
const activeCustomGroupRows = computed(() => rows.value.filter((row) => String(row.watchGroup || '').trim() === activeCustomGroup.value))
const groupModalTitle = computed(() => {
  if (groupModalMode.value === 'assign') return '设置基金分组'
  if (groupModalMode.value === 'rename') return '重命名页签'
  return '新建自定义页签'
})
const groupModalHint = computed(() => {
  if (groupModalMode.value === 'assign') return '直接选择已有页签；选“未分组”会把基金移出自定义页签。'
  if (groupModalMode.value === 'rename') return '重命名后，这个页签下的基金会一起迁移到新名称。'
  return '先建一个分组页签，后面可以把基金归进去。'
})
const groupModalActionText = computed(() => {
  if (groupModalMode.value === 'rename') return '重命名'
  if (groupModalMode.value === 'create') return '创建'
  return '保存'
})

function renderSortableTitle(label, sortKey, defaultOrder = 'desc') {
  const active = watchSortState.value.key === sortKey
  const order = active ? watchSortState.value.order : defaultOrder
  return h(
    'span',
    {
      role: 'button',
      tabindex: 0,
      'aria-label': `${label}排序`,
      style: {
        display: 'inline-flex',
        alignItems: 'center',
        gap: '4px',
        cursor: 'pointer',
        userSelect: 'none',
        color: active ? '#eef5ff' : 'inherit',
        font: 'inherit',
        lineHeight: 1,
        whiteSpace: 'nowrap'
      },
      onClick: (event) => {
        event.stopPropagation()
        toggleWatchSort(sortKey)
      },
      onKeydown: (event) => {
        if (event.key === 'Enter' || event.key === ' ') {
          event.preventDefault()
          toggleWatchSort(sortKey)
        }
      }
    },
    [
      h('span', null, label),
      h(
        'span',
        {
          'aria-hidden': 'true',
          style: {
            display: 'inline-flex',
            alignItems: 'center',
            fontSize: '11px',
            color: active ? 'var(--primary-color)' : 'rgba(222, 234, 255, 0.42)'
          }
        },
        order === 'asc' ? '↑' : '↓'
      )
    ]
  )
}

function toggleWatchSort(sortKey) {
  if (watchSortState.value.key === sortKey) {
    watchSortState.value.order = watchSortState.value.order === 'desc' ? 'asc' : 'desc'
    return
  }
  watchSortState.value = {
    key: sortKey,
    order: 'desc'
  }
}

function getWatchSortValue(row, sortKey) {
  switch (sortKey) {
    case 'estimate':
      return row.netEstimatedRate ?? row.latestDailyRate ?? null
    case 'netGrowth1':
      return row.fundBasic?.netGrowth1
    case 'netGrowth3':
      return row.fundBasic?.netGrowth3
    case 'netGrowth6':
      return row.fundBasic?.netGrowth6
    default:
      return null
  }
}

function compareNullableValues(left, right, order = 'desc') {
  const leftMissing = left === null || left === undefined || Number.isNaN(Number(left))
  const rightMissing = right === null || right === undefined || Number.isNaN(Number(right))
  if (leftMissing && rightMissing) return 0
  if (leftMissing) return 1
  if (rightMissing) return -1
  const diff = Number(left) - Number(right)
  return order === 'asc' ? diff : -diff
}

function normalizeWatchKeyword(value) {
  return String(value || '').trim().toLowerCase()
}

function matchesWatchKeyword(row, keyword) {
  const fields = [
    row?.name,
    row?.code,
    row?.watchGroup,
    row?.fundBasic?.type,
    row?.fundBasic?.trackingTarget,
    row?.trackingTarget,
    row?.fundBasic?.company,
    row?.fundBasic?.manager,
    row?.categoryLabel
  ]
  return fields
    .map((item) => normalizeWatchKeyword(item))
    .filter(Boolean)
    .some((item) => item.includes(keyword))
}

const columns = [
  {
    type: 'selection',
    multiple: true
  },
  {
    title: '基金',
    key: 'name',
    width: 240,
    fixed: 'left',
    render: (row) => h('div', { class: 'cell-main' }, [
      h(NButton, { text: true, type: 'primary', onClick: () => openDetail(row) }, () => row.name || row.code),
      h('div', { class: 'cell-meta' }, row.code)
    ])
  },
  {
    title: '类型',
    key: 'fundType',
    width: 180,
    render: (row) => h(NTag, { size: 'small', bordered: false, type: 'info' }, { default: () => row.fundBasic?.type || '类型待同步' })
  },
  {
    title: '估算 / 最近1日',
    title: () => renderSortableTitle('估算 / 最近1日', 'estimate'),
    key: 'estimate',
    width: 180,
    render: (row) => h('div', { class: 'cell-stack' }, [
      h(
        NText,
        {
          type: row.netEstimatedUnit
            ? (Number(row.netEstimatedRate || 0) >= 0 ? 'error' : 'success')
            : 'default'
        },
        {
          default: () => (
            row.netEstimatedUnit
              ? Number(row.netEstimatedUnit).toFixed(4) + ' / ' + signedPercent(row.netEstimatedRate) + '%'
              : (row.netUnitValue ? Number(row.netUnitValue).toFixed(4) + ' / 最新净值' : '暂无估算')
          )
        }
      ),
      h('div', { class: 'cell-meta' }, row.netEstimatedUnitTime || row.netUnitValueDate || '-')
    ])
  },
  {
    title: () => renderSortableTitle('近1月', 'netGrowth1'),
    key: 'netGrowth1',
    width: 100,
    render: (row) => renderPercent(row.fundBasic?.netGrowth1)
  },
  {
    title: () => renderSortableTitle('近3月', 'netGrowth3'),
    key: 'netGrowth3',
    width: 100,
    render: (row) => renderPercent(row.fundBasic?.netGrowth3)
  },
  {
    title: () => renderSortableTitle('近6月', 'netGrowth6'),
    key: 'netGrowth6',
    width: 100,
    render: (row) => renderPercent(row.fundBasic?.netGrowth6)
  },
  {
    title: '操作',
    key: 'actions',
    width: 360,
    render: (row) => h('div', { class: 'table-actions' }, [
      h(NButton, { size: 'tiny', secondary: true, type: 'warning', onClick: () => openSingleAI(row) }, () => 'AI分析'),
      h(NButton, { size: 'tiny', secondary: true, type: 'success', onClick: () => openBetter(row) }, () => '对比推荐'),
      h(NButton, { size: 'tiny', secondary: true, type: 'warning', onClick: () => openGroupModal(row) }, () => '分组'),
      h(NButton, { size: 'tiny', secondary: true, type: 'primary', onClick: () => unfollow(row) }, () => '取消')
    ])
  }
]

const betterColumns = computed(() => [
  {
    title: () => renderBetterColumnHeader('排序', 'score'),
    key: 'recommendationRank',
    width: 160,
    render: (row) => renderBetterScoreCell(row)
  },
  {
    title: () => renderBetterColumnHeader('候选基金'),
    key: 'name',
    width: 250,
    render: (row) => h('div', { class: 'better-candidate-cell' }, [
      h(NButton, { text: true, type: 'primary', onClick: () => openExternal(row.code) }, () => row.name || row.code),
      h('div', { class: 'better-candidate-meta' }, [
        h('span', { class: 'cell-meta better-candidate-code' }, row.code),
        h(NTag, { size: 'small', bordered: false, type: 'success' }, { default: () => row.categoryLabel || row.fundType || '基金' })
      ])
    ])
  },
  {
    title: () => renderBetterColumnHeader('主要指标对比'),
    key: 'comparison',
    width: 560,
    render: (row) => renderBetterComparisonTable(row)
  },
  {
    title: () => renderBetterColumnHeader('官网同类排名'),
    key: 'peerRanks',
    width: 380,
    render: (row) => renderBetterOfficialRankTable(row)
  },
  {
    title: () => renderBetterColumnHeader('结论'),
    key: 'reasons',
    width: 220,
    render: (row) => renderBetterReasonCell(row)
  }
])

function renderBetterColumnHeader(title, helpKey = '') {
  return h('div', { class: 'better-column-header' }, [
    h('span', null, title),
    helpKey ? renderBetterHeaderHelp(helpKey) : null
  ])
}

function renderBetterHeaderHelp(helpKey) {
  return h(
    NTooltip,
    { placement: 'top-start', style: { maxWidth: '360px' } },
    {
      trigger: () => h(
        'span',
        { class: 'help-icon-trigger', title: '评分说明', 'aria-label': '评分说明' },
        [h(NIcon, { class: 'help-icon', size: 15 }, { default: () => h(AlertCircleOutline) })]
      ),
      default: () => buildBetterHelpContent(helpKey)
    }
  )
}

function buildBetterHelpContent(helpKey) {
  if (helpKey !== 'score') {
    return h('div', { class: 'score-tooltip' }, '当前列说明')
  }
  return h('div', { class: 'score-tooltip' }, [
    h('div', { class: 'score-tooltip-title' }, '得分说明'),
    h('div', { class: 'score-tooltip-line' }, '这是候选基金相对当前基金的加权优势分，不是绝对买入分。'),
    h('div', { class: 'score-tooltip-line' }, '近期 1-3 个月收益和回撤权重更高，同类排名、夏普和 Calmar 也会参与加分。'),
    h('div', { class: 'score-tooltip-line' }, '分段只用于当前参考基金和当前样本池的相对比较：'),
    h('div', { class: 'score-tooltip-line' }, '100 分以上：强烈优先关注'),
    h('div', { class: 'score-tooltip-line' }, '60-100 分：优先关注'),
    h('div', { class: 'score-tooltip-line' }, '25-60 分：可做备选'),
    h('div', { class: 'score-tooltip-line' }, '25 分以下：仅观察')
  ])
}

function renderBetterScoreCell(row) {
  return h('div', { class: 'better-order' }, [
    h('strong', { class: 'better-order-rank' }, `#${row.recommendationRank || '-'}`),
    h('span', { class: 'cell-meta' }, `得分 ${formatBetterScore(row.betterScore)}`)
  ])
}

function renderBetterMiniTable(title, columns, rows) {
  return h('div', { class: 'better-panel-card' }, [
    h('table', { class: 'better-mini-table' }, [
      h('colgroup', {}, columns.map((column) => h('col', {
        key: column.key,
        style: { width: column.width }
      }))),
      h('thead', {}, [
        h('tr', { class: 'better-mini-table-caption-row' }, [
          h('th', { colspan: columns.length, class: 'better-mini-table-caption' }, title)
        ]),
        h('tr', {}, columns.map((column) => h('th', {
          key: column.key,
          class: column.numeric ? 'is-number' : ''
        }, column.label)))
      ]),
      h('tbody', {}, rows)
    ])
  ])
}

function comparisonMetricKeys() {
  return ['growth1', 'drawdown1', 'growth3', 'drawdown3', 'growth6', 'growth12', 'drawdown6', 'drawdown12', 'sharpe12', 'calmar12']
}

function officialRankKeys() {
  return ['rank1m', 'rank3m', 'rank6m', 'rank12m']
}

function findBetterMetric(row, key) {
  return (row?.metrics || []).find((item) => item?.key === key) || null
}

function betterMetricLabel(key) {
  switch (key) {
    case 'growth7': return '近7天'
    case 'growth1': return '近1月'
    case 'drawdown1': return '1月最大回撤'
    case 'growth3': return '近3月'
    case 'growth6': return '近6月'
    case 'growth12': return '近1年'
    case 'drawdown3': return '3月最大回撤'
    case 'drawdown6': return '6月最大回撤'
    case 'drawdown12': return '近1年最大回撤'
    case 'volatility12': return '近1年波动'
    case 'sharpe12': return '近1年夏普'
    case 'calmar12': return 'Calmar'
    case 'rank1m': return '近1月'
    case 'rank3m': return '近3月'
    case 'rank6m': return '近6月'
    case 'rank12m': return '近1年'
    default: return key
  }
}

function isRatioMetric(key) {
  return key === 'sharpe12' || key === 'calmar12'
}

function isLowerBetterMetric(key) {
  return key === 'drawdown1' || key === 'drawdown3' || key === 'drawdown6' || key === 'drawdown12' || key === 'volatility12'
}

function formatMetricValue(value, key) {
  if (value === null || value === undefined) {
    return '-'
  }
  const num = Number(value)
  if (isRatioMetric(key)) {
    return num.toFixed(2)
  }
  if (isLowerBetterMetric(key)) {
    return `${num.toFixed(2)}%`
  }
  return `${num >= 0 ? '+' : ''}${num.toFixed(2)}%`
}

function formatBetterScore(value) {
  if (value === null || value === undefined) {
    return '-'
  }
  return Number(value).toFixed(2)
}

function metricTone(metric, key) {
  if (!metric) return ''
  if (metric.advantage != null && Number(metric.advantage) > 0) {
    return 'profit-text'
  }
  if (metric.delta == null) {
    return ''
  }
  const delta = Number(metric.delta)
  const better = isLowerBetterMetric(key) ? delta < 0 : delta > 0
  return better ? 'profit-text' : 'loss-text'
}

function getMetricPair(row, key) {
  const reference = betterResult.value?.reference || null
  if (!reference) return null

  switch (key) {
    case 'growth1':
      return { candidate: row?.netGrowth1, reference: reference.netGrowth1 }
    case 'drawdown1':
      return { candidate: row?.maxDrawdown1, reference: reference.maxDrawdown1 }
    case 'growth3':
      return { candidate: row?.netGrowth3, reference: reference.netGrowth3 }
    case 'growth6':
      return { candidate: row?.netGrowth6, reference: reference.netGrowth6 }
    case 'growth12':
      return { candidate: row?.netGrowth12, reference: reference.netGrowth12 }
    case 'drawdown3':
      return { candidate: row?.maxDrawdown3, reference: reference.maxDrawdown3 }
    case 'drawdown6':
      return { candidate: row?.maxDrawdown6, reference: reference.maxDrawdown6 }
    case 'drawdown12':
      return { candidate: row?.maxDrawdown12, reference: reference.maxDrawdown12 }
    case 'sharpe12':
      return { candidate: row?.sharpe12, reference: reference.sharpe12 }
    case 'calmar12':
      return { candidate: row?.calmar12, reference: reference.calmar12 }
    default:
      return null
  }
}

function comparePair(candidateValue, referenceValue, lowerBetter = false) {
  if (candidateValue === null || candidateValue === undefined || referenceValue === null || referenceValue === undefined) {
    return { icon: '—', label: '待补', tone: 'muted' }
  }
  const candidate = Number(candidateValue)
  const reference = Number(referenceValue)
  if (Math.abs(candidate - reference) < 0.0001) {
    return { icon: '→', label: '持平', tone: 'muted' }
  }
  const candidateBetter = lowerBetter ? candidate < reference : candidate > reference
  return candidateBetter
    ? { icon: '↑', label: '候选更优', tone: 'profit-text' }
    : { icon: '↓', label: '当前更优', tone: 'loss-text' }
}

function renderBetterComparisonTable(row) {
  const rows = comparisonMetricKeys().map((key) => {
    const pair = getMetricPair(row, key)
    const decision = comparePair(pair?.candidate, pair?.reference, isLowerBetterMetric(key))
    return h('tr', { key }, [
      h('td', { class: 'better-table-label' }, betterMetricLabel(key)),
      h('td', { class: ['is-number', metricTone(findBetterMetric(row, key), key)] }, pair ? formatMetricValue(pair.candidate, key) : '-'),
      h('td', { class: 'is-number' }, pair ? formatMetricValue(pair.reference, key) : '-'),
      h('td', { class: 'better-table-decision' }, [
        h('span', { class: ['better-direction', decision.tone] }, `${decision.icon} ${decision.label}`)
      ])
    ])
  })

  return renderBetterMiniTable('收益 / 回撤 / 风险质量', [
    { key: 'label', label: '指标', width: '31%' },
    { key: 'candidate', label: '候选', width: '17%', numeric: true },
    { key: 'reference', label: '当前', width: '17%', numeric: true },
    { key: 'decision', label: '判断', width: '35%' }
  ], rows)
}

function getOfficialRankPair(row, key) {
  const reference = betterResult.value?.reference || null
  if (!reference) return null
  switch (key) {
    case 'rank1m':
      return {
        candidateRank: Number(row?.stageRank1m || 0),
        candidateTotal: Number(row?.stageRank1mTotal || 0),
        referenceRank: Number(reference.stageRank1m || 0),
        referenceTotal: Number(reference.stageRank1mTotal || 0)
      }
    case 'rank3m':
      return {
        candidateRank: Number(row?.stageRank3m || 0),
        candidateTotal: Number(row?.stageRank3mTotal || 0),
        referenceRank: Number(reference.stageRank3m || 0),
        referenceTotal: Number(reference.stageRank3mTotal || 0)
      }
    case 'rank6m':
      return {
        candidateRank: Number(row?.stageRank6m || 0),
        candidateTotal: Number(row?.stageRank6mTotal || 0),
        referenceRank: Number(reference.stageRank6m || 0),
        referenceTotal: Number(reference.stageRank6mTotal || 0)
      }
    case 'rank12m':
      return {
        candidateRank: Number(row?.stageRank12m || 0),
        candidateTotal: Number(row?.stageRank12mTotal || 0),
        referenceRank: Number(reference.stageRank12m || 0),
        referenceTotal: Number(reference.stageRank12mTotal || 0)
      }
    default:
      return null
  }
}

function formatRankValue(rank, total) {
  if (!rank || !total) {
    return '-'
  }
  return `${rank}/${total}`
}

function compareRankPair(pair) {
  if (!pair || !pair.candidateRank || !pair.referenceRank) {
    return { icon: '—', label: '待补', tone: 'muted' }
  }
  if (pair.candidateRank === pair.referenceRank) {
    return { icon: '→', label: '持平', tone: 'muted' }
  }
  return pair.candidateRank < pair.referenceRank
    ? { icon: '↑', label: '候选更优', tone: 'profit-text' }
    : { icon: '↓', label: '当前更优', tone: 'loss-text' }
}

function renderBetterOfficialRankTable(row) {
  const rows = officialRankKeys().map((key) => {
    const pair = getOfficialRankPair(row, key)
    const decision = compareRankPair(pair)
    return h('tr', { key }, [
      h('td', { class: 'better-table-label' }, betterMetricLabel(key)),
      h('td', { class: 'is-number' }, formatRankValue(pair?.candidateRank, pair?.candidateTotal)),
      h('td', { class: 'is-number' }, formatRankValue(pair?.referenceRank, pair?.referenceTotal)),
      h('td', { class: 'better-table-decision' }, [
        h('span', { class: ['better-direction', decision.tone] }, `${decision.icon} ${decision.label}`)
      ])
    ])
  })

  return renderBetterMiniTable('东方财富官网同类排名', [
    { key: 'label', label: '周期', width: '22%' },
    { key: 'candidate', label: '候选', width: '22%', numeric: true },
    { key: 'reference', label: '当前', width: '22%', numeric: true },
    { key: 'decision', label: '判断', width: '34%' }
  ], rows)
}

function renderBetterReasonCell(row) {
  const highlights = (row.reasons || []).slice(0, 3)
  return h('div', { class: 'better-reason-card' }, [
    h('div', { class: 'better-reason-summary' }, row.reasonSummary || '综合表现更优'),
    h('div', { class: 'better-reason-meta muted' }, `${row.scopeLabel || '基金池对比'} · 综合分 ${formatBetterScore(row.betterScore)}`),
    highlights.length
      ? h('div', { class: 'better-reason-list' }, highlights.map((item, index) => h('div', { class: 'better-reason-item', key: `${row.code}-${index}` }, item)))
      : null
  ])
}

function renderPercent(value) {
  if (value === null || value === undefined) {
    return '-'
  }
  return h(NText, { type: Number(value) >= 0 ? 'error' : 'success' }, { default: () => signedPercent(value) + '%' })
}

function signedPercent(value) {
  const num = Number(value || 0)
  return (num >= 0 ? '+' : '') + num.toFixed(2)
}

async function loadData() {
  loading.value = true
  try {
    rows.value = (await GetFollowedFund()) || []
    checkedRowKeys.value = checkedRowKeys.value.filter((code) => rows.value.some((row) => row.code === code))
    loadRecommendationStatus(true)
  } catch (error) {
    console.error(error)
    message.error('基金自选加载失败')
  } finally {
    loading.value = false
  }
}

async function loadRecommendationStatus(autoStart = false) {
  const getter = window.go?.main?.App?.GetFundRecommendationRefreshStatus
  if (typeof getter !== 'function') {
    recommendationStatus.value = null
    return
  }
  try {
    recommendationStatus.value = await getter(Boolean(autoStart))
  } catch (error) {
    console.error(error)
  }
}

async function searchFunds(keyword) {
  const value = String(keyword || '').trim()
  if (!value) {
    fundOptions.value = []
    return
  }
  const result = (await GetfundList(value)) || []
  fundOptions.value = result.map((item) => ({
    label: item.name + ' [' + item.code + ']',
    value: item.code
  }))
}

function handleKeywordUpdate(value) {
  addKeyword.value = value
}

async function handleSelectFund(code) {
  try {
    const result = await FollowFund(code)
    if (result) {
      const group = activeGroupTab.value !== ALL_GROUP_TAB && activeGroupTab.value !== UNGROUPED_GROUP_TAB
        ? activeGroupTab.value
        : ''
      if (group && window.go?.main?.App?.UpdateFundWatchGroup) {
        await window.go.main.App.UpdateFundWatchGroup(code, group)
      }
      message.success(result)
      addKeyword.value = ''
      fundOptions.value = []
      showAddFundModal.value = false
      await loadData()
    }
  } catch (error) {
    console.error(error)
    message.error('加入自选失败')
  }
}

function openAddFundModal() {
  addKeyword.value = ''
  fundOptions.value = []
  showAddFundModal.value = true
}

async function unfollow(row) {
  try {
    const result = await UnFollowFund(row.code)
    if (result) {
      message.success(result)
      await loadData()
    }
  } catch (error) {
    console.error(error)
    message.error('取消自选失败')
  }
}

function handleCheckedRowKeys(keys) {
  const uniqueKeys = Array.from(new Set((keys || []).map((item) => String(item || '').trim()).filter(Boolean)))
  if (uniqueKeys.length > 10) {
    checkedRowKeys.value = uniqueKeys.slice(0, 10)
    message.warning('基金对比最多选择 10 只')
    return
  }
  checkedRowKeys.value = uniqueKeys
}

function loadStoredGroups() {
  try {
    const raw = window.localStorage?.getItem(CUSTOM_GROUP_STORAGE_KEY)
    if (!raw) return []
    const parsed = JSON.parse(raw)
    return Array.isArray(parsed) ? parsed.map((item) => String(item || '').trim()).filter(Boolean) : []
  } catch (error) {
    console.error(error)
    return []
  }
}

function saveStoredGroups(groups) {
  customGroups.value = Array.from(new Set(groups.map((item) => String(item || '').trim()).filter(Boolean)))
  window.localStorage?.setItem(CUSTOM_GROUP_STORAGE_KEY, JSON.stringify(customGroups.value))
}

function openCreateGroup() {
  groupModalMode.value = 'create'
  editingGroupName.value = ''
  showGroupModal.value = true
}

function openGroupModal(row) {
  groupModalMode.value = 'assign'
  groupingFund.value = row
  editingGroupSource.value = String(row.watchGroup || '').trim()
  editingGroupName.value = String(row.watchGroup || '').trim()
  showGroupModal.value = true
}

function openRenameGroup() {
  if (!activeCustomGroup.value) return
  groupModalMode.value = 'rename'
  editingGroupSource.value = activeCustomGroup.value
  editingGroupName.value = activeCustomGroup.value
  showGroupModal.value = true
}

function closeGroupModal() {
  showGroupModal.value = false
}

async function saveGroup() {
  const group = String(editingGroupName.value || '').trim()
  if (groupModalMode.value === 'create') {
    if (!group) {
      message.warning('请先输入页签名称')
      return
    }
    if (!isValidGroupName(group)) {
      message.warning('这个页签名称不可用，请换一个名字')
      return
    }
    if (hasGroupName(group)) {
      message.warning('页签名称已存在')
      return
    }
    saveStoredGroups([...customGroups.value, group])
    activeGroupTab.value = group
    closeGroupModal()
    message.success('自定义页签已创建')
    return
  }

  if (groupModalMode.value === 'rename') {
    const source = String(editingGroupSource.value || '').trim()
    if (!source) {
      message.error('当前页签不存在')
      return
    }
    if (!group) {
      message.warning('请先输入新的页签名称')
      return
    }
    if (!isValidGroupName(group)) {
      message.warning('这个页签名称不可用，请换一个名字')
      return
    }
    if (group === source) {
      closeGroupModal()
      return
    }
    if (hasGroupName(group, source)) {
      message.warning('页签名称已存在')
      return
    }
    try {
      const result = await renameGroup(source, group)
      saveStoredGroups([...customGroups.value.filter((item) => item !== source), group])
      activeGroupTab.value = group
      closeGroupModal()
      await loadData()
      message.success(result || '页签已重命名')
    } catch (error) {
      console.error(error)
      message.error('页签重命名失败')
    }
    return
  }

  if (group && !isValidGroupName(group)) {
    message.warning('这个页签名称不可用，请换一个名字')
    return
  }

  if (!window.go?.main?.App?.UpdateFundWatchGroup) {
    message.error('当前版本暂不支持基金分组')
    return
  }

  try {
    const result = await window.go.main.App.UpdateFundWatchGroup(groupingFund.value.code, group)
    if (group) {
      saveStoredGroups([...customGroups.value, group])
      activeGroupTab.value = group
    }
    closeGroupModal()
    await loadData()
    message.success(result || '分组已更新')
  } catch (error) {
    console.error(error)
    message.error('分组更新失败')
  }
}

function handleDeleteGroup() {
  const group = activeCustomGroup.value
  if (!group) return
  const fundCount = activeCustomGroupRows.value.length
  dialog.warning({
    title: '删除页签',
    content: fundCount
      ? `删除“${group}”后，页签会移除，这 ${fundCount} 只基金会进入“未分组”。确认继续吗？`
      : `删除“${group}”后，这个空页签会被移除。确认继续吗？`,
    positiveText: '删除页签',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        const result = await deleteGroup(group)
        saveStoredGroups(customGroups.value.filter((item) => item !== group))
        activeGroupTab.value = fundCount ? UNGROUPED_GROUP_TAB : ALL_GROUP_TAB
        await loadData()
        message.success(result || '页签已删除')
      } catch (error) {
        console.error(error)
        message.error('页签删除失败')
      }
    }
  })
}

function openDetail(row) {
  activeFund.value = normalizeWatchFund(row)
  showDetail.value = true
}

function openSingleAI(row) {
  activeFund.value = normalizeWatchFund(row)
  showSingleAI.value = true
}

function openBetter(row) {
  activeFund.value = normalizeWatchFund(row)
  betterReferenceCode.value = row.code
  sameTypeOnly.value = true
  betterSameSubTypeOnly.value = false
  betterFeeFree7.value = true
  betterFeeFree30.value = true
  betterIncludeAClass.value = false
  betterDimension.value = 'balanced'
  showBetterAI.value = false
  showBetterModal.value = true
  loadBetterFunds()
}

function openBetterAI() {
  if (!betterReferenceCode.value || !betterRows.value.length) {
    message.warning('当前栏目还没有可供分析的推荐基金')
    return
  }
  showBetterAI.value = true
}

async function loadBetterFunds(forceRefresh = false) {
  if (!betterReferenceCode.value) return
  betterLoading.value = true
  try {
    const result = await GetBetterFunds({
      referenceCode: betterReferenceCode.value,
      sameTypeOnly: sameTypeOnly.value,
      sameSubTypeOnly: betterSameSubTypeOnly.value,
      dimension: betterDimension.value,
      networkRefresh: Boolean(forceRefresh),
      feeFree7: betterFeeFree7.value,
      feeFree30: betterFeeFree30.value,
      includeAClass: betterIncludeAClass.value,
      onlyAClass: betterIncludeAClass.value,
      page: 1,
      pageSize: 12
    })
    betterResult.value = result || null
  } catch (error) {
    console.error(error)
    betterResult.value = null
    message.error('对比推荐筛选失败')
  } finally {
    betterLoading.value = false
  }
}

function normalizeWatchFund(row) {
  return {
    stockCode: row.code,
    stockName: row.name,
    trackingTarget: row.fundBasic?.trackingTarget || row.trackingTarget || '',
    fundType: row.fundBasic?.type,
    fundCompany: row.fundBasic?.company,
    fundManager: row.fundBasic?.manager,
    fundScale: row.fundBasic?.scale,
    fundRating: row.fundBasic?.rating,
    netEstimatedUnit: row.netEstimatedUnit,
    netEstimatedTime: row.netEstimatedUnitTime,
    netEstimatedRate: row.netEstimatedRate,
    netUnitValue: row.netUnitValue,
    netUnitValueDate: row.netUnitValueDate,
    netGrowth1: row.fundBasic?.netGrowth1,
    netGrowth3: row.fundBasic?.netGrowth3,
    netGrowth6: row.fundBasic?.netGrowth6,
    netGrowth12: row.fundBasic?.netGrowth12,
    netGrowth36: row.fundBasic?.netGrowth36,
    netGrowthYTD: row.fundBasic?.netGrowthYTD,
    currentPrice: row.netEstimatedUnit || row.netUnitValue || 0,
    estimateUpdated: Boolean(row.netEstimatedUnitTime),
    estimateStatus: row.netEstimatedUnitTime ? ('估算更新 ' + row.netEstimatedUnitTime) : '暂无盘中估算'
  }
}

function openExternal(code) {
  OpenURL('https://fund.eastmoney.com/' + code + '.html')
}

function isCustomGroupTab(value) {
  const group = String(value || '').trim()
  return Boolean(group) && group !== ALL_GROUP_TAB && group !== UNGROUPED_GROUP_TAB
}

function isValidGroupName(name) {
  return !RESERVED_GROUP_NAMES.includes(String(name || '').trim())
}

function hasGroupName(name, exclude = '') {
  return groupNames.value.some((item) => item === name && item !== exclude)
}

async function renameGroup(source, target) {
  if (window.go?.main?.App?.RenameFundWatchGroup) {
    return window.go.main.App.RenameFundWatchGroup(source, target)
  }
  if (!window.go?.main?.App?.UpdateFundWatchGroup) {
    throw new Error('rename_group_unavailable')
  }
  const members = rows.value.filter((row) => String(row.watchGroup || '').trim() === source)
  for (const row of members) {
    await window.go.main.App.UpdateFundWatchGroup(row.code, target)
  }
  return '页签已重命名'
}

async function deleteGroup(group) {
  if (window.go?.main?.App?.DeleteFundWatchGroup) {
    return window.go.main.App.DeleteFundWatchGroup(group)
  }
  if (!window.go?.main?.App?.UpdateFundWatchGroup) {
    throw new Error('delete_group_unavailable')
  }
  const members = rows.value.filter((row) => String(row.watchGroup || '').trim() === group)
  for (const row of members) {
    await window.go.main.App.UpdateFundWatchGroup(row.code, '')
  }
  return '页签已删除，组内基金已移到未分组'
}

function resetGroupModal() {
  groupingFund.value = null
  editingGroupName.value = ''
  groupModalMode.value = 'create'
  editingGroupSource.value = ''
}

let recommendationStatusTimer = 0

onMounted(() => {
  customGroups.value = loadStoredGroups()
  loadData()
  recommendationStatusTimer = window.setInterval(() => {
    if (recommendationStatus.value?.refreshing || Number(recommendationStatus.value?.pendingCount || 0) > 0) {
      loadRecommendationStatus(false)
    }
  }, 15000)
})

onUnmounted(() => {
  if (recommendationStatusTimer) {
    window.clearInterval(recommendationStatusTimer)
    recommendationStatusTimer = 0
  }
})

watch(sameTypeOnly, (checked) => {
  if (!checked) {
    betterSameSubTypeOnly.value = false
  }
})

watch(betterSubTypeDisabled, (disabled) => {
  if (disabled) {
    betterSameSubTypeOnly.value = false
  }
})

watch([sameTypeOnly, betterSameSubTypeOnly, betterFeeFree7, betterFeeFree30, betterIncludeAClass], () => {
  if (showBetterModal.value) {
    loadBetterFunds()
  }
})

watch(showBetterModal, (show) => {
  if (!show) {
    showBetterAI.value = false
  }
})

watch(showAddFundModal, (show) => {
  if (!show) {
    addKeyword.value = ''
    fundOptions.value = []
  }
})

watch(showGroupModal, (show) => {
  if (!show) resetGroupModal()
})

watch(watchTabs, (tabs) => {
  if (!tabs.some((tab) => tab.value === activeGroupTab.value)) {
    activeGroupTab.value = ALL_GROUP_TAB
  }
})
</script>

<style scoped>
.fund-watch-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.add-shell,
.watch-toolbar-card,
.recommendation-status-card,
.table-shell {
  padding: 16px;
}

.watch-toolbar {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.watch-search-bar {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.watch-search-main {
  display: flex;
  align-items: center;
  gap: 12px;
}

.watch-search-main :deep(.n-input) {
  flex: 1;
}

.watch-search-meta {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.table-toolbar {
  display: flex;
  justify-content: flex-start;
  align-items: center;
  margin-bottom: 16px;
}

.recommendation-status-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.recommendation-status-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.recommendation-status-title {
  font-size: 15px;
  font-weight: 700;
}

.recommendation-status-badge {
  padding: 4px 10px;
  border-radius: 999px;
  background: rgba(148, 163, 184, 0.14);
  color: var(--text-secondary);
  font-size: 12px;
}

.recommendation-status-badge.running {
  background: rgba(250, 204, 21, 0.16);
  color: #facc15;
}

.recommendation-status-badge.done {
  background: rgba(16, 185, 129, 0.16);
  color: #6ee7b7;
}

.recommendation-status-summary {
  font-weight: 600;
  line-height: 1.6;
}

.recommendation-status-hint {
  color: var(--text-secondary);
  line-height: 1.6;
  font-size: 13px;
}

.cell-main,
.cell-stack {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.cell-meta {
  font-size: 12px;
  color: var(--text-secondary);
}

.better-toolbar {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.better-toolbar-left,
.better-toolbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.better-filter-group {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
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

.better-toolbar {
  justify-content: space-between;
  margin-bottom: 12px;
}

.better-context {
  margin-bottom: 12px;
  padding: 14px 16px;
  border-radius: 16px;
  border: 1px solid rgba(97, 118, 148, 0.18);
  background: rgba(11, 19, 31, 0.5);
}

.better-context.warning {
  border-color: rgba(250, 204, 21, 0.28);
  background: rgba(120, 78, 15, 0.14);
}

.better-context-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.better-context-title {
  font-size: 15px;
  font-weight: 700;
}

.better-context-meta {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  color: var(--text-secondary);
  font-size: 12px;
}

.better-context-hint {
  margin-top: 8px;
  color: var(--text-secondary);
  line-height: 1.6;
}

.better-table-wrap {
  overflow-x: auto;
}

.better-column-header {
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 24px;
  font-size: 13px;
  font-weight: 700;
  color: #e8eef8;
  text-align: left;
}

.help-icon-trigger {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  cursor: help;
}

.help-icon {
  color: rgba(226, 232, 240, 0.88);
  transition: color 0.2s ease, opacity 0.2s ease;
}

.help-icon-trigger:hover .help-icon {
  color: #ffffff;
}

.better-order {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.better-order-rank {
  font-size: 20px;
  color: var(--primary-color);
}

.score-tooltip {
  display: flex;
  flex-direction: column;
  gap: 6px;
  line-height: 1.6;
}

.score-tooltip-title {
  font-weight: 700;
  color: #eef5ff;
}

.score-tooltip-line {
  color: rgba(226, 232, 240, 0.92);
  font-size: 12px;
}

.better-tag-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.better-candidate-cell {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.better-candidate-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.better-candidate-code {
  font-variant-numeric: tabular-nums;
}

.better-reason-card,
.better-panel-card {
  min-width: 0;
}

.better-result-table :deep(.n-data-table-th__title) {
  width: 100%;
  justify-content: flex-start;
  text-align: left;
}

.better-result-table :deep(.n-data-table-td) {
  vertical-align: top;
}

.better-reason-meta {
  font-size: 12px;
  line-height: 1.5;
}

.better-panel-card {
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.26);
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.015);
}

.better-mini-table {
  width: 100%;
  table-layout: fixed;
  border-collapse: collapse;
  font-size: 13px;
}

.better-mini-table th,
.better-mini-table td {
  padding: 7px 10px;
  text-align: left;
  border: 1px solid rgba(255, 255, 255, 0.2);
  white-space: nowrap;
  font-variant-numeric: tabular-nums;
}

.better-mini-table th {
  background: rgba(255, 255, 255, 0.045);
  color: #cbd5e1;
  font-size: 12px;
  font-weight: 700;
}

.better-mini-table-caption-row th {
  padding: 8px 10px;
  background: rgba(255, 255, 255, 0.07);
}

.better-mini-table-caption {
  text-align: left;
  color: #e6eef9;
  font-size: 13px;
  letter-spacing: 0.01em;
}

.better-mini-table .is-number {
  text-align: right;
}

.better-table-label {
  font-weight: 600;
  color: #e5edf8;
}

.better-table-decision {
  text-align: left;
}

.better-direction {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-weight: 700;
  white-space: nowrap;
}

.muted {
  color: var(--text-secondary);
}

.profit-text {
  color: #6ee7b7;
}

.loss-text {
  color: #fbbf24;
}

.better-reason-card {
  line-height: 1.6;
  white-space: normal;
  word-break: break-word;
}

.better-reason-summary {
  font-weight: 600;
}

.better-reason-meta {
  margin-top: 6px;
}

.better-reason-list {
  margin-top: 10px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.better-reason-item {
  padding: 8px 10px;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.04);
  color: var(--text-secondary);
  font-size: 12px;
  line-height: 1.5;
}

@media (max-width: 900px) {
  .watch-search-main {
    flex-direction: column;
    align-items: stretch;
  }

  .better-context-head {
    flex-direction: column;
    align-items: flex-start;
  }

  .better-mini-table th,
  .better-mini-table td {
    padding: 7px 8px;
    font-size: 12px;
  }
}
</style>
