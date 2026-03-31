package portfolio

import (
	"math"
	"time"

	"go-stock/backend/db"
	"go-stock/backend/logger"
)

// Service 持仓分析服务
type Service struct{}

// NewService 创建持仓分析服务
func NewService() *Service {
	return &Service{}
}

// --- 持仓 CRUD ---

// GetAllHoldings 获取所有持仓
func (s *Service) GetAllHoldings() []Holding {
	var holdings []Holding
	db.Dao.Where("quantity > 0").Order("holding_type asc, stock_code asc").Find(&holdings)
	return holdings
}

// GetHoldingsByType 按类型获取持仓
func (s *Service) GetHoldingsByType(holdingType string) []Holding {
	var holdings []Holding
	db.Dao.Where("holding_type = ? AND quantity > 0", holdingType).Find(&holdings)
	return holdings
}

// GetHoldingByCode 根据代码获取持仓
func (s *Service) GetHoldingByCode(stockCode string) *Holding {
	var h Holding
	err := db.Dao.Where("stock_code = ? AND quantity > 0", stockCode).First(&h).Error
	if err != nil {
		return nil
	}
	return &h
}

// CreateHolding 创建持仓
func (s *Service) CreateHolding(h Holding) *Holding {
	// 计算初始值
	h.TotalCost = h.AvgCost * float64(h.Quantity)
	h.TotalValue = h.CurrentPrice * float64(h.Quantity)
	h.ProfitLoss = h.TotalValue - h.TotalCost
	if h.TotalCost > 0 {
		h.ProfitRate = math.Round((h.ProfitLoss/h.TotalCost)*10000) / 100
	}

	err := db.Dao.Create(&h).Error
	if err != nil {
		logger.SugaredLogger.Errorf("创建持仓失败: %v", err)
		return nil
	}
	return &h
}

// UpdateHolding 更新持仓
func (s *Service) UpdateHolding(h Holding) *Holding {
	err := db.Dao.Save(&h).Error
	if err != nil {
		logger.SugaredLogger.Errorf("更新持仓失败: %v", err)
		return nil
	}
	return &h
}

// DeleteHolding 删除持仓（软删除）
func (s *Service) DeleteHolding(id uint) bool {
	err := db.Dao.Delete(&Holding{}, id).Error
	return err == nil
}

// --- 交易记录 ---

// AddTransaction 添加交易记录并更新持仓
func (s *Service) AddTransaction(tx Transaction) *Transaction {
	// 计算成交金额
	tx.Amount = tx.Price * float64(tx.Quantity)

	// 获取或创建持仓
	var holding Holding
	result := db.Dao.Where("stock_code = ? AND quantity > 0", tx.StockCode).First(&holding)

	if tx.Type == "buy" {
		if result.Error != nil {
			// 新建持仓
			now := time.Now()
			holding = Holding{
				StockCode:   tx.StockCode,
				StockName:   tx.StockName,
				HoldingType: "stock",
				AvgCost:     tx.Price,
				Quantity:    tx.Quantity,
				TotalCost:   tx.Amount + tx.Fee,
				BuyDate:     &now,
			}
			db.Dao.Create(&holding)
		} else {
			// 更新持仓：加权平均成本
			newTotalCost := holding.TotalCost + tx.Amount + tx.Fee
			newQuantity := holding.Quantity + tx.Quantity
			holding.AvgCost = newTotalCost / float64(newQuantity)
			holding.Quantity = newQuantity
			holding.TotalCost = newTotalCost
			db.Dao.Save(&holding)
		}
	} else if tx.Type == "sell" {
		if result.Error == nil {
			holding.Quantity -= tx.Quantity
			holding.TotalCost = holding.AvgCost * float64(holding.Quantity)
			db.Dao.Save(&holding)
		}
	}

	tx.HoldingID = holding.ID
	err := db.Dao.Create(&tx).Error
	if err != nil {
		logger.SugaredLogger.Errorf("添加交易记录失败: %v", err)
		return nil
	}
	return &tx
}

// GetTransactions 获取交易记录
func (s *Service) GetTransactions(stockCode string, page, pageSize int) ([]Transaction, int64) {
	var transactions []Transaction
	var total int64

	query := db.Dao.Model(&Transaction{})
	if stockCode != "" {
		query = query.Where("stock_code = ?", stockCode)
	}

	query.Count(&total)
	query.Order("trade_date desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&transactions)

	return transactions, total
}

// --- 实时收益更新 ---

// UpdateHoldingPrice 更新持仓当前价格并重算盈亏
func (s *Service) UpdateHoldingPrice(stockCode string, currentPrice, todayChange, todayRate float64) {
	var h Holding
	err := db.Dao.Where("stock_code = ? AND quantity > 0", stockCode).First(&h).Error
	if err != nil {
		return
	}

	h.CurrentPrice = currentPrice
	h.TotalValue = currentPrice * float64(h.Quantity)
	h.ProfitLoss = h.TotalValue - h.TotalCost
	if h.TotalCost > 0 {
		h.ProfitRate = math.Round((h.ProfitLoss/h.TotalCost)*10000) / 100
	}
	h.TodayChange = todayChange * float64(h.Quantity)
	h.TodayRate = todayRate

	db.Dao.Save(&h)
}

// --- 汇总 ---

// GetPortfolioSummary 获取持仓汇总
func (s *Service) GetPortfolioSummary() *PortfolioSummary {
	holdings := s.GetAllHoldings()
	summary := &PortfolioSummary{
		Holdings: holdings,
	}

	for _, h := range holdings {
		summary.TotalCost += h.TotalCost
		summary.TotalValue += h.TotalValue
		summary.TotalProfit += h.ProfitLoss
		summary.TodayProfit += h.TodayChange

		if h.HoldingType == "stock" {
			summary.StockCount++
		} else {
			summary.FundCount++
		}
	}

	if summary.TotalCost > 0 {
		summary.TotalProfitRate = math.Round((summary.TotalProfit/summary.TotalCost)*10000) / 100
	}

	return summary
}

// SaveDailySnapshot 保存当日收益快照
func (s *Service) SaveDailySnapshot() {
	summary := s.GetPortfolioSummary()
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	snapshot := ProfitSnapshot{
		SnapshotDate: &today,
		TotalCost:    summary.TotalCost,
		TotalValue:   summary.TotalValue,
		TotalProfit:  summary.TotalProfit,
		ProfitRate:   summary.TotalProfitRate,
	}

	// 分别计算股票和基金市值
	for _, h := range summary.Holdings {
		if h.HoldingType == "stock" {
			snapshot.StockValue += h.TotalValue
		} else {
			snapshot.FundValue += h.TotalValue
		}
	}

	// Upsert
	var existing ProfitSnapshot
	err := db.Dao.Where("snapshot_date = ?", today).First(&existing).Error
	if err != nil {
		db.Dao.Create(&snapshot)
	} else {
		snapshot.ID = existing.ID
		db.Dao.Save(&snapshot)
	}
}

// GetProfitHistory 获取收益历史
func (s *Service) GetProfitHistory(days int) []ProfitSnapshot {
	var snapshots []ProfitSnapshot
	startDate := time.Now().AddDate(0, 0, -days)
	db.Dao.Where("snapshot_date >= ?", startDate).
		Order("snapshot_date asc").
		Find(&snapshots)
	return snapshots
}
