<script setup>
import Stock from '../../gostock/components/stock.vue'
</script>

<template>
  <div class="invest-view-shell">
    <Stock />
  </div>
</template>

<style scoped>
.invest-view-shell {
  min-height: calc(100vh - 120px);
}
</style>
