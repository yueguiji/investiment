<script setup>
import PromptTemplateList from '../gostock/components/promptTemplateList.vue'
</script>

<template>
  <div class="fade-in">
    <n-page-header title="提示词模板" subtitle="管理系统 Prompt 与用户 Prompt 模板" />
    <div class="platform-card" style="margin-top: 20px;">
      <PromptTemplateList />
    </div>
  </div>
</template>
