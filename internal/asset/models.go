package asset

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

// AssetCategory 资产类别
type AssetCategory struct {
	gorm.Model
	Name        string                `json:"name" gorm:"uniqueIndex"`        // 类别名称：流动资产、固定资产、贷款
	Type        string                `json:"type"`                            // liquid(流动) / fixed(固定) / liability(负债)
	Icon        string                `json:"icon"`                            // 图标
	Description string                `json:"description"`                     // 描述
	SortOrder   int                   `json:"sortOrder"`                       // 排序
	IsDel       soft_delete.DeletedAt `gorm:"softDelete:flag" json:"-"`
}

func (a AssetCategory) TableName() string {
	return "asset_categories"
}

// Asset 资产记录
type Asset struct {
	gorm.Model
	CategoryID  uint                  `json:"categoryId" gorm:"index"`         // 所属类别
	Category    AssetCategory         `json:"category" gorm:"foreignKey:CategoryID"`
	Name        string                `json:"name"`                            // 资产名称（如：招商银行活期、房产-XXX小区）
	Type        string                `json:"type"`                            // liquid / fixed / liability
	Amount      float64               `json:"amount"`                          // 金额（正数为资产，负数为负债）
	Currency    string                `json:"currency" gorm:"default:CNY"`     // 币种
	Rate        float64               `json:"rate"`                            // 年化收益率/贷款利率
	StartDate   *time.Time            `json:"startDate"`                       // 起始日期
	EndDate     *time.Time            `json:"endDate"`                         // 到期日期
	Remark      string                `json:"remark"`                          // 备注
	IsDel       soft_delete.DeletedAt `gorm:"softDelete:flag" json:"-"`
}

func (a Asset) TableName() string {
	return "assets"
}

// AssetSummary 资产汇总（非持久化，计算用）
type AssetSummary struct {
	TotalLiquid    float64            `json:"totalLiquid"`    // 流动资产总额
	TotalFixed     float64            `json:"totalFixed"`     // 固定资产总额
	TotalLiability float64            `json:"totalLiability"` // 负债总额
	NetAsset       float64            `json:"netAsset"`       // 净资产
	InvestValue    float64            `json:"investValue"`    // 投资资产（来自持仓模块）
	Categories     []CategorySummary  `json:"categories"`     // 按类别汇总
}

// CategorySummary 类别汇总
type CategorySummary struct {
	CategoryID   uint    `json:"categoryId"`
	CategoryName string  `json:"categoryName"`
	Type         string  `json:"type"`
	TotalAmount  float64 `json:"totalAmount"`
	Count        int     `json:"count"`
}
