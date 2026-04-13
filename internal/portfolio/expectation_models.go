package portfolio

import "gorm.io/gorm"

type PortfolioExpectationItem struct {
	Code                      string  `json:"code"`
	Name                      string  `json:"name"`
	HoldingType               string  `json:"holdingType"`
	Category                  string  `json:"category"`
	CategoryLabel             string  `json:"categoryLabel"`
	TrackingTarget            string  `json:"trackingTarget"`
	Bucket                    string  `json:"bucket"`
	BucketLabel               string  `json:"bucketLabel"`
	BrokerName                string  `json:"brokerName"`
	AccountTag                string  `json:"accountTag"`
	TotalValue                float64 `json:"totalValue"`
	TotalCost                 float64 `json:"totalCost"`
	CurrentProfit             float64 `json:"currentProfit"`
	CurrentProfitRate         float64 `json:"currentProfitRate"`
	EstimatedAnnualReturnRate float64 `json:"estimatedAnnualReturnRate"`
	EstimatedAnnualProfit     float64 `json:"estimatedAnnualProfit"`
	WeightInPortfolio         float64 `json:"weightInPortfolio"`
	DaysHeld                  int     `json:"daysHeld"`
	BasisLabel                string  `json:"basisLabel"`
}

type PortfolioExpectationBucket struct {
	Key                       string  `json:"key"`
	Label                     string  `json:"label"`
	Value                     float64 `json:"value"`
	Weight                    float64 `json:"weight"`
	EstimatedAnnualReturnRate float64 `json:"estimatedAnnualReturnRate"`
	EstimatedAnnualProfit     float64 `json:"estimatedAnnualProfit"`
	Count                     int     `json:"count"`
}

type PortfolioExpectationSummary struct {
	GeneratedAt                       string                       `json:"generatedAt"`
	HouseholdLiquidAssets             float64                      `json:"householdLiquidAssets"`
	TargetAnnualReturnRate            float64                      `json:"targetAnnualReturnRate"`
	AnnualUntrackedProfit             float64                      `json:"annualUntrackedProfit"`
	TargetAnnualProfit                float64                      `json:"targetAnnualProfit"`
	TargetDifficultyLabel             string                       `json:"targetDifficultyLabel"`
	YearProgressRatio                 float64                      `json:"yearProgressRatio"`
	TargetProfitToDate                float64                      `json:"targetProfitToDate"`
	CurrentPortfolioValue             float64                      `json:"currentPortfolioValue"`
	CurrentTotalProfit                float64                      `json:"currentTotalProfit"`
	CurrentTotalProfitWithManual      float64                      `json:"currentTotalProfitWithManual"`
	CurrentTotalProfitRate            float64                      `json:"currentTotalProfitRate"`
	InvestedValue                     float64                      `json:"investedValue"`
	InvestedRatioOfLiquidAssets       float64                      `json:"investedRatioOfLiquidAssets"`
	IdleLiquidAssets                  float64                      `json:"idleLiquidAssets"`
	FundValue                         float64                      `json:"fundValue"`
	StockValue                        float64                      `json:"stockValue"`
	FundCount                         int                          `json:"fundCount"`
	StockCount                        int                          `json:"stockCount"`
	EstimatedFundAnnualProfit         float64                      `json:"estimatedFundAnnualProfit"`
	EstimatedStockAnnualProfit        float64                      `json:"estimatedStockAnnualProfit"`
	EstimatedPortfolioAnnualProfit    float64                      `json:"estimatedPortfolioAnnualProfit"`
	CombinedAnnualProfitProjection    float64                      `json:"combinedAnnualProfitProjection"`
	EstimatedHoldingsAnnualReturnRate float64                      `json:"estimatedHoldingsAnnualReturnRate"`
	EstimatedLiquidAnnualReturnRate   float64                      `json:"estimatedLiquidAnnualReturnRate"`
	CombinedLiquidAnnualReturnRate    float64                      `json:"combinedLiquidAnnualReturnRate"`
	ProjectedCompletionRatio          float64                      `json:"projectedCompletionRatio"`
	AnnualGap                         float64                      `json:"annualGap"`
	ToDateGap                         float64                      `json:"toDateGap"`
	RequiredReturnOnInvestedCapital   float64                      `json:"requiredReturnOnInvestedCapital"`
	RequiredReturnOnIdleLiquidAssets  float64                      `json:"requiredReturnOnIdleLiquidAssets"`
	ConservativeValue                 float64                      `json:"conservativeValue"`
	ConservativeRatio                 float64                      `json:"conservativeRatio"`
	ConservativeExpectedReturnRate    float64                      `json:"conservativeExpectedReturnRate"`
	GrowthValue                       float64                      `json:"growthValue"`
	GrowthRatio                       float64                      `json:"growthRatio"`
	GrowthExpectedReturnRate          float64                      `json:"growthExpectedReturnRate"`
	SuggestedFixedIncomeMaxRatio      float64                      `json:"suggestedFixedIncomeMaxRatio"`
	SuggestedFixedIncomeMaxAmount     float64                      `json:"suggestedFixedIncomeMaxAmount"`
	SuggestedGrowthMinAmount          float64                      `json:"suggestedGrowthMinAmount"`
	Buckets                           []PortfolioExpectationBucket `json:"buckets"`
	TopDrivers                        []PortfolioExpectationItem   `json:"topDrivers"`
	BottomDraggers                    []PortfolioExpectationItem   `json:"bottomDraggers"`
	Items                             []PortfolioExpectationItem   `json:"items"`
	Warnings                          []string                     `json:"warnings"`
}

type PortfolioExpectationAIAnalysis struct {
	gorm.Model
	TriggerSource    string `json:"triggerSource"`
	AIConfigID       int    `json:"aiConfigId"`
	PromptTemplateID int    `json:"promptTemplateId"`
	ModelName        string `json:"modelName"`
	Status           string `json:"status"`
	Prompt           string `json:"prompt" gorm:"type:text"`
	InputPayload     string `json:"inputPayload" gorm:"type:text"`
	AnalysisMarkdown string `json:"analysisMarkdown" gorm:"type:text"`
	ErrorMessage     string `json:"errorMessage" gorm:"type:text"`
}

func (PortfolioExpectationAIAnalysis) TableName() string {
	return "portfolio_expectation_ai_analyses"
}
