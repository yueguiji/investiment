import { createApp } from 'vue'
import { createRouter, createMemoryHistory, createWebHashHistory } from 'vue-router'
import App from './App.vue'
import { routes } from './router/index.js'

import './style.css'
import 'tdesign-vue-next/es/style/index.css'

function createDesktopAwareHistory() {
  const protocol = window.location?.protocol || ''
  const isDesktopRuntime = protocol !== 'http:' && protocol !== 'https:'
  return isDesktopRuntime ? createMemoryHistory() : createWebHashHistory()
}

const router = createRouter({
  history: createDesktopAwareHistory(),
  routes
})

const assetUnlockKey = 'investment.asset.unlocked'
const assetUnlockRoute = '/asset/unlock'
let assetUnlockRequiredPromise = null

async function isAssetUnlockRequired() {
  if (!assetUnlockRequiredPromise) {
    assetUnlockRequiredPromise = Promise.resolve()
      .then(async () => {
        const getVersionInfo = window.go?.main?.App?.GetVersionInfo
        if (!getVersionInfo) {
          return false
        }
        const versionInfo = await getVersionInfo()
        return Boolean(versionInfo?.assetUnlockEnabled)
      })
      .catch(() => false)
  }
  return assetUnlockRequiredPromise
}

router.beforeEach(async (to) => {
  if (!to.path.startsWith('/asset')) {
    return true
  }

  const unlockRequired = await isAssetUnlockRequired()
  if (!unlockRequired) {
    return true
  }

  const isUnlockRoute = to.path === assetUnlockRoute
  const unlocked = window.sessionStorage.getItem(assetUnlockKey) === 'granted'

  if (!unlocked && !isUnlockRoute) {
    return {
      path: assetUnlockRoute,
      query: { redirect: to.fullPath }
    }
  }

  if (unlocked && isUnlockRoute) {
    return to.query.redirect || '/asset/overview'
  }

  return true
})

function showRuntimeError(message) {
  if (shouldIgnoreRuntimeError(message)) {
    return
  }
  let el = document.getElementById('runtime-error-overlay')
  if (!el) {
    el = document.createElement('pre')
    el.id = 'runtime-error-overlay'
    el.style.position = 'fixed'
    el.style.top = '12px'
    el.style.right = '12px'
    el.style.zIndex = '99999'
    el.style.maxWidth = '50vw'
    el.style.maxHeight = '40vh'
    el.style.overflow = 'auto'
    el.style.margin = '0'
    el.style.padding = '12px'
    el.style.whiteSpace = 'pre-wrap'
    el.style.background = 'rgba(120, 12, 12, 0.92)'
    el.style.color = '#fff'
    el.style.fontSize = '12px'
    el.style.lineHeight = '1.4'
    el.style.border = '1px solid rgba(255,255,255,0.2)'
    el.style.borderRadius = '8px'
    document.body.appendChild(el)
  }
  el.textContent = String(message)
}

function shouldIgnoreRuntimeError(message) {
  const text = String(message || '')
  return (
    text.includes('ResizeObserver loop completed with undelivered notifications') ||
    text.includes('ResizeObserver loop limit exceeded')
  )
}

window.addEventListener('error', (event) => {
  if (shouldIgnoreRuntimeError(event.error?.stack || event.message)) {
    event.preventDefault()
    return
  }
  showRuntimeError(event.error?.stack || event.message)
})

window.addEventListener('unhandledrejection', (event) => {
  if (shouldIgnoreRuntimeError(event.reason?.stack || event.reason)) {
    event.preventDefault()
    return
  }
  showRuntimeError(event.reason?.stack || event.reason)
})

async function bootstrap() {
  const app = createApp(App)
  app.config.errorHandler = (err) => {
    console.error(err)
    if (shouldIgnoreRuntimeError(err?.stack || err?.message || err)) {
      return
    }
    showRuntimeError(err?.stack || err)
  }
  app.use(router)
  await router.isReady().catch((err) => {
    console.error(err)
    if (shouldIgnoreRuntimeError(err?.stack || err?.message || err)) {
      return
    }
    showRuntimeError(err?.stack || err)
  })
  app.mount('#app')
}

bootstrap()
