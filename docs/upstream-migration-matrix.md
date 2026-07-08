# go-stock 上游迁移矩阵

更新日期：2026-07-06

## 基线

- 当前产品：Rubin Investment，根仓库 `main`，当前提交 `e8b1b01`
- 上游项目：`ArvinLovegood/go-stock`
- 上游基线：`v2026.07.06.2-release`，提交 `17fef8d`
- 对比范围：本仓库 `go-stock/` 与上游 `go-stock` 源码树

## 当前差异概览

| 分类 | 结果 |
|---|---:|
| 上游源码/文本文件数 | 312 |
| 本地内嵌 go-stock 文件数 | 169 |
| 双方共有文件 | 163 |
| 共有但内容不同 | 76 |
| 上游独有文件 | 149 |
| 本地独有文件 | 6 |

本地不是简单落后上游，而是对 `go-stock` 做了裁剪、路径适配和产品化嵌入。更新必须按功能迁移，不能直接覆盖 `go-stock/`。

## 功能分型

| 分型 | 代表范围 | 迁移策略 | 风险 |
|---|---|---|---|
| Rubin 独有业务 | `internal/asset`、`internal/portfolio`、`internal/quant`、`frontend/src/views` | 保留为主产品边界，不接受上游覆盖 | 高 |
| 上游基础数据源 | 行情、新闻、K 线、基金净值、研究报告接口 | 按接口逐个移植，先后端后前端 | 中 |
| 上游 AI/Agent | Agent Mode、MCP、AI Assistant Web、工具调用、聊天记忆 | 独立隔离迁移，先不开主入口 | 高 |
| 上游图表能力 | 多周期 K 线、SATS 指标、TV 指标、轻量 K 线组件 | 先迁移数据模型，再接 UI | 高 |
| 上游基金增强 | 基金 K 线、基金前十大重仓、ETF 支持 | 适配到现有基金持仓/详情，不替换页面 | 中高 |
| 运行时适配 | `runtimeconfig`、`runtimepaths`、安装目录数据库保护 | 本地优先，禁止被上游覆盖 | 高 |
| 构建发布 | Wails 配置、release workflow、安装包脚本 | 只吸收必要修复 | 中 |

## 迁移批次

### P0：保护和审计

目标：确保后续迁移不覆盖本地数据和产品能力。

- 固定上游基线到 `v2026.07.06.2-release`
- 保留当前未提交的 AI 估值纠偏改动
- 跑 `scripts/compare-go-stock-upstream.ps1` 生成差异统计
- 禁止直接替换 `go-stock/` 整目录

验收：

```powershell
git status --short --branch
powershell -ExecutionPolicy Bypass -File .\scripts\compare-go-stock-upstream.ps1
```

### P1：低风险后端修复

目标：优先吸收不改变产品结构的上游修复。

候选：

- 通知协议兼容修复
- 数据源 URL 或解析小修
- Go 依赖安全/兼容更新
- 非 UI 的测试修复

排除：

- 数据库路径、运行时目录、安装数据路径
- 设置页整体覆盖
- AI Agent 大规模重构

验收：

```powershell
go test ./...
cd frontend
npm run build
cd ..
```

### P2：基金数据增强

目标：把上游基金 K 线、前十大持仓等能力接入 Rubin 的基金详情。

迁移方式：

1. 只迁移后端数据抓取和模型。
2. 新增服务方法，不替换现有 `FundHoldings.vue` 工作流。
3. 在 `FundInsightDrawer.vue` 增加独立 tab 或信息块。
4. 用真实持仓基金验证数据可用性。

验收：

```powershell
go test ./...
npm --prefix frontend run build
wails build
```

人工检查：

- 基金持仓列表可打开
- 基金详情抽屉不报错
- 今日估值来源选择仍可用
- AI 估值纠偏数据不被覆盖

### P3：K 线和指标体系

目标：逐步引入上游多周期 K 线和指标，而不破坏当前股票监控页。

迁移方式：

1. 后端 K 线数据接口先兼容。
2. 前端组件作为新组件接入，不替换老图表。
3. 增加开关或独立入口。

验收：

- 股票监控页可打开
- K 线弹窗可渲染
- 图表无空白、无控制台错误

### P4：Agent/MCP 能力

目标：上游 AI Agent 能力与本地 AI 分析系统融合。

迁移方式：

1. 保留本地 PromptTemplate 和 AIConfig。
2. 后端工具注册隔离为独立包。
3. 默认关闭新 Agent 入口。
4. 先在开发环境验证工具调用日志、超时和错误边界。

验收：

- 原股票 AI 分析可用
- 原基金/资产 AI 分析可用
- 新 Agent 失败不会影响主流程

## 禁止事项

- 禁止覆盖 `data/stock.db`
- 禁止覆盖 `data/private-overrides.json`
- 禁止整目录复制上游 `backend`、`frontend`
- 禁止把上游默认 runtime path 覆盖本地安装目录路径
- 禁止在未跑真实库验证前发布基金估值相关改动

## 每批迁移交付标准

每个迁移批次必须包含：

- 迁移的上游提交或 tag
- 文件清单
- 风险说明
- 自动化验证结果
- 需要人工点击验证的页面
- 回滚方式

## 执行记录

### 2026-07-06 P0/P1

基线：

- 上游：`v2026.07.06.2-release` / `17fef8d`
- 本地：Rubin Investment `e8b1b01`，保留当前未提交 AI 估值纠偏改动

已完成：

- 新增 `scripts/compare-go-stock-upstream.ps1`，用于重复生成上游差异统计
- 补充 `StockGroupApi.GetAllGroupStocks`
- 补充 `StockGroupApi.UpdateGroup`
- 补充根应用兼容层 `GetAllGroupStocks` / `UpdateGroup`
- 补充内嵌 `go-stock` App 绑定
- 更新 Wails generated 调用声明

本批文件：

- `go-stock/backend/data/stock_group_api.go`
- `go-stock/app.go`
- `gostock_compat.go`
- `frontend/wailsjs/go/main/App.d.ts`
- `frontend/wailsjs/go/main/App.js`
- `go-stock/frontend/wailsjs/go/main/App.d.ts`
- `go-stock/frontend/wailsjs/go/main/App.js`
- `scripts/compare-go-stock-upstream.ps1`
- `docs/upstream-migration-matrix.md`

验证：

```powershell
go test ./...
npm --prefix frontend run build
wails build
```

结果：

- 根工程 `go test ./...` 通过
- 前端生产构建通过
- 根应用 `wails build` 通过，生成 `build/bin/investment-platform.exe`
- 单独执行 `go-stock` 子模块测试未作为通过门槛：该目录已被本产品裁剪，独立上游应用测试会因缺少截图资源和独立 DB 初始化失败而失败

本批风险：

- 无数据库 schema 变更
- 无基金估值逻辑变更
- 无运行路径变更
- 无安装目录数据覆盖
### 2026-07-06 P2 基金数据增强

基线：
- 上游：`v2026.07.06.2-release` / `17fef8d`
- 本地：Rubin Investment，延续 P0/P1 的保护策略，不覆盖本地数据库和 AI 估值纠偏改动。

已完成：
- 新增基金净值 K 线只读 API：`GetFundKLine(fundCode, klt, limit)`
- 新增基金前十大重仓股只读 API：`GetFundTop10Holdings(fundCode)`
- 前十大持仓支持东方财富 F10 抓取，并轻量补全 A/H/US 股票价格与涨跌幅
- 基金 K 线先接入场外基金净值 K 线，避免引入上游完整场内行情 K 线依赖
- 基金详情抽屉新增“重仓股”页签，展示排名、股票、市场、占净值、最新价、涨跌幅
- `GetSettingConfig()` 增加 DB 未初始化时的默认配置兜底，便于数据 API 独立测试
- 新增解析测试与 live 数据源验证测试，live 测试需显式设置 `RUN_FUND_DATA_LIVE=1`

本批文件：
- `app.go`
- `go-stock/app.go`
- `go-stock/backend/data/fund_kline_api.go`
- `go-stock/backend/data/fund_holdings_api.go`
- `go-stock/backend/data/fund_enhancement_test.go`
- `go-stock/backend/data/settings_api.go`
- `frontend/src/views/portfolio/components/FundInsightDrawer.vue`
- `frontend/wailsjs/go/main/App.d.ts`
- `frontend/wailsjs/go/main/App.js`
- `frontend/wailsjs/go/models.ts`
- `go-stock/frontend/wailsjs/go/main/App.d.ts`
- `go-stock/frontend/wailsjs/go/main/App.js`
- `go-stock/frontend/wailsjs/go/models.ts`

验证：
```powershell
go test ./...
go test go-stock/backend/data -run "TestParseFundTrendKLines|TestParseFundHoldingRows" -count=1
$env:RUN_FUND_DATA_LIVE='1'; go test go-stock/backend/data -run TestFundEnhancementLive -count=1 -v
npm --prefix frontend run build
wails build
```

结果：
- 根工程 `go test ./...` 通过
- 基金解析单测通过
- 真实东方财富 live 验证通过：`161725` 可返回净值 K 线和前十大持仓
- 前端生产构建通过
- Wails 打包通过，生成 `build/bin/investment-platform.exe`

本批风险边界：
- 未迁移上游完整场内基金 K 线、Sina/Tencent/EastMoney 股票 K 线 fallback；该能力放入 P3
- 未改变基金估值纠偏算法，也未写入历史估值快照
- 未修改数据库 schema
- 未覆盖安装目录数据、`stock.db` 或本地私有配置
### 2026-07-06 P3 股票 K 线兼容入口

基线：
- 上游：`v2026.07.06.2-release` / `17fef8d`
- 本地：Rubin Investment，仍不直接替换上游完整图表系统。

已完成：
- 新增 `StockDataApi.GetStockKLineWithFallback(stockCode, kLineType, days)`
- `GetStockKLine` 改为使用兼容回退入口：A 股优先腾讯 K 线，港/美走原港美路径，失败后回退新浪 K 线
- 根应用补齐 `GetStockCommonKLine`，与内嵌 `go-stock` App 绑定保持一致
- 新增 K 线代码归一化测试和 live 股票 K 线验证测试

本批文件：
- `go-stock/backend/data/stock_kline_fallback_api.go`
- `go-stock/backend/data/stock_kline_fallback_api_test.go`
- `gostock_compat.go`
- `go-stock/app.go`
- Wails 生成的 `frontend/wailsjs/go/main/App.*`

验证：
```powershell
go test go-stock/backend/data -run TestNormalizeStockKLineCode -count=1
$env:RUN_STOCK_KLINE_LIVE='1'; go test go-stock/backend/data -run TestStockKLineFallbackLive -count=1 -v
go test ./...
npm --prefix frontend run build
wails build
```

结果：
- 股票 K 线 live 验证通过：`000001.SZ` 返回 2026-07-06 最新日 K 数据
- 根工程 `go test ./...` 通过
- 前端生产构建通过
- Wails 打包通过，生成 `build/bin/investment-platform.exe`

本批风险边界：
- 未迁移上游 `EastMoneyKLineApi` 完整实现
- 未迁移 `lightweight-charts` 绘图组件、画线工具、SATS/TV 指标系统
- 未改现有股票监控页布局，只增强原 `GetStockKLine` 数据入口稳定性
