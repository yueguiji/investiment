<template>
  <div class="fade-in asset-members-page">
    <n-page-header title="家庭成员" subtitle="统一维护家庭成员与家庭画像，供资产归属和数字分析联动使用">
      <template #extra>
        <n-space>
          <n-button @click="$router.push('/asset/detail')">资产台账</n-button>
          <n-button @click="$router.push('/asset/digital-analysis')">数字分析</n-button>
          <n-button secondary @click="loadAll">刷新</n-button>
          <n-button type="primary" @click="openCreateMember">新增成员</n-button>
        </n-space>
      </template>
    </n-page-header>

    <n-grid :cols="3" :x-gap="16" :y-gap="16" style="margin-top: 20px;">
      <n-gi v-for="card in statCards" :key="card.label">
        <div class="metric-card">
          <div class="metric-label">{{ card.label }}</div>
          <div class="metric-value">{{ card.value }}</div>
          <div class="metric-sub">{{ card.sub }}</div>
        </div>
      </n-gi>
    </n-grid>

    <div class="platform-card" style="margin-top: 20px;">
      <div class="section-title">家庭画像</div>
      <div class="section-sub">这里维护地区、家庭规模、风险偏好和年度支出，数字分析会直接使用这些信息。</div>

      <n-form label-placement="top" style="margin-top: 18px;">
        <n-grid :cols="3" :x-gap="16" :y-gap="8">
          <n-gi>
            <n-form-item label="家庭名称">
              <n-input v-model:value="profile.householdName" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="所在地区">
              <n-input v-model:value="profile.region" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="城市能级">
              <n-input v-model:value="profile.cityTier" />
            </n-form-item>
          </n-gi>

          <n-gi>
            <n-form-item label="住房状态">
              <n-input v-model:value="profile.housingStatus" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="风险偏好">
              <n-input v-model:value="profile.riskPreference" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="主要收入来源">
              <n-input v-model:value="profile.primaryIncomeSource" />
            </n-form-item>
          </n-gi>

          <n-gi>
            <n-form-item label="家庭月均支出">
              <n-input-number v-model:value="profile.monthlyHouseholdSpend" :show-button="false" :min="0" style="width: 100%;" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="年度家庭支出">
              <n-input-number v-model:value="profile.annualHouseholdSpend" :show-button="false" :min="0" style="width: 100%;" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="个人社保(月)">
              <n-input-number v-model:value="profile.monthlyPersonalInsuranceContribution" :show-button="false" :min="0" style="width: 100%;" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="个人公积金(月)">
              <n-input-number v-model:value="profile.monthlyHousingFundContribution" :show-button="false" :min="0" style="width: 100%;" />
            </n-form-item>
          </n-gi>

          <n-gi>
            <n-form-item label="其他税前扣除(月)">
              <n-input-number v-model:value="profile.monthlyOtherPretaxDeduction" :show-button="false" :min="0" style="width: 100%;" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="育儿抵扣(月)">
              <n-input-number v-model:value="profile.monthlyChildcareDeduction" :show-button="false" :min="0" style="width: 100%;" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="房贷抵扣(月)">
              <n-input-number v-model:value="profile.monthlyHousingLoanDeduction" :show-button="false" :min="0" style="width: 100%;" />
            </n-form-item>
          </n-gi>

          <n-gi>
            <n-form-item label="赡养老人抵扣(月)">
              <n-input-number v-model:value="profile.monthlyElderlyCareDeduction" :show-button="false" :min="0" style="width: 100%;" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="其他专项附加抵扣(月)">
              <n-input-number v-model:value="profile.monthlyOtherSpecialDeduction" :show-button="false" :min="0" style="width: 100%;" />
            </n-form-item>
          </n-gi>
          <n-gi :span="3">
            <n-form-item label="备注">
              <n-input v-model:value="profile.notes" type="textarea" :rows="3" />
            </n-form-item>
          </n-gi>
        </n-grid>
      </n-form>

      <div class="action-row">
        <n-button type="primary" @click="saveProfile">保存家庭画像</n-button>
      </div>
    </div>

    <div class="platform-card" style="margin-top: 20px;">
      <div class="table-header">
        <div>
          <div class="section-title">成员列表</div>
          <div class="section-sub">收入、负债、固定资产和保障里的所属人会统一从这里选择。</div>
        </div>
      </div>

      <n-data-table :columns="columns" :data="members" :pagination="{ pageSize: 8 }" :bordered="false" />
    </div>

    <n-modal v-model:show="showEditor" preset="card" :title="editingId ? '编辑成员' : '新增成员'" style="width: 720px;">
      <n-form label-placement="top">
        <n-grid :cols="2" :x-gap="16" :y-gap="8">
          <n-gi>
            <n-form-item label="姓名">
              <n-input v-model:value="memberForm.name" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="关系">
              <n-input v-model:value="memberForm.relationship" placeholder="本人 / 配偶 / 子女 / 父母" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="性别">
              <n-select v-model:value="memberForm.gender" :options="genderOptions" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="出生日期">
              <n-date-picker v-model:value="memberForm.birthDate" type="date" clearable style="width: 100%;" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="职业">
              <n-input v-model:value="memberForm.occupation" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="所在城市">
              <n-input v-model:value="memberForm.city" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="年收入">
              <n-input-number v-model:value="memberForm.annualIncome" :show-button="false" :min="0" style="width: 100%;" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="启用">
              <n-switch v-model:value="memberForm.isActive" />
            </n-form-item>
          </n-gi>
          <n-gi :span="2">
            <n-form-item label="备注">
              <n-input v-model:value="memberForm.notes" type="textarea" :rows="3" />
            </n-form-item>
          </n-gi>
        </n-grid>
      </n-form>
      <template #action>
        <n-space justify="end">
          <n-button @click="showEditor = false">取消</n-button>
          <n-button type="primary" @click="saveMember">保存</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup>
import { computed, h, onMounted, ref } from 'vue'
import { NButton, NPopconfirm, NTag, useMessage } from 'naive-ui'

const message = useMessage()

const members = ref([])
const profile = ref(defaultProfile())
const showEditor = ref(false)
const editingId = ref(null)
const memberForm = ref(defaultMemberForm())

const genderOptions = [
  { label: '未设置', value: '' },
  { label: '男', value: 'male' },
  { label: '女', value: 'female' }
]

const statCards = computed(() => [
  { label: '成员人数', value: members.value.length, sub: '统一维护收入、负债和资产的所属人' },
  { label: '所在地区', value: profile.value.region || '未设置', sub: profile.value.householdName || '我的家庭' },
  { label: '住房状态', value: profile.value.housingStatus || '未填写', sub: profile.value.riskPreference || '未填写风险偏好' }
])

const columns = [
  { title: '姓名', key: 'name', width: 120 },
  { title: '关系', key: 'relationship', width: 100 },
  { title: '职业', key: 'occupation', width: 120, render: (row) => row.occupation || '-' },
  { title: '城市', key: 'city', width: 120, render: (row) => row.city || '-' },
  { title: '年收入', key: 'annualIncome', width: 130, render: (row) => formatMoney(row.annualIncome) },
  {
    title: '标签',
    key: 'tags',
    width: 100,
    render: (row) =>
      h('div', { style: { display: 'flex', gap: '8px', flexWrap: 'wrap' } }, [
        row.isActive ? h(NTag, { size: 'small', type: 'info', round: true }, () => '启用') : h(NTag, { size: 'small', round: true }, () => '停用')
      ].filter(Boolean))
  },
  {
    title: '操作',
    key: 'actions',
    width: 140,
    render: (row) => h('div', { style: { display: 'flex', gap: '8px' } }, [
      h(NButton, { size: 'tiny', quaternary: true, onClick: () => openEditMember(row) }, () => '编辑'),
      h(
        NPopconfirm,
        { onPositiveClick: () => deleteMember(row.ID) },
        {
          default: () => '删除后会清空关联台账里的所属人，确认继续吗？',
          trigger: () => h(NButton, { size: 'tiny', quaternary: true, type: 'error' }, () => '删除')
        }
      )
    ])
  }
]

function defaultProfile() {
  return {
    householdName: '我的家庭',
    region: '天津市',
    cityTier: '新一线',
    membersCount: 0,
    housingStatus: '',
    riskPreference: '',
    monthlyHouseholdSpend: 0,
    annualHouseholdSpend: 0,
    primaryIncomeSource: '',
    monthlyPersonalInsuranceContribution: 0,
    monthlyHousingFundContribution: 0,
    monthlyOtherPretaxDeduction: 0,
    monthlyChildcareDeduction: 0,
    monthlyHousingLoanDeduction: 0,
    monthlyElderlyCareDeduction: 0,
    monthlyOtherSpecialDeduction: 0,
    notes: ''
  }
}

function defaultMemberForm() {
  return {
    name: '',
    relationship: '本人',
    gender: '',
    birthDate: null,
    occupation: '',
    city: '',
    annualIncome: 0,
    notes: '',
    isActive: true
  }
}

function formatMoney(value) {
  return `¥ ${Number(value || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

function toTimestamp(value) {
  return value ? new Date(value).getTime() : null
}

function toISOStringOrNull(value) {
  return value ? new Date(value).toISOString() : null
}

async function callApp(name, ...args) {
  const fn = window.go?.main?.App?.[name]
  if (!fn) throw new Error(`missing app method: ${name}`)
  return fn(...args)
}

async function loadAll() {
  profile.value = { ...defaultProfile(), ...((await callApp('GetHouseholdProfile')) || {}) }
  members.value = (await callApp('GetHouseholdMembers')) || []
  profile.value.membersCount = members.value.length
}

function openCreateMember() {
  editingId.value = null
  memberForm.value = defaultMemberForm()
  showEditor.value = true
}

function openEditMember(row) {
  editingId.value = row.ID
  memberForm.value = {
    ...defaultMemberForm(),
    ...row,
    birthDate: toTimestamp(row.birthDate)
  }
  showEditor.value = true
}

async function saveProfile() {
  try {
    const payload = {
      ...profile.value,
      annualHouseholdSpend: Number(profile.value.annualHouseholdSpend || 0) || Number(profile.value.monthlyHouseholdSpend || 0) * 12,
      membersCount: members.value.length
    }
    await callApp('UpsertHouseholdProfile', payload)
    await loadAll()
    message.success('家庭画像已保存')
  } catch (error) {
    console.error(error)
    message.error('家庭画像保存失败')
  }
}

async function saveMember() {
  const name = String(memberForm.value.name || '').trim()
  if (!name) {
    message.warning('请先填写成员姓名')
    return
  }
  try {
    const payload = {
      ...memberForm.value,
      birthDate: toISOStringOrNull(memberForm.value.birthDate)
    }
    if (editingId.value) {
      payload.ID = editingId.value
      await callApp('UpdateHouseholdMember', payload)
    } else {
      await callApp('CreateHouseholdMember', payload)
    }
    showEditor.value = false
    await loadAll()
    message.success('成员信息已保存')
  } catch (error) {
    console.error(error)
    message.error('成员信息保存失败')
  }
}

async function deleteMember(id) {
  try {
    await callApp('DeleteHouseholdMember', id)
    await loadAll()
    message.success('成员已删除')
  } catch (error) {
    console.error(error)
    message.error('成员删除失败')
  }
}

onMounted(loadAll)
</script>

<style scoped>
.asset-members-page {
  max-width: 1280px;
  margin: 0 auto;
}

.metric-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 18px;
}

.metric-label {
  font-size: 12px;
  color: var(--text-secondary);
}

.metric-value {
  margin-top: 10px;
  font-size: 28px;
  font-weight: 700;
}

.metric-sub,
.section-sub {
  margin-top: 8px;
  font-size: 12px;
  line-height: 1.6;
  color: var(--text-muted);
}

.section-title {
  font-size: 16px;
  font-weight: 700;
}

.action-row {
  display: flex;
  justify-content: flex-end;
  margin-top: 8px;
}

.table-header {
  margin-bottom: 16px;
}
</style>
