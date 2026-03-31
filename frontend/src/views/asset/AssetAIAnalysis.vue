<template>
  <div class="fade-in household-ai-analysis">
    <n-page-header title="AI资产分析" subtitle="基于家庭台账、负债计划、快照镜像和统一 AI 配置生成分析结论">
      <template #extra>
        <n-space>
          <n-button @click="$router.push('/asset/benchmarks')">基准数据页</n-button>
          <n-button secondary @click="loadBootstrap">刷新配置</n-button>
          <n-button type="primary" :loading="running" @click="runAnalysis('manual')">手动重新分析</n-button>
        </n-space>
      </template>
    </n-page-header>

    <n-grid :cols="3" :x-gap="16" :y-gap="16" style="margin-top: 20px;">
      <n-gi>
        <div class="metric-card">
          <div class="metric-label">当前地区</div>
          <div class="metric-value">{{ region }}</div>
          <div class="metric-sub">分析会优先结合该地区和全国口径进行对比</div>
        </div>
      </n-gi>
      <n-gi>
        <div class="metric-card">
          <div class="metric-label">AI 模型</div>
          <div class="metric-value">{{ selectedAiLabel }}</div>
          <div class="metric-sub">复用设置页维护的统一 AI 源</div>
        </div>
      </n-gi>
      <n-gi>
        <div class="metric-card">
          <div class="metric-label">最近分析</div>
          <div class="metric-value">{{ latestStatusLabel }}</div>
          <div class="metric-sub">{{ latestMetaText }}</div>
        </div>
      </n-gi>
    </n-grid>

    <div class="platform-card" style="margin-top: 20px;">
      <div class="section-row">
        <div>
          <div class="section-title">家庭画像</div>
          <div class="section-sub">这部分会和资产、负债、基准一起进入 AI 上下文，帮助模型理解家庭背景。</div>
        </div>
        <n-space>
          <n-button secondary @click="loadProfile">刷新画像</n-button>
          <n-button type="primary" @click="showProfileEditor = true">编辑画像</n-button>
        </n-space>
      </div>

      <n-grid :cols="4" :x-gap="16" :y-gap="16">
        <n-gi>
          <div class="mini-card">
            <div class="mini-label">家庭名称</div>
            <div class="mini-value">{{ profile.householdName || '我的家庭' }}</div>
            <div class="mini-sub">{{ profile.region || region }}</div>
          </div>
        </n-gi>
        <n-gi>
          <div class="mini-card">
            <div class="mini-label">家庭成员</div>
            <div class="mini-value">{{ profile.membersCount || 1 }} 人</div>
            <div class="mini-sub">抚养人数 {{ profile.dependentsCount || 0 }} 人</div>
          </div>
        </n-gi>
        <n-gi>
          <div class="mini-card">
            <div class="mini-label">住房状态</div>
            <div class="mini-value">{{ profile.housingStatus || '未填写' }}</div>
            <div class="mini-sub">{{ profile.cityTier || '未填写城市能级' }}</div>
          </div>
        </n-gi>
        <n-gi>
          <div class="mini-card">
            <div class="mini-label">风险偏好</div>
            <div class="mini-value">{{ profile.riskPreference || '稳健' }}</div>
            <div class="mini-sub">年家庭支出 {{ formatMoney(profile.annualHouseholdSpend || 0) }}</div>
          </div>
        </n-gi>
        <n-gi>
          <div class="mini-card">
            <div class="mini-label">个人缴存</div>
            <div class="mini-value">{{ formatMoney((profile.monthlyPersonalInsuranceContribution || 0) + (profile.monthlyHousingFundContribution || 0)) }}</div>
            <div class="mini-sub">社保 {{ formatMoney(profile.monthlyPersonalInsuranceContribution || 0) }} / 公积金 {{ formatMoney(profile.monthlyHousingFundContribution || 0) }}</div>
          </div>
        </n-gi>
        <n-gi>
          <div class="mini-card">
            <div class="mini-label">专项附加抵扣</div>
            <div class="mini-value">{{ formatMoney(totalSpecialDeduction) }}</div>
            <div class="mini-sub">育儿、房贷、赡养和其他专项附加合计</div>
          </div>
        </n-gi>
      </n-grid>
    </div>

    <div class="platform-card" style="margin-top: 20px;">
      <div class="section-title">分析配置</div>
      <div class="section-sub">这里不单独管理密钥，直接复用设置页和提示词模板页。</div>
      <n-grid :cols="3" :x-gap="16" style="margin-top: 16px;">
        <n-form-item label="地区">
          <n-input v-model:value="region" placeholder="例如：天津市" />
        </n-form-item>
        <n-form-item label="AI 源">
          <n-select v-model:value="selectedAiConfigId" :options="aiOptions" />
        </n-form-item>
        <n-form-item label="系统提示词模板">
          <n-select v-model:value="selectedPromptTemplateId" :options="promptOptions" />
        </n-form-item>
      </n-grid>
    </div>

    <div class="platform-card" style="margin-top: 20px;">
      <div class="section-row">
        <div>
          <div class="section-title">最近分析结果</div>
          <div class="section-sub">资产台账保存或删除后会自动触发一次后台分析，这里也支持手动重跑。</div>
        </div>
      </div>

      <n-alert v-if="statusMessage" :type="statusType" :show-icon="false" style="margin-bottom: 16px;">
        {{ statusMessage }}
      </n-alert>

      <n-descriptions v-if="latestRecord" label-placement="left" :column="3" bordered>
        <n-descriptions-item label="触发来源">{{ latestRecord.triggerSource || '-' }}</n-descriptions-item>
        <n-descriptions-item label="模型">{{ latestRecord.modelName || '-' }}</n-descriptions-item>
        <n-descriptions-item label="基准版本">{{ latestRecord.benchmarkVersion || '-' }}</n-descriptions-item>
      </n-descriptions>

      <n-grid :cols="2" :x-gap="16" :y-gap="16" style="margin-top: 20px;">
        <n-gi v-for="section in parsedSections" :key="section.title">
          <div class="analysis-card">
            <div class="analysis-title">{{ section.title }}</div>
            <MdPreview class="analysis-preview" :editor-id="analysisSectionEditorId(section.title)" :model-value="section.content" theme="dark" preview-theme="github" />
          </div>
        </n-gi>
      </n-grid>

      <div class="section-row" style="margin-top: 20px; margin-bottom: 12px;">
        <div>
          <div class="section-title">大致水平</div>
          <div class="section-sub">基于当前内置天津/全国基准做粗略判断，只用于辅助观察，不替代权威分布统计。</div>
        </div>
      </div>
      <n-grid :cols="2" :x-gap="16" :y-gap="16">
        <n-gi v-for="item in rankingCards" :key="item.label">
          <div class="mini-card ranking-card">
            <div class="mini-label">{{ item.label }}</div>
            <div class="mini-value">{{ item.value }}</div>
            <div class="mini-sub">{{ item.sub }}</div>
          </div>
        </n-gi>
      </n-grid>

      <n-collapse style="margin-top: 20px;">
        <n-collapse-item title="本次实际 Prompt" name="prompt">
          <pre class="code-block">{{ latestRecord?.prompt || usedPrompt || '暂无 Prompt' }}</pre>
        </n-collapse-item>
        <n-collapse-item title="输入快照(JSON)" name="payload">
          <pre class="code-block">{{ latestRecord?.inputPayload || latestPayload || '暂无输入快照' }}</pre>
        </n-collapse-item>
        <n-collapse-item title="完整原始分析" name="raw">
          <pre class="code-block">{{ latestAnalysis || '暂无分析结果' }}</pre>
        </n-collapse-item>
      </n-collapse>
    </div>

    <div class="platform-card" style="margin-top: 20px;">
      <div class="section-row">
        <div>
          <div class="section-title">基准数据</div>
          <div class="section-sub">这里的数据会直接进入 AI 分析上下文，用于天津市/全国对比。</div>
        </div>
        <n-space>
          <n-button secondary @click="loadBenchmarks">刷新基准</n-button>
          <n-button @click="$router.push('/asset/benchmarks')">独立管理页</n-button>
          <n-button type="primary" @click="openBenchmarkEditor()">新增基准</n-button>
        </n-space>
      </div>

      <n-data-table
        :columns="benchmarkColumns"
        :data="benchmarks"
        :pagination="{ pageSize: 6 }"
        :bordered="false"
      />
    </div>

    <n-modal v-model:show="showBenchmarkEditor" preset="card" :title="editingBenchmarkId ? '编辑基准数据' : '新增基准数据'" style="width: 720px;">
      <n-form label-placement="top">
        <n-grid :cols="2" :x-gap="16">
          <n-form-item label="名称">
            <n-input v-model:value="benchmarkForm.name" />
          </n-form-item>
          <n-form-item label="地区">
            <n-input v-model:value="benchmarkForm.region" placeholder="例如：天津市 / 全国" />
          </n-form-item>
          <n-form-item label="范围">
            <n-select v-model:value="benchmarkForm.scope" :options="scopeOptions" />
          </n-form-item>
          <n-form-item label="分类">
            <n-select v-model:value="benchmarkForm.category" :options="categoryOptions" />
          </n-form-item>
          <n-form-item label="数值">
            <n-input-number v-model:value="benchmarkForm.value" :show-button="false" style="width: 100%;" />
          </n-form-item>
          <n-form-item label="单位">
            <n-input v-model:value="benchmarkForm.unit" placeholder="例如：元/年" />
          </n-form-item>
          <n-form-item label="年份">
            <n-input-number v-model:value="benchmarkForm.year" :show-button="false" style="width: 100%;" />
          </n-form-item>
          <n-form-item label="版本">
            <n-input v-model:value="benchmarkForm.version" placeholder="例如：built-in-2026.03" />
          </n-form-item>
          <n-form-item label="说明" style="grid-column: span 2;">
            <n-input v-model:value="benchmarkForm.description" type="textarea" :rows="3" />
          </n-form-item>
        </n-grid>
      </n-form>
      <template #action>
        <n-space justify="end">
          <n-button @click="showBenchmarkEditor = false">取消</n-button>
          <n-button type="primary" @click="saveBenchmark">保存</n-button>
        </n-space>
      </template>
    </n-modal>

    <n-modal v-model:show="showProfileEditor" preset="card" title="编辑家庭画像" style="width: 760px;">
      <n-form label-placement="top">
        <n-grid :cols="2" :x-gap="16">
          <n-form-item label="家庭名称">
            <n-input v-model:value="profileForm.householdName" placeholder="例如：三口之家" />
          </n-form-item>
          <n-form-item label="地区">
            <n-input v-model:value="profileForm.region" placeholder="例如：天津市" />
          </n-form-item>
          <n-form-item label="城市能级">
            <n-select v-model:value="profileForm.cityTier" :options="cityTierOptions" />
          </n-form-item>
          <n-form-item label="住房状态">
            <n-select v-model:value="profileForm.housingStatus" :options="housingStatusOptions" />
          </n-form-item>
          <n-form-item label="家庭成员人数">
            <n-input-number v-model:value="profileForm.membersCount" :show-button="false" style="width: 100%;" :min="1" />
          </n-form-item>
          <n-form-item label="抚养人数">
            <n-input-number v-model:value="profileForm.dependentsCount" :show-button="false" style="width: 100%;" :min="0" />
          </n-form-item>
          <n-form-item label="风险偏好">
            <n-select v-model:value="profileForm.riskPreference" :options="riskPreferenceOptions" />
          </n-form-item>
          <n-form-item label="主要收入来源">
            <n-input v-model:value="profileForm.primaryIncomeSource" placeholder="例如：工资收入" />
          </n-form-item>
          <n-form-item label="月个人社保" >
            <n-input-number v-model:value="profileForm.monthlyPersonalInsuranceContribution" :show-button="false" style="width: 100%;" :min="0" />
          </n-form-item>
          <n-form-item label="月个人公积金">
            <n-input-number v-model:value="profileForm.monthlyHousingFundContribution" :show-button="false" style="width: 100%;" :min="0" />
          </n-form-item>
          <n-form-item label="其他税前扣除">
            <n-input-number v-model:value="profileForm.monthlyOtherPretaxDeduction" :show-button="false" style="width: 100%;" :min="0" />
          </n-form-item>
          <n-form-item label="育儿个税抵扣">
            <n-input-number v-model:value="profileForm.monthlyChildcareDeduction" :show-button="false" style="width: 100%;" :min="0" />
          </n-form-item>
          <n-form-item label="房贷个税抵扣">
            <n-input-number v-model:value="profileForm.monthlyHousingLoanDeduction" :show-button="false" style="width: 100%;" :min="0" />
          </n-form-item>
          <n-form-item label="赡养老人个税抵扣">
            <n-input-number v-model:value="profileForm.monthlyElderlyCareDeduction" :show-button="false" style="width: 100%;" :min="0" />
          </n-form-item>
          <n-form-item label="其他专项附加抵扣">
            <n-input-number v-model:value="profileForm.monthlyOtherSpecialDeduction" :show-button="false" style="width: 100%;" :min="0" />
          </n-form-item>
          <n-form-item label="年家庭支出" style="grid-column: span 2;">
            <n-input-number v-model:value="profileForm.annualHouseholdSpend" :show-button="false" style="width: 100%;" :min="0" />
          </n-form-item>
          <n-form-item label="备注" style="grid-column: span 2;">
            <n-input v-model:value="profileForm.notes" type="textarea" :rows="3" placeholder="补充工作阶段、孩子教育、未来大额支出计划等。" />
          </n-form-item>
        </n-grid>
      </n-form>
      <template #action>
        <n-space justify="end">
          <n-button @click="showProfileEditor = false">取消</n-button>
          <n-button type="primary" @click="saveProfile">保存</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup>
import { computed, h, onMounted, ref } from 'vue'
import { NButton, NPopconfirm, useMessage } from 'naive-ui'
import { MdPreview } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'

const message = useMessage()
const running = ref(false)
const region = ref('天津市')
const profile = ref(createProfileForm())
const profileForm = ref(createProfileForm())
const aiConfigs = ref([])
const promptTemplates = ref([])
const selectedAiConfigId = ref(null)
const selectedPromptTemplateId = ref(null)
const latestRecord = ref(null)
const latestAnalysis = ref('')
const latestPayload = ref('')
const usedPrompt = ref('')
const statusMessage = ref('')
const statusType = ref('info')
const benchmarks = ref([])
const showBenchmarkEditor = ref(false)
const showProfileEditor = ref(false)
const editingBenchmarkId = ref(0)
const benchmarkForm = ref(createBenchmarkForm())

const cityTierOptions = [
  { label: '一线', value: '一线' },
  { label: '新一线', value: '新一线' },
  { label: '二线', value: '二线' },
  { label: '三四线', value: '三四线' }
]

const housingStatusOptions = [
  { label: '自有住房', value: '自有住房' },
  { label: '按揭中', value: '按揭中' },
  { label: '租房', value: '租房' },
  { label: '与父母同住', value: '与父母同住' }
]

const riskPreferenceOptions = [
  { label: '保守', value: '保守' },
  { label: '稳健', value: '稳健' },
  { label: '平衡', value: '平衡' },
  { label: '积极', value: '积极' }
]

const scopeOptions = [
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

const aiOptions = computed(() => aiConfigs.value.map((item) => ({
  label: `${item.Name || item.name || '未命名'} / ${item.ModelName || item.modelName || '未设模型'}`,
  value: Number(item.ID || item.id || 0)
})))

const promptOptions = computed(() => promptTemplates.value.map((item) => ({
  label: item.name,
  value: item.ID
})))

const selectedAiLabel = computed(() => aiOptions.value.find((item) => item.value === selectedAiConfigId.value)?.label || '未选择')
const latestStatusLabel = computed(() => latestRecord.value?.status || '未运行')
const latestMetaText = computed(() => {
  if (!latestRecord.value?.createdAt) return '还没有分析记录'
  return new Date(latestRecord.value.createdAt).toLocaleString('zh-CN')
})
const totalSpecialDeduction = computed(() =>
  Number(profile.value.monthlyChildcareDeduction || 0) +
  Number(profile.value.monthlyHousingLoanDeduction || 0) +
  Number(profile.value.monthlyElderlyCareDeduction || 0) +
  Number(profile.value.monthlyOtherSpecialDeduction || 0)
)

const rankingCards = computed(() => {
  const members = Math.max(1, Number(profile.value.membersCount || 1))
  const netAssetsPerAdult = Number(latestPayloadSummary.value.netAssets || 0) / members
  const debtRatio = Number(latestPayloadSummary.value.debtRatio || 0)
  return [
    buildAssetRankingCard('天津资产水平', '天津市', netAssetsPerAdult, members),
    buildAssetRankingCard('全国资产水平', '全国', netAssetsPerAdult, members),
    buildDebtRankingCard('天津负债率水平', '天津市', debtRatio),
    buildDebtRankingCard('全国负债率水平', '全国', debtRatio)
  ]
})

const latestPayloadSummary = computed(() => {
  if (!latestPayload.value) return {}
  try {
    return JSON.parse(latestPayload.value)?.summary || {}
  } catch {
    return {}
  }
})

const parsedSections = computed(() => {
  if (!latestAnalysis.value) return []
  const titles = ['核心结论', '关键指标表', '风险点', '优化建议', '地区/全国对比', '后续关注项']
  return titles.map((title, index) => {
    const nextTitle = titles[index + 1]
    const pattern = nextTitle
      ? new RegExp(`(?:^|\\n)#{0,3}\\s*${title}[：:\\n\\r]*([\\s\\S]*?)(?=(?:\\n#{0,3}\\s*${nextTitle}[：:\\n\\r]))`)
      : new RegExp(`(?:^|\\n)#{0,3}\\s*${title}[：:\\n\\r]*([\\s\\S]*)`)
    const match = latestAnalysis.value.match(pattern)
    return {
      title,
      content: (match?.[1] || '暂无该部分内容').trim().replace(/^---+\s*/g, '').trim()
    }
  })
})

function analysisSectionEditorId(title) {
  return `household-ai-section-${String(title || '')
    .replace(/[^a-zA-Z0-9\u4e00-\u9fa5_-]+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '') || 'section'}`
}

const benchmarkColumns = [
  { title: '名称', key: 'name', width: 180 },
  { title: '地区', key: 'region', width: 100 },
  { title: '范围', key: 'scope', width: 90 },
  { title: '分类', key: 'category', width: 90, render: (row) => categoryLabel(row.category) },
  { title: '数值', key: 'value', width: 140, render: (row) => `${Number(row.value || 0).toLocaleString('zh-CN')} ${row.unit || ''}` },
  { title: '年份', key: 'year', width: 80 },
  { title: '版本', key: 'version', width: 120 },
  {
    title: '操作',
    key: 'actions',
    width: 150,
    render: (row) => h('div', { style: { display: 'flex', gap: '8px' } }, [
      h(NButton, { size: 'tiny', quaternary: true, onClick: () => openBenchmarkEditor(row) }, () => '编辑'),
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

function createBenchmarkForm() {
  return {
    name: '',
    scope: '地区',
    region: '天津市',
    category: 'income',
    value: 0,
    unit: '元/年',
    year: 2025,
    version: 'manual',
    description: ''
  }
}

function createProfileForm() {
  return {
    ID: 0,
    householdName: '我的家庭',
    region: '天津市',
    cityTier: '新一线',
    membersCount: 2,
    dependentsCount: 0,
    housingStatus: '自有住房',
    riskPreference: '稳健',
    annualHouseholdSpend: 0,
    primaryIncomeSource: '工资收入',
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

function categoryLabel(value) {
  return {
    income: '收入',
    debt: '负债',
    asset: '资产',
    protection: '保障',
    other: '其他'
  }[value] || value || '-'
}

function formatMoney(value) {
  return `¥ ${Number(value || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

function findBenchmark(regionName, candidates) {
  return benchmarks.value.find((item) => item.region === regionName && candidates.includes(item.name))
}

function assetLevelByMultiple(multiple) {
  if (multiple <= 0) return '暂缺数据'
  if (multiple < 1) return '约后 50%'
  if (multiple < 2) return '约前 50%-35%'
  if (multiple < 4) return '约前 35%-20%'
  if (multiple < 8) return '约前 20%-10%'
  if (multiple < 15) return '约前 10%-5%'
  return '约前 5%以内'
}

function debtLevelByRatio(relativeRatio) {
  if (relativeRatio <= 0) return '暂缺数据'
  if (relativeRatio <= 0.3) return '优于约 90%'
  if (relativeRatio <= 0.6) return '优于约 75%'
  if (relativeRatio <= 1.0) return '优于约 60%'
  if (relativeRatio <= 1.4) return '接近中位'
  if (relativeRatio <= 2.0) return '约后 40%'
  return '约后 20%'
}

function buildAssetRankingCard(label, regionName, netAssetsPerAdult, members) {
  const incomeBenchmark = findBenchmark(regionName, ['天津居民人均可支配收入', '全国居民人均可支配收入'])
  if (!incomeBenchmark?.value || netAssetsPerAdult <= 0) {
    return { label, value: '基准不足', sub: '缺少可支配收入基准，暂不能估算资产层级。' }
  }
  const multiple = netAssetsPerAdult / Number(incomeBenchmark.value)
  return {
    label,
    value: assetLevelByMultiple(multiple),
    sub: `按人均净资产 ${formatMoney(netAssetsPerAdult)} 对比 ${regionName} 人均收入 ${formatMoney(incomeBenchmark.value)}，约 ${multiple.toFixed(1)} 倍。`
  }
}

function buildDebtRankingCard(label, regionName, debtRatio) {
  const benchmark = findBenchmark(regionName, ['天津负债率参考', '全国负债率参考', '天津贷款/存款余额比', '全国住户贷款/存款余额比'])
  if (benchmark?.value == null) {
    return { label, value: '基准不足', sub: '缺少地区杠杆代理指标，暂不能估算负债率层级。' }
  }
  const relativeRatio = debtRatio / Number(benchmark.value)
  return {
    label,
    value: debtLevelByRatio(relativeRatio),
    sub: `当前负债率 ${Number(debtRatio).toFixed(2)}%，对比 ${regionName} 代理杠杆 ${Number(benchmark.value).toFixed(2)}%。`
  }
}

function pickDefaultPromptId() {
  const exact = promptTemplates.value.find((item) => item.name === '家庭资产分析-标准模板')
  if (exact) return exact.ID
  return promptTemplates.value[0]?.ID || null
}

async function loadBootstrap() {
  await loadProfile()
  if (window.go?.main?.App?.GetAiConfigs) {
    aiConfigs.value = (await window.go.main.App.GetAiConfigs()) || []
  }
  if (window.go?.main?.App?.GetPromptTemplates) {
    promptTemplates.value = (await window.go.main.App.GetPromptTemplates('', '模型系统Prompt')) || []
  }
  if (!selectedAiConfigId.value && aiConfigs.value.length > 0) {
    selectedAiConfigId.value = Number(aiConfigs.value[0].ID || aiConfigs.value[0].id || 0)
  }
  if (!selectedPromptTemplateId.value && promptTemplates.value.length > 0) {
    selectedPromptTemplateId.value = pickDefaultPromptId()
  }
  await Promise.all([loadLatest(), loadBenchmarks()])
}

async function loadProfile() {
  if (!window.go?.main?.App?.GetHouseholdProfile) return
  const data = await window.go.main.App.GetHouseholdProfile()
  if (!data) return
  profile.value = {
    ...createProfileForm(),
    ...data
  }
  profileForm.value = {
    ...createProfileForm(),
    ...data
  }
  if (!region.value || region.value === '天津市') {
    region.value = profile.value.region || '天津市'
  }
}

async function loadLatest() {
  if (!window.go?.main?.App?.GetLatestHouseholdAIAnalysis) return
  latestRecord.value = await window.go.main.App.GetLatestHouseholdAIAnalysis()
  latestAnalysis.value = latestRecord.value?.analysisMarkdown || ''
  latestPayload.value = latestRecord.value?.inputPayload || ''
  usedPrompt.value = latestRecord.value?.prompt || ''
}

async function loadBenchmarks() {
  if (!window.go?.main?.App?.GetHouseholdBenchmarks) return
  benchmarks.value = (await window.go.main.App.GetHouseholdBenchmarks(region.value)) || []
}

async function runAnalysis(triggerSource) {
  if (!window.go?.main?.App?.RunHouseholdAIAnalysis) {
    message.error('当前版本未接入家庭资产 AI 分析接口')
    return
  }
  running.value = true
  statusMessage.value = ''
  try {
    const result = await window.go.main.App.RunHouseholdAIAnalysis(
      region.value,
      selectedAiConfigId.value || 0,
      selectedPromptTemplateId.value || 0,
      triggerSource
    )
    statusType.value = result.success ? 'success' : 'warning'
    statusMessage.value = result.message || (result.success ? '分析完成' : '分析失败')
    latestAnalysis.value = result.analysis || ''
    usedPrompt.value = result.prompt || ''
    await loadLatest()
    if (result.success) {
      message.success('家庭资产分析已更新')
    } else {
      message.warning(result.message || '分析未完成')
    }
  } catch (error) {
    console.error(error)
    statusType.value = 'error'
    statusMessage.value = error?.message || '分析失败'
    message.error(statusMessage.value)
  } finally {
    running.value = false
  }
}

function openBenchmarkEditor(row = null) {
  editingBenchmarkId.value = row?.ID || 0
  benchmarkForm.value = row
    ? {
      name: row.name || '',
      scope: row.scope || '地区',
      region: row.region || '天津市',
      category: row.category || 'income',
      value: Number(row.value || 0),
      unit: row.unit || '元/年',
      year: Number(row.year || 2025),
      version: row.version || 'manual',
      description: row.description || ''
    }
    : createBenchmarkForm()
  showBenchmarkEditor.value = true
}

async function saveBenchmark() {
  try {
    await window.go.main.App.UpsertHouseholdBenchmark({
      ID: editingBenchmarkId.value || 0,
      ...benchmarkForm.value,
      isActive: true
    })
    showBenchmarkEditor.value = false
    await loadBenchmarks()
    message.success('基准数据已保存')
  } catch (error) {
    console.error(error)
    message.error('保存基准数据失败')
  }
}

async function saveProfile() {
  try {
    const saved = await window.go.main.App.UpsertHouseholdProfile({
      ...profileForm.value,
      householdName: (profileForm.value.householdName || '').trim() || '我的家庭',
      region: (profileForm.value.region || '').trim() || '天津市'
    })
    profile.value = {
      ...createProfileForm(),
      ...saved
    }
    profileForm.value = {
      ...createProfileForm(),
      ...saved
    }
    region.value = profile.value.region || region.value
    showProfileEditor.value = false
    await loadBenchmarks()
    message.success('家庭画像已保存')
  } catch (error) {
    console.error(error)
    message.error('保存家庭画像失败')
  }
}

async function removeBenchmark(id) {
  try {
    await window.go.main.App.DeleteHouseholdBenchmark(id)
    await loadBenchmarks()
    message.success('基准数据已删除')
  } catch (error) {
    console.error(error)
    message.error('删除基准数据失败')
  }
}

onMounted(loadBootstrap)
</script>

<style scoped>
.household-ai-analysis { max-width: 1280px; margin: 0 auto; }
.metric-card { background: var(--bg-card); border: 1px solid var(--border-color); border-radius: var(--radius-md); padding: 18px; min-height: 138px; }
.metric-label { font-size: 13px; color: var(--text-secondary); }
.metric-value { margin-top: 10px; font-size: 24px; font-weight: 700; }
.metric-sub { margin-top: 8px; font-size: 12px; color: var(--text-muted); line-height: 1.5; }
.mini-card { background: linear-gradient(180deg, rgba(15, 23, 42, 0.92), rgba(30, 41, 59, 0.84)); border: 1px solid rgba(148, 163, 184, 0.16); border-radius: 14px; padding: 16px; min-height: 120px; }
.mini-label { font-size: 12px; color: var(--text-secondary); }
.mini-value { margin-top: 10px; font-size: 20px; font-weight: 700; color: #f8fafc; }
.mini-sub { margin-top: 8px; font-size: 12px; color: var(--text-muted); line-height: 1.5; }
.section-title { font-size: 16px; font-weight: 700; }
.section-sub { margin-top: 6px; color: var(--text-muted); font-size: 12px; }
.section-row { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; margin-bottom: 16px; }
.analysis-card { background: linear-gradient(180deg, rgba(15, 23, 42, 0.94), rgba(30, 41, 59, 0.86)); border: 1px solid rgba(148, 163, 184, 0.14); border-radius: 14px; padding: 16px; min-height: 180px; }
.analysis-title { font-size: 14px; font-weight: 700; margin-bottom: 10px; }
.code-block { white-space: pre-wrap; line-height: 1.6; font-size: 12px; margin: 0; }
.ranking-card { min-height: 132px; }

:deep(.analysis-preview) {
  background: transparent;
  color: #dbe4f0;
}

:deep(.analysis-preview .md-editor-preview) {
  background: transparent;
  padding: 0;
}

:deep(.analysis-preview table) {
  width: 100%;
  border-collapse: collapse;
}

:deep(.analysis-preview th),
:deep(.analysis-preview td) {
  border: 1px solid rgba(148, 163, 184, 0.18);
  padding: 8px 10px;
}
</style>
