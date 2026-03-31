<template>
  <div class="fade-in">
    <n-page-header title="交易记录" subtitle="录入并查看买卖交易流水">
      <template #extra>
        <n-space>
          <n-input v-model:value="filterCode" clearable placeholder="按代码筛选" style="width: 180px" />
          <n-button @click="loadData">刷新</n-button>
          <n-button type="primary" @click="showAddModal = true">新增交易</n-button>
        </n-space>
      </template>
    </n-page-header>

    <div class="platform-card" style="margin-top: 20px; padding: 0;">
      <n-data-table :columns="columns" :data="transactions" :pagination="{ pageSize: 15 }" striped />
    </div>

    <n-modal v-model:show="showAddModal" preset="card" title="新增交易记录" style="width: 560px;">
      <n-form label-placement="top">
        <n-grid :cols="2" :x-gap="12">
          <n-gi>
            <n-form-item label="代码">
              <n-input v-model:value="form.stockCode" placeholder="如：sh600519" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="名称">
              <n-input v-model:value="form.stockName" placeholder="如：贵州茅台" />
            </n-form-item>
          </n-gi>
        </n-grid>
        <n-grid :cols="2" :x-gap="12">
          <n-gi>
            <n-form-item label="类型">
              <n-select v-model:value="form.type" :options="typeOptions" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="交易日期">
              <n-date-picker v-model:value="tradeDateTs" type="datetime" style="width: 100%;" clearable />
            </n-form-item>
          </n-gi>
        </n-grid>
        <n-grid :cols="2" :x-gap="12">
          <n-gi>
            <n-form-item label="价格">
              <n-input-number v-model:value="form.price" :min="0" :precision="3" style="width: 100%;" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="数量">
              <n-input-number v-model:value="form.quantity" :min="0" style="width: 100%;" />
            </n-form-item>
          </n-gi>
        </n-grid>
        <n-grid :cols="2" :x-gap="12">
          <n-gi>
            <n-form-item label="手续费">
              <n-input-number v-model:value="form.fee" :min="0" :precision="2" style="width: 100%;" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="备注">
              <n-input v-model:value="form.remark" />
            </n-form-item>
          </n-gi>
        </n-grid>
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
import { onMounted, reactive, ref, watch } from 'vue'
import { useMessage } from 'naive-ui'

const message = useMessage()
const transactions = ref([])
const filterCode = ref('')
const showAddModal = ref(false)
const saving = ref(false)
const tradeDateTs = ref(null)

const form = reactive({
  stockCode: '',
  stockName: '',
  type: 'buy',
  price: 0,
  quantity: 0,
  fee: 0,
  remark: ''
})

const typeOptions = [
  { label: '买入', value: 'buy' },
  { label: '卖出', value: 'sell' }
]

const columns = [
  { title: '日期', key: 'tradeDate', width: 160, render: (row) => formatDate(row.tradeDate || row.CreatedAt) },
  { title: '代码', key: 'stockCode', width: 110 },
  { title: '名称', key: 'stockName', width: 140 },
  { title: '类型', key: 'type', width: 80, render: (row) => row.type === 'buy' ? '买入' : '卖出' },
  { title: '价格', key: 'price', width: 100, render: (row) => formatPrice(row.price) },
  { title: '数量', key: 'quantity', width: 90 },
  { title: '成交额', key: 'amount', width: 120, render: (row) => formatPrice(row.amount) },
  { title: '手续费', key: 'fee', width: 100, render: (row) => formatPrice(row.fee) },
  { title: '备注', key: 'remark' }
]

function formatPrice(value) {
  return `¥ ${Number(value || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

function formatDate(value) {
  if (!value) return '-'
  return new Date(value).toLocaleString('zh-CN')
}

async function loadData() {
  if (!window.go?.main?.App?.GetTransactions) return
  const result = await window.go.main.App.GetTransactions(filterCode.value || '', 1, 200)
  transactions.value = Array.isArray(result) ? result[0] || [] : []
}

async function handleCreate() {
  if (!form.stockCode.trim() || !form.stockName.trim() || !form.price || !form.quantity) {
    message.warning('请先填写完整交易信息')
    return
  }
  saving.value = true
  try {
    await window.go.main.App.AddTransaction({
      ...form,
      tradeDate: tradeDateTs.value ? new Date(tradeDateTs.value).toISOString() : undefined
    })
    message.success('交易记录已保存')
    showAddModal.value = false
    Object.assign(form, { stockCode: '', stockName: '', type: 'buy', price: 0, quantity: 0, fee: 0, remark: '' })
    tradeDateTs.value = null
    await loadData()
  } catch (error) {
    console.error(error)
    message.error('保存失败')
  } finally {
    saving.value = false
  }
}

watch(filterCode, () => {
  loadData()
})

onMounted(loadData)
</script>
