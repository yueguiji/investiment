# Runtime Configuration

This repository keeps private credentials and local data sources outside Git.

## Where Runtime Overrides Are Read From

The app reads local-only runtime overrides from:

1. `INVESTMENT_PRIVATE_CONFIG_PATH`
2. `data/private-overrides.json`

The loader is implemented in [go-stock/backend/runtimeconfig/runtimeconfig.go](../go-stock/backend/runtimeconfig/runtimeconfig.go).

## Supported Fields

```json
{
  "seedDbPaths": [
    "D:/data/stock.db"
  ],
  "eastmoneyQgqpBId": "",
  "assetUnlockPassword": "",
  "xueqiuCookie": "",
  "jiuyangongsheToken": "",
  "jiuyangongsheCookie": "",
  "newsSyncUrl": "http://go-stock.sparkmemory.top:16666/FinancialNews/json",
  "shareUploadUrl": "http://go-stock.sparkmemory.top:16688/upload",
  "stockBasicUrl": "http://8.134.249.145:18080/go-stock/stock_basic.json",
  "stockBaseInfoHkUrl": "http://8.134.249.145:18080/go-stock/stock_base_info_hk.json",
  "stockBaseInfoUsUrl": "http://8.134.249.145:18080/go-stock/stock_base_info_us.json",
  "danmuWebsocketUrl": "ws://8.134.249.145:16688/ws",
  "messageWallUrl": "https://go-stock.sparkmemory.top:16667/go-stock",
  "releaseLatestUrl": "https://api.github.com/repos/ArvinLovegood/go-stock/releases/latest",
  "releaseTagBaseUrl": "https://api.github.com/repos/ArvinLovegood/go-stock/git/ref/tags",
  "releaseDownloadBaseUrl": "https://github.com/ArvinLovegood/go-stock/releases/download",
  "releaseProxyDownloadBaseUrl": "https://gitproxy.click/https://github.com/ArvinLovegood/go-stock/releases/download",
  "releasePageUrl": "https://github.com/ArvinLovegood/go-stock/releases"
}
```

## Recommended Local Setup

### Option A: Use the default local file

Create `data/private-overrides.json` locally and keep it out of Git.

### Option B: Use an environment variable

```powershell
$env:INVESTMENT_PRIVATE_CONFIG_PATH = "D:\private\investment\private-overrides.json"
```

## Seed Database Bootstrap

The app can bootstrap runtime data from local SQLite seed databases. Candidate lookup is described in [internal/shared/bootstrap.go](../internal/shared/bootstrap.go).

Recommended approach:

- keep your real database outside the repository
- point `seedDbPaths` at the local copy
- let the application seed runtime data on first launch

## Publishing Rules

- Never commit `data/private-overrides.json`
- Never commit a personal `data/stock.db`
- Never commit browser cookies or account tokens
- If a feature depends on a private token, document the public fallback behavior

## Service Endpoint Overrides

The following fields are useful when you want to keep source code clean while still using your preferred upstream or self-hosted services locally:

- `assetUnlockPassword`
- `newsSyncUrl`
- `shareUploadUrl`
- `stockBasicUrl`
- `stockBaseInfoHkUrl`
- `stockBaseInfoUsUrl`
- `danmuWebsocketUrl`
- `messageWallUrl`
- `releaseLatestUrl`
- `releaseTagBaseUrl`
- `releaseDownloadBaseUrl`
- `releaseProxyDownloadBaseUrl`
- `releasePageUrl`
