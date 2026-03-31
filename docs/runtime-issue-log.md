# Runtime Issue Log

Updated: 2026-03-16

## Confirmed Issues

### 1. Sidebar navigation crashed on parameterized routes

- Symptom: clicking some sidebar items raised a Vue Router stack similar to `resolve -> push -> onClick`
- Root cause: the sidebar menu generated direct links from child route paths, including parameterized paths like `/quant/editor/:id?`
- Fix: parameterized child routes are now hidden from the sidebar menu, and menu navigation failures are logged instead of surfacing as an uncaught crash
- Files: `frontend/src/App.vue`, `frontend/src/router/index.js`

### 2. ECharts `setOption` crashed on null or disposed instances

- Symptom: runtime error similar to `TypeError: Cannot read properties of null (reading 'setOption')`
- Root cause: chart components updated ECharts instances before DOM mount completed or after component unmount, especially in async callbacks and reactive watchers
- Fix: added container checks, instance reuse, unmount disposal, and async result guards in chart components
- Files: `frontend/src/gostock/components/KLineChart.vue`, `frontend/src/gostock/components/moneyTrend.vue`, `frontend/src/gostock/components/stockSparkLine.vue`, `frontend/src/gostock/components/AnalyzeMartket.vue`

### 3. Hot discovery opened on an empty first page

- Symptom: the first tab in `热点发现` looked blank when the hot-events feed had no returned items
- Root cause: the page defaulted to the less stable `HotEvent` feed and had no explicit empty state
- Fix: switched the default tab to `热话题` and added an empty-state message for `热事件`
- Files: `frontend/src/views/invest/HotDiscovery.vue`, `frontend/src/gostock/components/HotEvents.vue`

### 4. Stock monitor notice and report actions still pointed at an upstream-only route

- Symptom: clicking `公告` or `研报` inside stock monitor raised a Vue Router resolve/push error
- Root cause: `frontend/src/gostock/components/stock.vue` still pushed to the old upstream route name `market`, but the host desktop app uses a different routed page layout
- Fix: rewired both actions to `/invest/research-feeds` and made that page accept `tab` and `stockCode` query parameters
- Files: `frontend/src/gostock/components/stock.vue`, `frontend/src/views/invest/ResearchFeeds.vue`

## Pending Verification

- Hot discovery `events` may still require a packaged-app click-through after adding a fallback from Xueqiu hot events to local telegraph/news data
- Re-run full menu click-through against the packaged app and append any remaining page-specific runtime errors
- Verify newly exposed pages that depend on multi-value Wails bindings, especially transactions and strategy linkage
