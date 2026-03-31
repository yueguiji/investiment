<script setup>
import { h, onMounted, ref } from 'vue'
import { NTag, useMessage } from 'naive-ui'
import { GetConfig, UpdateConfig } from '../../wailsjs/go/main/App'
import { data } from '../../wailsjs/go/models'
import { EventsEmit } from '../../wailsjs/runtime'

const message = useMessage()
const loading = ref(false)
const configId = ref(1)
const assetUnlockPassword = ref('')

async function loadConfig() {
  const config = await GetConfig()
  configId.value = config.ID || 1
  assetUnlockPassword.value = config.assetUnlockPassword || ''
}

async function saveSecurityConfig() {
  loading.value = true
  try {
    const current = await GetConfig()
    const nextConfig = new data.SettingConfig({
      ...current,
      ID: current.ID || configId.value,
      aiConfigs: current.aiConfigs || [],
      assetUnlockPassword: assetUnlockPassword.value || '',
    })
    const saveMessage = await UpdateConfig(nextConfig)
    message.success(saveMessage)
    EventsEmit('updateSettings', {
      ...current,
      assetUnlockPassword: assetUnlockPassword.value || '',
    })
  } catch (error) {
    message.error(error?.message || 'Failed to save security settings')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadConfig().catch((error) => {
    console.error(error)
    message.error(error?.message || 'Failed to load security settings')
  })
})
</script>

<template>
  <div class="security-shell">
    <n-space vertical size="large">
      <n-card :title="() => h(NTag, { type: 'warning', bordered: false }, () => 'Security')" size="small">
        <n-space vertical size="large">
          <n-alert type="info" :show-icon="false">
            Use this password to protect access to the asset analysis area. Leave it blank to disable the lock.
          </n-alert>

          <n-form label-placement="left" label-align="left">
            <n-grid :cols="24" :x-gap="24">
              <n-form-item-gi :span="12" label="Asset access password" path="assetUnlockPassword">
                <n-input
                  v-model:value="assetUnlockPassword"
                  type="password"
                  show-password-on="click"
                  clearable
                  placeholder="Leave blank to disable the asset access password"
                />
              </n-form-item-gi>
            </n-grid>
          </n-form>

          <n-space justify="end">
            <n-button type="primary" :loading="loading" @click="saveSecurityConfig">Save security settings</n-button>
          </n-space>
        </n-space>
      </n-card>
    </n-space>
  </div>
</template>

<style scoped>
.security-shell {
  text-align: left;
}
</style>
