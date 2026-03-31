<script setup>
import LongTigerRankList from '../../gostock/components/LongTigerRankList.vue'
</script>

<template>
  <div class="fade-in">
    <n-page-header title="龙虎榜" subtitle="查看龙虎榜席位与上榜原因" />

    <div class="platform-card" style="margin-top: 20px; padding: 20px 20px 0;">
      <LongTigerRankList />
    </div>
  </div>
</template>
