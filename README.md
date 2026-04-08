# Rubin Investment

`Rubin Investment` 是一个基于 `Wails + Vue + Go` 的本地优先桌面投资工具。项目复用了上游 [`go-stock`](https://github.com/ArvinLovegood/go-stock) 的部分行情、资讯和研究能力，并在此基础上扩展了资产分析、持仓分析、基金工作流、量化模板和 AI 辅助能力。

[English Summary](#english-summary)

## 当前定位

这个仓库面向公开代码托管整理，默认遵循下面的边界：

- 私有数据库、私有覆盖配置、Cookie、Token 不进入 Git
- 本地运行数据默认落在仓库根目录 `data/`、`logs/`，并由 `.gitignore` 排除
- 私有运行配置通过 `data/private-overrides.json` 或环境变量注入
- 仓库保留上游 `go-stock` 的许可证与归属说明

## 主要能力

### 1. 资产分析

- 资产总览、资产明细、负债计划、家庭成员、基准数据
- 家庭数字分析与资产解锁入口

### 2. 投资分析

- 股票监控、市场行情、研究中心、AI 智能体
- 基金自选：支持分组、AI 分析、基金对比、基金详情抽屉
- 基金筛选：支持按类型、阶段收益、回撤和行业筛选，并直接打开基金详情
- 公告与研报、热点发现、投资日历、龙虎榜、资金排行
- 全市场股票、股票资料库、AI 荐股记录

### 3. 持仓分析

- 持仓总览
- 股票持仓
- 基金持仓：支持按金额录入、查看详情、AI 解读
- 收益历史
- 交易记录

### 4. 量化模板

- 模板库
- AI 生成
- 脚本搜索
- 联动推荐
- 模板编辑、导出、启用

## 技术栈

- 桌面壳：`Wails`
- 前端：`Vue 3`、`Naive UI`、`Vite`
- 后端：`Go`
- 本地数据：`SQLite`

## 仓库结构

- `main.go`、`app.go`、`internal/`：桌面宿主与本仓库新增业务模块
- `frontend/`：当前桌面前端
- `go-stock/`：复用与适配的上游市场分析代码
- `docs/`：配置、审计、分支与运行记录
- `scripts/`：开发、打包、导出辅助脚本

## 快速开始

### 环境要求

- Go `1.26`
- Node.js
- Wails CLI `v2`

安装 Wails CLI：

```powershell
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### 安装依赖

```powershell
cd frontend
npm ci
cd ..
```

### 本地校验

```powershell
go test ./...
cd frontend
npm run build
cd ..
```

### 启动开发环境

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\start-wails-dev.ps1
```

### 打包导出

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\build-and-export.ps1
```

如果需要指定导出目录：

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\build-and-export.ps1 -OutputDir D:\apps\rubin-investment -SkipShortcut
```

## 运行时配置

本仓库支持把本地敏感配置放在 Git 之外。

运行时会从下面位置读取私有覆盖配置：

1. 环境变量 `INVESTMENT_PRIVATE_CONFIG_PATH`
2. 本地文件 `data/private-overrides.json`

常见用途：

- 指定本地种子数据库路径 `seedDbPaths`
- 注入本地测试用 Cookie / Token
- 覆盖资讯、下载、消息墙等运行时接口地址

更多说明见：

- [docs/runtime-configuration.md](docs/runtime-configuration.md)
- [docs/examples/private-overrides.example.json](docs/examples/private-overrides.example.json)

## 分支与发布

当前阶段建议使用 beta 版本发布：

- `main`：对外公开、可发布代码
- `develop`：日常集成分支
- `feature/*`：功能分支
- `fix/*`：非紧急修复
- `hotfix/*`：已发布 beta 的紧急修复

Tag 约定：

- `v0.x.y-beta.N`

更多说明见：

- [docs/branching-and-beta-plan.md](docs/branching-and-beta-plan.md)

## 开源边界

请不要提交以下内容：

- `data/stock.db`
- `data/private-overrides.json`
- `go-stock/data/stock.db`
- 浏览器 Cookie、账号 Token、API Key
- 本机路径、日志、截图、打包产物

如果你发现敏感信息泄露或潜在漏洞，请按 [SECURITY.md](SECURITY.md) 处理。

## 贡献

贡献说明见 [CONTRIBUTING.md](CONTRIBUTING.md)。

## 上游归属

本仓库包含基于 `go-stock` 的衍生与适配代码：

- 上游仓库：<https://github.com/ArvinLovegood/go-stock>
- 上游协议：Apache License 2.0

更多归属信息见 [NOTICE](NOTICE)。

## English Summary

`Rubin Investment` is a local-first desktop investment app built with `Wails + Vue + Go`.

Highlights:

- Asset analysis workflows for households and liabilities
- Investment research pages built on top of `go-stock`
- Portfolio overview, stock holdings, fund holdings, and transaction tracking
- Fund watch, fund screener, fund comparison, and fund detail workflows
- Quant template management, AI generation, and strategy linkage
- Private runtime overrides kept out of Git

Quick start:

```powershell
cd frontend
npm ci
cd ..
go test ./...
cd frontend
npm run build
cd ..
powershell -ExecutionPolicy Bypass -File .\scripts\start-wails-dev.ps1
```

Build:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\build-and-export.ps1
```

Tagged GitHub releases publish:

- `investment-platform-<tag>-windows-amd64.exe`
- `investment-platform-<tag>-windows-amd64.zip`
- `investment-platform-<tag>-macos-universal.zip`
- `investment-platform-<tag>-macos-universal.tar.gz`

For runtime override details, see:

- [docs/runtime-configuration.md](docs/runtime-configuration.md)
- [docs/examples/private-overrides.example.json](docs/examples/private-overrides.example.json)
