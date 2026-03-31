# Current Desktop vs Open Source Audit

Updated: 2026-03-16

## Summary

The desktop app now covers the main `go-stock` workflows and most bundled open-source pages. Core gaps are no longer about missing screens; they are mostly about how deep those screens go.

Status buckets used below:

- `Integrated`: routed and usable in the desktop app
- `Partial`: usable, but simplified or still missing a richer workflow
- `Available / Internal`: helper component or backend support exists, but it is not meant to be a first-class page on its own

## 1. Investment Analysis

| Capability | Open-source source | Desktop status | Notes |
|---|---|---|---|
| Stock monitor | `frontend/src/gostock/components/stock.vue` | Integrated | Main monitoring workflow preserved |
| Market overview | `frontend/src/gostock/components/market.vue` | Integrated | Core market analysis retained |
| Fund watch | `frontend/src/gostock/components/fund.vue` | Integrated | Main fund watch flow retained |
| Research center | `frontend/src/gostock/components/researchIndex.vue` | Integrated | Main research flow retained |
| Agent chat | `frontend/src/gostock/components/agent-chat.vue` | Integrated | Streaming flow repaired |
| Full market stock list | `frontend/src/gostock/components/allStockList.vue` | Integrated | Routed through dedicated wrapper page |
| All stock info list | `frontend/src/gostock/components/allStockInfoList.vue` | Integrated | Routed as stock library |
| AI recommendation list | `frontend/src/gostock/components/aiRecommendStocksList.vue` | Integrated | Routed as AI recommendation history |
| Prompt template list | `frontend/src/gostock/components/promptTemplateList.vue` | Integrated | Routed as prompt template page |
| Stock notice list | `frontend/src/gostock/components/StockNoticeList.vue` | Integrated | Surfaced in research feeds page |
| Stock research report list | `frontend/src/gostock/components/StockResearchReportList.vue` | Integrated | Surfaced in research feeds page |
| Industry research report list | `frontend/src/gostock/components/IndustryResearchReportList.vue` | Integrated | Surfaced in research feeds page |
| Hot events / topics / hot stocks | `HotEvents.vue`, `HotTopics.vue`, `HotStockList.vue` | Integrated | Surfaced in hot discovery page |
| Calendars / long-tiger / rank views | `ClsCalendarTimeLine.vue`, `InvestCalendarTimeLine.vue`, `LongTigerRankList.vue`, `rankTable.vue` | Integrated | Surfaced through dedicated calendar, 龙虎榜, and money rank pages |
| Helper market widgets | `KLineChart.vue`, `moneyTrend.vue`, `stockSparkLine.vue`, `industryMoneyRank.vue`, `AnalyzeMartket.vue` | Available / Internal | Used as embedded widgets or lower-level building blocks |

## 2. Asset Analysis

| Capability | Source | Desktop status | Notes |
|---|---|---|---|
| Asset overview | custom desktop view | Integrated | CRUD + summary available |
| Asset detail | custom desktop view | Integrated | Grouped detail and totals page added |
| Asset AI analysis | custom desktop view | Partial | Heuristic analysis and allocation suggestions available |
| Asset categories | backend available | Partial | No dedicated category management page yet |

## 3. Portfolio Analysis

| Capability | Source | Desktop status | Notes |
|---|---|---|---|
| Portfolio overview | custom desktop view | Integrated | Summary + add/delete available |
| Stock holdings detail | custom desktop view | Integrated | Dedicated filtered holdings page added |
| Fund holdings detail | custom desktop view | Integrated | Dedicated filtered holdings page added |
| Profit history | custom desktop view | Integrated | Snapshot trend and detail table added |
| Transactions | backend available via `GetTransactions` / `AddTransaction` | Integrated | Dedicated transaction page added |

## 4. Quant Templates

| Capability | Source | Desktop status | Notes |
|---|---|---|---|
| Template list | custom desktop view | Integrated | Loads list from backend |
| AI generator | custom desktop view | Partial | Prompt-based generation works, not a richer code agent workflow |
| Template editor | custom desktop view | Integrated | CRUD, activate, export, delete exposed |
| Strategy linkage | custom desktop view | Partial | Matching and suggestion flow added, but still heuristic |
| Export template | backend available | Integrated | Exposed in template editor |
| Activate template | backend available | Integrated | Exposed in template editor |

## 5. Settings / About

| Capability | Source | Desktop status | Notes |
|---|---|---|---|
| go-stock settings | `frontend/src/gostock/components/settings.vue` | Integrated | Core settings retained |
| About | `frontend/src/gostock/components/about.vue` | Integrated | Retained |

## Remaining Gaps

- Add a dedicated asset category management page if category-level maintenance becomes important
- Upgrade `AssetAIAnalysis` from heuristic summaries to true model-driven analysis if required
- Upgrade `StrategyLinkage` from rule-based matching to a fuller execution workflow if required
- Do a full click-through regression pass on every newly surfaced desktop page
