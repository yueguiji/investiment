package portfolio

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"go-stock/backend/db"
	"go-stock/backend/logger"
)

func (s *Service) BuildExpectationSummary(liquidAssets float64, targetAnnualReturnRate float64, annualUntrackedProfit float64) *PortfolioExpectationSummary {
	now := time.Now()
	summary := s.GetPortfolioSummary()
	fundDashboard := s.GetFundDashboard()
	stockHoldings := s.GetHoldingsByType("stock")

	result := &PortfolioExpectationSummary{
		GeneratedAt:            now.Format("2006-01-02 15:04:05"),
		HouseholdLiquidAssets:  roundMoney(maxFloat64(liquidAssets, 0)),
		TargetAnnualReturnRate: roundPercent(clampFloat(targetAnnualReturnRate, 0, 100)),
		AnnualUntrackedProfit:  roundMoney(annualUntrackedProfit),
		CurrentPortfolioValue:  roundMoney(summary.TotalValue),
		CurrentTotalProfit:     roundMoney(summary.TotalProfit),
		CurrentTotalProfitRate: roundPercent(summary.TotalProfitRate),
		FundValue:              roundMoney(summary.FundValue),
		StockValue:             roundMoney(summary.StockValue),
		FundCount:              summary.FundCount,
		StockCount:             summary.StockCount,
	}
	result.CurrentTotalProfitWithManual = roundMoney(result.CurrentTotalProfit + result.AnnualUntrackedProfit)

	if result.HouseholdLiquidAssets > 0 && result.TargetAnnualReturnRate > 0 {
		result.TargetAnnualProfit = roundMoney(result.HouseholdLiquidAssets * result.TargetAnnualReturnRate / 100)
	}
	result.TargetDifficultyLabel = expectationDifficultyLabel(result.TargetAnnualReturnRate)
	result.YearProgressRatio = roundPercent(yearProgressRatio(now) * 100)
	result.TargetProfitToDate = roundMoney(result.TargetAnnualProfit * result.YearProgressRatio / 100)

	items := make([]PortfolioExpectationItem, 0, len(fundDashboard.Positions)+len(stockHoldings))
	for _, item := range fundDashboard.Positions {
		items = append(items, buildFundExpectationItem(item))
	}
	for _, holding := range stockHoldings {
		items = append(items, buildStockExpectationItem(holding, now))
	}

	result.Items = items
	result.InvestedValue = roundMoney(sumExpectationItemValue(items))
	if result.HouseholdLiquidAssets > 0 {
		result.InvestedRatioOfLiquidAssets = roundPercent(result.InvestedValue / result.HouseholdLiquidAssets * 100)
		result.IdleLiquidAssets = roundMoney(maxFloat64(result.HouseholdLiquidAssets-result.InvestedValue, 0))
	}

	for index := range result.Items {
		if result.InvestedValue > 0 {
			result.Items[index].WeightInPortfolio = roundPercent(result.Items[index].TotalValue / result.InvestedValue * 100)
		}
		switch result.Items[index].HoldingType {
		case "stock":
			result.EstimatedStockAnnualProfit += result.Items[index].EstimatedAnnualProfit
		default:
			result.EstimatedFundAnnualProfit += result.Items[index].EstimatedAnnualProfit
		}
	}
	result.EstimatedFundAnnualProfit = roundMoney(result.EstimatedFundAnnualProfit)
	result.EstimatedStockAnnualProfit = roundMoney(result.EstimatedStockAnnualProfit)
	result.EstimatedPortfolioAnnualProfit = roundMoney(result.EstimatedFundAnnualProfit + result.EstimatedStockAnnualProfit)
	result.CombinedAnnualProfitProjection = roundMoney(result.EstimatedPortfolioAnnualProfit + result.AnnualUntrackedProfit)
	if result.InvestedValue > 0 {
		result.EstimatedHoldingsAnnualReturnRate = roundPercent(result.EstimatedPortfolioAnnualProfit / result.InvestedValue * 100)
	}
	if result.HouseholdLiquidAssets > 0 {
		result.EstimatedLiquidAnnualReturnRate = roundPercent(result.EstimatedPortfolioAnnualProfit / result.HouseholdLiquidAssets * 100)
		result.CombinedLiquidAnnualReturnRate = roundPercent(result.CombinedAnnualProfitProjection / result.HouseholdLiquidAssets * 100)
	}
	if result.TargetAnnualProfit > 0 {
		result.ProjectedCompletionRatio = roundPercent(result.CombinedAnnualProfitProjection / result.TargetAnnualProfit * 100)
	}
	result.AnnualGap = roundMoney(result.TargetAnnualProfit - result.CombinedAnnualProfitProjection)
	result.ToDateGap = roundMoney(result.TargetProfitToDate - result.CurrentTotalProfitWithManual)
	if result.InvestedValue > 0 && result.TargetAnnualProfit > 0 {
		result.RequiredReturnOnInvestedCapital = roundPercent(result.TargetAnnualProfit / result.InvestedValue * 100)
	}
	if result.IdleLiquidAssets > 0 && result.AnnualGap > 0 {
		result.RequiredReturnOnIdleLiquidAssets = roundPercent(result.AnnualGap / result.IdleLiquidAssets * 100)
	}

	result.Buckets = buildExpectationBuckets(result.Items, result.InvestedValue)
	for _, bucket := range result.Buckets {
		switch bucket.Key {
		case "conservative":
			result.ConservativeValue = bucket.Value
			result.ConservativeRatio = bucket.Weight
			result.ConservativeExpectedReturnRate = bucket.EstimatedAnnualReturnRate
		case "growth", "stock":
			result.GrowthValue += bucket.Value
		}
	}
	result.GrowthValue = roundMoney(result.GrowthValue)
	if result.InvestedValue > 0 {
		result.GrowthRatio = roundPercent(result.GrowthValue / result.InvestedValue * 100)
	}
	if result.GrowthValue > 0 {
		growthProfit := 0.0
		for _, item := range result.Items {
			if item.Bucket == "growth" || item.Bucket == "stock" {
				growthProfit += item.EstimatedAnnualProfit
			}
		}
		result.GrowthExpectedReturnRate = roundPercent(growthProfit / result.GrowthValue * 100)
	}

	result.SuggestedFixedIncomeMaxRatio, result.SuggestedFixedIncomeMaxAmount, result.SuggestedGrowthMinAmount =
		calcSuggestedFixedIncomeRange(result)

	sort.Slice(result.Items, func(i, j int) bool {
		if result.Items[i].EstimatedAnnualProfit == result.Items[j].EstimatedAnnualProfit {
			return result.Items[i].TotalValue > result.Items[j].TotalValue
		}
		return result.Items[i].EstimatedAnnualProfit > result.Items[j].EstimatedAnnualProfit
	})
	result.TopDrivers = append(result.TopDrivers, result.Items[:minInt(len(result.Items), 6)]...)

	draggers := append([]PortfolioExpectationItem{}, result.Items...)
	sort.Slice(draggers, func(i, j int) bool {
		if draggers[i].EstimatedAnnualProfit == draggers[j].EstimatedAnnualProfit {
			return draggers[i].TotalValue > draggers[j].TotalValue
		}
		return draggers[i].EstimatedAnnualProfit < draggers[j].EstimatedAnnualProfit
	})
	result.BottomDraggers = append(result.BottomDraggers, draggers[:minInt(len(draggers), 4)]...)
	result.Warnings = buildExpectationWarnings(result)
	return result
}

func (s *Service) SavePortfolioExpectationAIAnalysis(item PortfolioExpectationAIAnalysis) *PortfolioExpectationAIAnalysis {
	if err := db.Dao.Create(&item).Error; err != nil {
		logger.SugaredLogger.Errorf("save portfolio expectation ai analysis failed: %v", err)
		return nil
	}
	return &item
}

func (s *Service) GetLatestPortfolioExpectationAIAnalysis() *PortfolioExpectationAIAnalysis {
	var item PortfolioExpectationAIAnalysis
	if err := db.Dao.Order("created_at desc, id desc").First(&item).Error; err != nil {
		return nil
	}
	return &item
}

func buildFundExpectationItem(view FundHoldingView) PortfolioExpectationItem {
	rate, basis := estimateFundAnnualReturn(view)
	return PortfolioExpectationItem{
		Code:                      view.StockCode,
		Name:                      view.StockName,
		HoldingType:               "fund",
		Category:                  view.Category,
		CategoryLabel:             view.CategoryLabel,
		TrackingTarget:            strings.TrimSpace(view.TrackingTarget),
		Bucket:                    expectationBucketForFund(view.Category),
		BucketLabel:               expectationBucketLabel(expectationBucketForFund(view.Category)),
		BrokerName:                view.BrokerName,
		AccountTag:                view.AccountTag,
		TotalValue:                roundMoney(view.TotalValue),
		TotalCost:                 roundMoney(view.TotalCost),
		CurrentProfit:             roundMoney(view.ProfitLoss),
		CurrentProfitRate:         roundPercent(view.ProfitRate),
		EstimatedAnnualReturnRate: roundPercent(rate),
		EstimatedAnnualProfit:     roundMoney(view.TotalValue * rate / 100),
		BasisLabel:                basis,
	}
}

func buildStockExpectationItem(holding Holding, now time.Time) PortfolioExpectationItem {
	rate, basis, daysHeld := estimateStockAnnualReturn(holding, now)
	return PortfolioExpectationItem{
		Code:                      holding.StockCode,
		Name:                      holding.StockName,
		HoldingType:               "stock",
		Category:                  "stock",
		CategoryLabel:             "股票持仓",
		Bucket:                    "stock",
		BucketLabel:               expectationBucketLabel("stock"),
		BrokerName:                holding.BrokerName,
		AccountTag:                holding.AccountTag,
		TotalValue:                roundMoney(holding.TotalValue),
		TotalCost:                 roundMoney(holding.TotalCost),
		CurrentProfit:             roundMoney(holding.ProfitLoss),
		CurrentProfitRate:         roundPercent(holding.ProfitRate),
		EstimatedAnnualReturnRate: roundPercent(rate),
		EstimatedAnnualProfit:     roundMoney(holding.TotalValue * rate / 100),
		DaysHeld:                  daysHeld,
		BasisLabel:                basis,
	}
}

func estimateFundAnnualReturn(view FundHoldingView) (float64, string) {
	type sample struct {
		label  string
		value  float64
		weight float64
	}

	samples := make([]sample, 0, 4)
	if view.NetGrowth1 != nil {
		samples = append(samples, sample{label: "近1月年化", value: annualizeMonthlyReturn(*view.NetGrowth1, 1), weight: 0.22})
	}
	if view.NetGrowth3 != nil {
		samples = append(samples, sample{label: "近3月年化", value: annualizeMonthlyReturn(*view.NetGrowth3, 3), weight: 0.30})
	}
	if view.NetGrowth6 != nil {
		samples = append(samples, sample{label: "近6月年化", value: annualizeMonthlyReturn(*view.NetGrowth6, 6), weight: 0.23})
	}
	if view.NetGrowth12 != nil {
		samples = append(samples, sample{label: "近1年", value: *view.NetGrowth12, weight: 0.25})
	}
	if len(samples) == 0 {
		return 0, "按历史收益数据缺失暂按 0% 估算"
	}

	totalWeight := 0.0
	totalValue := 0.0
	basisParts := make([]string, 0, len(samples))
	for _, item := range samples {
		totalWeight += item.weight
		totalValue += item.value * item.weight
		basisParts = append(basisParts, item.label)
	}
	if totalWeight <= 0 {
		return 0, "按历史收益数据缺失暂按 0% 估算"
	}

	estimated := totalValue / totalWeight
	switch view.Category {
	case "cash":
		estimated = clampFloat(estimated, 0.5, 4.5)
	case "bond":
		estimated = clampFloat(estimated, -4, 12)
	case "equity":
		estimated = clampFloat(estimated, -30, 36)
	default:
		estimated = clampFloat(estimated, -24, 28)
	}
	return estimated, "按" + strings.Join(basisParts, " + ") + "加权估算"
}

func estimateStockAnnualReturn(holding Holding, now time.Time) (float64, string, int) {
	if holding.TotalValue <= 0 {
		return 0, "按当前仓位为空暂按 0% 估算", 0
	}

	daysHeld := 0
	if holding.BuyDate != nil {
		daysHeld = int(now.Sub(*holding.BuyDate).Hours() / 24)
		if daysHeld < 0 {
			daysHeld = 0
		}
	}

	if holding.TotalCost <= 0 {
		return clampFloat(holding.ProfitRate, -20, 25), "按当前累计收益率估算", daysHeld
	}

	growth := 1 + holding.ProfitRate/100
	if growth <= 0 {
		return -100, "按累计亏损折算", daysHeld
	}

	switch {
	case daysHeld >= 90:
		return clampFloat(annualizeDayReturn(holding.ProfitRate, daysHeld), -60, 80),
			fmt.Sprintf("按持有 %d 天累计收益率年化估算", daysHeld),
			daysHeld
	case daysHeld >= 30:
		annualized := annualizeDayReturn(holding.ProfitRate, daysHeld)
		blended := annualized*0.6 + holding.ProfitRate*0.4
		return clampFloat(blended, -45, 60),
			fmt.Sprintf("按持有 %d 天收益率与当前累计收益折中估算", daysHeld),
			daysHeld
	case daysHeld > 0:
		blended := holding.ProfitRate*2.2 + holding.TodayRate*3
		return clampFloat(blended, -35, 45),
			fmt.Sprintf("按持有 %d 天短期收益节奏估算", daysHeld),
			daysHeld
	default:
		return clampFloat(holding.ProfitRate, -25, 35), "按当前累计收益率估算", daysHeld
	}
}

func buildExpectationBuckets(items []PortfolioExpectationItem, totalValue float64) []PortfolioExpectationBucket {
	type bucketAccumulator struct {
		label  string
		value  float64
		profit float64
		count  int
	}

	acc := map[string]*bucketAccumulator{}
	for _, item := range items {
		key := item.Bucket
		if strings.TrimSpace(key) == "" {
			key = "growth"
		}
		if acc[key] == nil {
			acc[key] = &bucketAccumulator{label: expectationBucketLabel(key)}
		}
		acc[key].value += item.TotalValue
		acc[key].profit += item.EstimatedAnnualProfit
		acc[key].count++
	}

	keys := []string{"conservative", "growth", "stock"}
	result := make([]PortfolioExpectationBucket, 0, len(keys))
	for _, key := range keys {
		item := acc[key]
		if item == nil || item.value <= 0 {
			continue
		}
		bucket := PortfolioExpectationBucket{
			Key:                   key,
			Label:                 item.label,
			Value:                 roundMoney(item.value),
			EstimatedAnnualProfit: roundMoney(item.profit),
			Count:                 item.count,
		}
		if totalValue > 0 {
			bucket.Weight = roundPercent(item.value / totalValue * 100)
		}
		if item.value > 0 {
			bucket.EstimatedAnnualReturnRate = roundPercent(item.profit / item.value * 100)
		}
		result = append(result, bucket)
	}
	return result
}

func buildExpectationWarnings(summary *PortfolioExpectationSummary) []string {
	if summary == nil {
		return nil
	}

	warnings := make([]string, 0, 6)
	switch {
	case summary.HouseholdLiquidAssets <= 0:
		warnings = append(warnings, "家庭资产里的流动资产还没录完整，当前收益目标只能按持仓口径粗略估算。")
	case summary.InvestedValue > summary.HouseholdLiquidAssets*1.05:
		warnings = append(warnings, "当前持仓市值已经高于家庭流动资产口径，建议确认家庭账户里是否已把投资账户余额录入。")
	case summary.IdleLiquidAssets > 0 && summary.HouseholdLiquidAssets > 0 && summary.IdleLiquidAssets/summary.HouseholdLiquidAssets >= 0.35:
		warnings = append(warnings, "仍有较多流动资产没有进入当前持仓，年度目标是否达成会明显受闲置资金影响。")
	}

	if summary.TargetAnnualReturnRate <= 0 {
		warnings = append(warnings, "还没有设置目标年收益率，AI 只能先按当前持仓做结构判断。")
	} else if summary.TargetAnnualReturnRate >= 12 {
		warnings = append(warnings, "12% 以上的家庭年度目标通常已经属于高挑战区间，往往需要更高权益暴露或更强波动承受能力。")
	}

	if summary.RequiredReturnOnIdleLiquidAssets >= 15 {
		warnings = append(warnings, "如果仅靠剩余闲置流动资产去补足收益缺口，所需收益率已经偏高，单靠固收类资产通常很难覆盖。")
	}

	if summary.ProjectedCompletionRatio > 0 && summary.ProjectedCompletionRatio < 60 {
		warnings = append(warnings, "按当前持仓结构推算，年度目标达成度偏低，更适合尽早调整而不是继续被动等待。")
	}
	if math.Abs(summary.AnnualUntrackedProfit) > 0.005 {
		warnings = append(warnings, fmt.Sprintf("已手动补充 %.2f 元年度未计入收益，年度缺口和达成度已按该口径一起计算。", summary.AnnualUntrackedProfit))
	}
	return warnings
}

func calcSuggestedFixedIncomeRange(summary *PortfolioExpectationSummary) (float64, float64, float64) {
	if summary == nil || summary.HouseholdLiquidAssets <= 0 {
		return 0, 0, 0
	}

	conservativeRate := summary.ConservativeExpectedReturnRate
	growthRate := summary.GrowthExpectedReturnRate
	if growthRate <= 0 {
		growthRate = maxFloat64(summary.EstimatedHoldingsAnnualReturnRate, conservativeRate)
	}

	ratio := 0.0
	switch {
	case summary.TargetAnnualReturnRate <= 0:
		ratio = summary.ConservativeRatio
	case summary.TargetAnnualReturnRate <= conservativeRate:
		ratio = 100
	case growthRate <= conservativeRate:
		ratio = 0
	default:
		ratio = (growthRate - summary.TargetAnnualReturnRate) / (growthRate - conservativeRate) * 100
	}

	ratio = roundPercent(clampFloat(ratio, 0, 100))
	fixedIncomeAmount := roundMoney(summary.HouseholdLiquidAssets * ratio / 100)
	growthAmount := roundMoney(maxFloat64(summary.HouseholdLiquidAssets-fixedIncomeAmount, 0))
	return ratio, fixedIncomeAmount, growthAmount
}

func annualizeMonthlyReturn(rate float64, months int) float64 {
	if months <= 0 {
		return 0
	}
	decimal := 1 + rate/100
	if decimal <= 0 {
		return -100
	}
	return (math.Pow(decimal, 12/float64(months)) - 1) * 100
}

func annualizeDayReturn(rate float64, days int) float64 {
	if days <= 0 {
		return 0
	}
	decimal := 1 + rate/100
	if decimal <= 0 {
		return -100
	}
	return (math.Pow(decimal, 365/float64(days)) - 1) * 100
}

func expectationDifficultyLabel(rate float64) string {
	switch {
	case rate <= 0:
		return "未设定"
	case rate <= 4:
		return "稳健"
	case rate <= 8:
		return "平衡"
	case rate <= 12:
		return "进取"
	case rate <= 18:
		return "高挑战"
	default:
		return "极高挑战"
	}
}

func expectationBucketForFund(category string) string {
	switch strings.ToLower(strings.TrimSpace(category)) {
	case "bond", "cash":
		return "conservative"
	default:
		return "growth"
	}
}

func expectationBucketLabel(bucket string) string {
	switch strings.ToLower(strings.TrimSpace(bucket)) {
	case "conservative":
		return "固收 / 现金类基金"
	case "stock":
		return "股票持仓"
	default:
		return "权益 / 主题基金"
	}
}

func sumExpectationItemValue(items []PortfolioExpectationItem) float64 {
	total := 0.0
	for _, item := range items {
		total += item.TotalValue
	}
	return total
}

func yearProgressRatio(now time.Time) float64 {
	start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	end := start.AddDate(1, 0, 0)
	totalDays := end.Sub(start).Hours() / 24
	if totalDays <= 0 {
		return 0
	}
	elapsedDays := now.Sub(start).Hours() / 24
	if elapsedDays < 0 {
		elapsedDays = 0
	}
	if elapsedDays > totalDays {
		elapsedDays = totalDays
	}
	return elapsedDays / totalDays
}

func roundMoney(value float64) float64 {
	return math.Round(value*100) / 100
}

func clampFloat(value float64, low float64, high float64) float64 {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func maxFloat64(left float64, right float64) float64 {
	if left > right {
		return left
	}
	return right
}

func minInt(left int, right int) int {
	if left < right {
		return left
	}
	return right
}
