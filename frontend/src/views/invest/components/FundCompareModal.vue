<template>
  <n-modal
    :show="show"
    preset="card"
    title="基金横向对比"
    style="width: 1520px; max-width: calc(100vw - 32px);"
    @update:show="emit('update:show', $event)"
  >
    <n-space vertical size="large">
      <div class="compare-summary">
        <div class="compare-summary-head">
          <div class="compare-summary-text">
            <n-text depth="3">
              支持 1-10 只基金横向看近期核心表现。当前已选 {{ codes.length }} 只，
              更新时间：{{ result?.refreshedAt || '-' }}
            </n-text>
            <n-text v-if="result?.missingCodes?.length" type="warning">
              以下基金暂时没有拿到数据：{{ result.missingCodes.join('、') }}
            </n-text>
          </div>
          <n-button :disabled="codes.length < 2" @click="showSelectionAI = true">
            AI分析勾选基金
          </n-button>
        </div>
      </div>

      <n-spin :show="loading">
        <div class="compare-scroll">
          <table class="compare-table">
            <thead>
              <tr>
                <th class="metric-col">对比项</th>
                <th v-for="item in items" :key="item.code" class="fund-col">
                  <div class="fund-head">
                    <n-button text type="primary" @click="openFund(item.code)">
                      {{ item.name || item.code }}
                    </n-button>
                    <div class="fund-code">{{ item.code }}</div>
                    <div class="fund-tags">
                      <n-tag size="small" type="info" :bordered="false">
                        {{ item.categoryLabel || '基金' }}
                      </n-tag>
                      <n-tag size="small" :bordered="false">
                        {{ item.fundType || '类型待补' }}
                      </n-tag>
                    </div>
                  </div>
                </th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="metric in metricRows" :key="metric.key">
                <th class="metric-col">{{ metric.label }}</th>
                <td
                  v-for="item in items"
                  :key="metric.key + '-' + item.code"
                  :class="['value-cell', { best: isBest(metric, item) }]"
                >
                  {{ renderMetric(metric, item) }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </n-spin>
    </n-space>

    <FundAIAnalysisModal
      v-model:show="showSelectionAI"
      mode="selection"
      :fund-codes="codes"
      :scope-label="tabLabel"
      :title="compareAITitle"
    />
  </n-modal>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useMessage } from 'naive-ui'
import { OpenURL } from '../../../../wailsjs/go/main/App'
import FundAIAnalysisModal from '../../portfolio/components/FundAIAnalysisModal.vue'

const props = defineProps({
  show: { type: Boolean, default: false },
  codes: { type: Array, default: () => [] },
  tabLabel: { type: String, default: '' }
})

const emit = defineEmits(['update:show'])

const message = useMessage()
const loading = ref(false)
const result = ref(null)
const showSelectionAI = ref(false)

const items = computed(() => result.value?.items || [])
const compareAITitle = computed(() => {
  const label = String(props.tabLabel || '').trim()
  return label ? `${label} · 勾选基金 AI 对比` : '勾选基金 AI 对比'
})

const metricRows = [
  { key: 'netGrowth7', label: '近7天', type: 'percent', better: 'higher' },
  { key: 'netGrowth1', label: '近1月', type: 'percent', better: 'higher' },
  { key: 'maxDrawdown1', label: '近1月最大回撤', type: 'percent', better: 'lower' },
  { key: 'netGrowth3', label: '近3月', type: 'percent', better: 'higher' },
  { key: 'maxDrawdown3', label: '近3月最大回撤', type: 'percent', better: 'lower' },
  { key: 'netGrowth6', label: '近6月', type: 'percent', better: 'higher' },
  { key: 'netGrowth12', label: '近1年', type: 'percent', better: 'higher' },
  { key: 'maxDrawdown12', label: '近1年最大回撤', type: 'percent', better: 'lower' },
  { key: 'topIndustry', label: '所属行业', type: 'text' },
  { key: 'company', label: '基金公司', type: 'text' },
  { key: 'manager', label: '基金经理', type: 'text' },
  { key: 'ratingScale', label: '评级 / 规模', type: 'custom' },
  { key: 'screenUpdatedAt', label: '更新时间', type: 'text' }
]

watch(
  () => [props.show, ...(props.codes || [])],
  async ([show]) => {
    if (!show) return
    await loadCompare()
  },
  { deep: true }
)

watch(
  () => props.show,
  (show) => {
    if (!show) {
      showSelectionAI.value = false
    }
  }
)

async function loadCompare() {
  const codes = Array.isArray(props.codes) ? props.codes.filter(Boolean).slice(0, 10) : []
  if (codes.length === 0) {
    result.value = null
    return
  }
  if (!window.go?.main?.App?.CompareFunds) {
    message.error('当前版本暂不支持基金对比')
    return
  }

  loading.value = true
  try {
    result.value = await window.go.main.App.CompareFunds({ codes })
  } catch (error) {
    console.error(error)
    message.error('基金对比加载失败')
  } finally {
    loading.value = false
  }
}

function renderMetric(metric, item) {
  if (!item) return '-'
  if (metric.key === 'ratingScale') {
    return [item.rating || '暂无评级', item.scale || '规模待补'].join(' / ')
  }
  if (metric.type === 'percent') {
    const value = item[metric.key]
    if (value === null || value === undefined || Number.isNaN(Number(value))) {
      return '-'
    }
    const num = Number(value)
    return `${num >= 0 ? '+' : ''}${num.toFixed(2)}%`
  }
  return item[metric.key] || '-'
}

function isBest(metric, item) {
  if (metric.type !== 'percent') return false
  const values = items.value
    .map((entry) => Number(entry[metric.key]))
    .filter((value) => !Number.isNaN(value))
  if (values.length < 2) return false

  const current = Number(item[metric.key])
  if (Number.isNaN(current)) return false

  const target = metric.better === 'lower' ? Math.min(...values) : Math.max(...values)
  return current === target
}

function openFund(code) {
  OpenURL(`https://fund.eastmoney.com/${code}.html`)
}
</script>

<style scoped>
.compare-summary {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.compare-summary-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}

.compare-summary-text {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.compare-scroll {
  overflow-x: auto;
  padding-bottom: 8px;
}

.compare-table {
  width: max-content;
  min-width: 100%;
  border-collapse: separate;
  border-spacing: 0;
}

.compare-table th,
.compare-table td {
  min-width: 220px;
  padding: 14px 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  vertical-align: top;
  background: rgba(15, 23, 42, 0.9);
}

.compare-table thead th {
  background: rgba(30, 41, 59, 0.96);
}

.metric-col {
  position: sticky;
  left: 0;
  z-index: 2;
  min-width: 160px !important;
  max-width: 160px;
  text-align: left;
  white-space: nowrap;
}

.fund-col {
  z-index: 1;
}

.fund-head {
  display: flex;
  flex-direction: column;
  gap: 8px;
  align-items: flex-start;
}

.fund-code {
  font-size: 12px;
  color: var(--text-secondary);
}

.fund-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.value-cell {
  color: var(--text-primary);
  line-height: 1.6;
}

.value-cell.best {
  background: rgba(20, 184, 166, 0.12);
  color: #6ee7d8;
  font-weight: 600;
}

@media (max-width: 900px) {
  .compare-summary-head {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
