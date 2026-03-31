package shared

import (
	"os"
	"path/filepath"
	"strings"

	"go-stock/backend/data"
	"go-stock/backend/db"
	"go-stock/backend/logger"
	"go-stock/backend/models"
	"go-stock/backend/runtimeconfig"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// SeedRuntimeDataIfNeeded imports go-stock data/config into a freshly created
// runtime database so packaged builds keep working after being moved.
func SeedRuntimeDataIfNeeded(targetDBPath string, appBaseDir string) {
	targetPath := sqliteFilePath(targetDBPath)
	sourcePath := findSeedDBPath(targetPath, appBaseDir)
	if sourcePath == "" {
		logger.SugaredLogger.Info("no seed database found for runtime bootstrap")
		ensureDefaultSettings()
		return
	}

	sourceDB, err := gorm.Open(sqlite.Open(sourcePath), &gorm.Config{
		Logger:                 gormlogger.Default.LogMode(gormlogger.Silent),
		SkipDefaultTransaction: true,
	})
	if err != nil {
		logger.SugaredLogger.Warnf("open seed database failed: %v", err)
		return
	}

	copyTableIfEmpty[data.Settings](sourceDB)
	copyTableIfEmpty[data.AIConfig](sourceDB)
	copyTableIfEmpty[data.StockBasic](sourceDB)
	copyTableIfEmpty[data.IndexBasic](sourceDB)
	copyTableIfEmpty[data.FollowedStock](sourceDB)
	copyTableIfEmpty[data.FollowedFund](sourceDB)
	copyTableIfEmpty[data.FundBasic](sourceDB)
	copyTableIfEmpty[data.Group](sourceDB)
	copyTableIfEmpty[data.GroupStock](sourceDB)
	copyTableIfEmpty[models.PromptTemplate](sourceDB)
	copyTableIfEmpty[models.StockInfoHK](sourceDB)
	copyTableIfEmpty[models.StockInfoUS](sourceDB)
	copyTableIfEmpty[models.Tags](sourceDB)
	copyTableIfEmpty[models.Telegraph](sourceDB)
	copyTableIfEmpty[models.TelegraphTags](sourceDB)
	copyTableIfEmpty[models.LongTigerRankData](sourceDB)
	copyTableIfEmpty[models.AIResponseResult](sourceDB)
	copyTableIfEmpty[models.AiRecommendStocks](sourceDB)
	copyTableIfEmpty[models.AllStockInfo](sourceDB)
	copyTableIfEmpty[models.BKDict](sourceDB)

	copyTemplatesIfNeeded(appBaseDir, sourcePath)
	ensureDefaultSettings()
}

func copyTableIfEmpty[T any](source *gorm.DB) {
	var targetCount int64
	if err := dataDB().Model(new(T)).Count(&targetCount).Error; err != nil {
		return
	}
	if targetCount > 0 {
		return
	}

	var sourceCount int64
	if err := source.Model(new(T)).Count(&sourceCount).Error; err != nil || sourceCount == 0 {
		return
	}

	rows := make([]T, 0, sourceCount)
	if err := source.Model(new(T)).Find(&rows).Error; err != nil || len(rows) == 0 {
		return
	}

	if err := dataDB().CreateInBatches(rows, 200).Error; err != nil {
		logger.SugaredLogger.Warnf("seed %T failed: %v", *new(T), err)
		return
	}
	logger.SugaredLogger.Infof("seeded %d rows into %T", len(rows), *new(T))
}

func copyTemplatesIfNeeded(appBaseDir string, sourceDBPath string) {
	targetDir := filepath.Join(appBaseDir, "data", "quant_templates")
	entries, _ := os.ReadDir(targetDir)
	if len(entries) > 0 {
		return
	}

	sourceDataDir := filepath.Dir(sourceDBPath)
	sourceTemplates := filepath.Join(sourceDataDir, "quant_templates")
	if sourceTemplates == targetDir {
		return
	}
	if info, err := os.Stat(sourceTemplates); err != nil || !info.IsDir() {
		return
	}

	_ = os.MkdirAll(targetDir, os.ModePerm)
	sourceEntries, err := os.ReadDir(sourceTemplates)
	if err != nil {
		return
	}
	for _, entry := range sourceEntries {
		if entry.IsDir() {
			continue
		}
		src := filepath.Join(sourceTemplates, entry.Name())
		dst := filepath.Join(targetDir, entry.Name())
		if _, err := os.Stat(dst); err == nil {
			continue
		}
		content, err := os.ReadFile(src)
		if err != nil {
			continue
		}
		_ = os.WriteFile(dst, content, 0644)
	}
}

func findSeedDBPath(targetPath string, appBaseDir string) string {
	candidates := []string{}

	for _, root := range searchRoots(appBaseDir) {
		if root == "" {
			continue
		}
		candidates = append(candidates, filepath.Join(root, "data", "stock.db"))
		candidates = append(candidates, filepath.Join(root, "go-stock", "data", "stock.db"))
		candidates = append(candidates, filepath.Join(root, "seed", "stock.db"))
		candidates = append(candidates, filepath.Join(root, "bootstrap", "stock.db"))
	}
	candidates = append(candidates, filepath.Join(".", "data", "stock.db"))
	candidates = append(candidates, runtimeconfig.Current().SeedDBPaths...)

	bestPath := ""
	bestScore := int64(-1)
	seen := map[string]bool{}
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		resolved, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		if seen[resolved] || sameFilePath(resolved, targetPath) {
			continue
		}
		seen[resolved] = true
		info, err := os.Stat(resolved)
		if err == nil && !info.IsDir() && info.Size() > 0 {
			score := scoreSeedDB(resolved)
			if score > bestScore {
				bestScore = score
				bestPath = resolved
			}
		}
	}
	if bestPath != "" {
		logger.SugaredLogger.Infof("using external seed database for runtime bootstrap (score=%d)", bestScore)
	}
	return bestPath
}

func searchRoots(appBaseDir string) []string {
	roots := []string{}
	if wd, err := os.Getwd(); err == nil {
		roots = append(roots, findProjectRoot(wd))
	}
	roots = append(roots, findProjectRoot(appBaseDir))
	return roots
}

func scoreSeedDB(dbPath string) int64 {
	sourceDB, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger:                 gormlogger.Default.LogMode(gormlogger.Silent),
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return -1
	}

	weights := []struct {
		table  string
		weight int64
	}{
		{"settings", 1000},
		{"ai_config", 500},
		{"tushare_stock_basic", 10},
		{"stock_base_info_hk", 3},
		{"stock_base_info_us", 1},
		{"all_stock_info", 4},
		{"telegraph_list", 20},
		{"tags", 20},
		{"word_analyzes", 30},
		{"sentiment_result_analyzes", 30},
		{"fund_basic", 10},
	}

	var score int64
	for _, item := range weights {
		var count int64
		if err := sourceDB.Table(item.table).Count(&count).Error; err == nil {
			score += count * item.weight
		}
	}
	return score
}

func ensureDefaultSettings() {
	var count int64
	if err := dataDB().Model(&data.Settings{}).Count(&count).Error; err != nil || count > 0 {
		return
	}

	defaults := &data.Settings{
		LocalPushEnable:        false,
		DingPushEnable:         false,
		UpdateBasicInfoOnStart: false,
		RefreshInterval:        60,
		OpenAiEnable:           false,
		CheckUpdate:            false,
		CrawlTimeOut:           60,
		KDays:                  60,
		EnableDanmu:            false,
		EnableNews:             true,
		DarkTheme:              false,
		BrowserPoolSize:        1,
		EnableFund:             true,
		EnablePushNews:         false,
		EnableOnlyPushRedNews:  false,
		EnableAgent:            false,
		AssetUnlockPassword:    runtimeconfig.Current().AssetUnlockPassword,
	}
	if err := dataDB().Create(defaults).Error; err != nil {
		logger.SugaredLogger.Warnf("create default settings failed: %v", err)
		return
	}
	logger.SugaredLogger.Info("created default go-stock settings for integrated runtime")
}

func findProjectRoot(start string) string {
	dir := start
	for dir != "" && dir != filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "wails.json")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	return ""
}

func sameFilePath(a string, b string) bool {
	return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
}

func sqliteFilePath(dsn string) string {
	if idx := strings.Index(dsn, "?"); idx >= 0 {
		return dsn[:idx]
	}
	return dsn
}

func dataDB() *gorm.DB {
	return db.Dao
}
