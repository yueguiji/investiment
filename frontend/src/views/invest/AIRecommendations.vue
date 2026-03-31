<script setup>
import AIRecommendStocksList from '../../gostock/components/aiRecommendStocksList.vue'
</script>

<template>
  <div class="invest-view-shell">
    <AIRecommendStocksList />
  </div>
</template>

<style scoped>
.invest-view-shell {
  min-height: calc(100vh - 120px);
}
</style>
