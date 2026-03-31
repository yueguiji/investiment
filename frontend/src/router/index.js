import Dashboard from '../views/Dashboard.vue'
import Settings from '../views/Settings.vue'
import About from '../views/About.vue'
import PromptTemplates from '../views/PromptTemplates.vue'

import AssetOverview from '../views/asset/AssetOverview.vue'
import AssetDetail from '../views/asset/AssetDetail.vue'
import AssetDebtPlans from '../views/asset/AssetDebtPlans.vue'
import DigitalAnalysis from '../views/asset/DigitalAnalysis.vue'
import AssetBenchmarks from '../views/asset/AssetBenchmarks.vue'
import AssetMembers from '../views/asset/AssetMembers.vue'
import AssetUnlock from '../views/asset/AssetUnlock.vue'

import StockMonitor from '../views/invest/StockMonitor.vue'
import MarketOverview from '../views/invest/MarketOverview.vue'
import FundWatch from '../views/invest/FundWatch.vue'
import ResearchCenter from '../views/invest/AIAnalysis.vue'
import AgentChat from '../views/invest/AgentChat.vue'
import AllStocks from '../views/invest/AllStocks.vue'
import AIRecommendations from '../views/invest/AIRecommendations.vue'
import StockLibrary from '../views/invest/StockLibrary.vue'
import ResearchFeeds from '../views/invest/ResearchFeeds.vue'
import HotDiscovery from '../views/invest/HotDiscovery.vue'
import Calendars from '../views/invest/Calendars.vue'
import LongTiger from '../views/invest/LongTiger.vue'
import MoneyRank from '../views/invest/MoneyRank.vue'

import PortfolioOverview from '../views/portfolio/PortfolioOverview.vue'
import StockHoldings from '../views/portfolio/StockHoldings.vue'
import FundHoldings from '../views/portfolio/FundHoldings.vue'
import ProfitHistory from '../views/portfolio/ProfitHistory.vue'
import Transactions from '../views/portfolio/Transactions.vue'

import TemplateList from '../views/quant/TemplateList.vue'
import TemplateEditor from '../views/quant/TemplateEditor.vue'
import AIGenerator from '../views/quant/AIGenerator.vue'
import StrategyLinkage from '../views/quant/StrategyLinkage.vue'
import ScriptSearch from '../views/quant/ScriptSearch.vue'

export const routes = [
  {
    path: '/',
    redirect: '/invest/monitor'
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: Dashboard,
    meta: { title: '仪表盘', icon: 'GridOutline' }
  },
  {
    path: '/asset',
    name: 'Asset',
    redirect: '/asset/overview',
    meta: { title: '资产分析', icon: 'DiamondOutline' },
    children: [
      {
        path: 'overview',
        name: 'AssetOverview',
        component: AssetOverview,
        meta: { title: '资产总览', icon: 'PieChartOutline' }
      },
      {
        path: 'detail',
        name: 'AssetDetail',
        component: AssetDetail,
        meta: { title: '资产明细', icon: 'ListOutline' }
      },
      {
        path: 'debt-plans',
        name: 'AssetDebtPlans',
        component: AssetDebtPlans,
        meta: { title: '负债计划', icon: 'BarChartOutline' }
      },
      {
        path: 'digital-analysis',
        alias: 'ai-analysis',
        name: 'AssetDigitalAnalysis',
        component: DigitalAnalysis,
        meta: { title: '数字分析', icon: 'ChatbubblesOutline' }
      },
      {
        path: 'benchmarks',
        name: 'AssetBenchmarks',
        component: AssetBenchmarks,
        meta: { title: '基准数据', icon: 'StatsChartOutline' }
      },
      {
        path: 'members',
        name: 'AssetMembers',
        component: AssetMembers,
        meta: { title: '家庭成员', icon: 'PeopleOutline' }
      },
      {
        path: 'unlock',
        name: 'AssetUnlock',
        component: AssetUnlock,
        meta: { title: '资产解锁', icon: 'LockClosedOutline', hideInMenu: true }
      }
    ]
  },
  {
    path: '/invest',
    name: 'Invest',
    redirect: '/invest/monitor',
    meta: { title: '投资分析', icon: 'TrendingUpOutline' },
    children: [
      {
        path: 'monitor',
        name: 'StockMonitor',
        component: StockMonitor,
        meta: { title: '股票监控', icon: 'EyeOutline' }
      },
      {
        path: 'market',
        name: 'MarketOverview',
        component: MarketOverview,
        meta: { title: '市场行情', icon: 'BarChartOutline' }
      },
      {
        path: 'fund',
        name: 'FundWatch',
        component: FundWatch,
        meta: { title: '基金自选', icon: 'CashOutline' }
      },
      {
        path: 'research',
        alias: 'ai-analysis',
        name: 'ResearchCenter',
        component: ResearchCenter,
        meta: { title: '研究中心', icon: 'AnalyticsOutline' }
      },
      {
        path: 'agent',
        name: 'AgentChat',
        component: AgentChat,
        meta: { title: 'AI 智能体', icon: 'ChatbubblesOutline' }
      },
      {
        path: 'all-stocks',
        name: 'AllStocks',
        component: AllStocks,
        meta: { title: '全市场股票', icon: 'ListOutline' }
      },
      {
        path: 'stock-library',
        name: 'StockLibrary',
        component: StockLibrary,
        meta: { title: '股票资料库', icon: 'StorefrontOutline' }
      },
      {
        path: 'ai-recommendations',
        name: 'AIRecommendations',
        component: AIRecommendations,
        meta: { title: 'AI 荐股记录', icon: 'FlashOutline' }
      },
      {
        path: 'research-feeds',
        name: 'ResearchFeeds',
        component: ResearchFeeds,
        meta: { title: '公告与研报', icon: 'AnalyticsOutline' }
      },
      {
        path: 'hot-discovery',
        name: 'HotDiscovery',
        component: HotDiscovery,
        meta: { title: '热点发现', icon: 'FlashOutline' }
      },
      {
        path: 'calendars',
        name: 'Calendars',
        component: Calendars,
        meta: { title: '投资日历', icon: 'TimeOutline' }
      },
      {
        path: 'long-tiger',
        name: 'LongTiger',
        component: LongTiger,
        meta: { title: '龙虎榜', icon: 'BarChartOutline' }
      },
      {
        path: 'money-rank',
        name: 'MoneyRank',
        component: MoneyRank,
        meta: { title: '资金排行', icon: 'TrendingUpOutline' }
      }
    ]
  },
  {
    path: '/portfolio',
    name: 'Portfolio',
    redirect: '/portfolio/overview',
    meta: { title: '持仓分析', icon: 'BriefcaseOutline' },
    children: [
      {
        path: 'overview',
        name: 'PortfolioOverview',
        component: PortfolioOverview,
        meta: { title: '持仓总览', icon: 'WalletOutline' }
      },
      {
        path: 'stocks',
        name: 'StockHoldings',
        component: StockHoldings,
        meta: { title: '股票持仓', icon: 'StorefrontOutline' }
      },
      {
        path: 'funds',
        name: 'FundHoldings',
        component: FundHoldings,
        meta: { title: '基金持仓', icon: 'StorefrontOutline' }
      },
      {
        path: 'history',
        name: 'ProfitHistory',
        component: ProfitHistory,
        meta: { title: '收益历史', icon: 'TimeOutline' }
      },
      {
        path: 'transactions',
        name: 'Transactions',
        component: Transactions,
        meta: { title: '交易记录', icon: 'CreateOutline' }
      }
    ]
  },
  {
    path: '/quant',
    name: 'Quant',
    redirect: '/quant/templates',
    meta: { title: '量化模板库', icon: 'CodeSlashOutline' },
    children: [
      {
        path: 'templates',
        name: 'TemplateList',
        component: TemplateList,
        meta: { title: '模板库', icon: 'ListOutline' }
      },
      {
        path: 'editor/:id?',
        name: 'TemplateEditor',
        component: TemplateEditor,
        meta: { title: '脚本编辑', icon: 'CreateOutline', hideInMenu: true }
      },
      {
        path: 'ai-generate',
        name: 'AIGenerator',
        component: AIGenerator,
        meta: { title: 'AI 生成', icon: 'FlashOutline' }
      },
      {
        path: 'search',
        name: 'ScriptSearch',
        component: ScriptSearch,
        meta: { title: '脚本搜索', icon: 'AnalyticsOutline' }
      },
      {
        path: 'linkage',
        name: 'StrategyLinkage',
        component: StrategyLinkage,
        meta: { title: '联动推荐', icon: 'LinkOutline' }
      }
    ]
  },
  {
    path: '/prompts',
    name: 'PromptTemplates',
    component: PromptTemplates,
    meta: { title: '提示词模板', icon: 'CreateOutline' }
  },
  {
    path: '/settings',
    name: 'Settings',
    component: Settings,
    meta: { title: '设置', icon: 'SettingsOutline' }
  },
  {
    path: '/about',
    name: 'About',
    component: About,
    meta: { title: '关于', icon: 'InformationCircleOutline' }
  }
]
