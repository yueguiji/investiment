package portfolio

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

// Holding 持仓记录
type Holding struct {
	gorm.Model
	StockCode    string                `json:"stockCode" gorm:"index"`          // 股票/基金代码
	StockName    string                `json:"stockName"`                       // 名称
	HoldingType  string                `json:"holdingType" gorm:"index"`        // stock / fund
	Market       string                `json:"market"`                          // A股/港股/美股: a / hk / us
	AvgCost      float64               `json:"avgCost"`                         // 平均成本
	Quantity     int64                 `json:"quantity"`                        // 持有数量（基金为份额×100）
	CurrentPrice float64               `json:"currentPrice"`                    // 当前价格（实时更新）
	ProfitLoss   float64               `json:"profitLoss"`                      // 浮动盈亏金额
	ProfitRate   float64               `json:"profitRate"`                      // 收益率 %
	TotalCost    float64               `json:"totalCost"`                       // 总成本
	TotalValue   float64               `json:"totalValue"`                      // 当前市值
	TodayChange  float64               `json:"todayChange"`                     // 今日涨跌额
	TodayRate    float64               `json:"todayRate"`                       // 今日涨跌幅 %
	BuyDate      *time.Time            `json:"buyDate"`                         // 首次建仓日期
	BrokerName   string                `json:"brokerName"`                      // 券商名称
	AccountTag   string                `json:"accountTag"`                      // 账户标签（如：主账户/量化账户）
	Remark       string                `json:"remark"`                          // 备注
	IsDel        soft_delete.DeletedAt `gorm:"softDelete:flag" json:"-"`
}

func (h Holding) TableName() string {
	return "holdings"
}

// Transaction 交易记录
type Transaction struct {
	gorm.Model
	HoldingID   uint                  `json:"holdingId" gorm:"index"`          // 关联持仓
	StockCode   string                `json:"stockCode" gorm:"index"`
	StockName   string                `json:"stockName"`
	Type        string                `json:"type"`                            // buy / sell
	Price       float64               `json:"price"`                           // 成交价
	Quantity    int64                 `json:"quantity"`                        // 成交数量
	Amount      float64               `json:"amount"`                          // 成交金额
	Fee         float64               `json:"fee"`                             // 手续费
	TradeDate   *time.Time            `json:"tradeDate" gorm:"index"`          // 交易日期
	Remark      string                `json:"remark"`
	IsDel       soft_delete.DeletedAt `gorm:"softDelete:flag" json:"-"`
}

func (t Transaction) TableName() string {
	return "transactions"
}

// ProfitSnapshot 收益快照（每日记录一次）
type ProfitSnapshot struct {
	gorm.Model
	SnapshotDate *time.Time `json:"snapshotDate" gorm:"uniqueIndex"` // 快照日期
	TotalCost    float64    `json:"totalCost"`                       // 总成本
	TotalValue   float64    `json:"totalValue"`                      // 总市值
	TotalProfit  float64    `json:"totalProfit"`                     // 总盈亏
	ProfitRate   float64    `json:"profitRate"`                      // 总收益率 %
	StockValue   float64    `json:"stockValue"`                      // 股票总市值
	FundValue    float64    `json:"fundValue"`                       // 基金总市值
}

func (p ProfitSnapshot) TableName() string {
	return "profit_snapshots"
}

// PortfolioSummary 持仓汇总（计算用，不持久化）
type PortfolioSummary struct {
	TotalCost       float64         `json:"totalCost"`       // 总成本
	TotalValue      float64         `json:"totalValue"`      // 总市值
	TotalProfit     float64         `json:"totalProfit"`     // 总盈亏
	TotalProfitRate float64         `json:"totalProfitRate"` // 总收益率
	TodayProfit     float64         `json:"todayProfit"`     // 今日盈亏
	StockCount      int             `json:"stockCount"`      // 股票持仓数
	FundCount       int             `json:"fundCount"`       // 基金持仓数
	Holdings        []Holding       `json:"holdings"`        // 所有持仓
}
