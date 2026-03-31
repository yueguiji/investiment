# Rubin Investment

`Rubin Investment` 是一个基于 `Wails + Vue` 的本地优先投资研究桌面应用。项目复用了 `go-stock` 的市场数据与分析能力，并在此基础上增加了资产管理、组合跟踪、量化模板工作流和 AI 辅助分析能力。

[English Summary](#english-summary)

本仓库已经按公开发布方式整理：

- 不提交私有数据库、私有覆盖配置和个人凭据
- 本地运行数据默认位于 `data/`、`logs/`
- 私有配置通过仓库外部或本地忽略文件注入
- 仓库保留了 `go-stock` 的上游归属与许可证说明

## 当前状态

- 持仓模块仍在开发中，目前还不是完整交付状态
- 量化联动推荐功能已接入，但当前稳定性一般，结果质量和联动表现仍需继续优化

## 项目范围

- 基于 `Wails` 的桌面壳应用
- 复用 `go-stock` 的行情、资讯和部分分析页面
- 新增家庭资产管理、投资组合跟踪、量化模板能力
- 使用本地 `SQLite` 数据库，支持初始化种子数据

## 仓库结构

- `main.go`、`app.go`、`internal/`：桌面应用宿主与新增业务模块
- `frontend/`：当前桌面应用前端
- `go-stock/`：来源于上游并在本仓库中复用的市场分析相关代码
- `docs/`：配置说明、审查记录和发布文档
- `scripts/`：本地开发、构建和导出脚本

## 快速开始

### 环境要求

- 与 `go.mod` 匹配的 Go 工具链
- Node.js
- Wails CLI

如未安装 `Wails`，可以执行：

```powershell
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### 安装前端依赖

```powershell
cd frontend
npm ci
cd ..
```

### 运行测试

```powershell
go test ./...
cd frontend
npm run build
```

### 启动开发环境

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\start-wails-dev.ps1
```

### 构建与导出

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\build-and-export.ps1
```

默认会导出到 `build/export/`。如果你想指定目录：

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\build-and-export.ps1 -OutputDir D:\apps\rubin-investment -SkipShortcut
```

## 运行时配置

私有运行时配置不会提交到仓库。程序会从以下位置读取本地覆盖配置：

- `data/private-overrides.json`
- 或环境变量 `INVESTMENT_PRIVATE_CONFIG_PATH`

支持的字段和示例见：

- [docs/runtime-configuration.md](docs/runtime-configuration.md)
- [docs/examples/private-overrides.example.json](docs/examples/private-overrides.example.json)
- [docs/branching-and-beta-plan.md](docs/branching-and-beta-plan.md)

## 开源说明

- 本仓库包含基于 `go-stock` 的衍生与扩展代码
- `go-stock` 上游项目采用 Apache License 2.0
- 本仓库已保留上游许可证文本与归属说明
- 某些市场页面仍依赖第三方公开接口或上游兼容地址，这些能力应视为可选集成，而非本仓库自有基础设施
- 请不要提交本地 `data/stock.db`、`data/private-overrides.json`、Cookie、Token 或其他个人敏感信息

## 安全

如果你发现漏洞或意外泄露的敏感信息，请按照 [SECURITY.md](SECURITY.md) 处理，不要在公开 Issue 中直接披露细节。

## 贡献

贡献说明见 [CONTRIBUTING.md](CONTRIBUTING.md)。

## 上游归属

本项目包含基于 `go-stock` 的衍生工作：

- 上游仓库：https://github.com/ArvinLovegood/go-stock
- 上游协议：Apache License 2.0

更多归属信息见 [NOTICE](NOTICE)。

## English Summary

`Rubin Investment` is a local-first desktop app for investment research, built with `Wails + Vue`. It reuses parts of the upstream `go-stock` project and extends them with portfolio tracking, household asset workflows, quant templates, and AI-assisted analysis.

### Highlights

- Desktop application built with `Wails`
- Local SQLite runtime data
- Reused `go-stock` market and news capabilities
- Extended with asset management, portfolio tracking, and quant workflows
- Private runtime overrides are kept out of Git

### Current Status

- The portfolio/holding module is still under active development
- The quant linkage recommendation feature is available, but it is not fully stable yet and still needs refinement

### Quick Start

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

### Build

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\build-and-export.ps1
```

### Runtime Overrides

Local private overrides are loaded from:

- `data/private-overrides.json`
- or `INVESTMENT_PRIVATE_CONFIG_PATH`

See:

- [docs/runtime-configuration.md](docs/runtime-configuration.md)
- [docs/examples/private-overrides.example.json](docs/examples/private-overrides.example.json)

### Upstream Attribution

This repository contains derivative work based on `go-stock`.

- Upstream repository: https://github.com/ArvinLovegood/go-stock
- Upstream license: Apache License 2.0
