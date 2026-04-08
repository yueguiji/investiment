package shared

import (
	"path/filepath"

	"go-stock/backend/data"
	"go-stock/backend/db"
	"go-stock/backend/logger"
	"go-stock/backend/models"
	"investment-platform/internal/asset"
	"investment-platform/internal/portfolio"
	"investment-platform/internal/quant"
)

// InitDB initializes the shared database used by both the host app and go-stock features.
func InitDB(dbPath string) {
	db.Init(dbPath)

	logger.SugaredLogger.Info("starting database migrations")

	hadFundWatchlistColumn := db.Dao.Migrator().HasColumn(&data.FollowedFund{}, "is_watchlist")

	// go-stock tables required by the integrated investment pages
	db.Dao.AutoMigrate(&data.StockInfo{})
	db.Dao.AutoMigrate(&data.StockBasic{})
	db.Dao.AutoMigrate(&data.FollowedStock{})
	db.Dao.AutoMigrate(&data.IndexBasic{})
	db.Dao.AutoMigrate(&data.Settings{})
	db.Dao.AutoMigrate(&models.AIResponseResult{})
	db.Dao.AutoMigrate(&models.StockChangeHistory{})
	db.Dao.AutoMigrate(&models.StockInfoHK{})
	db.Dao.AutoMigrate(&models.StockInfoUS{})
	db.Dao.AutoMigrate(&data.FollowedFund{})
	db.Dao.AutoMigrate(&data.FundEstimateSnapshot{})
	db.Dao.AutoMigrate(&data.FundBasic{})
	db.Dao.AutoMigrate(&models.PromptTemplate{})
	db.Dao.AutoMigrate(&data.Group{})
	db.Dao.AutoMigrate(&data.GroupStock{})
	db.Dao.AutoMigrate(&models.Tags{})
	db.Dao.AutoMigrate(&models.Telegraph{})
	db.Dao.AutoMigrate(&models.TelegraphTags{})
	db.Dao.AutoMigrate(&models.LongTigerRankData{})
	db.Dao.AutoMigrate(&data.AIConfig{})
	db.Dao.AutoMigrate(&models.BKDict{})
	db.Dao.AutoMigrate(&models.WordAnalyze{})
	db.Dao.AutoMigrate(&models.SentimentResultAnalyze{})
	db.Dao.AutoMigrate(&models.AiRecommendStocks{})
	db.Dao.AutoMigrate(&models.AllStockInfo{})

	// host app tables
	db.Dao.AutoMigrate(&asset.Asset{})
	db.Dao.AutoMigrate(&asset.AssetCategory{})
	db.Dao.AutoMigrate(&asset.HouseholdAccount{})
	db.Dao.AutoMigrate(&asset.HouseholdFixedAsset{})
	db.Dao.AutoMigrate(&asset.HouseholdIncome{})
	db.Dao.AutoMigrate(&asset.HouseholdProtection{})
	db.Dao.AutoMigrate(&asset.HouseholdLiability{})
	db.Dao.AutoMigrate(&asset.HouseholdLiabilitySchedule{})
	db.Dao.AutoMigrate(&asset.HouseholdSnapshot{})
	db.Dao.AutoMigrate(&asset.HouseholdAIAnalysis{})
	db.Dao.AutoMigrate(&asset.HouseholdBenchmarkRecord{})
	db.Dao.AutoMigrate(&asset.HouseholdProfile{})
	db.Dao.AutoMigrate(&asset.HouseholdMember{})
	db.Dao.AutoMigrate(&portfolio.Holding{})
	db.Dao.AutoMigrate(&portfolio.Transaction{})
	db.Dao.AutoMigrate(&portfolio.ProfitSnapshot{})
	db.Dao.AutoMigrate(&quant.Template{})
	db.Dao.AutoMigrate(&quant.TemplateCategory{})

	if !hadFundWatchlistColumn {
		logger.SugaredLogger.Info("migrating followed_fund watchlist flags")
		db.Dao.Exec(`
			UPDATE followed_fund
			SET is_watchlist = CASE
				WHEN code IN (SELECT stock_code FROM holdings WHERE quantity > 0) THEN 0
				ELSE 1
			END
		`)
	}

	SeedRuntimeDataIfNeeded(dbPath, runtimeBaseDir(dbPath))

	logger.SugaredLogger.Info("database migrations completed")
}

func runtimeBaseDir(dbPath string) string {
	path := sqliteFilePath(dbPath)
	return filepath.Dir(filepath.Dir(path))
}
