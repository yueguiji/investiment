<template>
  <div class="fade-in household-detail">
    <n-page-header title="资产明细" subtitle="统一管理家庭账户、固定资产、收入、保障与负债计划">
      <template #extra>
        <n-space>
          <n-button @click="$router.push('/asset/members')">家庭成员</n-button>
          <n-button @click="$router.push('/asset/debt-plans')">负债计划页</n-button>
          <n-button @click="$router.push('/asset/benchmarks')">基准数据页</n-button>
          <n-button secondary @click="loadAll">刷新</n-button>
          <n-button type="primary" @click="startCreate">新增当前台账</n-button>
        </n-space>
      </template>
    </n-page-header>

    <n-grid :cols="5" :x-gap="16" :y-gap="16" style="margin: 20px 0;">
      <n-gi v-for="card in statCards" :key="card.label">
        <div class="metric-card">
          <div class="metric-label">{{ card.label }}</div>
          <div class="metric-value">{{ card.value }}</div>
          <div class="metric-sub">{{ card.sub }}</div>
        </div>
      </n-gi>
    </n-grid>

    <div class="platform-card">
      <div class="section-row">
        <div>
          <div class="section-title">家庭月均支出</div>
          <div class="section-sub">这里的支出会同步写入家庭画像，并直接带入数字分析助手，帮助判断现金流、结余和安全垫。</div>
        </div>
      </div>
      <n-grid :cols="3" :x-gap="16" :y-gap="12">
        <n-gi>
          <n-form-item label="家庭月均支出">
            <n-input-number v-model:value="profile.monthlyHouseholdSpend" :show-button="false" :min="0" :step="100" style="width: 100%;" />
          </n-form-item>
        </n-gi>
        <n-gi>
          <n-form-item label="折算年支出">
            <n-input-number :value="annualSpendPreview" :show-button="false" disabled style="width: 100%;" />
          </n-form-item>
        </n-gi>
        <n-gi>
          <n-form-item label="预计月结余">
            <n-input-number :value="monthlyFreeCashflow" :show-button="false" disabled style="width: 100%;" />
          </n-form-item>
        </n-gi>
      </n-grid>
      <div class="toolbar-actions">
        <n-button type="primary" @click="saveHouseholdSpend">保存月均支出</n-button>
      </div>
    </div>

    <n-grid :cols="2" :x-gap="16" :y-gap="16" style="margin-top: 20px;">
      <n-gi>
        <div class="platform-card chart-card">
          <div class="section-title">流动资金趋势图</div>
          <div class="section-sub">基于历史家庭快照，观察流动资金余额随时间的变化。</div>
          <div v-if="liquidTrend.length > 1" class="chart-shell">
            <div ref="liquidTrendChartRef" class="chart-canvas"></div>
          </div>
          <div v-else class="empty-text">录入并更新资产后，这里会逐步形成趋势。</div>
        </div>
      </n-gi>
      <n-gi>
        <div class="platform-card chart-card">
          <div class="section-title">流动资金分布饼图</div>
          <div class="section-sub">按每个流动资金细项拆分，查看资金集中度。</div>
          <div v-if="liquidDistribution.length" class="chart-shell">
            <div ref="liquidDistributionChartRef" class="chart-canvas"></div>
          </div>
          <div v-else class="empty-text">当前还没有可计入流动资产的账户明细。</div>
        </div>
      </n-gi>
    </n-grid>

    <div class="platform-card" style="margin-top: 20px;">
      <n-tabs v-model:value="activeTab" type="line" animated>
        <n-tab-pane v-for="tab in tabs" :key="tab.key" :name="tab.key" :tab="tab.label">
          <div class="tab-toolbar">
            <div>
              <div class="section-title">{{ tab.heading }}</div>
              <div class="section-sub">{{ tab.description }}</div>
            </div>
            <n-space>
              <n-button @click="loadTab(tab.key)">刷新当前页</n-button>
              <n-button type="primary" @click="startCreate">新增{{ tab.shortLabel }}</n-button>
            </n-space>
          </div>

          <n-data-table :columns="tableColumns" :data="tableData" :pagination="{ pageSize: 8 }" :bordered="false" />
        </n-tab-pane>
      </n-tabs>
    </div>

    <n-modal v-model:show="showEditor" preset="card" :title="editorMode === 'create' ? `新增${currentTab.shortLabel}` : `编辑${currentTab.shortLabel}`" style="width: 820px;">
      <n-form label-placement="top">
        <n-grid :cols="2" :x-gap="16">
          <n-gi v-for="field in currentTab.fields" :key="field.key" :span="field.span || 1">
            <n-form-item :label="field.label">
              <n-input v-if="field.type === 'text'" v-model:value="formState[field.key]" :placeholder="field.placeholder || ''" />
              <n-input v-else-if="field.type === 'textarea'" v-model:value="formState[field.key]" type="textarea" :rows="3" :placeholder="field.placeholder || ''" />
              <n-input-number v-else-if="field.type === 'number'" v-model:value="formState[field.key]" style="width: 100%;" :show-button="false" :min="field.min" :step="field.step || 1" />
              <n-select v-else-if="field.type === 'select'" v-model:value="formState[field.key]" :options="field.options" />
              <n-select v-else-if="field.type === 'member-select'" v-model:value="formState[field.key]" :options="memberOptions" clearable filterable placeholder="请选择家庭成员" />
              <n-switch v-else-if="field.type === 'switch'" v-model:value="formState[field.key]" />
              <n-date-picker v-else-if="field.type === 'date'" v-model:value="formState[field.key]" type="date" clearable style="width: 100%;" />
            </n-form-item>
          </n-gi>
        </n-grid>
      </n-form>

      <n-alert v-if="formHint" type="info" :show-icon="false" style="margin-top: 8px;">
        {{ formHint }}
      </n-alert>

      <template #action>
        <n-space justify="space-between">
          <n-text depth="3">{{ currentTab.validationTip }}</n-text>
          <n-space>
            <n-button @click="showEditor = false">取消</n-button>
            <n-button type="primary" @click="handleSave">保存</n-button>
          </n-space>
        </n-space>
      </template>
    </n-modal>

    <n-modal v-model:show="showSchedule" preset="card" title="负债还款计划" style="width: 880px;">
      <div class="section-sub" style="margin-bottom: 16px;">{{ scheduleTitle }}</div>
      <n-data-table :columns="scheduleColumns" :data="scheduleRows" :pagination="{ pageSize: 10 }" :bordered="false" />
    </n-modal>
  </div>
</template>

<script setup>
import { computed, h, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NButton, NPopconfirm, NTag, NText, useMessage } from 'naive-ui'
import * as echarts from 'echarts'

const route = useRoute()
const router = useRouter()
const message = useMessage()

const validTabs = ['accounts', 'fixedAssets', 'incomes', 'protections', 'liabilities']
const activeTab = ref(validTabs.includes(route.query.tab) ? route.query.tab : 'accounts')
const showEditor = ref(false)
const showSchedule = ref(false)
const editorMode = ref('create')
const editingId = ref(null)
const formState = ref({})
const scheduleRows = ref([])
const scheduleTitle = ref('')
const members = ref([])
const profile = ref(defaultProfile())
const summary = ref({ monthlyNetIncome: 0, monthlyDebtPayment: 0, monthlyEffectiveDebtPayment: 0 })
const liquidTrend = ref([])
const liquidDistribution = ref([])
const liquidTrendChartRef = ref(null)
const liquidDistributionChartRef = ref(null)

let liquidTrendChartInstance = null
let liquidDistributionChartInstance = null

const datasets = ref({
  accounts: [],
  fixedAssets: [],
  incomes: [],
  protections: [],
  liabilities: []
})

const tabs = [
  {
    key: 'accounts',
    label: '家庭账户',
    shortLabel: '账户',
    heading: '流动账户与数字钱包',
    description: '适合银行卡、支付宝、微信、抖音、现金以及其他自定义流动账户。',
    validationTip: '至少填写名称和账户类型；余额不能为负数。',
    listMethod: 'GetHouseholdAccounts',
    createMethod: 'CreateHouseholdAccount',
    updateMethod: 'UpdateHouseholdAccount',
    deleteMethod: 'DeleteHouseholdAccount',
    emptyForm: { name: '', accountType: 'bank', provider: '', owner: '', balance: 0, isLiquid: true, isActive: true, remark: '' },
    fields: [
      { key: 'name', label: '名称', type: 'text' },
      { key: 'accountType', label: '账户类型', type: 'select', options: accountTypeOptions() },
      { key: 'provider', label: '机构/平台', type: 'text' },
      { key: 'owner', label: '归属人', type: 'member-select' },
      { key: 'balance', label: '当前余额', type: 'number', min: 0, step: 100 },
      { key: 'isLiquid', label: '计入流动资产', type: 'switch' },
      { key: 'isActive', label: '启用', type: 'switch' },
      { key: 'remark', label: '备注', type: 'textarea', span: 2 }
    ],
    columns: [
      textColumn('名称', 'name', 170),
      tagColumn('账户类型', 'accountType', accountTypeLabel),
      textColumn('机构/平台', 'provider', 140),
      moneyColumn('当前余额', 'balance'),
      booleanColumn('流动资产', 'isLiquid'),
      textColumn('归属人', 'owner', 120),
      textColumn('备注', 'remark')
    ]
  },
  {
    key: 'fixedAssets',
    label: '固定资产',
    shortLabel: '固定资产',
    heading: '房产、车位、车辆等长期资产',
    description: '支持记录位置、估值日期、持有比例等固定资产信息。',
    validationTip: '估值不能为负数；持有比例会自动校正到 0~1。',
    listMethod: 'GetHouseholdFixedAssets',
    createMethod: 'CreateHouseholdFixedAsset',
    updateMethod: 'UpdateHouseholdFixedAsset',
    deleteMethod: 'DeleteHouseholdFixedAsset',
    emptyForm: { name: '', assetType: 'property', owner: '', ownershipRatio: 1, currentValue: 0, costBasis: 0, location: '', referenceCode: '', purchasedAt: null, valuationDate: null, isActive: true, remark: '' },
    fields: [
      { key: 'name', label: '资产名称', type: 'text' },
      { key: 'assetType', label: '资产类型', type: 'select', options: fixedAssetTypeOptions() },
      { key: 'owner', label: '归属人', type: 'member-select' },
      { key: 'ownershipRatio', label: '持有比例', type: 'number', min: 0, step: 0.1 },
      { key: 'currentValue', label: '当前估值', type: 'number', min: 0, step: 1000 },
      { key: 'costBasis', label: '购置成本', type: 'number', min: 0, step: 1000 },
      { key: 'location', label: '位置/地址', type: 'text' },
      { key: 'referenceCode', label: '资产编号', type: 'text' },
      { key: 'purchasedAt', label: '购置日期', type: 'date' },
      { key: 'valuationDate', label: '估值日期', type: 'date' },
      { key: 'isActive', label: '启用', type: 'switch' },
      { key: 'remark', label: '备注', type: 'textarea', span: 2 }
    ],
    columns: [
      textColumn('资产名称', 'name', 180),
      tagColumn('类型', 'assetType', fixedAssetTypeLabel),
      moneyColumn('当前估值', 'currentValue'),
      moneyColumn('购置成本', 'costBasis'),
      ratioColumn('持有比例', 'ownershipRatio'),
      textColumn('归属人', 'owner', 100),
      dateColumn('估值日期', 'valuationDate', 120)
    ]
  },
  {
    key: 'incomes',
    label: '收入来源',
    shortLabel: '收入',
    heading: '工资与非工资收入',
    description: '统一管理月度和年度收入；税后收入测算会结合家庭画像里的社保、公积金和专项附加扣除参数。',
    validationTip: '月度与年度金额至少填一个，系统会自动补齐另一项。',
    listMethod: 'GetHouseholdIncomes',
    createMethod: 'CreateHouseholdIncome',
    updateMethod: 'UpdateHouseholdIncome',
    deleteMethod: 'DeleteHouseholdIncome',
    emptyForm: { name: '', incomeType: 'salary', owner: '', employer: '', frequency: 'monthly', monthlyAmount: 0, annualAmount: 0, monthlyPersonalInsuranceContribution: 0, monthlyEmployerInsuranceContribution: 0, monthlyPersonalHousingFundContribution: 0, monthlyEmployerHousingFundContribution: 0, isActive: true, remark: '' },
    fields: [
      { key: 'name', label: '收入名称', type: 'text' },
      { key: 'incomeType', label: '收入类型', type: 'select', options: incomeTypeOptions() },
      { key: 'owner', label: '归属人', type: 'member-select' },
      { key: 'employer', label: '单位/发放主体', type: 'text' },
      { key: 'frequency', label: '频率', type: 'select', options: incomeFrequencyOptions() },
      { key: 'monthlyAmount', label: '月度金额', type: 'number', min: 0, step: 100 },
      { key: 'annualAmount', label: '年度金额', type: 'number', min: 0, step: 1000 },
      { key: 'monthlyPersonalInsuranceContribution', label: '个人社保', type: 'number', min: 0, step: 10 },
      { key: 'monthlyEmployerInsuranceContribution', label: '单位社保', type: 'number', min: 0, step: 10 },
      { key: 'monthlyPersonalHousingFundContribution', label: '个人公积金', type: 'number', min: 0, step: 10 },
      { key: 'monthlyEmployerHousingFundContribution', label: '单位公积金', type: 'number', min: 0, step: 10 },
      { key: 'isActive', label: '启用', type: 'switch' },
      { key: 'remark', label: '备注', type: 'textarea', span: 2 }
    ],
    columns: [
      textColumn('收入名称', 'name', 180),
      tagColumn('类型', 'incomeType', incomeTypeLabel),
      textColumn('归属人', 'owner', 100),
      textColumn('单位', 'employer', 120),
      textColumn('频率', 'frequency', 90, (row) => frequencyLabel(row.frequency)),
      moneyColumn('月度金额', 'monthlyAmount'),
      moneyColumn('五险一金', 'monthlyPersonalInsuranceContribution', (row) => Number(row.monthlyPersonalInsuranceContribution || 0) + Number(row.monthlyEmployerInsuranceContribution || 0) + Number(row.monthlyPersonalHousingFundContribution || 0) + Number(row.monthlyEmployerHousingFundContribution || 0)),
      moneyColumn('年度金额', 'annualAmount'),
      booleanColumn('启用', 'isActive')
    ]
  },
  {
    key: 'protections',
    label: '保障福利',
    shortLabel: '保障',
    heading: '五险一金、企业年金、商业保险',
    description: '沉淀保障与福利类信息，后续 AI 会基于它判断保障结构与缺口。',
    validationTip: '名称和类型建议必填；保费、保障额度都允许按需补录。',
    listMethod: 'GetHouseholdProtections',
    createMethod: 'CreateHouseholdProtection',
    updateMethod: 'UpdateHouseholdProtection',
    deleteMethod: 'DeleteHouseholdProtection',
    emptyForm: { name: '', protectionType: 'social_insurance', owner: '', provider: '', employer: '', currentBalance: 0, monthlyPersonalContribution: 0, monthlyEmployerContribution: 0, monthlyPremium: 0, coverageAmount: 0, nextDueDate: null, isActive: true, remark: '' },
    fields: [
      { key: 'name', label: '名称', type: 'text' },
      { key: 'protectionType', label: '类型', type: 'select', options: protectionTypeOptions() },
      { key: 'owner', label: '归属人', type: 'member-select' },
      { key: 'provider', label: '机构', type: 'text' },
      { key: 'employer', label: '单位', type: 'text' },
      { key: 'currentBalance', label: '当前余额', type: 'number', min: 0, step: 100 },
      { key: 'monthlyPersonalContribution', label: '个人月缴', type: 'number', min: 0, step: 10 },
      { key: 'monthlyEmployerContribution', label: '单位月缴', type: 'number', min: 0, step: 10 },
      { key: 'monthlyPremium', label: '月保费', type: 'number', min: 0, step: 10 },
      { key: 'coverageAmount', label: '保障额度', type: 'number', min: 0, step: 1000 },
      { key: 'nextDueDate', label: '下次到期日', type: 'date' },
      { key: 'isActive', label: '启用', type: 'switch' },
      { key: 'remark', label: '备注', type: 'textarea', span: 2 }
    ],
    columns: [
      textColumn('名称', 'name', 170),
      tagColumn('类型', 'protectionType', protectionTypeLabel),
      textColumn('归属人', 'owner', 100),
      moneyColumn('当前余额', 'currentBalance'),
      moneyColumn('月度成本/储存', 'monthlyPremium', (row) => Number(row.monthlyPremium || 0) + Number(row.monthlyPersonalContribution || 0) + Number(row.monthlyEmployerContribution || 0)),
      moneyColumn('保障额度', 'coverageAmount'),
      dateColumn('下次到期', 'nextDueDate', 120)
    ]
  },
  {
    key: 'liabilities',
    label: '负债计划',
    shortLabel: '负债',
    heading: '房贷、车贷、消费贷等还款计划',
    description: '保存贷款核心参数并自动生成还款计划，剩余本金会随日期自动演进。',
    validationTip: '名称、本金、剩余本金、利率、期限和还款方式建议一次录全，月供可自动估算。',
    listMethod: 'GetHouseholdLiabilities',
    createMethod: 'CreateHouseholdLiability',
    updateMethod: 'UpdateHouseholdLiability',
    deleteMethod: 'DeleteHouseholdLiability',
    extraActions: liabilityActions,
    emptyForm: { name: '', liabilityType: 'mortgage', lender: '', owner: '', principal: 0, outstandingPrincipal: 0, annualRate: 0, loanTermMonths: 360, repaymentMethod: 'equal_installment', monthlyPayment: 0, extraMonthlyPayment: 0, startDate: null, firstPaymentDate: null, autoAmortize: true, isActive: true, remark: '' },
    fields: [
      { key: 'name', label: '名称', type: 'text' },
      { key: 'liabilityType', label: '负债类型', type: 'select', options: liabilityTypeOptions() },
      { key: 'lender', label: '贷款机构', type: 'text' },
      { key: 'owner', label: '归属人', type: 'member-select' },
      { key: 'principal', label: '贷款本金', type: 'number', min: 0, step: 1000 },
      { key: 'outstandingPrincipal', label: '当前剩余本金', type: 'number', min: 0, step: 1000 },
      { key: 'annualRate', label: '年利率(%)', type: 'number', min: 0, step: 0.01 },
      { key: 'loanTermMonths', label: '期限(月)', type: 'number', min: 1, step: 1 },
      { key: 'repaymentMethod', label: '还款方式', type: 'select', options: repaymentMethodOptions() },
      { key: 'monthlyPayment', label: '月供', type: 'number', min: 0, step: 10 },
      { key: 'extraMonthlyPayment', label: '额外还款', type: 'number', min: 0, step: 10 },
      { key: 'startDate', label: '起始日期', type: 'date' },
      { key: 'firstPaymentDate', label: '首期还款日', type: 'date' },
      { key: 'autoAmortize', label: '自动摊还', type: 'switch' },
      { key: 'isActive', label: '启用', type: 'switch' },
      { key: 'remark', label: '备注', type: 'textarea', span: 2 }
    ],
    columns: [
      textColumn('名称', 'name', 160),
      tagColumn('类型', 'liabilityType', liabilityTypeLabel),
      textColumn('机构', 'lender', 120),
      moneyColumn('贷款本金', 'principal'),
      moneyColumn('剩余本金', 'outstandingPrincipal'),
      textColumn('利率/期限', 'annualRate', 120, (row) => `${Number(row.annualRate || 0).toFixed(2)}% / ${row.loanTermMonths || 0}月`),
      moneyColumn('月供', 'monthlyPayment', (row) => Number(row.monthlyPayment || 0) + Number(row.extraMonthlyPayment || 0))
    ]
  }
]

const currentTab = computed(() => tabs.find((item) => item.key === activeTab.value) || tabs[0])
const tableData = computed(() => datasets.value[activeTab.value] || [])
const tableColumns = computed(() => [...currentTab.value.columns, actionColumn()])
const memberOptions = computed(() => members.value.map((item) => ({ label: item.name, value: item.name })))
const annualSpendPreview = computed(() => Number((Number(profile.value.monthlyHouseholdSpend || 0) * 12).toFixed(2)))
const monthlyFreeCashflow = computed(() => Number((Number(summary.value.monthlyNetIncome || 0) - Number(summary.value.monthlyEffectiveDebtPayment || summary.value.monthlyDebtPayment || 0) - Number(profile.value.monthlyHouseholdSpend || 0)).toFixed(2)))
const statCards = computed(() => [
  { label: '家庭账户', value: datasets.value.accounts.length, sub: '银行卡、钱包、现金、自定义账户' },
  { label: '固定资产', value: datasets.value.fixedAssets.length, sub: '房产、车位、车辆及其他长期资产' },
  { label: '收入来源', value: datasets.value.incomes.length, sub: '工资与非工资收入' },
  { label: '保障福利', value: datasets.value.protections.length, sub: '五险一金、企业年金、商业保险' },
  { label: '负债计划', value: datasets.value.liabilities.length, sub: '房贷、车贷与其他贷款计划' }
])

const formHint = computed(() => {
  if (currentTab.value.key !== 'liabilities') return ''
  const principal = Number(formState.value.principal || 0)
  const rate = Number(formState.value.annualRate || 0)
  const term = Number(formState.value.loanTermMonths || 0)
  const outstanding = Number(formState.value.outstandingPrincipal || 0)
  if (!principal || !term) return '填写本金、利率和期限后，系统会自动帮你补齐剩余本金和参考月供。'
  const monthlyRate = rate / 100 / 12
  let estimate = principal / term
  if ((formState.value.repaymentMethod || 'equal_installment') === 'equal_installment' && monthlyRate > 0) {
    const factor = Math.pow(1 + monthlyRate, term)
    estimate = (principal * monthlyRate * factor) / (factor - 1)
  }
  return `当前参考月供约 ${formatMoney(estimate)}，剩余本金 ${formatMoney(outstanding || principal)}。`
})

const scheduleColumns = [
  { title: '期次', key: 'periodNumber', width: 70 },
  { title: '还款日', key: 'dueDate', width: 120, render: (row) => formatDate(row.dueDate) },
  { title: '期初本金', key: 'openingPrincipal', width: 130, render: (row) => formatMoney(row.openingPrincipal) },
  { title: '本金', key: 'principalPaid', width: 120, render: (row) => formatMoney(row.principalPaid) },
  { title: '利息', key: 'interestPaid', width: 120, render: (row) => formatMoney(row.interestPaid) },
  { title: '本期还款', key: 'paymentAmount', width: 130, render: (row) => formatMoney(row.paymentAmount) },
  { title: '期末本金', key: 'closingPrincipal', width: 130, render: (row) => formatMoney(row.closingPrincipal) }
]

watch(activeTab, (value) => {
  router.replace({ path: route.path, query: { ...route.query, tab: value } })
})

function defaultProfile() {
  return {
    householdName: '我的家庭',
    region: '天津市',
    monthlyHouseholdSpend: 0,
    annualHouseholdSpend: 0
  }
}

function accountTypeOptions() { return [{ label: '银行卡', value: 'bank' }, { label: '支付宝', value: 'alipay' }, { label: '微信', value: 'wechat' }, { label: '抖音', value: 'douyin' }, { label: '现金', value: 'cash' }, { label: '自定义', value: 'custom' }] }
function fixedAssetTypeOptions() { return [{ label: '房产', value: 'property' }, { label: '车位', value: 'parking' }, { label: '车辆', value: 'vehicle' }, { label: '其他固定资产', value: 'other' }] }
function incomeTypeOptions() { return [{ label: '工资', value: 'salary' }, { label: '奖金', value: 'bonus' }, { label: '经营收入', value: 'business' }, { label: '其他收入', value: 'other' }] }
function incomeFrequencyOptions() { return [{ label: '月度', value: 'monthly' }, { label: '年度', value: 'annual' }] }
function protectionTypeOptions() { return [{ label: '五险', value: 'social_insurance' }, { label: '公积金', value: 'housing_fund' }, { label: '企业年金', value: 'enterprise_annuity' }, { label: '商业保险', value: 'commercial_insurance' }] }
function liabilityTypeOptions() { return [{ label: '房贷', value: 'mortgage' }, { label: '车贷', value: 'car_loan' }, { label: '消费贷', value: 'consumer_loan' }, { label: '其他负债', value: 'other' }] }
function repaymentMethodOptions() { return [{ label: '等额本息', value: 'equal_installment' }, { label: '等额本金', value: 'equal_principal' }] }

function textColumn(title, key, width = 140, renderFn = null) { return { title, key, width, render: (row) => (renderFn ? renderFn(row) : (row[key] || '-')) } }
function moneyColumn(title, key, renderValue = null) { return { title, key, width: 140, render: (row) => formatMoney(renderValue ? renderValue(row) : row[key]) } }
function dateColumn(title, key, width = 110) { return { title, key, width, render: (row) => formatDate(row[key]) } }
function ratioColumn(title, key) { return { title, key, width: 100, render: (row) => `${Number(row[key] || 0).toFixed(2)}` } }
function booleanColumn(title, key) { return { title, key, width: 90, render: (row) => h(NTag, { size: 'small', type: row[key] ? 'success' : 'default', round: true }, () => (row[key] ? '是' : '否')) } }
function tagColumn(title, key, labelGetter) { return { title, key, width: 120, render: (row) => h(NTag, { size: 'small', type: 'info', round: true }, () => labelGetter(row[key])) } }

function actionColumn() {
  return {
    title: '操作',
    key: 'actions',
    width: currentTab.value.key === 'liabilities' ? 220 : 140,
    render: (row) => h('div', { style: { display: 'flex', gap: '8px', flexWrap: 'wrap' } }, [
      ...(currentTab.value.extraActions ? currentTab.value.extraActions(row) : []),
      h(NButton, { size: 'tiny', quaternary: true, onClick: () => startEdit(row) }, () => '编辑'),
      h(NPopconfirm, { onPositiveClick: () => handleDelete(row.ID) }, { default: () => '确认删除这条记录吗？', trigger: () => h(NButton, { size: 'tiny', quaternary: true, type: 'error' }, () => '删除') })
    ])
  }
}

function liabilityActions(row) { return [h(NButton, { size: 'tiny', quaternary: true, type: 'primary', onClick: () => openSchedule(row) }, () => '还款计划')] }
function accountTypeLabel(value) { return ({ bank: '银行卡', alipay: '支付宝', wechat: '微信', douyin: '抖音', cash: '现金', custom: '自定义' })[value] || value || '-' }
function fixedAssetTypeLabel(value) { return ({ property: '房产', parking: '车位', vehicle: '车辆', other: '其他固定资产' })[value] || value || '-' }
function incomeTypeLabel(value) { return ({ salary: '工资', bonus: '奖金', business: '经营收入', other: '其他收入' })[value] || value || '-' }
function frequencyLabel(value) { return ({ monthly: '月度', annual: '年度' })[value] || value || '-' }
function protectionTypeLabel(value) { return ({ social_insurance: '五险', housing_fund: '公积金', enterprise_annuity: '企业年金', commercial_insurance: '商业保险' })[value] || value || '-' }
function liabilityTypeLabel(value) { return ({ mortgage: '房贷', car_loan: '车贷', consumer_loan: '消费贷', other: '其他负债' })[value] || value || '-' }
function cloneForm(form) { return JSON.parse(JSON.stringify(form)) }
function formatMoney(value) { return `¥ ${Number(value || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}` }
function formatDate(value) { return value ? new Date(value).toLocaleDateString('zh-CN') : '-' }
function toTimestamp(value) { return value ? new Date(value).getTime() : null }
function toISOStringOrNull(value) { return value ? new Date(value).toISOString() : null }
function normalizeRowForEdit(row, fields) { const next = { ...row }; for (const field of fields) { if (field.type === 'date') next[field.key] = toTimestamp(row[field.key]) } return next }
function serializeForm(form, fields) { const next = { ...form }; for (const field of fields) { if (field.type === 'date') next[field.key] = toISOStringOrNull(form[field.key]) } return next }

function normalizePayloadByTab(tabKey, payload) {
  const next = { ...payload }
  if (tabKey === 'fixedAssets') next.ownershipRatio = Math.min(1, Math.max(0, Number(next.ownershipRatio || 0)))
  if (tabKey === 'incomes') {
    const monthly = Number(next.monthlyAmount || 0)
    const annual = Number(next.annualAmount || 0)
    if (!monthly && annual) next.monthlyAmount = Number((annual / 12).toFixed(2))
    if (!annual && monthly) next.annualAmount = Number((monthly * 12).toFixed(2))
  }
  if (tabKey === 'liabilities') {
    const principal = Number(next.principal || 0)
    const outstanding = Number(next.outstandingPrincipal || 0)
    const term = Number(next.loanTermMonths || 0)
    const annualRate = Number(next.annualRate || 0)
    next.outstandingPrincipal = outstanding > 0 ? Math.min(outstanding, principal || outstanding) : principal
    if (!next.monthlyPayment && principal > 0 && term > 0) {
      const monthlyRate = annualRate / 100 / 12
      next.monthlyPayment = next.repaymentMethod === 'equal_installment' && monthlyRate > 0
        ? Number(((principal * monthlyRate * Math.pow(1 + monthlyRate, term)) / (Math.pow(1 + monthlyRate, term) - 1)).toFixed(2))
        : Number((principal / term).toFixed(2))
    }
  }
  return next
}

function validatePayload(tabKey, payload) {
  if (!String(payload.name || '').trim()) return '请先填写名称'
  if (tabKey === 'accounts' && Number(payload.balance || 0) < 0) return '账户余额不能小于 0'
  if (tabKey === 'fixedAssets' && (Number(payload.currentValue || 0) < 0 || Number(payload.ownershipRatio || 0) < 0 || Number(payload.ownershipRatio || 0) > 1)) return '请检查固定资产估值和持有比例'
  if (tabKey === 'incomes' && Number(payload.monthlyAmount || 0) <= 0 && Number(payload.annualAmount || 0) <= 0) return '月度金额和年度金额至少填写一个'
  if (tabKey === 'liabilities') {
    if (Number(payload.principal || 0) <= 0) return '贷款本金必须大于 0'
    if (Number(payload.loanTermMonths || 0) <= 0) return '贷款期限必须大于 0'
    if (Number(payload.outstandingPrincipal || 0) < 0) return '剩余本金不能小于 0'
    if (Number(payload.outstandingPrincipal || 0) > Number(payload.principal || 0)) return '剩余本金不能大于贷款本金'
  }
  return ''
}

function ensureChart(container, currentInstance) {
  if (!container) return null
  return currentInstance || echarts.init(container)
}

function renderLiquidTrendChart() {
  if (!liquidTrendChartRef.value || liquidTrend.value.length <= 1) return
  liquidTrendChartInstance = ensureChart(liquidTrendChartRef.value, liquidTrendChartInstance)
  liquidTrendChartInstance.setOption({
    animation: false,
    grid: { left: 56, right: 56, top: 28, bottom: 48 },
    tooltip: { trigger: 'axis' },
    legend: { top: 0, textStyle: { color: '#cbd5e1' } },
    xAxis: { type: 'category', data: liquidTrend.value.map((item) => item.date || '-'), boundaryGap: false, axisLabel: { color: '#94a3b8' } },
    yAxis: [{ type: 'value', axisLabel: { color: '#94a3b8', formatter: (value) => `${(Number(value) / 10000).toFixed(0)}万` } }, { type: 'value', axisLabel: { color: '#94a3b8', formatter: (value) => `${(Number(value) / 1000).toFixed(0)}k` } }],
    series: [
      { name: '流动资金', type: 'line', smooth: true, showSymbol: false, data: liquidTrend.value.map((item) => Number(item.totalLiquidAssets || 0)), lineStyle: { width: 3, color: '#14b8a6' } },
      { name: '月税后收入', type: 'bar', yAxisIndex: 1, barWidth: 12, data: liquidTrend.value.map((item) => Number(item.monthlyNetIncome || 0)), itemStyle: { color: 'rgba(96, 165, 250, 0.72)' } },
      { name: '月供', type: 'line', yAxisIndex: 1, smooth: true, showSymbol: false, data: liquidTrend.value.map((item) => Number(item.monthlyDebtPayment || 0)), lineStyle: { width: 2, color: '#f59e0b' } }
    ]
  })
}

function renderLiquidDistributionChart() {
  if (!liquidDistributionChartRef.value || !liquidDistribution.value.length) return
  liquidDistributionChartInstance = ensureChart(liquidDistributionChartRef.value, liquidDistributionChartInstance)
  liquidDistributionChartInstance.setOption({
    animation: false,
    tooltip: { trigger: 'item' },
    legend: { bottom: 0, textStyle: { color: '#cbd5e1' } },
    series: [{ type: 'pie', radius: ['38%', '70%'], center: ['50%', '45%'], data: liquidDistribution.value.map((item) => ({ name: item.name, value: Number(item.balance || 0) })) }]
  })
}

async function callAppMethod(name, ...args) {
  const fn = window.go?.main?.App?.[name]
  if (!fn) throw new Error(`missing app method: ${name}`)
  return fn(...args)
}

async function triggerAutoAnalysis(triggerSource) {
  try {
    const aiConfigs = (await callAppMethod('GetAiConfigs')) || []
    const promptTemplates = (await callAppMethod('GetPromptTemplates', '', '模型系统Prompt')) || []
    const promptTemplate = promptTemplates.find((item) => item.name === '家庭资产分析-标准模板') || promptTemplates[0]
    await callAppMethod('RunHouseholdAIAnalysis', profile.value.region || '天津市', Number(aiConfigs[0]?.ID || 0), Number(promptTemplate?.ID || 0), triggerSource)
  } catch (error) {
    console.error('auto household analysis failed', error)
  }
}

async function loadLiquidData() {
  summary.value = (await callAppMethod('GetHouseholdDashboardSummary')) || summary.value
  profile.value = { ...profile.value, ...((await callAppMethod('GetHouseholdProfile')) || {}) }
  liquidTrend.value = (await callAppMethod('GetHouseholdLiquidAssetTrend', 180)) || []
  liquidDistribution.value = (await callAppMethod('GetHouseholdLiquidAssetDistribution')) || []
  await nextTick()
  renderLiquidTrendChart()
  renderLiquidDistributionChart()
}

async function loadTab(tabKey) {
  const tab = tabs.find((item) => item.key === tabKey)
  if (!tab) return
  datasets.value[tabKey] = (await callAppMethod(tab.listMethod)) || []
}

async function loadAll() {
  members.value = (await callAppMethod('GetHouseholdMembers')) || []
  for (const tab of tabs) await loadTab(tab.key)
  await loadLiquidData()
}

function startCreate() {
  editorMode.value = 'create'
  editingId.value = null
  formState.value = cloneForm(currentTab.value.emptyForm)
  showEditor.value = true
}

function startEdit(row) {
  editorMode.value = 'edit'
  editingId.value = row.ID
  formState.value = normalizeRowForEdit(row, currentTab.value.fields)
  showEditor.value = true
}

async function saveHouseholdSpend() {
  try {
    await callAppMethod('UpsertHouseholdProfile', { ...profile.value, monthlyHouseholdSpend: Number(profile.value.monthlyHouseholdSpend || 0), annualHouseholdSpend: annualSpendPreview.value, membersCount: members.value.length || Number(profile.value.membersCount || 0) || 1 })
    await loadLiquidData()
    triggerAutoAnalysis('asset-detail:monthly-spend')
    message.success('家庭月均支出已保存')
  } catch (error) {
    console.error(error)
    message.error('保存家庭月均支出失败')
  }
}

async function handleSave() {
  try {
    let payload = serializeForm(formState.value, currentTab.value.fields)
    payload = normalizePayloadByTab(currentTab.value.key, payload)
    const validationMessage = validatePayload(currentTab.value.key, payload)
    if (validationMessage) return message.warning(validationMessage)
    if (editorMode.value === 'edit') {
      payload.ID = editingId.value
      await callAppMethod(currentTab.value.updateMethod, payload)
    } else {
      await callAppMethod(currentTab.value.createMethod, payload)
    }
    showEditor.value = false
    await loadAll()
    triggerAutoAnalysis(`asset-detail:${currentTab.value.key}:${editorMode.value}`)
    message.success('保存成功')
  } catch (error) {
    console.error(error)
    message.error('保存失败，请检查输入后重试')
  }
}

async function handleDelete(id) {
  try {
    await callAppMethod(currentTab.value.deleteMethod, id)
    await loadAll()
    triggerAutoAnalysis(`asset-detail:delete:${currentTab.value.key}`)
    message.success('删除成功')
  } catch (error) {
    console.error(error)
    message.error('删除失败')
  }
}

async function openSchedule(row) {
  try {
    scheduleTitle.value = `${row.name || '负债'} - 自动摊还计划`
    scheduleRows.value = (await callAppMethod('GetHouseholdLiabilitySchedules', row.ID)) || []
    showSchedule.value = true
  } catch (error) {
    console.error(error)
    message.error('加载还款计划失败')
  }
}

function handleResize() {
  liquidTrendChartInstance?.resize()
  liquidDistributionChartInstance?.resize()
}

watch(liquidTrend, async () => { await nextTick(); renderLiquidTrendChart() }, { deep: true })
watch(liquidDistribution, async () => { await nextTick(); renderLiquidDistributionChart() }, { deep: true })

onMounted(async () => {
  window.addEventListener('resize', handleResize)
  await loadAll()
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  liquidTrendChartInstance?.dispose()
  liquidDistributionChartInstance?.dispose()
})
</script>

<style scoped>
.household-detail { max-width: 1280px; margin: 0 auto; }
.metric-card { background: var(--bg-card); border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 18px; }
.metric-label { font-size: 13px; color: var(--text-secondary); }
.metric-value { margin-top: 10px; font-size: 28px; font-weight: 700; }
.metric-sub { margin-top: 8px; font-size: 12px; color: var(--text-muted); line-height: 1.5; }
.section-row { display: flex; justify-content: space-between; align-items: flex-start; gap: 16px; margin-bottom: 16px; }
.section-title { font-size: 16px; font-weight: 700; }
.section-sub { margin-top: 6px; font-size: 12px; color: var(--text-muted); line-height: 1.6; }
.tab-toolbar { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; margin-bottom: 18px; }
.toolbar-actions { display: flex; justify-content: flex-end; margin-top: 8px; }
.chart-card { min-height: 360px; }
.chart-shell { height: 280px; }
.chart-canvas { width: 100%; height: 100%; }
.empty-text { color: var(--text-secondary); padding-top: 28px; }

@media (max-width: 980px) {
  .tab-toolbar,
  .section-row {
    flex-direction: column;
  }
}
</style>
