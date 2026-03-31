<template>
  <div class="fade-in asset-unlock-page">
    <div class="unlock-shell">
      <div class="unlock-badge">资产分析已加锁</div>
      <h1>输入访问密码</h1>
      <p>
        为了保护家庭资产、负债和数字分析内容，进入资产分析模块前需要先解锁一次。
      </p>

      <n-alert v-if="errorMessage" type="error" :show-icon="false" class="unlock-alert">
        {{ errorMessage }}
      </n-alert>

      <n-form label-placement="top">
        <n-form-item label="访问密码">
          <n-input
            v-model:value="password"
            type="password"
            show-password-on="click"
            placeholder="请输入资产分析访问密码"
            :disabled="checkingConfig"
            @keydown.enter.prevent="handleUnlock"
          />
        </n-form-item>
      </n-form>

      <div class="unlock-actions">
        <n-button @click="$router.push('/dashboard')">返回仪表盘</n-button>
        <n-button type="primary" :loading="checkingConfig" @click="handleUnlock">解锁进入</n-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const assetUnlockKey = 'investment.asset.unlocked'
const password = ref('')
const errorMessage = ref('')
const checkingConfig = ref(true)

function grantAccess() {
  window.sessionStorage.setItem(assetUnlockKey, 'granted')
  errorMessage.value = ''
  router.push(String(route.query.redirect || '/asset/overview'))
}

onMounted(async () => {
  try {
    const getVersionInfo = window.go?.main?.App?.GetVersionInfo
    if (!getVersionInfo) {
      return
    }

    const versionInfo = await getVersionInfo()
    if (!versionInfo?.assetUnlockEnabled) {
      grantAccess()
      return
    }
  } catch (error) {
    console.error(error)
  } finally {
    checkingConfig.value = false
  }
})

async function handleUnlock() {
  if (checkingConfig.value) {
    return
  }

  const verifyAssetUnlockPassword = window.go?.main?.App?.VerifyAssetUnlockPassword
  if (!verifyAssetUnlockPassword) {
    errorMessage.value = '当前运行环境不支持密码校验。'
    return
  }

  const passed = await verifyAssetUnlockPassword((password.value || '').trim())
  if (!passed) {
    errorMessage.value = '密码不正确，请重新输入。'
    return
  }

  grantAccess()
}
</script>

<style scoped>
.asset-unlock-page {
  min-height: calc(100vh - 120px);
  display: grid;
  place-items: center;
}

.unlock-shell {
  width: min(460px, 100%);
  padding: 32px;
  border-radius: 24px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  background:
    radial-gradient(circle at top right, rgba(56, 189, 248, 0.14), transparent 38%),
    linear-gradient(180deg, rgba(15, 23, 42, 0.96), rgba(30, 41, 59, 0.88));
  box-shadow: 0 28px 60px rgba(2, 6, 23, 0.28);
}

.unlock-badge {
  display: inline-flex;
  padding: 6px 12px;
  border-radius: 999px;
  background: rgba(59, 130, 246, 0.14);
  border: 1px solid rgba(59, 130, 246, 0.28);
  color: #bfdbfe;
  font-size: 12px;
}

h1 {
  margin: 18px 0 12px;
  font-size: 30px;
  line-height: 1.2;
}

p {
  margin: 0 0 18px;
  color: var(--text-secondary);
  line-height: 1.7;
}

.unlock-alert {
  margin-bottom: 14px;
}

.unlock-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>
