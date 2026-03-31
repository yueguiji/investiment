package asset

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-stock/backend/db"
	"go-stock/backend/logger"
)

type HouseholdBenchmark struct {
	Name        string  `json:"name"`
	Value       float64 `json:"value"`
	Unit        string  `json:"unit"`
	Scope       string  `json:"scope"`
	Region      string  `json:"region"`
	Category    string  `json:"category"`
	Year        int     `json:"year"`
	Version     string  `json:"version"`
	Description string  `json:"description"`
}

type HouseholdAIContext struct {
	GeneratedAt             time.Time                              `json:"generatedAt"`
	Region                  string                                 `json:"region"`
	BenchmarkVersion        string                                 `json:"benchmarkVersion"`
	Profile                 *HouseholdProfile                      `json:"profile"`
	Members                 []HouseholdMember                      `json:"members"`
	MemberProfiles          []HouseholdMemberProfile               `json:"memberProfiles"`
	Benchmarks              []HouseholdBenchmark                   `json:"benchmarks"`
	Summary                 *HouseholdDashboardSummary             `json:"summary"`
	Accounts                []HouseholdAccount                     `json:"accounts"`
	FixedAssets             []HouseholdFixedAsset                  `json:"fixedAssets"`
	Incomes                 []HouseholdIncome                      `json:"incomes"`
	Protections             []HouseholdProtection                  `json:"protections"`
	Liabilities             []HouseholdLiability                   `json:"liabilities"`
	Snapshots               []HouseholdSnapshot                    `json:"snapshots"`
	LiabilityTrend          []HouseholdLiabilityTrendPoint         `json:"liabilityTrend"`
	LiquidAssetTrend        []HouseholdLiquidAssetTrendPoint       `json:"liquidAssetTrend"`
	LiquidAssetDistribution []HouseholdLiquidAssetDistributionItem `json:"liquidAssetDistribution"`
	AssetDetails            []HouseholdAssetBreakdown              `json:"assetDetails"`
	TopAssetDetails         []HouseholdAssetBreakdown              `json:"topAssetDetails"`
	IncomeDetails           []HouseholdIncomeBreakdown             `json:"incomeDetails"`
	LiabilityDetails        []HouseholdLiabilityBreakdown          `json:"liabilityDetails"`
}

type HouseholdMemberProfile struct {
	Name             string                    `json:"name"`
	Relationship     string                    `json:"relationship"`
	Gender           string                    `json:"gender"`
	Age              int                       `json:"age"`
	Occupation       string                    `json:"occupation"`
	City             string                    `json:"city"`
	AnnualIncome     float64                   `json:"annualIncome"`
	MonthlyIncome    float64                   `json:"monthlyIncome"`
	ProtectionStatus HouseholdProtectionStatus `json:"protectionStatus"`
}

type HouseholdProtectionStatus struct {
	HasSocialInsurance bool     `json:"hasSocialInsurance"`
	HasHousingFund     bool     `json:"hasHousingFund"`
	SocialInsurance    []string `json:"socialInsurance"`
	CommercialCoverage []string `json:"commercialCoverage"`
	MonthlyPersonal    float64  `json:"monthlyPersonal"`
	MonthlyEmployer    float64  `json:"monthlyEmployer"`
	CurrentBalance     float64  `json:"currentBalance"`
}

type HouseholdAssetBreakdown struct {
	Name         string  `json:"name"`
	Category     string  `json:"category"`
	Owner        string  `json:"owner"`
	Value        float64 `json:"value"`
	ShareOfTotal float64 `json:"shareOfTotal"`
	Provider     string  `json:"provider,omitempty"`
	Location     string  `json:"location,omitempty"`
}

type HouseholdIncomeBreakdown struct {
	Name              string  `json:"name"`
	Type              string  `json:"type"`
	Owner             string  `json:"owner"`
	Employer          string  `json:"employer"`
	MonthlyGross      float64 `json:"monthlyGross"`
	MonthlyNet        float64 `json:"monthlyNet"`
	MonthlyTax        float64 `json:"monthlyTax"`
	PretaxDeduction   float64 `json:"pretaxDeduction"`
	OtherPretaxDeduct float64 `json:"otherPretaxDeduction"`
	SpecialDeduction  float64 `json:"specialDeduction"`
	BasicDeduction    float64 `json:"basicDeduction"`
	TaxableIncome     float64 `json:"taxableIncome"`
	InsurancePersonal float64 `json:"insurancePersonal"`
	HousingFundPerson float64 `json:"housingFundPersonal"`
	FormulaText       string  `json:"formulaText"`
}

type HouseholdLiabilityBreakdown struct {
	Name                 string  `json:"name"`
	Type                 string  `json:"type"`
	Owner                string  `json:"owner"`
	Lender               string  `json:"lender"`
	OutstandingPrincipal float64 `json:"outstandingPrincipal"`
	MonthlyPayment       float64 `json:"monthlyPayment"`
	AnnualRate           float64 `json:"annualRate"`
	RemainingMonths      int     `json:"remainingMonths"`
	RepaymentMethod      string  `json:"repaymentMethod"`
}

func (s *Service) InitDefaultHouseholdBenchmarks() {
	defaults := []HouseholdBenchmarkRecord{
		{
			Name:        "全国居民人均可支配收入",
			Scope:       "全国",
			Region:      "全国",
			Category:    "income",
			Value:       43377,
			Unit:        "元/年",
			Year:        2025,
			Version:     "built-in-2026.03",
			Description: "国家统计局 2025 年居民收入口径，用于家庭收入水平全国对比。",
			IsActive:    true,
		},
		{
			Name:        "全国居民人均消费支出",
			Scope:       "全国",
			Region:      "全国",
			Category:    "income",
			Value:       29476,
			Unit:        "元/年",
			Year:        2025,
			Version:     "built-in-2026.03",
			Description: "国家统计局 2025 年居民消费支出口径，可与家庭年支出做对照。",
			IsActive:    true,
		},
		{
			Name:        "全国住户存款余额",
			Scope:       "全国",
			Region:      "全国",
			Category:    "asset",
			Value:       1620255,
			Unit:        "亿元",
			Year:        2025,
			Version:     "built-in-2026.03",
			Description: "2025 年二季度全国住户存款余额，用作家庭存款环境参考。",
			IsActive:    true,
		},
		{
			Name:        "全国平均家庭存款",
			Scope:       "全国",
			Region:      "全国",
			Category:    "asset",
			Value:       327886,
			Unit:        "元/户",
			Year:        2025,
			Version:     "built-in-2026.03",
			Description: "按全国住户存款余额与第七次人口普查家庭户数折算的平均家庭存款估算值。",
			IsActive:    true,
		},
		{
			Name:        "全国住户贷款余额",
			Scope:       "全国",
			Region:      "全国",
			Category:    "debt",
			Value:       840100,
			Unit:        "亿元",
			Year:        2025,
			Version:     "built-in-2026.03",
			Description: "2025 年二季度全国住户贷款余额，用作居民负债环境参考。",
			IsActive:    true,
		},
		{
			Name:        "全国负债率参考",
			Scope:       "全国",
			Region:      "全国",
			Category:    "debt",
			Value:       51.85,
			Unit:        "%",
			Year:        2025,
			Version:     "built-in-2026.03",
			Description: "按全国住户贷款余额与住户存款余额计算的代理负债率，用于全国居民杠杆观察。",
			IsActive:    true,
		},
		{
			Name:        "天津居民人均可支配收入",
			Scope:       "地区",
			Region:      "天津市",
			Category:    "income",
			Value:       46361,
			Unit:        "元/年",
			Year:        2025,
			Version:     "built-in-2026.03",
			Description: "天津市 2025 年居民人均可支配收入，用于地区家庭收入水平对照。",
			IsActive:    true,
		},
		{
			Name:        "天津居民人均消费支出",
			Scope:       "地区",
			Region:      "天津市",
			Category:    "income",
			Value:       32421,
			Unit:        "元/年",
			Year:        2025,
			Version:     "built-in-2026.03",
			Description: "天津市 2025 年居民人均消费支出，可与家庭年支出做地区对照。",
			IsActive:    true,
		},
		{
			Name:        "天津金融机构人民币存款余额",
			Scope:       "地区",
			Region:      "天津市",
			Category:    "asset",
			Value:       50006.61,
			Unit:        "亿元",
			Year:        2025,
			Version:     "built-in-2026.03",
			Description: "天津市金融机构人民币各项存款余额，属于区域金融环境代理指标，不等同于纯住户口径。",
			IsActive:    true,
		},
		{
			Name:        "天津平均家庭存款",
			Scope:       "地区",
			Region:      "天津市",
			Category:    "asset",
			Value:       1027484,
			Unit:        "元/户",
			Year:        2025,
			Version:     "built-in-2026.03",
			Description: "按天津金融机构人民币存款余额与第七次人口普查家庭户数折算的平均家庭存款代理值。",
			IsActive:    true,
		},
		{
			Name:        "天津金融机构人民币贷款余额",
			Scope:       "地区",
			Region:      "天津市",
			Category:    "debt",
			Value:       48285.29,
			Unit:        "亿元",
			Year:        2025,
			Version:     "built-in-2026.03",
			Description: "天津市金融机构人民币各项贷款余额，属于区域金融环境代理指标，不等同于纯住户口径。",
			IsActive:    true,
		},
		{
			Name:        "天津负债率参考",
			Scope:       "地区",
			Region:      "天津市",
			Category:    "debt",
			Value:       96.56,
			Unit:        "%",
			Year:        2025,
			Version:     "built-in-2026.03",
			Description: "按天津金融机构贷款余额与存款余额计算的代理杠杆指标，用于地区债务压力观察。",
			IsActive:    true,
		},
	}

	for _, item := range defaults {
		record := HouseholdBenchmarkRecord{}
		if err := db.Dao.
			Where(&HouseholdBenchmarkRecord{
				Name:   item.Name,
				Region: item.Region,
				Year:   item.Year,
			}).
			Assign(HouseholdBenchmarkRecord{
				Scope:       item.Scope,
				Category:    item.Category,
				Value:       item.Value,
				Unit:        item.Unit,
				Version:     item.Version,
				Description: item.Description,
				IsActive:    item.IsActive,
			}).
			FirstOrCreate(&record).Error; err != nil {
			logger.SugaredLogger.Errorf("seed household benchmark failed: %v", err)
		}
	}
}

func (s *Service) GetHouseholdBenchmarks(region string) []HouseholdBenchmarkRecord {
	query := db.Dao.Model(&HouseholdBenchmarkRecord{}).Where("is_active = ?", true)
	if region != "" {
		query = query.Where("region = ? OR region = ?", region, "全国")
	}

	var items []HouseholdBenchmarkRecord
	query.Order("year desc, scope asc, name asc").Find(&items)
	return items
}

func (s *Service) UpsertHouseholdBenchmark(item HouseholdBenchmarkRecord) *HouseholdBenchmarkRecord {
	if item.Version == "" {
		item.Version = "manual"
	}
	if err := db.Dao.Save(&item).Error; err != nil {
		logger.SugaredLogger.Errorf("upsert household benchmark failed: %v", err)
		return nil
	}
	return &item
}

func (s *Service) DeleteHouseholdBenchmark(id uint) bool {
	if err := db.Dao.Delete(&HouseholdBenchmarkRecord{}, id).Error; err != nil {
		logger.SugaredLogger.Errorf("delete household benchmark failed: %v", err)
		return false
	}
	return true
}

func (s *Service) GetHouseholdProfile() *HouseholdProfile {
	var item HouseholdProfile
	if err := db.Dao.Order("updated_at desc, id desc").First(&item).Error; err != nil {
		return &HouseholdProfile{
			HouseholdName:                        "我的家庭",
			Region:                               "天津市",
			CityTier:                             "新一线",
			MembersCount:                         2,
			DependentsCount:                      0,
			HousingStatus:                        "自有住房",
			RiskPreference:                       "稳健",
			MonthlyHouseholdSpend:                0,
			AnnualHouseholdSpend:                 0,
			PrimaryIncomeSource:                  "工资收入",
			MonthlyPersonalInsuranceContribution: 0,
			MonthlyHousingFundContribution:       0,
			MonthlyOtherPretaxDeduction:          0,
			MonthlyChildcareDeduction:            0,
			MonthlyHousingLoanDeduction:          0,
			MonthlyElderlyCareDeduction:          0,
			MonthlyOtherSpecialDeduction:         0,
			Notes:                                "",
		}
	}
	if item.Region == "" {
		item.Region = "天津市"
	}
	if item.HouseholdName == "" {
		item.HouseholdName = "我的家庭"
	}
	if item.MembersCount <= 0 {
		item.MembersCount = len(s.GetHouseholdMembers())
		if item.MembersCount <= 0 {
			item.MembersCount = 1
		}
	}
	return &item
}

func (s *Service) UpsertHouseholdProfile(item HouseholdProfile) *HouseholdProfile {
	if item.Region == "" {
		item.Region = "天津市"
	}
	if item.HouseholdName == "" {
		item.HouseholdName = "我的家庭"
	}
	if item.MembersCount <= 0 {
		item.MembersCount = 1
	}
	if item.DependentsCount < 0 {
		item.DependentsCount = 0
	}
	if item.MonthlyHouseholdSpend < 0 {
		item.MonthlyHouseholdSpend = 0
	}
	if item.AnnualHouseholdSpend < 0 {
		item.AnnualHouseholdSpend = 0
	}
	if item.MonthlyHouseholdSpend <= 0 && item.AnnualHouseholdSpend > 0 {
		item.MonthlyHouseholdSpend = roundTo(item.AnnualHouseholdSpend/12, 2)
	}
	if item.AnnualHouseholdSpend <= 0 && item.MonthlyHouseholdSpend > 0 {
		item.AnnualHouseholdSpend = roundTo(item.MonthlyHouseholdSpend*12, 2)
	}
	if err := db.Dao.Save(&item).Error; err != nil {
		logger.SugaredLogger.Errorf("upsert household profile failed: %v", err)
		return nil
	}
	return &item
}

func (s *Service) BuildHouseholdAIContext(region string) *HouseholdAIContext {
	profile := s.GetHouseholdProfile()
	if region == "" {
		region = profile.Region
	}
	members := s.GetHouseholdMembers()
	if profile.MembersCount <= 0 && len(members) > 0 {
		profile.MembersCount = len(members)
	}

	benchmarks := s.GetHouseholdBenchmarks(region)
	accounts := s.GetHouseholdAccounts()
	fixedAssets := s.GetHouseholdFixedAssets()
	incomes := s.GetHouseholdIncomes()
	protections := s.GetHouseholdProtections()
	liabilities := s.GetHouseholdLiabilities()
	summary := s.GetHouseholdDashboardSummary()
	assetDetails := buildAllAssetDetails(accounts, fixedAssets, summary.TotalAssets)
	contextBenchmarks := make([]HouseholdBenchmark, 0, len(benchmarks))
	benchmarkVersion := "manual"
	for _, item := range benchmarks {
		contextBenchmarks = append(contextBenchmarks, HouseholdBenchmark{
			Name:        item.Name,
			Value:       item.Value,
			Unit:        item.Unit,
			Scope:       item.Scope,
			Region:      item.Region,
			Category:    item.Category,
			Year:        item.Year,
			Version:     item.Version,
			Description: item.Description,
		})
		if item.Version != "" && benchmarkVersion == "manual" {
			benchmarkVersion = item.Version
		}
	}

	return &HouseholdAIContext{
		GeneratedAt:             time.Now(),
		Region:                  region,
		BenchmarkVersion:        benchmarkVersion,
		Profile:                 profile,
		Members:                 members,
		MemberProfiles:          buildHouseholdMemberProfiles(members, incomes, protections),
		Benchmarks:              contextBenchmarks,
		Summary:                 summary,
		Accounts:                accounts,
		FixedAssets:             fixedAssets,
		Incomes:                 incomes,
		Protections:             protections,
		Liabilities:             liabilities,
		Snapshots:               s.GetHouseholdSnapshots(180),
		LiabilityTrend:          s.GetHouseholdLiabilityTrend(0, 23),
		LiquidAssetTrend:        s.GetHouseholdLiquidAssetTrend(180),
		LiquidAssetDistribution: s.GetHouseholdLiquidAssetDistribution(),
		AssetDetails:            assetDetails,
		TopAssetDetails:         pickTopEightyPercentAssets(assetDetails, summary.TotalAssets),
		IncomeDetails:           buildIncomeDetails(incomes, *profile),
		LiabilityDetails:        buildLiabilityDetails(liabilities),
	}
}

func (s *Service) BuildHouseholdAIContextJSON(region string) string {
	payload, err := json.MarshalIndent(s.BuildHouseholdAIContext(region), "", "  ")
	if err != nil {
		logger.SugaredLogger.Errorf("marshal household ai context failed: %v", err)
		return "{}"
	}
	return string(payload)
}

func (s *Service) SaveHouseholdAIAnalysis(item HouseholdAIAnalysis) *HouseholdAIAnalysis {
	if err := db.Dao.Create(&item).Error; err != nil {
		logger.SugaredLogger.Errorf("save household ai analysis failed: %v", err)
		return nil
	}
	return &item
}

func (s *Service) GetLatestHouseholdAIAnalysis() *HouseholdAIAnalysis {
	var item HouseholdAIAnalysis
	if err := db.Dao.Order("created_at desc, id desc").First(&item).Error; err != nil {
		return nil
	}
	return &item
}

func buildHouseholdMemberProfiles(members []HouseholdMember, incomes []HouseholdIncome, protections []HouseholdProtection) []HouseholdMemberProfile {
	result := make([]HouseholdMemberProfile, 0, len(members))
	for _, member := range members {
		status := HouseholdProtectionStatus{}
		for _, item := range protections {
			if item.Owner != member.Name {
				continue
			}
			status.MonthlyPersonal += item.MonthlyPersonalContribution
			status.MonthlyEmployer += item.MonthlyEmployerContribution
			status.CurrentBalance += item.CurrentBalance
			switch item.ProtectionType {
			case "social_insurance":
				status.HasSocialInsurance = true
				status.SocialInsurance = append(status.SocialInsurance, item.Name)
			case "housing_fund":
				status.HasHousingFund = true
				status.SocialInsurance = append(status.SocialInsurance, item.Name)
			default:
				status.CommercialCoverage = append(status.CommercialCoverage, item.Name)
			}
		}
		result = append(result, HouseholdMemberProfile{
			Name:             member.Name,
			Relationship:     member.Relationship,
			Gender:           member.Gender,
			Age:              calculateAge(member.BirthDate),
			Occupation:       member.Occupation,
			City:             member.City,
			AnnualIncome:     member.AnnualIncome,
			MonthlyIncome:    sumMemberMonthlyIncome(member.Name, incomes),
			ProtectionStatus: status,
		})
	}
	return result
}

func buildAllAssetDetails(accounts []HouseholdAccount, fixedAssets []HouseholdFixedAsset, totalAssets float64) []HouseholdAssetBreakdown {
	items := make([]HouseholdAssetBreakdown, 0, len(accounts)+len(fixedAssets))
	for _, item := range accounts {
		if item.Balance <= 0 {
			continue
		}
		items = append(items, HouseholdAssetBreakdown{
			Name:         item.Name,
			Category:     item.AccountType,
			Owner:        item.Owner,
			Value:        item.Balance,
			ShareOfTotal: shareOfTotal(item.Balance, totalAssets),
			Provider:     item.Provider,
		})
	}
	for _, item := range fixedAssets {
		if item.CurrentValue <= 0 {
			continue
		}
		items = append(items, HouseholdAssetBreakdown{
			Name:         item.Name,
			Category:     item.AssetType,
			Owner:        item.Owner,
			Value:        item.CurrentValue,
			ShareOfTotal: shareOfTotal(item.CurrentValue, totalAssets),
			Location:     item.Location,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Value > items[j].Value
	})
	return items
}

func buildIncomeDetails(incomes []HouseholdIncome, profile HouseholdProfile) []HouseholdIncomeBreakdown {
	result := make([]HouseholdIncomeBreakdown, 0, len(incomes))
	for _, item := range incomes {
		monthlyGross := normalizeIncomeToMonthly(item)
		pretax := item.MonthlyPersonalInsuranceContribution + item.MonthlyPersonalHousingFundContribution
		monthlyNet := monthlyGross
		monthlyTax := 0.0
		otherPretax := 0.0
		special := 0.0
		basicDeduction := 0.0
		taxableIncome := 0.0
		formulaText := ""
		if item.IncomeType == "salary" {
			otherPretax = roundTo(NumberOrZero(profile.MonthlyOtherPretaxDeduction), 2)
			special = roundTo(calculateNormalizedMonthlySpecialDeductions(&profile), 2)
			basicDeduction = 5000
			taxableIncome = roundTo(maxFloat(monthlyGross-pretax-otherPretax-special-basicDeduction, 0), 2)
			monthlyTax = calculateMonthlyComprehensiveTax(monthlyGross, pretax+otherPretax, special)
			monthlyNet = monthlyGross - pretax - monthlyTax
			formulaText = buildSalaryIncomeFormulaText(monthlyGross, pretax, otherPretax, special, basicDeduction, taxableIncome, monthlyTax, monthlyNet)
		} else if item.IncomeType == "bonus" {
			monthlyTax = calculateMonthlyBonusTax(monthlyGross)
			monthlyNet = monthlyGross - monthlyTax
			taxableIncome = roundTo(maxFloat(monthlyGross, 0), 2)
			formulaText = buildBonusIncomeFormulaText(monthlyGross, monthlyTax, monthlyNet)
		} else {
			taxableIncome = roundTo(maxFloat(monthlyGross, 0), 2)
			formulaText = buildOtherIncomeFormulaText(monthlyGross)
		}
		result = append(result, HouseholdIncomeBreakdown{
			Name:              item.Name,
			Type:              item.IncomeType,
			Owner:             item.Owner,
			Employer:          item.Employer,
			MonthlyGross:      roundTo(monthlyGross, 2),
			MonthlyNet:        roundTo(monthlyNet, 2),
			MonthlyTax:        roundTo(monthlyTax, 2),
			PretaxDeduction:   roundTo(pretax, 2),
			OtherPretaxDeduct: otherPretax,
			SpecialDeduction:  special,
			BasicDeduction:    basicDeduction,
			TaxableIncome:     taxableIncome,
			InsurancePersonal: roundTo(item.MonthlyPersonalInsuranceContribution, 2),
			HousingFundPerson: roundTo(item.MonthlyPersonalHousingFundContribution, 2),
			FormulaText:       formulaText,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].MonthlyGross > result[j].MonthlyGross
	})
	return result
}

func buildSalaryIncomeFormulaText(monthlyGross, pretax, otherPretax, special, basicDeduction, taxableIncome, monthlyTax, monthlyNet float64) string {
	return "税后 = 税前" +
		formatCompactMoney(monthlyGross) +
		" - 个人五险一金" + formatCompactMoney(pretax) +
		" - 其他税前扣除" + formatCompactMoney(otherPretax) +
		" - 专项附加扣除" + formatCompactMoney(special) +
		" - 基本减除费用" + formatCompactMoney(basicDeduction) +
		" => 应纳税所得额" + formatCompactMoney(taxableIncome) +
		" => 个税" + formatCompactMoney(monthlyTax) +
		" => 税后" + formatCompactMoney(monthlyNet)
}

func buildBonusIncomeFormulaText(monthlyGross, monthlyTax, monthlyNet float64) string {
	return "税后 = 税前" +
		formatCompactMoney(monthlyGross) +
		" - 个税" + formatCompactMoney(monthlyTax) +
		" => 税后" + formatCompactMoney(monthlyNet)
}

func buildOtherIncomeFormulaText(monthlyGross float64) string {
	return "税后 = 税前" + formatCompactMoney(monthlyGross) + "（当前按未额外扣税口径计入）"
}

func formatCompactMoney(value float64) string {
	return "¥" + formatFloatTrimmed(roundTo(value, 2))
}

func formatFloatTrimmed(value float64) string {
	text := strconv.FormatFloat(value, 'f', 2, 64)
	text = strings.TrimRight(text, "0")
	text = strings.TrimRight(text, ".")
	if text == "" {
		return "0"
	}
	return text
}

func buildLiabilityDetails(liabilities []HouseholdLiability) []HouseholdLiabilityBreakdown {
	result := make([]HouseholdLiabilityBreakdown, 0, len(liabilities))
	for _, item := range liabilities {
		result = append(result, HouseholdLiabilityBreakdown{
			Name:                 item.Name,
			Type:                 item.LiabilityType,
			Owner:                item.Owner,
			Lender:               item.Lender,
			OutstandingPrincipal: roundTo(item.OutstandingPrincipal, 2),
			MonthlyPayment:       roundTo(item.MonthlyPayment, 2),
			AnnualRate:           roundTo(item.AnnualRate, 4),
			RemainingMonths:      item.LoanTermMonths,
			RepaymentMethod:      item.RepaymentMethod,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].OutstandingPrincipal > result[j].OutstandingPrincipal
	})
	return result
}

func pickTopEightyPercentAssets(items []HouseholdAssetBreakdown, totalAssets float64) []HouseholdAssetBreakdown {
	if len(items) == 0 {
		return items
	}
	if totalAssets <= 0 {
		if len(items) > 5 {
			return items[:5]
		}
		return items
	}
	threshold := totalAssets * 0.8
	sum := 0.0
	limit := 0
	for index, item := range items {
		sum += item.Value
		limit = index + 1
		if sum >= threshold {
			break
		}
	}
	if limit < len(items) && limit < 3 {
		limit = 3
	}
	if limit > len(items) {
		limit = len(items)
	}
	return items[:limit]
}

func shareOfTotal(value float64, total float64) float64 {
	if value <= 0 || total <= 0 {
		return 0
	}
	return roundTo(value/total*100, 2)
}

func sumMemberMonthlyIncome(owner string, incomes []HouseholdIncome) float64 {
	total := 0.0
	for _, item := range incomes {
		if item.Owner != owner {
			continue
		}
		total += normalizeIncomeToMonthly(item)
	}
	return roundTo(total, 2)
}

func normalizeIncomeToMonthly(item HouseholdIncome) float64 {
	if item.MonthlyAmount > 0 {
		return item.MonthlyAmount
	}
	if item.AnnualAmount > 0 {
		return item.AnnualAmount / 12
	}
	return 0
}

func calculateAge(birthDate *time.Time) int {
	if birthDate == nil {
		return 0
	}
	now := time.Now()
	age := now.Year() - birthDate.Year()
	if now.YearDay() < birthDate.YearDay() {
		age--
	}
	if age < 0 {
		return 0
	}
	return age
}
