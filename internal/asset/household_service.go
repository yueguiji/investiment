package asset

import (
	"math"
	"sort"
	"strings"
	"time"

	"go-stock/backend/db"
	"go-stock/backend/logger"
)

func (s *Service) GetHouseholdAccounts() []HouseholdAccount {
	var items []HouseholdAccount
	db.Dao.Order("account_type asc, name asc").Find(&items)
	return items
}

func (s *Service) GetHouseholdMembers() []HouseholdMember {
	var items []HouseholdMember
	db.Dao.Order("relationship asc, name asc").Find(&items)
	return items
}

func (s *Service) CreateHouseholdMember(item HouseholdMember) *HouseholdMember {
	item.Name = strings.TrimSpace(item.Name)
	if item.Name == "" {
		return nil
	}
	if err := db.Dao.Create(&item).Error; err != nil {
		logger.SugaredLogger.Errorf("create household member failed: %v", err)
		return nil
	}
	s.refreshHouseholdMemberCounts()
	s.SaveHouseholdSnapshot("member:create")
	return &item
}

func (s *Service) UpdateHouseholdMember(item HouseholdMember) *HouseholdMember {
	item.Name = strings.TrimSpace(item.Name)
	if item.Name == "" {
		return nil
	}
	if err := db.Dao.Save(&item).Error; err != nil {
		logger.SugaredLogger.Errorf("update household member failed: %v", err)
		return nil
	}
	s.refreshHouseholdMemberCounts()
	s.SaveHouseholdSnapshot("member:update")
	return &item
}

func (s *Service) DeleteHouseholdMember(id uint) bool {
	var existing HouseholdMember
	if err := db.Dao.First(&existing, id).Error; err != nil {
		logger.SugaredLogger.Errorf("load household member failed: %v", err)
		return false
	}
	if err := db.Dao.Delete(&HouseholdMember{}, id).Error; err != nil {
		logger.SugaredLogger.Errorf("delete household member failed: %v", err)
		return false
	}
	s.clearMemberOwnerReferences(existing.Name)
	s.refreshHouseholdMemberCounts()
	s.SaveHouseholdSnapshot("member:delete")
	return true
}

func (s *Service) CreateHouseholdAccount(item HouseholdAccount) *HouseholdAccount {
	now := time.Now()
	if item.LastUpdatedAt == nil {
		item.LastUpdatedAt = &now
	}
	if err := db.Dao.Create(&item).Error; err != nil {
		logger.SugaredLogger.Errorf("create household account failed: %v", err)
		return nil
	}
	s.SaveHouseholdSnapshot("account:create")
	return &item
}

func (s *Service) UpdateHouseholdAccount(item HouseholdAccount) *HouseholdAccount {
	now := time.Now()
	item.LastUpdatedAt = &now
	if err := db.Dao.Save(&item).Error; err != nil {
		logger.SugaredLogger.Errorf("update household account failed: %v", err)
		return nil
	}
	s.SaveHouseholdSnapshot("account:update")
	return &item
}

func (s *Service) DeleteHouseholdAccount(id uint) bool {
	if err := db.Dao.Delete(&HouseholdAccount{}, id).Error; err != nil {
		logger.SugaredLogger.Errorf("delete household account failed: %v", err)
		return false
	}
	s.SaveHouseholdSnapshot("account:delete")
	return true
}

func (s *Service) GetHouseholdFixedAssets() []HouseholdFixedAsset {
	var items []HouseholdFixedAsset
	db.Dao.Order("asset_type asc, name asc").Find(&items)
	return items
}

func (s *Service) CreateHouseholdFixedAsset(item HouseholdFixedAsset) *HouseholdFixedAsset {
	if item.OwnershipRatio <= 0 {
		item.OwnershipRatio = 1
	}
	if err := db.Dao.Create(&item).Error; err != nil {
		logger.SugaredLogger.Errorf("create household fixed asset failed: %v", err)
		return nil
	}
	s.SaveHouseholdSnapshot("fixed_asset:create")
	return &item
}

func (s *Service) UpdateHouseholdFixedAsset(item HouseholdFixedAsset) *HouseholdFixedAsset {
	if item.OwnershipRatio <= 0 {
		item.OwnershipRatio = 1
	}
	if err := db.Dao.Save(&item).Error; err != nil {
		logger.SugaredLogger.Errorf("update household fixed asset failed: %v", err)
		return nil
	}
	s.SaveHouseholdSnapshot("fixed_asset:update")
	return &item
}

func (s *Service) DeleteHouseholdFixedAsset(id uint) bool {
	if err := db.Dao.Delete(&HouseholdFixedAsset{}, id).Error; err != nil {
		logger.SugaredLogger.Errorf("delete household fixed asset failed: %v", err)
		return false
	}
	s.SaveHouseholdSnapshot("fixed_asset:delete")
	return true
}

func (s *Service) GetHouseholdIncomes() []HouseholdIncome {
	var items []HouseholdIncome
	db.Dao.Order("income_type asc, name asc").Find(&items)
	return items
}

func (s *Service) CreateHouseholdIncome(item HouseholdIncome) *HouseholdIncome {
	s.normalizeIncomeAmounts(&item)
	if err := db.Dao.Create(&item).Error; err != nil {
		logger.SugaredLogger.Errorf("create household income failed: %v", err)
		return nil
	}
	s.syncIncomeProtections(item)
	s.SaveHouseholdSnapshot("income:create")
	return &item
}

func (s *Service) UpdateHouseholdIncome(item HouseholdIncome) *HouseholdIncome {
	s.normalizeIncomeAmounts(&item)
	if err := db.Dao.Save(&item).Error; err != nil {
		logger.SugaredLogger.Errorf("update household income failed: %v", err)
		return nil
	}
	s.syncIncomeProtections(item)
	s.SaveHouseholdSnapshot("income:update")
	return &item
}

func (s *Service) DeleteHouseholdIncome(id uint) bool {
	if err := db.Dao.Where("source_income_id = ?", id).Delete(&HouseholdProtection{}).Error; err != nil {
		logger.SugaredLogger.Errorf("delete linked household protections failed: %v", err)
		return false
	}
	if err := db.Dao.Delete(&HouseholdIncome{}, id).Error; err != nil {
		logger.SugaredLogger.Errorf("delete household income failed: %v", err)
		return false
	}
	s.SaveHouseholdSnapshot("income:delete")
	return true
}

func (s *Service) GetHouseholdProtections() []HouseholdProtection {
	var items []HouseholdProtection
	db.Dao.Order("protection_type asc, name asc").Find(&items)
	return items
}

func (s *Service) CreateHouseholdProtection(item HouseholdProtection) *HouseholdProtection {
	if err := db.Dao.Create(&item).Error; err != nil {
		logger.SugaredLogger.Errorf("create household protection failed: %v", err)
		return nil
	}
	s.syncProtectionIncomeContributions(item)
	s.SaveHouseholdSnapshot("protection:create")
	return &item
}

func (s *Service) UpdateHouseholdProtection(item HouseholdProtection) *HouseholdProtection {
	if err := db.Dao.Save(&item).Error; err != nil {
		logger.SugaredLogger.Errorf("update household protection failed: %v", err)
		return nil
	}
	s.syncProtectionIncomeContributions(item)
	s.SaveHouseholdSnapshot("protection:update")
	return &item
}

func (s *Service) DeleteHouseholdProtection(id uint) bool {
	var existing HouseholdProtection
	if err := db.Dao.First(&existing, id).Error; err != nil {
		logger.SugaredLogger.Errorf("load household protection failed: %v", err)
		return false
	}
	if err := db.Dao.Delete(&HouseholdProtection{}, id).Error; err != nil {
		logger.SugaredLogger.Errorf("delete household protection failed: %v", err)
		return false
	}
	s.clearIncomeContributionForProtection(existing)
	s.SaveHouseholdSnapshot("protection:delete")
	return true
}

func (s *Service) GetHouseholdLiabilities() []HouseholdLiability {
	var items []HouseholdLiability
	db.Dao.Order("liability_type asc, name asc").Find(&items)
	return items
}

func (s *Service) CreateHouseholdLiability(item HouseholdLiability) *HouseholdLiability {
	s.normalizeLiability(&item)
	if err := db.Dao.Create(&item).Error; err != nil {
		logger.SugaredLogger.Errorf("create household liability failed: %v", err)
		return nil
	}
	if !s.rebuildLiabilitySchedule(item.ID) {
		return nil
	}
	s.SaveHouseholdSnapshot("liability:create")
	db.Dao.First(&item, item.ID)
	return &item
}

func (s *Service) UpdateHouseholdLiability(item HouseholdLiability) *HouseholdLiability {
	s.normalizeLiability(&item)
	if err := db.Dao.Save(&item).Error; err != nil {
		logger.SugaredLogger.Errorf("update household liability failed: %v", err)
		return nil
	}
	if !s.rebuildLiabilitySchedule(item.ID) {
		return nil
	}
	s.SaveHouseholdSnapshot("liability:update")
	db.Dao.First(&item, item.ID)
	return &item
}

func (s *Service) DeleteHouseholdLiability(id uint) bool {
	if err := db.Dao.Where("liability_id = ?", id).Delete(&HouseholdLiabilitySchedule{}).Error; err != nil {
		logger.SugaredLogger.Errorf("delete household liability schedules failed: %v", err)
		return false
	}
	if err := db.Dao.Delete(&HouseholdLiability{}, id).Error; err != nil {
		logger.SugaredLogger.Errorf("delete household liability failed: %v", err)
		return false
	}
	s.SaveHouseholdSnapshot("liability:delete")
	return true
}

func (s *Service) GetHouseholdLiabilitySchedules(liabilityID uint) []HouseholdLiabilitySchedule {
	var items []HouseholdLiabilitySchedule
	db.Dao.Where("liability_id = ?", liabilityID).Order("due_date asc").Find(&items)
	return items
}

func (s *Service) RebuildHouseholdLiabilitySchedule(liabilityID uint) bool {
	return s.rebuildLiabilitySchedule(liabilityID)
}

func (s *Service) GetHouseholdLiabilityTrend(monthsBack, monthsForward int) []HouseholdLiabilityTrendPoint {
	if monthsBack < 0 {
		monthsBack = 0
	}
	if monthsForward < 0 {
		monthsForward = 0
	}

	var liabilities []HouseholdLiability
	db.Dao.Find(&liabilities)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, -monthsBack, 0)
	end := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, monthsForward, 0)

	var result []HouseholdLiabilityTrendPoint
	for cursor := start; !cursor.After(end); cursor = cursor.AddDate(0, 1, 0) {
		monthEnd := endOfMonth(cursor)
		point := HouseholdLiabilityTrendPoint{
			Month: cursor.Format("2006-01"),
		}
		for _, liability := range liabilities {
			if !liability.IsActive {
				continue
			}
			outstanding, payment, principalPaid, interestPaid := s.getLiabilityPositionAt(liability, monthEnd)
			point.TotalOutstanding += outstanding
			point.TotalPayment += payment
			point.PrincipalPaid += principalPaid
			point.InterestPaid += interestPaid
		}
		point.TotalOutstanding = roundTo(point.TotalOutstanding, 2)
		point.TotalPayment = roundTo(point.TotalPayment, 2)
		point.PrincipalPaid = roundTo(point.PrincipalPaid, 2)
		point.InterestPaid = roundTo(point.InterestPaid, 2)
		result = append(result, point)
	}

	return result
}

func (s *Service) GetHouseholdLiquidAssetTrend(days int) []HouseholdLiquidAssetTrendPoint {
	snapshots := s.GetHouseholdSnapshots(days)
	result := make([]HouseholdLiquidAssetTrendPoint, 0, len(snapshots))
	for _, item := range snapshots {
		date := ""
		if item.SnapshotDate != nil {
			date = item.SnapshotDate.Format("2006-01-02")
		}
		result = append(result, HouseholdLiquidAssetTrendPoint{
			Date:               date,
			TotalLiquidAssets:  roundTo(item.TotalLiquidAssets, 2),
			MonthlyNetIncome:   roundTo(item.MonthlyNetIncome, 2),
			MonthlyDebtPayment: roundTo(item.MonthlyDebtPayment, 2),
		})
	}
	return result
}

func (s *Service) GetHouseholdLiquidAssetDistribution() []HouseholdLiquidAssetDistributionItem {
	accounts := s.GetHouseholdAccounts()
	totalLiquid := 0.0
	grouped := make(map[string]*HouseholdLiquidAssetDistributionItem)
	for _, item := range accounts {
		if !item.IsActive || !item.IsLiquid || item.Balance <= 0 {
			continue
		}
		totalLiquid += item.Balance
		name := strings.TrimSpace(item.Name)
		if name == "" {
			name = "未命名账户"
		}
		if existing, ok := grouped[name]; ok {
			existing.Balance = roundTo(existing.Balance+item.Balance, 2)
			if existing.AccountType == "" {
				existing.AccountType = item.AccountType
			}
			if existing.Provider == "" {
				existing.Provider = item.Provider
			}
			if existing.Owner == "" {
				existing.Owner = item.Owner
			}
			continue
		}
		grouped[name] = &HouseholdLiquidAssetDistributionItem{
			Name:        name,
			AccountType: item.AccountType,
			Provider:    item.Provider,
			Owner:       item.Owner,
			Balance:     roundTo(item.Balance, 2),
		}
	}

	result := make([]HouseholdLiquidAssetDistributionItem, 0, len(grouped))
	for _, item := range grouped {
		item.ShareOfLiquid = shareOfTotal(item.Balance, totalLiquid)
		result = append(result, *item)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Balance == result[j].Balance {
			return result[i].Name < result[j].Name
		}
		return result[i].Balance > result[j].Balance
	})
	return result
}

func (s *Service) GetHouseholdDashboardSummary() *HouseholdDashboardSummary {
	var (
		accounts    []HouseholdAccount
		fixedAssets []HouseholdFixedAsset
		incomes     []HouseholdIncome
		protections []HouseholdProtection
		liabilities []HouseholdLiability
	)

	db.Dao.Find(&accounts)
	db.Dao.Find(&fixedAssets)
	db.Dao.Find(&incomes)
	db.Dao.Find(&protections)
	db.Dao.Find(&liabilities)
	profile := s.GetHouseholdProfile()

	summary := &HouseholdDashboardSummary{
		AccountCount:    len(accounts),
		FixedAssetCount: len(fixedAssets),
		IncomeCount:     len(incomes),
		ProtectionCount: len(protections),
		LiabilityCount:  len(liabilities),
	}

	for _, item := range accounts {
		summary.TotalAssets += item.Balance
		if item.IsLiquid {
			summary.TotalLiquidAssets += item.Balance
		}
	}

	for _, item := range fixedAssets {
		value := item.CurrentValue * maxOwnershipRatio(item.OwnershipRatio)
		summary.TotalAssets += value
		summary.TotalFixedAssets += value
	}

	for _, item := range protections {
		summary.TotalProtection += item.CurrentBalance
		summary.TotalAssets += item.CurrentBalance
	}

	for _, item := range incomes {
		if !item.IsActive {
			continue
		}
		monthly := item.MonthlyAmount
		if monthly <= 0 && item.AnnualAmount > 0 {
			monthly = item.AnnualAmount / 12
		}
		summary.MonthlyIncome += monthly
	}

	salaryMonthly, bonusMonthly, otherMonthly := s.classifyMonthlyIncomes(incomes)
	incomeInsurance, incomeHousingFund, _ := s.sumSalaryPretaxContributions(incomes)
	pretaxInsurance := incomeInsurance
	if pretaxInsurance <= 0 {
		pretaxInsurance = NumberOrZero(profile.MonthlyPersonalInsuranceContribution)
	}
	pretaxHousingFund := incomeHousingFund
	if pretaxHousingFund <= 0 {
		pretaxHousingFund = NumberOrZero(profile.MonthlyHousingFundContribution)
	}
	pretaxCosts := roundTo(
		pretaxInsurance+
			pretaxHousingFund+
			NumberOrZero(profile.MonthlyOtherPretaxDeduction),
		2,
	)
	specialDeductions := calculateNormalizedMonthlySpecialDeductions(profile)
	salaryTax := calculateMonthlyComprehensiveTax(salaryMonthly, pretaxCosts, specialDeductions)
	bonusTax := calculateMonthlyBonusTax(bonusMonthly)
	summary.MonthlyPretaxCosts = pretaxCosts
	summary.MonthlyHousingFundInflows = s.sumMergedHousingFundInflows(incomes, protections, profile)
	summary.MonthlyIncomeTax = roundTo(salaryTax+bonusTax, 2)
	summary.MonthlyNetIncome = roundTo(
		maxFloat(salaryMonthly-pretaxCosts-salaryTax, 0)+
			maxFloat(bonusMonthly-bonusTax, 0)+
			otherMonthly,
		2,
	)

	for _, item := range liabilities {
		if !item.IsActive {
			continue
		}
		outstanding, payment, _, _ := s.getLiabilityPositionAt(item, time.Now())
		summary.TotalLiabilities += outstanding
		if payment <= 0 {
			payment = item.MonthlyPayment + item.ExtraMonthlyPayment
		}
		summary.MonthlyDebtPayment += payment
	}

	summary.NetAssets = summary.TotalAssets - summary.TotalLiabilities
	if summary.TotalAssets > 0 {
		summary.DebtRatio = roundTo(summary.TotalLiabilities/summary.TotalAssets*100, 2)
	}
	if summary.MonthlyDebtPayment > 0 {
		incomeBase := summary.MonthlyNetIncome
		if incomeBase <= 0 {
			incomeBase = summary.MonthlyIncome
		}
		summary.MonthlyCoverageRate = roundTo(incomeBase/summary.MonthlyDebtPayment, 2)
		summary.MonthlyEffectiveDebtPayment = roundTo(maxFloat(summary.MonthlyDebtPayment-summary.MonthlyHousingFundInflows, 0), 2)
		if summary.MonthlyEffectiveDebtPayment > 0 {
			summary.MonthlyEffectiveCoverageRate = roundTo(incomeBase/summary.MonthlyEffectiveDebtPayment, 2)
		}
	}

	return summary
}

func (s *Service) SaveHouseholdSnapshot(triggerSource string) *HouseholdSnapshot {
	summary := s.GetHouseholdDashboardSummary()
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	snapshot := HouseholdSnapshot{
		SnapshotDate:                 &today,
		TotalAssets:                  summary.TotalAssets,
		TotalLiquidAssets:            summary.TotalLiquidAssets,
		TotalFixedAssets:             summary.TotalFixedAssets,
		TotalProtection:              summary.TotalProtection,
		TotalLiabilities:             summary.TotalLiabilities,
		NetAssets:                    summary.NetAssets,
		DebtRatio:                    summary.DebtRatio,
		MonthlyIncome:                summary.MonthlyIncome,
		MonthlyNetIncome:             summary.MonthlyNetIncome,
		MonthlyIncomeTax:             summary.MonthlyIncomeTax,
		MonthlyPretaxCosts:           summary.MonthlyPretaxCosts,
		MonthlyHousingFundInflows:    summary.MonthlyHousingFundInflows,
		MonthlyDebtPayment:           summary.MonthlyDebtPayment,
		MonthlyEffectiveDebtPayment:  summary.MonthlyEffectiveDebtPayment,
		MonthlyCoverageRate:          summary.MonthlyCoverageRate,
		MonthlyEffectiveCoverageRate: summary.MonthlyEffectiveCoverageRate,
		TriggerSource:                triggerSource,
	}

	var existing HouseholdSnapshot
	err := db.Dao.Where("snapshot_date = ?", today).First(&existing).Error
	if err != nil {
		if err := db.Dao.Create(&snapshot).Error; err != nil {
			logger.SugaredLogger.Errorf("create household snapshot failed: %v", err)
			return nil
		}
		return &snapshot
	}

	snapshot.ID = existing.ID
	if err := db.Dao.Save(&snapshot).Error; err != nil {
		logger.SugaredLogger.Errorf("update household snapshot failed: %v", err)
		return nil
	}
	return &snapshot
}

func (s *Service) GetHouseholdSnapshots(days int) []HouseholdSnapshot {
	var items []HouseholdSnapshot
	if days <= 0 {
		days = 365
	}
	startDate := time.Now().AddDate(0, 0, -days)
	db.Dao.Where("snapshot_date >= ?", startDate).Order("snapshot_date asc").Find(&items)
	return items
}

func (s *Service) normalizeIncomeAmounts(item *HouseholdIncome) {
	switch item.Frequency {
	case "annual":
		if item.AnnualAmount <= 0 && item.MonthlyAmount > 0 {
			item.AnnualAmount = item.MonthlyAmount * 12
		}
		if item.MonthlyAmount <= 0 && item.AnnualAmount > 0 {
			item.MonthlyAmount = item.AnnualAmount / 12
		}
	default:
		if item.MonthlyAmount <= 0 && item.AnnualAmount > 0 {
			item.MonthlyAmount = item.AnnualAmount / 12
		}
		if item.AnnualAmount <= 0 && item.MonthlyAmount > 0 {
			item.AnnualAmount = item.MonthlyAmount * 12
		}
	}
}

func (s *Service) classifyMonthlyIncomes(items []HouseholdIncome) (float64, float64, float64) {
	var salaryMonthly float64
	var bonusMonthly float64
	var otherMonthly float64
	for _, item := range items {
		if !item.IsActive {
			continue
		}
		monthly := item.MonthlyAmount
		if monthly <= 0 && item.AnnualAmount > 0 {
			monthly = item.AnnualAmount / 12
		}
		switch item.IncomeType {
		case "salary":
			salaryMonthly += monthly
		case "bonus":
			bonusMonthly += monthly
		default:
			otherMonthly += monthly
		}
	}
	return roundTo(salaryMonthly, 2), roundTo(bonusMonthly, 2), roundTo(otherMonthly, 2)
}

func (s *Service) sumSalaryPretaxContributions(items []HouseholdIncome) (float64, float64, float64) {
	var insurance float64
	var housingFund float64
	var employerHousingFund float64
	for _, item := range items {
		if !item.IsActive || item.IncomeType != "salary" {
			continue
		}
		insurance += NumberOrZero(item.MonthlyPersonalInsuranceContribution)
		housingFund += NumberOrZero(item.MonthlyPersonalHousingFundContribution)
		employerHousingFund += NumberOrZero(item.MonthlyEmployerHousingFundContribution)
	}
	return roundTo(insurance, 2), roundTo(housingFund, 2), roundTo(employerHousingFund, 2)
}

func (s *Service) sumMergedHousingFundInflows(incomes []HouseholdIncome, protections []HouseholdProtection, profile *HouseholdProfile) float64 {
	type ownerHousingFund struct {
		personal float64
		employer float64
	}

	byOwner := map[string]*ownerHousingFund{}
	ensureOwner := func(owner string) *ownerHousingFund {
		key := strings.TrimSpace(owner)
		if key == "" {
			key = "_default"
		}
		if existing, ok := byOwner[key]; ok {
			return existing
		}
		entry := &ownerHousingFund{}
		byOwner[key] = entry
		return entry
	}

	for _, item := range incomes {
		if !item.IsActive {
			continue
		}
		entry := ensureOwner(item.Owner)
		entry.personal = maxFloat(entry.personal, NumberOrZero(item.MonthlyPersonalHousingFundContribution))
		entry.employer = maxFloat(entry.employer, NumberOrZero(item.MonthlyEmployerHousingFundContribution))
	}

	for _, item := range protections {
		if !item.IsActive || item.ProtectionType != "housing_fund" {
			continue
		}
		entry := ensureOwner(item.Owner)
		entry.personal = maxFloat(entry.personal, NumberOrZero(item.MonthlyPersonalContribution))
		entry.employer = maxFloat(entry.employer, NumberOrZero(item.MonthlyEmployerContribution))
	}

	total := 0.0
	for _, entry := range byOwner {
		total += entry.personal + entry.employer
	}
	if total <= 0 && profile != nil {
		total = NumberOrZero(profile.MonthlyHousingFundContribution)
	}
	return roundTo(total, 2)
}

func (s *Service) syncIncomeProtections(item HouseholdIncome) {
	s.upsertIncomeProtection(item, "social_insurance",
		item.MonthlyPersonalInsuranceContribution,
		item.MonthlyEmployerInsuranceContribution,
	)
	s.upsertIncomeProtection(item, "housing_fund",
		item.MonthlyPersonalHousingFundContribution,
		item.MonthlyEmployerHousingFundContribution,
	)
}

func (s *Service) syncProtectionIncomeContributions(item HouseholdProtection) {
	if item.SourceIncomeID == 0 {
		return
	}

	var income HouseholdIncome
	if err := db.Dao.First(&income, item.SourceIncomeID).Error; err != nil {
		logger.SugaredLogger.Errorf("load linked household income failed: %v", err)
		return
	}

	switch item.ProtectionType {
	case "social_insurance":
		income.MonthlyPersonalInsuranceContribution = roundTo(NumberOrZero(item.MonthlyPersonalContribution), 2)
		income.MonthlyEmployerInsuranceContribution = roundTo(NumberOrZero(item.MonthlyEmployerContribution), 2)
	case "housing_fund":
		income.MonthlyPersonalHousingFundContribution = roundTo(NumberOrZero(item.MonthlyPersonalContribution), 2)
		income.MonthlyEmployerHousingFundContribution = roundTo(NumberOrZero(item.MonthlyEmployerContribution), 2)
	default:
		return
	}

	if err := db.Dao.Save(&income).Error; err != nil {
		logger.SugaredLogger.Errorf("sync household protection back to income failed: %v", err)
	}
}

func (s *Service) clearIncomeContributionForProtection(item HouseholdProtection) {
	if item.SourceIncomeID == 0 {
		return
	}

	var income HouseholdIncome
	if err := db.Dao.First(&income, item.SourceIncomeID).Error; err != nil {
		logger.SugaredLogger.Errorf("load linked household income failed: %v", err)
		return
	}

	switch item.ProtectionType {
	case "social_insurance":
		income.MonthlyPersonalInsuranceContribution = 0
		income.MonthlyEmployerInsuranceContribution = 0
	case "housing_fund":
		income.MonthlyPersonalHousingFundContribution = 0
		income.MonthlyEmployerHousingFundContribution = 0
	default:
		return
	}

	if err := db.Dao.Save(&income).Error; err != nil {
		logger.SugaredLogger.Errorf("clear linked household income contribution failed: %v", err)
	}
}

func (s *Service) upsertIncomeProtection(item HouseholdIncome, protectionType string, personalAmount, employerAmount float64) {
	if item.ID == 0 {
		return
	}

	if !item.IsActive || (NumberOrZero(personalAmount) <= 0 && NumberOrZero(employerAmount) <= 0) {
		db.Dao.Where("source_income_id = ? AND protection_type = ?", item.ID, protectionType).Delete(&HouseholdProtection{})
		return
	}

	name := incomeProtectionName(item, protectionType)
	var existing HouseholdProtection
	err := db.Dao.Where("source_income_id = ? AND protection_type = ?", item.ID, protectionType).First(&existing).Error
	if err != nil {
		existing = HouseholdProtection{
			SourceIncomeID: item.ID,
			ProtectionType: protectionType,
		}
	}

	existing.Name = name
	existing.Owner = item.Owner
	existing.Provider = item.Employer
	existing.Employer = item.Employer
	existing.MonthlyPersonalContribution = roundTo(NumberOrZero(personalAmount), 2)
	existing.MonthlyEmployerContribution = roundTo(NumberOrZero(employerAmount), 2)
	existing.MonthlyPremium = 0
	existing.IsActive = item.IsActive
	if existing.Remark == "" || existing.SourceIncomeID == item.ID {
		existing.Remark = "由收入台账自动同步"
	}

	if err := db.Dao.Save(&existing).Error; err != nil {
		logger.SugaredLogger.Errorf("upsert linked household protection failed: %v", err)
	}
}

func incomeProtectionName(item HouseholdIncome, protectionType string) string {
	base := strings.TrimSpace(item.Owner)
	if base == "" {
		base = strings.TrimSpace(item.Name)
	}
	if base == "" {
		base = "收入自动同步"
	}
	switch protectionType {
	case "housing_fund":
		return base + "-公积金"
	default:
		return base + "-五险"
	}
}

func calculateMonthlyComprehensiveTax(salaryMonthly, pretaxCosts, specialDeductions float64) float64 {
	taxable := salaryMonthly - pretaxCosts - specialDeductions - 5000
	if taxable <= 0 {
		return 0
	}
	switch {
	case taxable <= 3000:
		return roundTo(taxable*0.03, 2)
	case taxable <= 12000:
		return roundTo(taxable*0.10-210, 2)
	case taxable <= 25000:
		return roundTo(taxable*0.20-1410, 2)
	case taxable <= 35000:
		return roundTo(taxable*0.25-2660, 2)
	case taxable <= 55000:
		return roundTo(taxable*0.30-4410, 2)
	case taxable <= 80000:
		return roundTo(taxable*0.35-7160, 2)
	default:
		return roundTo(taxable*0.45-15160, 2)
	}
}

func calculateMonthlyBonusTax(bonusMonthly float64) float64 {
	if bonusMonthly <= 0 {
		return 0
	}
	annualBonus := bonusMonthly * 12
	bracketBase := annualBonus / 12
	var rate float64
	var quickDeduction float64
	switch {
	case bracketBase <= 3000:
		rate, quickDeduction = 0.03, 0
	case bracketBase <= 12000:
		rate, quickDeduction = 0.10, 210
	case bracketBase <= 25000:
		rate, quickDeduction = 0.20, 1410
	case bracketBase <= 35000:
		rate, quickDeduction = 0.25, 2660
	case bracketBase <= 55000:
		rate, quickDeduction = 0.30, 4410
	case bracketBase <= 80000:
		rate, quickDeduction = 0.35, 7160
	default:
		rate, quickDeduction = 0.45, 15160
	}
	annualTax := annualBonus*rate - quickDeduction
	if annualTax < 0 {
		return 0
	}
	return roundTo(annualTax/12, 2)
}

func calculateNormalizedMonthlySpecialDeductions(profile *HouseholdProfile) float64 {
	if profile == nil {
		return 0
	}
	return roundTo(
		normalizeLikelyAnnualDeduction(profile.MonthlyChildcareDeduction)+
			normalizeLikelyAnnualDeduction(profile.MonthlyHousingLoanDeduction)+
			normalizeLikelyAnnualDeduction(profile.MonthlyElderlyCareDeduction)+
			normalizeLikelyAnnualDeduction(profile.MonthlyOtherSpecialDeduction),
		2,
	)
}

func normalizeLikelyAnnualDeduction(value float64) float64 {
	value = NumberOrZero(value)
	if value > 12000 {
		return roundTo(value/12, 2)
	}
	return value
}

func NumberOrZero(value float64) float64 {
	if value < 0 {
		return 0
	}
	return value
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func (s *Service) refreshHouseholdMemberCounts() {
	members := s.GetHouseholdMembers()
	if len(members) == 0 {
		profile := s.GetHouseholdProfile()
		profile.MembersCount = 0
		profile.DependentsCount = 0
		s.UpsertHouseholdProfile(*profile)
		return
	}
	profile := s.GetHouseholdProfile()
	profile.MembersCount = len(members)
	profile.DependentsCount = 0
	s.UpsertHouseholdProfile(*profile)
}

func (s *Service) clearMemberOwnerReferences(owner string) {
	if strings.TrimSpace(owner) == "" {
		return
	}
	db.Dao.Model(&HouseholdAccount{}).Where("owner = ?", owner).Update("owner", "")
	db.Dao.Model(&HouseholdFixedAsset{}).Where("owner = ?", owner).Update("owner", "")
	db.Dao.Model(&HouseholdIncome{}).Where("owner = ?", owner).Update("owner", "")
	db.Dao.Model(&HouseholdProtection{}).Where("owner = ?", owner).Update("owner", "")
	db.Dao.Model(&HouseholdLiability{}).Where("owner = ?", owner).Update("owner", "")
}

func (s *Service) normalizeLiability(item *HouseholdLiability) {
	if item.OutstandingPrincipal <= 0 {
		item.OutstandingPrincipal = item.Principal
	}
	if item.LoanTermMonths > 0 && item.MonthlyPayment <= 0 {
		item.MonthlyPayment = s.calculateMonthlyPayment(item.Principal, item.AnnualRate, item.LoanTermMonths, item.RepaymentMethod)
	}
	if item.MaturityDate == nil && item.FirstPaymentDate != nil && item.LoanTermMonths > 0 {
		maturity := item.FirstPaymentDate.AddDate(0, item.LoanTermMonths-1, 0)
		item.MaturityDate = &maturity
	}
}

func (s *Service) rebuildLiabilitySchedule(liabilityID uint) bool {
	var liability HouseholdLiability
	if err := db.Dao.First(&liability, liabilityID).Error; err != nil {
		logger.SugaredLogger.Errorf("load household liability failed: %v", err)
		return false
	}

	schedules := s.buildLiabilitySchedule(liability)

	if err := db.Dao.Where("liability_id = ?", liabilityID).Delete(&HouseholdLiabilitySchedule{}).Error; err != nil {
		logger.SugaredLogger.Errorf("clear household liability schedules failed: %v", err)
		return false
	}

	if len(schedules) > 0 {
		if err := db.Dao.Create(&schedules).Error; err != nil {
			logger.SugaredLogger.Errorf("create household liability schedules failed: %v", err)
			return false
		}
	}

	outstanding, _, _, _ := s.getLiabilityPositionAt(liability, time.Now())
	if outstanding <= 0 && liability.Principal > 0 {
		outstanding = 0
	}
	if err := db.Dao.Model(&HouseholdLiability{}).Where("id = ?", liabilityID).Update("outstanding_principal", outstanding).Error; err != nil {
		logger.SugaredLogger.Errorf("update household liability outstanding failed: %v", err)
		return false
	}

	return true
}

func (s *Service) buildLiabilitySchedule(liability HouseholdLiability) []HouseholdLiabilitySchedule {
	if liability.Principal <= 0 || liability.LoanTermMonths <= 0 {
		return nil
	}

	firstPayment := liability.FirstPaymentDate
	if firstPayment == nil {
		now := time.Now()
		if liability.StartDate != nil {
			start := *liability.StartDate
			first := start.AddDate(0, 1, 0)
			firstPayment = &first
		} else {
			firstPayment = &now
		}
	}

	monthlyRate := liability.AnnualRate / 12 / 100
	basePayment := liability.MonthlyPayment
	if basePayment <= 0 {
		basePayment = s.calculateMonthlyPayment(liability.Principal, liability.AnnualRate, liability.LoanTermMonths, liability.RepaymentMethod)
	}

	outstanding := liability.Principal
	equalPrincipalBase := 0.0
	if liability.RepaymentMethod == "equal_principal" {
		equalPrincipalBase = liability.Principal / float64(liability.LoanTermMonths)
	}

	schedules := make([]HouseholdLiabilitySchedule, 0, liability.LoanTermMonths)
	for i := 0; i < liability.LoanTermMonths && outstanding > 0; i++ {
		dueDate := firstPayment.AddDate(0, i, 0)
		opening := outstanding
		interest := roundTo(opening*monthlyRate, 2)
		payment := basePayment + liability.ExtraMonthlyPayment

		var principalPaid float64
		if liability.RepaymentMethod == "equal_principal" {
			principalPaid = equalPrincipalBase + liability.ExtraMonthlyPayment
			if monthlyRate <= 0 {
				payment = principalPaid
			} else {
				payment = principalPaid + interest
			}
		} else {
			principalPaid = payment - interest
		}

		if monthlyRate <= 0 && liability.RepaymentMethod != "equal_principal" {
			principalPaid = payment
		}

		if principalPaid < 0 {
			principalPaid = 0
		}
		if principalPaid > outstanding {
			principalPaid = outstanding
			payment = principalPaid + interest
		}

		closing := roundTo(outstanding-principalPaid, 2)
		if closing < 0 {
			closing = 0
		}

		dueDateCopy := dueDate
		schedules = append(schedules, HouseholdLiabilitySchedule{
			LiabilityID:      liability.ID,
			DueDate:          &dueDateCopy,
			PeriodNumber:     i + 1,
			OpeningPrincipal: roundTo(opening, 2),
			PrincipalPaid:    roundTo(principalPaid, 2),
			InterestPaid:     interest,
			PaymentAmount:    roundTo(payment, 2),
			ClosingPrincipal: closing,
			Status:           "planned",
		})

		outstanding = closing
	}

	return schedules
}

func (s *Service) getLiabilityPositionAt(liability HouseholdLiability, asOf time.Time) (float64, float64, float64, float64) {
	outstanding := liability.OutstandingPrincipal
	if outstanding <= 0 {
		outstanding = liability.Principal
	}

	if !liability.AutoAmortize {
		return roundTo(outstanding, 2), roundTo(liability.MonthlyPayment+liability.ExtraMonthlyPayment, 2), 0, 0
	}

	schedules := s.GetHouseholdLiabilitySchedules(liability.ID)
	if len(schedules) == 0 {
		return roundTo(outstanding, 2), roundTo(liability.MonthlyPayment+liability.ExtraMonthlyPayment, 2), 0, 0
	}

	monthStart := time.Date(asOf.Year(), asOf.Month(), 1, 0, 0, 0, 0, asOf.Location())
	monthEnd := endOfMonth(asOf)

	latestOutstanding := liability.Principal
	var payment float64
	var principalPaid float64
	var interestPaid float64

	for _, item := range schedules {
		if item.DueDate == nil {
			continue
		}
		if !item.DueDate.After(asOf) {
			latestOutstanding = item.ClosingPrincipal
		}
		if !item.DueDate.Before(monthStart) && !item.DueDate.After(monthEnd) {
			payment += item.PaymentAmount
			principalPaid += item.PrincipalPaid
			interestPaid += item.InterestPaid
		}
	}

	return roundTo(latestOutstanding, 2), roundTo(payment, 2), roundTo(principalPaid, 2), roundTo(interestPaid, 2)
}

func endOfMonth(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month()+1, 0, 23, 59, 59, 0, value.Location())
}

func (s *Service) calculateMonthlyPayment(principal, annualRate float64, termMonths int, repaymentMethod string) float64 {
	if principal <= 0 || termMonths <= 0 {
		return 0
	}
	monthlyRate := annualRate / 12 / 100
	switch repaymentMethod {
	case "equal_principal":
		return roundTo(principal/float64(termMonths)+principal*monthlyRate, 2)
	default:
		if monthlyRate <= 0 {
			return roundTo(principal/float64(termMonths), 2)
		}
		factor := math.Pow(1+monthlyRate, float64(termMonths))
		return roundTo(principal*monthlyRate*factor/(factor-1), 2)
	}
}

func maxOwnershipRatio(ratio float64) float64 {
	if ratio <= 0 {
		return 1
	}
	if ratio > 1 {
		return 1
	}
	return ratio
}

func roundTo(value float64, decimals int) float64 {
	pow := math.Pow(10, float64(decimals))
	return math.Round(value*pow) / pow
}
