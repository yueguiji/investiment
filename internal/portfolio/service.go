package portfolio

import (
	"math"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-stock/backend/data"
	"go-stock/backend/db"
	"go-stock/backend/logger"
	"gorm.io/gorm"
)

type Service struct{}

type fundRefreshRuntimeState struct {
	mu            sync.Mutex
	refreshing    bool
	scope         string
	progressNow   int64
	progressTotal int64
	currentCode   string
	lastStarted   time.Time
	lastFinished  time.Time
}

var fundRefreshState fundRefreshRuntimeState

const (
	fundRefreshStateNotStarted = "not_started"
	fundRefreshStatePartial    = "partial"
	fundRefreshStateCompleted  = "completed"
	fundRefreshScopeFocused    = "watchlist_related"
	fundRefreshScopeAll        = "all_pending"
)

type betterCandidateUniverse struct {
	Basics           []data.FundBasic
	ScopeLabel       string
	FallbackApplied  bool
	ComparedUniverse int
}

type fundRefreshScopeSnapshot struct {
	Scope        string
	TargetCount  int64
	UpdatedToday int64
	PendingCount int64
}

type betterMetricSpec struct {
	Key     string
	Label   string
	Better  string
	Weight  float64
	Format  string
	ValueOf func(data.FundBasic) *float64
}

type betterMetricRank struct {
	Rank       int
	Total      int
	Percentile *float64
}

type fundRiskSnapshot struct {
	MaxDrawdown12 *float64
	Volatility12  *float64
	Sharpe12      *float64
	Calmar12      *float64
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) GetAllHoldings() []Holding {
	var holdings []Holding
	db.Dao.Where("quantity > 0").Order("holding_type asc, total_value desc, stock_code asc").Find(&holdings)
	return holdings
}

func (s *Service) GetHoldingsByType(holdingType string) []Holding {
	var holdings []Holding
	db.Dao.Where("holding_type = ? AND quantity > 0", normalizeHoldingType(holdingType)).Order("total_value desc, stock_code asc").Find(&holdings)
	return holdings
}

func (s *Service) GetHoldingByCode(stockCode string) *Holding {
	var holding Holding
	if err := db.Dao.Where("stock_code = ? AND quantity > 0", strings.TrimSpace(stockCode)).First(&holding).Error; err != nil {
		return nil
	}
	return &holding
}

func (s *Service) CreateHolding(h Holding) *Holding {
	h.StockCode = strings.TrimSpace(h.StockCode)
	h.StockName = strings.TrimSpace(h.StockName)
	h.BrokerName = strings.TrimSpace(h.BrokerName)
	h.AccountTag = strings.TrimSpace(h.AccountTag)
	h.HoldingType = normalizeHoldingType(h.HoldingType)
	if h.HoldingType == "" {
		h.HoldingType = inferHoldingTypeByCode(h.StockCode)
	}

	h.TotalCost = h.AvgCost * h.Quantity
	h.TotalValue = h.CurrentPrice * h.Quantity
	h.ProfitLoss = h.TotalValue - h.TotalCost
	if h.TotalCost > 0 {
		h.ProfitRate = roundPercent(h.ProfitLoss / h.TotalCost * 100)
	}

	if err := db.Dao.Create(&h).Error; err != nil {
		logger.SugaredLogger.Errorf("create holding failed: %v", err)
		return nil
	}

	switch h.HoldingType {
	case "fund":
		s.refreshSingleFundHolding(&h)
	case "stock":
		s.refreshSingleStockHolding(&h)
	}
	db.Dao.Save(&h)
	return &h
}

func (s *Service) UpdateHolding(h Holding) *Holding {
	h.StockCode = strings.TrimSpace(h.StockCode)
	h.StockName = strings.TrimSpace(h.StockName)
	h.BrokerName = strings.TrimSpace(h.BrokerName)
	h.AccountTag = strings.TrimSpace(h.AccountTag)
	h.HoldingType = normalizeHoldingType(h.HoldingType)
	if h.HoldingType == "" {
		h.HoldingType = inferHoldingTypeByCode(h.StockCode)
	}

	h.TotalCost = h.AvgCost * h.Quantity
	h.TotalValue = h.CurrentPrice * h.Quantity
	h.ProfitLoss = h.TotalValue - h.TotalCost
	if h.TotalCost > 0 {
		h.ProfitRate = roundPercent(h.ProfitLoss / h.TotalCost * 100)
	}

	if err := db.Dao.Save(&h).Error; err != nil {
		logger.SugaredLogger.Errorf("update holding failed: %v", err)
		return nil
	}
	return &h
}

func (s *Service) DeleteHolding(id uint) bool {
	return db.Dao.Delete(&Holding{}, id).Error == nil
}

func (s *Service) AddTransaction(tx Transaction) *Transaction {
	tx.StockCode = strings.TrimSpace(tx.StockCode)
	tx.StockName = strings.TrimSpace(tx.StockName)
	tx.BrokerName = strings.TrimSpace(tx.BrokerName)
	tx.AccountTag = strings.TrimSpace(tx.AccountTag)
	tx.HoldingType = normalizeHoldingType(tx.HoldingType)
	if tx.HoldingType == "" {
		tx.HoldingType = inferHoldingTypeByCode(tx.StockCode)
	}
	tx.Amount = tx.Price * tx.Quantity

	var holding Holding
	result := db.Dao.Where("stock_code = ? AND quantity > 0", tx.StockCode).First(&holding)
	if result.Error == nil && holding.HoldingType != "" {
		tx.HoldingType = holding.HoldingType
	}

	switch tx.Type {
	case "buy":
		if result.Error != nil {
			now := time.Now()
			holding = Holding{
				StockCode:    tx.StockCode,
				StockName:    tx.StockName,
				HoldingType:  tx.HoldingType,
				AvgCost:      tx.Price,
				Quantity:     tx.Quantity,
				CurrentPrice: tx.Price,
				TotalCost:    tx.Amount + tx.Fee,
				TotalValue:   tx.Amount,
				BuyDate:      &now,
				BrokerName:   tx.BrokerName,
				AccountTag:   tx.AccountTag,
			}
			db.Dao.Create(&holding)
		} else {
			newTotalCost := holding.TotalCost + tx.Amount + tx.Fee
			newQuantity := holding.Quantity + tx.Quantity
			if newQuantity > 0 {
				holding.AvgCost = newTotalCost / newQuantity
			}
			holding.Quantity = newQuantity
			holding.TotalCost = newTotalCost
			if holding.BrokerName == "" && tx.BrokerName != "" {
				holding.BrokerName = tx.BrokerName
			}
			if holding.AccountTag == "" && tx.AccountTag != "" {
				holding.AccountTag = tx.AccountTag
			}
			db.Dao.Save(&holding)
		}
	case "sell":
		if result.Error == nil {
			holding.Quantity -= tx.Quantity
			if holding.Quantity < 0 {
				holding.Quantity = 0
			}
			holding.TotalCost = holding.AvgCost * holding.Quantity
			db.Dao.Save(&holding)
		}
	}

	tx.HoldingID = holding.ID
	if err := db.Dao.Create(&tx).Error; err != nil {
		logger.SugaredLogger.Errorf("add transaction failed: %v", err)
		return nil
	}

	switch holding.HoldingType {
	case "fund":
		s.refreshSingleFundHolding(&holding)
	case "stock":
		s.refreshSingleStockHolding(&holding)
	}
	db.Dao.Save(&holding)
	return &tx
}

func (s *Service) UpsertFundHoldingByAmount(input FundPositionInput) *Holding {
	input.StockCode = strings.TrimSpace(input.StockCode)
	input.StockName = strings.TrimSpace(input.StockName)
	input.BrokerName = strings.TrimSpace(input.BrokerName)
	input.AccountTag = strings.TrimSpace(input.AccountTag)
	input.Remark = strings.TrimSpace(input.Remark)

	if input.StockCode == "" || input.PositionAmount <= 0 {
		return nil
	}

	existing := s.GetHoldingByCode(input.StockCode)
	holding := Holding{
		StockCode:   input.StockCode,
		StockName:   input.StockName,
		HoldingType: "fund",
		BrokerName:  defaultLabel(input.BrokerName, "支付宝"),
		AccountTag:  defaultLabel(input.AccountTag, "主账户"),
		Remark:      input.Remark,
	}
	if existing != nil {
		holding = *existing
		if input.StockName != "" {
			holding.StockName = input.StockName
		}
		if input.BrokerName != "" {
			holding.BrokerName = input.BrokerName
		}
		if input.AccountTag != "" {
			holding.AccountTag = input.AccountTag
		}
		if input.Remark != "" {
			holding.Remark = input.Remark
		}
	}

	s.refreshSingleFundHolding(&holding)
	if holding.CurrentPrice <= 0 {
		return nil
	}

	holding.Quantity = input.PositionAmount / holding.CurrentPrice
	if holding.Quantity <= 0 {
		return nil
	}

	costAmount := input.CostAmount
	if costAmount <= 0 {
		costAmount = input.PositionAmount
	}

	holding.TotalCost = costAmount
	holding.AvgCost = costAmount / holding.Quantity
	holding.TotalValue = holding.CurrentPrice * holding.Quantity
	holding.ProfitLoss = holding.TotalValue - holding.TotalCost
	if holding.TotalCost > 0 {
		holding.ProfitRate = roundPercent(holding.ProfitLoss / holding.TotalCost * 100)
	}

	if existing == nil {
		now := time.Now()
		holding.BuyDate = &now
		if err := db.Dao.Create(&holding).Error; err != nil {
			logger.SugaredLogger.Errorf("upsert fund holding by amount failed: %v", err)
			return nil
		}
	} else {
		if err := db.Dao.Save(&holding).Error; err != nil {
			logger.SugaredLogger.Errorf("upsert fund holding by amount failed: %v", err)
			return nil
		}
	}

	return &holding
}

func (s *Service) GetTransactions(stockCode string, page, pageSize int) ([]Transaction, int64) {
	var transactions []Transaction
	var total int64

	query := db.Dao.Model(&Transaction{})
	if trimmed := strings.TrimSpace(stockCode); trimmed != "" {
		query = query.Where("stock_code = ?", trimmed)
	}

	query.Count(&total)
	query.Order("trade_date desc, id desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&transactions)

	return transactions, total
}

func (s *Service) UpdateHoldingPrice(stockCode string, currentPrice, todayChange, todayRate float64) {
	var holding Holding
	if err := db.Dao.Where("stock_code = ? AND quantity > 0", stockCode).First(&holding).Error; err != nil {
		return
	}

	holding.CurrentPrice = currentPrice
	holding.TotalValue = currentPrice * holding.Quantity
	holding.ProfitLoss = holding.TotalValue - holding.TotalCost
	if holding.TotalCost > 0 {
		holding.ProfitRate = roundPercent(holding.ProfitLoss / holding.TotalCost * 100)
	}
	holding.TodayChange = todayChange * holding.Quantity
	holding.TodayRate = todayRate
	db.Dao.Save(&holding)
}

func (s *Service) GetPortfolioSummary() *PortfolioSummary {
	holdings := s.GetAllHoldings()
	summary := &PortfolioSummary{Holdings: holdings}

	for _, holding := range holdings {
		summary.TotalCost += holding.TotalCost
		summary.TotalValue += holding.TotalValue
		summary.TotalProfit += holding.ProfitLoss
		summary.TodayProfit += holding.TodayChange

		if holding.HoldingType == "stock" {
			summary.StockCount++
			summary.StockValue += holding.TotalValue
			continue
		}

		summary.FundCount++
		summary.FundValue += holding.TotalValue
		switch classifyFundTypeByHolding(holding.StockCode).Category {
		case "bond":
			summary.BondFundCount++
			summary.BondFundValue += holding.TotalValue
		case "cash":
			summary.CashFundCount++
			summary.CashFundValue += holding.TotalValue
		case "equity":
			summary.EquityFundCount++
			summary.EquityFundValue += holding.TotalValue
		}
	}

	if summary.TotalCost > 0 {
		summary.TotalProfitRate = roundPercent(summary.TotalProfit / summary.TotalCost * 100)
	}
	return summary
}

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
		StockValue:   summary.StockValue,
		FundValue:    summary.FundValue,
	}

	var existing ProfitSnapshot
	if err := db.Dao.Where("snapshot_date = ?", today).First(&existing).Error; err != nil {
		db.Dao.Create(&snapshot)
		return
	}

	snapshot.ID = existing.ID
	db.Dao.Save(&snapshot)
}

func (s *Service) SaveAndReturnDailySnapshot() *ProfitSnapshot {
	s.SaveDailySnapshot()
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var snapshot ProfitSnapshot
	if err := db.Dao.Where("snapshot_date = ?", today).First(&snapshot).Error; err != nil {
		return nil
	}
	return &snapshot
}

func (s *Service) GetProfitHistory(days int) []ProfitSnapshot {
	var snapshots []ProfitSnapshot
	startDate := time.Now().AddDate(0, 0, -days)
	db.Dao.Where("snapshot_date >= ?", startDate).Order("snapshot_date asc").Find(&snapshots)
	return snapshots
}

func (s *Service) SyncPortfolioQuotes() *PortfolioSummary {
	holdings := s.GetAllHoldings()
	for i := range holdings {
		holding := holdings[i]
		switch holding.HoldingType {
		case "fund":
			s.refreshSingleFundHolding(&holding)
		default:
			s.refreshSingleStockHolding(&holding)
		}
		db.Dao.Save(&holding)
	}
	return s.GetPortfolioSummary()
}

func (s *Service) SyncHoldingFundEstimates() int {
	holdings := s.GetHoldingsByType("fund")
	updated := 0
	for i := range holdings {
		holding := holdings[i]
		s.refreshSingleFundHolding(&holding)
		db.Dao.Save(&holding)
		updated++
	}
	return updated
}

func (s *Service) SyncHoldingStockQuotes() int {
	holdings := s.GetHoldingsByType("stock")
	updated := 0
	for i := range holdings {
		holding := holdings[i]
		s.refreshSingleStockHolding(&holding)
		db.Dao.Save(&holding)
		updated++
	}
	return updated
}

func (s *Service) GetFundDashboard() *FundPortfolioDashboard {
	holdings := s.GetHoldingsByType("fund")
	positions := make([]FundHoldingView, 0, len(holdings))
	typeMap := map[string]*AllocationItem{}
	platformMap := map[string]*AllocationItem{}
	accountMap := map[string]*AllocationItem{}
	companyMap := map[string]*AllocationItem{}

	dashboard := &FundPortfolioDashboard{
		Summary: PortfolioSummary{},
	}

	for _, holding := range holdings {
		view := s.buildFundHoldingView(holding)
		positions = append(positions, view)

		dashboard.Summary.TotalCost += holding.TotalCost
		dashboard.Summary.TotalValue += holding.TotalValue
		dashboard.Summary.TotalProfit += holding.ProfitLoss
		dashboard.Summary.TodayProfit += holding.TodayChange
		dashboard.Summary.FundCount++
		dashboard.Summary.FundValue += holding.TotalValue

		switch view.Category {
		case "bond":
			dashboard.Summary.BondFundCount++
			dashboard.Summary.BondFundValue += holding.TotalValue
		case "cash":
			dashboard.Summary.CashFundCount++
			dashboard.Summary.CashFundValue += holding.TotalValue
		case "equity":
			dashboard.Summary.EquityFundCount++
			dashboard.Summary.EquityFundValue += holding.TotalValue
		}

		addAllocation(typeMap, view.CategoryLabel, holding.TotalValue)
		addAllocation(platformMap, defaultLabel(holding.BrokerName, "未标记平台"), holding.TotalValue)
		addAllocation(accountMap, defaultLabel(holding.AccountTag, "未分组账户"), holding.TotalValue)
		addAllocation(companyMap, defaultLabel(view.FundCompany, "未披露基金公司"), holding.TotalValue)

		if view.EstimateUpdated && view.NetEstimatedUnit != nil && view.NetUnitValue != nil {
			dashboard.EstimatedProfitToday += (*view.NetEstimatedUnit - *view.NetUnitValue) * holding.Quantity
		} else {
			dashboard.EstimatedProfitToday += holding.TodayChange
		}
	}

	if dashboard.Summary.TotalCost > 0 {
		dashboard.Summary.TotalProfitRate = roundPercent(dashboard.Summary.TotalProfit / dashboard.Summary.TotalCost * 100)
	}
	if dashboard.Summary.FundValue > 0 {
		dashboard.ConservativeRatio = roundPercent((dashboard.Summary.BondFundValue + dashboard.Summary.CashFundValue) / dashboard.Summary.FundValue * 100)
		dashboard.BondAllocationRatio = roundPercent(dashboard.Summary.BondFundValue / dashboard.Summary.FundValue * 100)
	}

	sort.Slice(positions, func(i, j int) bool {
		if positions[i].TotalValue == positions[j].TotalValue {
			return positions[i].StockCode < positions[j].StockCode
		}
		return positions[i].TotalValue > positions[j].TotalValue
	})

	dashboard.Positions = positions
	dashboard.Summary.Holdings = holdings
	dashboard.TypeAllocation = allocationSlice(typeMap, dashboard.Summary.FundValue)
	dashboard.PlatformAllocation = allocationSlice(platformMap, dashboard.Summary.FundValue)
	dashboard.AccountAllocation = allocationSlice(accountMap, dashboard.Summary.FundValue)
	dashboard.CompanyAllocation = allocationSlice(companyMap, dashboard.Summary.FundValue)
	return dashboard
}

func (s *Service) GetFundProfile(code string) *FundProfile {
	return s.getFundProfile(code, false)
}

func (s *Service) RefreshFundProfile(code string) *FundProfile {
	return s.getFundProfile(code, true)
}

func (s *Service) getFundProfile(code string, refresh bool) *FundProfile {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil
	}

	api := data.NewFundApi()
	holding := s.GetHoldingByCode(code)
	current := Holding{
		StockCode:   code,
		HoldingType: "fund",
	}
	if holding != nil {
		current = *holding
	}

	if refresh {
		s.ensureFundMetricsFreshToday(code)
		if holding != nil {
			s.refreshSingleFundHolding(&current)
			db.Dao.Save(&current)
		} else {
			ensureFollowedFund(code, current.StockName)
			if basic, err := dbQueryFundBasic(code); err != nil || fundBasicNeedsRefresh(basic) {
				api.CrawlFundBasic(code)
			}
			api.CrawlFundNetUnitValue(code)
			api.CrawlFundNetEstimatedUnit(code)
		}
	}

	view := s.buildFundHoldingView(current)

	points := make([]FundTrendPoint, 0)
	updatedAt := strings.TrimSpace(view.LatestDailyUpdatedAt)
	var latestReturn *float64
	if view.LatestDailyRate != nil {
		value := *view.LatestDailyRate
		latestReturn = &value
	}

	if refresh {
		trend, trendUpdatedAt, trendLatestReturn, err := api.GetFundTrend(code)
		if err != nil {
			logger.SugaredLogger.Warnf("get fund trend failed for %s: %v", code, err)
		} else {
			updatedAt = trendUpdatedAt
			latestReturn = trendLatestReturn
			points = make([]FundTrendPoint, 0, len(trend))
			for _, point := range trend {
				points = append(points, FundTrendPoint{
					Timestamp:   point.Timestamp,
					Date:        point.Date,
					Value:       point.Value,
					DailyReturn: point.DailyReturn,
				})
			}
		}
	}

	estimateTrendRaw, estimateUpdatedAt, estimateLatestRate, estimateErr := api.GetFundEstimatedTrend(code, time.Now())
	if estimateErr != nil {
		logger.SugaredLogger.Warnf("get fund estimate trend failed for %s: %v", code, estimateErr)
	}
	estimatePoints := make([]FundEstimatePoint, 0, len(estimateTrendRaw))
	for _, point := range estimateTrendRaw {
		estimatePoints = append(estimatePoints, FundEstimatePoint{
			Timestamp:     point.Timestamp,
			Time:          point.Time,
			EstimatedUnit: point.EstimatedUnit,
			EstimatedRate: point.EstimatedRate,
		})
	}

	stageRankings := make([]FundStageRanking, 0)
	if refresh {
		stageRankingsRaw, rankErr := api.GetFundStageRankings(code)
		if rankErr != nil {
			logger.SugaredLogger.Warnf("get fund stage rankings failed for %s: %v", code, rankErr)
		} else {
			stageRankings = make([]FundStageRanking, 0, len(stageRankingsRaw))
			for _, item := range stageRankingsRaw {
				stageRankings = append(stageRankings, FundStageRanking{
					Period:             item.Period,
					ReturnRate:         item.ReturnRate,
					SimilarAverageRate: item.SimilarAverageRate,
					BenchmarkLabel:     item.BenchmarkLabel,
					BenchmarkRate:      item.BenchmarkRate,
					Rank:               item.Rank,
					RankTotal:          item.RankTotal,
					RankPercentile:     item.RankPercentile,
					RankDelta:          item.RankDelta,
					RankDeltaDirection: item.RankDeltaDirection,
					Quartile:           item.Quartile,
				})
			}
		}
	}

	return &FundProfile{
		FundHoldingView:        view,
		Trend:                  points,
		TrendUpdatedAt:         updatedAt,
		LatestReturn:           latestReturn,
		EstimateTrend:          estimatePoints,
		EstimateTrendUpdatedAt: estimateUpdatedAt,
		EstimateLatestRate:     estimateLatestRate,
		StageRankings:          stageRankings,
		StageRankingsUpdatedAt: view.NetUnitValueDate,
	}
}

func (s *Service) EnsureFundUniverse() int64 {
	var total int64
	db.Dao.Model(&data.FundBasic{}).Count(&total)
	if total >= 3000 {
		return total
	}

	data.NewFundApi().AllFund()
	db.Dao.Model(&data.FundBasic{}).Count(&total)
	return total
}

func resolveFundRefreshScope(limit int) (string, int) {
	if limit < 0 {
		return fundRefreshScopeAll, 0
	}
	if limit == 0 {
		return fundRefreshScopeFocused, 0
	}
	return fundRefreshScopeFocused, limit
}

func defaultFundRefreshScope() string {
	return fundRefreshScopeFocused
}

func (s *Service) ensureFundScreenerFreshAsync(limit int) FundRefreshStatus {
	return s.startFundScreenerRefresh(limit, false)
}

func (s *Service) startFundScreenerRefresh(limit int, force bool) FundRefreshStatus {
	universeCount := s.EnsureFundUniverse()
	today := time.Now().Format("2006-01-02")
	scope, batchLimit := resolveFundRefreshScope(limit)
	scopeStatus := s.loadFundRefreshScopeSnapshot(scope, today, universeCount)
	status := s.buildFundRefreshStatus(universeCount)
	if status.Refreshing {
		return status
	}
	if scopeStatus.PendingCount == 0 {
		status.Scope = scope
		status.TargetCount = scopeStatus.TargetCount
		status.TargetUpdated = scopeStatus.UpdatedToday
		status.TargetPending = scopeStatus.PendingCount
		status.State, status.StateLabel = resolveFundRefreshState(scopeStatus.UpdatedToday, scopeStatus.TargetCount, scopeStatus.PendingCount)
		status.NeedsRefresh = false
		if force {
			switch scope {
			case fundRefreshScopeAll:
				status.Message = "Today's full-universe fund metrics are already complete."
			default:
				status.Message = "Today's watchlist-related fund metrics are already complete."
			}
		}
		return status
	}

	fundRefreshState.mu.Lock()
	if fundRefreshState.refreshing {
		fundRefreshState.mu.Unlock()
		return s.buildFundRefreshStatus(universeCount)
	}
	fundRefreshState.refreshing = true
	fundRefreshState.scope = scope
	fundRefreshState.progressNow = 0
	fundRefreshState.progressTotal = 0
	fundRefreshState.currentCode = ""
	fundRefreshState.lastStarted = time.Now()
	fundRefreshState.mu.Unlock()

	go s.runFundScreenerRefresh(batchLimit, scope)

	status = s.buildFundRefreshStatus(universeCount)
	status.Triggered = true
	switch {
	case scope == fundRefreshScopeAll && force:
		status.Message = "Manual full refresh started in the background. Remaining funds across the whole universe will be caught up today."
	case scope == fundRefreshScopeAll:
		status.Message = "A background full refresh has resumed for the whole fund universe."
	case batchLimit > 0:
		status.Message = "Manual watchlist-related refresh started in the background. This batch will update your followed and holding funds first."
	case force:
		status.Message = "Manual watchlist-related refresh started in the background. Followed funds, holdings, and recommendation candidates will be caught up today."
	case status.State == fundRefreshStatePartial:
		status.Message = "Today's watchlist-related metrics are only partially updated. A background refresh has resumed for followed funds, holdings, and recommendation candidates."
	default:
		status.Message = "Today's first request started a background watchlist-related refresh. Followed funds, holdings, and recommendation candidates will improve first."
	}
	return status
}

func (s *Service) runFundScreenerRefresh(limit int, scope string) {
	var queuedCount int
	var successCount int
	today := time.Now().Format("2006-01-02")
	defer func() {
		if r := recover(); r != nil {
			logger.SugaredLogger.Errorf("fund screener refresh panic: %v", r)
			logger.SugaredLogger.Errorf("fund screener refresh stack: %s", string(debug.Stack()))
		}
		fundRefreshState.mu.Lock()
		fundRefreshState.refreshing = false
		fundRefreshState.scope = ""
		fundRefreshState.currentCode = ""
		fundRefreshState.lastFinished = time.Now()
		fundRefreshState.mu.Unlock()
	}()

	var basics []data.FundBasic
	query := s.buildFundRefreshQuery(scope, today).
		Order("CASE WHEN screen_updated_at IS NULL OR screen_updated_at = '' THEN 0 ELSE 1 END asc").
		Order("screen_updated_at asc").
		Order("updated_at asc")
	if limit > 0 {
		query = query.Limit(limit)
	}
	query.Find(&basics)
	queuedCount = len(basics)

	fundRefreshState.mu.Lock()
	fundRefreshState.progressTotal = int64(len(basics))
	fundRefreshState.mu.Unlock()

	logger.SugaredLogger.Infof("fund screener refresh started: scope=%s limit=%d queued=%d", scope, limit, queuedCount)

	for idx, basic := range basics {
		fundRefreshState.mu.Lock()
		fundRefreshState.currentCode = basic.Code
		fundRefreshState.mu.Unlock()

		refreshed := false
		func(code string) {
			defer func() {
				if r := recover(); r != nil {
					logger.SugaredLogger.Errorf("fund screener refresh panic for %s: %v", code, r)
					logger.SugaredLogger.Errorf("fund screener refresh stack for %s: %s", code, string(debug.Stack()))
				}
			}()
			refreshed = s.refreshFundScreeningMetrics(code)
		}(basic.Code)
		if refreshed {
			successCount++
		}

		fundRefreshState.mu.Lock()
		fundRefreshState.progressNow = int64(idx + 1)
		fundRefreshState.mu.Unlock()
	}

	fundRefreshState.mu.Lock()
	fundRefreshState.progressNow = int64(len(basics))
	fundRefreshState.mu.Unlock()

	logger.SugaredLogger.Infof("fund screener refresh finished: scope=%s limit=%d queued=%d refreshed=%d", scope, limit, queuedCount, successCount)
}

func resolveFundRefreshState(updatedToday, targetCount, pendingCount int64) (string, string) {
	switch {
	case targetCount <= 0:
		return fundRefreshStateCompleted, "今日完成"
	case updatedToday <= 0:
		return fundRefreshStateNotStarted, "未开始"
	case pendingCount > 0:
		return fundRefreshStatePartial, "部分更新"
	default:
		return fundRefreshStateCompleted, "今日完成"
	}
}

func (s *Service) buildFundRefreshStatus(universeCount int64) FundRefreshStatus {
	today := time.Now().Format("2006-01-02")
	lastRefreshHint := loadFundScreenRefreshHint()
	screenedCount := loadScreenedFundCount()
	updatedToday := loadUpdatedTodayFundCount(today)

	fundRefreshState.mu.Lock()
	refreshing := fundRefreshState.refreshing
	scope := fundRefreshState.scope
	progressNow := fundRefreshState.progressNow
	progressTotal := fundRefreshState.progressTotal
	currentCode := fundRefreshState.currentCode
	fundRefreshState.mu.Unlock()

	if strings.TrimSpace(scope) == "" {
		scope = defaultFundRefreshScope()
	}

	scopeStatus := s.loadFundRefreshScopeSnapshot(scope, today, universeCount)
	state, stateLabel := resolveFundRefreshState(scopeStatus.UpdatedToday, scopeStatus.TargetCount, scopeStatus.PendingCount)
	needsRefresh := scopeStatus.PendingCount > 0

	message := ""
	switch {
	case refreshing:
		if scope == fundRefreshScopeAll {
			message = "A full-universe fund refresh is running in the background."
		} else {
			message = "A watchlist-related fund refresh is running in the background."
		}
	case scopeStatus.TargetCount == 0:
		message = "There are no watchlist or holding funds yet, so daily focused refresh is idle."
	case state == fundRefreshStateNotStarted:
		message = "Today's watchlist-related fund metrics have not started refreshing yet."
	case state == fundRefreshStatePartial:
		message = "Today's watchlist-related fund metrics are only partially updated."
	default:
		if scope == fundRefreshScopeAll {
			message = "Today's full-universe fund metrics are ready."
		} else {
			message = "Today's watchlist-related fund metrics are ready."
		}
	}

	return FundRefreshStatus{
		State:           state,
		StateLabel:      stateLabel,
		Scope:           scope,
		Refreshing:      refreshing,
		NeedsRefresh:    needsRefresh,
		CurrentDate:     today,
		LastRefreshHint: lastRefreshHint,
		UpdatedToday:    updatedToday,
		ScreenedCount:   screenedCount,
		UniverseCount:   universeCount,
		TargetCount:     scopeStatus.TargetCount,
		TargetUpdated:   scopeStatus.UpdatedToday,
		TargetPending:   scopeStatus.PendingCount,
		ProgressCurrent: progressNow,
		ProgressTotal:   progressTotal,
		CurrentCode:     currentCode,
		Message:         message,
	}
}

func (s *Service) ensureFundMetricsFreshToday(code string) {
	code = strings.TrimSpace(code)
	if code == "" {
		return
	}

	var basic data.FundBasic
	if err := db.Dao.Where("code = ?", code).First(&basic).Error; err != nil {
		s.refreshFundScreeningMetrics(code)
		return
	}
	if !isSameTradingDayString(basic.ScreenUpdatedAt, time.Now()) {
		s.refreshFundScreeningMetrics(code)
	}
}

func (s *Service) RefreshFundScreenerData(limit int) map[string]any {
	universeCount := s.EnsureFundUniverse()
	status := s.startFundScreenerRefresh(limit, true)
	scope, _ := resolveFundRefreshScope(limit)

	return map[string]any{
		"universeCount": universeCount,
		"started":       status.Triggered,
		"scope":         scope,
		"screenedCount": loadScreenedFundCount(),
		"updatedToday":  loadUpdatedTodayFundCount(time.Now().Format("2006-01-02")),
		"refreshStatus": status,
	}
}

func buildPendingFundRefreshQuery(today string) *gorm.DB {
	return db.Dao.Where(
		"screen_updated_at IS NULL OR screen_updated_at = '' OR substr(screen_updated_at, 1, 10) <> ?",
		strings.TrimSpace(today),
	)
}

func loadPendingFundRefreshCount(today string) int64 {
	var count int64
	buildPendingFundRefreshQuery(today).Model(&data.FundBasic{}).Count(&count)
	return count
}

func (s *Service) buildFundRefreshQuery(scope string, today string) *gorm.DB {
	switch scope {
	case fundRefreshScopeAll:
		return buildPendingFundRefreshQuery(today)
	default:
		codes := s.loadFocusedFundRefreshCodes()
		if len(codes) == 0 {
			return db.Dao.Where("1 = 0")
		}
		s.ensureFundBasicsForCodes(codes)
		return buildPendingFundRefreshQuery(today).Where("code IN ?", codes)
	}
}

func (s *Service) loadFundRefreshScopeSnapshot(scope string, today string, universeCount int64) fundRefreshScopeSnapshot {
	switch scope {
	case fundRefreshScopeAll:
		return fundRefreshScopeSnapshot{
			Scope:        scope,
			TargetCount:  universeCount,
			UpdatedToday: loadUpdatedTodayFundCount(today),
			PendingCount: loadPendingFundRefreshCount(today),
		}
	default:
		codes := s.loadFocusedFundRefreshCodes()
		return fundRefreshScopeSnapshot{
			Scope:        fundRefreshScopeFocused,
			TargetCount:  int64(len(codes)),
			UpdatedToday: loadUpdatedTodayFundCountByCodes(codes, today),
			PendingCount: loadPendingFundRefreshCountByCodes(codes, today),
		}
	}
}

func loadUpdatedTodayFundCountByCodes(codes []string, today string) int64 {
	if len(codes) == 0 {
		return 0
	}
	var count int64
	db.Dao.Model(&data.FundBasic{}).
		Where("code IN ?", codes).
		Where("substr(screen_updated_at, 1, 10) = ?", strings.TrimSpace(today)).
		Count(&count)
	return count
}

func loadPendingFundRefreshCountByCodes(codes []string, today string) int64 {
	if len(codes) == 0 {
		return 0
	}
	var count int64
	buildPendingFundRefreshQuery(today).
		Model(&data.FundBasic{}).
		Where("code IN ?", codes).
		Count(&count)
	return count
}

func (s *Service) loadFocusedFundRefreshCodes() []string {
	codeSet := make(map[string]struct{})
	for _, code := range loadFundWatchlistCodes() {
		if trimmed := strings.TrimSpace(code); trimmed != "" {
			codeSet[trimmed] = struct{}{}
		}
	}
	holdingCodes := loadFundHoldingCodes()
	for _, code := range holdingCodes {
		if trimmed := strings.TrimSpace(code); trimmed != "" {
			codeSet[trimmed] = struct{}{}
		}
	}

	seedCodes := sortedCodeKeys(codeSet)
	if len(seedCodes) == 0 {
		return seedCodes
	}

	s.ensureFundBasicsForCodes(seedCodes)
	for _, code := range s.collectRecommendedFundCodes(loadFundWatchlistCodes(), 3) {
		if trimmed := strings.TrimSpace(code); trimmed != "" {
			codeSet[trimmed] = struct{}{}
		}
	}

	return sortedCodeKeys(codeSet)
}

func loadFundWatchlistCodes() []string {
	var codes []string
	db.Dao.Model(&data.FollowedFund{}).
		Where("is_watchlist = ?", true).
		Pluck("code", &codes)
	return dedupeCodes(codes)
}

func loadFundHoldingCodes() []string {
	var codes []string
	db.Dao.Model(&Holding{}).
		Where("holding_type = ? AND quantity > 0", "fund").
		Pluck("stock_code", &codes)
	return dedupeCodes(codes)
}

func (s *Service) ensureFundBasicsForCodes(codes []string) {
	if len(codes) == 0 {
		return
	}

	var existing []string
	db.Dao.Model(&data.FundBasic{}).Where("code IN ?", codes).Pluck("code", &existing)
	existingSet := make(map[string]struct{}, len(existing))
	for _, code := range existing {
		existingSet[strings.TrimSpace(code)] = struct{}{}
	}

	api := data.NewFundApi()
	for _, code := range codes {
		trimmed := strings.TrimSpace(code)
		if trimmed == "" {
			continue
		}
		if _, ok := existingSet[trimmed]; ok {
			continue
		}
		api.CrawlFundBasic(trimmed)
	}
}

func (s *Service) collectRecommendedFundCodes(referenceCodes []string, perDimension int) []string {
	if len(referenceCodes) == 0 || perDimension <= 0 {
		return []string{}
	}

	codeSet := make(map[string]struct{})
	dimensions := []string{"balanced", "higher_return", "lower_drawdown"}
	for _, code := range dedupeCodes(referenceCodes) {
		reference, err := dbQueryFundBasic(code)
		if err != nil || reference == nil {
			continue
		}
		refCategory := classifyFundType(defaultLabel(reference.Type, reference.Name))
		universe := loadBetterFundUniverse(*reference, refCategory.Category, true)
		if len(universe.Basics) == 0 {
			continue
		}
		for _, dimension := range dimensions {
			candidates := buildBetterFundCandidates(*reference, universe, dimension)
			sort.Slice(candidates, func(i, j int) bool {
				if candidates[i].BetterScore == candidates[j].BetterScore {
					return candidates[i].Code < candidates[j].Code
				}
				return candidates[i].BetterScore > candidates[j].BetterScore
			})
			for i := 0; i < len(candidates) && i < perDimension; i++ {
				if strings.TrimSpace(candidates[i].Code) == "" || candidates[i].BetterScore <= 0 {
					continue
				}
				codeSet[candidates[i].Code] = struct{}{}
			}
		}
	}
	return sortedCodeKeys(codeSet)
}

func dedupeCodes(codes []string) []string {
	codeSet := make(map[string]struct{})
	for _, code := range codes {
		if trimmed := strings.TrimSpace(code); trimmed != "" {
			codeSet[trimmed] = struct{}{}
		}
	}
	return sortedCodeKeys(codeSet)
}

func sortedCodeKeys(codeSet map[string]struct{}) []string {
	if len(codeSet) == 0 {
		return []string{}
	}
	codes := make([]string, 0, len(codeSet))
	for code := range codeSet {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	return codes
}

func (s *Service) GetFundScreener(query FundScreenerQuery) *FundScreenerResult {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	universeCount := s.EnsureFundUniverse()
	refreshStatus := s.ensureFundScreenerFreshAsync(0)

	dbQuery := db.Dao.Model(&data.FundBasic{})
	keyword := strings.TrimSpace(query.Keyword)
	if keyword != "" {
		pattern := "%" + keyword + "%"
		dbQuery = dbQuery.Where(
			"code LIKE ? OR name LIKE ? OR company LIKE ? OR manager LIKE ? OR top_industry LIKE ?",
			pattern, pattern, pattern, pattern, pattern,
		)
	}
	if trimmed := strings.TrimSpace(query.FundType); trimmed != "" {
		dbQuery = dbQuery.Where("type = ?", trimmed)
	}
	if trimmed := strings.TrimSpace(query.Industry); trimmed != "" {
		dbQuery = dbQuery.Where("top_industry LIKE ?", "%"+trimmed+"%")
	}
	dbQuery = applyFundCategoryDBFilter(dbQuery, query.Category)
	if query.MinReturn7 != nil {
		dbQuery = dbQuery.Where("net_growth7 >= ?", *query.MinReturn7)
	}
	if query.MinReturn1 != nil {
		dbQuery = dbQuery.Where("net_growth1 >= ?", *query.MinReturn1)
	}
	if query.MinReturn3 != nil {
		dbQuery = dbQuery.Where("net_growth3 >= ?", *query.MinReturn3)
	}
	if query.MaxDrawdown12 != nil {
		dbQuery = dbQuery.Where("max_drawdown12 <= ?", *query.MaxDrawdown12)
	}
	if query.OnlyWatchlist {
		dbQuery = dbQuery.Where(
			"code IN (?)",
			db.Dao.Model(&data.FollowedFund{}).Select("code").Where("is_watchlist = ?", true),
		)
	}

	var total int64
	dbQuery.Count(&total)

	sortColumn := mapFundScreenerSortColumn(query.SortBy)
	sortOrder := "desc"
	if strings.EqualFold(strings.TrimSpace(query.SortOrder), "asc") {
		sortOrder = "asc"
	}

	var basics []data.FundBasic
	dbQuery.
		Order(sortColumn + " " + sortOrder).
		Order("updated_at desc").
		Order("code asc").
		Offset((query.Page - 1) * query.PageSize).
		Limit(query.PageSize).
		Find(&basics)

	watchlistMap := loadFundWatchlistMap(basics)
	items := make([]FundScreenerItem, 0, len(basics))
	for _, basic := range basics {
		items = append(items, buildFundScreenerItem(basic, watchlistMap[basic.Code]))
	}

	var screenedCount int64
	db.Dao.Model(&data.FundBasic{}).
		Where("screen_updated_at IS NOT NULL AND screen_updated_at <> ''").
		Count(&screenedCount)

	return &FundScreenerResult{
		Items:           items,
		Total:           total,
		Page:            query.Page,
		PageSize:        query.PageSize,
		UniverseCount:   universeCount,
		ScreenedCount:   screenedCount,
		TypeOptions:     loadFundTypeOptions(),
		CategoryOptions: []string{"bond", "cash", "equity", "other"},
		IndustryOptions: loadFundIndustryOptions(),
		LastRefreshHint: loadFundScreenRefreshHint(),
		RefreshStatus:   refreshStatus,
	}
}

func (s *Service) GetBetterFunds(query BetterFundQuery) *BetterFundResult {
	code := strings.TrimSpace(query.ReferenceCode)
	if code == "" {
		return nil
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}
	if query.PageSize > 50 {
		query.PageSize = 50
	}
	query.Dimension = normalizeBetterFundDimension(query.Dimension)

	var reference data.FundBasic
	if err := db.Dao.Where("code = ?", code).First(&reference).Error; err != nil {
		return nil
	}
	if !isSameTradingDayString(reference.ScreenUpdatedAt, time.Now()) {
		s.refreshFundScreeningMetrics(code)
		if refreshed, err := dbQueryFundBasic(code); err == nil && refreshed != nil {
			reference = *refreshed
		}
	}

	refItem := buildFundScreenerItem(reference, isFundWatchlisted(code))
	refCategory := classifyFundType(defaultLabel(reference.Type, reference.Name))
	refreshStatus := s.ensureFundScreenerFreshAsync(0)
	universe := loadBetterFundUniverse(reference, refCategory.Category, query.SameTypeOnly)
	candidates := buildBetterFundCandidates(reference, universe, query.Dimension)
	if false && query.SameTypeOnly && len(candidates) == 0 {
		candidates = loadBetterFundCandidates(reference, refCategory.Category, query.Dimension, false)
		for i := range candidates {
			candidates[i].Reasons = append([]string{"同类型暂无更优，已放宽到同大类继续筛选"}, candidates[i].Reasons...)
			candidates[i].BetterScore = roundPercent(candidates[i].BetterScore * 0.92)
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].BetterScore == candidates[j].BetterScore {
			return candidates[i].Code < candidates[j].Code
		}
		return candidates[i].BetterScore > candidates[j].BetterScore
	})
	for i := range candidates {
		candidates[i].RecommendationRank = i + 1
		if universe.FallbackApplied {
			candidates[i].Reasons = append([]string{"同类型暂无更优，已放宽到同大类继续筛选"}, candidates[i].Reasons...)
			candidates[i].ReasonSummary = strings.Join(topStrings(candidates[i].Reasons, 3), " / ")
			candidates[i].BetterScore = roundPercent(candidates[i].BetterScore * 0.92)
		}
	}

	total := int64(len(candidates))
	start := (query.Page - 1) * query.PageSize
	if start > len(candidates) {
		start = len(candidates)
	}
	end := start + query.PageSize
	if end > len(candidates) {
		end = len(candidates)
	}

	return &BetterFundResult{
		Reference:        refItem,
		Candidates:       candidates[start:end],
		Dimension:        query.Dimension,
		SortLabel:        betterSortLabel(query.Dimension),
		ScopeLabel:       universe.ScopeLabel,
		ComparedUniverse: universe.ComparedUniverse,
		FallbackApplied:  universe.FallbackApplied,
		DataHint:         betterDataHint(refreshStatus, universe),
		Total:            total,
		Page:             query.Page,
		PageSize:         query.PageSize,
		RefreshStatus:    refreshStatus,
	}
}

func (s *Service) CompareFunds(query FundCompareQuery) *FundCompareResult {
	seen := make(map[string]struct{})
	codes := make([]string, 0, len(query.Codes))
	for _, raw := range query.Codes {
		code := strings.TrimSpace(raw)
		if code == "" {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		codes = append(codes, code)
		if len(codes) >= 10 {
			break
		}
	}
	if len(codes) == 0 {
		return &FundCompareResult{}
	}

	now := time.Now()
	api := data.NewFundApi()
	for _, code := range codes {
		basic, err := dbQueryFundBasic(code)
		if err != nil {
			api.CrawlFundBasic(code)
			basic, err = dbQueryFundBasic(code)
		}
		if err == nil && !isSameTradingDayString(basic.ScreenUpdatedAt, now) {
			s.refreshFundScreeningMetrics(code)
		}
	}

	var basics []data.FundBasic
	db.Dao.Where("code IN ?", codes).Find(&basics)
	basicMap := make(map[string]data.FundBasic, len(basics))
	for _, basic := range basics {
		basicMap[basic.Code] = basic
	}

	watchlistMap := loadFundWatchlistMap(basics)
	items := make([]FundScreenerItem, 0, len(codes))
	missing := make([]string, 0)
	for _, code := range codes {
		basic, ok := basicMap[code]
		if !ok {
			missing = append(missing, code)
			continue
		}
		items = append(items, buildFundScreenerItem(basic, watchlistMap[code]))
	}

	return &FundCompareResult{
		Items:        items,
		Total:        len(items),
		MissingCodes: missing,
		RefreshedAt:  now.Format("2006-01-02 15:04:05"),
	}
}

func loadBetterFundCandidates(reference data.FundBasic, refCategory string, dimension string, sameTypeOnly bool) []BetterFundCandidate {
	dbQuery := db.Dao.Model(&data.FundBasic{}).
		Where("code <> ?", reference.Code).
		Where("screen_updated_at IS NOT NULL AND screen_updated_at <> ''")
	if sameTypeOnly && strings.TrimSpace(reference.Type) != "" {
		dbQuery = dbQuery.Where("type = ?", reference.Type)
	} else {
		dbQuery = applyFundCategoryDBFilter(dbQuery, refCategory)
	}

	var basics []data.FundBasic
	dbQuery.Order("updated_at desc").Order("code asc").Find(&basics)

	watchlistMap := loadFundWatchlistMap(basics)
	candidates := make([]BetterFundCandidate, 0, len(basics))
	for _, item := range basics {
		candidate, ok := buildBetterFundCandidateByDimension(reference, item, watchlistMap[item.Code], dimension)
		if ok {
			candidates = append(candidates, candidate)
		}
	}
	return candidates
}

func loadBetterFundUniverse(reference data.FundBasic, refCategory string, sameTypeOnly bool) betterCandidateUniverse {
	if sameTypeOnly && strings.TrimSpace(reference.Type) != "" {
		basics := queryBetterFundBasics(reference, refCategory, true)
		if len(basics) > 0 {
			return betterCandidateUniverse{
				Basics:           basics,
				ScopeLabel:       "同类型精确匹配：" + strings.TrimSpace(reference.Type),
				ComparedUniverse: len(basics) + 1,
			}
		}
		basics = queryBetterFundBasics(reference, refCategory, false)
		return betterCandidateUniverse{
			Basics:           basics,
			ScopeLabel:       "同类型暂无更优，已放宽到同大类：" + betterCategoryLabel(refCategory),
			FallbackApplied:  true,
			ComparedUniverse: len(basics) + 1,
		}
	}

	basics := queryBetterFundBasics(reference, refCategory, false)
	return betterCandidateUniverse{
		Basics:           basics,
		ScopeLabel:       "同大类匹配：" + betterCategoryLabel(refCategory),
		ComparedUniverse: len(basics) + 1,
	}
}

func queryBetterFundBasics(reference data.FundBasic, refCategory string, exactType bool) []data.FundBasic {
	dbQuery := db.Dao.Model(&data.FundBasic{}).
		Where("code <> ?", reference.Code).
		Where("screen_updated_at IS NOT NULL AND screen_updated_at <> ''")
	if exactType && strings.TrimSpace(reference.Type) != "" {
		dbQuery = dbQuery.Where("type = ?", reference.Type)
	} else {
		dbQuery = applyFundCategoryDBFilter(dbQuery, refCategory)
	}

	var basics []data.FundBasic
	dbQuery.Order("updated_at desc").Order("code asc").Find(&basics)
	return basics
}

func buildBetterFundCandidates(reference data.FundBasic, universe betterCandidateUniverse, dimension string) []BetterFundCandidate {
	if len(universe.Basics) == 0 {
		return []BetterFundCandidate{}
	}

	specs := betterMetricSpecsForDimension(dimension)
	rankUniverse := make([]data.FundBasic, 0, len(universe.Basics)+1)
	rankUniverse = append(rankUniverse, reference)
	rankUniverse = append(rankUniverse, universe.Basics...)

	rankMaps := make(map[string]map[string]betterMetricRank, len(specs))
	for _, spec := range specs {
		rankMaps[spec.Key] = buildBetterMetricRanks(rankUniverse, spec)
	}

	watchlistMap := loadFundWatchlistMap(universe.Basics)
	candidates := make([]BetterFundCandidate, 0, len(universe.Basics))
	for _, basic := range universe.Basics {
		candidate, ok := buildBetterFundCandidateV2(reference, basic, watchlistMap[basic.Code], dimension, specs, rankMaps, universe)
		if ok {
			candidates = append(candidates, candidate)
		}
	}
	return candidates
}

func betterMetricSpecsForDimension(dimension string) []betterMetricSpec {
	specs := []betterMetricSpec{
		{Key: "growth3", Label: "近3月", Better: "higher", Weight: 0.75, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.NetGrowth3 }},
		{Key: "growth6", Label: "近6月", Better: "higher", Weight: 0.95, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.NetGrowth6 }},
		{Key: "growth12", Label: "近1年", Better: "higher", Weight: 1.05, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.NetGrowth12 }},
		{Key: "drawdown12", Label: "近1年最大回撤", Better: "lower", Weight: 1.10, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.MaxDrawdown12 }},
		{Key: "sharpe12", Label: "近1年夏普", Better: "higher", Weight: 0.85, Format: "ratio", ValueOf: func(item data.FundBasic) *float64 { return item.Sharpe12 }},
		{Key: "calmar12", Label: "Calmar", Better: "higher", Weight: 0.70, Format: "ratio", ValueOf: func(item data.FundBasic) *float64 { return item.Calmar12 }},
		{Key: "volatility12", Label: "近1年波动", Better: "lower", Weight: 0.40, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.Volatility12 }},
	}

	switch normalizeBetterFundDimension(dimension) {
	case "lower_drawdown":
		return []betterMetricSpec{
			{Key: "drawdown12", Label: "近1年最大回撤", Better: "lower", Weight: 1.80, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.MaxDrawdown12 }},
			{Key: "volatility12", Label: "近1年波动", Better: "lower", Weight: 0.95, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.Volatility12 }},
			{Key: "sharpe12", Label: "近1年夏普", Better: "higher", Weight: 0.60, Format: "ratio", ValueOf: func(item data.FundBasic) *float64 { return item.Sharpe12 }},
			{Key: "growth3", Label: "近3月", Better: "higher", Weight: 0.35, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.NetGrowth3 }},
			{Key: "growth6", Label: "近6月", Better: "higher", Weight: 0.45, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.NetGrowth6 }},
			{Key: "growth12", Label: "近1年", Better: "higher", Weight: 0.55, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.NetGrowth12 }},
		}
	case "higher_return":
		return []betterMetricSpec{
			{Key: "growth7", Label: "近7天", Better: "higher", Weight: 0.25, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.NetGrowth7 }},
			{Key: "growth1", Label: "近1月", Better: "higher", Weight: 0.55, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.NetGrowth1 }},
			{Key: "growth3", Label: "近3月", Better: "higher", Weight: 0.95, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.NetGrowth3 }},
			{Key: "growth6", Label: "近6月", Better: "higher", Weight: 1.25, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.NetGrowth6 }},
			{Key: "growth12", Label: "近1年", Better: "higher", Weight: 1.15, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.NetGrowth12 }},
			{Key: "sharpe12", Label: "近1年夏普", Better: "higher", Weight: 0.55, Format: "ratio", ValueOf: func(item data.FundBasic) *float64 { return item.Sharpe12 }},
			{Key: "drawdown12", Label: "近1年最大回撤", Better: "lower", Weight: 0.25, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.MaxDrawdown12 }},
		}
	default:
		return specs
	}
}

func buildBetterFundCandidateV2(reference data.FundBasic, candidate data.FundBasic, watchlist bool, dimension string, specs []betterMetricSpec, rankMaps map[string]map[string]betterMetricRank, universe betterCandidateUniverse) (BetterFundCandidate, bool) {
	item := buildFundScreenerItem(candidate, watchlist)
	metrics := make([]BetterFundMetric, 0, len(specs))
	type scoredReason struct {
		text  string
		score float64
	}
	scoredReasons := make([]scoredReason, 0, len(specs))
	score := 0.0
	hasReturnAdvantage := false
	hasRiskAdvantage := false

	for _, spec := range specs {
		metric := buildBetterMetric(reference, candidate, spec, rankMaps[spec.Key])
		if metric.Contribution > 0 {
			score += metric.Contribution
			scoredReasons = append(scoredReasons, scoredReason{
				text:  describeBetterMetric(metric, spec.Format),
				score: metric.Contribution,
			})
			if isReturnBetterMetric(spec.Key) {
				hasReturnAdvantage = true
			}
			if isRiskBetterMetric(spec.Key) {
				hasRiskAdvantage = true
			}
		}
		metrics = append(metrics, metric)
	}

	switch normalizeBetterFundDimension(dimension) {
	case "lower_drawdown":
		if !hasRiskAdvantage {
			return BetterFundCandidate{}, false
		}
	case "higher_return":
		if !hasReturnAdvantage {
			return BetterFundCandidate{}, false
		}
	default:
		if !hasReturnAdvantage && !hasRiskAdvantage {
			return BetterFundCandidate{}, false
		}
	}

	if score <= 0 {
		return BetterFundCandidate{}, false
	}

	sort.Slice(scoredReasons, func(i, j int) bool {
		if scoredReasons[i].score == scoredReasons[j].score {
			return scoredReasons[i].text < scoredReasons[j].text
		}
		return scoredReasons[i].score > scoredReasons[j].score
	})

	reasons := make([]string, 0, len(scoredReasons))
	for _, item := range scoredReasons {
		reasons = append(reasons, item.text)
	}

	return BetterFundCandidate{
		FundScreenerItem: item,
		BetterScore:      roundPercent(score),
		ReasonSummary:    strings.Join(topStrings(reasons, 3), " / "),
		ScopeLabel:       universe.ScopeLabel,
		ComparedUniverse: universe.ComparedUniverse,
		Reasons:          reasons,
		Metrics:          metrics,
	}, true
}

func buildBetterMetric(reference data.FundBasic, candidate data.FundBasic, spec betterMetricSpec, rankMap map[string]betterMetricRank) BetterFundMetric {
	candidateValue := cloneFloat(spec.ValueOf(candidate))
	referenceValue := cloneFloat(spec.ValueOf(reference))
	metric := BetterFundMetric{
		Key:            spec.Key,
		Label:          spec.Label,
		Better:         spec.Better,
		CandidateValue: candidateValue,
		ReferenceValue: referenceValue,
		Weight:         spec.Weight,
	}

	if candidateValue != nil && referenceValue != nil {
		delta := roundPercent(*candidateValue - *referenceValue)
		metric.Delta = &delta

		advantage := *candidateValue - *referenceValue
		if spec.Better == "lower" {
			advantage = *referenceValue - *candidateValue
		}
		if advantage > 0 {
			metricAdvantage := roundPercent(advantage)
			metric.Advantage = &metricAdvantage
			metric.Contribution = roundPercent(metric.Contribution + metricAdvantage*spec.Weight)
		}
	}

	if rank, ok := rankMap[candidate.Code]; ok {
		metric.CandidateRank = rank.Rank
		metric.RankTotal = rank.Total
		metric.CandidatePercentile = cloneFloat(rank.Percentile)
	}
	if rank, ok := rankMap[reference.Code]; ok {
		metric.ReferenceRank = rank.Rank
		if metric.RankTotal == 0 {
			metric.RankTotal = rank.Total
		}
		metric.ReferencePercentile = cloneFloat(rank.Percentile)
	}

	rankEdge := 0.0
	if metric.CandidatePercentile != nil {
		if metric.ReferencePercentile != nil {
			rankEdge = *metric.CandidatePercentile - *metric.ReferencePercentile
		} else {
			rankEdge = *metric.CandidatePercentile - 50
		}
	}
	if rankEdge > 0 {
		metric.Contribution = roundPercent(metric.Contribution + rankEdge*spec.Weight*0.03)
	}

	return metric
}

func buildBetterMetricRanks(universe []data.FundBasic, spec betterMetricSpec) map[string]betterMetricRank {
	type rankItem struct {
		Code  string
		Value float64
	}
	items := make([]rankItem, 0, len(universe))
	for _, basic := range universe {
		value := spec.ValueOf(basic)
		if value == nil {
			continue
		}
		items = append(items, rankItem{Code: basic.Code, Value: *value})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Value == items[j].Value {
			return items[i].Code < items[j].Code
		}
		if spec.Better == "lower" {
			return items[i].Value < items[j].Value
		}
		return items[i].Value > items[j].Value
	})

	ranks := make(map[string]betterMetricRank, len(items))
	total := len(items)
	for idx, item := range items {
		rank := idx + 1
		ranks[item.Code] = betterMetricRank{
			Rank:       rank,
			Total:      total,
			Percentile: calcComparablePercentile(rank, total),
		}
	}
	return ranks
}

func describeBetterMetric(metric BetterFundMetric, format string) string {
	if metric.Advantage != nil {
		switch metric.Better {
		case "lower":
			return metric.Label + "更低 " + formatComparableValue(*metric.Advantage, format) + withRankSuffix(metric)
		default:
			return metric.Label + "更高 " + formatComparableValue(*metric.Advantage, format) + withRankSuffix(metric)
		}
	}
	if metric.CandidateRank > 0 && metric.RankTotal > 0 {
		return metric.Label + "同类第 " + strconv.Itoa(metric.CandidateRank) + "/" + strconv.Itoa(metric.RankTotal)
	}
	return metric.Label + "指标更稳"
}

func withRankSuffix(metric BetterFundMetric) string {
	if metric.CandidateRank > 0 && metric.RankTotal > 0 {
		return "，同类第 " + strconv.Itoa(metric.CandidateRank) + "/" + strconv.Itoa(metric.RankTotal)
	}
	return ""
}

func topStrings(items []string, limit int) []string {
	if len(items) == 0 {
		return []string{}
	}
	if limit <= 0 || len(items) <= limit {
		return items
	}
	return items[:limit]
}

func betterCategoryLabel(category string) string {
	switch strings.TrimSpace(category) {
	case "bond":
		return "债基"
	case "cash":
		return "现金管理"
	case "equity":
		return "权益类"
	default:
		return "已更新基金池"
	}
}

func betterSortLabel(dimension string) string {
	switch normalizeBetterFundDimension(dimension) {
	case "lower_drawdown":
		return "按风险得分排序：回撤、波动、夏普，再参考近3月和近6月收益"
	case "higher_return":
		return "按收益得分排序：近1月、近3月、近6月、近1年收益优先，辅以夏普和回撤"
	default:
		return "按综合得分排序：近3月、近6月、近1年收益联动回撤、夏普和 Calmar"
	}
}

func betterDataHint(refreshStatus FundRefreshStatus, universe betterCandidateUniverse) string {
	hint := refreshStatus.Message
	if strings.TrimSpace(hint) == "" {
		hint = "推荐基于本地基金指标缓存"
	}
	if universe.ComparedUniverse > 0 {
		hint += " 当前比较范围：" + universe.ScopeLabel + "，样本 " + strconv.Itoa(universe.ComparedUniverse) + " 只。"
	}
	return hint
}

func isReturnBetterMetric(key string) bool {
	switch key {
	case "growth7", "growth1", "growth3", "growth6", "growth12", "sharpe12", "calmar12":
		return true
	default:
		return false
	}
}

func isRiskBetterMetric(key string) bool {
	switch key {
	case "drawdown12", "volatility12", "sharpe12", "calmar12":
		return true
	default:
		return false
	}
}

func calcComparablePercentile(rank int, total int) *float64 {
	if rank <= 0 || total <= 0 {
		return nil
	}
	value := roundPercent((float64(total-rank+1) / float64(total)) * 100)
	return &value
}

func formatComparableValue(value float64, format string) string {
	switch format {
	case "ratio":
		return strconv.FormatFloat(roundPercent(value), 'f', 2, 64)
	default:
		return strconv.FormatFloat(roundPercent(value), 'f', 2, 64) + "%"
	}
}

func cloneFloat(value *float64) *float64 {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func (s *Service) refreshSingleStockHolding(h *Holding) {
	stockDataAPI := data.NewStockDataApi()
	stockInfos, err := stockDataAPI.GetStockCodeRealTimeData(h.StockCode)
	if err != nil || stockInfos == nil || len(*stockInfos) == 0 {
		return
	}

	stockInfo := (*stockInfos)[0]
	price := parseStockPrice(stockInfo.Price)
	if price <= 0 {
		return
	}

	h.CurrentPrice = price
	h.TotalValue = price * h.Quantity
	h.ProfitLoss = h.TotalValue - h.TotalCost
	if h.TotalCost > 0 {
		h.ProfitRate = roundPercent(h.ProfitLoss / h.TotalCost * 100)
	}
	h.TodayChange = stockInfo.ChangePrice * h.Quantity
	h.TodayRate = stockInfo.ChangePercent
}

func (s *Service) refreshFundScreeningMetrics(code string) bool {
	code = strings.TrimSpace(code)
	if code == "" {
		return false
	}

	api := data.NewFundApi()
	api.CrawlFundBasic(code)

	updates := map[string]any{
		"screen_updated_at": time.Now().Format("2006-01-02 15:04:05"),
	}

	if rankings, err := api.GetFundStageRankings(code); err == nil {
		for _, item := range rankings {
			switch normalizeFundStagePeriod(item.Period) {
			case "7d":
				updates["net_growth7"] = item.ReturnRate
			case "1m":
				updates["net_growth1"] = item.ReturnRate
			case "3m":
				updates["net_growth3"] = item.ReturnRate
			case "6m":
				updates["net_growth6"] = item.ReturnRate
			case "12m":
				updates["net_growth12"] = item.ReturnRate
			}
		}
	}

	if trend, _, _, err := api.GetFundTrend(code); err == nil {
		riskSnapshot := calcFundRiskSnapshot(trend, time.Now().AddDate(-1, 0, 0))
		if riskSnapshot.MaxDrawdown12 != nil {
			updates["max_drawdown12"] = riskSnapshot.MaxDrawdown12
		}
		if riskSnapshot.Volatility12 != nil {
			updates["volatility12"] = riskSnapshot.Volatility12
		}
		if riskSnapshot.Sharpe12 != nil {
			updates["sharpe12"] = riskSnapshot.Sharpe12
		}
		if riskSnapshot.Calmar12 != nil {
			updates["calmar12"] = riskSnapshot.Calmar12
		}
	}

	if industry, err := api.GetFundTopIndustry(code); err == nil && industry != nil {
		updates["top_industry"] = industry.Industry
		updates["top_industry_weight"] = industry.Weight
		updates["top_industry_date"] = industry.ReportDate
	}

	return db.Dao.Model(&data.FundBasic{}).Where("code = ?", code).Updates(updates).Error == nil
}

func normalizeFundStagePeriod(raw string) string {
	period := strings.TrimSpace(strings.ReplaceAll(raw, " ", ""))
	switch {
	case strings.Contains(period, "近1周"), strings.Contains(period, "近7天"), strings.Contains(period, "7天"), strings.Contains(period, "1周"):
		return "7d"
	case strings.Contains(period, "近1月"), strings.Contains(period, "1月"):
		return "1m"
	case strings.Contains(period, "近3月"), strings.Contains(period, "3月"):
		return "3m"
	case strings.Contains(period, "近6月"), strings.Contains(period, "6月"):
		return "6m"
	case strings.Contains(period, "近1年"), strings.Contains(period, "1年"):
		return "12m"
	default:
		return ""
	}
}

func (s *Service) refreshSingleFundHolding(h *Holding) {
	api := data.NewFundApi()
	ensureFollowedFund(h.StockCode, h.StockName)

	if basic, err := dbQueryFundBasic(h.StockCode); err != nil || fundBasicNeedsRefresh(basic) {
		if fund, crawlErr := api.CrawlFundBasic(h.StockCode); crawlErr == nil {
			if strings.TrimSpace(h.StockName) == "" {
				h.StockName = fund.Name
			}
		}
	}

	api.CrawlFundNetUnitValue(h.StockCode)
	api.CrawlFundNetEstimatedUnit(h.StockCode)

	var followed data.FollowedFund
	if err := db.Dao.Preload("FundBasic").Where("code = ?", h.StockCode).First(&followed).Error; err != nil {
		return
	}

	currentPrice := 0.0
	estimateFresh := isFreshEstimatedValue(followed.NetEstimatedTime)
	if estimateFresh && followed.NetEstimatedUnit != nil && *followed.NetEstimatedUnit > 0 {
		currentPrice = *followed.NetEstimatedUnit
	} else if followed.NetUnitValue != nil && *followed.NetUnitValue > 0 {
		currentPrice = *followed.NetUnitValue
	}
	if currentPrice <= 0 {
		return
	}

	if strings.TrimSpace(h.StockName) == "" {
		h.StockName = defaultLabel(followed.Name, followed.FundBasic.Name)
	}

	h.CurrentPrice = currentPrice
	h.TotalValue = currentPrice * h.Quantity
	h.ProfitLoss = h.TotalValue - h.TotalCost
	if h.TotalCost > 0 {
		h.ProfitRate = roundPercent(h.ProfitLoss / h.TotalCost * 100)
	}
	h.LatestDailyRate = nil
	h.LatestDailyUpdatedAt = ""
	h.TodayChange = 0
	h.TodayRate = 0
	if estimateFresh && followed.NetUnitValue != nil && followed.NetEstimatedUnit != nil {
		h.TodayChange = (*followed.NetEstimatedUnit - *followed.NetUnitValue) * h.Quantity
		if followed.NetEstimatedRate != nil {
			h.TodayRate = *followed.NetEstimatedRate
		}
	}

	_, trendUpdatedAt, latestReturn, err := api.GetFundTrend(h.StockCode)
	if latestReturn != nil {
		latestRate := roundPercent(*latestReturn)
		h.LatestDailyRate = &latestRate
	}
	if err == nil && strings.TrimSpace(trendUpdatedAt) != "" {
		h.LatestDailyUpdatedAt = trendUpdatedAt
	} else {
		h.LatestDailyUpdatedAt = normalizeFundUpdateTime(followed.NetUnitValueDate)
	}
}

func (s *Service) buildFundHoldingView(h Holding) FundHoldingView {
	view := FundHoldingView{Holding: h}

	var followed data.FollowedFund
	db.Dao.Preload("FundBasic").Where("code = ?", h.StockCode).First(&followed)

	if strings.TrimSpace(view.StockName) == "" {
		view.StockName = defaultLabel(followed.Name, followed.FundBasic.Name)
	}

	view.FundType = followed.FundBasic.Type
	view.FundCompany = followed.FundBasic.Company
	view.FundManager = followed.FundBasic.Manager
	view.FundRating = followed.FundBasic.Rating
	view.FundScale = followed.FundBasic.Scale
	view.NetUnitValue = followed.NetUnitValue
	view.NetUnitValueDate = followed.NetUnitValueDate
	view.NetEstimatedUnit = followed.NetEstimatedUnit
	view.NetEstimatedTime = followed.NetEstimatedTime
	view.NetEstimatedRate = followed.NetEstimatedRate
	view.NetGrowth1 = followed.FundBasic.NetGrowth1
	view.NetGrowth3 = followed.FundBasic.NetGrowth3
	view.NetGrowth6 = followed.FundBasic.NetGrowth6
	view.NetGrowth12 = followed.FundBasic.NetGrowth12
	view.NetGrowth36 = followed.FundBasic.NetGrowth36
	view.NetGrowthYTD = followed.FundBasic.NetGrowthYTD
	view.EstimateUpdated = isFreshEstimatedValue(followed.NetEstimatedTime)
	if view.CurrentPrice <= 0 {
		if view.EstimateUpdated && followed.NetEstimatedUnit != nil && *followed.NetEstimatedUnit > 0 {
			view.CurrentPrice = *followed.NetEstimatedUnit
		} else if followed.NetUnitValue != nil && *followed.NetUnitValue > 0 {
			view.CurrentPrice = *followed.NetUnitValue
		}
	}
	if view.LatestDailyRate == nil && followed.NetEstimatedRate != nil && view.EstimateUpdated {
		latestRate := roundPercent(*followed.NetEstimatedRate)
		view.LatestDailyRate = &latestRate
	}
	if strings.TrimSpace(view.LatestDailyUpdatedAt) == "" {
		if view.EstimateUpdated {
			view.LatestDailyUpdatedAt = normalizeFundUpdateTime(followed.NetEstimatedTime)
		} else {
			view.LatestDailyUpdatedAt = normalizeFundUpdateTime(followed.NetUnitValueDate)
		}
	}
	if view.EstimateUpdated {
		view.EstimateStatus = "今日估算已更新"
	} else if strings.TrimSpace(followed.NetEstimatedTime) != "" {
		view.EstimateStatus = "上次估值 " + followed.NetEstimatedTime
	} else {
		view.EstimateStatus = "暂无估值"
	}

	classification := classifyFundType(defaultLabel(followed.FundBasic.Type, defaultLabel(followed.FundBasic.Name, h.StockName)))
	view.Category = classification.Category
	view.CategoryLabel = classification.Label
	view.RiskLevel = classification.RiskLevel
	return view
}

type fundClassification struct {
	Category  string
	Label     string
	RiskLevel string
}

func classifyFundTypeByHolding(code string) fundClassification {
	var followed data.FollowedFund
	db.Dao.Preload("FundBasic").Where("code = ?", code).First(&followed)
	return classifyFundType(defaultLabel(followed.FundBasic.Type, defaultLabel(followed.FundBasic.Name, followed.Name)))
}

func classifyFundType(fundType string) fundClassification {
	normalized := strings.TrimSpace(strings.ToUpper(fundType))
	switch {
	case normalized == "":
		return fundClassification{Category: "other", Label: "未分类基金", RiskLevel: "中性"}
	case strings.Contains(normalized, "货币"), strings.Contains(normalized, "现金管理"), strings.Contains(normalized, "现金"):
		return fundClassification{Category: "cash", Label: "现金管理", RiskLevel: "低风险"}
	case strings.Contains(normalized, "同业存单"):
		return fundClassification{Category: "cash", Label: "同业存单基金", RiskLevel: "低风险"}
	case strings.Contains(normalized, "中短债"):
		return fundClassification{Category: "bond", Label: "中短债基金", RiskLevel: "低风险+"}
	case strings.Contains(normalized, "纯债"), strings.Contains(normalized, "长债"):
		return fundClassification{Category: "bond", Label: "纯债基金", RiskLevel: "低风险+"}
	case strings.Contains(normalized, "一级债"):
		return fundClassification{Category: "bond", Label: "一级债基", RiskLevel: "低风险+"}
	case strings.Contains(normalized, "二级债"):
		return fundClassification{Category: "bond", Label: "二级债基", RiskLevel: "中低风险"}
	case strings.Contains(normalized, "偏债"), strings.Contains(normalized, "债券型-混合"), strings.Contains(normalized, "债券混合"):
		return fundClassification{Category: "bond", Label: "偏债混合", RiskLevel: "中低风险"}
	case strings.Contains(normalized, "可转债"):
		return fundClassification{Category: "bond", Label: "可转债基金", RiskLevel: "中风险"}
	case strings.Contains(normalized, "债券指数"), strings.Contains(normalized, "债券"), strings.Contains(normalized, "债基"):
		return fundClassification{Category: "bond", Label: "债券基金", RiskLevel: "低风险"}
	case strings.Contains(normalized, "REIT"):
		return fundClassification{Category: "equity", Label: "REITs", RiskLevel: "中风险"}
	case strings.Contains(normalized, "QDII"):
		return fundClassification{Category: "equity", Label: "QDII基金", RiskLevel: "中高风险"}
	case strings.Contains(normalized, "FOF"):
		return fundClassification{Category: "equity", Label: "FOF基金", RiskLevel: "中风险"}
	case strings.Contains(normalized, "指数增强"):
		return fundClassification{Category: "equity", Label: "指数增强", RiskLevel: "中高风险"}
	case strings.Contains(normalized, "ETF联接"):
		return fundClassification{Category: "equity", Label: "ETF联接基金", RiskLevel: "中风险"}
	case strings.Contains(normalized, "ETF"):
		return fundClassification{Category: "equity", Label: "ETF基金", RiskLevel: "中风险"}
	case strings.Contains(normalized, "指数"):
		return fundClassification{Category: "equity", Label: "指数基金", RiskLevel: "中风险"}
	case strings.Contains(normalized, "偏股"):
		return fundClassification{Category: "equity", Label: "偏股混合", RiskLevel: "中高风险"}
	case strings.Contains(normalized, "灵活配置"):
		return fundClassification{Category: "equity", Label: "灵活配置混合", RiskLevel: "中风险"}
	case strings.Contains(normalized, "平衡"):
		return fundClassification{Category: "equity", Label: "平衡混合", RiskLevel: "中风险"}
	case strings.Contains(normalized, "混合"):
		return fundClassification{Category: "equity", Label: "混合基金", RiskLevel: "中风险"}
	case strings.Contains(normalized, "股票"):
		return fundClassification{Category: "equity", Label: "股票基金", RiskLevel: "中高风险"}
	default:
		return fundClassification{Category: "other", Label: "其他基金", RiskLevel: "中性"}
	}
}

func applyFundCategoryDBFilter(query *gorm.DB, category string) *gorm.DB {
	switch strings.TrimSpace(category) {
	case "bond":
		return query.Where(
			"type LIKE ? OR type LIKE ? OR type LIKE ? OR type LIKE ? OR type LIKE ? OR type LIKE ? OR type LIKE ? OR type LIKE ? OR type LIKE ?",
			"%债券%", "%债基%", "%纯债%", "%偏债%", "%一级债%", "%二级债%", "%可转债%", "%中短债%", "%长债%",
		)
	case "cash":
		return query.Where("type LIKE ? OR type LIKE ? OR type LIKE ? OR type LIKE ?", "%货币%", "%现金管理%", "%现金%", "%同业存单%")
	case "equity":
		return query.Where(
			"type LIKE ? OR type LIKE ? OR type LIKE ? OR type LIKE ? OR type LIKE ? OR type LIKE ? OR type LIKE ? OR type LIKE ? OR type LIKE ? OR type LIKE ?",
			"%混合%", "%股票%", "%指数%", "%ETF%", "%QDII%", "%FOF%", "%REIT%", "%灵活配置%", "%偏股%", "%平衡%",
		)
	case "other":
		return query.Where("(type IS NULL OR type = '')")
	default:
		return query
	}
}
func mapFundScreenerSortColumn(sortBy string) string {
	switch strings.TrimSpace(sortBy) {
	case "growth7":
		return "net_growth7"
	case "growth1":
		return "net_growth1"
	case "growth3":
		return "net_growth3"
	case "growth6":
		return "net_growth6"
	case "growth12":
		return "net_growth12"
	case "drawdown12":
		return "max_drawdown12"
	case "industry":
		return "top_industry"
	case "updatedAt":
		return "screen_updated_at"
	case "company":
		return "company"
	default:
		return "net_growth3"
	}
}

func isFundWatchlisted(code string) bool {
	var count int64
	db.Dao.Model(&data.FollowedFund{}).
		Where("code = ? AND is_watchlist = ?", strings.TrimSpace(code), true).
		Count(&count)
	return count > 0
}

func loadFundWatchlistMap(basics []data.FundBasic) map[string]bool {
	if len(basics) == 0 {
		return map[string]bool{}
	}

	codes := make([]string, 0, len(basics))
	for _, item := range basics {
		codes = append(codes, item.Code)
	}

	var followed []data.FollowedFund
	db.Dao.Where("is_watchlist = ? AND code IN ?", true, codes).Find(&followed)

	result := make(map[string]bool, len(followed))
	for _, item := range followed {
		result[item.Code] = true
	}
	return result
}

func buildBetterFundCandidate(reference data.FundBasic, candidate data.FundBasic, watchlist bool) BetterFundCandidate {
	item := buildFundScreenerItem(candidate, watchlist)
	reasons := make([]string, 0, 5)
	score := 0.0

	appendReason := func(label string, candidateVal *float64, referenceVal *float64, reverse bool) {
		if candidateVal == nil || referenceVal == nil {
			return
		}
		diff := *candidateVal - *referenceVal
		if reverse {
			diff = *referenceVal - *candidateVal
		}
		if diff <= 0 {
			return
		}
		score += diff
		reasons = append(reasons, label+"领先 "+formatFloatPtrDiff(diff)+"%")
	}

	appendReason("近7天", candidate.NetGrowth7, reference.NetGrowth7, false)
	appendReason("近1月", candidate.NetGrowth1, reference.NetGrowth1, false)
	appendReason("近3月", candidate.NetGrowth3, reference.NetGrowth3, false)
	appendReason("近6月", candidate.NetGrowth6, reference.NetGrowth6, false)
	appendReason("最大回撤", candidate.MaxDrawdown12, reference.MaxDrawdown12, true)

	return BetterFundCandidate{
		FundScreenerItem: item,
		BetterScore:      roundPercent(score),
		Reasons:          reasons,
	}
}

func formatFloatPtrDiff(val float64) string {
	return strconv.FormatFloat(roundPercent(val), 'f', 2, 64)
}

func buildBetterFundCandidateByDimension(reference data.FundBasic, candidate data.FundBasic, watchlist bool, dimension string) (BetterFundCandidate, bool) {
	item := buildFundScreenerItem(candidate, watchlist)
	reasons := make([]string, 0, 5)
	score := 0.0
	hasReturnAdvantage := false
	hasDrawdownAdvantage := false

	appendReturnReason := func(label string, candidateVal *float64, referenceVal *float64, weight float64) {
		if candidateVal == nil {
			return
		}
		ref := 0.0
		if referenceVal != nil {
			ref = *referenceVal
		}
		diff := *candidateVal - ref
		if diff <= 0 {
			return
		}
		score += diff * weight
		hasReturnAdvantage = true
		reasons = append(reasons, label+"更高 "+formatFloatPtrDiff(diff)+"%")
	}

	appendDrawdownReason := func(candidateVal *float64, referenceVal *float64, weight float64) {
		if candidateVal == nil {
			return
		}
		ref := 999.0
		if referenceVal != nil {
			ref = *referenceVal
		}
		diff := ref - *candidateVal
		if diff <= 0 {
			return
		}
		score += diff * weight
		hasDrawdownAdvantage = true
		reasons = append(reasons, "最大回撤更低 "+formatFloatPtrDiff(diff)+"%")
	}

	switch normalizeBetterFundDimension(dimension) {
	case "lower_drawdown":
		if candidate.MaxDrawdown12 == nil {
			return BetterFundCandidate{}, false
		}
		appendDrawdownReason(candidate.MaxDrawdown12, reference.MaxDrawdown12, 1.8)
		appendReturnReason("近3月", candidate.NetGrowth3, reference.NetGrowth3, 0.35)
		appendReturnReason("近6月", candidate.NetGrowth6, reference.NetGrowth6, 0.45)
		if !hasDrawdownAdvantage {
			return BetterFundCandidate{}, false
		}
	case "higher_return":
		appendReturnReason("近7天", candidate.NetGrowth7, reference.NetGrowth7, 0.25)
		appendReturnReason("近1月", candidate.NetGrowth1, reference.NetGrowth1, 0.55)
		appendReturnReason("近3月", candidate.NetGrowth3, reference.NetGrowth3, 0.90)
		appendReturnReason("近6月", candidate.NetGrowth6, reference.NetGrowth6, 1.20)
		appendDrawdownReason(candidate.MaxDrawdown12, reference.MaxDrawdown12, 0.25)
		if !hasReturnAdvantage {
			return BetterFundCandidate{}, false
		}
	default:
		appendReturnReason("近7天", candidate.NetGrowth7, reference.NetGrowth7, 0.20)
		appendReturnReason("近1月", candidate.NetGrowth1, reference.NetGrowth1, 0.50)
		appendReturnReason("近3月", candidate.NetGrowth3, reference.NetGrowth3, 0.80)
		appendReturnReason("近6月", candidate.NetGrowth6, reference.NetGrowth6, 1.00)
		appendDrawdownReason(candidate.MaxDrawdown12, reference.MaxDrawdown12, 1.10)
	}

	if len(reasons) == 0 || score <= 0 {
		return BetterFundCandidate{}, false
	}

	return BetterFundCandidate{
		FundScreenerItem: item,
		BetterScore:      roundPercent(score),
		Reasons:          reasons,
	}, true
}

func normalizeBetterFundDimension(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "lower_drawdown":
		return "lower_drawdown"
	case "higher_return":
		return "higher_return"
	default:
		return "balanced"
	}
}

func buildFundScreenerItem(basic data.FundBasic, watchlist bool) FundScreenerItem {
	classification := classifyFundType(defaultLabel(basic.Type, basic.Name))
	return FundScreenerItem{
		Code:              basic.Code,
		Name:              defaultLabel(basic.Name, basic.FullName),
		FundType:          basic.Type,
		Category:          classification.Category,
		CategoryLabel:     classification.Label,
		RiskLevel:         classification.RiskLevel,
		Company:           basic.Company,
		Manager:           basic.Manager,
		Rating:            basic.Rating,
		Scale:             basic.Scale,
		TopIndustry:       basic.TopIndustry,
		TopIndustryWeight: basic.TopIndustryWeight,
		TopIndustryDate:   basic.TopIndustryDate,
		NetGrowth7:        basic.NetGrowth7,
		NetGrowth1:        basic.NetGrowth1,
		NetGrowth3:        basic.NetGrowth3,
		NetGrowth6:        basic.NetGrowth6,
		NetGrowth12:       basic.NetGrowth12,
		MaxDrawdown12:     basic.MaxDrawdown12,
		Volatility12:      basic.Volatility12,
		Sharpe12:          basic.Sharpe12,
		Calmar12:          basic.Calmar12,
		ScreenUpdatedAt:   basic.ScreenUpdatedAt,
		Watchlist:         watchlist,
	}
}

func loadFundTypeOptions() []string {
	var items []string
	db.Dao.Model(&data.FundBasic{}).
		Where("type IS NOT NULL AND type <> ''").
		Distinct().
		Order("type asc").
		Pluck("type", &items)
	return items
}

func loadFundIndustryOptions() []string {
	var items []string
	db.Dao.Model(&data.FundBasic{}).
		Where("top_industry IS NOT NULL AND top_industry <> ''").
		Distinct().
		Order("top_industry asc").
		Pluck("top_industry", &items)
	return items
}

func loadFundScreenRefreshHint() string {
	type refreshRow struct {
		Value string
	}
	var row refreshRow
	db.Dao.Raw("SELECT MAX(screen_updated_at) AS value FROM fund_basic").Scan(&row)
	return strings.TrimSpace(row.Value)
}

func loadScreenedFundCount() int64 {
	var count int64
	db.Dao.Model(&data.FundBasic{}).
		Where("screen_updated_at IS NOT NULL AND screen_updated_at <> ''").
		Count(&count)
	return count
}

func loadUpdatedTodayFundCount(today string) int64 {
	var count int64
	db.Dao.Model(&data.FundBasic{}).
		Where("screen_updated_at LIKE ?", strings.TrimSpace(today)+"%").
		Count(&count)
	return count
}

func calcFundMaxDrawdown(points []data.FundTrendPoint, since time.Time) *float64 {
	if len(points) == 0 {
		return nil
	}

	filtered := make([]data.FundTrendPoint, 0, len(points))
	for _, point := range points {
		if point.Timestamp <= 0 {
			continue
		}
		if time.UnixMilli(point.Timestamp).Before(since) {
			continue
		}
		filtered = append(filtered, point)
	}
	if len(filtered) == 0 {
		filtered = points
	}

	peak := 0.0
	maxDrawdown := 0.0
	for _, point := range filtered {
		if point.Value <= 0 {
			continue
		}
		if point.Value > peak {
			peak = point.Value
			continue
		}
		if peak <= 0 {
			continue
		}
		drawdown := (peak - point.Value) / peak * 100
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}

	value := roundPercent(maxDrawdown)
	return &value
}

func calcFundRiskSnapshot(points []data.FundTrendPoint, since time.Time) fundRiskSnapshot {
	snapshot := fundRiskSnapshot{
		MaxDrawdown12: calcFundMaxDrawdown(points, since),
	}
	if len(points) < 2 {
		return snapshot
	}

	filtered := make([]data.FundTrendPoint, 0, len(points))
	for _, point := range points {
		if point.Timestamp <= 0 {
			continue
		}
		if time.UnixMilli(point.Timestamp).Before(since) {
			continue
		}
		filtered = append(filtered, point)
	}
	if len(filtered) < 2 {
		filtered = points
	}
	if len(filtered) < 2 {
		return snapshot
	}

	dailyReturns := make([]float64, 0, len(filtered)-1)
	for i := 1; i < len(filtered); i++ {
		if filtered[i].DailyReturn != nil {
			dailyReturns = append(dailyReturns, *filtered[i].DailyReturn/100)
			continue
		}
		prev := filtered[i-1].Value
		if prev <= 0 {
			continue
		}
		dailyReturns = append(dailyReturns, (filtered[i].Value/prev)-1)
	}
	if len(dailyReturns) < 2 {
		return snapshot
	}

	dailyMean := 0.0
	for _, item := range dailyReturns {
		dailyMean += item
	}
	dailyMean /= float64(len(dailyReturns))

	variance := 0.0
	for _, item := range dailyReturns {
		diff := item - dailyMean
		variance += diff * diff
	}
	variance /= float64(len(dailyReturns) - 1)
	if variance < 0 {
		variance = 0
	}
	annualVolatility := math.Sqrt(variance) * math.Sqrt(252)
	if annualVolatility > 0 {
		value := roundPercent(annualVolatility * 100)
		snapshot.Volatility12 = &value
	}

	firstValue := filtered[0].Value
	lastValue := filtered[len(filtered)-1].Value
	if firstValue <= 0 || lastValue <= 0 {
		return snapshot
	}
	days := time.UnixMilli(filtered[len(filtered)-1].Timestamp).Sub(time.UnixMilli(filtered[0].Timestamp)).Hours() / 24
	if days <= 0 {
		days = float64(len(dailyReturns))
	}
	years := days / 365
	if years <= 0 {
		years = float64(len(dailyReturns)) / 252
	}
	if years <= 0 {
		return snapshot
	}

	annualReturn := math.Pow(lastValue/firstValue, 1/years) - 1
	if annualVolatility > 0 && !math.IsNaN(annualReturn) && !math.IsInf(annualReturn, 0) {
		value := roundPercent(annualReturn / annualVolatility)
		snapshot.Sharpe12 = &value
	}
	if snapshot.MaxDrawdown12 != nil && *snapshot.MaxDrawdown12 > 0 && !math.IsNaN(annualReturn) && !math.IsInf(annualReturn, 0) {
		drawdownDecimal := *snapshot.MaxDrawdown12 / 100
		if drawdownDecimal > 0 {
			value := roundPercent(annualReturn / drawdownDecimal)
			snapshot.Calmar12 = &value
		}
	}

	return snapshot
}

func normalizeHoldingType(val string) string {
	switch strings.ToLower(strings.TrimSpace(val)) {
	case "fund":
		return "fund"
	case "stock":
		return "stock"
	default:
		return ""
	}
}

func inferHoldingTypeByCode(code string) string {
	trimmed := strings.ToLower(strings.TrimSpace(code))
	switch {
	case strings.HasPrefix(trimmed, "sh"), strings.HasPrefix(trimmed, "sz"), strings.HasPrefix(trimmed, "hk"), strings.HasPrefix(trimmed, "us"):
		return "stock"
	case len(trimmed) == 6:
		return "fund"
	default:
		return "stock"
	}
}

func defaultLabel(val string, fallback string) string {
	if strings.TrimSpace(val) == "" {
		return fallback
	}
	return strings.TrimSpace(val)
}
func roundPercent(val float64) float64 {
	return math.Round(val*100) / 100
}

func addAllocation(m map[string]*AllocationItem, label string, value float64) {
	item, ok := m[label]
	if !ok {
		item = &AllocationItem{Label: label}
		m[label] = item
	}
	item.Value += value
	item.Count++
}

func allocationSlice(m map[string]*AllocationItem, total float64) []AllocationItem {
	items := make([]AllocationItem, 0, len(m))
	for _, item := range m {
		if total > 0 {
			item.Ratio = roundPercent(item.Value / total * 100)
		}
		items = append(items, *item)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Value == items[j].Value {
			return items[i].Label < items[j].Label
		}
		return items[i].Value > items[j].Value
	})
	return items
}

func ensureFollowedFund(code string, name string) {
	var followed data.FollowedFund
	err := db.Dao.Where("code = ?", code).First(&followed).Error
	if err == nil {
		return
	}
	if err == gorm.ErrRecordNotFound {
		db.Dao.Create(&data.FollowedFund{Code: code, Name: name, IsWatchlist: false})
	}
}

func dbQueryFundBasic(code string) (*data.FundBasic, error) {
	var fund data.FundBasic
	if err := db.Dao.Where("code = ?", code).First(&fund).Error; err != nil {
		return nil, err
	}
	return &fund, nil
}

func fundBasicNeedsRefresh(basic *data.FundBasic) bool {
	if basic == nil {
		return true
	}
	return strings.TrimSpace(basic.Type) == "" ||
		strings.TrimSpace(basic.Company) == "" ||
		strings.TrimSpace(basic.Manager) == ""
}

func isFreshEstimatedValue(estimatedTime string) bool {
	estimatedTime = strings.TrimSpace(estimatedTime)
	if estimatedTime == "" {
		return false
	}

	location := time.Now().Location()
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		time.RFC3339,
	}

	var parsed time.Time
	var err error
	for _, layout := range layouts {
		parsed, err = time.ParseInLocation(layout, estimatedTime, location)
		if err == nil {
			break
		}
	}
	if err != nil {
		return false
	}

	today := time.Now().Format("2006-01-02")
	return parsed.Format("2006-01-02") == today
}

func isSameTradingDayString(value string, target time.Time) bool {
	text := strings.TrimSpace(value)
	if text == "" {
		return false
	}

	location := target.Location()
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
		time.RFC3339,
	}
	for _, layout := range layouts {
		parsed, err := time.ParseInLocation(layout, text, location)
		if err == nil {
			return parsed.Format("2006-01-02") == target.Format("2006-01-02")
		}
	}
	return strings.HasPrefix(text, target.Format("2006-01-02"))
}

func normalizeFundUpdateTime(value string) string {
	text := strings.TrimSpace(value)
	if text == "" {
		return ""
	}
	if strings.Contains(text, ":") {
		return text
	}
	return text + " 15:00"
}

func parseStockPrice(raw string) float64 {
	if strings.TrimSpace(raw) == "" {
		return 0
	}
	val, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0
	}
	return val
}
