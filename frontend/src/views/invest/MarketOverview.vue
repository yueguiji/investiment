<script setup>
import Market from '../../gostock/components/market.vue'
</script>

<template>
  <div class="invest-view-shell">
    <Market />
  </div>
</template>

<style scoped>
.invest-view-shell {
  min-height: calc(100vh - 120px);
}
</style>
