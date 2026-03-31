<script setup>
import GoStockSettings from '../gostock/components/settings.vue'
import AssetSecuritySettings from './AssetSecuritySettings.vue'
import About from '../gostock/components/about.vue'
</script>

<template>
  <div class="fade-in">
    <n-page-header title="Settings" subtitle="Manage runtime options, security access, and project information." />
    <div class="platform-card settings-shell">
      <n-tabs type="line" animated>
        <n-tab-pane name="settings" tab="System">
          <GoStockSettings />
        </n-tab-pane>
        <n-tab-pane name="security" tab="Security">
          <AssetSecuritySettings />
        </n-tab-pane>
        <n-tab-pane name="about" tab="About">
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
