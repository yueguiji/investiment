<script setup>
import { ref } from 'vue'
import RankTable from '../../gostock/components/rankTable.vue'

const activeTab = ref('netamount')

const tabOptions = [
  { name: 'netamount', label: '净流入额', sort: 'netamount' },
  { name: 'changepercent', label: '涨跌幅', sort: 'changepercent' },
  { name: 'r0_net', label: '主力净流入', sort: 'r0_net' },
  { name: 'r3_net', label: '散户净流入', sort: 'r3_net' }
]
</script>

<template>
  <div class="fade-in">
    <n-page-header class="view-header" title="资金排行" subtitle="按资金流向和涨跌强度筛选强势标的" />

    <div class="platform-card tabs-shell">
      <n-tabs v-model:value="activeTab" type="line" animated pane-style="padding: 20px;">
        <n-tab-pane
          v-for="item in tabOptions"
          :key="item.name"
          :name="item.name"
          :tab="item.label"
        >
          <RankTable :header-title="item.label" :sort="item.sort" />
        </n-tab-pane>
      </n-tabs>
    </div>
  </div>
</template>
