<script setup>
import AllStockInfoList from '../../gostock/components/allStockInfoList.vue'
</script>

<template>
  <div class="invest-view-shell">
    <AllStockInfoList />
  </div>
</template>

<style scoped>
.invest-view-shell {
  min-height: calc(100vh - 120px);
}
</style>
