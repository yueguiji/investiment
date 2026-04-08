<script setup>
import GoStockSettings from '../gostock/components/settings.vue'
import AssetSecuritySettings from './AssetSecuritySettings.vue'
import About from '../gostock/components/about.vue'
</script>

<template>
  <div class="fade-in">
    <n-page-header title="设置" subtitle="管理运行时选项、安全访问控制和项目信息。" />
    <div class="platform-card settings-shell">
      <n-tabs type="line" animated>
        <n-tab-pane name="settings" tab="系统设置">
          <GoStockSettings />
        </n-tab-pane>
        <n-tab-pane name="security" tab="安全设置">
          <AssetSecuritySettings />
        </n-tab-pane>
        <n-tab-pane name="about" tab="关于">
          <About />
        </n-tab-pane>
      </n-tabs>
    </div>
  </div>
</template>

<style scoped>
.settings-shell {
  margin-top: 20px;
  padding: 16px;
}
</style>
