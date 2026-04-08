<template>
  <div class="fade-in transaction-page">
    <n-page-header title="交易记录" subtitle="支持基金为主的买入卖出录入，平台和账户分组会直接回写到持仓">
      <template #extra>
        <n-space>
          <n-input v-model:value="keyword" clearable placeholder="搜索代码或名称" style="width: 220px;" />
          <n-select v-model:value="holdingTypeFilter" :options="holdingTypeOptions" style="width: 120px;" />
          <n-select v-model:value="tradeTypeFilter" :options="tradeTypeOptions" style="width: 120px;" />
          <n-button @click="loadData">刷新</n-button>
          <n-button type="primary" @click="showAddModal = true">录入持仓</n-button>
        </n-space>
      </template>
    </n-page-header>

    <div class="summary-card">
      <div class="summary-item">
        <div class="summary-label">总记录</div>
        <div class="summary-value">{{ filteredTransactions.length }}</div>
      </div>
      <div class="summary-item">
        <div class="summary-label">基金交易</div>
        <div class="summary-value">{{ fundTransactionCount }}</div>
      </div>
      <div class="summary-item">
        <div class="summary-label">买入金额</div>
        <div class="summary-value">{{ formatMoney(buyAmount) }}</div>
      </div>
      <div class="summary-item">
        <div class="summary-label">卖出金额</div>
        <div class="summary-value">{{ formatMoney(sellAmount) }}</div>
      </div>
    </div>

    <div class="table-card">
      <n-data-table
        :columns="columns"
        :data="filteredTransactions"
        :pagination="{ pageSize: 15 }"
        :scroll-x="1280"
        striped
      />
    </div>

    <n-modal v-model:show="showAddModal" preset="card" title="持仓录入与交易记录" style="width: 700px;">
      <n-form label-placement="top">
        <n-form-item label="录入方式">
          <n-radio-group v-model:value="entryMode" name="entry-mode">
            <n-space>
              <n-radio-button value="position">当前仓位</n-radio-button>
              <n-radio-button value="trade">明细交易</n-radio-button>
            </n-space>
          </n-radio-group>
        </n-form-item>

        <template v-if="entryMode === 'position'">
          <n-grid :cols="2" :x-gap="12">
            <n-gi>
              <n-form-item label="基金代码">
                <n-input v-model:value="positionForm.stockCode" placeholder="例如 013449" />
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="基金名称">
                <n-input v-model:value="positionForm.stockName" placeholder="例如 广发景宁纯债债券C" />
              </n-form-item>
            </n-gi>
          </n-grid>

          <n-grid :cols="2" :x-gap="12">
            <n-gi>
              <n-form-item label="平台">
                <n-input v-model:value="positionForm.brokerName" placeholder="默认支付宝" />
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="账户分组">
                <n-input v-model:value="positionForm.accountTag" placeholder="例如 主账户 / 家庭账户" />
              </n-form-item>
            </n-gi>
          </n-grid>

          <n-grid :cols="2" :x-gap="12">
            <n-gi>
              <n-form-item label="当前持仓金额">
                <n-input-number v-model:value="positionForm.positionAmount" :min="0" :precision="2" style="width: 100%;" />
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="累计成本金额（可选）">
                <n-input-number v-model:value="positionForm.costAmount" :min="0" :precision="2" style="width: 100%;" />
              </n-form-item>
            </n-gi>
          </n-grid>

          <n-form-item label="备注">
            <n-input v-model:value="positionForm.remark" placeholder="例如 稳健仓 / 养老仓 / 备用金" />
          </n-form-item>

          <div class="entry-note">
            保存时会自动抓盘中估算净值并换算份额。这样你只管“现在有多少钱在这只基金里”，不用自己算份额。
          </div>
        </template>

        <template v-else>
        <n-grid :cols="2" :x-gap="12">
          <n-gi>
            <n-form-item label="品类">
              <n-select v-model:value="form.holdingType" :options="holdingTypeOptions.filter(item => item.value !== 'all')" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="交易方向">
              <n-select v-model:value="form.type" :options="tradeTypeOptions.filter(item => item.value !== 'all')" />
            </n-form-item>
          </n-gi>
        </n-grid>

        <n-grid :cols="2" :x-gap="12">
          <n-gi>
            <n-form-item label="代码">
              <n-input v-model:value="form.stockCode" placeholder="基金填 6 位代码，例如 006327" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="名称">
              <n-input v-model:value="form.stockName" placeholder="例如 某某中短债A" />
            </n-form-item>
          </n-gi>
        </n-grid>

        <n-grid :cols="2" :x-gap="12">
          <n-gi>
            <n-form-item label="平台">
              <n-input v-model:value="form.brokerName" placeholder="默认支付宝" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="账户分组">
              <n-input v-model:value="form.accountTag" placeholder="例如 主账户 / 家庭账户" />
            </n-form-item>
          </n-gi>
        </n-grid>

        <n-grid :cols="3" :x-gap="12">
          <n-gi>
            <n-form-item label="成交净值 / 价格">
              <n-input-number v-model:value="form.price" :min="0" :precision="4" style="width: 100%;" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="份额 / 数量">
              <n-input-number v-model:value="form.quantity" :min="0" :precision="2" style="width: 100%;" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="手续费">
              <n-input-number v-model:value="form.fee" :min="0" :precision="2" style="width: 100%;" />
            </n-form-item>
          </n-gi>
        </n-grid>

        <n-grid :cols="2" :x-gap="12">
          <n-gi>
            <n-form-item label="交易时间">
              <n-date-picker v-model:value="tradeDateTs" type="datetime" style="width: 100%;" clearable />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="备注">
              <n-input v-model:value="form.remark" placeholder="例如 每月定投 / 临时赎回" />
            </n-form-item>
          </n-gi>
        </n-grid>
        </template>
      </n-form>

      <template #action>
        <n-space justify="end">
          <n-button @click="showAddModal = false">取消</n-button>
          <n-button type="primary" :loading="saving" @click="handleCreate">保存</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useMessage } from 'naive-ui'

const message = useMessage()

const keyword = ref('')
const holdingTypeFilter = ref('all')
const tradeTypeFilter = ref('all')
const showAddModal = ref(false)
const saving = ref(false)
const tradeDateTs = ref(Date.now())
const entryMode = ref('position')
const transactions = ref([])
const positionLookupTimer = ref(null)
const tradeLookupTimer = ref(null)

const positionForm = reactive({
  stockCode: '',
  stockName: '',
  positionAmount: 0,
  costAmount: 0,
  brokerName: '支付宝',
  accountTag: '主账户',
  remark: ''
})

const form = reactive({
  stockCode: '',
  stockName: '',
  holdingType: 'fund',
  brokerName: '支付宝',
  accountTag: '主账户',
  type: 'buy',
  price: 0,
  quantity: 0,
  fee: 0,
  remark: ''
})

const holdingTypeOptions = [
  { label: '全部品类', value: 'all' },
  { label: '基金', value: 'fund' },
  { label: '股票', value: 'stock' }
]

const tradeTypeOptions = [
  { label: '全部方向', value: 'all' },
  { label: '买入', value: 'buy' },
  { label: '卖出', value: 'sell' }
]

const filteredTransactions = computed(() => {
  const search = keyword.value.trim().toLowerCase()
  return (transactions.value || []).filter((item) => {
    const matchesKeyword = !search || [item.stockCode, item.stockName, item.brokerName, item.accountTag]
      .filter(Boolean)
      .some((value) => String(value).toLowerCase().includes(search))
    const matchesHoldingType = holdingTypeFilter.value === 'all' || (item.holdingType || 'stock') === holdingTypeFilter.value
    const matchesTradeType = tradeTypeFilter.value === 'all' || item.type === tradeTypeFilter.value
    return matchesKeyword && matchesHoldingType && matchesTradeType
  })
})

const fundTransactionCount = computed(() => filteredTransactions.value.filter((item) => item.holdingType === 'fund').length)
const buyAmount = computed(() => filteredTransactions.value.filter((item) => item.type === 'buy').reduce((sum, item) => sum + Number(item.amount || 0) + Number(item.fee || 0), 0))
const sellAmount = computed(() => filteredTransactions.value.filter((item) => item.type === 'sell').reduce((sum, item) => sum + Number(item.amount || 0), 0))

const columns = [
  {
    title: '时间',
    key: 'tradeDate',
    width: 170,
    render: (row) => formatDateTime(row.tradeDate || row.CreatedAt)
  },
  {
    title: '代码 / 名称',
    key: 'stockName',
    width: 220,
    render: (row) => `${row.stockCode} · ${row.stockName || '-'}`
  },
  {
    title: '品类',
    key: 'holdingType',
    width: 90,
    render: (row) => row.holdingType === 'fund' ? '基金' : '股票'
  },
  {
    title: '方向',
    key: 'type',
    width: 80,
    render: (row) => row.type === 'buy' ? '买入' : '卖出'
  },
  {
    title: '平台 / 账户',
    key: 'brokerName',
    width: 180,
    render: (row) => `${row.brokerName || '-'} / ${row.accountTag || '-'}`
  },
  {
    title: '价格',
    key: 'price',
    width: 100,
    render: (row) => formatPrice(row.price)
  },
  {
    title: '数量',
    key: 'quantity',
    width: 110,
    render: (row) => formatUnits(row.quantity)
  },
  {
    title: '成交金额',
    key: 'amount',
    width: 120,
    render: (row) => formatMoney(row.amount)
  },
  {
    title: '手续费',
    key: 'fee',
    width: 100,
    render: (row) => formatMoney(row.fee)
  },
  {
    title: '备注',
    key: 'remark',
    width: 160,
    render: (row) => row.remark || '-'
  }
]

function formatMoney(value) {
  return `¥${Number(value || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

function formatPrice(value) {
  return Number(value || 0).toFixed(4)
}

function formatUnits(value) {
  return Number(value || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function formatDateTime(value) {
  if (!value) return '-'
  return new Date(value).toLocaleString('zh-CN')
}

function resetForm() {
  Object.assign(positionForm, {
    stockCode: '',
    stockName: '',
    positionAmount: 0,
    costAmount: 0,
    brokerName: '支付宝',
    accountTag: '主账户',
    remark: ''
  })
  Object.assign(form, {
    stockCode: '',
    stockName: '',
    holdingType: 'fund',
    brokerName: '支付宝',
    accountTag: '主账户',
    type: 'buy',
    price: 0,
    quantity: 0,
    fee: 0,
    remark: ''
  })
  tradeDateTs.value = Date.now()
}

async function loadData() {
  if (!window.go?.main?.App?.GetTransactions) return
  const result = await window.go.main.App.GetTransactions('', 1, 500)
  transactions.value = Array.isArray(result) ? (result[0] || []) : []
}

async function autofillFundName(code, target) {
  const trimmed = String(code || '').trim()
  if (!trimmed || trimmed.length < 4 || !window.go?.main?.App?.GetfundList) return
  const result = await window.go.main.App.GetfundList(trimmed)
  const exact = (result || []).find((item) => item.code === trimmed)
  if (exact?.name) {
    target.stockName = exact.name
  }
}

async function handleCreate() {
  if (entryMode.value === 'position') {
    if (!positionForm.stockCode.trim() || !positionForm.positionAmount) {
      message.warning('请先填写基金代码和当前持仓金额')
      return
    }

    saving.value = true
    try {
      const result = await window.go.main.App.UpsertFundHoldingByAmount({ ...positionForm })
      if (!result) {
        message.error('暂时没有拿到这只基金的净值，稍后再试')
        return
      }
      showAddModal.value = false
      resetForm()
      await loadData()
      message.success('当前仓位已保存')
    } catch (error) {
      console.error(error)
      message.error('保存失败')
    } finally {
      saving.value = false
    }
    return
  }

  if (!form.stockCode.trim() || !form.stockName.trim() || !form.price || !form.quantity) {
    message.warning('请先填写完整的交易信息')
    return
  }

  saving.value = true
  try {
    await window.go.main.App.AddTransaction({
      ...form,
      tradeDate: tradeDateTs.value ? new Date(tradeDateTs.value).toISOString() : undefined
    })
    showAddModal.value = false
    resetForm()
    await loadData()
    message.success('交易记录已保存')
  } catch (error) {
    console.error(error)
    message.error('保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(loadData)

watch(() => positionForm.stockCode, (value) => {
  if (positionLookupTimer.value) clearTimeout(positionLookupTimer.value)
  positionLookupTimer.value = setTimeout(() => {
    autofillFundName(value, positionForm)
  }, 220)
})

watch(() => form.stockCode, (value) => {
  if (form.holdingType !== 'fund') return
  if (tradeLookupTimer.value) clearTimeout(tradeLookupTimer.value)
  tradeLookupTimer.value = setTimeout(() => {
    autofillFundName(value, form)
  }, 220)
})
</script>

<style scoped>
.transaction-page {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.summary-card,
.table-card {
  border-radius: var(--radius-lg);
  border: 1px solid rgba(97, 118, 148, 0.24);
  background: linear-gradient(180deg, rgba(14, 24, 39, 0.96), rgba(18, 30, 48, 0.98));
  box-shadow: 0 18px 40px rgba(7, 12, 18, 0.16);
}

.summary-card {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 1px;
  overflow: hidden;
}

.summary-item {
  padding: 18px;
  background: rgba(11, 19, 31, 0.24);
}

.summary-label {
  color: var(--text-secondary);
  font-size: 13px;
}

.summary-value {
  margin-top: 10px;
  font-size: 24px;
  font-weight: 700;
}

.table-card {
  padding: 16px;
}

.entry-note {
  color: var(--text-secondary);
  font-size: 12px;
  line-height: 1.6;
}
</style>
