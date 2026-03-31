<script setup>
import Fund from '../../gostock/components/fund.vue'
</script>

<template>
  <div class="invest-view-shell">
    <Fund />
  </div>
</template>

<style scoped>
.invest-view-shell {
  min-height: calc(100vh - 120px);
}
</style>
