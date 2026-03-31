package bridge

import (
	"strconv"
	"time"

	"go-stock/backend/data"
	"go-stock/backend/db"
	"go-stock/backend/logger"
)

// Bridge adapts reusable go-stock backend capabilities for this app.
type Bridge struct{}

func NewBridge() *Bridge {
	return &Bridge{}
}

func (b *Bridge) GetFollowedStockCodes() []string {
	var followedStocks []data.FollowedStock
	db.Dao.Find(&followedStocks)

	codes := make([]string, 0, len(followedStocks))
	for _, s := range followedStocks {
		codes = append(codes, s.StockCode)
	}
	return codes
}

func (b *Bridge) GetStockRealtimePrice(stockCode string) (float64, float64, float64) {
	stockDataAPI := data.NewStockDataApi()
	stockInfos, err := stockDataAPI.GetStockCodeRealTimeData(stockCode)
	if err != nil || stockInfos == nil || len(*stockInfos) == 0 {
		logger.SugaredLogger.Warnf("failed to get realtime stock data: %s, err=%v", stockCode, err)
		return 0, 0, 0
	}

	stockInfo := (*stockInfos)[0]
	price, err := strconv.ParseFloat(stockInfo.Price, 64)
	if err != nil {
		logger.SugaredLogger.Warnf("failed to parse realtime price for %s: %v", stockCode, err)
		return 0, 0, 0
	}

	return price, stockInfo.ChangePrice, stockInfo.ChangePercent
}

func (b *Bridge) GetSettingConfig() *data.SettingConfig {
	return data.GetSettingConfig()
}

func (b *Bridge) IsMarketOpen() bool {
	now := time.Now()
	weekday := now.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}

	currentMinutes := now.Hour()*60 + now.Minute()
	morningOpen := 9*60 + 30
	morningClose := 11*60 + 30
	afternoonOpen := 13 * 60
	afternoonClose := 15 * 60

	return (currentMinutes >= morningOpen && currentMinutes <= morningClose) ||
		(currentMinutes >= afternoonOpen && currentMinutes <= afternoonClose)
}
