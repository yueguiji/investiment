<script setup>
import AllStockList from '../../gostock/components/allStockList.vue'
</script>

<template>
  <div class="invest-view-shell">
    <AllStockList />
  </div>
</template>

<style scoped>
.invest-view-shell {
  min-height: calc(100vh - 120px);
}
</style>
