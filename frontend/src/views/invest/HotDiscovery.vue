<script setup>
import { ref } from 'vue'
import HotEvents from '../../gostock/components/HotEvents.vue'
import HotTopics from '../../gostock/components/HotTopics.vue'
import HotStockList from '../../gostock/components/HotStockList.vue'

const activeTab = ref('topics')
</script>

<template>
  <div class="fade-in">
    <n-page-header class="view-header" title="热点发现" subtitle="聚合市场热事件、热话题和热股追踪" />

    <div class="platform-card tabs-shell">
      <n-tabs v-model:value="activeTab" type="line" animated pane-style="padding: 20px;">
        <n-tab-pane name="topics" tab="热话题">
          <HotTopics />
        </n-tab-pane>
        <n-tab-pane name="events" tab="热事件">
          <HotEvents />
        </n-tab-pane>
        <n-tab-pane name="stocks" tab="热股榜">
          <HotStockList />
        </n-tab-pane>
      </n-tabs>
    </div>
  </div>
</template>
