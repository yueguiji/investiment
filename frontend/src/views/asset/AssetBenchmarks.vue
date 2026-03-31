<template>
  <div class="fade-in household-benchmarks">
    <n-page-header title="基准数据" subtitle="维护天津市、全国等地区口径，为家庭资产 AI 分析提供稳定参照">
      <template #extra>
        <n-space>
          <n-button @click="$router.push('/asset/overview')">返回资产总览</n-button>
          <n-button secondary @click="loadBenchmarks">刷新</n-button>
          <n-button type="primary" @click="openEditor()">新增基准</n-button>
        </n-space>
      </template>
    </n-page-header>

    <n-grid :cols="4" :x-gap="16" :y-gap="16" style="margin-top: 20px;">
      <n-gi v-for="card in summaryCards" :key="card.label">
        <div class="metric-card">
          <div class="metric-label">{{ card.label }}</div>
          <div class="metric-value">{{ card.value }}</div>
          <div class="metric-sub">{{ card.sub }}</div>
        </div>
      </n-gi>
    </n-grid>

    <div class="platform-card" style="margin-top: 20px;">
      <div class="section-title">筛选条件</div>
      <div class="section-sub">默认会拉取当前地区和全国两组基准，方便同时维护对比口径。</div>
      <n-grid :cols="4" :x-gap="16" style="margin-top: 16px;">
        <n-form-item label="当前地区">
          <n-input v-model:value="region" placeholder="例如：天津市" />
        </n-form-item>
        <n-form-item label="范围">
          <n-select v-model:value="scopeFilter" :options="scopeOptions" />
        </n-form-item>
        <n-form-item label="分类">
          <n-select v-model:value="categoryFilter" :options="categoryFilterOptions" />
        </n-form-item>
        <n-form-item label="关键词">
          <n-input v-model:value="keyword" placeholder="按名称或说明筛选" />
        </n-form-item>
      </n-grid>
    </div>

    <div class="platform-card" style="margin-top: 20px;">
      <div class="section-row">
        <div>
          <div class="section-title">基准数据列表</div>
          <div class="section-sub">这些记录会直接进入家庭资产 AI 分析上下文。</div>
        </div>
      </div>
      <n-data-table
        :columns="columns"
        :data="filteredBenchmarks"
        :pagination="{ pageSize: 10 }"
        :bordered="false"
      />
    </div>

    <n-modal v-model:show="showEditor" preset="card" :title="editingId ? '编辑基准数据' : '新增基准数据'" style="width: 720px;">
      <n-form label-placement="top">
        <n-grid :cols="2" :x-gap="16">
          <n-form-item label="名称">
            <n-input v-model:value="formState.name" />
          </n-form-item>
          <n-form-item label="地区">
            <n-input v-model:value="formState.region" placeholder="例如：天津市 / 全国" />
          </n-form-item>
          <n-form-item label="范围">
            <n-select v-model:value="formState.scope" :options="scopeOptions.slice(1)" />
          </n-form-item>
          <n-form-item label="分类">
            <n-select v-model:value="formState.category" :options="categoryOptions" />
          </n-form-item>
          <n-form-item label="数值">
            <n-input-number v-model:value="formState.value" :show-button="false" style="width: 100%;" />
          </n-form-item>
          <n-form-item label="单位">
            <n-input v-model:value="formState.unit" placeholder="例如：元/年" />
          </n-form-item>
          <n-form-item label="年份">
            <n-input-number v-model:value="formState.year" :show-button="false" style="width: 100%;" />
          </n-form-item>
          <n-form-item label="版本">
            <n-input v-model:value="formState.version" placeholder="例如：built-in-2026.03" />
          </n-form-item>
          <n-form-item label="说明" style="grid-column: span 2;">
            <n-input v-model:value="formState.description" type="textarea" :rows="3" />
          </n-form-item>
        </n-grid>
      </n-form>
      <template #action>
        <n-space justify="end">
          <n-button @click="showEditor = false">取消</n-button>
          <n-button type="primary" @click="saveBenchmark">保存</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup>
import { computed, h, onMounted, ref, watch } from 'vue'
import { NButton, NPopconfirm, useMessage } from 'naive-ui'

const message = useMessage()
const region = ref('天津市')
const scopeFilter = ref('all')
const categoryFilter = ref('all')
const keyword = ref('')
const benchmarks = ref([])
const showEditor = ref(false)
const editingId = ref(0)
const formState = ref(createForm())

const scopeOptions = [
  { label: '全部', value: 'all' },
  { label: '全国', value: '全国' },
  { label: '地区', value: '地区' }
]

const categoryOptions = [
  { label: '收入', value: 'income' },
  { label: '负债', value: 'debt' },
  { label: '资产', value: 'asset' },
  { label: '保障', value: 'protection' },
  { label: '其他', value: 'other' }
]

const categoryFilterOptions = [
  { label: '全部', value: 'all' },
  ...categoryOptions
]

const filteredBenchmarks = computed(() => benchmarks.value.filter((item) => {
  if (scopeFilter.value !== 'all' && item.scope !== scopeFilter.value) return false
  if (categoryFilter.value !== 'all' && item.category !== categoryFilter.value) return false
  if (!keyword.value) return true
  const haystack = `${item.name || ''} ${item.description || ''}`.toLowerCase()
  return haystack.includes(keyword.value.toLowerCase())
}))

const summaryCards = computed(() => {
  const localCount = benchmarks.value.filter((item) => item.region === region.value).length
  const nationalCount = benchmarks.value.filter((item) => item.region === '全国').length
  const latestYear = benchmarks.value.reduce((max, item) => Math.max(max, Number(item.year || 0)), 0)
  const versions = new Set(benchmarks.value.map((item) => item.version).filter(Boolean))
  return [
    {
      label: '当前地区记录',
      value: `${localCount}`,
      sub: `${region.value} 口径`
    },
    {
      label: '全国记录',
      value: `${nationalCount}`,
      sub: '全国对比口径'
    },
    {
      label: '最新年份',
      value: latestYear ? `${latestYear}` : '-',
      sub: '优先确认是否已更新到最新统计年'
    },
    {
      label: '版本数',
      value: `${versions.size}`,
      sub: '便于区分内置与手工维护数据'
    }
  ]
})

const columns = [
  { title: '名称', key: 'name', width: 180 },
  { title: '地区', key: 'region', width: 100 },
  { title: '范围', key: 'scope', width: 90 },
  { title: '分类', key: 'category', width: 90, render: (row) => categoryLabel(row.category) },
  { title: '数值', key: 'value', width: 150, render: (row) => `${Number(row.value || 0).toLocaleString('zh-CN')} ${row.unit || ''}` },
  { title: '年份', key: 'year', width: 80 },
  { title: '版本', key: 'version', width: 120 },
  { title: '说明', key: 'description' },
  {
    title: '操作',
    key: 'actions',
    width: 150,
    render: (row) => h('div', { style: { display: 'flex', gap: '8px' } }, [
      h(NButton, { size: 'tiny', quaternary: true, onClick: () => openEditor(row) }, () => '编辑'),
      h(
        NPopconfirm,
        { onPositiveClick: () => removeBenchmark(row.ID) },
        {
          default: () => '确认删除这条基准数据吗？',
          trigger: () => h(NButton, { size: 'tiny', quaternary: true, type: 'error' }, () => '删除')
        }
      )
    ])
  }
]

function createForm() {
  return {
    name: '',
    scope: '地区',
    region: region.value,
    category: 'income',
    value: 0,
    unit: '元/年',
    year: 2025,
    version: 'manual',
    description: ''
  }
}

function categoryLabel(value) {
  return {
    income: '收入',
    debt: '负债',
    asset: '资产',
    protection: '保障',
    other: '其他'
  }[value] || value || '-'
}

async function loadBenchmarks() {
  if (!window.go?.main?.App?.GetHouseholdBenchmarks) return
  benchmarks.value = (await window.go.main.App.GetHouseholdBenchmarks(region.value)) || []
}

async function loadProfileRegion() {
  if (!window.go?.main?.App?.GetHouseholdProfile) return
  const profile = await window.go.main.App.GetHouseholdProfile()
  if (profile?.region) {
    region.value = profile.region
  }
}

function openEditor(row = null) {
  editingId.value = row?.ID || 0
  formState.value = row
    ? {
      name: row.name || '',
      scope: row.scope || '地区',
      region: row.region || region.value,
      category: row.category || 'income',
      value: Number(row.value || 0),
      unit: row.unit || '元/年',
      year: Number(row.year || 2025),
      version: row.version || 'manual',
      description: row.description || ''
    }
    : createForm()
  showEditor.value = true
}

async function saveBenchmark() {
  if (!formState.value.name?.trim()) {
    message.warning('请先填写基准名称')
    return
  }
  if (!formState.value.region?.trim()) {
    message.warning('请先填写地区')
    return
  }
  try {
    await window.go.main.App.UpsertHouseholdBenchmark({
      ID: editingId.value || 0,
      ...formState.value,
      name: formState.value.name.trim(),
      region: formState.value.region.trim()
    })
    showEditor.value = false
    await loadBenchmarks()
    message.success('基准数据已保存')
  } catch (error) {
    console.error(error)
    message.error(error?.message || '保存失败')
  }
}

async function removeBenchmark(id) {
  try {
    await window.go.main.App.DeleteHouseholdBenchmark(id)
    await loadBenchmarks()
    message.success('基准数据已删除')
  } catch (error) {
    console.error(error)
    message.error(error?.message || '删除失败')
  }
}

watch(region, async () => {
  await loadBenchmarks()
})

onMounted(async () => {
  await loadProfileRegion()
  await loadBenchmarks()
})
</script>

<style scoped>
.household-benchmarks {
  max-width: 1280px;
  margin: 0 auto;
}

.metric-card {
  background: linear-gradient(180deg, rgba(15, 23, 42, 0.94), rgba(30, 41, 59, 0.86));
  border: 1px solid rgba(148, 163, 184, 0.16);
  border-radius: var(--radius-md);
  padding: 20px;
  min-height: 132px;
}

.metric-label {
  font-size: 13px;
  color: var(--text-secondary);
}

.metric-value {
  margin-top: 14px;
  font-size: 28px;
  font-weight: 700;
  font-family: var(--font-mono);
  color: #f8fafc;
}

.metric-sub {
  margin-top: 10px;
  color: var(--text-muted);
  font-size: 12px;
  line-height: 1.5;
}

.section-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 18px;
}

.section-title {
  font-size: 16px;
  font-weight: 700;
}

.section-sub {
  margin-top: 6px;
  color: var(--text-muted);
  font-size: 12px;
}
</style>
