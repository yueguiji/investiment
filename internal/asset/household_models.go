package asset

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type HouseholdAccount struct {
	gorm.Model
	Name          string                `json:"name" gorm:"index"`
	AccountType   string                `json:"accountType" gorm:"index"`
	Provider      string                `json:"provider"`
	Owner         string                `json:"owner"`
	Currency      string                `json:"currency" gorm:"default:CNY"`
	Balance       float64               `json:"balance"`
	IsLiquid      bool                  `json:"isLiquid" gorm:"default:true"`
	LastUpdatedAt *time.Time            `json:"lastUpdatedAt"`
	Remark        string                `json:"remark"`
	IsActive      bool                  `json:"isActive" gorm:"default:true"`
	IsDel         soft_delete.DeletedAt `gorm:"softDelete:flag" json:"-"`
}

func (HouseholdAccount) TableName() string {
	return "household_accounts"
}

type HouseholdFixedAsset struct {
	gorm.Model
	Name           string                `json:"name" gorm:"index"`
	AssetType      string                `json:"assetType" gorm:"index"`
	Owner          string                `json:"owner"`
	OwnershipRatio float64               `json:"ownershipRatio" gorm:"default:1"`
	CurrentValue   float64               `json:"currentValue"`
	CostBasis      float64               `json:"costBasis"`
	Location       string                `json:"location"`
	ReferenceCode  string                `json:"referenceCode"`
	PurchasedAt    *time.Time            `json:"purchasedAt"`
	ValuationDate  *time.Time            `json:"valuationDate"`
	Remark         string                `json:"remark"`
	IsActive       bool                  `json:"isActive" gorm:"default:true"`
	IsDel          soft_delete.DeletedAt `gorm:"softDelete:flag" json:"-"`
}

func (HouseholdFixedAsset) TableName() string {
	return "household_fixed_assets"
}

type HouseholdIncome struct {
	gorm.Model
	Name                                   string                `json:"name" gorm:"index"`
	IncomeType                             string                `json:"incomeType" gorm:"index"`
	Owner                                  string                `json:"owner"`
	Employer                               string                `json:"employer"`
	Frequency                              string                `json:"frequency" gorm:"default:monthly"`
	MonthlyAmount                          float64               `json:"monthlyAmount"`
	AnnualAmount                           float64               `json:"annualAmount"`
	MonthlyPersonalInsuranceContribution   float64               `json:"monthlyPersonalInsuranceContribution"`
	MonthlyEmployerInsuranceContribution   float64               `json:"monthlyEmployerInsuranceContribution"`
	MonthlyPersonalHousingFundContribution float64               `json:"monthlyPersonalHousingFundContribution"`
	MonthlyEmployerHousingFundContribution float64               `json:"monthlyEmployerHousingFundContribution"`
	LastReceivedAt                         *time.Time            `json:"lastReceivedAt"`
	Remark                                 string                `json:"remark"`
	IsActive                               bool                  `json:"isActive" gorm:"default:true"`
	IsDel                                  soft_delete.DeletedAt `gorm:"softDelete:flag" json:"-"`
}

func (HouseholdIncome) TableName() string {
	return "household_incomes"
}

type HouseholdProtection struct {
	gorm.Model
	Name                        string                `json:"name" gorm:"index"`
	ProtectionType              string                `json:"protectionType" gorm:"index"`
	SourceIncomeID              uint                  `json:"sourceIncomeId" gorm:"index"`
	Owner                       string                `json:"owner"`
	Provider                    string                `json:"provider"`
	Employer                    string                `json:"employer"`
	CurrentBalance              float64               `json:"currentBalance"`
	MonthlyPersonalContribution float64               `json:"monthlyPersonalContribution"`
	MonthlyEmployerContribution float64               `json:"monthlyEmployerContribution"`
	MonthlyPremium              float64               `json:"monthlyPremium"`
	CoverageAmount              float64               `json:"coverageAmount"`
	NextDueDate                 *time.Time            `json:"nextDueDate"`
	Remark                      string                `json:"remark"`
	IsActive                    bool                  `json:"isActive" gorm:"default:true"`
	IsDel                       soft_delete.DeletedAt `gorm:"softDelete:flag" json:"-"`
}

func (HouseholdProtection) TableName() string {
	return "household_protections"
}

type HouseholdLiability struct {
	gorm.Model
	Name                 string                `json:"name" gorm:"index"`
	LiabilityType        string                `json:"liabilityType" gorm:"index"`
	Lender               string                `json:"lender"`
	Owner                string                `json:"owner"`
	Currency             string                `json:"currency" gorm:"default:CNY"`
	Principal            float64               `json:"principal"`
	OutstandingPrincipal float64               `json:"outstandingPrincipal"`
	AnnualRate           float64               `json:"annualRate"`
	LoanTermMonths       int                   `json:"loanTermMonths"`
	RepaymentMethod      string                `json:"repaymentMethod" gorm:"default:equal_installment"`
	MonthlyPayment       float64               `json:"monthlyPayment"`
	ExtraMonthlyPayment  float64               `json:"extraMonthlyPayment"`
	StartDate            *time.Time            `json:"startDate"`
	FirstPaymentDate     *time.Time            `json:"firstPaymentDate"`
	MaturityDate         *time.Time            `json:"maturityDate"`
	AutoAmortize         bool                  `json:"autoAmortize" gorm:"default:true"`
	Remark               string                `json:"remark"`
	IsActive             bool                  `json:"isActive" gorm:"default:true"`
	IsDel                soft_delete.DeletedAt `gorm:"softDelete:flag" json:"-"`
}

func (HouseholdLiability) TableName() string {
	return "household_liabilities"
}

type HouseholdLiabilitySchedule struct {
	gorm.Model
	LiabilityID      uint                  `json:"liabilityId" gorm:"index"`
	DueDate          *time.Time            `json:"dueDate" gorm:"index"`
	PeriodNumber     int                   `json:"periodNumber"`
	OpeningPrincipal float64               `json:"openingPrincipal"`
	PrincipalPaid    float64               `json:"principalPaid"`
	InterestPaid     float64               `json:"interestPaid"`
	PaymentAmount    float64               `json:"paymentAmount"`
	ClosingPrincipal float64               `json:"closingPrincipal"`
	Status           string                `json:"status" gorm:"default:planned"`
	IsDel            soft_delete.DeletedAt `gorm:"softDelete:flag" json:"-"`
}

func (HouseholdLiabilitySchedule) TableName() string {
	return "household_liability_schedules"
}

type HouseholdSnapshot struct {
	gorm.Model
	SnapshotDate                 *time.Time `json:"snapshotDate" gorm:"uniqueIndex"`
	TotalAssets                  float64    `json:"totalAssets"`
	TotalLiquidAssets            float64    `json:"totalLiquidAssets"`
	TotalFixedAssets             float64    `json:"totalFixedAssets"`
	TotalProtection              float64    `json:"totalProtection"`
	TotalLiabilities             float64    `json:"totalLiabilities"`
	NetAssets                    float64    `json:"netAssets"`
	DebtRatio                    float64    `json:"debtRatio"`
	MonthlyIncome                float64    `json:"monthlyIncome"`
	MonthlyNetIncome             float64    `json:"monthlyNetIncome"`
	MonthlyIncomeTax             float64    `json:"monthlyIncomeTax"`
	MonthlyPretaxCosts           float64    `json:"monthlyPretaxCosts"`
	MonthlyHousingFundInflows    float64    `json:"monthlyHousingFundInflows"`
	MonthlyDebtPayment           float64    `json:"monthlyDebtPayment"`
	MonthlyEffectiveDebtPayment  float64    `json:"monthlyEffectiveDebtPayment"`
	MonthlyCoverageRate          float64    `json:"monthlyCoverageRate"`
	MonthlyEffectiveCoverageRate float64    `json:"monthlyEffectiveCoverageRate"`
	TriggerSource                string     `json:"triggerSource"`
}

func (HouseholdSnapshot) TableName() string {
	return "household_snapshots"
}

type HouseholdAIAnalysis struct {
	gorm.Model
	TriggerSource    string `json:"triggerSource"`
	Region           string `json:"region"`
	BenchmarkVersion string `json:"benchmarkVersion"`
	AIConfigID       int    `json:"aiConfigId"`
	PromptTemplateID int    `json:"promptTemplateId"`
	ModelName        string `json:"modelName"`
	Status           string `json:"status"`
	Prompt           string `json:"prompt" gorm:"type:text"`
	InputPayload     string `json:"inputPayload" gorm:"type:text"`
	AnalysisMarkdown string `json:"analysisMarkdown" gorm:"type:text"`
	ErrorMessage     string `json:"errorMessage" gorm:"type:text"`
}

func (HouseholdAIAnalysis) TableName() string {
	return "household_ai_analyses"
}

type HouseholdBenchmarkRecord struct {
	gorm.Model
	Name        string  `json:"name" gorm:"index"`
	Scope       string  `json:"scope" gorm:"index"`
	Region      string  `json:"region" gorm:"index"`
	Category    string  `json:"category" gorm:"index"`
	Value       float64 `json:"value"`
	Unit        string  `json:"unit"`
	Year        int     `json:"year" gorm:"index"`
	Version     string  `json:"version" gorm:"index"`
	Description string  `json:"description"`
	IsActive    bool    `json:"isActive" gorm:"default:true"`
}

func (HouseholdBenchmarkRecord) TableName() string {
	return "household_benchmark_records"
}

type HouseholdProfile struct {
	gorm.Model
	HouseholdName                        string  `json:"householdName"`
	Region                               string  `json:"region" gorm:"index"`
	CityTier                             string  `json:"cityTier"`
	MembersCount                         int     `json:"membersCount"`
	DependentsCount                      int     `json:"dependentsCount"`
	HousingStatus                        string  `json:"housingStatus"`
	RiskPreference                       string  `json:"riskPreference"`
	MonthlyHouseholdSpend                float64 `json:"monthlyHouseholdSpend"`
	AnnualHouseholdSpend                 float64 `json:"annualHouseholdSpend"`
	PrimaryIncomeSource                  string  `json:"primaryIncomeSource"`
	MonthlyPersonalInsuranceContribution float64 `json:"monthlyPersonalInsuranceContribution"`
	MonthlyHousingFundContribution       float64 `json:"monthlyHousingFundContribution"`
	MonthlyOtherPretaxDeduction          float64 `json:"monthlyOtherPretaxDeduction"`
	MonthlyChildcareDeduction            float64 `json:"monthlyChildcareDeduction"`
	MonthlyHousingLoanDeduction          float64 `json:"monthlyHousingLoanDeduction"`
	MonthlyElderlyCareDeduction          float64 `json:"monthlyElderlyCareDeduction"`
	MonthlyOtherSpecialDeduction         float64 `json:"monthlyOtherSpecialDeduction"`
	Notes                                string  `json:"notes"`
}

func (HouseholdProfile) TableName() string {
	return "household_profiles"
}

type HouseholdMember struct {
	gorm.Model
	Name         string                `json:"name" gorm:"index"`
	Relationship string                `json:"relationship" gorm:"index"`
	Gender       string                `json:"gender"`
	BirthDate    *time.Time            `json:"birthDate"`
	Occupation   string                `json:"occupation"`
	City         string                `json:"city"`
	AnnualIncome float64               `json:"annualIncome"`
	Notes        string                `json:"notes"`
	IsPrimary    bool                  `json:"isPrimary" gorm:"default:false"`
	IsDependent  bool                  `json:"isDependent" gorm:"default:false"`
	IsActive     bool                  `json:"isActive" gorm:"default:true"`
	IsDel        soft_delete.DeletedAt `gorm:"softDelete:flag" json:"-"`
}

func (HouseholdMember) TableName() string {
	return "household_members"
}

type HouseholdDashboardSummary struct {
	TotalAssets                  float64 `json:"totalAssets"`
	TotalLiquidAssets            float64 `json:"totalLiquidAssets"`
	TotalFixedAssets             float64 `json:"totalFixedAssets"`
	TotalProtection              float64 `json:"totalProtection"`
	TotalLiabilities             float64 `json:"totalLiabilities"`
	NetAssets                    float64 `json:"netAssets"`
	DebtRatio                    float64 `json:"debtRatio"`
	MonthlyIncome                float64 `json:"monthlyIncome"`
	MonthlyNetIncome             float64 `json:"monthlyNetIncome"`
	MonthlyIncomeTax             float64 `json:"monthlyIncomeTax"`
	MonthlyPretaxCosts           float64 `json:"monthlyPretaxCosts"`
	MonthlyHousingFundInflows    float64 `json:"monthlyHousingFundInflows"`
	MonthlyDebtPayment           float64 `json:"monthlyDebtPayment"`
	MonthlyEffectiveDebtPayment  float64 `json:"monthlyEffectiveDebtPayment"`
	MonthlyCoverageRate          float64 `json:"monthlyCoverageRate"`
	MonthlyEffectiveCoverageRate float64 `json:"monthlyEffectiveCoverageRate"`
	AccountCount                 int     `json:"accountCount"`
	FixedAssetCount              int     `json:"fixedAssetCount"`
	IncomeCount                  int     `json:"incomeCount"`
	ProtectionCount              int     `json:"protectionCount"`
	LiabilityCount               int     `json:"liabilityCount"`
}

type HouseholdLiabilityTrendPoint struct {
	Month            string  `json:"month"`
	TotalOutstanding float64 `json:"totalOutstanding"`
	TotalPayment     float64 `json:"totalPayment"`
	PrincipalPaid    float64 `json:"principalPaid"`
	InterestPaid     float64 `json:"interestPaid"`
}

type HouseholdLiquidAssetTrendPoint struct {
	Date               string  `json:"date"`
	TotalLiquidAssets  float64 `json:"totalLiquidAssets"`
	MonthlyNetIncome   float64 `json:"monthlyNetIncome"`
	MonthlyDebtPayment float64 `json:"monthlyDebtPayment"`
}

type HouseholdLiquidAssetDistributionItem struct {
	Name          string  `json:"name"`
	AccountType   string  `json:"accountType"`
	Provider      string  `json:"provider"`
	Owner         string  `json:"owner"`
	Balance       float64 `json:"balance"`
	ShareOfLiquid float64 `json:"shareOfLiquid"`
}

type HouseholdChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
