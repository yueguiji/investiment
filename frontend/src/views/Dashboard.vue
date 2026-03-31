<template>
  <div class="dashboard fade-in">
    <h2 style="margin-bottom: 20px; font-weight: 600;">
      投资总览
      <span style="font-size: 13px; color: var(--text-muted); font-weight: 400; margin-left: 12px;">
        {{ new Date().toLocaleDateString('zh-CN', { year: 'numeric', month: 'long', day: 'numeric', weekday: 'long' }) }}
      </span>
    </h2>

    <n-grid :cols="4" :x-gap="16" :y-gap="16" style="margin-bottom: 24px;">
      <n-gi>
        <div class="metric-card" style="--accent: var(--primary);">
          <div class="metric-label">净资产</div>
          <div class="metric-value">¥ {{ formatNum(dashData.netAsset) }}</div>
          <div class="metric-sub">含投资市值</div>
        </div>
      </n-gi>
      <n-gi>
        <div class="metric-card" :style="{ '--accent': dashData.totalProfit >= 0 ? 'var(--profit)' : 'var(--loss)' }">
          <div class="metric-label">投资总盈亏</div>
          <div class="metric-value" :class="dashData.totalProfit >= 0 ? 'profit' : 'loss'">
            {{ dashData.totalProfit >= 0 ? '+' : '' }}¥ {{ formatNum(dashData.totalProfit) }}
          </div>
          <div class="metric-sub">收益率 {{ dashData.totalProfitRate }}%</div>
        </div>
      </n-gi>
      <n-gi>
        <div class="metric-card" :style="{ '--accent': dashData.todayProfit >= 0 ? 'var(--profit)' : 'var(--loss)' }">
          <div class="metric-label">今日盈亏</div>
          <div class="metric-value" :class="dashData.todayProfit >= 0 ? 'profit' : 'loss'">
            {{ dashData.todayProfit >= 0 ? '+' : '' }}¥ {{ formatNum(dashData.todayProfit) }}
          </div>
          <div class="metric-sub">{{ dashData.stockCount }} 股票 / {{ dashData.fundCount }} 基金</div>
        </div>
      </n-gi>
      <n-gi>
        <div class="metric-card" style="--accent: var(--info);">
          <div class="metric-label">量化策略</div>
          <div class="metric-value">{{ dashData.activeStrategies }}</div>
          <div class="metric-sub">{{ dashData.totalStrategies }} 个模板</div>
        </div>
      </n-gi>
    </n-grid>

    <n-grid :cols="2" :x-gap="16" :y-gap="16">
      <n-gi>
        <div class="module-card" @click="$router.push('/portfolio/overview')">
          <div class="module-header">
            <span class="module-icon">持</span>
            <span class="module-title">持仓概览</span>
            <n-tag size="tiny" :type="dashData.todayProfit >= 0 ? 'success' : 'error'" round>
              {{ dashData.todayProfit >= 0 ? '盈' : '亏' }}
            </n-tag>
          </div>
          <div class="module-desc">
            持有 {{ dashData.stockCount + dashData.fundCount }} 个标的，总市值 ¥{{ formatNum(dashData.totalValue) }}
          </div>
        </div>
      </n-gi>
      <n-gi>
        <div class="module-card" @click="$router.push('/invest/monitor')">
          <div class="module-header">
            <span class="module-icon">投</span>
            <span class="module-title">投资分析</span>
            <n-tag size="tiny" type="info" round>go-stock</n-tag>
          </div>
          <div class="module-desc">
            AI 赋能股票分析，支持 A 股、港股、美股实时行情与智能选股。
          </div>
        </div>
      </n-gi>
      <n-gi>
        <div class="module-card" @click="$router.push('/asset/overview')">
          <div class="module-header">
            <span class="module-icon">资</span>
            <span class="module-title">资产分析</span>
            <n-tag size="tiny" type="warning" round>基础版</n-tag>
          </div>
          <div class="module-desc">
            管理流动资产、固定资产和负债，全面掌握个人财务状况。
          </div>
        </div>
      </n-gi>
      <n-gi>
        <div class="module-card" @click="$router.push('/quant/templates')">
          <div class="module-header">
            <span class="module-icon">量</span>
            <span class="module-title">量化模板</span>
            <n-tag size="tiny" type="default" round>{{ dashData.activeStrategies }} 运行中</n-tag>
          </div>
          <div class="module-desc">
            AI 生成 Python 量化策略，并与投资分析模块联动。
          </div>
        </div>
      </n-gi>
    </n-grid>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const dashData = ref({
  netAsset: 0,
  totalProfit: 0,
  totalProfitRate: 0,
  todayProfit: 0,
  totalValue: 0,
  stockCount: 0,
  fundCount: 0,
  activeStrategies: 0,
  totalStrategies: 0
})

function formatNum(num) {
  return Number(num || 0).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

onMounted(async () => {
  try {
    if (window.go?.main?.App?.GetPortfolioSummary) {
      const summary = await window.go.main.App.GetPortfolioSummary()
      if (summary) {
        dashData.value.totalProfit = summary.totalProfit
        dashData.value.totalProfitRate = summary.totalProfitRate
        dashData.value.todayProfit = summary.todayProfit
        dashData.value.totalValue = summary.totalValue
        dashData.value.stockCount = summary.stockCount
        dashData.value.fundCount = summary.fundCount
      }
    }

    if (window.go?.main?.App?.GetAssetSummary) {
      const assetSummary = await window.go.main.App.GetAssetSummary()
      if (assetSummary) {
        dashData.value.netAsset = assetSummary.netAsset + (dashData.value.totalValue || 0)
      }
    }
  } catch (e) {
    console.log('Dashboard data loading...', e)
  }
})
</script>

<style scoped>
.dashboard {
  max-width: 1200px;
  margin: 0 auto;
}

.metric-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 20px;
  position: relative;
  overflow: hidden;
  transition: all 0.3s ease;
}

.metric-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: var(--accent);
  border-radius: var(--radius-md) var(--radius-md) 0 0;
}

.metric-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.metric-label {
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 8px;
}

.metric-value {
  font-size: 24px;
  font-weight: 700;
  font-family: var(--font-mono);
  margin-bottom: 4px;
}

.metric-sub {
  font-size: 12px;
  color: var(--text-muted);
}

.module-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 20px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.module-card:hover {
  border-color: var(--primary);
  box-shadow: var(--shadow-glow);
  transform: translateY(-2px);
}

.module-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
}

.module-icon {
  font-size: 20px;
}

.module-title {
  font-size: 15px;
  font-weight: 600;
  flex: 1;
}

.module-desc {
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.6;
}
</style>
