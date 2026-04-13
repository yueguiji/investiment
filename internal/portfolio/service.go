package portfolio

import (
	"encoding/json"
	"math"
	"regexp"
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

type focusedRefreshCodesRuntimeCache struct {
	mu      sync.RWMutex
	key     string
	expires time.Time
	codes   []string
}

type recommendationRefreshRuntimeState struct {
	mu            sync.Mutex
	refreshing    bool
	progressNow   int64
	progressTotal int64
	currentCode   string
	lastStarted   time.Time
	lastFinished  time.Time
}

var fundRefreshState fundRefreshRuntimeState
var focusedRefreshCodesCache focusedRefreshCodesRuntimeCache
var recommendationRefreshState recommendationRefreshRuntimeState

const (
	fundRefreshStateNotStarted    = "not_started"
	fundRefreshStatePartial       = "partial"
	fundRefreshStateCompleted     = "completed"
	fundRefreshScopeFocused       = "watchlist_related"
	fundRefreshScopeAll           = "all_pending"
	recommendationStatusPending   = "pending"
	recommendationStatusRunning   = "running"
	recommendationStatusCompleted = "completed"
	recommendationStatusFailed    = "failed"
	betterFundRefreshWorkers      = 6
	betterFundRefreshCap          = 600
	betterFundTargetPerDimension  = 5
	betterFundRefreshProbeBudget  = 24
)

type betterCandidateUniverse struct {
	Basics           []data.FundBasic
	ScopeLabel       string
	FallbackApplied  bool
	ComparedUniverse int
	UniverseTotal    int
	NetworkRefresh   bool
	RefreshedCount   int
	Limited          bool
}

type betterUniverseRefreshStats struct {
	UniverseTotal  int
	RefreshedCount int
	NetworkRefresh bool
	Limited        bool
}

type fundRefreshScopeSnapshot struct {
	Scope        string
	TargetCount  int64
	UpdatedToday int64
	PendingCount int64
}

type focusedFundRefreshProgress struct {
	TargetCodes  []string
	PendingCodes []string
	TrackedSet   map[string]struct{}
	Completed    int64
	Pending      int64
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
	MaxDrawdown1  *float64
	MaxDrawdown3  *float64
	MaxDrawdown6  *float64
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
	trackedSet := map[string]struct{}{}
	if scope == fundRefreshScopeFocused {
		trackedSet = s.loadFocusedFundRefreshProgress(today).TrackedSet
	}
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
		if _, tracked := trackedSet[strings.TrimSpace(basic.Code)]; tracked {
			s.refreshFollowedFundMarketData(basic.Code, basic.Name)
		}
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
	if scope == fundRefreshScopeFocused {
		go s.GetFundRecommendationRefreshStatus(true)
	}
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
		progress := s.loadFocusedFundRefreshProgress(today)
		if len(progress.PendingCodes) == 0 {
			return db.Dao.Where("1 = 0")
		}
		return db.Dao.Where("code IN ?", progress.PendingCodes)
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
		progress := s.loadFocusedFundRefreshProgress(today)
		return fundRefreshScopeSnapshot{
			Scope:        fundRefreshScopeFocused,
			TargetCount:  int64(len(progress.TargetCodes)),
			UpdatedToday: progress.Completed,
			PendingCount: progress.Pending,
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
	for _, code := range loadTrackedFundCodes() {
		if trimmed := strings.TrimSpace(code); trimmed != "" {
			codeSet[trimmed] = struct{}{}
		}
	}

	seedCodes := sortedCodeKeys(codeSet)
	if len(seedCodes) == 0 {
		return seedCodes
	}

	cacheKey := strings.Join(seedCodes, ",")
	focusedRefreshCodesCache.mu.RLock()
	if focusedRefreshCodesCache.key == cacheKey && time.Now().Before(focusedRefreshCodesCache.expires) && len(focusedRefreshCodesCache.codes) > 0 {
		cached := append([]string(nil), focusedRefreshCodesCache.codes...)
		focusedRefreshCodesCache.mu.RUnlock()
		return cached
	}
	focusedRefreshCodesCache.mu.RUnlock()

	for _, code := range s.collectRecommendedFundCodes(loadFundWatchlistCodes(), 3) {
		if trimmed := strings.TrimSpace(code); trimmed != "" {
			codeSet[trimmed] = struct{}{}
		}
	}

	result := sortedCodeKeys(codeSet)
	focusedRefreshCodesCache.mu.Lock()
	focusedRefreshCodesCache.key = cacheKey
	focusedRefreshCodesCache.codes = append([]string(nil), result...)
	focusedRefreshCodesCache.expires = time.Now().Add(30 * time.Second)
	focusedRefreshCodesCache.mu.Unlock()
	return result
}

func loadTrackedFundCodes() []string {
	codeSet := make(map[string]struct{})
	for _, code := range loadFundWatchlistCodes() {
		if trimmed := strings.TrimSpace(code); trimmed != "" {
			codeSet[trimmed] = struct{}{}
		}
	}
	for _, code := range loadFundHoldingCodes() {
		if trimmed := strings.TrimSpace(code); trimmed != "" {
			codeSet[trimmed] = struct{}{}
		}
	}
	return sortedCodeKeys(codeSet)
}

func (s *Service) loadFocusedFundRefreshProgress(today string) focusedFundRefreshProgress {
	targetCodes := s.loadFocusedFundRefreshCodes()
	if len(targetCodes) == 0 {
		return focusedFundRefreshProgress{
			TargetCodes:  []string{},
			PendingCodes: []string{},
			TrackedSet:   map[string]struct{}{},
		}
	}

	trackedSet := make(map[string]struct{})
	for _, code := range loadTrackedFundCodes() {
		if trimmed := strings.TrimSpace(code); trimmed != "" {
			trackedSet[trimmed] = struct{}{}
		}
	}

	var basics []data.FundBasic
	db.Dao.Where("code IN ?", targetCodes).Find(&basics)
	basicMap := make(map[string]data.FundBasic, len(basics))
	for _, basic := range basics {
		basicMap[strings.TrimSpace(basic.Code)] = basic
	}

	followedMap := make(map[string]data.FollowedFund)
	if len(trackedSet) > 0 {
		trackedCodes := make([]string, 0, len(trackedSet))
		for code := range trackedSet {
			trackedCodes = append(trackedCodes, code)
		}

		var followed []data.FollowedFund
		db.Dao.Where("code IN ?", trackedCodes).Find(&followed)
		for _, item := range followed {
			followedMap[strings.TrimSpace(item.Code)] = item
		}
	}

	now := time.Now()
	var completed int64
	pendingCodes := make([]string, 0, len(targetCodes))
	for _, code := range targetCodes {
		basic, ok := basicMap[code]
		if !ok || shouldRefreshBetterFundCandidate(basic, now) {
			pendingCodes = append(pendingCodes, code)
			continue
		}
		if _, tracked := trackedSet[code]; tracked {
			followed, ok := followedMap[code]
			if !ok || !isSameTradingDayTime(followed.UpdatedAt, now) {
				pendingCodes = append(pendingCodes, code)
				continue
			}
		}
		completed++
	}

	return focusedFundRefreshProgress{
		TargetCodes:  targetCodes,
		PendingCodes: pendingCodes,
		TrackedSet:   trackedSet,
		Completed:    completed,
		Pending:      int64(len(pendingCodes)),
	}
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

func loadFundRecommendationLastRefreshHint(today string) string {
	var progress FundRecommendationProgress
	if err := db.Dao.
		Where("refresh_date = ? AND status = ?", today, recommendationStatusCompleted).
		Order("updated_at desc").
		First(&progress).Error; err == nil {
		return progress.UpdatedAt.Format("2006-01-02 15:04:05")
	}
	return ""
}

func normalizeFundRecommendationRunningRows(today string) {
	recommendationRefreshState.mu.Lock()
	refreshing := recommendationRefreshState.refreshing
	recommendationRefreshState.mu.Unlock()
	if refreshing {
		return
	}
	db.Dao.Model(&FundRecommendationProgress{}).
		Where("refresh_date = ? AND status = ?", today, recommendationStatusRunning).
		Updates(map[string]any{
			"status":        recommendationStatusPending,
			"current_phase": "",
			"message":       "",
		})
}

func seedFundRecommendationProgress(today string, codes []string) {
	if len(codes) == 0 {
		return
	}

	var rows []FundRecommendationProgress
	db.Dao.Where("refresh_date = ? AND reference_code IN ?", today, codes).Find(&rows)
	existing := make(map[string]FundRecommendationProgress, len(rows))
	for _, row := range rows {
		existing[strings.TrimSpace(row.ReferenceCode)] = row
	}

	for _, code := range codes {
		trimmed := strings.TrimSpace(code)
		if trimmed == "" {
			continue
		}
		row, ok := existing[trimmed]
		if !ok {
			db.Dao.Create(&FundRecommendationProgress{
				ReferenceCode: trimmed,
				RefreshDate:   today,
				Status:        recommendationStatusPending,
			})
			continue
		}
		if row.Status == recommendationStatusCompleted && !hasRequiredRecommendationCachesForToday(trimmed, today) {
			db.Dao.Model(&FundRecommendationProgress{}).
				Where("id = ?", row.ID).
				Updates(map[string]any{
					"status":        recommendationStatusPending,
					"current_phase": "",
					"message":       "推荐缓存缺少必要维度，等待后台补齐",
				})
			continue
		}
		if row.Status == recommendationStatusRunning {
			db.Dao.Model(&FundRecommendationProgress{}).
				Where("id = ?", row.ID).
				Updates(map[string]any{
					"status":        recommendationStatusPending,
					"current_phase": "",
					"message":       "",
				})
		}
	}
}

func hasRequiredRecommendationCachesForToday(code string, today string) bool {
	requiredDimensions := []string{"balanced", "higher_return", "lower_drawdown"}
	var count int64
	db.Dao.Model(&FundRecommendationCache{}).
		Where(
			"reference_code = ? AND refresh_date = ? AND same_type_only = ? AND IFNULL(same_sub_type_only, 0) = ? AND fee_free7 = ? AND fee_free30 = ? AND include_a_class = ? AND IFNULL(only_a_class, 0) = ? AND dimension IN ?",
			strings.TrimSpace(code), today, true, false, true, true, false, false, requiredDimensions,
		).
		Distinct("dimension").
		Count(&count)
	return count == int64(len(requiredDimensions))
}

func resolveFundRecommendationRefreshState(completed, pending int64, watchlistCount int64) (string, string) {
	switch {
	case watchlistCount <= 0:
		return fundRefreshStateCompleted, "今日完成"
	case completed <= 0 && pending > 0:
		return fundRefreshStateNotStarted, "未开始"
	case pending > 0:
		return fundRefreshStatePartial, "部分更新"
	default:
		return fundRefreshStateCompleted, "今日完成"
	}
}

func (s *Service) buildFundRecommendationRefreshStatus(autoStart bool) FundRecommendationRefreshStatus {
	today := time.Now().Format("2006-01-02")
	codes := loadFundWatchlistCodes()
	normalizeFundRecommendationRunningRows(today)
	seedFundRecommendationProgress(today, codes)

	recommendationRefreshState.mu.Lock()
	refreshing := recommendationRefreshState.refreshing
	progressNow := recommendationRefreshState.progressNow
	progressTotal := recommendationRefreshState.progressTotal
	currentCode := recommendationRefreshState.currentCode
	recommendationRefreshState.mu.Unlock()

	var rows []FundRecommendationProgress
	if len(codes) > 0 {
		db.Dao.Where("refresh_date = ? AND reference_code IN ?", today, codes).Find(&rows)
	}

	completed := int64(0)
	failed := int64(0)
	pending := int64(0)
	for _, row := range rows {
		switch row.Status {
		case recommendationStatusCompleted:
			completed++
		case recommendationStatusFailed:
			failed++
			pending++
		default:
			pending++
		}
	}

	state, stateLabel := resolveFundRecommendationRefreshState(completed, pending, int64(len(codes)))
	lastRefreshHint := loadFundRecommendationLastRefreshHint(today)
	message := ""
	fundRefreshStatus := s.buildFundRefreshStatus(s.EnsureFundUniverse())
	if autoStart && len(codes) > 0 && fundRefreshStatus.TargetPending > 0 && !fundRefreshStatus.Refreshing {
		fundRefreshStatus = s.ensureFundScreenerFreshAsync(0)
	}
	switch {
	case len(codes) == 0:
		message = "当前还没有自选基金，推荐检索任务处于空闲状态。"
	case fundRefreshStatus.Refreshing && fundRefreshStatus.Scope == fundRefreshScopeFocused:
		message = "正在等待今日自选相关基金指标更新完成，随后会自动继续推荐检索。"
	case refreshing:
		message = "后台正在检索自选基金的推荐结果，未完成的基金稍后会继续补齐。"
	case state == fundRefreshStateCompleted:
		message = "今日自选基金推荐缓存已完成，可以直接打开对比推荐查看。"
	case state == fundRefreshStatePartial:
		message = "今日自选基金推荐缓存部分完成，后续会从剩余基金继续补齐。"
	default:
		message = "今日自选基金推荐缓存尚未开始。"
	}

	status := FundRecommendationRefreshStatus{
		State:           state,
		StateLabel:      stateLabel,
		Refreshing:      refreshing,
		CurrentDate:     today,
		WatchlistCount:  int64(len(codes)),
		CompletedCount:  completed,
		PendingCount:    pending,
		FailedCount:     failed,
		ProgressCurrent: progressNow,
		ProgressTotal:   progressTotal,
		CurrentCode:     currentCode,
		LastRefreshHint: lastRefreshHint,
		Message:         message,
	}

	if autoStart && len(codes) > 0 && !refreshing && !(fundRefreshStatus.Refreshing && fundRefreshStatus.Scope == fundRefreshScopeFocused) && pending > 0 {
		if s.startFundRecommendationRefresh(today, codes) {
			status.Triggered = true
			status.Refreshing = true
			status.Message = "已启动今日自选基金推荐检索，结果会在后台逐只补齐，未完成时下次会从剩余基金继续。"
		}
	}

	return status
}

func (s *Service) startFundRecommendationRefresh(today string, codes []string) bool {
	recommendationRefreshState.mu.Lock()
	if recommendationRefreshState.refreshing {
		recommendationRefreshState.mu.Unlock()
		return false
	}
	recommendationRefreshState.refreshing = true
	recommendationRefreshState.progressNow = 0
	recommendationRefreshState.progressTotal = 0
	recommendationRefreshState.currentCode = ""
	recommendationRefreshState.lastStarted = time.Now()
	recommendationRefreshState.mu.Unlock()

	go s.runFundRecommendationRefresh(today, codes)
	return true
}

func (s *Service) GetFundRecommendationRefreshStatus(autoStart bool) FundRecommendationRefreshStatus {
	return s.buildFundRecommendationRefreshStatus(autoStart)
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
	preferredQuery := betterFundDefaultCacheQuery()
	for _, code := range dedupeCodes(referenceCodes) {
		reference, err := dbQueryFundBasic(code)
		if err != nil || reference == nil {
			continue
		}
		refCategory := classifyFundType(defaultLabel(reference.Type, reference.Name))
		universe := loadBetterFundUniverse(*reference, refCategory.Category, true, false)
		if len(universe.Basics) == 0 {
			continue
		}
		for _, dimension := range dimensions {
			preferredQuery.Dimension = dimension
			candidates := buildBetterFundCandidates(*reference, universe, dimension)
			candidates = finalizeBetterFundCandidates(candidates, universe)
			candidates = applyBetterFundPreferenceFilters(candidates, preferredQuery)
			candidates, _ = filterCandidatesStrongerThanReference(candidates)
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

func upsertFundRecommendationProgressStatus(today string, code string, updates map[string]any) {
	trimmed := strings.TrimSpace(code)
	if trimmed == "" {
		return
	}
	result := db.Dao.Model(&FundRecommendationProgress{}).
		Where("refresh_date = ? AND reference_code = ?", today, trimmed).
		Updates(updates)
	if result.Error == nil && result.RowsAffected > 0 {
		return
	}
	row := FundRecommendationProgress{
		ReferenceCode: trimmed,
		RefreshDate:   today,
	}
	if value, ok := updates["status"].(string); ok {
		row.Status = value
	}
	if value, ok := updates["current_phase"].(string); ok {
		row.CurrentPhase = value
	}
	if value, ok := updates["message"].(string); ok {
		row.Message = value
	}
	if value, ok := updates["last_error"].(string); ok {
		row.LastError = value
	}
	if value, ok := updates["compared_universe"].(int); ok {
		row.ComparedUniverse = value
	}
	if value, ok := updates["universe_total"].(int); ok {
		row.UniverseTotal = value
	}
	db.Dao.Create(&row)
}

func betterFundDefaultCacheQuery() BetterFundQuery {
	return BetterFundQuery{
		FeeFree7:      true,
		FeeFree30:     true,
		IncludeAClass: false,
		OnlyAClass:    false,
		Page:          1,
		PageSize:      30,
	}
}

func betterFundUnrestrictedCacheQuery() BetterFundQuery {
	return BetterFundQuery{
		FeeFree7:      false,
		FeeFree30:     false,
		IncludeAClass: true,
		OnlyAClass:    false,
		Page:          1,
		PageSize:      30,
	}
}

func sameBetterFundPreferenceProfile(left BetterFundQuery, right BetterFundQuery) bool {
	return left.FeeFree7 == right.FeeFree7 &&
		left.FeeFree30 == right.FeeFree30 &&
		left.IncludeAClass == right.IncludeAClass &&
		left.OnlyAClass == right.OnlyAClass
}

func buildBetterFundCacheProfiles(base BetterFundQuery) []BetterFundQuery {
	profiles := make([]BetterFundQuery, 0, 2)

	preferred := base
	defaults := betterFundDefaultCacheQuery()
	preferred.FeeFree7 = defaults.FeeFree7
	preferred.FeeFree30 = defaults.FeeFree30
	preferred.IncludeAClass = defaults.IncludeAClass
	preferred.OnlyAClass = defaults.OnlyAClass
	profiles = append(profiles, preferred)

	open := base
	unrestricted := betterFundUnrestrictedCacheQuery()
	open.FeeFree7 = unrestricted.FeeFree7
	open.FeeFree30 = unrestricted.FeeFree30
	open.IncludeAClass = unrestricted.IncludeAClass
	open.OnlyAClass = unrestricted.OnlyAClass
	if !sameBetterFundPreferenceProfile(preferred, open) {
		profiles = append(profiles, open)
	}

	return profiles
}

func betterFundCacheFallbackProfiles(query BetterFundQuery) []BetterFundQuery {
	fallbacks := make([]BetterFundQuery, 0, 2)

	unrestricted := query
	open := betterFundUnrestrictedCacheQuery()
	unrestricted.FeeFree7 = open.FeeFree7
	unrestricted.FeeFree30 = open.FeeFree30
	unrestricted.IncludeAClass = open.IncludeAClass
	unrestricted.OnlyAClass = open.OnlyAClass
	if !sameBetterFundPreferenceProfile(query, unrestricted) {
		fallbacks = append(fallbacks, unrestricted)
	}

	preferred := query
	defaults := betterFundDefaultCacheQuery()
	preferred.FeeFree7 = defaults.FeeFree7
	preferred.FeeFree30 = defaults.FeeFree30
	preferred.IncludeAClass = defaults.IncludeAClass
	preferred.OnlyAClass = defaults.OnlyAClass
	if !sameBetterFundPreferenceProfile(query, preferred) && !sameBetterFundPreferenceProfile(unrestricted, preferred) {
		fallbacks = append(fallbacks, preferred)
	}

	return fallbacks
}

func saveFundRecommendationCache(referenceCode string, refreshDate string, query BetterFundQuery, result *BetterFundResult) error {
	if result == nil {
		return nil
	}
	candidates := result.Candidates
	if len(candidates) > 30 {
		candidates = candidates[:30]
	}
	payload, err := json.Marshal(candidates)
	if err != nil {
		return err
	}

	cache := FundRecommendationCache{
		ReferenceCode:    strings.TrimSpace(referenceCode),
		RefreshDate:      refreshDate,
		SameTypeOnly:     query.SameTypeOnly,
		SameSubTypeOnly:  query.SameSubTypeOnly,
		Dimension:        strings.TrimSpace(query.Dimension),
		FeeFree7:         query.FeeFree7,
		FeeFree30:        query.FeeFree30,
		IncludeAClass:    query.IncludeAClass,
		OnlyAClass:       query.OnlyAClass,
		ScopeLabel:       result.ScopeLabel,
		SortLabel:        result.SortLabel,
		FallbackApplied:  result.FallbackApplied,
		ComparedUniverse: result.ComparedUniverse,
		UniverseTotal:    result.UniverseTotal,
		DataHint:         result.DataHint,
		CandidatesJSON:   string(payload),
	}

	var existing FundRecommendationCache
	err = db.Dao.Where(
		"reference_code = ? AND refresh_date = ? AND same_type_only = ? AND IFNULL(same_sub_type_only, 0) = ? AND dimension = ? AND fee_free7 = ? AND fee_free30 = ? AND include_a_class = ? AND IFNULL(only_a_class, 0) = ?",
		cache.ReferenceCode, cache.RefreshDate, cache.SameTypeOnly, cache.SameSubTypeOnly, cache.Dimension,
		cache.FeeFree7, cache.FeeFree30, cache.IncludeAClass, cache.OnlyAClass,
	).First(&existing).Error
	if err == nil {
		return db.Dao.Model(&FundRecommendationCache{}).
			Where("id = ?", existing.ID).
			Updates(map[string]any{
				"scope_label":       cache.ScopeLabel,
				"sort_label":        cache.SortLabel,
				"fallback_applied":  cache.FallbackApplied,
				"compared_universe": cache.ComparedUniverse,
				"universe_total":    cache.UniverseTotal,
				"data_hint":         cache.DataHint,
				"candidates_json":   cache.CandidatesJSON,
			}).Error
	}
	return db.Dao.Create(&cache).Error
}

func loadFundRecommendationCache(referenceCode string, query BetterFundQuery, refreshDate string) (*FundRecommendationCache, bool) {
	var cache FundRecommendationCache
	err := db.Dao.
		Where(
			"reference_code = ? AND same_type_only = ? AND IFNULL(same_sub_type_only, 0) = ? AND dimension = ? AND fee_free7 = ? AND fee_free30 = ? AND include_a_class = ? AND IFNULL(only_a_class, 0) = ?",
			strings.TrimSpace(referenceCode), query.SameTypeOnly, query.SameSubTypeOnly, strings.TrimSpace(query.Dimension),
			query.FeeFree7, query.FeeFree30, query.IncludeAClass, query.OnlyAClass,
		).
		Order("CASE WHEN refresh_date = '" + strings.TrimSpace(refreshDate) + "' THEN 0 ELSE 1 END").
		Order("refresh_date desc").
		Order("updated_at desc").
		First(&cache).Error
	if err != nil {
		return nil, false
	}
	return &cache, cache.RefreshDate == strings.TrimSpace(refreshDate)
}

func paginateBetterCandidates(candidates []BetterFundCandidate, page int, pageSize int) ([]BetterFundCandidate, int64) {
	total := int64(len(candidates))
	start := (page - 1) * pageSize
	if start > len(candidates) {
		start = len(candidates)
	}
	end := start + pageSize
	if end > len(candidates) {
		end = len(candidates)
	}
	return candidates[start:end], total
}

func finalizeBetterFundCandidates(candidates []BetterFundCandidate, universe betterCandidateUniverse) []BetterFundCandidate {
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
	return candidates
}

func betterFundCompositeScoreFromMetrics(metrics []BetterFundMetric, candidate bool) float64 {
	score := 0.0
	for _, metric := range metrics {
		var percentile *float64
		if candidate {
			percentile = metric.CandidatePercentile
		} else {
			percentile = metric.ReferencePercentile
		}
		if percentile == nil {
			continue
		}
		score += (*percentile / 100.0) * metric.Weight
	}
	return roundPercent(score)
}

func filterCandidatesStrongerThanReference(candidates []BetterFundCandidate) ([]BetterFundCandidate, bool) {
	if len(candidates) == 0 {
		return candidates, false
	}

	filtered := make([]BetterFundCandidate, 0, len(candidates))
	referenceBest := true
	for _, candidate := range candidates {
		candidateScore := betterFundCompositeScoreFromMetrics(candidate.Metrics, true)
		referenceScore := betterFundCompositeScoreFromMetrics(candidate.Metrics, false)
		if candidateScore > referenceScore {
			filtered = append(filtered, candidate)
			referenceBest = false
		}
	}
	return filtered, referenceBest
}

func buildBetterFundResultFromUniverse(reference data.FundBasic, watchlist bool, query BetterFundQuery, universe betterCandidateUniverse, refreshStatus FundRefreshStatus, dataHint string) *BetterFundResult {
	refItem := buildFundScreenerItem(reference, watchlist)
	candidates := buildBetterFundCandidates(reference, universe, query.Dimension)
	candidates = finalizeBetterFundCandidates(candidates, universe)
	candidates = applyBetterFundPreferenceFilters(candidates, query)
	comparedUniverse := len(candidates) + 1
	var referenceBest bool
	candidates, referenceBest = filterCandidatesStrongerThanReference(candidates)
	if referenceBest {
		dataHint = strings.TrimSpace(strings.Join([]string{strings.TrimSpace(dataHint), "当前产品已是当前筛选范围内的更优选择，暂无更优推荐。"}, " "))
	}
	pageItems, total := paginateBetterCandidates(candidates, query.Page, query.PageSize)
	return &BetterFundResult{
		Reference:        refItem,
		Candidates:       pageItems,
		Dimension:        query.Dimension,
		SortLabel:        betterSortLabel(query.Dimension),
		ScopeLabel:       universe.ScopeLabel,
		ComparedUniverse: comparedUniverse,
		UniverseTotal:    universe.UniverseTotal,
		RefreshedCount:   universe.RefreshedCount,
		NetworkRefresh:   universe.NetworkRefresh,
		FallbackApplied:  universe.FallbackApplied,
		DataHint:         dataHint,
		Total:            total,
		Page:             query.Page,
		PageSize:         query.PageSize,
		RefreshStatus:    refreshStatus,
	}
}

func buildCachedBetterFundResult(reference data.FundBasic, watchlist bool, query BetterFundQuery, cache *FundRecommendationCache, isToday bool) *BetterFundResult {
	if cache == nil {
		return nil
	}

	var candidates []BetterFundCandidate
	if strings.TrimSpace(cache.CandidatesJSON) != "" {
		_ = json.Unmarshal([]byte(cache.CandidatesJSON), &candidates)
	}
	candidates = applyBetterFundPreferenceFilters(candidates, query)
	comparedUniverse := cache.ComparedUniverse
	if comparedUniverse <= 1 {
		if parsed := parseComparedUniverseFromHint(cache.DataHint); parsed > 0 {
			comparedUniverse = parsed
		}
	}
	if comparedUniverse <= 0 {
		comparedUniverse = len(candidates) + 1
	}
	var referenceBest bool
	candidates, referenceBest = filterCandidatesStrongerThanReference(candidates)
	pageItems, total := paginateBetterCandidates(candidates, query.Page, query.PageSize)
	hint := strings.TrimSpace(cache.DataHint)
	if hint == "" {
		hint = "推荐结果来自后台已生成的推荐缓存。"
	}
	if !isToday {
		hint = "当前展示的是最近一次缓存结果，今天的推荐缓存还没有完成。"
	}
	if referenceBest {
		hint = strings.TrimSpace(strings.Join([]string{hint, "当前产品已是当前筛选范围内的更优选择，暂无更优推荐。"}, " "))
	}

	return &BetterFundResult{
		Reference:        buildFundScreenerItem(reference, watchlist),
		Candidates:       pageItems,
		Dimension:        query.Dimension,
		SortLabel:        defaultLabel(cache.SortLabel, betterSortLabel(query.Dimension)),
		ScopeLabel:       cache.ScopeLabel,
		ComparedUniverse: comparedUniverse,
		UniverseTotal:    cache.UniverseTotal,
		FallbackApplied:  cache.FallbackApplied,
		DataHint:         hint,
		Total:            total,
		Page:             query.Page,
		PageSize:         query.PageSize,
		RefreshStatus: FundRefreshStatus{
			LastRefreshHint: cache.UpdatedAt.Format("2006-01-02 15:04:05"),
		},
	}
}

func betterRecommendationCacheStale(cache *FundRecommendationCache, dimension string) bool {
	if cache == nil {
		return true
	}
	if strings.TrimSpace(cache.SortLabel) != betterSortLabel(dimension) {
		return true
	}
	if strings.TrimSpace(cache.CandidatesJSON) == "" {
		return false
	}

	var candidates []BetterFundCandidate
	if err := json.Unmarshal([]byte(cache.CandidatesJSON), &candidates); err != nil {
		return true
	}
	for _, candidate := range candidates {
		if len(candidate.Metrics) == 0 {
			continue
		}
		for _, metric := range candidate.Metrics {
			if metric.Key == "drawdown1" {
				return false
			}
		}
		return true
	}
	return false
}

func parseComparedUniverseFromHint(hint string) int {
	hint = strings.TrimSpace(hint)
	if hint == "" {
		return 0
	}
	patterns := []string{
		`有效样本\s*(\d+)\s*只`,
		`样本\s*(\d+)\s*只`,
	}
	for _, pattern := range patterns {
		matches := regexp.MustCompile(pattern).FindStringSubmatch(hint)
		if len(matches) != 2 {
			continue
		}
		value, err := strconv.Atoi(matches[1])
		if err == nil && value > 0 {
			return value
		}
	}
	return 0
}

func (s *Service) runFundRecommendationRefresh(today string, codes []string) {
	defer func() {
		if r := recover(); r != nil {
			logger.SugaredLogger.Errorf("fund recommendation refresh panic: %v", r)
			logger.SugaredLogger.Errorf("fund recommendation refresh stack: %s", string(debug.Stack()))
		}
		recommendationRefreshState.mu.Lock()
		recommendationRefreshState.refreshing = false
		recommendationRefreshState.currentCode = ""
		recommendationRefreshState.lastFinished = time.Now()
		recommendationRefreshState.mu.Unlock()
	}()

	activeCodes := dedupeCodes(codes)
	if len(activeCodes) == 0 {
		return
	}

	seedFundRecommendationProgress(today, activeCodes)
	var rows []FundRecommendationProgress
	db.Dao.Where("refresh_date = ? AND reference_code IN ?", today, activeCodes).Find(&rows)
	pendingCodes := make([]string, 0, len(rows))
	for _, row := range rows {
		if row.Status != recommendationStatusCompleted {
			pendingCodes = append(pendingCodes, strings.TrimSpace(row.ReferenceCode))
		}
	}
	if len(pendingCodes) == 0 {
		recommendationRefreshState.mu.Lock()
		recommendationRefreshState.progressTotal = 0
		recommendationRefreshState.progressNow = 0
		recommendationRefreshState.mu.Unlock()
		return
	}

	recommendationRefreshState.mu.Lock()
	recommendationRefreshState.progressTotal = int64(len(pendingCodes))
	recommendationRefreshState.progressNow = 0
	recommendationRefreshState.mu.Unlock()

	dimensions := []string{"balanced", "higher_return", "lower_drawdown"}
	for idx, code := range pendingCodes {
		trimmed := strings.TrimSpace(code)
		recommendationRefreshState.mu.Lock()
		recommendationRefreshState.currentCode = trimmed
		recommendationRefreshState.progressNow = int64(idx)
		recommendationRefreshState.mu.Unlock()

		upsertFundRecommendationProgressStatus(today, trimmed, map[string]any{
			"status":        recommendationStatusRunning,
			"current_phase": "refresh_reference",
			"message":       "正在补齐参考基金指标",
			"last_error":    "",
		})

		reference, err := dbQueryFundBasic(trimmed)
		if err != nil || reference == nil {
			if !s.refreshFundScreeningMetrics(trimmed) {
				upsertFundRecommendationProgressStatus(today, trimmed, map[string]any{
					"status":        recommendationStatusFailed,
					"current_phase": "refresh_reference",
					"last_error":    "参考基金指标刷新失败",
					"message":       "参考基金指标刷新失败",
				})
				continue
			}
			reference, err = dbQueryFundBasic(trimmed)
		}
		if err != nil || reference == nil {
			upsertFundRecommendationProgressStatus(today, trimmed, map[string]any{
				"status":        recommendationStatusFailed,
				"current_phase": "refresh_reference",
				"last_error":    "未找到参考基金基础数据",
				"message":       "未找到参考基金基础数据",
			})
			continue
		}

		if !isSameTradingDayString(reference.ScreenUpdatedAt, time.Now()) {
			s.refreshFundScreeningMetrics(trimmed)
			if refreshed, queryErr := dbQueryFundBasic(trimmed); queryErr == nil && refreshed != nil {
				reference = refreshed
			}
		}

		upsertFundRecommendationProgressStatus(today, trimmed, map[string]any{
			"status":        recommendationStatusRunning,
			"current_phase": "refresh_candidates",
			"message":       "正在补齐同类候选基金指标",
		})

		refCategory := classifyFundType(defaultLabel(reference.Type, reference.Name))
		exactUniverse, _ := s.refreshBetterFundUniverse(*reference, refCategory.Category, true, false)
		exactHint := betterDataHint(FundRefreshStatus{}, exactUniverse)

		query := BetterFundQuery{
			ReferenceCode:   trimmed,
			SameTypeOnly:    true,
			SameSubTypeOnly: false,
			Page:            1,
			PageSize:        30,
		}
		for _, dimension := range dimensions {
			query.Dimension = dimension
			for _, cacheQuery := range buildBetterFundCacheProfiles(query) {
				result := buildBetterFundResultFromUniverse(*reference, isFundWatchlisted(trimmed), cacheQuery, exactUniverse, FundRefreshStatus{}, exactHint)
				if err := saveFundRecommendationCache(trimmed, today, cacheQuery, result); err != nil {
					logger.SugaredLogger.Warnf("save exact recommendation cache failed for %s/%s: %v", trimmed, dimension, err)
				}
			}
		}

		subTypeUniverse, _ := s.refreshBetterFundUniverse(*reference, refCategory.Category, true, true)
		subTypeHint := betterDataHint(FundRefreshStatus{}, subTypeUniverse)
		query.SameSubTypeOnly = true
		for _, dimension := range dimensions {
			query.Dimension = dimension
			for _, cacheQuery := range buildBetterFundCacheProfiles(query) {
				result := buildBetterFundResultFromUniverse(*reference, isFundWatchlisted(trimmed), cacheQuery, subTypeUniverse, FundRefreshStatus{}, subTypeHint)
				if err := saveFundRecommendationCache(trimmed, today, cacheQuery, result); err != nil {
					logger.SugaredLogger.Warnf("save subtype recommendation cache failed for %s/%s: %v", trimmed, dimension, err)
				}
			}
		}

		wideUniverse := loadBetterFundUniverseDetailed(*reference, refCategory.Category, false, false)
		wideHint := "推荐缓存基于今日已落库的基金指标生成。"
		query.SameTypeOnly = false
		query.SameSubTypeOnly = false
		for _, dimension := range dimensions {
			query.Dimension = dimension
			for _, cacheQuery := range buildBetterFundCacheProfiles(query) {
				result := buildBetterFundResultFromUniverse(*reference, isFundWatchlisted(trimmed), cacheQuery, wideUniverse, FundRefreshStatus{}, wideHint)
				if err := saveFundRecommendationCache(trimmed, today, cacheQuery, result); err != nil {
					logger.SugaredLogger.Warnf("save wide recommendation cache failed for %s/%s: %v", trimmed, dimension, err)
				}
			}
		}

		upsertFundRecommendationProgressStatus(today, trimmed, map[string]any{
			"status":            recommendationStatusCompleted,
			"current_phase":     "",
			"message":           "今日推荐缓存已完成",
			"compared_universe": exactUniverse.ComparedUniverse,
			"universe_total":    exactUniverse.UniverseTotal,
			"last_error":        "",
		})

		recommendationRefreshState.mu.Lock()
		recommendationRefreshState.progressNow = int64(idx + 1)
		recommendationRefreshState.mu.Unlock()
	}
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
	if query.SameSubTypeOnly {
		query.SameTypeOnly = true
	}
	if query.OnlyAClass {
		query.IncludeAClass = true
	}

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
	universe := loadBetterFundUniverseDetailed(reference, refCategory.Category, query.SameTypeOnly, query.SameSubTypeOnly)
	if query.NetworkRefresh {
		refreshedUniverse, refreshStats := s.refreshBetterFundUniverse(reference, refCategory.Category, query.SameTypeOnly, query.SameSubTypeOnly)
		refreshedUniverse.NetworkRefresh = refreshStats.NetworkRefresh
		refreshedUniverse.RefreshedCount = refreshStats.RefreshedCount
		if refreshStats.UniverseTotal > 0 {
			refreshedUniverse.UniverseTotal = refreshStats.UniverseTotal
		}
		refreshedUniverse.Limited = refreshStats.Limited
		universe = refreshedUniverse
	}
	candidates := buildBetterFundCandidates(reference, universe, query.Dimension)
	if false && query.SameTypeOnly && len(candidates) == 0 {
		candidates = loadBetterFundCandidates(reference, refCategory.Category, query.Dimension, false, false)
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
		UniverseTotal:    universe.UniverseTotal,
		RefreshedCount:   universe.RefreshedCount,
		NetworkRefresh:   universe.NetworkRefresh,
		FallbackApplied:  universe.FallbackApplied,
		DataHint:         betterDataHint(refreshStatus, universe),
		Total:            total,
		Page:             query.Page,
		PageSize:         query.PageSize,
		RefreshStatus:    refreshStatus,
	}
}

func (s *Service) GetBetterFundsCached(query BetterFundQuery) *BetterFundResult {
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
	if query.SameSubTypeOnly {
		query.SameTypeOnly = true
	}
	if query.OnlyAClass {
		query.IncludeAClass = true
	}
	today := time.Now().Format("2006-01-02")

	if query.NetworkRefresh {
		s.GetFundRecommendationRefreshStatus(true)
	}

	var reference data.FundBasic
	if err := db.Dao.Where("code = ?", code).First(&reference).Error; err != nil {
		return nil
	}

	if query.NetworkRefresh {
		refCategory := classifyFundType(defaultLabel(reference.Type, reference.Name))
		refreshStatus := s.buildFundRefreshStatus(s.EnsureFundUniverse())
		universe, _ := s.refreshBetterFundUniverse(reference, refCategory.Category, query.SameTypeOnly, query.SameSubTypeOnly)
		result := buildBetterFundResultFromUniverse(reference, isFundWatchlisted(code), query, universe, refreshStatus, betterDataHint(refreshStatus, universe))
		if err := saveFundRecommendationCache(code, today, query, result); err != nil {
			logger.SugaredLogger.Warnf("save refreshed recommendation cache failed for %s/%s: %v", code, query.Dimension, err)
		}
		return result
	}

	if cache, isToday := loadFundRecommendationCache(code, query, today); cache != nil && !betterRecommendationCacheStale(cache, query.Dimension) {
		return buildCachedBetterFundResult(reference, isFundWatchlisted(code), query, cache, isToday)
	}
	for _, fallbackQuery := range betterFundCacheFallbackProfiles(query) {
		if cache, isToday := loadFundRecommendationCache(code, fallbackQuery, today); cache != nil && !betterRecommendationCacheStale(cache, fallbackQuery.Dimension) {
			return buildCachedBetterFundResult(reference, isFundWatchlisted(code), query, cache, isToday)
		}
	}

	refCategory := classifyFundType(defaultLabel(reference.Type, reference.Name))
	refreshStatus := s.buildFundRefreshStatus(s.EnsureFundUniverse())
	universe := loadBetterFundUniverseDetailed(reference, refCategory.Category, query.SameTypeOnly, query.SameSubTypeOnly)
	dataHint := "今日推荐缓存尚未完成，当前先展示本地已落库的基金指标结果。"
	if universe.ComparedUniverse <= 1 {
		dataHint = "当前推荐缓存尚未完成，且本地可比较样本还不足，请等待后台推荐检索完成。"
	}
	return buildBetterFundResultFromUniverse(reference, isFundWatchlisted(code), query, universe, refreshStatus, dataHint)
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
		if err == nil && (!isSameTradingDayString(basic.ScreenUpdatedAt, now) || basic.MaxDrawdown1 == nil) {
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

func normalizeBetterFundSubType(value string) string {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	if normalized == "" {
		return ""
	}
	replacer := strings.NewReplacer(
		" ", "",
		"\t", "",
		"\r", "",
		"\n", "",
		"（", "",
		"）", "",
		"(", "",
		")", "",
		"-", "",
		"_", "",
		"·", "",
		"，", "",
		",", "",
		"跟踪标的", "",
		"标的", "",
		"指数", "",
		"ETF联接", "",
		"ETF", "",
		"联接基金", "",
		"联接", "",
		"增强", "",
		"基金", "",
	)
	return strings.TrimSpace(replacer.Replace(normalized))
}

func betterFundSubTypeLabel(fund data.FundBasic) string {
	return strings.TrimSpace(fund.TrackingTarget)
}

func betterFundSubTypeKey(fund data.FundBasic) string {
	return normalizeBetterFundSubType(fund.TrackingTarget)
}

func filterBetterFundBasicsBySubType(reference data.FundBasic, basics []data.FundBasic) []data.FundBasic {
	refKey := betterFundSubTypeKey(reference)
	if refKey == "" || len(basics) == 0 {
		return basics
	}

	filtered := make([]data.FundBasic, 0, len(basics))
	for _, basic := range basics {
		if betterFundSubTypeKey(basic) == refKey {
			filtered = append(filtered, basic)
		}
	}
	return filtered
}

func fundSupportsSubTypeFilter(fund data.FundBasic) bool {
	normalized := strings.ToUpper(strings.TrimSpace(fund.Type))
	if normalized == "" {
		return strings.TrimSpace(fund.TrackingTarget) != ""
	}
	return strings.Contains(normalized, "指数") ||
		strings.Contains(normalized, "ETF") ||
		strings.Contains(normalized, "联接") ||
		strings.Contains(normalized, "增强") ||
		strings.TrimSpace(fund.TrackingTarget) != ""
}

func normalizeBetterFundTypeKey(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if parts := strings.SplitN(value, "|", 2); len(parts) > 0 {
		value = strings.TrimSpace(parts[0])
	}
	return value
}

func fundShareClassSuffix(name string) string {
	normalized := strings.ToUpper(strings.TrimSpace(name))
	if normalized == "" {
		return ""
	}
	match := regexp.MustCompile(`([A-Z])$`).FindStringSubmatch(normalized)
	if len(match) != 2 {
		return ""
	}
	return match[1]
}

func queryOnlineBetterFundSeeds(reference data.FundBasic, refCategory string, exactType bool) []data.FundCatalogItem {
	catalog, err := data.NewFundApi().GetEastmoneyFundCatalog(false)
	if err != nil || len(catalog) == 0 {
		return []data.FundCatalogItem{}
	}

	refType := normalizeBetterFundTypeKey(reference.Type)
	items := make([]data.FundCatalogItem, 0, len(catalog))
	for _, item := range catalog {
		if strings.TrimSpace(item.Code) == "" || strings.TrimSpace(item.Code) == strings.TrimSpace(reference.Code) {
			continue
		}
		itemType := normalizeBetterFundTypeKey(item.Type)
		if exactType && refType != "" {
			if itemType != refType {
				continue
			}
		} else {
			classification := classifyFundType(defaultLabel(itemType, item.Name))
			if classification.Category != refCategory {
				continue
			}
		}
		items = append(items, data.FundCatalogItem{
			Code: strings.TrimSpace(item.Code),
			Name: strings.TrimSpace(item.Name),
			Type: itemType,
		})
	}
	return items
}

func loadFundBasicsMapByCodes(codes []string) map[string]data.FundBasic {
	if len(codes) == 0 {
		return map[string]data.FundBasic{}
	}

	var basics []data.FundBasic
	db.Dao.Where("code IN ?", dedupeCodes(codes)).Find(&basics)
	result := make(map[string]data.FundBasic, len(basics))
	for _, basic := range basics {
		result[strings.TrimSpace(basic.Code)] = basic
	}
	return result
}

func ensureFundBasicShells(entries []data.FundCatalogItem) {
	if len(entries) == 0 {
		return
	}

	codes := make([]string, 0, len(entries))
	for _, item := range entries {
		if trimmed := strings.TrimSpace(item.Code); trimmed != "" {
			codes = append(codes, trimmed)
		}
	}
	if len(codes) == 0 {
		return
	}

	var existing []string
	db.Dao.Model(&data.FundBasic{}).Where("code IN ?", dedupeCodes(codes)).Pluck("code", &existing)
	existingSet := make(map[string]struct{}, len(existing))
	for _, code := range existing {
		existingSet[strings.TrimSpace(code)] = struct{}{}
	}

	for _, item := range entries {
		code := strings.TrimSpace(item.Code)
		if code == "" {
			continue
		}
		if _, ok := existingSet[code]; ok {
			continue
		}
		db.Dao.Create(&data.FundBasic{
			Code: code,
			Name: strings.TrimSpace(item.Name),
			Type: normalizeBetterFundTypeKey(item.Type),
		})
	}
}

func prioritizeBetterFundSeedEntries(reference data.FundBasic, seeds []data.FundCatalogItem, basics map[string]data.FundBasic) []data.FundCatalogItem {
	if len(seeds) <= 1 {
		return seeds
	}

	refShare := fundShareClassSuffix(reference.Name)
	preferredProfile := betterFundDefaultCacheQuery()
	prioritized := append([]data.FundCatalogItem(nil), seeds...)
	sort.SliceStable(prioritized, func(i, j int) bool {
		left := prioritized[i]
		right := prioritized[j]

		leftBasic, leftExists := basics[left.Code]
		rightBasic, rightExists := basics[right.Code]

		leftScore := 0
		rightScore := 0
		if leftExists {
			leftScore -= 4
			if isUsableBetterFundCandidate(leftBasic) {
				leftScore -= 6
			}
		}
		if rightExists {
			rightScore -= 4
			if isUsableBetterFundCandidate(rightBasic) {
				rightScore -= 6
			}
		}
		if refShare != "" {
			if fundShareClassSuffix(left.Name) == refShare {
				leftScore -= 3
			}
			if fundShareClassSuffix(right.Name) == refShare {
				rightScore -= 3
			}
		}
		if !preferredProfile.IncludeAClass {
			if isAFundShareClass(left.Name) {
				leftScore += 6
			} else {
				leftScore -= 2
			}
			if isAFundShareClass(right.Name) {
				rightScore += 6
			} else {
				rightScore -= 2
			}
		}
		if preferredProfile.FeeFree7 || preferredProfile.FeeFree30 {
			if leftExists {
				switch {
				case preferredProfile.FeeFree7 && leftBasic.RedeemFeeFreeDays > 0 && leftBasic.RedeemFeeFreeDays <= 7:
					leftScore -= 5
				case preferredProfile.FeeFree30 && leftBasic.RedeemFeeFreeDays > 7 && leftBasic.RedeemFeeFreeDays <= 30:
					leftScore -= 3
				case leftBasic.RedeemFeeFreeDays > 0:
					leftScore += 3
				default:
					leftScore += 1
				}
			}
			if rightExists {
				switch {
				case preferredProfile.FeeFree7 && rightBasic.RedeemFeeFreeDays > 0 && rightBasic.RedeemFeeFreeDays <= 7:
					rightScore -= 5
				case preferredProfile.FeeFree30 && rightBasic.RedeemFeeFreeDays > 7 && rightBasic.RedeemFeeFreeDays <= 30:
					rightScore -= 3
				case rightBasic.RedeemFeeFreeDays > 0:
					rightScore += 3
				default:
					rightScore += 1
				}
			}
		}
		if leftScore == rightScore {
			return left.Code < right.Code
		}
		return leftScore < rightScore
	})
	return prioritized
}

func loadBetterFundCandidates(reference data.FundBasic, refCategory string, dimension string, sameTypeOnly bool, sameSubTypeOnly bool) []BetterFundCandidate {
	basics := queryBetterFundBasics(reference, refCategory, sameTypeOnly && strings.TrimSpace(reference.Type) != "", sameSubTypeOnly)

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

func loadBetterFundUniverse(reference data.FundBasic, refCategory string, sameTypeOnly bool, sameSubTypeOnly bool) betterCandidateUniverse {
	if sameSubTypeOnly {
		sameTypeOnly = true
	}
	if sameSubTypeOnly {
		subTypeLabel := betterFundSubTypeLabel(reference)
		if subTypeLabel != "" {
			basics := queryBetterFundBasics(reference, refCategory, true, true)
			return betterCandidateUniverse{
				Basics:           basics,
				ScopeLabel:       "同子类型精确匹配：" + subTypeLabel,
				ComparedUniverse: len(basics) + 1,
			}
		}

		basics := queryBetterFundBasics(reference, refCategory, true, false)
		return betterCandidateUniverse{
			Basics:           basics,
			ScopeLabel:       "参考基金暂无明确子类型，已按同类型匹配：" + strings.TrimSpace(reference.Type),
			ComparedUniverse: len(basics) + 1,
		}
	}
	if sameTypeOnly && strings.TrimSpace(reference.Type) != "" {
		basics := queryBetterFundBasics(reference, refCategory, true, false)
		if len(basics) > 0 {
			return betterCandidateUniverse{
				Basics:           basics,
				ScopeLabel:       "同类型精确匹配：" + strings.TrimSpace(reference.Type),
				ComparedUniverse: len(basics) + 1,
			}
		}
		basics = queryBetterFundBasics(reference, refCategory, false, false)
		return betterCandidateUniverse{
			Basics:           basics,
			ScopeLabel:       "同类型暂无更优，已放宽到同大类：" + betterCategoryLabel(refCategory),
			FallbackApplied:  true,
			ComparedUniverse: len(basics) + 1,
		}
	}

	basics := queryBetterFundBasics(reference, refCategory, false, false)
	return betterCandidateUniverse{
		Basics:           basics,
		ScopeLabel:       "同大类匹配：" + betterCategoryLabel(refCategory),
		ComparedUniverse: len(basics) + 1,
	}
}

func queryBetterFundBasics(reference data.FundBasic, refCategory string, exactType bool, sameSubTypeOnly bool) []data.FundBasic {
	if exactType {
		seedEntries := queryOnlineBetterFundSeeds(reference, refCategory, true)
		if len(seedEntries) > 0 {
			codes := make([]string, 0, len(seedEntries))
			for _, item := range seedEntries {
				if trimmed := strings.TrimSpace(item.Code); trimmed != "" {
					codes = append(codes, trimmed)
				}
			}
			basicMap := loadFundBasicsMapByCodes(codes)
			basics := make([]data.FundBasic, 0, len(seedEntries))
			for _, item := range seedEntries {
				if basic, ok := basicMap[item.Code]; ok && strings.TrimSpace(basic.ScreenUpdatedAt) != "" {
					basics = append(basics, basic)
				}
			}
			if sameSubTypeOnly {
				basics = filterBetterFundBasicsBySubType(reference, basics)
			}
			return basics
		}
	}

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
	if sameSubTypeOnly {
		basics = filterBetterFundBasicsBySubType(reference, basics)
	}
	return basics
}

func loadBetterFundUniverseDetailed(reference data.FundBasic, refCategory string, sameTypeOnly bool, sameSubTypeOnly bool) betterCandidateUniverse {
	universe := loadBetterFundUniverse(reference, refCategory, sameTypeOnly, sameSubTypeOnly)
	if sameSubTypeOnly {
		if betterFundSubTypeKey(reference) != "" {
			universe.UniverseTotal = countBetterFundUniverse(reference, refCategory, true, true) + 1
			return universe
		}
		universe.UniverseTotal = countBetterFundUniverse(reference, refCategory, true, false) + 1
		return universe
	}
	if sameTypeOnly && strings.TrimSpace(reference.Type) != "" {
		universe.UniverseTotal = countBetterFundUniverse(reference, refCategory, true, false) + 1
		if len(universe.Basics) == 0 {
			universe.UniverseTotal = countBetterFundUniverse(reference, refCategory, false, false) + 1
		}
		return universe
	}
	universe.UniverseTotal = countBetterFundUniverse(reference, refCategory, false, false) + 1
	return universe
}

func queryBetterFundUniverseSeeds(reference data.FundBasic, refCategory string, exactType bool, sameSubTypeOnly bool, limit int) []data.FundBasic {
	if exactType {
		seedEntries := queryOnlineBetterFundSeeds(reference, refCategory, true)
		if len(seedEntries) > 0 {
			if limit > 0 && len(seedEntries) > limit {
				seedEntries = seedEntries[:limit]
			}
			codes := make([]string, 0, len(seedEntries))
			for _, item := range seedEntries {
				codes = append(codes, item.Code)
			}
			basicMap := loadFundBasicsMapByCodes(codes)
			basics := make([]data.FundBasic, 0, len(seedEntries))
			for _, item := range seedEntries {
				if basic, ok := basicMap[item.Code]; ok {
					basics = append(basics, basic)
				}
			}
			if sameSubTypeOnly {
				basics = filterBetterFundBasicsBySubType(reference, basics)
			}
			return basics
		}
	}

	dbQuery := db.Dao.Model(&data.FundBasic{}).
		Where("code <> ?", reference.Code)
	if exactType && strings.TrimSpace(reference.Type) != "" {
		dbQuery = dbQuery.Where("type = ?", reference.Type)
	} else {
		dbQuery = applyFundCategoryDBFilter(dbQuery, refCategory)
	}
	if limit > 0 {
		dbQuery = dbQuery.Limit(limit)
	}

	var basics []data.FundBasic
	dbQuery.Order("updated_at desc").Order("code asc").Find(&basics)
	if sameSubTypeOnly {
		basics = filterBetterFundBasicsBySubType(reference, basics)
	}
	if limit > 0 && len(basics) > limit {
		basics = basics[:limit]
	}
	return basics
}

func countBetterFundUniverse(reference data.FundBasic, refCategory string, exactType bool, sameSubTypeOnly bool) int {
	if exactType {
		seedEntries := queryOnlineBetterFundSeeds(reference, refCategory, true)
		if sameSubTypeOnly {
			basics := queryBetterFundBasics(reference, refCategory, true, true)
			return len(basics)
		}
		if len(seedEntries) > 0 {
			return len(seedEntries)
		}
	}

	if sameSubTypeOnly {
		return len(queryBetterFundBasics(reference, refCategory, exactType, true))
	}
	var count int64
	dbQuery := db.Dao.Model(&data.FundBasic{}).
		Where("code <> ?", reference.Code)
	if exactType && strings.TrimSpace(reference.Type) != "" {
		dbQuery = dbQuery.Where("type = ?", reference.Type)
	} else {
		dbQuery = applyFundCategoryDBFilter(dbQuery, refCategory)
	}
	dbQuery.Count(&count)
	return int(count)
}

func parseFundEstablishmentDate(value string, location *time.Location) (time.Time, bool) {
	text := strings.TrimSpace(value)
	if text == "" {
		return time.Time{}, false
	}
	if location == nil {
		location = time.Local
	}
	layouts := []string{"2006-01-02", "2006/01/02", "2006.01.02"}
	for _, layout := range layouts {
		parsed, err := time.ParseInLocation(layout, text, location)
		if err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func fundHasHistoryForWindow(basic data.FundBasic, today time.Time, months int) bool {
	if months <= 0 {
		return true
	}
	established, ok := parseFundEstablishmentDate(basic.Establishment, today.Location())
	if !ok {
		return true
	}
	return !established.After(today.AddDate(0, -months, 0))
}

func isAlmostZeroMetric(value *float64) bool {
	if value == nil {
		return false
	}
	return math.Abs(*value) < 1e-9
}

func allowsMissingSharpeForFlatFund(basic data.FundBasic) bool {
	return basic.Sharpe12 == nil &&
		isAlmostZeroMetric(basic.NetGrowth3) &&
		isAlmostZeroMetric(basic.NetGrowth6) &&
		isAlmostZeroMetric(basic.NetGrowth12) &&
		isAlmostZeroMetric(basic.MaxDrawdown3) &&
		isAlmostZeroMetric(basic.MaxDrawdown6) &&
		isAlmostZeroMetric(basic.MaxDrawdown12)
}

func allowsMissingCalmarForZeroDrawdown(basic data.FundBasic) bool {
	return basic.Calmar12 == nil && isAlmostZeroMetric(basic.MaxDrawdown12)
}

func shouldRefreshBetterFundCandidate(basic data.FundBasic, today time.Time) bool {
	if !isSameTradingDayString(basic.ScreenUpdatedAt, today) {
		return true
	}
	require1M := fundHasHistoryForWindow(basic, today, 1)
	require3M := fundHasHistoryForWindow(basic, today, 3)
	require6M := fundHasHistoryForWindow(basic, today, 6)
	require12M := fundHasHistoryForWindow(basic, today, 12)

	if fundSupportsSubTypeFilter(basic) && betterFundSubTypeKey(basic) == "" {
		return true
	}
	if basic.RedeemFeeFreeDays <= 0 {
		return true
	}
	if require1M && (basic.StageRank1M <= 0 || basic.StageRank1MTotal <= 0) {
		return true
	}
	if require3M && (basic.NetGrowth3 == nil || basic.MaxDrawdown3 == nil || basic.StageRank3M <= 0 || basic.StageRank3MTotal <= 0) {
		return true
	}
	if require6M && (basic.NetGrowth6 == nil || basic.MaxDrawdown6 == nil || basic.StageRank6M <= 0 || basic.StageRank6MTotal <= 0) {
		return true
	}
	if require12M {
		if basic.NetGrowth12 == nil || basic.MaxDrawdown12 == nil || basic.StageRank12M <= 0 || basic.StageRank12MTotal <= 0 {
			return true
		}
		if basic.Sharpe12 == nil && !allowsMissingSharpeForFlatFund(basic) {
			return true
		}
		if basic.Calmar12 == nil && !allowsMissingCalmarForZeroDrawdown(basic) {
			return true
		}
	}
	return false
}

func isUsableBetterFundCandidate(basic data.FundBasic) bool {
	return basic.NetGrowth3 != nil &&
		basic.NetGrowth6 != nil &&
		basic.NetGrowth12 != nil &&
		basic.MaxDrawdown3 != nil &&
		basic.MaxDrawdown6 != nil &&
		basic.MaxDrawdown12 != nil &&
		basic.Sharpe12 != nil &&
		basic.Calmar12 != nil
}

func buildBetterUniverseFromSeedEntries(reference data.FundBasic, entries []data.FundCatalogItem, sameSubTypeOnly bool) betterCandidateUniverse {
	now := time.Now()
	codes := make([]string, 0, len(entries))
	for _, item := range entries {
		if trimmed := strings.TrimSpace(item.Code); trimmed != "" {
			codes = append(codes, trimmed)
		}
	}
	basicMap := loadFundBasicsMapByCodes(codes)
	basics := make([]data.FundBasic, 0, len(entries))
	for _, item := range entries {
		basic, ok := basicMap[item.Code]
		if !ok || strings.TrimSpace(basic.ScreenUpdatedAt) == "" || shouldRefreshBetterFundCandidate(basic, now) {
			continue
		}
		basics = append(basics, basic)
	}
	if sameSubTypeOnly {
		basics = filterBetterFundBasicsBySubType(reference, basics)
	}
	scopeLabel := "同类型精确匹配：" + strings.TrimSpace(reference.Type)
	if sameSubTypeOnly {
		if subTypeLabel := betterFundSubTypeLabel(reference); subTypeLabel != "" {
			scopeLabel = "同子类型精确匹配：" + subTypeLabel
		} else {
			scopeLabel = "参考基金暂无明确子类型，已按同类型匹配：" + strings.TrimSpace(reference.Type)
		}
	}
	return betterCandidateUniverse{
		Basics:           basics,
		ScopeLabel:       scopeLabel,
		ComparedUniverse: len(basics) + 1,
		UniverseTotal:    len(entries) + 1,
	}
}

func hasEnoughBetterCandidatesForDailyCache(reference data.FundBasic, universe betterCandidateUniverse) bool {
	if len(universe.Basics) == 0 {
		return false
	}
	defaultQuery := betterFundDefaultCacheQuery()
	for _, dimension := range []string{"balanced", "higher_return", "lower_drawdown"} {
		defaultQuery.Dimension = dimension
		candidates := buildBetterFundCandidates(reference, universe, dimension)
		candidates = finalizeBetterFundCandidates(candidates, universe)
		candidates = applyBetterFundPreferenceFilters(candidates, defaultQuery)
		candidates, _ = filterCandidatesStrongerThanReference(candidates)
		if len(candidates) < betterFundTargetPerDimension {
			return false
		}
	}
	return true
}

func (s *Service) refreshBetterFundUniverse(reference data.FundBasic, refCategory string, sameTypeOnly bool, sameSubTypeOnly bool) (betterCandidateUniverse, betterUniverseRefreshStats) {
	stats := betterUniverseRefreshStats{NetworkRefresh: true}
	today := time.Now()

	if sameSubTypeOnly {
		sameTypeOnly = true
	}
	exactType := sameTypeOnly && strings.TrimSpace(reference.Type) != ""
	if exactType {
		seedEntries := queryOnlineBetterFundSeeds(reference, refCategory, true)
		if len(seedEntries) > 0 {
			stats.UniverseTotal = len(seedEntries) + 1
			localBasics := loadFundBasicsMapByCodes(func() []string {
				codes := make([]string, 0, len(seedEntries))
				for _, item := range seedEntries {
					codes = append(codes, item.Code)
				}
				return codes
			}())
			prioritized := prioritizeBetterFundSeedEntries(reference, seedEntries, localBasics)
			initialUniverse := buildBetterUniverseFromSeedEntries(reference, seedEntries, sameSubTypeOnly)
			if hasEnoughBetterCandidatesForDailyCache(reference, initialUniverse) {
				initialUniverse.NetworkRefresh = true
				initialUniverse.UniverseTotal = stats.UniverseTotal
				return initialUniverse, stats
			}

			pendingEntries := make([]data.FundCatalogItem, 0, len(prioritized))
			for _, item := range prioritized {
				basic, ok := localBasics[item.Code]
				if ok && isUsableBetterFundCandidate(basic) && !shouldRefreshBetterFundCandidate(basic, today) {
					continue
				}
				pendingEntries = append(pendingEntries, item)
			}

			probeEntries := pendingEntries
			if len(probeEntries) > betterFundRefreshProbeBudget {
				probeEntries = probeEntries[:betterFundRefreshProbeBudget]
				stats.Limited = true
			}
			ensureFundBasicShells(probeEntries)

			for _, item := range probeEntries {
				if s.refreshFundScreeningMetrics(item.Code) {
					stats.RefreshedCount++
				}
				currentUniverse := buildBetterUniverseFromSeedEntries(reference, seedEntries, sameSubTypeOnly)
				if hasEnoughBetterCandidatesForDailyCache(reference, currentUniverse) {
					currentUniverse.NetworkRefresh = true
					currentUniverse.RefreshedCount = stats.RefreshedCount
					currentUniverse.UniverseTotal = stats.UniverseTotal
					currentUniverse.Limited = stats.Limited
					return currentUniverse, stats
				}
			}

			universe := buildBetterUniverseFromSeedEntries(reference, seedEntries, sameSubTypeOnly)
			universe.NetworkRefresh = true
			universe.RefreshedCount = stats.RefreshedCount
			universe.UniverseTotal = stats.UniverseTotal
			universe.Limited = stats.Limited
			if len(universe.Basics) > 0 {
				return universe, stats
			}
		}
	}

	seeds := queryBetterFundUniverseSeeds(reference, refCategory, exactType, sameSubTypeOnly, 0)
	if exactType && len(seeds) == 0 {
		exactType = false
		seeds = queryBetterFundUniverseSeeds(reference, refCategory, false, false, 0)
	}

	stats.UniverseTotal = len(seeds) + 1
	if len(seeds) == 0 {
		universe := loadBetterFundUniverseDetailed(reference, refCategory, sameTypeOnly, sameSubTypeOnly)
		universe.NetworkRefresh = true
		universe.UniverseTotal = stats.UniverseTotal
		return universe, stats
	}

	targetSeeds := seeds
	if len(targetSeeds) > betterFundRefreshCap {
		targetSeeds = targetSeeds[:betterFundRefreshCap]
		stats.Limited = true
	}

	targetUsable := 80
	if !sameTypeOnly {
		targetUsable = 120
	}

	usableCount := 0
	pendingCodes := make([]string, 0, len(targetSeeds))
	for _, basic := range targetSeeds {
		if isUsableBetterFundCandidate(basic) && !shouldRefreshBetterFundCandidate(basic, today) {
			usableCount++
			continue
		}
		pendingCodes = append(pendingCodes, basic.Code)
	}

	for _, code := range pendingCodes {
		if usableCount >= targetUsable {
			break
		}
		if s.refreshFundScreeningMetrics(code) {
			stats.RefreshedCount++
			if refreshed, err := dbQueryFundBasic(code); err == nil && refreshed != nil && isUsableBetterFundCandidate(*refreshed) {
				usableCount++
			}
		}
	}

	universe := loadBetterFundUniverseDetailed(reference, refCategory, sameTypeOnly, sameSubTypeOnly)
	universe.NetworkRefresh = true
	universe.RefreshedCount = stats.RefreshedCount
	universe.UniverseTotal = stats.UniverseTotal
	universe.Limited = stats.Limited
	return universe, stats
}

func buildBetterFundCandidates(reference data.FundBasic, universe betterCandidateUniverse, dimension string) []BetterFundCandidate {
	if len(universe.Basics) == 0 {
		return []BetterFundCandidate{}
	}

	specs := betterMetricSpecsForDimensionV2(dimension)
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

func betterMetricSpecsForDimensionV2(dimension string) []betterMetricSpec {
	specs := []betterMetricSpec{
		{Key: "growth1", Label: "近1月收益", Better: "higher", Weight: 0.90, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.NetGrowth1 }},
		{Key: "drawdown1", Label: "近1月最大回撤", Better: "lower", Weight: 0.85, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.MaxDrawdown1 }},
		{Key: "growth3", Label: "近3月收益", Better: "higher", Weight: 1.20, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.NetGrowth3 }},
		{Key: "drawdown3", Label: "近3月最大回撤", Better: "lower", Weight: 1.10, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.MaxDrawdown3 }},
		{Key: "growth6", Label: "近6月收益", Better: "higher", Weight: 0.95, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.NetGrowth6 }},
		{Key: "growth12", Label: "近1年收益", Better: "higher", Weight: 0.80, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.NetGrowth12 }},
		{Key: "drawdown6", Label: "近6月最大回撤", Better: "lower", Weight: 0.80, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.MaxDrawdown6 }},
		{Key: "drawdown12", Label: "近1年最大回撤", Better: "lower", Weight: 0.65, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.MaxDrawdown12 }},
		{Key: "sharpe12", Label: "近1年夏普", Better: "higher", Weight: 0.65, Format: "ratio", ValueOf: func(item data.FundBasic) *float64 { return item.Sharpe12 }},
		{Key: "calmar12", Label: "Calmar", Better: "higher", Weight: 0.60, Format: "ratio", ValueOf: func(item data.FundBasic) *float64 { return item.Calmar12 }},
	}

	switch normalizeBetterFundDimension(dimension) {
	case "lower_drawdown":
		return []betterMetricSpec{
			{Key: "drawdown1", Label: "近1月最大回撤", Better: "lower", Weight: 1.25, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.MaxDrawdown1 }},
			{Key: "drawdown3", Label: "近3月最大回撤", Better: "lower", Weight: 1.40, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.MaxDrawdown3 }},
			{Key: "drawdown6", Label: "近6月最大回撤", Better: "lower", Weight: 1.05, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.MaxDrawdown6 }},
			{Key: "drawdown12", Label: "近1年最大回撤", Better: "lower", Weight: 0.80, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.MaxDrawdown12 }},
			{Key: "volatility12", Label: "近1年波动", Better: "lower", Weight: 0.55, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.Volatility12 }},
			{Key: "sharpe12", Label: "近1年夏普", Better: "higher", Weight: 0.40, Format: "ratio", ValueOf: func(item data.FundBasic) *float64 { return item.Sharpe12 }},
			{Key: "growth3", Label: "近3月收益", Better: "higher", Weight: 0.35, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.NetGrowth3 }},
			{Key: "growth6", Label: "近6月收益", Better: "higher", Weight: 0.25, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.NetGrowth6 }},
		}
	case "higher_return":
		return []betterMetricSpec{
			{Key: "growth1", Label: "近1月收益", Better: "higher", Weight: 0.95, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.NetGrowth1 }},
			{Key: "growth3", Label: "近3月收益", Better: "higher", Weight: 1.25, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.NetGrowth3 }},
			{Key: "growth6", Label: "近6月收益", Better: "higher", Weight: 0.95, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.NetGrowth6 }},
			{Key: "growth12", Label: "近1年收益", Better: "higher", Weight: 0.75, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.NetGrowth12 }},
			{Key: "drawdown1", Label: "近1月最大回撤", Better: "lower", Weight: 0.45, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.MaxDrawdown1 }},
			{Key: "drawdown3", Label: "近3月最大回撤", Better: "lower", Weight: 0.55, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.MaxDrawdown3 }},
			{Key: "drawdown6", Label: "近6月最大回撤", Better: "lower", Weight: 0.25, Format: "percent", ValueOf: func(item data.FundBasic) *float64 { return item.MaxDrawdown6 }},
			{Key: "sharpe12", Label: "近1年夏普", Better: "higher", Weight: 0.35, Format: "ratio", ValueOf: func(item data.FundBasic) *float64 { return item.Sharpe12 }},
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
		return "按风险得分排序：近1月、近3月回撤权重更高，再参考近6月回撤、波动和收益"
	case "higher_return":
		return "按收益得分排序：近1月、近3月收益权重更高，再参考近6月、近1年收益与回撤"
	default:
		return "按综合得分排序：近期1-3个月收益与回撤权重更高，兼顾6月、1年、夏普和 Calmar"
	}
}

func betterDataHint(refreshStatus FundRefreshStatus, universe betterCandidateUniverse) string {
	if universe.NetworkRefresh {
		hint := "当前推荐已按同类基金池联网补抓后重排"
		if universe.RefreshedCount > 0 {
			hint += "，本次新增刷新 " + strconv.Itoa(universe.RefreshedCount) + " 只"
		}
		if universe.Limited {
			hint += "。当前范围过大，已对联网样本做限流补抓"
		}
		if universe.ComparedUniverse > 0 {
			hint += " 当前比较范围：" + universe.ScopeLabel + "，有效样本 " + strconv.Itoa(universe.ComparedUniverse) + " 只"
			if universe.UniverseTotal > 0 {
				hint += " / 目标池 " + strconv.Itoa(universe.UniverseTotal) + " 只"
			}
			hint += "。"
		}
		return hint
	}
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
	case "drawdown1", "drawdown3", "drawdown6", "drawdown12", "volatility12", "sharpe12", "calmar12":
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
	updates := map[string]any{
		"screen_updated_at": time.Now().Format("2006-01-02 15:04:05"),
	}
	if basic, err := api.CrawlFundBasic(code); err == nil && basic != nil {
		if strings.TrimSpace(basic.TrackingTarget) != "" {
			updates["tracking_target"] = strings.TrimSpace(basic.TrackingTarget)
		}
		if basic.NetGrowth1 != nil {
			updates["net_growth1"] = basic.NetGrowth1
		}
		if basic.NetGrowth3 != nil {
			updates["net_growth3"] = basic.NetGrowth3
		}
		if basic.NetGrowth6 != nil {
			updates["net_growth6"] = basic.NetGrowth6
		}
		if basic.NetGrowth12 != nil {
			updates["net_growth12"] = basic.NetGrowth12
		}
		if basic.RedeemFeeFreeDays > 0 {
			updates["redeem_fee_free_days"] = basic.RedeemFeeFreeDays
		}
	} else if err != nil {
		logger.SugaredLogger.Warnf("crawl fund basic failed for %s: %v", code, err)
	}

	if rankings, err := api.GetFundStageRankings(code); err == nil {
		for _, item := range rankings {
			switch normalizeFundStagePeriod(item.Period) {
			case "7d":
				updates["net_growth7"] = item.ReturnRate
			case "1m":
				updates["net_growth1"] = item.ReturnRate
				updates["stage_rank1_m"] = item.Rank
				updates["stage_rank1_m_total"] = item.RankTotal
			case "3m":
				updates["net_growth3"] = item.ReturnRate
				updates["stage_rank3_m"] = item.Rank
				updates["stage_rank3_m_total"] = item.RankTotal
			case "6m":
				updates["net_growth6"] = item.ReturnRate
				updates["stage_rank6_m"] = item.Rank
				updates["stage_rank6_m_total"] = item.RankTotal
			case "12m":
				updates["net_growth12"] = item.ReturnRate
				updates["stage_rank12_m"] = item.Rank
				updates["stage_rank12_m_total"] = item.RankTotal
			}
		}
	} else {
		logger.SugaredLogger.Warnf("get fund stage rankings failed for %s: %v", code, err)
	}

	if trend, _, _, err := api.GetFundTrend(code); err == nil {
		riskSnapshot := calcFundRiskSnapshot(trend, time.Now())
		if riskSnapshot.MaxDrawdown1 != nil {
			updates["max_drawdown1"] = riskSnapshot.MaxDrawdown1
		}
		if riskSnapshot.MaxDrawdown3 != nil {
			updates["max_drawdown3"] = riskSnapshot.MaxDrawdown3
		}
		if riskSnapshot.MaxDrawdown6 != nil {
			updates["max_drawdown6"] = riskSnapshot.MaxDrawdown6
		}
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
	} else {
		logger.SugaredLogger.Warnf("get fund trend failed for %s: %v", code, err)
	}

	if industry, err := api.GetFundTopIndustry(code); err == nil && industry != nil {
		updates["top_industry"] = industry.Industry
		updates["top_industry_weight"] = industry.Weight
		updates["top_industry_date"] = industry.ReportDate
	}

	return db.Dao.Model(&data.FundBasic{}).Where("code = ?", code).Updates(updates).Error == nil
}

func (s *Service) refreshFollowedFundMarketData(code string, fallbackName string) {
	code = strings.TrimSpace(code)
	if code == "" {
		return
	}

	ensureFollowedFund(code, strings.TrimSpace(fallbackName))

	api := data.NewFundApi()
	api.CrawlFundNetUnitValue(code)
	api.CrawlFundNetEstimatedUnit(code)
}

func normalizeFundStagePeriod(raw string) string {
	period := strings.TrimSpace(strings.ReplaceAll(raw, " ", ""))
	period = strings.ReplaceAll(period, "\u00a0", "")
	switch {
	case strings.Contains(period, "近1周"), strings.Contains(period, "近7天"), strings.Contains(period, "7天"), strings.Contains(period, "7日"), strings.Contains(period, "1周"), strings.Contains(period, "1星期"):
		return "7d"
	case strings.Contains(period, "近1月"), strings.Contains(period, "1月"), strings.Contains(period, "1个月"):
		return "1m"
	case strings.Contains(period, "近3月"), strings.Contains(period, "3月"), strings.Contains(period, "3个月"):
		return "3m"
	case strings.Contains(period, "近6月"), strings.Contains(period, "6月"), strings.Contains(period, "6个月"):
		return "6m"
	case strings.Contains(period, "近1年"), strings.Contains(period, "1年"):
		return "12m"
	}
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
	view.TrackingTarget = strings.TrimSpace(followed.FundBasic.TrackingTarget)
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
		TrackingTarget:    strings.TrimSpace(basic.TrackingTarget),
		Category:          classification.Category,
		CategoryLabel:     classification.Label,
		RiskLevel:         classification.RiskLevel,
		RedeemFeeFreeDays: basic.RedeemFeeFreeDays,
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
		MaxDrawdown1:      basic.MaxDrawdown1,
		MaxDrawdown3:      basic.MaxDrawdown3,
		MaxDrawdown6:      basic.MaxDrawdown6,
		MaxDrawdown12:     basic.MaxDrawdown12,
		Volatility12:      basic.Volatility12,
		Sharpe12:          basic.Sharpe12,
		Calmar12:          basic.Calmar12,
		StageRank1M:       basic.StageRank1M,
		StageRank1MTotal:  basic.StageRank1MTotal,
		StageRank3M:       basic.StageRank3M,
		StageRank3MTotal:  basic.StageRank3MTotal,
		StageRank6M:       basic.StageRank6M,
		StageRank6MTotal:  basic.StageRank6MTotal,
		StageRank12M:      basic.StageRank12M,
		StageRank12MTotal: basic.StageRank12MTotal,
		ScreenUpdatedAt:   basic.ScreenUpdatedAt,
		Watchlist:         watchlist,
	}
}

func applyBetterFundPreferenceFilters(candidates []BetterFundCandidate, query BetterFundQuery) []BetterFundCandidate {
	if len(candidates) == 0 {
		return candidates
	}

	hydrateBetterCandidatePreferences(candidates)
	filtered := make([]BetterFundCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if !passesBetterFundPreferenceFilters(candidate, query) {
			continue
		}
		filtered = append(filtered, candidate)
	}

	comparedUniverse := len(filtered) + 1
	for index := range filtered {
		filtered[index].RecommendationRank = index + 1
		filtered[index].ComparedUniverse = comparedUniverse
	}
	return filtered
}

func hydrateBetterCandidatePreferences(candidates []BetterFundCandidate) {
	codes := make([]string, 0, len(candidates))
	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		code := strings.TrimSpace(candidate.Code)
		if code == "" {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		codes = append(codes, code)
	}
	if len(codes) == 0 {
		return
	}

	var basics []data.FundBasic
	if err := db.Dao.Select(
		"code, redeem_fee_free_days, max_drawdown1, max_drawdown3, max_drawdown6, max_drawdown12, "+
			"stage_rank1_m, stage_rank1_m_total, stage_rank3_m, stage_rank3_m_total, "+
			"stage_rank6_m, stage_rank6_m_total, stage_rank12_m, stage_rank12_m_total",
	).Where("code IN ?", codes).Find(&basics).Error; err != nil {
		return
	}

	basicMap := make(map[string]data.FundBasic, len(basics))
	for _, basic := range basics {
		basicMap[basic.Code] = basic
	}
	for index := range candidates {
		basic, ok := basicMap[candidates[index].Code]
		if !ok {
			continue
		}
		if candidates[index].RedeemFeeFreeDays <= 0 {
			candidates[index].RedeemFeeFreeDays = basic.RedeemFeeFreeDays
		}
		if candidates[index].MaxDrawdown1 == nil {
			candidates[index].MaxDrawdown1 = cloneFloat(basic.MaxDrawdown1)
		}
		if candidates[index].MaxDrawdown3 == nil {
			candidates[index].MaxDrawdown3 = cloneFloat(basic.MaxDrawdown3)
		}
		if candidates[index].MaxDrawdown6 == nil {
			candidates[index].MaxDrawdown6 = cloneFloat(basic.MaxDrawdown6)
		}
		if candidates[index].MaxDrawdown12 == nil {
			candidates[index].MaxDrawdown12 = cloneFloat(basic.MaxDrawdown12)
		}
		if candidates[index].StageRank1M <= 0 {
			candidates[index].StageRank1M = basic.StageRank1M
			candidates[index].StageRank1MTotal = basic.StageRank1MTotal
		}
		if candidates[index].StageRank3M <= 0 {
			candidates[index].StageRank3M = basic.StageRank3M
			candidates[index].StageRank3MTotal = basic.StageRank3MTotal
		}
		if candidates[index].StageRank6M <= 0 {
			candidates[index].StageRank6M = basic.StageRank6M
			candidates[index].StageRank6MTotal = basic.StageRank6MTotal
		}
		if candidates[index].StageRank12M <= 0 {
			candidates[index].StageRank12M = basic.StageRank12M
			candidates[index].StageRank12MTotal = basic.StageRank12MTotal
		}
	}
}

func passesBetterFundPreferenceFilters(candidate BetterFundCandidate, query BetterFundQuery) bool {
	isAClass := isAFundShareClass(candidate.Name)
	if query.OnlyAClass {
		if !isAClass {
			return false
		}
	} else if !query.IncludeAClass && isAClass {
		return false
	}
	if !query.FeeFree7 && !query.FeeFree30 {
		return true
	}

	days := candidate.RedeemFeeFreeDays
	if days <= 0 {
		return false
	}
	if query.FeeFree7 && days <= 7 {
		return true
	}
	if query.FeeFree30 && days > 7 && days <= 30 {
		return true
	}
	return false
}

func isAFundShareClass(name string) bool {
	normalized := strings.ToUpper(strings.TrimSpace(name))
	if normalized == "" {
		return false
	}
	match := regexp.MustCompile(`([A-Z])$`).FindStringSubmatch(normalized)
	return len(match) == 2 && match[1] == "A"
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

func calcFundRiskSnapshot(points []data.FundTrendPoint, now time.Time) fundRiskSnapshot {
	snapshot := fundRiskSnapshot{
		MaxDrawdown1:  calcFundMaxDrawdown(points, now.AddDate(0, -1, 0)),
		MaxDrawdown3:  calcFundMaxDrawdown(points, now.AddDate(0, -3, 0)),
		MaxDrawdown6:  calcFundMaxDrawdown(points, now.AddDate(0, -6, 0)),
		MaxDrawdown12: calcFundMaxDrawdown(points, now.AddDate(-1, 0, 0)),
	}
	if len(points) < 2 {
		return snapshot
	}

	filtered := make([]data.FundTrendPoint, 0, len(points))
	for _, point := range points {
		if point.Timestamp <= 0 {
			continue
		}
		if time.UnixMilli(point.Timestamp).Before(now.AddDate(-1, 0, 0)) {
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

func isSameTradingDayTime(value time.Time, target time.Time) bool {
	if value.IsZero() {
		return false
	}
	return value.In(target.Location()).Format("2006-01-02") == target.Format("2006-01-02")
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
