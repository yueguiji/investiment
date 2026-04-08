<template>
  <n-config-provider :theme="darkTheme" :locale="zhCN" :date-locale="dateZhCN">
    <n-message-provider>
      <n-notification-provider>
        <n-dialog-provider>
          <n-layout has-sider style="height: 100vh;">
            <n-layout-sider
              bordered
              :collapsed="collapsed"
              collapse-mode="width"
              :collapsed-width="64"
              :width="228"
              :native-scrollbar="false"
              :style="{ background: 'var(--bg-sidebar)' }"
            >
              <div class="sidebar-logo" @click="collapsed = !collapsed">
                <img class="logo-icon" :src="brandIcon" alt="Rubin Investment" />
                <transition name="fade">
                  <span v-if="!collapsed" class="logo-text">Rubin Investment</span>
                </transition>
              </div>

              <n-menu
                :collapsed="collapsed"
                :collapsed-width="64"
                :collapsed-icon-size="22"
                :options="menuOptions"
                :value="activeMenu"
                @update:value="handleMenuClick"
              />

              <div class="sidebar-toggle" @click="collapsed = !collapsed">
                <n-icon size="18">
                  <ChevronBackOutline v-if="!collapsed" />
                  <ChevronForwardOutline v-else />
                </n-icon>
              </div>
            </n-layout-sider>

            <n-layout :native-scrollbar="false">
              <n-layout-header bordered class="app-header">
                <n-breadcrumb>
                  <n-breadcrumb-item v-for="item in breadcrumbs" :key="item.path">
                    {{ item.title }}
                  </n-breadcrumb-item>
                </n-breadcrumb>
                <div class="header-actions">
                  <n-tag :type="marketOpen ? 'success' : 'default'" size="small" round>
                    {{ marketOpen ? '交易时段' : '非交易时段' }}
                  </n-tag>
                  <n-button quaternary circle size="small" @click="$router.push('/settings')">
                    <template #icon>
                      <n-icon><SettingsOutline /></n-icon>
                    </template>
                  </n-button>
                </div>
              </n-layout-header>

              <n-layout-content class="app-content">
                <router-view v-slot="{ Component }">
                  <transition name="page" mode="out-in">
                    <component :is="Component" />
                  </transition>
                </router-view>
              </n-layout-content>
            </n-layout>
          </n-layout>
        </n-dialog-provider>
      </n-notification-provider>
    </n-message-provider>
  </n-config-provider>
</template>

<script setup>
import { computed, h, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { dateZhCN, darkTheme, NIcon, zhCN } from 'naive-ui'
import brandIcon from './assets/rubin-cat.png'
import {
  AnalyticsOutline,
  BarChartOutline,
  BriefcaseOutline,
  CashOutline,
  ChatbubblesOutline,
  ChevronBackOutline,
  ChevronForwardOutline,
  CodeSlashOutline,
  CreateOutline,
  DiamondOutline,
  EyeOutline,
  FlashOutline,
  FunnelOutline,
  GridOutline,
  InformationCircleOutline,
  LockClosedOutline,
  LinkOutline,
  ListOutline,
  PieChartOutline,
  PeopleOutline,
  SettingsOutline,
  StatsChartOutline,
  StorefrontOutline,
  TimeOutline,
  TrendingUpOutline,
  WalletOutline,
  BulbOutline
} from '@vicons/ionicons5'
import { routes } from './router'

const route = useRoute()
const router = useRouter()
const collapsed = ref(false)

const iconMap = {
  AnalyticsOutline,
  BarChartOutline,
  BriefcaseOutline,
  BulbOutline,
  CashOutline,
  ChatbubblesOutline,
  CodeSlashOutline,
  CreateOutline,
  DiamondOutline,
  EyeOutline,
  FlashOutline,
  FunnelOutline,
  GridOutline,
  InformationCircleOutline,
  LockClosedOutline,
  LinkOutline,
  ListOutline,
  PieChartOutline,
  PeopleOutline,
  SettingsOutline,
  StatsChartOutline,
  StorefrontOutline,
  TimeOutline,
  TrendingUpOutline,
  WalletOutline
}

const activeMenu = computed(() => route.path)

const breadcrumbs = computed(() =>
  route.matched
    .filter((item) => item.meta?.title)
    .map((item) => ({
      path: item.path,
      title: item.meta.title
    }))
)

const marketOpen = computed(() => {
  const now = new Date()
  const day = now.getDay()
  if (day === 0 || day === 6) {
    return false
  }
  const totalMinutes = now.getHours() * 60 + now.getMinutes()
  const morning = totalMinutes >= 9 * 60 + 15 && totalMinutes <= 11 * 60 + 30
  const afternoon = totalMinutes >= 13 * 60 && totalMinutes <= 15 * 60
  return morning || afternoon
})

function renderIcon(iconName) {
  const icon = iconMap[iconName]
  if (!icon) {
    return undefined
  }
  return () => h(NIcon, null, { default: () => h(icon) })
}

function toMenuOption(record) {
  const option = {
    label: record.meta?.title || record.name,
    key: record.redirect || record.path,
    icon: renderIcon(record.meta?.icon)
  }

  if (record.children?.length) {
    option.children = record.children
      .filter((child) => child.meta?.title && !child.meta?.hideInMenu && !child.path.includes(':'))
      .map((child) => ({
        label: child.meta.title,
        key: record.path + '/' + child.path.replace(/^\/+/, ''),
        icon: renderIcon(child.meta.icon)
      }))
  }

  return option
}

const menuOptions = computed(() =>
  routes
    .filter((record) => record.meta?.title)
    .map((record) => toMenuOption(record))
)

function handleMenuClick(key) {
  router.push(key).catch((err) => {
    console.error('Menu navigation failed:', key, err)
  })
}
</script>

<style scoped>
.sidebar-logo {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px 20px;
  cursor: pointer;
  border-bottom: 1px solid var(--border-color);
  user-select: none;
}

.logo-icon {
  width: 30px;
  height: 30px;
  border-radius: 10px;
  object-fit: cover;
  flex-shrink: 0;
  box-shadow: 0 6px 18px rgba(15, 23, 42, 0.22);
  border: 1px solid rgba(255, 255, 255, 0.16);
}

.logo-text {
  font-size: 15px;
  font-weight: 600;
  color: #f8fafc;
  white-space: nowrap;
}

.sidebar-toggle {
  position: absolute;
  bottom: 12px;
  left: 0;
  right: 0;
  display: flex;
  justify-content: center;
  padding: 8px;
  cursor: pointer;
  opacity: 0.5;
  transition: opacity 0.2s;
}

.sidebar-toggle:hover {
  opacity: 1;
}

.app-header {
  padding: 0 24px;
  height: 52px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--bg-secondary);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.app-content {
  padding: 20px;
  background: var(--bg-primary);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
