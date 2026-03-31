package runtimeconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go-stock/backend/logger"
	"go-stock/backend/runtimepaths"
)

const privateConfigEnvVar = "INVESTMENT_PRIVATE_CONFIG_PATH"

type PrivateConfig struct {
	SeedDBPaths                 []string `json:"seedDbPaths"`
	EastmoneyQgqpBId            string   `json:"eastmoneyQgqpBId"`
	XueqiuCookie                string   `json:"xueqiuCookie"`
	JiuyangongsheToken          string   `json:"jiuyangongsheToken"`
	JiuyangongsheCookie         string   `json:"jiuyangongsheCookie"`
	AssetUnlockPassword         string   `json:"assetUnlockPassword"`
	ReleaseLatestURL            string   `json:"releaseLatestUrl"`
	ReleaseTagBaseURL           string   `json:"releaseTagBaseUrl"`
	ReleaseDownloadBaseURL      string   `json:"releaseDownloadBaseUrl"`
	ReleaseProxyDownloadBaseURL string   `json:"releaseProxyDownloadBaseUrl"`
	ReleasePageURL              string   `json:"releasePageUrl"`
	NewsSyncURL                 string   `json:"newsSyncUrl"`
	ShareUploadURL              string   `json:"shareUploadUrl"`
	StockBasicURL               string   `json:"stockBasicUrl"`
	StockBaseInfoHKURL          string   `json:"stockBaseInfoHkUrl"`
	StockBaseInfoUSURL          string   `json:"stockBaseInfoUsUrl"`
	DanmuWebsocketURL           string   `json:"danmuWebsocketUrl"`
	MessageWallURL              string   `json:"messageWallUrl"`
}

const (
	defaultReleaseLatestURL            = "https://api.github.com/repos/ArvinLovegood/go-stock/releases/latest"
	defaultReleaseTagBaseURL           = "https://api.github.com/repos/ArvinLovegood/go-stock/git/ref/tags"
	defaultReleaseDownloadBaseURL      = "https://github.com/ArvinLovegood/go-stock/releases/download"
	defaultReleaseProxyDownloadBaseURL = "https://gitproxy.click/https://github.com/ArvinLovegood/go-stock/releases/download"
	defaultReleasePageURL              = "https://github.com/ArvinLovegood/go-stock/releases"
	defaultNewsSyncURL                 = "http://go-stock.sparkmemory.top:16666/FinancialNews/json"
	defaultShareUploadURL              = "http://go-stock.sparkmemory.top:16688/upload"
	defaultStockBasicURL               = "http://8.134.249.145:18080/go-stock/stock_basic.json"
	defaultStockBaseInfoHKURL          = "http://8.134.249.145:18080/go-stock/stock_base_info_hk.json"
	defaultStockBaseInfoUSURL          = "http://8.134.249.145:18080/go-stock/stock_base_info_us.json"
	defaultDanmuWebsocketURL           = "ws://8.134.249.145:16688/ws"
	defaultMessageWallURL              = "https://go-stock.sparkmemory.top:16667/go-stock"
)

var (
	loadOnce sync.Once
	current  *PrivateConfig
)

func Current() *PrivateConfig {
	loadOnce.Do(func() {
		current = load()
	})
	return current
}

func ConfigPath() string {
	if value := strings.TrimSpace(os.Getenv(privateConfigEnvVar)); value != "" {
		return value
	}
	return filepath.Join(runtimepaths.DataDir(), "private-overrides.json")
}

func load() *PrivateConfig {
	cfg := &PrivateConfig{}
	content, err := os.ReadFile(ConfigPath())
	if err != nil {
		if !os.IsNotExist(err) {
			logger.SugaredLogger.Warnf("load private runtime config failed: %v", err)
		}
		return cfg
	}

	if err := json.Unmarshal(content, cfg); err != nil {
		logger.SugaredLogger.Warnf("parse private runtime config failed: %v", err)
		return &PrivateConfig{}
	}

	cfg.EastmoneyQgqpBId = strings.TrimSpace(cfg.EastmoneyQgqpBId)
	cfg.XueqiuCookie = strings.TrimSpace(cfg.XueqiuCookie)
	cfg.JiuyangongsheToken = strings.TrimSpace(cfg.JiuyangongsheToken)
	cfg.JiuyangongsheCookie = strings.TrimSpace(cfg.JiuyangongsheCookie)
	cfg.AssetUnlockPassword = strings.TrimSpace(cfg.AssetUnlockPassword)
	cfg.ReleaseLatestURL = strings.TrimSpace(cfg.ReleaseLatestURL)
	cfg.ReleaseTagBaseURL = strings.TrimSpace(cfg.ReleaseTagBaseURL)
	cfg.ReleaseDownloadBaseURL = strings.TrimSpace(cfg.ReleaseDownloadBaseURL)
	cfg.ReleaseProxyDownloadBaseURL = strings.TrimSpace(cfg.ReleaseProxyDownloadBaseURL)
	cfg.ReleasePageURL = strings.TrimSpace(cfg.ReleasePageURL)
	cfg.NewsSyncURL = strings.TrimSpace(cfg.NewsSyncURL)
	cfg.ShareUploadURL = strings.TrimSpace(cfg.ShareUploadURL)
	cfg.StockBasicURL = strings.TrimSpace(cfg.StockBasicURL)
	cfg.StockBaseInfoHKURL = strings.TrimSpace(cfg.StockBaseInfoHKURL)
	cfg.StockBaseInfoUSURL = strings.TrimSpace(cfg.StockBaseInfoUSURL)
	cfg.DanmuWebsocketURL = strings.TrimSpace(cfg.DanmuWebsocketURL)
	cfg.MessageWallURL = strings.TrimSpace(cfg.MessageWallURL)
	cfg.SeedDBPaths = cleanedPaths(cfg.SeedDBPaths)

	logger.SugaredLogger.Info("private runtime config loaded")
	return cfg
}

func cleanedPaths(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	result := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		resolved, err := filepath.Abs(value)
		if err == nil {
			value = resolved
		}
		key := strings.ToLower(filepath.Clean(value))
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, value)
	}
	return result
}

func (cfg *PrivateConfig) ResolveReleaseLatestURL() string {
	return firstNonEmpty(cfg.ReleaseLatestURL, defaultReleaseLatestURL)
}

func (cfg *PrivateConfig) ResolveReleaseTagURL(tag string) string {
	return joinURL(firstNonEmpty(cfg.ReleaseTagBaseURL, defaultReleaseTagBaseURL), tag)
}

func (cfg *PrivateConfig) ResolveReleaseDownloadURL(tag string, fileName string, useProxy bool) string {
	base := cfg.ReleaseDownloadBaseURL
	if useProxy {
		base = cfg.ReleaseProxyDownloadBaseURL
	}
	if useProxy && strings.TrimSpace(base) == "" {
		base = defaultReleaseProxyDownloadBaseURL
	}
	if !useProxy && strings.TrimSpace(base) == "" {
		base = defaultReleaseDownloadBaseURL
	}
	return joinURL(base, tag, fileName)
}

func (cfg *PrivateConfig) ResolveReleasePageURL() string {
	return firstNonEmpty(cfg.ReleasePageURL, defaultReleasePageURL)
}

func (cfg *PrivateConfig) ResolveNewsSyncURL(sinceUnix int64) string {
	base := firstNonEmpty(cfg.NewsSyncURL, defaultNewsSyncURL)
	if strings.Contains(base, "?") {
		return base
	}
	return fmt.Sprintf("%s?since=%d", strings.TrimRight(base, "/"), sinceUnix)
}

func (cfg *PrivateConfig) ResolveShareUploadURL() string {
	return firstNonEmpty(cfg.ShareUploadURL, defaultShareUploadURL)
}

func (cfg *PrivateConfig) ResolveStockBasicURL() string {
	return firstNonEmpty(cfg.StockBasicURL, defaultStockBasicURL)
}

func (cfg *PrivateConfig) ResolveStockBaseInfoHKURL() string {
	return firstNonEmpty(cfg.StockBaseInfoHKURL, defaultStockBaseInfoHKURL)
}

func (cfg *PrivateConfig) ResolveStockBaseInfoUSURL() string {
	return firstNonEmpty(cfg.StockBaseInfoUSURL, defaultStockBaseInfoUSURL)
}

func (cfg *PrivateConfig) ResolveDanmuWebsocketURL() string {
	return firstNonEmpty(cfg.DanmuWebsocketURL, defaultDanmuWebsocketURL)
}

func (cfg *PrivateConfig) ResolveMessageWallURL() string {
	return firstNonEmpty(cfg.MessageWallURL, defaultMessageWallURL)
}

func (cfg *PrivateConfig) HasAssetUnlockPassword() bool {
	return strings.TrimSpace(cfg.AssetUnlockPassword) != ""
}

func (cfg *PrivateConfig) VerifyAssetUnlockPassword(password string) bool {
	if !cfg.HasAssetUnlockPassword() {
		return true
	}
	return strings.TrimSpace(password) == strings.TrimSpace(cfg.AssetUnlockPassword)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func joinURL(base string, parts ...string) string {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	for _, part := range parts {
		part = strings.Trim(strings.TrimSpace(part), "/")
		if part == "" {
			continue
		}
		base += "/" + part
	}
	return base
}
