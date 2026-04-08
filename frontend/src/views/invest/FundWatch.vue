<template>
  <div class="fade-in fund-watch-page">
    <n-page-header title="基金自选" subtitle="跟踪关注基金的估算、阶段收益，并支持对比推荐和 AI 解读。">
      <template #extra>
        <n-space>
          <n-button :disabled="checkedRowKeys.length < 2" @click="showCompareModal = true">
            基金对比{{ checkedRowKeys.length ? ` (${checkedRowKeys.length})` : '' }}
          </n-button>
          <n-button @click="showCollectionAI = true">AI 分析关注池</n-button>
          <n-button :loading="loading" @click="loadData">刷新列表</n-button>
        </n-space>
      </template>
    </n-page-header>

    <div class="platform-card add-shell">
      <n-grid :cols="24" :x-gap="12">
        <n-gi :span="10">
          <n-auto-complete
            v-model:value="addKeyword"
            :options="fundOptions"
            placeholder="输入基金代码或名称，加入自选"
            clearable
            @search="searchFunds"
            @update:value="handleKeywordUpdate"
            @select="handleSelectFund"
          />
        </n-gi>
        <n-gi :span="8">
          <n-text depth="3">支持代码、名称模糊检索，加入后可直接做 AI 分析和对比推荐。</n-text>
        </n-gi>
      </n-grid>
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
        <n-space>
          <n-button v-if="activeCustomGroup" size="small" @click="openRenameGroup">重命名</n-button>
          <n-button v-if="activeCustomGroup" size="small" type="error" ghost @click="handleDeleteGroup">删除页签</n-button>
          <n-button size="small" @click="openCreateGroup">新建页签</n-button>
        </n-space>
      </div>
    </div>

    <div class="platform-card table-shell">
      <n-data-table
        :loading="loading"
        :columns="columns"
        :data="filteredRows"
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
      scope="watchlist"
      title="基金自选 AI 分析"
    />

    <FundCompareModal
      v-model:show="showCompareModal"
      :codes="checkedRowKeys"
    />

    <n-modal
      v-model:show="showGroupModal"
      preset="card"
      :title="groupModalTitle"
      style="width: 420px;"
    >
      <n-space vertical size="large">
        <n-input
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
        <n-checkbox v-model:checked="sameTypeOnly">只看同类型</n-checkbox>
        <n-button :loading="betterLoading" @click="loadBetterFunds">重新筛选</n-button>
      </div>
      <n-tabs v-model:value="betterDimension" type="segment" animated @update:value="loadBetterFunds">
        <n-tab name="lower_drawdown" tab="回撤更低" />
        <n-tab name="higher_return" tab="收益更高" />
        <n-tab name="balanced" tab="实力均衡更优" />
      </n-tabs>
      <div class="better-table-wrap">
        <n-data-table
          :loading="betterLoading"
          :columns="betterColumns"
          :data="betterRows"
          :pagination="{ pageSize: 8 }"
          :scroll-x="1360"
        />
      </div>
    </n-modal>
  </div>
</template>

<script setup>
import { computed, h, onMounted, ref, watch } from 'vue'
import { NButton, NTag, NText, useDialog, useMessage } from 'naive-ui'
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
const activeFund = ref(null)
const showDetail = ref(false)
const showSingleAI = ref(false)
const showCollectionAI = ref(false)
const showBetterModal = ref(false)
const betterLoading = ref(false)
const sameTypeOnly = ref(false)
const betterDimension = ref('balanced')
const betterRows = ref([])
const checkedRowKeys = ref([])
const showCompareModal = ref(false)
const activeGroupTab = ref('__all__')
const customGroups = ref([])
const showGroupModal = ref(false)
const groupingFund = ref(null)
const editingGroupName = ref('')
const groupModalMode = ref('create')
const editingGroupSource = ref('')

const betterReferenceCode = ref('')
const ALL_GROUP_TAB = '__all__'
const UNGROUPED_GROUP_TAB = '__ungrouped__'
const CUSTOM_GROUP_STORAGE_KEY = 'fund-watch-custom-groups'
const RESERVED_GROUP_NAMES = [ALL_GROUP_TAB, UNGROUPED_GROUP_TAB, '全部', '未分组']

const groupNames = computed(() => Array.from(new Set([
  ...customGroups.value,
  ...rows.value.map((row) => String(row.watchGroup || '').trim()).filter(Boolean)
])))

const watchTabs = computed(() => {
  const counts = new Map()
  let ungroupedCount = 0
  for (const row of rows.value) {
    const group = String(row.watchGroup || '').trim()
    if (group) {
      counts.set(group, (counts.get(group) || 0) + 1)
    } else {
      ungroupedCount += 1
    }
  }

  return [
    { label: `全部 (${rows.value.length})`, value: ALL_GROUP_TAB },
    ...groupNames.value.map((group) => ({ label: `${group} (${counts.get(group) || 0})`, value: group })),
    { label: `未分组 (${ungroupedCount})`, value: UNGROUPED_GROUP_TAB }
  ]
})

const filteredRows = computed(() => {
  if (activeGroupTab.value === ALL_GROUP_TAB) return rows.value
  if (activeGroupTab.value === UNGROUPED_GROUP_TAB) {
    return rows.value.filter((row) => !String(row.watchGroup || '').trim())
  }
  return rows.value.filter((row) => String(row.watchGroup || '').trim() === activeGroupTab.value)
})

const betterTitle = computed(() => {
  const name = activeFund.value?.name || activeFund.value?.stockName || activeFund.value?.code || activeFund.value?.stockCode || '当前基金'
  return name + ' 的对比推荐'
})
const activeCustomGroup = computed(() => isCustomGroupTab(activeGroupTab.value) ? activeGroupTab.value : '')
const activeCustomGroupRows = computed(() => rows.value.filter((row) => String(row.watchGroup || '').trim() === activeCustomGroup.value))
const groupModalTitle = computed(() => {
  if (groupModalMode.value === 'assign') return '设置基金分组'
  if (groupModalMode.value === 'rename') return '重命名页签'
  return '新建自定义页签'
})
const groupModalHint = computed(() => {
  if (groupModalMode.value === 'assign') return '留空表示移出自定义页签，进入“未分组”。'
  if (groupModalMode.value === 'rename') return '重命名后，这个页签下的基金会一起迁移到新名称。'
  return '先建一个分组页签，后面可以把基金归进去。'
})
const groupModalActionText = computed(() => {
  if (groupModalMode.value === 'rename') return '重命名'
  if (groupModalMode.value === 'create') return '创建'
  return '保存'
})

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
    key: 'estimate',
    width: 180,
    render: (row) => h('div', { class: 'cell-stack' }, [
      h(
        NText,
        { type: Number(row.netEstimatedRate || 0) >= 0 ? 'error' : 'success' },
        {
          default: () => (
            row.netEstimatedUnit
              ? Number(row.netEstimatedUnit).toFixed(4) + ' / ' + signedPercent(row.netEstimatedRate) + '%'
              : '暂无估算'
          )
        }
      ),
      h('div', { class: 'cell-meta' }, row.netEstimatedUnitTime || row.netUnitValueDate || '-')
    ])
  },
  {
    title: '近1月',
    key: 'netGrowth1',
    width: 100,
    render: (row) => renderPercent(row.fundBasic?.netGrowth1)
  },
  {
    title: '近3月',
    key: 'netGrowth3',
    width: 100,
    render: (row) => renderPercent(row.fundBasic?.netGrowth3)
  },
  {
    title: '近6月',
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

const betterColumns = [
  {
    title: '候选基金',
    key: 'name',
    width: 220,
    render: (row) => h('div', { class: 'cell-main' }, [
      h(NButton, { text: true, type: 'primary', onClick: () => openExternal(row.code) }, () => row.name || row.code),
      h('div', { class: 'cell-meta' }, row.code)
    ])
  },
  {
    title: '类型',
    key: 'fundType',
    width: 180,
    render: (row) => h('div', { class: 'cell-stack' }, [
      h(NTag, { size: 'small', bordered: false, type: 'success' }, { default: () => row.categoryLabel || row.fundType || '基金' }),
      h('div', { class: 'cell-meta' }, row.fundType || '-')
    ])
  },
  {
    title: '近7天',
    key: 'netGrowth7',
    width: 90,
    render: (row) => renderPercent(row.netGrowth7)
  },
  {
    title: '近1月',
    key: 'netGrowth1',
    width: 90,
    render: (row) => renderPercent(row.netGrowth1)
  },
  {
    title: '近3月',
    key: 'netGrowth3',
    width: 90,
    render: (row) => renderPercent(row.netGrowth3)
  },
  {
    title: '近6月',
    key: 'netGrowth6',
    width: 90,
    render: (row) => renderPercent(row.netGrowth6)
  },
  {
    title: '回撤',
    key: 'maxDrawdown12',
    width: 90,
    render: (row) => row.maxDrawdown12 == null ? '-' : Number(row.maxDrawdown12).toFixed(2) + '%'
  },
  {
    title: '更优原因',
    key: 'reasons',
    width: 320,
    render: (row) => h('div', { class: 'better-reasons' }, (row.reasons || []).join(' / ') || '-')
  }
]

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
  } catch (error) {
    console.error(error)
    message.error('基金自选加载失败')
  } finally {
    loading.value = false
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
      await loadData()
    }
  } catch (error) {
    console.error(error)
    message.error('加入自选失败')
  }
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
  sameTypeOnly.value = false
  betterDimension.value = 'balanced'
  showBetterModal.value = true
  loadBetterFunds()
}

async function loadBetterFunds() {
  if (!betterReferenceCode.value) return
  betterLoading.value = true
  try {
    const result = await GetBetterFunds({
      referenceCode: betterReferenceCode.value,
      sameTypeOnly: sameTypeOnly.value,
      dimension: betterDimension.value,
      page: 1,
      pageSize: 12
    })
    betterRows.value = result?.candidates || []
  } catch (error) {
    console.error(error)
    message.error('对比推荐筛选失败')
  } finally {
    betterLoading.value = false
  }
}

function normalizeWatchFund(row) {
  return {
    stockCode: row.code,
    stockName: row.name,
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

onMounted(() => {
  customGroups.value = loadStoredGroups()
  loadData()
})

watch(sameTypeOnly, () => {
  if (showBetterModal.value) {
    loadBetterFunds()
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
.table-shell {
  padding: 16px;
}

.watch-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
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

.better-table-wrap {
  overflow-x: auto;
}

.better-reasons {
  min-width: 260px;
  line-height: 1.6;
  white-space: normal;
  word-break: break-word;
}
</style>
