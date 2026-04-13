package portfolio

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type Holding struct {
	gorm.Model
	StockCode            string                `json:"stockCode" gorm:"index"`
	StockName            string                `json:"stockName"`
	HoldingType          string                `json:"holdingType" gorm:"index"`
	Market               string                `json:"market"`
	AvgCost              float64               `json:"avgCost"`
	Quantity             float64               `json:"quantity"`
	CurrentPrice         float64               `json:"currentPrice"`
	LatestDailyRate      *float64              `json:"latestDailyRate"`
	LatestDailyUpdatedAt string                `json:"latestDailyUpdatedAt"`
	ProfitLoss           float64               `json:"profitLoss"`
	ProfitRate           float64               `json:"profitRate"`
	TotalCost            float64               `json:"totalCost"`
	TotalValue           float64               `json:"totalValue"`
	TodayChange          float64               `json:"todayChange"`
	TodayRate            float64               `json:"todayRate"`
	BuyDate              *time.Time            `json:"buyDate"`
	BrokerName           string                `json:"brokerName"`
	AccountTag           string                `json:"accountTag"`
	Remark               string                `json:"remark"`
	IsDel                soft_delete.DeletedAt `gorm:"softDelete:flag" json:"-"`
}

func (Holding) TableName() string {
	return "holdings"
}

type Transaction struct {
	gorm.Model
	HoldingID   uint                  `json:"holdingId" gorm:"index"`
	StockCode   string                `json:"stockCode" gorm:"index"`
	StockName   string                `json:"stockName"`
	HoldingType string                `json:"holdingType"`
	BrokerName  string                `json:"brokerName"`
	AccountTag  string                `json:"accountTag"`
	Type        string                `json:"type"`
	Price       float64               `json:"price"`
	Quantity    float64               `json:"quantity"`
	Amount      float64               `json:"amount"`
	Fee         float64               `json:"fee"`
	TradeDate   *time.Time            `json:"tradeDate" gorm:"index"`
	Remark      string                `json:"remark"`
	IsDel       soft_delete.DeletedAt `gorm:"softDelete:flag" json:"-"`
}

func (Transaction) TableName() string {
	return "transactions"
}

type ProfitSnapshot struct {
	gorm.Model
	SnapshotDate *time.Time `json:"snapshotDate" gorm:"uniqueIndex"`
	TotalCost    float64    `json:"totalCost"`
	TotalValue   float64    `json:"totalValue"`
	TotalProfit  float64    `json:"totalProfit"`
	ProfitRate   float64    `json:"profitRate"`
	StockValue   float64    `json:"stockValue"`
	FundValue    float64    `json:"fundValue"`
}

func (ProfitSnapshot) TableName() string {
	return "profit_snapshots"
}

type PortfolioSummary struct {
	TotalCost       float64   `json:"totalCost"`
	TotalValue      float64   `json:"totalValue"`
	TotalProfit     float64   `json:"totalProfit"`
	TotalProfitRate float64   `json:"totalProfitRate"`
	TodayProfit     float64   `json:"todayProfit"`
	StockCount      int       `json:"stockCount"`
	FundCount       int       `json:"fundCount"`
	BondFundCount   int       `json:"bondFundCount"`
	CashFundCount   int       `json:"cashFundCount"`
	EquityFundCount int       `json:"equityFundCount"`
	FundValue       float64   `json:"fundValue"`
	StockValue      float64   `json:"stockValue"`
	BondFundValue   float64   `json:"bondFundValue"`
	CashFundValue   float64   `json:"cashFundValue"`
	EquityFundValue float64   `json:"equityFundValue"`
	Holdings        []Holding `json:"holdings"`
}

type AllocationItem struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
	Ratio float64 `json:"ratio"`
	Count int     `json:"count"`
}

type FundHoldingView struct {
	Holding
	FundType         string   `json:"fundType"`
	FundCompany      string   `json:"fundCompany"`
	FundManager      string   `json:"fundManager"`
	FundRating       string   `json:"fundRating"`
	FundScale        string   `json:"fundScale"`
	TrackingTarget   string   `json:"trackingTarget"`
	Category         string   `json:"category"`
	CategoryLabel    string   `json:"categoryLabel"`
	RiskLevel        string   `json:"riskLevel"`
	NetUnitValue     *float64 `json:"netUnitValue"`
	NetUnitValueDate string   `json:"netUnitValueDate"`
	NetEstimatedUnit *float64 `json:"netEstimatedUnit"`
	NetEstimatedTime string   `json:"netEstimatedTime"`
	NetEstimatedRate *float64 `json:"netEstimatedRate"`
	NetGrowth1       *float64 `json:"netGrowth1"`
	NetGrowth3       *float64 `json:"netGrowth3"`
	NetGrowth6       *float64 `json:"netGrowth6"`
	NetGrowth12      *float64 `json:"netGrowth12"`
	NetGrowth36      *float64 `json:"netGrowth36"`
	NetGrowthYTD     *float64 `json:"netGrowthYTD"`
	EstimateUpdated  bool     `json:"estimateUpdated"`
	EstimateStatus   string   `json:"estimateStatus"`
}

type FundTrendPoint struct {
	Timestamp   int64    `json:"timestamp"`
	Date        string   `json:"date"`
	Value       float64  `json:"value"`
	DailyReturn *float64 `json:"dailyReturn,omitempty"`
}

type FundEstimatePoint struct {
	Timestamp     int64    `json:"timestamp"`
	Time          string   `json:"time"`
	EstimatedUnit float64  `json:"estimatedUnit"`
	EstimatedRate *float64 `json:"estimatedRate"`
}

type FundStageRanking struct {
	Period             string   `json:"period"`
	ReturnRate         *float64 `json:"returnRate"`
	SimilarAverageRate *float64 `json:"similarAverageRate"`
	BenchmarkLabel     string   `json:"benchmarkLabel"`
	BenchmarkRate      *float64 `json:"benchmarkRate"`
	Rank               int      `json:"rank"`
	RankTotal          int      `json:"rankTotal"`
	RankPercentile     *float64 `json:"rankPercentile"`
	RankDelta          int      `json:"rankDelta"`
	RankDeltaDirection string   `json:"rankDeltaDirection"`
	Quartile           string   `json:"quartile"`
}

type FundProfile struct {
	FundHoldingView
	Trend                  []FundTrendPoint    `json:"trend"`
	TrendUpdatedAt         string              `json:"trendUpdatedAt"`
	LatestReturn           *float64            `json:"latestReturn"`
	EstimateTrend          []FundEstimatePoint `json:"estimateTrend"`
	EstimateTrendUpdatedAt string              `json:"estimateTrendUpdatedAt"`
	EstimateLatestRate     *float64            `json:"estimateLatestRate"`
	StageRankings          []FundStageRanking  `json:"stageRankings"`
	StageRankingsUpdatedAt string              `json:"stageRankingsUpdatedAt"`
}

type FundScreenerQuery struct {
	Keyword       string   `json:"keyword"`
	FundType      string   `json:"fundType"`
	Category      string   `json:"category"`
	Industry      string   `json:"industry"`
	MinReturn7    *float64 `json:"minReturn7"`
	MinReturn1    *float64 `json:"minReturn1"`
	MinReturn3    *float64 `json:"minReturn3"`
	MaxDrawdown12 *float64 `json:"maxDrawdown12"`
	OnlyWatchlist bool     `json:"onlyWatchlist"`
	Page          int      `json:"page"`
	PageSize      int      `json:"pageSize"`
	SortBy        string   `json:"sortBy"`
	SortOrder     string   `json:"sortOrder"`
}

type FundScreenerItem struct {
	Code              string   `json:"code"`
	Name              string   `json:"name"`
	FundType          string   `json:"fundType"`
	TrackingTarget    string   `json:"trackingTarget"`
	Category          string   `json:"category"`
	CategoryLabel     string   `json:"categoryLabel"`
	RiskLevel         string   `json:"riskLevel"`
	RedeemFeeFreeDays int      `json:"redeemFeeFreeDays"`
	Company           string   `json:"company"`
	Manager           string   `json:"manager"`
	Rating            string   `json:"rating"`
	Scale             string   `json:"scale"`
	TopIndustry       string   `json:"topIndustry"`
	TopIndustryWeight *float64 `json:"topIndustryWeight"`
	TopIndustryDate   string   `json:"topIndustryDate"`
	NetGrowth7        *float64 `json:"netGrowth7"`
	NetGrowth1        *float64 `json:"netGrowth1"`
	NetGrowth3        *float64 `json:"netGrowth3"`
	NetGrowth6        *float64 `json:"netGrowth6"`
	NetGrowth12       *float64 `json:"netGrowth12"`
	MaxDrawdown1      *float64 `json:"maxDrawdown1"`
	MaxDrawdown3      *float64 `json:"maxDrawdown3"`
	MaxDrawdown6      *float64 `json:"maxDrawdown6"`
	MaxDrawdown12     *float64 `json:"maxDrawdown12"`
	Volatility12      *float64 `json:"volatility12"`
	Sharpe12          *float64 `json:"sharpe12"`
	Calmar12          *float64 `json:"calmar12"`
	StageRank1M       int      `json:"stageRank1m"`
	StageRank1MTotal  int      `json:"stageRank1mTotal"`
	StageRank3M       int      `json:"stageRank3m"`
	StageRank3MTotal  int      `json:"stageRank3mTotal"`
	StageRank6M       int      `json:"stageRank6m"`
	StageRank6MTotal  int      `json:"stageRank6mTotal"`
	StageRank12M      int      `json:"stageRank12m"`
	StageRank12MTotal int      `json:"stageRank12mTotal"`
	ScreenUpdatedAt   string   `json:"screenUpdatedAt"`
	Watchlist         bool     `json:"watchlist"`
}

type FundScreenerResult struct {
	Items           []FundScreenerItem `json:"items"`
	Total           int64              `json:"total"`
	Page            int                `json:"page"`
	PageSize        int                `json:"pageSize"`
	UniverseCount   int64              `json:"universeCount"`
	ScreenedCount   int64              `json:"screenedCount"`
	TypeOptions     []string           `json:"typeOptions"`
	CategoryOptions []string           `json:"categoryOptions"`
	IndustryOptions []string           `json:"industryOptions"`
	LastRefreshHint string             `json:"lastRefreshHint"`
	RefreshStatus   FundRefreshStatus  `json:"refreshStatus"`
}

type FundRefreshStatus struct {
	State           string `json:"state"`
	StateLabel      string `json:"stateLabel"`
	Scope           string `json:"scope"`
	Refreshing      bool   `json:"refreshing"`
	NeedsRefresh    bool   `json:"needsRefresh"`
	Triggered       bool   `json:"triggered"`
	CurrentDate     string `json:"currentDate"`
	LastRefreshHint string `json:"lastRefreshHint"`
	UpdatedToday    int64  `json:"updatedToday"`
	ScreenedCount   int64  `json:"screenedCount"`
	UniverseCount   int64  `json:"universeCount"`
	TargetCount     int64  `json:"targetCount"`
	TargetUpdated   int64  `json:"targetUpdated"`
	TargetPending   int64  `json:"targetPending"`
	ProgressCurrent int64  `json:"progressCurrent"`
	ProgressTotal   int64  `json:"progressTotal"`
	CurrentCode     string `json:"currentCode"`
	Message         string `json:"message"`
}

type BetterFundQuery struct {
	ReferenceCode   string `json:"referenceCode"`
	SameTypeOnly    bool   `json:"sameTypeOnly"`
	SameSubTypeOnly bool   `json:"sameSubTypeOnly"`
	Dimension       string `json:"dimension"`
	NetworkRefresh  bool   `json:"networkRefresh"`
	FeeFree7        bool   `json:"feeFree7"`
	FeeFree30       bool   `json:"feeFree30"`
	IncludeAClass   bool   `json:"includeAClass"`
	OnlyAClass      bool   `json:"onlyAClass"`
	Page            int    `json:"page"`
	PageSize        int    `json:"pageSize"`
}

type BetterFundCandidate struct {
	FundScreenerItem
	RecommendationRank int                `json:"recommendationRank"`
	BetterScore        float64            `json:"betterScore"`
	ReasonSummary      string             `json:"reasonSummary"`
	ScopeLabel         string             `json:"scopeLabel"`
	ComparedUniverse   int                `json:"comparedUniverse"`
	Reasons            []string           `json:"reasons"`
	Metrics            []BetterFundMetric `json:"metrics"`
}

type BetterFundMetric struct {
	Key                 string   `json:"key"`
	Label               string   `json:"label"`
	Better              string   `json:"better"`
	CandidateValue      *float64 `json:"candidateValue"`
	ReferenceValue      *float64 `json:"referenceValue"`
	Delta               *float64 `json:"delta"`
	Advantage           *float64 `json:"advantage"`
	Weight              float64  `json:"weight"`
	Contribution        float64  `json:"contribution"`
	CandidateRank       int      `json:"candidateRank"`
	ReferenceRank       int      `json:"referenceRank"`
	RankTotal           int      `json:"rankTotal"`
	CandidatePercentile *float64 `json:"candidatePercentile"`
	ReferencePercentile *float64 `json:"referencePercentile"`
}

type BetterFundResult struct {
	Reference        FundScreenerItem      `json:"reference"`
	Candidates       []BetterFundCandidate `json:"candidates"`
	Dimension        string                `json:"dimension"`
	SortLabel        string                `json:"sortLabel"`
	ScopeLabel       string                `json:"scopeLabel"`
	ComparedUniverse int                   `json:"comparedUniverse"`
	UniverseTotal    int                   `json:"universeTotal"`
	RefreshedCount   int                   `json:"refreshedCount"`
	NetworkRefresh   bool                  `json:"networkRefresh"`
	FallbackApplied  bool                  `json:"fallbackApplied"`
	DataHint         string                `json:"dataHint"`
	Total            int64                 `json:"total"`
	Page             int                   `json:"page"`
	PageSize         int                   `json:"pageSize"`
	RefreshStatus    FundRefreshStatus     `json:"refreshStatus"`
}

type FundRecommendationCache struct {
	gorm.Model
	ReferenceCode    string `json:"referenceCode" gorm:"index:idx_fund_recommendation_cache_ref_date_scope,priority:1"`
	RefreshDate      string `json:"refreshDate" gorm:"index:idx_fund_recommendation_cache_ref_date_scope,priority:2;index"`
	SameTypeOnly     bool   `json:"sameTypeOnly" gorm:"index:idx_fund_recommendation_cache_ref_date_scope,priority:3"`
	SameSubTypeOnly  bool   `json:"sameSubTypeOnly" gorm:"index:idx_fund_recommendation_cache_ref_date_scope,priority:4"`
	Dimension        string `json:"dimension" gorm:"index:idx_fund_recommendation_cache_ref_date_scope,priority:5"`
	FeeFree7         bool   `json:"feeFree7" gorm:"index:idx_fund_recommendation_cache_ref_date_scope,priority:6"`
	FeeFree30        bool   `json:"feeFree30" gorm:"index:idx_fund_recommendation_cache_ref_date_scope,priority:7"`
	IncludeAClass    bool   `json:"includeAClass" gorm:"index:idx_fund_recommendation_cache_ref_date_scope,priority:8"`
	OnlyAClass       bool   `json:"onlyAClass" gorm:"index:idx_fund_recommendation_cache_ref_date_scope,priority:9"`
	ScopeLabel       string `json:"scopeLabel"`
	SortLabel        string `json:"sortLabel"`
	FallbackApplied  bool   `json:"fallbackApplied"`
	ComparedUniverse int    `json:"comparedUniverse"`
	UniverseTotal    int    `json:"universeTotal"`
	DataHint         string `json:"dataHint" gorm:"type:text"`
	CandidatesJSON   string `json:"candidatesJson" gorm:"type:text"`
}

func (FundRecommendationCache) TableName() string {
	return "fund_recommendation_cache"
}

type FundRecommendationProgress struct {
	gorm.Model
	ReferenceCode    string `json:"referenceCode" gorm:"index:idx_fund_recommendation_progress_ref_date,priority:1"`
	RefreshDate      string `json:"refreshDate" gorm:"index:idx_fund_recommendation_progress_ref_date,priority:2;index"`
	Status           string `json:"status" gorm:"index"`
	CurrentPhase     string `json:"currentPhase"`
	Message          string `json:"message" gorm:"type:text"`
	LastError        string `json:"lastError" gorm:"type:text"`
	ComparedUniverse int    `json:"comparedUniverse"`
	UniverseTotal    int    `json:"universeTotal"`
}

func (FundRecommendationProgress) TableName() string {
	return "fund_recommendation_progress"
}

type FundRecommendationRefreshStatus struct {
	State           string `json:"state"`
	StateLabel      string `json:"stateLabel"`
	Refreshing      bool   `json:"refreshing"`
	Triggered       bool   `json:"triggered"`
	CurrentDate     string `json:"currentDate"`
	WatchlistCount  int64  `json:"watchlistCount"`
	CompletedCount  int64  `json:"completedCount"`
	PendingCount    int64  `json:"pendingCount"`
	FailedCount     int64  `json:"failedCount"`
	ProgressCurrent int64  `json:"progressCurrent"`
	ProgressTotal   int64  `json:"progressTotal"`
	CurrentCode     string `json:"currentCode"`
	LastRefreshHint string `json:"lastRefreshHint"`
	Message         string `json:"message"`
}

type FundCompareQuery struct {
	Codes []string `json:"codes"`
}

type FundCompareResult struct {
	Items        []FundScreenerItem `json:"items"`
	Total        int                `json:"total"`
	MissingCodes []string           `json:"missingCodes"`
	RefreshedAt  string             `json:"refreshedAt"`
}

type FundPortfolioDashboard struct {
	Summary              PortfolioSummary  `json:"summary"`
	Positions            []FundHoldingView `json:"positions"`
	TypeAllocation       []AllocationItem  `json:"typeAllocation"`
	PlatformAllocation   []AllocationItem  `json:"platformAllocation"`
	AccountAllocation    []AllocationItem  `json:"accountAllocation"`
	CompanyAllocation    []AllocationItem  `json:"companyAllocation"`
	ConservativeRatio    float64           `json:"conservativeRatio"`
	BondAllocationRatio  float64           `json:"bondAllocationRatio"`
	EstimatedProfitToday float64           `json:"estimatedProfitToday"`
}

type FundPositionInput struct {
	StockCode      string  `json:"stockCode"`
	StockName      string  `json:"stockName"`
	PositionAmount float64 `json:"positionAmount"`
	CostAmount     float64 `json:"costAmount"`
	BrokerName     string  `json:"brokerName"`
	AccountTag     string  `json:"accountTag"`
	Remark         string  `json:"remark"`
}
