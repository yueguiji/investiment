# go-stock Upstream Gap Report - 2026-07-06

Baseline:

- Local product: current repository working tree, branch `main`, HEAD `e8b1b01`.
- Local upstream subtree: `go-stock/` is not a git repository.
- Upstream source: `https://gitcode.com/gh_mirrors/go/go-stock.git`, branch `dev`.
- Upstream latest checked: `17fef8d feat(notification):升级飞书卡片协议至2.0并更新@所有人语法`.
- Upstream release tag at HEAD: `v2026.07.06.2-release`.
- GitHub direct access timed out on this machine, so the GitCode mirror was used.

## Size Of The Gap

Source/config files compared after excluding `.git`, `node_modules`, `dist`, logs, binaries, images, local databases, and large font/doc artifacts:

| Metric | Count |
| --- | ---: |
| Upstream source/config files | 289 |
| Local `go-stock/` source/config files | 155 |
| Common files | 145 |
| Identical common files | 71 |
| Changed common files | 74 |
| Upstream-only files | 144 |
| Local-only files | 10 |

Main diff scale:

| Area | Scale |
| --- | ---: |
| `backend/data` | 98 files changed, about 27,300 insertions and 5,718 deletions |
| `backend/agent` | 30 files changed, about 12,462 insertions and 261 deletions |
| `frontend/src` | 46 files changed, about 27,515 insertions and 1,133 deletions |
| `go-stock/app.go` | about 2,093 insertions and 791 deletions |

App binding gap:

| Metric | Count |
| --- | ---: |
| Upstream App methods | 212 |
| Local App methods | 98 |
| Upstream-only App methods | 115 |
| Local-only App methods | 1 |

## What Is Missing By Function

### 1. AI Agent And Tool System

Missing or heavily changed:

- Agent tracing, memory, response guard, token utilities, model factory.
- Tool grouping and question-based tool selection.
- MCP server/tool integration.
- Skill management API.
- Cron task API for scheduled agent tasks.
- New model providers and SDK upgrades: Claude, Gemini, Ollama, OpenRouter, Qwen, MCP tool support.

Representative files:

- `backend/agent/chat_model_factory.go`
- `backend/agent/cron_task_api.go`
- `backend/agent/tools/data_tools_wrapper.go`
- `backend/agent/tools/mcp_skill_tools.go`
- `backend/data/mcp_server_api.go`
- `backend/data/skill_api.go`
- `frontend/src/components/cron-task-manager.vue`
- `frontend/src/components/mcp-server-manager.vue`
- `frontend/src/components/skill-manager.vue`

### 2. Market Data And K-Line

Missing or heavily changed:

- Eastmoney K-line and page APIs.
- TDX market data APIs: finance info, company info, XDXR, board relation, call auction.
- Sina K-line fallback.
- Wallstreetcn lives/market/calendar/kline.
- Lightweight K-line frontend with drawing/indicator helpers.
- Chip distribution.
- Market statistic APIs.

Representative files:

- `backend/data/eastmoney_api.go`
- `backend/data/eastmoney_kline_api.go`
- `backend/data/tdx_kline_api.go`
- `backend/data/sina_kline_api.go`
- `backend/data/wallstreetcn_api.go`
- `backend/data/chip_distribution.go`
- `backend/data/market_statistic_api.go`
- `frontend/src/components/StockLightweightKlineChart.vue`
- `frontend/src/components/kline/*`

### 3. Stock Monitoring And Research

Missing or heavily changed:

- Stock changes monitoring and history stats.
- Up-limit ladder and hot tables.
- Research/announcement/F10 helpers.
- Iwencai markdown/query flow.
- Custom strategy APIs.

Representative files:

- `backend/data/stock_changes_api.go`
- `backend/data/stock_change_history_api.go`
- `backend/data/f10_data_api.go`
- `backend/data/iwencai_api.go`
- `backend/data/iwencai_markdown.go`
- `frontend/src/components/stockChangesMonitor.vue`
- `frontend/src/components/uplimitLadder.vue`

### 4. Fund Feature Surface

Missing or diverged:

- Upstream has fund follow/ranking/K-line UI.
- `fund_data_api.go` has a large rewrite and conflicts with local fund portfolio work.
- Local has `fund_holdings_api.go` and fund enhancement tests that upstream does not have.

Representative files:

- `backend/data/fund_data_api.go`
- `backend/data/fund_kline_api.go`
- `frontend/src/components/FundFollow.vue`
- `frontend/src/components/FundKlineChart.vue`
- `frontend/src/components/FundRanking.vue`

### 5. Notification, Sharing, Trading Records

Missing or heavily changed:

- Feishu notification API and updated card protocol.
- DingTalk API changes.
- Trading record manager.
- Alert reset and frequent trading checks.
- Share text/user manual helpers.

Representative files:

- `backend/data/feishu_api.go`
- `backend/data/dingding_api.go`
- `frontend/src/components/TradingRecordManager.vue`
- App methods: `SendFeishuMessage`, `SendFeishuMessageByType`, `AddTradingRecord`, `CheckFrequentTrading`.

## Merge Decision Matrix

### Can Be Directly Added First

These are mostly new isolated files. They still need imports, Wails bindings, and tests, but they do not require replacing Rubin's main application shell.

- Data clients/utilities: `httpclient.go`, `eastmoney_api.go`, `eastmoney_kline_api.go`, `sina_kline_api.go`, `tdx_kline_api.go`, `wallstreetcn_api.go`, `web_search_api.go`, `xueqiu_chromedp.go`.
- Market/fund-flow data: `bk_fund_flow_api.go`, `concept_fund_flow_api.go`, `chip_distribution.go`, `market_statistic_api.go`.
- Frontend isolated components: `StockLightweightKlineChart.vue`, `frontend/src/components/kline/*`, `FundRanking.vue`, `FundKlineChart.vue`, `bkFundFlowChart.vue`, `conceptFundFlowChart.vue`.
- Tests for the above should be copied with the feature where they do not require browser/private credentials.

### Needs Adaptation, Not Direct Replace

These touch local Rubin architecture or changed schemas/bindings.

- `go-stock/app.go` and `app_common.go`: upstream has 115 App methods not present locally, but Rubin uses its own root `app.go`, `gostock_compat.go`, and portfolio/asset modules. Merge method-by-method.
- `backend/models/models.go`: upstream model changes must be checked against `internal/shared/db.go` AutoMigrate and existing SQLite data.
- `backend/data/fund_data_api.go`: large conflict with Rubin fund holding/portfolio workflow.
- `backend/data/settings_api.go` and `openai_api.go`: upstream has major AI provider/config refactor; merge only after preserving local runtime override behavior.
- `frontend/src/App.vue`, router, settings, stock/market pages: Rubin has different sidebar/layout and business pages. Port features into existing views instead of wholesale replacing.
- `frontend/wailsjs/*`: regenerate from Wails after backend binding changes, do not manually merge.
- `go.mod/go.sum`: dependency upgrade is required for agent/MCP/model features, but should be a separate phase because it pulls many SDK upgrades.

### Not Recommended To Merge As-Is

- Upstream `main.go`, platform app files, tray/update/sponsor/device-binding code: conflicts with Rubin product identity, runtime paths, database location, and release behavior.
- `ai-assistant-web/`: separate web app; only useful if Rubin wants a standalone web assistant.
- `.github`, release scripts, sponsor docs/screenshots: not core product functionality.
- Upstream build assets and docs can be cherry-picked only if needed.

## Recommended Update Order

1. Data-source layer first:
   - Eastmoney/Sina/TDX/Wallstreetcn/WebSearch/Xueqiu clients.
   - Verify with live tests and no UI changes.

2. K-line and market pages:
   - Add K-line fallback/page APIs and lightweight K-line component.
   - Wire into existing Rubin stock/fund detail surfaces.

3. Stock monitor and market statistics:
   - Stock changes history, daily stats, up-limit ladder, fund flow charts.
   - Add backend tests and one manual UI pass.

4. Fund workflow:
   - Reconcile upstream `fund_data_api.go` with local fund holding, fund estimate, AI correction, and recommendation code.
   - Do not overwrite local fund APIs.

5. AI Agent/MCP/Skills:
   - Upgrade `go.mod` in a controlled branch.
   - Port agent memory/model factory/tool groups.
   - Add UI managers only after backend compiles.

6. Final Wails binding regeneration:
   - Regenerate `frontend/wailsjs/*`.
   - Run `go test ./...`, frontend build, Wails build.

## Quality Gate

For every phase:

- No database overwrite. Schema changes must go through GORM `AutoMigrate` or explicit additive migrations only.
- Keep Rubin root app shell intact.
- Run `go test ./...`.
- Run `npm --prefix frontend run build`.
- Run `wails build`.
- For live market data features, run focused live tests behind env flags.
- For UI features, verify at least the affected page manually or with browser automation.
