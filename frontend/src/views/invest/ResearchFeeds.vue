<script setup>
import { ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import StockNoticeList from '../../gostock/components/StockNoticeList.vue'
import StockResearchReportList from '../../gostock/components/StockResearchReportList.vue'
import IndustryResearchReportList from '../../gostock/components/IndustryResearchReportList.vue'

const route = useRoute()
const activeTab = ref('notices')
const stockCode = ref('')

function syncFromQuery() {
  const tab = String(route.query.tab || '')
  const queryStockCode = String(route.query.stockCode || '')

  if (['notices', 'stock-reports', 'industry-reports'].includes(tab)) {
    activeTab.value = tab
  }

  stockCode.value = queryStockCode
}

syncFromQuery()

watch(
  () => route.query,
  () => {
    syncFromQuery()
  }
)
</script>

<template>
  <div class="fade-in">
    <n-page-header class="view-header" title="公告与研报" subtitle="集中查看公告、个股研报和行业研报" />

    <div class="platform-card tabs-shell">
      <n-tabs v-model:value="activeTab" type="line" animated pane-style="padding: 20px;">
        <n-tab-pane name="notices" tab="公告">
          <StockNoticeList :stock-code="stockCode" />
        </n-tab-pane>
        <n-tab-pane name="stock-reports" tab="个股研报">
          <StockResearchReportList :stock-code="stockCode" />
        </n-tab-pane>
        <n-tab-pane name="industry-reports" tab="行业研报">
          <IndustryResearchReportList />
        </n-tab-pane>
      </n-tabs>
    </div>
  </div>
</template>
