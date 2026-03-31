package asset

import (
	"math"
	"testing"
	"time"

	"go-stock/backend/db"
	"gorm.io/gorm"
)

func setupHouseholdTestDB(t *testing.T) {
	t.Helper()

	dbPath := "file:household-test-" + t.Name() + "?mode=memory&cache=shared"
	db.Init(dbPath)
	if err := db.Dao.AutoMigrate(
		&HouseholdAccount{},
		&HouseholdFixedAsset{},
		&HouseholdIncome{},
		&HouseholdProtection{},
		&HouseholdLiability{},
		&HouseholdLiabilitySchedule{},
		&HouseholdSnapshot{},
		&HouseholdAIAnalysis{},
		&HouseholdBenchmarkRecord{},
		&HouseholdProfile{},
		&HouseholdMember{},
	); err != nil {
		t.Fatalf("migrate test db failed: %v", err)
	}
}

func TestCalculateMonthlyPaymentEqualInstallment(t *testing.T) {
	service := NewService()

	payment := service.calculateMonthlyPayment(100000, 12, 12, "equal_installment")

	if math.Abs(payment-8884.88) > 0.02 {
		t.Fatalf("unexpected monthly payment: got %.2f", payment)
	}
}

func TestBuildLiabilityScheduleEqualPrincipal(t *testing.T) {
	service := NewService()
	firstPayment := time.Date(2026, 1, 10, 0, 0, 0, 0, time.Local)

	schedules := service.buildLiabilitySchedule(HouseholdLiability{
		Model:               gorm.Model{ID: 1},
		Principal:           120000,
		AnnualRate:          3.6,
		LoanTermMonths:      12,
		RepaymentMethod:     "equal_principal",
		ExtraMonthlyPayment: 0,
		FirstPaymentDate:    &firstPayment,
	})

	if len(schedules) != 12 {
		t.Fatalf("expected 12 schedule rows, got %d", len(schedules))
	}

	var principalPaid float64
	for _, item := range schedules {
		principalPaid += item.PrincipalPaid
	}

	if math.Abs(principalPaid-120000) > 1 {
		t.Fatalf("unexpected total principal paid: got %.2f", principalPaid)
	}

	last := schedules[len(schedules)-1]
	if math.Abs(last.ClosingPrincipal) > 0.01 {
		t.Fatalf("expected final closing principal to be zero, got %.2f", last.ClosingPrincipal)
	}
}

func TestBuildLiabilityScheduleEqualInstallmentDecreasesOutstanding(t *testing.T) {
	service := NewService()
	firstPayment := time.Date(2026, 1, 10, 0, 0, 0, 0, time.Local)

	schedules := service.buildLiabilitySchedule(HouseholdLiability{
		Model:               gorm.Model{ID: 2},
		Principal:           500000,
		AnnualRate:          4.2,
		LoanTermMonths:      24,
		RepaymentMethod:     "equal_installment",
		FirstPaymentDate:    &firstPayment,
		ExtraMonthlyPayment: 0,
	})

	if len(schedules) != 24 {
		t.Fatalf("expected 24 schedule rows, got %d", len(schedules))
	}

	prev := schedules[0].OpeningPrincipal
	for _, item := range schedules {
		if item.ClosingPrincipal > prev {
			t.Fatalf("outstanding principal increased: prev %.2f current %.2f", prev, item.ClosingPrincipal)
		}
		prev = item.ClosingPrincipal
	}

	if math.Abs(schedules[len(schedules)-1].ClosingPrincipal) > 0.5 {
		t.Fatalf("expected near-zero final outstanding, got %.2f", schedules[len(schedules)-1].ClosingPrincipal)
	}
}

func TestGetHouseholdLiabilityTrendCurrent24Months(t *testing.T) {
	setupHouseholdTestDB(t)
	service := NewService()

	startDate := time.Now()
	firstPayment := time.Date(startDate.Year(), startDate.Month(), 15, 0, 0, 0, 0, startDate.Location())
	item := service.CreateHouseholdLiability(HouseholdLiability{
		Name:                 "测试房贷",
		LiabilityType:        "mortgage",
		Lender:               "测试银行",
		Owner:                "测试",
		Principal:            300000,
		OutstandingPrincipal: 240000,
		AnnualRate:           3.1,
		LoanTermMonths:       24,
		RepaymentMethod:      "equal_installment",
		MonthlyPayment:       10319.33,
		FirstPaymentDate:     &firstPayment,
		StartDate:            &startDate,
		AutoAmortize:         true,
		IsActive:             true,
	})
	if item == nil {
		t.Fatal("expected liability to be created")
	}

	trend := service.GetHouseholdLiabilityTrend(0, 23)
	if len(trend) != 24 {
		t.Fatalf("expected 24 months trend, got %d", len(trend))
	}
	if trend[0].Month != time.Now().Format("2006-01") {
		t.Fatalf("expected first trend month %s, got %s", time.Now().Format("2006-01"), trend[0].Month)
	}
}

func TestDashboardSummaryCalculatesNetIncome(t *testing.T) {
	setupHouseholdTestDB(t)
	service := NewService()

	profile := service.UpsertHouseholdProfile(HouseholdProfile{
		HouseholdName:                        "税务测试",
		Region:                               "天津市",
		MembersCount:                         2,
		MonthlyPersonalInsuranceContribution: 1500,
		MonthlyHousingFundContribution:       1200,
		MonthlyChildcareDeduction:            1000,
		MonthlyHousingLoanDeduction:          1000,
		MonthlyElderlyCareDeduction:          1500,
	})
	if profile == nil {
		t.Fatal("expected household profile")
	}

	income := service.CreateHouseholdIncome(HouseholdIncome{
		Name:          "工资",
		IncomeType:    "salary",
		Frequency:     "monthly",
		MonthlyAmount: 25000,
		IsActive:      true,
	})
	if income == nil {
		t.Fatal("expected salary income")
	}
	bonus := service.CreateHouseholdIncome(HouseholdIncome{
		Name:          "奖金",
		IncomeType:    "bonus",
		Frequency:     "monthly",
		MonthlyAmount: 2000,
		IsActive:      true,
	})
	if bonus == nil {
		t.Fatal("expected bonus income")
	}

	summary := service.GetHouseholdDashboardSummary()
	if summary.MonthlyIncome != 27000 {
		t.Fatalf("expected gross income 27000, got %.2f", summary.MonthlyIncome)
	}
	if summary.MonthlyNetIncome <= 0 || summary.MonthlyNetIncome >= summary.MonthlyIncome {
		t.Fatalf("expected net income between 0 and gross, got net=%.2f gross=%.2f", summary.MonthlyNetIncome, summary.MonthlyIncome)
	}
	if summary.MonthlyIncomeTax <= 0 {
		t.Fatalf("expected income tax to be positive, got %.2f", summary.MonthlyIncomeTax)
	}
	if summary.MonthlyPretaxCosts != 2700 {
		t.Fatalf("expected pretax costs 2700, got %.2f", summary.MonthlyPretaxCosts)
	}
}

func TestDashboardSummaryNormalizesAnnualLikeSpecialDeductionAndEffectiveDebtPayment(t *testing.T) {
	setupHouseholdTestDB(t)
	service := NewService()

	profile := service.UpsertHouseholdProfile(HouseholdProfile{
		HouseholdName:                "专项抵扣测试",
		Region:                       "天津市",
		MembersCount:                 2,
		MonthlyOtherSpecialDeduction: 72000,
	})
	if profile == nil {
		t.Fatal("expected household profile")
	}

	income := service.CreateHouseholdIncome(HouseholdIncome{
		Name:                                   "工资",
		IncomeType:                             "salary",
		Frequency:                              "monthly",
		MonthlyAmount:                          25000,
		MonthlyPersonalHousingFundContribution: 1200,
		MonthlyEmployerHousingFundContribution: 1200,
		IsActive:                               true,
	})
	if income == nil {
		t.Fatal("expected income")
	}
	firstPayment := time.Date(2026, 1, 10, 0, 0, 0, 0, time.Local)
	liability := service.CreateHouseholdLiability(HouseholdLiability{
		Name:                 "房贷",
		LiabilityType:        "mortgage",
		Principal:            300000,
		OutstandingPrincipal: 300000,
		AnnualRate:           0,
		LoanTermMonths:       360,
		RepaymentMethod:      "equal_installment",
		MonthlyPayment:       5000,
		FirstPaymentDate:     &firstPayment,
		AutoAmortize:         false,
		IsActive:             true,
	})
	if liability == nil {
		t.Fatal("expected liability")
	}

	summary := service.GetHouseholdDashboardSummary()
	if summary.MonthlyHousingFundInflows != 2400 {
		t.Fatalf("expected monthly housing fund inflows 2400, got %.2f", summary.MonthlyHousingFundInflows)
	}
	if summary.MonthlyEffectiveDebtPayment != 2600 {
		t.Fatalf("expected monthly effective debt payment 2600, got %.2f", summary.MonthlyEffectiveDebtPayment)
	}
	if summary.MonthlyIncomeTax <= 0 {
		t.Fatalf("expected positive income tax after normalizing annual-like deduction, got %.2f", summary.MonthlyIncomeTax)
	}
}

func TestDashboardSummaryDoesNotDoubleCountProfilePretaxOrOtherIncomeContributions(t *testing.T) {
	setupHouseholdTestDB(t)
	service := NewService()

	profile := service.UpsertHouseholdProfile(HouseholdProfile{
		HouseholdName:                        "家庭画像税前口径",
		Region:                               "天津市",
		MembersCount:                         2,
		MonthlyPersonalInsuranceContribution: 2286,
		MonthlyHousingFundContribution:       1089,
		MonthlyChildcareDeduction:            2000,
		MonthlyHousingLoanDeduction:          1500,
		MonthlyElderlyCareDeduction:          2500,
	})
	if profile == nil {
		t.Fatal("expected household profile")
	}

	salary := service.CreateHouseholdIncome(HouseholdIncome{
		Name:                                   "工资",
		IncomeType:                             "salary",
		Frequency:                              "monthly",
		MonthlyAmount:                          25000,
		MonthlyPersonalInsuranceContribution:   2286,
		MonthlyPersonalHousingFundContribution: 1089,
		IsActive:                               true,
	})
	if salary == nil {
		t.Fatal("expected salary income")
	}

	other := service.CreateHouseholdIncome(HouseholdIncome{
		Name:                                 "副业",
		IncomeType:                           "other",
		Frequency:                            "monthly",
		MonthlyAmount:                        2500,
		MonthlyPersonalInsuranceContribution: 1400,
		IsActive:                             true,
	})
	if other == nil {
		t.Fatal("expected other income")
	}

	bonus := service.CreateHouseholdIncome(HouseholdIncome{
		Name:          "奖金",
		IncomeType:    "bonus",
		Frequency:     "monthly",
		MonthlyAmount: 2083.33,
		IsActive:      true,
	})
	if bonus == nil {
		t.Fatal("expected bonus income")
	}

	summary := service.GetHouseholdDashboardSummary()
	if summary.MonthlyPretaxCosts != 3375 {
		t.Fatalf("expected pretax costs 3375 without double counting profile fields, got %.2f", summary.MonthlyPretaxCosts)
	}
	if summary.MonthlyIncomeTax != 915 {
		t.Fatalf("expected income tax 915.00, got %.2f", summary.MonthlyIncomeTax)
	}
	if summary.MonthlyNetIncome != 25293.33 {
		t.Fatalf("expected monthly net income 25293.33, got %.2f", summary.MonthlyNetIncome)
	}
}

func TestIncomeSyncsProtections(t *testing.T) {
	setupHouseholdTestDB(t)
	service := NewService()

	income := service.CreateHouseholdIncome(HouseholdIncome{
		Name:                                   "工资",
		IncomeType:                             "salary",
		Owner:                                  "张三",
		Employer:                               "测试公司",
		Frequency:                              "monthly",
		MonthlyAmount:                          20000,
		MonthlyPersonalInsuranceContribution:   1200,
		MonthlyEmployerInsuranceContribution:   2400,
		MonthlyPersonalHousingFundContribution: 800,
		MonthlyEmployerHousingFundContribution: 800,
		IsActive:                               true,
	})
	if income == nil {
		t.Fatal("expected income to be created")
	}

	protections := service.GetHouseholdProtections()
	if len(protections) != 2 {
		t.Fatalf("expected 2 linked protections, got %d", len(protections))
	}

	income.MonthlyPersonalHousingFundContribution = 0
	income.MonthlyEmployerHousingFundContribution = 0
	updated := service.UpdateHouseholdIncome(*income)
	if updated == nil {
		t.Fatal("expected income to be updated")
	}

	protections = service.GetHouseholdProtections()
	if len(protections) != 1 {
		t.Fatalf("expected 1 linked protection after removing housing fund, got %d", len(protections))
	}

	if !service.DeleteHouseholdIncome(income.ID) {
		t.Fatal("expected income to be deleted")
	}
	protections = service.GetHouseholdProtections()
	if len(protections) != 0 {
		t.Fatalf("expected linked protections to be deleted, got %d", len(protections))
	}
}

func TestProtectionUpdateSyncsBackToIncomeAndSummary(t *testing.T) {
	setupHouseholdTestDB(t)
	service := NewService()

	income := service.CreateHouseholdIncome(HouseholdIncome{
		Name:                                   "工资",
		IncomeType:                             "salary",
		Owner:                                  "张三",
		Employer:                               "测试公司",
		Frequency:                              "monthly",
		MonthlyAmount:                          25000,
		MonthlyPersonalHousingFundContribution: 1089,
		MonthlyEmployerHousingFundContribution: 0,
		IsActive:                               true,
	})
	if income == nil {
		t.Fatal("expected income to be created")
	}

	protections := service.GetHouseholdProtections()
	if len(protections) != 1 {
		t.Fatalf("expected 1 linked protection, got %d", len(protections))
	}

	housingFund := protections[0]
	housingFund.MonthlyEmployerContribution = 1089
	updatedProtection := service.UpdateHouseholdProtection(housingFund)
	if updatedProtection == nil {
		t.Fatal("expected housing fund protection to be updated")
	}

	var refreshed HouseholdIncome
	if err := db.Dao.First(&refreshed, income.ID).Error; err != nil {
		t.Fatalf("load refreshed income failed: %v", err)
	}
	if refreshed.MonthlyEmployerHousingFundContribution != 1089 {
		t.Fatalf("expected employer housing fund contribution to sync back as 1089, got %.2f", refreshed.MonthlyEmployerHousingFundContribution)
	}

	firstPayment := time.Date(2026, 1, 10, 0, 0, 0, 0, time.Local)
	liability := service.CreateHouseholdLiability(HouseholdLiability{
		Name:                 "房贷",
		LiabilityType:        "mortgage",
		Principal:            300000,
		OutstandingPrincipal: 300000,
		AnnualRate:           0,
		LoanTermMonths:       360,
		RepaymentMethod:      "equal_installment",
		MonthlyPayment:       5000,
		FirstPaymentDate:     &firstPayment,
		AutoAmortize:         false,
		IsActive:             true,
	})
	if liability == nil {
		t.Fatal("expected liability")
	}

	summary := service.GetHouseholdDashboardSummary()
	if summary.MonthlyHousingFundInflows != 2178 {
		t.Fatalf("expected monthly housing fund inflows 2178, got %.2f", summary.MonthlyHousingFundInflows)
	}
	if summary.MonthlyEffectiveDebtPayment != 2822 {
		t.Fatalf("expected monthly effective debt payment 2822, got %.2f", summary.MonthlyEffectiveDebtPayment)
	}
}

func TestDashboardSummaryMergesManualHousingFundProtectionByOwner(t *testing.T) {
	setupHouseholdTestDB(t)
	service := NewService()

	income := service.CreateHouseholdIncome(HouseholdIncome{
		Name:                                   "工资",
		IncomeType:                             "salary",
		Owner:                                  "张斌",
		Employer:                               "测试公司",
		Frequency:                              "monthly",
		MonthlyAmount:                          25000,
		MonthlyPersonalHousingFundContribution: 1089,
		MonthlyEmployerHousingFundContribution: 0,
		IsActive:                               true,
	})
	if income == nil {
		t.Fatal("expected income")
	}

	manualProtection := service.CreateHouseholdProtection(HouseholdProtection{
		Name:                        "单位 - 公积金",
		ProtectionType:              "housing_fund",
		Owner:                       "张斌",
		MonthlyPersonalContribution: 1089,
		MonthlyEmployerContribution: 1089,
		IsActive:                    true,
	})
	if manualProtection == nil {
		t.Fatal("expected manual housing fund protection")
	}

	firstPayment := time.Date(2026, 1, 10, 0, 0, 0, 0, time.Local)
	liability := service.CreateHouseholdLiability(HouseholdLiability{
		Name:                 "房贷",
		LiabilityType:        "mortgage",
		Principal:            300000,
		OutstandingPrincipal: 300000,
		AnnualRate:           0,
		LoanTermMonths:       360,
		RepaymentMethod:      "equal_installment",
		MonthlyPayment:       5000,
		FirstPaymentDate:     &firstPayment,
		AutoAmortize:         false,
		IsActive:             true,
	})
	if liability == nil {
		t.Fatal("expected liability")
	}

	summary := service.GetHouseholdDashboardSummary()
	if summary.MonthlyHousingFundInflows != 2178 {
		t.Fatalf("expected merged housing fund inflows 2178, got %.2f", summary.MonthlyHousingFundInflows)
	}
	if summary.MonthlyEffectiveDebtPayment != 2822 {
		t.Fatalf("expected effective debt payment 2822, got %.2f", summary.MonthlyEffectiveDebtPayment)
	}
}

func TestDeleteLinkedProtectionClearsIncomeContribution(t *testing.T) {
	setupHouseholdTestDB(t)
	service := NewService()

	income := service.CreateHouseholdIncome(HouseholdIncome{
		Name:                                   "工资",
		IncomeType:                             "salary",
		Owner:                                  "张三",
		Employer:                               "测试公司",
		Frequency:                              "monthly",
		MonthlyAmount:                          20000,
		MonthlyPersonalHousingFundContribution: 800,
		MonthlyEmployerHousingFundContribution: 800,
		IsActive:                               true,
	})
	if income == nil {
		t.Fatal("expected income to be created")
	}

	protections := service.GetHouseholdProtections()
	if len(protections) != 1 {
		t.Fatalf("expected 1 linked protection, got %d", len(protections))
	}

	if !service.DeleteHouseholdProtection(protections[0].ID) {
		t.Fatal("expected linked protection to be deleted")
	}

	var refreshed HouseholdIncome
	if err := db.Dao.First(&refreshed, income.ID).Error; err != nil {
		t.Fatalf("load refreshed income failed: %v", err)
	}
	if refreshed.MonthlyPersonalHousingFundContribution != 0 || refreshed.MonthlyEmployerHousingFundContribution != 0 {
		t.Fatalf("expected linked income housing fund contributions to be cleared, got personal=%.2f employer=%.2f",
			refreshed.MonthlyPersonalHousingFundContribution,
			refreshed.MonthlyEmployerHousingFundContribution,
		)
	}
}

func TestInitDefaultHouseholdBenchmarksAndBuildContext(t *testing.T) {
	setupHouseholdTestDB(t)
	service := NewService()

	profile := service.UpsertHouseholdProfile(HouseholdProfile{
		HouseholdName:         "测试家庭",
		Region:                "天津市",
		CityTier:              "新一线",
		MembersCount:          3,
		DependentsCount:       1,
		HousingStatus:         "按揭中",
		RiskPreference:        "稳健",
		MonthlyHouseholdSpend: 15000,
		AnnualHouseholdSpend:  180000,
		PrimaryIncomeSource:   "工资收入",
	})
	if profile == nil {
		t.Fatal("expected household profile to be saved")
	}

	service.InitDefaultHouseholdBenchmarks()
	benchmarks := service.GetHouseholdBenchmarks("天津市")
	if len(benchmarks) < 8 {
		t.Fatalf("expected seeded benchmarks, got %d", len(benchmarks))
	}
	foundDebtRatio := false
	foundDeposit := false
	for _, item := range benchmarks {
		if item.Name == "天津负债率参考" {
			foundDebtRatio = true
		}
		if item.Name == "全国平均家庭存款" {
			foundDeposit = true
		}
	}
	if !foundDebtRatio || !foundDeposit {
		t.Fatalf("expected seeded debt ratio and deposit benchmarks, got debtRatio=%v deposit=%v", foundDebtRatio, foundDeposit)
	}

	context := service.BuildHouseholdAIContext("天津市")
	if context.Region != "天津市" {
		t.Fatalf("unexpected context region: %s", context.Region)
	}
	if context.BenchmarkVersion == "" {
		t.Fatal("expected benchmark version to be set")
	}
	if context.Profile == nil {
		t.Fatal("expected household profile in context")
	}
	if context.Profile.HouseholdName != "测试家庭" {
		t.Fatalf("unexpected household name: %s", context.Profile.HouseholdName)
	}
	if context.Profile.MonthlyHouseholdSpend != 15000 {
		t.Fatalf("expected monthly household spend 15000, got %.2f", context.Profile.MonthlyHouseholdSpend)
	}
	if len(context.Benchmarks) == 0 {
		t.Fatal("expected context benchmarks to be populated")
	}
}

func TestGetHouseholdProfileDefault(t *testing.T) {
	setupHouseholdTestDB(t)
	service := NewService()

	profile := service.GetHouseholdProfile()
	if profile == nil {
		t.Fatal("expected default profile")
	}
	if profile.Region != "天津市" {
		t.Fatalf("unexpected default region: %s", profile.Region)
	}
	if profile.MembersCount <= 0 {
		t.Fatalf("expected positive default members count, got %d", profile.MembersCount)
	}
}

func TestUpsertHouseholdProfileSyncsMonthlyAndAnnualSpend(t *testing.T) {
	setupHouseholdTestDB(t)
	service := NewService()

	profile := service.UpsertHouseholdProfile(HouseholdProfile{
		HouseholdName:         "支出测试",
		Region:                "天津市",
		MembersCount:          2,
		MonthlyHouseholdSpend: 12000,
	})
	if profile == nil {
		t.Fatal("expected profile to be saved")
	}
	if profile.AnnualHouseholdSpend != 144000 {
		t.Fatalf("expected annual spend 144000, got %.2f", profile.AnnualHouseholdSpend)
	}
}

func TestSaveLatestHouseholdAIAnalysis(t *testing.T) {
	setupHouseholdTestDB(t)
	service := NewService()

	first := service.SaveHouseholdAIAnalysis(HouseholdAIAnalysis{
		TriggerSource:    "test:first",
		Region:           "天津市",
		BenchmarkVersion: "v1",
		Status:           "success",
		AnalysisMarkdown: "核心结论\n第一版",
	})
	if first == nil {
		t.Fatal("expected first analysis record")
	}

	second := service.SaveHouseholdAIAnalysis(HouseholdAIAnalysis{
		TriggerSource:    "test:second",
		Region:           "天津市",
		BenchmarkVersion: "v2",
		Status:           "success",
		AnalysisMarkdown: "核心结论\n第二版",
	})
	if second == nil {
		t.Fatal("expected second analysis record")
	}

	latest := service.GetLatestHouseholdAIAnalysis()
	if latest == nil {
		t.Fatal("expected latest analysis record")
	}
	if latest.TriggerSource != "test:second" {
		t.Fatalf("unexpected latest trigger source: %s", latest.TriggerSource)
	}
}

func TestHouseholdMembersRefreshProfileAndContext(t *testing.T) {
	setupHouseholdTestDB(t)
	service := NewService()

	member := service.CreateHouseholdMember(HouseholdMember{
		Name:         "张斌",
		Relationship: "本人",
		IsPrimary:    true,
		IsActive:     true,
	})
	if member == nil {
		t.Fatal("expected member to be created")
	}
	second := service.CreateHouseholdMember(HouseholdMember{
		Name:         "配偶",
		Relationship: "配偶",
		IsActive:     true,
	})
	if second == nil {
		t.Fatal("expected second member to be created")
	}

	profile := service.GetHouseholdProfile()
	if profile.MembersCount != 2 {
		t.Fatalf("expected members count 2, got %d", profile.MembersCount)
	}

	context := service.BuildHouseholdAIContext("天津市")
	if len(context.Members) != 2 {
		t.Fatalf("expected 2 members in context, got %d", len(context.Members))
	}

	income := service.CreateHouseholdIncome(HouseholdIncome{
		Name:          "工资",
		IncomeType:    "salary",
		Owner:         "张斌",
		Frequency:     "monthly",
		MonthlyAmount: 10000,
		IsActive:      true,
	})
	if income == nil {
		t.Fatal("expected linked income")
	}

	if !service.DeleteHouseholdMember(member.ID) {
		t.Fatal("expected member to be deleted")
	}
	incomes := service.GetHouseholdIncomes()
	if len(incomes) == 0 || incomes[0].Owner != "" {
		t.Fatalf("expected owner reference to be cleared, got %#v", incomes)
	}
}

func TestGetHouseholdLiquidAssetTrendAndDistribution(t *testing.T) {
	setupHouseholdTestDB(t)
	service := NewService()

	first := service.CreateHouseholdAccount(HouseholdAccount{
		Name:        "工资卡",
		AccountType: "bank",
		Balance:     80000,
		IsLiquid:    true,
		IsActive:    true,
	})
	if first == nil {
		t.Fatal("expected first account")
	}
	second := service.CreateHouseholdAccount(HouseholdAccount{
		Name:        "零钱",
		AccountType: "wechat",
		Balance:     20000,
		IsLiquid:    true,
		IsActive:    true,
	})
	if second == nil {
		t.Fatal("expected second account")
	}

	trend := service.GetHouseholdLiquidAssetTrend(30)
	if len(trend) == 0 {
		t.Fatal("expected liquid asset trend")
	}
	if trend[len(trend)-1].TotalLiquidAssets != 100000 {
		t.Fatalf("expected liquid assets 100000, got %.2f", trend[len(trend)-1].TotalLiquidAssets)
	}

	distribution := service.GetHouseholdLiquidAssetDistribution()
	if len(distribution) != 2 {
		t.Fatalf("expected 2 liquid asset distribution items, got %d", len(distribution))
	}
	totalShare := distribution[0].ShareOfLiquid + distribution[1].ShareOfLiquid
	if math.Abs(totalShare-100) > 0.01 {
		t.Fatalf("expected liquid shares to total 100, got %.2f", totalShare)
	}
}
