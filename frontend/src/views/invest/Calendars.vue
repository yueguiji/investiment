<script setup>
import { ref } from 'vue'
import ClsCalendarTimeLine from '../../gostock/components/ClsCalendarTimeLine.vue'
import InvestCalendarTimeLine from '../../gostock/components/InvestCalendarTimeLine.vue'

const activeTab = ref('cls')
</script>

<template>
  <div class="fade-in">
    <n-page-header class="view-header" title="投资日历" subtitle="查看财联社与财经日历时间线" />

    <div class="platform-card tabs-shell">
      <n-tabs v-model:value="activeTab" type="line" animated pane-style="padding: 20px;">
        <n-tab-pane name="cls" tab="财联社日历">
          <ClsCalendarTimeLine />
        </n-tab-pane>
        <n-tab-pane name="invest" tab="财经日历">
          <InvestCalendarTimeLine />
        </n-tab-pane>
      </n-tabs>
    </div>
  </div>
</template>
